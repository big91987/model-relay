# seq34 固定门禁来源审查（QA35）

日期2026-10-08，Run4cd62ced6a4f98f50e3cca6f7596eb06；这是QA读回的宿主Connector证据，不是QA原生执行make verify。

## 回执与版本

- tests执行34：原完整`make verify` exit0，94335 bytes，归档truncated=false，timeout900s。
- 开头来源：HEAD98511771871cf0951ecef716bbb55deb13e18da5，124文件，Go1.25.6，指纹50fb764aac5443687e376c18906b36b765f02e9b05788fa4d5ac73617cbc98a4。
- QA在当前checkout使用项目缓存实际执行`go run ./tests/evidence`，source-before.json与上述完全一致；source-after.json另存最终指纹。不是仅凭HEAD放行未提交产品文件。
- `read_command_output`主链offset0→32768→65536→94335/eof=true，truncated均false；补读37000、49500、55500以查看工程门禁及新控制台断言。工具呈现长块时有中段省略；本文件不伪造原始全日志，不将未逐行呈现的部分当新独立断言。完整留存仍由seq34归档提供，下游可按同seq/offset读回。

## 实际看到的结果

| 范围 | 宿主结果与边界 |
|---|---|
| health-history | 隔离重建旧基线9b6994ed的预期红灯exit1：21个精确method/status/Allow失败、12个query/recovery通过；历史assessment判定PASS。不是当前产品失败，也不是原开发时成功记录 |
| 工程 | format-check、modules verified、go vet、165 Node全通过/skip0、Go race全包、build通过；darwin链接器LC_DYSYMTAB警告不隐藏，未导致exit失败 |
| 当前console API | GovernanceActualHTTP、ResourceDimensionsActualHTTP、PeopleScopeActualHTTP、PauseUsageActualHTTP、ActualHTTPQuotaBillingAndSSE全部PASS；RealProcessCrashRecovery、迁移、精度、身份、Key collection及核验revision契约PASS |
| 既有真实网络/UI | 正式CLI schema1/2→3迁移，各1280/1440/390；旧native UI和多上游路由、SSE/cancel受控回归保持通过 |
| 新console | console-real-go-ui三视口通过；仅证明该脚本动作，不概括为全部AC |
| continuous | 1280/1440/390分别J01、J02、J03、J04、J05、J07、J06通过，共21个检查点；均controlled_http_not_supplier |

## 产物关联和保存

实际读取并校验`test-results/continuous-1280-96921.json`、`continuous-1440-96937.json`、`continuous-390-96963.json`，各result=passed、7项checkpoint唯一且齐全；文件修改时间位于宿主执行窗口（09:24–09:25 +08）。同PID五类脱敏截图及JSON已逐字节复制到`host-seq34/`，共18件。来源、SHA256、大小及时间在artifact-index.json，收集器没有运行材料包脚本。

JSON本身没有source digest，不宣称签名或独立制品attestation；版本关联依据Connector来源戳、QA同checkout指纹、实际runner与时间/PID关联。仅日志和JSON有passed也不足以覆盖脚本从未执行的失败路径。

## 红转绿与剩余边界

seq32的1280/J06首账目误作累计读值，本轮同原路径已绿；当前断言累计fee21/cost7/Tokens3+4、entries2、原revision0和价格快照不变、差额revision=before+1。Go Governance补强原项不变、delta fee5/cost16和原bill19/60均绿。没有为测试改原账本。

QA审查仍发现：console.mjs:137只做签发丢响应后首次读回成功；没有首次读回失败后的重试，以及轮换丢响应分支。journeys.mjs的J03均收到轮换成功响应；call helper每次等待上一调用结束，不是快速双提交。当前绿灯不能证明这些AC08/AC16路径。

QA没有重新执行已知无法监听的原生make verify、没有提权或用static browser.check代替真实后端；原生限制保留在执行22的qa-make-verify.log。当前独立非监听命令及另一个真实源码处理函数的失败复现分开保存。
