# 宿主 seq9 失败边界（保留）

Run 4cd62ced6a4f98f50e3cca6f7596eb06，tests执行9；完整 `make verify` exit2。
Connector保留83947 bytes，truncated=false。development执行10通过 `read_command_output(seq=9,offset=81000)` 读取末段到83947/eof；未读取的中段仅按正式Connector回执描述，不推断额外通过项。完整日志由平台保留，不能用局部显示代替原回执。

被验来源：HEAD 98511771871cf0951ecef716bbb55deb13e18da5；117文件工作树 SHA256 f059c17ab3a1b0edf181d612d2907b0a755ba50605bca737901061abe99739ff。此指纹只对应migration-driver-r2，不是本轮新候选。

## 已证事实

- Go race包及生产build在回执中通过，`TestMigrationUpgradeTargetContracts`两个目标通过；health-history重建旧版的预期失败不等于当前产品失败。
- 浏览器兼容schema1→2输出PASS，包含原UI、映射、Key、历史、受控HTTP、旧二进制回滚；不是正式CLI/供应商/部署事实。
- 正式生产迁移控制台脚本失败于 `keys-1280`，捕获器隐藏原异常；没有该旅程或后续全部native旅程通过证据。这是测试/产品待修失败，不是沙箱bind限制。
- 源码确定性错配：`controlledKeys`调用 `replyCollection` 返回 `{items,total,page,page_size}`，旧脚本却读取 `body.data.length`。新增 `TestConsoleKeyCollectionContract` 使用真实登录/创建/列表Go handler及SQLite复核该包络；Node回归确认旧data包络被拒绝、正确items能完整检查两Key启停及计数。

## 归因边界及下一验证

当前错配是必修驱动缺陷，但seq9粗粒度捕获不能证明它是宿主本次唯一异常：DOM等待、异步scope和实际HTTP结果仍需新宿主回执排除。不轮换其他假设性修复，不改产品/数据库/旧断言。新驱动分 `keys-navigation` / `keys-http` / `keys-contract` / `history`，每视口重置白名单状态/schema/count，错误不输出秘密或原文。

重新通过同固定 `make verify` 验证两迁移目标及正式schema1/2→3和1280/1440/390页面，再执行所有后续旅程；若失败，按新stage及HTTP/schema/count最小事实返回development。完整exit0前不得进入通过裁决或搬用旧指纹局部Pass。
