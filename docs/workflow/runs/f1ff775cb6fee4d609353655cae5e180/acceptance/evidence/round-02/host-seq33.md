# 宿主 seq33 固定门禁审核

来源：平台会话 previous_results seq33，kind=command，**退出0**。不是独立QA亲自在沙箱执行的 `make verify`；沙箱限制沿用首轮报告和研发限制，不扩大权限、不重复已确认受限监听测试。

执行：2026-10-06 12:17:56（+08:00起始health-history时间），当前Go测试包112/113秒量级，两屏截图生成于12:20:40和12:21:25。仓库 `big91987/model-relay`，分支 `workflow/f1ff775cb6fee4d609353655cae5e180`，HEAD `d62cd51e8de2127548e1ac7469d349443d1bfe3d`；56文件工作树摘要 `1b2f7026692d0633f6f050d585cec0bb67571a7ecf411848d8213d2fd609cdc6`。QA独立重算相同，不将旧HEAD当新实现。Go 1.25.6，Node 26.0.0，npm 11.12.1，Playwright锁定1.59.1；宿主未输出Chromium完整版本，不伪造浏览器build。产品实际运行原生前端和短命真实Go服务，版本回归从真实healthz核对dev/原值；不是正式部署候选SHA已上线。

回执中段省略，以下只摘录可见尾部，不声称恢复完整stdout：

```text
PASS
ok model-relay/internal/relay 113.488s
PASS
ok model-relay/tests/health-history 4.944s
make build
go build -trimpath -o bin/model-relay ./cmd/model-relay
make browser
npm ci --ignore-scripts --omit=optional --no-audit --no-fund --cache .cache/npm
go run ./tests/browser
model_diagnostic_execution=98536
PASS runtime version actual health, refresh, failure/reentry, nonblocking and text/layout 1280px
PASS QA recovery, precise older evidence, started/unknown/404/injected503, native pagination 1280px (controlled upstream; injected failures explicitly scoped)
PASS actual Go service browser journey 1280px (controlled upstream, not supplier validation)
PASS runtime version actual health, refresh, failure/reentry, nonblocking and text/layout 390px
PASS QA recovery, precise older evidence, started/unknown/404/injected503, native pagination 390px (controlled upstream; injected failures explicitly scoped)
PASS actual Go service browser journey 390px (controlled upstream, not supplier validation)
```

`health-history`头部旧baseline=9b6994e的exit1是严格隔离重建预期红测，不能解释为当前Go失败；固定入口最终exit0才是结果。旧seq7/11未知/截断及所有失败保留，seq31箭头失败精确行仍在delivery/verification，不用seq33绿色改写。

浏览器源核对：`tests/browser/main.go`创建隔离临时installation/动态端口/受控上游，启动实际Go Handler；`journey.mjs`两屏逐次执行。1280空业务配置全程UI；390承接已持久化状态，新别名/密钥仍网页创建。控制端restart/release-stream/upstream-count/audit-logs只承担生命周期准备、夹具、对账，不用于首配预填。网络/管理503/剪贴板拒绝明确为测试注入，不是真实供应商或OS拒绝证据。管理员页面第二执行者停用模型、52次未知key真实UI请求和并发新行均未SQL/API伪造业务对象。

归档核对见 [archive-checks.log](archive-checks.log)：execution98536每屏71行全部精确目标成功；只归档每屏首行含LF原字节，390px首次targetIndex=1/selectedIndex=1，1280首次targetIndex=0不能单独证明移动；5阶段首末观测保留。全部字符串已检查只含固定stage/focus标签，无别名值、凭据、消息、URL或响应体。行摘要用于本地归档一致性，seq33尾部未给单行SHA，不冒称与宿主回执公布SHA比对。

宿主清密截图 [1280px](host-seq33-console-1280.png)/[390px](host-seq33-console-390.png)由QA目视核对：概览有真实dev、配置状态、总数/启用数、同ID/time/revision/alias/scope与下一步，未见凭据/对话；截图前源明确断言所有敏感字段/输出清空。截图只是收尾页面，不是每个失败/焦点步骤的截图证明。摘要、mtime与该次运行时段吻合。

实际环境恢复：main.go deferred shutdown/close/remove、journey finally解除route/清测试剪贴板/撤销context权限/关闭浏览器；exit0与正常清理相符。QA没有独立OS进程清理审计，不补造。正式供应商、PR/合并/测试部署及部署后体验无本轮回执，Not Run。
