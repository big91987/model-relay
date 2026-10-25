# Issue63 tests6 返工：冷却边界回归去除墙钟竞态

Run `ba434da8039cca16f6829481112e3d68`，development 执行7，2026-10-25。仅为研发实施记录，不覆盖首次QA Fail、QA63-01、冻结计划或任何旧回执，不代签独立QA。

## 完整失败证据与定位

- tests6 固定候选为171件源码SHA `7f98fee14c6ca5337d981b66b0d7dd1ed88579123c492674c7ed1ed3da8f69e0`，HEAD `f06f05de20ffac5b8c8e772958c808379001f0e3`，功能仍在工作树。
- 原固定回执 `test-results/verification/20261025T135914Z-1386142975/state.json` / `verify.log` 保持原字节；verify.log 1,069,973 bytes，SHA `7d5029d2addda20056a3a3d8cbaf3ff62f1054e33dab729772fdcd3b918e08fe`。实际以failure-audit.json逐bytes/SHA验证为准，完整读取到EOF。
- 正式工具导出执行6平台全文到 `workflow-output/ba434da8039cca16f6829481112e3d68/command-6.log`，1,074,864 bytes，SHA `225937af5544c5c6014e2891642dd324d3cd798451e172cc79ba92adb7c143b5`，truncated=false。没有用上下文截断摘要替代原日志。
- 主gate 13:59:29.048721061Z—14:03:35.939624279Z，246.890903218秒；主gate exit2，runner state exit1，外层make/Connector exit2，未超时，source/candidate_unchanged=true。失败提前终止，不能作为完整gate成本或绿色。
- 主测试唯一失败：`TestRoutingPassiveCooldown`，`routing_test.go:190: cooling candidate visited`。该用例冷却仅10ms；故障后还有备用上游、终态记账和新增诊断持久化，再发下一请求。原测试没有控制这些真实存储/调度耗时，断言时冷却可能已合法到期；它不能据此证明产品违反冷却。
- health-history阶段旧版本真实HTTP的405负例以及runner的失败/取消/超时夹具均是原有预期负例。原健康方法主测试和runner包通过；不删除、修复或掩盖这些预期红字。

## 实际修改

相对tests6仅改 `internal/relay/routing_test.go` 中原用例，生产代码、Makefile、runner、UI、实验和其他已有源码均不变。原配置10ms、3次调用、503故障/备用成功、冷却内不派发及冷却后恢复的全部原断言保留。

使用现有 `server.latencyClock` 注入atomic时钟，初次失败时间固定；第二次请求设为故障后9ms，仍必须只到备用；第三次请求设为10ms，恰好到期必须恢复。删除原15ms sleep；不加重试、不延长产品冷却、不改期望次数或返回码。按真实路由时钟直接验证边界，而不是假定宿主存储耗时小于10ms；计费/额度/请求超时的原时钟不改。

patch工具自动生成的 `internal/relay/routing_test.go.orig` 已原字节移到 `harness-output/issue63-rework-2/routing_test.go.patch-backup.txt`，与修改前文本副本及HEAD原测试bytes/SHA一致；路径映射见manifest。曾包含该非编译备份的172件中间摘要单独保留为 `source-evidence-with-patch-backup.log`，明确不是交测版本；最终源码171件。源码快照以txt/patch保存在证据目录，不递归参与Go编译，不排除正常测试。

## 研发定向回归

- 首次本地命令因本节点PATH没有Go而exit127，完整保留 `cooldown-boundary-regression.log`；使用已存在 `/usr/local/go/bin` 修正本进程PATH，不安装新工具、不改权限。这不是测试断言失败，也没有重跑完整gate。
- `go test -race -count=10 -run '^TestRoutingPassiveCooldown$' -v ./internal/relay`：10/10通过，包6.789秒，完整命令47.330秒（包含本节点首次依赖准备）；EOF/exit0见 `cooldown-boundary-regression-run.log`。
- `go test -race -count=1 -v -run '^(TestRouting(PassiveCooldown|SelectionAndEvidence|BoundedAndStop)$|TestLatency|TestHealthMethods)' ./internal/relay`：exit0，包35.226秒，完整命令36.099秒。涵盖实际HTTP非零RPM/TPM/并发/预算、普通/SSE同跨池水位、测量年龄/配置迟到、首有效输出与长尾语义、选择诊断和健康方法。完整原日志见 `routing-health-regression.log`；不是完整固定入口或独立QA。
- gofmt、diff whitespace和局部vet结果另见 `static-check.log`。未重复实验、双屏、前端或全套make verify，旧结果不冒充本候选绿色。

## 预算核对与交接

沿正式User Input18223已批准的保守成本模型：`(1215.138100 + 46.759 + 91.440) × 1.10 = 1488.670810秒`，距1500秒11.329190秒；旧阻塞1577.2881秒、批准记录、估计风险与原字节都保留。本次不加入任何固定阶段、用例、调用、依赖或轮数，仅在原用例以受控时钟替换15ms等待；没有六文件之外会新增固定执行成本的改动。研发10次定向重复不是固定阶段，不据此追加固定重复，也不抵扣已批准的保守预留；新36.099秒相关组亦未超过已计入的46.759秒。失败gate只有246.891秒是提前终止，绝不拿它降低完整预算。预算明细见新budget.json；没有改变25分钟内层/30分钟外层或2480样本及80%/35%/75%/40次阈值。

本次实质修复原测试的时间依赖后，交tests对最终171件精确源码进行一次完整冷隔离make verify。保留旧QA Fail及两份exit1；QA63-01待本候选固定门禁和独立QA原负例/同跨池恢复复验，研发不关闭。后续仍为tests→独立QA→approval→report→固定publisher，到同版draft PR及释放后留存核验，不合并部署。

供应商/生产性能和raw trace仍Not Run；本轮没有读凭据、操作用户数据库、推送、PR、发布或改权限。命令原回执保留实际墙钟与退出码；开发驻留含平台归档、等待，不称Agent净耗时，不估token/费用。
