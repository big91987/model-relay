# 需求—旅程—状态—测试—证据矩阵

权威基线：本 Run product-definition.md、password-visibility-architecture.md。快照为 HEAD 73681a7d + 工作树摘要 0f6ee2d6894067323697f632231ee0d62b098be0261a4da7e975edb6e1efb37b。H = evidence/host-connector-receipt.md（平台 seq7 退出 0；截断边界已披露）；L = evidence/qa-non-listening-checks.log（QA 亲自执行）；S = evidence/snapshot.log、final-integrity.log。

UI-E2E / API-E2E 的 Pass 来源是同轮固定宿主门禁，而不是 QA 本地点击。固定 Makefile 串行 fail-fast 且 browser Harness 将 Node journey 退出码透传，因此宿主退出 0 能证明脚本断言全部完成；没有细项原始尾部日志，不填造单项耗时、截图或测试数量。VM 不代替真实键盘或网页旅程。旅程执行契约见 journeys.md。

| ID | 基线要求 | 角色/范围 | 起始状态 | 连续用户步骤 | 最终事实 | 证据类型 | 结果 | 证据位置 |
|---|---|---|---|---|---|---|---|---|
| AC01-UI | 首次/刷新默认隐藏 | 管理员/两个 viewport | 未登录、空初始化 | J01 初态和刷新 | 空值、type=password、动作/状态隐藏 | UI-E2E | Pass | H；journey.mjs passwordVisibilityJourney/passwordState |
| AC02-UI | 空值/特殊输入保值多次切换 | 同上 | J01 输入 | 反复 Enter/Space/click 切换 | value 完全一致，原 input 未重建 | UI-E2E | Pass | H；J01—02 |
| AC02-Contract | 输入节点与值稳定 | 未登录/VM | 空值/特殊输入 | 运行 app.js 同步 click | 节点 identity、value 和状态一致 | Contract | Pass | L；ui.test.mjs password toggle |
| AC03-UI | Tab、Enter、Space、可见焦点 | 管理员 | J02 | 密码输入 Tab→切换→登录；两种键激活 | 焦点保持、outline>=3px，原生 button | UI-E2E | Pass | H；J02 |
| AC04-UI | 清楚名称/状态 | 同上 | 隐藏/显示 | 每次切换读按钮/status AX snapshot | 显示/隐藏动作名与 type、说明一致；不混用 aria-pressed | UI-E2E | Pass | H；passwordState；仅 AX 语义，非音频 |
| AC05-UI | 不提交、不校验、不发请求 | 同上 | session 初始化完成 | 空/非空值鼠标与键盘切换 | submit=0、invalid=0、所有 request 增量=0 | UI-E2E | Pass | H；J02 |
| AC06-UI | 两态真实登录、错误恢复 | 管理员/真实 Go | 无会话 | J03 拒绝→隐藏 Enter 登录→退出→显示态登录 | 错误仍 401、可重试；成功管理台、输入清空隐藏 | UI-E2E | Pass | H；journey.mjs 两态 login；不 mock 成功 |
| AC07-UI | 退出后隐藏空态 | 同上 | 有效会话 | 退出、重新输入、未登录刷新 | 显隐状态与名称复位、新输入仍隐藏 | UI-E2E | Pass | H；J03—04、journey 末段 |
| AC08-UI | 有效恢复/失效回登录 | 同上 | 有效会话→退出或测试重启 | reload 恢复；logout/reload；重启后 reload/重登 | 有效恢复管理台、失效登录；架构约定 logout 失效路径明确验证隐藏空值 | UI-E2E | Pass | H；runtimeVersionJourney、logout/reload、restart |
| AC08-Contract | 管理 401 清理/恢复时清理 | VM/管理员状态机 | 显示态有输入 | session 恢复、管理请求 401 | signedIn/signedOut 清空且隐藏，版本失效 | Contract | Pass | L；ui.test.mjs restoration/real management 401 |
| AC09-UI | 不持久化/额外发送密码 | 管理员 | J01—J05 | URL/storage 比较；整个管理生命周期 | local/sessionStorage、IndexedDB、SW 无新记录；合法会话 cookie HttpOnly | UI-E2E | Pass | H；J05；不采集请求正文 |
| AC09-Data | 日志和记录不泄露 | 隔离实例/受控 upstream | 调用成功/失败 | UI 操作→records 文本检查→测试 control 日志对账 | 敏感哨兵排除、错误 upstream 不回显 | Data Reconciliation | Pass | H；audit-logs/record-list 断言；控制端只作事实对账 |
| AC09-Static | 无秘密副本和新记录 | 前端源码 | 当前 diff | 独立检查 type/textContent、onclick、清理入口 | 不新增存储、URL、网络、日志、后端变更 | Static Review | Pass | evidence/static-review.md；不替代 UI |
| AC10-UI | 真实管理/首次配置/窄屏无回归 | 管理员/1280、390 | 空库→存量 | J06 浏览器配置→调用/流式/取消/错误→重开/停用 | 首次和日常链连续完成，布局/操作可达 | UI-E2E | Pass | H；journey.mjs 现有完整管理链 |
| AC11-API | 健康方法/异常状态回归 | 无登录/HTTP | healthy/degraded | GET/HEAD/拒绝方法及恢复 | 既有方法、响应头和状态契约成立 | API-E2E | Pass | H；TestHealthMethodsHTTP；旧 baseline 故意失败非当前失败 |
| AC11-Contract | 健康/版本 handler | QA/当前源码 | 无监听 handler | race 下执行 handler/version 测试 | 方法矩阵、原始版本正确 | Contract | Pass | L；不是当前 HTTP-E2E |
| AC11-UI | 实际版本、未知、重试、迟到隔离 | 管理员 | 已登录 | actual /health、refresh、route 故障夹具、退出/返回 | actual dev 一致；未知可恢复；迟到不污染；长值安全 | UI-E2E | Pass | H；runtimeVersionJourney；故障 route 明确为夹具 |
| AC12-Host | 固定门禁完整真实运行 | 平台宿主 | 当前工作树 | 固定 make verify | 退出 0；摘要与 QA 相同 | Contract | Pass | H；平台 seq7，非历史 |
| AC12-QA | QA 独立固定入口尝试 | QA 沙箱 | 同源码 | make verify | 到 health-history 无法监听，退出 2；后续未运行 | Contract | Blocked | evidence/qa-make-verify.log；以 AC12-Host 同轮等价门禁满足发布条件 |
| QA-UI | QA 自己重跑真实 UI | QA 沙箱 | 本地门禁前置失败 | 未进入 browser 阶段 | 没有本地真实 UI 结果 | UI-E2E | Not Run | 同上；J01—J06 的宿主行分别提供同轮用户结果，不改此行 Pass |
| AC13-Integrity | 文档/日志保留、路径和来源正确 | QA/交付 | 10 文档+62日志基线 | 保存新历史重建、恢复旧路径、哈希/diff/产物检查 | 72 份历史哈希全 OK；新增契约产物位于本 Run；来源分层 | Static Review | Pass | S；host-health-history.log、qa-health-history.log；本轮 P3 副作用另列 |

所有强制 AC 有当前同轮证据；Not Run/Blocked 不计入本地通过率。未测生产、真实供应商、屏幕阅读器音频；不是本轮验收失败项。
