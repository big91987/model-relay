# Issue30 QA8 验收矩阵与连续旅程

基线：Owner Issue30 最小复现/验收1–6、T001、design-v0.2.0 的 AC01/AC08/AC15/AC17 和 J01–J07。冻结材料保持 frozen-for-test-implementation，不改称重新批准。快照：main HEAD `9bf6ecaf53758ca2258498968b830c535e02e1c0` 加未提交候选，150文件摘要 `0384f115adbbd545e8ce01021f238010ba66c89eb06e6acea08ce872ce0fa474`。不是声称 main 已有修复。

以下 `host/` 均指本文件相邻 `evidence/qa-08/host-tests-07/`；日志与state原字节归档。UI-E2E Pass 均来自正式 tests7 宿主真实Go/Chromium，不是QA自己驱动的浏览器，也不是供应商联调。native独立动态QA单独列Blocked。JSON、源码与函数测试是补强证据，不拼接成独立UI通过。

## 需求—旅程—状态—测试—证据

| ID | 基线要求 | 角色/范围 | 起始状态 | 连续用户步骤 | 最终事实 | 证据类型 | 结果 | 证据位置 |
|---|---|---|---|---|---|---|---|---|
| S01-UI | 本次PATCH后同scope tenants/catalog刷新完成才save返回；false/true及窄屏 | 平台管理员，同tenant/model页/context | UI已创建两模型，目标启用 | 精确模型行→启停→checkbox false/true→保存；点击前route布置，仅PATCH200后GET响应hold/release | 12组、24个阶段均dialog已隐藏、loading/notice/pending/锁保持、save未返回；原样释放后唯一行/列/状态正确 | UI-E2E，宿主受控 | Pass | host/browser/model-save-readback-{1280,1440,390}.json；tests/browser/model-save-readback.mjs |
| S01-C | 操作、scope、page屏障拒绝错误完成；严格负例仍红 | 实际函数与驱动契约 | VM fixture，不是浏览器 | 当前operation终态一次性判断；注入错状态/缺列/重复行 | 原AssertionError身份、唯一行/列/值不降级；不sleep/networkidle/加timeout/重试assert | Contract，QA独立Node | Pass | evidence/qa-08/independent-node.log（75/75）；model-save-readback/save-lifecycle/console-summary tests |
| S02-UI | GET失败不能以旧行假成功，恢复只读 | 平台管理员，同tenant | 写入已确认；真实GET200后显式替换为503 | 启停保存→GET失败→看到读取错误→重试读取 | 6组save拒绝、旧行不呈现、GET200恢复且严格状态正确；503是注入不是数据库失效 | UI-E2E，宿主受控故障 | Pass | host/browser/model-save-readback-{1280,1440,390}.json；代码boundary handler |
| S03-UI | scope切换与第二次相反保存不被第一次迟到快照覆盖 | 同一平台管理员context；原tenant→全部租户→原tenant | 第一次停用后的真实catalog响应held | 保存停用→切范围→第一次save拒绝→切回→相反启用保存→释放旧快照 | 3组第二次完成先于旧响应释放，最终原范围模型启用 | UI-E2E，宿主受控 | Pass | 同上 case=scope-late-opposite |
| S03-C | 旧epoch tenants/catalog和finally不改新范围/新编辑；busy不允许替换 | 实际workspace函数 | deferred原函数fixture | 旧读迟到、新编辑、重复submit、同/不同epoch路径 | 原scope读隔离、busy锁/取消/close/secret/unknown兼容 | Contract，QA独立Node | Pass | evidence/qa-08/independent-node.log；tests/browser/save-lifecycle.test.mjs |
| S04-UI | 原取消/关闭/新编辑/secret/幂等保持 | 平台/租户身份受服务端判断 | 同轮独立安装/context | editor interruption、签发/轮换丢响应读回、双普通/SSE提交 | 新编辑save可用、取消编辑零写；秘密不重放；双击实际仅一次dispatch | UI-E2E，宿主受控 | Pass | host/browser/editor-interruption-{1280,1440,390}.json；fault-{1280,1440,390}.json |
| S05-GATE | 完整固定make verify，不用局部绿放行 | 私有候选/临时服务 | node_modules/npm cache均不存在 | 项目verify→verify-gates→format/module/vet/Node/race/build/browser | exit0；213 Node通过；全包race和真实HTTP；历史红分类符合原错误；源码before/after一致 | Contract/API-E2E，宿主 | Pass | host/verify.log；host/state.json；evidence/qa-08/archive-index.json |
| S05-DATA | 零软预算、有效UTC、迁移保留事实 | 隔离安装与测试时钟 | 标准夹具/历史种子，只作Start State | HTTP零预算与回拨；实际进程恢复；CLI迁移1/2→3并原生UI核对 | 零预算无dispatch；截止/回拨不复活；identity/key/history/usage守恒 | API-E2E / Data Reconciliation | Pass | host/verify.log：ConsoleReviewZeroSoftBudgetActualHTTP、ConsoleReviewClockRollbackActualHTTP、ClockRecoveryActualHTTP及migration phase |
| S05-SKIP | 单独worker顶层运行 | 无MODEL_RELAY_CLOCK_STATE | 未配置子进程专用状态 | 普通go test枚举worker | 顶层worker实际skip，不能写Pass；父ClockRecoveryInProcess/ActualHTTP分别以真实子进程断言CLOCK_RECOVERY_PASS，覆盖同一恢复不变量 | Contract | Not Run | host/verify.log 534–536；internal/relay/console_clock_recovery_test.go:117/139/147 |
| S06-UI | 三视口原summary/10页visual/连续旅程 | 平台/租户管理员、开发者、财务 | 每视口新私有安装，真实角色contexts | 见下方独立旅程行 | 21条J01–J07全部完成；summary严格状态/API一致；30页rootFits/secretsClear | UI-E2E，宿主受控 | Pass | host/browser/continuous-1280-86354.json、continuous-1440-86373.json、continuous-390-86389.json；management-summary/visual-* |
| S06-INDEPENDENT-UI | QA本节点独立运行真实页面 | native沙箱 | listener/依赖连接限制未变 | 无法在当前工具内启动/驱动同版真实后端旅程 | 不能用静态browser.check或VM代真浏览器；同轮宿主S01–S06覆盖候选发布门槛，不改此行结果 | UI-E2E，独立动态 | Blocked | verification/native-verify-04.log/state；frontend-checks-06.log；本轮报告 |
| S06-SUPPLIER | 真实供应商响应与费用事实 | 真实provider | 未提供凭据/供应商条件 | 未执行真实供应商请求 | controlled_http_not_supplier仅验证产品协议/管理链 | External State | Not Run | host/browser/*.json provider；Owner边界 |
| S06-DEPLOY | 新精确提交合并、Owner正式preview及实际登录 | Owner/协调者 | 当前仍未提交候选 | 本节点不提交/部署，不操作已有账号 | 新prepare/部署/Deployment/health/schema3及实际入口待后续；旧PR29不是新候选证据 | External State / UI-E2E | Not Run | Owner授权；交report |

## 连续用户旅程（每条单独UI-E2E）

每个视口从新安装的空collections起步，初始化/测试时钟/临时受控HTTP是显式准备；管理写入均由真实页面完成，不用API/SQL替代UI中间步骤。J01–J07共享该视口运行；不同角色按真实邀请/登录使用各自context，每个角色段不更换身份或用假角色切换。API读取用于最终对账，测试control仅准备故障/时钟/清理，不是供应商事实。Cleanup均为private directories、servers、contexts，state记录owned group absent及candidate removed。

| Journey ID | User Story/AC | Actor | Start State | UI Steps | Expected UI Result | Final Fact | Cleanup | Result |
|---|---|---|---|---|---|---|---|---|
| J01-UI | AC01–04/06/12/16 | 平台管理员；邀请后的成员身份context | 无预配置collections | 接入服务→核验→发布模型→价格/租户授权→预算→邀请→Key→测试调用→详情 | 配置及可调用结果连续完成 | 请求/用量/价格快照相符；controlled HTTP | 同轮隔离清理 | Pass |
| J02-UI | AC03/06/16 | 租户管理员，同tenant context | J01正式创建并授权 | 目录→Key→一次性值→调用→收窄→拒绝→修正→新调用 | 拒绝与修正路径明确 | 拒绝零dispatch；授权按服务端交集 | 同上 | Pass |
| J03-UI | AC07/08/11/17 | Key管理员原context | 正式Key及已用量 | 详情→立即/过渡轮换→秘密→关闭→旧值状态→提前撤销 | 秘密清除，过渡不叠加 | 新旧共享使用池/历史，截止边界 | 同上 | Pass |
| J04-UI | AC09–11 | 租户管理员 | 已有预算与实际使用 | 查看已用→耗尽→429→提高限额→显式重发→详情 | 已用不清零、拒绝可恢复 | 硬预算/无dispatch/最终计量一致 | 同上 | Pass |
| J05-UI | AC04/05 | 平台管理员及受影响tenant contexts | UI创建租户、成员/Key | 授权→使用→暂停→拒绝→恢复→当前授权使用 | 另一租户独立 | 暂停限制所有Key，历史保持 | 同上 | Pass |
| J06-UI | AC12–14 | 平台管理员/租户读取及财务context | 真实调用历史与测试时间边界 | 改价→旧/新快照→账单核对→拒绝当前/待确认结账→结账→更正→导出 | 不覆写原金额，范围正确 | charge/adjustment-period守恒、跨租户拒绝 | 同上 | Pass |
| J07-UI | AC15/17 | 同角色同范围操作段 | 同轮已建对象/多页列表 | 详情跳日志→页2→切tenant→旧读迟到→错误重试→reload | 新scope不被覆盖，安全页面恢复 | scope由服务端确定，秘密不恢复 | 同上 | Pass |

三宽度共同证据：host/browser/continuous-1280-86354.json、continuous-1440-86373.json、continuous-390-86389.json（逐J checkpoint）及host/verify.log的逐旅程结果。源码步骤为 tests/browser/journeys.mjs；这些结果是宿主自动化而非Owner独立预览。
