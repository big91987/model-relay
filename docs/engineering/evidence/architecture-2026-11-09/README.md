# 本轮证据怎么读

先看[治理前后说明](../../architecture-remediation-2026-11-09.md)和[包地图](../../architecture.md)，再查下列证据。源码基线为 `064eb0b9c52d9b04a72ed3cd1dd70eeac756e58b`；最终架构材料快照为 `328ccea19f55b02826c37f3b48d89730e3153188fc695003288d1115b7a45fbe`。报告中的 `git_head` 是执行时尚未提交的基线，实际内容由逐文件摘要和 snapshot 标识，不能只按 HEAD 判断。

| 文件 | 证明什么 |
| --- | --- |
| [structure.json](structure.json) | 实际生产文件和功能目录的前后数量、路径，以及安装工具的摘要无漂移 |
| [final/report.md](final/report.md)、[report.json](final/report.json)、[model-result.json](final/model-result.json) | 标准入口接受的最终规则和模型结果；规则通过，模型保留 4 条待复核候选 |
| [dispositions.json](dispositions.json) | 原四项落实位置，以及最终四条模型候选的独立复核和处理依据；不改写原报告 |
| [test-migration.json](test-migration.json) | 原 118 个 Go 测试入口迁移关系及历史断言核对 |
| [dependency-probes.json](dependency-probes.json) | 隔离副本的反向依赖被规则拒绝，真实源码未注入探针 |
| [observation-mutation.json](observation-mutation.json) | 旧代次守卫的三个隔离删保护变体在预期断言失败 |
| [policy-counterexamples.json](policy-counterexamples.json) | 保留/移除额度预留与锁后时钟检查的真实 HTTP 对照结果 |
| [clock-disposition.json](clock-disposition.json) | 展示/审计时间与有效时钟的既有差异及保留限制 |
| [model-run-index.json](model-run-index.json) | 六轮模型调用的版本绑定、状态、错误、用量和原始报告摘要；`incomplete` 的原始候选不计为已接受发现 |
| [citation-correction.json](citation-correction.json) | 同一材料上仅更正两条不合法引用；所有判断、发现、顺序和其他引用保持 |
| [raw-report-manifest.json](raw-report-manifest.json) | 本目录原始标准报告的逐文件 SHA256 与来源，相应报告逐字保留 |
| [verify-state.json](verify-state.json)、[verify.log](verify.log)、[verification-summary.json](verification-summary.json) | 最终同版完整 `make verify` 退出 0；冷依赖安装、源码不变、真实回归与清理回执。状态与日志逐字复制，日志摘要与原回执一致 |

最后一次原报告因两条检查依据引用错误而被拒收，原始[拒收报告](rejected-citations/report.json)和[模型回复](rejected-citations/model-result.json)保持原样。使用忽略目录中的诊断 reviewer adapter，让同一 DeepSeek 模型针对原冻结材料修正引用，仍通过原安装工具的 `--review-command` 和校验器；不是手改 JSON、删除发现或修改校验器。更正轮 22.454 秒，输入 267312 / 输出 6094 Token，其中 261632 输入 Token 命中缓存。

前六轮各约 25–26 万输入 Token、108–223 秒；第一轮材料在扫描中变化而拒收，另一轮出现额外字段而拒收。中间报告只作诊断，已接受的实际职责问题进入源码修复；不以反复扫描后的问题数量作为质量分数。最终引用合法也只证明报告可定位，不证明每条整改建议适合当前代码。

本目录不包含模型凭据、完整发送材料、数据库、截图或生成工作区。真实 HTTP 反例的命令路径用 `$ROOT` 表示仓库根目录。更多原始执行记录保留在忽略的 `harness-output/governance-2026-11-09/`，公开证据不包含私有宿主路径。
