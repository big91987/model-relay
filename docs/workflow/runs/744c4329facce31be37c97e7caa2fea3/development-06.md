# 固定 tests 执行 5 超时后的研发修复（执行 6）

2026-10-14，Issue #44，Run `744c4329facce31be37c97e7caa2fea3`，冻结 Workflow revision 2。共享工作区及 HEAD `1d581652736fbac4e8862b70367181c8a9f2f8b2` 未重建。本轮是项目验证设施的累计预算修复，不是 QA 缺章返工或产品功能开发。

## 冻结输入、原文件与完整失败回执

读取根 AGENTS.md、既有冻结身份摘录、首轮及执行 4 实施材料、候选接续说明和本轮失败 state。未发现 docs/tests/internal 下的额外 AGENTS.md。核对三个候选源码、两份旧报告和接续说明与首轮归档一致；旧报告仍是历史，不将其私有日志引用当作本 Run 证据。凭据只核对 GITHUB_TOKEN/GH_TOKEN/GIT_ASKPASS 存在性，均 false，未读取值。

项目原始回执位于 `test-results/verification/20261014T154026Z-2753749201/`。完整 verify.log 的 1054470 字节已全量读取、逐行解析及校验，SHA-256 `8361a3d64091f4557be14e80d436754d016bccfa1395b4f91bf3e041e1265735` 与 state 一致。原字节归档在本轮 `development-06/verify.log`，state 同目录保存；可复验解析器为 `check.cjs`。这不是平台外层 1075469 字节日志；未把用户显示的省略片段当作完整外层回执，也没有声称读取全部平台外层日志。

固定 Connector 退出 2，项目验证器退出 1，gate_exit=-1。冷起点实测 node_modules/npm 缓存不存在，manual_preinstall=none；源码及候选前后摘要一致，owned_group_absent=true，candidate_removed。原源码摘要 `a6f4ec9a2a4e9f87f501aee7f48251ccede244d46082fc09b415ea9953a03b3e`，158 文件。

## 真实归因及未验边界

- 历史 HTTP 红例仍按原分类通过：21 method failures 与 12 query/recovery passes。日志中的 26 个 JSON fail 事件含父测试/包聚合，全部属于该历史 HealthMethodsHTTP 红例，不能当作本版产品断言失败。
- 日志显示格式、模块校验、vet、278 个前端单测、race 包测试及 build 已经通过并进入 browser。本轮先前 candidate 不完整 Go 归档编译问题未再出现。
- 累计 gate context 在 prerequisite-npm 完成后开始，约 `15:40:41.730Z`；`15:52:41` 被内部 12 分钟预算取消。migration、legacy-ui、routing 的 browser phase end 均为 ok=true，分别耗时 15408、178165、214066 毫秒。
- console 于 `15:52:36.969904929Z` 开始，约 5 秒后 make browser/verify-gates 被 Terminated；项目末尾是 signal:terminated/context deadline exceeded。console 没有成功 phase end，full-console 尚未开始。取消时 console 输出失败诊断，不据此假定取消前后所有 UI 都正确，亦不预言修后全套必绿。
- 此失败不是 Connector 的 1800 秒外层预算耗尽，不是原 2Gi 环境的 signal:killed/OOM 归因。无需修改冻结准备、安装依赖、资源配额或平台代码。

## 实质最小修复与范围补充

仅修改 `tests/verify/main.go`、`tests/verify/main_test.go`。原 Makefile 入口/门禁顺序、锁文件、三个候选源码和产品实现保持不变。本轮新增修改也是测试设施源码，后续独立 QA 和最终报告必须覆盖，不能只审查原三源码。

将**累计执行 make verify-gates 的设施总预算**由 12 分钟改为固定 24 分钟，覆盖同一次冷启动构建和串行 browser 阶段；仍低于固定 Connector 的 30 分钟外层预算，正常路径为准备和清理留出余量，异常路径仍可能被外层限制，不承诺极端情况下所有辅助命令都能耗满其预算。未增加重试，未跳过任何门禁，未改变 HTTP/SSE/冷却期限、单项 browser 超时、并行或业务断言，也未改变辅助命令的 2 分钟上限及进程组清理策略。首轮说明所保留的验收期限不因此改变；24 分钟是本次公开记录的设施总预算修订，不将其伪装为冻结候选原值。

新增 gate_started_utc/gate_budget_ms/gate_execute_ms/gate_cancelled 回执字段及 GATE_START/GATE_END 日志标记，使下一次失败能直接核验预算和父 context 取消。父级更短 deadline 或取消继续优先；预算不会覆盖操作者或平台取消。

新增真实进程回归：验证总预算固定有界，50ms 父截止不被扩大，进程被取消且自有进程组消失。既有 green/nonzero/probe runner fixture 额外断言 gate 时序及 8 秒父预算；原真实失败 exit=23 分类不变。不需要等待 24 分钟验证 context 的截止元数据。

## 本轮实际定向验证

通过已预装 Go 1.25.6 使用测试自有临时 GOCACHE 运行 `go test -race -count=1 -v ./tests/verify`，实际退出 0，完整 stdout/stderr 和状态保存于本轮目录。包含新增预算/父取消回归、原进程组取消及全部既有 runner 反例。`go vet ./tests/verify` 实际退出 0；`go run ./tests/evidence` 实际退出 0，新 158 文件摘要为 `79467e60b36340bad4df1b090cd1518e74c8e29d6c2503c5632585a5ecf7316c`。临时 Go 缓存已删除，未手装任何环境依赖。

本轮没有代跑 make verify、浏览器全套或旧路由反例。已有固定失败仍失败，不追认绿色。首轮 11 项校验和六份候选材料逐字节核对通过，首轮说明和快照仍一致；授权缺章继续保留，尚无 QA 正式拒绝，不能提前补齐。研发不撰写 qa.md/report.json，不提交、push、PR、发布、合并或改数据库记录。真实供应商、支付、外部身份均 NotRun。

## 同版交接

本轮归档包含完整固定项目失败回执、新实现 patch、定向验证 stdout/stderr/退出码、新源码证据、解析器与结果。SHA256SUMS 绑定当前源码、所有本 Run 交付文件和本说明，排除清单自身，不仅依赖 HEAD 或不含文档的 158 文件摘要。

基于本轮实质累计预算修复与新可复验定向证据，正式 handoff tests 执行原 make verify，保留完整 stdout/stderr、退出码、GATE 时序、冷起点、材料及源码前后摘要和清理结果。若失败，先完整定位；没有实质修复或新的可复验证据不得同版反复求绿，涉及冻结准备/环境/凭据则正式 blocked。

全量通过后仍必须由独立 QA 审查原三个源码、执行 4 归档模块边界及本轮两个验证器源码，明确拒绝授权缺章并正式 revise；研发补齐后重新固定测试和 QA 复核，再操作者确认、报告及真实同版草稿 PR。本轮不构成 QA 放行或链路完成。
