# 代码与架构规范检查

- 规则扫描：未发现规则违规（仅限已声明的规则）。
- 模型评审：发现 4 项待复核的架构问题。

规则检查依据已声明的依赖约束；模型意见需复核。本报告不判定产品业务功能通过或失败。

## 规则违规

声明范围内未发现规则违规。

## 模型架构发现（待复核）

### 1. internal/relay 单包把控制台 HTTP、事务 SQL 与计费/密钥规则耦合在同一 handler（高）

**涉及位置与源码对象**

- `internal/relay/console_api.go:92` · `manageLimits`
- `internal/relay/billing.go:110` · `manageBill`
- `internal/relay/console_api.go:322` · `manageModelFinance`
- `docs/engineering/architecture.md:27`

**问题依据**

module_responsibilities、business_rule_ownership、saas_responsibilities、testability；architecture.json 声明 relay 模块，docs/engineering/architecture.md 要求复杂操作保持事务边界，但不要求把规则留在 HTTP 层。

**影响**

价格发布、限额、账单和密钥轮换等规则变更需要同时修改 HTTP、SQL 和事务编排；规则无法通过应用服务接口独立验证，跨用例复用困难，事务语义散落在 handler 中。

**调整建议**

在 internal 下按用例提取应用服务（至少覆盖 limits、pricing、billing、key lifecycle），由服务接收普通值并返回决策或结果；SQL 写入封装为仓储或 store 操作，HTTP handler 只做协议映射和调用。保留现有事务范围、幂等键和审计写入，不要为简单 CRUD 造空层。

**复查建议**

重新检查 console_api.go、billing.go、console_keys.go 的 handler 是否不再直接执行业务 SQL；用现有 HTTP 集成测试与 make verify 作为事务和幂等行为保持的回归，不宣称本次执行。

**原文证据**

`internal/relay/console_api.go:92` · `manageLimits`

    func (server *Server) manageLimits(writer http.ResponseWriter, request *http.Request, transaction *sql.Tx, principal Principal, tenantID, role, ownerType, ownerID string) {

`internal/relay/billing.go:110` · `manageBill`

    func (server *Server) manageBill(writer http.ResponseWriter, request *http.Request, transaction *sql.Tx, principal Principal, parts []string) {

`internal/relay/console_api.go:322` · `manageModelFinance`

    if price.FeeIn != input.Prices[0].FeeIn || price.FeeOut != input.Prices[0].FeeOut {

`docs/engineering/architecture.md:27`

    relay 内的控制台 HTTP、应用操作与 SQL 仍有混合。这属于后续分步治理范围，不能因当前导入规则通过而宣称所有 SaaS 领域分层完成。

### 2. Server 聚合会话、限流、上游核验和网关配置状态，缺少独立状态所有者（中）

**涉及位置与源码对象**

- `internal/relay/server.go:36-52` · `Server`
- `docs/engineering/architecture.md:21`

**问题依据**

state_ownership；docs/engineering/architecture.md 承认会话与上游检查状态尚未整体拆出。

**影响**

登录与邀请限流、会话、upstream check 和 gateway 配置及成功证据共享 server.mu/authMu；任一能力扩展都会扩大共享锁和耦合，状态遗漏与锁竞争风险只能在运行时暴露。

**调整建议**

把 sessions、passwordSlots、attempts、invitationAttempts 拆为认证会话与限流所有者；checking 拆为上游核验注册表；successID、configEpoch 归配置协调器；Server 仅保留组合入口，并为这些所有者定义窄接口。

**复查建议**

检查登录、邀请、upstreams/check、generate 的字段访问是否只经各自接口；保留现有 HTTP 与并发回归测试作为行为保持，不宣称已执行。

**原文证据**

`internal/relay/server.go:36-52` · `Server`

    type Server struct {
    	Store              *Store
    	Config             Config
    	mu                 sync.RWMutex
    	authMu             sync.Mutex
    	sessions           map[string]session
    	attempts           []time.Time
    	invitationAttempts []time.Time
    	passwordSlots      chan struct{}
    	degraded           atomic.Bool
    	client             *http.Client
    	configEpoch        uint64
    	successID          string
    	router             *routing.Engine
    	checking           map[int64]bool
    	latencyClock       func() time.Time
    }

`docs/engineering/architecture.md:21`

    原有共享配置锁仍保护请求与配置的协调。本轮不以结构调整声称已消除锁竞争；会话与上游检查状态尚未整体拆出。

### 3. 上游 HTTP/协议实现直接嵌入 routedGenerate 和 checkUpstream，缺少显式适配接口（中）

**涉及位置与源码对象**

- `internal/relay/routed_proxy.go:116` · `routedGenerate`
- `internal/relay/routed_proxy.go:141` · `routedGenerate`
- `internal/relay/configuration.go:631` · `checkUpstream`
- `internal/relay/configuration.go:577` · `checkUpstream`

**问题依据**

gateway_upstream_separation；routing 负责选择策略与运行状态且禁止 HTTP/SQL，上游适配器应拥有网络与协议实现。

**影响**

网关协调代码同时决定 URL 拼接、请求头、响应 Content-Type 与 JSON/SSE 解析；生成和核验各自构造请求，协议变化会扩散到 gateway 逻辑，难用适配器替身验证。

**调整建议**

定义上游适配接口（如 ChatCompletion 与 ListModels/Check），由 relay 内的上游适配器实现 HTTP/SSE 细节；routedGenerate 和 checkUpstream 只依赖接口与归一化结果，routing 保持无 HTTP/SQL。

**复查建议**

重新检查 internal/routing 导入，以及 routed_proxy.go、configuration.go 是否不再直接构造上游请求；现有 HTTP/SSE 与上游核验回归作为行为保持。

**原文证据**

`internal/relay/routed_proxy.go:116` · `routedGenerate`

    outgoing, err := http.NewRequestWithContext(output.child, http.MethodPost, upstream.URL+"/chat/completions", bytes.NewReader(body))

`internal/relay/routed_proxy.go:141` · `routedGenerate`

    response, err := server.client.Do(outgoing)

`internal/relay/configuration.go:631` · `checkUpstream`

    outgoing, err := http.NewRequestWithContext(ctx, http.MethodGet, value.URL+"/models", nil)

`internal/relay/configuration.go:577` · `checkUpstream`

    func (server *Server) checkUpstream(writer http.ResponseWriter, request *http.Request, id int64) {

### 4. workspace.js 仍集中多数管理页业务 UI、交互状态和业务 API（中）

**涉及位置与源码对象**

- `internal/relay/web/workspace.js:16`
- `internal/relay/web/workspace.js:232` · `requestDetail`
- `docs/engineering/architecture.md:33`

**问题依据**

frontend_features；playground 已按功能公开入口拆分，但 architecture.md 仅声明调用测试这一条功能边界，其他管理页仍集中在 workspace.js。

**影响**

records、models、keys、billing、limits 的页面状态和 API 调用共享模块级 state 与 invalidate/clearSensitive；修改一个页面需理解全局状态，功能清理边界只对 playground 成立。

**调整建议**

按管理页提取功能模块（如 records、models/services、keys、billing/limits），每个功能拥有自身 DOM、交互状态、业务 API 与 dispose；workspace.js 保留租户范围、导航和组合，shared/transport 继续业务中立。

**复查建议**

检查 workspace.js 是否只依赖功能公开入口、各功能是否自行持有状态与清理；保留现有前端与浏览器回归作为行为保持，不宣称已执行。

**原文证据**

`internal/relay/web/workspace.js:16`

    const state = {user:null,csrf:'',tenant:'',tenants:[],page:'overview',epoch:0,busy:false,invitation:'',filters:{},pageNumber:1};

`internal/relay/web/workspace.js:232` · `requestDetail`

    async function requestDetail(id) { const value=await api(scoped('workspace/requests/'+encodeURIComponent(id))); const controls=[];

`docs/engineering/architecture.md:33`

    其他管理页仍集中在 workspace.js，旧版 app.js 也未重构；本轮只验证调用测试这一条功能边界，不宣称前端已整体模块化。


## 逐项评审与限制

- `module_responsibilities`：存在偏差。architecture.json 将 internal/relay/*.go 定义为一个 relay 模块，但该包同时承载网关请求循环、控制台 HTTP、身份与租户、计费、密钥生命周期、备份升级与 SQLite 事务；模块内没有按用例或能力拆分职责，导出面也较大。文档已承认混合，但导入规则通过不能等同于职责内聚。
- `business_rule_ownership`：存在偏差。accounting 包拥有金额与准入基元，但价格发布、账单累计、限额管理和密钥轮换等规则仍直接写在 HTTP handler 中并与 SQL 事务耦合；同一业务规则难以从模块接口单独验证，存在规则归属和重复边界不清的风险。
- `dependency_direction`：符合。routing 与 accounting 只导入标准库，未导入 net/http 或 database/sql；workspace.js 仅通过公开的 shared/transport 与 features/playground 入口依赖前端功能；architecture.json 的 forbid_imports 与源码导入一致。
- `state_ownership`：存在偏差。Server 把会话、登录与邀请限流、上游核验中标志、配置代次、成功证据、路由引擎和 HTTP client 放在同一可变对象；auth、admin 与 gateway 路径通过 server.mu/authMu 共享访问，状态所有权不清晰，且文档已承认会话与上游检查状态尚未整体拆出。
- `testability`：存在偏差。routing 与 accounting 有可经包接口验证的 Select、CheckAdmission 等入口，HTTP 也有集成 seam；但 billing、limit 与 key 规则被锁在接收 *sql.Tx 的 handler 中，若要按规则验证只能走完整 HTTP+SQL 事务，缺少应用服务层 seam。
- `necessary_abstractions`：符合。现有接口窄且有当前职责：materialQuery 与 identityQuery 只抽象查询，routing.Engine 直接持有算法和运行状态，未发现为形式而造出的空层或通用引擎。
- `standards_alignment`：符合。架构文档与 architecture.json 对包边界和禁止导入的描述一致，并明确规则通过不等于职责分层或产品验收完成；未把当前缺口表述为已满足目标标准。
- `saas_responsibilities`：存在偏差。限额、价格、账单和密钥等规则丰富的用例由 HTTP handler 直接开启或接收 *sql.Tx 并执行 SQL，未形成独立应用用例与仓储边界；事务虽存在，但应用规则和基础设施实现未分离。
- `gateway_request_lifecycle`：符合。请求/尝试循环、尝试输出承诺、取消/超时和重试上限集中在 routedGenerate、attemptOutput 与 stream；路由包不持有 HTTP 生命周期。静态评审只确认职责归属，不判断运行时正确性。
- `gateway_upstream_separation`：存在偏差。routing 包不含 HTTP/SQL，符合策略与运行状态归属；但网关没有显式上游适配接口，routedGenerate 与 checkUpstream 直接构造 /chat/completions、/models HTTP 请求并解析响应，网络与协议实现和协调代码混在 relay。
- `frontend_features`：存在偏差。playground 已作为功能模块并由 workspace.js 通过公开入口使用，transport 也保持业务中立；但 workspace.js 仍集中 records、models、keys、billing、limits 等页面的业务 UI、业务 API 调用和共享 state，功能边界只覆盖调用测试。
- 范围说明：本次为静态架构评审，未运行 make architecture-check、make verify、浏览器或真实上游调用；不对运行时正确性、锁竞争或超时/取消效果下结论。
- 范围说明：审阅范围为 architecture.json 指定的生产源文件与上下文文档；测试文件及 web 静态资产（index.html、style.css、workspace.html、workspace.css）未纳入内容判断。
- 范围说明：architecture.json 的规则状态按输入报告为 conformant；该结果只说明导入规则通过，不代表职责分层或产品验收通过。
- 范围说明：前端仅审阅 workspace.js、shared/transport.mjs、features/playground/index.mjs 与旧 app.js 的模块边界；HTML/CSS 资产未审阅。

## 执行记录与检查范围

工具版本：0.4.1
Git HEAD：42b79a00d3ff7b8ed8957f87a97fd5b4c83abe64
源码摘要：d698c922cac97afea3c81b4044805dcc7ad092c0d0a13f6a1ae3a75d8b84c117

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
