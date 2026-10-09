# 【产品测试】补齐受控RPM/TPM与安全调用计量回归

Refs #34

## 交付结论

同Issue34、Run及PR36的联合候选补正：整合已合入PR35的main `6929376089535038112da48aed715fd729615d76`，保留PR36受控计量/安全回归与全部原断言。tests22完整冷门禁及独立QA23结论 **Go，仅限M01联合受控回归资产**；T001 Done、M01 Accepted。**确认生产缺陷0，不宣称修复生产RPM/TPM算法故障；原#33 Incomplete-No-Go保持，不关闭#33，不声明#34所有真实验收完成。**

## 变更与复现结论

- 八组Go受控回归覆盖正式bounds到普通/SSE有效cap、非零RPM/TPM阈值/窗口、known/unknown/失败/取消/超时、并发原子性、Key/tenant共享隔离及轮换/恢复守恒。受控fixture不是产品默认供应商。
- 同ID、上游dispatch、双scope精确金额与池副作用采证接入既有make verify；18个既有相关测试保留parallel=4有界调度，冷却补只读诊断。未删断言或延长原720秒deadline。
- DS-SAFETY-01“后台无默认cap”推断内部未复现：正式bounds64到实际受控普通/SSE默认64、显式32有证据；真实收费/隐藏推理/思考安全路径仍待核验，候选不伪关闭。
- 联合保留PR35首次自托管/可读性及主线所有行为，本任务增量是回归资产/采证/文档，不是另行产品修复；不改冻结设计、生产算法、协议/schema、价格或供应商名称，不接管#30/#32或混入平台/Dashboard。
- 旧tests9/QA10只证明旧单候选。tests5/7、exec17 Git拒绝、exec21代理EPERM及首失败原样保留，不能以旧绿发布联合版本，也不将历史环境失败伪称算法首红。

## 联合验证与版本

任务分支 `workflow/637705186406b80a2ee4afaf0d6b2167`；旧已推送HEAD `8917f942cb38af2f56aa93914ab62e763bf9cc6e`，origin/main及MERGE_HEAD `6929376089535038112da48aed715fd729615d76`。**当前受测对象是尚未提交的联合工作树，不是旧HEAD对应制品，也尚无新联合提交/推送回执。** 158文件完整源码SHA256 `6e359667d5f7521f6b8c421454c6fc7cbe5ccb8c74aa87bb1e5955c194c202cc`，tests22/QA23前后一致，report24重算一致。联合库存核对PR35全部3255个main变更路径及PR36全部231路径；接手时仅4份授权交付状态/计划文档不同，源码/原断言保留。

- 固定tests22实际完整冷make verify：Connector/runner/gate均exit0，653.728秒；上海2026-10-10 00:22:25–00:33:19（UTC 2026-10-09），原720秒不变。node_modules/npm cache起初不存在，NODE_PATH unset，无手工预装。
- Go1.25.6、Node26.0.0、npm11.12.1、Playwright/core1.59.1、Chromium147.0.7727.15；格式/modules/vet、锁定npm ci、278项前端0 fail/skip、全包race、实际HTTP/进程/build及migration/legacy/routing/console/full-console均运行通过。
- 1280/1440/390三宽provider/self_hosted真实Go网页、J01–J07、首次空安装各12步骤通过；普通/SSE同ID均input3/output4、租户费21微美元/成本7微美元。该价与计量fixture租户费7/成本33不同，不混算，不是DeepSeek账单。
- QA23实际8组无监听race 21.991秒exit0、53项Node exit0，独立计划A/B合格，C仅合法前置依赖图。关键数值与45PNG integrity复核一致；QA本人抽看截图但未重跑浏览器，动态运行主体是tests22。
- 完整宿主日志959452字节、SHA256 `ad3d5808e2f16420ed7ffd0ebda3955d37c1d1932091528e715c91ecccaf8746`；Connector959549字节含97字节前缀。未声称独立逐字节比对平台整段输出，落盘完整日志/回执/同版具体用例支持结论。嵌套runner预期故障不当顶层失败，外层worker Skip不改Pass。

关键受控数值：RPM2第1/2受理、第3拒绝；TPM B200/U7下206拒绝207接受，P200后406/407、跨窗399/400；known3+4=7，unknown保留200而非免费；8竞争者1受理7拒绝，occupied207/C1→最终14/C0；双scope无部分扣计。窗口注入与API数值不冒充自然时间网页。

## 未验边界与后续

- 原#33两屏非零RPM/TPM、自然分钟、独立DA双网页竞争/池、真实普通/SSE成功/恢复及Token/费用/实扣 **Not Run/Blocked**。一般三宽Go UI、routing30秒冷却恢复、受控HTTP不替代这些范围。
- 真实DeepSeek完整安全/收费输入输出及隐藏推理/思考策略、精确flash/pro可用性未核验。无私有Provider配置读取、无新供应商请求，不换供应商或静默重命名deepseek-v4-flash。
- 原跨attempt 24次/每次64 tokens/90秒、budget not_attested和未知历史保持，不重置或扩额。专用合法配置测试Pipeline须核H/X、可信安全B与逐用例潜在attempt（H+X+潜在attempt≤24），独立评审后才可复测；未知/失败/取消不自动返还预算。
- 当前完整门禁约66秒余量不能保证慢环境永不超时；源码/测试/锁/契约若改变需重验，不能弱化断言或增时求绿。
- 原publish/pr负责提交未提交整合、推送及更新同PR36、映射精确新提交与正式Actions。本文整理时这些联合发布步骤、合并/部署均尚未发生；report未启动部署或重跑宿主全门禁。无本任务新增生产状态变更/迁移，不冒称已演练上线/回滚。

## 可审阅证据

以下核对实际origin为big91987/model-relay、同任务分支及本地文件；分支斜杠URL编码。**当前远端仍为旧HEAD，新QA/文档及联合候选未提交推送；这些是原Connector推送本次产物后的访问目标，不宣称现在已有远端内容。发布正文前必须确认本候选与所有链接目标已推送，保留精确新SHA与Actions映射；不要链接旧main/旧SHA冒充新增证据。**

- [公共总览](https://github.com/big91987/model-relay/blob/workflow%2F637705186406b80a2ee4afaf0d6b2167/docs/workflow/runs/637705186406b80a2ee4afaf0d6b2167/README.md)
- [任务T001](https://github.com/big91987/model-relay/blob/workflow%2F637705186406b80a2ee4afaf0d6b2167/docs/workflow/runs/637705186406b80a2ee4afaf0d6b2167/tasks/T001-metering.md)
- [里程碑M01](https://github.com/big91987/model-relay/blob/workflow%2F637705186406b80a2ee4afaf0d6b2167/docs/workflow/runs/637705186406b80a2ee4afaf0d6b2167/milestones/M01-metering.md)
- [QA23独立验收](https://github.com/big91987/model-relay/blob/workflow%2F637705186406b80a2ee4afaf0d6b2167/docs/workflow/runs/637705186406b80a2ee4afaf0d6b2167/acceptance/release-assessment-qa-23.md)
- [逐项分层矩阵](https://github.com/big91987/model-relay/blob/workflow%2F637705186406b80a2ee4afaf0d6b2167/docs/workflow/runs/637705186406b80a2ee4afaf0d6b2167/acceptance/acceptance-matrix-qa-23.md)
- [独立计划评审](https://github.com/big91987/model-relay/blob/workflow%2F637705186406b80a2ee4afaf0d6b2167/docs/workflow/runs/637705186406b80a2ee4afaf0d6b2167/acceptance/test-plan-review-qa-23.md)
- [三宽实际网页旅程](https://github.com/big91987/model-relay/blob/workflow%2F637705186406b80a2ee4afaf0d6b2167/docs/workflow/runs/637705186406b80a2ee4afaf0d6b2167/acceptance/core-journeys-qa-23.md)
- [发现与下一责任](https://github.com/big91987/model-relay/blob/workflow%2F637705186406b80a2ee4afaf0d6b2167/docs/workflow/runs/637705186406b80a2ee4afaf0d6b2167/acceptance/findings-and-next-actions-qa-23.md)
- [证据总索引](https://github.com/big91987/model-relay/blob/workflow%2F637705186406b80a2ee4afaf0d6b2167/docs/workflow/runs/637705186406b80a2ee4afaf0d6b2167/acceptance/evidence/index.md)
- [联合库存复核](https://github.com/big91987/model-relay/blob/workflow%2F637705186406b80a2ee4afaf0d6b2167/docs/workflow/runs/637705186406b80a2ee4afaf0d6b2167/acceptance/evidence/qa-23/integration-recomputed.json)
- [完整门禁state](https://github.com/big91987/model-relay/blob/workflow%2F637705186406b80a2ee4afaf0d6b2167/docs/workflow/runs/637705186406b80a2ee4afaf0d6b2167/acceptance/evidence/qa-23/host-tests22/state.json)
- [完整宿主日志](https://github.com/big91987/model-relay/blob/workflow%2F637705186406b80a2ee4afaf0d6b2167/docs/workflow/runs/637705186406b80a2ee4afaf0d6b2167/acceptance/evidence/qa-23/host-tests22/verify.log)
- [归档摘要](https://github.com/big91987/model-relay/blob/workflow%2F637705186406b80a2ee4afaf0d6b2167/docs/workflow/runs/637705186406b80a2ee4afaf0d6b2167/acceptance/evidence/qa-23/archive-index.json)
- [测试计划](https://github.com/big91987/model-relay/blob/workflow%2F637705186406b80a2ee4afaf0d6b2167/docs/workflow/runs/637705186406b80a2ee4afaf0d6b2167/verification/test-plan.md)
- [历史公共文件快照](https://github.com/big91987/model-relay/blob/workflow%2F637705186406b80a2ee4afaf0d6b2167/docs/workflow/runs/637705186406b80a2ee4afaf0d6b2167/verification/public-state-before-report24.md)

材料使用product-test33-handoff-r1，ZIP SHA256 `a13fb6b9089cefbb9d2b6f9a333830e0ee6866bd7ff45d5ba881da97e9a66a49`，61项匹配，冻结design-v0.2.0逐字保持；包版本不是设计改版。材料恢复与实际引用路径见公共总览及QA23材料完整性记录。
