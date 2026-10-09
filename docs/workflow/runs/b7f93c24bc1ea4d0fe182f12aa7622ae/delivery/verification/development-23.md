# 执行23：重复空图的浏览器合成边界诊断

## 基线与恢复
- 基线HEAD仍 `3f9cee1615e87e002a2c2300f2d29778072ec3f5`，继承执行21候选154件 `66cc3cd66fd062d0716ceff32bc9889cc2140145164e0cabc4b116ae3707a3c8`。未分支、提交、推送、改数据库或部署。
- 恢复同Run总览、T001、M01、执行21记录及执行22完整宿主回执；冻结design-v0.2.0、archive SHA256 `5f51d35fdc15ed7214749f9413879b4238097f1d6b6bbe2cc0d508d8e103cef7`保持。实际材料manifest与引用为 `.workflow-input/b7f93c24bc1ea4d0fe182f12aa7622ae/5f51d35fdc15ed7214749f9413879b4238097f1d6b6bbe2cc0d508d8e103cef7/files/manifest.json` 和同root `docs/prd.md`、`docs/interactions.md`、`docs/acceptance.md`，状态仍frozen-for-test-implementation。保留秘密一次展示、同ID费用、网页业务写入、受控供应商边界。
- 使用研发交付Skill与public-delivery-contract/delivery-review-checklist维护既有公共任务，不建重复任务；trellis-before-dev实际package发现为single repo/backend+frontend，读取frontend索引、workspace/DOM/console/verification/dependency规范及两项共享thinking guides。现有规范有源码依据，不重新初始化或声称prepare initialized=true；本轮只涉及浏览器证据层。

## 原始宿主失败与最小归因
- 执行22入口UTC `2026-10-09T03:14:08.370792Z` → `2026-10-09T03:25:05.678127Z`；Connector exit2，runner exit1/gate exit2，源码戳前后66cc3c…相同。冷起点node_modules/npm cache均不存在，原入口安装原锁定依赖，无手装，candidate已移除/owned group absent。原state不改。
- 91件完整原字节产物复制并逐件SHA核对，见 `host-tests-22/archive-index.json`。原runner日志442181bytes，SHA256 `6bef2179294ecb3c6e99638903f37f883ca5b6f60032297109582cf23f8d8196`；Connector全文442362bytes含外包装，二者不宣称逐字节相同。
- 宿主245Node全部通过，前置Go race/HTTP/预算UTC迁移及legacy/routing/console阶段日志保留；1280 J01–J07、当前周期、summary已通过。full-console失败在1280 visual-keys原像素门禁，当前完整visual三视口及新self_hosted未开始；不能用历史执行13绿覆盖此次未执行项。
- 实看原 `host-tests-22/browser/visual-1280/keys-right-columns-1.png` 全空白，原像素白名单均RGB243/246/250，四锚点均零墨；SHA256 `78a7b304df77dac78964e27f6571f81f2ac48376e545786f9778a183028bff53` 与执行20相同。原 `.png.integrity.json` 保留。scrollLeft37/clientWidth988/scrollWidth1025末列已达，title/content前后几何稳定，根不溢出/秘密清理通过不代表图片可用。
- render前后ready complete、visibility visible、字体loaded、动画0；五个固定scope祖先均非hidden/inert、opacity1、visible，无transform/filter/clip。title因scrollY148不在当前视口，其余四scope命中自身树，不据此认定title CSS隐藏。白名单摘要 `development-23/host-failure-summary.json`。同次十个原截图sidecar比较 `development-23/capture-comparison.json`：其余九图y0有墨，失败图y148全空白。此关联不能证明滚动/合成/截图因果，不能仅凭它改产品、归零滚动或重拍求绿。

## 实施边界与回归
- 编码前缺口：CSS/DOM诊断不能区分浏览器合成与截图栅格层；原错误仍真失败。只改 `tests/browser/capture-integrity.mjs` / `.test.mjs`，并在既有workspace-contracts记录可执行证据契约；公共文档仅更新任务/里程碑/记录。不改Go或UI产品、API/DB/费用/Token/授权、锁文件、Makefile、Platform/Connector。
- `observeCaptureCompositor(page)`使用page-owned CDP session，只启用LayerTree并读Page.getLayoutMetrics。原单次fullPage截图前后采集CSS布局/视口；被动保留各最多8个tree/paint事件、每tree16层、每层16数字transform；记录总计数/遗漏层数，finally detach。不等事件、不加sleep/raf/重试/超时/强制重绘，不取第二截图，不改变原截图选项。只白名单数字/布尔/UTC，不存ID/URL/正文/凭据/任意协议payload或原错误。
- 不支持domain/layout、无事件或detach失败分别忠实可识别；都不是渲染通过或根因事实。新增3个Node用例进入原通配入口，验证投影容量/泄漏防护、不可用边界、清理与单次原像素失败保持。夹具证明契约，不证明真实Chromium机制；宿主需实际采集。
- 本地 `node --test` 四相关文件64项通过；`make format-check vet build`、Go race普通/SSE非监听fixture、syntax/diff exit0。实际命令/UTC/退出码见 `development-23/local-checks.json`，完整日志同目录。Go链接器LC_DYSYMTAB warning保留，exit0，不改无关测试。
- 未重跑已证稳定沙箱监听失败的完整make verify，首次 `native-verify-state.json` / `native-verify.log`保留。锁定playwright-core未安装的本地全通配失败亦保留在执行21，不手装、不以定向测试替代全门禁；执行22宿主245项是实际测得旧候选全通配绿，当前新增3项宿主数量须实际测量，不能推算Pass。
- trellis-check：新函数有回归，单一截图所有调用复用原helper；白名单投影无跨业务层变更，源/依赖/生命周期合同保持。原mandatory anchors、像素阈值、几何稳定、secret guards、原键盘/scrollend/wheel/2000ms和业务断言无修改。更新规范遵循trellis-update-spec；前后154件候选戳 `6ce878eb3a41ec7fa0a18488c6625c09f9ef6c3d243ee66812036550ccf6a65b`相同，`development-23/source.json` / `source-after.json`可核。

## 结论与下一步
- T001 In Review、M01 Needs Revision；执行23不是产品根因修复、不是完整交付，QA14-01/QA58-01未Closed，历史空图及R28-04根因未知。
- 交tests原固定冷起点make verify；保持原锁定Playwright1.59.1/实际Chromium、全部Node/Go race/HTTP/预算UTC迁移/三视口J01–J07 summary visual与首次self_hosted。保存完整安全日志、UTC/退出/清理、源前后、原PNG、render和compositor sidecar。若再失败，先比较DOM/CSS事实、CDP布局变化和事件及原像素，不轮换假设修复；无事件/观测不可用不得当作无绘制证据，observer本身可能影响时序也需保持归因限制。
- 完整绿仍交独立QA实看安全截图、逐项核QA58-01与新的自托管覆盖；不以绿证明历史根因。requirements/design跳过沿用总览依据，冻结范围未变。动态原生独立QA、实际Owner preview已登录操作、真实供应商/支付/身份分别Not Run；受控HTTP不冒称供应商联调。发布由Connector负责。
