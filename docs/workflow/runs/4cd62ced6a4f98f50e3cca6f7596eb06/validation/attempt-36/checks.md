# 执行36 · safe-readback-resource-matrix-r1

2026-10-08；同Run研发整改。T001 **In Review** / M01 **Ready for Review**，退出仍 **Blocked by Validation**；QA35 No-Go没有自行撤销。本节点没有需求/设计歧义、用户决定项或权限扩大。

## 承接、版本与变更边界

权威输入：`../../architecture.md`、`../../implementation-contract.md`、`../../verification-design.md`、`../../delivery/tasks/T001-console.md`，以及 `../../acceptance/remediation-handoff-35.md`、`acceptance-report-35.md`、`acceptance-matrix-35.md`、`user-journeys-35.md` 和原生产handler红灯。初始来源继承QA35/seq34的HEAD `98511771871cf0951ecef716bbb55deb13e18da5`、124文件 `50fb764aac5443687e376c18906b36b765f02e9b05788fa4d5ac73617cbc98a4`；见 `source-before.json`，不伪称初次本地指纹命令成功。

本轮设计材料仍为 **design-v0.2.0** / Frozen for Test Implementation。archive SHA256 `5f51d35fdc15ed7214749f9413879b4238097f1d6b6bbe2cc0d508d8e103cef7`，manifest为 `.workflow-input/4cd62ced6a4f98f50e3cca6f7596eb06/5f51d35fdc15ed7214749f9413879b4238097f1d6b6bbe2cc0d508d8e103cef7/files/manifest.json`。实际引用该root中的 `docs/prd.md`（Key未知结果、代际/额度共享、独立限额、未知用量）、`docs/acceptance.md`（AC08/09/16及全AC/H退出线）、`docs/interactions.md`、`docs/decisions.md`；原型仍是A工作台参考，不是后端事实。`material-integrity.json`记录archive及19件逐文件实际摘要，全部匹配。未改冻结文件、未执行包内脚本、未访问旧Run/New API源码或数据。

Trellis现场有 `.trellis/.version=0.6.15`、scripts与已实现的frontend/backend/guides规范；不是prepare初始false状态，也无需重复初始化/bootstrap。已按trellis-before-dev执行 `python3 .trellis/scripts/get_context.py --mode packages`（单仓frontend/backend），恢复具体规范及共享reuse/cross-layer指引。已按trellis-update-spec更新两个契约文件，按trellis-check核对本轮变更、生命周期、跨层事实/秘密边界及原入口。

最小行为差距位于生产 `internal/relay/web/workspace.js` 的 `edit`：未知写入首次安全GET失败时父节点textContent删除retry且没有操作引用。本轮只修编辑器所有权、补必要测试/规范/公共进度，不改Go鉴权/计费/迁移/API/DB规则，不处理无关布局或新高级功能。公共商用对照及全量分层矩阵沿用原任务与QA35矩阵，不宣称New API/LiteLLM完整对齐。

## 实现与强制回归映射

| 目标 | 本轮实际实现/用例 | 局部检查 | 宿主完整测试 | 独立QA / 发布 |
|---|---|---|---|---|
| QA35-01 / AC08恢复 | 原editor独立status/reference/retry；未知结果锁住保存及再次form提交；GET busy去重；epoch及dialog-open抑制迟到 | 生产抽取Contract通过；原红灯保留 | 当前asset真实UI仍待验 | 尚未Closed / Not Run |
| AC08签发 | 原console入口新增POST201已提交丢响应→GET503→GET404→同operation GET200；native双击/Enter仅一个已提交POST；对象1、无secret重放 | helper/语法通过，不是浏览器证据 | 1280/1440/390均已接入原browser | 待验 / Not Run |
| AC08轮换 | 已使用Key丢响应恢复；revision/generations仅+1、过渡1、usage及完整原请求/价快照不变；第二显式轮换409不叠加 | 原identity/key回归保留 | 三视口native双提交、读回失败恢复待验 | 待验 / Not Run |
| AC16快速发送 | native双击/Enter普通与SSE；受控真实上游350ms有界延迟，GET测试facts核对actual dispatch delta1；proxyPOST1/准入1/尝试1、Tokens3+4及同ID详情 | 原cancel/SSE路径不删；语法通过 | 三视口故障旅程待验 | 待验 / Not Run |
| AC09两层五维 | `TestConsoleResourceMatrix{InProcess,ActualHTTP}`：tenant与Key各calls/RPM/TPM/C/hard费用，0拒绝无dispatch→正有限精确预留→竞争拒绝两scope不扣→null恢复不清零 | 10个分例race通过 | ActualHTTP进入原go test，待宿主 | 待验 / Not Run |
| AC09共享竞争 | `TestConsoleSharedResources{InProcess,ActualHTTP}`：2个应用Key共享tenant；旧新代共享Key；barrier8争2、6拒绝；占满时其他tenant仍调用；历史及已确认金额不变 | 2个分例race通过 | ActualHTTP待宿主 | 待验 / Not Run |
| 原J01–J07及J06账本 | 原三视口连续21检查点、J06 fee21/cost7/原项快照不变/delta，以及Governance原19/60均未删 | 既有6项定向race通过 | seq34是旧50fb版本；新127文件须全量 | 不跨版本Pass |

新增主入口文件 `tests/browser/console-faults.mjs` 由原 `console.mjs` 直接调用；`tests/browser/readback.test.mjs` 由原frontend-test glob收集；Go矩阵由原 `go test -race ./...` 收集。**Makefile原门禁未更改**，没有单独弱化的放行入口。受控facts只在测试上游，非产品API；所有操作通过正式管理/代理入口及独立临时目录，未改业务数据满足断言。

`previousEditorAsset` 在测试页route内重建整改前editor，并记录其SHA256：真实Go已提交201后丢响应、读503，要求原retry确实消失，然后解除asset route并reload真实当前asset做完整绿色恢复。它是**修复后隔离重建旧行为**，不是原开发时已经跑过Chromium的记录，更不计当前产品Pass。原生不能监听，本轮仍未执行这两个真实浏览器阶段；不能拿当前Contract补造UI红/绿。JSON与安全截图由宿主实际执行产生；失败时也保存已达到的cases/phase/status，父driver输出固定阶段与安全快照。

## 真实本地检查与失败

| 命令/证据 | 实际exit | 层级/边界 |
|---|---:|---|
| QA35原 `readback-recovery.test.mjs`，修改生产之前重放；`readback-red.log` | 1 | 生产抽取Contract：mutation1/read1/retry=false/operation-visible=false；非Chromium |
| 首次新增资源矩阵，`resource-driver-red.log` | 1 | 新driver错用tenant期望revision2；实际fixture初值1，409拒绝。按实际契约纠正，不改产品 |
| 第二轮新增矩阵，`resource-unbounded-fixture-red.log` | 1 | null无界时产品不注入max_tokens，旧受控fixture仍要求100；非环境限制。仅新matrix显式允许无界省略，原fixture默认100断言继续生效 |
| `GOCACHE=$PWD/.cache/go-build go test -race -count=1 -v -run '^TestConsole(ResourceMatrix\|SharedResources)InProcess$' ./internal/relay`；`resource-check.log` | 0 | 2顶层/12分例；handler+SQLite+受控Transport，不是实际HTTP |
| 原Governance/ResourceDimensions/PeopleScope/PauseUsage/KeyCollection/FreshInstallation定向race；`existing-regressions.log` | 0 | 6个原顶层检查，保留原追加计费/隔离/暂停守恒等断言 |
| `make frontend-test`；`frontend-check-final.log` | 0 | 169Node检查，含旧editor重建红行为与当前恢复Contract；不是nativeUI |
| `make format-check`、`go mod verify`、`make vet`、`make build`；`static-check-final.log` | 各0 | Go格式/锁/静态/真实构建；无新增生产依赖 |
| 三个修改JS文件 `node --check`、`git diff --check` | 各0 | 语法/空白检查 |

原日志仅目录脱敏，`log-redaction.json`保留raw/redacted摘要及字节数，所有失败断言/退出事实和链接器LC_DYSYMTAB警告保留；警告未阻止exit0。最初 `go run ./tests/evidence` 使用系统默认cache受沙箱拒绝exit1，改用项目GOCACHE后成功，不能把失败初次输出补成绿。已知原生监听限制及其首次失败继续引用 `../attempt-33/checks.md` 和既有记录：环境没有相关变化，**没有重复原生必败的make verify、没有提权、没有用static browser.check替代真实Go页面**。宿主seq34的绿色不适用于本轮修改。

## 当前来源与交接结论

当前 `source-evidence.json`：HEAD不变，127文件指纹 **3ee2122eadac33a67342523a082f791d8a81154cc86b7321d53de856866a15ef**，Go1.25.6；交接前再生成 `source-after.json` 核对一致。文档回执不参与自引用指纹。模型执行器配置来自平台Codex/gpt-6.1-sol授权；本机沒有可独立验证实际原生模型的回执，本节点不伪造模型认证事实。

开发整改与新增回归已完成，**只请求tests执行同指纹原完整make verify**；不是产品完整Go、QA关单或发布。QA35-01、QA22-01残余覆盖必须等新宿主真实回执和独立QA才可关闭；QA22-02已经Closed的旧事实不重开。供应商/外部身份/支付、Harness负向/实际模型、PR/merge/Actions/Deployment/固定URL及health发布一致性继续Not Run或平台责任。没有提交、推送、创建PR、合并或部署。
