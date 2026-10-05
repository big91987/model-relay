# Model Relay

面向团队自用的大模型 API 中转站。管理员接入上游、配置模型映射并为应用签发访问密钥；应用使用统一的 OpenAI 兼容入口调用模型，管理员查看请求结果与用量。

从零独立实现；不使用 New API 的源码、数据库模型或实现。它是 Agent Platform 原生编排的真实产品实践仓库，需求、设计、开发、独立 QA 和交付证据随代码维护。

## 首版目标

- 单管理员管理网页，上游配置、模型别名、访问密钥和调用记录。
- `/v1/models` 和 `/v1/chat/completions`，支持普通响应与 SSE 流式转发。
- 密钥保护、明确的错误边界、请求取消与数据持久化。
- 完整服务的启动、停止、重启与备份说明，真实 HTTP 和浏览器验证。

后端统一使用 **Go**，管理网页由同一 Go 服务提供。

- [产品路线图与功能拆分](docs/roadmap.md)
- [当前阶段：v0.1.0 基础中转](docs/releases/v0.1.0.md)
- [产品委托](docs/product-brief.md)

首版已通过 PR #2 合并到 `main`，并在测试主机完成首次 Actions 发布。后续合并只自动准备，Owner 仍须在 Actions 手动部署；最新发布状态、版本和效果地址从 GitHub Deployment 或 Agent Platform 的原 Run 页面核对，操作说明见 [本机持续预览](docs/local-deployment.md)。`make verify` 的短命服务和手工预览不算正式发布。

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

1. “上游”保存完整 HTTPS base URL（例如供应商 `/v1` 地址）与凭据。初次和改地址必须有新凭据；相同地址留空保留原值。保存只代表本地提交，不是联调成功。
2. “模型”建立稳定别名与真实模型映射，可编辑映射、停用或重新启用。
3. “应用密钥”签发命名密钥，仅创建时展示完整值，可复制后关闭。遗失后签发新的并停用旧的；查询不找回原值。
4. “调用测试”明确输入应用密钥、别名及消息，手动选择普通/SSE 发送。可能产生真实供应商费用；支持取消，不自动重试。页面离开、退出和取消会清空敏感输入。
5. “调用记录”只显示真实请求元数据、状态、耗时及可空用量；未知不等于零。分页最多 100 条，不记录对话或供应商诊断正文。

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

仅支持 `GET /v1/models` 和文本/函数工具子集的 `POST /v1/chat/completions`。协议清单、来源规则与错误契约以 `docs/workflow/design.md` 第 4–6 节为准：支持常见生成参数、function tools/choice、response_format、stream_options.include_usage，未知字段和多模态转发前拒绝。只替换请求 model；普通结果及合法 SSE chunk 原样返回，响应中的真实 model 不改写。不转发客户端 Cookie/自定义头或供应商 Set-Cookie；不会访问客户端指定目的地址。

默认拒绝 HTTP、私网/loopback/特殊地址、环境代理、所有重定向；配置和实际 Dial 均复验 DNS。管理写操作须精确 Origin，登录后的写操作还需会话 CSRF；管理会话与应用 Bearer 隔离。上游 token AES-GCM 加密，管理员密码 Argon2id，随机应用密钥仅存 SHA-256 摘要。本版只支持 loopback，无公网 TLS/生产部署承诺，同机特权进程不在秘密保护边界内。

生成请求最大 1 MiB、管理 JSON 64 KiB、普通响应 8 MiB、SSE 事件 256 KiB/全流 32 MiB；响应头 30 秒、整个请求 300 秒、有效 SSE 事件空闲 60 秒；逐次下游写 10 秒。全局登录 10 次/分钟，最多两个并行口令验证。每次生成至多一次供应商 POST，不做重试。SSE 只有记录落盘后才发送 `[DONE]`；`event: error` 或缺少 `[DONE]` 均为不完整，不以 finish_reason 代替结束。取消不能保证供应商撤销费用。

## 验证

前提：上述 Go/C 工具链，以及 Node 18+、npm、Playwright Chromium（仅测试依赖）。

```sh
npm ci --ignore-scripts --omit=optional --no-audit --no-fund --cache .cache/npm
npx playwright install chromium
make verify
```

固定 `make verify` 顺序执行工作树代码摘要/HEAD、旧版健康方法真实 HTTP 红灯重建与精确分类、格式、module 校验、vet、前端异步状态回归、race 单元与真实 HTTP/CLI 测试、构建、实际 Go 服务浏览器旅程（1280/390px，包含延迟配置加载屏障）。`make frontend-test` 可单独运行实际 app.js 的隔离状态测试，不代替真实浏览器验收。任何失败都会非零退出，HTTP 或浏览器不可用不算通过。浏览器依赖锁在 package-lock 中，不用于产品后端；受控上游只在测试程序内以指定 authority 放行，正式 CLI 没有测试开关、没有默认假供应商。测试拥有临时数据及随机端口，不使用用户安装或付费凭据。

验证默认只把本次日志输出到 stdout/stderr，不改写已跟踪的历史任务文档或日志。`make health-history` 同样只输出当前重建回执；其成功表示旧版的 21 个方法断言失败与 12 个查询/恢复通过被精确识别，不表示旧版全绿。分类异常或 stdout 写入失败均非零退出，诊断位于 stderr。需要保存时由调用方选择一个新的独立证据位置，例如先设置 `VERIFY_LOG` 为本次新日志路径，再执行 `make verify > "$VERIFY_LOG" 2>&1`；直接重定向保留 make 退出码，不用未启用 pipefail 的 `tee` 管道，不指向旧日志。验证程序没有硬编码本次 Issue、Run 或机器路径；调用方负责目录存在、可写及不覆盖旧证据。

登录后概览的「运行版本」读取当前服务 `/healthz.version` 原始值（包括 `dev`），与静态品牌标签无关。健康请求失败、版本缺失/为空或等待超过 5 秒时显示「版本未知」，不阻塞管理操作；离开再进入概览或刷新页面会重新获取。该显示不证明供应商联调或多实例版本一致。版本回归已纳入同一 `make verify`，无需新运行命令。

配置生命周期回归包括不同模型编辑、新地址/新凭据及旧目标计数、上游禁用无消耗后恢复。两屏网页旅程创建两个不同映射、编辑后正常停止 HTTP 服务并关闭存储、重新打开同一临时安装和启动 Go 服务，再登录核对管理响应与恢复网页；同时检查调用记录及捕获的服务日志不含完整凭据、哨兵对话或敏感上游错误。浏览器 runner 的独立 loopback 生命周期/日志检查服务仅在测试内存在，不是产品 API，也不修改数据库记录。实际进程 SIGTERM/崩溃恢复另由原 Go/CLI 测试覆盖。

研发环境监听受沙箱限制，完整通过回执需已配置 host-only 项目测试节点实际执行；参见 `docs/workflow/delivery.md` 的失败与待核验清单。真实供应商联调未测，受控测试、构建成功及草稿 PR 都不代替独立 QA 或用户最终验收。
