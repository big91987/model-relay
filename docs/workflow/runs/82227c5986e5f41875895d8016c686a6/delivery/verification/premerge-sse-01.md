# PR #19 合并前两项 SSE Important 返工

## 对象、基线与范围

2026-10-06，研发节点接同Issue #18／Run返工。固定已审HEAD为 `8ae4e5173aa0dc45c28ad61dbc64df20afe229a6`（Connector已推送并创建Draft PR #19；未合并部署）。需求仍为[PRD](../../product-definition.md) R-3/R-4/R-5、AC-05～09/12；设计仍为[路由架构](../../routing-architecture.md) 协议白名单、逐尝试及终态合同，无需返回需求或设计。

最小行为缺口：Scanner超长行是大小／协议错误，不能当暂时传输故障；已记success后DONE下游失败／取消应纠正终态并清成功归属。实现拥有者为proxy.go的Scanner和请求finish闭包、routed_proxy.go的streamFinish。只改这两个实现文件，加一个测试文件，补原Trellis合同和公共D3/D5状态；不改schema、调用API、调度策略、限额、UI/driver、原门禁、平台源／安装／Workflow／DB，不加兜底，不重做需求设计或提前路线3/4。

## 有界诊断：事实与假设

- 平台提供独立已审HEAD管理/代理Handler＋隔离Transport＋真实SQLite红控，不是实际网络或供应商验收。本节点在产品未改的该HEAD上添加回归，独立复现相同错误。
- [初次命令](sse-terminal-red-01.log) exit1是新增测试误用不存在的roundTripFunc导致编译失败；改为仓库已有roundTrip，保留首次失败，不归为环境或产品问题。
- [正式红控](sse-terminal-red-02.log) exit1：首超长行访问2次Transport、请求success、2 attempts；提交后超长行错误code是stream_interrupted。DONE写错误及取消带错误仍留下成功上游；DONE写内取消但返回nil时请求和attempt错误保留success。success对照不失败。命令为 `GOCACHE=.cache/go-build go test -race -count=1 -run '^TestSSE(OverlongLineNeverSwitchesOrCools|DoneTerminalAttribution)$' ./internal/relay`，Go1.25.6。
- 可证根因：Scanner.ErrTooLong进入classify(stream=true)，而routing将code stream_interrupted列为暂时故障；streamFinish只在success设置归属但从不清先前指针；DONE成功保存后写失败／写后取消缺完整纠正。不是locator、异步按键或预算环境假设，无需换驱动或加sleep。

## 正式修复与退出断言

1. Scanner.ErrTooLong显式归类 `stream_interrupted / upstream_protocol_error`；HTTP提交前502、提交后仍200但流不完整。错误码不可切换，协议错误不触发候选冷却，限额不变。
2. 请求finish在父取消覆写终态后，统一对非success清空 `SuccessfulUpstream` 再落盘；streamFinish非success亦清归属。DONE写前取消／写失败／写成功返回后已取消均纠正attempt和request；只有success可形成概览证据。
3. 新 `sse_terminal_regression_test.go` 自动纳入原 `go test ./...` 和固定 `make verify`，不删除／放宽已有断言。免监听管理/代理Handler＋SQLite读回核对同Request ID：请求和attempt精确终态、成功上游null、概览call_evidence=null；success对照仍有正确upstream及概览。双候选首／提交后超长行各只有一次Transport、无备用、无冷却。
4. 新 `TestSSERealHTTPOverlongLineNeverSwitchesOrCools` 提供真实隔离HTTP首／提交后超长行，核对primary=1、secondary=0、精确协议错误和无冷却；此处尚未执行，留宿主原入口。DONE失败使用确定性下游Writer，仅证明代码错误返回／取消边界，不冒称真实TCP下游故障重现；原真实HTTP取消和双屏旅程完整保留。

## 实际验证与版本

| 命令／证据 | 实际结果 | 证明范围 |
|---|---|---|
| [专项green-01](sse-terminal-green-01.log) | Go输出PASS；首轮管道未启pipefail，不单独用其进程码放行 | 辅助历史输出，已完整复跑 |
| [专项green-02](sse-terminal-green-02.log) | exit0，race，2个顶层／6个子场景通过，15.911秒 | 同版实现Handler/Transport/SQLite协议和DONE归属；非网络 |
| [相邻契约回归](sse-terminal-adjacent-01.log) | exit0，race，44.888秒 | 权重优先级、临时白名单、多尝试未知、取消停止、被动冷却、概览精确查询；非网络 |
| [项目局部检查](sse-terminal-validation-01.log) | exit0，119/119前端VM通过 | format、modules、vet、全包零测试编译、build、VM、diff；零测试编译不是行为验收 |
| [最终专项green-03](sse-terminal-green-03.log) | exit0，race，2个顶层／6个子场景通过 | 新测试参数统一为testContext后重跑format/vet/全包零测试编译和专项race；仅命名变化，行为不变 |
| [最终版本13](final-source-13.json) | 当前待验工作树摘要 | 含新增回归、规范和公共状态；[12](final-source-12.json)及首[审计](sse-terminal-audit-01.log)保留命名整理前快照，旧seq21/version11不能代用 |
| 新固定 `make verify`／独立QA | Pending Host／Not Run | 900秒完整入口→同版本QA逐项复验；不提前Accepted |

本机稳定listen/Mach沙箱限制没有相关变化，按既有[首次限制回执](make-verify-01.log)不再次跑必然失败的完整入口，不扩权。新HTTP用例已编译但待宿主执行，所有尚可执行局部检查完成后next交既有tests Connector。原seq21完整成功和[QA report-02](../../acceptance/report-02.md)、[matrix-02](../../acceptance/matrix-02.md)、所有原失败／QA01保持历史证据原样；不改测试预期，不把历史Go写成新版本Go。

## Trellis及交付审查

before-dev已运行项目get_context packages（single-repo），读取真实backend索引、runtime/console/routing合同及reuse/cross-layer指南；规范已有源码依据，无需bootstrap或补造初始化。update-spec在routing-contracts加入7节可执行SSE边界合同。check已复核：错误类别→temporary→observe无协议冷却；finish→SQLite→管理详情→overview成功候选失效闭环；复用roundTrip/consoleFixture/真实HTTPfixture，无新依赖、迁移、秘密日志或无关重构。代码通过格式、vet、编译和局部race；完整网络与浏览器证据未冒充完成。

公共任务仍使用原D3/D5，状态In Review；M02/M03 Ready for Review，整体Verifying；未改任务保留原接受范围，无新里程碑／重复事实源。PR正文和唯一路线同步当前待验，不覆盖QA历史。任务负责人研发、评审独立QA；退出条件是新宿主完整门禁及QA，关键路径及范围没有需求／设计偏差。

## 宿主与QA交接

核对最终版本13产品摘要和[模板来源](platform-template-provenance-01.json)中Workflow／manifest联合身份，再用既有900秒 `make verify` 完整执行。保存新日志、真实exit、Go/Node和源码摘要于本Run，不覆盖旧证据。失败保存精确test/subtest/断言，先最小归因产品、测试或环境，不换更弱入口。QA逐项核对超长行首／提交后无备用与无冷却，以及DONE写失败／两种取消的请求、attempt、成功上游和概览一致；保留成功对照及原全部验收范围。

成功后走QA→报告→原发布Connector更新原PR #19，本节点不提交／推送／切分支／建Issue或PR／合并／部署。可信controller新旧Go联合链、正式Actions非空升级及可信隔离失败恢复、真实供应商仍Not Run；平台维护源和授权发布边界不变。
