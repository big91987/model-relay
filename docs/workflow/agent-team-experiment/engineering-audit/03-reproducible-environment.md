# 可复现环境与依赖初始化静态审查

## 结论与范围

本次仅读取仓库文件和 Git 元数据，没有执行初始化、下载、构建、测试、服务或页面，也没有读取常驻凭据。当前同版候选为 `122c9abece28f562616483e991098d020d236d6d`；三项前置已依次合入：严格 Bearer 认证 `2d48ee08f12368e1b806fcc0975943e2eca0c8ee`、单请求单次时钟推进 `532e713acc998162c682b310c8f280e1651e6f95`、严格备份清单 `122c9abece28f562616483e991098d020d236d6d`。这是 Git 历史中已验证的候选/合入关系，不代表本审查重新验证了环境或产品。

总体判断：产品 Go 依赖、测试 Node 包和仓库内 Harness Go 依赖均有版本及校验锁；固定 `make verify` 还隔离项目缓存、环境和临时副本，基础较好。但干净机器仍不能仅凭仓库、在无网络或无预制宿主镜像时完成验证：Go 补丁版本、Node/npm、C 编译器、Chromium 可执行文件及系统库没有统一的机器可执行环境清单或安装/校验入口。当前没有 P0；有两个 P1（工具链/浏览器未完整锁定、冷缓存离线不可复现）和若干 P2 文档及供应链可追溯性问题。

状态词：**已验证**仅表示本次静态读取或 Git 元数据核对；**推断**表示由配置行为推导、未执行；**未执行**表示本任务明确禁止的动态验证。历史日志只作为仓库留存对象核对，不升级为本次执行证据。

## 可复现入口清单

| 入口 | 静态依据 | 锁定/隔离能力 | 本次状态 |
| --- | --- | --- | --- |
| 产品 Go 模块 | `go.mod:1-10`；`go.sum:1-6` | `go 1.25.0`，三个模块版本固定并有 `h1` 校验；SQLite 使用 CGO 驱动 | 已验证文件；未执行 `go mod verify` |
| Node/浏览器测试包 | `package.json:1-6`；`package-lock.json:1-49` | Playwright 与 core 均精确为 `1.59.1`，lockfile v3 含 registry URL/integrity；可选 `fsevents 2.3.2` 也锁定 | 已验证文件；未执行 `npm ci` |
| 项目门禁 | `Makefile:1-47` | `GOCACHE` 落在工作树，`CGO_ENABLED=1`、`GOTOOLCHAIN=local`；`verify` 委托 Go 验证器，gates 包含历史、格式、模块校验、vet、前端、race、build、延迟和浏览器 | 已验证入口；未执行任何 target |
| Node 初始化 | `Makefile:36-40,46-47` | 前端和浏览器目标均先执行锁文件驱动的 `npm ci --ignore-scripts --omit=optional --no-audit --no-fund --cache .cache/npm`，失败阻断消费者 | 已验证配置；未下载 |
| 人工从零说明 | `README.md:46-58,120-132` | 说明 Go/C/CGO、Node/npm、Chromium、构建/初始化/服务顺序，以及干净 worktree 不依赖预装 `node_modules` | 已验证文档；动态可行性未验证 |
| 本机持续预览 | `docs/local-deployment.md:1-9` | 受信 Runner 在独立工作树核 main SHA、跑完整门禁并绑定二进制 SHA；部署与准备分离 | 已验证文档；平台私有配置/宿主未检查 |
| OpenSandbox Harness | `docs/harness/opensandbox-entry.md:1-23`；`.harness/agent-platform/SOURCE.json:1-3,804-812`；`.harness/agent-platform/go.mod:1-22`、`go.sum:1-30` | 仓库复制 SDK/模板/工具；源 commit、逐文件摘要和独立 Go 模块锁存在；固定 commit 并要求完整可达历史 | 已验证仓库对象；未核外部维护源或运行 Harness |
| 编排本地状态 | `.trellis/config.yaml:102-108,160-163`；`.gitignore:16-25` | worker 有空闲/数量限额；自动提交关闭、inline dispatch；运行态和私有输入被忽略 | 已验证配置 |
| 发布/入口来源指纹 | `.github/agent-platform-entry-source.json:1-4`、`.github/agent-platform-sandbox-entry-source.json:1-4`、`deploy/platform-preview-source.json:1-4` | 三个入口均记录目标路径与 SHA-256；前两个摘要也出现在 Harness `SOURCE.json:554-557,739-742` | 已验证声明一致性（文本比对）；未重新计算摘要 |
| 持久/生成物排除 | `.gitignore:1-14` | 缓存、数据、二进制、Node 包、测试结果、数据库、密钥、备份和 env 文件不入库 | 已验证规则；未扫描被忽略内容 |

## 正面能力

1. **依赖消费顺序明确。** `verification-deps` 是前端与浏览器的显式先决条件（`Makefile:36-47`），不是依靠全局模块。仓库历史基线还记录该依赖顺序及失败短路测试通过（`docs/workflow/runs/3e7496815b9922704072482cf44e73e2/acceptance/evidence/dependencies-qa.log:1-17`）。这是历史执行证据，不是本次复跑。
2. **验证器有较强环境收敛。** 静态代码显示它在临时目录复制候选，并重设工作树 `GOCACHE`、`GOTOOLCHAIN=local`、`CGO_ENABLED=1` 和项目内 npm user/global config（`tests/verify/main.go:124-192,255-277`）；测试普遍使用 `t.TempDir()`/`os.MkdirTemp` 和 `127.0.0.1:0`，例如 `cmd/model-relay/main_test.go:53,143`、`tests/browser/main.go:50,64,213`。这降低用户数据、固定端口和外部 npm 配置污染。
3. **锁定对象覆盖产品与 Harness 两套 Go 模块。** 产品和仓库内 Harness 都同时保留 `go.mod`/`go.sum`；Node 使用 lockfile v3 和 integrity。Harness 另以 `SOURCE.json` 固定上游 commit 和文件摘要，而非只写“最新版”。
4. **仓库知道沙箱与宿主的边界。** README 明确原生沙箱不能监听、完整门禁交宿主 Connector（`README.md:120-124`）；历史对象也记录 `127.0.0.1:0` 被拒以及 npm 网络失败，例如 `docs/workflow/health-methods-verify-initial.log:41-49`、`docs/workflow/health-methods-browser.log:1-10`。本次没有复跑这些既知故障。
5. **历史环境对象能交叉核锁。** 留存 QA 对象记录 Go 1.25.6、Playwright/core 1.59.1、无 `NODE_PATH`，并区分“QA 开始时 node_modules 已存在、非干净环境”（`environment-qa.log:1-13`）；材料锁也记录 Chromium revision 路径（`material-locks-qa.json:1-10`）。其局限已保留，没有把非干净 QA 环境误称为冷启动证明。

## 隐含前提

- 宿主必须已有兼容的 Go 1.25 工具链；`GOTOOLCHAIN=local` 会阻止自动补齐缺失工具链（`Makefile:3-5`）。README 推荐 1.25.6，但 `go.mod` 只声明 1.25.0 语言/最低工具链线（`README.md:48`、`go.mod:3`）。
- 必须有可用 C 编译器、平台 linker/header，并允许 CGO；SQLite 无外部数据库不等于无原生构建依赖（`README.md:48`、`Makefile:4`）。
- 必须已有 Node 18+、npm 和与 Playwright 匹配的 Chromium 及系统库（`README.md:124-130`）；仓库没有 Node/npm 版本管理文件或 `engines` 字段。
- 冷 Go module cache 需要访问模块代理/源码，冷 npm cache 需要访问 lockfile 中的 npm registry；`.cache/npm` 只是缓存位置，命令没有 `--offline`。Playwright 浏览器安装还是单独的手工网络步骤。
- 动态测试需要允许创建临时目录、子进程、CGO 编译，以及绑定多个 loopback ephemeral ports；受限原生沙箱不满足最后一项，宿主 Connector 才是预期执行面。
- 历史重建依赖完整可达 Git 祖先和系统 `tar`（`tests/health-history/main.go:39-66`、`tests/review-history/main.go:27-50`）；浅克隆必须由平台准备层补足，仓库命令自身不拉历史。
- 本机预览依赖仓库外私有 Runner 配置、环境变量和目标目录（`docs/local-deployment.md:5-7`）；这些合理地不进仓库，但意味着仓库本身不足以重建部署宿主。

## 风险与建议

### P1：工具链与浏览器没有形成单一、机器可校验的环境契约

**影响：** 两台“符合 README”的机器可使用不同 Go 1.25 补丁、Node/npm、C 工具链、Chromium revision/系统库；`GOTOOLCHAIN=local` 又使结果完全取决于宿主预装版本。历史前置任务实际回执使用 Go 1.25.6、Node 22.14.0/npm 10.9.2 或另一个历史 QA 的 Node 26/npm 11.12.1，说明当前文档允许较大漂移。浏览器包虽锁为 1.59.1，但二进制只在历史对象中出现为宿主缓存路径，不是仓库环境契约。

**依据：** `go.mod:3`；`README.md:48,124-130`；`package.json:1-6`；`material-locks-qa.json:5-10`；前置任务控制器回执（本任务 `get_task` 返回值，2026-11-01）。

**建议归属：仓库基建。** 增加受版本控制的工具链清单（至少 Go 精确补丁、Node、npm、Playwright、Chromium revision、支持 OS/arch、C 编译器最低线）和只读 preflight；CI/Connector 镜像以不可变 digest 绑定同一清单并在门禁开头打印/校验。保留 `GOTOOLCHAIN=local`，但让版本不符快速失败。不要仅靠个人版本管理器文件，因为 C/浏览器系统依赖也需覆盖。

### P1：冷缓存、离线和镜像恢复前提没有闭环

**影响：** lockfile 保证下载内容身份，不保证内容可获得。新环境没有 Go module/npm/Chromium 缓存时需要外网；仓库既无离线 bundle/SBOM 制品说明，也无缓存预热、校验、保留期和失效恢复契约。历史 `ENOTFOUND`/`EPERM connect` 已显示网络受限会在产品测试前失败，但本次未复跑。

**依据：** `Makefile:36-47`；`package-lock.json:13-49`；`README.md:124-130`；`docs/workflow/health-methods-browser.log:1-10`；`docs/workflow/development-version-browser.log:1-27`。

**建议归属：通用平台 + 工程工具。** 平台提供按 lockfile/go.sum/浏览器 revision 键控、带校验和与保留策略的只读依赖缓存或不可变验证镜像；项目 preflight 明确报告 cache hit/miss 和网络需求。若要求真正离线验证，应产出可审计依赖包及来源清单并提供 `GONOSUMDB/GOPROXY`、npm cache 和浏览器路径的受控注入方案；在此之前文档应明确“冷启动需要网络”，不要声称离线可复现。

### P2：从零初始化说明分散，缺少无副作用的统一诊断入口

**影响：** README 同时混合产品运行、测试依赖和沙箱例外；操作者要到 Harness/部署文档才能理解完整历史、私有配置与执行面。发现缺 C 编译器、Node 版本、Chromium 系统库或 loopback 权限时，通常已进入门禁。

**依据：** `README.md:46-58,120-132`；`docs/harness/opensandbox-entry.md:5-13,23`；`docs/local-deployment.md:5-9`。

**建议归属：仓库流程。** 增加一个不安装、不监听、不读取凭据的 `make doctor`/`scripts/preflight`，输出支持矩阵、精确版本、临时目录可写性和“端口能力仅待宿主门禁验证”；README 给出 fresh clone → preflight → dependency provisioning → verify 的唯一顺序，并链接平台缓存/镜像契约。

### P2：来源指纹可检测漂移，但元数据语义不一致且缺少仓库内校验说明

**影响：** Harness `SOURCE.json` 有上游 commit、bytes、mode、sha256；三个 `*entry-source.json` 只有 path/sha256，部署来源文件甚至没有维护源 commit。人工可比对摘要，却不清楚哪个固定入口负责重新计算、漂移时如何失败。

**依据：** `.harness/agent-platform/SOURCE.json:1-7,554-557,739-742`；`.github/agent-platform-entry-source.json:1-4`；`.github/agent-platform-sandbox-entry-source.json:1-4`；`deploy/platform-preview-source.json:1-4`；`docs/harness/opensandbox-entry.md:17-21`。

**建议归属：工程工具。** 统一 source manifest schema（来源仓库/commit、目标 path、mode、bytes、sha256），提供纯读取校验命令并纳入固定证据门禁。部署入口也应记录维护源版本。本次未重新计算摘要，因此这里只评估结构与文本一致性。

### P2：Harness 是第二套构建生态，但新环境说明主要覆盖产品模块

**影响：** `.harness/agent-platform` 有独立 Go 依赖和 Python SDK/脚本；产品 README 的 Go/Node 准备不能自动说明 Harness 的 Python、系统命令和运行要求。平台升级者可能把“产品可构建”误当作“入口可重建”。

**依据：** `.harness/agent-platform/go.mod:1-22`；`.harness/agent-platform/SOURCE.json:804-832`；`docs/harness/opensandbox-entry.md:1-3,17-23`。

**建议归属：通用平台。** 由维护源发布 Harness 运行时兼容矩阵和不可变制品/镜像摘要；仓库只保留所用源 commit、导出摘要和升级差异，不在产品 `make verify` 中隐式安装平台 SDK。

## 沙箱、临时目录与端口判定

- **已验证（静态）：** 产品验证器和测试广泛使用系统临时目录/`t.TempDir`；项目缓存写入工作树 `.cache`，生成二进制写 `bin`，均被 `.gitignore` 排除。测试监听以 `127.0.0.1:0` 为主，避免固定端口冲突。
- **推断：** 文件沙箱至少需允许写工作树 `.cache`、`bin`、`test-results` 和系统临时目录，并允许临时副本执行子进程；浏览器测试会同时占用多个 loopback 端口，不能用“允许一个固定端口”替代。
- **历史已知、未复跑：** 原生沙箱曾对 `127.0.0.1:0` 返回 `operation not permitted`；npm 也曾因 DNS/代理连接受限失败。README 已将完整验证移交宿主固定 Connector，这是正确分层，但平台仍需把该能力作为环境契约而非偶然宿主状态。

## 未执行与未验证

- 未运行 `go env`、`go mod verify/download`、`npm ci`、Playwright 安装、任何 `make`、测试、构建、服务、页面、监听探测或摘要重算。
- 未确认当前机器实际 Go/Node/npm/C/Chromium 版本、缓存内容、网络、临时目录容量/挂载选项或端口权限。
- 未读取仓库外私有平台配置、token file、Runner 镜像、部署目录或凭据；未访问维护源、registry、module proxy、GitHub/Actions/Deployment。
- 三项前置任务的绿色 `make verify` 回执用于候选背景和版本交叉核对，不是本次静态审查的执行结果，也不证明新机器冷启动、离线重建或真实供应商联调。

因此，本任务可标记 completed（静态审查及报告完成），但环境与产品验证状态保持“未由本任务执行”。
