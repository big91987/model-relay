# 修复 /healthz 方法边界：仅接受 GET/HEAD，其他方法返回 405

建议草稿 PR 标题：**fix: 限制 /healthz 为 GET/HEAD 并保留故障状态与 HEAD 空体**

## 背景、原故障与预期

运行中的 Model Relay 原先仅按路径匹配 `/healthz`，未限制 HTTP 方法。健康状态下 `POST /healthz` 也返回 200，非查询请求被错误地当作健康检查；存储故障时这些方法返回 503，且没有明确允许的方法列表。

本次仅修复该接口的方法边界。验收依据为 `docs/workflow/task.md` 的“目标、现象与预期”和“范围与验收依据”：GET 保留原 status/version JSON 及存储故障语义；HEAD 在相同健康状态下与 GET 同状态码且无响应体；所有非 GET/HEAD 方法返回 405，`Allow` 精确为 `GET, HEAD`。不改其他路由或 Host 校验，不引入供应商调用，不部署。

当前独立 QA 结论为 **Go，QA-H8-01 Closed**，可交 Connector 创建草稿 PR，最终合并由人决定。本文不表示 PR 已创建，也不表示生产或真实供应商验收通过。

## 具体行为与实现取舍

| 请求 | 健康 | degraded / 数据库不可用 |
|---|---|---|
| GET /healthz | 200，原 status/version JSON | 503，原 storage_unavailable 错误 JSON |
| HEAD /healthz | 200，空响应体 | 503，空响应体 |
| 非 GET/HEAD 方法 | 405，Allow: GET, HEAD | 405，Allow: GET, HEAD |

- `internal/relay/server.go` 仅修改健康检查分支：先判断方法，再查询健康状态，使非查询方法不受存储故障影响。原 Host 校验仍在路由前执行，上表以通过该校验的请求为前提。
- GET/HEAD 共用 degraded/数据库 Ping 的状态判断；HEAD 显式写入 JSON 内容类型和状态码后返回，不依赖 HTTP 服务自动丢弃响应体。GET 的成功载荷与故障错误契约不变。
- `internal/relay/health_test.go` 新增真实 HTTP 和辅助 handler 两层回归。三种状态各覆盖 GET、HEAD、POST、PUT、PATCH、DELETE、OPTIONS、TRACE、CUSTOM 以及恢复后 GET/HEAD，每层 33 个叶用例；使用隔离临时安装和随机端口。
- `tests/health-history/main.go` 与 `main_test.go` 补充隔离旧版网络复现及判定器。只读归档入口提交，复制最终回归，既不回滚正式源码，也不修改数据库来制造通过。
- `Makefile` 将 `health-history` 加入固定 `make verify`，保留原格式、依赖、vet、前端、全量 race、构建和实际 Go 服务浏览器门禁。旧版子测试 exit 1 仅在全部精确预期断言齐全时被视为复现成功；监听、构建、超时、race、skip 或额外断言失败不能冒充红灯。

## 实际经过的阶段

入口判断 → 开发 → 宿主项目测试 → 独立 QA → 补证开发 → 宿主完整复验 → 独立 QA 最终复验 → 本次交付说明。

本次 requirements/design 阶段按用户明确授权跳过：问题、范围、预期和验收入口明确，无新增产品或架构决策。仓库已有同名文档属于首版建设，不作为此次修复需求/设计基线；未补造占位 PRD 或设计。研发记录保存在 `delivery.md`，其中“等待宿主测试”等为当时状态；最终结论以 `qa.md` 顶部最终复验和下述 seq 9 回执为准。

原开发先加入真实 HTTP 回归，但沙箱在监听阶段失败，没有发送请求；辅助 handler 在旧版抵达了错误断言。QA 首轮因此阻塞 QA-H8-01。后续按照 QA 整改方案在宿主隔离重建旧业务版本，取得真实网络红灯并再次运行完整门禁，QA 才关闭缺口。**旧版网络证据是修复后重建补证，不是原修复前当时成功运行的记录**，不能把实际顺序写成完美的先红后绿。

## 实际验证与证据

提交前 HEAD：`9b6994ed80de2bdba79ff267419150dccaac1360`。实现尚在工作树，不能仅凭 HEAD 标识被验代码；35 文件摘要为：

```text
edb4224576eb1ef488c33d9e8ab354bdba90a8a39668bb75f0c548a616f7dd48
```

摘要由 `tests/evidence/main.go` 计算，覆盖产品、测试、构建锁、README 和验收/设计文件，排除交付报告与日志避免自引用。该值与 seq 9 宿主回执、最终独立 QA 及 `health-methods-history-version.json` 一致。本报告节点使用项目本地缓存重新计算得到相同摘要，`git diff --check` 成功；首次直接计算因默认用户缓存不可读失败，改用项目已有 `.cache/go-build` 后成功，未扩大权限。

| 验证 | 真实结果与依据 |
|---|---|
| 旧版真实 HTTP 复现 | seq 9 与 `health-methods-history.log`：healthy POST 实际 200/预期 405、Allow 空；三状态共 21 个非查询方法失败，42 条 status/Allow 诊断，12 个查询/恢复用例通过；子 go test exit 1，判定器输出 PASS expected historical HTTP red |
| 当前完整 make verify | 平台 previous_results seq 9，tests Connector，command exit_code 0；串行包含 evidence、历史复现、格式、module、vet、前端、全量 race、build、browser，无忽略失败 |
| 当前方法契约 | 真实 HTTP 三状态 33 叶用例由完整门禁验证；GET JSON/503、HEAD 同状态空体、七种非查询方法 405/Allow 及恢复行为全部覆盖 |
| 独立 QA | `qa.md` 最终结论 Go；`health-methods-qa-recheck.log` 记录摘要、diff、旧版/测试哈希、完整历史事件独立解析、判定器正常及 14 类拒绝路径 race 测试、handler 33 叶用例成功 |
| 实际服务 UI | 完整入口包含真实 Go 服务的 1280px/390px 连续浏览器旅程，使用受控测试上游，不是静态页面或真实供应商联调 |

seq 9 会话 stdout 在历史事件段后截断，未逐行看到本轮 test/build/browser 后半段输出。完整成功依据是真实整体 exit 0 和独立 QA 核对的串行 Makefile/脚本错误传播，不捏造截断行。seq 6 未截断回执提供同一未变更浏览器资产的两视口明确 PASS，但其旧 33 文件摘要不替代 seq 9 的新摘要验收。

旧版来源：入口提交归档 SHA-256 `e8307563c976a77940851da8254d70b8255fa4caa690d898a6452ee15de4eaaf`；旧 server.go `d61e4e747231654aa7895404b94c2539497cef2074186b3bbe05b90a94252383`；最终 health_test.go `15b13fbd3cb3265c8bedb9a34f78980fea195351e458f5a4b83ca675513f7c3e`，均由独立 QA 核对。

失败证据未删除：`health-methods-before.log` 保留最初 HTTP 监听阻塞与 handler 旧版失败；`health-methods-handler-recovery-failure.log` 保留测试安装锁错误；`health-methods-verify-final.log`、`health-methods-browser.log`、`health-methods-browser-direct.log` 保留沙箱 bind/DNS 失败；`health-methods-history-verify-initial.log`、`health-methods-history-sandbox.log`、`health-methods-history-verify.log` 保留补证初次判定失败及监听阻塞。它们不算产品通过，后续宿主成功也不改写原失败。

## 已知限制

- 事后重建补证的时序限制永久保留；旧版真实 HEAD 空体本来由 Go HTTP 层抑制响应体，不能将 handler 的旧版 HEAD body 失败混称网络 HEAD 失败。
- 历史门禁依赖上述基线提交仍可被 `git archive` 读取，浅克隆缺失该对象时应补齐历史，不得跳过门禁。
- 原生 Agent 沙箱不能监听，网络/浏览器验收由固定宿主 Connector 执行；未生成截图、trace 或单独宿主进程盘点，不声称另做这些检查。
- macOS LC_DYSYMTAB 链接警告有真实记录，未导致宿主整体失败；历史判定器只允许该精确已知警告，不泛化忽略构建错误。
- 无真实供应商、生产部署、性能/SLA 或额外 UI/读屏验收。本次不新增用户界面、权限、配置、依赖或数据迁移。

## 试用方式

前提为项目 Go 1.25 工具链、CGO/C 编译器。使用全新的独立数据目录和密钥文件，避免覆盖现有安装；初始化密码只向终端展示，不保存到仓库日志。运维注意事项见 `README.md`。

```sh
make build
bin/model-relay init --data-dir ../relay-health-demo --key-file ../relay-health-demo.key
bin/model-relay serve --data-dir ../relay-health-demo --key-file ../relay-health-demo.key --listen 127.0.0.1:5544
```

在另一个终端检查，不需要供应商凭据或发送生成请求：

```sh
curl -i http://127.0.0.1:5544/healthz
curl -I http://127.0.0.1:5544/healthz
curl -i -X POST http://127.0.0.1:5544/healthz
```

预期依次为 200 + status/version JSON、200 + 无响应体、405 + `Allow: GET, HEAD`；不要把 authority 改为 localhost。Ctrl-C 停止服务。degraded/数据库不可用及恢复由隔离自动回归验证，不建议破坏试用数据库来手工模拟。

完整复验使用 `make verify`，浏览器测试需项目锁定的 npm 依赖和 Playwright Chromium，安装方式见 README。受限沙箱沿用宿主验证 Connector，不删除历史门禁、不以静态页面测试替代。

## 后续 Connector 交接

走 `next → publish`，由授权 Connector 回写 Issue #8 的真实证据、提交当前工作树、推送并创建草稿 PR，使用本文具体修复标题及正文，最终由人审阅合并。保持不部署、不触发 GitHub Workflow、不扩大访问权限。本节点只整理交付说明，未创建/合并 PR 或执行 Git 发布。

提交会改变 HEAD，但不改写提交前验收事实。若发布前改变产品、测试、构建或验收资产导致内容摘要变化，必须重新执行固定 `make verify` 并交独立 QA，不能沿用本次放行。
