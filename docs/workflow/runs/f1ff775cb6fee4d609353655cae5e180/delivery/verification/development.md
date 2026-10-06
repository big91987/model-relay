# 研发验证与项目测试交接

2026-10-06；执行人：本节点研发。本文件保留首轮验证及18AC映射；第一次补充见[supplement-rework.md](supplement-rework.md)，第二次补充后的当前实现、指纹及回执见[expiry-watchdog.md](expiry-watchdog.md)。首轮工具提示、会话分支及悬挂检查的到期边界已在同Run修正。结论：**固定门禁 Blocked（沙箱监听限制），实现 Ready for Review**。不是QA通过、里程碑Accepted、供应商认证、版本发布或产品全量验收。

## 1. 对象与版本

- 权威输入：[PRD](../../product-definition.md)、[架构](../../console-architecture.md)、[交互](../../console-interaction.md)、[G2](../../g2-review.md)。
- 基线/HEAD `d62cd51e8de2127548e1ac7469d349443d1bfe3d`；本节点未提交、分支/推送由Connector负责，HEAD不能代表本轮未提交实现。
- 首轮源码指纹见[source-current.json](source-current.json)，`working_tree_sha256=513c77cf6f24388e188ae2b60e3f7258eb0af8465fcfb7e428b914c2f8be90fc`，覆盖56文件，Go `go1.25.6`。同指纹在首轮最后固定命令[verify-attempt-04.log](verify-attempt-04.log)首部；不是补充后的当前实现，请以下述补充记录和新指纹为准。
- 环境：原生受限沙箱、Node测试和Go/CGO本地编译；未扩大权限、未读取其他项目或真实供应商凭据。测试数据均属隔离临时安装；Go构造SQL计数/终态夹具不是首次UI旅程证据。
- 日志仅将临时/工作站绝对路径脱敏为`<local-path>`、将系统make路径归一为命令名；保留时间、源码标识、断言、退出码和原始失败。不覆盖历史记录。

## 2. 实际命令与回执

| 命令 | 实际结果 | 证据 |
|---|---|---|
| `make verify`，共4次 | 每次退出2；可信历史HTTP重现的子命令退出1，因`listen tcp 127.0.0.1:0: bind: operation not permitted`被严格分类REJECTED；**没有绕过门禁** | [01](verify-attempt-01.log)、[02](verify-attempt-02.log)、[03](verify-attempt-03.log)、[04当前](verify-attempt-04.log) |
| `make format-check vet frontend-test build` | 3次均退出0；最后41项VM测试全部通过，Go格式、vet和构建通过 | [01](checks-attempt-01.log)、[02](checks-attempt-02.log)、[03](checks-attempt-03.log) |
| `go test -race -count=1 -v ./internal/relay -run '<所列专项>'`，首次含ProcessCrashRecovery | 退出1；Console三专项及不监听回归通过，但Crash需要真HTTP监听而失败；保留该失败，不掩盖为绿 | [go-attempt-01.log](go-attempt-01.log) |
| 同上，移除明确需要监听的Crash，仅不监听专项 | 退出0、race通过；Console三专项、InProcess、Health handler/version、安装/版本/协议/路径/usage回归 | [go-attempt-02.log](go-attempt-02.log) |
| `go run ./tests/browser`，补充执行真实原入口 | 退出1，在受控上游监听处同样bind受限；没有启动Chromium旅程，没有截图，**不称浏览器通过** | [browser-attempt-01.log](browser-attempt-01.log) |
| `node --test tests/browser/ui.test.mjs`，开发期6次 | 27→35→35→39→40→41测试通过；最后的固定子目标也再次41通过。VM不证明真实浏览器键盘/AX | [01](ui-attempt-01.log)、[02](ui-attempt-02.log)、[03](ui-attempt-03.log)、[04](ui-attempt-04.log)、[05](ui-attempt-05.log)、[06](ui-attempt-06.log) |
| `go mod verify`；`node --check` app/journey；`git diff --check` | 均退出0；模块all modules verified；JS语法和补丁空白通过 | 本节点工具真实回执；模块/锁文件未改 |
| Trellis packages、DOM ID/本地引用/历史差异检查 | packages返回单仓backend/frontend；实际版本0.6.15；DOM ID唯一且静态element引用均存在；历史指定文件与HEAD无差异 | 本节点工具真实回执及下述规范审查 |

不监听专项的完整选择表达式（继承Makefile GOCACHE及本地Go toolchain，race启用）：

```sh
go test -race -count=1 -v ./internal/relay -run 'Test(ConsoleOverviewAndExactRecords|ConsoleEvidenceConfigurationRace|ConsoleDeniedMetadataAndStorage|InProcessContract|HealthMethodsHandler|HealthVersionContract|InstallationRecovery|InterruptedAndVersion|DestinationAndProtocol|UsageSnapshots|SpecialInstallationPath)$'
```

该选择只用于有限补充反馈；共享入口仍是原样`make verify`，没有从入口删除任何HTTP、历史、race、build、前端或浏览器阶段。最后Go专项后仅修改原生UI、浏览器验证与公共文档，不改变受检Go逻辑；最终格式/vet/build和41VM再次检查当前实现。失败不能仅以重新限定专项转为版本通过。

## 3. 18条AC追踪（研发自审，不代替QA）

| AC | 实现与自动回归入口 | 本节点证据/剩余验证 |
|---|---|---|
| 01 | hash规范化、导航current、六页刷新；journey每页深链/刷新/back/forward | VM未知hash通过；真实浏览器待宿主 |
| 02 | expires_at、到期/恢复检查、当前管理401同步清除、应用401不退出；密码显隐保留 | VM密码、session和迟到401通过；真实重启/过期旅程待宿主 |
| 03 | 响应式导航、局部表格滚动、焦点和原生控件；journey1280/390键盘登录/导航/编辑取消/发送/诊断/复制/取消 | 脚本存在，原生布局/键盘/AX尚未执行 |
| 04 | overview只读事务COUNT/SUM、null/0、独立healthz原值5秒 | 1006模型/密钥与503未知、原版本VM/Go通过；真实HTTP待宿主 |
| 05 | 成功写后配置代次清候选，Finish成功后同代次登记，持久行/启用映射复核 | 全部配置写/迟到/重启/后来取消通过Go race；UI证据展示待宿主 |
| 06 | 上游、别名、签发、普通/SSE、同ID均由journey操作UI，空业务配置无API/SQL预填 | 首次UI闭环未执行；不能以Go夹具替代 |
| 07 | 依赖入口、真空状态、创建折叠/取消保草稿、模型编辑/启停、密钥启停 | VM取消/编辑忙锁通过；原真服务保存/重启/启停扩展待宿主 |
| 08 | 签发只当次返回，默认掩码、显式显隐/复制、关闭/离页清除、issuedEpoch | 迟到签发/复制隔离通过VM；真实URL/存储/日志/截图保护待宿主 |
| 09 | origin/v1、变量先赋值的curl、凭据/模型占位符、复制手动后备 | VM无秘密示例通过；原生键盘复制及零生成检查待宿主 |
| 10 | 选择启用别名，失效旧选择清空；普通只读文本，工具/空内容提示原始诊断 | VM文本/工具/0/null通过；真实普通内容待宿主 |
| 11 | SSE UTF8/CRLF分帧、DONE确认、错误/断流、真实服务显式流结束门闩、同步防重复 | VM跨UTF8/CRLF/缺DONE/error/取消/重复提交通过；真实增量/POST数/上游取消待宿主 |
| 12 | 会话/导航/操作身份、load/find序列、issuedEpoch、每个await/catch/finally检查，离页/退出/过期/取消清除 | VM旧header/JSON/错误/reader/finally不污染新操作通过；原生生命周期待宿主 |
| 13 | X-Request-ID按钮→精确Record，独立于分页，key ID、alias/model、状态、duration、0/null | Go精确查询/身份/race通过，VM精确与分页分离；journey逐字段核对待宿主 |
| 14 | 未知/停用密钥统一401、拒绝Start/Finish、未知归属空；读503/404不造行 | Go过大/未知/停用/16并发/存储拒绝及零上游通过；真HTTP和保密审计待宿主 |
| 15 | 中文code映射、安全通用网络/协议说明、字段焦点、依赖入口；外部协议不变 | VM恢复/错误与原InProcess通过；真HTTP边界/供应商拒绝脱敏待宿主 |
| 16 | 可恢复启停/输入/签发并明确手动新请求，不重放不自动付费 | VM零重复/取消和Go拒绝通过；原浏览器错误修正/启用链路待宿主 |
| 17 | 原位README/roadmap、真实Trellis、公共任务/里程碑、原验证门禁 | 文档与源码审查通过；项目路线2–4未实现、未创建后续Issue |
| 18 | 本Run留失败/版本/命令和范围，next宿主tests后独立QA | 交接就绪；QA/PR/合并/部署/部署后体验均未发生 |

## 4. Trellis check与跨层复核

before-dev已读取当前Run实际需求、架构/交互与全部适用spec index/checklist、runtime-contracts、dom-lifecycle、verification及共享思考指南。虽prepare历史回执initialized=false，当前checkout已存在真实`.trellis/.version=0.6.15`、脚本和有源码依据的规范；packages真实执行成功。无需bootstrap，不伪造初始化动作。

按update-spec新增backend/console-contracts、frontend/console-lifecycle七节可执行契约，更新index及旧dom-lifecycle签名。收尾按trellis-check审查改动、规格、格式/vet/测试/build、跨层和一致性；静态约束符合，但其“全部tests通过”条件因沙箱监听限制未满足，明确交宿主而不是降低门禁。

- 读链：SQL COUNT/Record/启用映射→admin授权/事务→overview或精确ID→textContent；不存在从截断列表推导总数、假的历史行或html注入。
- 写链：UI显式提交→原Origin/CSRF/jsonInput→SQLite成功写→同锁代次清除；生成保留双授权、Start先于至多一次上游POST、Finish先于成功候选。拒绝仅安全元数据，未知不解析正文或猜key归属。
- 生命周期：导航/过期/退出先失效所有权再停止生成/清DOM；旧HTTP401/签发/reader/clipboard/finally不触达新状态。退出keepalive撤销服务端会话，未确认撤销不报确认。
- 复用：沿用原Store、Record、failure协议、passwordVisibility/clearSecrets/loadVersion和可信Makefile；新列清单同时用于列表和精确查询，未新建后端服务、数据库表或依赖。
- 测试/日志：无调试正文、无凭据持久浏览器存储、未加入真实供应商默认值；原密码/版本/健康方法/历史日志保护/HTTP/SSE/安装恢复继续保留。
- 审查范围：无无关重构或平台补丁；本节点未运行发布动作。frontend无TypeScript，不伪称通过不存在的类型检查；原生键盘/截图尚无事实证据。

## 5. 宿主项目测试交接

1. 在同一源码上原样运行`make verify`，保留当前指纹与真实退出码。失败必须存新的本Run证据，不能覆盖上述失败或历史文档。不得绕过历史HTTP分类器。
2. 真Go浏览器原入口已包含空配置首旅程（1280px）及390px完整旅程；业务上游/模型/应用密钥只由UI创建。Go单元SQL夹具、静态原型、API预填均不能代替该证据。
3. 受控上游明确只为测试；正常SSE在首段后等待控制端`/release-stream`，UI先断言增量/接收中和重复submit零新增POST，再释放DONE。取消分支等待实际HTTP context取消。不得将此夹具称真实供应商。
4. 成功旅程仅在所有秘密和输出清除后生成`test-results/console-first-call-1280.png`、`test-results/console-first-call-390.png`。这是忽略的运行目录，**本节点尚未生成**。tests/QA确认无秘密后归档实际截图到本Run根并列出真实文件，不伪造截图或AX回执。
5. 独立QA对照全部18AC、代码/healthz版本、真实失败/用量/同请求以及两屏/键盘；检查基线日志未改。当前仅实现第1切片；2–4已授权持续目标仍在项目roadmap，不提前实现。

公共任务全In Review、里程碑Ready for Review，未标Done/Accepted/Delivered。固定门禁的解除条件是宿主真实执行，不需要用户额外授权监听。供应商联调未测；PR、合并、正式测试部署及部署后体验没有回执。后续目标维护者按既有授权和独立QA结果自主接续，阶段Agent不抢跑发布。
