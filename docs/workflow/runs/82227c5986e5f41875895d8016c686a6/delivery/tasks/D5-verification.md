# D5 可复验集成

- 状态：Done；负责人：研发节点；指定评审：独立QA（2026-10-06 report-02）；收尾同步：交付节点；类型：Validation；里程碑：M03；硬依赖：D1–D4实现。
- 权威输入：[验证计划](../../routing-validation.md)、[当前验收矩阵](../../acceptance/matrix-02.md)、原Makefile可信门禁；需求均以[本Run PRD](../../product-definition.md)为准。
- 范围／非目标：固定make verify、真实Go两屏UI和精确旧SHA页面生成非空数据后的backup/upgrade/restore；API仅补边界。非目标为替代独立QA、正式Actions或供应商验收。
- 冲突范围：原测试／浏览器driver及证据指纹；不修改固定入口顺序或删除历史红控。
- 交付物／实施结果：保留原格式／modules／vet／VM／Go race/HTTP／build／真服务浏览器入口，纳入迁移、多候选核验及停止回归。tests/evidence绑定实际工作树；Workflow/manifest独立版本锁。
- 验收结果：版本11 seq21原make verify完整exit0、900秒预算、79939 bytes retained/truncated=false；独立QA专项exit0／13.473秒及119/119VM exit0。13AC产品门槛满足，四项P1原层级Closed，report-02 Go。旧health-history预期红控保留，不把它作为当前失败。
- 验证契约：隔离临时数据及明确标注受控上游；共享入口为 `make verify`，宿主真实执行；[seq21回执](../../acceptance/evidence/host-seq21.md)、[独立QA](../../acceptance/report-02.md)、[逐AC矩阵](../../acceptance/matrix-02.md)、[版本锁](../../acceptance/evidence/version-lock-02.json)可定位。源码不是已发布commit。
- 已知限制／交接：可交付Connector草稿PR；本机listen/Mach限制单独保留，QA没有重跑完整监听链。DEP02–04、DEP05未覆盖边缘、DEP06–07、真实供应商Not Run。Git提交／推送／合并／部署未发生。
- 最近成功验证：seq21版本11完整门禁及seq22独立QA；无待补产品验收或新需求／设计偏差。历史失败和返工保持在[验证记录](../verification/report.md)、[交付总览](../README.md)及原QA报告中，不改历史结论。下一步由Connector发布草稿PR，维护者继续核对真实发布依赖。
