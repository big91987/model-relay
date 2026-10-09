# M01 内部安全与计量可复验

- 状态：Accepted（2026-10-10，QA23联合受控验收；report24汇总）
- 负责人：development；独立评审：QA23；交付负责人：report24。
- 硬依赖/进入依据：同版源码、61项材料/交接ZIP、既有实现契约及Trellis规范。
- 覆盖：原AC09–13/16及关联AC07/17，PR35主线与PR36回归资产联合；不承诺原#33真实供应商验收。

## 阶段结果及退出条件

[T001](../tasks/T001-metering.md) Done，形成实际受控普通/SSE安全默认、非零阈值/窗口、终态计量、并发原子性及池共享/隔离数值证据；共同保留主线首次自托管/可读性行为及全部断言。确认生产缺陷0，不能称修好了算法故障。

- [x] prepare18受信整合精确main，全部原行为、断言、失败证据及固定入口保留，冲突0。
- [x] 联合158文件源码指纹与tests22/QA23/report24一致。
- [x] tests22固定完整冷make verify exit0，原deadline未扩展。
- [x] QA23计划A/B独立评审与受影响复验、三宽宿主网页及版本/数值/截图证据复核完成。
- [x] 确认缺陷、内部未复现及真实未验层分列；原#33 No-Go、budget not_attested保持。

## 集成与验证证据

HEAD `8917f942cb38af2f56aa93914ab62e763bf9cc6e`；main/MERGE_HEAD `6929376089535038112da48aed715fd729615d76`；158文件SHA256 `6e359667d5f7521f6b8c421454c6fc7cbe5ccb8c74aa87bb1e5955c194c202cc`，仍为未提交联合工作树、非干净，不是新发布SHA。report仅文本更新，源码未变。

[QA23验收](../acceptance/release-assessment-qa-23.md)、[矩阵](../acceptance/acceptance-matrix-qa-23.md)、[计划评审](../acceptance/test-plan-review-qa-23.md)、[旅程](../acceptance/core-journeys-qa-23.md)、[宿主state](../acceptance/evidence/qa-23/host-tests22/state.json)支持本范围Accepted。环境Go1.25.6、Node26.0.0、Playwright1.59.1、Chromium147.0.7727.15；临时SQLite、受控上游，非部署。完整门禁653.728秒；QA本人8组race及53项Node通过，不冒充本人重跑浏览器。

## 已知风险与后续

Accepted是内部集成验收，不等于G4版本发布。report整理交付已完成；原publish/pr提交推送并更新同PR36、精确新提交及正式Actions门禁仍待执行，不将发布混入已满足的内部退出条件。源变更需重新验证；约66秒余量不保证慢环境永不超时。

供应商安全/隐藏推理/思考、成功/恢复/真账及原非零自然窗/双会话网页仍Blocked/Not Run；合法专用测试入口承担后续，原24/64/90及未知预算保持，新增外部请求0。历史状态见[接手前快照](../verification/public-state-before-report24.md)；旧tests9/QA10不替联合验收，tests5/7及exec17/21失败证据不重写。
