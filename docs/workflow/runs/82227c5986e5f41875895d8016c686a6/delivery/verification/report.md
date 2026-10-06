# 研发验证与宿主交接记录

## 当前产品验证结论（2026-10-06）

结论：Passed（版本11产品切片）；宿主 seq21 原 `make verify` 完整exit0，同版本[独立QA report-02](../../acceptance/report-02.md) Go，四项P1 Closed。当前[矩阵](../../acceptance/matrix-02.md)、[回执](../../acceptance/evidence/host-seq21.md)及[版本锁](../../acceptance/evidence/version-lock-02.json)是完成依据；不是本节点或QA完整入口重跑。D1–D5 Done、M01–M03 Accepted，只允许草稿PR，不代表正式发布或供应商Go。文档收尾及版本差异见[审查](report-closeout.md)。DEP02–04/06–07和供应商继续Not Run，DEP05仅标准模板子范围通过。

## 历史验证记录（以下均为当时快照，不是当前状态）

### seq19失败后版本11待验记录
最新实际门禁是10/seq19/900秒/exit2，不是当前版本通过；完整平台日志83109 bytes、truncated=false，末段已读至eof。见[失败归因、边界与复验步骤](host-failure-05.md)和[原始选择轨迹归档](host-seq19-selection.jsonl)。修复driver无合同正文等待，强化真实overview证据→精确GET→同ID正文/输入，保留原生折叠/焦点及所有路由断言；只扩展既有VM回归，未改业务/QA负控/可信门禁或平台。

11局部命令结果见[回看回归日志](review-identity-validation-01.log)；最终待验源码见[11指纹](final-source-11.json)。真实浏览器/完整make verify仍Pending Host，原沙箱listen/Mach限制未变不重复，不能以VM预测宿主绿灯；独立QA后续裁定。原[QA整改](qa-remediation-01.md)、No-Go及历次失败保留，DEP02～04/06/07和供应商仍Not Run。以下条目均为历史时间点。

本轮9条真实局部检查均exit0：`make frontend-test` 119/119（扩展completed-row原节点/同ID详情回归）；两份driver/VM语法、format、module verify、vet、全包零测试编译、build、diff。零测试编译不计作Go行为复验，VM不证明native details或真实网络。11指纹85文件、`852f7bfc90323b6253658c8b9c2167fa22c43a5f59389f0bbcb92539865120e7`；Trellis已恢复实际frontend规范、共享reuse指南并按check复查本轮单层driver/测试/安全诊断边界，无业务跨层写改动。公共计划按交付审查保持Ready for Review，退出条件没有降低；当前关键路径是新11宿主完整门禁→独立QA，不需用户重新批准普通修复。

审计首次失败记录 `review-identity-audit-01.log`：该独立命令遗漏项目GOCACHE，默认用户缓存被沙箱拒绝，未产新摘要；原回执保留、公共摘要脱敏。按既有 `.cache/go-build` 重跑 `review-identity-audit-02.log` 真实exit0：6文档/105相对链接无问题，42行原生选择轨迹严格白名单及字节摘要匹配；独立QA测试/截图/轨迹、Workflow/manifest均与QA原锁一致，原Makefile/health-history/历史Run未改，最终11指纹重算一致。此失败不是产品行为失败，也不替代或解除宿主seq19真实exit2。

### seq17整改后版本10待验记录

seq16原08的完整宿主固定入口exit0已确认；seq17独立QA首次验收No-Go，详见[原报告](../../acceptance/report-01.md)和[整改包](../../acceptance/remediation-01.md)。本节点不改原QA测试、报告、矩阵或失败证据。QA-01～04已按原契约修实现并加入强制边界回归，当前事实及宿主验法集中在[本轮整改记录](qa-remediation-01.md)。当前不是08或QA新增负控前的源码，需核对[10指纹](final-source-10.json)，重新跑900秒原 `make verify` 后独立QA判定，不能沿用seq16放行。

原本机监听/Mach限制未变；本轮不重复必然受限的完整入口，不扩大权限，保留首次[固定入口限制](make-verify-01.log)和历次宿主失败。局部真实执行仅支持其各自范围；新增真实HTTP的10秒头/body、工具EOF/idle/cancel以及新增双屏原生键盘/折叠/核验旅程尚未在本节点运行。DEP-02～04/06/07与供应商继续Not Run，平台/发布责任不变。以下均是保留的历史记录，不重写为当前绿灯。

## 2026-10-06 DEP-07标准模板同步

同Issue/Run的新增用户授权只要求受管理模板同步，不重放需求/设计、不解除seq14真实键盘失败。依据原部署契约的标准install_workflow.py路径，本节点实际执行 `python3 <platform-source>/examples/github/model-relay-preview/install_workflow.py --project . --upgrade` 两次；日志中源目录仅脱敏为占位符，实际命令使用用户指定的维护源。未手写Workflow/manifest补丁或绕过manifest漂移检查。

| 证据层 | 固定身份与实际结果 | 验证边界 |
|---|---|---|
| 本节点读取维护源 | 固定 `294ddb33a18a02d413af9e790e4d5caaf26eca51`；实际平台HEAD `e999c011325a8d1591dfc91640e7fde9382524fb`，示例目录工作树干净且与固定版本无差异，模板/安装器逐字节与固定Git对象相同 | 平台HEAD并非冒充294ddb3；本次只使用完全相同的固定示例字节，未修改平台仓 |
| 标准安装与复跑 | [首次原回执](platform-template-install-01.log)：Workflow installed，exit0；[重复原回执](platform-template-repeat-01.log)：Workflow unchanged，exit0 | installer原manifest校验通过；仅写 `.github/workflows/deploy-local.yml` 和 `deploy/platform-preview-source.json`，无控制器复制、数据修改或部署 |
| 安装身份核对 | [真实来源及摘要](platform-template-provenance-01.json)：YAML SHA256 `a3d814fa47da0c74e6934ab46138f2b854c55cf1de177f7fba538b4365ab2f45`；标准安装器SHA256 `1d09d8bd6069781fb2b33cfd9c998a35577906077dfe851d8242c81ba3b85da8`；目标YAML精确等于固定模板，manifest摘要匹配目标 | 原目标摘要为 `f77b9e76caa018b4c53c8ad38d8b4e03ee232ea2fe0ea1ac4ae50f8e25779002`；本次没有目标文件漂移或读取失败，不做人工修复 |
| 用户提供的平台进展 | 源Draft PR #5版本294ddb3，45项维护源回归及独立Important复审修复；预览根原install.py升级/重复核对，控制器摘要 `e73b92b46ea62a7520d6e93c2bf306e8cdc5dee0322fff4d83c406e51de93fba`，配置/deployed/current/key不变，实际health仍精确旧main | 这些由用户提供，本节点未重跑45项测试/预览根安装、读取控制器或访问实际health；不是本节点工具回执，更不是Go/正式Actions恢复通过 |

实际安装模板可读出：workflow_dispatch target为preview/validation，默认preview；push仍只prepare，main/Owner及既有控制器prepare/activate路径保留；validation传入protected preview root且使用独立local-validation Environment。独立根/端口/服务/敏感文件隔离判定由平台控制器负责，本节点未运行activate或证明真实隔离恢复。DEP-02～04/06/07仍 **Not Run**，维护者继续负责隔离旧安装及真实发布验收；源PR不自动合并。下方较早平台补充表是当时历史记录，最新源/安装事实以本节分层为准。

产品和测试仍为[08](final-source-08.json)，82文件摘要 `c340664b231d83d153f0d2569065db08c8be34661a314f51e7831f74f808b27e`，type-ahead driver及全部业务断言本轮未改。tests/evidence现有范围不包括受管理Workflow/manifest，故本次以08产品身份＋上述模板来源JSON联合绑定，而不声称08单独覆盖全部发布文件。[模板同步后审计](platform-template-audit-01.log)核对源码复算、模板/manifest与固定对象、旧门禁不变及文档引用。此前08审计的Workflow未变描述仅适用于新增授权前；本次Workflow变更恰为用户明确要求的标准安装器输出。

Trellis-check：已恢复backend index/runtime部署责任与既有标准模板约定；无需改产品规范、架构或重建计划。实际installer两次结果和逐字节来源校验均通过，git diff --check通过；没有增加新依赖或要求宿主外广泛提权。完整固定入口仍待宿主900秒执行，本机已知监听/Mach限制未变，不重复同一受限路径；安装Unchanged不等于make verify Passed。next保留新完整日志/exit/联合版本与候选JSONL，再独立QA。真实供应商仍 **Not Run**。

## 历史待验状态（2026-10-06，seq14→08）
当前[08源码身份](final-source-08.json)：82文件，`c340664b231d83d153f0d2569065db08c8be34661a314f51e7831f74f808b27e`；[审计](candidate-keyboard-audit-02.log)exit0，文档相对链接问题0、真实源码复算一致、原门禁/上一Run未改，归档JSONL与宿主原文件逐字节一致。交付回执不参与源码指纹。

06 make verify在900秒真实exit2，候选primary selection1280仍失败。平台完整存档77652 bytes/truncated=false，读取72500至eof；原[完整JSONL](host-seq14-selection.jsonl)10事件/5边界未截断。键到达聚焦原select但index始终0，无input/change/mutation；失败先于模型POST。依据[归因与修订](host-failure-04.md)，08仅修driver为原生type-ahead→Tab及受控名唯一ASCII前缀，不改产品/Go/API/权限/门禁/平台；不是再猜产品重绘或加时序等待。

[VM检查](candidate-keyboard-vm-02.log)116/116 exit0；[局部检查](candidate-keyboard-validation-02.log)语法、format、module、vet、全包零测试编译、build、diff exit0。VM只证明调用顺序/唯一前缀拒绝/不程序设值/脱敏，不证明原生选择成功。现有精确label、真实Go UI创建及所有业务断言保留，完整08门禁仍Pending Host；下文旧待验身份为历史记录，不作当前结论。已知本机监听/Mach限制未变，不反复执行同一受限路径。独立QA、正式Actions非空升级/可信隔离失败恢复与真实供应商未通过。

## 最新返工状态（2026-10-06，seq12→06）
05固定make verify在900秒真实exit2；原双屏及旧UI非空迁移局部PASS，多候选primary selection1280 FAIL。精确locator已命中，目标index1、Tab后index0；旧快照缺乏逐键留存，无法区分native移动/提交、事件拦截或产品重绘。[最新诊断](host-failure-03.md)只补有界旁观轨迹与两项脱敏/cleanup回归，保留Home/ArrowDown/Tab及原最终断言，不改产品或假设性换按键。114项VM PASS不证明真实机制。局部命令见[candidate-trace-validation-01.log](candidate-trace-validation-01.log)，当前版本见[06身份](final-source-06.json)；完整固定入口仍Failed/Pending Host，不升级Done/Accepted。

本机监听/Mach限制未变，不重复已证受限完整命令；交900秒宿主同入口。失败回执只留紧凑白名单，完整有界轨迹在ignored `test-results/routing-selection-<execution>.jsonl`，宿主须按输出PID留档实际文件并提供相对路径，不依赖截断回执。所有旧失败保留；正式Actions非空升级/可信隔离失败恢复与供应商仍Not Run。

日期：2026-10-06。节点：研发实现；评审人：后续独立 QA。范围：Issue #18 / 公共路线2。结论：**Ready for Review；完整固定门禁 Pending Host；未形成 Accepted/Done 或本轮发布许可**。

最新返工：seq9在300秒预算中断；seq10同04版本900秒完整运行exit2。实际CLI/Go测试、旧UI非空迁移与回退、原控制台双屏通过；多候选1280创建阶段失败。详见[宿主04与标签修订](host-failure-02.md)。当前待验版本为[05](final-source-05.json)，仅候选独立label关联、VM回归、原driver白名单诊断和规范变更；历史下文04“待验”保留为当时记录，不作当前通过结论。正式Actions/隔离恢复/供应商仍Not Run。

本次trellis-check：frontend index所指DOM/lifecycle/verification与共享reuse/cross-layer已核对；只改DOM标签关联，无API/存储/权限改动。模式+行号ID在创建/编辑各两行唯一，标签不再包含option文本，原生select及候选值保持；新增VM回归112/112 exit0。[局部检查](candidate-label-validation-01.log)语法/格式/module/vet/全包零测试编译/build/diff均exit0。驱动诊断只读且有界，不打印错误message、option文本、input值、URL或响应正文。固定门禁不改，已知本机监听及新Chromium Mach限制明确保留，05完整门禁不能勾选Passed。公共交付审查维持既有D4/D5、M01–03 In Review，不创建重复任务体系、不以局部绿色或仪器化绿色声称根因完全消除。

05实测指纹：80文件，`f251735964e9fcc330e4f70387e3a3b073fd63537f3960e6cbd8a4daca560510`。[最终审计](candidate-label-audit-01.log)exit0，6份文档/56个相对链接问题0，实际源码复算一致，Makefile/Workflow/历史health-history与上一Run未改。交付回执不参与指纹避免自引用；本次未重跑已确认受限的完整固定入口，next由宿主900秒按同一make verify复验。

权威输入：[PRD](../../product-definition.md)、[架构](../../routing-architecture.md)、[交互](../../routing-interaction.md)、[迁移ADR](../../adr-001-storage-upgrade.md)、[发布契约](../../deployment-upgrade-contract.md)、[原验证矩阵](../../routing-validation.md)。公共状态：[总览](../README.md)及D1–D5任务文件；不复制新的需求事实源。

## 1. 代码与门禁身份
- HEAD/精确旧安装基线：`85f6d17c07c5fddeb2d6a39a5a067769a11bb4cc`。研发未提交/建分支/推送。当前工作树包括入口/需求/设计已有产物及本节点实现。
- 首次固定门禁 [make-verify-01.log](make-verify-01.log) 的源码摘要是 `35c1ea27954b90df8e640e2bb59d61a2ca8835a1625d9b485a51b9fef1eaa4b0`，不能当最终工作树的通过回执。
- 03版本 [final-source-03.json](final-source-03.json) 的摘要为 `39c1a53b0851952ced57519c97a5e5a146c3c0a53f49158df24a676a035287b9`；该版本宿主固定入口已实测 exit 2，见 [失败摘要](host-failure-01.md)。03相对[02](final-source-02.json)仅同步平台文档证据；02相对01补原生候选键盘流程。历次快照及失败不覆盖。
- 当前修复待验版本为 [final-source-04.json](final-source-04.json)，真实 `go run ./tests/evidence` 输出，80文件，摘要 `5b30123a1570e5cef7c1cf18ddf6243753d6e5518d2962523984f2cdef4743e0`。相对03仅改 `cmd/model-relay/main_test.go` 和 `.trellis/spec/backend/routing-contracts.md` 的CLI契约回归，产品代码/浏览器/Makefile未变。覆盖范围同此前，排除交付回执避免自引用。宿主须核对04并重新执行固定入口，不用03失败或04局部绿灯作为完整通过；其后源码或规范修改必须重新生成身份。
- 固定 `make verify` 顺序、Makefile及health-history源保持不变；没有新绕过入口。扩展原测试程序纳入多上游/迁移旅程。健康GET因实际schema新增字段兼容历史重建，当前version_test仍严格断言schema2，HEAD/其他方法/拒绝门禁保留。

## 2. 真实执行记录与失败处理

全部命令在项目根执行；Go独立命令使用 `GOCACHE="$PWD/.cache/go-build"`，Makefile本已导出该缓存。局部命令不是项目验收替代入口。

| 命令 / 证据 | 真实结果 | 解释与边界 |
|---|---|---|
| `make verify`，[首次日志](make-verify-01.log) | exit 2 | health-history真实重建因 `listen tcp 127.0.0.1:0: bind: operation not permitted` 非零；不算预期红灯通过，后续固定链未执行 |
| `node --test tests/browser/ui.test.mjs`，[frontend-09.log](frontend-09.log) | exit 0，111/111 | 原async/会话/版本/工具诊断回归及新增创建busy/读回失败/账户失效晚到；VM不是实际DOM/浏览器/网络 |
| `go test -race -count=1 -v ./internal/relay -run 'Test(ManagementResponse\|RoutingDeadline\|RoutingPrecommit\|AttemptOutput)'`，[boundaries-01.log](boundaries-01.log) | exit 0 | 慢管理写前锁释放、整体deadline/取消零备用、提交前超时切换、首提交停止timer；缩短timer是边界单测，不证明默认秒数真实等待 |
| `go test -race -count=1 -v ./internal/relay -run 'Test(ConnectionCheck\|CandidateWrite)'`，[configuration-01.log](configuration-01.log) | exit 0 | 列表GET非生成、核验中改凭据可提交且旧结果不适用、重复候选原子拒绝；模拟transport证据 |
| 广非监听选择第03轮，[non-listening-03.log](non-listening-03.log) | exit 1 | 误把旧TestStorageFailClosed纳入，该真实HTTP测试仍在监听处受限；不是存储产品断言失败。保存失败后不重跑该监听用例，仅修正非监听选择 |
| 真正非监听集合第04轮，[non-listening-04.log](non-listening-04.log) | exit 0，race | 下面完整选择表达式，含升级/路由/配置/记录/安全/版本；不替代HTTP/SSE实际传输 |
| CLI/历史分类第01轮，[cli-history-01.log](cli-history-01.log) | exit 1 | 新deployment-contract及历史分类单测通过；旧CLI安装全旅程在端口占用验证处受限，未跑服务后半段 |
| `go test -race -count=1 -v ./cmd/model-relay ./tests/health-history -run 'Test(DeploymentContractAndIdempotentUpgradeCLI\|HistoryCommandOutput\|HistoricalAssessment)'`，[cli-history-02.log](cli-history-02.log) | exit 0 | 实际CLI构建、只读契约、schema2 init/重复upgrade/JSON、历史回执输出/精确分类；历史夹具不是真实网络重建通过 |
| `go test -run '^$' ./...`，[compile-01.log](compile-01.log) | exit 0 | 最后新增真实HTTP/迁移driver可编译，零测试运行，不记功能Passed |
| 格式/module/vet/build/所有driver语法，[static-build-03.log](static-build-03.log) | exit 0 | 最终源码：`make format-check; go mod verify; make vet; make build`，及app.js与全部浏览器mjs的`node --check` |
| 最终工作树/driver语法/摘要/文档审计，[final-audit.log](final-audit.log) | exit 0 | 新native候选driver语法、最终摘要一致、相对链接/本机路径/历史门禁无修改；不是浏览器执行 |
| 平台用户补充后的文档/摘要审计，[platform-update-audit.log](platform-update-audit.log) | exit 0 | 当前03摘要/相对链接/固定平台SHA/安装摘要/Not Run边界；仅文档接缝同步，不复跑平台测试或产品完整门禁 |
| 宿主03 `make verify`，[已收到失败摘要](host-failure-01.md) | exit 2 | 当前relay/历史分类包通过但Go总体FAIL；中段截断，不能推断唯一故障。build/browser未通过，不归为本机沙箱限制 |
| 本机CLI包级最小诊断，[cli-repair-before-01.log](cli-repair-before-01.log) | exit 1 | 03源码；deployment-contract通过，安装旅程在端口占用检查受限，未执行后半段。保留并停止重跑完整监听路径 |
| 新免监听管理契约回归，[cli-management-01.log](cli-management-01.log) | exit 0，race | 实际handler确认旧管理410/旧模型400、新分页/候选写入成功；不是进程、网络或浏览器证据 |
| 最终04局部检查，[cli-repair-validation-01.log](cli-repair-validation-01.log) | 各exit 0 | 格式/module/vet、管理/部署CLI契约及历史分类race、全包编译、构建、diff检查；完整make verify仍待宿主 |

第04轮完整命令：

```sh
GOCACHE="$PWD/.cache/go-build" go test -race -count=1 -v ./internal/relay -run 'Test(ExplicitUpgrade|Upgrade|RoutingSelection|RoutingTemporary|RoutingBounded|RoutingPassive|ManagementResponse|RoutingDeadline|RoutingPrecommit|AttemptOutput|ConnectionCheck|CandidateWrite|InProcess|ConsoleOverview|ConsoleEvidence|ConsoleDenied|DestinationAndProtocol|HealthMethodsHandler|HealthVersionContract|InterruptedAndVersion|SourceAndLoginLimits|SpecialInstallationPath|HeaderBoundary)'
```

所有失败保留：frontend-01至04为接口/夹具适配过程（01/02包含挂起/取消，不能从不完整日志反推退出0）；05/06/07早期局部绿灯不替代最终09。frontend-08退出1是新测试未传submitter、VM的form按钮查询为空，与产品正常native submitter不同；补submitter后09通过。static-build-01退出2是version_test.go未格式化，gofmt后02/03通过。non-listening-01/02、routing-upgrade-01和vet-01是早期检查，保留其原范围。macOS race链接器LC_DYSYMTAB warning保留，未导致绿灯命令失败。

同一沙箱、执行环境未改变，不再次执行必然在监听处退出的`make verify`，没有请求扩权。误选监听用例的两份后续局部失败也保留，不能将整份日志写Passed。最终代码与首次失败通过两个真实摘要关联，完整门禁仍必须由宿主运行。

## 3. 实现追踪与待验矩阵

以下“局部已检查”只表示开发检查，不是AC验收。每项最终状态均 **Pending Host/QA**；真实供应商另为 **Not Run**。

| AC / 任务 | 实现及已检查层 | 宿主固定入口需核对 |
|---|---|---|
| 01 / D2,D4 | configuration.go安全多对象/revision；UI列表/新建/编辑/取消/凭据清理 | 两个真实页面配置、独立启停与重启、名称/地址错误恢复及另一对象不污染 |
| 02 / D2,D4 | 显式GET模型列表、核验记录/代次/适用范围；no-generation/迟到单测 | 成功/401/429/非法列表/超时/取消/双操作的实际HTTP、费用文案、GET/POST独立计数、不清生成冷却 |
| 03 / D2,D4 | 原子候选/revision/重复拒绝；原生fieldset、取消草稿 | UI不同远端映射、所有priority/weight边界与无效引用、普通/工具参数透明、重启回读 |
| 04 / D3,D5 | SWRR/优先级单测与真实HTTP并发回归实现；新浏览器40次3:1、20次1:1、10次优先级10:0 | 实际接收计数、三层停用、快照与提交停用边界；driver原生键盘激活动作（原journey保留按键填写/选择） |
| 05 / D3,D5 | 临时白名单、最多3不同上游、总预算；单测通过；真实HTTP逐状态/4候选上限新回归已编译 | 网络/TLS/传输/HTTP/SSE切换与不可切换、真实次数/无候选/耗尽、缩短时限与默认预算区分 |
| 06 / D3,D4 | runtime30s候选冷却、顺序观察与配置隔离；单测被动恢复 | 实等31秒双屏恢复、无后台探活、全部冷却/重复失败/启停保留/重启未知/跨模型隔离 |
| 07 / D3,D4 | precommit timer、整体deadline、取消及首输出不重放边界单测 | 实际SSE/工具首提交后的断流/idle/写失败；取消/跳页/退出/过期时备用计数0与当前上游取消 |
| 08 / D3,D4 | 逐尝试/选择/安全分类/已知部分及未知合计；原console安全边界 | 同ID对账、溢出/零/未知/全已知合计、真实存储Start/attempt/Finish失败、中断与日志无秘密 |
| 09 / D4,D5 | 原journey保留两屏键盘流程，新routingdriver两个宽度独立建A/B及候选 | native焦点/表格局部滚动/连续操作/错误恢复与截图；截图仅辅助，不替代backend行为 |
| 10 / D4 | 概览短摘要+折叠身份/时间/范围，账户分组；111项async回归 | 两屏真实反馈/冲突/读回失败/复制失败/会话恢复；关键业务步骤不能由API预置 |
| 11 / D1,D5 | 显式升级/备份/atomic/conservation边界，实际CLI契约；旧UI迁移主driver已扩展 | 精确旧main真运行→UI上游/两别名/两密钥/启停→成功与失败历史→旧backup→candidate upgrade→schema2/旧key→新UI对象→旧restore/旧服务再调用；不能用SQL夹具替代 |
| 12 / D1,D3,D4 | 保留原协议/安全/health-history测试，非监听安全检查通过 | 真实普通/SSE/工具/DNS/重定向/大小/鉴权/记录fail-closed/初始化/CLI重启/备份及空安装全程UI |
| 13 / D5 | 原固定入口保留、源码摘要扩充、任务/规范/历次失败记录 | 完整make verify真实exit0、同版本独立QA及完整验收矩阵；不提前报PR或部署 |

原validation矩阵是最低退出门槛，不因新driver只覆盖部分分支而缩减。宿主若发现缺失边界或驱动/时序/产品失败，保留回执并返回本节点补齐；不能把driver存在或模拟transport绿灯当该AC通过。

## 4. Trellis-check与跨层审查
- 已恢复实际Trellis初始化/规范（prepare回执旧标记不覆盖磁盘事实），用before-dev读取层索引、runtime/console、DOM/ownership/verification及shared guides；未凭空声称bootstrap或派生Agent。
- 写流：原生fieldsets→revision DTO→configuration事务→meta/candidates；读流：持久对象→API安全DTO→原生列表/详情。候选整体事务而非浏览器双写，旧入口410无id1旁路。新增 [可执行规范](../../../../../../.trellis/spec/backend/routing-contracts.md)及两层既有规范同步真实签名/字段/错误矩阵/正反例/测试，七节契约已维护。
- 配置短锁与provider IO分离；补management buffered response及TryLock回归，避免慢响应持锁；开始门复验授权/停用，旧快照不混用新地址/凭据。停止预算与响应提交分离，未知多尝试不能只记末次。
- 管理会话/CSRF与应用密钥隔离；token不返回、密文AAD绑定ID/地址；公开错误只安全分类。逐尝试/摘要不包含URL/原供应商正文/提示词/完成文。
- 原门禁/历史证据不覆盖；No suppressed warnings、无新增后端语言/默认供应商/外部调度依赖。格式/vet/module/build已检查；全测试、native focus、真实跨层网络/迁移尚未可执行，不勾选全绿。

## 5. 迁移与正式发布证据分层
产品schema1 SQL边界夹具仅测守恒/atomic错误分支，不是AC-11主路径。新主driver用git archive精确旧SHA、未修改旧CLI，以及显式隔离服务runner运行旧实际Go/UI，仅允许其测试上游端口；此权限只在测试中，正式CLI无放行开关。旧UI创建所有主配置并真实调用后才查询SQL对账；新对象也从升级后UI新增，旧归档恢复验证不保留它，再真实启动旧服务调用。密码/测试key只在私有子进程缓冲/环境，日志不打印，截图只在清空秘密后。

DEP-01：产品只读契约/当前schema幂等CLI局部检查通过，完整旧UI/CLI迁移仍待宿主。DEP-02～05的最新平台来源与安装进展见下方分层记录；不得将平台夹具或安装检查当作真实新旧Go联合验收。DEP-06/07：正式Actions非空升级与隔离失败恢复 **Not Run**，仍是发布依赖。产品仓未复制平台控制器、手改安装/Workflow/平台DB，未人工预upgrade替代activate。真实供应商 **Not Run**；受控网络和浏览器成功也不能升级此状态。

### 2026-10-06 平台维护源补充

证据来源：用户在本Run提供的维护者进展。本产品节点仅归档引用，未读取平台工作区、独立复跑平台测试或将用户描述补造成工具回执。不新增产品需求、不重放需求/设计阶段。

| 证据层 | 用户提供的固定身份和结果 | 可以证明 / 不能证明 |
|---|---|---|
| 唯一平台维护源 | [d0564d05e11eec271d7707ee7a444346e7be09ae](https://github.com/big91987/agent_platform/commit/d0564d05e11eec271d7707ee7a444346e7be09ae)；源PR #5仍Draft，未合main；实现契约绑定、候选显式upgrade、版本/schema健康检查、有限子进程与中断/失败快照恢复 | 记录本轮接缝实现来源，不证明该PR已合并、真实Go迁移或产品已发布 |
| 平台隔离回归与复审 | 38项Git/进程/脚本数据夹具回归通过；监督者异常退出与提交记录中断问题已修复并复审 | 平台逻辑与异常路径夹具证据，不关闭DEP-02～04真实新旧Go联合迁移/恢复/legacy验收 |
| 标准安装升级 | 先备份原可信控制器，经原install.py升级及重复安装核对；已安装控制器SHA-256 `fc51d9e7b329f3fdf9bc4cbe46fc0541fb411e096a6fbccf4aecff3a21fd226a`；服务配置、原deployed记录、current未变，旧Go实际healthz仍`85f6d17c` | 已报告标准安装和旧安装保护结果；未触发schema2产品升级，不能推断新版本兼容/恢复通过 |
| Workflow来源 | 原部署Workflow YAML与该平台源模板逐字节一致 | 模板unchanged，无需产品复制controller或手改Workflow；不削弱原安装漂移保护 |
| 正式prepare入口 | [Actions 37422891093](https://github.com/big91987/model-relay/actions/runs/37422891093)，deploy=false；真实日志为`Already deployed 85f6d17c`、`deploy skipped` | 升级后旧安装识别/健康及正式入口可用；没有重新执行make verify、候选upgrade或activate，不是DEP-06非空升级/DEP-07失败恢复证据 |

目前明确保留：真实新旧Go与控制器联合升级/恢复 **Not Run**；原Actions非空schema1→schema2升级 **Not Run**；同可信路径获授权隔离失败恢复 **Not Run**。源PR仍Draft，产品门禁、独立QA和发布条件不因上述进展自动关闭。后续各层必须用真实执行回执和固定版本逐项更新，不默认破坏当前部署根、不扩大权限、不人工预迁移。

## 6. 证据保留与下一步
本Run新日志只替换本机路径为`<workspace>`/`<user-home>`/`<temporary-path>`/`<make-binary>`，保留错误类别、行号、警告、断言、时长和失败；不改历史Run日志，不补造退出码。私有缓存、二进制、临时安装、浏览器截图维持ignore，不提交密钥/数据库/node_modules。日志路径脱敏不是修复或验收通过。

宿主tests Connector：核对final-source后实际`make verify`，保留新的完整回执及源码摘要；浏览器已有默认受控外部行为，不使用用户配置或付费凭据。任何失败先最小归因（产品/driver/异步/环境），不得轮换假设或降门禁。通过后才由独立QA对13AC/迁移/协议/secret/双屏逐项判断；本报告不是QA签字。之后草稿PR及获授权正式部署由各Connector/维护者执行，平台源PR不自动合并。
