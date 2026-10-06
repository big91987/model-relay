# ADR-001：版本2配置与尝试存储，显式升级及备份回退

状态：Accepted（依据本轮自动推荐授权，2026-10-06；仅设计决策，不等于迁移已执行）。负责人：本阶段设计／研发维护者；独立用户环境破坏性恢复仍须单独授权，已授权正式activate的自动失败恢复按窄契约处理，不重复普通阶段审批。范围：[PRD](product-definition.md) R-7、US-11/12，数据契约见 [架构§4](routing-architecture.md#4-持久数据契约schema2)。

## 背景与决定

基线schema1只有一个上游及单映射，无法承载独立身份、候选和逐尝试事实。SQLite持久事实源保持单一。新init创建schema2，serve只运行schema2；schema1清楚拒绝并给正式upgrade指引，未知版本拒绝。升级由Go CLI新命令完成，不把普通Open/serve/restore变成隐式不可逆迁移。命令接口：`model-relay upgrade --data-dir <installation> --key-file <key-file> --output <new-backup-file>`，沿用现有安装/密钥/备份输出flag命名，upgrade下output必填；路径由运营者传入，不写本机路径或凭据到文档/仓库。

upgrade先独占安装锁、核对原meta版本与sentinel及全部旧凭据可解密，再通过同锁持有的备份内部函数生成一致schema1归档（不可嵌套重新取锁）。备份目的必须新文件、0600；完成checksum/fsync/关闭确认后才迁移。外部密钥文件不在现有backup归档内，操作说明要求运营者独立保管原密钥，不重新生成密钥替换。错误密钥、占用、备份失败或未知版本均零迁移。已是schema2校验后返回already_current、零配置写/零重复候选；schema版本和产品Version分开。机器可读结果及重复失败规则见 [正式部署窄契约](deployment-upgrade-contract.md)。

## 正式发布路径（本轮补充）

显式upgrade不是维护者在正式发布前手工执行的步骤。产品新增只读 `deployment-contract` 声明schema2启动/升级能力；平台可信controller须在prepare/activate协商并绑定候选版本，停旧→旧binary backup→candidate upgrade→切current/start→核对health版本及实际schema，失败使用旧binary/旧backup恢复。旧binary无upgrade能力不影响它提供backup/restore，新候选承担迁移；未知版本或schema2到legacy降级安全拒绝。

平台维护源 `examples/github/model-relay-preview/controller.py` 当前没有这一步，标准安装/升级也须由目标维护者在平台维护源修复后走正式入口落地。本节点只读核对，不在产品仓复制/修改控制器，不以手动CLI替代正式发布。两份独立备份、重复/崩溃阶段、回退和旧版兼容的完整契约及验收在 [deployment-upgrade-contract.md](deployment-upgrade-contract.md)；当前正式Actions升级/恢复 **Not Run**。产品实现可进入研发，正式发布受此明确依赖约束。

## 原子迁移步骤

1. 同一事务内重建upstream，去id=1约束；原id=1、URL、启停、revision原值保留，名称“原有上游”，execution_revision=1。先解密原 `upstream:{url}`，再以 `upstream:v2:{id}:{url}`重新seal；不持久或打印明文。sentinel仍保留安装原值，admin/hash/salt、app_keys完整不动。
2. models按原alias保留enabled，revision=1；有旧上游时为每个alias创建 `(alias,1)` 候选，remote_model=旧model、priority=0、weight=1、候选enabled=1。上游原enabled仍控制最终可参与性，不用候选启用伪造上游恢复。
3. 无旧上游时不得造假供应商或关联不存在id。将旧远端名放models.legacy_remote_model；零候选，公共名保留，UI显示“尚未关联上游（已保留原模型名）”，管理员从真实编辑入口完成关联后清该字段。空库同一迁移路径，不掩盖非空升级测试。
4. requests只ADD列，旧id、rowid、时间、key/alias/model/revision、状态/错误/status/duration/usage原值不改，evidence_version=1；不反推历史尝试、核验或成功上游。新表建立索引/FK/checks；meta.version最后更新为2，config_revision=1。
5. 事务内检查对象数、映射正确、外键及所有新关联密文可解；提交后检查完整性再退出，任何失败不能启动半迁移serve。进程在提交前中断时由事务恢复到1，提交后中断按完整schema2识别，不重复seal/插入。持有锁直到所有验证和关闭完成。

重建models须先建新表、复制后再建立候选FK，不能因SQLite表重命名误改引用；不重建requests避免分页rowid丢失。事务的foreign_key_check/integrity_check必须失败即拒绝；不在事务外先DROP旧表。失败注入按步骤覆盖，不能只测最终成功。

## 备份、恢复、回退

Go Backup扩展支持schema1/2且清楚记录manifest.Version为实际schema；仍停止服务、独占锁、checkpoint、一致归档与checksum。Restore验证归档版本、条目、checksum、密钥及数据完整性后写入**空目标**，恢复原schema，不调用有副作用的Open来改历史状态或自动升级。共用只校验安装的内部函数；未知版本/损坏/错误key/目标非空拒绝，不覆盖已有安装。

回退路径：停止新服务→可另存schema2新备份保留新增资产→用精确旧基线二进制和schema1升级前归档恢复到新的隔离空目录→启动旧版并用原管理员/应用密钥及公共别名验证。不得让旧二进制直接读schema2，不提供SQL“降级”或自动混并。升级后新上游、配置、密钥和请求不在升级前归档里；停用旧密钥等升级后安全变更也不会被该快照保留，恢复后管理员必须核对这些资产及禁用状态，不能静默承诺不丢数据/不恢复权限。新备份保留不等于旧版可使用。

实际用户环境删除/覆盖/恢复不是本节点授权；测试只在测试拥有临时目录及副本里完成。backup归档、数据库、密钥、口令及完整应用密钥不提交；提交脱敏守恒结果、命令类别、源SHA与退出码。

## 替代与后果

| 方案 | 结论与原因 |
|---|---|
| 继续schema1，加第二份JSON路由 | 不能统一管理事实源/尝试证据，Rejected |
| serve启动自动升级且可直接旧版读回 | 改变启动的风险语义，旧schema无法表达多候选，Rejected |
| 显式事务upgrade＋升级前备份恢复 | 本轮采用；需要短维护窗口，保留身份及历史；不保证保留升级后新增资产的旧版回退 |

退出验证：固定make verify内至少一条主旅程由精确旧SHA隔离真实运行，经旧版用户页面创建上游、多个别名、密钥并完成实际调用形成非空安装，再正式备份/upgrade/新serve/旧备份restore。必须比对管理员认证、应用密钥digest/id、上游/公共名、历史ID/rowid和nullable usage，核对旧凭据新版本仍能向受控上游认证；归档成功但未restore不算回退通过。API可补计数、并发与错误/空库等边界，不替代主UI配置和调用，也不要求每个边缘分支重复完整UI。证据分层见验证交接§3。无法满足守恒或损坏恢复边界则阻止交付，修复维护源，不降低R-7。
