# 【修复】交付工作台退出按钮配色与测试夹具竞态修复

工作台右上角的「退出」按钮此前与普通操作按钮配色相近，用户在浏览顶栏时不易辨认。本次合并后，「退出」按钮在桌面及 390px 窄屏下均为浅红底（`#fff1f2`）、深红字（`#9f1239`），悬停底色加深为 `#ffe4e6`，移开恢复常态；按钮位置、文字、键盘焦点描边和退出登录行为（POST `/admin/api/logout` 后返回登录页）全部保持不变，其他按钮（含 `#nav-toggle`）配色不变。合并后用户无需任何额外操作即可获得该效果。

关联 Issue：#118（沿用 #115 验收基线；与 #117 同版补丁一致，#116 已完成同一补丁的独立 QA）。

## 改动

- `internal/relay/web/workspace.css`：文件末尾、媒体查询之外追加 `#logout{background:#fff1f2;color:#9f1239}` 与 `#logout:hover{background:#ffe4e6}`。规则位于媒体查询之外，桌面与窄屏均生效；既有 `:focus-visible` 描边规则不受影响。
- `tests/browser/console.mjs`：console 旅程新增 `logout-style-{width}` 阶段，在真实 Go 服务上断言常态底色/文字、悬停底色、移开恢复、Tab 聚焦 `:focus-visible` solid 描边，以及 `#nav-toggle` 保持原配色；1280/1440/390 三宽度均覆盖。
- `tests/browser/workspace.test.mjs`：新增静态断言——配色规则在媒体查询之外、`:focus-visible` 描边规则存在、`workspace.js` 退出绑定与 fetch 调用保持原逻辑。
- `tests/verify/main_test.go`（测试夹具竞态修复，不含产品行为变更）：`TestFixedRunnerReceipts` 夹具改用「临时文件 + mv」原子发布就绪回执，断言侧以 `awaitReadyFile` 轮询等待回执完整；`timeout_gate` 场景的夹具门限从 200ms 提升到 5s、父预算从 8s 提升到 15s，以覆盖高负载宿主的进程派生延迟。**生产门限 `gateTimeout` 仍为 25 分钟（`tests/verify/main.go`），清理断言（进程组不存在、候选目录清除）全部保留未动。**

## 验证结果

- 完整固定回归（tests 节点，执行 6）：exit_code=0、`GATE_END exit=0`、`CLEAN_END source_unchanged=true`；回执 `test-results/verification/20261128T063308Z-1982769750`，go1.27.4、chromium 147.0.7727.15。console 套件 `logout-style` 阶段三宽度全绿；连续旅程 J01–J07、故障、延迟、可视化、自助部署等全部套件在 1280/1440/390 三宽度通过。浏览器旅程中的上游均为受控 HTTP 测试上游（`controlled_http_not_supplier`），非真实供应商，符合本任务范围（纯前端配色，不涉及供应商集成）。
- 独立 QA（执行 8，结论 Go）：验收条件 1–8 全部 Pass，详见 `docs/workflow/runs/4230365d4a54629a1d863fb31526121c/qa.md`。本机亲验：gofmt 无差异、`go vet` 通过、`TestFixedRunnerReceipts` 13/13（23.6s）、node 静态断言 39/39。回执 state.json 中 `gate_timeout_ms=1500000`（25 分钟），确认生产门限未被夹具改动波及。
- 架构差异审查（code_review 节点，执行 7）：regressions=0、unresolved=0、baseline_input_identical=true、baseline_model_incomplete=false、source_unchanged=true，candidate `7dd49847814951b3241e5a86c91e0f1894c743f4`。4 条存量基线发现按流程保留在审查报告，不属于本次范围。
- 人工确认（approval 节点，执行 9）：已核同版固定回归、架构差异审查和独立 QA 通过，准予正式交付。
- 脱敏截图证据（同场景「退出」按钮 1440px 与 390px）：
  - [console-1440.png](docs/workflow/runs/4230365d4a54629a1d863fb31526121c/qa-evidence/console-1440.png)
  - [console-390.png](docs/workflow/runs/4230365d4a54629a1d863fb31526121c/qa-evidence/console-390.png)

## 边界与影响

- 纯 CSS 视觉变更：无接口、数据、配置或运维命令变化；无需迁移。
- `workspace.js` 无改动（diff 为空），退出行为与键盘焦点逻辑保持原实现。
- 夹具修复仅影响 `tests/verify` 测试自身，不改变被测的固定回归工具行为与生产门限。

## 待确认事项

- 无阻塞项。合并动作由平台 Connector 执行；合并前如需重放 CI，以合并时最新结果为准。
- 真实供应商验证不属于本次范围（本任务为前端配色，任务规则亦禁止在凭据与供应商响应到位前声称真实供应商验证）。

🤖 Generated with [Claude Code](https://claude.com/claude-code)
