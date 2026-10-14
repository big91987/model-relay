# 同 Run 续接：冷却 HTTP fixture 归因及阶段边界修复

## 身份、授权及旧失败

2026-10-14，沿用 Issue42 / Run `e65c964ce9faf29e35271270d751cedb` / Workflow revision 2 的现有共享工作区；HEAD 仍为 `5ff456cb6de0ce8d45bb2faac4c964cda9efcbf0`。先读取 AGENTS.md、首轮冻结输入摘要、执行 6 诊断、既有源码和归档；development-e6/SHA256SUMS 全部核对成功。没有重建工作区或重新生成需求/设计。

用户本次正式授权先归因；仅在确认测试设施可移植性问题时最小修 fixture，保持实际 HTTP、排除/回退/冷却到期恢复断言和全部固定门禁，不能在本任务修产品行为。本轮源码新增修改只有 `internal/relay/routing_http_test.go` 中原失败测试；此前两个 review-history 源文件的授权修复原样保留。

tests 执行 5 平台原始回执由 Builder 正式读 EOF 保存：exit 2、1010274 bytes、truncated=false（依据用户正式续接消息，不声称研发本轮重新读取平台全部原始事件）。独立的项目 verify.log 是 1003960 bytes / 3844 行 / SHA-256 `bf350234304a88bc9935c8229c6f3a6c60ad413337494d1bd6c431a6a49e1bc4`，与对应 state.json 一致，保留在 `harness-output/development-e6/tests-seq5-verify.log`。两种长度对应不同日志层，不能混同。项目回执 gate_exit=2、exit_code=1；历史 JSON 修复已越过，真实当前失败为 `TestRoutingRealHTTPSelectionFailoverAndRecovery` 的冷却排除断言。

原 tests5 没有故障请求的冷却时间戳；本轮不回填或声称已证实其具体超窗耗时。执行 3 原失败、执行 4 验证器红绿和执行 5 原失败均保持真实失败，不升级为通过。

## 最小归因实验：合法到期也能触发旧断言

源码核对：routing.go 的 beginAttempt 用 time.Now 与 health.Cooldown.After(now) 判断冷却，observe 用 time.Now 加 Config.Cooldown 生成截止；现有 Config.Clock / Store.clock 控制的是其他持久化时钟，不能控制这两处。为接入路由可控时钟而改产品源码不属于本轮范围。

在原 HTTP 测试上临时加入仅用于诊断的受控备用响应耗时：由实际 primary 503 产生原 15 毫秒截止，在 secondary HTTP handler 中用 timer 等待该截止后 1 毫秒（附 request cancellation 边界），随后仍以真实 HTTP 获取下一请求选路记录。没有篡改数据或供应商记录。诊断 timer 只用于构造反例，不是延迟求绿方案，最终交付源码不含此 timer。

真实命令 `go test -race -count=1 -v ./internal/relay -run '^TestRoutingRealHTTPSelectionFailoverAndRecovery$'` 退出 1；原冷却排除断言仍失败。实际管理 API 记录显示原截止 `2026-10-14T12:46:50.634931478Z`，下一请求 primary 选择时刻 `2026-10-14T12:46:50.650582133Z`、reason=selected、primary_delta=1：已超截止约 15.65 毫秒，选中 primary 是合法到期行为；后续 primary 再次 503 才建立新的 cooling，而非原截止仍有效。

实际诊断源码及完整输出保存在 `harness-output/cooldown-fixture/routing-http-diagnostic.go`、`elapsed-diagnostic-red.log`、`diagnostic-status.txt`；原源码保存在 `routing-http-before.go`。该构造实验确证旧测试把“下一请求”错误等同于“尚未到期”，依赖 HTTP/持久化/调度速度；它不是 tests5 当时负载的精确重放，不证明 tests5 的唯一具体原因，也未发现需修的产品冷却错行为。

## 最小 fixture 修复：独立建立未到期与已到期阶段

改动仅限原测试，未接入或改变产品时钟：

- 保留 40 个并发真实 HTTP 请求和 SWRR primary/secondary=30/10，以及真实 primary 503 到 secondary 的两次尝试、映射、凭据哨兵及计量断言。
- fixture 配置 Cooldown 为一小时，使“未到期阶段”不再依赖 15 毫秒的调度窗口；不等待一小时、不增加 sleep/retry。核验真实故障生成的 Cooldown-Observed 必须精确等于该配置，不能用长窗口隐藏错误期限。
- 未到期请求必须只有 secondary 一次尝试、primary 计数不变，并从真实管理 API 核验 primary 决策 reason=cooling、选择时间早于真实截止、返回截止完全一致。不是只检查 HTTP 200。
- 在无在途请求的测试阶段边界，以 server.mu 保护并校验同一 primary 健康实例未被冷却请求替换，再将测试拥有的内存 health.Cooldown 调到其 Observed 前 1 纳秒。只调整 fixture 的内存截止，不写数据库、历史请求、尝试或决策记录；不改产品源码，也不伪装成产品自然时间经过。
- 已到期阶段仍执行真实 HTTP：必须只有 primary 一次成功尝试、primary 计数恰好增加 1，管理 API 的选择时刻不得早于阶段截止、reason=selected 且没有 cooling 标记。原 25 毫秒 sleep 被移除。

这里明确以测试阶段截止模拟“已经到期”的输入状态，验证真实选择/回退/恢复和审计输出；自然真实时间到期可使 primary 再入选的反例由前述诊断另行观察。独立 QA 必须核对这一区分、内存阶段调整以及所有保留/新增断言，不能将阶段调整误报为可控产品时钟或自然等待验收。

## 真实定向结果及非通过结果

- 修后同一精确 HTTP 测试 `go test -race -count=1 -v ...` 真实退出 0，见 `phase-green.log` / `targeted-status.txt`。
- 在测试拥有的新临时候选中复制本版 internal 和锁定模块，不改变共享工作区产品源文件；分别注入三种单点产品反例，每次独立执行同一真实 HTTP 测试且仅一次：忽略冷却、永不到期、生成期限减半。三次实际退出均为 1，分别因 secondary 不独占、到期 primary 不恢复、实际期限 30 分钟被拒绝。实际补丁、完整输出和退出码分别归档为对应 `*.patch`、`*-red.log`、`mutation-status.txt`。候选目录已清理；这些隔离故障变体不是交付的产品代码。
- `go vet ./internal/relay` 真实退出 0；gofmt 执行，git diff --check 无错误。工作区 routing.go、server.go 与基线无差异。
- 扩展命令 `timeout 120 go test -race -count=1 -v ./internal/relay -run '^TestRoutingRealHTTP' -parallel=4` 真实退出 1，日志完整保留为 `routing-http-incomplete.log`，不得称整个路由 HTTP 集合通过。它先通过 committed tools 边界与修后 selection 测试，随后在 temporary allowlist 并行子例中 `signal: killed`，包耗时 14.569 秒；未出现该测试集合的断言 FAIL 输出。此本地扩展命令与 vet 曾重叠，不是固定 make verify 回执。
- 扩展命令结束后容器只读 memory 快照显示 memory.max=2147483648、memory.peak=2147495936、累计 oom=1/oom_kill=1，见 `memory-snapshot.txt`。没有运行前计数差或进程关联，因此不宣称已精确归属该信号原因。完整固定测试仍待独立 tests；不重试本地扩展集、不扩资源、不删并行门禁。

隔离 mutation 命令最初因执行器禁止 rm 风格清理而在执行前被拒绝，没有产生测试副作用；随后改用仅测试新建临时目录的安全清理方式执行上述三个独立反例，不将拒绝当作测试运行或成功。

## 新版本、交接与未验

项目 evidence 入口实际退出 0，新 working_tree_sha256=`a6f4ec9a2a4e9f87f501aee7f48251ccede244d46082fc09b415ea9953a03b3e`、158 文件，HEAD 不变；原修复摘要 `12d91213...de70` 属于此前版本，不能混用于新 QA。完整新源证据、文件摘要和修后源码归档于 `harness-output/cooldown-fixture/source-evidence.json`、`SHA256SUMS`、`routing-http-after.go`。

说明首轮 SHA-256 仍为 `c0ec7667d07fb4eccbd346f69a64ebff141178a9609015dc71a5dec56818504b`，授权缺章继续保留，尚无独立 QA 正式拒绝，不能提前补章。凭据只核对 GITHUB_TOKEN/GH_TOKEN/GIT_ASKPASS 存在性，均 false；未读取值。未修改 Makefile、锁文件、产品业务、通用平台、冻结准备或数据库；未执行提交、push、PR、发布、合并、真实供应商调用。

以本轮实质 fixture 修复及真实归因/正反例为依据，原生 handoff 交固定 tests 对新摘要执行原 make verify，保留完整原始回执、所有门禁与期限，并单独报告本地扩展 signal failure。若固定失败，先读完整回执定位，不重复求同版绿例。固定完整通过后仍需独立 QA 核对两次范围补充及版本、正式拒绝首轮缺章并 handoff 返工、修正后重测/复核、正式人工确认及报告/草稿 PR。研发不写 qa.md/report.json，不将流程或隔离 upstream 验证当作真实供应商验收。
