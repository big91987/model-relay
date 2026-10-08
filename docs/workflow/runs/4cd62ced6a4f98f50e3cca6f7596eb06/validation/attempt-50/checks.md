# development50：editor-lock-r1 修复候选

## 基线、范围与授权

2026-10-08，同Run4cd62ced6a4f98f50e3cca6f7596eb06；HEAD98511771871cf0951ecef716bbb55deb13e18da5，未提交、未切分支。实际待验工作树以[source-evidence.json](source-evidence.json)为准。直接承接architecture-r1、T001、management-summaries-r1及QA47，不恢复旧工作区、不改冻结输入。原生模型无法从文件或Go版本独立核实；Codex/gpt-6.1-sol配置与进程核对仍属平台事实，未自行标记通过。

先读取输入manifest：`.workflow-input/4cd62ced6a4f98f50e3cca6f7596eb06/5f51d35fdc15ed7214749f9413879b4238097f1d6b6bbe2cc0d508d8e103cef7/files/manifest.json`。version design-v0.2.0 / frozen-for-test-implementation，archive SHA2565f51d35fdc15ed7214749f9413879b4238097f1d6b6bbe2cc0d508d8e103cef7。实际引用同root的docs/prd.md、docs/interactions.md（编辑/忙状态/未知结果）、docs/acceptance.md（AC01/三视口）；19件SHA/大小及archive复核见[material-integrity.json](material-integrity.json)。未执行材料脚本，无需再次审批。

## 最小诊断与实施

[seq49失败摘要](host-seq49-summary.md)保留原make verify exit2与准确指纹。恢复段未包含最终snapshot，不臆断原现场disabled值。实际源码VM红灯证明相符的产品缺陷：写确认后readback被导航中断，旧epoch finally不解锁，新edit未清旧disabled。此缺陷已修，宿主异常是否全部消失仍待实测。

- `internal/relay/web/workspace.js`仅新edit初始化Save disabled=false；保持同步busy、未知锁、原epoch finally及安全读回。
- `tests/browser/workspace.test.mjs`新增实际函数红/绿回归及旧finally不能解锁下一提交、单响应等待与driver启停/取消/清理/回执测试。
- `tests/browser/console.mjs`的save要求真实Save可用，单响应等待与click由Promise.all共同拥有，返回同response供创建租户/服务读取，不再产生原service超时的重复孤立等待器。新增三视口实际服务201后的controlled GET hold→导航→新编辑可保存→取消→真实服务POST总数1的有界断言。
- `.trellis/spec/frontend/workspace-contracts.md`保存生命周期与好坏案例。后端、SQL/金额/快照、Go测试、依赖、Makefile、原视觉/continuous/fault断言不变。

不扩产品/权限，不改数据库来过测试，不统一重写旧driver或增加生产重试。受控时序用于证明新编辑锁恢复，不证明供应商或操作系统取消机制。

## 实际验证

| 命令与证据 | exit | 范围 |
|---|---|---|
| `node --test --test-name-pattern='a new editor owns' tests/browser/workspace.test.mjs`修复前，[editor-red.log](editor-red.log) | 1 | 新edit disabled仍true，非浏览器现场 |
| 同命令修复后，[editor-green.log](editor-green.log) | 0 | 同断言绿，含旧finally隔离下一提交 |
| `make frontend-test`，[frontend.log](frontend.log) | 0 | 187 tests，零fail/skip，含原184及新增3项 |
| `make format-check && go mod verify && make vet && make build && node --check internal/relay/web/workspace.js && node --check tests/browser/console.mjs && git diff --check`，[static.log](static.log) | 0 | 原格式/锁定模块/vet/Go构建/JS解析/diff，无新依赖 |
| `go run ./tests/evidence`，[source-evidence.json](source-evidence.json) | 0 | 当前指纹，非health/模型/发布证明 |

不重复跑已确认原生不能bind的完整make verify；首次限制保留attempt-01/make-verify.log，环境未变化。本轮未改Go，不把seq49旧版本race继承成当前网络通过。日志去本机路径见[log-redaction.json](log-redaction.json)，无秘密、storage、任意DOM或供应商正文。文档patch第一次因错误标题定位拒绝、未写入；更正为实际标题后应用，不影响源码或测试。

## trellis-check 与公共交付

实际Trellis0.6.15及源码规范非模板；get_context.py --mode packages确认single-repo backend/frontend。按before-dev恢复frontend索引、DOM/lifecycle/verification/workspace及code-reuse指南，确定行为所有者与最小写集；按update-spec保存锁契约。无需bootstrap、无后端规范变化。trellis-check：新增行为有回归，无debug/类型绕过/新依赖，DOM ID/原epoch匹配；未以VM代替Chromium，所有原入口保留。研发Skill公共契约/审查清单沿T001/M01维护，不重建PRD/架构/任务。结论Blocked by Validation（可交tests，非用户业务阻塞）；T001仍In Review，M01非Accepted。

## 固定宿主复验与未验证项

本次交接被平台暂拒后收到2026-10-08用户正式部署查看指令，已承接到原[T001](../../delivery/tasks/T001-console.md#2026-10-08用户补充正式部署后查看最新实际页面)，不改变代码/门禁或新建任务。只读核对正式deploy-local workflow与本机部署文档；tests→QA→report→publish→pr后由协调者核验并触发已授权测试仓合并/正式Actions。后续report负责当前候选/QA/PR绑定、正式main部署步骤、实际页面入口与安全首次登录指引，禁止旧main冒充最新或临时服务替代。本节点不代发GitHub动作；用户提供workflow375348269、preview5545，运行成功及效果URL尚未验证。

1. 对source-evidence.json一致候选跑原完整`make verify`，timeout900s。保留seq49失败及此次完整日志/exit/指纹；失败返回development，不拉长等待、删断言或归因到沙箱限制。
2. 原console1280/1440/390须产出`test-results/editor-interruption-{width}.json`：真实服务POST1、取消POST0、新Save enabled，明确controlled GET hold。此为预期输出，非已有本地证据。
3. 承接attempt48新增TestConsoleSummariesActualHTTP、`test-results/management-summary-{width}.json`，原连续J01–J07/视觉63图/fault/资源/迁移/恢复/费用门禁全部保留。
4. 完整通过后独立QA继续同19件摘要及当前指纹，决定QA47-01 P1/AC01 Fail与QA47-02 P2是否关闭；研发不改历史QA或自签Go。

真实供应商/外部身份/支付、平台模型及负向Harness、PR/merge/正式Actions部署/healthz效果仍Not Run；当前未发布，Git/部署由Connector与协调者处理。
