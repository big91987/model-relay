# seq16 概览跨屏状态返工与宿主交接

## 实际失败与对象

2026-10-06，seq16宿主原样 `make verify` 退出2，HEAD `d62cd51e8de2127548e1ac7469d349443d1bfe3d`，56文件摘要 `74d8f7b07b8afab67e88e983d86e03bbe2b7a1e90f813c7be11ec701bb8c6dfc`。入口仅保留首尾，中间省略，不补造完整日志。

回执当前Go race/HTTP PASS（113.862s）、历史分类器测试PASS、build完成；真实浏览器1280px运行版本及完整控制台旅程均输出PASS（受控上游，不是供应商验证）。随后390px失败：

```json
{"viewport":390,"checkpoint":"visible #summary contains 上游","location":"journey.mjs:14:56","error":"TimeoutError","passwordProbe":null,"keyboardProbe":null}
```

最后已收到真实POST login 200及GET overview 200。这是第二屏概览等待失败，不是监听限制或历史旧版本预期红测。390px未完成，不能将1280px通过称为两屏或18AC验收通过。[上一轮](host-return-05.md)、seq11未知结果及历次失败均保留。

## 根因与最小修正

`journey.mjs`以 `[1280,390]` 顺序共用隔离Go服务，不按视口重建数据。1280px通过UI创建两模型和两密钥、停用第一个密钥后，用第二个完成真实调用；退出登录不清配置或成功记录。390px起点是已配置且当前启动有成功证据，非空安装。

`app.js` overview成功分支正确显示 `当前配置曾成功调用 · 请求 {request_id}`；只有已保存且缺成功证据分支包含 `上游配置已保存`。旧旅程两次登录后硬等 `上游`，与真实成功分支不兼容。不能改产品文案、丢成功证据或重置数据库来掩盖测试假设。

- 仅修改旅程/VM回归、现有frontend verification规范和公共任务，不改产品、协议、鉴权、CSS、数据库、Makefile、依赖锁或平台。
- `nextOverview(page)`在两次登录前登记真实GET overview响应；`loginOverview(page, completed, width)`要求200、非degraded、真实配置/凭据状态和确切模型/密钥计数。1280px仍要求空配置/null证据；390px要求2/2模型、2/1密钥、持久别名alpha-1280和当前成功ID，并核对对应可见summary/cards。
- 没有API/SQL预填、额外模拟overview或睡眠替代。登录显隐/清密、原生Tab/焦点、版本、首次配置全UI、HTTP/SSE、取消、防重复和日志/截图保密断言全部保留。
- 新VM回归调用真实app.js覆盖空配置→已保存无证据→离开后成功证据恢复，核对summary与计数；不是布局或真实浏览器证据。

## 当前验证

| 检查 | 实际结果 | 证据 |
|---|---|---|
| VM回归 | 退出0，60/60，无跳过 | [日志](overview-reentry-vm.log) |
| JS语法、格式、vet、build、模块、diff及门禁/锁文件不变 | 全部退出0 | [分项日志](checks-overview-reentry.log) |
| 当前工作树摘要 | 56文件，Go1.25.6，HEAD同上 | [指纹](source-overview-reentry.json) |
| 新对象完整make verify/原生两屏 | 本次Not Run；稳定沙箱监听/Chromium限制未改变，交宿主 | [限制证据](host-return-03.md)、[首次记录](development.md) |

首次直接指纹命令未使用Makefile项目GOCACHE，默认用户缓存访问被沙箱拒绝（退出1）；分项日志保留该结果，随后使用Makefile同等GOCACHE/CGO_ENABLED/GOTOOLCHAIN生成成功指纹（退出0）。不是产品失败或权限扩大，不因此重跑已确认必然受限监听命令。

当前工作树摘要 `dc9990248951dbfa54fda356975ffbef0dcfc13fef4add95e4a001eda74d69f4`，delivery回执排除以免自引用；不得用seq16旧对象当作本次通过。

Trellis before-dev packages发现成功；已有初始化及源码规范，bootstrap不适用。恢复PRD/架构/交互、公共非终态任务和三里程碑，读取frontend checklist对应DOM、console所有权、验证及复用/跨层guide。trellis-check核对单层改动、真实overview JSON字段alias、计数与共享旅程终态、原断言/秘密保护/门禁不变；update-spec记录跨屏签名、好/坏用例及测试边界，无需求/设计偏差。

## 公共状态与下一步

计划仍Verifying，T001/T002/T003 In Review，M01–M03 Ready for Review，不标Done/Accepted/Delivered。T003已修正确定的起点假设，T002两屏门槛仍待新对象宿主结果。依赖与编号不变，无第二套路线或用户审批。

正式next交tests原样 `make verify`，核对新指纹并保留真实退出码/脱敏首尾。失败按正式边返工，不放宽断言；通过后独立QA逐条18AC及架构/交互核对。无新原生截图、供应商、QA或PR/合并/部署证据；供应商Not Run，路线2–4未实现，按已授权独立Issue接续。不宣称全量门禁或两屏已通过。
