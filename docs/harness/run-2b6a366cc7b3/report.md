# Harness模板升级后的同版QA缺章返工、确认与草稿PR发布交接

## 真实Issue、目标与当前结论

关联 big91987/model-relay Issue #48（https://github.com/big91987/model-relay/issues/48），明确替代 #44；本报告不提前关闭Issue。Run `2b6a366cc7b3d48870e1b887757f229f`，冻结Workflow revision3，report执行9。目标是验证项目自有GitHub入口、共享OpenSandbox角色交接、固定tests、独立QA真实拒绝返工、正式操作者确认和固定publisher草稿PR的连续性，不新增Model Relay业务、不重做产品设计。

已确认：本Run两轮原make verify正式exit0；QA04实际缺章拒绝后development05补齐，QA07完成完整原日志独立复核并最终PASS；approval8已按授权保存有效next确认。报告只交固定publish生成真实草稿PR；真实PR、远端同版核验、材料保存后正式生命周期释放仍是后续责任，M1/完整交付不能在此宣告完成。没有合并或部署授权。

## 相对真实main的全部改动与历史继承

基点为可信候选 `codex/issue44-reviewed-candidate`，固定HEAD `30b4e13c19e36d51b72ef1186468d6adf39ac822`；任务指定main `6d2552d331fd7a0ec8d205f5d18ba3bcd34d6995`，本地Git确认main是HEAD祖先。此处不是report联网刷新远端main的结论；固定publisher应核实际目标与远端差异，保留升级而非覆盖回旧模板。

相对该main已有94项提交内变更（33160插入/37删除），逐文件A/M清单为 `docs/workflow/runs/2b6a366cc7b3d48870e1b887757f229f/qa-07/full-review/inherited-main-diff.txt`，随本轮提交。它们来自#47材料接续候选，不是本轮新写、不重复塞入本轮files：

- 11项测试设施源码：`internal/relay/routing_http_test.go`、`tests/review-history/main.go`/`main_test.go`、`tests/browser/journeys.mjs`/`stream-capture.mjs`/`stream-capture.test.mjs`、`tests/verify/main.go`/`main_test.go`/`gate.go`/`resources.go`/`gate_resources_test.go`。历史回归分离严格JSON stdout与诊断stderr，路由夹具保留真实HTTP、40并发SWRR、503回退、冷却与凭据哨兵；固定Go runner保留24分钟累计预算、父截止/短预算、进程组清理、冷候选和同版receipt、只读资源采样及非Linux边界；SSE采集克隆同次原生Response、finally恢复fetch，保留唯一dispatch、usage/账本和缺DONE反例。没有产品业务实现变更；旧授权预算校准已继承，本Run未增限。
- 83项历史文档/原层证据：Issue42/44接续说明、`docs/harness/issue44-delivery/qa.md`/`report-input.json`/`report.md`、完整旧文档 `docs/harness/opensandbox-run-sharing.md`、旧Run `744c4329facce31be37c97e7caa2fea3` 的研发、原失败/返工/测试/QA取证以及更早Run的两项Harness修复说明。原可编译源码归档被构建器递归发现的失败、后续非编译文本映射、预算超限和SSE设施失败均保留，不删除红灯、不排除正常测试。详情以94项清单和原Issue44报告为准。

#44旧Run已正式取消释放，17步历史与publish失败回执保留，不恢复旧Run；旧tests12/QA13不是本Run结果。#47只通过正式API原字节保存90项原选定文件及QA/报告作为固定候选，不是最终Pipeline PR、不合并。94项是相对本次已升级main的Git差异，90是历史选定文件口径，不能混为一谈。#44原publish17因模板自己的测试字符串误触扫描而失败且没有Pipeline PR；旧报告里的“待完成”保持原时态。本候选继承的旧测试设施有旧QA13原审查依据，本轮QA07又审查当前设施与本Run完整回执，不以历史放行替代本轮独立结论。

已合main的模板升级来自唯一维护源bc43314（完整提交见QA），经PR #46合入；146项模板逐文件匹配，不回退、不放宽凭据规则。平台afebd27确认上下文修复和后续日志导出revision8ad93945da127d5da7cdeefde68b92dccf9a2d15是平台维护输入，不是本PR产品代码，也不改变冻结Workflow revision3。

## 本轮实际新增与失败恢复

1. 新增 `docs/harness/opensandbox-publish-recovery.md`，覆盖项目模板安装、共享工作区/角色交接、固定验证与独立QA返工、正式确认、发布失败、材料保存、停止/恢复/环境释放。研发02依授权故意遗漏最后一章，首轮8166字节/SHA256 `35e5eb5af37a83c81344cbad95cb89b866193179605ba5b9c772dc2dc1562be9` 保存在first-round原件；已有opensandbox-run-sharing.md未改为缺章。
2. tests03通过不等于文档验收通过：QA04实际拒绝QA04-D01并正式revise；development05只补齐第七章及取证，原六章为原字节前缀。修订文档12999字节/SHA256 `d1a7112600467ae85d9af9000e3349f78ee9590e0317fc935cd4e1560d5f4209`，返工映射、patch.txt、首轮原文和QA04拒绝全部随交付保留。
3. tests06再次固定验证通过后，QA07因外层中段未完整审阅/可靠归档而真实拒绝并正式blocked。维护者旁证未替代QA。平台增加正式read_command_output保存能力后，同一QA7自己导出、全读、核摘要及内外层关联，关闭QA04-E01/QA07-E01和能力阻塞；旧拒绝/blocked原件仍保持当时OPEN，不把用户消息或导出修复冒称额外研发返工。
4. QA自身早期python调用/Go PATH缺失，以及完整复核脚本错误要求两种独立毫秒时间绝对相等的失败均保留；后者按真实1ms量化边界纠正检查脚本，未改日志、产品、原tests或预算。report9首次原生serverOverloaded失败及同一会话一次有界恢复，依用户正式输入记录于report-09/approval-and-recovery-context.md；首次原生容量失败前无command/MCP/报告/发布副作用；第一次恢复已生成三份报告后再遇第二次容量错误，这次已有文件写入副作用、validation/handoff未完且未publish。至少五分钟后的最后一次有界恢复先核三文件原始字节/摘要全部一致，仅补记本事实并完成校验，详见上述上下文转录；未重做tests/QA/decision，未换模型/权限/预算或手工Git发布，若再遇容量失败则保留现场停止、不连续重试。

## 已验版本、完整回执与结果

版本 `development05-lifecycle-rework` = 上述HEAD + 本轮修订文档与证据，不能仅凭HEAD代表实际工作树。163项源码/测试/锁与合同清单组合SHA256为 `991c79ca919b3d0ccf148c066c3b24d993090fcb78ab864c0e4637cfff8040ba`；report重新只读核各文件字节/摘要、组合摘要及模板，没有执行产品构建或重跑同版验证。

| 本Run固定轮次 | 主attempt | 正式结果与原件 |
| --- | --- | --- |
| tests03 | 20261014T192852Z-1089328441 | make verify exit0；外层1174629字节/SHA256 754f5242f9a567dd52b2f40b10667b34c18819099b6094d28940e48f661d310e；内层1144558字节/SHA256 afdbb354a9181d8fd950d2d16159283c9379d56552269eadbe4cd3f283ed101a |
| tests06 | 20261014T195523Z-3543396171 | make verify exit0；外层1177433字节/SHA256 ec2540257de3d6a6c0717f165a045ceee8712b30380bb1cba4c56691e4bf6caf；内层1147466字节/SHA256 61dadc32700ffe860c9a03800b30e13b6c406f2a56dd43bedc1d71c14c3ba797 |

两份外层各选唯一 `workflow-output/<本Run>/command-3.log`/`command-6.log` 原件，truncated=false；对应完整内层/state在qa-04/tests03与qa-07/tests06。QA7实际完整独立核验，report仅复核文件身份并采用其结论，不冒称report重新逐项审阅百万字节日志。去除真实平台事件后的全部载荷仅作CR/LF传输关联，未改原档字节；completion本身没有退出码，正式exit0来自平台固定回执及state交叉核对。

原入口make verify→Go隔离runner→原verify-gates，格式、锁模块、vet、前端283/283、Go race七包、构建与实际Go服务浏览器migration/legacy-ui/routing/console/full-console均通过。QA7另有6项限定runner/资源/路径/取消及3项真实受控HTTP回归通过，不当作替代全套。每轮25历史FAIL、15注入FAILED和3个非零fixture gate均由QA完整归因至已通过的预期红灯测试，当前门禁无失败，原失败全文不删。

内层1440000ms/外层1800s未改，冷候选node_modules/npm cache起点false、manual_preinstall none、NODE_PATH unset，前后源码一致、owned group清理与candidate_removed均核对通过。8Gi/2CPU/GOMAXPROCS2及无OOM增量来自采样；共享工作区不是全局冷环境，采样不作独占资源因果论断。

最终QA为 `harness-output/qa.md`，11472字节/SHA256 `e13253c86426e806b63cb7229c066d3e1e4490717f8fa11330b629856b1a5cf3`，与qa-07/full-review/qa.md原字节一致；下方完整附原文。独立QA署名/判断由QA节点形成，report未代QA改结论。

## 已保存的操作者确认与报告边界

直接采用平台传入的approval8有效next结果：Harness Builder依据用户既有授权通过正式Owner API确认，不是用户本人点击。实际stop/重复stop均受理，正式GET确认stopped且全部原生进程stopped；同Run同seq8 resume回waiting，未新增步骤，前7步及冻结definition/input/environment逐值保持，恢复后QA/文档/清单/QA04拒绝字节不变。确认前30份原件4794292字节已通过正式文件API保存。重复stop的去重/无新步骤依据为该保存确认，report未重演副作用或自行独立核控制API。

首次decision因不匹配route被HTTP400拒绝且无result，随后依冻结next出边更正才有效，未改图。确认结果及容量恢复输入作为明确标源的上下文转录保存；不伪装成原始API回执，不把approval当命令日志，不再次要求用户点击。

## 文件选择、归档、限制与下一责任

files采用QA推荐81项具体本轮文件，加report.json、report-09/report.md、approval-and-recovery-context.md及validation.json，共85项；每项≤5MiB（非总包5MiB），不重列94项已提交继承文件。必要原层日志全文、原失败/blocked、首轮/返工、最终QA/来源映射均包含；提交项不含构建目录、node_modules、凭据、数据库、原生会话历史或可编译源码副本。证据以原路径/大小/摘要清单、.txt文本/补丁与不可变原层日志归档，和构建输入隔离；保留原归档失败及更名映射，不改源码字节、不排除正常测试。旧SHA清单保持不变，旧harness-output/qa.md按supersession-map映射到qa-04/qa.md核对，而不是拿新QA误校旧摘要。

报告侧只做JSON/具体路径/大小/摘要/源码及模板身份/原全仓离线扫描校验，结果在report-09/validation.json；这些不是再次tests/QA或发布通过证据。未读Git凭据值、未改扫描规则/锁/测试/预算/模板/产品。本轮只有文档及交付证据新增。

本报告只对现有Issue和历史PR使用可访问GitHub关联；上述本轮产物路径是具体仓库相对路径，待固定publisher提交，不假装旧main或未发布路径已有可访问URL。固定publisher须创建本Run真实草稿PR，核实际远端提交与85项选定文件、同版文档/源码/QA及确认一致，并保留main模板升级。未知发布结果先查原操作，不盲重试；不得Agent手工Git发布、合并或部署。材料实际保存后由授权角色走正式生命周期释放，不靠删除目录。受理/排队/恢复不算最终交付。

未验真实供应商、支付、外部身份、Mac实机；受控HTTP/浏览器不是供应商接入证据。角色模型、完整4小时/40步控制面配置来源于冻结/维护上下文，QA未独立核全部；无资源或权限扩张。真实草稿PR、远端同版复核及最终释放仍由下游完成并据实回执。

## 同版独立QA7原文（完整原字节附录，保留QA写作时态）

# 独立QA最终复核：PASS，仅放行正式approval

2026-10-14；Issue #48；Run `2b6a366cc7b3d48870e1b887757f229f`；QA执行7接续；冻结Workflow revision3。版本 `development05-lifecycle-rework`；HEAD `30b4e13c19e36d51b72ef1186468d6adf39ac822`。**文档返工及本Run固定验证证据门槛通过，可交approval；不表示正式确认、真实草稿PR、远端同版核验、环境释放或产品全范围通过。**

## 验收依据、版本和继承关系

以本Run用户冻结任务1–7、实际QA04拒绝、研发05返工、固定tests03/06原件和QA07阻塞记录为依据。没有新产品设计/业务实现，没有修改产品、原测试、锁文件、模板、凭据规则、预算或准备。当前main `6d2552d331fd7a0ec8d205f5d18ba3bcd34d6995` 是HEAD祖先；146项模板逐文件匹配bc43314标准导出，保留已合main模板修复。相对main的94项继承变更列于inherited-main-diff.txt，来自#47材料接续候选，不冒称都是本轮新增。下游报告应覆盖它们与本Run新增交付的完整关系。

#44旧Run已取消释放，旧tests12/QA13及#47仅历史/材料接续，不当成本Run通过证据。旧docs/harness/issue44-candidate-continuation.md、issue44-delivery/qa.md/report.md、旧Run原件保留其当时“待完成”状态，不能倒写旧结论。已有完整opensandbox-run-sharing.md与HEAD原字节一致。本轮交付新文档和证据，没有覆盖模板升级。

当前源码163项逐文件原字节/大小和组合摘要独立复核，仍为 `991c79ca919b3d0ccf148c066c3b24d993090fcb78ab864c0e4637cfff8040ba`；cmd/internal/tests/.trellis/spec无额外构建输入，跟踪产品差异为空。当前新文档12999字节、SHA-256 `d1a7112600467ae85d9af9000e3349f78ee9590e0317fc935cd4e1560d5f4209`，与研发05快照一致。

## 原失败、返工与阻塞的真实接续

QA04-D01：首轮新文档只有六章，缺停止/恢复/释放，QA执行4真实拒绝并交研发05补齐，不把旧Run返工当本轮。现七章，原六章作为原字节前缀保留，首轮快照摘要35e5eb5af37a83c81344cbad95cb89b866193179605ba5b9c772dc2dc1562be9未变。新增章覆盖正式安全点、重复stop实际去重、同Run同节点resume/材料保持、未知结果先核原操作、正式decision、保存后正式生命周期释放及不恢复旧Run等边界；明确这些操作说明不等于执行完成。缺章实质修复，关闭QA04-D01。

QA04-E01/QA07-E01：此前外层中段未完整审阅/归档，QA07实际拒绝且随后正式blocked；维护者完整阅读/摘要仅作旁证，没有替代独立QA。当时development handoff因用户输入待送达未受理，不冒称发生了额外研发返工。QA04拒绝、QA07拒绝、capability-blocker.md及旁证记录原字节保留。

本次平台新增的原生导出能力实际可用：本QA分别调用read_command_output(seq=3/6,save_to_workspace=true)，取得固定共享工作区路径、字节数、摘要、truncated=false，再亲自读取整份UTF-8原档并程序化完整核验。没有由维护者补造工作区文本，没有拼接呈现省略内容，没有重跑同版make verify。平台部署revision8ad93945da127d5da7cdeefde68b92dccf9a2d15来自用户维护输入；实际可独立观测的是两次正式导出成功及原文件内容，不混同冻结Workflow revision3。

## 两轮完整内外层原件独立复核

- seq3正式导出：workflow-output/2b6a366cc7b3d48870e1b887757f229f/command-3.log，1174629字节，SHA-256 `754f5242f9a567dd52b2f40b10667b34c18819099b6094d28940e48f661d310e`，truncated=false；init sandbox48b9533c70bf4d6285ac76aca19dd530，342条ping及一个真实completion，execute_time1024495ms。对应内层4588行/1144558字节，SHA-256 `afdbb354a9181d8fd950d2d16159283c9379d56552269eadbe4cd3f283ed101a`，原attempt20261014T192852Z-1089328441。
- seq6正式导出：workflow-output/2b6a366cc7b3d48870e1b887757f229f/command-6.log，1177433字节，SHA-256 `ec2540257de3d6a6c0717f165a045ceee8712b30380bb1cba4c56691e4bf6caf`，truncated=false；init sandbox354f6de569b44a4fa596e8576fc55edb，341条ping及一个真实completion，execute_time1020683ms。对应内层4592行/1147466字节，SHA-256 `61dadc32700ffe860c9a03800b30e13b6c406f2a56dd43bedc1d71c14c3ba797`，原attempt20261014T195523Z-3543396171。

两份外层全文去除已解析的init/ping/completion事件后，剩余**全部**载荷与对应完整内层日志在仅去CR/LF的传输比较下完全相同，外加精确原make verify命令/attempt前缀，无未解释额外内容。该归一化只用于内存关联检查；两份外层原档、内层原件绝未改写。不能声称内外层原字节相同或用内层取代外层。完整原档中无回执呈现省略标记。

exit0/timeout1800来自正式传入平台tests回执；completion事件自身不含退出码，不能仅凭完成事件推导exit0。state gate_exit/exit_code均0，内层预算1440000ms不变，seq3/6 gate execute分别1000606/1004725ms，真实内外层均未超限；候选冷起点node_modules/npm cache false、manual_preinstall none、NODE_PATH unset，source_before/after一致、candidate_unchanged true；辅助命令退出0，owned_group_absent true、candidate_removed。共享工作区不是全局冷起点。8Gi/2CPU/GOMAXPROCS2与无OOM增量来自state资源采样，不能作独占资源因果归因。

两轮设施仍为原make verify→Go隔离runner→原verify-gates，包含格式、模块校验、vet、前端283/283/fail0、当前Go race包、构建以及真正Go服务浏览器migration/legacy-ui/routing/console/full-console全部ok=true。不是静态browser.check。

完整归因：每轮25条历史FAIL输出均属旧TestHealthMethodsHTTP（21方法子项、4父项），有精确预期红灯及12查询恢复通过证明；每轮15条FAILED诊断均位于通过的TestInventoryCancellationReceipt/TestFixedRunnerReceipts故障注入，3条非零fixture gate也全部在已通过的故障测试中。完整当前顶层测试无FAIL/not ok/panic/DATA RACE，routing end均ok=true；7个当前Go包结果通过。原失败全文保留，未过滤红灯或排除正常测试。上述证据关闭QA04-E01/QA07-E01及原能力阻塞，**现在关闭**，不修改旧记录中的当时OPEN状态。

## 同版独立行为复查和本次检查失败

QA07此前实际运行6项runner/资源/路径/取消限定回归和3项真实受控HTTP回归（管理与普通请求、SSE首片屏障/usage、console quota/billing/SSE），9项均通过，两条命令退出0，完整原输出和命令文件保留；当前源码摘要未变。本次不为日志取证重跑同版tests，不把限定回归充作固定全套。

QA07最初python命令缺失和Go PATH缺失原失败记录仍在，只纠正已有工具调用，无手工预装固定依赖。本次独立复核脚本最初错误要求平台execution_time与毫秒时间戳差绝对相等而失败；两轮原件均相差精确1ms（独立毫秒采样/量化边界），后以最多1ms量化边界和原1800秒上限检查通过。初始脚本/退出1/完整stderr另存，未改原日志/测试/预算，不能把检查脚本纠正当产品修复。

## 未验、遗留及approval边界

- 已验范围仅本任务文档返工、版本/模板/历史完整性、两轮本Run固定受控验证完整证据及限定独立HTTP/runner复查。没有真实供应商凭据/响应证据，不验支付、外部身份、Mac实机；本次未重跑浏览器全套。
- 角色模型会话、4小时/40步及平台修复提交来自冻结/维护上下文，未独立核对全部控制平面配置。无额外权限或Git凭据授权；GITHUB_TOKEN/GH_TOKEN/GIT_ASKPASS仅核存在布尔值且均false，未读值。
- 待approval责任角色核正式安全点、stop/重复stop去重、同Run同节点resume、材料保持，再正式decision；report应直接使用平台保存的确认结果，不冒称用户本人点击或读取approval“命令日志”。
- 待report覆盖相对真实main继承94项和本Run新增、附同版QA原文；固定publisher产真实草稿PR并远端核同版。未发布、未合并、未部署，不把本QA放行当完整交付。
- 材料实际保存后方可正式生命周期释放；本QA未操作stop/resume/decision/publish/释放。已保存路径只是仓库相对产物，不冒称GitHub可访问链接。

## 归档与历史摘要映射

两份完整外层日志各选唯一workflow-output原件，不另复制同日志；内层必要全文与state沿用qa-04/tests03.*和qa-07/tests06.*，源attempt与正式导出回执由full-review.json/export-receipts.json对应。新最终harness-output/qa.md与本报告原字节一致；其旧QA04拒绝已完整保存在qa-04/qa.md（原SHA1d25ef3827de7846b4af52493832a1dba74ba2e3fb295ef118903d1c45f0f1f4），supersession-map.json明确旧路径→不可变历史原件及新报告映射。旧SHA256SUMS不修改；校验其旧harness-output路径时使用该历史原件，不把新报告误判成原件被篡改。

源码仅使用原路径/大小/摘要清单和非编译文本/补丁，未放入可编译副本。新报告/映射/清单及原日志经过原有无网络全仓发布扫描，未改扫描规则。files列具体文件且每项≤5MiB；推荐交付清单仅列本轮未提交必要文件及显式最终QA，不重列已提交旧材料，仍给report/控制回执预留100项上限空间。

## files（本次复核与必需原件）
- `docs/workflow/runs/2b6a366cc7b3d48870e1b887757f229f/qa-04/tests03.state.json`
- `docs/workflow/runs/2b6a366cc7b3d48870e1b887757f229f/qa-04/tests03.verify.log`
- `docs/workflow/runs/2b6a366cc7b3d48870e1b887757f229f/qa-07/full-review/SHA256SUMS`
- `docs/workflow/runs/2b6a366cc7b3d48870e1b887757f229f/qa-07/full-review/checks.exit.txt`
- `docs/workflow/runs/2b6a366cc7b3d48870e1b887757f229f/qa-07/full-review/checks.stderr.log`
- `docs/workflow/runs/2b6a366cc7b3d48870e1b887757f229f/qa-07/full-review/checks.stdout.log`
- `docs/workflow/runs/2b6a366cc7b3d48870e1b887757f229f/qa-07/full-review/delivery-files-recommendation.json`
- `docs/workflow/runs/2b6a366cc7b3d48870e1b887757f229f/qa-07/full-review/export-receipts.json`
- `docs/workflow/runs/2b6a366cc7b3d48870e1b887757f229f/qa-07/full-review/files.json`
- `docs/workflow/runs/2b6a366cc7b3d48870e1b887757f229f/qa-07/full-review/full-review.json`
- `docs/workflow/runs/2b6a366cc7b3d48870e1b887757f229f/qa-07/full-review/independent-full-review.py.txt`
- `docs/workflow/runs/2b6a366cc7b3d48870e1b887757f229f/qa-07/full-review/inherited-main-diff.txt`
- `docs/workflow/runs/2b6a366cc7b3d48870e1b887757f229f/qa-07/full-review/initial-checks.exit.txt`
- `docs/workflow/runs/2b6a366cc7b3d48870e1b887757f229f/qa-07/full-review/initial-checks.stderr.log`
- `docs/workflow/runs/2b6a366cc7b3d48870e1b887757f229f/qa-07/full-review/initial-checks.stdout.log`
- `docs/workflow/runs/2b6a366cc7b3d48870e1b887757f229f/qa-07/full-review/initial-full-review.py.txt`
- `docs/workflow/runs/2b6a366cc7b3d48870e1b887757f229f/qa-07/full-review/qa.md`
- `docs/workflow/runs/2b6a366cc7b3d48870e1b887757f229f/qa-07/full-review/supersession-map.json`
- `docs/workflow/runs/2b6a366cc7b3d48870e1b887757f229f/qa-07/tests06.state.json`
- `docs/workflow/runs/2b6a366cc7b3d48870e1b887757f229f/qa-07/tests06.verify.log`
- `harness-output/qa.md`
- `workflow-output/2b6a366cc7b3d48870e1b887757f229f/command-3.log`
- `workflow-output/2b6a366cc7b3d48870e1b887757f229f/command-6.log`
