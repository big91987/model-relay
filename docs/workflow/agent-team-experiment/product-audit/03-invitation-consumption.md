# 邀请消费原子性与权限静态审查

## 结论与范围

本报告仅静态审查 Schema 3 的 invitation schema、`tenantInvitations`、`acceptInvitation` 及其直接 HTTP/UI/存储合同；没有启动服务、运行测试、访问真实邀请、读取数据库或修改实现。成员变更/退出/最后管理员等生命周期不在本报告重复展开。

结论：未发现可由普通租户成员创建、列出或撤销邀请，或通过篡改请求提升目标角色的直接路径；新用户、membership、邀请状态和审计写入在源码上共享一个 SQL 事务，失败路径由延迟回滚兜底。但“已接受”分支把持有旧令牌的匿名重放当作无条件成功，绕过当前租户状态、到期及既有账户登录绑定检查（P1）；公开端点还使用所有调用者共享的 10 次/分钟进程级桶，存在低成本阻断合法消费的风险（P1）。另有展示时钟与消费时钟不一致（P2），以及缺少针对事务失败/重复消费的直接合同证据（P1 验证缺口）。

## 确认事实与已有防护

### Schema、令牌与账户绑定

- `invitations` 将租户、规范化账户、目标角色、唯一 `digest`、到期时间、三态状态及 `accepted_by` 持久化；角色和状态均有 `CHECK`，摘要有 `UNIQUE`，租户/接受者有外键。[internal/relay/identity.go:15](../../../../../internal/relay/identity.go#L15) [internal/relay/identity.go:18](../../../../../internal/relay/identity.go#L18)
- 创建端生成 32 字节 CSPRNG 随机值（Raw URL-safe Base64，当前为 43 字符），数据库只写 SHA-256 摘要；明文只在创建响应中返回，UI明确提示“一次显示、手工通过受信渠道交付”。对高熵随机 bearer secret，未加盐 SHA-256 摘要不构成可行的离线枚举面。[internal/relay/security.go:20](../../../../../internal/relay/security.go#L20) [internal/relay/security.go:27](../../../../../internal/relay/security.go#L27) [internal/relay/identity.go:292](../../../../../internal/relay/identity.go#L292) [internal/relay/identity.go:294](../../../../../internal/relay/identity.go#L294) [internal/relay/identity.go:298](../../../../../internal/relay/identity.go#L298) [internal/relay/web/workspace.js:170](../../../../../internal/relay/web/workspace.js#L170)
- 接受时按令牌摘要查询，不接受客户端提供账户或角色；membership 使用邀请记录中的 `account`、`tenant_id` 和 `role`，因此请求体不能改绑账户或抬升角色。[internal/relay/identity.go:426](../../../../../internal/relay/identity.go#L426) [internal/relay/identity.go:448](../../../../../internal/relay/identity.go#L448) [internal/relay/identity.go:495](../../../../../internal/relay/identity.go#L495)
- 若规范化账户已存在，消费必须带该用户的有效 session 且 CSRF 常量时间匹配，并再次确认用户启用；仅持邀请令牌不能给既有账户增加 membership。若账户不存在，令牌本身是创建该邮箱名义账户的能力，密码须满足 12–1024 字符等规则；响应明确 `email_verified:false`，实现不声称验证邮箱所有权。[internal/relay/identity.go:52](../../../../../internal/relay/identity.go#L52) [internal/relay/identity.go:58](../../../../../internal/relay/identity.go#L58) [internal/relay/identity.go:459](../../../../../internal/relay/identity.go#L459) [internal/relay/identity.go:461](../../../../../internal/relay/identity.go#L461) [internal/relay/identity.go:467](../../../../../internal/relay/identity.go#L467) [internal/relay/identity.go:471](../../../../../internal/relay/identity.go#L471) [internal/relay/identity.go:499](../../../../../internal/relay/identity.go#L499)

### 过期、撤销、重复消费与原子性

- 首次消费只允许有效时钟严格早于 `expires`、状态为 `pending` 且租户启用；恰好到期即拒绝。撤销仅把同租户的 `pending` 记录改为 `revoked`，并要求恰好更新一行。[internal/relay/identity.go:448](../../../../../internal/relay/identity.go#L448) [internal/relay/identity.go:454](../../../../../internal/relay/identity.go#L454) [internal/relay/identity.go:301](../../../../../internal/relay/identity.go#L301) [internal/relay/identity.go:307](../../../../../internal/relay/identity.go#L307) [internal/relay/identity.go:312](../../../../../internal/relay/identity.go#L312)
- 接受入口在进程内持有 `server.mu`，存储又限制为单连接；事务开始后，新用户插入、membership 插入、邀请从 `pending` 条件更新为 `accepted`、`accepted_by`、审计和提交处于同一 `sql.Tx`。任何显式错误或 commit 失败都会返回 503，延迟 `Rollback` 覆盖未提交路径，因此源码结构上不会提交“创建用户但未写 membership”或“membership 已写但邀请未消费”的局部状态。[internal/relay/identity.go:434](../../../../../internal/relay/identity.go#L434) [internal/relay/identity.go:439](../../../../../internal/relay/identity.go#L439) [internal/relay/identity.go:444](../../../../../internal/relay/identity.go#L444) [internal/relay/identity.go:478](../../../../../internal/relay/identity.go#L478) [internal/relay/identity.go:495](../../../../../internal/relay/identity.go#L495) [internal/relay/store.go:69](../../../../../internal/relay/store.go#L69)
- membership 的复合主键和邀请摘要唯一约束是第二层防护；接受前也显式拒绝目标用户已在该租户的情况。[internal/relay/identity.go:17](../../../../../internal/relay/identity.go#L17) [internal/relay/identity.go:18](../../../../../internal/relay/identity.go#L18) [internal/relay/identity.go:486](../../../../../internal/relay/identity.go#L486)

### 邀请者与目标角色权限

- 所有租户 invitation 操作先通过 `scopedRole`；平台管理员对存在的租户被映射为 `administrator`，普通用户必须有该租户有效用户 membership。`tenantInvitations` 再明确要求 `administrator`，故 developer/finance 不能列出、创建或撤销邀请；跨租户找不到角色时返回 404。[internal/relay/identity.go:99](../../../../../internal/relay/identity.go#L99) [internal/relay/identity.go:203](../../../../../internal/relay/identity.go#L203) [internal/relay/identity.go:243](../../../../../internal/relay/identity.go#L243)
- 创建端只允许 `administrator`、`developer`、`finance` 三个目标角色，1–168 小时；租户管理员可以邀请另一管理员。这与当前角色集合/UI直接合同一致，源码中没有“只有平台管理员才能授予 administrator”的策略，因此本审查不把同级管理员授予视为实现越权，但产品策略若另有要求需单独收紧。[internal/relay/identity.go:48](../../../../../internal/relay/identity.go#L48) [internal/relay/identity.go:274](../../../../../internal/relay/identity.go#L274) [internal/relay/identity.go:288](../../../../../internal/relay/identity.go#L288) [internal/relay/web/workspace.js:254](../../../../../internal/relay/web/workspace.js#L254)
- 接受路由是特意开放的未登录端点，但所有写请求仍先受精确同源 `Origin` 检查；既有账户路径另加 session/CSRF 绑定。[internal/relay/server.go:200](../../../../../internal/relay/server.go#L200) [internal/relay/server.go:210](../../../../../internal/relay/server.go#L210) [internal/relay/identity.go:461](../../../../../internal/relay/identity.go#L461)

## 缺陷假设与影响分级

### P1 — 已接受令牌可匿名、无条件重放为成功

`acceptInvitation` 查到 `state == "accepted"` 后立即返回 200/`already_applied:true`，发生在到期、租户启用、既有账户 session/CSRF 和 `accepted_by` 检查之前。[internal/relay/identity.go:445](../../../../../internal/relay/identity.go#L445) [internal/relay/identity.go:450](../../../../../internal/relay/identity.go#L450) [internal/relay/identity.go:454](../../../../../internal/relay/identity.go#L454) [internal/relay/identity.go:461](../../../../../internal/relay/identity.go#L461)

影响：泄露或留存的旧链接长期充当邀请存在/已消费状态 oracle；租户停用或邀请到期后仍得到“成功”语义；对既有账户，未登录持令牌者也能得到成功，破坏账户绑定合同的一致性。该分支不再次写 membership，故不是重复提权，但响应缺少首次成功返回的 `account`/`tenant_id`，UI仍拼接 `result.account`，会显示 `undefined` 并误导用户。[internal/relay/web/workspace.js:285](../../../../../internal/relay/web/workspace.js#L285)

建议合同：重复消费默认返回统一的 `invitation_unavailable`；若业务必须支持幂等，只能对与 `accepted_by` 相符的已登录账户返回一致响应，并明确是否仍要求租户启用。令牌错误、撤销、过期和已接受应避免形成不必要的外部状态 oracle。

### P1 — 公开接受端点的全局限流可被任意来源耗尽

每次请求在解析 JSON/验证令牌前就追加到 `server.invitationAttempts`；该切片没有按来源、邀请摘要或账户分桶，达到 10 次/分钟后所有邀请接受均返回 429。[internal/relay/identity.go:401](../../../../../internal/relay/identity.go#L401) [internal/relay/identity.go:404](../../../../../internal/relay/identity.go#L404) [internal/relay/identity.go:411](../../../../../internal/relay/identity.go#L411) [internal/relay/identity.go:417](../../../../../internal/relay/identity.go#L417) [internal/relay/console_contract_test.go:315](../../../../../internal/relay/console_contract_test.go#L315)

影响：能向同源端点发请求的攻击者可持续阻断全部租户的合法邀请消费；重启又会清空桶，且多实例时各实例独立，保护强度不稳定。`passwordSlots` 可限制昂贵 Argon2 并发，但不能消除全局可用性问题。建议使用受控的多维/有界持久或边缘限流（来源加令牌摘要，另保留全局高水位），并确保无界 key 不造成内存攻击。

### P1 — 原子性和一次消费缺少直接故障/竞争合同证据（验证缺口）

当前复合 `Exec` 同时执行 membership 插入和邀请条件更新，但未检查邀请更新恰好影响一行；安全性依赖进程互斥、单数据库连接、安装目录排他锁及 SQLite 事务失败语义。[internal/relay/identity.go:495](../../../../../internal/relay/identity.go#L495) [internal/relay/store.go:48](../../../../../internal/relay/store.go#L48) [internal/relay/store.go:69](../../../../../internal/relay/store.go#L69)

源码结构支持原子提交，不能替代可执行证据。直接合同测试目前能证明三种角色的正常创建/接受和限流边界，但在本次静态检索中未见：两个同时消费同一令牌、撤销与消费竞争、用户插入后 membership/审计/commit 故障回滚、既有用户账户绑定、重复已接受响应、membership 已存在时新用户不残留等专项断言。[internal/relay/console_contract_test.go:554](../../../../../internal/relay/console_contract_test.go#L554) [internal/relay/console_contract_test.go:564](../../../../../internal/relay/console_contract_test.go#L564) [internal/relay/console_contract_test.go:571](../../../../../internal/relay/console_contract_test.go#L571)

建议后续将邀请状态更新拆成可检查 `RowsAffected()==1` 的语句，仍置于同一事务，并用独立临时库做竞争与故障注入回归。本项是证据缺口，不是已确认发生部分提交。

### P2 — 邀请列表与消费使用不同时间源

列表用 `time.Now().UTC()` 把 pending 投影为 `expired`，消费用防回拨的 `Store.now()`；创建也用 `Store.now()`。[internal/relay/identity.go:262](../../../../../internal/relay/identity.go#L262) [internal/relay/identity.go:293](../../../../../internal/relay/identity.go#L293) [internal/relay/identity.go:403](../../../../../internal/relay/identity.go#L403) [internal/relay/identity.go:454](../../../../../internal/relay/identity.go#L454)

影响：系统时钟回拨/有效时钟高水位场景下，管理员看到的状态可能与实际能否消费不一致；列表只是展示投影，不改变数据库状态。建议统一使用本请求已推进的有效时钟。

### P2 — 已是成员的邀请保持 pending，管理端缺少前置约束

若目标账户对应用户已拥有该租户 membership，消费返回 409，但邀请不撤销、不接受，直到到期或管理员手动撤销。[internal/relay/identity.go:486](../../../../../internal/relay/identity.go#L486) [internal/relay/identity.go:491](../../../../../internal/relay/identity.go#L491)

影响主要是重复邀请记录、持续可重试及管理员困惑，不产生额外 membership。可在创建时拒绝当前成员，或在消费时采用明确的终态/幂等合同；需避免把该调整扩展为成员生命周期重构。

## 未执行项与判定边界

- 按任务约束，未运行 `go test`、`make verify`、服务、浏览器或任何数据库命令；未复用依赖任务的绿色门禁作为本审查的邀请结论。
- 未做真实邀请交付、邮件所有权、代理/CDN 来源识别、多进程竞争、SQLite 故障注入、崩溃恢复、备份恢复后的令牌行为或性能/DoS 实测。
- 未检查或评价成员修改、移除、自退、最后管理员保护等成员生命周期。
- P0：未发现。P1：已接受令牌无条件成功重放；全局公开限流可用性风险；事务/竞争专项证据缺口。P2：列表时间源不一致；既有成员邀请保持 pending。

