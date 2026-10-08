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
