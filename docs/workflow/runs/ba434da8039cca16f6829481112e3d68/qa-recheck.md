# 【产品功能】Issue63 独立 QA — 新候选通过

- Issue：https://github.com/big91987/model-relay/issues/63
- Run：ba434da8039cca16f6829481112e3d68；qa执行9；2026-10-25 UTC。
- **结论：Pass（本功能候选独立QA），允许进入approval。** QA63-01已在新同版固定门禁、原独立负例、普通/SSE同池及跨池恢复复验后关闭。旧QA Fail、两份exit1和tests6真实失败保持原字节，不把历史绿色替代当前结果。
- 这不是合并、部署、真实供应商认证或PR交付完成。approval→report→固定publisher→精确远端head核对→释放后全新下载核验仍待实际完成。
- QA只写验收报告与证据；没有改产品/既有测试/冻结标准、业务DB、权限、凭据、分支或发布状态。以下严格区分独立动态、原始证据审阅/复算、静态审查及Not Run。

## 1. 验收依据与精确版本

依据为本次Issue63用户结果、正式用户17021批准的冻结v2、后续18223预算更新及本轮真实实现，不是旧Issue51或设施PR62的绿色。冻结文件中“待审批/Not Run”是设计阶段原文，正式批准由用户输入证明，不覆盖原件。

|核对项|结果|
|---|---|
|plan-v2 SHA|0944af31943b26633e3813280dfe84efc5120f13dfc340221dcd67356be80a2f|
|experiment-v2 SHA|4f72f2061bc9c61c7b1f33e3ca1d8f2ab86667c7ea5743d1ed716d5b04727675|
|v2 manifest SHA|64b550ad798e248ad96cbaa1a463a5e7822f4e4ac7ee9cc3db44b24f55af0544|
|HEAD|f06f05de20ffac5b8c8e772958c808379001f0e3；仅基线，功能仍在工作树，不冒称功能提交|
|当前已验源码|171件，SHA6316a232322de48fdce31a4420a872495f9e7ea2237c36e9e2742a4e3566a7e0|
|独立同版核对|tests/evidence的source-before.json与source-after.json逐字节相同，并与tests8 state.source_before/source_after相同|
|清单核对|冻结9件、当前源码171件、研发交接149件均逐bytes/SHA核验；audit.log保留完整结果|
|相对tests6|独立比较前候选改动清单与其余HEAD原件，仅internal/relay/routing_test.go改变；不是产品修复后偷偷另改实现|

源码摘要范围由tests/evidence定义（产品/测试/构建锁/契约及历史验收设计），不包含所有本轮交付文档；冻结v2和当前46件实际仓库差异另由逐文件清单绑定，避免把摘要范围说大。当前完整源码文本快照及原路径/bytes/SHA沿用已核验的issue63-rework-2/source-snapshot.txt与source-manifest.json；证据不作为Go编译输入。

## 2. 完整固定门禁：审阅原始证据，不冒称QA重跑全套

- 固定入口make verify→冷隔离candidate的make verify-gates；tests8实际执行成功。内部1500000ms、外层1800秒未变；node_modules/npm cache初始不存在，无手工预装，candidate/source unchanged，owned group清理确认。
- 主gate：2026-10-25T14:18:51.403544406Z—14:39:16.399035540Z，**1224.995491134秒，exit0，未取消**。含准备清理1241.135218549秒；Connector execution_time=1241222ms是不同口径，不混加。
- verify.log：1133068bytes，SHAf24391a7c9b4194da1b35f3c2e32c5bbc13ee092a12c9fae493bfd037a03764d。独立校验全文件与state、末尾CLEAN_END/EOF。
- 正式read_command_output(seq=8,save_to_workspace=true)导出command-8.log：1169995bytes，SHA219242268d200152d7c7dc9503bf20bbc570d67b590f278f23f382de6ff32af3，truncated=false，包含execution_complete。不是使用对话中省略的输出作为完整证据。
- Makefile静态差异仅新增唯一无条件routing-experiment，位于build后browser前；完整log中该命令恰好一次。原format/模块锁/vet/frontend/Go race/build/迁移/legacy-ui/routing/console/full-console全部保留；runner自身预期负例与health-history旧红字不等于当前主gate失败。
- 浏览器实际阶段秒数：migration15.715、legacy-ui179.039、routing242.072、console35.948、full-console254.393。完整固定原始浏览器378件已无损分片，独立新临时目录解压后逐bytes/SHA核验378/378。
- 正式预算1488.670810秒只作已批准成本预测，不是通过依据。本次实际主gate距1500秒275.004508866秒；没有利用余量增加预算或裁原覆盖。

原tests3成功仍只绑定旧源9f7a…；原tests6主gate exit2、runner state exit1、Connector exit2及246.890903218秒提前失败继续留存，不能当完整成本或绿色。完整回执见交接具体files。

## 3. 本节点实际独立执行

真实Go HTTP服务、隔离临时数据、环回临时端口和明确标记受控上游；没有静态browser.check代替API。复用项目测试入口进行独立执行不等于独立重新实现全部测试；原QA overlay为执行4独立编写且本次保持原断言。

|独立命令/内容|真实结果与回执|
|---|---|
|go test -race -count=1 -v -overlay=harness-output/issue63-qa-recheck/overlay.json ./internal/relay -run TestQA63 -parallel=4|原两并发负例2/2 Pass；包2.777秒、整命令215.123秒（冷编译/依赖）；original-negative-recheck.log exit0/EOF|
|go test -race -count=1 -v ./internal/relay -run 'TestLatency\|TestRoutingPassiveCooldown\|Test.*(Admission\|Budget\|Isolation\|RPM\|TPM\|Concurrent\|Unknown\|Rout.*Grant)' -parallel=4|实际匹配21个顶层测试Pass；包51.175秒，整命令237.476秒；independent-go.log exit0/EOF。以原log实际匹配为准，不把regex未匹配测试当已执行|
|安全/权重补充组（完整9个精确测试名见原log）|weighted实际HTTP十维矩阵、治理/人员隔离、拒绝元数据、最高优先级/选择证据、3attempt上限、已输出SSE不重放、超时取消及健康代次；包11.084秒、整命令11.947秒；safety-weighted-http.log exit0/EOF|
|go run ./tests/latency-ui|实际Go服务+Playwright，1280及390各独立临时安装/受控上游；两屏Pass；independent-ui.log exit0/EOF，整命令204.024秒（含冷依赖编译）|
|node audit.mjs / integrity.mjs；tests/evidence；git diff --check|原始样本复算、完整回执/归档校验、执行前后源码一致；最终均通过，原回执留存|

延迟组实际涵盖：普通/SSE测量59/60/61秒、逐样本过期、完成不续期；1/2/3/100候选有界探索与分池；新候选/无输出未知；role-only/空content/工具片段/refusal；存储失败、终端write/flush失败、取消；16种并发代次/序列场景；schema1/2/3及扩展损坏；低延迟策略tenant/key各calls、非零RPM、非零TPM、并发、金额预算的独立十维准入。补充组执行原weighted十维，拒绝时验证上游派发不增、独立资源不互相替代。

### 双屏实际用户旅程

本节点启动的真实Go服务中，通过浏览器原生界面创建A/B两个同模型上游、选择low_latency、保存并重入、设置授权/价格/可信上界/Key、发普通与SSE、打开同Request ID日志/选择/观测/费用、重启服务读回策略并在页面切回weighted，再重入确认。没有用API/SQL替代承诺的配置/调用/诊断UI；临时租户准备和重启控制为显式夹具。

|屏宽|普通Request ID / SSE Request ID|同ID验收|
|---|---|---|
|1280|cSDqga6DAEYBB-bhskfrTQ / UC1o82Iz4-ryudShhkTMFA|均success、1attempt、1费用条目、fee=14；low_latency选择、对应独立metric、accepted|
|390|-m3F7YF8srkRUTJll5ss5g / _yIq9WAknYHyT3FM1tnH3g|同上；无页面横向溢出|

16张安全模态截图和两份结构化JSON按准确文件名复制归档，没有glob混入旧失败。人工查看390策略、SSE选择/观测、1280普通诊断截图：字段可读、模式命名清楚、未见提示/回复/凭据；诊断内容较长需纵向滚动，保留少数底层字段/非选中原因英文标识，属于非阻断可用性限制，不将其称全面视觉重设计。

## 4. AC—结果—证据映射

|AC|结果|验证方式/证据|
|---|---|---|
|01 默认、合法性、CAS、重入、重启、升级|Pass|独立LatencyPolicyMigrationAndCompatibility、MigrationOldSchemaAndCorruption、双屏JSON/截图；固定配置/frontend回归审阅|
|02 合格最高优先级、授权隔离、原冷却|Pass|独立RoutingSelectionAndEvidence、GovernanceActualHTTP、PeopleScopeActualHTTP、DeniedMetadataAndStorage、PassiveCooldown、HealthEpochSequenceDisableAndRestart；beginAttempt先过滤再排序静态审查|
|03 无样本、新候选、过期、减速恢复、5n|Pass|独立ExplorationBoundAndModeIsolation、NoOutputAndNewCandidate、SamplesExpireFromMeasurement；固定2480条复算四core cell与40次适应段|
|04 SSE心跳不能冒充输出；role/空/工具/长尾|Pass|独立OutputDeltaContract/HeartbeatCancellation；固定160心跳+80尾部原样本复算，不混传输body与有效输出|
|05 测量TTL、失败取消、迟到、配置代次|Pass|独立原QA两负例、FailureWatermark八组合（普通/SSE同跨池×有无既有成功）、MeasurementAgeAndLateConfiguration、ConcurrentGenerationAndSequence、TerminalWrite、StorageFailure|
|06 输出/取消不重放、最多3attempt、未知用量|Pass|独立CommittedStreamNeverReplays、DeadlineAndCancellationNeverRestart、FourCandidatesStopAfterThree、TemporaryAllowlistAndUnknownTotals、HeartbeatCancellation；固定SSE工具/错误回归审阅|
|07 独立额度、非零RPM/TPM、并发、金额预算|Pass|独立low_latency与weighted两份ActualHTTP十维资源矩阵，原准入先于dispatch保持|
|08 同ID选择/日志/费用、权限、隐私|Pass|双屏四个实际请求JSON/截图及脚本内敏感哨兵断言；独立治理/人员/拒绝元数据负例；诊断typed scalar结构和授权查询静态审查|
|09 冻结两轮规模、阈值、分布与代价|Pass（原证据独立复算）|audit.log复算2480条、成功率/attempt/费用条目、均值/p50/p95/完整duration/窗口；未在QA重复2480请求|
|10 完整同版固定与独立QA、可审阅交付|Pass（至QA）；后续Not Run|完整tests8日志+source前后+归档还原；approval/report/draft PR/精确远端head/释放后下载核验由后续节点与Builder完成，不能提前标全链路已交付|

## 5. 受控效果复算与代价

固定新候选采样2026-10-25T14:23:53.196491390Z—14:27:07.547717223Z，194.351秒。原始2480唯一Request ID=2240 core+160 heartbeat+80 tail，全部status200/state success、1attempt、1charge、有效观测accepted；总attempt2480、成功率100%。这是受控环境结果，不代表真实供应商可靠性。

|轮/模式|weighted稳态快端|low_latency稳态快端|均值改善|互换恢复计分段快端|
|---|---:|---:|---:|---:|
|1/普通|50%|90%|52.6653%|90%|
|1/SSE|50%|90%|52.5077%|90%|
|2/普通|50%|90%|52.5690%|90%|
|2/SSE|50%|90%|52.5368%|90%|

每core cell学习40、稳态120、互换适应40、恢复80单列；核心80%/35%/75%门槛全部满足，40次适应窗口内出现基于新评分的lowest_latency新快端选择。心跳陷阱稳态亦90%选真正早输出端，不是早body端；逐慢端样本确认body早但metric等待有效输出。尾部两端首输出均约20ms，完整尾部20/200ms，评分不跟完整duration混淆。

独立复算每窗口metric及client_duration的mean/p50/p95、成功数、attempt数、缺失值数与summary一致，完整结果见audit.log。第一轮普通稳态均值60.911956→28.832487ms；首输出收益不能称完整生成收益或TTFT。p95仍约101ms，探索保留慢尾，冷启/样本不足/重启清池和成功结束后才学习均有代价；长流>60秒可能没有可用评分。没有隐藏后台付费探测。可视化SVG及experiment-results.md是明确标识的研发原实验，不把其旧数值冒称tests8；最终报告应引用本QA新复算。

## 6. 缺陷闭环与差异审查

**QA63-01（曾阻断，本候选关闭）**：较新请求503清池后，故障前已在途的旧成功被采纳、重新污染恢复池并可能清除冷却。原late-after-failure.log与sequence-watermark.log均exit_code=1/EOF；原QA全文保存为qa-execution4-fail.md，原SHA/路径映射见最终manifest。研发将失效水位保留在跨普通/SSE的health identity外侧，临时失败将已派发序列全部失效；清样本/TTL/后续成功不擦除水位。拒绝诊断invalidated_by_failure有中文说明。

本节点原两个overlay断言不改，实测旧成功invalidated_by_failure，故障后池0；八组真实HTTP同跨池/已存在成功组合确认故障前在途不能越过冷却，之后正常请求可重新accepted且旧结果不能覆盖恢复。结合新候选固定门禁，满足关闭条件，不沿用研发自签。

**tests6冷却测试时间依赖（测试缺陷，关闭）**：原10ms墙钟在诊断/记账完成后可合法过期，断言误把合法恢复当失败。研发仅给原用例注入既有atomic clock，第二次固定故障后9ms、第三次恰10ms，仍10ms配置/3调用/原故障备用冷却恢复断言。独立静态diff确认没有删断言、用例/阶段或产品实现；本节点PassiveCooldown和新完整固定均Pass，旧失败日志不删除。

静态审查配置默认/CAS事务、扩展schema兼容/损坏拒绝、TTL逐样本和EWMA、优先级/准入前置、选择快照不可变、完成后采纳/失败拒绝、代次、SSE有效输出与flush、受限typed诊断及scope查询。差异审查原件见diff-review.txt及reviewed-repository-files.json。未发现新的阻断产品缺陷；有限受控测试不能证明所有生产交错或供应商输出变体。

## 7. 限制、过程失败与留存

- 本节点命令开始前PATH不含Go，仅显式加已有/usr/local/go/bin，Go1.25.6；Node22.22.0/Playwright1.59.1/Chromium147与固定依赖一致。没有提权、安装产品后台或更改固定环境。
- QA首次用了不存在的/usr/bin/time，exit127，测试未启动；随后两次overlay准备路径/patch格式错误亦未启动Go测试，保留missing-time-command.log、missing-overlay-command.log、overlay-path-command.log及平台工具原日志。改用Bash time并修正证据overlay路径后执行一次真实负例；不是产品红后盲重试。
- QA归档审计脚本先误把研发31件改动清单当完整171清单，后误认旧log退出码大小写，三份审计失败原log保留；修正为“前改动清单+其余HEAD基线”比较与原exit_code标签，未改任何产品/测试断言/冻结阈值/原始样本。最终integrity.log exit0/EOF。
- 动态Go/UI命令并行等待冷依赖和C编译，包耗时与整命令墙钟分开记录；整命令不可相加当QA净耗时。当前QA最早留存动作约14:40:43Z，实际执行、等待、核验与归档均计本节点驻留，不宣称Agent净活动时间。
- 新固定378原件共44272250bytes（198 PNG、162 JSON、17 JSONL、1 SVG）分8个tar.gz，每分片<8MiB；fixed-browser-archive-manifest.json映射每个原路径/bytes/SHA与分片。新临时目录解压后378/378逐字节摘要验证；原文件未移动删除。不是声称释放后远端重新下载已经完成。
- 所有源码副本使用.txt或.patch.txt且位于harness-output；没有源码副本加入编译树。新16截图、2个JSON、QA命令stdout/stderr至EOF/原退出码、摘要、原失败和冻结材料通过最终manifest列具体文件，单文件/数量/总量再核验。
- 旧研发handoff-manifest引用的harness-output/qa.md是执行4 Fail。覆盖当前标准报告前已原字节另存qa-execution4-fail.md；最终保留映射和两者摘要。旧清单仍历史原件，不能按新报告重写旧签字或谎称旧清单当前路径仍同hash。
- **Not Run**：真实供应商集成/性能（无凭据）、生产负载/多进程分布式、任意供应商私有输出类型、未经测量的线上成本保证。未读取历史聊天密钥。
- **Not Run（隐私）**：原始浏览器trace，不可靠脱敏就不采集；安全结构化回执与截图替代。未留存认证header/session/密钥、提示与回复正文/工具参数，不把raw trace说成“已脱敏通过”。截图人工抽查不是对全部二进制图像进行自动秘密检测的证明。
- **后续Not Run**：正式approval/report/publisher草稿PR、远端精确功能head与QA/报告绑定、沙箱释放与全新下载归档核对。不得合并、部署或改用户预览环境。

## 8. 计时与下游交接

沿正式用户回执：Issue创建11:14:20Z；Run受理11:17:27.847529522Z；prepare结束/development开始11:19:02.534241469Z/11:19:02.534319500Z。设计两次wait/恢复、17020握手前故障恢复和Builder介入见冻结execution-v2及正式输入；不能把等待/Builder核验从总体等待中抹去。当前固定tests8完整起止如上；本QA实际命令有单独UTC/real/package回执。approval、报告、发布及释放核验尚无完成回执；**Issue到真实PR总墙钟待真实PR生成，不能以当前QA结束替代**。原生token/缓存/费用本节点无可得原值，Not Available，不估算。

交接approval：核对本候选源码6316…与当前QA报告，然后按正式report/publisher创建仅草稿PR；报告标题须反映已交付的可选低延迟路由与选择诊断，类别【产品功能】并引用Issue63。发布files只使用下列已核46件仓库改动（QA/报告由标准publisher复制），不把workflow-output/harness-output原始临时输入当业务源文件；原日志/截图/样本通过本Run产物具体清单访问。产品摘要之外的文档也按最终bytes/SHA绑定。

## 9. files（已核验实际仓库差异，非部署授权）

- `docs/workflow/runs/ba434da8039cca16f6829481112e3d68/budget-baseline-update.md`
- `docs/workflow/runs/ba434da8039cca16f6829481112e3d68/execution-v2.md`
- `docs/workflow/runs/ba434da8039cca16f6829481112e3d68/execution.md`
- `docs/workflow/runs/ba434da8039cca16f6829481112e3d68/experiment-results.md`
- `docs/workflow/runs/ba434da8039cca16f6829481112e3d68/experiment-v2.md`
- `docs/workflow/runs/ba434da8039cca16f6829481112e3d68/experiment.md`
- `docs/workflow/runs/ba434da8039cca16f6829481112e3d68/frozen-input.md`
- `docs/workflow/runs/ba434da8039cca16f6829481112e3d68/implementation.md`
- `docs/workflow/runs/ba434da8039cca16f6829481112e3d68/latency-comparison.svg`
- `docs/workflow/runs/ba434da8039cca16f6829481112e3d68/plan-manifest-v1.json`
- `docs/workflow/runs/ba434da8039cca16f6829481112e3d68/plan-v2.md`
- `docs/workflow/runs/ba434da8039cca16f6829481112e3d68/plan.md`
- `docs/workflow/runs/ba434da8039cca16f6829481112e3d68/rework-qa63-01.md`
- `docs/workflow/runs/ba434da8039cca16f6829481112e3d68/rework-tests6-cooldown.md`
- `internal/relay/routing_latency.go`
- `internal/relay/routing_latency_recovery_test.go`
- `internal/relay/routing_latency_storage.go`
- `internal/relay/routing_latency_test.go`
- `tests/browser/latency-focus.mjs`
- `tests/browser/latency-journey.mjs`
- `tests/latency-experiment/main.go`
- `tests/latency-experiment/main_test.go`
- `tests/latency-ui/main.go`
- `tests/latencyfixture/fixture.go`
- `tests/latencyfixture/fixture_test.go`
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
- `internal/relay/routing_test.go`
- `internal/relay/server.go`
- `internal/relay/store.go`
- `internal/relay/upgrade.go`
- `internal/relay/web/app.js`
- `internal/relay/web/index.html`
- `internal/relay/web/workspace.js`
- `tests/browser/journeys.go`
- `tests/browser/journeys.mjs`
- `tests/browser/workspace.test.mjs`
