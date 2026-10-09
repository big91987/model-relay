# T001 安全路径与资源计量可复验

- 状态：Done（2026-10-10，tests22＋QA23联合验收，report24更新）
- 里程碑：M01；负责人 development；评审人 QA23；开始2026-10-09。
- 模块 internal/relay；优先级 High；执行类型 Validation + Review。
- 硬依赖：[总览](../README.md)所列材料及既有Go实现/固定make verify。
- 权威输入：原AC09–13/16及关联AC07/17、PLAN-01–12；本Run测试计划及项目Trellis计量契约。
- 软依赖：真实供应商前置只影响外部验收，不阻塞受控回归。
- 冲突边界：测试fixture与交付文档；整合保留PR35，不接管其任务或改冻结设计。
- 交付物：Go受控回归、数值采证、完整联合门禁回执、独立评审及公共说明。

## 目标与变更边界

当前缺口首先是测试资产/证据，不是已确认产品故障。行为所有者为 accounting.go 的 admit/usageFor/boundedPayload/settleAdmission，protocol/proxy负责实际请求及终态。补测前不改生产实现。

实际修改 `internal/relay/console_contract_test.go`：允许同一个隔离fixture选择明确的受控HTTP/SSE响应，不改默认或现有断言；新增 `internal/relay/metering_regression_test.go`：实际上游请求cap、非零RPM/TPM、固定窗、已知/未知及共享池的数值断言；更新相关Trellis规范，保留所有旧门禁。测试自动进入make verify既有Go race阶段，不增旁路命令。隔离fixture调度使用原parallel=4，浏览器补只读冷却诊断及测试，不改产品前端。

非目标：调用表单新增猜测控件、供应商替换或重命名、扩大预算、数据库手改、平台功能、报告Dashboard。必要时先同层复现红，再最小修产品；没有复现不虚构红绿。


## 验证契约


受控单候选输入上界100/输出100（安全专测64）；费率1,000,000微美元/百万tokens；已知usage3+4=7；B=200、fee预留200。RPM按服务端固定UTC分钟，TPM=本分钟已知U+跨窗未结算P。等值U+P+B接受，少1拒绝且租户/Key均不部分扣计。并发使用开始屏障与阻塞上游，拒绝先读回再释放，不用串行冒充竞争。回归使用临时隔离安装，不用SQL更改业务数据。

安全链必须从正式管理bounds配置到实际受控上游payload，同ID诊断和池数值关联。有限TPM/硬预算与无限路径分开；设置cap本身不是输入/推理收费上界的供应商证明。

固定入口 `make verify` 在隔离临时目录/端口、受控上游执行；Go1.25.6与锁定npm依赖，完整环境与恢复见[QA23报告](../acceptance/release-assessment-qa-23.md)。成功要求原门禁全部exit0、数值/副作用与矩阵一致、源码前后一致；失败保留日志及同版身份，不削弱断言或手改库。受控可执行用例与人工网页步骤见[测试计划](../verification/test-plan.md)、[独立计划评审](../acceptance/test-plan-review-qa-23.md)、[旅程](../acceptance/core-journeys-qa-23.md)。

## 验收结果

- [x] 正式bounds64→实际受控普通/SSE默认64/显式32与拒绝路径。
- [x] 双scope RPM2边界；TPM B200/U7下206拒绝207接受、P200后406/407、跨窗399/400及未知carry。
- [x] known7/租户fee7/成本33与unknown预留200分列；失败/未dispatch/超时/取消、重复同ID不重放。
- [x] 8屏障竞争者1受理7拒绝、双Key/轮换同池/跨tenant隔离及各维度无部分扣计。
- [x] 次数/并发/预算独立语义、改额/价格/暂停/重启守恒及全部原回归保留。
- [x] PR35 main与PR36受控资产联合完整冷门禁tests22通过；QA23独立计划A/B及受影响复验完成。

## 实施结果与版本

prepare18整合main `6929376089535038112da48aed715fd729615d76`；tests22于上海2026-10-10 00:22:25–00:33:19及其后QA23/report24采证时，旧HEAD为 `8917f942cb38af2f56aa93914ab62e763bf9cc6e`，联合工作树当时尚未提交、冲突0。publish25已于上海2026-10-10 00:55:39形成并推送联合提交 `5602f563313427ea9b1c7fbd495f5ce60a1de10a`（exit0/pushed=true），pr26已更新同PR36。report28接手工作树干净；158文件源码指纹 `6e359667d5f7521f6b8c421454c6fc7cbe5ccb8c74aa87bb1e5955c194c202cc` 与原验收一致，361件QA归档bytes/SHA及联合提交对象核对一致。tests22完整冷make verify约653.728秒exit0，278项前端无skip、全包race/真实HTTP/进程/build与三宽受控Go UI全绿。QA本人8组进程内race 21.991秒exit0、53项Node exit0，未声称本人重跑HTTP/浏览器。

[QA23结论](../acceptance/release-assessment-qa-23.md)、[逐项矩阵](../acceptance/acceptance-matrix-qa-23.md)、[证据索引](../acceptance/evidence/index.md)为当前完成依据。生产实现确认缺陷0；内部“无默认cap”未复现，没有生产算法红绿，只有采证/测试工程历史失败及当前完整门禁绿。无本任务新增生产接口/schema/价格/供应商改动；联合主线行为保留。设计偏差无；不重做PRD/架构。

历史[tests5](../verification/host-return-exec6.md)、[tests7](../verification/host-return-exec8.md)、[exec17阻塞](../verification/integration-blocked-exec17.md)、[exec21记录](../verification/integration-review-exec21.md)保留；旧公共状态原文见[历史快照](../verification/public-state-before-report24.md)。旧tests9/QA10只适用单候选，已解除的环境阻塞不再作为当前任务状态。

## 交接与限制

- 可依赖：联合M01受控资产已独立验收，publish25已提交推送、pr26已更新同PR36；report28补正[PR正文](../pr.md)及公共当前状态，原publish/pr仅需提交推送本轮文档并更新同PR36，Refs #34，不新建PR。最终新HEAD及正式Actions映射待Connector，不预编SHA，不声称PR36已合并/部署。report24版本保留于publish25提交对象，原始证据不变。
- 未验：原#33两屏非零RPM/TPM/自然窗/双会话、真实供应商安全/隐藏推理/思考、成功/恢复/费用实扣Not Run/Blocked，Incomplete-No-Go保持。
- 专用测试责任Pipeline核合法配置、精确模型/上界/思考和H/X预算再独立评审复测；C仅前置图。budget not_attested、24次/每次64 tokens/90秒及未知历史不变，新增真实请求0，无私有Provider读取。
- 最近成功验证tests22/QA23，不再等待已完成的验证；约66秒门禁余量有风险。源/测试/锁/契约变化需重验，纯报告更新不冒称重跑。普通阶段无新增用户决定。
