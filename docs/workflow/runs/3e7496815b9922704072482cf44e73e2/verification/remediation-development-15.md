# 执行15：执行14超时的最小归因诊断

## 结论与候选
本轮是诊断候选，不是超时根因已修复，更不是完整冷验或部署通过。T001 In Review、M01 Blocked、独立QA No-Go保持。HEAD/base仍 `cd413b5aa24a372a19c19af999a6ef61b5c41572`；当前147件源码摘要 `24725bbb7833cfc2665d44b67de4982cc045aa375ad5ec65c68c6c7d6f2b044e`，见 [源码回执](source-development-15.json)。执行13的144件dc73ac仅是执行14实际被测版本。

## 已证事实与证据边界
正式tests执行14外层exit2，Connector声明249788bytes/truncated=false。本轮读取保留输出尾段至249788 EOF，但工具投影有省略；不声称已逐页恢复Connector全文。以下归档来自实际固定入口生成的唯一目录 `test-results/verification/20261008T180607Z-863663337`，不是手抄摘要：

| 实际文件 | bytes | SHA256 |
|---|---:|---|
| [tests-14-verify.log](tests-14-verify.log) | 249576 | e37932ca114a297ca2e0cd69b8ca3efb6fbe1d14f1d5a866d478e5c7cb616e5e |
| [tests-14-state.json](tests-14-state.json) | 2764 | 5365b7e5c0a16b1408fbeae661d1434f145b509db7d43e04e64f4b7b4dacc496 |
| [tests-14-routing-selection.jsonl](tests-14-routing-selection.jsonl) | 994093 | 24392a850e10f8c0cc5359dbc1543b0911e09ca1b5fe18f2c239a20575b375cd |

runner完整log与state的log_bytes/log_sha256实际一致，失败原样保留；与Connector全文不是同一文件/长度。公开副本未改变字节。检查日志无用户主目录/本机临时目录路径或常见token模式；选择JSONL只含受控阶段、执行ID、索引、焦点/事件/布尔边界，不含输入值、响应体、HTTP头。未新增公开截图。

实际开始 `2026-10-08T18:06:07.567014Z`、结束 `2026-10-08T18:18:11.700773Z`；开始node_modules/npm cache=false、NODE_PATH=unset、未手工预装、source/candidate_unchanged=true。日志有原历史预期红分类器、193Node、Go/race/build、budget-clock-r1零软预算/UTC及schema1/2→3受控浏览器迁移输出；原三视口J01–J07/十页尚未完成，不能放行。

末尾 `make[1]: [browser] Terminated: 15`、`gate_exit=-1`、`context deadline exceeded`，与runner源码12分钟WithTimeout一致。Node泛化routing Error出现在父门禁终止后，不能当作先前产品断言失败。外层900秒、内层12分钟；不是MODULE_NOT_FOUND，也不是Owner stop。state记录exit_code=1、finished_utc非空、gate及全部已启动辅助进程组owned_group_absent=true、candidate_removed。不同于seq12未闭合state；没有改写seq8/12事实。

选择文件153条：两视口各40次权重阶段选择、31次折叠阶段后编辑/优先级/禁用相关调用选择；1280已输出完整路由Pass，390在failover阶段有两条选择记录。末次原始事件selectedIndex=targetIndex=5、sameOptions=true。证明选项当时匹配，不证明后续请求/记录/冷却完成。没有单步耗时，累计正常耗时、异步等待还是驱动卡住均未判定。产品行为、驱动、时序仍待区分；旧中断与本轮超时不混为一个原因。

原生 `ps -o pid,ppid,pgid,command -p 48396,48293` 实际被沙箱拒绝（operation not permitted），未杀任何进程。执行14owned groups闭合以实际runner回执为依据，不声称清查了旧seq8/12或所有全局进程/端口。

## 最小变更边界
缺口在真实浏览器驱动原层：总deadline只留下泛化阶段，不能区分累计工作与某个未完成边界。本轮不猜测修复，只新增计时观测：
- `tests/browser/main.go`：既有migration/legacy-ui/routing/console/full-console调用的UTC开始/结束、单调耗时和ok，保持调用顺序/错误传播。
- `tests/browser/routing-progress.mjs`：每个真实call的固定序号/视口及panel/alias/key/message/stream/send/complete耗时；某步5秒未结束只输出一次pending，finally清timer。不重试、不跳步、不延迟或取消原action，不记录值或错误内容。
- `tests/browser/routing.mjs`：在原边界使用诊断，泛化failover stage细分fixture/request/record action/rendering/counts/cooldown；原生Tab/typing、请求次数、权重、全部断言、31秒恢复和SSE取消保持。
- Node/Go新回归验证实际timer清理、值/错误identity、一次调用、UTC及白名单输出，进入原门禁；夹具不证明浏览器时序。
- Trellis依赖准备规范增加同一契约；维护既有公共任务/里程碑/索引，不新建PRD或设计。

不改Makefile门禁、runner12分钟、Connector900秒、锁/版本、业务/API/DOM/DB/预算UTC、权限或部署控制器。不凭累计耗时猜测优化驱动、改fill/selectOption、降低断言或延长deadline。

## 检查及失败保留
before-dev恢复本Run任务、冻结manifest/AC、既有implementation-contract、实际非空Trellis验证/依赖/运行时规范及共享指南；get_context --mode packages确认单repo。prepare.initialized=false不据此重初始化；已有源码规范无需bootstrap。
- [development-15-tests-initial.log](development-15-tests-initial.log)、[development-15-static-initial.log](development-15-static-initial.log)：局部采集zsh保留变量status使collector exit1，保留、不作最终成功；改用bash普通变量重新执行。不是产品或宿主门禁失败。
- [development-15-static-gate-failure.log](development-15-static-gate-failure.log)：包装隐藏既有源码门禁要求的 `consoleJourney()`/`fullConsoleJourneys()` 字面调用，194Node中193pass/1fail。恢复显式调用闭包，原测试不改。
- [development-15-tests.log](development-15-tests.log)：中间Node/syntax/新Go回归及两包race exit0，不替代最终候选或完整门禁。
- [development-15-tests-final.log](development-15-tests-final.log)：最终Go新回归及runner/依赖两包race逐项exit0，含原实际取消/SIGTERM、文件边界及两路日志回归。
- [development-15-static.log](development-15-static.log)：最终format/module/vet/build/diff逐项exit0；原已有工作区依赖下194Node全pass（原193+1新增），不是冷安装或真实浏览器证据。

Trellis-check核对原断言/调用顺序/原生动作、error identity、timer清理、隐私及spec同步；update-spec记录诊断。公共交付审查为Ready for Dispatch（有界归因实测），同时Blocked by Validation，不是Accepted/Delivered。稳定原生监听限制未变，不重复必然受限的完整make verify；完整新门禁仍由正式宿主执行。

## 下一次正式tests责任
只用冻结argv原 `make verify`，当前候选自行生成新独立冷副本。不手工安装/删依赖、不运行旧推荐脚本、不改Connector或增限。归档新完整log/state及隐私审查后的具体browser文件到本Run，更新索引；校验开始无node_modules/npm cache、NODE_PATH unset、未预装、147件当前摘要，结束源码不变、完整log哈希/bytes、清理及候选内锁/Chromium探针。

先从BROWSER_PHASE UTC定位browser启动和各旅程累计耗时，再从routing_call序号、durations/pending识别活动边界，结合原断言/原始失败。再次deadline失败须带此最小诊断返development，不把No-Go变Pass、不盲目轮换修复或无限重跑。实际归因后才针对性修复及正式完整复验。

仅完整新冷门禁exit0、全部三视口J01–J07/十页和独立QA达标后，才能走授权的新PR/精确已验提交合并与Owner正式preview部署。Deployment/environment_url、healthz version/schema3、真实admin/旧数据密码守恒必须实际闭合。本轮未Git发布/部署，也未重新查询5545。供应商Not Run、QA6 No-Go和原生浏览器Blocked保持。

材料沿用design-v0.2.0，archive SHA256 `5f51d35fdc15ed7214749f9413879b4238097f1d6b6bbe2cc0d508d8e103cef7`，manifest/AC/交互路径沿T001追踪，未改变冻结状态或执行材料脚本。
