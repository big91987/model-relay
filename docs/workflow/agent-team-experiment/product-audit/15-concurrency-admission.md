# 并发准入、排队与取消静态审查

## 结论

本报告只审查候选源码，不运行服务、并发测试或产品门禁，也不对产品作通过结论。范围为 `internal/relay/accounting.go` 的并发资源准入、`server.go`/`proxy.go` 的互斥与请求 context、`routing.go`/`routed_proxy.go` 的 attempt deadline 和直接调用合同；不复述 RPM/TPM 或重试策略。

发现两个 P1 缺陷假设和一个 P2 设计风险，没有确认到 P0：

1. **P1：全局互斥等待不受客户端取消或总截止时间约束，且没有 FIFO 公平合同。** 所有非健康检查先在 `ServeHTTP` 无条件等待 `server.mu`，但总超时 context 到进入 `generate` 后才创建。锁竞争可使已断开的请求继续占用 goroutine、推进时钟并进入后续处理，也能让实际端到端时延超过配置的 `TotalTimeout`；Go `sync.RWMutex` 没有这里可依赖的逐请求 FIFO 合同。
2. **P1：已占用槽位没有词法级 release/defer；panic 或终态持久化失败可把 `running=1` 留到进程重启恢复。** 正常早退会汇入 `finish`，但准入成功后没有围绕整个请求生命周期的 deferred settlement/recover。`Store.Finish` 失败时事务回滚，当前进程仅进入 degraded，未见后台重试或本地补偿。影响是租户和 Key 的后续请求可能持续收到并发超限，直至成功的后续终结调用或重启恢复。
3. **P2：所谓“排队”只有全局实现锁竞争，不是并发额度等待队列。** 槽位已满会立即返回 429；不同租户/Key 仍共享入口、路由和持久化互斥。若产品合同期待按租户/Key 有界排队、公平等待或取消等待，当前实现并未提供；若合同就是 fail-fast，则应明确记录，避免把 mutex 等待误称为准入队列。

## 已确认事实

### 槽位的获取、维度和原子性

- 并发使用量来自 `admissions.running=1` 的计数；tenant 查询按 `tenant_id`，Key 查询按 `key_id`，见 `accounting.go:82-89`。因此同一请求同时占用其租户池和 Key 池，而不是只按 API Key 或只按租户计数。
- `admit` 在同一 SQL 事务中读取租户/Key 限额（`accounting.go:171-186`）、依次检查两个 owner 的并发用量（`accounting.go:272-296`），再插入 `running=1` 的 admission 并提交（`accounting.go:297-308`）。调用方还以全局写锁包住整个准入（`proxy.go:154-157`）。在当前单进程调用合同下，检查和占用被串行化，未发现 check-then-insert 的本进程超发窗口。
- 任一维度超限都立即返回 `tenant_concurrency_exceeded` 或 `key_concurrency_exceeded` 的 429（`accounting.go:290-291`）；代码没有槽位 waiter、条件变量或按 owner 排队结构。因此“并发准入排队公平性”不存在可验证的产品算法。
- 存储读取、插入或 commit 失败均 fail closed 为 503（例如 `accounting.go:277-279,297-306`），不会把未提交事务算作已占槽；`defer transaction.Rollback()` 覆盖准入的所有早退（`accounting.go:131-135`）。

### 槽位释放与正常早退

- 正常释放发生在 `Store.Finish` 的同一终态事务中：更新请求后调用 `settleAdmission`（`store.go:343-368`），后者最终将 `running=0`（`accounting.go:448-449`）。这使终态与释放要么一起提交，要么一起回滚。
- 请求在准入后进入 `routedGenerate`；其取消、总超时、无候选、attempt 创建失败和各类响应路径都通过局部 `respond`/`finish` 合同收口（`routed_proxy.go:13-30,31-68,283-307`）。已取消且尚未 dispatch 的请求会在循环顶部终结，不会调用上游（`routed_proxy.go:31-36`）。
- `finish` 再次以原始 `request.Context()` 判定客户端取消，并持久化最终记录（`proxy.go:38-53`）。这保护了大部分显式 return 路径；重复/迟到 Finish 的 Schema 3 更新受 `state='started'` 限制（`store.go:355-364`），settlement 将 `running` 置零，具有值层面的幂等效果。
- 崩溃/重启保护存在：初始化扫描全部 `running=1` admission 并 settlement（`console_storage.go:103-129`）。这能在成功重启时释放本地并发，但不是当前进程内的及时恢复。

### 请求互斥、取消和截止时间

- 所有非 `/healthz` 请求在路由前争用同一个 `server.mu` 写锁，并在锁内推进有效时钟（`server.go:132-146`）。`TryLock` 失败后使用阻塞式 `Lock`，没有 `select` 或 context-aware acquisition（`server.go:136-141`）。这同时影响不同租户、Key、管理请求和生成请求。
- 总超时仅在 `generate` 开头由 `context.WithTimeout(request.Context(), TotalTimeout)` 创建（`proxy.go:19-21`），所以不包含上述 `ServeHTTP` 入口锁等待。后续 `appID`/snapshot、admit 和 `beginAttempt` 也用阻塞式 mutex；已建立的总 context 虽会到期，却不能中断锁等待（`proxy.go:129-137,154-157`; `routing.go:171-179`）。
- `beginAttempt` 获锁后先执行独立存储时钟推进，再检查 `ctx.Err()`（`routing.go:171-179`）。因此在锁等待期间取消的调用仍可能先做一次存储事务，之后才返回 timeout 合同；客户端取消也会先被折叠成内部 `request_timeout`，随后由外层 `respond`/`finish` 改写成 `client_canceled`（`routed_proxy.go:15-24,36-53`）。

### attempt deadline 的直接合同

- 每次选定 candidate 并完成密钥解封后创建 `attemptOutput`（`routed_proxy.go:68-77`）。它用父级总 context 派生可取消 child，并以 timer 实现 `AttemptTimeout`（`routing.go:388-399`）；上游 HTTP 请求直接使用该 child（`routed_proxy.go:116-126`），因此客户端取消/总超时和 attempt timer 都能取消 transport。
- 非流式响应在完整读取并校验后才 `commitNormal`；timer 在此处停止，超时则拒绝 commit（`routed_proxy.go:233-270`; `routing.go:453-464`）。因此 attempt deadline 覆盖连接、响应头和完整非流式响应体读取。
- 流式响应在首个合法 SSE 写入时停止 attempt timer并标记 committed（`routing.go:403-423`）；后续存活由父级总 context 和 stream idle timeout负责，而不是 attempt timeout。这个“首字节/首次提交 deadline”是代码的直接合同，不是整个 stream 的最长时长。
- attempt 结束的存储写使用独立 `context.Background()` + 2 秒清理窗口（`routing.go:322-330`），因此客户端取消不会立即打断审计落盘；这是有界清理保护，但也意味着 handler 可在取消后最多继续等待该存储操作。

## 缺陷假设与影响分级

### P1 — 不可取消、无端到端 deadline 的全局锁队列

**证据链：** `server.go:132-143` 在创建生成总 context 之前阻塞；`proxy.go:19-21` 才开始 `TotalTimeout`；后续 `proxy.go:129-137,154-157` 与 `routing.go:171-179` 也采用普通 `Lock`。

**影响：** 长事务或慢存储持锁时，其他租户/Key 的请求均可堆积；客户端断开不能撤销 waiter。入口等待不计入 `TotalTimeout`，所以用户观察到的总时延没有该配置所暗示的上界。锁调度没有项目级 FIFO/租户公平保证，热门或管理流量可干扰无关租户。

**为何仍列“缺陷假设”：** 静态代码已确认不可取消等待和 timeout 起点；实际饥饿程度、goroutine 增长与延迟幅度需要并发实验，本任务明确禁止执行。

### P1 — panic/终态存储失败后的槽位滞留

**证据链：** 占用在 `accounting.go:297-308` 提交；释放仅由 `finish -> Store.Finish -> settleAdmission` 完成（`proxy.go:38-53`; `store.go:343-368`），而 `generate` 在准入成功后直接调用 `routedGenerate`，没有 deferred release/recover（`proxy.go:154-170`）。`Finish` 任一步失败都会 rollback（`store.go:346-368`），调用方只设置 degraded（`proxy.go:50-52`）。重启才扫描恢复（`console_storage.go:103-129`）。

**影响：** 若准入后发生 panic，或 finish/settlement 因存储故障失败，该 admission 保持 `running=1`。在小额度（尤其 concurrency=1）下，单个滞留项即可阻断同租户和同 Key 的新请求；degraded 状态又会使当前进程全局拒绝，恢复存储本身不会清除 degraded 或自动补偿。

**为何仍列“缺陷假设”：** 未在审查路径中确认一个必然可触发的 panic 点，也未注入 Finish 故障；这是由缺少词法释放保障推导出的故障面，不声称已动态复现。

### P2 — 准入“排队”合同缺失/命名风险

**证据链：** `accounting.go:290-291` 明确立即 429；`Server` 没有按 tenant/Key 的 waiter 状态，只有全局 `sync.RWMutex`（`server.go:34-52`）。

**影响：** 如果调用方期待槽位释放后自动按先来顺序获准，会发生不受控重试和惊群；反之若预期 fail-fast，当前行为可接受，但必须把入口 mutex 等待与资源准入区分开。静态范围内未发现 FIFO、最大等待人数、队列超时、取消出队或租户隔离合同。

## 已有防护

- 准入使用 SQL 事务并在所有显式早退 rollback；存储错误 fail closed。
- tenant 和 Key 两个维度在同一事务中共同判定、共同占用，不会只占一个维度。
- 正常取消/超时路径会终结记录并 settlement；总 context 传入上游 transport。
- attempt timer 对非流式完整响应、流式首次提交建立明确边界；stream 另有 idle timeout。
- 终态更新与槽位释放同事务，Schema 3 防止迟到结果覆盖终态。
- 重启恢复会扫描并释放遗留的 `running=1`，同时保留未知用量的后续核对语义。

## 未执行项与限制

- 按任务约束，**未启动服务、未运行 Go 测试、未运行并发/取消/故障注入测试，也未运行 `make verify`**。
- 未测 mutex 饥饿、跨租户延迟、排队 goroutine 上限、客户端断开时的锁等待、锁等待超过 `TotalTimeout`、panic 注入、`Store.Finish`/commit 故障、重启恢复耗时或多进程共享数据库行为。
- 未验证真实供应商、真实网络或操作系统 socket deadline；所有结论仅来自当前候选源码静态分析。
- 本报告不评价 RPM/TPM、预算计算或重试选择语义；仅在它们与并发槽位事务边界共用代码时引用行号。
