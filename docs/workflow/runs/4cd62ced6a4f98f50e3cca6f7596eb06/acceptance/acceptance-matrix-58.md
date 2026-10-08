# QA58 当前逐项分层验收矩阵

design-v0.2.0 / records-detail-entry-r1；HEAD98511771871cf0951ecef716bbb55deb13e18da5；134件fb4066fa37d1e7865d3a18f72dd78fef98ae79bde17e20875af373b24771c3a6。archive5f51d35fdc15ed7214749f9413879b4238097f1d6b6bbe2cc0d508d8e103cef7。完整源文件、材料路径与摘要见evidence/attempt-58/integrity.json。

这是**当前QA权威矩阵**，acceptance-matrix-47.md及Run根研发追踪矩阵继续保留历史状态，不把其旧Open/Fail移植为当前结论。本轮不改根研发矩阵，因为它属于已测试134件指纹；验证索引指向本文件。

证据别名：H57=本Runtests seq57原make verify完整回执+evidence/attempt-58/host-seq57-review.md；J58=user-journeys-58.md及evidence/attempt-58/host/continuous-{1280-68740,1440-68790,390-68850}.json；V58=visual-review-58.md及同host/visual-{1280,1440,390}/review.json/PNG；S58=host/management-summary-{1280,1440,390}.json；F58=host/fault及editor-interruption三视口JSON；Q58=qa-static-node-final.log及qa-inprocess-final.log；G=internal/relay现有Go测试在H57的执行结果；M=H57正式CLI schema1/2→3迁移三视口。H57是宿主实际执行，Q58是QA亲自执行，不互相替名。Static Review不替代UI-E2E。

## 产品AC01–AC18

| ID | 基线要求 | 角色/范围 | 起始状态 | 连续用户/实际测试步骤 | 最终事实 | 证据类型 | 结果 | 证据位置 |
|---|---|---|---|---|---|---|---|---|
| AC01-UI | A十页、三视口、多对象/表格/统一弹层/键盘 | 四真实角色/两租户 | 空安装到多对象 | 登录→创建→导航→编辑/详情→键盘/移动菜单 | native操作可完成、根不溢出、秘密清除 | UI-E2E | Pass | H57/J58/V58/F58；QA58-01 P3另列 |
| AC01-design | 批准主要摘要和动作不删 | 平台及当前tenant | 两模型/两租户/五Key（跨租户合计） | 原模型/服务/租户/工作台路径查看→上下文启停/调用/管理 | 来源/能力/价/延迟、核验时间/revision、实际负责人/预算/状态与API同值 | UI-E2E / Data Reconciliation | Pass | S58/V58；QA47-01 Closed |
| AC02-UI-provider | 服务接入失败恢复、核验、发布 | 平台/受控provider | 无目标服务/模型 | 非法URL→保存失败保留公开草稿→修正→核验→发布 | 无半对象、秘密清除、核验当前适用 | UI-E2E | Pass | J01/H57；非供应商联调 |
| AC02-API-source | provider/self_hosted、配置变更使旧核验无效 | 平台/受控HTTP | 旧revision/未核验 | 两来源保存→核验→修改/晚到→再次核验 | 来源持久、旧结果不能覆盖当前 | API-E2E / Contract | Pass | H57 GovernanceActualHTTP/ServiceCheckRevision/EvidenceConfigurationRace |
| AC02-UI-self-hosted | 首次self_hosted网页独立分支 | 平台 | 空服务 | 本轮未单独执行该分支 | 无该分支当前UI回执 | UI-E2E | Not Run | 与provider共用正式表单+两来源API是内部门槛等价补强，不把本行改Pass |
| AC03-UI | 新模型默认无授权，tenant/Key交集 | 平台/租户管理员 | 新模型/空授权 | 发布→授权→Key→收窄→403→修正→显式再调用 | 拒绝零dispatch、授权不隐式扩大 | UI-E2E | Pass | J01/J02/H57 |
| AC03-API | 新身份/映射组合变更安全 | 应用/两scope | 有旧授权配置 | 直接越权/空grant、映射/候选变化后调用 | 无自动增加模型权限，零dispatch拒绝 | API-E2E / Contract | Pass | H57 Governance/IdentityAndKeys/G |
| AC04-UI | 真实邀请接受/角色/诊断分配/退出 | 四角色/两tenant | 空租户/新身份 | 邀请→接受→登录→分配→收回/降权/移除→最后管理员保护 | 人员失权、应用Key仍服务，最后管理员409 | UI-E2E | Pass | J01/J05/J58/H57 |
| AC04-API | 跨tenant读改详情/搜索/分页/导出拒绝 | 四真实身份 | 有应用/日志 | 直接ID/旧session/查询与CSV越权 | 403/404不泄另一scope及未分配应用摘要 | API-E2E | Pass | H57 PeopleScopeActualHTTP/SummariesActualHTTP；Q58补强非HTTP |
| AC05-UI | 暂停所属全部Key，新请求拒绝，其他tenant独立 | 两tenant/多Key | 非零已用 | 暂停→普通/SSE拒绝→Beta成功→恢复 | 拒绝零dispatch、恢复不补发 | UI-E2E | Pass | J05/H57 |
| AC05-DATA | 暂停恢复不清历史额度 | 同两scope | 已有调用费/pending | 前后读取对账 | calls/费用/预留/终态保持 | Data Reconciliation | Pass | H57 PauseUsageActualHTTP/J05 |
| AC06-UI | 一次秘密、关闭/刷新/退出不回显 | 租户管理员 | 无Key | 签发→一次值→调用→关闭/切scope/刷新重登 | 安全metadata保留、秘密不恢复 | UI-E2E | Pass | J01/J03/J07/F58 |
| AC06-API | 安全存储与重启不可重新取秘密 | 人员/应用 | 已签发Key | metadata/持久库只读/重启恢复读取 | 摘要存储、no-store、不重放/泄漏 | API-E2E / Contract | Pass | H57 IdentityAndKeys/ProcessCrashRecovery/安全回归 |
| AC07-UI | 立即/1h/24h/提前撤销/期限轮换 | 租户管理员/同应用 | 已用Key | 轮换→新秘密→关闭→旧/新调用→409叠加保护→结束过渡/撤销 | 同稳定Key历史及共池不清零 | UI-E2E | Pass | J03+console立即轮换；H57三视口 |
| AC07-API | 截止前/等/后、最长24h及代际共享 | 两代秘密/并发 | 显式clock及资源池 | 边界/非法过渡/8争2名额 | 精确失效、应用池共享无部分扣计 | API-E2E / Contract | Pass | H57 GenerationExactDeadline/SharedResourcesActualHTTP |
| AC08-UI | 写后丢响应/双提交/读失败安全恢复 | 管理员同context | 签发/轮换未提交 | 双Enter→unknown→GET503/404/200读回 | mutation1、不重签、不假成功/重放秘密 | UI-E2E | Pass | F58三视口/H57；历史红重建另标 |
| AC09-API | 双scope次数/RPM/TPM/并发/预算独立0/null，共池 | 两tenant/多Key/两代 | 矩阵/并发隔离夹具 | 各维命中与修复、8争2 | accepted2/denied6/other_tenant1，另一scope不部分扣计 | API-E2E | Pass | H57 ResourceDimensions/ResourceMatrixActualHTTP/SharedResourcesActualHTTP |
| AC09-UI | 预算有效上限/已用/耗尽与恢复 | 平台/租户 | 非零已用 | 预算设已用→429→提高→显式再调用 | 已用不清、同请求计量 | UI-E2E | Pass | J04；不把此链当全部五维独立UI分支 |
| AC10-API | 硬预算原子有界预留、软告警、未知保守 | 应用/有界无界候选 | 有可信bounds/故障 | 并发准入→多dispatch→取消/未知/核对 | 未知不作零/不释放可花余额，按实际追加守恒 | API-E2E / Data Reconciliation | Pass | H57 Admission/Accounting/QuotaBillingAndSSE/G |
| AC11-API | 改限额/轮换/跨窗/回拨/重启/迟到守恒 | 应用/运维 | 已用/pending | crash子进程恢复/时间边界/迟到重复终态 | 不二次扣计释放，不更改既定终态 | API-E2E / Contract | Pass | H57 ConsoleRealProcessCrashRecovery/ProcessCrashRecovery/G |
| AC12-UI | 独立成本/租户价、币种单位/版本 | 平台/租户 | 旧价调用 | 发新价→新旧同ID详情→账单 | 旧fee14/cost7、新fee21/cost7与快照一致 | UI-E2E | Pass | J06/V58/S58/H57 |
| AC12-DATA | 整数微美元精度/舍入及快照 | 费用账本 | 边界/半微美元 | 精度测试→dispatch/更价→读旧事实 | 不溢出、不回算，尝试舍入汇总 | Data Reconciliation / Contract | Pass | H57 MicroDollarPrecision/G；Q58 |
| AC13-API | 全已dispatch尝试已知收费、未知待确认、幂等 | 平台/应用 | 已知失败→未知→成功同ID | HTTP/SSE多attempt→核对→重复/迟到 | 成本/费分别追加，原attempt/终态/价快照不改 | API-E2E / Data Reconciliation | Pass | H57 Accounting/SummariesActualHTTP/G；Q58 |
| AC14-UI | 当前/pending拒结，结束平台结账/更正/导出 | 平台/财务/两tenant | 旧请求+真实unknown | 拒结→跨月→网页核对→结账→+5更正→详情/CSV | 原金额/原因/调整期保留，不泄另一tenant | UI-E2E | Pass | J06/H57，clock非生产日期 |
| AC14-DATA | 结账与更正幂等/金额非负/明细对账 | 平台/费用账本 | 可关闭周期 | 重复/负更正/对账/跨scope | 一次追加、不覆原账单和请求 | Data Reconciliation / API-E2E | Pass | H57 Accounting/G/J06；Q58 |
| AC15-UI-operation | 筛选/分页/导出/同ID/迟到及读错恢复 | 租户管理员/平台 | UI产生多页日志 | 同ID→页2→筛CSV→切tenant迟到→503→重试/刷新 | scope不覆盖、错误非空0、秘密不恢复 | UI-E2E | Pass | J07/H57 |
| AC15-UI-meaning | 原始与已确认Token同一事实可理解 | 平台/current tenant | unknown待确认请求 | 同ID网页核对3/4→records精确行详情→usage/CSV | 原始总量仍未知、确认3/4/pending0、无重复fee | UI-E2E | Pass | J06-reconciled-token-meaning/V58，QA47-02 Closed |
| AC16-UI | 普通/SSE/取消/同ID/双发送一次 | 管理员/应用Key | 有效Key | 输入→发送→输出/取消或拒绝→同ID | 不自动重放、不伪称取消免上游费 | UI-E2E | Pass | J01/J02/J04/J05/F58/H57 |
| AC16-DATA | dispatch/准入/attempt守恒及首块后不重放 | 受控HTTP/SSE | 延迟/断流 | 双击/取消/路由故障后读计数 | double普通/SSE均proxy/admission/attempt/dispatch1，3+4Tokens | API-E2E / Data Reconciliation | Pass | F58/H57 SSE/routing |
| AC17-API | 秘密/完整对话不泄，审计、服务端身份 | 四角色/应用Key | 敏感输入/非法scope | HTTP来源/伪角色/CSV/存储拒写/审计 | 无秘密完整对话落日志，敏感写可追责 | API-E2E / Contract | Pass | H57 DeniedMetadataAndStorage/PeopleScope/安全回归；Q58补强 |
| AC18-CLI | 安装/schema1/2→3/旧身份历史/backup restore | 运维/旧应用 | 固定版本旧UI夹具 | 正式CLI升级→native控制台/调用→备份恢复及失败路径 | 身份/历史/额度/映射守恒，失败不破坏原库 | UI-E2E / API-E2E / Contract | Pass | H57 M三视口/MigrationConservation/InstallationRecovery；非实际正式部署 |
| QA-local-fixed-entry | 原生QA监听环境 | 当前沙箱 | 已知限制不变 | 未再次启动必败入口 | 无提权/绕门禁；H57同版替代 | Contract | Blocked | QA22限制历史；本行不改Pass |

## H01–H08及外部事实

| ID | 要求/责任 | 证据类型 | 结果 | 当前证据或未决项 |
|---|---|---|---|---|
| H01 | Issue版本包、可访问SHA/entry | Contract | Pass | 会话issue25/prepare、QA58 archive/19件重算 |
| H02 | 原Issue→唯一Run，重复入站不重建 | External State | Not Run | 正向关系可见，平台重复事件演练未提供 |
| H03-material | 各阶段实际读取同基线与规则 | Contract | Pass | 需求/architecture/研发56/QA58材料路径、角色/预算/费用/AC和具体缺口 |
| H03-runtime | 实际Codex/gpt-6.1-sol凭证 | External State | Not Run | 配置声明不是原生模型可核对回执 |
| H04 | 同基线交接，缺文件/错SHA/过期安全恢复 | External State | Not Run | 正向材料一致；负向Harness回执缺失，未执行包脚本 |
| H05-fixed-entry | Agent候选源码与固定完整门禁 | Contract | Pass | development56正式交接、H57完整门禁、QA58独立指纹与Q58；非协调者代写实现 |
| H05-design | 设计主要信息/动作与实际实现相符 | UI-E2E | Pass | S58/V58/J58；QA47-01当前原层关闭，P3另列 |
| H06 | 独立检出→返原层→同版复验 | Contract | Pass | QA22/35/38/47历史缺陷、development多轮、seq57/QA58关单；旧No-Go保留 |
| H07 | 原产品PR/base/head/merge/Actions/Deployment/health/实际页 | External State | Not Run | 未发布；report→publish→pr、Connector/协调者负责闭合，不能用8793或旧main替代 |
| H08 | Harness唯一维护源/模板升级/原入口复验 | External State | Not Run | 平台完整事实未提供；产品修复不替代平台交付 |
| EXT-provider | 真实供应商/真实已有推理端点响应与报价 | External State | Not Run | 无凭据/响应；controlled_http_not_supplier不改本行 |
| EXT-identity/payment | 外部身份/真实支付联调 | External State | Not Run | 内部真实身份/USD测试账本是本期；外部不默认承诺 |
| DEPLOY-current-product | 正式latest页面/admin/和/healthz同版 | External State | Not Run | 等真实Deployment environment_url与merge SHA，非临时测试服务 |

## 商用能力五层对照与未来差距

对照范围沿用冻结references.md和用户已指定New API/LiteLLM官方资料的能力主题；本轮没有读取其实现/本机资料，没有重新扩需求或声称全功能对齐。

| 能力 | 已实现/Static Review | 局部测试 | 宿主完整测试 | 当前独立QA | 已发布 |
|---|---|---|---|---|---|
| 服务/渠道接入核验 | Go/UI两来源/revision/关联摘要 | 193Node/源码合同；非独立self_hosted UI | Pass H57 | provider UI及来源API Pass，真实供应商Not Run；P3枚举 | Not Run |
| 模型发布/授权 | 公共目录/交集/路由/价格/启停摘要 | Summary/授权回归 | Pass H57 | J01/J02/summary/设计主要结果Pass | Not Run |
| Key创建/轮换/撤销 | 一次秘密/共池/unknown读回 | 合同/边界/fault | Pass H57 | J03/F58原层Pass，历史缺陷Closed | Not Run |
| 身份/成员/租户隔离 | 内部真实身份/邀请/责任/暂停摘要 | QA PeopleScope race Pass | Pass H57 | 多身份UI及越权API Pass；外部身份Not Run | Not Run |
| 额度/预算/计费 | 双scope五维/保守预留/快照/追加账本 | QA Accounting/Precision/Summary Pass | Pass H57 | J04/J06及数值守恒Pass，测试USD非供应商报价 | Not Run |
| 用量/日志/诊断 | 同ID多尝试/筛选分页/CSV/确认Token | QA ExactRecords/Summaries Pass | Pass H57 | J06/J07原层Pass，QA47-02Closed；P3技术字段 | Not Run |
| 表格/编辑/详情/移动 | A现代明亮后台/统一弹层/三视口 | 193Node含键盘/所有权 | Pass H57/63visual PNG门槛 | 30页及关键图独立对照，AC01Pass/P3断词 | Not Run |

冻结外的支付充值/退款税务、GPU算力部署、其他模型协议、多区域HA/跨实例协调、生产性能/SLO、外部身份及自动秘密分发仍是后续差距；不因参考商用产品自动纳入本Run，也不声称其已完成。当前产品候选Go with known issues可交report；全部AC/H最终退出及实际页面交付仍待责任方正式事实，未执行行不隐藏。
