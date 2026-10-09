# 执行17：J06账单详情截图完成边界修补

## 基线与权威输入
- Run `b7f93c24bc1ea4d0fe182f12aa7622ae`，HEAD仍 `3f9cee1615e87e002a2c2300f2d29778072ec3f5`；没有提交、切分支、推送、发布或部署。
- 材料 `design-v0.2.0`、SHA256 `5f51d35fdc15ed7214749f9413879b4238097f1d6b6bbe2cc0d508d8e103cef7`；读取prepare原路径 `.workflow-input/b7f93c24bc1ea4d0fe182f12aa7622ae/5f51d35fdc15ed7214749f9413879b4238097f1d6b6bbe2cc0d508d8e103cef7/files/manifest.json` 与其 `docs/acceptance.md` / `docs/interactions.md`，冻结状态不变。不重建PRD/架构。
- 恢复本Run delivery总览、T001、M01、执行15记录和源码规范；使用研发交付Skill公共契约与审查清单、trellis-before-dev/frontend实际规范与共享指南，完成后按trellis-check核对。现有规范非空模板，不重复bootstrap。

## 宿主原失败与最小归因
- 正式tests执行16原 `make verify`：2026-10-09T02:05:25.156229Z → 02:16:25.574845Z，gate exit2 / runner exit1，Connector exit2。前后154件源码戳 `4a5046dee6b66429a75652f7ad8c5344a60dff1448a10c7293f9423bad1cf10c` 一致，冷依赖起点、无手装、进程组清理/候选移除记录保留。
- 原68件回执/浏览器产物逐字节归档 `host-tests-16/`，完整安全日志401083 bytes / SHA256 `14c76b25a180472050f271bf57c64223b366cf9bcdb0e2692981b27b33a36b02`；`archive-index.json`含逐文件hash，`development-17/host-failure-summary.json`是有界派生摘要。不把它等同Connector401264 bytes日志，不据截断展示推断成功。
- 原Node234/234通过，原Go race/HTTP/UTC等阶段及routing/console完成；full-console1280完成J01/J02/J03/J04/J05/J07，在J06 after-correction截图完整性失败。后续summary/visual、self_hosted及1440/390完整旅程未到达，不能写Pass。
- 实际查看 `host-tests-16/browser/continuous-1280-after-correction-78043.png`：完整账单列表而非已打开详情，不是全单色。对应sidecar六个锚点均inside/ink_present=true；截图前dialog=false、title=(258,104)、content=(258,252,990,509)，截图后dialog=true、title=(645,25)、content=(620,0,660,900)。根/viewport/main未变，失败在geometry stable处。
- 代码事实：J06驱动只await按钮click就截图；生产异步按钮await账单GET后才调用 `detail(...).showModal()`。click完成不等于异步详情完成。这是本次失败的可区分驱动缺口，不是监听限制，不证明Chromium内部机制或QA14旧空图原因。页面固定2026-11-01为隔离UTC预算场景，侧车started/finished保留真实采集UTC，不混用。

## 实施边界
- 本轮只改 `tests/browser/journeys.mjs` 的 `billDetail(page,period,id)`和两处J06调用、`journeys.test.mjs`四项回归，以及 `.trellis/spec/frontend/workspace-contracts.md`合同。不新增布局框架/依赖。
- 监听精确账单GET后一次原行按钮click，要求200/精确bill ID，等待实际 `#detail` 可见并精确核完整标题，返回后执行原截图/导出。首个打开也复用此入口。GET由原网页动作触发，不额外发请求，不用API/SQL代网页写入。
- 原秘密/存储/根溢出门槛、单次fullPage PNG与像素/几何门禁、键盘scrollend→唯一wheel→2000ms精确末列断言及全部原业务断言不动；无sleep、断言重试、timeout扩大或失败转绿。
- 本轮产品JS/CSS、Go API/DB、账号/租户/授权/费用/Token配置零新增改动；前序候选保留，不将驱动修补描述为QA14根治或完整交付。

## 本地验证与Trellis-check
- 当前154件候选SHA256 `ec88f5c79c35689c7587e83e50b0c3c68ba99728dc382a2fe196088e74957619`；检查后的 `development-17/source.json` 与最终 `source-after.json` 记录一致终态。HEAD不变。
- 实际命令/UTC/退出码见 `development-17/local-checks.json`：相关Node54/54、format/vet/build、非监听Go race普通/SSE fixture、syntax/diff均exit0。原通配Node新增四项预期238，以宿主实测为准。夹具证明延迟完成/单次点击/GET路径/错误身份/status-ID-title拒绝合同，不证明原生时序。
- 首次独立源码戳命令未带项目GOCACHE，系统只读cache写入失败exit1，不算通过。改用项目既有 `.cache/go-build` 后exit0。无手装依赖；全量Node缺锁定Playwright前次失败继续保留执行15，不重复已知缺依赖命令冒充绿。
- 沙箱稳定监听限制/执行环境未变，完整入口首次受限证据仍 `native-verify-state.json` / `native-verify.log`；不重复必败入口，它不解释执行16真实宿主失败。当前完整make verify明确留给tests，局部绿不能替代冷入口。
- Trellis-check审查共用详情/截图、无敏感诊断新增、无产品跨层数据改变、原权限/预算/撤销/取消/错误恢复保留；无TypeScript。按trellis-update-spec写明签名/错误矩阵/好坏例/原入口必需测试，不建重复状态源。

## 交接与未决
- T001 In Review仅指同步修补可交tests，M01仍Needs Revision；QA14-01、QA58-01整体未Closed，R28-04历史原因仍未知。
- tests在当前戳运行原固定 `make verify`：冷Go/npm、隔离SQLite/受控HTTP、完整原Node/Go race/HTTP/预算UTC/迁移、三视口J01–J07/summary/visual及self_hosted；保留日志/UTC/源码前后/退出/清理、锁定Playwright1.59.1与实际Chromium信息，每张原PNG/sidecar。未运行后续阶段不补造Pass。
- 新失败先返回原PNG/sidecar及精确边界归因，不反复完整求绿；完整绿仍需独立QA实际逐张审图、对照冻结材料/术语分项和新self_hosted覆盖。QA14旧空图不能由此次同步推断根治。
- 动态独立原生浏览器、Owner preview登录实操、真实供应商/支付/身份继续分别Not Run；受控HTTP不是真供应商联调。无需求/架构偏差，不重写冻结输入。
