# 历史查询与诊断隔离静态审查

## 结论与范围

本轮只静态阅读候选源码和已有测试文件，未启动产品、未连接数据库或上游、未读取真实历史/日志，也未运行测试。审查对象为 `server.go` 的旧版 records 入口、Schema 3 的 `console_reports.go`，请求/尝试/路由证据读取、CSV 输出和 `admin_response.go`。通用登录、会话和 CSRF 控制不在本报告重复评价。

结论：未发现 P0。请求详情的父记录查询具有租户/成员行级范围，未发现仅凭已知 request ID 跨租户读取子表的可达路径；持久化 schema 也不保存 prompt、completion 正文或上游 token，因此本范围内没有这些正文/凭据由历史接口直接返回的事实。发现 2 项 P1 缺陷假设和 2 项 P2 加固项：Schema 3 对 administrator、finance、developer 缺少响应字段裁剪；人工 `reason/evidence` 被复用到更宽的读取/CSV 面；敏感历史响应没有显式 `no-store`；深页 OFFSET 缺少稳定快照/游标语义。

## 事实与现有防护

1. **旧 records 只对平台面开放，但没有租户语义。** Schema 3 请求先由 `consoleAPI` 接管；无 tenant 的非 workspace requests 会退回旧入口，但非平台 principal 随后被拒绝（`internal/relay/console_api.go:12-23`，`internal/relay/server.go:250-263`）。旧列表使用 `rowid < cursor`、倒序和 `limit+1`，limit 为 1..100（`server.go:463-509`）；旧单记录要求 22 字符合法 ID 并参数化查询（`console.go:28-45`）。因此旧入口是平台级兼容面，不是租户成员查询面。
2. **Schema 3 行级范围在父查询中生效。** 平台未指定 tenant 时可查询全局；非平台必须指定 tenant，且 `scopedRole` 失败统一为 404。developer 还必须命中该 key 的 `key_diagnostic_members`（`console_reports.go:39-56`）。模型、key、state、request_id 都用参数绑定；时间必须通过 RFC3339Nano 解析后再绑定（`console_reports.go:138-153`）。列表 count、数据和详情父记录均复用同一 `where`（`console_reports.go:154-158,236-250`）。
3. **单记录的子证据读取依赖已通过范围检查的父记录。** 父查询若不是恰好一条即 404；随后 attempts、prices、charges、routing evidence、historical evidence 才按 request ID 查询，且均在同一 transaction 内（`console_reports.go:155-200`；`routing_latency_storage.go:62-108`）。request ID 是 `requests` 主键（`store.go:82,104`），子表也以 request ID 建外键/复合主键（`store.go:89-90`）。这避免了“子查询自身无 tenant 条件”形成当前可达的 IDOR；未来若拆分父子查询或允许非唯一 ID，必须保留该不变量。
4. **历史存储不含正文和 token。** `requests` 只保存标识、alias/远端 model、状态、错误码、HTTP 状态、耗时和 token 数；`request_attempts` 保存供应商/远端模型及同类诊断元数据（`store.go:82-90`）。`recordColumns` 及 Schema 3 SELECT 与此一致（`console.go:9`；`console_reports.go:158,167,250`）。上游 token 位于 `upstream.token`，不在这些 SELECT 中（`store.go:84-86`）。已有受控测试还检查管理查询不包含 key、管理员密码、上游 token、prompt 和 answer 哨兵（`inprocess_test.go:102-113`；`relay_test.go:358-365`），但本轮没有执行这些测试。
5. **错误返回做了归一化。** 公共 failure 只组合内部 code，不回显 SQL error 或供应商响应正文（`server.go:82-87`）；报告查询失败统一为 `storage_unavailable`。历史会返回结构化 `error_code` 和数值 `upstream_status`，但没有原始供应商错误消息（`console_reports.go:158,167,250`）。
6. **分页/导出有边界。** JSON page 为 1..100000、page_size 为 1..100；CSV 在过滤后总数超过 10000 时拒绝，并固定最多 10000 行（`console_reports.go:219-260`）。CSV 使用标准库 quoting，并对首字符为 `= + - @ TAB CR` 的非数值内容前置单引号（`console_reports.go:281-307`）。导出列是显式 allowlist，不含远端 model、upstream 或自由文本 evidence（`console_reports.go:260`）。
7. **`adminResponse` 不做字段变换。** 它克隆 header、缓冲完整 body，最终原样复制 header/status/body（`admin_response.go:8-40`）。因此它既不会额外泄漏内部对象，也不会替历史/CSV 响应做脱敏、缓存控制或大小限制。

## 缺陷与优先级

### P1 — 请求详情没有按角色裁剪供应商与财务诊断字段（高可信缺陷假设）

`reportScope` 只产生行级 where，没有把 role 返回给响应投影；父记录命中后所有角色共用同一 SELECT 和组装逻辑（`console_reports.go:39-56,155-200`）。这意味着：

- developer（仅被分配某 key 的诊断成员）可收到 `r.model`、`upstream_status`、attempt 的 `upstream_id`、`upstream_name`、`remote_model`，以及路由选择/延迟证据；关键选择见 `console_reports.go:158,167,182-189`。
- finance 和 tenant administrator 同样收到 attempts、routing、price snapshots、charge entries 和 historical price evidence；后四类还包含供应商 ID、内部价格、成本、费率和操作证据（`console_reports.go:172-199`）。
- 路由证据明确包含 upstream ID、算法、样本数、EWMA、TTL、探索序列等内部拓扑/策略信息（`routing_latency.go:27-54`）。

现有角色契约证明系统在其他资源会裁剪：developer 被禁止 tenant limits/bills/audit/upstreams（`console_contract_test.go:593-600`），账单详情也显式拒绝 developer（`billing.go:121-129`）；但未见针对 request detail 各角色字段集合的回归。若产品明确授权 key 诊断成员查看供应商身份和完整商业证据，此项可降级；当前代码与最小权限角色模型不一致，且任务边界明确要求防止远端模型/供应商秘密泄露，故列 P1。

**影响：** 被分配单个应用诊断权限的用户可枚举该应用历史并获知供应商名称、远端模型、路由策略和内部价格；finance 可获得超出结算所需的运行拓扑。供应商身份或模型名属于商业秘密时是实质信息披露。

**建议防护：** 在范围解析时保留 role，并为列表、详情、CSV 建独立 DTO allowlist。developer 仅返回请求 ID、alias、时间、状态、归一化错误码、耗时/用量及必要 attempt ordinal；隐藏 upstream ID/name、remote model、路由算法/样本和成本。finance 仅返回结算所需用量/金额/期间，隐藏 attempts 与路由。administrator/platform 才返回完整运营诊断。增加角色 × list/detail/export × assigned/unassigned/cross-tenant 矩阵测试，并用远端模型/供应商哨兵做负断言。

### P1 — 自由文本 `reason/evidence` 从高权限写入流向较宽读取面（中高可信缺陷）

usage reconciliation 仅平台可写，但 `reason`/`evidence` 的验证只是 1..128 字符且排除 CR/LF/NUL（`billing.go:205-223`；`server.go:110-112`）；legacy reconciliation 同样将这两个字段写入 `historical_price_evidence` 或 charge/reconciliation 记录（`legacy_accounting.go:9-30,126-133`）。请求详情随后把 `charge_entries.reason/evidence` 和 `historical_price_evidence.reason/evidence` 原样返回给任何通过上述行级 scope 的角色（`console_reports.go:177-199`）。账单 CSV 也把 charge `reason/evidence` 导出给非 developer 的租户角色（`billing.go:121-142`）。

这不是当前代码已存 prompt/token 的事实，而是明确的二次泄露通道：平台操作员若把工单摘录、供应商回复、prompt、token 或其他秘密填入 evidence，developer/finance 会在不同权限面读到它。长度和换行限制不构成内容分类或脱敏。

**影响：** 人工诊断习惯或误操作可把敏感正文/凭据带入持久历史及 CSV，扩大到租户角色和离线文件。

**建议防护：** evidence 改为受控枚举/不可逆引用（例如 receipt ID），显示字段另设严格非敏感 reason code；至少在写入 UI/API 明示“不得放入 prompt、completion、凭据或供应商响应”，并在 developer/finance DTO 中删除 evidence。不要尝试用 token 形状黑名单代替 allowlist。增加含 prompt/answer/token/供应商诊断哨兵的写入后角色读取负测试。

### P2 — 历史 JSON/CSV 缺少显式禁止缓存

`reply` 只设置 `Content-Type`（`server.go:78-80`）；CSV 只设置类型和附件名（`console_reports.go:281-283`）；`adminResponse.flush` 原样转发这些 header（`admin_response.go:31-40`）。未见 `Cache-Control: no-store`、`Pragma: no-cache` 或防 MIME sniffing。响应包含 key ID、请求状态、费用及可能的供应商诊断，即使会话授权正确，也可能留在共享浏览器/中间缓存或下载历史中。

**建议防护：** 在所有 `/admin/api/` 响应的统一边界设置 `Cache-Control: no-store`（兼容需要时加 `Pragma: no-cache`）和 `X-Content-Type-Options: nosniff`；CSV 保留 attachment，并测试成功/4xx/5xx 均带缓存策略。该项不等同于通用授权复述。

### P2 — OFFSET 深分页没有跨请求稳定性，且可制造昂贵扫描

Schema 3 接受 page 到 100000、page_size 到 100，并先 count 再 `LIMIT/OFFSET`（`console_reports.go:219-250`）。单次请求内使用同一 transaction 是一致的，但翻页发生在不同请求；新记录插入会造成重复/遗漏，最大 offset 约一千万行也可能形成昂贵扫描。旧入口已经采用 rowid cursor（`server.go:463-509`）。

**影响：** 运营导出之外的长历史浏览可能不稳定；授权用户可反复请求深页增加数据库负载。它不直接造成越权或正文泄漏，故为 P2。

**建议防护：** Schema 3 列表改用 `(time,id)` 或不可猜测签名游标，响应带 next cursor；若必须保留 page，降低最大页并设置查询预算。导出仍保持 10000 行过滤上限。

## 未发现的高危路径与限制

- **P0：无。** 未发现未认证访问、跨租户 request ID 直接读取、SQL 拼接用户值、上游 token SELECT、prompt/completion 正文持久化或原始供应商错误正文回显。
- `queryMaps` 会把任意选中列直接映射并将 `[]byte` 转字符串（`console_reports.go:59-93`），本身不是脱敏层；安全性完全依赖每条 SELECT 的列 allowlist。后续增加列时应视为安全敏感变更。
- 子表查询自身没有 tenant predicate；当前安全依赖“父记录先按 scope 唯一命中且所有读取在同一 transaction”。建议测试并文档化这一不变量，或让子查询 join scoped parent 以降低未来回归风险。
- CSV 首字符防公式控制覆盖当前常见触发字符，但本轮未用 Excel/LibreOffice 验证 Unicode、前导空白或特定客户端解释差异；不能据静态阅读声称所有表格软件安全。
- `adminResponse` 缓冲完整响应且忽略最终 write 错误（`admin_response.go:24-40`）。当前历史列表/CSV已有 100/10000 上限，未形成已证实的无界数据泄露；仍建议将 CSV 直接流式写出并记录 flush/write 失败，但这属于可靠性加固。

## 未执行项

- 未运行 Go、HTTP、浏览器、CSV 客户端或性能测试；未启动 Go 服务。
- 未访问真实历史、SQLite、日志、备份、凭据、prompt、completion 或供应商。
- 未验证真实角色会话、并发插入分页、缓存代理、十万/万行导出、损坏 routing evidence JSON 或客户端断连。
- 依赖任务的既有 `make verify` 回执只说明其候选变更当时的门禁结果，不是本静态审查的实测证据，也不据此给产品通过结论。

## 建议验收顺序

1. P1：先定义 administrator/developer/finance/platform 的 request list/detail/export 字段矩阵并实现 DTO 投影；用供应商名、remote model、价格和 evidence 哨兵做负测试。
2. P1：把自由文本 evidence 改为受控非敏感引用/代码，清理 developer/finance 返回面，并决定既有数据迁移/隐藏策略。
3. P2：统一 admin API `no-store`/`nosniff`，覆盖 JSON、CSV 和错误响应。
4. P2：用 keyset cursor 替代深 OFFSET，并在隔离临时数据库上验证并发翻页无重复/遗漏及查询成本。
