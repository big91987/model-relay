# 工程知识契约静态审查

## 1. 审查结论与边界

结论：三项前置成果均已合入本候选，具备审查前提。仓库已经具备较强的分层材料、Run 内需求/架构/验收追踪和诚实的证据边界；但项目级“当前入口”没有随产品演进收敛，至少四个时代的状态并存。最严重的问题是 README 的协议权威链接已经指向被局部功能覆盖的全局设计文件，工程人员无法按该链接取得所声明的协议契约。另有项目委托、路线图、Run 状态和全局 workflow 入口互相错位，易导致错误选取 AC、范围或发布状态。

本报告仅审查知识归档、导航、版本关系、冲突与可执行性；不复核三项产品改动的业务合同，也不判断产品 QA 是否通过。

### 候选与前置合入记录

| 项目 | 已验证事实 |
|---|---|
| 候选 SHA | `122c9abece28f562616483e991098d020d236d6d` |
| 工作树 | 分支 `agent-team/task-audit-knowledge-contracts-1`；审查开始时 `git status --short --branch` 仅显示分支标题，无修改项 |
| 严格 Bearer 认证 | integration commit `2d48ee08f12368e1b806fcc0975943e2eca0c8ee`，`git merge-base --is-ancestor … HEAD` 退出 0 |
| 单请求单次时钟推进 | integration commit `532e713acc998162c682b310c8f280e1651e6f95`，祖先检查退出 0 |
| 严格备份清单 | integration commit `122c9abece28f562616483e991098d020d236d6d`，即当前 HEAD，祖先检查退出 0 |

状态说明：全文“已验证”表示由本次文件或只读 Git 检查直接确认；“推断”表示由已验证材料关系得出的工程影响；“未执行”表示本任务按约束未运行构建、测试、服务或页面。审查读取了指定八份文档及其直接引用的当前 Run 入口，历史材料仅抽样索引和状态页，没有遍历日志。

## 2. 知识关系图

| 关系 | 仓库当前表达 | 审查判断 |
|---|---|---|
| 委托 | `docs/product-brief.md` | 项目级委托，但正文仍以 Issue #18/M2 为“最新” |
| 公共版本路线 | `docs/roadmap.md` | 自称唯一公共路线；顶部又承接 Issue #25/design-v0.2.0 |
| 基础版本契约 | `docs/releases/v0.1.0.md` | 明确定位为历史基础能力与持续回归契约 |
| 当前广义 AC | Run `4cd…/product-definition.md` → 冻结材料 `docs/prd.md`、`docs/acceptance.md` 等 | 关系定义清楚，但权威正文位于 `.workflow-input` 外部材料根，项目入口只做承接 |
| 当前设计 | Run `4cd…/architecture.md`（由 delivery 索引引用） | Run 内关系清楚；项目根 `docs/workflow/design.md` 并非当前总体设计 |
| 工程规则 | `AGENTS.md` | Go、验证、秘密、真实服务与发布边界清楚 |
| 当前操作入口 | README 首屏链接 Run `4cd…`；但后续段落还并列 Issue #18、Issue #28、Issue #63 | 没有一个机器可判定的当前候选/状态入口 |
| 全局 workflow 三件套 | `task.md` → `requirements.md` → `design.md` | 内部一致，但只对应历史 Issue #9“运行版本”局部功能 |

## 3. 发现

### K-01 — README 的协议权威链接已失真

- 严重度/状态：**P0 / 已验证**。
- 证据：`README.md:95-103` 声明协议清单、来源规则和错误契约以 `docs/workflow/design.md` 第 4–6 节为准；但该文件标题是“管理台运行服务版本”，并明确其基线仅为运行版本 AC（`docs/workflow/design.md:1-7`）。其第 4 节是版本展示的存储/5 秒等待/错误边界（`:40-47`），第 5 节是该局部功能的测试契约（`:49-66`），第 6 节是版本展示推荐决策（`:68-74`），均不是 README 所述的 Chat Completions 协议、转发或错误契约。
- 影响（推断）：按 README 实施或审查协议变更时会取得错误权威材料；未知字段、多模态、SSE、重试及安全输出上限等要求只有 README 自身陈述，失去声明的设计依据，属于不可执行导航。
- 建议归属：**仓库流程**。建议把链接改到当前、版本固定的协议/架构章节，并增加链接目标章节存在性与语义标签检查；不要继续用可被局部任务覆盖的全局文件名承担长期协议权威。

### K-02 — 项目级委托未随“当前范围”更新

- 严重度/状态：**P1 / 已验证**。
- 证据：`docs/product-brief.md:9` 仍称 Issue #18/M2 为“最新”，状态为 Delivered 但未合并部署；同文件 `:34` 也把 Issue #18 描述为“本轮”。相对地，`docs/roadmap.md:5-7` 已声明 2026-10-07 当前委托为 Issue #25 / Run `4cd…`，完整多租户 AC01–AC18/H01–H08；README 首屏当前入口同样指向该 Run（`README.md:7-22`）。
- 影响（推断）：从“产品委托”进入的需求分析会把完整多租户范围误判为多上游 M2，且会携带错误的发布状态和 PR 责任。
- 建议归属：**仓库流程**。项目级 brief 应只保留稳定委托与明确的 current-run 指针；历史切片状态迁入带日期的历史索引，或在顶部提供 superseded 标记。

### K-03 — 全局 workflow 入口名义通用、实际只属于历史局部任务

- 严重度/状态：**P1 / 已验证**。
- 证据：`docs/workflow/task.md:1-9`、`requirements.md:1-11`、`design.md:1-7` 都对应 Issue #9“管理台运行服务版本”。requirements 甚至声明它“替换此前 v0.1 整体产品需求基线”（`requirements.md:7`），design 又声明替换历史 v0.1 总体设计（`design.md:5`）。与此同时，当前广义需求入口明确是 Run `4cd…/product-definition.md`，其 `:6-24` 将冻结包中的 PRD、acceptance、interactions、decisions 和 manifest 定义为事实源；README 当前范围也指向该 Run（`README.md:19-20`）。
- 影响（推断）：工具或新人按常见路径 `docs/workflow/{requirements,design,task}.md` 自动发现“当前需求/设计”时，会选择一个已经完成的局部功能，而不是当前产品契约；K-01 已证明这种歧义会外溢到实际运维文档。
- 建议归属：**仓库基建**。建议建立稳定的 `docs/current/` 或清单文件，显式给出 current delegation、requirements、AC、architecture、candidate 与 evidence 指针；全局 workflow 文件改成索引或带作用域/失效元数据，不再被任务原位覆盖。

### K-04 — 路线图同时保留互斥范围，缺少结构化适用性

- 严重度/状态：**P1 / 已验证**。
- 证据：`docs/roadmap.md:7` 说明 Issue #25 已明确覆盖完整多租户与成本分账，并声明后文旧限制不约束当前范围；但 `:11` 仍把用户约束写为单管理员，`:19-22` 又把多租户计费列为后续/默认不承诺。`docs/releases/v0.1.0.md:5` 使用类似文字解释下述单管理员/唯一上游只是历史，而 `:15-17` 随即继续以现在时写“只做一个管理员、一个上游”。
- 影响（推断）：人可以通过上下文勉强消歧，但自动抽取和局部引用极易把历史限制当当前规则；范围审查结果依赖读者是否看到了顶部免责声明。
- 建议归属：**仓库流程**。建议把历史版本内容放入不可变版本文档并用“适用于 v0.1.0”字段标注；当前路线只保留有效约束。对被替代规则建立显式 `supersedes` 关系，而非在同一正文中用自然语言否定后文。

### K-05 — “当前实现与验证状态”链接本身是过期快照

- 严重度/状态：**P1 / 已验证**。
- 证据：README 将 Run `4cd…/delivery/README.md` 标为“当前实现与验证状态”（`README.md:20`）；该索引却记录候选 HEAD `66f1328…` 加未提交修复、未合并、下一步 publish/PR27（Run delivery `README.md:7-15`）。本次候选 HEAD 已是 `122c9ab…`，并额外包含三项实验集成提交。该状态页也明确说仍 Verifying、不是 Delivered（`:9-14`）。
- 影响（推断）：入口没有“截至 SHA/是否仍当前”的校验，读者可能把旧候选状态用于新候选发布或验收决策；三项新修复也没有被项目知识导航关联。
- 建议归属：**工程工具**。建议在当前入口生成并校验 `candidate_sha`、`source_run`、`generated_at`、`superseded_by`，当 HEAD 不匹配时在文档检查中失败或醒目标红。

### K-06 — 当前广义需求依赖工作流输入目录，长期可用性契约不充分

- 严重度/状态：**P2 / 已验证 + 推断**。
- 证据：Run `4cd…/product-definition.md:8-24` 把正式业务规则放在 `.workflow-input/.../files/docs/{prd,acceptance,interactions,decisions}.md`，并记录摘要；该 Run 文件只做承接和补充。项目级路线和 README 链到承接文件，而没有直接、仓库内稳定的 AC 正文入口。
- 影响（推断）：如果工作流输入未随 clone、归档或发布保留，项目仓库中的 current requirements 无法独立执行；摘要能证明材料身份，不能保证未来可取用。本文未检查外部归档生命周期，因此不判定材料已经丢失。
- 建议归属：**通用平台**。平台应为冻结输入提供内容寻址、只读、可长期解析的仓库内或制品 URI，并在 Run 索引中记录可用性检查；仓库则保留最小的正式 AC 映射，避免只有本机路径。

### K-07 — 根 README 聚合运行说明、历史状态和多个当前增强，导航负荷过高

- 严重度/状态：**P2 / 已验证**。
- 证据：README 同时把多租户 Run `4cd…` 称为当前目标（`:7-22`），在历史段落把 Issue #18 称为当前（`:24-32`），在验证段落又指向 Issue #28（`:120-148`），最后新增 Issue #63 的冻结设计与实验（`:150-158`）。这些段落没有统一的状态表或优先级规则。
- 影响（推断）：当前入口需要读者手工比较日期、Issue 和 Run；版本、AC、设计、工程规则与交付状态无法一跳闭合。
- 建议归属：**仓库基建**。README 首屏应给出唯一 current matrix（范围、AC、设计、候选、QA、发布），历史和专题增强移至按日期/版本索引；运行手册与项目状态分文件维护。

## 4. 有证据的正面能力

1. **工程边界明确（已验证）**：`AGENTS.md:3-17` 清楚规定原创实现、Go 后端、真实 Go 服务验证、秘密与数据保护、受控上游不等于供应商联调、Connector/Git/部署责任；规则可直接转成审查清单。
2. **版本层次已有雏形（已验证）**：路线图明确“公共阶段目标 → 基础版本契约 → 最新 Issue/Run PRD”的关系（`docs/roadmap.md:39-46`）；v0.1.0 也明确自己只证明基础契约、不证明后续增强（`docs/releases/v0.1.0.md:1-11`）。问题在于入口维护没有完全落实该模型，而非完全没有模型。
3. **冻结需求可追溯（已验证）**：Run `4cd…/product-definition.md:6-24` 给出材料根、文件职责和 SHA256，并说明冲突优先级、原型非事实源及完整性不等于产品通过。
4. **状态措辞克制（已验证）**：当前 Run delivery 把 Ready for Review、Verifying、Go with known issues、未合并/未发布/Not Run 分开（Run delivery `README.md:7-15,31-37`），没有把文档或受控测试冒充正式发布。
5. **局部任务 AC/设计映射完整（已验证）**：Issue #9 的 requirements AC-01～08 与 design 测试表逐项对应（`docs/workflow/requirements.md:34-47`；`docs/workflow/design.md:49-66`），且设计明确未运行测试、不得宣称 AC 通过（`design.md:66,78-84`）。其缺陷是作用域和“当前性”标识，而非内部可执行性。
6. **直接引用目标存在（已验证）**：本次抽查的 README、路线图、brief、release 直接引用的 Run product-definition、delivery、validation、architecture、Issue #28 检查及 Issue #63 设计/实验文件均存在；未发现抽样范围内的物理断链。

## 5. 优先处置建议（均未实现）

1. **P0 / 仓库流程**：立即修复 README 协议契约链接，目标必须是当前且不会被局部任务覆盖的版本化架构/协议文档。
2. **P1 / 仓库基建**：增加一个结构化 current manifest，至少包含 delegation、release、requirements、AC、architecture、engineering rules、candidate SHA、verification status、release status 及 supersedes。
3. **P1 / 仓库流程**：将 `docs/workflow/task.md`、`requirements.md`、`design.md` 改为 current manifest 的稳定索引，或迁入按 Run/Issue 命名的不可变目录；禁止局部任务覆盖全局语义文件。
4. **P1 / 工程工具**：在文档门禁中检查 Markdown 目标、章节锚点、目标文档 scope/status，以及状态页 candidate SHA 与当前候选关系；仅检查“文件存在”不足以发现 K-01。
5. **P2 / 通用平台**：为 `.workflow-input` 冻结包提供内容寻址的长期只读引用与可用性验证，避免正式 AC 只依赖执行机路径。

## 6. 未执行与局限

- **未执行**：构建、单元/集成测试、`make verify`、浏览器、服务启动、供应商调用、产品业务合同复核。
- **未执行**：外部 GitHub/Deployment/PR 链接真实性复验；报告只评价仓库如何表达这些状态。
- **未遍历**：全部历史日志、全部 Run 证据和三项前置成果的产品实现；仅抽样了建立权威/状态关系所需的索引。
- 因此本报告完成只表示静态工程知识审查已交付，不表示产品 QA、发布或供应商验证通过。
