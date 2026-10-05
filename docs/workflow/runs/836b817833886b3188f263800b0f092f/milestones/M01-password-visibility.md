# M01 管理员能够安全核对密码并继续管理

- 状态：Ready for Review
- 负责人：研发节点；评审：独立 QA
- 硬依赖：本 Run G1／G2 已就绪
- 范围：PRD AC01—AC13；任务 T001

## 进入条件
需求与可访问、生命周期契约明确；Go 服务、原生 UI、make verify 和真实浏览器 Harness 已存在。Trellis 已初始化，空白规范需在实现前源码 bootstrap。

## 退出条件
原生切换保值、不提交、可键盘操作且名称／状态同步；秘密清理与真实登录、会话恢复、退出、管理台、健康和版本回归通过；历史文档保留；固定门禁真实通过并由独立 QA 核验。只有这些退出证据成立才可 Accepted。

## 集成与验证
使用原有 make verify，Node VM 快速反馈，真实 Go 服务在 1280px／390px 下验证原生交互。证据在 ../verification/，任务结果在 ../tasks/T001-password-visibility.md。隔离测试上游不代表供应商联调。

## 风险
沙箱监听受限需平台宿主 Connector；失败保留，不以局部通过冒充里程碑验收。显示密码存在旁观风险，只允许主动操作。

2026-10-06 研发切片与静态／非监听回归就绪，见 [研发验证](../verification/M01-development.md)。完整 make verify 受沙箱监听限制；退出条件中的真实 HTTP、浏览器与独立 QA 未满足，里程碑不是 Accepted。
