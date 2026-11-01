# 前端异步状态与凭据展示静态审查

## 结论

审查对象是任务所称 `web/app.js`、`web/workspace.js` 及对应 HTML 在本仓库中的实际路径：`internal/relay/web/app.js`、`internal/relay/web/workspace.js`、`internal/relay/web/index.html`、`internal/relay/web/workspace.html`。本次仅静态审查，未启动浏览器、Go 服务或测试。

发现 1 个 P1、2 个 P2；未发现 P0。最明确的问题是工作台调用测试在普通失败或 SSE 失败后仍保留并重新启用 API Key 与消息，允许敏感值在失败后继续留在 DOM。其余两项是注销非 2xx 未报告，以及工作台部分异步操作没有统一 guarded/epoch 提交入口。两套界面均有陈旧响应防护；租户切换会同步递增 epoch、取消管理请求与调用、关闭弹窗并清除敏感值，没有从静态路径发现已证实的跨租户响应覆盖。`app.js` 的调用测试在失败、取消、离页、会话不确定及注销路径上清密和取消 reader 的闭环更完整。

## 发现

### P1 — 工作台调用失败后 API Key 与消息仍留在 DOM

**事实。** 工作台把调用 Key 设为 `password`/`.secret`，把消息标为 `data-sensitive`（[`workspace.js:264`](../../../../internal/relay/web/workspace.js#L264)-[`workspace.js:268`](../../../../internal/relay/web/workspace.js#L268)）。提交时读取 Key，并将消息直接放入请求体（[`workspace.js:270`](../../../../internal/relay/web/workspace.js#L270)-[`workspace.js:271`](../../../../internal/relay/web/workspace.js#L271)）。失败分支只向输出追加提示；`finally` 随后重新启用 Key、消息及发送控件，没有清空字段（[`workspace.js:272`](../../../../internal/relay/web/workspace.js#L272)-[`workspace.js:273`](../../../../internal/relay/web/workspace.js#L273)）。这覆盖 HTTP 非 2xx、JSON/SSE 解析错误、流未正常结束以及用户按“停止本地接收”触发的 `AbortError`。相比之下，旧控制台失败时显式清空 `test-key` 和 `test-message`（[`app.js:544`](../../../../internal/relay/web/app.js#L544)-[`app.js:547`](../../../../internal/relay/web/app.js#L547)）。

**缺陷假设。** 用户以为失败或停止已结束敏感上下文，但 Key 和提示词仍可被同页脚本、后续操作者、浏览器内存检查或误触再次提交。字段仅在导航、切租户、退出、刷新/`pagehide` 等调用 `invalidate()` 时清除，失败本身不触发清除。

**影响。** 敏感值暴露窗口延长，且失败后重复提交成本/计费风险增加。优先级为 P1：需要真实用户动作或同页访问才能利用，但直接违背“失败后敏感值残留”审查目标。

**已有防护。** 导航、租户切换、退出和 `pagehide` 会经 `invalidate()` 中止请求并调用 `clearSensitive()`（[`workspace.js:60`](../../../../internal/relay/web/workspace.js#L60)-[`workspace.js:62`](../../../../internal/relay/web/workspace.js#L62)、[`workspace.js:276`](../../../../internal/relay/web/workspace.js#L276)、[`workspace.js:287`](../../../../internal/relay/web/workspace.js#L287)）；调用在途期间也禁止重复提交（[`workspace.js:270`](../../../../internal/relay/web/workspace.js#L270)）。

### P2 — 工作台注销只报告网络异常，不检查服务端拒绝

**事实。** 工作台先执行本地 `signedOut()`，再 POST `/admin/api/logout`；只有 `fetch` 抛异常才显示“服务端注销未确认”，任何 HTTP 4xx/5xx 都会被当作成功且没有提示（[`workspace.js:282`](../../../../internal/relay/web/workspace.js#L282)）。后端合同要求合法请求后删除服务端会话并返回 200（[`server.go:236`](../../../../internal/relay/server.go#L236)-[`server.go:247`](../../../../internal/relay/server.go#L247)）。旧控制台会检查 `response.ok`，仅把 401 作为会话已失效的可接受结果（[`app.js:465`](../../../../internal/relay/web/app.js#L465)-[`app.js:469`](../../../../internal/relay/web/app.js#L469)）。

**缺陷假设。** CSRF、来源、存储或服务器错误造成 403/5xx 时，本地界面已退出，但服务端会话可能仍有效，用户没有得到关闭浏览器/等待过期的风险提示。

**影响。** 共享终端上可能误认为服务端会话已撤销。优先级 P2：本地敏感状态已经清除，且需要服务端注销失败才出现。

**已有防护。** `signedOut()` 先递增 epoch、取消请求、清空用户/CSRF/租户和敏感 DOM（[`workspace.js:61`](../../../../internal/relay/web/workspace.js#L61)-[`workspace.js:62`](../../../../internal/relay/web/workspace.js#L62)）；网络异常有明确提示。

### P2 — 工作台存在绕过统一 guarded 状态的异步操作入口

**事实。** 通用 `actions()` 只是把 handler 直接交给按钮，不做按钮禁用或 operation guard（[`workspace.js:92`](../../../../internal/relay/web/workspace.js#L92)）；`runAction()` 只做 epoch 过滤和错误提示（[`workspace.js:18`](../../../../internal/relay/web/workspace.js#L18)）。例如服务连接核验按钮可连续触发 POST，而按钮本身不会进入 busy 状态（[`workspace.js:252`](../../../../internal/relay/web/workspace.js#L252)）；页脚刷新也没有提交 guard（[`workspace.js:281`](../../../../internal/relay/web/workspace.js#L281)）。编辑器写入路径则有 `state.busy`、按钮禁用、固定 idempotency operation 以及“结果未知时安全读回”（[`workspace.js:113`](../../../../internal/relay/web/workspace.js#L113)-[`workspace.js:160`](../../../../internal/relay/web/workspace.js#L160)），说明防护并未覆盖全部异步入口。

**缺陷假设。** 快速双击连接核验可产生并发 POST；最终通常会由后端的 execution revision / `check_in_progress` 合同拒绝第二次，但 UI 可能显示一次失败并打开多份详情，且不能单凭前端保证不会重复副作用。并发刷新主要是体验和额外负载风险；每次随后调用 `navigate()` 会使较旧 epoch 失效，未发现它导致跨租户落屏的确定路径。

**影响。** 重复操作、混乱反馈和无谓请求；当前证据不足以证明会产生重复计费或持久化损坏，因此定为 P2。

**已有防护。** 编辑器写入统一 guarded；API 请求受 AbortController/epoch/user 校验保护（[`workspace.js:70`](../../../../internal/relay/web/workspace.js#L70)-[`workspace.js:78`](../../../../internal/relay/web/workspace.js#L78)）；后端仍负责授权与并发合同，前端不是授权边界。

## 已核对且未发现缺陷的关键路径

- **陈旧响应覆盖。** `app.js` 的 owner token 同时绑定 session、navigation、load/find sequence，并在 fetch、JSON 解析和落屏前断言当前性（[`app.js:52`](../../../../internal/relay/web/app.js#L52)-[`app.js:72`](../../../../internal/relay/web/app.js#L72)）；导航递增版本并停止调用（[`app.js:398`](../../../../internal/relay/web/app.js#L398)-[`app.js:414`](../../../../internal/relay/web/app.js#L414)）。`workspace.js` 的 API 在解析响应后检查 epoch 与 user id，页面提交前再次检查 epoch（[`workspace.js:70`](../../../../internal/relay/web/workspace.js#L70)-[`workspace.js:78`](../../../../internal/relay/web/workspace.js#L78)、[`workspace.js:261`](../../../../internal/relay/web/workspace.js#L261)-[`workspace.js:262`](../../../../internal/relay/web/workspace.js#L262)）。
- **租户切换。** select change 先更新 tenant，随即进入 `navigate()`；`navigate()` 的第一个同步动作是 `invalidate()`，会递增 epoch、abort 全部管理请求及当前调用、关闭弹窗并清敏感值（[`workspace.js:276`](../../../../internal/relay/web/workspace.js#L276)、[`workspace.js:280`](../../../../internal/relay/web/workspace.js#L280)）。租户资源 URL 使用编码后的当前 tenant（[`workspace.js:34`](../../../../internal/relay/web/workspace.js#L34)-[`workspace.js:36`](../../../../internal/relay/web/workspace.js#L36)）。未发现旧租户响应能在切换后提交到新租户视图的静态路径。
- **表单 guarded 状态。** 旧控制台 `guarded()` 在同一 owner token 下禁用提交按钮并抑制 stale 错误（[`app.js:172`](../../../../internal/relay/web/app.js#L172)-[`app.js:177`](../../../../internal/relay/web/app.js#L177)）；模型创建、编辑和上游操作还有独立 operation 状态。工作台登录、邀请接受和编辑器提交均防双击；编辑器对未知写入结果隐藏保存按钮并提供基于同一 operation id 的安全读回。
- **失败提示。** 两套界面都区分读取失败、写入已确认但读回失败及结果未知。工作台在读回失败时移除旧内容，明确不把旧数据当当前事实（[`workspace.js:125`](../../../../internal/relay/web/workspace.js#L125)-[`workspace.js:135`](../../../../internal/relay/web/workspace.js#L135)）。
- **SSE 取消。** `app.js` 同时 abort fetch 与 cancel reader，并在取消/离页/注销时清密（[`app.js:96`](../../../../internal/relay/web/app.js#L96)-[`app.js:108`](../../../../internal/relay/web/app.js#L108)）；读取循环每个 chunk 后校验 owner（[`app.js:517`](../../../../internal/relay/web/app.js#L517)-[`app.js:536`](../../../../internal/relay/web/app.js#L536)）。工作台通过同一 AbortController 取消 fetch，epoch 阻止取消后的落屏（[`workspace.js:270`](../../../../internal/relay/web/workspace.js#L270)-[`workspace.js:273`](../../../../internal/relay/web/workspace.js#L273)）。静态上看取消可达，但工作台没有显式保存并 `reader.cancel()`；浏览器/服务端终止时序仍需动态验证。
- **一次性凭据。** 旧控制台签发前清除上一次秘密，成功后以 password input 展示，支持显式显示/复制/关闭，导航与注销会清除（[`app.js:453`](../../../../internal/relay/web/app.js#L453)-[`app.js:463`](../../../../internal/relay/web/app.js#L463)；对应文案见 [`index.html:11`](../../../../internal/relay/web/index.html#L11)）。工作台同样默认遮蔽、只读、显式显示/复制，并声明关闭、切租户、退出或刷新后不可恢复（[`workspace.js:164`](../../../../internal/relay/web/workspace.js#L164)-[`workspace.js:170`](../../../../internal/relay/web/workspace.js#L170)）。未知结果的 operation 读回只返回 `secret_recoverable:false`（[`console_api.go:40`](../../../../internal/relay/console_api.go#L40)-[`console_api.go:59`](../../../../internal/relay/console_api.go#L59)）。
- **密码展示。** 旧控制台提供有状态的显示/隐藏按钮，并在登录、清密、注销时恢复 password 类型（[`app.js:74`](../../../../internal/relay/web/app.js#L74)-[`app.js:75`](../../../../internal/relay/web/app.js#L75)、[`app.js:96`](../../../../internal/relay/web/app.js#L96)-[`app.js:101`](../../../../internal/relay/web/app.js#L101)）。工作台不提供密码显示按钮，登录和邀请提交的 `finally` 均清空密码（[`workspace.js:279`](../../../../internal/relay/web/workspace.js#L279)、[`workspace.js:285`](../../../../internal/relay/web/workspace.js#L285)）。HTML 对密码、新密码、一次性 Key 使用相应 input 类型和 autocomplete 属性（[`workspace.html:5`](../../../../internal/relay/web/workspace.html#L5)-[`workspace.html:6`](../../../../internal/relay/web/workspace.html#L6)）。

## 未执行项与局限

- 按任务边界未运行浏览器、服务、单元/集成测试，也未修改前端或测试；因此双击时序、慢响应乱序、切租户在途响应、SSE TCP 终止、bfcache、剪贴板权限和屏幕阅读器提示均未实测。
- 后端仅作为合同引用；未全面审计授权实现。前端租户参数和隐藏控件不能替代服务端授权。
- 未使用真实 Key、密码、邀请 token 或上游凭据；报告不含运行日志、截图或敏感值。
- **工具回执历史。** 前一轮执行时任务 MCP 未在可调用工具中暴露，因此当时无法取得 `get_task`/`handoff` 回执；这是该轮 exit 0 但未完成平台交接的真实历史事实。本次正式接续已成功调用 `agent_team_factory.get_task`：任务 `product-audit-frontend-state` 状态为 `running`、attempt 2，允许写入路径仅为本报告；依赖 `strict-bearer-auth`、`single-clock-advance`、`strict-backup-manifest` 均为 `merged`。用户消息再次确认本任务只做静态审查、不运行产品测试。本次将在报告完成后以真实 `agent_team_factory.handoff` 回执交付。

## 建议的后续验证（未执行）

1. 以可控延迟响应覆盖调用失败、SSE 中途断流、用户取消和切租户，断言 Key/消息/输出 DOM 的清理以及服务端连接终止。
2. 对连接核验、刷新、登录、邀请接受和所有编辑动作做快速双击，记录请求数、operation/idempotency header 与最终提示。
3. 让注销端点分别返回 200、401、403、500 和网络错误，断言本地清理始终发生，且除 200/401 外均提示服务端撤销未确认。
4. 构造 A 租户慢响应后立即切到 B，断言 A 的列表、详情、一次性秘密与失败提示均不能出现在 B 的 DOM。
