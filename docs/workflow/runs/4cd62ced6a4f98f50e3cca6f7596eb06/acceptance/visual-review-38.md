# QA38 批准原型与实际UI证据对照

冻结design-v0.2.0，archive5f51d35fdc15ed7214749f9413879b4238097f1d6b6bbe2cc0d508d8e103cef7；实际材料相对路径及19件摘要见evidence/attempt-38/material-integrity.json。权威为材料docs/interactions.md「页面及行为」及docs/acceptance.md AC01；本Run verification-design.md§2明确要求十页实际页面、原型对照截图和动作结果。不是要求复刻演示金额/图标或擅自新增商用高级功能。

## 已能作出的判断

本轮查看批准材料docs/screenshots/keys.jpg与当前console-1280.png：稳定左栏、明亮布局、Key非空紧凑表格、名称/应用/安全标识/授权/状态/期限/预算摘要和实际操作可见；原型的前端角色切换、B/C切换、演示标记不应带入真实身份产品。当前故障截图显示居中统一editor及右侧详情drawer；390折叠导航、局部表格滚动和恢复按钮可触达。当前1440账单更正截图显示测试USD、租户费和成本、原周期与开放调整周期分别列出。品牌、蓝色色值、表格装饰简化本身不是阻断缺陷。完整秘密不出现在这些查看样本中。

上述是这些截图/动作的局部对照，不是另外七页的实际视觉结论。当前产品源码提供overview真实趋势/费用分布、模型/服务/成员/预算/用量/日志/调用入口，CSS提供drawer与响应式；Static Review只能证明预期实现位置，不能代替真实页面对照。

| 页面 | 批准信息与关键动作 | 当前对应资产 | 本轮实际对照证据/限制 | 分层结果 |
|---|---|---|---|---|
| 工作台 | 今日调用/成功/费用/延迟、趋势/租户费用/风险及导航 | workspace.js overview；OverviewAndExactRecords；console十页循环 | API聚合有测试；十页循环仅h1/根宽度；没有当前工作台截图与原型对照 | Static Review Pass；UI设计对照 Not Run |
| 模型托管 | 公共名/来源/能力/状态/价格、发布/编辑/授权/调用 | models、modelEditor、J01/J02 | 生命周期动作通过；无当前模型页列表/详情/授权弹层的原型对照截图。当前列表仅alias/协议说明/状态/revision/操作，其他信息在何处满足日常管理须实际核对 | UI旅程 Pass；UI设计对照 Not Run |
| 供应商与服务 | 来源/地址/状态/核验事实、保存/失败修正/核验 | services、serviceEditor、J01 | provider路径已通过；无当前服务页及核验详情原型对照截图 | UI旅程 Pass；UI设计对照 Not Run |
| API Keys | 安全标识/应用/授权/用量/期限/状态、生命周期/恢复 | console三截图、fault截图、J03 | 1280Key及1280签发恢复/390轮换503人工查看；三视口有状态/unknown/恢复自动断言 | 本行局部UI设计对照 Pass |
| 租户与成员 | 租户范围/状态/角色、邀请接受/降权/移除/暂停 | J01/J05、PeopleScope | 真实身份旅程已通过；没有当前租户成员页及角色弹层对照截图 | UI旅程 Pass；UI设计对照 Not Run |
| 配额与预算 | 各独立维度/有效上限/已用/剩余/硬软与未知口径、调整 | limits、limitSummary、J04、资源矩阵 | 硬预算旅程及API组合通过；Key卡片的JSON展示不能由代码推演为全部管理信息可读；缺该页真实对照 | UI旅程 Pass；UI设计对照 Not Run |
| 用量分析 | Tokens/费用/待确认、模型租户Key时间筛选/导出 | usage、filters、J07及HTTP聚合 | 查询资产存在；没有当前分析页面筛选/非空结果对照截图 | Static Review Pass；UI设计对照 Not Run |
| 定价与账单 | 价格版本/成本租户价/周期/核对/结账/更正/导出 | J06及三视口账单截图 | 1440追加更正页面人工查看，J06动作与账本事实已通过 | 本行局部UI设计对照 Pass |
| 调用日志 | ID/范围/状态/Token/费用/延迟、分页/详情多尝试/导出 | J07与requestDetail | 正式过滤/分页/迟到旅程已通过；缺当前列表/多尝试详情对照截图 | UI旅程 Pass；UI设计对照 Not Run |
| 调用测试 | Key/公共名/消息/SSE、普通/取消/同ID、不重放 | console/fault/J01/J02/J04/J05 | 真实普通/SSE/双发送通过；最终截图回Key页，不是调用页对照证据 | UI旅程 Pass；UI设计对照 Not Run |

## QA38-01：强制AC01证据缺口

结果为**Not Run（完整UI设计对照）**，不是“页面不存在”或“产品运行失败”。tests/browser/console.mjs:140的十页循环只检查一个可见h1和document.scrollWidth≤innerWidth，:142保存截图时已回API Keys。J01–J07及fault的当前PNG主要是Key和账单；artifacts明确可查。QA35矩阵AC01-UI-full-review已明确Not Run，本轮整改和宿主绿色未提供对应的新证据，不能自动提升。

按冻结验证契约，研发应在原真实Go浏览器入口收集十页非空/适用空态、关键编辑/详情和390操作可触达证据，保存准确页面名/角色/scope及实际字段/动作断言；提交页面对照映射供独立QA逐页判断。只补必要证据，不要求所有状态组合截图、不要求像素一比一、不改变冻结规则、不用API填补UI旅程。若既有当前同版证据能覆盖，给出具体文件和真实断言即可，不必重写产品；但现有63件图与h1循环不满足这个要求。

此项是强制门槛未验证，owner=development（测试资产）；不是给可能布局差异直接定P1产品bug。尚无新增实际产品P0/P1；原QA35-01可关闭，但完整产品Go仍不能判定。真实自托管供应商凭据、外部身份/支付不在这项整改中。
