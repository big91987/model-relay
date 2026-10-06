# seq14 原生键盘焦点返工与宿主复验

2026-10-06，研发节点，同Run续接。权威输入：[PRD AC-01/02及完整18AC](../../product-definition.md)、[交互设计](../../console-interaction.md)、[架构验收边界](../../console-architecture.md)。不修改范围、设计或产品协议。

## 实际失败与可诊断范围

seq14宿主原样 `make verify` 实际退出2，对象摘要为 `37b9d52f1d2458e5333a4675648ca28c13a513af8eeb094a37120a46d78de861`，HEAD为 `d62cd51e8de2127548e1ac7469d349443d1bfe3d`。入口提供首尾，中间省略，不能补造完整输出。

尾部当前 `internal/relay` race/HTTP PASS（113.048s）、历史分类器测试PASS、build完成；浏览器明确输出1280px运行版本旅程PASS，随后失败：

```json
{"viewport":1280,"checkpoint":"native keyboard action with visible focus","location":"journey.mjs:26:10","error":"AssertionError","passwordProbe":null}
```

这是当前旅程焦点断言失败，不是历史隔离旧版本的预期红测，也不是沙箱限制。原生390px及后续首次调用/SSE完整旅程未通过。登录完成边界修改已使本次旅程推进到新失败点，不等于所有登录/生命周期或AC已验收。[前次记录](host-return-04.md)及所有历史失败保留。

## 源码缺口与变更边界

`keyboardAction`原实现先发一个Tab，再`locator.focus()`，立即将activeElement和outline合成一个布尔断言。这不能证明真实Tab能到达目标，也不能从seq14回执区分未获焦、样式缺失或异步抢焦。按旅程顺序，第一个调用用于取消模型编辑；旧回执并未记录目标与焦点细节，不能直接判定产品CSS或焦点转移根因。

本次修正明确的测试契约缺口，不以猜测修改产品：

- 等待目标可见，通过原生Tab最多80次遍历；不调用程序化focus，不补偿点击，不替换Enter/Space、不注入产品keydown。
- 先要求目标确为activeElement，再要求`:focus-visible`、outline非none且宽度大于0，然后仅一次原生激活。现有取消零PATCH、占位复制、诊断展开、发送/取消、导航与两屏断言全部保留。
- 失败只记录固定安全目标类别、key、Tab次数、焦点布尔、outline枚举/宽度；不读取密码/密钥/提示输入、DOM文本或HTML。安全控制ID/导航页名白名单，不输出动态模型名称。成功后清除探针；全局catch仅焦点checkpoint带该探针。
- 仅修改 `tests/browser/journey.mjs`、前端verification规范与本Run公共文档。没有产品实现、HTTP/存储、Makefile、锁文件或平台变更。新工具和供应商依赖均未引入。

## 本地真实检查与限制

| 检查 | 实际结果 | 证据 |
|---|---|---|
| 原有VM生命周期回归 | 退出0，59/59，无跳过；不模拟或证明原生Tab | [VM日志](keyboard-traversal-vm.log) |
| JS语法、格式、vet、build、模块、diff | 各退出0，固定门禁/锁文件diff为空 | [分项日志](checks-keyboard-traversal.log) |
| 版本摘要 | 56文件，Go1.25.6，HEAD同上 | [新对象指纹](source-keyboard-traversal.json) |
| 新对象完整make verify/原生浏览器 | 本次Not Run；既有沙箱监听/Chromium限制未改变，不重复必然受限命令，正式留宿主 | [已确认限制](host-return-03.md)、[首次限制](development.md) |

新工作树指纹：`74d8f7b07b8afab67e88e983d86e03bbe2b7a1e90f813c7be11ec701bb8c6dfc`。回执摘要覆盖产品、测试、契约和规范，排除delivery回执避免自引用；不能使用seq14旧对象当作本次结果。

Trellis before-dev实际packages发现成功，恢复现有真实frontend规范、DOM生命周期、console所有权、复用和跨层guide；初始化与非模板规范已存在，bootstrap不适用。trellis-check核对Tab目标、原有激活/业务断言、安全诊断白名单及门禁不变；update-spec将上述签名、好/坏用例和两屏验证约束写入现有verification规范。仅测试层变更，无公共接口数据流变化，不需要需求/设计返工。没有新增VM来假冒浏览器焦点。

## 公共状态与正式交接

按交付Skill恢复既有非终态任务、三里程碑及权威链接，编号/负责人/依赖不变；审查确认当前修订属于T003 Harness，T002真实两屏退出条件仍未满足。T001/T002/T003 In Review，M01–M03 Ready for Review，阶段Verifying；不标Done/Accepted/Delivered。新文档引用与diff检查另见分项日志。

next交tests宿主原样 `make verify`，核对新摘要并保留真实脱敏首尾/退出码。若失败，请保留keyboardProbe（目标、Tab数、focused/focusVisible、outline）并按正式失败边返工，不能放宽门禁；通过后由独立QA逐条18AC及设计验收。原生失败尚未复验，不宣称产品根因已修复或门禁通过。

无新截图/供应商联调/QA/PR/合并/部署证据，供应商Not Run。路线2–4未实现，按既有持续授权后续独立Issue接续。不需要用户额外批准或扩大权限。
