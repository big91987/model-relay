# #34 精确候选独立验收报告

> 日期：2026-10-09，Asia/Shanghai；Run 637705186406b80a2ee4afaf0d6b2167，qa执行10。
> 仓库：big91987/model-relay；分支 workflow/637705186406b80a2ee4afaf0d6b2167。
> HEAD：3f9cee1615e87e002a2c2300f2d29778072ec3f5；**这是基线提交，不是已经提交的整改版本**。
> 受测候选：154文件工作树摘要 2dfe14a3c305c13621f23b7dd287a8313b7e4f39ab64b8281d283450d125d8ed，包含未提交/新增测试与规范；QA产物不在此源码指纹范围。
> 模式：受影响安全/计量全量验收，加原公共完整门禁；真实供应商部分只读评估。
> 阶段结论：**Go，仅限M01内部回归资产及测试调度整改交付，可整理PR交付说明**。
> 产品/原#33结论：**Blocked；历史Incomplete-No-Go保留，不宣布真实调用或费用验收通过**。

## 1. 结论摘要

tests执行9的真实完整 `make verify` 成功，内部 `make verify-gates` 和Connector均exit0；QA重新计算源码指纹与前后回执完全一致。新增安全、非零RPM/TPM、已知/未知计量、并发和池隔离已有受控HTTP数值证据；QA独立复跑8组无监听计量race与3项诊断断言通过。原完整真实Go UI回归在1280/1440/390通过，但它不是#33指定DeepSeek两屏生成/非零RPM/TPM旅程的替代证据。

当前确认生产实现缺陷仍为0；没有为“关闭Issue”改产品。DS-SAFETY-01关于“缺调用控件意味着后台没有安全默认”的推断未成立，内部API已动态观察默认max_tokens，真实供应商收费/隐藏推理的完整安全路径仍Blocked。此交接完成的是独立QA责任及内部整改交付，不关闭外部验收，不把未运行项记成Pass。

## 2. 本轮覆盖与基线

材料由prepare提供，QA重新核验ZIP及全部61项大小/SHA256，结果一致。版本 `product-test33-handoff-r1`，交接ZIP SHA256 `a13fb6b9089cefbb9d2b6f9a333830e0ee6866bd7ff45d5ba881da97e9a66a49`；原设计ZIP声明 `5f51d35fdc15ed7214749f9413879b4238097f1d6b6bbe2cc0d508d8e103cef7` 保留，未声称此次重建原ZIP。实际路径、逐文件摘要在 `evidence/qa-10/material-integrity.json`，不是从别的Run或个人机器恢复材料。

权威原文是同材料根 `design-v0.2.0/docs/prd.md` 的配额/预算/计量条款、`acceptance.md` AC07/09–13/16/17、`interactions.md` J01–07及模型上界入口；逐字节未改。固定UTC分钟/受理计数/未知carry及有界默认等架构细则引用项目已有 `docs/workflow/runs/4cd62ced6a4f98f50e3cca6f7596eb06/implementation-contract.md` §3–5，不以实现反写规格。读取原report、independent-review、DS-SAFETY候选、PLAN评审/整改表，历史规则与结论没有重置。未发现需要返回requirements/design的实质冲突。

覆盖角色包括隔离安装的真实平台管理员、租户管理员、开发者和财务观察者；对象/账户由正式UI/API创建，测试专有受控上游不是产品默认。核心网页旅程按角色独立会话连续完成，跨角色授权是显式交接，不冒称单人单会话完成所有权限动作。J01从空服务/模型/Key安装开始；J02日常调用/授权收窄及恢复；J03轮换/提前撤销/秘密清除；J04硬预算拒绝与改额恢复；J05暂停、角色收回、另一租户隔离；J06改价、未知核对、结账/更正；J07分页/作用域晚到/故障恢复。详细UI与最终事实见 `core-journeys-qa-10.md`，测试期时钟与故障注入已披露。

## 3. 缺陷及进展

- P0/P1/P2/P3确认生产缺陷：均0；不为漏测制造业务缺陷，也不把测试工程失败升级为生产故障。
- tests5与tests7原完整门禁失败保留。tests7受测078e1aa1候选累计十二分钟在full-console1440 J03被终止，不是独立产品J03红。tests9新2dfe14a3候选相同入口/断言/时限完整成功，关闭本轮**门禁未完成**阻断，不倒改历史结果。
- 本轮只核对已有测试并行调度和回归资产，无生产源码/协议/锁/价格/供应商选择/Makefile改动。不能称修复了原先未确认的RPM/TPM算法故障。
- DS-SAFETY-01保持待评供应商能力候选；内部“不存在默认cap”未复现；原控件检查及真实供应商未验不删除。责任及复测包见 `findings-and-next-actions-qa-10.md`。

## 4. 实际证据与执行来源

宿主tests9：2026-10-09 13:13:47.905552Z–13:25:26.407351Z；隔离冷安装，node_modules/cache初始不存在，NODE_PATH unset，无手工预装。Go1.25.6 darwin/arm64、Node26.0.0、npm11.12.1、Playwright/core1.59.1、Chromium147.0.7727.15；本地隔离schema3及测试构建，不是正式线上部署，未提供发布二进制或merge SHA。

完整入口包含source/history/health-history、格式、锁定模块校验、vet、npm ci、前端测试、全包race、build、迁移/旧UI/routing/console/full-console。relay包304.603秒；7组新增ActualHTTP计量均Pass。浏览器阶段耗时14.841/97.695/131.265/15.432/52.717秒，所有相位完成；总回执耗时698.502秒，未改内部720秒或外层900秒限制。单次成功不承诺所有宿主负载下都具备足够余量。

`evidence/qa-10/host-tests9/state.json` 和 `verify.log` 是本工作区宿主保留的原文件，日志499499字节、SHA256 `c93e9625fdd4cd5c51ebf97c7838cd37c887972ccf672e67182d6b5ce4484021`，QA逐字节校验与state一致。Connector回执499596字节、未截断；其头/尾通过平台日志工具核对。另存的 `connector-output-reconstructed.log` 是97字节已见包装+原runner日志的**重构件**，不是独立取得全包装原始日志；不宣称中间包装日志逐字节比对。原完整runner文件、state、退出值、前后source和阶段输出共同作为成功依据，不以截断展示或缺日志推断成功。

日志中11处 `FAILED:` 属 `tests/verify` 内部隔离故障注入预期，通过对应顶层测试确认，不混作产品红，也不单独当成功证据。外层 `TestConsoleClockRecoveryWorker` 明确Skip，矩阵记Not Run；同轮父测试启动真实子进程、校验CLOCK_RECOVERY_PASS，另行记录恢复结果Pass，而非把Skip改成Pass。

QA亲自执行：原source-evidence默认使用个人Go缓存时权限拒绝；设置工作区GOCACHE后exit0且同154文件摘要。`git diff --check` exit0；`go test -race -count=1 -v ./internal/relay -parallel=4 -run '^TestMetering.*InProcess$'` exit0/29.082秒（8组，非监听/非外部）；两项Node测试文件3个断言exit0。额外初次Node smoke含selection-trace普通模块，仅3个业务断言，不多算一个验收用例。本节点没有重复宿主监听完整门禁，没有使用静态browser.check冒充后台网页测试，也没有请求提权。独立日志/检查/采集失败及脱敏转换保留在evidence/qa-10。

## 5. 未执行与环境限制

专用配置入口/凭据、精确deepseek-v4-flash/pro当时可用性、可信输入/收费输出及隐藏推理上界、思考策略和跨attempt历史预算未获证明；真实普通/SSE成功、成功恢复、自然RPM窗口、真实TPM/费用/供应商实扣、双浏览器并发及多Key/tenant真实消耗池均Blocked/Not Run。**新真实生成0**。原24次跨attempt总量、每次64 tokens、90秒不扩大，历史未知不清零；没有已证剩余预算就不能安排新生成。

新非零数值测试是Contract/API-E2E + 数据对账，不是通过网页输入非零RPM/TPM的UI-E2E。本轮未执行#33原角色/精确模型/两屏数值网页全链；该UI行保留Not Run，绝不将后台用例与普通UI截图拼成Pass。受控注入窗口证明精确边界不变量，不证明自然供应商一分钟重置；routing自然30秒恢复证明候选冷却而非RPM窗口。真实供应商不可故意无界造失败/取消/未知usage来追求绿色。

## 6. 阶段边界

本任务是#33报告整改，用户已授权直达development；不机械重建需求/设计、不实施未承接里程碑。#32管理术语/首次自托管工作和#30保存同步仅运行已有回归，不接管在途分支。独立评审只批准 `verification/test-plan.md` A/B作为受控可执行方案；C是正确的停机/安全前置说明，**不是已批准可执行的供应商预算计划**。未执行平台Harness能力、Dashboard发布、供应商切换、Git发布/Actions/部署；这些不能补造Pass。

## 7. 环境恢复

宿主回执owned_group_absent=true，临时候选已删除；浏览器隔离目录/服务/contexts清理声明随原JSON归档。QA无共享环境业务写入，InProcess测试使用自己的临时目录；缓存/原始本机路径日志留忽略目录，公共归档仅脱敏内容。PNG由测试安全截图函数断言secret字段/浏览器存储清空；QA抽看1280及390最终屏，不宣称人工逐张视觉全审。未留运行中测试服务，未改生产SQLite或用户上游记录。

## 8. 质量趋势

**内部验证能力提高**：原PLAN漏项转成可重复、带数值、带同ID/零dispatch/双scope断言的回归，并接入固定门禁；原两次宿主不完整变为本轮完整成功。无新增/复发确认生产缺陷；不能因没有产品改动就说“产品修好了”，也不能因受控绿说真实供应商质量已被证明。剩余风险是外部前置、原数值UI未验及门禁时间余量，分别保留责任。

## 9. Go / No-Go与交接

M01内部交付Go：同候选固定完整门禁通过、独立受控计划评审完成、QA受影响回归通过且无确认实现缺陷，可以交report整理明确范围的测试资产PR说明。产品真实DeepSeek验收Blocked，原#33 No-Go不改，不自动关闭Issue。report必须带未验项、原预算和版本摘要；后续publish/PR由Connector办理，源码/测试/锁/契约变化需重验。merge/Actions/部署及专用供应商复测不是本节点已完成的事实。

当前权威矩阵：`acceptance-matrix-qa-10.md`；独立计划评审：`test-plan-review-qa-10.md`；证据索引：`evidence/index.md`。研发执行8的“待验/Ready for Review”是历史状态，本报告是tests9后新的验收快照，不覆盖历史失败。
