# 软件工厂实验评估（2026-11-01）

> **本轮真实结果：产品候选 `NO-GO`；原 QA 执行仍为 `BLOCKED—RETURN TO PLAN REVIEW`。** 这是已合入的[独立 Go/No-Go](qa/04-go-no-go.md)裁决，不是本报告作者重新测试所得。总体成熟度是“可运行、可并发、可追溯的单机实验原型”，尚不是生产软件工厂。生成本报告本身不等于产品通过。

本报告使用三种标签：**已核对事实**指 Git、源码、控制器回执或合入证据可交叉核对的事实；**静态风险/缺陷假设**指静态审查提出但未被动态复现的风险；**建议**不表示已实现或已获授权。

## 1. 结论与证据边界

- **已核对事实**：报告启动 HEAD 为 `abe28416703013c204037545ee68a2837381b933`，初始工作树干净。QA 执行候选/最终 readiness 集成点为 `ca6ba0129b58cd07673c9a122512a094e9a57ec3`；独立复核候选 `a339f906e2ea21e655a108c0ddb530be4f83419a` squash 合入为 `d2e70fd79dfe35126ef47e65b5fbe54cac546f14`，且后者是 HEAD 祖先；QA 归档合入为当前 HEAD。归档 QA 源、执行报告和索引的 SHA-256 分别为 `de72fa13…07968`、`e6417597…60f2bf`、`7acd487b…333ff`，与控制器回执一致。
- **已核对事实**：独立复核确认 A02/A03 Bearer 合法形式被拒、B05 最终 rename 失败会删除 existing-empty 目标，任一项足以 NO-GO；B03/B04 未完成，X01–X03（包括该冻结版本的原样 `make verify`）未执行；C05 在[执行者报告](qa/03-execution.md)中自报 Pass，但[独立复核](qa/04-go-no-go.md)判为“证据不足”，后者是最终裁决。
- **已核对事实**：20 份产品静态审查最终都已合入；这不改写首轮 17/20 有效 handoff、3 个 Native exit 0 但没有 handoff 的失败。最初恢复阶段只到 19/20；`model-authorization` attempt 3 后才由 `75b16707ae1275306e6b781ad75ecfa566ab8e1a` 合入，故最终是 20/20。
- **静态风险/缺陷假设**：8 份工程审查和 20 份产品审查中的未动态验证条目仍只是假设，不能称漏洞。报告/PR 数、Worker 完成数、历史绿色门禁都不能替代产品 QA。
- **已核对事实**：没有真实供应商联调、生产部署、main/master 合并或公开网页部署；本 HTML 只是仓库下载附件。Token 与缓存量是运行遥测，不是账单；没有人工基线，不给出提效倍数。

## 2. 原型最小架构

**已核对事实**：维护源是 `agent_platform` 的 `codex/agent-team-factory-experiment` 分支，固定版本 `f264161e7961e4fc4a2e0d65c9bf7f9959e85c32`，源码入口为 [`examples/agent-team-factory`](https://github.com/big91987/agent_platform/tree/f264161e7961e4fc4a2e0d65c9bf7f9959e85c32/examples/agent-team-factory)。测试仓与工厂源仓是两个仓库。

源码关系如下：

1. `runner.py` 在显式 `plan` 时启动一次真实 Leader，解析任务后写入任务库；`run` 再用固定并发池领取任务。Leader 不是常驻、自驱循环，当前仍由协调者手动启动批次、检查评审、决定恢复与选择汇合。
2. `factory_store.py` 的 SQLite 是任务状态事实源，持久化任务、依赖、消息、submission/result、执行和事件；消息不等于授权或完成。状态与 Native 进程退出分开，`runner.py` 明确拒绝“exit 0 但没有 handoff”。
3. `git_workspace.py` 为每个任务建立 `agent-team/task-<id>-<attempt>` 分支和独立 worktree，检查允许写路径与禁入物；Worker 不能 commit/push/merge。`native.py` 启动真实 Native 会话并保留沙箱边界，`mcp_server.py` 暴露绑定任务的 `get_task`、`send_message`、`handoff`。
4. handoff 后控制器检查 artifacts/范围并保存私有候选。只有人工显式选择 `ready` 任务，`runner.py` 才调用 `git_workspace.py` 以 squash 汇入 `agent-team/experiment`；私有 candidate commit 与公开 `integration_commit` 分开。`publish.py` 对 bundle 摘要、目标 HEAD、非主线分支及远端读回做约束。

这是一套单机 worktree + Native 沙箱原型，不是多机队列，也不是多租户容器隔离。独立 worktree 降低文件冲突，但不自动解决共享 CPU、缓存、网络、凭据边界或跨租户对抗隔离。

## 3. 实际工作组成与结果

**已核对事实**：本轮不是“20 个产品功能”。工作由以下不同性质的条目组成：

| 类型 | 数量/内容 | 结果边界 |
|---|---|---|
| 产品实现候选 | [Bearer 认证](strict-bearer-auth.md)、[严格备份清单](strict-backup-manifest.md)、[单次时钟推进](single-clock-advance.md) | 三项都曾原样运行完整 `make verify`，单项约 20.5 分钟；第六轮并发批次为 `1264.136s`。这是候选门禁，不替代最终 QA。 |
| 测试设施 | [Restore rename seam](qa/restore-testability.md)、[时钟持久化故障夹具](qa/clock-fault-fixture.md) | seam 保留红断言并暴露 existing-target 被删除；时钟夹具提供四阶段确定性故障。 |
| 工程静态审查 | 8 项 | 静态审查，不是动态产品验收。 |
| 产品静态审查 | 首轮 20 项；最终 20 份合入 | 首轮 17 handoff；3 个 exit 0 无 handoff。恢复中先达 19/20，最终 attempt 3 达 20/20。 |
| QA | 计划、多轮 readiness、执行、独立复核 | v4 readiness 为可执行许可；执行中止；独立裁决 NO-GO。 |
| 工厂工具回归 | 修复后 20 个独立 SID 的纯工具链回归 | 20/20 实际调用 get_task/send_message/handoff、无文件改动；不是产品审查或产品验收。 |

首轮失败会话使用同 SID/worktree 恢复。最初两个恢复成功、累计 19/20；当时 `model-authorization` 仍按 `get_task` 指令停止，没有伪造报告。后来 attempt 3 成功交接并合入。延迟工具发现修复对应维护源 [Issue 7](https://github.com/big91987/agent_platform/issues/7)，仍开放待主线评审，不能声称已根治。产品缺陷另有 Issue 75（Bearer 合法形式）和 Issue 76（Restore 空目标目录失败原子性），本任务未创建或修改这些 Issue。

## 4. 首轮 20 项并发时序

**已核对事实**：逐项值由控制器从原始 `native/<id>-1/status.json` 并核 events 后提供；[离线看板](index.html)直接使用相同 20 条数值绘图。共同时间原点是最早 start `1793509458.2377396`；最晚 end 与最早 start 之差为 `191.08635807037354s`。持续时间中位数为 `115.54100668430328s`，首轮有效 handoff 17，Worker 峰值 20；计入同时规划 Leader 的全部 Native 峰值 21。所有 20 个 Native 都 exit 0，但三项无 handoff，证明进程成功退出不等于任务成功。

| 任务 | 相对开始(s) | 相对结束(s) | 持续(s) | 首轮 handoff |
|---|---:|---:|---:|---|
| auth-session | 0.039 | 110.921 | 110.883 | 是 |
| members-roles | 0.000 | 104.562 | 104.562 | 是 |
| invitation-consumption | 0.045 | 170.189 | 170.144 | 是 |
| console-isolation | 0.022 | 125.291 | 125.269 | 是 |
| app-key-lifecycle | 0.053 | 95.793 | 95.740 | 是 |
| model-authorization | 0.108 | 94.577 | 94.469 | **否** |
| upstream-secrets | 0.146 | 111.717 | 111.572 | 是 |
| request-boundaries | 0.212 | 117.060 | 116.848 | 是 |
| weighted-routing | 0.310 | 138.923 | 138.613 | 是 |
| latency-routing | 0.228 | 130.920 | 130.692 | 是 |
| health-cooldown | 0.270 | 117.976 | 117.706 | 是 |
| retry-replay | 0.333 | 130.960 | 130.627 | 是 |
| rate-limits | 0.426 | 110.688 | 110.262 | 是 |
| quota-billing | 0.357 | 191.086 | 190.729 | 是 |
| concurrency-admission | 0.471 | 121.383 | 120.913 | 是 |
| sse-lifecycle | 0.443 | 109.931 | 109.488 | 是 |
| history-diagnostics | 0.450 | 154.501 | 154.050 | 是 |
| network-boundaries | 0.546 | 114.781 | 114.234 | **否** |
| sqlite-integrity | 0.538 | 108.870 | 108.332 | 是 |
| frontend-state | 0.543 | 110.017 | 109.474 | **否** |

并发显著压缩了静态审查的墙钟跨度，但不能据此称“20 倍效率”。截至控制器采集点，端到端累计 `17196.41137933731s`（约 4.78 小时），包含准备、工具根因修复、重试、QA 多轮计划/评审、诊断等待与人工协调，不只是模型编码时间。实现门禁约 20.5 分钟、评审返工、环境准备及人工汇合仍是主要串行成本；本报告曾因缺逐项时序和源码入口真实 blocked 后接续，也是协调成本实例。

## 5. 质量证据、反例与未验证范围

- **已核对事实**：较早执行版本 `f20cc091a20b75e92c5c63ce1fb9e854297d9180` 曾在本机/执行机通过 37 项源码隔离回归；维护源随后更新到 `4eedbe…` 的 41 项，最终固定 `f264161…` 在本机和执行机 112 上通过完整 44 项源码隔离回归，Git 为 2.55.0。升级证据不能倒推产品 QA 通过。
- **已核对事实**：三个代码候选各自取得原样 `make verify` 完整门禁，约 20.5 分钟；第六轮并发批次 `1264.136s`。受控测试 upstream 不是供应商联调。
- **已核对事实**：独立复核否定了实现报告中“Bearer scheme 必须精确大小写/空格”的提案；RFC 合同要求 scheme 大小写不敏感并接受 `1*SP`。这说明绿色候选门禁仍可能固化错误合同。
- **已核对事实**：Restore 测试设施在 existing-empty 目标的最终 rename 注入失败后观测到目标被删除，红断言保留；独立 QA B05 又得到 10/10 同向结果。
- **未验证**：B03 的完整 tar entry/type/path/size/order 矩阵、B04 每个负例的 absent+existing-empty 原子性、X01–X03、该冻结 QA 版本的原样 `make verify`；C05 缺完整业务表及受控 upstream 零点机械断言。
- **静态风险/缺陷假设**：审查目录里的其他发现没有动态复现证据时，只能作为后续验证输入，不能汇总成“已确认漏洞数”。

## 6. 瓶颈与优先建议

### 测试仓工程设施

| 优先级 | 建议 | 依据与预期改善 | 不解决的边界 |
|---|---|---|---|
| P0 | 修正 Bearer 合同实现与 Restore 最终安装原子性，并为新冻结候选补齐 B03/B04/C05 后从头跑 A/C/B/X。 | 两个动态失败直接导致 NO-GO；补齐后才能形成可裁决同版证据。 | 不由工厂平台替产品决定 HTTP 合同或恢复算法。 |
| P1 | 让仓库固定验证入口在执行前机械检查计划矩阵、测试名和 receipt 覆盖映射。 | 本轮到 B03/B04 才发现冻结测试缺项，造成整轮返工。 | 不能用清单完整性替代断言质量与独立 QA。 |
| P2 | 继续缩短隔离门禁的缓存准备与历史基线获取，并保留失败证据。 | 每候选约 20.5 分钟，是串行关键路径。 | 不以跳过 race/browser/历史回归换速度。 |

### 通用工具 / Agent Platform

| 优先级 | 建议 | 依据与预期改善 | 不解决的边界 |
|---|---|---|---|
| P0 | 将工具延迟发现、启动就绪、首调前用户补充投递和 handoff 缺失检测作为稳定协议，并完成 Issue 7 主线评审。 | 首轮 3 个 exit 0 无 handoff，报告又因缺输入 blocked；可减少“进程成功、任务失败”和人工补料。 | 不保证 Agent 业务判断正确，也不是产品 QA。 |
| P1 | 提供脱敏事件导出：逐项时序、版本、candidate/integration 映射、峰值与状态快照。 | 本报告最初只收到汇总，无法复验时间条。 | 不导出原始会话、凭据或把 token 解释成费用。 |
| P2 | 增加不确定集成/冲突的受控恢复与跨机队列设计评估。 | 当前保留现场但无完整自动恢复，且仅单机。 | 不声称现有 worktree 沙箱等于容器或多租户隔离。 |

### 软件工厂流程

| 优先级 | 建议 | 依据与预期改善 | 不解决的边界 |
|---|---|---|---|
| P0 | 在实现前冻结合同，计划→独立 readiness→冻结测试/候选→执行→独立裁决；任何缺项回到计划，不拼接旧绿证据。 | Bearer 错合同与 B03/B04 缺项说明合同和矩阵必须前置。 | 流程不能替代技术修复。 |
| P1 | 把 candidate commit、公开 integration commit、证据哈希和最终裁决设为交接必填。 | squash 下私有 candidate 与公开父链不同；当前做法已能追溯。 | 数量完整不代表质量通过。 |
| P2 | 下一轮只做最小闭环：修两项 P0 → 补齐并独立评审唯一 QA → 新 F 全量执行 → 独立 Go/No-Go → 再决定实验分支汇合/发布。 | 限制并行扩张，优先消除已知失败和缺证。 | 不授权部署、主线合并或供应商联调。 |

## 7. 发布边界

**已核对事实**：维护源 `agent_platform`，分支 `codex/agent-team-factory-experiment`，原型目录 `examples/agent-team-factory`。首批精确实验 HEAD `aade8d2f72d42ca66349d6127cfd78a3a012b1cf` 已由正式 bundle 发布器发布并远端读回。后续 QA 归档与本报告仍需控制器最终汇合、bundle 发布和读回；Worker 没有提交、推送、合并或部署。

没有合入 main/master，没有生产部署，没有真实供应商联调。`index.html` 是可离线打开的仓库附件，不是公开部署网页。

## 8. 证据索引

### QA（最终链）

- [最终合同计划 v4（计划，不是结论）](qa/01-contract-plan-v4.md)
- [最终 readiness v4（执行许可，不是产品通过）](qa/02-readiness-v4.md)
- [QA 执行报告（BLOCKED；含执行者 C05 自报）](qa/03-execution.md)
- [脱敏 receipt 哈希索引](qa/evidence/03-execution/receipt-index.md)
- [独立 Go/No-Go（最终产品裁决 NO-GO）](qa/04-go-no-go.md)

### 8 项工程静态审查

[知识合同](engineering-audit/01-knowledge-contracts.md)、[模块边界](engineering-audit/02-relay-boundaries.md)、[环境复现](engineering-audit/03-reproducible-environment.md)、[验证证据](engineering-audit/04-verification-evidence.md)、[CI/发布控制](engineering-audit/05-ci-release-controls.md)、[测试组合](engineering-audit/06-test-portfolio.md)、[运维恢复](engineering-audit/07-operations-recovery.md)、[交付追溯](engineering-audit/08-delivery-traceability.md)。这些是静态风险输入，不是动态漏洞清单。

### 20 项产品静态审查（最终合入；首轮仅 17 handoff）

[认证会话](product-audit/01-auth-session.md)、[成员角色](product-audit/02-members-roles.md)、[邀请消费](product-audit/03-invitation-consumption.md)、[控制台隔离](product-audit/04-console-isolation.md)、[Key 生命周期](product-audit/05-app-key-lifecycle.md)、[模型授权（attempt 3 最终合入）](product-audit/06-model-authorization.md)、[上游密钥](product-audit/07-upstream-secrets.md)、[请求边界](product-audit/08-request-boundaries.md)、[权重路由](product-audit/09-weighted-routing.md)、[延迟路由](product-audit/10-latency-routing.md)、[健康冷却](product-audit/11-health-cooldown.md)、[重试重放](product-audit/12-retry-replay.md)、[速率限制](product-audit/13-rate-limits.md)、[额度计量](product-audit/14-quota-billing.md)、[并发准入](product-audit/15-concurrency-admission.md)、[SSE 生命周期](product-audit/16-sse-lifecycle.md)、[历史诊断](product-audit/17-history-diagnostics.md)、[网络边界](product-audit/18-network-boundaries.md)、[SQLite 完整性](product-audit/19-sqlite-integrity.md)、[前端状态](product-audit/20-frontend-state.md)。

### 实现、设施与源码

- 三项实现报告：[Bearer](strict-bearer-auth.md)、[Backup manifest](strict-backup-manifest.md)、[Clock advance](single-clock-advance.md)
- 测试设施：[Restore seam](qa/restore-testability.md)、[Clock fault fixture](qa/clock-fault-fixture.md)
- 产品源码入口：[认证实现](../../../internal/relay/identity.go)、[恢复实现](../../../internal/relay/backup.go)、[时钟实现](../../../internal/relay/store.go)、[冻结独立 QA 源](../../../internal/relay/qa_independent_regression_test.go)
- 工厂源码固定入口：[`examples/agent-team-factory` @ `f264161…`](https://github.com/big91987/agent_platform/tree/f264161e7961e4fc4a2e0d65c9bf7f9959e85c32/examples/agent-team-factory)
- 精确提交：实现/设施公开汇合可由当前 Git 历史复验；Restore seam candidate `031093b6e46d4399a1129b1d3e8e20454aaaad92`，时钟夹具 candidate `185439c788acac3034d64ff405aeb053bf815a13` / integration `2aeeb3af8db73a33f6fcb99371769b79f865e367`，QA candidate/integration 见本报告第 1 节。

## 9. 本报告自身验证

只检查本目录三个报告文件，未启动 Model Relay、未运行产品 QA、`make verify`、Go 测试或浏览器产品旅程。执行摘要（命令只读取三个报告文件及其相对链接目标）：

```text
node <内联报告校验器>                             # JSON/HTML/链接/敏感模式/一致性
node <内联最小 DOM smoke>                        # 执行内嵌脚本，确认生成 20 行
git diff --no-index --check /dev/null <报告文件>  # 逐文件空白检查
git status --short                               # 写路径范围
```

- JSON 解析、稳定字段类型与 20 条时序数量检查：通过。
- 由 20 条原始数据重算 span、持续时间中位数、handoff 数：分别等于 `191.08635807037354`、`115.54100668430328`、`17`。
- HTML 本地离线基本展示检查：结构与自包含资源检查通过；最小 DOM smoke 实际执行内嵌脚本并生成 20 行时间条。无外部 CDN、字体、图片、分析/追踪请求。
- README 中本地相对链接存在性、三文件关键数字一致性、绝对私有路径和常见敏感模式扫描：通过。
- `git diff --check` 与写路径范围检查：通过；仅 `index.html`、`README.md`、`summary.json`。

这些检查只是报告制品验证，不是产品浏览器 QA，也不改变最终 `NO-GO`。
