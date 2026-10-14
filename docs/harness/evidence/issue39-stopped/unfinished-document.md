# 隔离环境中的 Run 工作区共享

本文说明 Model Relay 独立 OpenSandbox Harness 的使用与证据边界。本 Issue 只验证受控研发流程，不修改业务功能，不调用真实供应商，也不把流程通过等同于产品或商业验收。

## 选择独立流程

Owner 创建标题以 `【Harness验证】` 开头、目标明确的新 Issue，并加 `harness-opensandbox` 标签。`.github/workflows/agent-platform-sandbox-entry.yml` 使用独立的 `AGENT_PLATFORM_SANDBOX_CONFIG`；旧入口 `.github/workflows/agent-platform-entry.yml` 排除此标签。入口检查 Owner 身份、当前 Issue 标签及非 PR 条件；手动入口也重新检查，不能用手动触发绕过标签。

当前安装流程 ID 为 `83f08f81f7040fcd0c12e6ff4e2acd2d`，平台维护源版本为 `3261b340a7d3b20b29f524739d346dd1b4025cda`，入口详情见 `docs/harness/opensandbox-entry.md`。私有配置、Token 和宿主路径不属于仓库产物；不要复制到 Issue、会话或报告。

## 固定源码进入工作区

Runner 运行仓库自有适配器，不创建 Agent 工作区；Actions checkout 仅用于执行入口脚本，并关闭凭据持久化。新任务通过 GitHub API 读取配置基线（本入口为 `main`）的完整 SHA，将 `repository`、`repository_commit` 随正式入站提交，要求 `expected_environment=opensandbox`。重复事件先找原 Run，不重新解析最新 main 来替换原版本。

平台准备节点在 Run 所属共享工作区拉取这个固定提交，校验 `FETCH_HEAD`，并保存 `.git/agent-platform-prepared.json` 的 repository/commit/run_id 身份。已有内容必须与标记匹配才能复用；不匹配时停止并核对既有副作用，而非删除目录重新准备。准备和发布使用各自凭据引用，Agent 和 tests 不继承 Git 凭据。检查凭据隔离时只报告环境变量是否存在，不读取值、不打印环境、不读取密钥文件。

本 Run `b0dfee0c70121cc2709e881fb94690e5` 的准备回执报告私有仓库 `big91987/model-relay` 准备成功，固定提交为 `9e5531b1149e00748013f25e9672bd5dd6041dc4`，`reused=false`，退出码 0。研发实读 HEAD 与身份标记一致。此证据不独立证明入站时远端 main 的实际 SHA；该一致性仍需协调者核对正式入站及 GitHub 快照，不能以此时重新查询 main 代替历史证据。

`.harness/agent-platform/SOURCE.json` 记录导入源 `f4ded463db0e401d5ad242b53cc1f3adb3fdab56` 及文件摘要。它是仓库内模板来源，不是本次 repository_commit，也不是平台运行二进制版本；三者必须分别记录。

## 原生角色交接与独立返工

流程顺序为准备 → 研发 → 固定 tests → 独立 QA → 人工确认 → 报告 → 草稿 PR。仓库模板声明 Agent 使用原生 Codex / `gpt-6.1-sol`，节点有各自持久会话；tests 是执行节点，固定执行 `make verify`，不是 Agent 会话。实际模型和会话身份应以平台原生记录核验，模板配置不能代替运行证明。

所有角色必须实际读取同 Run 的前序文件，交接传递仓库相对路径，不依赖口头完成声明。研发使用原生 `handoff` 交 `tests`；固定测试失败返回研发，独立 QA 对失败或缺证据通过正式 `handoff` 交 `development`。研发接到返工后先核对现有文件、历史回执和版本，只补当前 Issue 范围，不重建工作区或需求设计。补齐后重新经过完整固定 tests，再由独立 QA 检查同版文件。

本任务授权的故障注入要求首轮文档遗漏一个指定章节，保留该轮快照及实施记录；QA 应读取实际文件、指出缺项并正式返工。不得由研发自称 QA 已拒绝，或用普通聊天代替 QA handoff，也不能为该注入另开产品修复 Issue。

## 固定测试与原始证据

固定 tests 入口是项目现有 `make verify`，不得缩减为静态文档检查或删除门禁。`Makefile` 与 `tests/verify/main.go` 定义隔离候选目录、格式检查、模块校验、vet、前端测试、Go 测试、构建和浏览器验证。真实 HTTP/UI 验证必须使用实际 Go 服务，静态页面检查不证明后台工作正常。

验证入口输出 `verification_receipt`，实际证据在 `test-results/verification/` 下的运行目录；固定节点平台回执与该目录日志/receipt 应同时保留。逐次记录命令、版本、退出码及限制，不用历史产品日志替代本 Run 结果。平台早前命令日志用 `read_command_output` 按 seq/offset 读到 eof；缺失或截断不是通过证据。

原测试若因环境失败，保留首次完整失败、受影响门禁和具体错误，交研发定位项目问题，或由正式消息报告 Harness 环境缺口。不能反复求绿、改库数据、删门禁或把环境失败冒称产品缺陷。本 Run 固定 tests 执行 3 已失败：准备工作区为深度 1 的浅克隆，缺少历史 HTTP 门禁所需的基线提交，`tests/review-history` 报 `baseline unavailable: exit status 128`，外层 `make verify` 退出码 2。首个历史门禁未完成，后续门禁未执行，不能据此判定产品不通过或完整测试通过。原始日志和状态原样归档于 `harness-output/tests-execution-3/`，定位与恢复前提见本 Run 的 `development-execution-4.md`。需要由有权的平台准备/协调入口补充历史 Git 对象，在保留当前 HEAD、文件及身份标记的前提下核验后再测试；Agent 不自行取用凭据、重新准备目录或删除历史门禁。当前没有独立 QA 拒绝、人工确认或发布回执。

## 重复输入与未知执行结果

入口以 GitHub repository 与 Issue 的稳定 ID 构造 request_id 找回原 Run；Owner 评论按原 comment ID 构造消息 request_id，重复输入核对并复用原回执。机器人输出不是用户输入，修改旧评论也不是新指令。新补充应在原 Issue 新增评论，不能为重试新建替代 Run；已完成 Run 不复活。

任务受理、评论通知、Git push 和 PR 创建结果未知时，先通过正式平台入口及外部可见回执核对原请求和已执行副作用。核对前不得盲重发评论、重推送、重建 PR 或自造状态。状态查询失败只说明未知，不说明动作未发生。保留原请求标识和平台记录，由协调者执行需要的安全恢复；本研发节点不执行这些外部动作。

## QA、人工确认与草稿 PR 查看

独立 QA 读取本版文档、实施记录、完整固定 tests 回执以及前轮拒绝/返工证据；通过后保存 `harness-output/qa.md`，写清平台已验、未验及产品未验范围，再交人工确认。人工确认必须等待正式审批结果，不能以自动交接代替批准。

报告角色在批准后生成 `harness-output/report.json`，其 files 只列本版经 QA 确认的仓库相对文件，正文关联真实 Issue、版本、测试与限制。发布 Connector 保存 QA 原文字节和报告正文到 `docs/harness/run-<Run ID 前12位>/qa.md`、`report.md`，提交审核文件并创建真实草稿 PR。查看发布回执中的 PR 链接和该 PR 文件；没有真实发布回执时不填写虚构 URL，也不宣称 PR 已创建。草稿 PR、人工合并、部署是三个独立结果，本流程不自动合并或部署。

GitHub Issue 应能看到真实节点、原生回复、审批/执行卡点及处理入口，由 `big91987-agent-platform-bot[bot]` 回写。作者、Run/会话链接和通知去重需协调者以实际 GitHub 评论核验；本文不把安装说明当作本次已发送通知的证据。

## 证据边界与可追溯版本

运行页核对冻结 Workflow 版本、节点执行号、原生会话和执行回执；GitHub Issue 核对通知及处理入口；草稿 PR 核对实际提交文件与 QA 原文。三处证据互相对应，不靠“节点结束”推断通过。

当前证据包括用户提供的准备回执、研发实际读取的固定工作区，以及固定 tests 执行 3 的原始失败记录；完整固定测试尚未通过。远端 main 快照、入口事件去重、各角色实际模型、QA 拒绝、正式审批、通知、PR 和生命周期验证仍待对应角色提供真实回执。本文不声称供应商连接、计费、生产部署或商用能力已验证。
