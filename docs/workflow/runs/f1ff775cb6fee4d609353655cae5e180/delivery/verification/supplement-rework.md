# 维护者补充：工具摘要与会话检查

2026-10-06，本Run研发补齐，T002/T003仍In Review，M02/M03仍Ready for Review。结论：分项检查通过；**固定门禁Blocked by Validation（沙箱禁止监听）**。不是独立QA、全量产品、真实供应商、PR/合并/部署通过。

本文件保留第一次补充的实现快照与证据。维护者后续发现的悬挂检查已知到期缺口及正常提示问题，在同Run第二次补充修正；**当前对象和最新证据以[expiry-watchdog.md](expiry-watchdog.md)为准**，下述48VM与指纹不是最新对象。

## 契约与根因

不另建Issue/Run，不修改已接受[PRD](../../product-definition.md)或[架构§6](../../console-architecture.md)/[交互](../../console-interaction.md)来取得绿灯。

| 边界 | 首轮实际缺口 | 当前实现与回归 |
|---|---|---|
| AC-10合法工具结果 | 仅提取content，tool-only测试响应不完整且只断言无文本 | 普通message.tool_calls；SSE choice/tool index增量名称在DONE后汇总。数量+前5名称、每名128字符、最多1024身份，超限明确“至少”。不执行工具、不以arguments伪造文本；保留空内容、未知用量及原始诊断。普通/流式合法结构断言真实摘要、结束前无摘要、无额外请求。 |
| AC-02/12及架构§6会话 | 60秒轮询；非401失败仅提示，秘密和请求未清 | 可见生成30秒；已知过期更早，提交前也检查边界。网络/不可读响应失效会话异步所有权、abort/cancel、清除全部秘密及输出，保留安全目标/模型/已取得ID；发送锁定、停止自动检查，不称401。仅显式检查成功/登录恢复，不自动重发。管理401登出；应用401不登出。 |
| 迟到与恢复 | 网络失败未隔离尚在读取/签发的结果 | 单飞session检查；旧JSON/SSE/read/finally/401/签发不污染新状态。VM覆盖长流网络失败、30秒定时、迟到增量、重新检查、已知过期、401/登录、重新登录入口、隐藏页跳过轮询和旧检查隔离。 |
| 产品事实 | 所有用户产品概览固定声明供应商未测 | 改为保存/当前成功证据边界和费用/记录说明，不伪造供应商结果。本轮真实供应商Not Run继续在本报告保留。 |

可读与原始展示独立有界256Ki字符；工具摘要参数不执行，名称按textContent显示。Trellis真实规范原位维护在`.trellis/spec/frontend/console-lifecycle.md`。此次未改变后端协议、表、权限、上游结构、限流或多路路由。

## 第一次补充对象与真实回执

HEAD仍`d62cd51e8de2127548e1ac7469d349443d1bfe3d`，未提交；当前未提交源码摘要见[source-rework-02.json](source-rework-02.json)：`57c6ead69e3a79d749f49757a04c94b07639603fbe8fdf026d99849700e78c1d`，56文件，Go1.25.6，与[verify-attempt-06.log](verify-attempt-06.log)首部一致。新增证据不覆盖首轮失败、分项结果或历史日志。

| 实际命令 | 结果 | 证据 |
|---|---|---|
| make verify，补充中间及最终各一次 | 均退出2；可信历史HTTP门禁因`listen tcp 127.0.0.1:0: bind: operation not permitted`严格REJECTED，未进入后续固定阶段 | [05](verify-attempt-05.log)、[06最终](verify-attempt-06.log) |
| make format-check vet frontend-test build | 中间46VM、最终48VM全通过，Go格式/vet/build均退出0 | [04](checks-attempt-04.log)、[05最终](checks-attempt-05.log) |
| node --check app.js/journey.mjs；git diff --check | 最终均退出0 | [checks-attempt-05.log](checks-attempt-05.log) |
| go run ./tests/browser，补充中间及最终 | 均退出1，受控上游监听受限，未启动Chromium旅程或产生截图 | [02](browser-attempt-02.log)、[03最终](browser-attempt-03.log) |

中间[source-rework-01.json](source-rework-01.json)不是最终验收对象。开发快速测试曾发现断言将“并非已过期”中的词也误判为认证失败；改为明确“网络异常不等于认证失效”的产品反馈，保留“不出现已过期”和“发送被锁定/秘密清除/无重发”的强断言。没有以宽松测试代替实现。

## Trellis-check与宿主交接

- 实际packages显示单仓backend/frontend；恢复适用索引、生命周期、验证及复用规范。已有真实初始化/源码规范，不伪造bootstrap。
- 复用既有clearSecrets、stopCall、owner/current、api和原Go浏览器入口；无运行时依赖、无调试日志、无平台补丁，无inline HTML执行。Makefile及历史健康日志/历史task/requirements/design与HEAD无差异。
- 原生Go两屏旅程新增合法普通/工具增量夹具及同IDsuccess未知用量记录；工具流由测试控制端释放，先断言未完成无假文本再释放DONE。网络故障仅拦截测试session检查，不预填配置，不修改产品认证或数据库；真实30秒长流等待后验证停止/清密，再解除网络故障并读取真实session手动恢复。供应商夹具仅是受控测试。
- 下一节点在同一源码原样执行make verify，保留新回执；特别核对工具流释放、30秒长流故障、48VM和全部原HTTP/race/浏览器门禁。真实浏览器尚未执行，不能从VM或Go编译推出布局、键盘、AX/会话/协议验收通过。
- 独立QA必须逐条核对PRD全部18AC和架构/交互边界，并使用首轮[18AC映射](development.md)及本补充，不以通过用例数替代验收。截图只在秘密/输出清除后归档。
- **真实供应商联调Not Run**：本轮无真实供应商凭据及响应证据；不得将受控夹具称为真实供应商成功。路线2–4尚未实现，按项目唯一roadmap接续；当前切片不代表产品/平台全量通过。PR/合并/正式测试部署/部署后体验均尚无回执。
