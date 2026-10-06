# 本 Run 独立验收索引

2026-10-06，合并前全量覆盖复验。当前报告：[report-03.md](report-03.md)，结论 **Go（版本13当前产品切片，next→report→原PR19更新）**。两项SSE Important/P1已在原发现层关闭；正式可信联调、Actions发布及真实供应商仍Not Run，不是发布或供应商验收通过。

- 当前权威矩阵：[matrix-03.md](matrix-03.md)，13AC本版宿主完整重跑与SSE原层级/真实HTTP分开列。
- 当前联合版本锁：[evidence/version-lock-03.json](evidence/version-lock-03.json)，86文件e3cde103…与宿主seq28/研发13一致，Workflow/manifest及原QA锁未变。
- 当前固定入口：[evidence/host-seq28.md](evidence/host-seq28.md)，实际exit0/900秒预算/81349 bytes完整平台存档；索引不是完整日志。
- QA独立：[evidence/sse-retest-03.log](evidence/sse-retest-03.log) race exit0/28.424秒；[evidence/frontend-retest-03.log](evidence/frontend-retest-03.log) 119/119 exit0；[evidence/final-audit-03.log](evidence/final-audit-03.log)最终静态审计。
- 当前辅助：[390回看截图](evidence/routing-390-03.png)、[PID8419选择轨迹](evidence/routing-selection-8419.jsonl)，不代替实际浏览器回执。

## 第二轮全量历史（版本11，不作为当前放行依据）

[report-02.md](report-02.md)保留当时产品范围Go；合并前审查揭示两项遗漏后经新版本13与seq28重新复验，不覆盖原结论：

- 当前权威矩阵：[matrix-02.md](matrix-02.md)。原四项P1均在原发现层Closed；保留未执行/Blocked行与证据层。
- 当前版本及辅助产物：[evidence/version-lock-02.json](evidence/version-lock-02.json)。11／85文件工作树852f7bfc…与宿主seq21一致，Workflow/manifest分别联合锁定；本次未新增测试改变已验源码。
- 当前宿主固定入口：[evidence/host-seq21.md](evidence/host-seq21.md)，实际exit0/900秒/79939 bytes retained log；索引不冒充完整日志，历史seq19 exit2仍保留。
- QA实际复测：[evidence/contract-retest-02.log](evidence/contract-retest-02.log) exit0；[evidence/frontend-retest-02.log](evidence/frontend-retest-02.log) 119/119 exit0。宿主真实HTTP/UI与QA免监听契约/VM分开。
- 本轮双屏概览、重启回看截图及85660选择JSONL路径与摘要在version-lock-02中；实际产物是宿主本地辅助证据，不声称独立重跑浏览器。

## 首次验收与失败历史（不再作为当前结果）

[report-01.md](report-01.md)保持当时 **No-Go**，不将历史失败改为通过：

- 权威验收结果：[matrix-01.md](matrix-01.md)，包含逐 AC、分层结果和连续旅程。
- 开放缺陷及研发行动：[remediation-01.md](remediation-01.md)。三项核验 P1、一个强制概览合同 P1；缺失用例继续补齐，不以固定门禁绿色替代。
- 宿主门禁真实成功及历次失败索引：[evidence/host-seq16.md](evidence/host-seq16.md)。完整日志保留在平台 seq16，不把摘要伪装成全量日志。
- QA 独立执行的失败：[evidence/connection-contract-01.log](evidence/connection-contract-01.log)；回归资产位于仓库 `internal/relay/qa_connection_contract_test.go`。
- 原产品与新增 QA 测试的版本分离：[evidence/version-lock-01.json](evidence/version-lock-01.json)、[evidence/qa-source-01.json](evidence/qa-source-01.json)。原 82 文件摘要仍为08；QA 测试加入后为新的83文件测试快照，不能沿用 seq16 宣称新快照门禁通过。
- 宿主本地产物归档：[1280px概览](evidence/console-first-call-1280.png)、[390px概览](evidence/console-first-call-390.png)、[1280px路由收尾](evidence/routing-1280.png)、[390px路由收尾](evidence/routing-390.png)、[原生候选选择轨迹](evidence/routing-selection-75584.jsonl)。截图仅辅助；不替代真实 Go 浏览器操作回执。

历史报告及研发失败不覆盖。当前已新增report-03/matrix-03及seq28实际门禁索引，继续使用本Run根；当前结论以上方版本13索引为准。
