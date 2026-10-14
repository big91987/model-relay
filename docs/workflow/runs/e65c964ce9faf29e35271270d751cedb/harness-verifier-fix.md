# 用户授权续接：历史验证器最小修复

## 同 Run 身份与范围补充

2026-10-14，继续 Issue42 / Run `e65c964ce9faf29e35271270d751cedb`，冻结 Workflow revision 2，HEAD 仍为 `5ff456cb6de0ce8d45bb2faac4c964cda9efcbf0`。读取原工作区 AGENTS.md、冻结 tests 输入身份白名单、本地准备标记和既有文件；首轮及 development-e4 归档摘要均核对成功，没有重建工作区。

用户正式补充范围：允许本 Pipeline 最小修复项目历史验证器，分离机器 stdout 和完整 stderr 诊断；不得改旧历史基线、历史行为断言、故障分类、完整门禁和期限。此授权解决执行 4 记录的范围阻塞；旧失败与原阻塞记录原样保留，不能改成通过。本轮未修改通用平台或冻结准备步骤。

实际源码修改仅 `tests/review-history/main.go`、`tests/review-history/main_test.go`。新增本记录；此前说明和实施记录继续保留。未改产品行为、供应商、数据库、Makefile、锁文件、模板或既有部署配置。

## 先回归红例，再修复

新增命令边界测试通过 Go 测试子进程和明确隔离的 git/tar/go 命令夹具执行真实 `run()` / `main()` 路径。夹具不是供应商或历史产品集成证据，不修改工作区源码，不读取凭据，只在测试拥有的临时目录构造导出目录。实际 Go 历史 HTTP 重建另外核验。

九个场景覆盖合法 JSON 的预期历史红例、依赖下载非 JSON stderr、非法 stdout、真实子进程退出 2、意外退出 0、缺历史断言，以及 stderr 上的构建、race、listener 故障；所有场景检查 stderr 首尾诊断完整保留。非法 stdout 检查原始未知内容被保留，不能通过过滤行求绿。

修复前先运行 `go test -count=1 -v ./tests/review-history`，真实退出 1；新增九个场景失败，原 assess 回归通过。真实输出 `harness-output/verifier-fix/regression-red.log`；修复前源码、新测试原文和红例摘要保存在同目录 `main-before.go`、`regression-red-main_test.go`、`RED-SHA256SUMS`。没有把编译失败或人写的预期文本当作红例。

随后最小修复：

- exec.Cmd 的 stdout 与 stderr 分别接独立缓冲区，逐行 JSON 解析只读取 stdout；不跳过非法 stdout。
- 完整 stderr 在成功或拒绝路径均输出；拒绝路径另输出完整原始 stdout 和具体错误。
- 保留原 stdout 历史断言 assess，stderr 不能补造缺少的 stdout 行为断言；另结合 stderr 执行相同 assess，以免隐藏原故障分类。
- 只有真实退出 1 可以进入预期历史红例判断；启动错误、信号或其他退出码仍拒绝，不把执行故障当业务红例。
- 旧基线 `66f1328ada75ea90e5638fbf2bd58663ac86347e`、测试正则、四个预期失败事件、历史实际行为断言、120 秒期限、扫描上限和后续全部门禁未变。

## 定向验证与版本

- `go test -race -count=1 -v ./tests/review-history` 真实退出 0，九个场景及既有 assess 测试通过。日志 `harness-output/verifier-fix/regression-green.log`；这不是固定 make verify 全量通过。
- `go vet ./tests/review-history` 真实退出 0，见 `harness-output/verifier-fix/vet.log`。gofmt 已执行；git diff --check 无错误。
- 使用修复后的真实入口 `go run ./tests/review-history` 重建历史实际 HTTP，验证器真实退出 0；内部历史 go test 仍为预期退出 1，实际观察两个零软预算 status=200/dispatch_delta=1 及回滚 status=200/dispatch=1。见 `harness-output/verifier-fix/actual-history-green.log`、`targeted-status.txt`。这是受控历史重建，不是原开发时回执、当前产品全量绿例或真实供应商证据。
- 项目原有 `go run ./tests/evidence` 真实退出 0；158 文件新源摘要 `12d91213e103d69050c427579d36ec20595d5284a9ab6d75926761380784de70`，见 `harness-output/verifier-fix/source-evidence.json`。HEAD 不变不能代表工作树未改，固定 tests 和 QA 必须核对此新源摘要及文件摘要。
- 首轮说明 SHA-256 仍为 `c0ec7667d07fb4eccbd346f69a64ebff141178a9609015dc71a5dec56818504b`；授权缺章仍保留，未提前补齐、代写 QA 或虚构 QA 拒绝。
- GITHUB_TOKEN、GH_TOKEN、GIT_ASKPASS 存在性均 false，只检查名称存在性，未读取值。无提交、push、PR、发布、合并或真实供应商调用。

## 原失败、固定 tests 与后续交付

tests 执行 3 的原 make verify 失败及完整项目回执保留于 `harness-output/development-e4/tests-seq3-state.json`、`tests-seq3-verify.log`。原日志 SHA-256 为 `e90c9d5baddb5812e1ffc1ac946ce7f6d5499241081227247ff6477ddad9ae9c`；原退出 2，不覆盖或升级为通过。原开发诊断可能填充默认 Go 模块缓存，此局部验证不冒充原冷启动全量通过。

本轮以实质验证器修复及真实回归红绿为依据，原生 handoff 交固定 tests 执行原 `make verify`，不替换门禁。请保留完整回执、退出码、source_before/after 和日志摘要；失败先读取完整诊断，不反复重交同版。随后由独立 QA 读取真实同版文件与固定回执，正式拒绝首轮授权缺章并 handoff 研发补齐；研发不得绕开这次故障注入。之后还需重新固定测试、QA 同版复核及正式人工确认。

QA/report/发布角色应在最终报告与 PR 说明此次范围补充、真实回归红绿、原固定失败和新固定结果，以及未验真实供应商边界。研发不提前创建 `qa.md`、`report.json` 或草稿 PR，也不代为宣称审批或交付完成。
