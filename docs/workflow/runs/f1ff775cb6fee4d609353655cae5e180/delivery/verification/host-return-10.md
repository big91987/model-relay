# seq27 闭合模型选择返工

## 对象与真实失败

2026-10-06，宿主原样 `make verify` seq27退出2；HEAD `d62cd51e8de2127548e1ac7469d349443d1bfe3d`，56文件摘要 `9ebdb69472beb0ffb66565f1ec6d3bd47ea0d76b573fdb6338377ce5cc00a6f6`。沿用[PRD](../../product-definition.md)、[交互§7](../../console-interaction.md)、[架构](../../console-architecture.md)、[QA整改](../../acceptance/remediation-01.md)及[上次记录](host-return-09.md)，不重建方案。以下为平台提供回执的脱敏摘录，不是新执行或完整日志：

```text
PASS runtime version actual health, refresh, failure/reentry, nonblocking and text/layout 1280px
PASS QA recovery, precise older evidence, started/unknown/404/injected503, native pagination 1280px
PASS actual Go service browser journey 1280px (controlled upstream, not supplier validation)
PASS runtime version actual health, refresh, failure/reentry, nonblocking and text/layout 390px
FAIL viewport=390 checkpoint=native model selection commits enabled target
location=journey.mjs:170:10 error=AssertionError
modelProbe={stage:"after commit",optionCount:4,targetIndex:1,selectedIndex:0,targetEnabled:true,matchesTarget:false}
exit status 1
make[1]: *** [browser] Error 1
make: *** [verify] Error 2
```

当前Go/race/HTTP与build通过，1280px完整扩展旅程通过，但390px未完成。探针证明旧Space/Home/ArrowDown/Enter/Tab序列没有选中目标；它不证明OS弹层内部如何处理按键，也不是产品CSS/模型协议缺陷证据。保留seq25及seq27失败，不用旧部分绿色放行。

## 变化边界

变化只在原项目Harness：`keyboardModel(page, alias)` 保持原生Tab可达、focus-visible和outline检查，从真实控件读取公开选项与启用态。推荐避免OS弹层序列：保持闭合，按optionCount次ArrowUp抵达首个启用项，再按目标启用序号ArrowDown；Tab离开后仍要求原option索引与精确inputValue同时匹配。实际按键有界，不轮询未提交值、不补点、不程序赋值、不使用selectOption、不重发业务请求。enabledTargetIndex作为纯数值加入原白名单probe；失败仍硬失败，不接受“至少选中某个模型”。

修改原 `tests/browser/journey.mjs`、`ui.test.mjs`、[Trellis验证规范](../../../../../../.trellis/spec/frontend/verification.md)及公共交付状态。产品app.js/index.html/CSS、Go/SQL/接口/地址策略、PRD/设计、QA资产、平台、Makefile和锁文件不改。无需新依赖、组织流程或路线2–4实现。Trellis初始化及真实规范存在，packages发现成功，bootstrap不适用；before-dev恢复索引及相关规范，check复核单层边界及旧断言，update-spec更新已有选择契约而非创建第二套规范。

## 验证证据

| 检查 | 实际结果 | 证据与边界 |
|---|---|---|
| 修复前闭合选择契约 | 84项，79绿5红，退出1 | [红测](closed-select-before.log)；helper契约不是OS复现 |
| 修复后frontend-test | 85/85，退出0 | [日志](closed-select-vm.log)；77产品VM+8真实helper源码契约 |
| JS语法/Go格式/模块/vet/build/diff | 退出0 | [分项](checks-closed-select.log)；固定门禁和锁文件相对HEAD无差异 |
| 公共文档本地引用与脱敏检查 | 6文件90引用无缺失，退出0 | 同分项日志；新增证据不含绝对工作区路径 |
| 新对象 | 56文件，指纹命令退出0 | [指纹](source-closed-select.json)：`d9d6abe3f2599bc9dfdeaff4a58f13e86f002779bec4793d2434f76417cf21a8` |
| 新完整make verify/真实Go两屏 | Not Run | 稳定沙箱监听/Chromium限制未变，沿用已有受限证据，交宿主原入口 |
| 独立QA/供应商/发布 | 无新通过 | 首轮No-Go与QA-03驱动差异保留；供应商Not Run，无发布回执 |

新增helper契约覆盖立即/离焦提交、从最后项选择第一/最后目标、跳过前置禁用项、缺失/禁用目标在按键前失败，以及所有箭头被忽略时必须失败并保留index0/matchesTarget=false。模拟只验证Harness算法和拒绝行为，不宣称真实浏览器修复已确证。日志去除本机工作区路径，不含输入值/凭据。

## 交付审查与下一步

交付审查为Ready for Dispatch至tests，不是版本通过。任务T001/T002/T003保持In Review，M01/M02/M03保持Ready for Review，不Done/Accepted，不关闭QA缺陷。公共计划仅投影本轮事实，历史不覆盖。关键路径：新指纹原样 `make verify` → 失败正式返工或独立QA逐条18AC/整改原层复验。宿主必须保留两屏精确模型、原生键盘、依赖启用恢复、发送取消、同ID记录及分页全部断言；失败保留新modelProbe数字/布尔字段和真实退出码。不把85单元绿色或旧1280px通过称新两屏通过。
