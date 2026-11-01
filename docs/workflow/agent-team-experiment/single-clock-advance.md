# 单次 HTTP 请求有效时钟推进审计

日期：2026-11-01

## 契约与实现

除 `/healthz` 外，每个外部 HTTP 请求只在 `Server.ServeHTTP` 推进一次持久化有效时钟。该推进与 `server.mu` 的管理/路由事务序列化边界对齐：未竞争时直接推进；若请求已在该锁后排队，先保留一次非持久化到达采样，取得锁后再执行唯一的 `Store.advanceClock`，因此截止时间仍取排队结束后的墙钟。推进失败继续设置 degraded 并返回 `503 storage_unavailable`。

删除的重复调用：

- `Server.listModels`：必经 `ServeHTTP`，删除列表处理内第二次推进。
- `Server.identityAPI`：必经 `/admin/api/` 的 `ServeHTTP` 分派，删除身份事务前第二次推进。
- `Server.consoleAPI`：必经 `/admin/api/` 的 `ServeHTTP` 分派，删除控制台事务前第二次推进。

保留的调用：

- `Server.acceptInvitation`（`identity.go`）：保留其独立事务时点。该入口包含密码限流/哈希等事务前工作，截止校验必须紧邻其事务；当前 `ServeHTTP` 的统一推进仍是外部请求推进，内部调用用于邀请接受事务时点保证。
- `Store.admit`（`accounting.go`）：保留准入独立事务时点；它也可由不以 `ServeHTTP` 为事务边界的内部流程调用，并负责配额窗口/凭据时效的一致快照。
- `Server.beginAttempt`（`routing.go`）：保留每次路由尝试的独立时点；重试可跨越冷却或配置边界，不属于外部请求入口的重复写。
- 测试中的两个 `fixture.store.advanceClock`（`console_contract_test.go`）：显式测试布置，不是生产 HTTP 调用链。
- `Server.advanceClock` 与 `Store.advanceClock` 定义：保留统一入口包装、SQLite 高水位、防回拨、日志和故障关闭实现。

## 回归证据

新增 `internal/relay/clock_advance_regression_test.go`：

- 使用可观测测试时钟证明模型列表、身份租户列表、控制台概览三类成功请求各调用一次墙钟并只执行一次持久化推进。
- 将墙钟回拨一小时，证明内存有效时间和 `console_clock.effective` 均保持高水位，且该请求只推进一次。
- 关闭隔离测试数据库，证明推进事务启动失败时返回 503、设置 degraded，且不进入后续请求处理。
- 复跑既有 `TestConsoleClockQueuedDeadlineInProcess`，证明排队跨越密钥/邀请截止边界的语义保持。

## 实际命令与退出状态

1. `go test ./internal/relay -run 'Test(ExternalHTTPRequestAdvancesEffectiveClockOnce|SingleRequestClockAdvance)' -count=1`：exit 1，未进入编译；默认 Go cache 为只读，自动 Go 1.25 工具链首次报告标准库解析失败。此结果不记为产品失败。
2. `GOCACHE=/tmp/model-relay-go-cache GOTOOLCHAIN=local go test ...`：exit 1；本机 local Go 为 1.21.6，低于 `go.mod` 要求 1.25.0。
3. `GOCACHE=/tmp/model-relay-go-cache GOTOOLCHAIN=auto go test ./internal/relay -run 'Test(ExternalHTTPRequestAdvancesEffectiveClockOnce|SingleRequestClockAdvance)' -count=1`：exit 0。
4. `GOCACHE=/tmp/model-relay-go-cache GOTOOLCHAIN=auto go test ./internal/relay -count=1`：exit 1。测试实际运行 35.752 秒；大量 ActualHTTP/进程用例因沙箱禁止 `listen tcp 127.0.0.1:0` 失败。该轮还捕获一个真实的 in-process 排队截止回归（模型列表得到 200 而非 401），随后通过将统一入口推进与 `server.mu` 边界对齐修复。
5. `GOCACHE=/tmp/model-relay-go-cache GOTOOLCHAIN=auto go test ./internal/relay -run 'Test(ExternalHTTPRequestAdvancesEffectiveClockOnce|SingleRequestClockAdvance|ConsoleClockQueuedDeadlineInProcess)' -count=1`：修复前 exit 1（入口等待锁导致既有测试屏障超时）；加入排队时非持久化到达采样后重跑 exit 0，2.040 秒。
6. `GOCACHE=/tmp/model-relay-go-cache GOTOOLCHAIN=auto go test -race ./internal/relay -run 'Test(ExternalHTTPRequestAdvancesEffectiveClockOnce|SingleRequestClockAdvance|ConsoleClockQueuedDeadlineInProcess)' -count=1`：exit 0，4.149 秒。
7. `make verify`：exit 2；Makefile 强制 `GOTOOLCHAIN=local`，本机 Go 1.21.6 不满足 go.mod 的 1.25.0。
8. `make GOTOOLCHAIN=auto GOCACHE=/tmp/model-relay-go-cache verify`：exit 2；验证器完成 source-inventory 后因基线中已有但工作树不存在的 `docs/workflow/development-browser-diagnostics-browser.log` 执行 `statat` 失败。完整 format/vet/test/build/browser 链未运行，不能声明 `make verify` 通过。
9. `git diff --check`：exit 0。
10. attempt 3 固定 PATH 核对：`go version` 为 `go1.25.6 linux/amd64`，Go 路径为锁定 toolchain；`node --version` 为 `v22.14.0`，`npm --version` 为 `10.9.2`，exit 0。
11. attempt 3 原样 `make verify`：exit 2。隔离验证回执为 `test-results/verification/20261101T033444Z-1092849137`；source inventory、Git directory、两次 source evidence 和 Go/Node/npm prerequisites 均 exit 0，工作树摘要为 `91159542c9637fbec4c4b489b134bf01d533450472b901fa809ee1429ff6155c`。`make verify-gates` 在 `go run ./tests/review-history` 失败，原始错误为 `review historical HTTP rejected: baseline unavailable: exit status 128`；未进入 format/vet/frontend/full test/build/browser，不能声明完整门禁通过。
12. attempt 3 首次直接专项命令：exit 1，因默认 `GOCACHE=/home/moffett/.cache/go-build` 在原生文件系统中只读；这是命令环境问题，未进入测试。
13. `GOCACHE=$PWD/.cache/go-build go test -race ./internal/relay -run 'Test(ExternalHTTPRequestAdvancesEffectiveClockOnce|SingleRequestClockAdvance|ConsoleClockQueuedDeadlineInProcess)' -count=1`：exit 0，4.531 秒。
14. `GOCACHE=$PWD/.cache/go-build go test ./internal/relay -count=1`：exit 0，62.972 秒；固定测试网络下整个受影响包（含真实临时端口用例）通过。
15. attempt 4 原样 `make verify`：exit 2，隔离回执为 `test-results/verification/20261101T034451Z-3001124188`。Go 1.25.6、Node 22.14.0、npm 10.9.2、source inventory/evidence 和 Git directory 均通过，工作树摘要仍为 `91159542c9637fbec4c4b489b134bf01d533450472b901fa809ee1429ff6155c`；`go run ./tests/review-history` 再次报 `review historical HTTP rejected: baseline unavailable: exit status 128`，后续必跑阶段未执行。
16. attempt 4 只读定位：当前任务 worktree 中 `git cat-file -t 66f1328ada75ea90e5638fbf2bd58663ac86347e` 返回 `commit`，`git cat-file -e ...^{tree}` exit 0，`git rev-parse --is-shallow-repository` 返回 `false`。因此维护源恢复的对象在当前 worktree 可见；失败发生于 `tests/verify` 创建的私有隔离候选中执行 `git archive baseline`，属于隔离候选未继承该历史对象的设施/验证器问题，而不是本次产品代码或测试断言失败。
17. attempt 5 原样 `make verify`：exit 2，隔离回执为 `test-results/verification/20261101T035007Z-1808404718`，工作树摘要仍为 `91159542c9637fbec4c4b489b134bf01d533450472b901fa809ee1429ff6155c`。本轮 `review-history` 已真实通过：成功归档 `66f1328ada75ea90e5638fbf2bd58663ac86347e`，运行历史 actual HTTP/race 重建并观测到预期三个旧断言失败，archive SHA-256 为 `4b9b3e1b583d0ee0f9e17e007085389ac20414dc5b360cdb2416bb0a48e1e3b1`。随后首个失败发生在 `make health-history`：`tests/health-history` 对其基线 `9b6994ed80de2bdba79ff267419150dccaac1360` 报 `historical HTTP reproduction rejected: baseline archive unavailable: exit status 128`。因此此前对“私有候选普遍不能继承历史对象”的推断不充分；本轮证据只支持第二个历史基线仍未准备/不可见。format/vet/frontend/current full test/build/browser 仍未执行。
18. attempt 6 原样 `make verify`：exit 0，完整隔离回执为 `test-results/verification/20261101T035407Z-1882954909`。执行时工作树摘要为 `91159542c9637fbec4c4b489b134bf01d533450472b901fa809ee1429ff6155c`，结束时 source unchanged。三个历史重建门禁通过；format-check、`go mod verify`、vet 通过；前端 Node 测试 286/286 通过；`go test -race -count=1 -v ./... -parallel=4` 全部通过，其中新增模型列表、身份、控制台单次推进及回拨/存储失败回归均通过；build 通过；latency experiment 2480 样本、0 失败；真实 Go 浏览器 migration、legacy UI、routing、console、full-console 在 1280/1440/390 相关视口全部通过。dependency probe 确认 Playwright 1.59.1、playwright-core 1.59.1、Chrome for Testing 147.0.7727.15。该回执为受控测试上游证据，不是供应商联调。

## 限制

- 当前沙箱不允许 host-only 监听测试，因此整个受影响包及完整门禁没有绿色回执；专项、故障、回拨与既有排队边界均已在 in-process race 测试中真实通过。
- 首轮 `agent_team_factory.get_task` 被运行环境以“approval policy is never”拒绝；该失败证据保留。attempt 2 中工具注册授权已由维护源修复，`get_task` 成功，确认无依赖，并收到“保留原结果与验证限制、限制妨碍完成则 blocked handoff”的维护消息。
- attempt 3 的 Go/Node/临时端口设施已恢复，专项及整个 relay 包通过；但项目历史基线门禁仍在产品测试前以 exit 128 失败。该基线不在本任务允许修改路径内，且不得修改测试门禁，因此仍无法取得 `make verify` 绿色回执。
- attempt 4 已确认原 worktree 非 shallow 且目标 commit/tree 可解析，但隔离验证候选仍报告 baseline unavailable；完整门禁仍未通过。
- attempt 5 已证明 `review-history` 的第一个基线可在隔离候选中成功归档和执行；新的阻塞是 `health-history` 所需的另一个固定基线 `9b6994ed80de2bdba79ff267419150dccaac1360` 无法归档。该对象准备仍属于维护源/Git 历史设施，不在本任务写范围内。
- attempt 6 在维护源准备全部显式历史对象后，完整固定门禁通过；此前各轮失败仍保留为历史，不用于替代最终成功回执。
- 未执行真实供应商联调；本变更不涉及供应商协议或凭据。
