# 宿主 seq19 失败归因边界

- tests 执行19，固定 `make verify` exit2；日志88101字节，Connector声明truncated=false。实际恢复offset82000→88101/eof/truncated=false；完整日志保留于平台seq19，不把这段尾部当作全部日志已读取。
- 候选来源：HEAD `98511771871cf0951ecef716bbb55deb13e18da5`，119文件，工作树SHA256 `b61394208bd1f0c4fd58d466390fbfc794c703c5eafeec47b6fe903ce6bc5bdf`，与attempt18一致。
- 开头旧版本health测试失败属于固定health-history预期red复现；当前Go race包332.449s、migration目标测试、build及随后native旅程继续执行。不得将旧版红当作本次产品失败或据它重写health。
- 尾部确认schema1/2→3生产CLI与1280/1440/390迁移控制台通过；旧控制台和路由1280/390通过，范围均为受控上游，不是供应商或正式部署事实。
- 当前schema3控制台停在`keyboard-1280` AssertionError；editor_open=true，page_error_count=0，Key创建201且秘密/id存在，最后call_status=429为quota阶段预期拒绝，request_detail_status=200。因为执行顺序与原断言，旧Key普通调用200、同ID详情和后续流程已越过，但整套控制台三视口未通过。
- 当前粗阶段包含初次Tab、12次Tab循环、Escape可见性和触发点恢复。失败时editor仍开，不能据此唯一认定焦点逃逸；也不能将call429误判为普通调用失败。无新的原生bind限制证据，宿主失败不是环境免责。
- 源码当前仅依赖原生dialog进行Tab/Escape/焦点恢复；存在行为待证，未增加猜测性焦点trap。进一步源码对照确认keyEditor以Promise.all.then异步打开editor，而旧keyboard.press不等待可见dialog；这是确定的驱动准备边界缺口，不能证明是seq19唯一根因。执行20补dialog可见性等待并给原始断言增加细阶段与只读焦点投影；延迟打开的actual-driver回归先红后绿，但不是Chromium机制证明。完整host复验须据新stage/trace选择进一步修复。若仍失败返回development，不直接交QA。

保留seq17、attempt18及更早失败。seq19部分通过仅绑定旧候选，不能跨版本提升全量AC/H、独立QA或发布状态。
