# QA35 当前分层验收矩阵

版本design-v0.2.0；HEAD98511771871cf0951ecef716bbb55deb13e18da5，124文件指纹50fb764aac5443687e376c18906b36b765f02e9b05788fa4d5ac73617cbc98a4。当前报告acceptance-report-35.md；材料完整摘要和实际路径见evidence/attempt-35/material-integrity.json。上一矩阵不覆盖，旧119文件结论仅属历史。

证据缩写：H34=evidence/attempt-35/host-seq34-review.md及平台seq34完整归档；C=tests/browser/console.mjs；J=tests/browser/journeys.mjs与evidence/attempt-35/host-seq34三个JSON/15截图；G=internal/relay/console_contract_test.go、console_http_test.go、console_v3_test.go、console_process_test.go；L35=evidence/attempt-35/qa-independent-checks.log；R35=evidence/attempt-35/readback-recovery.test.mjs及readback-red.log；M=tests/browser/migration-console.mjs和Go正式CLI迁移runner。代码仅定位，Pass需对应H34/L35真实执行。受控HTTP/SSE不是供应商证据。

## AC01–AC18：本轮实际范围

同一AC拆成不同层/分支，Pass只对该行明确场景有效；遗漏的强制分支仍Not Run，不能把子场景拼成完整Go。J01–J07是冻结主链，不包含AC要求的所有异常组合。

| ID | 基线要求/本行范围 | 角色/范围 | 起始状态 | 连续步骤 | 最终事实 | 证据类型 | 结果 | 证据位置/边界 |
|---|---|---|---|---|---|---|---|---|
| AC01-UI-main | A十页、多对象、三视口、键盘创建和错误恢复 | 四角色/两租户 | 空安装后多对象 | 导航/创建/校验/Tab提交/编辑/详情/移动菜单 | 当前页可操作、根无横向溢出 | UI-E2E | Pass | H34+C三视口、J01键盘创建、J03列表；QA查看Key/账单截图，非每页像素比较 |
| AC01-UI-full-review | 十页批准原型逐页主要信息/动作及所有关键失败态 | 各角色 | 完整业务对象 | 全页原型对照 | 无私自减少结果 | UI-E2E/Static Review | Not Run | 当前截图集中Key/账单，逐页完整视觉验收尚不齐；局部对照不能写全量通过 |
| AC02-UI-service | 正式服务保存失败→保留草稿→修正→核验→发布 | 平台 | 无服务 | 非法URL400→修正→真实GET核验→模型 | 无失败半对象、秘密清除、核验applicable | UI-E2E | Pass | J01、H34；真实供应商/自托管端点另Not Run；当前UI明确选择provider，其他来源不能据此推定 |
| AC02-API-revision | 配置修改使旧核验失效/迟到不覆盖 | 平台 | 旧revision核验 | 更新→旧结果返回→重新核验 | 当前配置只适用当前证据 | API-E2E/Contract | Pass | H34 TestConsoleServiceCheckRevisionContract、ConsoleEvidenceConfigurationRace及连接晚到契约 |
| AC03-UI-intersection | Key授权收窄、零dispatch拒绝、修正不改历史 | 租户管理员 | 两授权模型/旧请求 | 目录→编辑Key→403→修正→新调用 | model_forbidden零dispatch；旧快照/账目保留 | UI-E2E/Data Reconciliation | Pass | J02、H34；默认空grant/非法范围补G实际API正反例，新增映射场景不泛化 |
| AC04-UI-people | 正式邀请接受、分配、降权收回、最后管理员、成员移除 | 平台/租户/开发者/财务 | 尚未接受邀请 | 各人真实登录→分配→409保护→降权→移除→旧会话刷新 | 原身份失权、应用Key继续调用、诊断分配清空 | UI-E2E | Pass | J01/J05；不是前端角色切换；服务端身份/范围与直接API拒绝由H34 G PeopleScope另证明 |
| AC04-API-scope | 多角色和跨租户列表/详情/写入/导出边界 | 四真实用户 | 两租户多应用 | 绕开UI直接请求 | 404/403、不带其他应用事实 | API-E2E | Pass | H34 TestConsolePeopleScopeActualHTTP；L35同handler受控transport。具体断言范围见G，不把隐藏菜单当隔离 |
| AC05-UI-pause | 暂停所有本测试租户Key普通/SSE、另一租户不受影响、恢复守恒 | 两租户/多Key | 已有费用、用量 | 暂停→普通/SSE/备用Key403→另一租户200→恢复→两Key200 | pause前后calls/fee/reserve/pending/C相同，拒绝零dispatch | UI-E2E/Data Reconciliation | Pass | J05三视口+H34 PauseUsageActualHTTP+L35；不声称暂停撤销已dispatch费用 |
| AC06-UI-secret | 租户创建/一次展示/关闭/刷新/退出不恢复 | 租户管理员 | 无目标Key | 创建→真实调用→关闭→刷新/退出再登录 | 安全metadata无完整secret，存储空 | UI-E2E | Pass | J01/J03/J07、C秘密清理、截图预条件；重启/持久安全由G真实进程与身份Key测试补证，不把API行代替UI |
| AC07-UI-generation | 立即、1h/24h、过渡禁止叠加、提前结束、启停、过期、撤销 | 租户管理员 | 已用Key/旧秘密 | 轮换→新旧分别调用→409→结束→旧401→状态/历史 | 共池不清零、秘密不重取、历史保留 | UI-E2E/Data Reconciliation | Pass | C立即轮换；J03三视口；精确截止Contract由H34/L35 GenerationExactDeadline单列，仅该测试是非网络身份断言 |
| AC08-UI-issued-readback | 已提交签发丢响应后首次安全读回成功 | 平台 | 无目标Key | POST成功后abort→未知→安全读回→刷新 | 对象只有1个、未再次签发、秘密不可恢复 | UI-E2E | Pass | C:137+H34三视口，仅首次读回成功 |
| AC08-CONTRACT-readback-failure | 未知结果第一次安全读回失败后保持可重试 | 当前生产editor函数 | 模拟写入丢响应和临时读故障 | submit→安全读回→失败→找重试 | 应保留重试入口及安全操作标识 | Contract | Fail | R35 exit1：retry=false、operation-visible=false；QA35-01 P1。DOM替身不是真实UI |
| AC08-UI-required-faults | 轮换丢响应、读回失败/恢复、快速双提交 | 管理员同会话 | 有应用Key/未知操作 | 真实页面注入故障→重复点击→读回失败→再读回 | 不重放秘密/操作、事务只一次、事实可安全确认 | UI-E2E | Not Run | J03仅成功轮换，C仅成功读回；QA22-01残余，不用Contract拼成该行Pass |
| AC09-API-current | tenant五维各0/null、Key Calls=0、单Key并发、无部分扣计 | 当前API/应用 | 有界候选/空用量 | 分别设0→429→恢复→在途竞争 | 零dispatch、tenant未部分扣计、600原子预留 | API-E2E | Pass | H34 ResourceDimensionsActualHTTP+ActualHTTPQuotaBillingAndSSE；L35对应两场景 |
| AC09-API-required-combinations | Key RPM/TPM/C/费用分别独立命中、多Key/双代竞争同tenant资源 | 两租户/多Key/两代 | 有独立正有限上限和已用量 | barrier争用→拒绝→修改恢复 | 各scope预算/Token/count/C守恒，拒绝无部分扣计 | API-E2E | Not Run | G resourceDimensionScenario:354–427只对tenant轮五维，Key单独Calls；8并发用同一Key。未见这些组合的当前实际断言 |
| AC10-UI-budget | 已用费用耗尽→硬预算429→提高不清零→重新成功 | 租户应用/平台限额 | 接近阈值 | 页面改预算→拒绝→提高→同ID详情 | tenant_budget_exceeded零dispatch、已有费用保留 | UI-E2E | Pass | J04，G有界三尝试/未知预留补API证明；不将受控上界当真实供应商可信声明 |
| AC11-API-recovery | 跨月/回拨保留pending、取消/崩溃恢复不重放和释放守恒 | 应用/运维 | 已受理/待确认 | 跨月→回拨→实际子进程崩溃→重开→核对 | 历史不改、并发释放一次、未知保守预留 | API-E2E/Contract | Pass | H34 GovernanceActualHTTP、RealProcessCrashRecovery、HTTPRestartAndBackup；L35回拨/幂等；不泛化为所有任意崩溃点覆盖 |
| AC12-UI-price | 旧请求旧价、新请求新价、成本与租户价分开 | 平台/租户 | 旧价请求 | 发布新价→调用→看旧/新详情 | fee14→21、cost7，旧快照不变 | UI-E2E/Data Reconciliation | Pass | J01/J06、H34；微美元精度/舍入由L35 MicroDollarPrecision与G补Contract |
| AC13-API-attempt-ledger | 同ID失败已知→未知→成功，追加差额与幂等/过期拒绝 | 平台/三个受控上游 | 不同候选快照 | 三dispatch→pending→跨窗→核对→重复/低值/旧revision | 原项/快照不改，delta fee5/cost16，原bill19/60 | API-E2E/Data Reconciliation | Pass | H34 GovernanceActualHTTP+L35；不能把未知revision0的0费用误称完整已结算 |
| AC13-UI-reconciliation | 单次未知调用正式页面核对 | 平台 | 真实未知请求 | 找同ID→核对→读回 | 累计fee21/cost7/Tokens3+4，entries2，pending/reserve0 | UI-E2E/Data Reconciliation | Pass | J06及实际assertReconciledLedger；未知三尝试界面的完整逐项操作不由该单次场景替代 |
| AC14-UI-bill | 当前/pending拒结→结束核对→结账→更正→导出与角色隔离 | 平台/财务/两租户 | 已有旧价和pending | 实际页面结账/核对/更正/导出 | 原金额保留、原因/调整期+5、另租户404 | UI-E2E/Data Reconciliation | Pass | J06；G补close重复、负更正、同幂等ID一次；不同并发close/correction组合不泛化 |
| AC15-UI-scope | 真实>25调用、过滤页2、详情导航、切scope迟到隔离、错误重试 | 租户管理员/平台 | 当前Key历史超过一页 | 筛选→页2→详情跳日志→切租户→旧结果释放→503→重试→导出 | 当前scope行数/ID精确、错误不是空、秘密刷新不恢复 | UI-E2E | Pass | J07三视口，故障route只延迟/注入读取失败，成功数据来自Go；非VM拼接 |
| AC16-UI-main | 普通/SSE输出→取消→同ID，无自动重发 | 平台调用 | 有真实Key/受控上游 | 普通→SSE→输出后取消→日志 | 实际响应/本地停止、费用可能待确认、不重放 | UI-E2E | Pass | C三视口，H34真实HTTP/SSE/路由补后端终态；J普通调用每次单次dispatch |
| AC16-UI-double-submit | 真实页面快速双击/Enter双提交只生成一次 | 当前新控制台 | 可调用且有延迟响应 | 快速重复用户动作→等待→同ID与上游计数 | 一次准入/dispatch，按钮和取消状态正确 | UI-E2E | Not Run | C/J均未执行快速双提交；state.call静态guard不是UI回执；QA22-01残余 |
| AC17-API-security | 服务端身份/CSRF/SSRF、元数据/审计隐私 | 四角色/应用 | 真实认证与敏感输入 | 直接越权/非法源→安全详情/CSV/审计 | 拒绝不泄secret/完整对话，敏感写审计 | API-E2E/Contract | Pass | H34安全旧回归、PeopleScope、DeniedMetadataAndStorage及身份Key持久测试；截图/浏览器存储另由C/J断言 |
| AC18-UI-CLI-migration | 正式CLI schema1/2→3、旧身份/Key/历史守恒、备份/恢复失败闭环 | 运维/存量应用 | 版本化旧UI夹具 | 旧UI建数据→正式CLI→新UI/调用→备份恢复 | 原身份/Key/历史保留、未知不补零；失败不迁移 | UI-E2E/API-E2E/Contract | Pass | H34+M三视口、Go迁移/CLI/backup/restart；旧schema1→2兼容测试单列，不当正式部署 |
| QA-LOCAL-GATE | QA原生固定入口监听限制 | 沙箱 | 同环境限制 | 本轮未重复已知失败 | 不越权/不删门禁 | Contract | Blocked | 原执行22 evidence/qa-make-verify.log；本轮同版宿主H34为门禁替代执行，Blocked原用例不改Pass |

## H01–H08与外部事实

| ID | 当前责任/事实 | 证据类型 | 结果 | 证据/具体未决 |
|---|---|---|---|---|
| H01 | Issue25版本包、prepare入口、archive与19文件同摘要 | Contract | Pass | 会话prepare/Issue回执及QA35 material-integrity.json |
| H02 | 本次Issue→本Run已承接；唯一性/重复事件不得重建 | External State | Not Run | 原Run关系有回执，重复入站实际演练无可核对证据；协调者负责 |
| H03-material | 各阶段/QA实际阅读同冻结规则与路径 | Contract | Pass | architecture-r1/研发交接及本轮具体对照/摘要，不只是“已阅读” |
| H03-runtime | 指定Codex/gpt-6.1-sol实际原生模型核对 | External State | Not Run | 本节点无原生执行模型回执；配置/自报不替代 |
| H04 | 同批准基线交接及错摘要/缺文件/旧版本安全恢复 | External State | Not Run | 同材料handoff已完成，材料负向平台恢复无完整真实回执 |
| H05 | Agent源码与设计全部兑现、实际固定门禁/旅程 | UI-E2E/Contract | Not Run | 124文件与当前宿主绿有效；QA35-01和强制缺用例仍阻断完整符合声明 |
| H06 | 独立检出缺口→原节点整改→同版原层复验 | Contract | Pass | QA22返研、seq34新主链、QA35关P2/发现P1/正式返研；不表示所有缺陷已闭合 |
| H07 | 同PR精确head/base/merge、Actions/Deployment/health与效果 | External State | Not Run | 尚未发布，本节点不推送/合并/部署；8793是编排端口而非产品效果 |
| H08 | Harness缺陷唯一维护源/模板安装升级原入口复验 | External State | Not Run | 维护者责任，当前无完整证据；不补造平台修复事实 |
| EXT-PROVIDER | 真实供应商/真实已有推理端点 | External State | Not Run | 无凭据；受控HTTP/SSE只能证明本项目语义，不阻断当前内部修复 |
| EXT-IDENTITY/PAYMENT | 外部身份/支付 | External State | Not Run | 当前为真实内部身份和USD测试分账，外部不在默认实现承诺 |

## 商用能力五层对照

不声称全面对齐New API/LiteLLM；参考仅沿用冻结资料公开功能/交互，不读取其实现。每一层状态分开，宿主“Pass测试集”不等于能力全验收。

| 能力 | 已实现/静态评估 | 局部测试 | 宿主完整make verify | 独立QA（当前范围） | 已发布 |
|---|---|---|---|---|---|
| 服务/渠道接入与核验 | 正式Go/UI入口 | Pass保存/核验/旧revision | Pass seq34测试集 | Pass J01受控provider；真实端点Not Run | Not Run |
| 模型发布/授权 | 目录/授权交集 | Pass API正反例与J02 | Pass seq34测试集 | Pass J01/J02列明场景 | Not Run |
| Key创建/轮换/撤销 | 代际/元数据/共池入口 | Pass生命周期；Fail读回恢复契约 | Pass seq34既有测试集 | Fail QA35-01 Contract；异常UI Not Run | Not Run |
| 身份/成员/租户隔离 | 真实内部身份/角色 | Pass邀请/降权/隔离/暂停 | Pass seq34测试集 | Pass J01/J05+PeopleScope具体边界 | Not Run |
| 额度/费用预算 | 独立资源/原子预留 | Pass已覆盖维度/硬预算 | Pass seq34测试集 | Pass J04；多Key/双代完整组合Not Run | Not Run |
| 定价/计费/账单 | 快照/追加账本/结账更正 | Pass金额/幂等/多尝试 | Pass seq34测试集 | Pass J06及API列明事实 | Not Run |
| 用量日志/分析/诊断 | 范围过滤/分页/导出/同ID | Pass实际分页及故障恢复 | Pass seq34测试集 | Pass J07；分析全部组合不泛化 | Not Run |
| 表格/筛选/编辑/详情/移动 | A真实后台/三视口 | Pass关键操作/键盘；列表状态修复 | Pass seq34测试集 | Pass主链；全页视觉与双提交Not Run | Not Run |

冻结外支付税务、多协议、自动秘密分发、外部身份、生产SLO/跨实例协调仍为后续差距，不默认纳入整改。本轮No-Go依据内部强制AC/P1，不以外部凭据或发布尚未发生作为研发实现缺陷。
