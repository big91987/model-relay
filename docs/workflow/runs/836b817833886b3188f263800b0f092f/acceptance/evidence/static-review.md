# QA 独立静态与证据审查

2026-10-06，实际读取本轮 PRD、架构、研发任务/验证与既有 Go/前端/browser/evidence/health-history 源码和差异；完整读取挂载 validating-platform-releases-cn/SKILL.md、references/acceptance-sop.md、references/artifact-contracts.md，采用首次全量（薄切片+公共回归），不套组织级平台验收。

## 实现契约
- HTML 原密码输入 type=password、required 与 autocomplete=off；显式 label for；切换为原生 type=button，aria-controls/password、aria-describedby/status、动态动作名称、role=status/live=polite。没有固定名称与动态文字矛盾，也没有混用 aria-pressed。
- app.js 唯一显隐事实源是原 input.type；passwordVisibility 只写 type、按钮文字、状态说明，不复制 value、不请求网络、不提交表单、不校验、不保存偏好。只处理 onclick，无 Enter/Space 双处理。signedIn 清空/隐藏；clearSecrets 清空/隐藏；signedOut、API 401 沿用现有入口。没有认证/Go/数据库/模块锁/Makefile 修改。
- CSS 仅新增切换按钮右侧间距，沿用既有可见焦点和自适应。具体可达性/尺寸依赖真实旅程，不从 CSS 阅读推成 UI Pass。
- ui.test.mjs 验证原节点/值/空值、多次切换、不 fetch、登录/退出/错误/恢复清理以及管理/版本竞态；VM 不实现浏览器原生语义，已与真实证据分层。
- journey.mjs 独立核对了真实 Tab/Enter/Space/click、focus、button/status AX snapshot、submit/invalid/全部请求计数、storage、两态正常认证、错误恢复、退出/reload、有效 session/重启、1280px/390px、完整管理/版本/日志链；秘密比较用布尔结果，AX 只截按钮和说明。故障 health route、重启 control 与受控 upstream 为明确夹具，不把供应商或生产写成已验证。

## 证据真实性和 Trellis
- Makefile 未修改、没有跳过/错误吞掉；Go browser 启动真实 relay、隔离临时状态和随机端口，Node 非产品后端，退出码透传；宿主退出 0 可对应完整串行门禁。health-history 对指定旧版故意失败进行分类，拒绝沙箱错误；因此本地失败没有被误当作正常旧版红灯。
- 宿主 evidence 摘要与 QA 独立使用项目 GOCACHE 的摘要均为 0f6ee2d6894067323697f632231ee0d62b098be0261a4da7e975edb6e1efb37b；HEAD 73681a7d 相同。最终源码哈希与研发 final-integrity 对应。新 Run 文档与规范不在这 36 文件摘要内，额外 SHA-256 位于 final-integrity.log。
- 当前 .trellis/spec 共 8 份基于 Go/原生 DOM/真实浏览器的规则，非空白框架模板；前端显隐状态与测试契约已同步。研发验证记录区分实际 bootstrap、packages、before-dev/update-spec/check 阅读与执行和未跑门禁，不声称 QA 重新执行研发流程。本轮 QA 不改 Trellis 业务规范。
- Git/Issue 发布和 PR 由平台后续处理，QA 未自行评论、提交、推送、部署或调用供应商。历史文档/日志按哈希而非过去的 Pass 保护；本轮新旧重建证据均保留，未替旧任务写新结论。
