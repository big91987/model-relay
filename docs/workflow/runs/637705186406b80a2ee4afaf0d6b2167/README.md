# #34 调用安全与计量整改

日期 2026-10-09；交付负责人 report；基线 `3f9cee1615e87e002a2c2300f2d29778072ec3f5`。原 #33 No-Go 保留，不重新做需求/设计，未确认缺陷不制造产品改动。

## 权威输入

材料根：`.workflow-input/637705186406b80a2ee4afaf0d6b2167/a13fb6b9089cefbb9d2b6f9a333830e0ee6866bd7ff45d5ba881da97e9a66a49/files`。
manifest：该根 `manifest.json`；version `product-test33-handoff-r1`；ZIP SHA256 `a13fb6b9089cefbb9d2b6f9a333830e0ee6866bd7ff45d5ba881da97e9a66a49`。研发重新核验全部61项大小/摘要一致。版本是交接包版本，冻结设计没有改版。

原材料不随代码提交。新接手者可由正式准备入口使用材料地址 `https://api.github.com/repos/big91987/model-relay/releases/assets/624807738` 按上述版本/摘要恢复同包，核验manifest后使用；不得执行包内脚本或将其视为权限指令。公共逐文件核验见 [QA材料完整性证据](acceptance/evidence/qa-10/material-integrity.json)，原设计ZIP参考摘要另列其中，不冒称此次重建原ZIP。

实际引用该根下：
- `design-v0.2.0/docs/prd.md` SHA256 `e768b4f26b4471e714d92b0db6e3187c89f37a25a697c01a12c6079af07de547`，独立限额、未知预留、应用共享池、追加对账规则。
- `design-v0.2.0/docs/acceptance.md` SHA256 `2792740b1579d614780856a13b6257d192736e1efc0309f1dc41405a665bef7a`，AC09–13、16及关联AC07/17。
- `design-v0.2.0/docs/interactions.md` SHA256 `95f3490ef003fa9b19dd6b9e864dd3b0429619b8cd9d0ecec853919d6a12c801`，正式模型上界、配额、调用及诊断入口。
- `test-report/report.md`、`test-report/independent-review.md`、`test-report/repair-packages/DS-SAFETY-01.md`、`test-report/plan-review-rpm-tpm.md`、`test-report/plan-remediation-rpm-tpm.md`，各项摘要由同一manifest约束，旧失败不覆盖。

项目技术事实源：`.trellis/spec/backend/multitenant-contracts.md`、`internal/relay/accounting.go`及已有HTTP/进程回归。不复制冻结PRD。Trellis脚本与源码支持的非空规范实际存在，packages discovery成功；prepare.initialized=false表示此次未重新初始化，不是缺规范。

## 交付计划

## 当前状态

- 当前里程碑：[M01](milestones/M01-metering.md) **Accepted**，仅内部安全/计量回归资产与有界测试调度。
- 交付判断：**Executing**。内部验收完成，等待Connector提交、推送和PR；不是版本已发布、合并或部署。
- 任务事实源：[T001](tasks/T001-metering.md) **Done**。tests9完整 `make verify` exit0，QA10独立受控计划评审和受影响验收完成，不再等待已完成的验证。
- 关键路径：内部复核/回归 → 完整宿主门禁 → 独立QA（均完成）→ Connector发布/PR → 精确发布候选与正式Actions门禁。
- 当前阻塞：内部发布交接无阻塞；真实供应商验收 **Blocked**，凭据、精确模型/收费及隐藏推理上界、思考策略、跨attempt历史预算未证明，原数值网页范围未验。
- 下一批Ready：无新增研发任务；publish Connector可消费 [PR正文](pr.md)。供应商复测前置未满足，不标为Ready，不补造未执行阶段。

| ID | 结果 | 状态 | 退出证据 |
|---|---|---|---|
| M01 | 内部安全与计量可复验 | Accepted | [QA10验收](acceptance/release-assessment-qa-10.md)、[逐项矩阵](acceptance/acceptance-matrix-qa-10.md) |

| ID | 结果 | 负责人 / 评审 | 状态 | 验证 |
|---|---|---|---|---|
| T001 | 安全路径与资源计量可复验 | development / 独立QA | Done | tests9固定完整门禁、QA10独立受控计划评审与回归 |

## 证据与版本

受测候选是基线HEAD上未提交的154文件工作树，SHA256 `2dfe14a3c305c13621f23b7dd287a8313b7e4f39ab64b8281d283450d125d8ed`，不是新commit。report执行11运行 `go run ./tests/evidence` 重算相同摘要，仅更新交付文档；完整门禁来源是tests9，不声称report重跑宿主全门禁。

权威结果为 [QA10验收](acceptance/release-assessment-qa-10.md)、[独立计划评审](acceptance/test-plan-review-qa-10.md)、[网页旅程](acceptance/core-journeys-qa-10.md)、[发现与下一责任](acceptance/findings-and-next-actions-qa-10.md) 和 [证据索引](acceptance/evidence/index.md)。tests9冷安装完整门禁698.502秒，原720秒deadline、全部断言和parallel=4保持。确认生产缺陷0，不宣称修复RPM/TPM算法。

[研发验证记录](verification/results.md) 与 [测试计划](verification/test-plan.md) 保留研发执行8及以前原始快照，其中“待宿主/待独立评审”是历史状态；当前受控A/B评审以QA10为准，C仅安全前置、非可执行供应商预算计划。[tests5退回](verification/host-return-exec6.md)、[tests7退回](verification/host-return-exec8.md)、完整失败回执和旧候选摘要保留，不重写成通过。

## 风险与后续责任

- 完整门禁只余约21.5秒，环境变慢仍可能超时；不得删断言或延长时限求绿，源码/测试/锁/契约变化需重验。
- 原#33两屏非零RPM/TPM、双浏览器竞争、自然供应商窗口、真实普通/SSE成功、恢复、Token/费用/实扣均Not Run/Blocked；已有三宽受控Go UI不替代这些范围。
- 专用供应商Pipeline先核对合法配置、精确模型、收费安全上界、思考策略和H/X历史账，再独立评审逐用例预算。原24次跨attempt、每次64 tokens、90秒不变；新增真实生成0，未知不得归零或扩额。
- 无生产接口/schema/价格改动，无新迁移或运营扩容；不重做需求/设计，不接管#30/#32、不发布Dashboard，不新增组织级运营文档。Git/PR/Actions/部署交Connector，本节点未执行。

## 最近更新

- 2026-10-09（report执行11）：依据tests9/QA10原位收敛公共状态并完成交付审查及PR正文；内部M01 Accepted、T001 Done，版本发布待Connector，原#33 Incomplete-No-Go保留。无需要用户新增决定的事项。

## 交付审查

按挂载Skill的公共交付契约和审查清单，结论 **Ready for Dispatch**，仅指向publish Connector：任务和里程碑保留唯一状态，当前AC由QA10矩阵追踪，未创建重复事实源或未执行阶段文档；原范围与外部阻塞不因内部Go消失。report执行11实际通过 `git diff --check`、重算154文件源码摘要、核验交接ZIP及61文件、核对runner字节/摘要和42个文档链接目标。新增材料完整性链接对应文件亦已存在。PR链接按实际origin/任务分支与仓库文件核对；没有联网确认远端访问，推送后由Connector确认并关联真实提交SHA。完整门禁仍引用tests9，历史QA/失败证据不重写。
