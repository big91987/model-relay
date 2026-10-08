# Model Relay：design-v0.2.0 多租户控制台 / budget-clock-r1

Refs #25 — https://github.com/big91987/model-relay/issues/25

更新已有同产品 PR27：https://github.com/big91987/model-relay/pull/27 。不新建重复 PR。

## 交付结论与版本绑定

**当前受控产品候选 Go with known issues，可以更新 PR27；不是整套 Harness 最终 Go，也不是最新版本已部署。**

| 绑定项 | 真实事实 |
|---|---|
| Run / 工作流 | 4cd62ced6a4f98f50e3cca6f7596eb06 / revision 1；report69（文件级发布重试纠正，保留report66） |
| 产品 / 架构 | budget-clock-r1 / architecture-r1 |
| 材料 | design-v0.2.0，Frozen for Test Implementation；archive SHA256 5f51d35fdc15ed7214749f9413879b4238097f1d6b6bbe2cc0d508d8e103cef7 |
| 仓库 / 目标 | big91987/model-relay；base main；head workflow/4cd62ced6a4f98f50e3cca6f7596eb06 |
| 验收时 HEAD | 66f1328ada75ea90e5638fbf2bd58663ac86347e + 未提交修复；旧head不是包含本轮修复的新提交 |
| 被测源戳 | 140件，SHA256 ca8ec50aa168cf16ac2e24edfe214aa44d06a05846c34deadbc3c7cb6c02904a；范围以QA65源文件清单为准，不是全仓任意文件哈希 |
| 固定测试 | tests seq64原make verify，timeout900s，exit0；110897 bytes，truncated=false；QA完整读取至eof |
| 独立 QA | QA65，Go with known issues；零软预算P1-1、回拨复活P1-2 Closed；QA58-01 P3 Open/非阻断 |
| 发布事实 | seq60已推送旧head、seq61已有PR27；seq67发布exit1/978 bytes/truncated=false，因目录artifact被拒，未提交推送。QA68/report69已纠正文件级交接，新修复/QA65/报告仍待正式publish；尚无本候选merge、正式Actions/Deployment或最新页面回执 |

publish Connector须将修复及QA65证据提交推送到同一任务分支；pr Connector更新本PR并绑定实际新提交SHA/base/head。仅报告/脱敏证据变化也要复核140件源戳；产品、测试、锁或哈希内契约变化须重新固定入口和受影响独立QA。不能用旧PR head、QA58/seq57证明当前修复已发布。

本次重试仅纠正交接路径：发布失败原因为 `acceptance/evidence/attempt-65/host` 是目录而非同Run根内非空文件。report69承接QA68，核对73件旧清单、140件源戳、19件材料、51件UI字节一致，不修改产品或Frozen输入、不重跑已受限入口，QA65两项P1 Closed和Go with known issues不变。旧清单保留为执行68时快照；更新后的文件清单供publish逐项使用，不放宽权限或递归目录。正式推送后才核对新增目标远端可访问性。

## 本次实现与整改

- Go管理API、代理、SQLite schema3及A明亮十页工作台；稳定侧栏、紧凑表格、筛选分页、统一编辑对话框与详情抽屉，真实服务端身份/租户范围。
- 供应商API/已有自托管端点配置与核验、公共模型发布/授权/启停和管理摘要；不部署GPU，不默认使用假供应商。
- 应用Key授权交集、期限、一次性秘密、启停/撤销、立即或最长24小时轮换；新旧共享额度/历史，人员退出与应用Key责任分开。
- 租户和Key同时生效的次数/RPM/TPM/并发/硬软预算；未知保守预留，修改与轮换不清零。
- 精确USD测试价格快照、独立成本/租户费、全部dispatch已知收费/未知待确认、幂等核对；仅已结束且无待确认周期结账，更正追加原金额/原因/调整周期。
- 普通/SSE、同Request ID多尝试、诊断/取消/重启/迟到隔离、分析/CSV/审计及schema1/2→3正式CLI迁移恢复。
- P1-1修复显式零软预算：tenant/Key实际HTTP 429/零dispatch；NULL无限和正软预算超限告警不变。
- P1-2通过持久UTC高水位、锁后期限、跨进程Open及Backup/Restore防止过期未retire代际回拨复活，期限和旧使用事实不因回拨重置。

## 实际验证与来源

宿主seq64原固定入口保留格式/模块/vet、193Node、全仓Go race/实际HTTP/SSE/费用恢复、build与browser。在临时隔离数据/端口和受控HTTP/SSE上游，启动真实Go服务完成1280/1440/390各J01–J07、十页、管理摘要、故障恢复、失响应/双提交、编辑中断和正式CLI迁移；不使用用户上游或数据。

准确PR27旧版实际HTTP红灯是**修复后隔离重建**，不是原开发时网络回执：零软预算200/dispatch1，旧代际截止401后回拨200/dispatch1；当前ActualHTTP绿另行记录，两项P1由QA65原层关闭。health-history旧版预期红与当前HealthMethodsHTTP绿也分开保留；历史expected failure不能计为当前失败或代替当前绿灯。

QA65独立重算140件源字节及archive/19材料，亲自执行六组非监听Go race（24.750s/exit0）和193Node（exit0/无skip）；复核51件当前UI留存文件、30页结构化事实，亲看三张PNG。宿主执行、QA非监听执行、静态图片审查不互换：**QA原生独立浏览器实操仍Blocked**，未提权或重跑已知禁监听make verify。ClockRecoveryWorker辅助顶层skip保持原事实，父例同轮实际子进程恢复另有通过证据。

### 可审阅仓库证据

以下URL的实际仓库、转义分支和目标路径已在checkout核对；新QA65/修复/报告须publish先提交推送，pr再更新正文并核对远端可访问性。report69不声称新增目标此刻远端可访问，不链旧main或旧提交。冻结材料由prepare保留，公开摘要见integrity，不链接忽略的材料缓存。

- [当前交付总览](https://github.com/big91987/model-relay/blob/workflow%2F4cd62ced6a4f98f50e3cca6f7596eb06/docs/workflow/runs/4cd62ced6a4f98f50e3cca6f7596eb06/delivery/README.md) / [任务包T001](https://github.com/big91987/model-relay/blob/workflow%2F4cd62ced6a4f98f50e3cca6f7596eb06/docs/workflow/runs/4cd62ced6a4f98f50e3cca6f7596eb06/delivery/tasks/T001-console.md) / [report66复核](https://github.com/big91987/model-relay/blob/workflow%2F4cd62ced6a4f98f50e3cca6f7596eb06/docs/workflow/runs/4cd62ced6a4f98f50e3cca6f7596eb06/delivery/verification/report-66.md)。
- [独立QA65报告](https://github.com/big91987/model-relay/blob/workflow%2F4cd62ced6a4f98f50e3cca6f7596eb06/docs/workflow/runs/4cd62ced6a4f98f50e3cca6f7596eb06/acceptance/acceptance-report-65.md) / [全部AC01–AC18、H01–H08、J01–J07和商用五层矩阵](https://github.com/big91987/model-relay/blob/workflow%2F4cd62ced6a4f98f50e3cca6f7596eb06/docs/workflow/runs/4cd62ced6a4f98f50e3cca6f7596eb06/acceptance/acceptance-matrix-65.md)。
- [seq64完整日志恢复及来源审查](https://github.com/big91987/model-relay/blob/workflow%2F4cd62ced6a4f98f50e3cca6f7596eb06/docs/workflow/runs/4cd62ced6a4f98f50e3cca6f7596eb06/acceptance/evidence/attempt-65/host-seq64-review.md) / [140件源戳与19件材料摘要](https://github.com/big91987/model-relay/blob/workflow%2F4cd62ced6a4f98f50e3cca6f7596eb06/docs/workflow/runs/4cd62ced6a4f98f50e3cca6f7596eb06/acceptance/evidence/attempt-65/integrity.json)。
- [51件当前UI文件索引](https://github.com/big91987/model-relay/blob/workflow%2F4cd62ced6a4f98f50e3cca6f7596eb06/docs/workflow/runs/4cd62ced6a4f98f50e3cca6f7596eb06/acceptance/evidence/attempt-65/ui-artifact-index.json) / [当前宿主页面证据目录](https://github.com/big91987/model-relay/tree/workflow%2F4cd62ced6a4f98f50e3cca6f7596eb06/docs/workflow/runs/4cd62ced6a4f98f50e3cca6f7596eb06/acceptance/evidence/attempt-65/host)。这是受控验收页面，不是已部署效果页。
- [QA亲自Go日志](https://github.com/big91987/model-relay/blob/workflow%2F4cd62ced6a4f98f50e3cca6f7596eb06/docs/workflow/runs/4cd62ced6a4f98f50e3cca6f7596eb06/acceptance/evidence/attempt-65/qa-inprocess.log) / [QA亲自Node日志](https://github.com/big91987/model-relay/blob/workflow%2F4cd62ced6a4f98f50e3cca6f7596eb06/docs/workflow/runs/4cd62ced6a4f98f50e3cca6f7596eb06/acceptance/evidence/attempt-65/qa-node.log)。
- [研发63整改交接](https://github.com/big91987/model-relay/blob/workflow%2F4cd62ced6a4f98f50e3cca6f7596eb06/docs/workflow/runs/4cd62ced6a4f98f50e3cca6f7596eb06/delivery/verification/development-63.md) / [历史失败及验证记录](https://github.com/big91987/model-relay/blob/workflow%2F4cd62ced6a4f98f50e3cca6f7596eb06/docs/workflow/runs/4cd62ced6a4f98f50e3cca6f7596eb06/validation/report.md)。
- [QA68文件级纠正与seq67失败](https://github.com/big91987/model-relay/blob/workflow%2F4cd62ced6a4f98f50e3cca6f7596eb06/docs/workflow/runs/4cd62ced6a4f98f50e3cca6f7596eb06/acceptance/handoff-correction-68.md) / [report69发布重试复核](https://github.com/big91987/model-relay/blob/workflow%2F4cd62ced6a4f98f50e3cca6f7596eb06/docs/workflow/runs/4cd62ced6a4f98f50e3cca6f7596eb06/delivery/verification/report-69.md) / [实际非空文件清单](https://github.com/big91987/model-relay/blob/workflow%2F4cd62ced6a4f98f50e3cca6f7596eb06/docs/workflow/runs/4cd62ced6a4f98f50e3cca6f7596eb06/delivery/verification/artifact-files-69.json)。网页证据目录链接仅供浏览，不作为handoff artifact。

## 未验范围与已知问题

- QA58-01 P3技术枚举/窄列断词Open、非阻断，development后续改进；report不改已验实现。
- 独立self_hosted首次网页分支Not Run；共用provider完整UI与两来源HTTP补强内部候选，用户允许本候选先发布后补测，原分支不改Pass。
- 真实供应商、已有端点现货/凭据/报价、外部身份/支付Not Run；受控HTTP/SSE只证明本项目语义。可信收费Token上界须由运营者依据供应商契约提供；无法证明则有限TPM/硬费用拒绝，不能无条件声称硬上限。
- H02重复入站、H03-runtime实际原生模型凭证、H04-negative缺文件/错SHA/过期恢复、H08平台维护源/正式升级完整回执Not Run；Codex/gpt-6.1-sol是指定配置，不是实际运行模型凭证。
- H07已有PR27/旧推送，但当前新head/merge/Actions/Deployment/实际页面及health同版闭环Not Run。独立宿主残留进程/端口External State核查Not Run。
- 支付充值/退款税务、GPU、其他协议、多区HA/跨实例协调、生产性能/SLO和自动秘密分发是后续差距，不声称全面商用对齐。

## 通过正式产品Pipeline查看最新页面

2026-10-08用户已授权测试仓合并和正式测试部署；以下为**待执行/待最新正式安装实测的要求，不是已验证部署手册或成功回执**。责任为协调者/目标维护者及Connector，report不代发GitHub动作、不临时启动服务。

1. publish/pr先更新同PR27，绑定新提交/140件源戳/QA65/seq64，保留Issue25关联。协调者核验后合并测试仓main，记录实际merge SHA；不能合并旧head冒充修复。
2. main push自动触发 [Deploy Model Relay locally工作流](https://github.com/big91987/model-relay/blob/workflow%2F4cd62ced6a4f98f50e3cca6f7596eb06/.github/workflows/deploy-local.yml) 的prepare：独立工作树精确main SHA、完整make verify、按SHA构建。push/prepare只准备，不等于上线。
3. 获授权仓库owner在正式Actions选择main、deploy=true、target=preview。dispatch先prepare，再由deploy job激活，环境为local-preview。用户提供workflow ID375348269、preview5545仅是目标配置；validation隔离不替代preview。遵循 [已有正式部署说明](https://github.com/big91987/model-relay/blob/workflow%2F4cd62ced6a4f98f50e3cca6f7596eb06/docs/local-deployment.md)及 [本Run迁移运营契约](https://github.com/big91987/model-relay/blob/workflow%2F4cd62ced6a4f98f50e3cca6f7596eb06/docs/workflow/runs/4cd62ced6a4f98f50e3cca6f7596eb06/migration-and-operations.md)，不另设绕过门禁的部署器。
4. 旧安装须正式控制器停服务/独占、旧binary完整备份、候选显式upgrade到schema3、切换启动及health核验。失败保留回执走正式备份恢复，不手改数据库、不把升级后写入并进旧快照。受控CLI升级通过不证明正式Actions非空升级/失败恢复已发生。
5. 核对deploy job及GitHub Deployment成功、部署SHA与 /healthz.version 一致、storage_schema=3。main前进则证明部署SHA包含PR merge并核验实际部署SHA门禁；不能重部署旧main冒充最新成果。
6. 从**实际成功Deployment的environment_url**取得效果根地址，提供其 /admin/ 和 /healthz，真实登录查看管理页。保留PR新head/base/merge、Actions run、Deployment ID/environment_url、运行version与页面验收绑定。未得回执不预写成功URL；8793是编排平台，测试截图不能替代最新效果。

### 安全首次初始化与登录（待最新正式安装复核）

- 正式preview由现有安装/部署流程管理。仅新空安装允许init；已有安装不得重复init、清数据或覆盖key-file。
- 首次账号admin；获授权维护者从目标主机私有initial-admin-password.txt安全取得并私密交付首次密码。产品CLI init一次输出秘密，正式控制器必须捕获到私有文件，绝不进Actions日志/PR/聊天/截图/仓库；首次值不保证重置后有效。
- 在实际environment_url/admin/登录，Host保持与监听authority一致，不随意把127.0.0.1换成localhost；不公布密码、应用Key、主密钥或本机绝对路径。
- 密码遗失/重置，由受信维护者停止正式服务、完整备份，按产品reset-password --password-stdin和正式data-dir/key-file从私有输入设置并重启原管理服务；不把密码放命令参数、不绕过身份。见 [README运营入口](https://github.com/big91987/model-relay/blob/workflow%2F4cd62ced6a4f98f50e3cca6f7596eb06/README.md)。
- 首次链路：真实管理员登录→服务接入/显式核验→模型/独立价格/租户授权→邀请接受→Key一次秘密→普通/SSE→同ID日志/用量核对。默认空安装无演示供应商或免费生成；保存、核验、生成成功、用量完整和结账是不同事实，无凭据时供应商联调仍Not Run。

当前已完成受控产品门禁和QA65；最新发布/正式部署/实际页面待后续真实回执。完整退出闭合并记录后才可更新Done/Accepted/Delivered。
