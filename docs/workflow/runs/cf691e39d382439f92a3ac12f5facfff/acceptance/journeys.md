# 短轮复验旅程与范围

基线为用户 Issue #14 原始委托、本 Run T001 AC、既有运行契约及历史 QA12-HISTORY-WRITE；预期是固定验证不改写历史、当前输出可定位、异常非零且完整门禁保留。实现不反向定义预期。

采用短轮复验：从 prepare HEAD 仅修改 health-history 输出、回归、README 与 Trellis 运行契约，产品、UI、认证、数据、锁文件和 Makefile 未变。已有全量旅程仅恢复执行契约，不拿历史 Pass 替代本轮；本轮宿主实际重新执行全部固定门禁。无新增产品首次授权、异步执行、权限或供应商能力，相关独立新矩阵裁剪；既有能力仍在公共门禁中。

| Journey ID | User Story/AC | Actor | Start State | UI Steps | Expected UI Result | Final Fact | Cleanup | Result |
|---|---|---|---|---|---|---|---|---|
| H01 | 默认历史保护/新输出 | 验证调用者/隔离 checkout | 已跟踪历史清单与同版源码 | CLI 执行 make verify，调用方指定新输出；无新增 UI | stdout 有版本、历史分类与后续门禁回执 | 宿主 exit 0，106 文件不变 | 程序移除临时数据；不恢复旧文件 | Pass（同轮宿主；历史独立对账） |
| H02 | 异常不吞退出码 | QA/隔离 fixture | 正常、rejected、closed_stdout 三情形 | CLI 子进程，不涉及 UI | 成功 receipt；失败 stderr 定位 | expected_red exit 0，其余 exit 1，历史哨兵保持 | t.TempDir 清理 | Pass（QA 实际 race） |
| H03 | 固定入口真实失败 | QA/受限环境 | 沙箱不能监听 | 原样 make verify 重定向本 Run 新日志 | 不把环境失败当旧版预期红灯 | REJECTED、make exit 2，历史不变 | run defer 清理 | Pass（失败传播）；成功网络路径为 Blocked |
| UI01 | 既有核心 UI-E2E 公共门禁 | 同一管理员/同一隔离 context，每 viewport 一个会话 | relay.Init 空数据库；受控上游仅测试夹具 | 密码可见性/键盘/两态登录→管理台版本→上游/别名/令牌创建编辑→调用/流式→记录→失败恢复→重启重新登录与持久事实→停用/撤销；分别 1280/390px | 用户从真实 Go 页面完成操作与恢复，无隐藏 API 补产品步骤 | journey 断言真实请求、记录、秘密排除、持久状态；版本接口独立核验 | browser context 关闭，Go 服务/数据库/临时目录关闭移除 | Pass（同轮宿主总回执层，细粒度输出截断） |
| UI02 | QA 亲自重复上述 UI-E2E | 同一管理员 | QA 固定入口在历史监听阶段受阻 | 未进入浏览器 | 不可声称页面验证成功 | 无本地 UI 事实 | 无启动服务遗留 | Not Run；同版同轮 UI01 为发布门禁依据 |

UI01 的正式旅程为 `tests/browser/journey.mjs`，Go harness 为 `tests/browser/main.go`。测试生命周期 restart/audit 仅用于故障准备与秘密排除对账；管理对象与产品操作由页面完成，不是 API 预制完成态。不存在供应商 credentials/实际响应，本报告不声称真实供应商验证。
