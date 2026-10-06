# 既有交互契约收尾：研发修复与待验对象

## 恢复与边界

2026-10-06，同Issue/Run接续维护者静态对照反馈。权威输入：[交互§1/4/6/7](../../console-interaction.md)、[PRD](../../product-definition.md)、[架构](../../console-architecture.md)、[原QA整改](../../acceptance/remediation-01.md)、[独立QA首轮报告](../../acceptance/report-01.md)。反馈不是新的原生失败回执；本记录是研发静态核对及实现记录，不替代独立QA结论。

HEAD仍 `d62cd51e8de2127548e1ac7469d349443d1bfe3d`。seq27失败对象 `9ebdb69472beb0ffb66565f1ec6d3bd47ea0d76b573fdb6338377ce5cc00a6f6` 保留；[闭合选择修正](host-return-10.md)的 `d9d6abe3f2599bc9dfdeaff4a58f13e86f002779bec4793d2434f76417cf21a8` 尚无新的宿主结果。当前对象包含其选择修正及本次四项收尾，不能复用旧seq18绿色或seq27的1280px部分通过。

写集限原生前端、原项目VM/浏览器测试、Trellis真实规范与本Run交付文档；不改Go业务协议/数据库/需求设计/QA资产/平台/固定门禁/锁文件，无新对象或迁移，不扩展路线2–4。记录定位、普通导航、占位符示例等允许的等价行为不改，不因线框按钮数量或文字差异新增产品范围。

## 四项静态发现与实施

| 既有契约 | 静态依据（非原生回执） | 修复与约束 | AC与测试层 |
|---|---|---|---|
| §4 撤销/恢复 | refresh-upstream只load；上游GET不清输入token；save和restore各有独立button锁，可交叉开始 | 复用“重新读取上游”，增加明确说明；立即清token、GET-only恢复公开地址/启停；失败保公开草稿且说明恢复未确认。upstreamAction持有共享操作身份，保存含其GET读回全程禁用两动作及相关输入；重复submit不写；导航/登出/pagehide/不明会话清忙态，迟到旧finally不得解锁新操作。保留PUT确认与GET失败边界 | AC-07/08/12/15；VM与待宿主GET/PUT屏障 |
| §7 提交参数 | 原提交只改变send/cancel，四输入无disabled边界 | 同步snapshot后锁定model/key/message/SSE原控件，保留实际值（密钥仍密码形态）；当前成功/失败finally解锁，取消/跳页/失效清密解锁，旧finally仅在call与owner同时匹配时操作。无自动重发或协议变化 | AC-10/11/12；VM与待宿主普通/SSE |
| §6 创建时间 | keys GET已有created，renderer忽略字段 | 同一记录行以“创建时间”显示known(value.created)；null/缺失/空为未知，无客户端时钟或新表 | AC-07及交互§6；VM真实/未知与待宿主同ID GET交叉核对 |
| §1 行动作焦点 | 启停后replaceChildren，旧按钮被销毁，无显式身份恢复 | 点击前捕获是否拥有焦点；成功读回按model alias/key ID找同一行新状态动作。仅旧按钮仍持焦点或因删除落body且owner有效时恢复；重排行/同名key不跳对象，用户已移到别处/跳页/对象消失不抢焦点 | AC-03/07；VM焦点模型及待宿主即时原生焦点 |

焦点第4项静态证明缺少恢复路径，VM证明新实现的身份/不抢焦点规则；本次没有真实浏览器复现或截图，不把VM虚拟focus称原生缺陷关闭。恢复读取503注入是明确测试夹具，不是实际存储故障。

## 回归与共享入口

项目 `tests/browser/ui.test.mjs` 新增16项四契约检查，保留77产品VM及8选择helper契约，共101项。覆盖restore立即清密/零PUT/确认恢复/503保公开草稿、save/restore互斥及重复submit、PUT200/GET503、旧restore versus新save、pagehide/不明会话忙态、生成三个终态/旧finally不能解锁新调用、created/未知、两种对象重排行/同名key/焦点移走及跨页。

原 `tests/browser/journey.mjs` 同时扩展1280px/390px：
- 真实GET恢复屏障：公开草稿与凭据经键盘输入；未释放GET前凭据已空，输入/两按钮禁用，重复Enter零PUT；实际GET200后恢复持久公开值。另注入GET503检查失败反馈、保草稿和零写，再真实恢复。
- 真实PUT保存屏障：保存/恢复共享忙态，重复requestSubmit仍一条PUT，释放后继续实际Go写入与读取；已有保存确认/读取失败断言不删。
- 普通请求测试级屏障：释放前实际提交值仍在原控件且四输入禁用，然后继续真实Go调用；实际受控SSE接收时同样检查。成功/取消/供应商失败后可再编辑，秘密清理保持。请求值比较只返回布尔，不打印凭据/消息。
- keys实际GET响应按本条ID核对created行文；启停真实PATCH/GET后，不再按Tab就检查同对象新按钮activeElement/focus-visible。原首次全UI、依赖恢复、精确ID、分页、历史保护/版本/密码/保密门禁保留。

仍是原样 `make verify` 内的frontend-test与browser入口，无API/SQL业务预填，无替代静态browser.check，无新依赖。闭合选择精确断言保留，不能把补充项通过冒作390px选择已通过。

## 原QA脚本兼容性与语义差异

复跑原QA脚本首先发现项目前次选择helper已置于首个密码测试之前，包含import.meta及test注册；原QA按历史fixture前缀new Function读取，导致SyntaxError，退出1。见[原样失败](closure-qa-original-initial.log)。这是项目测试布局对原fixture读取的破坏，不是新产品/平台故障。仅将项目自身module-only helper及注册移动到文件尾部，fixture定义保持前缀位置；没有修改原QA脚本、原始红测或其断言。

原脚本恢复执行后，[最新原样回执](closure-qa-original.log)仍4项中3绿1红、退出1，唯一失败QA-03仍调用普通records导航，而新实现提供overview-record精确动作。保持这一事实；独立QA必须复核普通导航清上下文与明确证据动作的语义等价性。不能把该脚本改成新动作再覆盖旧红测宣称4/4，也不能由研发关闭QA No-Go。

## 当前验证事实

| 层/命令 | 真实结果 | 证据与限制 |
|---|---|---|
| 研发静态源码核对 | 四项缺口确认并修复 | 不是原生失败回执，也不是独立QA验收 |
| 初轮新增VM | 99/99，退出0 | [初轮](closure-vm-attempt-01.log)，未捕获四项实现前红测，不补造 |
| 最终frontend VM/helper | 101/101，退出0 | [最终](closure-vm-final.log)；[中间99项](closure-vm.log)保留 |
| 三项Go in-process race契约 | 退出0，8.560s | [回执](closure-go.log)；非新增HTTP监听/浏览器证明，编译器临时绝对路径归一为temp，warning原义保留 |
| JS语法/格式/模块/vet/build/diff | 退出0 | [分项](checks-closure.log)，固定Makefile与Go/npm锁文件对HEAD无差异 |
| 原独立QA脚本执行 | 3/4，退出1 | 原QA-03语义差异仍红；历史与新回执保留 |
| 新完整make verify / HTTP / 真实Go两屏 | Not Run | 稳定沙箱监听/Chromium限制未变，见[已知真实受限命令](expiry-watchdog.md)，不重复必然受限执行；交宿主原入口 |
| 供应商/PR/合并/部署 | Not Run / 无新增回执 | 受控上游不等于供应商验证；路线2–4未实现 |

最终对象：[source-closure.json](source-closure.json)，56文件、Go1.25.6，摘要 `475091a58b28ca779261cb8c1dd24a55118d9c74df1c4d1be8fb5aa4fc82c63a`，指纹命令退出0。交付收据不参与自身指纹。原失败/旧指纹/QA首轮报告均未覆盖。

交接前复核：6份公共交付文档的100条本地引用、67个app.js直接DOM ID、9份收尾产物本机路径扫描通过；diff退出0，按项目Go缓存重算56文件摘要与上述对象一致。动态DOM绑定及敏感字段行为仍由VM与待执行原生旅程核对，不把静态扫描扩大为保密验收通过。

Trellis before-dev恢复frontend索引、DOM/lifecycle/verification及shared guides，packages发现成功；真实初始化/规范已存在，bootstrap不适用。update-spec维护同一规范的共享忙态、调用锁/焦点/时间和原层验证约束；check核对生命周期、DOM IDs、静态边界、测试和无扩权。公共交付审查Ready for Dispatch至tests，不是G4通过。

T001/T002/T003仍In Review，M01/M02/M03仍Ready for Review，独立QA No-Go不自行关闭。下一步：宿主对新指纹原样 `make verify`，保留真实退出码/脱敏失败/两屏清密截图；失败按正式边返工，通过后独立QA逐条18AC、四项补充及QA-03独立语义评审，再由Connector处理后续发布。不把VM/旧局部通过或静态完成称本切片/产品/平台全量通过。
