# 联合测试计划独立评审

评审者：独立QA执行23，2026-10-10。对象：本Run `verification/test-plan.md` r1及执行21联合补正；非回写tests22为已进行文档评审。结论：**A/B受控交付计划合格；C仅是合法复测前置与依赖图，不是可直接执行的新供应商调用预算。** 旧QA10仅历史，未据其放行新候选。

权威材料根和sha见 `evidence/qa-23/material-integrity.json`，实际引用design-v0.2.0/docs/{prd,interactions,acceptance}.md、test-report/{report,independent-review,plan-review-rpm-tpm,plan-remediation-rpm-tpm}.md和repair-packages/DS-SAFETY-01.md。PRD§配额与预算要求独立维度、双scope与轮换共享；AC07/09–13/16要求原子性、未知不归零、守恒和安全调用；implementation-contract §4明确UTC固定分钟、受理次数/RPM、TPM已知+预留、可信上界、安全默认。预期不是由当前实现反推。

| PLAN | 独立判断 / 当前证据 | 范围限制 |
|---|---|---|
| 01 | 61项材料、ZIP及原文hash匹配；Pass | 不追写#33当时材料缺失为通过 |
| 02 | 正式bounds64→普通/SSE默认64及显式32，无冲突cap；租户TPM/KeyTPM/硬预算三种trigger；Pass | 真实收费输入/输出、隐藏推理、thinking策略Blocked |
| 03 | RPM2，1/2接受、3拒绝，边界前1ns拒绝、边界后恢复；两scope无部分扣计；Pass | 注入Store.now seam，不证明自然时间网页 |
| 04 | U7+P0+B200，206拒绝/207接受；406拒绝/407接受；跨窗399拒绝/400接受；旧窗核对不污染本分钟；Pass | 真实B不可用100+100夹具代替 |
| 05 | 窗口/回拨/重启与进程门禁同版执行；Pass | 时钟注入明确；不等于供应商自然跨窗 |
| 06 | 普通/SSE known、missing/inconsistent usage；成本33/租户费7分列；重复operation409单dispatch；Pass | 无真实账单核验 |
| 07 | failure known/unknown、dispatch前取消、输出后HTTP取消、timeout2秒分别有终态与池快照；Pass | undispatched只内部契约；取消不声称远端免收费 |
| 08 | 8竞争者屏障、1受理7拒绝，同ID零attempt和最终守恒；Pass | 不能替原DA双网页会话 |
| 09 | 多Key共享、轮换代际同池、另租户独立；原独立Key资源矩阵保留；Pass | 未声称网页非零池完整执行 |
| 10 | 改额/轮换/暂停/价格快照/晚到幂等/真实进程恢复既有门禁保留；Pass | Open/handler证据与真实进程证据分层 |
| 11 | limits/usage/period/effectiveUTC/attempt/hits/金额字符串数值日志和拒绝前后比较；Pass | minute来源是测试时钟，非GET新增字段；不含秘密或对话 |
| 12 | 禁止新供应商生成，H+X+潜在attempt≤24，单次64/90与unknown保守占位；前置说明合格 | H/X、精确模型、真实上界与思考能力未证；执行计划Blocked |

新候选Makefile未弱化：history/health-history、gofmt/modules/vet、锁定npm ci、278项前端、全包race、build和真实Go UI均执行。三宽provider/self_hosted及原PNG/几何/键盘/模态断言保留，tests22是联合全绿而非串接两个旧绿。首红是历史门禁/设施失败，不伪造不存在的生产算法首红。历史QA35-01重建旧编辑器的expected_failure仅证明测试敏感性，不作为当前产品失败。

供应商计划仍需专用测试Pipeline独立绑定每条生成的最大attempt、可信B、共享同ID证据、H/X及剩余额度后执行；当前不具备这些事实，不能自动推荐生成或扩大预算。原两屏、独立角色会话、自然UTC窗口、非零RPM/TPM、并发/池隔离、真实费用与恢复全部保留，未从范围删除。若无法在剩余授权完成，记录有证子集与未执行，不把内部矩阵伪装为供应商全验收。
