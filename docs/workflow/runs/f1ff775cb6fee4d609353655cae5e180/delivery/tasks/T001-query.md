# T001 当前配置和同请求可核对

- 当前状态：**Done（本切片）**。第二轮QA核对同快照seq33真实HTTP/两屏最终事实及独立3Go race契约，退出条件满足；见[当前报告](../../acceptance/report-02.md)。以下状态均属历史；未宣称正式发布。

- 独立QA当前：seq18宿主当前Go HTTP及独立3项race契约通过，仍In Review，本切片不Accepted。下方“真实HTTP待宿主”是旧记录。同ID概览动作由T002/T003返工；见[报告](../../acceptance/report-01.md)。

- 状态：In Review；里程碑 M01；负责人研发；评审项目测试/独立QA；执行 TDD。2026-10-06领取，基线 d62cd51e8de2127548e1ac7469d349443d1bfe3d。
- 硬依赖：G2；冲突范围 internal/relay/server.go、proxy.go、store.go 及新增 console.go。
- 权威输入：[架构§4–5](../../console-architecture.md)，PRD AC-04/05/13/14/16。
- 目标：SQL全量计数、当前启动当前配置持久success证据、精确ID查询、拒绝安全归属与expires_at。
- 非目标：新表、持久核验实体、多上游、扩权。
- 验收：超过1000计数；变更/迟到/重启失效；拒绝401统一且零转发；未知归属空；查询权限/404/503。
- 验证：隔离Go测试及 make verify；证据 ../verification/development.md。
- 实施结果：console.go聚合/精确查询/内部摘要识别；server.go成功写入配置代次与会话expiry；proxy.go拒绝落盘和成功候选隔离。无迁移、无鉴权放宽、无外部协议变更。
- 已验证：TestConsoleOverviewAndExactRecords（1006对象、真实0/null、权限与精确边界）、TestConsoleEvidenceConfigurationRace（全部配置写入/迟到/重启/终态修正）、TestConsoleDeniedMetadataAndStorage（安全拒绝/过大正文/并发拒绝/存储失败）；Go race专项通过。见[实际验证](../verification/development.md)。
- 当前可依赖：上述管理JSON契约与实现；未完成：真实HTTP/固定入口和独立评审。下一步宿主运行现有HTTP全量回归及TestConsoleRealHTTP；不标Done。
