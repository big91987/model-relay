# 架构检查

执行 `make architecture-check`；完整 `make verify` 的 verify-gates 同时包含这项检查。GitHub Architecture checks 对 PR 和主线运行同一工具，展示结果并保留同版 JSON/Markdown 附件。

architecture.json 声明 cmd、internal 中的Go与JavaScript生产代码，测试单列排除。当前启动包只依赖relay；relay和控制台分别归属。relay仍是一个集中业务包，控制台仍是原生JavaScript，本轮没有完成领域分包或Vue迁移。检查通过只证明已声明的依赖规则满足，不能作为领域设计或独立QA通过。

报告位于忽略的 harness-output/architecture/，每次独立保留；固定latest-report文件供平台逐步骤归档。退出码0为规则通过、1为违规、2为分析未完成。未完成不可降级为通过；工具失败由Harness维护源修复。

工具唯一维护源为 Agent Platform examples/arch-template，安装清单 SOURCE.json 保存版本与逐文件摘要。首次使用该维护源的 scaffold.py --install-tool PROJECT；升级使用 --update-tool PROJECT，拒绝本地源码漂移，重新安装锁定npm依赖后验证。不在本仓手工修补已分发的工具。

初检：当前范围29个生产文件、排除35个测试、资源4件，声明规则0违规。27个Go文件包含1个启动文件和26个relay文件，2个JavaScript文件尚无feature边界。模型语义评审和真实产品QA未执行；后续领域拆分时同步收紧规则与验收。

维护源固定提交：`0308f6e`（工具0.3.1）；具体工具字节以 SOURCE.json 为准。平台侧后续归档和通知修复不会在本仓留第二份工作流维护源。


## 0.4.0 逐项模型评审

本仓通过维护源提交02204e2的标准升级入口安装0.4.0，工具字节以SOURCE.json为准。规则结果位于rules，模型判断位于model_review；模型不能覆盖确定违规。architecture.json额外采用Go SaaS、模型网关和前端场景的4项职责要求，加上内置7项共11项；这是本项目的目标架构要求，不代表代码已经符合。

需要模型评审时，按已安装工具README提供显式URL、模型及私有凭据环境，使用inspect_project.py --review-command标准入口。密钥不写入仓库或命令参数。材料范围和提供商发送授权由调用方明确配置；正常make architecture-check仍只运行确定规则，不隐式付费或调用外部模型。模型仅输出精确原文，工具定位唯一引用；scope_notes不导致incomplete，缺少必需覆盖或blockers则拒绝。原始材料、模型结果和JSON/Markdown报告均留在忽略的harness-output中。

0.3报告和已发生运行保持历史原样；旧评审协议不被0.4新执行入口接受。当前平台已冻结的旧图和Run未在本次升级中改写。本轮没有完成产品模块迁移或产品QA。
