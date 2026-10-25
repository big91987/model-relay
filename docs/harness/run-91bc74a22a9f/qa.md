# 独立 QA 报告

- 日期：2026-10-25 UTC；Run：`91bc74a22a9f2c71662a6b9ee832b275`；节点：qa，执行6。
- 结论：**本轮研发交付及固定验证门禁通过，交人工确认**。这不是整条 Harness 流程完成、产品发布批准、真实供应商验证或旧轮失败改判。
- QA 不修改产品、测试源码、数据库或门禁配置；未重复执行全套 `make verify`，未推送、建分支、创建PR、合并或部署。

## 验收依据与同版性

本轮依据是用户当前 Harness 验证任务：先检查用例与环境，再把项目内部整套预算由12分钟明确调整为25分钟，外层1800秒保持不变，并保留全部历史、迁移、前后端、race、真实浏览器和严格失败恢复。旧业务需求仅作持续回归背景，不冒充本轮新增需求；未扩大网关业务范围。

冻结 HEAD 为 `12df9eb7a8522a679e3ccf98ca4e22b019774c5d`。固定tests5的前后源码、candidate及QA实际 `go run ./tests/evidence` 一致：160文件，`working_tree_sha256=3864d081e13800056ea0f849c598373c1eac8c03443990453752cd70b4be7c4d`。逐文件重算development4 manifest中的11个源码和22个证据，以及原development manifest字节摘要，全部匹配。交付patch是文本证据，浏览器包装保留为 `.go.txt`，没有把源码副本递归加入编译。

## 改动与用例合理性

| 改动/层级 | 独立核对结论 |
| --- | --- |
| `tests/verify`预算及回执 | 默认25分钟只改变执行条件，增加实际预算、起止UTC与取消标志；没有删门禁、放宽产品断言或修改外层期限。父取消、超时非零退出与进程组/候选清理仍被测试约束。 |
| 历史Go检查stdout/stderr分离 | JSON仅从stdout解析，stderr仍输出并参与错误分类；0/1退出及下载诊断回归保留。历史严格失败集合、额外/缺失失败、race、构建和监听失败拒绝逻辑仍在。预期历史红不等于本版主门禁失败。 |
| 键盘驱动 | 已聚焦时不先Tab离开；未聚焦仍键盘遍历、上限与可见焦点门槛不变。新增结构性回归有实现耦合，不能单独作为浏览器证据。 |
| 路由HTTP冷却夹具 | 15ms改2秒，等待自然到期后仍断言恢复；40请求3:1计数、两次attempt、冷却期间不访问primary及恢复一次均保留。只修改测试文件，产品默认冷却未变。独立真实HTTP/race复验通过。 |
| SSE浏览器采集返工 | 改用同一次页面原生fetch响应clone读取，不重放，不额外dispatch；finally恢复fetch并清状态。HTTP状态、stream请求标志、唯一dispatch、UI完成、data/DONE、同ID详情/日志、用量与费用等断言未删。5项采集正负测试涵盖缺失/外域/重复、坏body、fetch及action失败。 |
| 测试分层 | Go/race与HTTP覆盖持久化、租户、安全、预算、路由、SSE与恢复；Node/VM适合前端竞态与驱动负例，部分结构/白盒断言并非独立于实现。真实Go服务/HTTP/Chromium旅程不可被静态页面替代，本次固定回执实际执行了这些层。 |

所有新测试使用临时目录/受控上游和测试端口，没有更改用户配置上游或业务记录。受控HTTP不是供应商集成证据。

## 固定 tests5 完整回执

通过正式 `read_command_output(seq=5, save_to_workspace=true)` 导出原始输出，`truncated=false`；QA读取全部字节至EOF并校验执行完成及CLEAN_END，不以界面省略摘要代替完整输出。

- 原始命令日志：1,103,560 bytes，SHA256 `e01f395f0276bba35df538064d762ba7aea68c4edaf806a32b1fc484dfdf4f35`。
- runner日志：1,074,290 bytes，SHA256 `da15351c27805125c035937e9ab438a17da5f764fb0e9a822245fefe92c31b40`，与state逐字节摘要相符。
- 正式固定入口 `make verify`：外层exit0、gate_exit0、state.exit_code0；预算1,500,000ms，cancelled=false。
- gate：00:47:50.951至01:04:11.032 UTC，980.081秒；含准备清理总计995.135秒，外层回执995.219秒，均低于1800秒。
- candidate的node_modules/npm cache开始不存在，无手工预装，NODE_PATH unset；依赖来自candidate，Playwright/Playwright-core均1.59.1，Chromium147.0.7727.15。源码构建使用独立candidate Go build cache；**不声称共享Go模块缓存或浏览器二进制也全冷**。
- 清理回执确认所属进程组不存在、candidate已删、源码与candidate摘要不变。QA另查当前candidate目录为0、源码node_modules/npm cache仍不存在。

| 必需阶段 | 本轮证据 |
| --- | --- |
| 历史/health、格式、依赖校验、vet | 完整日志保留真实执行，主门禁继续；历史预期失败由分类器严格判断。 |
| 前端 | 285 tests，285 pass，0 fail。 |
| Go race与build | 固定 `go test -race -count=1 -v ./... -parallel=4` 各包通过，真实HTTP集成包括在内；Go构建完成。 |
| migration | 15.802秒，通过。 |
| legacy-ui | 178.558秒，通过。 |
| routing | 240.367秒，通过。 |
| console严格失败恢复 | 36.175秒，通过；1280/1440/390失响应及双发防护记录存在。 |
| full-console | 243.808秒，通过；三视口连续旅程、视觉与独立self-hosted各12步记录齐全。 |

进入浏览器前约265.143秒；legacy和routing仍是主要耗时。原2CPU/8GiB及本轮相同cgroup观测不证明环境是唯一根因。键盘无效遍历、窄时序窗口、下载stderr和SSE CDP读取失效均有具体设施解释；25分钟让完整流程可执行，**不代表慢因已修复或性能指标放宽**。

## QA 独立执行

首条复验命令中Go不在默认PATH，两条Go命令实际未执行，随后Node126项成功使组合命令最后退出为0。QA发现并原样保留该日志，**不将它计为Go通过**。随后显式使用已安装的 `/usr/local/go/bin`，以 `set -euo pipefail` 重新执行针对性命令；没有重跑全套或改变candidate。

- `go test -race -count=1 -v ./tests/verify ./tests/review-history`：通过；包含成功/失败夹具、gate timeout、cancel、SIGTERM、子进程终止、候选删除、日志摘要、stderr与严格历史错误分类。
- `go test -race -count=1 -v ./internal/relay -run '^TestRoutingRealHTTPSelectionFailoverAndRecovery$'`：通过，真实本机HTTP，测试耗时2.87秒；未申请额外权限。
- `node --test tests/browser/stream-capture.test.mjs tests/browser/ui.test.mjs`：126项通过，0失败。
- `go run ./tests/evidence`：与固定tests5同版；`git diff --check`：通过。
- QA未再次运行整套浏览器或逐页人工交互。真实浏览器结论来源于本轮固定tests5完整回执与原始JSON/PNG，QA对移动端同ID SSE详情截图作可视抽查，不把静态图当动态执行。

## 原始失败与归档

本轮tests3仍为真实exit2：817.388秒、25分钟未超时，1280自托管SSE采集失败。development4保留的command/state/runner字节摘要已重核，未改判。旧#59两次12分钟失败和旧Run环境HTTP404/平台ready状态缺口保持原事实；本节点未取得旧PR实现，不声称逐字比较或重新下载了旧证据。

tests5浏览器342个文件（含134个截图完整性侧车）逐件计算字节/SHA256，三视口self-hosted均12步通过。为满足8MiB单文件上限，原始浏览器目录仅打包并分成3个有序片段，不改原文件内容、不参与编译。QA本地重组、临时解包、逐件重算342个成员完全一致，临时目录清理。重组：`cat harness-output/tests5-browser.tar.gz.part-* > tests5-browser.tar.gz`；原路径映射与每件摘要见audit.json。压缩包整体SHA256 `68218942c1997db4f118daf65423c86d2e28f9b47d0e926d1027f62927c9c81c`。

## 未验证范围、限制与后续

- 没有真实供应商凭据/调用，没有生产性能或公网部署验收；不授权发布或合并。
- 正式人工确认、同版report、真实draft PR、最终逐件摘要、正式释放环境及从持久归档重新下载校验仍待后续节点完成，不能以本地打包或本次handoff申请冒称已完成归档下载。
- 未重验旧Run已失效环境；未修改冻结图、权限、定时任务或平台生命周期状态。
- 无阻断本次研发交付的未解决问题；既有长耗时、部分断言实现耦合、Go默认PATH差异及旧平台生命周期缺口如实保留。

## files

- `harness-output/qa.md`
- `docs/workflow/runs/91bc74a22a9f2c71662a6b9ee832b275/qa/report.md`
- `harness-output/qa-regressions.log`
- `harness-output/qa-regressions-retry.log`
- `harness-output/qa-environment.txt`
- `harness-output/audit.mjs`
- `harness-output/audit.json`
- `harness-output/audit-summary.log`
- `harness-output/archive-verification.json`
- `harness-output/manifest.json`
- `harness-output/tests5-state.json`
- `harness-output/tests5-verify.log`
- `harness-output/tests5-browser.tar.gz.part-00`
- `harness-output/tests5-browser.tar.gz.part-01`
- `harness-output/tests5-browser.tar.gz.part-02`
- `workflow-output/91bc74a22a9f2c71662a6b9ee832b275/command-5.log`
- `docs/workflow/runs/91bc74a22a9f2c71662a6b9ee832b275/development/review.md`
- `docs/workflow/runs/91bc74a22a9f2c71662a6b9ee832b275/development4/rework-review.md`
- `docs/workflow/runs/91bc74a22a9f2c71662a6b9ee832b275/development4/implementation.patch`
- `docs/workflow/runs/91bc74a22a9f2c71662a6b9ee832b275/development4/manifest.json`
- `docs/workflow/runs/91bc74a22a9f2c71662a6b9ee832b275/development4/tests3-command.log`
- `docs/workflow/runs/91bc74a22a9f2c71662a6b9ee832b275/development4/tests3-state.json`
- `docs/workflow/runs/91bc74a22a9f2c71662a6b9ee832b275/development4/tests3-verify.log`
