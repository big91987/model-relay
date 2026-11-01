# QA 合同执行计划修订版（v2）

## 1. 状态、输入与边界

**本修订计划状态：可交付；产品候选仍预期 `FAIL / NO-GO`；QA 执行仍为 `BLOCKED`。** 本文只修订计划，不覆盖或改写历史的 `01-contract-plan.md` 与 `02-readiness-review.md`。阻塞评审来自提交 `426ee922082461eb3b907013b5d70a36c599ddb1`，其结论继续是历史事实：在原独立评审 Agent 重新复核并明确给出 `READY` 前，不得开始 QA 执行。本文完成不表示产品通过、不自批 readiness，也不允许修改产品或既有回归使当前候选变绿。

静态研究基线为当前集成提交 `68005200c4187089065e70ea14a7a9e57e40db06`；其中产品树仍是 `122c9abece28f562616483e991098d020d236d6d`，只额外合入了原计划。输入包括 `AGENTS.md`、`docs/product-brief.md`、`docs/roadmap.md`、`docs/releases/v0.1.0.md`、正式 product-definition、原计划、上述 BLOCKED 评审、三项 Worker 文档，以及 `internal/relay` 中的真实实现与回归。Worker 绿回执只作历史输入，不能替代独立证据。

正式依据保持不变：RFC 9110 §11.1 要求 authentication scheme ASCII 大小写不敏感；RFC 6750 §2.1 的 `Bearer 1*SP b64token` 接受一个或多个 SP；RFC 5234 §2.3 规定 ABNF 字符串字面量默认大小写不敏感；SQLite `CREATE TRIGGER` §7 说明 TEMP trigger 仅对创建它的连接可见。项目可收紧 `mr_` credential 本体、生命周期和授权，但不能收紧 scheme 大小写或把 `1*SP` 改成单个 SP。

## 2. 阻塞评审逐项闭环

| 评审项 | 处理 | 验证方法 | 必须交付的证据 |
|---|---|---|---|
| R-CLOCK-01 TEMP trigger 连接可能漏计 | 采用真实 `Store.DB` 的单连接事实：`database()` 在 `internal/relay/store.go` 调用 `SetMaxOpenConns(1)`；测试先硬断言 `DB.Stats().MaxOpenConnections == 1`，再在该 DB 完成 TEMP 表/trigger 安装。不得另开 `sql.DB`、不得持有 rows/事务、不得在审计期改变连接池配置 | 夹具全部完成后安装；清零；同步执行唯一请求；用同一 DB 读审计；删除 trigger/table。若最大连接数不是 1、安装连接丢失或审计读取失败，测试失败而非跳过 | 连接池断言、审计 DDL 摘要、逐行 old/new effective、请求状态、清理后 TEMP 对象为 0 的摘要 |
| R-CLOCK-02 来源与夹具污染 | 不给同一 UPDATE 伪造来源标签；以隔离的差分场景判别：外层-only=1；外层+admit/no-attempt=2；一或多 attempt=2+N；邀请接受=外层1+邀请内部1。每案独立安装/清零并用 `request_id` 的 `request_attempts` 数 N 交叉对账 | 每案在清零后禁止任何辅助 HTTP；单请求完成 channel 关闭后才读 DB；比较相邻审计值、attempt 数与预期增量。排队另记非持久 `wallTime` sample，不混入 UPDATE | 每案名称、唯一请求/脱敏 request ID、HTTP 状态、audit count、attempt count、外层/内部推导式；任何额外请求或不明 UPDATE 都失败 |
| C-CLOCK-01 现回归只数函数调用 | 现有 `clock_advance_regression_test.go` 仅作辅助；核心结论必须来自上述持久 UPDATE 审计 | 对 models/identity/console 外层-only 请求分别证明 1；对生成和邀请证明差分 | SQL 审计摘要及辅助调用计数分栏，二者不得互相替代 |
| D-RESTORE-02 最终 rename 无确定注入 | 静态确认 `Restore` 在 `backup.go` 先 `os.Remove(directory)` 后 `os.Rename(temporary,directory)`，且当前无 filesystem/rename seam。设立第 4 节独立工程设施前置；在设施合入、重新冻结并复核前 QA 保持 BLOCKED，禁止 `Not Run`、删例或伪造 Pass | seam 就绪后令最终 rename 确定返回 sentinel error，只作用于 `t.TempDir()` 父目录；重复运行并硬断言原目标身份/权限/内容仍在、未部分替换、无 `.restore-*` | 注入命中一次、sentinel error、父目录前后清单及元数据摘要、目标内容摘要、残留为 0；当前代码预期会暴露目标丢失并判 Fail/NO-GO |
| R-EVIDENCE-01 原始证据误入 Git | 原始 stdout/stderr、截图及运行产物只写 Git 忽略的 `test-results/qa-contract/<candidate>/...` 或控制器私有运行目录；Git 只写脱敏报告/哈希/receipt 索引 | 执行前用 `git check-ignore` 验证项目目录；执行后逐文件记录 bytes/SHA-256，并核对 Git status 无禁入物 | 相对 receipt ID、原始产物相对路径或控制器 receipt ID、bytes、SHA-256、退出码、时间；不登记不可移植绝对路径 |
| R-SCOPE-01 测试写路径歧义 | 后续 QA 唯一测试写路径冻结为 `internal/relay/qa_independent_regression_test.go`；不得修改三个既有产品回归，也不得新增其他测试文件。执行报告固定为 `docs/workflow/agent-team-experiment/qa/03-execution-report.md` | 执行前后用 Git path diff 对账；任何其他源码/测试写入立即停止并判执行范围失败 | before/after path manifest、候选 SHA、唯一测试文件摘要和报告摘要 |

## 3. 时钟可执行协议

### 3.1 审计夹具与严格边界

每个时钟子例使用新的 `t.TempDir()`、Schema 3 安装、确定性 `store.clock`、受控 transport/临时端口。先通过产品/既有测试夹具完成管理员、租户、key、模型、候选、邀请等前置数据；夹具阶段不取计数。确认所有 `Rows`、事务及辅助 goroutine 已结束，硬断言 `fixture.store.DB.Stats().MaxOpenConnections == 1`，然后经 `Store.DB.Exec` 在唯一连接上创建 TEMP 审计表和 `AFTER UPDATE OF effective ON console_clock` TEMP trigger，记录自增序号及 `OLD.effective`、`NEW.effective`。安装后执行 `DELETE FROM temp_clock_audit` 并确认 count=0；这就是审计零点。

零点以后至读取前只能发生一个被测 HTTP 请求；禁止登录、创建夹具、健康探测、详情查询或重试等辅助 HTTP。同步调用直接返回即为完成边界；并发/actual HTTP 用 `started`、`release`、`finished` channel/barrier，只有收到 `finished` 才读取。读取只允许直接 SQL 查询 TEMP 审计和业务表，不经过 handler，因此不会引入外层推进。读取完立即 `DROP TRIGGER`、`DROP TABLE`，确认 `sqlite_temp_master` 中对象为 0；最后关闭临时安装。安装、清零、请求、读取、清理任一步失败均为 Fail，不能用调用计数补证。

### 3.2 来源可判别场景

| 场景 | 前置数据与唯一入口 | 预期持久 UPDATE | 断言与失败判定 |
|---|---|---:|---|
| C01a 外层-only | 合法 key；`GET /v1/models`（另以独立子例覆盖 identity/console GET） | 1 | HTTP 预期成功，audit 恰 1；`/healthz` 独立子例为 0。多/少一行均失败 |
| C02a admit/no-attempt | 合法生成身份，但构造确定性准入后无可分派候选的受控业务路径；`POST /v1/chat/completions` | 2 = 外层1 + `Store.admit`1 | `request_attempts` 为 0；audit 恰2。若当前夹具不能保证走过 admit，以业务记录/响应码证明路径，否则夹具失败，不得猜测 |
| C02b N attempts | 合法单候选成功 N=1；另有受控失败链产生 N=3；同一生成入口 | 2+N | 用响应 `X-Request-ID` 直接查询 `request_attempts`，ordinal 连续且 N 与受控上游命中相同；audit 分别3、5。`beginAttempt` 每次调用 `Store.advanceClock`；不以总数单独宣称来源 |
| C02c 邀请内部 | 已创建且未过期邀请；唯一 `POST /admin/api/invitations/accept` | 2 = `ServeHTTP` 外层1 + `identity.go` 接受事务前内部1 | 接受结果与邀请状态一致且 audit=2；任何夹具创建请求必须在安装/清零前完成 |
| C03 排队 | 预置恰好跨越的 rotation/invitation deadline；持有 `server.mu`，启动唯一请求 | 外层持久1；邀请路径另加内部1 | `store.clock` 首次非持久采样通过 `observed` channel 通知已排队；主 goroutine 推钟并释放锁；`finished` 后断言 audit。禁止 sleep 判断排队。非持久 sample 与持久行分栏，资源过期且上游0 |

审计行必须逐行满足新值不小于旧值；C04 将墙钟从 T1 回拨到 T0 后，内存 `Store.now()` 与数据库 effective 均不低于 T1，过期资源不复活。C05 分别覆盖 begin/read/update/commit 失败：503、degraded、handler 不继续、上游0、无部分业务写；不能只以关闭 DB 的 begin 失败代表全部。C06 关闭重开与项目 `Backup`→`Restore` 后均保持 T1；C07 Schema 1/2 明确无 `console_clock` 更新要求。所有恢复/故障用独立临时安装。

## 4. 恢复最终安装失败：工程设施前置任务

当前 `Restore(input,directory,keyFile)` 直接调用 `os.Remove`/`os.Rename`，QA 文件无法仅靠普通权限、同一临时父目录且无竞态地保证“remove 已成功而 rename 必失败”；chmod、竞争 goroutine、跨文件系统或环境特权均不确定或越界。因此 readiness 不得预设 READY，QA 执行必须等待以下独立工程设施任务合入：

- **最小入口与修改路径：** 仅修改 `internal/relay/backup.go`，新增未导出的参数化 helper（例如 `restoreWithInstall(..., install func(temporary,directory string) error)`）；公开 `Restore` 仍以真实安装函数调用 helper。测试可把最终安装函数替换为返回唯一 sentinel error 的闭包。依赖通过参数传入，禁止可变包级全局，避免并行测试串扰；不得新增 HTTP/CLI/debug API。
- **设施验收：** 工程任务自有测试证明正常 `Restore` 行为不变；注入闭包只接收位于同一个 `t.TempDir()` 父目录内的已验证 temporary/target，恰调用一次并返回 sentinel；`go test -race` 不发生跨例污染。设施回执只证明 seam 可用，不证明恢复原子性通过。
- **QA 硬断言：** 创建合法项目 Backup 与 existing-empty target，在 target 放置不含秘密的 sentinel 文件并记录目录身份（允许的平台 inode/file ID）、mode、entry 清单和内容 SHA-256；调用注入路径。必须返回 sentinel error；原 target 必须仍存在且身份、mode、sentinel 摘要完全相同，不含 `relay.db`，父目录无 `.restore-*`。再以 absent target 运行，失败后仍 absent、无残留。每种状态至少重复 10 次以证明确定性，任一次不同即 Fail/NO-GO。
- **非目标：** 不改变 manifest/tar 合同，不扩大公开 API，不实现通用虚拟文件系统，不依赖 root/mount/FUSE/真实备份，不顺带修复 Bearer 或时钟，不把失败改写为成功。若工程任务同时修复安装算法，那是独立产品变更，必须新候选重新冻结并由独立评审裁决；QA 仍运行上述失败注入断言。

设施缺失、注入未命中或断言无法观测时，状态只能是 **执行前 BLOCKED**；不得把 B05 标成 N/A/Not Run，也不得移除该门禁。按当前先删除目标再 rename 的代码，注入预计将使 existing target 消失，这是候选缺陷的预期 Fail，不是测试设施失败。

## 5. 执行矩阵（前置、步骤、断言、隔离）

| ID | 前置数据/夹具 | 入口与步骤 | 硬断言 / 失败判定 | 隔离与 receipt |
|---|---|---|---|---|
| A01–A03 | 临时 Schema3、一把仅授权单模型的合法 `mr_` key、零命中受控上游 | handler 与 actual HTTP：合法形式；`Bearer/bearer/BEARER/混合大小写`；单/双/多 SP；缺失、裸 key、Basic、空值、前导、HTAB、尾随、附加内容 | scheme 大小写变体与合法 credential 结果相同；双/多 SP 与单 SP 相同。其他非法结构拒绝且上游0。当前 `controlledAppID` 的精确 `CutPrefix("Bearer ")` 及现回归会令规范正例失败，必须记 Fail/NO-GO | 临时 DB/端口；receipt `A01-A03`，不得打印 key |
| A04–A05 | 合法/随机/停用/撤销/过期/轮换 key；两租户/grant | models、普通/SSE、并发合法/非法请求 | 仅有效 generation 成功；无越权；拒绝请求不分派/扣费；race 无报告 | 受控上游，零供应商；`A04-A05` |
| C01–C07 | 第3节各独立夹具 | 严格按安装→清零→单请求→读取→清理；排队用 channel/barrier；故障、回拨、重启/恢复分别运行 | audit 数、attempt 数、HTTP/业务结果均满足第3节；任何漏计、污染、回拨、失败后继续均 No-Go | 每案新 temp；`C01-C07` SQL 仅脱敏摘要 |
| B01 | 项目 API 生成 Schema1/2/3 最小合法安装，Schema3 含 T1 | `Backup`→检查 tar→`Restore`→Open | 两个普通 entry、规范 manifest/摘要，状态与 T1 守恒 | 每版新 temp；`B01` |
| B02–B03 | 从 B01 仅改测试 archive | 重复/未知/尾随/类型/缺失/非对象/非规范 digest；重复/未知/缺 entry、非普通/路径/大小/截断/校验/version/schema/顺序 | 非法均在目标安装前拒绝；合法 whitespace/允许顺序成功；错误不泄密 | 每 case 新 archive/target；`B02-B03` |
| B04 | B02/B03 每案各运行 absent 与 existing-empty sentinel target | 失败后比较目标及父目录快照 | absent仍absent；existing仍同一状态；无 relay.db、替换、删除或 `.restore-*` | 仅 temp parent；`B04` |
| B05 | 第4节 seam 已合入；合法 archive；absent/existing target | 最终安装闭包确定返回 sentinel，重复10次 | 第4节全部硬断言；不得 Not Run。当前代码若目标丢失即 Fail/NO-GO | 仅 temp parent；`B05` |
| X01–X02 | 恢复后的 Schema3 key/grant/T1；另设时钟写失败 | 回拨启动后用大小写变体和多 SP 调 models/生成；写失败时用各 Authorization | 高水位连续，RFC 认证不变且 grant 不扩大；时钟失败优先503、上游0，恢复后再按 A02/A03 | 受控 HTTP；`X01-X02` |
| X03 | 已冻结候选和源码摘要 | 原样 `make verify`，不得吞退出码 | 全阶段 exit0 才通过；当前 A02/A03/B05 任一失败即总体 NO-GO；旧 Worker 回执不可替代 | 私有原始 receipt；Git 仅索引 |

真实供应商、真实备份和 UI 新视觉均不适用：合同可用本地确定性夹具验证，且三项不改 UI；`make verify` 仍负责既有 UI 回归。manifest JSON 解析没有独立并行语义，目标变化/最终安装边界由 B04/B05 覆盖。

## 6. 证据、同版对账与唯一写路径

后续 QA 执行开始前必须同时满足：独立评审将本计划判为 READY；第4节工程设施已合入并重新冻结候选；执行任务写权限明确包含且仅包含 `internal/relay/qa_independent_regression_test.go`、`docs/workflow/agent-team-experiment/qa/03-execution-report.md` 和 Git 忽略的私有产物目录。本文任务不创建这些文件。

执行 Agent 只能把全部新 QA 用例写入 `internal/relay/qa_independent_regression_test.go`。禁止修改 `bearer_auth_regression_test.go`、`clock_advance_regression_test.go`、`backup_manifest_regression_test.go` 或任何产品文件，禁止新增第二个测试文件/runner。正式报告只写 `docs/workflow/agent-team-experiment/qa/03-execution-report.md`。

原始根目录固定二选一：项目内 `test-results/qa-contract/<candidate-full-sha>/<run-id>/`（执行前必须由 `git check-ignore` 证明忽略），或控制器提供的运行私有目录。完整 stdout、stderr、截图、SQL/tar 临时产物和 `make verify` 内部 receipt 只在原始根目录；Git 中报告仅存脱敏摘要与索引。严禁提交 SQLite、key/credential、截图、完整日志、workspace 副本、临时 archive、会话/原生历史或本机绝对路径。

同版步骤如下：

1. 记录 UTC 起始时间、candidate full SHA、branch/status、`git diff --check`、被测源码与唯一 QA 文件逐文件 SHA-256，形成 source-manifest；任何候选或摘要变化均废弃旧运行。
2. 每条命令生成独立 receipt ID（例如 `qa-contract/<sha>/<run-id>/C02b`），保存原始 argv/工作目录的脱敏相对表示、开始/结束 UTC、exit code、stdout bytes/SHA-256、stderr bytes/SHA-256、关联产物 bytes/SHA-256。命令不得通过会吞掉退出码的管道。
3. 先执行 A02/A03 暴露当前预期缺陷，再按 A01–A05、C01–C07、B01–B05、X01–X02；设施缺失立即 Blocked，不继续伪造覆盖。若控制器仍要求全量，再在同版运行原样 `make verify`。
4. 运行后重采 candidate SHA、status 和 source-manifest；必须逐字节匹配步骤1（只允许预声明的唯一 QA 文件、报告和 ignored receipts）。报告逐矩阵登记 Pass/Fail/Blocked，B05 不允许 Not Run。
5. Git 报告中的每个索引行包含：candidate SHA、source-manifest SHA-256、命令 ID/脱敏命令、起止 UTC、exit code、stdout/stderr bytes 与 SHA-256、相对 receipt ID、脱敏结论。控制器私有目录仅登记其稳定 receipt ID，不登记绝对路径。哈希/字节数与原始文件不一致即证据无效。

## 7. 不可变裁决

1. Bearer scheme ASCII 大小写不敏感；scheme 与 credential 间 RFC 6750 `1*SP` 的双/多 SP 是正向断言。当前候选预期失败并必须 NO-GO，QA 不得改产品或原回归求绿。
2. Schema3 每个非 `/healthz` 外部请求外层持久推进一次；admit、每 attempt、邀请内部推进另计。连接、清零、单请求同步、读取、清理边界缺一不可；排队不用 sleep。
3. 任一时钟写失败 fail-closed；高水位不回拨；重启和合法恢复连续；认证/时钟失败不得触达上游、扣费或扩大授权。
4. 任一 `Restore` 失败不得创建、部分替换或删除目标，且不得遗留 `.restore-*`。最终安装失败是必跑门禁；无 seam 时 QA 执行保持 Blocked。
5. 所有数据、端口、SQLite、archive、key 与上游均为任务临时夹具；不访问常驻服务、用户数据、真实备份、凭据或供应商。
6. 本计划须由原独立评审 Agent 复核决定 READY。计划交付 completed 仅表示修订可执行；在 READY 与工程设施前置均满足前，后续 QA 执行不得开始。
