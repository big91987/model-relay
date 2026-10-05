# Issue #8：/healthz HTTP 方法边界修复入口

## 目标、现象与预期

用户报告运行服务的 `POST /healthz` 返回 200。核对 `internal/relay/server.go` 的 `ServeHTTP`：健康检查分支仅匹配路径，没有方法限制；随后检查 degraded 状态及数据库 Ping，成功返回 status/version JSON。这与报告一致；本节点未发送运行时请求，不冒充复现成功。

- GET 保持现有 status/version JSON 和存储故障 503 语义。
- HEAD 与 GET 在相同健康状态下返回相同状态码且无响应体，包含存储故障场景。
- POST、PUT、PATCH、DELETE 等所有非 GET/HEAD 方法返回 405，Allow 明确为 `GET, HEAD`；存储故障不应把这些方法改判为健康查询。
- 不改其他路由、现有 Host 校验及无关安全边界，不引入供应商调用、不部署。

## 范围与验收依据

限于健康检查方法处理及相邻回归测试，必要时更新与此次修复直接相关的说明。已有 `internal/relay/relay_test.go` 提供真实 loopback HTTP 服务和隔离安装辅助函数；可沿用，不需要新存储、接口体系或跨模块架构。

开发先添加通过真实 HTTP 请求断言上述预期的回归，保留修复前在旧代码上失败的证据，再修复并保留通过证据。覆盖 GET JSON、HEAD 空体、非查询方法的 405/Allow，以及 degraded 和数据库不可用时 GET/HEAD 的一致性；测试使用临时目录及随机端口，不访问用户安装或供应商。仅 Recorder 测试不能替代真实 HTTP 回归。

固定完整验证入口为 `make verify`：现有 Makefile 包含代码/版本证据、格式、模块校验、vet、前端状态测试、race Go 测试、构建及实际 Go 服务浏览器旅程，不缩减入口。开发及独立 QA 以真实回执和代码版本一致性为准；失败或监听沙箱限制单独保留，必要时由既有 host-only 验证 Connector 执行，不扩大权限、不宣称未运行的检查通过。本节点仅判断起点，不运行验收。

## 起点、既有资料与后续授权

选择 `development`，跳过本次 requirements 和 design 阶段：用户明确要求直接开发，现象、预期、非目标、验证方法均已明确，实际实现支持局部修复，无需新增产品范围或架构决策；无需重复审批。

已有 `docs/workflow/requirements.md`、`design.md` 和原 task.md 属于首版 v0.1/Issue #1 的建设任务（其中仍有“尚无实现”历史描述），不是本次修复的需求或设计基线。本记录替换旧入口记录；不创建跳过阶段的占位文档，也不以历史验证日志证明本次通过。

前序 Connector 首次 fetch 超时（退出 1）；显式重试后准备成功（退出 0），已确认分支 `workflow/1382a96a9e697c58ba68fe3c3c26f34b`，入口 HEAD `9b6994ed80de2bdba79ff267419150dccaac1360`，Issue #8 已创建。不重复准备、创建 Issue 或执行 Git 发布操作。

继续沿用用户授权：开发完成后交独立 QA，实际执行阶段分别更新 delivery.md、qa.md、pr.md 并明确为 Issue #8；在本 Issue 回写真实证据并由 Connector 提交草稿 PR。Git 提交、推送和草稿 PR 属于后续 Connector，本节点不执行；不部署、不启动 GitHub Workflow。
