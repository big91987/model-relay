# 执行25：恢复原输入事件证据，不宣称空图根治

## 基线与责任
- HEAD仍 `3f9cee1615e87e002a2c2300f2d29778072ec3f5`；同Run T001 In Review、M01 Needs Revision，恢复执行23及独立QA14整改，不重建方案。
- 材料保持 design-v0.2.0，archive SHA256 `5f51d35fdc15ed7214749f9413879b4238097f1d6b6bbe2cc0d508d8e103cef7`。实际manifest：`.workflow-input/b7f93c24bc1ea4d0fe182f12aa7622ae/5f51d35fdc15ed7214749f9413879b4238097f1d6b6bbe2cc0d508d8e103cef7/files/manifest.json`；沿用其root `docs/prd.md`、`docs/interactions.md`、`docs/acceptance.md`。状态frozen-for-test-implementation未改；AC01/02/06/15/16/17仍要求真实网页、秘密一次展示、同ID费用与受控供应商边界。
- trellis-before-dev实际运行 `.trellis/scripts/get_context.py --mode packages`，识别单repo frontend/backend；读取frontend index及DOM/lifecycle/verification/dependency/workspace相关合同与共享reuse/cross-layer。规范已有源码依据，未重新初始化/补模板。研发Skill沿用公共总览/T001/M01及两份references检查，不建第二状态源。

## 执行24真实失败与最小归因
- 原固定宿主 `make verify`，冷起点UTC `2026-10-09T03:36:44.095917Z` 至 `2026-10-09T03:47:40.215191Z`；gate_exit2/runner_exit1/Connector_exit2。源前后154件戳 `6ce878eb3a41ec7fa0a18488c6625c09f9ef6c3d243ee66812036550ccf6a65b`一致，owned_group_absent=true、candidate_removed。
- 原248Node全过，原Go race/HTTP/预算UTC及migration/legacy/routing/console阶段通过；full-console 1280 J01–J07和summary通过，在visual-keys像素断言失败。后续完整三视口visual及本候选self_hosted未执行，不能挪用执行13绿作当前绿。
- 完整91件原字节产物已复制并逐件SHA核对，见 `host-tests-24/archive-index.json`；原runner日志455775bytes、SHA `3a9b740790f8d2e54f04c0f8492b21e52bb4e3152d337d1c223f71abddd19c01`，Connector外包装全文455956bytes，不冒称二者相同。已调用read_command_output定位执行24 retained日志；实际原runner文件全文及state为保留事实源，不以工具首段输出替代全日志。
- 实看原 `host-tests-24/browser/visual-1280/keys-right-columns-1.png`：1280x1149全单色RGB243/246/250，四锚点零墨；SHA `78a7b304df77dac78964e27f6571f81f2ac48376e545786f9778a183028bff53` 与执行20/22相同。前后root scrollY148、几何稳定，末列已达，均不能当作可用截图。
- 新CDP诊断实际available=true，3个layerTree事件、13个paint事件、detach=true；CSS布局前后1280x900、内容1280x1149，pageY148。说明“没有观测到绘制”不是本次事实，仍不能将paint事件等同PNG内容或证明Chromium原因。完整原sidecar保留，安全摘要见 `development-25/host-failure-summary.json`。
- 确定的证据缺陷：`readabilityEvidence`在截图回调前已finish/dispose事件handle，失败catch又以null调用finish，将已有事件覆盖成available:false。本次before/after scrollLeft35+client988=scrollWidth1023，键盘已到末列；仅末列predicate不能说明后续wheel投递完成。尚无本次真实wheel/capture顺序，不能据此做时序修复或推断空图因果。

## 改动边界
- 只改 `tests/browser/readability.mjs` / `.test.mjs` 和既有frontend规范、公共任务记录；不改产品UI/Go/API/DB、费用/Token/授权/秘密口径、Makefile/锁文件、Platform/Connector/权限。
- 原被动输入观察延续到原单次captureRight回调结束，finally一次finish/dispose，失败catch保留已完成receipt；原截图异常始终原样抛出。输入事件最多32个，新增最多3个固定标记wheel-dispatch-start/capture-start/capture-end，共用performance.now，输出仅有限数字root/region滚动值和枚举/布尔。标记包围整个回调，不是精确CDP栅格命令起止；额外观察可能影响时序，不能以转绿证明旧根因。
- 保留唯一ArrowRight、keyboard-scrollend、原wheel和2000ms末列断言、PNG参数/mandatory anchors/墨量/几何/secret guard。没有新增sleep、raf等待、同步等待、重拍、重试或超时扩大。观察失败明确unavailable；无wheel事件不是未投递证明。
- 新增一项失败捕获回归（含可用/不可用两分支），扩展原被动观察及投影安全用例：证实事件在采集回调期间可被记录、保留原错误、一次回调/finish/dispose及有界秘密白名单。夹具只证合同，不证实际浏览器机制。

## 本地检查与证据
- 65项相关Node全过；`make format-check vet build`、syntax/diff通过；普通/SSE受控Go fixture race两项通过。实际命令/UTC/退出码见 `development-25/local-checks.json` 和同目录日志。
- 首个定向回归因VM Array跨realm deepStrictEqual失败，原安全日志 `node-first-cross-realm-failure.log` 保留；landmarks投影改用本地Array.from后15项readability及65项相关测试通过，不降低断言。
- 初次直接go run/go test误用个人默认GOCACHE，真实operation-not-permitted失败保留；随后显式使用与原Makefile相同 `.cache/go-build`、CGO_ENABLED1、GOTOOLCHAINlocal，无手装/权限扩大，源戳及Go fixture复验exit0。链接器LC_DYSYMTAB warning保留，仅路径安全脱敏，不当失败或删除。
- trellis-check完成范围/复用/原异常/秘密投影/测试/规范同步检查，无跨业务层修改。源码前后154件戳 `bce88d0683991cef837feea8b321cc67a75a18da9b09a71760105b816f4c7824`相同，见 `development-25/source.json` / `source-after.json`。固定完整门禁仍Pending；未重跑已证稳定沙箱监听必败入口，原首次native-verify.log/state保留。

## 交tests
- 原固定冷起点 `make verify`，原锁定Playwright1.59.1及实际Chromium，全部原Node/Go race/HTTP/预算UTC/迁移/三视口J01–J07/summary/visual及首次self_hosted。新Node宿主数量必须实测，不推算Pass；完整UTC/log/退出/清理/源前后和原PNG/sidecar保留。
- 若keys再次失败，优先读其failure.readability.scroll_probe.events的同钟landmarks及事件阶段，对照PNG/合成层。区分观测不可用、wheel在回调前/中被观察到、仅几何已达和原PNG结果，不再用available:false推断浏览器行为。即使未观察到或转绿，仍不闭合历史空图/R28-04因果。
- 完整绿后仍独立QA实看当前安全截图并核QA58-01/self_hosted；QA14/QA58未Closed。requirements/design沿用已冻结范围，无偏差故跳过，未补造文档。动态原生独立浏览器、实际Owner preview已登录操作、真实供应商/支付/身份分别Not Run；受控HTTP不冒称真实供应商。发布仍归Connector。
