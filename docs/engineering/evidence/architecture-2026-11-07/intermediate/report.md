# 代码与架构规范检查

- 规则扫描：未发现规则违规（仅限已声明的规则）。
- 模型评审：发现 3 项待复核的架构问题。

规则检查依据已声明的依赖约束；模型意见需复核。本报告不判定产品业务功能通过或失败。

## 规则违规

声明范围内未发现规则违规。

## 模型架构发现（待复核）

### 1. internal/relay 将计费、限额与密钥治理用例与 HTTP、SQL 混合，并重复实现模型授权规则（高）

**涉及位置与源码对象**

- `internal/relay/billing.go:15` · `tenantBills`
- `internal/relay/accounting.go:65` · `admit`
- `internal/relay/console_keys.go:418` · `authorizeModel`
- `internal/relay/accounting.go:79` · `admit`
- `internal/relay/console_keys.go:421` · `authorizeModel`
- `docs/engineering/architecture.md:27`

**问题依据**

saas_responsibilities：规则丰富的 SaaS 能力应由应用用例承载、独立领域包拥有业务规则、基础设施实现操作级原子存储；business_rule_ownership：规则应归属于其所属能力而非 HTTP/SQL 或重复用例；module_responsibilities：模块须内聚并暴露小型显式接口。依据 docs/engineering/architecture.md 与 architecture.json 的模块声明。

**影响**

账单关闭/更正、用量核对、准入与结算既有 HTTP 处理又有内联 SQL 与判定，规则归属分散；key_model_grants/tenant_model_grants 授权谓词在 console_keys 与 accounting.admit 各写一份，修改一处易遗漏另一处，增加维护成本与不一致风险（属静态判断，非运行期已证故障）。

**调整建议**

最小变更：把 billing.go 的 tenantBills/manageBill/reconcileUsage、accounting.go 的 admit/settleAdmission、console_keys.go 的 key 生命周期中与传输无关的判定抽为接受普通值与 *sql.Tx 的应用用例函数（保留原有事务边界），并将 authorizeModel 与 admit 内联的授权谓词统一为单一实现，由 console_keys.authorizeModel 与准入流程共同调用；不新增空层，不改变产品规则。

**复查建议**

复检 internal/relay 是否仍存在同一授权谓词的多份 SQL 与处理函数内联的计费规则；以现有 make verify 的 HTTP 集成回归作为行为保持手段，不新增业务验收条件。

**原文证据**

`internal/relay/billing.go:15` · `tenantBills`

    func (server *Server) tenantBills(writer http.ResponseWriter, request *http.Request, transaction *sql.Tx, principal Principal, tenantID, role string) {

`internal/relay/accounting.go:65` · `admit`

    func (store *Store) admit(record Record, snapshot routeSnapshot, payload map[string]json.RawMessage, request *http.Request) (string, int, string) {

`internal/relay/console_keys.go:418` · `authorizeModel`

    func (store *Store) authorizeModel(keyID, alias string) (bool, string) {

`internal/relay/accounting.go:79` · `admit`

    "SELECT count(*) FROM key_model_grants k JOIN key_controls c ON c.key_id=k.key_id JOIN tenant_model_grants g ON g.alias=k.alias AND g.tenant_id=c.tenant_id WHERE k.key_id=? AND k.alias=?"

`internal/relay/console_keys.go:421` · `authorizeModel`

    "SELECT t.active,(SELECT count(*) FROM key_model_grants k JOIN tenant_model_grants g ON g.alias=k.alias AND g.tenant_id=c.tenant_id WHERE k.key_id=c.key_id AND k.alias=?) FROM key_controls c JOIN tenants t ON t.id=c.tenant_id WHERE c.key_id=?"

`docs/engineering/architecture.md:27`

    relay 内的控制台 HTTP、应用操作与 SQL 仍有混合。这属于后续分步治理范围

### 2. Server 与 Store 聚合互不相关的可变状态：会话、限流、配置代次、上游检查与时钟（中）

**涉及位置与源码对象**

- `internal/relay/server.go:41-44` · `Server`
- `internal/relay/store.go:28-31` · `Store`
- `docs/engineering/architecture.md:21`

**问题依据**

state_ownership：会话、请求、路由、健康与业务状态应有明确持有者，无关能力不应依赖同一共享可变对象；依据 docs/engineering/architecture.md 对共享配置锁与会话/上游检查未拆分的说明。

**影响**

会话表、登录与邀请限流历史、密码并发槽、配置代次、成功证据与上游检查集合共用 Server 及其锁；Store 混淆数据库句柄与有效时钟。任一路径的锁持有范围变化都可能影响其他能力，扩展与测试时难以界定状态边界（属风险描述，未测量）。

**调整建议**

最小变更：将会话表与限流历史、上游检查集合、配置代次分别交由独立持有对象（如 sessionStore、rateLimiter、checkRegistry）管理，Server 仅保留组合与协调；Store 的 clockMu/effectiveTime/clockWarned 从数据库字段中分离为独立时钟持有者。保持现有对外行为不变。

**复查建议**

复检 Server/Store 字段与锁使用是否按能力收敛、是否存在跨能力共享同一互斥量；以现有集成测试作为行为保持手段。

**原文证据**

`internal/relay/server.go:41-44` · `Server`

    	sessions           map[string]session
    	attempts           []time.Time
    	invitationAttempts []time.Time
    	passwordSlots      chan struct{}

`internal/relay/store.go:28-31` · `Store`

    	clock         func() time.Time
    	clockMu       sync.RWMutex
    	effectiveTime time.Time
    	clockWarned   bool

`docs/engineering/architecture.md:21`

    本轮不以结构调整声称已消除锁竞争；会话与上游检查状态尚未整体拆出。

### 3. 上游网络与协议实现内联于网关包，缺少显式上游适配接口（中）

**涉及位置与源码对象**

- `internal/relay/routed_proxy.go:141` · `routedGenerate`
- `internal/relay/proxy.go:185` · `stream`
- `internal/relay/server.go:73` · `New`

**问题依据**

gateway_upstream_separation：网关通过显式接口协调，上游适配器拥有网络/协议实现；routing 拥有策略与运行状态且不含 HTTP/SQL。

**影响**

请求构造、Authorization 头、HTTP 客户端调用与 SSE 逐行解析直接写在 relay 的网关文件中，网关协调与上游协议细节耦合；替换或新增上游协议时需改动协调逻辑，协议实现也难以独立验证（属静态判断）。

**调整建议**

最小变更：在 relay 内引入显式上游适配接口（负责 URL 组装、认证头、发送与 SSE 帧解析），routed_proxy.go 与 proxy.go 通过该接口调用而非直接持有 server.client 与解析器；routing 保持无 HTTP。避免为算法新增层次。

**复查建议**

复检 routed_proxy.go/proxy.go 是否不再直接依赖 *http.Client 与 SSE 解析细节、适配接口是否被网关通过接口调用；以现有集成测试作为协议行为保持手段。

**原文证据**

`internal/relay/routed_proxy.go:141` · `routedGenerate`

    response, err := server.client.Do(outgoing)

`internal/relay/proxy.go:185` · `stream`

    scanner := bufio.NewScanner(response.Body)

`internal/relay/server.go:73` · `New`

    transport := &http.Transport{Proxy: nil, DialContext: config.Policy.Dial, DisableKeepAlives: true, ForceAttemptHTTP2: false, TLSHandshakeTimeout: 5 * time.Second, ResponseHeaderTimeout: config.HeaderTimeout}


## 逐项评审与限制

- `module_responsibilities`：存在偏差。internal/relay 作为单一模块同时承载 HTTP 协议适配、网关请求协调、控制台管理与 SQLite 持久化，并暴露大量以 *Server 方法形式存在的处理函数；职责只能靠文件名区分，未形成内聚模块与小型显式接口。项目文档亦承认 relay 内部职责混合。
- `business_rule_ownership`：存在偏差。模型/租户授权与额度类判定分别以不同 SQL 出现在 console_keys.authorizeModel、accounting.admit 内联查询以及 proxy.generate 分支中，属于跨位置的规则重复与耦合，而非由单一所属能力持有。
- `dependency_direction`：符合。依赖方向与声明一致：cmd → relay；relay → routing、accounting；routing 与 accounting 仅依赖标准库，未引入 net/http 或 database/sql；前端共享传输与 playground 无业务功能依赖。
- `state_ownership`：存在偏差。Server 单一对象聚合会话表、登录/邀请限流历史、并发密码槽、降级标志、配置代次、成功证据 ID、上游检查集合与路由引擎及延迟时钟；Store 同时持有数据库句柄与有效时钟状态，无显式归属划分。
- `testability`：符合。规则与策略可通过模块接口验证：routing.RecordLatency 通过 isCurrent 与 save 回调隔离配置校验与证据持久化，accounting 提供纯函数可单测，架构声明排除 *_test.go 且存在 HTTP 集成测试接缝（未执行）。
- `necessary_abstractions`：符合。现有接口各有当前职责：Resolver 提供 DNS 解析接缝，materialQuery 与 identityQuery 提供查询抽象以支持校验与测试，adminResponse 缓冲管理响应以保证事务后统一写出；未见需要四层化的冗余空层。
- `standards_alignment`：符合。工程规范与可执行声明互相一致，且对未完成的分层以显式披露而非宣称达标：architecture.json 声明约束，architecture.md 说明新增包须归属模块并保持复杂操作事务边界，AGENTS.md 明确规则通过不等于领域质量或产品验收。
- `saas_responsibilities`：存在偏差。对规则丰富的多租户计费、限额与密钥治理，HTTP 处理函数直接内联 SQL 与业务判定（账单关闭、更正、用量核对、准入结算），缺少独立的应用用例层；基础设施虽在事务内执行，但规则未与 HTTP/SQL 分离。
- `gateway_request_lifecycle`：符合。网关在本包内定义请求/尝试生命周期：routedGenerate 控制轮次与重试上限，attemptOutput 负责输出提交与超时取消；生命周期归属明确，但运行期正确性未经本次验证。
- `gateway_upstream_separation`：存在偏差。上游网络与协议实现未抽象为适配接口，而是与网关协调同处 relay 包：HTTP 客户端、SSE 解析与请求构造直接内联在网关文件，routing 虽已独立且无 HTTP，但适配层边界缺失。
- `frontend_features`：符合。调用测试功能自持 DOM、应用密钥请求、流接收与 dispose 清理，工作台仅通过公开入口创建并传递通知与查看记录回调；共享传输只承载范围绑定请求与取消，属业务中性。其他管理页集中属声明范围内，且不以缺少 Vue 作为缺陷。
- 范围说明：静态评审仅基于本次提供的源码与文档，未执行构建、测试或运行时验证，运行期行为一律按风险而非结论描述。
- 范围说明：被排除的 *_test.go 与前端静态资源（index.html、style.css、workspace.html、workspace.css）未作为评审证据；仅参考其存在性判断测试接缝。
- 范围说明：业务结果正确性（价格、限额阈值、权限决策）不在评审范围；本次只评估包职责、依赖方向、接口、规则归属、状态归属、测试接缝与抽象必要性。

## 执行记录与检查范围

工具版本：0.4.1
Git HEAD：42b79a00d3ff7b8ed8957f87a97fd5b4c83abe64
源码摘要：b1a00128c1a8ccb3137dacfeacddf9670eaf2555dc076bf3800a4c1793597a5c

```json
{
  "scanned": [
    "cmd/model-relay/main.go",
    "internal/accounting/rules.go",
    "internal/relay/accounting.go",
    "internal/relay/admin_response.go",
    "internal/relay/backup.go",
    "internal/relay/billing.go",
    "internal/relay/configuration.go",
    "internal/relay/console.go",
    "internal/relay/console_api.go",
    "internal/relay/console_clock.go",
    "internal/relay/console_keys.go",
    "internal/relay/console_overview.go",
    "internal/relay/console_reports.go",
    "internal/relay/console_storage.go",
    "internal/relay/console_summaries.go",
    "internal/relay/gateway_attempts.go",
    "internal/relay/gateway_latency.go",
    "internal/relay/gateway_output.go",
    "internal/relay/identity.go",
    "internal/relay/legacy_accounting.go",
    "internal/relay/protocol.go",
    "internal/relay/proxy.go",
    "internal/relay/routed_proxy.go",
    "internal/relay/routing_latency_storage.go",
    "internal/relay/security.go",
    "internal/relay/server.go",
    "internal/relay/store.go",
    "internal/relay/upgrade.go",
    "internal/relay/web.go",
    "internal/relay/web/app.js",
    "internal/relay/web/features/playground/index.mjs",
    "internal/relay/web/shared/transport.mjs",
    "internal/relay/web/workspace.js",
    "internal/routing/engine.go",
    "internal/routing/latency.go",
    "internal/routing/types.go"
  ],
  "excluded": [
    "cmd/model-relay/deployment_test.go",
    "cmd/model-relay/main_test.go",
    "internal/accounting/limits_test.go",
    "internal/accounting/money_test.go",
    "internal/relay/boundary_test.go",
    "internal/relay/configuration_test.go",
    "internal/relay/connection_lifecycle_test.go",
    "internal/relay/console_clock_queue_test.go",
    "internal/relay/console_clock_recovery_test.go",
    "internal/relay/console_contract_test.go",
    "internal/relay/console_http_test.go",
    "internal/relay/console_migration_test.go",
    "internal/relay/console_process_test.go",
    "internal/relay/console_resource_matrix_test.go",
    "internal/relay/console_review_regression_test.go",
    "internal/relay/console_summaries_test.go",
    "internal/relay/console_test.go",
    "internal/relay/console_v3_test.go",
    "internal/relay/crash_test.go",
    "internal/relay/health_test.go",
    "internal/relay/inprocess_test.go",
    "internal/relay/lifecycle_test.go",
    "internal/relay/metering_regression_test.go",
    "internal/relay/persistence_test.go",
    "internal/relay/protocol_role_test.go",
    "internal/relay/qa_connection_contract_test.go",
    "internal/relay/relay_test.go",
    "internal/relay/routing_boundary_test.go",
    "internal/relay/routing_contract_regression_test.go",
    "internal/relay/routing_http_test.go",
    "internal/relay/routing_latency_recovery_test.go",
    "internal/relay/routing_latency_test.go",
    "internal/relay/routing_test.go",
    "internal/relay/single_fixture_test.go",
    "internal/relay/sse_terminal_regression_test.go",
    "internal/relay/upgrade_test.go",
    "internal/relay/version_test.go",
    "internal/routing/engine_test.go",
    "internal/routing/latency_test.go"
  ],
  "assets": [
    "internal/relay/web/index.html",
    "internal/relay/web/style.css",
    "internal/relay/web/workspace.css",
    "internal/relay/web/workspace.html"
  ],
  "context": [
    ".trellis/spec/backend/index.md",
    "AGENTS.md",
    "architecture.json",
    "docs/engineering/architecture-checks.md",
    "docs/engineering/architecture.md",
    "docs/roadmap.md",
    "go.mod",
    "go.sum",
    "package-lock.json",
    "package.json"
  ]
}
```
