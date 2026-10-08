# 本Run证据索引

## 报告22交付审查
- [一致性复核](report-22-review.json)：重新计算147件候选摘要，核对148件QA归档长度/hash、原seq18失败、当前锁与材料；不是新增make verify、独立动态浏览器或部署回执。
- [报告交接审计](report-22-handoff.json)：本轮实际产物、文档/链接检查与发布前置；QA21的handoff-artifacts.json仍是历史快照，本轮公共状态更新不改写其哈希。
- [对外PR正文](../pr.md)：使用实际仓库与已转义任务分支URL；尚未推送，发布Connector须先提交推送并确认目标文件存在。
- 公共总览Executing、T001 In Review、M01 Ready for Review：冷门禁及独立QA已完成，授权发布部署待办。下列各轮记录是历史证据，不是并行的当前任务状态。

## 执行21 当前权威入口
- [历史summary补充独立评估](../acceptance/summary-assessment-qa-21.md)：旧失败Open/根因未知；P2验证可靠性风险，产品P1未实证。没有把一次20绿写为修复关闭，保留原断言；动态独立浏览器Blocked/供应商Not Run；若再红的最小真实GET屏障/同步契约见该文，不盲重跑整套。
- [独立QA报告](../acceptance/report-qa-21.md)、[矩阵](../acceptance/matrix-qa-21.md)、[连续旅程](../acceptance/journeys-qa-21.md)、[缺陷/风险](../acceptance/defects-qa-21.md)：147件f8befd候选Go with known issues，仅进入精确发布；正式部署Not Run。
- [本轮具体证据索引](../acceptance/evidence/qa-21/index.md)、[逐件路径/长度/hash](../acceptance/evidence/qa-21/archive-index.json)：原样148件宿主20资产，完整runner310718bytes/hash7c4d3952…abb7b与state一致；Connector310815bytes/未截断/EOF分开记录。
- [QA当前源码](../acceptance/evidence/qa-21/source-qa-21.json)、[独立检查回执](../acceptance/evidence/qa-21/independent-review.json)：定向Go race/197Node/静态exit0；QA原生固定入口监听Blocked保留原log/state。
- [发布交接](../acceptance/release-handoff-qa-21.md)：新PR/精确merge/Owner正式preview及版本/页面闭合由授权节点继续，不能停草稿或冒称已上线。以下全部旧候选/失败/中断/No-Go保持历史，不用于替代新版本。

## 执行19 历史诊断候选（非修复闭合）
- [当前归因/宿主契约](remediation-development-19.md)：tests18真实冷入口628.083秒、摘要原断言失败；无增限/同步假设修复。下一步原make verify产出精确边界。
- [当前源码摘要](source-development-19.json)、[本地回归/静态日志](development-19-check.log)：仅局部证据，不能替代新宿主冷全绿或独立QA。
- [tests18实际runner完整日志](tests-18-verify.log)：307276bytes，SHA256 cb92549be85d96006a9d9131a12067982e41abccf5e75147c7353f0803e7c430；不是307458bytes Connector全文。
- [tests18原state](tests-18-state.json)：前后147件28571f候选一致、冷起点已测、gate_exit2、cleanup闭合；失败不放行。
- [tests18具体安全浏览器失败JSON](tests-18-continuous-1280.json)：1280px summary-model-context-toggle、原断言失败、旧failure_boundary unavailable。未上传raw trace/未经审查截图。
- [最终一致性检查](development-19-final.log)：摘要、文件非空、原收据bytes/hash及新日志隐私检查。以下全部保留为历史，不替当前候选放行。

## 执行17历史证据入口
当前候选身份以 [source-development-17.json](source-development-17.json) 为准；新完整冷门禁待宿主，QA No-Go保持。执行16已证明累计deadline失败；不是绿灯，也未证cooldown渲染缺陷。下面都是实际非空文件，不使用目录占位：
- [remediation-development-17.md](remediation-development-17.md)：归因、最小改动、检查结论、限制及正式tests归档契约。
- [tests-16-verify.log](tests-16-verify.log)：完整原runner293790bytes，SHA256与原state相同；非Connector294002bytes全文。
- [tests-16-state.json](tests-16-state.json)：原样冷起点/失败/源不变/清理回执，不改写为Pass。
- [tests-16-routing-selection.jsonl](tests-16-routing-selection.jsonl)：原始白名单选择诊断，非后续完成证明。
- [development-17-serial.log](development-17-serial.log)：首次局部命令默认Go cache越权失败exit1，保留并脱敏路径。
- [development-17-serial-project-cache.log](development-17-serial-project-cache.log)：同13状态码原串行race exit0/32.244秒。
- [development-17-parallel.log](development-17-parallel.log)：同13状态码原断言并行race exit0/11.765秒。
- [development-17-inprocess-race.log](development-17-inprocess-race.log)：五组独立场景联合race exit0/31.548秒；不是网络证明。
- [development-17-check.log](development-17-check.log)：194Node及runner/依赖race通过，随后browser Go包实际HTTP监听受限exit1；静态后续未执行，不把组合写为通过。
- [development-17-static.log](development-17-static.log)：仅无监听诊断回归和格式/module/vet/build/diff检查，最终退出码见日志；不是完整入口。
- [development-17-final.log](development-17-final.log)：最终源码摘要重算一致、格式/diff、15项交接产物非空及原失败日志hash/长度核对exit0；不包括该检查日志自身的计数。

正式tests仅运行原argv make verify；归档新唯一receipt的完整log/state和隐私审查后具体browser文件，独立QA核对冷环境/源码/UTC/hash/清理及全部原门禁。再次失败按真实边界返工，不盲增限或用旧绿；下述执行15及更早均历史。

## 执行15历史证据入口
当前147件候选 `24725bbb7833cfc2665d44b67de4982cc045aa375ad5ec65c68c6c7d6f2b044e` 是有界归因诊断，不是超时修复或完整冷绿灯。任务In Review/M01 Blocked/独立QA No-Go保持。以下均实际非空文件：
- [remediation-development-15.md](remediation-development-15.md)：原deadline/清理与未证根因、最小诊断、失败保留、Trellis-check及tests契约。
- [source-development-15.json](source-development-15.json)：当前HEAD及147件摘要，执行13为历史被测版。
- [tests-14-verify.log](tests-14-verify.log)：原固定入口完整runner日志249576bytes，hash与state吻合，不是249788bytes Connector全文。
- [tests-14-state.json](tests-14-state.json)：原样失败receipt，冷起点/结束/source不变/cleanup闭合，但gate_exit=-1、exit_code=1。
- [tests-14-routing-selection.jsonl](tests-14-routing-selection.jsonl)：153条原始白名单事件，994093bytes，不证明后续未完成步骤。
- [development-15-tests-initial.log](development-15-tests-initial.log)、[development-15-static-initial.log](development-15-static-initial.log)：局部zsh采集错误exit1，保留。
- [development-15-static-gate-failure.log](development-15-static-gate-failure.log)：原字面调用门禁红灯，保持断言、已修包装调用。
- [development-15-tests.log](development-15-tests.log)：中间定向检查exit0，不替代最终候选或完整门禁。
- [development-15-tests-final.log](development-15-tests-final.log)：最终Go新诊断回归及原runner/依赖race exit0。
- [development-15-static.log](development-15-static.log)：最终格式/模块/vet/build/diff/194Node逐项exit0；原已有依赖环境，不是冷证明。

下一步仅原make verify新独立冷实测，按实际timing归因再选修复，不改时限/Connector/断言或用旧绿放行。完整新绿后才独立QA/授权发布。下述执行13及更早均历史。

## 执行13历史证据入口
当前144件候选 `dc73ac404db4daaf7e264552c0940abb294b2b1b48684e792eec90822b6eb665`。三项P2最小复现成立并修复，局部检查Pass；完整新宿主冷门禁仍Not Run，独立QA及发布No-Go保持。以下均为本Run实际非空文件，非目录占位。

- [remediation-development-13.md](remediation-development-13.md)：红→绿、最小边界、Trellis-check、旧中断/现场限制及新原固定入口归档契约。
- [source-development-13.json](source-development-13.json)：当前HEAD、144件摘要；执行11/9/142件为历史。
- [development-13-red.log](development-13-red.log)：父目录越界实际复制、两路诊断丢失、库存取消未生效三项新增断言的真实红灯。
- [development-13-initial.log](development-13-initial.log)：初轮回归旧输出格式断言失败，保留。
- [development-13-tests.log](development-13-tests.log)：第一次两包定向race exit0。
- [development-13-tests-final.log](development-13-tests-final.log)：真实SIGTERM新回归的测试自身竞态exit1，原堆栈保留；不是最终绿灯。
- [development-13-tests-candidate.log](development-13-tests-candidate.log)：修复测试同步后最终两包race exit0，含路径边界、stdout/stderr、实际子进程取消/SIGTERM落盘。
- [development-13-cancellation-repeat.log](development-13-cancellation-repeat.log)：三个取消/SIGTERM回归各重复三次，race exit0。
- [development-13-static.log](development-13-static.log)：格式/module/vet/build/diff逐项exit0。
- [tests-12-state.json](tests-12-state.json)：seq12同次runner实际原样state，finished_utc空、cleanup=not_started/log_bytes0，非通过；未改写。
- [tests-12-partial-verify.log](tests-12-partial-verify.log)：同次实际runner中断前文件脱敏副本60709bytes，非61016bytes Connector全文或新绿；出处/原摘要见执行13记录。原平台全文read_command_output至EOF，seq12外层exit-1保持。

正式tests须原 `make verify` 得到新唯一receipt目录并归档其完整log/state与审查后的具体browser文件。停止、缺结束、源码漂移或残留组不算通过；不能使用旧推荐脚本/手装依赖/更改Connector或旧seq5放行。

## 研发返工（执行7）
- [remediation-development-7.md](remediation-development-7.md)：同版宿主隔离执行与完整日志导出契约，仍待宿主补证。
- [host-clean-verify.sh](host-clean-verify.sh)：实际make verify外的干净副本准备/采集辅助脚本，不是通过证据。
- [development-7-check.log](development-7-check.log)：本轮定向回归、脚本语法、格式/diff及同版源码检查；不替代完整宿主绿灯。

## 当前独立QA（执行6）
同版宿主tests执行5的make verify exit0已核对，候选142件源码戳与QA一致；原生QA完整入口仍监听受限exit2。当前结论Blocked：正式宿主独立候选、开始无node_modules/NODE_PATH unset/未预装的强制干净证明缺失，宿主完整日志尚需本Run具体文件补档。不是产品实现被证实再次失败；不宣称上线。

- 当前报告：../acceptance/report-qa-6.md
- 权威矩阵：../acceptance/matrix-qa-6.md
- 连续旅程：../acceptance/journeys-qa-6.md
- 问题与补证：../acceptance/defects-qa-6.md、../acceptance/remediation-handoff-qa-6.md
- 实际QA证据索引：../acceptance/evidence/index.md（各具体文件及来源/结果分开）

## 研发原始证据（保留）

以下均为本Run实际文件。日志保留完整命令输出/退出码，机器工作区前缀改为仓库相对路径、npm内部路径改为`<npm>/`，私有debug日志地址移除；错误类型/堆栈位置/数量/退出码不变。完整入口失败不代表健康历史分类器通过；定向绿灯不代表宿主验收通过。

- [clean-before.log](clean-before.log)：未安装前HEAD、无node_modules、无NODE_PATH、工具链和锁/材料摘要。
- [frontend-before.log](frontend-before.log)：真实旧入口192/191/1缺包红灯，make exit2。
- [dependency-regression-before.log](dependency-regression-before.log)：新回归在原Makefile的红灯，Go exit1。
- [verify-before.log](verify-before.log)：首次原生完整入口源码戳、历史HTTP监听限制、exit2。
- [dependency-regression-after.log](dependency-regression-after.log)：修复后定向race顺序与失败停止回归exit0。
- [frontend-after.log](frontend-after.log)：实际新入口先npm ci，网络EPERM导致停止exit2，未运行frontend消费者。
- [static-after.log](static-after.log)：格式/module/vet/build/diff-check逐项exit0。
- [environment-after.log](environment-after.log)：网络失败后依赖状态及候选Makefile/测试/锁摘要。
- [source-after.json](source-after.json)：待宿主验证候选HEAD、142件源码摘要。
- [changed-files.log](changed-files.log)：最终diff-check退出0及实际修改/新增文件清单，无提交或发布。
- [development-check.md](development-check.md)：结果/范围/Trellis-check及正式宿主、QA、发布待办。

不把目录当成日志，不引用旧Run验收替代当前候选。后续测试Connector和QA在本目录或本Run其他实际文件保存新回执并更新索引；平台命令日志若截断需read_command_output续读，不能推断丢失内容为成功。
# 执行9历史证据入口

- [remediation-development-9.md](remediation-development-9.md)：固定入口实施、诊断、Trellis-check与宿主归档契约。
- [source-development-9.json](source-development-9.json)：变更后的候选身份；旧142件8e82aaa仅历史。
- [development-9-tests.log](development-9-tests.log)：定向race回归完整输出；夹具非全量冷安装。
- [development-9-tests-final.log](development-9-tests-final.log)：补齐结束状态/版本检查后的定向race复验；保留前一次输出。
- [development-9-tests-candidate.log](development-9-tests-candidate.log)：清理状态措辞修正后的中间复验。
- [development-9-check.log](development-9-check.log)：最终候选race、静态门禁及源码前后匹配检查；此前定向回执保留。
- [development-9-static.log](development-9-static.log)：格式、module、vet、build、diff实际检查。
- seq8由Owner stop，exit-1、102563bytes、truncated=false；中断在浏览器，不通过。seq5原日志API及QA6原报告继续保留，不改历史回执。
- 正式tests仍只执行 `make verify`。入口打印本次 `test-results/verification/<unique>/`；其中verify.log/state.json和审查后的具体browser文件须归档本Run并增补索引，未生成前不作为已存在证据。

# 执行11历史证据入口
- [remediation-development-11.md](remediation-development-11.md)：正式npm前置缺陷、最小诊断、修复边界、Trellis-check及原固定入口宿主下一步。
- [tests-10-verify.log](tests-10-verify.log)：从正式runner目录直接复制的40bytes完整失败日志，非手抄摘要。
- [tests-10-state.json](tests-10-state.json)：同次失败真实状态、时间、144件旧候选身份及日志摘要；gate_exit=-1，不是完整冷绿。
- seq10外层完整239bytes平台输出已read_command_output读EOF，正式API路径见执行11记录；没有用本机诊断补写旧宿主stderr。
- [development-11-npm-before.log](development-11-npm-before.log)：实际npm重复配置最小复现exit1，无网络安装/监听。
- [source-development-11.json](source-development-11.json)：当前144件待宿主候选，与执行9摘要不同。
- [development-11-tests.log](development-11-tests.log)：当前候选定向race及真实npm解析测试exit0；不等于完整冷安装。
- [development-11-static.log](development-11-static.log)：当前候选格式/module/vet/build/diff及最终摘要，overall_exit0。
- [development-11-tests-initial.log](development-11-tests-initial.log)、[development-11-static-initial.log](development-11-static-initial.log)：本轮较早定向成功回执，非最终候选完整门禁，不覆盖当前源码戳。
- [development-11-tests-intermediate.log](development-11-tests-intermediate.log)、[development-11-static-intermediate.log](development-11-static-intermediate.log)：前置日志写入失败传播显式检查前的中间成功回执，最终候选以无后缀日志为准。
- 新正式完整日志/浏览器回执尚未产生，不列未存在文件，不用目录冒充日志。任务仍In Review，QA与正式部署仍待执行。
