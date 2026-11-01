# 三项合同独立复核与 QA 执行计划

## 1. 结论与版本冻结

**计划交付裁决：COMPLETED；当前候选合同裁决：预期 FAIL / NO-GO。** 三项成果已实际合入当前实验分支，因而满足本次复核的开始条件；正式合同能够由 RFC 和项目材料裁决，覆盖矩阵也可直接执行，所以发现产品缺陷不阻塞 QA 计划交付。当前 Bearer 实现和回归把认证 scheme 当成大小写敏感，并把 RFC 6750 的 `1*SP` 收紧为恰好一个 SP，均与正式协议语法冲突。独立执行应保留这些预期失败用例并据实判定候选 NO-GO，不得用现有实现测试自证合同，不得自行改产品或放宽断言。

第一次交接曾将“产品存在协议阻塞”等同于“计划任务 blocked”；该原始报告结论和发给 Leader 的缺陷消息保留在控制器历史中。控制器随后明确：只要正式合同可裁决且计划可执行，本任务应 completed，同时把候选标为可能 NO-GO。本版据此只修正计划交付状态，不撤销 B-AUTH-01，也不把任何预期失败改成通过。

本报告只做静态合同复核和执行计划，没有运行专项测试或 `make verify`，也不把三名 Worker 的报告或其第 6 轮门禁结果视为独立 QA 结论。

### 1.1 候选和合入证据

| 对象 | commit | Git 事实 |
|---|---|---|
| QA 候选 / 当前 HEAD | `122c9abece28f562616483e991098d020d236d6d` | 当前分支 `agent-team/task-qa-contract-plan-1`；也是最后一项合入后的实验分支头 |
| strict-bearer-auth 工作提交 | `67f464494f4d59be8a20ef99d9b15175ecd00e78` | 集成提交 `2d48ee08f12368e1b806fcc0975943e2eca0c8ee` |
| single-clock-advance 工作提交 | `318639d819165f3309f123133fdaf6229003270c` | 集成提交 `532e713acc998162c682b310c8f280e1651e6f95` |
| strict-backup-manifest 工作提交 | `7dc57094557b8883c130a41ed306d1223c0d92f8` | 集成提交 `122c9abece28f562616483e991098d020d236d6d` |
| 三项合入前实验基线 | `521751f1a0afaca26bd50b989f6f0df2d09f10e5` | `521751f..122c9ab` 为本轮相关总差异 |

复核开始时 `git status --short --branch` 仅输出分支行 `## agent-team/task-qa-contract-plan-1`，无已跟踪或未跟踪改动。三项总差异是 11 个文件、706 行新增、13 行删除：3 份 Worker 文档、3 个产品文件组及 3 个新增回归文件；实际产品变化集中于 `console_keys.go`、`server.go`、`identity.go`、`console_api.go`、`backup.go`。本报告写入后工作树预期仅新增本文件；执行阶段必须重新采集状态和差异，不能复用此快照。

### 1.2 正式依据和项目依据

- RFC 9110 §11.1：HTTP authentication scheme 是大小写不敏感的 token；凭据结构为 `auth-scheme [ 1*SP ( token68 / #auth-param ) ]`。<https://www.rfc-editor.org/rfc/rfc9110.html#section-11.1>
- RFC 6750 §2.1：Bearer 用于 `Authorization`，语法为 `credentials = "Bearer" 1*SP b64token`，且客户端不得用多种传递方式提交同一 token。`1*SP` 明确是一个或多个 SP；其 `b64token` 字符集比本项目密钥格式宽，本项目可合法进一步限制凭据本体。<https://www.rfc-editor.org/rfc/rfc6750.html#section-2.1>
- RFC 5234 §2.3：ABNF 字符串字面量默认大小写不敏感；RFC 6750 中的字面量 `"Bearer"` 因此也不能解释为精确大小写。<https://www.rfc-editor.org/rfc/rfc5234.html#section-2.3>
- Go `encoding/json.Decoder` 官方文档用于解释逐 token 解码、`More` 和 `Token` 的输入边界；具体“唯一键”是项目在 token 层自行实施的约束，不是 Go 默认保证。<https://pkg.go.dev/encoding/json#Decoder.Token>
- SQLite 官方备份说明只作为一致快照背景；本项目的正式合同仍是仓库 `Backup`/`Restore`、独占目录锁、checkpoint、摘要和原子安装。<https://www.sqlite.org/backup.html>
- 项目范围依据：`AGENTS.md`、`docs/product-brief.md`、`docs/roadmap.md`、`docs/releases/v0.1.0.md`、`docs/workflow/runs/4cd62ced6a4f98f50e3cca6f7596eb06/product-definition.md` 及其设计/迁移运营文档。基础合同要求 Go 后端、密钥保护、普通/SSE、持久化、重启和隔离备份恢复持续回归；本轮不新增协议、供应商或产品功能。
- Worker 文档 `strict-bearer-auth.md`、`single-clock-advance.md`、`strict-backup-manifest.md` 仅用于定位其声称意图和历史执行记录，结论均由本报告依据实际代码重新裁决。

## 2. 已确认合同

### 2.1 Bearer

1. `Authorization` 的 authentication scheme 必须按 ASCII 大小写不敏感匹配 `Bearer`。`Bearer`、`bearer`、`BEARER` 以及混合大小写都是同一 scheme；项目无权把标准 scheme 名称改成大小写敏感。
2. RFC 6750 要求 scheme 与 credential 之间 `1*SP`，即一个或多个 SP；RFC 9110 通用框架相同。当前 Worker 任务虽写了“精确结构/拒绝空白变体”，但它是实现任务说明，不能把标准明确接受的两个或更多 SP 重新定义为非法 Bearer。候选拒绝额外分隔 SP 应作为 B-AUTH-02 预期失败。HTAB 不是 SP，前导空白、尾随空白、逗号附加内容也不属于该凭据语法，仍应拒绝。
3. 项目可合法把 bearer credential 从通用 `b64token` 收紧为当前生成的 `mr_` 加固定总长度 46，并继续要求摘要命中、generation 有效、key 启用、未撤销、未过期、租户/模型授权正确。空 credential、裸 credential、Basic、错误长度/前缀、附加内容均拒绝。
4. `/v1/models` 和生成入口必须使用同一解析合同；认证失败在任何上游访问、计量或业务授权之前发生，并返回现有 401 安全错误，不泄漏密钥存在性。

**已证静态冲突 B-AUTH-01：** `controlledAppID` 使用 `strings.CutPrefix(header, "Bearer ")`，`identifyKey` 先用 `strings.HasPrefix(header, "Bearer mr_")`；新增回归明确期待 `bearer <合法密钥>` 返回 401。这违反 RFC 9110 scheme 大小写不敏感规则以及 RFC 5234 对 ABNF 字面量的解释。即使原 Worker 任务文字要求拒绝“大小写变体”，也不能覆盖正式外部协议。执行 A02 应预期在当前候选失败，并据此给出 NO-GO。

**已证静态冲突 B-AUTH-02：** `CutPrefix("Bearer ")` 之后固定长度检查会让第二个分隔 SP 成为 credential 的首字符并拒绝；新增回归也期待 “extra separator whitespace” 为 401。RFC 6750 的 `1*SP` 明确允许多个 SP，因此 A03 对两个及多个 SP 的不可变期望是与单 SP 相同的认证结果。项目仍可严格拒绝 HTAB、前导/尾随空白和 credential 内附加内容；这些不能与合法的分隔 SP 混为一谈。

### 2.2 有效时钟

1. Schema 3 的每个外部 HTTP 请求除 `/healthz` 外，在 `Server.ServeHTTP` 统一执行一次**持久化外层推进**；推进发生在与管理/路由共享的 `server.mu` 排队边界。失败必须先停止分派，设置 degraded 并返回 `503 storage_unavailable`。
2. “一次”的可观察量不是源码调用点数量，也不是 `Store.clock`/`wallTime` 调用次数。排队分支故意在等待前做一次非持久化墙钟采样，持久化推进仍只有一次。独立 QA 应在临时 SQLite 上对 `console_clock` 安装 TEMP/测试 trigger 审计 UPDATE 次数，同时读取持久化 `effective`，并结合 HTTP 状态和上游命中计数。
3. `Store.advanceClock` 在事务中读取高水位，取 `max(previous, wallTime)`，写回并提交；只有提交成功才更新内存 `effectiveTime`。墙钟回拨不能降低内存或持久化高水位；重启及 `Backup`→`Restore` 后必须连续。
4. 必要内部推进保留：`Store.admit` 的准入事务、`beginAttempt` 的逐尝试事务，以及不经普通已认证管理入口完成邀请接受的事务时点。它们不是同一“外部入口推进”计数的重复项；QA 必须分别断言其业务必要性和失败关闭。
5. 并发请求通过互斥/SQLite 排队形成单调高水位。应测试排队期间墙钟越过密钥、邀请、配额窗口截止点，最终业务判定采用取得锁后的有效时间，且无过期凭据复活、无额度回补。
6. Schema 1/2 的 `advanceClock` 是 no-op；本轮严格一次持久化只适用于 Schema 3。`/healthz` 不推进是明确例外，因为健康探针不应产生持续写事务。

静态风险 C-CLOCK-01：新增 `TestExternalHTTPRequestAdvancesEffectiveClockOnce` 仅统计墙钟函数调用，不能完整证明数据库只 UPDATE 一次，也不覆盖竞争分支的额外非持久化采样。它可作辅助回归，不能作为核心合同证据。

### 2.3 备份清单与恢复

1. 项目 `Backup` 在目录锁内 checkpoint WAL，读取支持的 schema 1/2/3，生成仅含 `manifest.json` 和 `relay.db` 的 tar；manifest 为整数 `version` 和数据库 SHA-256 的 64 位小写十六进制。新输出路径必须不存在，失败输出应删除。
2. `Restore` 只接受单一 JSON 对象，`version`、`sha256` 各恰好一次，无未知键；拒绝重复键、错误类型、缺字段、尾随第二个 JSON 值、空/短/大写/非十六进制/非规范摘要。JSON 末尾空白合法。
3. tar 必须恰有一个普通文件 `manifest.json` 和一个普通文件 `relay.db`；拒绝重复名、未知名、目录/链接/设备等非普通类型、负数/超限大小、超大 manifest、缺项、截断/损坏、unsupported version、摘要不符和 schema/version 不符。顺序可为任一，因为现实现按名称收集并在末尾校验。
4. 合法的项目 `Backup` 必须能往返到新目录，重开后保留 schema、配置、密钥状态、记录和 `console_clock` 高水位。测试不得读取真实备份或真实 key file。
5. 所有失败只能影响父目录内专用 `.restore-*` 临时目录，退出后清理；目标若不存在仍不存在，目标若已存在且为空则内容/身份不得被部分替换，非空目标必须在解包前拒绝。最终安装只能在全部数据库、schema、摘要检查成功后 rename。

静态缺口 B-RESTORE-01：现有严格清单回归只对“不存在的 target”断言 `os.IsNotExist`；尚未证明已存在空目标在每条失败路径均未被移除/替换，也未系统覆盖 tar 结构异常。计划补足，但此项本身是测试缺口而非已证产品缺陷。最终阶段“删除空目录再 rename”若 rename 失败，可能留下目标缺失；需用可控文件系统失败验证，若可复现则升级为恢复原子性缺陷，不以难以注入为由放行。

## 3. 未决差异与进入条件

| ID | 等级/状态 | 差异 | 执行前要求 |
|---|---|---|---|
| B-AUTH-01 | Candidate blocker / 已证静态冲突 | scheme 被精确大小写匹配；小写回归与 RFC 9110/RFC 5234 相反 | A02 保留预期失败；若实测仍拒绝则候选 NO-GO，后续产品任务修正，QA 不代改 |
| B-AUTH-02 | Candidate blocker / 已证静态冲突 | 额外分隔 SP 被拒绝；回归与 RFC 6750 `1*SP` 相反 | A03 保留双 SP/多 SP 预期成功；若实测拒绝则候选 NO-GO，HTAB/前导/尾随空白仍预期拒绝 |
| C-CLOCK-01 | Coverage / 待执行 | 现测试把 wall clock observation 当持久化推进代理 | 新独立测试用 SQLite trigger/持久值观测外层 UPDATE；竞争路径单列采样数，不改产品 |
| B-RESTORE-01 | Coverage / 待执行 | 失败原子性仅测不存在目标，结构异常覆盖不全 | 对每类失败同时跑 absent target 与 existing-empty sentinel target；核对无 `.restore-*` 残留 |
| D-RESTORE-02 | Defect hypothesis | 删除既有空目标后最终 rename 失败可能使目标消失 | 设计可控同文件系统失败；若无法可靠注入，明确 Not Run，不能宣称覆盖；若复现则阻塞恢复原子性 |

## 4. 可执行覆盖矩阵

所有用例固定使用 `t.TempDir()`、测试生成的 SQLite/key、`httptest.Server` 或 `127.0.0.1:0` 临时端口，以及名称明确的受控上游。禁止环境中的常驻服务、用户数据目录、真实归档、供应商配置/凭据及真实供应商网络。原始证据按候选 SHA 存入 QA 回执目录；不得写入本报告或产品源码目录，也不得记录 credential 全文。

| ID / 类别 | 前置数据 | 实际入口与操作 | 可观察断言 / 失败判定 | 隔离与原始证据 |
|---|---|---|---|---|
| A01 正常 | 临时 Schema 3；一把只授权一个模型的合法 `mr_` key；零命中受控上游 | 实际 Go handler/HTTP：`GET /v1/models` 与普通生成，`Authorization: Bearer <key>` | 200；模型范围不扩大；生成才命中受控上游。任何 401、越权模型或提前上游访问失败 | 临时 DB/端口；`auth-normal.log`、HTTP 状态/脱敏响应、上游计数 |
| A02 scheme 边界 | 同 A01 | 表驱动 `Bearer`、`bearer`、`BEARER`、混合大小写 | 全部视为同一 scheme 并到达相同 credential 校验结果。任一仅因 scheme 大小写 401 即复现 B-AUTH-01 | 不打印 key；`auth-scheme-case.log` |
| A03 空白/结构边界 | 同 A01 | 缺 header、裸 key、Basic、空 credential、前导空白、HTAB、两个及多个 SP、尾随空白、逗号/附加 token | 缺失/错误 scheme、空 credential、前导/HTAB/尾随/附加拒绝；两个及多个分隔 SP 与单 SP 同结果。当前候选拒绝多 SP 是 B-AUTH-02 的预期 Fail；任何无 scheme 或附加内容成功也失败 | handler 与 actual HTTP 各一组；`auth-syntax.log` |
| A04 credential/生命周期 | 合法、错误前缀、45/47 长度、随机 `mr_`、禁用、撤销、过期、轮换宽限、新 generation | models 与普通/SSE 生成 | 仅当前有效/宽限内 generation 成功；其余 401；失败均上游 0、无 prompt/credential 日志 | 临时受控上游；`auth-lifecycle.log` |
| A05 隔离/并发 | 两租户、不同 grant；并发合法/非法请求 | actual HTTP 并发 models 与生成 | 无跨租户/跨模型泄漏；非法请求 401 且不消耗/分派；race 无报告 | 临时 DB、随机端口；`auth-isolation-race.log` |
| C01 正常外层推进 | Schema 3；`console_clock` audit trigger；模型、identity、console 三类请求 | 每类单次 actual HTTP；请求前后读 audit 与 effective | 每个非 health 外部请求恰好一个外层 UPDATE；`/healthz` 为 0。仅数 wallTime 不算证据 | 测试事务/trigger 仅临时库；`clock-entry.log`、SQL 快照 |
| C02 内部事务 | 受控单候选和多尝试模型、可准入请求 | 普通与 SSE 生成；记录入口、admit、每次 beginAttempt 推进 | 外层一次；admit 一次；每个实际 attempt 一次；数量与 attempt 记录一致，不能把内部推进误报重复 | 受控失败/成功上游；`clock-internal.log` |
| C03 排队并发 | 轮换/邀请/配额截止点；锁住 `server.mu`，请求排队后推进测试钟 | in-process 与 actual HTTP；等待采样后跨越 deadline 再放锁 | 持久外层 UPDATE 仍一次；允许额外非持久 wallTime sample；过期/额度按排队后时间判定且上游 0 | channel 同步，禁止 sleep 判定；`clock-queue.log` |
| C04 高水位/回拨 | 先推进到 T1，再把 wall clock 设 T0<T1 | HTTP 读、管理写、生成 | 内存/SQLite effective 均不低于 T1；过期资源不复活；月份/分钟/预算不回退 | 确定性测试钟；`clock-rollback.log`、脱敏 SQL |
| C05 写失败 fail-closed | 分别注入 begin/read/update/commit 失败（能由临时 DB/锁/连接安全触发者） | models、console、identity、生成 | 503 + degraded；无 handler 继续、无上游、无部分业务写；不能只测已关闭 DB 的 begin 失败 | 每案独立临时安装；`clock-failclosed.log` |
| C06 重启/恢复 | T1 高水位、截止资源、用量；关闭后重开；项目 Backup→新目录 Restore | 实际 Open/New 后请求 | effective=T1；回拨启动不复活资源/额度；恢复后相同；合法当前 key 仍按 grant 工作 | 子进程仅接收临时状态文件；`clock-recovery.log` |
| C07 Schema 1/2 | 测试生成旧 schema 夹具 | 非 health HTTP 与升级前读取 | 无 `console_clock` 持久更新要求；既有请求行为不因本轮改变。N/A 理由：严格高水位表是 Schema 3 合同 | 仓库自有旧版夹具；`clock-legacy.log` |
| B01 合法往返 | 项目 API 创建 Schema 1/2/3 最小合法安装，含配置、key 状态、记录、Schema 3 clock | `Backup`，检查 tar，再 `Restore` 到新目录并 Open | 恰有两普通 entry；manifest 规范；摘要匹配；所有状态和高水位守恒 | 每版独立 temp；`backup-roundtrip.log`、tar listing/摘要 |
| B02 JSON 严格性 | 从 B01 归档仅替换测试 manifest | 重复 version/sha256、未知键、尾随 JSON、类型错、缺字段、非对象、空/短/大写/非 hex digest；另测末尾空白 | 所有非法输入在 DB open 前拒绝；末尾 JSON whitespace 成功；错误不得泄漏内容 | 每 case 新 archive/target；`backup-manifest.log` |
| B03 tar 结构 | 测试构造 tar | 重复 entry、未知 entry、缺项、非普通类型、路径名、负/超限大小、manifest >4096、DB >1GiB header、截断/校验错、version/schema 不符、两种 entry 顺序 | 非法全部拒绝；合法反序归档按当前名称合同成功；不得创建目标 | 不分配 1GiB payload，只写受控 header；`backup-archive.log` |
| B04 失败原子性 | B02/B03 每个 case 各配 absent target 和 existing-empty target（含 sentinel 身份/权限快照） | 调用 Restore，失败后检查 target、父目录和临时项 | absent 仍 absent；existing-empty 仍为同一预期目录状态；无 relay.db、无替换、无 `.restore-*`；任何部分安装失败 | 独占 temp parent；`backup-atomicity.log`、目录前后清单 |
| B05 最终安装失败 | 合法归档、既有空 target；可控 rename/权限/竞争条件，且不越出 temp parent | 在最终 rename 边界制造失败 | 返回错误且原 target 仍存在、无部分 DB；否则复现 D-RESTORE-02。若平台无法确定性注入则标 Not Run，不可标 Pass | 专用 temp filesystem；`backup-finalize-failure.log` |
| X01 跨改动：恢复后认证/时钟 | Schema 3 合法安装含 key/grant、T1 clock；Backup/Restore | 回拨时钟启动恢复实例；大小写变体 scheme 请求 models/生成 | 高水位连续；认证 scheme 仍大小写不敏感；grant 不扩大；非法 credential 上游 0 | 受控 HTTP；`cross-restore-auth-clock.log` |
| X02 跨改动：失败优先级 | 时钟写失败 + 各种 Authorization | 非 health models/生成 | 外层时钟推进先 fail-closed 为现有 503，任何请求都不触达上游；恢复存储后认证再按 A02/A03 判定 | 每案新临时实例；`cross-failure-order.log` |
| X03 全量回归 | 已修复且冻结的候选 SHA，干净源码快照 | 原样执行 `make verify`，不包裹吞码管道 | 所有阶段真实 exit 0；任何 skip/截断/非零按门禁合同处理；Worker 第6轮结果不可替代 | 独立 receipt，完整 stdout/stderr 与 exit code |

并发不适用于 manifest 的 JSON 字段解析本身：`Restore` 是单次离线归档操作，目标目录并发改变已由“restore target changed”边界处理；计划在 B05 覆盖最终竞争，而不虚构并行 JSON 语义。真实供应商测试不适用：三项合同均可由本地 HTTP/tar/SQLite 确定性验证，且项目明确禁止把受控上游当真实供应商证据。UI 视觉测试不适用：三项不改变页面或交互；完整 `make verify` 仍保留现有 UI 回归以防跨改动。

## 5. 执行顺序与同版证明

计划已经完成；当前候选含 B-AUTH-01/B-AUTH-02，执行节点应先复现并可在确认候选 NO-GO 后停止昂贵全量门禁。若控制器仍要求收集完整失败面，则按以下顺序执行；后续修复候选必须从冻结步骤重来：

1. **冻结候选。** 记录 UTC 时间、`git rev-parse HEAD`、`git status --porcelain=v1 --branch`、`git diff --binary <三项前基线>..HEAD`、`git diff --check`、submodule（若有）和全部被测源码路径逐文件 SHA-256。候选 SHA 变化即废弃后续旧回执并从此步重来。
2. **记录工具链。** 原字节保存 `go version`、`go env GOVERSION GOOS GOARCH GOMOD GOWORK`、`make --version`、SQLite driver/module lock 状态，以及 `make verify` 自身报告的 Node/npm/Playwright/Chrome 版本；不得修改工具链凑通过。
3. **先复现候选阻塞。** 运行 A02 和 A03，保存小写/混合 scheme 及双 SP/多 SP 的真实状态。它们的正式期望来自 RFC，不来自同一实现的回归；若当前候选如静态分析所示拒绝，则记录 Fail/NO-GO。修复不属于本任务；新候选须重新冻结后再验证。
4. **专项由窄到宽。** A01–A05 → C01–C07 → B01–B05 → X01–X02。每条命令用直接重定向或能保留真实退出码的收集器记录完整 stdout/stderr、开始/结束时间、退出码、test 名称和随机种子（如有）；失败即保留现场摘要，不篡改数据库记录求绿。
5. **固定全量入口。** 在同一候选、同一源码摘要上原样运行 `make verify` 一次作为放行回执。不得用 Worker 第 6 轮、分项 Go 测试、静态 browser.check 或重跑片段替代。
6. **运行后对账。** 重采 HEAD、status、源码 SHA-256；必须与步骤 1 一致，允许新增的仅是预先声明且不含秘密的 QA receipt。核对完整 stdout/stderr 文件大小/摘要、exit code、门禁内部 receipt、源码 unchanged 标志和所有专项证据索引。
7. **独立裁决。** 每个矩阵 ID 标 Pass/Fail/Blocked/Not Run，Not Run 必须引用本报告的不适用理由或具体设施边界；任何 B-AUTH-01、恢复部分替换、高水位回拨、推进失败后继续分派、跨租户授权均为 No-Go。

建议回执目录结构（执行任务创建，不在本计划创建）：`docs/workflow/agent-team-experiment/qa/evidence/<candidate-short-sha>/`，包含 `candidate-before.txt`、`source-manifest.sha256`、`toolchain.txt`、逐矩阵日志、`make-verify.stdout`、`make-verify.stderr`、`make-verify.exit`、`candidate-after.txt`、`receipt-index.json`。若项目范围不允许写该目录，则由控制器指定外部持久证据目录并在最终报告登记绝对来源到可移植相对索引；不得把 `/tmp` 当长期证据。

## 6. 允许新增的单独 QA 测试文件

以下只是后续执行任务可申请的最小测试资产清单，本任务不创建；不得改产品实现来制造可观察性。

- `internal/relay/bearer_contract_qa_test.go`：RFC scheme 大小写、空白、credential/lifecycle、models/generation 一致性。
- `internal/relay/clock_persistence_qa_test.go`：临时 SQLite audit trigger、排队、内部事务点、begin/update/commit 失败与高水位。
- `internal/relay/backup_atomicity_qa_test.go`：严格 JSON、tar 结构、absent/existing-empty 目标和临时目录清理。
- `internal/relay/cross_contract_qa_test.go`：仅放恢复后认证/时钟及故障优先级，避免复制前三文件单项覆盖。

若既有 `bearer_auth_regression_test.go`、`clock_advance_regression_test.go`、`backup_manifest_regression_test.go` 能表达用例，应优先扩充其表驱动数据；但 B-AUTH-01 的错误期待必须由产品修复任务明确改正，QA 不得悄悄删断言。禁止新增第二套 verification runner、产品调试端点、假供应商默认配置或持久测试表。

## 7. 不可变断言

1. HTTP authentication scheme 大小写不敏感；Bearer scheme 与 credential 之间一个或多个 SP 均合法。任何候选不得以项目“严格”名义推翻 RFC 9110、RFC 6750 和 RFC 5234。
2. 只有有效 `mr_` generation 能认证；密钥全文、prompt、completion 和供应商 credential 不进入日志/证据。
3. 认证/时钟失败不得触达受控上游、产生调用扣费或扩大授权；跨租户和未授权模型始终拒绝。
4. 每个 Schema 3 非 health 外部请求只有一次外层持久时钟推进；必要的 admit/attempt/invitation 事务推进单独计数，不能机械删除。
5. 有效时间和持久高水位永不回拨；写失败 fail-closed；重启和合法备份恢复不丢连续性。
6. manifest 精确、唯一、规范；归档结构精确；摘要覆盖实际 `relay.db`；项目自身合法 Backup 必须可 Restore。
7. 任一 Restore 失败不得部分创建、替换或删除目标，不遗留临时安装；不得通过直接改数据库让测试通过。
8. 所有测试仅用临时目录/SQLite/端口和显式受控 HTTP 夹具，零真实供应商、零常驻服务、零用户备份/配置/凭据访问。
9. 放行只认冻结候选上的原始 `make verify` 与独立专项同版证据；命令、完整 stdout/stderr、exit code、工具链、源码前后摘要和 receipt 缺一不可。
10. 本轮不新增产品功能、不扩大协议面、不提交、推送、合并或部署；候选的 FAIL/NO-GO 不能被 Worker 报告、旧绿色回执或静态审查覆盖。
