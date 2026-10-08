# QA6 实际证据索引

当前快照：source-qa.json；结论Blocked。当前报告../report-qa-6.md；矩阵../matrix-qa-6.md；旅程../journeys-qa-6.md；问题../defects-qa-6.md；返工../remediation-handoff-qa-6.md。

| 文件 | 来源/结果 | 边界 |
|---|---|---|
| source-qa.json | QA项目GOCACHE实算142件摘要，匹配tests seq5 | 含未提交源码，不是部署制品 |
| environment-qa.log | QA起始日期/HEAD/分支/工具/锁/材料archive摘要，diff_check exit0 | QA开始模块已存在；NODE_PATH unset不能补写为宿主开始证明 |
| material-locks-qa.json | QA校验19件材料，当前lock/安装均1.59.1，Chromium解析1217 | 无材料批准状态升级；路径脱敏，无browser启动成功声明 |
| verify-native-qa.log | QA实际make verify完整输出，exit2，历史HTTP监听受限 | Blocked，不是全量Pass或业务红灯 |
| dependencies-qa.log | QA实际go test -race定向顺序/失败停止，exit0 | 临时隔离dry-run/失败npmfixture，不是干净真实npm+全量绿 |
| route-cleanup-qa.log | QA实际原route模块2/2，exit0 | 当前已有Connector安装依赖，非无模块起点 |
| connector-5-review.md | 同版宿主exit0与关键段审查 | 不是完整日志副本；完整平台seq5可继续读，须补档及干净证明 |
| host-visual-review-projection.json | 现有test-results三个视口review的脱敏白名单投影，保留原文件SHA256 | QA未新跑浏览器，不是正式preview或供应商 |
| host-controlled-overview-1280.png | 现有宿主受控1280工作台实际截图，QA已查看 | schema3测试制品/显式测试时钟，非正式5545截图或全旅程证明 |

研发首次真实红灯及失败证据保持在 ../../verification/index.md 引用的具体log，未覆盖。当前平台完整宿主日志尚未落入仓库，这是已报告交接缺口，不伪造为本地完整文件。
