# T001 安全路径与资源计量可复验

- 状态：Done
- 里程碑：M01；负责人 development；评审人 tests/独立QA
- 硬依赖：README所列材料已核验，现有Go实现与固定make verify。
- 权威范围：原AC09–13/16、PLAN-01–12；不接管#30/#32。
- 交付物：Go受控回归、数值采证、原门禁回执、测试计划和复核结论。
- 开始：2026-10-09；模块 internal/relay；优先级 High；执行类型 Validation + Review。
- 软依赖：供应商前置只影响真实验收，不影响内部回归；冲突范围为测试fixture，未改生产接口/存储/前端。

## 变更边界

当前缺口首先是测试资产/证据，不是已确认产品故障。行为所有者为 accounting.go 的 admit/usageFor/boundedPayload/settleAdmission，protocol/proxy负责实际请求及终态。补测前不改生产实现。

实际修改 `internal/relay/console_contract_test.go`：允许同一个隔离fixture选择明确的受控HTTP/SSE响应，不改默认或现有断言；新增 `internal/relay/metering_regression_test.go`：实际上游请求cap、非零RPM/TPM、固定窗、已知/未知及共享池的数值断言；更新相关Trellis规范，保留所有旧门禁。测试自动进入make verify既有Go race阶段，不增旁路命令。隔离fixture调度使用原parallel=4，浏览器补只读冷却诊断及测试，不改产品前端。

非目标：调用表单新增猜测控件、供应商替换或重命名、扩大预算、数据库手改、平台功能、报告Dashboard。必要时先同层复现红，再最小修产品；没有复现不虚构红绿。

## 验证契约

受控单候选输入上界100/输出100（安全专测64）；费率1,000,000微美元/百万tokens；已知usage3+4=7；B=200、fee预留200。RPM按服务端固定UTC分钟，TPM=本分钟已知U+跨窗未结算P。等值U+P+B接受，少1拒绝且租户/Key均不部分扣计。并发使用开始屏障与阻塞上游，拒绝先读回再释放，不用串行冒充竞争。回归使用临时隔离安装，不用SQL更改业务数据。

安全链必须从正式管理bounds配置到实际受控上游payload，同ID诊断和池数值关联。有限TPM/硬预算与无限路径分开；设置cap本身不是输入/推理收费上界的供应商证明。

基线make verify已经实际执行，外层make退出2、内部gate退出2，失败为 historical listener unavailable；不是业务红。原始回执在 `test-results/verification/20261009T114858Z-733498716`。当前环境不再反复执行必然监听失败的全门禁，继续可执行的格式/vet/编译/handler回归；完整HTTP/进程/浏览器必须由tests宿主运行。新源码摘要与该环境失败分列。

## 完成判定

任务需实际产物、逐行验证、独立评审才Done；现由tests9及QA10证据满足内部交付条件，2026-10-09 report执行11原位更新。原#33数值网页/真实供应商仍Not Run/Blocked，不属于本次Done承诺；历史首失败不重写。

- [x] 正式bounds64→受控普通/SSE请求观察默认64/显式32与拒绝路径。
- [x] tenant/key RPM2边界；TPM B200/U7的206拒绝207接受、P200后406/407、跨窗399/400及未知carry。
- [x] known7、fee7/cost33与unknown预留200分列；失败/未dispatch/超时/取消及同ID不重放证据。
- [x] 8屏障竞争者1受理7拒绝、双Key/轮换稳定池/跨tenant隔离及各维度无部分扣计。
- [x] 次数/并发/预算独立语义、改额/价格/暂停/重启守恒和全部原回归保留。
- [x] 固定完整门禁tests9 exit0、独立计划A/B评审及QA10受影响验收完成，未验范围和责任明确。

## 实施结果与下一责任

tests9真实完整 `make verify` exit0/698.502秒，原720秒未变；HTTP/race/进程和1280/1440/390受控Go UI完成。QA10独立8组InProcess race exit0/29.082秒、3项Node断言通过，指定计划与证据评审完成。[QA10结论](../acceptance/release-assessment-qa-10.md)、[逐项矩阵](../acceptance/acceptance-matrix-qa-10.md)、[独立计划评审](../acceptance/test-plan-review-qa-10.md) 是tests9后的权威验收。

版本：HEAD仍 `3f9cee1615e87e002a2c2300f2d29778072ec3f5`，受测154文件工作树SHA256 `2dfe14a3c305c13621f23b7dd287a8313b7e4f39ab64b8281d283450d125d8ed`，report执行11重算一致。没有新commit、PR、merge或部署，不用基线冒充整改提交。

确认生产缺陷0，DS-SAFETY“服务端无默认cap”内部未复现；没有生产源码/协议/schema/模块锁/价格/Makefile改变，无迁移/配置/接口变更，无设计偏差。八组回归、默认nil fixture和原治理/资源矩阵等断言均保留；执行8仅为18个既有计量相关测试补有界调度。测试工程首失败和修后门禁绿见 [执行6](../verification/host-return-exec6.md)、[执行8](../verification/host-return-exec8.md) 及QA10，不是RPM/TPM产品故障红绿。

[研发results](../verification/results.md) 和 [测试计划](../verification/test-plan.md) 保留旧研发快照，旧“待宿主/待评审”文本不作为当前任务状态。原生沙箱监听限制及tests5/7失败单列保存，不由静态页面替代后台。

## 交接

- 当前可依赖：M01内部回归资产已独立验收；下一责任publish Connector提交推送、建立PR并保留正式候选/Actions门禁；[PR正文](../pr.md) 按任务分支链接，推送前不称远端可访问。
- 未完成：原#33两屏非零RPM/TPM、自然供应商窗口、双浏览器竞争、真实普通/SSE成功/恢复及Token/费用/实扣Not Run/Blocked，原Incomplete-No-Go保留；计划C仅安全前置，不是可执行供应商预算计划。
- 专用供应商Pipeline取得合法配置、精确模型/收费输入输出及隐藏推理上界、思考策略和H/X历史预算，按原24/64/90独立评审后复测；新增真实生成0，未知不得归零。缺私有前置不要求协调者代验或广泛提权。
- 最近成功验证：tests9完整门禁及QA10独立验证，日志/receipt/hash见 [证据索引](../acceptance/evidence/index.md)。门禁余量约21.5秒；源码/测试/锁/契约变化需重验，纯报告更新不冒称重跑门禁。无需要用户重新批准的普通选择。
