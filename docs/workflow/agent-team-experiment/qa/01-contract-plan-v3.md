# 后续独立 QA 合同执行计划（v3，唯一有效版）

## 1. 效力、候选与裁决边界

本文是后续 QA 的**唯一执行计划**。`01-contract-plan.md`、`01-contract-plan-v2.md` 和提交 `426ee922082461eb3b907013b5d70a36c599ddb1` 中的 `02-readiness-review.md` 仅是历史输入；v2 提议的 `docs/workflow/agent-team-experiment/qa/03-execution-report.md` 自本文起明确废止，禁止创建或使用。Git 中唯一执行报告路径是 `docs/workflow/agent-team-experiment/qa/03-execution.md`。

本计划静态冻结于候选 `aade8d2f72d42ca66349d6127cfd78a3a012b1cf`。只读祖先核对确认：v2 的集成提交 `ef379a26fbe7c6159134a01a6fb0164003c33ef8` 与 Restore 设施集成提交 `aade8d2f72d42ca66349d6127cfd78a3a012b1cf` 均已进入当前实验分支。设施提供 `restoreWithRename(input, directory, keyFile, rename func(string,string) error)`，公开 `Restore` 仍传入真实 `os.Rename`。设施专项已证明 absent-target 与公开正常路径可运行，同时真实暴露 existing-target 在最终 rename sentinel 失败前被 `os.Remove` 删除。故 B05 已具备确定执行条件，但当前候选预期 **FAIL/NO-GO**；这不是 `Not Run`、N/A 或设施回归可替代的结论。

计划修订完成只表示本文可供独立 readiness 评审，不表示产品通过、本文 `READY` 或允许执行。只有独立评审另行明确批准后才能开始执行；执行 Agent 不得自批。若评审发现设施接口与本合同仍有不可执行冲突，应记录具体函数、调用边界、复现条件和所需独立处理，停止执行，不修改设施、产品或断言。

正式依据为 `docs/product-brief.md`、`docs/releases/v0.1.0.md`、`docs/workflow/requirements.md`、`docs/workflow/design.md` 与正式 product-definition；实现依据包括 `controlledAppID`、`Server.ServeHTTP`、`Store.advanceClock`、`Store.admit`、`Server.beginAttempt`、`Server.acceptInvitation`、`Backup`、`Restore/restoreWithRename` 及相关既有回归。RFC 9110 §11.1 与 RFC 5234 §2.3 要求 authentication scheme ASCII 大小写不敏感；RFC 6750 §2.1 的 `Bearer 1*SP b64token` 接受一个或多个 SP；SQLite TEMP trigger 仅对创建连接可见。Worker、设施及既有回归回执都是输入，不能代替本次独立 QA。

## 2. 冻结的写路径、环境与证据合同

### 2.1 唯一写路径

后续 QA 唯一允许新增或修改的测试文件是 `internal/relay/qa_independent_regression_test.go`；不得改 `bearer_auth_regression_test.go`、`clock_advance_regression_test.go`、`backup_manifest_regression_test.go`、`backup_restore_failure_seam_test.go`、产品代码或任何第二个测试文件/runner。Git 中唯一报告是 `docs/workflow/agent-team-experiment/qa/03-execution.md`。执行前后若 tracked path diff 出现除此二者外的执行变更，立即判范围失败并停止。

不得创建或提交任何 QA evidence 原始产物目录。完整 stdout/stderr、截图、SQLite、临时 archive、工作区副本、密钥夹具、Native/会话历史只能进入：(a) 经 `git check-ignore -q test-results/` 和具体 run root 双重确认的 `test-results/qa-contract/<candidate-full-sha>/<run-id>/`，或 (b) 控制器私有运行目录。当前仓库 `.gitignore` 含 `test-results/`，但执行时仍必须复核。禁止把绝对私有路径写进 Git。

`03-execution.md` 只保存：receipt ID/相对索引、候选与 source-manifest SHA-256、脱敏命令、UTC 起止时间、退出码、stdout/stderr 各自字节数与 SHA-256、关联原始产物的字节数与 SHA-256、同版来源、脱敏断言摘要。不得保存完整日志、数据库、archive、截图、credential、prompt/completion、工作区副本或会话历史。

### 2.2 同版可复验对账

执行在同一冻结 worktree 中依次完成，命令不得用吞退出码的管道包装：

1. **运行前冻结：**记录 `git rev-parse HEAD` 的 full SHA、`git status --porcelain=v1 --branch`、`git diff --check`；用 `git ls-files -s` 加每个 tracked 文件 SHA-256 形成源码 manifest，并单列 Go 源、`go.mod`、`go.sum`、Makefile、唯一 QA 文件和报告。记录 manifest 文件 bytes/SHA-256。任何未声明改动或候选不是获评审 SHA 均停止。
2. **工具链：**分别记录 `go version`、`go env GOTOOLCHAIN GOOS GOARCH CGO_ENABLED`、`node --version`、`npm --version`、`git --version`、`make --version` 的脱敏 stdout/stderr、退出码、时间、bytes/SHA-256；工具缺失或与项目锁定要求不符如实 Fail/Blocked，不替换固定入口。
3. **逐命令 receipt：**每次命令先写 UTC start，结束后立即写 UTC end 和真实 exit code；stdout 与 stderr 分文件保存。对每个文件计算 bytes 和 SHA-256，receipt ID 固定为 `qa-contract/<sha>/<run-id>/<matrix-id>/<attempt>`。重跑不覆盖旧 receipt；报告说明为何重跑。
4. **来源绑定：**每条 receipt 引用候选 full SHA 与运行前 source-manifest SHA-256。专项依次覆盖 A、C、B、X；无论专项成败，仍原样执行一次 `make verify` 并保留真实退出码。专项绿不能替代全量，专项失败和产品缺陷不能从报告删除。
5. **运行后冻结：**重新采集 full SHA、status、diff-check 与同算法 source manifest。排除预先声明的唯一 QA 文件、唯一报告及 ignored receipts 后，运行前后源码必须逐字节一致；否则本轮证据作废。QA 文件若在执行期间完成，先冻结其最终摘要，再开始正式矩阵，正式运行期间不得再改。
6. **复验：**复核者凭 receipt ID 找到私有原件，重算 bytes/SHA-256，对照报告的命令、起止时间和退出码；再重算 source manifest 并确认同一候选。任一原件缺失、哈希/字节数不符、来源不一致或时间边界不闭合，相关结论无效。

所有夹具仅用 `t.TempDir()`、临时 SQLite、临时端口及进程内或临时受控 HTTP/SSE 上游。不得访问真实供应商、常驻服务、真实备份、用户数据或凭据；不得把测试上游称作供应商集成。

## 3. 通用计划项格式与执行顺序

下列每项均冻结七个字段：**前置数据、实际入口、操作顺序、可观察断言、失败判定、隔离边界、receipt**。未写在单项中的共同操作是：先完成夹具、再设观测零点、执行唯一被测动作、读取观测、清理；任何 setup/cleanup 失败也使该项失败。顺序为 A01–A05、C01–C07、B01–B05、X01–X03；不得因预期失败删项。报告逐项只能写 Pass、Fail 或“执行前真实设施冲突导致 Blocked”；B05 禁止 Not Run/N/A。

## 4. Bearer 正常、边界、隔离与并发

| ID | 前置数据 | 实际入口与操作顺序 | 可观察断言与失败判定 | 隔离边界与 receipt |
|---|---|---|---|---|
| A01 正常 | Schema 3 临时安装；一把有效 `mr_` key，仅 grant 一个模型；受控上游命中计数为零 | 分别经直接 handler 与 actual HTTP 调 `GET /v1/models`、普通生成、SSE；使用规范 `Bearer <key>` | models 只列授权模型；普通/SSE 成功且协议、request/attempt/usage 记录一致。任一越权、错误状态、记录不闭合或非预期上游次数即 Fail | 每子例新 DB/端口/upstream；不打印 key，只记摘要；`A01/<case>` |
| A02 scheme 大小写 | 同 A01，每个变体独立零点 | 对 `Bearer`、`bearer`、`BEARER` 和至少两个混合大小写 scheme，分别调用 models 与生成 | 所有变体与基准状态、授权集合及上游行为相同。ASCII 大小写敏感即规范 Fail/NO-GO；当前精确 `CutPrefix("Bearer ")` 实现预期失败，绝不改产品/旧回归求绿 | 每变体新请求与计数；`A02/<scheme>` |
| A03 `1*SP` 与非法结构 | 有效 key；上游零点 | 正向依次发送单 SP、双 SP、多 SP；负向发送缺头、裸 key、Basic、空 credential、scheme 前导空白、HTAB 分隔、credential 尾随空白/附加内容 | 双 SP、多 SP 必须与单 SP 一样成功；RFC 6750 `1*SP` 不是“精确一个 SP”。负向均拒绝且上游0、无扣费/attempt。当前实现预期正向失败即 NO-GO | 逐格式独立 case；header/credential 只存类型和摘要；`A03/<form>` |
| A04 生命周期/授权 | 有效、随机、停用、撤销、过期、轮换前后 key；两租户与不相交 grant | models、普通和 SSE 各走全组；在请求前用直接 SQL/产品夹具固定状态 | 仅当时有效 key 成功；租户/grant 不扩大；拒绝请求无分派、attempt、usage 或扣费。任何生命周期穿透或信息泄露 Fail | 不复用真实 key；每状态新安装或恢复确定快照；`A04/<state>` |
| A05 并发隔离 | 两租户，多把有效/无效 key；barrier 同时释放请求；受控 upstream 可阻塞并计数 | 以 channel/barrier 建立全部 goroutine ready→统一 release→全部 done；混合 models、普通、SSE、撤销/轮换边界；运行在 race-enabled 专项 | 每个响应只对应自身身份；无跨租户模型/记录/usage；无效请求上游0；有效并发不被错误身份污染；race 报告或 barrier 超时/泄漏均 Fail | 禁止 sleep 猜并发；临时 DB/端口；`A05/concurrent` |

## 5. 时钟：真实单连接 TEMP trigger 协议

### 5.1 强制观测边界

每个 C 子例使用新 Schema 3 安装和确定性 `store.clock`。夹具创建管理员、租户、key、模型、候选、邀请等全部在观测前完成，并等待所有辅助 goroutine、事务和 `Rows` 关闭。硬断言真实 `fixture.store.DB` 即产品 `Store.DB`，且 `DB.Stats().MaxOpenConnections == 1`（`database()` 当前调用 `SetMaxOpenConns(1)`）；不得另开 `sql.DB`、改变池配置或在观测期占用连接。

通过这个真实 `Store.DB.Exec` 创建 TEMP 审计表和 `AFTER UPDATE OF effective ON console_clock` TEMP trigger，记录自增序号、`OLD.effective`、`NEW.effective`。因单连接池，安装 trigger 与被测请求的持久 UPDATE 必须落在同一连接。安装后 `DELETE` 清零并确认 count=0；随后只允许**一个**被测请求，禁止登录、健康探测、fixture API、详情查询或其他辅助 HTTP。请求完成 channel 关闭后，才以同一 `Store.DB` 直接读取 TEMP 审计、`request_attempts` 和必要业务记录；直接 SQL 读取不经过 handler。最后 DROP trigger/table，并断言 `sqlite_temp_master` 中相关对象为零。安装、清零、唯一请求、读取、清理任一步失败均 Fail。

每行须满足 `NEW.effective >= OLD.effective`。来源不用伪造标签，而以隔离场景差分：外层-only=1；外层+admit/no-attempt=2；N attempts=2+N；邀请=外层1+邀请内部1。attempt 数由该唯一请求的脱敏 request ID 查询 `request_attempts`，并与 upstream 命中和 ordinal 对账。夹具及辅助请求不得位于零点后。

| ID | 前置数据 | 实际入口与操作顺序 | 可观察断言与失败判定 | 隔离边界与 receipt |
|---|---|---|---|---|
| C01 外层推进 | 合法 key；另建管理身份；按 5.1 安装并清零 | 独立子例调用 models、identity GET、console GET；另以 `/healthz` 子例 | 三类非 health 请求各 audit=1；health=0；HTTP/业务结果正确。多/少一行、连接约束丢失或额外请求均 Fail | 每入口新安装，不共享 trigger；`C01/<entry>` |
| C02 内部推进差分 | (a) 确定经过 admit 但无候选；(b) 单候选成功 N=1；(c) 受控失败链 N=3；(d) 未过期邀请 | 各自执行 5.1；唯一入口分别为生成或 `POST /admin/api/invitations/accept`；响应后直接 SQL 对账 | (a) audit=2、attempt=0；(b/c) audit=2+N、attempt ordinal 连续且 N=upstream 命中；(d) audit=2 且邀请消费结果正确。业务记录不能证明路径或出现不明 UPDATE 即 Fail | 各差分独立 DB/trigger；`C02/<no-attempt|N1|N3|invite>` |
| C03 排队与截止 | deadline 恰可跨越的 key/邀请；真实 server mutex；确定性时钟；观测已清零 | 主 goroutine 持锁；请求 goroutine 发出 `started`；测试 wallTime 在首次排队 sample 时关闭 `observed`；主 goroutine收到后推进时间并 release；等待 `finished` 后读取。channel/barrier 建立 happens-before | 不使用 sleep。外层持久 audit=1，邀请再加内部1；非持久 sample 单列；资源按新截止失效且 upstream=0。任一事件顺序不成立、超时或用 wall sample 冒充 UPDATE 即 Fail | 独立 goroutine 必须 join；`C03/<key|invite>` |
| C04 防回拨/高水位 | 先推进至 T1并完成观测清理，再把 wall clock 回拨 T0<T1，重新安装/清零 | 唯一请求后读 audit、内存 now、DB effective 与资源状态 | OLD/NEW、内存及 DB 均不低于 T1；已过期资源不复活；推进次数符合场景。回退或复活 Fail | 时间仅测试函数控制；`C04/rollback` |
| C05 写失败 fail-closed | 分别建立 begin、read、update、commit 的确定性 DB 故障入口/既有夹具；业务/upstream 零点 | 每种故障独立唯一请求，等完成后查可得的业务记录；不得用“关闭 DB 的 begin 失败”替代其余阶段 | 503 `storage_unavailable`、degraded、handler 不继续、upstream0、无 attempt/usage/部分业务写；错误阶段必须匹配。无法命中特定阶段是测试设施/夹具 Fail，不得假定通过 | 每故障新 DB；`C05/<phase>` |
| C06 重启/备份恢复连续 | Schema3 已达 T1；项目 `Backup` 生成临时合法 archive | 关闭重开后回拨并请求；另将 archive Restore 到临时新目标、Open、回拨并请求；按 5.1 观测可用阶段 | 重启及合法恢复后的 effective 均不低于 T1，过期资源不复活，业务正常；archive/DB 只在 temp | 不使用真实备份；`C06/<restart|restore>` |
| C07 旧 schema | 项目 API 生成 Schema1、Schema2 最小安装 | 各自运行合法可用入口并检查 schema/业务，不强装 Schema3 trigger | 旧 schema 行为按正式兼容合同工作；明确不要求 `console_clock` UPDATE。崩溃、错误迁移或意外创建时钟表 Fail | 每 schema 新目录；`C07/<schema>` |

## 6. 备份、manifest、恢复与最终 rename 失败

| ID | 前置数据 | 实际入口与操作顺序 | 可观察断言与失败判定 | 隔离边界与 receipt |
|---|---|---|---|---|
| B01 正常往返 | 项目 API 分别生成 Schema1/2/3 最小合法安装，Schema3 effective=T1 | `Backup`→只读解析 tar/manifest→`Restore` 到 absent temp target→`Open` | 恰有普通 `manifest.json`、`relay.db`；规范 digest/version/schema 一致；数据与 T1 守恒且可打开。任何手工伪造基线、真实备份或状态丢失 Fail | 每 schema 新 parent/archive；`B01/<schema>` |
| B02 manifest 边界 | 仅从 B01 测试 archive 改 manifest | 表驱动覆盖重复 version/SHA、未知键、尾随值、错误类型、缺字段、非对象、空/短/大写/非 hex digest；另验合法 whitespace/允许字段顺序 | 非法均在安装前拒绝、错误不泄密；合法变体成功。误收/误拒或残留 Fail | 每 case 新 archive/target；`B02/<case>` |
| B03 tar 边界 | B01 archive 的测试副本 | 覆盖重复/未知/缺 entry、非普通类型、路径名、负/超限 size、截断、校验、version/schema 不符及 entry 顺序 | 所有非法 archive 在目标安装前拒绝，错误稳定且无内容泄露；允许顺序成功。任一半安装 Fail | 不访问真实 archive；`B03/<case>` |
| B04 普通失败原子性 | B02/B03 每个负例各准备 absent target 与 existing-empty target；对 existing 记录 `FileInfo`/`SameFile` 身份、mode、排序清单、内容摘要与父目录清单 | 调公开 `Restore`，失败后立即重采快照并查 `.restore-*` | absent 仍 absent；existing 仍存在且身份/权限/内容/清单不变；无 `relay.db`、部分替换、删除或临时残留。任一差异 Fail | 仅 `t.TempDir` 父目录；`B04/<case>/<target-state>` |
| B05 最终 rename 失败 | 项目 `Backup` 生成合法 archive；使用已合入显式参数 `restoreWithRename`；absent 与 existing-empty 两状态，各重复10次。existing 记录与 B04 相同快照 | 注入唯一 sentinel 的 rename 函数；闭包断言仅命中已验证的 `.restore-*` temporary→精确 target，二者同一 temp parent，调用恰一次，然后返回 sentinel；调用后重采目标/父目录 | `errors.Is(err,sentinel)`；existing 必须仍存在且身份/权限/内容/清单不变，无 `relay.db`、替换或半恢复；absent 仍 absent；两者 `.restore-*`=0。注入未命中/多命中、错误丢失或任一次状态变化即 Fail。当前先 `os.Remove` 再注入 rename 的实现预计令 existing 失败并总体 NO-GO；设施专项结果不得替代此项 | 禁止 chmod race/goroutine race/root/mount/FUSE/全局 hook；仅 temp parent；`B05/<state>/<iteration>` |

B05 不允许 Not Run、N/A、弱化为“临时目录已清理”，也不允许把 `backup_restore_failure_seam_test.go` 的结果当作独立 QA。若合入的函数签名或调用位置不再能只替换最终 rename，应在执行前以源码位置和编译证据报 Blocked，请求独立设施处理；不得自行改 `backup.go`。

## 7. 跨改动场景与固定全量入口

| ID | 前置数据 | 实际入口与操作顺序 | 可观察断言与失败判定 | 隔离边界与 receipt |
|---|---|---|---|---|
| X01 恢复×时钟×认证 | B01 恢复后的 Schema3，key/grant、effective=T1；wall clock 回拨；受控 upstream | 用 scheme 大小写变体及双/多 SP 调 models、普通生成和 SSE，再直接查 clock/attempt/usage | 高水位连续；RFC 合法认证均等价；grant 不扩大；记录闭合。任一认证格式误拒、时钟回拨或越权即 Fail/NO-GO | 新恢复目标/端口；`X01/<form>/<entry>` |
| X02 时钟失败×认证 | Schema3，构造 C05 各写失败；合法及非法 Authorization 全组 | 各格式独立请求，先确认故障命中，恢复 DB 后再跑 A02/A03 基准 | 时钟失败优先 503、upstream0、无扣费/attempt；恢复后合法 RFC 形式成功、非法仍拒绝。认证不能绕过 fail-closed | 每故障新 DB；`X02/<phase>/<form>` |
| X03 固定全量 | 最终 QA 文件已冻结；运行前 source manifest 与工具链 receipt 完整 | 在同一候选、同一工作树**原样**执行 `make verify` 一次，保存完整 stdout/stderr 与真实退出码；随后做运行后源码对账 | 只有 exit 0 且 source unchanged 才是全量 Pass。任何专项失败、产品缺陷、门禁失败或源码变化均如实保留；旧 Worker 绿色、设施绿色或专项绿色不能替代。即使 A02/A03/B05 已失败，也仍执行并报告 X03 | 原始输出仅 ignored/private；`X03/make-verify` |

## 8. 报告结构与最终裁决

`03-execution.md` 至少包含：候选与评审授权、运行前后 source-manifest 对账、工具链表、A/C/B/X 每项七字段对应的实际结果、receipt 索引表、缺陷与复现摘要、全量入口结果、禁入物检查及最终裁决。每个索引行必须能从私有原件重算字节数和 SHA-256；不得粘贴完整 stdout/stderr。

最终规则不可变：

1. Bearer scheme 按 ASCII 大小写不敏感，且 `1*SP` 的双 SP、多 SP 都是正向断言；当前不符合只能 Fail/NO-GO。
2. Schema3 每个非 `/healthz` 外部请求外层持久推进一次，admit、每 attempt、邀请内部推进另计；TEMP trigger 必须绑定真实单连接 `Store.DB`，夹具和辅助请求不得污染。
3. 排队与截止必须用 channel/barrier 建立确定 happens-before，禁止 sleep 猜测。
4. 任一 Restore 失败不得创建、删除、替换或半恢复目标；最终 rename 失败必须使用显式参数化设施真实执行，current existing-target 丢失即产品 Fail/NO-GO。
5. 原样 `make verify` 是固定全量入口；分项绿不替代全量，失败也不得隐去。
6. 所有运行数据、端口、SQLite、archive、credential 和 upstream 都是临时测试夹具；无真实供应商、真实备份、常驻服务或凭据。
7. 本文的 `completed` 只代表计划修订可交独立评审，不是产品通过、计划 `READY` 或执行许可。
