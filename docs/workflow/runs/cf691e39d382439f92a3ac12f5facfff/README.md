# 验证历史保护交付计划

## 目标与边界
Issue #14：修复固定验证覆盖历史证据的工程缺陷，完整保留历史红灯分类、当前 HTTP/版本、Go race、前端和真实 Go 浏览器门禁。原始授权与可观察 AC 已落实在 [T001](tasks/T001-history-output.md)；沿用 [.trellis 运行契约](../../../../.trellis/spec/backend/runtime-contracts.md) 与 [项目入口](../../../../Makefile)。小型 bug 仅设一个可验收里程碑，不新增 PRD、原型或设计；intake 已判断业务、协议与架构不变，跳过 requirements/design。

基线 HEAD：`066f035bc551b2b4d031166cebdab0849084afd4`。历史背景为 [Issue #12 独立验收](../836b817833886b3188f263800b0f092f/acceptance/acceptance-report.md) 中 QA12-HISTORY-WRITE；旧任务/里程碑的待验状态与旧 QA 后续成功存在矛盾，保留原文而不复制成当前事实。本 Run 是当前状态事实源；历史验收不作为本轮绿色证据。

不改认证、数据库、UI、供应商或部署；不提交/推送/合并，不扩权，不重写旧 Issue #8/#9/#12 文档与日志。普通选择自动推荐：stdout-only，不增加无必要的输出配置，调用方选择独立新文件保存。

## 当前状态
- 报告对账时点：2026-10-06 05:08 +08:00；交付判断 **Delivered**，仅表示本轮工程修复、验收及公共说明交付，不表示 Git 发布、合并或部署。
- 当前里程碑 M01：**Accepted**；任务 T001：**Done**，宿主固定门禁及独立 QA 已完成。
- 关键路径：实现/回归 → 宿主固定 make verify exit 0 → 独立 QA Go → 报告对账和 pr.md 均完成；剩余平台 Connector 提交/推送/草稿 PR。
- 当前阻塞：无未关闭产品缺陷或研发阻塞。研发和 QA 本地监听受限 exit 2/Blocked 保留，不能改写为本地完整通过。
- 下一批 Ready 编码任务：无；发布 Connector 可消费 [PR 正文](pr.md)，无需再次安排已完成的宿主测试或独立 QA。

| 里程碑 | 结果 | 状态 | 依赖 | 退出证据 |
|---|---|---|---|---|
| [M01](milestones/M01-history-protection.md) | 验证不覆盖历史且全门禁可追溯 | Accepted | T001、宿主测试、独立 QA（均完成） | [报告对账](verification/M01-development.md#报告阶段对账)、[独立验收 Go](acceptance/acceptance-report.md) |

| 任务 | 负责人 | 状态 | 硬依赖 | 验证 |
|---|---|---|---|---|
| [T001](tasks/T001-history-output.md) | development；独立 QA 评审 | Done | 已有需求与运行契约 | 同版宿主固定门禁 exit 0；独立红转绿/失败路径；106 份历史校验通过 |

## 风险与下一步
deterministic 子进程回归不是实际网络/浏览器证据，不能替代 make verify。本轮完整门禁依据 [宿主回执核验](acceptance/evidence/host-connector-receipt.md)：seq=5 exit 0、同版摘要及未修改的串行 fail-fast 链；宿主文本截断，未取得后段原始日志/截图，不虚构细粒度输出或视觉质量。独立 QA 正常/异常/输出失败及历史完整性证据见 [验收矩阵](acceptance/acceptance-matrix.md)；既有 P3 QA12-HISTORY-WRITE 本轮关闭。全部本地失败保留，不恢复旧历史来掩盖副作用。

报告已按本 Run 任务事实源及本轮同版宿主/QA 证据同步 T001、M01、本总览和验证结论，并形成 `pr.md`。研发验证和 QA 验收文件中的原阶段时点/待报告描述作为历史快照保留，当前状态以任务与本总览为准。提交、推送、PR、合并、部署在报告时点均未发生；后续 Connector 负责 Git 发布和草稿 PR，不自行部署或合并。无需用户新增决定。

## 最近更新
2026-10-06：删除历史路径写入，检查 stdout 错误，补进程级成功/失败/保护回归，更新 README 和已有 Trellis 契约，保留红绿及受限门禁日志；公共计划审查结论 Ready for Dispatch（分发宿主验证，不代表里程碑验收）。无需用户新增决定。

2026-10-06 05:08 +08:00：恢复 seq=5 宿主成功和 05:01 独立 QA Go，逐条核对 AC 后完成报告对账；T001 In Review → Done，M01 Ready for Review → Accepted，总览 Verifying → Delivered（工程交付范围），交平台草稿 PR 发布。
