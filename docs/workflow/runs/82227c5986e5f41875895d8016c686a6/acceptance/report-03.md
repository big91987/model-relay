# 版本13合并前独立验收报告

> 日期：2026-10-06；仓库：big91987/model-relay；分支：workflow/82227c5986e5f41875895d8016c686a6。
> Commit：8ae4e5173aa0dc45c28ad61dbc64df20afe229a6；当前未提交版本13工作树86文件e3cde10304e0bbce0dcebf671f9469ab7e5c494d8524cb9a7b7e04228d71d54f。
> 环境：Go1.25.6，实际Go服务/SQLite/受控HTTP与浏览器由宿主Connector验证；QA免监听Handler/隔离Transport/SQLite及前端VM独立复验。
> 模式：全量覆盖复验（SSE终态变化及合并前Important遗漏发现）；结论：**Go，仅当前产品切片与更新原Draft PR #19，不是正式发布或供应商Go。**

## 1. 结论摘要

两项合并前Important均回到原发现层验证：超长SSE行在首事件前和输出后均明确为协议错误，只访问一个上游，不切换且不冷却；DONE写失败或取消后请求与attempt一致取消，成功归属为空，概览不再留成功证据。成功对照保留正确归属，现有有限切换、工具停止、首配、两屏完整运营和精确旧main非空升级/回退在当前版本宿主固定入口重新通过。QA自己独立契约race及119项VM通过，未沿用旧seq21/report-02放行。无新发现阻断项；平台真实联合链、正式Actions和供应商仍未验。

## 2. 版本与基线

预期来自本Run [PRD](../product-definition.md) R-3/R-4/R-5、AC05～08/12、[架构](../routing-architecture.md)、[交互](../routing-interaction.md)、[验证计划](../routing-validation.md)及平台输入的合并前审查。实现不能改写协议大小错误不切换、只有正常完成可标成功上游的合同。

相对已审HEAD只改两个产品实现文件proxy.go/routed_proxy.go，新增sse_terminal_regression_test.go及Trellis合同/公共状态文档；无schema/API/调度限额、权限、UI驱动、锁和模板变化。当前指纹与seq28及final-source-13相同，Workflow/manifest联合身份与原锁相同，原13个锁定文件和历史QA证据哈希仍相同，见 [version-lock-03](evidence/version-lock-03.json)。验收产物不在原源工具范围内；本节点未添加测试改变源版本。

## 3. 本轮覆盖

完整13AC分层映射见 [matrix-03](matrix-03.md)。原单管理员/单实例范围不变，采用当前同版宿主完整链而非拼接旧版Pass：首配、配置创建/编辑/取消、核验及错误恢复、密钥启停/复制、优先级与权重真实计数、有限切换、未知用量、健康启停/晚到隔离、输出与取消停止、同请求精确诊断、账户恢复及旧数据迁移。

1280/390多上游UI-E2E各自同管理员、同浏览器会话连续完成，390实际创建本屏对象而非查看桌面收尾。服务启动和外部故障/计数属于准备与对账；API/SQL不代替UI配置中间步骤。原首配两屏包含真实clipboard授权/拒绝恢复，routing转录步骤并不伪装clipboard验证。精确旧85f6d17旧UI创建非空对象及真实历史，再backup/upgrade/restore，是独立迁移旅程。上述均由seq28本版真实宿主执行；QA没有另起监听链。

新协议回归分层：真实HTTP测试验证外部访问primary1/secondary0、协议码、无冷却及持久终态；Handler/Transport用例独立验证管理详情及SQLite与概览。DONE写失败以确定性Writer产生真实Write错误/父取消，包括写返回nil但取消已发生；不称真实TCP断线复现。成功对照、原真实HTTP取消/工具停止和两屏SSE停止继续通过。

## 4. 缺陷及进展

沿用原审查两项Important，因违反核心停止和数据一致性AC，映射本验收P1，不降级放行。

| ID | 等级 | 用户能看到的问题 | 复现步骤 | 影响 | 定位信息 | 当前状态 | 回归资产 |
|---|---|---|---|---|---|---|---|
| SSE-01（审查1） | P1 / Important | 协议超长错误触发备用生成并误冷却，可能增加费用 | 双候选→首事件前/输出后超长data行→查看同ID和计数 | 重放风险、未来请求健康误伤，AC05/06/07/12 | proxy Scanner.ErrTooLong；routed临时白名单 | Closed | TestSSEOverlongLineNeverSwitchesOrCools及TestSSERealHTTPOverlongLineNeverSwitchesOrCools |
| SSE-02（审查2） | P1 / Important | 流结束写失败或取消却仍指认成功上游 | 正常增量→DONE Write失败/取消→管理详情/概览 | 终态/归属不一致，AC07/08 | proxy finish及DONE前后；routed streamFinish | Closed | TestSSEDoneTerminalAttribution的成功对照及三失败/取消子例 |
| QA-01～04（原全量） | P1 | 核验协议/预算/状态与完成步骤问题 | 原负控及连续双屏路径 | 原AC02/09 | 原remediation-01 | Closed，当前回归未复发 | 原QA测试/connection生命周期/两屏routing资产未降低预期 |

研发 [premerge-sse-01](../delivery/verification/premerge-sse-01.md)保存已审HEAD红控和历次失败；本节点审查其对应断言，不冒称自己复跑旧版红控。当前QA独立同层复测及宿主本版新HTTP绿控均存在。

## 5. 实际证据

- [宿主seq28](evidence/host-seq28.md)：固定make verify，900秒预算，exit0，81349 bytes完整存档；重点读取新测试/相邻停止/浏览器段至eof。原预期历史红控与本版结果严格区分。
- [QA专项](evidence/sse-retest-03.log)：QA实际GOCACHE项目缓存/GOTOOLCHAIN=local，go test -race -count=1 -v，原SSE两主题及QA核验/取消/整数/健康代次7个顶层，exit0，28.424秒；不包含监听HTTP用例。
- [QA前端](evidence/frontend-retest-03.log)：原node --test，119/119、0 fail/cancel/skip，exit0，183.962292ms；不是实际浏览器证据。
- [版本锁](evidence/version-lock-03.json)含当前源码/原门禁/模板/辅助产物摘要；[最终审计](evidence/final-audit-03.log)核对引用、13AC、身份与旧证据保留。
- [首次静态审计](evidence/final-audit-03-before.log)真实exit1：仅最终审计日志自身尚未落盘导致两条引用缺失；保留失败，再实际复跑，不归为产品故障。
- 本版PID8419/156行安全选择轨迹及当前390回看截图已归档。QA查看该窄屏截图；辅助文件没有Connector制品签名，不能替代主回执，也不是QA新启动浏览器。

## 6. 未执行和阶段边界

QA沙箱原listen/Mach限制未变，QA自己完整监听/Chromium重跑Blocked；不扩权、不重复已知受限入口，同轮同版本seq28替代相同用户结果门槛，但Blocked行不改Pass。

DEP02～04真实可信controller旧/新Go联合链、DEP06正式Actions非空升级、DEP07同可信路径隔离失败恢复、DEP05未覆盖安装边缘及真实供应商仍Not Run。标准模板联合锁未变，不代表平台全部升级成功。缩短超时的真实停止回归不等同300秒生产SLA；受控上游不证明真实费用或供应商能力。原PR19尚未合并或部署。本结论只允许进入报告并由既定Connector更新原Draft PR19，维护者发布授权持续有效但不能预报发生；不自动合平台源PR、不人工预upgrade、不改平台数据库。

## 7. 环境恢复

QA专项使用隔离临时数据和Transport，正常退出/cleanup；未触碰用户数据、凭据和真实上游。seq28 runner正常退出，原cleanup回收临时目录、Go服务及browser；QA未独立做宿主全机残留审计。cache/node_modules/test-results继续ignored，只归档无凭据截图和白名单轨迹。未修改产品、门禁、Workflow或平台安装/数据，未提交推送。

## 8. 质量趋势

本轮关闭2项合并前遗漏P1；原四项P1复发0，新未关闭缺陷0。相对report-02扩大Scanner大小错误和DONE取消纠正覆盖，避免“正常流测试通过”漏掉异常终态。自动化资产直接纳入原go test及make verify；真实HTTP与确定性Writer证据分开，既有UI/迁移全量回归没有降级。剩余风险仍集中正式可信部署链和供应商。

## 9. Go / No-Go

**Go，next→report。** 当前版本13同版强制入口通过，原发现层两项Important/P1已关闭，核心受影响旅程和最终事实通过，无开放P0/P1。报告节点更新公共任务状态和本Run pr.md，保留report-01/02历史与各次失败，不重建需求设计；交既定发布节点更新原PR19。后续行为/模板变化须新版本重验，不使用本结论替代正式部署验收。
