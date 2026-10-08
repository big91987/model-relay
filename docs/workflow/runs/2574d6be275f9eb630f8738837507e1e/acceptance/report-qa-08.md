# Issue30 模型保存读回契约验收报告

> 日期：2026-10-09（Asia/Shanghai）；宿主UTC 2026-10-08T22:19:51.041183Z–22:30:55.912316Z。
> 仓库：big91987/model-relay；分支 workflow/2574d6be275f9eb630f8738837507e1e；HEAD 9bf6ecaf53758ca2258498968b830c535e02e1c0 + 未提交候选。
> 快照：150文件，0384f115adbbd545e8ce01021f238010ba66c89eb06e6acea08ce872ce0fa474；QA前后及development6/tests7一致。
> 环境：正式tests7私有candidate、独立临时Go安装/受控HTTP、真实Chromium；不是Owner现有preview或真实供应商。
> 模式：本Bug全量受影响验收 + 原完整项目公共门禁；QA独立源码/回执/契约/截图复核，独立动态浏览器单列Blocked。
> 结论：Go with known issues，仅该精确候选可交report及授权精确发布链；不是已合并/已部署/全面真实环境验收。

## 1. 结论摘要

正式tests7一次完整make verify exit0，同版三宽度真实Go页面保留原严格断言通过。12组false/true保存共24个tenants/catalog真实响应暂停阶段均没有提前完成；原样释放后准确模型行/列/状态正确。显式GET503、scope切换、相反第二次保存、旧快照晚到等9组边界也通过。已证R30完成契约缺口可关闭，R28-04仍Open/P2、旧tests18根因未知，未证持续产品P1。独立动态QA Blocked、供应商及Owner实际登录预览Not Run分别保留，不拿宿主受控自动化冒充这些结果。

## 2. 本轮覆盖

权威映射及逐角色/起点/步骤/事实/清理见 acceptance-matrix-qa-08.md，不重建PRD/HLD。恢复T001/M01/总览及development6实际交接；按Bug短路径跳过requirements/design，原design-v0.2.0不变。材料manifest、archive和19份文件全部核对摘要，实际引用prd、interactions、acceptance与原型资源路径记录在evidence/qa-08/material-check.json；材料不授予权限，不执行包内脚本。

UI主证据为同轮宿主真实Go与Chromium：每宽度空安装后连续J01–J07，平台/租户管理员、开发者、财务按正式邀请/登录各自context；每角色段连续操作。UI写入不由API/SQL代办。测试控制面只准备受控HTTP故障、显式时间及临时对象清理；API读取只对账。管理summary和30页visual按scope/真实事实核验；QA亲自查看1280/1440/390模型页归档截图，列/状态/控制均存在，390局部表格横向容器不等于缺列。截图不证明hold时序，hold的真实安全JSON和代码严格断言才是行为证据。

代码差异限于workspace保存生命周期/epoch、journeys完成屏障、测试及规范；无Go HTTP/schema/鉴权/依赖锁/固定入口/Connector/权限/时限变化。复用operation、epoch、busy、read/render；不引入公共刷新协议。QA没有修改业务实现。

## 3. 缺陷及进展

见defects-qa-08.md：R30-SYNC-01已证范围Closed；R28-04 Open/P2不更级/不关单。本轮未发现释放后持续错误，故不以测试暂态升级持久P1，也不宣称修复旧根因。没有研发整改需求；下一失败必须保留首个实际错误，返回development，不能无限全套重跑求绿。

## 4. 证据及真实性

- 固定入口仍make verify→go run ./tests/verify→isolated make verify-gates。host-tests-07/state.json实测cold start：node_modules_before=false、npm_cache_before=false、NODE_PATH=unset、manual_preinstall=none；candidate/source前后未变。
- 项目完整verify.log为337036字节、SHA256 9197fff5ef1e77fd658f1a5a4c2a9e89b33598808f33e8708919967cee457c3f，QA逐字节与state比对并原样归档。平台tests7回执log.bytes=337132、truncated=false，是另一输出口径；96字节差额不推断成缺失内容或人为补齐。未声称重新执行或读遍平台retained输出；本轮使用完整项目文件和实际Connector元数据。
- 宿主依赖：Go1.25.6、Node v26.0.0、npm11.12.1、Playwright/core1.59.1 candidate/node_modules；Chrome for Testing147.0.7727.15，三个probe实际browser.version()均记录147.0.7727.15。
- 公共format/module/vet、213/213全量Node、全包go test -race -count=1 -v ./... -parallel=4、build与真实HTTP均exit0。旧HTTP21个method红/12个query绿分类、零软预算与有效UTC旧红由历史重构捕获，不是当前失败；无删门禁/降race。日志中的runner cancel/预期历史红属于测试验证隔离及历史分类，不把它们抹掉。ld LC_DYSYMTAB警告保留，不声称零警告。
- 本QA自己执行的75/75 Node为Contract，完整日志independent-node.log；源码before/after和material核验是本节点实际执行。最初直接go run source检查因默认HOME Go cache权限报错，未跑测试；按原Makefile的workspace GOCACHE/GOTOOLCHAIN/CGO环境完成核对，不改权限/依赖或冻结配置。
- 26个宿主文件含完整log/state、三宽度model-save/continuous/summary/fault/editor/visual JSON和3张安全模型页截图；archive-index.json列出每个具体路径/bytes/SHA。known-secret及本机home路径扫描无匹配；保留原JSON不替换scope/UTC/原错误。

## 5. 未执行和环境限制

QA本节点没有重跑make verify或监听浏览器：native限制已有同Run原始证据、环境未变，项目明确要求消费host-only正式tests而非扩大沙箱。独立动态QA仍Blocked，静态browser.check不适用于Go后端，未使用。对应候选强制结果已由同轮同版宿主原层回归覆盖，关联S01–S06，但Blocked行不改Pass。

TestConsoleClockRecoveryWorker顶层明确skip，为Not Run；父ClockRecoveryInProcess及ActualHTTP启动两次真实worker（重启/备份恢复），必须得到CLOCK_RECOVERY_PASS才通过。该同轮等价恢复证据满足门槛，不冒称顶层worker未skip。其他供应商/Owner预览及新正式部署Not Run；原PR29版本和schema3部署只作进入基线，非本候选发布证据。

## 6. 阶段边界

本轮接受的是模型保存读回契约及现有产品受影响回归；不追加design-v0.2.0能力、不验证GPU/支付/多区域运营、不修改Platform或复制部署控制器。精确已验源码尚未形成新commit，不可用现HEAD main作为已修复版本发布。report应在本Run pr.md如实写结论；实际提交/推送/PR/合并/Owner正式prepare、部署、Deployment、health/version/schema3与用户入口由既有授权Connector/协调者完成。

## 7. 环境恢复

宿主state：cleanup=owned_process_group_cleanup_attempted; candidate_removed、owned_group_absent=true。continuous JSON：isolated_directory_servers_contexts；21组本Bugcase cleanup=true。host临时安装/route/context已清理；node_modules_after/npm_cache_after为临时候选内冷安装测量，不是泄漏到源码树。QA仅生成证据/报告并运行无监听Node及source检查，无用户DB、账号、master key、checkpoint、供应商或preview更改。当前checkout无node_modules；使用原入口指定的已有.cache，不删除或提交缓存；历史失败原样保留。

## 8. 质量趋势

较tests5：四组真实hold证实save早完成的红，变为tests7三套各四组hold且save不早完成的绿；最终状态两轮均正确，定位为完成契约而非持久数据错。新增GET失败及迟到相反操作原层回归、operation关联终态、VM/原summary负例资产，质量提升。较旧tests18：边界观测更完整，但原错误仍无法归因，R28-04未关闭；单轮绿不能证明永不复发。独立动态、供应商和正式preview风险未消除。

## 9. Go / No-Go及下游约束

Go with known issues：无当前证据P0/P1，完整同版强制候选门禁已跑且通过；既有P2不默认阻塞。handoff→report，不因未执行真实供应商或native监听限制伪造通过/额外索权，也不回development无缺陷重跑。此结论只允许精确候选继续已授权交付；正式prepare/部署若红或源码不一致必须No-Go并返回原责任节点。Owner独立实际登录预览仍需后续完成；不得在PR/部署回复写“已全面验收”或“旧问题根治”。
