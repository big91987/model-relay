# 多上游路由目标架构

2026-10-06；状态：Ready for Development（设计就绪，不代表实现或验收通过）。依据 [本轮 PRD](product-definition.md) R-1～R-7、[G1](g1-review.md) 与用户自动推荐授权制定；评审见 [G2](g2-review.md)。本文件是本轮技术契约；[交互契约](routing-interaction.md)、[验证交接](routing-validation.md)、[迁移决策](adr-001-storage-upgrade.md)和 [正式部署窄契约](deployment-upgrade-contract.md)分别承担界面、验证、不可逆边界及可信部署协作职责。

## 1. 整体影响、范围与责任

目标：管理员运营多个独立上游及公共模型，应用依赖稳定公共名，网关按明确规则完成一次可解释的、有界调用。延续单 Go 服务、单管理员、单安装锁、SQLite 和原生前端。路由、记录及管理必须在同一版本部署，不能只有多上游表单。

管理员拥有供应商接入授权、费用风险、启停、备份及支持责任；应用拥有终态后显式新请求的决策；供应商拥有实际生成及收费；维护者负责实现、隔离验证与故障修复。平台发布由 Connector 负责。本轮不定义商业计费、配额、多租户、外部调度器、后台生成探活或工具执行。

已确认事实：HEAD 基线为 `85f6d17c07c5fddeb2d6a39a5a067769a11bb4cc`；`store.go` 的 schema=1 限定 upstream id=1，`proxy.go` 固定取该行；上游密文使用 URL 关联数据，公共别名对应一个远端名，requests 无逐尝试字段；现有 `Backup/Restore/Open` 都对版本1作检查。已有 `.trellis/spec` 可用（入口 prepare 曾报告未初始化，那是当时状态）。上一发布事实沿用用户输入，未在本节点复验。

本轮改变：配置存储升级为 schema=2；新增候选及尝试证据；生成改为有界循环；概览聚合多个对象；管理 API 和 UI 同步更新。公开 `/v1/models`、`/v1/chat/completions`、Request ID、普通/SSE/工具参数边界继续兼容。旧管理 API 的唯一上游语义不再作为真实路由入口。

## 2. 能力、领域与唯一事实源

| 能力／拥有者 | 对象及生命周期 | 权威事实源、命令路径和退出 |
|---|---|---|
| 供应商接入／管理员 | 上游创建、改名、改地址或凭据、启停 | SQLite upstream；管理 UI→管理员 API→Store 事务；停用保留身份及历史，不硬删除 |
| 连接核验／管理员与网关 | 显式列表访问、完成或取消、旧结果历史化 | SQLite upstream_checks 保存安全终态；单进程运行表表示正在核验；POST check，不由保存或页面加载隐式触发 |
| 公共模型运营／管理员 | 公共名及候选创建、整体编辑、启停 | SQLite models/route_candidates；一次事务保存整组；公共名不可隐式重命名；停用退出 |
| 应用接入／管理员与应用 | 原有应用密钥签发、启停、调用 | app_keys 摘要及既有 API；一次性明文只留当前页面；本轮不改变治理规则 |
| 调度和暂时恢复／网关 | 选择、冷却、到期允许参与、成功或再失败 | 当前配置的运行期健康与平滑加权状态；只由真实生成结果更新；重启未知，持久历史仍在 |
| 请求诊断／网关与管理员 | 受理、零或多尝试、响应提交、终态或中断 | requests、request_attempts、request_route_decisions；精确 Request ID 查询；不以日志、浏览器或最后一条记录替代 |

核验、生成健康、生成成功是三种独立证据，不互相升级。配置是期望，核验是观察，历史是事实。浏览器没有另一套持久路由配置。

## 3. 逻辑架构、部署与关键流程

```text
管理员浏览器 ─ 会话/CSRF/来源校验 ─ 管理 Handler ─ Store 配置事务
       │                              │             │
       └─ 显式连接核验 ─ 列表 HTTP ─ 安全核验终态       │
                                                     ▼
应用 ─ 调用鉴权/协议校验 ─ requests.Start ─ 路由快照/开始门 ─ 上游 HTTP
                                       │                   │
                                       └─ 决策/逐尝试事实 ◀─┘
                                             │
                    requests/attempts 查询 ◀─┴─ 终态/未知用量
```

Go ServeHTTP 继续服务静态 UI、管理 API 和代理；SQLite/密钥文件仍在源代码之外；运行期状态在同进程内。独占安装锁阻止两个进程同时访问安装。网络请求期间不持有配置锁或 SQL 事务，单个请求生成顺序执行，不同请求可以并行。没有异步事件总线、外部控制器或共享健康服务。

管理写：验证主体及输入→事务检查当前 revision→原子更新配置与全局 config_revision→提交→在同一配置锁内失效相关运行期状态→返回确认。确认只证明配置已保存。读回失败不撤销确认、不自动重复写。全局读快照只在短事务内获取；所有 rows 及时 Close，不嵌套占用单连接。

现有admin在所有业务Handler外统一Lock/deferUnlock，研发须改为由具体管理写/读快照持有短锁，不能直接把联网check接在原全程锁里面。jsonInput、地址Policy预校验及DNS等待在配置锁外完成，再在锁内按expected_revision复查并提交；核验锁只覆盖去重登记/配置读取及结果范围校验，不覆盖HTTP。authMu只负责会话表，释放后才进入配置锁；后台收尾不反向持SQL事务等待配置锁。通过延迟核验同时读取/停用测试证明管理不会等10秒网络，不能只凭代码推断。

核验：管理员 POST→同一上游去重／登记→读取当前执行配置→10秒内 GET `{base}/models`→验证合法列表且有界解析→保存安全结果→回读。200且application/json、最多1MiB、合法JSON对象、`object:list`及`data`数组中每项有非空字符串id才成功，空data合法；不保存列表正文。10秒覆盖联网与body读取，整体取消优先；本地存储收尾失败显示未知/degraded而非核验成功。请求取消关闭出站访问，记录 canceled/unknown，不把取消改成成功；若进程消失，重启将 running 的核验历史标 unknown。改配置后的旧结果仍可存为历史，但不会成为当前配置成功。

调用：受理产生一个网关 ID及整体 deadline→现有鉴权、请求体和工具协议检查→读取完整快照→Start 落盘→过滤并选择→开始门→逐尝试登记→访问该远端模型→落盘尝试结果→允许时选择未试上游→落盘请求终态→普通输出或结束 SSE。每次切换必须在前次连接／body关闭、取消与 deadline 检查之后。

遥测／成本：合法协议中解析出的各项 usage→逐尝试事实→带完整性合计；不解析错误正文、不估计价格、不发送遥测正文。维护者通过安全类别、Request ID、healthz 和本轮验证证据处理事件。

## 4. 持久数据契约（schema=2）

保留 admin、app_keys、安装密钥和 sentinel 的身份。meta 增加 `config_revision INTEGER NOT NULL DEFAULT 1`；每次成功配置写递增，核验或请求结果写不递增。revision 是内部一致性标记，不是新产品实体。

| 表／主键 | 字段与约束 | 责任 |
|---|---|---|
| upstream / 整数 id | name UNIQUE、url、token BLOB、enabled 0/1、revision、execution_revision；取消 id=1 CHECK | id 不改；规范化非空名称 ≤128字节，名称精确区分大小写且 trim 后唯一；地址与凭据校验沿用 Policy；新建必须凭据，历史缺失独立呈现 |
| models / alias | enabled、revision、legacy_remote_model nullable | alias 沿用 validAlias；revision 单调；legacy字段只保留旧库无上游时尚不能关联的映射信息，不参与路由，修复关联后清除 |
| route_candidates / (alias, upstream_id) | remote_model、enabled、priority 0..100、weight 1..100、execution_revision；FK alias/upstream，禁止级联删除历史 | 同一公共名/上游最多一条；无额外候选ID；完整集合在一事务中校验保存；新建执行代次=1 |
| upstream_checks / 整数 id | upstream_id、execution_revision、state、started_at、finished_at nullable、duration_ms nullable、error_code、upstream_status nullable | 核验历史不含 token、URL、原始响应或模型列表正文；只存安全类别和适用执行代次 |
| requests / 原 id | 保留原列及 rowid；增加 evidence_version、config_revision nullable、model_revision nullable、committed_at nullable、successful_upstream_id nullable、attempts_complete 0/1 | 旧行 evidence_version=1，新行=2；旧 model/revision/usage 原值保留；新行 model/status 为最后真实尝试兼容摘要，只有 success 可填成功上游 |
| request_attempts / (request_id, ordinal) | upstream_id、upstream_name、remote_model、upstream_revision、upstream_execution_revision、candidate_execution_revision、started_at、finished_at、duration_ms、state、error_code、upstream_status、usage 三个 nullable 整数 | 不存 URL、密钥或正文；名称是当时快照，改名后仍可解释；ordinal 从1递增，最多3；终态只写本序号 |
| request_route_decisions / (request_id, selection_round, upstream_id) | snapshot_enabled、snapshot_priority、snapshot_weight、reason、observed_at、cooldown_until nullable | 安全选择解释，不是付费尝试；reason 为 selected/disabled_upstream/disabled_model/disabled_candidate/credential_missing/cooling/already_attempted/lower_priority/not_selected_same_priority；不存正文 |

usage 为非负 int64 或 null，溢出保持合计未知并给安全完整性说明，不 wrap。新请求三项兼容 summary 仅当该项所有实际尝试都已知且 attempts_complete 时存完整合计，否则 null；精确详情另外给 `known_sum, unknown_attempts, completeness`。零尝试的用量是“不适用”（null，非供应商返回0）；旧行使用 original usage 并标历史无逐尝试证据。

上游 revision 每次编辑递增；execution_revision 只在规范化地址或提供新凭据时递增。候选 execution_revision 只在远端名变更时递增；优先级/权重/启停不会清未到期冷却。模型 revision 每次集合或启停改变递增。名字改变不清核验；地址/凭据改变使旧核验历史化；远端名变化只使该候选生成健康失效。

requests 不重建，以保留分页 rowid 和原历史值；新增列使用显式列名，不能继续依赖 `INSERT ... VALUES` 全表列顺序。历史 FK 不依赖当前候选存在。候选在编辑中移出只删除当前关联，历史尝试快照留存；移出再加回相同上游/远端仍不能绕过未到期冷却（运行期健康键见§6）。

## 5. 权威管理接口与兼容

管理接口继续使用已有会话、CSRF、Host/Origin、no-store 与严格JSON。新增对象不能由应用密钥管理；错误不含供应商原始诊断。列表完整计数来自 SQL，列表分页使用 limit 1..100及稳定游标，读失败503不是空数组。上游详情地址仅管理员可读，诊断和概览不传播地址。

| 命令／查询 | 输入或结果与成功含义 |
|---|---|
| GET/POST `/admin/api/upstreams` | 列表 `{data,next_cursor}`；创建 `{name,url,token,enabled}`→201 安全对象，token永不返回 |
| GET/PATCH `/admin/api/upstreams/{id}` | GET完整安全配置、核验及运行范围；PATCH `{expected_revision,name,url,token,enabled}`，全部非秘密配置字段必填；空token保留同址凭据，变址需新凭据 |
| POST `/admin/api/upstreams/{id}/check` | `{expected_execution_revision}`→200安全核验结果；后台仅此显式请求拥有操作；同上游已有核验返回409 `check_in_progress`；不排队、不重试 |
| GET/POST `/admin/api/models` | v2列表 `{data,next_cursor}`及每模型候选摘要；POST `{alias,enabled,candidates:[{upstream_id,remote_model,enabled,priority,weight}]}`→201 |
| GET/PATCH `/admin/api/models/{alias}` | PATCH `{expected_revision,enabled,candidates:[...]}`整体替换；URL alias是身份、不接受改名；空集合可保存未就绪模型但不能生成，UI明示 |
| GET `/admin/api/overview` | upstream/model/key全表计数；可调用配置数；尚未完成步骤；当前运行/配置的一次真实成功证据或null；失败读取独立显示 |
| GET `/admin/api/requests`、`/{id}` | 列表保留现有分页和安全摘要；精确返回原Record字段+evidence_version、attempts、selection_decisions、usage_summary及不完整性；历史 attempts=null而非伪造一条 |

非法字段或整数400；对象不存在404；重名／重复alias409；revision冲突409 `configuration_conflict`，返回安全当前revision且让UI重读，不自动覆盖；旧地址留空新凭据400 `new_credential_required`；非法候选400 `invalid_route_candidate`，带无敏感值的字段位置如 `candidates[1].weight`。写失败503 `storage_unavailable`。同一次创建响应丢失后 UI只能GET查找/确认，不能自动POST重复；名称/alias唯一约束防重复创建，编辑revision防丢失更新。

公共协议兼容；管理 v2 是内部 UI/API 同版本契约，旧 `/admin/api/upstream` 返回410 `management_api_changed`，不保留固定id1写旁路。旧管理 models 的 `{model:...}` 输入按未知字段拒绝，不悄悄选首候选；研发更新 README 管理示例及相应测试，不移除安全断言。旧缓存UI无法操作时提供刷新入口；静态资源和API已有no-store。

`GET /v1/models` 继续仅返回公共alias；列出启用模型且至少有配置完整、手动启用上游/候选的公共名。暂时冷却不从目录抹去稳定名称，调用仍可能503；禁用公共模型保持404 `model_unavailable`。健康、核验状态不改变公开模型身份。

## 6. 快照、调度与开始门

受理快照包含全局配置代次、模型revision、全部候选远端名/优先级/权重/启用、上游身份/名称/地址/密文和执行代次。受理以完成鉴权及协议校验后读取快照为配置线性化点；整体deadline从handler受理开始计时，不从这里重置。解密失败是本地存储错误，零新访问。

同一请求永远使用快照映射，不把运行中编辑的新远端名/新凭据混入旧请求。每次开始门在 server 配置互斥锁下复查 context、全局degraded、当前密钥授权及公共模型/上游/候选的手动启用；当前关联已移出等同停用。当前停用即禁止开始；只允许快照与当前都启用者。旧快照未启用的候选不会因为后来启用而进入旧请求。

开始门与配置写提交使用同一锁：门内完成尝试可记录性检查、决策与发起登记，然后交给 transport；登记获得执行权就是新尝试的线性化起点。之后的拨号/发包属于已开始尝试，禁用确认不撤销它。锁在网络等待前释放；绝不锁住直到返回响应头。需要测试在门前和门后分别暂停：禁用确认先于门→零新访问；门先于禁用→该次可结束但无后续尝试。物理网络到达时间不能冒充开始门时间。

过滤顺序：模型启用→候选/上游启用及存在→凭据完整→当前配置对应健康未冷却→本请求未试该上游→最高priority→同级权重。过滤结果落盘，不计尝试。开始门发现变化时重新过滤剩余快照候选；无候选且零实际访问503 `upstream_unavailable`；已有访问但剩余候选耗尽502 `upstream_attempts_exhausted`，详情仍说明手动停用/冷却等排除原因。

平滑加权轮转在短锁内执行：对最高优先级可参与组每项 current += weight；选最大current，平局按整数upstream_id升序；选中项 current -= 组总weight。顺序分配依据开始门获准发起序列，不按完成时间；测试独占组1:3连续40次=10:30、1:1连续20次=10:10。配置/可参与集合改变重置该组，重启清空。状态按 `(alias, request.config_revision, model_revision)` 隔离，内部仅保留最近有序可参与组的fingerprint（含priority/weight/执行代次）和current值；每次集合变化重新置零，即使随后回到曾出现的组也不能恢复旧周期。活跃旧快照单独持有旧代次状态，旧配置请求不得修改新代次；请求释放后回收无人引用的旧代次，当前每模型只保留当前组，避免历史配置无限积累。

候选健康键为 `(alias, upstream_id, upstream_execution_revision, remote_model)`；远端名本身区分执行配置，并同时校验候选执行代次隔离迟到更新。当前同键的移出/重加保留冷却，模型改名不可用；地址/凭据变化产生新键未知。远端名显式变化时清该pair旧观察，切回旧名也为未知；单纯移出不清未到期冷却，保留到期前的短期墓碑，重加同名沿用，过期且无活跃引用后回收。只改优先级、权重、名称、启停保持原冷却。旧执行代次结果只写历史，不能更新当前健康。

同一执行配置并发结果以开始门发起序号去重，只有不早于最后已采用发起序号的观察可更新最近状态，防止旧成功晚到冲掉较新失败；冷却从被采用失败的真实完成时刻起30秒。已有并发尝试可结束，不新增半开单探针机制。到期计算状态“待重新尝试”，不启动网络。成功恢复最近成功，白名单故障重新冷却；其他失败显示最近失败但仍可参与，手动启用不由健康修改。重启健康未知，页面给运行开始时间及范围。

## 7. 有限切换与响应提交状态机

固定上限沿用 PRD：3个不同上游，整体默认300秒，响应头30秒，SSE空闲60秒，输出前每尝试60秒。deadline取客户端/配置更短者；不能在切换或冷却等待中重置。Transport 禁重定向，GetBody=nil，关闭隐式生成重放；保留禁代理、Dial地址防护和现有协议限制。

```text
accepted → recorded → selecting → dispatching → reading_uncommitted
                                      │              │
                                      │       可切换且有预算 → selecting
                                      │              │
                                      │              ├─普通完整合法结果 → finish → committed → complete
                                      │              └─首合法SSE事件 → committed → streaming → finish → DONE
                                      └─取消/总体超时/本地故障 → terminal
committed 后任何失败 → incomplete/canceled/timeout；永不回 selecting
```

只有网络建立/TLS/传输错误、尝试超时及HTTP 408/429/500/502/503/504允许在未提交时切换。其他HTTP（含其他5xx、重定向）、非法JSON、类型/工具/SSE协议错误不切换也不冷却。整体超时、取消、存储及鉴权失败不切换且不冷却。流已开始后网络或供应商超时可影响未来冷却，但本次不重放。缺DONE/非法工具协议是协议不完整，不因EOF自动等同可切换网络错误；首事件之前干净EOF也是协议错误。

首SSE事件完整解析且合法后才可能提交；首事件前只接收注释不向客户端flush，不用提前200或心跳逃避60秒上限。保持现有256KiB事件、32MiB流累计、普通8MiB及请求1MiB限额，完整验证工具结构，透传除model映射外所有现有合法生成参数，不执行工具。

提交标记在任何 WriteHeader/Write/Flush 之前置位；即便写入失败也不得重放。普通结果先合法解析及落盘success后写出，写失败按原机制纠正为canceled/downstream_write_failed并清成功证明。SSE terminal成功必须在合法DONE前落盘，失败不能发送DONE冒充成功；已提交则发送现有安全 `event:error`（连接尚可写时），HTTP200不代表完整。

输出前60秒采用可停止计时器与尝试child context，不用永不可撤销的60秒WithTimeout贯穿整个SSE。计时器、首提交和取消共享小状态锁：首提交前核对timer未触发/parent未取消并stop timer；失败则不写；stop成功后保留parent整体deadline和流idle timer。parent取消永久，不能因首包重建无父context。idle从等待合法事件开始计，合法事件重置；仅注释不无限续命。WriteDeadline不得超过整体deadline，保留单次写10秒上限。

| 终态 | 外部HTTP/机器码 | 内部记录 |
|---|---|---|
| 无有效配置/全冷却 | 503 `upstream_unavailable` | rejected，零尝试及排除原因 |
| 可切换故障后耗尽/第三次失败 | 502 `upstream_attempts_exhausted` | upstream_failed，逐尝试保留原原因 |
| 总体deadline | 504 `request_timeout` | timeout；不是尝试timeout触发新切换 |
| 单次不可切换失败 | 保留现有安全HTTP/码如 upstream_rejected/upstream_redirect/upstream_protocol_error | upstream_rejected/upstream_failed，零备用访问 |
| 客户端取消 | 不伪造已送达HTTP；内部499语义 | canceled/client_canceled，有可得用量仍按事实记 |
| 已提交后错误 | 保持原HTTP，可能流内error | stream_interrupted/timeout/canceled；不完整标记 |
| 本地记录失败 | 未提交503 `storage_unavailable`；提交后安全流错误/关闭 | degraded，不伪造已落盘终态 |

## 8. 持久化、未知与恢复

访问供应商前必须 Start 和可写尝试登记成功；失败全局degraded、禁止任何新生成。选择解释和登记在同一短事务；候选被过滤只写decision不生成attempt。尝试完成与对应usage在同一事务更新；下次尝试必须先确认前次终态落盘成功。最终请求及合计在一事务提交。关闭和取消所有本请求body/timer；数据库收尾用独立、最长2秒context，不能延长网络或响应deadline，也不能启动新生成。

SQLite提交和远端访问不能构成原子事务。登记后到实际发送前崩溃时，`dispatching`只能表达“发起已登记，实际执行结果未知”，不能证明供应商已收到、已经收费或确实零访问。正常完成回填实测发起/结束和结果；受控上游接收计数才证明验收中的真实访问。重启将未完请求及尝试标process_interrupted，attempts_complete=false，终态/用量未知，绝不自动补发。正在处理详情明确登记与执行证据的边界，不能把预选候选当作成功上游。

Start使新请求attempts_complete=false，全部实际尝试已终态且无记录缺口时最终事务才设true（成功或确定失败均可），处理中/崩溃维持false。概览仍使用保守运行期successID：只有已持久success且受理config_revision与当前相等、所用公共模型/候选/上游/key仍启用时注册；每次配置写清证明，重启为空，普通写失败纠正仅清自己的ID。读取时再次核对持久终态与关联，不用一次核验或旧成功替代当前证明。

已提交后的收尾存储失败不能撤回已输出的内容；全局degraded和healthz503，详情显示最近持久状态/不完整证据，不在内存造一份“成功”。日志只输出静态安全类别及必要非秘密关联，保留原有无正文规则。恢复由维护者处理磁盘/数据库并受支持重启，不能手工把行改success。

多尝试usage例：A网络失败三项null、B成功10/20/30→完整合计null，已知部分10/20/30且1次未知；两个成功协议中已知usage（例如首个流已有合法usage后网络失败、后续只在未提交时才有可能切换）仍按实际可得事实计算，不为了制造全已知多尝试读取错误正文。每项分别计算完整性；不把prompt+completion推算成供应商未返回的total。

## 9. 信任边界与质量场景

| 威胁／故障 | 控制、影响与责任／验证 |
|---|---|
| 应用密钥进入管理API、CSRF跨站写 | 原会话/CSRF/来源校验；管理401不泄露配置；原HTTP安全测试持续执行 |
| 地址变更夹带旧token、重定向/SSRF | 规范化Policy地址及每次Dial保护；变址必须新token；不跟随重定向；正式测试夹具只在显式隔离策略使用 |
| 多上游密文串用或记录泄漏 | schema2用 `upstream:v2:{id}:{normalized_url}`作AEAD关联；token无GET/日志/记录/截图；备份含敏感材料不进仓库 |
| 重复点击／迟到结果越过会话 | 既有operation/session/navigation身份和同步busy锁；只清自己秘密；abort非身份保证；恢复不自动生成 |
| 供应商故障风暴或永久冷却 | 顺序最多3次、30秒冷却、无后台探活及冷却等待；管理员看原因并修正或停用；故障计数验证 |
| SQLite单连接/慢外部请求拖管理 | 网络不持锁/事务；写busy及收尾有界；沿用HTTP体大小与时限；并发测试验证无死锁和无误分配 |
| 容量/磁盘耗尽 | 现有体大小限制、fail-closed、healthz503及备份；维护者负责实际容量，不宣称吞吐SLA或引入配额功能 |

度量只取当前承诺：精确次数与分配、超过配置deadline零新尝试、取消/输出后零重放、已确认停用边界、未知用量不变零、升级身份/行值守恒、两屏键盘旅程。小团队尚无服务窗口/吞吐/数据增长样本，不新设可用性百分比、错误预算图表或7×24值班组织。故障发现入口是真实调用/记录/healthz；管理员首接支持，维护者排产品/存储故障，供应商负责执行和账单；不把UI成功率当供应商SLA。

## 10. 方案比较与决策

目标模块是配置、路由及诊断；负责事实源、有限生成与解释，不负责供应商执行、账单和跨实例一致性。必需契约是PRD规则、公开协议及旧数据资产；当前阶段为团队单实例路线2。

| 方向 | 覆盖与代价、退出／状态 |
|---|---|
| 仅扩展表单、全程单上游 | 不满足真实多候选/切换；Rejected |
| 单进程SQLite事务配置＋运行期调度/健康＋持久尝试 | 覆盖当前边界；管理与网络锁分离，重启健康保守未知；schema升级需停机备份；本轮采用，按授权进入研发 |
| 多服务、持久外部熔断/任务调度 | 跨实例共享可用，但需要新基础设施、分布式一致性和运营负担，无本轮规模证据；Deferred，不采购/引入新组件 |

技术组件沿用项目Go标准库、锁定SQLite驱动、现有AEAD/鉴权及原生UI；不评估新的厂商或库。高影响迁移/回退另见ADR；其他内部函数拆分由研发在现有代码风格内落实，不预建插件API。健康不持久的代价是重启后可能再试刚失败供应商，UI必须明确运行范围；记录持久保证故障仍可追溯。

## 11. 实施接缝、风险及官方依据

实施顺序见 [验证交接](routing-validation.md)：先迁移与受支持命令→配置管理→开始门和有界生成→证据查询→真实UI连续旅程。成功条件是全部13组AC及原回归真实通过；技术建议不是独立QA签字。未来路线3以既有key_id和逐尝试事实接入治理，路线4复用升级命令及版本边界；没有预做治理/部署平台。若实际并发容量导致单连接瓶颈，先测量管理时延、busy和收尾失败，再重评存储，不提前增加共享系统。

剩余风险：真实供应商凭据及费用预算未提供，供应商Not Run；数据库/网络跨边界无法exactly-once，取消不保证撤费；旧库无上游但有映射属于合法边缘，升级必须保留并提示待关联；旧管理客户端需升级；schema回退可能舍弃升级后新增资产，必须由备份边界明确处理。新增静态核对发现可信平台activate尚无candidate upgrade，现状无法直接发布schema2旧数据升级；目标维护者须修平台维护源及标准安装路径，产品提供Go deployment-contract/upgrade和实际health storage_schema，不在产品仓复制控制器。正式Actions升级/失败恢复Not Run且作为发布依赖，不能以独立CLI通过或人工预迁移替代。无方向性需求缺口，无法通过某项验证时保留失败并修复，不能静默改变PRD。

2026-10-06查阅官方机制资料（仅原理/协议文档，非实现复制；Go文档已选仓库锁定Go1.25.0，仍须本地编译/race验证实现）：

- Go net/http，Request.Context、RoundTripper与ResponseWriter：[官方文档](https://pkg.go.dev/net/http@go1.25.0#Request.Context)。用于取消与提交边界；不据此宣称远端撤销或供应商执行一次。
- Go context：[官方文档](https://pkg.go.dev/context@go1.25.0)。用于父请求deadline、取消和有界收尾；60秒首提交计时器是本产品设计。
- SQLite事务：[官方文档](https://www.sqlite.org/lang_transaction.html)。用于原子迁移/配置事务和读快照；网络访问不在事务里。
- SQLite备份：[官方文档](https://www.sqlite.org/backup.html)。用于确认备份必须一致；本项目采用已有停止服务、安装锁与WAL checkpoint路径，不新增在线备份系统。
- AWS Reliability Pillar：[官方文档](https://docs.aws.amazon.com/wellarchitected/latest/reliability-pillar/welcome.html)；Google SRE Handling Overload：[官方章节](https://sre.google/sre-book/handling-overload/)。用于有界失败、恢复验证及不过度重试，不引入组织级SLO体系。
- NIST SP800-207：[官方出版页](https://csrc.nist.gov/pubs/sp/800/207/final)。用于管理/应用主体无隐式互信；不是合规认证。
- W3C Reflow：[官方说明](https://www.w3.org/WAI/WCAG22/Understanding/reflow.html)。用于窄屏重排；交互见独立文档，不声称全WCAG认证。
- 产品参考已在 [reference-observations.md](reference-observations.md) 中复核；只承接行为，不重复第三方功能调查、不读本机New API代码/数据/资产。
