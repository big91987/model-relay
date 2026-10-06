# 宿主固定门禁证据索引

## 真实来源和版本

来源：平台 tests Connector seq16，命令 `make verify`，预算900秒，实际 exit_code=0。保留日志74864 bytes，truncated=false。QA 使用平台 `read_command_output` 检索 seq16：offset 0、32768、36000、39000、46500、62000，检索到 eof（next_offset=74864）。本文件是重点摘录与审查索引，**不是完整原始日志，也不是 QA 自己执行 make verify**；工具显示长段时存在展示裁剪，完整档案以平台 retained log 为准。

日志开头身份：Go go1.25.6；HEAD `85f6d17c07c5fddeb2d6a39a5a067769a11bb4cc`；82文件 working_tree_sha256 `c340664b231d83d153f0d2569065db08c8be34661a314f51e7831f74f808b27e`。QA 开始独立执行原 evidence 命令得到同一值；添加 QA 测试后另独立重构原清单、排除唯一新测试再次得到同一值，见 version-lock-01.json。

当前受管理 Workflow sha256 `a3d814fa47da0c74e6934ab46138f2b854c55cf1de177f7fba538b4365ab2f45` 与 manifest一致。其安装来源见研发 `platform-template-provenance-01.json`，固定平台源294ddb33；产品08摘要不覆盖Workflow/manifest，必须合读。QA没有复跑平台安装器或确认正式部署控制器升级。

## 重点实际输出

```text
PASS expected historical HTTP red: 21 method failures with exact status/Allow diagnostics; 12 query/recovery passes
all modules verified
go vet ./...
tests 116 / pass 116 / fail 0 / skipped 0
go test -race -count=1 -v ./...
PASS TestDeploymentContractAndIdempotentUpgradeCLI
PASS TestCLIManagementConfigurationContract
PASS TestCLIInstallationCommands
ok model-relay/cmd/model-relay 8.920s
ok model-relay/internal/relay 224.934s
ok model-relay/tests/health-history 5.052s
go build -trimpath -o bin/model-relay ./cmd/model-relay
go run ./tests/browser
PASS exact legacy UI migration baseline=85f6d17c07c5fddeb2d6a39a5a067769a11bb4cc schema=1->2 aliases=2 keys=2 historical_requests=2 controlled_posts=4; old-binary rollback retains pre-upgrade identities; not formal Actions or supplier validation
model_diagnostic_execution=75436
PASS actual Go service browser journey 1280px (controlled upstream, not supplier validation)
PASS actual Go service browser journey 390px (controlled upstream, not supplier validation)
routing_diagnostic_execution=75584
PASS real multi-upstream UI 1280px: connection GET, priority/3:1 visits, disable, failover/unknown usage, passive recovery, SSE/cancel no replay; controlled upstreams only
PASS real multi-upstream UI 390px: connection GET, priority/3:1 visits, disable, failover/unknown usage, passive recovery, SSE/cancel no replay; controlled upstreams only
```

上面 PASS 测试名行是便于查阅的归一化摘录，原日志为 Go 的 `--- PASS` 格式。历史网络红灯是隔离旧9b6994ed代码的预期负控，不是当前版本 health 回归失败；不要将其与当前核验缺陷混淆。链接器warning非测试失败。宿主全链绿色只证明已实现断言执行成功，QA新增契约负控仍真实失败。

## 历次失败保留

| seq | 产品摘要版本 | 实际结果 | 说明 |
|---|---|---|---|
| 7 | 03 / 39c1a53… | exit2 | 摘要中段缺失；研发CLI新API适配修复不能伪称完整归因 |
| 9 | 04 / 5b30123… | exit-1 | 原300秒预算中断，部分效果；不是功能通过 |
| 10 | 04 / 5b30123… | exit2，74586 bytes retained | 多候选创建1280阶段失败 |
| 12 | 05 / f251735… | exit2，74970 bytes retained | primary原生选择未改变索引 |
| 14 | 06 / 28e1a56… | exit2，77652 bytes retained | 70677原始轨迹：事件到达，原select无变更 |
| 16 | 08 / c340664… | exit0，74864 bytes retained | 当前宿主全入口通过，不替代逐AC评估 |

早期无log字段的回执不能恢复缺失日志。10/12/14完整日志仍可使用其seq分页读取；本轮未冒充逐字复核所有历史日志。研发 host-failure-01～04 与70677轨迹在同Run delivery/verification保留。

## 各证据层限制

75584 JSONL与两屏截图来自 seq16 后本地 `test-results`，本轮复制并计算摘要；原Connector没有给截图摘要，所以文件时间/当前存在只是辅助溯源，不作为单独制品身份凭证。新 QA HTTP handler/持久化回归无监听、无真实供应商，仅验证当前 Go 代码对受控 Transport 返回的处理。正式Actions升级/失败恢复、平台新旧Go联合链、真实供应商仍 Not Run。
