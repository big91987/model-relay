# seq14 原生轨迹与08驱动修订

2026-10-06；负责人：研发节点；状态：Ready for Review / Pending Host，非Done或Accepted。

## 真实失败与已证边界

固定命令 `make verify`，平台tests预算900秒，seq14实际exit2；版本06摘要 `28e1a5652d32410125bd4df1c74c7b0b552b091d802978f2de6094c43c994752`、81文件，HEAD为精确旧main `85f6d17c07c5fddeb2d6a39a5a067769a11bb4cc`。平台保留77652 bytes/truncated=false，本节点从72500读取到77652/eof核对失败末段。历史health预期红灯不是根因；relay包221.715s PASS、health-history分类5.339s PASS、build完成，原双屏控制台与旧UI非空迁移/回退局部PASS，不是整体通过。

实际失败 `candidate primary selection 1280` / AssertionError，models只有GET200、无保存POST。[70677原始完整有界JSONL](host-seq14-selection.jsonl)从实际ignored宿主文件原样归档，10事件/5边界、omittedEvents0，无值/URL/正文/凭据/本机路径。它表明：

- before Home聚焦原select，documentFocused=true、enabled/connected，匹配1、目标index1、options3。
- Home与ArrowDown可信keydown/keyup到达original-select、冒泡prevented=false；ArrowDown后仍index0，无input/change。
- Tab前后仍index0；只发生focusout，原option身份保持、mutationCount0。不是“ArrowDown已移动但Tab撤销”，未见产品重绘/选项替换。

这足以定位该次driver动作没有产生原生选择变化；不能把冒泡prevented=false当成所有系统层机制证明，也不推导Chromium版本通用缺陷。当前安全证据不包含OS原生弹层轨迹，精确底层机制仍未证实。不改产品键盘事件，不把测试夹具行为伪装成浏览器机制。

## 最小修改与不变契约

仅测试driver调整：既有journey.mjs已用native type-ahead→Tab且seq14双屏PASS，复用该操作方式而不导入有启动副作用的旅程模块。新 `typeAheadCandidate(control, keyboard, label, sample)` 负责唯一可输入ASCII前缀校验、键盘输入与边界采样。两受控上游仍从UI创建、名称仍明确“受控”，将A/B及屏宽前置为 `A-<width> 受控` / `B-<width> 受控`。这不是产品命名限制，也不修改用户数据；选择为原生控件输入首个ASCII token再Tab。

先只读验证前缀在全部options中唯一且对应启用的精确目标；非ASCII/歧义/禁用/缺目标在任何键之前失败。不使用selectOption、不赋selectedIndex/value、不派发事件、不API预置、不兜底重试、不加sleep/RAF、不给产品添加特殊分支。原末端精确selected label断言、用户入口操作、候选映射/优先级/3:1外部计数、禁用/有限切换/未知合计/冷却恢复/SSE及取消不重放断言均保留。产品、Go、Makefile、health-history、Workflow、依赖锁不改。

旁观轨迹保留原node/options身份与有界JSONL/cleanup；仅新增printable枚举（不留真实字母）及before/after type-ahead边界。两项新增VM测试只证明helper动作顺序、不程序设值、失败前不发送键；既有轨迹测试补可输入键脱敏。模拟keyboard不会改变selectedIndex，显式断言仍0，故不能被误读为真实原生选择PASS。

## 检查、版本与接力

交接前源码审查发现两屏共享隔离服务，1280创建的对象会保留到390；首轮仅A/B单字母前缀在第二屏会歧义。未等宿主失败，按同一type-ahead策略将屏宽纳入ASCII token，VM基准改为四个名字同时存在并逐一匹配；没有改变原生选择断言或产品。保留本节点中间[07身份](final-source-07.json)、[第一轮VM](candidate-keyboard-vm-01.log)、[第一轮局部检查](candidate-keyboard-validation-01.log)与[第一轮审计](candidate-keyboard-audit-01.log)，它们不是当前08版本通过证据；最终以02检查及08源码身份交接。未覆盖任何宿主失败。

[VM](candidate-keyboard-vm-02.log)116/116 exit0；[语法/format/module/vet/零测试全包编译/build/diff](candidate-keyboard-validation-02.log)exit0。已读取Trellis frontend index/checklist、DOM/lifecycle/verification、共享reuse/cross-layer与交付契约/审查清单；当前规范同步helper签名、错误条件、好/基准/坏例及证据层。跨执行面数据只读options→校验→native keyboard→只读精确断言→脱敏有界文件，秘密未进入诊断。公共D4/D5和总览保持In Review，已有PRD AC-09/10与架构不改，无新任务体系。

当前源码身份由[08](final-source-08.json)绑定；交付回执排除以免自引用。本机监听/Mach限制已在前序失败保存，执行环境未变，不重复完整必然受限入口，不扩大权限。next由宿主900秒执行同一固定make verify并保留完整新日志/exit/版本和实际JSONL。须同时看到type-ahead前后index、input/change、Tab提交及全部后续业务断言成功，才可认定本层修复有效；若仍失败，先看可信printable事件/前后index/身份/变更与API边界，再作最小归因，不轮换按键或削弱断言。

08摘要 `c340664b231d83d153f0d2569065db08c8be34661a314f51e7831f74f808b27e`、82文件。[审计](candidate-keyboard-audit-02.log)真实exit0，相对链接问题0，源码身份复算一致，原门禁/历史Run未改；原轨迹逐字节一致，SHA-256 `df08132d7ca2f3d1b3b6b77dd548d838fcea4b154d86bdbb293eb6dc6ca52d7d`。

独立QA在宿主真实门禁后判断13AC。正式Actions非空升级、获授权可信隔离失败恢复与真实供应商仍Not Run；平台维护源、安装/Workflow/DB未改，无Git发布，不人工预upgrade代替activate。seq9预算中断及seq10/12/14失败与03–06身份全部保留。
