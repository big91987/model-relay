# 执行21：PR35 + PR36 联合受控候选研发复核

日期2026-10-10；Run637705186406b80a2ee4afaf0d6b2167；Issue34 / PR36。结论 **Ready for host tests，非联合Go**。本轮不新增需求/产品功能，不修改供应商或冻结设计；确认生产缺陷仍0。

## 版本、恢复与材料

prepare18保留日志953字节，read_command_output读取至eof=true、truncated=false。分支为workflow/637705186406b80a2ee4afaf0d6b2167；HEAD `8917f942cb38af2f56aa93914ab62e763bf9cc6e`；origin/main与MERGE_HEAD均 `6929376089535038112da48aed715fd729615d76`；共同基线 `3f9cee1615e87e002a2c2300f2d29778072ec3f5`。prepare已受信整合，integration_pending=true，无unmerged路径。研发接手时3255项暂存、三份状态文档未暂存、执行17阻塞文档未跟踪；没有清理或覆盖这些输入。

材料版本product-test33-handoff-r1，ZIP实算SHA256 `a13fb6b9089cefbb9d2b6f9a333830e0ee6866bd7ff45d5ba881da97e9a66a49`；61项bytes/SHA256全部匹配manifest。实际路径见README及[integration-inventory-exec21.json](integration-inventory-exec21.json)的material项。报告、独立复核、DS-SAFETY-01、两份RPM/TPM计划实际在material_root/test-report/下，原prd/interactions/acceptance在material_root/design-v0.2.0/docs/下；不能用遗漏test-report前缀的简写定位。沿用原材料，不执行包内脚本。

Trellis0.6.15及源码支持的非空backend/frontend/guides规范实际存在，packages discovery为single-repo。prepare.initialized=false是此次未重新初始化；不声称本轮初始化，不需要bootstrap。按trellis-before-dev读取索引、计量/运行时/控制台/依赖验证和相关模态规范；按trellis-check核查两边差异、重用、边界与验证。此次只维护交付证据，无新可执行产品契约，不需改写Trellis事实源。

## 联合保留核对，不以无冲突代替验收

[只读采证脚本](collect-integration-exec21.mjs)可从仓库根运行，读取Git树、工作文件及材料，不写.git、启动服务或调用网络。它对3255个main变更路径及231个PR36变更路径逐文件采集大小/SHA256并比较Git blob：PR35全部逐字保持；PR36仅README/T001/M01公共状态文档不同，全部源码、原回归、失败证据逐字保持。库存包含大于1MiB的JSONL，不依赖截断输出。

首次临时诊断使用execFileSync默认1MiB上限，将过大的JSONL读取异常列为差异，未据此修改文件；最小归因确认大小超限。提高只读诊断缓冲至32MiB复算，main差异0、PR36差异仅三份状态文档。正式脚本使用ls-tree加本地blob摘要，无逐文件子进程输出上限；该修正是采证驱动修正，不是生产修复。完成公共文档维护后另有verification/test-plan.md的联合说明差异，最终库存显式列出这四份允许更新的交付文档；原矩阵与所有源码断言不变。

项目 `go run ./tests/evidence` 记录158文件源码指纹 `6e359667d5f7521f6b8c421454c6fc7cbe5ccb8c74aa87bb1e5955c194c202cc`，见[candidate-source-exec21.json](candidate-source-exec21.json)；正式库存列出同范围每文件SHA256并复算一致。范围为产品、tests、构建锁、项目契约与Trellis，不含交付回执避免自引用；两边所有变更路径另有完整库存。不是新提交SHA，不用旧2dfe14指纹放行。

Makefile与两边版本相同：verify→Go tests/verify隔离冷副本→verify-gates（source/history/health/format/modules/vet/锁定frontend/Go race -parallel=4/build/真实Go浏览器）。十二分钟原deadline、所有原断言和依赖锁不变。metering_regression_test.go全部保留并由Go race wildcard纳入；cooldown诊断及其测试保留。PR35的workspace.js/css、readability/capture-integrity、self_hosted旅程和runner计时测试逐字main；仅测试隔离父级原Parallel调度来自两项原改动，未新改生产准入、费用或取消语义。

安全路径仍为正式模型bounds→admit→boundedPayload→上游cap；有限TPM/硬预算触发时注入64或保留合法显式32。RPM/TPM独立语义、U/P/B、未知预留、池共享/隔离及数值断言见原test-plan A。未复现“后端没有默认值”，不能凭表单无思考控件制造产品改动；受控上游证据不证明DeepSeek收费/推理约束。

## 实际检查与失败保留

Go1.25.6 darwin/arm64、Node26.0.0、原workspace-write沙箱。以下日志中的私有工作区、home和临时路径作等值标记脱敏；退出码、错误类型、数值和用例不变。

| 实际命令/检查 | 结果 | 证据 |
|---|---|---|
| 只读联合库存及61项材料校验 | exit0；valid=true；两边源码完整保持 | integration-inventory-exec21.json |
| go run ./tests/evidence（GOCACHE=项目.cache/go-build，GOTOOLCHAIN=local） | exit0；158文件；与库存一致 | candidate-source-exec21.json |
| make format-check / go mod verify / go vet ./... / make build | 各exit0 | static-exec21.log |
| go test ./... -run '^$' | exit0；全包编译，不是执行全包测试 | static-exec21.log |
| go test -race -count=1 -v ./internal/relay -parallel=4 -run 选定InProcess矩阵（完整表达式在日志） | exit0；18个顶层组，34.258秒；无race报告 | metering-combined-exec21.log |
| go test -race -count=1 -v ./tests/verify -run '^TestCaptureCommandTiming$' | exit0；成功/非零/已取消计时边界，包1.619秒 | runner-timing-exec21.log |
| make frontend-test | make exit2，npm exit1；锁定冷安装connect EPERM 127.0.0.1:7897；Node wildcard未启动 | frontend-exec21.log |
| node --test 八个无外部依赖的受影响.test.mjs（路径在日志） | exit0；136项，0失败/跳过；仅补充VM/诊断合同子集 | node-subset-exec21.log |
| git diff --check HEAD（全整合） | exit2；main带入的历史首失败.log存在尾空白，不改写历史 | static-exec21.log |
| git diff --check HEAD -- internal tests .trellis Makefile go.mod go.sum package.json package-lock.json 与本Run文档 | exit0；范围检查不替代上一失败或固定门禁 | scope-check-exec21.log |

首次独立go run未设项目缓存，exit1：open <home>/Library/Caches/go-build/49/49b5812c6c9fc9e61e0716ad1441671047f1cbc09e2fab7bc6e0d458b96e6094-d: operation not permitted。frontend日志保留相同首失败；按Makefile已有GOCACHE路径重新采源码成功，不提权。一次只读shell枚举使用zsh特殊变量path，导致head查找exit127；改用file完成枚举，非产品失败。Darwin linker LC_DYSYMTAB警告保留，未影响上述race成功。

make frontend-test失败已确定发生在npm读取锁定playwright-core公开包时的本机代理connect权限，非产品断言失败；未修改代理配置、凭据、npm锁、测试入口或权限，也未重试求绿。纯Node子集不依赖被安装包，不宣称替代完整冷安装/原wildcard。

原基线make verify监听EPERM和实际HTTP最小诊断及源码关联见[历史results](results.md)与listener-diagnostic.log。沙箱监听能力未改变，不重复同一必然受限命令；本联合候选完整make verify **Not Run locally / Pending host**，不是Pass。静态组合命令总体exit2明确保留，其前序format/module/vet/build/compile各exit0；尾空白仅在逐字main的历史证据中，属于非本轮源码格式问题，不能删除原失败或重写他Run证据。

## 下游责任、未验边界与公共审查

固定tests对本候选完整冷make verify，保留冷安装、HTTP/SSE/取消、进程重启、原PNG/几何/关闭/键盘/三宽及新增self_hosted原门禁；比较158文件指纹、HEAD/main/MERGE_HEAD及冲突标记。按平台既有规则处理暂存/候选身份，本Agent不代运行Connector或写.git。失败回原development，重复同类失败先最小归因，不扩deadline、不放宽断言。之后新独立QA复核A/B计划、来源/版本、真实受控浏览器和未验项；旧tests9/QA10不能证明联合绿。

report随后原位更新本Run pr.md，publish/pr负责提交推送并更新同一PR36，标题“【产品测试】补齐受控RPM/TPM与安全调用计量回归”，默认Refs #34；旧pr.md当前仅历史单候选，不可直接发布。不关闭#33，不在本节点merge PR、部署或运行Actions。

delivery-review-checklist自审：复用M01/T001、原PRD/AC和test-plan，无新阶段/重复事实源；有逐文件版本、失败路径与检查命令；T001 In Review、M01 Ready for Review，不谎报Done/Accepted。当前可分发tests，产品/架构无待澄清冲突，发布Blocked by Validation。仍需完整宿主门禁和新独立QA。

原Issue33产品No-Go、真实普通/SSE成功及恢复、真实Token/费用、非零数值网页/自然窗/双浏览器供应商验收保持Not Run/Blocked。新增供应商/真实生成0，未读私有Provider配置，原24次/每次64tokens/90秒及未知历史不重置，budget not_attested不变；受控测试价33成本/7租户费不冒充供应商账单。
