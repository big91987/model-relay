# 独立 QA 核心旅程执行契约

日期：2026-10-06；模式：本轮首次全量验收，范围为 Issue #12 薄切片及固定 make verify 全回归，不扩展路线图。测试前冻结的权威基线为本 Run product-definition.md（US1—US8 / AC01—AC13）和 password-visibility-architecture.md。不将实现作为规格。代码身份见 evidence/snapshot.log。

## 连续 UI 旅程

所有 UI 旅程使用 tests/browser/journey.mjs 的同一管理员角色、同一隔离浏览器 context 连续完成，分别覆盖 1280px / 390px。Start State：Go Harness 通过正式 relay.Init 初始化临时空数据库及随机管理员凭据，启动真实 Go 服务；测试专用受控 upstream 仅为夹具，不是供应商验证。脚本生命周期接口只用于故障准备、日志排除对账；产品操作在浏览器完成，不用 API 预制管理对象补中间步骤。每个 viewport 关闭 context；Harness 关闭服务并移除临时目录。

| Journey ID | User Story/AC | Actor | Start State | UI Steps | Expected UI Result | Final Fact | Cleanup | Result |
|---|---|---|---|---|---|---|---|---|
| J01 | US1/AC01—02 | 管理员/未登录 | 空安装、登录页、无持久对象 | 刷新；空值和含空格特殊字符的输入；反复显示隐藏 | 初态隐藏、值不变、空值仍为空 | 同一 input 的 type/value 一致 | 清空输入 | Pass（同轮宿主 Connector） |
| J02 | US2—3/AC03—05 | 同一管理员 | J01 后 | Tab 到按钮；Enter/Space/鼠标切换；Tab 到登录 | 焦点可见且保留；动作名称、AX 状态一致 | submit/invalid/所有请求增量为零 | 清空输入 | Pass（同轮宿主 Connector） |
| J03 | US4/AC06 | 同一管理员 | J02 后 | 错误密码主动登录；隐藏态 Enter 登录；退出；显示态按钮登录 | 错误可恢复，两态成功进入真实管理台，秘密清空隐藏 | 真实认证与管理数据返回，不拦截成功登录 | 旅程继续、最终退出 | Pass（同轮宿主 Connector） |
| J04 | US5/AC07—08 | 同一管理员 | 登录成功 | 有效会话刷新；退出后输入与刷新；服务重启后刷新并重新登录 | 恢复管理台；失效/退出回隐藏空登录；重试成功 | 会话失效返回 401，重新登录后管理 API 200 | 关闭 context、服务、临时目录 | Pass（同轮宿主 Connector） |
| J05 | US6/AC09 | 同一管理员 | J01 到 J04 生命周期 | 比较切换前后 URL/storage；登录/退出后检查；操作记录与日志排除对账 | 无密码/偏好持久化或敏感内容展示 | 仅主动登录发认证，合法 HttpOnly 会话 cookie 仍允许 | 同 J04 | Pass（同轮宿主 Connector） |
| J06 | US7/AC10—11 | 同一管理员 | 首次空管理台，随后存量对象 | UI 配置 upstream、模型、key；调用/流式/取消/错误恢复；查看记录；版本加载/刷新/错误重试/迟到隔离 | 首次配置与日常使用均完成；布局可达；版本正确或未知可恢复 | 真实 Go API、隔离 upstream、持久化重开和日志事实对账 | 停用测试 key、退出、关闭临时环境 | Pass（同轮宿主 Connector） |

US8 / AC12—13 为工程与证据治理，不造 UI 旅程；以固定门禁、代码摘要、哈希和差异核验。没有新增多角色、迁移、异步控制器或组织运营能力，相关平台级场景不适用；密码同步反向操作覆盖，现有版本异步竞态仍回归。真正供应商、部署、屏幕阅读器音频和设备密码管理器不在承诺范围，均不声称已验证。

结果在 acceptance-matrix.md 给出最终 Pass / Fail / Blocked / Not Run；本表是执行前契约，不是通过证据。
