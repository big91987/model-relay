# 连续核心网页旅程（执行57）

基线：冻结材料 `docs/interactions.md` J01–J07及Issue32首次self_hosted。执行前原合同在材料、`tests/browser/journeys.mjs` 与同Run任务中已存在，本文件是QA结果映射，不临时重写预期。执行者宿主56；各宽度1280/1440/390均实测。QA57核验代码、源戳、原回执与实际图，未另跑浏览器。结果证据根为 `evidence/attempt-57/host/browser/`。

## J01–J07

| Journey ID | User Story/AC | Actor | Start State | UI Steps | Expected UI Result | Final Fact | Cleanup | Result |
|---|---|---|---|---|---|---|---|---|
| J01-UI-E2E | US01/02；AC02/03/06/16 | 平台管理员接入、租户管理员使用；显式角色会话 | 原供应商模式新临时安装 | 服务→核验→模型→价格/租户授权→预算→成员邀请接受→应用Key→调用→详情 | 已核验、授权、一次秘密、同ID调用结果 | 真实受控HTTP dispatch及请求归属/用量 | harness关闭服务/删除临时目录，browser finally关闭上下文 | Pass |
| J02-UI-E2E | US03/04；AC03/06/16 | 原租户管理员、同会话 | 原授权目录与应用 | 新Key→一次值→调用→收窄→拒绝→修正→显式调用 | 拒绝可定位，修正可继续 | 拒绝零dispatch、不自动重放 | 同上隔离清理 | Pass |
| J03-UI-E2E | US04；AC07/08 | 原租户管理员、同scope | 原Key | 详情→轮换选择→新值→关闭→旧值状态→过渡提前撤销 | 新秘密不恢复、不叠加过渡 | 新旧共享用量与历史 | 同上 | Pass |
| J04-UI-E2E | US05；AC09/10/11 | 原租户管理员 | 已用预算 | 查看→耗尽429→提高限额→新发送→日志 | 调额不清已用、拒绝后可恢复 | 零dispatch拒绝与同ID费用 | 同上；测试时钟明确专用 | Pass |
| J05-UI-E2E | US02/04；AC04/05 | 平台/租户管理员，明确作用域 | 两独立租户 | 创建/负责人/授权→暂停→所有所属Key拒绝→恢复 | 状态与资格正确、他租户不受影响 | 当前授权拒绝/恢复与历史保留 | 同上 | Pass |
| J06-UI-E2E | US06；AC12/13/14 | 原平台管理员 | 旧价与多种请求终态 | 改价→旧/新快照→确认/待确认→当前期拒结→跨月→结账→更正→导出 | 原金额/原因/周期保持 | 账单明细汇总、成本/租户费分离 | 同上；control仅推动专用测试钟 | Pass |
| J07-UI-E2E | US07；AC15/17 | 原有权角色与当前scope | 多页日志 | 详情跳日志→分页→切scope→迟到响应→失败重试→刷新 | 旧数据不覆盖、错误非空态、秘密不恢复 | 读/写scope和安全状态 | 同上 | Pass |

三份同版continuous回执：`continuous-1280-52412.json`、`continuous-1440-52442.json`、`continuous-390-52483.json`；各checkpoints含完整J01–07，无skip。每条内部角色动作沿用同一角色页面/浏览器会话/作用域；J01/J05必要的跨角色动作明确分上下文，不把一人伪装多个角色。既有两来源routing、普通/SSE取消/恢复、权限与故障另有独立浏览器/Go门禁，不能用它们补self_hosted创建中间步骤。

## 首次 self_hosted 的执行契约

| Journey ID | User Story/AC | Actor | Start State | UI Steps | Expected UI Result | Final Fact | Cleanup | Result |
|---|---|---|---|---|---|---|---|---|
| SELF-FIRST-UI-E2E | Issue32-3；AC02/03/04/06/16 | 平台管理员admin创建与授权；网页邀请加入的租户管理员owner普通调用；网页邀请加入的developer SSE调用 | 每视口独立新SQLite；UI只读初始upstreams/models/tenants全0；只有初始化管理员/bootstrap及隔离HTTP/control夹具 | 登录→服务表单选self_hosted→错误地址核验→网页纠正→错误凭据401→网页纠正→核验通过→模型表单关联→创建租户→模型租户授权→价格表单→邀请/接受两角色→owner创建应用Key并分配developer→owner普通与developer SSE→各自同ID详情/日志→owner用量和长名称 | 来源“已有自托管”；失败原因可见；配置变更旧核验失效；新核验通过；模型/租户/Key正确；调用成功、同ID日志可定位 | 源kind为self_hosted；revision递进；真实HTTP普通/SSE；每次确认3/4Token、费21/成本7微美元、pending0；两次合计6/8、费42/成本14 | harness defer关监听/control/upstream/store与临时目录；browser finally关闭全部角色上下文；宿主owned_group_absent/candidate_removed | Pass |
| SELF-FAIL-RECOVERY-UI-E2E | Issue32必要失败与恢复 | 同一平台管理员页面与会话 | 刚创建来源/无模型 | /unreachable核验→upstream_connection_failed→编辑基础地址/错凭据→401 upstream_rejected→编辑凭据→200核验通过 | 原因码精确、失败不伪造通过；修正后当前配置适用且可继续关联 | 原核验状态API保持failed/failed/success，页面中文；修改后先“未知/配置已变更” | 同上 | Pass |
| SELF-CALL-UI-E2E | AC16/17 | owner普通/已分配developer流式，各自会话同租户 | 上述原生Key与授权已完成 | 调用表单填Key/模型/消息/流式→发送→同ID日志→读回尝试/快照 | 普通响应及SSE data/[DONE]，成功可追踪，无秘密留图 | 请求tenant/key/alias/价格快照及charge_entries对账 | 采图前清秘密/消息；隔离清理 | Pass |
| SELF-LEDGER-READ | AC12/13/15；费用Token不变 | owner同scope | 上述两次真实调用 | 查看日志/用量；只读API增强核对 | 成本与租户费、已确认与pending明确 | 每次21/7与预算增量21、总42/14、6/8、pending0 | 不改SQL/不补业务写入 | Pass |

三份独立首次回执：`self-hosted-1280-52430.json`、`self-hosted-1440-52473.json`、`self-hosted-390-52507.json`；每份12个记录步骤、10张安全图。模型创建/租户/授权/价/邀请/Key是源码内网页动作且完整门禁实际完成，12记录点不代表只有12个用户点击。普通与SSE同ID最终核对在当前 `tests/browser/journeys.mjs:386` 后的强断言中，回执有ledger事实；不杜撰一份未保存的逐请求HAR。

准备与诊断边界：Go使用Open/Bootstrap建立全新数据库和初始管理员，起HTTP供应商/control监听；浏览器只读API检查空计数/已保存对象/用量。save读取浏览器实际响应，表单fill/select/click生成写入。内部邀请接受也由网页完成。control只控制隔离时间/dispatch查询，不代替来源/模型/租户/Key写入；没有往数据库写记录让旅程通过。/unreachable为隔离上游显式模拟连接失败，不是外部真实供应商故障。

## 退出、故障与未执行

已有J02–05与faults覆盖收窄/撤销/暂停/预算、丢响应读回、双普通/SSE只一次；routing覆盖取消/断流后不重放与未知费用。三视口编辑中断都service_posts1/canceled_editor_posts0。11种modal每宽度原生打开/滚动/关闭/精确焦点完成，原四项关闭断言不放宽，当前body锁同步释放。

QA原生独立动态复跑、真实供应商/支付/身份、Owner preview已登录操作均各自Not Run，见矩阵。此文不为未验证真实部署操作手册；内部受控HTTP路径可复用固定make verify回归，但不能包装成已验生产操作说明。
