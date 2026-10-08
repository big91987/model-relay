# Development63：预算/时钟复审

基线 HEAD 66f1328ada75ea90e5638fbf2bd58663ac86347e；既有PR27未合并。seq57/QA58仅为修改前历史，本候选必须重新tests→QA。

材料 design-v0.2.0 / Frozen for Test Implementation，archive SHA256 5f51d35fdc15ed7214749f9413879b4238097f1d6b6bbe2cc0d508d8e103cef7。实际入口 `.workflow-input/4cd62ced6a4f98f50e3cca6f7596eb06/5f51d35fdc15ed7214749f9413879b4238097f1d6b6bbe2cc0d508d8e103cef7/files/manifest.json`；引用 docs/prd.md（e768b4f26b4471e714d92b0db6e3187c89f37a25a697c01a12c6079af07de547）及docs/acceptance.md（2792740b1579d614780856a13b6257d192736e1efc0309f1dc41405a665bef7a）。architecture-r1/implementation-contract第4节保持原决策。

## 变更边界

两项production handler/SQLite回归已在修改前exit1，见red.log：tenant/key零软预算200/1dispatch；旧Key截止401后回拨200/1dispatch。尚非真实HTTP复现，宿主须运行同场景并保留红→绿。没有基于只读评审就直接宣称失败。

零预算在accounting.go双scope检查修复，NULL无限、正软预算允许超额并告警不变。Store新增兼容schema3持久UTC高水位小表，在正常Open/recovery初始化；HTTP入口、受理及dispatch前事务采样，所有schema3 now读取已提交高水位，不在业务事务内嵌套DB访问，不持事务等待网络。初始化从旧admissions精确解析时间；旧版从未存储的拒绝观察无法重建。不修改既有账本/代际历史。旧回拨断言与合同矛盾（允许返回旧月），纠正为月份只前进。

文件边界：store/console_clock/console_storage/server/routing/accounting及回归、项目规范与公共任务。发现入口采样到配置锁之间仍可能排队，按同一精确期限契约将采样放到真实锁取得后、业务事务开始前；因此必要增加identity.go/console_api.go的同源采样，及真实锁/原子clock排队回归，不是另改身份/计费范围。非目标：UI/P3、商用高级范围、供应商凭据、Git/部署、冻结输入。已读取Trellis backend及guides实际源码规范；get_context报告单模块，不重新初始化/bootstrap。Skill公共事实沿用T001/M01。原生稳定监听限制沿用既有记录，不重复必败完整make verify；原固定门禁保留，完整HTTP/SSE/重启/迁移和1280/1440/390留tests宿主与独立QA。

## 保留失败及边界

- red.log：修改前预算两scope及回拨三项实际handler断言失败，exit1。
- green.log：修复后新测试费用断言未通过（同一effective timestamp的价格并列、随机ID排序使测试选价不确定），exit1；按公开API将测试价格生效时间分别前进1秒，非改DB记录、非改价格选择契约。
- green-2.log：对应调整后原Console InProcess及精确期限检查exit0。
- clock-regression.log：首个子进程exit0但未开启-v，仅输出PASS，父测试无法确认专用收据，因此exit1。加-test.v后仍要求专用标记，未弱化断言。
- race.log：扩大Console InProcess race、代际、当前Key/邀请、跨进程和正式backup/restoreexit0；其后追加有限Calls/RPM窗口守恒断言与fresh-init建表。race-final.log保留该次检查，增加锁排队精确期限后最终以race-current.log为准。
- static.log：make format-check、go mod verify、make vet、193/193Node、make build及历史红灯分类器race测试exit0。
- 历史监听环境限制：[attempt01](../attempt-01/make-verify.log)，bind operation not permitted、make exit2，首次工作树c5bddf89cbd94f19075acc1e38725213221342b249dd9e9ad7873db6d0e0b8c0。本次环境相关权限未变，未重复完整make verify。局部成功不是整套通过。
- 平台配置来源为本轮User Input的Codex/gpt-6.1-sol；本节点未取得独立的原生模型进程证明，不伪造模型切换或运行认证。

## 宿主/QA复验要求

在当前工作树运行原make verify。新增tests/review-history在隔离临时目录git archive准确PR27 head，复制当前最小回归，以实际Go HTTP要求tenant/key预算0 soft均200/1dispatch、旧代际到截止401后回拨200/1dispatch；若监听、build、race、超时或不符断言，不当作预期红灯。成功仅证明修复后重建旧版红灯，非原开发时网络成功记录。随后原Go race suite要求当前ActualHTTP预算矩阵和clockRecovery、全量原HTTP/SSE/迁移/重启与1280/1440/390实际Go UI全部通过。独立QA须读完整宿主日志至eof、核对新源戳/19材料，并逐项复验P1，不凭解释或历史绿灯关闭。

clockRecovery使用隔离tmp私密0600状态文件与单独Go测试子进程，仅测试秘密、不输出/提交。验证expired generation未retire仍401、current Key期限401、invite400；回拨和重启/官方backup-restore不重新开放已耗尽月份的次数额度，保留确认费用/预留/Pending，live generation仍可读模型。正常schema3 init、Open兼容已安装schema3，schema1/2正式CLI迁移和恢复走原门禁，旧账本不改写。

## 最终本地检查及源戳

2026-10-08，HEAD66f1328ada75ea90e5638fbf2bd58663ac86347e + budget-clock-r1工作树。140文件SHA256 **ca8ec50aa168cf16ac2e24edfe214aa44d06a05846c34deadbc3c7cb6c02904a**，Go1.25.6；实际Go inventory回执source-evidence.json。material-integrity.json重新核对archive及全部19文件一致（all_match=true，状态仍frozen-for-test-implementation）。材料/源码都无旧Run工作区导入。

| 检查 | 命令/范围 | 退出及证据 |
|---|---|---|
| 当前产品/回归race | `GOCACHE=.cache/go-build GOTOOLCHAIN=local go test -race -count=1 -v -run '^TestConsoleClock.*InProcess$\|^TestConsoleReview.*InProcess$\|^TestConsoleGenerationExactDeadline$\|^TestConsole.*InProcess$' ./internal/relay`；表中反斜杠为Markdown转义，实际正则使用普通管道 | exit0，85.696s，race-current.log；真实生产handler/SQLite+Go子进程，但非实际HTTP |
| 当前静态/前端/构建 | make format-check、go mod verify、make vet、make frontend-test、make build、Go historical classifier race | exit0，static-current.log；Node193/193 |
| 差异/敏感路径 | git diff --check；本次证据无本机绝对路径/供应商凭据 | exit0；链接警告的临时地址脱敏，原警告文字与失败仍保留 |
| 原固定完整make verify | 未在原生稳定禁监听沙箱重复执行 | Pending宿主；历史限制exit2不冒充当前失败/成功 |
| 实际HTTP历史红→当前绿 / 原全量三视口 / 独立QA / PR更新 / 部署 | 通过正式tests→QA→report/publish/pr继续 | Pending，P1-1/2未标Closed；不得沿用QA58源戳 |

Trellis-check已完成源码/spec及同层复用/跨层审核和可执行检查：高水位在业务事务之前提交，取得配置锁后重新采样避免排队过期；server.advanceClock统一503 fail-closed；没有嵌套SQLite事务/网络持锁、有期限/窗口/费用不变量回归。完整质量结论待宿主与独立QA。trellis-update-spec已记入backend/multitenant-contracts及索引；无需bootstrap或新建组织级文档。新正软预算仍可超额并返回soft_budget_exceeded；零预算/NULL/一次性秘密/标准迁移、原门禁没有放宽。
