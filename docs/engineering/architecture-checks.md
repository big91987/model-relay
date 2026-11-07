# 架构检查

本仓安装架构检查工具 0.4.1，唯一维护源为 Agent Platform 的 `examples/arch-template`，来源提交 `0540677ca693896b58ecf8afdad0eceda033a6f1`（包含维护源 PR #8 的工具及安装清单修复）。工具通过标准 `scaffold.py --update-tool PROJECT` 升级，`SOURCE.json` 保存版本与逐文件摘要；不在本仓手工修改已分发的工具。

## 执行与报告

运行 `make architecture-check`。完整 `make verify` 的 `verify-gates` 包含同一检查；GitHub Architecture checks 在 PR 和主线执行，并展示及保留同版 JSON、Markdown 报告。

默认只运行确定规则，不隐式调用外部模型。报告位于忽略的 `harness-output/architecture/`，每次独立保留，`latest-report` 文件用于固定归档入口。退出码 0 为声明规则通过，1 为违规或需复核，2 为检查未完成；未完成不当作通过。

`architecture.json` 声明 `cmd`、`internal` 中的 Go 与 JavaScript 生产代码，测试从依赖规则排除。当前启动包依赖 `relay`；`relay` 依赖独立 `routing` 与 `accounting` 包，两个规则包禁止 HTTP/SQL 和第三方库依赖。控制台为原生 JavaScript，工作台仅通过公开入口依赖共享传输与调用测试功能，其他管理页仍集中。通过当前规则不代表已经符合目标职责划分，也不代表产品验收通过。

## 模型架构评审

需要模型评审时，按已安装工具 README 显式配置服务 URL、模型和私有凭据，运行 `inspect_project.py --review-command`。密钥不写入仓库或命令参数。材料发送范围和提供商使用由调用方明确授权。

内置 7 项及本仓声明的 4 项场景要求，共 11 项检查。规则结果与模型判断分别报告；模型只能提出候选架构问题，不能清除规则违规。工具检查职责、依赖、规则归属、状态和抽象，不判定产品业务计算结果。报告列出具体位置、解析器识别的函数或类型、依据及调整建议，保留模型原文。旧版历史报告保持原样，不作为当前工具已执行的证据。

模型格式、引用及语义仍有波动，不能把一次接受报告当成稳定性证明。0.4.1 的真实复验有结构和歧义引用拒收记录，见维护源[治理范围核对](https://github.com/big91987/agent_platform/blob/9fa9f7a/docs/03-delivery/architecture-tool-governance-review-2026-11-06.md)。

## 安装升级

首次采用维护源的 `scaffold.py --install-tool PROJECT`，后续使用 `--update-tool PROJECT`；本地工具源码漂移和运行中的检查会阻止升级。升级后重新安装锁定 npm 依赖并运行项目检查入口。模型评审、产品测试和平台工作流执行分别验收，工具初始接入不包含产品模块重构或产品 QA。后续整改记录见 [工具驱动整改](architecture-remediation-2026-11-07.md)，本轮重构的回归不能代替独立产品 QA。
