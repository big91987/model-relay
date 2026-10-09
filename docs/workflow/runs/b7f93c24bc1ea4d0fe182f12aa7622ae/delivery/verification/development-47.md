# 执行47：模态背景滚动边界修复，待正式宿主复验

## 恢复与版本
- 基线 HEAD 仍为 main `3f9cee1615e87e002a2c2300f2d29778072ec3f5`；保留执行45的既有修改及所有历史失败，不重放不确定写入或供应商调用。
- 执行46正式冷 `make verify` 的275件原字节产物已保存至 `host-tests-46/archive-index.json`。runner日志654614字节，SHA256 `2319748e441f3e87f0351a61a8946dc3a20582e750d6c6443bbbcb573e10e3b0`；Connector封装日志654795字节，不混同。state证明源不变、owned进程组不存在及候选已清理；UTC 2026-10-09T09:16:32.660132Z→09:27:55.881033Z。Connector/gate退出2、runner退出1，整体失败。
- 原候选154件 `88794346a0c678b418671a625365a555a28bd0626c2533492c692428cfc9a0a8` 与交接匹配。264 Node、原Go race/实际HTTP通过，relay包282.668s；migration/legacy-ui/routing/console通过。三视口J01–J07/summary通过，1280/1440完整visual/self_hosted通过；390在keys-editor原截图几何门禁失败，后续visual和self_hosted未完成。不是超时或原生监听限制，不能拼接宽屏绿判整体通过。
- 同一design-v0.2.0材料archive SHA256 `5f51d35fdc15ed7214749f9413879b4238097f1d6b6bbe2cc0d508d8e103cef7`，manifest和实际引用的PRD/interactions/acceptance逐件摘要复核见 `development-47/failure-summary.json`；状态仍为 frozen-for-test-implementation，不修改材料或冻结设计。
- 已使用现有Trellis scripts/spec；get_context报告single-repo、frontend/backend规范。此次CSS/原固定浏览器验证边界遵循frontend DOM/lifecycle/verification/workspace/dependency和复用规范；源码规范不是空模板，不重新bootstrap或声称本次prepare初始化成功。

## 最小归因与修改边界
- 本轮只读检查原390 keys-editor安全PNG，实际可见非空编辑表单，未显示秘密。三视口原PNG摘要和原字节hash复核见 `development-47/original-png-audit.json`；这不是275件图的独立人工QA。
- 执行46新增computed证据确认弹窗前后均为fixed。根scrollY182→327；弹窗文档top254.75→399.75、标题top273.75→418.75；弹窗视口top始终72.75。窗口resize/scroll事件发生在原截图期间，原root/body/main文档几何不变。已证边界是背景根滚动导致固定弹窗文档坐标变化，不是表单重排、空图或业务失败；具体Chromium/采图内部触发机制仍未证明。
- 最小产品修复仅在 `internal/relay/web/workspace.css` 增加 `html:has(dialog:modal){overflow:clip}`。仅原生模态打开时禁止背景滚动，长表单仍由原dialog `overflow:auto`承载；不改居中编辑/右侧详情、JS秘密清理、原生关闭/Escape/焦点、API原值、授权或费用。关闭最后一个modal由原生状态撤除CSS锁，不新增生命周期标志。精确背景滚动偏移恢复不作额外承诺，焦点必须返回原触发点。
- `tests/browser/console-visual.mjs` 在原采图前增加实际根滚动边界检查：仅存在open dialog时执行一次有界scrollBy，要求native modal、根clip、dialog auto、位移零、焦点在弹层；成功与失败均保留至有界 `modal_scroll_checks`。每个详情/编辑关闭后还要求根解除锁定并恢复精确触发点焦点。原安全截图、一次fullPage采集、原像素锚点和原document几何断言完全保留，没有换截图方法、归一化、豁免、sleep或重试。
- 新Node用例检查CSS限定作用域及打开/关闭断言的负例；只证明契约，不模拟或冒称浏览器机制。新增回归沿原通配frontend-test与实际Go UI visual入口进入 `make verify`。Trellis workspace合同保存上述签名、白名单、错误矩阵和禁例，不改验收标准。

## trellis-check与局部证据
- 最新113相关Node用例通过；JS语法、git diff --check、原format/vet/build/go mod verify通过；verification wiring race以原项目workspace缓存通过。UTC/退出码/完整日志SHA见 `development-47/local-checks.json`。
- 两个局部失败同样保留：首次直接全部Node有265项、264通过、1文件失败，因原生工作区没有playwright-core/package.json，未手装依赖；首次直接Go wiring使用不可写用户缓存而失败，改为原Makefile既有workspace GOCACHE后通过。前者依赖准备由宿主原固定入口负责，后者是调用环境，不归作产品失败。日志中的绝对本机路径仅脱敏，保留原日志摘要与脱敏事实。未用这些局部绿替完整门禁。
- 原生监听限制未发生已知相关环境改变，不重复必然受限的原生完整命令；历史实际命令/退出和范围仍见 `development-check.md`。正式完整 `make verify` 明确留给已配置tests宿主，不申请权限、不改Connector或通用控制器。
- 完成源码/规范/复用/同层数据流复核：新产品变更是CSS一条规则，浏览器断言附加而非替换；无新框架、依赖、backend/schema/管理写入、生产日志或凭据。本轮只有文档、源码、局部构建/缓存副作用；没有分支、提交、推送、部署、数据库或账号操作。
- 最终154件候选 `dcdab7fc4bbdbbcc944377450a30d36cb6e20a67c65443cfb950de22a779f193`，Go1.25.6，HEAD同基线，见 `development-47/source-final.json`。最终完整性索引见 `development-47/final-integrity.json`。中间源码戳与先前局部日志保留，不覆写为最终通过。

## 交接与未决
T001 In Review、M01 Needs Revision、QA14 No-Go保持。交tests从上述新候选运行原冷 `make verify`：十二分钟总门禁/九百秒外层、parallel4、Playwright1.59.1/Chromium147锁、原权限不变。保留全部旧Node/Go race/HTTP/预算UTC迁移/权限撤销取消恢复，三视口J01–J07/summary/visual和self_hosted首次完整页面普通/SSE同ID日志费用/不可达及核验失败纠正恢复；不要拼接旧绿。

修复的真实Chromium效果仍待宿主，不能说已根治。若同边界再失败，先读modal_scroll_checks、原capture_integrity/原PNG及computed坐标，区分锁未生效、焦点/内部滚动和原采集几何，保留失败并返回development，不改弱门禁。全绿后由独立QA逐页检查安全图、冻结设计、QA58-01和self_hosted覆盖。历史空图/R28-04因果、独立动态原生浏览器、真实供应商/支付/身份及Owner preview已登录实操仍分别未闭合。受控HTTP不是真实供应商事实。
