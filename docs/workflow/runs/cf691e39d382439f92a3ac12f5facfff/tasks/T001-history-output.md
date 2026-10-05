# T001 当前验证回执不再覆盖历史文件

- 状态：Done；报告对账时点：2026-10-06 05:08 +08:00。研发时点 In Review 保留在研发验证记录中。
- 里程碑：M01；模块：tests/health-history；优先级：High。
- 负责人：development；评审人：独立 QA；执行类型：TDD。
- 领取/基线：2026-10-06，`066f035bc551b2b4d031166cebdab0849084afd4`。
- 硬依赖：入口已确认 Issue #14 范围与已有运行契约；软依赖：无。
- 冲突范围：main.go、main_test.go、README、runtime-contracts.md；下游不改源码后沿用旧摘要。

## 目标与权威输入
Issue #14 要求：默认 make verify 不修改已跟踪历史文件，本次输出可定位；不删减既有门禁、不吞失败码；覆盖正常、异常分类和输出失败；本 Run 状态以真实测试/QA 闭环。

依据：[旧独立验收 QA12-HISTORY-WRITE](../../836b817833886b3188f263800b0f092f/acceptance/acceptance-report.md)、[运行契约](../../../../../.trellis/spec/backend/runtime-contracts.md)、[共享验证入口](../../../../../Makefile)、[项目说明](../../../../../README.md)。不需要新 UI/接口/存储设计。

## 变更边界
最小行为差距位于 run() 最后输出处：原先分类后无条件 os.WriteFile 历史路径且忽略 fmt.Print 错误；现在仅输出 stdout 并传播写入错误。不在调用方拦截、缓存或人工恢复历史文件。

仅改 main.go 的输出，main_test.go 的进程级回归，README 输出约定，已有 Trellis 源码契约；不改 Makefile、assess 分类器、旧基线、健康测试、模块/锁、产品源码或历史任务。无局部重构，无设计偏差；stdout-only 是最小默认选择，未加文件 flag/env。

## 交付物与验收条件
| AC | 当前结果 | 证据/剩余要求 |
|---|---|---|
| 默认保护历史日志/任务 | Passed | 独立 QA 正常/拒绝/关闭 stdout 场景哨兵不变；宿主后及 QA 前后 106 份历史哈希全 OK，无恢复操作 |
| 本轮日志独立可定位 | Passed | QA 新日志接收 stdout/stderr 的 REJECTED 回执；成功输出经独立子进程断言；调用方选择路径，无硬编码 |
| 精确分类且异常非零 | Passed | 独立 QA 保留 14 类负例及 rejected exit 1；实际监听失败分类 REJECTED，make exit 2；宿主正常历史分类依据同版固定总退出 0 |
| stdout 写入失败不吞退出码 | Passed | 独立 QA closed_stdout 子进程 exit 1，stderr 定位输出失败，历史哨兵不变 |
| 完整固定门禁与当前健康/版本/浏览器 | Passed（宿主回执层） | 本 Run seq=5 固定 make verify exit 0，与 QA 源码摘要相同；Makefile 原文及串行 fail-fast 链未变。宿主后段输出截断，不能虚构逐项尾部或截图；QA 本地完整入口仍 Blocked |
| 必要规范与交付状态同步 | Passed | 研发已实际 before-dev/check/update-spec；报告已同步 T001 Done、M01 Accepted、总览和验证结论，并写本 Run pr.md |
| 指定评审完成 | Passed | [本轮独立 QA](../acceptance/acceptance-report.md) 结论 Go，既有 P3 关闭；不使用旧 QA 放行 |

## 验证契约
环境：Go 1.25、CGO、锁定模块，Node 测试依赖；完整验证需宿主可监听及浏览器依赖。无真实供应商/用户数据库。

固定入口 `make verify`；专用回归 `GOCACHE="$PWD/.cache/go-build" GOTOOLCHAIN=local go test -race -count=1 -v ./tests/health-history`。回归以临时 workspace、真实命令 main 和受控 git/go 事件检查输出，非网络替身验收。每种情形同时核对历史哨兵；真正 make verify 的历史保护用 `verification/history-before.sha256` 与 shasum 校验。

成功：正常输出包含 baseline/source 哈希、子命令、exit_code 与精确 PASS，完整入口退出 0 且旧历史全未变化。失败：分类/监听/输出失败均非零，输出错误定位 stderr，不自动恢复历史。调用方用直接重定向到新文件存日志，保存实际 exit code；不得使用吞退出码的 tee 管道。

## 实施结果与交接
源码已形成、无 Commit/PR。新增回归先在隔离 fixture 对原实现失败（覆盖历史/输出错误被吞），修复后 race 通过。原仓库历史不受红灯复现污染。细节与全部证据见 [研发验证](../verification/M01-development.md)。

当前可依赖：stdout-only 实现、回归与规范、同轮宿主固定门禁和独立 QA Go 均已完成。验收源码摘要 `b33aa9f6f848a8cda1bf5af08ba9c2e540d0aa09f91090930ccafa6d4e636702`，详见 [验收矩阵](../acceptance/acceptance-matrix.md)、[宿主回执核验](../acceptance/evidence/host-connector-receipt.md) 和 [报告对账](../verification/M01-development.md#报告阶段对账)。报告更新不改受验源码。原实现红灯、研发/QA 本地 Blocked、宿主文本截断均保留，不把受控回归等同真实 HTTP/浏览器。无未关闭产品缺陷或研发任务阻塞。

下一步：平台 Connector 消费 [pr.md](../pr.md) 提交/推送并创建草稿 PR；这些发布动作、合并及部署在报告时点均未发生。本任务 Done 仅表示本轮修复、测试、指定评审及公共状态闭环完成，不表示代码已合并或上线。无需用户新增决定。
