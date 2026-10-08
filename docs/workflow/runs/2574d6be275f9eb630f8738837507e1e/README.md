# Issue30 保存读回契约修复

## 目标与范围
- 在 main `9bf6ecaf53758ca2258498968b830c535e02e1c0` 上定位并修复模型保存后同角色、同租户、同页面刷新及 DOM 提交的完成契约。当前不是功能扩展或重新设计。
- 唯一任务事实源：[T001](tasks/T001-model-save-readback.md)；里程碑：[M01](milestones/M01-model-save-readback.md)；当前验收：[QA8](acceptance/report-qa-08.md) 与 [矩阵](acceptance/acceptance-matrix-qa-08.md)。研发 [development-06](verification/development-06.md)、修前 [development-04](verification/development-04.md) 保留历史。
- 沿用 design-v0.2.0；需求/设计阶段按 intake 的 Bug 短路径跳过，不复制 PRD/架构。不修改 Platform、DB、已有凭据、部署控制器或历史回执。

## 当前状态
- 交付判断：Verifying。研发候选验收已完成，当前进入交付发布链；新提交、合并及正式预览部署尚未发生，不标 Delivered。
| 对象 | 状态 | 可依赖结果 |
|---|---|---|
| M01 保存结果可安全断言 | Accepted | 仅研发候选范围；tests7完整同版门禁及QA8评审成立，不包含正式部署验收 |
| T001 定位并修复保存完成契约 | Done | 完成契约原层红转绿，QA8 Go with known issues；R28-04保留Open/P2 |

关键路径：修前tests5真实失败 → development6候选 → tests7一次完整门禁通过 → QA8证据复核完成 → report及授权Connector精确发布/协调者核证。当前HEAD仍旧main；已验候选150件摘要0384f115adbbd545e8ce01021f238010ba66c89eb06e6acea08ce872ce0fa474，尚未新提交/部署。native限制保留，不重复必败完整入口，不扩大权限；tests5红不改绿。

下一步：publish Connector提交并推送该精确候选，消费本Run [PR正文](pr.md)，核对远端文件链接后继续既有授权发布链。没有新增待研发Ready任务；当前无阻断交付说明的验收矛盾。发布若源码失配或同版门禁红，须No-Go返回原责任节点，不沿用候选绿灯。

## 风险与边界
- R28-04 仍 Open/P2；旧 tests18 根因未知。新受控暂停即使证实早返回，也不能倒推旧执行根因。
- 若释放刷新后严格唯一行/列/状态仍错误，按核心管理用户影响升级 P1，不当作驱动暂态。
- 动态独立 QA：Blocked（native未独立驱动浏览器）；同版宿主受控UI通过与QA独立Contract/回执复核分别记录；供应商、Owner真实登录预览、发布/部署：NotRun。
- 候选建议Go with known issues，不宣布全面验收或部署。旧PR29/Owner预览事实保持历史，不补写旧状态；后续正式门禁红须No-Go返development。

更新时间：2026-10-09（Asia/Shanghai）；实际执行 UTC 见验证回执。
report9按交付Skill审查：沿用单一总览/任务/里程碑及QA证据，Bug短路径无新增需求或架构产物。2026-10-08T22:42:44.388Z复核材料19文件及归档log/state摘要；源码150件与QA8一致。仅整理交付文档，未重跑host-only动态门禁或进行Git发布。
收尾检查（UTC 2026-10-08T22:45:12.416Z）：26个QA归档文件逐一bytes/SHA256匹配；PR的16个完整仓库URL目标在本地均实际非空、ref编码匹配任务分支或既有历史提交；公共文档25个相对引用有效。`git diff --check`通过，文档整理后源码摘要仍与QA8一致。远端链接尚未验证，须publish在实际推送后核证，不将本地目标存在写成远端已可访问。
