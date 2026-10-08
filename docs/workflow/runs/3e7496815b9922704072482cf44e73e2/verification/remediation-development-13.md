# 执行13：验证工具三项P2返工

## 结论、版本与边界
2026-10-09 Asia/Shanghai。三项最小复现均成立，已在本工作树修复；不是据审查直接认定宿主中断根因。T001仍In Review，M01仍Blocked by Validation，QA6 No-Go未解除。HEAD/base `cd413b5aa24a372a19c19af999a6ef61b5c41572`，未提交、推送或部署。

最终144件候选源码戳 `dc73ac404db4daaf7e264552c0940abb294b2b1b48684e792eec90822b6eb665`，见 [source-development-13.json](source-development-13.json)。`tests/verify/main.go` SHA256 `a57141093ee44755e7f138bdf9a0cf90889a6bce2fe74ff390ce546a9d344126`；`main_test.go` SHA256 `a3a2040ae4f5d2c9d7bdff35b05600b4581063d8305d253e9462a89b0db42cdc`。旧142件、执行11的144件4d484086、seq5不是新候选验收。

本轮仅修改 `tests/verify/main.go`、`tests/verify/main_test.go`、`.trellis/spec/frontend/dependency-preparation.md`及本Run公共记录。沿用执行11两个空私有npmrc。Makefile原层门禁、锁、业务Go、UI、DB、预算UTC、部署控制器及冻结Connector未本轮改动；已有依赖顺序修复保留。不重做需求、架构、原型，不扩大权限或时限。

## 实证红→绿
原输出：[development-13-red.log](development-13-red.log)。命令使用项目既有GOCACHE，不监听、不安装包。

| 项目 | 最小真实红灯 | 修复与当前回归 |
| --- | --- | --- |
| 父目录越界 | 临时source/internal链接到根外普通input.txt，copyFiles实际复制成功；新增拒绝断言失败，Go exit1 | os.OpenRoot守住读写根；每层Lstat拒绝链接/非目录，最终源需普通文件，目标O_EXCL不覆盖。正常嵌套、最终源链接、内/外父链接、源/目标根链接、目标父/最终链接、失败候选清理均覆盖 |
| 诊断丢失 | 实际shell分别stdout/stderr并exit25/26；原evidence/probe日志只剩exit status；新增输出断言失败，Go exit1 | 复用execute/cleanLog，stdout tee日志并独立解析，stderr完整采集；COMMAND_START/END记录阶段/退出码/取消。成功JSON伴随stderr仍可解析。实际Node调用明确test-only的Chromium测试进程exit37，验证两路输出及状态，不冒充真实Chromium验收 |
| ctx取消无效 | inventory子进程sleep0.6；run传入50ms取消仍等待约1.0375s，新增有界断言失败，Go exit1 | Git/evidence/version/probe均受ctx监督：独立进程组、两分钟辅助期限、原TERM/5s wait/KILL/组消失检查。每阶段实际子进程PID消失及向runner真实发送SIGTERM后失败状态落盘均覆盖 |

state的 `auxiliary_commands` 保留阶段、退出码、取消及实际owned_group_absent。机器stdout与stderr不混合解析。Git库存NUL仅在文字日志转义为 `\0`，解析保持原字节。最终Chromium失败转发spawnSync的stdout/stderr及status/signal/error-code，不再吞掉原层原因。

固定 `make verify` 仍自动隔离，再原顺序执行 `make verify-gates` 全部可信历史、格式、模块、vet、193Node、Go/race、build、真实Go受控browser。门禁设12分钟内部上限，低于已有宿主900秒；辅助命令受父取消约束。正常SIGTERM由NotifyContext取消，defer完成日志摘要/结束UTC/失败状态及候选移除。外层强杀仍可能阻止defer，不能事后补写清理或完成状态。

## 回归与中间失败
- [development-13-initial.log](development-13-initial.log)：初轮输出，一处旧测试仍期待CombinedOutput旧格式失败，工具输出exit1。改为stderr及阶段exit24直接断言，未修改产品/历史门禁预期。
- [development-13-tests.log](development-13-tests.log)：第一次两包race exit0，含路径/各辅助阶段取消与失败落盘。
- [development-13-tests-final.log](development-13-tests-final.log)：新增真实SIGTERM回归后race检出**测试自身**跨goroutine读取exec.Cmd.Process的竞态，exit1，原堆栈保留。不是产品runner竞态或环境失败。
- [development-13-tests-candidate.log](development-13-tests-candidate.log)：测试从实际子进程就绪文件获取其父/子PID，通过独立os.FindProcess发SIGTERM，不跨goroutine读取Cmd.Process。最终 `go test -race -count=1 -v ./tests/verify ./tests/verification` exit0；原顺序/安装失败停止断言保留。
- [development-13-static.log](development-13-static.log)：make format-check、go mod verify、make vet、make build、git diff --check均exit0。
- [development-13-cancellation-repeat.log](development-13-cancellation-repeat.log)：SIGTERM、库存取消、owned-group取消各三次race复验exit0。

这些由本节点实际执行。日志隐私转换仅工作树/用户路径、GOROOT/临时路径及NUL文字转义，错误类别、堆栈行、阶段、退出码保留。定向进程测试不是完整冷安装、真实HTTP/browser或供应商通过。

## seq12中断与现场
seq12正式输出通过read_command_output offset0→32768→61016读至EOF，truncated=false；外层exit-1，193Node通过，中断于Go/race，不是完整通过。原全文位于 `/api/workflow-runs/3e7496815b9922704072482cf44e73e2/steps/12/output?format=text`，不是本次新候选证据。

同次runner目录 `test-results/verification/20261008T174328Z-4124892975` 的 [tests-12-state.json](tests-12-state.json) 原样复制：finished_utc空、gate_exit=-1、exit_code=1、log_bytes0、cleanup=not_started，不能把外层stopped当内层清理通过。原state无归属PID，不猜测其残留子进程或杀其他进程。

[tests-12-partial-verify.log](tests-12-partial-verify.log) 直接复制同目录实际verify.log，仅标准宿主make绝对路径换为 `<make>`。原文件60919bytes/SHA256 `ef18a6069f80ca56f703fc4906fc65bce9fd7e9e4982cc707dc2bf3638250530`；脱敏副本60709bytes/SHA256 `9c8af1a736373d3b100df42712f69bc6d11f1f65343ad1413e527ced0115a08c`。这是中断前runner采集，非手抄，非61016bytes Connector全文或正常结束的新绿灯；外层97bytes前缀不在runner文件。state不因归档被修正。

本节点实际尝试只读 `ps -axo pid,ppid,pgid,command`，沙箱返回 `operation not permitted: ps`。未扩大权限、未kill历史/非归属进程；旧seq8/12残留状态仍未知。新回归受控进程PID/group退出已测，不证明旧现场已清理。本轮未重新访问5545；用户所述旧9851177/schema2不是本节点新查询回执。seq12未执行defer的原因仍需外层终止时序证据，不归因于这三项缺陷。

## Trellis与公共交付检查
恢复实际Trellis及任务；get_context --mode packages确认单Go模块，读取验证/依赖准备和共享指南，规范非空且有源码依据；prepare.initialized=false不是重初始化依据。before-dev先界定最小行为/文件集，再TDD；update-spec维护根/命令/错误/回归契约。check核对格式、模块、vet/build、两包race、stdout解析与stderr日志分离、execute复用、原锁/门禁/产品界限及隐私。无跨产品API/DB/DOM改动，不作无关重构。

交付Skill裁剪为既有T001/M01及验证记录，任务In Review、里程碑Blocked；结论Ready for Dispatch（宿主验证），不是Delivered/Accepted。原生HTTP监听限制已在本Run既有verify-before及QA报告实证，环境未改变，不重复同一必然失败的完整命令；当前源码与旧受限执行不是同戳绿灯。

## tests → 独立QA → 授权发布
1. tests仅用冻结argv原 `make verify`，不运行旧host-clean-verify.sh、不手工npm安装/删依赖、不改Connector。固定入口自行创建独立私有副本；本次真实开始无node_modules/npm cache、NODE_PATH unset、未预装，源码匹配source-development-13。
2. 从本次实际打印的唯一 `test-results/verification/<UTC>-<unique>` 获取新完整verify.log/state.json及隐私审查后的具体browser文件，归档本Run并更新索引。不硬编码旧目录，不以目录代替文件，不以截断尾部推断通过。
3. 校验clean_start_measured=true；before/after及当前同戳；finished_utc有效UTC；source_unchanged/candidate_unchanged=true；gate_exit=0/exit_code=0；log_bytes/SHA256与完整归档吻合；候选内锁解析/版本/Chromium探针、gate owned_group_absent=true、全部启动的auxiliary group absence=true、candidate_removed。停止/缺结束/存活组均非通过。
4. 完整日志覆盖原历史预期红分类器、零软预算/有效UTC/迁移、193Node、全部Go/race/build、三视口J01–J07/十页真实Go管理旅程。失败先定位阶段及原层输出，不轮换猜测或无限重跑。
5. 再由独立QA对新源码、原材料AC09–11/18/H05–07及既有UI复核；原生browser Blocked与供应商Not Run单列。QA6旧No-Go保留，seq5/8/10/12不放行本候选。
6. 合格后报告/Connector及获授权协调者完成新PR、精确已验提交合并、正式prepare/Owner deploy=true target=preview；实际Deployment/environment_url、healthz version/schema3、真实admin和旧数据密码守恒闭合才宣布上线。本节点不执行Git发布或部署。

材料design-v0.2.0，archive sha256 `5f51d35fdc15ed7214749f9413879b4238097f1d6b6bbe2cc0d508d8e103cef7`；实际manifest、AC/交互路径及单文件摘要沿用T001材料追踪，不执行包内脚本或改写材料状态。
