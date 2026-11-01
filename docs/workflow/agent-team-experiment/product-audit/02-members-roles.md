# 租户成员与角色生命周期静态审查

## 范围与结论

本报告只静态审查 `internal/relay/identity.go` 中 `tenants`、`memberships`、`tenantMembers`、`scopedRole`、用户/租户 active 状态及直接数据库/HTTP 合同；邀请仅作为“成员如何产生”的入口边界说明，不审查邀请消费实现本身，也不扩展到通用控制台授权。

结论：未发现 P0。确认 1 项 P1 生命周期缺陷：清理或降权一个**停用的管理员成员**时，“最后管理员”判断会错误地阻止操作，即使租户仍有一个有效管理员。另有 2 项 P2 合同缺口/假设需要产品确认：租户停用并不冻结成员管理；用户 active 有鉴权语义但没有本范围内可见的状态管理合同。成员增删改角、一般自我退出、跨租户隔离和真实最后有效管理员保护均已有明确防护。

## 确认事实与直接合同

| 生命周期 | 静态确认事实 | 影响/边界 |
|---|---|---|
| 数据模型 | `users.active`、`tenants.active` 均受 0/1 CHECK；membership 以 `(tenant_id,user_id)` 为主键，角色只允许 `administrator/developer/finance`（`identity.go:15-20`）。 | 同一用户在同一租户最多一个角色；角色值有数据库与应用双重约束（`validRole`，`identity.go:48-50`）。外键未声明 membership 删除级联。 |
| 成员新增 | 成员 API 没有直接 `POST /members`；membership 的产品新增路径是邀请接受后插入（`identity.go:486-495`），而本任务不继续审查消费细节。 | 不能通过 `tenantMembers` 任意指定用户新增成员；“新增”原子性/账号绑定属于明确排除项。 |
| 读成员 | 只有该租户管理员可列出成员；查询以 `m.tenant_id=?` 限定并返回用户 active（`identity.go:327-354`）。 | 开发者/财务不能枚举成员；停用用户仍作为成员记录可见，便于治理。 |
| 改角/移除 | 目标 membership 先以 `(tenant_id,user_id)` 精确读取，PATCH 复验角色，随后同一事务更新或删除并写审计（`identity.go:356-398`）。 | 目标无法跨租户串改；不存在的目标统一 404。角色变更不是删除/重建。 |
| 自我退出/降权 | 非管理员只有一个写例外：删除自己的 membership（`identity.go:208-211,360-364`）。自我 PATCH 降权仍要求当前角色为管理员；若存在其他有效管理员则允许（`identity.go:377-392`）。 | 普通成员可退出但不能自助改角；管理员可自我降权/退出，但受最后管理员保护。 |
| 最后管理员 | 当目标原角色是管理员且操作将其删除或改成非管理员时，代码统计该租户 active 用户中的管理员，`<=1` 返回 409 `last_administrator`（`identity.go:377-386`）。 | 对“目标是当前唯一有效管理员”的正常场景提供保护。所有身份写路径受 `server.mu` 串行并处于事务内（`identity.go:130-137`），同进程并发不会同时绕过计数。 |
| 用户停用 | 登录和 session principal 重载都要求 `users.active=1`（`identity.go:75-96`）；普通成员的 `scopedRole` 也再次要求用户 active（`identity.go:99-106`）。成员列表仍显示 active 字段（`identity.go:333-347`）。 | 用户一旦在存储中停用，既不能新登录，已有 session 的下一次身份 API 请求也会 401；即使绕过 principal 入口，租户角色查询也失败。 |
| 租户停用 | 平台用户可 PATCH tenant active（`identity.go:212-229`）；tenant 列表继续返回 inactive tenant（`identity.go:138-175`）。`scopedRole` 只验证 tenant 存在/用户 membership，不检查 `tenants.active`（`identity.go:99-106`）。 | 停用不是删除，也不自动移除成员或角色；成员管理入口对 inactive tenant 仍可达。 |
| 跨租户边界 | `scopedRole` 对非平台用户以 tenant ID、user ID 和 active 用户联合查询；失败在入口映射为 404（`identity.go:99-106,202-207`）。成员读写再以相同 tenant ID 约束目标（`identity.go:333,366,390-392`）。 | 非成员看不到目标租户是否存在，且不能用另一个租户中的管理员身份管理本租户。平台用户是显式全租户例外（`identity.go:101-103`）。 |
| 派生清理 | membership 删除通过复合外键 `ON DELETE CASCADE` 清理诊断成员；角色离开 developer 时触发器也清理（`console_storage.go:17-18`）。 | 成员移除/改角不会遗留诊断分配。直接合同测试源码覆盖改角后清理、删除后作用域撤销（`console_contract_test.go:609-639`）。 |

## 发现

### P1 — 停用管理员无法在仅有一个有效管理员时被清理或降权（确认缺陷）

触发条件：租户有管理员 A（`users.active=1`）和管理员 B（`users.active=0`），由 A 或平台管理员删除 B、或把 B 改为 developer/finance。

代码先确认 B 的 membership 角色是 administrator（`identity.go:365-377`），随后统计的是**全租户有效管理员总数**，而不是“操作后剩余的有效管理员数”（`identity.go:379`）。此时计数为 1（只有 A），`administrators <= 1` 导致 409（`identity.go:383-385`）。B 本来不在有效管理员集合中，删除/降权不会减少有效管理员数，因此该拒绝是错误的。

影响：管理员名册无法清理停用账号，停用身份会永久保留高权限角色标签；若其他表或未来流程以 membership role 而非 active 联合判断，风险会被放大。当前鉴权 `principal`/`scopedRole` 已阻止 B 实际使用该角色，所以这不是即时权限绕过，定为 P1 而非 P0。

建议修复条件：只在“目标用户当前 active 且为 administrator，并且操作后将不剩其他 active administrator”时拒绝；或直接按 `user_id<>目标` 统计剩余 active administrator 并在结果为 0 时拒绝。应同时保持事务与 `server.mu` 的现有串行边界。

### P2 — inactive tenant 是否允许继续改成员没有显式合同（缺陷假设）

`scopedRole` 不读取 tenant active（`identity.go:99-106`），因此 tenant 被平台停用后，现有管理员仍能列成员、改角、移除和自我退出；平台也能管理。若“停用”定义为冻结业务流量但保留治理/恢复能力，这是合理防护设计；若定义为冻结全部租户变更，则这是策略缺口。

影响取决于产品语义：最坏情况下，停用期间仍可扩大或重排管理员面。由于当前代码和直接合同仅明确了 active 状态持久化，未给出成员治理语义，记录为 P2 假设，不作为确认漏洞。建议把允许的 inactive-tenant 操作矩阵写入合同；若禁止扩权，至少区分恢复性收缩（删除/降权）与扩权（升为 administrator）。

### P2 — 用户 active 有执行语义但缺少本范围内的生命周期入口（确认合同缺口）

仓库 Go 源码中可见的 `users.active` 写入仅为初始化/新用户创建为 active；本范围未发现停用/恢复用户的 HTTP 管理入口。读取侧则严格执行 active（`identity.go:84-95,105`）。

影响：代码能够消费“停用”状态，却没有可审计的直接产品路径说明谁能改变它、停用前如何保证租户仍有管理员、恢复后 membership 是否原样生效。这也使上述 P1 场景难以通过公开合同构造，但并不消除数据库状态及迁移/未来入口的一致性问题。建议先补合同，再决定是否提供入口；不要用直接改库作为运营流程。

## 已有防护

- 平台管理员不是 membership 伪装：`scopedRole` 明确对存在的 tenant 返回 administrator（`identity.go:101-103`）；非平台用户必须具备本租户 membership 且 active（`identity.go:105`）。
- 非管理员跨租户访问在入口返回 404，不泄露租户存在性（`identity.go:202-207`）；直接合同源码也覆盖不同角色访问另一租户 members 返回 404（`console_contract_test.go:583-587`）。
- 角色更改即时从数据库读取；已有 session 不缓存 membership role。直接合同源码验证改角后租户列表立即回读新角色（`console_contract_test.go:609-615`）。
- 删除 membership 后，下次作用域查询失败；直接合同源码验证旧 session 对原租户资源变为 404、tenant 列表清空（`console_contract_test.go:627-635`）。
- 真实最后管理员自删返回 409 的直接合同已存在（`console_v3_test.go:218-225`；`console_contract_test.go:658-660`），但没有覆盖“目标是 inactive 管理员且另有一个 active 管理员”的反例。

## 未执行项与证据限制

- 按任务约束，仅进行了静态代码审查；未启动服务、未运行 Go/HTTP/浏览器测试、未执行 `make verify`，不作产品通过或运行时可复现声明。
- 未审查邀请 token 消费、密码建立、重复接受等流程；仅引用其 membership 插入点以界定新增成员来源。
- 未扩展审查通用控制台资源授权、应用 Key、账单或模型权限；仅引用 membership 外键/触发器及直接合同中与角色生命周期紧邻的证据。
- P1 的触发状态依赖停用用户已存在；当前公开入口是否能创建该状态不在可见合同中，因此报告同时明确记录用户 active 生命周期缺口。

