# Issue28 干净安装交付计划

## 目标与边界
修复新安装缺少浏览器依赖导致验证失败的问题，完成同版冷验证与独立QA，并按已有授权推进新PR、精确合并及正式preview部署。基线main为 `cd413b5aa24a372a19c19af999a6ef61b5c41572`；Run `3e7496815b9922704072482cf44e73e2`、冻结工作流版本2、Issue28。不得恢复Issue25/旧Run或重复PR27。

产品输入沿用design-v0.2.0，版本、材料摘要、实际引用路径及权威架构见 [T001](tasks/T001-clean-install.md)。本次不改产品范围、后台运行时、数据库、UI资产、依赖锁或部署控制器。仅一个Bug任务和一个里程碑；需求、架构与产品设计阶段不重复执行，不补造新PRD/设计。已有Trellis规范继续维护，不据prepare.initialized=false重初始化。

## 当前状态
- 交付判断：**Executing**。
- 当前里程碑：M01 **Ready for Review**；冷验证与独立QA已完成，正式发布部署尚未完成。
- 关键路径：已验候选 → Connector提交推送/新PR → 授权审查精确合并 → 新main正式prepare → Owner deploy=true target=preview → Deployment、healthz与真实admin闭环。
- 当前阻塞：无阻止该候选进入发布的已证P0/P1；P2摘要验证可靠性风险Open，不冒称已接受风险或已修复。正式prepare若再红不得越过。
- 下一责任：合法publish Connector；不是重新等待tests/QA。不再询问是否批准下一阶段。
- 候选：147件，SHA256 `f8befd158a543afeb1ba77778e4caf72e876f6e4b0b988dd4e6ab73de5e5eaab`，当前HEAD仍是旧base，尚无新提交/推送/合并/部署回执。

## 里程碑与任务投影
| ID | 可观察结果 | 状态 | 依赖 | 证据 |
|---|---|---|---|---|
| [M01](milestones/M01-clean-install.md) | 干净安装通过完整验证并具备正式交付证据 | Ready for Review | T001与授权发布/部署 | tests20冷全门禁、QA21；正式上线证据待补 |
| [T001](tasks/T001-clean-install.md) | 锁定依赖先于消费者并完成发布闭环 | In Review | 宿主工具链、独立QA、发布Connector | [QA报告](acceptance/report-qa-21.md)、[矩阵](acceptance/matrix-qa-21.md) |

## 已完成与剩余范围
`make verify` 自建私有无模块/冷缓存候选，锁定npm安装先于消费者，保留所有历史红分类器、Go/race/HTTP、迁移与浏览器门禁。tests20正式宿主exit0；QA21完成同版证据审查和局部race/197Node/static复验，结论 **Go with known issues，仅候选进入精确发布**。完整runner为310718bytes，与Connector310815bytes不同；原样日志/state及三视口J01–J07/summary/visual见 [逐件归档](acceptance/evidence/qa-21/archive-index.json)。报告22复核148件归档及源码/锁/材料一致，见 [审查记录](verification/report-22-review.json)，不是新增全门禁。

QA原生完整入口监听Blocked、动态独立浏览器Blocked、供应商Not Run分别保留。顶层clock worker为子进程专用skip，不记Pass，实际HTTP/in-process恢复覆盖见QA矩阵。正式prepare、Owner部署、Deployment/environment_url、healthz新版本/schema3、真实admin和旧数据密码保持尚未取证，因此不是Delivered。

## 主要风险与责任
- [summary历史独立评估](acceptance/summary-assessment-qa-21.md)：seq18原断言失败、根因未知；development19仅补诊断。seq20一次真实绿不关闭历史。P2验证可靠性风险Open，产品持续错状态P1未实证；再红按原层最小GET屏障/同步契约返研发，不盲目重跑或放弱断言。
- 原生沙箱限制不等于宿主通过；受控HTTP/SSE不等于真实供应商能力。
- 通用控制器同SHA日志覆盖归Platform维护源；本仓不复制或修改控制器。
- 下游不得init/reset旧安装、改密码/DB、启临时替代服务、扩大权限；发布授权不等于已经完成。

## 证据与交接
[验证索引](verification/index.md) → [QA21报告](acceptance/report-qa-21.md) → [发布闭环契约](acceptance/release-handoff-qa-21.md)。对外PR正文为 [pr.md](pr.md)，其仓库URL指向任务分支，报告时尚未推送；发布Connector须先确保对应文件进入精确提交并推送，不能链接旧main。

历史返工、失败和中断在原 [验证记录](verification/index.md)、QA6与各轮remediation文件中保留。公共状态以本总览投影、T001事实源和M01门槛为准，不再叠加相互矛盾的“当前”段落。QA21交接文件哈希是当时快照，报告22仅更新本Run交付文档，不覆盖该快照。

2026-10-09交付审查：**Ready with Non-blocking Gaps**，可进入合法publish；冷证据缺口已闭合，剩余已公开风险与正式发布/部署必须继续，不宣称任务全部完成。
