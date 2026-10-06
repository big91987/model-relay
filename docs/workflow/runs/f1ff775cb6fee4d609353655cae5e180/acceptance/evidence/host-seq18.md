# 固定验证宿主 seq18 回执审查

来源：平台本节点输入 `previous_results` 的 tests / seq18，`connector_receipt.kind=command`，`exit_code=0`。这是 **宿主 Connector 实际执行**，不是 QA 本地执行或研发自报。输入正文只有首尾，中段明确省略；本文件保留可见事实摘要，不伪造全量原日志。

命令为原样 `make verify`；源指纹输出：Go `go1.25.6`，HEAD `d62cd51e8de2127548e1ac7469d349443d1bfe3d`，56 文件，摘要 `dc9990248951dbfa54fda356975ffbef0dcfc13fef4add95e4a001eda74d69f4`。QA重新运行指纹一致，见 [baseline.log](baseline.log)。Makefile、Go和npm锁文件对 HEAD 无差异。

可见首段为隔离重建旧版本 health 方法红测，baseline `9b6994ed80de2bdba79ff267419150dccaac1360`；其 `exit_code=1` 是历史分类器预期输入，不是当前产品失败。不能单独把该段红测当本轮 No-Go，也不能把它当当前修复前的原始执行。

可见尾段：当前 `model-relay/internal/relay` race/HTTP PASS，耗时112.539s；health-history 分类器/输出回归 PASS，耗时4.867s；`make build` 完成；`make browser` 锁定依赖安装后执行 `go run ./tests/browser`，最后四行：

```text
PASS runtime version actual health, refresh, failure/reentry, nonblocking and text/layout 1280px
PASS actual Go service browser journey 1280px (controlled upstream, not supplier validation)
PASS runtime version actual health, refresh, failure/reentry, nonblocking and text/layout 390px
PASS actual Go service browser journey 390px (controlled upstream, not supplier validation)
```

中段格式、模块校验、vet、前端测试没有逐行回执；原 Makefile 严格顺序且每步失败即退出，Connector 最终退出0支持全入口完成。QA另外实际运行前端60项及Go三项专项，不能补称自己执行了完整宿主命令。

浏览器入口审查：`tests/browser/main.go` 创建隔离临时数据、初始化管理员、启动真实 relay Handler/SQLite 服务及受控上游；生成版本为 `dev`。上游/模型/应用密钥由 `journey.mjs` 页内表单创建，不用API/SQL预填。控制端只承担流结束/进程重启/日志审计，真实管理API用于结果核对，非替用户中间配置。1280px为空业务起点；390px复用同服务已有2模型/2密钥并另建对象，不称第二次空安装。

`test-results/console-first-call-*.png` 的修改时间为2026-10-06 10:11，内容匹配本旅程两屏终态（2/2模型、2/1密钥和4/4模型、4/2密钥），截图前源码断言所有秘密/消息/输出为空。QA目视无完整凭据/对话，复制到当前证据目录并留SHA256；截图自身没有嵌入版本指纹/执行ID，不宣称另有签名制品溯源。截图证明概览终态和布局，不证明所有表单写前/写后或新增异常路径。

环境：Darwin/arm64，Go1.25.6，Node v26.0.0，npm11.12.1；Playwright锁定1.59.1。Chromium具体build未在回执输出，未补造。仅隔离受控环境，无真实供应商或当前部署制品验收。上次正式部署的main SHA不能代替本次工作树。
