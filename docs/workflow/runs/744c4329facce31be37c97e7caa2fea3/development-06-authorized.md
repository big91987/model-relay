# 维护者授权后的累计预算校准与资源回执

2026-10-14，Issue #44，Run `744c4329facce31be37c97e7caa2fea3`，研发执行 6 同一工作区接续，冻结 Workflow revision 2，HEAD `1d581652736fbac4e8862b70367181c8a9f2f8b2` 不变。本记录是**归因完成后的新授权候选**，不是恢复已撤回提案，也不是复用其定向绿作为新验收。

## 决策及完整历史保留

维护者已独立核对执行 5 正式全文、主 state 与执行 6 纠正材料，确认 2CPU/8Gi 是当前标准冻结环境，无资源配置漂移；不能据累计 CPU 节流断言 CPU 是唯一根因或扩资源。维护者明确解除此前“先归因并保持 12 分钟”的临时约束，授权一次累计设施预算校准，上限 24 分钟；外层固定 Connector 仍 1800 秒。授权范围和不变项另存 `development-06-authorized/decision.json`。

本轮开始时验证器确实已恢复为冻结 HEAD 原字节。执行 5 失败完整项目回执、维护者外层全文 EOF 身份、第一次 24 分钟提案、撤回/纠正和后验资源归因全部保留，分别在 `development-06/`、`development-06-correction/` 及对应实施记录中，未覆盖。其旧清单、解析器和定向结果具有历史版本边界，不证明本次新候选通过。

原 blocked 已由维护者明确的新项目设施授权解除，不再等待扩资源或改冻结准备。本轮没有资源/权限/冻结工作流修改，也没有重建工作区、恢复旧 Run、手装依赖或改产品业务。

## 预算依据及空间

原主 attempt `20261014T154026Z-2753749201` 从 state 起点到第一个 browser phase 约 322.279 秒，其中包含门禁外准备；已成功的 migration/legacy-ui/routing phase 合计 407.639 秒。console 在主全程约第 730.134 秒才开始，被内部累计截止截断，full-console 未启动。原 12 分钟是验证设施累计预算，**不是产品单请求响应 SLA**；不能把其超时当作业务响应失败。

项目 `tests/verify/gate.go` 固定源码常量为 **24 分钟（1440000ms）**，只作用于同次隔离 `make verify-gates` 的累计运行。无环境变量、命令行或 Run 专属预算开关。以旧关键路径估算，约 307 秒 gate 内前置加约 408 秒已完成浏览器阶段后，新预算约剩 **725 秒**给完整 console/full-console 和阶段结束；这是空间估算，不是对未运行阶段耗时或通过的保证。

固定外层 1800 秒与 gate 上限相差 360 秒，供门禁外准备、源码前后核验、依赖探针、资源结束采样、进程与候选清理使用。仍保留原辅助命令每次 2 分钟上限；多次异常辅助命令若各自耗满可能被外层切断，不能宣称六分钟足以覆盖所有病理路径。父级更短截止或取消优先；gate 不允许超过源码 24 分钟上限。若本版仍总超时，必须正式 blocked，以新主阶段证据再决策，**不自动加时、不全套自动重试**。

## 最小实现及不变范围

修改 `tests/verify/main.go`、`tests/verify/main_test.go`，新增 `tests/verify/gate.go`、`tests/verify/resources.go`、`tests/verify/gate_resources_test.go`。五文件原字节快照以 `.go.txt` 存本轮目录，避免把证据副本编译为根模块测试包；实施 patch 包含全部五文件。

- gate 封装仍执行原 `make verify-gates`，没有改变入口或门禁顺序。原 execute 的 SIGTERM/SIGKILL、WaitDelay、进程组隔离和 ownedGroupAbsent 判定未改。gate 非零返回保留，取消仍非成功退出。
- receipt 保存固定 configured budget、实际有效 budget、主 gate 起点/耗时/取消标记及进程组结果；完整日志包含 GATE_START/GATE_END，使主 gate 与嵌套回归 attempt 可区分。测试用短预算参数仅为私有函数的回归注入，不是公开配置；0、负数和超过 24 分钟直接拒绝，未启动子进程。
- Makefile、模块锁、原三份候选测试源码、health-history 分类器及 browser 用例原字节未变。全部严格断言、历史分类、HTTP/SSE/冷却/单动作期限、race 并行和冷隔离保留；未用降低响应要求、增加重试或 sleep 来求绿。

## 起止资源快照/差值及失败语义

Go 验证器入口在 receipt/自有候选创建后、源码 inventory 前只读自身 `/proc/self/cgroup`，读取当前 v2 可见 cgroup 的 cpu.max/cpu.stat、memory.max/memory.events、cpuset，以及进程 affinity；记录 Go runtime.NumCPU 与 GOMAXPROCS 的查询值，不设运行时并发或资源限额。不读取环境凭据或其他 cgroup。cgroup 身份记录 SHA-256 而非主机路径；失败只记录固定文件标签，不暴露底层路径。

资源开始/结束均保存 UTC 采样窗口；结束采样位于归档 defer，在候选及自有进程清理后，仍先于最终日志哈希/receipt 保存。差值只在同 cgroup、配额/affinity 不变且全部所需计数有效单调时生成。cpu.stat 与 memory.events 原样计数是累计值，resource_delta 是这两次本次采样间差值；**可见 cgroup 可能包含其他活动，不称进程独占用量或唯一因果**，文件读取也不是原子快照。

样本覆盖 Go 验证进程内部准备、门禁与清理，不包含 `go run` 在进入 Go main 之前的启动编译或整个外层 Connector 生命周期。缺失/读取错误、非法计数/限额、计数下降或环境身份变化都拒绝产生误导性差值，并使最终验证非零。开始采样失败不运行门禁；结束采样失败不把已完成 gate=0 提升成最终绿色，同时保留原 gate 非零错误。采样错误不会覆盖或清除原测试失败。

## 新的实际定向验证

本次使用已预装 Go 1.25.6 和独立临时 GOCACHE，重新执行 `go test -race -count=1 -v ./tests/verify`，实际退出 0，包耗时 2.010 秒；`go vet ./tests/verify` 实际退出 0。完整 stdout/stderr、退出码和状态均在本轮目录，临时缓存已删除。没有运行项目全套或浏览器全套，也不引用旧 24 分钟提案的测试结果作本轮证据。

新真实子进程回归覆盖正常 stdout/stderr、真实非零 23、100ms 短 gate 截止、100ms 更短父截止、预先父取消以及非法 0/负数/超上限预算。短截止实际终止有长等待子进程的独立进程组，核对组消失及子 PID ESRCH；等待仅是隔离取消 fixture 的受控阻塞，不是产品延时或为绿例加 sleep/retry。既有 runner 的真实取消、清理、失败分类仍通过。

新资源回归覆盖正常解析及差值、读取失败、非法 quota、非法/缺失 counter、counter 下降、cgroup/limit 变化。完整 runner 注入的开始采样失败、结束采样失败、非零 gate 同时结束采样失败均被拒绝；原 gate exit=23 仍可识别，证明不能产生假绿。

新 runner fixture 实际开始样本显示 cpu.max=`200000 100000`、memory.max=`8589934592`、Go runtime.NumCPU=112、GOMAXPROCS=2；这只是本次定向 fixture 的真实资源观察，不是下一次完整固定测试的起点。下一次固定入口必须自行生成同版起止样本及差值。

`go run ./tests/evidence` 新执行退出 0，161 文件，源码摘要 **`a026e6b9122f78b43c62d53834f5eba6e7f6242186fbc12c302695831f5bd165`**，HEAD 不变。`check.cjs` 对五份源码快照、首轮 11 项、原三源码/两旧报告/接续说明、入口/锁/分类器/浏览器不变及新回归实际结果进行核验。凭据仅检查 GITHUB_TOKEN/GH_TOKEN/GIT_ASKPASS 存在性，不读值。

## 当前同版、交接及独立 QA 边界

当前清单 `development-06-authorized/SHA256SUMS` 绑定全部本 Run 材料及当前源码，排除自身；历史清单不再作为当前版本核对入口。原交付说明与首轮快照仍原字节，授权缺少“停止、恢复与环境释放”章节保持，没有提前补章或代写 qa.md/report.json。

基于明确授权、当前实质修复和新定向证据，仅正式 handoff tests **一次原 make verify**。tests 必须保存完整 stdout/stderr、退出码、主 gate configured/effective budget、起止资源/差值、冷起点、源码及全部交付材料前后摘要和清理结果；不混用嵌套 fixture attempt。若 24 分钟仍超时，正式 blocked 后凭新的阶段证据再决策，不自动加时、换资源或同版反复重跑。

完整固定测试通过后独立 QA 应审查原三个候选源码、执行 4 归档模块边界和本次五文件累计预算/资源回执修复，明确验证原断言与失败/取消语义完整性，不能称性能 SLA 验收。QA 必须实际拒绝首轮缺章并正式 revise，之后研发才补齐并重新完整固定测试与 QA 复核，再操作者确认/报告/真实同版草稿 PR。

本 Run 当前仍无全量通过、独立 QA、操作者确认、报告或新草稿 PR。真实供应商、支付、外部身份 NotRun。无凭据值读取、提交、push、PR、发布、合并、资源或权限变更。
