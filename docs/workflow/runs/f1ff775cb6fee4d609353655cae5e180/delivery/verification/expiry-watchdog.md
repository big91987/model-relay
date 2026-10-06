# 已知到期不依赖悬挂会话检查

2026-10-06，同Run研发第二次补充。权威输入仍为[PRD AC-02/12](../../product-definition.md)、[架构§6](../../console-architecture.md)和[交互](../../console-interaction.md)。T002/T003 In Review、M02/M03 Ready for Review；未修改已接受需求、设计或原门禁。不是全量产品、独立QA、浏览器/网络或真实供应商验收通过。

## 根因、修复与证据边界

第一次补充的scheduleSession仅在checkSession的finally安排下一次检查/到期触发。当30秒检查一直pending，一次性timer已经消费且finally不运行，已知45秒到期失去本地触发机会。另一个缺口是正常成功检查无条件提示重新输入，即使调用仍在接收。

仅修改原生app.js的会话生命周期、ui.test.mjs、真实Trellis规范及本Run公共交付文件，不改Go协议、服务端会话寿命、存储、认证或平台。修复方式：

- `scheduleExpiry()`使用独立sessionExpiryTimer，按会话epoch及原始截止时间持有；不依赖查询结束，跨安全页面仍生效。查询pending及不明状态/手动恢复pending期间均保持已知到期清理。
- `expireSession()`在已知到期时登出并取消生成/reader、清秘密及输出；在接收查询结果、提交、pageshow/可见恢复时也检查截止，防止迟到结果复活。登出/pagehide清两类timer；恢复页面重建已知截止约束。
- 当前会话查询可首次建立或缩短到期，但不得推后已知expires_at。退出/重新登录用新的身份；旧成功、网络错误、401及finally不修改新会话/页面/生成或timer，不自动重发。
- 正常周期成功保留真实生成状态、输入、输出和提示；只从“不明”恢复才要求重新输入并手动发送。
- 复用既有owner/current、api、signedOut、stopCall、clearSecrets；未添加超时重试、自动登录或生成重试。

测试fixture现对所有timeout实行一次性语义：调用callback之前从计时器集合/delay/deadline集合移除；再次调用旧callback不重复触发。fake clock与实际app.js构成确定性隔离回归，并不是真实浏览器/网络/供应商证据。

## 正式回归覆盖

新增10项测试，进入原make verify的frontend-test入口，总计58项；原48项和原真Go浏览器入口保留。

| 组合 | 强断言 |
|---|---|
| 初始45秒到期、30秒检查pending、46秒触发独立截止 | 未resolve/reject GET之前即取消已开始SSE的signal/reader、清密钥/消息/已有可读及原始输出、显示登录，保持安全hash；到期后timer集合清空，无生成重发 |
| 该悬挂GET迟到成功/网络失败/401 | 三分支分别验证不能恢复旧会话、不能覆盖过期提示或污染输出、不能重建timer；成功返回更晚期限也不延寿 |
| pending GET跨页→原截止→重新登录并新调用→旧结果到达 | 三分支分别核对上游页敏感草稿到期清除、目标恢复、新调用未abort/输入保留、截止集合不变、旧结果不影响新输出 |
| 不明状态→手动检查pending→原截止 | 保持本地已知截止，迟到恢复成功不复活已过期会话 |
| 正常周期成功、仍在SSE接收 | 接收状态、输出、输入和已有notice保持；不提示重新输入/发送；服务器返回更晚期限不延长原截止 |
| pagehide/pageshow及不明时visibility恢复 | 卸载停止timer/生成；重新可见时检查已知截止，即使旧查询悬挂或timer投递暂停也不能恢复过期内容 |

## 当前对象与真实命令

HEAD仍`d62cd51e8de2127548e1ac7469d349443d1bfe3d`，本节点未提交、推送或发布。当前工作树摘要见[source-expiry-fix.json](source-expiry-fix.json)：`9254855737fbe400acf1d8c3a038767ca7807e0e1ea43160b7b5ed87d5456d0a`，56文件，Go1.25.6；与[verify-attempt-07.log](verify-attempt-07.log)首部一致。

| 实际命令/对象 | 结果 | 回执 |
|---|---|---|
| node --test tests/browser/ui.test.mjs，产品修复前，一次性fixture和最初4项强回归已加入 | 52项中48通过、4失败，退出1；三次pending检查均无剩余截止timer，正常成功提示也失败。日志仅脱敏workspace路径，保留断言和退出码 | [ui-expiry-before-fix.log](ui-expiry-before-fix.log) |
| make verify，第7次 | 退出2；可信历史HTTP门禁严格REJECTED：`listen tcp 127.0.0.1:0: bind: operation not permitted`；未跳过或移除门禁 | [verify-attempt-07.log](verify-attempt-07.log) |
| make format-check vet frontend-test build；app/journey语法；diff检查 | 58VM全通过，Go格式/vet/build及语法/diff均退出0；只是分项反馈，不是固定入口全通过 | [checks-attempt-06.log](checks-attempt-06.log) |
| go run ./tests/browser，第4次 | 退出1，受控上游监听受限；未启动Chromium旅程，未生成截图 | [browser-attempt-04.log](browser-attempt-04.log) |

## Trellis-check及后续责任

恢复并遵循前端索引、DOM生命周期、console所有权、验证及复用规范；`.trellis/spec/frontend/console-lifecycle.md`原位记录独立到期/一次性timer/正常恢复文案契约。bootstrap不适用，已有真实初始化/规范；无平台补丁、无新依赖、无凭据或会话文件。Makefile及指定历史健康/task/requirements/design与HEAD无差异。

保留[首轮验证和18AC映射](development.md)、[第一次补充](supplement-rework.md)及其失败日志/源码快照；不得把新分项成功改写为此前门禁通过。当前固定门禁仍Blocked by Validation（原生沙箱监听限制），唯一解除条件是原宿主测试Connector实际执行同一源码的make verify。独立QA仍须逐条18AC及设计边界核对，而非仅计数58项测试。

**真实供应商联调Not Run**：无真实供应商凭据或响应证据，受控上游与VM均不能替代。路线2–4未实现，仍按项目唯一roadmap接续。PR/合并/正式测试部署/部署后体验均尚无回执。本节点未扩大权限或改变平台状态。
