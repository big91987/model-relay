## 多上游路由、有限故障切换与可解释控制台

Refs #18。承接公共路线第2项；基线 main `85f6d17c07c5fddeb2d6a39a5a067769a11bb4cc`。Go后端、单实例、单管理员，不提前实现路线3/4的应用治理或分布式设施。

**当前结论：版本13产品切片Go，两项合并前Important/P1已关闭，允许更新原Draft PR #19；不是正式发布或供应商Go。** 已审HEAD `8ae4e5173aa0dc45c28ad61dbc64df20afe229a6` 已由Connector提交推送并创建本PR。本轮版本13修复与新增验收尚未提交推送，本文分支链接在原发布Connector推送并更新PR后反映本版；不能用旧HEAD证明本版远端可读。尚未合并或部署。

### 合并前返工（2026-10-06）

- 两项Important：首SSE data行超过Scanner 256 KiB上限错误切换／冷却，以及DONE写失败／取消纠正后仍留下成功上游。按既有协议和成功事实契约修复，无需求或设计变更。
- 明确将ErrTooLong归类为不可切换的协议错误；提交前后均不生成备用请求、不冷却。非成功终态清空成功归属，DONE写后取消也纠正请求及attempt，详情与概览一致。
- 新增管理/代理Handler＋隔离Transport＋真实SQLite红绿回归，及宿主真实HTTP超长行回归，自动进入原make verify。版本13宿主900秒固定make verify完整exit0，独立QA按原发现层复验Closed；真实HTTP超长行与确定性Writer故障证据分开。
- [返工红绿记录](https://github.com/big91987/model-relay/blob/workflow%2F82227c5986e5f41875895d8016c686a6/docs/workflow/runs/82227c5986e5f41875895d8016c686a6/delivery/verification/premerge-sse-01.md)保留旧版失败与修复过程；[当前QA报告](https://github.com/big91987/model-relay/blob/workflow%2F82227c5986e5f41875895d8016c686a6/docs/workflow/runs/82227c5986e5f41875895d8016c686a6/acceptance/report-03.md)关闭两项Important，原四项P1本版未复发。DONE确定性Writer错误/取消不是精确TCP故障注入；后者仍Not Run，不冒称网络故障复现。平台联合链／正式Actions升级恢复和真实供应商仍Not Run。

### 用户得到的结果

- 从上游页面运营多个独立上游：名称、地址、凭据状态、启停、核验及时间；保存不等于连接成功。核验仅GET模型列表，不生成，严格验证合法列表，10秒覆盖头与body；费用可能性在操作旁提示，已知失败与未核验／未知分开。
- 公共模型可创建、编辑和取消多候选映射并纠正错误输入。统一“大数优先”，同级健康候选平滑加权轮转；受控双上游真实访问计数验证不同远端映射、3:1／1:1、优先级与停用，不仅检查配置JSON。
- 顺序最多尝试3个不同上游，仅提交前白名单暂时错误允许切换；整体300秒预算不重置，输出前单次60秒上限，候选30秒冷却后被动恢复。手动停用和暂时故障区分，已输出或取消后不重放生成；普通、SSE及工具协议持续回归。
- 同一Request ID展示路由决策、逐尝试、安全错误和用量；任一尝试用量未知，不把总消耗伪装成末次或0。凭据、提示词、完成文及原始供应商敏感错误不进入诊断。
- 概览突出未完成操作，完成步骤默认折叠且可回看精确请求；业务导航与账户／恢复分组。1280px及390px各自从页面完成连续原生键盘配置、核验纠错、调用、诊断、停止和重启，不以API／SQL预建业务步骤替代。
- schema2提供显式原子升级、只读deployment-contract与实际health schema；精确旧main真实Go用户页面形成非空配置、密钥和成功／失败历史后，经正式backup／upgrade／restore验证身份及历史守恒。

权威输入：[PRD](https://github.com/big91987/model-relay/blob/workflow%2F82227c5986e5f41875895d8016c686a6/docs/workflow/runs/82227c5986e5f41875895d8016c686a6/product-definition.md)、[路由架构](https://github.com/big91987/model-relay/blob/workflow%2F82227c5986e5f41875895d8016c686a6/docs/workflow/runs/82227c5986e5f41875895d8016c686a6/routing-architecture.md)、[交互契约](https://github.com/big91987/model-relay/blob/workflow%2F82227c5986e5f41875895d8016c686a6/docs/workflow/runs/82227c5986e5f41875895d8016c686a6/routing-interaction.md)。原创实现，仅借鉴公开行为，不复制New API实现、资产、数据或凭据。

### 当前版本13验证与独立QA

| 证据层 | 实际结果与范围 |
|---|---|
| 原固定入口 make verify | 宿主tests Connector seq28，900秒预算，exit0；81349 bytes完整保留、truncated=false。格式、锁模块、vet、前端、Go race/真实HTTP、build及真实Go浏览器链完整通过；历史health负控按原门禁分类保留 |
| 版本绑定 | 验收版本13：86文件工作树摘要e3cde10304e0bbce0dcebf671f9469ab7e5c494d8524cb9a7b7e04228d71d54f，不能仅用旧HEAD证明新实现；Workflow a3d814fa47da0c74e6934ab46138f2b854c55cf1de177f7fba538b4365ab2f45及manifest单独锁定 |
| 独立QA | report-03 Go（仅产品），13组AC产品范围满足；两项SSE Important/P1原层级Closed，原QA-01～04四项P1未复发。QA独立契约race 7顶层exit0／28.424秒、前端119/119 exit0；QA没有重跑完整监听／浏览器入口 |
| SSE新回归 | 实际HTTP超长行首事件前及提交后：primary1/secondary0、协议错误、无冷却；DONE成功对照及写失败/取消带错误/写返回nil时取消：SQLite详情request/attempt与概览一致，非成功归属为空 |
| 真实受控旅程 | 两屏原首配及多候选完整键盘旅程、真实路由计数、核验取消／超时／纠错、完成折叠同ID回看、冷却、SSE取消与工具EOF/idle/cancel零备用，有同版宿主证据 |
| 非空旧数据 | 旧SHA用户UI形成2别名／2密钥／2历史请求，再backup→schema1至2→原key调用→新增UI对象→旧backup回退／旧服务调用；不是空库或手写SQL替代主旅程 |

主证据：[独立验收报告](https://github.com/big91987/model-relay/blob/workflow%2F82227c5986e5f41875895d8016c686a6/docs/workflow/runs/82227c5986e5f41875895d8016c686a6/acceptance/report-03.md)、[逐AC矩阵](https://github.com/big91987/model-relay/blob/workflow%2F82227c5986e5f41875895d8016c686a6/docs/workflow/runs/82227c5986e5f41875895d8016c686a6/acceptance/matrix-03.md)、[seq28回执索引](https://github.com/big91987/model-relay/blob/workflow%2F82227c5986e5f41875895d8016c686a6/docs/workflow/runs/82227c5986e5f41875895d8016c686a6/acceptance/evidence/host-seq28.md)、[版本锁](https://github.com/big91987/model-relay/blob/workflow%2F82227c5986e5f41875895d8016c686a6/docs/workflow/runs/82227c5986e5f41875895d8016c686a6/acceptance/evidence/version-lock-03.json)。回执索引不是完整日志副本；平台保存完整seq28日志。截图及PID选择轨迹仅为辅助证据，不单独证明全部旅程。

[公共交付总览](https://github.com/big91987/model-relay/blob/workflow%2F82227c5986e5f41875895d8016c686a6/docs/workflow/runs/82227c5986e5f41875895d8016c686a6/delivery/README.md)已同步D1–D5 Done、M01–M03 Accepted，限定当前产品切片。原[首次No-Go](https://github.com/big91987/model-relay/blob/workflow%2F82227c5986e5f41875895d8016c686a6/docs/workflow/runs/82227c5986e5f41875895d8016c686a6/acceptance/report-01.md)、沙箱失败、seq9中断及历次宿主返工保留。相对已审HEAD，本轮修复只改proxy.go/routed_proxy.go、新增SSE回归及相关规范；版本13同版门禁和QA重新通过，未沿用版本11。报告收尾仅更新公共文档，证据工具覆盖的文档状态变化会产生新摘要，不能说整树仍等于已验快照；收尾不改行为、测试、锁、规范或模板，差异与实际静态核对见[收尾审查](https://github.com/big91987/model-relay/blob/workflow%2F82227c5986e5f41875895d8016c686a6/docs/workflow/runs/82227c5986e5f41875895d8016c686a6/delivery/verification/report-closeout-02.md)。报告节点未再执行完整make verify。

### 升级、回退与发布限制

- 回退只恢复升级前完整备份，不合并升级后新增配置、密钥、请求或停用状态。CLI备份／升级／回退成功不等于正式可信部署链已执行。见[迁移ADR](https://github.com/big91987/model-relay/blob/workflow%2F82227c5986e5f41875895d8016c686a6/docs/workflow/runs/82227c5986e5f41875895d8016c686a6/adr-001-storage-upgrade.md)。
- **Not Run：** DEP02–04真实可信controller与旧／新Go联合顺序、失败恢复及legacy保护；DEP06正式Actions非空升级；DEP07同可信路径获授权隔离失败恢复。DEP05仅受管理模板Installed／Unchanged及manifest子范围通过，全部安装漂移／失败边缘未独立复跑。
- 平台固定源294ddb33a18a02d413af9e790e4d5caaf26eca51标准模板经原安装器同步；用户提供45项平台回归／复审及控制器安装按来源归档，不冒充本产品节点复跑，不关闭真实联合链。见[部署契约](https://github.com/big91987/model-relay/blob/workflow%2F82227c5986e5f41875895d8016c686a6/docs/workflow/runs/82227c5986e5f41875895d8016c686a6/deployment-upgrade-contract.md)。不复制controller、不手改平台DB、不人工预upgrade替代activate，不自动合并平台源PR。
- **真实供应商 Not Run：** 无凭据和费用授权，实际外部行为全部来自标注受控夹具。核验成功不证明生成／费用可用，取消不承诺撤费，缩短配置的停止测试不冒称完整300秒生产SLA。
- 本PR此前已由Connector创建；当前报告节点未提交推送本轮修复、未创建新PR、合并或部署。原Connector负责更新本Draft PR #19，不另建PR；目标维护者依既有授权核对并推进测试仓合并及正式Actions，保留真实merge SHA／Deployment／healthz版本／schema／非空数据／浏览器一致证据。产品Go不替代尚未完成的发布依赖。
