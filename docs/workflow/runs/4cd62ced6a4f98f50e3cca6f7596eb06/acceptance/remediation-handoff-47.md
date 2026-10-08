# QA47 → development 整改与复验

候选visual-editor-owner-r1；HEAD98511771871cf0951ecef716bbb55deb13e18da5；129文件指纹7e0b20ddbdea4585959365d8ad3b6121abaa91cd3f450fa2c6c21f7afc05a9ab。seq46完整make verify exit0；**No-Go源于真实设计对照Fail，不是测试驱动再失败或缺供应商权限。** 冻结material version/SHA/manifest继续沿用本Runprepare；19件核对见evidence/attempt-47/material-integrity.json。

## 1. 必修QA47-01（P1 / Open）

在已有正式UI/API补回原interactions表和PRD§7的主要管理信息，来源必须是当前服务端身份/范围的真实事实：

1. 模型托管：多对象摘要表达来源服务/类型、支持能力、当前价格与可解释的延迟/未测状态；提供上下文启停及进入调用入口或等价操作。价格来源可直接关联既有版本，未知延迟明确未测，不填演示值。
2. 供应商与服务：给关联模型数量/核验时间与当前revision适用性；候选优先级按实际模型路由含义展示或提供明确关联入口，不杜撰统一服务优先级。
3. 租户与成员：各租户负责人（按有效管理员派生）、成员、授权模型、Key与预算边界/已用真实摘要；相关管理入口保持scope，不把“调用者administrator资格”当负责人。可表格、展开摘要或直接租户详情，不要求复制原型卡片像素。
4. 工作台：补当前范围模型状态摘要及直接管理入口，保持真实趋势/风险与现有费用口径。

发现证据：`visual-review-47.md`，保留的三视口models/services/tenants/overview PNG与review。冻结依据为材料docs/interactions.md“页面及行为”、PRD§7、AC01，不是临时增New API高级需求。已提供详情中的成本四费率可以保留等价组织，不要求每种信息都机械变成表格列；但多对象日常管理主要摘要不能只靠切页或原始机器字段自行拼事实。

## 2. QA47-02（P2 / Open，单独不阻断）

同已核对Request ID在usage确认输入3/输出4，records“Token总量”未知、pending0。应明确原始上游用量与已核对计费用量，不把原始不可变未知伪装成仍无可信Token。可增加已确认累计/待确认语义或明确标签及可直达已核对值。**不得覆写原requests/attempts或原价快照，不得重复收费，不得把未知转0。** 当前账本正确性未判Fail；本项无需为了阶段通过擅自升P1。

## 3. 必须沉淀的回归

- 在原真实Go浏览器路径验证两模型/两租户/多Key的上述真实摘要及上下文动作；三视口均断言对应实际字段和值，不只列标题。未核验/未知延迟可显示Not verified而非假success。
- UI数值与同scope管理API/用量事实对账；跨租户角色负向不泄摘要，切租户迟到隔离保持。不要为收图用SQL改数据、API补用户步骤、注入前端演示角色。
- 若修P2，用同Request ID核对前未知→UI核对后，检查列表/详情/分析/CSV的原始及已确认口径；原金额/历史/终态/幂等保持。补代码契约可辅助，但原层必须真实网页复验。
- 保留所有现有make verify门禁、182Node、真实HTTP/race、五维/双scope/shared resources、迁移/backup/restore、fault unknown及普通/SSE双发送、J01–J07、visual-activity和63图采集；不要削弱capture断言。

## 4. 复验与交接

最小修改原业务UI/必要只读聚合及测试，写明变更范围与新源戳。新指纹交固定tests宿主执行原完整make verify（900s入口约定），保留首个真实失败边界，不拿seq46绿放行新版本。QA原生已知监听受限不要重试/提权；按同Run原tests节点合法入口调度。随后QA同材料/版本逐页对照和原层回归，决定QA47-01Closed及AC01Pass。现有75件证据和此前seq44/各attempt失败必须保留。

没有需求/架构澄清或权限扩大事项。平台H02/H03-runtime/H04/H07/H08及真实供应商/支付/外部身份仍Not Run，研发不补造模型/发布回执。Git提交推送草稿PR、合并和正式部署仍由Connector/协调者负责；不另建演示服务、不恢复旧Run、不重启本Run。
