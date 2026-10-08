# report66 更新前公共文档历史快照

2026-10-08，Asia/Shanghai。下列是更新前原文，仅保存历史，不是当前状态；当前事实见原位置及 report-66.md。QA65、研发63与所有历史失败证据继续原位保留。

## 原位置：delivery/README.md

# design-v0.2.0 研发交付

## 目标与边界

交付冻结 design-v0.2.0 的 A 多租户模型服务管理控制台：Go 后台、真实内部身份、服务与模型、应用 Key、双范围治理、费用账本、日志诊断及迁移恢复。权威需求为 [product-definition](../product-definition.md)，架构为 [architecture-r1](../architecture.md)；全部 AC/H 与商用功能五层对照沿用 [QA58 矩阵](../acceptance/acceptance-matrix-58.md)作为历史基线，当前P1整改须新宿主门禁/QA验收。外部支付、GPU 部署、多区域 HA 等不自动扩入本期。

## 当前状态

- 当前里程碑：M01 多租户完整服务管理控制台。
- 交付判断：**Verifying**。候选budget-clock-r1两项P1实现/非监听回归完成，宿主完整tests与独立QA待验；旧records-detail-entry-r1的QA58只作为历史，不是当前Go。
- 当前基线：HEAD 66f1328ada75ea90e5638fbf2bd58663ac86347e；同分支、同PR27，未合并。本次工作树未提交。
- 历史门禁：seq57 make verify exit0/QA58，QA47-01/02 Closed；当前P1整改必须重新固定tests与独立QA。
- 关键路径：development63 → tests → QA → report → publish → pr更新PR27 → 协调者核验合并 → 正式verified-main Actions → Deployment/health/真实页面。
- 当前未完成：P1-1/2宿主红→绿及完整门禁/QA；H02、H03-runtime、H04、H08及H07正式发布事实。无需用户澄清。
- 下一责任：tests宿主；沿用稳定原生监听限制，不提权、不重复必败make verify。详见[development63](verification/development-63.md)。

## 里程碑

| ID | 结果 | 状态 | 依赖 | 退出证据 / 未满足条件 |
|---|---|---|---|---|
| M01 | 多租户完整服务管理控制台 | Active | 冻结需求、architecture-r1 | [QA58](../acceptance/acceptance-report-58.md)为历史；两项P1整改后须新make verify与QA，完整AC/H发布退出仍未闭合 |

沿用单任务/单里程碑裁剪，M01 的结果、进入/退出条件仍由 T001 与上述矩阵承载，不另建平行状态文件。

## 任务摘要

| ID | 结果 | 里程碑 | 负责人 / 评审 | 状态 | 硬依赖 | 验证 |
|---|---|---|---|---|---|---|
| [T001](tasks/T001-console.md) | 完整多租户服务管理 | M01 | development / independent QA；发布由 Connector/协调者 | In Review | requirements-r1、architecture-r1 | P1实现/局部回归完成，当前完整tests/QA待验；正式发布未完成 |

## 主要风险与变化

- QA58-01 P3：技术枚举/窄列断词 Open，由 development 后续改进；非阻断，不在 report 擅改已验代码。
- 供应商、真实已有自托管端点、外部身份/支付仍 Not Run；受控真实 HTTP/SSE 不证明供应商报价或联调。
- 1280/1440/390 J01–J07、30 页独立视觉对照和 63 图采集门槛通过；独立 self_hosted 首次网页分支仍 Not Run，详见矩阵。
- report 只改交付文档，不改变 134 件产品/合同指纹。产品、测试、锁或合同变动必须重新固定入口和 QA。

## 最近更新

2026-10-08 development63：合并前P1-1/2整改；公共状态退回实施、完整tests/QA待验，同PR27而非新建PR。局部结果和下一门槛见[development63](verification/development-63.md)。以下report59为历史。

2026-10-08 report59：消除重复“当前”段落，将原公共文件完整保存在 [历史快照](verification/history-before-report-59.md)，历次验证失败及 QA 报告原位保留；[交付复核](verification/report-59.md)与 [PR 正文](../pr.md)已形成。审查结论 Ready with Non-blocking Gaps，允许交接 publish，未宣告版本发布。

## 原位置：delivery/tasks/T001-console.md

# T001 完整多租户服务管理

- 状态：**In Review**。development63完成P1-1/2实现和必要非监听回归，当前完整宿主make verify/独立QA待重新验证；QA58为历史，非当前候选Go。
- 里程碑：M01，多租户完整服务管理控制台，Active。
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

- 当前候选：**budget-clock-r1**；仓库 big91987/model-relay；同分支 workflow/4cd62ced6a4f98f50e3cca6f7596eb06，PR27基线HEAD66f1328ada75ea90e5638fbf2bd58663ac86347e + 待验工作树。当前源戳见[attempt63](../../validation/attempt-63/checks.md)。
- 历史records-detail-entry-r1的134件源戳：fb4066fa37d1e7865d3a18f72dd78fef98ae79bde17e20875af373b24771c3a6。内部身份/隔离、代际、资源预留、追加费用、十页A工作台和schema3迁移仍为实现基线；当前修复修改产品/测试/契约，旧源戳不覆盖本候选。
- QA47-01 P1 管理摘要与 QA47-02 P2 确认 Token 口径已由 QA58 原层关闭；最终详情驱动修复和相邻红→绿保存在 [attempt56](../../validation/attempt-56/checks.md)。
- 费用为精确 USD 测试计价，供应商成本与租户费分离；不是实测供应商报价或付款。
- 已知限制：QA58-01 P3 技术枚举/窄列断词 Open（development 后续改进，非阻断）；首次 self_hosted 网页独立分支未单独运行；外部供应商/实际自托管凭据/外部身份支付 Not Run。完整边界见当前 QA 报告。

## 历史QA58验收与验证契约

- [x] 冻结输入与19件材料摘要一致，产品/测试/合同134件指纹一致。
- [x] 固定 make verify 在宿主 seq57 完整 exit0（timeout900s，106669 bytes，truncated=false），真实 HTTP/SSE、权限/预算/费用/重启恢复/迁移与三视口 UI 运行完成。
- [x] 独立 QA58：193 Node、静态检查、五项非监听 Go race exit0；J01–J07、30页及关键图对照完成，产品 Go with known issues。
- [ ] 正式同产品 PR/base/head/merge、verified-main Actions、Deployment environment_url、运行 health 与实际页面同版（H07/US08）。
- [ ] H02、H03-runtime、H04、H08 对应平台责任方提供真实退出证据；不由产品测试代签。

复现：按仓库 README 准备锁定 Go/浏览器依赖，在可监听的 host-only 隔离环境运行原 make verify，测试使用临时目录/端口且不访问用户上游。失败保留原日志与证据再返研发；原生稳定监听限制单独保留，不提权或重跑必败命令。成功证据为 [seq57 审查](../../acceptance/evidence/attempt-58/host-seq57-review.md)、[QA58 报告](../../acceptance/acceptance-report-58.md)、[用户旅程](../../acceptance/user-journeys-58.md)、[视觉审查](../../acceptance/visual-review-58.md)。宿主执行与 QA 审查不是同一个运行来源。

## 交接与正式页面要求

当前可依赖结果为P1-1/2根因修复与局部回归，完整门禁/QA尚待重新运行。下一步tests原make verify→独立QA→report→publish→pr更新同一个PR27（关联Issue25），不能重建PR、沿用seq57/QA58为新代码背书。见[development63交接](../verification/development-63.md)。

2026-10-08用户已授权测试仓合并与正式 Actions 测试部署。协调者核验 PR/QA 后负责合并和正式 verified-main Deploy Model Relay locally，preview 为用户指定5545目标，validation独立不替代preview；Agent不代发动作、不临时启动演示服务。步骤、安全私密文件初始化/登录与迁移恢复条件见 [本Run PR正文](../../pr.md)；当前未部署，不能把8793平台或旧main页面当最新产品。

最终必须给实际成功 Deployment 的 environment_url、对应 /admin/ 和 /healthz，并绑定 merge/Actions SHA/运行 version（若 main前进，证明包含关系并验证实际部署SHA）。H/供应商未验项继续 Not Run，不以“可发布候选”冒充完整最终 Go。无需用户再次批准或澄清。

原任务全文及历次状态保存在 [report59前历史快照](../verification/history-before-report-59.md)；[验证记录](../../validation/report.md)与QA22/35/38/47原报告、失败回执继续原位保留，不重写历史。

## 原位置：pr.md

# Model Relay：design-v0.2.0 多租户控制台交付候选

Refs #25 — https://github.com/big91987/model-relay/issues/25

## 交付结论与版本绑定

**产品候选 Go with known issues，可提交同产品 PR；不是整套最终 Go，也不是已部署最新页面。**

| 绑定项 | 真实事实 |
|---|---|
| Run / 工作流 | 4cd62ced6a4f98f50e3cca6f7596eb06 / revision 1；report59 |
| 产品 / 架构 | records-detail-entry-r1 / architecture-r1 |
| 材料 | design-v0.2.0，Frozen for Test Implementation；archive SHA256 5f51d35fdc15ed7214749f9413879b4238097f1d6b6bbe2cc0d508d8e103cef7 |
| 仓库 / 目标 | big91987/model-relay；base main；head workflow/4cd62ced6a4f98f50e3cca6f7596eb06 |
| 验收时 HEAD | 98511771871cf0951ecef716bbb55deb13e18da5（未提交候选基线，不是新产品提交） |
| 被测源戳 | 134 件，SHA256 fb4066fa37d1e7865d3a18f72dd78fef98ae79bde17e20875af373b24771c3a6 |
| 固定测试 | 本 Run tests seq57，make verify，timeout900s，exit0；106669 bytes，truncated=false |
| 独立 QA | QA58，产品 Go with known issues；QA47-01 P1 / QA47-02 P2 Closed；QA58-01 P3 Open/非阻断 |
| 正式发布 | report编写时尚未提交/推送、无本Run真实PR号/merge SHA/Actions部署/最新页面回执；不得预填成功 |

此正文由后续 PR Connector 原样使用，届时本 PR 页面即是同产品 PR 的真实绑定；publish/pr 回执必须记录实际提交 SHA、base/head 和 PR URL，并将已提交的134件文件与 QA58 清单逐项核对。后续 merge/部署另需实际回执，不能拿旧 main 或历史 PR 代替。

## 本次实现

- Go 管理 API、代理、SQLite schema3 与 A 明亮管理工作台十页；稳定侧栏、紧凑表格、筛选分页、统一编辑对话框及详情抽屉，服务端真实角色和租户范围。
- 供应商 API/已有自托管端点配置与核验、公共模型发布/授权/启停及真实管理摘要；不部署 GPU，不默认接假供应商。
- 应用 Key 授权交集、有效期、一次性秘密、启停/撤销、立即或最长24小时轮换；新旧代际共享额度与历史，人员退出与应用秘密责任分开。
- 租户和 Key 同时生效的次数/RPM/TPM/并发/硬软预算；未知用量保守预留，修改与轮换不清零。
- 每请求 USD 测试价格快照，成本/租户费独立、全部 dispatch 已知用量计费、未知待确认、幂等核对；已结束且无待确认才结账，追加更正保留原金额/原因/周期。
- 普通/SSE、同 Request ID 多尝试、诊断/取消/重启/迟到隔离、用量分析和 CSV、审计及旧 schema1/2→3 正式 CLI 迁移恢复。

## 实际测试与 QA

固定入口在宿主启动实际 Go 服务、临时隔离数据和受控 HTTP/SSE 上游，完成格式/模块/vet、193 Node、Go race/HTTP/恢复、build、浏览器和历史负向门禁。1280/1440/390 各自 J01–J07、management-summary、故障恢复、编辑中断、迁移及63张 visual PNG采集门槛通过。QA独立重算产品与材料摘要，另执行193 Node及5项非监听 Go race，实际审查30页和8张关键图。

宿主执行不是本report重新运行；QA原生不能监听的限制继续保留，未提权重跑。health-history 中旧版本预期红不是当前 seq57 失败；历史失败、链接器 warning、各轮 No-Go 均保留。

### 可审阅证据

以下为**已核对存在的任务分支仓库 URL**，分支与路径均显式转义，不依赖本正文位置。report编写时这些新文件尚未推送，不能宣称远端已可访问；publish 必须先推送并核对链接再创建 PR。不得改指旧 main；若推送后的目标不存在，应停止PR创建并修正交接，而非借旧文件冒充。

- [当前公共交付总览](https://github.com/big91987/model-relay/blob/workflow%2F4cd62ced6a4f98f50e3cca6f7596eb06/docs/workflow/runs/4cd62ced6a4f98f50e3cca6f7596eb06/delivery/README.md) / [任务包 T001](https://github.com/big91987/model-relay/blob/workflow%2F4cd62ced6a4f98f50e3cca6f7596eb06/docs/workflow/runs/4cd62ced6a4f98f50e3cca6f7596eb06/delivery/tasks/T001-console.md)。
- [当前独立 QA58 报告](https://github.com/big91987/model-relay/blob/workflow%2F4cd62ced6a4f98f50e3cca6f7596eb06/docs/workflow/runs/4cd62ced6a4f98f50e3cca6f7596eb06/acceptance/acceptance-report-58.md) / [全部 AC01–AC18、H01–H08及商用五层矩阵](https://github.com/big91987/model-relay/blob/workflow%2F4cd62ced6a4f98f50e3cca6f7596eb06/docs/workflow/runs/4cd62ced6a4f98f50e3cca6f7596eb06/acceptance/acceptance-matrix-58.md)。
- [连续用户旅程 J01–J07](https://github.com/big91987/model-relay/blob/workflow%2F4cd62ced6a4f98f50e3cca6f7596eb06/docs/workflow/runs/4cd62ced6a4f98f50e3cca6f7596eb06/acceptance/user-journeys-58.md) / [实际页面独立审查及P3](https://github.com/big91987/model-relay/blob/workflow%2F4cd62ced6a4f98f50e3cca6f7596eb06/docs/workflow/runs/4cd62ced6a4f98f50e3cca6f7596eb06/acceptance/visual-review-58.md)。
- [134件源戳、19件材料及84件脱敏宿主证据](https://github.com/big91987/model-relay/blob/workflow%2F4cd62ced6a4f98f50e3cca6f7596eb06/docs/workflow/runs/4cd62ced6a4f98f50e3cca6f7596eb06/acceptance/evidence/attempt-58/integrity.json) / [宿主seq57完整日志审查索引](https://github.com/big91987/model-relay/blob/workflow%2F4cd62ced6a4f98f50e3cca6f7596eb06/docs/workflow/runs/4cd62ced6a4f98f50e3cca6f7596eb06/acceptance/evidence/attempt-58/host-seq57-review.md)（完整原日志保留于平台seq57，不是这里的摘要）。
- [交付复核与后续门禁](https://github.com/big91987/model-relay/blob/workflow%2F4cd62ced6a4f98f50e3cca6f7596eb06/docs/workflow/runs/4cd62ced6a4f98f50e3cca6f7596eb06/delivery/verification/report-59.md) / [历史失败及验证记录](https://github.com/big91987/model-relay/blob/workflow%2F4cd62ced6a4f98f50e3cca6f7596eb06/docs/workflow/runs/4cd62ced6a4f98f50e3cca6f7596eb06/validation/report.md)。

## 未验范围与已知问题

- QA58-01 P3：部分技术枚举/机器字段及窄列断词，development 后续改进；不阻断当前候选，report不修改已验实现。
- 首次 self_hosted 网页独立分支 Not Run；provider真实网页与两来源API覆盖作为内部候选的等价补强，不把该分支写成已跑。
- 真实供应商及用户已有端点响应/凭据、供应商实际报价、外部身份/支付 Not Run；受控真实 HTTP/SSE证明项目语义，不证明供应商联调。
- H02重复入站、H03实际原生模型、H04负向材料恢复、H08平台唯一维护源/模板升级完整回执 Not Run。Codex/gpt-6.1-sol是指定配置，不是实际运行模型凭证。
- H07正式PR/merge/Actions/Deployment、真实最新页面及health同版 Not Run。当前候选不是全套AC/H退出通过。
- 支付充值/退款税务、GPU、其他模型协议、多区域HA/跨实例协调、生产性能/SLO及自动秘密分发属于后续差距；不声称全面对齐 New API/LiteLLM。

## 通过正式产品 Pipeline 查看最新页面

2026-10-08用户已授权测试仓合并和正式测试部署，无需再问阶段审批；以下是**待执行的正式操作要求，不是已验证部署回执**。执行责任是获授权协调者/目标维护者和 Connector，不由report代发GitHub动作。

1. publish/pr 完成本Run同产品草稿 PR，核对134件源戳和QA58绑定；产品/测试/锁/合同若变，先重新固定make verify与相应独立QA。协调者审查后合并至测试仓 main，取得真实 merge SHA。
2. main push 自动触发 [Deploy Model Relay locally 工作流](https://github.com/big91987/model-relay/blob/workflow%2F4cd62ced6a4f98f50e3cca6f7596eb06/.github/workflows/deploy-local.yml) 的 prepare：精确 main SHA、独立工作树完整make verify、按SHA构建。push只准备，不等于上线。
3. prepare 成功后，获授权仓库owner在正式 Actions 选择 main、deploy=true、target=preview；验证prepare输出与候选包含关系。用户提供 workflow ID375348269、preview端口5545为目标配置，尚不是本Run实际成功事实。validation是隔离目标，不能替代用户要看的preview。[已有正式部署说明](https://github.com/big91987/model-relay/blob/workflow%2F4cd62ced6a4f98f50e3cca6f7596eb06/docs/local-deployment.md)及[本Run迁移恢复契约](https://github.com/big91987/model-relay/blob/workflow%2F4cd62ced6a4f98f50e3cca6f7596eb06/docs/workflow/runs/4cd62ced6a4f98f50e3cca6f7596eb06/migration-and-operations.md)继续适用。
4. 旧安装必须经正式控制器停止/独占、旧binary完整备份、候选显式upgrade到schema3、切换启动及health核验；失败保留回执并走正式备份恢复，不手改数据库，不把升级后写入合并到旧快照。受控CLI迁移通过不证明正式Actions非空升级/失败恢复已完成。
5. 核对 deploy job与GitHub Deployment成功、部署SHA与运行 /healthz 的 version一致且 storage_schema=3。若main有后续提交，证明部署SHA包含PR merge并对实际SHA门禁复验；不能以旧main重新部署冒充最新成果。
6. 从实际成功Deployment的 environment_url 获取效果根地址，明确提供该根下 /admin/ 与 /healthz，并实际登录检查管理页面。保留PR merge、Actions run、Deployment ID/URL、运行version和页面验收绑定。8793是编排平台，不是产品地址；不启动临时服务或把测试截图当实际页面。

### 安全首次初始化与登录

- 服务由正式preview安装/部署流程管理，不在仓库内手动启动替代品。仅新空安装允许初始化；已有安装不得再init、清空数据或覆盖key-file。
- 首次管理员账号为 admin；获授权维护者从目标主机私有的 initial-admin-password.txt 安全取得首次密码并私密交付。密码、主密钥、应用秘密不进入PR、聊天、Actions日志、截图或仓库。初始密码文件仅记录首次值，不保证重置后仍有效。
- 在实际 environment_url/admin/ 登录。忘记/已重置密码时由受信维护者停止正式服务、备份，再使用产品 reset-password --password-stdin 从私有文件输入并重启原管理服务；不绕过身份、不在命令参数放密码。产品CLI参数见 [README运营入口](https://github.com/big91987/model-relay/blob/workflow%2F4cd62ced6a4f98f50e3cca6f7596eb06/README.md)。
- 首次可用链路：管理员真实登录→实际服务接入/核验→模型/价格与租户授权→邀请接受/应用Key一次秘密→普通/SSE→同ID日志/用量核对。无真实供应商凭据仍标Not Run；不能预设演示数据或免费调用成功。

当前发布和部署均待后续真实回执；取得实际页面后由责任方更新公共状态和退出证据，再判定完整交付。
