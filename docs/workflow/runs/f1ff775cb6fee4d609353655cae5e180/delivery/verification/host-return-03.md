# 宿主密码显隐失败：原生事件同步与诊断

2026-10-06；本 Run 研发返工，HEAD d62cd51e8de2127548e1ac7469d349443d1bfe3d。验证对象为当前工作树，不是干净 HEAD。结论：完整门禁 Failed（seq9），新增测试同步/诊断待宿主复验；不宣称产品缺陷已修复。

## 新宿主证据与范围

seq9 固定 `make verify` 退出2。尾部当前 Go `internal/relay` race/HTTP 测试 PASS（112.913s），health-history 分类器测试 PASS，build 完成；浏览器阶段退出1。首个失败为1280px、`password hidden with accessible action and state`、`journey.mjs:43:10` 的密码 type 严格断言；只见两次 GET session 401。该检查点被多次调用，回执没有实际 type 或触发动作，不能确定是 Space/点击调度还是产品行为缺陷。历史旧版本的预期红测不是当前失败根因。seq7/seq9 与前序日志保留，不改写为通过。

恢复现有 PRD、架构、交互、公共任务和 Trellis 原生 DOM/验证/所有权规范；packages 实际输出 single-repo、backend/frontend，已有真实规范，bootstrap 不适用。范围只涉及浏览器密码动作同步/失败诊断和公共记录，不修改产品、鉴权、Go、CSS、锁文件、make verify 或平台。T002/T003 仍 In Review，M02/M03 Ready for Review，未 Accepted；无需退需求或设计。

## 实施与未确定项

`tests/browser/journey.mjs` 在原生按钮上加仅测试使用的 click 计数：每次 Enter/Space/click 后最多等2秒观察对应事件；随后严格断言恰好一次，以及原有 type、AX、状态、焦点、值不变、零提交/验证/网络/存储。没有重试点击、跳过原断言或伪造事件。此同步避免把键盘命令回执直接当作 DOM 动作已经完成；是否为 seq9 根因仍须宿主确认。

失败附加白名单诊断：空/受控输入阶段（不含值）、动作名、动作计数、预期显隐、实际 type 枚举、焦点布尔和数字事件计数。不输出凭据、响应正文、DOM/AX完整树、原始 error.message，不拍带秘密的失败截图。即使事件等待失败，仍取状态并执行原断言，不把超时作为通过。

`.trellis/spec/frontend/verification.md` 保存此契约。产品 `passwordVisibility` 仍保持原节点和值，由原生 onclick 驱动，clearSecrets/signedIn/signedOut 的清密不变；无证据支持新增产品 keydown 或认证补丁。

## 本节点实际验证

- `node --check tests/browser/journey.mjs` 退出0；58项既有 VM 回归退出0，见 [VM回执](password-probe-vm.log)。不是原生浏览器通过证据。
- 本节点尝试无监听的 `chromium.launch({headless:true})` 后关闭，退出1：`browserType.launch: Target page, context or browser has been closed`；Chromium进程启动后退出。未加载页面、未运行密码旅程、未生成截图。不能据此诊断 seq9（宿主曾成功启动 Chromium）。省略本机绝对可执行路径和临时profile路径，不存储凭据。
- 本环境既有监听限制未变，不重复必然受限的 `make verify`、`make browser` 或 Go HTTP 监听。前次真实失败在 [expiry-watchdog.md](expiry-watchdog.md)、[host-return-01.md](host-return-01.md) 保留；它们不覆盖本次新增测试。完整新代码门禁明确留给宿主。
- 无监听分项与当前指纹见 [checks-password-probe.log](checks-password-probe.log)、[source-password-probe.json](source-password-probe.json)。Trellis check 核对范围、原断言、秘密白名单、DOM/生命周期复用；不新增跨层协议。公共审查保持原任务编号和关键路径、权威需求/设计不变。
- 实际 `make format-check vet build`、JS语法及 `git diff --check` 各退出0；5份当前公共交付文档的36条本地引用存在检查通过。56文件当前指纹为 `4387d20a8cfcbbbe57f01d67541a901de7d8d0279377cf6b21cf1d3a25c7208c`（Go1.25.6），不同于seq9的 `92548557…`：差异为本轮旅程与验证规范，不沿用旧通过/失败当作新代码完整验证。

## 下一节点

tests 原样执行 `make verify`。若失败，新白名单诊断定位具体密码动作和观测状态，正常返工，不能称此轮已经修复原生显隐；若通过，再由独立 QA 对照18AC判断真实Go全UI首次配置、两屏、HTTP/SSE/晚到隔离等。保留新完整失败尾部和当前指纹；供应商 Not Run，QA/PR/合并/部署未发生，路线2–4未实现。
