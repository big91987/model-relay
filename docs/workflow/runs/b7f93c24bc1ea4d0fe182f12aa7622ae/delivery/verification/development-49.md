# 执行49：原生模态背景锁失败后的最小产品修复

## 接续与范围
- HEAD仍为main `3f9cee1615e87e002a2c2300f2d29778072ec3f5`，未创建分支、提交、推送或发布。原执行47候选154件 `dcdab7fc4bbdbbcc944377450a30d36cb6e20a67c65443cfb950de22a779f193` 与执行48前后源码戳一致。本轮接续的是失败门槛，不重放不确定供应商调用、部署或账号操作。
- 读取交接任务、里程碑、开发记录、Trellis frontend checklist/DOM/lifecycle/verification/依赖和复用规范；实际packages discovery为single-repo。现有源码规范可用，不重复初始化/bootstrap。使用研发交付Skill及public-delivery-contract/delivery-review-checklist，在同Run维护原公共任务，不新建平行事实源。
- 原材料design-v0.2.0，archive SHA256 `5f51d35fdc15ed7214749f9413879b4238097f1d6b6bbe2cc0d508d8e103cef7`；manifest及实际引用仍为 `.workflow-input/b7f93c24bc1ea4d0fe182f12aa7622ae/5f51d35fdc15ed7214749f9413879b4238097f1d6b6bbe2cc0d508d8e103cef7/files/manifest.json` 和同root `docs/prd.md`、`docs/interactions.md`、`docs/acceptance.md`。版本与frozen-for-test-implementation状态不变，不改写冻结产品基线。引用/版本已由原总览记录，不执行材料脚本。
- 最小行为缺口在产品模态背景滚动边界：原CSS根clip被实际程序滚动穿透。修改workspace CSS/共享modal生命周期和对应回归；不改API、SQL、费用、Token、授权、原生弹窗定位、秘密清理、截图机制或Platform。requirements/design不适用，理由是没有改变冻结业务或架构契约。

## 执行48实际失败与保留
- 原固定冷 `make verify` 宿主收据完整归档 `host-tests-48/`：93件原字节，689931byte完整内部日志，SHA256 `ecfa54c7d1f88d4b779a89ae2c3451bf2bbc93b71e0975f32f681110568c4404`，与原state核对一致。Connector外层日志690113byte是另一层输出，不混同内部日志。UTC、gate退出2/runner退出1及owned清理见原state；Connector回执退出2。源/候选不变，owned_group_absent=true，候选已移除，不据缺失输出推断通过。
- 267项Node、原Go race/HTTP等前置门禁通过；relay耗时274.591s。migration、legacy-ui、routing、console阶段通过。1280 J01–J07/current-period/summary通过，但full-console visual keys-detail严格失败。该次尚未进入后续完整1440/390 visual与新self_hosted，不能拼接旧绿。
- 最小诊断有五个实际modal探针：前四个位移0；keys-detail native_modal=true、root_overflow=clip、dialog_overflow=auto、focus_inside=true，但scrollBy后background_scroll_delta=100。因此“根clip足以禁止背景程序滚动”已被实测否定；前四个局部0也不能作为通用锁证据。失败发生在keys-detail原PNG采集前，不伪造该截图，不把失败归入原生监听限制。实际Chromium内部成因仍未证明。
- 原summary和失败边界见 `development-49/failure-summary.json`；全部原日志/现有PNG/sidecar见 `host-tests-48/archive-index.json`。执行44与46各275件原证据、执行47实际产物再次核hash未变，见 `development-49/retained-integrity.json`。

## 修复与原门禁
- `openModal(dialog)`统一详情、编辑和一次性秘密/邀请的showModal入口：首次打开前捕获原scrollX/Y，用仅属于该锁的body CSS变量和class把背景固定在原偏移，宽度100%。保留原根clip、dialog auto和native showModal/close/Escape/焦点，不transform或移动弹窗，不换原fullPage采图。
- `releaseModalScroll()`只在最后一个native modal关闭后移除本锁class/变量，恢复原背景位置且清除持有；已有close事件处理Escape/按钮，invalidate在关两弹窗后同步释放再清理秘密。立即重开或跨弹窗转换时，旧排队close不能解开新modal。原生showModal异常释放锁并原样抛出。没有observer、sleep、轮询、额外状态持久化或新框架。
- 浏览器原单次scrollBy/零位移、root clip、dialog auto、focus_inside断言全部保留，附加body_position=fixed；白名单仅补root scroll/client heights用于失败归因。关闭附加body解除固定，原精确trigger焦点/根解锁断言不减。原PNG像素及document几何、秘密、三视口横滚/根宽度门槛完全保留，不豁免或归一化modal。
- 新VM回归检查只捕获一次原偏移、重开/排队close守护、幂等恢复、原生打开失败和invalidate释放顺序；实际机制仍必须宿主证明。两个原抽取函数测试显式接入openModal依赖，不削弱原保存/日志导航/秘密断言。新增用例自动进入原frontend-test和真实Go UI visual，Makefile/锁/总预算/Connector未变。Trellis update-spec记录执行48否定旧提案与新签名/错误边界/必验项。

## trellis-check及当前候选
- 最终116相关Node通过，原format/vet/build/go mod verify、JS语法、git diff --check、workspace缓存下verification wiring race通过。UTC/命令/退出码/完整日志SHA见 `development-49/local-checks.json`。
- 首次相关Node116项中115通过、1失败：抽取detail函数的既有VM测试缺openModal依赖，ReferenceError原日志保留。已补测试fixture明确依赖并116项复验通过，不将其归类环境限制，也不覆盖首败记录。
- 完成frontend合同、共享函数复用/close与invalidate/秘密清理及范围复核。没有backend/schema/业务写入变化；无凭据、生产日志或外部调用。局部构建与workspace缓存是本轮执行副作用，非发布/账号/DB改变。
- 最终154件候选 `2741f58b06c600037e412c7216df8cd0d59fae02cad1fbf8687cc02513b8dac2` / Go1.25.6 / HEAD不变，见 `development-49/source-final.json`。最终完整性索引见 `development-49/final-integrity.json`。本轮修改后的浏览器机制、完整门禁和独立QA没有实测通过。
- 原生监听限制没有已知相关环境变化，按既定约束不重复必败原生完整make verify，也不手装缺失Playwright依赖。历史实际命令/退出码/代码关联仍在development-check及执行47局部失败。完整验证明确交已配置tests宿主原入口；局部Node/构建不是替代门禁。

## 交接
T001 In Review、M01 Needs Revision、QA14 No-Go维持。tests按当前源码戳执行原冷 `make verify`，保留十二分钟总门禁/九百秒外层、parallel4、Playwright1.59.1/Chromium147锁及原权限；全部Node/Go race/HTTP/预算UTC迁移/权限撤销取消恢复、三视口J01–J07/summary/visual、自托管首次页面普通/SSE同ID费用与不可达/核验失败纠正恢复必须完整，不拼旧绿。新锁真实有效性未验，不宣称根治。

若同边界失败，先读取原modal_scroll_checks的body_position/root高度/实际位移、原PNG/sidecar computed/dialog坐标、关闭解锁与焦点结果归因，再返回development；不轮换未经证据支持的布局假设、不改弱断言。完整绿后独立QA检查实际安全图/冻结基线/QA58-01/self_hosted；历史空图/R28-04成因、独立动态原生、真实供应商/支付/身份及Owner preview已登录实操分别保持未闭合。受控HTTP不代表真实供应商联调。
