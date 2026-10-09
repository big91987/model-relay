# 执行43：总门禁超时归因与隔离测试调度

## 接续、输入和版本
- 原Run、HEAD `3f9cee1615e87e002a2c2300f2d29778072ec3f5`、工作分支及全部权限/预算保持；本轮没有供应商调用、网页业务写入、依赖安装、数据库/部署/分支/提交/推送操作。执行42已结束，runner回执记录owned process group已不存在、候选已清理、源码前后一致。不重放不确定操作。原生进程枚举被拒绝，不能声称本轮独立枚举了所有进程；这不推翻runner的owned清理回执。
- 冻结材料仍为design-v0.2.0 / frozen-for-test-implementation，archive SHA256 `5f51d35fdc15ed7214749f9413879b4238097f1d6b6bbe2cc0d508d8e103cef7`；manifest及实际引用 `docs/prd.md`、`docs/interactions.md`、`docs/acceptance.md` 的相对路径、摘要和本轮核验见 `development-43/final-integrity.json`。材料仅作事实输入，不执行包内脚本，不改为批准。
- 恢复同Run总览、T001、M01、执行41记录及执行42实际文件。使用研发交付Skill及两个公共契约references；trellis-before-dev确认单Go模块，读取实际frontend/backend索引、适用源码合同及shared guides。现有Trellis scripts和源码支持规范存在，没有重新初始化或把prepare的initialized=false描述为本轮初始化通过；无需bootstrap空白规范。
- 接续版本核对采用只读重建：三个测试文件还原至HEAD字节，移除本轮新增规范段落后，其余现有文件不变，按原source-evidence清单重算应与执行41/42的154件候选 `d041dcdacbc5375fd5adb99bde3a917418b249ad69f0f64b276a2527ad578ab6` 一致。不是在工作树回滚或重新执行旧操作。当前候选154件 `6be11e323f52fc3bf47f20397ec84786f725c9bccf401cacebb797cc82e1d11f`，局部检查前后相同，见 `development-43/source-final.json`。

## 执行42已证事实与待证范围
- 完整runner日志576152 bytes / SHA256 `fdb3367c2e0e53321fd3cf86a10a03c428d33a871b6c12065399d4773dc99db9`，runner exit1、gate exit-1、外层make exit2；执行时间 `2026-10-09T07:49:21.873088Z` 至 `2026-10-09T08:01:30.131956Z`。242件原字节日志/状态/JSON/PNG/sidecar保留于 `host-tests-42/archive-index.json`，逐件摘要，不覆盖执行40或更早失败。
- 原260 Node与Go race/实际HTTP、迁移/UTC/预算等前置门禁通过；relay race包实际320.136秒。browser从07:56:06.971448Z开始；migration14312ms、legacy-ui96808ms、routing128419ms、console15299ms均结束成功；full-console从08:00:21.953867Z开始，但没有成功结束标记。
- 1280与1440分别完成J01–J07/current-period/management-summary、10页visual和12步首次self_hosted；真实Chromium版本147.0.7727.15。对应review及self-hosted独立新安装收据在归档browser目录。两种self_hosted都是受控HTTP，不能宣称真实供应商。1280原keys-right-columns截图本轮SHA `df43d194f64a4b67888fba562aefd9f1177e3065cf6b19db0206f49305098591`、1280x1149、非单色且像素门禁通过，本轮已实际查看图像有内容；1440不存在该right-columns文件，按该视口实际review/keys-page核验，不能猜造同名产物。
- 390完成J01/J02/J03/J04/J05/J07，J06-close-correction/personnel-login期间总门禁十二分钟到期，make/browser收到SIGTERM；快照 unavailable 是中止后的状态，不是人员登录产品缺陷证据。390后续summary/visual/self_hosted没有完成，不写Pass。源码/候选未变、清理完成，不属于原生监听限制。
- 已证：总预算耗尽；三类前置browser耗时及Go race包耗时已可区分。待证：冷门禁在当前候选能否完整结束；历史R28-04/QA14空图的唯一根因没有闭合，新绿不追认历史根治。QA14 No-Go、M01 Needs Revision保持。

## 最小实现边界
- 缺口位于测试调度，不在管理控制台的人员登录或后台授权。本轮产品CSS/JS、API/存储/费用/Token/安全契约均不再修改，不添加任何生产快捷路径。
- 仅在 `internal/relay/relay_test.go`（7项）、`console_test.go`（4项）、`routing_test.go`（2项）增加13行`t.Parallel()`。执行42这些父测试串行耗时合计54.23秒；名单与原时间见 `development-43/failure-summary.json`。这是缩减独立串行工作的测量依据，不能据此保证实际节省多少时间。
- 已逐项核fixture：`installation`使用t.TempDir和单独SQLite/key/lock；`newFixture`使用独占临时安装、临时端口及Cleanup；`newConsoleFixture`独占安装。Server.New的session/passwordSlots/client.Transport/routing状态是实例所有。没有修改默认HTTP客户端/transport、进程环境或全局时钟；共享一个fixture的子测试/步骤保持串行。
- 原make verify及Go race入口的`-parallel=4`保持。登录哈希成本、SSE/取消/存储失败/配置竞争断言、channel barrier、截止时间、cleanup、原wheel/keyboard-scrollend/单次截图/像素/秘密门禁均不变，不加sleep、重试、扩预算、隐藏列或测试skip。Trellis update-spec仅把本轮可执行合同接入既有dependency-preparation场景。

## Trellis检查与本地结果
- 9条检查均有UTC、exit/signal、日志字节/SHA及范围，见 `development-43/local-checks.json`。8个新增并发的非监听Go场景在原race和parallel4下全部通过；没有把另外5个实际监听HTTP/SSE场景当本地通过。108相关Node通过；runner、依赖顺序wiring及受控普通/SSE非监听合同race通过；format-check/vet/build/go mod verify、JS语法/git diff检查通过。
- 局部race是隔离与原断言检查，不是完整冷门禁性能证明。没有本轮局部失败；链接器临时绝对object路径已脱敏，保留raw/safe字节数及事实。syntax-diff成功日志为0字节，不能以非空要求误判失败；交接只列非空实际产物。
- trellis-check核对13行调度变更外原测试正文精确不变、Makefile/Go及npm锁与HEAD一致；无新依赖、框架、注释、生产调试输出或数据流变更。当前154件源码戳前后相同。历史稳定原生监听限制仍见development-check及native-verify-state/log，环境未变，不重复必败原生完整命令；完整make verify明确留tests宿主，绝不写成本地Pass。

## 交接硬门槛
交tests从当前候选执行原固定冷 `make verify`，不手装依赖、不改变Connector权限/十二分钟门禁/九百秒外层预算/Playwright和Chromium锁；保留完整日志、UTC、退出码、冷缓存状态、源戳及owned清理。必须完整执行全部原Node/Go race/实际HTTP、UTC预算迁移、三视口J01–J07/summary/visual与self_hosted网页普通/SSE/同ID日志费用/不可达和核验失败恢复，不能拼接执行42局部绿放行。

若仍超时，先用本轮对照relay包wall time与browser阶段UTC区分节省未达标、冷编译成本或新的网页停顿；若原空图再次出现，检查实际PNG/边界策略/主frame/loss与marker白名单再修，不循环猜CSS、增加sleep或放弱断言。完整通过之后独立QA核源码、同源收据、实际安全截图及QA58-01，QA14是否闭合由QA判断；正式供应商/支付/身份、独立动态原生浏览器及Owner preview已登录操作分别仍Not Run。
