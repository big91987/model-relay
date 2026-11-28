# 研发回执：交付工作台退出按钮配色与已证测试夹具修复

关联:Issue #118;基线 `70a4f55b5cfd737e0882936ab8f65bbd66fcaed2`(落地前已核对 HEAD 一致)。

## 1. 改动内容

复用 #117 已完整回归、#116 已独立 QA 的同一补丁版本,`git apply --check` 校验上下文后干净应用,未做任何修改:

- `internal/relay/web/workspace.css`:文件末尾、媒体查询之外追加 `#logout{background:#fff1f2;color:#9f1239}` 与 `#logout:hover{background:#ffe4e6}`。
- `tests/browser/console.mjs`:新增 `logout-style-{width}` 阶段,断言常态/悬停/移开恢复/Tab 聚焦 `:focus-visible` 描边及 `#nav-toggle` 配色不变。
- `tests/browser/workspace.test.mjs`:新增静态断言(配色规则在媒体查询之外、焦点描边规则存在、退出绑定与 POST /admin/api/logout 保持原逻辑)。
- `tests/verify/main_test.go`:`TestFixedRunnerReceipts` 夹具就绪竞态修复——回执 tmp+mv 原子发布、`awaitReadyFile` 轮询、`timeout_gate` 夹具门限 200ms→5s、父预算 8s→15s;生产门限 25 分钟与全部清理断言保留。

## 2. 定向验证结果(本机,GOCACHE 指向沙箱可写目录)

| 检查 | 结果 |
|---|---|
| `git apply --check`(补丁与当前主干上下文一致) | 通过 |
| `gofmt -l .` | 无输出,格式合规 |
| `go vet ./tests/verify/` | 通过 |
| `go test ./tests/verify/ -run TestFixedRunnerReceipts -count=1` | 13/13 场景通过(含 `cancel_gate`、`timeout_gate`),22.26s |
| `node --test tests/browser/workspace.test.mjs` | 39/39 通过(含新增退出按钮配色静态断言) |

## 3. `make verify` 本机结果与沙箱限制

`make verify` 执行到 `go run ./tests/review-history` 时失败,报错 `review historical HTTP rejected: historical listener unavailable; not business red`。该工具自身将此归类为环境不可用(本沙箱禁止本地端口监听)而非业务红。此前同补丁在 #117 已完整回归通过;本次完整回归由固定节点在宿主机环境执行,本机沙箱限制按工程规则单独保留,不作为业务失败证据。`make verify` 中先于该步骤的 `architecture-check` 与源码证据核对均通过(conformant)。

## 4. 交接结论

实现完成,补丁与已验收版本逐字节一致,定向检查全部通过;完整回归、浏览器实际 HTTP 旅程交固定回归节点。无未决设计问题。
