# 管理台概览显示当前运行服务版本，失败时局部降级并可重试

建议草稿 PR 标题：**feat: 在登录后概览展示运行版本并支持异常降级与重试**

## 背景、原问题与预期

管理员需要确认测试环境当前运行的是哪一版服务。原管理台顶部只有静态「v0.1 · 团队模型网关」品牌标签，概览没有运行实例版本，不能据此判断升级结果。这是现有管理台的新增展示能力，不是从零实现中转服务，也不是修改服务版本来源。

依据 `docs/workflow/task.md` 的已确认行为和验收目标，用户已在原 Issue #9 的接续会话确认：仅登录后概览展示现有 `/healthz.version` 原始值（包括 `dev`）；不可用时显示「版本未知」，不阻塞其他管理操作，重入概览或刷新可重试。预期解决静态标签无法反映运行版本的问题，而不是证明构建标记、生产升级或多实例一致性正确。

需求为本次 `requirements.md` AC-01～08，设计为本次 `design.md`。研发记录见 `delivery.md` 的 Issue 9 章节；独立 QA 见 `qa.md` 开头的本次版本展示报告。旧 v0.1 记录及旧日志不是本次验收依据。

结论：**独立 QA Go，允许交授权 Connector 提交、推送并创建草稿 PR，最终合并由人决定；不部署。** 本报告不表示草稿 PR 已创建或原 Issue 最终回复已完成。

## 具体功能与行为

- 登录或恢复有效会话进入概览，在概览标题下显示「运行版本」；登录前不展示、不发起该健康请求，其他管理页与退出后不显示新增版本区域。原静态品牌标签保持不变。
- 进入概览独立 GET `/healthz`，等待时显示「加载中」；有效非空字符串按原文展示，保留前后空白，包含 `dev`，不从静态标签或 GitHub 推断版本。
- 网络失败、非成功 HTTP 状态、无效 JSON、版本缺失、空字符串、仅空白或非字符串均显示准确文案「版本未知」，不保留上次成功值。5 秒超时涵盖响应头和响应体读取。
- 版本获取不阻塞概览其余信息或导航，也不把健康请求的 401 当作管理会话失效；真正管理 API 的 401 仍沿用退出登录行为。
- 离开再进入概览或刷新页面重新获取，可从失败恢复；切页、退出开始、实际会话失效都会使旧请求失效，迟到成功或失败不能覆盖新版本或清理新请求的计时器。
- 版本仅通过 `textContent` 显示，标记不执行；长文本可折行。没有轮询、持久化、专用刷新按钮、配置写入或供应商调用。

## 关键实现取舍与变更范围

| 文件 | 本次变更 |
|---|---|
| `internal/relay/web/index.html` | 概览增加只读版本文本，带 `role="status"` 与 `aria-live="polite"`。 |
| `internal/relay/web/app.js` | 独立 `cache: no-store` 健康请求、专用 AbortController、5 秒全请求超时；请求身份、导航快照、当前页面及登录状态联合限制写回。 |
| `internal/relay/web/style.css` | 版本文本使用 `white-space: pre-wrap` 与 `overflow-wrap: anywhere`。 |
| `internal/relay/version_test.go` | 新增进程内健康原值契约与真实 HTTP 的 dev/发布字符串回归。 |
| `tests/browser/ui.test.mjs` | 总计 24 项确定性前端回归，含原有 4 项；覆盖异常输入、重入恢复、超时 headers/body、迟到响应与退出/管理401。 |
| `tests/browser/journey.mjs`、`tests/browser/main.go` | 在既有真实 Go 服务旅程加入版本检查，夹具明确注入 dev；覆盖 1280px/390px。 |
| `README.md` | 说明版本来源、降级、恢复及验证边界。 |

沿用现有健康接口和启动版本链，不修改生产 Go 后端、数据库、依赖锁或 Makefile 门禁。独立于管理 API 包装和调用测试控制器，避免局部健康失败触发全局错误或取消其他操作。新增测试进入原固定 `make verify`，没有用静态页面检查替代真实后端。

## 实际阶段与裁剪

实际经过：平台分支/Issue 准备 → 入口与用户关键澄清 → 精简需求 → 局部设计 → 实现与回归 → 宿主项目测试 → 独立 QA → 本交付说明。当前待执行为 Connector Git 发布与草稿 PR，Issue 自然语言进展和最终回复由平台 Hook 自动回写。

依据 `task.md`，没有跳过必要业务阶段；澄清已完成，不重复访谈、阶段批准或前序分支/Issue 操作。只读局部功能不适用组织级 PRD、额外平台运营设计、原型/PoC；没有为这些不适用阶段补造文档。生产部署和真实供应商联调明确不在本次范围。

## 实际验证与证据

本轮平台 **tests 节点 seq 7** 的宿主固定 `make verify` 回执为 **exit 0**。完整输出在平台回执，仓库摘录为 `docs/workflow/qa-version-host-receipt.md`，不能将摘录称为 QA 新执行的完整日志。

| 实际门禁 | 结果 |
|---|---|
| format-check、go mod verify、go vet | 通过，未削弱原门禁。 |
| 前端 Node/VM 回归 | 24/24 通过，0 fail、0 skipped；这是确定性异步测试，不冒充真实浏览器。 |
| `go test -race -count=1 -v ./...` | 全部通过，包括既有实际 HTTP/CLI 回归、新健康原值契约3个子项及真实 HTTP 的 dev/发布版本2个子项。 |
| `go build -trimpath` | 通过。 |
| 真实 Go 服务 Playwright 旅程 | 1280px 与 390px 均通过；真实未拦截健康值与 UI dev 一致、刷新、401降级/重入恢复、健康挂起时模型读取、文本安全/长文本、退出边界及既有管理旅程。异常和文本夹具由 page.route 显式注入。 |

独立 QA 本轮 seq 8 结论 **Go**：AC-01～07 与 AC-08 当前阶段责任通过；原 Issue/草稿 PR 是后续平台职责，未冒充已完成。QA 沙箱另跑前端24/24、进程内健康契约3/3和 diff 检查通过，输出保留于 `docs/workflow/qa-version-local.log`；不是再跑一套完整 HTTP/UI 门禁。

被测分支为 `workflow/509e5c927cb88a6bbcefa274656ac6d2`，提交前 HEAD 为 `9b6994ed80de2bdba79ff267419150dccaac1360`。实现尚在工作树，不能仅用 HEAD 代表被测代码。宿主、独立 QA 与本报告节点再次执行 `go run ./tests/evidence` 的33文件摘要均为：

```text
f96017d3daf52e8372e1ef106397bfb31224e049010bb53a6cd033d69f27d414
```

摘要边界见 `tests/evidence/main.go`：产品、测试、构建锁、README、task/requirements/design；排除 delivery、QA、本报告与回执以免自引用。本报告节点只修改 `pr.md`，不改代码、测试或基线，不声称重新执行宿主完整门禁。

失败证据不覆盖、不删除：`development-version-verify-initial.log` 记录修复前健康 Host/Authority 夹具403和沙箱监听拒绝；夹具修正后当前匹配版本的契约/宿主 HTTP通过。`development-version-verify.log` 的研发沙箱完整入口仍为 exit 2，串行 build/browser 未到达；`development-version-targeted.log` 为独立契约/build/语法成功；`development-version-browser.log` 为依赖安装 EPERM、`development-version-browser-runner.log` 为监听 EPERM，均未执行浏览器断言。后续同摘要宿主成功覆盖受限项，不将早期失败写成通过。macOS LC_DYSYMTAB 链接警告保留，未导致失败；`storage_write_failed` 为既有故障注入用例输出，相应测试通过。

## 已知限制

- 只展示当前响应服务自报版本；`dev` 如实展示，不保证构建标记正确、部署到指定发行版或负载均衡下各实例一致。不提供版本编辑或升级控制。
- 非阻塞指版本请求不会额外阻塞管理操作，不承诺全服务故障时其他管理接口仍成功。无轮询，状态更新需重入概览或刷新。
- 异常输入、全部超时和竞态由 VM 分层覆盖，未逐项另跑真实浏览器；真实浏览器已验证正常、失败恢复和非阻塞路径。
- 没有成功截图/视频、人工读屏、无障碍全量验收、生产升级验证或真实供应商验证。既有受控测试上游仅用于隔离回归，不证明供应商可用；没有付费调用或部署。

## 试用方式

在已有测试安装中使用原参数启动服务、登录 `/admin/` 并查看「概览」的「运行版本」。普通开发构建默认显示 `dev`；可在同一监听地址 GET `/healthz`，核对 `version` 与页面一致。切换到模型页再返回，或刷新并恢复有效会话，会重新读取；不要以顶部静态 v0.1 标签作为运行版本。

没有现有安装时，沿用 README 的隔离初始化方式（Go 1.25+、CGO/C 编译器）；使用新的目录，不覆盖用户已有安装，不将密码或主密钥写入仓库：

```sh
make build
bin/model-relay init --data-dir ../relay-data --key-file ../relay.key
bin/model-relay serve --data-dir ../relay-data --key-file ../relay.key --listen 127.0.0.1:5544
```

浏览器访问 `http://127.0.0.1:5544/admin/`，使用初始化时一次展示的管理员密码。保持 Host 与监听 authority 一致，不将地址换成 localhost。本次查看版本无需配置供应商或发送模型请求。

失败/恢复、超时/迟到、安全文本和双视口的自动试用纳入 `make verify`；浏览器测试依赖准备见 README。受限 Agent 沙箱无法监听时交固定宿主测试 Connector，不提权、不修改业务数据凑通过。

## Connector 交接

走 `next → publish`，沿用上述具体 PR 标题和正文。授权 Connector 负责提交当前工作树、推送并创建草稿 PR；合并由人决定。本节点不自行 GitHub 读写、提交、切分支、推送、创建/合并 PR、启动 Workflow 或部署。

发布提交后 HEAD 自然变化，保留提交前 HEAD 与内容摘要关联。若交付前产品、测试、构建锁或验收基线变化，必须重新验证并交独立 QA，不能继续沿用当前摘要的 Go。平台共享 PR 标题修复及安全重启是用户提供的恢复上下文，不是本报告执行或验收的产品变更。
