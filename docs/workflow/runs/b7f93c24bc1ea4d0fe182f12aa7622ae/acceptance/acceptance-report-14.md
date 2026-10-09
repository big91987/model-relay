# Issue32 候选独立验收报告（执行14）

> 日期：2026-10-09（UTC）；Run b7f93c24bc1ea4d0fe182f12aa7622ae；全量验收。
> 仓库：big91987/model-relay；分支 workflow/b7f93c24bc1ea4d0fe182f12aa7622ae。
> HEAD：3f9cee1615e87e002a2c2300f2d29778072ec3f5；未提交候选152件 SHA256：65201f488ae0969684625a92c23fb2c1e76bed30916b2c92a405b1e9e00f92a9。
> 结论：**No-Go：强制三视口实际视觉证据未闭合；返回 development。** 不是把全部行为测试判失败，也不是把 P2 自动升级为产品 P1。

## 1. 结论摘要

正式 tests 执行13的固定 `make verify` 冷起点完整通过，当前源码戳、前后戳及交接执行12候选一致。首次 self_hosted 三视口网页创建、两类核验失败、纠正恢复、授权、Key、普通/SSE及同ID费用链均实际执行成功，受控HTTP证据不代表真实供应商联调。

独立QA实际查看30张主页面和19张补充图片，发现部分截图全空白、主体消失或整体偏移，与绿色DOM/尺寸断言不一致。解码全部181张PNG确认4张逐像素单色，其中1280 API Keys主图与右列图均空白，1440租户右列图及self_hosted角色图也空白；另外存在非单色但主体不可审查的图。截图SHA与回执一致，不是QA复制或图像查看器传输造成的摘要失配。

因此服务中文核验状态与390长标识等可单项认可，但QA58-01整体和本次三视口视觉强制AC不能关闭。原因目前只能定位为当前原生运行的渲染/采集事实与DOM回执失配，尚不能断定产品CSS错误、Chromium机制或截图时序。无需需求/设计返工；先由研发最小归因和补验证，再走正式tests及独立QA，不反复完整求绿。

## 2. 基线、版本与本轮覆盖

- 冻结输入仍为 design-v0.2.0 / frozen-for-test-implementation；archive SHA256 5f51d35fdc15ed7214749f9413879b4238097f1d6b6bbe2cc0d508d8e103cef7。manifest及实际PRD/交互/AC路径、各文件摘要复核见 `evidence/attempt-14/audit-summary.json`；未改写为用户批准。
- 当前任务事实沿用 `../delivery/README.md`、`../delivery/tasks/T001-console.md`、`../delivery/milestones/M01-console.md`；定位盘点沿用 `../implementation/terminology-inventory.md`。旧QA58-01依据为旧Run `4cd62ced6a4f98f50e3cca6f7596eb06/acceptance/visual-review-58.md`，不把旧图当候选图片。
- 核对现有 `docs/workflow/runs/4cd62ced6a4f98f50e3cca6f7596eb06/implementation-contract.md` 的人员身份、权限和安全契约；旧首次调用架构与 `docs/workflow/design.md` 属于既有能力历史，不代替 design-v0.2.0 冻结规格。本次没有重建架构或复制PRD。
- 产品修改限于 workspace.js/css；Go业务实现、API/SQLite schema/费用/Token/权限及默认上游没有本次修改。测试驱动、隔离HTTP夹具与回归增加，不是产品默认供应商。
- 宿主时间：2026-10-09T01:21:34.676495Z至01:33:40.695851Z；Go1.25.6、Node26.0.0、npm11.12.1；锁定Playwright/playwright-core1.59.1、Chromium147.0.7727.15。实例标签 continuous-controlled-ui / schema3 是隔离测试标签，不是正式部署SHA。
- 角色：平台管理员、两个租户管理员、开发者、财务观察者。已有J01–J07按固定脚本连续执行，各角色保持自己的真实会话和租户边界；新self_hosted明确由平台管理员配置，受邀管理员签发/普通调用，受邀开发者使用同应用授权SSE，不是前端角色切换。
- 证据层级：宿主UI-E2E、API/HTTP、计量读回、Node/Go契约、实际PNG与QA静态审查分别记结果。详细映射和连续步骤见 `acceptance-matrix-14.md`、`journeys-14.md`。

## 3. 缺陷及进展

| ID | 等级 | 用户可理解的问题 / 影响 | 定位与复现 | 状态 / 回归 |
|---|---|---|---|---|
| QA58-01 | P3（保留原分级） | success英文、窄列断词影响状态与安全标识理解 | 已知状态中文/未知原值/完整文本region已实现；当前服务图及390分支可复核，但1280/1440部分图缺失可读主体 | Ready for Retest，分项结果见 visual-review-14；不整体Closed；workspace/readability/三视口原生资产保留 |
| QA14-01 | P2，验收工程证据缺陷；产品用户影响待归因 | 绿色回执所指真实截图为空白或偏移，验收者无法判断对应页面可读性 | 同候选执行13 make verify，直接打开所列PNG；4单色图可用audit-png重复解码；DOM结果均为true但图没有内容 | Open；本次强制视觉AC证据缺口阻止Go，而非P2默认阻止；development负责补原层采集/验证 |
| R28-04 | 沿用历史问题，不重新分级 | 原精确历史断言原因未知 | 新全绿没有原根因证明 | 未宣称根治，继续历史风险 |

没有发现已确证的新业务P0/P1。QA14-01先按证据缺陷处理，不把空白图直接归因“线上页面全白”，也不以不确定性放行。整改条件见 `remediation-14.md`。

## 4. 执行结果与证据

- 宿主：固定make verify退出0；227/227 Node、fail/cancelled/skipped/todo均0（执行11旧数225不能替本轮227）。format-check、go mod verify、vet、Go race/count1实际HTTP/预算UTC/迁移、build、全部浏览器阶段完成。原历史红例是显式历史重构/历史断言门禁，不是当前资产红灯；完整日志原样保存，不只保留绿色尾部。
- 原网页回归：三视口J01–J07、summary、当前周期0→1/21微美元/closed bill不改、保存读回/未知提交/一次性秘密/双发送/轮换/预算/成员/取消/路由错误恢复仍通过。1280/1440/390各10页visual的DOM断言通过，但实际PNG部分不合格，不能把 `result:passed` 当独立视觉Pass。
- self_hosted：每视口独立SQLite新安装，初始upstreams/models/tenants均0。网页POST/PATCH/PUT由点击保存产生；测试控制仅操作隔离时钟/受控HTTP，GET仅读回。连接关闭失败→401→编辑纠正→200；配置revision1→2→3且旧核验失效。普通/SSE各输入3输出4、租户费21微美元、成本7微美元、pending0；合计6/8、42/14。详情、日志、价格快照、额度增量与用量页面同结果，不以SQL补写网页。
- 原生键盘scrollend后唯一wheel及2000ms精确末列断言本轮行为通过；不从绿色推断Chromium内部原因，不放宽原断言。截图缺陷是本轮独立发现的新证据边界。
- 归档：`evidence/attempt-14/host/`原210件（181PNG、29文本），24,952,711字节；逐件原/副本摘要一致。`integrity.json`逐件具体路径/字节/SHA/原mtime。visual回执引用89张PNG摘要全部核对，self_hosted另有30张图和3份完整步骤回执。
- 原verify.log 373,368字节、SHA256 eeba757fc1f4728aeaab9c844a8aa274969f44c0638b9f7fa04b1609242b105c，与state一致；Connector日志373,465字节为命令包装，不误称相同文件。完整本地原日志可访问且有摘要，不依赖会话中省略输出判断通过。
- QA自行执行：源码戳复算、material摘要核对、210件归档完整性、89图片声明摘要核对、181PNG只读解码、49图人工查看。初始默认Go缓存读取被沙箱拒绝，使用Makefile同工作区GOCACHE复算成功，未扩大权限。不是QA亲自另跑宿主浏览器。

## 5. 未执行与限制

| 项 | 结果 | 原因 / 替代关系 |
|---|---|---|
| QA原生动态独立浏览器重跑 | Not Run | 没有可用动态原生控制通道，已有监听限制不变；没有使用静态browser.check伪代动态Go |
| QA本机再跑make verify | Not Run | 不重复已知沙箱监听必失败路径；同候选宿主执行13是行为门禁证据，本用例仍非Pass |
| Owner preview已登录实操 | Not Run | 未操作真实预览，入口health与旧main同字节不是全面UI验收 |
| 真实供应商、外部支付、外部身份 | Not Run（各自） | 无当前真实响应/条件；受控HTTP、内部邀请和测试计价不能替代 |
| 发布/新PR/精确同源合并/Owner Actions部署 | Not Run | 尚未进入对应Connector，QA不自行发布 |

以上独立边界不阻止承认宿主自动UI旅程的真实通过；但没有等价证据覆盖空白/偏移图所缺的三视口实际视觉结果。

## 6. 阶段边界

只验本任务的术语、窄屏可读性、新首次自托管分支和原严格回归。H01/H02/平台传输等不是本节点新增产品验收；H07发布部署交后续Connector，不重建平台运营资料。无需求矛盾/权限扩大，跳过requirements/design返工有冻结设计和同Run边界依据。不得用原Issue30/Run2574的completed事实替当前验收。

## 7. 环境恢复与安全

state记录source_unchanged/candidate_unchanged、owned_group_absent=true，临时候选移除；Go harness关闭自有服务/上游/control并删除临时目录，browser finally关闭会话。QA没有启动产品服务、触碰预览、重置账号、修改数据库或共享配置。归档29文本未发现个人主机路径及完整Key模式；已查看图没有完整秘密，截图前秘密/密码/邀请/对话清除门槛保留。模式检查不是全秘密不存在的数学证明。

保留此前05/07/09/11失败归档，不覆盖旧结论。真实预览console_clock水位与备份不同的已知边界仍保留，未宣称整库所有行逐字节不变。

## 8. 质量趋势

相对执行11，完整原生行为从1280右列失败前终止提升为全阶段通过；首次self_hosted从Not Run提升为三个独立安装UI-E2E实测。术语和390长值可读性改善；227项Node与新分支沉淀持续回归。独立QA新增发现截图质量漏检，说明DOM/安全检查不足以保证采集可用于验收；截图层没有红转绿证据，不能整体宣布交付完成。R28-04和真实外部/预览边界不变。

## 9. Go / No-Go

**No-Go，返回 development 修订验证与补齐原生视觉证据。** 必须先最小归因QA14-01，再保留完整固定make verify与原门槛，补当前候选三视口安全可读图及独立复验。不得增加sleep/断言重试、删除关键列、修改真实DB或平台权限来求绿。行为通过不作废，截图缺口也不隐藏；Task仍未Done、Milestone未Accepted。
