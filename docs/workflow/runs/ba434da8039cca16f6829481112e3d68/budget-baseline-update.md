# Issue63 返工候选测试预算基线更新（正式输入接续）

Run `ba434da8039cca16f6829481112e3d68`；2026-10-25。本文仅追加预算处理依据，不修改冻结v1/v2、旧阻塞记录、旧固定回执或首次QA Fail，不是QA签字、功能通过或发布许可。

## 正式依据与核验

来源：本研发节点恢复时收到的Harness Builder正式User Input，首句为“Harness Builder依据用户既有普通选择/持续完成本轮功能评测授权，正式确认本次测试成本估计更新”。输入明确批准以本Run tests3实测完整主gate替代旧设施任务成本参考，并对基线及两条完整增量命令一起加10%；要求保留所有旧记录、限时及覆盖，记录后直接handoff tests，无须对同一成本选择再次等待。

本次只执行文件/bytes/SHA与源码摘要核验，没有运行完整make verify、实验runner、浏览器旅程、依赖安装或隔离candidate预热。核验原QA交接74件未改；核验148条manifest记录（含交叉引用重复项，不称148个独立文件），批准前提全部成立。详见 `harness-output/issue63-budget-update/validation.log`。

- 新候选：171件全源码摘要 `7f98fee14c6ca5337d981b66b0d7dd1ed88579123c492674c7ed1ed3da8f69e0`，HEAD仍为基线 `f06f05de20ffac5b8c8e772958c808379001f0e3`，功能修改在工作树；新源码没有固定/独立QA通过结果。
- 返工仍仅六文件：routing.go、routing_latency.go、routing_latency_recovery_test.go、web/workspace.js、tests/browser/workspace.test.mjs、.trellis/spec/backend/routing-contracts.md；失效水位、拒绝诊断、相关回归/契约之外无变化。Makefile、固定runner及原旅程阶段均匹配此前候选bytes/SHA。
- plan-v2 SHA `0944af31943b26633e3813280dfe84efc5120f13dfc340221dcd67356be80a2f`；experiment-v2 SHA `4f72f2061bc9c61c7b1f33e3ca1d8f2ab86667c7ea5743d1ed716d5b04727675`，v1/v2原件不变。
- 原返工记录SHA `7eb1717b1f8a6a2166ba4a532b5dcd412a51874dd66d9a4013ec55a0e187a6be`，原budget.json SHA `faead4fe00634faba2ab9d2ef40c8578141ee1b8734f373a2b1baa1a514e6e41`，原manifest亦未重建或覆盖。

## 已批准的成本估计

|项|秒|原始依据|
|---|---:|---|
|本Run tests3完整主gate成本参考|1215.138100|test-results/verification/20261025T125233Z-3805794132/state.json及verify.log；精确时间差1215.138099975秒，依正式输入六位小数计账|
|完整定向watermark命令再次作为增量|46.759|harness-output/issue63-rework-1/watermark-regression-run.log，exit0/EOF，含首次依赖准备|
|完整静态/frontend/vet/build命令再次作为增量|91.440|harness-output/issue63-rework-1/static-frontend-vet-build.log，exit0/EOF，含原有覆盖和冷编译|
|合计|1353.337100|重复计入既有成本，未通过删测试抵扣|
|整个合计10%预留|135.333710|包括增量，不仅对旧gate预留|
|预测主gate|**1488.670810**|`(1215.138100 + 46.759 + 91.440) × 1.10`|
|距1500秒余量|**11.329190**|外层1800秒不借给主gate|

精确计算用整数微秒；JSON保留算法、原回执bytes/SHA、候选摘要、批准范围及风险。原冻结公式预测1577.2881秒及其停止事实继续保存在原budget.json/返工记录中；此前建议1474.850910秒仅是历史提议，本次按正式输入采用更保守的1488.670810秒。替代的是成本参考和估计方法，不是门禁结果、超时配置或验收阈值。

余量很小；调度/冷构建/归档/IO波动仍可能真实失败。旧gate属于旧源码，只是同Run成本参考，不能代替新候选固定绿色。估计不保证运行完成或通过，不据失败降低阈值、缩样本、跳阶段或循环求绿。

## 交接约束与未决项

按原入口make verify对新候选完整执行；所有原覆盖与阶段、2480请求规模、80%/35%/75%及40次门槛、冷起点/隔离清理、内部25分钟/外层30分钟原样保留。研发不提前自跑全套、不预热candidate、不改权限/凭据、推送、创建PR或发布。若新完整门禁红，保留首份完整日志至EOF/原退出码，定位并实质修复或获得新可复验证据后再交测；不盲重试。

QA63-01仍待新候选固定门禁及独立QA原负例、普通/SSE同跨池失效/冷却/恢复复验后关闭；研发定向通过不代签QA。首次QA报告及原两份exit1继续留存，新固定与QA必须绑定精确源码。真实供应商/生产性能与raw trace仍Not Run。

本次预算处理解除的是交测前估计停止条件，不解除QA63-01或最终交付门禁。后续仍走tests→独立QA→approval→report→固定publisher，仅到可审阅同版draft PR，不合并、不部署、不修改用户预览；最终报告与释放后留存核验继续由原流程完成。节点驻留包含预算等待/归档，不称Agent净耗时；正式输入精确平台时间/token/费用在本节点不可得，不估算。
