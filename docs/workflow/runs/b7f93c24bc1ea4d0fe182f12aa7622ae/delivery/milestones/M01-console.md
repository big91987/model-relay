# M01 可读控制台与首次自托管网页链路

- 状态：**Accepted**；负责人：development；独立评审：QA57；交付核对：report58。
- 硬依赖：冻结design-v0.2.0/main、既有规范、固定make verify宿主及独立QA。
- 覆盖：Issue32六项AC、相关design AC01–18和J01–J07严格回归；任务：[T001](../tasks/T001-console.md) Done。

## 阶段结果
已知业务状态可理解；长地址/模型名/安全标识在1280/1440/390完整可访问；新安装首次self_hosted网页链与必要失败纠正真实完成。Accepted仅任务受控退出条件，不表示Git发布、真实供应商或Owner preview已登录通过。

## 进入条件
- [x] 冻结材料version/SHA、main基线及真实引用路径恢复，见[总览](../README.md#目标与边界)。
- [x] 原QA58-01、self_hosted UI Not Run、源码规范与固定门禁可定位。
- [x] 真实Go/临时SQLite/受控HTTP/锁定浏览器可复现；不改平台权限或重置真实账号。

## 退出条件
- [x] 逐页术语与未知值/必要技术配置保持，QA58-01当前候选分项闭合。
- [x] 三视口完整长值/横滚/关键动作与列、根不溢出及安全图实测。
- [x] 全新self_hosted网页创建→失联/401→纠正核验→关联/授权/角色/Key→普通/SSE→同ID日志费用完成。
- [x] 原Node、七包Go race/真实HTTP、预算UTC/迁移、权限/撤销/取消/错误恢复、J01–J07/summary/visual及R30-SYNC-01保持。
- [x] 原UTC/退出码/冷依赖/清理/源码一致，独立QA核实际安全图；当前QA14视觉与执行54body-unlock原路径闭合。

## 集成与验证证据
- 宿主Connector56真实Go、独立临时SQLite/受控HTTP；原固定 `make verify` 一次冷隔离exit0，2026-10-09T11:00:42.841521Z至11:11:53.351221Z。
- 154件候选 `dc667be6f0f5db892748ac8a214d799264c2d451004da8c80a5e06e6c38f338c`；HEAD `3f9cee1615e87e002a2c2300f2d29778072ec3f5`。研发/宿主/QA/report复算一致，不仅比较旧HEAD。
- 276Node零失败/零skip、七包race/HTTP/UTC迁移/build与全部三视口phase。顶层ClockRecoveryWorker设计skip保持Not Run，同轮父测试子进程恢复证据Pass。
- QA57 **Go with known issues**，限任务受控范围：[报告](../../acceptance/acceptance-report-57.md)、[矩阵](../../acceptance/acceptance-matrix-57.md)、[旅程](../../acceptance/journeys-57.md)、[图对照](../../acceptance/visual-review-57.md)。
- [344件索引](../../acceptance/evidence/attempt-57/integrity.json)、[原日志](../../acceptance/evidence/attempt-57/host/verify.log)、[原state](../../acceptance/evidence/attempt-57/host/state.json)。QA核181非单色PNG/134侧车，人工看27当前图与4基线图，不冒称全图人工或动态原生重跑。

## 已知风险与后续
- 旧空图全部历史成因/R28-04仍未知，新绿不证明根治；[QA14原No-Go](../../acceptance/acceptance-report-14.md)、[宿主54原失败](../verification/host-tests-54/archive-index.json)及全部development/host-tests原位保留。
- QA独立原生动态、真实供应商/支付/身份、Owner preview已登录分别Not Run，不用宿主受控UI拼接外部验收。
- 提交/推送/新PR/精确同源合并/Owner Actions preview尚未执行；交publish核精确候选。源码改变须重新评估，不把Accepted改述为已部署。
- 无需求/设计冲突、用户必答或权限扩展；既有授权持续有效。小型应用按范围裁剪，不补造未执行阶段文档。
