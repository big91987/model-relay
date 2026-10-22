## 具体工作与结果
本PR交付 Issue #54 的统一显式目标 handoff 验证材料和必要的测试设施修复，不扩展模型网关业务，不恢复 #51。固定 tests9 的原始 make verify 已通过；独立 QA12 结论 PASS_TO_APPROVAL、QA-10-E1 已关闭；approval13 已正式允许形成报告并由固定发布节点创建草稿PR。报告执行14只核对、整理和交接，没有改产品或测试、没有重跑已通过的全门禁。草稿PR、commit/push及远端发布回执由后续发布节点完成，不以本报告或任务结束冒称整链已交付；未授权合并或部署。

## 问题与实际改动
- 历史红测试将 go test -json 的 stdout 与 race/build 诊断 stderr 混合，可能误判业务红：分离输出、原样保留 stderr，明确拒绝 DATA RACE/build failed，并增加回归覆盖。
- 实际HTTP故障转移测试只给15ms冷却、再固定等待25ms，race下可能在备用上游/SQLite/请求详情期间过期：夹具改为1秒窗口，校验真实 selection decision 的 cooldown_until 并等到该截止时间。原SWRR、故障转移、冷却排除/恢复、用量断言保留；不改生产路由。
- 固定tests5在原12分钟runner门禁预算耗尽：runner有界预算最小调整至25分钟并输出预算/增加回归，外层Connector仍1800秒；Makefile、race和门禁顺序不改。明确这是实际预算变更，不包装成“预算从未修改”。此后tests7失败不是预算耗尽，没有再次延预算或热重跑求绿。
- tests7的自托管SSE旅程在HTTP200且UI已输出后，Chromium CDP无法取得 response.text()：仅浏览器测试改用原fetch响应clone观察，验证同Request ID、单次POST/dispatch、finally恢复fetch。未另发请求、未加重试、未伪造响应；原SSE data:/[DONE]、账本与计量断言保留。新增观察器及反向/清理单测。
- 仓库提交范围严格为本报告 files 中QA确认的十项文件（八项测试设施、两项Harness说明/索引）。产品实现、依赖锁、Makefile及原生历史均不修改；候选a033eb20未获取、未比较或复用。源码快照用.txt/patch隔离归档，不把可编译副本加入递归构建输入，不排除正常测试。

## 已验版本、环境与固定测试
- 基线HEAD：868530b3e5c15d65bee0ed4ff6af69e96d428b81；工作树为本PR候选，并非已提交main。验收输入160文件SHA-256：44107d4db03a944c27a7d9b36b03e1802a770b3219c9bae61b19f0d81d4795fd。tests9前后、QA10/12及report14实测相同；该摘要有意排除交付文档以避免自引用，十项待提交文件另按字节/SHA清单锁定。
- tests9实际命令 make verify；2026-10-22 02:42:49.440854150Z–02:59:53.800960021Z，1024.360秒，外层/runner/gate均exit0，非截断日志。Go go1.25.6；Playwright/playwright-core 1.59.1，Chrome for Testing 147.0.7727.15，按锁安装隔离依赖。
- 格式、模块核验、go vet、七包全量race、285项前端单测（0失败）、build通过；migration、legacy-ui、routing、console、full-console五浏览器阶段均完成。实际Go服务和受控HTTP/SSE上游，1280/1440/390三宽度continuous/self-hosted/visual共九份结果passed，不是静态browser.check或真实供应商接入验证。
- 完整平台原生日志 harness-output/qa-12-tests9-platform-export.log：1093069字节，truncated=false，SHA-256 868edba207d2a972ce2c0b67c3ce5ec2485ed49f61334908d6a7cae30bf0e4ca；runner原日志 harness-output/tests-9-verify.log：1062735字节，SHA-256 c27a7fce90c4e6769363f5f5d85ea412de3aa3acc7c8d3b2a8fd7ba9cdfc9c5e。原件/state及原失败日志随平台artifacts保留，不把提示中省略的日志当作完整证据。

## 独立QA、确认与交接
- QA10独立实际HTTP、review-history/verify race、观察器7项反向/清理用例共六份命令回执exit0，但因平台证据不足真实No-Go。原报告 qa-10-report-preserved.md 和原失败均保留，未被QA12覆盖。
- development11只补正式平台导出和说明，未改160项验收输入、未重跑全测。冻结development唯一出口tests会真实再跑make verify，故依User Input11625由Owner正式停止并人工回退QA12；这是人工流程恢复，不是成功handoff，不修改冻结图或历史拒绝。
- QA12独立重取tests9完整日志，逐字节核对167项上轮索引、110项原证据、八份源码快照和development8历史120项artifacts/12906095字节，审计exit0；未新跑全浏览器或定点测试。Builder正式API选段经User Input11016/11328归因补证，关闭QA-10-E1。QA文件原文由固定发布节点复制到docs/harness/run-197ab169bc1cad067466300038ab6fac，本报告body原文亦由该节点复制；harness-output本身不进入files。
- development8原生handoff事件31753/call_ZzFP9Esbj348tnZaiHMlN12l核对完整target=tests、message、结构化context、artifacts及accepted=true。development4因待投递User Input被拒后处理消息再交接，development11的拒绝与人工恢复仍按真实记录保留，不重写为成功。
- approval13的已保存人工确认允许report/固定发布生成草稿PR，不批准合并/部署。report14已实际执行本地版本/完整性核对、重放QA12审计（不是新的独立QA），未再次运行make verify。当前正在准备target=publish的真实handoff；是否受理由随后的工具原生回执证明，不提前写accepted。报告节点的原生身份/模型/会话回执仍需平台后续核验，不虚构conversation/thread/model事件。
- docs/harness中development.md保留development11“待人工回退”历史时点，不改动已QA锁定原文；当前状态以QA12、approval13及本节更新为准，不能据旧pending否定较新的正式结论，也不能删除旧红。

## 时间、返工及可证实阻塞
固定测试失败后的研发返工3次：tests3于01:30:18.711Z–01:34:18.413Z（239.702秒，runner1/gate2，冷却夹具）；tests5于01:48:05.455Z–02:00:20.907Z（735.452秒，runner1/gate-1，预算耗尽）；tests7于02:11:56.635Z–02:27:34.994Z（938.360秒，runner1/gate2，SSE取体）；每次均保留原红并在实质修复后才进入新全门禁。tests9通过耗时见上。QA补证返工1次，不是又一次产品返工。development8可证实驻留下界619.764秒，development11首段补证下界168.256秒；QA10下界439.047秒，QA12从03:35:45Z采样至03:41:13.031Z报告下界328.031秒。其余研发轮次采样和失败启动详见development.md/manifest及原命令回执，精确入场和Owner等待/approval驻留无独立计时则不补造；report14从03:49:57Z首次明确采样的驻留下界记录于report-14-audit.json。候选不可读、平台补证等待、待投递输入拒绝及人工恢复均保留来源，不能把未知时间算成执行时长。

## 限制与真实Issue关联
- 关联 big91987/model-relay#54（https://github.com/big91987/model-relay/issues/54），Run 197ab169bc1cad067466300038ab6fac，冻结工作流7；Issue编号和入站Actions39703355037/39703356253来自正式User Input9738，未独立获取Actions回执。Refs #54；#48/#49仅历史参考，不作为本轮通过证据；#51是独立冻结产品任务，不被本PR替代或恢复。
- 未调用真实供应商、未读取或复用用户凭据、未改用户数据/配置。无真实供应商模型鉴定；无逐图人工视觉验收，无新增独立QA全浏览器复跑。
- development/qa/report配置codex/gpt-6.1-sol及不同会话/thread身份、共享工作区来自Builder正式API选段与本地一致性核对，不冒称QA自行直取API、完整HTTP封套或穷尽内部隔离审计。thread.started没有持久化model字段；配置+实际工具执行+外部提供的安装源运行时Model匹配校验链仅支持操作层验收，不伪造缺字段或供应商证明。报告实际执行/受理后续以平台原生回执核对。
- 当前尚无本PR真实远端编号/URL或发布成功回执。本节点不commit/push、不合并、不部署；固定发布节点须使用本版十项具体文件和实际远端回执，保持draft，不能把本报告当作已发布。
