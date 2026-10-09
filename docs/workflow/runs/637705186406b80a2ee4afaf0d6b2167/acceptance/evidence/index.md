# 当前独立验收证据索引

当前：qa执行23，2026-10-10；HEAD8917f942cb38af2f56aa93914ab62e763bf9cc6e，main/MERGE_HEAD6929376089535038112da48aed715fd729615d76，未提交联合158文件指纹6e359667d5f7521f6b8c421454c6fc7cbe5ccb8c74aa87bb1e5955c194c202cc。

结论：联合M01受控回归资产Go，可整理交付；tests22完整冷make verify及QA23独立复验通过；真实DeepSeek/原#33仍Blocked与历史No-Go，不关闭未验项。旧tests9/QA10只历史单候选。

当前产物相对本Run根：
- 报告：`acceptance/release-assessment-qa-23.md`。
- 权威分层矩阵：`acceptance/acceptance-matrix-qa-23.md`。
- 独立计划评审：`acceptance/test-plan-review-qa-23.md`。
- 核心旅程：`acceptance/core-journeys-qa-23.md`。
- 发现/候选/责任：`acceptance/findings-and-next-actions-qa-23.md`。
- 采证工具与恢复记录：`acceptance/evidence/qa-23/audit.mjs`、`finalize.mjs`、`collection-attempts.md`、`local-log-transforms.json`。
- 材料/逐文件库存/回执：`acceptance/evidence/qa-23/material-integrity.json`、`integration-recomputed.json`、`receipt-review.json`。
- 当前宿主完整证据：`acceptance/evidence/qa-23/host-tests22/state.json`、`verify.log`、`browser/`；原文件hash在`archive-index.json`。
- QA本人执行：`acceptance/evidence/qa-23/source-before.json`、`source-after.json`、`independent-checks.json`、`independent-go.log`、`independent-node.log`。

## 历史QA10（不作为联合放行依据）

本索引目录相对于本Run根：
- 历史报告：`acceptance/release-assessment-qa-10.md`。
- 历史分层矩阵：`acceptance/acceptance-matrix-qa-10.md`。
- 独立文档评审：`acceptance/test-plan-review-qa-10.md`。
- 连续旅程契约及三宽回执：`acceptance/core-journeys-qa-10.md`。
- 候选/历史事实/后续责任：`acceptance/findings-and-next-actions-qa-10.md`。
- 材料version/原路径/61文件SHA：`acceptance/evidence/qa-10/material-integrity.json`。
- 宿主tests9完整原runner日志/状态：`acceptance/evidence/qa-10/host-tests9/verify.log`、`state.json`；白名单摘录`selection.log`不可代替完整日志。
- 所有浏览器PNG/JSON/JSONL：`acceptance/evidence/qa-10/host-tests9/browser/`，具体路径与SHA由`archive-index.json`给出。
- QA本地原source指纹：`acceptance/evidence/qa-10/source-qa.json`；独立执行：`independent-go.log`、`independent-node-focused.log`、`independent-node.log`与`independent-checks.json`。
- 采证工具/限制/失败恢复：`acceptance/evidence/qa-10/collect.mjs`、`collection-attempts.md`、`receipt-review.json`、`local-log-transforms.json`。全包装log重构件不可冒称平台全段独立字节比对。
- QA归档逐文件摘要：`acceptance/evidence/qa-10/archive-index.json`。此索引不自含自己的hash，不改变源码指纹。

历史失败及研发证据继续由 `verification/evidence-index.json` 引用，tests5/7和原材料旧report未覆盖；本轮当前结论在该索引的current_qa字段关联，旧待验文档是执行8历史快照。原#33报告只能由合法真实复测证据改变，不能因本索引发布自动转Go。
