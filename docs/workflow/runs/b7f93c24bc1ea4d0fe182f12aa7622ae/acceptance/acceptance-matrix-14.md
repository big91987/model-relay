# 执行14 需求—旅程—状态—测试—证据矩阵

版本和冻结材料摘要以 acceptance-report-14.md / evidence/attempt-14/source-current.json 为准。H=`evidence/attempt-14/host`；UI-E2E均来自正式宿主执行13，QA没有亲自重跑。截图层结果单独记录；中文状态是展示映射，不改API原值。

| ID | 基线要求 | 角色/范围 | 起始状态 | 连续用户步骤 | 最终事实 | 证据类型 | 结果 | 证据位置 |
|---|---|---|---|---|---|---|---|---|
| B01 | 当前版本匹配 | 候选工作树 | main+未提交改动 | 复算源码戳、对照before/after | 152件65201f…相同 | Static Review | Pass | evidence/attempt-14/audit-summary.json、source-current.json |
| B02 | 冻结材料不失配 | design-v0.2.0 | 原manifest | 核archive和引用文档摘要 | sha一致、状态未改 | Contract | Pass | evidence/attempt-14/audit-summary.json/material |
| G01 | 固定完整make verify | 隔离宿主 | 冷npm/候选目录 | 原入口全部阶段 | exit0、227Node/race/HTTP/build/browser | Contract | Pass | H/state.json、verify.log |
| G02 | 清理、无源修改 | 同候选 | 私有进程组 | finally及源码复算 | source/candidate unchanged、owned group absent | External State | Pass | H/state.json |
| T01 | 已知中文、未知保留 | 全页面字段 | 同API值 | 状态/角色/来源映射与未知输入 | 原值未重写、不伪报已知 | Contract | Pass | tests/browser/workspace.test.mjs、H/verify.log |
| T02 | 核验状态可理解 | 平台管理员/来源 | 首次self_hosted | 核验失败、纠正后再核验 | 中文失败/通过，精确原因与revision | UI-E2E | Pass | H/browser/self-hosted-{1280-48143,1440-48273,390-48450}.json |
| L01 | 表格/长值可触达、根无溢出 | 当前租户/30页面 | J01–J07之后 | keyboard→scrollend→唯一wheel、精确末列；完整值region键盘 | 三视口DOM/尺寸/完整文本事实满足 | UI-E2E | Pass | H/browser/visual-{1280,1440,390}/review.json |
| L02 | 三视口实际视觉可读 | QA/1280 | 原生当前截图 | 看十页和右列补图 | Key图空白、usage主体消失、billing偏移 | Static Review | Fail | visual-review-14.md；H/browser/visual-1280 |
| L03 | 三视口实际视觉可读 | QA/1440 | 原生当前截图 | 看十页和补图 | tenants主体偏移/右列单色；limits侧栏偏移 | Static Review | Fail | visual-review-14.md；H/browser/visual-1440 |
| L04 | 390实际视觉可读 | QA/390 | 原生当前截图 | 看十页/服务右列/首次分支 | 已查看表头/标识不字母内断词、关键动作局部横滚 | Static Review | Pass | visual-review-14.md；H/browser/visual-390及self-hosted390图片 |
| S01 | 独立新安装首次网页配置 | 平台管理员 | 三安装upstream/model/tenant均0 | 页面建self_hosted→模型→租户授权→发布价格 | 来源类型/模型/授权从页面保存 | UI-E2E | Pass | H/browser/self-hosted-*.json、tests/browser/journeys.mjs:selfHostedJourney |
| S02 | 不可达/认证错误恢复 | 同管理员同来源 | 无核验 | 连接关闭→修改401→修改正确凭据→核验 | failed/code/null或401→success/200、旧核验失效 | UI-E2E | Pass | H/browser/self-hosted-*.json及390失败/verified PNG |
| S03 | 真实角色及Key | 受邀管理员/开发者 | 新租户无Key | 网页邀请接受→成员→签发一次性应用Key | 真实服务器身份，Key诊断分配 | UI-E2E | Pass | H/browser/self-hosted-*.json、journeys-14.md |
| S04 | 普通/SSE真实调用 | owner普通/developerSSE | 同Key同tenant/model | 页面发送→同请求详情→同ID日志 | HTTP200/SSE DONE、正常终态/Key租户一致 | UI-E2E | Pass | H/browser/self-hosted-*.json normal/sse-same-request |
| S05 | 同请求费用事实 | owner只读核对 | 普通/SSE各1次 | GET详情/日志/limits/usage核对页面 | 每次3/4、费21成本7、合计6/8/42/14、pending0 | Data Reconciliation | Pass | H/browser/self-hosted-*.json normal/sse-ledger、tests/browser/journeys.mjs |
| S06 | 真实HTTP受控自托管 | 隔离HTTP fixture | 临时端口 | 核验GET/生成POST/SSE | Go服务实际dispatch、非产品fake默认 | API-E2E | Pass | tests/browser/journeys.go、journeys_test.go、H/verify.log |
| S07 | 首次分支三视口安全可审图片 | QA/1280、1440、390 | 30张分支图 | 查看关键长值/角色/Key | 1440roles单色、服务及1280Key主体缺失 | Static Review | Fail | visual-review-14.md、evidence/attempt-14/png-audit.json |
| R01 | J01–J07原严格网页回归 | 平台/租户/开发者/财务 | 每视口独立安装 | 见journeys-14 | 接入/权限/轮换/撤销/预算/租户/账单/导航均保留 | UI-E2E | Pass | H/browser/continuous-1280-48108.json、continuous-1440-48209.json、continuous-390-48299.json |
| R02 | summary/当前周期口径 | 平台管理员/当前租户 | 跨期隔离时钟 | 点模型/预算入口、统计对API | 数量/负责人/价格/延迟真实；闭合账单不改 | Data Reconciliation | Pass | H/browser/management-summary-{1280,1440,390}.json、visual-activity-*.json |
| R03 | R30-SYNC-01/错误/竞态 | 三视口编辑/生成 | 未知写结果/保存读取失败 | 丢响应、读回503/404→200、双提交、取消、中断 | 不重复签发/秘密重放、不重复dispatch、迟到隔离 | UI-E2E | Pass | H/browser/fault-{1280,1440,390}.json、editor-interruption-*.json、model-save-readback-*.json |
| R04 | 费用/预算UTC/角色/迁移 | 独立Go临时数据 | schema1/2及边界夹具 | race/HTTP、UTC边界、生产CLI迁移、native身份恢复 | 原断言保留且退出0 | API-E2E | Pass | H/verify.log migration/race各阶段 |
| P01 | 秘密不进入安全截图 | 三视口实际操作 | 一次性秘密关闭后 | 清空敏感输入/对话、验证storage空、查看49图 | 清除门槛通过；未见完整秘密 | UI-E2E | Pass | H/visual review、self-hosted脚本、visual-review-14.md |
| X01 | 独立QA动态原生浏览器 | QA | 沙箱监听限制未变 | 未执行另一条动态链 | 无独立动态结果；宿主仅替行为门禁 | UI-E2E | Not Run | acceptance-report-14.md §5 |
| X02 | QA再执行固定make verify | QA沙箱 | 稳定限制 | 不重复必失败入口 | 宿主同版本可用，本条不改绿 | Contract | Not Run | acceptance-report-14.md §5 |
| X03 | Owner preview已登录实操 | Owner/真实预览 | 既有main | 未执行 | 无当前已登录原生QA | UI-E2E | Not Run | acceptance-report-14.md §5 |
| X04 | 真实供应商 | 外部supplier | 无凭据/响应 | 未联调 | 受控HTTP不能替代 | External State | Not Run | acceptance-report-14.md §5 |
| X05 | 实际支付 | 外部支付 | 无条件 | 未执行 | 测试计价不是付款 | External State | Not Run | acceptance-report-14.md §5 |
| X06 | 外部身份 | 外部认证 | 无条件 | 未执行 | 内部邀请不是外部验证 | External State | Not Run | acceptance-report-14.md §5 |
| X07 | 新发布/PR/精确合并/部署 | 后续Connector | QA未放行 | 未执行 | 没有发布事实 | External State | Not Run | acceptance-report-14.md §9 |

QA58-01的分项改善不得把L02/L03/S07 Fail覆盖；R28-04历史根因未知不由R04当前Pass关闭。No-Go依据是明确强制视觉验收缺口，不是X04–X06外部未执行自动阻塞。
