# design-v0.2.0 研发交付

## 目标与边界

交付冻结 design-v0.2.0 的 A 多租户模型服务管理控制台：Go 后台、真实内部身份、服务与模型、应用 Key、双范围治理、费用账本、日志诊断及迁移恢复。权威需求为 [product-definition](../product-definition.md)，架构为 [architecture-r1](../architecture.md)；全部 AC01–AC18/H01–H08、连续旅程及商用功能五层对照以 [QA65 矩阵](../acceptance/acceptance-matrix-65.md)为当前验收事实。外部支付、GPU 部署、多区域 HA 等不自动扩入本期，不声称全面对齐 New API/LiteLLM。

## 当前状态

- 当前里程碑：M01 多租户完整服务管理控制台，**Ready for Review**。
- 交付判断：**Verifying**。budget-clock-r1 的宿主完整固定入口和独立 QA65 已完成，产品候选 Go with known issues；整套 Harness、正式发布和最新实际页面退出尚未闭合，不是 Delivered。
- 当前基线：HEAD `66f1328ada75ea90e5638fbf2bd58663ac86347e` + 未提交修复；140 件被测源文件 SHA256 `ca8ec50aa168cf16ac2e24edfe214aa44d06a05846c34deadbc3c7cb6c02904a`。同分支、同 PR27，旧 PR head 不含本轮修复，未合并。
- 已完成门禁：seq64 原 `make verify` exit0，110897 bytes、truncated=false；QA65 全量证据/独立源戳审查、六组非监听 race 与 193 Node exit0；P1-1 零软预算与 P1-2 回拨复活 Closed。seq57/QA58仅保留历史，不替本候选背书。
- 关键路径：report69 文件级纠正 → publish 正式重试新提交/推送 → pr 更新 PR27 → 协调者核验合并 → verified-main 正式 Actions → Deployment、运行 health 与实际页面闭环。seq67 因目录 artifact 失败（exit1/978 bytes/truncated=false），未提交推送；不是产品门禁失败。[QA68纠正](../acceptance/handoff-correction-68.md)已复核原验收不变。
- 当前未完成：H02、H03-runtime、H04-negative、H08，以及 H07 最新候选提交/合并/正式部署。产品侧独立 self_hosted 首次 UI 分支 Not Run，用户允许本候选先发布后补测；原生 QA UI 实操 Blocked、外部凭据验收 Not Run。不重复把已完成 tests/QA 写为待验。
- 下一责任：publish Connector 使用 [report69文件清单](verification/artifact-files-69.json)中的实际非空文件路径，核对源戳未变后提交并推送；不得传目录。pr Connector 更新同一个 PR27，关联 Issue25，核对新增链接可访问。无需新需求/架构审批或用户澄清。

## 里程碑

| ID | 结果 | 状态 | 依赖 | 退出证据 / 未满足条件 |
|---|---|---|---|---|
| M01 | 多租户完整服务管理控制台 | Ready for Review | 冻结需求、architecture-r1 | [QA65](../acceptance/acceptance-report-65.md)与 seq64 同版受控产品候选 Go with known issues；完整 AC/H 的正式发布/外部退出未闭合，未 Accepted |

沿用单任务/单里程碑裁剪，M01 的结果、进入/退出条件由 T001 与上述矩阵承载，不另建平行状态文件。

## 任务摘要

| ID | 结果 | 里程碑 | 负责人 / 评审 | 状态 | 硬依赖 | 验证 |
|---|---|---|---|---|---|---|
| [T001](tasks/T001-console.md) | 完整多租户服务管理 | M01 | development / independent QA；发布由 Connector/协调者 | In Review | requirements-r1、architecture-r1 | 当前固定入口/QA65完成；正式页面/完整退出未完成，非 Done |

## 主要风险与变化

- QA58-01 P3 技术枚举/窄列断词仍 Open、非阻断，由 development 后续改进；report 不修改已验代码。
- 供应商、真实已有自托管端点、外部身份/支付仍 Not Run；受控真实 HTTP/SSE 不证明供应商报价或联调。
- 当前宿主 1280/1440/390 的 J01–J07、十页、摘要、故障及正式 CLI 迁移通过；QA65 复核 51 件 UI 文件与 30 页结构化事实、亲看三张图。不能把证据审查写成原生 QA 亲自浏览器实操。
- 140 件源戳以 QA65 清单为准，不覆盖任意仓库文件或正式部署器。产品/测试/锁/哈希内契约变动须重走固定入口及受影响 QA；仅报告/脱敏证据变化仍需绑定新提交并复核源戳。
- preview5545 是部署目标，不是当前已交付地址；8793 为编排平台，不是产品页面。正式部署步骤和安全登录说明见 [PR正文](../pr.md)，尚待目标维护者实测。

## 最近更新

2026-10-08 report69：承接 QA68，只修正文档和文件级交接；73件旧清单、140件源戳、19件材料及51件UI摘要复核一致，seq67失败保留，验收仍为QA65 Go with known issues。形成 [发布重试复核](verification/report-69.md)并更新同 PR27 的 [PR正文](../pr.md)；审查结论 Ready with Non-blocking Gaps，可交 publish 正式重试，未宣告发布或完整版本交付。[report66](verification/report-66.md)、[历史快照](verification/history-before-report-66.md)、report59/研发63及历次失败证据原位保留。
