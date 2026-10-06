# 宿主固定入口回执索引 21

2026-10-06。**这是 Connector 执行的证据索引，不是 QA 自己执行 make verify 的日志，也不是完整日志副本。**

| 字段 | 实际值 |
|---|---|
| 入口 | 原 `make verify`，tests Connector，900秒预算 |
| 结果 | exit_code=0，正常完成 |
| 版本 | HEAD `85f6d17c07c5fddeb2d6a39a5a067769a11bb4cc`；85文件；工作树 `852f7bfc90323b6253658c8b9c2167fa22c43a5f59389f0bbcb92539865120e7` |
| 环境 | 宿主 macOS、Go go1.25.6；锁定 Playwright 1.59.1，测试不使用真实供应商 |
| 完整存档 | 平台本 Run seq21 retained command log，79939 bytes，truncated=false |
| 本轮读取 | read_command_output 的 offsets 0、32768、38500、52000、60000、65536；已取到 next_offset=79939 / eof=true；显示端长消息会省略中段，因此本索引只陈述实际检索到的重点与完整回执的退出状态 |

## 重点结果

- 旧 `9b6994e` health-history 的 HTTP 405 负控仍为预期失败，门禁明确分类；**不是当前版本失败，也不是原开发时成功记录**。当前 HealthMethodsHTTP/Handler 通过。未删除历史红控。
- 原格式、锁模块、vet、前端、Go race/HTTP、build、真实 Go 浏览器链依原顺序完成。当前 relay 包 255.077秒、health-history测试包5.203秒均PASS。
- 原独立 QA 两项 TestQAConnectionCheck* 全绿，含合法空列表、5种格式错误、坏JSON、认证、限流、服务器、网络错误；不是删除负控或放宽预期。
- TestConnectionRealHTTPHeaderAndBodyBudget 的 headers/body 各PASS，整个子例含安装/清理约12.3秒；实际核验时间的断言为9～12秒，**不把包含fixture开销的子例时长误称核验耗时**。
- 新候选整数/引用原子边界、健康代次/旧序号/启停/重启和双候选工具EOF/idle/cancel零备用专项PASS。
- 精确旧main用户UI非空迁移PASS：schema1→2，2个别名、2把密钥、2条历史请求、4次受控POST；旧二进制回退保留升级前身份。不是正式Actions/controller联合链。
- 原控制台1280/390运行版本、恢复/未知/404/注入503、原生分页及连续真实Go旅程PASS，model_diagnostic_execution=85575。
- 多上游两屏旅程PASS，routing_diagnostic_execution=85660：创建/核验/错误恢复、候选键盘选择、真实1:3/1:1/优先级计数、停用、有限切换/未知用量、被动恢复、SSE/取消零重放。成功行是收尾摘要；本轮对照同版driver确认其中也执行了完成步骤0/5折叠、精确成功请求身份回看、收起以及重启2/3待办范围。

## 对应本地辅助产物

归档的 [版本锁](version-lock-02.json) 包含本轮截图/JSONL字节摘要。`routing-selection-85660.jsonl` 有156行且execution全部为85660；它只记录选择轨迹，不能证明所有核验/折叠步骤。截图来自本地宿主执行后的产物，Connector未给截图原始摘要，不单凭截图文件时间证明版本；主要依据是seq21同版回执及driver实际断言。

## 失败保留与未执行

seq19的10快照 `make verify` exit2（完成回看组TimeoutError）仍保留；11修订后的seq21才是本轮有效完整绿灯。seq9预算中断、seq10/12/14及首次QA失败也不覆盖。DEP02～04、DEP06～07、真实供应商继续Not Run；不从模板安装或旧deploy=false prepare推断候选发布完成。
