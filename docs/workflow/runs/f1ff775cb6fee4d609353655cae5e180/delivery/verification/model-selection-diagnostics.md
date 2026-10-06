# seq29 后原生选择诊断收敛

## 最新回执容量补充（2026-10-06）

用户核验平台只保留命令末尾12 KiB，40条完整trace静态估算约13.1 KiB且未含外层。本次只修项目输出载体，不改平台上限、产品或按键策略。`serializeFailure`对最终FAIL前缀＋JSON＋换行计算UTF-8总字节，预算8192。重复显式阶段按各阶段第一/最后一次选取并保持时序；到达、两类方向键前后、Tab前后及commit不因尾部切片丢失。保留boundaryCount/samples及omittedBoundaries/omittedEvents；超预算时先去事件明细、请求摘要和辅助探针详情，仍保留关键边界，必要时再去重复阶段并更新省略数。不截字节流或无效JSON。首次方向键没有发生时不会伪造其阶段。

完整有界白名单JSONL仍追加到忽略的test-results，失败回执提供相对文件名、execution、bytes和sha256。SHA-256对应单条UTF-8 JSONL行（包含结尾换行），不是不断追加的整文件。宿主下轮需取匹配execution且行摘要一致的记录，归档到本Run再分析；本沙箱没有运行页面，因此实际JSONL文件尚未生成，不列为已交付产物。秘密/模型实际值/消息/任意DOM/原始错误不进入轨迹。输出长度测试不是原生选择失败根因证据。

新增容量测试使用24个显式边界、40个事件源、12条响应、多字节中文checkpoint及超大辅助探针，断言完整可解析整行≤8192字节、关键阶段首末都在、省略计数准确及无私密字段；helper验证行摘要/字节数。第一次107/107见[容量契约](model-summary-budget-vm.log)，补齐24边界与摘要校验后107/107见[最终契约](model-summary-budget-vm-final.log)。语法/format/module/vet/build/diff检查见[分项回执](checks-model-summary-budget.log)。按before-dev读取真实规范，check核对数据链：只读观测→白名单投影→阶段摘要→总字节预算→整行输出；update-spec维护验证规范，已有真实Trellis无需bootstrap。

补充防止写诊断失败时引用前次行：每次keyboardModel开头清空文件回执指针，当前行成功追加后才给摘要。闭合107/107见[最终闭合契约](model-summary-budget-vm-closure.log)。本次最终对象见[source-model-summary-budget-final.json](source-model-summary-budget-final.json)，56文件摘要 `11c3732c72f61b19512bcf83523c87d2c5dd35bc4b971016b370177bdc9b0067`；指针补充前的[source-model-summary-budget.json](source-model-summary-budget.json)独立保留。旧[source-model-boundaries.json](source-model-boundaries.json)及历次失败保留，不适用于新对象。稳定沙箱监听/Chromium限制未改变，不重复必然受限完整make verify；新原生两屏/完整门禁仍Not Run留宿主原入口。QA No-Go不改变，任务In Review、里程碑Ready for Review；无供应商、PR、合并、部署或产品全量通过声明。

## 当前委托与版本边界

本Issue/Run正常返工，按维护者最新反馈先取得原生边界证据再选择按键策略。需求/设计/QA资产和产品实现不改，不增建Run、不改平台/运行数据库/固定门禁。公共任务仍T001/T002/T003 In Review，M01/M02/M03 Ready for Review，独立QA未放行。

实际已失败对象是seq29的56文件摘要 `475091a58b28ca779261cb8c1dd24a55118d9c74df1c4d1be8fb5aa4fc82c63a`、HEAD `d62cd51e8de2127548e1ac7469d349443d1bfe3d`，宿主原样make verify退出2。390px探针targetIndex=1、enabledTargetIndex=1、selectedIndex=0、targetEnabled=true；1280px主目标为首项，其通过不能证明箭头改变控件。seq25/27/29原生失败分别保留在既有报告，不将helper人为实现的按键语义当作根因证据。

[上一次未交接建议](host-return-11.md)的类型前缀方案现已撤回：平台交接未接受，期间没有宿主运行该对象。保留其日志及[source-model-typeahead.json](source-model-typeahead.json)作为历史，不称其已验证。本轮恢复seq29的ArrowUp／ArrowDown／Tab顺序及初始选项次数，不再替换输入策略或改产品焦点。新对象指纹见[source-model-boundaries.json](source-model-boundaries.json)，后续只能关联新对象的真实宿主回执。

## 源码依据与待证假设

`internal/relay/web/app.js` 的 `navigate` 增加navigationVersion、显示page-loading，await load后仅当前owner在finally隐藏loading、显示目标panel并focus其h1；`load('test')` await真实models/overview，重建选项，并按selectedModel恢复公开选择。导航按钮onclick调用navigate，另外存在hashchange监听。本轮静态查见这些路径，但没有新的原生证据证明哪个路径竞争或重置选择，不以“代码存在”认定产品根因。未加入人工load完成等待来遮盖竞争，未改标题焦点、hash行为或选项重绘。

诊断仍嵌入原 `tests/browser/journey.mjs`，由原 `make verify` 的browser入口运行实际Go及两屏。每次keyboardModel在键盘到达前开始只读观察：

| 边界／事件 | 允许记录的事实 |
|---|---|
| 到达前后、每次ArrowUp/Down前后、Tab前后、下一动画帧、cleanup前 | 同一个select／document是否聚焦、焦点名枚举、原节点仍连接／同身份、disabled／optionCount／selectedIndex |
| keydown/keyup、focusin/out | 仅ArrowUp/ArrowDown/Tab与select／other目标枚举及焦点快照；无任意按键、文本或DOM |
| options／loading／panel有关mutation | 原option节点身份、重绘次数、select替换次数；不读取选项内容或序列化节点 |
| 当前页面／导航变化 | navigationVersion／loadSequence数字、test view/hash布尔、loading／panel状态；hashchange/popstate固定事件标签 |
| models／overview实际GET | request／requestfinished／requestfailed计数、pending响应体数；不读取body/header/token/完整URL |

`readsSettled`只表示两类读取均被观测、当前都无pending响应体且当前test panel已结束loading；不是读取成功证明，失败数另列。它不独立识别管理API owner token，需要结合当前navigation/load数字和时间线判断本次导航。requestfinished是网络响应体完成观测，page-loading/panel是应用load/finally观测；两者不能互相替代。

显式边界最多24条、事件快照最多40条，保留最初8条与最新尾部，并记录总次数／截断布尔。所有输出仅数字、布尔和代码固定标签，focusName为test-model/test-heading/nav-test/other/none；不读密钥、消息、任意DOM内容或原始错误进入探针。Observer/listeners/测试全局在finally移除，不修改焦点、选择、网络结果、产品state或数据。

成功和失败都追加到忽略目录 `test-results/model-selection-1280.jsonl`、`test-results/model-selection-390.jsonl`，每行含当前execution PID／viewport数字，stdout输出同一model_diagnostic_execution计数便于关联当次记录。失败尾部仍带modelProbe，但压缩为白名单显式边界及最后8条事件快照，完整有界轨迹保留JSONL，避免重复快照撑满脱敏失败尾部；成功首项轨迹也保存，不再仅靠首项最终值推断方向键有效。宿主／QA归档当前execution行到本Run证据目录，旧追加行不能冒作新对象。当前沙箱未运行原生页面，因此这些JSONL尚未产生，不列入本阶段实际artifacts。

## 验证事实与局限

- [初次诊断契约](model-boundaries-vm-initial.log)：102/102退出0。
- [观测投影契约](model-boundaries-vm.log)、[清理与读取判定](model-boundaries-vm-final.log)、[最终导航事件契约](model-boundaries-vm-navigation.log)：105/105退出0。执行实际项目helper及采样函数的double，覆盖逐键边界、原始精确断言不变、忽略箭头仍失败并留证、选项重建／节点替换／标题焦点／navigation/load变化、pending与读取失败、白名单无值、40条截断与清理。只验证观测逻辑，不复现OS键盘机制或证明根因。
- [当前失败投影契约](model-boundaries-vm-projection.log)：106/106退出0，追加真实projection函数丢弃任意字段及压缩尾部的契约。上述105项与[初次指纹](source-model-boundaries-initial.json)保留，不覆盖为当前对象；最终分项见[最新检查](checks-model-boundaries-final.log)。
- 最终分项语法/格式/模块/vet/build/diff及门禁锁文件检查全部退出0；5份公共文档84条本地引用、10份新证据路径检查通过。指纹命令退出0，当前56文件摘要 `7479105426d1dd8ee6132d018665ce6a36445703f09a9912f9c5f6b132fc45a0`，HEAD仍为上述基线；交付记录不参与自身指纹。
- [分项检查](checks-model-boundaries.log)：记录实际命令/退出码、公共引用与新证据本机路径扫描；原Makefile及锁文件差异检查。当前不改后端，无需为此重跑已通过的无关业务回归，但后续固定make verify仍完整保留Go/race/HTTP/build/两屏及历史门禁。
- 新原生逐边界诊断、完整make verify：Not Run。稳定沙箱监听/Chromium限制未变，依据[实际受限证据](expiry-watchdog.md)不重复必然受限命令；下一节点宿主原入口收集。

采样增加page.evaluate与一个动画帧边界，可能暴露或减少竞争；即使新对象绿色，也不能因此宣称某假设确证或产品问题已修复。不能把观测契约绿色替代真实两屏、当前对象版本一致性或独立QA。

## 公共交接与后续判定

Trellis before-dev恢复适用frontend规范与共享guide，初始化/真实规范存在，bootstrap不适用；check复核单层观测无扩权/无敏感投影/无门禁删除，update-spec原位维护诊断契约。交付审查仅Ready for Dispatch至tests，不是G4通过；不需要新需求/设计或用户再批准。

关键路径：新指纹原样make verify → 真实失败／成功边界归档 → 按证据决定研发根因修复或独立QA原层复核。若一次按键已改变索引随后重置，依据节点/option身份、navigation/load变化与标题focus轨迹定位实际重置路径；若按键前或事件时焦点／document focus／load状态不正确，先处理实际焦点／就绪路径；若在同一就绪控件收到键但值未变，保留原生机制未知，不从double推断OS结论。诊断未收敛前不再换键、不用selectOption/DOM赋值/静态browser.check、不删390px或后续同请求路由断言。

QA首轮No-Go、QA-03语义差异及所有历史失败保留，独立QA不由研发关闭。供应商Not Run，路线2–4未实现，无新PR/合并/部署或产品/平台全量通过声明。
