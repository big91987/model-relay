# 执行19：services像素锚点裁剪修补与待验交接

## 恢复与范围
- Run b7f93c24bc1ea4d0fe182f12aa7622ae；HEAD仍 `3f9cee1615e87e002a2c2300f2d29778072ec3f5`。恢复delivery总览/T001/M01、执行17交接及执行18原宿主回执，不重建方案或另建任务状态源。
- 冻结material仍design-v0.2.0，archive SHA256 `5f51d35fdc15ed7214749f9413879b4238097f1d6b6bbe2cc0d508d8e103cef7`。读取 `.workflow-input/b7f93c24bc1ea4d0fe182f12aa7622ae/5f51d35fdc15ed7214749f9413879b4238097f1d6b6bbe2cc0d508d8e103cef7/files/manifest.json`；实际沿用该root的 `docs/prd.md`、`docs/interactions.md`、`docs/acceptance.md` 及本Run `implementation/terminology-inventory.md`。不改材料评审状态。
- 使用managing-engineering-delivery-cn及其公共契约/审查references维护既有任务；trellis-before-dev读实际frontend/guides规范，get_context返回single-repo/backend+frontend。已有脚本与源码支撑规范，无待填模板，不重复bootstrap、不声称prepare重新初始化。无需需求/设计返工；仅原Harness采样缺口。

## 原宿主失败与最小归因
- 执行18真实 `make verify`：2026-10-09T02:28:07.374121Z → 02:39:16.546503Z；Connector exit2、gate exit2、runner exit1。前后154件候选 `ec88f5c79c35689c7587e83e50b0c3c68ba99728dc382a2fe196088e74957619` 相同；冷依赖/清理与源码一致性仍在原state，不将其失败归为沙箱监听。
- `host-tests-18/archive-index.json`保存83件原字节文件及逐件SHA；原安全日志415987 bytes / SHA256 `75dd381e6062510abcc62287ce5bec5d91930a603071521629961652d07b245c`。它不同于Connector416169byte外层日志，不以工具截断展示代替全文。`development-19/host-failure-summary.json`是有界派生事实，不覆盖回执。
- 原238Node通过，Go race/HTTP/UTC/迁移等前置门禁、routing/console均完成；full-console1280完成J01/J02/J03/J04/J05/J07/J06、current-period与management-summary。实际复看J06 after-correction原PNG：详情已打开，sidecar前后dialog=true、两锚点均有墨、integrity passed。本轮不改该已实测同步。
- 随后services最右列截图失败。原驱动scrollLeft40→67，clientWidth988/scrollWidth1055，最末列right1246.71875位于region right1248内；错误stack到readability.mjs:179的captureRight回调，非末列断言。截图前后标题/主体几何稳定。实际复看原图：完整服务表、中文核验通过与右端全部动作可见，不是单色或sidebar-only。
- 唯一无墨锚点left192,width110落到table-wrap left258之外的区域；它来自通用candidate只检查文档宽度、不检查父表格裁剪。dark_pixels0/ink=false；标题和单独表头/数据锚点均true。失效的是采样框，不是末列交互或已证产品内容缺失。
- 离线在同一原PNG按已记录的完整可见表头重建第二锚点，原像素门禁通过；原锚点原断言依然失败。这仅验证该PNG的像素归因，明确不是新原生采集、不是QA14历史空图或Chromium内部根因证明。页面固定2026-11-01是既有隔离UTC预算场景；采集UTC仍为实际2026-10-09，不混淆。

## 最小实施与原层断言
- 本轮编码只改 `tests/browser/capture-integrity.mjs`、对应Node测试及 `.trellis/spec/frontend/workspace-contracts.md`；公共任务/证据跟进，不新增产品、依赖、架构或布局框架。
- `captureBoundaryFacts(page)`对位于table-wrap内的通用内容候选，要求整框位于 `tableBox.left + table.clientLeft` 到 `clipLeft + table.clientWidth` 的真实client横向裁剪区。不可见候选被排除后仍必须存在内容锚点；没有可见内容保留null并失败。表头/数据独立采样、单次fullPage PNG、全部墨像素/几何阈值不变。
- 原秘密/存储安全门禁、键盘scrollend→一次wheel→2000ms精确末列断言、每表头/数据门禁及业务写入/费用/权限/撤销/预算/取消/错误恢复保持。无sleep、重试断言、增加timeout、重置滚动代替末列截图、API/SQL代网页操作或Platform修改。
- 新3项Node回归：左/右/全部裁剪文字不能成为content anchor；client边框参与范围；document Y坐标准确；无可见内容仍拒绝，blank/sidebar-only仍拒绝。原测试保持；夹具不证明原生浏览器机制。

## 实际本地验证与Trellis-check
- `development-19/local-checks.json`记录实际命令/UTC/退出码及日志：57相关Node、format/vet/build、非监听Go race普通/SSE fixture、syntax/diff均exit0。新增3项进入原通配门禁，完整Node预期241，以宿主实测为准；没有手装依赖或将缺Playwright的本地全量测试冒称通过。
- 新用例修补前7Pass/2Fail记录 `anchor-red.log`；过程中夹具缺DOM clientLeft、单元格尺寸小于原24px门槛分别保留 `anchor-fixture-missing-border.log` / `anchor-fixture-too-small.log`，均exit1。只补夹具正确DOM属性/尺寸，不更改原24px门槛。这三次早期命令UTC未独立采集，不编造时间；日志本机路径已脱敏，失败不是产品/环境通过证据。
- Trellis-check检查变更范围、共用采集职责、无文本/secret诊断、原safeCapture顺序和单次截图/throw合同、跨层数据流无新增改动；无TypeScript层。trellis-update-spec记录签名/裁剪合同/错误与好坏例/必需测试，不建立重复规范。
- 新154件源码戳 `566a5eb651efd268740b9d2f876b8e860015da062acc7e3fcd7bee63403f2e15`，`source.json`和`source-after.json`应一致；delivery回执依项目源码清单排除自引用。
- 原生沙箱稳定监听限制未变，首次原 `native-verify-state.json` / `native-verify.log`仍保留，不重复必败完整入口。它不解释宿主执行18失败；当前完整make verify只交已配置tests宿主，不把本地子集变绿当完整交付。

## 交接结论与未决
- T001 In Review仅表示采样修补可交tests；M01仍Needs Revision，QA14-01/QA58-01未Closed，旧空图与R28-04历史根因未知。当前未到达的后续visual、self_hosted及1440/390完整旅程不写Pass。
- tests须冷起点运行原固定make verify，在当前戳保留全部Node/Go race/真实HTTP/预算UTC/迁移、三视口J01–J07/summary/visual和首次self_hosted；保留完整安全日志/真实UTC/退出/清理/源码前后、锁定Playwright1.59.1与实际Chromium事实及全部原PNG/sidecar。不能手装依赖、改Connector或放弱门禁；新失败先返回精确原层证据，不反复完整求绿。
- 完整宿主绿后仍交独立QA实际逐张检查冻结设计/QA58分项与self_hosted覆盖；动态独立原生浏览器、Owner preview已登录实操、真实供应商/支付/身份继续分别Not Run，受控HTTP不是供应商联调。保持原账号/租户/配置只读，不声称console_clock等整库所有行逐字节不变；发布/PR/部署由Connector负责。
