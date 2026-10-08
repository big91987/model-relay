# development6：消费首个宿主失败后的候选修复

## 结论与版本
- Issue30 / Run `2574d6be275f9eb630f8738837507e1e` / 冻结工作流2；HEAD保持正式main `9bf6ecaf53758ca2258498968b830c535e02e1c0`，没有研发提交或发布副作用。
- 当前150文件候选源码戳：`0384f115adbbd545e8ce01021f238010ba66c89eb06e6acea08ce872ce0fa474`。项目原 `go run ./tests/evidence` 生成的 [before](candidate-06-before.json)/[after](candidate-06-after.json) 在最终局部检查前后相同；交接后tests必须自行重新锁定/测量，不用这个局部结果替代宿主门禁。
- 实施结果：修复候选已完成，**修后真实Chromium及完整make verify尚未运行**。T001 In Progress / M01 Active。R28-04仍Open/P2，旧tests18根因未知；当前没有释放后持久最终错状态的P1实证。

## tests5完整真实失败的消费与归因
- 正式固定入口 `make verify`，UTC `2026-10-08T21:47:11.167634Z` 至 `2026-10-08T21:57:26.682523Z`。项目gate exit2 / runner exit1，Connector exit2。这是失败，不是预期红计通过。
- 完整 [verify.log](host-tests-05/verify.log)、[state.json](host-tests-05/state.json)、[archive-index](host-tests-05/archive-index.json) 原样归档。项目日志329252字节、SHA256 `a5a3c4b6d12086c04108c7c474a9484232414e27051522fd303072ffcae3ab59` 与state吻合；Connector包装输出329434字节与项目owned日志不是同一字节流，不能互相替代。10个原始归档文件逐一与保留的test-results源文件比对字节摘要相同，无截断推断。
- 修前源码149文件 `ba8816e8f4cb38bb3b6c782195da2324bd16964f98cc2ff00448d91e78940d6f` before/after不变；Go1.25.6、Node v26.0.0、npm11.12.1。cold start实际测得无node_modules/npm cache，NODE_PATH unset，无手工安装；候选移除且owned process group absent=true。此前回执未记录实际Chromium版本，明确NotRecorded，不从依赖锁倒推；新回归实际记录browser.version()。
- [model-save-readback-1280.json](host-tests-05/model-save-readback-1280.json)四组实际UI PATCH200（1280/390，各false/true）均在同角色同scope tenants真实响应held时save_completed=true、dialog=false、loading=false；catalog held时save仍完成、loading=true且无行。两次hold分别原样转发真实响应，无catalog伪造。
- 每组释放后严格最终唯一行/列/状态均通过，cleanup=true。因此本次已证原因是“驱动的dialog-hidden/loading-hidden完成条件可在本次读回前成立”，不是实证持续错误数据/渲染。候选修复依据是这组实测；不能把它当旧tests18根因或旧缺陷已关闭。
- 原 [continuous-1280-79647.json](host-tests-05/continuous-1280-79647.json)保留 `console-summary.mjs:32:12` 的原AssertionError：row_count1、column_count8、状态列存在、期望停用却观察启用。顶层stage是四组诊断结束时的enable，原错误实际来自首组disable的tenants阶段；不能用最后stage覆盖原错误归属。
- 原门禁在full-console1280第一次诊断后红退出；此后的1440/390连续J01–J07、summary、visual未运行，不能把此前独立legacy/routing/console三视口通过外推为完整连续旅程通过。三视口 [editor interruption](host-tests-05/editor-interruption-1280.json)与 [fault classifier](host-tests-05/fault-1280.json)回执及1440/390对应文件一并保留。

## 最小实施与回归
- 只改产品 `internal/relay/web/workspace.js` 既有edit/syncTenants/renderPage：复用operation UUID、epoch、loading、busy。普通成功写后，表单内部readback pending覆盖tenants及页面全部读与DOM提交；complete仅在本scope成功commit后成立。readback期间不展示旧表、不释放保存锁、不允许同scope替换编辑器；导航invalidate仍允许新scope编辑。
- syncTenants在collection后检查epoch，保存者在sync/render后再次检查epoch。旧scope异步结果/旧finally不能render新scope或解锁新编辑。没有新增通用刷新领域对象、公用API协议或配置框架。
- GET失败是“写入已确认但读回失败”，清旧数据、failed终态、显式GET-only重试；不转成写未知、不自动重发写、不把原失败保存后来改为成功。renderPage明确返回是否完成DOM提交，原调用者可忽略返回值。
- `tests/browser/journeys.mjs` 的save在点击前捕获operation/tenant/page；`waitForSaveReadback`只等待关联终态/owner替换，然后各断言一次，失败/换scope/换editor立即拒绝。秘密/邀请/unknown-write的原规则保持。
- 原 `model-save-readback.mjs` 诊断转成回归：两个真实hold阶段均断言同scope/page、pending、loading、save锁与未完成，再释放真实response，原严格summary只在完成后断言。当前宽度和390各两方向保留；不删原断言、不sleep/networkidle/超时扩展/重试assert。
- 增加分别tenants/catalog的**显式注入GET503**，要求save拒绝、error视图无旧行、写已确认notice、用户GET-only恢复与一次严格DOM核对。它不是真实存储故障证据。
- 真实catalog快照hold跨“当前租户→管理员已有全部租户→原租户”scope切换，原save必须拒绝。第二次相反UI保存成功后再释放第一次快照，第二次状态与scope不能被覆盖。这是待宿主证明的真实浏览器回归，不拿VM解释Chromium取消机制。
- `save-lifecycle.test.mjs`覆盖实际edit/sync/render函数的锁、重复submit、busy替换、迟到tenant/catalog、新editor隔离、GET失败与secret路径；readback/workspace测试只适配DOM fixture。原summary错状态/缺列/重复行负例与原错误身份、历史红/UTC/零预算/迁移/幂等/所有用户旅程均保留。

## 本节点实际验证（非完整门禁）
| 检查 | 真实结果与证据 | 限制 |
|---|---|---|
| 定向Node 6文件 | 最终75/75，exit0，[local-unit-06.log](local-unit-06.log) | 函数/驱动契约，不是Chromium |
| 可执行Node子集 | 最终211/211，exit0，[local-node-subset-06.log](local-node-subset-06.log)，UTC22:14:04Z | 仅排除依赖尚不可用的route-cleanup.test.mjs；不是make frontend-test通过 |
| format / vet / build / go mod verify / JS syntax / diff | 各exit0，[static-checks-06-final.log](static-checks-06-final.log)，UTC22:14:04Z–22:14:07Z | 没有全包测试或真实HTTP |
| 原controlled响应契约 | `go test -race -count=1 ./tests/browser -run '^TestContinuousUpstreamResponseContract$'` exit0，同上 | 非监听函数契约；非全包race、非供应商 |
| 原make frontend-test | exit2，[frontend-checks-06.log](frontend-checks-06.log)：标准npm ci连接127.0.0.1:7897 EPERM，完整Node门禁未开始 | 环境连接限制，未绕过/改安装配置/扩大权限；不重试求绿 |
| 本轮native完整make verify | NotRun；复用首次 [native log](native-verify-04.log)/[state](native-verify-04-state.json) 的historical listener unavailable | 稳定监听限制且环境未变；完整验证留正式tests，不计通过 |

可执行Node子集的明确命令为 `node --test $(rg --files tests/browser -g '*.test.mjs' | rg -v '/route-cleanup.test.mjs$')`；这是补充检查，**Makefile的全量通配门禁没有改动**。宿主仍须包含route-cleanup和锁定Playwright依赖。

首个新VM测试失败及修复保留：[initial](local-unit-06-initial.log)（72/73、exit1，测pending时VM宿主异步submit尚未进入sync）、[later](local-unit-06.log)。实际收集fixture调用边界后，改为等待sync/render函数入口deferred，不用sleep或假定微任务数量，不改生产时序来迁就fixture。更早静态/source与Node检查分别保留static-checks-06.log、static-checks-06-intermediate.log、candidate-06-intermediate.json、candidate-06-static-before/after.json、local-node-subset-06-initial.log，不覆盖曾发生的检查。

局部ld保留LC_DYSYMTAB警告（exit0），不冒称无警告；本机绝对workspace/temp路径单独脱敏，命令/退出码/断言未删除，[local-redactions-06.json](local-redactions-06.json)记录原/归档字节摘要。原宿主日志无此路径替换，完整保持原字节。

## Trellis与交付检查
- 恢复T001/M01与现有总览，不新建重复状态源；已读挂载研发交付Skill及公共契约/审查references，按单Bug单任务裁剪，不另写PRD/HLD/组织级发布清单。
- trellis-before-dev先发现single-repo backend/frontend，读取frontend索引、实际生命周期/验证/依赖/工作台规范及共享thinking guides。checkout已有Trellis初始化文件与非空源码依据规范，prepare的initialized=false不被改写为本轮新初始化；无需bootstrap或重建既有规范。
- trellis-update-spec更新workspace契约，保留修前诊断说明与真实证据边界。trellis-check复核变更范围、operation/epoch/read/render/finally数据流、依赖/复用、秘密/错误/权限、原门禁和严格负例。无HTTP/schema/鉴权/锁依赖/Connector/时限变更，Go真实后台保持；本节点没有借静态browser.check证明API。
- 固定完整入口仍为make verify。实现、Node契约通过和局部构建不等于本里程碑Accepted；独立QA负责最后逐AC判断。

## 正式tests交接要求
1. 对上述当前候选运行一次原固定 `make verify`；保持cold installation、源码before/after、候选锁、原全包race/真实HTTP、UTC/零软预算/迁移、1280/1440/390 J01–J07/summary/visual及清理。不要扩Connector权限/预算/超时，不手工安装、不从旧绿灯推定新版本。
2. 将实际生成的model-save-readback各视口JSON（含实际Chromium版本）、continuous各实际pid JSON、完整verify.log/state/source/清理归档本Run具体非空路径。故障注入与真实Go响应、product断言与route转发/取消/环境错误分别归因。
3. 若失败，保留首个实际失败的原错误/安全JSON/UTC/具体阶段返回development，优先收集区分原因的最小事实，不无限全套重试。释放后最终状态持续错误须按管理核心用户影响评估P1。完整通过再独立QA，不能跳到发布。
4. 动态独立QA仍Blocked（待本Run真实修后门禁/独立复核）；真实供应商NotRun；Owner实际登录预览验收NotRun。用户已授权后续发布仍由Connector/协调者执行，本节点不动DB、凭据、旧部署、PR29、Platform或部署控制器。
