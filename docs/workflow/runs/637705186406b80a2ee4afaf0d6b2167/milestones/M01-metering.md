# M01 内部安全与计量可复验

- 状态：Accepted
- 负责人：development；独立评审：QA执行10；状态汇总：report执行11（2026-10-09）。
- 硬依赖/进入依据：同版源码、61项材料及交接ZIP核验、既有实现契约及非空Trellis规范。
- 覆盖范围：原AC09–13/16及关联AC07/17的内部安全/计量反馈环，不承诺原#33真实供应商验收。

## 阶段结果与退出条件

[T001](../tasks/T001-metering.md) 已形成受控普通/SSE安全默认、非零阈值/窗口、known/unknown计量、并发原子性及池隔离数值回归；接入原门禁，既有HTTP/真实进程/三宽Go UI完整通过。确认生产缺陷0，不称修复未证实的算法故障。

- [x] 权威原文和同版实现可定位，冻结设计未改。
- [x] 安全/计量矩阵具备同ID、dispatch、双scope及精确金额副作用证据。
- [x] tests9同候选完整 `make verify` 冷安装exit0，原门禁全部保留。
- [x] QA10独立受控计划A/B评审、受影响回归及版本/证据核对完成。
- [x] 确认缺陷与未验范围分列，无确认生产缺陷，不伪关闭供应商候选。

## 集成与验证证据

基线HEAD `3f9cee1615e87e002a2c2300f2d29778072ec3f5`；154文件受测工作树SHA256 `2dfe14a3c305c13621f23b7dd287a8313b7e4f39ab64b8281d283450d125d8ed`，尚无新增提交/发布版本，report重算一致。tests9为自有隔离冷安装、Go1.25.6、真实HTTP/进程及Playwright受控上游；`make verify` 总698.502秒，原720秒deadline不变，Connector/state/gate均0。

[独立验收](../acceptance/release-assessment-qa-10.md)、[逐项矩阵](../acceptance/acceptance-matrix-qa-10.md)、[计划评审](../acceptance/test-plan-review-qa-10.md)、[证据索引](../acceptance/evidence/index.md) 支持Accepted。[测试计划](../verification/test-plan.md) 是研发快照；独立评审结论以QA10为准。[tests5失败](../verification/host-return-exec6.md)、[tests7失败](../verification/host-return-exec8.md) 保留，不把累计超时归因成确认冷却/J03产品缺陷。

## 已知风险与后续

Accepted只适用于M01内部资产。原#33真实普通/SSE、恢复、Token/费用、非零数值网页/自然窗口/双浏览器仍Not Run/Blocked，外层worker Skip不伪记Pass。门禁容量余量约21.5秒，不保证未来运行不超时。

关键路径剩余为Connector发布/PR与正式候选/Actions门禁；尚未提交、推送、合并或部署。真实复测交专用合法配置Pipeline，按原24/64/90及未知历史保守预算先完成前置/独立计划评审；新增真实生成0。无新迁移/生产配置/运营扩容，不另造上线运营文档，无需用户重批普通阶段。
