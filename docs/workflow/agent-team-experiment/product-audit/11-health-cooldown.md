# 11. 冷却与健康状态收敛静态审查

## 结论与范围

本项只做源码静态审查。未启动服务，未发送手工或自动探测请求，未运行测试、构建或产品验证，未修改产品代码；因此本文不是产品通过结论。重试次数、请求重放和“三次尝试”上限留给第 12 项，本项只在解释健康状态写入时引用调用点。

结论：冷却的路由判定、并发迟到结果抑制、成功清除、配置执行代次隔离和重启清空形成了基本闭环；手工 upstream check 与生成路由健康明确分离。发现两个 P2 缺口；未发现 P0/P1。

## 事实链

### 1. 状态、键与持久化边界

- **事实**：`routeHealth` 保存完成序号、失效水位、冷却截止、最近状态和观察时间；`Server.health`、`dispatchSequence` 都是进程内字段，新建 `Server` 时 health 是空 map（`internal/relay/routing.go:21-28`；`internal/relay/server.go:45-52,55-75`）。所以重启不会恢复旧冷却，也不会从数据库重建“健康”。
- **事实**：健康键是 `(alias, upstream_id, upstream_execution_revision, remote_model)`（`internal/relay/routing.go:167-169`）。候选 execution revision 不在字符串键中，但 `observe` 接受迟到结果前会再次读取当前候选，并要求 upstream id、remote model、candidate execution revision 同时匹配，还要求当前 upstream execution revision 等于 attempt 记录值（`internal/relay/routing.go:336-351`）。
- **事实**：`upstream_checks` 是持久健康证据表，记录 upstream id、execution revision、状态、起止时间、时长、安全错误码和 HTTP 状态；不含路由冷却字段（`internal/relay/store.go:88-92`）。读取最新一条 check 时仅用 execution revision 计算 `applicable`（`internal/relay/configuration.go:41-62`）。
- **事实**：`request_route_decisions` 持久化每轮选择原因和当时的 `cooldown_until`，它是请求选择解释，不是重启恢复源（`internal/relay/store.go:89-92`；`internal/relay/routing.go:293-305`）。

### 2. 临时/永久分类与冷却截止

- **事实**：HTTP 仅把 408、429、500、502、503、504 归为临时故障（`internal/relay/routing.go:371-373`）。301/其他 3xx、一般 4xx（含 401/403）、501、505 及协议格式错误不进入冷却；传输错误除主动取消和目的地策略拒绝外归为临时故障（`internal/relay/routed_proxy.go:145-178`；`internal/relay/routing.go:467-469`）。客户端取消和总请求 deadline 在最终写健康前被排除或改成非临时（`internal/relay/routed_proxy.go:277-300`）。
- **事实**：临时故障观察以 `routingNow()` 为起点，将截止设为 `Observed + Config.Cooldown`；默认冷却 30 秒（`internal/relay/routing.go:358-368`；`internal/relay/server.go:68-73`）。选择时只在 `Cooldown.After(now)` 时排除，等于截止时间即自然恢复参与（`internal/relay/routing.go:195-235`）。
- **事实**：永久/非临时故障仍写入最近状态，但 cooldown 为零，因此下一请求仍可选择该候选。管理读显示为“最近生成失败（仍参与）”（`internal/relay/configuration.go:321-327`）。这是明确的 fail-open 参与策略，而非遗漏清除。
- **直接合同（源码存在，未执行）**：临时 HTTP allowlist 的实际 HTTP 表覆盖 301、400、401、403、404、408、429、500、501、502、503、504、505（`internal/relay/routing_http_test.go:148-174`）；被测时钟合同要求截止前不访问、恰好到截止时恢复（`internal/relay/routing_test.go:172-200`）。

### 3. 成功清除与并发收敛

- **事实**：每次 dispatch 在 `server.mu` 内取得单调进程序号（`internal/relay/routing.go:307-312`）。临时失败把 `InvalidatedThrough` 提升到当时全局已分派序号，使已在途的较早成功不能清除这次冷却；更旧 sequence 或处于失效水位内的观察会被丢弃（`internal/relay/routing.go:353-367`）。
- **事实**：较新的成功通过同一个 `observe(..., false)` 路径覆盖 health，保留既有失效水位但 cooldown 归零，因而清除已到它之前的冷却；普通成功调用点在请求结果和终态持久化成功之后（`internal/relay/routed_proxy.go:243-254`）。已提交流式响应也只在 attempt 已保存时观察（`internal/relay/routed_proxy.go:200-220`）。
- **事实**：health 的读取、dispatch sequence 和 observe 都在 `server.mu` 下；配置 PATCH 也持有该锁，因此没有 map 数据竞争窗口，配置提交与迟到 observe 串行（`internal/relay/routing.go:171-174,333-336`；`internal/relay/configuration.go:193-215,466-467`）。数据库写失败的 attempt 不会更新 health（普通/流式调用点同上）。
- **直接合同（源码存在，未执行）**：较旧成功不得覆盖较新失败、启停不得清冷却、旧上游/候选执行代次不得污染新状态、重启 health 为空（`internal/relay/routing_contract_regression_test.go:45-101`）；失败发生前已分派的迟到成功还要被 latency/health 水位拒绝（`internal/relay/routing_latency_recovery_test.go:140-165`）。

### 4. 配置 revision 隔离

- **事实**：请求快照带全局 config revision、model revision 和进程 config epoch（`internal/relay/routing.go:29-35,143-164`）。每次配置写在事务内递增全局 config revision，提交后递增 epoch 并清路由轮转/延迟运行态（`internal/relay/configuration.go:77-80,277-281,554-570`；`internal/relay/console.go:21-26`）。开始每次 attempt 时复核 epoch、model revision 和 upstream execution revision，变化则以 `configuration_changed` 排除（`internal/relay/routing.go:191-223`）。
- **事实**：地址或凭据变化递增 upstream execution revision（`internal/relay/configuration.go:230-255`）；remote model 变化递增 candidate execution revision，并在提交后删除该 alias/upstream 的旧 health（`internal/relay/configuration.go:532-569`）。单纯名称、启停、优先级或权重变化不产生新的健康执行代次，因此保留原冷却。
- **事实**：即使快照的全局 config revision 已过期，迟到 `observe` 没有直接比较 config revision；它按“执行身份仍相同”收敛。也就是说无关配置变更不会阻止真实结果更新同一候选健康，而地址/凭据/remote model 变化会被执行代次复核阻止。这与 health key 的执行身份语义一致。

### 5. 重启与手工探测

- **事实**：启动恢复只把未完成 request/attempt 标成 `process_interrupted`，把未完成 upstream check 标成 `unknown`；没有恢复 routeHealth（`internal/relay/store.go:235-245`）。所以重启后所有候选路由健康为“本次启动未知”，旧冷却立即消失；历史 check 仍保留并按 execution revision 展示。
- **事实**：手工 check 在锁内读取并登记同 upstream 去重标志，联网期间释放配置锁；完成后持久化结果，再读当前 execution revision 计算适用性（`internal/relay/configuration.go:583-630,631-695`）。配置可在联网期间变化，旧结果会保留但 `applicable=false`。
- **事实**：check 的成功、失败、取消均不调用 `observe`，不读写 `server.health`，也不改变 enabled/execution revision。因此手工探测不能解除或建立路由冷却。直接合同明确断言取消探测保持 execution 与 health 不变，随后允许再次核验（`internal/relay/connection_lifecycle_test.go:16-68`，未执行）。

## 发现与分级

### P2 — 管理面冷却状态使用不同的“现在”，可与实际路由判定不一致

**事实**：实际选择和观察统一调用 `server.routingNow()`（`internal/relay/routing.go:195,358`），该时钟允许由 `latencyClock` 替换（`internal/relay/routing_latency.go:85-90`）；管理读 `describeModel` 却直接使用 `time.Now()` 判断是否仍冷却（`internal/relay/configuration.go:305-327`）。

**影响**：在使用内部替代路由时钟的测试/诊断环境（现有直接合同确实这样做）中，管理 API 可能显示“暂时冷却”而路由已经恢复，或显示“待重新尝试”而路由仍排除。它不改变实际路由安全；仓库未显示该替代时钟是生产配置面，因此按 P2 记录测试与诊断证据不一致，而不推断生产用户必然受影响。

**已有防护**：实际路由的截止比较内部一致；持久化 route decision 记录实际选择时的 observed/cooldown 值。

**建议**：`describeModel` 在持有 `server.mu` 的现有调用边界内使用 `server.routingNow()`，并增加同一替代时钟下管理展示与 beginAttempt 判定一致的直接合同。若产品明确只允许墙钟，则应移除/封装可替换路由时钟并在合同中说明，而不是保留两套“现在”。

### P2 — 过期/旧执行代次 health 无回收路径，配置 churn 可令内存单调增长

**事实**：health 只有 remote model 改名时按 alias/upstream 前缀删除；`configurationChanged` 不清 health，冷却到期只是不再排除候选，upstream 地址/凭据换代也不删除旧 key（`internal/relay/configuration.go:558-570`；`internal/relay/console.go:21-26`；`internal/relay/routing.go:212-235`）。未找到过期墓碑或无活跃引用回收逻辑。

**影响**：长期运行且频繁轮换地址/凭据、增删候选或模型的实例会保留不可再命中的 health 条目。单条很小，正常规模下影响有限，但增长无上界；这也偏离架构文档所述“过期且无活跃引用后回收”的直接设计合同（`docs/workflow/runs/82227c5986e5f41875895d8016c686a6/routing-architecture.md:106`）。

**已有防护**：进程重启整体清空；remote model 变更会清该 alias/upstream 的全部 key；每个 key 只保留一条最近观察。

**建议**：在锁内配置变更或选择扫描时，按当前执行身份、冷却截止和活跃快照引用做有界清理；至少在 upstream execution revision 递增后清该 upstream 的旧代次，并加入 churn 后 map 大小有界的直接合同。

## 风险矩阵与未发现项

| 等级 | 项目 | 防护 | 用户/运营影响 |
|---|---|---|---|
| P0 | 未发现 | 执行代次复核、序号水位、同锁串行 | 本次静态范围未见可直接导致新配置被旧结果污染或永久无路由的缺陷 |
| P1 | 未发现 | 临时故障 allowlist、成功/失败序号水位和执行代次隔离 | 本次静态范围未见高概率影响正常生产请求的健康收敛缺陷 |
| P2 | 管理展示与路由使用不同“现在” | 实际路由判定自洽；选择证据持久化 | 内部替代时钟下管理诊断与实际参与性可能相反 |
| P2 | 旧 health 无有界回收 | 重启、remote model 改名可清部分状态 | 长期配置 churn 下进程内存缓慢、无界增长 |

## 假设与未执行

- **假设**：`latencyClock` 是测试/可控运行时钟而非正式外部配置；即便如此，同一进程内管理描述与路由事实不一致仍是可复现的静态契约风险。本文未通过执行确认。
- **假设**：非临时故障继续参与是产品有意的“永久故障不自动熔断”策略，依据是明确的 allowlist 和中文状态“仍参与”。若产品期望 401/403、协议错误等在人工修复前隔离，则需另立产品边界；本项不擅自改分类。
- **未执行**：所有引用测试仅作为直接合同源码证据，均未在本任务运行；未验证竞态检测、真实 HTTP、重启进程、实际时间推进、数据库故障或真实供应商行为。
- **未覆盖**：重试次数、候选重放、已提交流式响应是否重放由第 12 项审查；本文不对此作通过或失败判断。
