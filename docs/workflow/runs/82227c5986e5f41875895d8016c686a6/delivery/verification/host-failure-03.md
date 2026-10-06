# 宿主05失败：原生候选选择归因诊断

2026-10-06；D4/D5，AC-02/03/12；In Review，完整门禁 Failed，未验收。

## 实际证据与已证边界

seq12：900秒预算、固定 `make verify`，真实exit2。源码05摘要 `f251735964e9fcc330e4f70387e3a3b073fd63537f3960e6cbd8a4daca560510`，HEAD `85f6d17c07c5fddeb2d6a39a5a067769a11bb4cc`，Go1.25.6。平台保留74970 bytes、truncated=false；本节点按71000读取至74970/eof核对末段，完整日志仍由平台按seq12保留。seq9预算中断与seq10失败、05原指纹均不覆盖。

宿主relay包PASS（220.531s）、health-history分类PASS、build完成，浏览器实际旧main UI非空迁移/回退及原1280/390旅程PASS。这些是05局部实测，不是整体验收；历史旧health预期红灯不是本次根因。实际失败为 `candidate primary selection 1280`、AssertionError，modelResponses只有GET200，无模型POST。安全诊断原文：

```json
{"boundary":"after Tab","matches":1,"targetIndex":1,"selectedIndex":0,"optionCount":3,"focused":false,"disabled":false,"connected":true}
```

这排除了该次精确标签未命中/目标不存在/最终节点禁用等解释，且失败先于保存API。不能推断上游调度失败，也不能证明产品change事件或重绘导致重置。05产品标签修订仍有VM依据，但未修复此次原生选择结果；不宣称唯一根因已消除。

## 缺失的区分证据与本次最小边界

原sample覆盖每一步却只留下最后一个快照，无法判定Home/ArrowDown是否移动、Tab是否撤销、聚焦是否持续、同节点选项是否重绘或键盘事件被拦截。历史公共模型选择确有native arrows未移动的真实归档，但不是当前候选控件的本次证据；不直接套用type-ahead策略，尤其本旅程候选名称含中文。

本轮只在原routing driver增加旁观轨迹，保持Home/ArrowDown/Tab、精确最终label断言、所有API状态/请求计数/权重/停用/冷却/SSE/取消断言不变；不改产品、配置、Go、Makefile、Workflow、超时或平台控制器。观察函数在selection-trace.mjs，返回sample/stop，经Playwright handle持有原节点与原option身份；留before focus、before/after Home、每次ArrowDown后、Tab后。监听白名单键/焦点/input/change事件以及父节点变更计数，只返回固定枚举、indices/counts/booleans。defaultPrevented仅是冒泡观察值，不等于证明系统原生处理策略；事件回调不preventDefault、不赋值、不发网络、不注入产品键盘处理。

每个选择finally清理观察器/六个监听器/handle，将完整有界轨迹写入忽略目录 `test-results/routing-selection-<execution>.jsonl`。首4+尾20事件、末12边界、总量/截断计数；失败回执只留末6事件及边界和相对文件路径。数值PID由 `routing_diagnostic_execution`输出，与宿主留档关联；不得输出选项名、值、正文、URL、凭据或绝对路径。旁观执行也可能改变时序，诊断下绿灯不是归因成功。

## 实际检查、限制及下一步

[首份VM记录](candidate-trace-vm-01.log)含114/114 PASS，但外层zsh记录退出码时使用只读status变量而返回1；这是本轮命令包装错误，不是产品或VM测试失败。保留该记录，[复跑VM记录](candidate-trace-vm-02.log)明确114/114、exit0；新增两项只验证轨迹投影/白名单/边界留存/有界及cleanup，不模拟并声称原生行为。[局部检查](candidate-trace-validation-01.log)含三份driver/module测试语法、format、module、vet、build、diff实际exit0；[06身份](final-source-06.json)用于版本绑定。本机已证监听/Mach限制，执行环境未变，不重复此前必然失败的完整make verify，不扩大权限。完整原生执行留给900秒宿主同固定入口。

Trellis-check：已核对当前前端规范、唯一label、测试模块import、事件不写产品/正文的跨执行面数据流及finally清理；无需更新后端契约或重建规范。新helper仅复用固定轨迹概念，不导入有浏览器启动副作用的journey模块。原Makefile/历史门禁/Workflow及上一Run未改；质量检查是局部Ready for Review，不是完整Passed。

首次单独生成identity未沿用Makefile缓存环境，默认Go cache受限而无有效JSON，连带首次文档审计解析失败，见[失败原文](candidate-trace-identity-01.log)及[首次部分审计](candidate-trace-audit-01.log)。按项目既有GOCACHE约定重试，不扩大权限；最终身份与[复验审计](candidate-trace-audit-02.log)才是交接依据。

下一步按06指纹执行完整make verify；若再失败，必须将回执指定JSONL作为宿主产物保留（目录忽略，不把进程或秘密文件提交），提交实际相对留档路径及关键边界摘要。先看before Home是否聚焦、ArrowDown后是否移动、keyup目标及prevented、Tab前后是否回零、原节点/option身份与mutation计数，再决定driver、产品重绘/时序或原生环境修复。不用旧局部PASS放行；通过后独立QA判断13AC。正式Actions非空升级/可信隔离失败恢复与真实供应商仍Not Run；无Git发布或平台修改。
