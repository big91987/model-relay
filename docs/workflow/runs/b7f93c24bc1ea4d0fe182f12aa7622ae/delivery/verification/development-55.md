# 执行55：已证正文关闭锁缺口的同步释放修复

## 恢复与版本
- Run/节点/执行：b7f93c24bc1ea4d0fe182f12aa7622ae / development / 55；HEAD仍为 `3f9cee1615e87e002a2c2300f2d29778072ec3f5`，未提交、切分支或推送。继续已有T001与M01，不重建需求/架构。
- 停止前执行53候选 `00d8519e5b54d97eccce2cae8652735bec46319108044d0d075c96da9b33ce8c` 与执行54实际前后源码戳一致，宿主报告源与候选未改变，owned进程组消失、临时候选清理。没有需要重放的不确定业务操作；执行55只修改源码、运行隔离非监听检查并生成证据，无账号、业务DB、供应商、发布或部署副作用。
- 材料保持design-v0.2.0、archive SHA256 `5f51d35fdc15ed7214749f9413879b4238097f1d6b6bbe2cc0d508d8e103cef7`、frozen-for-test-implementation。manifest及实际引用的同root docs/prd.md、docs/interactions.md、docs/acceptance.md路径/摘要核验见 `development-55/retained-integrity.json`；未把材料改为批准。关闭返回原触发点契约仍保留。
- 已有Trellis脚本与源码规范存在；get_context确认单仓库frontend/backend层。按trellis-before-dev读取frontend索引、workspace合同、DOM/console生命周期、验证及依赖约束、共享复用指南。规范并非待填模板，无需重新bootstrap，未把prepare initialized=false描述为本次初始化成功。

## 执行54原失败与归因
- 原回执79件逐字节保留在 `host-tests-54/archive-index.json`；原verify.log 717118bytes，SHA256 `2a6422e33ca5daae37e6e0b3a719cd67e4824de6c3a8579c31014c90c40727cd` 与state一致。Connector外层717300bytes不等于内层日志字节数。UTC 2026-10-09T10:41:44.570333Z至10:51:56.042574Z；gate_exit2、runner_exit1、Connector exit2，未把原失败改绿。
- 原冷起点273 Node、七包Go race与前置migration/legacy-ui/routing/console通过；1280 J01–J07、当前周期visual与management-summary通过。full-console在models-detail之后 `close-detail` 的 `body-unlock` 失败；self_hosted steps/screenshots为空，1440/390 full-console未完成，不拼历史绿。
- 原models-detail.png与sidecar passed，实际查看有实质详情内容；原一次scrollBy位移0、root clip/body fixed/dialog auto/focus_inside与像素/document几何通过。失败时native modal已关闭，root overflow visible，精确trigger已聚焦，但body仍fixed且锁class存在，close_events=0。此前四边界未知现已区分，不能继续解释成空图。
- 源码只在异步close监听中释放body锁；HTML Standard dialog closing算法先同步移除modal与恢复焦点，再排队close事件，见官方 https://html.spec.whatwg.org/multipage/interactive-elements.html#close-the-dialog 。因此产品显式关闭路径存在同步所有权缺口，与原实测边界吻合。零事件本身不证明后续事件不会投递，也不证明历史空图原因；真实修复效果仍需宿主复验。

## 最小实现边界
- 新增 `closeModal(dialog)`：先原生dialog.close，再调用既有releaseModalScroll。详情/编辑关闭按钮、写入完成、显式秘密转入调用页、详情转编辑使用同一入口。invalidate保留原关闭两窗后同步释放路径；原native close监听继续负责取消后的兜底释放与秘密清理，不移除任何原语义。
- releaseModalScroll仍幂等且有任意native modal存在就不释放；不手动改焦点，不修改CSS、API、后端、schema、Token/费用、授权、安全掩码或冻结设计。不加sleep、RAF、assert重试、截图替换、弱化断言或依赖/权限/门禁预算调整。已知显式关闭缺口在产品所有权层修复，不在测试调用层等待掩盖。
- 新增三项实际共享helper回归：不交付close事件也立即释放/恢复坐标，迟到事件不释放重开窗且幂等；另一native modal存在时继续锁定；UI显式关闭统一入口/native close先于释放且无focus或等待替代。抽取编辑器的四个VM文件显式注入closeModal委托，保留原mutation/readback/重复提交/迟到scope/秘密转移断言。夹具不证明Chromium事件机制。

## trellis-check与真实局部证据
- `development-55/local-checks.json`包含准确命令、UTC、退出码及日志SHA：134相关Node、JS语法、git diff检查、format、vet、build、模块校验与verification wiring race全部退出0。新增测试首次正则多一个右括号导致失败；原脱敏日志保留 `affected-first-failure.log`，只改测试正则后复验134通过，未当环境限制。首次命令未采UTC，明确null而不补造。
- 源码及调用点检查：复用同一稳定锁概念，所有显式产品关闭路径接入，invalidate有原同步释放；没有跨层数据流、公共接口或模块边界变更。原visual原生所有权/root解除/body解除/精确焦点四断言、一次fullPage、document/像素、安全与全部旧业务门禁未改。Trellis合同同步；本轮只改workspace.js、四个既有VM测试及合同/本Run公共证据。
- 原生监听限制与执行环境未变，沿用原失败证据，不重复已知必败完整make verify，不手装缺失锁定Playwright；完整wildcard Node/真实Go HTTP/浏览器交正式宿主固定入口。静态与局部绿不是完整固定门禁、self_hosted或独立QA通过。
- 新154件候选 `dc667be6f0f5db892748ac8a214d799264c2d451004da8c80a5e06e6c38f338c`，Go1.25.6。完整戳见 `development-55/source-final.json`，产物一致性见 `development-55/final-integrity.json`。

## 交接与状态
T001 In Review、M01 Needs Revision、QA14 No-Go保持；无用户必答问题或权限扩展。公共交付审查继续单一任务/里程碑/验证事实，关键路径为tests完整验证→独立QA→已授权Connector发布。评审结论Ready for Dispatch仅指可分发宿主验证，不指产品Accepted。

交tests原冷make verify，预计276 Node（原273加三项），全部Go race/真实HTTP/预算UTC迁移、权限撤销预算取消错误恢复、三视口J01–J07/summary/visual及self_hosted首次页面写入、不可达/核验失败纠正、普通/SSE同ID日志费用必须完整。保留十二分钟总门禁、900秒外层、parallel4、Playwright1.59.1/Chromium147与原权限；不得拼旧绿。失败先依据modal_close原精确边界、被动事件和原PNG/sidecar归因，不做循环猜修。完整绿后由独立QA核design、QA58-01、安全截图与self_hosted覆盖。

旧空图/R28-04精确历史原因、动态原生独立浏览器、真实供应商/支付/身份、Owner preview已登录实操仍分别未闭合；本轮不宣称根治这些问题。受控HTTP仅证明受控自托管协议路径。
