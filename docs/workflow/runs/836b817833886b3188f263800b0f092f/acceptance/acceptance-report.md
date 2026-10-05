# Issue #12 密码显隐独立验收报告

> 日期：2026-10-06，Asia/Shanghai（开始冻结 04:24，本轮实际命令时间见日志）
> 仓库：model-relay；分支：workflow/836b817833886b3188f263800b0f092f
> HEAD：73681a7df4c5cbe5914056ac13348154d58726e0，未提交工作树摘要：0f6ee2d6894067323697f632231ee0d62b098be0261a4da7e975edb6e1efb37b
> 制品/环境：macOS arm64、Go 1.25.6、Node 26.0.0；锁定 Playwright 1.59.1、go-sqlite3 1.14.52、x/crypto 0.55.0；真实 Go 测试服务 Version=dev；隔离临时数据库/随机端口；无生产部署/集群/真实供应商
> 模式：首次全量验收（新增薄切片及项目公共门禁）
> 结论：Go with known issues（P3 验证工程历史日志写入副作用，本轮已恢复现场；无未关闭 P0/P1）

## 1. 结论摘要
已独立核对 AC01—AC13 的基线、实现、测试契约、同轮宿主固定 make verify 回执和源码一致性。默认隐藏、保值、键盘/AX、零提交、两态真实登录、拒绝恢复、退出/会话恢复与既有管理/健康/版本链由同轮宿主退出 0 支撑。QA 自己执行 make verify 退出 2，原因是沙箱禁止 HTTP 监听，不写成 QA 本地全绿；独立非监听检查通过。旧 10 份文档及 62 份日志全部恢复并哈希核验，失败/新旧重建材料分别保留。相较研发仅非监听通过和门禁受阻，本轮增加同版本宿主真实旅程证据与独立复核，可交报告/PR 节点，不代表已经发布、合并或完成用户最终业务验收。

## 2. 本轮覆盖
权威需求与设计分别为本 Run product-definition.md / password-visibility-architecture.md，不以旧 QA 或当前实现改写预期。逐项证据见 acceptance-matrix.md，核心执行契约见 journeys.md。

- J01—J02：同一未登录管理员在同一 context 完成首次/刷新、空值/特殊输入保值、原生键盘与鼠标切换、可访问动作/状态、零 submit/invalid/network；宿主 UI-E2E Pass。
- J03—J04：同一管理员连续错误重试、隐藏态 Enter 登录、退出、显示态登录、有效会话刷新、失效回登录、重新输入/重登；宿主 UI-E2E Pass。测试服务重启是显式环境故障准备；产品重登在 UI 完成。架构规定的 logout→reload 失效路径明确断言隐藏/空值，管理请求 401 另有 VM 清理回归。
- J05：同会话 URL/storage 不变与完整生命周期，记录/服务日志秘密排除对账；UI 与 Data Reconciliation 分行，不靠 API 补 UI 步骤。
- J06：1280px/390px 都连续完成空库首次配置到存量管理、普通调用/SSE/取消/错误恢复、停用、持久化重开和版本读取/未知/重试/迟到隔离；真实 Go、显式受控 upstream，不是真实供应商。
- 健康：同轮宿主 race/HTTP 公共回归 Pass；QA 独立 handler/version race Pass，二者不混称。旧版重建中的拒绝方法 FAIL 是指定历史版本预期红灯，不是当前缺陷。
- 小增量无新多租户/迁移/异步控制器，组织级运营测试不适用；现有版本竞态/错误、权限失效和首次空配置仍保留。不制作不适用 PRD/原型或重复设计。

## 3. 缺陷及进展
见 defects.md。产品 P0/P1/P2 未发现；已确认既有 P3 QA12-HISTORY-WRITE，脚本可重写旧日志，当前现场已完整恢复，未来仍需输出保护，不静默标为 Accepted Risk。没有虚构历史缺陷关单，也不把沙箱权限错误归咎产品。P3 默认非阻塞。

## 4. 证据与版本
| 来源 | 实际内容/结果 | 文件 |
|---|---|---|
| 平台同轮 seq7 | 固定宿主 make verify 退出 0；摘要匹配；回执截断不伪造原始尾部 | evidence/host-connector-receipt.md |
| 宿主生成文件 | 指定旧版本 HTTP 重建及预期拒绝方法失败 | evidence/host-health-history.log |
| QA 本轮 | make verify 退出 2，到 health-history 监听受限；后续未跑 | evidence/qa-make-verify.log、qa-health-history.log |
| QA 本轮 | 格式、vet、27 项前端 VM、模块验证、build、race handler/version、journey 语法退出 0 | evidence/qa-non-listening-checks.log |
| QA 本轮 | 冻结/最终 SHA、72 历史哈希、diff、当前源码及新规范标识 | evidence/snapshot.log、final-integrity.log |
| QA 本轮 | 实现、测试真实性、Trellis 记录/规范及数据保护审查 | evidence/static-review.md |
| 研发本轮（非 QA） | 第一次固定门禁失败、补充检查、Trellis 动作 | 本 Run verification/M01-development.md 及其关联日志 |

源码/锁/README/已有 acceptance-design 的 36 文件摘要在宿主、QA 初始、最终一致；本次仅新增/更新 QA 文档和恢复历史路径，未修改业务实现。QA build 制品摘要另存 final-integrity.log，未运行生产部署。旧 Issue #8/#9 文档和日志只作保护对象/背景，不能作为本轮成功证据。

## 5. 未执行与环境限制
QA 本地固定门禁为 Blocked；QA 本地当前完整 HTTP 和浏览器阶段为 Not Run。逐项保留这些行，以同轮同摘要、同用户结果的固定宿主门禁作为发布条件替代证据，而不是改本地行为 Pass。静态 browser.check 不适用真实 Go，因此没有调用。初次直接 go run 遭系统缓存拒绝，沿用现有项目 cache 重跑成功，原错误已保留并脱敏。未扩大权限。

平台回执文本被截断，QA 未收到完整当前 HTTP/浏览器尾部和截图；只基于可信退出码、相同快照、未修改 fail-fast 入口和实际断言范围作门禁判定，不虚构个别宿主测试数量/耗时或声称亲自观察浏览器。无视觉截图/音频证据；AX 用例只验证浏览器语义。生产、付费、真实供应商、设备扩展/密码管理器和屏幕阅读器音频未验证，也不属于本轮交付承诺。

## 6. 阶段边界
此次只验 Issue #12 和旧健康方法/版本回归。PR 正文由报告节点写本 Run/pr.md，Issue 更新、Git 分支提交推送、草稿 PR 均由平台处理。不修改产品/门禁、不直接发布。不以当前结果代替未来架构闭环、新功能、真实供应商集成或用户最终合并决定。

## 7. 环境恢复
固定 Go Harness 以 defer 关闭服务、store、listener 并清理临时目录；宿主退出 0 且无共享数据库操作。QA 受限重建临时目录由程序清理，缓存/bin 为既有忽略目录，无凭据或 node_modules 入提交。两轮重写旧日志均先保存本轮材料再恢复 HEAD；72 份历史文件最终全 OK，历史 diff 为空。回执本机路径已脱敏，密码/请求正文/环境凭据未收集。

## 8. 质量趋势
相较本轮研发交接，验证强度提高：固定宿主真实 HTTP/UI 已成功且快照匹配；QA 独立补充检查与历史保护。原有受限失败继续保留，未删除历史红灯或伪造修复。新增功能回归资产为 VM 状态/保值和真实原生键盘/AX/零请求/两态登录；未发现新增产品缺陷或已修问题复发。剩余风险为 P3 历史日志写入副作用、截断回执的诊断完整性与显示密码固有旁观风险；不声称相较历史 Issue 的测试数量增长或其缺陷关闭。

## 9. Go / No-Go
Go with known issues：强制 AC 有同轮同版本证据，无未关闭 P0/P1；P3 非阻塞，本轮历史保留条件已恢复并验证。走 next/report。后续不得改业务/测试/锁再沿用此绿色；若源码摘要变化，必须重新固定 make verify 与独立复核。任何再验证仍须保留失败、保护旧日志，不可扩权或用静态展示替代真实 Go。
