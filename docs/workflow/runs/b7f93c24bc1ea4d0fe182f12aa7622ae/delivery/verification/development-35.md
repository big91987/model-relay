# 执行35：稳定生命周期后的原单次截图绘制帧诊断

## 真实结论与版本
- 本轮完成最小诊断扩展与本地必要回归，不是空图修复或完整交付；T001 In Review、M01 Needs Revision、QA14 No-Go保持。QA58-01与本次完整三视口/self_hosted、独立QA仍未闭合。
- HEAD仍为main基线 `3f9cee1615e87e002a2c2300f2d29778072ec3f5`。输入154件候选 `f700fb6ebcf8c8ec62ffb7e0621d62c4b86ac2e9f3a8956d0c211e1d0c1f85d8` 在执行34前后一致。当前154件候选 `d885da49ab2fe52cbc062c56dbd67293078c3fa8ed45235f776159b3572d519c`，本轮检查前后相同，见 `development-35/source.json`、`development-35/source-after.json`。本节点没有分支/提交/推送/发布。
- 材料沿用design-v0.2.0，archive SHA256 `5f51d35fdc15ed7214749f9413879b4238097f1d6b6bbe2cc0d508d8e103cef7`。实际manifest `.workflow-input/b7f93c24bc1ea4d0fe182f12aa7622ae/5f51d35fdc15ed7214749f9413879b4238097f1d6b6bbe2cc0d508d8e103cef7/files/manifest.json`；实际引用其root下 `docs/prd.md`（`e768b4f26b4471e714d92b0db6e3187c89f37a25a697c01a12c6079af07de547`）、`docs/interactions.md`（`95f3490ef003fa9b19dd6b9e864dd3b0429619b8cd9d0ecec853919d6a12c801`）、`docs/acceptance.md`（`2792740b1579d614780856a13b6257d192736e1efc0309f1dc41405a665bef7a`）。保留frozen-for-test-implementation，不改为批准；AC01/02/06/15/16/17继续要求真实UI/安全掩码/同ID用量/费用与受控供应商边界。

## 执行34原失败与最小归因
- 原固定宿主make verify：`2026-10-09T05:28:47.691491Z` → `2026-10-09T05:39:22.953765Z`，verify-gates退出2、runner退出1、外层make/Connector退出2。隔离冷起点、源码前后不变、owned group absent=true有原state。完整runner日志533487byte，SHA256 `4a51a3fc33e3cd6c91e103971c8dbbcbe1b7a1bf51a5c0ed60cfc0e4a8034332` 与state一致；Connector外包装日志533669byte，不混用两层字节/哈希。已用read_command_output读取执行34末段至eof，与原文件/收据对应；全文不依赖会话截断。
- 91件原字节文件逐件SHA核对归档于 `host-tests-34/`，具体清单 `host-tests-34/archive-index.json`。原PNG、sidecar、过程JSON、完整日志/state均保留，不改旧证据或数据库。
- 宿主255Node、原Go race/HTTP/UTC预算前置门禁、迁移/legacy/routing/console，以及1280 J01–J07/summary已经运行通过；完整控制台失败于visual-keys截图像素门禁。当前完整1440/390 visual及首次self_hosted未运行到，不能用旧绿或前置页面证明本轮完整通过。
- 实际查看原keys图，重新解码独立核对原sidecar：1280x1149，全像素RGB243/246/250；四个内容锚点全部contrast=0/dark=0。原PNG SHA256 `78a7b304df77dac78964e27f6571f81f2ac48376e545786f9778a183028bff53` 与执行32相同；DOM中标题/表格/末列均存在也不能让它可用。
- 同钟scrollend161ms、wheel-dispatch171ms、可信表内wheel191ms、capture-start197ms、capture-end285ms：投递先于截图，原同步边界已成立。不能继续以“未投递”解释空图。
- 原lifecycle available=true、6样本、omitted=0：armed0ms、screenshot-call3ms、window-resize9ms、visual-resize9ms、screenshot-return48ms、finished84ms；所有样本window/visual viewport均1280x900/scale1，window x0/y148，document/body/content可见且连接。resize通知不代表采样中出现尺寸变化；没有隐藏事实。采样不能排除未观测的合成内部状态，更不能证明因果。
- 原LayerTree三树/八paint通知不证明实际输出像素；当前缺少截图期间实际draw-time屏幕四边形与目标viewport证据。栅格/截图输出、合成坐标/裁剪以及瞬态输入效果均为待证假设；不归因产品、供应商或原生监听限制。安全派生摘要 `development-35/host-failure-summary.json`。

## 编码前边界与实际变更
- 只改 `tests/browser/capture-integrity.mjs`、所属 `capture-integrity.test.mjs` 和 `.trellis/spec/frontend/workspace-contracts.md`；公共任务/里程碑/验证记录同步维护。没有改产品Go/UI/CSS/API/数据库/费用/Token/授权、锁定依赖、Makefile、原输入或部署控制器。
- 新 `observeCaptureFrameTrace(session)` 复用原页CDP session，在原单次fullPage截图前开始受限Tracing，两个固定clock-sync marker括住同一次截图，最后结束/flush并清理监听。没有timer/sleep/RAF、断言重试、重拍/替代采集、强制重绘、时间扩张或权限变化。原截图选项仍 `{path,fullPage:true}`；原秘密/几何/全像素/锚点/末列门禁与异常身份保持。
- 512KiB recordUntilFull、禁用JS采样/system tracing、排除截图类别，限三个已知cc/DevTools timeline类别。协议原始trace、DOM、URL、截图内嵌、业务值/凭据/任意事件args都不落盘。最多8绘制帧、每帧16层、16行优先screen transform数字、8quad数字、4固定marker以及计数、loss/complete布尔。丢失/不可用/无帧明确未知。
- 浏览器Tracing是全局流；用Page.getFrameTree和SetLayerTreeId在内存关联当前主frame，仅允许显式numeric layerTreeId，分matched/other/unattributed。标识仅内存使用，结束清理，不猜测旧版hex对象ID。缺关联不能将其他页面帧算成当前页面证据；原LayerTree列优先transform也不能与trace行优先矩阵混用。
- start失败不发送end，防止结束非本轮trace；end失败/会话close保持complete=false，清理本轮监听，原CDP在finally detach。flush完成只表示采集完成，不表示PNG有效。诊断可能扰动调度；即使新完整绿，也不能据此宣称历史空图机制根治。

### 一手接口/实现参考（诊断依据，不是本机因果证明）
- Playwright锁定v1.59.1 `packages/playwright-core/src/server/chromium/crPage.ts` 的takeScreenshot，以及 `src/client/cdpSession.ts` 的session事件/清理：`https://raw.githubusercontent.com/microsoft/playwright/v1.59.1/packages/playwright-core/src/server/chromium/crPage.ts`、`https://raw.githubusercontent.com/microsoft/playwright/v1.59.1/packages/playwright-core/src/client/cdpSession.ts`。
- CDP Tracing配置/flush/clock markers：`https://raw.githubusercontent.com/ChromeDevTools/devtools-protocol/master/pdl/domains/Tracing.pdl`。
- Chromium绘制快照/层投影/主frame关联：`https://raw.githubusercontent.com/chromium/chromium/main/cc/trees/layer_tree_host_impl.cc`、`https://raw.githubusercontent.com/chromium/chromium/main/cc/layers/layer_impl.cc`、`https://raw.githubusercontent.com/chromium/chromium/main/third_party/blink/renderer/core/frame/local_frame_view.cc`。这些main接口只约束待观察白名单，不假设锁定本机Chromium一定提供所有字段；缺失不回退猜测或放绿。

## 本地验证与trellis-check
- 已读取挂载研发交付Skill及公共契约/审查清单，恢复同Run总览/T001/M01与原失败。执行trellis-before-dev package context返回single repo backend/frontend；读取实际frontend索引、DOM/console生命周期、验证/依赖/workspace合同与共享reuse/cross-layer指南。现有Trellis scripts/spec具源码合同，不是空模板；prepare initialized=false未改为重新初始化成功，无需bootstrap。trellis-update-spec维护新的可执行诊断合同；trellis-check复核最小scope、保密/清理/失败身份与原公共入口。
- 初次23项capture Node退出0，实际完整输出 `development-35/initial-node.log`。最终74项相关Node退出0；完整Go runner与wiring race、静态format/vet/build、非监听普通/SSE fixture、JS syntax及diff检查均退出0。实际命令/UTC/退出/范围/日志位置见 `development-35/local-checks.json`。全部新3项回归进入原Node wildcard，不新增弱入口，不手装依赖或更改Chromium/Playwright锁。
- 单元覆盖有界隐私、late主frame关联、unsupported/start/end失败、关闭/丢失/无关联/malformed，以及原单次截图成功/空图/异常的选项/异常身份/监听清理。Node fixture证明驱动契约，不证明浏览器合成机制。
- Go linker原LC_DYSYMTAB警告及退出0均保留；`go-fixture.log`只脱敏一处owned临时路径，local-checks记录原SHA与redactions=1，没有删除警告或伪造无告警。
- 已知原生沙箱监听环境未变，沿用 `development-check.md` 与 `native-verify-state.json` / `native-verify.log` 首次固定入口受限事实（2026-10-08T23:56:58.265233Z → 23:57:25.256346Z，shell2/runner1/gate2），本轮不重复必然同处退出的完整命令。上述本地通过不等于make verify全绿；完整race/HTTP/native browser归配置宿主tests。
- README/T001/M01沿用唯一事实源，没有第二套PRD/设计/状态表；依研发交付review-checklist核对边界、权威输入、退出条件与可消费证据，当前结论Blocked by Validation（可交tests，不是需用户输入的外部阻塞）。requirements/design均跳过：没有新增产品/架构边界。

## tests下一步与验收边界
1. 对当前候选执行原固定冷 `make verify`，保留新安装/锁定Chromium、完整安全日志/UTC/退出/清理、源码前后戳以及原PNG/sidecar；不用本地fixture或前置health替完整回执。
2. 若同空图再现，先查 `compositor.frame_trace` 是否available、complete、data_loss=false、同trace marker齐全、matched主frame。比较调用/返回标记内device viewport与draw-time quads/opacity/transform，结合原PNG及原LayerTree/lifecycle。未关联、缺失/丢失、不完整或截断均只说明诊断覆盖不足；不要由此反复切换推测修复。
3. 全部原Node、Go race/HTTP/迁移/UTC预算、三视口1280/1440/390 J01–J07/summary/visual，以及独立首次self_hosted网页写入/核验失败与纠正/授权/Key/普通/SSE/同请求用量费用仍必须完整验证。完整绿后独立QA实际核源码、旧QA58-01、新self_hosted与安全截图；历史空图及R28-04因果不因新绿自动闭合。
4. 动态原生独立浏览器QA、真实供应商/支付/身份、实际Owner preview已登录操作继续分别Not Run。受控HTTP自托管旅程不能冒称真实供应商联调；本节点不重置账号/改业务数据库/控制器/部署或通知Issue。
