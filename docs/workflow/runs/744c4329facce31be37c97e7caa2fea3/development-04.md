# 固定 tests 失败后的研发最小修复（执行 4）

2026-10-14，Issue #44，Run `744c4329facce31be37c97e7caa2fea3`，冻结 Workflow revision 2。复用当前共享工作区，HEAD 仍为冻结候选 `1d581652736fbac4e8862b70367181c8a9f2f8b2`。本轮是固定 tests 执行 3 失败后的设施归档修复，不是 QA 缺章返工。

## 输入及完整项目回执核验

先读取项目 AGENTS.md、prepare 冻结参数身份白名单和既有实施记录、首轮材料。候选三个源码、两份旧报告及接续说明再次阅读/核对；没有重新生成需求/设计、读取凭据值或重建工作区。历史基线对象存在且仍为 HEAD 祖先。旧报告继续是历史，不冒充本 Run 的结果。

原生 `read_command_output(seq=3, offset=0)` 返回外层保留日志首段，`eof=false`；未将此片段当作完整外层日志。另在工作区读取实际项目回执 `test-results/verification/20261014T152930Z-569886809/state.json` 和完整 `verify.log`。全量读取并解析所有 JSON 行，原字节归档到本轮 `development-04/tests-03-state.json` 和 `tests-03-verify.log`；日志为 704694 字节，SHA-256 `9294aeca06d921ba05c74d2a6266bb25fead4b7d447ee9bc6e6f14fcc8d11efd`，与 state 一致。这是项目保存的完整合并输出，不声称是 711632 字节的平台外层日志，也不声称可还原每条输出的 stdout/stderr 归属。

固定 Connector 退出 2；内部 verify-gates 退出 2；项目验证器回执 exit_code=1。历史 review 重建接受真实预期红例，health-history 接受精确 21 个方法失败及 12 个查询/恢复通过；这些不是本版产品全量通过。format-check、模块校验之后，原 `go vet ./...` 把首轮归档的 `candidate/internal/relay/routing_http_test.go` 当成独立 Go 包，报第 77 行 `undefined: fixture`。归档仅保存一个测试文件，没有该包其余夹具；失败归因是研发执行 2 的证据归档布局，不是产品业务故障、环境 OOM 或历史红例分类故障。

该固定执行未到 frontend-test、test、build、browser，不能声称这些门禁通过。state 记录源码及临时候选未变，冷起点 node_modules/npm cache 不存在，NODE_PATH=unset，manual_preinstall=none；清理结果为 owned_process_group_cleanup_attempted; candidate_removed，owned_group_absent=true。以上是回执观察，不扩大为全局冷缓存、全环境释放或当前 Run 已结束。

## 实质修复和范围

只在 `first-round/candidate` 新增无依赖 `go.mod` 和解释用 README，将不完整证据副本设为独立、不可运行的归档模块边界。根模块原有 `./...` 门禁继续检查全部产品及测试包；不删门禁、不改 glob、不加 build tags、不降低断言或期限，不改根 go.mod/go.sum、Makefile、三候选源码或业务实现。不在归档模块跑测试。

首轮原清单未增删或改写，SHA-256 仍为 `9b5c1f5e8a87a229d324c5a72ee910ad6e542fe2cf924c45939e8a7a80a11a7f`；原列出的 11 文件校验全部成功，三个副本与实际候选逐字节一致。新增模块边界是执行 4 的材料，不冒充首轮文件。

## 实际定向验证及版本

- 开始定向检查时裸 `go` 命令因当前会话 PATH 未含现有 Go 目录而退出 127，未运行 Go 检查。随后定位预装 `/usr/local/go/bin/go`，仅为命令设置 PATH，未安装工具或依赖、未改冻结准备步骤。其版本为 go1.25.6。
- `go list ./...` 实际退出 0，完整 stdout/stderr 保存在 `packages-after.stdout` / `packages-after.stderr`；11 个根包保留，包括 internal/relay、tests/review-history、tests/verify 和既有其他归档 acceptance 包，唯本次不完整 candidate 副本不再纳入。Go 自身输出三条锁定模块下载诊断，未手工预装。
- 原 `go vet ./...` 实际退出 0，完整 stdout/stderr 及退出记录为 `vet.stdout` / `vet.stderr` / `vet-status.txt`。这是对本次失败门禁的直接检查，不是完整 make verify。
- 可复验 `node docs/workflow/runs/744c4329facce31be37c97e7caa2fea3/development-04/check.cjs` 实际退出 0，结果为 `check-result.json`；验证完整项目日志长度/摘要、原失败位置、首轮 11 项摘要、三个源码副本和根包发现范围。此脚本是研发诊断，不是独立 QA 报告。
- `git diff --check` 无错误；原候选源码、根模块/锁和 Makefile 与 HEAD 无差异。HEAD 及项目 158 文件摘要不因归档边界代表新的交付版本；本轮交付以 `development-04/SHA256SUMS` 的路径/内容摘要区分，后续 tests/QA 必须另核实际工作树及全部新增材料。

本轮没有运行全套 make verify、重复旧定向 HTTP/mutation 诊断或浏览器，没有改用户数据库、读取凭据、提交、push、PR、发布或合并。没有创建 qa.md/report.json。真实供应商、支付、外部身份均 NotRun。

## 正式交接及未决事项

本轮有实质归档模块边界修复及直接 vet 通过证据，交 tests 对当前工作树执行原 make verify；不是同版失败后无改动循环求绿。请保存完整 stdout/stderr/退出码、冷起点、源码前后摘要、归档文件摘要及清理结果。任何新失败先读完整证据再归因。

工作说明仍是执行 2 的首轮事实快照，原字节及授权故障注入保持不变；“停止、恢复与环境释放”仍故意缺章。固定测试完整通过后，独立 QA 必须审查原三个候选源码、归档修复及边界，明确拒绝缺章并正式 revise，研发才补齐。当前无独立 QA 拒绝/复核、正式操作者确认、最终报告或本 Run 草稿 PR；本轮 vet 通过不能替代其中任何步骤。
