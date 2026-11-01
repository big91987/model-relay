# 自适应低延迟路由静态审查

## 结论

本次只做源码与直接合同的静态审查。审查范围内未发现 P0/P1 的已确认正确性缺陷；样本隔离、两类指标、保鲜、冷启动探索、并发串行化、异常样本拒绝和 weighted 回退均有直接实现防护。重启后会恢复持久化策略，但不会恢复近期评分、平滑加权游标或探索游标；这与现有产品合同“重启保留策略，近期评分重新学习”一致，不能当作当前缺陷。

有两项 P2 级缺陷假设值得后续澄清：其一，持久化的 latency evidence 并非路由状态恢复日志；若产品边界未来要求重启后延续 60 秒窗口内评分，当前实现不满足。其二，`generation` 是进程内计数，重启后可能复用旧代次编号，跨重启诊断存在歧义，但不影响样本键和实际选路。

## 已确认事实与用户影响

### 样本键与配置隔离

- 样本键由模型 alias、模型 revision、上游 ID、上游 execution revision、流式标志和候选 execution revision共同组成（`internal/relay/routing_latency.go:91-93`）。候选 remote model 变化会推进候选 execution revision（`internal/relay/configuration.go:532-542`），上游执行配置变化由 upstream execution revision 隔离；因此不同模型代次、上游凭据/执行代次、remote model 及流式模式不会共用评分。
- 任一配置提交会递增进程内配置代次并清空 routing/latency 内存（`internal/relay/console.go:21-26`，调用点 `internal/relay/configuration.go:554-570`）。这是偏保守的全局失效：无关模型配置变化也会让所有模型重新冷启动，可能短时失去低延迟收益，但不会继承错误评分。

### 流式与非流式指标

- 非流式指标名为 `response_body_first_byte_ms`，body wrapper 在首次读到正字节数时记时（`internal/relay/routing_latency.go:66-83`）；wrapper 仅用于非流式 200 响应（`internal/relay/routed_proxy.go:140-144`）。它是 relay 读取响应体首字节的时间，不是完整生成耗时。
- 流式指标名为 `first_output_delta_ms`。SSE 只有在已通过响应合同解析、且 delta 含非空 content/refusal 或工具函数 name/arguments 时才算有效输出（`internal/relay/routing_latency.go:194-218`）；心跳、注释、role-only、空 delta 和 usage-only 不记时。测量点在有效事件写向下游前取得并标记（`internal/relay/proxy.go:284-304`），所以它是 relay 观察到的首个有效增量，不宣称用户端首绘时间。
- 两类指标通过样本键中的 stream 布尔值分池，探索状态也按 stream 分开（`internal/relay/routing_latency.go:132-139`）。用户切换 stream 模式不会互相污染排序。

### 样本保鲜与评分

- TTL 固定 60 秒（`internal/relay/routing_latency.go:11`）；每次描述时按每个样本自己的 measurement time 过滤，仅保留年龄在 `[0,60s]` 的样本（`internal/relay/routing_latency.go:94-115`）。新样本不会给旧样本续期。
- 少于 3 个新鲜样本时状态为 `insufficient` 且不产生分数；达到 3 个后按时间切片顺序计算 alpha=0.25 的 EWMA（`internal/relay/routing_latency.go:116-128`）。每池最多保留最近 32 个已接受样本（`internal/relay/routing_latency.go:277-282`）。
- 用户影响：新候选、配置变更、重启或 60 秒无新鲜样本后会回到 weighted 行为，直到某候选积累至少 3 个新鲜成功样本；不会把“未知”误当作 0 ms。

### 冷启动探索与 weighted 回退

- low_latency 每 5 次同一 runtime/stream 选择执行一次探索；探索选择 `Last` 最小的候选并更新时间戳，按已排序 upstream ID 打破初始平局（`internal/relay/routing_latency.go:155-180`）。因此稳定候选组内每个候选最迟在 `5 × 候选数` 次选择内得到一次强制探索机会。
- 其余 4/5 请求从调用方已计算的 weighted 候选开始；只有存在至少 3 个新鲜样本的候选才可替换它，并取最低 EWMA（`internal/relay/routing_latency.go:153-180`）。这确认了“无有效评分回退 weighted”，且低延迟逻辑只作用于已经完成资格、优先级与冷却筛选的最高优先级 group（`internal/relay/routing.go:197-244,255-267`）。本报告不重复评价一般权重算法或健康冷却算法。
- fingerprint 包含当前 group 的上游、优先级、权重和候选 execution revision；变化时平滑权重状态与探索状态一起重置（`internal/relay/routing.go:246-254`）。用户影响是候选组变化后探索从头开始，避免新候选永久饥饿。

### 并发状态与异常样本接受

- 选择入口 `beginAttempt`、完成采纳 `acceptLatency` 和失败观察 `observe` 都持有同一 `server.mu`（`internal/relay/routing.go:171-174,333-335`；`internal/relay/routing_latency.go:220-223`）。因此 weighted 游标、探索计数、dispatch sequence、health watermark 和 latency pool 的读写在进程内串行化；未看到直接 data race 路径。
- 样本只在 attempt 已持久化为 success 且请求 context 未失败时继续评估；存储 degraded、未测得有效输出、时钟倒退/负值、测量已超过 TTL、配置代次变化都会拒绝（`internal/relay/routing_latency.go:224-247`）。随后还重新读取当前模型与上游，核对启用、模型 revision、上游 execution revision，并拒绝被失败 watermark 覆盖或 sequence/measurement time 乱序的完成（`internal/relay/routing_latency.go:248-260`）。
- 每次接受或拒绝先写 `route_latency_evidence`；证据写失败会进入 storage failure 且不更新内存池（`internal/relay/routing_latency.go:262-282`）。这避免“终态/证据未落库但路由已学习”的异常样本。临时失败还会同时清除该候选的流式与非流式池（`internal/relay/routing.go:353-366`）。
- `measurement.Value` 没有独立的最大值/finite 检查，但生产值来自两个 `time.Time` 的差，且超过 60 秒的测量会在完成采纳时因 TTL 被拒绝；当前未形成可利用的异常数值路径。测试注入时钟仍应避免制造非生产语义。

### 策略持久化与重启恢复

- 策略表通过数据库 CHECK 只允许 `weighted`/`low_latency`（`internal/relay/routing_latency_storage.go:9-13`），启动检查扩展完整性、版本和非法策略行（同文件 `15-44`）。读取时无策略行默认 weighted（`internal/relay/configuration.go:341-347`）；PATCH 未携带策略会保留现有策略，显式值也在 API 边界校验（同文件 `438-458,517-524`）。
- selection/latency evidence 持久化用于请求诊断和读取（`internal/relay/routing_latency_storage.go:62-108`），但 `New` 总是创建空 latency/routing/health map，未从 evidence 回灌（`internal/relay/server.go:55-75`）。因此重启恢复的是策略和历史诊断，不是近期评分、失败水位、weighted 或探索游标。
- 直接产品合同明确“重启保留策略，近期评分重新学习”（`docs/harness/run-e11d82e7f5d2/report.md:11`），并明确 60 秒、至少 3 样本、EWMA alpha=0.25、每 5 次探索和无分回退 weighted（同文件 `13-15`）。据此，当前重启行为是已确认设计，而非恢复缺陷。

## 缺陷假设与优先级

| 优先级 | 状态 | 假设与证据 | 用户影响 / 建议 |
| --- | --- | --- | --- |
| P0 | 无 | 未发现可导致授权绕过、跨候选错误路由或持久错误评分的确认问题。 | 无。 |
| P1 | 无 | 未发现会稳定破坏 low_latency 核心选择、并发安全或 weighted 回退的确认问题。 | 无。 |
| P2 | 合同变化时才成立 | evidence 已持久化，但启动只建空 map，完全不回灌 60 秒窗口内样本（`internal/relay/server.go:55-75`；`internal/relay/routing_latency_storage.go:62-108`）。当前合同明确重新学习，所以不是现行缺陷。 | 若未来承诺滚动重启无学习抖动，需要设计有界、严格校验、按完整样本键恢复的状态；不能直接把诊断 JSON 当可信运行状态。 |
| P2 | 诊断歧义假设 | `SelectionObservation.Generation`/`LatencyResult.Generation` 来自进程内 `configEpoch`（`internal/relay/routing_latency.go:151,224`），新 Server 默认从 0 开始（`internal/relay/server.go:55-75`）；跨重启可重复代次号。实际样本键仍含持久 revision，选路不受影响。 | 长期审计若仅按 generation 关联，可能把不同进程生命周期误认为同代；建议文档明确 generation 仅进程内有效，或以后增加 boot/session 标识。 |
| P2 | 保守失效的产品取舍 | 任一配置变化清空全局 latency/routing map（`internal/relay/console.go:21-26`），即使变化与某模型无关。 | 多租户频繁管理变更时，所有模型可能反复回到 weighted 冷启动；如有真实用户数据证明影响，再考虑按模型/上游精确失效。当前做法优先正确性，不建议无证据改动。 |

## 已有静态合同与防护

- 源码测试覆盖 output delta 分类、TTL 边界/逐样本过期、最多 `5N` 探索与 stream 隔离（`internal/relay/routing_latency_test.go:19-96`）。
- 源码测试还声明覆盖 16 个并发在途请求遇配置变化、旧请求晚完成的 out-of-order 拒绝（`internal/relay/routing_latency_test.go:286-410`），以及新候选不继承旧代评分且不会饥饿（同文件 `576-640`）。
- failure watermark 的直接合同覆盖成功/失败乱序及流式模式组合，并检查 `invalidated_by_failure`（`internal/relay/routing_latency_recovery_test.go:14-168`）。这些仅作为静态存在的保护证据；本任务没有执行它们，不能据此声称当前候选已通过测试。

## 未执行项与限制

- 按任务边界，未运行任何 Go 测试、race detector、延迟实验、服务、浏览器旅程或真实供应商请求；未复用依赖任务的 QA 结果作本任务通过结论。
- 未测进程重启、崩溃窗口、系统时钟跳变、高并发真实调度、SQLite 故障或 60 秒边界的运行表现。
- 未修改实现、策略、数据库、测试或其他文档；本报告是唯一产物。
- 本审查不评价一般权重选择的公平性，也不重复健康冷却机制；仅检查它们与低延迟层的直接衔接。
