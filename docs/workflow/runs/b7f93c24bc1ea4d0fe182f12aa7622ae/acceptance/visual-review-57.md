# QA57 实际截图与 QA58-01 对照

候选dc667be6…，宿主执行56、原PNG/JSON完整保留在 `evidence/attempt-57/host/browser/`；不是原型或旧main当前预览。QA使用本地image viewer实际查看下面27张当前图；没有人工看过的其余图只声称全量摘要/解码核验，不声称逐图人工通过。

## 实际人工查看清单

- 当前三视口services-page各1，共3张。
- 当前390的overview/models/playground/keys/tenants/limits/usage/billing/records-page各1，共9张（加services即全部10页）。
- 当前1280与1440的models-page/keys-page，各2，共4张。
- 当前390的services-right-columns-1、models-detail、keys-editor，共3张。
- 当前390 self_hosted八张：upstream_connection_failed、upstream_rejected、verified、service-long-address、model-long-name、key-safe-identifier、normal-same-request、sse-same-request，文件共同后缀52507。

另外看了四张基线图，不计本候选27张：冻结材料 `docs/screenshots/overview.jpg`、`tenant.jpg`（原型）；历史QA58 `acceptance/evidence/attempt-58/host/visual-{1280,390}/services-page.png`（旧受控实现）。原QA58报告的30页/38张范围保持原事实，不声称本轮重看了旧全部38。

## 十页当前盘点与用户结果

| 页面/原QA编号 | 当前真实可观察结果 | 用户选择或结果、布局/术语判断 |
|---|---|---|
| 工作台；QA58-01长来源 | 标题/业务指标、21微美元、UTC当前周期、趋势、费用/风险和模型状态 | “发起调用/管理模型/查看预算”保留；表格窄屏局部滚动，业务指标不是原型虚构值。三视口summary核值与入口 |
| 模型托管；QA58-01长公共名/技术类型 | 1280/1440模型/来源/能力/状态/价格/调用延迟与动作完整；390可见来源/能力 | 已知来源中文，普通/SSE保留；self_hosted超长公共名有焦点框和完整可访问滚动区域，非ellipsis；必要技术配置仍保留 |
| 供应商与服务；QA58-01明确success英文 | 1280/1440“核验通过”，地址不拆IP/端口；390表头不纵拆单字 | 来源/启用/核验语义分开，不重复假状态；最右列图可见编辑/核验/详情/关联模型；横滚不删除关键列 |
| 调用测试；QA58原安全口径 | 390输入/发送/取消/同ID入口清楚；未使用表单“尚未调用” | 采集安全策略清空Key/消息/响应，不把空表单截图当调用成功；普通/SSE结果来自真实旅程；不新增协议独白 |
| API Keys；QA58-01安全标识 | 三视口安全hint完整单行；已启用/已撤销/已过期、期限/授权/用量/轮换入口 | 仅安全hint非秘密；窄屏表格横滚、应用与公共名不丢失；keys-editor焦点/字段/保存可见；一次秘密清理回归保留 |
| 租户与成员；QA58-01角色/邀请 | 390租户管理员/开发者/财务观察者中文，成员与邀请分区 | 用户可识别人员资格，不是原型角色切换；scope真实会话。邀请状态/角色映射及未知保留由同轮Node与summary补强 |
| 配额与预算；QA58必要维度 | 五维、UTC、已用/预留、无限/零与精确USD | 不把未知费用当0，不改硬预算含义；局部滚动可达其余列，原J04真实拒绝与恢复保留 |
| 用量分析；QA58-01长模型 | 390筛选/导出、132确认输入Token及模型行 | 数据由受控请求生成，不伪造供应商计费；完整长模型可滚动/定位，列不删；self_hosted账实另核42/14微美元 |
| 定价账单；QA58-01快照 | 390当前价与当前/历史账单金额，外部支付未接通标明 | 输入/输出价、USD每百万Token、成本独立不改；结账/更正/右列由原旅程/滚动实测，不能把截图当支付Pass |
| 调用日志；QA58-01技术字段 | 390完整Request ID可见、筛选/分页/详情入口，Token口径说明保留 | self_hosted两张同ID详情状态“成功”，3/4/7原上游Token；必要dispatch_possible等精确字段仍在，不替换成猜测语义 |

## 与冻结设计及旧缺陷的对照

现有226px侧栏、明亮主面、统一表格、标题/主要动作、右侧详情/居中编辑继续符合A工作台基线；390菜单和表格局部滚动而非根页面横滚。原型overview/tenant只是预期的信息层级，不拿演示金额/角色切换替实现。当前聚焦滚动区有明确蓝色轮廓；整行长数据需要横滚在该产品规范允许范围，不能误把截图左侧局部截取当值被删除。

QA58旧1280核验列把success拆成succ/ess，旧390地址IP/端口拆碎且“状态”两字纵排；当前同页核验中文且地址/标识有完整可访问文本，表头和关键动作不再被窄列强行纵拆。保留HTTP/Chat Completions/SSE/USD/Token、原因码、快照/配置版本等必要名称；未知状态保留原值，不制造成功。未加协议说明段或无必要布局框架，未删列换绿。

## 失败恢复与秘密

self_hosted当前390三张核验详情依次真实显示：失败/连接错误无HTTP值，失败/401，成功/200；配置执行版本1→2→3，与同一来源修正路径相符。长来源地址/长公共名截图显示聚焦局部文本容器而非省略号；完整文本与DOM可访问名称一致、keyboard scroll达成的事实在三份self_hosted回执。key-safe-identifier只显示hint，已查看图未发现完整Key/凭据/密码。

普通/SSE同请求详情各有不同真实请求ID、长公共模型/尝试/状态/原上游Token；日志详情不是调用输出，不保存完整对话。按钮和标题在很窄详情里可自然换行、仍完整可识别；当前没有证据证明不可取回或关键动作隐藏。像素检查不能证明所有辅助技术体验，未运行原生屏幕阅读器，不据aria属性冒称完整无障碍认证。

## 全量证据与原交互回归

独立解码181PNG，没有单色图片。三视口visual实际89PNG（27/24/38），30张self_hosted图及其它原故障图归档保留；134侧车全部result passed、SHA吻合、标题/实质内容锚点存在、采集前后title/content文档几何稳定、scroll delta零。181图并非全有侧车；全量清单不代替人工观察。

三视口11种modal共33段native ownership/root clip/body fixed/dialog auto/focus_inside/background scroll0原门禁通过。关闭使用原4断言(native关、root解锁、body非fixed、精确trigger focus)，用户关闭后才继续下一页；执行54失败的models-detail原路径已在当前完整序列回归，fullPage单次截图与原assert没有被换图、sleep/retry或强制聚焦放宽。当前图片可用与body-unlock实测同时成立，不依据被动close事件是否已经交付猜测根因。

QA14旧空图/后续移位的全部历史成因和R28-04旧失败仍未知；只关闭本候选实际视觉和原层交互证据缺口。当前图不是正式Owner preview登录实操，也不证明真实供应商/外部支付/身份。
