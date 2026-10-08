# 服务端与工作台实现契约

architecture-r1；2026-10-07；同 [architecture.md](architecture.md) 的Run、冻结材料和摘要。本文给研发明确推荐机制，正式业务规则以原包PRD为准。

## 1. 真实身份与授权

### 人员生命周期

- 新安装Go `init` 创建平台用户 `admin` 和随机一次性密码，沿用Argon2与现有凭据输出安全规则。增加用户表和用户状态，旧密码兼容详见迁移文档。平台管理员不是租户管理员的客户端标记。
- 登录输入账号/密码。会话只在服务端存随机token摘要、user ID、到期和CSRF，不把成员/角色长期复制进会话；每请求读取当前有效用户和成员。会话可沿用内存保存，重启要求重新登录，不丢成员事实。注销、密码重置、禁用用户删除其会话；角色降权/退出即使旧会话仍存在也立即失去权限。
- 租户邀请包含规范化邮箱、角色、到期、token摘要、状态。管理员创建后一次展示接受链接，通过受支持手工安全渠道交付；UI明确“待接受 / 未发送邮件”，不能标已邮件验证。邀请默认24小时，最长7天，服务端精确UTC判断 `now >= expires_at` 失效；可撤销并重邀。token不放URL查询/日志，链接fragment由接受页读入、立即清除地址并POST提交。
- 接受入口允许新用户凭有效邀请设定密码建账户并加入；既有同账号用户必须用自己的密码登录再接受，邀请token不能覆盖既有账户密码。接受事务校验邀请未使用/未过期/未撤销和邮箱对应，创建用户/成员及消费邀请原子完成。token持有仅证明本次受邀凭证，非外部邮箱所有权认证；重复接受返回安全的已接受事实，不再次授予角色，撤销或移除后旧token不能重加入。
- 成员主键tenant/user；修改角色、移除、自己退出、接受邀请均串行检查最后有效管理员。负责人展示派生自有效管理员和待接受邀请；invited不算有效管理员；先邀请并由另一管理员接受才能移除最后管理员。
- 管理员在Key编辑选本租户active开发者作为诊断成员；角色降为财务/成员移除在同事务删除诊断分配。应用Key生命周期由管理员显式控制，人员退出不改变Key状态。

### 资源权限矩阵

| 资源 / 动作 | 平台管理员 | 租户管理员 | 开发者 | 财务观察者 |
|---|---|---|---|---|
| 服务/模型/价格写、租户创建暂停恢复、平台上限 | 全局 | 无 | 无 | 无 |
| 公共可调用模型读 | 全局 | 本租户授权交集 | 本租户授权交集 | 本租户价格/元数据，不生成 |
| 成员/邀请/诊断分配、Key签发/编辑/轮换/撤销 | 全局审计下管理 | 本租户 | 无 | 无 |
| 配额预算写 | 平台上限/全部 | 本租户且不越平台上限 | 无 | 无 |
| Key元数据、调用日志、用量 | 全局 | 本租户 | 仅active分配Key | 本租户用量；无秘密/完整对话 |
| 账单/价格/导出 | 全局 | 本租户读 | 分配应用用量，不读租户账单 | 本租户读 |
| 用量核对、结账、更正 | 全局 | 无 | 无 | 无 |
| /v1生成 | 有效调用Key，不使用人员角色代替 | 同左 | 同左 | 不由财务角色自动授予调用Key |

平台用户可在UI选择全局或具体租户；租户成员只能选择自己的有效租户。租户暂停不剥夺成员的只读诊断/账单资格，但禁止签新Key、扩大授权及新的生成；恢复后按当前资格生效。平台仍能修复，租户管理员仍可收窄/撤销Key及管理退出。暂停不是隐式删除。

所有资源先按范围定位；越租户ID返回404，角色不足返回403，未认证401。范围拒绝不返回另一租户是否存在、数量、模型、价、Key或账单。聚合、CSV、游标和搜索与列表同一范围；API层而非隐藏菜单负责隔离。

### 安全约束

保留Host/同源Origin/CSRF/CSP、HttpOnly/SameSite、登录限速和密码计算并发保护；TLS配置时Cookie必须Secure，受支持loopback HTTP开发入口不伪称TLS。管理资源不接受任意X-Role/X-Tenant头；参数化SQL。凭据密文使用现有AES-GCM及AAD，secret只存摘要；不把完整Key、密码、上游秘密、提示词/回答写日志、审计、幂等响应、URL或浏览器存储。

服务接入沿用SSRF阻断和DNS解析/实际dial校验、禁重定向/代理。生产自托管接入默认使用可被现有策略接受的合法地址；私网端点仅在运营者显式配置精确host/port allowlist时允许，仍禁止metadata、回环管理端点、动态客户端URL和重定向绕行；此配置是运维授权，不由普通模型表单放开网络。受控测试沿用仅测试Authority，不带进产品默认策略。

高风险威胁路径：伪角色/跨租户ID→服务端scope与负向测试；邀请重放/账号接管→摘要、期限、一次消费和既有账号认证；Key泄漏→一次秘密、撤销/轮换和审计；SSRF→配置/dial双校验；费用重放/竞态→原子事务和唯一账本键；CSV公式注入→文本字段转义危险开头、数值保持精确；SSE/存储DoS→时限、大小上界、并发背压。主机管理员越权改库为本期剩余特权风险，不宣称密码学不可抵赖。

## 2. 控制对象的权威命令路径

新管理资源推荐前缀 `/admin/api/v1`；现有login/session/logout入口保留兼容并返回真实user/资格，现有安全配置及历史读入口仍需回归，不能让旧入口绕过新授权。旧全局写端点只能由平台管理员操作且进入相同新业务事务。

| 对象与事实源 | 正式UI / HTTP命令 | 校验、生效、变更和退出 |
|---|---|---|
| 用户/会话 | 登录页；现有login/session/logout；Go reset-password | Argon2、限速、会话到期；重置撤会话并审计；无前端角色切换 |
| 服务/核验 | 供应商与服务；`POST /services`，`PATCH /services/{id}`，`POST /services/{id}/checks` | 地址/能力/凭据校验；每次配置变更revision；核验绑定execution_revision；历史失败/过期留存；启停通过PATCH |
| 模型/候选/租户授权 | 模型托管；`POST /models`，`PATCH /models/{alias}`，`PUT /models/{alias}/tenant-grants` | 公共名独立远端名；新模型授权空；映射变化不自动加授权；能力/价格/有界策略明确才可发布；停用保留历史 |
| 租户/预算 | 租户与成员；`POST /tenants`，`PATCH /tenants/{id}`；配额页 `PUT /tenants/{id}/limits` | 创建预算和负责人邀请可同事务；暂无active负责人显示待加入；暂停/恢复不清零；平台上限仅平台写 |
| 邀请/成员 | 成员表；`POST /tenants/{id}/invitations`，`POST /invitations/accept`，`POST /invitations/{id}/revoke`；`PATCH/DELETE /tenants/{id}/members/{user}` | 期限/角色/最后管理员；消费原子；角色变更与诊断范围同事务；不删除应用Key |
| Key及诊断分配 | API Keys；`POST /tenants/{id}/keys`，`PATCH /keys/{id}`，`POST /keys/{id}/rotations`，`POST /keys/{id}/revoke`，`POST /keys/{id}/retire-previous` | 模型必须属本租户grant；诊断成员必须属本租户active开发者；启停/有效期/代际受当前控制池约束；撤销不可复活 |
| 价格 | 定价；`POST /models/{alias}/prices`，`GET /models/{alias}/prices` | 四费率USD微美元/百万Tokens；新增版本及生效时间；不能修改旧快照；输入展示测试价 |
| 请求/尝试/用量 | 日志/用量/调用测试；`GET /requests`、`GET /requests/{id}`、`GET /usage`；真实 `/v1/chat/completions` | 查读scope；生成只用调用Key；系统唯一创建请求/尝试；诊断不改生成结果 |
| 未知用量核对 | 请求详情待确认区；`POST /requests/{id}/attempts/{ordinal}/reconciliations` | 平台管理员；证据引用、权威累计输入/输出、原因、revision；补已知差额/解除有依据预留；无证据拒绝，无手改库 |
| 账单 | 定价与账单；`GET /tenants/{id}/bills`，`GET /bills/{id}`，`POST /bills/{id}/close`，`POST /bills/{id}/corrections` | 平台关闭/更正、租户读；结束且无pending；更正原因/原账单/调整周期/非负；记录不可覆写 |
| 导出/审计/工作台 | 对应页；`GET /usage/export`，`GET /requests/export`，`GET /bills/{id}/export`，`GET /audit`，`GET /overview` | SQL范围与列表一致；脱敏精确金额；页级错误独立；审计仅有权限的管理员读 |

请求模型alias保持稳定；数据库之外没有第二个业务对象配置入口。后台核对、重启恢复、运营命令同样调用域内方法并留审计。

接受邀请是唯一新增可在无会话时执行的人员加入命令：仍校验受信任同源Origin、大小/速率限制及一次token；已有用户流程还需有效会话和CSRF。其他新管理命令全部走会话及CSRF，不能把accept白名单扩成资源访问白名单。邀请元数据公开读只返回接受所需最小字段，失效token不泄漏成员名单。

价格发布体明确 `tenant_input_rate`、`tenant_output_rate` 和各候选的 `upstream_id/cost_input_rate/cost_output_rate`，简单单候选可使用四费率输入；UI按模型统一租户价、按来源显示独立供应商成本。每候选成本与统一租户价组合为一份受理快照，缺任一可用候选价格不能在调用中偷偷补当前价。候选成本不同不需要新公共模型或重复租户授权。

迁移历史价格核对的正式命令为 `POST /requests/{id}/legacy-reconciliations`，只允许平台管理员，载荷含历史价格/dispatch依据引用、可确认Token字段、reason及expected_revision；不能核对时返回明确待确认原因。这与attempt核对入口共用追加账本/审计事务，不是修改旧请求快照或SQL修补入口。

## 3. 数据、版本和幂等契约

在现有schema2上增加schema3，保留旧表ID及密文AAD；schema1先沿既有升级路径再升3。建议表职责如下，研发可按现有Go风格拆文件，但不可减少持久事实：

| 数据 | 约束/必要字段 |
|---|---|
| users / memberships / invitations | 用户密码hash及状态；成员tenant/user复合PK；邀请token digest UNIQUE、角色/到期/消费；有效资格读当前事实 |
| tenants / tenant_model_grants / key_diagnostic_members | tenant状态；grant tenant/alias唯一；诊断tenant/user/key FK保证同租户 |
| app_keys / key_generations / key_model_grants | 原Key ID保留，tenant、应用名、状态/期限、revision；代际digest UNIQUE、expires/retired；所有代际引用稳定Key ID |
| limits / quota_windows / reservations | owner tenant或Key及平台上限，nullable独立限制、mode、revision；窗口身份；request/owner唯一预留、资源维度、余额、释放标记 |
| requests / request_attempts | 原Request ID及ordinal保留；tenant/key、受理时间/周期、授权/配置版本；尝试dispatch可能性、执行终态、完整/部分/未知usage；价格快照及上界 |
| price_versions / request_price_snapshots | 四独立费率、USD、版本/生效；请求受理冻结所有可能候选价格，尝试选用其中一份，配置改价不更改它 |
| usage_reconciliations / charge_entries | attempt/evidence唯一；累计已知输入输出和可信来源、operator/time；账本delta、成本/租户费、原周期/快照版本/来源唯一 |
| bills / bill_corrections | tenant/UTC月周期唯一；closed_at、immutable original金额/明细范围；追加更正signed delta、原bill、调整周期、reason/actor/time |
| management_operations / audit_events / migration_meta | 主体/命令/Idempotency-Key唯一，payload digest、安全结果引用；不存秘密；审计与业务事务绑定；迁移版本、兼容来源、时间高水位 |

UTC时间持久化精确边界，JSON使用RFC3339；金额int64微美元，JSON/CSV以十进制整数字符串及人读USD显示，避免JS浮点丢精度；rate亦为整数微美元/百万Tokens。次数和Token计数为非负整数，NULL表示未知/无上限，0是明确拒绝，负数/溢出/混合null计费均校验失败。尝试complete必须满足输入/输出可确认且total一致；上游仅total不够拆输入输出费用。

管理写入采用 `Idempotency-Key`，命令scope含真实actor、tenant、动作及目标；同键同payload返回同一安全事实，不同payload409。Key签发/轮换事务只持久安全operation→Key/代际引用，秘密仅提交成功的首个响应有一次展示机会；双提交的其他响应永不返回秘密。响应丢失视为未知，`GET /operations/{key}` 先读回；确认写入后秘密无法找回，明确提示管理员显式轮换，不能重复创建。过渡中不可叠加，先显式结束旧过渡才能另起轮换。对未确认存在的命令只可用原键重查/同payload重试，不能改键制造重复。

保留幂等事实至相关业务记录退出保留期；本期不自动删除账本/安全操作索引。可逆编辑带 `expected_revision`，不符409且保留非敏感草稿。同一实体状态/权限变更与审计、幂等结果在同一事务，不用“先返回成功再保存”。秘密失败窗口仍不持久化明文/可恢复密文。

状态机：邀请pending→accepted/revoked/expired；Key active↔disabled、active→revoked、期限到达expired；代际current→previous→retired（上一代最多一份）。请求accepted→running→succeeded/upstream_failed/cancelled/timed_out/process_interrupted；零dispatch拒绝为rejected。请求终态单向，只允许计量状态 partial/pending→complete 与追加核对进展，不用迟到成功覆盖取消。幂等终结条件使用request状态及released字段CAS。

## 4. 原子限额与未知预留

### 口径和安全上界

次数是UTC月内受理的逻辑生成请求数量，RPM是UTC固定一分钟受理数量，TPM是该窗口各已dispatch尝试的Tokens及未释放预留；C是当前本地活跃逻辑生成数。tenant和稳定Key两个scope均检查，平台上限作额外tenant策略上界。限额未设是unlimited，0拒绝；Key与tenant不互相覆盖。请求次数/RPM受理后保留，不因取消或失败退回；零dispatch的认证/授权/配额拒绝不增用量。受理后因配置变更导致零dispatch仍是已受理计数，费用为零，UI应区分。

费用硬预算使用租户费而非供应商成本；tenant/Key同周期都承担该请求的租户费，不是把两scope相加作为收费。费用页同时显示成本。任意显式0上限均为禁止受理，包括软费用0；正值软费用超限告警继续，不能用软模式把0解释为无限。未知预留跨窗口仍占相应scope可用量，不因月界/分钟界重置成可花余额；确认后归原窗口/周期，解除carry但不抹原事实。降低硬上限至低于已用/预留可保存，剩余显示0和超限，新受理拒绝；正值软费用只警告；不截断旧账本。

硬费用或有限TPM必须在受理前有**可信的每尝试输入/输出收费Token上界**。模型服务配置增加有界协议策略：支持的生成字段、最大输入收费Tokens、最大输出收费Tokens、输出上限参数（max_tokens或max_completion_tokens）、用量能力、上界依据。最简安全实现预留完整模型输入上界而非把字符估算当精确tokenizer；输出采用校验后显式发送的上限。输入类型/工具/多模态/多choice若无法证明在上界内则在有限TPM/硬费用模式下零dispatch拒绝 `usage_bound_unavailable`，不悄悄走软预算；支持的普通文本/SSE应有可验证的有界配置及正式UI入口。

未给输出上限时服务使用模型已配置安全默认；显式超模型上限、n>1或不受支持字段明确拒绝，不能静默改变语义。服务核验与“远端遵守收费上界”的依据独立显示；一次连接成功不能认证绝对收费上界。任意第三方可能违反声明，产品只对声明且验证的有界契约兑现硬准入，不宣称能控制不可信远端的无限收费。受控真实HTTP可证明本项目预留算法；真实供应商上界缺证Not Run，不能自动标核验通过。无法界定未知部分时保留全部剩余预留，必须用可信核对关闭。

### 受理与执行算法

1. 在短写事务中读取当前身份Key代际、租户active、模型授权交集、服务候选/能力/上界/价格、所有限制和时间高水位。没有授权/有效价时零dispatch拒绝；有限TPM或硬费用缺安全上界时亦拒绝，记录安全拒绝原因。没有有限TPM且只有软/无限费用时可按已支持协议受理，无界未知显式记为unbounded，不伪造数值预留或硬保护。旧未定价Key的兼容调用仅遵循迁移文档的legacy模式，历史仍待核对。
2. 候选列表及最多3尝试计划固定；有界模式按每个可能尝试输入/输出上界计算租户费上界，分别四舍五入后求和。Token预留求所有尝试上界和；不同候选价格从同一次受理快照取值。软/无限模式有界时仍保守预留，无界时记录未决责任状态而不填0；转硬/有限TPM之前必须核对此scope未决无界责任。无需猜将来改价或无限重试；未纳入受理计划的新候选不加入本请求。
3. 先验证所有scope：count+1、RPM+1、C+1、TPM已用+活跃/未知预留+新预留、周期费用已确认+未决预留+新预留均不越各自上限；软费用只告警继续。全部通过才一次提交请求、各scope预留、窗口增量和价格/授权快照。SQL并发控制使用短事务、唯一键/CAS及条件更新；BEGIN IMMEDIATE或现有单连接事务均需并发测试证明无lost update。禁止部分scope先扣后返回429；事务中不做HTTP。
4. dispatch前再检查停止/撤销和候选可用，先持久化attempt可能dispatch标记与快照，再发上游。禁止先发远端后补记录。提交到网络调用之间崩溃视为可能dispatch未知，不伪称零费用；能证明未dispatch的候选才释放对应预留。
5. 已知/部分用量按尝试累计可信字段计算账本差额；原子将已知预留转换为已确认，未dispatch/确认无用量部分才释放；未知余额继续保留。执行终态/并发持有CAS释放一次，计费与执行终态解耦。费用变化不增加第二份count/RPM。
6. 网络全部在事务外；预算竞争失败返回429及安全scope/维度/上限/恢复动作；DB busy超过既有限时返回503，不伪装预算耗尽。取消不会阻止持久化终态：用有界内部恢复context而非已取消请求context写事实；写失败全局storage-degraded阻止新dispatch并保留恢复待办。

预留守恒：每scope的费用已确认 + 未知/执行中预留 + 可用余额 = 上限（超限保存/供应商违约时可用余额取0且展示超额）；它是资源视图，不是第二次收费。每attempt账本合计等于按快照算出的累计确认费用；每request等于全部attempt合计；tenant等于所属request及明确调整项之和；Key代际都计同一Key。新旧Key并发竞争不增加应用池。可用余额不因软件重启、暂停/恢复或时间回拨增加。

TPM按实际已知Tokens+预留保证保守，不声称估算等于真实。持久时间高水位采用 `effective_now=max(wall_now,last_now)`，只允许窗口前进；期限、过渡和账单周期也用同一有效UTC时钟，回拨不给已过期秘密重新生效。持续大幅时钟漂移提供告警与运维修正指引，不手改窗口。已dispatch但本地进程终结只释放本地C，远端未知消费仍由预算/TPM预留保护；不声称该C可强制供应商停止计算。

## 5. 双费用、核对、关闭和更正

四个费率为 cost_input/cost_output/tenant_input/tenant_output，单位整数微美元每百万Tokens。每个尝试每一侧总费用：`round_half_up((input_tokens * input_rate + output_tokens * output_rate) / 1000000)`，再汇总尝试；部分已知字段先计算已知部分，待完整后用最终累计费用减先前累计已记费用，不把两次舍入简单相加。中间乘加使用有溢出检查的整数/大整数；费率、Tokens和金额绝不用float。正好半微美元向上舍入；例：rate=500000、input=1、output=0，结果1微美元。成本与租户价分别计算，任何一边未知仍明确显示待确认。

受理冻结价格版本及四费率，尝试从其候选快照取价。调用成功/失败/取消不影响“全部dispatch已知用量均收费”；零dispatch无费用。明细含累计原用量、证据来源、四费率、价版本、确认费用、未知维度与预留，不把上游敏感body作为核对凭证保存。

平台核对入口输入attempt、权威累计input/output、evidence_reference、reason、expected_usage_revision和Idempotency-Key。证据可以为本次受控上游凭据化收据或供应商安全用量记录引用，必须注明证据类型，不允许将运营者估算认证为供应商事实。用量新增记录，不覆写原证据；不同凭证相同累计量delta为0，相同凭证不同payload409；较小累计量或闭账后的原用量改写拒绝，走更正。NULL→0必须有可核验“未发生”证据，不提供“忽略未知并结账”按钮。业务核对和账本/预留调整同事务；输出终态保持。

账单周期为UTC自然月，UI显示精确开始/结束及本地可读时间。读账单派生自同一追加账本；关闭事务检查 `effective_now >= period_end`、该周期无任何待确认/活跃请求和价格缺失迁移项，保存原金额/明细界限/关闭人时间。tenant+period唯一，同键重试或重复close返回同一closed事实，不扣第二次费。并发关闭与核对/更正有版本校验，不能读完pending=0之后在事务外关闭。

更正仅对已关闭账单追加signed delta，必填原因、原账单和一个未关闭调整周期（默认当前UTC月）、operator/time，幂等唯一。原金额与原请求不变；同原账单的累计 `original + corrections >= 0`，并发负调整串行校验。原账单界面“当前应计=原金额+更正合计”是追溯视图，调整周期财务账本只记录更正一次，跨周期汇总不得把追溯后的原账再加一次更正。更正作用于租户调整周期预算的signed金额，Key实际请求费用不伪改；正调整可产生预算超额并阻止后续受理、负调整不改变Token/次数。外部支付始终标未接通，不提供已支付成功状态。

## 6. HTTP与调用契约

管理JSON含安全 `request_id`、版本/范围及结果；错误含code、字段错误或安全恢复动作，不透出上游敏感错误。GET列表page_size默认25、最大100，返回total及当前过滤条件；排序稳定（时间+ID）。游标/offset不能跨scope复用，切scope重置到首页。导出同过滤/授权，超过公开限制返回明确分批指引，不返回被静默截断的“完整导出”。

`GET /v1/models` 只返回active tenant、有效Key授权交集和enabled可调用模型；`POST /v1/chat/completions` 在同一次准入快照判断。无效/过期Key401，模型无授权403，不可调用能力400，已暂停/停用明确错误，资源超限429，无安全上界/配置未准备503；均无未经授权上游访问。新逻辑调用由服务端生成Request ID，客户端可用单次调用标识防UI双击；重复生成不自动重放完成内容（默认不存内容），只返回已有安全Request ID/状态。不能用同Idempotency-Key启动第二次有费用生成。

普通/SSE沿用已有Go协议实现，确保增量真正flush。stream usage以供应商支持能力请求include_usage，缺final usage为pending；收到[DONE]不自动补零。透明协议字段仅在支持策略内保留，预算无法包住的字段明确拒绝。已有最多3尝试路由在输出commit前可切换，所有dispatch计量；第一字节/帧已commit或上游生成不可安全重试后不再重放。按现有超时/冷却规则有界退出，不无限重试。

客户端Abort/连接断开触发Go context取消；SSE流后错误按既有事件/终止契约表现，不能把中断当完整成功。后端记录取消/未知供应商费用并给Request ID；UI清晰区分“本地停止接收”和“上游费用可能待确认”。后台恢复不自动重发prompt。

## 7. A工作台与前后端职责

沿用原包十页导航和J01–J07，侧栏226px、顶栏72px、8px间距、主区域最大1640px，明亮背景/白业务表面/低饱和蓝主动作；正式UI去掉演示角色切换、B/C开关、模拟耗尽与静态成功数字。账户菜单只读当前服务端角色/租户资格。运行版本及帮助是次级信息，首屏仍以业务任务为中心。

工作台的调用趋势、租户费用分布、健康和风险来自真实聚合及核验；缺数据为空/未知，读取失败不是0。供应商/模型页给当前revision核验、价格/能力及授权入口；成员页给真实邀请接受链接状态；Key页给诊断选择、未知写入读回和过渡提前撤销；预算页给零/无限、已确认/预留/剩余及硬软模式；定价账单页给四费率、核对/关闭/更正；日志页给多尝试及价版本。不能用说明文字替代这些动作。

调用测试手动粘贴已持有调用Key或接续当前一次性值（仅内存且需显式操作），不从后端再取秘密，不使用人员会话生成未授权调用。开发者只可诊断自己被分配应用，不能借调用测试拿到其他Key。提示词/输出仅在当前内存，刷新不恢复；离开/退出/权限变化清除。

弹层使用真实可访问dialog或等价焦点管理，打开聚焦、Tab圈定、Escape/关闭回触发点；详情抽屉同要求。写入状态editing/invalid/submitting/succeeded/failed/unknown；忙时防双击，明确失败保留非敏感草稿，unknown禁盲目重签/计费，只提供安全读回。秘密离开弹层、切租户、登出、刷新即清，无local/sessionStorage恢复。

每一次scope/身份/导航变化递增scope epoch，abort旧读取并在响应写入前比较epoch+user+tenant+query；退出开始立即失效所有请求和敏感DOM，迟到响应不得重新填回。后端401统一清身份，403重新读取资格并清受影响诊断数据；读取失败保留“读取失败/重试”，不展示过期数据为当前事实。安全页面路径/非敏感筛选可恢复，秘密和对话不可恢复。

1280/1440保持多对象紧凑表格/详情；390折叠导航，表格局部滚动，页面不溢出，核心操作不能仅因secondary-action样式消失，改为明确更多菜单/可触达操作。图表有范围/时间/未知口径及文本汇总，不以颜色唯一表达状态。所有写入口同时满足键盘与触屏，按冻结原型逐页对照而非只看工作台截图。
