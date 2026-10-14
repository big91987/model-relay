# 交接前的既有 macOS 入口兼容修复

2026-10-14，Issue #44，Run `744c4329facce31be37c97e7caa2fea3`，研发执行 6 同一工作区接续，Workflow revision 2；HEAD 仍 `1d581652736fbac4e8862b70367181c8a9f2f8b2`。先读取 AGENTS.md、冻结身份、当前源码、上轮真实材料与已有部署说明，没有重建工作区。上轮 handoff 因用户输入待投递而未被接受，本轮先处理该兼容性纠正，不把上轮建议称作已经执行的固定 tests。

## 原支持范围的真实依据

- `.github/workflows/deploy-local.yml:27` 的 prepare job 明确为 `[self-hosted, macOS, ARM64, he-full]`，第 59 行 deploy job 亦为该宿主类型。
- `docs/local-deployment.md:5` 明确要求同机 Runner 在独立工作树执行完整 `make verify`。原 Makefile 的入口没有 Linux-only 限制。

这证实既有 macOS 本机/自托管路径，不能猜测仅支持 Linux。上述原说明和工作流原字节未改；未读取/调用控制器凭据、触发 Actions、发布或改部署图。本轮不是新增操作系统支持，而是修复上轮新增的无条件 Linux 观测指标把既有入口锁死的回归。

开始编辑前，当前五份源码与上轮授权归档逐字节核对成功。用户提供 main.go 短前缀 `62997a1a` 与本地全文一致；所报 main_test.go `f477b4c2` 与实际本地上轮归档 `37f05d86...` 不一致，未据不完整外部摘要推断其字节或覆盖本地文件。完整既有摘要和来源记录在 `development-06-compatibility/input-identity.json`，新版本再以全文 SHA-256 绑定。

## 最小兼容处理及指标状态

相对上轮仅修改 `tests/verify/resources.go`、`main_test.go`、`gate_resources_test.go`；主 `main.go` 和 `gate.go` 与上轮授权归档仍原字节相同。源码固定 24 分钟上限、有效父截止、原取消/清理和全部门禁保持，没有新增 Profile、私有环境特批、公开采样跳过开关或预算配置。

生产采样器使用 **Go 编译目标的真实 `runtime.GOOS`**，记录 `goos`、runtime.NumCPU、GOMAXPROCS 及 UTC 采样窗口：

| 状态 | 行为及回执 |
| --- | --- |
| 非 Linux 的 `unsupported`（既有 Darwin 路径） | 完全不读取 `/proc` 或 `/sys/fs/cgroup`；保存明确不可用原因和运行时 CPU，cgroup identity/values/counters 均为空，没有伪造配额、指标或零差值。该观测指标自然不可用不改变原 gate 的成功/失败结果，也不算指标已验证。 |
| 当前 Linux v2 目标的 `available` | 保留实际自身 cgroup 读取、严格解析和前后差值条件；最终指标有完整来源。 |
| Linux 的 `error` | 读取/解析失败仍明确失败；不将缺失、无权限、非法内容或 counter/身份异常降级成 unsupported。失败采样不会把原 gate 失败清除或变成绿色。 |

差值仅在同一可用 Linux 平台和原身份/限额/计数约束满足时生成；同为 unsupported 时返回“无差值、无观测错误”，不生成零 counter 假记录。平台/状态发生变化明确失败。试图把 Linux 标成 unsupported 亦被校验拒绝。

GOOS 注入参数仅存在于私有采样函数的隔离单测，生产路径总是 runtime.GOOS；不是新的执行模式或当前 Run 特批。Linux 本次目标仍必须完成有效前后采样与审计。没有增加 cgroup v1 指标适配或宣称任何新平台的整个产品都可运行。

## 新版本真实定向结果

在当前 Linux 标准环境，使用已预装 Go 1.25.6 和独立临时 GOCACHE **重新**执行 `go test -race -count=1 -v ./tests/verify`，最终版实际退出 0，包耗时 2.009s；`go vet ./tests/verify` 实际退出 0。上轮绿不作为本版验收证据。新完整 stdout/stderr 与状态归档在本兼容目录。首个兼容版本的定向结果、源码快照、摘要及实施记录保留于 `attempt-01/`；之后为加强真实读取错误证据，仅修改定向回归，用自有临时目录的缺失文件产生真实 ENOENT，再对最终版重新执行上述定向检查和 Darwin 编译，不是同版全量求绿。

新回归包括：

- 注入 Darwin 的不可用场景，用 fatal reader 保证从未读取 Linux 路径；记录 runtime CPU 与 unsupported 原因，无 cgroup data/delta。
- 完整隔离 runner 的 unsupported 正常 gate=0 路径仍成功；相同 unavailable 指标加真实 gate exit=23 仍失败，证明不锁死正常入口也不隐藏测试失败。
- Linux 实际读取失败独立回归通过 `os.ReadFile` 读取自有临时目录的缺失文件，核验真实 ENOENT 后仍为 error/fatal/无差值，没有破坏真实 cgroup 文件；既有非法 quota/counter、缺失、下降、身份/限额变化与合法前后差值回归都重新通过。Linux available 的原“必须存在真实差值”断言保留，只有天然 unsupported 的平台按明确不同的指标契约断言无差值。
- 24 分钟有界策略、100ms 真实子进程短预算/父截止取消及进程组/子 PID 清理、原真实非零和起止采样失败不假绿回归均在本版重新运行通过。

另以 `GOOS=darwin GOARCH=arm64 CGO_ENABLED=0 go test -c` 对**验证器测试包**做编译检查，实际退出 0，输出格式是 Mach-O 64-bit arm64。该设置仅是隔离临时编译器目标，不是运行 Profile 或当前固定 tests 的环境特批；二进制未执行并已删除。这不能冒充 macOS 实机、完整 Go/UI 或供应商验收，Darwin 实机运行仍 NotRun。临时 Go 缓存也已删除，没有手装依赖。

新 `go run ./tests/evidence` 实际退出 0：161 文件，最终源码摘要 **`e637c8f9a92859a8766e9c6e23e2c08f53d9b40d96b5c0a247c455e7e19bac9b`**；此前 `e5efe502...` 只属于 attempt-01。五份当前源码以 `.go.txt` 原字节保存，避免再次把不完整证据副本编译为根模块包。Makefile、模块锁、原三候选源码/历史材料、health-history 与 browser 原文件保持不变；首轮 11 项清单和授权缺章仍原字节，未提前补章。

## 本版交接与未验

所有原执行 5 失败、旧 24 分钟提案/撤回、归因纠正、新授权决策及本兼容修复保留。当前只用 `development-06-compatibility/SHA256SUMS` 绑定本版全部交付；旧清单/源码摘要/回归仅用于历史，不证明本版通过。当前源码快照、fresh stdout hash、各项真实状态和关键不变项由本轮 check.cjs 核验。

完成本版实质兼容修复和新定向验证后，正式 handoff tests 一次原 `make verify`，外层仍 1800s，内部累计上限 24 分钟。当前 Linux 入口须输出真实 GOOS、available/error 指标状态、起止快照与有效差值，同时保留完整日志/退出码、主 gate 与嵌套 fixture 身份、冷起点、源码/全部材料前后摘要及清理结果。若 24 分钟仍超时则正式 blocked，依据新阶段证据再决策，不自动加时或全套盲重跑。

独立 QA 必须复核既有 macOS 支持依据、unsupported 不锁死/不伪造指标、Linux 读取解析失败严格性与真实合法差值，连同五份预算/观测设施源码、原三候选源码及归档模块边界审查原断言/取消语义完整性。不能称性能 SLA 或 macOS 实机验收；仍须实际拒绝首轮缺少“停止、恢复与环境释放”章，正式 revise 后研发才补齐并再次完整固定测试/QA复核。

目前无本 Run 全量通过、独立 QA、正式操作者确认、报告或新草稿 PR。真实供应商、支付、外部身份 NotRun。没有读取凭据值、提交、push、PR、发布、合并、改业务/数据、资源、权限或冻结准备。
