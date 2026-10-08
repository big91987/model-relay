# T001 完整多租户服务管理

- 状态：**In Review**。产品实施、固定宿主测试及独立 QA 已完成；正式发布和完整 AC/H 退出未完成，所以不标 Done。
- 里程碑：M01，多租户完整服务管理控制台，Ready for Review。
- 模块：internal/relay、cmd/model-relay、Go 嵌入 UI 与 tests。
- 优先级：Critical；负责人：development；评审人：独立 QA；执行类型：TDD / Validation。
- 硬依赖：requirements-r1、architecture-r1（G2 Ready for Development）；无新需求/架构返工。
- 冲突范围：身份、Store、准入、账本、迁移与原固定验证入口；report 不修改这些文件。

## 目标与权威输入

从首次安装到真实管理员登录、服务/模型/租户授权、邀请与应用 Key、普通/SSE、预算拒绝、同 ID 诊断/核对、结账更正及恢复，完成一个可接力验收的产品，不用静态原型或默认假供应商替代。

- 冻结范围及实际材料路径：[product-definition](../../product-definition.md)。材料 version design-v0.2.0；archive SHA256 5f51d35fdc15ed7214749f9413879b4238097f1d6b6bbe2cc0d508d8e103cef7；prepare manifest、19 文件摘要见 [QA58 integrity](../../acceptance/evidence/attempt-58/integrity.json)。Frozen for Test Implementation 状态不改为发布批准。
- 架构、接口和运维：[architecture](../../architecture.md)、[implementation-contract](../../implementation-contract.md)、[migration-and-operations](../../migration-and-operations.md)。
- AC01–AC18/H01–H08、商用功能五层与后续差距：[当前 QA58 矩阵](../../acceptance/acceptance-matrix-58.md)。不重写已测试的 Run 根追踪矩阵，不自称全面对齐 New API/LiteLLM。

## 交付物与实施结果

- 候选：**records-detail-entry-r1**；仓库 big91987/model-relay；分支 workflow/4cd62ced6a4f98f50e3cca6f7596eb06；未提交基线 HEAD 98511771871cf0951ecef716bbb55deb13e18da5。
- 134 件源戳：fb4066fa37d1e7865d3a18f72dd78fef98ae79bde17e20875af373b24771c3a6。已实现内部服务端身份/隔离、稳定应用与秘密代际、双 scope 五维预留、追加费用、十页 A 工作台、日志/分析/诊断与 schema3 迁移；逐文件在 integrity.json，不靠旧 main SHA 证明新候选。
- QA47-01 P1 管理摘要与 QA47-02 P2 确认 Token 口径已由 QA58 原层关闭；最终详情驱动修复和相邻红→绿保存在 [attempt56](../../validation/attempt-56/checks.md)。
- 费用为精确 USD 测试计价，供应商成本与租户费分离；不是实测供应商报价或付款。
- 已知限制：QA58-01 P3 技术枚举/窄列断词 Open（development 后续改进，非阻断）；首次 self_hosted 网页独立分支未单独运行；外部供应商/实际自托管凭据/外部身份支付 Not Run。完整边界见当前 QA 报告。

## 验收条件与验证契约

- [x] 冻结输入与19件材料摘要一致，产品/测试/合同134件指纹一致。
- [x] 固定 make verify 在宿主 seq57 完整 exit0（timeout900s，106669 bytes，truncated=false），真实 HTTP/SSE、权限/预算/费用/重启恢复/迁移与三视口 UI 运行完成。
- [x] 独立 QA58：193 Node、静态检查、五项非监听 Go race exit0；J01–J07、30页及关键图对照完成，产品 Go with known issues。
- [ ] 正式同产品 PR/base/head/merge、verified-main Actions、Deployment environment_url、运行 health 与实际页面同版（H07/US08）。
- [ ] H02、H03-runtime、H04、H08 对应平台责任方提供真实退出证据；不由产品测试代签。

复现：按仓库 README 准备锁定 Go/浏览器依赖，在可监听的 host-only 隔离环境运行原 make verify，测试使用临时目录/端口且不访问用户上游。失败保留原日志与证据再返研发；原生稳定监听限制单独保留，不提权或重跑必败命令。成功证据为 [seq57 审查](../../acceptance/evidence/attempt-58/host-seq57-review.md)、[QA58 报告](../../acceptance/acceptance-report-58.md)、[用户旅程](../../acceptance/user-journeys-58.md)、[视觉审查](../../acceptance/visual-review-58.md)。宿主执行与 QA 审查不是同一个运行来源。

## 交接与正式页面要求

当前可依赖结果为同指纹内部产品候选及 QA Go with known issues；剩余责任不是已经完成的 tests/QA。下一步 publish→pr，由 Connector 提交/推送、创建关联 Issue #25 的同产品 PR，返回真实 commit/PR 绑定。提交后核对134件清单与摘要，业务/测试/锁/合同变化必须返回固定 tests+QA；仅交付文档不改变产品门禁对象。

2026-10-08用户已授权测试仓合并与正式 Actions 测试部署。协调者核验 PR/QA 后负责合并和正式 verified-main Deploy Model Relay locally，preview 为用户指定5545目标，validation独立不替代preview；Agent不代发动作、不临时启动演示服务。步骤、安全私密文件初始化/登录与迁移恢复条件见 [本Run PR正文](../../pr.md)；当前未部署，不能把8793平台或旧main页面当最新产品。

最终必须给实际成功 Deployment 的 environment_url、对应 /admin/ 和 /healthz，并绑定 merge/Actions SHA/运行 version（若 main前进，证明包含关系并验证实际部署SHA）。H/供应商未验项继续 Not Run，不以“可发布候选”冒充完整最终 Go。无需用户再次批准或澄清。

原任务全文及历次状态保存在 [report59前历史快照](../verification/history-before-report-59.md)；[验证记录](../../validation/report.md)与QA22/35/38/47原报告、失败回执继续原位保留，不重写历史。
