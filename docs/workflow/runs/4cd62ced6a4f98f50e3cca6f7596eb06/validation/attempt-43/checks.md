# 执行43：visual-current-period-r1

2026-10-08；Run4cd62ced6a4f98f50e3cca6f7596eb06；development。HEAD `98511771871cf0951ecef716bbb55deb13e18da5`，当前129文件工作树 `31bbf962868170e32d089a83217b8f057afce3a359fa9f0d5b25c1f4089d6bc5`，见source-evidence.json/source-after.json。**T001 In Review / M01 Ready for Review；QA38-01 Open / No-Go；当前完整门禁、十页三视口独立QA、发布均Not Run**。本轮是已证测试准备缺陷的修复候选，不是整套产品验收。

## 已恢复的事实

- 承接attempt41的diagnostic候选和QA38整改；没有恢复其他Run或旧checkout，没有重复读取/发送Issue评论，没有Git提交/分支/推送/部署副作用。
- seq42正式`make verify` exit2，平台归档103232 bytes/truncated=false，指纹129文件fb766c0aca5f33fbb0960e13286a13461d20999485e93a19a841eb51114f92a4。精确读取最新失败尾段offset98304→103232/eof；offset0页仅用于版本与历史红重建核对，本节点不将被显示省略的中段称为完整重新审查。完整原日志仍由平台seq42保留。
- 实际既有test-results/visual-1280/review.json与最新continuous-1280-30213.json复制为本目录host-seq42文件，其字段与最新正式尾段一致：continuous1280 J01–J07全检查点后失败；visual pages=[]/overview/capture-facts/固定ready断言失败；1440/390新连续及十页采集未到达。日志前面的旧健康版本预期红不是本次根因。
- 安全DOM：HTTP overview GET200/scope_matches=true；title/scope/role、rootFits/secretsClear、7项正文/3个enabled动作通过；loading=false/error_nodes=0/notice=false；总计与可见empty各1，seven_day_no_calls=1/no_data=0/other=0；1280 viewport/root1280、content990。因此本次实际失败是合法近7日空态碰到非空采集断言，不是此前J06预期的账单404，更非已证供应商错误。
- 源码归因：J06为结账推进到下一月，overview以服务端当前UTC日和近6日前下界聚合；生产workspace在无近7日调用时正确显示固定空态。采集器原ready拒绝任何.empty。三者与宿主事实一致，缺口在`journeys.mjs`的当前周期活动准备，不在Go计费或页面空态规则。

## 变更边界及实现

最小缺口：非空视觉采集发生在J06跨月之后，原旅程当月无请求。不降低要求：保持console-visual.mjs/所有字段、row数量、角色scope、真实请求/预算/账单/秘密/布局门槛不变。

仅改`tests/browser/journeys.mjs`、`journeys.test.mjs`及`.trellis/spec/frontend/workspace-contracts.md`，维护现有公共T001/总览；没有生产Go/UI/Makefile/DB/测试钟/包锁修改。新`prepareVisualActivity`在J06全部原断言和跨租户账单404之后执行：

1. GET证据核对当前日/月、当天0调用和空趋势；保留合法空态事实，不伪装为非空。
2. 用既有admin身份/当前租户及已轮换current.key，调用原生调用测试UI的`call`；既有helper要求proxy200、普通协议及真实受控上游dispatch差1。业务创建不经API seed或SQL。
3. 原生同Request ID详情要求当前周期、无待确认、唯一charge fee21/cost7（J06已发布新价格）；保持历史同ID及费用断言不动。
4. authoritative overview核对当天1调用、100%成功率、确认费21、唯一当日趋势1调用/1成功；结束周期bill前后完整对象深等，包括原金额、closed_at和更正列表。
5. 原生导航离开调用页清理秘密/输出，再执行原非空十页采集。不在截图前改DOM、渲染演示值、倒退钟或修数据库。新的白名单activity回执只含数值和守恒boolean，进入原make verify输出及既有test-results目录。

## 实际局部检查

| 命令/检查 | 真实结果 | 证据/边界 |
|---|---|---|
| `node --test tests/browser/journeys.test.mjs` | 16 tests / exit0 | targeted-first.log及targeted-check.log；第一次shell包装用了zsh只读变量status而exit1，但测试日志16/16通过；纠正变量后命令实际exit0，不隐藏包装失败 |
| `make frontend-test` | 179 tests / exit0 | frontend-check.log；原176加3项实际准备函数Contract，不是原生浏览器证明 |
| 两个修改JS的`node --check` | exit0 | static-check.log |
| `make format-check` / `go mod verify` / `make vet` / `make build` / `git diff --check` | 各exit0 | static-check.log；未改Go故未另重跑需要监听的全仓测试 |
| manifest19项长度/SHA及archive SHA | 全一致 | material-integrity.json；未执行材料脚本 |
| `go run ./tests/evidence`前后来源 | exit0 / 相同指纹 | source-evidence.json/source-after.json |
| 本轮完整`make verify`及native1280/1440/390 | Not Run | 原生稳定监听限制沿attempt41及历史首次失败保留；环境未变化，不重复必败命令；按host-retest.md交tests实际完整执行 |

新增3项回归在既有入口内：执行真实准备函数及原旅程绑定，要求调用/证据读/清理路径；拒绝缺失/旧周期/错误费用/未知用量；拒绝已结账原金额或更正变化。VM读/call夹具仅证明该函数断言，不证明网络、Chromium、费用供应商事实。最新seq42保留为真实旧候选RED，不把新局部绿写成宿主修复完成。

## Trellis / 公共交付审查

使用managing-engineering-delivery-cn并读public-delivery-contract/delivery-review-checklist，恢复T001、总览、QA38矩阵和整改、architecture-r1/page-mapping及当前失败；不新建重复任务或另造PRD。保持AC01–AC18/H01–H08及八类商用/J01–J07五层状态，未验项目不提升。

trellis-before-dev：实际get_context.py packages返回单仓/backend/frontend；读取frontend index、DOM/lifecycle/verification/workspace、共享reuse/cross-layer指南，并追踪overview实际Go响应与UI范围。现有Trellis规范已经是有源码依据的契约，不是待填模板，沿前阶段bootstrap成果，不宣称prepare首次initialized=false即已初始化。

trellis-check：检查本轮差异为测试驱动和契约，不改变HTTP/CSRF/存储/权限；新增回归、JS语法/项目静态检查通过；沿trellis-update-spec记录跨月→当前周期真实活动→原非空采集的七项契约。所有原生产解析/限额/费用/迁移/故障门禁保留。固定完整测试责任明确在宿主；本节点未新增依赖、许可、凭据、绝对本机路径或调试输出。

权威材料保持design-v0.2.0/frozen-for-test-implementation，archive5f51d35fdc15ed7214749f9413879b4238097f1d6b6bbe2cc0d508d8e103cef7；入口为prepare指定`.workflow-input/4cd62ced6a4f98f50e3cca6f7596eb06/5f51d35fdc15ed7214749f9413879b4238097f1d6b6bbe2cc0d508d8e103cef7/files/manifest.json`。实际本轮引用其docs/prd.md、docs/interactions.md、docs/acceptance.md和同Runarchitecture-r1/verification-design/page-mapping；19文件完整路径/摘要见material-integrity。Codex/gpt-6.1-sol为用户指定及平台配置，本节点无法独立核对底层实际模型，不自造模型运行回执。

## 交接与退出

关键路径：tests对当前同指纹执行原完整make verify并保留新活动/十页/三视口产物 → 失败按首次真实边界回development，或完整exit0交独立QA逐页核对原型与AC01 → Connector负责发布。不得跨用seq37绿色或以179 Node关QA38-01。暂无需求/架构澄清、普通选择审批或权限扩展。当前独立QA/供应商/外部支付身份/正式Actions及生产URL核对仍未执行本候选；不宣称商业全面对齐或已发布。有效同Run60步授权沿用，本次43，不新Run、不重放副作用。
