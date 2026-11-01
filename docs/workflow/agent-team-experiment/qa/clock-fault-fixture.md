# 时钟持久化确定性故障夹具

## 基线与选择

本设施依据提交 `8f81c1faae3b91a46c8621f133c81fa5eff9f428` 的 `02-readiness-v3.md`、真实 `Store.advanceClock`、`InitConsole`/`Open` 初始化和既有时钟回归建立。它只位于 `internal/relay/clock_persistence_fault_fixture_test.go`，不改变产品代码、API、schema、依赖或全局 hook。

夹具使用 `sql.OpenDB` 和测试内 `driver.Connector` 包装现有 `github.com/mattn/go-sqlite3` 驱动。未使用 `sql.Register`，因此没有不可注销的全局 driver 名称或共享故障状态。每个子例由 `t.TempDir`、新的 schema 3 安装、独立连接池、连接级状态和清理函数隔离；连接池保持产品的 `MaxOpenConns(1)` 和相同 SQLite DSN 参数。无 sleep、并发竞态、权限故障或常驻服务。

## 调用与计数

故障参数是包内枚举 `clockFaultStage`：`clockFaultNone`、`clockFaultBegin`、`clockFaultRead`、`clockFaultUpdate`、`clockFaultCommit`。调用 `newClockFaultFixture(t, stage)` 后直接执行真实 `fixture.store.advanceClock()`。

`clockFaultCounts` 的 `begin/read/update/commit` 分别计数：事务 Begin、匹配 `SELECT effective FROM console_clock`、匹配 `UPDATE console_clock SET effective=...`、`driver.Tx.Commit`。预期向量依次为：正常 `[1,1,1,1]`，begin `[1,0,0,0]`，read `[1,1,0,0]`，update `[1,1,1,0]`，commit `[1,1,1,1]`。另有 `commitHookHits` 单独证明底层 SQLite commit hook 恰命中一次，且所有非 commit 模式为零。前三种失败用可 `errors.Is` 核对的 `errClockFault` sentinel；包装文本携带阶段值，使失败可区分。

## 正常委托与 commit-only 证明

正常模式把 Begin、Query、Exec 和 Commit 全部委托给真实 SQLite，并额外读取 `console_clock` 机械确认指定 RFC3339Nano 时间已持久化。

commit 模式不在 Begin、读取或 UPDATE 返回错误。连接创建时仅对该物理 SQLite 连接安装 mattn 驱动的 `RegisterCommitHook`；hook 返回非零，使底层 SQLite 在真正执行 COMMIT 时返回扩展码 `SQLITE_CONSTRAINT_COMMITHOOK`。测试同时要求阶段计数为 `[1,1,1,1]`、`commitHookHits == 1`，并以 `errors.As` 核对该扩展码，因此证明前三阶段已成功、错误首次出现于真实提交，而不是 begin/update 替代故障。该模式不把 commit 错误改写成 sentinel，也不吞掉产品看到的真实错误。

## 隔离、清理与边界

初始化先走真实 `InitConsole` 和 `Open`，随后仅把该测试 Store 的 `DB` 换为指向同一临时 `relay.db` 的包装连接池；Store 的目录锁保持到清理。`t.Cleanup` 关闭连接池和锁，临时目录由 testing 清理。故障状态不是包级变量；唯一包级 sentinel 是不可变错误身份。该设施仅验证四阶段入口本身，不是独立产品 QA，也不裁决 Bearer 或 Restore 结果。

## 设施验证

- `gofmt -w internal/relay/clock_persistence_fault_fixture_test.go && go test ./internal/relay -run '^TestClockPersistenceFaultFixture$' -count=1`：退出码 1；执行环境的默认 Go build cache 为只读，测试未进入编译。
- `env GOCACHE="$GOCACHE" GOTMPDIR="$GOTMPDIR" go test ./internal/relay -run '^TestClockPersistenceFaultFixture$' -count=1`（`GOCACHE`、`GOTMPDIR` 指向隔离的可写临时目录，首次）：退出码 1；正常委托持久化辅助查询被计入 read，实际 `[1,2,1,1]`，据此把阶段计数快照移到辅助查询之前。
- 同一定向命令（修正后）：退出码 0，`ok model-relay/internal/relay 0.983s`。
- `env GOCACHE="$GOCACHE" GOTMPDIR="$GOTMPDIR" go test ./internal/relay -count=1`（变量同上）：退出码 1；最小受影响包运行 66.362s，既有 `TestRestoreFinalRenameFailureSeam/existing-target` 失败（failed rename 后 target 不存在）。这与冻结评审记录的既知 Restore 产品缺陷一致，不由本设施吞错或修改。
- 增加底层 hook 独立命中断言后的最终定向命令，并串行执行 `git diff --check`、`git status --short`、`git diff --stat`：退出码 0；测试输出 `ok model-relay/internal/relay 0.979s`，差异检查无错误，状态仅有两个获授权交付路径。

以上均为设施定向验证或最小受影响包检查；没有运行 qa-final 产品矩阵、`make verify` 或旧流水线。

## 限制

SQL 阶段识别刻意绑定 `advanceClock` 当前的两条具体语句前缀；产品 SQL 若改变，计数断言会失败并要求同步复核设施。commit 错误身份采用 SQLite 稳定扩展码而非 Go sentinel，因为错误由真实 SQLite commit hook 产生。
