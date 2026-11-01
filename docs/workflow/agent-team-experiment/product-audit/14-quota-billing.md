# 额度、账期与扣减返还静态审查

## 审查边界与结论

- 本报告只静态阅读候选源码；未运行测试、未启动产品、未访问或修改数据库，因而不作运行通过结论。RPM/TPM 不在本项结论范围。
- 审查对象为 `internal/relay/accounting.go` 的预留/结算、`internal/relay/billing.go` 的账期/账单/调账，并沿调用链核对 `store.go`、`routing.go`、`routed_proxy.go`、`console_api.go`、`console_storage.go`、`legacy_accounting.go`、`identity.go`。
- 结论：P0 0 项；P1 4 项；P2 3 项。P1 均需用隔离数据库与受控 HTTP 上游复验，当前只属于有代码证据的缺陷假设。

## 已确认的实现事实与防护

1. 金额单位是“微美元/百万 token”。正常结算采用四舍五入（加 `500000`），预留采用向上取整（加 `999999`）；乘加在 `big.Int` 中完成，返回前检查 `int64`，负 token/费率被拒绝（`internal/relay/accounting.go:51-70`）。价格发布将单项费率限制在 `0..10^12`（`internal/relay/console_api.go:325-339`），可信边界将单方向 token 限制在 `1..10^9`（`internal/relay/console_api.go:239-254`）。
2. admission 使用有效时钟高水位，并以 `YYYY-MM` 固化账期、以分钟字符串固化窗口（`internal/relay/accounting.go:116-125,203,272,297`）；账单只允许关闭严格早于当前有效月份的期间（`internal/relay/billing.go:42-47`）。Schema 3 的有效时钟防回拨并持久化 UTC（`internal/relay/console_clock.go:62-98`），因此月界不会随宿主时钟回拨而重新打开。
3. 预授权不是直接写费用账本：admission 保存 `tokens_reserved`/`fee_reserved`，预算判断使用“已结费用 + 未释放预留 + 新预留”（`internal/relay/accounting.go:267-297`）。候选最多尝试三次的数据库约束见 `internal/relay/store.go:89`；预留取最大的三份 token/费用上界（`internal/relay/accounting.go:251-271`），路由不会重复尝试同一 upstream（`internal/relay/routing.go:228-230`）。这是多尝试最坏情形的保守覆盖。
4. 价格在 admission 时按 alias、upstream、上下游 execution revision 和 effective time选择，并复制到 `request_prices`，后续结算读取快照而非当前价格（`internal/relay/accounting.go:204-237,300-303,352-364`）。价格发布要求一次覆盖全部候选，且租户输入/输出费率在各 upstream 间一致（`internal/relay/console_api.go:315-347`）。边界证据明确是运营方声明、非供应商验证（`internal/relay/console_api.go:229-254`）；这属于直接合同输入，不应被表述为供应商实证。
5. 只有标为 `attempt_dispatches.possible=1` 的实际可能已发出尝试进入结算；标志在发起 HTTP 前写入（`internal/relay/routed_proxy.go:131-141`），结算汇总每个尝试的 input/output 并分别按其价格快照计费（`internal/relay/accounting.go:352-447`）。因此重试产生的供应商用量会累计，而非只计成功尝试。
6. 重复结算采用增量账本：先求同一 request/ordinal 的累计已记金额，再只插入差额；`UNIQUE(request_id,ordinal,revision)` 提供数据库防重（`internal/relay/accounting.go:429-441`；`internal/relay/console_storage.go:31`）。`Finish`、reconciliation 与账本更新处于各自同一事务内（`internal/relay/store.go:343-368`；`internal/relay/billing.go:267-271`）。账单也有 `UNIQUE(tenant_id,period)`，控制台 operation 有 `(actor,operation)` 主键（`internal/relay/console_storage.go:19,34`）。
7. 关账前强制该期间无 running/pending admission、无未解决 legacy usage（`internal/relay/billing.go:70-96`）。账单保存关账时 original fee/cost，关账后的变化只能走 correction；已关账期间禁止 usage reconciliation（`internal/relay/billing.go:255-261`）。
8. 发布价格、关账、调账、原生/历史 reconciliation 均要求 platform principal；租户管理员只能读账单，developer 被拒绝（`internal/relay/console_api.go:176-180`；`internal/relay/billing.go:15-40,121-129,153-160,204-208`；`internal/relay/legacy_accounting.go:9-13`）。reconciliation 要求 reason/evidence、乐观 revision，且请求仍运行时拒绝（`internal/relay/billing.go:214-243`）。调账记录 actor；usage reconciliation 的 reason/evidence进入不可变增量账本。

## 缺陷假设与财务影响

### P1-1：失败/取消且供应商未返回用量时不会自动返还预授权

**代码证据。** 对任何已标记 possible 的尝试，只要 input 或 output 为未知，`settleAdmission` 就继续持有对应可信上界；完整性不足还增加 `pending`，并把保留额写回 admission（`internal/relay/accounting.go:375-428,448-449`）。逻辑不检查 attempt 的 success/failed/canceled 状态。只有完全未 dispatch 的尝试被查询排除。人工 reconciliation 又只能提交确定 token，且必须由 platform 操作（`internal/relay/billing.go:204-267`）。

**影响。** 网络失败、超时或客户端取消即使供应商实际未执行，也会冻结最多三次尝试的预算，阻断后续请求并阻止关账；若没有供应商收据，运营人员无法有证据地安全释放。反面是立即归零会漏计“请求已到供应商但响应丢失”的成本，因此不能简单按状态退款。

**建议/复验。** 建立明确的“确定未 dispatch / 确定未接受 / 接受状态未知”合同；前两类自动释放，未知类保留并进入有期限的异常队列。用受控连接失败、写后断连、客户端取消、429 带/不带 usage 覆盖 `fee_reserved/pending` 与关账行为。

### P1-2：原生供应商的 total 与 input+output 矛盾时仍按“完整”释放保留

**代码证据。** `usageFrom` 只独立解析三个非负整数，不校验 `prompt+completion==total`（`internal/relay/protocol.go:311-323`）。结算中的 `complete` 在没有人工 reconciliation 时直接因 `reconciled == nil` 为真；总数一致性只在 reconciliation 分支检查（`internal/relay/accounting.go:379-390`）。所以原生响应 `{3,4,8}` 会按 input=3/output=4 计费且 pending=0。`aggregateUsage` 同样只求各字段和，不检查字段间恒等式（`internal/relay/routing.go:71-140`）。

**影响。** 直接合同证据自相矛盾却可进入确定账单，无法区分供应商错误、字段语义变化或解析问题；可能低收/少付，也可能使争议账单缺少异常标记。

**建议/复验。** 原生和人工证据使用同一完整性判定；不一致时保留合理上界并 pending，或按供应商书面合同明确权威字段。复验缺字段、矛盾 total、三次尝试中单次矛盾及后续 reconciliation。

### P1-3：人工 reconciliation 只能增加累计 token，无法纠正已确认的高估

**代码证据。** API 将新 input/output 与已有 `charge_entries` 累计值比较，任一减少即 409（`internal/relay/billing.go:245-252`）；legacy 路径也有相同单调约束（`internal/relay/legacy_accounting.go:119-123`）。而增量账本本身支持负差额，因为插入的是 `input-priorInput`、`fee-priorFee`，表也未禁止负增量（`internal/relay/accounting.go:429-439`；`internal/relay/console_storage.go:31`）。

**影响。** 若供应商先给出错误偏高 usage，运营方在关账前不能以更可靠收据下调；只能让错误金额关账后用 bill correction 抵减。成本账 `original_cost` 也没有对应 correction 字段，毛利/供应商应付仍会错误。

**建议/复验。** 在 platform + evidence + revision + 审计约束下允许累计值上调或下调，并保留追加式负差额；另外定义 cost correction。验证下调不产生负总 token/fee、重复提交防重、已关账仍只能走明确调账。

### P1-4：调账期间可任意指向未来，且账单总额的 Go 加法没有溢出保护

**代码证据。** correction period 只拒绝早于当前月或已存在账单的月份，没有未来上界（`internal/relay/billing.go:179-185`）。金额单笔限 `±10^12`，但 `originalFee+sum+input.Amount < 0` 使用普通 `int64` 加法（`internal/relay/billing.go:161-195`），读账单的 `originalFee+adjustments`（`internal/relay/billing.go:145-150`）及关账的 `fee += adjustment`（`internal/relay/billing.go:87-99`）也无显式溢出检查。数据库允许同一 bill 无限 corrections（`internal/relay/console_storage.go:35`）。

**影响。** platform 操作误把 adjustment 放到遥远未来，会立即改变旧账单 current total，却把预算/期间报表影响推迟到未来；大量合法单笔调账可在 Go 聚合加法处回绕，绕过 negative total 检查或返回错误金额。SQLite `sum` 自身可能先报 integer overflow，但不能替代应用层边界合同。

**建议/复验。** 将 correction period 限定为当前开放期间（若业务需要未来则限定一个明确窗口）；所有金额聚合采用 checked add 或 `big.Int` 后检查 `int64`。验证 `MaxInt64/MinInt64` 邻界、跨月、未来月与无限多小额 correction。

## P2 风险与改进项

### P2-1：相同/极近 effective time 的价格选择不是稳定的发布顺序

价格查询以 `julianday(effective) DESC,id DESC` 排序（`internal/relay/accounting.go:210-212`），而 ID 是随机 secret（`internal/relay/console_api.go:341-347`）。同一 effective timestamp 可重复发布，SQLite 浮点 Julian day 对极近纳秒时间也可能等价；此时随机 ID 决定生效价格。建议增加单调 publication revision/created sequence 或禁止同一 alias/upstream/execution/effective 重复，并以该序列稳定排序。财务影响是边界时刻可能快照到非预期费率。

### P2-2：账期是固定 UTC 自然月，但 API/账单没有显式时区合同字段

实现使用有效 UTC 时钟格式化 `YYYY-MM`，边界本身一致；账单响应只有 period/currency/payment（`internal/relay/billing.go:35,107,150`），未返回 `period_start/end` 或 timezone。若租户按本地月理解，会在月界产生争议。建议接口与 UI 明示 UTC 半开区间 `[start,end)`。

### P2-3：关账后的 correction 只修正租户 fee，不修正供应商 cost

`bill_corrections` 只有 amount，列表与当前账单只把它加到 fee；original cost 永不变（`internal/relay/console_storage.go:35`；`internal/relay/billing.go:131-150`）。这可视为当前产品只支持客户侧贷/借项的设计，但与“人工 reconciliation”容易混淆。建议明确命名为 tenant fee correction，并另设有证据和审批的 cost correction/供应商应付流程。

## 未执行项与复验清单

- 未运行任何 Go/HTTP/浏览器测试，未执行 `make verify`，未启动服务，未接触真实或测试供应商；所有优先级均为静态审查判断。
- 未验证 SQLite 在累计 `sum()` 溢出时的具体 API 映射、并发事务下重复 Finish/reconciliation 的实际返回、崩溃恢复中 reservation 的最终状态。
- 未验证月末最后纳秒与次月第一纳秒、同 effective 价格并发发布、三次 mixed success/failure/cancel、非流式错误体 usage、流式中断 usage、关账与 reconciliation 竞争。
- 未验证真实供应商是否对失败/取消请求计费，也未获得任何直接合同或价格表；源码中的 operator-declared bounds/evidence 不能替代该输入。

## 优先级汇总

| 优先级 | 数量 | 财务含义 |
|---|---:|---|
| P0 | 0 | 静态证据不足以确认可直接造成无权限扣款、跨租户记账或不可逆账本破坏的确定缺陷。 |
| P1 | 4 | 可能冻结客户预算并阻断关账、把矛盾用量确定入账、无法纠正高估、或使调账期间/总额回绕。 |
| P2 | 3 | 价格边界确定性、UTC 合同可见性、fee/cost 调账语义需要收紧。 |
