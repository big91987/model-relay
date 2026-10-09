# 当前候选分层验收矩阵

2026-10-09，qa执行10；HEAD3f9cee16、154文件指纹2dfe14a3（完整值见报告）。所有Pass只对当前同指纹tests9或QA本地证据；缩写 E=`acceptance/evidence/qa-10`（路径相对于本Run根），H=`E/host-tests9/verify.log`。Contract/API测试UI步骤列明确“不适用”，不能从它推出UI通过。原材料根和各文件摘要见E/material-integrity.json。

| ID | 基线要求 | 角色/范围 | 起始状态 | 连续用户步骤 | 最终事实 | 证据类型 | 结果 | 证据位置 |
|---|---|---|---|---|---|---|---|---|
| PLAN01 | 原设计/材料可核验 | QA/61文件 | prepare同包 | 读取manifest，逐项核验 | version、ZIP及61摘要一致 | Static Review | Pass | E/material-integrity.json |
| GATE | 固定完整make verify | 宿主/当前候选 | 冷依赖/隔离目录 | 执行全部原入口 | exit0、gate0、source/candidate unchanged | Contract | Pass | E/host-tests9/state.json、H |
| VERSION | 实际受测源码一致 | QA/产品+tests+locks+contracts | 未提交候选 | 重算原source入口 | 154/2dfe14a3与宿主前后一致 | Static Review | Pass | E/source-qa.json |
| SAFETY-API | AC10/16、PLAN02默认输出限制 | 管理员/tenant TPM、key TPM、硬预算 | 有界模型output64 | UI不适用；正式API配置→普通/SSE→诊断 | cap64默认/显式32、单attempt；6次RPM6/TPM42/fee42 | API-E2E | Pass | H TestMeteringSafetyActualHTTP/METERING_CAP |
| SAFETY-UI | 上界正式入口 | 平台管理员/受控模型 | 空安装 | J01网页接入→登记100/100依据→配置价格/授权→调用 | 表单PUT200，UI调用和同ID详情成功；不证明64或供应商安全 | UI-E2E | Pass | E/host-tests9/browser/continuous-*-*.json；tests/browser/journeys.mjs J01 |
| BOUND | 声明违反后拒绝新增有界请求 | 管理员/双scope | 受控声明100 | UI不适用；上游报output101→新请求 | 实际TPM104/fee104不截断；后503/零attempt | API-E2E | Pass | H TestMeteringBoundViolationActualHTTP |
| RPM-API | AC09/PLAN03非零阈值 | tenant及key分别 | 同一注入UTC分钟、RPM2 | UI不适用；2接受→第3拒绝→末1ns拒绝 | rejection双scope全值不变、零dispatch | API-E2E | Pass | H TestMeteringWindowsActualHTTP/{tenant,key}/rpm |
| RPM-WINDOW | AC11/PLAN05边界与显式恢复 | tenant及key | RPM2/Calls2/fee14 | UI不适用；精确跨分钟→新SSE | Calls3、RPM1、TPM7、fee21；历史不清零 | Contract | Pass | H METERING_SNAPSHOT rpm_recovered |
| TPM-API | AC09/10、PLAN04 U+P+B等值/少1 | tenant及key | U7/P0/B200 | UI不适用；改额206/207→406/407→399/400 | 拒绝无部分扣计；未知P200 carry；旧核对后当前TPM7/fee28 | API-E2E | Pass | H TestMeteringWindowsActualHTTP/tpm |
| TERMINAL-API | AC09/12/16、PLAN06模式计量 | 双scope | 单受控candidate | UI不适用；普通/SSE known/missing/inconsistent→重放 | known Calls/RPM1/TPM7/C0/fee7/reserve0；unknown TPM200/Pending1 | API-E2E | Pass | H TestMeteringTerminalActualHTTP |
| TERMINAL-DATA | AC12/13成本/费用/重复守恒 | request/tenant/key | 实际受控usage3+4 | UI不适用；只读同ID账 | fee7、cost33；重复409原ID/无第二dispatch | Data Reconciliation | Pass | H terminal METERING_SNAPSHOT及TestMeteringTerminalActualHTTP |
| FAILURE-API | AC10/13、PLAN07已dispatch失败 | 双scope | 受控HTTP500 | UI不适用；已知/未知失败 | 最终502；known仍fee7/cost33；unknown P200/Pending1、RPM1 | API-E2E | Pass | H failure_known/failure_unknown |
| UNDISPATCHED | 受理后未dispatch终态 | 内部执行层 | Calls/RPM1、C1、P200 | UI不适用；admit→cancel→routedGenerate/Finish→晚到 | Calls/RPM1保留、C0/TPM0/fee0/attempt0，晚到不改cancel | Contract | Pass | H及E/independent-go.log TestMeteringUndispatchedInProcess |
| CANCEL-HTTP | AC16首输出后取消、不重放 | 双scope | 活跃SSE C1/P200 | UI不适用；HTTP首行→客户端取消→等待上游context | 单attempt、cancel终态/C0、RPM保留、P200跨窗 | API-E2E | Pass | H TestMeteringCancellationActualHTTP |
| TIMEOUT-HTTP | AC10/11未知超时守恒 | 双scope | 受控持有连接 | UI不适用；2秒超时→读诊断 | 504 request_timeout、attempt1、RPM1、P200/Pending1 | API-E2E | Pass | H TestMeteringTimeoutActualHTTP |
| COMPETE-API | AC09/PLAN08并发原子性 | 2Key/轮换两代 | U7、只余1名额 | UI不适用；8屏障并发→保持上游→读7拒绝→释放 | 1受理7拒绝、occupied207/C1/fee7、final14/C0/fee14 | API-E2E | Pass | H TestMeteringCompetitionActualHTTP |
| POOLS-DATA | AC07/09、PLAN09共享/隔离 | tenantA K1/K2、tenantB、两代 | A并发已占余量 | UI不适用；比对双Key之和→B独立调用 | Key总值=tenantA；B Calls/RPM1/TPM7/fee7、A不被B扣计 | Data Reconciliation | Pass | H METERING_COMPETITION及相关快照 |
| DIMENSIONS | AC09独立次数/RPM/TPM/C/预算 | tenant/key各5维 | 隔离资源 | UI不适用；zero→positive→null、竞争、改额 | 零拒绝、等值预留、无部分扣计、改额不清历史 | API-E2E | Pass | H TestConsoleResourceMatrixActualHTTP/SharedResourcesActualHTTP |
| CONSERVE | AC11–13/PLAN10价格/核对/暂停守恒 | 多attempt/双scope | 有非零历史与未知项 | UI不适用；原治理与暂停用例 | 受理价快照/全部attempt/差额幂等；暂停不清计量 | API-E2E | Pass | H TestConsoleGovernanceActualHTTP/ResourceDimensionsActualHTTP/PauseUsageActualHTTP |
| RECOVERY | AC11真实进程与有效UTC | 临时安装/旧新凭据 | 持久已用、pending、高水位 | UI不适用；真实重启/备份restore/回拨 | 旧凭据不复活、用量/未知守恒 | Contract | Pass | H TestConsoleRealProcessCrashRecovery、ClockRecovery{InProcess,ActualHTTP}；父测试必须校验子进程marker |
| WORKER-OUTER | helper直接启动前置 | 全包test父进程 | 缺专有state env | 单独worker不执行 | 外层Skip，不是恢复Pass | Contract | Not Run | H TestConsoleClockRecoveryWorker；替代恢复事实仅引用RECOVERY |
| OBSERVE | AC17/PLAN11数值与隐私 | QA/白名单归档 | runner与本地日志 | 核对limits/usage/period/clock/ID/dispatch | 金额精确字符串，text无秘密/个人路径，转换留摘要 | Static Review | Pass | E/receipt-review.json、local-log-transforms.json |
| QA-RACE | 新计量无监听独立回归 | QA/8组InProcess | 自有临时夹具 | 运行race/parallel4 | exit0、29.082秒；不是HTTP/UI | Contract | Pass | E/independent-go.log |
| QA-NODE | 新冷却读边界/诊断资产 | QA/3断言 | 无后台服务 | 两test文件运行 | 3pass/0fail/0skip | Contract | Pass | E/independent-node-focused.log |
| UI-BASE | 原J01–07连续回归 | 四角色各独立会话/1280/1440/390 | 空安装及创建后对象 | 见独立journey表，页面写入/调用/恢复 | 七旅程各三宽完成，same ID/额度/价快照/账单最终事实 | UI-E2E | Pass | E/host-tests9/browser/continuous-*.json，H |
| UI-FAULT | AC08/16重复/未知恢复 | 已签Key/三个宽度 | 故障注入与页面 | 丢响应→503/404→读回；双击ordinary/SSE | secret不重放、一次POST/admission/attempt/dispatch，usage3+4 | UI-E2E | Pass | E/host-tests9/browser/fault-{1280,1440,390}.json |
| ROUTING | 原冷却/自然恢复、不重放 | 控制服务/1280/390 | 2候选 | 页面调用→故障→冷却→自然30秒恢复/SSE取消 | status200/target1/cooling1/expired0、DOM冷却true；非RPM窗口 | UI-E2E | Pass | H routing_cooldown与routing阶段；E/host-tests9/browser/routing-*.png |
| PLAN-UI-NONZERO | #33原两屏RPM/TPM、自然跨窗/双浏览器数值链 | DA两会话/1280/390 | 正式合法模型/预算 | 尚未执行；按原PLAN03–10 | 不用API数值或旧zero-calls拒绝代替 | UI-E2E | Not Run | 原材料test-report/user-journeys.md及本Runtest-plan C；前置见SUPPLIER |
| SUPPLIER-SAFETY | DS-SAFETY/PLAN02与12 | DeepSeek精确模型 | 凭据/收费上界/思考策略/历史账未证明 | 禁止新生成 | 无法证明真实输出/隐藏推理/剩余预算 | External State | Blocked | findings-and-next-actions-qa-10.md；原DS-SAFETY候选 |
| SUPPLIER-GENERATE | 原普通/SSE成功、恢复 | DA/1280/390 | 同上 | 未执行目标模型新生成 | 没有supplier成功ID/usage | External State | Not Run | 原report No-Go保留；qa新增真实生成0 |
| SUPPLIER-METER | PLAN03–10真实Token/费用/自然窗口 | provider/request/tenant/key | 同上 | 未执行原数值旅程 | 内部7/33/200不作为supplier事实 | Data Reconciliation | Blocked | test-plan-review-qa-10.md、原PLAN整改表 |
| RELEASE | merge/Actions/生产部署同版 | 发布Connector | 尚未PR/publish | 本节点不执行 | 未提供merge/Deployment/线上版本 | External State | Not Run | 后续report/publish责任，不用门禁替代 |

未创建产品缺陷的缺失用例不默认为合格：PLAN-UI-NONZERO及供应商行明确留空执行结果，并交专用测试Pipeline。阶段Go只对M01内部资产，不把上述Not Run/Blocked整体合并成产品Go。
