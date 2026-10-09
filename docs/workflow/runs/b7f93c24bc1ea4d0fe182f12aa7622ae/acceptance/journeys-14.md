# 执行14 核心连续网页旅程

所有UI-E2E结果是同候选正式宿主执行13，不冒称QA独立动态执行。每个角色有真实登录和自己的浏览器会话；跨角色步骤是显式人员责任交接，范围不串租户。正常/失败/退出均在现有脚本连续操作；准备为独立Go init、临时SQLite/端口和受控HTTP，上游失败/时钟控制不写业务DB；清理为自有Go/browser finally及临时候选移除。

| Journey ID | User Story/AC | Actor | Start State | UI Steps | Expected UI Result | Final Fact | Cleanup | Result |
|---|---|---|---|---|---|---|---|---|
| SH-SETUP | 任务3/AC02/03/04/06 | platform admin→受邀tenant admin/developer | 每视口独立新安装，0来源/模型/租户 | admin页面建已有自托管；核验并纠正；建模型、租户、授权、价格；网页邀请/接受；owner网页创建Key与diagnostic成员 | 来源/授权/角色与安全标识可读；秘密仅一次 | self_hosted真实保存、当前revision核验、授权交集与Key身份 | 独立实例关闭/目录删除 | Pass |
| SH-FAIL | 任务3失败恢复/AC02 | 同admin、同来源、同会话 | 刚创建不可达地址 | 点核验→看失败原因；页面编辑正确地址+错误凭据→看401；再次编辑纠正→点核验 | 核验失败且可定位原code；编辑失效旧核验；核验通过 | upstream_connection_failed/null、upstream_rejected/401、success/200；revision1/2/3 | 同上 | Pass |
| SH-NORMAL | 任务3/AC15/16 | owner、同tenant、同会话 | SH-SETUP创建的Key/model | 调用测试填Key/model/message；点发送；同请求详情；同ID日志 | 普通完成、成功状态、费用事实同ID | confirmed3/4、fee21、cost7、pending0，额度增加21 | 清敏感内容/browser finally | Pass |
| SH-SSE | 任务3/AC15/16 | developer、同tenant、自己的会话 | 诊断分配的同应用Key | 填表勾SSE发送→DONE→同请求详情→同ID日志 | SSE完成，不冒称免收费 | 同Key/tenant/model、3/4/21/7/pending0，价格快照/追加明细一致 | 同上 | Pass |
| SH-USAGE | 任务3/AC12/14/15 | owner、原会话原scope | 两次真实调用后 | 用量页核模型费用；GET只读对账；模型页完整名可访问 | 已确认租户费0.000042USD | aggregate6/8、42/14、pending0 | 同上 | Pass |
| J01 | 冻结首次供应商链 | admin及真实受邀角色 | 独立安装 | 供应商接入/失败修正/核验→模型/价格/授权→成员/Key→调用/详情 | 页面步骤真实保存，不API补写 | 同ID与计量核对 | 隔离实例finally | Pass |
| J02 | 授权拒绝修正 | tenant admin、原scope | 已授权Key | 创建→调用→收窄→拒绝→恢复授权→显式重发 | 拒绝可理解、修正再调用 | 拒绝不dispatch、修正不会自动重放 | 同上 | Pass |
| J03 | Key退出/轮换 | tenant admin | 已有Key历史/用量 | 轮换/一次值/过渡/到期/提前撤销及历史 | 秘密关闭后不可取回、状态准确 | 共享池、旧值失效、拒绝叠加 | 同上 | Pass |
| J04 | 硬预算/纠正 | tenant admin | 已有用量 | 耗尽→429→调限→再次发送/同ID核对 | 不清已用、显式再发 | 零dispatch拒绝、预留/结算一致 | 同上 | Pass |
| J05 | 租户退出/恢复 | admin及两租户角色 | 两租户独立 | 创建/负责人/授权→暂停→全部Key拒绝→恢复 | 另一租户不受影响 | 历史与授权保持、按当前授权可用 | 同上 | Pass |
| J06 | 账单/价格/费用 | admin/tenant admin/finance | 跨周期隔离fixture | 改价→旧新请求→pending/当前拒结→跨期结账→追加更正→导出 | 金额、原价、原因及调整期不混淆 | 同快照/尝试/费用守恒，closed bill不改 | 时钟仅隔离控制、finally | Pass |
| J07 | 导航/读取失败 | 原角色/scope | 详情/多页 | 详情跳日志→页2→换scope→旧响应晚到→错误重试→刷新 | 旧scope不覆盖，失败不是空数据，安全刷新 | 秘密不恢复、读回所有权正确 | 同上 | Pass |
| QA-NATIVE | 独立动态原生操作 | QA | 无可用动态工具 | 未执行 | 不用静态页面/源码替代 | 没有本条结果 | 不操作环境 | Not Run |

三视口SH各12个记录checkpoint（含2个只读ledger事实）；这不是“12次页面点击”计数，完整业务步骤以上述脚本与回执为准。视觉结果另看matrix L02/L03/S07，不能借行为Pass覆盖截图Fail。
