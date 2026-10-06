# D1 保留旧资产的显式升级

- 状态：Done；负责人：研发节点；指定评审：独立QA（2026-10-06 report-02）；收尾同步：交付节点；类型：SDD/TDD；里程碑：M01；硬依赖：G2。
- 权威输入：[ADR](../../adr-001-storage-upgrade.md)、[部署契约](../../deployment-upgrade-contract.md)、PRD AC-11/12；需求均以[本Run PRD](../../product-definition.md)为准。
- 范围／非目标：schema2、身份／历史守恒、版本校验、备份恢复、只读deployment-contract及实际health schema；非目标为平台激活器、隐式迁移和新增数据合并回退。
- 冲突范围：store.go、backup.go、main.go、health及迁移测试。
- 交付物／实施结果：upgrade.go显式原子升级、密钥校验与备份；store拒绝旧schema serve；CLI只读契约／JSON结果与实际health schema。
- 验收结果：seq21精确85f6d17旧Go用户页面形成非空上游／两别名／启停密钥／成功失败历史，再正式backup→upgrade→schema2原key调用→新增UI对象→旧restore及旧服务调用；SQL仅对账。错误密钥／占用／错误版本／输出失败零迁移及幂等在固定Go回归内通过。
- 验证契约：隔离临时数据及明确标注受控上游；共享入口为 `make verify`，宿主真实执行；[seq21回执](../../acceptance/evidence/host-seq21.md)、[独立QA](../../acceptance/report-02.md)、[逐AC矩阵](../../acceptance/matrix-02.md)、[版本锁](../../acceptance/evidence/version-lock-02.json)可定位。源码不是已发布commit。
- 已知限制／交接：旧CLI主旅程及产品DEP01已验收；DEP02–04/06–07可信controller／Actions仍Not Run。回退仅恢复升级前备份，升级后新增不保留。
- 最近成功验证：seq21版本11完整门禁及seq22独立QA；无待补产品验收或新需求／设计偏差。历史失败和返工保持在[验证记录](../verification/report.md)、[交付总览](../README.md)及原QA报告中，不改历史结论。下一步由Connector发布草稿PR，维护者继续核对真实发布依赖。
