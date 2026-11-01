# SSE 流式生命周期静态审查

日期：2026-11-01  
范围：`internal/relay/proxy.go` 的 `stream`、`routing.go` 的 `attemptOutput`、`routed_proxy.go` 流式分支、`protocol.go` 响应校验，以及与服务端终止合同直接相关的持久化、计量和前端消费。  
方法边界：仅静态阅读；未启动服务，未发送流请求，未运行测试、构建或验证入口，未修改产品代码。本文不是产品通过结论，也不主张真实供应商兼容性。

## 结论

整体设计已有明确的“首个有效帧提交后不换路”屏障、逐帧 flush、请求上下文向上游传递、严格帧/总量上限、截断检测、错误脱敏和未知用量保留机制。但终止顺序存在一个 P0：`[DONE]` 写给客户端之前先持久化 success；若终止帧写失败或此刻客户端取消，补写 canceled 会被数据库单次终态条件静默拒绝，导致客户端未完整结束而请求、尝试和计费仍可能保持 success。

分级：P0 1 项，P1 2 项，P2 2 项。

## 关键生命周期（事实）

1. 外层请求以客户端 context 派生总超时 context（`proxy.go:19-21`），每次上游尝试再以它派生可取消 context，并在首个下游提交前运行 attempt timer（`routing.go:375-400`）。上游请求使用该 child context（`routed_proxy.go:116`），所以客户端断连、总超时、尝试超时都会取消传输。
2. 只有 200 且媒体类型精确解析为 `text/event-stream` 才进入流解析（`routed_proxy.go:149-179`）。非 200 不转发供应商正文；只在受控大小且唯一 JSON 时提取 usage（`routed_proxy.go:150-170`）。
3. `bufio.Scanner` 按行读取，单行上限 256 KiB、累计流上限 32 MiB；仅接受 comment 和 `data` 字段，空行结帧，多 `data` 行以换行合并；非法 UTF-8、其他 SSE 字段、超限均为协议错误（`proxy.go:184-257`）。上游 EOF 未见 `[DONE]` 一律为截断/协议错误（`proxy.go:305-324`）。
4. 非终止 data 先经唯一键 JSON 和响应结构校验，再写给下游并 flush（`proxy.go:284-304`，`protocol.go:325-374`）。`attemptOutput.Write` 在持锁状态下检查父 context/attempt timeout、停止 timer、设置 committed 和时间，再首次写到底层 writer（`routing.go:403-424`）。因此首个“通过校验且成功进入底层 Write 的 data 帧”构成不可换路屏障；提交后的错误返回流内 `event: error`，不会重试其他候选（`proxy.go:203-214`，`routed_proxy.go:213-221`）。
5. 每个转发 data 帧都调用 `ResponseController.Flush`（`proxy.go:192-202`）；comment 仅在已经提交后转发并 flush（`proxy.go:248-255`）。未提交前 comment 被丢弃，不构成提交。
6. usage 仅在含对象型 `usage` 的 data 帧出现时写入当前 attempt（`proxy.go:289-290`）；attempt 结束时持久化，request 汇总所有 attempt。字段缺失/非法/负数成为 unknown，而不是零；结算保留 pending 和额度（`protocol.go:311-323`，`routing.go:71-140`，`accounting.go:343-449`）。
7. 前端管理页将多 `data:` 行合并，识别 `[DONE]` 和流内 JSON error；EOF 无 `[DONE]` 报中断（`web/app.js:518-536`）。这与服务端主要终止合同一致。工作台的简化消费者逐行 JSON 解析，见 P2-2。

## 发现

### P0-1 `[DONE]` 未成功交付前提前固化 success，补偿终态被静默丢弃

事实：`stream` 识别 `[DONE]` 后先调用 `finish("success", "")`，随后才检查客户端 context、写并 flush `[DONE]`（`proxy.go:258-281`）。`streamFinish` 会先把 attempt 持久化为 success，再把 request 持久化为 success并结算（`routed_proxy.go:181-211`）。若之后 `[DONE]` 写/flush 失败或客户端取消，代码再次调用 canceled 收尾（`proxy.go:273-280`）。但 Schema 3 的 attempt 更新只允许 `finished_at IS NULL` 且 request 仍为 started（`routing.go:322-330`），request 更新只允许 `state='started'`（`store.go:343-368`）；两处均不检查 `RowsAffected`。因此第二次调用可以返回“成功”但实际更新零行。

影响：客户端看到 EOF/取消而没有 `[DONE]`，服务端却可能保留 request/attempt=`success`、`successful_upstream_id` 和已完成结算。它破坏直接合同“完整终止才成功”、诊断一致性及最终计量状态；在硬预算场景还会把本应标记为交付不完整的请求当作成功证据。现有 `sse_terminal_regression_test.go:129-178` 静态上意图覆盖该边界，但其期望与上述 SQL 单次终态条件冲突；本审查未执行测试，不能断言测试当前结果。

建议防护：将终止形成单一原子状态机。至少在 `[DONE]` 成功写入并 flush、且再次确认客户端 context 未取消后，才一次性持久化 attempt/request success；失败则一次性持久化 canceled。所有条件更新检查影响行数，禁止把零行更新当成功。若业务必须先落账再写终止帧，则需要明确不可逆的 `delivered_unknown`/`stream_interrupted` 状态，而不是事后伪装可回滚。

### P1-1 usage 帧可被后续部分 usage 对象覆盖，已知计量可能退化为 unknown

事实：每遇到对象型 `usage` 就整体执行 `record.Usage = usageFrom(value)`（`proxy.go:289-290`）；`usageFrom` 对缺字段返回 nil（`protocol.go:311-323`）。`validResponse` 对 choices 为空的流帧只要求 `usage` 是对象，不要求 token 字段的类型、非负性、字段间一致性或单调性（`protocol.go:325-344`）。因此先收到完整 usage、后收到 `{}` 或部分 usage 的合法帧，会覆盖先前已知值。反方向则后帧覆盖前帧，合同依赖供应商发帧顺序。

影响：最终 attempt/request 的 token 可由已知变 unknown，触发 pending/保留额度；或不同字段来自同一“最后 usage 帧”而丢失更早证据。不会把 unknown 当零，这是已有安全防护，但会产生不必要的未决计量和人工对账。

建议防护：定义 usage 合并合同（每字段只从 unknown 变 known；冲突则协议错误或保留冲突证据），并校验非负整数及可选的 `prompt+completion=total` 一致性。为多 usage 帧、部分字段、冲突值和 usage 后截断增加回归。

### P1-2 comment 心跳不续 idle timeout，兼容供应商可能被误判超时

事实：idle timer 只在成功验证 JSON data 帧后 reset（`proxy.go:284-295`）；合法 SSE comment 即使已转发也不 reset（`proxy.go:248-255`）。

影响：用 `: keep-alive` 保持长生成连接、但数据间隔超过配置 idle timeout 的上游会被主动关闭并记 `upstream_timeout`。这是安全的资源上限，但若“idle”合同指连接无活动而非无有效模型数据，就属于兼容性错误；未提交前还可能换路，造成另一候选执行同一生成。

建议防护：明确 idle 的产品语义。若是字节/帧活动超时，经过大小和 UTF-8 校验的 comment 应续期；若是有效数据超时，文档应明确 comment 不续期并用受控用例固定该行为。

### P2-1 服务端只实现受限 SSE 子集，错误分类全部归并为协议错误

事实：解析器拒绝标准 SSE 的 `event`、`id`、`retry` 字段以及未知字段（`proxy.go:226-230`），供应商 JSON error data 也因存在 `error` 被 `validResponse` 拒绝（`protocol.go:333-335`），最终统一映射 `upstream_protocol_error` 并向下游发送脱敏 relay error（`proxy.go:203-214,284-287`）。

影响：不会泄漏供应商诊断，这是正向防护；但标准 SSE 合法扩展和明确的上游错误帧不可区分，运营诊断只能看到协议错误，可能掩盖供应商限流/安全拒绝。提交前该错误目前可能被视为非 temporary 而不换路；提交后一定不能换路。

建议防护：把“OpenAI 兼容 data-only SSE”写成直接合同；若需兼容标准字段，允许并安全忽略 `id/retry`，受控识别 `event:error`/JSON error 到独立脱敏错误码，仍不得透传敏感正文。

### P2-2 工作台简化 SSE 消费者不支持多行 data 合同

事实：管理页消费者正确合并多 `data:` 行（`web/app.js:518-535`）；工作台消费者却逐行取 `data:` 并立即 `JSON.parse`（`web/workspace.js:271`）。服务端会把上游多行 data 原样重建为多行 data（`proxy.go:257,298`）。

影响：服务端接受并转发的合法多行 JSON event，工作台可能在第一行即解析失败。此问题只影响内置前端体验，不改变 API 客户端合同。

建议防护：工作台复用管理页的逐 event 合并逻辑，并统一 CRLF、EOF、error event 与缓冲上限处理。

## 已有防护与未发现问题

- 首帧提交前，协议错误/attempt timeout 不会向下游写数据，可按路由策略重试；首帧提交后绝不换路，避免重复已可见输出（`routing.go:403-424`，`routed_proxy.go:213-232`）。
- 写入均设置不晚于总超时的 10 秒 write deadline；每帧显式 flush（`routed_proxy.go:312-318`，`proxy.go:192-202`）。
- 客户端 context 传递至上游请求；写失败和取消均有 canceled 分类路径。除 P0 的终止先后问题外，普通数据帧写失败会立即停止（`proxy.go:298-301`）。
- `[DONE]` 不能作为首帧；无任何已提交帧时收到它会返回 502 协议错误（`proxy.go:258-262`）。上游正常 EOF、过长行、非法 UTF-8、累计超限均不会伪造 `[DONE]`。
- 错误帧由 relay 自建且不包含供应商正文；已提交流保持 HTTP 200 并使用 `event: error`，未提交时返回普通 JSON 失败，这是 HTTP/SSE 不可逆提交后的合理分界。
- usage 缺失不会记零；多尝试 usage 逐 attempt 汇总，未知字段维持 unknown/pending。最终 request usage 来源是 attempt 表而非易变内存 record（`store.go:351-359`）。

## 假设与未执行

- 假设审查时 Schema 3 是当前主要运行模式；旧 schema 缺少 `state='started'` 条件，P0 的具体最终行值可能不同，但“成功先于终止交付”仍成立。
- 假设 API 成功合同以客户端收到并 flush `[DONE]` 为完整流终止；路线与发布文档要求“保持合法协议、失败无成功伪装”，支持该解释（`docs/roadmap.md:36`，`docs/releases/v0.1.0.md:26`）。
- 未执行任何现有测试，包括 `sse_terminal_regression_test.go`、routing、metering 或前端测试；未验证真实网络断连时 Go transport/writer 的具体时序。
- 未验证供应商是否会发送 comment 心跳、多 usage 帧、标准 SSE 扩展或多行 JSON；这些兼容性判断均为静态合同分析，不是现场事实。
- 依赖任务的源码候选已由控制器标为合入；本报告未复验其动态行为，且三项依赖与本次 SSE 核心发现无直接因果关系。

