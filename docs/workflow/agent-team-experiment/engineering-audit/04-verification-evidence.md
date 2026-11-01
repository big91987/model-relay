# 验证入口与证据成本静态审查

## 结论与边界

本报告只对候选源码做静态阅读和只读 Git 查询；没有运行 `make verify`、任何 Go/Node 测试、历史重建或浏览器旅程，也没有重复独立 QA。因此，本审查**不能证明门禁通过**，也不裁决 Bearer、时钟或备份清单三项产品行为。下文的“通过”仅描述入口代码预期或前置报告的自述，不是本次执行结论。

审查候选为 `122c9abece28f562616483e991098d020d236d6d`。只读 `git log` 验证三项前置已经按顺序进入候选：Bearer integration `2d48ee08f12368e1b806fcc0975943e2eca0c8ee`，单次时钟 integration `532e713acc998162c682b310c8f280e1651e6f95`，严格备份清单 integration `122c9abece28f562616483e991098d020d236d6d`。任务控制器记录的三个实现提交分别为 `67f46449…`、`318639d8…`、`7dc57094…`。

状态标记：**已验证**指本次由源码或只读 Git 直接确认；**推断**指由控制流推得、未经执行；**未执行**指本任务明确禁止的动态验证。

## 固定入口与阶段清单

```text
make verify
└─ go run ./tests/verify
   ├─ 创建 test-results/verification/<UTC>-<random>/ 与 verify.log/state.json
   ├─ git ls-files（tracked + untracked, exclude-standard）→ 私有临时候选
   ├─ 候选/原工作树 source evidence 对比；go/node/npm 版本探测
   └─ 25 分钟总预算：make verify-gates（串行、首错停止）
      1. source evidence
      2. review-history：旧 commit + 当前回归测试，预期精确失败
      3. health-history：旧 commit + 当前测试，预期精确失败
      4. gofmt check
      5. go mod verify
      6. go vet ./...
      7. npm ci → frontend Node tests
      8. go test -race -count=1 -v ./... -parallel=4
      9. go build -trimpath
     10. latency experiment
     11. npm ci → actual-Go-service browser runner
   ├─ 候选/原工作树结束摘要对比
   ├─ 仅 gate exit=0 时做 Playwright/Chromium dependency probe
   └─ 复制候选 test-results、写最终 state.json、哈希 verify.log、清理候选
```

依据：`Makefile:8-22,24-47`；`tests/verify/main.go:80-235,242-277,382-455`。Make recipe 的逐行失败语义以及 runner 汇总 `execute` 错误使阶段首错即停（**已验证控制流，未执行**）。`tests/verification/dependencies_test.go:13-64,66-97` 还以 dry-run/故障夹具约束依赖顺序和 npm 失败不得启动消费者。

## 正向发现

1. **P2 / 已验证：候选隔离和失败收据设计较完整。** runner 复制 Git inventory 到私有临时目录，拒绝非普通文件、路径逃逸和父目录 symlink；清除 `NODE_*`、`NPM_CONFIG_*`、`GIT_*` 等继承注入，固定本地 toolchain/cache，并用独立进程组、超时、TERM/KILL 和残留检查收口。即使前置阶段失败，defer 仍落盘最终日志哈希、字节数、结束时间和清理状态。依据：`tests/verify/main.go:100-167,267-364,382-455`；相应边界测试在 `tests/verify/main_test.go:18-151,223-290`。影响：降低本机依赖污染、挂起子进程及“失败无收据”的概率。

2. **P2 / 已验证：当前候选同版检查覆盖主要实现、测试、锁文件和契约。** evidence 对路径、长度、内容做排序 SHA-256，并记录 HEAD、Go 版本、文件数与 scope；runner 在执行前比较原工作树和隔离候选，执行后分别检查两边未变。依据：`tests/evidence/main.go:17-58`；`tests/verify/main.go:158-168,210-223`。影响：正常情况下可把执行结果绑定到工作树内容，而不只绑定 commit。

3. **P2 / 已验证：历史红基线不是泛化“命令失败即成功”。** 两个重建器都要求固定 commit、当前测试文件、race、限定测试名及精确失败集合/诊断，并排除监听失败、build/race/超时等设施红；输出声明这是重建而非原开发收据。依据：`tests/review-history/main.go:17-91,102-128`；`tests/health-history/main.go:18-96,99-176`。影响：能防止错误原因漂移后仍把历史门禁误判为有效。

4. **P2 / 已验证：日志诊断对阶段和超时较友好。** 辅助命令有 `COMMAND_START/END`、预算、耗时、exit、cancelled、进程组状态；总 gate 有 `GATE_START/END`，最终有 `CLEAN_END`。日志按行做工作树、候选、HOME、temp 路径替换。依据：`tests/verify/main.go:188-205,234,382-417,457-501`。影响：保留首错位置和超时/清理区别，同时减少本地路径泄露。

## 风险与成本

### P1：历史门禁隐含“固定旧对象已经本地可用”，与 partial clone 不相容

**已验证：** `review-history` 和 `health-history` 分别硬编码 `66f1328a…`、`9b6994ed…`，直接执行 `git archive <sha>`，失败只报告 `baseline unavailable`，没有预检、对象包、获取策略或“缺对象/缺 commit/网络失败”的细分（`tests/review-history/main.go:17,26-39`；`tests/health-history/main.go:18,34-51`）。只读 Git 配置显示当前仓库 `extensions.partialClone=true`、`remote.origin.promisor=true`、filter=`blob:none`；本次 `git cat-file -t` 确认两个 commit 当前可见，但这不证明其全部 blobs 在新的离线/受限 clone 中可归档。

前置报告提供了实际成本样本：三个任务的早期 `make verify` 都在产品测试之前因 review baseline exit 128 停止；对象准备后又在 health baseline exit 128 停止，直到第六轮才自述绿色（`strict-bearer-auth.md:81-97,114-148,166-178`；`single-clock-advance.md:44-51`；`strict-backup-manifest.md:40-43,60-63,80-93`）。

**影响：** 固定入口对 clone hydration 状态敏感；历史设施红阻断 format/vet/current tests/build/browser，导致昂贵整轮重试并掩盖当前候选诊断。**推断：** 在允许 lazy fetch 的环境，`git archive` 还可能发生未显式记录的网络取对象；在离线环境则直接失败。

**建议（仓库基建，P1）：** 将两条历史基线及所需对象做版本化、可校验的离线 fixture/bundle，或在正式门禁前增加只读 baseline preflight，逐个报告 commit/tree/blob 完整性及补齐指引；不要在 gate 内隐式联网。把基线 SHA、fixture SHA 和重建测试 SHA 纳入统一 machine-readable receipt。**通用平台：** partial-clone 创建/调度阶段根据仓库声明预取固定基线，并返回 hydration receipt，而不是由任务反复试错。

### P1：绿色证据不可携带，收据与完整日志/浏览器产物缺少统一不可变索引

**已验证：** runner 把 `state.json`、`verify.log` 和复制的 browser tree 写到 ignored 的 `test-results/verification/...`；`state.json` 仅保存 `verify.log` 的 SHA-256/字节数，没有列举或哈希复制后的 browser 文件，也没有给 `state.json` 自身生成外部摘要（`tests/verify/main.go:85-120,231-251`）。source evidence 明确排除 delivery receipts 以避免自引用（`tests/evidence/main.go:58`）。当前审查 worktree 中不存在三份报告引用的绿色 receipt 目录，只剩报告中的路径和结果摘要；因而本次无法把 `GATE_END`、完整测试日志、源码摘要和浏览器收据重新串联验证。

**影响：** “报告 → receipt → 完整日志 → 源码同版 → 子产物”的链条跨 worktree/清理边界断开；报告叙述可能是真实的，但独立审查者无法仅凭候选仓库复核，且复制后的 browser 产物可被单独改写而不触发顶层日志摘要。静态源码只能证明未来执行会尝试写收据，不能证明前置报告所述那次执行实际发生或保持完整。

**建议（仓库流程，P1）：** 每次用于合入结论的运行应产出一个小型、脱敏、可提交的 evidence index，记录 candidate integration SHA、working-tree digest、state/完整日志/每个子产物的 SHA-256、命令、exit、时间和执行环境标识；大日志可放受控制品库，但索引需持久、内容寻址并带保留期。报告引用内容摘要而不只引用本地路径。**通用平台：** handoff/合入控制器自动接收 receipt 目录、校验摘要并返回稳定 artifact ID；禁止仅凭自由文本“绿色”完成证据关联。

### P1：摘要 scope 与实际候选 inventory 不完全等价

**已验证：** 候选复制范围是所有 tracked 加所有未忽略 untracked 文件（`tests/verify/main.go:137-143`），但 evidence 只覆盖固定顶层文件、`cmd/internal/tests/.trellis/spec` 和少数按 basename 匹配的历史契约（`tests/evidence/main.go:18-41`）。例如其他未匹配的 `docs/`、脚本或新增顶层构建配置会进入候选却不进入 digest；文件模式也不进入摘要。`CandidateUnchanged`/`SourceUnchanged` 实际比较的是该有限 JSON，而字段名容易被理解成完整候选不变。

**影响：** 未被 scope 覆盖但会影响命令的文件可能在运行中改变，仍显示 `source_unchanged=true`；新增入口文件未更新 evidence 清单时也可能形成同版盲区。**推断：** 当前主要 Go/Node gate 文件已覆盖，但不存在“执行所读文件集合等于摘要集合”的机械保证。

**建议（工程工具，P1）：** 以经审计的 Git inventory 为摘要基础，仅显式排除生成目录和 receipt，并把 path、mode、size、content 全部纳入 Merkle/index；输出 include/exclude 数量及排除理由。增加回归：在任意 tracked build/doc-config 文件及未忽略 untracked 文件变更时，摘要必须变化。

### P2：重复安装、历史重建和全量门禁使局部复验成本偏高

**已验证：** 每次完整门禁在最前执行两次旧版本 race/HTTP 重建；随后 `frontend-test` 和 `browser` 各自依赖 `verification-deps`，因此同一候选中 `npm ci` 执行两次；再执行全仓 `go test -race -v`、延迟实验和全浏览器旅程（`Makefile:11-22,36-47`）。公开目标提供 `format-check`、`vet`、`frontend-test`、`test`、`build`、`browser`、`health-history`，但没有统一的阶段选择/从失败阶段复验接口，`review-history` 也没有 Make target。

**影响：** 精确首错诊断是好的，但修复单一晚期失败后只能手工拼目标或重跑前面昂贵阶段；重复 `npm ci` 增加网络/cache波动和时间，25 分钟总预算可能让晚期浏览器阶段获得不稳定余量。前置报告的六轮尝试体现了该放大效应，但本次未计时，不能给出绝对成本。

**建议（工程工具，P2）：** 让一次锁定安装成为 `verify-gates` 的显式阶段并供两个消费者复用；为每阶段提供一致命名的只读目标和 machine-readable stage receipt，支持开发者局部复验。正式合入仍跑全链，不以缓存替代最终绿色。可把对象 preflight 放在昂贵执行之前，而历史重建在 preflight 成功后并行或按稳定性分层；任何并行化需保持日志分段和 fail attribution。

### P2：诊断仍缺 gate 内部的结构化阶段结果

**已验证：** 顶层 receipt 的 `AuxiliaryCommands` 只记录 inventory、git-directory、source-evidence、prerequisite 和 dependency-probe；整个 `make verify-gates` 在 receipt 中是一个 `GateExit`，内部阶段只存在文本日志和 Make 输出（`tests/verify/main.go:22-62,137-199,224-229,382-417`）。dependency probe 只在 gate 绿色后运行。

**影响：** 自动系统要判断“哪些阶段未执行、首错在哪、各阶段耗时”必须解析非稳定文本；早期失败时也拿不到依赖解析/Chromium 状态，容易把基础设施与产品红混在一起。

**建议（仓库基建，P2）：** 由小型 orchestrator 或阶段 wrapper 为每阶段记录 started/finished/exit/cancelled/duration/skipped_reason，并保留原始 stdout/stderr；前置依赖探测拆成不安装的早期 probe 与绿色后的完整 provenance probe。

## 已验证、推断与未执行汇总

- **已验证：** 当前候选及三项 integration SHA；两个基线 commit 当前 `cat-file -t` 为 commit；仓库为 `blob:none` partial clone；入口顺序、超时、短路、清理、摘要 scope、receipt schema；前置报告确实引用三条最终绿色本地 receipt 路径，也保留多轮 baseline unavailable 记录。
- **推断：** Make 串行首错停止、对象缺失会在当前测试之前阻断、未统一哈希的产物可脱离顶层 receipt 改写、两次 `npm ci` 和重复全链会增加复验成本。均来自明确控制流，但本次没有动态复现或计时。
- **未执行：** `make verify`/`verify-gates`、任何 Go/Node/npm/browser/latency 命令、历史归档和重建、前置 receipt 内容校验、真实供应商联调。静态审查不构成任何门禁绿色证明。

## 优先级与归属清单

| 优先级 | 问题/建议 | 主要归属 | 用户/交付影响 |
|---|---|---|---|
| P0 | 无。关键入口均存在且可读；未发现静态上必然绕过所有门禁的路径 | — | 不代表门禁已通过 |
| P1 | 固定历史对象离线化或显式 preflight；平台预取并出 hydration receipt | 仓库基建 / 通用平台 | 避免 partial clone 设施红阻断当前验证 |
| P1 | 持久化内容寻址 evidence index，平台接管 artifact ID 与保留 | 仓库流程 / 通用平台 | 恢复跨 worktree 的源码—收据—完整日志链 |
| P1 | 摘要覆盖完整候选 inventory、mode 和明确排除集 | 工程工具 | 防止 `source_unchanged` 超出真实 scope |
| P2 | 单次 npm 安装、统一阶段目标、可局部复验但最终仍全跑 | 工程工具 | 降低重复运行成本和晚期失败重试成本 |
| P2 | gate 内部阶段结构化 receipt/skip 原因 | 仓库基建 | 提升自动诊断与设施红/产品红分类 |

最终判断：固定入口的隔离、严格历史红分类、首错传播和失败收据值得保留；最需要优先修复的是 **partial-clone 下历史对象的可获得性契约** 与 **绿色证据跨 worktree 的持久可验证性**。在这些问题解决前，前置报告可以作为执行陈述，但不能替代可携带、可独立校验的完整证据链。
