# 研发与独立QA验证交接

状态：产品验证已完成（版本13 seq28及独立QA report-03）；正式可信部署与供应商未执行。2026-10-06。实际逐项结果以[当前QA矩阵](acceptance/matrix-03.md)及[report-03](acceptance/report-03.md)为准，以下保留验证契约而非第二份验收结果。固定入口仍为 `make verify`；不可替换、删除历史重建/格式/vet/modules/前端/Go race/HTTP/build/真服务浏览器各级门禁。引用 [PRD](product-definition.md)、[架构](routing-architecture.md)、[交互](routing-interaction.md)、[ADR](adr-001-storage-upgrade.md)、[正式部署窄契约](deployment-upgrade-contract.md)，本文件不是QA报告。

## 1. 进入证据、实施顺序与退出条件

进入：精确基线已知、G1及G2设计契约就绪、普通技术选择已授权；无真实供应商凭据。按本轮单应用裁剪任务，不补建平台运营体系。

| 顺序／实施范围 | 应改维护源 | 阶段自检及退出（由研发实际执行） |
|---|---|---|
| D1 版本化数据与支持升级 | store.go、backup.go、cmd CLI、healthz、正式操作文档与迁移测试 | 非空旧main守恒、真实Go deployment-contract/upgrade、错误key/备份/中断/版本拒绝、schema1/2真实restore、health实际schema；不能先改路由使旧库失效 |
| D2 权威配置与核验 | server.go/console.go及必要小型Go文件；管理/前端契约 | 两个上游、整体候选保存、revision冲突、核验只GET列表、10秒/取消/迟到/费用及安全 |
| D3 有界数据面 | proxy.go/protocol.go和内聚路由函数 | 开始门、真实分配、白名单/预算/冷却、响应提交和取消、工具透传、逐尝试落盘 |
| D4 同请求诊断与连续UI | console.go、app.js/style.css/index.html、前端专项测试 | 同ID合计/未知、概览下一步和账户分组、错误恢复与两屏键盘连续使用 |
| D5 固定回归与交接 | Makefile已有verify链内扩展tests、Trellis真实契约 | 真实make verify结果、源SHA/工作树标识、失败保留、逐AC证据；独立QA审查后才讨论草稿PR |

设计只更新已有规范的路线2适用范围；实现后研发必须将精确签名、错误矩阵、例子、测试同步回 `.trellis/spec`。不得为了新接口把旧安全/健康/历史证据测试移出verify或将Not Run改PASS。

退出：PRD全部13组P0 AC及既有协议/首配/安全/备份回归完成；make verify真实通过不替代逐AC独立QA。若固定宿主缺依赖/监听或平台回执故障，保留真实失败交维护源修复，再以同入口复验，不能扩大权限或提供自造回执。

新增发布依赖：设计时可信controller无candidate upgrade；后续维护源／标准安装进展已分层归档，真实旧／新Go联合链仍Not Run（见[版本锁](acceptance/evidence/version-lock-02.json)及[研发记录](delivery/verification/report.md)）；目标维护者必须在平台维护源修prepare/activate及标准安装升级，产品仓只实现版本/升级契约和自身测试。按§5分层验收，产品CLI独立验证不是正式发布路径验收；不能要求管理员手工预upgrade后运行旧activate作为替代。产品实现/独立QA可报告其实际证据范围，正式Actions升级/恢复保持Not Run，发布闭环未就绪，直至维护源修复、标准安装和真实复验。

## 2. 验证矩阵

| AC | 固定入口内最低验证内容 | 证据和关键否定断言 |
|---|---|---|
| 01 | 真Go浏览器从空配置UI创建A/B，编辑/取消/启停/重启；名/址/凭据校验 | 每对象回读及UI状态，B不受A失败污染；无API/SQL预建关键配置 |
| 02 | 合法列表、401/429/类型/JSON错误、10秒超时、取消、双点击、配置变化/重启 | 受控模型列表GET和生成POST分开计数，核验零生成；旧结果不是当前成功、不清冷却 |
| 03 | UI候选A/B远端名不同，非法引用/重复、所有priority/weight边界；正常及工具调用 | 实际远端接收model匹配、其他参数不变；编辑失败零部分写，重启回读一致 |
| 04 | priority10:0连续10；1:1连续20；1:3连续40，无其他流量；停三层；并发快照 | 外部接收计数10:0、10:10、10:30及发起序列；停用确认门前零新访问，门后仅已开始允许完成 |
| 05 | 网络/TLS/传输/尝试超时、408/429/500/502/503/504各自切换；其他HTTP和协议各自不切换；2/3/4候选 | 同ID A→B；两个失败只2次、四候选最多3次且不重复上游；耗尽502、新码、无候选503；不只用配置JSON证明 |
| 06 | 冷却前/边界/到期、重复失败、高优先恢复、无流量等待、启停保留、配置旧结果隔离、重启 | 冷却内零A、到期有新请求才A；全部冷却503零访问；列表核验不清健康；跨公共模型不误封 |
| 07 | 普通首结果前取消、已提交普通写失败、SSE首增量/工具增量后断流/缺DONE/idle超时；离页/退出/expiry | 备用外部计数0；活动上游context取消；后续新ID、迟到不能改变新owner；不得靠本地abort状态猜后台停止 |
| 08 | 切换成功/耗尽/零尝试/处理中/取消/中断；逐项全已知/部分/全未知usage、0、溢出；存储Start/attempt/Finish故障 | 精确ID查询与真实响应/接收计数对账；全已知多尝试用量聚合可由Store合同输入验证，不能伪造错误响应usage；历史attempts=null；敏感正文不存在 |
| 09 | 桌面1280与390两次独立完整连续真实UI键盘旅程，至少一次空安装；配置/核验/映射/调用/故障/恢复 | 真实按键顺序、焦点、对象ID、事件、截图，另一屏不得只复用桌面已配置收尾；表格局部滚动和首屏关键动作可用 |
| 10 | 两屏验证就近错误、取消编辑、保存未确认、保存已确认而读失败、冲突、读取未知、复制失败、会话恢复 | 原有成功身份/时间/范围可查，generic记录不误定位；重读不写/不生成；凭据和消息清理无越权 |
| 11 | 精确旧SHA隔离真实运行→旧版用户页面创建上游/多个别名/密钥及启停→UI实际调用形成非空历史→正式backup→upgrade→新serve→既有key真实生成→升级后新增对象→旧归档空目录restore/旧serve | 至少一条完整主UI旅程，对齐US-11；比对身份/历史行/rowid/nullable usage与密钥可用性；API仅补计数/并发/错误边界，不能代替主配置和调用；新增资产不在旧回退里明确标注 |
| 12 | 原HTTP/SSE/工具协议、名单/格式/大小/结束、鉴权来源/登录显隐/限速、SSRF/DNS/重定向、初始化防覆盖、备份、health-history | 原有真实门禁及首次使用不断；假供应商非默认，未授权key零上游；记录fail-closed和日志/截图保密 |
| 13 | 原make verify完整链及独立QA同版本复查 | 命令/真实退出/源版本/证据层/失败历次归档本Run；供应商Not Run，不预报PR/合并/部署 |

时间证据：生产默认300/30/60/60/30（整体/响应头/SSE空闲/输出前尝试/冷却）由Config默认及测试断言核对；受控时钟验证精确边界，缩短配置的真实HTTP等待验证传播和终止。两层都要有，不把毫秒夹具当生产默认，不能用sleep截图替代时间对账。首包合法事件后总时长超过60但未超过整体的流应可继续，证明首提交计时器真正停止；不断注释但无合法事件不能续命。

并发故障专项：快照读取后改映射→当前请求不混入新映射；备用开始门前停用→备用零访问；旧成功晚于新配置故障→新健康不改变；同执行配置旧序号晚到→不覆盖新观察；Start或前次Finish失败→零后续访问；暂停首提交计时器与取消竞态→无额外输出重放；单连接查询/事务无死锁。Go race与HTTP外部接收计数共同验证，不用单纯mock调用次数覆盖网络层。

## 3. 旧基线真实性与夹具权限

主升级旅程必须使用 Git对象 `85f6d17c07c5fddeb2d6a39a5a067769a11bb4cc` 的真实旧Go源码/二进制，在测试临时目录通过正式init并启动旧Go服务。**至少一条真实非空旧安装必须从旧版用户页面**登录、创建上游、多个公共别名、应用密钥并操作启停，通过旧版调用测试页面完成实际生成调用和失败诊断，留下可定位的旧Request ID及用量历史，再以正式命令backup→upgrade→新serve核对→旧归档restore/旧serve回退。该主路径对齐PRD US-11/AC-11，不能以管理API预建上游/模型/密钥或API直接生成替代这些用户步骤（本轮新多上游连续旅程也仍全部UI）。记录旧源SHA、构建结果、安装特征、UI操作轨迹、对象身份及实际受控上游接收；不能用新schema的手写SQL、空库或只改meta.version代替旧版本。

主旧安装覆盖一个加密上游、多个别名、启用/停用模型与key、已知usage成功、未知usage失败、历史拒绝及可定位旧Request ID，关键配置和调用事实来自上述旧版UI旅程。API可只读补充计数、身份/历史对账，并在明确区分的隔离副本补测并发、故障/非法输入、无上游但有模型及空库等边缘分支；这些边缘分支不必各重复完整UI，但报告必须注明“API/边界补测”，不能用其通过关闭主UI升级旅程。测试知道的明文凭据仅临时内存用于旧新HTTP认证，持久对账不保存明文。迁移前先关闭旧服务、完成同版本备份，再升级副本。旧源码不从本机New API或其他仓库取，既有health-history重建不会被此夹具改写。

受控上游仅提供列表、JSON/SSE/工具、可切换HTTP、传输故障、阻塞、计数及时间；不提供产品路由、不预填网关配置。A/B不同远端名明确标“受控测试上游”。第三/第四候选是同样隔离端口的故障边界测试，不在产品安装默认出现。

## 4. 证据保存与结果语义

本Run根：`docs/workflow/runs/82227c5986e5f41875895d8016c686a6`。研发可在该根 `verification/` 归档每次脱敏命令回执，QA在该根 `acceptance/` 保存真实报告及证据；这些目录由执行阶段产生，本节点不补造文件。报告显式列PASS/FAIL/Not Run/环境限制，源SHA和未提交树摘要、命令、退出码、时间、AC关联、真实后端/浏览器/夹具/供应商证据层。

每次失败保留，后续成功另归档；截图在秘密清除后截取，工具原始诊断及内容不进入提交；日志/DB对账脱敏或只报守恒/排除结果。数据库、backup、key、node_modules、个人会话和绝对本机路径保持忽略。历史Run报告不覆盖。

设计阶段未运行make verify、真实UI、迁移或供应商联调；静态线框和本节点自审不替代Go服务证据。真正供应商测试必须由管理员提供合法凭据、模型和费用授权后执行，目前Not Run。平台浏览器静态根工具仅适用设计原型，本轮未调用，不能用其结果替代真实后端旅程。

## 5. 可信部署链升级补充矩阵（R-7／AC-11～13）

每项分别报告实际状态，不以后一层未执行掩盖前一层失败，也不以前一层通过代替后一层：

| 项 | 维护源与入口 | 预期／否定断言与证据 | 当前状态 |
|---|---|---|---|
| DEP-01 产品真实CLI | 产品原make verify链内扩展Go CLI/HTTP及真实旧版浏览器主旅程 | `deployment-contract`不打开安装/联网且buildVersion正确；upgrade统一 `--data-dir --key-file --output`；错误 `--data`不作为别名；§3旧UI建配置并实际调用的非空旧库→备份/迁移JSON→healthz实际schema2→回退；重复/故障可边界补测但不替代主UI | Pass（产品DEP01）：版本11 seq21真实旧UI非空迁移／回退及完整固定门禁exit0，独立QA report-02复核；不扩为Actions/controller通过 |
| DEP-02 可信控制器顺序 | 平台 `examples/github/model-relay-preview/tests/controller_test.py`及维护源受支持测试入口，目标维护者执行 | 与真实精确旧/新Go二进制联合，旧backup→候选upgrade→切current/start顺序；receipt/schema/版本解析错误拒绝，不先serve试运气；原私有fetch令牌/精确main/二进制摘要/设置保护不退化 | 38项平台隔离夹具/复审进展见下方补充；真实新旧Go联合验收 **Not Run** |
| DEP-03 失败与重入 | 同平台维护源测试，临时部署根/受控启动与健康故障 | backup失败零迁移；upgrade/提交后启动/health失败均旧binary restore旧schema/身份/history；超时进程停完才恢复；restore失败明确未恢复；重复activate零副作用、崩溃阶段不重新旧backup已升级库 | 38项平台隔离夹具/复审进展见下方补充；真实新旧Go联合验收 **Not Run** |
| DEP-04 legacy兼容 | 产品旧源码真实二进制＋可信控制器维护源测试 | 旧binary无deployment-contract/upgrade仍完成backup/restore；已知legacy schema1→schema1原路径；schema2→legacy和未知/损坏契约停旧前拒绝，不把任意命令非零当旧协议 | 38项平台隔离夹具/复审进展见下方补充；真实新旧Go联合验收 **Not Run** |
| DEP-05 标准安装升级 | 平台install.py；模板变化才install_workflow.py --upgrade，原manifest/设置检查 | 源修复后安装控制器摘要/源版本可核对，重复安装不改数据/key/port/current；必要Workflow升级保留标准漂移校验；不得手工拷贝controller、改manifest或新建Issue特制条件 | 当前仅受管理模板子范围Pass：原install_workflow.py授权Installed/Unchanged、固定294ddb33模板及manifest版本锁一致；用户控制器安装进展分层保留；全部安装错误／漂移边缘Not Run，见[QA矩阵](acceptance/matrix-03.md) |
| DEP-06 正式Actions成功 | 既有正式prepare＋获授权workflow_dispatch/activate | 已有非空schema1→实际候选upgrade→schema2；merge SHA/Actions run/Deployment/controller版本/二进制摘要/实际healthz和控制台一致；旧登录/key/模型/历史及非敏感数据守恒；回执脱敏归档本Run | **Not Run** |
| DEP-07 正式失败恢复 | 同一可信Actions路径的获授权隔离部署根 | 独立失败回执与旧binary/schema/data真实恢复，保留失败库；不得默认破坏当前正式根、改变产品权限、加入产品故障开关或临时Issue分支；未有受支持隔离根及授权则明确未执行 | **Not Run** |

平台层mock顺序/故障单测仅证明控制器逻辑，不能替代真实Go二进制数据升级；真实隔离控制器链仍不能替代DEP-06正式Actions。维护源修复/安装/Actions证据使用各自源SHA和摘要，不能拿本节点只读指纹证明部署根已更新。DEP-06/07不作为草稿PR前必然已发生的事件，交付/发布节点必须保留Not Run并明确发布依赖，直到真实执行后原位追加证据；用户既有正式部署授权持续有效。

2026-10-06平台维护者进展已按用户提供证据记录于 [本Run研发验证记录](delivery/verification/report.md#2026-10-06-平台维护源补充)：固定源`d0564d05e11eec271d7707ee7a444346e7be09ae`，源PR #5仍Draft；38项隔离夹具/复审、原install.py升级/重复核对及模板unchanged。DEP-02～04的真实新旧Go联合迁移/恢复仍Not Run；DEP-05记录标准安装已报告实施、控制器摘要`fc51d9e7b329f3fdf9bc4cbe46fc0541fb411e096a6fbccf4aecff3a21fd226a`，不从该安装核对推断所有错误/漂移边界通过。正式Actions37422891093只是deploy=false prepare识别原85f6d17c（Already deployed/deploy skipped），未执行产品make verify或候选upgrade；DEP-06/07保持Not Run。下游验收以本补充和表中原断言合读，不降低任何退出条件。
