# 需求—旅程—接口/状态—测试—证据矩阵

证据简称：H=[seq18宿主回执审核](evidence/host-seq18.md)；B=`tests/browser/journey.mjs`；V=[独立60VM](evidence/frontend-independent.log)；G=[独立3Go契约](evidence/console-contract-independent.log)；Q=[新增4边界失败第三轮](evidence/independent-boundaries-attempt-03.log)，对应[入口](evidence/independent-boundaries.mjs)。各行结果只适用所列证据层/子场景，AC需所有强制分支满足才能整体接受。所有行同当前56文件快照，不借用其他Run绿色。

| ID | 基线要求 | 角色/范围 | 起始状态 | 连续用户步骤 | 最终事实 | 证据类型 | 结果 | 证据位置 |
|---|---|---|---|---|---|---|---|---|
| AC01-UI | 六页导航/直达/刷新/前后 | 两屏管理员 | 登录有效 | 每页goto/reload，模型Tab导航/前后 | hash/aria-current/页面一致，无溢出 | UI-E2E | Pass | H/B六页循环 |
| AC01-C | 未知hash安全归一 | 前端所有权 | 非白名单hash | 进入页面 | overview且无敏感URL | Contract | Pass | V unknown navigation |
| AC02-UI | 登录显隐/管理401与应用401分离 | 两屏管理员 | 错误密码、重启后失效会话、停用应用key | 显隐→登录拒绝→登录→重启→重登录目标；生成401留管理台 | 密码清除，安全hash恢复，不重发 | UI-E2E | Pass | H/B rejectedLogin/restart/disabled |
| AC02-C | 长流known期限/pending/新登录隔离 | 前端会话 | 悬挂GET和已知expires_at | 到期/跨页/新登录→释放旧查询 | 清密/取消，不延期、不复活旧会话 | Contract | Pass | V expiry相关11组合；非原生浏览器 |
| AC03-UI | 1280/390布局、指定控件键盘/焦点/AX属性 | 两屏管理员 | 真实表单/调用 | 密码、发送/取消、复制示例、diagnostics、编辑取消、导航Tab | activeElement/:focus-visible/outline/role，整页无横滚 | UI-E2E | Pass | H/B keyboardAction及版本测试；归档两屏截图 |
| AC03-FULL | 设计要求全程键盘包括创建/启停/精确查询/分页 | 两屏管理员 | 隔离配置和足量记录 | J03-FULL-KEYBOARD | 全链路原生键盘最终事实 | UI-E2E | Not Run | journeys-01.md；QA-G01 |
| AC04-UI | health原值/未知/真实数量 | 两屏管理员 | 空/成功存量 | 登录→overview/health→刷新/受控异常恢复 | dev原值，0/已知计数，未知不阻塞 | UI-E2E | Pass | H/B runtimeVersion/loginOverview |
| AC04-G | 超1000聚合不截断/0与null | 管理API | 隔离SQL准备1006对象（非UI首配） | 查询overview/record | SQL全量口径，权限与故障正确 | Contract / Data Reconciliation | Pass | G OverviewAndExactRecords |
| AC05-G | 当前配置真实success/旧配置/重启/迟到隔离 | 后端 | 受控transport/持久record | 成功→改配置→迟到成功→恢复/终态修正 | 运行期证据与持久终态匹配 | Contract / Data Reconciliation | Pass | G EvidenceConfigurationRace |
| AC05-UI | 空/保存/成功区分 | 两屏管理员 | 首配/存量 | 保存→概览→普通调用→概览 | 保存不称核验，success显示ID/范围 | UI-E2E | Pass | H/B，归档概览 |
| AC05-C | 刷新后成功证据可操作核对同ID | 管理员恢复会话 | call_evidence存在、本地ID空 | 概览→“核对同请求记录” | 精确输入/查询应承载显示ID | Contract | Fail | Q QA-03；QA-D03，交互§3 |
| AC06-UI | 空配置全UI首调用闭环 | 1280管理员 | 正式Init后业务空 | J01 UI保存/创建/签发/普通/SSE/记录 | 真实Go/SQLite结果，同ID和流终止 | UI-E2E | Pass | H/B/main.go；无API/SQL预填 |
| AC07-UI | 创建/编辑/取消/签发关闭/密钥停用 | 两屏管理员 | 新建/已配置 | 冲突→改映射/取消→签发→关闭→停用 | 行状态与持久映射一致，取消零PATCH | UI-E2E | Pass | H/B |
| AC07-C | 失败清上游秘密，保公开输入 | 前端表单 | 非法地址和非空凭据 | 保存→400 | 公开地址保留，凭据应清除 | Contract | Fail | Q QA-01；QA-D01，交互§4 |
| AC07-C2 | PUT成功/GET失败分别告知 | 前端表单 | 写入前配置 | PUT200→GET503 | 明确已保存、最新读取未确认，不重写 | Contract | Fail | Q QA-02；QA-D02，交互§4 |
| AC08-UI | 一次密钥、存储/日志/截图排除 | 两屏管理员 | 新签发完整值 | 关闭→刷新→离页/退出→审计 | 完整值不再取得；存储/日志无秘密 | UI-E2E | Pass | H/B logAudit/storage/screenshot前断言；截图目视 |
| AC08-C | 迟到签发/clipboard不污染新页 | 前端 | pending签发或复制 | 关闭/重签/离页/新登录→旧完成 | 旧秘密/notice无晚到写 | Contract | Pass | V late issuance / clipboard |
| AC09-UI | 网关Base URL/占位复制零生成 | 两屏管理员 | 调用页 | 展开/键盘复制 | origin/v1、占位key/model、无真实key/POST | UI-E2E | Pass | H/B example；失败或成功中文分支均允许 |
| AC09-C | 复制失败手动路径 | 前端clipboard | 拒绝剪贴板 | 显式复制→失败 | 中文手动选择、不触发调用 | Contract | Pass | V clipboard failure；实际成功剪贴板并未单独证明 |
| AC10-UI | 启用选择、普通可读/SSE增量/工具摘要 | 两屏管理员 | 无依赖/合法配置 | 普通→diagnostics→SSE首段→release DONE→工具 | 文本/工具不伪造、不执行，未知usage可核对 | UI-E2E | Pass | H/B；受控上游 |
| AC11-UI | busy防重复、取消真实生效 | 两屏管理员 | 接收中流 | repeat requestSubmit→取消/释放 | 单次POST，canceled记录，上游context感知 | UI-E2E | Pass | H/B/main.go |
| AC11-C | 无DONE/事件error/旧finally | 前端流 | split UTF8/CRLF/异常流 | 读取→异常/取消→晚到 | 失败不完整、不重发、不改新请求 | Contract | Pass | V SSE及cancel isolation |
| AC12-UI | 离页/退出/会话不明清密，不重放 | 两屏管理员 | 流和秘密存在 | 网络检查失败→恢复检查→离页/退出 | 字段/输出为空，生成数不增 | UI-E2E | Pass | H/B session failure及sameRecord清除 |
| AC12-C | 忽略abort的JSON/stream/401/收尾隔离 | 前端所有权 | pending旧调用/会话 | cancel/跨页/新登录→释放旧任务 | 旧任务不能写DOM或解锁新任务 | Contract | Pass | V 多组合；不是原生网络演示 |
| AC12-C2 | 交互§7失败清key/message、保公开选择 | 前端生成 | 应用拒绝401 | 显式提交→拒绝 | 不登出管理员；秘密输入应清、模型保留 | Contract | Fail | Q QA-04；QA-D01 |
| AC13-UI | 普通/SSE/失败/停用密钥同ID和usage | 两屏管理员 | 当前页刚结束生成 | 查看本次记录 | ID/alias/keyID/state/0或未知对应 | UI-E2E | Pass | H/B sameRecord；duration精确值由G补强，非UI逐值对账 |
| AC13-C | 重登录概览同ID入口 | 前端记录页 | overview成功、本地ID空 | 点击核对 | 不丢精确身份，不用最新行顶替 | Contract | Fail | Q QA-03；QA-D03 |
| AC14-G | 安全拒绝/未知归属/并发/存储失败 | 管理API/应用 | 隔离合法/未知/停用key、关闭存储 | 受理/拒绝→精确查询 | 统一401/零上游、安全元数据、无伪归属 | Contract / Data Reconciliation | Pass | G DeniedMetadataAndStorage，H真实HTTP |
| AC14-C | 404/读取失败不借另一行 | 前端记录 | 不存在ID/读取故障 | 精确查询 | unavailable/手动定位，无假行 | Static Review | Pass | app.js findRecord（仅审查；需补原生异常验证） |
| AC15-UI | 地址错误/供应商拒绝/停用key中文 | 两屏管理员 | 输入错误/已停用key | 保存错误→修正；调用拒绝→查记录 | 安全文案，应用401不登出，无敏感上游message | UI-E2E | Pass | H/B失败及audit |
| AC15-G | 既有HTTP协议/来源/大小/网络/存储边界 | Go协议 | 隔离HTTP夹具 | 原HTTP/race全量 | 兼容错误结构/拒绝不访问上游 | API-E2E | Pass | H当前Go全量尾部；不是所有中文恢复UI已验 |
| AC16-UI | 显式恢复、取消零写/不自动请求 | 两屏管理员 | 地址错误/映射编辑/停用key | 修正地址、编辑取消、有效第二key新调用 | 新请求独立成功/记录，取消无PATCH | UI-E2E | Pass | H/B |
| AC16-REENABLE | 上游/模型/密钥停用→重新启用→页面新调用/错误恢复 | 两屏管理员 | 各依赖停用 | 页面启用再显式生成/记录 | 原对象恢复、不创建绕行、不重放 | UI-E2E | Not Run | QA-G01；现有Go配置启停不替代页面链 |
| AC17-S | 单一公共路线/授权/Trellis/原创交互 | 维护者 | 项目/Run文档 | 对照范围与路径 | 路线2–4独立接续，Go/单实例保持，未虚构发布 | Static Review | Pass | README/product-brief/roadmap/releases/local-deployment、.trellis/spec |
| AC18-H | 固定门禁/快照/失败保留 | 维护者 | 当前工作树 | 原样make verify→独立指纹/证据核对 | seq18退出0；旧失败/未知不改写 | Contract | Pass | H/baseline.log及delivery历史 |
| AC18-Q | 独立完整AC可接受 | 维护者 | 当前验收候选 | 审查→独立边界复现→整改交接 | 强制分支失败不得放行 | Contract | Fail | 本报告/矩阵；并非门禁失败 |
| SUPPLIER / DEPLOY | 真实供应商与正式部署后验证 | 获授权维护者 | 无凭据/无新发布回执 | 正式接入/发布后浏览器 | 实际响应/版本 | External State / UI-E2E | Not Run | 不在受控上游或main历史部署中推导 |

矩阵已标注未执行子项，不能把AC整体按通过率放行。QA-G01还包括上游同地址空凭据/变址缺凭据的页面恢复、模型启停导致选择失效、精确记录分页及读取故障、未知密钥拒绝页面路径；当前HTTP/VM补强不冒充这些UI结果。需要补充用例后复验，而不是修改基线降低门槛。
