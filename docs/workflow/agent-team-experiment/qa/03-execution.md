# QA v4 执行报告

## 结论

**`BLOCKED—RETURN TO PLAN REVIEW`**。本轮不能声称计划执行 completed，也不代替最终 Go/No-Go。执行到 B03/B04 时确认，冻结候选缺少 v4 要求的完整 tar 畸形矩阵和逐负例双目标原子性测试；正式矩阵开始后再修改唯一 QA 文件会改变 F，违反 v4 的冻结门禁，因此立即停止。X01–X03（包括原样 `make verify`）未执行。

停止前得到的产品事实仍保留：Bearer 合同为 Fail；Restore 最终 rename 的 existing-empty 目标 10/10 被删除，为 Fail。没有修改产品实现、既有测试、断言或门禁。

## 授权、版本和同版证明

- 独立授权：`02-readiness-v4.md` 唯一结论 `READY`；公共 integration commit `ca6ba0129b58cd07673c9a122512a094e9a57ec3`。
- E/执行前后 HEAD：`ca6ba0129b58cd07673c9a122512a094e9a57ec3`。公共 P：`2aeeb3af8db73a33f6fcb99371769b79f865e367`；计划集成点 `348e297f63cd4528fc05e5eaa1d7a0c2e3156d81`。三者祖先检查均为 exit 0。
- P manifest：36 项，SHA-256 `ed7f68947ecc410cf80d8d8769b9547cfcb8d498959621a8f6490ced6edd8005`，与 readiness 获审值一致。
- QA 文件 SHA-256：`de72fa137400ed00ca3fb806691bb62d0f6393aa820a4b5f1ca12b6a0fd07968`；完整新增 diff SHA-256 `25262bcd13ac82fc3b8ba381596ae72e08abcae5e8ed80dd3bd25742950f8540`。
- F：`1771d3037afa4c06187378171238bd1aa00d058219a231dcce0fc402ba59fa0d`。正式命令期间 QA 与产品/构建内容未变化；停止时工作树仅有该 QA 文件和本报告/摘要交付路径。
- 计划 SHA-256 `9c7510299982e7f15a2e89a69f7e1a90a8465dbb7b5a185a4db0ae65285d8fe4`；readiness SHA-256 `7cbc07dce28c34a03d54e48632c67dde6d26c5a6b2aaeb9b217d86312d8c830b`。
- 工具链：Go `go1.25.6 linux/amd64`，GOTOOLCHAIN `local`，CGO enabled，Node `v22.14.0`，npm `10.9.2`，Git `2.55.0`，GNU Make `4.3`。
- 原始 receipt 根索引：`test-results/qa-contract/<F>/20261101T063742Z/`（Git ignored）。freeze receipt SHA-256 `dd8b81d2f756a4be48d202018ad6f092ffa275b1525123c4acc6a31c935f7e24`。

首次预检因默认用户 Go cache 只读而未进入编译；正式执行固定使用 ignored 的隔离 `GOCACHE`/`GOMODCACHE`。该设施事实与产品失败分开。

## 矩阵结果

| ID | 结果 | 实际入口与事实 |
|---|---|---|
| A01 | Pass | handler 与实际 HTTP 的治理、普通生成和 SSE 专项 exit 0。 |
| A02 | Fail | 实际 handler 对 `bearer`、`BEARER` 和两种混合大小写返回 401；基准 `Bearer` 成功。 |
| A03 | Fail | 单空格成功；双空格、多空格被 401 拒绝。既有负向边界保持拒绝且未扩权。 |
| A04 | Pass | key/identity、人员/共享资源及隔离专项 exit 0。 |
| A05 | Pass | `go test -race` 并发配置、生成和 metering 隔离专项 exit 0。 |
| C01 | Pass | 外层推进及 health 例外专项 exit 0。 |
| C02 | Pass | 请求推进、邀请和三候选失败链专项 exit 0。 |
| C03 | Pass | in-process/实际 HTTP 排队截止 barrier 专项 exit 0。 |
| C04 | Pass | 高水位防回拨与资源不复活专项 exit 0。 |
| C05 | Pass | normal/begin/read/update/commit 真实设施计数分别匹配；commit 得到 SQLite commit-hook 扩展码；四阶段真实 handler 均 503/degraded。 |
| C06 | Pass | worker、重启、Backup/Restore 时钟恢复专项 exit 0。 |
| C07 | Pass | schema 1/2 迁移兼容专项 exit 0。 |
| B01 | Pass | 项目 Backup/Restore、重启和安装恢复专项 exit 0。 |
| B02 | Pass | 已执行的 manifest 严格对象、重复/未知/类型/缺字段/digest 负例均 exit 0。 |
| B03 | Blocked | 命令本身 exit 0，但只重复了 manifest 专项；没有执行 v4 明列的完整 tar entry/type/path/size/order 矩阵，故不得标 Pass。 |
| B04 | Blocked | 缺少每个 B02/B03 负例对 absent 与 existing-empty 的完整原子性断言。attempt 1 混入 B05 seam 后 exit 1，不可冒充 B04 证据。 |
| B05 | Fail | absent 10/10 保持 absent；existing-empty 10/10 在最终 rename sentinel 后被删除。错误保真且 rename 恰一次。 |
| X01 | Blocked | 因冻结门禁停止，未执行。 |
| X02 | Blocked | 因冻结门禁停止，未执行。 |
| X03 | Blocked | 原样 `make verify` 未执行；不得声称全量绿色。 |

## 缺陷与阻断复现

1. **Bearer 合同失败**：用有效 `mr_` key 请求真实 `/v1/models` handler，把 `Bearer` 分别替换为 ASCII 大小写变体或在 scheme 后使用双/多 SP。期望按 v4/RFC 6750 正向成功且授权范围不扩大；实际为 401。影响所有兼容客户端。证据：A02/A03 receipt。
2. **Restore 原子性失败**：项目生成合法 archive，预建 existing-empty target，通过 `restoreWithRename` 的最终 rename seam 返回唯一 sentinel。期望 sentinel 保真且 target 身份/mode/listing 不变；实际 10/10 target 不存在。影响失败恢复时既有目标。证据：B05 receipt。
3. **执行阻断**：要满足 B03/B04，必须向已冻结的 `qa_independent_regression_test.go` 增加 v4 已列的 tar 及双目标断言并生成新 F；本轮无权在正式矩阵后这样做。返工条件是计划评审明确授权新 QA 候选并要求从头在唯一新 F 上重跑 A/C/B/X 与原样 `make verify`。

## 禁入物和证据边界

完整 stdout/stderr、时间、exit、命令、哈希及临时 cache 只在 ignored `test-results/`。Git 交付仅含本报告、唯一 QA 源文件和脱敏 receipt 索引；未加入数据库、archive、密钥、截图、绝对运行路径、会话历史或凭据，未访问真实供应商。
