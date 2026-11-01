# 06 测试组合与断言质量审查

## 结论

本次为静态工程审查，未执行 Go、Node、浏览器、服务或供应商请求。测试组合已经形成单元/handler、实际 HTTP、独立进程、浏览器旅程和历史回归的多层防线，且 `make verify` 把 race、构建、受控延迟实验与真实 Go 服务浏览器旅程纳入同一入口。当前未发现需要阻断候选的 P0 测试缺陷。

主要改进面是：门禁没有覆盖率或测试清单漂移约束；部分回归直接观察私有函数、SQLite 表和测试时钟，重构抗性偏低；受控上游及管理配置夹具在多个层次重复，维护时可能产生语义漂移。所有“上游”证据均为本地受控 HTTP fixture、stub/mock 或自托管测试 handler；没有凭据、外部供应商端点及真实供应商响应，因此**不是供应商集成验证**。

## 范围、方法与限制

- 先用静态文件枚举和 `func Test`/`test(` 搜索量化测试入口，再按风险域抽取代表文件阅读：认证（`bearer_auth_regression_test.go`）、备份恢复（`backup_manifest_regression_test.go`）、有效时钟（`clock_advance_regression_test.go`）、路由/SSE/计量（`routing_http_test.go`、`sse_terminal_regression_test.go`、`metering_regression_test.go`）、进程恢复（`console_process_test.go`）、命令与部署（`cmd/model-relay/*_test.go`）、浏览器编排与前端 VM 测试（`tests/browser/`）、延迟 fixture（`tests/latencyfixture/`）及 `Makefile`。
- 这是按风险域和代表文件抽样，不是逐行通读全部测试。顶层入口数量不等于实际 case 数；表驱动 subtest、动态场景和嵌套 Node test 会使实际 case 更多。
- 静态计数：`internal/relay` 与 `cmd/model-relay` 共 131 个顶层 Go `Test`；`tests/browser` 有 6 个 Go `Test`，15 个 `*.test.mjs` 中至少 204 个行首 `test(` 声明；`tests/latencyfixture` 有 1 个 Go `Test`。这些是职责规模指标，不是覆盖率。
- 未运行任何命令来验证通过率、耗时、flakiness、race 结果或浏览器像素结果；下文“已验证”仅指代码/配置静态可见，“推断”明确标注。未读取凭据，也未访问供应商。

## 前置候选记录

任务控制器显示三项依赖均已合入当前基线：

| 候选 | 候选提交 SHA | 合入 SHA | 静态审查关注点 |
|---|---|---|---|
| strict-bearer-auth | `67f464494f4d59be8a20ef99d9b15175ecd00e78` | `2d48ee08f12368e1b806fcc0975943e2eca0c8ee` | 非法 Authorization 形态、合法授权范围、生成请求不触达上游 |
| single-clock-advance | `318639d819165f3309f123133fdaf6229003270c` | `532e713acc998162c682b310c8f280e1651e6f95` | 每请求一次推进、回拨高水位、存储失败关闭 |
| strict-backup-manifest | `7dc57094557b8883c130a41ed306d1223c0d92f8` | `122c9abece28f562616483e991098d020d236d6d` | 重复/未知字段、尾随值、类型与摘要、失败不部分安装 |

上述 SHA 来自任务依赖元数据；本审查未用 Git 命令复验提交图。

## 组合职责与正面能力

| 层次 | 代表职责与依据 | 断言性质 | 判断 |
|---|---|---|---|
| 单元/边界 | `internal/relay/backup_manifest_regression_test.go:66-122` 表驱动拒绝歧义 manifest，并检查失败目标不存在；`routing_boundary_test.go:43-101` 覆盖 deadline、取消及首次提交边界 | 同时观察输入拒绝、外部结果与少量私有 seam | 已验证：失败恢复断言明确；部分白盒合理服务于难注入边界 |
| handler/HTTP 集成 | `bearer_auth_regression_test.go:10-70` 经 `ServeHTTP` 检查 401/200、模型授权范围及上游命中为零；`relay_test.go:337-697` 覆盖管理/普通请求、SSE、取消、超时、重定向、网络与存储失败 | 主要观察状态码、响应体、header、调用次数和持久结果 | 已验证：用户可观察行为占主导，安全回归不是只测解析函数 |
| 实际监听 HTTP | `console_contract_test.go:39-96` 通过 `actual` 分支在 `httptest.Server` 与进程内 handler 间复用场景；`routing_http_test.go:85-248` 覆盖真实 HTTP 选择、故障转移、恢复、禁用和 committed stream 不重放 | 网络栈行为加持久化证据 | 已验证：同一合同常有 in-process/actual HTTP 配对；但重复执行成本较高 |
| 并发/race | `Makefile:33-34` 使用 `go test -race -count=1 -v ./... -parallel=4`；`metering_regression_test.go:386-529`、`routing_latency_test.go:286-411`、`console_http_test.go:192-210` 有竞争/并发场景；多处 `t.Parallel()` | 并发结果、计量上界、序列与持久化证据 | 已验证：race 被门禁声明，且不仅依赖 race detector；未执行，不能确认当前无 race |
| 进程/崩溃恢复 | `crash_test.go:16-127` 与 `console_process_test.go:19-180` 通过当前测试二进制的 worker 模式、动态监听端口、kill/reopen 验证中断状态；`console_clock_recovery_test.go:103-169` 覆盖跨进程时钟/恢复 | 观察真实进程退出、HTTP 断开、重启后状态 | 已验证：不是单纯函数模拟；启动/终止时序仍可能是慢测与抖动来源 |
| CLI/部署 | `cmd/model-relay/main_test.go:52-164` 与 `deployment_test.go:11` 覆盖命令合同、安装/备份/恢复/升级与幂等升级 | 命令输出、退出效果和落盘状态 | 已验证：运营入口有独立职责，而非只复用库测试 |
| 前端逻辑 | `tests/browser/ui.test.mjs:1-80` 在 VM 中构造 DOM/fetch/timer stub；该文件可见密码清理、会话、取消、迟到响应与控件恢复等断言；其余 `*.test.mjs` 分担视觉摘要、可读性、迁移、保存生命周期、捕获完整性 | 多数观察可见文案、可用状态、请求载荷和焦点；也绑定具体 DOM id/中文文案 | 已验证：快速而细；DOM stub 不是浏览器实现，需由旅程层兜底 |
| 浏览器旅程 | `Makefile:46-47` 运行 `go run ./tests/browser`；`tests/browser/main.go:63-305` 启动隔离 Store、受控上游、动态端口和实际 Go 服务，再调用 Node 浏览器脚本；`migration.go:65-376` 覆盖旧 schema/升级后浏览器路径 | 用户操作、网络读回、响应/视觉/进度行为 | 已验证：后端 API 不是由静态根替代；推断：需要本机 Node/浏览器依赖，环境敏感 |
| 延迟实验/fixture | `tests/latencyfixture/fixture.go:57-112` 使用 `httptest.Server`、`127.0.0.1:0`、独立 Store 并集中关闭；`fixture_test.go:5-21` 证明配置走公开管理面 | 公开管理配置和受控场景 | 已验证：隔离良好；单个 fixture 自测只覆盖 happy path，主要价值由上层实验消费 |
| 历史回归与统一门禁 | `Makefile:8-30` 顺序执行 evidence/review-history/health-history、format、module verify、vet、前端、race Go、build、latency experiment、browser；`verify` 委托 `go run ./tests/verify` | 编排级完整性 | 已验证：入口完整且 `-count=1` 避免缓存假绿；未执行，不能复述历史收据为本次结果 |

## 隔离与证据边界

- 临时目录/SQLite：抽样中的安装、备份、恢复和迁移普遍使用 `t.TempDir()`，例如 `upgrade_test.go:12-23`、`backup_manifest_regression_test.go:55-76`、`tests/browser/journeys_test.go:111-120`。fixture 自行创建并关闭 Store；静态上未见指向用户配置目录的路径。
- 端口：实际 HTTP/浏览器/延迟 fixture 使用 `httptest.NewServer` 或 `net.Listen("tcp", "127.0.0.1:0")`，例如 `tests/latencyfixture/fixture.go:57-81`、`tests/browser/main.go:64-72,213-221`、`console_process_test.go:30-38`。测试请求中出现的 `127.0.0.1:5544` 多为 handler URL/Authority 字符串，不代表监听固定端口。
- SQLite 故障注入：`boundary_test.go:130-144`、`routing_latency_test.go:540-564` 通过测试库上的 trigger/关闭连接制造写失败；`backup_manifest_regression_test.go:96-101` 检查恢复失败后没有部分目录。这些是隔离故障注入，不是篡改生产记录。
- mock/stub：`tests/browser/ui.test.mjs` 的 DOM、fetch、timer 是手写 stub；`httptest.NewRecorder` 是进程内 handler 夹具；`httptest.Server`/本地 listener 是受控 HTTP fixture。它们可证明协议与恢复逻辑，但不能证明 DNS、TLS、供应商限流、供应商流式方言或真实计费语义。
- 浏览器中的“self-hosted”/“controlled” upstream 仍由仓库 handler 生成。`tests/browser/journeys_test.go:108-203` 明示受控响应和 “not supplier validation” evidence；因此它只证明 relay 对已知 HTTP/SSE 合同的处理。

## 缺口、优先级与建议归属

### P1：缺少可审计的覆盖率与测试清单漂移门槛

**证据（已验证）**：`Makefile:8-47` 的门禁包含 race、前端、构建、延迟和浏览器，但没有 `go test -coverprofile`、覆盖率阈值、关键风险域清单或 Node coverage。静态入口数量很多，但无法回答认证、恢复、错误码或浏览器旅程是否因重命名/删除而悄然消失。

**影响**：大规模测试资产可能“全绿但保护面缩水”；审计只能用文件名和入口数近似职责，无法发现未覆盖的新分支。

**建议/归属**：仓库流程负责维护一份轻量风险域—代表测试映射并评审删除/改名；仓库基建在不显著拉长浏览器门禁的前提下产出 Go/Node 覆盖率工件，先设观察基线，再对安全与恢复关键包设置温和的差分阈值。不要用单一总覆盖率替代合同测试。

### P1：部分回归断言绑定内部表、私有函数或时钟调用次数

**证据（已验证）**：`clock_advance_regression_test.go:11-112` 直接替换 `store.clock`、读取 `console_clock` 并以函数调用次数为核心；`backup_manifest_regression_test.go:108-122` 直接调用私有 `decodeManifest`；`routing_latency_test.go:100-403` 多处直接调用 `routeModel`、`upstreamByID`、读取 evidence 表；`upgrade_test.go:66-95` 直接匹配内部列。

**影响**：这些断言能精准保护事务/迁移不变量，但表或内部 seam 的等价重构会产生较多非用户行为失败；反之，仅检查内部计数可能漏掉一次请求中其他 SQLite 写事务。时钟回归已有 HTTP 状态和高水位断言，仍缺面向“每请求写事务数/锁竞争”的稳定外部遥测。

**建议/归属**：工程工具提供测试专用、语义化的事务/时钟观察器（仅测试构建或 package-private helper），让回归断言“一个外部请求只产生一次有效时钟推进”而非具体函数实现；仓库流程要求白盒断言旁至少保留一个 HTTP/CLI 可观察结果。迁移测试可继续读 schema，因为 schema 本身就是兼容合同。

### P1：受控上游和管理配置夹具重复，存在语义漂移成本

**证据（已验证）**：`relay_test.go:36-189`、`console_contract_test.go:34-164`、`tests/browser/journeys.go:150-226`、`tests/browser/journeys_test.go:108-203`、`tests/latencyfixture/fixture.go:55-133` 都建立本地 upstream 与管理配置；多个场景重复 tenant/upstream/model/grant/bounds/prices/key 的长配置序列。in-process/actual HTTP 成对入口也广泛重复同一 scenario。

**影响**：协议字段或管理流程变化需要跨层同步；某层 fixture 的默认行为不同可能制造假阳性或不一致失败。重复是跨层独立性的代价，不应合并成一个全局万能 fixture。

**建议/归属**：仓库基建抽取“小而显式”的受控 HTTP 上游协议构造器和公开管理面 setup builder，保留每层独立生命周期与可覆盖行为；工程工具增加 fixture contract test，固定 JSON/SSE/错误/延迟能力矩阵。通用平台可缓存只读浏览器依赖，但不得共享 SQLite、端口或运行态 fixture。

### P2：浏览器 VM stub 与真实浏览器职责边界需要显式清单

**证据（已验证）**：`tests/browser/ui.test.mjs:1-80` 自行实现有限 DOM、timer、fetch；`Makefile:39-40` 把全部 `*.test.mjs` 作为 frontend-test，`Makefile:46-47` 另有实际服务浏览器阶段。文件命名并不能自动说明某个交互同时由两层覆盖。

**影响**：VM stub 对焦点、事件传播、Abort、表单和布局的模拟可能偏离浏览器；真实旅程若未覆盖同一关键交互，stub 绿色会被误解为端到端保证。

**建议/归属**：仓库流程为登录/密钥、保存读回、取消/迟到响应、SSE、迁移和移动视口维护最小旅程矩阵；前端 stub 专注状态机和请求合同，浏览器层专注平台行为与可访问交互。无需逐个 Node test 在浏览器重复。

### P2：进程与浏览器测试存在时限式同步和诊断信息损失

**证据（已验证）**：`console_process_test.go:119-151` 以 stdout ready 行和 10 秒 timeout 同步，并把 worker stderr 丢弃；后续多处 5 秒 timeout。`tests/browser/migration.go`、`main.go` 通过子进程串联多个 Node 脚本。

**影响**：慢机器上可能产生偶发超时；stderr 丢弃使真实失败难定位。race 模式及并行 4 会放大资源竞争。

**建议/归属**：工程工具保留每个子进程的有界 stderr ring buffer，仅在失败输出；用 readiness HTTP/control channel 替代仅靠日志行，统一超时诊断（阶段、PID、端口、最后状态）。通用平台提供稳定的浏览器/Go 资源配额和依赖缓存，但仓库测试仍须拥有独立端口与临时目录。

### P2：延迟 fixture 自身失败路径测试较薄

**证据（已验证）**：`tests/latencyfixture/fixture_test.go:5-21` 只有一个 happy-path 测试，验证公开管理配置、注册和发钥匙；fixture 实现包含重启、关闭、场景切换与并发状态，但本文件未见对应生命周期/错误单测。

**影响**：上层延迟实验失败时较难区分产品回归与 fixture 生命周期缺陷。

**建议/归属**：仓库基建为 fixture 增加少量自测：重复 Close、Gateway 重启、未知场景、连接中断及并发 Configure/Register；这些是 fixture 健康测试，不应宣称产品或供应商验证。

## 优先级汇总

| 优先级 | 数量 | 判断 |
|---|---:|---|
| P0 | 0 | 未发现会使当前候选测试证据整体失效的静态缺陷 |
| P1 | 3 | 覆盖保护面不可量化、白盒耦合、跨层 fixture 漂移 |
| P2 | 3 | 浏览器职责映射、子进程诊断、延迟 fixture 自测 |

## 验证状态声明

- **已验证**：上述文件、入口数量、Make 编排、静态断言与隔离方式均来自本次代码/配置阅读。
- **推断**：测试运行成本、慢机 flakiness、维护漂移和重构噪声依据结构推断，未通过执行数据证明。
- **未执行**：未运行 `make verify`、`go test`、Node test、浏览器、延迟实验、服务或外部请求；未验证当前分支实际通过率、race、覆盖率、耗时或供应商行为。

