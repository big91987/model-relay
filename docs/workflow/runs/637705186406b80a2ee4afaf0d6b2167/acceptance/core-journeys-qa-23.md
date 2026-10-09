# 当前联合候选连续网页旅程

QA执行23；实际执行者为宿主tests22，真实Go服务＋Playwright/Chromium，不是QA本人网页实操。HEAD8917f942＋main6929376、158文件6e359667候选。每宽独立临时安装，同角色保持context/page/session；跨角色通过邀请、接受、真实登录显式交接。编号J01–07来自冻结interactions，不与原#33扩展J09–12混同。

用户业务写入由原页面操作；只读API用于最终同ID/账/初始空状态采证。Go测试control提供受控上游故障、计数及显式测试钟，不代替UI创建或调用；不使用SQL补业务步骤。J03期限、J06跨月为测试钟，非自然时间。故障/未知提交/迟到使用明确route注入；旧编辑器红是隔离重建，不是当前asset。秘密使用后清除、截图不展示秘密。首次自托管另一个隔离安装，初始upstream/model/tenant全部0，business_writes=native_user_actions_only。

| Journey ID | User Story/AC | Actor | Start State | UI Steps | Expected UI Result | Final Fact | Cleanup | Result |
|---|---|---|---|---|---|---|---|---|
| J01-UI | AC02/03/04/06/10/16首次使用 | PA→TA/DA/财务独立会话 | 空服务/模型/Key | 创建tenant→服务错误草稿/纠正→核验→模型发布/授权→bounds/四费率/预算→邀请接受登录→创建Key→普通调用→详情 | 错误不伪成功、秘密一次性、首次调用完成 | 当前revision适用；同ID/价快照/计量一致，角色边界保留 | contexts/临时服务和目录由宿主清理 | Pass |
| J02-UI | AC03/16日常失败恢复 | TA/分配的DA | J01有效对象 | 收窄Key授权→调用403→诊断→网页修正授权→显式新调用 | 拒绝原因可读；不后台补发 | 拒绝零dispatch，新请求成功 | 同上 | Pass |
| J03-UI | AC06/07/11轮换退出 | TA | 非零已用Key | 1h过渡轮换→新秘密关闭→新旧调用→叠加拒绝→提前撤销/到期 | 关闭不回显，历史共享 | 稳定池保留；旧值撤销/期限拒绝；测试钟不回拨复活 | 同上；披露测试钟 | Pass |
| J04-UI | AC09/10/11预算与恢复 | PA调额/TA调用 | 已确认费非零 | 读已用→硬预算耗尽→429→详情→提高额度→新请求 | 提高不清历史；明确剩余与拒绝 | 拒绝零dispatch/双scope不变，恢复受控Fee14 | 同上 | Pass |
| J05-UI | AC04/05/11/17生命周期/隔离 | PA及两tenant真实会话 | A多个Key，B新建 | 创建授权B→暂停A→普通/SSE/另Key拒绝→B调用→A恢复→成员降权/移除→范围读回 | A新调用停、B独立，成员范围立即收回 | 暂停计量不变；恢复成功；跨tenant404、应用Key不私撤 | 同上 | Pass |
| J06-UI | AC12–14改价/核对/账单 | PA/TA/财务 | 旧价调用和本期账 | 新价→新旧请求对照→unknown→当前/pending结账拒绝→测试月界→页面核对→结账→更正→导出 | 未知单列、更正保留原账/原因/周期 | 新Fee21/旧快照不变；核对幂等；closed bill保持 | 同上；测试钟跨月 | Pass |
| J07-UI | AC08/15/17导航/故障/恢复 | 同角色同scope | 页面创建>25请求 | 历史第二页→筛选CSV→详情跳日志→切tenant→旧响应晚到→读故障重试→刷新/退出登录 | 不串scope、错误非空、秘密不恢复 | CSV无另一tenant；注销401，旧scope不覆盖 | 同上；显式route注入 | Pass |
| SELF-UI | PR35首次自托管＋AC02/03/04/06/16 | PA→TA/DA各真实会话 | 无upstream/model/tenant | 不可达→拒绝→纠正核验→长地址服务→成员/角色→Key安全标识→普通→同ID账→SSE→同ID账→用量→长模型名 | 接入失败可恢复；长技术值局部滚动/键盘可读；输出与诊断同ID | 两模式均input3/output4/Fee21/cost7；三宽各12步骤 | 独立安装由宿主清理 | Pass |
| FAULT-UI | AC08/16未知提交/双击 | TA/DA | 当前编辑器 | 签发/轮换提交丢响应→503/404/200读回；普通/SSE双击 | 未知不重复写，秘密不重放，禁止叠过渡 | 每mutation1、proxy/admit/attempt/dispatch各1；3+4 tokens | 同上；故障拦截恢复 | Pass |
| ORIG33-UI | 原非零RPM/TPM/自然窗/并发/真实成功与费用 | 原DA两会话/1280及390 | 当前供应商安全/预算未具备 | 未执行 | 须由原合法专用Pipeline网页连续操作 | 无新真实生成usage/费用/自然窗事实 | 未产生外部负载 | Not Run |

## 当前证据定位

全部路径相对本Run，目录 `acceptance/evidence/qa-23/host-tests22/browser/`：

- continuous-1280-43465.json、continuous-1440-43504.json、continuous-390-43542.json：各有J01/J02/J03/J04/J05/J07/J06的完整checkpoints。
- self-hosted-1280-43494.json、self-hosted-1440-43522.json、self-hosted-390-43563.json：独立空安装、原生写入、开始结束时间及12项最终事实；对应截图有integrity JSON。
- fault-1280.json、fault-1440.json、fault-390.json与editor-interruption三宽：丢响应恢复/双击/旧GET晚到和原编辑器首红敏感性证据。
- visual-1280/1440/390的review.json与页面PNG：三宽原几何、模态关闭、键盘和技术值断言。QA实际抽看390同请求和limits截图；没有声称人工逐张视觉评审全部PNG。
- 全部文件SHA256在 `acceptance/evidence/qa-23/archive-index.json`；原runner完整输出另在host-tests22/verify.log。

当前正常网页通过不覆盖“网页设置非零RPM/TPM→自然窗口→双会话竞争→真实供应商计量”整链；该未执行层保留在矩阵。静态原型、API构造的请求与截图均不单独替代UI-E2E。
