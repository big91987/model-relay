# Issue 63：实施与证据登记

## 当前实际状态

2026-10-25：development执行2接续prepare冻结材料；本地HEAD为f06f05de20ffac5b8c8e772958c808379001f0e3，初始git status为空；harness-output为空，未找到本Run已有计划或已评审设计。此次只形成设计检查候选，不修改产品、不运行实验/固定门禁、不创建PR。

已读取冻结执行输入的任务文本和parameters（issue_number=63、repository=big91987/model-relay、repository_commit同HEAD），未读取凭据环境或密钥。节点编号“执行2”不代表本Run已完成一次设计检查。输入归档为frozen-input.md；与本轮用户任务一致。

prepare用户回执：执行92450ms，exit0、日志3636bytes/truncated=false；基线准备时间不能从总等待中隐去。本节点不把该准备通过视为功能通过。

## 文件登记

- docs/workflow/runs/ba434da8039cca16f6829481112e3d68/frozen-input.md：冻结用户任务原文及安全parameters，不含凭据。
- docs/workflow/runs/ba434da8039cca16f6829481112e3d68/plan.md：本轮需求、设计、参数、存储/兼容与AC矩阵。
- docs/workflow/runs/ba434da8039cca16f6829481112e3d68/experiment.md：观测前冻结的样本、控制、阈值与验收。
- docs/workflow/runs/ba434da8039cca16f6829481112e3d68/execution.md：实际状态、后续归档和计时口径。
- harness-output/issue63-plan-manifest.json：以上四文件原相对路径、bytes、SHA256及源码基线、归档生成UTC。

工具限制：当前执行器没有独立apply_patch工具或同名可执行文件。仅新增规划文档使用shell内apply_patch函数（标准patch封装）应用统一diff；没有改产品源码或安装工具。

## 后续证据位置（仅预定，不冒称已存在）

后续研发原始回执/JSONL/截图/trace/退出码/统计在harness-output下按candidate与轮次登记；可读计划/验收汇总/对照SVG在本Run docs目录，QA报告固定harness-output/qa.md。handoff只逐项列真实文件，不能用本段目录作为artifacts。

每候选生成独立source-manifest.json：原路径、bytes、SHA256，绑定源码patch.txt（不是可编译.go副本）、定向回执、固定完整verify.log/receipt、QA实际回执。证据目录不置入Go包源码路径；不删失败原件、不改正常测试排除规则。日志超8MiB按字节分片并保留次序、总SHA、原退出码和EOF标记；单次最多256文件/128MiB，按真实清单安排，不把截断输出当完整通过。

## 计时和交付约束

Issue创建时刻、入口排队与既有平台阶段事件尚未由可用冻结材料给出，记Unknown，不按本地首次命令时间伪造。后续记录设计等待开始/正式接续、实现、定向、tests、独立QA、返工、approval、report、publisher及释放后留存核验的实际UTC起止与证据。阶段驻留含等待/归档，不称Agent净耗时。原生token/缓存仅引用平台原值，当前不可得；费用Unknown不估算。Issue60仅设施开销参考，不算不同任务提效百分比。

下一动作：正式wait_for_input(kind=clarification)请求一次整体设计检查；未收到正式User Input前不改产品、不交tests。获检查接续后同工作区实现和定向验证，handoff(target=tests)交完整真实候选。后续由独立QA→approval→report→固定publisher创建draft PR。研发不读凭据、不git push、不建PR、不部署、不改预览、不恢复旧Run。
