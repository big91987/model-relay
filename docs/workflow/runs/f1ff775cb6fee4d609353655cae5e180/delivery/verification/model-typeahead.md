# seq31 原生模型选择证据与限定返工

2026-10-06；负责人研发；T003 In Review，M03 Ready for Review。交付审查 Ready for Dispatch 至宿主 tests，不是完整验证或 G4 通过。沿用本 Run PRD、交互、架构与独立 QA 整改包，不重建需求/设计，不关闭 QA No-Go。

## 实测对象和诊断恢复

平台 seq31 原样 `make verify` 退出2，对象56文件摘要 `11c3732c72f61b19512bcf83523c87d2c5dd35bc4b971016b370177bdc9b0067`，HEAD `d62cd51e8de2127548e1ac7469d349443d1bfe3d`。当前 Go race/HTTP、build、1280px版本及QA恢复/精查/分页完整旅程通过；390px版本通过，但模型选择失败。旧历史health红测是隔离重建预期失败，不是本次失败源。

从忽略的宿主JSONL中按 execution=96972 和失败回执精确行摘要恢复：

| 归档证据 | 字节（含LF） | 行SHA-256 |
| --- | --- | --- |
| [390px失败完整有界行](seq31-model-390.jsonl) | 21986 | `92132f4369afafdd97177e5f9fde9b83e7ced900add813d9a9c9d924daba5515` |
| [1280px第一次成功行](seq31-model-1280.jsonl) | 18789 | `b2fcd5f4ddaefd17e735a4a8d94ffa88dd317e8d3fc74f8d5157cc8bd7887cd7` |

仅归档相关行，原忽略目录全部追加记录未覆盖。归档只有既有固定标签、计数、布尔、索引，不含alias值、凭据、消息、URL或响应体。1280px首次目标index0最终仍0，不能证明方向键移动成功。

390px逐键：Tab已到达test-model；四次ArrowUp与一次ArrowDown的keydown/keyup均送至select，documentFocused=true，控件启用，node/options身份不变，optionRenders/replacements=0。索引始终0，目标1。navigation=12、load=13不变，models 14/14、overview 17/17，pending均0，loading=false。Tab后正常离开。此次观察不支持迟到读取、重绘重置或按键前失焦；没有观察到“曾选中1又重置”。它确定旧Harness闭合箭头序列未执行选择意图，但不确定所有Chromium/OS的底层机制。此前要求先收敛真实边界已经取得证据，不再盲目增加相同诊断或改产品CSS。

## 最小修改边界

行为缺口位于 `tests/browser/journey.mjs` 的 `keyboardModel(page, alias)`：对就绪原生select发箭头并不等于选择了目标。只修改该helper及 `ui.test.mjs` 契约、适用Trellis验证规范和公共T003/M03/计划。业务实现、后端、鉴权、schema、产品焦点策略、需求/交互/架构、QA资产、固定Makefile与锁文件不改。

推荐并实现单一原生type-ahead路径：真实Tab到达并验证activeElement/focus-visible/outline后，`page.keyboard.type(alias)`输入受控ASCII别名，再Tab离开；仍强制最终index等于目标、完整inputValue相等。不是失败后补偿或循环尝试，不调用selectOption、DOM赋值、insertText、程序化focus、click，也不以API/SQL预填或只用首选项替代390px非首项。是否在宿主Chromium实际奏效仍须新对象验证，不能由double推断OS行为。

诊断保留原边界能力，新增before/after type-ahead白名单阶段；仅select的单字符事件标记为printable，不输出字符，其他输入字符不采集。8192字节失败整行预算保留新阶段；原箭头阶段压缩测试仍保留。JSONL继续只读观测与finally清理，成功/失败均追加并报告精确行摘要。

## 验证与状态

- [修复前契约](model-typeahead-before.log)：退出1，107项中旧箭头helper不满足新原生文字选择序列；不是新浏览器复现。日志本机工作区路径已替换为占位符。
- [初次修复后](model-typeahead-vm.log)：107/107，退出0；[最终专项](model-typeahead-vm-final.log)：108/108，退出0。包括立即/离焦提交、同前缀非首项、末项/停用项、缺失目标、输入无效果必须失败且保留精确诊断、printable不泄露及容量压缩。模拟只验证项目helper序列和门禁，不验证浏览器type-ahead机制。
- [分项检查](checks-model-typeahead.log)：JS语法、Go格式、模块校验、vet、build、diff及Makefile/锁文件不变检查全部退出0。
- [当前对象](source-model-typeahead.json)：HEAD不变，56文件摘要 `1b2f7026692d0633f6f050d585cec0bb67571a7ecf411848d8213d2fd609cdc6`。交付回执不参与自身指纹。
- 新对象完整 `make verify`、真实Go两屏type-ahead：Not Run。稳定沙箱监听/Chromium限制未变，沿用[首次真实限制](expiry-watchdog.md)，不重复已确认必然受限命令；交宿主固定入口原样执行。不把旧对象1280px绿色或108项double替代新两屏通过。

Trellis before-dev已恢复实际frontend索引/DOM生命周期/验证规范/共享guide及任务基线；初始化和真实规范存在，bootstrap不适用。check核对同层复用、保密与固定门禁；update-spec原位记录这次实测边界和选择契约。公共交付审查复用原T003/M03，不增平行任务、不扩权限、不改平台。

关键路径：新对象宿主 `make verify` → 如失败保留实际退出码与新execution精确JSONL行按正常边返工 → 如通过交独立QA逐条18AC/四项补充/原QA整改复验。QA-03泛records驱动与明确overview-record动作差异继续独立评审，研发不修改红测或宣布关闭。无需新增用户阶段批准。供应商Not Run；公共路线2–4未实现；本节点无PR、合并、部署回执，不宣称产品或平台全量通过。
