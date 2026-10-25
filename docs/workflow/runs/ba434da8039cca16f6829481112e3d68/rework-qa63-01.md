# Issue63 QA63-01 实质返工与预算停止回执（development 执行5）

Run `ba434da8039cca16f6829481112e3d68`，2026-10-25 UTC。产品候选已实质修复；**研发定向验证通过，固定重测尚未执行，独立QA尚未复验，因冻结v2预算停止条件暂停交测**。这不是QA签字、功能最终通过或发布许可。

## 输入、保留与实际改动

先读取根AGENTS、冻结输入、获17021批准的plan-v2/experiment-v2、实施记录、完整QA63-01说明及两份原始失败。未重建工作区、修改数据库、冻结阈值、QA断言或旧证据。v1/v2九件bytes/SHA重新验证，旧失败原件保留，详见返工manifest。

根因：临时失败delete普通/SSE评分池，池内Sequence随之丢失；原acceptLatency仅在评分池存在时拒绝旧序列，故较早派发的成功能重建故障恢复池。此前固定摘要 `9f7a656a749ef67af8604af68a04cced5fdc31d368ac2e61005dd4bf9ffcfd66` 的绿色不能用于新候选放行。

实际返工共六个仓库文件：
- `internal/relay/routing.go`：健康状态保留独立于评分池的 `InvalidatedThrough`。临时失败在同一server.mu临界区记录当时dispatchSequence水位，并清两模式样本；所有当时已派发的同健康身份请求均失效，包括派发序列高于失败请求、但早于失败被观察的请求。后续健康成功保留单调失效水位；失效在途成功不能清冷却。
- `internal/relay/routing_latency.go`：通过配置/执行代次检查后、评分池顺序检查前，拒绝水位内成功；持久化 `invalidated_by_failure`，而非accepted或未知。普通/SSE共用健康身份失效边界，各自评分仍分池。其他失败、取消、TTL、配置迟到、存储降级检查保留。
- `internal/relay/web/workspace.js`：新增该拒绝状态中文标签“故障前在途观测已失效，不采纳”。
- `internal/relay/routing_latency_recovery_test.go`：八组真实HTTP/race：普通/SSE旧请求×普通/SSE失败×有/无较新已采纳成功；barrier额外覆盖比失败请求序列更高、却已在失败前派发的在途成功。逐请求读取持久化证据，断言清池为0、旧成功拒绝、不触达冷却上游、正常新请求恢复采样、恢复成功后仍拒绝最老请求、另一模式新请求也能学习。
- `tests/browser/workspace.test.mjs`：执行真实enumLabel逻辑，区别失效、成功与未知。
- `.trellis/spec/backend/routing-contracts.md`：补单调失效/恢复及诊断契约，不改冻结产品目标。

现有healthKey按模型别名、上游ID、上游执行版本与远端模型隔离。失效水位不随评分TTL或采样清空丢失；重启本就终止在途请求且不保留学习评分。配置变更仍由原Epoch/Revision拒绝迟到结果。没有新增持久化表、供应商探测、重放或通用策略框架。

## 真实定向命令与结果

均为研发执行，不冒称独立QA。只使用隔离临时数据及环回临时端口，不触达用户上游。全部原回执至EOF/原退出码见 `harness-output/issue63-rework-1/` 具体清单。

|回执|实际内容|结果/墙钟|
|---|---|---|
|watermark-regression.log|首次调用不存在的/usr/bin/time，Go测试未启动|exit127，保留工具失败，不称产品失败|
|watermark-regression-run.log|go test -race -count=1 -v ./internal/relay -run '^TestLatencyRealHTTPFailureWatermark$'|exit0，八组4.76秒，包5.842秒，整条46.759秒；含首次模块下载/编译|
|latency-and-health-regression.log|go test -race -count=1 -v ./internal/relay -run 'TestLatency\|TestRoutingHealthEpochSequenceDisableAndRestart' -parallel=4（竖线为正则或）|exit0，包36.970秒，整条37.851秒；覆盖原延迟、16并发配置代次、迁移、冷却恢复、SSE取消/终端写失败、存储失败、low_latency准入矩阵及新增八组|
|late-after-failure-fixed.log|原QA overlay.json及TestQA63LateSuccessAfterNewerFailure原断言不变|exit0，5.946秒；旧成功invalidated_by_failure，池0|
|sequence-watermark-fixed.log|原QA watermark-overlay.json及TestQA63SequenceWatermarkSurvivesFailure原断言不变|exit0，6.190秒；较新成功accepted→503 not_success→最旧成功invalidated_by_failure|
|static-frontend-vet-build.log|gofmt检查、git diff --check、node --test tests/browser/*.test.mjs、go vet ./...、go build -trimpath实际Go程序|exit0，前端286/286，完整命令91.440秒，含既有阶段冷编译，未提前跑make verify|
|source-evidence.log|go run ./tests/evidence|exit0，新全源码171件摘要7f98fee14c6ca5337d981b66b0d7dd1ed88579123c492674c7ed1ed3da8f69e0|

原QA两份exit1/EOF原件、qa.md及旧完整固定回执未覆盖。新测试自动进入既有make test（go test -race ./...），没有跳过开关、削减断言或更改25分钟/外层30分钟。未再次执行2480请求实验或双屏UI；这些必须由新候选固定入口和后续独立QA重新绑定，不复用旧结果声称新候选通过。

## 冻结预算核算与正式阻塞

严格沿用experiment-v2 §7逐项max及原980.081+98.0081漂移公式，而非自行用本轮旧绿色放行。原runner236、双屏100仍保留（研发原实测195.802/15.588秒未被当新候选性能结果）；纯状态/race20、前端5保留。保守HTTP记最长实际命令46.759秒（同组全量定向37.851秒另记，不叠加重复覆盖），编译/vet/build保守把整条91.440秒全部计入，虽含旧阶段及前端而未扣除，以免低估。

`980.081 + max(236,195.802) + max(25,46.759) + 20 + 5 + max(20,91.440) + 100 + 98.0081 = 1577.2881秒`，超1500秒 **77.2881秒**。即只计37.851秒HTTP同组仍为1568.3801秒；结论不依赖是否把首次下载算HTTP。故不能交固定节点盲试。新静态命令的91.440秒包括完整原有vet/build，并不能据此声称产品增量就有91秒，也不能称完整门禁实际已超时；这是冻结保守预测触发线，不是新固定失败。

建议等待Builder正式处理预算估计口径，而非扩大时间/删覆盖/预热candidate：已有同Run真实固定主gate1215.138100秒（完整原回执），可作为更直接成本基线；保守再加本轮两条最长完整定向命令46.759+91.440秒和基线10%漂移，得 **1474.85091秒**、余25.14909秒。这一替代仍有风险，并且把很多已存在成本再次计入；它是待正式批准的预算模型建议，不是已获许可的v2修改或运行保证。确认前保持阻塞，不重跑求绿。全部固定阶段、2480请求规模、原效果阈值、25/30分钟不变。

## 候选、隐私与后续

新全源码摘要见上文；31件功能改动完整快照、相对旧候选六文件差异、原路径/bytes/SHA和原失败/冻结输入校验均单独归档，源码快照为.txt，不进入Go构建输入。固定入口与QA尚未绑定新候选；旧QA仍Fail历史原件，不由研发覆盖为Pass。新候选没有真实供应商凭据；供应商/生产性能和raw trace仍Not Run。

环境：本节点初始PATH不含/usr/local/go/bin，显式补标准Go工具路径后使用go1.25.6；/usr/bin/time缺失，改用Bash time，保留原exit127。原生apply_patch入口不存在，以临时同名薄包装调用系统patch应用统一diff；工具包装在/tmp而非产品中。未读取凭据/用户会话、推送、创建PR、改权限、发布、修改预览或代签QA。

计时：首个带时间回执13:33:37Z；新实际命令起止/耗时各自留存。节点驻留/冷依赖/等待不称Agent净耗时；Issue创建与Run受理等平台原时刻继续沿用既有实施记录。后续等待、固定、独立QA、approval/report/publisher及留存核验尚待真实事件，不估token或费用。
