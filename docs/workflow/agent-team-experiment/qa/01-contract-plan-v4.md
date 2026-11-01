# 后续独立 QA 合同执行计划（v4，唯一有效版）

## 1. 效力、输入与裁决边界

本文取代 v3，成为后续 QA 的唯一执行计划。`01-contract-plan.md`、v2、v3、`02-readiness-v3.md` 及更早评审均只读保留为历史输入；v2 的 `03-execution-report.md` 继续废止。后续唯一测试文件是 `internal/relay/qa_independent_regression_test.go`，Git 中唯一执行报告是 `docs/workflow/agent-team-experiment/qa/03-execution.md`。

本版只读核对的公共候选 HEAD 是 `2aeeb3af8db73a33f6fcb99371769b79f865e367`。控制器回执和公共历史确认 v3 的公共集成提交 `d1893ec08737074df8ee14a59d8da59413dc75e3` 与时钟设施公共集成提交 `2aeeb3af8db73a33f6fcb99371769b79f865e367` 均已合入；私有 worker candidate 不要求成为公共 HEAD 的祖先。输入摘要：v3 `3e74a646…`、readiness v3 `2e88dbd6…`、时钟设施源码 `31f13f92…`、设施报告 `8e793dce…`（均为 SHA-256）。正式依据仍为 `docs/product-brief.md`、`docs/releases/v0.1.0.md`、`docs/workflow/requirements.md`、`docs/workflow/design.md` 和 product-definition；真实入口包括 `Server.ServeHTTP`、`Store.advanceClock`、`Store.admit`、`Server.beginAttempt`、`Server.acceptInvitation`、`Backup`、`Restore/restoreWithRename`。

计划完成只表示 v4 可供独立 readiness 评审，不表示 READY、产品通过或允许执行。独立评审明确批准前不得执行；执行者不得自批。Bearer 与 Restore 的已知产品缺陷允许产生 Fail/NO-GO，它们不是 readiness 障碍，也不得被弱化或删除。

## 2. 分层同版模型（机械门禁）

### 2.1 三层对象

1. **获审实质候选 P：**独立评审记录产品/构建 source manifest 的 SHA-256，以及能代表该内容的公共祖先产品 commit。P 覆盖全部产品 Go/前端源码、`go.mod`、`go.sum`、schema/migration、Makefile、`scripts/` 和一切 build/verify/gate 配置；按 tracked path、Git mode、blob ID、文件 bytes 与 SHA-256 排序形成 manifest。P 不包含控制文档，也不要求包含计划文档的 HEAD 等于较早产品 SHA。
2. **控制层 C：**获审后允许合入的白名单只能逐项记录精确路径和内容 SHA-256：本计划、后续 readiness、唯一执行报告等控制文档；不得以“仅文档变化”概括。控制层不得改变 P 中任何路径或门禁语义。
3. **执行候选 E：**开始执行时的真实 `git rev-parse HEAD`，必须以 Git 祖先检查证明包含独立评审所指定的公共基点，并由路径级 diff/manifest 证明，相对获审 P 的全部差异仅为已登记 C 和唯一获批测试文件 `internal/relay/qa_independent_regression_test.go`。不在本计划硬编码尚未产生的未来 SHA。

### 2.2 执行前冻结和停止条件

执行开始先记录：真实 E HEAD；`git status --porcelain=v1 --branch`；公共基点到 E 的祖先检查命令/退出码；P 的产品/构建 manifest 与摘要；白名单每个精确路径、bytes/SHA-256/blob；唯一 QA 文件的完整 diff、bytes/SHA-256/blob；由 P 加该最终 QA 文件得到的**最终候选源指纹 F**。报告文件在运行前可不存在；创建后只按登记的控制路径变化处理，不进入 F。禁止未跟踪的产品/构建输入。

所有 A/C/B/X 专项及原样 `make verify` 必须在同一 E worktree、同一 F 上运行；正式矩阵开始后 QA 文件冻结。每条 receipt 同时绑定 E、P-manifest、QA 摘要和 F。运行后重算，除登记中的报告内容外必须逐字节一致。任何产品/构建内容变化、依赖/schema/门禁变化、非白名单路径、白名单摘要不符、manifest 不一致、QA 文件之外的测试变化或运行间 F 改变，立即停止、证据作废并回到独立计划评审；不得自行解释或续跑。

## 3. 写入、隔离与 receipt 合同

执行阶段只可修改上述唯一 QA 文件和唯一 Git 报告。完整 stdout/stderr、截图、SQLite、archive、临时 workspace、密钥夹具及 Native/会话/运行历史，只能进入执行时用 `git check-ignore -q` 同时确认目录和具体文件均 ignored 的 `test-results/qa-contract/<F>/<run-id>/`，或控制器私有目录；绝不写入 Git。Git 报告只存 receipt ID/相对来源索引、脱敏命令、UTC 起止时间、退出码、stdout/stderr 及关联产物各自 bytes/SHA-256、E/P/F 同版来源和脱敏断言摘要，不写绝对私有路径或原件。

每次命令不使用吞退出码的管道；stdout/stderr 分文件，先记 start、结束立即记 end/exit/bytes/SHA-256。receipt 为 `qa-contract/<F>/<run-id>/<matrix-id>/<attempt>`，重跑不得覆盖。另记录 `go version`、`go env GOTOOLCHAIN GOOS GOARCH CGO_ENABLED`、Git/make 版本及项目实际需要的浏览器工具版本。原件缺失、摘要不符、时间不闭合或来源不一致，相关结论无效。

所有夹具只用 `t.TempDir()`、临时 SQLite/端口和进程内或临时受控 HTTP/SSE 上游；不得访问真实服务、供应商、备份、数据或凭据。通用顺序为：构造前置数据→建立观测零点→唯一被测动作→读取断言→清理。setup/cleanup 失败即 Fail。逐项报告只能为 Pass、Fail 或“执行前真实设施冲突导致 Blocked”；不得删项。

## 4. A：Bearer 正常、边界、隔离与并发

| ID | 前置、入口与顺序 | 必须断言、隔离与 receipt |
|---|---|---|
| A01 正常 | Schema 3 临时安装、有效 `mr_` key 仅 grant 一模型、上游零点；直接 handler 和真实 HTTP 分别调用 models、普通生成、SSE | 仅列授权模型；协议及 request/attempt/usage 闭合；越权或次数异常 Fail。每例新 DB/端口/upstream；`A01/<case>` |
| A02 scheme 大小写 | 每变体独立零点，对 `Bearer`、`bearer`、`BEARER` 及至少两种混合大小写调用 models 与生成 | ASCII scheme 大小写不敏感，结果与基准等价；敏感即 Fail/NO-GO。`A02/<scheme>` |
| A03 RFC 6750 `1*SP` | 有效 key、上游零点；单 SP、双 SP、多 SP 均作正向；缺头、裸 key、Basic、空 credential、前导空白、HTAB、尾随空白/附加内容作负向 | 三个正向均成功；负向均拒绝且 upstream/attempt/usage/扣费为零。逐格式隔离；`A03/<form>` |
| A04 生命周期/授权 | 有效、随机、停用、撤销、过期、轮换前后 key；两租户不相交 grant；models/普通/SSE 全组 | 仅当时有效身份成功，无 grant/租户扩张；拒绝路径无分派或记录。每状态新安装/快照；`A04/<state>` |
| A05 并发隔离 | 两租户、多种 key；所有 goroutine ready 后 barrier 同时 release，受控上游可阻塞；race-enabled 专项 | 响应、模型、记录、usage 不串租户；无效请求 upstream=0；race、超时或泄漏 Fail。禁止 sleep 猜测；`A05/concurrent` |

## 5. C：时钟合同

### 5.1 TEMP-trigger 与 barrier 公共协议

除 C05/C07 的特殊入口外，每子例用新 Schema 3 与确定 `store.clock`。观测前完成全部夹具并关闭辅助 goroutine/Rows；硬断言被测真实 `Store.DB` 且 `MaxOpenConnections==1`。在该 DB 创建 TEMP 审计表及 `AFTER UPDATE OF effective ON console_clock` trigger，安装后清零并确认 0；零点后只准一个被测请求。完成 channel 关闭后，直接 SQL 读取审计、attempt 和业务记录；最后 DROP 并断言 `sqlite_temp_master` 无残留。每行 `NEW.effective>=OLD.effective`。差分固定：外层-only=1；外层+admit/no-attempt=2；N attempts=`2+N`；邀请=2；attempt ordinal、上游命中与业务记录对账，夹具请求不得污染。

排队用 channel/barrier 建立 started→wall sample observed→推进时间→release→finished 的 happens-before，禁止 sleep；非持久 wall sample 不冒充持久 UPDATE。

| ID | 前置、入口与顺序 | 必须断言、隔离与 receipt |
|---|---|---|
| C01 外层 | 合法 key，按 5.1 分别唯一调用 models、identity GET、console GET、health | 三类非 health audit=1，health=0，业务正确；每入口独立 trigger；`C01/<entry>` |
| C02 差分 | 独立建立 admit 无候选、成功 N=1、失败链 N=3、未过期邀请；各执行唯一生成/accept 请求 | audit 分别 2、3、5、2；attempt ordinal=N=upstream；邀请正确消费；`C02/<case>` |
| C03 排队截止 | 主 goroutine 持 server mutex，请求报告 started；clock 回调报告 observed；收到后推进再 release | 次序全成立；外层 audit=1、邀请另加1；资源按新截止失效且 upstream=0；goroutine join；`C03/<key|invite>` |
| C04 防回拨 | 先推进 T1并清理，再 wall 回拨 T0<T1，重新安装/清零后唯一请求 | OLD/NEW、内存 effective 与持久值均不低于 T1，过期资源不复活；`C04/rollback` |
| C05 四阶段 fail-closed | 严格执行 5.2；每阶段独立 fixture/请求与零点 | 阶段身份、计数、持久/内存不变、503/degraded、upstream/业务为零全部满足；`C05/<phase>` |
| C06 重启/恢复 | Schema3 达 T1；项目 Backup 到临时 archive；关闭重开后回拨请求；另 Restore 到新临时目标、Open、回拨请求 | 两路径 effective≥T1、过期不复活、业务正常；`C06/<restart|restore>` |
| C07 旧 schema | 项目 API 生成 Schema1/2 最小安装，各走合法入口 | 按兼容合同工作，不要求/强装 Schema3 trigger；崩溃、错误迁移或意外时钟表 Fail；`C07/<schema>` |

### 5.2 C05 机械执行合同

真实包内设施位于 `clock_persistence_fault_fixture_test.go`。类型/入口精确为：`clockFaultStage`；参数 `clockFaultBegin`、`clockFaultRead`、`clockFaultUpdate`、`clockFaultCommit`（正常校准另用 `clockFaultNone`）；构造 `newClockFaultFixture(t, stage) *clockFaultFixture`；真实调用 `fixture.store.advanceClock()`；清理由构造函数注册的 `t.Cleanup` 执行 `store.Close()`，临时安装由 `t.TempDir()` 回收。计数读取为 `fixture.state.counts.begin/read/update/commit.Load()`；commit 专有计数为 `fixture.state.commitHookHits.Load()`；前三阶段错误 sentinel 为 `errClockFault`，commit 错误必须 `errors.As` 为 `sqlite3.Error` 且 `ExtendedCode==sqlite3.ErrConstraintCommitHook`。

每个阶段先新建 fixture，固定 `fixture.store.clock` 为 Tnew，直接 SQL 读取并保存原持久 `console_clock.effective`，并在锁保护下保存 `store.effectiveTime`；再建立 request/attempt/usage/其他业务表及受控 upstream 的计数零点。因为基线 SQL 也经过包装连接，完成基线后必须对四个 atomic 计数逐一 `Store(0)`、确认 `[0,0,0,0]` 且 `commitHookHits==0`，然后才进入正式零点。正式动作只能调用一次真实 `fixture.store.advanceClock()`；随后立即快照计数和错误，再用独立只读查询核对持久值（该辅助查询发生在计数快照之后，避免污染 read 计数），锁保护下核对内存值，最后核对业务/upstream 仍为零。直接设施调用没有 handler 响应，因此不得虚构 503；同一阶段另以 `New(fixture.store, Config{...})` 构造真实 Server，在零点前完成合法身份/路由数据，随后通过 `ServeHTTP` 发出唯一非-health 请求，核对 503 `storage_unavailable`、`server.degraded.Load()==true`、handler 停止、upstream/attempt/usage/业务零点。两子例都必须绑定同一 stage receipt，前者证明机械阶段，后者证明 fail-closed；无法以真实 `New`/`ServeHTTP` 承载该 Store 时 C05 保持 BLOCKED，不能以设施自测替代。

| 故障 | 唯一调用顺序与期望计数 `[begin,read,update,commit]` | 错误、持久/内存与零点断言 |
|---|---|---|
| begin | 构造 `clockFaultBegin`→设 Tnew→取旧值/零点→一次 `advanceClock`→计数快照；`[1,0,0,0]`，commitHook=0 | `errors.Is(err,errClockFault)`；持久值与 `effectiveTime` 均等于调用前；业务/upstream=0 |
| read | 同序，参数 `clockFaultRead`；`[1,1,0,0]`，commitHook=0 | 同 sentinel；真实 begin 已完成但 read 返回失败，事务 rollback；持久/内存不变，业务/upstream=0 |
| update | 同序，参数 `clockFaultUpdate`；`[1,1,1,0]`，commitHook=0 | 同 sentinel；begin/read 已完成，仅 UPDATE 返回失败；持久/内存不变，业务/upstream=0 |
| commit | 同序，参数 `clockFaultCommit`；`[1,1,1,1]` 且 commitHook=1 | 必须为 `SQLITE_CONSTRAINT_COMMITHOOK`，不得接受 sentinel 或其他阶段错误；计数机械证明 begin/read/update 已完成，且仅真实 COMMIT 失败；SQLite 回滚后持久值不变，`effectiveTime` 未发布，业务/upstream=0 |

每阶段使用独立临时 DB、连接池、driver connector 和状态，不共享全局可变 hook，不 sleep、不改权限、不关闭 DB 冒充阶段。正常校准 `clockFaultNone` 必须在正式故障前单独证明快照计数 `[1,1,1,1]`、hook=0、无错误、持久值=Tnew、内存 effective=Tnew；校准 receipt 为 `C05/normal-delegation`，不能替代四个失败项。任何目标计数不符、非目标计数非上表值、错误身份不符、持久/内存改变、业务/upstream 非零，均为 Fail；若设施标识/SQL匹配或 commit-only 证明因源码变化不可用，则计划执行结论保持 BLOCKED并回独立评审，不得换用 begin/update/关闭 DB 等替代。

## 6. B：备份、manifest、恢复

| ID | 前置、入口与顺序 | 必须断言、隔离与 receipt |
|---|---|---|
| B01 正常往返 | 项目 API 生成 Schema1/2/3（v3 effective=T1）；`Backup`→只读 tar/manifest→`Restore` 到 absent temp target→`Open` | 恰有普通 manifest/DB，digest/version/schema 一致，数据/T1 守恒；`B01/<schema>` |
| B02 manifest | 只改 B01 测试副本 manifest：重复 version/SHA、未知键、尾随值、类型/缺字段/非对象、空短大写非hex digest；另验合法空白/字段顺序 | 非法均安装前拒绝且不泄密，合法均成功；每例新 target；`B02/<case>` |
| B03 tar | 覆盖重复/未知/缺 entry、非普通类型、路径名、负/超限 size、截断/校验、version/schema 不符及顺序 | 非法均安装前拒绝且无半安装，允许顺序成功；`B03/<case>` |
| B04 普通失败原子性 | 每个 B02/B03 负例用 absent 与 existing-empty；existing 预存 FileInfo/SameFile、mode、排序清单、内容摘要、父目录清单 | absent 仍 absent；existing 身份/权限/内容/清单不变；无 relay.db、替换、删除或 `.restore-*`；`B04/<case>/<state>` |
| B05 最终 rename | 合法 Backup；调用真实 `restoreWithRename(input,directory,keyFile,rename)`；absent/existing-empty 各 10 次；闭包仅允许同一 temp parent 的 `.restore-*`→精确 target，恰一次后返回唯一 sentinel | `errors.Is` 保真；existing 身份/权限/内容/清单完全不变且无 relay.db，absent 仍 absent，临时目录为0。当前 rename 前 `os.Remove(directory)` 导致 existing 预计 Fail/NO-GO；不得 Not Run/N/A 或以设施测试替代。`B05/<state>/<iteration>` |

B05 禁止 chmod/goroutine race、root/mount/FUSE 或全局 hook。若签名或注入点不再仅替换最终 rename，执行前以源码/编译证据报 Blocked并回独立评审，不改 `backup.go`。

## 7. X：跨改动与固定全量

| ID | 前置、入口与顺序 | 必须断言、隔离与 receipt |
|---|---|---|
| X01 恢复×时钟×认证 | B01 恢复 Schema3、effective=T1，回拨 wall；用大小写变体和单/双/多 SP 调 models、普通、SSE | 高水位连续、合法认证等价、grant 不扩大、attempt/usage 闭合；`X01/<form>/<entry>` |
| X02 时钟失败×认证 | 对 C05 四阶段分别跑合法/非法 Authorization；先确认机械命中，恢复 DB 后复跑 A02/A03 基准 | 时钟失败优先 503、degraded、upstream/attempt/usage/扣费=0；恢复后合法成功、非法拒绝；`X02/<phase>/<form>` |
| X03 固定全量 | QA 文件及 F 已冻结，前述 receipt 完整；在同一 E/F worktree **原样**执行一次 `make verify` | 仅 exit 0 且运行后 F 一致为 Pass。即使专项失败也必须运行并保留真实结果；设施/专项绿不可替代；`X03/make-verify` |

## 8. 执行报告与最终裁决

`03-execution.md` 必须含：独立评审授权；E/P/F 与祖先/manifest/白名单/QA diff 的前后对账；工具链；A01–A05、C01–C07、B01–B05、X01–X03 的实际前置、入口、顺序、断言、失败、隔离和 receipt；缺陷摘要；原样全量结果；禁入物检查；最终裁决。

最终硬规则：Bearer scheme ASCII 大小写不敏感，RFC 6750 `1*SP` 的单/双/多 SP 都是正向；TEMP trigger 绑定真实单连接 Store.DB 且保持冻结差分；排队使用 barrier；C05 必须按四阶段真实设施和 commit-only 证明执行；Restore 最终 rename 的 existing target 硬断言不可弱化；所有专项和原样 `make verify` 绑定同一 F。任一必跑项未执行、证据不可复验、同版门禁失败或断言失败，均不得裁为 GO。

本计划任务只进行了源码、文档与 Git 的静态研究并新增本文；没有运行产品/专项测试或 `make verify`，没有启动服务或访问真实备份、数据、凭据、供应商。本文须经独立评审；作者不裁 READY。
