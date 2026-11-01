# 07 运行诊断与数据恢复保障审查

## 审查口径与候选

本报告是候选代码、仓库文档和只读 Git 元数据的静态接线审查。**已验证**表示可由当前文件/行号或 Git 对象直接确认；**推断**表示由控制流或文件系统语义推导、尚未故障注入；**未执行**表示本任务明确没有启动服务、运行 CLI/业务页面、打开常驻数据目录、SQLite、备份、日志、密钥或凭据，也没有执行备份、恢复、升级、测试或 `make verify`。因此下文不把文档合同、测试源码或建议写成已实测保障，也不复核 strict-backup-manifest 的产品断言。

当前审查候选/基线为 `122c9abece28f562616483e991098d020d236d6d`。三项前置的任务提交与合入 SHA（只读 Git 已验证）为：strict-bearer-auth `67f464494f4d59be8a20ef99d9b15175ecd00e78` → `2d48ee08f12368e1b806fcc0975943e2eca0c8ee`；single-clock-advance `318639d819165f3309f123133fdaf6229003270c` → `532e713acc998162c682b310c8f280e1651e6f95`；strict-backup-manifest `7dc57094557b8883c130a41ed306d1223c0d92f8` → `122c9abece28f562616483e991098d020d236d6d`。

## 结论

未发现静态证据表明存在一个必然导致所有安装立即丢失的 P0。已有独占锁、WAL/FULL 同步、启动校验、事务迁移、迁移前备份、归档摘要、临时目录恢复、进程中断状态收敛和密钥分离等防护。但仍有 3 项 P1：管理面存储错误可能不进入健康/日志故障状态；恢复成功缺少目录持久化屏障；备份/升级失败后的诊断粒度不足。另有 3 项 P2，集中在初始化失败残留、健康覆盖范围及仓库内恢复手册可操作性。它们意味着可能出现“页面报错但探针仍绿”、恢复成功后断电却回退/丢失目录项、或运营者无法从单一错误定位阶段的情况。

## 发现

### P1-1 管理与报表存储故障没有统一接入 degraded 与运行日志

**已验证。** `storageFailure` 会设置进程内 sticky degraded 并仅记录固定字符串 `storage_write_failed`（`internal/relay/server.go:85-87`），健康接口只依据该标志或 `DB.PingContext` 返回 503（`internal/relay/server.go:149-168`）。生成链和时钟推进显式调用该入口（例如 `internal/relay/console_clock.go:13-19`），但控制台事务开启失败只向用户返回 `storage_unavailable`（`internal/relay/console_api.go:29-36`），报表查询失败也只返回相同 503（`internal/relay/console_reports.go:130-135`、`158-185`、`236-265`），没有置位 degraded 或记录请求/操作/阶段。用户可见错误本身采用稳定、去敏的通用 envelope（`internal/relay/server.go:77-84`），是已有防护。

**推断影响。** 当 SQLite 仍能 `Ping`、每请求时钟写仍成功，但某张表损坏、schema 漂移或某一查询失败时，管理页面会 503，而 `/healthz` 仍可能 200；运营日志也可能没有对应事件。这是不可观测的局部故障，可能延误处置，并让部署控制器把业务不可用实例视为健康。另一方面，不能把所有 4xx/查询错误都永久置 degraded，否则会造成不必要停服。

**建议归属：仓库基建 + 工程工具。** 建立按错误类别统一的存储故障上报函数：记录去敏的 operation/stage/error class/request ID，针对完整性/schema/I/O/只读/磁盘满等致命类别置 degraded；为健康增加“最近致命存储错误”原因码或独立 readiness（仍不泄露底层错误）。增加静态/单元检查，防止新 503 存储分支绕过该入口。

### P1-2 恢复安装的目录切换缺少可见的断电持久化保证

**已验证。** 恢复先写父目录下临时目录、验证后关闭数据库，再删除临时锁并以 `os.Rename` 安装（`internal/relay/backup.go:140-153`、`188-230`）；目标非空会拒绝，失败时 defer 清理临时目录（`141-153`、`222-230`），避免代码路径上的部分安装。备份输出文件在成功前会删除，且 archive 关闭后调用文件 `Sync`（`internal/relay/backup.go:107-138`）。但恢复写出的 `relay.db` 关闭前没有 `Sync`，临时目录与目标父目录也没有目录 `Sync`，目标为空目录时还会先删除再 rename（`internal/relay/backup.go:188-230`）。

**推断影响。** 在 `Restore` 返回成功附近发生掉电/内核崩溃时，文件内容或 rename 的目录项未必具有仓库代码层承诺的持久性；目标可能消失、恢复目录可能回退或数据库不完整。这是恢复路径上的数据丢失/不可启动风险。普通进程错误下的原子 rename 防护不等于断电耐久保证；本任务未在任何文件系统上实测。

**建议归属：仓库基建。** 明确支持的文件系统/平台耐久合同；恢复数据库关闭前 fsync 文件，rename 前后 fsync 临时目录及父目录，并用同父目录的可恢复交换/标记避免“删空目标—rename”窗口。增加受控崩溃点测试，但不要把普通往返测试当断电证明。

### P1-3 备份、升级和启动错误过度归一化，难以安全判断是否可重试

**已验证。** CLI 最终只打印 `Model Relay operation failed: <error>`（`cmd/model-relay/main.go:148-152`）。升级把打开安装阶段的 key、锁、数据库、schema/材料错误统一成 `upgrade installation/key/lock validation failed`，备份阶段统一成 `upgrade backup failed; no migration performed`（`internal/relay/upgrade.go:110-137`）。启动监听失败同样只返回 `listen failed; installation retained`（`cmd/model-relay/main.go:103-114`）。代码具有先拿锁、迁移前备份、单事务迁移、迁移内 integrity/foreign-key 检查的防护（`internal/relay/store.go:48-75`，`internal/relay/upgrade.go:49-99`、`114-137`、`138-219`），但没有结构化阶段、错误类别、备份路径/候选版本等运行诊断输出。

**推断影响。** 运营者无法仅凭正式入口区分“服务仍在运行”“错误 key”“数据库损坏”“磁盘满/权限”“输出已存在”等，容易重复危险操作、误选恢复材料或延长停机。通用错误避免泄密是优点，但本机 CLI 可以在不输出密钥/内容/绝对路径的前提下提供稳定原因码。

**建议归属：仓库流程 + 仓库基建。** 为 CLI stderr 增加去敏、稳定的 command/stage/code 和明确的“源是否变更/输出是否创建”字段；文档给出按原因码分流的停止、保全、重试、恢复步骤。日志不得包含提示词、补全、凭据或解密材料。

### P2-1 初始化失败可能留下不可重试的半成品材料

**已验证。** 初始化先创建空目录和独立 key 文件，再打开数据库、建 schema、提交，最后 chmod 数据库（`internal/relay/store.go:114-187`）。key 写入后或数据库事务/commit/chmod 失败时，没有删除新 key、数据库/WAL/SHM 或目录的回滚路径；而重试要求数据目录为空且 key 文件必须全新（`121-138`）。

**推断影响。** 首次初始化遇到磁盘/权限/进程故障可能留下“不完整但看似有材料”的安装，后续命令拒绝或需要人工判定删除哪些文件。它更偏可恢复性/停机风险；在原本为空的新安装上不构成既有业务数据丢失。

**建议归属：仓库基建。** 在同父目录 staging 完整安装并校验、同步后原子发布；或输出明确残留清单和安全清理命令。任何清理仅限本次新建且已验证的路径。

### P2-2 `/healthz` 是进程/连接活性探针，不是写入、完整性或恢复就绪证明

**已验证。** 健康仅检查 degraded 标志和 DB ping，并在健康时暴露 version/schema（`internal/relay/server.go:149-168`）。测试源码覆盖 GET/HEAD、degraded、关闭数据库和人工重新打开后的状态（`internal/relay/health_test.go:20-132`）；本任务没有执行这些测试。启动时才执行材料解密、schema、完整性/外键校验和中断恢复（`internal/relay/store.go:193-252`，`internal/relay/upgrade.go:65-100`）。运行中探针不验证可写、WAL checkpoint、剩余空间、表级查询、备份新鲜度或 key 可恢复性。

**推断影响。** 探针为绿不能证明下一次持久化成功或备份可恢复；把它作为唯一部署/运营保障会掩盖部分故障。另一方面，每次探针跑全库 integrity 或写事务会放大锁与 I/O 风险，不应直接塞进当前 endpoint。

**建议归属：通用平台 + 仓库基建。** 平台区分 liveness/readiness/周期性深度诊断；深检在受控频率和锁预算下验证写能力、空间/WAL、最近备份/恢复演练回执，并保留 `/healthz` 的轻量用途。

### P2-3 恢复文档入口分散，缺少仓库自包含的故障决策与演练回执合同

**已验证。** README 给出停止、独占备份、空目录恢复、独立 key、损坏时保留原安装及 `process_interrupted`/degraded 行为（`README.md:79-93`），也明确升级回退会丢弃升级后新增数据（`README.md:107-118`），这些是良好的风险披露。密码重置文档说明独占锁、事务更新和会话失效（`docs/workflow/admin-password-reset.md:3-7`），且代码确实使用 `Open` 的锁/key 验证和事务更新两处管理员哈希（`internal/relay/store.go:263-295`）。但是本地部署文档把安装、升级和回退边界委托给仓库外的平台维护源（`docs/local-deployment.md:5-9`）；`conflict-recovery.md` 是历史文档冲突处置，不是数据恢复手册（`docs/workflow/conflict-recovery.md:1-15`）。仓库内没有统一说明备份保留/轮换、空间预检、恢复到隔离目标后的验收清单、RPO/RTO、谁可访问 key 与备份、或演练回执格式。

**推断影响。** 外部平台源不可用或版本漂移时，值班人员难以从当前候选独立完成安全诊断；只备份数据库却遗失 key 会造成永久不可解密，误把升级前快照当合并恢复会造成升级后数据丢失。README 已明确 key 必需且不在备份中（`README.md:91-93`），但没有可审计的新鲜度/配对检查。

**建议归属：仓库流程 + 通用平台。** 在仓库建立版本化 runbook，引用而非复制平台实现，列出故障树、停服/保全顺序、备份与 key 的独立保管和配对标识、隔离恢复验收、回切、RPO/RTO、授权和去敏回执。周期演练必须记录候选 SHA、备份标识、schema、校验阶段和实际结果；未演练必须标 Not Run。

## 已有防护与接线总览

- **启动/停止：已验证。** `serve` 限制显式 loopback，打开并验证安装后才监听；SIGINT/SIGTERM 最多优雅排空 10 秒，失败则取消并关闭（`cmd/model-relay/main.go:84-143`）。标准输出只有一次启动地址，标准错误走统一 CLI 失败前缀；没有结构化启动/停止事件。
- **SQLite：已验证。** 单连接、2 秒 busy timeout、foreign keys、WAL、`synchronous=FULL`（`internal/relay/store.go:59-74`）；安装目录有非阻塞排他 flock（`48-57`）。这降低并发写/热拷贝风险，但不证明文件系统断电语义。
- **schema/迁移：已验证。** 新安装 schema 3，serve 拒绝非 3；open 检查唯一 meta/admin、列契约、材料解密、integrity 和 foreign keys（`internal/relay/store.go:101-104`、`189-252`，`internal/relay/upgrade.go:17-40`、`65-100`）。升级先在锁内备份、再单事务迁移并做对象数量守恒/完整性检查（`internal/relay/upgrade.go:110-219`）。README 所称“原子迁移”与事务接线一致，但本任务未执行迁移。
- **异常恢复：已验证。** 启动把未完成 request/attempt/check 标为中断或未知（`internal/relay/store.go:242-250`），并在事务内结算仍 running 的 admission、写审计、恢复有效时钟（`internal/relay/console_storage.go:94-142`）。这保留未知/待核对状态而非伪造成功；无法证明最后一笔磁盘写已落盘。
- **备份/恢复：已验证。** 备份要求锁、checkpoint truncate 成功后复制单一数据库，归档含版本与 SHA-256，输出必须新建（`internal/relay/backup.go:60-138`）。恢复限制条目类型/大小/名称，先写临时目录，再用 key 打开并校验 schema 后 rename（`140-230`）。独立 key 不进归档是设计上的秘密分离，也是恢复材料单点依赖。
- **秘密：已验证。** key 文件要求 32 字节，哨兵和上游材料用 AES-GCM/AAD 验证（`internal/relay/store.go:197-227`，`internal/relay/security.go:28-56`）；初始化要求 key 与数据库路径分离且以 0600 新建（`internal/relay/store.go:114-145`）。README 明确同机特权进程不在保护边界（`README.md:95-101`）。
- **用户可见诊断：已验证。** HTTP 错误返回稳定 code/request_id 而不暴露底层错误（`internal/relay/server.go:77-84`）；报表保留 `process_interrupted`、错误码、attempt、路由和历史核对证据（`internal/relay/console_reports.go:154-200`），CSV 对公式前缀做转义（`281-307`）。`adminResponse` 缓冲管理响应再一次 flush（`internal/relay/admin_response.go:8-41`），降低半响应可见性，但忽略最终写错误，不能作为客户端已收到的证明。

## 未执行与证据边界

未运行任何 Go 测试、`make verify`、CLI、HTTP 服务、浏览器或外部平台流程；未验证 WAL checkpoint 在真实负载下的行为、锁跨平台语义、磁盘满/只读/损坏/断电、1 GiB 边界、备份内容、错误 key、恢复 rename、崩溃窗口、真实部署回退或真实供应商。`health_test.go`、README、密码重置文档及前置任务回执只作为源码/历史陈述存在，不是本任务运行结果。尤其没有复核 strict-backup-manifest 的正确性或测试结论。
