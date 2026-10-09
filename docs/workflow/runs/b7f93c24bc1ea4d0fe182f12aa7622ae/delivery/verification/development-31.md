# 执行31：修复已证输入完成边界，空图根因仍待宿主核验

## 恢复与变更边界
- 同Run T001 / M01及执行29交接继续有效；HEAD仍为main基线 `3f9cee1615e87e002a2c2300f2d29778072ec3f5`，未提交、推送或发布。本节点不重建需求或设计，不把局部绿色当交付。
- material版本 `design-v0.2.0`，archive SHA256 `5f51d35fdc15ed7214749f9413879b4238097f1d6b6bbe2cc0d508d8e103cef7`。读取 `.workflow-input/b7f93c24bc1ea4d0fe182f12aa7622ae/5f51d35fdc15ed7214749f9413879b4238097f1d6b6bbe2cc0d508d8e103cef7/files/manifest.json`，继续引用同root的 `docs/prd.md`、`docs/interactions.md`、`docs/acceptance.md`；材料状态不改为批准。引用文件SHA与manifest复核见 `development-31/final-integrity.json`。
- trellis-before-dev恢复实际frontend/backend索引、DOM/lifecycle、workspace、verification/dependency及共享思考规范。现有Trellis初始化文件与源码规范均存在，未执行初始化或bootstrap；没有空模板缺口。本轮按trellis-update-spec维护既有workspace合同，不另建事实源。
- 最小缺口在 `tests/browser/readability.mjs`：键盘已到表格右缘时，末列几何predicate在wheel前就已成立；测试仍在wheel实际投递前进入截图。仅修改该驱动、其单元回归、既有规范和本Run交付/证据。没有产品JS/CSS、后台、HTTP/API、SQLite、账号、费用/Token、权限、依赖锁、平台或部署修改。

## 执行30原始失败与已证/待证事实
- 原固定冷 `make verify` UTC `2026-10-09T04:40:43.016107Z` 至 `2026-10-09T04:51:23.768920Z`，Connector退出2，runner退出1、gate退出2。原前后154件候选均为 `b446c17fa141aeb559558c1bbe1ac230f797000108ea5d1920d98dc88d5ccfc9`；clean_start_measured/source_unchanged/candidate_unchanged/owned_group_absent均true。
- 250Node与原Go race/HTTP/预算UTC/迁移等前置门禁通过；原green_fixture此次通过，不证明执行28超时原因消失。browser migration/legacy/routing/console阶段通过；full-console在1280 visual-keys失败。1280 J01–J07及summary已运行通过，当前完整三视口visual与self_hosted未完成，self_hosted_progress为空。不得借执行13的历史结果补本次覆盖。
- 91件原runner文件逐字节归档并核SHA，见 `host-tests-30/archive-index.json`。完整runner日志505303bytes，SHA `5c9a8fe9ef85d558fbd025f65b99929dd3b716f71288138b1f1597ea9e38975c`；Connector包装505485bytes是不同对象。读取本地保留全文归档，不依赖会话首尾截断输出。
- 原 `host-tests-30/browser/visual-1280/keys-right-columns-1.png` 已实际查看：1280×1149整幅单色RGB243/246/250，四锚点均无墨。SHA `78a7b304df77dac78964e27f6571f81f2ac48376e545786f9778a183028bff53`。原integrity sidecar/DOM可见和4树13paint观测不能使原图合格。
- 同一performance时间轴：keyboard scrollend161ms，wheel-dispatch-start169ms，capture-start183ms，可信wheel189ms（表格内、坐标753/112、deltaX100000/deltaY0、观察时未prevent），capture-end272ms。wheel之前scrollLeft33+client988=scrollWidth1021，末列已在clip内。由此已证旧驱动不能保证wheel被观察后才开始capture回调；时间标记包围整个回调，不等于精确CDP栅格命令起止。
- 此次排除了“只因旧region listener漏记”的证据缺口；没有证明Chromium事件循环、默认动作或合成栅格怎样导致空图。微任务中的prevented只代表观察时刻，不代表所有后续监听均未取消。白名单摘要 `development-31/host-failure-summary.json`，原QA14/QA58、历史空图、执行28超时和R28-04因果均未闭合。

## 最小修复及严格回归
- `startWheelDispatchCompletion(region)`在原唯一wheel前注册ownerDocument被动capture listener，仅接受可信、属于该region、精确像素delta的事件，再在微任务中给出布尔确认；观察到prevented则失败。2000ms内未收到匹配事件也失败，cancel/timeout/确认移除listener与timer，driver所有退出都dispose。只证明该输入被DOM观察，不声称默认动作、paint或供应商事实已完成。
- 驱动先要求 `native-wheel-dispatch` 确认，再执行原2000ms末列wait、精确bounds、唯一fullPage PNG及全部几何/像素/secret gates。保留原ArrowRight、keyboard-scrollend、mouse.move、唯一100000wheel、末列/标识键盘访问、失败身份与scroll trace。没有sleep、RAF、重试、重拍、超时扩大或替代滚动；原入口与锁定Playwright1.59.1/Chromium设置不改。
- 新两项Node测试含可信/合成/区域外/错误delta、观察时prevented、timeout/cancel、预先到边缘但仍阻止capture直到确认、失败零capture及原dispatch异常身份和一次清理。原rightmost timeout与capture失败保留测试继续通过。VM证明驱动契约，不证明浏览器机制；新用例仍由原Node wildcard进入make verify。

## 本地验证与trellis-check
- 首轮68相关Node、runner/wiring race、静态与非监听普通/SSE Go通过，见 `development-31/initial-checks/local-checks.json`；随后仅澄清规范不得把微任务解释为完整传播/default-action完成，首轮记录原样保留，不冒充最终源码戳。
- 最终UTC `2026-10-09T05:00:22.640Z` 至 `2026-10-09T05:00:39.502Z`：68相关Node、完整tests/verify race（原8秒green_fixture/取消/SIGTERM等）、tests/verification race、format-check/vet/build、普通/SSE非监听Go race、syntax/diff全部退出0。命令/UTC/退出/日志见 `development-31/local-checks.json`。使用原Makefile工作区Go cache/toolchain/CGO；不是冷宿主完整门禁。空stdout的syntax-diff.log不列交接产物，真实退出保存在JSON。
- 最终154件源前后同SHA `6151725283c508e9f5ec6f8cf3419489fc1274ea954e787037cc6511d0b7a7a3`，见source.json/source-after.json。最终审计核归档91件SHA、runner完整日志、当前源码与material引用一致；见final-integrity.json。
- trellis-check核本轮源码范围、事件所有权、promise/timer/listener和handle清理、失败原身份、原门禁/输入/像素/安全断言及规范同步。无跨产品HTTP/存储层变化，不引入框架/依赖或新管理对象。稳定原生监听限制未改变，沿用 `native-verify-state.json` / `native-verify.log` 首败，不重复必败全入口；当前实际浏览器验留给已配置宿主tests。

## 交接与门槛
- T001 In Review，M01 Needs Revision，QA14 No-Go仍未解除。requirements/design跳过：没有新的产品边界或冻结设计偏差。
- tests运行原固定冷 `make verify`：新源码戳、无预装依赖/缓存起点、锁定版本、全日志UTC/退出/清理、全部原Node/Go race/HTTP/预算UTC迁移、1280/1440/390 J01–J07/summary/visual及独立新安装self_hosted网页分支。网页写入、不可达/核验失败→修正恢复、普通/SSE→同请求日志/费用仍须真实执行，不能用API/SQL替代。
- 若仍同处空图，先读新的 `native-wheel-dispatch` 结果、同钟wheel/capture顺序、原PNG/sidecar；确认实际输入观察已在capture前后再诊断剩余栅格边界，不轮换猜测性产品修复或完整重复求绿。若先fixture失败，读原阶段UTC/budget_ms/execute_ms/group_check_ms定位；不扩时限。绿色也不证明历史根因，后续独立QA核原材料与实际安全图片。
- 受控HTTP自托管证明与真实供应商联调分开；真实供应商/支付/身份、动态原生独立QA、Owner preview已登录操作继续分别Not Run。发布、新PR、精确同源合并及预览部署仍由已授权Connector核证完成，本节点不操作。
