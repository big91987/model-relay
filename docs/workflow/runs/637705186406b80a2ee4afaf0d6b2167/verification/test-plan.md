# #34 可执行复核计划 r1

状态：研发自审完成，待tests独立文档评审及真实网页复测；不是已批准的独立评审。权威来源和材料摘要见本Run README。冻结原设计不修改；原报告/截图/首失败保留。每行区分内部契约、真实HTTP、浏览器和供应商。

## A. 受控矩阵：接入原 make verify

共享配置：隔离临时Go/SQLite安装；正式管理API创建对象（不是SQL预填）；单受控候选；默认输入/输出上界100+100，B=200；真实用量声明3+4=7；租户费率各1,000,000微美元/百万tokens，候选成本费率3,000,000/6,000,000。普通/SSE都走产品代理，fake只在测试内，未作为产品默认。

| PLAN / 原AC | 可执行用例 | 准确断言与副作用证据 |
|---|---|---|
| 01 / 材料 | README材料清单，manifest重新核验61项 | 原PRD/交互/AC09–13/16可访问、摘要一致；旧Run缺原文不追写成当时通过 |
| 02 / AC10/16 | TestMeteringSafety{InProcess,ActualHTTP} | 正式bounds输出64；tenant TPM、Key TPM及硬预算分别触发；普通/SSE无参数→上游max_tokens64；显式32原样，max_completion_tokens32不混入另一cap；同ID单attempt；6次每scope Calls/RPM6、TPM42、Fee42 |
| 02 / AC10 | TestMeteringBoundViolation* | 控制响应输出101违反100；记录104真实声明Tokens而不是裁掉；下一请求503 trusted_bound_required、零dispatch/零部分扣计。只证明内部失信保护，不证明真实供应商安全 |
| 03 / AC09 | TestMeteringWindows*/{tenant,key}/rpm | RPM2：第1/2请求接受、第3拒绝；所有维度和两scope保持；下一分钟前1ns仍拒绝；精确边界后RPM/TPM0、Calls2/Fee14保留，显式新SSE→RPM1/TPM7/Calls3/Fee21 |
| 04 / AC09/10 | TestMeteringWindows*/{tenant,key}/tpm | U7、P0、B200，L206拒绝/L207接受未知SSE；后U7/P200，L406拒绝/L407接受已知；U14/P200，跨窗只转出U14；L399拒绝/L400接受；对旧分钟unknown追加核对后本分钟TPM仍7，Fee28，重复核对不二次释放 |
| 05 / AC11 | 同上，已有ConsoleReviewClockRollback*、ClockRecovery* | 固定有效UTC、旧known转出、pending跨窗；注入时钟边界是契约证据，不是自然时间网页证据。实际进程/备份恢复由原make verify宿主执行 |
| 06 / AC09/12/16 | TestMeteringTerminal* | 普通/SSE分别known7、missing usage、inconsistent total；known Calls/RPM1、TPM7、C0、Fee7、Reserve0、Pending0；unknown Calls/RPM1、TPM200、C0、Fee0、Reserve200、Pending1；成本33与租户费7独立；重复operation409返回原ID、无第二dispatch |
| 07 / AC10/11/13 | TestMeteringTerminal*/failure_* | 实际受控500 known也收费7/成本33；unknown留P200/Pending1，RPM不返还；单候选最后502且同ID单attempt |
| 07 / AC10/11 | TestMeteringUndispatchedInProcess | 直接生产admit→已受理后取消→routedGenerate/Finish：从Calls/RPM1、C1、P200到Calls/RPM1、C0、TPM0/Fee0/Pending0、零attempt；晚到success不改cancel终态。此行是内部执行层契约，不冒充完整HTTP |
| 07 / AC10/16 | TestMeteringCancellation*、TestMeteringTimeout* | HTTP取消必须下游读取首SSE行后cancel并看到受控上游context结束；C从1到0，Calls/RPM保留1、P200/Pending1跨窗；timeout2秒504 request_timeout，单attempt保留未知。InProcess cancel只观察上游prefix，不宣称下游输出 |
| 08 / AC09 | TestMeteringCompetition* | RPM仅余1或TPM207（U7+B200）8竞争者屏障同时开始；上游保持阻塞直到7个429、各ID零attempt已读回；只1个接受，占用Calls/RPM2、TPM207、C1、Fee7、Reserve200；释放后Calls/RPM2、TPM14、C0、Fee14 |
| 09 / AC07/09 | 同上multi_key / old_new_generation + 既有ResourceMatrix* | tenant双Key共享总池，两个Key快照求和等于tenant；轮换前后非零已用保留，新旧值同应用池；另一个tenant能接受且不改变原池；另tenant最终Calls/RPM1、TPM7/Fee7。独立Key各限额仍由原资源矩阵覆盖，不删旧断言 |
| 10 / AC11/12/13 | Windows/Competition、既有ConsoleGovernance*、ConsolePauseUsage*、RealProcessCrashRecovery、ClockRecovery* | 改额、轮换、暂停不清历史；价格快照、所有attempt账、核对差额幂等；真实杀进程/重启仍为独立原门禁宿主项。不能把Open/handler当真实进程 |
| 11 / 全部 | METERING_SNAPSHOT/CAP/REJECT/COMPETITION/CANCEL/TIMEOUT日志 | 每phase白名单limits/usage/period/effectiveUTC/window来源/上游hit，拒绝两scope不变；同IDattempt；金额字符串保精度。GET没有minute字段，窗口明确来自Store.now seam。日志不含消息、响应、秘密；保留失败和版本摘要 |
| 12 / 安全 | 下节专用真实测试前置与预算图 | 本研发新增真实生成0、外部供应商请求0；受控请求不占真实授权，也不填补真实用量事实 |

## B. 宿主执行与归因

执行8更新：[host-return-exec8.md](host-return-exec8.md)保留tests7完整失败：原两屏冷却/routing通过，累计十二分钟在full-console1440px终止。新18个独占计量相关顶层测试只改变有界调度，A节产品断言/数值及C节真实供应商前置不变，独立计划评审仍未完成；必须新候选完整宿主复验。以下执行5/6说明保留为历史。

唯一完整门禁 `make verify`；不绕过review-history、health-history、npm clean install、race、进程、真实Go服务浏览器阶段。新Go测试被原test阶段自动发现。宿主保留state.json/verify.log与candidate工作树摘要，结果非零即返回研发；最终独立QA核对同版及本计划后才判验收。

原生沙箱首次make verify失败于历史HTTP监听，外层make2/内部gate2。当前稳定同环境不重复必然失败的完整门禁。执行5旧候选HTTP/进程/计量race已通过，完整门禁在390px冷却展示处被十二分钟deadline终止，浏览器仅部分完成。执行6新候选完整门禁仍Not Run，静态/handler是补充。若宿主失败先区分产品、驱动、时序和环境，采最小诊断再改，不能删除原断言。新增读边界诊断和fixture调度不改变本计划产品断言；详见host-return-exec6.md。

## C. 真实网页与DeepSeek复测（专用测试Pipeline责任）

前置仍Blocked：专用配置入口取得凭据；精确 deepseek-v4-flash / deepseek-v4-pro 的当时可用性；正式模型bounds背后的输入及收费输出/隐藏推理真实依据；max_tokens语义和思考关闭（不支持就明确不能完成）；所有attempt历史预算权威账。公开包不含凭据。产品拒绝thinking字段，不能承诺供应商策略已关闭。不得重命名模型、调整原授权或把原24计数清零。

保留原桌面1280/390、DA独立会话、网页写入（API仅只读采证）及J05/J06/J09/J10/J11/J12、T09原范围。按A节数值关系从真实B计算限额，不能把100+100测试cap移植成供应商声明，不能把64输出当收费总Token64。自然跨一个固定服务端UTC分钟需受理时间/同ID/服务器窗口来源，不能仅客户端睡60秒。

以下为依赖图，不是已分配调用预算：普通成功A→RPM N与Token/费用对账；同窗拒绝R→经证零attempt才预算0；SSE成功B→SSE/RPM/known usage；自然窗口新请求C→恢复；TPM精确边界D→U+P+B；双会话并发E→只合法名额（但预算先预留两请求的最坏dispatch数）；跨Key/tenant F→共享/隔离；轮换后G→守恒及成功恢复。A/B等可以共享同ID事实，但不能冒充另一屏或另一角色。

预算必须按 `历史已证H + 历史未知保守占位X + 每个潜在supplier attempt预留 <= 24` 原子登记，而非按成功逻辑请求计数；每次<=64、请求<=90秒。独立评审先绑定每个新生成行、可能的最多3候选尝试、共享证据和剩余额度。超时/取消/可能dispatch不自动返还。H/X当前未知，所以可执行新生成预算为未获证明，不是24。若全部两屏/双模型/并发旅程不能在剩余预算内完成，选择有证依据的子集并列未验项，不扩大额度或宣称全验收。

供应商missing-usage、失败、取消、未知carry不能故意无界造付费失败；无法一次有界观察的分支保持Not Run/Blocked，内部用例分栏。真实usage和公开服务价格核验、供应商收费事实、平台测试定价分别记录。原No-Go只在专用Pipeline实际证据充分后改变。

## 计划自审 / 待独立评审

自审已按delivery-review-checklist：原AC可追踪，单任务事实源、原入口不弱化、同层首红仅限真故障、成本/费用与unknown分列、环境失败独立、恢复不依赖私有会话。独立评审未执行：tests负责确认断言源与预算，网页/供应商缺前置保持Blocked；任何实际业务规则冲突再返回requirements/design。
