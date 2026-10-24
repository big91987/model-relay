# 独立 QA12：同版补证复核通过，交正式确认

## 具体工作与结论

- Run：197ab169bc1cad067466300038ab6fac；冻结版本7；Issue：big91987/model-relay #54。独立QA执行12在同一 /workspace 复核研发11补证及QA10未决项，不修改产品、测试、Makefile或发布说明。
- **结论：PASS_TO_APPROVAL。QA-10-E1关闭。** tests9完整固定门禁与QA10独立技术复测成立；新增正式平台来源材料补齐历史显式交接、配置/执行和会话身份链。本结论只允许进入正式确认，不等于批准发布、报告已完成或草稿PR已创建。
- 本轮执行独立证据审计 node harness-output/qa-12-audit.mjs，退出0；另执行 go run ./tests/evidence，退出0。没有重跑make verify、浏览器或已通过的定点测试，不热重跑求绿。审计逐字节核对而非照搬研发“通过”摘要。
- 复核的实际研发改动仅测试设施：历史race诊断与JSON输出分离并保留失败；冷却恢复测试按实际截止时间断言；runner有界预算从12分钟调整至25分钟（不是本轮再延预算，Makefile入口/race未降级）；SSE观察器读取原fetch响应clone、同Request ID、恰好一次dispatch、finally清理，不新增上游请求/重试/假响应。业务断言、账本计量和SSE data:/[DONE]仍保留。没有生产实现改动，未扩展#51。

## 同版源码和固定测试原件

- 基线HEAD：868530b3e5c15d65bee0ed4ff6af69e96d428b81；验收输入160文件SHA-256：44107d4db03a944c27a7d9b36b03e1802a770b3219c9bae61b19f0d81d4795fd。与tests9前后、QA10前后及本轮最终采样一致。工作树有本任务研发差异，不冒称等同main或已提交。八份执行8最终源码.txt快照与当前源码字节相等；上轮索引167项及原110项证据完整核对。
- 本轮通过 read_command_output(seq=9, save_to_workspace=true) 重新独立取得固定tests原生导出，工具回执bytes=1093069、truncated=false、SHA-256=868edba207d2a972ce2c0b67c3ce5ec2485ed49f61334908d6a7cae30bf0e4ca。完整读取并核对与旧归档相同，不能以提示中的省略输出作为完整日志证据。
- tests9实际执行原始make verify，从2026-10-22T02:42:49.44085415Z至02:59:53.800960021Z，1024.360秒；外层、runner、gate均exit0。七个Go包race通过；前端285项通过、0失败；build完成；migration/legacy-ui/routing/console/full-console五阶段结束均ok=true。1280/1440/390三个宽度的continuous/self-hosted/visual九份结果均passed。
- runner原日志1062735字节，SHA-256=c27a7fce90c4e6769363f5f5d85ea412de3aa3acc7c8d3b2a8fd7ba9cdfc9c5e；与state中的大小/摘要相符；source/candidate未变，所属进程组已清理。未把测试中的预期取消/错误注入fixture日志误判为整套失败或忽略真正失败。
- QA10独立复测原日志及六份命令回执均保留并核对exit0：观察器7项反向/清理用例；review-history/verify的race；实际HTTP路由故障转移、冷却恢复、已提交流不重放、连续HTTP及自托管受控边界。该技术结果属于QA10执行，不写成本轮新跑的测试。此前失败、诊断及返工3/5/7原日志继续留存。

## QA-10-E1关闭依据与证据边界

1. 正式User Input11016由Builder转交正式Run/Conversation/Event API只读结果，归档为platform-api-supplement-user-input.json。development、qa、report冻结配置均codex/gpt-6.1-sol。development8 conversation=273dda67cdbc6cc708ecd76619af26b2/thread=01a1616f-7d51-7a50-ba1e-605ab2d5b7bf；QA10 conversation=7b70d2cc1fa48416c05136d2fb144c89/thread=01a1618d-1211-7543-86fe-cf064e9d28c7；会话与原生thread不同，工作区均/workspace。这支持角色会话身份分离，不是穷尽式内部prompt注入边界审计。
2. 正式User Input11328补入development8原生MCP事件31753，call_ZzFP9Esbj348tnZaiHMlN12l，server=registered_workflow_node/tool=handoff/status=completed/error=null；target=tests，非空完整message，结构化context，120个实际artifacts。受理structuredContent.accepted=true且Run/seq8/target绑定一致，content内JSON与其相同，时间2026-10-22T02:40:30.455838329Z与11016摘要一致。不是用聊天句子或本地模板冒充调用。
3. 对120项历史artifacts独立核对存在性/版本：110项证据及八个最终源码不变，两份历史docs映射至rework-11-before-*.txt保留原字节，总计12906095字节，等于原context；不以当前文档冒充历史版本。完整映射/摘要在qa-12-analysis.json。
4. thread.started未持久化model字段。11016/11328的证据链是冻结配置+实际Codex thread/tool执行+平台提供的安装源9b03a305aa838f43b4db88e73ccdb9e15e8f608e中thread/start返回Model必须匹配Snapshot.Model才发送turn/start的运行时校验说明；安装二进制SHA-256=602988ff44bff020077edc64cfabdca673fb8a763b6647c0ef19fd242dd32256。本QA接受该正式平台外部声明和可关联事件满足操作层角色/模型配置验收，不再要求不存在的事件字段；未自行读取安装源码/二进制、未直连平台API、未获得原HTTP响应封套，不声称供应商侧模型鉴定。
5. 因完整交接原字段现已补齐、模型缺字段及来源已明确，QA10的缺证据影响已解除；不是撤销技术门槛或要求研发凭空补字段。QA10原No-Go和read_command_output(seq8)找不到记录的历史均保留，不能改写成当时通过或handoff失败。report尚未执行，其实际执行及会话证据是后续报告阶段责任，不要求未来证据提前存在。

## 人工恢复、阶段与后续责任

- 正式User Input11625明确development唯一出口tests会重新执行make verify，不能当作转QA；development11补证后等待Owner正式停止/回退。本轮由用户正式人工回退到qa恢复，是人工流程恢复，不是development11成功handoff；不修改冻结图、不补造工具受理。此前两次待投递User Input导致的handoff拒绝和wait状态冲突原件继续保留。
- docs/harness开发说明/manifest仍是development11写入时“等待Owner回退/待QA复核”的历史状态，本QA不改被审材料；当前较新的QA结论以本报告为准，后续report应在正式确认后按时间线同步发布说明，不篡改历史。
- development：实现与补证完成；固定tests9：通过；独立QA12：通过并准备交approval；approval：待正式确认；report：未运行；固定发布：未创建本任务草稿PR。没有commit/push/新分支/合并/部署。当前handoff是否受理以本轮工具实际回执为准，本文件不预写受理成功。
- 后续正式确认应审阅本结论及来源限制，再进入report及固定发布；report实际executor/model/会话/工具受理须在其运行后记录。发布仅限原publication_files十个具体文件，不包含harness-output或ignored临时文件。不得将本QA通过称为整链已交付。

## 未验 / 限制 / 遗留

- 没有调用真实模型供应商，没有读取/复用用户凭据；所有HTTP/SSE证据使用隔离受控上游和实际Go服务，不是供应商接入验证。未修改数据库记录或用户运行配置。
- 候选a033eb20ea848d3aead77231da367703b1b43fb8不可读，未比较/复用，不声称完成该候选验证。旧#48/#49仅为历史，不能替代本轮证据。
- Issue#54和入站Actions39703355037/39703356253关联来自正式用户补充，未独立获取GitHub Actions回执。
- 本轮没有新开完整QA浏览器或逐图人工视觉验收；固定门禁原始浏览器回执和QA10独立HTTP复测已核对。没有广泛sandbox扩权；平台外部模型证据的来源限制如前。
- report/发布的未来真实结果尚未出现，不作为当前通过范围。没有遗留产品缺陷被本轮复核发现；本结论也不证明所有潜在缺陷不存在。

## 归档、时间及可复核性

- QA10原报告原字节保存在qa-10-report-preserved.md，10728字节，SHA-256=a37ae518ef2761961fbeaab966a8ff99a1e15c223dc2cdaf4690e21cddfc50f4；原qa-10-files.json中的qa.md映射至此文件。不覆盖原失败证据；旧索引是历史版本。
- 本次原生导出workflow-output/197ab169bc1cad067466300038ab6fac/command-9.log等字节移至harness-output/qa-12-tests9-platform-export.log，原文件名/摘要/映射见analysis；临时空目录已清理，不将工作流日志或可编译源码副本污染构建输入。源码快照均.txt，字节不改，不排除正常测试。
- 本轮工具及PATH无apply_patch；仅以Node写独立QA报告/审计文件和保留原件，不改产品。审计执行退出0；最终source摘要再次相同。files索引列每个实际文件的大小/SHA-256（索引自身不自哈希），单文件8MiB、256件/128MiB限制已检查。
- 首次明确UTC采样2026-10-22T03:35:45Z；报告写入2026-10-22T03:41:13.031Z；可证实驻留下界328.031秒，精确入场未知。此前固定测试返工3次，QA补证返工1次；development11各恢复/等待采样保留在其记录，人工恢复未当研发新测试。没有捏造Owner等待时长或未来驻留时间。

## files（实际文件；harness-output仅归档，不发布）

- `docs/harness/run-197ab169bc1cad067466300038ab6fac/development.md`
- `docs/harness/run-197ab169bc1cad067466300038ab6fac/evidence-manifest.json`
- `harness-output/Makefile.txt`
- `harness-output/archive-evidence.mjs`
- `harness-output/baseline-command.json`
- `harness-output/baseline-make-verify.log`
- `harness-output/baseline-runner-state.json`
- `harness-output/baseline-runner-verify.log`
- `harness-output/development-checks-command.json`
- `harness-output/development-checks.log`
- `harness-output/development8-event-before-development.md.txt`
- `harness-output/development8-event-before-evidence-manifest.json.txt`
- `harness-output/development8-full-supplement-handoff-rejected.json`
- `harness-output/development8-handoff-audit.json`
- `harness-output/development8-handoff-audit.mjs`
- `harness-output/development8-handoff-files.json`
- `harness-output/development8-handoff-platform-event-user-input.json`
- `harness-output/development8-handoff-review.json`
- `harness-output/handoff-contract-command.json`
- `harness-output/handoff-contract-tests.log`
- `harness-output/history-diagnostic-command.json`
- `harness-output/history-diagnostic.stderr.log`
- `harness-output/history-diagnostic.stdout.log`
- `harness-output/implementation.patch`
- `harness-output/linux-candidate-response.txt`
- `harness-output/owner-qa-rollback-audit.json`
- `harness-output/owner-qa-rollback-before-development.md.txt`
- `harness-output/owner-qa-rollback-before-evidence-manifest.json.txt`
- `harness-output/owner-qa-rollback-files.json`
- `harness-output/owner-qa-rollback-request.json`
- `harness-output/owner-qa-rollback-source-after.json`
- `harness-output/owner-qa-rollback-user-input.json`
- `harness-output/platform-api-supplement-user-input.json`
- `harness-output/platform-supplement-assessment.json`
- `harness-output/platform-supplement-before-development.md.txt`
- `harness-output/platform-supplement-before-evidence-manifest.json.txt`
- `harness-output/platform-supplement-files.json`
- `harness-output/platform-supplement-handoff-rejected.json`
- `harness-output/qa-10-analysis.json`
- `harness-output/qa-10-audit.mjs`
- `harness-output/qa-10-checks.json`
- `harness-output/qa-10-diff.log`
- `harness-output/qa-10-files.json`
- `harness-output/qa-10-http.log`
- `harness-output/qa-10-observer.log`
- `harness-output/qa-10-report-preserved.md`
- `harness-output/qa-10-runner-history.log`
- `harness-output/qa-10-source-after.log`
- `harness-output/qa-10-source-before.log`
- `harness-output/qa-12-analysis.json`
- `harness-output/qa-12-audit.mjs`
- `harness-output/qa-12-files.json`
- `harness-output/qa-12-resumption.json`
- `harness-output/qa-12-source-final.log`
- `harness-output/qa-12-tests9-platform-export.log`
- `harness-output/qa.md`
- `harness-output/review-history-before-command.json`
- `harness-output/review-history-before-main.go.txt`
- `harness-output/review-history-before-regression.log`
- `harness-output/review-history-final-command.json`
- `harness-output/review-history-final-regression.log`
- `harness-output/review-history-main.go.txt`
- `harness-output/review-history-main_test.go.txt`
- `harness-output/review-history-regression.log`
- `harness-output/rework-11-before-development.md.txt`
- `harness-output/rework-11-before-evidence-manifest.json.txt`
- `harness-output/rework-11-evidence.json`
- `harness-output/rework-11-files.json`
- `harness-output/rework-4-Makefile.txt`
- `harness-output/rework-4-archive.mjs`
- `harness-output/rework-4-before-delayed.go.txt`
- `harness-output/rework-4-before-delayed.log`
- `harness-output/rework-4-checks.mjs`
- `harness-output/rework-4-commands.json`
- `harness-output/rework-4-implementation.patch`
- `harness-output/rework-4-path-after-delayed.go.txt`
- `harness-output/rework-4-path-after-delayed.log`
- `harness-output/rework-4-path-before-delayed.go.txt`
- `harness-output/rework-4-path-before-delayed.log`
- `harness-output/rework-4-path-checks.mjs`
- `harness-output/rework-4-path-commands.json`
- `harness-output/rework-4-path-format.log`
- `harness-output/rework-4-path-targeted.log`
- `harness-output/rework-4-path-vet.log`
- `harness-output/rework-4-review-history-main.go.txt`
- `harness-output/rework-4-review-history-main_test.go.txt`
- `harness-output/rework-4-routing_http_test-before.go.txt`
- `harness-output/rework-4-routing_http_test.go.txt`
- `harness-output/rework-4-source-current.json`
- `harness-output/rework-6-archive.mjs`
- `harness-output/rework-6-checks.json`
- `harness-output/rework-6-checks.mjs`
- `harness-output/rework-6-diff.log`
- `harness-output/rework-6-final-audit.json`
- `harness-output/rework-6-format.log`
- `harness-output/rework-6-implementation.patch`
- `harness-output/rework-6-runner-race.log`
- `harness-output/rework-6-runner-vet.log`
- `harness-output/rework-6-source.log`
- `harness-output/rework-6-verify-main-before.go.txt`
- `harness-output/rework-6-verify-main.go.txt`
- `harness-output/rework-6-verify-main_test-before.go.txt`
- `harness-output/rework-6-verify-main_test.go.txt`
- `harness-output/rework-8-after-browser.log`
- `harness-output/rework-8-after-journeys_test.go.txt`
- `harness-output/rework-8-after-overlay.json`
- `harness-output/rework-8-archive.mjs`
- `harness-output/rework-8-browser-race.log`
- `harness-output/rework-8-browser-vet.log`
- `harness-output/rework-8-checks.json`
- `harness-output/rework-8-checks.mjs`
- `harness-output/rework-8-commands.json`
- `harness-output/rework-8-dependencies.log`
- `harness-output/rework-8-diagnostic-journeys.mjs.txt`
- `harness-output/rework-8-diagnostic-journeys_test.go.txt`
- `harness-output/rework-8-diagnostic-overlay.json`
- `harness-output/rework-8-diagnostic-unit.log`
- `harness-output/rework-8-diagnostic.log`
- `harness-output/rework-8-diff.log`
- `harness-output/rework-8-final-audit.json`
- `harness-output/rework-8-final-internal-relay-routing_http_test.go.txt`
- `harness-output/rework-8-final-source.log`
- `harness-output/rework-8-final-tests-browser-call-response.mjs.txt`
- `harness-output/rework-8-final-tests-browser-call-response.test.mjs.txt`
- `harness-output/rework-8-final-tests-browser-journeys.mjs.txt`
- `harness-output/rework-8-final-tests-review-history-main.go.txt`
- `harness-output/rework-8-final-tests-review-history-main_test.go.txt`
- `harness-output/rework-8-final-tests-verify-main.go.txt`
- `harness-output/rework-8-final-tests-verify-main_test.go.txt`
- `harness-output/rework-8-frontend.log`
- `harness-output/rework-8-implementation.patch`
- `harness-output/rework-8-journeys-before.mjs.txt`
- `harness-output/rework-8-journeys-test-before.mjs.txt`
- `harness-output/rework-8-source.log`
- `harness-output/rework-8-targeted-unit.log`
- `harness-output/rework-8-tracked-only-implementation.patch`
- `harness-output/source-current.json`
- `harness-output/tests-3-platform-export.log`
- `harness-output/tests-3-platform.log`
- `harness-output/tests-3-state.json`
- `harness-output/tests-3-verify.log`
- `harness-output/tests-5-analysis.json`
- `harness-output/tests-5-model-selection-1280.jsonl`
- `harness-output/tests-5-model-selection-390.jsonl`
- `harness-output/tests-5-platform-export.log`
- `harness-output/tests-5-routing-selection-30296.jsonl`
- `harness-output/tests-5-state.json`
- `harness-output/tests-5-verify.log`
- `harness-output/tests-7-analysis.json`
- `harness-output/tests-7-platform-export.log`
- `harness-output/tests-7-self-hosted-1440.json`
- `harness-output/tests-7-state.json`
- `harness-output/tests-7-verify.log`
- `harness-output/tests-9-continuous-1280.json`
- `harness-output/tests-9-continuous-1440.json`
- `harness-output/tests-9-continuous-390.json`
- `harness-output/tests-9-platform-export.log`
- `harness-output/tests-9-self-hosted-1280.json`
- `harness-output/tests-9-self-hosted-1440.json`
- `harness-output/tests-9-self-hosted-390.json`
- `harness-output/tests-9-state.json`
- `harness-output/tests-9-verify.log`
- `harness-output/tests-9-visual-1280.json`
- `harness-output/tests-9-visual-1440.json`
- `harness-output/tests-9-visual-390.json`
- `harness-output/user-input-issue-54.json`
- `harness-output/verify-main.go.txt`
- `internal/relay/routing_http_test.go`
- `tests/browser/call-response.mjs`
- `tests/browser/call-response.test.mjs`
- `tests/browser/journeys.mjs`
- `tests/review-history/main.go`
- `tests/review-history/main_test.go`
- `tests/verify/main.go`
- `tests/verify/main_test.go`
