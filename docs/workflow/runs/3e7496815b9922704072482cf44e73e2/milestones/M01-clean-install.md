# M01 干净安装可通过完整验证

- 状态：**Ready for Review**。
- 负责人：交付负责人；实现development、验证tests/独立QA、发布部署由授权协调者/Connector负责。
- 任务：[T001](../tasks/T001-clean-install.md)；硬依赖：宿主工具链与合法发布部署入口。
- 覆盖：Issue28依赖顺序、冷起点/完整证据、既有budget-clock-r1/UTC/迁移/UI不变量与正式preview闭环。

## 阶段结果与进入条件
全新安装无需手工npm预装即可使用项目固定make verify完成全部验证；正式新main部署具有可核对的版本/schema3/真实admin证据。Issue28已授权最短Bug路径，main基线、design-v0.2.0及架构边界明确；不重做需求/设计。

## 退出条件
- [x] 原缺包红灯有可定位证据，消费者依赖先准备，失败阻止消费。
- [x] 当前147件f8befd候选正式宿主固定make verify冷全门禁exit0；前后源不变、完整日志/state、最终本地锁版本/Chromium探针与自有清理闭合。
- [x] QA21独立核对同版HTTP/零软预算/UTC/迁移/UI及错误覆盖；结论Go with known issues，QA原生监听/动态浏览器Blocked与供应商Not Run分别保留。
- [ ] T001发布闭环完成：新PR/精确merge、新main正式prepare、Owner deploy=true target=preview成功、Deployment/environment_url、healthz version/schema3、真实admin及旧数据/密码保持。

冷验证和独立QA已完成，不再以QA6旧证据缺口将当前门槛标Blocked；因发布/部署未完成且T001仍In Review，本里程碑不标Accepted/Delivered。

## 集成证据
[QA报告](../acceptance/report-qa-21.md)、[逐项矩阵](../acceptance/matrix-qa-21.md)、[实际旅程](../acceptance/journeys-qa-21.md)、[148件原样归档索引](../acceptance/evidence/qa-21/archive-index.json)。tests20 runner310718bytes/SHA256 `7c4d3952943721deae9188a6f52f860572037ac8ea39ba55d278d6bf2c7abb7b` 与state一致，Connector310815bytes/未截断/EOF证据单独记录。报告22复核见 [一致性审查](../verification/report-22-review.json)，不是重新运行完整门禁。

## 风险与继续条件
[summary历史评估](../acceptance/summary-assessment-qa-21.md) 保持Open/根因未知/P2；development19只加诊断，20一次绿不证明根治，当前业务P1未实证。正式prepare仍必须跑原入口；若再红带具体安全JSON与完整log/state返研发，不放弱断言、增限或猜测修复。

下游按 [发布交接](../acceptance/release-handoff-qa-21.md) 继续已有授权；未部署之前不能把旧9851177/schema2页面称为新版。不得reset数据/密码、手改数据库或启动替代服务。控制器日志覆盖另由Platform维护源负责。历次原失败与中断记录在 [验证索引](../verification/index.md) 保留，不被新证据改写。
