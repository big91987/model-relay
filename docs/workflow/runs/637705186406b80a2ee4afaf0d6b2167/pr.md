## 交付结论

关联 #34（承接产品测试 #33 报告整改；**不自动关闭 #33，不声明 #34 全部真实验收完成**）。本PR交付 **M01内部安全/计量回归资产与有界测试调度**，独立QA10结论为该范围Go。确认生产实现缺陷仍为0，不称“修好了RPM/TPM算法”。原#33 **Incomplete-No-Go** 和真实供应商验收 **Blocked** 保留。

## 变更

- 新增八组Go受控安全/计量回归：正式bounds到普通/SSE有效cap、RPM/TPM非零阈值与窗口、known/unknown/失败/取消/超时、并发原子性、Key/tenant共享隔离及轮换恢复守恒。fixture仅用于隔离测试，不是产品默认供应商。
- 补充同ID、上游dispatch、双scope全维度、精确金额、窗口来源的白名单采证，自动接入既有 `make verify`。
- 隔离独占fixture使用既有parallel=4；执行8为18个既有计量相关测试补调度。冷却展示补只读API/DOM诊断及测试，不删除原断言、不改720秒deadline。
- 同步Trellis规范及本Run公共交付状态/验收证据，保留原材料、首失败、tests5/7失败和旧版本。
- 无生产代码/协议/schema/模块锁/价格/供应商选择/Makefile变更，无迁移/生产配置调整；不接管#30/#32，不改冻结设计，不混入平台功能或Dashboard。

## 验证与版本

基线HEAD：`3f9cee1615e87e002a2c2300f2d29778072ec3f5`。受测对象是其上的**未提交154文件工作树**，SHA256 `2dfe14a3c305c13621f23b7dd287a8313b7e4f39ab64b8281d283450d125d8ed`，不是新commit/merge/部署版本。report执行11重算同摘要，仅更新交付文档。Connector应以实际提交/推送SHA关联本候选及正式Actions结果；本文不预造发布成功。

- tests9宿主完整 `make verify`：Connector/state/gate均exit0；隔离冷安装，无手工预装，源码前后一致。Go1.25.6、Node26.0.0、Playwright1.59.1；698.502秒，原720秒不变。
- 原全部HTTP/race/真实进程/1280、1440、390受控Go UI通过；新ActualHTTP计量7组通过。非零数值是API/契约/对账证据，不冒充原数值UI。
- QA10独立8组InProcess race exit0/29.082秒及3项Node断言通过；61材料及ZIP、171交付证据、42历史索引摘要一致。计划A/B独立评审合格；C仅安全前置，尚不可执行。
- 完整runner日志499499字节，SHA256 `c93e9625fdd4cd5c51ebf97c7838cd37c887972ccf672e67182d6b5ce4484021`；与Connector日志499596字节是不同归档对象。report重核runner摘要一致。
- 11处 `FAILED:` 为runner预期隔离故障输出，不是顶层失败；外层worker明确Skip/Not Run，父测试真实子进程marker提供恢复Pass，不将Skip改绿。沙箱监听限制单列，report未重跑宿主全门禁或提权。

受控数值：RPM2第1/2受理第3拒绝；TPM B200/U7下206拒绝207接受，P200后406/407、跨窗399/400；known3+4=7、租户fee7与cost33独立，unknown保留预留200而非免费；8竞争者只有1受理7拒绝，occupied207/C1、最终14/C0。**不是DeepSeek实际token或账单。**

## 未验项与后续

- DS-SAFETY-01“后台无默认cap”推断内部未复现：正式bounds64到受控普通/SSE默认64/显式32有动态证据；真实收费输入/输出及隐藏推理、思考策略的完整安全路径仍待核验。
- 原#33两屏非零RPM/TPM、双浏览器同步竞争、自然供应商窗口、真实普通/SSE成功、恢复及Token/费用/实扣 **Not Run/Blocked**。三宽Go UI不替代该范围，自然30秒routing恢复不是RPM窗口证明。
- 专用供应商Pipeline先取得合法配置/凭据、精确模型、收费安全依据、思考策略及跨attempt权威H/X历史账，再独立评审逐用例预算。不得换供应商、静默重命名 `deepseek-v4-flash` 或扩大授权。
- 原整个Run24次、每次64 tokens、90秒不变；新增真实生成0，未知历史不得清零，剩余额度未证不得新生成。失败/取消不能故意无界造付费请求求绿。
- 门禁余量约21.5秒，环境变慢仍可能超时；源码/测试/锁/契约变化需重验。发布、正式候选/Actions、合并/部署由Connector负责，尚未发生。无新生产状态变更，不需新增数据回滚；既有部署恢复规范保持，不称已演练部署/回滚。

## 可审阅产物

以下明确指向仓库 `big91987/model-relay` 的任务分支 `workflow/637705186406b80a2ee4afaf0d6b2167`，分支斜杠已URL编码。**本文生成时尚未提交/推送；链接是Connector推送本候选及文档后供PR访问的目标，不宣称当前已远端可访问，不链接旧main/基线冒充新增产物。** Connector须在发布正文前确认文件已推送且分支匹配。

- [公共总览](https://github.com/big91987/model-relay/blob/workflow%2F637705186406b80a2ee4afaf0d6b2167/docs/workflow/runs/637705186406b80a2ee4afaf0d6b2167/README.md)
- [任务T001](https://github.com/big91987/model-relay/blob/workflow%2F637705186406b80a2ee4afaf0d6b2167/docs/workflow/runs/637705186406b80a2ee4afaf0d6b2167/tasks/T001-metering.md) · [里程碑M01](https://github.com/big91987/model-relay/blob/workflow%2F637705186406b80a2ee4afaf0d6b2167/docs/workflow/runs/637705186406b80a2ee4afaf0d6b2167/milestones/M01-metering.md)
- [QA结论](https://github.com/big91987/model-relay/blob/workflow%2F637705186406b80a2ee4afaf0d6b2167/docs/workflow/runs/637705186406b80a2ee4afaf0d6b2167/acceptance/release-assessment-qa-10.md) · [逐项矩阵](https://github.com/big91987/model-relay/blob/workflow%2F637705186406b80a2ee4afaf0d6b2167/docs/workflow/runs/637705186406b80a2ee4afaf0d6b2167/acceptance/acceptance-matrix-qa-10.md)
- [独立计划评审](https://github.com/big91987/model-relay/blob/workflow%2F637705186406b80a2ee4afaf0d6b2167/docs/workflow/runs/637705186406b80a2ee4afaf0d6b2167/acceptance/test-plan-review-qa-10.md) · [Go受控网页旅程](https://github.com/big91987/model-relay/blob/workflow%2F637705186406b80a2ee4afaf0d6b2167/docs/workflow/runs/637705186406b80a2ee4afaf0d6b2167/acceptance/core-journeys-qa-10.md)
- [证据索引及历史失败](https://github.com/big91987/model-relay/blob/workflow%2F637705186406b80a2ee4afaf0d6b2167/docs/workflow/runs/637705186406b80a2ee4afaf0d6b2167/acceptance/evidence/index.md) · [门禁state](https://github.com/big91987/model-relay/blob/workflow%2F637705186406b80a2ee4afaf0d6b2167/docs/workflow/runs/637705186406b80a2ee4afaf0d6b2167/acceptance/evidence/qa-10/host-tests9/state.json) · [完整runner日志](https://github.com/big91987/model-relay/blob/workflow%2F637705186406b80a2ee4afaf0d6b2167/docs/workflow/runs/637705186406b80a2ee4afaf0d6b2167/acceptance/evidence/qa-10/host-tests9/verify.log)
- [候选与未验责任](https://github.com/big91987/model-relay/blob/workflow%2F637705186406b80a2ee4afaf0d6b2167/docs/workflow/runs/637705186406b80a2ee4afaf0d6b2167/acceptance/findings-and-next-actions-qa-10.md)

材料版本 `product-test33-handoff-r1`，交接ZIP SHA256 `a13fb6b9089cefbb9d2b6f9a333830e0ee6866bd7ff45d5ba881da97e9a66a49`；原设计ZIP声明SHA256 `5f51d35fdc15ed7214749f9413879b4238097f1d6b6bbe2cc0d508d8e103cef7`。核验的是交接包/61文件，不冒称重建原设计ZIP。输入材料保持忽略不随PR提交；公共材料引用及逐文件核验在总览/QA证据中。
