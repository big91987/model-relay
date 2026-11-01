# 13. RPM／TPM 限流边界静态审查

## 结论

本项仅审查候选源码，没有启动服务、发送 HTTP 请求、运行测试或修改产品。未发现可仅凭静态证据定为 P0 的 RPM／TPM 绕过。确认的合同是固定 UTC 日历分钟、租户与逻辑 Key 双层聚合、受理时保守预留、未知 usage 持续占用预留；`nil` 是无限，显式 `0` 是零额度。发现 1 项 P1 缺陷假设：token 计数在 SQL 聚合结果与 Go 加法交界处缺少显式溢出检查，极端但合法的上游 `int64` usage 可能使 TPM 使用量回绕，从而错误放行或返回存储故障。另列 2 项 P2 加固建议，均不等同于已确认线上缺陷。

金额额度不在本项定级（留第 14 项）；并发槽位不在本项定级（留第 15 项）。本文只在它们与 RPM／TPM 原子受理共享代码时说明边界，不评价其产品正确性。

## 审查口径

- **事实**：可由当前候选源码直接推出。
- **缺陷假设**：存在可达代码路径，但本项未执行构造性验证，不能写成已复现故障。
- **防护**：当前实现对相应风险的明确约束。
- **未执行**：因任务授权为静态审查而没有取得的运行证据。
- 严重度：P0 为普遍或低门槛失守；P1 为可造成限流错误且需要修复；P2 为纵深防御、合同清晰度或异常数据韧性问题。

## 事实与直接合同

1. `ResourceLimits` 的五个数值维度均为 `*int64`，因此 JSON `null`／省略字段进入 `nil`；`ResourceUsage` 使用 `int64`（`internal/relay/accounting.go:16-34`）。`overLimit` 仅在指针非空时判断，并使用“已用已超”或“新增量大于剩余额度”的减法形式（`accounting.go:112-114`）。由此直接推出：`nil` 不限、`0` 对任何新增请求拒绝；不是“0 表示无限”。
2. 管理 PUT 使用严格 JSON 解码，拒绝未知字段和重复键（`internal/relay/server.go:89-108`）；所有非空限制必须在 `[0, 10^12]`（`internal/relay/console_api.go:128-138`）。更新要求 revision 精确匹配，并把限制、审计和提交置于同一事务（`console_api.go:139-173`）。PUT 是完整替换语义：省略字段会写成 SQL `NULL`，即取消该维度限制。
3. `resource_limits` 对非空五维均有非负 CHECK；新租户和新 Key 初始均插入全 `NULL` 限制（`internal/relay/console_storage.go:24,41-42`；`internal/relay/console_keys.go:235-238`）。`platform_limits` 本身没有同等 CHECK（`console_storage.go:25`），但受支持的管理入口执行上述范围校验。
4. 窗口不是滑动 60 秒。受理时把有效时钟格式化为 `YYYY-MM-DDTHH:MM`，RPM 统计相同 `minute` 的 admission 数；月 Calls 使用 `YYYY-MM`（`accounting.go:82-103,272-288`）。Store 时钟统一转 UTC，Schema 3 使用持久化有效时钟（`internal/relay/store.go:32-45`）；`admissionClock` 还会保留 admissions 中较新的已接受时间，防止回拨重开旧窗（`accounting.go:116-125`）。边界因此是 UTC 分钟 `[HH:MM:00, 下一分钟)`。
5. 每次新逻辑调用先读取 tenant 限制和逻辑 Key 限制，再在同一事务内按 tenant、Key 顺序检查 Calls/RPM/TPM，最后插入唯一 admission 并提交（`accounting.go:179-186,272-308`）。tenant 聚合列为 `tenant_id`，Key 聚合列为 `key_id`（`accounting.go:82-88`）。轮换 generation 最终解析回同一个 `key_id`（`internal/relay/console_keys.go:403-418`），所以新旧 generation 共享 Key 池；同租户多个 Key 分别计 Key 池，同时汇入同一 tenant 池。
6. TPM 是“本分钟已确认 token + 当前保留 token”。`usageFor` 汇总所有 admissions 的 `tokens_reserved`，再加 admission 分钟与当前分钟相同的账本 input/output token（`accounting.go:88-97`）。已跨分钟但仍未知的 admission 的保留量不带分钟过滤，因此会保守 carry 到以后分钟；已确认旧分钟 token 不进入新分钟。
7. 只要 tenant 或 Key 的 TPM 有限（或硬金额有限），请求进入 bounded 模式（`accounting.go:187-202`）。每个可用候选必须有正的可信 input/output bound，`n` 只能为 1；负载只允许文本消息子集，两个 output cap 不能并存，显式 cap 必须为正且不超过所有候选中最小 output bound，缺 cap 时服务端注入该 bound（`accounting.go:221-249,311-340`）。
8. admission 预留最多三个候选的最大 token bound 之和，与最多三次 attempt 的模型一致；排序后取前三并在任何限额检查前计算（`accounting.go:251-271`；attempt ordinal 的 schema 约束见 `internal/relay/store.go:89`）。单候选 input/output bound 各不得超过 `10^9`，所以该处最多约 `6×10^9`，不会在 `int64` 内溢出（`accounting.go:253-258`）。
9. 结算时，缺 input 或 output 分量分别保留对应 bound；已知两分量但 total 缺失／不一致时保留 bound 与已知值之间的余量，并标 pending。只有已知且一致的 token 才进入追加账本；最终以一次 UPDATE 释放或缩减 admission 预留（`accounting.go:375-448`）。上游负 token 在协议解析时变为 unknown，而不是负扣减（`internal/relay/protocol.go:311-323`）。
10. 同一进程中，生成受理在 `server.mu` 下调用 `admit`（`internal/relay/proxy.go:129-157`）；单 Store 又将 SQLite 连接数限制为 1，安装目录有排他文件锁（`internal/relay/store.go:48-69`）。因此受支持的单实例部署中，读 usage、判限、插 admission、提交是串行事务，不存在两个请求同时读到同一旧余额后都成功提交的静态路径。拒绝发生在 INSERT 前，不会部分扣 RPM/TPM（`accounting.go:272-299`）。

## 发现与优先级

### P1（缺陷假设）：TPM 聚合边界缺少 checked addition

`usageFor` 分别让 SQLite 求 `sum(tokens_reserved)` 与 `sum(input_tokens+output_tokens)`，随后直接执行 `value.TPM += knownTokens`（`accounting.go:88-97`）。`overLimit` 假定 `used`、`additional` 均为未溢出的非负数（`accounting.go:112-114`）。上游 usage 的单字段只约束为非负 `int64`，没有按可信模型 bound 截顶后再入账（`protocol.go:316-323`；实际超 bound 只删除未来 bound，见 `accounting.go:442-445`）。因此在费率为零或足够小、金额计算仍不溢出的组合下，极大的已确认 token 与正预留相加可能在 Go 中回绕为负数，继而让 TPM 比较错误放行；SQLite 自身 `sum()` 超范围则可能报错并转成 503，但这不能覆盖“两个各自可表示的聚合值在 Go 相加时溢出”的交界。

影响：有限 TPM 的 tenant 与 Key 都可能在异常／恶意上游 usage 下失去 fail-closed 性；管理 usage 也可能不可用或显示错误。触发需要供应商返回接近 `MaxInt64` 的 token 数，现实门槛高，但这是外部不可信计量边界，故列 P1 而非 P2。建议对所有 token 加法使用 checked non-negative addition，并在异常时以明确的计量完整性错误失败关闭；不要用饱和值静默改写账本事实。该项未执行，仍是缺陷假设。

### P2（加固）：schema 对直接写入的约束不对称

`resource_limits` 有非负 CHECK，但 `platform_limits` 没有，`admissions.tokens_reserved`、`charge_entries.input_tokens/output_tokens` 也没有非负 CHECK（`console_storage.go:24-32`）。受支持的 API 与协议解析目前阻止负管理值和负上游 usage，故没有确认正常路径可写入负数；不过恢复、迁移缺陷或未来内部调用若写入异常值，会破坏 `overLimit` 的非负前提。建议 schema 与读取边界同时校验。影响以异常数据韧性为主，P2。

### P2（合同清晰度）：管理 PUT 的 `null`／省略均会扩大到无限

严格解码能阻止拼错字段，但 `ResourceLimits` 无法区分“字段省略”与“显式 null”，PUT 随后完整覆写全部列（`console_api.go:128-169`）。这与当前完整替换合同一致，不是实现自相矛盾；风险在于客户端若误把 PUT 当 PATCH，省略 RPM/TPM 会解除限制。建议在管理 API 文档/UI 明示“完整替换、null/省略=无限”，或要求五维字段显式出现。影响是配置误操作，P2。

## 防护汇总

- `nil` 与 `0` 语义清楚，管理上限 `10^12`，resource schema 再做非负约束。
- UTC 有效时钟和 admission 高水位避免主机回拨提前重置 RPM。
- tenant 与 Key 在一个事务中依次检查，任一失败均在 admission 写入前返回。
- bounded 模式要求 execution 版本绑定的正 input/output bound，并限制为可证明有界的文本负载。
- 未报告、部分报告、不一致 total、失败／取消后的未知 usage 均保留 token 预留；跨分钟未知量继续占用 TPM。
- 单实例互斥、单 SQLite 连接和安装目录排他锁共同保护受支持部署的原子判限。

## 未执行与边界

- 未运行任何 Go/前端/浏览器测试，未执行 `make verify`，未启动实际 Go 服务，未发送管理或生成请求；因此没有运行时 Pass 结论。
- 未构造 `MaxInt64` usage、SQL `sum()` 溢出、分钟边界、回拨、重启、多 Key、轮换 generation 或并发竞争；上述行为均为源码推导。
- 未验证真实供应商是否总会提供 usage、是否遵守可信 bound，亦未做真实 RPM/TPM 消耗。
- 未审计金额额度算法与舍入（第 14 项）；未审计并发槽位的释放、崩溃恢复和竞争公平性（第 15 项）。
- 未评价绕过安装目录锁的非受支持多进程／直接数据库写入部署；当前进程模型静态上通过排他锁拒绝第二实例，而不是实现跨进程分布式限流。

