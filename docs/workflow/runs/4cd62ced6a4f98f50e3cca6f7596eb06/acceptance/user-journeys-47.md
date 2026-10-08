# QA47 当前连续用户旅程与最终事实

同指纹seq46；三个连续回执保存于evidence/attempt-47/host-seq46/continuous-1280-34454.json、continuous-1440-34476.json、continuous-390-34501.json，checkpoints都为J01/J02/J03/J04/J05/J07/J06。场景契约沿用QA38原设计，当前不是复制旧green：seq46重新执行所有这些链，QA核对新回执与当前journeys.mjs/runner及源戳。

UI-E2E由宿主实际Chromium操作，不是QA新开浏览器。测试初始仅正式CLI创建admin/空schema3、受控HTTP服务和显式UTCclock；服务、模型、租户、邀请、Key、价、权限、预算、27次查询记录等用户中间步骤由原正式网页完成。读取API与账本只用于Final Fact，不替网页提交。角色交接使用各自真实身份context，单Actor链保持同scope/session。J06最后执行；跨月后为visual补1次当前周期真实调用，不改DB或闭合旧账金额。

| Journey ID | Story/AC | Actor | Start State | UI Steps | Expected UI Result | Final Fact | Cleanup | Result |
|---|---|---|---|---|---|---|---|---|
| J01 UI-E2E | US01–03 / AC02–06 | 平台admin→受邀租户管理员/开发者 | 正式空安装，无目标模型/tenant/Key | 登录→非法服务/修正→核验→两模型/价→tenant/grant/budget→邀请接受登录→Key→普通/SSE→同ID | 首次可用，非法草稿保留，秘密一次，角色由后端判定 | 无半对象，fee14/cost7，允许禁止正确，最后管理员409 | 隔离目录/contexts/server清理 | Pass，三视口；受控provider |
| J02 UI-E2E | US04 / AC03/06 | 租户管理员同context | J01应用Key/两模型 | 收窄→403→修正→显式新调用→同ID | 拒绝可解释与修正可用 | 拒绝零dispatch、旧历史/快照保留 | 同上 | Pass，三视口 |
| J03 UI-E2E | US04 / AC07/11 | 同管理员/旧新应用Key | 非零用量/无过渡 | 1h新旧→叠加409→提前结束401；24h新旧/结束→启停撤销期限→历史 | 不叠加，准确生命周期、秘密不可回取 | 两代共池，历史>=8；立即分支console、精确截止由独立API证据 | 同上；clock显式测试用 | Pass，三视口；不拼API到未跑UI精确时间边界 |
| J04 UI-E2E | US05 / AC09–11 | 平台/租户应用 | 已用费非0 | 看已用→硬预算设已用→429→提高→显式新调用→详情 | 耗尽拒绝，提高恢复不清零 | tenant_budget_exceeded零dispatch、恢复fee14 | 同上 | Pass，三视口；五维HTTP矩阵另计 |
| J05 UI-E2E | US02/04 / AC04/05/11 | 平台/两租户管理员/开发者 | Alpha应用已用，Beta无目标对象 | Beta创建邀请授权预算调用；Alpha暂停多Key普通/SSE→Beta可用→恢复；降权/收回/移除→应用Key继续 | 成员失权不盲撤应用；scope隔离 | 原用量守恒，removed范围404/目录0，Beta独立 | 同上 | Pass，三视口 |
| J06 UI-E2E | US06 / AC12–14 | 平台/租户调用/财务与另一tenant只读 | 旧价调用+真实pending | 新价→新旧详情→当前周期拒结→clock次月重登/pending拒结→同ID核对→结账→+5更正→导出 | 旧金额/快照/原因/调整周期可追溯 | fee21/cost7、确认输入3输出4、entries2、原revision0，delta revision+1，原bill保留 | 同上；clock不是生产自然月份 | Pass，三视口；同已核对请求日志解释另QA47-02Fail |
| J07 UI-E2E | US07 / AC01/15/17 | 租户管理员context/平台换scope | 正式页面27调用形成分页 | 页2→过滤→CSV→同ID保过滤→旧读晚到/换scope→503重试→刷新/退出重登 | 当前scope真数据，错误非空0，不恢复秘密 | 行ID/导出scope正确，旧读不覆盖，session401/storage空 | 同上；读故障仅route显式注入 | Pass，三视口；不证明全部摘要/语义 |
| F08-ISSUE UI-E2E | AC08 / QA35-01 | 管理员同context | 无目标Key | 双击/Enter保存→已提交201丢响应→unknown→503/404/200安全读回 | 可恢复且不重签、不回显秘密 | mutations1/reads3/secret_replayed=false | 同上 | Pass，三视口 |
| F08-ROTATE UI-E2E | AC07/08/11 | 管理员同scope | 已用Key | 保存200丢响应→503/404/200→再轮换409 | 不叠过渡，历史价用量保留 | generation/revision+1，transition1，mutation1 | 同上 | Pass，三视口 |
| F16-DOUBLE UI-E2E | AC16 | 同管理员/应用Key | 350ms受控有效上游 | 普通/SSE各双击+Enter→结果→同ID详情 | busy仅一次，结束控件恢复 | 每种proxy/admission/attempt/dispatch1，Tokens3+4 | 同上 | Pass，三视口 |

历史旧editor红为修复后隔离重建expected_failure_reproduced，不称开发原时浏览器红。自托管首次UI独立来源、实际供应商、真实部署和外部身份支付仍Not Run。

核心操作链通过不等于全部产品Go：V47从真实页面检出主要摘要缺项，AC01Fail/P1；J06确认数据已存在而日志无口径“Token总量未知”是单独P2，不覆写原链账本通过结论。
