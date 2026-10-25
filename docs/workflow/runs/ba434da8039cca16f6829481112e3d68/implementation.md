# Issue 63 实施与交测记录（研发结论，不是QA签署）

- 类别：【产品功能】；Issue：https://github.com/big91987/model-relay/issues/63
- Run：`ba434da8039cca16f6829481112e3d68`；基线 `f06f05de20ffac5b8c8e772958c808379001f0e3`。未提交、未推送、未创建PR、未改权限、未部署或更改已有预览。
- 状态：实现与研发定向验证完成，准备交固定 `make verify`；完整门禁、独立QA、approval/report/publisher仍待后续节点真实执行，不能引用旧绿色代替本轮。
- 设计基线：Builder第二轮完整性检查批准v2实施。v1/v2和两版manifest原字节保留；冻结文件中“待检查”是当时状态，不回写历史。第一轮纠正了SSE首body目标失配、长流TTL续期风险和串行预算；第二轮批准并非QA通过。
- 本记录新增，不覆盖冻结的计划、实验规模或80%/35%/75%/40次阈值。原生17020握手前失败和停止、单次恢复授权，依据Builder本会话正式输入；未取得该失败原始日志，不冒称产品命令曾执行或已复核EOF。

## 实际改动

1. 在原模型编辑页面增加持久化策略 `weighted` / `low_latency`，默认权重；旧POST省略默认、旧PATCH省略保留。策略与模型候选CAS写入同事务。目录和旧管理页面同步读写，服务重启保留策略、运行时学习重置。
2. 独立版本化routing扩展表，主schema1/2/3兼容操作入口不改；schema2/3旧库在正常open或upgrade中事务创建扩展，未知/部分扩展拒绝。无离线SQL改业务数据求绿。
3. 普通响应首body、SSE首有效输出增量分池。SSE忽略心跳/role/空/usage/结束元数据，不称TTFT。每池32条、逐条60秒TTL（测量时刻起，60秒含边界）、最低3条、EWMA .25。稳定合格同优先级集合每5次探索一次，5n机会界限；其余最低有效评分，无评分回退原权重。
4. 选择前仍执行原授权、启用、优先级、冷却及租户/Key准入；探索是正常请求，无隐藏付费探测。实际测量不等于采纳，完整成功及持久化后才入池；取消、失败、终端写失败、降级存储、配置代次变化和旧序列迟到拒绝。SSE原有已输出/取消不重放、记账和未知usage不改。
5. 同Request ID诊断永久保留当时选择和本次采纳事实（安全标量/枚举/ID），未知为null。修正过期池重复描述从expired退回unknown的问题，并保留低优先级等原排除理由。配置变更清空池，长流完成不续TTL。
6. 固定入口原阶段不删；在build之后/browser之前无条件执行一次routing-experiment。Go test只运行轻量统计单测，不重复2480采样；双屏原生UI旅程嵌入原full-console的1280/390，1440及原各阶段完整保留。新增独立开发用UI入口不在固定阶段重复运行。

## AC / 研发证据映射

下表“研发通过”仅表示列明范围的实操/定向断言，不是独立QA或完整固定入口通过。

|AC|研发结论与范围|留存回执/后续|
|---|---|---|
|01 策略配置/兼容/重启|双屏页面创建两服务、候选、低延迟策略，保存重入、重启保留、切回权重；API默认/省略/非法枚举、旧库扩展迁移/幂等/拒绝通过|ui-candidate.log、ui/*.json/PNG、new-stage-timing.log；固定旧二进制迁移旅程待执行|
|02 合格集合与原准入|低延迟10个租户/Key calls、非零RPM/TPM、并发、金额预算矩阵及原路由健康/升级定向通过；选择仍在同最高优先级合格组|candidate-edges.log、targeted-regression.log、new-stage-timing.log；完整权限隔离矩阵留固定入口|
|03 冷启/探索/互换/恢复|2480全样本、每cell阈值通过；n=1/2/3/100探索界限、普通/SSE新候选12请求各一组、冷却恢复120请求通过|experiment/*.jsonl、summary.json、experiment-results.md、new-stage-timing.log|
|04 延迟口径与心跳陷阱|普通首body、SSE首有效输出；160成功心跳陷阱、80长尾；role/空/工具/refusal/metadata分类和合法无输出真实HTTP通过|实验逐请求传输/有效输出对照；TestLatencyOutputDeltaContract、TestLatencyRealHTTPNoOutputAndNewCandidate|
|05 TTL/迟到/失败|真实HTTP两模式完成59/60/61秒（注入测量钟、barrier）、逐样本过期、16并发×两模式×策略/禁用/远端/凭据/候选移除重建；旧序列、存储失败不采纳，race无报告|new-stage-timing.log；终端强制写错误用in-process writer，真实HTTP取消/中断另见candidate-edges.log，不冒称强制writer测试是TCP实测|
|06 终态/重放/未知|本轮定向原实际HTTP终态、取消、已输出不重放；原未知用量与最多3次尝试逻辑不改|candidate-edges.log及targeted-regression.log；完整原记账/未知矩阵仍须固定节点|
|07 原准入资源|低延迟租户/Key calls、非零RPM/TPM、并发、金额预算10场景；正常实验100%成功、每请求一次尝试一次charge|new-stage-timing.log、experiment/*.jsonl；其余原权限隔离由固定完整入口复核|
|08 诊断/隐私|普通/SSE同ID、上游、选择依据、采纳状态与费用14微美元，页面实操及安全JSON断言通过|ui-candidate.log、ui/安全模态截图；raw trace未采集，不能声称trace已脱敏；完整跨租户权限矩阵留固定入口|
|09 全量受控对照|两轮两模式核心2240+心跳160+长尾80，不补样不降阈值，全部通过；分布/代价/窗口原样留存|experiment-candidate.log、experiment-results.md、summary/JSONL/SVG；固定在最终候选重验|
|10 固定/独立QA/交付|研发候选已绑定源清单；make verify本节点未跑，QA尚未执行，无QA签名|固定节点需完整25分钟门禁；独立QA必须写harness-output/qa.md及映射/真实命令/同版源码核验|

## 实际命令、失败与修复

所有回执位于Run产物 `harness-output/issue63-development/`，逐文件列在evidence-manifest.json；不是只有本机目录提示。源码以`.txt`快照/patch留存，不把Go副本放入构建树。

- `new-stage-timing.log`：`go test -race -parallel=4 -count=1 -json ./internal/relay -run '^TestLatency'`，最终新增覆盖全组实际27.193秒，命令墙钟28.147秒，exit0/EOF。
- `candidate-edges.log`：新增组加原实际HTTP committed stream/tools、metering terminal/cancel定向，包33.849秒，命令37.332秒，exit0/EOF。包含旧覆盖，预算不把此整条再加到新增阶段。
- `targeted-regression.log`：低延迟resource矩阵、原fallback/3次上限/健康代次与显式升级拒绝及原子性等实际命中用例，命令7.727秒，exit0/EOF。早期探索命令曾no tests to run，不计通过证据。
- `experiment-candidate.log`：唯一一次研发runner，2480样本、零断言失败；summary194.999秒/END195.001秒，go run墙钟195.802秒，exit0/EOF。采样全串行，无浏览器/编译/其他实验并跑。核心两轮两模式快端90%、稳态均值改善52.5%左右，互换后新快端90%；160心跳陷阱90%且改善52.77%。各cell不混均值，详细分布、学习/适应窗口、选择代价见experiment-results.md及summary。
- `ui-initial.log`：第一次UI失败exit1/EOF，因旅程使用不存在的“概览”导航；改为实际“工作台”后通过。原失败安全JSON/截图保留，非产品缺陷、非独立QA发现。增加滚动至新诊断区截图后，最终`ui-candidate.log`双屏通过，整命令15.588秒，各进程7.090/6.911秒。所有承诺配置/调用/诊断由页面完成，夹具只准备隔离身份/受控上游/测试重启。
- 静态审查发现durable finish失败仍可能留下attempt成功态而被评分；新增degraded storage保护并以真实HTTP存储失败注入回归通过。其他阶段未制造失败，未降断言、未重跑求绿；早期通过回执也保留但不冒充最终全套。
- `handoff-static.log`：gofmt检查、git diff --check、go vet ./...、Go build、共享UI JS语法和原前端单测（共285条）通过，墙钟3.984秒，exit0/EOF。没有跑完整make verify、没有预热固定隔离candidate。

## 候选与实验版本绑定

- 最终source-manifest.json摘要：`5607062a855ff7151718973ed5762ac5a83cc49fdbde2ba9be791f7847214420`，29个实际改动的产品/测试/README/契约文件，逐路径bytes/SHA。source-snapshot.txt保留全部29文件原字节；tracked-diff.patch.txt含已跟踪diff（新文件由snapshot及清单补全）。完整仓库非改动部分继承冻结HEAD。
- 实验当时source摘要`50827f37eb32171d27311a3304e92147fe6f89f67c2662140fe547ebc978f097`原manifest/snapshot单独保留。实验后仅增强`routing_latency_test.go`反例和`tests/browser/latency-journey.mjs`安全截图；实际网关、fixture、runner及依赖未变，证据核验脚本逐文件比较。**不说实验已在两个完全相同的源码摘要上执行**；固定入口会在最终候选重新执行唯一runner，完整门禁/QA必须绑定最终候选而非研发旧测试文件摘要。
- 冻结计划及所有结果文档另行清单，避免结果生成令可执行候选摘要自引用。后续若源码变更，必须重建对应candidate绑定；未经固定和QA确认不发布。

## 交测预算（v2第7节）

|新增项|冻结规划秒|实测依据秒|入账秒|
|---|---:|---:|---:|
|runner|236|195.802（cells/overhead单列）|236|
|新增HTTP边界|25|28.147（保守把含纯状态/race/编译的整个最终新组命令计入）|28.147|
|纯状态/race|20|已含于上一实测，仍额外保留规划|20|
|前端|5|最终整条静态命令3.984、原前端单测1.613|5|
|新增编译/vet/build|20|静态整条3.984，非冷candidate保证|20|
|双屏UI增量|100|独立开发入口含准备整条15.588|100|

新增保守409.147秒；历史主gate980.081秒仅成本参考，合1389.228秒；再加原阶段10%漂移98.0081秒＝**1487.2361秒**，25分钟余**12.7639秒**，未达>=1500阻塞线。budget.json保留公式与数值。原阶段/固定完整full-console实际仍待测，余量很小，不能把估算称通过；独立开发UI不是原full-console增量精确差分，保留100秒规划而不以15秒替换。外层30分钟不借给25分钟主gate。任何真实失败留完整EOF/退出码，只有实质修复或新证据才重交，不静默增限/删原阶段/弱断言/加跳过开关。

## 计时来源与后续

以下时刻来自Builder转交的平台/GitHub原回执，不是本节点重新查询结果，不是Agent净活动：

|事件|UTC|
|---|---|
|Issue63创建|2026-10-25T11:14:20Z|
|Run受理|2026-10-25T11:17:27.847529522Z|
|prepare结束|2026-10-25T11:19:02.534241469Z|
|development2开始|2026-10-25T11:19:02.5343195Z|
|首次wait accepted / native stopped|11:26:21.257601192Z / 11:26:30.171867724Z|
|16636后同一步恢复running|11:30:37.032639151Z|
|第二次wait accepted / native stopped|11:43:05.507155016Z / 11:43:17.301582802Z|
|本次研发实验窗口（本节点真实日志）|12:32:51.678978614Z 至约12:36:06.680Z|

入口opened40188019710/labeled40188020080 success且同Run。Issue到受理187.847529522秒，prepare阶段驻留94.686711947秒，prepare命令92.450秒是不同口径；后续设计等待/Builder核验/握手恢复不能计作Agent净耗时。最终Issue到PR总墙钟尚无真实PR结束点，由report/publisher补齐排队/准备/设计/实现/固定/QA/返工/正式确认/发布/留存核验；原生token/缓存/费用不可得，不估算。

供应商/生产性能均Not Run；没有凭据、不带历史密钥；没有独立QA结论或完整make verify回执；旧设施PR62仅基线，不复用其绿证据。固定测试后还须独立QA真实关键旅程/负例/版本复核，正式approval→report→publisher只创建可审阅draft PR。Builder后续精确head/产物核验、释放及全新下载检查亦尚未执行，不能预报成功。
