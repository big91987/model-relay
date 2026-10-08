# design-v0.2.0 研发交付

## 目标与边界

交付冻结 design-v0.2.0 的 A 多租户模型服务管理控制台：Go 后台、真实内部身份、服务与模型、应用 Key、双范围治理、费用账本、日志诊断及迁移恢复。权威需求为 [product-definition](../product-definition.md)，架构为 [architecture-r1](../architecture.md)；全部 AC/H 与商用功能五层对照以 [QA58 矩阵](../acceptance/acceptance-matrix-58.md)为当前验收事实。外部支付、GPU 部署、多区域 HA 等不自动扩入本期。

## 当前状态

- 当前里程碑：M01 多租户完整服务管理控制台。
- 交付判断：**Verifying**。产品候选 **records-detail-entry-r1 / Go with known issues**；正式发布与完整 Pipeline 退出尚未完成，不是 Delivered。
- 基线：HEAD 98511771871cf0951ecef716bbb55deb13e18da5；分支 workflow/4cd62ced6a4f98f50e3cca6f7596eb06。产品尚未提交，不能把 HEAD 当新产品版本。
- 已完成门禁：seq57 原 make verify exit0；QA58 独立验收完成，QA47-01/02 Closed。
- 关键路径：report59 → publish 提交/推送并绑定源戳 → pr 创建同产品 PR → 协调者核验并合并测试仓 → 正式 verified-main Actions → Deployment、health 版本及真实页面验收。
- 当前未完成：H02、H03-runtime、H04、H08 平台事实与 H07 正式发布事实；不再等待已经通过的产品测试或 QA。没有需用户澄清的阻塞。
- 下一可领取责任：publish Connector；不是重新开发或重跑稳定受限的原生监听测试。

## 里程碑

| ID | 结果 | 状态 | 依赖 | 退出证据 / 未满足条件 |
|---|---|---|---|---|
| M01 | 多租户完整服务管理控制台 | Ready for Review | 冻结需求、architecture-r1 | [QA58 报告](../acceptance/acceptance-report-58.md)与当前矩阵；产品受控验收完成，完整 AC/H 发布退出未闭合，未 Accepted |

沿用单任务/单里程碑裁剪，M01 的结果、进入/退出条件仍由 T001 与上述矩阵承载，不另建平行状态文件。

## 任务摘要

| ID | 结果 | 里程碑 | 负责人 / 评审 | 状态 | 硬依赖 | 验证 |
|---|---|---|---|---|---|---|
| [T001](tasks/T001-console.md) | 完整多租户服务管理 | M01 | development / independent QA；发布由 Connector/协调者 | In Review | requirements-r1、architecture-r1 | 产品固定入口与 QA 已通过；正式发布/平台退出待证据 |

## 主要风险与变化

- QA58-01 P3：技术枚举/窄列断词 Open，由 development 后续改进；非阻断，不在 report 擅改已验代码。
- 供应商、真实已有自托管端点、外部身份/支付仍 Not Run；受控真实 HTTP/SSE 不证明供应商报价或联调。
- 1280/1440/390 J01–J07、30 页独立视觉对照和 63 图采集门槛通过；独立 self_hosted 首次网页分支仍 Not Run，详见矩阵。
- report 只改交付文档，不改变 134 件产品/合同指纹。产品、测试、锁或合同变动必须重新固定入口和 QA。

## 最近更新

2026-10-08 report59：消除重复“当前”段落，将原公共文件完整保存在 [历史快照](verification/history-before-report-59.md)，历次验证失败及 QA 报告原位保留；[交付复核](verification/report-59.md)与 [PR 正文](../pr.md)已形成。审查结论 Ready with Non-blocking Gaps，允许交接 publish，未宣告版本发布。
