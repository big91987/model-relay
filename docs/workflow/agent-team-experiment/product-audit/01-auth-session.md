# 登录、密码与会话生命周期静态审查

## 范围与结论

本报告仅静态审查 `internal/relay/server.go` 的 login/session/logout、`internal/relay/identity.go` 的密码规则与认证、`internal/relay/security.go` 的随机材料，以及直接合同和相关测试名称。未运行构建、测试、浏览器或服务，未读取真实凭据，也未改动实现或测试。因此下文“确认”表示源码或测试合同可直接确认，不表示运行通过。

结论：未发现可仅凭本次静态证据定为 P0 的认证绕过。密码散列、随机会话材料、CSRF、服务端过期和注销撤销均有明确实现。发现 2 项 P1 缺陷假设（Cookie 缺少 `Secure`；会话生命周期使用墙钟而非项目有效时钟）和 3 项 P2 缺陷假设（密码策略不一致、停用用户的令牌不主动清除、登录限流为进程级共享且成功也计数）。

## 确认事实与已有防护

### 密码规则、散列与认证

- 新邀请用户密码要求 12–1024 字节且不得包含 CR、LF、NUL（`internal/relay/identity.go:58-60`）；邀请接收失败使用明确的 `password_requires_12_characters`（`identity.go:471-478`）。该规则是长度规则，不是字符类别或已泄露密码规则。
- 新用户密码使用 Argon2id：独立 16 字节盐、3 次迭代、64 MiB、单 lane、32 字节输出（`identity.go:62-65`）。盐来自 `crypto/rand` 生成的 16 随机字节所编码字符串再做 SHA-256 截断；随机源仍提供 128 bit 原始熵（`internal/relay/security.go:20-27`）。
- 认证只查询 `active=1` 用户，计算相同参数的 Argon2id，并用 `subtle.ConstantTimeCompare` 比较（`identity.go:75-90`）。账号不存在时仍以固定哑盐和 32 字节哑散列执行一次 Argon2id，降低账号存在性计时差异；对外登录失败统一为 401 `unauthorized`（`server.go:330-338`），未直接泄露“账号不存在/已停用/密码错误”。
- 空账号映射为 `admin`；非 admin 账号在认证时做 trim 和小写规范化（`identity.go:75-81`）。数据库账号列唯一（`identity.go:15`），邀请账号在创建路径另有邮件地址规范化合同。
- 旧 schema 管理员密码也使用同一 Argon2id 参数及常量时间比较（`internal/relay/store.go:297-303`）。管理员离线重置同步更新 schema 3 的 platform-admin 散列，且注释明确依赖停服独占锁、重启清除内存会话（`store.go:263-294`）。
- 密码型工作设有 2 个并发槽，登录与邀请接受共享该限制（`server.go:313-318`; `identity.go:419-424`）；登录另有每进程每分钟 10 次尝试上限（`server.go:294-312`）。JSON 请求限制 64 KiB、要求对象、拒绝重复/未知字段（`server.go:89-109`）。

### 会话、Cookie、CSRF、过期与注销

- 登录成功分别生成 32 随机字节的会话 secret 和 CSRF secret，使用 URL-safe 无填充 Base64；`randomSecret` 直接调用 `crypto/rand.Read`，失败会 panic 而不会退化为弱随机（`security.go:20-25`; `server.go:340-353`）。两者各自约 256 bit 原始熵。
- 服务端仅以会话 secret 的 SHA-256 摘要作为 map key，不保存明文 secret；会话内容为 CSRF、绝对过期时刻和 user ID（`server.go:29-33,185-198,353`）。会话只驻留进程内存，重启自然撤销全部会话。
- 会话固定 8 小时，无滑动续期。登录时清理已过期会话并限制总数为 128；每次取会话再次检查绝对过期并删除过期项（`server.go:185-198,342-356`）。`/admin/api/session` 返回 CSRF、过期时间及（schema 3）重新查询得到的活动 principal（`server.go:223-233`）。
- Cookie 为 `HttpOnly`、`SameSite=Strict`、`Path=/admin`、`MaxAge=8h`（`server.go:355`）。所有管理响应还设置 `Cache-Control: no-store`、CSP、`nosniff`，并严格校验 Host（`server.go:124-130`）。
- 所有非 GET 管理请求要求 Origin 精确等于 `http://<Authority>`；登录也在该检查之后。除登录和邀请接收外，写请求还要求常量时间比对 `X-CSRF-Token`（`server.go:200-220`）。邀请接收对既有账号额外要求当前用户会话与目标 user ID 相同并校验 CSRF（`identity.go:458-469`）。
- 注销本身需要有效会话、可信 Origin 和 CSRF；之后按 Cookie secret 摘要删除服务端会话并发送同路径的删除 Cookie（`server.go:214-247`）。这构成即时的当前令牌撤销。
- schema 3 的活动用户校验并不只发生在登录：session 端点重新调用 `principal(active=1)`（`server.go:223-230`; `identity.go:93-97`），identity API 与 console API 入口也分别重新调用它（`identity.go:114-124`; `internal/relay/console_api.go:12-29`），其余 legacy 管理分派前亦有 principal 校验（`server.go:250-263`）。因此直接停用用户后，旧会话条目虽仍在 map 中，但静态路径显示管理请求会因活动用户重查而失去授权；成员移除也通过每次请求重新读取 membership/role 收回租户范围（`identity.go:99-106,327-399`）。

## 缺陷假设与影响分级

### P1 — 会话 Cookie 未设置 `Secure`

证据：登录 Cookie 和注销删除 Cookie 均设置 `HttpOnly`、`SameSite=Strict`，但没有 `Secure: true`（`server.go:246,355`）；Origin 合同还硬编码为 `http://`（`server.go:202`）。

影响：如果管理入口可经不可信网络直接使用 HTTP，浏览器会允许明文发送 bearer-like 会话 Cookie；同站 CSRF 防护不能抵御链路窃听或篡改。反向代理外部终止 TLS 时，应用仍需确保浏览器收到 `Secure` Cookie。若部署合同明确限定可信本机 HTTP，则风险可下调，但本次范围内没有足够静态证据证明该限制。

已有防护：`HttpOnly`、`SameSite=Strict`、严格 Host/Origin、CSRF、8 小时固定过期。建议后续明确 HTTPS/代理合同，并在生产模式设置 `Secure`（同时处理可信代理与开发模式），增加 Cookie 属性合同测试。

### P1 — 会话过期使用 `time.Now`，未使用项目有效时钟

证据：登录创建和读取会话均使用 `time.Now()`（`server.go:194-195,295,353-356`），而请求入口及邀请过期采用可持久化、防回拨的 `Store.now()`/有效时钟（`server.go:132-147`; `identity.go:403-406,449-455`）。

影响：系统墙钟回拨可延长会话实际有效期；配置的测试/有效时钟无法一致控制会话生命周期。墙钟前跳则可能提前注销用户。该问题不直接产生无限会话，但破坏项目已有的防回拨时间语义。

已有防护：绝对 8 小时截止、每次访问检查、登录时清理。建议后续把会话时刻来源纳入统一时钟合同，并用回拨/前跳测试验证；需谨慎评估进程内单调时钟与持久化高水位的组合。

### P2 — 密码策略在邀请用户与管理员重置之间不一致

证据：邀请用户要求至少 12 字节（`identity.go:58-60`），但 `ResetAdminPassword` 只要求 trim 后非空、合法 UTF-8 单行且不超过 1024 字节（`store.go:263-268`）。因此单字符管理员密码可通过离线重置合同。

影响：最高权限 platform-admin 可被配置为明显弱于普通用户的密码，降低离线/在线猜测成本。在线限流可缓解，但不能替代一致的密码强度下限。

已有防护：Argon2id、离线独占锁和 key file 要求、在线限流。建议后续统一最低长度并为重置边界增加弱密码拒绝测试；迁移/运维兼容性需单独评估。

### P2 — 用户停用不主动清除其内存会话

证据：`principal` 与权限查询会过滤活动用户（`identity.go:93-106`），但 session map 只在过期、注销或进程重启时删除（`server.go:185-198,236-247,342-347`）；当前审查源码中也未见面向用户停用的管理写路径。

影响：现有请求路径会动态拒绝停用用户，故这不是已确认的认证绕过；但令牌材料仍存活至 8 小时截止。如果未来新增某条只信任 `getSession` 而不重查 principal 的端点，残留令牌会成为易回归的授权风险；用户重新启用时，旧令牌也可能恢复有效。

已有防护：目前 schema 3 受审查路径重查 active，进程重启清空会话。建议后续在用户停用/密码变更事件上显式撤销该用户全部会话，或把用户状态/会话版本纳入 `getSession` 的统一认证边界。

### P2 — 登录限流为全局进程桶，成功与格式错误也消耗额度

证据：所有登录在解析请求或验证凭据前即写入同一个 `server.attempts` 切片，没有账号/IP维度，成功也不清除（`server.go:294-324`）。

影响：任意可达客户端可用 10 次请求使该进程所有用户最多约一分钟无法登录；多实例又各自独立，降低抗分布式猜测效果。这是可用性与防爆破精度问题，不是凭据绕过。

已有防护：严格 Origin 会阻止普通跨站浏览器直接触发，另有 2 个 Argon2 并发槽和 `Retry-After`。建议后续采用总量保护加账号/IP维度的组合策略，并保留统一错误以避免账号枚举。

### P0

本次静态范围内无已确认 P0。特别是：未发现密码明文存储、可预测会话随机数、注销后当前会话仍可直接使用、或停用用户在受审查 schema 3 路径继续获得权限的直接证据。

## 错误泄露审查

- 登录凭据失败统一返回 401 `unauthorized`，账号不存在和 inactive 也走同一外部错误；哑 Argon2 工作降低粗粒度时序枚举（`identity.go:84-90`; `server.go:336-338`）。
- JSON/字段错误统一为 400 `invalid_request`；并发槽和速率上限统一为 `login_rate_limited`。这些错误不会回显密码或账号（`server.go:313-326`）。
- session 无效、用户停用统一为 401 `unauthorized`；CSRF 和 Origin 分别暴露 `invalid_csrf` / `untrusted_origin`。这种区分不含凭据，但可向调用者区分请求边界失败（`server.go:200-228`）。
- 邀请的新密码过短返回具体策略错误，而无效/过期/停用租户邀请统一为 `invitation_unavailable`（`identity.go:449-475`）。已存在账号则返回 `existing_account_login_required`，会泄露该邀请账号已存在于系统这一状态；因调用者必须持有 256-bit 邀请 token 才能到达该分支，静态评估为低风险、未单列等级。
- `randomSecret` 的熵源失败会 panic（`security.go:20-24`）。这避免弱随机继续运行，但造成进程可用性失败；错误不会把随机材料写入响应。

## 直接合同与相关测试名称（仅静态列举，未执行）

- `TestSourceAndLoginLimits`：声明 Origin 拒绝、10 次失败后全局登录限流合同（`internal/relay/relay_test.go:271-299`）。
- `TestInProcessContract`：声明登录、CSRF 管理调用、注销后会话 401 的合同（`internal/relay/inprocess_test.go:19-46,121-126`）。
- `TestInstallationRecovery`：声明正确/错误密码隔离及恢复后密码有效（`relay_test.go:191-231`）。
- `newConsoleFixture` 的直接断言：声明登录响应过期时间至少约 7 小时（`internal/relay/console_test.go:25-45`）；它是 fixture helper，不是独立命名测试。
- `TestConsolePeopleScopeInProcess` / `TestConsolePeopleScopeActualHTTP`（共同调用 `peopleScopeScenario`）：声明角色变更在现有会话实时生效、移除成员后现有会话失去该租户范围（`internal/relay/console_contract_test.go:458-459,609-635`）。它们未直接覆盖 `users.active=0` 后的完整会话撤销。
- `TestConsoleIdentityAndKeys`：相关身份/邀请/登录综合合同名称（`internal/relay/console_v3_test.go:144`）。
- `TestConsoleInvitationRateLimit`：相关邀请密码工作与限流合同名称（`internal/relay/console_contract_test.go:315`）。

静态搜索未发现专门覆盖以下边界的测试名称或直接断言：Cookie `Secure` 属性；会话恰好在 8 小时边界失效；墙钟回拨/前跳下的 session 行为；用户 `active` 从 1 变 0 后所有管理端点及重新启用行为；密码重置与邀请密码最低长度一致性；多会话逐用户撤销；登录成功是否重置限流窗口；账号存在性细粒度时序测量。

## 未执行项与限制

- 按任务边界未运行任何 Go 测试、`make verify`、构建、浏览器或服务；上述测试仅阅读名称和断言，不能声称通过。
- 未做动态计时分析，无法量化不存在账号、错误密码、停用账号之间的剩余时序差异。
- 未检查真实部署的 TLS、反向代理、Cookie 改写或网络隔离配置，因此 Cookie 风险保留为缺陷假设。
- 未读取真实凭据、数据库或会话，也未验证生产中的用户停用操作是否由范围外系统触发。
- 未修改实现或测试；唯一产物为本报告。
