# 缺陷与环境问题

| ID | 等级 | 用户能看到的问题 | 复现步骤 | 影响 | 定位信息 | 当前状态 | 回归资产 |
|---|---|---|---|---|---|---|---|
| QA12-HISTORY-WRITE | P3 | 重跑验证后旧健康复现日志被本轮结果替换，可能误读旧证据 | 运行固定 make verify；检查旧 history log 哈希/差异 | 验证工程历史整洁，无产品行为影响；本轮已留存生成文件并完整恢复旧历史，因此 AC13 最终保留成立 | tests/health-history/main.go 写 docs/workflow/health-methods-history.log；未改可信门禁 | Open（既有脚本副作用；本轮现场已恢复；研发 Owner 后续维护） | historical-logs.sha256；本轮 host-health-history.log / qa-health-history.log；final-integrity.log |

P0/P1/P2 产品缺陷：本轮未发现；无虚构关闭或降级。此 P3 在研发交接已提示，本轮独立确认，并非首次新增功能缺陷或复发修复。默认不阻塞交付，不把未裁决问题写成 Accepted Risk。

后续维护验收语句：固定验证入口不得削弱；生成本轮重建结果仍须留存，但旧 62 份日志和 10 份文档哈希始终保留。可由研发在独立维护范围评估输出路径/保护包装，不由 QA 修改产品或门禁。当前发布和任何后续验证先保存输出、再恢复基线、再核验，禁止覆盖现有证据文件或删除失败。

## 非产品缺陷
- QA 沙箱：固定 make verify 退出 2，health-history 的真实 HTTP listen 127.0.0.1:0 被拒绝；保留 Blocked。未扩大权限，同版本宿主固定门禁退出 0 为同轮替代证据。
- Connector 输出截断：只能核验总体退出码、首段源码摘要与脚本断言范围，没有完整浏览器原始尾部/截图，不声称提供这些材料；保留为证据可诊断性风险，不擅自捏造日志。
- 初次冻结时直接 go run 使用系统 GOCACHE 被拒绝；已保留并按现有 Makefile 的 workspace cache 环境重跑，摘要一致。不是产品失败。
