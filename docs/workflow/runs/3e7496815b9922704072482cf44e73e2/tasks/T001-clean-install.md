# T001 锁定验证依赖先于消费者

- 状态：**In Review**。
- 里程碑：[M01](../milestones/M01-clean-install.md)；优先级：Critical；模块：项目验证供应链。
- 负责人：development完成实现，report完成说明，publish Connector接续发布；评审：tests20、独立QA21，精确提交由授权协调者审查。
- 执行类型：TDD/Validation/Review；硬依赖：宿主Go/C、Node/npm、匹配Chromium/系统库与包源，合法发布Connector。无新增产品/架构决策依赖。
- 基线：main `cd413b5aa24a372a19c19af999a6ef61b5c41572`；任务分支 `workflow/3e7496815b9922704072482cf44e73e2`。
- 已验候选：147件 `f8befd158a543afeb1ba77778e4caf72e876f6e4b0b988dd4e6ab73de5e5eaab`，见 [源戳](../acceptance/evidence/qa-21/source-qa-21.json)。HEAD仍旧base，不冒称新已验commit。

## 目标与非目标
无node_modules的全新clone/worktree可直接使用固定 `make verify` 完成全部门禁，无手工预装、NODE_PATH或全局模块依赖；按既有授权完成新PR、精确合并与正式preview部署闭环。产品运行时、冻结需求/设计、锁定版本、现有预算/时间/迁移/UI不改变；不复制控制器、不重启旧Run、不重复PR27。

## 权威输入与材料追踪
Issue28修复边界及当前授权为任务需求，不复制冻结PRD。材料版本design-v0.2.0，archive sha256 `5f51d35fdc15ed7214749f9413879b4238097f1d6b6bbe2cc0d508d8e103cef7`，manifest状态`frozen-for-test-implementation`，未改写为新批准。

实际引用前缀：`.workflow-input/3e7496815b9922704072482cf44e73e2/5f51d35fdc15ed7214749f9413879b4238097f1d6b6bbe2cc0d508d8e103cef7/files/`。
- `manifest.json`：入口及冻结状态。
- `docs/acceptance.md`：sha256 `2792740b1579d614780856a13b6257d192736e1efc0309f1dc41405a665bef7a`，AC09–11预算/时间、AC18安装迁移，H05–07实现/QA/发布同版；其余既有UI行为不变。
- `docs/interactions.md`：sha256 `95f3490ef003fa9b19dd6b9e864dd3b0429619b8cd9d0ecec853919d6a12c801`，保持产品交互，本次不改变UI。
架构沿用历史 `docs/workflow/runs/4cd62ced6a4f98f50e3cca6f7596eb06/implementation-contract.md` 及 `docs/local-deployment.md`，仅作为契约，旧验证不证明本次通过。

## 实施结果与变更边界
- Makefile新增共享phony `verification-deps`，frontend-test/browser均先锁定npm ci；安装失败停止消费者。原全部verify门禁迁入内部verify-gates，由固定verify调用Go `tests/verify`，不能直接用内部目标替代发布入口。
- Go runner自建私有冷候选、空隔离npm配置、冷Go/npm缓存，记录前后源戳、完整stdout/stderr、退出码、最终依赖/Chromium探针与自有进程清理。源根/父/最终链接越界拒绝、取消/SIGTERM与失败诊断已最小红→绿并纳入回归。
- 隔离Go测试有界并行保留原断言；browser补路由阶段计时及摘要白名单诊断。summary19没有修产品或保存同步，不能当成历史故障根治。README与Trellis依赖规范同步。
- 无产品后台/API/DB/前端资产或锁版本变更；报告22再次核对运行时、锁及deploy workflow无差异。报告文档不进入147件源码摘要。

## 验收条件与真实结果
- [x] 原无模块frontend失败已保留：192/191/1、MODULE_NOT_FOUND，见 [原红灯](../verification/frontend-before.log) 与 [起点](../verification/clean-before.log)。正式Owner历史失败仅用户提供，不伪称恢复被覆盖的首轮日志。
- [x] 消费者安装先行、失败停止，回归在原门禁内；见 [QA独立Go回归](../acceptance/evidence/qa-21/independent-go.log)。
- [x] 正式宿主tests20原make verify冷完整exit0，source/candidate前后一致、NODE_PATH unset、无手装、清理闭合；见 [runner](../acceptance/evidence/qa-21/host-seq20/verify.log)、[原state](../acceptance/evidence/qa-21/host-seq20/state.json)。
- [x] QA21同版核对零软预算/有效UTC/schema1、2→3/UI；三视口原层J01–J07、summary、十页visual通过。独立动态浏览器Blocked和供应商Not Run不转为Pass；见 [权威矩阵](../acceptance/matrix-qa-21.md)。
- [ ] 新PR/精确已验提交合并与新main正式prepare/Owner部署、Deployment/environment_url、healthz version/schema3、真实admin及旧数据密码保持回执。无此证据不标Done。

## 验证与证据契约
共享入口只有 `make verify`；工具链/浏览器前置按项目README，入口自己准备锁定模块，不要求协调者执行本Run旧脚本。成功必须是全部原门禁结束、前后版本一致与cleanup/probe闭合；失败保留原层log/state/浏览器安全JSON，不增限、跳过或改预期。证据见 [总索引](../verification/index.md)、[QA归档索引](../acceptance/evidence/qa-21/archive-index.json)、[报告22审查](../verification/report-22-review.json)。

## 已知限制与交接
候选放行结论Go with known issues。R28-04历史Open/P2/根因未知，具体用户影响、非阻止发布的证据与再红最小同步契约见 [summary独立评估](../acceptance/summary-assessment-qa-21.md)。若完成刷新后错状态实证按P1；再红门禁No-Go，不能以一次绿或PATCH200关闭问题。

下一责任人publish按 [发布契约](../acceptance/release-handoff-qa-21.md) 与 [PR正文](../pr.md) 提交推送/新PR，授权协调者继续精确merge与Owner正式preview，不止草稿。仅本Run交付文档改变不需要重新跑旧完整失败循环；产品/测试/锁/规范摘要改变须新tests→QA。没有本轮供应商、部署或真实admin成功证据。

历次证据原位保留：QA6缺口、seq8/12中断、seq10配置红灯、seq14/16超时、seq18摘要失败及development7/9/11/13/15/17/19修订均见验证索引；新state不得回填旧state。旧144件/其他源戳不用于本次放行。QA21交接哈希保留为当时快照，不替换为本轮文档哈希。
