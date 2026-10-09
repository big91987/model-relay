# T001 控制台展示修复与独立自托管网页回归

- 状态：**Done**；里程碑：[M01](../milestones/M01-console.md)。
- 模块：Go管理控制台与固定验证Harness；优先级：Normal（原QA58-01为P3，不改分级）。
- 负责人：development；指定评审：独立QA57；交付核对：report58；执行类型：实现与回归 / Review。
- 硬依赖：冻结design-v0.2.0、main基线、正式tests宿主完整门禁与独立QA；无新增软依赖。
- 冲突范围：workspace.js/css、既有Go/浏览器验证资产及Trellis合同；report只整理文档。

## 目标与权威输入
已知状态中文可理解；长来源地址/模型名/安全标识完整可访问；新安装self_hosted网页创建、失败纠正、关联/授权/角色/Key、普通/SSE与同ID费用核对。
权威需求/架构及材料version/SHA/实际引用路径见[总览基线](../README.md#目标与边界)；[编码前逐页盘点](../../implementation/terminology-inventory.md)和[QA57真实图对照](../../acceptance/visual-review-57.md)绑定原QA58-01与当前十页，原型不是后台功能或供应商真实性证据。

## 范围与非目标
- 展示层核验/来源/角色/邀请/请求已知值中文，未知值忠实保留；必要技术名、原始原因码与API值保持。
- 现有表格原生局部滚动、键盘可达、完整公开技术值查看；不删关键列、不不可逆截断、不暴露秘密、不增加布局框架。
- Go Harness每视口独立临时SQLite/角色会话；网页fill/select/click完成业务写入。GET只读起点/同ID/费用，专用control只操作隔离时钟/受控HTTP，不用API/SQL替网页创建。
- 保留原失败断言、单次截图/像素/几何/精确焦点/秘密门禁；最小修正驱动同步、裁剪锚点及诊断；产品统一显式关闭同步释放既有body锁，原生close兜底和秘密清理不变。
- 非目标：API、生产Go后台、费用/Token、schema/迁移、权限调整、供应商默认替身、平台/部署控制器、真实账号/业务库、PR31同步。Node只浏览器测试，后台及运营命令仍Go。

## 交付物
- `internal/relay/web/workspace.js`、`internal/relay/web/workspace.css`；现有Go/Node回归、`tests/browser/journeys.go`、`tests/browser/journeys.mjs`、`tests/browser/readability.mjs`、`tests/browser/capture-integrity.mjs`、固定 `tests/verify`入口。
- `.trellis/spec/frontend/workspace-contracts.md`、`.trellis/spec/frontend/dependency-preparation.md`合同同步；实际文件全集由任务分支diff核对，不重建需求/设计。
- [QA57报告](../../acceptance/acceptance-report-57.md)、[矩阵](../../acceptance/acceptance-matrix-57.md)、[旅程](../../acceptance/journeys-57.md)、[图对照](../../acceptance/visual-review-57.md)、[344件原证据](../../acceptance/evidence/attempt-57/integrity.json)。

## 验收条件与结果
| 条件 | 结果 | 证据 |
|---|---|---|
| QA58-01逐页术语、未知值/必要技术名/API值保持 | Pass | QA57 ISSUE32-01-UI/CONTRACT与十页图对照 |
| 1280/1440/390完整长值/关键列动作、局部横滚/根无溢出、安全图 | Pass | 三visual/self_hosted、134侧车及QA人工查看范围 |
| 新安装网页创建→失联/401→纠正核验→关联/授权/角色/Key→普通/SSE→同ID费用 | Pass | 三self-hosted回执；每次3/4Token、费21/成本7微美元；合计6/8Token、42/14微美元、pending0 |
| 全原Node/Go race/真实HTTP/预算UTC迁移、J01–J07/summary/visual及R30-SYNC-01 | Pass | 宿主56原make verify exit0及全部phase |
| 当前视觉缺口/四项modal关闭原层回归与独立QA | Pass | QA57 Go with known issues；181PNG/134侧车、三视口33个modal打开关闭 |
| 失败/中断/UTC/退出清理/源一致保留，不弱化断言 | Pass | 历次development/host-tests、QA57索引及复算 |

## 验证契约
- 共享入口：`make verify`。固定宿主自动准备锁定Playwright/core1.59.1、Chromium147.0.7727.15；真实Go、隔离SQLite/受控HTTP；不手装依赖、改Connector权限或用户数据。
- 成功：完整exit0、所有规定phase、源/候选不变、自有进程组消失及候选移除。失败保留日志/UTC/退出码/PNG/侧车，先最小归因，不sleep、重试断言或放弱门禁。
- 最近成功：宿主56，2026-10-09T11:00:42.841521Z至11:11:53.351221Z；[state](../../acceptance/evidence/attempt-57/host/state.json)、[完整原日志](../../acceptance/evidence/attempt-57/host/verify.log)，不是QA/report本地重跑。
- 顶层ClockRecoveryWorker无专用环境设计skip列Not Run；同轮恢复父测试实际子进程等价证据Pass，详见QA57报告§4–5。

## 实施结果与交接
- HEAD/基线 `3f9cee1615e87e002a2c2300f2d29778072ec3f5`；分支 `workflow/b7f93c24bc1ea4d0fe182f12aa7622ae`；未提交候选154件源码戳 `dc667be6f0f5db892748ac8a214d799264c2d451004da8c80a5e06e6c38f338c`，研发55/宿主56/QA57/report58一致。
- 实际完成术语/布局、self_hosted链、modal同步释放及全部原回归，无产品/架构偏差；实现及失败依据保留[研发55](../verification/development-55.md)和各历史引用。新PR/提交/部署尚无成功回执。
- 已知限制：旧全部空图成因/R28-04未明；QA独立原生动态、真实供应商/支付/身份、Owner preview已登录及发布部署分别Not Run。不追认旧根治，不冒称真实供应商。
- 当前可依赖：精确候选受控任务退出条件已完成；M01 Accepted。下一步publish核源提交/推送，后续Connector按既有授权处理PR/精确同源合并/preview。
- report58只整理公共投影及[PR正文](../../pr.md)，历史QA/失败证据不改；无用户必须决定事项。源码变更须重评，不能只据旧HEAD复用验收。
