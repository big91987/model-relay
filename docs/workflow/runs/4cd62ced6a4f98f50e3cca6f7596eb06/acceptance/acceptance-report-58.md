# design-v0.2.0 / records-detail-entry-r1 独立验收报告

> 日期：2026-10-08，Asia/Shanghai；Run 4cd62ced6a4f98f50e3cca6f7596eb06，QA执行58。
> 仓库/分支：big91987/model-relay / workflow/4cd62ced6a4f98f50e3cca6f7596eb06。
> HEAD：98511771871cf0951ecef716bbb55deb13e18da5；候选仍为未提交工作树，134件指纹 fb4066fa37d1e7865d3a18f72dd78fef98ae79bde17e20875af373b24771c3a6。
> 验收模式：全量范围恢复与当前宿主全量回归审查；QA独立非监听复验、逐页设计对照。不是仅验证研发改过的定位器。
> 环境：宿主实际Go临时HTTP服务、SQLite schema3、真实内部用户会话、受控HTTP/SSE上游、Playwright 1.59.1；Go1.25.6，Node v26.0.0。不是生产部署或供应商事实。
> 产品候选结论：**Go with known issues，可交 report 整理交付**。完整产品/平台/部署最终退出尚未完成，不是H01–H08全部Go。

## 1. 结论摘要

固定宿主入口 `make verify` seq57 exit0，106669 bytes，归档truncated=false；按next_offset恢复至eof。QA用独立Node实现重新计算134件路径/字节源戳，结果与研发56、宿主57完全一致；材料archive及19件manifest条目全部匹配。测试不是上轮seq46或seq55的绿色移植。

三视口1280/1440/390本轮J01–J07全部完成，包含J06原调用日志行的native详情入口、同Request ID核对前后及CSV/账本守恒。QA独立查看本轮30张十页截图、三张请求抽屉、一张移动Key编辑、四张账单更正前后图，并审查对应动作、真实API对账和所有63张visual图片的摘要/采集合同。**QA47-01 P1及QA47-02 P2在原层关闭**；AC01主要管理摘要、AC15原始与确认用量语义恢复。

QA另实际执行193项Node回归、格式/模块/vet/diff及五项Go race非监听测试，均exit0。没有重复原生已知不能监听的完整入口，没有用静态browser.check证明后端。新增一项P3：若干服务状态、角色和请求详情仍显示英文枚举/机器字段且窄列断词，不阻止业务链、权限或金额核对；后续由development改进，不因此追加阻断轮。

## 2. 基线与本轮覆盖

唯一冻结输入仍为design-v0.2.0 / Frozen for Test Implementation，archive SHA256 `5f51d35fdc15ed7214749f9413879b4238097f1d6b6bbe2cc0d508d8e103cef7`。实际manifest及所有引用路径/逐文件摘要见 [integrity.json](evidence/attempt-58/integrity.json)。本轮恢复PRD、interactions、acceptance、decisions及baseline-review最后闭环，以最后整改结论为准，不重开R02–R04。原型A只是布局/动作基线，不是供应商/费用事实；未执行包内脚本。architecture-r1、implementation-contract与verification-design继续承接。

产品覆盖包含：平台管理员、租户管理员、开发者、财务观察者；首次空安装、正式接入/发布/授权/邀请/接受/登录/签发；普通/SSE/取消/同ID诊断；授权收窄、密钥立即/过渡轮换/截止/撤销；两scope五维限额及共享池并发；保守预留/未知用量/多尝试/独立成本租户费；价格快照、结账拒绝与合法结账、追加更正/CSV；搜索/分页/迟到/错误重试/退出清除；schema1/2→3正式CLI升级、备份恢复与失败保护。逐项层级结果见 [当前矩阵](acceptance-matrix-58.md)，完整角色/起始状态/连续步骤见 [旅程表](user-journeys-58.md)。

J01/J05是按正式责任交接的多用户链：平台登录会话创建前置对象，受邀负责人在独立真实会话接受/登录，然后管理本租户；不宣称一个用户跨角色伪切换。J02/J03/J04在同一租户管理员会话连续执行，J06在平台管理员会话连续操作（与财务只读反例分开），J07在当前真实身份与scope中执行导航/分页/故障隔离。没有API/SQL替用户补UI中间步骤；只读API用于最终事实核对。Go runner仅初始化首次安装、提供受控上游/隔离目录以及显式测试时钟。测试中的2026-11周期来自跨月测试时钟，不是现实系统日期或生产账单。

## 3. 缺陷及进展

| ID | 等级 | 用户问题/复现入口 | 影响及本轮事实 | 状态 | 回归资产 |
|---|---|---|---|---|---|
| QA47-01 | P1 | 两模型/两租户/多Key，查看模型、服务、租户、工作台管理摘要 | 当前补齐来源/能力/价格/真实或未测延迟、模型启停/调用入口；服务关联数/候选优先级/核验时间及适用revision；实际负责人/成员/grant/Key/预算；工作台状态与管理入口。三视口native动作和API数值匹配 | Closed | tests/browser/console-summary.mjs、console_summaries_test.go；本轮management-summary三JSON及visual四页 |
| QA47-02 | P2 | 同未知Request ID经网页核对输入3/输出4，再打开日志/详情/分析/导出 | 日志解释原始上游总量未知，另列确认计费3/4，pending0；抽屉及CSV同口径，原attempt/快照不变、核对幂等不重收费 | Closed | journeys.mjs J06-reconciled-token-meaning；Summaries API/race、账本对账；三视口records图 |
| QA58-01 | P3 | 服务列表核验/人员角色及日志详情查看，1280/390窄列显示success、administrator、dispatch_possible等 | 技术枚举与断词降低扫读效率；并非缺少核验事实、角色权限或详情数据。核心操作和同ID核对仍可完成 | Open，非阻断 | 本轮services/tenants/records-detail图片；Owner development，建议后续统一文案与列宽，不擅自改冻结范围 |

QA22连续UI、QA35未知写入读回及双发送、QA38逐页采集、后续编辑所有权和route清理主题均保留历史编号/失败证据。本轮宿主对应链重新通过，没有通过删断言/缩超时/改变金额/跳过J06关闭。seq55驱动误用详情入口的原失败和研发56三项红转绿仍位于validation/attempt-56。本报告不抹除旧No-Go。

## 4. 实际证据与版本一致

| 执行来源 | 实际命令/结果 | 证据/边界 |
|---|---|---|
| tests Connector seq57 | make verify，timeout900s，exit0；106669 bytes/truncated=false | [宿主审查](evidence/attempt-58/host-seq57-review.md)；正式日志用read_command_output(seq57)恢复，不是QA亲自执行 |
| QA58本地 | 独立源戳/材料重算及84件脱敏宿主产物保留，exit0 | [可复验脚本](evidence/attempt-58/check-integrity.mjs)、integrity.json；134源文件与材料逐件SHA，69张PNG/15个JSON |
| QA58本地 | make format-check、go mod verify、make vet、193 Node、git diff --check，exit0 | [qa-static-node-final.log](evidence/attempt-58/qa-static-node-final.log)；193 pass/0 fail/0 skip |
| QA58本地 | 五项Go非监听回归，-race -count=1，exit0 | [qa-inprocess-final.log](evidence/attempt-58/qa-inprocess-final.log)；Summaries/OverviewAndExactRecords/Accounting/PeopleScope/MicroDollarPrecision，不冒充HTTP/UI执行 |
| QA58独立视觉/合同审查 | 30页逐页及8张关键图，对照批准材料与native断言 | [visual-review-58.md](visual-review-58.md)；截图、UI-E2E、API、账本各层分别记录 |

QA两个最初日志包装命令在测试后使用zsh只读变量status，包装exit1。保留qa-static-node.log及qa-inprocess.log，不用其作为门禁通过；改用result变量重跑同一检查，最终真实exit0日志列于上表。这是QA命令包装错误，不是产品/宿主门禁失败，没有修改产品源码。

Go日志保留macOS链接器LC_DYSYMTAB warning；仅将本机临时目录前缀脱敏为<temp>，没有删除warning或改变测试输出/结果。以实际exit0及无race失败判定，不把warning伪装成无警告。

健康历史门禁首段有隔离旧版本预期红，当前版本HealthMethodsHTTP/Handler后来真实绿；不能把该历史exit1误读为seq57失败，亦不把历史红重建称作原开发成功。动态页面服务标记continuous-controlled-ui/schema3属于测试runner；当前源码由入口指纹绑定，**该标记不证明未来Actions制品的merge SHA**。

## 5. 未执行与限制

- 原生QA完整make verify/新浏览器监听重跑仍Blocked（原限制未变，本轮未重复）。替代为同指纹宿主seq57真实HTTP/race/UI；本行不改Pass，也不拼API与截图去假造未执行UI。
- 独立首次self_hosted网页分支未单独运行，仍Not Run；本轮provider网页接入/核验与两来源ActualHTTP合同覆盖同一编辑器/后端来源事实及故障恢复，作为当前AC02内部受控门槛的等价补强，不称该分支UI已跑。
- 真实供应商与用户已有自托管端点凭据/响应、供应商真实计价、外部支付/外部身份Not Run；当前内部身份及USD测试账本已验证，按用户授权不阻断内部候选验收。
- H02重复入站唯一性、H03原生实际模型、H04负向材料安全恢复、H08平台维护源/模板升级完整回执仍Not Run；配置声明不是实际模型证明，产品测试不是Harness负向证明。
- H07同产品PR/base/head/merge、Actions/Deployment、实际/admin/页面及/healthz同版尚Not Run。HEAD是未改main，不可当最新产品制品；8793是编排平台，不是产品效果URL。

## 6. 阶段边界与下游要求

本节点只对当前内部产品候选作Go with known issues。允许继续同Run report→publish→pr，不因此把所有H/发布写Pass；最终完整退出须由对应Connector/协调者以正式事实闭合。本轮没有需求/设计返工或用户澄清事项。

report应在Run根pr.md绑定本候选、134件指纹、seq57和本QA报告、Issue25与随后真实PR/commit；不可预写PR号、merge SHA、部署URL或成功。发布节点提交后如产品/测试/锁/源戳合同变化，必须重新make verify及相应QA；只添加报告/脱敏证据仍应检查指纹不变。

用户2026-10-08授权测试仓合并和正式Actions部署持续有效。源码已有Deploy Model Relay locally仅verified main prepare，正式workflow_dispatch执行deploy=true/target=preview（validation隔离不替代preview）。preview5545仅为用户给出的目标配置；最终入口必须从实际Deployment environment_url核对，再给/admin/及/healthz。Coordinator/Connector负责核验后执行，不由本Agent发GitHub动作、不用旧main/临时服务冒充最新。report给安全的私密文件初始化/首次登录指引，不写密码/密钥、不把未实跑部署手册称验证通过。

## 7. 环境恢复

宿主Go runner拥有随机监听端口/临时目录，defer关闭测试HTTP服务/Store并RemoveAll；browser contexts在finally关闭，controlled上游不访问用户服务。宿主入口正常完成并给isolated_directory_servers_contexts清理标识；未由QA额外探测宿主进程表，因此运行后无残留端口的独立External State检查Not Run。QA五项inprocess只用临时测试数据，不改生产/fixture库造通过。本轮只新增QA报告/验证资产/脱敏产物和索引，不改业务实现、冻结材料、测试门禁或依赖；缓存不作为交付产物。

## 8. 质量趋势

较QA47提高：关闭1个P1和1个P2，管理摘要由缺失到真实多对象事实与上下文动作；已核对用量由易误解到列表/详情/分析/CSV一致。J06原层全绿、三视口完整采集与193Node持续保留；此前宿主seq49/51/53/55失败没有被隐藏。新增一个P3扫读问题，不涉及权限/账本或核心功能。剩余风险主要是正式部署与平台/供应商事实，不能由产品测试数量代替。

## 9. 判定

当前产品候选无Open P0/P1，固定入口同版通过，核心连续旅程具真实宿主执行证据及独立QA审查，**Go with known issues → report**。P3由development后续处理，无须为低风险文案再阻断交付。完整US08/H07和用户可直接查看的正式页面尚未完成；在实际发布事实齐备前，禁止声称“全套最终Go”“已部署最新版本”或“已全面对齐New API/LiteLLM商用能力”。
