# 执行29：暂停费用守恒最小诊断

候选pause-conservation-diagnostic-r1；日期2026-10-08；完整交付仍Blocked by Validation。恢复已确认design-v0.2.0/architecture-r1/T001/QA22-01及QA22-02，不重建方案、编号或AC。固定入口仍make verify，无需求/设计返工或用户澄清项。

## 材料、规范与边界

材料同prepare：`.workflow-input/4cd62ced6a4f98f50e3cca6f7596eb06/5f51d35fdc15ed7214749f9413879b4238097f1d6b6bbe2cc0d508d8e103cef7/files/manifest.json`；version design-v0.2.0；archive SHA256 5f51d35fdc15ed7214749f9413879b4238097f1d6b6bbe2cc0d508d8e103cef7。实际引用root下docs/prd.md、docs/interactions.md（J05）、docs/acceptance.md、docs/decisions.md，逐文件摘要在material-source.log。未执行包内脚本、修改冻结材料、读取旧Run或New API源码。

trellis-before-dev运行get_context --mode packages得到单仓库backend/frontend；读取两层index及runtime/console/routing/multitenant和DOM/lifecycle/verification/workspace实质规范、guides及两个thinking guide。已有初始化目录和源码支持规范可用，prepare initialized=false仍保留，不伪称首次受信初始化通过，不重建空模板。沿delivery Skill两份references恢复公共任务；本轮仅继续同T001/M01和验证记录，无组织级新增结构。trellis-update-spec维护workspace契约，trellis-check核对改动范围、数据流、秘密投影、原门禁与新增回归。

实施前的最小边界：seq28 J05恢复读取200后费用比较失败，缺少比较值，尚不能断言是生产计费错误。行为所有者分别是生产manageLimits/usageFor/admit/settleAdmission及实际连续driver；仅在console_contract_test.go增加原生产路径守恒回归、journeys.mjs添加每个拒绝/隔离边界只读核对和白名单投影、journeys.test.mjs验证实际辅助段。生产Go/UI、DB、角色、计价、暂停/恢复协议和时间均未改。保留原恢复费用相等断言；新增calls/pending/concurrency/reserved和period守恒更严格，未削弱门槛。

## 已证与未证

- 宿主失败见host-seq28-summary.md；J01–J04部分绿不能变成完整通过。
- pause-diagnostic.log：新增用例首次exit1，停在Key签发409，因为测试缺少Idempotency-Key；这是新增测试缺陷，不是seq28产品根因。补唯一操作ID后pause-diagnostic-2.log exit0。两份原始失败/成功均保留，未改生产强制幂等契约。
- TestConsolePauseUsageInProcess调用实际schema3管理/proxy/SQLite，以明确受控transport fixture生成非零费用；两租户、第一租户两Key、普通/SSE拒绝零dispatch、第二租户独立收费、暂停/恢复费用/预留/调用数不清零及恢复后追加调用通过。不能证明Chromium/网络机制或真实供应商。
- 新TestConsolePauseUsageActualHTTP进入原go test门禁，仍待宿主运行；同场景实际临时HTTP listeners，不是更弱替代。
- 新usage_facts仅期间相同布尔、fee/reserved十进制字符串和calls/pending/concurrency安全计数，非法转null并失败；不输出ID/账户/消息/原始API内容。分段J05-paused-ordinary-usage/stream/standby/other-tenant-usage先定位差异首次出现处，再执行原resume费用断言。VM测试实际函数段确认大于Number安全范围金额不丢精度、各差异/非法值失败和秘密不进入投影，但不作为原native机制证明。
- 原生环境未改变，不重复已知稳定监听失败make verify；完成可运行补充检查后交配置宿主。真实供应商/外部身份/支付Not Run，发布/Harness/实际模型证据仍由相应平台回执判断。指定Codex/gpt-6.1-sol不是实际模型证明。

## 宿主下一步及独立QA

对source-evidence.json同指纹运行原make verify并保留before/after指纹、完整非截断日志、exitcode。新增actual HTTP回归必须运行，原所有Go/Node/health-history/迁移/HTTP/SSE/取消/keyboard保留。连续1280/1440/390各需J01–J07七项actual passed；保存continuous-{width}-{pid}.json及五类脱敏截图到本Run验证目录。若失败，保留新的usage_facts/stage/action/call_facts，用首次差异区分产品状态、驱动范围、异步边界；不要直接改计费或把fixture通过当原宿主根因。全部通过后独立QA按同材料/指纹逐AC01–AC18/H01–H08和八类商用/J01–J07五层状态复验QA22-01/02。本节点未关闭缺陷、未宣布商用对齐/完整交付/发布。

具体补充命令和真实退出码见checks.log；当前源码版本见source-evidence.json。T001 In Review、M01 Ready for Review；此轮是获取归因所需的诊断交接，不是最终Done。

## 最终可运行检查结果

`make format-check`、`go mod verify`、`make vet`、`make frontend-test`（161 tests/161 pass/0 fail）、定向四项顶层Go race回归（PauseUsageInProcess、PeopleScopeInProcess、AccountingInProcess、ContinuousUpstreamResponseContract）、`make build`、browser测试编译、journeys.mjs语法、`git diff --check`均exit0。journey-test.log单文件九项通过。链接器LC_DYSYMTAB warning保留，临时本机路径已脱敏，不伪作失败或隐藏警告。新actual HTTP回归与完整native UI尚未运行，源文件和编译通过不算实测Pass。

最终HEAD98511771871cf0951ecef716bbb55deb13e18da5；124文件、Go go1.25.6；工作树SHA256 `6263734f52bb1631801f5db0f27d44f7a12de0e64f1fad4ffdf003c244b6dd92`。本Run原生产候选未提交，HEAD不能代替完整指纹。材料archive和四份实际引用文档SHA匹配manifest。新增证据扫描未出现凭据或绝对本机路径；未提交、切分支、推送、评论Issue或部署。
