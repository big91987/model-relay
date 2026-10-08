# 宿主 seq7 失败归因（摘要，不是完整日志）

- 来源：本Run tests 执行7 Connector 回执；`read_command_output(seq=7,offset=0/65536)`核对首段来源和末段失败。完整保留日志由平台提供，82637 bytes，truncated=false；此文件只保留归因边界，不冒充完整日志或成功证明。
- 固定命令：`make verify`，exit_code=2，timeout_seconds=900。
- 输入HEAD：98511771871cf0951ecef716bbb55deb13e18da5；114文件；Go go1.25.6；工作树SHA256：93b2b67d21d7e1dfbeffd02b2f26a3b414cca7fc0dd781c358c3c91871e44fc2。
- 末段事实：`model-relay/internal/relay` Go race包PASS（325.006s）；health-history测试包PASS；build完成；npm ci完成；`go run ./tests/browser`失败：`migration browser verification failed: candidate upgrade response invalid`，browser子入口exit1，make exit2。旧health方法历史复现的预期红测试不是本次门禁失败点。
- 宿主已经运行真实HTTP回归、旧原生UI种子及升级命令，不能解释成原生沙箱bind限制。后续普通/路由/schema3控制台浏览器旅程未到达，不能计Pass。

## 最小证据和归因

旧驱动`tests/browser/migration.go`构建`./cmd/model-relay`作为candidate，并检查`result.Schema==2`；生产CLI的upgrade实际调用`UpgradeConsole`，目标schema3。后续health和`migration-after.mjs`仍属schema2契约。原报错不含实际schema；本轮`TestMigrationUpgradeTargetContracts`实际构建执行两类命令并复开数据库，确认兼容入口保持2，生产CLI从2迁移到3（attempt09 exit0）。证据确定测试目标绑定不一致，而非网络、浏览器异步或供应商失败。

修复不改产品CLI/数据库/认证/UI。保留schema1→2守恒、旧Key真实调用、旧UI新对象与旧二进制回滚断言，以显式test-only target2执行；追加同源旧UI快照分别schema1/2→正式CLI target3、守恒、旧Key受控HTTP和三视口原生控制台。追加浏览器结果仍待新宿主回执。
