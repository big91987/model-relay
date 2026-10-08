# 当前验收矩阵（QA6）

版本：HEAD cd413b5aa24a372a19c19af999a6ef61b5c41572，142件源码戳8e82aaa123df25a4758a1939fb2d018713831507aee337e18dfa6100c49f619e。证据类型分开，不拼接不同层Pass。`evidence/connector-5-review.md`引用宿主执行5，不代表QA新跑。

| ID | 基线要求 | 角色/范围 | 起始状态 | 连续用户步骤 | 最终事实 | 证据类型 | 结果 | 证据位置 |
|---|---|---|---|---|---|---|---|---|
| CI01-RED | 旧入口真实缺包红灯 | 安装者/旧Makefile | 无node_modules/NODE_PATH | make frontend-test | 192/191/1 MODULE_NOT_FOUND exit2 | Contract | Pass | ../verification/clean-before.log、frontend-before.log（研发原始证据） |
| CI02-STATIC | 锁定准备先于消费者，原门禁保持 | 本候选 | 未提交五文件范围 | diff审查、对照lock与README | 前置共享、无锁变更、无skip/期待变更 | Static Review | Pass | report-qa-6.md；evidence/material-locks-qa.json |
| CI03-ORDER | 三入口顺序及安装失败停止 | QA隔离临时目录 | 复制Makefile，失败npm显式fixture | Go race顺序与两消费者失败回归 | 无消费越过安装失败 | Contract | Pass | evidence/dependencies-qa.log |
| CI04-ROUTE | 原route-cleanup不能跳过 | QA本地已有依赖 | 非干净环境 | node --test route-cleanup.test.mjs | 两例真实加载锁定core并通过 | Contract | Pass | evidence/route-cleanup-qa.log |
| CI05-HOST | 完整公共固定门禁 | tests Connector | 干净状态未知 | make verify全部原步骤 | exit0，源码戳同版，frontend193/193、Go/build/browser | Contract | Pass | evidence/connector-5-review.md；平台seq5完整保留日志（未本地补档） |
| CI06-CLEAN | 全新独立候选无模块、不手工预装直接全量绿 | 宿主安装者 | 必须无node_modules且NODE_PATH unset | 正式宿主make verify | 本轮起始条件+同版全量绿及安装后锁版本 | Contract | Not Run | defects-qa-6.md E28-01；seq5未记录这些条件 |
| CI07-NATIVE | QA自行固定入口 | 原生sandbox | 已有node_modules | make verify | review-history监听受限，exit2；后续未运行 | Contract | Blocked | evidence/verify-native-qa.log |
| BC01-HTTP | AC09–11零软预算、deadline/UTC/恢复 | 租户/Key、受控Go HTTP | 隔离store/测试时钟 | 原预算与回拨HTTP回归 | 拒绝与dispatch约束、UTC高水位、重启恢复保持 | API-E2E | Pass | seq5 TestConsoleReviewZeroSoftBudgetActualHTTP、ClockRollbackActualHTTP、ClockRecoveryActualHTTP、ClockQueuedDeadlineActualHTTP；connector-5-review.md |
| BC02-WORKER | 顶层私有恢复worker | Go子进程 | 无私有状态环境变量 | 顶层go test枚举 | 顶层skip，不把此例转绿 | Contract | Not Run | seq5；同轮父用例实际子进程见report第5节 |
| MIG01 | AC18 schema1/2→3身份/历史守恒 | 正式CLI、隔离夹具 | 旧schema存量 | 原CLI迁移+三视口控制台 | schema3、身份/Keys/历史保持 | Data Reconciliation | Pass | seq5 production CLI migration、TestConsoleMigrationConservation；connector-5-review.md |
| UI01-HOST | AC01–17 J01–07连续UI回归 | 管理员/租户/开发者/财务 | 全新临时Go安装、受控上游 | 各角色context页面操作，详见旅程表 | 三视口21个journey结果、十页各视口、失败恢复/秘密清理 | UI-E2E | Pass | journeys-qa-6.md；seq5；evidence/host-visual-review-projection.json |
| UI02-QA | 独立QA原生重复实际Go UI | 同上 | sandbox监听不允许 | 固定入口未到browser | 未执行连续UI | UI-E2E | Blocked | evidence/verify-native-qa.log（宿主UI单独替代回归，不改此行） |
| SUP01 | 真实供应商事实 | 外部供应商 | 未提供条件 | 正式凭据真实调用 | 尚无真实响应 | External State | Not Run | 本轮无此证据；受控服务不能替代 |
| PUB01 | 新PR、精确已验合并、正式prepare/Owner preview | 授权Connector | 尚无本候选发布回执 | 新PR→精确合并→prepare→deploy=true target=preview | Deployment/environment_url/healthz.version/schema3/admin一致 | External State | Not Run | 交接发布契约，待干净AC闭合后执行 |
| PUB02-UI | 正式真实admin同版可用、旧数据密码保持 | Owner/存量preview | 不重置旧安装 | 正式页面登录/版本与存量核对 | 同Deployment版本、schema3、身份/历史保持 | UI-E2E | Not Run | 没有正式部署及网页回执，不用受控截图替代 |
