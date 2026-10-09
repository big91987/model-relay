# 执行41：主页面绘制位移与局部横滚边界

## 当前结论与恢复
- 本Run沿用基线HEAD `3f9cee1615e87e002a2c2300f2d29778072ec3f5`、冻结design-v0.2.0、原配置/时钟/权限/预算及原固定 `make verify`。执行40宿主失败不是监听限制。T001 In Review、M01 Needs Revision、QA14 No-Go保持；本轮完成最小边界修补和必要局部检查，不宣称空图根因闭合、完整交付或独立QA通过。
- 恢复时154件候选 `7f0e8c2074c493a6176a94dad788c9c65a291d2464a70e484d1fde01e0b1a658` 与执行40 source-before/after一致，见 `development-41/source-before.json`。91件原字节收据归档 `host-tests-40/` 并逐件SHA核对，完整日志574082字节/SHA `c42d8b10034c61395ba7734a1cee3cb8f261c99754c8b376b37d90283ff62c54` 与原state相同。Connector包装log.bytes574264不是原verify.log的字节数，不混同；不靠会话截断摘要当完整日志。
- 原入口UTC `2026-10-09T07:12:27.744874Z`→`07:23:37.293448Z`；gate_exit2、runner exit1（Connector包装exit2）。冷起点/源与候选不变、无手装依赖、owned_group_absent=true、candidate_removed均来自原state，仅证明本次owned清理，不猜机器其他进程。
- 材料manifest先恢复于 `.workflow-input/b7f93c24bc1ea4d0fe182f12aa7622ae/5f51d35fdc15ed7214749f9413879b4238097f1d6b6bbe2cc0d508d8e103cef7/files/manifest.json`；archive SHA256 `5f51d35fdc15ed7214749f9413879b4238097f1d6b6bbe2cc0d508d8e103cef7`。实际引用同root `docs/prd.md`、`docs/interactions.md`、`docs/acceptance.md`，特别是interactions的表格局部横滚/页面不横溢出规则；材料仍frozen-for-test-implementation，不改为批准。沿用本Run任务、盘点及QA14整改，不另建事实源或重建架构。Trellis scripts/spec实际存在且为源码合同，get_context确认单包/frontend/backend；本轮不重初始化或bootstrap空模板。

## 已证事实与待证原因
- 执行40有260Node及原前置Go race/HTTP/UTC预算/迁移/旧UI/权限等通过，full-console1280 J01–J07/summary后在visual-keys原像素门禁失败；该次完整1440/390 visual及self_hosted未执行到，不用其他套件或历史绿色外推。原PNG已人工查看、重新核SHA：1280×1149、全单色RGB243/246/250，SHA仍 `78a7b304df77dac78964e27f6571f81f2ac48376e545786f9778a183028bff53`。图与sidecar见 `host-tests-40/browser/visual-1280/keys-right-columns-1.png` 及 `.png.integrity.json`。
- 相比执行38，七个保留快照现在关联至同进程主页面树，一个归other。call marker之后主页面内容layer_quad右边界-2399至-320，绘制内容全部移到视口左边，期间有1149高度全页绘制帧；DOM/window scrollX0、末列scrollLeft27+988=1015、可见CSS及生命周期尺寸仍稳定。已证区别是主页面实际绘制位移，而不是表格未达末列、未知其他树或DOM根横向布局溢出。
- trace complete=true、loss=true、仍缺return marker；没有完整截图区间，不能仅凭上述关联证明弹性overscroll是唯一根因、绘制帧就是返回PNG所用帧或Tracing没有时序影响。最小白名单摘要在 `development-41/failure-summary.json`；完整16层投影保留原sidecar，不把大转储塞入工具摘要。受控页面before.utc为测试时钟，不当宿主UTC；宿主收据时间来自state/capture receipt。

## 编码边界与修补
- 编码前明确仅改局部横滚容器CSS、readability门禁/回归、现有frame trace类别与对应严格配置回归、已有Trellis规范。具体主页面位移提供了新的最小归因依据，选择原滚动边界所在层修补；不是继续轮换等待/截图参数/输入假设。仍需宿主证实边界修补实际消除该用户影响，不称已根治历史空图或R28-04。
- `internal/relay/web/workspace.css` 对 `.table-wrap,.technical-value` 设置 `overscroll-behavior-x:none`，阻止局部横向到边后的滚动链/局部边缘位移；保留native overflow、焦点、全文标识、全部列/动作和纵向页面滚动。没有root overflow隐藏、JS事件截获、重新定位、布局框架、API/数据库/费用/Token/授权/秘密改变。
- `tests/browser/readability.mjs` 从实际getComputedStyle读取所有表格及标识横向边界策略，新增 `horizontal_boundaries_isolated` 严格true门禁及安全失败字段；Node回归拒绝任一容器auto/contain/缺失策略，保留横滚/纵滚。fixture证明投影和合同，不证明浏览器实现。
- `tests/browser/capture-integrity.mjs` 去除广泛 `disabled-by-default-cc.debug`，仍启用原layers与timeline类别。固定Chromium147官方CategoryLayerTree按cc.debug/viz.quads/devtools.timeline.layers分组，layers仍启用快照，减少无关debug事件是loss诊断的最小类别收敛。512KiB、recordUntilFull、8快照/16层/4标记、同进程关联与清理均不变；是否实际无loss/双marker只能由宿主回执证明。
- 原唯一wheel100000、keyboard-scrollend/投递完成、精确末列bounds/2000ms、单次fullPage PNG、像素/几何/秘密及全部原业务断言保持；不加sleep、重试/重拍、弱化断言、扩预算或手装依赖，不改Platform/Connector/预览/账号/DB/分支/提交/推送。
- 官方主参考与可执行合同已写 `.trellis/spec/frontend/workspace-contracts.md` 的execution41场景：CSS Overscroll §4（none的两类边界行为与单轴语义），Chromium147 `cc/tiles/frame_viewer_instrumentation.{h,cc}`（快照类别分组）。来源只用于合同解释，不是产品抄写或供应商事实。

## Trellis检查与必要本地回归
- `development-41/local-checks.json` 保存各命令UTC/退出/日志；108相关Node通过，包括原单次截图/空图/不可用/loss/身份/清理回归及2个新边界测试。Go runner/wiring race、非监听受控上游Response/Stream合同、format-check/vet/build/go mod verify、JS语法/git diff检查均exit0。原生完整入口稳定监听限制已在 `development-check.md` / `native-verify-state.json` / `native-verify.log` 保留，环境未变，本轮不重复必败完整命令；局部绿不是完整make verify或真实HTTP网页证据。
- 非监听Go夹具链接器warning保留，临时绝对object路径作 `[temporary-linker-object]` 脱敏；local-checks记录原字节数/摘要和脱敏事实，不提交本机路径。没有本轮局部测试失败；一次辅助文档探测路径不存在，后续读取Skill明确要求的public-delivery-contract及delivery-review-checklist，无编辑副作用，不当阶段阻塞。
- Trellis检查：变更在既有容器展示/采集层，不改业务数据流或ownership；无新增框架/客户端依赖/内联注释/调试秘密；原API原值、mask、费用/权限/迁移及原Makefile/Go与npm锁保留。新的行为有严格回归，规范同步，scope只覆盖有事实依据的横滚边界及既有诊断。完整门禁和真实PNG效果Pending Host，不能填Pass。
- 最终154件候选 `d041dcdacbc5375fd5adb99bde3a917418b249ad69f0f64b276a2527ad578ab6`，局部验证前后相同，见 `development-41/source.log` / `source-after.log`。交接前再次核对与物料/归档完整性见 `development-41/final-integrity.json`。公共delivery日志不参与源码戳，避免自引用。

## 下一节点硬门槛
交tests对该候选执行原冷起点 `make verify`，保持原锁定Playwright/Chromium147、依赖/预算/权限。首先核原keys右列PNG与主页面关联trace：loss/双marker/区间以及新增horizontal_boundaries_isolated；不可用/丢失仍明确Unknown，不据绿追认历史根治。同失败若再次出现，先用已有白名单区分“边界策略未生效”“主页面仍移位”“PNG采集缺失/trace不足”，不要再循环猜CSS或扩大等待。

完整原Node/Go race/实际HTTP/预算UTC/迁移、三视口J01–J07/summary/visual和新self_hosted完整网页普通/SSE/同ID费用/失败恢复后，由独立QA逐页审实际安全图、QA58-01及QA14。正式供应商/支付/身份、独立动态原生浏览器、Owner preview已登录实操分别保持Not Run；受控HTTP不等于真实供应商。失败回development，产品修复仍本节点负责；无新增需求或架构冲突，不需澄清或改冻结输入。

## 本执行再次stop升级后的接续审计
- 用户报告平台正式stop升级，原Run/权限/模型/配置/预算不变。本轮先只读核文件与local-checks：八条本地命令已全部exit0、signal=null并结束，最后一条 `2026-10-09T07:35:57.740Z`；没有未知供应商调用或待重放业务操作。产品/测试/spec改动和公共文档均在，source.log/source-after.log/source-final.json三份候选相同。
- 停止前最后的产物审计已明确exit1：不正确地要求每个成功命令stdout/stderr非空，实际syntax-diff退出0且日志0字节。source-final.json已写、final-integrity.json尚未写，不能当最终审计成功，也不当产品测试失败。首败安全字段保留 `development-41/integrity-audit-first-failure.json`；raw工具堆栈含本机绝对路径，不提交该路径。
- 只修产物审计对“成功静默命令”的处理，不改原项目门禁或其断言，不重跑已有测试、网页/供应商调用或不确定操作。接续只读重新计算 `development-41/source-resumed.json`，仍为154件候选 `d041dcdacbc5375fd5adb99bde3a917418b249ad69f0f64b276a2527ad578ab6`；91件归档、物料三个引用及archive SHA、原入口/锁未变均重新核实，真实完成的最终审计在 `development-41/final-integrity.json`。所有局部门禁事实和仍待宿主/独立QA的范围保持。
