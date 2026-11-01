# QA 计划实质门禁评审

## 1. 门禁结论

**最终结论：`BLOCKED`。后续 QA 执行不得开始。**

`qa-contract-plan` 正确识别了候选中的两项 Bearer 协议冲突，也覆盖了时钟与备份的大部分必要风险，但计划仍有三项会造成误导性通过或证据不可交付的实质缺口：时钟核心观测未约束 SQLite 连接且不能可靠区分外层与内部推进；最终安装失败允许在无法注入时标 `Not Run`，与恢复失败原子性的不可变断言冲突；原始证据落点与控制器最新私有产物边界相冲突。计划必须先返工并再次接受 readiness review，执行节点不得临时改断言、证据位置或范围。

本评审只做静态复核和文档交付，没有运行专项测试或 `make verify`；没有修改产品、测试或原计划。

## 2. 候选与审阅材料

- 本评审候选 commit：`68005200c4187089065e70ea14a7a9e57e40db06`；开始时 `git status --short` 无输出。
- 被评审计划：`docs/workflow/agent-team-experiment/qa/01-contract-plan.md`，其产出提交 `d42266af1194a6540809b43835d781f129cd06cf`，集成提交即本评审候选。计划记录的产品候选 `122c9abece28f562616483e991098d020d236d6d` 是三项实现合入后的产品树；`68005200...` 仅再合入计划文档，产品代码未随之改变。执行阶段仍须按计划重新冻结当时 HEAD，不能复用任一旧回执。
- 正式项目材料：`AGENTS.md`、`docs/product-brief.md`、`docs/roadmap.md`、`docs/releases/v0.1.0.md`、`docs/workflow/runs/4cd62ced6a4f98f50e3cca6f7596eb06/product-definition.md`。
- 三项差异说明：`strict-bearer-auth.md`、`single-clock-advance.md`、`strict-backup-manifest.md`；仅作为实施意图与历史回执，不作为合同正确性的自证。
- 候选代码与回归：`server.go`、`console_keys.go`、`identity.go`、`console_api.go`、`console_clock.go`、`accounting.go`、`routing.go`、`backup.go` 及三个新增 `*_regression_test.go`。
- 官方依据：RFC 9110 §11.1（authentication scheme 大小写不敏感及通用凭据语法）、RFC 6750 §2.1（`Bearer` + `1*SP` + token）、RFC 5234 §2.3（ABNF 字符串字面量默认大小写不敏感）、SQLite `CREATE TRIGGER` 文档 §7（TEMP trigger 的连接可见性边界）。
- 控制器消息 seq 29：完整原始日志、截图只保存在项目 Git 忽略的 `test-results` 或运行私有目录；仓库 `qa/evidence` 只允许脱敏白名单摘要，不提交完整日志、截图、数据库、密钥、工作区副本或原生会话历史。

## 3. “合同 → 行为 → 断言 → 证据”逐项结论

### 3.1 Bearer：合同正确，候选实现阻塞，计划能暴露错误

正式合同要求 authentication scheme 按 ASCII 大小写不敏感；RFC 6750 的分隔是 `1*SP`，即一个或多个 SP。项目可以继续把 credential 本体收紧为 `mr_` 固定格式，并拒绝 HTAB、前导/尾随空白、空值、附加内容、错误生命周期或授权范围，但不能把标准 scheme 改成大小写敏感，也不能拒绝合法的多个 SP。

候选 `controlledAppID` 使用 `strings.CutPrefix(header, "Bearer ")`，生成入口另有精确 `"Bearer mr_"` 检查；新增回归明确把小写 scheme 与双 SP 断言为 401。实现和回归因此与正式协议冲突。计划 A02/A03 没有固化该错误：它明确要求大小写变体及双/多 SP 与规范形式得到相同 credential 校验结果，并把当前候选的拒绝视为预期 Fail/NO-GO。A01、A04、A05 还覆盖正常 credential、生命周期、授权隔离、并发和上游零命中。

结论：计划的 Bearer 合同与断言 **通过 readiness**；候选本身存在 `Critical` 协议缺陷 B-AUTH-01/B-AUTH-02，执行若发生只能 Fail，不能因现有回归绿色而放行。

### 3.2 时钟：覆盖主题齐全，但核心观测不可可靠复验

正式合同是 Schema 3 每个非 `/healthz` 外部请求在统一入口进行一次持久化外层推进；必要的 `admit`、逐 attempt 和邀请事务推进另计；高水位不回拨，写失败 fail-closed，重启和备份恢复连续。代码在 `ServeHTTP` 持 `server.mu` 调用一次 `advanceClock`，排队分支另做非持久 `wallTime` 采样；管理和 identity 的旧调用点已删除。现回归只数时钟函数调用，确实不足以证明持久 UPDATE 次数。

计划 C01–C07 覆盖单请求、内部事务、排队并发、高水位、防回拨、begin/read/update/commit 失败、重启/恢复和旧 schema，主题完整；但证据设计仍有两项歧义：

1. C01 写“TEMP/测试 trigger”，却没有要求 `SetMaxOpenConns(1)`、固定 `*sql.Conn`，或改用数据库 schema 中的测试审计表并在结束后清理。SQLite TEMP trigger 只对创建它的连接生效，而 `database/sql` 可让被测请求使用池中另一连接；漏计可把实际多次 UPDATE 错判为一次。
2. C02 要分别断言“外层一次、admit 一次、每个 attempt 一次”，但未定义如何给相同 `console_clock` UPDATE 标注来源、如何在夹具初始化及辅助 HTTP 请求后清零，以及怎样把一次业务请求产生的审计行与其 attempt 记录原子关联。仅凭总数无法证明各推进点各一次，也可能被管理夹具/额外观测请求污染。

结论：时钟计划 **不通过 readiness**。返工必须规定连接无关或固定连接的持久 UPDATE 观测机制、审计清零边界、单一被测请求标识/同步边界，以及外层和内部推进的可判别方法；排队测试必须用 channel/barrier，不得靠 sleep，并分别报告 wall-time sample 与持久 UPDATE。

### 3.3 备份与恢复：清单覆盖充分，最终安装失败门禁不可执行

计划 B01 覆盖项目自身 `Backup` 产物对 Schema 1/2/3 的往返及高水位；B02 覆盖重复键、未知键、尾随值、类型、缺字段、非对象和非规范摘要；B03 覆盖重复/未知/缺失 entry、非普通类型、路径名、大小边界、截断、校验、版本/schema 及顺序；B04 要求每个失败案同时验证 absent target 与 existing-empty target，且无 `.restore-*` 残留。这些比现有仅验证 absent target 的回归完整，断言方向正确。

实际 `Restore` 在所有内容校验后会先 `os.Remove(directory)`，再 `os.Rename(temporary, directory)`。若 rename 失败，原空目标已经消失，违反“任何失败不得删除或替换目标”。计划 B05 识别了 D-RESTORE-02，却允许“平台无法确定性注入则 `Not Run`”。这不是可裁剪项目：它正是候选代码可见的失败窗口，也是本任务明确要求的失败恢复门禁。允许 Not Run 后仍继续裁决，会造成误导性通过。

结论：manifest/tar/普通失败原子性覆盖 **通过 readiness**；最终安装失败原子性 **不通过 readiness**。返工必须给出在临时父目录、无扩权条件下可重复的最终边界故障注入方案及明确命令/断言；若当前实现没有可测试 seam 且不允许产品改动，计划应把这一事实列为执行前阻塞条件，而不是执行后的可选 Not Run。

### 3.4 分类、范围与不适用项

- 正常：A01、C01、B01 有真实结果断言。
- 边界：A02–A04、C03–C05、B02–B03 覆盖关键语法、时间和归档边界。
- 隔离：A05、全矩阵临时目录/SQLite/端口、受控 HTTP 上游和零真实供应商要求充分。
- 并发：A05、C03 真实覆盖；manifest JSON 解析本身标 N/A 有依据，离线单归档没有独立并行语义，目标变化由 B05 边界承担。
- 失败恢复：C05/C06、B04/B05 已列，但 B05 的 Not Run 逃生口使该分类不能算可执行完成。
- UI 视觉 N/A 有依据：三项不改变页面交互；固定 `make verify` 保留现有 UI 回归。真实供应商 N/A 有依据且必须保持：这些合同可由本地确定性夹具验证，不能把受控上游冒充供应商验证。

### 3.5 同版证明、日志、回执与隔离

计划要求候选前后 SHA/status/源码摘要、工具链、完整 stdout/stderr、退出码、receipt、原样 `make verify` 和零真实服务访问，基本同版链充分；也明确旧 Worker 绿回执、静态 browser.check 和分项测试不能替代全量门禁。

但计划第 118 行建议把 `make-verify.stdout`、`make-verify.stderr` 和逐矩阵完整日志放入 `docs/.../qa/evidence/<sha>/`。这与 seq 29 的私有产物约束直接冲突，且“若项目范围不允许再由控制器指定”把必要落点推迟到执行阶段，违反 readiness 的无歧义要求。绝对私有路径也不是可移植索引。

结论：证据方案 **不通过 readiness**。返工必须预先固定：原始文件写入 Git 忽略的 `test-results/...` 或控制器私有运行目录；报告只登记候选 SHA、命令、起止时间、退出码、字节数、SHA-256、脱敏内容摘要和控制器可解析的相对/receipt ID；仓库 `qa/evidence` 若使用，只能保存脱敏白名单摘要。必须明确凭据、数据库、截图、完整临时日志、工作区副本和原生会话历史不进入 Git。

## 4. 差异清单与所需处理

| ID | 严重度 | 差异及依据 | 所需处理 |
|---|---|---|---|
| B-AUTH-01 | Critical / candidate | 精确 `Bearer ` 匹配和“小写应 401”回归违反 RFC 9110 §11.1、RFC 5234 §2.3；项目需求只要求有效密钥 401/授权隔离，未授权收紧 scheme 大小写 | 保留 A02 的规范断言；产品候选修复后重新冻结。QA 不得改成精确大小写期待 |
| B-AUTH-02 | Critical / candidate | 精确单 SP 与“双 SP 应 401”回归违反 RFC 6750 §2.1 `1*SP` | 保留 A03 双/多 SP 成功、HTAB/前后/附加内容失败的区分；产品修复后复验 |
| R-CLOCK-01 | High / plan blocker | SQLite TEMP trigger 的连接范围未约束；连接池可能漏计，造成假 Pass | 固定连接或采用连接无关审计；写明安装、清零、读取、清理和连接池配置 |
| R-CLOCK-02 | High / plan blocker | 外层/admit/attempt 都更新同表，计划没有来源判别和夹具副作用隔离 | 定义单请求同步边界与可判别证据；每次辅助请求不得混入计数 |
| C-CLOCK-01 | Medium / candidate coverage | 现回归只统计 `wallTime` 调用，不能证明持久推进次数 | 仅作辅助回归，不能作为 C01/C02 核心证据；由返工后的观测补足 |
| B-RESTORE-01 | High / candidate coverage | 现回归只验证 absent target，未系统验证 existing-empty、tar 异常和临时目录清理 | 保留 B01–B04，逐 case 双目标状态和父目录快照 |
| D-RESTORE-02 | Critical / candidate hypothesis + plan blocker | `os.Remove(directory)` 成功后 `os.Rename` 失败会丢失原目标；计划允许无法注入时 Not Run | 执行前提供确定性临时文件系统故障方案；否则 readiness 继续 BLOCKED，不得以 Not Run 放行 |
| R-EVIDENCE-01 | High / plan blocker | 完整日志建议进入仓库 evidence，与 seq 29 私有产物边界冲突 | 原始证据固定到 Git 忽略/私有目录；Git 内只留脱敏摘要、哈希和可复验 receipt 索引 |
| R-SCOPE-01 | Medium / ambiguity | 第 120–127 行说只允许四个独立 QA 文件，第 129 行又建议优先修改三个既有回归，执行写路径不明确 | 返工明确允许写路径；不得由执行 Agent 临时选择或修改产品实现 |

## 5. 返工完成条件

只有以下条件全部满足后，才可重新申请 readiness review：

1. 不改变 A02/A03 的 RFC 断言，明确当前候选预期 NO-GO；候选修复属于产品任务，不由 QA 执行临时处理。
2. C01/C02 给出连接可靠、来源可判别、不会被夹具或额外请求污染的持久推进观测协议；C03–C06 保留排队、并发、高水位、回拨、各写失败、重启和恢复。
3. B05 有确定性、仅限临时父目录的最终 rename 边界故障方案；不能执行就作为执行前 blocker，不得标 N/A/Not Run 后放行。
4. 证据落点完全符合 seq 29：私有原始证据 + Git 内脱敏摘要/receipt 索引，并明确同版校验和保留边界。
5. 明确后续 QA 任务的唯一允许写路径和测试资产；矩阵每项均有前置数据、实际入口、操作、可观察断言、失败判定、隔离方式和证据索引。

在上述返工完成并再次评审为 `READY` 之前，后续执行必须保持停止。
