# Issue42 候选接续说明

本分支保存 Issue42 Pipeline Agent 已实际写出的两处项目测试设施修复，供新隔离环境继续原固定测试和独立 QA。它没有修改产品业务逻辑，也不是已通过验收的发布版本。

冷环境依赖下载的 stderr 曾混入历史回归 JSON stdout，解析失败；验证器现在分别保留并解析两路输出。路由 HTTP 测试曾假定下一请求一定发生在 15ms 冷却窗口内；fixture 现在明确建立未到期和已到期阶段，同时保留真实 HTTP、回退、冷却排除、恢复及期限检查。原始研发说明见本分支 docs/workflow/runs/e65c964ce9faf29e35271270d751cedb/ 两份修复报告。

源码由该 Run 的 development4、development6 产生，协调者从正式文件 API 读取并按原字节保存；本次只转存候选，没有代写测试或代 QA 放行。历史报告中的 harness-output 路径属于原 Run 的私有原始材料，未全部发布到仓库，不能当作本分支可直接读取的文件。

真实验证：验证器 9 项定向回归红→绿，原历史 HTTP 回归通过；路由精确 HTTP race 测试通过，忽略冷却、永不到期、期限减半三个变体均被拒绝。扩展路由集合出现 signal:killed，未通过；容器事后 memory.max=2Gi、oom_kill=1，但无起点计数，不能精确归属进程。两次原固定 make verify 退出2；最后一次在原1小时环境期限前正式停止，exit=-1。没有全量绿或独立 QA，不合并此草稿。

原 Run 已正式取消并释放。44份原材料共2528387字节已保存；新平台通过标准升级提供新 Run 的4小时/8Gi配置，旧 Run 不变。后继任务必须从本分支固定提交，经原 make verify、独立 QA 的缺章拒绝/返工/复核、正式确认及 Pipeline 自行发布同版草稿 PR。后继 PR 应包含本候选三份源码的最终审查结论，本草稿仅作接续材料。

关联：https://github.com/big91987/model-relay/issues/42

原文件校验：

- `tests/review-history/main.go`：`eaa988821fccf011d9fc63795ea5933f3f7acb734ffe22830cbff8f6eafcf355`
- `tests/review-history/main_test.go`：`ebdd0554efd7f73155324f9d5034e71a72fe1baaced6747c066485f5443a8bfa`
- `internal/relay/routing_http_test.go`：`0018956b4e78f792ad26b789b215e66c31a75bf12a7f683e4ba9eb2249ae8572`
- `docs/workflow/runs/e65c964ce9faf29e35271270d751cedb/harness-verifier-fix.md`：`e2bb231fb451ad07fed06ea1da632597e9a9e54ece6f4a56f0ec8dfec1592e1f`
- `docs/workflow/runs/e65c964ce9faf29e35271270d751cedb/harness-cooldown-fixture-fix.md`：`68c7df51b776ee067182c7e751051f72b23d4bcefcc3e86656302c47d579559c`