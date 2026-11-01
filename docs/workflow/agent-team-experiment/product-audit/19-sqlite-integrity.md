# SQLite 启动、迁移与完整性静态审查

## 结论与范围

本次只读审查当前候选源码中的 `store.go` 初始化/Open/进程锁/WAL/schema、`upgrade.go` 迁移与校验，以及 identity、console、routing 扩展 schema 和直接合同测试。没有打开任何 SQLite 文件，没有运行初始化、迁移、测试或服务，也不据此给出产品通过结论。backup manifest 严格解析专案明确排除；高层备份、监控、灾备等运维成熟度不重复审查。

结论：未发现可由静态证据直接定为 P0 的问题；发现 3 项 P1、2 项 P2。迁移 DDL/DML 置于单一事务且提交前做完整性检查，是可靠的主要防护；但新装初始化的文件系统动作不具原子性，启动恢复跨越非事务写与另一事务，schema 合同校验也不足以识别约束/索引/触发器漂移。

## 事实与已有防护

- 单实例与连接模型：数据目录 `.lock` 使用非阻塞排他 `flock`，失败即关闭句柄并拒绝打开（`internal/relay/store.go:48-57`）；数据库池限制为一个连接（`:59-74`）。这既将 SQLite 写并发串行化，也保证 DSN 中 `_foreign_keys=on`、`_journal_mode=WAL`、`_synchronous=FULL`、2 秒 busy timeout 作用于实际使用连接（`:64-70`）。这是同一主机/同一文件系统、且所有入口遵守该锁时的防护，不等于网络文件系统上的分布式锁保证。
- 新装 schema、meta/admin 数据、console/identity 和 routing 扩展均在一个数据库事务中创建并提交（`store.go:152-183`）；失败路径 defer rollback。identity 的角色/状态/布尔约束及成员外键见 `identity.go:15-22`，routing 扩展的版本行、策略 CHECK 与复合外键见 `routing_latency_storage.go:9-13`。
- Open 先验证版本只有 1/2/3，默认拒绝 schema 1，核验 sentinel/key、加密上游材料、schema 及完整性，任一失败均 Close 并释放锁（`store.go:211-238`）。外键与物理完整性检查使用 `PRAGMA foreign_key_check` 和 `PRAGMA integrity_check`（`upgrade.go:49-62`）。
- schema 1→2/3 的重命名、复制、对象计数守恒、证据表创建、版本更新、材料复验、routing 扩展和最终完整性检查都位于同一事务，只有最后提交才生效（`upgrade.go:133-219`）。schema 2→3 同样把 console 创建、版本更新、合同/完整性检查置于单事务（`:138-154`）。迁移前必须成功创建备份；失败明确不开始迁移（`:110-137`）。
- 扩展的半成品会被拒绝：routing 四张表必须全无或全有，版本表必须恰有一行且版本为 1（`routing_latency_storage.go:15-44`）。普通 Open 可在独立事务中为 schema 2/3 原子补建完全缺失的扩展（`:47-59`）。
- console migration 把 identity、console 表、遗留映射及有效时钟初始化纳入调用方事务（`console_storage.go:46-54`）；恢复结算本身也在单事务中完成，提交后才发布内存有效时间（`:94-142`）。邀请接受把用户、membership、邀请状态和审计放在同一事务（`identity.go:439-499`）。routing 的决策、延迟证据和 attempt 创建一起提交（`routing.go:286-319`）。
- 直接合同测试文本覆盖：schema 1 显式升级、对象/密钥材料守恒和重复升级（`upgrade_test.go:46-113`）；错误 key、已有输出、锁冲突、未知版本不迁移（`:116-164`）；非法旧值导致回滚及缺表拒绝 Open（`:166-197`）。另有 console migration 合同入口 `console_migration_test.go:13` 和 routing 扩展损坏入口 `routing_latency_test.go:450`。这些是现存测试代码的静态事实，本次未执行。

## 发现与缺陷假设

### P1-1：新装初始化可留下不可重试、或凭据结果不可交付的半成品

事实：`initSchema` 先创建并 chmod 数据目录、取得会生成 `.lock` 的锁，再以 `O_EXCL` 创建并写入 key 文件，之后才打开数据库和开始 schema 事务（`store.go:114-156`）。任一后续建表、加密、console/routing 初始化或 commit 失败，只回滚 SQLite 事务；没有删除 key、数据库、WAL/SHM 或目录。更严重的是 commit 成功后才 chmod `relay.db`，chmod 失败会返回错误且不返回已生成管理员密码，但数据库已提交（`:157-187`）。重试又会被“目录必须为空”或“key file must be new”拒绝（`:121-139`）。

缺陷假设与影响：磁盘满、权限/文件系统错误或 commit 后 chmod 失败，可形成调用者看到失败、现场却有部分或完整安装的状态；管理员密码可能永久丢失，人工清理若判断错误还可能误删有效安装。影响为安装可恢复性和凭据可用性，定为 P1。

建议防护：在同一父目录建立私有 staging 目录和 staging key，完成数据库、校验、权限设置、fsync 后再以明确的发布顺序原子 rename；失败清理仅限本次已确认创建的 staging 项。若 key 必须位于任意外部路径，应定义跨文件系统无法原子发布时的恢复清单。至少把数据库 chmod 移到 commit 前可验证阶段，并新增每个注入失败点的“无已发布安装/可安全重试”合同。

### P1-2：启动恢复不是一个原子启动事务

事实：校验及 routing 补建完成后，普通 Open 用一次无显式事务的多语句 `db.Exec` 改写 `requests`、`request_attempts`、`upstream_checks`（`store.go:235-245`），随后 schema 3 才进入另一事务执行 admission 结算、审计与 clock 恢复（`:246-252`；`console_storage.go:94-142`）。若首段语句间、首段之后或第二段恢复中失败，Open 返回错误，但先前已经提交的状态可能保留。源码没有把三张证据表的恢复与 console settlement/audit 绑定为一个事务。

缺陷假设与影响：启动期 I/O/约束错误可能留下 requests 已标记 interrupted、attempt/check 尚未完全改写，或基础证据已改写但 admission 尚未结算的中间态。再次启动可能收敛，但审计/计费证据在失败窗口内不是单次原子转换，定为 P1。

建议防护：把所有启动恢复 SQL、`settleAdmission`、恢复审计和 clock 读取放进一个事务，提交后再发布内存 clock；对每条更新检查错误，并用故障注入覆盖每一语句/commit 失败后的整体现状及第二次启动幂等性。

### P1-3：Open 的 schema 合同不能证明约束、外键、索引和触发器完整

事实：`validateStorage`/`validateConsole` 主要执行 `SELECT ... LIMIT 0` 来证明表和列可读（`upgrade.go:65-100`；`console_storage.go:56-91`）。`integrity_check`/`foreign_key_check` 能发现页面损坏和现存孤儿，但不会证明预期的 CHECK、UNIQUE、FK 定义、部分索引或触发器仍存在。routing 扩展额外检查表数量、版本及现存策略值，却也没有核对 `sqlite_master` 定义（`routing_latency_storage.go:15-44`）。例如 `current_generation` 部分唯一索引、`diagnostic_role_cleanup` 与 `register_dispatch` 触发器是关键直接合同（`console_storage.go:14,18,23`），Open 不验证其存在。

缺陷假设与影响：被旧二进制、手工操作或不完整外部恢复重建成“同名同列但少约束”的库会通过启动；之后可出现多个 current generation、角色变更后诊断授权残留、attempt 无 dispatch 记录等。完整性检查可能持续返回 ok，直到业务不变量已损坏，定为 P1。

建议防护：为每个 schema 版本维护规范化 schema 指纹，或逐项核验 `pragma_table_info`、`foreign_key_list`、`index_list/index_xinfo` 和关键 trigger SQL；同时执行少量只回滚的负向约束探针。迁移提交前和 Open 使用同一验证器，避免两套合同漂移。

### P2-1：核心表若干业务域仅靠应用层约束

`app_keys.enabled`、`requests.stream/attempts_complete` 及多种 state/error/evidence 字段没有数据库 CHECK；console 中 `revoked/retired/running/pending/hard` 等也多为无 CHECK 的 INTEGER（`store.go:77-104`；`console_storage.go:10-42`）。相对地 upstream/models/routing 和 identity 的部分布尔/枚举已有 CHECK。直接 SQL 写入、未来迁移或缺陷可产生应用未预期值，而启动合同通常只验证可查询。影响主要是损坏隔离和故障定位，定为 P2。建议按 schema 版本补齐布尔、枚举、非负值以及时间/状态组合约束；无法直接 ALTER 的表用事务化重建并做负向测试。

### P2-2：初始化目录与 key 路径只防止一种冲突，路径合同不完整

代码只比较清理后的字符串，拒绝 key 路径恰好等于 `<directory>/relay.db`（`store.go:114-117`）；没有使用绝对路径/符号链接解析，也未拒绝 key 位于待要求为空的数据目录内的其他名称。若 key 位于数据目录，先检查空目录后创建 key 并可继续，但后续备份/目录生命周期与“key must be separate”的安全意图可能分叉。影响取决于 CLI 上层是否另有路径限制，故定为 P2 假设。建议对两个路径求绝对路径并解析现有父目录 symlink，明确要求 key 不在数据目录树内；把该规则加入 Init/Open/Upgrade 一致合同。

## 优先级汇总

| 优先级 | 项目 | 主要影响 |
|---|---|---|
| P0 | 无静态确认项 | — |
| P1 | 初始化文件系统发布非原子 | 半成品、不可重试、管理员密码丢失 |
| P1 | 启动恢复跨非事务写和另一事务 | 失败启动留下跨表中间态 |
| P1 | schema 验证不核对约束/索引/触发器 | 漂移库可通过 Open，业务不变量静默失效 |
| P2 | 数据库业务域约束不一致 | 非法状态可由直接写/未来缺陷持久化 |
| P2 | key/data 路径分离校验不完整 | key 可落入数据目录或经别名绕过意图 |

## 未执行项与限制

- 按任务禁令未打开真实或临时 SQLite，未执行 Init/Open/Upgrade/Restore、PRAGMA、迁移、故障注入、Go 测试、`make verify` 或浏览器测试。
- 未验证 SQLite 驱动对单次多语句 `Exec` 的逐句提交细节；因此 P1-2 以“源码没有显式原子保证”的缺陷假设表述，而非声称已复现部分提交。
- 未验证 `flock` 在 NFS/容器卷/异常文件系统上的语义，也未做断电/WAL 恢复测试。
- 未核对 backup manifest 严格解析实现；只将迁移前备份成功作为迁移入口的现有防护记录。
- 引用的测试仅经静态阅读，不能作为本轮运行证据。报告不修改任何实现，也不包含数据库、凭据、运行日志或本机路径证据。
