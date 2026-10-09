# 研发验证与事实分类

执行8当前结论见 [tests7完整退回与调度整改](host-return-exec8.md)；旧078e1aa1候选完整make verify失败、两屏冷却/routing通过。新候选仅测试调度改进，完整宿主/独立QA未验，不能称产品修复或门禁已绿。历次结果保持。

执行6最新结论见 [宿主退回归因与复验记录](host-return-exec6.md)；以下保留执行4原结果与首失败。执行5取得旧候选受控实际HTTP/race证据但完整门禁仍失败，不能把下文历史Not Run当作未曾执行，也不能将旧版通过转为新版绿。当前版本见candidate-source.json，旧版见candidate-source-exec4.json。

日期2026-10-09；Run `637705186406b80a2ee4afaf0d6b2167`；结论 **Ready for host tests / 非产品Go**。研发责任产物已完成，原#33 No-Go不变，独立计划评审/QA未执行。

## 1. 版本与材料

- 原受测版/当前Git HEAD `3f9cee1615e87e002a2c2300f2d29778072ec3f5`。本节点未提交/推送；当前是其上新增测试和规范的工作树，不伪称新commit。
- 基线工作树摘要 `0384f115adbbd545e8ce01021f238010ba66c89eb06e6acea08ce872ce0fa474`；最终候选摘要 `ebfc78a9bc25c69e3bacffe21f4aab956fd70d4d9af07dc12a8c97f7b61c7a9d`，由项目tests/evidence实际生成，见candidate-source.json，最终race日志头包含同一值。其范围是产品/测试/锁/契约/Trellis规范，故不受后续结果文档影响。
- 测试源码摘要：console_contract_test.go `0939ab1f06cd079f00b1129853fe01147c2c88d84a8e325a8571dc3fc51ea4ac`；metering_regression_test.go `e29ec00ebe3064597d4399ade21548297f983f2d9e8fe7dded76500823c44c9d`，见candidate-test-code.sha256。后续Connector提交后应重新核对工作树摘要与候选HEAD，不用dev版本字符串代替制品定位。
- 材料版本、实际路径、ZIP摘要及61项manifest核验见README；原design-v0.2.0字节保持。原ZIP参考摘要 `5f51d35fdc15ed7214749f9413879b4238097f1d6b6bbe2cc0d508d8e103cef7` 是原材料声明；本节点实际重新核验的是此次a13f交接ZIP及其文件manifest，不冒称重新下载原ZIP。

## 2. 报告事实、候选与未验项逐类处置

| 输入类别 | 复核/复现结论 | 处理责任与证据 |
|---|---|---|
| #33桌面/390本地16拒绝、同ID零attempt | 保留历史已测子集，不在本研发重演UI | 原材料report/independent-review及两屏证据；不是RPM/TPM通过 |
| 调用form没有输出/思考字段 | 历史页面事实不反驳，但不推导后端无默认 | 不改猜测控件；正式模型bounds→admit→boundedPayload→上游payload已补可执行用例 |
| DS-SAFETY-01广义缺安全路径候选 | **内部契约不复现“后端没有默认”**；当前真实HTTP/DeepSeek仍Blocked，不能关闭为全面不复现 | Safety三种有限触发、普通/SSE实际受控handler收到64/32；thinking仍unsupported_field，真实推理收费语义不获保证。已有技术契约明确保证只适用于可信有界模式 |
| RPM/TPM生产实现已坏 | 原报告没有确认，此次也未确认 | 新用例按原AC/架构数值复核正确的准入/结算/共享池；只改测试资产，不宣称修好未证故障 |
| PLAN-01原文缺失 | 本Run已解除，旧Run的Blocked不改写 | 61项manifest再次校验通过，原AC09–13/16和实际实现同规则 |
| PLAN-03/04/05边界、窗口 | 内部race通过；自然分钟真实网页Not Run | RPM2/N+1、TPM U/P/B等值与少1、unknown跨窗、核对；服务端测试时钟明确不是客户端等待或真实时间验收 |
| PLAN-06普通/SSE与PLAN-07失败/取消 | 内部对应分支通过；真实供应商成功/收费未运行 | 普通/SSE known/missing/inconsistent、500known/unknown、timeout、取消、已受理未dispatch；ActualHTTP尚待宿主（首次选择用例启动即bind失败） |
| PLAN-08/09并发和消耗池 | 内部屏障/阻塞上游矩阵通过；双浏览器会话Not Run | 8竞争者只有1槽、7拒绝先读回零attempt；双Key总和/稳定Key新旧值、独立tenant。生产HTTP/浏览器不能由此推导已通过 |
| PLAN-10守恒 | 现有handler回归通过；真实进程门禁待宿主 | 改额/轮换/暂停/时钟、三attempt费用/差额；原RealProcessCrashRecovery/ClockRecovery保留，不替代为内存测试 |
| PLAN-11数值采证缺口 | 新回归已保存白名单数值，不为旧attempt补造 | METERING日志及test-plan字段约定；服务端GET无minute字段，测试窗口来源标清；真实网页采证仍由专用Pipeline实现/评审 |
| PLAN-12与FLASH缺名/预算未知 | 仍外部前置Blocked，不是产品故障 | 精确名称不改；没有凭据/真实收费及推理上界/历史预算权威账；新供应商生成0，未把历史unknown归零 |
| 真实普通/SSE成功、纠正后成功恢复、真实Token/费用对账 | **Not Run / Blocked** | 交专用测试Pipeline合法配置入口；本项目受控价/Token账不是供应商账单。完整网页原范围由测试责任方独立评审/复测 |

详细逐条数值及可执行入口见test-plan.md。本轮没有可以提交“产品故障首红→修后绿”的已确认缺陷。下面驱动错误均不计生产缺陷。

## 3. 实际命令与回执

环境Go1.25.6/darwin-arm64，Node26.0.0/npm11.12.1，原生workspace-write沙箱。所有业务数据属测试临时安装，未使用供应商/用户数据；日志中的2026-10-01是注入的UTC计量测试时钟，不是执行日期。

| 命令 | 实际结果 | 文件 |
|---|---|---|
| make verify（基线） | 外层make退出2；go wrapper1；内部gate2；fail-fast于review-history历史HTTP监听，后续门禁未跑 | baseline-make-verify.json、baseline-make-verify-full.log；原始test-results/verification/20261009T114858Z-733498716也保留 |
| go test -race -count=1 -v ./internal/relay -run '^TestMetering.*InProcess$'（首次） | 1；新驱动错误读code而API是error_code | metering-inprocess-first.log；不是产品红 |
| 同入口第二次 | 1；错误把所有400归同码；且buffered mock不能到达真实下游first-output屏障 | metering-inprocess-second.log；protocol.go明确0为invalid_request、thinking为unsupported_field；helper传输边界已区分，不猜测产品修复 |
| Safety/Cancellation针对驱动复验 | 0 | metering-driver-corrections.log |
| 第三次较早完整进程内矩阵 | 0；当时尚未新增undispatched/timeout两行 | metering-inprocess-final.log，保留而不覆盖 |
| Undispatched/Timeout新增分支 | 0 | metering-interruption.log |
| 最终同摘要完整TestMetering进程内矩阵 | **0，8组，race无报告，60.658秒** | metering-candidate-final.log，含command、Go版本、完整candidate摘要及test_exit=0 |
| 既有Governance/ResourceMatrix/SharedResources/Accounting/Pause/SoftBudget/ClockRollback/MicroDollarPrecision进程内回归 | **0，25.243秒**，保留原fixture默认断言 | existing-inprocess-regression.log；之后仅wrapper test.Helper归因修正，无生产行为变更 |
| format-check / go mod verify / vet / build / go test ./... -run '^$' / git diff --check | 全部0；compile-only不是运行全部测试 | static-candidate-final.log；较早static-checks.log也保留 |
| 静态采证首次shell | 1；zsh的modules是只读变量，format/module已经完成，后续尚未执行 | static-checks-first.log；改结果变量名后复验，不计产品故障 |
| 当前候选最小监听诊断：go test -race -count=1 -v ./internal/relay -run '^TestMeteringSafetyActualHTTP/tenant_tpm$' | **1，bind: operation not permitted**，在启动受控上游时失败，零生成 | listener-diagnostic.log含同candidate摘要；明确环境限制，非安全默认断言失败 |

当前环境不重跑相同必然失败的完整make verify；宿主必须对**新候选**实际跑完整原入口。监听故障不能当门禁通过，当前HTTP/进程/浏览器除上述setup诊断外均Not Run。保留Darwin linker的LC_DYSYMTAB警告，未将其当失败原因（静态/进程内成功、监听诊断明确EPERM）。

首次失败中的storage_write_failed出现在新驱动Fatal后未排空fixture worker的清理过程；并发回归已补有界排空，未按这些fixture清理日志另立生产故障。取消InProcess最终明确first_output_observed=false，实际HTTP用例才在client首行后cancel；该HTTP断言尚未运行，不用模拟证明网络。

## 4. Trellis检查、裁剪与安全

实际完成trellis-before-dev packages discovery和backend/guides的具体规范阅读；现有规范非空且有源码引用，未伪称prepare新初始化，也不重建规范。按trellis-update-spec新增7段metering-regression契约，更新index。trellis-check核对变更、范围、复用、imports、格式、编译、vet/race及副作用；完整project check诚实Blocked。

生产backend/UI/API/schema/依赖不变，只新增测试与可接力事实；不扩展到3+生产层，跨层检查按实际管理配置→admit→payload/dispatch→settle→同IDdiagnostic链核对。未读/复制New API，不执行材料脚本，不改用户SQL，不使用静态browser.check冒充后端验证。为此单一整改保留一任务/一里程碑，未建新PRD/设计或组织运营文档。

公共日志只遮蔽Darwin编译器临时宿主路径为<temporary>，保留原错误、退出码、数字及摘要。完整未脱敏日志仅在忽略的本地缓存，公共源/脱敏SHA映射见evidence-index.json；baseline完整日志与原state中的SHA一致。秘密、完整对话、数据库、缓存、绝对本机路径均不加入公开产物。

## 5. 交接退出条件

tests收到test-plan.md先独立审查AC断言及分层，再在宿主对精确candidate运行make verify（全部旧门禁保留），保留真实回执/失败并由独立QA判定。宿主成功不等于DeepSeek或费用验收。真实专用Pipeline须核准精确模型、账/推理上界和历史预算后才新生成；前置未知就Blocked，不能由本节点替代或扩大授权。

本节点不启动Git发布/部署；下一步Ready为宿主测试与独立评审，而非版本可发布。无需要用户再次确认的普通选择，也无冻结设计实质冲突需要返回requirements/design。
