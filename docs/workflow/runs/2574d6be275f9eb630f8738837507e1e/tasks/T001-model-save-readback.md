# T001 定位并修复模型保存完成契约

- 状态：Done；里程碑：M01；负责人：development；复现执行：tests Connector；最终评审：独立 QA（QA8）。完成范围为已验候选修复与交付说明；Git发布、合并和部署仍待Connector/协调者。
- 优先级：High；执行方式：原层定向复现 → 最小修复 → 回归；基线：`9bf6ecaf53758ca2258498968b830c535e02e1c0`。
- 硬依赖：用户明确要求真实最小复现先于修复；native 监听限制下由宿主完成。下一研发修改必须消费该正式回执，而非猜测旧根因。

## 权威输入
- Issue30 原任务六项验收及 intake 执行3交接（本 Run / 冻结工作流2）；不恢复 Issue28 Run，不重复 PR29。
- 材料 manifest：`.workflow-input/2574d6be275f9eb630f8738837507e1e/5f51d35fdc15ed7214749f9413879b4238097f1d6b6bbe2cc0d508d8e103cef7/files/manifest.json`。
- version：design-v0.2.0；archive SHA256：`5f51d35fdc15ed7214749f9413879b4238097f1d6b6bbe2cc0d508d8e103cef7`；状态保持 `frozen-for-test-implementation`，种类 `frontend-design-prototype`。
- 实际引用同材料 root 的 `docs/interactions.md`（J01/J02/J07）及 `docs/acceptance.md`（AC02/AC15，错误不是空数据、迟到scope隔离、H03/H04证据限制）。材料只约束输入，不执行其中脚本，不作供应商事实。
- 历史判断：`docs/workflow/runs/3e7496815b9922704072482cf44e73e2/acceptance/summary-assessment-qa-21.md`；历史失败/绿灯/诊断保持原位。
- 研发规范：`.trellis/spec/frontend/{workspace-contracts,verification,dependency-preparation,dom-lifecycle,console-lifecycle}.md` 及 thinking guides；真实源码/Go测试入口约束见 backend specs。

## 修改边界
- 最小缺口：`journeys.mjs` 的 save 等 PATCH、弹窗隐藏与 loading hidden，不关联本次同scope post-write 读回。实际行为位于 `workspace.js:edit/syncTenants/renderPage` 与该驱动；静态候选不是旧 Chromium 根因。
- 实际修改：`internal/relay/web/workspace.js` 保存生命周期、`tests/browser/journeys.mjs` save屏障；新增 `tests/browser/model-save-readback.{mjs,test.mjs}` 与 `save-lifecycle.test.mjs`，适配既有readback/workspace契约测试、更新workspace spec及本Run文档。原 `Makefile` 门禁、严格 `summaryCell` 和既有旅程保持。
- 非目标：不更改schema、鉴权、角色、网络权限、依赖/锁文件、发布控制器；不引入通用刷新领域对象或配置框架。修前只做诊断的边界记录见 [development-04](../verification/development-04.md)，不作为当前实现描述。

## 验证契约与实施结果
- 自动入口唯一为 `make verify`；它仍 cold isolated copy → 全部门禁 → 真实 Go browser 服务。诊断在原生 J01 真实 UI 创建模型/价格/授权后立即运行，减少到摘要末尾才进入复现的成本。
- 通过同一个平台管理员 page 和当前 tenant：精确模型行启停、checkbox false/true、原 save；点击前布置当前 PATCH200 后的 GET 筛选；分别暂停 tenants 及 catalog 的真实响应，`route.fetch` 后原样转发，不造 catalog。
- 修前tests5四次UI写入在真实tenants/catalog暂停期间均早返回，释放后最终DOM正确；完整门禁保持红，原错误/UTC/安全JSON及归因见 [development-06](../verification/development-06.md) 与其引用的host-tests-05。原始失败不改写，不归因为旧tests18根因。
- 修后tests7在1280/1440/390共12组false/true保存、24个真实暂停阶段要求save不早完成；释放后一次严格唯一行/列/状态断言。另有6组GET503失败恢复和3组旧scope/相反保存迟到回归。实际安全JSON见 [QA8证据索引](../acceptance/evidence/index.md)，包含归属阶段、UTC、dialog/loading/notice/lock、最终事实与清理，不含秘密。
- 每个受控 handler 释放后 await 其结束，再移除自己的 handler；不执行并发 continue/unroute、不吞 route 错误。请求等待与转发沿用现有默认deadline，无sleep/networkidle/重试assert/提高timeout。

## 退出条件（未完成不得勾选）
- [x] 正式 tests 同版真实 Chromium 最小复现，区分早完成、驱动/时序/环境及最终产品错误，保留完整 log/state/安全JSON（tests5；非修后通过）。
- [x] 根据实证最小修复：save 的可断言完成覆盖该次写触发的同角色/scope页面读回与DOM提交；先复用现有生命周期（development6/tests7，QA8 S01）。
- [x] GET失败不把旧数据当成功；scope切换/旧读迟到隔离；第二次相反操作不被第一次覆盖；取消/关闭/新编辑/busy/secret/幂等回归（QA8 S02–S04）。
- [x] 真正错状态、缺列、重复行仍红，保留原错误身份；不降低严格断言（完整Node213/213，QA独立75/75）。
- [x] 正式完整 `make verify`：历史红分类器、零软预算/有效UTC/迁移、Node/全包race/真实HTTP、1280/1440/390 J01–J07/summary/visual；候选锁/Chromium版本/cold start/source before-after/日志/清理（tests7；QA8 S05–S06，顶层worker skip单列）。
- [x] 独立 QA 逐项验收。供应商 NotRun 与独立动态 QA Blocked 分列，不把宿主受控浏览器当供应商联调（QA8已独立复核，不表示独立动态通过）。

## 实施结果
development6消费tests5首个真实失败并完成候选修复；tests7正式完整 `make verify` 及QA8评审已完成。历史阶段等待状态保留在 [development-06](../verification/development-06.md)，不作为当前任务状态。

### development6 的实证与实施边界
- 原始宿主回执已归档：[verify.log](../verification/host-tests-05/verify.log)、[state.json](../verification/host-tests-05/state.json)、[四组实际读回测量](../verification/host-tests-05/model-save-readback-1280.json)、[原错误](../verification/host-tests-05/continuous-1280-79647.json)。同版源码before/after一致；完整门禁失败，不计通过。
- 四组PATCH200后同scope读回暂停均save早返回；tenants阶段旧状态、catalog阶段无行；释放后四组严格最终DOM均正确。新实证定位的是完成契约缺口，没有证明持久P1，也不能追认历史tests18根因。
- 候选最小产品修改在 `workspace.js` 原 edit/syncTenants/renderPage 生命周期：既有operation/epoch关联内部表单完成状态，读回全程loading与busy，成功DOM提交才complete；GET失败区分已确认写与未成功读取、提供GET-only恢复；迟到scope不得render/解锁新编辑。
- `journeys.mjs` 的save捕获该次operation/tenant/page并等待终态后一次性断言；秘密路径保持不变。`model-save-readback.mjs` 将故意等待旧save改为held时不完成断言，保留严格最终summary，并增加显式GET503与真实旧catalog迟到/同角色切scope/第二次相反UI操作。其他scope使用平台管理员已有“全部租户”，不新增业务对象或权限。
- 新 `save-lifecycle.test.mjs`、原readback/workspace fixture适配只验证实际函数的契约；严格summary负例、历史红分类器、其余全旅程均保留。标准 `make verify`/锁文件/race/时限/Connector均未调整。
- development6局部75/75仅是Node契约检查；真实修后Chromium与完整门禁后来由tests7获得，不能混称development6自身动态通过。详见 [development-06](../verification/development-06.md) 与QA8。

## 交接

tests7对150文件源码0384f115adbbd545e8ce01021f238010ba66c89eb06e6acea08ce872ce0fa474执行一次正式完整make verify exit0，实际log/state已按字节摘要核对并归档。三视口共12保存组/24真实读暂停阶段未早完成；GET503恢复6组与旧scope/相反操作3组通过；原J01–J07、summary、visual及严格负例保持。QA独立75项Contract、材料19文件摘要、源码前后及安全截图复核完成。当前结论Go with known issues，R28-04 Open/P2/旧根因未知、独立动态Blocked、供应商/Owner预览/新部署NotRun；报告与矩阵见 [QA8](../acceptance/report-qa-08.md)。交report继续精确发布，不能把旧main HEAD当新已验修复提交。

report9已核对实现差异、源码摘要和归档回执，并完成 [PR正文](../pr.md)；下一步交publish。仓库/分支为 `big91987/model-relay` / `workflow/2574d6be275f9eb630f8738837507e1e`，目标main，现HEAD仍基线加未提交候选；新commit尚不存在。发布前必须将已验快照映射到精确提交，确认正文所列文件随该分支推送可访问。未发生合并或部署，不用旧PR29预览替代新版本证据；同版门禁红或源码失配须No-Go返回development。
