# 逐AC分层验收矩阵02

2026-10-06，全量整改复验，当前11快照。预期来自 [PRD](../product-definition.md)、[验证计划](../routing-validation.md)及 [原整改](remediation-01.md)，不是从实现反推。**每行结果只涵盖所述层与步骤。** H=[宿主seq21同版本完整固定入口](evidence/host-seq21.md)；Q=[QA独立契约exit0](evidence/contract-retest-02.log)；V=[QA独立前端VM119/119](evidence/frontend-retest-02.log)；L=[版本与辅助证据锁](evidence/version-lock-02.json)。原01矩阵保留为历史。

## 需求—旅程—接口／状态—测试—证据

| ID | 基线要求 | 角色/范围 | 起始状态 | 连续用户步骤 | 最终事实 | 证据类型 | 结果 | 证据位置 |
|---|---|---|---|---|---|---|---|---|
| AC01-UI | 两个独立上游，菜单创建/回读/清密 | 管理员，每屏同一会话 | 无本屏A/B，其他对象可共存 | 概览→上游→新建A→保存→新建B→保存 | 各POST201，凭据清空，候选准确选两身份 | UI-E2E | Pass | H；routing.mjs创建步骤；L中156行选择轨迹仅辅助 |
| AC01-EDIT | 取消、重名拒绝、变址无凭据拒绝与修正 | 管理员，A/B | 两上游已保存 | B草稿→取消→改重名→拒绝→变址无凭据→拒绝→还原/补凭据→保存再核验 | PATCH409/400/200，原名称取消恢复，生成映射仍可用 | UI-E2E / API响应对账 | Pass | H；routing.mjs upstream edit rejection recovery每屏 |
| AC02-UI | 非生成核验，合法才成功，确定失败/取消可恢复 | 管理员，每屏 | 本屏UI建A/B | 分别核验→非法/401/429→取消→重读→超时→离页→回来修正再核验 | 失败不伪成功、GET与POST分计数；取消持久canceled、外部停止；修正成功 | UI-E2E / External State | Pass | H；routing.mjs check阶段及真实counts/读取取消终态 |
| AC02-PROTOCOL | Content-Type/object/每项id/JSON严格契约 | 管理员，隔离安装 | 测试API配置上游 | handler check→合法空列表和各错误响应→回读 | API与SQLite状态一致；5非法格式拒绝，坏JSON失败 | Contract / Data Reconciliation | Pass | Q；原qa_connection_contract_test.go未改预期；H也复跑 |
| AC02-BUDGET | 10秒总预算，头/body都包含 | 管理员，隔离HTTP | 已配置 | 显式check→阻塞头或body→等终态 | 出站停止；failed/upstream_timeout，9～12秒真实等待；持久一致 | API-E2E / External State / Contract | Pass | H；TestConnectionRealHTTPHeaderAndBodyBudget；Q原预算探针；两者不混称 |
| AC02-FAILURE | 429/503/network/协议确定失败，不是unknown | 管理员，隔离安装 | 已配置 | 各受控响应→查询结果 | check.State=failed，API/持久一致；旧历史未知不改写 | Contract / Data Reconciliation | Pass | Q；TestQAConnectionCheckProtocolAndFailureStates全部原子例 |
| AC02-DEDUPE | 单上游去重与取消释放，不改路由 | 管理员，隔离安装 | 一次核验pending | 再check→409→取消第一条→重新核验 | 出站访问1次、context终止、enabled/revision/health不变，再入成功 | Contract / Data Reconciliation | Pass | Q；TestConnectionCancelDeduplicatesAndPreservesRouting；网页忙时disabled由H |
| AC02-ASYNC | 离页/会话过期/旧结果晚到隔离 | 管理员，网页owner与后端scope | 延迟核验 | UI离页取消、后端核验期间改配置；VM过期及迟到 | 旧结果不污染新页/新配置；当前检查不自动重发 | UI-E2E / Contract | Pass | H真实离页；V cancellation/navigation/expiry；H TestConnectionCheckDoesNotGenerateAndScopesLateResults，各证明自己的分支 |
| AC03-UI | 双候选不同远端名、权重错误纠正/取消/保存 | 管理员，每屏 | UI新建A/B | 模型创建双行→weight0→保持编辑→纠正取消→后续保存权重/优先级 | 公共别名稳定，实际供应商收到各远端名，编辑保存生效 | UI-E2E / External State | Pass | H；routing.mjs；main.go受控上游remote-model断言 |
| AC03-BOUNDARY | 整数端点/越界/分数/空远端/引用/重复整体拒绝 | 管理员，配置API | 隔离两候选 | 各非法PATCH→回读→合法端点 | 400零部分写，0/100 priority与1/100 weight合法；重复拒绝 | Contract / Data Reconciliation | Pass | Q TestCandidateIntegerAndReferenceBoundariesAreAtomic；H TestCandidateWriteRejectsDuplicatesAtomically |
| AC04-ROUTE | 1:3/1:1/大数优先实际选择 | 管理员，每屏→应用调用 | 两完整候选 | UI改规则→40/20/10次UI发送 | 30:10、10:10、0:10独立外部计数与记录；不是配置JSON证明 | UI-E2E / External State | Pass | H；routing.mjs精确计数 |
| AC04-DISABLE | 三层启停与备用开始门止损 | 管理员/应用，HTTP边界 | 旧配置有效 | 停候选UI调用；HTTP停上游/模型与fallback门前停用 | 停用确认后没有新尝试，不混入新映射；已开始允许完成 | UI-E2E / API-E2E | Pass | H；routing.mjs、TestRoutingRealHTTPDisableBeforeFallback、TestHTTPConfigurationChangesAndReenable；快照/门契约测试分层 |
| AC05-HTTP | 暂时错误白名单/其他不切换/最多三不同上游 | 应用，2～4候选 | API显式准备边界配置 | 请求→受控状态/网络/输出前超时→fallback或耗尽 | 同ID真实尝试、不可切换零B、第四零访问、稳定错误 | API-E2E / External State | Pass | H routing_http_test.go allowlist/four candidates/selection；routing_boundary_test.go；API仅边界准备 |
| AC05-TIME | 整体预算不重置、首提交停止尝试timer | 应用，缩短隔离超时 | 两候选 | 超时/取消/首提交→继续或终止 | 整体到期不新尝试；precommit可一次切换；提交后不被尝试timer截断 | Contract / API-E2E | Pass | H TestRoutingDeadlineAndCancellationNeverRestart、TestRoutingPrecommitTimeoutCanSwitchOnce、TestAttemptOutputStopsTimerOnFirstCommit；不是300秒SLA实跑 |
| AC06-UI | 冷却跳过与30秒到期新请求恢复 | 管理员，每屏 | A503、B正常 | 故障→再调用→模型冷却状态→外部恢复→31秒后新调用 | 冷却内零新A，到期新请求才A；手动禁用与冷却不混称 | UI-E2E / External State | Pass | H；routing.mjs passive recovery；受控上游仅恢复其响应，不操作网关健康 |
| AC06-EPOCH | 全冷却零访问、启停保留、旧结果/重启隔离 | 管理员/应用，隔离状态机 | 同代次故障 | 较旧成功晚到→全冷却调用→启停→换凭据/映射→旧观察→重启 | 零出站，启停不清冷却，新执行不被旧健康污染，重启未知 | Contract / Data Reconciliation | Pass | Q/H TestRoutingHealthEpochSequenceDisableAndRestart；H短冷却专项；重启待办H实际UI |
| AC07-STREAM | 已提交SSE与用户取消不重放 | 管理员每屏/双候选 | A首增量已输出 | SSE断流→查看失败；阻塞流→UI取消 | B外部计数0，不伪造DONE；context取消；敏感输入清理 | UI-E2E / External State | Pass | H routing.mjs committed stream/cancel；原Cancellation/SSEBarrier测试 |
| AC07-TOOLS | 工具首增量后EOF/idle/cancel零备用 | 应用，真实双候选HTTP | 工具事件已提交 | 接收工具增量→各异常终止 | 不重复工具增量、不再B，取消传出站，不完整结束真实记录 | API-E2E / External State | Pass | H TestRoutingRealHTTPCommittedToolsStopBoundaries三项，不拿普通文本流替代工具组合 |
| AC08-UI | 同Request ID逐尝试/多次未知消耗 | 管理员，每屏 | A失败→B成功 | UI结果→查看本次记录 | 两真实上游/远端/顺序；完整合计未知、未知尝试1，不只显示末次用量 | UI-E2E / Data Reconciliation | Pass | H routing.mjs同ID精确详情；H TestRoutingTemporaryAllowlistAndUnknownTotals |
| AC08-STATES | started/unknown/404/503/历史无明细与用量聚合 | 管理员/记录API | 隔离存量及显式故障注入 | 精确查询/分页/恢复；Store输入测聚合 | 不替换成最新成功；未知不报0，历史明细不造；写故障fail-closed | UI-E2E / Contract / Data Reconciliation | Pass | H journey.mjs QA recovery、TestUsageSnapshots、TestStorageFailClosed及migration历史对账；注入只证明对应分支 |
| AC09-FLOW | 双屏完整连续路由键盘任务/焦点 | 单管理员，每屏独立会话 | 无本屏A/B/alias/key | J-ROUTE各屏连续创建/核验/编辑取消/签发/发送/停止/详情恢复 | 原生Tab/Enter/Space/type-ahead及focus-visible、真实规则/计数、无隐藏业务API步骤 | UI-E2E | Pass | H；routing.mjs；L中选择JSONL只辅助候选动作 |
| AC09-COPY | 每屏密钥复制成功、失败就近恢复 | 单管理员，每屏原首配会话 | UI刚签发密钥 | 原生按钮复制→注入拒绝→手动提示→原生复制重试 | 本机clipboard与当前签发值相等，只断言布尔；清空剪贴板/撤销测试授权 | UI-E2E / External State | Pass | H；journey.mjs copyIssuedKey，每屏；注入仅拒绝分支，成功使用真实clipboard |
| AC09-ROUTE-COPY | 在routing.mjs本屏路由对象上另外执行原生复制 | 单管理员，双候选会话 | routing-key已签发 | 此driver读取一次性UI值并按键输入，没有点复制 | 本单独组合未执行；同轮AC09-COPY验证完全相同控件/处理器、各屏同角色复制用户结果 | UI-E2E | Not Run | 门槛等价证据为AC09-COPY；不是API＋页面存在性拼接；J-ROUTE仍真实创建/用户按键发送 |
| AC09-OVERVIEW | 已完成回看/当前待办/账户分组/身份可查 | 单管理员，每屏 | 当前配置生成成功 | 回概览0/5默认关闭→原生展开→精确成功ID详情→返回收起→重启登录2/3 | 原节点/证据身份保持，未知不造健康，焦点可见、表格局部滚动 | UI-E2E / Static Review辅助 | Pass | H completed review各细分阶段；L中双屏新概览/重启截图；关闭QA-04 |
| AC10-RECOVERY | 写拒绝/部分成功只读恢复/过期清密 | 管理员，两屏 | 存量任务/故障 | 原journey键盘纠错/重读/会话失效→安全重入 | 不自动重写或重生成，非敏感草稿可修，秘密/旧输出清理 | UI-E2E / Contract | Pass | H原journey完整复跑＋新核验恢复；V异步迟到/过期，各层不替代 |
| AC11-MIGRATION | 精确旧main非空UI升级/原key调用/旧备份回退 | 旧/新单管理员，隔离安装 | 真85f6d17 schema1空安装 | 旧UI建上游/2别名/2keys启停/成功失败调用→backup→upgrade→新UI/原key→旧restore/调用 | schema1→2；身份/历史/nullable usage/凭据守恒，新增不在旧backup | UI-E2E / CLI / Data Reconciliation | Pass | H；migration.go、legacy.mjs、migration-after.mjs；DEP01，不是Actions |
| AC11-REFUSAL | 错key/锁/未知版本/原子失败/重复 | 运维，隔离副本 | schema1/2 | CLI/Store各边界→回读 | 拒绝不迁移，完整性与重复运行有界 | Contract / CLI | Pass | H upgrade_test.go及deployment_test.go，非空主UI另列 |
| AC12-REGRESSION | 普通/SSE/工具/鉴权/SSRF/限额/首配/backup | 管理员＋应用，隔离服务 | init及受控协议 | 原HTTP/CLI/浏览器真实回归 | 参数透明、非法权限零出站、来源与凭据隔离、fail-closed、无默认假供应商 | API-E2E / UI-E2E / Contract | Pass | H固定入口原main/relay/boundary/security/health及journey资产 |
| AC13-GATE | 当前同版完整固定入口及QA契约 | 工程，11快照＋模板联合锁 | 85文件852f7bfc… | 原宿主make verify→QA身份/回归审查 | seq21 exit0；原QA资产保留；真实证据和历史失败分开 | Host Receipt / Contract | Pass | H/Q/V/L；不是08历史绿，不是QA重跑监听链 |
| QA-HOST-REPLAY | QA自己重复完整监听/Chromium链 | QA沙箱 | listen/Mach限制未改变 | 未重复已知受限链 | 无QA独立监听回执；同版H真实链覆盖本轮门槛 | Execution Environment | Blocked | 原make-verify-01.log/浏览器探针；本轮用既定宿主，不扩大权限 |
| DEP02-04 | 真controller旧/新Go联合顺序/失败/legacy | 目标维护者，隔离部署根 | 固定维护源/标准安装 | 真实可信联调尚未执行 | 不能从平台脚本夹具推断Go数据守恒/恢复 | Platform Integration | Not Run | 本轮未操作控制器；发布依赖不解除 |
| DEP05-INSTALL | 受管理模板标准安装/manifest身份 | 维护者/研发，受管理文件 | 原标准安装 | 授权安装器upgrade/repeat→QA只读比对 | Installed/Unchanged回执；当前模板/manifest摘要相同 | Installer Receipt / Static Review | Pass | delivery/verification/platform-template-provenance-01.json＋L；仅模板子范围 |
| DEP05-EDGES | 控制器全部安装拒绝/漂移/失败分支独立复跑 | 平台维护者 | 用户提供源测试/安装进展 | 产品QA未复跑维护源测试 | 不从模板一致扩为全部边缘成功 | Platform Integration | Not Run | 维护者45项回归为用户输入来源，不冒充本节点执行 |
| DEP06-07 | 正式Actions非空升级/同可信路径隔离失败恢复 | 目标维护者，授权后续正式路径 | 本轮尚未发布 | 合并后实际Actions/Deployment核对 | merge/schema/非空数据/回退需真实一致 | Production Journey | Not Run | 设计明确草稿PR前无需预发生；报告必须保留 |
| SUPPLIER | 真实供应商/费用/能力联调 | 用户授权管理员 | 无供应商凭据与费用授权 | 未执行 | 受控上游不等于真实供应商 | Supplier Integration | Not Run | 所有本轮实际外部行为为标注受控fixture |

## 核心用户旅程执行契约

同版H实际运行。单管理员；1280/390各自context/page，路由每屏建本屏新对象，390不只是看桌面结果。服务及受控上游只做Start State；准备/清理不代替UI Steps。密钥不写证据，只从UI签发；SQL用于迁移历史/身份对账，外部control API用于故障/计数；所有用户中间配置仍在产品页面完成。

| Journey ID | User Story/AC | Actor | Start State | UI Steps | Expected UI Result | Final Fact | Cleanup | Result |
|---|---|---|---|---|---|---|---|---|
| J-FIRST-1280 | 原首配、AC09/10/12 | 单管理员1280会话 | init无有效上游 | 登录→上游错误/纠正→模型→签key/复制成功及拒绝恢复→调用→同ID详情→过期重入 | 首配入口/安全纠错/复制可用/清密重入 | 真实Go/外部POST/同ID记录 | clipboard清空/权限撤销，context关闭，runner收尾 | Pass |
| J-FIRST-390 | 同上 | 单管理员390会话 | 新别名/key，已有共享上游 | 本屏实际编辑/签发复制/调用/故障恢复，不只截图 | 窄屏同结果/焦点 | 当前key/请求/记录可对账 | 同上 | Pass |
| J-ROUTE-1280 | AC01～10 | 单管理员1280同一会话 | 无本屏A/B/双候选alias/key | 菜单创建A/B→核验成功/失败/取消/超时/离页再入→上游编辑修正→双候选创建/错误/取消/编辑→签key→UI实际分配/故障→详情→回看/恢复/停止→重启登录 | 每个实际任务均原生键盘完成，无隐藏业务配置API | 真计数30:10/10:10/0:10，停止零备用，终态/同ID，完成0/5→重启2/3 | context/browser.close、HTTP shutdown、临时目录defer；保留安全截图轨迹 | Pass |
| J-ROUTE-390 | 同上 | 单管理员390独立会话 | 桌面对象仍在，创建本屏新对象 | 同样完整本屏连续任务 | 窄屏真实流程/焦点/无关键遮挡 | 本屏映射、计数、终态及同ID真实 | 同上 | Pass |
| J-MIGRATE | AC11、DEP01 | 精确旧/新管理员，隔离旧安装 | 85f6d17真实旧binary/schema1 | 旧UI形成上游/多个别名/启停keys/真实历史→正式backup/upgrade→新UI操作与原key调用→旧restore/旧服务调用 | 旧身份/历史继续可用，新增不在回退内 | 2别名/2keys/2历史/4受控POST、schema1→2及旧backup守恒 | 旧新进程stop，runner临时目录cleanup | Pass |

UI步骤的Pass不将所有API边缘分支加入其内部；API用例各自有层级。未执行的controller、Actions与供应商是后续验收，而不是在当前网页步骤中靠管理员后台补救才能完成。当前交付结论及缺陷闭环见 [report-02](report-02.md)。
