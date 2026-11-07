# 代码与架构规范检查

- 规则扫描：未发现规则违规（仅限已声明的规则）。
- 模型评审：发现 5 项待复核的架构问题。

规则检查依据已声明的依赖约束；模型意见需复核。本报告不判定产品业务功能通过或失败。

## 规则违规

声明范围内未发现规则违规。

## 模型架构发现（待复核）

### 1. internal/relay 单包混合 HTTP、SQL 与业务规则（高）

**涉及位置与源码对象**

- `internal/relay/accounting.go:127` · `admit`
- `internal/relay/billing.go:15` · `tenantBills`
- `internal/relay/console_api.go:179` · `manageModelFinance`

**问题依据**

module_responsibilities、business_rule_ownership、saas_responsibilities：规则应归能力 owner，HTTP 只适配用例，存储实现 operation 级原子性而不是承载业务判断。

**影响**

准入、限额、账单费用和价格授权规则分散在 handler 与 SQL 中，调整协议或存储会牵连业务判断；规则难以独立复用和验证，形成维护与边界风险。

**调整建议**

以现有 Store.admit、usageFor、tenantBills、manageModelFinance 为切分点：把准入/限额/账单费用规则移到独立领域或应用包，HTTP handler 只解析请求与响应，SQL 与事务保留在存储适配器；不新增空层。

**复查建议**

复查 internal/relay 业务文件的 database/sql、net/http 依赖及 handler 到 SQL 的直接调用；以现有集成测试作为行为保持护栏。

**原文证据**

`internal/relay/accounting.go:127` · `admit`

    func (store *Store) admit(record Record, snapshot routeSnapshot, payload map[string]json.RawMessage, request *http.Request) (string, int, string) {

`internal/relay/billing.go:15` · `tenantBills`

    func (server *Server) tenantBills(writer http.ResponseWriter, request *http.Request, transaction *sql.Tx, principal Principal, tenantID, role string) {

`internal/relay/console_api.go:179` · `manageModelFinance`

    func (server *Server) manageModelFinance(writer http.ResponseWriter, request *http.Request, transaction *sql.Tx, principal Principal, alias, resource string) {

### 2. 路由与上游协调直接耦合 HTTP 请求和 SQL（高）

**涉及位置与源码对象**

- `internal/relay/routing.go:171` · `beginAttempt`
- `internal/relay/routing.go:288` · `beginAttempt`
- `internal/relay/routed_proxy.go:116` · `routedGenerate`

**问题依据**

gateway_upstream_separation 与 dependency_direction：routing 只拥有策略和运行时状态，上游 adapter 拥有网络/协议实现，策略不应依赖 HTTP 或 SQL。

**影响**

路由选择与持久化、HTTP 请求处理交织，替换上游协议或改变存储会使 routing 变更；无法通过清晰接口替换或模拟上游适配器。

**调整建议**

让 routing 的选择、健康、探索状态不依赖 net/http 与 database/sql，通过仓库接口持久化选择证据；定义 upstream adapter 接口，由 HTTP 实现负责请求构造、响应解析和流协议，routedGenerate 只协调 attempt。

**复查建议**

检查 routing 包导入和 beginAttempt 签名是否仍出现 net/http、database/sql；检查 routedGenerate 是否只调用 adapter 接口；保留现有路由集成测试。

**原文证据**

`internal/relay/routing.go:171` · `beginAttempt`

    func (server *Server) beginAttempt(ctx context.Context, request *http.Request, record *Record, snapshot routeSnapshot, tried map[int64]bool, round int) (*Candidate, *Attempt, string) {

`internal/relay/routing.go:288` · `beginAttempt`

    transaction, err := server.Store.DB.BeginTx(cleanup, nil)

`internal/relay/routed_proxy.go:116` · `routedGenerate`

    outgoing, err := http.NewRequestWithContext(output.child, http.MethodPost, upstream.URL+"/chat/completions", bytes.NewReader(body))

### 3. Server 共享可变状态横跨会话、路由、健康与延迟（中）

**涉及位置与源码对象**

- `internal/relay/server.go:47` · `Server`
- `internal/relay/server.go:49` · `Server`
- `internal/relay/console_api.go:29` · `consoleAPI`

**问题依据**

state_ownership：session、request、routing、health 和业务状态应有明确 owner，不相关能力不应依赖同一共享可变对象。

**影响**

不同能力共享 Server 与同一互斥锁，console/config 操作会与路由、健康、延迟状态纠缠；测试和并发边界难以按能力隔离。未测量运行时影响，仅构成架构风险。

**调整建议**

把 sessions、routing runtime/health、latency pool、checking 等状态移入各自拥有的结构或服务，Server 仅做组合与依赖注入；console handler 通过对应服务访问，而非直接操作共享 map。

**复查建议**

复查 Server 字段与 server.mu 的使用范围，确认状态访问经由能力 owner；以现有配置、路由和 console 集成测试保持行为。

**原文证据**

`internal/relay/server.go:47` · `Server`

    routing            map[string]*routeRuntime

`internal/relay/server.go:49` · `Server`

    health             map[string]*routeHealth

`internal/relay/console_api.go:29` · `consoleAPI`

    server.mu.Lock()

### 4. 前端各业务页共享单一全局状态与渲染分支（中）

**涉及位置与源码对象**

- `internal/relay/web/workspace.js:15`
- `internal/relay/web/workspace.js:238` · `renderPage`
- `internal/relay/web/app.js:4`

**问题依据**

frontend_features：功能应拥有业务 UI、交互状态、业务 API 与清理；共享传输保持业务中立，公共入口约束依赖。

**影响**

workspace.js 与 app.js 的全局状态和集中 renderPage/navigate 逻辑让页面交互状态、请求和清理互相牵连，增加跨 feature 回归风险。

**调整建议**

按现有 pages/renderPage 分支拆出 feature 模块，各自持有状态、业务 API 调用和清理逻辑；保留 api()/fetch 作为业务中立的共享传输层，入口模块只组合 feature。

**复查建议**

复查每个 feature 是否只通过公共传输与自身状态通信，页面切换时是否由 owner 清理自己的 controller/secret；以现有浏览器旅程作为行为保持护栏。

**原文证据**

`internal/relay/web/workspace.js:15`

    const state = {user:null,csrf:'',tenant:'',tenants:[],page:'overview',epoch:0,controllers:new Set(),busy:false,call:null,invitation:'',filters:{},pageNumber:1};

`internal/relay/web/workspace.js:238` · `renderPage`

    async function renderPage() {

`internal/relay/web/app.js:4`

    let csrf = '', active = 'overview', nextCursor = null, navigationVersion = 0, sessionEpoch = 0;

### 5. 业务规则只能通过 SQL 事务或 HTTP 请求验证（中）

**涉及位置与源码对象**

- `internal/relay/accounting.go:82` · `usageFor`
- `internal/relay/accounting.go:73` · `limitsFor`

**问题依据**

testability：规则与策略应能通过模块接口验证，外部行为保留集成测试接缝，而不是把规则埋入存储查询。

**影响**

usageFor、limitsFor 等规则需要 *sql.Tx，admit 还需要 *http.Request；规则级验证必须构造数据库/HTTP 夹层，降低可替换性和精确回归能力。

**调整建议**

将资源使用、限额判断和准入策略提取为接收普通值并返回决策的领域函数；存储适配器负责加载数据和事务写入，HTTP handler 负责输入输出。

**复查建议**

复查规则函数是否可在无 database/sql、net/http 依赖下调用；现有 HTTP/数据库集成测试作为行为保持护栏。

**原文证据**

`internal/relay/accounting.go:82` · `usageFor`

    func usageFor(transaction *sql.Tx, ownerType, ownerID, period, minute string) (ResourceUsage, error) {

`internal/relay/accounting.go:73` · `limitsFor`

    func limitsFor(transaction *sql.Tx, ownerType, ownerID string) (ResourceLimits, error) {


## 逐项评审与限制

- `module_responsibilities`：存在偏差。internal/relay 一个包内同时存在 HTTP handler、SQL、业务策略、加解密、备份/升级和静态资源服务；文件命名区分但不构成模块边界。例如准入与账单规则直接写在带 *http.Request 或 *sql.Tx 的函数中。
- `business_rule_ownership`：存在偏差。费用、限额和准入规则以 SQL 和 handler 分支形式存在于 internal/relay，没有独立的能力拥有者；tenantBills 在 HTTP handler 内完成查询和 fee 调整，admit/usageFor 在存储层计算规则。
- `dependency_direction`：存在偏差。routing.go 直接依赖 *http.Request 和 *sql.Tx；routed_proxy.go 直接构造上游 HTTP 请求并解析协议。策略/用例依赖协议和存储细节，方向反转。
- `state_ownership`：存在偏差。Server 集中持有 routing、health、latency、sessions、checking 等不同能力状态，console/config handlers 也通过同一 server.mu 访问；没有按能力拆出显式 owner。
- `testability`：存在偏差。usageFor/limitsFor 等规则以 *sql.Tx 查询实现，admit 还需要 *http.Request；规则无法通过纯模块接口验证，单元验证需要数据库或 HTTP 夹层。
- `necessary_abstractions`：符合。已存在 materialQuery、Resolver、rowScanner、identityQuery 等小接口用于测试/解析接缝；未观察到空层或通用引擎。缺少上游 adapter 的问题归入 gateway_upstream_separation。
- `standards_alignment`：符合。architecture.json 声明模块与 review_checks，文档明确当前规则通过不等于目标职责划分已达成，没有用描述现状来豁免要求。
- `saas_responsibilities`：存在偏差。规则丰富的计费/限额/准入能力没有独立领域包；HTTP adapters 直接执行 SQL 和业务分支，基础设施与用例未按 operation 分离。事务虽存在，但职责归属不符合要求。
- `gateway_request_lifecycle`：符合。generate/routedGenerate 以及 attemptOutput 集中处理请求/尝试、取消、超时、输出提交和重试；静态审查只能确认这些职责位置，不能证明运行时正确性。
- `gateway_upstream_separation`：存在偏差。routing 同时依赖 HTTP 请求和 SQL，routedGenerate 直接实现上游 HTTP 与协议解析；没有显式的 upstream adapter 接口来承载网络/协议实现。
- `frontend_features`：存在偏差。workspace.js 用单一全局 state 和 renderPage 的大分支承载所有业务页；app.js 也用全局可变状态承载会话、导航和调用。业务交互状态、API 与清理未按 feature 归属。
- 范围说明：本次为静态架构审查，仅基于提供的快照、源码和文档；未运行 make architecture-check、服务或测试，不据此判断运行时行为。
- 范围说明：coverage 中的测试文件和静态资源未作为逐行架构证据；文档用于边界与标准对齐，不用于证明实现缺陷。

## 执行记录与检查范围

工具版本：0.4.1
Git HEAD：42b79a00d3ff7b8ed8957f87a97fd5b4c83abe64
源码摘要：d67816a9746ed4996cfd26820d018c3f895c21df130386423c37d6dbe1dab28b

```json
{
  "scanned": [
    "cmd/model-relay/main.go",
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
    "internal/relay/identity.go",
    "internal/relay/legacy_accounting.go",
    "internal/relay/protocol.go",
    "internal/relay/proxy.go",
    "internal/relay/routed_proxy.go",
    "internal/relay/routing.go",
    "internal/relay/routing_latency.go",
    "internal/relay/routing_latency_storage.go",
    "internal/relay/security.go",
    "internal/relay/server.go",
    "internal/relay/store.go",
    "internal/relay/upgrade.go",
    "internal/relay/web.go",
    "internal/relay/web/app.js",
    "internal/relay/web/workspace.js"
  ],
  "excluded": [
    "cmd/model-relay/deployment_test.go",
    "cmd/model-relay/main_test.go",
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
    "internal/relay/version_test.go"
  ],
  "assets": [
    "internal/relay/web/index.html",
    "internal/relay/web/style.css",
    "internal/relay/web/workspace.css",
    "internal/relay/web/workspace.html"
  ],
  "context": [
    "AGENTS.md",
    "architecture.json",
    "docs/engineering/architecture-checks.md",
    "docs/roadmap.md",
    "go.mod",
    "go.sum",
    "package-lock.json",
    "package.json"
  ]
}
```
