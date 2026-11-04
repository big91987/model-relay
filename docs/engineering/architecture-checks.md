# 架构检查

执行 `make architecture-check`；完整 `make verify` 的 verify-gates 同时包含这项检查。GitHub Architecture checks 对 PR 和主线运行同一工具，展示结果并保留同版 JSON/Markdown 附件。

architecture.json 声明 cmd、internal 中的Go与JavaScript生产代码，测试单列排除。当前启动包只依赖relay；relay和控制台分别归属。relay仍是一个集中业务包，控制台仍是原生JavaScript，本轮没有完成领域分包或Vue迁移。检查通过只证明已声明的依赖规则满足，不能作为领域设计或独立QA通过。

报告位于忽略的 harness-output/architecture/，每次独立保留；固定latest-report文件供平台逐步骤归档。退出码0为规则通过、1为违规、2为分析未完成。未完成不可降级为通过；工具失败由Harness维护源修复。

工具唯一维护源为 Agent Platform examples/arch-template，安装清单 SOURCE.json 保存版本与逐文件摘要。首次使用该维护源的 scaffold.py --install-tool PROJECT；升级使用 --update-tool PROJECT，拒绝本地源码漂移，重新安装锁定npm依赖后验证。不在本仓手工修补已分发的工具。

初检：当前范围29个生产文件、排除35个测试、资源4件，声明规则0违规。26个Go文件仍主要集中在relay，3个JavaScript文件尚无feature边界。模型语义评审和真实产品QA未执行；后续领域拆分时同步收紧规则与验收。

维护源固定提交：`0308f6e`（工具0.3.1）；具体工具字节以 SOURCE.json 为准。平台侧后续归档和通知修复不会在本仓留第二份工作流维护源。
