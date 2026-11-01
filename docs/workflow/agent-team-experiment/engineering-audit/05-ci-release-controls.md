# CI/CD 与发布回滚控制静态审查

## 结论与边界

审查基线为候选 `122c9abece28f562616483e991098d020d236d6d`。三项前置在本地 Git 历史中均可达：Bearer 认证合入 `2d48ee08f12368e1b806fcc0975943e2eca0c8ee`（任务提交 `67f464494f4d59be8a20ef99d9b15175ecd00e78`）、单次时钟推进合入 `532e713acc998162c682b310c8f280e1651e6f95`（任务提交 `318639d819165f3309f123133fdaf6229003270c`）、严格备份清单合入 `122c9abece28f562616483e991098d020d236d6d`（任务提交 `7dc57094557b8883c130a41ed306d1223c0d92f8`）。这是 **已验证（本地 Git 对象和祖先关系）**，不是远端分支或 GitHub 合入状态复核。

结论：仓库已有一条设计较完整的本机预览发布控制链，包含精确 main SHA、项目固定验证入口、摘要绑定、显式部署、健康核对、迁移前备份、失败回滚和中断恢复。然而，仓库没有在 PR 或普通 push 上运行产品 `make verify` 的独立 CI；验证第一次成为发布强制条件是在代码已经进入 main 后的 prepare。发布制品、验证日志、计划和回滚证据仅保存在受信自托管主机的私有部署根，没有进入 GitHub Actions artifact/attestation 留存链。故本次静态审查完成不代表候选可发布、部署可用或产品 QA 通过。

审查仅读取配置、维护源和 Git 历史；未调用 GitHub/平台 Connector，未读取 Secrets，未运行 workflow、`make verify`、控制器、预览或部署。下文“已验证”均指仓库静态内容或本地 Git；“推断”指依赖 GitHub/Runner/Environment 外部状态；“未执行”明确表示没有运行态证据。

## 发现（按优先级）

### P1-01：没有合入前的产品 CI 门禁

- **已验证：** 四个 workflow 中，Issue 入口只转发事件（`.github/workflows/agent-platform-entry.yml:1-74`、`.github/workflows/agent-platform-sandbox-entry.yml:1-74`）；独立审查工作流执行平台审查而非产品 `make verify`（`.github/workflows/agent-platform-code-review.yml:1-90`）；唯一调用项目验证的链路是 main push/手动发布的部署 prepare，且实际调用藏在主机控制器 `prepare` 中（`.github/workflows/deploy-local.yml:3-55`；`.harness/agent-platform/examples/github/model-relay-preview/controller.py:541-619`）。
- **影响：** 缺陷可以先进入 main，随后 prepare 才失败。若分支保护把某个审查 status 当成测试，它仍不能替代 Go race、HTTP、浏览器和 build 门禁。main 可保持未部署，但集成分支质量与可发布性会被破坏。
- **建议（仓库基建，P1）：** 增加只读 PR/push CI，直接调用项目拥有的 `make verify` 或等价拆分门禁；使用 GitHub 托管或隔离的一次性 Runner，最小 `contents: read` 权限，并把该 check 配为 main 合入必要条件。部署 prepare 仍应重新验证精确 SHA，不能因新增 CI 而删除。
- **状态：** 分支保护和 required checks 是 **未执行/不可从仓库验证**，因此不声称当前合入确实无门禁。

### P1-02：发布信任集中于常驻自托管 Runner、外置控制器和可变主机路径

- **已验证：** 所有入口均运行在 `[self-hosted, macOS, ARM64, he-full]`（例如 `.github/workflows/deploy-local.yml:27,59`）。部署 job 不 checkout 仓库控制器，而是从仓库变量选择的 `DEPLOY_ROOT/controller.py` 直接执行（`:35-55,67-84`）。维护源的安装记录包含控制器摘要，文档要求固定版本安装（`.harness/agent-platform/examples/github/model-relay-preview/README.md:14-23`），但 workflow 每次执行前没有独立重算并与仓库内期望摘要比较。
- **影响：** Runner 用户、部署根或仓库变量一旦被错误修改/攻陷，可在 GitHub token 和本机产品数据边界内执行不同控制器；代码 SHA 和产品二进制摘要无法证明控制器本身未漂移。Issue 入口也在同一类常驻主机执行仓库代码，扩大持久化污染面。
- **建议（通用平台 + 仓库基建，P1）：** 将事件/PR 验证迁到短生命周期隔离 Runner；部署 Runner 单独分组并只允许受保护 environment/job。激活前让一个最小、仓库内按完整提交固定的启动器校验 `controller.py`、安装记录和配置摘要，或把签名/摘要固定进受保护部署配置。主机权限、Runner group、变量修改审计和轮换属于平台控制。
- **状态：** 主机 ACL、Runner group 和变量保护为 **未执行**；风险判断为 **推断**，不等同于发现主机已被篡改。

### P1-03：发布制品和验证/回滚回执没有 Actions 级留存与版本证明

- **已验证：** prepare 在私有根创建独立 worktree，日志命名 `verify-<sha>-<随机>.log`，构建后把 `sha` 与 `binary_sha256` 写入 `ready.json`（控制器 `:568-632`）；激活再次比较计划、制品摘要、main 和契约（`:837-859`），部署记录保存 SHA、前版 SHA、摘要、契约和 schema（`:673-681`）。但 deploy workflow 没有 `upload-artifact`、artifact retention、签名或 provenance 步骤（`.github/workflows/deploy-local.yml:1-84`）。只有独立审查报告上传 30 天，名称为 `code-review-<engine>-<run_id>`（`.github/workflows/agent-platform-code-review.yml:83-90`）。
- **影响：** 主机丢失、被清理或被入侵后，GitHub 侧无法独立重建“哪份二进制通过了哪次验证、由哪次计划部署、回滚用了哪份备份”的证据；发布审计依赖同一故障域。review artifact 名称也未包含 head SHA/run attempt，人工关联更易出错。
- **建议（仓库基建 + 工程工具，P1）：** prepare 结束后上传不含秘密的 manifest（目标 SHA、二进制摘要、契约摘要、门禁结果、controller 摘要、run id/attempt），以 SHA+run attempt 命名并设明确留存期；需要跨主机发布时，再引入签名 provenance。私有验证全文与含路径/可能敏感输出的备份不得直接上传。回滚结果应产生对应的脱敏状态 manifest。
- **状态：** 私有根中的真实文件与 GitHub artifact 列表均 **未执行**。

### P2-01：发布审批强度部分依赖仓库外 Environment 配置

- **已验证：** push main 只运行 prepare；deploy job 仅在 `workflow_dispatch`、actor 为 repository owner、`deploy=true` 且 prepare 输出允许时运行（`.github/workflows/deploy-local.yml:4-18,25-30,56-63`）。job 绑定 `local-preview`/`local-validation` Environment，并限制 workflow ref 必须为 main（`:26,61-63`）。文档明确私有仓库套餐可能没有 required reviewer，因此主要控制是 Owner 显式触发（维护源 README `:21-23`）。
- **影响：** 审批不是双人控制；Owner 账户失陷或误操作可准备并激活。Environment 的 branch restriction/reviewer 状态不在 Git 中，配置漂移无法由 PR review 捕获。
- **建议（仓库流程 + 通用平台，P2）：** 在套餐允许时启用 required reviewers、禁止 self-review，并定期导出/核对 Environment 与 Runner group 配置；至少把 Owner、目标、SHA、run id 和批准时间写入脱敏发布 manifest。若项目仍明确接受单人授权，应把这是风险接受而非“双人审批”写清。
- **状态：** Environment 当前真实设置为 **未执行**；文档中的“限制 main”是宣称，不作为已验证事实。

### P2-02：失败恢复设计强，但恢复未完成后的运营闭环仍靠人工重入

- **已验证：** 激活前对旧版本执行正式 backup、检查非空、持久化 SHA-256 后才升级（控制器 `:885-920`）；候选健康失败会调用 recovery（`:939-955`）。恢复会校验旧二进制、主密钥和回滚备份摘要，隔离失败数据，用旧二进制 restore 到新目录，原子移回并核验旧版本 health/schema（`:711-816`）。中断 activation 会阻止新部署，要求从相同显式入口恢复并再次部署（`:819-836`）。文档也明确 fault fixture 的 exit 0 不证明恢复，需另核对并再次正常发布（维护源 README `:35-48,54-60`）。
- **影响：** 这是合理的故障关闭行为，但 workflow 没有自动通知、恢复 runbook 链接、超时后升级路径或外部备份保留策略；`recovery incomplete` 时服务可能持续不可用，必须依赖维护者发现并重入。
- **建议（仓库流程 + 通用平台，P2）：** 为失败/中断状态增加不含秘密的告警和 runbook 链接，定义责任人、RTO、备份保留/容量监控以及恢复后对账清单；继续禁止自动反复重试迁移或手改 activation/数据库。
- **状态：** 回滚代码路径为 **已验证（静态）**；任何真实升级、失败恢复、RTO 或数据守恒均 **未执行**。

## 控制矩阵

| 领域 | 仓库接线（已验证） | 文档宣称/外部依赖（推断） | 缺失或未执行 |
|---|---|---|---|
| 触发 | main push 仅 prepare；main 上手动 dispatch + owner + `deploy=true` 才 activate（deploy workflow `:3-18,25-26,56-63`） | Environment 限 main | 无 PR 产品 CI；未查 branch protection/Environment |
| 权限 | workflow 顶层仅 `contents: read`（`:19-20`）；checkout 均 `persist-credentials:false`（入口 workflows `:60-65`，review `:63-71`） | 自托管 Runner/Owner/变量管理员可信 | 未查 Runner group、主机 ACL、token 审计 |
| 不受信代码边界 | `pull_request_target` review 仅 checkout base SHA 且 sparse checkout 维护源，job 限 owner、同仓 PR（review workflow `:3-5,25,63-82`） | review 脚本正确隔离 PR head | 未执行恶意 PR 测试；review status 写权限仍依赖脚本安全 |
| 第三方 action | `actions/checkout` 固定完整 SHA `11d596…`；`upload-artifact` 固定完整 SHA `ea165f…` | 固定 SHA 对应预期上游版本 | 未联网核验 action tag/来源；无依赖更新策略证据 |
| 秘密暴露 | GitHub token 仅传控制器；控制器 `pop` 后不传给产品命令（控制器 `:541-548,819-823`）；代理校验禁止 URL 用户名/密码并 mask（入口 workflows `:35-59`）；初始密码 0600 写私有文件（控制器 `:930-935`） | 配置变量内容和主机日志权限正确 | 未读 Secrets/vars/日志；无法证明无运行时泄漏 |
| 源完整性 | 两个入口和部署 workflow 有 path+SHA256 source 文件；本地实算分别匹配 `b8be4d…`、`4a9f35…`、`a3d814…`（各 source JSON `:1-4`） | `.harness/agent-platform/SOURCE.json` 是维护源供应清单 | code-review workflow 无同级 source JSON；未执行安装器一致性测试 |
| 产品门禁 | `Makefile:8-22` 的 `verify` 委托 Go verifier，门禁含 evidence/history、format、module verify、vet、Node、race tests、build、latency、真实 Go browser；prepare 对精确 SHA 调用（控制器 `:568-619`） | 依赖固定主机工具链和锁文件 | 本审查未运行；合入前未接线 |
| 制品关联 | 版本用完整 SHA 注入，二进制 SHA-256、契约、previous SHA 与 plan/deployed 绑定（控制器 `:298-305,618-629,673-681,837-859`） | 私有根持久、安全、可备份 | 无 GitHub artifact/attestation/SBOM；未查实际记录 |
| 审批/目标 | 目标仅枚举 preview/validation，目录来自变量，不接受任意输入；validation 传 protected root（deploy workflow `:13-18,35-53,67-83`） | local environments 配置正确，loopback/launchd 安装正确 | 单 Owner 审批；未验证实际 URL/根隔离 |
| 回滚/恢复 | 持久 activation 阶段、备份/密钥/二进制摘要、failed-data 隔离、旧版 restore/health 核对、未知状态故障关闭（控制器 `:706-816,878-955`） | 主机存储耐久、备份容量充足 | 未执行真实故障；无异地主机恢复、RTO/告警证据 |

## 已有安全与可追溯设计

- 工作流权限显式最小化；checkout 不保存凭据；第三方 actions 使用完整 commit SHA，而非可移动 tag。
- `pull_request_target` 没有直接 checkout PR head，且限定仓库 Owner、同仓 head 和受信 base 中的 sparse 维护源，降低 fork/PR 代码借 token 执行的风险。
- 部署拒绝短 SHA、非当前远端 main、旧 plan、漂移二进制、漂移契约和不匹配 current 指针；重复部署只有记录与真实健康一致时才无副作用返回。
- prepare 与 activate 分离；main push 不自动切服，手动 deploy 每次重新 prepare 当前 main。目标为固定枚举，validation 显式保护 preview root。
- token 被从控制器环境移除；产品验证、构建和服务不应继承仓库 token。私有根使用 `umask 077`，初始密码单独 0600 保存；控制器限制产品输出并不把原始错误输出打印到 Actions（控制器 `:320-425,960-961`）。
- 验证日志每次使用唯一文件，不覆盖同 SHA 的失败历史；ready/deployed/activation 记录绑定 SHA 和摘要。失败恢复保存 failed-data 和部分恢复证据，不把不确定目录冒充可用数据。
- 文档准确区分维护源回归、受控替身、真实 Go 生命周期、Actions/launchd、页面旅程与真实供应商联调，没有把控制器单测或 fault fixture 冒充整条发布通过（维护源 README `:46-48,62-64,67-89`）。

## 建议归属与次序

1. **仓库基建 / P1：** 增加合入前产品 CI 和 required check；同时保留发布时精确 SHA 复验。
2. **通用平台 + 仓库基建 / P1：** 隔离部署 Runner，并建立外置控制器执行前的摘要/签名校验和配置漂移审计。
3. **仓库基建 + 工程工具 / P1：** 生成并留存脱敏、版本绑定的验证/部署/回滚 manifest；避免上传秘密、数据库和原始私有日志。
4. **仓库流程 + 通用平台 / P2：** 明确单人授权风险接受，条件允许时启用双人 Environment approval，并定期审计 branch/Environment/Runner 配置。
5. **仓库流程 / P2：** 为中断或 recovery incomplete 建立告警、责任人、RTO、容量/保留与恢复对账 runbook。

没有识别到仅凭当前静态配置即可定为 P0 的问题；这不代表运行环境没有 P0。关键的 GitHub 保护规则、Environment、Runner 主机和真实部署根均未访问。

## 本次实际检查

- **已验证：** 读取指定 workflow/source JSON、部署维护源、Makefile、部署文档、当前 PR/交付段落；使用本地 `git log`、`git merge-base --is-ancestor` 核对候选与三项前置；使用本地 `sha256sum` 核对三个已提供 source JSON 的 workflow 摘要。
- **未执行：** workflow 语法在线解析、Actions、GitHub API、平台 Connector、`make verify`、Python/Go 测试、构建、上传制品、Environment 审批、预览、部署、健康探测、备份或回滚演练。
- **限制：** `.github` 配置无法证明远端 required checks、Environment reviewers/branch policy、变量值、Runner group/主机权限和 artifact 实存状态；维护源代码只能证明设计存在，不能证明已安装版本、真实控制器摘要或运行结果。
