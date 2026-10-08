# 兼容迁移、恢复与正式运营契约

architecture-r1；2026-10-07；来源与摘要见 [architecture.md](architecture.md)。本文件是目标操作契约，不是已执行升级或部署回执。

## 1. 已有基线与必要变化

本checkout基线 `98511771871cf0951ecef716bbb55deb13e18da5` 已有schema1/2、SQLite WAL、数据目录排他锁、加密sentinel、Argon2管理员、应用Key摘要、多个候选、最多3个attempt及Go init/serve/reset-password/upgrade/backup/restore。现有 `session` 仅CSRF/Expiry，无人员或成员scope；现有requests可能有未知Tokens但无价格/tenant；现有版本2不能因此视为完整多租户基础。

现有 `upgrade --output` 是备份后原位事务迁移，`backup` 在独占目录锁下checkpoint后归档数据库，key-file分开保管，`restore` 要求新目标。继续扩展这些**正式Go入口**，不新造Node/Python服务，不改平台部署控制器/数据库。运营命令准确参数由研发实现后更新README；本设计沿用现有 `reset-password` 名称，而非另造第二个重置入口。

重要根因：当前openInstallation在验证后会将started请求改process_interrupted。新升级/备份必须将“只读校验/打开”和“服务启动恢复”分开，不能调用带恢复副作用的open后才备份。迁移前备份必须保存原始可恢复事实，不因预检查先改源库。

## 2. schema3兼容映射

| 原有资产 | 升级结果 / 不变量 |
|---|---|
| admin salt/hash | 复制为唯一bootstrap平台用户admin的密码事实，算法与旧密码保留；旧password-only登录兼容此用户；新账号登录进入同一会话身份路径；reset-password更新新事实、失效会话且留审计，不留两套可独立改变的认证源 |
| 原上游及密文、AAD、配置revision | schema2原ID/配置/候选/密文及AAD不变；schema1用已有受支持密文迁移，再升3；核验版本保持，旧核验不能证明新上界/价格能力 |
| 公共模型alias、候选 | 保持ID与enabled/路由语义；新增能力/价格状态标legacy-needs-review，不假造已核验收费能力；新增模型默认无授权 |
| app_keys digest/hint/id/状态 | 每条原Key成为稳定Key池，原digest成为其current代际；原Key字符串继续有效，不要求读回明文；期限未配置为无期限而非过期；启停原值不改 |
| 原Key授权 | 创建明确“原有应用”租户，bootstrap admin成为其active管理员；tenant及Key grant显式快照迁移时全部已有alias（保留原有既得范围），以后新增alias无自动grant；不把空集合当全模型 |
| 请求/attempt/route证据 | 保留Request ID、ordinal、状态、时间、Tokens、revision及用量未知字段；已知key_id映射原有租户；缺失Key归属的旧拒绝/历史仅平台可读，标legacy-unattributed，不猜其他租户 |
| 原有未定价历史 | 标legacy-unpriced/待核对；不填零价、不用今日价回算、不能关闭含这些项的周期。给平台正式“历史定价依据核对”入口，追加来源/历史价证据及快照，不覆写原记录；无凭据保持Not Run/待确认 |
| 原有计数/限制 | 无原有限额不伪造旧限额；从现有可归属请求保留数量/Tokens视图，无法判断的受理或dispatch标legacy-unknown而非补成0；原Key默认无费用硬上限、有限TPM，保持此前可用性；启用有界策略时必须核对当前周期历史负债，不能跳过待确认 |
| 已有schema3账本/更正/窗口 | 备份恢复完整保留ID、精确金额、snapshot、幂等键、释放标记及时间高水位；same-version upgrade仅验证返回already_current |

旧系统没有价格/账单并不等于费用为零；未知历史不进入“已确认费用”分母或成功结账。兼容调用可在原有无硬限制模式继续，UI明确预算保护能力尚未准备；设硬费用/有限TPM之前校验可信上界和该scope未决历史，不能声称绝对硬上限已实现。原有Key改授权只能通过正式管理入口， migration不是今后配置事实源。

## 3. 升级与失败恢复顺序

1. 运营者停止服务；upgrade获取同一目录排他锁，校验目标路径、源schema、key sentinel、完整性、外键、所需磁盘/目标输出不存在。未知schema、错误key、占用锁、备份目标已存在均安全失败。升级不能自动覆盖用户目录或凭据。
2. 只读读取源事实并记录脱敏迁移前计数/ID/金额/未知项摘要；不执行启动恢复。以现有一致备份机制写全新归档并fsync，备份不含key-file；key由运营者独立安全保管。备份失败则没有任何migration写入。
3. 单一写事务完成schema1→2→3或2→3及上述映射，运行凭据校验、integrity/FK及业务守恒检查。旧Key digest/ID、公共alias、请求/attempt数量、已知Tokens、费用/账单/更正和unknown数量均可核对；没有数据的字段保持null/未定价。meta version最后更新；失败rollback整个事务，备份保留。
4. commit后fsync/close，JSON回执含from_schema/schema/result和脱敏守恒计数，非秘密。未知提交结果通过再次只读schema+完整性+迁移ID检查确定，不盲目重复迁移。schema3 serve执行启动恢复后 `/healthz` 返回storage_schema=3与实际构建version；UI读身份/范围后才业务可用。
5. 旧二进制看到schema3必须明确拒绝，不写不认识字段。回退需停止新服务、保留新数据目录与失败日志，使用升级前正式备份恢复到新隔离目录及对应旧key、再用兼容旧二进制验收；不要就地“降schema”。升级后产生的新业务记录不能通过旧快照恢复假装无损，必须明确回退RPO边界及人工决定，不能自行毁弃新账本。

设计新增历史定价核对由平台管理员在待确认队列/请求详情发命令，输入旧请求/尝试、历史价格证据引用及价版本、已知/未知字段来源、原因、幂等键；通过同一核对/快照/账本事务追加，不能将“当前配置定价”作过去事实。若旧schema1无尝试证据，可追加标legacy推定的证据记录供读取，但在确认dispatch/价之前不把其标供应商已核验。必要人工核对保留而非造数据完成AC18。

## 4. 服务重启与未决用量

启动在目录锁内先验证schema/密钥/事务完整性，再用恢复事务扫描accepted/running请求及未结束attempt：逻辑执行终态记process_interrupted；attempt尚未确认的记录标未知。可证明从未dispatch候选释放资源；可能dispatch未知部分保留原预留。所有活跃本地并发持有归零/释放一次，并留恢复关联记录；不是清零计费/次数/RPM/TPM/历史。

已提交账本唯一来源键可重算验证；未应用的可信证据按差额补账一次。恢复事务失败serve不开始受理，不把坏数据库当空安装。迟到核对只补用量/费用，不修改已结束执行状态。没有可信证据的未知不能定时TTL释放；待确认队列显示最早时间、Request ID、残余预留、核对入口及对预算的影响，平台管理员负责持续处理。

## 5. 故障、发现和责任

| 事件 | 用户行为/影响 | 信号与隔离 | 恢复及责任 |
|---|---|---|---|
| 供应商连接/生成失败 | 当前候选失败，未commit可有限切换；已有输出不重放 | 同IDattempt与脱敏error，候选冷却；用量可能pending | 平台管理员改配置/核验；提供者负责远端；应用显式新调用 |
| 权限/预算/期限拒绝 | 明确401/403/429/恢复动作，零未经授权dispatch | 请求拒绝维度/安全scope；角色收窄即刻读取当前事实 | 租户管理员修授权/限额/Key，平台修tenant边界 |
| DB busy或写入失败 | 503/storage-degraded，新请求不dispatch；已有输出不伪报完整结算 | health降级、日志只记code/ID，保留预留 | 平台管理员查磁盘/锁/完整性，受支持恢复后核对；不手改库 |
| HTTP结果/管理响应丢失 | UI写入unknown，不重签、不把空读当成功 | operation安全读回/读回失败状态 | 同幂等键查询；无法再取秘密时显式轮换 |
| 取消/断流/进程崩溃 | cancelled/interrupted终态，费用可能待确认 | pending及恢复审计；本地C一次释放 | 正式证据核对，保守资源不透支；不后台重发 |
| 邀请泄漏/Key泄漏 | 撤销邀请或代际/Key，已消费资格需要正式移除 | 邀请/Key审计、不记录secret | 管理员撤销及交接身份；不自动撤其他应用 |
| 上游实际超声明上界 | 显示实际确认费用及预算超额，停止该候选有限预算新调用 | bound_violation风险/审计，不截断真实账本 | 平台与提供者核对上界/退款或更正证据；真实保证重新验证 |
| 迁移/恢复失败 | 安全失败，原库/备份保留，不“重新初始化解决” | 命令非零、阶段、脱敏校验原因 | 运营者正式restore到新目录验证；保留历次失败，不覆盖证据 |

默认日志只含Request ID、tenant/Key安全ID、阶段、error、status、duration、计量完整性和版本；无对话、token或敏感响应。审计对敏感配置、邀请、角色、授权、Key轮换/撤销、限额、价、核对、结账、更正、重置及恢复保留actor/scope/对象/旧新安全差异/time/operation。租户只看本范围可读事件，供应商秘密变更仅记“已变更”。

## 6. 部署和后续维护边界

本节点不运行正式服务、不接触运行中实例、不扩权限、不创建分支/commit/PR，不推送或触发Actions。研发需让README/正式命令帮助与实现同版。目录、key、生成密码、SQLite、会话、node_modules、个人缓存和绝对本机路径不能进仓库；验证证据需脱敏。

测试/部署须沿 `docs/local-deployment.md` 的现有main prepare→授权手动deploy路径。若控制器不支持schema3或新命令，研发提交产品兼容说明给协调者，由唯一平台维护源/模板升级后正式复验；不在产品写临时部署器、不手改平台数据库。最终由Connector/协调者核对同产品PR、merge SHA、Actions/Deployment、health版本和固定效果URL；不能拿8793编排端口或另起演示Go实例作为发布。
