# QA35 核心旅程与主链复验

同冻结design-v0.2.0/interactions.md J01–J07，不改故事。证据是宿主seq34真实Go/SQLite/Chromium，不是QA原生再次执行或静态原型；来源124文件指纹50fb764aac5443687e376c18906b36b765f02e9b05788fa4d5ac73617cbc98a4。每视口同一独立安装，平台/租户/开发者/财务各自真实登录context；每条人员子旅程连续执行，不使用前端角色切换。职责交付处换真实Actor，不谎称跨人员全部是同角色。

## 准备和清理边界

Start State只用正式Go InitConsole空安装：服务/模型/租户数组0，无预制业务完成态。独立受控上游和测试时钟明确标注；clock按UTC推进而非修改SQLite或系统时间。页面实际提交建立对象/授权/限额/价格/成员/Key/账单；没有API或SQL补中间业务步骤。只读API用于身份/计量/导出对账，响应读取一次秘密/邀请token仅承担明确的安全手工交付和测试调用，不作为隐藏管理动作。J07 route只延迟真实读响应或注入明确503，不伪造成功对象。

Cleanup统一由Go runner defer关闭服务、store、upstream、control及删除临时目录，Node finally关闭真实browser/context；不是共享用户数据。没有全机进程审计。每视口JSON和五类脱敏截图已复制到evidence/attempt-35/host-seq34/，原test-results仍保留，artifact-index.json保存摘要/时间/PID。

| Journey ID | User Story/AC | Actor | Start State | UI Steps（实际主链） | Expected UI Result | Final Fact | Cleanup | UI-E2E Result |
|---|---|---|---|---|---|---|---|---|
| J01 | US01/02/03；AC01/02/04/06/16 | 平台→真实接受邀请的租户管理员；开发者/财务各自登录 | 全新空安装 | 键盘租户校验错误→创建；服务非法URL→修正→核验；发布两模型→授权/上界/价格/预算；邀请接受；租户签发并分配诊断成员；调用→同ID；最后管理员移除409 | 草稿保留但秘密清除、核验成功、完整秘密一次、邀请身份真实、可用结果 | 初始数组0、失败无半对象、tenant/Key归属正确、fee14/cost7、developer有诊断Key但无创建按钮、finance只读 | 统一 | Pass，1280/1440/390；受控provider来源，真实供应商另Not Run |
| J02 | US03/04；AC03/06 | 租户管理员同context/scope | 获两模型/有旧请求 | 看目录→编辑Key收窄→模型403→修正→新调用→同ID→安全截图 | 授权交集拒绝可解释并可恢复 | 拒绝零dispatch、旧快照与旧账项不改 | 统一 | Pass三视口；Key首次创建在前置J01同context完成 |
| J03 | US04；AC07/11/列表状态 | 租户管理员 | 有费用与历史Key、无过渡 | 1h→一次秘密→旧新分别调用→禁止叠加→提前结束旧401；24h→旧新调用→结束；撤销另Key401；主Key停用401/启用；到期Key推进60秒→过期401；同Key历史 | 过渡/停用/启用/过期/撤销状态准确，完整秘密不回取 | rotation前后usage相等、generation期限精确3600/86400、Key历史>=8、旧值拒绝；C另做立即轮换 | 统一；明确测试clock | Pass三视口；真实轮换丢响应/读回失败未执行，不能覆盖AC08 |
| J04 | US05；AC09/10/11 | 平台修改限额+租户应用 | 已有非零费用 | 看已用→把硬预算设为已用→调用429→提高→新调用→同ID | 费用拒绝原因准确，提高后可用 | tenant_budget_exceeded零dispatch、已用保持、恢复请求fee14 | 统一 | Pass三视口；不是只测Calls=3后的429，也不是所有资源组合 |
| J05 | US02/05；AC04/05/11 | 平台/两租户管理员/开发者 | Alpha已有Key/费用，Beta无对象 | Beta创建→邀请/授权/预算→调用；Alpha备用Key调用；暂停→普通/SSE/备用Key拒绝→Beta成功→恢复两Key成功；开发者改财务→收回诊断→移除→旧scope消失→应用Key仍成功 | 暂停/恢复及成员权限显示真实，另一租户独立 | 每个暂停拒绝与恢复边界usage守恒；removed scope404/tenants0/select options0、应用Key不盲撤销 | 统一 | Pass三视口；API核对与页面写操作分开记录 |
| J06 | US06；AC12/13/14 | 平台；租户调用；财务和另租户只读 | 已有旧价请求、真实未知调用形成pending | 改价→新调用→旧新详情；当前周期拒结；advance至次月/pending拒结；同ID页面核对→结账→+5更正→详情/导出；财务无写、另一租户账单404 | 旧快照保留、未知待确认→可核对、原账单和调整期保留 | 新fee21/cost7；核对entries2/累计Tokens3+4、pending/reserve0、原revision0与快照不改；correction原因/amount5/新调整期，original_fee不变 | 统一；clock只推进、无DB改账 | Pass三视口；J06实际最后执行，次月需重新真实登录，不假装会话不失效 |
| J07 | US07；AC01/15/17 | 租户管理员同context；平台scope切换 | 27次正式页面调用形成超页历史 | Key历史→页2→筛选→CSV；同ID详情跳日志保持过滤；平台页2→延迟旧Alpha读→切Beta→释放旧响应；503→重试；租户刷新/退出/重登 | 当前scope不被迟到覆盖、失败不是0、秘密不恢复 | row数量及ID来自Beta真实响应、导出无另一租户、401 session退出、存储为空 | 统一；读取故障显式注入 | Pass三视口；不是旧native页2或VM测试替代 |

## 不得由主链绿色推定的分支

| 独立场景 | 要求 | 当前层/结果 | 下轮执行契约 |
|---|---|---|---|
| AC08 首次安全读回失败 | 保留可重试入口和安全operation标识，不重签 | Contract Fail（QA35-01）；真实UI Not Run | 真实POST成功后丢响应→GET operations首次503/abort→等待未知状态→同页面恢复读回；后端操作/代际数量只一次 |
| AC08 轮换丢响应/双提交 | 不能多代链或重复展示已发秘密 | UI-E2E Not Run | 已有用量Key，真实轮换写入后丢响应，重复用户动作和安全读回，继续沿正式生命周期修复丢失秘密，不隐藏API补步 |
| AC16 快速重复发送 | 普通与SSE busy中重复动作不发第二次请求 | UI-E2E Not Run | 同真实页面快速双击/Enter，响应受控延迟，上游计数/同ID/最终账本证明一次，并确认取消提示与控制恢复 |
| AC09 多Key/两代独立资源 | tenant和Key各维度同时生效与竞争守恒 | API-E2E Not Run指定组合 | 实际HTTP barrier竞争有界资源，各Key维度单独命中/恢复，拒绝另一scope不部分扣计；同时另一tenant成功 |

这些内部可完成用例不归外部凭据缺失；本轮不能将“Not Run”改成已执行Fail，也不能以Static Review/Contract Pass替代真实UI关单。J01–J07主链已执行事实保留，下轮添加故障回归，不能删除原健康历史/迁移/keyboard等门槛。
