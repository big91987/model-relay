# 模型别名、权限与发现一致性静态审查

## 结论

本次仅做静态审查，未启动产品、未发 HTTP 请求、未运行测试，也未审查路由算法或 Bearer parser。审查范围为 `configuration.go` 的模型/候选配置、`console_keys.go` 的 `authorizeModel`、`server.go` 的 `/v1/models`、生成入口及其直接合同。

未发现无需 grant 即可调用模型的静态授权绕过。发现 1 项 P1 和 1 项 P2：模型 alias 在远端模型映射变化后仍沿用既有 tenant/key grants，可能让旧授权无感获得语义上不同的新模型；alias 接受大小写但不规范化，精确匹配虽保持执行一致，却容易形成大小写近似的不同授权对象。未发现 P0。

## 直接合同与一致性矩阵（事实）

| 环节 | 静态合同 | 一致性判断 |
|---|---|---|
| alias 输入 | `validAlias` 接受 1–128 字节名称，字符仅限 ASCII 字母、数字、`.`、`_`、`-`，但不改写大小写（`internal/relay/server.go:110-122`）；模型创建/更新路径用它校验 alias（`internal/relay/configuration.go:459-465`），生成请求也用同一函数校验 `model`（`internal/relay/protocol.go:169-183`）。 | 配置与调用语法一致；没有规范化，`Foo` 与 `foo` 是不同对象。 |
| 模型及候选配置 | 模型记录含 alias、enabled、revision；候选含 upstream、remote model、enabled、priority、weight、execution revision（`internal/relay/configuration.go:294-303,332-359`）。更新使用 expected revision，整组删除再插入候选；远端模型变化会递增候选 execution revision（`internal/relay/configuration.go:481-542`）。 | 乐观并发和远端映射版本有防护；alias 本身不可在 PATCH 中改名（`internal/relay/configuration.go:451-452`）。 |
| Key grant 写入 | Key 的模型集合必须是无重复、且当时存在于该 tenant grants 中的精确 alias（`internal/relay/console_keys.go:122-139`），随后替换写入 `key_model_grants`（`internal/relay/console_keys.go:151-159`）。 | 新建/更新 Key 时保证 key grant 是 tenant grant 子集。 |
| 调用授权 | `authorizeModel` 要求 tenant active，并以精确 alias 同时命中 key grant 与该 key 所属 tenant grant；否则返回 `tenant_paused` 或 `model_forbidden`（`internal/relay/console_keys.go:421-434`）。生成在解析请求后调用一次（`internal/relay/proxy.go:110-128`），派发候选前再次调用（`internal/relay/routing.go:183-190`）。 | 双层授权与二次检查能覆盖授权在请求处理中被收回的主要窗口。 |
| 模型发现 | `/v1/models` 先验证 key；Schema 3 下检查 tenant active，并只查询 enabled 模型、至少一个 enabled candidate + enabled upstream + 非空凭据，且同时命中 tenant/key grants（`internal/relay/server.go:524-550`）。结果返回 alias，最多 1000 项（`internal/relay/server.go:550-569`）。 | 对同一有效 key，发现不会列出无双层 grant 或静态上不可调用的模型。Key 输入本身最多 1000 个模型（`internal/relay/console_keys.go:122-124`），因此该上限不会截断一个合规 Key 的授权集合。 |
| 实际生成 | 生成先验证请求 model alias，再授权，再取得 snapshot 并要求模型 enabled（`internal/relay/proxy.go:110-145`）。派发阶段重新验证 key 与 grant，并检查当前模型、配置 epoch/revision、上游/候选 enabled、凭据存在（`internal/relay/routing.go:183-231`）。发送上游前把 payload 的 `model` 替换为候选 `RemoteModel`（`internal/relay/routed_proxy.go:68-70,108-126`）。 | 发现使用外部 alias，实际调用使用其候选 remote model；启停、凭据和 grant 的核心条件一致。生成比发现多做动态快照/冷却等检查，因此“已发现”不保证调用必定成功。 |
| 持久化约束 | tenant/key grants 的 alias 都外键引用 `models(alias)`（`internal/relay/console_storage.go:15-16`）；候选也外键引用模型及上游（`internal/relay/store.go:84-86`）。 | 防止 grant 指向不存在模型，但不能表达 grant 对某次远端映射 revision 的认可。 |

## 发现

### P1：远端模型重映射后，旧 grant 自动授权新语义

**事实。** tenant/key grant 只保存 alias，不保存模型 revision、候选 execution revision 或 remote model（`internal/relay/console_storage.go:15-16`）。管理员可以在保留同一 alias 时替换候选及其 `remote_model`；代码仅递增模型 revision 和候选 execution revision（`internal/relay/configuration.go:511,528-542`），没有撤销、暂停或要求重新确认已有 grants。`authorizeModel` 也只按 alias 交集判断（`internal/relay/console_keys.go:421-434`）。因此 alias 若从低能力/低敏感模型重新映射到更强或不同合规模型，所有既有 tenant/key grants 会立即继承该映射；先禁用、修改、再启用也不会使 grants 失效。

**影响。** 这是配置变更后的陈旧授权：数据库关系仍合法，但授权时所认可的资源语义已经改变。可能造成权限扩张、成本/数据驻留/合规边界漂移，以及审计中“Key 从何时被批准访问该远端模型”无法回答。风险需要有权修改模型配置的管理员参与，不是匿名绕过，因此定为 P1 而非 P0。

**建议防护。** 将授权绑定到稳定且不可重用的模型身份，并记录获批的授权 revision；当任一 enabled 候选的 upstream 或 remote model 发生变化时，默认把模型置为待批准/不可发现，显式重新批准 tenant grants，再要求 Key grants 重新确认或至少证明仍是 tenant grant 的当前 revision。若产品明确把 alias 定义为可变服务等级，则合同必须明确这一继承语义，并对影响授权边界的重映射提供变更预览、强审计事件和可选择的自动撤权。增加回归合同：已授权 alias 重映射后，旧 key 在 `/v1/models` 与生成上均不可使用，直到完成规定的重新批准流程。

### P2：alias 不规范化，允许大小写近似但权限独立的对象

**事实。** `validAlias` 同时接受大写和小写字母且不做 canonicalization（`internal/relay/server.go:113-122`）；模型查询、grants、发现和调用均以原字符串精确匹配（例如 `internal/relay/configuration.go:343,363-365`、`internal/relay/console_keys.go:424`、`internal/relay/server.go:546`）。所以 `Model-A` 与 `model-a` 可同时存在，并拥有不同候选及 grants。

**影响。** 静态执行是一致的，不会把一个 alias 的 grant 当成另一个；风险在管理面的人因混淆、错误授权、审计检索遗漏和客户端大小写处理差异。定为 P2。

**建议防护。** 选择并强制单一 canonical form（建议小写 ASCII），在模型创建、grant 写入和请求解析的共同边界规范化；迁移前检测大小写折叠冲突。若必须大小写敏感，则在 API/UI 合同中明确说明，并禁止仅大小写不同的 alias。

## 已有防护

- 模型配置有 expected revision，候选 remote model 改变会递增 execution revision（`internal/relay/configuration.go:481-542`）。
- 配置提交会递增全局 config revision 并通知运行态（`internal/relay/configuration.go:554-570`）；派发前会比较 epoch、模型 revision 和上游 execution revision，变化时拒绝使用旧 snapshot（`internal/relay/routing.go:191-223`）。
- 模型、候选、上游启停以及凭据存在性同时参与发现；生成派发再次验证相应条件（`internal/relay/server.go:532-550`；`internal/relay/routing.go:214-231`）。
- Key grant 写入时强制为当时 tenant grant 的子集；调用时仍动态求 tenant/key grant 交集，所以撤销 tenant grant 可立即收回调用权限（`internal/relay/console_keys.go:133-139,421-434`）。
- 生成授权不是一次性检查：请求解析后检查一次，实际选择候选前再次检查（`internal/relay/proxy.go:119-128`；`internal/relay/routing.go:183-190`）。

## 假设与未执行

**假设。** 本报告把“remote model/upstream 的变化可能改变授权资源语义”视为安全边界；若产品合同明确规定 grant 授权的是永久可变的 alias 服务而不是其当前供应模型，P1 的安全定级可下调，但仍是需要显式记录的治理风险。还假设 SQLite 外键在运行时按项目初始化方式启用；即使未启用，也不改变上述 P1 结论。

**未执行。** 按任务要求未运行服务、测试、数据库迁移或任何动态验证；未验证真实供应商响应。未审查候选选择/权重/冷却等路由算法，也未审查 Bearer parser。

**接续与工具回执。** 前一轮会话的直接工具列表未显示 factory 工具，当时未按延迟发现机制查询并且没有 handoff；平台将该次退出保留为真实失败。本次接续先通过原生执行入口的 `ALL_TOOLS` 找到 `mcp__agent_team_factory__get_task`、`send_message`、`handoff`，随后成功调用 `get_task`；回执确认本任务为 attempt 3、状态 running，三项源码依赖均为 merged。只读核对显示当前 worktree HEAD 为 `ef379a26fbe7c6159134a01a6fb0164003c33ef8`，但本报告涉及的冻结范围源码相对用户给定公开实验 HEAD `348e297f63cd4528fc05e5eaa1d7a0c2e3156d81` 无路径差异，因此无需重做审查。此项仅是源码静态一致性核对，不是产品测试或通过结论。

## 优先级汇总

| 优先级 | 数量 | 项目 |
|---|---:|---|
| P0 | 0 | 无 |
| P1 | 1 | alias 重映射后旧 tenant/key grants 自动继承新远端模型语义 |
| P2 | 1 | alias 大小写不规范化导致近似但权限独立的对象 |
