# 权重与优先级路由静态审查

## 结论

审查范围内未发现 P0。优先级、同级平滑加权、禁用候选、配置代次、已尝试排除和无候选的主体合同在代码中闭合；记录会保存请求级配置/模型 revision，并逐轮保存每个快照候选的选择原因。发现 1 项 P1 缺陷假设：平滑加权游标在选择证据与 attempt 事务成功前已推进，事务失败会留下无持久化证据的内存调度状态变化。另列 2 项 P2 可审计性/防回归缺口。本文不对低延迟自适应、健康冷却算法或重试策略作评价。

## 已核对事实与直接合同

1. **配置输入边界。** `Candidate` 明确定义 `priority`、`weight` 与执行 revision（`internal/relay/configuration.go:294-303`）。写 API 要求候选上游唯一且存在、远端模型名有效、`priority` 为 0..100、`weight` 为 1..100，并在开启事务前整体拒绝非法集合（`internal/relay/configuration.go:438-501`）；数据库亦有相同 CHECK 与 `(alias, upstream_id)` 主键（`internal/relay/store.go:86`）。模型写以 `expected_revision` 作 CAS，冲突返回 409（`internal/relay/configuration.go:468-483`），成功在同一事务中递增模型 revision、替换完整候选集合并递增全局 `config_revision`（`internal/relay/configuration.go:502-555`）。远端模型改变才递增候选 `execution_revision`；启停、优先级、权重变化由模型/global revision 隔离（`internal/relay/configuration.go:532-542`）。
2. **快照一致性防护。** 生成入口持有 `server.mu` 建立快照，之后才写请求记录（`internal/relay/proxy.go:129-149`）。快照读取模型、全局配置 revision、候选对应上游，并以 `alias:config_revision:model_revision` 选择运行时权重状态（`internal/relay/routing.go:143-164`）。管理写同样在 `server.mu` 内完成，成功后增加进程 epoch 并重建 routing map（`internal/relay/configuration.go:466-467,554-570`；`internal/relay/console.go:21-25`）。请求保存 `config_revision` 和 `model_revision`（`internal/relay/store.go:324-340`）。开始每次派发前又核对 epoch、模型 revision、上游执行 revision；变化者标为 `configuration_changed`，不会用新凭据替换旧快照继续派发（`internal/relay/routing.go:191-217`）。
3. **资格过滤与优先级。** 每轮按快照候选逐项判定：模型、上游、候选启用状态，凭据存在性，随后才排除已尝试上游；不合格项不进入 `eligible`（`internal/relay/routing.go:199-236`）。本审查只确认冷却判断在已尝试排除之前，不评价冷却机制本身。`priority` 取合格候选中的最大数值，再仅保留相等者，因此数值越大优先级越高，低组不会参与同轮权重选择（`internal/relay/routing.go:232-244`）。
4. **权重选择。** 最高优先级组先按上游 ID 升序，fingerprint 包含上游 ID、priority、weight、候选执行 revision；组或相关配置改变即清零游标（`internal/relay/routing.go:244-254`）。选择实现为平滑加权轮转：各 current 加 weight、取最大 current（升序保证平局确定性）、选中项减组总 weight（`internal/relay/routing.go:255-266`）。这不是随机抽样；权重表达长期确定性份额，而非每次独立概率抽签。
5. **已尝试排除与无候选。** `tried` 以 upstream ID 为键，成功取得候选后才置位（`internal/relay/routed_proxy.go:14,36,68`）；下一轮该上游被记录为 `already_attempted`（`internal/relay/routing.go:228-229`）。若无选中候选，零尝试返回 503 `upstream_unavailable`，已有尝试则返回 502 `upstream_attempts_exhausted`（`internal/relay/routed_proxy.go:56-62`）。空候选集合自然产生 `selected == nil`，不会创建 attempt。
6. **决策持久化。** 每轮为快照中的每个候选生成一条原因：具体过滤原因、`selected`、`lower_priority` 或 `not_selected_same_priority`（`internal/relay/routing.go:199-280`）。这些决策与选中 attempt 在一个数据库事务中插入并提交；无候选时也提交已有候选的排除原因，空配置则提交零条决策（`internal/relay/routing.go:286-319`）。详情 API 按轮次/上游读回 snapshot enabled/priority/weight、原因与观察时间（`internal/relay/console.go:45-69`）。表主键禁止同一请求、轮次、上游重复（`internal/relay/store.go:90`）。

## 发现与缺陷假设

### P1 — 持久化失败会提前消耗平滑加权槽位

**事实：** `snapshot.Runtime.Current` 在 `internal/relay/routing.go:255-266` 被原地推进；数据库事务直到 `:286` 才开始，决策/attempt 插入或 commit 可在 `:294-317` 失败。失败路径返回 `storage_unavailable`，没有回滚内存 current。

**缺陷假设：** 若 Begin/INSERT/Commit 失败但进程继续服务或存储恢复，下一请求从已推进游标继续，虽然前一选择没有可靠决策/attempt 证据且没有派发。于是实际派发序列不再等同于持久化“开始门获准序列”，短窗口权重份额可能偏移，且无法从数据库重建为何跳过某个槽位。

**影响：** 不会绕过最高优先级、启停或鉴权，也不会在失败事务中发起上游访问；影响是调度公平性与审计可解释性，故定为 P1 而非 P0。

**建议防护：** 在局部副本上计算 next current，只有选择决策与 attempt 事务 commit 成功后才发布；或保存旧值并在所有事务失败路径恢复。需保持 `server.mu` 覆盖计算、提交和发布，避免并发覆盖。应补一个注入 Begin/decision insert/attempt insert/commit 失败后恢复存储、核对下一次选择与无故障基线一致的确定性回归；本任务未实现或执行。

### P2 — 决策证据没有直接保存配置 revision

**事实：** 请求主表保存 config/model revision，attempt 保存上游/候选执行 revision；`request_route_decisions` 只保存候选启用、priority、weight、原因和时间（`internal/relay/store.go:89-90`）。详情通过同一 request 聚合这些表，能够关联请求级 revisions。

**缺陷假设：** 单独导出或查询决策行时无法自描述其配置代次，且决策未保存候选 remote model / candidate execution revision。当前详情合同可间接关联，因此不是行为错误，但会提高离线审计、迁移和事故取证误关联风险。

**建议防护：** 明确“必须经 request_id 关联 requests/request_attempts”是存储合同；若未来支持独立决策导出，再在版本化 schema 中加入所需代次字段，不应在本审查中直接改表。

### P2 — 关键语义依赖实现顺序，缺少局部不变量注释

**事实：** “更大 priority 胜出”“同权重平局按上游 ID”“空配置零决策但 503”“同一 upstream 不因不同 remote model 在单请求重复尝试”均由当前循环/排序/map 键自然形成（`internal/relay/routing.go:228-266`；`internal/relay/routed_proxy.go:14,56-68`），函数签名和类型未直接声明这些合同。

**影响：** 当前行为与架构文档一致，不是已证实缺陷；后续重构排序、候选标识或持久化位置时较易发生静默语义漂移。

**建议防护：** 在 `beginAttempt` 邻近处写最小不变量注释，并以确定性合同测试分别锁定边界；测试建议不等于本次已执行证据。

## 风险分级汇总

| 等级 | 数量 | 影响 |
|---|---:|---|
| P0 | 0 | 未见绕过资格/优先级、错误派发或决策与 attempt 分裂提交的静态证据。 |
| P1 | 1 | 存储失败可无证据推进权重游标，影响恢复后的公平性和可解释性。 |
| P2 | 2 | 决策行代次需间接关联；若干直接合同仅由实现顺序表达，增加维护风险。 |

## 未执行项与限制

- 仅静态阅读当前候选源码；未运行 Go 测试、`make verify`、统计/分布实验、服务、数据库迁移或真实供应商调用，不能据此宣称产品通过。
- 未修改产品代码、schema、测试、配置或数据；唯一产物是本报告。
- 按任务边界未深入审查 `chooseLatency` 的自适应选择、健康冷却正确性、失败分类、最多尝试次数及重试/SSE策略；仅记录它们与本次资格/权重路径的直接调用边界。
- 依赖任务的合入结果只作为当前源码背景；本报告没有复验 Bearer、有效时钟或备份清单功能。
