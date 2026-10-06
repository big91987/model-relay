# 独立 QA 第1轮整改与新对象宿主交接

## 基线、责任和变更边界

2026-10-06研发接回[整改包](../../acceptance/remediation-01.md)、[报告](../../acceptance/report-01.md)、[矩阵](../../acceptance/matrix-01.md)及[有效红灯](../../acceptance/evidence/independent-boundaries-attempt-03.log)。seq18固定 `make verify` 退出0只适用于旧工作树 `dc9990248951dbfa54fda356975ffbef0dcfc13fef4add95e4a001eda74d69f4`；seq19独立QA No-Go不是门禁绿色的撤销，也不是本次新实现已通过。

继续沿用[PRD](../../product-definition.md)、[架构](../../console-architecture.md)、[交互](../../console-interaction.md)。任务T002/T003顺序返工，单负责人研发，不新增Agent、路线、模块、数据库迁移或平台补丁。Trellis初始化/源码规范已存在，packages发现为single-repo/backend/frontend；before-dev读取frontend三规范、backend运行/console契约、复用/跨层guide和公共任务，bootstrap不适用。后端业务协议未变，Go只扩展隔离浏览器上游夹具。

最小行为差距在真实 `app.js`：当前操作失败未清字段、保存与回读共用反馈、概览有成功ID却没有明确精确动作。本次只改该状态责任和HTML动作，测试文件补原层异常/键盘/分页回归，Trellis沉淀执行契约及现有公共任务更新。保留原门禁/锁文件/QA原始资产；不重建产品或需求/设计。

## 实施与验收追踪

| 缺陷 / 范围 | 根因修正 | 可执行证据 / 剩余门槛 |
|---|---|---|
| QA-D01 / AC-07/08/12/15 | 上游catch仅当前owner清凭据并保地址/启用草稿；生成catch仅清key/message，保模型/SSE/已有输出/原始诊断/真实ID及记录入口；应用401不登出管理会话 | VM写拒绝/网络、应用/供应商/JSON/SSE/网络清密、旧catch跨页隔离；真实Go两屏拒绝→恢复→新调用→同ID用例已进入原browser门禁，待宿主 |
| QA-D02 / AC-07/15 | PUT确认和GET回读分离；GET故障明示已保存但最新读取未确认，保故障类别/公开草稿，清秘密，增加GET-only重读；PUT不明结果不谎称已保存 | VM回读失败；原生真实PUT200+明确注入GET503、零自动重写、重读恢复；网络注入明确非服务端存储损坏 |
| QA-D03 / AC-05/13/14 | 概览显示实际alias/scope，加载前清旧动作ID；独立overview-record携带显示ID，view-record携带本次ID；普通导航始终泛列表，清旧精确输入/详情 | VM不同本地ID/会话恢复/新登录/无证据及失败刷新；原生首分页之外成功ID、更新行、刷新/重登录均要求精确GET等于显示ID，待宿主 |
| QA-G01 / AC-03/06/07/09–16 | 真实Tab门禁复用到输入、模型选择、复选框、创建/取消/编辑/启停、密钥复制/关闭、发送/取消/诊断、定位/分页；响应屏障避免旧notice误判完成 | 原click/密码旅程保留，新增全控件键盘覆盖而非声称整段无鼠标；实际HTTP动作和两屏焦点尚待宿主 |
| 依赖恢复 | 原对象上游同地址空凭据保留/变址缺凭据拒绝；上游停用恢复，原选模型停用不静默换beta，全部模型/密钥停用与原集合恢复、原密钥拒绝后启用新生成 | UI无API/SQL预填；生成POST在浏览器屏障中等待另一个实际管理页停用模型，然后实际Go拒绝；受控上游计数必须不变，无产品重试 |
| 记录边界 | 真实UI产生52次未知key拒绝，把先前成功移到第2页；下一页/旧ID精查/真实404、显式注入503后恢复、另一页面观察真实started；普通0和未知用量逐字段核对 | sameRecord核对真实精确响应ID/alias/model/key_id/state/duration_ms/usage及可见详情；未知key拒绝不得捏造提交alias或key归属；52次及概览显示后另一页面并发新增拒绝都须上游零触达 |

明确夹具：`tests/browser/main.go`受控上游新增非法JSON、无DONE的截断SSE、无usage合法响应。保存变址用同一测试authority的不同路径，不依赖外网DNS。分页不是伪造历史SQL数据，而是同UI真实拒绝记录。注入PUT网络/管理GET503只证明前端状态恢复，不能叫真实存储503；既有真实Go storage fail-closed回归未删。

## 原独立 QA-03 驱动不一致（留给独立QA核对）

原[独立脚本](../../acceptance/evidence/independent-boundaries.mjs)未改。其QA-03调用fixture的 `app.navigate('records')`，该helper激活泛 `[data-view=records]` 按钮，不是新增独立概览成功动作；仍要求泛导航隐式继承证据ID，与整改包“普通调用记录仍为泛列表，不要所有导航都隐式改到某请求”的要求冲突。

本轮不以改QA脚本、改fixture导航行为或重新隐式过滤泛列表来刷绿。原脚本当前3/4通过、QA-03仍失败，见[首次修后](qa-boundaries-original-after.log)/[最终修后](qa-boundaries-original-final.log)。QA-01/02/04红转绿可直接比较原有效红灯。项目门禁中的新增测试显式激活 `overview-record.onclick()` 并同时验证泛导航无精查，覆盖不同本地ID、会话恢复、新登录、旧成功首分页之外和失败刷新。最终是否关闭QA-D03由独立QA按真实宿主动作判断；不能把原4条报告改写成4/4通过。

## 验证、历史和环境边界

最终对象：56文件，Go1.25.6，HEAD `d62cd51e8de2127548e1ac7469d349443d1bfe3d`，工作树 `6cdd533c1dd449f9aa28dcf1422228c399e94ed2a589175f465ff16839cd7a52`。`make frontend-test`退出0（74/74、无跳过）；3条Go race专项退出0（8.603s，保留非致命链接器warning）；最终JS/格式/vet/build/模块/diff及固定门禁/锁文件检查退出0。6份本轮公共文档74个本地引用存在，83个DOM ID唯一且源绑定完整。见实际日志，不含完整宿主门禁通过声明。

- [VM初次](qa-rework-vm-attempt-01.log)69项有4失败：旧摘要精确字串、网络fixture未移除已reject项、新概览fixture误用未配置快照、旧工具记录测试调用泛导航而非精确动作。日志保留；修正快照/队列，摘要增加alias/scope仍精确比较，工具记录改激活本次动作且保留原unknown/zero断言，不删门禁。
- [VM第二次](qa-rework-vm-attempt-02.log)69/69通过；[第三次](qa-rework-vm-attempt-03.log)74/74通过，增加恢复/新登录/迟到上游/失败刷新边界。最终原 `make frontend-test` 结果见[最终VM](qa-rework-vm-final.log)。这些均是Contract证据，不是Chromium焦点或网络实测。
- 格式、Go模块校验、vet/build、JS语法、diff及门禁/锁文件原样检查见[分项](checks-qa-rework.log)；不监听的3条Go race console契约见[Go专项](qa-rework-go.log)。项目代码版本见[当前指纹](source-qa-rework.json)，HEAD仍基线d62cd51；不将旧绿色用于新对象。
- 新对象完整 `make verify`、真实Go/Chromium两屏异常及分页 **Not Run，待宿主tests**。稳定沙箱监听限制和执行环境未变，按既有[限制记录](host-return-03.md)/[首轮](development.md)不重复必然在监听处退出的完整命令，不索要扩大权限、不把VM称原生通过。此次Go夹具和浏览器用例已扩展，实际原生失败须由宿主回执诊断，不能先称通过。
- 原独立QA报告、红灯、seq18截图与所有历次失败不覆盖。新增日志只脱敏工作区绝对路径，不修改结果。截图仍仅清密后生成；新key/message/原文排除已加入实际服务日志和记录检查。无新增供应商联调、PR、合并、部署或部署后浏览器回执。

## 公共状态、门槛和交接

T001仍In Review/M01 Ready for Review；T002/T003整改实现与可运行专项完成，进入In Review；M02/M03 Ready for Review，不Accepted/Done/Delivered。独立QA首轮No-Go保持历史有效，三缺陷与QA-G01待新对象宿主原层证据及独立复评，不由研发自行关闭。

下游原样 `make verify`：核对当前指纹、保留真实退出码、脱敏失败首尾和新两屏截图。失败按正式边返工；通过交独立QA逐条18AC、整改包及驱动不一致复验。供应商Not Run；后续路线2–4留项目公共计划，不在本Run实现。Git发布继续Connector负责，不启动外部工作流。
