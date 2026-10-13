# Model Relay 的独立 OpenSandbox Harness 入口

本仓库复制并维护 `.harness/agent-platform` 内的 SDK、样例及工程工具。GitHub Issue 入口直接执行这些文件；Runner 不再查找另一份 Agent Platform checkout 或使用其虚拟环境。`SOURCE.json` 记录初始固定源版本和文件摘要，项目调整由本仓库 Git 记录，升级需审阅差异。

## 两条入口

- 原产品研发入口保留原配置变量 `AGENT_PLATFORM_WORKFLOW_CONFIG`、Owner 身份检查、Mac ARM64/he-full Runner 和每 Issue 串行策略。已有 Issue/Run 沿用原流程。
- 新验证入口由标题 `【Harness验证】具体目标` 和标签 `harness-opensandbox` 共同选择，使用独立 `AGENT_PLATFORM_SANDBOX_CONFIG`。新标签被旧入口排除；手动 dispatch 也由适配器读取当前 Issue 再核验，不能把新任务送入旧本机流程。
- 新入口把固定 repository_commit、用户目标和原始 GitHub 关联提交给平台，要求 `expected_environment=opensandbox`，不在 Runner 创建 Agent 工作区。平台中的执行节点负责准备同 Run 的 `/workspace`，各角色共享该工作区、保持各自会话。
- 初次和重复投递使用同一 GitHub Issue 请求键；已受理后不因 Issue 编辑改变原启动输入。评论属于原 Issue、由 Owner 发出时才送到现有会话，不能创建替代 Run。

## 新入口的启用条件

本次只安装入口与仓库自有工程文件，**尚未创建新的验证 Issue、安装实际 Agent Pipeline 或验证原生模型**。创建验证任务前须完成：

1. 112 隔离平台的原生 Codex/gpt-6.1-sol、工作区权限、角色交接、停止与恢复实际验证。
2. 通过正式 API 安装独立 OpenSandbox Workflow，保存准备、Agent、固定验证、独立 QA/报告和人工确认路径；不复用旧产品 Run。
3. 在 Runner 的私有目录设置该 Workflow 的配置与 API 凭据文件，再设置仓库变量 `AGENT_PLATFORM_SANDBOX_CONFIG` 指向配置文件。配置中的 `repository` 为 `big91987/model-relay`，`issue_label` 为 `harness-opensandbox`，`base_url` 指向独立实例，`workflow_id` 为正式安装回执，`base` 为 `main`，`token_file` 指向私有凭据。实际路径、token 和供应商凭据不得进入仓库或 Issue。
4. 确认 Runner 能访问隔离实例后，再由 Owner 创建唯一新 Issue 并附此标签。缺条件的入站失败或任务受理不代表测试通过。

验证应覆盖真实独立角色与共享文件、交接、固定测试、独立 QA、返工、重复投递、停止/接续、外部失败及报告发布。产品实现与产品验收仍交对应 Pipeline Agent，本次入口迁移不改 Model Relay 业务逻辑。已有产品任务、部署入口、账户及数据不在本次范围。

## 项目维护

更新工程文件时从明确源版本导出到新目录，审阅并合并本项目调整；不要覆盖已有定制。修改入口后更新对应 `.github/agent-platform*-entry-source.json` 的实际文件摘要。原配置变量及在途任务保持，配置缺失先补齐条件，不自动重建任务或放宽权限。
