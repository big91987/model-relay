# 独立 QA 整改交接

结论：No-Go；返回研发，不回需求或设计。预期在 PRD R-1/AC-02 与 routing-architecture §3 的10秒/合法列表/状态合同明确一致，实现未满足；不得按当前代码把规格改成30秒或放宽合法列表。Owner：产品研发；QA负责修复后同版本复验。所有缺陷为本轮新发现，不是之前选择driver故障复发。

## 缺陷表

| ID | 等级 | 用户能看到的问题 | 复现步骤 | 影响 | 定位信息 | 当前状态 | 回归资产 |
|---|---|---|---|---|---|---|---|
| QA-01 | P1 | 不是合法模型列表也显示“核验成功” | 已配置上游显式核验；供应商返回text/html但body含data数组，或缺object:list、缺模型id；结果success且落库 | 核心运营事实错误，AC-02假阳性；成功不等于生成成功也不能掩盖核验本身不合法 | configuration.go checkUpstream只检查data非nil，未检查类型、object及元素id；browser fixture本身缺object:list | Open | TestQAConnectionCheckProtocolAndFailureStates的5个非法列表子例，独立race日志实际失败 |
| QA-02 | P1 | 核验最多10秒的承诺未实现 | 发起显式核验，观察出站context deadline；实际剩余30秒而非≤10秒 | 强制时间界限违约，慢/悬挂供应商让管理员多等；与同页取消缺口联动 | configuration.go使用30*time.Second；不是生成的30秒响应头合同 | Open | TestQAConnectionCheckTenSecondBudget，30s断言失败；目前测的是出站预算，不伪称已等待30秒 |
| QA-03 | P1 | 已知失败被记成未知 | 核验返回坏JSON、429、503或受控网络错误；API/持久记录state=unknown | “未取得事实”与“明确失败”混淆，核验失败不能按契约诊断；401正控确实failed | 默认unknown，仅401/403改failed；超时分支也unknown，PRD明确超时是已知失败 | Open | TestQAConnectionCheckProtocolAndFailureStates的4个失败子例，API与持久化均失败 |
| QA-04 | P1 | 完成首配后仍在“当前下一步”展示整套已完成步骤 | 同版本宿主390px首配成功截图→返回概览；完整1～5列表常驻，源码无已完成折叠区 | 违反当前强制AC-09/R-6明确的完成后折叠回看，不是另提美化建议；长任务占位问题未闭环 | index.html setup-steps，app.js仅completed class；状态字段隐藏不等于整个完成步骤归入折叠区 | Open | 本轮实际console-first-call-390.png＋同版本DOM/JS对照；需增加真两屏完成后展开/收起/焦点回归，当前缺该自动断言 |

独立运行：`GOCACHE="$PWD/.cache/go-build" go test -race -count=1 -v -run '^TestQAConnectionCheck' ./internal/relay`，实际exit1，日志见 [connection-contract-01.log](evidence/connection-contract-01.log)。合法空列表和401正控Pass，9个负向子例Fail，另10秒预算Fail。没有访问用户供应商、读取安装凭据或改真实数据库。Transport夹具只用于这一层归因，不能替代真实UI核验的失败恢复验收。

## 必须补齐的覆盖／界面契约

1. 将受控列表成功夹具修成合法 `application/json`、`object:list`、每项非空字符串id；新增真实Go/浏览器两屏的非法列表、401/429、超时、安全错误、失败后编辑修正再核验。核验API/落库/页面状态/供应商GET计数一致，零生成、不打印原文。不要只把QA测试删掉或把预期改success/unknown。
2. 核验忙时提供真正的“取消核验”及可见反馈、单上游去重、离页/过期abort与迟到隔离。当前源码 `setUpstreamBusy` 禁用 `cancel-upstream`，其handler在operation存在时return；没有取消核验按钮。此项为静态确认的合同缺口，未取得真实浏览器负控，不宣称该用例执行过。补入原宿主旅程，测取消传递给出站以及取消不改手动启停/冷却。时间测试覆盖头与body总预算，取消/无可靠终态仍可未知，不把确定失败混进去。
3. AC-09每屏完整原生键盘多候选编辑/取消/恢复仍缺证据：routing.mjs的edit回调使用fill/check/uncheck，call使用selectOption/fill/setChecked，密钥输入也fill；75584只证明创建候选原生type-ahead。旧journey的单候选键盘通过不能补齐新双候选旅程。保留现有严格label和外部计数断言，用原生Tab/输入/提交跑完两屏，验证focus-visible；不能改DOM值、API预填或把截图当完成。
4. 概览已完成步骤未按R-6/交互§2归入可回看的折叠区域：本轮390px实际截图在“当前下一步”仍展示完整1～5列表；代码仅切completed class。需落实“下一步突出未完成、已完成可展开回看”，保持原请求身份/时间/范围可查，不删除证据入口。补两屏正常/失败/重启的展开、焦点和任务连续性断言；当前截图观察不包装成已通过该AC。
5. 补矩阵标Not Run的强制边界：核验并发/取消/错误；候选整数边界；健康冷却启停不清除、全部冷却零访问、配置改变旧健康失效和较旧成功晚到隔离、重启未知；已提交工具增量/网络/idle停止。优先小型状态/HTTP专项，再必要UI主闭环；不强迫每个API边缘分支从头UI。代码存在不等于测试通过。

## 不得回退与下轮顺序

- 不改变大数优先、SWRR 1:3/1:1、三不同上游、整体300秒/头30秒/尝试输出前60秒/冷却30秒、取消与提交后不重放、未知多尝试用量及旧schema数据守恒。
- 当前门禁08/seq16 exit0保留，新QA测试自然进入既有 `go test ./...`。修复后先让负控绿，增强真HTTP/UI原层级，再由宿主900秒原 `make verify`完整执行、输出新摘要/完整日志；独立QA后续判13组AC。不要以局部绿或seq16替代新版本。
- 研发同步Trellis与验证记录，但不改原设计预期、可信Makefile/历史门禁或旧Run报告。新增宿主失败与修复证据留本Run，不覆盖本报告。
- DEP-02～04真实Go联合链、DEP-06/07正式Actions与真实供应商仍Not Run；平台维护源/安装/发布由维护者承担，产品节点不触发、不人工预upgrade、不改平台数据或自动合源PR。
