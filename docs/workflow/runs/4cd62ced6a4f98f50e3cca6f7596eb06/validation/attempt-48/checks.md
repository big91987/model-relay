# development48：management-summaries-r1 真实验证与宿主交接

2026-10-08；Run4cd62ced6a4f98f50e3cca6f7596eb06；研发不是独立QA。结论：**候选整改已实现、局部检查通过；当前完整make verify/实际HTTP/Chromium/独立QA仍待执行。QA47 No-Go未由研发关闭。**

## 1. 输入、版本与承接

- 主线HEAD仍为`98511771871cf0951ecef716bbb55deb13e18da5`，候选包含本Run既有未提交实现，不把HEAD当全部候选。当前133件源戳`2b1297e865a8b595af98cd7594900d4a6c0506729c0e46de5ba3a05640a2a962`，生成命令`go run ./tests/evidence`，实际Go1.25.6；见[source-evidence.json](source-evidence.json)。源戳沿用项目公共算法，包括产品/测试/锁文件/项目契约/根验收矩阵/Trellis，排除交付回执防自指。
- 材料`design-v0.2.0`，SHA256`5f51d35fdc15ed7214749f9413879b4238097f1d6b6bbe2cc0d508d8e103cef7`，实际入口`.workflow-input/4cd62ced6a4f98f50e3cca6f7596eb06/5f51d35fdc15ed7214749f9413879b4238097f1d6b6bbe2cc0d508d8e103cef7/files/manifest.json`。本轮直接复核原archive摘要及manifest19件大小/摘要，全部匹配；具体引用PRD§7/interactions页面行为、acceptance全部AC/H、decisions和A原型三文件的仓库相对路径见[material-integrity.json](material-integrity.json)。不执行材料包脚本，不改冻结状态。
- 直接承接`architecture-r1`、`implementation-contract.md`和[QA47整改包](../../acceptance/remediation-handoff-47.md)；[QA47原报告](../../acceptance/acceptance-report-47.md)与[逐页对照](../../acceptance/visual-review-47.md)原位保留。上游seq46是旧129件指纹`7e0b20ddbdea4585959365d8ad3b6121abaa91cd3f450fa2c6c21f7afc05a9ab`完整门禁exit0/104299bytes/truncated=false；本轮没有重新执行seq46或冒充读取完整日志，使用QA47保留的审核与回执，不给新候选复用Pass。
- Codex/gpt-6.1-sol为用户指定及平台配置声明；本节点工具不提供原生模型的独立运行证明，仍交平台核对，未换模型/重建Run/恢复旧代码。

## 2. 实施边界与真实事实源

行为缺口在Go只读聚合和正式workspace原表格，而不是要扩大冻结范围。新增`internal/relay/console_summaries.go`，原`identity.go`/`console_api.go`/`console_reports.go`补派生字段；`workspace.js`补摘要与原入口上下文操作；原Trellis规格写回可执行契约。没有新表、迁移、身份授予、数据库手改或计费/结账写入修改。

| QA项 | 本轮实现及事实源 | 新回归 | 当前状态 |
|---|---|---|---|
| QA47-01 模型 | 当前候选关联服务名/来源类型、普通/SSE能力、按执行版本与生效时间有效的每候选租户输入输出价；scope内实际已结束dispatch均值，没有样本明确未测；原PATCH启停及公共名上下文调用 | production catalog/SQLite已知/未知/失效价；浏览器两模型精确值、固定3000000/2000000费率、启停/调用名/不填秘密 | Implemented + Local Pass；Native/QA Pending |
| QA47-01 服务 | 实际模型路由关联数量/公共名/各模型候选优先级；核验完成时间、当前执行版本、核验版本/适用性，关联详情进入原编辑 | 浏览器实际创建1服务/2模型；与正式service/model API逐字段对照、详情实际按钮 | Driver implemented；Native/QA Pending |
| QA47-01 租户 | 有效实际管理员成员为负责人，不把平台资格或待接受邀请当负责人；实际成员/模型授权/稳定Key数含撤销历史；同limits/usageFor的周期已确认/预留/上限/硬软/待确认/平台cap | Go创建/接受真实成员、权限负向及预算对账；浏览器2租户多Key，对账members/keys/catalog/limits及范围入口；开发者摘要受限、扩目录403、非成员404 | Implemented + Local Pass；Native/QA Pending |
| QA47-01 工作台 | 当前scope目录模型启停/来源/延迟及模型/预算入口；原真实趋势/费用/风险不替换 | 原visual-activity保留；浏览器精确模型状态及入口范围 | Driver implemented；Native/QA Pending |
| QA47-02 原始/确认用量 | 原始Token/attempts/snapshots不变；请求列表/详情/CSV另附charge_entries累计输入/输出字符串；UI解释原始与确认并保留pending | Go全部多尝试6/8+pending→核对9/12且重放一次不多收；原entry/attempt/snapshot守恒、CSV/分析对账；原J06实际网页unknown/0+0/pending1→原始仍unknown/确认3+4/pending0，详情/CSV/分析 | Implemented + Local Pass；Native/QA Pending，原P2不单独阻断 |

这是AC01/02/04/13/14/15/17相关整改，不替代AC01–AC18/H01–H08全矩阵。原普通/SSE/双发/未知签发/轮换/并发预算/迁移/恢复/J01–J07门槛保持。

## 3. 实际局部命令与结果

| 命令/日志 | 实际exit | 证明范围/边界 |
|---|---|---|
| `go test -race -count=1 -run '^TestConsoleSummariesInProcess$' ./internal/relay` / [summary-first.log](summary-first.log) | 0 | 初稿生产handler/SQLite摘要回归，非最终完整源码证明；14.733s |
| `make frontend-test` / [frontend-first.log](frontend-first.log) | 2 | 首次3失败：既有J06源码片段边界把新增native段纳入VM、旧service VM夹具未供models读取/sourceKind。不是native/后端故障；原失败完整保留并去本机路径 |
| `make frontend-test` / [frontend-second.log](frontend-second.log) | 0 | 修夹具读取依赖及源码测试片段边界，保留原账本/核验revision/失败断言；184/184，未把VM当浏览器 |
| format/module/vet/build/JS/diff链 / [static-first.log](static-first.log) | 0 | 中间静态检查，不跨版本充当完整门禁 |
| InProcess等定向race / [go-regression.log](go-regression.log) | 0 | 中间handler回归73.076s，之后继续补scope负向故仍保留为历史 |
| `make format-check && go mod verify && make vet && make frontend-test && make build`，再JS语法/`git diff --check` / [static-final.log](static-final.log) | 0 | 最终格式、锁模块、vet、184Node、Go产品构建、三处JS语法/diff检查；无新增依赖、无lint抑制 |
| `go test -race -count=1 -v -run '(InProcess$\|^TestConsoleFreshInstallationEmptyCollections$\|^TestConsoleIdentityAndKeys$\|^TestConsoleServiceCheckRevisionContract$\|^TestConsoleKeysEnvelopeContract$)' ./internal/relay` / [go-final.log](go-final.log) | 0 | 实际11个顶层测试通过、75.013s；包括Governance/Resources/PeopleScope/Pause/Accounting/五维与共享资源/新Summaries/新安装/核验/Identity。此正则中的最后一个名字未选到测试，不将其另算覆盖。No real listening |
| archive/19件material SHA/size复核 / [material-integrity.json](material-integrity.json) | 0 | 同原冻结输入；没有下载其他Run/旧机器数据 |
| `go run ./tests/evidence` / [source-evidence.json](source-evidence.json) | 0 | 当前版本指纹与实际工具Go版本，非运行health/模型/部署证明 |

Go链接器原`malformed LC_DYSYMTAB`警告出现在这些本机Go测试记录，命令仍exit0；未屏蔽或改成完整门禁Pass。白名单去路径记录见[log-redaction.json](log-redaction.json)，保留日志错误类别/测试定位/时序/结果和去标识前后摘要，未转存凭据、DB或绝对本机路径。

## 4. trellis-check 与公共交付审查

实际Trellis`.trellis/.version`=0.6.15；`get_context.py --mode packages`成功确认single-repo backend/frontend。已有源码规范并非prepare待填模板，无须bootstrap。恢复相关索引、runtime/console/routing/multitenant及DOM/lifecycle/verification/workspace指南、code-reuse/cross-layer与原任务/架构；按before-dev确定只读聚合/UI/回归边界，按update-spec把新增API字段/权限/空值/原始与确认口径及错误/好坏案例/断言写回原规格。

本轮trellis-check：格式/vet/build/Node/定向Go race已实跑；参数绑定、先scope选请求再聚合、权限没有只靠隐藏按钮、developer延迟遵守诊断分配、epoch拥有公共名转移、无新增秘密日志、无循环依赖/跨层类型不匹配。原账本写方法/快照/迁移/Makefile未变。公共交付审查保持T001/M01，不重建阶段/复制事实源，不新建PRD或额外审批。不能把未执行网络/浏览器/独立QA勾为已通过，最终退出仍**Blocked by Validation**（可继续tests，不是外部业务阻塞）。

## 5. 宿主固定入口必须执行的复验

1. 对同133件指纹候选执行**原完整`make verify`**，900s既有约定，不抽取更弱子集或删health-history/182原Node/原HTTP-race/迁移/restore/fault/J01–J07/visual-activity/63图断言。首个失败保留原命令、exit、完整分段日志和当前指纹，按scope/产品/驱动/时序事实再返研，不将产品失败归入沙箱。
2. 新`TestConsoleSummariesActualHTTP`必须实跑；当前未执行。`tests/browser/journeys.mjs`原路径在1280/1440/390均必须得到新`suite=management-summary`真实回执，预期产物`test-results/management-summary-{width}.json`。这些文件尚未产生，不能作为本轮已有artifact。
3. 原J02开发者范围负向与新restricted摘要；原J06同ID真实网页核对前unknown/确认0+0/pending1及核对后原始unknown/确认3+4/pending0；实际详情、同ID分析及CSV字段对账；原金额/快照/终态/幂等断言继续生效。
4. 原visualReview仍生成十页/三视口/63PNG及review，新增必需摘要标题/字段/上下文按钮断言，不只截图。新的management-summary断言再核两模型、两租户、多Key实际值和对scope/API/受控创建事实的三方对照。保留历史75件QA47产物，不用新采集覆盖原失败。
5. tests完整通过后交独立QA，同冻结manifest/version/SHA、新指纹/回执/实际页面再次决定QA47-01/AC01与QA47-02。研发此时只提交修复候选，原QA结论没有被改写。

## 6. 未验证/不能外推

原生监听稳定限制见[首次make verify失败](../attempt-01/make-verify.log)和先前报告：`listen tcp 127.0.0.1:0: bind: operation not permitted`，外层make exit2；执行环境未改变，本轮不反复运行同一必败门禁、不提权、不改用静态browser工具放行。尚未取得本版本真实HTTP/SSE/进程恢复/Chromium/视口/键盘与整套门禁结果，必须宿主执行。

真实供应商、外部支付/身份、平台实际模型及负向Harness、PR/merge/正式Actions部署/healthz效果一致仍Not Run；受控真实HTTP只证明本项目语义，不证明供应商联调。Git分支/提交/推送/PR/合并/正式部署由Connector/协调者负责。本节点没有制造演示服务/重启Pipeline/手改数据库/读取New API本机源码、数据或秘密。

## 7. 交接产物

以[artifact-index.json](artifact-index.json)列出的实际仓库文件及摘要为准；当前公共任务、根验收追踪和验证记录已经更新。继续同Run的tests，失败回development；没有需要requirements/design修订或用户再次批准的事项。
