# 产品与可信正式部署器的窄升级契约

2026-10-06，设计补充为当时快照；产品版本13已实现并经seq28固定门禁与独立QA report-03通过，正式Actions验收未执行。依据本轮新增用户反馈及只读代码核对补齐 [ADR-001](adr-001-storage-upgrade.md)，不新建Issue、Run或路线。产品范围仍为路线2所需schema1→2；不是路线4全套运维平台。

## 1. 设计时缺口、维护源与责任

设计阶段静态快照（非当前能力声明）：平台仓唯一维护源 `examples/github/model-relay-preview/controller.py` 的 `activate` 当前执行停旧服务→旧二进制backup→切current→启动候选→health失败隔离数据并由旧二进制restore。`prepare` 校验精确main、make verify、二进制摘要和计划；没有candidate upgrade或存储契约协商。平台 `deploy-local.yml` 从可信已安装控制器执行prepare/activate，不是从产品提交读取控制器。与本轮“serve仅接受schema2”合用时，schema1安装直接启动候选无法成功；这是静态不一致推断，不是已发生的正式失败部署。

本次只读源核对指纹：平台HEAD `ae41e3071048b8438d6cc3e2b3702c2be67133df`，上述示例路径工作树无改动；controller.py SHA256 `4f9fd39c30ae73ae82c7f794100baa7751fdabb529a974a2c51c807f66a653eb`。该指纹是被阅读的维护源，不是已安装控制器或正式Deployment版本证明。产品HEAD仍是本轮精确基线。

目标维护者同时负责协调以下两个维护源，产品阶段不得越界修改平台仓：

| 维护源／责任 | 本轮所需变更及受支持路径 |
|---|---|
| 产品仓，研发负责人 | Go版本契约、upgrade、schema校验、backup/restore、CLI真实测试及发布说明；固定make verify不变 |
| 平台仓上述示例，目标维护者 | 修复controller prepare/activate、对应标准安装升级及测试；平台源PR不自动合并。本产品仓不复制controller、不改运行目录副本、不写某Issue临时条件 |
| 可信控制器的标准安装 | 从经维护者审查的同一平台源运行 `examples/github/model-relay-preview/install.py --root <deployment-root> --repository <owner/repo> --port <existing-port>`；保留原设置、端口、主密钥、数据及current。此入口当前原子安装controller且不部署产品；不能用手工cp或编辑已安装controller替代 |
| 正式Workflow安装升级 | 若平台源Workflow确有变化，使用 `examples/github/model-relay-preview/install_workflow.py --project <product-checkout> --upgrade`及原安装manifest漂移校验；不手改 `.github/workflows/deploy-local.yml` 或manifest。若模板无变化，记录unchanged，不为本Issue特制Workflow |
| 正式发布与复验，目标维护者/Connector | 修复和标准安装生效后，通过现有Actions精确main准备与已授权workflow_dispatch部署；核对Deployment、实际healthz/控制台和迁移后的数据。安装控制器不等于产品升级成功 |

以上入口为平台源相对路径，部署根由正式配置提供，不提交本机路径、密钥或归档。设计节点仅只读核对上述源和测试；未运行安装、Actions或修改平台源。研发可先完成产品能力，但正式发布必须等待可信控制器维护源与标准安装路径修复，不能让维护者先手工upgrade数据库再假装原activate验证了升级。

## 2. 产品提供的机器可读版本与执行入口

新增Go只读命令 `model-relay deployment-contract`：不接收安装路径/凭据、不打开数据库、不进行网络调用；stdout只输出单一JSON，非零或不可解析时不协商成功。`binary_version`与既有buildVersion一致；正式候选必须匹配计划SHA。产品已实现此命令并在版本11固定门禁验证，下面保留设计契约示例（不是发布SHA）：

```json
{
  "contract_version": 1,
  "binary_version": "<exact-product-sha>",
  "storage": {
    "init_schema": 2,
    "serve_schemas": [2],
    "upgrade_from": [1],
    "explicit_upgrade": true,
    "backup_schemas": [1, 2],
    "restore_schemas": [1, 2]
  }
}
```

`explicit_upgrade=true`表示既有安装必须在启动之前经候选正式upgrade校验；不能把“serve会拒绝”当升级过程。控制器只允许已知contract_version=1及schema集合，JSON有界读取（最多4KiB）、严格校验类型/字段语义；没有可执行shell片段或任意命令字段。可信控制器使用固定命令名和参数数组，不执行产品JSON提供的命令字符串。prepare把合法契约/摘要及可信控制器契约版本绑定到计划；activate重新核对候选二进制、契约、旧版本和精确main，漂移时停止前拒绝并要求prepare，不静默采用旧计划。

正式迁移入口统一为：

```text
<candidate-binary> upgrade --data-dir <installation> --key-file <existing-key-file> --output <new-migration-backup-file>
```

只有 `--data-dir`，不增加 `--data` 别名；key-file沿用原主密钥，output必填且正常迁移必须新文件。迁移内部在安装锁内验证数据和密钥、备份、事务迁移、校验，再返回。成功stdout为单一JSON，至少包含 `binary_version, result, from_schema, schema`；`result=migrated, from_schema=1, schema=2`或`result=already_current, from_schema=2, schema=2`，退出0。失败非零及安全stderr类别，不打印秘密/路径/原始数据库内容，不返回可被误认成功的JSON；进程退出0但结果缺失/不合法也算部署失败。控制器不直接读取或修改业务表。

重复成功调用同一upgrade：schema2先校验密钥/完整性，再返回already_current，零新备份、零重seal/候选插入/配置写；原output可以已存在但绝不覆盖，output参数仍须给出。提交前失败且schema1与output已经存在：安全拒绝 `backup_destination_exists`、零迁移，不能盲用该旧文件；控制器按失败恢复路径处理，新的显式activation尝试分配新的output。这样重复调用不重复副作用，不把失败伪装成功，也不自动无限尝试。

候选 `init`用于首次空安装，直接建立schema2；有existing data/key不能重复init或自动删除。已有schema2同schema发布仍执行upgrade校验，结果already_current，不必再次迁移；旧服务已停止及恢复快照仍由控制器负责。schema未知、错误key、损坏、升级来源不支持或降级schema都拒绝，不回退到直接serve试运气。

候选healthy `/healthz`保持原version字段并添加 `storage_schema:2`，该字段来自已验证打开的实际meta版本，不由声明常量伪造。正式部署以实际version和storage_schema同时核对候选结果，而非只检查200；这是兼容的新增诊断字段。失败/存储degraded保持503，不能靠静态deployment-contract证明库已升级。

## 3. 正式activate顺序及备份边界

prepare仍在隔离工作树通过产品固定make verify并stamp精确SHA，读取无副作用产品契约；现有服务继续运行。准备通过只代表可进入activate，不表示安装已迁移。

activate在可信部署锁内执行以下受支持流程，所有动作纳入既有异常恢复范围：

1. 复查main/计划/二进制/控制器版本、既有current/deployed一致和旧服务健康，协商候选存储能力；旧版无新命令按§5处理。任何不兼容在停旧之前拒绝。
2. 停旧并确认端口退出；保持current仍指旧二进制。用旧二进制正式 `backup --data-dir <installation> --output <new-rollback-backup>` 形成可信回退快照，校验完成/归档schema、保存摘要。原master.key单独保留不覆盖。备份失败不启动候选、不迁移，恢复旧服务并验证旧version；恢复失败报告而非宣称回滚成功。
3. **切current之前**调用经校验的candidate upgrade，参数见§2。候选独立产生自己的迁移前backup；它与控制器旧二进制backup是两份不同新文件：前者保证CLI独立安全，后者是activate恢复旧版的唯一权威快照。不得给upgrade传已经存在的rollback备份路径，也不跳过产品内部备份校验。只接受退出0、合法结果及期望schema。
4. 成功后切current→启动候选→实际healthz验证精确SHA及schema2→保存成功deployed/Deployment回执。只有这一步完成才可报告已部署；连接核验/供应商生成验收依然独立。新库直接init分支无schema1迁移，但同样验证真实schema2，不把它当旧库升级证据。

迁移、启动、健康验证任一失败：确保候选/升级进程退出且安装锁已释放→保留失败数据到受控failed-data新目录（不能删库掩盖失败）→使用**旧二进制**正式 `restore --input <rollback-backup> --data-dir <empty-original-installation-path> --key-file <existing-key-file>`→恢复current指向旧release→启动旧服务并核对旧version→返回失败。不让旧二进制读取schema2，不用candidate backup/restore替代旧回退语义，不手改meta版本或混并候选写入。

候选升级到2后启动失败也必须恢复1，不能仅切二进制；upgrade在current切换前失败同样纳入restore。整个过程有有限子进程/健康等待预算，并为恢复留出作业时间；超时须先终止并等待子进程/安装锁释放，不能一边升级一边restore。若restore或旧服务健康失败，保留原快照/失败库/版本身份，明确“恢复未完成”，禁止启动可能不兼容的旧服务或写成功deployed记录。异常恢复应由同一可信控制器支持重入，不要求手工修改业务库。

该恢复恢复的是停旧时快照；升级后或健康窗口内的候选新增记录保留在隔离失败数据，不隐式合并回旧schema。沿用ADR的新增密钥/配置/请求及停用状态回退边界；不承诺取消收费或候选写入零丢失。

## 4. 重复activate、崩溃与幂等

保持现有部署锁、不可变SHA/二进制摘要和精确main检查。可信控制器需对该activation持久保存最小安全阶段记录：目标/前版SHA及摘要、rollback归档身份/摘要、阶段 `backup_complete → upgrade_confirmed → candidate_healthy → committed` 或 `recovery_pending → rolled_back`；内容不含密钥、密码、请求正文。归档及阶段记录属于可信部署根，不属于产品数据库/仓库，也不是产品新后台服务。

- 已提交同SHA且实际health/version/schema一致：already-deployed，零停服、backup、upgrade或restore；deployed写入后中断可据健康及阶段记录对账，不新建版本。
- upgrade结果丢失/进程中断：先确认原进程已退出，保留原rollback快照；候选同一upgrade可返回already_current，结合阶段及结果验证后继续。不得重新用旧二进制backup已升级的schema2覆盖回退快照。
- schema1且先前output存在/升级非零：按原快照恢复，明确本次失败；下一次授权activate使用新尝试和新文件，不无限自动重试。
- 回退处理中断：同一阶段记录在数据被隔离、restore到空目标、指针恢复各步重入，保留并验证已经恢复的结果；不重复移动或覆盖未知目录。状态不一致停止并保留现场，不能盲重放restore/迁移或重新init。

这些要求是本轮显式升级的最小崩溃边界，由平台维护源实现和测试；设计时controller尚无此升级阶段记录；后续维护源进展见§6，未经真实旧／新Go联合验证不能描述为已验能力。

## 5. 旧版本无升级/协商命令时的兼容

精确旧main `85f6d17c07c5fddeb2d6a39a5a067769a11bb4cc`不支持upgrade或deployment-contract，但其schema1 backup/restore有效。旧版只负责停旧前运行与停旧后backup、失败后restore/start；绝不调用它的upgrade。升级总由新候选执行，所以不需回补旧二进制。

可信控制器只能将经平台维护源明确识别的既有legacy schema1 release（包括本轮精确基线、校验release二进制摘要）作为旧协议兼容输入；无命令的任意二进制、契约命令异常退出/乱码不能一概当schema1。候选为有契约schema2时按上述upgrade；**明确已知**legacy候选且现有安装也为schema1时保留原无迁移backup→start路径，无需它提供新命令。schema2安装不能发布legacy schema1候选，未知兼容情况停旧前拒绝；不做自动降级或新权限策略。识别旧协议属于可信控制器的通用受支持版本策略，基线SHA用于核验来源和验收，不在产品或平台代码加本Issue/SHA特例执行分支。

## 6. 验收层级、依赖和真实状态

当前产品切片已验收，见[report-02](acceptance/report-02.md)及[逐项矩阵](acceptance/matrix-02.md)。平台源最新已记录身份为294ddb33a18a02d413af9e790e4d5caaf26eca51；受管理Workflow/manifest经原安装器Installed/Unchanged及QA只读比对，见[来源记录](delivery/verification/platform-template-provenance-01.json)。用户提供45项平台回归及控制器安装是分层来源，不是产品节点复跑。下表平台早期d0564d05/38项记录为历史进展，不能代替真实联调或DEP05全部边缘；DEP02–04/06–07仍Not Run，正式发布尚未发生。

| 层级 | 要验证的事实 | 当前状态 |
|---|---|---|
| 产品make verify | Go deployment-contract的无副作用/版本/schema；至少一条精确旧SHA用户页面建上游/模型/key并实际调用的非空安装主旅程，再正式备份/upgrade/restore守恒；API仅补重复/错误/并发边界；普通协议/health新增字段回归 | Pass（仅产品）：版本13 seq28原make verify完整exit0，独立QA report-03 Go；旧UI非空schema1→2及旧backup回退通过，不扩为正式Actions |
| 平台维护源测试 | prepare契约/漂移拒绝、旧协议兼容、backup→candidate upgrade→start顺序、升级/启动/健康/恢复失败、重复和崩溃阶段；真实新旧Go二进制配合控制器，不能仅shell伪命令 | 用户报告固定源d0564d05的38项隔离夹具及复审通过；真实新旧Go联合迁移/恢复 **Not Run**，不据夹具关闭本层 |
| 标准安装升级 | 修复后的平台源经install.py生效；若Workflow变化经install_workflow.py --upgrade；重复安装不改数据/设置、模板漂移拒绝、Controller源版本/摘要可核对 | 用户报告原install.py升级/重复安装核对；配置/deployed/current未变，Workflow模板unchanged；控制器摘要和证据来源见下方引用，未迁移产品 |
| 正式Actions旧数据升级 | 从已有正式schema1非空安装，经精确merge SHA prepare/activate，真实执行候选upgrade；同Deployment实际health/schema/版本与控制台、旧身份和历史守恒 | **Not Run**，CLI/受控本地控制器通过不能替代 |
| 正式失败恢复 | 隔离测试部署根通过相同可信Actions路径注入升级/启动/健康失败，实际旧binary/旧schema/旧数据恢复，真实回执分层 | **Not Run**，不把当前生产根破坏作为默认测试；如需真实环境破坏性测试另获授权 |

本轮正式发布可通过已有授权由维护者推进，但必须先完成跨维护源依赖；产品草稿PR/独立QA与正式发布状态分开。正式Actions未复验前只报告实现或隔离验证结果，不报告“正式迁移已通过/已部署”。如果平台修复尚未落地，报告“发布被可信部署升级能力阻断”，不是让管理员手动升级的操作说明。详细逐AC及新部署矩阵见 [验证交接](routing-validation.md)。

2026-10-06用户补充的固定维护源：[d0564d05e11eec271d7707ee7a444346e7be09ae](https://github.com/big91987/agent_platform/commit/d0564d05e11eec271d7707ee7a444346e7be09ae)，源PR #5仍Draft、未合main；标准安装控制器SHA-256为`fc51d9e7b329f3fdf9bc4cbe46fc0541fb411e096a6fbccf4aecff3a21fd226a`。[Actions 37422891093](https://github.com/big91987/model-relay/actions/runs/37422891093)仅以deploy=false进行prepare，日志`Already deployed 85f6d17c`/`deploy skipped`，未重跑make verify或迁移。证据由用户提供，本产品节点未独立复跑；完整分层来源见 [研发验证记录](delivery/verification/report.md#2026-10-06-平台维护源补充)。因此不能继续把“维护源尚未实现/标准安装尚未实施”当当前事实，也不能将接缝进展提升成“真实Go联合/正式迁移通过”：这些实测仍为Not Run，后续按真实回执推进。
