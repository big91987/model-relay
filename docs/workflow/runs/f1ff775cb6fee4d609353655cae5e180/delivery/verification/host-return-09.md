# seq25 原生模型选择提交返工

## 恢复对象与真实失败

2026-10-06固定宿主 `make verify` seq25退出2，HEAD `d62cd51e8de2127548e1ac7469d349443d1bfe3d`，56文件指纹 `7f6faaac87ad3501cf76108b850db068439f0c6a47c8e10003e5caa4694775c9`。恢复[PRD](../../product-definition.md)、[交互](../../console-interaction.md)、[架构](../../console-architecture.md)、[QA整改](../../acceptance/remediation-01.md)、公共任务和[前次记录](evidence-metadata.md)，不重建方案或覆盖历史。

回执当前Go/race/HTTP与build通过；1280px版本、QA新增恢复/精确旧证据/started/unknown/404/注入503/原生分页和完整控制台旅程通过；390px版本旅程通过，随后失败。会话提供的脱敏尾部为：

```text
FAIL actual Go service browser journey:
viewport=390
checkpoint=native keyboard action with visible focus
location=journey.mjs:161:10
error=AssertionError
passwordProbe=null; keyboardProbe=null; clipboardProbe=null
exit status 1
make[1]: *** [browser] Error 1
make: *** [verify] Error 2
```

这是seq25回执字段摘录，不是重跑日志或完整原输出。行161实际是 `keyboardModel` 的最终 `inputValue() === alias`，此前 `keyboardFocus` 已成功且将probe清空。不能据泛化checkpoint把失败归为CSS或焦点不可见；旧回执未包含实际选项/值，不能确证OS popup状态。历史预期红测仍不等于当前产品失败。

## 最小变化与依据

旅程原helper执行Home后循环ArrowDown，读尚未提交的DOM值，最多1000次；没有Enter或离开下拉的提交边界。两屏共用真实Go服务/持久模型，390px目标不是列表第一项，1280px默认值通过不证明后续项可正确提交。推荐只修项目Harness的原生选择契约，不猜测修改产品模型/CSS/鉴权。

- `keyboardModel(page, alias)` 保留真实Tab可达、activeElement、focus-visible和outline断言，读取既有公开option索引及启用态，缺失/禁用直接失败。
- Space打开→Home首项→按目标索引ArrowDown→Enter提交→Tab离开，再要求选中索引与精确inputValue都匹配。没有selectOption、程序赋值、补发请求或API/SQL业务预填。
- 新checkpoint与modelProbe只含固定阶段、计数/索引和布尔值，不包含凭据、提示词、任意DOM或原始错误正文。成功清probe，失败保留提交前后边界。
- 四个helper契约测试从真实journey源码提取该函数执行，覆盖立即/延后提交及缺失/禁用目标；选择夹具是Harness单元模拟，不是浏览器或OS复现。77条原产品app.js VM断言不删。

写集仅原 `journey.mjs` / `ui.test.mjs`、Trellis验证规范与公共任务/里程碑/本记录。产品、Go协议、数据库、需求/设计、QA资产、Makefile/锁文件和平台均不改；无新依赖。Trellis初始化及真实规范存在、packages发现成功，bootstrap不适用；before-dev读取frontend索引/三规范/shared guides，update-spec记录模型选择提交契约，check核对原门禁与分层证据。

## 验证与状态

| 检查 | 结果 | 证据与边界 |
|---|---|---|
| 修复前新增helper契约 | 81项，77绿4红，退出1 | [红测](model-selection-before.log)，非原生复现 |
| 修复后原frontend-test | 81/81，退出0 | [日志](model-selection-vm.log)，77产品VM+4helper契约，非原生证据 |
| JS语法/格式/模块/vet/build/diff | 退出0 | [分项](checks-model-selection.log)，门禁/锁文件相对HEAD无差异，不代替完整门禁 |
| 公共文档本地引用 | 6文档84引用无缺失，退出0 | 本轮交付总览/两个任务/两个里程碑/本记录，见分项追加回执 |
| 新对象 | 56文件，退出0 | [指纹](source-model-selection.json)：`9ebdb69472beb0ffb66565f1ec6d3bd47ea0d76b573fdb6338377ce5cc00a6f6`，沿用HEAD |
| 新完整make verify/真实Go两屏 | Not Run | 已确认稳定沙箱监听/Chromium限制未改变，不重复必然受限命令，交宿主固定入口 |
| 独立QA/供应商/发布 | 未新增通过 | 保留首轮QA No-Go/QA-03驱动差异；供应商Not Run；无PR/合并/部署回执 |

任务T001/T002/T003 In Review、M01/M02/M03 Ready for Review；没有Done/Accepted。公共交付评审：Ready for Dispatch至tests，关键路径新指纹原样 `make verify` → 真实失败返工或独立QA逐条18AC及整改包复验。成功仅由新宿主回执判断；不能把81单元通过、seq18旧绿色或seq25的1280px部分通过当作新两屏验收。路线2–4仍未实现。
