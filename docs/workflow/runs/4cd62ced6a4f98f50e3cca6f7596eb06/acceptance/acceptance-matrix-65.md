# QA65 当前分层验收矩阵

候选budget-clock-r1，冻结design-v0.2.0；HEAD66f1328ada75ea90e5638fbf2bd58663ac86347e + 未提交工作树。140件源戳ca8ec50aa168cf16ac2e24edfe214aa44d06a05846c34deadbc3c7cb6c02904a；材料archive5f51d35fdc15ed7214749f9413879b4238097f1d6b6bbe2cc0d508d8e103cef7。QA65为当前权威矩阵，QA58/47保留历史，不沿用其版本作为当前通过证据。

证据：H64=宿主seq64新完整make verify及evidence/attempt-65/host-seq64-review.md；G64=该回执当前Go实际测试；J64=host/continuous-{1280-14624,1440-14640,390-14654}.json；V64=host/visual-{1280,1440,390}/review.json及SHA匹配PNG；Q65=qa-inprocess.log/qa-node.log；I65=integrity.json；以上host/与日志均在evidence/attempt-65/。H64为宿主执行，Q65为QA亲自执行，不能互换。UI行Pass表示本轮宿主真实浏览器执行+QA证据审查，不表示QA原生亲自浏览器执行。

## 产品AC01–AC18

| ID | 基线要求 | 角色/范围 | 起始状态 | 连续用户/实际测试步骤 | 最终事实 | 证据类型 | 结果 | 证据位置 |
|---|---|---|---|---|---|---|---|---|
| AC01-UI | A十页/三视口/统一表单详情/键盘 | 四角色、两tenant、多对象 | 空安装→真实多对象 | 登录/创建→编辑/详情/导航/键盘/移动菜单 | 实际操作可完成，敏感值清除，根布局不溢出 | UI-E2E | Pass | H64/J64/V64，P3另列 |
| AC01-design | 批准A主要层级与管理摘要动作 | 平台/当前tenant | 两模型/两tenant/五Key | 逐页查看/上下文启停/调用/管理 | 当前管理事实/价格/负责人/状态与API一致 | UI-E2E | Pass | H64 management-summary三视口 |
| AC01-visual-QA | 当前页面静态设计对照 | 当前tenant/平台 | 当前采集完成 | QA查看390工作台、1280模型、1440额度图/30页结构化事实 | 明亮A布局和主要结果保留；长名字断词仍在 | Static Review | Pass | V64/host-seq64-review；只补强，非实操 |
| AC02-UI-provider | 接入/保存失败恢复/核验/发布 | 平台/受控provider | 无目标服务/模型 | 错URL→失败保留公开草稿→修正→核验→发布 | 配置与核验区别，失败无半对象 | UI-E2E | Pass | J01/H64；非真实供应商 |
| AC02-API-source | 两来源与核验revision/晚到保护 | 平台/受控HTTP | 未测/旧配置 | provider/self_hosted保存→核验→修改/晚到→重核验 | 来源持久，旧核验不冒充当前适用 | API-E2E | Pass | G64 GovernanceActualHTTP/ServiceCheckRevision/EvidenceConfigurationRace |
| AC02-UI-self-hosted | 独立self_hosted首次网页分支 | 平台 | 空服务 | 本轮未独立操作来源分支 | 无独立该分支UI回执 | UI-E2E | Not Run | 共用表单provider链+两来源HTTP补强；用户允许本候选先发布，原行不改Pass |
| AC03-UI | 新模型无隐式tenant/Key授权 | 平台/租户管理员 | 空授权 | 发布→授权→Key→收窄→拒绝→修复→显式调用 | 授权交集成立，拒绝零dispatch | UI-E2E | Pass | J01/J02/H64 |
| AC03-API | 新身份/映射组合变化不扩大授权 | 应用Key | 既有授权 | 越权/空grant/候选修改后调用 | 403/零dispatch，不自动扩权 | API-E2E | Pass | G64 Governance/IdentityAndKeys |
| AC04-UI | 邀请/接受/角色/分配/退出 | 四身份/两tenant | 首次成员 | 邀请→接受登录→分配→收回/降权/移除→最后管理员保护 | 身份失权，应用Key仍归tenant，最后管理员409 | UI-E2E | Pass | J01/J05/H64 |
| AC04-API | 跨tenant读改详情/搜索/导出拒绝 | 四身份/诊断scope | 多应用/请求 | 直接ID/旧session/查询/CSV越权 | 无另一tenant或未分配应用元数据泄漏 | API-E2E | Pass | G64 PeopleScopeActualHTTP/SummariesActualHTTP |
| AC05-UI | 暂停所属全部Key新调用，其他tenant可用 | 两tenant/多Key | 有已用额度 | 暂停Alpha→普通/SSE拒绝→Beta成功→恢复Alpha | 新拒绝零dispatch，恢复不补发 | UI-E2E | Pass | J05/H64 |
| AC05-DATA | 暂停恢复保留旧事实 | 两scope | 已有历史/pending | 读取前后对账 | calls/费用/预留/终态不清零 | Data Reconciliation | Pass | G64 PauseUsageActualHTTP/J05 |
| AC06-UI | 受限Key一次值 | 租户管理员 | 无Key | 签发→一次值→调用→关闭/换scope/刷新重登 | 秘密不重新取回 | UI-E2E | Pass | J01/J03/J07/H64故障链 |
| AC06-API | 安全存储、重启不重放秘密 | 人员/应用 | 已签发 | 元数据/持久库核对/重启 | digest/no-store/无完整值重放 | API-E2E | Pass | G64 IdentityAndKeys/ProcessCrashRecovery |
| AC07-UI | 立即/过渡/撤销/期限且历史共池 | 租户管理员同Key | 非零已用历史 | 立即/1h/24h轮换→新旧调用→叠加409→提前撤销/期限 | 一次值，共享应用池/历史，无叠加过渡 | UI-E2E | Pass | J03/H64 |
| AC07-API | 精确截止/回拨不复活/最长24h | 应用Key | 旧代际未retire | 截止前/等/后→回拨→新Key | 旧值始终401/零dispatch、新值可用 | API-E2E | Pass | G64 GenerationExactDeadline/ReviewClockRollbackActualHTTP/ClockRecoveryActualHTTP |
| AC08-UI | 签发/轮换未知回执安全恢复 | 管理员 | 正常对象 | 写后丢响应→unknown→503/404/200读回→明确再轮换 | mutation1/secret不重放，409防叠加，不伪报成功 | UI-E2E | Pass | H64 console-faults三视口 |
| AC09-UI | 可查看调整有效限额和拒绝 | 租户管理员 | 非零已用 | 查看→降低硬预算→429→提高→显式重发 | 当前两scope显示真实上限/已用/剩余 | UI-E2E | Pass | J04/H64 |
| AC09-API | 双scope五维/零无限/多Key竞争 | tenant/稳定Key/两tenant | 隔离并发夹具 | PUT限额→并发调用→收据与资源核对 | 全部scope原子受理，无部分扣计，共池不串tenant | API-E2E | Pass | G64 ResourceMatrixActualHTTP/SharedResourcesActualHTTP/ResourceDimensionsActualHTTP |
| AC09-zero-soft | 0软预算仍禁止受理，NULL/正软不变硬 | tenant和Key分别 | 已定价有授权 | PUT budget0/hard=false→付费及免费chat→NULL→正soft | 0为429/零dispatch/费用预留不变；NULL成功；正soft超限警告 | API-E2E | Pass | G64 ReviewZeroSoftBudgetActualHTTP原层；Q65非HTTP补强 |
| AC10-API | 保守硬预算预留/未知/取消 | 两scope | 有限预算/输出上界 | 原子预留→多尝试/取消/已知/未知→核对 | 已确认+未知预留守恒，无重复扣费或释放 | API-E2E | Pass | G64 ActualHTTPQuotaBillingAndSSE/ResourceMatrix/SharedResources |
| AC10-DATA | 费用与保守预留是不同事实 | 账本/资源池 | 有实际请求 | 比较attempt/charge/admission资源汇总 | 费用只记一次，未知不视零 | Data Reconciliation | Pass | G64 Accounting/QuotaBillingAndSSE/J06 |
| AC11-API | 修改/轮换/跨窗/回拨/重启不清旧事实 | 应用/两scope | 已耗额度/到期秘密 | 跨窗→回拨→子进程Open→Backup/Restore→旧新调用 | UTC高水位/月份不退，秘密不复活，calls2/fee14/reserved0/pending0保持 | API-E2E | Pass | G64 ClockRecoveryActualHTTP/ReviewClockRollbackActualHTTP/GovernanceActualHTTP |
| AC11-queue | 锁等待过截止再鉴权 | 模型目录/邀请 | 请求排队未到期 | 先取锁→请求到达→推进deadline→释放锁 | models401/invite400，零dispatch/零calls | API-E2E | Pass | G64 ClockQueuedDeadlineActualHTTP，Q65排队补强 |
| AC11-worker-top | 私密状态辅助worker顶层 | 专用子进程 | 无环境状态文件 | 常规全仓顶层调用 | 此单条顶层skip未执行专用场景 | Contract | Not Run | G64顶层SKIP；父ClockRecovery两模式实际子进程+restore绿为同轮替代 |
| AC12-UI | 独立成本/租户价/版本/币种 | 平台/财务 | 原价历史 | 查看→发布新价→新旧请求明细对比 | USD测试计价，旧价快照不回算 | UI-E2E | Pass | J06/H64 |
| AC12-precision | 微美元整数/舍入/溢出 | 双费用 | 边界费率与Tokens | 小数/半微美元/大整数/累计差额 | 与约定精度舍入一致 | Contract | Pass | G64 MicroDollarPrecision/Accounting，非供应商报价 |
| AC13-API | 同ID已知失败→未知→成功、多attempt | 受控HTTP | 失败/未知上游 | 多尝试→核对累计→重复/迟到 | 每dispatch快照双费，未知pending，差额幂等/终态保持 | API-E2E | Pass | G64 QuotaBillingAndSSE/Accounting/路由回归 |
| AC13-UI | 可区分原始未知与确认用量 | 平台 | 真实unknown请求 | 日志同ID→核对3/4→原行详情/分析/CSV | 原attempt不覆写，确认3/4/pending0一致 | UI-E2E | Pass | J06/H64；QA47-02当前回归 |
| AC14-UI | 拒当前/待确认结账，合法结束账与追加更正 | 平台管理员，财务只读反例 | 当前/pending→已结束 | 两次拒结→核对→关闭→+5更正→详情/导出 | 原额不变，调整周期记一次，责任正确 | UI-E2E | Pass | J06/H64 |
| AC14-API | 结账/更正幂等并发/跨tenant/禁止负透支 | 账单/两tenant | 结束/未结束/待确认 | 重复close/correction→金额与权限核对 | 账本与明细总额一致，无越权或重算旧事实 | API-E2E | Pass | G64 Accounting/QuotaBillingAndSSE/PeopleScope |
| AC15-UI | 真实筛选分页详情/错误恢复/CSV | 当前身份/scope | 超一页请求 | 跳同ID→页2/筛选→导出→切tenant迟到→503重试 | 当前scope明确，错误非空成功 | UI-E2E | Pass | J06/J07/H64 |
| AC15-API | 列表/汇总/CSV范围和确认事实一致 | 四身份 | 多页/unknown | 直接请求与非法scope、确认前后核对 | 正确原始/确认语义，拒绝无权限事实 | API-E2E | Pass | G64 OverviewAndExactRecords/SummariesActualHTTP/PeopleScope |
| AC16-UI | 普通/SSE/取消/双击/重进 | 当前应用/真实UI | 已授权 | 普通→SSE增量→取消→详情/恢复→双提交 | 不输出后重放，不声称取消免供应商费，单dispatch | UI-E2E | Pass | J02/H64故障与路由UI |
| AC16-API | 失败/取消/SSE终止边界和重启 | 受控上游 | 执行中请求 | DONE/断流/长行/取消/崩溃恢复 | 终态守恒，unknown保留，输出后不重放 | API-E2E | Pass | G64 SSE/RealProcessCrashRecovery/路由HTTP |
| AC17-API | 服务端权限/审计/不泄对话秘密 | 所有人/Key | 恶意输入或非法scope | 伪role/Origin/Host/CSV/存储拒写/审计核对 | 后端身份判断，管理事件可追责，无完整秘密对话日志 | API-E2E | Pass | G64 DeniedMetadataAndStorage/PeopleScope/安全回归 |
| AC17-UI | 退出/晚到隔离/秘密清除 | 当前context | 当前敏感编辑/调用 | 退出/切scope/取消旧editor/重登 | 旧响应无权写新界面，秘密不恢复 | UI-E2E | Pass | J07/H64 editor-interruption；193Node仅补强 |
| AC18-CLI-UI | schema1/2→3旧身份历史可用 | 运维→真实人员/旧应用 | 固定旧UI公开夹具 | 正式CLI upgrade→native登录/Key/pending历史→调用 | 历史/映射/旧身份守恒，新控制台可用 | UI-E2E | Pass | H64 production_cli迁移两分支×三视口；非正式部署 |
| AC18-recovery | schema3兼容Open/备份恢复/失败保护 | 隔离持久库 | 旧或当前schema | 标准安装/Open/升级/backup restore→拒绝故障 | 原账本不清，高水位恢复，无破坏原库 | Contract | Pass | G64 MigrationConservation/ClockRecovery/InstallationRecovery/Upgrade；Q65补强 |
| QA-native-UI | QA自己实际浏览器和完整入口 | 原生受限环境 | 已知监听限制 | 未重复必败make verify/无动态工具 | 无独立浏览器实操证据；不提权 | UI-E2E | Blocked | 本轮H64同版真实完整UI作为发布门槛替代，本行仍Blocked |

## 连续核心旅程执行契约与当前结果

步骤契约承接既有user-journeys-58.md，不修改其历史结果；下表只给本轮新H64/J64结果。共同清理：runner隔离目录/随机端口、受控上游，context finally关闭；没有人工SQL/API补UI步骤。只读API为最终事实核对，首次CLI初始化/测试clock为明确准备。

| Journey ID | User Story/AC | Actor | Start State | UI Steps | Expected UI Result | Final Fact | Cleanup | Result |
|---|---|---|---|---|---|---|---|---|
| J01-UI | US01/02/03，AC02/03/04/06 | 平台真实会话→受邀负责人真实会话，分别scope | CLI首次空安装/无目标对象 | 登录→服务非法URL失败/修正保存核验→发布两模型→建两tenant授权/预算→邀请→新身份接受登录→Key/首次调用 | 无隐藏预填，正式入口闭环，一次秘密、可修正错误 | 对象/授权/身份/审计/调用真实，零隐式授权 | 共同清理 | Pass |
| J02-UI | US03/05，AC03/09/16 | 同租户管理员context | J01真实应用 | 普通调用→SSE→取消→同ID诊断→收窄授权拒绝→恢复显式发送 | 输出/Request ID/拒绝恢复连贯，不自动重放 | 零dispatch拒绝、历史保留 | 共同清理 | Pass |
| J03-UI | US04，AC06/07/08/11 | 同租户管理员/稳定Key | 非零已用/历史 | Key编辑轮换立即/1h/24h→关闭秘密→旧新调用→叠加拒绝→结束旧代/撤销/期限 | 过渡明确，一次值不恢复 | 同应用池、旧历史保持；精确UTC/回拨另API行 | 共同清理 | Pass |
| J04-UI | US05，AC09/10/11 | 同租户管理员scope | 非零费用 | 额度查看→硬预算降至已用→调用429→提高→显式发送 | 可理解有效上限与拒绝原因 | 拒绝无上游/费用，修改不清零 | 共同清理 | Pass |
| J05-UI | US02/04，AC04/05 | 平台/两tenant负责人按正式职责交接 | 已接受身份/多Key | 角色分配/收回/移除→保护最后管理员；暂停Alpha→全Key普通/SSE拒绝→Beta成功→恢复 | 真正身份失权/另一scope可用 | 人员离开不盲撤应用Key，历史额度保持 | 共同清理 | Pass |
| J06-UI | US06/07，AC12/13/14/15 | 平台管理员同context；财务只读独立反例 | 原价/unknown请求 | 新价→对新旧明细→当前期拒结→测试跨月→pending拒结→网页核对3/4→原日志行native详情/分析/CSV→结束结账→+5更正详情/导出 | 未知与确认分明、责任/恢复可操作 | 旧fee14/new21/成本独立；原账不改、调整期守恒、幂等不重费 | 共同清理，clock只限隔离测试 | Pass |
| J07-UI | US07，AC15/16/17 | 当前真实用户同context/scope | 超一页记录 | 精确详情跳日志→分页/筛选/导出→切tenant隔离迟到→读取503→重试→刷新/退出重登 | 错误非0条、不混scope、不恢复秘密 | 读改CSV仅当前授权，晚到隔离 | 共同清理/释放测试GET故障 | Pass |

三个视口上述每行分别通过（H64逐journeyJSON、J64各七checkpoint）；J01/provider不是独立self_hosted首次UI；J04不是把API五维并发所有组合说成UI全组合。QA自己同会话实操仍Blocked。

## H01–H08、外部和发布事实

| ID | 要求/责任 | 证据类型 | 结果 | 当前证据/缺口 |
|---|---|---|---|---|
| H01 | Issue25包版本/实际SHA/入口 | Contract | Pass | prepare回执/会话输入/I65 archive+19件核对 |
| H02 | 唯一Run/重复事件不重建 | External State | Not Run | 原Issue/Run关系可见，未提供重复事件全套演练 |
| H03-material | 阶段实际同基线读取/关键规则 | Contract | Pass | requirements/architecture/研发63/本报告具体预算/期限/费用/AC与材料路径 |
| H03-runtime | 真实原生Codex/gpt-6.1-sol可核对身份 | External State | Not Run | 配置声明不当执行凭证 |
| H04-positive | 同批准基线正式交接和材料摘要 | Contract | Pass | development63→tests64→QA65/I65/实际路径 |
| H04-negative | 缺文件/错SHA/过期安全失败与恢复 | External State | Not Run | 缺平台负向实证；未执行包脚本 |
| H05-product | 执行Agent候选与固定入口/设计主要结果 | Contract | Pass | development63源码、H64完整测试、I65和本QA；不是协调者代写实现 |
| H06 | 独立检出→原层整改→同版复验 | Contract | Pass | 历史QA22/35/47闭环及本轮P1-1/2真实HTTP红→绿，独立QA65 |
| H07 | 唯一PR/merge/Actions/Deployment/health/页面同版 | External State | Not Run | PR27当前旧HEAD已知；新候选还需提交更新/合并/正式部署，不能用5545配置或8793代替 |
| H08 | Harness唯一源/模板安装升级及原入口复验 | External State | Not Run | 本节点不代做平台实现，完整平台事实缺失 |
| EXT-provider | 真实供应商/已有自托管端点现货与报价 | External State | Not Run | 无凭据/供应商响应，不以受控HTTP替代 |
| EXT-identity/payment | 外部身份/支付事实 | External State | Not Run | 内部身份/USD测试账本与外部事实分开 |
| DEPLOY-latest | 最新正式效果页面/health与merge对应 | External State | Not Run | report/publish/pr/协调者按verified-main正式Actions闭合 |
| CLEANUP-host-independent | QA独立查宿主无残留监听/进程 | External State | Not Run | runner清理声明与代码不是独立进程表证据 |

## 商用功能五层与未决差距

承接用户已指定官方能力主题，不重新扩冻结范围、不读取New API代码/本机资料、不声称全功能对齐。

| 能力/关键旅程 | 已实现（源码静态） | 局部测试 | 宿主完整测试 | 当前独立QA | 已发布 |
|---|---|---|---|---|---|
| 服务/渠道接入核验，J01 | Go两来源/表单/revision核验 | API/193Node | Pass H64 | 证据审查Pass；独立self_hosted UI Not Run、QA实操Blocked | Not Run |
| 模型发布/授权，J01/J02 | 公共目录/交集/价/启停 | 权限/摘要回归 | Pass H64 | 同版新UI/API证据审查Pass | Not Run |
| Key创建/轮换/撤销，J03 | 一次秘密/共池/期限 | Q65期限/回拨race Pass | Pass H64 | 原层HTTP时钟缺陷Closed，UI证据审查Pass | Not Run |
| 身份/成员/tenant隔离，J05 | 内部真实用户/邀请/角色/范围 | 当前ClockRecovery/queue Pass | Pass H64 | 越权/邀请原层及UI审查Pass；外部身份Not Run | Not Run |
| 额度/预算/计费，J04/J06 | 五维/零软/持久钟/保守预留/追加账本 | Q65零soft/守恒Pass | Pass H64 | 两P1 Closed；API/数据同版核对Pass，测试USD非供应商价 | Not Run |
| 日志/用量/诊断，J06/J07 | 同ID多尝试/筛选分页CSV/核对 | Q65 193Node Pass | Pass H64 | 原层J06/摘要事实证据审查Pass | Not Run |
| 表格/编辑/详情/移动 | A/三视口/统一表单 | 193Node机制Pass | Pass H64/三视口 | 当前三图审查+30页结构化核对；其余图留存/P3，QA实操Blocked | Not Run |

未来差距：QA58-01 P3改善；独立self_hosted首次UI分支补测；真实供应商/外部身份支付需实际凭据后再验；生产性能/SLO、税务退款、多协议、GPU部署、多区HA不在本冻结版本。当前候选Go with known issues可交report，完整产品/平台/正式页面退出未完成。
