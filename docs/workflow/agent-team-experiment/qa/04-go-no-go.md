# QA v4 独立 Go/No-Go 复核

## 唯一结论

**`NO-GO`**。

这是产品候选的业务裁决，不改变原执行任务的 **`BLOCKED—RETURN TO PLAN REVIEW`** 状态，也不声称 v4 测试矩阵已完成。冻结候选已有两个可复验的关键产品失败：Bearer 认证 scheme 大小写变体及双/多 SP 正向形式被错误拒绝；Restore 在最终 rename 失败时删除既有空目标。任一项都足以禁止 GO。此外 B03/B04、X01–X03 未完成，C05 的完整业务零点证据不足，故本报告不释放产品执行、合并或部署。

## 复核对象、版本与证据索引

- 唯一正式输入为 `docs/workflow/agent-team-experiment/qa/01-contract-plan-v4.md`（SHA-256 `9c751029…85d8fe4`）、`02-readiness-v4.md`（`7cbc07dc…d8c830b`）和执行 worktree 中的 `03-execution.md`（`e6417597…60f2bf`）。v3 与旧 BLOCKED 评审仅是历史，未作为通过依据。
- 当前复核基线及执行 E 均为 `ca6ba0129b58cd07673c9a122512a094e9a57ec3`；执行原件记录 P=`2aeeb3af…65e367`、36 项 P manifest=`ed7f6894…d8005`、冻结 QA=`de72fa13…07968`、F=`1771d303…fa0d`。本任务开始时工作树干净；执行 worktree 的仅有未跟踪交付为唯一 QA、执行报告及其脱敏索引。
- 独立重算确认执行 worktree 的上述三份交付文件 bytes/hash 与控制器只读回执一致；`freeze.txt` SHA-256=`dd8b81d2…f7e24`，`p-manifest.txt` SHA-256=`ed7f6894…d8005`。P 到 E 的 tracked 差异仅新增 v4 计划与 v4 readiness；产品/构建内容未见漂移。
- 原始证据索引为 ignored 的 `test-results/qa-contract/<F>/20261101T063742Z/`。逐项检查 A01–A05、C01–C07、B01–B05 的 `command/start/end/exit/stdout/stderr/hashes`；全部哈希校验通过、UTC 时间闭合、真实 exit 与 stdout 相符。`test-results/` 及具体 receipt 均由项目 `.gitignore` 命中。X01–X03 没有 receipt。
- 执行报告与脱敏索引 SHA-256 分别为 `e6417597…60f2bf`、`7acd487b…333ff`；后者只作定位，裁决来自原始 receipt、冻结测试源和产品源码的交叉核验。

## v4 矩阵独立对账

下表的“Pass”只表示该项已有同版、完整度足够的已执行证据；“证据不足”不因命令 exit 0 而算 Pass。

| ID | 独立结果 | 复核依据 |
|---|---|---|
| A01 | Pass | 原始命令覆盖 handler/实际 HTTP、普通生成与 SSE，exit 0，哈希闭合。 |
| A02 | **Fail** | 原始 stdout 显示 `bearer`、`BEARER` 和两种混合大小写均返回 401，基准 `Bearer` 成功。产品 `identifyKey` 以大小写敏感的 `HasPrefix("Bearer mr_")`/`TrimPrefix("Bearer ")` 实现。 |
| A03 | **Fail** | 单 SP 成功，双 SP、四 SP 返回 401；负向测试未将这些正向误固化为通过条件。 |
| A04 | Pass | key/identity、人员/共享资源与隔离命令 exit 0，原始回执闭合。 |
| A05 | Pass | race-enabled 并发配置、生成与 metering 命令 exit 0；原件未见 race 报告。 |
| C01 | Pass | 外层单次推进及 health 例外专项 exit 0，命令/测试名映射成立。 |
| C02 | Pass | admit/邀请/三候选失败链专项 exit 0，命令/测试名映射成立。 |
| C03 | Pass | in-process 与实际 HTTP 排队截止 barrier 专项 exit 0；不是 sleep 型 receipt。 |
| C04 | Pass | 高水位防回拨及资源不复活专项 exit 0。 |
| C05 | **证据不足** | 专项 exit 0，冻结 QA 源真实断言 normal/begin/read/update/commit 计数、commit-hook 扩展码、持久/内存不变及 handler 503/degraded；但未按 v4 建立并机械断言 request/attempt/usage/其他业务表及受控 upstream 的零点，不能裁为完整 Pass。 |
| C06 | Pass | 重启、worker、Backup/Restore 时钟恢复命令 exit 0。 |
| C07 | Pass | schema 1/2 迁移兼容相关命令 exit 0。 |
| B01 | Pass | 合法 Backup/Restore、重启和安装恢复命令 exit 0。 |
| B02 | Pass（已执行范围） | manifest 严格对象、重复/未知/类型/缺字段/digest 负例命令 exit 0；这不替代 B04 的逐例双目标原子性。 |
| B03 | **未完成** | receipt exit 0，但命令只运行 `TestRestoreStrictManifest`；未执行 v4 的 tar entry/type/path/size/order、截断/校验等完整矩阵。 |
| B04 | **未完成** | 唯一 attempt 混入 B05 seam 且 exit 1；没有对每个 B02/B03 负例执行 absent 与 existing-empty 的完整不变性断言。 |
| B05 | **Fail** | 原始 stdout 显示 existing-target 00–09 共 10/10 在 sentinel 后消失；absent 分支未失败。源码在调用注入的最终 `rename` 前先 `os.Remove(directory)`，与现象一致。 |
| X01 | **未执行** | 无 receipt；冻结门禁触发后停止。 |
| X02 | **未执行** | 无 receipt；冻结门禁触发后停止。 |
| X03 | **未执行** | 无 receipt；原样 `make verify` 未运行。因此不得宣称所有必跑阶段执行或全量绿色。 |

## 合同、真实性与隔离复核

- Bearer 裁决符合正式合同及 HTTP 认证框架：RFC 9110 §11.1 将 `auth-scheme` 定义为大小写不敏感 token，credentials 语法允许 `1*SP`；RFC 6750 §2.1 的 Bearer credentials 同样使用 `1*SP`。因此 A02/A03 是产品缺陷，不是错误断言。官方依据：<https://www.rfc-editor.org/rfc/rfc9110.html#section-11.1>、<https://www.rfc-editor.org/rfc/rfc6750.html#section-2.1>。
- 时钟原件支持单次推进、barrier 排队、高水位、重启/Restore 与四阶段故障核心事实；commit 子例同时要求 `[1,1,1,1]`、hook=1 和 `SQLITE_CONSTRAINT_COMMITHOOK`。但 C05 缺少 v4 要求的完整业务/upstream 零点机械断言，必须随新冻结候选补齐并从头重跑，不能由设施自测或 exit 0 替代。
- Restore 合法往返与 manifest 拒绝有原件；完整 tar 畸形矩阵和逐负例无部分恢复证据缺失。B05 则有两套同向证据：冻结 QA 的 10 次硬断言和既有 seam 测试均复现 existing target 被删除。
- 命令均限定 Go 包内测试，冻结 QA 使用 `t.TempDir`、`httptest`、临时 SQLite/受控入口；receipt 未显示访问常驻服务、真实数据、凭据或供应商。原始运行物留在 ignored 目录。未发现为通过而修改产品、既有测试、门禁或数据库的证据；正式矩阵开始后没有改写冻结 QA，而是停止。由于 X 项未运行，本结论只确认已执行部分的隔离证据，不把缺失范围推定为通过。

## 遗留风险与返工依据

1. 修复认证解析：auth-scheme 必须 ASCII 大小写不敏感，并接受合同规定的单/双/多 SP；随后在 models、普通生成、SSE 及拒绝路径上重跑 A02/A03/X01/X02，确认无授权或计费扩张。
2. 修复 Restore 最终安装的原子性：注入的最终 rename 失败时，existing-empty 的身份、mode、内容/清单和父目录必须不变，absent 必须仍 absent，且无临时残留；重跑 B05 两状态各 10 次。
3. 经新的独立计划/readiness 授权，补齐 B03 全部 tar 畸形矩阵、B02/B03 每个负例的 B04 双目标断言，以及 C05 业务/upstream 零点断言；生成新 F 后从头重跑 A/C/B/X，最后原样执行一次 `make verify`。不得用本轮已有绿色 receipt 拼接成新候选的 GO。

本 `NO-GO` 仅供 Leader 安排修复与重新验收；不授权本任务修复产品、修改测试、创建外部事项、提交、推送、合并或部署。
