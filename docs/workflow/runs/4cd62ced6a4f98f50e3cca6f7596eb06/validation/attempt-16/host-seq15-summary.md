# 宿主 seq15 失败恢复

- 入口：tests Connector `make verify`；实际 exit2，非完整通过。
- 版本：HEAD `98511771871cf0951ecef716bbb55deb13e18da5`，119文件，工作树SHA256 `34f3a82f0c7687194a60e0cbce6af66e6badff55b0b50a373a0690764e6e7fa3`，对应attempt14，不是本轮诊断候选。
- 保留日志：本Run平台 `read_command_output seq15`；总86806 bytes、truncated=false。本轮恢复offset80000→86806，eof=true；不是声称重新读取/另存全量日志，完整原日志仍由平台保留。
- 尾段可见：Go relay race包PASS（332.565s）、迁移目标契约及health-history测试PASS、build完成；兼容schema1→2和正式schema1/2→3的1280/1440/390迁移UI通过；旧Go UI/多路由1280/390通过。这些只是旧候选局部证据。
- 新控制台从原model失败推进到`key-1280`，错误TimeoutError/timeout。前序服务核验200，revision存在，success/applicable/版本匹配均true，远端模型精确label匹配1。
- `write_status:201`是save helper最近一次成功状态；粗粒度key阶段并不能证明Key POST完成，可能仍是之前价格发布。原回执没有Key专属状态、具体等待动作或安全DOM边界。
- 阶段覆盖导航、异步catalog/members依赖、字段/模型选择、签发、一次秘密弹层、显式转交、普通调用及同Request ID详情。无法据该回执判断是产品、驱动、时序还是网络；不将宿主失败归类为原生沙箱bind限制。
- seq15开头历史health旧版重建有预期红；最终退出位置是browser新控制台Key超时，不是该历史红。

## 本轮结论

源码对照确认上述多个等待点确实处于同一stage。Key后台直接返回一次性`id/key`，编辑依赖读取真实catalog/members，显式转交经dialog.close和navigate；这些源码事实不能确定此次宿主卡在哪一点，更不能单独证明Chromium close事件的实际顺序。本轮不轮换假设性修复，只将最小诊断加入原固定验证链路。

`key-diagnostic-r1`保留原动作、定位、HTTP成功状态及同ID断言，新增11个Key子阶段、按路由分类的数字HTTP状态、安全DOM投影、pageerror/requestfailed计数。未改Go/UI业务、等待时长、鉴权、依赖、Makefile、数据库或权限；没有API预置Key、额外重试/延迟或削弱原断言。新结果仍须宿主运行完整入口；失败携新增stage/status/booleans/count返回development，选择有证据的修复。
