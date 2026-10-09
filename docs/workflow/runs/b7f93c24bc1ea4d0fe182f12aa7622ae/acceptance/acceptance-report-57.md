# Issue32 独立验收报告（执行57）

> 日期：2026-10-09；Run b7f93c24bc1ea4d0fe182f12aa7622ae，工作流版本2。
> 仓库：big91987/model-relay；分支 workflow/b7f93c24bc1ea4d0fe182f12aa7622ae。
> HEAD：3f9cee1615e87e002a2c2300f2d29778072ec3f5；未提交候选154文件，源码戳 dc667be6f0f5db892748ac8a214d799264c2d451004da8c80a5e06e6c38f338c。
> 环境：执行56宿主冷隔离候选、真实Go服务、独立临时SQLite、真实Chromium、受控HTTP上游；不是生产或Owner preview。
> 模式：任务范围全量验收；QA独立审查同版本宿主实测与实际截图，未自行重跑动态浏览器。
> 结论：**Go with known issues（限本任务受控验收范围）**。无本候选已发现的未关闭P0/P1；历史根因未明及外部未验边界保留，不授予供应商/正式部署全面通过。

## 1. 结论摘要

执行56原固定 `make verify` 冷起点完整退出0，276 Node全部通过、零跳过，七个Go测试包含race/真实HTTP通过，构建及三视口J01–J07、summary、visual、新self_hosted全部完成。研发55、宿主before/after与QA自行复算的154文件候选戳完全一致，不是只比较旧HEAD后拼历史绿色。QA逐字节归档344件当前原证据，独立解码181张PNG均非单色，核验134张像素/几何侧车，并人工查看27张当前安全截图，关闭本轮强制视觉证据缺口。

QA58-01原P3的核验英文、机器状态与窄列断词已在当前页面和原路径复验；必要配置字段、协议名称、原始原因码仍忠实可查看。新增首次自托管链在每个视口各用全新临时安装，从空服务/模型/租户开始，网页创建/纠正/核验/关联/授权/邀请/Key/普通与SSE调用直到同ID日志与费用，真实完成，没有以API或SQL替代网页业务写入。执行54的modal关闭后body仍锁定缺陷由当前真实浏览器原四项严格关闭断言复验通过；不是仅据VM测试或非空图关单。

旧空图的全部历史成因、R28-04旧精确历史断言原因仍未明；本轮不同源码的绿不能追认历史根治。独立原生动态浏览器、真实供应商/支付/外部身份、Owner preview已登录实操、后续发布部署分别未执行。本结论只允许进入report与既有授权的发布核证流程，不能声称已经部署或全面真实联调。

## 2. 权威基线与覆盖

- 使用同一材料回执：design-v0.2.0，archive SHA256 `5f51d35fdc15ed7214749f9413879b4238097f1d6b6bbe2cc0d508d8e103cef7`。实际读取 `.workflow-input/b7f93c24bc1ea4d0fe182f12aa7622ae/5f51d35fdc15ed7214749f9413879b4238097f1d6b6bbe2cc0d508d8e103cef7/files/manifest.json` 及同root `docs/prd.md`（US01–08、§7–8）、`docs/interactions.md`（A工作台、十页、J01–J07、弹层与错误规则）、`docs/acceptance.md`（AC01–18）。参考原型overview.jpg/tenant.jpg，不执行包内脚本。
- 材料archive和manifest全部19项长度/SHA一致；逐项实际路径、SHA、状态见 `evidence/attempt-57/audit-summary.json`。材料仍 `frozen-for-test-implementation`，未改写批准状态。
- 用户Issue32范围优先；沿用 `../delivery/README.md`、T001/M01、`../implementation/terminology-inventory.md` 和历史QA58/本Run QA14，不重建PRD、架构或第二套任务状态。准备回执没有提供根README；现有delivery总览为本Run已确立入口。
- UI结果：平台管理员、租户管理员、开发者分角色独立会话，财务范围/禁止动作保留原回归；J01–J07按原角色/作用域连续执行，首次self_hosted是另一套冷安装，不与既有供应商UI混用。
- 交互与数据：来源失联/401/纠正核验、授权收窄与恢复、轮换/撤销/过渡、暂停/恢复、预算耗尽无dispatch、取消不重放、双提交一次、未知签发/轮换读回、迟到scope隔离、账单快照与更正、过滤/导出、秘密清理均保留。
- 工程与运营事实：锁定依赖、format/vet/mod校验、Go race、真实HTTP、UTC水位/预算/恢复、schema1/2→3迁移和原身份/历史守恒。R30-SYNC-01的保存完成/同步读回/失效隔离契约继续由原Node与真实编辑中断/faults资产覆盖，没有把成功定义改成仅POST成功。

详细范围与分层结果见 [矩阵](acceptance-matrix-57.md)、[连续旅程](journeys-57.md)、[实际图对照](visual-review-57.md)。不把截图、API和代码证据拼成未执行的UI Pass。

## 3. 缺陷及进展

| ID | 等级 | 用户能看到的问题 / 原路径 | 影响与当前状态 | 定位及回归资产 |
|---|---|---|---|---|
| QA58-01 状态术语 | P3（原级别不变） | 服务核验success英文，成员/邀请/请求机器枚举不易理解 | Closed（本候选）；服务核验通过/失败、角色/邀请/请求已知值中文，未知值忠实保留 | workspace状态转换；276 Node及当前services/tenants/records图、self_hosted三种核验图 |
| QA58-01 长值与窄列 | P3（原级别不变） | 地址、安全标识、长名称/字段在窄列内拆碎 | Closed（本候选）；表头/行操作不拆字，完整公开值可聚焦局部滚动，三视口根不溢出、列未隐藏 | readability/technicalValue；三视口89 visual图及30 self_hosted图；全部侧车、当前人工对照 |
| QA14-01 当前视觉证据 | 沿用原强制AC阻断，不重写历史分级 | 旧候选空白/偏移图不能证明实际页面 | Closed（当前候选证据缺口）；181PNG解码非单色、134强制侧车通过、27当前图人工查看；旧故障的全部成因不因此关闭 | capture-integrity、PNG审计与归档；原attempt-14全部保留 |
| 执行54 modal关闭锁缺口 | 强制交互回归阻断主题 | 点击模型详情关闭后modal已关、trigger正确，但body仍fixed/locked | Closed（当前真实原层回归）；三视口33个modal打开/关闭段全部通过，不据截图猜时序 | closeModal/releaseModalScroll；assertModalClosed原四断言与被动close诊断保持；原host-tests-54失败保留 |
| R28-04 历史精确断言原因 | 原历史级别/原因未知，不自行重定级 | 旧精确历史断言失败原因尚未获证 | Open（历史原因未明）；不据不同候选新绿关单，也不虚构Owner风险批准；有复发须回原证据最小归因 | 保留历史断言/读回回归与原日志；新绿不证明根治 |

本候选未发现新的P0/P1或需要退回requirements/design的范围冲突。仍保留 `legacy_remote_model`、`dispatch_possible`、`upstream_name`、快照/版本等精确诊断名称；这是必要技术配置/事实，不把它们改为假业务结果。中文按钮/标题在极窄详情中可自然换行，但仍完整可识别；截图中的局部裁切不是删除值，键盘横滚与完整可访问名称有真实自动证据。

## 4. 真实证据与版本一致性

宿主执行56：2026-10-09T11:00:42.841521Z至11:11:53.351221Z，670,510ms，低于原十二分钟门禁与900秒外层。原 `make verify → isolated make verify-gates`，parallel4不变。候选开始无node_modules/npm cache、NODE_PATH unset、manual_preinstall none，依赖仅固定入口安装。Go1.25.6、Node26.0.0、npm11.12.1、Playwright/core1.59.1、Chromium147.0.7727.15。

- `evidence/attempt-57/host/state.json`：原UTC、退出码、cold start、源不变、候选不变、owned group absent和candidate removed。
- `evidence/attempt-57/host/verify.log`：729,471字节，SHA256 `5f9b70c0e3c0837abd821048aca9a8b2787626441497ad735891481719b32f22`，与state精确匹配。Connector包装日志729,568字节不是同一文件；完整原日志已本地保留，未靠会话省略输出放行。
- `evidence/attempt-57/source-current.json`：QA自行运行 `GOCACHE="$PWD/.cache/go-build" GOTOOLCHAIN=local go run ./tests/evidence`，与开发55及宿主before/after完全相同。
- `evidence/attempt-57/integrity.json`：344件、40,258,790字节，原相对路径/mtime/长度/SHA及归档路径全部保留，字节一致。`png-audit.json`独立解码181PNG非单色；`sidecar-audit.json`核134张PNG摘要/锚点/单次采集前后几何与零scroll delta；不是全181张均有侧车或均人工看过。
- 三视口visual review共89PNG；其中页面条目直接引用56PNG已核摘要，另外33modal截图由侧车及全量索引绑定。J01–J07、summary/self_hosted与失败恢复各有真实JSON。
- QA自行做的是源码复算、材料和日志摘要、证据归档、全量PNG只读解码、侧车核验、27图人工查看和报告。宿主make verify/Go监听/浏览器步骤属于Connector执行56，不冒称QA亲自重跑。

原完整日志内 `TestFixedRunnerReceipts` 的exit23、cancelled/FAILED文字是有意失败/取消夹具，父测试全部Pass；不是顶层门禁失败。Go顶层 `TestConsoleClockRecoveryWorker` 跳过单列Not Run；同轮两种恢复父测试真实启动该子进程两次、强制核CLOCK_RECOVERY_PASS，是恢复AC等价证据，不能把顶层skip本身改为Pass。

## 5. 未执行与环境限制

| 项目 | 结果 | 范围与替代关系 |
|---|---|---|
| QA独立原生动态浏览器 | Not Run | 当前没有可用动态原生控制通道，原监听限制不变；静态browser.check未使用，不能替代Go。宿主真实自动UI结果仍单独Pass |
| QA本地再次make verify | Not Run | 不重复已知沙箱监听失败，不申请宽权限；同候选执行56为强制门禁证据，本用例不写Pass |
| 顶层ClockRecoveryWorker无专用环境直接运行 | Not Run | 原设计skip；同轮TestConsoleClockRecoveryInProcess/ActualHTTP子进程恢复已Pass，明确等价关系 |
| Owner preview已登录实际操作 | Not Run | 未访问已部署真实账号/数据库；旧main health/schema/字节一致只说明旧入口与版本，不代表当前候选已部署 |
| 真实供应商联调 | Not Run | 仅受控HTTP/SSE/401/失联测试，没有真实供应商凭据或响应证明 |
| 真实外部支付 | Not Run | 内部USD测试计价与账单不是支付 |
| 外部身份系统 | Not Run | 内部邀请/角色/会话不是外部身份联调 |
| 发布/新PR/精确同源合并/Owner Actions preview部署 | Not Run | 由后续合法Connector按既有授权核证，不由QA触发 |

## 6. 阶段边界与裁剪

本次小型自托管应用仅做术语/布局/首次分支和完整原回归，不建立组织级运营平台文件。冻结材料H01/H02等平台传输/协调者产品能力不属于本节点新增验收；H07发布部署尚待后续，不伪造。需求/设计无新增冲突，跳过返回requirements/design有现有A工作台与明确Issue32范围依据。未复制New API实现、未改Platform/部署控制器、未新建下一节点或Run。

## 7. 环境恢复与安全

宿主state确认自有进程组消失、候选移除、源码不变；Go harness defer关闭服务/上游/control并移除临时SQLite目录，浏览器finally关闭会话。self_hosted三套数据均独立，不触碰用户上游、原账号、租户、业务配置或真实预览数据。未重放中断或不确定操作，原执行14–54失败与development55首次测试失败日志均保留。

归档文本未检出绝对个人路径，未携带DB、会话、凭据或node_modules；PNG采集保留原秘密/密码/邀请/对话清除或掩码门禁，人工查看图未发现完整秘密。此为采集/检查证据，不是任意秘密不存在的数学证明。console_clock运行水位与备份可能不同，未声称真实库所有行字节不变；本节点没有改库。

## 8. 质量趋势

相对QA14及执行54显著提高：强制视觉证据与body-unlock原层缺口闭合，三视口完整冷门禁和首次self_hosted恢复链首次在该候选一起完成。227→276 Node是资产增长，不作为质量本身的替代；当前像素/几何/滚动/精确焦点、同步关闭helper和失败层投影有回归资产。没有新增当前业务缺陷，旧失败没有删除或洗白，外部边界和历史原因未明仍原样保留。

## 9. Go / No-Go 与交接

**Go with known issues，限 dc667be6…候选的本任务受控验收。** T001 Done、M01 Accepted指产品任务范围退出条件已具备，不表示PR已合并或真实供应商/预览已验。当前不需development返工；下一节点report引用本报告/矩阵/视觉/旅程/完整索引，在同Run `pr.md` 写实际交付说明并保留上述Not Run与历史根因边界。后续发布必须核对精确源码同源；若源码改动须重新评估，不能仅据旧HEAD复用本结论。无需追加阶段批准询问，既有发布授权继续有效。
