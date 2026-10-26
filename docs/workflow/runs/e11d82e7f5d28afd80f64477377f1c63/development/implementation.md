# Issue65 交付接续：研发同版准备记录

Run：e11d82e7f5d28afd80f64477377f1c63；development执行2；生成时间：2026-10-25T15:59:50.832Z。

## 结论与实际改动

本节点仅核对并保留Issue63原Pipeline已验候选，准备交给本Run固定tests；没有重新实现产品，没有重写PRD、架构、验收基线或冻结计划。未发现需修改产品的新实质问题。旧QA通过仅为输入，不代表本Run测试或QA通过。

- 当前HEAD：65cc0ad786123a7e38c6118ea450db42b7bc7ddf；父提交精确为产品main 7709fe2734c370238a49bbb98efac9eddba81c7a。
- 本地源码171件，按tests/evidence的原路径、NUL、字节长度、NUL、原字节有序计算，SHA256：6316a232322de48fdce31a4420a872495f9e7ea2237c36e9e2742a4e3566a7e0。source-manifest.json列出每个原路径、bytes/SHA；source-snapshot.txt仅文本归档，不参与Go编译。
- 本地标准工具包SOURCE.json记录来源643c128dfbde94b41aa6a2ba4c0a20d944c036c4；150件逐bytes/SHA及Git blob核对，均与来源清单及7709fe2中的文件一致；另核对SOURCE.json自身，共151个对象。没有执行发布器，不能据此宣布真实publisher已验证。
- 原delivery-handoff.json列48件，加其自身共49件，精确等于相对产品main的真实差异；逐bytes/SHA/Git blob一致。不是相对prepared HEAD的空diff。publication-files.json保留完整49件路径及摘要，包含原.trellis/spec/backend/routing-contracts.md，绝不迁移或省略。
- plan-v2、experiment-v2、qa-first-no-go、qa-recheck已按入口给定SHA逐件匹配，原件不改；旧首次拒绝及两轮返工记录保留。所有原tracked文件保持未改。
- 实际新增仅此development目录的准备记录、audit.log、audit.json、source-manifest.json、source-snapshot.txt、publication-files.json、go-source-evidence.log和artifact-manifest.json。未更改产品、测试、锁文件、忽略规则、业务数据、权限或凭据。环境无apply_patch工具/可执行文件；无源码补丁，新增文件均为检查生成的归档证据。

## 本节点检查与证据口径

实际检查命令包含git rev-parse HEAD/HEAD^、git diff --name-only 7709fe2734c370238a49bbb98efac9eddba81c7a HEAD、git hash-object/rev-parse逐blob比较，以及Node标准库文件摘要核验；完整检查结果在audit.log到AUDIT_EOF exit=0，结构化结果在audit.json。另直接执行项目原Go入口go run ./tests/evidence，完整输出及原退出码见go-source-evidence.log。这是源码身份检查，不是单测/完整门禁或供应商测试。未安装依赖、未预热隔离candidate；初始工作区无node_modules及.cache/npm。

本地harness-output初始为空，旧原始212件最终/671件累计归档与原负例脚本不在本节点工作区；149149552bytes下载核验属于用户提供的上游事实，本节点没有重新下载或独立验证。原报告中提及的外部原始日志不能冒称本轮已持有。QA/报告需从正式留存获取并绑定旧证据，或明确不可用范围；新固定门禁与QA实际操作必须独立产生新原始回执，不能用旧报告替代。

## 固定tests交接（本Run尚未执行）

- 唯一完整入口make verify；本节点不预先自跑完整门禁。固定Connector负责host-only执行，显式将已存在/usr/local/go/bin加入本进程PATH；不手装依赖、不改权限、不读取旧聊天密钥。Makefile内原verification-deps依锁正常执行，不是手工预热。
- 内部25分钟/1500000ms，外层30分钟/1800秒。保留全部原阶段、冷隔离起点、source_before/after、stdout/stderr全文至EOF、原退出码、所属进程组清理与不存在证明。无/usr/bin/time可使用Bash time。不要抹掉失败、盲重试、增限或弱化断言。
- 原tests8主gate1224.995491秒仅供成本参考；加10%为1347.495040秒，小于1500。没有新增固定阶段、源码或产品执行成本，不以设施保守重复成本阻塞本次执行；并非通过保证。
- 唯一2480请求实验不变：2240核心+160心跳+80尾部；80%/35%/75%与40次适应门槛保持。普通首body、SSE首有效输出增量口径分开，不称TTFT或完整生成收益。
- 如失败，保留完整证据并定位，实质修复或新可复验证据后才返测；若同版重复外部失败需修环境/凭据/冻结准备，则正式blocked，不循环求绿。

## 独立QA、审批与最终发布要求

固定tests之后仍须独立QA：当前源码与完整门禁、原迟到负例、普通/SSE同池跨池失效恢复、两策略非零RPM/TPM/次数/并发/金额预算及隔离、SSE取消不重放、1280/390真实Go服务浏览器旅程。2480原始样本独立复算与实际重新发送请求分开报告；保留结构化逐请求/日志/费用核对和安全截图。raw trace因隐私Not Run；真实供应商、生产性能与线上费用因无凭据Not Run。

QA正式通过后必须approval，由Builder核对版本证据并正式确认，随后report和固定publisher。本节点不push、不建PR、不发布；最终只到产品draft PR，不合并、不部署、不更改原预览。推荐标题为“【产品功能】支持同模型上游低延迟优先路由与选择诊断”，主关联Issue63，Issue65仅为接续记录。

report.json.files须包含publication-files.json中的49件及本轮必要交付文档，标准publisher另复制最终QA/报告；harness-output/workflow-output仅为归档输入，不直接当产品文件提交。真实PR后核对精确远端head、全部171源码、QA/报告/清单及原证据同版，再归档、正式release、全新下载核验。不能宣称跨Run自动恢复或在途升级已实现。独立QA/待审批/发布失败等状态须由授权节点在Issue说明影响和下一步。

## 时间、历史失败与费用

全过程总墙钟从2026-10-25T11:14:20Z计，不以本Run或Issue65重置。原Run受理11:17:27.847529522Z，QA9归档15:02:20.553507846Z，approval停止15:05:47.043649725Z；原Run未有产品PR/报告/发布。旧tests3虽绿被QA4拒绝、tests6真实exit2及QA63-01旧负例exit1均不能抹除；tests6由9/10ms受控钟保留原断言修复，源码匹配证明本次沿用最终修复而非重新实现。

原阶段驻留秒数（用户输入，含等待/工具/归档，不称Agent净耗时）：prepare94.686355、development2 5603.005297、tests3 1239.498280、QA4 960.103246、development5 1808.653143、tests6 263.748324、development7 898.255873、tests8 1242.746796、QA9 1382.007475。并行源码修复与QA不重复相加；设计修订、有界握手前恢复/预算等待、Builder修复与分支准备、共享Runner排队都纳入全过程分析。

原8个原生turn累计计数（用户已核）：input40078725，cachedInput38891136，output214330（reasoning58969），total40293055；cached为input子集、reasoning为output子集，不再次相加。此为原生上报计数，非费用发票；新Run另记，本节点无原生token/货币费用回执，不估造。原受控普通/SSE稳态快端50%→90%、均值改善52.5077%–52.6653%、p95约101ms及约10%探索慢尾仅为旧输入效果，本轮以新测试与QA为准。
