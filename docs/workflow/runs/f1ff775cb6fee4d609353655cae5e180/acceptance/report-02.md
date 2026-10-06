# 首次调用控制台第二轮独立验收报告

> 日期：2026-10-06（Asia/Shanghai）
> 仓库/分支/Commit：big91987/model-relay / workflow/f1ff775cb6fee4d609353655cae5e180 / d62cd51e8de2127548e1ac7469d349443d1bfe3d
> 被验工作树：56文件SHA-256 `1b2f7026692d0633f6f050d585cec0bb67571a7ecf411848d8213d2fd609cdc6`；包含未提交产品、测试、契约及Trellis变更，不是旧HEAD实现
> 制品/环境：宿主seq33隔离临时实际Go服务+原生前端+受控上游+Chromium/Playwright；非正式部署，页面真实healthz版本dev
> 验收模式：全量复验（首轮No-Go、跨页/异步状态改变、多轮修订，不能按小改短轮放行）
> 结论：**Go，仅公共路线第1条切片；交 report，不代表全产品/平台或新发布通过**

## 1. 结论摘要

当前快照与宿主seq33固定 `make verify` 退出0的56文件摘要完全一致。该次1280px从空配置全UI首配、390px承接存量的主链，以及新增失败恢复、精确旧请求、分页和任务键盘两屏均通过。QA独立108项VM/helper、3项Go race契约和5项明确证据动作契约通过，保留原QA脚本3/4失败及本轮夹具失败，不机械改写历史。QA-D01/P1清密、QA-D03/P1概览同ID和QA-D02/P2部分成功反馈在原发现Contract层与新增真实UI层得到复验，QA-G01缺用例已补齐，无未关闭阻断项。真实供应商、当前候选正式部署和部署后体验仍Not Run。

## 2. 本轮覆盖

权威基线依次恢复用户授权、[PRD18AC](../product-definition.md)、[架构](../console-architecture.md)、[交互](../console-interaction.md)、[首轮报告](report-01.md)/[整改包](remediation-01.md)，并核对后续四项收尾：GET-only清密恢复与共享忙态、调用输入snapshot锁定、真实created、按对象恢复焦点。验收没有改需求或业务实现。

- 角色/范围：单管理员维护唯一上游与别名/密钥；应用密钥只生成；管理401与应用401区分；第二管理页仅验证同scope多执行者停用竞态。无需团队角色/多租户/多上游验收。
- J01首次使用：同隔离安装、同角色会话从正式页面登录、真实空计数，经UI上游/别名/密钥/普通与SSE/同请求记录；没有API/SQL预填业务步骤。既有restart为生命周期准备，并有真实重新登录恢复，不冒称重启前后同一个cookie会话。
- J02存量使用：390px与1280px共用隔离Go服务，读取真实既有success后维护/创建本屏对象/调用/精查。它不是第二份空安装；1280空安装单独承担首次使用AC。
- J03任务键盘：原生Tab到达输入/创建取消/编辑启停/复制关闭/发送取消/诊断/定位分页，focus-visible和outline分开强制核对。模型使用真实type-ahead+Tab且强制精确index和值，390px首个目标index1成功；未用selectOption/DOM赋值或脚本补首配。密码历史回归仍含click/Enter/Space，不把每个回归动作都说成纯键盘。
- J04失败恢复：地址400/变址缺凭据、写网络失败、PUT成功后注入GET503、停用密钥/模型、供应商受控拒绝、非法JSON/断SSE/网络生成失败。清秘密保公开草稿/安全诊断/ID，GET-only手动恢复，取消零写/拒绝零上游，无自动重试。
- J05运营事实：真实52次UI未知key拒绝形成50行以外旧成功，分页/并发新行/started/404/注入503和刷新/重登录精查。明确概览动作携带显示ID，并对照同记录time/revision/alias/scope；泛导航不隐式精查；真实0与null区分，不捏造归属或记录。
- J06四项收尾：恢复立即清凭据、恢复和保存共享锁；处理中参数保持实际值、四控件禁用，终态解锁；密钥created来自同ID Go响应；启停完成恢复原对象行可见焦点。迟到不解锁新操作/不抢跨页焦点通过V补强。
- 工程与安全：原Makefile/锁文件不变，Go格式、模块、vet、前端、race/实际HTTP、构建与浏览器在固定门禁顺序执行；原历史health红灯分类与日志保密保留。页面/服务日志/存储/URL/截图保密核对、应用拒绝零上游、工具仅摘要不执行。无商业功能、平台补丁、迁移或权限扩大。

J01–J06的UI-E2E证据为**宿主seq33真实执行**，不是QA本地浏览器运行；来源、注入和清理边界见[宿主审核](evidence/round-02/host-seq33.md)。[旅程契约](journeys-02.md)保留QA本地Not Run，[矩阵](matrix-02.md)独立列H同层Pass与Contract/Data等，不拼下层绿灯补用户步骤。

## 3. 缺陷及进展

| ID | 等级 | 用户可见问题与原路径 | 影响/定位 | 当前状态 | 原层及同根复验 |
|---|---|---|---|---|---|
| QA-D01 | P1 | 上游保存拒绝仍留token；应用拒绝仍留key/message。输入受控秘密→提交拒绝→应清秘密保公开草稿/ID诊断 | 安全清密；upstream form/test-form catch owner | Closed | O QA-01/04当前绿；H两屏地址400/写网络/应用401/协议断流/网络失败→清密→恢复新请求；V迟到隔离 |
| QA-D03 | P1 | 刷新/重登录后概览显示成功ID却无法从步骤5找回同ID | 核心同请求；overview-record绑定call_evidence | Closed | Q明确动作精确路径/显示、失效与迟到5/5；H两屏旧ID第二页/并发/刷新/重登录精确GET及字段对账 |
| QA-D02 | P2 | PUT已写入而随后读取失败，被误报为保存失败；保存→GET503 | 用户误重写；saved状态与load分离 | Closed | O QA-02绿；H真实PUT200→注入GET503→明确已保存但未确认→GET-only恢复且一次PUT |
| QA-G01 | 覆盖缺口 | 缺完整键盘、原依赖恢复、记录异常/分页/旧ID连续UI路径 | 不能仅以API/VM替代UI | Closed | 原make verify浏览器入口扩展；H两屏QA恢复/精查/分页完整PASS；B逐动作、最终事实审核 |

**QA-03驱动裁决**：原独立脚本重跑仍Fail（O，3/4，退出1），该条以`app.navigate('records')`表示概览步骤5，当前实际为明确`overview-record`动作，普通nav必须保持泛列表；首轮整改包已明确不能把所有records导航隐式改为某请求。原脚本不修改、不删断言、不写4/4；保留旧红灯作为历史资产，新增Q驱动实际明确动作并验证泛导航无精查。H真实页面“5 · 核对概览成功请求记录”与同ID GET跨刷新/登录/旧分页通过，才关闭缺陷主题；不是用“旧测试不对”回避用户结果。

本轮新增Q首次4/5，失败因models响应后异步尚未发出upstream GET，夹具提前应答；补一轮tick后5/5，随后使用真实记录revision字段再复验。该失败不是产品缺陷，保留attempt-01。QA命令首次使用zsh只读变量`status`中断日志尾追加，108项Node已完成；另存final日志真实exit0，不伪造旧shell成功。无新增P0/P1/P2/P3用户缺陷，不把QA夹具错误或host harness OS序列失败算产品复发；未宣称所有Chromium/OS机制通用正确。

## 4. 证据

- [版本与依赖](evidence/round-02/baseline.log)：HEAD/分支/56文件摘要、Go1.25.6、Node26.0.0、npm11.12.1；Playwright锁1.59.1，宿主未输出完整Chromium build。
- [H宿主seq33](evidence/round-02/host-seq33.md)：原样make verify退出0，当前Go包113.488s、health-history4.944s、build、execution98536和两屏主链/恢复PASS；中段省略不补造完整stdout。
- [V独立前端final](evidence/round-02/frontend-final.log)：node --test tests/browser/ui.test.mjs，108/108，exit0、无skip；VM/helper而非真实浏览器。
- [G独立后端](evidence/round-02/console-contract.log)：3个Console契约，Go race -count=1 exit0、8.582s；不监听真实HTTP，不替代H。
- [O原边界](evidence/round-02/original-boundaries.log)：4项3绿1红exit1；[Q新明确动作资产](evidence/round-02/explicit-evidence.mjs)/[首次夹具失败](evidence/round-02/explicit-evidence-attempt-01.log)/[最终复验](evidence/round-02/explicit-evidence-verified.log)。所有证据层与真实退出码分开。
- [归档检查](evidence/round-02/archive-checks.log)：execution98536每屏71行全部目标index/值一致；只归档各屏首行含LF原字节，[1280精确行](evidence/round-02/seq33-model-1280.jsonl)/[390精确行](evidence/round-02/seq33-model-390.jsonl)摘要、字节及阶段一致。1280首次index0不用于证明移动；390首次1→1及非首项真实闭环有效。
- [1280截图](evidence/round-02/host-seq33-console-1280.png)/[390截图](evidence/round-02/host-seq33-console-390.png)：最新宿主清密概览，QA目视审核。仅证明收尾可见页面和布局，不冒充每个异常/焦点步骤截图。
- [最终检查](evidence/round-02/final-checks.log)：diff/语法/本地引用/无本机路径/原门禁与锁不改/最终源摘要。首轮及研发各次失败均保留，不覆盖其他Run。

## 5. 未执行与环境限制

QA没有在已确认受限沙箱重新监听并运行完整make verify/Chromium；按AGENTS采用固定host-only Connector同快照同层证据，无扩权、换入口、静态browser.check或VM冒充。矩阵LOCAL-UI仍Not Run，该次H实际两屏同用户结果覆盖强制门槛。极端忽略abort和已知deadline悬挂组合是确定性Contract，不称真实浏览器网络复现；已有真实session故障/恢复在H单列。

无真实供应商凭据，联调Not Run；协议夹具/管理503与网络注入不说明真实供应商/实际数据库故障经历。新PR、合并、正式测试部署、部署后实际页面/healthz对应合并SHA均无回执，Not Run，历史main部署不证明工作树已部署。无全WCAG/所有OS浏览器认证，无新正式操作手册截图，不把受控dev页面称生产健康。

## 6. 阶段边界与公共状态

本轮只验公共路线第1条。路线2–4已授权持续目标，由目标维护者依赖前切片完成后以独立Issue接续，当前未实现/未新增Issue编号；成员角色/额外协议按实际场景选。获授权目标维护者可在核对QA后自主完成测试仓合并与正式测试部署，不等用户逐阶段批准；本节点不代替Connector执行Git分支/提交/推送/PR/合并/部署，不自动合并平台源PR。

本轮T001/T002/T003退出条件已验，公共任务标Done；M01/M02/M03标Accepted，**限本切片实现与验收，不是Delivered或Published**。delivery历史待验和首轮No-Go保留，当前索引指向本报告。项目级PRD/产品/路线单一事实源未复制或改写；README/基础版本等较早“待QA/只完成需求”进度句是阶段快照，不扩大范围，后续report在对外说明同步最新实际结论，勿继续引用为当前状态。若后续改产品/测试/范围必须新指纹复验，不能沿用本报告放行变更对象。

## 7. 环境恢复

QA新Contract无网络/持久业务副作用，G隔离临时store/in-process transport由测试cleanup关闭删除。H源审查确认defer shutdown/close/remove，browser finally解除注入、清受控测试剪贴板、撤销隔离context权限并关闭浏览器，exit0与正常清理一致；没有独立OS级清理审计，不冒称已做。保留忽略cache/build/node_modules/test-results，归档只含清密截图/固定标签JSONL/脱敏回执，不提交DB、秘密、本机路径或runtime session。不触达用户上游/数据，不修改数据库造绿，无平台修改或权限扩大。

## 8. 质量趋势

相对首轮：关闭3缺陷主题（D01两分支合一）和1组覆盖缺口，无新用户缺陷/确定复发；108项独立VM/helper较60项增强，新增5项QA明确动作/失效/迟到回归资产。失败清密、部分写成功恢复、旧ID跨刷新/登录/分页、完整任务键盘和四项收尾已回到真实Go两屏；不再仅靠Contract断言关闭网页风险。seq31原箭头失效有完整白名单证据，seq33type-ahead精确结果有效，不反推OS底层根因。剩余风险集中在真实供应商、正式部署后验证、后续多上游治理以及更多浏览器/OS，而不是把第一条切片称完整产品。

## 9. Go / No-Go

**Go**：当前指定快照固定门禁真实exit0、全部强制AC有对应同层证据，QA-D01/D03原P1和D02原P2关闭，QA-G01覆盖补齐。旧QA-03驱动Fail仍保留但不满足当前明确动作语义，门槛由同轮Q/H同结果与关键不变量证据满足；未执行项原样列明，不按通过率隐藏。走合法`next → report`，交当前报告、矩阵、旅程及证据供后续PR说明；后续发布按真实回执继续。
