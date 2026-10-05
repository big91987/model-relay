# Model Relay v0.1 实现设计 / G2

## 1. 基线、结论和责任

设计依据为本次 `requirements.md` AC-01～19、`task.md`、`../releases/v0.1.0.md` 和路线图 M1。采用单 Go 进程、嵌入式网页、SQLite、一个上游的同步中转；管理员维护接入与权限，应用仅使用密钥调用，供应商负责生成和计费。仅完成 M1/v0.1，不引入多租户、调度、额度、队列、外部 IAM、集群或计费平台。

状态：`Ready for Development`。普通选择依据现有自动推荐委托采用，不代表实现、G3、QA 或发布验收已通过。当前仓库只有文档，无运行证据，真实供应商联调未测。

按架构 Skill 加载后端与分布式系统主模块、安全架构支持模块，由本节点综合，不创建额外智能体。后端问题边界是事实源、同步事务、一次转发、取消及恢复；安全边界是三类身份、浏览器来源、目的地址与秘密。证据来自需求和第 10 节官方资料，设计置信度足以实施，行为仍需受控测试验证。不决定供应商可用率、费用或公网部署。

## 2. 用户路径和网页状态

1. 运营者通过 Go CLI 初始化独立数据目录，保存一次输出的随机管理员密码和独立加密材料；启动后访问 loopback 管理页。
2. 登录进入概览；未配置时明确显示“待配置”，引导到上游，再建立至少两个模型别名。保存表示本地配置提交，不表示供应商连接已验证。
3. 密钥页输入名称，签发后一次展示完整值与复制按钮；刷新后只见名称、短标识、启用状态。停用按钮二次确认，提交成功才改变显示。
4. 调用测试页输入应用密钥、别名、文本和普通/流式选项，明确提示可能产生供应商费用；只有点击发送才调用 `/v1/chat/completions`。支持取消，不提供绕过鉴权的管理测试 API。
5. 记录页按时间倒序分页，展示请求 ID、别名、密钥标识、结果、延迟、实际用量/未知；概览显示配置和存储健康，不伪造调用统计。
6. 运营者按正式命令停止、备份、恢复到空目录、重启并重新登录；应用密钥和禁用状态保留，浏览器旧会话不恢复。

网页采用普通 HTML/CSS/JavaScript，Go `embed` 同源提供，无 CDN、前端框架或外部资源依赖。桌面导航与窄屏纵向布局；记录表窄屏可滚动，表单有 label、键盘焦点、错误区域和忙碌状态。配置加载失败不显示空配置成功态；保存失败保留输入，不乐观更新已保存值。

全页面状态为加载/未登录/待配置/可操作/提交中/失败；测试额外有等待首包/接收中/完成/取消/不完整。只有合法结束标记才显示流完成。调用结果使用 `textContent`，不把生成内容、模型名、上游诊断或输入当 HTML。密钥、密码、对话仅在当前页面内存；禁止 localStorage、sessionStorage、IndexedDB、URL 或 Service Worker 存储。测试页离开、退出和取消后清空敏感输入；创建结果关闭后不再可恢复。

不新增静态原型：路径和状态足以实现，小项目避免维护第二套 UI；因此不调用静态 browser.check，也不将原型当 AC-18 证据。

## 3. 模块、权威写入和 API

建议布局：`cmd/model-relay/` 为 Go CLI；`internal/server/` 为路由及网页；`internal/auth/` 为管理会话/应用鉴权；`internal/store/` 为 SQL 与迁移；`internal/relay/` 为协议和受约束传输；`internal/security/` 为加密/地址校验；`web/` 为嵌入资源；`tests/` 为隔离 HTTP/UI 测试。按责任拆包，不建立通用渠道插件、事件总线或多层 Repository 框架。

| 能力/对象 | 事实源与唯一写入路径 | 生效、退出和失败 |
|---|---|---|
| 安装、管理员 | CLI `init` 写数据库密码摘要与独立密钥文件 | 排他创建；已有目录拒绝覆盖；启动不自动补初始化 |
| 管理会话 | 登录建立进程内会话，退出删除 | 8 小时绝对到期；重启全部失效 |
| 唯一上游 | `PUT /admin/api/upstream`，事务提交唯一配置行 | 校验后提交；启停/改配置影响后续快照；不自动探测/生成 |
| 模型别名 | `POST /admin/api/models`、`PATCH /admin/api/models/{alias}` | 别名是唯一业务键，不改名；禁用保留历史；冲突 409 |
| 应用密钥 | `POST /admin/api/keys` 签发、`PATCH /admin/api/keys/{id}` 启停 | ID 稳定引用身份；摘要为权威认证值；无找回/删除历史 |
| 调用记录 | relay 在转发前创建、终结时更新；启动处理未完成项 | 管理只读；不开放修改、伪造成功或重放 |

管理查询包括 `GET /admin/api/session`、`GET /admin/api/upstream`、`GET /admin/api/models`、`GET /admin/api/keys`、`GET /admin/api/requests?cursor=…&limit=…`；登录/退出为 `POST /admin/api/login`、`POST /admin/api/logout`。管理接口均认证（登录除外），响应 `Cache-Control: no-store`；写操作成功须已提交，校验失败 400、未登录 401、来源/CSRF 拒绝 403、冲突 409、存储不可用 503。列表默认 50、最大 100，以时间及请求 ID 稳定游标分页，不做全表下载。

仅 `GET /v1/models` 与 `POST /v1/chat/completions` 对应用开放；管理 cookie 不能替代 Bearer 应用密钥，应用 Bearer 不能替代管理 cookie。`GET /healthz` 仅给 ready/degraded，不能泄露配置；未接入上游仍可 ready（可管理），存储故障则 503。除静态页面外无隐式 SPA API 回退，未实现路径 404，方法不符 405。

配置写事务与请求快照使用同一进程读写锁：提交并释放写锁后，之后取得读锁的请求必见新状态；请求一次读取启用密钥、上游、别名及对应配置版本，然后释放锁。在途请求持有不可变快照，不换地址、不因停用追溯取消。使用一套事实源，无长期认证缓存或第二配置副本；配置变更不需要后台调和。

## 4. 协议与一次转发契约

### 4.1 接入与兼容清单

上游配置为完整 base URL，如 `https://provider.example/v1`，服务仅追加固定 `/chat/completions`，不从客户端拼接地址或发现模型。`GET /v1/models` 根据本地启用上游和别名返回 `{"object":"list","data":[{"id":"alias","object":"model","owned_by":"model-relay"}]}`，无可路由模型为空数组，不查询供应商。

生成入口须 `Content-Type: application/json`（允许 charset），单个 JSON 对象、无尾随内容/重复对象键，`model` 非空别名、`messages` 非空数组、`stream` 为布尔（省略为 false）。`messages` 仅支持 system/developer/user/assistant/tool 角色，文本 content；assistant 工具调用允许 content 为 null，tool 须有 tool_call_id。拒绝数组形式多模态内容，不下载消息中 URL。

| 字段 | v0.1 行为 |
|---|---|
| `model` | 查本地映射后替换为真实模型名 |
| `messages` | 保留 role/content/name/tool_calls/tool_call_id；校验文本和工具结构后透传 |
| `temperature`, `top_p`, `max_tokens`, `max_completion_tokens`, `n`, `stop`, `presence_penalty`, `frequency_penalty`, `seed`, `logit_bias`, `logprobs`, `top_logprobs`, `user` | 校验 JSON 类型；token 限额和 n 为正整数、stop 为字符串/字符串数组；保留原值交供应商判断模型能力 |
| `tools`, `tool_choice`, `parallel_tool_calls`, `response_format` | 支持 function 工具及其 JSON Schema、auto/none/required 或指定 function；响应格式 text/json_object/json_schema；不执行工具，不重写 schema |
| `stream`, `stream_options.include_usage` | 支持 SSE；include_usage 仅允许布尔、只用于 stream=true，省略不自动添加，避免改变厂商行为 |
| 其他顶层或消息控制字段 | 转发前 400 `unsupported_field`，包括客户端 base_url/headers/api_key、旧 functions/function_call、音频/图片扩展；无静默删除 |

对象字段通过 `json.RawMessage` 保留嵌套数值/工具 schema，而不是反序列化为浮点后重新生成；仅替换 model。协议声明是中转接纳与透传范围，不保证任意供应商/模型支持这些参数。普通 200 必须 JSON 且含合法 choices，解析验证和用量提取后原响应透传；响应中的真实模型名、choices/tool_calls/扩展字段原样保留，不假称响应 model 已改回别名。非 200 一律走错误契约。

请求鉴权后在本地校验，随后创建持久记录，再只执行一次供应商 POST。不发送客户端 cookie、Authorization、Host、代理头、幂等头或自定义 URL 控制信息；服务仅生成 Content-Type、Accept、供应商 Authorization 和自有请求 ID。返回头仅白名单 Content-Type、自有 X-Request-ID、Cache-Control，绝不转发 Set-Cookie 或上游任意头。

使用专用 `http.Transport`：`Proxy=nil`、禁止重定向、`GetBody=nil`、不标记请求为幂等；M1 禁用连接复用及自动 HTTP/2，避免底层重用连接触发重试。地址解析/Dial 只选择一个经过校验的 IP，无候选地址重试。测试必须计数证明错误、取消、超时及重定向场景最多一次生成。此承诺只覆盖本次中转处理，不能保证客户端自行重试不产生重复消耗，也不保证超时意味着供应商未计费。

### 4.2 普通错误和 SSE

所有流前错误采用 `{"error":{"message":"固定脱敏说明","type":"relay_error","code":"稳定类别"},"request_id":"…"}`。非法/停用密钥统一 401 `unauthorized`；非法请求 400；请求过大 413；未知/禁用别名 404 `model_unavailable`；上游未配置/禁用 503 `upstream_unavailable`。供应商 400～499（429 除外）映射 502 `upstream_rejected`，429 映射 429 `upstream_rate_limited`；5xx、网络失败、重定向或协议异常映射 502 对应类别；等待超时 504。仅记录上游状态码，不返回/存储其原始错误 body、URL 或供应商诊断文本。

SSE 200 必须 `text/event-stream`，按 UTF-8 事件空行边界增量解析，支持 LF/CRLF、多 data 行和注释；接受 data 中合法 chat chunk（包含 choices，包括 usage-only 空 choices）或 `[DONE]`。不支持自定义 event/id/retry 厂商扩展，收到即协议失败。每个完整合法 chunk 当即 write+flush；注释可透传但不算生成完成。只缓冲单个事件，不聚合全流，工具 delta 原样保留。

首个合法数据事件之前不发送 200，错误以普通 JSON 非 2xx 返回。开始之后失败发送 `event: error` 加脱敏 JSON data（作为 Relay 自身错误扩展），flush 后关闭，不发送 `[DONE]`；写连接已断时仅关闭。上游 error 对象、畸形事件、越界事件、超时或缺少 `[DONE]` 的 EOF 均不能成功。正常 `[DONE]` 在终态记录落盘后才转发并立即关闭；不用等待供应商关闭 socket。浏览器显式解析 error 和 `[DONE]`，其他客户端必须以缺失 `[DONE]` 判不完整；客户端已收到 finish_reason 也不能替代正常流结束。

### 4.3 有限资源与取消

| 边界 | 默认规则/验收方式 |
|---|---|
| 入站 | 生成 body 最大 1 MiB，管理 JSON 最大 64 KiB，头最大 32 KiB；header 5 秒、body 10 秒读期限 |
| 上游 | Dial 5 秒、TLS 5 秒、响应头 30 秒；整个请求最大 300 秒，SSE 首事件/有效数据事件之间最长 60 秒 |
| 响应 | 普通最大 8 MiB；SSE 单事件最大 256 KiB、全流最大 32 MiB；超限协议错误并取消，不无限积压 |
| 写与存储 | 下游每次 write/flush 最大 10 秒（逐次更新 deadline）；SQL 等待最大 2 秒，不将全局短 WriteTimeout 用于长流 |
| 取消/停止 | 客户端断开/AbortController 传播 context 并关闭上游 body；测试上游在 5 秒截止前观察到取消；SIGTERM 最多排空 10 秒，随后取消余下请求 |

超时以类型分类而非匹配上游错误字符串；慢客户端用同步读写背压和有限缓冲，不启动无界 goroutine/队列。测试可注入短 deadline，但须验证与生产默认走同一逻辑。未声明吞吐 SLA，不引入应用额度或 M2 并发治理。

## 5. 信任边界、凭据与地址

管理员浏览器 → 同源 Go 管理 API；应用 → Bearer 代理 API；relay → 固定配置供应商；运维 CLI → 本机数据库/加密文件。上游响应和全部用户输入不可信。高风险路径是来源伪造、XSS、管理身份混用、DNS rebinding/重定向、变更地址继承凭据及日志泄漏，按以下规则关闭。

- 默认 `127.0.0.1:5544`，本次指南仅支持 loopback 监听，不通过公网发布；实际 Host 必须匹配配置监听 authority，拒绝任意 Host，忽略 X-Forwarded-*。浏览器管理写操作（包括登录）要求 Origin 精确匹配服务 origin；缺失 Origin、null、不匹配均 403，不以 Referer/省略校验作兜底。管理脚本必须明确发送 Origin，测试覆盖缺失来源不改变状态。
- 登录后随机 256-bit session，服务内存存摘要；cookie HttpOnly、SameSite=Strict、Path=/admin，HTTPS 时 Secure，本地 HTTP 明示其限制。认证写操作还要求会话绑定的随机 CSRF token（session 查询返回）和 JSON Content-Type。网页 CSP 限制 self 脚本/连接、禁止 frame、object 和内联脚本；不允许跨源 CORS，应用程序无需 CORS 可正常 HTTP 调用。
- 登录以服务全局滑动窗口最多 10 次/分钟（成功/失败都计数），超限 429 和 Retry-After；限制器不依赖可伪造代理 IP。超过 2 个并行密码验证立即 429，避免 Argon2 内存放大。密码采用 Argon2id、随机 16-byte salt、64 MiB、3 轮、并行度 1、32-byte 输出；管理员密码由 init 生成 32 随机字节，不从命令行明文参数输入。
- 应用密钥为 `mr_` 加 32-byte 随机秘密，SHA-256 摘要存库、常量时间比对；随机高熵秘密不采用口令 KDF。单独 ID/非认证短标识用于记录与展示，不提供完整值查询；日志不输出签发响应。上游 token 使用 AES-256-GCM，新 nonce、AAD 绑定固定上游身份及规范 base URL；读接口只返回 `credential_configured`。
- URL 限制 HTTPS、无 userinfo/query/fragment、无反斜杠/路径穿越/编码分隔符，端口默认 443，可显式配置合法端口；base path 规范化固定，保存前校验。拒绝 localhost、loopback、private、link-local、unspecified、multicast、reserved/非公共范围、IPv4-mapped 绕过和元数据地址；所有 A/AAAA 都须符合策略，混合公私地址整组拒绝。
- 保存时解析并校验；每次实际 Dial 再解析校验全部结果，将选定公共 IP 直接交 Dialer，不再让系统做第二次名字解析；TLS SNI/证书验证仍使用原始 hostname，不能关闭校验。禁止环境代理和所有重定向，含同源重定向；3xx 不访问目标。解析器可注入用于 DNS 变换测试。
- base URL 规范值变化时必须同时提交新的非空 token，不允许旧密文/留空自动继承；保持相同 URL 且 token 留空才保留原凭据。初次配置必须有 token，禁用不删除密文；异常不能改变原配置。
- 隔离集成测试可通过程序启动参数显式注入 loopback 测试目的地策略；仅放行指定测试 authority、允许其 HTTP，不放行整个私网、不读取环境开关。正式服务默认没有该例外；CLI 不将例外写入持久上游状态。即使例外，也必须重新解析和禁止重定向。

本节点不新增合规认证或企业角色。loopback 不代表自动可信；同机恶意进程或拥有数据库及主密钥的运维者仍能访问秘密，此风险由本机权限与管理员保管承担。

## 6. 存储、调用终态和恢复

SQLite 为唯一事实源；Go `database/sql` 配合 `modernc.org/sqlite`，单进程排他运行锁、单连接、外键约束、WAL、synchronous=FULL、有限 busy 等待。事务中不得等待上游或持有流。所有列表有界，M1 不自动删除记录/引入保留平台，指南提示磁盘容量责任。

最小结构：安装元数据含 schema 版本与主密钥加密校验哨兵；admin 单行密码摘要/KDF 参数；upstream 单行 base URL、凭据 nonce/密文、enabled、revision；models 以 alias 为主键，真实模型名、enabled；app_keys 以随机 ID 为主键，名称、digest、短标识、enabled、创建时间；requests 以服务生成请求 ID 为主键，时间、应用 ID 可空、别名/真实模型快照（未能合法解析时可空）、stream、配置 revision、state、脱敏 error_code、upstream_status、耗时、可空用量。禁止保存 prompt、completion、原始 payload、全量响应或 URL/token。名称/别名长度有限，非法输入不原样写日志。

调用生命周期：鉴权 → 校验与配置快照 → durable started → 一次 HTTP 转发 → durable 终态 → 交付普通响应/最终流标记。有效应用密钥的本地拒绝（模型、配置、字段、大小等）写 `rejected`，不访问上游；未鉴权请求不进调用表、不编造应用身份，仅记自有 ID 与固定 unauthorized 类别的最小运行日志。模型列表不是生成，不建调用记录。body 限额必须在鉴权后使用有界读取，不能先全量读。

终态为 `success`、`rejected`、`upstream_rejected`（包括限流）、`upstream_failed`、`timeout`、`canceled`、`stream_interrupted`、`process_interrupted`。流中上游错误/异常 EOF 为 stream_interrupted，超时/客户端取消优先保持 timeout/canceled，附固定失败类别区分位置。只有普通合法成功响应或流合法 `[DONE]` 才可 success；success 表示供应商完成且 relay 已具备交付结果，不是客户端确认收取。若此时客户端已取消则 canceled；最终 write 失败尽力更新 canceled/下游写失败，不宣称网络可提供事务式送达保证。

用量逐字段存 nullable 整数 `prompt_tokens`、`completion_tokens`、`total_tokens`；缺失/null/无效值为未知，不填 0、不相加估算、不累计 SSE 重复快照。保存最近收到的合法 usage 快照，已有部分字段允许单独未知；断流若已收真实 usage 则保留并仍标记失败，完全没有为未知。非负整数 0 为真实 0。普通/SSE 成功 payload 中 usage 原样交付，不拿存储摘要替换供应商结果。

开始记录落盘失败：503 storage_unavailable，不发供应商请求。终态落盘失败：普通结果尚未提交时改 503；已开流则 error+关闭、不发 `[DONE]`。取消的收尾使用独立 2 秒 context，不能复用已取消的请求 context。发出不含 payload 的 storage_write_failed 日志并将服务置 degraded，后续生成 fail-closed，healthz 503、网页提示重启修复；不将仅内存记录当落盘成功。若数据库本身失效，无法保证最后错误记录写入，必须如实保留这个失败边界。

启动持运行锁后，在事务中将上次残留 started 转成 process_interrupted，保留已知用量、耗时未知而非伪造；失败则拒绝启动，绝不自动重发。SQL 迁移版本初始 1，在事务内执行，未知更高版本或迁移失败退出且保留原数据，禁止删库重建。

CLI 公开入口：`model-relay init --data-dir <dir> --key-file <file>`、`serve --data-dir <dir> --key-file <file> --listen 127.0.0.1:5544`、`backup --data-dir <dir> --output <file>`、`restore --input <file> --data-dir <empty-dir> --key-file <file>`；运行指南使用相对示例/占位路径，不提交本机绝对路径。数据目录 0700，数据库、独立主密钥及备份 0600；init 排他创建并拒绝非空安装，缺失密钥的 serve/restore 不生成替代密钥。init 输出密码不是运行日志，不写入源码或证据。

备份仅支持服务停止后，获得同一排他锁、检查点并关闭 DB 后创建 Go tar 归档（数据库与版本/摘要 manifest）；不复制运行中的 SQLite 主文件冒充一致备份。主密钥不进该归档，管理员需另外保管对应 key-file，数据库仍含身份摘要且备份须按敏感材料保管。恢复拒绝已有数据、校验归档路径/大小/摘要/版本、验证密钥能解密已配置凭据，再从临时目录原子安装；无上游时通过初始化保存的加密校验哨兵验证 key-file。归档不含会话、锁或日志；失败不得替换原安装。测试检查错误/缺失主密钥、损坏归档和运行中备份拒绝。不提供 M4 在线备份或升级回滚系统。

## 7. 技术选择与代价（已按委托采用）

- Go 标准库 HTTP/CLI/embed，最低项目工具链推荐 Go 1.27.1（2026-10-05 官方下载索引可核对）；设计环境实际为 Go 1.25.6，未据此声称目标版本已构建。研发在 go.mod 指定工具链基线，验证记录实际版本，受限环境不能获得工具链时保留阻塞证据，由已配置 Connector 验证，不暗中降级声明。
- SQLite/pure-Go driver 优于 JSON 文件（事务和多类持久状态需一致）与外部数据库（当前无需额外服务）；代价是单机故障域和单写入吞吐，退出路径为保留应用层 SQL 接缝、未来独立迁移，当前不做多存储适配器。选择 driver 及 x/crypto 的兼容稳定版本写入 go.mod/go.sum，禁止浮动 @latest 验证；具体锁定版本及 license 收据由研发留存，不虚构已经下载的依赖。
- 简单同源网页优于 SPA 构建工具，降低运行依赖；浏览器自动化可用锁版本 Playwright 作为纯测试依赖，不进入后端运行依赖。测试锁文件进入版本管理，node_modules、浏览器报告、秘密、数据、运行锁和缓存加入 .gitignore。
- 不自动重试/故障转移，错误由应用/管理员决定后续动作，避免未知计费副作用。终态记录失败不继续报告成功，代价是数据库故障会阻止调用，符合真实记录承诺。

## 8. 开发顺序与验收映射

| 实现切片 | 交付与受控验证 | 对应 AC |
|---|---|---|
| 安装/存储/身份 | Go CLI、版本/锁、加密、登录退出、Origin/CSRF/限速、默认空态 | 01～03、06、16～17 |
| 配置/模型/密钥 | 真实管理页面及 API，事务/停用、单次展示、地址校验、改 URL 必换凭据 | 04～07、15 |
| 普通代理 | 认证隔离、模型列表、兼容字段、一次转发、响应/错误/用量/记录 | 08～09、12～15 |
| SSE/取消 | event parser、逐事件 flush、DONE/错误、上下文与限额 deadline | 10～12、14 |
| 运行闭环 | 记录/UI 测试、窄屏、停止重启、备份恢复、指南、verify | 13、16～19 |

研发在 `delivery.md` 记实际实现路径、选型偏差、命令/环境/退出码与遗留限制；独立 QA 在 `qa.md` 复核代码版本及逐条 AC，Connector 阶段维护 `pr.md`。本设计不预写未执行阶段报告。

`make verify` 为唯一固定门禁：非改写式 gofmt 检查 → `go mod verify` → `go vet ./...` → Go 单元/真实 HTTP 集成测试（建议含 race，支持平台须失败即非零）→ `go build` → 实际 Go 服务 UI 浏览器旅程。任一步失败整体非零；禁止捕获错误后继续输出通过，不修改可信 Connector 门禁。浏览器缺依赖/环境无法监听必须失败并给原因，不能 skip 后整体绿灯。Go 程序负责启动实际服务、隔离数据与随机端口及回收；浏览器 runner 只执行 UI 自动化。

HTTP 测试使用临时目录和 httptest 本地供应商，显式测试目的地策略，正式管理 API 配置后调用，不能靠预填业务 DB 伪造测试。覆盖正确透传、两模型两密钥、启停、认证隔离、同源/缺 Origin、URL/DNS 变化/混合 IP/IPv6/3xx、变更凭据、请求/响应上限、错误内容类型、usage 缺失/部分/0、慢读写、取消、超时和错误脱敏。少量低层单元测试允许注入失败存储/解析器；不能修改已生成业务记录制造验收。

SSE 实时性用 channel 屏障：供应商首包 flush 后阻塞，客户端收到首包才释放屏障；不是比较两个易抖动时间戳。取消测试等供应商 context.Done 并在 5 秒截止前断言；每个故障断言生成计数最多 1。流中失败须断言无 `[DONE]`；记录写入故障须断言 fail-closed 与 health 503；服务异常退出/重启须看到 process_interrupted 而非 success。

浏览器旅程必须连接运行中的 Go 服务，桌面及 390px 窄屏完成登录、配置、两个别名、两把密钥、普通/SSE/取消、记录、停用、刷新和退出；断言表单错误、未保存不成功及浏览器 storage 无秘密。截图/trace 避免一次展示密钥和对话原文，凭据哨兵仅在受控测试内生成；失败截图也不得泄密。秘密扫描覆盖日志、管理响应、DB 记录和浏览器持久存储。重启/备份测试走正式 CLI 与 HTTP，验证恢复后原应用密钥能用、停用仍拒绝，不读用户环境或付费供应商。

真实回执需命令、退出码、完整脱敏输出、工具链版本及代码标识（HEAD 加工作树内容摘要；存在未提交差异时不能仅写 HEAD），测试证据保存在 `docs/workflow` 对应阶段或其安全附件。若沙箱无法监听，保存实际失败，不申请扩大权限，交已配置 host-only 验证 Connector；独立 QA 复核真实回执与相同代码，不要求 Agent 在沙箱重复监听。没有回执就不能通过 AC-19。

## 9. G2 自查、风险与交接

- 边界：仅 M1/v0.1，单管理员/单上游/单实例；全部后端与运维 Go，前端由 Go 提供，原创；未读取 New API 实现。
- 完整性：用户路径、模块/权威命令、对象生命周期、存储与版本、来源隔离、地址安全、兼容清单、实时流/取消、失败类别、记录失败、备份材料均有明确实现及验收路径，覆盖 AC-01～19。
- 简化：不引入组织治理、抽象渠道平台、消息队列、假供应商默认值或静态成功原型；alias 沿用业务键，只为密钥/请求的独立生命周期保留 ID。
- 证据门槛：本节点只做设计，无 Makefile/Go 应用可运行，未执行 make verify、HTTP 或浏览器测试，不补造通过证据。G2 为文档自查就绪，不代替后续真实门禁。
- 非阻塞未知：供应商扩展、usage 与具体模型参数兼容尚未实测；依赖/工具链及浏览器安装需开发环境确认；纯 Go SQLite 单连接在实际负载下的延迟未测。不承诺吞吐/可用率；这些未知通过受控测试、锁版本和“供应商联调未测”管理，不扩大当前范围。
- 剩余风险：取消不能保证供应商撤销费用，流失败不能撤回已交付增量，磁盘故障无法保证最后记录持久化，同机特权者可读主密钥；产品以明确状态、fail-closed、敏感备份保管和不自动重放控制影响。
- 方向性阻塞：无，无需退回 requirements。后续遇到协议边界实质矛盾应返回修订，而非静默删 AC。交接 development 实现并获得真实 make verify 回执，随后独立 QA，最后由 Connector 提交草稿 PR；本节点不切分支/提交/推送/部署。

## 10. 官方参考与采用边界

以下为 2026-10-05 查阅的协议/运行时原始资料，不是供应商联调证据；本文限制是 Model Relay 自有契约，不宣称官方保证中转安全或全部兼容。

- [R1] OpenAI Chat API：`https://developers.openai.com/api/reference/resources/chat`；参考 messages、工具字段、流和 include_usage；本产品只支持第 4 节明确子集。
- [R2] WHATWG SSE：`https://html.spec.whatwg.org/multipage/server-sent-events.html`；参考 UTF-8、data 行、空行事件边界。浏览器用 fetch 解析 POST 流，不使用只能 GET 的 EventSource 自动重连。
- [R3] Go net/http：`https://pkg.go.dev/net/http`；参考 Transport 重试条件、CheckRedirect、context、MaxBytesReader、ResponseController deadline/flush，设计明确限制潜在重放。
- [R4] Go 官方版本索引：`https://go.dev/dl/?mode=json`；查询 stable go1.27.1，区分推荐工具链和本机实际 go1.25.6。
- [R5] modernc SQLite driver：`https://pkg.go.dev/modernc.org/sqlite`；参考 database/sql pure-Go 驱动，需研发锁定兼容依赖，不宣称本节点已构建。
- [R6] SQLite backup：`https://www.sqlite.org/backup.html`；参考一致备份约束；M1 选择停止后排他备份，不使用运行中裸复制。
- [R7] Go Argon2：`https://pkg.go.dev/golang.org/x/crypto/argon2`；管理员口令 KDF；高熵应用密钥采用不可逆摘要而非存储明文。
- [R8] OWASP SSRF Prevention：`https://cheatsheetseries.owasp.org/cheatsheets/Server_Side_Request_Forgery_Prevention_Cheat_Sheet.html`；参考公私地址验证、DNS 与重定向风险，本项目以实际 Dial 重新解析/固定 IP 并禁代理约束实现。
