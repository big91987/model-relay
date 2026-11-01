# 时钟设施与 v4 最终可执行性独立评审

## 1. 唯一业务结论

**`READY`**。

该结论只授权后续执行者严格按 `01-contract-plan-v4.md` 执行，不表示产品通过、GO，也不预判 Bearer 或 Restore 的结果。当前 Bearer 对 scheme/空格的已知不兼容和 Restore existing-target 在最终 rename 失败时被删除的已知产品 Fail，必须在正式执行中保留为真实 Fail/NO-GO；它们不是计划 readiness 阻断。

本评审仅做源码、文档和 Git 的静态核对，未运行产品测试、设施专项或 `make verify`，未启动服务，也未访问真实服务、数据、备份、凭据或供应商。

## 2. 开始条件、候选和指纹

- 原生任务回执确认 `clock-persistence-fault-fixture` 与 `qa-final-plan-v4` 均为 `merged`，公开 `integration_commit` 分别为 `2aeeb3af8db73a33f6fcb99371769b79f865e367` 与 `348e297f63cd4528fc05e5eaa1d7a0c2e3156d81`。评审开始时 HEAD 为后者，`git status --short` 无输出；两次 `git merge-base --is-ancestor <commit> HEAD` 均退出 0。
- `2aeeb3a..348e297` 只新增 `docs/workflow/agent-team-experiment/qa/01-contract-plan-v4.md`；产品/构建路径及 workflow 门禁无差异。获审产品/构建祖先 P 为 `2aeeb3af8db73a33f6fcb99371769b79f865e367`。
- P 的规范清单选取 `Makefile`、`go.mod`、`go.sum`、`cmd/`、`internal/`、`scripts/` 下所有非 `_test.go` tracked 文件，逐项记录 Git mode、blob、bytes、文件 SHA-256，按路径排序；共 36 项，规范清单 SHA-256 为 `ed7f68947ecc410cf80d8d8769b9547cfcb8d498959621a8f6490ced6edd8005`。P 与当前 HEAD 的相同路径集合摘要均为 `31aabf3d4a5c0756590fd7d28127230f275c96f0e40b429a03dc3b6f70a89f36`（mode/type/blob/path 清单），路径级 diff 退出 0。
- 输入内容指纹与 v4 声明一致：v4 `9c7510299982e7f15a2e89a69f7e1a90a8465dbb7b5a185a4db0ae65285d8fe4`；v3 readiness `2e88dbd67518fe7d77d13bd892f516de38885fbc49f26e3ad2c8e89602f371db`；时钟设施源码 `31f13f92fceaa1eec051c6f49cf598764c3e792188811d395096cbb8ebf4e9b2`；设施记录 `8e793dcebc0d17c42a716b121be7c854c96025487df8e786b93cc732eeb81257`。
- 唯一后续 QA 文件 `internal/relay/qa_independent_regression_test.go` 与唯一报告 `docs/workflow/agent-team-experiment/qa/03-execution.md` 当前均不存在，符合“后续唯一新增”的语义；`test-results/` 的目录规则实际命中示例原始 receipt 文件。

## 3. 合同到 receipt 的逐项追踪

| 门禁 | 正式合同与真实入口 | v4 操作、机械断言与 receipt | 评审、严重度与真实返工条件 |
|---|---|---|---|
| 时钟真实路径 | `Store.advanceClock` 依次执行 `DB.Begin`、读取 `console_clock`、UPDATE、`Tx.Commit`，成功后才发布 `effectiveTime`；`Server.ServeHTTP` 对每个非 health 请求调用 `server.advanceClock` | 设施由 `newClockFaultFixture(t, stage)` 构造真实 schema 3 Store，直接调用 `fixture.store.advanceClock()`；C05 同阶段另经 `New`/`ServeHTTP` 验 503、degraded 与业务/upstream 零点，receipt 为 `C05/<phase>` | **满足 / 无 blocker**。设施机械证明和产品 fail-closed 证据被明确分开，设施自测不能冒充产品 QA。若入口或 SQL 变化导致匹配失效，须停止并回独立评审 |
| 四阶段入口和计数 | 包内枚举真实存在：`clockFaultBegin/Read/Update/Commit/None`；connector/conn/tx 分别包装 Begin、目标 SELECT、目标 UPDATE、Commit | 独立 atomic 计数固定为 begin `[1,0,0,0]`、read `[1,1,0,0]`、update `[1,1,1,0]`、commit `[1,1,1,1]`；前三阶段用 `errClockFault`；辅助持久查询只在计数快照后进行 | **满足 / 无 blocker**。每阶段有具体调用入口且非目标阶段可机械判定；计数、错误身份或前后状态任一不符即 Fail，不能由执行者换入口 |
| commit-only 真实性 | `Connect` 仅在 commit 模式对该 SQLite 物理连接注册 `RegisterCommitHook`；Begin/Query/Exec 正常委托；`clockFaultTx.Commit` 调用真实底层 Commit | 必须同时得到 `[1,1,1,1]`、`commitHookHits==1`，并 `errors.As` 到 `sqlite3.Error` 且 `ExtendedCode==ErrConstraintCommitHook`；不接受 sentinel 或其他错误 | **满足 / 无 blocker**。该组合证明前三阶段已完成，错误首次来自真实 COMMIT；没有 sleep、竞态、权限、关闭 DB 或较早阶段替代。若扩展码或 hook 命中不能复现，真实返工条件是修复/复审设施，不得把它记为普通产品 Fail |
| 委托、隔离与清理 | 正常模式把四阶段全委托真实 SQLite；`sql.OpenDB` 使用每 fixture connector，不 `sql.Register`，状态非包级；`t.TempDir`、独立 DB/连接池和 `t.Cleanup(store.Close)` | `clockFaultNone` 必须无错误、`[1,1,1,1]`、hook=0，并验证持久/内存均为 Tnew；四种失败错误和计数互异，持久/内存不变 | **满足 / 无 blocker**。正常委托 receipt 为 `C05/normal-delegation`，不能代替四个失败项；setup/cleanup 失败即 Fail |
| 同版与回审 | v3 的 SHA=HEAD 悖论已改成 P/C/E/F 分层；P 是产品/构建内容与公共祖先，C 是逐路径逐摘要白名单，E 是真实执行 HEAD，F 是 P 加唯一冻结 QA 文件 | 执行前记录 E/status/祖先、P manifest、C 的 bytes/SHA/blob、唯一 QA 完整 diff 与摘要、F；每条 receipt 绑定 E/P/QA/F，运行后重算 | **满足 / High 防伪门禁**。任一产品/构建变化、非白名单差异、摘要不符、QA 之外测试变化或运行间 F 改变，证据立即作废并回独立计划评审。执行者无权用“仅文档”解释差异 |
| 范围与写路径 | 正式产品合同入口仍为 `ServeHTTP`、`advanceClock`、admit/attempt/invitation、Backup、`Restore/restoreWithRename`；v3 冻结范围为 A01–A05、C01–C07、B01–B05、X01–X03 | v4 逐项保留相同编号；唯一测试/报告路径固定；完整 stdout/stderr、截图、SQLite、archive 和运行历史只能进执行时再次验证 ignored 的 `test-results/...` 或控制器私有目录 | **满足 / High 范围门禁**。未扩张或缩水。出现额外 tracked 测试/报告、原件进 Git、receipt 缺 bytes/SHA/exit/time/source 即停止，不得临时发明范围 |
| Bearer、Restore 与固定全量 | Bearer 要求 ASCII scheme 大小写不敏感及 RFC 6750 `1*SP`；Restore 要求失败不破坏 existing target；项目固定入口为原样 `make verify` | A02 保留多种大小写；A03 把单/双/多 SP 全列正向；B05 对 absent/existing 各 10 次、唯一 sentinel、一次最终 rename、身份/mode/清单/内容和残留硬断言；X03 不论专项结果仍原样运行一次 `make verify` | **满足 / 无 readiness blocker**。已知实现失败必须成为 Fail/NO-GO，不能删项、弱化或改为 N/A；设施绿不能替代产品断言或 X03 |
| 其余 A/C/B/X | TEMP trigger 必须绑定真实单连接 Store.DB；排队须建立确定 happens-before；备份/manifest/tar/恢复须走项目真实 API | C01–C04/C06/C07 保留审计差分、barrier、回拨、重启/Restore、旧 schema；A/B/X 保留正常、边界、隔离、并发、失败恢复与跨改动，每项都有独立 receipt | **满足 / 无 blocker**。任何必跑项缺失、用 sleep 猜测、用 wall sample 冒充持久 UPDATE、或以设施测试代替产品矩阵，均不得 GO |

## 4. 执行授权边界与返工条件

后续执行者只能按 v4 操作，不得自批改计划。开始时必须重新生成完整 P manifest 和真实 E/C/QA/F 记录；本评审中的摘要是获审基线，不替代执行现场对账。正式矩阵开始后冻结唯一 QA 文件，所有 A/C/B/X 与原样 `make verify` 必须在同一 E/F 上完成。

以下任一情况立即撤销本 `READY` 并回到独立评审：产品或构建输入变化；公共祖先关系不成立；白名单路径或摘要不符；出现唯一 QA 文件以外的测试差异；时钟 SQL/driver/hook 接口变化；commit-only 不能同时证明前三阶段成功与真实 commit 唯一失败；Restore seam 不再只替换最终 rename；范围、断言、receipt 或 ignored/private 边界被弱化。

任务工作状态与业务结论分开：本文完整交付可标记 `completed`；`READY` 只表示 v4 现在具备最终可执行性。
