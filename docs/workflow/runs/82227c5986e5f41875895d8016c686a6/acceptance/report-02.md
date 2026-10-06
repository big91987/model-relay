# 路线2整改版本11独立验收报告

> 日期：2026-10-06（Asia/Shanghai）
> 仓库／分支／Commit：big91987/model-relay；workflow/82227c5986e5f41875895d8016c686a6；HEAD 85f6d17c07c5fddeb2d6a39a5a067769a11bb4cc，有未提交实现与本Run文档。
> 产品快照：85文件，852f7bfc90323b6253658c8b9c2167fa22c43a5f59389f0bbcb92539865120e7；同版宿主seq21 make verify exit0。
> 环境：隔离真实Go/SQLite、受控外部上游与真实Chromium双屏；单管理员、单实例。不是共享正式部署。
> 验收模式：全量整改复验；原核验、前端任务生命周期均改变，不能仅沿用08版绿色。
> 结论：**Go（当前产品切片交付／进入报告与草稿PR）。不是正式发布或真实供应商验收Go。**

## 1. 结论摘要

原QA-01～04四项P1在原发现层关闭：严格列表/失败状态的API与持久化契约已独立race复测；10秒头/body预算已由宿主真实HTTP复测；完成步骤默认折叠及精确请求回看已在两屏真实网页连续回归。新增核验取消/去重/超时/离页/纠错、健康代次及工具停止边界有同版本结果，不再把静态代码或VM绿灯当作浏览器通过。原完整固定入口、真实外部计数、旧main非空UI升级及备份回退均在当前11快照重新执行成功。本轮未发现新的未关闭P0/P1，允许按既定next进入报告；正式可信部署联合链及供应商未执行项不得由该结论补绿。

## 2. 权威基线与变化

验收预期仍为本Run [PRD](../product-definition.md)、[架构](../routing-architecture.md)、[交互](../routing-interaction.md)、[验证计划](../routing-validation.md)、[迁移ADR](../adr-001-storage-upgrade.md)与[部署合同](../deployment-upgrade-contract.md)。恢复了 [首次全量报告](report-01.md)、[原矩阵](matrix-01.md)及[整改包](remediation-01.md)，均保持历史结论不变。

上轮08／QA测试快照至当前11的受影响域：核验响应解析、10秒出站context与终态落库；前端取消/owner失效、概览完成步骤移动；原浏览器旅程及HTTP/状态回归增强。schema、迁移协议、调度算法及默认300/30/60/30预算方向未因整改改写；核验10秒本来就是基线。seq19发现的详情driver文案等待错误仅改为真实请求身份响应及页面首行核对，未降低产品契约。原QA测试SHA仍为aafefad09c238b976b850c44eda7960f74ea2e5cb6b7bb96b36794cfddb4e9b4。

## 3. 覆盖与连续用户旅程

详见当前权威 [matrix-02.md](matrix-02.md)。不是将首配页面存在与API创建拼成Pass：

- J-FIRST-1280/390：原真实Go控制台从初始化/每屏新对象进入，登录、配置纠错、密钥签发/原生复制及拒绝恢复、调用、同ID记录、会话过期安全重入在本屏同一角色/会话完成；390是实际任务而不是桌面收尾截图。
- J-ROUTE-1280/390：每屏同一管理员会话从菜单创建两个本屏上游、核验、失败/取消/超时/离页再入、上游编辑纠错、双候选创建/编辑/取消、签发密钥、权重/优先级调用、故障切换、详情、冷却恢复和停止直到重启再登录。各组实际生成调用由UI发出，业务对象无API/SQL预建；控制API仅切外部故障、读计数或执行显式隔离服务重启。
- 核验与生成分开：列表GET零生成；合法列表才成功，非法/401/429/超时为失败；取消时外部context停止且持久化canceled，重读不再次核验。10秒等待覆盖头与body，不伪称Transport探针是真实网络等待。
- 每屏原生Tab/Enter/Space/type-ahead编辑并验证focus-visible；无selectOption/fill/setChecked补路由中间步骤。密钥从一次性UI读取后用真实按键输入，是自动化模拟用户转录，未经管理API获取；原生复制成功/拒绝在J-FIRST同轮各屏单列真实剪贴板验证，不把其步骤说成routing.mjs执行过。
- 完成后待办0项、回看5项默认关闭，原生展开→原成功Request ID精确详情→返回收起；重启后待办2项、回看3项，不伪造当前启动健康。当前两屏概览截图也确认账户分组、简短状态及无技术栈页脚。
- J-MIGRATE：精确旧85f6d17真实Go用户页面形成非空上游/两别名/启停密钥/成功失败历史，再正式backup→upgrade→候选启动/原key调用→新增UI对象→旧backup restore/旧服务调用。SQL仅守恒对账，不造配置或请求。该Pass不扩展为controller/Actions升级Pass。

角色裁剪为当前单管理员和应用密钥调用方；多租户、组织权限、应用配额治理、分布式调度属于后续路线，不新增本轮门槛。

## 4. 缺陷状态与原层级关闭

| ID | 等级 | 用户问题／复现路径 | 当前状态 | 实际关闭证据／回归资产 |
|---|---|---|---|---|
| QA-01 | P1 | 核验错误Content-Type/object/id却报成功 | Closed | 原TestQAConnectionCheckProtocolAndFailureStates独立exit0；API/SQLite失败事实；同版真网页非法列表失败/合法列表成功 |
| QA-02 | P1 | 核验预算30秒违反10秒头/body界限 | Closed | 原预算负控独立通过；seq21 TestConnectionRealHTTPHeaderAndBodyBudget两个真实网络分支；网页duration_ms及外部取消计数断言 |
| QA-03 | P1 | 坏JSON/429/503/network确定失败误报未知 | Closed | 原失败分类负控独立通过且持久化一致；真网页401/429/非法/超时失败，取消独立canceled；不抹掉旧历史unknown |
| QA-04 | P1 | 完成首配后完整步骤仍占当前下一步 | Closed | seq21双屏真实完成0/5、默认关闭、原生展开/精确同ID详情/收起、重启2/3；新390/1280首配截图和DOM相符 |

开放P0=0、P1=0；本轮没有新增实现缺陷或重新分级。首轮缺失的取消/去重、原生双候选键盘、整数边界、全冷却/启停保留/旧健康晚到/重启未知和工具增量停止由当前资产补齐。实际浏览器不重复穷举所有API非法数值；对应API/持久化边界行独立Pass，不将其包装成UI边缘输入全部实测。

## 5. 证据及版本一致性

- [宿主seq21索引](evidence/host-seq21.md)：900秒入口实际exit0、完整79939 bytes retained存档，分页重点检索到eof；本文件只是索引，不伪装完整日志。08/seq16和10/seq19不用于当前11放行。
- [version-lock-02](evidence/version-lock-02.json)：QA自己go run ./tests/evidence与11/seq21一致，工作树而非仅HEAD绑定；Workflow a3d814fa…及manifest 731a6778…与原锁和[安装provenance](../delivery/verification/platform-template-provenance-01.json)一致。source工具实际清单不含acceptance/delivery回执或Workflow/manifest，后者分别锁定。
- [contract-retest-02](evidence/contract-retest-02.log)：QA独立 `GOCACHE="$PWD/.cache/go-build" GOTOOLCHAIN=local go test -race -count=1 -v -run '^(TestQAConnectionCheck|TestConnectionCancel|TestCandidateInteger|TestRoutingHealthEpoch)' ./internal/relay`，exit0／13.473秒；与宿主完整入口分开。日志只将链接器绝对临时路径替换为`<temp>`，保留警告/断言/实际结果。
- [frontend-retest-02](evidence/frontend-retest-02.log)：QA独立node --test原ui.test.mjs，119/119、0 fail/skip/cancel，exit0。VM用于异步owner/迟到/过期/重启状态契约，不代替真实UI。
- [final-audit-02](evidence/final-audit-02.log)：相对引用、13组AC、源摘要、辅助证据哈希、原QA负控/模板/manifest、历史Run及可信门禁未改检查，exit0；报告产物不改变已通过的85文件产品快照。
- 本轮归档 [1280概览](evidence/console-first-call-1280-02.png)、[390概览](evidence/console-first-call-390-02.png)、[1280重启回看](evidence/routing-1280-02.png)、[390重启回看](evidence/routing-390-02.png)与[85660选择轨迹](evidence/routing-selection-85660.jsonl)。QA实际查看概览双屏与390重启截图；本地截图不是独立启动浏览器，也未由Connector签署原始哈希，辅助证据不能替代主回执。JSONL关联PID/156行，但仅证明原生选择，不证明全部其他步骤。

## 6. 未执行、限制及后续边界

QA沙箱listen/Mach限制不变，完整make verify由既定宿主Connector执行；本轮未扩大权限、未用静态browser.check当后端验收、未反复跑必然受限完整命令。QA自己的HTTP/浏览器完整重跑记Blocked，同轮同版本宿主真实链是对应主用户结果的门槛证据，不将Blocked项改成Pass。

DEP02～04真实可信controller与旧/新Go联合顺序、失败恢复和legacy保护 **Not Run**；DEP06正式Actions非空旧安装升级、DEP07同可信路径隔离失败恢复 **Not Run**。标准安装只证明受管理模板/manifest一致，不覆盖所有DEP05安装漂移/失败边界；用户提供45项平台回归/复审和控制器安装不等于本节点复跑或Go数据迁移。真实供应商无凭据/费用授权仍 **Not Run**。这些明确属于既定发布后续链，不是已验产品实现失败，也不要求在草稿PR之前伪造发生。

产品13组AC在本轮实际证据范围满足交付门槛；未宣称所有外部环境和输入排列均测试过。生产默认时间值由配置契约与计时专项验证；缩短配置的HTTP停止证据不表示跑了完整300秒生产SLA。实供应商能力/费用、真实部署长期稳定性不从受控上游推断。

## 7. 环境恢复

QA专项沿用隔离临时安装/模拟Transport，测试正常退出并由cleanup回收，无共享数据/凭据或真实上游操作。seq21浏览器runner正常退出，源码有context/browser.close、服务shutdown与临时目录defer；本节点未独立枚举宿主进程，因此不宣称全机残留审计。临时DB/key/backup、node_modules/cache/test-results仍ignored；只归档已清秘密页面及白名单轨迹。QA未修改业务源码、测试预期、Workflow、平台安装或运行数据，未提交/推送/触发Actions。

## 8. 质量趋势

首轮四项P1从Open变Closed，新增0、已证复发0。原宽松成功fixture已改合法列表；负向格式/状态、真实头/body预算、核验取消/错误恢复、完成步骤和工具停止沉淀为原固定入口资产。seq19仍揭示driver详情身份契约漏测，本轮11精确身份替代脆弱文案并获得同层真实回归，不以VM代替。当前119项VM及完整宿主链都绿，但结论依据用户结果/最终事实而非单纯数量。剩余风险集中于平台真实可信联调、正式发布与供应商，层次比首轮更明确。

## 9. Go / No-Go

**Go，next→report。** 原四项P1已原层级关闭，当前快照完整固定门禁及受影响连续用户旅程通过；不再返回研发放宽规格。report据当前矩阵写本Run pr.md并交既定Connector完成草稿PR，正文必须明确未发布、DEP02～04/06～07及供应商Not Run。正式合并/Actions部署按用户既有授权由目标维护者后续执行并核对真实merge SHA、Deployment、healthz、schema、非空数据及浏览器；平台源PR不得自动合并。若后续产品或模板改变，当前版本锁失效，必须对应新版本重验，不沿用本结论。
