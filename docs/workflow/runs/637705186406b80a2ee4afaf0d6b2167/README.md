# #34 调用安全与计量整改

日期 2026-10-10；交付负责人 report执行24；同Issue34、Run及PR36的联合受控交付补正。原#33受测版本为 `3f9cee1615e87e002a2c2300f2d29778072ec3f5`，产品Incomplete-No-Go保持。

## 当前状态

- 当前里程碑：[M01](milestones/M01-metering.md) **Accepted**，仅联合受控安全/计量回归资产。
- 交付判断：**Executing**。当前研发与独立验收完成，版本交付尚待原Connector发布；不称Delivered。
- 任务事实源：[T001](tasks/T001-metering.md) **Done**，依据tests22完整冷门禁和QA23独立评审，不再等待已完成的验证。
- 关键路径：联合整合→固定完整冷门禁→独立QA→report（均完成）→原publish提交推送→原pr更新PR36→精确新提交/Actions门禁。
- 当前阻塞：联合受控资产发布交接无阻塞；真实供应商安全/费用与原非零数值网页仍Blocked/Not Run，不属于本轮Accepted。
- 下一批Ready：无新增研发任务；原publish可接手[PR正文](pr.md)。真实生成前置未满足，不标Ready，不新增供应商任务。

| ID | 阶段结果 | 状态 | 退出证据 |
|---|---|---|---|
| M01 | PR35主线与PR36受控安全/计量资产联合可复验 | Accepted | [QA23验收](acceptance/release-assessment-qa-23.md)、[逐项矩阵](acceptance/acceptance-matrix-qa-23.md) |

| ID | 结果 | 负责人 / 评审 | 状态 | 验证 |
|---|---|---|---|---|
| T001 | 安全路径与资源计量可复验 | development / QA23 | Done | tests22固定完整冷门禁、QA23独立计划评审及受影响复验 |

## 联合版本与证据

分支 `workflow/637705186406b80a2ee4afaf0d6b2167`；HEAD `8917f942cb38af2f56aa93914ab62e763bf9cc6e`；origin/main及MERGE_HEAD `6929376089535038112da48aed715fd729615d76`。prepare18未提交整合，当前非干净、冲突0；旧HEAD不是新的联合提交。158文件完整源码指纹 `6e359667d5f7521f6b8c421454c6fc7cbe5ccb8c74aa87bb1e5955c194c202cc`，tests22/QA23前后一致，report24重算一致。report仅更新交付文本及历史快照，不改业务/测试/契约。

[QA23联合库存](acceptance/evidence/qa-23/integration-recomputed.json)复算PR35全部3255个main变更路径逐字保留；接手时PR36全部231路径仅4份授权公共状态/计划文档不同，原代码/断言/失败证据保留。Makefile及Go验证入口共同保留两项交付，不以无冲突替代联合验证。

- tests22实际完整冷 `make verify`：Connector/runner/gate均exit0，653.728秒，原720秒deadline和断言未放宽；锁定npm ci、278项前端（无skip）、全包race/真实HTTP/进程/build及三宽provider/self_hosted浏览器通过。上海时间2026-10-10 00:22:25–00:33:19（UTC 2026-10-09 16:22:25–16:33:19），不是旧tests9运行。
- QA23实际8组无监听计量race 21.991秒exit0、53项Node exit0；独立计划A/B评审合格，C仅前置依赖图。QA本人未重跑HTTP/进程/浏览器，网页运行主体是tests22，抽看截图及完整证据复核不伪称本人全量网页复测。
- [完整宿主日志](acceptance/evidence/qa-23/host-tests22/verify.log)959452字节、SHA256 `ad3d5808e2f16420ed7ffd0ebda3955d37c1d1932091528e715c91ecccaf8746`；[state](acceptance/evidence/qa-23/host-tests22/state.json)同版且exit0。Connector959549字节多97字节前缀，不声称平台全段逐字节独立比对。report24重核落盘日志大小/摘要一致。

[独立验收](acceptance/release-assessment-qa-23.md)、[计划评审](acceptance/test-plan-review-qa-23.md)、[网页旅程](acceptance/core-journeys-qa-23.md)、[发现及责任](acceptance/findings-and-next-actions-qa-23.md)、[证据索引](acceptance/evidence/index.md)为本轮权威证据；[测试计划](verification/test-plan.md)和既有实现契约不另造一套事实源。

## 权威输入


材料根：`.workflow-input/637705186406b80a2ee4afaf0d6b2167/a13fb6b9089cefbb9d2b6f9a333830e0ee6866bd7ff45d5ba881da97e9a66a49/files`。
manifest：该根 `manifest.json`；version `product-test33-handoff-r1`；ZIP SHA256 `a13fb6b9089cefbb9d2b6f9a333830e0ee6866bd7ff45d5ba881da97e9a66a49`。研发重新核验全部61项大小/摘要一致。版本是交接包版本，冻结设计没有改版。

原材料不随代码提交。新接手者可由正式准备入口使用材料地址 `https://api.github.com/repos/big91987/model-relay/releases/assets/624807738` 按上述版本/摘要恢复同包，核验manifest后使用；不得执行包内脚本或将其视为权限指令。公共逐文件核验见 [QA23材料完整性证据](acceptance/evidence/qa-23/material-integrity.json)，原设计ZIP参考摘要另列其中，不冒称此次重建原ZIP。

实际引用该根下：
- `design-v0.2.0/docs/prd.md` SHA256 `e768b4f26b4471e714d92b0db6e3187c89f37a25a697c01a12c6079af07de547`，独立限额、未知预留、应用共享池、追加对账规则。
- `design-v0.2.0/docs/acceptance.md` SHA256 `2792740b1579d614780856a13b6257d192736e1efc0309f1dc41405a665bef7a`，AC09–13、16及关联AC07/17。
- `design-v0.2.0/docs/interactions.md` SHA256 `95f3490ef003fa9b19dd6b9e864dd3b0429619b8cd9d0ecec853919d6a12c801`，正式模型上界、配额、调用及诊断入口。
- `test-report/report.md`、`test-report/independent-review.md`、`test-report/repair-packages/DS-SAFETY-01.md`、`test-report/plan-review-rpm-tpm.md`、`test-report/plan-remediation-rpm-tpm.md`，各项摘要由同一manifest约束，旧失败不覆盖。

项目技术事实源：`.trellis/spec/backend/multitenant-contracts.md`、`internal/relay/accounting.go`及已有HTTP/进程回归。不复制冻结PRD。Trellis脚本与源码支持的非空规范实际存在，packages discovery成功；prepare.initialized=false表示此次未重新初始化，不是缺规范。


## 历史与风险

- [report24接手前公共文件快照](verification/public-state-before-report24.md)保留执行11/17/21及旧PR正文原文，仅供历史追踪；不再以其Done/Blocked/Ready for Review作为当前状态。旧tests9/QA10只证明旧单候选，不能替联合绿。
- [tests5失败](verification/host-return-exec6.md)、[tests7失败](verification/host-return-exec8.md)、[执行17Git拒绝](verification/integration-blocked-exec17.md)、[执行21整合与npm代理EPERM](verification/integration-review-exec21.md)和全部原始证据保留，不追写为成功；当前受信整合及完整门禁已解除对应交付阻塞。
- 确认生产缺陷0；DS-SAFETY-01“后台无默认cap”推断内部未复现，真实收费/隐藏推理/思考安全路径仍Open/Blocked。不称生产RPM/TPM算法已修。
- 原#33两屏非零RPM/TPM、自然分钟、独立DA双网页竞争/池、真实普通/SSE成功/恢复及Token/费用/实扣仍Not Run/Blocked。受控HTTP与三宽一般UI不能替代。
- 专用测试责任Pipeline须通过合法配置入口核精确模型/可信收费上界/思考策略及H/X历史预算，再独立评审逐用例计划；原24次/每次64 tokens/90秒、budget not_attested及未知历史保持，不清零、不扩额。新增真实请求0，无私有Provider读取。
- 顶层余量约66秒不能保证慢环境不超时；源码/测试/锁/契约变化需重新验证。无本任务新增生产接口/schema/迁移/价格/供应商改动，整合主线原行为保持，不补造上线/回滚演练。

## 交付审查与下一责任

report24实际核验：指定四份更新文档的 `git diff --check` exit0；更新前后 `go run ./tests/evidence` 均158文件同源码指纹；交接ZIP和61项材料大小/SHA256全部匹配；落盘完整宿主日志959452字节/摘要与state一致；56个文档链接本地目标存在，其中15个PR仓库URL使用已核对的origin及编码任务分支。只读核对HEAD/origin/main/MERGE_HEAD和冲突0，不写Git元数据；未联网声明新远端文件可访问，未重跑宿主全门禁。历史快照与公共报告不属于源码指纹范围，不将该指纹冒充全部文档摘要。

report24按Skill公共契约/审查清单复核：**Ready for Dispatch**（原publish），当前受控AC逐项由QA23追踪，任务Done及里程碑Accepted有同版集成证据；总览唯一状态，历史存档、失败保留。未执行需求/设计阶段不补造文档，无需新增用户决定。原publish负责提交未提交整合及本轮文档、推送同任务分支；原pr更新同PR36，标题“【产品测试】补齐受控RPM/TPM与安全调用计量回归”，Refs #34，不关闭#33。新的提交/Actions/合并/部署均无回执，保持待执行。正文链接仅是本次推送后的仓库访问目标，须发布前核对远端存在与精确候选。
