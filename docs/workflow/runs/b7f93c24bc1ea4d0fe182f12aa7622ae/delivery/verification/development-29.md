# 执行29：夹具前置超时诊断，浏览器归因仍待实测

## 恢复与实施边界
- 同Run已有README/T001/M01、执行27交接、执行26空图及QA14 No-Go继续为事实源；HEAD仍为 `3f9cee1615e87e002a2c2300f2d29778072ec3f5`，不重启Run或重建方案。
- 材料design-v0.2.0，archive SHA256 `5f51d35fdc15ed7214749f9413879b4238097f1d6b6bbe2cc0d508d8e103cef7`；读取 `.workflow-input/b7f93c24bc1ea4d0fe182f12aa7622ae/5f51d35fdc15ed7214749f9413879b4238097f1d6b6bbe2cc0d508d8e103cef7/files/manifest.json`，沿用同root `docs/prd.md`、`docs/interactions.md`、`docs/acceptance.md` 的网页写入、秘密与同请求费用契约。实际引用SHA复核见 `development-29/final-integrity.json`；材料frozen-for-test-implementation不改为批准，不执行包内脚本。
- 研发交付Skill及两份公共契约references用于续接现有任务；trellis-before-dev实际执行get_context识别single-repo，读取backend/frontend/guides索引及相关runtime、dependency、verification、DOM/lifecycle、workspace与reuse/cross-layer规则。现有初始化脚本及有源码依据的规范存在，bootstrap不适用；prepare initialized=false不解释为本轮缺少初始化。trellis-update-spec维护原dependency-preparation。
- 编码前界定：缺口是宿主无辅助阶段耗时，不能区分原8秒夹具预算耗尽与某阶段执行/进程组检查拖延。责任在项目测试入口captureCommand，不在产品UI或Platform控制器。仅修改 `tests/verify/main.go`、`tests/verify/main_test.go`、相关Trellis规范及同Run交付记录。
- 无产品/API/DB、账号、依赖、权限、Connector、部署或Git发布操作；没有sleep、raf、重试、截止延长或放弱断言。原页级wheel诊断和所有浏览器/像素门禁保持。

## 执行28已证与未知
- 原固定冷入口make verify：UTC `2026-10-09T04:21:23.935954Z` 至 `2026-10-09T04:27:39.074261Z`；Connector exit2、gate_exit2、runner exit1。source/candidate前后不变、cold测量与owned_group_absent=true、candidate_removed保留在 `host-tests-28/state.json`。
- 250Node通过后，原 `TestFixedRunnerReceipts/green_fixture` 失败：耗时8.01秒，prerequisite-npm cancelled=true、context deadline exceeded。此npm是只有echo的隔离shell夹具，不是真实npm、产品、供应商或监听失败证据。
- 宿主未到browser，没有新PNG/wheel事实；执行26原全空白PNG与页级输入归属未知继续保留。不把Go夹具失败当空图消失，也不把旧1280成功旅程移记到执行28。
- 2件原state/verify.log逐字节复制、逐件SHA核对，索引 `host-tests-28/archive-index.json`。完整runner日志423684bytes，SHA `2d677abf830b07f1cfea30c4b4e8911382d34f13b2b893027a9f5a297d54ec63`；Connector包装423866bytes是不同对象。调用read_command_output(seq28)确认日志可读，未将首段当全文；归档完整原runner文件，不依赖截断显示。
- 原前后154件源码戳均为 `0211a610cb72b4435455a6d746d7648e9668d414a967d021c1b645ff97825e96`。白名单摘要 `development-29/host-failure-summary.json` 区分已证与未知。
- 修改前本地单次 `go test -race ./tests/verify -run '^TestFixedRunnerReceipts/green_fixture$' -count=1 -v` 通过，green_fixture为2.14秒；仅工具回执保留，未另存全文，不当完整证据或宿主根因。局部绿色不能解释宿主超时，故不做猜测修复或扩大时限。

## 原入口内最小诊断
- 每个辅助command receipt新增started_utc、budget_ms、execute_ms、group_check_ms，日志同样只增加这些字段。UTC定位阶段，单调时钟测耗时；采样原上下文预算，取消/到期钳为零。
- execute_ms包含原Start/Wait/kill，group_check_ms只测其后ownedGroupAbsent；不能区分内核启动与shell调度。没有新命令、等待、执行参数、PID、环境或路径投影。
- 原execute、TERM/5秒WaitDelay/KILL、group检查、8秒夹具/2分钟辅助/12分钟gate及原失败身份/退出和业务断言均保持。意外夹具结果先读已终结receipt，仅输出上述字段与原stage/exit/cancelled再按原期望失败。
- 新TestCaptureCommandTiming用实际隔离shell检查成功/非零/预取消的时间格式、预算、stdout/stderr分离、退出与清理，加入原go test通配入口；不是Chromium或供应商机制证据。

## 本地回归与trellis-check
- 首次新测试传裸bytes.Buffer给stdout/stderr并发copy，race与capture断言真实失败，不能归入环境。保留 `development-29/go-runner-first-failure.log`、`local-checks-first.json`、`first-failure-summary.json`；日志是显式脱敏的本地派生对象，不冒称宿主原字节。
- 根因在新增夹具，真实run原本使用mutex保护的cleanLog。测试改用原cleanLog并flush后读，不改execute或降低race要求。first metadata的log名称为首次采集时名称，只有失败runner日志另存；其他原stdout未另存，不用后来文件反推旧证据。
- 修正后UTC `2026-10-09T04:34:32.282Z` 至 `2026-10-09T04:34:46.522Z`：完整tests/verify race（含原取消/SIGTERM/npm配置/fixture与3项新分支）、tests/verification race门禁顺序/fail-fast、66相关Node、format-check/vet/build、普通/SSE非监听Go race、syntax/diff全部通过。实际命令/UTC/退出/日志见 `development-29/local-checks.json`。
- 最终154件源前后均为 `b446c17fa141aeb559558c1bbe1ac230f797000108ea5d1920d98dc88d5ccfc9`，见source.json/source-after.json。没有手装依赖，使用原Makefile工作区缓存/local toolchain/CGO；不当冷宿主完整make verify。syntax-diff.log无stdout，真实退出0在local-checks中；交接不列空文件。
- trellis-check核本轮diff、并发输出安全、旧上下文/信号/退出/cleanup、秘密字段、规范同步和范围。新增字段仅测试证据，不跨产品HTTP/存储层，无需改业务接口或引入UI框架。稳定本地监听限制未变，沿用首次native失败，不重复必败全入口；完整race/HTTP/预算UTC/迁移/浏览器交tests宿主。

## 交接与门槛
- T001 In Review、M01 Needs Revision；QA58-01/QA14、历史空图、执行28超时及R28-04因果均未闭合。requirements/design跳过：没有新增产品边界或冻结设计偏差。
- tests运行原make verify冷起点。夹具再失败先读阶段budget_ms、execute_ms、group_check_ms及UTC，区分累计预算/单阶段耗时；不足仍保留未知，不扩时/重试求绿。
- 到旧空图边界先读readability.scroll_probe页级wheel归属/坐标/阶段、原PNG及integrity sidecar；原单次截图和像素/秘密断言保持。未观测不证明没有输入，绿色也不证明历史机制根治。
- 全部原Node/Go race/HTTP/预算UTC迁移和1280/1440/390 J01–J07/summary/visual、新独立self_hosted网页完整回执后，由独立QA核原材料AC和实际安全图片。
- 受控HTTP自托管与真实供应商联调分开；真实供应商/支付/身份、动态原生独立QA与Owner preview已登录操作分别Not Run。发布/PR/合并/部署由已授权Connector负责，本节点不发布。
