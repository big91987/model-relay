# 研发交付记录

## Issue 9 运行版本展示（2026-10-05）

基线：本次 task.md、requirements.md AC-01～08、design.md 及设计交接。用户澄清已完成，不再重复确认。下方 v0.1 记录及旧日志仅保留历史失败，不作为本次通过证据。

- VERSION-01：概览增加只读版本文本；独立健康请求、5 秒全请求超时、导航/退出失效。无生产后端变更、供应商调用或部署。
- VERSION-02：VM 回归覆盖原值、异常、非阻塞、重入、迟到响应与 body 超时；真实 Go 服务浏览器覆盖登录边界、刷新恢复、文本安全及双视口；Go HTTP 检查健康版本契约。
- VERSION-03：运行原固定 `make verify`（保留全部既有门禁），记录代码摘要及真实失败；若本机监听受限，交固定项目测试 Connector 实际执行，不标通过。

关键路径：VERSION-01 → VERSION-02 → VERSION-03 → 项目测试 → 独立 QA。负责人为研发实现节点；VERSION-01/02 实现与测试资产就绪，VERSION-03 本机验证已执行但未通过，整体 `Ready for host verification / In Review`。不是功能验收通过或 QA 放行，不自行提交、推送。

首轮 `development-version-verify-initial.log`：`make verify` exit 2，前端 23/23 通过；存在两个不同失败来源：沙箱拒绝真实 HTTP/CLI 监听，以及新增健康契约测试的 Host 夹具未设置导致 403。后者已修正为匹配 Authority 的请求并使用隔离真实 Store（健康接口会 Ping DB），不能将夹具失败归为沙箱。初轮摘要仅代表修复前代码；后续以新回执为准。

### 实现与 AC 追踪

- AC-01/02：概览标题下的 `#runtime-version` 展示原始文本，含 `dev`；未登录不请求。不修改品牌、启动版本链或生产健康接口。
- AC-03/04：独立 `/healthz` 请求，不复用管理 api/调用测试控制器；`cache:no-store`；未知只更新版本文本，既有概览与导航不等待版本 Promise。
- AC-05/06：每次概览进入重新加载；5 秒覆盖 headers/body；请求身份、导航快照、active 与登录状态共同约束写回。切页、退出开始、真实管理 401 清理，旧 success/catch/finally 不污染新请求。
- AC-07：仅 `textContent` 写入，`pre-wrap` 与 `overflow-wrap:anywhere`；不写存储、配置或调用供应商。README 增补运行/降级/重试说明。
- VERSION-02：`tests/browser/ui.test.mjs` 24 项包含原 4 项，新增原值、12 类异常、非阻塞、重入、超时/body、晚到成功/失败、旧 finally、退出/重登和真实管理 401；`internal/relay/version_test.go` 增加进程内健康契约及真实 HTTP 回归；`tests/browser/journey.mjs` 在原真实 Go 服务旅程插入 1280/390px 版本检查（真实健康一致性、刷新、401、重入恢复、挂起时模型读取、文本安全/长文本布局及退出边界）。浏览器测试服务明确注入 `dev`，重启沿用，避免测试夹具空值冒充实际 CLI 默认值。
- 所有回归均已纳入原 `make verify`：`frontend-test`、`go test -race ./...`、`browser`；Makefile 原有完整门禁保持不变，无另建或跳过关键断言。无需静态 browser.check：本功能依赖真实 Go 后端，静态服务不构成验收。

### 最终代码与本机验证

最终 HEAD：`9b6994ed80de2bdba79ff267419150dccaac1360`；33 文件工作树 SHA256：`f96017d3daf52e8372e1ef106397bfb31224e049010bb53a6cd033d69f27d414`。沿用 `tests/evidence` 摘要口径（排除交付回执，避免自引用），不代表提交已产生。初轮摘要 `858a0a646c7d835f1656c7ec20140af8a75be9e95314cfbbc8df3eafb20183f1` 对应夹具修正及最后一项回归之前，不能混用。

| 回执 | 实际结果与限制 |
|---|---|
| `development-version-verify.log` | 最终 `make verify` exit **2**：format/modules/vet 与前端 **24/24** 通过；健康契约 **3/3** 通过。Go race/HTTP/CLI 门禁在真实端口监听处 `bind: operation not permitted` 失败（包括新 `TestHTTPHealthVersion`）；串行 build/browser 未到达。最终未再出现健康 403 夹具断言失败。 |
| `development-version-targeted.log` | `go test -race -count=1 -v ./internal/relay -run '^TestHealthVersionContract$'` exit **0**，3/3；独立 `make build` exit **0**；journey JS 语法检查 exit **0**。不替代真实 HTTP 或浏览器通过。 |
| `development-version-browser.log` | 独立 `make browser` exit **2**：npm 锁定依赖安装连接本机代理被 EPERM 拒绝，未启动 Go/Chromium；无浏览器断言执行。 |
| `development-version-browser-runner.log` | 直接运行同一真实服务 runner `go run ./tests/browser` exit **1**：测试上游 `127.0.0.1:0` 监听被拒绝，未启动 Chromium。没有用静态服务代替。 |

原失败和 linker LC_DYSYMTAB 警告保留。新回执仅将本机绝对路径脱敏为 `[toolchain]`、`[temp]`、`[npm-runtime]` 或相对路径，未隐藏失败；缓存/bin/node_modules 仍受已有 .gitignore 排除，无凭据或个人会话入库。

### 交接与简化审查

按 Skill 检查上游/任务/关键路径/Harness/公共接续：局部功能不拆组织级文档或新增平台模型，任务与 AC 可从本节、requirements/design 和共享入口恢复；无范围或结构偏差需要返回需求/设计，无需用户裁决。验证门槛仍 **Blocked by Validation**，资产 **Ready for Dispatch** 到固定宿主项目测试，而非版本交付 Accepted。

下一步 `next → tests`：使用已配置宿主 Connector 完整运行 **make verify**，核对上述最终摘要；必须实际执行新增 HTTP 及 1280/390px 浏览器版本旅程与所有原测试。重点待核验：健康/UI 原值一致、登录边界、刷新/重入恢复、健康挂起时真实模型页读取、文本安全/长值布局、既有后台及恢复流程不回归。若真实断言失败返回研发；项目测试成功后再交独立 QA 核对 AC-01～08。现有 qa/pr 是历史材料，本节点不覆盖或自行 QA 放行；Issue Hook 和草稿 PR Connector 接续，不部署。

## R3 独立 QA 验收资产补齐计划（2026-10-05，seq 14）

恢复依据：task/requirements/design 基线不变；seq 13 对摘要 `97efa7f179cbc32edd3b99e59ee2b9902f12fce46866b9d3847a196f7abcef38` 的宿主 make verify exit 0 有效。独立 QA 在 qa.md 标出 QA-GAP-01（AC-04 配置变更/启停/两个网页别名重启）及 QA-GAP-02（AC-15 实际网页记录及敏感失败/服务日志）。这是原 AC 覆盖遗漏，不是已复现产品缺陷；保留旧失败和 QA No-Go，不重做需求/设计、不提前 QA 放行。

任务与验证计划：
- R3-A：真实 HTTP 补不同模型编辑、新地址/新 token、旧目标计数不增加、无新 token 拒绝、禁用无消耗/重新启用恢复；接入现有 go test。
- R3-B：原 1280/390px 真实服务旅程用网页编辑映射；网页创建两个不同映射后正常 Shutdown/关闭存储/重新 Open 与启动服务，重新登录，管理查询及恢复网页双重核对配置，禁止 SQL 补状态。
- R3-C：原旅程加入哨兵输入/回答/敏感上游错误，限定 record-list 检查完整秘密和对话排除、分类保留；捕获服务日志检查同一秘密集，失败仅报对象类别不输出内容。
- R3-D：保留全部原门禁/断言，真实执行 make verify、独立 browser（若串行入口因监听失败未到达）、静态检查与 build，更新新摘要及真实回执。沙箱拒绝监听时按已配置 next→tests 宿主运行完整入口；任一新真实断言失败仍回研发，不删断言或请求提权。

### R3 实施、最终回执与交接（容量中断后续接）

- QA-GAP-01 补测资产已完成：新增 `internal/relay/lifecycle_test.go` 的真实 HTTP 测试，经管理 API 编辑为不同模型并立即生成核对；改 URL 无新 token 拒绝后旧配置仍有效；合法更新 URL/token 后只访问新 URL、使用新认证，旧 URL 计数不增加；上游禁用后拒绝且不消耗，模型列表隐藏，重新启用后模型列表与普通调用恢复且仅增加一次调用。两个受控 URL 使用同一隔离 authority 的不同路径，逐目标计数，不扩大产品目的地址策略。
- 两屏原旅程已加入网页映射编辑，以及两个不同映射的刷新/重启恢复核验。runner 正常 Shutdown HTTP 服务、关闭 Store，再用原安装和独立 key-file 重新 Open/启动，浏览器旧会话失效，重新登录后真实管理查询和恢复网页都核对两条映射、上游地址/启用状态及凭据已配置但不回显。此次是服务/存储重启，独立进程 SIGTERM/崩溃恢复仍由既有 CLI/Go 用例验证，不冒充新 OS 进程测试。
- QA-GAP-02 补测资产已完成：两屏普通/SSE/取消后补敏感上游失败；对应本屏 alias 的记录行必须显示 `upstream_rejected`，受控上游失败计数必须每屏恰好一次。限定 `record-list` 检查完整管理员/上游/应用秘密、哨兵输入/回答及上游诊断排除，最终其他调用后重复检查。实际服务日志捕获在测试 runner 内存中，通过独立测试控制服务核查同一秘密集，不保存正文或秘密日志。该控制服务仅在测试中存在，不是产品端点、不直接改数据库；泄漏断言不输出秘密本体。
- 原 AC、产品实现和所有既有门禁保留；新增 HTTP 测试由现有 `go test -race -count=1 -v ./...` 自动纳入，新增两屏步骤由原 `make browser` 执行，全部属于固定 `make verify`。README 更新了运行/验证边界；无新需求或架构方向需要退回。
- 最终 `development-qa-gaps-static.log` **exit 0**：JS 语法、format、vet、4/4 前端异步回归、独立 build、源码摘要。当前 HEAD `fdcb16cb3426d403327da727ae2a8f95c001f1af`，32 文件工作树摘要 **`fbc07625a4d9819bba5aa809de5dfea636450cf42489d7d7c72ab66ccba06efe`**；续接后重新计算得到相同摘要，未修改补测源码。
- 最终 `development-qa-gaps-verify.log` **exit 2**：真实执行完整入口，format/module/vet/4 前端回归及非监听 Go/race 用例通过；CLI、原 HTTP 用例及新增 `TestHTTPConfigurationChangesAndReenable` 在 `127.0.0.1:0 bind: operation not permitted` 处失败。internal/relay 总耗时 55.592s，串行 build/browser 未运行；保留完整失败和 LC_DYSYMTAB 警告。此为沙箱环境限制，新增真实 HTTP 断言尚未执行，不标产品 Fail 或整体 Pass。
- 最终 `development-qa-gaps-browser.log` **exit 2**：锁定 npm 安装成功，实际上游监听被沙箱拒绝，未启动 Chromium；新增重启与记录/日志断言均未实际执行。独立 build 的成功不替代这些断言。
- 保留 `development-qa-gaps-static-attempt.log` **exit 1**：首次直接运行摘要程序未继承 Makefile 的项目内 GOCACHE，出现 open operation not permitted；显式沿用项目内缓存后成功，不是产品回归。`*-initial.log` 是较早摘要 `ec890408d79260760cf6b04a0d50d578104ddf74ca6b2863f2962863a91b299b` 的初轮证据；之后补上本屏敏感错误记录/真实上游失败计数核对并完整重跑，最终日志才对应上述 fbc 摘要。初轮证据不能与最终代码混用；旧 seq 13 成功也只证明旧摘要。

R3 状态：**补测资产就绪 / Ready for host verification**，不是 AC-04/15 已通过或独立 QA 放行。按用户续接指令正式 `next → tests`，固定宿主执行完整 `make verify` 并保留新摘要回执；任何真实断言失败返回研发，成功后由 QA 复验 QA-GAP-01/02 及原全部基线。不请求提权、不删除/跳过断言、不提交/推送/创建 PR；供应商联调仍 **Not Run**。QA 原 No-Go 及历次产品/harness失败证据保持原样。

## R2 浏览器恢复路径同步计划（2026-10-05，seq 11）

最新宿主 `make verify` exit 2，工作树摘要 `d4b105f594595fbe96e1b7b054a89a971d136469c9b420d5310e8676672b065a`：frontend-test 3/3、全部 Go/race/真实 HTTP/SSE/CLI/恢复及 build 通过；1280px browser 在 journey.mjs:67 的上游地址恢复断言失败。首次保存 PUT 200、非法地址 PUT 400 已实际发生，先前保存竞态不再是此次失败位置。完整回执保留于平台 seq 11，本轮保留脱敏摘录，不覆盖旧失败。

计划：用实际 app.js 的可控异步响应复现「拒绝保存后重新导航、GET 完成前立即 inputValue 读到旧输入」。区分隐藏表单状态与加载完成状态，不改产品存储或放松预期值；在真实旅程恢复导航加入延迟 GET 屏障和可见等待，继续验证持久化地址与凭据不回显。保留原超时和所有后续断言，将恢复状态回归纳入现有 frontend-test。执行 make verify、独立 build/browser，若仍受监听沙箱限制则交固定宿主 tests 验证，不声称整体或独立 QA 通过。

### R2 实施与验证结果

- 复现结果：真实 app.js 在拒绝 PUT 后保留错误输入，导航时隐藏表单，GET 完成后才回填数据库配置。这符合设计的保存失败保留输入。Playwright inputValue 不等待元素可见，原 journey:67 在点击完成后立即读取隐藏表单，取到 `http://127.0.0.1:1/v1` 而非已保存地址。`development-ui-recovery-before.log` exit 1 在同一异步状态读法下确定性复现预期值差异；不把它误报为 DB 未保存。最初 fixture 没等待概览的顺序 session 请求，另存 `development-ui-recovery-fixture-attempt.log` exit 1；修正 fixture 调度后才得到上述有效复现。
- 修复范围仅测试同步：将已有延迟上游 GET 屏障提取为 loadUpstreamPanel，首次进入和拒绝保存后重新进入都断言加载中面板不可见，然后等待真实字段可见，再读取地址与凭据。原相等断言、错误提示、成功保存、取消、记录、停用、秘密隔离及桌面/390px旅程均保留；超时仍为 10 秒，没有 sleep 或放宽预期。补充恢复断言的明确诊断 checkpoint。未修改产品、数据库、管理 API 或既有 Makefile 命令。
- 状态回归在 `make verify` 已有 frontend-test 内：加载期间断言隐藏和保留被拒输入，GET 完成后断言面板可见、恢复准确地址、秘密为空和错误清除。`development-ui-recovery-after.log` exit 0，4/4 通过；这是实际 app.js 的隔离状态测试，不冒充浏览器/HTTP 验收。
- `development-ui-recovery-static.log` exit 0：JS 语法、format/vet/build 和源码标识通过。最终 HEAD `fdcb16cb3426d403327da727ae2a8f95c001f1af`，31 文件工作树摘要 `97efa7f179cbc32edd3b99e59ee2b9902f12fce46866b9d3847a196f7abcef38`，未提交；本轮不变更验收/设计基线。摘要由固定入口 tests/evidence 生成，交付日志不参与自引用。
- `development-ui-recovery-verify.log` 为真实完整入口 exit 2：format/module/vet、frontend-test 4/4、非监听 Go/race 测试通过；真实 HTTP/CLI 在 bind operation not permitted 处失败，internal/relay 54.961s。该次串行入口的 build/browser 没有运行。保留链接器 LC_DYSYMTAB 警告和全部失败，没有删门禁。
- `development-ui-recovery-browser.log` 为独立 make browser exit 2：锁定 npm 安装成功，实际 Go 测试上游监听即被沙箱拒绝；未启动 Chromium、未验证实际恢复屏障。独立 build 通过不代表完整入口通过。日志脱敏，无本机绝对路径、秘密、正文或个人会话。

R2 状态：**Harness 同步修正与回归就绪，等待宿主完整验证**。R1 首次保存在 seq 11 已走到 PUT 200，但整个 AC-18/D4 仍未通过，旧 Go 成功回执不代表新摘要整体通过。交 `next → tests` 对新摘要执行完整 make verify，1280/390px 必须在首次进入及错误恢复路径通过延迟 GET 屏障并完成原全部旅程；真实断言失败仍回研发并保留 JSON。无需求/设计方向变更，无须回退这两个阶段；不申请权限、不做独立 QA 放行、不提交/推送/创建 PR。真实供应商联调未测。

## R1 定位后的修复计划（2026-10-05，seq 9）

最新宿主固定入口 exit 2，摘要 `cd8742dcd205f8f89098a1e4da31635e1c1d50820b3a9ac37541518f30fd8ec2`：Go/race/真实 HTTP/CLI 与 build 通过；1280px 浏览器在 `#notice contains 上游已保存` 超时。诊断只有 GET upstream，没有 PUT。该失败不是沙箱限制，原回执继续保留（平台 seq 9）。

局部修复计划：先用实际 app.js 与可控异步响应复现导航提前显示表单、GET 回填覆盖用户输入的问题；阻止初次加载完成前编辑，保持成功/失败提示与恢复路径；增加不依赖端口的回归测试并接入 make verify，真实浏览器旅程加入延迟加载用例，保留所有原断言/超时。随后执行本地固定入口，端口限制仍交既有宿主 tests Connector 验证，不宣称完整 UI 或独立 QA 通过。

### 本轮实施与真实证据

- 代码根因：旧 navigate 在 await load 前显示目标表单，load('upstream') 的 GET 随后回填 URL 和 enabled。首次空配置会覆盖已经输入的 URL，浏览器 required 校验阻止 submit，因此可出现「无 PUT、无成功提示」。宿主 seq 9 的 HTTP 诊断与该竞态一致；宿主浏览器的具体调度及修复后的整个旅程仍需新回执确认。
- 用真实 app.js 加隔离 DOM/异步响应替身做不监听端口的回归：修复前两个测试失败（加载未完成时上游表单已可见），修复后 3/3 通过，覆盖等待→填写→PUT→提示/清秘密、GET 失败→再次导航恢复、旧导航不能重新显示已离开的面板。该状态测试不是实际浏览器/HTTP 证据。原失败/修复输出为 `development-ui-race-before.log`（exit 1）/`development-ui-race-after.log`（exit 0）。
- 产品修复：加载期间隐藏目标操作面板并显示「加载中」，完成/失败后再展示；导航版本保护防止旧请求完成后显示旧面板或覆盖提示。没有改变管理 API、存储、地址策略、认证或提交成功契约，也没有延长超时。
- `make verify` 增加 frontend-test，所有原格式/module/vet/race/HTTP/build/browser 命令保留。真实服务浏览器旅程在 1280/390px 初次上游 GET 加入可控屏障：请求未释放时面板不可编辑，释放后填表并保留原「上游已保存」、错误地址、未保存数据/秘密等全部断言。请求屏障有 10 秒超时，不使用 sleep 掩盖竞态；它仍由 Go 测试程序启动实际服务，不使用静态 browser.check。
- `development-host-browser-save-failure.log` 是 seq 9 的脱敏摘录，完整原回执仍在平台；不覆盖 seq 7 和旧本地失败记录。
- 新版本：HEAD `fdcb16cb3426d403327da727ae2a8f95c001f1af`；工作树摘要 `d4b105f594595fbe96e1b7b054a89a971d136469c9b420d5310e8676672b065a`，31 文件。`development-ui-race-static.log` exit 0：两份 JS 语法检查、format/vet/build 和摘要。源码未提交；旧宿主 Go 通过只对应旧摘要。
- `development-ui-race-verify.log`：实际 `make verify` exit 2；format/module/vet、新前端 3/3、非监听 Go/race 用例通过，真实 HTTP/CLI 网络测试在 `bind: operation not permitted` 失败（internal/relay 54.897s），该串行入口的 build/browser 未执行。独立 build 在上一条通过，不能因此标整体绿灯。
- `development-ui-race-browser.log`：实际 `make browser` exit 2；锁定 npm 安装成功，真实测试上游监听被沙箱拒绝，尚未启动 Chromium或执行屏障/UI 断言。日志临时和本机路径脱敏；无秘密或对话截图。

R1 状态：**修复与回归就绪，等待宿主 make verify 确认**；D4 未验收，不是独立 QA 放行。交接固定 tests Connector 必须对以上新摘要执行完整入口，验证延迟加载屏障和全部原旅程；任何真实断言失败返回研发。无需用户提权；没有 Git 提交/推送/PR 操作。真实供应商联调未测。

## 宿主浏览器失败续修计划（2026-10-05）

固定 tests Connector 第一次 `make verify` 返回 exit 2：工作树摘要 `3964b300ba3734905090d4e565fbd3fcddf31b9b3288c6d8d1d1e50faeadbf2c`；format/module/vet、全部 race Go 测试（包括真实 HTTP/SSE/CLI/恢复）和 build 通过。browser 安装依赖成功，但真实旅程报 `TimeoutError locator.waitFor: Timeout 10000ms exceeded.`，未通过。该错误是宿主实际浏览器断言失败，不是沙箱监听限制；不能据此放行。

局部任务 R1：恢复失败位置，修复确认的根因，再由相同固定入口运行全部旅程。原 harness 仅打印 error.message 第一行，丢失 Playwright locator 调用日志，无法从现有回执确定哪个断言失败；本轮先补充 viewport、最后动作、断言 selector/预期和固定 API 路径/HTTP 状态，不记录请求/响应正文、密码、密钥、地址或截图。诊断必须保持原断言与超时，不以放宽等待、删除断言代替修复。关键路径：可定位宿主复现 → 根因修复 → make verify → 独立 QA。

本地复现 `make browser`：exit 2，真实 Go 测试上游监听在 `bind: operation not permitted` 失败，尚未触及原超时。单独 Chromium 启动同样失败（MachPortRendezvousServer bootstrap_check_in Permission denied 1100）；没有扩大权限。因此本轮不能声称已复现或修复宿主 UI 超时，需已配置 Connector 运行增强诊断的固定入口；若仍失败，研发应按具体失败断言修复。

R1 本轮交接证据：原宿主失败摘录保存在 `development-host-browser-timeout.log`（明确标记非完整日志，完整回执由平台 seq 7 保留）；增强诊断只改 `tests/browser/journey.mjs`，没有修改产品行为或削弱任何验收断言。错误输出包含 viewport、checkpoint、journey 源码行号、错误类别及最近 12 条白名单 API 的方法/路径/状态；不输出错误正文，避免 AssertionError 夹带秘密。启动失败也进入统一 catch，退出仍非零。没有截图或 trace 含秘密风险。

新工作树摘要 `cd8742dcd205f8f89098a1e4da31635e1c1d50820b3a9ac37541518f30fd8ec2`（30 文件，HEAD 不变）。`development-browser-diagnostics-static.log` exit 0：JS 语法、format/vet/build、证据摘要；`development-browser-diagnostics-verify.log` exit 2：真实固定 `make verify` 再次在 Go 网络用例监听被拒处停止（internal/relay 53.273s），build/browser 未在该入口轮次执行；`development-browser-diagnostics-browser.log` exit 2：独立 browser 再次在测试上游监听被拒处停止。原宿主 Go/HTTP 通过证据仍只对应旧摘要，不能当新摘要完整门禁通过。

R1 状态 **In Progress / 等待宿主可定位复现**，不是 Done；D4 尚未验收。交接 next 是诊断复跑，不是宣布修复完成：固定 Connector 必须继续执行原 `make verify`。若 browser 失败，保留新的 JSON 定位回执并返回研发修复；若整体通过，仍须独立 QA 复核。未确认产品根因，不做猜测性产品改动，不增加权限、不创建提交/推送/PR，真实供应商联调未测。

## 基线与计划（2026-10-05）

权威输入：task.md、requirements.md AC-01～19、design.md 第 3～6 节。没有前次研发失败回执；仓库尚无实现或 Makefile。仅 M1，原创 Go 后端，不读 New API、不调用真实供应商、不执行 Git 写操作。

本节点独立执行，不建立多 Agent 分工。关键路径：持久安装与身份 → 配置/密钥 → 普通/SSE/终态记录 → 真实网页 → 固定验证。任务包：

| 包/里程碑 | 输入、依赖与边界 | 退出证据 | 状态 |
|---|---|---|---|
| D1 可恢复安装 | AC01–03/06/16–17；SQLite、Argon2、加密、排他 CLI | 初始化/来源/登录/备份失败及恢复测试 | 实现就绪；基础单元通过，宿主闭环待验 |
| D2 可授权调用 | D1；AC04–09；配置提交、别名与密钥生命周期 | HTTP 映射、停用、SSRF、秘密隔离 | 实现就绪；进程内契约通过，真实 HTTP 待验 |
| D3 可追踪流 | D2；AC10–15；一次转发、取消、有限资源、记录 fail-closed | 屏障首包、取消、故障/未知用量测试 | 实现就绪；故障契约单元通过，网络行为待验 |
| D4 可操作交付 | D1–3；AC18–19；同源真实 UI、公开运行入口 | make verify 的 Go/真实浏览器/构建回执 | 实现就绪；构建通过，浏览器旅程待验 |

验证计划：锁 Go 模块与浏览器依赖；固定 make verify 执行格式、vet、Go 单元与真实 HTTP 测试、构建、启动实际 Go 服务的浏览器连续旅程（桌面/390px）。所有数据/端口归测试所有；无凭据及正文附件。故障保留，不跳过门禁。沙箱监听被拒绝则记录原失败并交既有 host-only tests Connector，不宣称通过或要求提权。QA 独立放行不属于本节点。

工具链推荐调整：实现使用本机已安装 Go 1.25.6（go.mod 最低 1.25），不依赖未获实证的设计 Go 1.27.1 推荐；不改变产品契约。具体依赖与实测回执将在完成后追加。

依赖获取失败原证据：隔离 GOPATH 的 go mod download 退出 1，默认代理 connect operation not permitted；去代理后 proxy.golang.org DNS no such host。未申请提权。普通实现推荐调整为本机缓存的 mattn/go-sqlite3 v1.14.52，仍采用 SQLite/database/sql/WAL 事务；增加 CGO_ENABLED=1 与 C 编译器构建前提，不引入非 Go 后端或外部数据库。官方 driver README 明确 CGO 前提；x/crypto v0.55.0 和 x/sys v0.47.0 使用已有缓存并锁定。选择代价是交叉编译须有目标 C 工具链，而非 pure-Go 单二进制构建；产品运行仍为一个 Go 服务。此局部构建实现调整不改变 API/AC/存储结构。

## 已实现范围与权威入口

- `cmd/model-relay`：init/serve/backup/restore；loopback、独占锁、SIGTERM 排空、端口占用拒绝。`internal/relay/store.go` / `backup.go`：SQLite schema v1、WAL/FULL、独立 AES-GCM key-file、Argon2id、摘要密钥、started 异常恢复、停机归档/校验恢复。路径以 URI 编码，避免 data-dir 中的 ?/# 改变 SQLite 打开方式。
- `internal/relay/server.go`：管理身份/8h 会话、严格 Host/Origin/CSRF、登录 10/min + 2 并发、配置与密钥生命周期、分页只读记录。管理状态写入要求显式布尔值，null/重复键/缺字段不悄悄停用；错误固定脱敏。
- `internal/relay/security.go` / `protocol.go` / `proxy.go`：URL 与实际 Dial 的 DNS/IP 复验、指定单 authority 的测试例外、禁止代理/重定向/连接复用/生成重放；文本/工具子集透传、有限体积和超时、SSE 单事件 flush、取消传播、未知用量、终态持久后才成功/发送 DONE，写记录失败 degraded/fail-closed。
- `internal/relay/web`：Go embed 的原创同源控制台，概览、上游、模型创建/编辑/启停、密钥一次展示/复制/启停、普通/SSE/取消、记录分页、错误/忙碌/窄屏状态。不保存浏览器凭据或对话，不提供假上游默认值。
- `README.md`：公开构建/安装/停止/恢复指南、恢复材料责任、兼容与资源边界。Go/C 运行，Node 只用于测试。运行锁采用 flock，推荐在 macOS/Linux 上构建运行；不声明 Windows 构建通过。
- `Makefile`：固定 verify 串行执行代码摘要、格式、module、vet、race 单元/实际 HTTP/CLI、构建、实际 Go 服务浏览器。没有跳过失败、成功兜底或仅静态网页服务；browser 子目标仍为必需门禁。

## AC → 测试与证据

| AC | 固定入口内的用例 | 当前真实观察 |
|---|---|---|
| 01–03 | CLIInstallationCommands、InstallationRecovery、SourceAndLoginLimits、HTTPManagementAndOrdinary、InProcessContract | 初始化/重复拒绝/密码/来源/限速/基础身份单元通过；CLI 测试执行至监听占用检查被沙箱拒绝，正常服务与完整 HTTP 登录待宿主 |
| 04–07 | HTTPManagementAndOrdinary、DestinationAndProtocol、HeaderBoundaryAndFailureContract、InProcessContract；浏览器两个别名/两把密钥 | 进程内真实管理/存储/映射、SSRF 规则、保护头、状态非法写入单元通过；真实 HTTP 停用与浏览器待宿主 |
| 08–09 | TransparentParametersAndLocalReject、HTTPManagementAndOrdinary、DestinationAndProtocol | 请求类型/重复键/未知字段/部分参数契约单元通过；真实 HTTP 参数/工具/大小拒绝待宿主 |
| 10–12 | SSEBarrierAndUsage、Cancellation、UpstreamFailures、TimeoutRedirectAndNetwork、HeaderBoundaryAndFailureContract | 进程内 chunk/DONE/error/timeout 分类通过；屏障首包、实际取消计数、HTTP 超时/拒绝/断流/重定向/连接失败未执行至断言 |
| 13–15 | UsageSnapshots、InProcessContract、HeaderBoundaryAndFailureContract、StorageFailClosed、HTTPManagementAndOrdinary | 进程内 nullable 用量、固定诊断、管理查询秘密不返还、保护头、记录入站故障 fail-closed 通过；真实网络终态故障和 DB/WAL 秘密检查待宿主 |
| 16–17 | InstallationRecovery、SpecialInstallationPath、HTTPRestartAndBackup、InterruptedAndVersion、InterruptedActualRequest、ProcessCrashRecovery、CLIInstallationCommands | 停机备份/错误材料/损坏拒绝/版本拒绝/基础恢复单元通过；实际 HTTP 重启/恢复、进程强制终止在途调用和 CLI 监听待宿主 |
| 18 | tests/browser/main.go + journey.mjs，Makefile browser 必执行 | 安装锁定 Playwright 成功，JS 语法通过；Go 测试上游监听就被拒绝，未启动 Chromium，桌面/390px UI 旅程均未验收 |
| 19 | make verify；tests/evidence 的 HEAD + 源码/测试/锁/基线摘要 | 必需门禁已接入；本机退出 2，不标记整体通过，tests Connector 须真正执行同一入口 |

进程内 RoundTripper 测试仅是补充单元证据，不冒充实际 HTTP、供应商联调或浏览器验收。crash helper 是强制终止子进程的测试支撑入口，其单独返回不代表异常恢复 AC 通过。网络上游夹具先读取完整受限 body，避免未消费请求体导致测试供应商不能观察连接取消；没有改业务记录制造结果。写入失败使用隔离测试数据库触发器故障注入，正式产品无故障开关。

## 版本、依赖与限制

Go 1.25.6；mattn/go-sqlite3 1.14.52（MIT、内含 SQLite、要求 CGO/C 编译器）；x/crypto 0.55.0、x/sys 0.47.0（BSD 系许可）；Playwright/核心 1.59.1（Apache-2.0），本机 Node 26.0.0 / npm 11.12.1。版本/完整性由 go.mod/go.sum/package-lock.json 锁定；未新增供应商 SDK。依赖 README/license 直接核对已缓存发行包。缓存、二进制、node_modules、数据、秘密均被忽略。

初次 npm 离线元数据查找退出 1（ENOTCACHED）；复用用户缓存的写尝试被沙箱拒绝（EPERM）。只读取已有缓存，将必要包复制至项目忽略缓存后生成锁并安装成功，没有修改个人缓存/权限。首次 browser 安装尝试耗时约 1 分钟：可选 fsevents 包因无缓存尝试网络；随后明确 omit=optional（只是不需要的文件监听依赖）、no-audit/no-fund，锁定 Playwright 核心仍强制安装，未删除浏览器检查。

一次直接 `go run ./tests/evidence` 未设置项目 GOCACHE，默认本机 cache 路径访问被拒绝（operation not permitted）；改用 Makefile 同等项目 .cache/go-build 后退出 0。本机 race 链接有 LC_DYSYMTAB 警告，保留在日志；没有据此虚构工具链问题已修复，单元/构建的真实退出码分别记录。

所有日志的临时/本机绝对路径已规范化为占位符；保留原错误、步骤、状态及退出码，不删失败。生产默认不启用测试目的地例外。真实供应商联调未测，不读取用户供应商秘密，不使用付费额度。

## 交接门槛

本节点是研发实现，不是独立 QA。交接 `next → tests` 固定项目测试 Connector，在宿主运行 **make verify**，不得只重复已通过的单元/静态检查，也不得用静态 browser.check 替代后端。其失败应真实返回研发；通过后才由独立 QA 复核 AC/回执/相同代码。没有提交、推送或 PR 操作，后续草稿 PR 仍由 Connector 执行。

## 最终修正与回执索引

- 开发自查补齐了正式 CLI 成功启动/health/默认空态/网页管理 API、两把密钥的 SIGTERM 后重启保留及旧会话拒绝用例；DB 秘密扫描包含主库、WAL/SHM。它们全在 make verify 的 Go test 内，不是入口外的人工待办；沙箱无法监听，仍未执行至这些网络断言。
- 自查修正 CLI 收到 SIGTERM 后 Serve 先返回导致提前关库的顺序问题：等待 Shutdown 协程实际完成后才退出/关闭存储，仍保持 10 秒排空和超时取消。
- 实际复现 AC14 用量回归：合法 usage(prompt_tokens=0) 后跟 usage:null 的 chunk 原先覆盖已知快照，`TestInProcessContract` 退出 1、失败于 record contract。修正为只有合法 usage 对象更新快照；同一触发用例保留并与未知/null 单元一起退出 0。失败/修复证据分别为 `development-usage-before.log` / `development-usage-after.log`，并再次纳入完整 make verify。此前未将该正确性失败误归因于沙箱。
- 最终源代码标识：HEAD `fdcb16cb3426d403327da727ae2a8f95c001f1af`，工作树摘要 `3964b300ba3734905090d4e565fbd3fcddf31b9b3288c6d8d1d1e50faeadbf2c`，30 个源/构建/锁/验收基线文件。仓库有未提交实现，不能仅用 HEAD 代表本次代码；摘要边界由 tests/evidence 明确列出，排除交付日志避免自引用。没有 Git 提交或发布。
- `development-verify-final.log`：最终固定 make verify 的完整脱敏输出，exit 2；格式/module/vet 与非监听 race 单元通过，实际 HTTP/CLI 监听检查均因 bind operation not permitted 失败。该次固定串行门禁在 test 处停止，所以其中 build/browser 未执行，不能报整体通过。
- `development-static-final.log`：相同最终源摘要的单独 make build exit 0、两份 JS 语法检查 exit 0、证据程序 exit 0；这不代替上一条未通过的完整门禁。
- `development-browser-final.log`：相同最终实现的单独 make browser exit 2；锁定依赖安装成功（约 0.4 秒），实际 Go 测试上游启动即被沙箱拒绝监听，未运行 UI 断言。静态 browser.check 不能启动后端，按项目规则不调用它来伪造 UI 验收。
- 原尝试 retained：`development-verify-initial.log` / `development-verify-attempt2.log` / `development-verify-sandbox.log` / `development-verify-attempt4.log`、`development-browser-initial.log` / `development-browser-sandbox.log`、`development-static-initial.log` / `development-static.log`，均为较早工作树的真实回执，不与最终代码混用。早期 raw 缓存从未作为提交产物。

本节点状态：**Ready for host verification，不是 G3/QA 放行**。必需 HTTP/SSE 实时性、取消、错误、终态存储故障、正式 CLI 正常路径/重启、强制进程退出恢复、真实桌面/窄屏浏览器均需 host-only tests Connector 实际完成；源码和测试用例已就绪，不申请沙箱提权、不删测试、不扩大产品范围。真实供应商联调依旧未测。
