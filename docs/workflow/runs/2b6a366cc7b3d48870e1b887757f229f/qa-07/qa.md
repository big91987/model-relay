# 独立QA执行7：文档修复确认，证据门禁未完成 / revise

2026-10-14；Issue #48；Run `2b6a366cc7b3d48870e1b887757f229f`；冻结revision3。审查版本 `development05-lifecycle-rework`；HEAD `30b4e13c19e36d51b72ef1186468d6adf39ac822`。**拒绝进入approval，交development补证；不宣称产品或完整Pipeline草稿PR交付通过。** 本轮没有修改产品、测试、模板、锁文件、数据库、预算或冻结准备，没有发布/合并/部署。

## 验收依据和返工结果

依据本Run冻结用户任务1–7以及QA04原拒绝记录，必须保留首轮缺章、独立拒绝、研发修订、每轮固定make verify完整内外层原件及同版独立QA，再进入正式控制/确认/报告/固定发布。旧tests12/QA13、旧Run17步与#47材料接续不代替本Run结果。当前真实main `6d2552d331fd7a0ec8d205f5d18ba3bcd34d6995` 是HEAD祖先；相对main的继承改动仍须由下游报告覆盖，不是本轮新产品设计。

QA04-D01的文档缺章已实质修复：当前文档12999字节、SHA-256 `d1a7112600467ae85d9af9000e3349f78ee9590e0317fc935cd4e1560d5f4209`，全文七个二级章节；原六章作为原字节前缀完整保留。首轮快照摘要仍为 `35e5eb5af37a83c81344cbad95cb89b866193179605ba5b9c772dc2dc1562be9`。新增章节明确正式approval安全点、stop与重复stop实际去重证据、同Run同节点resume及材料摘要保持、decision与平台已保存结果、保存后正式生命周期释放，禁止删除目录代替释放、不恢复旧取消Run、不把操作说明当完成事实。研发05快照与当前文档原字节一致；QA04原报告和研发05摘要清单检查通过。旧完整opensandbox-run-sharing.md未改。

源码163项重新逐项核字节/大小并重算组合摘要，仍为 `991c79ca919b3d0ccf148c066c3b24d993090fcb78ab864c0e4637cfff8040ba`；cmd/internal/tests/.trellis/spec无额外构建输入；146项模板逐文件匹配bc43314标准导出。跟踪差异为空。既有无网络check_publish_files全仓扫描通过，未修改扫描规则。Git凭据变量仅检查存在布尔值，GITHUB_TOKEN/GH_TOKEN/GIT_ASKPASS均false，未读取值。

## 本Run固定tests06内层与实际独立复查

固定入口仍为原 `make verify -> go run ./tests/verify -> 隔离候选make verify-gates`，未手装依赖、弱化断言或替换全套。平台传入执行6：exit_code=0、sandbox exit=0/stopped=true、外层1800秒、log.bytes=1177433/truncated=false；这些是正式传入元数据，不等于已完整审阅外层。

完整读取并程序化核验内层原verify.log的4592行、1147466字节，SHA-256 `61dadc32700ffe860c9a03800b30e13b6c406f2a56dd43bedc1d71c14c3ba797`，与state一致；两份原件按原字节归档并保存源路径映射。state开始19:55:23.561967759Z、结束20:12:24.157026612Z；内层预算1440000ms、gate execute_ms=1004725，exit/gate_exit均0且未取消。候选冷起点node_modules/npm cache均false、manual_preinstall=none，source_before/after一致、candidate_unchanged=true、owned_group_absent=true、candidate_removed。共享工作区不是全局冷起点。8Gi/2CPU及GOMAXPROCS=2来自state采样，无OOM增量，不能作独占资源因果归因。

设施/日志核对包括格式、go mod verify、vet、前端fail=0、当前Go race包、构建、实际Go服务migration/legacy-ui/routing/console/full-console阶段ok=true。日志25处`--- FAIL`保留；当前顶层测试没有FAIL，历史HTTP重建明确记录预期21项旧方法红灯/12项查询恢复通过，runner故障注入由通过的TestFixedRunnerReceipts/TestInventoryCancellationReceipt覆盖，不删历史红灯。

QA另外实际运行6项runner/资源/拷贝边界回归及3项受控真实HTTP回归：TestHTTPManagementAndOrdinary、TestSSEBarrierAndUsage、TestConsoleActualHTTPQuotaBillingAndSSE，全部通过，两条命令退出0。实际临时HTTP服务器/临时存储、SSE首片屏障、usage及管理凭据边界来自已有测试设施，不触用户上游。仅限定回归，不冒充固定make verify，不是静态browser.check，也不代表真实供应商。命令/完整输出/退出码已归档。

本QA首次检查调用python（不存在）退出127；改用已安装python3后检查退出0。首次限定回归shell找不到go，两条退出127；定位已安装/usr/local/go/bin/go，仅纠正PATH后原限定用例通过。以上失败原件保留，没有安装工具或改测试。Go按锁定go.mod下载模块为测试正常依赖解析，不是手工预装固定候选依赖。清除了本QA导入扫描模块生成的两份pycache；未删运行环境或其他材料。

## 阻断证据缺口 QA04-E01（仍OPEN）及 QA07-E01

- 复现：研发05明确保留QA04-E01 OPEN；原tests03外层1174629字节仅曾读取首段0→32768与尾段1146880→EOF，区间[32768,1146880)缺完整审阅和原字节归档，内层文件不能冒充外层。QA04报告/研发05原元数据未覆盖。
- 当前tests06：本QA实际原生读取offset0→32768（eof=false）及offset1146880→1177433（eof=true），原工具均truncated=false；中段[32768,1146880)仍未完整审阅，完整外层原件也未物化。尾段资源/清理/完成事件核对无冲突，但**尾页EOF不证明连续完整读取**。当前内层完整原件不关闭该缺口。
- 诊断：按照仓库SDK公开路径对原seq3做无凭据只读HTTP请求，得到401；没有读取/申请凭据，没有改权限。原生分页工具可用，不能据401宣称所有访问途径不可用。不存在完整原件时不重建/拼凑省略文本，不宣称原输出丢失。
- 影响：用户明确要求每轮完整原回执、退出/冷起点/摘要/清理且QA04已经提出补证；当前无法确认已完成该证据门禁。文档修复和局部绿灯不能覆盖外层原件缺口，禁止推进approval/publish。
- 返工要求：由研发/平台原有授权入口保全seq3及seq6完整原始脱敏日志，以正式逐页连续读取（保留offset/next_offset/EOF覆盖）或正式原字节导出归档；分别匹配1174629和1177433字节，保存具体源Run/seq、映射、摘要、原退出码及timeout元数据。原日志/历史失败/QA04原件不可覆盖；QA须继续完整独立审阅，不由研发代QA结论。若完整原档确实缺失或截断则正式记录不足，不用重跑替代旧回执。**证据补齐本身无需同版再跑make verify；不改产品/测试/凭据门禁/预算。若候选实质变更，才重新固定tests并QA。**

## 未验、限制与后续边界

未复跑浏览器全套，只核对本Run固定tests06完整内层和限定独立HTTP；未验证真实供应商、支付、外部身份、Mac实机。平台afebd27、4小时/40步和原生角色模型会话声明来自冻结上下文，未独立验证全部配置。外层原件门禁仍未过；approval安全点实际stop/重复stop/resume/材料保持、正式decision、report、真实草稿PR及远端同版核验、保存后正式环境释放均未执行，不能当本QA缺失证据的替代或提前声称完成。

本报告仅保存在qa-07/qa.md；当前harness-output/qa.md仍保留QA04原拒绝以维持其原摘要清单，不把它冒称本轮最终放行。以后通过的QA才写最终harness-output/qa.md，并记录历史移动/映射。归档仅文本日志/JSON/脚本.txt/清单，没有可被Go递归编译的源码副本。所有路径是尚未发布的仓库相对文件，不是GitHub可访问链接。

## files（本轮具体新增QA产物）
- `docs/workflow/runs/2b6a366cc7b3d48870e1b887757f229f/qa-07/SHA256SUMS`
- `docs/workflow/runs/2b6a366cc7b3d48870e1b887757f229f/qa-07/archive-mapping.json`
- `docs/workflow/runs/2b6a366cc7b3d48870e1b887757f229f/qa-07/checks.exit.txt`
- `docs/workflow/runs/2b6a366cc7b3d48870e1b887757f229f/qa-07/checks.json`
- `docs/workflow/runs/2b6a366cc7b3d48870e1b887757f229f/qa-07/checks.stderr.log`
- `docs/workflow/runs/2b6a366cc7b3d48870e1b887757f229f/qa-07/checks.stdout.log`
- `docs/workflow/runs/2b6a366cc7b3d48870e1b887757f229f/qa-07/files.json`
- `docs/workflow/runs/2b6a366cc7b3d48870e1b887757f229f/qa-07/independent-checks.py.txt`
- `docs/workflow/runs/2b6a366cc7b3d48870e1b887757f229f/qa-07/independent-regression.commands.txt`
- `docs/workflow/runs/2b6a366cc7b3d48870e1b887757f229f/qa-07/independent-regression.exit.txt`
- `docs/workflow/runs/2b6a366cc7b3d48870e1b887757f229f/qa-07/independent-regression.log`
- `docs/workflow/runs/2b6a366cc7b3d48870e1b887757f229f/qa-07/initial-checks.exit.txt`
- `docs/workflow/runs/2b6a366cc7b3d48870e1b887757f229f/qa-07/initial-checks.stderr.log`
- `docs/workflow/runs/2b6a366cc7b3d48870e1b887757f229f/qa-07/initial-checks.stdout.log`
- `docs/workflow/runs/2b6a366cc7b3d48870e1b887757f229f/qa-07/initial-regression.exit.txt`
- `docs/workflow/runs/2b6a366cc7b3d48870e1b887757f229f/qa-07/initial-regression.log`
- `docs/workflow/runs/2b6a366cc7b3d48870e1b887757f229f/qa-07/outer-receipt-metadata.json`
- `docs/workflow/runs/2b6a366cc7b3d48870e1b887757f229f/qa-07/qa.md`
- `docs/workflow/runs/2b6a366cc7b3d48870e1b887757f229f/qa-07/tests06.state.json`
- `docs/workflow/runs/2b6a366cc7b3d48870e1b887757f229f/qa-07/tests06.verify.log`
