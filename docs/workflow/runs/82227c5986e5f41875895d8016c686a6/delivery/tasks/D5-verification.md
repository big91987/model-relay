# D5 可复验集成

- 状态：Done（2026-10-06 report-03独立QA产品范围Go）；负责人：研发节点；指定评审：独立QA；类型：Validation；里程碑：M03；硬依赖：D1–D4实现。

2026-10-06：原PR #19合并前两项Important已修复并按原发现层独立复验Closed；版本13宿主seq28固定 `make verify` 完整exit0及report-03全量复验满足本任务退出条件。超长SSE首事件前/提交后均协议错误、primary1/secondary0且无冷却；DONE写失败/取消后request与attempt终态和空成功归属、概览一致。确定性Writer不冒称TCP故障。研发红控及修复过程见[返工证据](../verification/premerge-sse-01.md)，历史结果不覆盖。
- 权威输入：[验证计划](../../routing-validation.md)、[当前验收矩阵](../../acceptance/matrix-03.md)、原Makefile可信门禁；需求均以[本Run PRD](../../product-definition.md)为准。
- 范围／非目标：固定make verify、真实Go两屏UI和精确旧SHA页面生成非空数据后的backup/upgrade/restore；API仅补边界。非目标为替代独立QA、正式Actions或供应商验收。
- 冲突范围：原测试／浏览器driver及证据指纹；不修改固定入口顺序或删除历史红控。
- 交付物／实施结果：保留原格式／modules／vet／VM／Go race/HTTP／build／真服务浏览器入口，纳入迁移、多候选核验及停止回归。tests/evidence绑定实际工作树；Workflow/manifest独立版本锁。
- 验收结果：版本13 seq28原make verify完整exit0，900秒预算、81349 bytes完整retained/truncated=false。独立QA专项race 7顶层exit0/28.424秒及119/119VM exit0，13AC产品范围满足，两项合并前P1 Closed、原四项未复发；report-03 Go。旧health-history预期红控保留。
- 验证契约：隔离临时数据及明确标注受控上游；共享入口为 `make verify`，宿主真实执行；[seq28回执](../../acceptance/evidence/host-seq28.md)、[独立QA](../../acceptance/report-03.md)、[逐AC矩阵](../../acceptance/matrix-03.md)、[版本锁](../../acceptance/evidence/version-lock-03.json)可定位。版本13修复尚未提交推送；已审HEAD不是本版发布commit。
- 已知限制／交接：Connector已提交推送固定HEAD并创建Draft PR #19，尚未合并或部署。本轮不自行Git发布。本机listen/Mach限制单独保留，完整新入口已由seq28宿主通过，QA未重跑监听链；DEP02–04、DEP05未覆盖边缘、DEP06–07、真实供应商Not Run。
- 最近完整成功验证：版本13 seq28完整固定入口及同版独立QA report-03。下一步由报告和原发布Connector更新原Draft PR #19；无研发待验阻塞，不自行提交、推送、合并或部署。正式可信部署及供应商Not Run，历史失败和report-01/02不改。
