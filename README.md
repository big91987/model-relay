# Model Relay

面向团队自用的大模型 API 中转站。管理员接入上游、配置模型映射并为应用签发访问密钥；应用使用统一的 OpenAI 兼容入口调用模型，管理员查看请求结果与用量。

从零独立实现；不使用 New API 的源码、数据库模型或实现。它是 Agent Platform 原生编排的真实产品实践仓库，需求、设计、开发、独立 QA 和交付证据随代码维护。

## 当前目标：多租户管理控制台

- A管理工作台十页，真实平台/租户管理员、开发者、财务身份及邀请接受。
- 供应商API/已有自托管服务接入、公共模型发布/租户授权、稳定应用Key及代际轮换。
- 独立次数/RPM/TPM/并发/硬软费用预算；USD微美元价格快照、多尝试/未知费用、追加核对和账单更正。
- `/v1/models` 和 `/v1/chat/completions`，支持普通响应与 SSE 流式转发。
- 密钥保护、明确的错误边界、请求取消与数据持久化。
- 完整服务的启动、停止、重启与备份说明，真实 HTTP 和浏览器验证。

后端统一使用 **Go**，管理网页由同一 Go 服务提供。

- [产品路线图与功能拆分](docs/roadmap.md)
- [当前冻结范围及实际材料入口](docs/workflow/runs/4cd62ced6a4f98f50e3cca6f7596eb06/product-definition.md)
- [当前实现与验证状态](docs/workflow/runs/4cd62ced6a4f98f50e3cca6f7596eb06/delivery/README.md)
- [历史v0.1.0基础中转](docs/releases/v0.1.0.md)
- [产品委托](docs/product-brief.md)

## 历史演进：多上游路由与调用诊断

以下为历史发布与切片责任，不代表当前design-v0.2.0已验收。当前Issue #25从main独立实施完整冻结范围，不恢复旧工作区。源码、受控测试和公开运维说明已扩展到schema3；宿主固定门禁、独立QA、PR与正式部署仍须本Run真实回执，不能引用历史通过。

第 1 项首次配置与诊断已通过独立 QA，经 Issue #16 / PR #17 合并至 `85f6d17c07c5fddeb2d6a39a5a067769a11bb4cc` 并正式测试部署（用户提供 Deployment 6876076706 success）。这是上一切片的发布事实，不覆盖其 [历史验收报告](docs/workflow/runs/f1ff775cb6fee4d609353655cae5e180/acceptance/report-02.md)。当前 Issue #18 实现路线第 2 项：[多上游需求](docs/workflow/runs/82227c5986e5f41875895d8016c686a6/product-definition.md)、[架构](docs/workflow/runs/82227c5986e5f41875895d8016c686a6/routing-architecture.md)、[研发交付与验证状态](docs/workflow/runs/82227c5986e5f41875895d8016c686a6/delivery/README.md)。当前实现待宿主完整门禁与独立 QA，不代表本轮已合并或发布；真实供应商仍未测。

允许参考 New API 与 LiteLLM 的公开产品交互，禁止复制实现、资产及读取本机 New API 数据／凭据。[唯一公共路线](docs/roadmap.md) 第 3–4 项仍以独立 Issue 接续，本轮不提前实现应用治理或分布式设施。第 1 项通过不等于产品全量通过。

本轮研发须扩展固定 `make verify`，首次配置中间步骤全程 UI 操作，不能 API／SQL 预填；实际 Go 服务、受控上游夹具与真实供应商证据分开。需求、设计、项目测试、独立 QA、PR、合并、正式部署与部署后浏览器核验各按真实回执报告，历史发布不证明本轮交付。

首版已通过 PR #2 合并到 `main`，并在测试主机完成首次 Actions 发布。用户已授权本目标维护者核对独立 QA 后自主完成测试仓 PR 合并与测试部署，无需等待用户另行合并；阶段 Agent 仍如实交付草稿 PR，平台源仓 PR 不在此合并授权内。既有流程在合并后只自动准备，正式测试部署仍需获授权维护者通过 Actions 的部署入口显式触发，不等于自动部署或部署已成功。最新发布状态、版本和效果地址从 GitHub Deployment 或 Agent Platform 的原 Run 页面核对，操作说明见 [本机持续预览](docs/local-deployment.md)。`make verify` 的短命服务和手工预览不算正式发布。

## 重置管理员密码

本机运维者可在停止服务后，通过独立主密钥和正式命令重置密码；服务运行中会拒绝操作。先用 `backup` 保存安装，再从私有文件传入新密码：

```sh
bin/model-relay reset-password --data-dir <data-directory> --key-file <key-file> --password-stdin < <private-password-file>
```

输入为一行非空 UTF-8 文本（最多 1024 字节），命令不回显密码。重启原服务后使用新密码登录，旧密码和旧会话失效；上游、模型、应用密钥和调用记录保留。部署器生成的 `initial-admin-password.txt` 仅记录首次初始化密码，重置不会更新该历史文件。由 launchd 管理的实例需先 `bootout` 对应服务，重置后用原 plist `bootstrap`；直接杀进程会被 KeepAlive 自动拉起。

## 安装与运行

构建前提：Go 1.25.6 或兼容的 1.25+ 工具链、C 编译器及 `CGO_ENABLED=1`。SQLite driver 包含 SQLite 引擎，不需要另起数据库；本版不支持无 C 工具链的纯 Go 交叉编译。Go 依赖由 `go.mod` / `go.sum` 锁定，页面资源嵌入二进制，运行无需 Node。

```sh
make build
bin/model-relay init --data-dir ../relay-data --key-file ../relay.key
bin/model-relay serve --data-dir ../relay-data --key-file ../relay.key --listen 127.0.0.1:5544
```

`init` 创建随机管理员密码，仅向当前终端展示一次：立即安全保管，不重定向到仓库日志。独立主密钥文件也是恢复必需材料。非空安装/已存在的 key-file 拒绝覆盖；启动不自动初始化。不要让数据目录或 key-file 指向已有文件/链接，不使用共享可写目录。数据库与密钥为 0600，数据目录 0700。用户负责同机权限、磁盘容量及备份材料保管。

访问 `http://127.0.0.1:5544/admin/`（不要改为 `localhost`，Host 必须与实际监听 authority 一致），使用一次输出的密码登录。默认安装没有上游、模型或密钥，不自动发送任何生成请求：

1. 使用账号`admin`和首次密码登录。“供应商与服务”明确来源类型、基础地址和凭据；编辑同地址留空保留原凭据。显式核验不发送生成请求，保存不等于连接/生成成功。
2. “模型托管”发布公共模型及真实远端候选，选择租户授权。新模型没有隐式租户/Key授权。“定价与账单”一次发布全部候选的独立成本和统一租户输入/输出价格，单位为微美元/百万Token，保留历史版本。
3. “租户与成员”创建租户、设预算、邀请负责人。邀请链接手工受信交付，不发送邮件；新账户接受时设至少12字符密码，已有账户先登录后接受。人员与应用Key分离，移除成员收回身份/诊断权限，不擅自撤销应用Key。
4. 租户管理员在“API Keys”创建应用、模型交集、期限及诊断成员；完整秘密仅本次显示。遗失需轮换，立即或最长24小时过渡，新旧共享应用额度与历史。未知提交先安全读回，不盲目重签。
5. “配额与预算”独立配置租户/Key次数、RPM、TPM、并发及费用。空为无限、0为拒绝，改限额/暂停/轮换不清零。有限TPM或硬费用须先登记绑定当前执行配置的可信收费Token上界；不能证明时拒绝，而不是假称硬保证。上界依据由运营者承担供应商契约责任，测试上游不证明真实供应商。
6. “调用测试”显式使用已持有Key，普通/SSE输出、取消与同Request ID诊断；取消只停止本地接收，不能宣称上游免收费。页面离开/退出清敏感，不自动重发；“调用日志”支持模型/Key/状态/时间/ID筛选、分页、精确详情与导出。
7. 平台管理员按可追溯依据核对未知用量/历史价格；所有dispatch的已知用量均计成本与租户费。仅已结束且无未决的周期可结账；已关闭金额不可覆写，更正追加原因、操作者与开放调整周期。外部支付仍未接通。

正式schema3是十页A工作台，真实身份范围由服务端决定。概览趋势/租户费用/风险来自持久事实；缺数据为未知/空，错误不冒充0。运行版本独立读取`/healthz.version`及`storage_schema`。schema2六页仅留作兼容回归，不是本期交付UI。

应用端调用示例（环境变量仅供调用者自行填入，测试不读取用户密钥）：

```sh
curl -H "Authorization: Bearer $APPLICATION_KEY" http://127.0.0.1:5544/v1/models
curl --no-buffer http://127.0.0.1:5544/v1/chat/completions \
  -H "Authorization: Bearer $APPLICATION_KEY" -H 'Content-Type: application/json' \
  -d '{"model":"your-alias","messages":[{"role":"user","content":"Hello"}],"stream":true}'
```

## 停止、备份与恢复

使用 Ctrl-C 或向该进程发送 SIGTERM。最多排空 10 秒，超时取消剩余请求；不静默更换占用端口。服务排他持有安装锁；同目录多实例、运行中备份均明确失败。

```sh
# 先停止服务，再执行；输出文件必须不存在。
bin/model-relay backup --data-dir ../relay-data --output ../relay.backup
# 使用原 key-file，恢复到不存在或空的独立目录。
bin/model-relay restore --input ../relay.backup --data-dir ../relay-restored --key-file ../relay.key
bin/model-relay serve --data-dir ../relay-restored --key-file ../relay.key --listen 127.0.0.1:5544
```

备份含数据库、版本与摘要，不含独立主密钥或登录会话。数据库仍含认证摘要与加密上游凭据：备份同样属于敏感材料。请分别安全保存备份与原 key-file，丢失材料不能解密恢复。恢复先校验归档路径/类型/大小/摘要/版本与加密哨兵，失败不覆盖原安装；本版恢复的数据库上限 1 GiB。损坏、未知 schema 版本、错误/缺失主密钥必须人工处理，不删库重建。恢复及重启保留配置、应用密钥状态和历史；浏览器旧会话失效。

非正常退出的 `started` 记录在下次启动变为 `process_interrupted`，不会重放。记录写入失败使服务 degraded：健康接口 503、新生成 fail-closed。停止后排查磁盘与权限，保留原安装，再重启；不能假设磁盘故障时最后一条记录已保存。

## 协议与安全边界

仅支持 `GET /v1/models` 和文本/函数工具子集的 `POST /v1/chat/completions`。协议清单、来源规则与错误契约以 `docs/workflow/design.md` 第 4–6 节为准：软/无限模式支持既有常见生成参数、function tools/choice、response_format、stream_options.include_usage，未知字段和多模态转发前拒绝。有限TPM/硬费用模式仅支持可证明有界的普通文本子集，拒绝tools、多模态、n>1及超过已声明输出上限的请求，默认发送显式安全输出上限。只替换请求 model；普通结果及合法 SSE chunk 原样返回，响应中的真实 model 不改写。不转发客户端 Cookie/自定义头或供应商 Set-Cookie；不会访问客户端指定目的地址。

默认拒绝 HTTP、私网/loopback/特殊地址、环境代理、所有重定向；配置和实际 Dial 均复验 DNS。管理写操作须精确 Origin，登录后的写操作还需会话 CSRF；管理会话与应用 Bearer 隔离。上游 token AES-GCM 加密，管理员密码 Argon2id，随机应用密钥仅存 SHA-256 摘要。本版只支持 loopback，无公网 TLS/生产部署承诺，同机特权进程不在秘密保护边界内。

接入已有私网推理端点时，仅运营者在明确授权后为`serve`配置`--private-upstream-authorities host:port,host:port`精确白名单；未列明仍拒绝。该列表可允许私网HTTP/HTTPS，但不允许loopback、metadata/link-local及混合DNS越界，不自动扩大网络权限。产品默认不设置白名单，受控loopback只存在于测试Policy。

生成请求最大 1 MiB、管理 JSON 64 KiB、普通响应 8 MiB、SSE 事件 256 KiB/全流 32 MiB；响应头 30 秒、整个请求 300 秒、提交前每次尝试 60 秒、有效 SSE 事件空闲 60 秒；逐次下游写 10 秒。全局登录 10 次/分钟，最多两个并行口令验证。同请求最多尝试 3 个不同上游，不重复同一上游；优先级大数先用，同级平滑加权轮转。只有 HTTP 408/429/500/502/503/504、暂时网络或 TLS/传输失败可在响应提交前切换；暂时失败冷却 30 秒后被动恢复。协议/授权/存储失败、整体超时、取消或已向客户端提交响应均停止切换。SSE 只有记录落盘后才发送 `[DONE]`；`event: error` 或缺少 `[DONE]` 均为不完整，不以 finish_reason 代替结束。取消不能保证供应商撤销费用，多次尝试可能产生重复消耗；未知用量保留未知并展示已知部分。

上游与候选由控制台或同一管理 API 持久维护：`/admin/api/upstreams` 支持分页、创建、按 ID 编辑；编辑需 `expected_revision`。模型写入整个 `candidates` 集合（上游 ID、远端名、启用、0–100 优先级、1–100 权重），更新也需 `expected_revision`。旧单上游管理入口返回 410，无配置双写。连接核验显式调用 `POST /admin/api/upstreams/{id}/check`，携带 `expected_execution_revision`；仅查询上游模型列表，不生成，不保证供应商不收取查询费用，也不证明生成健康。

## 从旧安装升级

新安装使用 schema 3；`serve` 不会隐式升级 schema 1/2。停止旧服务、保留独立主密钥并先由旧 binary 备份，使用候选 binary 执行：

```sh
bin/model-relay deployment-contract
bin/model-relay upgrade --data-dir <data-directory> --key-file <key-file> --output <new-migration-backup.tar>
```

`upgrade` 在独占安装锁内校验凭据、创建新的升级前备份，再原子迁移；返回 JSON `migrated` 或 `already_current`。已是 schema 3 时不重复备份/迁移，仍必须提供 `--output`。保留旧管理员、应用密钥、模型及历史，旧Key归入“原有应用”租户；未知历史用量/价格不补0，平台必须通过带证据的正式历史核对入口处理。无上游的遗留映射保留远端名待关联。健康检查报告实际`storage_schema`，部署只接受CLI声明的schema3。

回退只承诺升级前完整快照：停止新服务，用旧 binary 的 `restore` 将旧归档恢复到空目标并配原主密钥，不合并升级后新增数据。不要把本机手动升级当正式发布；平台正式 activate 需完成旧 binary 备份→候选升级→切换启动/实际 schema 核对及失败恢复，用户已报告平台固定源d0564d05的接缝实现、隔离回归和标准安装升级；这不代替真实新旧Go联调及正式 Actions 非空升级/失败恢复复验，后两层仍是发布依赖。详见 [升级与发布契约](docs/workflow/runs/82227c5986e5f41875895d8016c686a6/deployment-upgrade-contract.md)。

## 验证

产品范围、AC映射与历史执行结果见[历史验证记录](docs/workflow/runs/4cd62ced6a4f98f50e3cca6f7596eb06/validation/report.md)，不代替新候选验收。本次干净安装修复与宿主待验状态见[Issue28研发检查](docs/workflow/runs/3e7496815b9922704072482cf44e73e2/verification/development-check.md)。原生沙箱已实测不能监听；完整`make verify`仍交宿主固定项目Connector，不将定向单测/构建当全量通过。门禁包含真实HTTP/SSE、人员范围、独立资源/费用、多尝试核对、schema1/2→3与崩溃重启、1280/1440/390实际Go UI；原可信历史/浏览器门禁保留。

前提：上述 Go/C 工具链，以及 Node 18+、npm、Playwright Chromium（仅测试依赖）。首次准备浏览器可使用以下命令；已提供匹配 Chromium 的干净 clone/worktree 直接执行 `make verify` 即可，无需预装 `node_modules`。`make frontend-test` 和 `make browser` 都先通过共享 `verification-deps` 执行锁定的 `npm ci --ignore-scripts --omit=optional --no-audit --no-fund --cache .cache/npm`，安装失败即停止，不使用全局模块或 `NODE_PATH`。完整入口分别进入这两个独立目标时会重新校验安装；npm 缓存只加速下载，不替代 lockfile。Chromium 二进制和其系统库仍是宿主前置条件，不能将缺失浏览器算作通过。

```sh
make verification-deps
./node_modules/.bin/playwright install chromium
make verify
```

固定 `make verify` 顺序执行工作树代码摘要/HEAD、旧版健康方法真实 HTTP 红灯重建与精确分类、格式、module 校验、vet、前端异步状态回归、race 单元与真实 HTTP/CLI 测试、构建、实际 Go 服务浏览器旅程（1280/390px，包含延迟配置加载屏障）。`make frontend-test` 可单独运行实际 app.js 的隔离状态测试，不代替真实浏览器验收。任何失败都会非零退出，HTTP 或浏览器不可用不算通过。浏览器依赖锁在 package-lock 中，不用于产品后端；受控上游只在测试程序内以指定 authority 放行，正式 CLI 没有测试开关、没有默认假供应商。测试拥有临时数据及随机端口，不使用用户安装或付费凭据。

固定 `make verify` 调用项目 Go 执行器 `tests/verify`，按 Git tracked/unignored inventory 复制当前候选到新临时目录（不复制 ignored 的依赖、cache或运行数据），复用原仓库只读历史并核对源码摘要，再执行保留全部原顺序的内部 `make verify-gates`。清除 NODE_PATH/NODE_OPTIONS、继承的 npm 配置变量和 Make 控制参数；npm 用户/全局配置分别指向副本内两个空文件，不共用 `/dev/null`（npm 会拒绝同文件重复加载）。副本内 npm cache与Go build cache从空开始。原工作区已有node_modules保持不动，不作为冷安装证明；`verify-gates` 不是独立验收替代入口。工具版本前置失败时，原始 stdout/stderr 也进入脱敏日志供归因。

2026-10-25 用户授权的执行条件：内部完整 `verify-gates` 预算为25分钟（`gate_timeout_ms=1500000`），取代原写死12分钟；冻结执行器外层30分钟不变，为候选准备和清理留余量。`GATE_START` / `GATE_END` 与 `state.json` 记录配置预算、开始/结束UTC和取消状态，父取消仍优先生效。超时继续非零退出并清理所属进程组和候选，不重试、不再自动提高时限。此调整不代表性能改善，不更改历史失败结论，也不省略历史/迁移、race、真实HTTP/浏览器或严格失败恢复断言。

入口自动打印 `test-results/verification/<UTC>-<unique>/`，流式输出并保存完整脱敏 `verify.log` 和原子更新的 `state.json`：实际干净起点、工具版本、原树/副本前后摘要、结束依赖状态、锁定Playwright/core本地解析及Chromium版本、退出码、日志字节数/SHA256、清理范围。每次新建目录，不覆盖历史，不硬编码Run、候选摘要或机器路径。SIGINT/SIGTERM只终止本次创建的进程组并清理自己的临时副本；强制杀死执行器可能留下未完成state，缺少最终finished_utc/exit0不算通过。原始浏览器产物保存在本次 `browser/`，仍为ignored本地产物，QA须审查隐私与同版后选取实际文件归档当前Run，不能直接提交整目录。

`make health-history` 仍只输出当前重建回执；其成功表示旧版21个方法断言失败与12个查询/恢复通过被精确识别，不表示旧版全绿。原门禁、摘要不一致、日志/回执写入失败均非零退出。新增执行器回归进入原race测试；夹具绿灯只证明隔离/采集/失败传播机制，不代替正式宿主完整冷 npm、HTTP与浏览器验证。

Go 门禁保留 `-race -count=1 -v ./...`，以 `-parallel=4` 有界执行显式标记 `testing.T.Parallel` 的独立场景。路由状态码、上游失败、连接阶段、资源维度、终端归属及人员/汇总测试各自创建临时安装和受控上游；共享同一安装的步骤仍顺序执行。未降低口令成本、缩短真实等待、调整时钟或删改断言。并行父测试显示的 `0.00s` 不包含子测试墙钟时间，分析总耗时应使用包级结果；完整冷入口仍须实际通过全部 HTTP/浏览器门禁。

登录后概览的「运行版本」读取当前服务 `/healthz.version` 原始值（包括 `dev`），与静态品牌标签无关。健康请求失败、版本缺失/为空或等待超过 5 秒时显示「版本未知」，不阻塞管理操作；离开再进入概览或刷新页面会重新获取。该显示不证明供应商联调或多实例版本一致。版本回归已纳入同一 `make verify`，无需新运行命令。

配置生命周期回归包括不同模型编辑、新地址/新凭据及旧目标计数、上游禁用无消耗后恢复。两屏网页旅程创建两个不同映射、编辑后正常停止 HTTP 服务并关闭存储、重新打开同一临时安装和启动 Go 服务，再登录核对管理响应与恢复网页；同时检查调用记录及捕获的服务日志不含完整凭据、哨兵对话或敏感上游错误。浏览器 runner 的独立 loopback 生命周期/日志检查服务仅在测试内存在，不是产品 API，也不修改数据库记录。实际进程 SIGTERM/崩溃恢复另由原 Go/CLI 测试覆盖。

研发环境监听受沙箱限制，完整通过回执需已配置 host-only 项目测试节点实际执行；参见本轮 [验证记录](docs/workflow/runs/82227c5986e5f41875895d8016c686a6/delivery/verification/report.md) 的失败与待核验清单。固定入口新增双上游真实路由旅程与精确旧 main 的非空用户页面→backup→upgrade→旧 binary restore 旅程；这些实现不等于已运行通过。真实供应商联调未测，受控测试、构建成功及草稿 PR 都不代替独立 QA 或获授权目标维护者的最终验收。

## 同模型低延迟优先（Issue 63）

管理员在“模型托管”的现有模型编辑器选择“权重（默认）”或“低延迟优先”，保存后重新进入仍保留；重启保留策略，但近期评分重新学习。旧数据及未指定策略的新模型默认权重，旧客户端 PATCH 不带策略不会重置已有策略。两种策略都只使用已授权、启用、健康且符合最高可用优先级的候选；原租户/Key、额度、非零 RPM/TPM、并发和金额预算规则不变。

低延迟优先只从正常付费请求学习，不发隐藏探测。普通响应测上游响应体首字节（不等于用户看到正文），SSE 测完整事件内首个非空内容/拒绝/工具函数输出增量；忽略心跳、注释、role-only 和空增量。两类评分分池，均不是首 token。成功完成、终端写出和存储提交后才采纳；失败、取消、配置变化或旧序列的迟到响应不进入评分。

每候选最多32个样本，按实际测量时刻逐个60秒过期，至少3个才用 EWMA（alpha=0.25）排序。每5次同组选择一次正常流量探索，稳定候选集合每候选至多5n次选择获探索；其余选近期最低评分，无评分回退权重。长流完成不会把旧测量续期；探索意味着慢尾仍存在，不保证 p95 或生产收益。

调用测试的“查看同Request ID日志”显示所选上游、策略、指标、当时样本数/年龄/EWMA、探索原因和本次采纳/拒绝依据，可关联同请求费用；未知保持 null，不记录提示词、回复正文、工具参数或凭据。详细冻结验收与受控实验方法见 [本轮设计](docs/workflow/runs/ba434da8039cca16f6829481112e3d68/plan-v2.md) 和 [预注册实验](docs/workflow/runs/ba434da8039cca16f6829481112e3d68/experiment-v2.md)。受控 HTTP/SSE 不是供应商验证；真实供应商为 Not Run，固定门禁与独立 QA 结果以各节点回执为准。
