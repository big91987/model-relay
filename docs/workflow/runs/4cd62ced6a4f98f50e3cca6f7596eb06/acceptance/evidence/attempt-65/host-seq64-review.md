# QA65 对宿主固定入口 seq64 的独立审查

## 来源和版本

真实执行者为 tests Connector，不是本原生 QA 进程。命令为原 `make verify`，timeout900s，exit0；归档110897 bytes、truncated=false。QA65实际调用 read_command_output(seq=64) 从0、32768、65536、98304读取到 next_offset110897/eof=true；另从48000补看前端/全仓回归上下文。保留日志仍在平台seq64，无重新运行、拼造或从摘要生成“完整原日志”。工具展示长度限制不改变归档truncated=false；本文件是审查摘要，不是原始日志副本。

日志最前源码收据：HEAD66f1328ada75ea90e5638fbf2bd58663ac86347e + 未提交budget-clock-r1工作树；140文件，Go1.25.6，SHA256 ca8ec50aa168cf16ac2e24edfe214aa44d06a05846c34deadbc3c7cb6c02904a。QA的独立Node路径/字节重算见integrity.json，完全相同。前端与后台来自同一个Go runner源码/编译入口；运行版本continuous-controlled-ui是测试构建标签，不是正式部署commit身份。

## 历史红灯不是当前失败

review-history隔离git archive准确PR27旧HEAD并复制最小回归，旧版实际HTTP进程退出1，明确观察：tenant和Key分别显式budget0/hard=false返回200、dispatch_delta1；旧代际精确截止401后回拨又200、dispatch1。archive SHA2564b9b3e1b583d0ee0f9e17e007085389ac20414dc5b360cdb2416bb0a48e1e3b1，回归SHA25611e3253c5465c2245841bf2cfdcece1eb83170fe45c5a5a10ba1bd7b3c49f3。

这是修复后受控历史HTTP重建，不是原开发时网络回执。分类器不把监听失败/build/race/timeout/缺断言当作预期红灯。旧health-history的POST/PUT方法红灯同样为明确标记的旧版重建，当前HealthMethodsHTTP/Handler绿灯另外记录。两类历史expected failure不能计为当前测试失败，也不能单独证明新实现绿灯。

## 当前绿灯及边界

- 原format-check、go mod verify、vet、193Node、全仓go test -race -count=1 -v、build、browser保留；internal/relay PASS 458.373s，tests/browser PASS10.287s；分类器测试PASS。
- 当前ReviewZeroSoftBudgetActualHTTP tenant/key全部PASS。零软预算429、上游零dispatch、两个scope费用/预留/计数不变；免费请求同样拒绝；NULL可用、正软预算可以超限并告警，三次真实dispatch最终calls3/fee14/reserved0/pending0。
- 当前ReviewClockRollbackActualHTTP PASS；ClockQueuedDeadlineActualHTTP models/invitation PASS；ClockRecoveryActualHTTP PASS；GenerationExactDeadline及Governance、ResourceMatrix、SharedResources均PASS。这里是正式handler/SQLite和真实HTTP，不是前端演示。
- ClockRecoveryWorker顶层SKIP是未给私密子进程状态文件的辅助入口，单独此行Not Run。父ClockRecoveryActualHTTP实际启动子进程，要求专用CLOCK_RECOVERY_PASS，含重启和Backup/Restore再次执行；父例PASS是同轮等价恢复链，不能把顶层skip改Pass。worker中邀请接受调用为inprocess，代际和模型调用为实际HTTP；邀请真实HTTP由父例和排队期限ActualHTTP覆盖。
- 原HTTP/SSE、未知账本、跨tenant/角色、并发/共享池、价格精度、结账更正/幂等及schema1/2→3守恒、失败保护通过。browser含正式CLI迁移两分支和三视口；schema2 compatibility是明确test-only入口，不冒充正式CLI。
- 1280/1440/390分别J01/J02/J03/J04/J05/J07/J06全部passed；console-real-go-ui十页、editor interruption、签发/轮换失响应、读回503/404/200及普通/SSE双提交全部passed。double cases=一次POST/一次admission/一次attempt/一次上游dispatch，输入3输出4；editor service_posts1/canceled_editor_posts0。
- 三视口management-summary都models2/tenants2/services1/keys5、api_values_match/context_actions=true；visual-current-period都requests0→1/trend_days0→1/fee21微美元、旧闭账不变；console-visual每视口十页通过。

## 当前UI证据留存和独立审查范围

当前continuous进程IDs为1280=14624、1440=14640、390=14654。JSON/visual文件修改时间为2026-10-08 22:43:45、22:44:01、22:44:20（Asia/Shanghai），符合seq64 22:31开始、约22:44结束的命令窗口；与每宽度七个checkpoint/当前visual结果及源码检查相符。绑定属于回执与采集时间/文件内容交叉核对，不声称具有签名制品证明。ui-artifact-index.json列出51件实际留存文件和SHA，全部visual截图匹配review声明摘要。

QA发现首次选择的96921/96937/96963是历史PID，立即依据mtime排除；rejected-artifact-selection.json保留选择失误事实和被拒摘要。历史原图仍保留，不计当前证据；当前索引只引用14624/14640/14654。这不是产品故障，也没有使用错误证据作Go结论。

QA本轮亲自查看当前390工作台、1280模型页、1440额度页PNG，并核对30页结构化事实/图片摘要和旅程执行脚本。其余PNG留存可复查，但不声称本轮逐图人工审查；三图Static/Visual Review不替代宿主连续UI-E2E。QA原生独立浏览器实际操作Blocked（既有监听限制），没有再次启动必败make verify或拿静态browser.check冒充后端操作。

全部上游为controlled_http_not_supplier；真实供应商、自托管真实现货、外部身份/支付、正式部署均Not Run。runner拥有隔离临时目录/随机端口，关闭context/服务/Store，当前三份continuous JSON包含isolated_directory_servers_contexts。QA未独立查宿主进程残留，因此清理External State核查Not Run，不把代码清理合同提升为外部实证。
