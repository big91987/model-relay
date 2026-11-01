# 重试、故障转移与请求重放静态审查

审查日期：2026-11-01  
审查方式：仅静态阅读；未运行测试、服务或供应商调用  
范围：`routed_proxy.go`、`routing.go` 的 attempt 输出与错误分类、`proxy.go`、直接合同及直接回归测试。按任务要求，不重复审查健康冷却或并发准入。

## 结论与优先级

| 优先级 | 结论 | 性质 | 影响 |
|---|---|---|---|
| P0 | 未发现可由本次静态证据确认的 P0。 | — | 不代表动态验证通过。 |
| P1 | 内部重试是至多三次、跨不同上游的 **at-least-once** 分发；供应商是否已执行但响应丢失不可判定，且客户端 `Idempotency-Key` 被刻意隔离、不下传。 | 已确认机制；“造成重复计费/副作用”的结果是缺陷假设 | 传输错误、408/429/5xx 或流式首个有效数据帧前断开时，原请求可能已经被供应商接受，随后又投递到另一上游。模型生成可能重复计费；若供应商侧请求触发外部副作用，风险更高。 |
| P1 | `transportTemporary` 除取消和目的地址禁止外，把几乎所有传输错误都当作可重试，未区分确定未送达与结果未知，也未收窄 TLS、永久 DNS、响应解析等错误。 | 已确认分类；错误是否确属永久需运行时证据 | 扩大不必要重放面；故障下最多向三个不同上游重复发送同一逻辑生成。 |
| P2 | 408、429、500、502、503、504 都立即切换；未读取 `Retry-After`，也没有 attempt 间退避。 | 已确认 | 最多三次的硬上限限制了放大，但可能加剧跨供应商限流/瞬时故障，并降低可解释性。 |
| P2 | 路线合同表述存在历史歧义：v0.1.0 S4 写“流前／流中错误分别处理，不自动重发生成”，当前 M2 路线则承诺“受限故障切换；开始输出后不重放”。实现符合后者：首个有效 SSE 数据帧前的临时错误可切换。 | 已确认文档差异 | 若读者仍以 v0.1.0 句子作为持续合同，会误判流前重试行为；应在当前合同处明确覆盖关系。 |

## 已确认事实

### 1. 最大尝试数与候选唯一性

- 外层虽然写成四轮循环（`internal/relay/routed_proxy.go:31`），但在第四次分发前检查 `len(tried) == 3` 并终止（`:64-67`），每次临时失败后也在三次处终止（`:305-307`）。因此实际最多产生/分发 **3 个 attempt**，不是 4 个。
- `tried` 以 `UpstreamID` 为键（`:14`, `:68`）；下一轮选择把已试上游标为 `already_attempted`（`internal/relay/routing.go:228-229`）。同一请求不会再次选择同一 upstream ID，即使远端模型相同。
- 每次选中后才插入 `request_attempts`，ordinal 为 `len(tried)+1`（`routing.go:307-313`）。选择轮次与实际 attempt 是不同证据：没有候选的轮次会记录 route decisions，但不会创建 attempt（`:293-315`）。
- 直接合同测试声明了四候选仍只调用三次并落三条 attempt（`internal/relay/routing_http_test.go:177-199`）。本审查只阅读该测试，未执行。

### 2. 哪些错误会切换

- HTTP 状态仅 `408, 429, 500, 502, 503, 504` 可重试（`routing.go:371-373`）。3xx 被归为 `upstream_redirect`；其余 4xx 为上游拒绝；这些以及 501/505 都不会切换（`routed_proxy.go:150-170`, `:301-303`）。
- 传输错误通过 `classify` 映射：取消为 499/`client_canceled`，deadline 或网络 timeout 为 504/`upstream_timeout`，其他非流式错误为 502/`upstream_connection_failed`，流式错误为 `stream_interrupted`（`internal/relay/proxy.go:171-183`）。
- 是否继续并不只看映射码：`transportTemporary` 对任何非 `context.Canceled`、非 `errDestinationForbidden` 的错误返回 true（`routing.go:467-469`）；attempt timer 超时也强制 temporary（`routed_proxy.go:145-147`, `:277-281`）。整体请求超时或客户端取消随后会把 temporary 清零并终止，不会重新开始（`:284-303`）。对应未执行的合同测试见 `routing_boundary_test.go:43-91`。
- 非 200 响应的 body 在 Schema 3 下最多读取 1 MiB，仅用于提取唯一 JSON 中的 usage（`routed_proxy.go:149-159`）；重试决策只由状态 allowlist 决定，不依赖响应 body。

### 3. 请求 body 重建与 header 隔离

- 入站 body 只读取一次，限制 1 MiB，随后解析成 `map[string]json.RawMessage`（`internal/relay/proxy.go:97-118`）。每次 attempt 将 `model` 改成该候选的 remote model，再重新 `json.Marshal(payload)`，并用新的 `bytes.Reader` 建立新的 POST（`routed_proxy.go:108-116`）。因此重试不依赖已消费的入站流；除每个候选的 `model` 映射外，逻辑 payload 被重建。
- 出站显式将 `GetBody=nil`（`:123`），只设置 relay 选择的 `Content-Type`、上游 `Authorization`、`X-Request-ID` 和 `Accept`（`:124-130`）。客户端 Cookie、转发头和 `Idempotency-Key` 不会透传。直接合同测试明确断言这些 header 不泄漏且 `GetBody == nil`（`internal/relay/boundary_test.go:77-99`）；未执行。
- `http.Client` 因此不能靠 `GetBody` 自行重放 POST；产品层的每次切换均显式创建新 request。3xx 也不会成为产品重试状态。

### 4. 幂等键的能力边界

- Schema 3 的 admission 将可选 `Idempotency-Key` 限制为 `validName`，并对规范化 payload JSON 取摘要（`internal/relay/accounting.go:147-157`）。相同 key + 不同 payload 返回 409 `idempotency_conflict`；相同 key + 相同 payload 返回 409 `logical_call_already_exists` 并给出原 request ID（`:158-169`）。映射在同一 admission 事务写入（`:260-264`）。
- 这能阻止客户端因网络不确定而再次创建同一网关逻辑调用，但不抑制该逻辑调用内部的最多三次 supplier attempt。键未下传供应商；供应商只收到同一个 relay `X-Request-ID`（`routed_proxy.go:126`），代码未建立供应商确认/去重协议。
- 没有幂等键时请求仍被接受，因而客户端自行重试会创建新的 request ID，并可再次触发最多三次内部尝试。

### 5. 响应提交后禁止切换

- 普通响应先完整读取（上限 8 MiB）、验证协议，再调用 `commitNormal`；该调用停止 attempt timer 并把 output 标成 committed（`routed_proxy.go:234-266`; `routing.go:453-464`）。之后任何 downstream write 失败只把当前记录改成 canceled，不进入切换循环（`routed_proxy.go:247-265`）。
- SSE 的提交点不是收到 HTTP 200，也不是供应商写 comment，而是 relay 第一次向客户端写合法数据事件：`stream` 先验证完整 SSE event 和 JSON（`proxy.go:218-288`），`attemptOutput.Write` 在首次 `text/event-stream` 写入时停止 attempt timer、标记 committed，再写到底层 writer（`routing.go:403-423`）。
- 流处理返回后，只要 `output.isCommitted()`，立即持久化/观察并 return，绝不选择下一上游（`routed_proxy.go:213-221`）。现有直接合同分别覆盖 committed stream 不重放（`routing_http_test.go:229-250`）以及首个提交后 timer 不再取消（`routing_boundary_test.go:94-105`）；均未执行。

### 6. 流式首字节前后边界

- **首个有效数据事件前**：transport EOF/timeout 经 `classify(..., true)` 成为 `stream_interrupted`/timeout，外层把这两类标为 temporary（`routed_proxy.go:213-214`）；如果 output 尚未 committed，当前 attempt 先落终态，然后可选择另一上游（`:223-233`, `:294-309`）。
- **协议错误即使在首字节前也不切换**：非法 SSE 行、过长事件、非法 JSON、只有 `[DONE]`、正常 EOF 无 `[DONE]` 都由 `stream.fail` 归为 `upstream_protocol_error`（`proxy.go:218-229`, `:258-261`, `:284-286`, `:306-316`），外层 temporary 条件不包含该 code。该保守规则避免对已被供应商处理、但返回坏协议的生成重放。
- **首个有效数据事件后**：任何中断、idle timeout、协议错误、客户端断开或下游写失败都只能结束当前 attempt/request；不会故障转移。中途失败若已 opened，会追加脱敏 relay error SSE，而不是改写已发送的 HTTP 状态（`proxy.go:203-214`）。
- `[DONE]` 前先持久化 success，再向客户端写终止帧（`proxy.go:258-282`）；终止帧写失败会把记录改为 canceled/downstream failure，但仍不重放。直接回归测试还区分 DONE 写失败/取消的归因（`sse_terminal_regression_test.go:129-178`）；未执行。

## 缺陷假设与建议防护

### P1：结果未知的 POST 被自动重放

**假设。** `client.Do` 的错误或可重试 HTTP 状态不证明供应商没有执行生成；特别是发送完成后读响应失败、网关收到 5xx/429、或 SSE 首个合法数据事件前断开时，另一个上游会收到同一逻辑请求。客户端幂等键不能约束供应商，因为它被隔离。

**建议防护。** 将产品合同明确写成“内部故障转移是 at-least-once，最多三个不同上游”，并在 UI/API 证据中继续暴露所有 attempt 与 `dispatch_possible`。若要把风险降为可控能力，应按供应商能力配置是否允许 pre-commit retry、是否支持供应商幂等令牌，并只在能够证明未发送时自动重试；不能把 relay request ID 假设成供应商去重保证。涉及工具/副作用请求时，默认禁用结果未知后的自动切换更安全。

### P1：传输临时性分类过宽

**假设。** 当前 `transportTemporary` 的否定列表过短，TLS 证书/握手、永久 DNS、畸形 HTTP 响应等确定性错误可能被当作临时故障并重放到其他供应商。是否出现取决于 Go transport 返回的具体错误链，本次未动态构造验证。

**建议防护。** 建立正向 allowlist：连接建立前的明确临时网络错误、timeout（且总 deadline 仍有预算）才允许切换；对 TLS 校验、URL/协议、永久 DNS、响应解析错误终止。为“请求已写出、响应未知”单独分类并在证据/UI 中显示，而非统称 connection failed。

### P2：状态重试无退避/Retry-After

**假设。** 对 429/503 的立即跨上游切换可能在共享供应商或共享配额场景形成短时放大。

**建议防护。** 保留三次硬上限，同时为可重试状态设置受总 deadline 约束的短抖动退避；仅在明确受信的响应上解析有界 `Retry-After`。若产品目标是立即跨独立供应商切换，也应在合同中明确忽略 Retry-After 是有意策略。

### P2：合同措辞需消歧

**建议防护。** 在当前 M2/公开 API 合同中直接说明：“流式首个有效数据事件前，仅传输中断/超时可切换；协议错误不切换；任何数据事件提交后绝不切换。”同时把 v0.1.0 的“不自动重发生成”标注为单上游历史合同，避免与当前路线产生两种解释。

## 现有防护汇总

- 最多 3 次，且每个 upstream ID 最多一次；整体 deadline/客户端取消优先终止。
- 只有固定 HTTP 状态 allowlist 和传输/attempt timeout 会切换；协议错误、普通 4xx、重定向、目的地址禁止不会切换。
- 入站 body 有 1 MiB 上限；普通上游响应有 8 MiB 上限；失败 body 用量解析有 1 MiB 上限；SSE 有单事件 256 KiB、总计 32 MiB 限制。
- 每次 attempt 在可能分发前登记，Schema 3 在 `client.Do` 前把 `attempt_dispatches.possible=1`（`routed_proxy.go:131-141`），便于保守计量/审计；登记不等价于供应商确认执行。
- 普通响应完整验证后才提交；流式合法首事件提交后停止 attempt timer，并以 committed 状态禁止切换。
- 客户端幂等键在网关内阻止同 key 同 payload 新建第二个逻辑调用；冲突 payload 明确 409；键和客户端敏感 header 不下传。
- 每个 attempt 的状态、error code、上游状态与 usage 分开落库；请求级 usage 聚合能表达 unknown/incomplete（`routing.go:71-140`, `:322-330`）。

## 未执行项与证据限制

- 按任务边界，未运行 `go test`、`make verify`、任何服务、浏览器或供应商调用；本文不作产品通过、QA 通过或真实供应商兼容结论。
- 未动态验证：请求已经写出后 read-side transport error 的实际重放；TLS/DNS/HTTP transport 错误链分类；429/503 的跨上游时序；三次 attempt 的 body 字节一致性；幂等并发竞争；首个 SSE 数据帧前后的真实 socket 断开；downstream partial write 的 committed 证据。
- 引用的 `*_test.go` 仅作为直接合同/预期行为的静态证据，不是本任务执行收据。
- 本次没有审查或评价健康冷却算法、权重/延迟选择质量、并发 admission、配额正确性，也没有调用真实供应商。

## 影响判断

当前实现对“有限故障切换”和“输出后不重放”有清晰、代码级的硬边界，最大尝试数、候选去重、body 重建和提交状态也相互一致。剩余核心风险不是无限重试，而是任何自动重试 POST 都无法从本地错误证明供应商未执行：现有幂等键只保护客户端到 relay 的重复提交，不能把内部 failover 变成 exactly-once。该风险应作为 P1 产品合同和策略问题处理；在没有真实供应商幂等能力前，不应宣称无重复执行或无重复计费。
