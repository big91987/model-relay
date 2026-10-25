## 任务与问题

本PR交付2026-10-25当前Harness验证任务的测试设施修正，不新增Model Relay网关业务功能，不调用真实供应商。本轮真实关联任务：Refs #60（https://github.com/big91987/model-relay/issues/60），依据启动输入在本次人工回退中明确补足的链接。历史关联：Refs #59；旧Run `09ce99d8fe8bf76d8e4f5024ab74e293` 两次固定测试真实exit2、12分钟期限失败保持原结论。原report8因入站缺少当前Issue链接而只列历史#59，现修正为本轮#60；旧报告保留，不冒用旧业务需求Issue #9、#25或#51，也不以本PR关闭历史平台问题。

旧观测为浏览器前约267秒、migration 15.951秒、legacy 185.691秒、routing 238.951秒，console临近总期限才开始，full-console未运行。研发检查区分了测试驱动缺陷、时序夹具及长耗时，不能仅凭2CPU/8GiB判断环境是唯一根因。用户明确授权项目内部整套期限从12分钟调整为25分钟，外层30分钟不变；这是执行条件调整，不是产品性能要求放宽或慢因修复。

## 当前基线到候选的实际改动

- `tests/verify`保持标准入口 `make verify`，内部预算改为25分钟（1,500,000ms），回执增加生效预算、门禁起止UTC及取消标志；补充超时、父取消、进程组与候选清理回归。README记录执行条件；未改外层执行器、冻结工作流、权限、数据库或定时任务。
- `tests/review-history`分离Go检查stdout/stderr，仅从stdout解析JSON，保留stderr诊断与严格失败分类，并增加下载诊断/退出分类回归。历史health分类器原已分离，不重复修改；历史预期红不算当前主门禁失败。
- 浏览器键盘驱动在目标已聚焦时不先Tab离开；未聚焦时仍走有界键盘遍历及可见焦点门槛，并补结构性回归。
- 真实HTTP路由测试把15ms冷却夹具调整为2秒，仍等待自然到期、检查40请求3:1分配、failover两次attempt、冷却期间不访问primary及恢复一次；产品默认冷却未变。
- 本轮首次固定tests3出现自托管SSE采集设施失败，返工新增 `stream-capture.mjs`：从同一次页面原生fetch响应clone采集SSE，不重放请求、不额外dispatch，finally恢复fetch与状态。新增5项正负回归，保留HTTP状态、stream标志、唯一dispatch、UI完成、data/DONE、同ID详情/日志和用量费用断言。
- 没有删除历史/迁移、前后端、Go race/build、真实浏览器或严格失败恢复门禁，没有预热隔离candidate或遇红继续抬高期限。

这些问题不是本轮首次发现。approval7中Builder补充正式GitHub API核对：旧draft PR #55的head `32213cf63995f6e3b06488bd0ec48c4ed7bc1b24` 已有25分钟及同一fetch响应clone采集的修正方向，尚未进入本轮冻结main；本轮在当前基线重新实现、补回归并验证。此来源说明不证明逐字等价，也不是Pipeline起跑前取得旧实现的证据。研发/QA当时未取得旧PR的记录保持原事实；旧PR #49/#55的绿色不作本轮验收。Builder未预先准备旧材料造成重复定位，是本轮输入准备效率问题，不归为产品故障或容量问题；不复制旧PR整包、不扩范围、不自动合并。

## 已验证版本与正式全量结果

- Run：`91bc74a22a9f2c71662a6b9ee832b275`；工作流 `83f08f81f7040fcd0c12e6ff4e2acd2d` revision10。冻结main/HEAD：`12df9eb7a8522a679e3ccf98ca4e22b019774c5d`。验证对象是该HEAD加本PR的未提交工作树，不把冻结HEAD单独称作已验证候选提交。
- 160文件源码摘要：`3864d081e13800056ea0f849c598373c1eac8c03443990453752cd70b4be7c4d`，研发、正式tests5前后/candidate、独立QA及report重算一致。摘要范围包含产品、测试、构建锁定与项目契约，排除交付回执避免自引用；发布节点应保持这11个变更文件字节不变并记录实际新提交。
- 正式固定tests5唯一修复后全量 `make verify`：外层exit0、gate/state exit0，cancelled=false，生效预算1,500,000ms。gate为2026-10-25 00:47:50.951至01:04:11.032 UTC，980.081秒；含准备清理995.135秒，外层执行回执995.219秒。report未再跑全套。
- 完整原始输出1,103,560 bytes，SHA256 `e01f395f0276bba35df538064d762ba7aea68c4edaf806a32b1fc484dfdf4f35`；QA由正式日志接口导出并读至EOF，truncated=false。runner输出1,074,290 bytes，SHA256 `da15351c27805125c035937e9ab438a17da5f764fb0e9a822245fefe92c31b40`，与state相符；不以UI省略摘要或命令节点completed代替完整通过。
- 历史/health、格式、依赖校验、vet、285项前端测试、Go race及build全部保留通过。实际Go/HTTP/Chromium五阶段：migration 15.802秒、legacy-ui 178.558秒、routing 240.367秒、console严格失败恢复36.175秒、full-console 243.808秒；1280/1440/390三视口连续旅程、视觉和各12步self-hosted完成。
- candidate的node_modules/npm cache开始不存在，无手工预装，NODE_PATH unset；依赖在candidate解析，Playwright及playwright-core 1.59.1、Chromium 147.0.7727.15。独立Go构建缓存；不声称共享Go模块缓存或浏览器二进制全冷。所属进程组消失、candidate删除、源码不变可核对。

## 独立QA与正式确认

QA6结论是本轮研发交付及固定门禁通过。QA未修改产品/测试/数据库/配置；独立执行runner与历史分类Go race、真实HTTP路由Go race、126项前端/采集回归和源码核验。首条复验因默认PATH没有Go而未执行Go，组合命令被后续Node成功掩盖的原始日志保留，不算Go通过；显式PATH并使用fail-fast后定向Go复验成功。QA未重复全套浏览器或逐页人工交互，浏览器结论依据tests5完整回执、原始JSON/PNG，移动端同ID SSE详情仅作可视抽查。部分键盘/前端结构断言有实现耦合，不能独立代替真实后端旅程。

approval7已保存正式人工确认：Builder依据用户授权代Owner确认进入同版报告和固定publisher真实draft PR，不是用户本人点击，也不替代独立QA；不批准合并或部署。report仅生成本报告，不创建分支、不提交、不推送、不创建PR。

## 首次发布拒绝与人工恢复

2026-10-25固定publish9真实exit1：report8标题使用发布器不支持的【测试设施】类别，在字段校验阶段抛出 `ValueError: report must describe the actual change and evidence`，未进入Git写入、提交、推送或创建PR。正式原始日志843 bytes、SHA256 `13641f4829a4e71c3682c900d323324bfe3b76915e0f5d5b0d098b6ec45eee4f`，已导出并读取至EOF、truncated=false。根据本次用户回退材料，Owner经既有授权停止并正式回退report；Builder通过正式GitHub入口核对Run分支 `codex/run-91bc74a22a9f` 无匹配PR。这不是自动成功交接，也不是本节点重新进行GitHub核对的结论。

本次report10仅将标题改为允许的【Harness验证】、补足真实Issue #60及本段失败/人工恢复记录；11个QA确认文件、160文件源码摘要和QA原文不变，不重跑 `make verify`、QA或approval。原report8按原字节保存为 `harness-output/report8-original.json`，SHA256 `e34dad060a0ec521ffec96749f5c0c8449dac136b25edaac23af6c7829f16412`；原发布失败日志保存为 `workflow-output/91bc74a22a9f2c71662a6b9ee832b275/command-9.log`。`read_command_output(seq=8)`返回record not found，report8是Agent节点，不能冒称取得其命令日志；核对依据是旧报告快照及上游已保存的report8交接说明。维护源ea38d07的入口/模板修正与133项模板回归、主图升revision11来自用户回退材料，不属于本PR或本节点验证；当前Run仍冻结revision10，未热改图。approval7继续只授权同版报告和draft PR，不授权合并、部署或扩大权限。当前仅完成材料修复和本地字段/同版校验，重新发布及整链终态尚待固定节点实际执行。

## 失败、归档与限制

- 本轮tests3真实exit2、817.388秒、25分钟未超时，1280自托管SSE采集失败；command/state/runner与返工证据原样保留。旧#59两次12分钟失败不能因tests5通过改判。旧Run文件API HTTP404而平台快照ready的生命周期缺口未修复，未重建旧环境或重放未知副作用。
- 本轮浏览器342成员（含134完整性侧车）保留原路径及字节摘要，打包为3个小于8MiB片段；QA本地重组解包逐件相符。源码快照以文本/补丁和 `.go.txt` 保存，原名映射与摘要见manifest，不让源码副本参与编译，不排除正常测试。
- QA6当时只完成本地留存核验；approval7另记录Builder核对87件平台归档、30,818,063 bytes逐size/SHA匹配，并对同一tests5 attempt重组的342成员、37,700,340 bytes独立核对匹配。这是留存完整性确认，不是Builder代做产品/视觉QA，也不等于最终发布后的环境释放与重新下载核验已完成。
- 没有真实供应商凭据/调用、生产性能、公网或生产部署验收。受控HTTP上游只是隔离测试。legacy/routing仍耗时较长，2CPU/8GiB不是唯一原因结论。
- 后续仅由固定publisher创建可审阅的同版draft PR，记录实际提交与PR链接，不合并、不部署；随后仍需最终产物逐文件核对、正式入口释放环境及从持久归档重新下载核验。上述后续尚未发生，不把本地文件或handoff受理称为整链完成。

## 提交与证据范围

`files`只包含development4 manifest列出的11个、经QA确认的仓库变更文件；report已重算各文件字节/SHA256并用 `git check-ignore --stdin` 核对无忽略项，exit1且无输出。未加入构建目录、node_modules、凭据、运行数据库或原生会话历史。`harness-output/report.json`、QA原文及原始日志/打包证据只作workflow artifacts，不列入提交files；研发检查与失败证据仍作为交接材料保留。固定发布节点按契约把QA和PR正文原文复制到 `docs/harness/run-*`，不批量提交本地证据树。

审阅入口：`harness-output/qa.md`；`docs/workflow/runs/91bc74a22a9f2c71662a6b9ee832b275/development/review.md`；`docs/workflow/runs/91bc74a22a9f2c71662a6b9ee832b275/development4/rework-review.md`；`harness-output/manifest.json`（QA逐件摘要）；`harness-output/audit.json`（浏览器成员映射/摘要）；`harness-output/tests5-state.json`；`workflow-output/91bc74a22a9f2c71662a6b9ee832b275/command-5.log`。