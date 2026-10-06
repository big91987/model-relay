# 独立QA第二轮：18AC分层验收矩阵

基线 [PRD](../product-definition.md)、[架构](../console-architecture.md)、[交互](../console-interaction.md)、[整改包](remediation-01.md)、[连续旅程](journeys-02.md)。本轮全量复验，不沿用首轮绿色。HEAD d62cd51，56文件摘要 `1b2f7026692d0633f6f050d585cec0bb67571a7ecf411848d8213d2fd609cdc6`。

证据别名：H=[宿主seq33原样make verify，退出0](evidence/round-02/host-seq33.md)，B=该次执行的`tests/browser/journey.mjs`/`main.go`；V=[QA独立108 VM/helper](evidence/round-02/frontend-final.log)，G=[QA独立3 Go race契约](evidence/round-02/console-contract.log)，Q=[QA独立5明确证据动作](evidence/round-02/explicit-evidence-final.log)，O=[旧4条边界重跑](evidence/round-02/original-boundaries.log)，A=[同execution归档核对](evidence/round-02/archive-checks.log)。UI-E2E Pass取H的实际浏览器结果，代码仅用于核对该结果覆盖，V/G不补UI中间步骤。

| ID | 基线要求 | 角色/范围 | 起始状态 | 连续用户步骤 | 最终事实 | 证据类型 | 结果 | 证据位置 |
|---|---|---|---|---|---|---|---|---|
| AC01-UI | 六页/当前页/深链刷新/前后恢复 | 两屏管理员 | 正式入口/未知hash/已登录 | 六页导航→reload→history→重登录 | 当前页/hash/标题一致，仅安全页面标识入URL | UI-E2E | Pass | H/B navigate/history及最终六页布局 |
| AC02-UI | 中文过期恢复/密码显隐/应用401不登出 | 两屏管理员 | 密码错误、服务重启和应用拒绝 | 错密→重新输入；重启401→登录原页；应用401→记录 | 密码清密、恢复安全页、无自动生成、管理会话仍在 | UI-E2E | Pass | H/B password/rejectedLogin/restart/dependencyRecovery |
| AC02-C | 已知期限不依赖悬挂GET session | 前端session owner | deadline已知/检查pending/跨页新登录 | 触发期限→迟到成功/失败/401 | 清密停止调用，不延寿不复活不误伤新会话 | Contract | Pass | V deadline/watchdog/恢复/正常周期各分支 |
| AC03-UI | 1280/390布局/完整任务键盘/焦点与AX | 两屏管理员 | 同角色会话中空/存量对象 | Tab输入/创建取消/编辑启停/复制关闭/发送取消/诊断/定位分页 | 原生Tab可达、focus-visible/outline；无阻断横滚，表格自身滚动 | UI-E2E | Pass | H/B keyboardReach/keyboardAction及J03；A当屏模型精确选择 |
| AC03-C | 模型选择必须提交目标非首项而非默认值 | Harness，不是OS机制测试 | 启用目标/末项/缺失/无效果 | 执行真实helper序列的double | 目标index/值不匹配时仍Fail，诊断脱敏有界 | Contract | Pass | V helper/容量契约；不替代AC03-UI |
| AC04-UI | healthz原值/5秒隔离/真实计数/失败未知 | 两屏管理员 | 空/存量/读取故障 | 概览→版本刷新/失败恢复→503状态→重读 | dev原值，版本未知不阻塞，数量总数/启用分开，不造0健康 | UI-E2E | Pass | H/B runtimeVersionJourney/overview响应对照；两屏清密截图 |
| AC04-G | SQL全量计数，不以列表长度统计 | 管理API/隔离SQLite | 1006对象/真实0/null/关闭存储 | 管理查询与最终事实核对 | 全量总数启用数、权限拒绝和未知保守语义 | Contract / Data Reconciliation | Pass | G TestConsoleOverviewAndExactRecords；H当前Go包 |
| AC05-UI | 保存非核验；同成功ID/time/revision/alias/scope | 两屏管理员 | 空/保存/成功/配置变化/刷新重登录 | 步骤入口→生成→概览→核对显示证据→改配置/失败读 | 真实证据逐字段与精确记录一致；失效/失败清旧time/revision/动作 | UI-E2E | Pass | H/B evidenceSummary/evidenceOverview/recordsBoundaryJourney |
| AC05-G | 配置变更/旧调用/重启保守失效 | Go配置代次 | 当前success/调用pending/重启 | 变配置→迟到终态→重开服务 | 旧成功不污染当前配置，历史仍保留 | Contract / Data Reconciliation | Pass | G TestConsoleEvidenceConfigurationRace |
| AC06-UI | 真实Go空配置首配全UI | 1280px管理员 | 无上游/模型/密钥 | 上游→创建别名→签发→普通/SSE→记录 | 页面对真实Go写入和受控上游核验，无API/SQL业务预填 | UI-E2E | Pass | H/B J01；390承接存量，不伪称第二次空安装 |
| AC07-UI | 创建/取消/编辑启停与原依赖恢复 | 两屏管理员 | 空列表/合法存量/公开草稿 | 创建取消/冲突修正/编辑保存取消/依赖停启；同址空凭据/变址400→重读 | 取消零写，原对象恢复无绕行新对象；公开草稿保留，秘密不回显 | UI-E2E | Pass | H/B upstreamFailure/dependencyRecovery/keyboardStateChange |
| AC07-C | 上游恢复共享忙态、GET-only与迟到隔离 | 前端owner | token草稿/GET或PUT挂起 | 恢复→重复提交→跨页→旧finally | 立即清密、五控件锁、只GET；旧操作不能写新页/解锁新操作 | Contract | Pass | V closure恢复/共享忙态/相反操作 |
| AC08-UI | 签发一次、显式复制/关闭、存储日志截图保密 | 两屏管理员/应用权限 | 新签发受控值 | 复制拒绝手动恢复→隔离权限复制→内容布尔核对→关闭/跳页/reload | 不从列表回取原值；URL/存储/日志/截图无秘密 | UI-E2E | Pass | H/B copyIssuedKey/audit-logs/storage/清密截图；注入拒绝不称OS拒绝 |
| AC08-C | 复制/签发跨页退出迟到隔离 | 前端issued owner | 复制/签发pending | 关闭/跳页/退出→迟到 | 不重现秘密/旧提示，不误伤新秘密 | Contract | Pass | V clipboard/signing epoch |
| AC09-UI | 网关Base URL/占位示例/复制恢复 | 两屏管理员 | 已登录测试页 | 展开示例→原生复制→反馈 | 当前origin/v1、占位key/model，无供应商地址/真实key/生成POST | UI-E2E | Pass | H/B client-example/零额外请求断言 |
| AC09-C | 自动复制失败可手动操作 | clipboard夹具 | API拒绝 | 点击示例复制→拒绝 | 中文手动选择，不触发调用 | Contract | Pass | V clipboard failure |
| AC10-UI | 启用模型、可见输入snapshot、文本/工具/SSE及诊断 | 两屏管理员 | 无依赖/合法配置/处理中 | 依赖入口→显式普通/SSE/工具→诊断 | 正常完成/首段在结束前可见；输入四控件锁定保值，终态解锁；不执行工具 | UI-E2E | Pass | H/B submittedValuesVisible/callControlState/tools |
| AC11-UI | 防重复/真实取消/无DONE不完整/不重放 | 两屏管理员 | held生成/增量流 | 重复提交→取消→核对记录；断流/非法JSON→新请求 | 单次POST，受控上游取消；不完整非完成，已有输出不重试 | UI-E2E | Pass | H/B cancellation/recordsBoundaryJourney；测试重复requestSubmit显式注入 |
| AC11-C | 分块UTF8/CRLF/流error/旧finally隔离 | 前端stream owner | 正常/异常流 | 读取→取消或错误→晚到 | 当前状态正确、旧输出/解锁不串新请求 | Contract | Pass | V SSE分块、异常及owner组合 |
| AC12-UI | 失败清key/message保安全诊断与ID，离页退出清输出 | 两屏管理员 | 拒绝/协议/网络/断流/会话不明 | 失败→清密修正→新请求；离页/退出→恢复 | 应用401不登出管理、未取得ID不伪造；恢复无自动生成 | UI-E2E | Pass | H/B J04，O QA-04当前绿；QA-D01关闭 |
| AC12-C | 忽略abort任务/跨页/新登录隔离 | 前端所有权 | 旧JSON/SSE/401 pending | cancel/跳页/新登录→释放旧任务 | 清秘密、旧catch/finally不清新值不解锁新请求 | Contract | Pass | V owner组合，Q精确查询迟到 |
| AC13-UI | 完成/失败同ID核对字段、旧ID分页与并发 | 两屏管理员 | success/拒绝/未知用量/旧ID | 本次记录→精确核对；泛列表→下一页→定位；概览明确动作→刷新重登录 | 同ID/alias/model/keyID/state/延迟、0/null区分；不借最新行 | UI-E2E / Data Reconciliation | Pass | H/B sameRecord/recordsBoundaryJourney，52次UI拒绝且零上游 |
| AC13-C | 明确概览动作携带ID，泛导航不隐式过滤 | 前端/恢复后本地ID空 | overview持久success | overview-record→GET列表→精确GET | 显示ID与GET路径一致，卡片可见；另一泛导航无精查 | Contract | Pass | Q 5项；QA-D03关闭 |
| AC13-OLD | 原QA-03驱动普通records导航却预期精查 | 前端旧资产 | overview成功 | app.navigate(records) | 旧断言预期自动精查与整改包泛导航要求冲突 | Contract | Fail | O 3/4退出1；保留旧失败，不作为现产品缺陷；由AC13-C/UI同用户结果证据替代门槛 |
| AC14-UI | started/404/503/未知归属准确、手动恢复 | 两屏管理员 | held流/不存在ID/管理503/未知key | 页面定位→待终态；404/503→清详情→手动重试 | 不造行不借其他记录、不捏造未知归属；零上游触达 | UI-E2E | Pass | H/B recordsBoundaryJourney；503为管理响应注入 |
| AC14-G | 拒绝安全记账/权限/存储fail-closed | 管理及应用API | 隔离拒绝/数据库故障 | HTTP/精确查询→对账 | 统一401、安全元数据/未知归属/无对话日志、无法存储不伪成功 | API-E2E / Data Reconciliation | Pass | H当前Go HTTP/race；G DeniedMetadataAndStorage补强 |
| AC15-UI | 地址/凭据/停用/供应商/协议/网络中文反馈 | 两屏管理员 | 不合法/拒绝/依赖停用 | 页面触发→说明→对应入口修正→新请求 | 不仅英文码，不回显敏感供应商错误；应用拒绝零上游 | UI-E2E | Pass | H/B upstreamFailure/dependencyRecovery/recordsBoundary/sentinel审计 |
| AC15-G | 既有错误结构/来源/超时/存储协议兼容 | Go API | 隔离HTTP夹具 | 固定make verify当前测试 | 保持协议/认证/大小/网络/fail-closed | API-E2E | Pass | H当前Go包113.488s，原Makefile/锁文件未改 |
| AC16-UI | 显式原依赖恢复/取消零写/不自动重试 | 两屏管理员 | 上游/模型/密钥停用、失败 | 同对象再启用→显式新请求→记录 | 新请求独立、原对象恢复，不以新对象绕行、不自动付费生成 | UI-E2E | Pass | H/B dependencyRecovery/keyboardSave/生成计数 |
| AC16-D02 | PUT成功GET失败区分且只重读 | 两屏管理员 | 真实PUT/注入GET503 | 保存→已保存但未确认→恢复 | 一次写、清密保公开草稿、手动GET恢复 | UI-E2E / Contract | Pass | H/B upstreamFailure，O QA-02绿；QA-D02关闭 |
| AC17-S | 单一公共路线、范围授权、Trellis与等价设计 | 维护者 | 项目/Run文档 | 对照公共授权/契约/本轮路径 | Go/单实例/唯一上游；路线2–4独立接续；不复制实现不抢跑发布 | Static Review | Pass | README/product-brief/roadmap/releases/local-deployment/.trellis/spec；阶段进度以当前验收索引为准 |
| AC18-H | 固定入口/真实快照/失败保留 | 维护者 | 当前56文件 | 宿主原样make verify→QA重算/资产核对 | seq33退出0且同指纹；历史未覆盖，无产品/门禁/平台修补 | Contract | Pass | H/A/baseline.log/最终复核 |
| AC18-Q | 独立验收/公共任务里程碑闭环 | 维护者 | 首轮No-Go后当前候选 | 全18AC+原层整改+状态更新 | 3主题Closed、QA-G01已补，新Go仅本切片 | Static Review / Contract | Pass | report-02/本矩阵/公共delivery状态 |
| LOCAL-UI | QA亲自在沙箱运行全make verify/真实Chromium | 独立QA | 已确认监听限制 | 不重复扩权运行 | 不以静态browser.check或VM替代真Go | UI-E2E | Not Run | 同轮H同快照同用户结果作为门槛证据；本行仍未运行 |
| SUPPLIER | 真实供应商联调 | 管理员/供应商 | 无真实授权凭据 | 正式UI接入和实际响应 | 真供应商效果/费用 | External State | Not Run | 受控上游不是供应商证据 |
| DEPLOY | 新PR/合并/正式测试部署/部署后浏览器 | 获授权目标维护者 | 本轮无新发布回执 | 后续正式节点 | 合并SHA/制品/healthz/浏览器一致 | External State / UI-E2E | Not Run | 本切片验收不预报后续发布 |

矩阵结论 **Go（本轮第一条切片）**。旧驱动Fail、QA本地UI Not Run、真实供应商/新部署Not Run均保留，不按总通过率放行。不存在未关闭P0/P1/P2；同轮H已覆盖强制真实Go两屏/异常/恢复/分页/完整任务键盘。
