# 同轮宿主 Connector 回执核验

来源：平台 previous_results，seq=5，node=tests；kind=command，exit_code=0，摘要“命令已结束 · 退出码 0”。这是本 Run 宿主固定 `make verify`，不是 QA 自己执行，也不是 Issue #12 的历史绿色。

可见输出首段：

```json
{"files":36,"go":"go1.25.6","head":"066f035bc551b2b4d031166cebdab0849084afd4","scope":"product + tests + build locks + README + acceptance/design; excludes delivery receipts to avoid self-reference","working_tree_sha256":"b33aa9f6f848a8cda1bf5af08ba9c2e540d0aa09f91090930ccafa6d4e636702"}
```

随后实际进入 health-history，输出 baseline `9b6994ed80de2bdba79ff267419150dccaac1360`，archive SHA256 `e8307563c976a77940851da8254d70b8255fa4caa690d898a6452ee15de4eaaf`，server SHA256 `d61e4e747231654aa7895404b94c2539497cef2074186b3bbe05b90a94252383`，final_test SHA256 `15b13fbd3cb3265c8bedb9a34f78980fea195351e458f5a4b83ca675513f7c3e`，子命令为 `go test -json -race -count=1 -run ^TestHealthMethodsHTTP$ ./internal/relay`，exit_code=1。可见 healthy 的拒绝方法返回 200/空 Allow，degraded 返回 503/空 Allow，GET/HEAD 和恢复查询通过。这是预期的旧版红灯，不是当前产品失败。

回执输出在旧版 HTTP 事件中截断，标记 `[output truncated]`。没有提供当前 HTTP、前端、浏览器的原始尾部或截图。本文件是事实摘录与核验，不是完整日志副本，不能据此虚构逐项宿主输出、测试数量或视觉检查结果。

QA 两次实际运行 evidence，与上述 HEAD 和工作树摘要一致；Makefile 与 HEAD 原文相同，SHA256 `048e33cc8936e629dc4421052c825f620984eac2fd622cb7c74b25b37d856c63`。其顺序为 evidence → health-history → format → module → vet → frontend-test → full Go race → build → npm ci → Go browser，串行 fail-fast，无忽略失败或删减。browser main 真实启动 Go 服务、临时数据库及受控测试上游，返回 node journey 退出错误；journey 对 1280/390px 连续执行，失败设 exitCode=1。

因此宿主总退出 0 与同版不可吞错的门禁链共同支持本轮完整固定门禁成功，包括真实网络历史分类、当前健康/版本、race、前端和真实 Go 浏览器。此结论属于 Connector 回执层，不等于 QA 亲自执行了浏览器，不用静态审查或 VM 测试补拼 UI Pass。细粒度尾部证据不足作为明确限制保留。

QA 初次和末次均核对研发执行前的 106 文件哈希清单，全 OK；初次核对发生于宿主执行后，因此也确认宿主未留下历史文件改写。本轮没有保存/恢复旧历史的操作。清单不覆盖忽略目录与本 Run 新证据，也不能证明执行中不存在瞬时写入；源码已移除原历史写入，受控成功/失败场景的哨兵断言补充验证输出边界。
