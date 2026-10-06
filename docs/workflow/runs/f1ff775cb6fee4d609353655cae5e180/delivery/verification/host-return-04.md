# seq12：登录拒绝完成边界返工

- 日期：2026-10-06；负责人研发；结论：同步缺陷已修改，待宿主固定门禁与独立QA，未验收通过。
- 对象：HEAD `d62cd51e8de2127548e1ac7469d349443d1bfe3d` 的未提交工作树；[56文件摘要](source-login-completion.json) SHA256 `37b9d52f1d2458e5333a4675648ca28c13a513af8eeb094a37120a46d78de861`。
- 权威范围：[PRD](../../product-definition.md) AC-03/08/17/18、[交互设计](../../console-interaction.md) 登录清除、[架构](../../console-architecture.md) 生命周期。未改这些契约或产品代码。

## 宿主失败与旧证据

seq11平台broken pipe导致回执缺失，结果未知，不能当成功或产品失败。维护者显式重跑同工作树；seq12 `make verify` 真退出码2，工作树摘要 `4387d20a8cfcbbbe57f01d67541a901de7d8d0279377cf6b21cf1d3a25c7208c`。输入提供首尾而非完整日志，不补造中间结果。

可观察尾部：当前Go包 `internal/relay` race/HTTP PASS（113.793s）、历史分类器测试PASS、build完成；1280px真实浏览器在 `journey.mjs:53:10` 密码隐藏断言失败。实际脱敏探针如下，不称浏览器通过：

```json
{"stage":"controlled input","activation":"click","activationCount":8,"expectedVisible":false,"type":"text","inputFocused":false,"toggleFocused":false,"events":{"submit":1,"invalid":0,"toggle":9}}
```

响应列表只有两次 GET `/admin/api/session` 401，尚未出现POST登录响应。旧 [host-return-03.md](host-return-03.md) 及所有历史回执保留。

## 根因和变更边界

stage与activation只在显隐循环中更新，之后一直保留。seq12不是“第8次显隐产生第9次click”的证明：循环结束后旅程还会点击一次“显示密码”再“登录”，恰好产生toggle=9、submit=1。真正失败处是随后错误密码登录清除断言，旧探针错误标记为上一个动作。

登录后旅程仅等待 `#error` 含“请”。启动session 401已展示 `管理会话已过期，请重新登录；不会自动重发调用。`，包含相同的“请”。`navigate('登录')`只保证click结束，不保证异步登录完成，因此旧提示立即满足等待，密码仍为text时隐藏断言失败。产品在真实登录401后才signedOut清除，此回执尚未到达这个边界，不应猜测改动产品显隐或鉴权。

- `rejectedLoginJourney(page)`使用测试专用路由屏障，只延迟后 `route.continue()`，不伪造Go响应。先验证请求已发起、密码显示/值不变、登录禁用、旧提示仍在；释放后等待真实POST响应401及完整特定错误，再保留原隐藏/清除、按钮与状态AX断言。两屏都执行。
- 显隐循环结束重置探针；拒绝、退出和刷新使用各自固定stage。原8次原生显隐、恰好一次激活、键盘焦点、零submit/invalid/network、URL/存储不变断言全部保留。首次配置全UI、HTTP/SSE和保密回归未删。
- 新增app.js VM悬挂登录回归：旧含“请”提示在pending时不变、密码保留、submit禁用、仅一个请求；模拟401后才清除/隐藏并展示专属错误。它证明生命周期与谓词碰撞，不替代浏览器。
- 只改旅程、VM测试、前端验证规范和公共文档；Makefile/锁文件未改，产品/协议/数据库/平台不变，无扩权。

## 真实验证与限制

| 检查 | 实际结果 | 证据 |
|---|---|---|
| 初次新VM | 退出1，58通过/1失败：误期望网络错误，实际为session401过期提示；修正测试及原生pending断言，保留失败 | [初次日志](login-completion-vm-initial.log) |
| 最终VM | 退出0，59/59通过，无跳过 | [最终日志](login-completion-vm.log) |
| JS语法、格式/vet/build、模块校验、diff | 各退出0，门禁及锁文件diff为空 | [分项日志](checks-login-completion.log) |
| 版本摘要 | 退出0，56文件、Go1.25.6、上述HEAD/摘要 | [当前指纹](source-login-completion.json) |
| 新对象完整make verify/Chromium | 本次Not Run；已稳定确认的沙箱监听/Chromium限制未变，完整验证明确留宿主 | [前次限制](host-return-03.md)、[首次失败](development.md) |

Trellis before-dev实际packages发现成功，读取真实frontend规范、生命周期和复用guide；初始化/规范已存在，bootstrap不适用。Trellis check核对新增路径、原状态/AX断言、安全白名单、锁文件/门禁不变；update-spec将异步拒绝完成边界和探针生命周期写入现有verification规范。无公共协议变化，不需要跨层设计返工。

公共交付审查沿用原编号、权威追踪、负责人和关键路径；5份更新文档的45个本地引用检查通过、文档修改后diff退出0。初次VM失败日志只将本机绝对堆栈路径脱敏为仓库相对路径，保留失败内容与退出码；没有覆盖历史证据。

## 公共状态与正式交接

T001/T002/T003仍In Review，M01–M03仍Ready for Review，阶段Verifying。关键路径：新指纹对象宿主原样 `make verify` → 真实失败正式返工，或独立QA逐条18AC/设计验收。不能复用seq12旧对象；仅宣称本地分项通过，未宣称原生回归通过。

保留全部历史与初次VM失败。不触及平台维护源/运行数据库，不请求权限扩大。未新建PR/合并/部署；供应商联调Not Run，路线2–4未实现，持续授权和后续独立Issue接续不变。
