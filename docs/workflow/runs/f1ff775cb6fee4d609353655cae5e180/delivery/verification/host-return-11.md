# seq29 原生模型选择返工

> 后续用户明确要求先收敛原生诊断：本文件的未宿主类型前缀建议已撤回、交接未接受，原日志和指纹保留为历史，不能用于当前放行。当前诊断对象与动作见[原生边界诊断](model-selection-diagnostics.md)。

## 恢复与范围

沿用同Run PRD、架构、交互及[公共任务](../README.md)，恢复[上一对象](closure-contracts.md)及[QA整改](../../acceptance/remediation-01.md)。本轮只处理T003的宿主选择失败，不重建需求/设计或路线。T001/T002/T003仍In Review，M01/M02/M03仍Ready for Review；独立QA首轮No-Go及QA-03驱动语义差异保留，研发不关闭。

seq29平台实际执行原样 `make verify`，退出2，对象56文件摘要 `475091a58b28ca779261cb8c1dd24a55118d9c74df1c4d1be8fb5aa4fc82c63a`，HEAD `d62cd51e8de2127548e1ac7469d349443d1bfe3d`。当前Go race/HTTP与build、1280px版本和完整QA恢复/精查/分页控制台旅程通过，390px版本通过；390px模型选择未通过。以下为会话回执的脱敏失败尾部摘录，不伪造完整源日志：

```text
FAIL actual Go service browser journey:
viewport=390
checkpoint=native model selection commits enabled target
location=journey.mjs:168:10
error=AssertionError
modelProbe={stage:after commit,optionCount:4,targetIndex:1,enabledTargetIndex:1,selectedIndex:0,targetEnabled:true,matchesTarget:false}
make[1]: *** [browser] Error 1
make: *** [verify] Error 2
```

不是focus断言，也不是供应商错误。旧seq18绿色和seq29局部绿色不能放行新对象。历史旧版本HTTP红测是门禁预期，不能视为本次根因。

## 源码定位及推荐修正

`keyboardModel`已确认目标存在且启用，Tab焦点门槛通过，但此前弹层及闭合箭头两种序列均未提交目标。回执仍不能证明是浏览器OS事件处理、按键路由或其他行为导致，不声称已找到产品根因；源码检查未发现需要修改产品模型列表或CSS的证据。

改动位于项目旅程的选择方式：选项文本由真实app.js以`alias + ' → ' + model`生成，对已有控件原生Tab到达并保留focus-visible/outline，然后一次 `page.keyboard.type(alias, {delay: 30})` 类型前缀选择、Tab离开，仍严格验证原索引和完整alias一致。不调用selectOption、不赋DOM值、不触发模拟change、不补偿click、不循环重试、不改产品键盘处理器。新增离焦前`typedSelectedIndex`安全诊断以区分选择和提交时状态；只记录固定阶段/数字/布尔，不输出选项文本、消息或凭据。

此为有边界的Harness修正推荐，真实效果必须由新对象宿主回执判定，不把替换输入序列称OS根因已解决。原首次全UI配置、两屏依赖恢复、发送取消、精确记录、分页、保密、版本、密码显隐、剪贴板和四项交互收尾门槛保留。无产品/Go协议/DB/需求设计/QA原资产/平台/Makefile/锁文件改动。

## 验证记录

- [修正前新增契约](model-typeahead-before.log)：102项101通过1失败，命令 `node --test tests/browser/ui.test.mjs` 退出1。新double明确箭头不改变首项、类型前缀可选目标；仅模拟算法，不是原生红测。
- [修正后契约](model-typeahead-vm.log)：102/102，退出0，含77产品VM与25旅程helper/收尾契约；立即/离焦提交、第一/最后目标、前置禁用、缺失/禁用目标、忽略所有输入必须失败均保留。
- [分项检查](checks-model-typeahead.log)：实际命令/退出码；JS语法、Go格式/模块/vet/build、diff与门禁锁文件检查。
- 初次新证据路径扫描退出1，发现红测堆栈含本机工作区路径；只归一为仓库相对路径，保留断言/行号/失败内容。修正后四份公共文档本地引用与新证据路径扫描通过，记录追加至分项日志，不删除初次扫描失败状态。
- [当前源码指纹](source-model-typeahead.json)：以项目GOCACHE运行 `go run ./tests/evidence`，退出0；56文件摘要 `f4998ab26ce6e372d52eec4f9a01c3efe3b6a8b6d085175d5a5a72a02cf5eb87`，HEAD仍为上述基线。交接必须核对此新对象而非seq29对象。
- 新完整`make verify`/真实Go两屏：Not Run。已确认稳定沙箱监听/Chromium限制未变，按[既有受限证据](expiry-watchdog.md)不重复必然受限命令。下一节点宿主原样完整执行。
- 独立QA、供应商、发布：无新增通过；供应商Not Run，路线2–4未实现，无PR/合并/部署回执。

Trellis before-dev已恢复frontend索引、DOM/lifecycle/verification与shared code reuse guide，packages发现成功，当前真实初始化/规范存在，bootstrap不适用。check按单层变更复核精确门槛、源码选项标签契约、无敏感输出、原QA文件未改及回归；update-spec原位更新验证规范，不建第二套规范。

## 交付审查

Ready for Dispatch至tests，不是G4版本通过，不需要扩大权限或阶段批准。关键路径：新指纹原样 `make verify` → 失败正式返工／通过后独立QA逐条18AC、原层整改、四项补充及QA-03语义评审 → Connector发布。失败保留stage/typedSelectedIndex/selectedIndex/targetIndex及真实退出码，不能放宽精确值/原生焦点/业务断言，不能用102单元绿色替代真实两屏。历史报告和证据不覆盖。
