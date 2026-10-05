# 宿主测试 seq 13 独立复核索引

日期：2026-10-05。原始完整回执由平台 previous_results seq 13 保存；本文件是独立 QA 的摘录与索引，不冒充重新执行的宿主测试或完整 stdout。执行者为已配置项目测试 Connector，命令 `make verify`，真实退出码 **0**。

## 版本摘录

```json
{"files":31,"go":"go1.25.6","head":"fdcb16cb3426d403327da727ae2a8f95c001f1af","scope":"product + tests + build locks + README + acceptance/design; excludes delivery receipts to avoid self-reference","working_tree_sha256":"97efa7f179cbc32edd3b99e59ee2b9902f12fce46866b9d3847a196f7abcef38"}
```

QA 本轮重新执行 `go run ./tests/evidence`，获得完全一致的 HEAD、31 文件边界与摘要，见 `qa-sandbox-checks.log`。检查摘要生成器覆盖产品、网页、Go/浏览器测试、构建入口、依赖锁及本次需求/设计；排除交付/QA 报告及运行缓存，避免自引用。不是仅凭 HEAD 接受未提交代码。

## 真实执行结果摘录

- format-check 成功；`go mod verify` 输出 `all modules verified`；`go vet ./...` 成功。
- `node --test tests/browser/ui.test.mjs`：4 项通过，fail/cancelled/skipped/todo 均为 0。
- `go test -race -count=1 -v ./...`：CLI 包成功；relay 包成功（95.373s）；browser/evidence 为测试运行器，无 Go test 文件，不算额外行为测试。
- CLI 用例 `TestCLIInstallationCommands` 通过。
- relay 用例 `TestSpecialInstallationPath`、`TestHeaderBoundaryAndFailureContract`、`TestProcessCrashRecovery`、`TestInProcessContract`、`TestHTTPRestartAndBackup`、`TestInterruptedActualRequest`、`TestUsageSnapshots`、`TestInstallationRecovery`、`TestInterruptedAndVersion`、`TestSourceAndLoginLimits`、`TestDestinationAndProtocol`、`TestHTTPManagementAndOrdinary`、`TestTransparentParametersAndLocalReject`、`TestSSEBarrierAndUsage`、`TestCancellation`、`TestUpstreamFailures`、`TestTimeoutRedirectAndNetwork`、`TestStorageFailClosed` 全部通过，包括输出中的子用例。
- `TestCrashProcessHelper` 是子进程辅助入口，单独 PASS 不代表另一个故障验收；实际故障证据来自 `TestProcessCrashRecovery`。
- `go build -trimpath -o bin/model-relay ./cmd/model-relay` 成功。
- 浏览器阶段实际 `npm ci` 后运行 `go run ./tests/browser`，启动真实 Go 服务与显式隔离受控上游，不是静态 browser.check，也不是假响应 route.fulfill。

浏览器原输出：

```text
PASS actual Go service browser journey 1280px (controlled upstream, not supplier validation)
PASS actual Go service browser journey 390px (controlled upstream, not supplier validation)
```

宿主日志的链接器 LC_DYSYMTAB warning 未导致失败；`storage_write_failed` 出现在有意注入的存储失败用例。不能把它们归为当前产品断言失败，也不能据此忽略测试覆盖缺口。

## 证据边界

该回执真实证明列出的测试在该摘要下通过。逐项验收还需复核测试断言本身；本轮发现的 AC-04/AC-15 缺口详见 `qa.md`。不把该出口 0 改写为所有原 AC 已覆盖。不附浏览器截图/trace：现有运行器未产生此类文件，本次使用实际旅程回执及对应源码审核，未编造截图。

真实供应商联调 **Not Run**：没有供应商凭据、真实响应或付费调用授权。
