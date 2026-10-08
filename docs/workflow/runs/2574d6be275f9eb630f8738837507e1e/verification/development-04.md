# development4 修前诊断交接

## 结论与恢复
2026-10-09（Asia/Shanghai）。Owner报告本轮原生容量失败 `serverOverloaded / Selected model is at capacity`；本次沿原会话有界恢复，仅核现场、归档及新源码检查，不新建Run、不切模型、不重复UI/API写入或补造回执。T001仍 In Progress，M01仍 Active。本轮交 tests 是**真实复现前置切片**，不是产品修复完成或预期红已通过。

工作区 HEAD实核仍 `9bf6ecaf53758ca2258498968b830c535e02e1c0`。Trellis实际有 `.trellis/.version=0.6.15`、脚本和源码支撑的 frontend/backend specs；prepare 的 `initialized=false` 不是本轮初始化成功声明。本轮按 trellis-before-dev 恢复适用规范，未运行安装或重做 bootstrap；通过 trellis-update-spec 在已有 workspace-contracts 记录修前诊断契约。公共任务按研发交付Skill维护，不复制上游设计。

## 保留的首次固定入口失败
| 项 | 实际事实 |
|---|---|
| 命令 | `make verify` → cold isolated `make verify-gates` |
| UTC | 2026-10-08T20:57:56.730301Z — 2026-10-08T20:58:24.150854Z |
| 源码 | 149文件，before=after=`a47fd28fc07b82e7fc9f2f15d610084c864f5cddf678c4448956e18cf19c58cf` |
| 原失败 | `review historical HTTP rejected: historical listener unavailable; not business red` |
| 退出 | 隔离 gate_exit=2；runner exit_code=1；外层make返回2 |
| 受限范围 | review-history实际HTTP监听不可用；后续格式/锁校验/vet/Node/race/build/Chromium等未执行，不计通过 |
| 完整原始日志 | [native-verify-04.log](native-verify-04.log)，148817字节；SHA256=`071064bd7defc9013482d0c637461c3608184bbdd599e2c7bb9f099bec4adec4` |
| 完整state | [native-verify-04-state.json](native-verify-04-state.json) |
| 清理/冷起点 | node_modules/npm_cache开始不存在；manual_preinstall=none，NODE_PATH=unset；owned_group_absent=true；candidate_removed；source_unchanged/candidate_unchanged=true |

恢复后只读核对上述log的实际字节数/摘要与state吻合，并检查没有未脱敏本机路径。没有修改state/log或把受限监听补写成通过。首次固定入口覆盖的是当时中间诊断版本；其后已移动probe到J01准备后、补充分阶段事实及spec，因此不能把它当当前候选完整验收。稳定监听限制未变，按委托不再次运行同一必败固定命令。宿主必须验证下列新候选。

## 当前候选与局部检查
- 新源码：149文件，`ba8816e8f4cb38bb3b6c782195da2324bd16964f98cc2ff00448d91e78940d6f`。实际 `go run ./tests/evidence` before/after一致，命令exit0，`cmp` exit0。记录：[candidate-04-before.json](candidate-04-before.json)、[candidate-04-after.json](candidate-04-after.json)。源码摘要包含产品/测试/锁/Trellis，不包含交付回执，避免自引用。
- UTC 2026-10-08T21:44:49Z；[local-checks-04.log](local-checks-04.log)保留命令、完整输出及每项exit。
- `node --test tests/browser/model-save-readback.test.mjs tests/browser/console-summary.test.mjs tests/browser/workspace.test.mjs`：35/35，exit0。证明新筛选器、原严格摘要负例/错误身份、已有编辑生命周期等**局部契约**；不证明真实浏览器转发/调度、HTTP持久化或供应商。
- 三个诊断/旅程文件 `node --check` 均exit0；`git diff --check` exit0。没有手装依赖，未改package锁、固定门禁、race参数/超时或Connector。
- 本轮未运行完整Node wildcard、全包race、真实HTTP、Chromium、visual/迁移；全部留给原固定宿主入口，不用上述子集替代。

## trellis-check 与交付审查
| 维度 | 当前判断 |
|---|---|
| 变更范围 | 仅test诊断与其原旅程挂接、既有workspace spec、本Run公共记录；产品 `workspace.js` 与原 `save` 未修改 |
| 跨层 | 复用现有独立Go安装/服务/平台管理员context；真实PATCH和GET只由UI/实际转发触发；不变更API/schema/auth/storage |
| 断言与错误 | `summaryCell`原唯一行、必需列及状态等值保留；持有期间的失败与最终DOM失败区分；早返回对应原error最终重抛；原负例仍在门禁 |
| 生命周期 | 持有的真实响应原样转发；已声明的barrier释放并await finished后unroute，当前仅源码检查，真实Playwright清理仍待宿主 |
| 保密 | 新JSON只取固定枚举/布尔/计数/UTC，无raw body/header/ID/URL/秘密/错误文本；旧log核对摘要与路径脱敏 |
| 验证/状态 | 固定门禁Blocked，局部检查有新源码证据；未称完整Passed/Done；单任务事实源与一个适用里程碑，无重复PRD/设计 |
| 未决 | 宿主原层未运行；修复、GET失败/scope迟到/相反第二操作等新增端到端回归尚未落实；按T001剩余条件继续 |

交付审查结果：Ready for Dispatch（仅派发正式tests复现切片）；不是Ready for Release。当前不需要用户新决策。

## tests 的具体责任与返回条件
1. 只执行已有正式固定 `make verify`；保持cold copy、候选锁、Chromium版本、源码before/after、全部原门禁和清理。使用当前候选摘要，不沿用首次native受限版本。
2. 原 `tests/browser/journeys.go` 在1280/1440/390分别建立独立临时安装、Go服务及角色context；诊断挂在J01真实创建模型/价格/授权之后。在第一次1280执行内已包含1280及390的两个相反方向，以便首次失败仍有窄屏最小证据；一旦红，后续整套/1440未运行须如实报告，不能补造通过。
3. 具体目标文件：首次实际生成的 `test-results/model-save-readback-1280.json`。将其完整安全内容归档到本Run验证目录；同时归档对应 `continuous-1280-<实际pid>.json`、完整verify.log/state及实际源码/版本/清理证据，携具体路径返回development。未生成就是NotRun，不能仅交模板文件名。
4. 当前诊断故意在tenants真实响应held时 await **旧save**：若早返回，记录该次PATCH200、相同scope/page、dialog/loading/notice及严格错误；先释放tenants再独立held catalog，最后释放真实响应并严格核对新DOM；false/true与窄屏四组后重抛最早原summary错误。它预期保持红来定位，不是把预期红当验收。
5. 如果旧save没有早返回，保留原默认deadline失败和部分JSON，注明候选机制未复现。若GET转发或清理先失败，利用read_error_kind/实际log辨别驱动/环境，勿认定产品根因。若释放后的最终DOM错误，记录 `final_summary_cell`，按核心管理影响评估P1。禁止无限全套重试或未经实证轮换修复。
6. tests得到首个实际失败/完整诊断后回 development，按实证选择最小驱动完成屏障或产品修复；下一轮修改必须删除诊断中故意等待旧save的结构，转成该次成功写触发的正确scope读回/DOM完成校验，保留本轮失败与严格原断言。不得直接QA/发布或宣告修复完成。

R28-04保持Open/P2；旧tests18根因仍未知；本轮静态候选不能追认历史。动态独立QA仍Blocked，真实供应商NotRun；发布/Owner实际登录预览验收NotRun。本节点没有commit/branch/push/PR/部署/DB修改副作用。
