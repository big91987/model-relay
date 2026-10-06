# D2 多对象权威配置

- 状态：Done；负责人：研发节点；指定评审：独立QA（2026-10-06 report-02）；收尾同步：交付节点；类型：SDD/TDD；里程碑：M01；硬依赖：D1契约。
- 权威输入：[架构](../../routing-architecture.md) §4/5、[交互](../../routing-interaction.md)、PRD AC-01/02/03/10；需求均以[本Run PRD](../../product-definition.md)为准。
- 范围／非目标：安全分页／详情／创建／乐观编辑、原子候选集合、不生成核验及适用代次；不硬删除、不新增默认供应商。
- 冲突范围：server.go、配置存储及配置回归。
- 交付物／实施结果：configuration.go唯一持久配置、revision冲突原子拒绝、GET模型列表严格合法JSON/list/id核验；10秒覆盖头/body，确定失败不记unknown，父取消独立、迟到结果隔离，不以核验清冷却；旧单上游写入口410。
- 验收结果：seq21真实配置／核验HTTP及双屏UI，独立QA原负控race复验exit0；合法／格式错误／401/429/503/network、头/body10秒、取消去重／迟到／重入、候选整数及引用原子边界通过。QA-01～03 Closed。
- 验证契约：隔离临时数据及明确标注受控上游；共享入口为 `make verify`，宿主真实执行；[seq21回执](../../acceptance/evidence/host-seq21.md)、[独立QA](../../acceptance/report-02.md)、[逐AC矩阵](../../acceptance/matrix-02.md)、[版本锁](../../acceptance/evidence/version-lock-02.json)可定位。源码不是已发布commit。
- 已知限制／交接：D3/D4可消费同一权威配置和安全状态；保存不等于连接成功，核验不生成也不证明生成可用。供应商费用及能力未联调。
- 最近成功验证：seq21版本11完整门禁及seq22独立QA；无待补产品验收或新需求／设计偏差。历史失败和返工保持在[验证记录](../verification/report.md)、[交付总览](../README.md)及原QA报告中，不改历史结论。下一步由Connector发布草稿PR，维护者继续核对真实发布依赖。
