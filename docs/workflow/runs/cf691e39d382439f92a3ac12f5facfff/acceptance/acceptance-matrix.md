# 本轮权威验收矩阵

版本与来源见 evidence/snapshot.log 和 evidence/host-connector-receipt.md。结果区分 QA 实际执行、同轮宿主与静态核验；不把本地 Blocked/Not Run 改成 Pass。

| ID | 基线要求 | 角色/范围 | 起始状态 | 连续用户步骤 | 最终事实 | 证据类型 | 结果 | 证据位置 |
|---|---|---|---|---|---|---|---|---|
| AC01-HOST | 完整固定门禁无删减 | 宿主/同版 | 当前工作树 | 固定 make verify 全链 | 总退出 0；具体后段由串行 fail-fast 链证明，非原始尾部可见 | Contract/Connector | Pass | evidence/host-connector-receipt.md |
| AC01-QA | QA 实际完整门禁 | QA/沙箱 | 同版 | 原样 make verify | 监听拒绝、精确分类 REJECTED、exit 2 | API-E2E | Blocked | evidence/qa-make-verify.log |
| AC02-DATA | 所有旧历史不改写 | QA/106 已跟踪文件 | 研发执行前清单 | 宿主后及 QA 前后 shasum -c | 106/106 OK，无恢复操作 | Data Reconciliation | Pass | evidence/snapshot.log、evidence/final-integrity.log、../verification/history-before.sha256 |
| AC02-REG | 原缺陷可重现并转绿 | QA/临时 fixture | HEAD 原 main overlay；修复 main | 同一新进程回归运行两版 | 旧版 exit 1：成功/拒绝均覆盖历史，输出关闭吞错；修复 race exit 0 | Contract | Pass | evidence/qa-original-regression.log、evidence/qa-output-regression.log |
| AC03-OUT | 新输出可定位无硬编码 | QA/CLI | 调用方选择新路径 | stdout/stderr 直接重定向，检查回执 | 新日志含 baseline、命令、exit、REJECTED；成功输出经 expected_red fixture 断言 | Contract | Pass | evidence/qa-make-verify.log、evidence/qa-output-regression.log |
| AC04-ERR | 分类/输出异常非零 | QA/CLI | rejected、closed_stdout | 实际 helper main 子进程 | exit 1 及 stderr 输出诊断、哨兵不变 | Contract | Pass | evidence/qa-output-regression.log |
| AC04-NEG | 分类不接受环境/额外错误 | QA/分类器 | 14 类异常 | race 全分类负例 | bind/build/warning+error/timeout/green/缺多断言/race/错状态/query/skip/重复/缺 package/invalid JSON 全拒绝 | Contract | Pass | evidence/qa-output-regression.log |
| AC04-MAKE | 实际固定入口失败传播 | QA/真实失败 | 沙箱监听被拒 | make verify | 程序 exit 1、make exit 2；未伪装 PASS | API-E2E | Pass | evidence/qa-make-verify.log |
| AC05-LOCAL | 格式/module/vet/前端/build | QA/同版 | 未改产品 | 单独检查 | exit 0，27 前端 VM 通过、零 skip；不代替完整门禁或 UI | Contract | Pass | evidence/qa-non-listening-checks.log |
| AC05-UI | 实际 Go 连续 UI 旅程 | 宿主/管理员/1280和390 | 空安装、受控上游 | journeys.md UI01 | 同轮固定 browser 阶段成功，非供应商验证；无逐项尾部/截图 | UI-E2E/Connector | Pass | evidence/host-connector-receipt.md、journeys.md |
| AC05-UI-QA | QA 自己的 UI 旅程 | QA/沙箱 | history 阶段已停止 | 未执行 browser | 未取得本地浏览器事实；由 AC05-UI 同轮替代 | UI-E2E | Not Run | evidence/qa-make-verify.log |
| AC06-STATIC | 门禁/规范/说明保持一致 | QA/工程 | HEAD 加 diff | 查看输出 diff、Makefile、harness 与契约 | 仅 stdout 检查替代历史 WriteFile；Makefile 无 diff，锁/业务未变 | Static Review | Pass | evidence/final-integrity.log、acceptance-report.md |
| DELIVERY | 报告同步公共状态 | report/本 Run | 研发时点文档 | 后续同步 T001/M01/总览再写 pr.md | QA 已完成；报告节点尚未执行，不提前称交付完成 | Contract | Not Run | acceptance-report.md 第 9 节 |
