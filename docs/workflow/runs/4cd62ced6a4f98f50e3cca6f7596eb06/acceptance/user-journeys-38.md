# QA38 连续用户旅程与故障复验

当前127文件工作树3ee2122eadac33a67342523a082f791d8a81154cc86b7321d53de856866a15ef；HEAD98511771871cf0951ecef716bbb55deb13e18da5；design-v0.2.0。同冻结J01–J07及QA35执行契约，本轮不是重定义旅程。宿主seq37实际执行，QA独立检查脚本/断言/回执及63件归档；不冒充QA浏览器操作。

共同准备/清理：Go runner初始化独立临时目录、临时真实服务和受控HTTP/SSE，真实人员context分别登录。API仅准备入口允许的测试数据或只读最终事实，不替用户填管理步骤；一次性Key和邀请令牌从真实响应取出后显式交付/页面输入，截图先清除敏感输入。测试clock只推进时间、不编辑数据库；finally关闭context/browser/server/store并清理临时目录。没有操作用户上游/配置或生产DB；无全机进程清理审计声明。

| Journey ID | User Story/AC | Actor | Start State | UI Steps | Expected UI Result | Final Fact | Cleanup | UI-E2E Result |
|---|---|---|---|---|---|---|---|---|
| J01 | 首次可用；AC01/02/03/04/06/12/16 | 平台→被邀请管理员/开发者/财务各真实context | 无服务/模型/租户 | 保存失败→修正核验→公共模型/价格/上界/授权/预算→邀请接受→创建及分配Key→调用→同ID→最后管理员删除 | 首次可用、秘密一次、身份真实、非法草稿可恢复 | 初始对象0、无失败半对象、fee14/cost7、角色允许/禁止符合、最后管理员409 | 共同清理 | Pass三视口，受控provider |
| J02 | 授权交集；AC03/06 | 同租户管理员context | J01所建Key/两模型/历史 | 目录→收窄Key授权→403→修正→新调用→同ID | 拒绝可解释、修正可用 | 拒绝零dispatch，旧历史/快照不变 | 共同清理 | Pass三视口 |
| J03 | 轮换/退出；AC07/11 | 同管理员context/scope | 已用Key/无过渡 | 1h→新旧调用→叠加409→结束旧401；24h→新旧→结束；撤销/停用启用/期限→历史 | 状态真实、完整秘密不回取 | 精确代际期限、usage保留、历史>=8；C另立即轮换 | 共同清理；clock显式 | Pass三视口 |
| J04 | 硬预算恢复；AC09/10/11 | 平台限额+租户应用 | 非零已用费用 | 看已用→限额设为已用→429→提高→新调用→详情 | 耗尽拒绝、提高恢复不清零 | tenant_budget_exceeded无dispatch、恢复fee14 | 共同清理 | Pass三视口，不泛称所有资源维度UI路径 |
| J05 | 租户/成员生命周期；AC04/05/11 | 平台/两租户管理员/开发者 | Alpha已有Key费用，Beta无对象 | Beta创建邀请授权预算调用；Alpha多Key→暂停普通/SSE拒绝→Beta成功→恢复两Key；降权/收回诊断/移除→应用Key继续 | 暂停恢复/人员失权与应用独立 | 每个边界非零usage不变、removed scope404/目录0、Beta独立 | 共同清理 | Pass三视口 |
| J06 | 定价/账单核对；AC12/13/14 | 平台；租户调用；财务/另一租户只读 | 旧价请求+真实pending | 新价→新旧详情；当前周期拒结→clock次月/pending拒结→同ID核对→结账→+5更正→详情/导出→隔离 | 已确认与待确认分别可核对、旧金额保留 | new fee21/cost7/Tokens3+4/entries2，原revision0/快照不改，delta revision+1；原账单/原因/调整期保留 | 共同清理；clock后真实重登 | Pass三视口；实际最后执行 |
| J07 | 查询与异步恢复；AC01/15/17 | 租户管理员context；平台换scope | 正式页面27次调用形成分页 | 历史→页2→筛选→CSV→同ID跳日志保留过滤→旧读延迟换scope→503重试→刷新/退出重登 | 当前scope真实、失败非0、秘密不恢复 | 新scope行ID/导出正确、旧读不覆盖、session401、storage空 | 共同清理；读故障显式注入 | Pass三视口 |
| F08-ISSUE | AC08/QA35-01 | 当前管理员context | 无目标Key | native双保存/Enter→已提交POST201丢响应→unknown→GET503→404→同operation200 | reference和retry保留、可恢复、秘密不能读回 | 一个新对象/mutation1、无隐藏替代签发/secret replay | 共同清理；route仅注入故障 | Pass三视口 |
| F08-ROTATE | AC08/07/11 | 同管理员/已用应用Key | Key已有用量、价快照与请求历史 | native双保存→POST200丢响应→GET503/404/200→再次轮换409 | 同页面安全恢复、不叠加过渡 | revision/generation+1、transition1、原用量/历史/价快照deepEqual | 共同清理 | Pass三视口 |
| F16-DOUBLE | AC16 | 同管理员页面/应用Key | 正式有效旧代，350ms受控上游 | 普通/SSE分别dblclick+Enter→输出→同ID详情 | 忙中不双发，控件恢复 | 每种proxy POST1/准入1/尝试1/上游dispatch1/Tokens3+4/pending0 | 共同清理 | Pass三视口 |

F历史红：在当前测试服务响应中替换为明确标记的旧editor handler重建版本，真实POST201后丢响应，GET503删retry，mutations1；这是**修复后隔离重建**，不是旧开发当时实际Chromium红灯。原QA35生产函数Contract红保留。当前asset复验未使用该替换，真实green与Contract共同关闭原缺陷。

AC09十个scope/维度及8争2竞争是API-E2E独立行，不拼到UI-E2E中。AC01完整十页原型实际对照仍Not Run，虽然这些旅程及十页入口烟测通过；参见visual-review-38.md。真实供应商/身份支付、平台模型/负向Harness/发布仍Not Run。
