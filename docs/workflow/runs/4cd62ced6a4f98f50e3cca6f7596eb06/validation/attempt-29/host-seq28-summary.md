# seq28真实宿主失败边界

- 入口：原 `make verify`；exit2；92097 bytes；truncated=false。平台回执为权威日志；本节点通过read_command_output恢复offset86000→92097/eof，未将显示省略部分当作通过证据。
- 来源：HEAD98511771871cf0951ecef716bbb55deb13e18da5；124文件指纹9fd82d0300ce6e3b0879662a8194522f7beb9b1a99e69ad3a32b153c9825a687；Go go1.25.6。不是本次修改后的指纹。
- 已证：TestContinuousUpstreamResponseContract与TestContinuousUpstreamRelayHTTP known/unknown通过；内部relay包PASS；tests/browser包PASS；构建通过。health-history开头旧版预期红不能误认当前产品失败。末段明确保留迁移/旧UI/路由和console-real-go-ui三视口通过。
- continuous-console width1280的J01、J02、J03、J04逐项passed；随后failed：stage=J05-resume/action=authoritative-read/last_status=200/error_kind=assertion/checkpoints=[J01,J02,J03,J04]，initial_counts三个集合均0；call_facts.expected_status=200/dispatch_delta=1。editor_open=false/editor_invalid=0/editor_failure=false/workspace_visible=true/secret_present=false。
- 对照实际driver：该边界在恢复租户并等待活跃行之后，GET第一租户limits成功，下一句比较resumedUsage.usage.fee_micro_usd与pauseUsage.usage.fee_micro_usd；未执行恢复后的两次调用。原回执不含比较金额/期间/预留/调用数，不能区分先前拒绝调用、另一租户调用、暂停恢复或观察时序导致的差异。call_facts来自上一Beta成功调用，不代表失败GET的计费事实。
- 未完成：J05成员降权/退出部分、J07、J06与1440/390连续旅程；不制造后续Pass。原QA22 No-Go不撤销，不归入沙箱监听限制。
- 新增最小诊断：in-process生产handler/SQLite守恒用例通过，但不能证明原Chromium失败唯一机制。宿主应运行同一入口，取得每次暂停拒绝、其他租户成功和恢复后的安全before/after投影后再定位。原200/403/dispatch/费用断言全部保留。
