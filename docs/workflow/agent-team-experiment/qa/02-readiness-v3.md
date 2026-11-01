# v3 QA 计划最终可执行性独立评审

## 1. 唯一结论

**`BLOCKED`**。

该结论只表示 `docs/workflow/agent-team-experiment/qa/01-contract-plan-v3.md` 尚不能作为无歧义、可复验的执行合同；不表示当前产品已通过或未通过 QA。已知 Bearer 实现会拒绝 RFC 合法形式、existing target 会在最终 rename 失败时丢失，均应由正确计划执行后得到 `Fail/NO-GO`，本评审没有把这些预期产品失败当作 readiness blocker。

真实阻断有两项：C05 没有为时钟事务的 begin/read/update/commit 四阶段给出可执行的确定性故障入口，尤其当前源码和既有夹具中不存在 commit 阶段入口；计划冻结的候选 `aade8d2...` 又不是包含 v3 的实际评审基线 `d1893ec...`，而运行前规则同时要求 HEAD 等于“获评审 SHA”，导致执行者无法无歧义确定候选与 source manifest。前者会造成漏测或由执行 Agent 临时发明设施，后者会使同版证据无法复验。

本评审仅静态读取 Git 历史、文档和源码；未运行产品测试、专项测试或 `make verify`，未启动服务，未访问真实数据、备份、凭据或供应商，也未修改计划、产品、设施或测试。

## 2. 候选、合入关系与输入

- 本评审 worktree 的 `HEAD` 为 `d1893ec08737074df8ee14a59d8da59413dc75e3`，开始检查时 `git status --short` 无输出。
- `git merge-base --is-ancestor` 静态核对均返回 0：v2 集成 `ef379a26fbe7c6159134a01a6fb0164003c33ef8`、Restore 设施集成 `aade8d2f72d42ca66349d6127cfd78a3a012b1cf`、v3 产出 `fb9645f8ffda321683638272e288f65f27d97bd5` 均为当前 HEAD 祖先。两项任务依赖的控制器状态均为 `merged`，开始条件满足。
- 提交 `426ee922082461eb3b907013b5d70a36c599ddb1` 中历史 BLOCKED 评审与当前 `02-readiness-review.md` 的 SHA-256 均为 `ca5263b8b83ca8e5b5d23ff768ccb81d0fba7cb49befd90320336129b11893bf`，历史输入未被改写。
- 已读输入包括 v3、v2、历史 BLOCKED 评审、`restore-testability.md`、正式 product brief、v0.1.0 release、requirements、design、product-definition，以及真实 `backup.go`、`store.go`、`console_clock.go`、`server.go`、认证/路由实现和相关既有回归。

## 3. 正式合同到证据的逐项追踪

| 项目 | 正式合同 | 实现/设施入口 | v3 断言与证据形式 | 评审结果、严重度与返工条件 |
|---|---|---|---|---|
| Restore rename seam | 任一恢复失败不得删除、替换或半恢复目标；错误与临时清理可观察 | `Restore` 显式调用未导出的 `restoreWithRename(..., os.Rename)`；rename 依赖是函数参数；仅最终 `.restore-*`→target 使用该参数；没有包级可变 hook。真实代码仍在 rename 前 `os.Remove(directory)` | B05 用唯一 sentinel、调用次数/参数、`errors.Is`、absent/existing 各 10 次、身份/mode/清单/摘要、无 `relay.db`/替换/残留；receipt 逐迭代索引 | **满足 / 无 blocker**。计划可确定性暴露 existing-target 的真实产品 Fail；不得以此阻止执行。若签名或注入位置变化，才按 v3 规定成为执行前设施冲突 |
| TEMP trigger 连接与差分 | Schema 3 每个非 health 外部请求外层持久推进一次；admit、每 attempt、邀请内部推进另计 | `database()` 设置 `SetMaxOpenConns(1)`；真实 `fixture.store.DB` 即产品 DB；`advanceClock` 的持久 UPDATE 在该连接池事务中发生 | §5.1 明确同一真实 DB、MaxOpenConnections=1、安装→清零并验 0→唯一请求→完成 channel→直接 SQL 读取→DROP 并验清理；公式为 1、2、2+N、2，attempt/上游/ordinal 对账 | **满足 / 无 blocker**。连接不会漂移到另一物理连接；零点后禁止辅助 HTTP，差分不会把夹具动作计入。各场景独立 DB，不能仅以业务记录替代 trigger 行 |
| 排队与截止 | 请求排队后必须观察新的截止时间；并发次序确定 | `ServeHTTP` 在 `TryLock` 失败时先调用非持久 `wallTime`，随后阻塞于 `server.mu.Lock`；既有 queue 回归有相同可控入口 | C03 主 goroutine 持锁；clock 回调发 `observed`；收到后推进时间再 release；等待 finished 后读取；明确禁止 sleep | **满足 / 无 blocker**。channel/barrier 建立了 sample→推进→解锁→持久推进的 happens-before；wall sample 与 audit UPDATE 分列 |
| Bearer 协议 | scheme ASCII 大小写不敏感；RFC 6750 `1*SP` 接受一个或多个 SP；非法 credential 仍拒绝 | 当前 `CutPrefix("Bearer ")`、`HasPrefix("Bearer mr_")` 等精确实现可能失败 | A02 保留多种大小写；A03 把单/双/多 SP 均列正向，HTAB、前导/尾随/附加内容列负向；X01/X02 跨场景复验；失败即 Fail/NO-GO | **满足 / 无 blocker**。当前产品可能 Fail 不是计划不可执行理由，执行 Agent 无权弱化断言 |
| 唯一写路径 | 执行不得临时扩展测试与报告范围 | 唯一测试文件尚不存在，符合后续新增语义 | 唯一测试路径精确为 `internal/relay/qa_independent_regression_test.go`；唯一 Git 报告精确为 `docs/workflow/agent-team-experiment/qa/03-execution.md`；v2 的 `03-execution-report.md` 明确废止 | **满足 / 无 blocker**。运行前后出现其他 tracked 执行变更即范围失败 |
| 原始证据隔离 | 原始日志、截图、SQLite、归档、workspace/native 历史不得进入 Git | `.gitignore` 的 `test-results/` 对目录和具体示例文件均由 `git check-ignore -v` 静态确认命中 | 原件仅进入执行时再次确认 ignored 的 `test-results/qa-contract/<sha>/<run-id>/` 或控制器私有目录；Git 报告只含 receipt、bytes、SHA-256、exit、时间、同版来源和脱敏摘要，禁止绝对私有路径 | **满足 / 无 blocker**。原件缺失或哈希/字节不符即证据无效 |
| 正常、边界、隔离、并发、恢复、跨改动 | 全矩阵必须有前置、入口、操作、断言、失败、隔离和 evidence；固定全量不可替代 | A/B/C/X 对应真实 handler、Store、受控 upstream、Backup/Restore；`make verify` 为项目固定入口 | A01–A05、C01–C07、B01–B05、X01–X03 均给出七字段表；专项不论成败仍原样执行一次 `make verify`；正式运行中不得改 QA 文件 | **部分满足，但被下述 C05 阻断**。其余正常/边界/隔离/并发/恢复/跨改动与 X03 均有明确失败判定及同版 receipt |
| C05 时钟写失败 | begin/read/update/commit 每阶段均须确定性命中并证明 fail-closed，不得以一个阶段替代其他阶段 | 当前 `Store.advanceClock` 直接使用具体 `*sql.DB/*sql.Tx`，没有参数 seam 或数据库 driver fault seam；现有回归仅通过关闭 DB 命中 begin。普通 SQLite trigger 可确定命中 UPDATE；删除/破坏表可造成 read 失败，但源码/既有夹具没有确定的 commit-only 故障入口 | C05 仅写“分别建立 ... 确定性 DB 故障入口/既有夹具”，没有给出每阶段的 SQL、fixture、函数或命令；又规定“无法命中特定阶段是测试设施/夹具 Fail” | **BLOCKER / High**。这不是实际入口，而是把设施设计留给执行 Agent；commit 阶段按当前允许写路径无法由既有入口确定性命中。返工须在执行前提供并核对四阶段各自的具体、隔离、可证明只命中该阶段的入口；若需产品/设施 seam，须由独立工程任务先合入，不能由 QA 临时修改产品或把未命中记作普通产品 Fail |
| 候选与同版绑定 | 每条证据必须能绑定被评审候选、最终 QA 文件及执行报告，并可前后复验 | 当前评审 HEAD 是包含 v3 的 `d1893ec...`；v3 自身不存在于其声明的 `aade8d2...` | §1 称“静态冻结于候选 `aade8d2...`”；§2.2 又要求运行前候选等于“获评审 SHA”，在同一冻结 worktree 执行并把每条 receipt 绑定 candidate full SHA | **BLOCKER / High**。执行若以 `aade8d2...` 为 HEAD，就没有 v3；若以 `d1893ec...` 或后续含本评审/QA 文件的 HEAD 执行，就违反文本中的 frozen candidate/获评审 SHA 条件。返工须唯一指定产品基线、计划/评审基线与执行 HEAD 的关系，明确允许的仅文档/QA 资产差异及其独立 manifest，给出执行时应比较的精确 SHA/祖先条件，不能让执行 Agent自行解释 |

## 4. 真实返工条件

再次申请 readiness 前必须同时满足：

1. 为 C05 的 begin、read、update、commit 分别写出当前仓库可调用的确定性入口、安装与清理顺序、阶段命中证明、业务/upstream 零点和 receipt。若 commit-only 故障需要新 seam，先由独立设施任务合入；v3 只能引用已存在入口，不能授权执行 Agent 临时扩产品范围。
2. 修正候选模型：明确产品代码候选、包含 v3/本评审的控制文档 HEAD、最终 QA 文件冻结点分别如何进入 source manifest，并写出执行前唯一可机械判定的 SHA 或 ancestor/diff 条件。报告生成导致的允许差异也须从“被测源码不变”中显式分层，而非要求 HEAD 同时等于不存在 v3 的旧提交。
3. 保持已经正确的硬合同不变：A02/A03 的大小写及双/多 SP 正向断言、§5.1 单连接 TEMP trigger 协议、C03 barrier、B05 的 absent/existing 原子性与错误保真、两个唯一写路径、ignored/private 原始证据边界，以及不论专项失败仍原样运行一次 `make verify`。

完成上述返工并再次独立评审为 `READY` 前，不应开始 v3 执行。业务结论为 `BLOCKED` 不影响本次静态评审文档本身按任务工作状态 `completed` 交付。
