# 密码可见性研发交付计划

## 目标与边界
管理员能够核对密码且正常登录、退出与恢复会话。权威需求为 [product-definition.md](product-definition.md)，架构为 [password-visibility-architecture.md](password-visibility-architecture.md)，G2 已就绪。仅改原生登录 DOM、局部样式与现有回归测试；不改认证、数据库、健康／版本协议，不部署或调用真实供应商。

## 当前状态
- 当前权威 QA：[独立验收报告](acceptance/acceptance-report.md)、[矩阵](acceptance/acceptance-matrix.md)、[快照与最终保护](acceptance/evidence/final-integrity.log)。2026-10-06 同轮宿主固定 make verify 退出 0，QA 本地受监听限制退出 2；独立结论 Go with known issues（既有 P3 日志副作用，本轮已恢复），交报告/PR，不等于已发布或合并。以下研发状态为阶段记录，不能替代当前 QA。
- 当前里程碑：[M01](milestones/M01-password-visibility.md)，Ready for Review；交付判断 Verifying，不是 Delivered。
- 关键路径：源码规范 bootstrap → T001 实现与本地验证 → 宿主固定验证 → 独立 QA → 平台 PR。
- 无产品／设计阻塞。make verify 退出 2，沙箱禁止 HTTP 监听；待固定宿主 Connector，不扩大权限。
- 当前任务：[T001](tasks/T001-password-visibility.md)，In Review，研发节点负责；独立 QA 评审。没有并行任务或软依赖，不启动第二套调度器。
- 下一批 Ready：完成研发后由平台 tests 执行固定门禁；QA 独立判断。

## 交付审查
2026-10-06 按 delivery-review-checklist 审查：Ready for Dispatch。小增量裁剪为一个端到端里程碑与一个任务包，AC01—AC13 全部追踪到任务内验证矩阵；测试 Harness 已存在，无额外建设任务。没有日历排期、组织运营文档或重复事实源。任务状态以任务文件为准；本页仅投影，不声称版本交付。

## 最近更新
2026-10-06 实现及非监听回归完成，[本轮验证](verification/M01-development.md) 包含失败、代码标识和历史保护。具备进入宿主测试与独立评审的交接条件，不具备最终交付条件；当前无需要用户决定的事项。旧健康重建脚本的历史日志写入副作用已恢复并记录，下游再次运行须保留旧基线。
