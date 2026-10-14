# 独立 QA：首轮拒绝 / revise

2026-10-14；Issue #48；Run `2b6a366cc7b3d48870e1b887757f229f`；Workflow revision3；QA 执行4。审查版本 `development02-authorized-missing-lifecycle-section`，HEAD `30b4e13c19e36d51b72ef1186468d6adf39ac822`。**本轮不放行 approval；交 development 返工。** 原 make verify 的退出0不等于本任务验收通过，也不等于产品或完整草稿PR交付通过。

## 验收依据与实际改动

依据本节点用户冻结任务：新文档必须包括项目自有模板安装、共享工作区/角色交接、固定验证与独立QA返工、正式确认、发布失败处理、材料保存、停止/恢复/环境释放；首轮授权故意缺章，但必须由本Run独立QA真实拒绝后研发补齐，再经固定tests及独立QA复核。旧 tests12/QA13、旧Run17步、#47材料接续均不是本轮结果。

独立读取当前文档、研发执行2记录、首轮快照、checks、delivery-files、frozen-identity、源码清单与 Makefile/runner/历史HTTP验证设施；核对未跟踪交付文件及跟踪差异。未修改产品、测试、锁文件、模板或旧完整文档。源码163项按路径排序/NUL/字节长度/原字节算法独立重算：`991c79ca919b3d0ccf148c066c3b24d993090fcb78ab864c0e4637cfff8040ba`；146项模板逐文件摘要匹配 `bc43314eca15f5af042afd9e3d212469eabd10ea`；main `6d2552d331fd7a0ec8d205f5d18ba3bcd34d6995` 是HEAD祖先；跟踪差异为空，研发原 SHA256SUMS 校验通过，旧 `opensandbox-run-sharing.md` 与HEAD原字节一致。

## 阻断缺陷 QA04-D01：停止、恢复与环境释放缺章

- 位置：`docs/harness/opensandbox-publish-recovery.md`，第45行起“材料保存与历史关联”直至文件末尾；全篇只有六个二级章节，无“停止、恢复与环境释放”章节，也无对应操作说明。研发记录第5行起明确这是首轮授权故障注入。
- 复现：读取当前文档并枚举所有Markdown标题；与首轮快照作原字节比较，两者一致，SHA-256均为 `35e5eb5af37a83c81344cbad95cb89b866193179605ba5b9c772dc2dc1562be9`。独立脚本 `qa-04/independent-checks-v2.py.txt` 实际执行并记录 `lifecycle_section_missing=true`、`qa_decision=revise`，退出0表示检查成功复现缺陷，不表示交付合格。
- 影响：操作说明缺少安全点停止、重复stop去重、同节点resume、材料保持及保存后正式生命周期释放的边界。无法支持用户要求的Harness停止/恢复验收与安全释放；这是文档交付阻断，不是产品业务新增要求。
- 返工：只由研发补齐当前新文档，并明确正式入口、状态/回执核验、重复stop幂等、同Run同节点resume、不重启旧已取消Run、不删除目录代替释放、材料先保存后释放；描述操作要求与实际完成证据须分开，不伪造已执行。
- 保留：首轮快照、故障说明、研发原清单和本QA拒绝报告不可覆盖。修订另存实施记录和新文档摘要；不修改旧完整文档、不改产品/测试/发布凭据规则。

## 本Run固定tests03证据核对

平台传入外层回执：执行3，exit_code=0，timeout_seconds=1800，log.bytes=1174629，truncated=false。工作区原层 `test-results/verification/20261014T192852Z-1089328441/verify.log` 和 `state.json` 存在，已按原字节复制至本Run `qa-04/tests03.verify.log`、`qa-04/tests03.state.json` 并写原路径/字节数/摘要映射。

内层日志完整读取4588行，1144558字节，SHA-256 `afdbb354a9181d8fd950d2d16159283c9379d56552269eadbe4cd3f283ed101a`，与state一致。state开始19:28:52.320239524Z、结束19:45:49.176535584Z；gate_exit/exit_code均0，configured_budget_ms=1440000，execute_ms=1000606；候选冷起点node_modules/npm cache均false，manual_preinstall=none，NODE_PATH=unset，源码前后与候选一致；owned_group_absent=true，cleanup记载candidate_removed。辅助命令均记录退出0。资源回执记录8Gi、2CPU配额、GOMAXPROCS=2，无OOM增量；共享cgroup统计不作独占因果归因。

设施仍是原 `make verify -> go run ./tests/verify -> 隔离 make verify-gates`。Makefile包含格式、模块、vet、前端、race、build、实际Go服务浏览器门禁；浏览器设施使用临时目录和127.0.0.1临时端口的受控HTTP上游，不是静态browser.check，不是供应商证据。内层记录模块校验、vet、前端283/283且fail=0、当前race包通过、构建、migration/legacy-ui/routing/console/full-console阶段ok=true。

历史health重建的JSON输出中确有 `--- FAIL`，这是 `tests/health-history` 针对旧基线的期望红灯，日志明确 `PASS expected historical HTTP red: 21 method failures with exact status/Allow diagnostics; 12 query/recovery passes`，并非当前候选测试失败。runner故障注入包含exit23/24/25/26、取消和资源读取失败，位于通过的TestFixedRunnerReceipts/TestInventoryCancellationReceipt测试中；原红灯保留，不删除、不弱化正常测试。

**外层完整审查限制 / QA04-E01：**本节点调用 `read_command_output(seq=3,offset=0)` 得到next_offset=32768、eof=false、truncated=false；未读完后续分页，也未将完整1174629字节外层原输出物化到可提交文件。因此不宣称已逐页核验外层完整回执，不将用户消息中省略的中段或内层日志冒充外层。此证据门禁仍未完成；本轮已经有独立复现的缺章阻断，不进入确认。返工期间须保留并归档执行3完整外层原回执及正式退出/timeout元数据；后续QA须按正式日志分页读到EOF。不得用重跑替代丢失的旧回执或把历史日志当新结果。

## 独立检查与QA自身失败保留

独立检查限于只读版本/摘要、文档缺章复现、原回执字节/state一致性及既有无网络 `repository.check_publish_files`。未手装依赖，未重复全套make verify，未修改数据库，未执行stop/resume/release/decision/发布。仅检查GITHUB_TOKEN/GH_TOKEN/GIT_ASKPASS存在性，均false，未读取值；不能倒推tests进程环境的独立观测。

QA首版检查脚本把日志任意位置的 `--- FAIL:` 当作当前失败，因期望历史红灯断言退出1。该QA解释错误不是产品失败；已保留首版脚本、initial-checks stdout/stderr/exit原件，核对历史验证器及真实包通过上下文后，新版检查按当前行级失败与明确期望历史红灯分别判定，退出0。这是改正QA日志分类，未改任何产品断言或原日志。当前工具集中无apply_patch，终端也未找到该命令；仅以heredoc/标准库写QA文本和证据文件，不安装工具。

## 返工后必须重新核验

1. 研发补齐QA04-D01，保留首轮失败和本报告，另存修订摘要/实施记录，不代QA放行。
2. 保全tests03内/外层原件、退出码、冷起点、源码与清理证据；具体files清单最多100项、单文件≤5MiB。证据与构建输入分离，不产生编译器递归发现的源码副本。
3. 修订版本仍通过固定tests原make verify（内层24分钟、外层1800秒不变），保存完整新回执，不同版循环求绿；再交独立QA复核。
4. 后续QA核对本次拒绝/正式revise、修订结果、完整同版回执及候选身份后才能交approval；本报告不能作为修订后通过报告。
5. QA放行之后才由责任角色核验安全点控制、decision、report、固定publisher草稿PR和远端同版；不得Agent手动发布，不合并、不部署。

## 未验、限制与遗留

本节点未对真实供应商、支付、外部身份、Mac实机作验证；未重新执行产品HTTP/浏览器全套，只审查本Run受控门禁原件和设施并独立复现文档缺章。平台afebd27/冻结revision、4小时/40步以及各角色独立模型会话来自任务上下文，未独立观测全部配置。外层EOF审查、修订后固定测试与QA、approval安全点实际控制回执、正式确认、报告、真实草稿PR、远端一致性和材料保存后的正式环境释放均未完成，不能将本节点任务结束视为产品或整条Harness链路通过。

## files（具体QA产物）

本段所列为仓库相对路径，尚未发布；不是GitHub远端链接。研发首轮12项由其 `delivery-files.json` 独立保留，不覆写其原清单。

- `harness-output/qa.md`
- `docs/workflow/runs/2b6a366cc7b3d48870e1b887757f229f/qa-04/qa.md`
- `docs/workflow/runs/2b6a366cc7b3d48870e1b887757f229f/qa-04/tests03.verify.log`
- `docs/workflow/runs/2b6a366cc7b3d48870e1b887757f229f/qa-04/tests03.state.json`
- `docs/workflow/runs/2b6a366cc7b3d48870e1b887757f229f/qa-04/archive-mapping.json`
- `docs/workflow/runs/2b6a366cc7b3d48870e1b887757f229f/qa-04/independent-checks.py.txt`
- `docs/workflow/runs/2b6a366cc7b3d48870e1b887757f229f/qa-04/independent-checks-v2.py.txt`
- `docs/workflow/runs/2b6a366cc7b3d48870e1b887757f229f/qa-04/initial-checks.stdout.log`
- `docs/workflow/runs/2b6a366cc7b3d48870e1b887757f229f/qa-04/initial-checks.stderr.log`
- `docs/workflow/runs/2b6a366cc7b3d48870e1b887757f229f/qa-04/initial-checks.exit.txt`
- `docs/workflow/runs/2b6a366cc7b3d48870e1b887757f229f/qa-04/checks.json`
- `docs/workflow/runs/2b6a366cc7b3d48870e1b887757f229f/qa-04/checks.stdout.log`
- `docs/workflow/runs/2b6a366cc7b3d48870e1b887757f229f/qa-04/checks.stderr.log`
- `docs/workflow/runs/2b6a366cc7b3d48870e1b887757f229f/qa-04/checks.exit.txt`
- `docs/workflow/runs/2b6a366cc7b3d48870e1b887757f229f/qa-04/files.json`
- `docs/workflow/runs/2b6a366cc7b3d48870e1b887757f229f/qa-04/SHA256SUMS`
