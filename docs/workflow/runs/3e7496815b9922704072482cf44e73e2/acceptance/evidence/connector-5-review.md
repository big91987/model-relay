# tests Connector 执行5审查摘要（非完整日志副本）

来源：本Run用户上下文tests回执及read_command_output(seq=5)。回执kind=command、exit_code=0、timeout_seconds=900、log.bytes=111871、log.truncated=false。原始命令固定入口make verify；输出首行为go run ./tests/evidence。

版本：Go1.25.6；HEAD cd413b5aa24a372a19c19af999a6ef61b5c41572；files142；working_tree_sha256=8e82aaa123df25a4758a1939fb2d018713831507aee337e18dfa6100c49f619e，与QA source-qa.json完全一致。

分段读取offset0→32768→65536→98304→111871/eof=true（retained truncated=false）；额外offset39500/53000核对安装与前端汇总。平台保留完整日志，工具长输出展示存在省略，QA审查关键段，不宣称本Markdown保存了全部111871字节。当前缺本Run完整宿主日志仓库文件。

已观察关键事实：
- review-history隔离旧版实际HTTP出现零预算与回拨错误，为历史精确红灯证据，非当前产品失败；health-history亦是旧版方法错误重建，不是本次缺包。
- format-check、go mod verify(all modules verified)、go vet依次执行。
- make frontend-test首先npm ci固定参数(added2)，再node --test；193 tests/pass193、fail0、cancelled0、skipped0、todo0。包含两条原locked Playwright route处理回归。
- go test -race -count=1 -v ./...：当前ZeroSoftBudgetActualHTTP、ClockRollbackActualHTTP、ClockQueuedDeadlineActualHTTP、ClockRecoveryActualHTTP、治理/角色/费用/SSE/共享池、schema迁移与进程恢复关键段PASS。internal/relay整体ok448.193s，cmd/model-relay整体ok9.818s。TestConsoleClockRecoveryWorker顶层skip单独保留；父用例通过才证明其私有子进程恢复。
- tests/verification的三入口顺序及两消费者安装失败停止PASS，package ok5.245s。其失败npm是显式夹具，不是真实供应商或真实包源故障。
- build实际go build执行；browser再次npm ci(added2)，再真实Go服务浏览器。schema1/2→3CLI+1280/1440/390迁移页面、runtime version与恢复、两视口路由、三视口secret/unknown/double-submit故障、J01–J07各三视口、十页visual及摘要金额/范围均报告PASS，controlled_http_not_supplier。

缺失条件：日志未记录开始node_modules缺失、NODE_PATH、独立工作树身份、未预先安装。不能仅凭npm ci出现或者后续浏览器成功推定它是要求的干净宿主运行。QA本地随后已有node_modules，与开始是否干净无等价关系。

当前没有新PR/合并/Actions prepare/Owner deploy/healthz正式版本/admin网页回执。本摘要不得用于宣布上线。
