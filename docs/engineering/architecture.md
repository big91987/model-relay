# 代码职责与依赖

Model Relay 是模块化单体：一个 Go module、一个服务进程和一套 SQLite 数据。按能力划分包，HTTP、业务操作、协议、运行状态和安装维护各有归属。人和 Agent 使用同一份包地图、公开方法与自动检查。

## 后端包地图

| 包 | 负责什么 | 典型入口与边界 |
| --- | --- | --- |
| `cmd/model-relay` | 解析运行、初始化、备份、升级等命令 | 调用 `app.New` 与 `installation`；不实现业务规则 |
| `app` | 组装依赖、HTTP 分流、安全响应头和健康检查 | `New`、`Server.ServeHTTP`；不查询业务表 |
| `httpconsole` | 管理接口的路径、方法、JSON、Cookie、CSRF、分页和错误映射 | `Controller.ServeHTTP`；调用能力服务，不写 SQL、不持有业务数据副本 |
| `httpapi` | 通用 JSON 输入、响应、分页和 CSV 编码 | 严格 JSON 对象解码；不处理租户、Key、价格等业务 |
| `identity` | 登录身份、租户、成员、邀请与接受邀请 | `Authenticate`、`CreateTenant`、`Invite`、`AcceptInvitation`；负责权限、规则、审计和原子提交 |
| `apikey` | 应用 Key、授权模型、诊断成员、密钥代次、轮换、撤销和过期 | `Create`、`Update`、`Rotate`、`Identify`；同一操作中的授权与写入在一个事务内 |
| `catalog` | 上游配置、模型候选、租户模型授权、连接检查 | `CreateUpstream`、`UpdateModel`、`Check`；拥有配置变更的协调锁、版本更新和路由失效 |
| `finance` | 额度配置、价格及可信边界、账单、用量补录和结算 | `UpdateLimits`、`PublishPrices`、`CorrectBill`、`ReconcileUsage`；操作拥有完整事务，金额规则交给 `accounting` |
| `admission` | 调用接纳：重新鉴权、幂等、价格快照、额度检查和资源预留 | `Admit`；上述判断与预留在同一事务中完成，不能分成数次独立提交 |
| `records` | 请求、尝试、派发可能性和路由证据的持久化 | `Start`、`BeginAttempt`、`FinishAttempt`、`Finish`；请求结束与财务结算原子提交 |
| `reporting` | 按当前人员、租户与诊断 Key 范围生成管理读模型 | `Tenants`、`Requests`、`Usage`、`Overview`、`Catalog`；一次读取使用同一事务快照，保留未知值和原始证据 |
| `relay` | 一次推理请求及多次尝试的生命周期 | `Gateway.Generate`、`ListModels`；负责取消、总超时、尝试超时、SSE 空闲超时、输出提交及重试，调用能力服务而不实现 SQL 或供应商协议解析 |
| `upstream` | 上游目标安全、HTTP 传输、请求准备、普通 JSON 与 SSE 协议 | `PrepareChat`、`Client.Send`、`Client.Check`、`NewStream`；不决定重试、结算或下游输出何时提交 |
| `routing` | 权重、优先级、冷却、探索、延迟样本与派发序列 | `Engine.Select/Observe/RecordLatency`；普通输入输出，无 HTTP、SQL 或全局状态 |
| `accounting` | 金额取整、溢出检查、多次尝试预留和多维额度比较 | 纯计算；无 HTTP、SQL 或第三方依赖 |
| `sessions` | 登录会话、失效清理、登录/邀请限速和共享密码计算并发槽 | `Create/Lookup/Delete`、`BeginLogin/BeginInvitation`；不暴露可改写的 map |
| `configstate` | 配置代次、当前成功证据和正在进行的连接检查 | `Changed`、`CompleteRequest`、`BeginCheck/EndCheck`；调用方持同一协调锁，不重复加锁 |
| `installation` | 加密材料、数据库结构、独占锁、有效时钟、恢复、备份和显式升级 | `InitConsole`、`Open`、`Backup`、`Restore`、`UpgradeConsole`；是 DDL 和安装生命周期的唯一归属 |
| `webui` | 内嵌静态资源与允许的页面/模块路径 | `Serve`；不访问业务数据库 |

`access` 提供身份与范围判断；`operations` 管理操作幂等记录和审计；`sqlread` 提供统一查询结果解码；`secure`、`names`、`jsoncodec` 分别承载密钥材料、名称和严格 JSON 的小型公共规则。这些包不调度业务流程。

## 业务分层如何落地

按业务能力组织，不建一个容纳所有业务的全局 `service` 或 `biz` 目录。管理 HTTP 在 `httpconsole`；可独立调用的操作在对应能力包；纯计算在 `accounting`、`routing` 等包；SQLite 实现随拥有事务的能力存放。简单 CRUD 不额外套四层，复杂操作也不能在 HTTP 函数里完成。

例如轮换 Key 由 HTTP 适配器解码 `Transition`，调用 `apikey.Service.Rotate`。服务在同一事务内检查租户权限、暂停状态、Key 修订和已有过渡期，更新密钥代次，写入幂等回执及审计，提交后返回一次性密钥。HTTP 只决定响应码和响应字段。非 HTTP 调用者可以复用相同操作。

租户列表的身份事实与管理摘要由 `reporting.Tenants` 在同一事务快照中组合；身份写操作保持在 `identity`。

价格快照、额度预留和结算跨多张表，仍保持原事务边界。`finance.Settle` 接受已有事务，供请求结束和启动恢复共同调用；它不是独立提交的远程服务。持久化调用顺序不能为了目录分层而打断。

## 网关调用与状态

请求先经 Key 识别与模型授权，再获取配置快照并保存请求。`admission.Admit` 原子接纳后，网关让路由引擎选择候选；`records.BeginAttempt` 保存选择证据并创建尝试。准备上游请求成功后，保存派发可能性，再开始网络调用。

上游解析器返回完整 JSON 或已验证 SSE 事件。网关保留计时、取消、输出提交和是否允许重试的判断。已经提交的流式输出不能切换候选重放。结果证据保存成功后才采用延迟样本；过期配置的结果不能更新当前代次。

会话由 `sessions.Manager` 管理；配置协调由 `configstate.Coordinator` 管理；路由运行状态由 `routing.Engine` 管理。`app.Server` 组装这些依赖，不直接持有这些状态的 map。协调器方法不自行重复加锁；普通 `catalog` 方法拥有锁，`ModelFacts`、`UpstreamFacts`、`ModelSnapshot` 是明确要求调用者已持锁的网关读取入口。

管理报表和账单读取仅使用各自的数据库事务快照，不占用配置写锁。管理修改入口保留原有串行协调范围，配置服务普通方法自行持锁；同一调用链只能选择一种持锁入口，不能重入。旧版总览读取与成功证据失效一起受协调锁保护。

有效时钟由 `installation` 独立维护。HTTP 入口推进后，取得配置协调锁的期限敏感路径还会重新采样，覆盖排队期间经过的期限；配置读锁保护配置快照，时钟事务由安装存储自己的同步规则保护。

通用上游协议与可信预留策略分别归 `upstream` 和 `admission`。前者接受完整的受支持请求协议；后者在有硬预留要求时只允许可证明上界的文本请求子集。工具、多模态和多候选输出的限制服务于资源预留，不应合并成通用协议的字段规则。

仍使用共享配置协调锁和单连接 SQLite。这轮职责拆分不宣称消除锁竞争、支持多进程并发写入或改变产品部署方式。

## 前端功能边界

静态资源位于 `internal/webui/web`。`workspace.js` 管会话、人员/租户范围、页面导航与通用编辑器；业务页面分别在 `features/overview`、`models`、`services`、`tenants`、`keys`、`limits`、`billing`、`reports` 和 `playground` 中。每个功能从自己的 `index.mjs` 创建，拥有该页的业务请求、交互状态和退出清理；`reports` 共同管理调用记录与用量筛选。

受支持的旧版界面同样按 `legacy-*` 功能划分，`app.js` 仅组装功能。`shared` 放通用传输、集合读取、界面值格式化和页面生命周期；不能依赖具体功能。功能不能相互读取内部状态，切换人员/租户或离开页面要取消请求，拒绝过期结果并清除临时密钥。

## 自动约束与验证

`architecture.json` 声明实际包边界、允许的依赖、前端公开入口及场景要求。新增生产包必须声明归属。纯规则和状态包禁止 HTTP/SQL；HTTP 适配与组装包禁止 SQL；能力服务禁止 HTTP。确定规则检查导入方向，模型评审检查同一包内的职责、事务、状态归属及是否存在无效抽象。

`make verify` 保留原有真实 HTTP、SSE、配置变更、资源维度、账单、重启恢复、迁移、浏览器和历史回归门禁。能力包新增隔离 SQLite 与协议测试。规则通过、模型无发现、回归通过及独立产品 QA 是不同结论，不能相互替代。
