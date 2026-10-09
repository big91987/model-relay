# 当前验收矩阵（执行57）

候选154文件dc667be6f0f5db892748ac8a214d799264c2d451004da8c80a5e06e6c38f338c；HEAD3f9cee1615e87e002a2c2300f2d29778072ec3f5。以下Pass的执行者是宿主Connector56，QA57独立审核；不冒称QA动态重跑。基线为Issue32 AC1–6、design-v0.2.0 AC01–18及J01–J07。证据根 `evidence/attempt-57/host/`，全部精确路径/SHA见 `evidence/attempt-57/integrity.json`。

| ID | 基线要求 | 角色/范围 | 起始状态 | 连续用户步骤 | 最终事实 | 证据类型 | 结果 | 证据位置 |
|---|---|---|---|---|---|---|---|---|
| ISSUE32-01-UI | QA58-01术语与逐页盘点 | 平台管理/当前租户；三视口 | 同J链多对象 | 打开十页/详情/服务核验与成员 | 已知中文状态，必要技术名/未知值保留 | 真实Go浏览器与当前图审核 | Pass | browser/visual-{1280,1440,390}/review.json；visual-review-57.md |
| ISSUE32-01-CONTRACT | 未知值忠实保留/API值不变 | 状态字典/诊断 | 已知及未识别值 | Node调用真实展示函数及反例 | 不把未知显示成功、不改原数据 | Node/代码合同（不是UI替代） | Pass | verify.log；workspace.test.mjs、readability.test.mjs |
| ISSUE32-02-UI | 1280/1440/390长值/完整入口 | 所有十页/详情 | 多对象、长值 | 键盘局部横滚、查看右端列及完整标识 | 根不溢出、列不删、动作可达、名称等于完整文本 | 真实浏览器/像素/可读性事实 | Pass | 三份visual review；三份self-hosted JSON；sidecar-audit.json |
| J01-UI | 接入发布到首次供应商调用 | 平台管理员→租户管理员 | 隔离空业务对象 | 来源/核验/模型/价/授权/预算/成员/Key/调用/详情 | 原共享供应商网页链完成 | 连续UI-E2E | Pass | browser/continuous-1280-52412.json；continuous-1440-52442.json；continuous-390-52483.json |
| J02-UI | 本租户创建/收窄/修正 | 租户管理员 | 已有授权模型 | Key/一次值/调用/收窄拒绝/修正新调用 | 拒绝不dispatch，恢复不重放 | 连续UI-E2E | Pass | 同三份continuous；verify.log |
| J03-UI | 轮换/撤销安全 | 租户管理员 | 已有应用Key | 立即/过渡/新秘密/关闭/旧值/提前撤销 | 新旧共享额度、无叠加、历史保持 | 连续UI-E2E | Pass | 同三份continuous；fault-{1280,1440,390}.json |
| J04-UI | 硬预算耗尽恢复 | 租户管理员 | 已有用量 | 查看/耗尽429/调额/显式再发/核对 | 已用不清零，拒绝零dispatch | 连续UI-E2E | Pass | 同三份continuous；verify.log |
| J05-UI | 租户暂停恢复 | 平台/租户管理员 | 两独立租户 | 创建授权/暂停/拒绝/恢复 | 所属Key拒绝、另一租户独立 | 连续UI-E2E | Pass | 同三份continuous |
| J06-UI | 快照账单更正 | 平台管理员 | 旧/新价及待确认请求 | 改价/旧新日志/拒结/跨月结账/更正/导出 | 原账单不覆写、明细一致 | 连续UI-E2E | Pass | 同三份continuous；visual-activity-{1280,1440,390}.json |
| J07-UI | 范围导航/迟到隔离 | 有权角色/当前租户 | 多页记录 | 详情→日志→分页→切租户→错误重试/刷新 | 旧响应不覆盖、秘密不恢复 | 连续UI-E2E | Pass | 同三份continuous；editor-interruption-{1280,1440,390}.json |
| SELF-UI | 首次自托管完整网页链 | 平台管理→租户管理→开发者，独立会话 | 每视口新SQLite；来源/模型/租户0 | 来源→核验失败→纠正→模型/授权/价/邀请→Key→普通/SSE→同ID日志 | 所有业务写入来自原生用户动作 | 独立冷安装连续UI-E2E | Pass | browser/self-hosted-1280-52430.json；self-hosted-1440-52473.json；self-hosted-390-52507.json |
| SELF-RECOVERY-UI | 失联/核验失败可纠正 | 同self_hosted平台管理员会话 | 地址错误、新来源 | 核验→连接失败→修地址/错凭据401→纠正→核验通过 | revision1/2/3与当前配置适用，失败原因可见 | UI-E2E与真实HTTP | Pass | 同self-hosted JSON；三类核验PNG/侧车 |
| SELF-DATA | 同ID/Token/费用 | 普通租户管理/SSE开发者 | 上述原生Key授权 | 读取浏览器真实响应/同ID详情/只读限额与用量 | 每次3/4 Token、21费/7成本微美元；合计6/8、42/14、pending0 | API/页面数据对账；不是UI补步骤 | Pass | self-hosted JSON normal-ledger/sse-ledger；journeys.mjs原强断言；verify.log |
| AC04-HTTP | 身份/权限/跨租户 | 管理/开发/财务 | 隔离身份夹具 | 实际HTTP合法/非法操作 | 降权/移除/跨租户拒绝 | 独立HTTP及race | Pass | verify.log TestConsolePeopleScopeActualHTTP等 |
| AC09-11-HTTP | 五维预算/UTC/恢复 | 租户/Key资源池 | 隔离数据库/进程 | 竞争/零无限/跨窗/回拨/重启 | 无部分扣计、无重复释放、UTC水位保留 | Go race/HTTP/子进程 | Pass | verify.log TestConsoleClockRecoveryInProcess/ActualHTTP、资源矩阵等 |
| AC18-MIGRATION | 旧身份/历史守恒 | schema1/2旧样本 | 隔离旧版样本 | 正式CLI升级→真实三视口旧账号/Key/历史 | schema3、身份/额度/快照/历史保持 | Go/实际CLI/浏览器 | Pass | verify.log migration phase |
| R30-SYNC-01 | 完成必须读回及失效隔离 | 原编辑/签发/轮换流程 | 延迟/丢响应/旧scope | 用户保存→读回→明确完成；中断不覆盖新编辑 | 一次写、旧结果不污染新状态 | Node及真实UI faults/interruption | Pass | verify.log；browser/fault-{1280,1440,390}.json；editor-interruption三JSON |
| MODAL-54-UI | 原四项关闭边界 | 同visual角色/作用域 | 打开原模型/服务/Key等11modal | 原用户关闭按钮→hidden→立即原断言 | native关/root解锁/body非fixed/精确trigger焦点 | 当前真实浏览器，保留被动close诊断 | Pass | 三visual review result passed；verify.log；console-visual.mjs assertModalClosed |
| VISUAL-14-EVIDENCE | 空图/偏移不能放行 | 三视口 | 当前一次fullPage采集 | 原像素/几何断言→独立解码/人工看图 | 181非单色；134锚点/几何侧车通过 | PNG/像素侧车/QA人工 | Pass | png-audit.json、sidecar-audit.json、visual-review-57.md |
| SUMMARY-UI | 当前周期及十页摘要真实 | 平台管理/当前租户 | 当前周期0调用 | 原生新调用→刷新摘要/入口 | 调用0→1、费用21微美元、旧closed bill不变 | 浏览器与API值逐列核对 | Pass | management-summary-{1280,1440,390}.json；visual-activity三JSON |
| GATE | 全部原固定门禁/冷依赖/退出清理 | 宿主候选 | 无node_modules/npm cache | 原make verify一次完整执行 | exit0、276Node/七包race/HTTP/build、全phase通过 | Connector实测（非QA重跑） | Pass | state.json、verify.log、audit-summary.json |
| CLOCK-WORKER-DIRECT | 无子进程环境直接执行helper | Go顶层发现 | 无专用状态env | 直接worker →原设计skip | 不作为单用例绿；同轮父测试子进程等价恢复证据 | Go skip | Not Run | verify.log第598–600行；恢复父测试Pass，报告§4说明 |
| QA-NATIVE | QA动态原生独立复跑 | 当前QA沙箱 | 无动态控制通道 | 未运行 | 宿主UI结果不能改写本行 | 环境边界 | Not Run | acceptance-report-57.md §5 |
| QA-LOCAL-GATE | QA本地重新make verify | 原监听权限 | 已有同版本宿主完整证据 | 未重复监听失败 | 不新增审批/手装依赖 | 环境边界 | Not Run | acceptance-report-57.md §5 |
| OWNER-PREVIEW | 实际预览已登录操作 | 既有账号/配置只读 | 旧main preview | 未操作 | 入口不是全面UI验收 | 正式环境边界 | Not Run | acceptance-report-57.md §5 |
| SUPPLIER | 真实供应商响应 | 外部供应商 | 无当前实证 | 未调用 | 受控HTTP不替代供应商事实 | 外部边界 | Not Run | acceptance-report-57.md §5 |
| PAYMENT | 真实外部支付 | 外部支付 | 未接通 | 未执行 | 测试计价非支付 | 外部边界 | Not Run | acceptance-report-57.md §5 |
| IDENTITY | 外部身份 | 外部身份系统 | 无联调条件 | 未执行 | 内部邀请不替代外部身份 | 外部边界 | Not Run | acceptance-report-57.md §5 |
| RELEASE | 新PR/精确同源合并/preview部署 | 后续Connector | 当前未提交候选 | 待report与核证发布 | 未声称已发布/已部署 | 后续阶段 | Not Run | acceptance-report-57.md §9 |

R28-04和旧空图全部历史成因未知保留为报告的已知风险，不制造“历史原因复验Pass”行；原失败证据仍在本Run及原Run。当前视觉缺口与当前body-lock可以在实际原层关闭，不等于替旧根因推断。无截图/受控HTTP证据可以覆盖SUPPLIER/PAYMENT/IDENTITY/OWNER-PREVIEW/QA-NATIVE行。
