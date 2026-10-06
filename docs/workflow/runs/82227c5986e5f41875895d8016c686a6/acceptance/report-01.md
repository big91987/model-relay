# 多上游切片独立验收报告 01

> 日期：2026-10-06（Asia/Shanghai）
> 仓库：model-relay；分支 workflow/82227c5986e5f41875895d8016c686a6；HEAD/精确旧main 85f6d17c07c5fddeb2d6a39a5a067769a11bb4cc；实现为未提交工作树，不把HEAD当新实现版本。
> 被验产品快照：08，82文件，sha256 c340664b231d83d153f0d2569065db08c8be34661a314f51e7831f74f808b27e。Workflow/manifest另联合核对 a3d814fa47da0c74e6934ab46138f2b854c55cf1de177f7fba538b4365ab2f45。
> 环境：宿主 tests Connector 临时实际Go/SQLite/受控上游/Playwright Chromium；Go1.25.6，Node26.0.0，锁定npm/mod；1280px及390px。测试运行版本dev，由原health/browser回归核对，不代表已部署新commit。
> 验收模式：全量（跨配置、数据面、迁移及UI）；宿主同版本实际门禁证据审查＋QA独立免监听负向执行。不冒称QA自己运行了宿主完整门禁。
> 结论：**No-Go，development整改；不得进入报告/草稿PR放行或宣布发布成功。**

## 1. 结论摘要

宿主 seq16 原固定 `make verify` 完整真实exit0，08产品及当前模板联合身份吻合。两个屏幕实际用户页面创建上游/候选/密钥、真实路由分配、有限切换、未知用量、冷却恢复、SSE/取消停止，以及精确旧main非空UI升级/回退均已有真实受控证据。但绿色门禁未测全核验合同：QA独立race负控确认非法列表也记success、出站预算30秒而非10秒、已知失败记unknown，三项P1开放。核验取消/失败恢复、完整多候选原生键盘及部分冷却竞态等强制覆盖仍缺；概览完成步骤折叠也与规格不符。故不签13AC全部通过，返回研发修复和补测试，不降低既有需求/设计。

## 2. 本轮覆盖

权威基线：本Run product-definition R-1～7/AC-01～13、routing-architecture、routing-interaction、ADR-001、deployment-upgrade-contract、routing-validation，及唯一公共roadmap/首版安全协议回归。规格先于当前实现。单管理员/单实例及应用调用方裁剪；多租户、应用治理、商业计费与路线3/4不适用，不造相应运营文档。

[matrix-01](matrix-01.md)是当前权威需求—旅程—接口/状态—测试—证据矩阵；分开UI、API/持久化、真实外部接收和未执行。J-FIRST两屏、J-RUNTIME两屏和J-MIGRATE分别由一个角色/作用域/会话连续完成，业务对象来自UI，fixture只提供外部行为。核验实际显示success却接收不合法列表，J-CHECK为Fail，不将该步骤从完整功能结论中隐藏。J-KEYBOARD多候选完整原生键盘为Not Run，已成功的创建候选type-ahead及旧单候选键盘仅作相邻回归。

工程/协议：格式、mod验证、vet、116项VM、Go race CLI/handler/真实HTTP、build及真服务浏览器由seq16实际运行。历史9b6994ed HTTP红控有正确分类，不是当前失败；新核验P1与该历史红控无关。多次调用真实外部计数、不同远端名、精确Request ID与逐尝试未知合计均核读原断言。SQL仅隔离迁移对账；旧请求unknown/started和UI503等显式注入单列范围，未声明供应商执行事实。

## 3. 缺陷及进展

新增 QA-01/02/03 三项核验P1及QA-04强制概览合同P1，详见 [remediation-01](remediation-01.md) 的复现/影响/Owner/回归资产。合法空列表和401正控成功、9个负向子例失败，另预算测试失败；独立命令exit1，不是权限阻塞或猜测。API结果与真实SQLite状态相同错误，属于实现缺陷，不是需求歧义。QA-04按当前强制AC-09/R-6分级，而非任意提高一般外观建议的严重程度。

未宣称新P0，未将未执行用例伪装成失败测试或接受风险。概览完成态静态/截图偏差及取消核验入口缺口均需落实原合同并在真实浏览器关闭；本轮不冒称已执行这些新UI负控。覆盖债必须补入固定入口才能完成剩余强制AC；不以默认P2机制绕过缺失验收。没有Owner接受的新豁免。

历史CLI旧API/候选label和macOS原生选择问题已有同层宿主seq16回归绿色，说明该受控路径较seq10/12/14改善；不声称每次历史失败唯一根因全部证明。失败证据仍保留，问题主题不丢弃。本轮三项核验缺陷是先前覆盖遗漏发现，不改旧报告为新失败。

## 4. 证据及版本一致性

- [host-seq16](evidence/host-seq16.md)：真实宿主exit0/900秒/74864 bytes retained log，检索到eof；完整档案继续由平台seq16持有，落盘文档只是重点摘录索引。
- [version-lock-01](evidence/version-lock-01.json)：QA开始原evidence命令返回08相同摘要；添加唯一QA测试后独立重构原82文件清单排除新资产，仍与08相同；当前YAML/manifest摘要匹配。没有改产品实现或门禁。
- [connection-contract-01](evidence/connection-contract-01.log)：QA自己执行的Go race/handler/Transport/持久化专项，真实exit1。仅链接器临时绝对路径替换为`<temp>`，失败、断言及退出码保留；不打印密码/token/正文。新资产 `internal/relay/qa_connection_contract_test.go` 原入口自动纳入。
- [qa-source-01](evidence/qa-source-01.json)：加入QA测试后的83文件摘要；这是测试扩展快照，不是已通过seq16的快照，不沿用旧绿放行。
- 两屏概览/路由截图及 [75584原生轨迹](evidence/routing-selection-75584.jsonl)已复制到本acceptance/evidence并做摘要。截图来源是宿主本地产物，QA只读观察，不是本轮另启动浏览器；Connector未返回其原摘要，文件本地存在/时间仅辅助身份。

## 5. 未执行和环境限制

当前沙箱已知listen/Mach不允许，按工程授权使用固定宿主回执，不扩大权限重跑或用静态browser.check冒充Go服务。QA本轮实际完成免监听负控，不重跑已知受限全入口。QA测试加入后完整新的make verify尚Not Run，而且局部明确红灯；修复后必须新的完整宿主回执。

核验错误/取消/超时UI闭环、多候选完整原生键盘、健康晚到/全部冷却/重启范围以及部分工具流备用停止组合仍Not Run，详列矩阵，不通过拆分底层或历史证据补成Pass。概览已完成列表依当前实际截图没有折叠回看，待修和真实路径复验。

DEP-01真实产品旧UI非空迁移Pass；DEP-02～04真实旧/新Go联合可信controller、DEP-06正式Actions非空升级、DEP-07可信隔离失败恢复及真实供应商均Not Run。标准Workflow安装/重复来自研发实际回执并由QA核对当前受管理文件身份；用户报告的平台45项夹具/独立复审与控制器安装不是本节点复跑或真实Go升级证据。原deploy=false Actions只prepare旧版本，不能证明新版本发布。

## 6. 阶段边界

本结论是当前产品切片的独立QA，不是合并/正式部署/供应商Go。设计已明确DEP-06/07不需草稿PR前发生；未发布不构成本轮P1的归因。真正本轮阻断是核验/概览合同及强制覆盖未完成，应返回development，不退回规格放宽。目标维护者后续负责平台源修复/标准路径/正式Actions，对未执行项继续如实标注；平台源PR不自动合并。本节点没有Git分支/提交/推送/触发Actions或改平台数据。

## 7. 环境恢复

QA新增测试使用已有 `installation(t)` 的独立临时目录、模拟Transport，不启动监听，不触及用户配置或供应商；测试正常完成收尾，临时资源由测试cleanup销毁。宿主浏览器/迁移runner有实际context.close、HTTP shutdown与目录defer清理，seq16正常退出；没有新的共享环境变更。QA进程枚举被沙箱拒绝，未独立核验宿主零进程残留，不把源码清理意图说成实测进程审计。截图/无秘密JSONL留作证据，数据库/key/归档未进仓库；原test-results/node_modules缓存继续ignored。

## 8. 质量趋势

相较研发最近四轮，宿主固定门禁从预算中断/候选原生选择失败改善为完整绿，真旧UI迁移与双屏多路由运行链有可重复资产。独立QA进一步发现核验假阳性/时间/状态漏洞及已明确承接的概览合同遗漏，说明门禁执行完整但需求负向覆盖不足，不能只报116/116。沉淀两项真实race回归入口（含11种列表/错误场景），本轮四项新增P1、无已关闭新P1、无已证核验复发；剩余UI/竞态/正式发布及供应商风险明确。既有选择driver改善保留，但质量尚未达到完整切片可交付标准。

## 9. Go / No-Go

**No-Go。** 四项P1和强制覆盖缺口使13组P0 AC不可整体通过；seq16绿色不豁免产品规格。development按整改包先修核验、概览、合法fixture与缺失验证，再宿主同一900秒 `make verify` 获得新版本完整实际结果，返回独立QA逐项复验。保留本报告、所有历史失败、原源08和新QA测试身份；不预报修复通过、新PR、merge或发布。
