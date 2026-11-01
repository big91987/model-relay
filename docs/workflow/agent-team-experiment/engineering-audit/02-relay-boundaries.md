# Relay 模块边界静态工程审查

## 范围、基线与证据状态

- 审查基线：`122c9abece28f562616483e991098d020d236d6d`（当前 `HEAD`）。三项前置候选均已合入：严格 Bearer 候选 `67f464494f4d59be8a20ef99d9b15175ecd00e78`、合入 `2d48ee08f12368e1b806fcc0975943e2eca0c8ee`；单次时钟推进候选 `318639d819165f3309f123133fdaf6229003270c`、合入 `532e713acc998162c682b310c8f280e1651e6f95`；严格备份清单候选 `7dc57094557b8883c130a41ed306d1223c0d92f8`、合入 `122c9abece28f562616483e991098d020d236d6d`。
- 已验证：只读 Git、静态搜索和代码阅读；材料限于 `cmd/model-relay/*.go`、`internal/relay/*.go`、`internal/relay/web/`、`go.mod`，测试只读取名称、包声明和少量调用点来判断耦合。
- 未执行：编译、`go test`、`go vet`、服务启动、浏览器、真实供应商、备份恢复演练。因此本文的调用/依赖事实是“已验证”，运行时竞争、性能和故障结果均只标为“推断”；不继承前置任务的运行结果作为本次实测。
- 结论：未发现仅凭本范围静态证据可定为 P0 的问题。当前产品是一个可工作的单 Go 包实现，但 `internal/relay` 是部署单元也是事实上的集成边界，不是由编译器强制隔离的子系统集合；主要维护风险来自共享 `Server`/`Store` 状态及裸 SQL，而非目录命名。

## 轻量职责地图

```text
cmd/model-relay/main.go
  ├─ init/open/serve ─> Store ─> SQLite + key/installation lock
  ├─ backup/restore ─> backup.go ─> Store open/validation
  └─ upgrade ────────> upgrade.go ─> backup + schema/console/routing validation

HTTP ServeHTTP
  ├─ /healthz
  ├─ /admin/api/* ─> session/CSRF ─> identity + console + billing/reports
  ├─ /v1/models  ──> application-key authentication/authorization
  ├─ /v1/chat/completions
  │    └─ auth → protocol → snapshot → accounting admission
  │             → candidate/attempt persistence → upstream proxy/SSE
  │             → health/latency observation → settlement
  └─ /admin/* ─────> embedded web assets ─> the same admin/generation APIs

Shared in-process state
  Server.mu: effective-clock entry ordering + configuration/routing/health/latency
  Server.authMu: sessions + login/invitation throttling
  Store: DB + encryption key + schema + effective clock
```

入口和进程生命周期集中在 `cmd/model-relay/main.go:23-146`，其中 CLI 直接调用升级、初始化、备份、恢复和 `relay.New`（`main.go:47-118`）。HTTP 总入口集中于 `internal/relay/server.go:124-183`。这使公开入口容易定位；代价是所有后端职责最终汇入同一个 `relay` 包和同一 `Server` 对象。

## 客观发现

### P1 — “每个外部请求恰好一次”时钟契约仍有可见例外

**已验证。** `ServeHTTP` 对所有非健康请求在统一入口持有 `server.mu` 并调用 `advanceClock`，注释明确声明一次（`internal/relay/server.go:132-146`）。但 `/admin/api/invitations/accept` 仍由同一入口转入 `acceptInvitation`（`server.go:210-212`），后者再次取得 `server.mu` 并调用 `advanceClock`（`internal/relay/identity.go:434-438`）。因此该请求至少触发两次 `Store.advanceClock`；每次都会开启事务、读取并更新 `console_clock` 后提交（`internal/relay/console_clock.go:62-98`）。

此外，生成请求在入口推进后，计量准入 `Store.admit` 会再次推进（`internal/relay/accounting.go:127-132`），每次候选尝试 `beginAttempt` 也会推进（`internal/relay/routing.go:171-175`）。后两者可以解释为准入/派发的独立事务时点，但当前注释没有表达这个分层契约。

**推断的影响：** 邀请接受必然增加一次 SQLite 写事务；生成链路的推进次数随尝试数增长。维护者若依照入口注释删除或新增下层推进，可能改变配额、回拨高水位和故障关闭语义。用户侧主要风险是写锁竞争和存储故障面扩大，不是已证明的功能错误。

**建议（仓库流程，未完成）：** 把契约写成“每个请求入口一次，另列允许的事务时点”，逐一列出 invitation/admission/attempt；为邀请接受增加计数回归，决定它是应删除的重复推进还是有意的二次时点。若有意，改正 `ServeHTTP` 注释与前置审计口径。优先级 P1。

### P1 — 单包和裸存储句柄使职责边界不能由编译器约束

**已验证。** `internal/relay` 的后端文件和测试均使用 `package relay`；`Store` 公开 `DB`、`Key`、`Schema`（`internal/relay/store.go:21-30`），`Server` 又公开持有 `*Store` 和完整 `Config`（`internal/relay/server.go:18-53`）。认证直接查 `Store.DB`（`internal/relay/console.go:146-163`、`console_keys.go:390-424`），路由快照及尝试直接查写数据库（`routing.go:143-164,286-329`），生成编排直接查候选表并调用 `Start`/`Finish`/`admit`（`proxy.go:19-169`），控制台、身份和计量也各自创建事务。`go.mod:1-10` 只有 SQLite、crypto 及间接系统依赖，外部依赖面小，但内部 SQL/schema 依赖面很宽。

**用户/维护影响：** schema、密钥表达、时钟或事务语义改变时，影响不会停在 `store.go`；可能波及认证、路由、计量、控制台、升级、恢复和白盒测试。由于没有接口或子包方向，循环概念依赖可继续增长，代码评审只能靠约定发现越界。

**建议（仓库基建，未完成）：** 不建议一次性拆包。先以窄接口收口最高风险能力：凭据查验、路由快照/尝试仓储、计量准入/结算、有效时钟；让 `*sql.Tx` 只在存储/领域仓储层流动。再用依赖检查（例如禁止 HTTP 文件直接引用 `Store.DB/Key` 的仓库脚本）作为渐进门禁。优先级 P1。

### P1 — 一个全局互斥锁跨越 HTTP、数据库和路由运行态

**已验证。** `Server.mu` 同时保护入口时钟排序（`server.go:132-143`）、控制台/身份事务（`console_api.go:29-36`、`identity.go:130-132`）、生成快照与准入（`proxy.go:129-157`）、候选选择与证据事务（`routing.go:171-175,286-319`）以及健康/延迟内存态（`routing.go:333-368`）。配置变更会整体清空 routing/latency 并递增 epoch（`internal/relay/console.go:21-25`）。这是清晰的一致性策略，但锁的所有权不是单一职责。

**推断的影响：** 较慢的 SQLite 事务会阻塞无关管理请求、时钟入口或路由选择；扩大任一控制台操作时容易扩大全服务等待半径。现有 `TryLock` 仅保留排队时的墙钟采样，不消除串行化。

**建议（工程工具，未完成）：** 先增加锁等待/持有时间的无敏感信息观测及静态锁顺序说明；用已有实际 HTTP 并发场景建立基线，再考虑拆为配置版本锁、路由运行态锁和短时钟提交边界。未有运行数据前不建议直接细分锁。优先级 P1。

### P2 — 认证/授权在生成路径被多次读取，边界虽防御性但容易漂移

**已验证。** `generate` 先调用 `identifyKey`，协议解析后调用 `authorizeModel`，随后锁内再次 `appID`，准入中又以 `generationIdentity` 和授权 SQL 重验（`proxy.go:29-31,119-137,154-157`；`accounting.go:136-145`）。Schema 3 的 `identifyKey` 本身先查询 generation，再委托 `controlledAppID`（`console.go:146-163`）；精确 Bearer 结构位于 `controlledAppID`（`console_keys.go:390-398`）。这是 TOCTOU 防护所需的重复确认的一部分，但“解析凭据”“身份”“授权”“事务内重验”没有类型化结果区分。

**影响：** 合法安全修复需要同时审阅多个调用点；只改一个解析入口可能造成 `/v1/models`、生成前检和事务内检验不一致。前置 Bearer 回归已降低当前格式漂移风险，但未形成编译期边界。

**建议（通用平台，未完成）：** 引入不可导出、不可伪造的认证结果和显式事务内 `Revalidate` 方法；保留事务内重验，不把它误删为重复工作。优先级 P2。

### P2 — 有效业务时钟、运行时截止和路由测量是三个时钟域，但命名/注入边界不完整

**已验证。** `Config.Clock` 被写入 `store.clock`（`server.go:55-58`），`Store.now` 返回持久化高水位而 `wallTime` 才调用注入时钟（`store.go:32-45`）。会话过期、登录限速、网络截止和耗时仍直接使用 `time.Now/time.Since`（例如 `server.go:185-198,294-312`、`proxy.go:19-25,48,78-100`）；路由延迟另有 `latencyClock`（`server.go:51-52`）。

**影响：** 这种区分在语义上合理：业务窗口不应等同单调耗时/网络截止。但新增逻辑时，维护者难以从 `Clock` 名称判断应使用哪个域，单元验证也需要包内改字段。

**建议（工程工具，未完成）：** 以 `EffectiveClock`、`WallClock`、`Monotonic/Deadline` 命名并写一页使用规则；只对业务判断注入有效/墙钟，对超时保留 Go 单调时间。优先级 P2。

### P2 — 嵌入式 Web 的文件隔离良好，API 合约仍与大包集成验证绑定

**已验证。** `web.go` 只负责嵌入和固定路径/Content-Type 映射（`internal/relay/web.go:9-64`），静态资产不接触数据库，是合理隔离点。`workspace.js` 直接拼接 `/admin/api/` 并直接消费 session、控制台、生成及 SSE 返回（例如 `internal/relay/web/workspace.js:73,236,271,278`）；HTTP 入口决定 schema 3 使用哪组资产（`server.go:175-180`）。

**影响：** 资产服务函数可小范围验证，但 UI/API 字段漂移只能通过契约或真实服务旅程发现；手工维护路径映射还会使新增资产同时修改 Go 和 Web。

**建议（仓库基建，未完成）：** 从后端响应类型/固定 JSON 样例生成或校验前端契约，并为嵌入资产建立最小 handler 测试；无需为此引入 Node 后端。优先级 P2。

## 合理封装和已有隔离点

- CLI 与后端包的公共面较窄：`main.go` 只编排 `InitConsole/Open/New/Backup/Restore/UpgradeConsole/ResetAdminPassword`，监听地址与优雅关闭留在命令层（`cmd/model-relay/main.go:41-143`）。
- 升级和恢复不是 HTTP handler 的附属分支。升级先锁定/打开安装、备份，再在事务中迁移并做完整性校验（`internal/relay/upgrade.go:110-154,189-219`）；恢复先严格解析 archive/manifest、校验数据库和 schema，最后 rename（`backup.go:140-230`）。这两个入口可在不启动服务时单独验证，虽然仍依赖共享 schema/open helpers。
- 严格 manifest 解码是纯输入边界（`backup.go:20-57`）；协议 JSON 校验函数集中在 `protocol.go`；加解密/地址策略集中在 `security.go`。这些是适合表驱动单元验证的天然隔离点。
- 路由将不可变请求快照、进程内健康/延迟态和持久化尝试证据显式建模（`routing.go:16-64,143-168`）；配置 epoch 和 execution revision 为并发配置变化提供了隔离机制。
- `adminResponse` 延迟管理响应的实际写出，使配置锁可在网络写前释放（`internal/relay/admin_response.go:8-31`）；这是明确且小型的 HTTP/锁边界。
- 嵌入式 Web 只经公开 HTTP API 工作，没有直接访问存储；Go 服务仍是唯一产品后端。

## 可独立验证程度与变更爆炸半径

| 职责 | 主要位置 | 可独立验证程度（静态判断） | 典型修改影响 |
|---|---|---|---|
| CLI/HTTP 入口 | `cmd/model-relay/main.go`, `server.go` | CLI `run` 可进程内验证；主路由可用 handler 验证 | 命令合同、监听安全、所有 HTTP 路径 |
| 协议与安全纯函数 | `protocol.go`, `security.go`, `backup.go:20-57` | 高：字节/值输入输出 | 请求兼容、SSRF/密钥、恢复输入边界 |
| 会话与应用认证 | `server.go`, `console.go`, `console_keys.go`, `identity.go` | 中低：多数需真实 SQLite；会话还需 `Server` 内存态 | admin、models、generation、邀请 |
| 路由/代理 | `routing.go`, `routed_proxy.go`, `proxy.go`, `routing_latency*.go` | 中：选择算法可白盒；提交/SSE/故障切换需 handler 或实际 HTTP | 上游调用、重试、证据、健康、延迟、计量 |
| 控制台/身份/账单/报告 | `console_*.go`, `identity.go`, `billing.go` | 低至中：共享事务、principal、schema、`Server.mu` | 权限、审计、计量、Web 合约 |
| 计量 | `accounting.go`, `legacy_accounting.go` | 中：金额纯函数可单测；准入/结算需 SQLite + 请求链 | 配额、幂等、账单、最终请求状态 |
| 时钟 | `console_clock.go`, `Store.now/wallTime` | 中：Store 可注入墙钟；请求次数需 handler + DB 观察 | 所有非健康请求、配额、邀请、路由尝试 |
| 持久化 | `store.go`, `console_storage.go`, `routing_latency_storage.go` | 低：裸 SQL 被全包直接使用 | 几乎所有后端职责、升级/恢复 |
| 升级/备份 | `upgrade.go`, `backup.go` | 高于服务路径：无需 HTTP；仍需真实临时 SQLite/文件系统 | 运维恢复、schema、密钥材料 |
| 嵌入式 Web | `web.go`, `web/*` | 资产 handler 高；用户旅程低，必须实际 Go 服务 | admin API、SSE、浏览器兼容 |

`internal/relay/*_test.go` 均使用同包 `package relay`（命令测试为 `package main`），且许多 relay 测试直接替换 `server.client.Transport`、读取 `server.health` 或调用未导出方法；这提供了细粒度白盒覆盖能力，但不是模块可替换性的证据。测试名称显示已有成对的 in-process/actual-HTTP 场景、恢复/升级、真实进程崩溃、时钟回归和资源矩阵；只有 HTTP/CLI/文件系统边界能证明跨职责组合，纯函数测试不能替代这些路径。本次没有执行任何一个测试。

## 建议归属汇总

- **仓库流程 / P1：** 明确三类时钟推进点及允许例外，补邀请接受的推进次数回归；代码注释、设计说明和回归断言必须一致。
- **仓库基建 / P1：** 渐进收口 `Store.DB/Key`，增加依赖方向门禁；不要以一次性目录重排冒充解耦。
- **工程工具 / P1：** 观测 `Server.mu` 等待/持有时间并记录锁顺序，以数据决定拆锁。
- **通用平台 / P2：** 用类型化认证结果与事务内重验接口统一 Bearer 身份/授权边界。
- **工程工具 / P2：** 命名并约束有效时钟、墙钟和单调截止三个域。
- **仓库基建 / P2：** 为嵌入 Web 建立后端响应契约和最小资产 handler 检查。

以上均为建议，不表示已经实现、验证或排期。
