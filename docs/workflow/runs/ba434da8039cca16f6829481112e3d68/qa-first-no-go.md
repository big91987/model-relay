# 【产品功能】Issue63 独立 QA — 不通过，退回研发

- Issue：https://github.com/big91987/model-relay/issues/63
- Run：ba434da8039cca16f6829481112e3d68；节点qa，执行4；2026-10-25 UTC。
- 结论：**Fail**。完整固定门禁通过、受控性能与主要UI旅程通过，不等于产品验收通过。独立新增并发负例发现 QA63-01，违反冻结v2的迟到结果/冷却恢复语义；本轮不得进入approval或publisher。
- 独立QA没有修改产品、业务数据库、既有测试或冻结计划，没有提交、推送、变更权限、创建PR、部署或使用真实供应商。

## 1. 版本与完整固定回执

验收基线为用户本轮Issue63、获正式接续批准的plan-v2.md和experiment-v2.md，而非旧Run或旧绿色。冻结计划的“待审批”标题是历史原字节，不能回写成已审批；本次依据正式用户17021接续输入。

|核对项|结果/证据|
|---|---|
|plan-v2 SHA|0944af31943b26633e3813280dfe84efc5120f13dfc340221dcd67356be80a2f|
|experiment-v2 SHA|4f72f2061bc9c61c7b1f33e3ca1d8f2ab86667c7ea5743d1ed716d5b04727675|
|v2 manifest SHA|64b550ad798e248ad96cbaa1a463a5e7822f4e4ac7ee9cc3db44b24f55af0544|
|HEAD/基线|f06f05de20ffac5b8c8e772958c808379001f0e3；改动仍在工作树，不把此HEAD称功能提交|
|固定/QA全源码摘要|9f7a656a749ef67af8604af68a04cced5fdc31d368ac2e61005dd4bf9ffcfd66；170文件，范围由tests/evidence定义|
|研发改动清单摘要|5607062a855ff7151718973ed5762ac5a83cc49fdbde2ba9be791f7847214420；29文件，和全源码摘要口径不同|
|同版验证|独立执行tests/evidence；source-before.json/source-after.json完全相同且匹配固定state的source_before/source_after；v2原9件、研发29件源码及103件证据逐bytes/SHA全部校验通过|
|固定入口|make verify → 私有干净candidate的make verify-gates；非本节点重跑全套|
|固定主gate|12:52:47.680374982Z—13:13:02.818474957Z，1215.138100秒，exit0，未取消；1500000ms未改|
|固定含准备清理|12:52:33.012354149Z—13:13:03.971724996Z，1230.959371秒；Connector外层原回执1238.149秒为不同口径|
|完整项目原日志|test-results/verification/20261025T125233Z-3805794132/verify.log，1125044bytes，SHA0572e51d7ec3eae3e94b6ee5999c9cb454a1a3649c5b9b9e15019e79af4a7586，至CLEAN_END/EOF|
|完整Connector原日志|workflow-output/ba434da8039cca16f6829481112e3d68/command-3.log，1161962bytes，SHA6874c61c65293a345b0608617949a0e5ae2e21fceb46b8a8c57f165a3b2420cf；正式read_command_output导出，truncated=false|

固定日志中嵌套的exit23/取消等是verification自身的受控负例，不能误当本轮主gate失败；真实主gate末尾exit0。原format、锁定模块校验、vet、frontend、Go race、build、迁移、legacy-ui、routing、console失败恢复和full-console均保留。独立审查Makefile确认routing-experiment在build后browser前无条件执行，完整日志精确出现一次该命令。浏览器阶段耗时：15.618/179.709/241.369/36.320/254.252秒，未裁旧覆盖。源码快照均为.txt/patch，不加入cmd/internal/tests编译树。

## 2. 实际独立执行（不是研发代签）

所有动态测试只访问临时目录和环回临时端口的真实Go网关/受控HTTP上游。

1. `go test -race -count=1 -v ./internal/relay -run 'TestLatency|Test.*(Admission|Budget|Isolation|RPM|TPM|Concurrent|Unknown|Rout.*Grant)' -parallel=4`：exit0/EOF，包运行48.840秒，命令墙钟140秒（含首次依赖准备），原回执independent-go.log。实际匹配范围以原日志为准，不把regex未匹配的测试冒称独立运行。
   - 实测59/60/61秒测量TTL边界、逐样本过期、普通/SSE分池、有界探索、无输出、新候选、冷却与恢复、配置变更16并发及成功间乱序、迁移/非法策略/部分扩展损坏、存储失败、SSE心跳取消和终端写失败。
   - 实测low_latency租户/Key各自calls、非零RPM、非零TPM、并发、金额预算矩阵，拒绝不触达上游；不是把这些规则相互替代。
2. `go run ./tests/latency-ui`：exit0/EOF，命令88秒（含编译/准备），重新启动真实网关；1280与390px分别完整页面配置两个服务及同模型候选、选择low_latency、保存重入、配置授权/价格/Key、普通及SSE调用、同Request ID诊断/日志/费用、重启保留、页面切回weighted并重入确认。非API/SQL替代业务页面步骤。夹具只准备租户、受控上游和重启能力。
   - 原回执independent-ui.log、latency-ui-1280.json、latency-ui-390.json与16张安全模态截图。每次成功1attempt、1费用条目、fee14；普通和SSE指标分别response_body_first_byte_ms/first_output_delta_ms。
   - 检查截图策略及390px选择依据，诊断展示当时策略、指标、代次、理由与未知样本，不展示调用输入/输出/凭据。脚本对诊断JSON及截图DOM检查敏感哨兵；没有采集原始trace。
3. `go test -race -count=1 -v ./internal/relay -run '^TestConsoleResourceMatrixActualHTTP$' -parallel=4`：独立复验默认weighted的同一准入矩阵；原退出码/耗时见weighted-admission.log。
4. 独立新增QA overlay负例两组：仅通过Go `-overlay`给测试编译注入文本快照，不修改产品或既有测试、不把源码副本放进构建扫描目录；原失败log保留，未循环求绿。
   - `go test -overlay harness-output/issue63-qa/overlay.json -race -count=1 -v ./internal/relay -run '^TestQA63LateSuccessAfterNewerFailure$'`：**exit1**，older成功在newer临时失败清池后被accepted，池重新有1样本。
   - `go test -overlay harness-output/issue63-qa/watermark-overlay.json -race -count=1 -v ./internal/relay -run '^TestQA63SequenceWatermarkSurvivesFailure$'`：**exit1**，更强序列反例，详见QA63-01。
5. `node harness-output/issue63-qa/audit.mjs`：exit0/EOF；独立从固定14份JSONL重新计算2480样本数量、请求ID唯一、成功/attempt/费用、每窗口分布、均值/p50/p95、原效果门槛、互换40次内适应、早心跳及长尾反例。此项是**原始证据复算**，不是本节点再发2480请求；没有声称独立重复全量实验。

## 3. 新发现与返工要求

### QA63-01 — P2，冷却清池丢失顺序水位，旧成功污染恢复样本

**状态：未修复；阻断AC03/AC05的完整通过。**

用户影响：在同一上游并发调用时，新请求临时失败会清除评分池，但失败前发出的旧请求稍后完成仍被当作可采纳的近期成功。诊断错误显示accepted；恢复池不再保证从故障后的正常请求重新学习。没有观察到绕过冷却直接派发，不能扩大为权限/计费泄漏或故障时重放。

独立最小复现（普通真实HTTP、单一已授权上游、low_latency）：

1. 派发R1，上游写有效JSON并flush，但用barrier暂缓结束，使R1保持在途。
2. 派发R2，正常结束；检查其持久化延迟证据为accepted，证明更高dispatch sequence已经被采纳。
3. 派发R3，上游返回503；检查其sample_acceptance=not_success，确认临时失败后池为0样本。
4. 释放R1，使旧成功结束，读取其持久化证据。
5. 预期：R1不得复建故障前评分，必须拒绝旧序列/恢复代次；实际：R1为accepted。原失败中R2=9GOjHmW62NgcbKz0OwtNnw、R3=WBf6at2cLpvWsoS_iRGHjw、R1=cTunOC-ZENLT431ZNlFhhQ，失败前后所有请求均为隔离合成流量。

根因静态核对：internal/relay/routing.go:359—361在临时失败时delete普通/SSE latency pool；internal/relay/routing_latency.go:253只在pool存在时检查pool.Sequence，而该顺序水位随pool一起丢失。acceptLatency未独立检查更高序列的健康/失效水位。随后internal/relay/routing_latency.go:270重建pool并采纳旧结果。diff-review.txt保留原代码带行号摘录及相关差异。

违反冻结plan-v2 §3.6“较低dispatch sequence的迟到成功不得覆盖已采纳的较新派发结果”、§3.7“冷却恢复重新用正常请求收集”，对应AC03/05。不是仅以测试数量或局部门禁替代用户结果。

**返工：** 保留独立于可评分样本的单调序列/失效水位或等效的恢复代次；清池不能抹掉旧派发拒绝依据；同时核对普通/SSE跨池临时失败及配置代次语义。诊断必须准确记录拒绝原因。将上述真实HTTP顺序反例纳入项目原固定make verify，并补普通/SSE、同模式/跨模式、故障后新请求可重新学习、不绕过冷却的回归。不得仅调整QA断言或弱化冻结标准。修复后先定向回归，再按原25分钟/30分钟固定入口对新候选重新交测及独立QA；保留本轮两份exit1原件，不能沿用旧候选绿色。

## 4. AC — 结论 — 证据映射

|AC|结论|证据与验证性质|
|---|---|---|
|01 默认/保存/CAS/重启/升级/非法策略|Pass（已测范围）|独立双屏UI；independent-go.log配置/迁移测试；固定原迁移阶段；静态事务/CAS差异审查|
|02 合格候选、隔离/授权/优先级/冷却|Pass（已测范围）|独立冷却/原错误allowlist/准入测试；固定routing原始回执及静态beginAttempt顺序；本节点未重做所有旧隔离场景|
|03 冷启/新候选/过期/减速/恢复/5n|**Fail（组合边界）**|独立常规用例通过、固定对照复算通过；QA63-01表明并发迟到破坏故障恢复清池|
|04 SSE有效增量、心跳/role/工具/长尾|Pass（受控协议）|独立输出判定/HTTP取消；固定160心跳+80尾部原样本复算；不是首token/供应商性能|
|05 测量TTL、失败/取消/迟到采纳|**Fail**|59/60/61、配置变化、成功间乱序、取消/终端写失败独立通过；序列水位被临时失败清除的新反例exit1|
|06 不重放、最多3attempt、未知用量|Pass（已测范围）|独立心跳取消/终端失败/unknown totals；固定完整回归与静态代理逻辑；本节点未逐项重发所有旧异常组合|
|07 独立calls/RPM/TPM/并发/金额预算|Pass|independent-go.log的low_latency矩阵；weighted-admission.log的默认矩阵；均真实HTTP|
|08 同ID选择/日志/费用/隐私/权限|Pass（已测范围）|独立双屏4次调用同ID及费用14；JSON/DOM敏感哨兵；原固定权限负例审阅；未人工逐像素检查全部198张历史截图|
|09 原规模/门槛/分布/代价|Pass（固定实验原件审阅+复算）|audit.json；固定14JSONL、summary.json/SVG；本节点未再独立跑全2480对照|
|10 固定同版/独立QA/可访问最终交付|部分Pass/最终Not Run|本候选固定及QA摘要匹配，原件已分片并逐文件核验；存在QA失败，approval/report/publisher/真实draft PR/释放后下载均未执行|

## 5. 性能对照审阅（不冒称生产性能）

固定2480逐请求原件：核心2240（两轮×普通/SSE×两策略×280），心跳陷阱160，长尾80；2480不同Request ID，100%成功，2480attempt，全部单attempt、单费用条目。固定窗口12:57:39.715001971Z—13:00:53.901170179Z，194.186秒。和研发旧实验是两次不同执行，不能混为同一份结果。

独立复算四个core稳态cell：weighted快端50%，low_latency快端90%；均值改善分别52.6484%、52.5020%、52.6282%、52.5274%，满足冻结80%/35%门槛；互换恢复计分段均90%，满足75%，40次适应窗口内实际出现新快端lowest_latency。JSONL中的未知/不足样本保留，低于3条不评分。

收益及代价：普通均值约60.85→28.81ms、SSE约60.86→28.91ms（第一轮），是启动指标而非完整生成耗时；普通首body与SSE首有效增量不混平均、不称首token。p95依旧约101ms，探索仍付出慢端成本，不能宣称尾延迟同幅改善。早心跳上游body先到但输出晚的反例按有效增量评分；长尾只影响duration，不代入评分。成功后才采纳、少于3条未知以及过期/重启清池都有学习代价。QA63-01是性能实验未覆盖的故障并发组合，不能用稳定效果数字抵消。

## 6. 证据留存、环境与未验证范围

- 原固定browser目录378件约43MB，包含其他历史旅程与本轮原始样本。为满足每节点256件/每件8MiB限制，独立无损分为8份fixed-browser-01…08.tar.gz；fixed-browser-archive-manifest.json逐原路径记录bytes/SHA与所在分片。所有分片新目录解压后，378原件逐一核bytes/SHA通过（archive-validation.json）；未移动、覆盖或删减原件。源码副本仅.txt快照，不进入构建输入。
- QA实际命令的完整stdout/stderr、原退出码、EOF及独立UI16截图另列；两份失败证据不覆盖。harness-output/issue63-qa/manifest.json列所有本节点交接files具体路径/bytes/SHA，不用目录代替。
- 初次shell的PATH不含Go，`go`未找到；未据此报产品失败。定位既有/usr/local/go/bin/go后显式补PATH，Go1.25.6；没有增加权限或安装后台。首次动态命令包含锁定模块下载/编译时间。网络只用于正常锁定依赖，无外部付费请求。独立动态不要求sandbox提权。
- 复用项目测试脚本执行和QA新编写overlay复现均明确标为独立执行；固定2480统计及其余历史浏览器结果为原始证据审阅/复算，不能叫全部独立重做。
- 真实供应商、真实供应商性能、生产负载/分布式场景：**Not Run**，本任务无凭据；不使用历史聊天密钥。原始浏览器trace：**Not Run（隐私）**；没有声称trace已脱敏。安全结构化回执/模态截图替代，未留存请求或回复正文/工具参数/认证header。
- 研发初次UI导航失败及修复前log仍在原103件证据中，通过bytes/SHA核验；本节点没有发现其他已修复缺陷，也没有替研发修改QA63-01。该新缺陷仍未解决。
- 无真实PR结束点，因此Issue创建到PR总耗时、approval/report/publish/释放下载阶段均待后续真实回执，不能估算或标完成。本QA读证/执行始于约13:13Z；命令时刻与耗时见原件，阶段驻留含准备等待与归档，不称Agent净耗时。原生token/缓存/费用没有可得原回执，不估算。

## 7. files（核对到的候选仓库改动，不是发布放行清单）

以下为研发source-manifest的29件实际改动，已逐bytes/SHA核对；不含harness-output临时输入。冻结设计/实施文档另见本轮docs/workflow/runs目录的具体交接路径。修复后必须生成新候选清单，不能直接沿用本表签发PR。
- `.trellis/spec/backend/routing-contracts.md`
- `Makefile`
- `README.md`
- `internal/relay/configuration.go`
- `internal/relay/console.go`
- `internal/relay/console_api.go`
- `internal/relay/console_reports.go`
- `internal/relay/console_resource_matrix_test.go`
- `internal/relay/proxy.go`
- `internal/relay/routed_proxy.go`
- `internal/relay/routing.go`
- `internal/relay/routing_latency.go`
- `internal/relay/routing_latency_storage.go`
- `internal/relay/routing_latency_test.go`
- `internal/relay/server.go`
- `internal/relay/store.go`
- `internal/relay/upgrade.go`
- `internal/relay/web/app.js`
- `internal/relay/web/index.html`
- `internal/relay/web/workspace.js`
- `tests/browser/journeys.go`
- `tests/browser/journeys.mjs`
- `tests/browser/latency-focus.mjs`
- `tests/browser/latency-journey.mjs`
- `tests/latency-experiment/main.go`
- `tests/latency-experiment/main_test.go`
- `tests/latency-ui/main.go`
- `tests/latencyfixture/fixture.go`
- `tests/latencyfixture/fixture_test.go`

## 8. 留存来源补充

独立UI输出截图复制时glob同时带入已有的失败JSON；经SHA核对它与研发ui-initial-failed.json完全相同（db02484c2cf9160544a92de10ba98375f8e1c1ac0acf5d09764959e19c9c9791）。已仅改归档名称为preexisting-development-ui-1280-failed.json，原字节及原文件未动，映射见provenance.json。这不是本节点新增UI失败；独立UI的真实exit0/EOF以independent-ui.log为准。
