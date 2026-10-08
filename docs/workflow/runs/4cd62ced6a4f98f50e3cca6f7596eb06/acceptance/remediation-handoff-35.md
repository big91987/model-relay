# QA35 → development 整改交接

2026-10-08；同Run4cd62ced6a4f98f50e3cca6f7596eb06、Issue25、design-v0.2.0/archive5f51d35f…、architecture-r1。候选reconciliation-ledger-r1，HEAD98511771871cf0951ecef716bbb55deb13e18da5，124文件指纹50fb764aac5443687e376c18906b36b765f02e9b05788fa4d5ac73617cbc98a4。**No-Go**；不是门禁执行失败：宿主seq34 make verify真实exit0，三视口J01–J07均绿。

权威当前产物：acceptance-report-35.md、acceptance-matrix-35.md、user-journeys-35.md；evidence/attempt-35保存本次来源、真实失败、独立绿灯、宿主产物/索引。旧QA22报告和历次失败仍保留，不覆盖为新结论。

## 1. 必修：QA35-01 P1，安全读回第一次失败后没有恢复入口

**用户路径**：管理员签发或轮换Key，后端已经提交，但浏览器丢响应；界面进入“写入结果未知”，保存隐藏。点“安全读回”时发生暂时读取失败，页面提示“请保留操作标识并稍后重试读回”，实际既看不到操作标识，又没有可再次点击的读回按钮。

**定位与证据**：`internal/relay/web/workspace.js:94`通用edit的readback catch使用`output.textContent=...`，删除此前append的button；operation只保存在该callback的closure。实际生产函数的独立Contract复现`node --test docs/workflow/runs/4cd62ced6a4f98f50e3cca6f7596eb06/acceptance/evidence/attempt-35/readback-recovery.test.mjs` exit1，readback-red.log保留mutation1/read1/retry=false/operation-visible=false。DOM替身明确只模拟textContent子节点替换，不是真实Chromium；原门禁只有首次读回成功，故未发现。

**验收语句**：

1. 写入结果未知时保留安全operation引用、原非敏感草稿和明确只读重试入口；不显示秘密、不自动重新提交、不引导重复签发/计费。临时读回失败后仍可在同页面恢复。
2. 首次GET operation丢响应或503/401/404分别说明当前事实/身份/未知边界；不能误报操作未提交，也不能通过重放写请求恢复秘密。
3. 成功读回只给安全元数据及后续明确生命周期操作。丢失一次秘密时通过用户可见正式流程处理，不能测试内部API偷偷创建替代Key/改代际。
4. 重试忙状态防双读/迟到：退出/租户切换/旧editor关闭后响应不恢复秘密/旧scope；至少沿用既有epoch与secret cleanup约束，不做无关重写。

**必须新增自动化**：先将当前缺陷在真实Go+Chromium层红灯复现（保留日志与安全截图），再修生产函数；把通用handler故障回归放项目既有Node集，并将下列真实UI路径加入原make verify。QA只新增了验证资产，没有修业务。

## 2. QA22-01残余强制覆盖（不能用新增主链标签代替）

| 用例 | 当前缺口 | 下轮实际执行与最终事实 |
|---|---|---|
| AC08签发丢响应→读回失败→恢复 | console.mjs:137只首次成功读回 | 真实POST201后route.abort；第一次operations读503/abort，按钮保留；恢复后同operation读成功，对象数量1，隐藏保存、无完整secret重放；提交/读回实际状态分开记录 |
| AC08轮换丢响应及写入双提交 | journeys.mjs的rotation均直接收到成功响应 | 有已用Key页面轮换，已提交后丢响应；快速重复保存/Enter，只一条operation和正确代际。读回失败后页面重试确认；原历史/用量/价快照保留，过渡不叠加；不能仅API测试 |
| AC16新控制台快速重复普通/SSE发送 | call helper一次发送后等待结束；静态state.call guard不是执行证据 | 用户级双击/Enter、上游受控延迟；普通/SSE各确认一次准入/dispatch、一个同ID及计量。取消仍显示费用可能待确认、开始输出不重放、按钮恢复 |
| AC09两层各维度与多Key/代际争用 | ResourceDimensions只对tenant五维逐个置0，Key只Calls=0；8并发为同一Key | 真实HTTP针对Key RPM/TPM/C/费用各自独立0/null/正有限边界命中与恢复；两个Key共享tenant、旧新代共享Key，用barrier并发争抢N名额；拒绝无另一scope部分扣计，修改不清零，另一tenant继续可用 |

这些是冻结AC/verification-design已有要求，不扩大产品范围，不要求真实供应商凭据，不替平台做H负向恢复。若已有等价当前测试，交接给出具体case/断言/同版完整运行证据供QA核对，而非只报测试名或状态机存在。

## 3. 已闭合与不能退回的门槛

- J01–J07主链1280/1440/390本轮已通过，不能再重报“全部未执行”；QA22-02 P2已Closed，保留真实状态/授权/期限/预算摘要与J03实际行断言。
- J06累计fee21/cost7/Tokens3+4、entries2、原revision0及快照不变、delta revision=before+1；G原bill19/60及delta fee5/cost16继续有效。禁止改原账本满足驱动，不删幂等/拒绝断言。
- 原健康方法预期红灯/恢复、真实HTTP/SSE/取消/重启/backup/迁移schema1/2→3、Governance/PeopleScope/PauseUsage、旧native/路由、签发未知、13Tab/Escape/返回点及键盘业务创建门槛不减少。
- Node165项与本轮七项定向Go race为当前实际绿色，不以测试总数替代新增用例；可合理扩展但不得删失败证据。

## 4. 同版复验与再次交接

1. 同Run沿用实际manifest路径和19文件原材料；仅修复必要业务/测试/对应文档，不改冻结规则、旧DB、平台检查点或其他Run。
2. 保存本次实际红灯及修复后绿灯；新before/after来源戳必须覆盖产品/测试/锁/契约，源码/测试变化后不能沿用50fb…的seq34放行。
3. 使用原完整`make verify`宿主Connector，保留真实exit、完整未截断日志和实际运行版本。原生监听限制不要求提权、不重复已知必败命令、不删make verify中的健康/HTTP/浏览器。
4. 新版本三视口连续7项检查点及新增故障场景均需真实通过；保存脱敏JSON、写入前后/unknown/读回失败/恢复截图与安全只读事实。截图等关键对象都等待真实UI落定，不在异步drawer打开前截图充当详情证据。
5. 交接独立QA携新指纹、完整日志索引、失败/修复证据、具体测试位置及仍Not Run项；不自报QA通过。保留外部供应商/身份/支付、实际模型和PR/Actions/Deployment/health固定URL各层边界。

## 5. 路由与授权

没有需求/设计矛盾或用户澄清，不回requirements/design；development负责通用editor安全恢复及上述内部用例。H02/实际模型/H04/H07/H08仍由平台/协调者核对，不以缺GitHub凭据阻断产品修复。新P1关闭且强制门禁/AC完整后再进入report；本QA不提交、推送、开PR、合并、部署或重新启动Run。
