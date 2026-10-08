# development执行27：controlled-completion-r1

## 来源、范围与公共任务

承接architecture-r1、QA22整改包和tests执行26失败，不重建方案。材料design-v0.2.0 / Frozen for Test Implementation，archive SHA256 `5f51d35fdc15ed7214749f9413879b4238097f1d6b6bbe2cc0d508d8e103cef7`；prepare实际manifest及本轮实际读取PRD§5、交互J01–J07、AC01–AC18/H01–H08路径与摘要见material-source.log。没有执行包内脚本、读取旧工作区或New API实现。用户授权继续有效，无须原型/实施审批。

最小差距：正常与未知用量受控响应应是合法assistant completion，实际遗漏role，被生产协议拒绝502。行为拥有者是tests/browser/journeys.go测试上游，不是validResponse。将原handler提取为continuousUpstream供实际runner和回归共用，只补两条响应role；保持3/4/7和未知用量缺省。新增实际fixture响应回归、生产parser负向回归及真实上游HTTP经schema3 handler/持久化对账回归。journeys.mjs失败时追加同Request ID GET-only安全投影，以区分协议/网络/上游状态/未知收费；原HTTP200、dispatch、J01–J07动作和三视口门槛均保留。

不改变产品HTTP协议、权限、初始化、预算、财务、SQL数据或冻结材料；不预填旅程业务对象、不扩大权限、不推送发布、不调整平台Run。无需求/设计返工或必须用户澄清的边界。局部辅助测试不是宿主/独立QA通过。

按trellis-before-dev运行get_context --mode packages（single-repo/backend/frontend），读取实际frontend生命周期/验证/workspace、backend runtime/console/routing/multitenant规范与共享guides，追踪handler→生产validator→费用明细。prepare initialized=false仍保留；现有Trellis脚本及源码支持规范实际存在、非待填模板，不伪称首次受信初始化通过，本轮不重建bootstrap。trellis-update-spec在workspace-contracts追加签名/协议/安全诊断/错误矩阵/正反例及回归。trellis-check检查本轮文件范围、原门禁、角色契约、已知/未知口径及秘密投影，执行可运行检查，结果见日志。

依据delivery Skill的public-delivery-contract/delivery-review-checklist沿用T001/M01，裁剪为同任务持续整改，不新增组织级对象或重复需求事实源。T001 In Review/M01 Ready for Review；退出Blocked by Validation，不Done/Accepted。当前原生模型没有可验证回执，任务指定Codex/gpt-6.1-sol不能代替平台实际模型核对。

## 真实验证

- expected-red.log：提取原实际handler并补回归、未修role时，`go test -count=1 -run '^TestContinuousUpstreamResponseContract$' ./tests/browser` exit1；known/unknown均失败于assistant role。
- protocol-fixture.log：定向非监听Go race，实际fixture known/unknown、生产缺role/错role拒绝及assistant已知/未知接受、已有新安装空集合和in-process accounting。命令/exit见checks.log；这不是Chromium或网络证据。
- journey-test.log：实际驱动/生产formatter、空态、日志导航及安全失败投影7项通过。只保留枚举/数值/布尔及最多3尝试；不输出ID/上游地址/正文/秘密。
- checks.log：format-check、go mod verify、vet、全部Node、build、browser测试编译、JS语法和diff检查真实命令/结果。
- 新TestContinuousUpstreamRelayHTTP自动进入原make verify的Go测试：真实临时受控上游HTTP、schema3服务handler、管理API和SQLite对账；已知费14/成本7、未知pending1。它不是完整relay网络或浏览器旅程。原生稳定不能监听，未运行此新增监听用例及完整make verify；留给配置宿主tests实际执行，源码不能证明它已Pass。

实际HEAD `98511771871cf0951ecef716bbb55deb13e18da5`；Go go1.25.6；124文件工作树指纹 `9fd82d0300ce6e3b0879662a8194522f7beb9b1a99e69ad3a32b153c9825a687`，完整投影见source-evidence.json。定向race四项顶层测试及子项exit0；全部Node159项exit0；format/module/vet/build/browser compile/JS/diff均exit0。macOS linker LC_DYSYMTAB警告保留，相关测试退出0，不把警告伪作失败或隐藏。原生监听限制未发生相关变化，不重跑必失败入口；首次限制和历次宿主失败保留。

## 宿主复验与退出门槛

tests同source-evidence指纹完整make verify，保留前后来源、完整非截断日志及exitcode；所有原Go/Node/health-history/迁移/HTTP/SSE/取消/未知签发/keyboard断言保持。continuous-console必须1280/1440/390各有J01–J07七项actual passed，保存test-results/continuous-{width}-{pid}.json和before-rotation/key-states/before-correction/after-correction/final脱敏截图到本Run验证目录。

失败保留原回执及call_facts（预期/实际状态、dispatch delta、读取状态、终态/错误白名单、upstream status、bounded attempts），据最小证据先归因再返development，不猜测轮换修复。GET诊断限5秒、失败不替换原status断言。全部通过后独立QA按同材料/指纹逐AC/H、八类商用/J01–J07五层状态复验。QA22-01/02未关闭，原No-Go不撤销；当前完整宿主/QA/发布待验。真实供应商/外部身份支付Not Run；H02/H03模型/H04负向恢复/H07发布/H08平台修复由真实回执判断，不补造。
