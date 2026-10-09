# 当前分层验收矩阵

统一候选：HEAD8917f942＋main/MERGE_HEAD6929376，158文件6e359667d5f7521f6b8c421454c6fc7cbe5ccb8c74aa87bb1e5955c194c202cc。材料根、版本和逐文件摘要见material-integrity；原AC按design-v0.2.0/docs/acceptance.md。以下E=`acceptance/evidence/qa-23/host-tests22/`（相对本Run），Q=`acceptance/evidence/qa-23/`。Pass的运行主体在证据栏明确，不合并未验层。

| ID | 基线要求 | 角色/范围 | 起始状态 | 连续步骤/可执行用例 | 最终事实 | 证据类型 | 结果 | 证据位置 |
|---|---|---|---|---|---|---|---|---|
| IDENTITY | 本轮精确联合候选 | QA | 未提交整合 | 复算158文件/3255main路径/231branch路径 | 无冲突、main完整保留、非干净状态诚实披露 | Static Review | Pass | Q/integration-recomputed.json、source-before.json |
| MATERIAL / PLAN01 | 同材料原文 | QA | prepare18包 | ZIP与61项sha重算 | version一致、零mismatch | Static Review | Pass | Q/material-integrity.json |
| GATE | 固定完整冷入口 | tests22 | node_modules/cache不存在 | make verify→原verify-gates全部阶段 | gate/Connector0、源码前后不变 | Contract / API-E2E / UI-E2E | Pass | E/state.json、完整verify.log；各层另列 |
| PLAN02-API | AC10/16正式安全默认 | admin、受控模型 | bounds64 | SafetyActualHTTP；普通/SSE默认/显式cap | 默认64、显式32；三trigger；6次Calls/RPM6、TPM42、Fee42；invalid零dispatch | API-E2E | Pass | E/verify.log METERING_CAP/REJECT/SNAPSHOT |
| PLAN02-BOUND | 违约不裁真实用量 | admin、受控响应 | output cap100 | BoundViolationActualHTTP | 104 Tokens记录、下一次503 trusted_bound_required、零dispatch | API-E2E / Data Reconciliation | Pass | E/verify.log TestMeteringBoundViolationActualHTTP |
| PLAN03 | AC09非零RPM和恢复 | tenant/key分别 | 第1已用 | WindowsActualHTTP rpm | 第2接受/第3拒绝；窗前1ns拒绝；窗后RPM1/TPM7/Calls3/Fee21 | API-E2E / Data Reconciliation | Pass | E/verify.log WindowsActualHTTP/{tenant,key}/rpm |
| PLAN04 | AC09/10非零TPM/unknown | tenant/key分别 | U7、B200 | WindowsActualHTTP tpm | 206拒绝207接受；406拒绝407接受；跨窗P200；399拒绝400接受；重复旧窗核对本分钟TPM7/Fee28 | API-E2E / Data Reconciliation | Pass | E/verify.log WindowsActualHTTP/{tenant,key}/tpm |
| PLAN05/10-PROCESS | AC11时钟/重启守恒 | 隔离进程 | 已用/未知/存量 | ClockRecovery/RealProcessCrashRecovery等原入口 | 回拨不增加额度、进程恢复/迁移保留账和历史 | API-E2E / Contract | Pass | E/verify.log相应PASS、migration浏览器阶段 |
| PLAN06/07 | AC12/13/16正常/失败计量 | 普通/SSE受控 | B200 | TerminalActualHTTP | known7、租户Fee7/成本33；unknownTPM200/Pending1/Reserve200/Fee0；失败已知也收费；重复operation409单attempt | API-E2E / Data Reconciliation | Pass | E/verify.log TerminalActualHTTP |
| PLAN07-CANCEL | AC10/11/16输出后取消/超时 | HTTP/SSE | 首行已读 | CancellationActualHTTP、TimeoutActualHTTP | 上游context结束、C释放一次、Calls/RPM1与P200保留；timeout504单attempt | API-E2E | Pass | E/verify.log CANCEL/TIMEOUT |
| PLAN07-NODISPATCH | 受理后取消无dispatch | 内部执行层 | 已admit/C1/P200 | UndispatchedInProcess | Calls/RPM1保留、C0/TPM0/Fee0/Pending0、零attempt、晚到不改cancel | Contract | Pass | E/verify.log；Q/independent-go.log |
| PLAN08/09 | AC07/09原子性/池 | 多Key/轮换代际/另一tenant | RPM仅余1或TPM207 | CompetitionActualHTTP：8屏障竞争、受理保持阻塞至拒绝诊断完成 | 1受理7拒绝；受理时TPM207/C1→释放TPM14/C0/Fee14；另一tenant7/7独立 | API-E2E / Data Reconciliation | Pass | E/verify.log METERING_COMPETITION/SNAPSHOT |
| PLAN10 | AC09/11/12/13改额/独立维度 | admin/tenant/key | 非零已用 | ResourceMatrix、Governance、PauseUsage等 | 次数/RPM/TPM/C/费用独立；改额/暂停/轮换不清账；价格快照/核对差额幂等 | API-E2E / Data Reconciliation | Pass | E/verify.log原对应测试 |
| PLAN11 | 白名单数值采证 | tests22/QA | 隔离数据 | 检查323条METERING日志及原断言 | 两scope/hits/attempt/窗口来源/精确金额可复核 | Data Reconciliation | Pass | E/verify.log；Q/independent-go.log |
| J01–07-UI | 原交互连续旅程 | PA/TA/DA/财务各真实会话 | 隔离安装 | 浏览器原角色链，三宽 | 7旅程checkpoints逐宽通过，不靠隐藏写入补步骤 | UI-E2E | Pass | E/browser/continuous-{1280,1440,390}-*.json；core-journeys-qa-23 |
| SELF-UI | PR35首次自托管与可读性 | PA/TA/DA | upstream/model/tenant均0 | 同宽12步骤：失败纠正→接入→授权→普通/SSE→同ID详情 | 3+4 Tokens；Fee21/成本7；长技术值/表格键盘可读 | UI-E2E / Data Reconciliation | Pass | E/browser/self-hosted-{1280,1440,390}-*.json及PNG/integrity |
| ASYNC-UI | AC08/16写结果未知/双击 | tenant admin/developer | 正常编辑器 | 丢响应读回503/404/200、轮换、双击普通/SSE | mutation1、不重放secret、并行过渡409；proxy/admit/attempt/dispatch各1 | UI-E2E | Pass | E/browser/fault-*.json、editor-interruption-*.json |
| QA-CONTRACT | 独立受影响复验 | QA | 本候选 | 8组进程内race＋53项Node | 实际exit0、不冒充HTTP或浏览器 | Contract | Pass | Q/independent-{go,node}.log、independent-checks.json |
| QA-UI | QA本人重复浏览器 | QA沙箱 | 无已授权动态执行入口 | 未重复host-only监听 | 本人未运行，由同版tests22替代交付门禁；本行仍未运行 | UI-E2E | Not Run | release-assessment §5 |
| ORIGINAL-RPMTPM-UI | 原两屏非零/自然窗/并发池 | 原DA独立会话 | 真实供应商前置缺失 | 专用Pipeline原旅程 | 不能由上述API/注入时钟/一般UI推导 | UI-E2E | Blocked | test-plan C；findings-and-next-actions-qa-23 |
| DS-SAFETY-EXTERNAL | 真实收费/隐藏推理/思考策略 | 精确DeepSeek模型 | 可信上界/预算未证 | 无新请求 | 完整供应商安全路径未完成 | External State | Blocked | 原DS-SAFETY-01＋本轮零外部调用边界 |
| REAL-SUCCESS/LEDGER | 成功普通/SSE/恢复/费用 | DeepSeek供应商 | H/X未知、无凭据 | 不生成/不查私有配置 | 历史not_attested不归零；24/64/90保持 | UI-E2E / External State | Blocked | 原报告；test-plan C |
| REAL-MODELS | 精确flash/pro当前可用性 | 供应商账户 | 当前无账户端点证据 | 未核验 | 原flash缺名只是历史事实，不rename | External State | Not Run | 原report/independent-review |
| H07-PUBLISH | 精确新提交/Actions/部署 | 后续Connector | merge尚未提交 | QA不执行发布 | 尚无新提交/部署事实，旧PR36head非联合发布版本 | External State | Not Run | prepare18/publish12/pr15历史回执 |
