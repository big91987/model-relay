# design-v0.2.0 当前 Run 验收追踪矩阵

## 执行56当前证据覆盖

当前records-detail-entry-r1，指纹见[attempt56 source](validation/attempt-56/source-evidence.json)。seq55上一指纹完整make verify exit2，编辑中断/console三视口通过，continuous1280 J01–J05/J07通过但J06核对后日志详情入口误用；management-summary及最终63图未到，不能据分段通过关闭QA47。仅驱动按当前行入口修复，193Node/静态检查exit0；本候选完整宿主、独立QA Pending。QA47-01 P1 Open/AC01 Fail、QA47-02 P2 Open保持。下面AC01–AC18/H01–H08及商用功能/旅程五层矩阵原职责继续，不提升通过或发布级别；供应商/外部身份支付、平台模型及负向Harness、PR/merge/Actions/实际部署页面仍Not Run。复验要求见[checks](validation/attempt-56/checks.md)与[seq55](validation/attempt-56/host-seq55-summary.md)，继续同Run正式部署，不以旧main/临时服务替代。以下旧“当前”均为历史。

## 执行54当前证据覆盖

当前editor-route-cleanup-r1，最终指纹见[attempt54 source](validation/attempt-54/source-evidence.json)。seq53原make verify exit2：1280受控中断原断言已通过，但后续路由清理失败，非监听限制。最小诊断及顺序清理修复后189Node/静态检查exit0；当前完整宿主/三视口console和summary/J01–J07/63图/独立QA仍Pending。QA47-01 P1 Open/AC01 Fail、QA47-02 P2 Open不升级；下方AC01–18/H01–08及功能旅程五层矩阵继续原职责，旧局部结果仅历史。供应商/外部身份支付、平台模型/负向Harness、PR/merge/Actions/实际部署页面Not Run。下一按[checks](validation/attempt-54/checks.md)与[seq53原失败](validation/attempt-54/host-seq53-summary.md)复验、独立QA及同Run正式发布，不能以旧main/临时页面替代。

## 执行52当前证据覆盖

当前editor-readback-route-r1，指纹见[attempt52 source](validation/attempt-52/source-evidence.json)。seq51上一指纹完整make verify exit2（POST201后受控读回检查点未到），不是监听限制；已证明分页查询与驱动glob不匹配并最小修复，187Node/静态检查exit0。当前完整宿主、三视口console/summary、J01–J07/63图及独立QA均Pending，不把旧部分PASS移植。QA47-01 P1 Open / AC01 Fail、QA47-02 P2 Open保留；全部AC01–18/H01–08及商用五层以下原矩阵继续适用，不提升通过级别。供应商/外部身份支付、平台模型/负向Harness、PR/merge/Actions/实际部署页面仍Not Run。下一按[checks](validation/attempt-52/checks.md)完整宿主复验，之后独立QA，正式部署实际页面要求沿原report→publish→pr。

## 执行50当前证据覆盖

当前editor-lock-r1。seq49原完整make verify **Fail/exit2**（attempt48指纹），不是本版Pass；真实首失败及局部红/绿见[attempt50](validation/attempt-50/checks.md)。已修一个跨导航新编辑器继承旧锁的产品缺陷，187Node/静态exit0。AC01–AC18/H01–H08、八类功能/J01–J07范围与下方映射不变，但下方执行48全部“当前宿主Pending”只属历史：本候选完整宿主仍Pending，独立QA仍Pending，发布Not Run；旧版本部分旅程通过不跨指纹继承。新增三视口editor受控hold/cancel断言以及attempt48的management-summary/全部原层门禁须原make verify运行。QA47-01 P1/AC01 Fail、QA47-02 P2保持Open，只有独立QA可关闭。未更改历史验收报告、冻结材料或费用数据；未自称商用对齐。

## 执行48当前版本证据边界

当前候选management-summaries-r1；冻结材料及完整AC01–AC18/H01–H08范围不变。[QA47独立矩阵](acceptance/acceptance-matrix-47.md)和No-Go保留；不是研发修改后的Pass表。下列当前状态覆盖本文件旧候选的“当前”描述，不覆盖历史证据。详细改动/实际局部结果/[新指纹](validation/attempt-48/source-evidence.json)见[attempt48](validation/attempt-48/checks.md)。

| 功能/关键旅程 | 实现 | 局部测试 | 当前宿主完整测试 | 当前独立QA | 已发布 |
|---|---|---|---|---|---|
| 服务接入/核验 J01 | 保留正式Go实现；补关联模型及核验时间/执行版本摘要 | 原契约回归 + 新真实值驱动；不能代替网络 | Pending，新指纹 | QA47-01待复验 | Not Run |
| 模型发布/授权 J01/J03 | 补来源/当前候选价/范围延迟、启停/上下文调用 | Go摘要/失效价/负向 + 新三视口驱动 | Pending | QA47-01待复验 | Not Run |
| Key创建/轮换/撤销 J02/J03 | 原共享代际实现保留；租户Key数摘要及入口 | 原不清零/诊断回归保留 | Pending | 不继承seq46/QA47旧版本通过 | Not Run |
| 真实身份/隔离 J01/J05/J07 | 实际成员派生负责人，限制角色不泄新摘要 | 新Go权限/未分配延迟负向；原原层隔离保留 | Pending | QA47-01待复验 | Not Run |
| 配额/预算 J04/J05 | 原准入保留；预算/用量/平台上界复用同事实源 | 原InProcess五维/共享资源及新摘要对账 | Pending | 当前待复验 | Not Run |
| 计费/账单 J06 | 不改写入；日志/详情/CSV附已确认Token | 原幂等/快照/追加守恒 + 新原始/确认口径 | Pending | QA47-02 P2待复验（不单独阻断） | Not Run |
| 用量日志/诊断 J06/J07 | 原始未知不覆写；确认多尝试累计独立展示 | 真实handler/SQLite对账；J06实际页面/分析/CSV断言待宿主 | Pending | QA47-02待复验 | Not Run |
| 表格/编辑/详情/移动端 J01–J07 | 原十页及全部原层保留；工作台模型状态/模型预算入口 | VM/精度/严格值断言；完整1280/1440/390待宿主 | Pending | QA47-01/AC01仍待独立判定 | Not Run |

本轮不把参考New API/LiteLLM的高级能力扩入冻结范围；既有后续差距记录继续有效。真实供应商/外部支付/外部身份及平台模型/负向Harness/正式发布事实仍未执行，不能以受控测试或局部通过自称商用对齐。

执行33当前：seq32同attempt31指纹原make verify exit2，新console三视口绿，continuous1280 J01–J05/J07绿，J06账本首条误作累计费用失败，continuous1440/390未运行。见[首失败](validation/attempt-33/host-seq32-summary.md)。本候选保留fee21并增强cost7/Token3+4/历史快照及差额断言，生产不改；[检查](validation/attempt-33/checks.md)及attempt-33/source-evidence.json绑定当前版本。AC12–14仍需完整实际旅程，QA22-01/02未关闭，独立No-Go和发布Not Run不变；全部AC/H、八类功能/J01–J07五层矩阵保留，不跨指纹复用宿主Pass。以下旧“当前”仅历史。

执行31当前覆盖：seq30同attempt29指纹完整make verify exit2，continuous1280 J01–J04过、J05未完，J06/J07与1440/390连续未运行；新console三视口部分Pass不等于完整Go。usage_facts守恒，旧resume标签不能定位到费用断言；见 [归因修正](validation/attempt-31/host-seq30-summary.md)。scope-refresh-attribution-r1只修已证driver刷新等待缺口并补成员/隔离定位及生产路径回归，未改生产策略，当前宿主仍待验。来源attempt-31/source-evidence.json、[检查](validation/attempt-31/checks.md)。AC/H全部矩阵、八类商用/J01–J07五层状态不减，QA22-01/02不关闭、独立No-Go及发布待验不变。以下旧当前/推断仅历史，不能跨指纹复用Pass。

执行29覆盖状态：seq28同attempt27指纹完整入口exit2；TestContinuousUpstreamRelayHTTP已知/未知实际通过，新console三视口通过，continuous1280/J01–J04实际通过，但J05-resume费用比较失败；J05未完成/J06/J07及1440/390连续旅程未到达。pause-conservation-diagnostic-r1仅补守恒回归和分段事实，不宣称修复或当前UI Pass；原生产代码未变，in-process费用守恒通过与native失败不能互相抵销。见 [attempt29](validation/attempt-29/checks.md)、[seq28](validation/attempt-29/host-seq28-summary.md)、当前来源attempt-29/source-evidence.json。AC/H全部矩阵、八类商用/J01–J07五层状态保留，QA22-01/02未关闭，独立No-Go与发布待验不变。以下旧当前仅历史。

执行27覆盖状态：seq26完整入口exit2，continuous1280/J01首调用502，J01–J07未完成、1440/390未到达。controlled-completion-r1修确定的受控响应role契约，不改生产；局部fixture/parser/Node通过不能升级为新UI Pass。宿主真实HTTP新增回归及完整三视口仍待验，QA22-01/02未关闭、独立No-Go不改。见 [attempt27](validation/attempt-27/checks.md)、[seq26](validation/attempt-27/host-seq26-summary.md)、[当前指纹](validation/attempt-27/source-evidence.json)。保留AC01–AC18/H01–H08及八类商用/J01–J07五层矩阵。以下旧记录仅历史。

执行25覆盖状态：seq24完整make verify exit2，新增continuous 1280/J01空态失败、J01–J07未完成；原三视口console通过不关闭QA22连续UI缺口。fresh-install-driver-r1修正确定的新安装/迁移契约错配；actual-init/SQLite/handler及actual-driver局部回归通过，当前完整宿主/独立QA/发布待验。见 [attempt25](validation/attempt-25/checks.md)、[seq24](validation/attempt-25/host-seq24-summary.md)、[当前指纹](validation/attempt-25/source-evidence.json)。AC01–AC18/H01–H08矩阵及八类商用/J01–J07五层表保持有效、未提升状态；QA22-01/02未Closed，不改独立QA No-Go。以下旧进度仅历史。

版本 requirements-r1；2026-10-07；Run 4cd62ced6a4f98f50e3cca6f7596eb06。

执行20状态：seq19在attempt18同指纹完整make verify exit2，已越过Key转交、普通调用和配额拒绝，失败缩至keyboard-1280。候选keyboard-ready-r1修复源码确认的异步dialog准备等待缺口（延迟打开actual-driver先红后绿），并增加安全细阶段/16条有界focus trace；未认定seq19唯一根因或产品焦点修复。原13 Tab/Escape/触发恢复断言、Go及三视口门禁完整保留。检查与来源见 [attempt20](validation/attempt-20/checks.md)、[seq19](validation/attempt-20/host-seq19-summary.md)及attempt20/source-evidence.json。AC01–AC18/H01–H08和八类/J01–J07五层矩阵保留；局部辅助回归不提升为完整宿主/独立QA/发布通过。tests需当前指纹完整复验，失败按新stage/trace返development，exit0才QA。以下旧版本状态仅历史。

权威正文：`.workflow-input/4cd62ced6a4f98f50e3cca6f7596eb06/5f51d35fdc15ed7214749f9413879b4238097f1d6b6bbe2cc0d508d8e103cef7/files/docs/acceptance.md`。
材料archive SHA256：`5f51d35fdc15ed7214749f9413879b4238097f1d6b6bbe2cc0d508d8e103cef7`。来源及故事前置/业务结果见 [需求入口](product-definition.md)。本矩阵不替代冻结正文。

## 产品P0

执行18当前：seq17完整make verify exit2，Key签发201、秘密存在但key-call-http-1280时call-key为空且无调用响应；详见 [宿主失败](validation/attempt-18/host-seq17-summary.md)。key-transfer-r1将editor延迟关闭清理限定在自身字段，保留全局invalidate秘密清理并加入转交epoch约束。四项实际handler回归和145 Node/非监听检查通过；转交值原生门禁是追加，原三视口调用/同ID详情/秘密清理等断言不减。当前源码指纹见attempt18/source-evidence.json，完整宿主/独立QA/发布未通过；全部AC01–AC18/H01–H08和商用八类/J01–J07五层矩阵仍未升级为完整Pass。旧候选局部成功不迁移为当前通过。

执行16当前：seq15对attempt14指纹完整make verify exit2，模型阶段推进但宽泛key-1280超时，不能据旧write_status:201认定Key签发已完成。key-diagnostic-r1仅细分原Key链路等待点并增加数字HTTP状态/安全DOM事实及三项非泄露回归；不改产品、不改原断言或超时、不轮换猜测性修复。141 Node及非监听检查通过，不是当前完整宿主/QA/发布Pass。见 [attempt16](validation/attempt-16/checks.md)、[seq15](validation/attempt-16/host-seq15-summary.md)、attempt16/source-evidence.json。全部AC01–AC18/H01–H08和八类商用功能/J01–J07的五层状态保持；新控制台三视口后续Key/SSE/费用/身份/键盘旅程仍须宿主完整验证，失败按新stage/status/控制状态返回development，exit0才交独立QA。下文“执行14当前”等均为历史候选，不跨版本挪用通过。

执行14当前：宿主seq13对attempt12指纹完整make verify exit2，服务核验成功后停于新控制台model-1280 Error。model-driver-r1仅修已证的实际编辑label/驱动子串匹配歧义，实际字段+驱动片段VM回归先红后绿；增加有界模型阶段/错误分类诊断，原所有native断言不减。138 Node及format/module/vet/build/JS exit0，不是完整宿主Pass；Chromium启动本地实际permission_denied，完整三视口Go UI须宿主。当前来源见 [attempt14检查](validation/attempt-14/checks.md)、[seq13边界](validation/attempt-14/host-seq13-summary.md)及attempt14/source-evidence.json。下列全部AC/H及八类功能/J01–J07五层状态仍未完整验收，旧候选部分成功只属历史；不宣称商用对齐、QA或发布通过。

执行12当前：宿主seq11对attempt10指纹make verify exit2；Go race/build、兼容迁移、正式schema1/2→3三视口及旧native/路由旅程通过，但新控制台service-1280失败，后续未验。已修service-check-r1：服务核验漏传execution_revision，并修正受控上游列表object标记，增加实际render/handler回归和白名单HTTP/核验结果诊断。137 Node、定向race/format/module/vet/build/JS检查exit0；这不是完整宿主通过。见 [attempt12检查](validation/attempt-12/checks.md)、[seq11边界](validation/attempt-12/host-seq11-summary.md)及attempt12/source-evidence.json。完整AC/H及八类功能/J01–J07五层状态不减少、不升为完整Pass；以下旧候选指纹和部分通过仅属历史。

执行10当前状态：宿主seq9对attempt09指纹完整make verify exit2，schema2兼容迁移通过但正式迁移控制台停在keys-1280；不是原生环境限制。migration-driver-r3修正确定的data/items包络缺陷并补actual-handler/SQLite和8个Node契约回归、安全分段诊断，135 Node及定向race/format/module/vet/build/JS检查exit0；实际native复验仍未通过，不能认定唯一原因。见 [attempt10检查](validation/attempt-10/checks.md)及 [seq9失败边界](validation/attempt-10/host-seq09-summary.md)。当前指纹取attempt10；下列旧状态保留历史，AC/H和商用功能/J01–J07五层状态不提升为Pass，完整宿主/独立QA/发布均待验。

执行8状态：tests执行7在attempt08旧指纹完整make verify的Go race包/build通过，但迁移浏览器期待2与生产CLI目标3不一致exit2，后续UI未到达。migration-driver-r2保留原schema2守恒/回滚并追加生产schema1/2→3 native旅程；attempt09实际双CLI/SQLite守恒定向race与127 Node/format/module/vet/build/JS检查exit0。当前完整宿主门禁、逐AC独立QA及发布仍未通过。本矩阵及商用功能/旅程五层状态不提升为Pass，seq7部分成功不可跨版本挪用。详见 [本轮证据](validation/attempt-09/checks.md)与 [旧失败归因](validation/attempt-09/host-seq07-summary.md)。

研发更新：implementation-r1，2026-10-08。下表“未验证”特指完整AC验收尚未通过，并不表示没有实现。逐行当前源码/用例及已跑与未跑的层级见 [研发验证记录](validation/report.md#逐项实现及待验映射)。补充非监听检查通过；固定make verify首次exit2受沙箱bind限制，当前候选完整宿主门禁/真实浏览器/独立QA待验。需求评审及定向回归不代表产品Pass。视觉与交互另由设计负责人对照A原型。不得用静态根浏览器检查替代运行中的Go服务。

| ID | 故事/旅程 | 必须可观察的正常与失败结果 | 验证方法及必留证据 | 当前状态 |
|---|---|---|---|---|
| AC01 | US07/J01–J07 | 多对象A工作台，侧栏/表格/抽屉/表单完整；1280/1440/390及键盘可操作，无整页横向溢出 | 真Go服务UI旅程、逐页原型对照及视口/焦点记录 | 未验证 |
| AC02 | US01/J01 | 服务与模型正式接入；公共/远端名、来源能力价格明确；保存不等于核验；改配置失效旧核验 | UI保存/连接失败修正/生成核验与HTTP记录；区分受控及供应商 | 未验证 |
| AC03 | US01/03/04/J01/02 | 新模型/组合/身份无隐式授权；租户与Key取交集、空授权拒绝 | 模型/映射变更与授权正反例，拒绝零dispatch计数 | 未验证 |
| AC04 | US02/J01/05 | 真实邀请/接受登录、诊断分配收回；移除/降权收身份不盲撤Key；最后管理员保护；跨租户读写导出拒绝 | 多身份真实UI/API正反例，篡改对象/范围及降权旧会话验证 | 未验证 |
| AC05 | US02/05/J05 | 暂停全部Key普通/SSE新调用，另一租户独立；恢复按现规则，历史和额度保持 | 多Key/租户HTTP与UI暂停恢复，dispatch及前后事实对照 | 未验证 |
| AC06 | US03/J02 | 受限Key；完整秘密首次一次；列表/刷新/关闭/重启不可再取回 | UI旅程、安全存储及管理API响应脱敏检查 | 未验证 |
| AC07 | US04/J03 | 立即/最长24h过渡，截止前/等/后，提前撤销、不叠加，共享池历史 | 可控边界时间的HTTP试验及元数据/使用量连续证据 | 未验证 |
| AC08 | US03/04/J03 | 写入后丢响应显未知，双提交不重签，先读回安全元数据；读回失败不伪成功 | 实际丢响应故障注入、双提交/读回失败及对象唯一性UI/API证据 | 未验证 |
| AC09 | US05/J04 | 次数/RPM/TPM/C/预算独立；租户Key同时生效；零/无限；跨Key与并发无部分扣计 | 并发真实HTTP、跨租户/Key、各独立限额与拒绝无dispatch证据 | 未验证 |
| AC10 | US05/J04 | 硬预算受理前原子预留、实际结算；未知保守预留、取消不绕过上限；软预算告警继续 | 有输出边界的并发预算试验，预留/已结算/待确认守恒及失败记录 | 未验证 |
| AC11 | US04/05/J03/04 | 改限额/轮换/跨窗/重启不清事实；时间回拨和重复/迟到终态不再扣计或释放 | HTTP+真实进程重启，边界时钟/重复通知及前后持久事实对照 | 未验证 |
| AC12 | US06/J06 | 成本与租户费独立、USD/输入输出单位明确；受理快照、微美元舍入，改价不回算 | 整数精度及舍入边界、改价前后请求/明细/导出核对 | 未验证 |
| AC13 | US06/J06 | 同ID失败已知→未知→成功；所有dispatch已知计双费用；未知非零；迟到/重复只幂等差额且终态保持 | 受控真实HTTP/SSE多尝试、核对入口、快照与尝试/请求/账单守恒 | 未验证 |
| AC14 | US06/J06 | 明细汇总可核对；当前周期/待确认拒结账；平台管理员关闭、更正幂等，原金额原因调整周期保留，不能负数、不泄租户 | 角色正反例、重复结账/更正、导出精确金额、原记录不可覆写 | 未验证 |
| AC15 | US07/J07 | 搜索筛选分页详情导出关联同ID/归属/尝试/快照；迟到scope隔离，读失败不是空 | 多页多范围真实数据UI/API，乱序响应/失败/重试证据 | 未验证 |
| AC16 | US03/07/J01/02 | 普通/SSE真实输出、取消；输出后不重放；双击一次；取消不宣称上游免收费 | HTTP/SSE时间线、上游实际尝试/取消计数、同ID终态费用UI旅程 | 未验证 |
| AC17 | 全故事 | 服务端身份范围；不泄秘密/完整对话；敏感管理操作审计，降权/切租户/退出清敏感内容 | UI/API/导出/日志安全检查，拒绝与管理事件审计，秘密证据脱敏 | 未验证 |
| AC18 | US08/首次可用及运营 | 正式安装/升级/备份恢复，旧身份/额度/价格/账单历史守恒；迁移失败安全恢复 | 隔离临时目录/端口，旧版本fixture迁移与失败恢复，Go入口及文档 | 未验证 |

## 商用控制台功能及旅程对照

本节承接用户补充的商用可用目标；**不是已对齐New API/LiteLLM或已达到商用水平的结论**。验收仍以design-v0.2.0和architecture-r1为准，冻结材料未改写。只研究官方公开产品文档，不读取本机New API源码、数据、资产或秘密，也不复制实现。正式A十页、服务端真实事实及完整AC优先；新增参考的高级功能不自动成为本Run门禁。

来源为本Run已核对摘要的材料files/docs/references.md（完整根路径见本页权威正文的files根）。本轮另实际打开以下官方页面，观察范围有界，不声称完整审计两产品：

- R1：New API Channel Management，`https://docs.newapi.pro/en/docs/guide/feature-guide/admin/channel`；参照服务配置、模型映射、优先级/权重、逐渠道测试及列表状态。页面还描述批量操作和多凭据轮询，仅记后续差距。
- R2：LiteLLM Admin UI，`https://docs.litellm.ai/docs/proxy/ui`；参照统一模型/Key/费用入口、数据库人员身份和邀请后的实际登录，不采用其环境主密钥登录约定替代本项目人员身份。
- R3：LiteLLM Virtual Keys，`https://docs.litellm.ai/docs/proxy/virtual_keys`；参照Key模型范围、用户/团队关联、期限及预算/RPM/TPM/并发上界。其归属、继承和自动轮换机制不替代本项目已冻结的授权交集、双scope、稳定代际共享池规则。

“已实现”仅指当前候选源码有受支持入口；“局部测试”仅指命名命令实测覆盖；“宿主完整测试”必须当前指纹make verify完整成功；“独立QA”必须独立报告逐AC裁决；“已发布”必须PR/merge/Actions/Deployment/health版本证据。五层独立，不能由前一层推导后一层。当前后面三层均尚无通过事实。

| 功能与参考 | 冻结范围及已实现候选 | 局部测试 | 宿主完整测试 | 独立QA | 已发布 | 具体缺口 / 未验证事项 |
|---|---|---|---|---|---|---|
| 服务/渠道接入与核验（R1） | AC02；供应商/已有自托管类型、地址/秘密、编辑启停、当前配置核验及详情；保存与核验区分 | attempt12实际render精确execution_revision及冲突回归、schema3登录/创建/核验handler/SQLite定向race通过；隔离transport非网络证据 | seq13旧指纹服务1280核验200/success/applicable/版本匹配，完整门禁仍exit2；当前待完整复验 | 未执行 | 否 | 真实供应商无凭据Not Run；其余视口、失败修正与版本失效仍待完整HTTP/浏览器；非泛供应商协议适配 |
| 模型发布/授权（R1/R2/R3） | AC02–03；公共/远端映射、候选路由、优先级/权重、价格及可信上界、显式租户/Key交集 | 既有IdentityAndKeys/PeopleScope/Governance局部；attempt14实际modelEditor/驱动字段契约先红后绿，精确选择不改附加路由；VM非native证据 | seq13旧指纹model-1280 Error/exit2；model-driver-r1当前待完整复验 | 未执行 | 否 | 新精确locator实测、模型/授权后续旅程及禁止隐式授权/空授权零dispatch待宿主；仅Chat Completions普通/SSE主链 |
| Key创建/轮换/撤销（R3） | AC06–08/11；期限/启停、一次秘密、立即/1h/24h、提前退役、不可叠加、共享池及未知写读回 | IdentityAndKeys/ExactDeadline/Accounting局部通过；新增元数据详情回归 | 待执行 | 未执行 | 否 | 实际秘密清理/clipboard/丢响应/双提交及截止边界真实HTTP待验；不是自动定时轮换或兼容LiteLLM Key API |
| 真实成员身份与租户隔离（R2/R3） | AC04–05/17；本地人员身份、邀请接受、角色/scope、诊断分配、移除/降权、最后管理员及租户暂停恢复 | PeopleScope/InvitationRateLimit局部通过 | 待执行 | 未执行 | 否 | 实际多人浏览器、降权旧session、全Key普通/SSE暂停及跨租户出口仍待验；邀请人工交付非邮件发送 |
| 额度与预算（R3） | AC09–11；次数/RPM/TPM/C/硬软费用独立、租户及稳定Key同时准入；未知carry、改限额不清事实 | ResourceDimensions/Governance/Accounting局部通过 | 待执行 | 未执行 | 否 | 实际并发、cancel/kill/restart、迟到和时间回拨待验；无可信收费Token上界不得宣称硬保证；Key“额度/用量”真实入口已补，native待验 |
| 定价及内部计费（R2/R3，冻结规则优先） | AC12–14/18；USD微美元独立成本/租户费、每请求快照、全dispatch已知计费、未知待确认、幂等核对、结束周期关闭及追加更正 | Precision/Governance/Migration补充通过；历史风险已补回归 | 待执行 | 未执行 | 否 | 原金额/调整周期/明细/导出独立对账及跨租户角色待验；不是充值/支付/税务发票或供应商结算认证 |
| 用量日志及诊断（R2，R1列表思想） | AC13/15–17；Request ID多尝试、筛选分页、费用/Token/耗时、追加核对/价格依据、CSV及审计；已确认输入/输出Token汇总从账本读取 | Accounting局部新增按请求筛选Token汇总；详情函数补充回归 | 待执行 | 未执行 | 否 | 有界导出、未知口径、迟到scope及真实SSE旅程仍待验；默认不存完整对话，不以外部平台日志替代本项目事实 |
| 表格/筛选/编辑/详情/移动端（R1/R2） | AC01/15/17；A侧栏十页、局部表格滚动、居中编辑/右抽屉、分页、本地当前列表搜索及日志服务端筛选，秘密不持久化 | attempt12 JS语法、137项VM/辅助用例通过；当前完整真实渲染与原型对照待验 | 待完整复验 | 未执行 | 否 | 1280/1440/390实际可触达、焦点返回/Tab/Escape、细字段/信息密度、表单失败保留草稿待QA；本地搜索不是全数据库查询 |

对应源码/用例完整路径仍以 [验证报告](validation/report.md#逐项实现及待验映射)为唯一证据索引，不复制另一套验收正文。上述“局部”最终命令结果以attempt-08回执为准，较早attempt03/04不能覆盖后来改动。

### 关键用户旅程与未验证边界

| 旅程 | 必须连贯的本期结果 | 候选实现 / 局部证据 | 必留完整证据及缺口 |
|---|---|---|---|
| J01 首次可用 | 管理登录→服务核验→模型/价/租户授权→租户预算/成员/Key→调用→同ID详情 | Go管理/代理入口及console.mjs已编排；局部身份/计费/详情摘要回归 | 宿主真实HTTP+native连贯成功/连接失败修正；真实供应商Not Run不冒充已核验 |
| J02 租户应用接入 | 本租户目录→一次秘密→调用→收窄授权拒绝→修正规则→显式新调用 | 授权交集、Key编辑、playground、日志；IdentityAndKeys/PeopleScope部分断言 | native授权收窄/纠错及拒绝零dispatch；无客户端角色切换；QA待裁决 |
| J03 代际轮换 | 立即/1h/24h→一次新值→旧值截止/提前退役→共享历史与额度 | rotations/retire-previous；精确截止回归通过 | 真实旧/新Key与浏览器秘密不可重取；已过渡不叠加、丢响应安全读回，仍待宿主 |
| J04 预算纠错 | 真实已用接近/耗尽→429零dispatch→提高限额不清零→显式新调用 | 原子admit/manageLimits；ResourceDimensions/Governance局部通过 | 实际dispatch barrier、双scope并发和预留守恒；未知不能当0；UI真实拒绝/修正待验 |
| J05 租户生命周期 | 创建→负责人加入→授权可用→暂停全Key拒绝→恢复；另一租户不受影响 | 邀请/成员/tenant状态及PeopleScope局部通过 | 真实多context登录/角色/成员退出、普通/SSE暂停恢复；Key不盲撤与最后管理员保护 |
| J06 账单责任 | 改价→新旧快照→已知/未知核对→当前/待确认拒关→结束周期关闭→原因/调整期更正→导出隔离 | Governance/Migration/Precision局部通过；价格/账单/历史依据入口 | 完整明细汇总/更正幂等/原金额不改与不同周期出口独立对账；不拿CSV存在代替对账Pass |
| J07 操作与导航 | 详情跳同范围日志→页2→切租户→迟到忽略→错误重试→刷新不恢复秘密 | epoch/AbortController、后端scope、筛选分页；摘要/本地搜索单测 | native乱序响应、失败重试、焦点/可触达及三视口；单测不是真实导航/移动端证据 |

这七条旅程的“完整通过/独立QA/已发布”均未成立。QA须同时使用原冻结interactions.md与本矩阵，不仅按测试脚本的有限动作裁决。源码入口存在但页面缺字段、错状态、不可达操作或权限错误属于**本基线内缺陷**，返回development；不能改名为未来需求豁免。

### 后续商用差距（不扩大本Run）

| 差距 | 当前边界及商用风险 | 后续建议，不作本期已完成事实 |
|---|---|---|
| 多供应商原生协议/多模态/更多端点 | 本期普通/SSE Chat Completions；不能把模型目录名称当可调用能力 | 逐协议定义契约、适配与真实凭据联调，继续明确已支持/不可调用边界 |
| 规模化渠道运营（R1） | 没有证明批量启停/批量核验、单渠道多凭据轮询、自动禁用恢复；大量渠道人工成本高 | 运营量达到需要后单独定义批量失败/审计、熔断恢复及凭据池责任，不照搬参考实现 |
| 企业身份和邀请交付（R2） | 本地人员身份与人工邀请链；未集成企业SSO/SCIM、MFA或邮件配送 | 商用外部身份需求另立基线、权限及凭据授权；不能以本地角色测试声称企业集成 |
| 自动凭据轮换（R3） | 本期人工触发代际及可控过渡，无自动轮换调度/安全分发 | 先定义消费者更新/失效补偿/秘密分发，再实施；不能改变当前共享池与一次秘密规则 |
| 对外售卖/财务合规 | 内部USD测试分账；无到账/退款/外部支付、税务发票/多币种正式责任 | 对外销售前另定义支付、退款、供应商对账及税务；不把内部bill当商业发票 |
| 高可用与运营保障 | 本候选Go+本地持久化、正式备份恢复；没有生产负载/SLO、跨实例预算协调/灾备验证事实 | 上商用前定容量、SLO、留存/告警、负载/故障演练和恢复目标；多区域/GPU调度仍本期非目标 |

后续差距不是认定所有商业场景必需，也不是默认授权实施。冻结范围外的重要能力先记录影响，由后续需求/设计明确；当前范围内的真实行为缺陷和未验证不能用这些非目标掩盖。

### 真实卡点与交接要求

固定make verify首次exit2（沙箱bind operation not permitted）未解决于原生环境；不能重跑必败监听或停止活跃原生进程伪造进度。attempt02格式失败已修，attempt05/07新增单测driver失败及attempt06/08复验分别保留；具体原因在validation/report.md，不归入环境限制。上述对照不能替代实现、固定门禁或QA。

tests/QA交接必须携带本页实际路径、validation/report.md、最终source-evidence.json、各失败/复验日志与同材料manifest。下游先核对指纹及未验证栏；完整make verify、实际UI/网络/重启、独立QA、发布、原生模型核对和真实供应商事实仍各按其责任层取真实回执。

## 材料与Harness（继续）

不将prepare成功、消息传输或本节点文档当作整体通过。矩阵事实栏目只是当前已知分层证据，最终由独立QA/协调者结合真实回执判定；历史无日志不得推定成功。

| ID | 责任/必须动作 | 当前事实与证据缺口 | 当前状态 |
|---|---|---|---|
| H01 | 入口/prepare：Issue版本化资料与SHA/entry | #25及prepare exit0由会话提供；本节点archive和19文件校验一致；不独立声称远端传输全场景通过 | 部分材料证据，待完整核验 |
| H02 | 平台/协调者：Issue→唯一Run、重复事件不重建 | 当前Issue/Run关联已提供；唯一性、摘要可查询与重复入站仍需真实平台证据 | 待验证 |
| H03 | 各执行节点：真实读取材料并承接规则 | 本节点product-definition列版本/摘要/路径、收费/身份/首次链/范围及缺口；模型实际回执尚缺 | 本节点读取已记录，整体待核验 |
| H04 | 各阶段：同基线正式handoff、安全失败恢复 | 将交接同根manifest/PRD/AC/交互；后续缺文件/错摘要/过期版本失败及恢复需保留证据 | 待验证 |
| H05 | 研发/设计：执行Agent源码、UI及测试符合基线 | implementation-r1为当前未提交候选；真实Go后台/工作台及回归已实现，源码指纹及补充检查在validation；完整固定测试及原型实际对照待宿主/QA，协调者不得代写产品 | 已实施，完整待验证 |
| H06 | 独立QA/项目测试Connector：逐项对照、缺陷返回复验 | 要求回执绑定摘要、提交与工作树指纹，保留失败历史；尚无本轮QA | 未验证 |
| H07 | 发布Connector/协调者：同产品PR关联#25、合并部署版本一致 | head/base/merge、Actions/Deployment、health版本和固定效果URL均待实际发布；不拿8793编排服务代替产品 | 未验证 |
| H08 | 平台维护者/协调者：缺陷修唯一源、正式安装升级及原入口复验 | 仅真实缺陷触发修复，不制造缺陷；若未触发仍记录当前版本原入口/升级边界，不无证据判Pass或擅自删此项 | 未验证 |

附件负向链还需平台证据：路径错误、下载失败、损坏、路径穿越、超大包、未知版本、权限不足、重复入站；包内脚本不能执行或扩大权限。H04/H08的项目证据不能替代这些平台责任。

## 裁决与证据契约

- 产品固定入口唯一为 `make verify`：格式、Go modules、vet、Go测试及真实HTTP集成、构建，扩展到本基线身份/预算/费用/迁移恢复及1280/1440/390实际UI旅程。设计/研发不能沿用旧入口exit0当作新增范围覆盖。
- QA报告和证据保存本Run文档根，记录实际代码版本、工作树指纹、材料SHA、命令/回执、逐项结果和失败复验路径。各层使用实际handoff路径，不猜固定文件名。
- 保留每次失败和沙箱限制；静态截图不替代UI操作，单测不替代完整HTTP/浏览器/恢复，受控HTTP/SSE不替代供应商事实。
- 真实供应商、外部身份与支付缺凭据：明确Not Run与影响，不拦截可完成内部认证/受控产品验收，也不冒充外部Pass。H事实缺证保持待验证，不能将局部产品通过称完整退出。
- 最终Go需全体当前AC/H相应层级真实证据；需求节点无实施/QA/合并/部署通过声明。
# 执行23当前整改映射（continuous-journeys-r1）

保留以下AC01–AC18/H01–H08和商用对照全文；独立QA22报告不改。QA22 No-Go仍有效。当前逐层状态见 [attempt23](validation/attempt-23/checks.md)，源码绑定source-evidence.json：

| 当前范围 | 实现/测试资产 | 本地局部 | 当前宿主完整 | 独立QA | 已发布 |
|---|---|---|---|---|---|
| Key状态、授权及费用用量；AC01/06/07；QA22-02 | 实际formatter与限额读回列；journeys.test及真实UI状态旅程 | Node契约；非原生证据 | Pending | QA22 Open，待复验 | Not Run |
| 服务/模型/首次负责人链路；AC02–04/06/16；J01/J02 | 空库配置、保存失败修正、核验、价/授权/预算、真实角色邀请/签发/收窄恢复 | 驱动静态契约及原后台定向回归 | Pending | QA22-01 Open | Not Run |
| 轮换/硬费用/两租户；AC05/07/09–11；J03/J04/J05 | 1h/24h共池/禁止叠加/提前结束/状态；费用429及提高；两Key暂停/另一租户/恢复/成员权限收回 | 原后台定向回归；新UI Not Run | Pending | QA22-01 Open | Not Run |
| 价快照/核对/结账/更正；AC12–14；J06 | 旧新请求、未知核对、拒绝/合法结账、原金额/调整周期/隔离CSV | 原后台定向回归；新UI Not Run | Pending | QA22-01 Open | Not Run |
| 日志/分页/迟到/错误恢复/安全刷新及键盘；AC01/15/17；J07 | UI生成>25记录、同范围详情跳日志、页2/筛选/导出、真实响应延迟与读取失败、键盘校验修正提交 | supplemental Node空态/驱动契约；新UI Not Run | Pending | QA22-01 Open | Not Run |

QA22整改是覆盖和状态候选，不是新增范围。外部真实供应商/支付/身份、H缺证及正式发布继续Not Run。固定原门禁、attempt01–20及QA22全部原失败/通过事实保留。
