# 逐 AC 分层验收矩阵 01

2026-10-06，全量验收；基线见 [报告](report-01.md)。结果只针对本行实际步骤，**Pass不代表整个关联AC通过**。H=[宿主seq16同版本固定入口](evidence/host-seq16.md)，Q=[QA独立核验回归exit1](evidence/connection-contract-01.log)，S=[390px实际概览](evidence/console-first-call-390.png)。代码和夹具入口均为仓库路径，不是外部发布证据。下文未执行项不得用代码阅读补绿。

## 需求—旅程—接口／状态—测试—证据

| ID | 基线要求 | 角色/范围 | 起始状态 | 连续用户步骤 | 最终事实 | 证据类型 | 结果 | 证据位置 |
|---|---|---|---|---|---|---|---|---|
| AC01-UI | 新建两个独立上游，不预填 | 管理员/同服务，每屏各一会话 | 1280先无路由A/B，390无本屏A/B | 概览→上游→新建A/填字段/保存→新建B/保存 | 实际POST201，凭据清空，两身份可被候选选择 | 真Go UI-E2E | Pass | H；tests/browser/routing.mjs创建步骤 |
| AC01-EDIT | 多对象独立取消、重名/空名、变址缺凭据 | 管理员/A和B | 两上游存在 | 编辑A错误→纠正/取消→B不变 | 回读及错误字段独立，取消无写 | 多上游UI＋API/数据负向 | Not Run | 原journey测单上游错误；现有routing不覆盖整组，不拼接Pass |
| AC02-UI | 合法列表才显示核验成功 | 管理员/A/B，两屏 | UI已保存 | 分别点核验→等待 | fixture返回缺object:list却显示成功，违反设计列表格式 | 真Go UI回执＋同版本夹具对照 | Fail | H；tests/browser/main.go列表响应；QA-01 |
| AC02-PROTOCOL | Content-Type/object/每项id | 管理员/隔离安装 | 独立新建上游 | handler显式check→受控返回→回读 | 5种非法列表success；坏JSONunknown；正确空列表success | QA实际Go handler/Transport/SQLite，不是UI | Fail | Q；internal/relay/qa_connection_contract_test.go |
| AC02-BUDGET | 核验10秒覆盖头与body | 管理员/隔离安装 | 已配置 | 显式check→读取出站deadline | 出站budget30s，不是10s | QA实际context预算断言 | Fail | Q；QA-02；未将预算探针说成真实等待测试 |
| AC02-FAILURE | 已知网络/协议/限流/超时失败，与未知分开 | 管理员/隔离安装 | 已配置 | 显式check→429/503/坏JSON/network→回读 | 四类已知失败unknown；API与持久化相同错误 | QA Go handler/SQLite | Fail | Q；QA-03；401正控Pass不关闭本行 |
| AC02-LIFECYCLE | 单项去重、取消、过期、超时、重入/纠错 | 管理员/双屏 | 延迟/失败列表 | 忙时取消、离页/过期、修改后显式核验 | 零重发、出站中止、旧结果隔离、状态正确 | 应补真HTTP/UI-E2E | Not Run | remediation-01；当前无取消核验入口，现有旧结果范围handler测试只覆盖其自身 |
| AC03-UI | 双候选映射/编辑/取消/非法权重纠正 | 管理员/同一屏会话 | UI建A/B | 模型→创建双候选→编辑weight0→修正→取消 | 公共名与各远端映射真实用于调用，非法草稿未提交 | 真Go UI-E2E/外部收到模型名 | Pass | H；routing.mjs；供应商fixture验证remote-model |
| AC03-BOUNDARY | 0～100整数优先级，1～100权重，引用及重复拒绝 | 管理员/配置API | 已有模型 | 提交各非法候选组合→回读 | 逐边界拒绝且零部分提交 | API/数据专项 | Not Run | duplicate原子拒绝由TestCandidateWriteRejectsDuplicatesAtomically/H覆盖；未找到全部数值/空远端/引用专项，不能扩为全部Pass |
| AC04-ROUTE | 1:3/1:1/优先级及候选停用真实选择 | 管理员/两屏及调用方 | 两完整候选 | UI改权重→40次；改1:1→20次；改priority10→10次；停候选→调用 | 实际30:10、10:10、0:10、停用无新访问 | 真Go UI操作/真实HTTP/外部计数 | Pass | H；routing.mjs计数断言；fixture control只改变外部行为/读计数 |
| AC04-DISABLE | 停上游/模型后新尝试止损 | 管理员＋应用 | 原配置有效 | handler/HTTP启停与切换边界 | 真实fallback前停用被过滤，旧公共协议停模型仍拒绝 | 实际HTTP/持久记录 | Pass | H；TestRoutingRealHTTPDisableBeforeFallback；TestHTTPConfigurationChangesAndReenable |
| AC05-HTTP | 白名单故障/不可切换/最多3个不同上游 | 应用/隔离HTTP | 2～4候选，API明确准备 | 生成→受控状态码→fallback/耗尽 | 白名单切换，其余零B；4候选只3次，无第四 | 真HTTP，不是UI | Pass | H；routing_http_test.go三项真实HTTP测试；API配置是边界补测，不代替创建UI |
| AC05-TIME | 整体预算不重置，取消与输出前计时 | 应用/缩短隔离超时 | 两候选 | 请求/超时/取消；首提交停止timer | 整体超时不再尝试，precommit超时允许一次切换 | Go context/handler/计时专项 | Pass | H；routing_boundary_test.go三项budget/stop回归；默认值仍按300/30/60，非耗时SLA |
| AC06-RECOVER | 冷却期跳过，到期被动恢复、无后台生成 | 管理员/两屏 | A失败，B成功 | UI调用→看同ID与冷却→恢复外部A→31秒后显式调用 | 冷却时零新增A，到期A计数增加，无列表检查或自动生成替代 | 真Go UI-E2E/实际HTTP/计数 | Pass | H；routing.mjs；TestRoutingRealHTTPSelectionFailoverAndRecovery |
| AC06-RACES | 启停不清冷却、全部冷却503、旧健康晚到、新配置及重启未知 | 管理员/并发与重启 | 故障/旧新执行代次 | 逆向操作/重启/延迟旧完成 | 手动状态不自动恢复，不污染当前健康，全部冷却零调用 | 待补状态/HTTP及必要UI | Not Run | 现有routing_test/routing_http仅普通冷却恢复；源码observe不代替竞态执行 |
| AC07-UI | SSE中断/管理员取消不重放 | 管理员/两屏 | 高优先级A、B可参与 | UI SSE→首包后断流；新SSE→取消→记录 | 不完整/已取消，B计数0 | 真Go UI-E2E/实际外部计数 | Pass | H；routing.mjs两个首包后停止路径 |
| AC07-HTTP | 现有普通取消、工具协议、输出失败/网络终止 | 应用/隔离HTTP | 合法密钥、受控stream | HTTP普通/SSE取消及协议回归 | 正在进行请求终止，参数/工具增量保留 | 真HTTP回归 | Pass | H；TestCancellation、TestSSEBarrierAndUsage、TestTransparentParametersAndLocalReject |
| AC07-MULTI-TOOLS | 多候选已提交工具增量/idle超时不访问B | 应用/双候选 | 工具流已首提交 | 工具增量→断流/idle/取消→核对B | 零备用/零重放且记录不完整 | 待补双候选真实HTTP专项 | Not Run | 单上游协议回归及content stream测试不自动等于此组合 |
| AC08-UI | 同ID逐尝试与已知+未知合计 | 管理员/两屏 | A503→B完成 | 结果→查看本次记录→精确ID详情 | 两尝试及1次未知，完整合计未知；不只报B3tokens | 真Go UI-E2E/记录/外部计数 | Pass | H；routing.mjs；TestRoutingTemporaryAllowlistAndUnknownTotals |
| AC08-RECOVERY | started/unknown/404/503/历史无明细，不替换为最新记录 | 管理员/两屏 | 真实服务/显式隔离故障注入 | 精确记录→读取/恢复/分页 | 不造用量/成功/最新记录；历史身份保留 | 真Go UI；人工夹具/注入只测声明分支 | Pass | H；journey.mjs QA recovery；migration.go旧行对账 |
| AC09-FLOW | 双屏路由配置→调用→故障→详情→恢复 | 管理员/每屏单会话 | 本屏新对象 | J-RUNTIME各屏连续步骤 | UI保存实际生效，实际计数、同ID详情和恢复 | 真Go UI-E2E（混合输入方式） | Pass | H；不是完整原生键盘通过，见下一行 |
| AC09-KEYBOARD | 每屏原生键盘连续创建/编辑/取消/核验/签发/发送/停止 | 管理员/两屏 | 本屏新对象 | 原生Tab及输入完成整个双候选旅程 | 可见焦点与真实变更全部成立 | 原生UI-E2E | Not Run | 75584仅创建候选type-ahead；routing编辑/签发/调用用fill/selectOption/check；旧单候选键盘不补双候选 |
| AC09-OVERVIEW | 已完成步骤可回看，下一步突出尚未完成 | 管理员/已有成功调用，390 | 完成首次调用 | 返回概览 | 完整1～5列表仍占当前下一步，没有设计要求的已完成折叠 | 当前宿主截图/同版本DOM与JS静态核读 | Fail | S；index.html setup-steps及app.js仅completed class；需真双屏回归，未冒充新点击实测 |
| AC10-UI | 原始恢复/写拒绝/部分成功只读恢复/会话清密 | 管理员/双屏单会话 | 已配置与受控故障 | 原journey键盘失败→修正→重读→过期→登录 | 无重复写/生成，明确未知，安全目标恢复 | 真Go UI-E2E/受控浏览器注入 | Pass | H；journey.mjs；注入503/旧响应显式范围，非供应商事实 |
| AC11-MIGRATION | 精确旧main真实非空旧UI→upgrade→旧backup rollback | 管理员/独立临时安装 | 精确85f6d17运行schema1 | 旧UI建上游/两别名/两keys启停/成功失败调用→backup→upgrade→新UI对象→旧restore/调用 | schema1→2，旧身份/历史/用量/加密凭据/key可用守恒；新增不在旧backup | 真实旧/新Go＋旧UI-E2E＋CLI＋SQL仅对账 | Pass | H；tests/browser/migration.go/legacy.mjs/migration-after.mjs；DEP-01，不是Actions |
| AC11-REFUSAL | 错key/锁/未知版本/失败原子/重复升级 | 运维/临时副本 | schema1/2临时安装 | 正式CLI与隔离边界错误输入 | 非迁移拒绝、完整性验证与幂等 | Go CLI/持久化边界测试 | Pass | H；upgrade_test.go/deployment_test.go |
| AC12-REGRESSION | 原普通/SSE/工具、安全鉴权/SSRF/限额/首配 | 管理员＋应用/隔离实例 | init空安装及受控上游 | 旧HTTP/Go/真浏览器回归 | 无默认供应商，非授权零外部，参数/密钥/恢复与fail-closed保持 | 固定入口Go race/HTTP/浏览器 | Pass | H；main_test.go/relay_test.go/boundary_test.go/journey.mjs |
| AC13-GATE08 | 固定make verify完整成功与当前产品身份 | 工程/08快照 | 82文件c340664…及模板联合身份 | tests Connector原入口900秒 | exit0，历史红控已正确分类，当前完整各级执行 | 宿主真实回执，QA独立身份核对 | Pass | H；version-lock-01.json |
| AC13-GATENEW | QA新增负控后修复快照门禁 | 工程/含QA新测试 | 83文件新快照 | 修复→宿主make verify→QA重验 | 必须新exit0与逐AC完整证据 | 尚未执行完整宿主复验 | Not Run | qa-source-01.json；新独立测试当前exit1；不能用GATE08关闭 |
| DEP02-04 | 真实旧/新Go＋可信controller顺序/失败/legacy | 维护者/隔离部署根 | 平台固定源及标准安装 | 受支持controller联合链 | 顺序/恢复/拒绝，真实data守恒 | 平台真实集成 | Not Run | 用户平台夹具/源复审不等于该链 |
| DEP05-INSTALL | 标准模板安装与manifest匹配 | 维护者/受管理文件 | 标准安装器 | 原授权install_workflow.py --upgrade及repeat | installed/unchanged；字节摘要与manifest一致 | 研发实际安装回执＋QA当前只读核对 | Pass | delivery/verification/platform-template-provenance-01.json；version-lock-01.json；不扩至所有控制器边界 |
| DEP06-07 | 正式Actions非空升级/可信隔离失败恢复 | 维护者/已授权正式路径 | 尚未本轮发布 | QA后发布链实际执行 | merge SHA/Deployment/health/browser/schema/data一致 | 正式环境 | Not Run | 设计明确不要求草稿PR前已发生；发布依赖保留 |
| SUPPLIER | 合法凭据真实供应商生成 | 管理员/合法预算 | 无供应商凭据/费用输入 | 取得授权后真实调用 | 供应商实际结果与费用边界 | 供应商 | Not Run | 不以受控夹具或历史发布关闭 |

## 核心用户旅程契约与实际结果

每屏单独browser context/page，由同一单管理员身份连续执行；实际Go服务是同一临时实例，两屏不是同一套对象的收尾截图。服务/外部上游由Go runner准备，**网关业务对象由UI创建**；控制端点只切外部行为/读取计数。调用方密钥从UI一次性签发，未API预置关键配置。UI注入/SQL历史对账的边界单列，不补步骤。

| Journey ID | User Story/AC | Actor | Start State | UI Steps | Expected UI Result | Final Fact | Cleanup | Result |
|---|---|---|---|---|---|---|---|---|
| J-FIRST-1280 | 原首配/AC10/12 | 单管理员1280会话 | init无有效上游，正式Go UI | 登录→上游错误/纠正/保存→模型→keys→调用→详情→过期恢复 | 首配/调用/同ID/清密与安全重入 | 外部调用与记录相符 | context.close；Go服务shutdown/临时目录defer清理 | Pass |
| J-FIRST-390 | 原首配/AC10/12 | 单管理员390会话 | 本屏新模型/key，共享已保存上游 | 本屏实际编辑/密钥/调用/错误恢复，不只截图 | 窄屏业务完整、焦点可见 | 当前key/请求/记录真实 | 同上 | Pass |
| J-RUNTIME-1280 | AC01/03～08运行子集 | 单管理员1280会话 | 无本屏A/B/公共模型/key | UI建A/B及模型→weight错误/取消→签key→权重/优先级调用→候选停用→503切换→同ID→冷却→SSE/取消 | 真实去向与诊断，未将核验成功合同计入本行 | 30:10/10:10/高优先级10次/备用零重放 | 同上 | Pass |
| J-RUNTIME-390 | 同上 | 单管理员390独立会话 | 无本屏A/B模型/key；1280对象共存 | 同样完整本屏任务 | 同上 | 同上 | 同上 | Pass |
| J-CHECK-1280/390 | AC02 | 每屏单管理员 | 本屏A/B已保存 | 显式列表核验→显示成功 | 只有合法列表才显示成功 | 实际缺object:list返回成功；不能按fixture绿色接受 | 同上 | Fail |
| J-KEYBOARD-1280/390 | AC09 | 每屏单管理员 | 新多候选业务 | 每步仅原生键盘连续完成 | 每个用户中间步骤及focus-visible验证 | 部分编辑/调用直接fill/selectOption；完整契约未执行 | 无新增执行 | Not Run |
| J-MIGRATE | AC11/DEP01 | 隔离旧/新单管理员 | 精确85f6d17旧binary/schema1空安装 | 旧UI建真实非空数据→backup/upgrade→新UI创建/调用→旧restore/旧调用 | 原身份/历史可用，新增不在旧backup | 2别名/2密钥/2历史，4真实受控POST；schema1→2→旧备份 | legacy进程停止，runner临时目录清理 | Pass |

J-RUNTIME通过不是J-CHECK通过；原首配键盘Pass不是J-KEYBOARD通过。各生命周期退出用停用而非硬删除，符合本轮范围；不引入多租户/配额/外部调度器验收。
