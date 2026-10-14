# development 执行 8：固定 tests7 失败定位与 SSE 测试采集修复

## 身份与范围

Issue #44；Run 744c4329facce31be37c97e7caa2fea3；Workflow revision2；HEAD 1d581652736fbac4e8862b70367181c8a9f2f8b2，加既有执行6未提交候选和本次三个 browser 测试设施文件。未重建工作区。执行前核对 AGENTS.md、frozen-identity.json、既有兼容记录及129文件清单；原清单全部吻合。frozen-identity.json 是已有白名单身份记录，不冒充完整冻结参数文件。保留原三候选源码、五份预算/资源设施源码、所有历史失败/提案/撤回/纠正记录及首轮故意缺章。

只修改 tests/browser/journeys.mjs，新增 stream-capture.mjs 和 stream-capture.test.mjs。没有修改产品业务、持久化数据、Makefile、预算、资源、冻结图、权限或锁文件；没有读取凭据值、提交、push、PR、发布或合并。真实供应商、支付、外部身份 NotRun。

## 完整失败证据与分类

先读取 seq7 正式回执首段及末段到 EOF；不是声称分页读完正式1170997字节。另完整读取/解析本地项目主 verify.log（1146109字节），SHA256 efd55d375788bd9620b35af4e370b5cd498a74325df22427839d648a383f40a9，与主 state 的 log_bytes/log_sha256 完全相符，末尾 FAILED exit status2。两个日志格式不同，不混用长度或SHA。主 state/完整项目日志原字节保存在 development-08/；正式回执 truncated=false、外层1800秒、exit2。

主 attempt 是 20261014T170210Z-2990963391。gate预算1440000ms、执行847083ms、exit2、cancelled=false、owned_group_absent=true；不是24分钟总超期、Owner stop、资源OOM或历史预期红。主 gate起点17:02:25.367922988Z，至migration前约314991ms；browser阶段migration15360ms、legacy-ui195875ms、routing224273ms、console31165ms均成功，full-console65144ms失败。此前完整console宽390成功，full-console宽1280中的J01–J07、visual和summary成功，第一次self-hosted旅程在self-hosted-sse/proxy-call失败。HTTP200且dispatch_delta=1，普通同ID日志/账本已经通过；原receipt仅error_kind=other，没有具体异常，不能据此称业务断言失败。

真实资源采样范围17:02:10.628927095Z至17:16:33.256719471Z：cpu.max=200000 100000、memory.max=8589934592、runtime_num_cpu112、GOMAXPROCS2；共享cgroup增量usage_usec1100252962、throttled_usec56440223、nr_throttled1506，memory.events所有差值0。不是单进程CPU归因或唯一根因。原state还记录冷起点node_modules/npm_cache均无、NODE_PATH未设、无手工预装、源码前后相同e637c8f9…、候选清理和进程组消失。固定tests7原失败永远保留为失败。

## 隔离最小实证与修复

在研发自有临时副本，保留真实Go服务、临时SQLite、临时端口、原用户动作/请求/原断言，仅添加一个定向Go test调用fullConsoleJourneyMode(1280,self_hosted)，以及响应读取异常的静态布尔诊断；不输出异常原文或凭据。实际退出1，12.469s：response-body、stream=true、status200、Error、protocol_error=true、no_data_found=true、no_resource=false。说明本次可复验边界是Chromium/Playwright对该SSE响应的协议体读取失败；这是复现证据，不声称原seq7未记录异常文本与复现逐字节一致。

本次隔离诊断只使用预装Go1.25.6、Node22.22.0、平台已存在且与锁相同的Playwright1.59.1及Chromium目录，无安装命令。临时候选node_modules链接至既有平台依赖，浏览器目录显式用于隔离定向诊断，Go构建缓存位于自有临时目录。这是暖定向实证，不能替代原冷隔离make verify，也不是新增运行Profile、固定tests环境特批或资源改动。

原call仍通过真实界面点击发送，用同一Playwright响应核验状态/stream参数/Request ID、真实dispatch差值和界面完成状态。仅成功SSE的原协议response.text读取改为：在测试页面一次性包装原生fetch，克隆同一个真实Response并读取实际流字节；原Response原样返回UI，不二次请求、不制造data/DONE、不改正文、不重试/sleep。不捕获其他管理读请求；成功、断言失败或取消均在finally恢复fetch并删除临时采集状态。克隆读取失败或无采集结果明确失败。仍保留原data:/DONE、同Request ID成功状态、key/tenant/model、pending、3/4 tokens、单笔fee21/cost7、账本/价格快照和最终6/8/42/14聚合断言，单请求/单步期限和所有gate顺序不变。

## 本版定向结果

- 五个新Node单测退出0：同一Response/唯一dispatch/原生消费者和克隆字节一致/恢复；缺失采集；读取失败；无关请求；真实transport rejection均明确核验。
- 对修后同版browser源码，仅重新执行同一真实1280首次self-hosted旅程，退出0、15.686s、12步通过，provider=controlled_http_not_supplier。其他宽度及完整make verify由固定tests负责，本轮没有代跑全套。
- 在同一自有副本将controlled upstream仅删除DONE帧，执行同一定向旅程一次，退出1、12.287s、stage=self-hosted-sse、error_kind=assertion、HTTP200、dispatch_delta1。证明原DONE断言仍真实拒绝错误流，而非通过移除/伪造帧求绿。变体不在交付源码，补丁和原fixture原字节保存。
- 新source-evidence退出0：163文件、源码摘要991c79ca919b3d0ccf148c066c3b24d993090fcb78ab864c0e4637cfff8040ba。此摘要不包含全部交付文档，交付版另由SHA256SUMS绑定。
- node --check journeys.mjs及git diff --check通过。没有声称本版全量通过、macOS实机、性能SLA、独立QA或供应商验收。

准备诊断时一次命令cwd错误、apply_patch不可用，后使用现有patch进行文件编辑；临时测试最初尚未生成就运行，返回no tests to run，保存为setup-no-tests，明确不计通过。新Node测试初次生成缺少尾部闭合造成SyntaxError并已补齐，保留initial-syntax输出；只将修后真实五测试和实际Go旅程作为本版定向证据。不会把这些准备失败隐藏成测试绿。

## 同版交接与尚未完成

development-08/SHA256SUMS绑定相对固定HEAD的所有现有未提交修改/新增交付文件，加原三候选源码、入口及锁。继承文件基线由明确HEAD锚定；不是只用163文件产品摘要代替交付材料。SOURCE-BASELINE.txt保留完整Git树身份列表（不读取凭据内容）。源码快照均以.txt保存，避免证据副本变成新的根模块测试包。原首轮文档与11项清单未变，“停止、恢复与环境释放”仍故意缺章，未进入QA不得提前补齐。

现有项目设施有实质最小修复和新红/绿/反例证据，因此正式handoff tests一次原make verify；内部24分钟一次校准、外层1800秒原样。保存完整stdout/stderr、退出码、主attempt身份、冷起点、Linux真实资源前后/差值、源码/完整交付前后摘要及清理。不要复用本暖定向绿为完整验收，不手装依赖或同版盲重跑。若24分钟仍总超期则正式blocked按新证据决策，不再自动加时；其他新失败先完整定位。

固定tests通过后，独立QA须审查本次3个browser文件、既有5个预算/观测文件、原3个候选源码、归档模块边界、原断言和取消/清理完整性；区分SSE协议体采集修复与真实业务断言、共享cgroup观测与性能SLA。必须实际拒绝首轮缺章并正式revise；研发之后才补章并再次完整tests/QA复核。当前无全量通过/QA/正式操作者确认/报告/新PR。
