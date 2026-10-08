# T001 完整多租户服务管理

- 状态：**In Review**。当前固定宿主 `make verify` 和独立 QA65 已完成，受控产品候选 Go with known issues；保留此状态仅因完整退出/正式发布未完成，不因等待已经完成的验证。
- 里程碑：M01 多租户完整服务管理控制台，Ready for Review；完整退出未闭合，未 Accepted。
- 模块：internal/relay、cmd/model-relay、Go 嵌入 UI 与 tests。
- 优先级：Critical；负责人：development；评审人：独立 QA；执行类型：TDD / Validation。
- 硬依赖：requirements-r1、architecture-r1（G2 Ready for Development）；无新需求/架构返工。
- 冲突范围：身份、Store、准入、账本、迁移与固定验证入口；report 不修改这些文件。

## 目标与权威输入

从首次安装到真实管理员登录、服务/模型/租户授权、邀请与应用 Key、普通/SSE、预算拒绝、同 ID 诊断/核对、结账更正及恢复，完成一个可接力验收的产品，不用静态原型或默认假供应商替代。

- 冻结范围和实际材料路径：[product-definition](../../product-definition.md)。version design-v0.2.0；archive SHA256 `5f51d35fdc15ed7214749f9413879b4238097f1d6b6bbe2cc0d508d8e103cef7`；prepare manifest 与 19 件实际摘要见 [QA65 integrity](../../acceptance/evidence/attempt-65/integrity.json)。Frozen for Test Implementation 不改写为完整发布批准。
- 架构、接口和运营：[architecture](../../architecture.md)、[implementation-contract](../../implementation-contract.md)、[migration-and-operations](../../migration-and-operations.md)。只引用既有事实源，不重新定义业务规则。
- 当前 AC01–AC18/H01–H08、J01–J07、商用功能五层和后续差距：[QA65矩阵](../../acceptance/acceptance-matrix-65.md)。New API/LiteLLM仅为公开功能对照，不读复制其实现；不声称全面商用对齐。

## 范围与交付结果

- 当前候选：**budget-clock-r1**；repo big91987/model-relay；分支 `workflow/4cd62ced6a4f98f50e3cca6f7596eb06`；已有 PR27/Issue25。验收 HEAD `66f1328ada75ea90e5638fbf2bd58663ac86347e` + 未提交修复；140 件 SHA256 `ca8ec50aa168cf16ac2e24edfe214aa44d06a05846c34deadbc3c7cb6c02904a`。旧已推送 head 不含本轮修复，后续提交由 Connector 绑定。
- 真实内部身份/租户隔离、应用代际与共享池、五维双范围预留、追加账本、十页 A 工作台与 schema1/2→3 正式 CLI 迁移为实现范围。费用为精确 USD 测试计价，供应商成本与租户费独立，不是供应商实价或付款证明。
- P1-1：显式零软预算拒绝；tenant/Key scope 实际 HTTP 429/零 dispatch，NULL无限和正软预算告警语义不变。P1-2：持久 UTC 高水位、锁后期限与跨进程/备份恢复阻止旧代际回拨复活；不重置既有使用量。
- 两项 P1 由 QA65 原层 Closed；QA47-01/02 历史 Closed，当前三视口摘要和 J06 同版回归通过。非目标仍为 GPU、外部支付、生产 SLO/多区 HA 等冻结基线外能力。

## 当前验收与验证契约

- [x] 冻结 archive 与 19 件材料逐条匹配；140 件源戳与研发63、seq64、QA65独立重算一致。
- [x] 固定 `make verify` 宿主 seq64 完整 exit0（timeout900s，110897 bytes，truncated=false），真实 HTTP/SSE、权限/预算/费用/重启/迁移与三视口 UI 完成。
- [x] 准确 PR27 旧版受控实际 HTTP 预期红与当前 ActualHTTP 绿分别留存，两项 P1 Closed，不用历史预期红代替当前通过。
- [x] 独立 QA65 全量审查、六组非监听 Go race exit0/24.750s、193 Node exit0/无skip；当前宿主 J01–J07、十页/摘要/故障/迁移及 UI 文件摘要复核完成。
- [ ] 独立 self_hosted 首次网页分支补测；当前以共用 provider UI 和两来源 HTTP 补强，原行 Not Run，按用户允许本候选先发布后补测。
- [ ] 正式当前候选提交、同 PR27 base/head/merge、verified-main Actions、Deployment environment_url、运行 health 和实际页面同版（H07/US08）。
- [ ] H02、H03-runtime、H04-negative、H08 平台责任方真实退出证据；不由产品测试代签。

复现：按仓库 README 准备锁定 Go/浏览器依赖，在可监听 host-only 隔离环境运行原 `make verify`，使用临时目录和端口、不访问用户配置。失败保留原日志及证据再返研发；原生稳定监听限制单独保留，不提权或重复必败命令。成功证据为 [seq64审查](../../acceptance/evidence/attempt-65/host-seq64-review.md)、[QA65报告](../../acceptance/acceptance-report-65.md)、[当前矩阵](../../acceptance/acceptance-matrix-65.md)与 [UI索引](../../acceptance/evidence/attempt-65/ui-artifact-index.json)。宿主执行、QA非监听执行、静态审查各按真实来源区分。

## 已知限制与交接

QA58-01 P3 技术枚举/窄列断词仍 Open、非阻断，由 development 后续改善。原生 QA 浏览器实操 Blocked；真实供应商/已有自托管响应与报价、外部身份/支付 Not Run。辅助 ClockRecoveryWorker 顶层 skip 不改 Pass，实际子进程恢复由父例同轮证据覆盖；完整界限见 QA65。

当前可依赖的是同版内部受控候选和已完成 tests64/QA65。seq67 发布因交接目录 artifact 失败（exit1/978 bytes/truncated=false），未提交推送；[QA68](../../acceptance/handoff-correction-68.md)已完成路径纠正。下一步 [report69](../verification/report-69.md) → publish 正式重试新提交推送 → pr 更新同一 PR27（Refs Issue25），只传 [文件清单](../verification/artifact-files-69.json)内非空实际文件，不传目录、不新建重复 PR、不重放测试副作用；产品/测试/锁/哈希内契约改变必须新门禁/受影响 QA。

2026-10-08 用户已授权测试仓合并与正式 Actions 测试部署；协调者核验后负责 verified-main Deploy Model Relay locally。preview5545仅为目标，validation隔离不替代 preview；Agent不代发动作、不启动演示服务、不把8793平台或旧main页面冒充最新产品。部署步骤及安全初始化/登录见 [本Run PR正文](../../pr.md)，未在最新正式安装实测的步骤仍待核对。

最终必须提供实际成功 Deployment 的 environment_url、对应 /admin/ 和 /healthz，绑定 merge、Actions SHA 和运行 version（main前进则证明包含关系并验证实际部署SHA）。完整退出未闭合前不标 Done/Delivered；无需用户再次阶段审批。原文在 [report66前快照](../verification/history-before-report-66.md)和 [report59前快照](../verification/history-before-report-59.md)，历史报告与失败证据均保留。
