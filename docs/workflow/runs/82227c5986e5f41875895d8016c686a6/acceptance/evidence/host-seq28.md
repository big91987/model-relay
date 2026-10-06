# 当前版本13固定门禁回执索引

来源为平台项目测试 Connector seq28，不是 QA 自己执行完整入口。2026-10-06，`make verify`，预算900秒，真实exit0；平台保留81349 bytes脱敏日志，truncated=false。完整存档通过 `read_command_output(seq=28)` 获取；本文件仅索引，不能冒充完整日志。

源证据：Go1.25.6，HEAD `8ae4e5173aa0dc45c28ad61dbc64df20afe229a6`，86文件工作树 `e3cde10304e0bbce0dcebf671f9469ab7e5c494d8524cb9a7b7e04228d71d54f`，与QA自己版本检查及研发 [final-source-13](../../delivery/verification/final-source-13.json) 相同。已审PR head尚未移动；当前未提交实现不等于远端已经发布。

QA重点读取offset56000、66000及72000到eof81349，核对新测试、相邻停止边界、迁移与两屏旅程。另以入口回执和原Makefile顺序核对完整exit0；没有声称逐字读取所有中段。历史health-history的旧版预期红控exit1不是本版门禁失败；其分类门禁和当前包均PASS。

| 当前实际执行 | 结果与证据层 |
|---|---|
| 原health-history/format/module/vet/VM/Go race/build/browser | 完整固定入口exit0；不是旧seq21或局部替代 |
| TestSSERealHTTPOverlongLineNeverSwitchesOrCools | before_commit/after_commit PASS，4.88秒；真实隔离HTTP，primary1/secondary0/无冷却/协议错误及持久终态 |
| TestSSEOverlongLineNeverSwitchesOrCools | 两子场景PASS，4.87秒；Handler/隔离Transport/真实SQLite，非监听 |
| TestSSEDoneTerminalAttribution | success/write_failure/cancel_with_error/cancel_after_write PASS，9.74秒；确定性Writer错误/取消，非真实TCP故障复现 |
| TestRoutingRealHTTPCommittedToolsStopBoundaries | EOF/idle/cancel全部PASS，7.60秒，双候选停止 |
| 当前relay包 | PASS，279.472秒；历史分类包PASS，4.244秒 |
| 精确旧main UI非空升级/回退 | PASS：85f6d17基线，schema1→2，2别名/2key/2历史/4受控POST；不是可信controller/正式Actions证据 |
| 原首配/身份诊断/恢复旅程 | 1280/390各自PASS，实际Go服务；model PID8299 |
| 多上游连续原生键盘旅程 | 1280/390各自PASS，routing PID8419；配置、核验、优先级/3:1/停用、有限切换/未知用量、被动恢复、SSE取消零重放 |

受控上游与测试安装均隔离，未使用供应商凭据。截图和PID轨迹是本地辅助证据，见 [version-lock-03](version-lock-03.json)，不能独立证明完整旅程或生产版本。QA自己的契约复测另见 [sse-retest-03](sse-retest-03.log)，不拼写成QA重跑了真实网络链。
