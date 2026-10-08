# 执行19：摘要原断言最小归因诊断

## 恢复与真实结论
- HEAD/base仍为 `cd413b5aa24a372a19c19af999a6ef61b5c41572`；既有研发改动保留，无Git发布、Connector或部署控制器改动。
- tests18正式make verify Connector exit2、307458bytes/truncated=false，本轮read_command_output从304500读到307458/EOF。实际runner完整日志307276bytes/SHA256 `cb92549be85d96006a9d9131a12067982e41abccf5e75147c7353f0803e7c430`，与原state一致。原样归档 [tests-18-verify.log](tests-18-verify.log)、[tests-18-state.json](tests-18-state.json)，不是Connector全文。
- UTC 2026-10-08T19:02:33.573258Z–19:13:01.656265Z、628.083秒。原state冷起点node_modules/npm cache不存在、NODE_PATH unset、manual_preinstall none；source/candidate前后147件摘要 `28571f199f257d5c0bb09fc679150ddb8c5beff822c1a8c813a3bb54c064eeb7`一致。gate_exit2/runner exit1。cleanup attempted/candidate_removed及owned_group_absent=true来自原runner持久化，不是本轮另测全机无残留；无广泛kill或替代服务。
- internal/relay race实际305.964秒；routing三宽度/console三宽度通过。full-console1280px J01–J07及current-period视觉活动通过后，summary-model-context-toggle/form-submit/last_status200/AssertionError失败。后续full-console1440/390、完整browser和最终依赖/Chromium probe未完成。这不是deadline或缺包，也不是已证产品错误。
- 安全浏览器失败JSON原样归档 [tests-18-continuous-1280.json](tests-18-continuous-1280.json)。failure_boundary.available=false，未记录唯一行/列/值哪项失败，也不能区分哪次启停。日志已检查本机home路径、长Bearer/provider key标记无匹配；不上传raw trace/密码或未经审查截图。

## 最小变化边界
问题在原摘要断言及failed receipt丢失边界。本轮只改tests/browser/console-summary.mjs、journeys.mjs和两者原有Node测试，维护现有Trellis规范。不改产品UI、Go运行时、DB、材料、锁、并行度、时限或原有断言；无重试/polling/sleep/跳门禁。

summaryCell仍依次断言唯一行、列存在、值相同，WeakMap仅关联安全测量并重抛原error。receipt新增summary_cell（边界、行/列计数、列存在、值相等，仅状态白名单）和content_loading布尔；源码位置仅允许两个固定文件及数值位置。不输出原始标题/值/expected、异常字符串/stack、路径或IDs。

源码事实：workspace.js的edit先dialog.close，再await syncTenants，后await renderPage；journeys.save在dialog隐藏后只检查当前loading隐藏。可能读到旧表格或新loading是待证假设，不是本次已证根因。因此尚不改同步或产品顺序。Node受控负例只证明诊断保留count/column/value红灯和原error身份，不能证明真实Chromium/网络调度。

## 回归与规范
使用trellis-before-dev恢复现有frontend/backend/guides索引及适用规范；Trellis实际已有非模板规范，prepare.initialized=false不是本轮初始化成功。任务沿用本Run，不重做冻结需求/架构。trellis-check复核范围、原门禁、隐私、契约和断言，真实结果见 [development-19-check.log](development-19-check.log)。新增3项Node回归进入原wildcard：精确摘要边界/私密值排除、locator原error身份、summary源码位置白名单/receipt接线。

本地完整Node及Go runner/依赖非监听回归、format/module/vet/build/diff结果以该日志退出码为准，不是新冷全门禁。不重复已确证监听限制的原生完整make verify；保留 [development-17-check.log](development-17-check.log)，待验新源码交宿主。

实际本地结果：Node197/197、0fail/0skip；Go tests/verify race13.094秒、tests/verification race3.394秒均exit0；format、module校验、vet/build、diff检查均exit0，日志LOCAL_CHECKS_EXIT=0。没有新增监听或浏览器实例；这些局部回归不验证本次真实summary失败原因。

## tests交接契约
仅原冻结argv make verify，不用推荐脚本/手工npm/NODE_PATH/修改业务记录。新候选见 [source-development-19.json](source-development-19.json)，旧147件戳不能放行当前源码。冷Go/npm缓存、独立无依赖候选、锁定版本、内层12分钟/Connector900秒、全部历史红分类器/预算UTC/迁移/HTTP/浏览器不变。

原样归档实际receipt目录的verify.log/state，核对bytes/hash及source/candidate前后。失败continuous JSON应有summary_cell/failure_boundary/content_loading；再失败依据真实边界返研发，不轮换猜测性修复或增限。完整绿才核对最终probe、cleanup和三宽度summary/visual，再独立QA。

旧seq5/8/10/12/14/16/18证据不覆盖。T001 In Review/M01 Blocked/QA6 No-Go保持。发布/精确merge/Owner deploy=true target=preview授权仍有效，须新完整冷绿及独立QA。未取得正式Deployment/environment_url/healthz version/schema3或admin页面新回执，不宣布上线；上游旧预览9851177/schema2未重新查询/重置/手改。真实供应商未验、原生QA浏览器Blocked分别保留。
