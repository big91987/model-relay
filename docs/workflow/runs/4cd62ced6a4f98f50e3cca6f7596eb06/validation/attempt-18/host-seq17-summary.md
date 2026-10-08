# tests 执行17失败边界

- 命令：`make verify`；Connector exit_code=2，完整日志87698 bytes，truncated=false。
- HEAD：`98511771871cf0951ecef716bbb55deb13e18da5`；119文件，旧候选工作树SHA256 `66b873dcdae343c12265140898fddf3fed9d5c298db1b89b507d6d806d029995`。
- 本轮通过read_command_output读取seq17 offset81500→87698，eof=true、truncated=false。完整日志仍由平台保留；此文件是安全摘要，不伪称完整日志副本。
- 旧候选Go race、build、兼容迁移、正式schema1/2→3三视口、旧UI及路由旅程有局部成功。health-history开头的旧版本预期失败是历史反例，不是本次最终失败点。
- 当前最终失败：console-real-go-ui / key-call-http-1280 / TimeoutError。key_create_status=201，issuance_has_secret=true，issuance_has_id=true，catalog_status=200，members_status=200。page_error_count=0、request_failed_count=0；active_page=playground，editor_open=false，secret_control_count=0，call_key_present=true，call_key_nonempty=false，call_invalid_controls=2。没有call_status。
- 已证：已签发一次秘密，调用等待点的Key控件为空且未观察到调用响应；不能归咎供应商、后台调用超时、沙箱监听或浏览器启动失败。两个invalid控件不用于推断另一控件的具体原因。
- 源码机制：secretDialog显式转交先close、await navigate、设置call-key；editor的close监听却调用全局clearSensitive。若关闭事件在转交写入后处理，将清掉目的页Key。新局部回归按两种顺序执行实际handler，证明旧代码违反所有权契约；这不是宿主native事件顺序的直接trace。
- 修复：关闭事件只擦除editor-fields；invalidate仍负责全局敏感字段。转交仅当前post-navigation epoch和playground可写。无sleep、重试、API预置Key或降低断言。
- 待证：新候选原生close/转交、完整调用链及后续全部旅程。宿主须完整make verify，同候选exit0才能进入独立QA；不能用本文判完整通过。
