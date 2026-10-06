# D3 有界真实调用

- 状态：Done（2026-10-06 report-03独立QA产品范围Go）；负责人：研发节点；指定评审：独立QA；类型：TDD；里程碑：M02；硬依赖：D2契约。

2026-10-06：原PR #19合并前两项Important已修复并按原发现层独立复验Closed；版本13宿主seq28固定 `make verify` 完整exit0及report-03全量复验满足本任务退出条件。超长SSE首事件前/提交后均协议错误、primary1/secondary0且无冷却；DONE写失败/取消后request与attempt终态和空成功归属、概览一致。确定性Writer不冒称TCP故障。研发红控及修复过程见[返工证据](../verification/premerge-sse-01.md)，历史结果不覆盖。
- 权威输入：[架构](../../routing-architecture.md) §6/7/8、PRD R-1～R-6及AC-04～09/12；需求均以[本Run PRD](../../product-definition.md)为准。
- 范围／非目标：快照、短锁开始门、最高优先级/SWRR、最多3个不同上游、整体300秒／提交前60秒、白名单切换、30秒冷却、取消及提交止损；无后台探活、分布式／持久健康、无限重试或撤费承诺。
- 冲突范围：proxy.go、调度及尝试存储；D4消费安全事实。
- 交付物／实施结果：routing.go/routed_proxy.go调度／冷却／有界预算／逐尝试持久化；提交与取消后不重放，开始门再核对停用，旧健康代次不污染新观察。
- 验收结果：版本13 seq28真实HTTP及双屏完整重跑路由映射/权重/启停/有限切换/未知用量/健康/工具停止；新超长SSE实际HTTP双边界通过，无备用与冷却；QA独立Handler/SQLite与DONE成功对照、失败/取消纠正及概览事实均通过，report-03两项P1 Closed。
- 验证契约：隔离临时数据及明确标注受控上游；共享入口为 `make verify`，宿主真实执行；[seq28回执](../../acceptance/evidence/host-seq28.md)、[独立QA](../../acceptance/report-03.md)、[逐AC矩阵](../../acceptance/matrix-03.md)、[版本锁](../../acceptance/evidence/version-lock-03.json)可定位。版本13修复尚未提交推送；已审HEAD不是本版发布commit。
- 已知限制／交接：受控证据不等于真实供应商；缩短配置的停止测试不表示跑满300秒生产SLA。取消不能撤销已发生外部费用。
- 最近完整成功验证：版本13 seq28完整固定入口及同版独立QA report-03。下一步由报告和原发布Connector更新原Draft PR #19；无研发待验阻塞，不自行提交、推送、合并或部署。正式可信部署及供应商Not Run，历史失败和report-01/02不改。
