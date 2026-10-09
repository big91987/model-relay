# 当前独立验收证据索引

当前：qa执行10，2026-10-09；HEAD3f9cee1615e87e002a2c2300f2d29778072ec3f5，154文件工作树指纹2dfe14a3c305c13621f23b7dd287a8313b7e4f39ab64b8281d283450d125d8ed。

结论：M01内部回归资产Go，可整理交付；真实DeepSeek/原#33仍Blocked与历史No-Go，不关闭未验项。

本索引目录相对于本Run根：
- 当前报告：`acceptance/release-assessment-qa-10.md`。
- 权威分层矩阵：`acceptance/acceptance-matrix-qa-10.md`。
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
