## 管理控制台术语与窄屏可读性，首次自托管网页旅程

关联 #32。承接正式main `3f9cee1615e87e002a2c2300f2d29778072ec3f5`（PR31），不重复Issue30/PR31或同步修复。沿用冻结design-v0.2.0，archive SHA256 `5f51d35fdc15ed7214749f9413879b4238097f1d6b6bbe2cc0d508d8e103cef7`；未重建PRD/架构或改写材料批准状态。

### 用户可见改进
- 服务核验、来源类型、成员角色/邀请和请求已知状态使用中文业务术语；未知值忠实可识别，必要技术配置名、原因码、API原始值保留。十页盘点对应QA58-01，不新增协议/实现独白或重复假状态。
- 保留现有表格/详情全部关键列及动作；1280/1440/390长来源地址、模型名、安全标识可聚焦、局部横滚并完整取回，页面根不溢出，不靠不可取回截断或隐藏列换绿。完整标识仅公开元数据，秘密一次性展示/清理不变。
- 统一modal显式关闭同步释放既有body滚动锁；保留原生close兜底、精确焦点、秘密清理，修复执行54原路径关闭锁缺口，不在断言处增加等待掩盖问题。
- 补首次self_hosted新安装网页链与失联/401核验失败纠正恢复，真实Go服务执行普通/SSE请求至隔离受控HTTP上游，未新增产品默认假供应商。

产品变更集中在workspace.js/css；必要Node/Go/浏览器回归、单次截图完整性诊断、固定验证入口及Trellis合同同步。生产Go后台、API/schema、费用/Token、预算/授权/安全口径保持，Node仅测试；未改Platform/部署控制器，未操作真实账号或业务库。

### 精确版本与真实验证
- 审核分支 `workflow/b7f93c24bc1ea4d0fe182f12aa7622ae`，目标main；HEAD/基线 `3f9cee1615e87e002a2c2300f2d29778072ec3f5`；**未提交候选154件源码戳** `dc667be6f0f5db892748ac8a214d799264c2d451004da8c80a5e06e6c38f338c`。该戳覆盖产品/测试/构建锁/指定项目合同/Trellis，不是全部交付文档的Git提交摘要。
- 研发55、宿主56 before/after、QA57与report58复算一致。publish须核精确候选提交/推送；源码变化须重评，不能只据旧HEAD沿用绿。
- 原固定 `make verify`：宿主56于2026-10-09T11:00:42.841521Z至11:11:53.351221Z一次冷隔离运行，670,510ms，**exit0**。无预置node_modules/npm cache，NODE_PATH unset，无手工预装；Go1.25.6、Node26.0.0、npm11.12.1、Playwright/core1.59.1、Chromium147.0.7727.15。
- 276 Node零失败/零skip；format/vet/modules/build、七包Go race/真实HTTP、预算UTC/schema1/2→3迁移及三视口J01–J07/summary/visual完成。原共享供应商、两来源路由、权限/收窄/轮换/撤销/预算/取消/未知费用/错误恢复与R30-SYNC-01严格回归保留。
- 新self_hosted每视口独立临时SQLite及平台管理员/租户管理员/开发者会话，从空业务对象开始，网页创建自托管来源→失联→纠正地址/错误凭据401→纠正核验成功→关联模型→租户授权/价格→网页邀请接受→应用Key→普通/SSE→各自同ID日志费用。业务写入无API/SQL替代，GET仅只读核对，control只隔离时钟/上游夹具。每次3/4 Token、费21/成本7微美元；合计6/8 Token、费42/成本14微美元、pending0。
- 宿主state记录源/候选不变、自有进程组消失及候选移除。原verify.log为729,471字节，SHA256 `5f9b70c0e3c0837abd821048aca9a8b2787626441497ad735891481719b32f22`；Connector包装日志729,568字节不是同一文件。
- 独立QA57：**Go with known issues，仅Issue32受控范围**；核344件/40,258,790字节原证据、181PNG非单色、134像素/几何侧车，人工看27当前图与4基线图。QA58-01、当前QA14视觉缺口、执行54body-unlock原层闭合；不是人工逐张看完181图，也不是QA亲自重跑浏览器。

### 未验范围与保留风险
| 项目 | 状态及限制 |
|---|---|
| QA独立原生动态浏览器、QA本地再次make verify | Not Run；原监听/控制通道限制不变，宿主真实自动UI Pass单独成立，静态browser.check不替代后台 |
| 真实供应商、外部支付、外部身份 | 分别Not Run；受控HTTP/SSE、内部USD计价/账单/邀请不代替外部联调 |
| Owner preview已登录真实操作 | Not Run；旧main health/schema/资源字节一致不代表当前候选部署或全面UI验收 |
| 顶层ClockRecoveryWorker无专用环境直接运行 | Not Run（设计skip）；同轮两个父测试真实子进程核恢复成功为AC等价证据，不把skip写Pass |
| 旧空图全部历史成因、R28-04旧精确历史断言 | 未明；不同源码新绿不证明历史根治，不删除失败证据或虚构Owner风险批准 |
| 本轮提交/推送、新PR、精确同源合并、Owner Actions preview部署 | 尚未执行；由后续授权Connector核证，本文不宣称已发生 |

本轮测试不触碰真实账号、租户、成员或业务配置；console_clock运行水位可能与备份不同，不宣称整库所有行逐字节不变。未重放中断或不确定写入/供应商调用，执行14–54失败与停止接续事实原位保留。回退沿用现有同源发布/升级契约，不自动倒写运行水位、重置账号或改DB求绿。本次小型应用切片无范围/架构偏差，跳过不适用需求/设计返工，不补造组织级运营文档。

### 可审阅资料
**以下是核对实际仓库与任务分支后生成的推送后完整仓库URL。report时候选尚未提交/推送，不声称新增文件现在远端可访问；publish须推送目标文件并在PR发布前确认链接，未把新产物链接到旧main或旧提交。**

- [公共交付总览](https://github.com/big91987/model-relay/blob/workflow%2Fb7f93c24bc1ea4d0fe182f12aa7622ae/docs/workflow/runs/b7f93c24bc1ea4d0fe182f12aa7622ae/delivery/README.md) · [T001任务](https://github.com/big91987/model-relay/blob/workflow%2Fb7f93c24bc1ea4d0fe182f12aa7622ae/docs/workflow/runs/b7f93c24bc1ea4d0fe182f12aa7622ae/delivery/tasks/T001-console.md) · [M01里程碑](https://github.com/big91987/model-relay/blob/workflow%2Fb7f93c24bc1ea4d0fe182f12aa7622ae/docs/workflow/runs/b7f93c24bc1ea4d0fe182f12aa7622ae/delivery/milestones/M01-console.md)。
- [独立QA57报告](https://github.com/big91987/model-relay/blob/workflow%2Fb7f93c24bc1ea4d0fe182f12aa7622ae/docs/workflow/runs/b7f93c24bc1ea4d0fe182f12aa7622ae/acceptance/acceptance-report-57.md) · [逐项矩阵](https://github.com/big91987/model-relay/blob/workflow%2Fb7f93c24bc1ea4d0fe182f12aa7622ae/docs/workflow/runs/b7f93c24bc1ea4d0fe182f12aa7622ae/acceptance/acceptance-matrix-57.md) · [连续网页旅程](https://github.com/big91987/model-relay/blob/workflow%2Fb7f93c24bc1ea4d0fe182f12aa7622ae/docs/workflow/runs/b7f93c24bc1ea4d0fe182f12aa7622ae/acceptance/journeys-57.md) · [真实图对照及人工范围](https://github.com/big91987/model-relay/blob/workflow%2Fb7f93c24bc1ea4d0fe182f12aa7622ae/docs/workflow/runs/b7f93c24bc1ea4d0fe182f12aa7622ae/acceptance/visual-review-57.md)。
- [344件原证据索引](https://github.com/big91987/model-relay/blob/workflow%2Fb7f93c24bc1ea4d0fe182f12aa7622ae/docs/workflow/runs/b7f93c24bc1ea4d0fe182f12aa7622ae/acceptance/evidence/attempt-57/integrity.json) · [原UTC/退出码/冷起点/清理](https://github.com/big91987/model-relay/blob/workflow%2Fb7f93c24bc1ea4d0fe182f12aa7622ae/docs/workflow/runs/b7f93c24bc1ea4d0fe182f12aa7622ae/acceptance/evidence/attempt-57/host/state.json) · [完整原日志](https://github.com/big91987/model-relay/blob/workflow%2Fb7f93c24bc1ea4d0fe182f12aa7622ae/docs/workflow/runs/b7f93c24bc1ea4d0fe182f12aa7622ae/acceptance/evidence/attempt-57/host/verify.log) · [源码/材料审核摘要](https://github.com/big91987/model-relay/blob/workflow%2Fb7f93c24bc1ea4d0fe182f12aa7622ae/docs/workflow/runs/b7f93c24bc1ea4d0fe182f12aa7622ae/acceptance/evidence/attempt-57/audit-summary.json)。
- [390服务右端动作安全图](https://github.com/big91987/model-relay/blob/workflow%2Fb7f93c24bc1ea4d0fe182f12aa7622ae/docs/workflow/runs/b7f93c24bc1ea4d0fe182f12aa7622ae/acceptance/evidence/attempt-57/host/browser/visual-390/services-right-columns-1.png) · [390模型完整详情安全图](https://github.com/big91987/model-relay/blob/workflow%2Fb7f93c24bc1ea4d0fe182f12aa7622ae/docs/workflow/runs/b7f93c24bc1ea4d0fe182f12aa7622ae/acceptance/evidence/attempt-57/host/browser/visual-390/models-detail.png)。
- [QA14历史No-Go](https://github.com/big91987/model-relay/blob/workflow%2Fb7f93c24bc1ea4d0fe182f12aa7622ae/docs/workflow/runs/b7f93c24bc1ea4d0fe182f12aa7622ae/acceptance/acceptance-report-14.md) · [执行54原失败](https://github.com/big91987/model-relay/blob/workflow%2Fb7f93c24bc1ea4d0fe182f12aa7622ae/docs/workflow/runs/b7f93c24bc1ea4d0fe182f12aa7622ae/delivery/verification/host-tests-54/archive-index.json) · [研发55修复及历史边界](https://github.com/big91987/model-relay/blob/workflow%2Fb7f93c24bc1ea4d0fe182f12aa7622ae/docs/workflow/runs/b7f93c24bc1ea4d0fe182f12aa7622ae/delivery/verification/development-55.md) · [部署与回退约束](https://github.com/big91987/model-relay/blob/workflow%2Fb7f93c24bc1ea4d0fe182f12aa7622ae/docs/local-deployment.md)。

T001 Done/M01 Accepted仅任务退出条件，公共交付判断Verifying仍等待Git发布核证。无需重复阶段批准；既有发布/精确同源合并/Owner preview授权继续有效，不扩大权限。
