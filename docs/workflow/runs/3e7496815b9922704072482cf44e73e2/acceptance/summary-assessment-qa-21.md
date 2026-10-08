# QA21补充：执行18模型摘要失败的独立评估

## 判定先行
**历史问题保持Open/根因未知；不将执行20一次通过写为修复或关闭。** 当前建议仍为 **Go with known issues，仅该候选进入精确发布及正式prepare**。已证用户影响是执行18整轮验收被该断言阻断，增加交付耗时；没有足够证据认定产品持续显示错误状态、模型操作不生效、权限/数据错误。按现有证据分级为P2验证可靠性风险，而不是已证产品P1。正式prepare或部署后原路径若再失败，该次门禁必须No-Go并返development；这个后续强制门禁不是承诺“能捕获所有偶发问题”，也不抹除当前残余风险。

本判断依据本轮同版完整冷门禁及原三宽度摘要路径实际通过、原断言未弱化、业务资产未改；不是“PATCH200等于UI正确”，更不是故障根因已消除。无需重复用户审批，report必须公开风险并由已授权协调者按真实门禁继续。若后续证实完成刷新后仍为错误状态、核心用户必须额外隐藏操作才能继续，属于AC01/模型管理核心结果错误，应提为P1并阻止发布；不能继续以测试抖动掩盖。

## 1. 旧失败的证据边界
旧候选147件28571f199f257d5c0bb09fc679150ddb8c5beff822c1a8c813a3bb54c064eeb7与 `verification/source-development-17.json` 一致。实际完整 `verification/tests-18-verify.log` 为307276bytes/hash cb92549be85d96006a9d9131a12067982e41abccf5e75147c7353f0803e7c430，与原state一致；gate_exit2、runner_exit1、cleanup原样保持。没有用seq20的state修补seq18。

`verification/tests-18-continuous-1280.json` 记录：1280px、stage=summary-model-context-toggle、action=form-submit、last_status200、error_kind=assertion、editor_open=false、editor_invalid=0、editor_failure=false、workspace_visible=true、secret_present=false、七条J01–J07已完成。full-console阶段以ok=false结束。**failure_boundary.available=false**，无summary_cell、content_loading，也没有哪一次false/true启停、唯一行数、列存在与实际/期望状态的原层记录。因此不能从旧收据确定行/列/值是哪一个断言失败，不能确定实际旧值/缺行/新loading或持续错误。

旧receipt的read_facts/usage_facts等是旅程累积数据，不是这一次模型刷新GET的逐请求归因，不能用其200或费用守恒推断这次页面已更新。last_status200最多证明该路径记录的写请求状态满足200预期，不能证明渲染完成、API最终模型值或供应商状态。没有记录到的数据保持未知，不回填猜测值。

## 2. 原断言与当前源码核对
源码证据 `evidence/qa-21/summary-source-review.json` 保留具体片段、文件hash与判断类别Static Review。以base HEAD源码做可访问对照，并交叉引用development17/18源码戳及development19变更说明；没有把base全树误称为旧147件未提交候选快照。

- base `tests/browser/console-summary.mjs` 的summaryCell依次要求：唯一精确行=1、指定标题列存在、该格trim后严格等于预期。当前15–37行保留这三项条件，仅加入WeakMap白名单facts并重抛同一error；没有捕获后吞错、改变成功标准、轮询、睡眠或重试。
- `tests/browser/console-summary.mjs:63` 的false→true两轮启停及提交后状态断言，与base同段逐字一致；development19没有修改此用户动作路径。
- `tests/browser/journeys.mjs:74` 的save注册精确PATCH响应等待，确认200、读JSON；无一次性秘密时只等editor隐藏后调用ready。`tests/browser/journeys.mjs:43` 的ready只等loading文本hidden，**未识别当前post-write刷新已开始/完成的代次**。
- `internal/relay/web/workspace.js:94` 在提交完成后先清busy、关闭dialog，再显示“写入已确认；正在读取当前事实。”，随后await syncTenants、await renderPage。41–46行syncTenants是一次异步GET；179行renderPage才换loading；201行才提交新section。该产品文件与base没有差异。
- 因此源码没有建立“dialog隐藏/loading不在→本次刷新完成”的可靠happens-before关系。若tenant GET暂未返回，旧表格仍可存在且loading不在；若render刚启动，行也可能暂时不存在。**这是静态可见的同步契约缺口/候选解释，不是执行18真实浏览器调度已重现或已证根因。** 其他产品或驱动原因未被排除。

本次只执行 `node --test tests/browser/console-summary.test.mjs`，4/4通过，保留行数/缺列/错值负例、原error身份、白名单诊断回归。该测试使用受控locator，仅证明断言/诊断契约，**不证明Chromium调度、异步GET实际顺序或历史UI故障已复现**。未模拟一个“绿色页面”冒充真实浏览器；未重跑完整make verify/宿主browser。

## 3. 用户影响与分级依据
| 事实或风险 | 证据强度 | 用户影响 | 本轮判断 |
|---|---|---|---|
| 执行18原摘要断言失败，后续视口/probe未完成 | 当次真实宿主原层log/state/JSON | 该次发布验证No-Go，交付延期 | 保持失败事实，不关闭 |
| 当前20三视口完整J01–J07/summary/visual及probe通过 | 当次真实宿主原层完整证据 | 当前候选可按授权进入精确prepare | 一次真实通过，不代表所有调度都正确 |
| save完成屏障不覆盖本次刷新代次 | 源码Static Review，未动态重现 | 可能导致自动化过早观察旧/空状态、门禁偶发阻断 | P2验证可靠性风险Open；非Accepted Risk伪声明 |
| 操作已提交后的最终页面持续错误或数据错误 | 当前无原层证据 | 若成立影响模型日常管理核心AC，可能误操作 | 根因/影响未知，不能宣称不存在；一旦实证按P1处理 |

暂态刷新期间旧表格与“正在读取当前事实”提示，并不自动等于最终状态错误。相反也不能因为提示或PATCH成功就免验新状态。现有证据不支持把未定位故障升级为已证业务P1，也不支持关闭它；自动推荐为P2风险公开随候选放行，不扩大权限、不请求重跑完整失败循环。

## 4. 再现/整改的最小原层契约（若再红，随handoff返研发）
这里是后续应验证的契约，不是本轮已运行或已修复。

1. 使用本仓真实Go服务及已有受控上游；独立临时安装/角色context。直接复用summary步骤的模型列表、精确模型行、“启停”、checkbox false→保存→真状态→true→保存。UI写步骤不得由API/SQL补齐，不修改业务记录或改断言。
2. 确认保存前状态和当前租户；在点击保存前布置只针对**本次成功PATCH之后**同scope刷新GET的受控屏障。分别对tenants GET和模型catalog GET做可控hold/release，不能混入初始导航或旧scope请求。拦截仅暂停真实请求转发/响应，不用假catalog替代真实后台。
3. PATCH200但post-write tenants GET held时，记录dialog、notice、loading、精确行/列/白名单状态及本次刷新阶段；测save是否提前resolve。释放tenant GET后再独立hold catalog GET，检查loading阶段不能被误当本次完成。最后释放GET，等该次正确scope的DOM提交后执行原唯一行/列/状态严格断言。两个启停方向及窄屏至少各一次；定位后再跑完整原门禁。
4. 同步完成的最低契约：响应200/弹窗关闭只证明写确认；save返回“结果可断言”必须发生在**本次成功写入引发的同角色、同scope、同页面刷新完成并提交DOM之后**。识别该次读响应与实际渲染事实，不能仅global networkidle/loading hidden/sleep/增加超时/重试assert。若修驱动，只修完成屏障，保留原错误行为红灯；若实际产品刷新不收敛，由development修产品，再原UI路径回归。
5. 邻近边界：刷新GET失败不作旧数据成功；scope切换/旧响应晚到不得覆盖新scope；第二次相反操作不得被第一次迟到结果覆盖；取消/关闭/新编辑保持已有busy/Secret/幂等规则。用失败用例证明仍能检出真正的错状态、缺列和重复行，不许poll-until-green遮盖错状态。
6. 原层证据应有具体JSON与完整log/state：toggle direction、write/read阶段归属、summary_cell/failure_boundary/content_loading、安全scope匹配布尔与原error_kind。不保存raw body、秘密、完整对话或绝对本机路径。确认稳定最小红→绿后再正式make verify→独立QA；不能用mock locator结果替代原层。

## 5. 本次结论与下游限制
本补充不改已验147件f8befd源码、不改旧state、不关闭R28-04、不弱化断言。动态独立浏览器仍Blocked，供应商仍Not Run；seq20正式宿主通过是独立一行，不冒称QA本地完整通过。发布部署Not Run保持。

交report继续既有授权精确PR/merge/Owner正式preview。report和协调者必须公开此Open风险，正式prepare继续原固定门禁，任何再红均不能宣布上线或手工越过；按上述契约携带真实失败到development。若正式门禁及发布后admin原路径实际成功，可如实记录安装/上线结果，但仍不得追认执行18为已复现修复。
