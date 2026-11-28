# 独立验收报告：交付工作台退出按钮配色与已证测试夹具修复

关联：Issue #118；基线 `70a4f55b5cfd737e0882936ab8f65bbd66fcaed2`（HEAD 已核对一致）。
验收时间：2026-11-28；验收人：独立 QA 节点（本轮执行 8）。

## 1. 被验版本冻结

| 项 | 值 |
|---|---|
| 仓库分支 | workflow/4230365d4a54629a1d863fb31526121c |
| HEAD | `70a4f55b5cfd737e0882936ab8f65bbd66fcaed2`（与 prepare、tests、code_review 回执一致） |
| 未提交改动 | 仅四处：workspace.css、console.mjs、workspace.test.mjs、tests/verify/main_test.go，与任务给定已验证补丁逐字节一致（`git diff` 全量比对） |
| 固定回归回执 | `test-results/verification/20261128T063308Z-1982769750/`（gate_exit=0、exit_code=0、source_unchanged=true、go1.27.4、chromium 147.0.7727.15） |
| 架构审查回执 | candidate `7dd49847814951b3241e5a86c91e0f1894c743f4`、regressions=0、unresolved=0、baseline_input_identical=true、baseline_model_incomplete=false |

## 2. 验收条件逐项结论

| # | 验收条件 | 结果 | 证据 |
|---|---|---|---|
| 1 | `#logout` 常态 rgb(255,241,242)/rgb(159,18,57)，悬停 rgb(255,228,230)，移开恢复 | Pass（UI-E2E） | 固定回归 console 套件新增 `logout-style` 阶段内联断言，1280/1440/390 三宽度全绿；截图人工复核（qa-evidence/console-1440.png、console-390.png） |
| 2 | Tab 聚焦 `:focus-visible`、solid 描边、常态底色 | Pass（UI-E2E） | 同上，logout-style 阶段断言 focus_visible/outline/background 三元组 |
| 3 | 桌面与 390px 窄屏均可辨认 | Pass（UI-E2E + 截图复核） | 三宽度套件通过；390px 截图中「退出」浅红底深红字与「菜单」白底明显区分 |
| 4 | `#nav-toggle` 等其他按钮配色不变 | Pass（UI-E2E） | logout-style 阶段断言 #nav-toggle 保持 rgb(255,255,255)/rgb(51,71,102) |
| 5 | 退出仍 POST `/admin/api/logout` 并返回登录页 | Pass（静态 + 旅程） | workspace.js 无改动（diff 为空）；静态断言 39/39；连续控制台旅程 J01–J07 三宽度通过（含登录/退出循环） |
| 6 | 静态断言：规则在媒体查询外、焦点描边规则存在、退出绑定与 fetch 保持原逻辑 | Pass（本机亲验） | `node --test tests/browser/workspace.test.mjs` 39/39 通过（本轮重跑） |
| 7 | `TestFixedRunnerReceipts` 13/13 场景通过；timeout_gate 回执门限 5000ms；生产门限 25 分钟与清理断言保留 | Pass（本机亲验 + 固定回归） | 本机 `go test` 23.6s 全过；`tests/verify/main.go:74` gateTimeout=25min 未动；回执 state.json gate_timeout_ms=1500000；清理断言（进程组不存在、候选目录清除）在补丁中保留 |
| 8 | `make verify` 整体通过（固定节点执行） | Pass（固定回归） | tests 节点执行 6：exit_code=0、GATE_END exit=0、CLEAN_END source_unchanged=true；console/latency/visual/self-hosted 等全部套件三宽度通过 |

## 3. 架构审查核对

- 本轮架构审查完成（与 #117 两次因引文核对失败停住不同）：baseline_model_incomplete=false、baseline_input_identical=true、regressions=0、unresolved=0，source_unchanged=true。
- 4 条基线发现为存量发现，按新流程保留在审查报告，不属于本次范围（需求第 4 节），不阻塞。

## 4. 本机亲验记录（独立于固定回归）

- `gofmt -l .`：无输出（格式合规）。
- `go vet ./tests/verify/`：通过。
- `go test ./tests/verify/ -run TestFixedRunnerReceipts -count=1`：ok，23.608s（含 cancel_gate、timeout_gate）。
- `node --test tests/browser/workspace.test.mjs`：39/39 通过。
- 截图人工复核：1440px 与 390px 顶栏「退出」均为浅红底深红字，位置与文字不变，与其他按钮可辨。

## 5. 缺陷与未验项

- 新增缺陷：无。复发缺陷：无。
- 未验项：无本范围内的 Not Run/Blocked。本机沙箱禁止本地端口监听，`make verify` 全程由固定节点在宿主机执行并有完整回执，属既定分工而非缺口。

## 6. 质量趋势与结论

- 相对上轮（#117）：完整回归一次通过保持不变；架构盘点阻塞本轮解除（基线模型完整、输入一致、零回归）；无新增风险。
- 结论：**Go**。验收通过，交人工确认交付。
