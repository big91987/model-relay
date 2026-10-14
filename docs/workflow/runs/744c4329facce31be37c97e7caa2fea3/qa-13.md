# 独立 QA 返工复核（执行 13）

## 结论：PASS，允许进入 approval；不是整条链路完成

2026-10-14，Issue #44，Run `744c4329facce31be37c97e7caa2fea3`，冻结 Workflow revision2。独立 QA 实际核对 development11 修订版、固定 tests12 原始项目日志及 state、完整交付清单与研发交接前基线，并在共享工作区独立复查必要行为。本版满足本节点交付文档与测试设施验收，QA10-D1 缺章问题已关闭。正式操作者确认、停止/接续去重、最终报告、新草稿 PR 及批准文件的发布字节核验仍待后续节点，不能据此称 Harness 链路完成、产品功能完成或真实供应商验收。

本节点只新增独立 QA 报告与证据，不修改产品、测试设施、交付说明、数据、原清单、预算、环境资源、冻结图或权限；不安装依赖，不重跑全量 make verify，不提交、push、创建 PR、合并或部署。定向检查使用已有 `/usr/local/go/bin/go`；Go 定向命令自行按锁下载模块，不是手装环境依赖。QA 命令日志完整保留合并 stdout/stderr、UTC 与退出码，不冒称分别保存的两路日志。

## 版本和同版依据

- 固定 HEAD：`1d581652736fbac4e8862b70367181c8a9f2f8b2` 加未提交候选。不能用 HEAD 单独识别交付版。
- 163 项源码摘要：`991c79ca919b3d0ccf148c066c3b24d993090fcb78ab864c0e4637cfff8040ba`。tests12 前后及本 QA 新运行 `go run ./tests/evidence` 均一致。
- development11 当前 214 项完整清单：`docs/workflow/runs/744c4329facce31be37c97e7caa2fea3/development-11-correction/SHA256SUMS`，SHA256 `1ce553146e21a910a47454d9c4cf6efa83a8527e19711c41ca7bc46e3801fdb3`。QA 定向复查前后逐文件重新读取核验，全部匹配；清单自身未变。具体 UTC、每项期望/实际摘要见 `qa-13/integrity-before.json`、`qa-13/integrity-after.json`。
- 当前交付文档 SHA256：`3d881a85cbeaa5fcec0c72de2568a143988e7bc967bfc0f2e3f1cffa0185aff8`；首轮原快照仍为 `7792bfaafcb7202f1f2ac5901a1672a078cec18eac5445705dc3bf8e0707184b`。
- QA 同时逐字节核对 handoff 绑定的研发基线 JSON、空 stderr 和退出状态文件，三摘要均吻合。研发基线实测窗口为 18:06:56.095Z–18:06:56.123Z；不是固定 runner 采样。

**同版等效依据判断：充分用于本 QA 复核。** 实际顺序是研发基线及正式交接 → 唯一固定 tests12（18:07:56Z–18:24:51Z）→ 独立 QA13；交接后至 tests 结束没有 Agent 修改交付物，tests 为冻结执行节点而非 Agent。研发基线逐项与 QA 后验同一 214 项清单一致，固定 runner 又核验自身隔离副本与工作区的 163 项源码前后不变。此分角色链路满足 QA10 允许的可审计等效依据，不是把源码摘要冒充完整交付清单。

固定 tests 仅执行冻结的 make verify，**没有执行额外 sha256sum 命令，也没有本 Issue 硬编码清单采样**。接续收到维护者通过本会话正式投递的只读旁证：18:08:57.909Z–18:13:14.034Z（两端seq12 running）已逐读清单中213个可见文件、6,493,619字节且摘要一致，AGENTS.md被其API策略排除、未绕行；清单自身前后与研发baseline.json原件也匹配。这是维护者报告的已完成运行期旁证，不是QA自行通过该API读取的原件；不称维护者已核全部214项。QA实际再次独立读全214项及三个基线文件均匹配，见 qa-13/maintainer-followup-integrity.json。维护者完成后第二次核验仍进行中，不提前当成完成依据。此旁证增强既有充分等效链路，独立判定不交给维护者；来源与边界另存 qa-13/maintainer-readonly-supplement.md。该判断也不证明所有文件在每个瞬间从未被写后恢复，更不替代最终 PR 的逐字节核验。新 QA 文件不在研发 214 项清单内，另以本节点 `qa-13/SHA256SUMS` 绑定。

## 固定 tests12 原始结果

主 attempt：`20261014T180756Z-2810915431`。完整项目 verify.log **1,155,178 字节**经程序读取，SHA256 `9856236d5b7cdd9e31970e46dc3a5020d9e166c3659f25626c366abfdb23dc82`，精确匹配 state。原字节备份为 `qa-13/tests12-verify.log`、`qa-13/tests12-state.json`。完整日志分析、原阶段、资源重算及图像摘要核验见 `qa-13/fixed-tests-audit.json`。

QA对正式 seq12 留存输出另实际读取首页和末页，末页 EOF=true、truncated=false；外层执行回执 exit0、1800秒、1,184,959字节。QA没有逐页遍历平台事件包裹日志的中间部分，**不声称QA自己已读平台全文**。接续维护者报告其已通过正式output API分页读取外层全文至EOF、truncated=false，SHA256 `704b03399b155dd851d66388b6c9ac10b3aa95545a9080aaa6b546ad3ca9f8f5`；该全文摘要按维护者来源记录，不冒称QA自己的全文核验。通过依据为完整项目日志、匹配 state、同版核验和真实外层回执；项目与平台日志格式及摘要不同，不能混用。

主 gate configured=1,440,000ms、effective=1,439,999ms、execute=998,870ms、exit0/cancelled=false。主 browser 阶段依次全部 ok=true：migration 15,738ms、legacy-ui 205,421ms、routing 235,595ms、console 36,198ms、full-console 239,821ms。完整日志保留原历史重建、health-history、格式、模块锁、vet、前端、race 全集、build 与 browser 门禁。日志中的原历史预期红和内嵌验证器 exit23 反例属于通过的分类/失败路径测试，不能误判为主 attempt 失败。

state 的新隔离候选 node_modules/npm cache 初始不存在、NODE_PATH unset、manual_preinstall none；这是候选冷起点，不证明全局 Go/系统缓存为空。清理为 candidate_removed 且 owned_group_absent=true；不能据命令 sandbox.stopped 推断整个 Run 或环境已释放。

本次真实 Linux 起止资源为 18:07:56.253647561Z–18:24:51.318704952Z，同 cgroup 身份；cpu.max 200000/100000，memory.max 8589934592，runtimeCPU112/GOMAXPROCS2。QA 重算全部差值：usage_usec 1,371,516,988，throttled_usec 87,762,430，OOM相关差值0。只代表共享可见 cgroup 的区间计数，不是单进程独占、CPU唯一归因或性能 SLA。

实际产物检查：1280/1440/390 的 first-self-hosted JSON 均为 passed、各12步、controlled_http_not_supplier；审查原生动作和独立临时 SQLite 声明。134份 PNG 原文件摘要均与其 integrity.json 相符。本 QA 没有重新人工逐屏视觉审美验收，也没有再次运行完整浏览器旅程；完整 UI/API 行为通过依据为固定 tests12，而非静态 browser.check。

## 测试设施源码独立审查

原候选三源码保持原字节，分别为：

| 文件 | SHA256 | 审查结论 |
| --- | --- | --- |
| tests/review-history/main.go | eaa988821fccf011d9fc63795ea5933f3f7acb734ffe22830cbff8f6eafcf355 | stdout严格JSON，完整stderr独立保留；非法stdout、真正执行错误、缺失/重复历史断言及race/build/listener仍拒绝。120秒与原四个失败名、两次零预算及回滚断言保留。 |
| tests/review-history/main_test.go | ebdd0554efd7f73155324f9d5034e71a72fe1baaced6747c066485f5443a8bfa | 隔离命令夹具覆盖冷下载stderr、非法stdout、非预期退出、意外绿、缺断言及诊断异常；完整stderr首尾哨兵保留。 |
| internal/relay/routing_http_test.go | 0018956b4e78f792ad26b789b215e66c31a75bf12a7f683e4ba9eb2249ae8572 | 真实HTTP、40并发SWRR 30/10、503回退/计量、映射和凭据哨兵保留。一小时原期限、未到期排除、管理API决定记录与已到期恢复严格核验。 |

路由阶段边界持锁核验同一健康实例后，仅调整测试自有内存 health.Cooldown 至 Observed 前1纳秒，构造已到期输入，再通过真实HTTP验证；不是自然等待一小时、产品可控时钟或数据库改值。其余允许列表、最多三候选、禁用后回退、已提交SSE不重放用例仍保留。

新增/修订设施审查覆盖 tests/verify/main.go、main_test.go、gate.go、resources.go、gate_resources_test.go，以及 tests/browser/journeys.mjs、stream-capture.mjs、stream-capture.test.mjs：

- 累计总预算在项目原入口固定24分钟并写入同版receipt，上限不可用环境变量放宽；更短父截止优先。原顺序、请求/单步期限、断言、失败退出及SIGTERM/SIGKILL进程组有界清理没有放宽。预算是冷门禁累计额度，不是产品响应 SLA。
- Linux采样读取或解析失败明确失败；采样失败通过 errors.Join 保留原gate错误，不能把失败刷绿。合法同cgroup/限额及不倒退计数才生成差值。
- 非Linux记录GOOS、runtimeCPU与unsupported，不读Linux文件、不编造指标或零差值，不锁死既有macOS入口。既有范围依据为 .github/workflows/deploy-local.yml 的 macOS ARM64 runner 与 docs/local-deployment.md 的完整make verify；隔离Darwin分支测试不等于macOS实机验证。
- SSE设施克隆同次原生fetch Response，返回原Response，读取真实SSE体，finally恢复fetch；无额外dispatch、制造帧、重试或sleep求绿。原请求stream、HTTP状态、唯一dispatch、Request ID、UI结束、data/[DONE]和账本断言保留。

## QA 自身必要行为复查

2026-10-14 18:26:49Z–18:29:10Z，未再跑全量：

- `go test -count=1 -v ./tests/verify ./tests/review-history`：exit0。真实100ms短预算/父截止子进程取消与清理、正常/非零/非法预算、SIGTERM持久回执、隔离Darwin不可用、Linux真实ENOENT、合法差值和采样失败保留exit23，以及历史分类反例均通过；完整输出在 `qa-13/go-tests.log`。
- `node --test tests/browser/stream-capture.test.mjs`：exit0，5项通过；原Response身份、单次dispatch、缺采集/读体失败/传输失败及恢复断言均保留；见 `qa-13/node-tests.log`。
- `go test -race -count=1 -v ./internal/relay -run '^TestRoutingRealHTTPSelectionFailoverAndRecovery$'`：exit0，包耗时1.959秒；真实临时HTTP、隔离持久化fixture，无race；见 `qa-13/routing-tests.log`。
- `go run ./tests/evidence`：exit0，源码摘要一致；git diff --check exit0；固定日志/资源/134张图像摘要审计exit0，完整214项交付及基线前后核验exit0。

局部测试不独立构成产品全量通过；上述必要复查与同版固定tests12共同支撑本节点结论。QA没有重跑旧研发三个变体反例或另一次实际历史HTTP重建；历史HTTP本次依据固定tests12，变体历史证据仅按原报告标注。

## 原失败、实际拒绝与返工关闭

development02 明确标注授权缺章注入，首轮原文确实没有“停止、恢复与环境释放”；首轮快照和 QA10 拒绝原文字节均保留且包含在214项清单。QA10 真实 REVISE，不是研发代写或协调者代验。development11 在该拒绝之后才补齐完整章节。

QA13 实际读新章节：安全停顿与Run停止/测试取消的区别、未知结果先核原执行、同Run/同版接续和去重、材料保存后平台释放、释放后不可恢复旧环境、旧Run已释放与本Run待确认边界均明确；未编造操作API、点击或未来验收。QA10-D1 影响已消除，关闭。文档明确按研发11时点记录未知未来；QA13更新事实保存在本独立报告，不为更新状态再修改已测试文件。

执行5原exit2/12分钟总量超期（非Owner stop、非业务失败）、旧24分钟提案/撤回、纠正和后续授权、兼容性返工、执行7 SSE采集设施失败、执行8修复、执行9原绿、QA10缺章拒绝、development11修订/边界纠正、tests12新版绿均保留。旧Run未全量通过/未QA的事实不改。原旧报告私有harness-output不全部可读，不声称本工作区有其原始日志。

## 未验、限制及后续要求

- 真实供应商、支付、外部身份、macOS实机：NotRun。没有性能SLA验收或新业务功能交付。
- 平台安全停顿点的停止/接续去重、整个Run取消/环境释放尚未由本QA实测。approval须由已授权操作者通过正式平台入口保存前后身份、文件清单与去重回执；如实称“操作者确认”，不能称用户本人点击。环境释放应在材料可访问保存后由平台执行，不删除工作区代替。
- 原生会话实际模型配置、Issue机器人全链路回写仍无独立QA回执。维护者运行期213项只读核验及平台日志全文读取按正式输入旁证记录，QA未自行抓取其API原件；维护者完成后第二次核验尚未结束，未宣称已完成。
- approval、report.json、主分支目标真实新草稿PR及同版逐字节发布仍未完成。report须覆盖相对真实main的全部实际变化，包括本候选三源码、预算/资源/SSE设施、文档和证据，附本QA原文；不能只报告新说明。当前仅有本地候选分支，未独立解析最新远端main或验证最终PR差异。
- 新QA报告/证据以 qa-13/SHA256SUMS 绑定，与研发214项清单一起用于审批及发布核验；后续如修改批准范围文件，不能沿用此QA结论冒称同版。不得合并或生产部署。

产物前缀：`docs/workflow/runs/744c4329facce31be37c97e7caa2fea3/`。本报告由独立QA13编写，正式handoff approval，不替人工确认或报告角色完成其职责。

接续补充：首次approval交接因本条用户输入待投递未被接受；现已处理该输入，原QA报告与231项清单保留为历史版本，当前报告及更新审批清单取代它们。新增证据后原214项候选仍不变，不需要因QA旁证记录重跑全量。
