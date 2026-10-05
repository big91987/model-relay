# Issue #8 独立 QA 最终复验（2026-10-06，seq 9 后）

## 当前结论：Go

按当前 `task.md` 和用户原始 Issue #8 短轮复验，**QA-H8-01 Closed，走 next → report**。旧业务版本已在宿主真实 HTTP 层复现预期方法断言失败，当前版本完整 `make verify` 的宿主 Connector 回执 exit 0；独立 QA 核对代码、摘要、旧版来源与每个红灯断言，并独立执行非监听补强测试。没有新增已复现产品缺陷或未关闭阻塞项。本结论不代表部署或真实供应商验收。

## 版本与执行者

- 分支：`workflow/1382a96a9e697c58ba68fe3c3c26f34b`；HEAD：`9b6994ed80de2bdba79ff267419150dccaac1360`。尚未提交，不以 HEAD 单独代表被验实现。
- QA 实际运行 `go run ./tests/evidence` 得到 35 文件摘要 `edb4224576eb1ef488c33d9e8ab354bdba90a8a39668bb75f0c548a616f7dd48`，与 seq 9 tests Connector 回执首段及 `health-methods-history-version.json` 完全一致。旧 seq 6 的 33 文件绿色不冒充新摘要验收。
- 进入本轮时 tracked 修改为 Makefile、task.md、delivery.md、qa.md、server.go；untracked 为 health_test.go、tests/health-history/、health-methods 系列日志与两份版本 JSON。完整 `git status --short` 存于本轮独立日志；无 staged 差异。QA 仅新增独立日志、更新本报告，未修改产品或测试实现。
- 宿主执行者：平台 `previous_results` seq 9 / node tests / kind command / exit_code 0，固定完整入口 `make verify`。Go 1.25.6，锁定依赖沿用项目；未取得宿主 Node/Chromium 精确版本，不编造。
- 独立执行者：本节点 Agent 沙箱。实际执行摘要、`git diff --check`、旧版与测试文件 SHA-256、独立 Node 回执断言、`go test -race -count=1 -v ./tests/health-history`、`go test -race -count=1 -v ./internal/relay -run '^TestHealthMethodsHandler$'`，均成功。详见 `health-methods-qa-recheck.log`，各验证退出码单独记录。

## 红转绿证据与验收矩阵

| 项目 | 执行与独立复核 | 结果 |
|---|---|---|
| 旧版真实 HTTP 红灯 | seq 9 与落盘 history.log：healthy POST 实际 200/预期 405，Allow 空；三状态七种非查询方法共 21 个叶用例失败，42 条精确 status/Allow 断言；12 个 GET/HEAD/恢复叶用例通过；旧 go test exit 1 | Pass（预期失败复现） |
| 旧版来源可信 | 独立 `git archive` / `git show` / 最终 health_test.go 哈希与宿主 history.log 完全一致；只在隔离目录添加最终回归，不回滚正式业务源码 | Pass |
| 拒绝伪红灯 | 核读判定器并独立运行 race 单测，正常事件及 14 类拒绝路径全部通过；bind、构建、超时、额外断言、race、skip、缺结果等均不能冒充预期旧版红灯 | Pass |
| GET JSON/故障语义 | 当前 TestHealthMethodsHTTP 三状态断言 healthy 恰为 status/version、故障 503/storage_unavailable；seq 9 完整 test 门禁成功 | Pass（宿主） |
| HEAD 状态/空体 | 当前真实 HTTP 三状态与 GET 同预期状态，body 为零；独立 handler 33 叶用例补强显式空体 | Pass（网络与 handler 分层） |
| 非查询方法与恢复 | POST/PUT/PATCH/DELETE/OPTIONS/TRACE/CUSTOM 三状态 405 + 精确 GET, HEAD；清除 degraded、正常 Close/Open 后 GET/HEAD 恢复 200 | Pass（宿主） |
| 完整公共门禁 | 当前 Makefile 串行且无忽略失败：evidence → health-history → format → module verify → vet → frontend → 全量 race → build → browser；seq 9 整体 exit 0 | Pass（宿主） |
| UI-E2E | 实际 Go 服务、1280px/390px 的同角色同会话连续浏览器脚本，由完整入口执行；非静态原型，非 handler/UI 单测替代 | Pass（宿主） |
| 范围/隔离 | server.go diff 仅健康分支，Host 校验/其他路由不变；临时安装与随机 loopback 端口；无供应商调用或用户数据库改写 | Pass（独立审阅） |

旧版来源独立核对值：archive `e8307563c976a77940851da8254d70b8255fa4caa690d898a6452ee15de4eaaf`；旧 server.go `d61e4e747231654aa7895404b94c2539497cef2074186b3bbe05b90a94252383`；最终 health_test.go `15b13fbd3cb3265c8bedb9a34f78980fea195351e458f5a4b83ca675513f7c3e`。独立解析宿主落盘 JSON，逐一验证 21 个失败的状态/Allow 文本、12 个查询通过及无 skip，不只相信判定器末行 PASS。

seq 9 会话内 stdout 在历史事件段被截断，**未逐行看到本轮后半段 test/build/browser 输出**；本轮完整门禁成功依据 Connector 真实整体 exit 0，以及所核对的 Makefile 串行执行和各脚本错误传播，不捏造被截断行。history.log 保存旧版完整事件和精确接受结果；seq 6 的未截断回执补充同一未变更浏览器资产的两个明确 PASS，但不是新摘要绿色的替代品。

## 真实网页旅程、恢复与未测范围

独立核读浏览器启动器及 journey.mjs：实际 Go relay、隔离持久存储、受控测试上游与控制端点；每个视口在自己的同一浏览器会话连续完成错误登录恢复、登录、上游配置/拒绝保存恢复、模型与密钥操作、普通及 SSE 调用/取消、记录查询、敏感信息排除、退出及退出后未授权检查。API 控制端点仅作显式故障夹具和日志审计，不代替网页表单步骤；受控上游不代表真实供应商。项目脚本无截图存储，本轮使用 Connector 回执，不补造截图。

Agent 没有重复监听/启动 Chromium：已有匹配摘要的宿主门禁，不要求扩大权限。原 bind/DNS/恢复锁失败及未到达检查日志全部保留；本轮独立 handler 成功仅证明 handler 层，不替代宿主真实 HTTP/UI。真实供应商、部署、额外 UI/读屏测试为 Not Run，非此次承诺，不能据此阻塞。LC_DYSYMTAB 链接 warning 为真实环境提示，未伪称产品失败。

历史 runner `defer os.RemoveAll` 清理隔离旧版本，HTTP 夹具 t.Cleanup 关闭服务与安装；浏览器 finally 关闭上下文/Chromium、Go defer 关闭服务/存储并清理临时目录。宿主成功回执与清理路径可复核，但没有单独宿主进程盘点回执，不声称另做了进程检查。未触及生产或用户安装。

## 缺口关闭与后续

QA-H8-01 从 Blocked 转为 Closed：新增真实历史 HTTP 网络红灯，与当前完整绿色形成原发现层闭环；新增判定器回归可重复拒绝环境错误伪红灯。没有新增或复发产品缺陷。原开发时的 HTTP 尝试仍是环境阻塞，**补证是修复后隔离重建，不能改写为修复前当时已成功运行**；这一时序限制永久保留，按上轮明确整改方案关闭证据缺口。

仅放行本次修复的汇报/交付流程。report 后由既有 Connector 回写 Issue 真实证据并提交草稿 PR，不提交/推送/评论/创建 PR/部署。若后续改变产品、测试或构建资产导致摘要变化，应重跑固定 `make verify` 并复核，不沿用本结论。

---

# Issue #8 独立 QA 首轮（2026-10-06，seq 7 历史结论）

## 当前结论

**Blocked / No-Go：修复后的行为与完整门禁通过，但强制的旧版真实 HTTP 红灯证据尚未完成。** 返回 `development` 补充验证资产，再经过固定项目 `tests` Connector；不要求改动已通过的业务实现，不扩大权限，不重做需求/设计，不部署。

本轮为局部 API 缺陷短轮复验，依据当前 `task.md` 和用户原始 Issue #8；下方 v0.1 QA 为历史记录，不作为本次结论。未发现新的已复现产品缺陷。阻塞项为 QA-H8-01，与用户明确要求的“先补真实 HTTP 回归、保留旧实现失败证据”直接相关，而非沙箱拒绝监听本身。修复后真实 HTTP 已通过，不能因此反推修复前的真实 HTTP 已复现。

## 被验版本与变更

- 仓库：`big91987/model-relay`；分支：`workflow/1382a96a9e697c58ba68fe3c3c26f34b`。
- HEAD：`9b6994ed80de2bdba79ff267419150dccaac1360`；实现尚未提交，HEAD 单独不足以标识被验代码。
- 独立执行 `go run ./tests/evidence`：33 文件摘要 `917c0548d72974f6df7dbed371036e070c67a28f958b242d532d884734721089`，与平台 seq 6 tests Connector 回执、研发交接和 `health-methods-version.json` 完全一致。摘要包括产品、测试、构建锁、README 和验收/设计基线，排除交付回执；本轮 QA 文档/日志不改变该摘要。
- QA 进入时已修改：`internal/relay/server.go`、`docs/workflow/task.md`、`docs/workflow/delivery.md`。未跟踪：`internal/relay/health_test.go`、`health-methods-version.json` 及八份 health-methods 开发日志。没有 staged 改动。业务 tracked diff 仅为健康检查分支，新增测试单独核读；其他路由与 Host 校验未改动。
- 工具链：独立核对 Go 1.25.6 darwin/arm64、Node v26.0.0，项目 Playwright 锁定 1.59.1。宿主回执确认 Go 1.25.6；没有宿主 Node/Chromium 精确版本回执，不编造一致性。

## 测试执行者与独立复核

**宿主测试：** 平台提供的 seq 6、node `tests`、kind `command` 回执，完整 `make verify` exit 0。QA 阅读原回执而不是仅接受开发者总结，核对 Makefile 仍依次执行 evidence、format-check、模块校验、vet、frontend-test、race Go 测试、build、browser，没有删除或缩减门禁。

回执包括前端 4/4、CLI/崩溃恢复、HTTP/SSE/取消/上游故障/存储 fail-closed、全量 race Go 测试、构建，以及 `TestHealthMethodsHTTP` 三状态全部 33 方法/恢复子用例通过。两个浏览器结果为 `PASS actual Go service browser journey 1280px` 和 `390px`，明确为受控上游而非真实供应商验证。macOS linker 的 LC_DYSYMTAB warning 未导致门禁失败，保留为环境提示。

**独立 QA：** 核对当前源码 diff、新回归断言、隔离夹具、浏览器启动器/连续旅程及摘要算法；实际执行摘要、`git diff --check` 和 `go test -race -count=1 -v ./internal/relay -run '^TestHealthMethodsHandler$'`，均 exit 0。辅助 handler 33 子用例通过，仅支持 handler 层，不替代宿主网络/UI 执行。独立执行证据见 `docs/workflow/health-methods-qa-independent.log`。

最初直接运行摘要时未配置项目 GOCACHE，访问默认用户缓存被沙箱拒绝，摘要未生成；改用 Makefile 同款 workspace `.cache/go-build` 后成功，未扩大权限。QA 未重复执行已由匹配版本宿主完成的监听门禁，也未将 sandbox bind/npm DNS 故障算作产品断言失败。

## 验收映射

| 项目 | 实际覆盖与证据 | 结果 |
|---|---|---|
| GET JSON 和故障语义 | HTTP healthy 校验恰有 status/version 两字段；degraded、database_closed 校验 503/storage_unavailable；宿主 seq 6 | Pass |
| HEAD 状态一致且空体 | 三状态真实 HTTP 与 GET 同预期状态码，读取 body 长度为零；辅助 handler 补强显式不写 body；seq 6 + QA 日志 | Pass |
| 非查询方法边界 | 三状态 POST/PUT/PATCH/DELETE/OPTIONS/TRACE/CUSTOM 全部 405，Allow 精确 GET, HEAD；代码为拒绝所有非 GET/HEAD 的通用分支；seq 6 | Pass |
| 恢复 | degraded 清除与数据库正常 Close/Open 后 recovered_GET/recovered_HEAD 均 200；seq 6 | Pass |
| 隔离和范围 | installation 使用 t.TempDir，localServer 监听 127.0.0.1:0；无用户配置/供应商访问，无数据库记录改写；diff 仅健康分支 | 独立代码复核符合；网络执行见 seq 6 |
| 完整公共门禁 | 当前摘要匹配的完整 make verify exit 0，含全量 race、build 与真实服务 browser | Pass（宿主执行） |
| 修复前真实 HTTP 复现 | health-methods-before.log 的 HTTP 三状态在 bind 时失败，未发送 POST、未抵达断言；仅 Handler 出现 POST 200/缺 Allow | Blocked：QA-H8-01 |

## 网页连续旅程与未测边界

本次没有 UI 变更，但保留完整入口的真实网页回归。独立核对 `tests/browser/main.go` 启动实际 Go relay、隔离临时存储和随机端口，执行 `journey.mjs`；不是静态原型。每个 viewport 使用自己的同一浏览器 context 连续执行登录、上游配置及非法地址恢复、模型创建/编辑、密钥创建、重启后重新登录与状态确认、普通/SSE/取消/上游拒绝、记录与秘密排除、密钥停用、第二密钥成功调用、退出。控制端仅承担显式隔离重启和日志审计，不代替用户表单操作；API 读取补充最终事实。

宿主回执记录两个完整旅程成功；QA 未自行启动 Chromium。现有资产不保存截图，本轮使用实际执行回执而非图片作为依据，不补造截图。真实供应商、部署、额外 UI/读屏检查不属于本次验收，Not Run，不因无关范围阻塞。研发初次安装锁恢复测试失败、sandbox bind/DNS 与未到达 browser 的旧失败均保留在原日志；匹配当前摘要的宿主完整通过关闭修复后运行缺口，不能关闭旧版网络复现缺口。

## QA-H8-01 整改交接

类别：强制验收证据缺口，不是已证实的产品功能故障；Owner：development。推荐不改当前 server.go、不回滚工作树、不重新设计，只补充可由既有宿主 Connector 运行的历史版本复现资产。

1. 在隔离临时目录构造入口 HEAD 的旧业务源码，加上最终健康回归测试（包含已修正的正常 Close/Open 恢复处理）。记录旧 HEAD、测试文件摘要和被测源码摘要；禁止修改正式 checkout 或用户数据。
2. 经项目验证入口/既有 tests Connector 执行旧版 `TestHealthMethodsHTTP`。必须实际监听随机端口并抵达断言，记录健康 POST 实际 200、预期 405 和缺少 Allow，以及其余非查询方法/故障状态结果。旧 Go net/http 可能已自动抑制 HEAD wire body，不要求 HEAD 必须红灯；不能把 bind 错误算成复现。
3. 验证资产应精确区分预期的旧版断言失败与意外启动/构建/恢复失败，不可把任意非零退出都接受为红灯；临时实例/目录清理，日志脱敏。
4. 保留当前回归，并在最终代码上再次运行不缩减的 `make verify`，提交旧版真实红灯和新版绿色两个版本的完整回执及对应摘要，重交独立 QA。

已发生的开发顺序无法事后改写：原修复前尝试的环境阻塞继续保留，新证据应明确标为“修复后隔离重建旧版网络复现”，不得伪称当时已成功。本轮未提交/推送/评论 Issue/创建 PR/部署；仅新增 QA 文档与日志。质量变化：修复后真实网络/浏览器缺口已由宿主关闭，新增回归资产稳定，剩余一个旧版网络红灯证据缺口，未发现新产品失败。待补证后再判断能否走 next，由后续 Connector 回写 Issue 和草稿 PR。

---

# Model Relay v0.1 独立 QA（历史记录，与 Issue #8 结论分开）

> **当前结论（2026-10-05，seq 16 复验）：Go。** QA-GAP-01/02 已由匹配当前 32 文件摘要的宿主完整门禁与真实网页旅程关闭；详见第 9 节。以下第 1～8 节是 seq 13 的历史 No-Go，保留失败与缺口证据，不代表当前结论。

> 日期：2026-10-05
> 模式：首次独立全量验收；包含此前修复的原层级复核，不重建需求或设计。
> 结论：**Blocked / No-Go（原强制 AC 的测试证据未完整覆盖）**。
> 路由：`development`，补充明确缺口后重新经过项目 `tests` Connector 与独立 QA。

## 1. 结论与版本

seq 13 的真实 `make verify` **exit 0** 有效，并与当前代码摘要完全匹配。真实 HTTP/SSE、CLI/停止/恢复、Go race、构建以及 1280px/390px 连续浏览器旅程均有成功回执；不依赖开发者“已通过”的说法。独立 QA 的非监听自查也实际成功。但 AC-04 的配置变更/完整多别名重启闭环和 AC-15 的网页记录秘密排除断言尚未被当前运行资产验证，不能据退出码将其写成 Pass。没有新增已复现的产品功能缺陷；以下是与原承诺直接相关的验收资产缺口，不是沙箱权限阻塞。

- 仓库：`big91987/model-relay`；分支：`workflow/c1089fdc410a9dcce0725c30ad190df4`。
- HEAD：`fdcb16cb3426d403327da727ae2a8f95c001f1af`。
- 当前产品/测试/构建/基线 31 文件摘要：`97efa7f179cbc32edd3b99e59ee2b9902f12fce46866b9d3847a196f7abcef38`；与 seq 13 相同。
- QA 开始时未提交差异：`.gitignore`、`README.md` 修改（合计 +78/-8）；`Makefile`、`cmd/`、`docs/workflow/`、`go.mod`、`go.sum`、`internal/`、`package-lock.json`、`package.json`、`tests/` 未跟踪，展开共 60 个未跟踪文件。实现尚未提交，不能仅以 HEAD 标识实现版本。
- 工具链：Go 1.25.6、CGO/C 编译器、SQLite driver v1.14.52、x/crypto v0.55.0、x/sys v0.47.0、Playwright 1.59.1。Node/npm 仅测试使用；未取得 Node/Chromium 精确版本回执，不编造版本号。设计选型调整已在 delivery/README 记录。
- 环境：宿主 Connector 上真实临时服务；本轮 Agent 沙箱只补跑不监听测试。未触及用户数据、生产上游或付费供应商。无生产部署制品/上线验收承诺。

## 2. 基线与证据责任

基线采用 `task.md`、`requirements.md` AC-01～19、`design.md` 的当前 M1/v0.1 契约及 README 正式运维入口；M2～M4 不适用。`delivery.md` 只作修订索引，不能反向改写 AC。

| 证据 | 执行者 | 本轮独立复核范围 |
|---|---|---|
| 平台 seq 13，`make verify` exit 0 | 项目测试 Connector（宿主） | 完整输出、所有阶段、源码摘要、测试预期、运行真实服务与连续网页路径 |
| `qa-host-seq13-review.md` | 独立 QA | 原回执摘录和定位；不是第二份执行回执 |
| `qa-sandbox-checks.log`，exit 0 | 独立 QA（Agent 沙箱） | 摘要、格式、模块校验、vet、4 项前端异步测试、7 项非监听 race 用例、build |
| seq 7/9/11 失败与 development 失败日志 | 原 Connector/研发 | 保留失败，不以旧版成功替代当前版本；原层级修复由 seq 13 复核 |

沙箱自查命令依次为 `go run ./tests/evidence`、`make format-check`、`go mod verify`、`make vet`、`make frontend-test`、`go test -race -count=1 -v ./internal/relay -run '^(TestInProcessContract|TestHeaderBoundaryAndFailureContract|TestDestinationAndProtocol|TestUsageSnapshots|TestInstallationRecovery|TestInterruptedAndVersion|TestSourceAndLoginLimits)$'`、`make build`。使用项目缓存、`GOTOOLCHAIN=local`、`CGO_ENABLED=1`。日志临时链接对象路径已脱敏，原 warning 内容保留。

## 3. 原 AC 逐项矩阵

Pass 表示本轮接受匹配摘要的真实运行回执并检查相应断言；不表示独立 QA 在沙箱重跑了监听或 Chromium，也不表示真实供应商兼容。

| AC | 角色/用户结果及证据 | 结果 |
|---|---|---|
| AC-01 | 空安装、拒绝重复 init、不覆盖身份、CLI loopback/health/未配置；`TestCLIInstallationCommands`、`TestInstallationRecovery` 与桌面首次登录 | Pass |
| AC-02 | 错误/正确登录、未登录管理、Cookie 与应用 Bearer 隔离、退出失效；HTTPManagement、InProcess、CLI、两屏实际浏览器 | Pass |
| AC-03 | 缺失/null/外站 Origin 拒绝、CSRF、10 次错误登录后 429；SourceAndLoginLimits、HTTPManagement；独立沙箱复跑 | Pass |
| AC-04 | 网页两个别名、保存/拒绝保存/刷新、HTTP 认证映射、重复别名、停用模型与上游拒绝、模型重新启用已通过。真实修改成不同配置后的转发、上游重新启用、两个别名重启均缺断言，见 QA-GAP-01 | Blocked |
| AC-05 | URL 协议/私网/混合 DNS 与 Dial 重新解析拒绝、目的地字段拒绝、受控测试例外；TimeoutRedirectAndNetwork 目标计数为 0、HTTPManagement 改地址空 token 拒绝；security.go 补充审查禁止代理/重定向 | Pass |
| AC-06 | 管理查询/刷新不回显、同地址空 token 保留且调用成功、改址空 token 拒绝、DB/WAL 哨兵扫描；security/store AES-GCM（nonce/AAD）代码复核，非明文/可逆编码 | Pass |
| AC-07 | 两把密钥网页一次展示/关闭/刷新，停用一把拒绝另一把可调用；HTTP 鉴权、DB/WAL 无完整秘密，存储摘要及提示标识复核 | Pass |
| AC-08 | 受控上游真实 HTTP 模型替换、普通内容返回；未知/停用模型拒绝且计数不增加；禁用上游模型列表为空；网页普通结果可见 | Pass |
| AC-09 | TransparentParametersAndLocalReject 验证参数/tools/tool_choice/大整数透明、非法/过大/未知模型不转发；DestinationAndProtocol 补充请求字段边界，protocol/proxy 补充审查响应透传 | Pass |
| AC-10 | SSEBarrierAndUsage 首包在释放屏障前可读，后续工具增量与 DONE 可见、部分 usage 正确；实际 Go 服务两屏 SSE 旅程通过。SSE HTTP 与网页证据分别接受 | Pass |
| AC-11 | Cancellation 普通及 SSE 上游 context 取消、有截止时间、次数/终态断言；网页取消后状态及记录 canceled。未以页面状态替代上游取消证据 | Pass |
| AC-12 | UpstreamFailures 与 TimeoutRedirectAndNetwork 含认证/429/500/类型/JSON/流错误/EOF/扩展/连接/超时/redirect；计数不重放、无 DONE、终态分类及敏感诊断排除；HeaderBoundary 补充错误事件 | Pass |
| AC-13 | HTTP 记录 ID/密钥归属/耗时/用量和失败终态；网页实际查询 success/canceled，表格字段映射复核；转发前 rejected、无身份不编造记录、应用不能枚举管理 API | Pass |
| AC-14 | UsageSnapshots、DestinationAndProtocol、InProcess、SSEBarrierAndUsage、stream_eof：0/null/缺字段/部分/无效 usage/后续空 snapshot；独立复跑。页面 `?? '未知'` 口径代码补强，不称额外 UI 用量专测通过 | Pass |
| AC-15 | 管理响应、DB/WAL、敏感上游错误排除与实际输出审阅通过；浏览器 storage 空、Cookie 隔离通过。网页调用记录的秘密/对话排除没有运行断言，见 QA-GAP-02 | Blocked |
| AC-16 | HTTPRestartAndBackup 中成功/rejected 历史、启用/停用密钥、失效旧会话、恢复路由实际调用；CLI SIGTERM/重启/端口占用/非 loopback 拒绝；异常进程终止与中断记录另有真实测试 | Pass |
| AC-17 | CLI init/backup/restore 与 HTTP 恢复后登录/调用/停用/记录；InstallationRecovery 运行中备份/错误 key/损坏备份/覆盖拒绝及 0600；InterruptedAndVersion 未来 schema 拒绝，README 明示独立 key 保管 | Pass |
| AC-18 | 同一会话真实 Go 服务网页登录→配置→模型→两密钥→普通/SSE/取消→记录→停用→另一把调用→退出；1280px 与 390px PASS；错误保存不显示成功、加载屏障与存储断言，无静态原型替代 | Pass |
| AC-19 | seq 13 固定 `make verify` 全阶段真实 exit 0，与当前摘要相同；锁定依赖、隔离临时目录/随机端口、串行 fail-fast；此前失败真实 exit 2 保留。此项 Pass 不等于全部 AC 已覆盖 | Pass |

## 4. 连续旅程与最终事实

| 旅程 | 起点、同一角色/会话及步骤 | 最终事实/边界 | 结果 |
|---|---|---|---|
| UI-E2E-1280 | 脚本仅用 Go Init 准备空安装管理员与隔离上游；同一管理员浏览器从错误登录到正确登录、首次保存、错误保存恢复、两别名/两密钥、普通/SSE/取消、记录、停用/另一把调用、退出 | 上游检查真实模型及认证；记录 success/canceled；停用调用 unauthorized，另一把完成；退出查询 401 | Pass |
| UI-E2E-390 | 新浏览器 context/同一管理员；同一 Go 实例已有桌面配置，使用新的 alias/key 名重复完整操作 | 窄屏无横向溢出/JS pageerror，持久存储无危险凭据；不是第二次空安装专测 | Pass |
| OPS/HTTP | 独立临时目录；正式 CLI 初始化/停止/备份/恢复/重启及真实 HTTP 请求；故障进程测试隔离执行 | 原密钥状态/登录/路由/历史与异常终态验证；不改用户数据库 | Pass |
| 配置变更生命周期补全 | 现有网页和 HTTP 用例没有把实际配置换成不同值再核对新转发；上游停用后也未走重新启用调用 | 不能用保存成功、同值重新提交或模型重新启用冒充上游恢复 | Not Run |
| 网页记录秘密专测 | 当前 record-list 只断言状态文字，没有断言哨兵文本/完整秘密不在记录中 | API/DB 通过和 storage 空不能冒充这个 DOM 结果已测 | Not Run |

浏览器 `page.route` 仅延迟配置 GET，后续 `route.continue` 请求真实 Go API，没有伪造成功响应；管理配置/签发/停用步骤全部由页面完成。浏览器退出后 `context.request.get` 只检查注销事实，不是隐蔽补齐用户步骤。无截图/trace 产物，使用真实运行回执，不将代码阅读写为 UI 执行。

## 5. 正式整改交接

两个项目均为**强制 AC 证据缺口**，不是已复现的产品 P0/P1 故障；未擅自改变原 AC 的 P0 级别，也不因没有复现功能故障而放行。

### QA-GAP-01：AC-04 的动态配置与持久化覆盖不足

定位：`internal/relay/relay_test.go:365` 改地址但空 token 被拒绝；`:367` 只用同地址留空 token 成功。`:370`/`:386` 模型停用/启用均为原 `provider-model`；`:389` 禁用上游后没有重新启用。`internal/relay/persistence_test.go:13` 重启/恢复 fixture 只有 alpha，模型列表断言也是单个。`tests/browser/journey.mjs:74` 创建两个别名但只刷新，不停止/重启服务。这些成功用例不验证变更后的新配置。

研发补充最小测试，不要求新增产品功能：

1. 通过真实管理 HTTP（模型编辑至少补网页操作）将 alias 映射改为**不同**真实模型，下一次生成由受控上游断言新模型；旧值不能被继续使用。
2. 用受控隔离目的地/凭据验证合法配置更新后新请求使用新地址/新 token，旧目的地计数不增加；保留“改址无新 token 拒绝且配置不变”。不得使用真实供应商或放开产品私网限制。
3. 上游禁用后调用拒绝、计数不变；重新启用后模型列表与普通调用恢复、上游计数增加一次。模型启停和密钥启停的已有回归保留。
4. 网页创建至少两个不同 alias 后，刷新查询并正常停止/重启，在真实管理查询及至少一次恢复网页读取中确认两个 alias/映射都保留；复用隔离 fixture，不手工 SQL 补齐状态。

### QA-GAP-02：AC-15 缺少网页记录的秘密排除断言

定位：`tests/browser/journey.mjs:123`～记录刷新只验证 success/canceled；`:141` 之后验证持久存储，不检查 record-list 的敏感内容。HTTPManagement 的管理响应/DB/WAL 排除断言有效，但不是实际网页记录检查。

在既有真实两屏连续旅程查看记录时，使用受控哨兵输入/回答/上游凭据及本轮签发密钥，断言 **record-list/记录页面结果** 不含完整秘密或对话；不要对正在用于输入/一次展示/调用结果的整页错误地断言“不含秘密”。覆盖敏感上游失败之后的记录，确认网页显示分类而非上游诊断正文，并审阅/扫描捕获的服务日志。断言失败只能输出对象类别/位置，不能把秘密本体写入失败回执；无需保存敏感截图或引入大型日志平台。

补测后更新 delivery 和摘要，重新完整运行 **`make verify`**，由宿主 Connector 保存新摘要的真实输出，再交独立 QA。旧 seq 13 只证明旧摘要，不代表修订后的完整门禁。若新增断言揭示实际缺陷，由研发修复并保留红转绿证据。无需需求/设计重做或用户扩大沙箱权限。

## 6. 失败保留与质量趋势

- seq 7：宿主浏览器 timeout 原失败，`development-host-browser-timeout.log` 保留诊断不足；seq 9 无 PUT 的首次保存竞态，`development-host-browser-save-failure.log` 保留。
- R1：表单加载完成后再展示的产品修复及回归，在 seq 13 的真实两屏首次配置/GET 屏障得到原网页层验证，当前 Closed。
- seq 11：恢复导航后即时读取隐藏旧字段的 harness AssertionError，`development-host-browser-recovery-failure.log` 保留。R2 改用真实 GET 完成屏障后保留原地址/凭据空断言；seq 13 两屏恢复路径成功，当前 Closed。
- 空 usage 覆盖已知 snapshot 的修复：`development-usage-before.log`/`development-usage-after.log` 保留；本轮匹配版本 InProcess/Usage 用例也通过，当前 Closed。
- 本轮未发现旧问题复发、未新增已复现产品功能缺陷；新增两个原 AC 的覆盖遗漏。自动化从 UI 超时/诊断/竞态/同步失败进展到完整门禁绿灯，但尚缺生命周期与网页记录安全的专门断言，不报告“全部验收通过”。

## 7. 未测、环境与阶段边界

- 真实供应商联调：**Not Run**，没有凭据/供应商响应/付费授权；不作为本次受控行为验收的新增阻断，不宣称任何供应商实际兼容性或可用性。
- 本轮 Agent 没有重跑监听服务或 Chromium；历史 `bind operation not permitted`/Chromium Mach 拒绝只属于环境限制，不是产品 Fail。seq 13 宿主证据可等价用于已覆盖场景，无需再向用户索要整体提权。
- 压测、公网生产发布、多上游、HA、计费、M2～M4 不在本次范围；不补造这些阶段文档或门禁。
- 测试使用临时目录和随机端口，runner 有 defer/cleanup 关闭服务、浏览器并删除目录；本轮非监听测试正常退出。宿主独立系统进程盘点未执行，不声称额外观察过宿主全局清理；没有共享生产资源需要恢复。
- 本轮只新增 QA 报告/证据并规范化自查日志临时路径；产品和测试未改，未 Git 提交/分支/推送/创建 PR。

## 8. Go / No-Go

固定工程门禁及已有真实连续旅程为 **Pass**；AC-04、AC-15 尚为 **Blocked**，其明确子场景为 **Not Run**。整体 **No-Go**，正式回 `development` 补测试、经 `tests` 后再 QA；不走 `next/report`。不把证据缺口当产品断言失败，不把能力不足或未测供应商冒充通过。

## 9. seq 16 独立复验与当前结论

### 9.1 模式、版本与责任

2026-10-05 执行短轮复验：恢复上次全量 QA 与整改要求，重点复核 QA-GAP-01/02，公共回归采用最新宿主完整 `make verify`。不重建已确认需求/设计，不修改产品或测试。基线仍为原 AC-01～19；以前的 No-Go 及开发红转绿日志全部保留。

- 分支 `workflow/c1089fdc410a9dcce0725c30ad190df4`；HEAD `fdcb16cb3426d403327da727ae2a8f95c001f1af`。
- 当前 32 文件 `working_tree_sha256`：`fbc07625a4d9819bba5aa809de5dfea636450cf42489d7d7c72ab66ccba06efe`。独立 QA 重新计算，与平台 seq 16 完全一致；不能仅以相同 HEAD 接受未提交实现。
- QA 入场差异：`.gitignore`、`README.md` 修改，合计 +80/-8；`Makefile`、`cmd/`、`docs/workflow/`、`go.mod`、`go.sum`、`internal/`、`package-lock.json`、`package.json`、`tests/` 未跟踪，展开 71 文件。此计数在本轮新增报告之前采集。Git 提交/发布仍未执行。
- 相对上轮的验收资产变更：新增 `internal/relay/lifecycle_test.go`；浏览器 runner 新增显式测试生命周期控制与服务日志捕获；旅程新增不同映射编辑、多映射重启核验、记录/日志秘密排除及敏感失败计数；README 更新测试说明。原门禁及产品实现未因补缺口放宽。
- 依赖仍为 Go 1.25.6、CGO、go-sqlite3 v1.14.52、x/crypto v0.55.0、x/sys v0.47.0、Playwright 1.59.1。宿主 Node/Chromium 精确版本未在回执报告，不臆造。

| 证据 | 执行者 | 独立 QA 复核 |
|---|---|---|
| 平台 seq 16 `make verify` exit 0 | 项目测试 Connector，宿主 | 完整阶段输出、32 文件摘要、所有新增用例及原层级网页旅程；不是开发自述 |
| `qa-host-seq16-review.md` | 独立 QA | 宿主原回执的索引、摘录、源码断言与证据边界；不是重新执行宿主测试 |
| `qa-sandbox-recheck.log` | 独立 QA，Agent 沙箱 | 摘要、格式、module、vet、4 前端异步回归、7 非监听 race 用例及构建 |

独立沙箱自查实际 exit 0，7 项非监听 race 用例包耗时 26.996s，前端 4/4 通过。沿用第 2 节列出的非监听命令集合，日志记录精确命令；临时链接路径规范化。报告落盘后再次计算摘要仍一致，`git diff --check` 成功。本轮未重跑沙箱 `make verify` 或 Chromium；完整入口的执行证据明确来自 seq 16 宿主，不混称 QA 自己运行了 HTTP/UI。

### 9.2 QA-GAP-01 关闭：配置生命周期

`TestHTTPConfigurationChangesAndReenable` 通过真实 HTTP 执行：初始调用、改址无新凭据拒绝且仍调用旧配置、alpha 改成不同真实模型、合法替换 URL 路径/token 后调用新目的地、禁用上游无调用消耗且模型列表为空、重新启用后列表和实际调用恢复。测试上游检查模型/认证，计数证明旧目标不增加、新目标只增加一次。替换目的地为同一隔离 authority 下的不同 base path，而非真实供应商或第二个公网主机；满足新配置生效的子场景，不宣称额外跨主机联调。

1280px 与 390px 真网页旅程分别创建 alpha/beta 的不同映射，通过页面编辑 alpha 为 `edited-controlled-model`。刷新确认后显式调用测试控制入口正常 Shutdown、Store.Close/Open、重新监听；原管理会话返回 401，管理员在网页重新登录。真实管理查询及恢复后的网页都断言两个映射正确，上游地址/启用状态保留、凭据空但已配置。后续页面普通/SSE 调用由受控上游检查编辑后的真实模型和认证。

测试控制 `/restart` 只承担已声明的运维夹具动作，不在产品路由中，也不替代页面创建/编辑配置；没有 SQL 写入或后台补对象。该网页测试验证服务/存储正常关闭重开；完整 CLI/真实进程停止、异常退出、备份恢复仍由原 Go 用例在同轮执行。**QA-GAP-01 Closed；AC-04 Pass。**

### 9.3 QA-GAP-02 关闭：记录秘密边界

两屏旅程均从页面发送含哨兵输入的普通/SSE/取消请求及敏感上游失败。受控上游实际返回含测试凭据、输入/回答及诊断的拒绝；网页错误呈现 `upstream_rejected`，记录区当前 alias 恰有一个该失败行。上游累积失败计数为桌面 1、窄屏 2，证明各屏真实触发一次而非仅借用旧记录。

在 `#record-list` DOM 中排除完整管理员密码、上游 token、该屏两把应用密钥、提示词、回答、取消/普通输入和敏感诊断；停用/另一把调用之后再次检查。只检查记录区域，不把输入框、一次密钥展示或调用输出的合法内容误判成记录泄漏。实际 Go 服务默认 logger 的输出被线程安全捕获，测试控制 `/audit-logs` 扫描同一秘密集合且返回 204；不是仅扫描模拟文本或源代码。失败断言只输出分类/位置，不输出秘密本体。

原管理响应、DB/WAL 排除及浏览器 storage/cookie 断言未删除，并在同轮完整入口继续通过。日志检查针对该真实测试服务捕获的默认 logger，不声称检查宿主所有进程或用户生产日志。**QA-GAP-02 Closed；AC-15 Pass。**

### 9.4 当前 AC 与旅程结果

| 范围 | 当前证据与结果 |
|---|---|
| AC-01～03 | 原初始化、鉴权隔离、来源与限速用例及网页登录/退出，本轮 seq 16 完整回归 Pass |
| AC-04 | 新配置 HTTP 生命周期与两屏不同映射编辑/重启核验，加原停用/拒绝用例，Pass |
| AC-05～09 | 原地址/重定向/凭据/密钥/普通转发/参数边界完整回归，Pass |
| AC-10～14 | 原 HTTP 首包屏障、工具增量、取消、故障、超时、记录及 nullable usage 回归，Pass；网页证据不替代 HTTP 时序证据 |
| AC-15 | 新真实记录 DOM 和默认服务日志扫描、敏感失败计数，加原管理/持久数据/浏览器存储边界，Pass |
| AC-16～17 | 原 CLI、正常/异常重启、备份恢复及版本材料失败回归，Pass |
| AC-18 | 1280/390px 同角色连续网页旅程完整通过，含恢复后重新登录、配置/模型/密钥/普通/SSE/取消/记录/停用/退出，Pass |
| AC-19 | seq 16 完整 `make verify` exit 0，摘要与当前代码一致，Pass |

桌面从空安装开始，窄屏新 context 使用同一隔离安装的已有配置，分别创建新 alias/key；不宣称窄屏另有一次空安装。浏览器使用真实 Go handler 与 SQLite，受控上游只有测试显式允许的 authority；请求屏障使用 route.continue 而非伪造产品 API 响应。未生成截图或 trace，本次证据为真实连续旅程回执与对应断言审阅，不补造截图。

### 9.5 趋势、剩余边界与交接

本轮关闭两个原 AC 验收资产缺口；没有新增已复现产品缺陷，旧 R1/R2/usage 回归未见复发。新增生命周期和记录安全用例已进入固定入口，原全量测试全部重跑成功。链接器 LC_DYSYMTAB warning 未导致失败，`storage_write_failed` 属于故意注入用例，不作为产品失败。历史沙箱监听/Chromium 拒绝保留，独立 QA 不申请提权，也不用静态 browser.check 替代后端 UI。

当前 **Go（M1/v0.1 受控验收）**，无未关闭 P0/P1 或原强制 AC 缺口，走 `next → report`。这不是生产发布授权，也不是已交付草稿 PR；Git/草稿 PR 仍由后续 Connector 完成。真实供应商联调 **Not Run**，没有真实凭据/响应或付费调用授权；压测、公网发布、HA、M2～M4 不适用。宿主全局清理未另行盘点；runner 的关闭/临时目录删除机制已审阅，沙箱自查正常退出，未触及共享用户数据。后续可直接复用完整 `make verify`，提交后的版本若更改产品或验收资产必须重新匹配摘要及复验。
