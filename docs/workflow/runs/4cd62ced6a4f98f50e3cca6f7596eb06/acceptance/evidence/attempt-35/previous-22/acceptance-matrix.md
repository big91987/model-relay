# 独立QA当前分层矩阵 — 执行22

本矩阵是本轮独立结论；上游根目录acceptance-matrix.md保留需求映射及历史，不覆盖已测试源码指纹所包含的原文。材料与候选摘要见acceptance-report.md。所有Pass均限定所列场景；不把局部Pass加总为完整AC通过。

证据缩写：H=平台seq21固定make verify及evidence/host-seq21-review.md；L=evidence/qa-independent-checks.log；B=tests/browser/console.mjs真实浏览器动作；G=internal/relay/console_contract_test.go、console_http_test.go、console_v3_test.go、console_process_test.go；M=tests/browser/migration-console.mjs及迁移Go runner；S=evidence/key-status-review.json。源码是定位，不单独构成行为Pass。Host UI的scope为受控真实Go服务，外部服务不是供应商。

## AC完整结果及缺口

| ID | 基线要求 | 角色/范围 | 起始状态 | 连续用户步骤 | 最终事实 | 证据类型 | 结果 | 证据位置/未完成原因 |
|---|---|---|---|---|---|---|---|---|
| AC01 | A工作台多对象/三视口/键盘 | 所有人员角色 | 多对象 | 表格/编辑/抽屉/完整键盘 | 可高效完成业务 | UI-E2E | Not Run | B十页及空编辑Tab通过；全键盘业务/逐页对照未全覆盖；QA22-02列表信息偏差 |
| AC02 | 接入/核验/发布及失败恢复 | 平台管理员 | 首次无服务 | 保存失败→修正→核验→改配置 | 旧证据失效，真实状态 | UI-E2E | Not Run | B正常接入通过；新workspace保存失败/草稿修正未执行 |
| AC03 | 默认无授权，交集且零dispatch | 平台/租户管理员 | 新模型/Key | 授权收窄/拒绝/修正 | 范围不扩大 | UI-E2E | Not Run | B显式授租户成功；J02实际网页收窄恢复缺失 |
| AC04 | 人员身份/诊断/降权/隔离 | 四角色/两租户 | 邀请未接受 | 接受→分配→降权/移除 | 旧身份失权、应用Key存续 | UI-E2E | Not Run | B成员接受登录通过；网页诊断分配/收回/跨租户导出未跑 |
| AC05 | 暂停全部Key/另一租户独立 | 平台/两租户 | 均有效 | 暂停→拒绝→恢复 | 历史/额度不变 | UI-E2E | Not Run | B无暂停/恢复；API证据不替代J05 |
| AC06 | 一次性受限秘密、刷新不可取回 | 租户管理员 | 无Key | 创建→使用→关闭/刷新/重启 | 秘密不可重取 | UI-E2E | Not Run | B平台签发/转交/关闭清理通过；租户签发和刷新重启UI未全覆盖 |
| AC07 | 期限/轮换/共享池/提前结束 | 租户管理员 | 有用量Key | 立即/1h/24h→结束→撤销 | 代际边界与历史正确 | UI-E2E | Not Run | B仅立即轮换；J03过渡与共享事实未走网页 |
| AC08 | 未知写入安全恢复 | 签发/轮换管理员 | 丢响应 | 安全读回/失败/再恢复 | 不重复创建/伪报成功 | UI-E2E | Not Run | B已提交后丢签发响应成功读回；轮换未知/读回失败缺实际浏览器用例 |
| AC09 | tenant/Key独立限制与竞争 | 多Key/两租户 | 已设资源 | 调整/并发/零和无限 | 无部分扣计 | API-E2E | Not Run | H/L验证多个独立维度及竞争；跨Key/跨租户全部组合未逐一证明，不能只凭测试名整体Pass |
| AC10 | 原子有界费用预算 | 多Key/多尝试 | 接近硬预算 | 耗尽→提高→重发 | 保守预留与实际结算 | UI-E2E | Not Run | G有预留/未知/竞争证据；B仅次数429，不是费用预算恢复 |
| AC11 | 改限制/轮换/跨窗/重启不清零 | 应用及运维 | 历史/未决 | 调整→跨窗→崩溃→迟到 | 不重复计量或释放 | Contract/API-E2E | Not Run | H真实进程恢复及G跨月/核对通过；所有指定回拨/迟到竞争点覆盖尚未逐条闭合 |
| AC12 | 独立价/精度/受理快照 | 平台/费用角色 | 有旧价请求 | 改价→新调用→旧价详情 | 历史不回算 | UI-E2E | Not Run | B首次调用前定价；网页改价前后快照对照缺失 |
| AC13 | 全dispatch已知收费、未知待确认 | 平台/多尝试 | 失败已知→未知→成功 | 同ID核对→重复/迟到 | 费用及终态守恒 | API-E2E | Not Run | H/G受控三尝试及幂等追加通过；全部迟到/用量形态不据概要全Pass |
| AC14 | 账单核对/结账/追加更正 | 平台/租户/财务 | 已结束及待确认周期 | 核对→拒结/合法结→更正→导出 | 原账、原因、调整周期保留 | UI-E2E | Not Run | H/G API结账更正通过；B没有该网页旅程 |
| AC15 | 搜索/分页/详情/导出与迟到隔离 | 各角色/租户 | 超一页记录 | 页2→详情→切租户→迟到→恢复 | 不串范围/不把错误当空 | UI-E2E | Not Run | B同ID详情通过；旧native页2不是新workspace旅程 |
| AC16 | 普通/SSE/取消/不重放 | 应用/诊断角色 | 授权Key | 发送→输出→取消→同ID | 状态/费用真实、不伪报免收费 | UI-E2E | Not Run | B平台普通/SSE/取消通过；其他角色完整末端与新页面双击覆盖缺失 |
| AC17 | 隐私/审计/服务端权限 | 全角色 | 含敏感输入 | 管理→降权/退出→审计 | 不泄对话或秘密 | API-E2E/UI-E2E | Not Run | H/L已有scope和秘密清理断言；全角色网页审计核对未跑 |
| AC18 | 正式迁移/恢复/持续运营 | 本地运维/原应用 | schema1/2存量 | init→upgrade→backup/restore | 旧资产守恒/失败安全恢复 | External State | Not Run | H本地CLI/三视口迁移及进程恢复通过；正式Actions安装发布及效果尚未发生 |

## 已执行局部证据（不拼接完整旅程）

| ID | 基线要求 | 角色/范围 | 起始状态 | 连续用户步骤 | 最终事实 | 证据类型 | 结果 | 证据位置 |
|---|---|---|---|---|---|---|---|---|
| AC01-UI-subset | 三视口十页/原13Tab/Escape | platform admin | 有数据/空创建编辑 | 导航→创建编辑→Tab→Escape | 焦点在弹层及返回原触发点 | UI-E2E | Pass | H+B；不是填写提交的全键盘路径 |
| AC02-UI-normal | 新服务与模型准备 | platform admin | 新租户无服务 | 保存→check→模型→授租户→价 | revision匹配且状态success/applicable | UI-E2E | Pass | H+B:95–118；POST输入execution revision实际校验 |
| AC04-UI-invite | 真实邀请/接受/登录 | invited administrator | 未接受资格 | 接受→登录→看范围→退出 | tenant正确、平台服务按钮不可用、session401 | UI-E2E | Pass | H+B:136；Start State为平台UI已签邀请 |
| AC04-API-subset | 降权/诊断/跨范围 | 多人员/应用 | 隔离受控夹具 | 管理API正反例 | 人员资格与应用Key分离 | API-E2E | Pass | H TestConsolePeopleScopeActualHTTP；L同场景InProcess，不能称L网络 |
| AC06-UI-subset | 秘密转交与清理 | platform admin | 新Key | 签发→秘密→显式转交→调用→关闭 | 一次性字段存在、值一致且清理 | UI-E2E | Pass | H+B:119–143；不记录秘密原值 |
| AC07-UI-immediate | 立即轮换 | platform admin | 有用量Key | UI轮换→关闭 | 新代际可用、旧代际401 | UI-E2E + API最终对账 | Pass | H+B:135；API /v1/models是Final Fact，不是隐藏UI步骤 |
| AC07-Contract-boundary | 精确24h截止与不叠加 | 管理/API | 已过渡Key | 前1ns→截止→再轮换 | 等于截止失效、不复活 | Contract | Pass | H TestConsoleGenerationExactDeadline；L独立重复 |
| AC08-UI-issuance | 丢签发响应安全恢复 | platform admin | 浏览器注入丢响应 | fetch已提交→abort→安全读回 | 未重复，列表仅1个新Key | UI-E2E | Pass | H+B:137；route仅丢响应不是产品默认上游 |
| AC09-API-dimensions | 零/无限/原子竞争 | tenant/Key | 受控多尝试fixture | 各维度零→解除→竞争 | 零dispatch/无部分扣计 | API-E2E | Pass | H TestConsoleResourceDimensionsActualHTTP；L InProcess重复 |
| AC10-API-reserve | 全候选上界/未知预留 | 应用/三尝试 | 有界模型/限额 | 准入→失败已知/未知→成功 | 保留未知费用和TPM | API-E2E | Pass | H TestConsoleGovernanceActualHTTP及ResourceDimensionsActualHTTP |
| AC11-API-process | schema3实际进程崩溃恢复 | 测试运维 | started/未决请求 | kill→restart→读回 | 不重放生成，未决守恒 | API-E2E | Pass | H TestConsoleRealProcessCrashRecovery；受控过程不是正式部署 |
| AC12-Contract-money | 微美元精度 | 计费 | 已知边界Token/费率 | 四舍五入边界/负数拒绝 | 计算符合断言 | Contract | Pass | H TestMicroDollarPrecision；L独立重复 |
| AC13-API-attempts | 多尝试计费及幂等核对 | 平台 | 已知/未知/成功 | 看详情→核对→重复 | pending/fee_reserved及追加账项准确 | API-E2E | Pass | H TestConsoleGovernanceActualHTTP；L同受控场景InProcess |
| AC14-API-bill-subset | 当前拒结/结束结账/幂等更正 | 平台 | 测试时钟生成账 | API关闭→更正两次→读回 | 原金额保留、一次调整 | API-E2E | Pass | H TestConsoleActualHTTPQuotaBillingAndSSE:168–184；L同场景InProcess |
| AC15-UI-detail | 同Request ID详情 | platform admin | UI新请求 | 调用→查看同ID | 详情关联当前请求 | UI-E2E | Pass | H+B:129–134 |
| AC16-UI-subset | 普通/SSE/取消 | platform admin | 有效Key | 普通→SSE→输出后取消→同ID | 实际200/输出/本地取消提示 | UI-E2E | Pass | H+B:129–134；上游独立受控 |
| AC17-UI-cleanup | 退出敏感清理 | platform admin | 已签发与调用 | 关弹层→退出 | secret/password控件无值 | UI-E2E | Pass | H+B:142；不代表整个隐私审计AC |
| AC18-UI-migration | schema1/2→3正式CLI本地迁移 | 原身份/应用 | 明确旧版测试fixture | 旧UI建数据→CLI升级→新UI读回 | 身份/Key/历史守恒 | UI-E2E/API-E2E | Pass | H+M三视口；schema1→2旧兼容测试另记，不冒充正式 |
| Q22-STATUS | 实际状态formatter | 停用/过期/有效Key | 静态抽取 | 传入revoked字段 | 三种状态无法区分 | Static Review | Fail | S；不是创建过期对象后的真实浏览器证据 |
| QA-LOCAL-GATE | 固定入口本地执行 | QA沙箱 | 同指纹 | make verify | 应允许真实监听 | Contract | Blocked | evidence/qa-make-verify.log exit2/bind denied；不能覆盖H的宿主exit0 |

## Harness与外部事实

| ID | 基线要求/范围 | 证据及具体边界 | 证据类型 | 结果 |
|---|---|---|---|---|
| H01 | Issue版本化材料/prepare | #25/prepare会话回执，19文件与archive摘要独立一致 | Contract | Pass |
| H02 | 唯一Run/重复入站 | 已有Issue/Run关系；重复事件和唯一性未有真实演练回执 | External State | Not Run |
| H03 | 读取材料+实际指定模型 | 本轮及上游有具体规则/路径/摘要；原生实际模型无可核对回执，配置说明不替代 | External State | Not Run |
| H04 | 同基线handoff/负向安全恢复 | 合法路径已承接；错摘要/缺文件/过期版本/损坏包恢复无完整实证 | External State | Not Run |
| H05 | Agent源码/设计/旅程完整符合 | 实际Go源码和119文件存在，宿主exit0；完整旅程缺用例，不整体Pass | UI-E2E/Contract | Not Run |
| H06 | 独立发现偏差/返回及复验 | 本轮独立矩阵/缺陷/正式返研；后续QA22-01修复尚无复验 | Contract | Not Run |
| H07 | 同PR合并/Actions/Deployment/health一致 | 尚未发布；不查旧Run或把8793当产品 | External State | Not Run |
| H08 | 真实平台缺陷唯一源正式升级复验 | 入口传递事实不替代维护源和升级回执；由协调者负责 | External State | Not Run |
| EXT-PROVIDER | 真实供应商/已有真实推理端点 | 无用户凭据；受控HTTP不算真实供应商 | External State | Not Run |
| EXT-IDENTITY/PAYMENT | 外部身份与支付 | 当前内部身份测试有效；外部未提供，非本轮内部产品阻断 | External State | Not Run |

## 商用能力五层对照

既有商业参考仅用于对照冻结范围，不声称全部对齐New API/LiteLLM，不读取其源码。保留上游八类及J01–J07；“局部测试”是明确子集，不是整体Pass。

| 能力 | 已实现/静态评估 | 局部测试 | 宿主完整make verify | 独立QA完整该能力 | 已发布 |
|---|---|---|---|---|---|
| 服务/渠道接入与核验 | 有正式入口 | Pass正常子集 | Pass现有测试集 | Not Run失败恢复UI | Not Run |
| 模型发布/授权 | 有入口/服务端grant | Pass部分API/UI | Pass现有测试集 | Not Run收窄恢复UI | Not Run |
| Key创建/轮换/撤销 | 有入口/代际池 | Pass立即/部分边界 | Pass现有测试集 | Not Run完整J02/J03；QA22-02 | Not Run |
| 身份/成员/租户隔离 | 有服务端身份 | Pass范围API/邀请UI | Pass现有测试集 | Not Run角色业务/暂停UI | Not Run |
| 配额/费用预算 | 有独立资源机制 | Pass原子API/次数UI | Pass现有测试集 | Not Run费用预算恢复UI | Not Run |
| 定价/计费/账单 | 有快照/追加账本 | PassAPI精度/核对/账单 | Pass现有测试集 | Not RunJ06 UI | Not Run |
| 日志/诊断/分析 | 有范围/详情/导出入口 | Pass详情/API/旧native | Pass现有测试集 | Not Run新workspace J07 | Not Run |
| 表格/筛选/编辑/详情/移动端 | 明亮A框架、列表差距 | Pass三视口入口及13Tab | Pass现有测试集 | Not Run逐页/完整键盘/所有关键旅程 | Not Run |

冻结范围外的支付税务、自动秘密分发、多协议、外部身份、生产SLO和跨实例协调继续沿用上游后续差距，不纳入本轮新增实现要求。
