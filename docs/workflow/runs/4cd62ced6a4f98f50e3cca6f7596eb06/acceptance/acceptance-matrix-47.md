# QA47 当前逐项分层验收矩阵

候选visual-editor-owner-r1；HEAD98511771871cf0951ecef716bbb55deb13e18da5；129文件7e0b20ddbdea4585959365d8ad3b6121abaa91cd3f450fa2c6c21f7afc05a9ab。冻结design-v0.2.0 archive5f51d35fdc15ed7214749f9413879b4238097f1d6b6bbe2cc0d508d8e103cef7。真实源码/材料摘要见evidence/attempt-47/source-before.json及material-integrity.json。

证据别名：H46=`evidence/attempt-47/host-seq46-review.md`+平台seq46当前完整日志；A47=`evidence/attempt-47/artifact-index.json`当前75件产物；J47=`user-journeys-47.md`及host-seq46/continuous-{1280-34454,1440-34476,390-34501}.json；V47=`visual-review-47.md`及host-seq46/visual-{1280,1440,390}/review.json、PNG；F47=同目录fault-{1280,1440,390}.json；G=当前internal/relay/console*test.go等在H46执行；M=tests/browser/migration-console.mjs及正式CLI runner的H46记录。代码只提供定位，不作为单独运行证据。H46是宿主实际执行，不是QA亲自运行。

## 产品AC01–AC18

| ID | 基线要求 / 范围 | 角色/范围 | 起始状态 | 连续用户步骤或实际测试步骤 | 最终事实 | 证据类型 | 结果 | 证据位置 |
|---|---|---|---|---|---|---|---|---|
| AC01-UI-main | 十页入口、多对象/三视口/键盘 | 四真实角色/两租户 | 空安装到多对象 | 登录/创建/导航/编辑/详情/键盘提交/菜单 | 操作可完成，根无横向溢出 | UI-E2E | Pass | H46/J47/console；不是设计符合性总Pass |
| AC01-UI-design | 冻结主要信息/动作不删减 | 平台/当前租户 | 两模型/两租户/四Key | 实际查看30页并核对原interactions/原型 | 模型/服务/租户/工作台管理摘要缺项 | UI-E2E | Fail | V47；QA47-01 P1；QA38-01采集门槛Closed |
| AC02-UI-provider | 接入保存失败恢复/核验/发布 | 平台 | 无目标对象 | 非法URL→修正保存→核验→发布 | 无半对象，秘密清除，核验当前适用 | UI-E2E | Pass | J01/H46；受控provider非供应商联调 |
| AC02-API-source | 两来源及配置变更使旧核验无效 | 平台/受控端点 | self_hosted/旧revision | 接入核验→修改→旧结果晚到→新核验 | 来源持久，旧结果不覆盖 | API-E2E / Contract | Pass | H46 Governance/ServiceCheckRevision/EvidenceConfigurationRace |
| AC02-UI-self-hosted | 独立首次自托管来源网页分支 | 平台 | 无服务 | 选择self_hosted→保存/核验 | 无本轮单独完整分支回执 | UI-E2E | Not Run | API上述Pass不改写本行；共享编辑器provider链非独立分支 |
| AC03-UI | 新模型无隐式授权/交集/修正 | 平台/租户管理员 | 无授权模型后有Key | 发布→授权→收窄Key→403→修正→新调用 | 拒绝无dispatch，历史不变 | UI-E2E | Pass | H46/J01/J02 |
| AC03-API | 新身份/映射变更授权安全 | 应用/两scope | 已配置模型/Key | 直接越权、映射/组合变化后请求 | 403/零dispatch，无自动扩大 | API-E2E / Contract | Pass | H46 G/授权资产 |
| AC04-UI | 真实邀请/角色/诊断分配/退出 | 四角色/两租户 | 空租户新身份 | 邀请接受→登录→分配→降权/收回/移除 | 旧身份失权但应用Key仍服务，最后管理员409 | UI-E2E | Pass | H46/J01/J05 |
| AC04-API | 跨租户读写及导出隔离 | 两租户/四角色 | 真实身份与数据 | 直接ID、旧会话、筛选/导出越权 | 404/403不泄另一scope | API-E2E | Pass | H46 PeopleScopeActualHTTP/G |
| AC05-UI | 暂停全部所属Key/恢复 | 两租户/多Key | 非零用量 | 暂停→普通/SSE拒绝→Beta成功→恢复 | 拒绝零dispatch、另一租户不受影响 | UI-E2E | Pass | J05三视口/H46 |
| AC05-DATA | 暂停恢复不清历史额度 | 两scope | 已有calls/fee/reserve | 暂停恢复前后读回对账 | 历史/费/预留/pending/C守恒 | Data Reconciliation | Pass | H46 PauseUsageActualHTTP/J05最终事实 |
| AC06-UI | 一次性秘密/离开与刷新不回显 | 租户管理员 | 无目标Key | 创建→使用→关闭/切scope/刷新/退出重登 | 秘密不恢复，安全标识保留 | UI-E2E | Pass | J01/J03/J07/F47/console |
| AC06-API | 安全存储与重启不可取回 | 人员/应用 | 已签发Key | 元数据/持久恢复/重启读取 | 摘要而非明文秘密，无重放 | API-E2E / Contract | Pass | H46 IdentityAndKeys/ProcessCrashRecovery/G |
| AC07-UI | 立即/1h/24h轮换、提前撤销、不叠加 | 管理员/旧新Key | 已用应用Key | 轮换→新旧调用→409→结束旧值→撤销/期限 | 共享池和历史不清零 | UI-E2E | Pass | J03+console立即，三视口 |
| AC07-API | 精确前/等/后失效与代际共池 | 应用/共享资源 | 显式clock/并发 | 截止边界/两代争8名额2 | 401边界准确，超额拒绝不部分扣计 | API-E2E / Contract | Pass | H46 ExactDeadline/SharedResourcesActualHTTP |
| AC08-UI | 已提交丢响应/读失败安全恢复 | 管理员同context | 未签发/已用Key | 双击Enter→unknown→GET503/404/200 | mutation1，无秘密重放/叠加过渡 | UI-E2E | Pass | F47三视口/H46，历史红单独重建 |
| AC09-API | tenant/Key五独立维度、0/null、共享并发 | 两scope/多Key | 隔离矩阵夹具 | 次数/RPM/TPM/C/预算分别命中与修复，8争2 | accepted2/denied6/other_tenant1，无部分扣计 | API-E2E | Pass | H46 ResourceMatrixActualHTTP/SharedResourcesActualHTTP |
| AC09-UI | 硬预算查看/耗尽拒绝/提高恢复 | 平台/租户 | 非零已用 | 设为已用→429→提高→新调用 | 同费用历史保留 | UI-E2E | Pass | J04；不泛称十维全UI分支已执行 |
| AC10-API | 原子硬预留/软告警/未知保守 | 应用/有界无界候选 | 可信上界及失败/取消 | 并发准入/dispatch/未知/结算对账 | 未知不释放为可花余额，双scope守恒 | API-E2E / Data Reconciliation | Pass | H46 Admission/Accounting/ActualHTTP资源资产 |
| AC11-API | 修改/轮换/跨窗/回拨/重启/迟到不清事实 | 应用/运维 | 已用与pending | 持久子进程中断恢复/迟到重复/窗口变更 | 已定终态不改、一次扣计释放 | API-E2E / Contract | Pass | H46 ConsoleRealProcessCrashRecovery/ProcessCrashRecovery/G |
| AC12-UI | 当前价版本/成本租户价/单位明确 | 平台/租户 | 旧价已调用 | 发布新价→新旧详情/账单核对 | fee与cost独立，历史快照仍旧价 | UI-E2E | Pass | J06/V47 billing详情；模型列表摘要缺项另AC01 |
| AC12-DATA | 微美元精度/舍入/快照 | 费用账本 | 输入输出/边界值 | 整数边界、半微美元、累计差额、改价后dispatch | 不回算历史，不溢出误计 | Data Reconciliation / Contract | Pass | H46 MicroDollarPrecision/G计费资产 |
| AC13-API | 全dispatch已知收费/未知pending、多尝试核对 | 平台/应用 | 失败已知→未知→成功同ID | HTTP/SSE多attempt→核对重复/迟到 | 成本费分别追加，未知非0，原终态不变 | API-E2E / Data Reconciliation | Pass | H46 G账本与J06最终事实 |
| AC14-UI | 当前/pending拒结，结束平台结账/更正/导出 | 平台/财务/两租户 | 旧价请求+真实pending | 拒结→跨月→核对→结账→+5更正→明细/导出 | 原账单/快照不覆盖，理由与调整期保留 | UI-E2E | Pass | J06三视口/H46；显式clock非生产日期 |
| AC14-DATA | 幂等汇总/更正/不负数/隔离 | 平台/账本 | 可结周期 | 重复/负更正/对账/跨scope | 一次追加，金额与明细相符 | Data Reconciliation / API-E2E | Pass | H46 Billing/G/J06 |
| AC15-UI-operation | 搜索分页筛选导出、scope迟到/503恢复 | 租户管理员/平台 | 正式UI27调用 | 页2→过滤CSV→同ID→切scope迟到→503重试/重登 | 新scope正确、错误非空0、存储无秘密 | UI-E2E | Pass | J07三视口/H46 |
| AC15-UI-meaning | 同ID用量可理解/已核对事实连贯 | 平台/当前tenant | 已核对输入3输出4 | 同Request IDusage→records→详情 | 日志无口径“总量未知”与确认3+4易混淆 | UI-E2E | Fail | V47 records/usage三视口；QA47-02 P2，不判账本丢失 |
| AC16-UI | 普通/SSE/取消/同ID、双发送一次 | 管理员/应用Key | 有效受限Key | 发送/增量→完成或取消/拒绝→同ID | 无自动重放，费用可能待确认 | UI-E2E | Pass | J01/02/04/05、F47 double-normal/sse |
| AC16-DATA | busy/取消/输出后不重放 | 受控真实HTTP/SSE | 延迟/断流/失败 | dispatch计数/第一块时间/取消后账本对账 | 各双发送proxy/准入/attempt/dispatch1，Token3+4 | API-E2E / Data Reconciliation | Pass | F47/H46 SSE/routing资产 |
| AC17-API | 身份服务端/隐私/审计 | 四角色/调用Key | 敏感输入与越权 | 非法来源/伪角色/直接API/CSV/存储拒写 | 无完整对话秘密泄露，敏感写审计 | API-E2E / Contract | Pass | H46 PeopleScope/DeniedMetadataAndStorage/安全回归 |
| AC18-CLI | 正式schema1/2→3、旧身份Key历史/备份恢复 | 运维/旧应用 | 版本化旧UI夹具 | CLI升级→新UI/调用→backup/restore/失败保护 | 守恒并失败安全；非真实生产旧数据 | UI-E2E / API-E2E / Contract | Pass | H46 M三视口；测试schema1→2单独标记 |
| QA-local-fixed-entry | 原生受限监听固定入口 | QA沙箱 | 已知限制未变 | 未再次重跑必败监听 | 无提权无删门禁；宿主H46同版替代 | Contract | Blocked | QA22原限制保留；不把本行改Pass |

## H01–H08与外部事实

| ID | 要求 / 责任 | 证据类型 | 结果 | 当前证据或缺口 |
|---|---|---|---|---|
| H01 | Issue25版本包与prepare真实入口 | Contract | Pass | 会话issue/prepare回执，本轮archive/19件重新核对 |
| H02 | 唯一Issue→Run、重复入站不重建 | External State | Not Run | 当前正向关系有回执，重复事件演练未提供，平台负责 |
| H03-material | 执行节点实际同材料及具体规则承接 | Contract | Pass | 需求/architecture-r1/研发45及QA47引用PRD角色/计费/预算/交互具体规则与实际path/hash |
| H03-runtime | 实际Codex/gpt-6.1-sol可核对 | External State | Not Run | 平台指定配置非原生实际模型凭证；QA无可验证回执 |
| H04 | 交接同基线及坏摘要/缺失/过期安全恢复 | External State | Not Run | 正向material一致；负向Harness恢复回执未提供 |
| H05-fixed-entry | 当前Agent候选源码与全部固定tests | Contract | Pass | development45交接/QA独立指纹/H46同版完整入口；不据此自证全部设计符合 |
| H05-design | 实现全部冻结主要信息与动作 | UI-E2E | Fail | V47/QA47-01；设计缩减不能放行 |
| H06 | 独立检出→返原层→真实复验 | Contract | Pass | QA22/35/38→研发多轮→seq46当前原层，旧主题关闭，QA47实际缺项返研 |
| H07 | 原PR/base/head/merge/Actions/Deployment/health固定URL | External State | Not Run | 未发布，Connector/协调者负责；8793是平台而非产品效果URL |
| H08 | 唯一平台维护源/模板安装升级/原入口复验 | External State | Not Run | 未提供平台完整维护证据，不以产品代码修复替代 |
| EXT-provider | 真实供应商/真实自托管端点联调 | External State | Not Run | 没有凭据/响应；受控HTTP/SSEPass不能改写本行，不阻断内部整改 |
| EXT-identity/payment | 外部身份/真实支付 | External State | Not Run | 内部真实身份与测试USD账本是本期范围，外部不默认承诺 |

## 商用能力五层对照

沿用冻结references.md及用户已指定New API/LiteLLM官方参考；没有新读其实现或本机资料，没有扩大高级功能范围。

| 能力 | 已实现 / Static Review | 局部测试 | 宿主完整测试 | 独立QA当前结论 | 已发布 |
|---|---|---|---|---|---|
| 服务/渠道接入核验 | Go/UI/两来源/revision | Pass provider UI及self_hosted API | Pass seq46 | 接入链Pass；服务主要摘要Fail QA47-01；真实供应商Not Run | Not Run |
| 模型发布/授权 | 目录/授权交集/候选/启停 | Pass J01/J02/权限反例 | Pass seq46 | 旅程Pass；模型摘要及上下文动作Fail QA47-01 | Not Run |
| Key创建/轮换/撤销 | 一次秘密/共池/unknown恢复 | Pass current fault/contracts | Pass seq46 | 当前原用户层Pass，旧故障Closed | Not Run |
| 身份/成员/租户隔离 | 内部真实身份/成员/暂停 | Pass UI/API失权/隔离 | Pass seq46 | 身份旅程Pass；租户运营摘要Fail QA47-01 | Not Run |
| 额度/预算/计费 | 双scope五维/预留/快照/账本 | Pass并发/未知/幂等/精度 | Pass seq46 | J04/J06及账本Pass；未知与核对后显示P2未闭合 | Not Run |
| 用量日志及诊断 | 筛选分页/导出/多尝试 | Pass J07/恢复/实际HTTP | Pass seq46 | 操作链Pass；Token原始/确认解释Fail P2 QA47-02 | Not Run |
| 表格/编辑/详情/移动 | A现代后台/三视口/统一弹层 | Pass owner合同及编辑/键盘 | Pass seq46/63图采集 | 30页已独立核对；冻结主要信息Fail AC01，不是仍未查看 | Not Run |

冻结外支付税务/GPU/其他协议/自动secret分发/生产SLO/跨实例协调保留未来差距。本轮No-Go来自实际强制AC01缺项P1；P2不单独阻断，H/发布继续由相关责任方闭合，不对外声称商用全对齐。
