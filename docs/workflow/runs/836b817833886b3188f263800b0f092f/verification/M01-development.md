# M01 研发验证与交接记录

- 验证对象：T001 登录密码可见性及相邻回归；执行人：研发节点
- 时间：2026-10-06（本机 Asia/Shanghai）；环境：macOS arm64，Go 1.25.6，Node 26.0.0，受限原生沙箱
- 基线 HEAD：73681a7df4c5cbe5914056ac13348154d58726e0；未提交工作树，代码标识见 final-integrity.log
- 结论：Blocked（完整固定门禁受监听限制），不是功能验收失败判定，也不是 Accepted／Delivered。

## 当前执行证据
| 检查 | 本轮结果 | 证据 |
| --- | --- | --- |
| make verify | 退出 2；evidence 成功，health-history 旧版 HTTP 隔离重建无法监听 127.0.0.1:0，operation not permitted，分类器拒绝该环境错误 | make-verify-attempt-1.log |
| make format-check / vet / build | 全部退出 0 | non-listening-checks.log |
| go mod verify | 退出 0，锁定模块验证成功 | non-listening-checks.log |
| make frontend-test | 27 项通过，含新增保值／空值／零请求与成功、退出、会话恢复、401 重置；既有版本竞态与管理导航回归仍通过 | non-listening-checks.log |
| Go handler 健康／版本回归 | race 下 TestHealthMethodsHandler、TestHealthVersionContract 通过；非真实 HTTP 证据 | non-listening-checks.log |
| journey.mjs 语法 | node --check 退出 0；没有运行真实浏览器 | non-listening-checks.log |
| 历史保护 | 10 份文档 SHA-256 OK；tracked 历史文档与日志 diff 为空；62 份历史日志已补充保护哈希 | final-integrity.log、historical-logs.sha256 |
| 差异与源码身份 | git diff --check 成功；最终源码哈希与 product/test/build 摘要已记录 | final-integrity.log |

固定门禁在 health-history 阶段停止，后面的完整 Go race/HTTP、browser 旅程没有执行。独立运行的非监听检查是补充反馈，不替代 make verify。本轮没有完整管理台、真实健康 HTTP、键盘或 AX 运行通过证据，没有部署、真实供应商、屏幕阅读器音频验证。

首次固定验证摘要为 2133707c5db422bbd2d94b269b0d710ae3064be4bf02554cc1f9e97560e55d02。随后补强旅程中隐藏态输入 Enter 登录及日志排除哨兵，最终摘要为 0f6ee2d6894067323697f632231ee0d62b098be0261a4da7e975edb6e1efb37b；两份身份都保留。宿主必须验证最终工作树，不把首次失败或补充检查冒充最终门禁。代码身份脚本明确仅覆盖 product/tests/build locks/README/既有 acceptance-design，不覆盖交付回执和新规范；规范变化见下节与实际文件。

## AC 状态与真实旅程契约
- AC01—02：VM 证明原节点和值、空值、多次切换；HTML 默认 password 已代码检查。真实初态／刷新仍待宿主旅程。
- AC03—05：已写入真实 Tab／Enter／Space／鼠标、焦点 outline、button/status AX snapshot、名称／描述关联和零 request/submit/invalid 断言；只通过语法检查，尚未真实执行。
- AC06—08：VM 生命周期与 401 清理通过；真实拒绝、隐藏态输入 Enter 登录、显示态按钮登录、成功清空、退出、会话刷新旅程已补齐，待宿主执行。
- AC09：实现只改 type 和状态文本，无新密码副本、网络／持久化／日志；旅程比较 URL/storageState，检查 local/sessionStorage、IndexedDB、service workers 并继续现有服务日志排除，尚待真实浏览器证据。
- AC10—11：保留原 1280px／390px 管理、SSE、取消和失败链路；版本 VM 与健康／版本 handler 当前通过；真实管理、健康 HTTP 和版本一致性仍待固定门禁。
- AC12：真实受限失败已记录，完整门禁待宿主；AC13：当前历史保护检查通过，独立 QA 仍需复核。

## Skill 与 Trellis 的实际动作
1. 完整读取 managing-engineering-delivery-cn 与 references/public-delivery-contract.md、delivery-review-checklist.md；恢复本 Run PRD、G2 与增量架构，未重建上游方案。公共目录映射到本 Run，保留 README／milestones／tasks／verification 职责；一个小切片裁剪为一里程碑、一任务，无组织运营文档、人工排期或第二套调度器。交付审查 Ready for Dispatch，验证后任务进入 In Review，不是 Done。
2. 完整读取 trellis-spec-bootstrap 与 repository-analysis.md、spec-task-planning.md、spec-writing.md；采用直接源码分析，不需 GitNexus／ABCoder 配置，因此条件式 mcp-setup 不适用。检查现有空模板、Go/module/Makefile、server/store、登录／版本／秘密清理、VM 与真实 Go 浏览器 Harness、健康／版本测试。保留单 repo frontend/backend 边界，合并空白规范为 8 份英文源码规则，删掉不适用的框架 hooks／ORM 模板。索引包含 Pre-Development Checklist 和 Quality Check。实际验证全部 8 文件无模板串，Markdown 导航均可解析。已有本地 bootstrap 任务标为 completed；任务与缓存受忽略，不作为公共交接事实源。
3. 完整读取 trellis-before-dev；实现前执行 python3 .trellis/scripts/get_context.py --mode packages，返回 single-repo、backend/frontend。读取实际 frontend/index、dom-lifecycle、verification，以及 backend/index、runtime-contracts 和三个 shared guides；任务文件写明最小行为缺口、文件边界、非目标和验证。所有实现都在预定 5 个文件中，Makefile、Go/API/存储未变。
4. 完整读取 trellis-update-spec：把新增 passwordVisibility 的可执行状态表、401 清理、原生按钮语义、正确／错误示例和所需测试同步到 frontend/dom-lifecycle.md。不改变跨层／基础设施契约，七段基础设施模板不适用。
5. 完整读取 trellis-check；实现后执行 diff/name/status、packages，再读 frontend 索引及其实际规则；检查 DOM ID、单一事实源、秘密副本、无 keydown 双触发、生命周期入口、无新增依赖或日志、作用域、测试覆盖与规范同步；执行项目检查并记录真实结果。单层 DOM 功能，无三层数据流变更，跨层专项不适用；复用检查已搜索 clearSecrets/signedIn/signedOut，统一 reset 而非重复密码状态。无静态质量违规；完整运行结果受限，不写 trellis-check 全绿。原生 JS 无配置的独立 type-check，不新增工具链。

## 历史日志副作用与恢复
固定 tests/health-history/main.go:93 本来会写 docs/workflow/health-methods-history.log。本轮 make verify 将受限失败写入该历史路径；失败全过程已保留在本 Run/make-verify-attempt-1.log，随后将旧日志精确恢复到进入本节点的 HEAD 版本，git diff 为零。没有删除失败，也没有把旧成功改为本轮结论；受限回执的本机绝对路径已脱敏。

下游宿主运行固定 make verify 时同样须先保护历史日志，保留新的 stdout／退出码到本 Run verification，并在执行后将该历史路径恢复到保护基线；若需要保存生成的历史重建日志，复制为本 Run 的新证据文件，不能覆盖旧 Issue #8/#9 日志。不修改或削弱可信门禁。historical-logs.sha256 是恢复后且已与 HEAD 比较一致的保护基线，不是历史测试重跑通过证据。

## 交接与门禁
交平台 tests 的固定 host-only Connector 实际运行 make verify，保留失败／后续全部证据；然后独立 QA 核对 AC01—AC13 和当前源码哈希。若宿主发现实现或测试问题，返回研发修复；本地没有需要需求／架构修订的证据。不要求扩大沙箱权限，不使用静态 browser.check 伪装后端验收。QA 与证据继续位于本 Run，最终 pr.md 由报告节点生成，Git／Issue 发布由平台负责。
