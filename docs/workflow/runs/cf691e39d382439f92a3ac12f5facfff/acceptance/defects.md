# 缺陷与环境限制

| ID | 等级 | 用户能看到的问题 | 复现步骤 | 影响 | 定位信息 | 当前状态 | 回归资产 |
|---|---|---|---|---|---|---|---|
| QA12-HISTORY-WRITE | P3（沿用原分级） | 每次验证覆盖旧日志，必须人为恢复，历史证据失真 | 原 main 运行固定入口；本轮用 HEAD main overlay 运行新回归在临时 fixture 重现 | 验证工程证据保护；本轮强制 AC | 原 main.go 无条件 WriteFile 后 unchecked Print；修复删历史写入并检查 stdout | Closed（本轮原实现红/修复绿、宿主正常及 QA 异常路径历史全未变） | TestHistoryCommandOutput；106 文件哈希；qa-original-regression.log 与 qa-output-regression.log |

新增产品缺陷 0；未关闭 P0/P1/P2/P3 0。不把沙箱权限失败列为产品缺陷，不静默重分级。此次 stdout 丢失错误由同根回归覆盖，不另造重复主题。

环境限制：QA 实际 make verify 在历史 HTTP 监听处 Blocked，之后当前 HTTP/full race/browser Not Run。同轮宿主同版固定入口 exit 0 为公共门禁依据，原用例结果仍保留。宿主文本截断，无法提供逐项尾部/截图；不是发现新失败，也不能虚构完整日志。

执行工具首次拒绝带递归删除临时目录的 overlay 命令，该命令未运行。随后使用显式删除两份临时文件和空目录的方式执行，得到原实现真实 exit 1；临时 fixture 未触碰原仓库历史。拒绝没有被算成红灯结果。
