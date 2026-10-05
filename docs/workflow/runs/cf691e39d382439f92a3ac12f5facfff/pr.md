# 修复验证回执覆盖历史文件并闭环交付状态

Closes #14

## 变更与边界

- `tests/health-history/main.go` 移除对已跟踪历史日志的无条件写入，仅向 stdout 输出本次回执；检查写入错误，stderr 给出诊断并非零退出，不吞掉分类失败。
- 增加真实 main 子进程的正常预期红灯、分类拒绝、关闭 stdout 三类回归，每类校验历史哨兵不变；保留精确分类 14 类负例。旧实现隔离复现红灯，修复版转绿，不污染旧历史。
- 项目 README 明确 stdout/stderr、调用方选择独立新日志并保存真实退出码；既有 Trellis 运行契约原位更新，研发实际执行 before-dev/check/update-spec。
- Makefile、历史基线、分类器和业务源码未改；完整保留旧版健康方法真实 HTTP 红灯重建、当前健康/版本、Go race、前端、构建及真实 Go 服务浏览器旅程，不以受控输出 fixture 替代网络门禁。

## 验证结果

- 本 Run 宿主固定 `make verify`：平台 seq=5 总退出 **0**。受验 HEAD `066f035bc551b2b4d031166cebdab0849084afd4` 加修复工作树，36 文件摘要 `b33aa9f6f848a8cda1bf5af08ba9c2e540d0aa09f91090930ccafa6d4e636702`，独立 QA 前后完全一致。未改的串行 fail-fast 链支持全部固定阶段成功。
- 独立 QA：原 main overlay 加本轮回归 race exit **1**；修复后的三类子进程及分类负例 race exit **0**；格式/module/vet/27 前端 VM/build 单独补验 exit **0**。既有 P3 `QA12-HISTORY-WRITE` 本轮 Closed，新增产品缺陷 0，结论 **Go**。
- 研发执行前 106 份已跟踪历史文档/日志清单在宿主后、QA 前后和报告核对时全部一致；无人工恢复旧日志。新输出由调用方重定向保存，独立位置和失败诊断可定位。

详细依据（均为仓库路径）：
- `docs/workflow/runs/cf691e39d382439f92a3ac12f5facfff/acceptance/acceptance-report.md`
- `docs/workflow/runs/cf691e39d382439f92a3ac12f5facfff/acceptance/acceptance-matrix.md`
- `docs/workflow/runs/cf691e39d382439f92a3ac12f5facfff/acceptance/evidence/host-connector-receipt.md`
- `docs/workflow/runs/cf691e39d382439f92a3ac12f5facfff/verification/M01-development.md`（研发记录及报告对账）

## 限制与发布状态

研发和 QA 各自实际 `make verify` 因沙箱监听拒绝 exit **2 / Blocked**，之后本地当前 HTTP/full race/browser 未执行；所有失败及 Not Run 保留，不拼局部通过为完整门禁。完整门禁依据是本轮同版宿主回执，不是旧 Run QA。宿主文本截断，无后段原始日志/截图，不能提供逐项尾部、细粒度宿主测试数量或视觉质量结论。

报告对账时点 **2026-10-06 05:08 +08:00**：本 Run T001 **Done**、M01 **Accepted**、总览 **Delivered**（仅工程修复、验收与公共交付说明范围）；不再将已完成的宿主测试/QA 写成等待执行。历史 Issue #8/#9/#12 文档及日志原文保留。

本正文交平台 Connector 用于提交、推送和创建**草稿 PR**；报告时点这些动作尚未发生。未合并、未部署、未扩权、未调用真实供应商；受控测试上游不代表真实供应商验证。无数据迁移或运行配置变化；若后续需回退代码，应保留本轮验收和失败证据，并明确旧覆盖缺陷会重新出现，不能以人工恢复历史替代修复。
