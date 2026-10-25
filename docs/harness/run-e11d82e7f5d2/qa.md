# Issue63 / Issue65 同版交付 — 独立 QA 最终复核

Run `e11d82e7f5d28afd80f64477377f1c63`，当前节点QA执行6；2026-10-25 UTC。

## 1. 正式结论

**通过已约定产品与交付证据验收，交 approval 正式确认。** 当前候选HEAD `65cc0ad786123a7e38c6118ea450db42b7bc7ddf`，171件源码SHA256 `6316a232322de48fdce31a4420a872495f9e7ea2237c36e9e2742a4e3566a7e0`。QA4已实际完成的完整门禁、独立Go/真实Go浏览器及两个原脚本正例有效；QA6仅独立核历史补件及同版一致性，**没有再次执行全量产品实操，没有重新实现产品**。

先前No-Go是当时真实证据不足结论，不删除或倒签：原始首次报告保存在 `harness-output/qa-original-recheck/prior-qa-evidence-insufficient.md`；人工补交脚本后仍缺旧失败原件的QA4报告原字节保存在 `harness-output/qa-history-review/prior-qa4-supplement-no-go.md`。两组原manifest与报告路径映射均保留，本次122条历史manifest记录按映射验证通过。QA65-E01（原脚本与原件重放）已于QA4补件实际执行关闭；QA65-E02（旧失败原日志、清单与版本链）由本次17件原件核验关闭。没有发现需产品返工的新实质问题。

本次放行不是产品PR、发布许可或上线完成；仍须approval正式确认→report→固定publisher创建产品**草稿PR**，止于draft，不合并、不部署、不改变原预览。

## 2. 本次补件来源与完整性

Owner经正式人工回退入口将本节点交QA；保留研发5取消语义，不伪造研发→QA工具交接，不改冻结图。Builder按普通选择委托的正式回退/确认另行记录，不能称用户亲自点击或无人介入。

本次输入为平台受控归档产物引用接口复制的旧Run `ba434da8039cca16f6829481112e3d68`、source_seq=9留存原件。**用户回执列明17件，本次实际17件全部逐size/SHA匹配**；来源路径→当前`.artifact`路径→bytes/SHA完整列于 `harness-output/qa-history-review/result.json`。不解包、不执行原件，不改原源码字节；文本源码仍为`.go.txt.artifact`，不进入Go编译树。21条原QA4/QA9 manifest交叉引用与本次原件匹配。

受控接口与同Owner授权/SHA来源是本次入站事实，本节点实际核验的是当前可读原件。旧QA9 retention-check自述为**本地交接前核验，不是远端全新下载**：209条清单/212件handoff；本节点没有声称独立下载核验旧212/671件全量。手工补交四文件、此次正式回退及17件引用均属人工介入；不称跨Run自动恢复或在途冻结升级已经实现。

## 3. 历史真实失败与修复链

|阶段|本次直接核得的原始事实|版本绑定|
|---|---|---|
|旧QA4 late-after-failure|514bytes，真实FAIL，exit_code=1，EOF；旧请求acceptance=accepted、清池后samples=1|旧QA4 manifest原路径/bytes/SHA一致；不能当当前候选失败|
|旧QA4 sequence-watermark|553bytes，真实FAIL，exit_code=1，EOF；较新失败后旧成功仍accepted|旧QA4/QA9 manifest同时保留这份失败原字节|
|旧研发5 / tests6|verify.log1069973bytes、state4135bytes、Connector1074864bytes全部读入并验SHA；主测试TestRoutingPassiveCooldown: cooling candidate visited|171件SHA `7f98fee14c6ca5337d981b66b0d7dd1ed88579123c492674c7ed1ed3da8f69e0`|
|旧tests6退出口径|**gate_exit=2、runner state exit_code=1、外层Connector=2**；非超时，source前后相等，所属组已清理|原state/log/Connector一致，不将任一口径改为0|
|旧研发7|仅internal/relay/routing_test.go从10ms墙钟竞态改9/10ms受控时钟；冷却内不派发、到期恢复及调用次数断言保留|source-manifest.previous_source指向7f98…，final171件6316…；逐文件匹配当前|
|旧QA9|original-negative-recheck.log1135bytes，两个原TestQA63通过，EXIT_CODE=0、EOF；迟到invalidated_by_failure、samples=0|QA9 manifest及retention source均6316…|
|当前gate / 独立QA|下节已执行绿结果，不以历史红替代|同版171件6316…，prepared HEAD65cc0ad…|

旧tests6主gate246.890903秒后失败，不能当完整gate成本或绿色；日志里的health-history历史负例及runner失败/取消/超时夹具为预期覆盖，不是额外主产品失败。不删除原失败，不重建假失败，不靠增限、弱断言或业务DB修改求绿。

旧Run HEAD f06f05…记录的是当时基线加工作树；本Run HEAD65cc0ad…是保留候选提交。两者HEAD不同不冒称相同commit，最终171件全源码摘要一致才构成同版绑定。新工具包来源commit643c128dfbde94b41aa6a2ba4c0a20d944c036c4的150条受跟踪文件本次重验；产品main7709fe2734c370238a49bbb98efac9eddba81c7a→候选真实49件差异及每件摘要重验，含.trellis项目契约，不拿prepared HEAD空diff做发布范围。

## 4. 本Run已实际完成的验收（QA4；QA6未重跑）

- **固定make verify**：tests执行3主gate **1403.175秒，exit0**；内部1500/外层1800秒；冷隔离无node_modules/npm cache起点、source/candidate前后不变、所属组清理通过。全部原阶段、唯一2480请求实验不变。当前完整Connector输出1177020bytes/SHA b72250062fb68152890bc018d374a4c1bbd282f975b9a80182bda0dc0351df82，含execution_complete；本次重新核验全文摘要与gate原verify/state，不依赖聊天截断展示。
- **独立Go**：主组30项、安全组6项均exit0；普通/SSE同池及跨池失效恢复8组合、TTL59/60/61、配置代次/水位、新候选/输出判定、未知用量、最多三attempt、两种策略非零calls/RPM/TPM/并发/金额预算及tenant/key授权隔离、SSE取消与已输出不重放。命令、真实stdout/stderr、退出码和Bash时间见qa-current/commands.md及原log。
- **原始QA负例正例重放**：QA4人工补四文件后，每个原overlay单次运行；late exit0/5.025秒，旧请求invalidated_by_failure、pool_samples=0；watermark exit0/5.147秒，较新accepted、后失败not_success、pool_before_late=0、旧请求invalidated_by_failure。前后171源码均6316…。本次将这些当前绿与补入旧红连成可审计链，而非再跑两个测试。
- **真实Go浏览器1280/390**：原生双服务/双候选配置，保存、重入、重启及切回weighted；普通/SSE同ID日志/选择/观测/费用一致，单attempt、单费用项fee14；16张安全截图和两JSON留存。最终1280 ID EUiMWsEy315hWDeFigPWOQ / 0LfqO5_mO8eWvWLNk82l_w，390 ID 42OQ9ZIDcrG5b0YLrYV-pA / zVEkA4Au7TGQP8kXZeJqGQ。
- **独立复算而非重新发送请求**：14份JSONL、2480唯一ID=2240核心+160心跳+80尾部；全200/success、单attempt/费用条目。mean/p50/p95、80%/35%/75%门槛及40次适应、心跳/尾部语义均通过。两轮普通/SSE稳态快端50%→90%、均值改善52.5718%—52.7667%，恢复计分90%；p95仍约101ms、探索保留慢尾。普通首body与SSE首有效输出增量，不是完整生成、供应商TTFT或生产收益。原Run52.5077%—52.6653%仅作历史分列，不替代本轮结果。

## 5. 执行边界、遗留问题与未验证范围

- 保留初次缺件、Node loader预检失败、apply_patch不可用、重建负例与原件负例的来源差异及真实命令噪声。QA4局部UI两次并发均exit0，共8实际请求；第二轮同名JSON/截图覆盖第一轮，只留最终匹配16图，不声称两套截图都在。QA6没有再跑UI或产品测试。
- 本环境无apply_patch工具/可执行文件；不安装依赖、不改权限，以shell/Node仅写QA证据和报告。无产品、冻结图、标准断言、配置DB或原始入站文件变动。历史manifest/result中的No-Go文字保留为历史，当前结论以本报告及qa-history-review/result.json为准。
- **真实供应商联调、生产性能与线上费用：Not Run**，无凭据，不读旧聊天密钥、不造供应商通过；受控HTTP/SSE是测试上游，不是产品默认假提供商。**原始浏览器trace：Not Run**，隐私原因，以安全截图/结构化收据替代。未做旧全量归档远端重新下载。
- **固定新publisher实跑、产品draft PR远端head/171源码/报告清单对齐、正式release与重新全新下载：尚未执行**，由后续固定阶段验证；不能以QA通过宣称这些成功。原预览未变更。无当前产品阻断项；上述明确不在本节点内的交付待办必须继续完成。
- 本节点无Issue63/65写入回执，不冒称已更新。Builder须把“QA65-E01/E02已关闭、QA正式通过、待approval/报告/固定publisher、仅draft不合并部署”和后续真实发布失败/状态用现有Connector体现到Issue65及关联Issue63。
- 总墙钟仍自**2026-10-25T11:14:20Z**算起；原QA9归档15:02:20.553507846Z、原approval停止15:05:47.043649725Z。本次补证等待/回退/归档属延续，不重置起点。阶段驻留含等待、工具与归档，不当Agent净耗时，源码修复与QA重叠不重复相加。旧tests3虽绿被拒、tests6失败、设计修订、有界握手前恢复、预算等待、Builder源修复及分支准备、Runner排队均须进入最终全过程分析。
- 原8turn上游确认input40078725、cachedInput38891136、output214330（含reasoning58969）、total40293055；cached为input子集，reasoning为output子集，不再加。这是原生计数不是发票；本Run token另记，本节点无新原生统计，不估造货币费用。

## 6. 正式确认与发布约束

交approval，由Builder依据既有授权核对当前版本/证据并记录**正式确认**，不能跳过或称用户亲自点击。随后report→固定publisher；最终标题 **【产品功能】支持同模型上游低延迟优先路由与选择诊断**，主需求Issue63，Issue65为同版交付接续，正文明确沿用已验实现而非从零重新开发。

report.json.files须以相对产品main的49件真实材料及本轮必要交付文档为基础，保留.trellis原路径，不只提交空diff或恢复旧源迁就publisher。harness-output/workflow-output是归档输入，不直接作为产品文件提交；标准publisher另复制最终QA/报告。本报告files是**QA归档清单，不是产品PR的report.json.files**。后续核精确远端head/171源码、完整证据及正式release/全新下载；发现实质源码变化须重新绑定新源完整测试，不得借旧通过放行。最终仅draft PR，禁止合并、部署、改预览或扩权。

## 7. files

下列96件为当前具体归档文件；artifact-manifest不自哈希，自身作为独立handoff文件保存。历史缺件结论、两轮失败原件、原环境噪声与当前绿色原日志均保留，不以目录代替文件。

- `docs/workflow/runs/e11d82e7f5d28afd80f64477377f1c63/development/implementation.md`
- `docs/workflow/runs/e11d82e7f5d28afd80f64477377f1c63/development/publication-files.json`
- `docs/workflow/runs/e11d82e7f5d28afd80f64477377f1c63/development/source-manifest.json`
- `docs/workflow/runs/e11d82e7f5d28afd80f64477377f1c63/qa/qa.md`
- `harness-output/imports/ba434da8039cca16f6829481112e3d68/9/harness-output/issue63-qa-recheck/manifest.json.artifact`
- `harness-output/imports/ba434da8039cca16f6829481112e3d68/9/harness-output/issue63-qa-recheck/original-negative-recheck.log.artifact`
- `harness-output/imports/ba434da8039cca16f6829481112e3d68/9/harness-output/issue63-qa-recheck/retention-check.json.artifact`
- `harness-output/imports/ba434da8039cca16f6829481112e3d68/9/harness-output/issue63-qa/late-after-failure.log.artifact`
- `harness-output/imports/ba434da8039cca16f6829481112e3d68/9/harness-output/issue63-qa/late-after-failure_test.go.txt.artifact`
- `harness-output/imports/ba434da8039cca16f6829481112e3d68/9/harness-output/issue63-qa/manifest.json.artifact`
- `harness-output/imports/ba434da8039cca16f6829481112e3d68/9/harness-output/issue63-qa/overlay.json.artifact`
- `harness-output/imports/ba434da8039cca16f6829481112e3d68/9/harness-output/issue63-qa/sequence-watermark.log.artifact`
- `harness-output/imports/ba434da8039cca16f6829481112e3d68/9/harness-output/issue63-qa/sequence-watermark_test.go.txt.artifact`
- `harness-output/imports/ba434da8039cca16f6829481112e3d68/9/harness-output/issue63-qa/watermark-overlay.json.artifact`
- `harness-output/imports/ba434da8039cca16f6829481112e3d68/9/harness-output/issue63-rework-1/manifest.json.artifact`
- `harness-output/imports/ba434da8039cca16f6829481112e3d68/9/harness-output/issue63-rework-1/source-manifest.json.artifact`
- `harness-output/imports/ba434da8039cca16f6829481112e3d68/9/harness-output/issue63-rework-2/handoff-manifest.json.artifact`
- `harness-output/imports/ba434da8039cca16f6829481112e3d68/9/harness-output/issue63-rework-2/source-manifest.json.artifact`
- `harness-output/imports/ba434da8039cca16f6829481112e3d68/9/test-results/verification/20261025T135914Z-1386142975/state.json.artifact`
- `harness-output/imports/ba434da8039cca16f6829481112e3d68/9/test-results/verification/20261025T135914Z-1386142975/verify.log.artifact`
- `harness-output/imports/ba434da8039cca16f6829481112e3d68/9/workflow-output/ba434da8039cca16f6829481112e3d68/command-6.log.artifact`
- `harness-output/issue63-qa/late-after-failure_test.go.txt`
- `harness-output/issue63-qa/overlay.json`
- `harness-output/issue63-qa/sequence-watermark_test.go.txt`
- `harness-output/issue63-qa/watermark-overlay.json`
- `harness-output/qa-current/archive-mappings.json`
- `harness-output/qa-current/artifact-manifest.json`
- `harness-output/qa-current/audit.json`
- `harness-output/qa-current/audit.log`
- `harness-output/qa-current/audit.mjs`
- `harness-output/qa-current/commands.md`
- `harness-output/qa-current/gate-experiment/comparison.svg`
- `harness-output/qa-current/gate-experiment/core-1-false-low_latency.jsonl`
- `harness-output/qa-current/gate-experiment/core-1-false-weighted.jsonl`
- `harness-output/qa-current/gate-experiment/core-1-true-low_latency.jsonl`
- `harness-output/qa-current/gate-experiment/core-1-true-weighted.jsonl`
- `harness-output/qa-current/gate-experiment/core-2-false-low_latency.jsonl`
- `harness-output/qa-current/gate-experiment/core-2-false-weighted.jsonl`
- `harness-output/qa-current/gate-experiment/core-2-true-low_latency.jsonl`
- `harness-output/qa-current/gate-experiment/core-2-true-weighted.jsonl`
- `harness-output/qa-current/gate-experiment/heartbeat-low_latency.jsonl`
- `harness-output/qa-current/gate-experiment/heartbeat-weighted.jsonl`
- `harness-output/qa-current/gate-experiment/summary.json`
- `harness-output/qa-current/gate-experiment/tail-false-low_latency.jsonl`
- `harness-output/qa-current/gate-experiment/tail-false-weighted.jsonl`
- `harness-output/qa-current/gate-experiment/tail-true-low_latency.jsonl`
- `harness-output/qa-current/gate-experiment/tail-true-weighted.jsonl`
- `harness-output/qa-current/gate-state.json`
- `harness-output/qa-current/gate-verify.log`
- `harness-output/qa-current/independent-go.log`
- `harness-output/qa-current/independent-ui-recheck.log`
- `harness-output/qa-current/independent-ui.log`
- `harness-output/qa-current/late-negative.go.txt`
- `harness-output/qa-current/lineage-missing-evidence.log`
- `harness-output/qa-current/overlay.json`
- `harness-output/qa-current/playwright-loader.mjs`
- `harness-output/qa-current/reconstructed-negatives.log`
- `harness-output/qa-current/result.json`
- `harness-output/qa-current/safety-weighted.log`
- `harness-output/qa-current/source-after.log`
- `harness-output/qa-current/ui/latency-ui-1280-ordinary-diagnostic.png`
- `harness-output/qa-current/ui/latency-ui-1280-ordinary-observation.png`
- `harness-output/qa-current/ui/latency-ui-1280-ordinary-selection.png`
- `harness-output/qa-current/ui/latency-ui-1280-restart.png`
- `harness-output/qa-current/ui/latency-ui-1280-sse-diagnostic.png`
- `harness-output/qa-current/ui/latency-ui-1280-sse-observation.png`
- `harness-output/qa-current/ui/latency-ui-1280-sse-selection.png`
- `harness-output/qa-current/ui/latency-ui-1280-strategy.png`
- `harness-output/qa-current/ui/latency-ui-1280.json`
- `harness-output/qa-current/ui/latency-ui-390-ordinary-diagnostic.png`
- `harness-output/qa-current/ui/latency-ui-390-ordinary-observation.png`
- `harness-output/qa-current/ui/latency-ui-390-ordinary-selection.png`
- `harness-output/qa-current/ui/latency-ui-390-restart.png`
- `harness-output/qa-current/ui/latency-ui-390-sse-diagnostic.png`
- `harness-output/qa-current/ui/latency-ui-390-sse-observation.png`
- `harness-output/qa-current/ui/latency-ui-390-sse-selection.png`
- `harness-output/qa-current/ui/latency-ui-390-strategy.png`
- `harness-output/qa-current/ui/latency-ui-390.json`
- `harness-output/qa-history-review/artifact-manifest.json`
- `harness-output/qa-history-review/commands.md`
- `harness-output/qa-history-review/finalize-report.mjs`
- `harness-output/qa-history-review/prior-qa4-supplement-no-go.md`
- `harness-output/qa-history-review/report-relocations.json`
- `harness-output/qa-history-review/result.json`
- `harness-output/qa-history-review/verify-history.log`
- `harness-output/qa-history-review/verify-history.mjs`
- `harness-output/qa-original-recheck/artifact-manifest.json`
- `harness-output/qa-original-recheck/final-source.log`
- `harness-output/qa-original-recheck/intake.json`
- `harness-output/qa-original-recheck/original-late.log`
- `harness-output/qa-original-recheck/original-watermark.log`
- `harness-output/qa-original-recheck/prior-qa-evidence-insufficient.md`
- `harness-output/qa-original-recheck/prior-report-relocations.json`
- `harness-output/qa-original-recheck/result.json`
- `harness-output/qa.md`
- `workflow-output/e11d82e7f5d28afd80f64477377f1c63/command-3.log`
