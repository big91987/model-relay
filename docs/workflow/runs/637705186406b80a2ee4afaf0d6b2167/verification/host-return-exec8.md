# 执行8：tests7完整失败归因与计量门禁调度整改

## 恢复与证据身份

HEAD仍为 `3f9cee1615e87e002a2c2300f2d29778072ec3f5`。原材料仍为README所列同一manifest、`product-test33-handoff-r1`和ZIP SHA256 `a13fb6b9089cefbb9d2b6f9a333830e0ee6866bd7ff45d5ba881da97e9a66a49`；本轮重核原PRD/interactions/acceptance、report、independent-review及plan-review对应文件摘要。冻结设计、PLAN-01–12及历史24次/64tokens/90秒授权不变。

本轮最终工作树SHA256 `2dfe14a3c305c13621f23b7dd287a8313b7e4f39ab64b8281d283450d125d8ed`（154个源码/测试/锁/契约文件），由原 `go run ./tests/evidence` 产生并闭环复核一致。旧078e1aa1候选与测试文件摘要另存candidate-source-exec6.json / candidate-test-code-exec6.sha256，不覆盖其失败版本身份。

直接恢复本工作区tests7实际保留的 `test-results/verification/20261009T124932Z-3450406578/{state.json,verify.log}`，分别归档为 `host-tests7-state.json` 和 `host-tests7-full.log`。后者496183字节、SHA256 `856d2529c78a86d16ad1f36bc8333b9f72354ad861ec904fe12907a1f88371f1`，与state.log_bytes/log_sha256逐字节匹配；Connector外层496396字节含包装输出，不混称同一文件。两次source摘要均 `078e1aa1c31f63fa98308db68e3e6fa5c29d2554409c3f498d2f1c6af4921019`，source/candidate unchanged、owned_group_absent均true，cleanup完成。旧tests5完整失败与执行6证据全部保留。

## 已证事实和待证假设

| 边界 | tests7真实结果 | 结论/责任 |
|---|---|---|
| 全部门禁 | Connector exit2；state.exit_code1/gate_exit=-1，内部12分钟deadline终止；不是900秒外层超时或本地监听限制 | Failed/Incomplete；tests须新候选完整复验 |
| Go计量/实际HTTP/进程 | 全包race通过；relay包363.407秒，metering每个ActualHTTP/InProcess通过 | 受控内部回归事实，不替代供应商；新调度仍待宿主 |
| 冷却 | 1280/390实际GET均200，matching/target各1、candidates2/cooling1/expired0/failed0；remaining29732/29752ms，DOM冷却true | 原两屏routing均完成，自然恢复原断言通过；未重现tests5冷却故障，不能归为已修产品缺陷 |
| 浏览器相位 | migration14799ms、legacy96835ms、routing130674ms、console15447ms均成功；full-console13:01:12.982084Z开始 | 全阶段累计预算，不能只看最后一步 |
| full-console | 1280完整旅程/视觉/汇总通过；1440仅J01/J02完成，J03-hour-grace navigation被SIGTERM打断；390未运行 | 不是独立于deadline的J03产品红；新候选同断言待测，不猜改产品 |
| 真实DeepSeek/收费/隐藏推理/历史预算 | 本研发新供应商生成0，前置仍未证明 | Blocked/Not Run；专用合法测试Pipeline，原No-Go不改 |

已证事实是：独占安全/计量相关fixture仍串行占用较大wall time；tests7采用新增metering顶层并行，观察到relay包较tests5减少60.065秒，但完整门禁仍未完成。这不证明宿主负载固定或全部差值因果，也不证明下一候选一定能在deadline内完成。未证假设是下一完整冷门禁是否足够、tests5冷却等待的瞬间根因；不以假设性产品修复关闭它们。

## 本轮最小变更边界

行为差距在验证资产调度，不在产品限额或UI。只对同一AC09–13/11相关的18个既有安全/计量/共享池/时钟与恢复顶层测试添加 `Parallel()`，复用已有 `-parallel=4`。涉及console_contract/http/resource_matrix/review_regression/clock_queue/clock_recovery六个Go测试文件；不修改helper实现、子用例顺序、竞争屏障、真实子进程、实际HTTP变体或任何断言。每个测试创建独占TempDir/SQLite/Server/clock/channels；ClockRecovery仅向自己的exec.Cmd.Env传独占状态文件，不变更父进程环境。原metering及冷却白名单诊断保持。

非目标：生产Go/UI/schema、密码哈希成本、Makefile或十二分钟deadline、供应商与调用预算、冻结设计、#30/#32在途功能、平台功能。未确认生产缺陷数仍0；没有虚构产品首红/修后绿。tests7累计失败是真实完整首失败，不是本地沙箱限制；本地既有监听限制未改变，不重复必然失败的make verify或npm网络安装。

## 本地验证与Trellis检查

同命令8组InProcess race基线31.251秒、调度后9.632秒，均exit0；日志保留为 `metering-scheduling-before-exec8.log` / `metering-scheduling-after-exec8.log`。这仅证明所选子集隔离与耗时变化，不证明全冷门禁适配。当前组合回归、静态检查与新候选摘要由 `evidence-index-exec8.json`、`candidate-source.json`和本Run结果页给出，后续独立QA必须逐项读取真实回执。

| 命令（Go统一使用工作区缓存、CGO_ENABLED=1、GOTOOLCHAIN=local） | 真实结果 | 保留日志 |
|---|---|---|
| `go test -race -count=1 -v ./internal/relay -parallel=4 -run '^TestConsole(Governance\|ResourceDimensions\|PauseUsage\|SharedResources\|ReviewZeroSoftBudget\|ReviewClockRollback\|ClockQueuedDeadline\|Accounting)InProcess$'` | 同命令前/后exit0；31.251s/9.632s；8个顶层测试均通过 | metering-scheduling-before/after-exec8.log |
| `go test -race -count=1 -v ./internal/relay -parallel=4 -run '^Test(Metering.*InProcess\|Console(Governance\|ResourceDimensions\|PauseUsage\|SharedResources\|ReviewZeroSoftBudget\|ReviewClockRollback\|ClockQueuedDeadline\|ClockRecovery\|Accounting\|ResourceMatrix)InProcess)$'` | exit0；18个顶层测试全部通过，40.805s；包含真实ClockRecovery子进程但不是HTTP变体 | metering-combined-exec8.log |
| `make format-check && go mod verify && make vet build && go test -run '^$' ./... && node --check tests/browser/routing.mjs && node --test tests/browser/cooldown-observation.test.mjs tests/browser/routing-progress.test.mjs` | exit0；格式、模块、vet/build、全包编译、3个诊断Node测试通过 | static-exec8.log |
| 新候选完整make verify/ActualHTTP/全部浏览器 | 本地Not Run，监听限制未改变；由tests宿主实际执行 | 原基线沙箱失败、tests5/7宿主失败均保留 |
| 本地完整frontend-test | Not Run；缺playwright-core和锁定npm ci网络EPERM的执行6首失败仍适用，未重复必然失败安装 | frontend-exec6.log / frontend-locked-exec6.log |

trellis-before-dev已执行packages discovery并恢复实际后端/前端规范、shared guides及交付Skill的公共契约/审查清单。规范非空且Trellis脚本存在，无需bootstrap；update-spec仅追加真实回执/源码支持的调度边界与冷却结果。trellis-check复核修改仅为独占测试调度、原断言与清理不变，无新协议/存储/跨层字段、依赖/警告抑制或秘密输出。完整宿主门禁未验不得标全部通过。

## 下一责任与退出条件

tests对新candidate-source精确摘要重新完整 `make verify`，保留冷安装、review/health历史、race、真实进程及全部浏览器宽度和原deadline，保存完整state/full log及前后source一致性。先确认完整结果，失败时根据相位、余量和原断言收集最小归因，不扩大超时/重试。只有完整门禁绿后再独立计划评审/QA；对原AC/PLAN逐行数值和真实供应商预算审查继续独立开展。M01仍Ready for Review，T001仍In Review，研发可交接不是产品Go；发布仅Connector处理。
