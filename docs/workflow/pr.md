# Model Relay v0.1：新增 Go 管理服务、应用密钥与普通/SSE 中转闭环

建议草稿 PR 标题：**feat: 实现 Model Relay v0.1 Go 管理与普通/SSE 中转及持久化运维**

## 背景、原问题与预期

原仓库只有产品范围文档，没有可运行服务、管理网页、Go 模块或验证入口。团队无法通过统一入口配置供应商、为应用签发独立权限、调用模型并追踪结果。本次是从零首版交付，不是对已有中转产品的局部修补。

依据 `docs/workflow/task.md` 的验收目标，仅实现 `docs/roadmap.md` M1 和 `docs/releases/v0.1.0.md`：从空目录初始化，经管理员配置上游与别名、签发应用密钥、普通/实时 SSE 调用、查看记录、停用密钥，到停止、重启保留数据和备份恢复。需求基线是 `docs/workflow/requirements.md` 的 AC-01～19；接口、安全、存储和失败契约见 `docs/workflow/design.md`。

当前结论：**M1/v0.1 受控验收通过，可交授权 Connector 创建草稿 PR，仍需人审阅。** 不是生产发布或真实供应商联调通过声明，本文也不表示草稿 PR 已创建。

## 具体功能与行为

- Go CLI 提供 `init`、`serve`、`backup`、`restore`；初始化生成管理员密码并仅向终端显示一次，拒绝覆盖现有安装或主密钥。数据存于独立 SQLite 安装目录，不写入源码。
- 同一 Go 服务嵌入管理网页：管理员登录/退出、单上游配置与启停、多个模型别名创建/编辑/启停、命名应用密钥签发与停用、手动普通/SSE 调用测试、调用记录分页。加载失败可恢复，桌面及 390px 窄屏操作真实后端。
- 应用使用 Bearer 密钥访问 `GET /v1/models`、`POST /v1/chat/completions`。别名映射到真实模型；生成参数及工具字段遵循明确兼容子集，不能支持的输入明确拒绝，不静默丢弃。
- 每次生成最多一次上游 POST，不自动重试或故障转移。SSE 增量及时转发，客户端取消中止上游；覆盖流前/流中错误、断流与有限超时。成功终态记录落盘后才发 `[DONE]`，失败不伪装正常结束。
- 记录请求标识、模型、状态、耗时、上游状态及可空用量；缺失用量不是零，默认不记录提示词、回答、凭据或上游诊断正文。写记录失败进入 degraded，健康检查返回 503，新生成 fail-closed。
- 正常停止和重启保留配置、应用密钥状态与历史；旧管理会话失效。异常终止的未完成记录在下次启动标记 `process_interrupted`，不重放生成。停止服务后可备份并用原独立主密钥恢复到空目录。

## 关键实现取舍

- 后端、管理 API、代理、持久化和运维全部 Go；HTML/CSS/JavaScript 仅为 Go 提供的前端。没有默认假供应商，没有供应商 SDK；原创实现，不读取或改编 New API 实现。
- 单进程、嵌入页面、SQLite 单事实源适合首版团队自用，不引入多服务或组织级运营平台。设计推荐的 pure-Go SQLite driver/Go 1.27.1 未作为最终构建基线：研发记录了获取依赖的失败证据，实际采用 Go 1.25.6 与 CGO SQLite driver，代价是需要 C 工具链、不支持无 C 工具链的纯 Go 交叉编译。调整依据和历史尝试见 `docs/workflow/delivery.md`。
- 锁定 `go-sqlite3 v1.14.52`、`x/crypto v0.55.0`、`x/sys v0.47.0`，Go 模块由 `go.mod`/`go.sum` 管理；Playwright 1.59.1 由 `package-lock.json` 锁定，仅用于测试，产品运行无需 Node。
- 管理会话与应用鉴权隔离；管理写操作校验来源，登录有限速。应用密钥摘要存储，上游凭据 AES-GCM 加密且不回显；改上游地址必须提交新凭据。地址校验包含实际 Dial 重新解析，禁止代理及重定向，默认拒绝私网目的地。
- 有界请求/响应与超时：生成输入 1 MiB、管理 JSON 64 KiB、普通响应 8 MiB、SSE 单事件 256 KiB/全流 32 MiB；上游响应头 30 秒、总请求 300 秒、有效 SSE 事件空闲 60 秒。默认 loopback，公网发布不在本次范围。
- 主要实现位置：`cmd/model-relay/`、`internal/relay/`、`internal/relay/web/`；固定验证入口 `Makefile`，真实浏览器旅程与摘要入口分别在 `tests/browser/`、`tests/evidence/`。

## 实际经过的阶段与修复

实际顺序为入口判断 → 需求 → 设计 → 开发 → 宿主固定验证及研发续修 → 独立 QA → 补验收资产 → 宿主全量验证 → 独立 QA 复验 → 本交付说明。需求、设计、开发和独立 QA 均未跳过；不存在短路径省略 PRD/设计的情况。独立原型、生产部署、压测和 M2～M4 不属于本次委托，不补造其产物。

研发与 QA 的失败并未删除：

| 原故障或缺口 | 预期、修正与验证 |
|---|---|
| 空 usage 覆盖已知用量快照 | 缺失字段保留已有快照、不伪造零；研发保存 `development-usage-before.log`/`development-usage-after.log`，最终全量回归继续覆盖 |
| 上游 GET 回填前展示表单，覆盖新输入，required 阻止保存；宿主 seq 9 无 PUT、提示超时 | 加载完成再展示表单，并保护过期导航；实际 app.js 红转绿回归及两屏延迟 GET 屏障证明加载期间不可编辑、加载后可保存 |
| 拒绝保存后重新导航，浏览器测试即时读取隐藏字段旧值；seq 11 断言失败 | 修正测试同步：恢复导航同样等待真实 GET 屏障；保留原持久值预期及 10 秒超时，未改产品或放宽断言 |
| seq 14 独立 QA No-Go：AC-04 动态配置生命周期、AC-15 网页记录秘密排除覆盖不足 | 新增真实 HTTP 配置/凭据更新、旧目标计数、禁用无消耗/重新启用，以及两屏编辑/多映射正常重启和敏感失败记录 DOM/服务日志排除；seq 16 全量通过后独立 QA 关闭 QA-GAP-01/02 |

`delivery.md` 保留各轮研发当时的“待宿主验证”状态；`qa.md` 第 1～8 节保留历史 No-Go。当前版本的最终状态以 **`qa.md` 第 9 节**及 seq 16 匹配回执为准，不把旧结论改写成成功。

## 实际验证与版本证据

2026-10-05，平台项目测试 Connector 在宿主实际运行 **`make verify`，seq 16，exit 0**。完整原回执保存在平台；仓库审核索引为 `docs/workflow/qa-host-seq16-review.md`，不是另一份测试执行，也不是完整原始回执附件。

验证时 HEAD：`fdcb16cb3426d403327da727ae2a8f95c001f1af`；实现仍未提交，不能只用 HEAD 证明版本。32 文件工作树摘要：

```text
fbc07625a4d9819bba5aa809de5dfea636450cf42489d7d7c72ab66ccba06efe
```

摘要边界由 `tests/evidence/main.go` 定义：产品、测试、构建锁、README 和 task/requirements/design；不包含 delivery、QA、本文及回执，避免自引用。本报告节点重新计算摘要，与 seq 16 和独立 QA 一致；未修改产品、测试或验收基线。

| 实际门禁 | 结果及范围 |
|---|---|
| 格式、Go module 校验、vet | 全部通过 |
| 前端异步状态回归 | 4/4 通过：加载后保存、加载失败恢复、旧导航保护、拒绝保存后的配置恢复 |
| `go test -race -count=1 -v ./...` | 通过，包括实际 HTTP、CLI、进程正常/异常退出、备份恢复、鉴权与来源、配置更新、凭据隔离、SSE 首包屏障、取消、超时/重定向/失败、记录存储 fail-closed 和用量快照 |
| `go build -trimpath` | 通过 |
| 实际 Go 服务浏览器旅程 | 1280px、390px 均通过；包含加载/恢复屏障、多映射重启、普通/SSE/取消、签发/停用/退出、真实敏感失败及记录区域/捕获服务日志秘密排除 |

独立 QA 在 `docs/workflow/qa.md` 第 9 节确认 AC-01～19 Pass、QA-GAP-01/02 Closed，结论 **Go（仅受控验收）**。独立 Agent 非监听自查另有 `docs/workflow/qa-sandbox-recheck.log`，exit 0；其格式/module/vet、4 前端回归、7 非监听 race 用例及 build 不混称宿主 HTTP/UI 执行。

沙箱历史 `make verify`/browser 因监听或 Chromium 权限限制失败，证据留在 `docs/workflow/development-*.log`；未申请提权、未删除门禁。宿主链接器 LC_DYSYMTAB warning 未导致失败；用例注入的 `storage_write_failed` 是预期故障覆盖，不是被忽略的产品错误。没有生成截图或 trace，不编造附件。

## 已知限制与未测边界

- **真实供应商联调 Not Run**：没有真实凭据、响应或付费调用授权。测试使用隔离目录/随机端口及显式允许的受控上游，不触达用户安装；受控成功不证明供应商实际可用或全协议兼容。
- 仅单管理员、单上游及明确 OpenAI 兼容子集；不含多渠道调度、自动故障转移、计费、角色、多租户、其他协议、HA 或后续 M2～M4。没有吞吐/SLA 承诺或生产部署验证。
- 调用测试必须由用户明确发送；配置保存不代表联调成功，取消也不能保证供应商撤销费用。流中失败可能已交付部分输出，不能撤回。
- 构建需 CGO/C 编译器；备份须先停服，并分别保管备份与原主密钥。备份不含主密钥或登录会话，丢失密钥不能解密恢复；恢复数据库上限 1 GiB。
- 宿主 Node/Chromium 精确版本未在验证回执中报告，不臆造。窄屏使用同一隔离安装的新浏览器 context，不宣称额外一次空安装；正常服务/存储重开由网页 runner 控制，真实 CLI/进程退出另有 Go 测试。

## 试用与审阅方式

前提：Go 1.25.6 或兼容 1.25+ 工具链、C 编译器及 `CGO_ENABLED=1`。使用新的独立目录和 key-file，不覆盖已有数据，不把密码输出重定向到仓库证据。

```sh
make build
bin/model-relay init --data-dir ../relay-data --key-file ../relay.key
bin/model-relay serve --data-dir ../relay-data --key-file ../relay.key --listen 127.0.0.1:5544
```

浏览器访问 `http://127.0.0.1:5544/admin/`，使用初始化一次展示的密码；不要替换为 localhost，Host 必须与监听 authority 一致。配置获授权的 HTTPS 上游和凭据 → 建立别名 → 签发应用密钥并安全保存 → 在调用测试页明确发送普通/SSE → 查看记录 → 停用密钥核对拒绝 → Ctrl-C 停止 → 用相同参数重启并重新登录，核对保留状态。真实调用可能产生费用，没有凭据时不要把未联调当失败或通过。

完整应用 curl 示例、`backup`/`restore` 命令和安全注意事项见 `README.md`。复跑自动验收需 Node 18+、npm 和 Playwright Chromium，仅用于测试：

```sh
npm ci --ignore-scripts --omit=optional --no-audit --no-fund --cache .cache/npm
npx playwright install chromium
make verify
```

受限 Agent 沙箱不能监听时沿用已配置宿主验证 Connector，不用静态页面检查代替。审阅优先关注鉴权/来源/地址和加密边界、SSE 完成与存储失败、运维恢复以及真实测试断言。

## 后续 Connector 交接

走 `next → publish`。仅由授权 Connector 对当前工作树提交、推送并创建草稿 PR，使用本文建议标题和正文；最终合并由人决定。本节点未创建 Issue、分支、提交、推送、PR、部署或 GitHub Workflow，也未扩大访问权限。

提交后 HEAD 自然变化，不改变上述提交前验证事实；保留原 HEAD 和内容摘要关联。若发布前修改产品、测试、构建锁或验收基线，必须重新计算匹配摘要、执行固定 `make verify` 并返回独立 QA，不沿用 seq 16 放行。
