## 修复范围

关联 #28；本次是从main `cd413b5aa24a372a19c19af999a6ef61b5c41572` 开始的新Pipeline，不重复PR27、不恢复Issue25/旧Run。

根因是干净安装时frontend-test先加载playwright-core，但旧Makefile直到最后browser目标才npm ci；缓存掩盖了依赖顺序。现在共享 `verification-deps` 是frontend-test/browser的显式前置，锁定npm安装失败即阻止消费者。固定 `make verify` 通过项目Go runner自建私有冷候选并执行全部原门禁，不靠协调者手装包、全局模块或NODE_PATH。

验证工具补齐源根/父/最终链接边界、两路原层日志、辅助命令有界取消/SIGTERM及清理回执；隔离测试有界并行、路由耗时/摘要安全诊断保留原断言。README与Trellis依赖准备契约同步。后台仍为Go；不改产品运行时、UI资产、DB、安全预算/时钟语义、依赖锁或平台部署控制器。

## 版本与验收

- 任务分支：`workflow/3e7496815b9922704072482cf44e73e2`。已验候选147件SHA256：`f8befd158a543afeb1ba77778e4caf72e876f6e4b0b988dd4e6ab73de5e5eaab`。报告时HEAD仍旧base，新提交须由Connector绑定此摘要；base HEAD不能冒称已验新提交。
- 正式宿主tests20原 `make verify` 在入口自建冷候选完整exit0：无node_modules/冷npm缓存、NODE_PATH unset、无手工预装，source/candidate前后一致、最终probe/自有清理闭合。Playwright及playwright-core 1.59.1，Chromium 147.0.7727.15；锁未变。
- 全部历史预算/回拨/HealthMethodsHTTP红分类器、格式/modules/vet、197Node、Go race/实际HTTP、build、schema1/2→3迁移、路由/console、1280/1440/390px J01–J07、summary与十页visual均完成。历史重建预期红及失败采集fixture不是当前门禁失败。
- Connector20为310815bytes、truncated=false、连续十页EOF；原runner为310718bytes，SHA256 `7c4d3952943721deae9188a6f52f860572037ac8ea39ba55d278d6bf2c7abb7b`，与state一致。两种日志范围不混称。148件原层文件已归档，报告22再次核对长度/hash与候选一致。
- 独立QA21结论 **Go with known issues，仅允许候选进入精确发布及强制正式prepare**。QA实际局部race/197Node/static绿；QA原生完整入口与动态独立浏览器为Blocked，不能写作QA本地完整绿。顶层clock worker为子进程专用skip，实际恢复覆盖详见矩阵，不把skip记Pass。

## 未决风险与未验范围

**R28-04保持Open/P2/根因未知。** seq18曾在summary-model-context-toggle/form-submit触发原严格行/列/状态断言失败，旧failure_boundary不足以确定具体失败；development19只加安全诊断，没有修同步或产品。源码可见保存关闭dialog先于异步刷新，而driver.ready只等loading hidden，缺少post-write刷新代次屏障；这是静态候选解释，不是已实证根因。seq20三视口一次真实通过，不代表历史故障已复现修复或关闭。

已证用户影响是当次验收阻断/交付延期；当前产品持续错状态P1未实证，故QA建议公开P2验证可靠性风险后继续精确prepare，并非伪称Owner已接受风险。若正式门禁再红，必须No-Go并按独立评估中的真实GET屏障最小重现、同scope post-write DOM提交契约及反向/失败回归返研发，不盲目整套重试、加睡眠/超时或放弱原断言；若刷新完成后仍错状态实证按P1处理。

受控HTTP/SSE只证明本产品行为，真实供应商为Not Run。正式prepare、合并、Owner部署、Deployment/environment_url、新healthz版本/schema3及真实admin均尚无本轮成功回执；不是已上线。旧预览9851177/schema2仅为输入历史事实，报告节点没有重新查询。平台同SHA日志覆盖问题另归Platform维护源，本仓未临时修改控制器。

## 发布与最终关闭要求

继续已有授权的新PR/精确已验提交审查合并，随后新main经 `Deploy Model Relay locally` 正式prepare，Owner `deploy=true target=preview`。必须取得真实Deployment成功、environment_url、healthz.version对应新main/schema3、真实admin页面烟测及存量数据/原密码保持证据后才能标记交付完成；不能止于草稿PR或把构建/QA放行当上线。禁止初始化/重置现有安装、手改DB、临时替代服务或扩大权限。任何产品/测试/锁/规范摘要变化需要新tests→独立QA。

## 可审阅证据

以下均为精确任务分支的仓库URL；**报告生成时尚未推送，远端可访问性待发布Connector确认**。正文发布前须将这些文件纳入同一交付提交并推送；不能链接旧main/旧commit。后续正式部署证据由责任节点追加真实回执，不在此预填成功。

- [公共交付总览](https://github.com/big91987/model-relay/blob/workflow%2F3e7496815b9922704072482cf44e73e2/docs/workflow/runs/3e7496815b9922704072482cf44e73e2/README.md)、[任务包 T001](https://github.com/big91987/model-relay/blob/workflow%2F3e7496815b9922704072482cf44e73e2/docs/workflow/runs/3e7496815b9922704072482cf44e73e2/tasks/T001-clean-install.md)、[里程碑 M01](https://github.com/big91987/model-relay/blob/workflow%2F3e7496815b9922704072482cf44e73e2/docs/workflow/runs/3e7496815b9922704072482cf44e73e2/milestones/M01-clean-install.md)。
- [独立QA21报告](https://github.com/big91987/model-relay/blob/workflow%2F3e7496815b9922704072482cf44e73e2/docs/workflow/runs/3e7496815b9922704072482cf44e73e2/acceptance/report-qa-21.md)、[权威验收矩阵](https://github.com/big91987/model-relay/blob/workflow%2F3e7496815b9922704072482cf44e73e2/docs/workflow/runs/3e7496815b9922704072482cf44e73e2/acceptance/matrix-qa-21.md)、[用户旅程](https://github.com/big91987/model-relay/blob/workflow%2F3e7496815b9922704072482cf44e73e2/docs/workflow/runs/3e7496815b9922704072482cf44e73e2/acceptance/journeys-qa-21.md)。
- [148件逐文件证据索引](https://github.com/big91987/model-relay/blob/workflow%2F3e7496815b9922704072482cf44e73e2/docs/workflow/runs/3e7496815b9922704072482cf44e73e2/acceptance/evidence/qa-21/archive-index.json)、[完整宿主runner](https://github.com/big91987/model-relay/blob/workflow%2F3e7496815b9922704072482cf44e73e2/docs/workflow/runs/3e7496815b9922704072482cf44e73e2/acceptance/evidence/qa-21/host-seq20/verify.log)、[原冷前后/清理state](https://github.com/big91987/model-relay/blob/workflow%2F3e7496815b9922704072482cf44e73e2/docs/workflow/runs/3e7496815b9922704072482cf44e73e2/acceptance/evidence/qa-21/host-seq20/state.json)、[候选源戳](https://github.com/big91987/model-relay/blob/workflow%2F3e7496815b9922704072482cf44e73e2/docs/workflow/runs/3e7496815b9922704072482cf44e73e2/acceptance/evidence/qa-21/source-qa-21.json)。
- [历史摘要独立评估与再红契约](https://github.com/big91987/model-relay/blob/workflow%2F3e7496815b9922704072482cf44e73e2/docs/workflow/runs/3e7496815b9922704072482cf44e73e2/acceptance/summary-assessment-qa-21.md)、[原seq18失败runner](https://github.com/big91987/model-relay/blob/workflow%2F3e7496815b9922704072482cf44e73e2/docs/workflow/runs/3e7496815b9922704072482cf44e73e2/verification/tests-18-verify.log)、[原seq18 state](https://github.com/big91987/model-relay/blob/workflow%2F3e7496815b9922704072482cf44e73e2/docs/workflow/runs/3e7496815b9922704072482cf44e73e2/verification/tests-18-state.json)、[原seq18安全失败JSON](https://github.com/big91987/model-relay/blob/workflow%2F3e7496815b9922704072482cf44e73e2/docs/workflow/runs/3e7496815b9922704072482cf44e73e2/verification/tests-18-continuous-1280.json)。
- [QA原生受限runner](https://github.com/big91987/model-relay/blob/workflow%2F3e7496815b9922704072482cf44e73e2/docs/workflow/runs/3e7496815b9922704072482cf44e73e2/acceptance/evidence/qa-21/qa-native/verify.log)、[报告22一致性复核](https://github.com/big91987/model-relay/blob/workflow%2F3e7496815b9922704072482cf44e73e2/docs/workflow/runs/3e7496815b9922704072482cf44e73e2/verification/report-22-review.json)、[精确发布/正式部署闭环契约](https://github.com/big91987/model-relay/blob/workflow%2F3e7496815b9922704072482cf44e73e2/docs/workflow/runs/3e7496815b9922704072482cf44e73e2/acceptance/release-handoff-qa-21.md)。

材料design-v0.2.0归档SHA256 `5f51d35fdc15ed7214749f9413879b4238097f1d6b6bbe2cc0d508d8e103cef7`；仍为frozen-for-test-implementation，不改写审批状态。材料实际引用路径/文件hash在任务包和报告22审查记录中，不将原型当后端/供应商事实。所有历次失败与中断原样保留，不用新state覆盖旧证明。
