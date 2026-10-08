# QA38 当前分层验收矩阵

design-v0.2.0 / safe-readback-resource-matrix-r1。HEAD98511771871cf0951ecef716bbb55deb13e18da5；127文件工作树3ee2122eadac33a67342523a082f791d8a81154cc86b7321d53de856866a15ef。材料archive5f51d35fdc15ed7214749f9413879b4238097f1d6b6bbe2cc0d508d8e103cef7；实际路径/摘要见evidence/attempt-38/material-integrity.json。此矩阵不改写QA35历史矩阵，也不把旧指纹的Pass搬来放行。

证据缩写：H37=evidence/attempt-38/host-seq37-review.md与平台seq37完整归档；A38=evidence/attempt-38/artifact-index.json（63件字节一致宿主产物）；F=tests/browser/console-faults.mjs及host-seq37/fault-{1280,1440,390}.json；J=tests/browser/journeys.mjs及本轮三个continuous JSON（PID5283/5310/5330）；C=tests/browser/console.mjs；G=internal/relay/console_contract_test.go、console_resource_matrix_test.go、console_http_test.go、console_v3_test.go、console_process_test.go；L38=qa-node.log、qa-go.log、qa-static-final.log；M=tests/browser/migration-console.mjs及正式CLI迁移runner；V38=visual-review-38.md。代码定位不等于执行证据；所有Pass需对应H37/L38/A38中的实际结果。

## 产品AC01–AC18

| ID | 基线要求/本行范围 | 角色/范围 | 起始状态 | 连续步骤 | 最终事实 | 证据类型 | 结果 | 证据位置/边界 |
|---|---|---|---|---|---|---|---|---|
| AC01-UI-main | A十页入口、多对象、三视口和键盘 | 四角色/两租户 | 空安装→多对象 | 导航/创建/校验/键盘提交/编辑/详情/移动菜单 | 正式操作成功、根无横向溢出 | UI-E2E | Pass | H37+C+J三视口；入口/动作不替代完整设计对照 |
| AC01-UI-design-review | 批准原型逐页主要信息与动作对照 | 各真实角色 | 非空各页及适用空态 | 按十页逐页检查实际页面与原型 | 未减承诺信息、主要动作可读可达 | UI-E2E | Not Run | V38列逐页缺口；已查看Key/账单/故障样本，不是十页完整对照；QA38-01强制证据门槛 |
| AC02-UI-provider | 保存失败保留草稿→修正→核验→发布 | 平台 | 无目标服务/模型 | 非法URL→修正→GET核验→模型发布 | 失败无半对象，凭据清除，当前核验applicable | UI-E2E | Pass | J01/H37；provider来源；真实供应商另Not Run |
| AC02-API-sources-revision | self_hosted接入及配置变更/旧核验隔离 | 平台/受控端点 | 正式创建受控self_hosted/旧revision | 接入→核验→更新→旧结果→新核验 | 来源持久、旧结果不覆盖当前配置 | API-E2E/Contract | Pass | H37 Governance/ServiceCheckRevision/EvidenceConfigurationRace；不升级为自托管UI首次链或真实供应商事实 |
| AC03 | 新模型空grant、目录与Key授权交集、拒绝零dispatch | 平台/租户 | 无授权模型→有Key | 发布/授权→Key收窄→403→修正→新调用 | 拒绝无dispatch、旧快照不改 | UI-E2E | Pass | J01/J02+H37 G正反例；API底层独立补证 |
| AC04-UI | 正式身份/邀请接受、诊断分配收回、成员退出/降权/最后管理员 | 平台/管理员/开发者/财务 | 空租户/新身份 | 邀请→真实登录→分配→降权→移除→应用Key继续调用 | 人员身份收回，应用Key不盲撤销，最后管理员409 | UI-E2E | Pass | J01/J05/H37；角色切换用不同真实context，不用前端角色模拟 |
| AC04-API-isolation | 跨租户列表/详情/修改/筛选/导出范围 | 两租户/四角色 | 两租户真实身份 | 直接越权/旧会话访问/搜索/导出 | 404/拒绝不泄另一租户 | API-E2E | Pass | H37 PeopleScopeActualHTTP/G权限资产；L38 InProcess只是补证 |
| AC05 | 暂停全所属Key普通/SSE、另一租户独立、恢复守恒 | 两租户/多Key | 已有非零费用/用量 | 暂停→多Key普通/SSE拒绝→Beta调用→恢复 | 拒绝零dispatch，calls/fee/reserve/pending/C不清零 | UI-E2E/Data Reconciliation | Pass | J05三视口+H37 PauseUsageActualHTTP；L38 InProcess补证 |
| AC06 | 一次性Secret/安全元数据、关闭/刷新/退出/重启不能取回 | 管理员/应用 | 无目标Key | 创建→真实调用→关闭/刷新/退出重登 | 完整值不持久、不回显，浏览器存储空 | UI-E2E | Pass | J01/J03/J07/C/F+H37身份Key及真实进程测试；重启是独立API/进程证据，不声称UI重启操作 |
| AC07 | 立即/1h/24h、截止前等后、过渡不叠加/提前结束、共享历史池 | 管理员/旧新Key | 已使用应用Key | 正常轮换→旧新调用→409→结束→启停/期限/撤销→历史 | 历史保留/共享池，边界401 | UI-E2E/Data Reconciliation | Pass | C立即；J03三视口；H37/L38 ExactDeadline及SharedResources两个入口补API/Contract边界 |
| AC08-UI | 签发/已用Key轮换已提交丢响应→读失败→恢复，双提交 | 管理员同context/scope | 无目标Key/有历史Key | 保存双击/Enter→unknown→GET503→404→200；轮换再尝试409 | mutations1、同operation、安全标识与retry保留，无secret重放；旧历史/usage/价快照不变 | UI-E2E/Data Reconciliation | Pass | F三视口与H37；旧editor红为修复后隔离重建，不冒充原始浏览器红灯；QA35-01Closed |
| AC08-CONTRACT | unknown锁写、GET去重、关闭/scope变化迟到隔离 | 当前生产editor | 写丢响应/受控未完成GET | repeated submit/retry→epoch变/close→迟到 | 一写一读、旧详情不恢复 | Contract | Pass | L38 169 Node内readback.test.mjs四例；此行不是浏览器覆盖所有迟到组合 |
| AC09-API | tenant/Key五维独立0/正有限/null及双scope原子性 | 两scope/应用 | 零用量→在途预留 | 0拒绝→finite占满→竞争拒绝→null恢复 | 零dispatch、另一scope不部分扣计、修改不清零，最终fee14/reserve0 | API-E2E | Pass | H37 ResourceMatrixActualHTTP十分例；L38 InProcess补证；不拿fixture的allowUnbounded当可信供应商上界 |
| AC09-API-shared | 多Key共享tenant/旧新代共享Key、8争2与另一tenant独立 | 两租户/多Key/两代 | 已有1次历史调用 | barrier8→2占位6拒绝→另tenant成功→完成 | N守恒、租户3calls/fee21/reserve0，另一tenant1call/fee7 | API-E2E/Data Reconciliation | Pass | H37 SharedResourcesActualHTTP；L38同InProcess；实际共享竞争选择C维度，不泛称所有资源组合穷举 |
| AC10 | 硬费用耗尽拒绝/提高不清已用，未知保守预留/软预算 | 平台/应用 | 已用非零/有界或无界候选 | 页面硬耗尽→429→提高→调用；API多尝试/未知/软预算 | 拒绝零dispatch，未知保留，可信上界不足拒绝 | UI-E2E | Pass | J04三视口；H37治理/预算API另补其不变量；受控上界不等于真实供应商证明 |
| AC11 | 改限额/轮换/跨窗/时间回拨/重启/迟到不改旧事实 | 应用/运维 | 受理或pending历史 | API跨月/回拨→进程崩溃→重开→核对 | 不重放/不重复扣计或释放，历史保留 | API-E2E/Contract | Pass | H37 GovernanceActualHTTP/RealProcessCrashRecovery/HTTPRestartAndBackup；L38治理/资源/期限补证 |
| AC12 | 成本/租户价独立、受理快照、USD微美元舍入 | 平台/租户 | 旧价请求 | 新价→新调用→旧新详情 | fee14→21、cost7、旧快照不变 | UI-E2E/Data Reconciliation | Pass | J01/J06/H37；L38 MicroDollarPrecision独立Contract |
| AC13-API-ledger | 失败已知→未知→成功各dispatch收费，追加差额与幂等 | 平台/三个受控上游 | 有界多候选 | 多尝试→pending→核对→重复/低值/旧revision | 原bill19/60与原快照保留、delta fee5/cost16 | API-E2E/Data Reconciliation | Pass | H37 GovernanceActualHTTP+L38 GovernanceInProcess；不把未知revision0费用0称完整结算 |
| AC13-UI-reconcile | 单次未知调用正式页面核对 | 平台 | 真实pending请求 | 同ID详情→核对→安全读回 | entries2、fee21/cost7/Tokens3+4、pending/reserve0，旧revision0不改 | UI-E2E/Data Reconciliation | Pass | J06+H37；UI单次场景不代表UI多尝试核对所有组合 |
| AC14 | 当前/待确认拒结→结束核对结账→追加更正→导出/隔离 | 平台/财务/两租户 | 旧价/pending/开放周期 | J06完整结账更正旅程 | 原金额保留、原因/调整期可追踪、财务无写、他租户404 | UI-E2E/Data Reconciliation | Pass | J06三视口/H37；G实际HTTP幂等和旧revision拒绝另补证 |
| AC15 | 搜索/筛选/分页/同ID详情/导出/迟到隔离/读取失败恢复 | 管理员/平台范围 | 27次正式页面调用 | 页2→筛选→详情跳日志→CSV→换scope→旧读晚到→503重试 | 只显示新scope，失败非空0，CSV不串租户 | UI-E2E | Pass | J07三视口/H37；不是DOM替身或API预制分页替代 |
| AC16 | 普通/SSE/输出后不重放/取消提示/快速双发送 | 应用同UI/context | 正式有效Key | 两次用户动作→延迟上游→同ID；取消后新请求 | 普通/SSE各一次准入dispatch/尝试，Tokens3+4；取消费用可待确认 | UI-E2E/Data Reconciliation | Pass | F三视口双发送，C/J/H37保留SSE/取消与恢复；不声称上游取消免收费 |
| AC17 | 隐私/安全身份/审计、无秘密完整对话 | 各人员/应用 | 真实身份敏感输入 | 非法来源/越权→安全metadata/审计/CSV；离开清理 | 拒绝不泄secret/完整对话，敏感写审计 | API-E2E/Contract | Pass | H37安全/DeniedMetadataAndStorage/PeopleScope；UI存储/截图另C/J/F |
| AC18 | 正式CLI schema1/2→3、兼容/备份恢复/失败守恒 | 运维/存量应用 | 版本化旧UI夹具 | 旧UI数据→CLI升级→新UI/调用→backup/restore | 身份/Key/历史/未知保留，失败不迁移 | UI-E2E/API-E2E/Contract | Pass | H37+M三视口；旧schema1→2单列，不当正式部署或真实生产数据证明 |
| QA-LOCAL-GATE | 原生沙箱监听限制下的固定入口 | QA沙箱 | 限制未变 | 本轮不重复已知失败 | 不提权不删门禁 | Contract | Blocked | QA22原qa-make-verify.log保留；H37同版宿主完整入口为允许替代，原Blocked不改Pass |

## H01–H08与外部事实

| ID | 当前责任/事实 | 证据类型 | 结果 | 具体证据/未决 |
|---|---|---|---|---|
| H01 | Issue25包、prepare材料与摘要/入口 | Contract | Pass | 会话prepare/issue回执+QA38 archive/19件独立hash |
| H02 | 唯一Issue→Run关系与重复入站安全 | External State | Not Run | 当前Run关系有回执；重复事件演练未提供，平台/协调者负责 |
| H03-material | 同版本材料实际读取与规则承接 | Contract | Pass | architecture-r1及研发/QA具体规则、引用与本轮复验产物 |
| H03-runtime | 用户指定Codex/gpt-6.1-sol实际原生执行 | External State | Not Run | 没有可核对原生执行模型回执；未切模型，不拿配置自报替代 |
| H04 | 同基线交接及缺文件/错摘要/旧版本负向恢复 | External State | Not Run | 正向承接19件一致；平台负向恢复无完整真实回执 |
| H05 | Agent实现与全部批准设计兑现、同版固定tests | UI-E2E/Contract | Not Run | H37源戳及真实tests有效；AC01完整实际设计对照仍缺，不能只据代码/标签勾全部符合 |
| H06 | 独立检出偏差→原阶段整改→原层复验 | Contract | Pass | QA22/35返研→development36→seq37同版→QA38关闭既有故障/资源主题并保留新证据门槛 |
| H07 | 同PR/base/head/merge、Actions/Deployment、health与固定页面 | External State | Not Run | 未发布；本节点不提交推送/开PR/合并/部署，8793不是产品URL |
| H08 | 唯一平台维护源/模板安装升级与原入口复验 | External State | Not Run | 平台责任；没有完整证据，不以产品仓修复替代 |
| EXT-PROVIDER | 真实供应商与真实已有推理端点 | External State | Not Run | 未提供凭据；受控HTTP/SSE是本项目语义证据，不阻断内部补证 |
| EXT-IDENTITY/PAYMENT | 外部身份/支付 | External State | Not Run | 当前真实内部身份、USD测试分账；外部不在本实现默认承诺 |

## 商用能力五层对照

参考沿用冻结docs/references.md与用户范围，不读取New API实现/本机数据，不声称已全面对齐New API/LiteLLM。

| 能力 | 已实现/静态评估 | 局部测试 | 宿主完整测试 | 独立QA当前范围 | 已发布 |
|---|---|---|---|---|---|
| 服务/渠道接入核验 | 正式Go/UI、两来源标识 | Pass保存/核验/revision | Pass seq37测试集 | Pass provider旅程/self_hosted API；全部页面设计对照Not Run | Not Run |
| 模型发布/授权 | 目录/授权交集/启停 | Pass UI/API正反例 | Pass seq37测试集 | Pass J01/J02；列表主要信息设计对照Not Run | Not Run |
| Key创建/轮换/撤销 | 生命周期/未知安全恢复/共池 | Pass当前readback Contract | Pass seq37测试集 | Pass J03与F三视口，QA35-01Closed | Not Run |
| 身份/成员/租户隔离 | 真实内部身份/角色与暂停 | Pass邀请/降权/收回/隔离 | Pass seq37测试集 | Pass J01/J05/PeopleScope；成员页设计对照Not Run | Not Run |
| 额度/费用预算 | 五维/双scope/原子预留 | Pass两scope五维/共享资源 | Pass seq37测试集 | Pass J04及ActualHTTP矩阵；预算页设计对照Not Run | Not Run |
| 定价/计费/账单 | 快照/追加账本/结账更正 | Pass精度/多尝试/幂等 | Pass seq37测试集 | Pass J06及列明账本事实/局部截图对照 | Not Run |
| 用量日志/分析/诊断 | 过滤/分页/导出/同ID | Pass真实分页及恢复 | Pass seq37测试集 | Pass J07；分析/日志逐页设计对照Not Run | Not Run |
| 表格/筛选/编辑/详情/移动 | A后台/三视口入口 | Pass键盘/生命周期/unknown/双提交 | Pass seq37测试集 | Pass列明主链/故障；AC01完整原型对照Not Run | Not Run |

冻结外支付税务/GPU/其他模型协议/自动秘密分发/生产SLO/跨实例协调均后续差距，不纳入本轮返研。**当前No-Go仅因强制AC01证据门槛未闭合；不是P2样式建议阻断，也不是要求研发补平台或真实供应商权限。** H最终退出在各责任方另闭合。
