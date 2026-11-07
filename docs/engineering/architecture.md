# 代码职责与依赖

仓库为一个 Go module。功能归属由 Go 包约束，不能以文件名代替模块边界，也不要求所有模块套用同一套四层目录。

| 位置 | 当前职责 | 允许的项目内依赖 |
| --- | --- | --- |
| cmd/model-relay | CLI、配置读取与入口组装 | relay |
| internal/relay | HTTP 协议适配、请求/尝试协调、当前配置读取和 SQLite 操作 | routing、accounting |
| internal/accounting | 金额计算、预留、资源限额及比较规则；输入普通值并返回决策 | 无；禁止 HTTP、SQL 与第三方库 |
| internal/routing | 候选选择、权重、探索、冷却、延迟样本与其运行状态 | 无；禁止 HTTP、SQL 与第三方库 |
| internal/relay/web/workspace.js | 管理页组装、人员及租户范围、编辑与导航 | shared/transport、features/playground 的公开入口 |
| internal/relay/web/shared/transport.mjs | 范围绑定的 JSON/下载传输、请求取消与过期响应拒绝 | 无业务功能依赖 |
| internal/relay/web/features/playground | 应用 Key 调用表单、流式接收、停止与退出清理 | 无其他功能依赖 |

## 路由操作

网关读取当前配置并将候选事实传入 routing.Engine.Select。Engine 持有权重、探索、健康、样本和派发序列；调用方不能直接读写这些 map。观察时间是明确输入，便于复现窗口边界，不从数据库或 HTTP 请求中获取。

Engine.Observe 接收由网关校验配置代次后的结果。RecordLatency 保留“证据保存成功后才采用样本”的顺序，配置核对和证据持久化由适配操作提供；失败直接返回，不补造证据。配置变化通过 Reset 或 ForgetCandidate 失效，不由控制台重建内部 map。

原有共享配置锁仍保护请求与配置的协调。本轮不以结构调整声称已消除锁竞争；会话与上游检查状态尚未整体拆出。

## 规则与尚未拆分的范围

architecture.json 是可执行的当前约束；新增生产包必须归属模块并声明依赖。路由与资源规则不得依赖 net/http 或 database/sql。根索引 AGENTS.md 引用本页，模型评审读取同一份规范。

relay 内的控制台 HTTP、应用操作与 SQL 仍有混合。这属于后续分步治理范围，不能因当前导入规则通过而宣称所有 SaaS 领域分层完成。简单查询和 CRUD 不为形式增加空层，复杂操作的事务边界必须保持。

## 浏览器功能

workspace.js 通过公开入口创建调用测试功能；功能只接收结果通知与查看记录回调，不访问工作台全局 state。离开页面或切换人员/租户时，工作台调用 dispose 取消调用并清除 Key、消息和输出。共享传输模块持有管理请求的 AbortController；每次响应绑定发起时的人员与范围，过期响应不能写回。

其他管理页仍集中在 workspace.js，旧版 app.js 也未重构；本轮只验证调用测试这一条功能边界，不宣称前端已整体模块化。
