# 请求解析、映射与参数边界静态审查

## 结论与范围

结论：未发现 P0。请求入口具备 1 MiB 限制、精确 JSON 媒体类型、全树重复键/尾随值拒绝、顶层及已建模对象未知字段拒绝、基本类型检查和别名到每次实际候选远端模型的重写。发现 2 项 P1、3 项 P2：普通（非有界）请求允许同时携带两个输出 token 上限；`response_format.json_schema.description` 未检查类型；工具相关字段缺少依赖关系检查；若干常用数值参数仅检查 JSON 数值类型而不检查协议范围；`stream_options` 空对象可通过但缺少明确合同价值。这些输入都会在通过入口验证后原样序列化给远端（仅 `model` 被替换），因此不同上游可能产生不同拒绝或解释。

本报告只审查 `internal/relay/protocol.go`、`proxy.go`、`server.go` 的 JSON 入口及其直接调用的 `accounting.go`、`routed_proxy.go` 映射边界。不覆盖认证、路由选择正确性或 SSE 生命周期。任务按授权只做静态检查；没有启动服务、发送请求、运行测试/构建或修改实现。

状态用语：**事实**为本次直接读到的源码/仓库合同；**影响**为由代码路径推得；**假设**为需要协议决策或动态测试确认；**未执行**见末节。

## 已有防护（事实）

| 边界 | 防护与关键行 |
|---|---|
| 唯一 JSON | `uniqueJSON` 先校验 UTF-8，递归记录每个对象的键并拒绝重复键，最后要求 EOF，故也拒绝第二个/尾随 JSON 值（`protocol.go:11-62`）。生成请求在解码前调用它（`:169-175`）；管理 JSON 同样调用（`server.go:89-108`）。 |
| 大小与读取 | 生成体经 `MaxBytesReader(..., 1<<20)`，超限明确 413；读取有 10 秒 deadline（`proxy.go:97-108`）。管理 JSON 上限 64 KiB，同样有 10 秒 deadline（`server.go:89-108`）。 |
| 媒体类型/对象 | 两个入口均用 `mime.ParseMediaType` 并要求 `application/json`；生成验证要求顶层对象（`proxy.go:92-110`; `protocol.go:173-175`），管理入口也要求对象（`server.go:90-108`）。 |
| 未知字段 | 生成顶层只允许明确清单，未知字段返回 `unsupported_field`（`protocol.go:177-179`）；messages、tool/function、tool choice、response format、stream options 的已建模对象也用 `onlyKeys`（`:195-199,261-303`）。管理入口使用 `DisallowUnknownFields`（`server.go:106-108`）。 |
| messages | 必须为非空数组；角色限 system/developer/user/assistant/tool；content 必须为字符串，唯一例外为 assistant 的 null content + 有效 tool calls；tool role 必须有非空 `tool_call_id`（`protocol.go:191-225`）。因此多模态 content 数组会在转发前拒绝。 |
| tools / response format | 只接受 function tool，限制对象键、名称和基本类型（`protocol.go:111-135,261-280`）；response format 只接受 text/json_object/json_schema，并对 schema 名称、schema 对象和 strict 布尔值建模（`:281-296`）。 |
| stream | 缺省 false；提供时必须为 JSON boolean；`stream_options` 只有 stream=true 时才允许，`include_usage` 若存在必须为 boolean（`protocol.go:184-190,297-303`）。 |
| token 参数 | `max_tokens`、`max_completion_tokens`、`n` 分别必须为正 int64（`protocol.go:92-95,231-232`）。有界模式额外拒绝同时存在两个上限、拒绝 n!=1、拒绝超过候选共同最大输出上限；缺省时写入显式安全上限（`accounting.go:230-249,311-340`），符合直接合同“使用 max_tokens 或 max_completion_tokens、缺省发送安全上限”（`implementation-contract.md:99-101`）。 |
| 别名重写 | 入口校验 alias 字符集和长度（`server.go:110-122`; `protocol.go:180-183`），授权/快照使用 alias（`proxy.go:110-146`）；每次实际尝试在发送前把 `payload["model"]` 替换为该候选 `RemoteModel`，再整体 JSON marshal（`routed_proxy.go:68-70,108-116`）。其余已验证字段不被删除。 |

## 发现

### RB-01 — 普通模式接受并转发两个互斥输出上限

- 严重度/状态：**P1 / 事实**。
- 证据：通用验证分别把 `max_tokens` 与 `max_completion_tokens` 当正整数，没有互斥检查（`protocol.go:226-232`）。只有进入有界准入时 `boundedPayload` 才拒绝两者并存（`accounting.go:311-340`）；非有界模式跳过该函数（`:187-188,242-249`）。发送路径除 model 外原样 marshal payload（`routed_proxy.go:108-116`）。直接合同将输出上限描述为“max_tokens 或 max_completion_tokens”（`implementation-contract.md:99`）。
- 影响：同一外部 API 输入会因租户/Key 是否启用有限 TPM 或硬预算而在本地得到不同边界；普通模式把歧义交给供应商，不同兼容端点可能拒绝、任选其一或采用不同优先级。调用者无法从 relay 合同预测语义。
- 建议防护：在 `validateRequest` 的公共协议层拒绝两者同时存在，避免预算模式改变协议合法性；补充普通和流式入口的表驱动边界测试。

### RB-02 — `json_schema.description` 未做字符串类型检查

- 严重度/状态：**P1 / 事实**。
- 证据：`json_schema` 允许键 `name description strict schema`，但代码只验证 name、schema 和可选 strict；完全未读取 description（`protocol.go:285-293`）。通过后该原始值被转发（`routed_proxy.go:108-116`）。相比之下，function description 明确要求字符串（`protocol.go:120-123`）。
- 影响：例如数字、数组或对象 description 会被 relay 认定为有效请求，随后由远端决定；这破坏入口“先按直接协议合同校验再透明转发”的一致性，并产生上游相关错误。
- 建议防护：description 存在时要求非 null JSON string；补充 null/number/array/object 拒绝和合法字符串保留测试。

### RB-03 — tools 相关字段缺少跨字段依赖检查

- 严重度/状态：**P2 / 事实 + 假设**。
- 证据：`tools`、`tool_choice`、`parallel_tool_calls` 分支各自独立验证（`protocol.go:239-240,261-280`）；代码允许空 tools 数组，也允许没有 tools 时单独出现命名 function choice、`required` choice 或 `parallel_tool_calls`。messages 内 tool call 的形状虽受严格检查，但没有校验 tool 消息引用的 ID 是否对应此前 assistant call（`:195-225`）。
- 影响（推断）：relay 接受内部无可执行工具或引用不闭合的请求并交给远端，不同端点可能拒绝或自行解释。若产品目标只是结构透明转发，这可能是有意边界；当前直接合同未明确这些依赖，故不升为 P1。
- 建议防护：先明确直接合同：至少规定 tool_choice/parallel_tool_calls 是否要求非空 tools；如承诺会话引用闭合，再验证 tool_call_id。不要静默删字段。

### RB-04 — 常用数值参数只检查类型，不检查协议取值范围

- 严重度/状态：**P2 / 事实 + 假设**。
- 证据：temperature、top_p、presence/frequency penalty 只需能转成有限 float64（`protocol.go:96-109,229-230`）；seed 接受任意 int64，top_logprobs 只设下界 0（`:233-238`）；logit_bias 的值只检查为数值（`:253-260`）。n 只要求正数，非有界模式没有上限（`:231-232`）。
- 影响（推断）：超出常见 Chat Completions 范围的值在验证后被原样转发，错误行为依赖远端；极大 n 还可能放大输出/费用风险，但有限 TPM/硬预算模式已强制 n=1，故未判 P1/P0。
- 建议防护：把项目直接协议合同的精确范围版本化后在公共验证层执行；在合同未冻结前，不应凭某单一供应商范围静默裁剪值。

### RB-05 — `stream_options:{}` 被视为有效但合同语义不清

- 严重度/状态：**P2 / 事实**。
- 证据：stream=true 时只检查对象键属于 `include_usage`，但不要求该键存在，因此空对象通过（`protocol.go:297-303`）。README 只明确承诺 `stream_options.include_usage`，没有为空对象定义行为（`README.md:97`）。
- 影响：空对象没有 relay 侧可见作用，仍被转发；严格上游可能接受或拒绝。风险较低，但这是另一处“验证通过不等于跨上游合同稳定”。
- 建议防护：合同明确允许空对象，或要求恰有 include_usage；保持原样转发，不静默补值。

## 验证后转发差异汇总

1. 通用路径解析为 `map[string]json.RawMessage`，所以合法字段的原始 JSON 表达在 marshal 时会被重新编码为对象，但值语义保留；唯一主动协议改写是 model（`protocol.go:169-309`; `routed_proxy.go:108-116`）。对象键顺序/空白不会保留，这是 JSON 语义内变化。
2. 有限 TPM/硬费用路径可能向 payload 注入缺省输出上限（`accounting.go:311-340`）；这是合同明确要求的安全改写。它发生在转发前，并会对后续实际候选共同生效。
3. 每次重试都在同一 payload 上写入当前候选远端模型，随后重新 marshal，因此 alias 不会泄漏给上游，后续候选也不会沿用前一候选模型名（`routed_proxy.go:31-36,68-70,108-116`）。
4. RB-01～05 所述字段均不会在发送前再次按候选能力验证；因此“relay 验证通过”不保证任意兼容上游接受。真实供应商能力与遵约按项目规则仍需独立证据。

## 优先级建议

1. **P1**：公共验证层拒绝同时出现两个 token 上限；补普通/流式及非有界路径回归。
2. **P1**：补 `response_format.json_schema.description` 字符串校验。
3. **P2**：冻结 tools 跨字段、数值范围和空 stream_options 的直接合同，再实现一致检查；不要按单一供应商猜测。
4. **P2**：为协议入口建立一张“字段/类型/范围/依赖/有界模式/是否改写”的版本化矩阵。当前 README 指向的 `docs/workflow/design.md` 不是协议设计，已由独立知识审查记录；本报告据源码、README 当前陈述及当前 Run 的 implementation contract 交叉核对，不把失真的链接当权威。

## 未执行与局限

- 未执行任何请求、服务启动、单元/集成测试、`make verify`、浏览器、受控上游或真实供应商调用；本报告不是产品通过结论。
- 未动态确认 Go `RawMessage` 的具体重编码字节、MaxBytesReader 的响应细节或各供应商对上述输入的实际行为；相关结论只到源码可证明的接受/转发边界。
- 未审查认证、候选路由算法、SSE 事件/完成生命周期、响应验证、计量结算正确性或管理 API 的各业务 DTO；`server.go` 仅审查共享 JSON 输入边界。
- 未使用外部协议网页。项目当前协议权威导航已知失真；需要精确外部兼容范围时，应另行用官方版本化文档冻结合同后再判定 RB-03～05 的最终规则。
