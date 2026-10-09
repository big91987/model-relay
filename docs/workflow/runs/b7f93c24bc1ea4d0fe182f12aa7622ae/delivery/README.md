# Issue32 控制台可读性与首次自托管网页交付

## 目标与边界
- Run：b7f93c24bc1ea4d0fe182f12aa7622ae；长期计划阶段2的本次切片，不重复Issue30/PR31或同步修复。
- 目标：已知状态使用中文业务术语；完整公开技术值在1280/1440/390可访问；全新self_hosted安装完成网页创建、失败纠正、授权、普通/SSE同ID日志费用链。
- 权威需求：Issue32六项AC、[冻结产品定义](../../4cd62ced6a4f98f50e3cca6f7596eb06/product-definition.md)、[原QA58-01](../../4cd62ced6a4f98f50e3cca6f7596eb06/acceptance/visual-review-58.md)；当前追踪见[QA57矩阵](../acceptance/acceptance-matrix-57.md)。
- 权威架构：沿用[产品架构](../../4cd62ced6a4f98f50e3cca6f7596eb06/architecture.md)及[实现契约](../../4cd62ced6a4f98f50e3cca6f7596eb06/implementation-contract.md)。未重建PRD/设计；API、存储、费用、Token、授权、秘密口径保持。
- 冻结材料：design-v0.2.0；archive SHA256 `5f51d35fdc15ed7214749f9413879b4238097f1d6b6bbe2cc0d508d8e103cef7`；manifest为 `.workflow-input/b7f93c24bc1ea4d0fe182f12aa7622ae/5f51d35fdc15ed7214749f9413879b4238097f1d6b6bbe2cc0d508d8e103cef7/files/manifest.json`。实际引用同root的 `docs/prd.md`、`docs/interactions.md`、`docs/acceptance.md`及原型截图。19项路径/长度/SHA见[QA材料审核摘要](../acceptance/evidence/attempt-57/audit-summary.json)；状态保持 `frozen-for-test-implementation`，不改为批准。
- 非目标：真实供应商/支付/外部身份联调、真实预览账号操作、重置业务库、改Platform/部署控制器或组织级运营文档。既有Trellis规范沿用，不声称重新初始化。

## 当前状态
- 当前里程碑：M01 **Accepted**；交付判断：**Verifying**，本任务实现及独立验收已完成，仅剩Connector的Git同源发布核证，尚无新PR/合并/部署回执。
- 任务事实源：[T001](tasks/T001-console.md) **Done**；QA57 **Go with known issues**，仅Issue32受控范围。
- 当前版本：HEAD/基线main `3f9cee1615e87e002a2c2300f2d29778072ec3f5`；分支 `workflow/b7f93c24bc1ea4d0fe182f12aa7622ae`；未提交候选154件源码戳 `dc667be6f0f5db892748ac8a214d799264c2d451004da8c80a5e06e6c38f338c`。旧HEAD不是本次变更已发布版本。
- 关键路径：已完成实现→已完成宿主56固定验证→已完成QA57→report58交付核对→**publish精确同源提交/推送**→后续授权PR/合并/Owner Actions preview核证。
- 当前阻塞：无本任务产品返工或必须用户澄清项；后续发布尚未执行，不把它写成已交付。
- 下一批Ready任务：无新增研发任务包；publish可领取现有候选及[PR正文](../pr.md)，不重新请求阶段批准。

## 里程碑与任务投影
| ID | 结果 | 状态 | 依赖与退出证据 |
|---|---|---|---|
| [M01](milestones/M01-console.md) | 可读控制台与首次自托管网页链路 | Accepted | 冻结基线、原固定门禁、[QA57报告](../acceptance/acceptance-report-57.md)与矩阵 |
| [T001](tasks/T001-console.md) | 展示修复与独立自托管网页回归 | Done | development实现、QA独立评审；[真实旅程](../acceptance/journeys-57.md)、[视觉审查](../acceptance/visual-review-57.md) |

## 风险与后续责任
| 项目 | 当前事实与处理 |
|---|---|
| 历史空图/R28-04 | 全部历史成因仍未明；当前视觉/关闭锁缺口已闭合。保留原失败证据，复发回原路径最小归因，不据新绿追认根治 |
| 原生动态与外部边界 | QA独立原生动态、真实供应商/支付/身份、Owner preview已登录分别Not Run。宿主受控HTTP/UI不能替代，见QA矩阵 |
| 发布同源 | 候选未提交/推送；report只更新文档。publish核154件源码戳及分支，源码改变须重新评估验收 |
| 真实数据/回退 | 本轮未操作真实账号/业务库；console_clock水位可能不同，不宣称整库字节不变。沿用项目[本地部署说明](../../../../local-deployment.md)及既有升级契约，不重置账号或倒写水位 |

## 交付审查（2026-10-09，report58）
- 原固定 `make verify`：宿主56于2026-10-09T11:00:42.841521Z至11:11:53.351221Z一次冷隔离exit0；276Node零失败/零skip、七包Go race/真实HTTP、UTC预算/迁移/build、三视口J01–J07/summary/visual/首次self_hosted全部完成。顶层ClockRecoveryWorker设计skip单列Not Run，同轮恢复父测试子进程证据Pass，未把skip改绿。
- QA57核344件原字节证据、181PNG非单色、134侧车，人工看27当前图与4基线图；不是人工看完181图，也不是QA/report自行重跑浏览器。
- report58在编辑前核QA文档原摘要及344件/40,258,790字节全部匹配，复算154件源码戳与研发55/宿主56/QA57一致；不重跑已知监听受限门禁，不重放不确定写入或供应商调用。
- 审查结论 **Ready for Dispatch**，仅指可交publish核证。上游AC、单一T001/M01、共享入口/失败证据可定位，无范围/架构偏差，跳过不适用需求/设计返工；无用户必答或新增权限问题。
- 本轮只整理总览、任务、里程碑、验收索引及PR正文。QA57的 `final-integrity.json`保留当时文档摘要，不重写历史；公共投影不在154件源码范围，QA报告与原证据不变。
- report58实际复算、原证据完整性、文档摘要与仓库链接检查见[交付核对回执](verification/report-58.json)；远端可访问性留待publish推送后核证，不把本地链接检查写成远端成功。

## 历史与当前证据
- [研发55](verification/development-55.md)、[宿主54原失败](verification/host-tests-54/archive-index.json)、[宿主52失败](verification/host-tests-52/archive-index.json)。
- [QA14原No-Go](../acceptance/acceptance-report-14.md)、[原整改](../acceptance/remediation-14.md)、[原210件索引](../acceptance/evidence/attempt-14/integrity.json)。
- [执行39停止接续审计](verification/development-resume.md)、[执行41接续记录](verification/development-41.md)及 `verification/`下全部development历史报告和host-tests归档原位保留；其当时状态不再重复投影为当前待验。
- 当前退出条件：[QA57报告](../acceptance/acceptance-report-57.md)、[矩阵](../acceptance/acceptance-matrix-57.md)、[344件索引](../acceptance/evidence/attempt-57/integrity.json)、[原UTC/退出清理](../acceptance/evidence/attempt-57/host/state.json)、[完整原日志](../acceptance/evidence/attempt-57/host/verify.log)。
