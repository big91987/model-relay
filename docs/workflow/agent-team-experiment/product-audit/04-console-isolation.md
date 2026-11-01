# 控制台管理接口隔离静态审查

## 结论与边界

结论：在本次指定源码范围内，**未发现可成立的 P0/P1 principal、tenant 或 role 越权链**。路径中的 tenant、model、bill、request、operation 等对象 ID 均在读取或写事务前进入对应的主体/租户/角色约束；平台级模型财务、账单关闭/更正、用量与历史对账均有显式平台管理员门槛。此结论仅是静态代码审查结果，不是运行测试、产品验收或真实部署结论。

审查对象为 `internal/relay/server.go`、`console_api.go`、`console_reports.go`、`admin_response.go`，并追踪它们直接调用的权限合同：`identity.go` 的 `Principal`/`principal`/`scopedRole`、`billing.go`、`legacy_accounting.go`、`console_overview.go` 与操作幂等记录。成员、邀请以及 Key 的签发、轮换、撤销等专属生命周期不在本报告结论内；只记录这些对象与本次路由隔离的交界。

严重度结果：P0 0 项，P1 0 项，P2 2 项（均为纵深防护/可维护性风险，未证明为当前可利用越权）。

## 判定口径

- **事实**：由本次所读候选源码直接支持，并引用行号。
- **假设**：静态推理依赖但本次没有运行验证的条件。
- **未执行**：没有启动服务、没有发 HTTP、没有运行 Go/浏览器/数据库测试、没有读取或修改真实数据；依赖任务的既有成功回执不作为本审查的执行证据。
- 影响等级：P0 为无需额外前提的全局控制面失守；P1 为明确跨租户/跨角色敏感读写；P2 为局部信息暴露、拒绝语义或容易在后续改动中破坏的纵深防护缺口。

## 入口、principal 与直接权限合同

### 会话入口

**事实**：所有 `/admin/api/` 请求先进入 `admin`；除登录和邀请接受外，必须取得服务端 session。所有非 GET 请求还要求精确 Origin 与 session 中的 CSRF 值，见 `server.go:169-170,200-220`。Schema 3 的控制台路由在调用旧管理处理器前，先尝试 `identityAPI`/`consoleAPI`；未被二者识别的路由会重新加载 principal，并拒绝非平台主体，见 `server.go:250-263`。因此 `/admin/api/upstreams`、旧 `/admin/api/models...`、旧全局 keys/requests 等回退路由不能被租户角色直接使用。

**事实**：principal 不取自请求参数或前端角色，而是由 session 的 `UserID` 查询 `users.id`，且要求用户仍 active，见 `identity.go:93-96`。角色也不取自请求体：平台 principal 对存在的 tenant 被映射为 administrator；非平台 principal 必须命中 `(tenant_id,user_id)` membership 且用户 active，见 `identity.go:99-106`。移除 membership 或停用 user 后，旧 session 不能继续获得该租户 scope。

**防护与影响**：principal、tenant membership 和 role 均在服务端重读，没有可见的 `role=`/`principal=` 客户端覆盖点。未知租户和无 membership 通常被映射为 404，减少对象存在性泄露。

### 路径解析与事务边界

**事实**：`consoleAPI` 只承认列举的 workspace、catalog、report、bill、model finance 和 tenant 子资源形状；model finance 要求恰好三个 path part，tenant 子资源至少三个 part，见 `console_api.go:12-23`。识别后重新加载 principal，在同一 `server.mu` 临界区内开启一个 SQL transaction，授权查询、对象查询、写入、审计与 commit 共用该事务，见 `console_api.go:24-36`。

**事实**：`/tenants/{tenant}/...` 在分派 limits/bills/keys 之前先调用 `scopedRole`，见 `console_api.go:60-74`。`/models/{alias}/{prices|bounds|tenant-grants}` 不接收 tenant scope，而是在处理器第一步要求 `principal.Platform`，再确认 alias 唯一存在，见 `console_api.go:176-185`。`/operations/{id}` 同时绑定 `actor=principal.ID` 与 operation ID；若记录带 tenant，再实时复核 tenant scope；无 tenant 的平台操作拒绝非平台主体，见 `console_api.go:40-59`。

**假设**：SQLite 连接实际启用了 schema 所声明的唯一键/外键约束；本次没有打开数据库或验证 PRAGMA。即使外键未启用，本次所见读写授权仍以 principal 与显式 tenant 条件为主，但孤儿数据可能改变 404/503 等行为。

## 逐类权限与对象 ID 混用审查

| 资源/动作 | principal、tenant、role 约束（事实） | 对象 ID 混用检查与影响 |
| --- | --- | --- |
| Workspace overview | `reportScope`：平台且无 tenant 可全局读；非平台必须给 tenant 并通过 `scopedRole`；developer 额外要求请求 Key 的诊断分配，见 `console_reports.go:39-56`、`console_overview.go:15-22`。 | overview 的所有 request/admission 聚合复用同一 `where/args`；tenant distribution 也带该 where，见 `console_overview.go:28-49`。未见把 tenant A 的筛选参数用于 tenant B 聚合。 |
| Tenant limits | tenant path 先取得 role；developer 在 tenant owner scope 被拒绝；GET 对 administrator/finance 开放，PUT 仅 administrator，见 `console_api.go:61-72,87-126`。非平台管理员写入还受对应 tenant 的 platform cap 限制；平台管理员才同时改 platform limit，见 `console_api.go:139-169`。 | `ownerType="tenant"` 与 `ownerID=tenantID` 由路由固定传入，不来自 body；读写和审计均使用同一 tenantID，见 `console_api.go:71,93-121,169`。未见 owner ID 替换。 |
| Model prices/bounds/tenant grants | 整类入口仅平台 principal，且 alias 先在 models 中确认，见 `console_api.go:176-185`。tenant grant body 中每个 tenant 必须存在且 active；bounds/price 的 upstream 必须是该 alias 的 route candidate，见 `console_api.go:203-217,249-254,315-347`。 | alias 与 upstream 的复合查询阻止把另一模型的 upstream ID 混入；写入、operation/audit object 都使用已校验 alias。tenant grant 是显式平台级批量替换，不受租户管理员控制。 |
| Catalog | 非平台必须提供 tenant；任何非空 tenant 都先过 `scopedRole`；`include_ungranted=true` 仅平台可用，见 `console_api.go:359-384`。 | tenant catalog 查询用 grant 的 `tenant_id=?`；每个 summary 继续收到同一 principal、tenantID、alias，见 `console_api.go:382-407`。未见 query tenant 与 summary tenant 分离。 |
| Audit | 先通过 `reportScope`；指定 tenant 时仅 administrator 可读，developer/finance 均 403；平台不指定 tenant 可读全局，见 `console_reports.go:109-135`。 | tenant audit SQL 显式 `WHERE tenant_id=?`。平台全局读取是授权能力，不是 tenant ID 绕过。 |
| Request list/usage/export/detail | `reportScope` 形成 `c.tenant_id=?`；developer 再要求 `(key_id,tenant_id,user_id)` 的诊断分配，finance/tenant administrator 可读本 tenant，平台可显式全局读，见 `console_reports.go:39-56,109-154`。筛选值均参数化，见 `console_reports.go:138-152`。 | 详情先以 scoped base + `r.id=?` 确认唯一父请求；只有确认后才按同一 request ID 读取 attempts、prices、charges、routing 与历史证据，见 `console_reports.go:155-200`。子查询本身不重复 tenant 条件，但建立在已授权父请求上；request ID 混入另一租户会在父查询得到 404。列表、聚合和 CSV 继续使用同一 base，见 `console_reports.go:203-267`。 |
| Usage/legacy reconciliation | 路由形状必须是 request ID + reconciliation 子资源；两个写处理器第一步均要求平台 principal，见 `console_reports.go:96-104`、`billing.go:204-208`、`legacy_accounting.go:9-13`。随后由 admissions/legacy record 反解 tenant，所有账单封闭检查、追加账本与 operation/audit 使用该 tenant，见 `billing.go:234-267`、`legacy_accounting.go:43-84,140-144`。 | request ID 不接受独立 tenant 参数，因此不存在“路径 tenant A + request B”双来源冲突；平台权限是显式前置条件。attempt ordinal 及 upstream/dispatch 关系另有复合检查。 |
| Tenant bill list/close | tenant path 已通过 `scopedRole`；developer 禁止，finance/administrator 可 GET；POST close 另要求平台 principal，见 `billing.go:15-41`。所有金额、pending 与 insert 都以路由 tenantID + body period 查询，见 `billing.go:58-103`。 | bill close 的 tenant 不能由 body 替换。平台对既有 bill 的幂等回放仍记录同一 tenant。 |
| Bill detail/export/correction | 先按 bill ID 取出其 tenant，再对该 tenant 调用 `scopedRole`；developer 拒绝。读允许 finance/administrator；任何非 GET 后续动作先要求平台，correction 再要求精确子路径，见 `billing.go:110-160`。 | 导出 usage 与 corrections 都使用由 bill 反解的 tenant/period；写 correction 固定使用该 bill ID，operation/audit tenant 也是反解值，见 `billing.go:131-150,170-200`。另一租户 bill ID 对非成员表现为 404。 |
| Operation receipt | 查询键为 `(actor,operation)`，随后对记录 tenant 重新鉴权；平台级空 tenant receipt 仅平台可读，见 `console_api.go:40-59`。幂等 fingerprint 包含 action、URL path 和 body，见 `console_keys.go:35-57`。 | actor 绑定阻止枚举他人的 operation ID；tenant 被保存于 finish 记录并在读取时复核，见 `console_keys.go:60-64`。 |
| 旧管理回退 | Schema 3 下，未被新控制台路由消费的旧 upstream/model/key/request 管理接口在分派前统一要求 `principal.Platform`，见 `server.go:250-289`。 | 旧全局 SQL 本身没有 tenant 条件，但只有平台 principal 能到达；这是平台控制面而非租户接口。 |
| Buffered admin response | `adminResponse` 仅包装上述平台回退处理器，克隆 headers、缓存 status/body 并在返回时一次 flush，见 `server.go:268-278`、`admin_response.go:8-40`。 | 它不解析 principal/tenant/object，也不改变授权分派；未发现其造成隔离绕过。 |

## 角色结论

- **平台管理员**：可跨 tenant 查看 workspace/report/audit，管理模型财务与授权、关闭/更正账单、执行 usage/legacy reconciliation，也可进入旧全局管理路由。所有这些能力均有 `Platform` 或平台专属回退门槛；这是显式平台权限，而非越权。
- **租户 administrator**：必须先通过目标 tenant membership；可读本租户聚合、详情、审计、账单与 limits，并可更新本租户 limits（受 platform cap 约束）。不能进入模型财务、平台对账、账单关闭/更正或旧全局管理路由。
- **developer**：请求/usage/overview 的 SQL scope 进一步绑定到本用户获分配诊断权限的 Key；tenant limits、bills、audit 被显式拒绝。catalog 可读已授权模型。成员与 Key 分配生命周期本次不作专项结论。
- **finance**：可读本租户 limits、请求/usage、账单与 catalog；audit 明确只允许 administrator，控制面写操作也因 role 或 `Platform` 门槛被拒绝。该“财务观察者可读租户费用与请求计量事实”的边界与代码直接合同一致；本次未发现 finance 写入路径。

## P2 观察与建议

### P2-01：详情子资源依赖“先授权父请求”的隐式合同

**事实**：request detail 先用 tenant/developer scoped 父查询确认 `r.id`，随后 attempts、price snapshots、charge entries、routing evidence、historical evidence 与 confirmed token 查询只带 request ID，不重复 tenant 条件，见 `console_reports.go:155-200,270-277`。

**防护**：当前控制流在 `len(values) != 1` 时立即返回，子查询不可先行；request ID 是这些子表的父标识。因此当前未形成跨租户泄露。

**影响**：这是容易被未来重构破坏的 caller-order 授权合同。例如把子记录读取抽成可独立调用的 handler、并发提前查询，或允许非唯一/可复用父标识时，会把隔离从 SQL 约束退化为调用顺序约束。等级 P2（纵深防护），不是当前可利用漏洞。

**建议**：将“已授权 request”封装成不可绕过的 scoped loader，或让详情子查询通过 `requests`/`key_controls` 重复关联同一 scope；补充跨 tenant request ID、developer 未分配 Key、finance/administrator 的矩阵回归。该建议不在本只读任务中实施。

### P2-02：授权规则分散在路由与下游处理器，缺少统一 capability 合同

**事实**：tenant limits 的 role 判断位于 `manageLimits`，bill detail 先反解 tenant 再判断，report 由 `reportScope` 拼 SQL，reconciliation/legacy/model finance 则各自在处理器用 `principal.Platform`；旧路由另由 `admin` 的统一平台检查保护，见 `console_api.go:60-82,87-126,176-180`、`console_reports.go:39-56`、`billing.go:110-160,204-208`、`legacy_accounting.go:9-13`、`server.go:250-289`。

**防护**：逐条追踪时，每个当前入口均找到相应门槛，且写事务使用已校验的对象/tenant；未发现遗漏。

**影响**：新增同名路由、改变 `known` 判定或直接复用 handler 时，容易遗漏调用者负责的那一半合同；错误更可能表现为跨角色能力扩大。等级 P2（维护性导致的授权回归风险）。

**建议**：建立显式 capability→role 表及统一授权函数，让 handler 自身声明并强制 capability；对每一 capability 固定覆盖 platform、tenant administrator、developer、finance、无 membership、另一 tenant 六类主体。保留 404/403 的存在性隐藏约定。该建议不在本任务中实施。

## 未发现项与未执行项

**未发现**：没有看到客户端可注入 principal/role；没有看到 tenant 路径校验后用另一个 body/query tenant 写入；没有看到 bill ID 绕过其反解 tenant；没有看到 request ID 在父 scope 校验失败后仍读取子记录；没有看到 developer 读取未分配 Key 的报告；没有看到 finance 或租户管理员进入平台写事务；没有看到 operation ID 跨 actor 读取。

**未执行**：没有运行 `make verify`、Go 单元/集成测试、实际 HTTP handler、浏览器、SQLite 查询计划或迁移/恢复；没有构造恶意编码路径、重复 query、并发角色变更、损坏/孤儿数据库或真实 session；没有检查成员与 Key 生命周期的完整实现。三项已合入依赖（严格 Bearer、单次 clock advance、严格 backup manifest）只作为候选源码背景，不把其历史执行结果计入本报告。

**假设/剩余风险**：Go `net/http` 提供正常解析后的 `URL.Path`；部署没有在服务前以不一致规则重写 path/query；数据库 schema 与候选源码一致；请求/账单/operation ID 保持其 schema 唯一性。上述条件任一不成立时，需要独立动态/数据完整性审查，不能由本报告外推产品通过。
