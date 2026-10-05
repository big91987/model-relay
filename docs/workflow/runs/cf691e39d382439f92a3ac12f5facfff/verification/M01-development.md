# M01 研发验证记录

- 对象：T001 输出边界修复；执行人：development。
- 基线：`066f035bc551b2b4d031166cebdab0849084afd4` 的当前工作树。
- 环境：Go 1.25.6、CGO、macOS 原生受限沙箱，Asia/Shanghai。
- 时间：2026-10-06 04:50–04:52 +08:00。
- 研发当时结论：**Blocked**（本地完整门禁监听受限），当时实现可交宿主验证；不改造此执行结果。当前报告门禁结论见末尾「报告阶段对账」。

## 真实执行与证据
| 检查 | 结果 | 证据 |
|---|---|---|
| 原实现的新回归 | Failed，exit 1 | [regression-before.log](regression-before.log)：正常/拒绝场景覆盖历史，关闭 stdout 时错误被吞；仅隔离 fixture 发生写入 |
| 修复后 health-history race | Passed，exit 0 | [regression-after.log](regression-after.log)：3 个进程级场景、原精确分类 14 个负例 |
| 固定 make verify | Blocked，exit 2 | [make-verify.log](make-verify.log)：evidence 成功，health-history 真正尝试旧版网络复现；三个状态均 listen bind operation not permitted，分类 REJECTED 并传播失败 |
| Trellis 补充检查 | Passed，exit 0 | [trellis-check.log](trellis-check.log)：上下文、diff-check、格式、module、vet、27 个前端 VM、build、race handler/version |
| 已跟踪历史保护 | Passed，106/106 | [history-before.sha256](history-before.sha256)、[history-after.log](history-after.log)：实际固定入口执行前后对比，无恢复操作，旧 Issue 文档/日志不变 |
| 交接前完整性 | Passed | [final-integrity.log](final-integrity.log)：相同源码摘要、历史再次全 OK、Makefile 无 diff、diff-check 成功 |

固定入口本地在 health-history 停止：其后格式/module/vet/前端/full race/current HTTP/build/browser 在这次固定执行中全部 Not Run。单独补充执行的成功仅对应表中补充检查，不拼成完整 make verify 通过。真实网络旧版预期红灯未成立；当前完整 HTTP、版本 HTTP 与浏览器待宿主。

新输出实际位于本 Run make-verify.log，stdout 中有 baseline、源码哈希、测试命令、exit_code、完整失败事件及 REJECTED；stderr 有 rejection。成功输出和 closed-stdout 的进程退出由隔离 regression 验证。受控事件不证明真实 HTTP/供应商/浏览器。stderr 输出失败不会被当成历史红灯成功。

## Trellis 实际动作与审查
before-dev 已执行 get_context.py --mode packages（single repo，backend/frontend），读取 backend/index 的 Pre-Development Checklist、runtime-contracts、guides/index 及 code-reuse/cross-layer 思考指南，并恢复本任务及旧 QA 事实。既有规范是有源码依据的完整基线，不是待填模板，因此不执行 bootstrap 或重建初始化。

update-spec 按原 runtime-contracts 原位补入输出契约：签名、stdout/stderr、错误矩阵、Good/Base/Bad、断言/真实宿主边界与 Wrong/Correct；未改变指南或业务契约。

check 再执行 package discovery，审查所有改动并完成日志中适用非监听校验。main.go 只移除持久化副作用并检查 stdout；assess、baseline、HTTP 状态断言、模块/锁和 Makefile 未变。源码复用/导入有效，无 debug 日志、抑制或产品默认假上游。单验证层变更，无存储→API→UI 数据流变化，跨业务层检查不适用。测试夹具只创建临时数据，历史保护断言覆盖全部进程场景。完整 tests/浏览器 check 不能完成，限制交宿主，未请求扩权。

公共交付审查按 Skill checklist 执行：单一结果/单一任务/单一里程碑，权威输入与 AC 可定位、关键路径清楚、无新产品范围和重复状态源、不虚构日期计划或百分比；结论 Ready for Dispatch，仅指宿主测试可接手。任务保持 In Review，里程碑 Ready for Review。

## 版本和门禁保护
本轮固定入口代码摘要：`b33aa9f6f848a8cda1bf5af08ba9c2e540d0aa09f91090930ccafa6d4e636702`（tests/evidence 的 36 文件范围，不含交付日志/.trellis）。Makefile SHA256：`048e33cc8936e629dc4421052c825f620984eac2fd622cb7c74b25b37d856c63`，与 HEAD 无 diff；顺序保留 evidence、health-history、format、module、vet、frontend-test、race/HTTP、build、实际 Go browser，任何失败 fail-fast。宿主必须核对相同源码摘要，规范另由 diff 审查。

仅对本 Run 新日志做本机路径脱敏：make 可执行路径替换为 `<make>`，linker 临时对象前缀替换为 `<temp>`；保留警告内容、测试结果、退出码及所有失败。旧文档/旧日志从未编辑。未收集凭据、供应商数据或用户安装；缓存/bin/私有 Trellis 状态仍被忽略。

## 下游复现及状态责任
tests：先对 history-before.sha256 执行 shasum 校验并冻结实际受保护文件，固定 `make verify` 输出重定向到本 Run 一个新文件，保留真实退出码和完整头尾；后续再检查哈希与 git diff（不得恢复旧日志掩盖写入）。固定 Go race 包含输出回归正常/拒绝/closed_stdout；独立 QA 可专用复跑 `go test -race -count=1 -v ./tests/health-history`，但完整入口仍必须实际通过。新日志路径由调用方决定，程序不自动选择。

QA：本 Run 保存验收与失败证据，审查 Makefile 未削弱及本轮版本匹配。报告：在 QA 后按真实证据同步 T001/M01/README 的所有当前状态与本验证引用，记录报告时点，随后生成本 Run pr.md。未发生的 Git/PR/合并/部署不写为完成；保留本地 Blocked 记录，即使宿主成功也不改造成本地 Passed。

## 报告阶段对账

- 对账人：report；时点：2026-10-06 05:08 +08:00；结论：**Passed**（当前工程交付门禁，不覆盖上文研发执行结论）。
- 验收对象：相同 HEAD 加修复工作树；源码摘要 `b33aa9f6f848a8cda1bf5af08ba9c2e540d0aa09f91090930ccafa6d4e636702`，Makefile 摘要及全门禁未变；[报告完整性记录](report-integrity.log) 保留核对退出码。报告未改源码，也未重新运行 host-only 全门禁。

| 退出条件 | 实际依据 | 当前结果 |
|---|---|---|
| 实现、回归及规范可定位 | main.go/main_test.go、README、runtime-contracts；研发 Trellis 实际执行记录；QA 静态复核 | Passed |
| 完整固定共享入口 | [同轮宿主 seq=5](../acceptance/evidence/host-connector-receipt.md) exit 0，与 QA 同版，未改 fail-fast 链支持全门禁 | Passed（Connector 层） |
| 正常、分类异常、stdout 失败 | [QA 矩阵](../acceptance/acceptance-matrix.md) AC02-REG、AC03-OUT、AC04-ERR/NEG/MAKE；原版 race exit 1，修复版 exit 0；真实沙箱 make exit 2 | Passed |
| 历史不变 | 原 106 文件清单、宿主后及 QA 前后哈希一致；报告再次校验；无恢复旧文件 | Passed |
| 指定评审及缺陷闭环 | [独立 QA Go](../acceptance/acceptance-report.md)，QA12-HISTORY-WRITE Closed，无未关闭产品缺陷 | Passed |
| 公共状态与交接一致 | T001 Done、M01 Accepted、README Delivered（仅工程范围）、本 Run pr.md | Passed |

公共交付 checklist 已核对：唯一任务事实源及其总览投影；需求/AC、运行契约、实现和验收逐条可追踪；单 bug 单任务单里程碑，无循环依赖、未决架构或重复规划；没有新产品/UI/迁移范围，requirements/design 不适用，不补造阶段文档；无下一批编码任务。审查结论 **Ready for Dispatch**，仅分发平台发布 Connector，不再分发已完成的 QA。

限制及边界：宿主后段原始日志/截图缺失，不声明逐项宿主测试数量或视觉结论；QA/研发本地 Blocked、Not Run、所有红灯和工具拒绝事实保留。QA DELIVERY 行及上文下游责任反映各自原时点，本节和当前任务/总览记录报告完成后的事实。没有真实供应商验证、提交/推送、草稿 PR、合并或部署；无新权限或不可逆操作。当前成果可交平台生成草稿 PR，不把工程 Accepted 当成已发布版本或上线。
