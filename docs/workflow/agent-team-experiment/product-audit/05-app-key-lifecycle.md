# 应用 Key 生命周期与授权静态审查

## 结论与范围

本报告仅基于源码和直接合同做静态审查，未运行测试、服务或数据库，未读取或生成真实 Key，也未修改实现。审查对象为 Schema 3 受控 Key 主路径、`server.go` 的旧 schema 兼容路径、相关 schema 及直接合同。Bearer scheme/空白 parser 已由 `strict-bearer-auth` 专案处理，本报告明确排除该问题，不复述其结论，也不据其测试结果宣称本任务动态通过。

未发现可仅凭当前静态证据定为 P0 的 Key 接管、跨租户授权或明文持久化缺陷。确认 1 项 P1 权限缩减持久性风险：平台撤回租户模型 grant 时未清除对应 Key grant，日后恢复租户 grant 会让旧 Key 自动重新获得模型权限。另有 1 项 P2 兼容面差异：Schema 3 以前的兼容 Key 仅有可逆的 enabled 开关，不具备受控 Key 的租户、模型、到期、轮换和不可逆撤销语义。

## 确认事实

### 凭据材料、摘要与一次性展示

- `randomSecret` 使用 `crypto/rand.Read`，并以无填充 URL-safe Base64 编码；应用 secret 为 `mr_` 加 32 个随机字节，随机部分具有 256 bit 原始熵，最终长度为 46 字符（`internal/relay/security.go:20-27`；`internal/relay/console_keys.go:235-237,352-353`）。随机源失败会 panic，而不是降级到弱随机。
- 持久化的是 SHA-256 摘要和前 11 字符 hint；`app_keys.digest` 与 `key_generations.digest` 均不保存明文（`internal/relay/security.go:27`；`internal/relay/console_keys.go:235-241,352-353`）。对高熵随机 secret 使用无盐摘要不形成现实的离线枚举弱点。Schema 对 generation digest 施加 UNIQUE，避免相同凭据映射多个 generation（`internal/relay/console_storage.go:13`）。
- 创建成功仅在 201 响应返回明文；相同幂等操作重放只返回对象 ID、`secret_recoverable:false`，元数据扫描只包含 ID/name/hint/enabled/created、控制信息、grants 和 generation 元数据，不查询 digest 或 secret（`internal/relay/console_keys.go:67-80,217-241`）。轮换同样只在首次成功响应返回新明文，幂等重放不恢复 secret（`internal/relay/console_keys.go:310-317,352-362`）。直接合同也静态断言创建重放与元数据读取不含 Key（`internal/relay/console_v3_test.go:175-188`）；本任务未执行该合同。

### 租户归属和管理授权

- `key_controls.key_id` 一对一绑定 `tenant_id`；模型 Key grant 以 `key_id` 为主体，诊断成员同时携带 tenant，并通过 `(tenant_id,user_id)` 外键绑定 membership（`internal/relay/console_storage.go:12,16-18`）。
- 所有受控 Key 路由先解析请求路径中的 tenant scope；对象读取后再次要求 `key.Tenant == tenantID`。写操作仅租户 administrator 可做；developer 只有在被显式分配为 diagnostic member 时才能读该 Key，越权对象统一隐藏为 404（`internal/relay/console_keys.go:101-119,168-180,183-210,248-255`）。
- 创建前验证目标 tenant 存在且当前 principal 在该 tenant 有 scope，并要求 tenant active；创建、控制记录、generation、初始 limits、Key grants、幂等记录和审计在同一事务提交（`internal/relay/console_keys.go:168-173,211-241`）。
- 模型和 diagnostic grants 的输入逐项验证归属：模型必须已 grant 给同一 tenant；诊断用户必须是同一 tenant 内 active developer；重复项拒绝（`internal/relay/console_keys.go:122-148`）。

### 创建、改权、启用、到期、轮换与撤销

- 创建允许显式设置 enabled、Key-level expiry、模型 grants 和 diagnostic members。启用且已到期的输入会被拒绝；过期判断使用有效存储时钟 `Store.now()`（`internal/relay/console_keys.go:25-33,122-131,226-237`）。
- 更新采用 expected revision 乐观并发控制；已撤销 Key 不可更新。name/enabled/application/expiry 和完整 grant 集在同一事务替换，revision 同步递增（`internal/relay/console_keys.go:258-294`）。
- tenant 暂停后只允许收敛：不得从 disabled 变 enabled、不得增加模型 grant、不得延长或移除既有到期时间；缩减模型集合、禁用或缩短到期仍可进行（`internal/relay/console_keys.go:281-290,365-388`）。
- 认证查询要求 generation 未 retired、Key enabled、control 未 revoked；随后同时检查 generation grace expiry 与 Key expiry，解析失败或 `now >= expiry` 均拒绝（`internal/relay/console_keys.go:403-418`）。
- 轮换创建全新 256-bit secret；旧 generation 最多保留 24 小时 grace，零小时立即 retired；活动 grace 未结束时拒绝叠加轮换，过期旧 generation 会先 retired。`retire-previous` 可提前终止全部带 expiry 的旧 generation（`internal/relay/console_keys.go:301-353`）。
- revoke 在一个事务内设置 `revoked=1`、禁用 app key、retire 全部 generations 并递增 revision；随后任何 generation 都无法通过认证，且 revoked 状态阻止后续更新或轮换（`internal/relay/console_keys.go:319-328,358-362,403-406`）。

### 模型授权的执行点

- 有效授权是 tenant grant 与 Key grant 的交集，不是任一集合单独生效。模型列表用双表 EXISTS 过滤；生成入口通过 `authorizeModel` 复查交集与 tenant active；在计量 admission 前又复查 generation 身份和同一授权交集，降低配置并发变化导致已撤权请求继续进入计量/上游的窗口（`internal/relay/server.go:524-550`；`internal/relay/console_keys.go:421-434`；`internal/relay/accounting.go:127-145`）。
- 生成路径在早期识别后，在 snapshot 前再次调用 `appID`；Schema 3 admission 再次调用 `generationIdentity`，所以禁用、撤销、到期或轮换 retirement 不是只在请求最初检查一次（`internal/relay/proxy.go:19-32,120-150`；`internal/relay/accounting.go:136-145`）。

## 缺陷与缺陷假设

### P1：租户模型 grant 撤回后，旧 Key grant 会在日后自动复活（确认的状态机缺口）

平台更新模型的 tenant grants 时，仅删除并重建 `tenant_model_grants`，没有同步删除或失效 `key_model_grants`（`internal/relay/console_api.go:186-227`）。运行时交集检查能保证撤回当下立即拒绝，因此这不是即时越权；但旧 Key grant 仍作为潜伏状态存在。若平台稍后把同一模型重新 grant 给该 tenant，交集重新成立，所有过去曾获该模型且其 `key_model_grants` 未被租户管理员主动改写的 Key 会自动恢复访问，无需 Key 管理员重新授权或 revision 变化（交集见 `internal/relay/console_keys.go:421-434` 与 `internal/relay/server.go:545-547`）。

影响：临时撤权、合规隔离或产品下架后重新开放模型时，历史应用可能在管理员无感知的情况下恢复能力；Key 数量多、存活期长时爆炸半径为该 tenant 内所有残留 grants。它不跨 tenant，也受 Key enabled/revoked/expiry 和 tenant active 继续约束。

建议防护：明确产品语义并选一项落地。若“撤回即永久缩减”，在删除 tenant grant 的同一事务中删除该 `(tenant,alias)` 下所有 Key grants，并为受影响 Key 增 revision/写可定位审计；恢复 tenant grant 时不自动恢复。若产品有意采用“暂停式 tenant gate”，则 UI/API 必须明确提示恢复会复活多少历史 Key，并提供确认和审计。补充静态合同对应的 HTTP 回归：撤回后 403/列表隐藏，重新 grant 后仍 403，直到租户管理员显式给 Key 加回 grant。

### P2：旧 schema 兼容 Key 缺少受控生命周期能力（确认的兼容差异）

Schema 3 禁止旧的全局创建入口并引导使用 tenant-scoped keys；但旧 schema 兼容路径仍可创建只有 `app_keys` 记录的 Key，管理面只能切换 enabled，认证也只检查 digest 与 enabled（`internal/relay/server.go:397-427,429-455,511-523`）。这些 Key 没有 tenant ownership、模型 grants、Key expiry、generation rotation 或不可逆 revoked 状态。迁移到 Schema 3 时，旧 Key 会被归入 `legacy` tenant，并自动获得所有现有模型的 tenant 与 Key grants（`internal/relay/console_storage.go:37-42`）。

影响：仍运行旧 schema 的部署无法表达最小模型权限、强制到期或不可逆撤销；disable 后可再次 enable。迁移时的全模型赋权保持旧行为兼容，但也保留宽权限，须在升级运营流程中显式盘点和收敛。该项不代表 Schema 3 主路径可绕过。

建议防护：将旧 schema 接口视为受限兼容面，文档标注其安全能力差异并推动升级；升级后默认生成 legacy Key 清单，要求管理员逐 Key 分配真实 tenant、缩减模型、设置到期或撤销。若仍承诺旧 schema 安全维护，应至少增加不可逆 revoke 与到期机制，而不是把 enabled=false 当作撤销。

### P2 假设：摘要比较的常量时间保证在 Schema 3 不可见

Schema 3 通过 SQLite `WHERE g.digest=?` 定位摘要，没有应用层常量时间比较（`internal/relay/console_keys.go:403-407`）；旧 schema 路径查询后另用 `subtle.ConstantTimeCompare`（`internal/relay/server.go:519-523`）。数据库 BLOB equality 是否产生可远程利用的前缀时序差异，本次静态证据不足，且 256-bit 随机 secret、HTTP/SQLite 噪声与只接受完整格式显著降低现实可利用性，因此仅列防御性假设，不定为确认漏洞。

建议防护：若威胁模型要求可证明的常量时间校验，可先用固定长度索引/lookup key 定位候选，再在 Go 中对完整 SHA-256 摘要做 constant-time compare；在无性能与泄漏证据前不建议为此牺牲索引查询或扩大数据库扫描。

## 已有防护汇总

- 高熵 CSPRNG secret；数据库仅存摘要/hint，管理读取不返回 digest 或明文。
- 创建与轮换采用不可恢复的一次展示及幂等重放抑制。
- tenant/object 双重 scope、角色检查、developer 细粒度诊断分配与 membership 外键清理。
- revision 防并发覆盖；创建、改权、轮换、撤销、审计和幂等状态均采用事务边界。
- enabled、revoked、双层 expiry、generation retirement、tenant active、Key grant 与 tenant grant 交集在请求链多处复核。
- tenant 暂停时仍允许权限收缩，但禁止重新启用、扩 grant 或延长有效期。

## 优先级

| 优先级 | 结论 | 处置 |
|---|---|---|
| P0 | 未发现可由静态证据确认的 P0。 | 无。若动态验证发现跨租户或撤销后仍可调用，应立即提升。 |
| P1 | tenant grant 撤回不清除历史 Key grant，重新 grant 会自动复活旧权限。 | 明确“永久缩减/临时 gate”语义；优先采用事务内清除 Key grants并增加回归。 |
| P2 | 旧 schema 兼容 Key 仅有 enabled，不具备租户、模型、到期、轮换、不可逆撤销；迁移初始为 legacy 全模型。 | 文档化差异、推动升级并在迁移后强制盘点收敛。 |
| P2（假设） | Schema 3 摘要 equality 的常量时间性质未由应用层保证。 | 仅在威胁模型或测量证据支持时加二次 constant-time 校验。 |

## 未执行项与证据限制

- 按任务授权未运行任何 Go 测试、`make verify`、HTTP 服务、浏览器、SQLite 查询或时序实验；直接合同仅作为源码证据阅读，不代表本任务执行通过。
- 未接触真实 Key、真实数据库、备份、供应商或生产配置；未验证随机源运行环境、响应经代理后的日志/缓存行为、数据库文件权限、备份加密或 TLS 终止。
- 未动态验证创建/轮换响应是否被外层访问日志记录，未验证并发撤权与在途 SSE 的终止语义，也未测量摘要查询时序。它们需要独立、隔离的后续任务授权。
- Bearer scheme、大小写、前后/附加空白及空凭据 parser 行为明确不在本审查范围；依赖专案已修改该边界，本报告不重复其已知结论。
