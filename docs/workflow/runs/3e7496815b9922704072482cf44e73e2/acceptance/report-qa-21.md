# Issue28 候选独立验收（执行21）

> 验收日期：2026-10-08 UTC（宿主执行20为19:23:36–19:34:44 UTC；上海时间2026-10-09）。
> 仓库：big91987/model-relay；分支 workflow/3e7496815b9922704072482cf44e73e2。
> HEAD/base：cd413b5aa24a372a19c19af999a6ef61b5c41572；有未提交研发改动，HEAD不是已验新提交。
> 已验候选：147件、SHA256 f8befd158a543afeb1ba77778e4caf72e876f6e4b0b988dd4e6ab73de5e5eaab。
> 模式：全量门禁证据验收＋独立定向执行；不是正式安装上线验收。
> 结论：**Go with known issues，仅允许该候选进入精确发布/正式prepare；正式部署完成仍为Not Run。**

## 1. 结论摘要
执行20的固定 `make verify` 在入口自行创建的私有冷候选完成全部原门禁，exit0；不是协调者推荐脚本或工作区依赖缓存绿灯。独立QA核对当前源码、前后候选、锁、原始完整runner日志、三视口连续旅程及最终依赖探测一致，关闭QA6的冷起点和完整日志证据缺口。QA自身执行的完整入口仍在历史HTTP监听前置被沙箱阻止；该项保持Blocked，不把宿主通过归为QA亲自完成浏览器。本轮没有发现未关闭的当前产品P0/P1；执行18摘要断言的间歇失败原因仍未知，保留非阻塞测试可靠性风险，不宣称诊断代码修好了产品竞态。交report继续既有授权的精确PR/merge/Owner preview部署；没有Deployment、environment_url、health版本/schema3与真实admin新证据，不能宣布上线。

## 2. 基线、版本与覆盖
- 本次Issue28及人工回退边界优先；沿用T001/M01、产品既有架构和 `docs/local-deployment.md`，不重建PRD/UI，不恢复Issue25/旧Run，不重复PR27。
- 原材料：`.workflow-input/3e7496815b9922704072482cf44e73e2/5f51d35fdc15ed7214749f9413879b4238097f1d6b6bbe2cc0d508d8e103cef7/files/manifest.json`。version=design-v0.2.0、status=frozen-for-test-implementation；archive SHA256为该目录摘要。QA重新核对19件全部字节数/hash，不提升材料评审状态。
- 实际引用该root下 `docs/acceptance.md`（2792740b1579d614780856a13b6257d192736e1efc0309f1dc41405a665bef7a）、`docs/interactions.md`（95f3490ef003fa9b19dd6b9e864dd3b0429619b8cd9d0ecec853919d6a12c801）、`docs/screenshots/overview.jpg`（143ba913b91324627bd57a48c25586335b208b50e8c5ad21c5ddf5df2d43715a）。完整材料路径/hash在 `evidence/qa-21/archive-index.json`。
- 当前差异：Makefile依赖前置、Go隔离验证/采集/监督工具、项目说明/规范；内部Go业务仅独立测试场景并行化，产品运行时/数据库/前端业务资产/部署控制器未改。browser变化是有界白名单诊断、计时及回归；原断言、race、全部视口、12分钟门禁不放宽。QA不修改业务实现。
- 工具链：Go1.25.6 darwin/arm64、Node v26.0.0、npm11.12.1；Playwright/core1.59.1，最终解析均candidate/node_modules；Chromium Google Chrome for Testing147.0.7727.15。冷证明覆盖node_modules/npm cache，不冒称Chromium下载缓存或全部宿主工具为首次安装。
- 覆盖：AC01–18回归，重点供应链、零软预算/有效UTC/迁移，H03–06材料/源码/QA衔接；H07正式发布闭环留后续。角色和同一context连续路径见 `journeys-qa-21.md`，权威逐项结果见 `matrix-qa-21.md`。

## 3. 缺陷与进展
维护者后续要求的独立历史摘要评估已完成，见 `summary-assessment-qa-21.md` 与 `evidence/qa-21/summary-history-review.json`/`summary-source-review.json`。旧失败保持Open/根因未知，当前只是一次真实通过；P2验证可靠性风险不阻止本候选进入精确正式prepare，未证实产品P1。源码显示save没有本次post-write渲染完成屏障，这仅是静态候选解释，不是浏览器根因复现。具体用户影响、分级依据及再红最小原层重现/同步/回归契约在补充报告；未再跑整套门禁。
见 `defects-qa-21.md`。原缺包顺序P1及QA6 E28-01/E28-02本次候选验证范围关闭；不等于正式Owner安装prepare已运行成功。执行12提出的父链接越界、辅助诊断丢失、有界取消三项P2，有执行13真实临时目录红灯和当前QA独立race原层绿灯，关闭验证工具缺陷。执行18的summary一次失败保留原日志和JSON，本次三宽度原路径通过，但没有红转绿根因修复证据，不写成产品缺陷已根治。

## 4. 证据核对与独立执行
- Connector20：exit0、310815bytes、truncated=false。通过平台read_command_output按0/32768/65536/98304/131072/163840/196608/229376/262144/294912连续游标取至310815/EOF。长工具消息显示会省略中段；QA还逐字节读取本地完整runner日志、校验并原样归档。没有以省略片段推断通过。
- 实际runner：310718bytes、SHA256 `7c4d3952943721deae9188a6f52f860572037ac8ea39ba55d278d6bf2c7abb7b`，与state一致；它不是310815bytes Connector全文。入口前缀/收据行与runner采集范围分开记录。可复核文件：`evidence/qa-21/host-seq20/verify.log`、`state.json`。
- 冷起点：clean_start_measured=true、node_modules/npm cache=false、NODE_PATH unset、manual_preinstall none。结束本地锁定依赖存在，source/candidate unchanged=true，147件前后均f8befd；gate_exit/runner exit0，cleanup包含candidate_removed，owned_group_absent=true，各辅助命令exit0/未取消/自有组不存在。
- 全门禁：历史预算/回拨旧版实际HTTP红断言准确；HealthMethodsHTTP旧版21项预期红/12项恢复通过分类器PASS；格式/module/vet、197Node（0fail/0skip）、完整Go race包、build、migration/legacy/routing/console/full-console、三宽度J01–J07/summary/十页visual、最终锁/Chromium probe均完整结束。日志内fixture失败是验证失败采集和取消的负例，不是候选门禁失败。
- QA亲自执行：`go test -race -count=1 -v ./tests/verify ./tests/verification` exit0（13.526s/2.573s）；197项Node exit0；format/module/vet/build/diff exit0。原固定 `make verify` exit2，runner exit1/gate_exit2，原层明确 historical listener unavailable; not business red。独立执行日志/原生失败完整runner/state均归档。
- 浏览器收据/截图148件原样归档含完整log/state；每件路径/长度/hash在archive-index。检查文本本机home路径、临时绝对路径、长Bearer/sk标记无匹配；截图捕获代码在清空敏感控件/输出后断言，visual还逐个检查已知秘密不在body中。QA实际查看1280工作台、390模型表、1440Key编辑，布局/当前角色/范围/schema3一致；对照原型信息层级，不声称逐像素一致或人工看完全部截图。

## 5. 未执行、限制与替代关系
- QA自身真实HTTP全套/浏览器：Blocked，固定入口未到browser；宿主同轮同版受控UI-E2E独立记Pass，未扩大沙箱权限、未用静态browser.check代替后端。
- 顶层TestConsoleClockRecoveryWorker：Not Run（skip，子进程专用）。同轮TestConsoleClockRecoveryInProcess/ActualHTTP通过，源码要求私有重启/备份恢复子进程实际输出CLOCK_RECOVERY_PASS，否则失败；它们替代发布不变量覆盖，不把顶层skip变成Pass。
- 正式部署prepare/Owner activate/Deployment/health/schema3/admin：Not Run。旧9851177/schema2仅用户已提供历史状态；本QA没有访问或更新5545。当前受控Go服务标识不是预览正式版本。
- 真实供应商、支付/外部身份集成：Not Run；受控HTTP/SSE只证明代理、故障及计量行为，不证明供应商实际响应、计费或端点能力。

## 6. 阶段边界与发布限制
允许已有授权的report/Connector完成新PR、精确已验提交审查合并和Owner `deploy=true target=preview`；QA不创建分支、提交、推送、合并或启动Actions。新提交须绑定本候选147件摘要，merge后的正式prepare必须重新运行项目入口并取得真实绿灯；摘要变化或门禁失败立即返development，不增限、不跳assert。控制器同SHA日志覆盖仍由Platform维护源负责，不复制修改。授权持续有效，不需要用户再次批准阶段；未知权限/不可逆扩大另按平台规则处理。

## 7. 环境恢复
宿主20原state报告自有进程组清理及候选移除；QA仅认可该次监督范围，不由此推断整台机器没有旧进程。QA自身固定入口失败state亦source/candidate不变、candidate_removed、owned_group_absent=true；定向测试临时目录与子进程由测试清理。本轮没有启动替代长期服务、访问用户上游、写共享数据/密码、改库或广泛kill。旧seq12 cleanup=not_started保持历史未闭合，不被本轮新state覆盖成成功。

## 8. 质量趋势
相比QA6只有暖工作区全绿，本次实际私有冷候选完成全门禁，完整前后状态与日志可追溯；相比18浏览器summary后未完成，现在三视口full-console及最终probe全绿。193→197来自新增诊断负例/计时测试，而非删除route-cleanup或放宽业务预期。父链接/失败stderr/取消已沉淀自动回归。剩余风险是summary历史偶发原因未定位、正式安装与真实供应商尚未验；不保证一次绿灯消除所有调度风险。

## 9. 判定
本轮无已证当前未关闭P0/P1，强制冷门禁证据已闭合；结论Go with known issues，范围仅该候选进入精确发布。P2可靠性风险记录且有边界诊断，不是Owner已接受风险的伪造声明；正式prepare若再失败按真实边界返研发。总体“正式部署成功”仍未完成，report不能将QA放行替代上线回执。发布后最终关闭条件见 `release-handoff-qa-21.md`。
