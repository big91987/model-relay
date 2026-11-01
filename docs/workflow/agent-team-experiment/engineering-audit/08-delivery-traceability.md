# 交付追溯与指标可收集性静态审查

## 1. 结论与范围

结论：仓库已经具备“需求/AC → 设计 → 验证 → QA → 交付”的文档骨架、按 AC 的验收矩阵、失败保留、源码摘要和部分运行时长等正面能力；但这些能力只在若干 run/harness 样本中较完整，尚未形成跨任务一致、机器可聚合的事实模型。当前顶层 workflow 文件把 Issue #9 当前记录与 v0.1 历史长篇混写，`.trellis/` 没有任务实例元数据，Git 提交也没有稳定的 Issue/AC/run/receipt 标识。本轮三项前置虽可由 Git 证明已合入，却无法仅从仓库离线重建它们的 Issue → AC → 验收 → 交付全链。因此，单次门禁耗时、已明确标注的失败尝试和部分 AC 覆盖可以按样本采集；端到端周期、排队、首次通过率、跨任务返工率、缺陷关闭时长及验证成本不能可靠汇总。

审查状态为 **completed（静态审查完成）**，不代表产品验收、供应商验证或发布通过。全程只读文件和 Git；未调用外部平台，未运行旧流水线、产品测试或大日志全量扫描。

抽样范围：顶层 `task.md`、`requirements.md`、`design.md`、`qa.md`、`pr.md`、`delivery.md`；run `637705…` 与 `cf691e…` 的总览/PR/矩阵/缺陷索引代表样本；harness `run-91bc…` 与 `run-e11…` 的 report/qa 代表样本；`.trellis/config.yaml`、`.trellis/workflow.md` 及目录结构；当前 Git 图和最近提交。没有读取所有大型 `.log`，仅使用报告中已记录的大小、摘要、退出码和时间。

证据状态：**已验证**指本轮直接读取静态文件或只读 Git；**推断**指由样本结构推导、未对全仓逐项穷举；**未执行**指外部平台、流水线或产品行为未复核。

## 2. 当前候选与 Git 可见状态

三项依赖在任务交接元数据中均为 `merged`；本轮在仓库中以只读 `git log/show` 验证了候选与合入对象：

| 前置项 | 候选提交 | 合入提交 | 当前 Git 可见关系 | 状态 |
|---|---|---|---|---|
| strict-bearer-auth | `67f464494f4d59be8a20ef99d9b15175ecd00e78` | `2d48ee08f12368e1b806fcc0975943e2eca0c8ee` | merge 第二父为候选；第一父 `521751f…` | 已验证 |
| single-clock-advance | `318639d819165f3309f123133fdaf6229003270c` | `532e713acc998162c682b310c8f280e1651e6f95` | merge 第二父为候选；第一父为上一合入 `2d48ee0…` | 已验证 |
| strict-backup-manifest | `7dc57094557b8883c130a41ed306d1223c0d92f8` | `122c9abece28f562616483e991098d020d236d6d` | merge 第二父为候选；第一父为上一合入 `532e713…`；当前 HEAD 即该合入 | 已验证 |

三个 merge 的 author/commit 时间均为 `2026-11-01T12:21:51+08:00`，候选提交时间分别为 12:15:03、12:15:01、12:14:58。Git 能证明拓扑和本地提交时间，不能证明外部 PR、审批、远端可见性、真实排队开始或部署。当前分支为 `agent-team/task-audit-delivery-traceability-1`，审查前工作树无改动；`origin/agent-team/experiment` 仍指 `521751f…`，故只能称本地三项已集成，不把远端发布写成完成。

## 3. 实际追溯矩阵

| 链路 | 正面证据 | 断点/歧义 | 评价 |
|---|---|---|---|
| Issue → 任务入口 | 顶层任务明确写 Issue #9、用户澄清与范围（`docs/workflow/task.md:5-9,17-30`）；run 样本标题直接写 #34，PR 正文有 `Refs #34`（`runs/637705…/README.md:1-10`；`pr.md:1-7`）。 | Issue 内容、创建/更新时间、状态和身份依赖外部平台；顶层还记录历史 404/审批拒绝（`task.md:38`），仓库无法独立验证原 Issue。三项前置提交主题无 Issue ID。 | 局部可追溯 |
| 任务 → 需求/AC | `requirements.md` 明确替换历史基线并定义 AC-01～08（`requirements.md:7-11,34-47`）；任务验收段与其一致（`task.md:26-34`）。 | 三项前置只有各自工程记录，没有统一 requirements/AC ID；`.trellis/workflow.md:42-76` 规定任务目录/`task.json` 生命周期，但当前 `.trellis/` 实际不存在 `tasks/` 或 `archive/` 实例。 | Issue #9 完整；本轮前置断链 |
| AC → 设计 | `design.md` 显式以 AC-01～08 为基线并列模块、事实链和失败边界（`design.md:3-7,21-45`）。 | 设计引用 AC 是段落级，缺机器可校验的逐 AC 设计映射；三项前置无同一编号体系。 | 人工可查，难聚合 |
| 设计 → 候选 | `delivery.md` 将 AC-01～07 映射到实现/测试（`delivery.md:15-22`），并用工作树摘要避免“未提交实现=HEAD”（`:24-35`）；`pr.md:24-36` 列变更范围。 | Issue #9 的受测对象是未提交工作树摘要，候选提交未在顶层交付文档回填；当前 Git HEAD 已是其他三项集成，顶层“当前”已过期。 | 有版本意识，交付闭环未回填 |
| 候选 → 验证 | QA 记录同版、门禁、逐 AC 结果和失败保留（`qa.md:7-13,15-31,35-48`）；PR 记录宿主 `make verify` exit 0 和摘要边界（`pr.md:44-66`）。现代 run 样本记录精确起止、653.728 秒、日志 bytes/SHA（`runs/637705…/README.md:22-34`）。 | 顶层成功回执主要是平台摘录，仓库不总含结构化 state；研发与宿主结果混在同一长文件。三项前置虽各自文档写 receipt，但无仓库级索引将 receipt 绑定 integration commit。 | 单任务较强，跨任务较弱 |
| 验证 → 独立验收 | QA 的 AC 矩阵区分 Pass/Not Run（`qa.md:23-31`），明确 QA 自己未重跑完整门禁（`:35-39`）。harness QA 也区分固定 tests5 与 QA 定向复验（`docs/harness/run-91bc74a22a9f/qa.md:26-58`）。 | “QA Go”与历史 No-Go 共存于同一顶层文件（`qa.md:58-60,168-170,172-190`）；消费者若只搜 Go/No-Go 会双计或误判。 | 证据诚实，但当前/历史隔离不足 |
| 验收 → PR/交付 | `pr.md:9-13,38-56` 汇总基线、阶段和验证；run 样本记录 publish/PR head/未部署边界（`runs/637705…/README.md:24-34,62-66`；其 `pr.md:17-27,31-37`）。 | 顶层 `pr.md:13,40` 明说草稿 PR 尚未创建，当前仓库又已存在后续提交；缺最终 PR URL、候选 SHA、合入 SHA、交付时间的统一回填记录。 | 文案边界好，终态字段缺失 |
| 缺陷 → 关闭 | 代表 run 有独立 `acceptance/defects.md`，QA 也以 QA-GAP-01/02 记录发现、整改及关闭（`qa.md:129-158,172-194`）。 | 缺陷没有统一 ID 生命周期事件、opened/closed timestamp、owner、关闭 commit；“未发现缺陷”“覆盖缺口”“环境失败”主要是自然语言。 | 可个案审计，不可可靠统计 |

## 4. 主要发现与建议

### F1 — P0：本轮三项前置缺少仓库内端到端追溯记录

- 状态：已验证。影响：Git 只能证明三个候选先后被 merge 到本地实验分支；无法离线回答各自来源 Issue、明确 AC、验收主体、receipt 与 integration commit 是否同版，也无法生成完整 Issue → 交付矩阵。
- 依据：最近 Git 主题仅为 `agent-team: …`，无 Issue/AC/run 标识；三项合入 merge 主题仅含候选 SHA。与此相对，旧 run 会在总览中显式绑定 Issue、Run、PR、HEAD 和 tests/QA（`runs/637705…/README.md:1-10,22-34`）。
- 建议归属：**仓库流程**。每个任务生成一个小型、不可变的 `trace.json`/Markdown front matter，至少含 `issue_ids`、`requirement_ids/ac_ids`、`candidate_sha`、`integration_sha`、`verification_receipts`、`qa_decision`、`pr`、`delivered_at`；合入后由控制器回填最终 SHA，禁止只留在平台任务状态。

### F2 — P1：当前文档与历史记录同文件混写，状态查询容易误判

- 状态：已验证。影响：`qa.md` 开头是 Issue #9 Go，后半又保留 v0.1 No-Go 和复验 Go；`delivery.md` 开头是 Issue #9，但后续包含 R1/R2/R3 与旧 v0.1 实现；自动抓取“结论/失败/通过”会重复计数并混淆当前版本。
- 依据：`qa.md:4-5,58-60,168-190`；`delivery.md:3-13,43-81,83-110`；基线自己也承认旧材料不是当前依据（`requirements.md:7`、`design.md:5`、`pr.md:11`）。
- 建议归属：**仓库流程**。顶层文件只保留当前任务指针和 latest decision；历史按 `docs/workflow/runs/<run-id>/` 冻结，增加 `supersedes`/`superseded_by`、`effective_from`、`subject_sha`。历史段不得继续使用“当前结论”字段。

### F3 — P1：运行证据丰富但没有统一目录索引与结构化状态

- 状态：已验证（代表样本），全仓完备率为推断。影响：`runs/` 下有 README、矩阵、缺陷、state、日志等多种形态，但根目录无统一 README/schema/index；同名语义分散，统计必须为每个 run 写特例。
- 依据：`runs/637705…/README.md:14-34` 提供状态、版本、时间和日志摘要；`cf691e…/acceptance/acceptance-matrix.md` 与 `defects.md` 提供矩阵/缺陷；另一些 run 仅有 README 或 delivery 子目录。`.trellis/workflow.md:42-76` 定义另一套任务生命周期，当前又无实例目录。
- 建议归属：**仓库基建**。增加 `docs/workflow/runs/README.md` 和 versioned schema；每个 run 必须有 `run.json`、`candidates.json`、`verification/*.json`、`acceptance/matrix.csv|json`、`defects.json`，Markdown 只作解释。建立仓库内索引，校验 ID 唯一和引用存在。

### F4 — P1：时间戳、身份和版本口径不统一

- 状态：已验证。影响：Git author/committer 时间可用，但文档混用上海时间、UTC、仅日期、seq/attempt；执行者有 Connector、runner、QA、Builder、Owner 代理等自然语言身份。无法可靠计算等待、排队、人工确认时长或按责任方归因。
- 依据：run 样本同时记录上海/UTC及 tests22 区间（`runs/637705…/README.md:24,30-32`）；harness 以 UTC 和节点/执行号记录（`docs/harness/run-91bc74a22a9f/qa.md:3,26-35`）；顶层 Issue #9 仅日期标题（`delivery.md:3`）。同一 harness 明确 QA 未重跑全套（`qa.md:5,50-58`），说明“谁执行”不可从文件路径推定。
- 建议归属：**通用平台**。所有事件写 RFC3339 UTC `started_at/finished_at/observed_at`、稳定 `actor_id/role`、`attempt_id`、`subject_sha/source_digest`；平台代理确认与用户确认用不同 actor type。

### F5 — P1：失败分类与缺陷生命周期仍是自然语言，首次通过率和缺陷关闭率会失真

- 状态：已验证。影响：环境限制、fixture 错误、产品断言、门禁配置、外部发布校验失败都被记为“失败”，但没有统一 taxonomy；若按 exit code 统计，首次通过率会混入未执行断言和嵌套预期失败。缺陷也缺 opened/closed 事件。
- 依据：顶层 QA 区分 403 fixture 与监听限制（`qa.md:42-48`）；harness QA 记录首次组合命令被后续 Node 成功掩盖、不得计为 Go 通过（`docs/harness/run-91bc74a22a9f/qa.md:50-58`）；其 report 还保留 publish 字段校验 exit1、未进入 Git 写入（`docs/harness/run-91bc74a22a9f/report.md:33-37`）。
- 建议归属：**工程工具**。receipt 增加 `outcome={pass,fail,blocked,not_run,cancelled}`、`failure_class={product,test_fixture,environment,tooling,policy,external}`、`stage`、`top_level`、`retry_of`；缺陷用稳定 ID 和 open/fix/verify/close 四事件。

### F6 — P2：源码摘要和原始日志完整性是成熟正面能力，应标准化复用

- 状态：已验证。影响：现有样本能避免把旧绿回执套到新代码，并能证明日志未截断，这是后续指标可信度的重要基础。
- 依据：Issue #9 明确 HEAD 不能代表未提交实现并记录 33 文件摘要（`pr.md:58-66`）；run 样本记录 158 文件摘要与日志 bytes/SHA（`runs/637705…/README.md:24-32`）；harness QA 记录 160 文件摘要、原始/runner 日志 SHA、`truncated=false` 和 candidate 清理（`docs/harness/run-91bc74a22a9f/qa.md:9-12,26-35`）。
- 建议归属：**工程工具**。把 `subject_sha/source_digest`、digest manifest、log SHA/bytes/truncated、toolchain、cleanup 状态固化到统一 receipt；禁止只在 Markdown 复制。

### F7 — P2：Git 提交可观察关联不稳定

- 状态：已验证。影响：旧 Git 提交有 `(#73)`、`Merge pull request #70` 等平台关联，三项前置提交却只有中文任务标题；跨样本无法统一从提交反查 Issue/PR/run。
- 依据：当前 `git log` 中 `521751f…` 含 `(#73)`、`a1cbbb8…` 含 PR #70，而 `67f4644…`、`318639d…`、`7dc5709…` 无 ID。
- 建议归属：**通用平台**。提交 trailer 使用稳定键：`Issue: #…`、`Run-ID:`、`Task-ID:`、`AC:`、`Receipt:`；merge 提交增加 `Candidate-SHA:`。标题可读性不能替代 trailer。

## 5. 指标可收集性与计算口径

| 指标 | 当前可收集性 | 可用口径/公式 | 关键限制 |
|---|---|---|---|
| 单次验证持续时间 | **可采（部分 run）** | `gate.finished_at - gate.started_at`；如 state 明列准备/清理，另报 wall、gate、phase 三层。样本：tests22 653.728 秒（`runs/637705…/README.md:30`）；tests5 gate 980.081 秒、总计 995.135 秒（`docs/harness/run-91bc74a22a9f/qa.md:32-35`）。 | 不能从 Markdown 日期或 Git 时间推算；冷/热缓存口径必须分组，harness 明说不声称模块/浏览器全冷（`:34`）。 |
| 阶段耗时/成本代理 | **可采（局部）** | 对有 phase timing 的 receipt 求和；同时记录 CPU/内存/工具链/样本数。harness 给出 migration/legacy/routing/console/full-console 分段（`:37-48`）。 | “验证成本”只能先用 wall time、runner 次数、制品字节作代理；没有 CPU-seconds、计费、人工工时，不能换算货币成本。 |
| 尝试次数/返工次数 | **可采（单 run）** | 同一 `subject_sha + gate + retry chain` 中，首次终态前的顶层 attempts 数；返工为候选内容变化后再次验证的次数。 | seq/attempt 命名不统一；嵌套预期失败不能计顶层失败；历史日志可能只保留摘录。跨 run 不可靠。 |
| 首次通过率 | **不可可靠汇总** | 未来：`首次合格顶层验证为 pass 的候选数 / 至少启动一次合格验证的候选数`；`blocked/not_run/cancelled` 排除并另报。 | 当前失败 taxonomy、candidate ID、top-level 标志缺失；监听限制、fixture 错误与产品失败混在自然语言，样本选择又偏向成功交付。 |
| 失败类别分布 | **人工可采样，不可自动汇总** | 以 F5 taxonomy 对每个顶层失败唯一主分类，允许 secondary causes；分母为终态 fail attempts。 | 现有报告虽诚实区分原因，但无结构化字段；同一 attempt 可能多个失败，容易重复计数。 |
| 端到端周期时间 | **不可采** | 未来：`delivered_at - issue_ready_at`；另报 `candidate_ready→verified`、`verified→qa_decision`、`qa→merged`。 | 缺 Issue 事件、ready 定义、最终交付时间；Git author/commit 时间不等价于开始/完成。外部平台数据未入库。 |
| 排队时间 | **不可采** | 未来：`execution_started_at - queued_at`，按 runner/pool 分组。 | 仓库没有 queued/claimed/started 事件；seq 间隔无法区分排队、人工等待和返工。 |
| 缺陷发现/关闭时间 | **不可可靠汇总** | 未来：`verified_closed_at - opened_at`；需 defect ID、severity、cause、fix SHA、verification receipt。 | 当前 defects/QA-GAP 多为 Markdown 状态，缺时间和统一关闭事件；“覆盖缺口”也不应混同产品缺陷。 |
| AC 覆盖率 | **单任务可采** | `有同版终态证据的适用 AC / 全部适用 AC`，Pass/Fail/Blocked/Not Run 分列，N/A 必须有理由。 | 不得只统计表格行或用门禁 exit0 代替 AC；当前 AC ID 只在部分任务存在，历史/当前混写会双计。 |
| 交付吞吐量 | **不可可靠汇总** | 未来按 integration/delivery event 的唯一 task 计数，不按 commit 数。 | squash/merge/文档补正、多提交一任务会造成重复；本地 merged 不等于远端 PR/发布。 |

建议首批只发布三项低争议指标：验证 gate 持续时间、按结构化分类的顶层 attempt 终态、AC 状态分布；在事件字段补齐前，不发布团队周期、排队、首次通过率或缺陷 MTTR 排名。

## 6. 样本偏差与口径限制

- **成功者偏差**：纳入仓库的 report/qa 多为可交付或被长期保留的复杂案例，失败后未形成报告的任务不可见；不能代表总体成功率。
- **复杂任务偏差**：抽样 run/harness 都有大量浏览器/冷门禁/归档证据，耗时远高于小改动；不可直接作为全项目基线。
- **幸存证据偏差**：大日志未全扫，本报告依赖其索引、摘要和 report 声明；验证了记录存在与自洽，不重新证明日志内容或产品行为。
- **时间口径差异**：UTC、Asia/Shanghai、仅日期、Git author/committer time 并存；同秒连续 merge 只能表示提交记录，不能说明无排队。
- **身份差异**：用户、Owner、Builder 代理、Connector、runner、QA 的行为边界不同；文件作者或 Git author 不能自动当执行者/批准者。
- **版本差异**：未提交工作树用 digest，发布后用 commit；若不同时保存两者及对应关系，旧回执可能误配新提交。Issue #9 文档已明确此风险（`pr.md:58-64`）。
- **外部平台缺失**：Issue/PR 评论、审核、Actions、队列、远端 branch、合并和部署均未查询；仓库文本中的外部状态是历史声明，不是本轮在线确认。
- **当前性限制**：顶层 workflow 文档仍描述 Issue #9，而当前 HEAD 是 2026-11-01 三项 agent-team 合入；“顶层=当前任务”的假设已经失效。

## 7. 最小落地方案

1. **P0 / 仓库流程**：为每次任务合入生成不可变 trace manifest，并立即为本轮三项前置补录 Issue/AC（若无外部 Issue，明确 `issue: null` 和授权任务 ID）、候选、receipt、QA、integration SHA。
2. **P1 / 仓库基建**：定义 `runs/v1` schema 和根索引；将 current pointer 与 immutable history 分开，CI 校验引用、时间格式、SHA 和状态枚举。
3. **P1 / 工程工具**：统一 receipt、失败 taxonomy、retry chain、subject digest、日志完整性和 phase timing；从工具直接产出 JSON，Markdown 引用它。
4. **P1 / 通用平台**：把 issue/queue/actor/PR/merge/delivery 事件以稳定 ID 和 UTC 时间写回仓库；提交 trailer 建立离线反查。
5. **P2 / 仓库流程**：把 AC matrix 与 defect lifecycle 结构化；QA 报告只消费同版 manifest，旧结论归档后不再保留“当前”措辞。

完成这些前，仓库适合做单次交付审计和证据核验，不适合据此评价团队效率或建立跨任务绩效指标。
