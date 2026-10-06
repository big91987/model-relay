# 首次可追踪调用交付计划

- 当前状态（2026-10-06第二轮独立QA）：**Go，本切片验收完成**。T001/T002/T003 Done；M01/M02/M03 Accepted，不是Delivered/Published。宿主seq33原样make verify退出0，56文件摘要 `1b2f7026692d0633f6f050d585cec0bb67571a7ecf411848d8213d2fd609cdc6` 与QA重算一致；见[当前验收报告](../acceptance/report-02.md)、[矩阵](../acceptance/matrix-02.md)。后续report/PR/合并/正式部署按真实回执推进；供应商和路线2–4未完成。以下所有待验/No-Go状态为历史记录，不覆盖。

- 2026-10-06 seq31真实逐键证据已归档并核对精确行摘要：390px箭头到达同一就绪select但未移动，无重绘/失焦/读取竞争证据。T003限定改原生type-ahead与精确值/索引门禁，108契约及分项通过，新对象两屏和完整make verify仍待宿主；任务In Review、M03 Ready for Review，QA No-Go不关闭。当前关键路径宿主新指纹验证→必要返工→独立QA。见[本次返工](verification/model-typeahead.md)，以下保留历史，不用于当前对象放行。

- 2026-10-06补充回执容量约束：最终整行UTF-8（前缀/JSON/换行）限定8192字节，保留关键阶段首末观测与省略计数，完整有界白名单JSONL给出相对路径及行摘要。107契约通过，非原生根因证据；当前对象指纹与实测边界见[诊断报告](verification/model-selection-diagnostics.md)，仍交原make verify宿主复验，独立QA未放行。

- 2026-10-06用户补充覆盖上条类型前缀建议：该未宿主方案已撤回，保持seq29原箭头序列，先在原browser入口采集到达／逐键／Tab／下一帧焦点、控件option身份、重绘、navigation/load及真实models/overview读体计数。成功/失败均白名单JSONL留存；106契约只验证观测与压缩投影，不证OS根因。当前新对象交宿主原样make verify收集诊断，未改产品、门禁或QA资产，任务In Review、里程碑Ready for Review、QA未放行。见[当前诊断与交接](verification/model-selection-diagnostics.md)。下条建议与指纹已被取代，旧证据不覆盖。

- 2026-10-06 seq29返工：原对象make verify退出2，1280px完整旅程通过，390px箭头选择仍停index0而目标为启用index1。本轮只改T003原生类型前缀选择与离焦前安全索引诊断，不改产品/门禁/QA资产；新增helper红转绿，102/102与分项通过，新对象完整两屏待宿主。任务In Review、里程碑Ready for Review，QA No-Go/QA-03语义差异保留。见[当前证据与交接](verification/host-return-11.md)，以下旧对象不用于新对象放行。

- 2026-10-06既有契约收尾：按维护者静态反馈修复GET-only撤销/共享保存恢复忙态、实际调用参数锁定、同记录created/未知、按alias/key ID即时行焦点恢复；不是新原生回执。101VM/helper、3Go in-process race及分项通过；原QA脚本恢复执行仍3/4，QA-03保持独立语义待评审，未改QA资产。新两屏/完整make verify交宿主，任务In Review、里程碑Ready for Review，QA No-Go与历史保留。见[当前对象与证据](verification/closure-contracts.md)；下方旧对象不能放行本对象。

- 2026-10-06 seq27：宿主make verify退出2，390px启用目标index1仍停index0；旧OS弹层选择序列未奏效。本轮仅T003改闭合ArrowUp/启用序号ArrowDown/Tab，保留精确值/焦点/全部业务断言，85VM/helper与分项通过，新对象两屏待宿主原入口。任务In Review、里程碑Ready for Review，QA No-Go/驱动差异保留，见[最新记录](verification/host-return-10.md)。以下为历史，不用于新对象放行。

- 2026-10-06 seq25返工：1280px完整控制台/QA原层新增旅程通过，390px在模型选择最终值断言失败（不是焦点assert）。T003修正原生下拉提交边界并补独立脱敏选择诊断；77产品VM与4旅程helper契约通过，新对象两屏/完整make verify仍待宿主。T001/T002/T003 In Review、M01/M02/M03 Ready for Review；关键路径宿主新指纹验证→独立QA，QA No-Go/驱动差异保留，见[最新记录](verification/host-return-09.md)。下方均历史，不将seq25部分通过用于新对象放行。

- 维护者补充交互§3：概览成功行补“调用时间／配置版本”，只引用该次call_evidence，不用当前时间或全局revision；旧值清除/未知/迟到隔离及两屏真实overview→同ID精确记录核对进入原门禁。77VM和分项通过，新指纹完整门禁/QA仍待验证，不干预在途原对象测试。状态不升格，见[补齐记录](verification/evidence-metadata.md)，后文为历史。

- 最新seq23返工：重读后等待旧credential文案可提前返回，迟到GET覆盖下一无效地址草稿。三次重读统一真实GET屏障、pending/settled及持久值核对，保留中文错误并新增实际PUT400；75VM与分项通过，新对象两屏及完整make verify待宿主。关键路径仍宿主验证→独立QA，任务In Review、里程碑Ready for Review，见[最新记录](verification/host-return-08.md)，后文为历史。

2026-10-06最新seq21固定make verify退出2：Go/race/HTTP、build及1280px版本通过，复制密钥成功反馈超时，新增QA两屏未完整完成。本次T003明确隔离origin剪贴板权限，强制拒绝恢复及原生内容核对，不改产品权限/门禁；74VM和分项通过，新对象待宿主。T002/T003 In Review、M02/M03 Ready for Review，不关闭QA。关键路径宿主→必要返工→独立QA。见[最新证据](verification/host-return-07.md)，以下均历史。

2026-10-06当前seq19整改实现完成：T002/T003 In Review，M02/M03 Ready for Review；74VM、3条不监听Go race契约及分项检查通过，原独立QA-01/02/04红转绿，QA-03旧driver激活泛导航仍红且原脚本未改，新增明确概览动作回归独立覆盖。完整新对象 `make verify` / 两屏原生异常和分页待宿主，旧seq18绿色不适用。关键路径：宿主固定门禁→必要返工→独立QA复评（含QA-03驱动/动作核对）。见[本轮证据与范围](verification/qa-rework-01.md)。以下seq19首轮及更早内容均为历史，不关闭QA/任务/里程碑、不发布。

2026-10-06独立QA第1轮：seq18当前指纹固定 `make verify` 退出0、真实Go两屏通过；独立新增失败清密/概览同ID动作P1、保存后读取反馈P2及完整键盘/恢复覆盖缺口，结论No-Go、返研发。T001 In Review（后端已验证），T002/T003 Needs Rework；M01 Ready for Review，M02/M03 Needs Rework，不完成/发布。见[验收索引](../acceptance/README.md)/[整改包](../acceptance/remediation-01.md)；以下为历史，不以“最新seq16待宿主”作当前事实。

2026-10-06最新seq16宿主退出2：Go race/HTTP、build、1280px版本与完整控制台旅程通过，但390px登录后概览等待旧关键词超时。T003按共享Go服务真实成功状态修正两次登录响应/文案/计数核对，保留首次全UI与安全门禁；60VM及分项通过，新对象两屏make verify仍待宿主。任务In Review、里程碑Ready for Review，关键路径仍宿主验证→必要返工→独立QA。见[最新证据](verification/host-return-06.md)，下方为历史。

2026-10-06最新seq14宿主 `make verify` 退出2：当前Go race/HTTP与build通过，1280px版本旅程通过，随后通用键盘焦点合并断言失败，未提供目标/焦点/outline分项。T003将程序化focus改为有界真实Tab到达目标，保留并增强焦点门禁及白名单诊断；59VM和分项检查通过，不称原生缺陷已修复。关键路径仍为新指纹对象宿主原样make verify→必要返工→独立QA；旧seq12及其余证据保留。见[键盘返工记录](verification/host-return-05.md)。任务In Review、里程碑Ready for Review，后续历史段不代表最新状态。

2026-10-06。研发准备结论：Ready for Dispatch。需求与设计引用本 Run 的
[PRD](../product-definition.md)、[架构](../console-architecture.md)、[交互](../console-interaction.md)；不重建项目路线。

## 边界与准备

只做路线第1项，Go / SQLite / 单上游 / 原生前端。变更在 server.go / proxy.go / 新增只读查询、web 三文件及项目测试；不迁移表、不放宽鉴权、不探测真实供应商、不发布。
行为缺口位于管理查询的真实聚合、安全拒绝持久化和浏览器页面/异步所有权，不以 CSS 或假数据遮盖。
prepare 曾报告 initialized=false，但当前 checkout 已存在 .trellis/.version、scripts/get_context.py 和有源码依据的前后端规范；实际 packages 命令成功。按 before-dev 读取全部适用规范，bootstrap 不适用，不重建初始化。

## 进度与关键路径

当前为G4发布前交付审查：实现与独立验收已完成，交付判断仍 **Verifying（仅后续发布闭环待完成）**，不标Delivered/Published。第二轮[QA报告](../acceptance/report-02.md)与[18AC矩阵](../acceptance/matrix-02.md)为当前退出依据；T001/T002/T003 Done，M01/M02/M03 Accepted。无本切片实现验收阻塞，不再等待已完成的宿主验证或独立QA。

剩余关键路径：报告收尾→发布Connector提交/推送/草稿PR真实回执→获授权维护者核对QA后合并测试仓→既有工作流准备→显式正式测试部署→版本SHA及部署后浏览器核验。当前Run没有下一批Ready研发任务；路线2–4由维护者按[唯一路线](../../../../roadmap.md)在依赖完成后另立Issue接续，不在本Run预建或实施。

本节点仅同步公共状态与PR正文；不改产品、测试、PRD/设计、可信门禁或权限。[交付审查](verification/report-closure.md)区分QA已验快照与收尾文档摘要，不将新文档摘要说成已运行make verify。研发/宿主/QA历史失败、seq7尾部丢失和seq11未知结果全部保留；原生沙箱限制不要求用户扩权。

| 里程碑 | 可观察结果 | 进入 / 退出 | 依赖 | 状态 |
|---|---|---|---|---|
| [M01](milestones/M01-traceable-facts.md) | 当前配置与同请求诊断有真实事实源 | G2基线 / 聚合、隔离、拒绝及权限回归；report-02 | 无 | Accepted |
| [M02](milestones/M02-first-call.md) | 管理员完成首次配置与普通/SSE调用 | M01契约 / 全UI真实Go两屏旅程、取消晚到保密；report-02 | M01 | Accepted |
| [M03](milestones/M03-verifiable-delivery.md) | 可复验交付给项目测试 | 实现 / seq33固定门禁退出0及独立QA Go；report-02 | M02 | Accepted |

任务状态以 [T001](tasks/T001-query.md)、[T002](tasks/T002-console.md)、[T003](tasks/T003-regression.md) 为事实源。

| 任务 | 结果 / 里程碑 | 负责人 / 评审 | 状态 | 硬依赖 | 当前验证 |
|---|---|---|---|---|---|
| T001 | 真实配置和同请求事实 / M01 | 研发 / 项目测试、独立QA | Done | G2 | report-02后端与HTTP对账 |
| T002 | 首次配置到可读调用 / M02 | 研发 / 项目测试、独立QA | Done | T001契约 | report-02两屏主链及失败恢复 |
| T003 | 固定门禁覆盖闭环 / M03 | 研发 / 项目测试、独立QA | Done | T001、T002 | seq33退出0，独立108/3/5分层回归 |

主要风险：沙箱监听受限须由宿主测试运行；真实供应商未测；成功证据必须对配置变更/重启保守失效；秘密不能进入证据。

## 历史实施与验证进度（非当前状态）

2026-10-06更新：实现和自动回归已存在，固定门禁受监听限制未通过；实际命令、失败、补充成功和全部AC映射见[研发验证](verification/development.md)。无需退需求或设计；按合法next交宿主tests，再由独立QA判断。不宣称M01–M03 Accepted、任务Done或产品全量完成；路线2–4仍未实现，按项目唯一roadmap接续。

2026-10-06维护者补充已在同Run处理：T002补齐合法工具数量/有限名称摘要及30秒会话安全检查；T003新增对应VM/真Go旅程，不放宽既有PRD/设计或门禁。最新48项VM与分项检查通过，完整门禁和真实浏览器仍受沙箱监听限制，关键路径仍为宿主tests→逐条18AC独立QA。见[补充验证及当前指纹](verification/supplement-rework.md)，首轮失败/分项回执保留。

2026-10-06第二次补充：已知到期改为独立会话级计时器，不受悬挂session请求或跨页影响；正常周期检查保留真实调用状态，仅不明状态恢复时提示重新输入。一次性timer夹具先复现失败，再扩展到58项VM通过。当前指纹及新回执见[悬挂检查验证](verification/expiry-watchdog.md)。首轮及第一次补充证据均保留；宿主make verify与独立QA仍未通过，公共状态不升格。

2026-10-06宿主seq7返回退出2，但会话回执在历史HTTP事件中截断，未提供最终失败阶段或错误。已恢复同指纹、复验分类器和58项VM等分项，make verify第8次仍受沙箱监听限制。详见[宿主退回诊断](verification/host-return-01.md)。当前等待完整宿主输出（至少最后80行）；不把历史预期红测误修成绿、不凭截断前缀猜测根因、不修改门禁或平台、不重复交接无修复版本。

2026-10-06最新维护反馈取代上一段的等待动作：seq7尾部已丢弃、不能恢复；维护者报告唯一平台源已修复首尾日志与截断前脱敏，d5e1403已通过原manifest升级开发实例。该信息为维护者回报，不是本项目门禁通过。本次仅重算56文件指纹并核对git状态，仍与seq6/7及第8次证据一致，无未解释产品变更；不重复监听受限命令。保留全部旧证据，T003回到In Review、M03 Ready for Review，正式next交固定make verify取得新宿主结果。见[新宿主交接记录](verification/host-return-02.md)。不猜测修复、不放宽门禁、不请求已丢失日志；真实新失败按正常路由返工。
