# PLAN-01–12 独立文档评审

日期2026-10-09；qa执行10；候选2dfe14a3、HEAD3f9cee16。评审输入为本Run `verification/test-plan.md` r1（含执行8补记），同材料根的原设计与PLAN整改表；版本、SHA256和实际路径由 `evidence/qa-10/material-integrity.json`、候选指纹锁定。此评审是现在真实完成的独立文档评审，不追记成旧#33或tests9执行前的批准。

## 评审准则与结论

按权威PRD/AC及既有架构契约检查可观察输入、受理窗口、scope、预留/费用不变量、同ID最终事实、权限、副作用、失败与恢复；代码用于核对实现了哪些断言，不用于创造预期。**A/B受控方案合格且tests9已实际执行；C外部方案只具备正确前置/预算保守策略，当前不可执行，不是全计划/供应商Pass**。未发现规格冲突；若真实供应商行为不支持有界协议，必须记录能力限制而不是强制测试或改模型名。

| PLAN | 权威规则与独立核对 | 覆盖/责任及判定 |
|---|---|---|
| 01 | 原设计及AC可访问，交接包version与冻结设计version分开；核验61项和ZIP摘要 | Pass，本轮材料准备缺口恢复；不倒改原Run材料Blocked |
| 02 | AC10/16与架构§4：有界路径默认输出cap及可信依据不能由连接检查替代；普通/SSE真实请求观察 | 受控正式API64及显式32、拒绝65/n/tools/thinking均有断言；UI登记100为独立旅程。内部合格，DeepSeek收费/推理路径Blocked |
| 03 | AC09：受理RPM，而不是成功数或零calls拒绝 | tenant/key独立RPM2，1/2接受、3拒绝；同窗口末1ns仍拒绝，零attempt且所有维度双scope守恒。Contract/API合格；原两屏非零UI Not Run |
| 04 | AC09/10：按U+P+B准入，不以文本估计Token | B200、U7：206拒绝207接受；P200后406/407，跨窗399/400；known7转换与未知carry分别断言。合格；真实B证据缺失Blocked |
| 05 | AC11与架构§4：固定有效UTC，旧known转出、unknown carry且核对幂等 | 注入精确边界、回拨、进程/备份恢复有固定门禁；不是自然分钟网页证据。内部合格；供应商自然跨窗Not Run/前置Blocked |
| 06 | AC09/12/16：ordinary/SSE独立、known/unknown/inconsistent分层 | known7、租户fee7与cost33分开；unknown预留200/Pending1/fee0不是“免费”；operation重放409且不再dispatch。合格，供应商usage实扣Blocked |
| 07 | AC10/11/13/16：受理前拒绝/受理后无dispatch/dispatch已知与未知、超时与取消独立 | 取消首SSE行后断开且上游context完成，只HTTP证首输出；handler未dispatch直接执行生产路径单列Contract；所有请求计数保留、C只释放一次、未知不清零。内部合格；供应商失败/取消Not Run |
| 08 | AC09：真重叠竞争、有限资源、无另一scope部分扣计 | 屏障8竞争者、上游保持直到7拒绝诊断完成、只1合法受理；occupied207/C1/fee7/reserve200，final14/C0/fee14。API合格；双浏览器同步竞争Not Run |
| 09 | AC07/09：双Key共享tenant池、独立Key限额、轮换共稳定Key池、跨tenant独立 | 新竞争用例验证两Key数值和等于tenant、旧新代际仅一个剩余名额、另一tenant7/fee7；既有ResourceMatrix/SharedResources逐scope逐维度保留。合格；供应商原网页矩阵Not Run |
| 10 | AC11–13：改额/轮换/重启不改历史，价格受理快照、追加核对与迟到终态守恒 | 原治理/暂停/资源矩阵/ClockRecovery/RealProcessCrashRecovery/多attempt门禁保留；helper外层Skip不冒充执行。API/Process合格；正式部署/原供应商恢复Not Run |
| 11 | AC17：保存数值、同ID、窗口来源、白名单，不保存秘密/对话 | METERING快照含limits/usage/period/effectiveUTC/Store.now seam/dispatch；金额字符串；每次拒绝两个scope全值相等。合格；外部原少数诊断字段不被提升 |
| 12 | 原安全授权：24跨attempt、64 tokens、90秒；未知H/X不能清零 | C要求H+X+最坏supplier attempts≤24，拒绝只有证实零dispatch才预算0，单候选/最多3尝试必须按行绑定；现在没有权威H/X，禁止新生成，Blocked而非预算24可用 |

## 旅程/用例缺口的处理

1. 原材料PLAN-03–10同时要求真实网页/供应商事实，A节自动化是新增**内部层**而不是把原UI行换掉。`acceptance-matrix-qa-10.md` 分行保留未执行。测试责任专用Pipeline必须重审按实际B、权威历史和剩余额度分配的可执行计划后才能发新生成。
2. 现有真实Go网页J01–07提供正式表单/授权/调用/预算/计量/恢复的连续回归，但没有#33指定模型的真实新供应商usage，也没有两浏览器非零RPM/TPM的UI数值旅程。这是未验范围，不是已确认产品缺陷；不要求研发为了缺凭据制造改动。
3. 受控fixture创建由公共管理API，UI旅程对象由页面，control仅测试时钟/只读上游计数、故障注入。费用都是USD测试价；没有直接改数据库求绿、没有固定假供应商作为产品默认。
4. `make verify`、原12分钟deadline、race并行上限4、冷npm安装、历史red重建与三宽UI均保留。最新完整成功只支持本候选这次运行；不凭8组本地改善保证冷门禁容量。
5. 没有需要返设计/需求的规则冲突；后续若确认真实层产品故障，按原发现层首红/修后绿返development。当前批准受控交付，外部缺口交专用测试责任方，而非协调者代验。
