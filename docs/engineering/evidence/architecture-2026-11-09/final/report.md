# 代码与架构规范检查

- 规则扫描：未发现规则违规（仅限已声明的规则）。
- 模型评审：发现 4 项待复核的架构问题。

规则检查依据已声明的依赖约束；模型意见需复核。本报告不判定产品业务功能通过或失败。

## 规则违规

声明范围内未发现规则违规。

## 模型架构发现（待复核）

### 1. reporting 的可见范围谓词通过 SQL 文本替换复用到子查询（中）

**涉及位置与源码对象**

- `internal/reporting/usage.go:22` · `Usage`
- `internal/reporting/request_filters.go:54`
- `internal/reporting/scope.go:23` · `reportScope`

**问题依据**

business_rule_ownership：租户/开发者可见范围规则应由 reporting 的 reportScope 单一拥有（docs/engineering/architecture.md 规定 reporting 按当前人员、租户与诊断 Key 范围生成读模型），不应以字符串改写的形式再分发到各查询。

**影响**

internal/reporting/usage.go 的 ledgerSum 必须在运行时把 where 子句中的别名 c. 与 r. 改写成 ec. 与 er.，与该包共享常量 requestsFrom 的别名集形成隐式契约；只要查询别名或连接结构变化，作用域谓词与费用/用量聚合就可能不一致，属于静态无法验证的运行期风险。

**调整建议**

让 internal/reporting/scope.go 的 reportScope 与 internal/reporting/request_filters.go 的 requestScope 返回可指定别名的谓词构造（例如接受表别名参数并生成条件），使 ledgerSum 与 requestsFrom 的使用方不再调用 strings.ReplaceAll；范围规则的判定仍只留在 reportScope。

**复查建议**

复查 internal/reporting 内不再出现对查询文本的别名替换，并保留 session 范围判定只在 reportScope；以现有 internal/app/console_contract_test.go 等 HTTP 回归作为行为保持的保障。

**原文证据**

`internal/reporting/usage.go:22` · `Usage`

    strings.ReplaceAll(strings.ReplaceAll(where, "c.", "ec."), "r.", "er.")

`internal/reporting/request_filters.go:54`

    const requestsFrom = " FROM requests r LEFT JOIN key_controls c ON c.key_id=r.key_id LEFT JOIN admissions a ON a.request_id=r.id WHERE "

`internal/reporting/scope.go:23` · `reportScope`

    where := "c.tenant_id=?"

### 2. 受支持请求字段集在 admission 与 upstream 各维护一份（低）

**涉及位置与源码对象**

- `internal/admission/bounds.go:6` · `boundedPayload`
- `internal/upstream/protocol.go:120` · `ValidateRequest`

**问题依据**

business_rule_ownership：协议接受面归 upstream、可信预留子集策略归 admission（docs/engineering/architecture.md 明确「后者在有硬预留要求时只允许可证明上界的文本请求子集」），但字段知识不应有两份独立声明。

**影响**

internal/admission/bounds.go 的 boundedPayload 内联一份允许字段 map，而 internal/upstream/protocol.go 的 ValidateRequest 用另一份字符串清单判定协议字段；两者需要同步维护，扩展协议字段时容易出现协议层接受、预留层拒绝（或反之）的规则漂移风险。

**调整建议**

保留 admission 决定「有界文本子集」的策略所有权，但把字段集合改为来自单一声明（例如由 upstream 导出已支持字段的分类，再由 admission 选出其文本子集），删除 boundedPayload 中第二份手工字段 map；不改变子集判定结果。

**复查建议**

复查 admission 不再重复声明字段名清单、仍由 admission 负责子集策略；以 internal/app/console_contract_test.go 中既有「unsupported bounded request」断言作为行为保持保障。

**原文证据**

`internal/admission/bounds.go:6` · `boundedPayload`

    allowed := map[string]bool{"model": true, "messages": true, "stream": true, "stream_options": true, "max_tokens": true, "max_completion_tokens": true, "n": true, "temperature": true, "top_p": true, "seed": true, "stop": true}

`internal/upstream/protocol.go:120` · `ValidateRequest`

    if !onlyKeys(value, "model messages stream temperature top_p max_tokens max_completion_tokens n stop presence_penalty frequency_penalty seed logit_bias logprobs top_logprobs user tools tool_choice parallel_tool_calls response_format stream_options") {

### 3. 存储 schema 版本分支散布在 relay、records、apikey 与 httpconsole（中）

**涉及位置与源码对象**

- `internal/relay/proxy.go:30` · `Generate`
- `internal/records/store.go:46` · `Finish`
- `internal/apikey/gateway.go:28` · `Observe`
- `internal/httpconsole/auth.go:30` · `ServeHTTP`

**问题依据**

dependency_direction：选定 installation 为存储结构唯一归属后，版本兼容决策应由存储侧能力拥有，调用方不应各自解释 Schema 数值（AGENTS.md 与 docs/engineering/architecture.md 的分包要求）。

**影响**

internal/relay/proxy.go 用 server.Store.Schema 决定请求时间来源，internal/records/store.go 用 store.Schema 拼接状态条件，internal/apikey/gateway.go 的 Observe 与 internal/httpconsole/auth.go 也各自按版本分支；同一「schema3 行为」散落多处，relay 还必须持有具体 *installation.Store 才能读到版本，版本语义变化需跨包同步，静态无法保证一致。

**调整建议**

把版本相关差异下沉到拥有的能力：由 records 与 apikey 用自身具名的兼容实现替代调用方对 Schema 数值的比较，httpconsole 通过能力服务给出的结果分支，relay 只消费这些接口而不直接比较存储版本；保留现有 schema2 兼容行为与安装期校验。

**复查建议**

复查 internal/relay/*.go、internal/records/store.go、internal/apikey/gateway.go、internal/httpconsole/auth.go 不再直接比较 schema 数值；以 internal/records/legacy_test.go 与 console 契约测试作为行为保持保障。

**原文证据**

`internal/relay/proxy.go:30` · `Generate`

    record.Time = server.Store.Now().Format(time.RFC3339Nano)

`internal/records/store.go:46` · `Finish`

    query += " AND state='started'"

`internal/apikey/gateway.go:28` · `Observe`

    func (service *Service) Observe(credential string, schema int) (string, bool) {

`internal/httpconsole/auth.go:30` · `ServeHTTP`

    if controller.Schema == 3 && request.URL.Path == "/admin/api/invitations/accept" && request.Method == http.MethodPost {

### 4. 中断请求的启动态迁移由 installation 直接改写 requests/attempt 状态（低）

**涉及位置与源码对象**

- `internal/installation/store.go:225` · `openInstallation`
- `.trellis/spec/backend/index.md:3`

**问题依据**

module_responsibilities：请求与尝试证据责任归 records（.trellis/spec/backend/index.md 明确 Request/attempt evidence persists in internal/records），恢复编排可留在 installation，但请求状态迁移规则不应在两处实现。

**影响**

internal/installation/store.go 的 openInstallation 用原始 SQL 把 started 请求改写成 process_interrupted，而 internal/records/store.go 的 Finish/FinishAttempt 也以同样的 state='started' 条件约束终态写入；请求状态机因此有两个实现者，状态取值与迁移条件需要跨包同步。

**调整建议**

在 internal/records 增加一个启动恢复入口承担 started→process_interrupted 的迁移（含 attempts_complete、duration_ms 的清空规则），由 installation.openInstallation/recoverConsole 在既有独占锁与事务内调用，保持 DDL、锁、时钟与调用顺序仍归 installation。

**复查建议**

复查 installation 不再直接写入请求状态字段、records 提供该迁移函数，并保留安装期事务边界；以 internal/records/legacy_test.go、internal/app/console_clock_queue_test.go 等既有回归作为行为保持保障。

**原文证据**

`internal/installation/store.go:225` · `openInstallation`

    UPDATE requests SET state='process_interrupted',error_code='process_interrupted',duration_ms=NULL,attempts_complete=0 WHERE state='started'

`.trellis/spec/backend/index.md:3`

    Request/attempt evidence persists in internal/records; DDL, effective clock, recovery and installation commands belong to internal/installation.


## 逐项评审与限制

- `module_responsibilities`：存在偏差。整体按能力分包且职责与文档包地图一致（app 组装、httpconsole 适配、admission/finance/identity 拥有用例、upstream 协议、routing 策略），但请求状态迁移的实现分布在 records 与 installation 两处，见 findings。
- `business_rule_ownership`：存在偏差。多数规则落在拥有者（accounting 纯计算、routing 选择策略、access.Role 身份范围、finance 账单事务），但 reporting 以字符串替换复用可见范围谓词，且 admission 与 upstream 各自维护一份受支持请求字段清单，见 findings。
- `dependency_direction`：存在偏差。上游/协议方向正确（relay 通过 upstream.Client/PrepareChat/NewStream 协作，upstream 不依赖 SQL，routing 无 HTTP/SQL），但存储 schema 版本判断分散到 relay、records、apikey 与 httpconsole，使版本兼容知识离开存储拥有者，见 findings。
- `state_ownership`：符合。配置代次/成功证据/检查登记由 configstate.Coordinator 显式拥有并要求调用方持锁，会话由 sessions.Manager 拥有，路由运行态由 routing.Engine 拥有；适配层在持锁路径上不调用自行加锁的 catalog 方法（如 tenant-grants 在加锁前分流），协调契约未被破坏。
- `testability`：符合。纯规则与策略可通过模块接口验证（accounting、routing、apikey.GenerationIdentity 事务级函数），外部行为有可注入接缝：upstream.NewClient 接受 *http.Client、relay.Config 可替换时钟，app 测试通过替换 Transport 与本地监听覆盖 HTTP 路径；未声称已执行这些测试。
- `necessary_abstractions`：符合。小型接口都有当下用途：identityQuery 让授权与接纳共享查询或事务、records.Query 复用证据读取、materialQuery 只为加密材料校验、Resolver 只为目标解析测试替换；fault/names/jsoncodec/accounting 等纯规则包均有真实调用者，未见空层或通用引擎。
- `standards_alignment`：符合。architecture.json 的模块与 allow/forbid 声明、docs/engineering/architecture.md 包地图、.trellis/spec/backend/index.md 职责描述与实际包集合一致（app/httpconsole 禁止 SQL、能力服务禁止 net/http、routing/accounting/configstate 同时禁止 HTTP/SQL），文档未以「现状如此」豁免要求；本评审提出的偏离位于代码内部耦合，而非声明与标准冲突。
- `saas_responsibilities`：符合。规则密集能力各自拥有用例与事务：admission.Admit 在同一事务内完成再鉴权、价格快照、双范围额度检查与预留，finance 的账单/更正/补录保留完整事务，identity 的邀请与成员变更原子提交；httpconsole 只做解码、错误映射与响应，不写 SQL、不复制业务判断。
- `gateway_request_lifecycle`：符合。请求/尝试生命周期明确归 relay：重试轮次与尝试上限、总超时与尝试超时、取消分类、输出提交（attemptOutput.commitNormal）都在该包内，能力服务只被调用；本项仅确认归属，未推断运行期正确性。
- `gateway_upstream_separation`：符合。upstream 拥有目标安全策略、传输、请求准备与 JSON/SSE 协议，不决定重试、结算或输出提交；relay 只通过 Endpoint/PreparedChat/Stream 等显式接口协调；routing 以普通结构体输入输出策略与运行态，无 HTTP/SQL。
- `frontend_features`：符合。两个壳层（app.js、workspace.js）只做组装、会话与导航；各 features/* 通过自身 index.mjs 导出创建函数，拥有页面业务请求、交互状态与 dispose 清理，共享模块仅提供无业务取向的传输、集合读取、格式化与生命周期；跨功能协作由壳层注入而非相互读内部状态。
- 范围说明：静态架构评审：未运行 make verify、go test、浏览器旅程或 architecture-check 工具本身；architecture.json 的规则结果以输入中给出的 conformant 声明为证据，不重复执行。
- 范围说明：运行期属性（锁竞争、并发、超时、SSE 提交时序）只能作为风险描述，未实测；本评审仅判断代码组织、依赖方向、规则与状态归属、接口与测试接缝。
- 范围说明：assets（internal/webui/web/index.html、style.css、workspace.css、workspace.html）未在输入中提供，未评审；excluded 的测试文件仅评审了 context 中列出的四份（app/health_test.go、app/console_contract_test.go、app/console_clock_queue_test.go、records/legacy_test.go）。
- 范围说明：前端 .mjs/.js 的模块依赖图以工具 facts 为准（共享模块仅导入 shared，功能模块仅导入 shared 与同目录 presentation.mjs）。
- 范围说明：未判断任何业务结果正确性（价格、额度阈值、权限决策、计费金额），仅判断规则与职责落位。

## 执行记录与检查范围

工具版本：0.4.1
Git HEAD：064eb0b9c52d9b04a72ed3cd1dd70eeac756e58b
源码摘要：328ccea19f55b02826c37f3b48d89730e3153188fc695003288d1115b7a45fbe

```json
{
  "scanned": [
    "cmd/model-relay/main.go",
    "internal/access/access.go",
    "internal/accounting/rules.go",
    "internal/admission/bounds.go",
    "internal/admission/service.go",
    "internal/apikey/authorization.go",
    "internal/apikey/gateway.go",
    "internal/apikey/legacy.go",
    "internal/apikey/models.go",
    "internal/apikey/persistence.go",
    "internal/apikey/rules.go",
    "internal/apikey/service.go",
    "internal/app/server.go",
    "internal/catalog/catalog.go",
    "internal/catalog/check.go",
    "internal/catalog/grants.go",
    "internal/catalog/models.go",
    "internal/catalog/read.go",
    "internal/catalog/upstreams.go",
    "internal/configstate/coordinator.go",
    "internal/fault/fault.go",
    "internal/finance/admission_reads.go",
    "internal/finance/billing.go",
    "internal/finance/legacy.go",
    "internal/finance/limits.go",
    "internal/finance/prices_read.go",
    "internal/finance/pricing.go",
    "internal/finance/reconciliation.go",
    "internal/finance/rules.go",
    "internal/finance/service.go",
    "internal/finance/settlement.go",
    "internal/finance/types.go",
    "internal/httpapi/http.go",
    "internal/httpconsole/auth.go",
    "internal/httpconsole/catalog.go",
    "internal/httpconsole/controller.go",
    "internal/httpconsole/finance.go",
    "internal/httpconsole/identity.go",
    "internal/httpconsole/keys.go",
    "internal/httpconsole/reports.go",
    "internal/httpconsole/response.go",
    "internal/identity/acceptance.go",
    "internal/identity/access.go",
    "internal/identity/authentication.go",
    "internal/identity/invitations.go",
    "internal/identity/members.go",
    "internal/identity/rules.go",
    "internal/identity/service.go",
    "internal/identity/tenants.go",
    "internal/installation/backup.go",
    "internal/installation/clock.go",
    "internal/installation/console.go",
    "internal/installation/routing.go",
    "internal/installation/schema.go",
    "internal/installation/store.go",
    "internal/installation/upgrade.go",
    "internal/jsoncodec/json.go",
    "internal/names/names.go",
    "internal/operations/operations.go",
    "internal/records/legacy.go",
    "internal/records/models.go",
    "internal/records/store.go",
    "internal/relay/gateway.go",
    "internal/relay/gateway_attempts.go",
    "internal/relay/gateway_latency.go",
    "internal/relay/gateway_output.go",
    "internal/relay/proxy.go",
    "internal/relay/routed_proxy.go",
    "internal/reporting/audit.go",
    "internal/reporting/catalog.go",
    "internal/reporting/legacy_overview.go",
    "internal/reporting/overview.go",
    "internal/reporting/request_filters.go",
    "internal/reporting/requests.go",
    "internal/reporting/scope.go",
    "internal/reporting/service.go",
    "internal/reporting/summaries.go",
    "internal/reporting/tenants.go",
    "internal/reporting/usage.go",
    "internal/routing/engine.go",
    "internal/routing/latency.go",
    "internal/routing/types.go",
    "internal/secure/secrets.go",
    "internal/sessions/sessions.go",
    "internal/sqlread/read.go",
    "internal/upstream/client.go",
    "internal/upstream/policy.go",
    "internal/upstream/protocol.go",
    "internal/upstream/response.go",
    "internal/upstream/stream.go",
    "internal/webui/web.go",
    "internal/webui/web/app.js",
    "internal/webui/web/features/billing/index.mjs",
    "internal/webui/web/features/keys/index.mjs",
    "internal/webui/web/features/legacy-keys/index.mjs",
    "internal/webui/web/features/legacy-models/index.mjs",
    "internal/webui/web/features/legacy-overview/index.mjs",
    "internal/webui/web/features/legacy-playground/index.mjs",
    "internal/webui/web/features/legacy-records/index.mjs",
    "internal/webui/web/features/legacy-upstreams/index.mjs",
    "internal/webui/web/features/limits/index.mjs",
    "internal/webui/web/features/limits/presentation.mjs",
    "internal/webui/web/features/models/index.mjs",
    "internal/webui/web/features/overview/index.mjs",
    "internal/webui/web/features/playground/index.mjs",
    "internal/webui/web/features/reports/index.mjs",
    "internal/webui/web/features/services/index.mjs",
    "internal/webui/web/features/tenants/index.mjs",
    "internal/webui/web/shared/catalog.mjs",
    "internal/webui/web/shared/data.mjs",
    "internal/webui/web/shared/facts.mjs",
    "internal/webui/web/shared/legacy-data.mjs",
    "internal/webui/web/shared/legacy-presentation.mjs",
    "internal/webui/web/shared/lifetime.mjs",
    "internal/webui/web/shared/transport.mjs",
    "internal/webui/web/workspace.js"
  ],
  "excluded": [
    "cmd/model-relay/deployment_test.go",
    "cmd/model-relay/main_test.go",
    "internal/accounting/limits_test.go",
    "internal/accounting/money_test.go",
    "internal/admission/service_test.go",
    "internal/apikey/fixture_test.go",
    "internal/apikey/legacy_publication_test.go",
    "internal/apikey/legacy_test.go",
    "internal/apikey/model_access_test.go",
    "internal/apikey/service_test.go",
    "internal/app/boundary_test.go",
    "internal/app/configuration_test.go",
    "internal/app/connection_lifecycle_test.go",
    "internal/app/console_clock_queue_test.go",
    "internal/app/console_clock_recovery_test.go",
    "internal/app/console_contract_test.go",
    "internal/app/console_http_test.go",
    "internal/app/console_migration_test.go",
    "internal/app/console_process_test.go",
    "internal/app/console_resource_matrix_test.go",
    "internal/app/console_review_regression_test.go",
    "internal/app/console_summaries_test.go",
    "internal/app/console_test.go",
    "internal/app/console_v3_test.go",
    "internal/app/crash_test.go",
    "internal/app/health_test.go",
    "internal/app/inprocess_test.go",
    "internal/app/legacy_fixture_test.go",
    "internal/app/lifecycle_test.go",
    "internal/app/metering_regression_test.go",
    "internal/app/owner_fixtures_test.go",
    "internal/app/persistence_test.go",
    "internal/app/qa_connection_contract_test.go",
    "internal/app/relay_test.go",
    "internal/app/routing_boundary_test.go",
    "internal/app/routing_contract_regression_test.go",
    "internal/app/routing_http_test.go",
    "internal/app/routing_latency_recovery_test.go",
    "internal/app/routing_latency_test.go",
    "internal/app/routing_test.go",
    "internal/app/single_fixture_test.go",
    "internal/app/sse_terminal_regression_test.go",
    "internal/app/version_test.go",
    "internal/catalog/catalog_test.go",
    "internal/configstate/coordinator_test.go",
    "internal/finance/finance_test.go",
    "internal/finance/prices_read_test.go",
    "internal/httpconsole/auth_contract_test.go",
    "internal/httpconsole/catalog_test.go",
    "internal/httpconsole/overview_test.go",
    "internal/httpconsole/read_coordination_test.go",
    "internal/identity/access_test.go",
    "internal/identity/identity_test.go",
    "internal/installation/recovery_test.go",
    "internal/installation/routing_original_test.go",
    "internal/installation/store_test.go",
    "internal/installation/upgrade_test.go",
    "internal/jsoncodec/json_test.go",
    "internal/records/legacy_test.go",
    "internal/relay/gateway_output_test.go",
    "internal/relay/metering_undispatched_test.go",
    "internal/relay/observation_guard_test.go",
    "internal/relay/protocol_adapter_test.go",
    "internal/reporting/legacy_overview_test.go",
    "internal/reporting/reporting_test.go",
    "internal/reporting/tenants_test.go",
    "internal/routing/engine_test.go",
    "internal/routing/latency_test.go",
    "internal/sessions/sessions_test.go",
    "internal/upstream/client_test.go",
    "internal/upstream/original_regression_test.go",
    "internal/upstream/protocol_role_test.go",
    "internal/upstream/protocol_test.go",
    "internal/webui/web_test.go"
  ],
  "assets": [
    "internal/webui/web/index.html",
    "internal/webui/web/style.css",
    "internal/webui/web/workspace.css",
    "internal/webui/web/workspace.html"
  ],
  "context": [
    ".trellis/spec/backend/index.md",
    ".trellis/spec/backend/multitenant-contracts.md",
    "AGENTS.md",
    "architecture.json",
    "docs/engineering/architecture-checks.md",
    "docs/engineering/architecture.md",
    "docs/roadmap.md",
    "go.mod",
    "go.sum",
    "internal/app/console_clock_queue_test.go",
    "internal/app/console_contract_test.go",
    "internal/app/health_test.go",
    "internal/records/legacy_test.go",
    "package-lock.json",
    "package.json"
  ]
}
```
