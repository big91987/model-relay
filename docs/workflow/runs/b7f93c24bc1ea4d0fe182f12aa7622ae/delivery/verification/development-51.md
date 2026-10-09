# 执行51：恢复宿主失败与编辑器VM依赖修复

## 基线与恢复
- HEAD仍为 `3f9cee1615e87e002a2c2300f2d29778072ec3f5`；执行49候选154件 `2741f58b06c600037e412c7216df8cd0d59fae02cad1fbf8687cc02513b8dac2` 与执行50before/after源码戳相同。现有实现/公共任务/QA14 No-Go保留，不重放任何供应商调用、账号配置或未知写入。
- 材料为design-v0.2.0，archive SHA256 `5f51d35fdc15ed7214749f9413879b4238097f1d6b6bbe2cc0d508d8e103cef7`；manifest为 `.workflow-input/b7f93c24bc1ea4d0fe182f12aa7622ae/5f51d35fdc15ed7214749f9413879b4238097f1d6b6bbe2cc0d508d8e103cef7/files/manifest.json`，引用同root的docs/prd.md、docs/interactions.md、docs/acceptance.md，冻结状态不变。沿用已有真实Trellis scripts/spec，未重新初始化。
- 执行50原state/verify.log逐字节归档在 `host-tests-50/archive-index.json`。保留日志579531bytes，SHA256 `67a7f8526194de3c5f46bee1b3a2711592cfc7298c261df7dc032f197dec39ed`，与state匹配；Connector外围总输出579713bytes不是此文件大小。UTC10:10:11.849930至10:11:08.980720，gate_exit2、runner exit1、Connector exit2，源码未变、冷起点实测、owned组清理确认。无本次浏览器截图，未伪造。
- 原44/46各275件、48的93件与49产物完整性复核见 `development-51/retained-integrity.json`。首次泛递归校验错误把材料相对docs/acceptance.md当仓库路径而ENOENT；当时host50归档已完成，不重放复制，改按明确entries数组校验。此错误不是产品或宿主门禁失败。

## 最小归因与范围
- 宿主执行50前置源码/历史/format/modules/vet完成，全量Node270项260通过、10失败：readback.test.mjs四项、save-lifecycle.test.mjs六项，均在抽取edit调用openModal时ReferenceError。失败发生在调用初始化，不是保存/安全读回/异步时序/浏览器或监听限制。该轮未进入Go race/完整browser，不能用前序绿拼接。
- 两夹具原有showModal方法，现显式注入 `openModal: dialog => dialog.showModal()`，与既有workspace/journeys抽取夹具一致。原历史缺陷重建、operation引用、读回恢复、重复写阻止、跨scope晚到、一次性秘密等断言不变。实际共享helper由workspace.test单独覆盖，VM不证明native锁机制。
- 本轮仅改上述两个测试及Trellis合同/本Run公共证据；不改workspace.js/css或真实浏览器驱动、截图、门槛、依赖锁、预算、Makefile、Connector、schema、费用、安全口径。执行49body固定锁的真实效果仍未验。

## trellis-check与实际验证
- 具体命令、UTC、退出码和日志摘要在 `development-51/local-checks.json`。相关12项Node通过。全量Node通配本地268通过、1测试模块装载失败，缺playwright-core/package.json；不手装依赖、不把这个结果写成全量通过，正式入口须冷npm ci加载原锁再执行全部270项。
- JS语法、git diff --check、format-check、go vet、build、go mod verify均通过；使用工作区Go缓存的verification wiring race通过。复核helper依赖、断言与秘密清理，无新增产品抽象或debug日志。未变更公共接口，跨层数据流无新改动。
- 原生监听限制环境未变，不重复已知受限完整make verify。以上检查不是完整固定门禁替代。执行副作用仅本地隔离测试、Go缓存/构建输出与证据文件，无账号/DB/供应商/发布操作。
- 当前154件候选 `454ba32f380c184b2bfde486bf3c38a1854c2290563c38ec26a72580c518fec8`，Go1.25.6；最终戳及完整性见 `development-51/source-final.json`、`development-51/final-integrity.json`。

## 交接与未决
T001 In Review、M01 Needs Revision、QA14 No-Go保持。tests按当前源码戳运行原冷make verify，保留十二分钟总门禁/九百秒外层、parallel4、Playwright1.59.1/Chromium147锁及原权限。全部Node/Go race/HTTP/预算UTC迁移/权限撤销取消恢复、三视口J01–J07/summary/visual，以及self_hosted首次页面写入/不可达和核验失败纠正/普通与SSE同ID日志费用必须完整，不拼旧绿。

modal原单次scrollBy零位移/root clip/dialog auto/focus_inside及body fixed/root高度、关闭精确焦点/解除fixed、一次fullPage/像素/document几何不变。同边界失败先依据原PNG/sidecar/实际位移诊断，不轮换猜测或放弱断言。完整绿后独立QA核安全截图、冻结设计和QA58-01/self_hosted；历史空图与R28-04根因、独立动态原生、真实供应商/支付/身份、Owner preview已登录实操仍分别未闭合，受控HTTP不能冒称供应商联调。
