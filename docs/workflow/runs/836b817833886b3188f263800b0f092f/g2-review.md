# G2 增量架构评审

基线：[password-visibility-architecture.md](password-visibility-architecture.md)、[product-definition.md](product-definition.md)。结论：Ready for Development；无阻塞，按自动推荐授权采用轻量方案 A。此状态仅为设计就绪，不是实现、make verify 或独立 QA 通过。

## 评审结果（按架构影响）

1. 整体影响：Go／原生前端、单管理员、既有认证与持久化边界保持；新增能力只在登录 DOM。角色、四项相邻能力、事实源和正式输入／登录／退出／凭据运维路径清楚，没有新控制面对象。
2. 交付影响：index.html、app.js 与两类前端测试为主要范围，样式仅局部；Trellis bootstrap／before-dev／check、make verify、独立 QA 和平台 Git 发布仍是下游责任。
3. 证据与结论：当前 password 输入、clearSecrets、signedIn／signedOut 和 401 处理足以承接局部状态。W3C 官方按钮模式支撑动态动作名称＋状态说明；不混用变化名称与 aria-pressed。
4. 缺口：新增交互、可访问树和真实浏览器行为尚未执行验证；不影响设计就绪，影响最终验收。不能用 VM click 或历史回执替代键盘和真实认证证据。
5. 决策：按用户普通选择自动推荐委托采用 A，低影响可逆，不建独立 ADR；安全手册本地审查覆盖秘密副本、意外提交、残留及证据泄漏，没有权限扩大或新合规承诺。
6. 开放问题：无需求方向级未知。遗漏重置、语义同步、焦点及证据脱敏为执行风险，由新增测试及 QA 阻断，不重复询问进入下一阶段。

## Skill 适用与裁剪记录

已完整读取 platform-architecture-cn 的 SKILL.md，以及 references/expert-routing.md、review-and-output-contracts.md；针对生命周期故障读取 architecture-lenses.md，针对密码数据保护选择并完整读取 expert-security-architecture.md（本地专业分析，不派生子智能体）。主专家边界、事实、风险、控制、证据与未知、验证／退出和不决定事项纳入方案第 3 节，并由总体评审综合为局部前端方案。

按小型增量裁剪：覆盖目标→角色／体验→能力／领域→流程／事实源→既有运行拓扑→实现契约；未引入企业运营模式、集群、AI 加速器、厂商选择、迁移或新公开契约。因此不加载条件式生态、设备、技术研究或详细 evolutionary-architecture 参考，也不产出独立 PoC、部署图、原型或 ADR。目标状态、方案比较、决策及风险职责分区；简化审查仅保留当前行为需要的节点和函数。

已恢复本 Run PRD／G1、项目 brief／roadmap／v0.1 范围、历史版本设计相关基线和实际登录／测试代码。未改业务代码、历史文档、既有准备变更或规范模板；未运行 make verify、浏览器旅程、独立 QA、部署或供应商调用。W3C 页面为本轮官方机制查阅，不是运行证据。

## 下游交接

研发使用本 Run 的 PRD 与本架构文件，不猜测历史 design.md 为本轮事实源。按第 4 节落实实现与 AC01—AC13 验证矩阵，完整执行挂载 Skill 和 Trellis；保留失败／沙箱证据，必要时使用固定 host-only Connector，不扩大权限。QA 报告和证据放本 Run；最终报告节点写本 Run 的 pr.md；旧 Issue #8/#9 与历史日志继续保留。

旧文档保护由需求阶段 historical-documents.sha256 提供。设计节点的核验结果另存 design-integrity-check.txt；哈希核验只是历史保留证据，不表示功能／健康／版本测试通过。原有 .gitignore、.gitattributes、.trellis 准备变更保留。
