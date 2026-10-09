# 执行33：单次截图内部生命周期诊断

## 结论与版本
- 本轮完成有界诊断及必要回归，不是空图修复或完整交付。T001 In Review、M01 Needs Revision、QA14 No-Go不解除；QA58-01处理及本轮完整self_hosted仍待实际宿主回执和独立QA。
- HEAD仍为main基线 `3f9cee1615e87e002a2c2300f2d29778072ec3f5`；154件候选前后同SHA `f700fb6ebcf8c8ec62ffb7e0621d62c4b86ac2e9f3a8956d0c211e1d0c1f85d8`，见 `development-33/source.json` 和 `development-33/source-after.json`。没有自行分支、提交、推送或发布。
- 材料沿用design-v0.2.0，archive SHA256 `5f51d35fdc15ed7214749f9413879b4238097f1d6b6bbe2cc0d508d8e103cef7`，原manifest `.workflow-input/b7f93c24bc1ea4d0fe182f12aa7622ae/5f51d35fdc15ed7214749f9413879b4238097f1d6b6bbe2cc0d508d8e103cef7/files/manifest.json`。继续引用同root的 `docs/prd.md`、`docs/interactions.md`、`docs/acceptance.md`，保留frozen-for-test-implementation状态，不重建冻结设计。

## 执行32真实失败与最小归因
- 宿主原固定make verify开始 `2026-10-09T05:04:25.277061Z`，结束 `2026-10-09T05:15:04.489351Z`；verify-gates退出2、runner退出1。新隔离目录与冷依赖起点有记录，候选未变化，owned process group已不存在。Connector日志519410byte，原runner完整日志519228byte，两者不是同一层输出，不混用SHA或字节数。
- 原runner日志SHA `118c913871d78f3fc8279acbd22c65947cd840a601d0d08b69669c51defbcb56` 与state一致；91件原字节文件归档 `host-tests-32/archive-index.json`，包括原PNG、sidecar、过程JSON、state及完整日志。所有历史失败原样保留。
- 250Node、前置门禁及1280 J01–J07/summary已经运行通过，失败仍是visual-keys像素完整性；本次完整1440/390 visual及self_hosted步骤未运行到，不从其他截图或旧绿推导本轮通过。
- 原keys图1280×1149，全像素RGB243/246/250，四锚点均零contrast/零dark pixels。原PNG SHA `78a7b304df77dac78964e27f6571f81f2ac48376e545786f9778a183028bff53`。本轮也实际查看原图确认空白，不用DOM表格存在替代像素证据。
- 同钟键盘scrollend162ms、wheel-dispatch-start171ms、可信表内wheel190ms、capture-start195ms、capture-end283ms。wheel已在capture前被观察，末列29+988=1017可达；执行31补的输入完成边界已实测成立，但空图状态未变。不能再把同一失败归因为“wheel尚未投递”，也不能据微任务确认声称默认动作或合成完成。
- 原sidecar前后CSS viewport均1280×900/scrollY148、content size1280×1149；3次LayerTree事件、8次paint通知及visible CSS依然不能证明PNG有效。此前采样仅在截图调用外部，不能区分调用期间的瞬态视口/可见性与合成/栅格链路问题。后者均为待证假设，不归入产品故障或原生监听限制。
- 白名单摘要见 `development-33/host-failure-summary.json`。不依赖会话截断输出；完整原日志在归档中可定位。

## 编码边界与实际变更
- 最小缺口在 `tests/browser/capture-integrity.mjs` 的唯一fullPage截图期间：前后状态不能呈现中间resize/visibility事件。只增加该helper、所属Node回归以及现有Trellis合同；不改产品、API、数据库、费用、授权、锁定依赖、Makefile、滚动输入或部署控制器。
- `observeCaptureLifecycle(page)` 在原单次截图前注册被动window/visualViewport resize、scroll及document visibilitychange。记录最多16个同钟样本和省略数，标记armed、screenshot-call、screenshot-return、finished。样本只有viewport/scroll数字、document可见布尔及body/content connected/checkVisibility布尔，没有业务文本、标识、CSS字符串、提示词、凭据或原始错误。
- 调用成功、空图、截图异常均在finally清理监听并dispose handle；观察不可用明确为available=false，不绕过像素、几何或秘密门禁，不改变原错误身份。首次采样在注册前，结束采样异常仍finally清理。
- 没有新增timer、sleep、RAF、断言重试、重拍、等待扩时或替代输入。原截图参数仍 `{path,fullPage:true}`，捕获仍只有一次，原失败断言继续失败。诊断样本不作为通过条件或浏览器机制证明。
- 新3项Node回归覆盖瞬态resize/隐藏、16件上界与省略计数、初始化/结束观察异常清理、原PNG成功/空图/截图异常的单次捕获及失败身份。由既有Node wildcard进入原make verify，不新增弱入口。VM只证明采样与驱动契约。

## 本地验证与trellis-check
- 读取现有Trellis package context（single repo/backend/frontend）、相关frontend索引、生命周期/验证/依赖与workspace合同、共享指南；已有scripts/spec具源码约束，不是空白模板，无需bootstrap或重新初始化。prepare initialized=false没有被改写为本次初始化成功。
- 最终71相关Node、完整tests/verify race、tests/verification race、format-check/vet/build、非监听普通/SSE Go fixture race、syntax/diff均退出0；原Makefile工作区Go cache/CGO/toolchain保持。命令、UTC、退出与日志见 `development-33/local-checks.json`。初始37项局部检查保留 `development-33/initial-node.log`，不当作最终源码戳门禁。
- trellis-check审查本轮单层边界、异常原身份、listener/handle所有权、失败清理、严格像素/秘密/几何断言、唯一输入/PNG、原入口与锁定依赖未改变；以trellis-update-spec在现有workspace合同保存这一诊断边界。没有跨产品HTTP/存储接口变化；不新增框架、依赖或独立状态源。
- 沿用稳定原生监听限制 `native-verify-state.json` / `native-verify.log`，环境未变化，本轮不重复必败完整命令、不请求扩大权限。局部通过不是make verify完整通过；新的诊断须由已配置tests宿主实际执行。
- 首次证据完整性扫描在本地go-fixture日志发现Go linker警告含临时绝对路径，保留 `development-33/initial-integrity.json`，仅遮蔽该路径前缀、保留完整警告及真实退出0。原日志在忽略的test-results留存；原SHA与单处脱敏说明写入local-checks，不把脱敏文本冒充原字节。宿主91件归档仍逐字节保持。最终审计见 `development-33/final-integrity.json`。

## 交接门槛与下一步
- 原固定冷 `make verify` 在新候选上完整执行，保留源码前后、锁定Playwright1.59.1/Chromium、完整日志/UTC/退出/清理及原PNG/sidecars；不手装依赖或改Connector权限。
- 若相同空图仍失败，先比较新的lifecycle samples在screenshot-call→return间是否resize/visibility变化，结合原compositor和PNG区分截图期间瞬态与稳定布局下栅格问题。样本若available=false先读观察边界，不轮换推测修复或完整重复求绿。resize存在也仅是证据，不自动证明根因。
- 全部原Node/Go race/HTTP/预算UTC迁移、1280/1440/390 J01–J07/summary/visual及独立新安装self_hosted首次网页创建→不可达/核验失败→纠正→模型→授权→Key→普通/SSE→同ID日志费用保持为必需门禁。不得用API/SQL代网页写入。完整绿后仍必须独立QA逐项核原材料、QA58-01和实际安全截图。
- requirements/design路由本轮不适用：没有新增产品范围或冻结设计偏差。动态原生独立QA、真实供应商/支付/身份、Owner preview已登录实操继续分别Not Run；受控HTTP不是供应商联调。R28-04及历史空图因果未闭合，绿色也不宣称根治。发布由已有授权Connector核证处理，本节点不发布。
