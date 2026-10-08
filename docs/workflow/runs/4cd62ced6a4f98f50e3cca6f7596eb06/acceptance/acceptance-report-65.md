# design-v0.2.0 / budget-clock-r1 独立验收报告

> 日期：2026-10-08，Asia/Shanghai；Run4cd62ced6a4f98f50e3cca6f7596eb06，QA65。
> 仓库：big91987/model-relay；分支workflow/4cd62ced6a4f98f50e3cca6f7596eb06；现有PR27/Issue25。
> HEAD66f1328ada75ea90e5638fbf2bd58663ac86347e + 未提交工作树；140文件SHA256 ca8ec50aa168cf16ac2e24edfe214aa44d06a05846c34deadbc3c7cb6c02904a。
> 环境：宿主实际Go/SQLite schema3/内部身份/Playwright1.59.1/受控HTTP/SSE；Go1.25.6。不是正式部署或真实供应商。
> 模式：全量范围验收；宿主全量固定入口、独立差异与证据审查、QA实际非监听回归。时钟涉及身份/预算/迁移，不能仅沿用QA58短轮绿灯。
> 产品候选判定：**Go with known issues → report**。本Run完整平台/发布退出未完成；不能称最新版本已部署。

## 1. 结论摘要

新宿主seq64原完整make verify exit0、110897bytes/truncated=false，QA实际按游标恢复到eof。当前140件源戳由QA独立Node按路径和实际字节重算，与development63和seq64完全一致；archive及19材料逐项匹配。准确旧PR27的HTTP红灯和当前HTTP绿灯在同轮完整入口里分开，P1-1与P1-2已具原层关闭证据，不凭源码解释关单。

QA另亲自执行六组非监听Go race场景（排队期限、持久时钟子进程/备份恢复、代际精确截止、schema1/2迁移、两scope零软预算、回拨），exit0/24.750s；193Node exit0/无skip。当前三视口J01–J07、十页、管理摘要、故障恢复及正式CLI迁移都由宿主真实Go服务执行，QA审查新证据而不是拿seq57代替。剩余QA58-01 P3及独立self_hosted首次网页分支/供应商/平台/部署未验项按原层保留，不能因门禁绿全部写Pass。

## 2. 权威基线、版本及覆盖

冻结design-v0.2.0 / frozen-for-test-implementation，archive SHA2565f51d35fdc15ed7214749f9413879b4238097f1d6b6bbe2cc0d508d8e103cef7。prepare实际入口为.workflow-input/本Run/该SHA/files/manifest.json；QA先读取manifest，再承接其docs/prd.md、interactions.md、acceptance.md、decisions.md、baseline-review及A原型路径。具体原路径和19件摘要在[evidence/attempt-65/integrity.json](evidence/attempt-65/integrity.json)。本轮重新核对PRD零/未设上限/软预算、期限、未知费用与角色责任，不改冻结资料，不自动运行包内脚本。architecture.md、implementation-contract.md（§4高水位/§5账单）、verification-design.md继续为项目架构事实源，未复制成第二套设计。

范围包括平台管理员、租户管理员、开发者、财务观察者，两租户、多Key/模型，首次空状态到授权/接受/登录/调用；签发轮换撤销、失响应、scope/角色收回；五维双scope准入和并发；未知保守预留、全dispatch价格快照/双费用、核对/结账/追加更正；日志分页筛选CSV/详情、普通/SSE/取消/重启；schema1/2正式CLI升级、原schema3兼容Open、备份恢复。

连续UI契约/当前结果见[acceptance-matrix-65.md](acceptance-matrix-65.md)。J01/J05是真实平台和受邀负责人按职责交接的不同会话，不伪造单人角色切换；J02/J03/J04在稳定租户管理员context，J06在平台管理员context，J07在当前真实身份/scope。只读API用来核对最终事实，测试runner只作首次安装、受控上游和明确时钟准备，不用SQL/API替用户补UI中间步骤。2026-11是显式测试clock的跨月场景，不是本机现实日期/供应商账单。

## 3. 缺陷与质量趋势

| ID | 等级 | 用户问题与原始复现 | 当前事实/状态 | 回归资产 |
|---|---|---|---|---|
| P1-1（QA65预算主题） | P1 | 管理员tenant或Key PUT limits设置budget_micro_usd0/hard=false，再用有授权有价Key chat；原版200/一次上游，而合同要求429/零dispatch/不改用量 | Closed：seq64准确历史HTTP重建红、当前ActualHTTP绿；QA独立handler/SQLite补强。免费请求仍拒绝；NULL无限、正软预算可超限告警不变硬；calls3/fee14/reserved0/pending0 | console_review_regression_test.go；tests/review-history |
| P1-2（QA65时钟主题） | P1 | 1h轮换过渡精确截止旧Key401，回拨截止前而未retire；原版旧值复活200/dispatch1 | Closed：seq64原层HTTP拒绝保持。持久UTC高水位、期限/邀请/月分钟、排队锁后采样、跨进程Open与正式Backup/Restore绿；QA独立race复验 | console_review_regression_test.go；console_clock_recovery_test.go；console_clock_queue_test.go |
| QA47-01/QA47-02 | P1/P2历史 | 管理摘要缺失、确认用量列表/详情歧义 | 既有Closed，本轮新三视口management-summary/J06/visual回归通过；不是继承旧图片作为当前回执 | console-summary.mjs、journeys.mjs、console_summaries_test.go |
| QA58-01 | P3 | 服务/角色/请求详情技术枚举与窄列断词降低扫读效率 | Open，非阻断；当前模型长名字断词仍可见；Owner development后续改善，未发现新核心影响，不擅自升级 | 后续需同页三视口回归 |

较QA58的质量提高但不是零风险：合并前只读复审发现两项原遗漏P1，本轮首次补齐0/soft组合和过期未retire回拨，关闭2个P1、新增自动化及持久时钟恢复/排队边界；本轮未发现新P0/P1。保留development63 red/green测试driver失败和全部历轮失败。QA自己曾选到历史PID，mtime核对后纠正，选择错误单独留存；没有把旧截图/旧门禁放行。

## 4. 证据与独立执行

- 宿主真实完整门禁：tests seq64；恢复/分类/关键结果/skip边界见[evidence/attempt-65/host-seq64-review.md](evidence/attempt-65/host-seq64-review.md)。原日志通过平台seq64可继续读取。
- QA独立重算：[integrity.json](evidence/attempt-65/integrity.json)、可复验脚本check-integrity.mjs；140件源戳不包含本目录新增报告/图片，列出的具体源文件集合为准，不宣称全仓任意文件均在哈希内。
- QA独立实际Go命令：`GOCACHE=.cache/go-build GOTOOLCHAIN=local go test -race -count=1 -v -run '^TestConsole(Review.*InProcess|Clock.*InProcess|GenerationExactDeadline|MigrationConservation)$' ./internal/relay`，exit0，24.750s；[qa-inprocess.log](evidence/attempt-65/qa-inprocess.log)。链接器警告保留，仅临时绝对路径脱敏，不隐去失败。
- QA独立实际`node --test tests/browser/*.test.mjs`，193/193、无skip、exit0；[qa-node.log](evidence/attempt-65/qa-node.log)。VM/单元机制证据不改名UI-E2E。git diff --check exit0。
- 当前UI留存51件：[ui-artifact-index.json](evidence/attempt-65/ui-artifact-index.json)。30页结构化证据与截图SHA核对；QA亲自看三张PNG，非声称人工看完所有图。当前宿主连续JSON为14624/14640/14654，时间与seq64相符，历史96921/96937/96963被排除。

## 5. 未执行、限制与替代

QA原生独立监听/浏览器实操Blocked；遵从已知限制不再次重跑必败make verify、不提权、不用static-root browser.check验证后台。新seq64同版完整宿主真实HTTP/UI-E2E是内部候选门槛的替代执行证据；原生受阻用例仍Blocked，不改Pass。

独立self_hosted首次网页分支Not Run；共用UI来源表单provider完整链、两来源实际HTTP/revision核验覆盖内部能力，与用户“先本候选发布、后补独立self_hosted分支”决定一致。真实供应商/已有端点现货/报价、外部身份和支付Not Run；受控上游只证明本项目语义。H02重复事件、H03实际原生模型身份、H04负向安全恢复、H08平台唯一维护源/正式升级等缺全套事实，仍Not Run。Codex/gpt-6.1-sol是配置/指定声明，非独立可核对的执行凭证。

## 6. 发布及阶段边界

Go with known issues仅针对本140件产品候选，允许report→publish→pr更新同一个PR27；并不代表现有PR27旧HEAD已经含修复。report必须更新本Run根pr.md，列HEAD+工作树源戳/seq64/QA65，随后绑定Connector提交后的新head，不能把66f1328旧已推送版本冒充最新修复。

协调者核验门禁/QA后按用户已授权的测试仓合并及verified-main正式Deploy Model Relay locally执行。preview5545是配置目标，不是发布事实；最终以真实merge SHA、prepare/Actions、Deployment environment_url、/admin/实际页和/healthz运行版本闭合H07。Agent不自行发GitHub动作、不临时启动演示服务、不以8793/旧main代替。report给安全首次初始化/登录指引，未实际部署执行的手册章节标待复核，不写密码/密钥。

支付/退款税务、GPU算力、多模型协议、多区HA/跨实例协调、生产SLO与商用全功能对齐属于后续版本差距；仅按冻结范围实现，不把New API/LiteLLM全部能力纳入当前freeze。

## 7. 环境恢复

本QA仅用独立临时数据/子进程执行非监听测试；testing.TempDir/defer Store关闭/子进程Wait无共享生产变更。宿主runner/context关闭与临时目录清理声明见当前continuous JSON及源审；独立宿主残留进程/端口External State检查Not Run。本节点没有修改业务源码/冻结输入/数据库事实/测试门禁/依赖，没有切分支、提交、推送或部署。

## 8. 判定

当前P0/P1无Open；两项新增合并前P1具准确旧版HTTP红→当前HTTP绿、对应守恒与QA独立补强；全量门禁与当前代码一致，核心UI真实同版宿主连续完成。**Go with known issues，交report**。QA58-01 P3不阻断，缺凭据外部Not Run不阻断当前内部受控候选。完整产品/Harness最终交付仍待H07等外部事实，不能写“完整最终Go”“已部署最新”或“已商用全对齐”。
