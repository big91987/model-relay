# Restore 最终 rename 可测性记录

## 基线候选与输入

- 工程基线：`ef379a26fbe7c6159134a01a6fb0164003c33ef8`。
- 合同输入：`docs/workflow/agent-team-experiment/qa/01-contract-plan-v2.md` 的 D-RESTORE-02、第 4 节与 B05；实现输入为该基线的 `internal/relay/backup.go`。
- 基线实现的最终安装顺序是：验证临时安装、删除临时 `.lock`、对 existing-empty target 执行 `os.Remove(directory)`，随后执行 `os.Rename(temporary, directory)`。因此最终 rename 失败时 existing target 已被删除；本任务只建立可测入口，不修复该算法。

## 设计边界

`Restore(input, directory, keyFile)` 的公开签名与归档/真实文件系统流程保持不变。它只把真实 `os.Rename` 作为显式函数参数传给未导出的 `restoreWithRename`。该参数只替代函数末尾的 temporary→target 最终安装调用；归档读取、校验、数据库打开验证、目标检查/删除和临时目录清理由原实现执行。没有包级 hook、可变全局、环境变量、公开调试接口或虚拟文件系统。

独立回归 `backup_restore_failure_seam_test.go` 使用项目 `Init`/`Backup` 生成合法临时备份，以唯一 sentinel error 注入最终 rename。闭包硬断言源路径中已经存在验证后的 `relay.db`；并记录一次调用的 source/target，断言二者同处该用例的 `t.TempDir()` 父目录、source 具有 `.restore-` 前缀且 target 精确匹配。

existing 状态遵循 v2 的 **existing-empty target** 前置，因为公开合同会在目标含任何 entry 时于注入点之前返回 `restore directory must be empty`。测试在调用前记录目录身份（`os.SameFile` 可比的 `FileInfo`）、mode、有序清单及清单 sentinel SHA-256；失败后要求全部不变、不得出现 `relay.db`。absent 状态要求失败后仍不存在。两种状态都要求 sentinel 可由 `errors.Is` 识别且父目录没有 `.restore-*` 残留。

## 正常路径代码依据

`backup.go` 中公开入口为：

```go
func Restore(input, directory, keyFile string) error {
    return restoreWithRename(input, directory, keyFile, os.Rename)
}
```

因此生产入口仍显式使用标准库真实 `os.Rename`，注入 seam 不能替代公开路径。`TestRestorePublicPathStillInstallsValidBackup` 另从合法项目备份调用公开 `Restore`，断言真实目标出现 `relay.db`、可由 `Open` 打开且没有临时残留。

## 针对性验证与结果

所有命令均在本任务 worktree 执行；Go 构建缓存显式设为 `/tmp/restore-failure-go-cache`，没有访问真实备份、数据、密钥或供应商。

1. 初次 `gofmt -w internal/relay/backup.go internal/relay/backup_restore_failure_seam_test.go` 与专项 `go test` 组合命令：退出码 1。先暴露测试文件缺少右花括号；同次测试还因默认用户 Go cache 只读而 setup failed。花括号已修正，后续命令改用 `/tmp` cache。这不是产品结论。
2. `GOCACHE=/tmp/restore-failure-go-cache go test ./internal/relay -run 'TestRestore(FinalRenameFailureSeam|PublicPathStillInstallsValidBackup)$' -count=1`：退出码 1。absent 与公开正常入口未报错；`TestRestoreFinalRenameFailureSeam/existing-target` 硬失败：`existing target was removed after failed rename: ... no such file or directory`。这确定性暴露基线的 remove→rename 失败窗口，断言按合同保留。
3. `GOCACHE=/tmp/restore-failure-go-cache go test ./internal/relay -run '^TestRestoreFinalRenameFailureSeam/absent-target$' -count=1`：退出码 0，`ok model-relay/internal/relay`。
4. `GOCACHE=/tmp/restore-failure-go-cache go test ./internal/relay -run '^TestRestorePublicPathStillInstallsValidBackup$' -count=1`：退出码 0，`ok model-relay/internal/relay`。

上述是开发专项回归，不是独立 QA 通过。按任务边界未运行、未替代原样 `make verify`；它明确留给后续独立 QA。由于硬合同已真实失败，本任务没有运行更宽的包回归来掩盖或归一化该结果。

## 限制与结论

设施本身达到最小参数化边界；absent-target 清理与公开正常恢复已有实测通过。existing-target 的保全合同在当前算法上实测失败，根因是注入点前已经执行 `os.Remove(directory)`，不是 sentinel 丢失或 seam 未命中。修复恢复算法不在本任务范围内，不能据此宣称 Restore 原子性或独立 QA 已通过。
