# QA58 当前连续用户旅程与事实边界

候选records-detail-entry-r1 / seq57 / 134件fb4066fa…，准确HEAD及材料摘要见acceptance-report-58.md。本表的UI-E2E Pass来源是当前宿主实际Go浏览器，不是QA重新监听执行，也不是API/截图拼接。三视口1280/1440/390全部同样门槛，当前continuous进程68740/68790/68850。

Start State：Go runner用正式InitConsole初始化隔离空安装并安全取得管理员密码，创建受控HTTP/SSE上游、显式测试clock和临时目录；服务/模型/租户/Key/价格/成员等均从正式页面创建，不做SQL改记录。只读管理API/受控dispatch计数用于最终事实核对。每个真实用户有独立context，角色由服务端session判定；明确的用户责任交接不伪称前端角色切换。

| Journey ID | Story/AC | Actor | Start State | UI Steps | Expected UI Result | Final Fact | Cleanup | Result |
|---|---|---|---|---|---|---|---|---|
| J01-UI | US01/02/03，AC02/03/04/06/16 | 平台管理员→正式受邀租户负责人，各自真实会话 | 空安装/无服务模型租户Key | 登录→空集合→建tenant→非法服务URL失败→修正/核验→两模型发布→租户授权/bounds/四费率→预算→邀请/接受/负责人登录→Key与诊断成员→普通/SSE→同ID详情 | 明确失败保留公开草稿/清secret、当前核验/可用目录、一次Key及实际输出 | 正式写入与请求归属/价格/已dispatch计量同值，无隐藏API补步骤 | runner关闭context/服务并删除隔离目录 | Pass |
| J02-UI | US03/04，AC03/06/16 | 租户管理员，原tenant/session | J01正式对象 | 目录→Key→一次值/调用→收窄授权→拒绝→恢复授权→显式新调用 | 明确403与可恢复授权，非自动重发 | 零dispatch拒绝；已用和历史保留 | 同上 | Pass |
| J03-UI | US04，AC06/07/08/11 | 租户管理员，原稳定应用/会话 | 已用Key/历史 | Key轮换对话框→立即或1h/24h→一次值→关闭→旧新调用→叠加409→结束旧代/撤销/期限→看共享历史 | 新旧状态/有效期/过渡边界可理解，秘密不能重取 | 同稳定Key共池及历史，期限精确；API截止前/等/后另关联行 | 同上 | Pass |
| J04-UI | US05，AC09/10/11 | 租户管理员，原scope/session | 非零调用费 | 配额查看真实已用→硬预算设已用→调用429→提高限额→显式发送→同ID | 可见拒绝原因及修正路径 | 429零上游、已用不清、新调用准确计量；五维并发API不冒充UI全分支 | 同上 | Pass |
| J05-UI | US02/04，AC04/05 | 平台/两tenant真实负责人按职责交接 | J01创建/接受的两个tenant和多Key | 授权/负责人加入→可用→暂停Alpha→其普通/SSE及多Key拒绝→Beta仍成功→恢复→当前授权调用；角色/成员退出保护 | 独立scope与正式失权，最后管理员409 | 另一tenant不受影响、退出人员立即失权但应用Key不盲撤，历史额度保持 | 同上 | Pass |
| J06-UI | US06/07，AC12/13/14/15 | 平台管理员同context/scope；财务只读反例独立会话 | 原价请求/真实unknown用量 | 查看价→发布新价→新旧同ID→当前周期拒结→显式跨月→pending拒结→日志筛同ID→详情网页核对3/4→回日志精确行打开详情→分析/CSV核对→已结束结账→+5更正→详情/导出 | 原始未知与确认3/4/pending0同时明确，结账/更正责任可操作 | 旧fee14/new21、成本独立；原attempt/快照/终态保留、差额幂等、原账单与调整周期守恒，不串tenant | 同上；测试clock不改系统/库日期 | Pass |
| J07-UI | US07，AC15/16/17 | 当前真实用户同context/授权scope | 原页面产生多页调用 | 详情跳日志保留范围→页2→同ID/模型/Key筛选与导出→切tenant→迟到旧响应→503→重试→刷新/退出重登 | 新scope不被旧响应覆盖，读取错误非0条，秘密不恢复 | 请求/CSV仅当前scope，另一tenant读改拒绝，安全存储为空 | 路由故障仅测试可控GET，结束移除 | Pass |
| S58-UI | AC01原QA47-01 | 平台管理员同context；两个正式tenant | J01–J07形成真实多对象 | 模型逐列→停/启→调用入口→服务关联详情→tenant应用Key入口→工作台模型/预算入口 | 真实摘要与上下文操作可用 | 两模型/两tenant/一服务/跨tenant五Key与API同值，实际负责人非调用者资格 | 同上 | Pass |
| F58-UI | AC08/16，原QA35/后续editor主题 | 管理员真实context | 正常签发或已用Key | 双Enter→写后丢响应→unknown→503/404/200读回；普通/SSE双发送；中断editor GET后新editor保存 | 安全未知恢复、新editor可操作、旧editor不会多写 | mutation1/secret不重放、double dispatch1、service_posts1/canceled0 | 控制故障路由有明确释放/移除，context finally关闭 | Pass |
| M58-UI | AC18旧数据可用 | 运维与迁移后的真实人员/旧应用 | 固定旧版公开UI夹具/schema1或2 | 正式CLI upgrade→native登录/identity/Key/pending历史→调用/新控制台 | 迁移后正式入口可用，schema3 | 旧身份/Key/历史/映射保持，backup/restore及失败保护由同版Go门禁补强 | Go临时目录与旧版进程归runner拥有 | Pass |
| SelfHosted-first-UI | AC02分支 | 平台 | 空服务 | 本轮未单独选self_hosted执行首次UI分支 | 无独立UI执行事实 | provider共表单链+两来源ActualHTTP仅等价补强，不改变本行 | 无新资源 | Not Run |
| Deployed-user-UI | US08/H07 | 部署后真实管理员/成员 | 正式最新Actions制品 | 尚无真实Deployment URL，未执行 | 应有/admin/和/healthz同merge版 | 不以临时runner/8793/旧main替代 | 无部署副作用 | Not Run |

## 原路径关闭而非测试驱动绿替代产品

QA47-01：summaryCell先找真实表头，再读对应cell值与已创建对象及只读API核对，native按钮实际启停/导航；不是只按实现采标题。QA47-02：requestDetail(entry='records')在确切Request ID行点击正式详情/核对，Promise.all观察真实响应/点击，保持200/ID/原始未知/确认3/4/pending0/CSV/账本不变断言；没有直接用API读detail替用户开抽屉。

seq55旧驱动误用call按钮的失败与attempt56实际helper三项红转绿继续保存；本轮seq57三视口J06才是最终网页回归。其他核心UI链没有通过SQL、后台脚本或API写入跨过用户应完成的步骤。角色被移除后的旧会话/财务只读、跨tenant负向属于不同Actor的独立反例，不挪用为同一用户会话的中间步骤。
