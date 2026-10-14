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
