## 多上游路由、有限故障切换与可解释控制台

Refs #18。承接公共路线第2项；基线 main `85f6d17c07c5fddeb2d6a39a5a067769a11bb4cc`。Go后端、单实例、单管理员，不提前实现路线3/4的应用治理或分布式设施。

**当前结论：产品切片独立QA Go，可交付草稿PR；不是正式发布或真实供应商验收Go。** 本文在提交／推送前整理；以下仓库链接指向实际任务分支，须由发布Connector提交并推送本轮产物后生效，不声称当前远端已可访问，不链接旧main冒充本轮证据。

### 用户得到的结果

- 从上游页面运营多个独立上游：名称、地址、凭据状态、启停、核验及时间；保存不等于连接成功。核验仅GET模型列表，不生成，严格验证合法列表，10秒覆盖头与body；费用可能性在操作旁提示，已知失败与未核验／未知分开。
- 公共模型可创建、编辑和取消多候选映射并纠正错误输入。统一“大数优先”，同级健康候选平滑加权轮转；受控双上游真实访问计数验证不同远端映射、3:1／1:1、优先级与停用，不仅检查配置JSON。
- 顺序最多尝试3个不同上游，仅提交前白名单暂时错误允许切换；整体300秒预算不重置，输出前单次60秒上限，候选30秒冷却后被动恢复。手动停用和暂时故障区分，已输出或取消后不重放生成；普通、SSE及工具协议持续回归。
- 同一Request ID展示路由决策、逐尝试、安全错误和用量；任一尝试用量未知，不把总消耗伪装成末次或0。凭据、提示词、完成文及原始供应商敏感错误不进入诊断。
- 概览突出未完成操作，完成步骤默认折叠且可回看精确请求；业务导航与账户／恢复分组。1280px及390px各自从页面完成连续原生键盘配置、核验纠错、调用、诊断、停止和重启，不以API／SQL预建业务步骤替代。
- schema2提供显式原子升级、只读deployment-contract与实际health schema；精确旧main真实Go用户页面形成非空配置、密钥和成功／失败历史后，经正式backup／upgrade／restore验证身份及历史守恒。

权威输入：[PRD](https://github.com/big91987/model-relay/blob/workflow%2F82227c5986e5f41875895d8016c686a6/docs/workflow/runs/82227c5986e5f41875895d8016c686a6/product-definition.md)、[路由架构](https://github.com/big91987/model-relay/blob/workflow%2F82227c5986e5f41875895d8016c686a6/docs/workflow/runs/82227c5986e5f41875895d8016c686a6/routing-architecture.md)、[交互契约](https://github.com/big91987/model-relay/blob/workflow%2F82227c5986e5f41875895d8016c686a6/docs/workflow/runs/82227c5986e5f41875895d8016c686a6/routing-interaction.md)。原创实现，仅借鉴公开行为，不复制New API实现、资产、数据或凭据。

### 验证与独立QA

| 证据层 | 实际结果与范围 |
|---|---|
| 原固定入口 make verify | 宿主tests Connector seq21，900秒预算，exit0；79939 bytes完整保留、truncated=false。格式、锁模块、vet、前端、Go race/真实HTTP、build及真实Go浏览器链完整通过；历史health负控按原门禁分类保留 |
| 版本绑定 | 验收版本11：85文件工作树摘要852f7bfc90323b6253658c8b9c2167fa22c43a5f59389f0bbcb92539865120e7，不能仅用旧HEAD证明新实现；Workflow a3d814fa47da0c74e6934ab46138f2b854c55cf1de177f7fba538b4365ab2f45及manifest单独锁定 |
| 独立QA | report-02 Go（仅产品），13组AC产品范围满足；原QA-01～04四项P1在原层级Closed。QA独立契约race exit0／13.473秒、前端119/119 exit0；QA没有重跑完整监听／浏览器入口 |
| 真实受控旅程 | 两屏原首配及多候选完整键盘旅程、真实路由计数、核验取消／超时／纠错、完成折叠同ID回看、冷却、SSE取消与工具EOF/idle/cancel零备用，有同版宿主证据 |
| 非空旧数据 | 旧SHA用户UI形成2别名／2密钥／2历史请求，再backup→schema1至2→原key调用→新增UI对象→旧backup回退／旧服务调用；不是空库或手写SQL替代主旅程 |

主证据：[独立验收报告](https://github.com/big91987/model-relay/blob/workflow%2F82227c5986e5f41875895d8016c686a6/docs/workflow/runs/82227c5986e5f41875895d8016c686a6/acceptance/report-02.md)、[逐AC矩阵](https://github.com/big91987/model-relay/blob/workflow%2F82227c5986e5f41875895d8016c686a6/docs/workflow/runs/82227c5986e5f41875895d8016c686a6/acceptance/matrix-02.md)、[seq21回执索引](https://github.com/big91987/model-relay/blob/workflow%2F82227c5986e5f41875895d8016c686a6/docs/workflow/runs/82227c5986e5f41875895d8016c686a6/acceptance/evidence/host-seq21.md)、[版本锁](https://github.com/big91987/model-relay/blob/workflow%2F82227c5986e5f41875895d8016c686a6/docs/workflow/runs/82227c5986e5f41875895d8016c686a6/acceptance/evidence/version-lock-02.json)。回执索引不是完整日志副本；平台保存完整seq21日志。截图及PID选择轨迹仅为辅助证据，不单独证明全部旅程。

[公共交付总览](https://github.com/big91987/model-relay/blob/workflow%2F82227c5986e5f41875895d8016c686a6/docs/workflow/runs/82227c5986e5f41875895d8016c686a6/delivery/README.md)已同步D1–D5 Done、M01–M03 Accepted，限定当前产品切片。原[首次No-Go](https://github.com/big91987/model-relay/blob/workflow%2F82227c5986e5f41875895d8016c686a6/docs/workflow/runs/82227c5986e5f41875895d8016c686a6/acceptance/report-01.md)、沙箱失败、seq9中断及历次宿主返工保留。本次只改文档，证据工具覆盖的公共文档变化会产生新摘要，不能说整树仍等于已验快照；代码／测试／锁／规范／模板未变，差异和实际审查见[收尾核对](https://github.com/big91987/model-relay/blob/workflow%2F82227c5986e5f41875895d8016c686a6/docs/workflow/runs/82227c5986e5f41875895d8016c686a6/delivery/verification/report-closeout.md)。本节点未再执行完整make verify。

### 升级、回退与发布限制

- 回退只恢复升级前完整备份，不合并升级后新增配置、密钥、请求或停用状态。CLI备份／升级／回退成功不等于正式可信部署链已执行。见[迁移ADR](https://github.com/big91987/model-relay/blob/workflow%2F82227c5986e5f41875895d8016c686a6/docs/workflow/runs/82227c5986e5f41875895d8016c686a6/adr-001-storage-upgrade.md)。
- **Not Run：** DEP02–04真实可信controller与旧／新Go联合顺序、失败恢复及legacy保护；DEP06正式Actions非空升级；DEP07同可信路径获授权隔离失败恢复。DEP05仅受管理模板Installed／Unchanged及manifest子范围通过，全部安装漂移／失败边缘未独立复跑。
- 平台固定源294ddb33a18a02d413af9e790e4d5caaf26eca51标准模板经原安装器同步；用户提供45项平台回归／复审及控制器安装按来源归档，不冒充本产品节点复跑，不关闭真实联合链。见[部署契约](https://github.com/big91987/model-relay/blob/workflow%2F82227c5986e5f41875895d8016c686a6/docs/workflow/runs/82227c5986e5f41875895d8016c686a6/deployment-upgrade-contract.md)。不复制controller、不手改平台DB、不人工预upgrade替代activate，不自动合并平台源PR。
- **真实供应商 Not Run：** 无凭据和费用授权，实际外部行为全部来自标注受控夹具。核验成功不证明生成／费用可用，取消不承诺撤费，缩短配置的停止测试不冒称完整300秒生产SLA。
- 本节点未提交、推送、创建PR、合并或部署。Connector负责草稿PR；目标维护者依既有授权核对并推进测试仓合并及正式Actions，保留真实merge SHA／Deployment／healthz版本／schema／非空数据／浏览器一致证据。产品Go不替代尚未完成的发布依赖。
