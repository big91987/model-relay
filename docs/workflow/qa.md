# 管理台运行服务版本：独立 QA

> 日期：2026-10-06（Asia/Shanghai）
> 模式：本次局部功能首次全量验收；固定入口执行全项目回归。执行者为宿主项目测试 Connector；本节点独立复核规格、版本、覆盖与回执，并在 Agent 沙箱补跑非监听测试。
> 结论：**Go，允许进入报告与草稿 PR 交付；不部署。** 未发现本次范围内未关闭 P0/P1。原 Issue 回写与草稿 PR 仍由平台后续交付，不在 QA 冒充已完成。

## 1. 本次基线与版本冻结

- 基线：task.md 已确认澄清、本次 requirements.md AC-01～08、design.md。用户确认仅登录后概览显示 `/healthz.version` 原值（含 dev），异常为「版本未知」，非阻塞，重入/刷新恢复，无供应商调用。澄清已完成，无需重复提问。
- 仓库：big91987/model-relay；分支 `workflow/509e5c927cb88a6bbcefa274656ac6d2`；HEAD `9b6994ed80de2bdba79ff267419150dccaac1360`。实现尚未提交，HEAD 本身不能证明被测改动。
- QA 独立运行 `go run ./tests/evidence` 得到 33 文件摘要 `f96017d3daf52e8372e1ef106397bfb31224e049010bb53a6cd033d69f27d414`，与本轮 tests seq 7 宿主回执一致。核对摘要程序覆盖产品、全部测试、Makefile、锁文件、README 和本次基线，非仅相信研发摘要声明。QA 文档及日志不参与自引用摘要。
- 起始未提交修改：README.md；docs/workflow/{task,requirements,design,delivery}.md；internal/relay/web/{app.js,index.html,style.css}；tests/browser/{journey.mjs,main.go,ui.test.mjs}。未跟踪：internal/relay/version_test.go 和五份 development-version 日志。QA 仅新增本轮证据、更新 qa.md，没有修改产品或测试实现。
- 测试制品：当前工作树 Go 服务、其内嵌 UI；宿主 Go go1.25.6，Playwright 锁定依赖与 Chromium。QA 本机 Node v26.0.0。服务测试版本明确设为 dev；真实 HTTP 另覆盖发布样式版本。没有生产部署制品。

## 2. 连续旅程与验收矩阵

Actor 为隔离测试实例的管理员。Go runner 使用临时目录、SQLite 和临时端口准备安装及密码，浏览器通过页面登录；夹具准备不是用户中间步骤。每个视口使用同一 context/page 连续完成入口、结果与退出；1280px 首次无上游配置，390px 为已配置状态。本次只读版本无独立创建/授权流程，不额外要求运营或供应商验收。

UI-E2E 连续步骤：打开登录页（无版本、无健康请求）→错误密码→正确登录概览→未拦截实际健康响应与 UI dev 精确一致→刷新一次请求→切模型页隐藏版本→拦截健康 401 后进入概览未知、概览仍可用→刷新仍未知→撤销拦截并重入/刷新恢复→挂起健康时可读模型页→返回概览验证标记字面值、长文本不溢出→恢复真实健康→继续既有上游、模型、密钥、调用与记录回归→退出隐藏版本及请求计数边界。两视口宿主 UI-E2E 均 Pass；没有用静态原型替代后端。

| 基线 | 分层证据与独立核对 | 结果 |
|---|---|---|
| AC-01 登录/概览边界 | 宿主真实 UI 入口无健康请求，登录显示，模型页及退出隐藏；审阅 journey 对应断言 | Pass |
| AC-02 原值与 dev | UI 与未拦截 `/healthz` dev 精确对账；HTTP dev/发布字符串；VM dev/发布/空白原值精确对账 | Pass |
| AC-03 异常未知 | VM 分别覆盖网络、401/503、缺失、空白/空字符串、非字符串、null/array/string body、无效 JSON；真实 UI 401 显示未知 | Pass |
| AC-04 独立非阻塞 | 真实 UI 健康挂起时概览/模型读取正常；401 不退出且 error 为空；VM 5 秒涵盖 headers/body；审阅独立 fetch 未使用管理错误包装 | Pass |
| AC-05 重入/刷新恢复 | 同一真实会话成功→失败→刷新失败→撤销拦截→重入恢复→刷新成功；健康请求计数，VM 每种异常后重入恢复 | Pass |
| AC-06 迟到隔离 | 实际 app.js 的 VM 确定性测试：旧成功/失败/finally 不清新计时器、旧响应不覆盖新结果、超时后迟到、退出立即失效/重登、真实管理401失效；真实 UI 挂起切页 | Pass |
| AC-07 文本安全/只读 | 真实 UI 含 img/onerror 原值仅文本、无子节点/注入、双视口无横向溢出；独立 diff 审查版本链路仅 GET /healthz、textContent 和局部状态，无配置写入/供应商调用 | Pass |
| AC-08 QA 与固定门禁 | 当前摘要宿主完整 make verify exit 0，新增回归确实在原入口执行；本报告保留失败和未测边界 | Pass（本节点责任） |
| AC-08 后续 Issue/草稿 PR | Hook/Connector 后续交付，QA 没有自行 GitHub 操作，也未部署 | Not Run（后续职责） |

VM 是异步状态测试，不等同真实浏览器。HTTP 是接口证据，不冒充 UI。表中真实 UI Pass 仅来源于宿主真实连续旅程的回执，不来源于代码阅读。

## 3. 实际执行与证据

- 宿主：本轮 tests seq 7 固定 `make verify` exit 0，format/modules/vet、前端 24/24、全 Go race/HTTP/CLI、build、真实 Go 服务 Playwright 双视口全部执行；详见 qa-version-host-receipt.md 摘录与平台完整回执。
- QA 沙箱：独立摘要核对；`make frontend-test` exit 0（24/24、skip=0）；`go test -race -count=1 -v ./internal/relay -run '^TestHealthVersionContract$'` exit 0（3 个版本子项）；`git diff --check` exit 0。实测输出在 qa-version-local.log，临时链接对象路径已脱敏。不是完整门禁或额外 UI 执行。
- 独立静态复核：Makefile 未削弱门禁；runner 启动真实 relay/Store 与测试上游，非 Node 产品后端；journey 真实浏览器断言失败会非零退出，不吞断言；健康正常路径没有拦截，异常和文本夹具明确用 page.route 注入。测试上游仅既有确定性回归，不是本次新增供应商访问。
- 截图/视频：本轮未生成成功截图，UI 证据是真实浏览器断言及 Connector 回执；不声称人工视觉、读屏音频或无障碍全量验收。role/status 与 aria-live 属性已在真实浏览器断言，但不等于读屏验收。

## 4. 失败、缺陷、限制与恢复

研发首轮健康契约 403 为测试 Host/Authority 夹具错误，不是沙箱问题；development-version-verify-initial.log 保留原失败，当前匹配版本的独立契约复跑与宿主 HTTP 全绿，确认已关闭。其余研发日志记录监听被 Agent 沙箱拒绝、依赖安装 EPERM；这些不是产品断言失败，不能算原轮通过。本节点不要求扩大权限重跑监听；当前同摘要宿主完整回执覆盖原受限项，不把旧失败日志改绿。

本次没有新增、复发或未关闭产品缺陷；macOS linker LC_DYSYMTAB 警告在宿主及 QA 存在，测试正常退出。相对此次研发交接，关键变化是完整宿主门禁与真实 UI 证据已补齐；此前 v0.1 QA 缺陷与旧摘要不作为本次通过依据。

未测项：真实供应商、生产环境升级、集群版本一致性、实际构建标记是否准确、人工读屏及未来里程碑；均非本次承诺。服务自报 dev 不等于已证明部署到指定发行版。所有异常种类和超时/竞态在 VM，未逐项另跑真实浏览器；符合本次 AC 的确定性回归方式，真实 UI 的正常、失败恢复及非阻塞路径已有独立执行证据。

环境：runner 使用独立临时数据，defer 关闭服务/监听并 RemoveAll，浏览器关闭 context/browser；回执没有清理错误，未接触用户上游/业务数据。QA 本地契约使用 testing 临时数据；仅保留忽略的构建缓存和本次报告，无凭据或本机绝对路径纳入证据。

## 5. 放行与交接

Go → report：本次 AC-01～07 与 AC-08 当前阶段要求已有匹配版本真实证据，无需求矛盾、设计错误或实现阻断。后续平台须按已授权流程回写原 Issue、Connector 提交/推送并创建草稿 PR，不合并、不部署。若交付前产品/测试/基线摘要发生变化，不能复用此绿色结论，需重新执行受影响验证并匹配回执。

---

# 历史记录：Model Relay v0.1 独立 QA（不作为本次版本展示验收依据）

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
