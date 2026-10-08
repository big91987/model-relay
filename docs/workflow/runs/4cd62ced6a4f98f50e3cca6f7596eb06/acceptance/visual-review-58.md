# QA58 实际页面对照与原缺陷关闭依据

对象为records-detail-entry-r1 / seq57 / 134件fb4066fa…；完整绑定见acceptance-report-58.md、evidence/attempt-58/integrity.json。材料design-v0.2.0原interactions“页面及行为”、PRD§7与原型A为预期，不用当前实现反向删减规则。

## 实际查看范围

QA通过本地image viewer实际看了visual-{1280,1440,390}的overview/models/services/playground/keys/tenants/limits/usage/billing/records十页，即30张；另外三个records-detail、一张390 keys-editor及continuous1280-before/after-correction、1440-after-correction、390-after-correction，共38张当前图片。额外查看材料overview.jpg和tenant.jpg作基线；不是重新运行原型。63张visual PNG全量摘要核对/保留，不声称已逐张人工查看全部63。

原文件在test-results为可覆盖运行目录；本轮脱敏副本在evidence/attempt-58/host，每件SHA/长度/原mtime见integrity.json，共84件（69PNG、15JSON）。页面截图不单独构成UI旅程Pass，原操作结果来自seq57真实Go浏览器及实际代码合同；QA审查不伪称亲自运行另一套浏览器。

## 逐页比较

| 页面 | 当前可观察事实 / 原型等价组织 | 操作/数据补强及结论 |
|---|---|---|
| 工作台 | 明亮业务首屏、真实调用/成功/21微美元/延迟、趋势与风险、两模型当前scope状态及来源，管理模型/预算入口 | 当前周期活动0→1且旧closed bill不变；三视口实际点入口保持scope。QA47-01工作台项Closed |
| 模型 | 两公共名、真实来源/类型/普通SSE边界、状态、当前价3/2 USD每百万、调用延迟或“未测”、上下文启停/调用 | summaryCell按真实API值逐列断言、停→启→读回；不是假延迟。QA47-01模型项Closed |
| 服务 | 地址/来源、关联两模型及各候选优先级、核验时间/执行revision/适用性，关联入口 | 真实GET核验及当前revision；关联模型抽屉可进入编辑。QA47-01服务项Closed；success英文/断词P3保留 |
| 租户 | 两租户实际管理员负责人、成员、授权模型、含撤销历史Key数量、当前预算/已用/预留/上界/周期 | owners由有效membership派生非调用者资格；跨租户角色API反例及native应用Key入口保持目标scope。QA47-01租户项Closed |
| API Keys | 多Key过期/撤销/启用、安全标识/授权/应用/预算/期限和轮换历史入口 | 真实创建/一次值/轮换/截止/收窄；三视口J03；移动editor focus/字段/保存可触达 |
| 配额 | tenant及各Key独立五维、有效上限/已用/剩余/平台边界，未知与预留说明 | J04硬预算耗尽429→修正→再调用；后台五维并发矩阵另记API，不把一条UI链当十维全UI |
| 用量 | 已确认输入132/输出176、费用与成本独立、pending0、筛选/导出 | J06核对同ID及J07筛选导出；费用保留精确微美元非供应商报价 |
| 定价账单 | 模型当前价/版本入口、周期原金额/成本/pending、结账/详情更正，外部支付未接通 | J06当前/pending拒结→跨月→合法结账→更正；前后更正图含读取中提示，是过程图，最终visual无未完成读取，不误判为卡死 |
| 日志 | 原始上游Token未知与已确认3/4、pending0同列，明示口径；同ID抽屉可核对 | J06原native行按钮、200/ID/字段、CSV同口径；原attempt/快照不改。QA47-02 Closed。技术字段P3不掩盖已确认值 |
| 调用测试 | 应用Key/模型/消息/流式、发送/取消/同ID入口，未使用表单空态没有假输出 | 三视口真实普通/SSE、双发送一次及取消/恢复另有native证据；截图空态只是隐私采集策略，不拼成调用Pass |

## 布局与边界

桌面稳定226px侧栏/72px顶栏、紧凑数据表与主次动作；390使用菜单，表格局部横滚，页面根无横向溢出。部分长英文/枚举和安全标识会断词，记录QA58-01 P3，不要求复制原型像素或演示卡片金额。原型通知/角色比较不是本期新增外部系统，实际角色由会话决定，没有前端角色切换。

弹层原native键盘firstTab/cycle/Escape/returnFocus、三视口编辑中断/faults等由seq57当前测试执行；未查看的其余editor/detail图片只声称保留和摘要一致，不夸大人工复核范围。私密值采集前清除或mask、localStorage/sessionStorage为空，保留的页面无API Key完整秘密、密码或供应商凭据。
