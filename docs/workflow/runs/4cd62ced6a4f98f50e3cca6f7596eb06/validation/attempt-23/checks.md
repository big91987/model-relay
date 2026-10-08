# 执行23研发整改与宿主复验契约

版本continuous-journeys-r1；HEAD `98511771871cf0951ecef716bbb55deb13e18da5`。未提交、未切分支、未推送。当前工作树指纹由同目录source-evidence.json提供；不能沿用seq21的119文件指纹。QA22 No-Go仍有效，两个缺陷均为待复验候选。

## 输入与恢复

Run `4cd62ced6a4f98f50e3cca6f7596eb06`；原任务T001/M01，不新建另一事实源。冻结版本design-v0.2.0，archive SHA256 `5f51d35fdc15ed7214749f9413879b4238097f1d6b6bbe2cc0d508d8e103cef7`；实际manifest：`.workflow-input/4cd62ced6a4f98f50e3cca6f7596eb06/5f51d35fdc15ed7214749f9413879b4238097f1d6b6bbe2cc0d508d8e103cef7/files/manifest.json`。本轮实际引用原根docs/prd.md、docs/interactions.md、docs/acceptance.md及既有architecture.md、implementation-contract.md、verification-design.md；逐文件及archive核对见material-source.json。

恢复acceptance/remediation-handoff.md、user-journeys.md、acceptance-report.md及evidence/host-seq21-review.md。seq21原门禁exit0与旧指纹有效，但QA22-01连续旅程覆盖P1、QA22-02状态P2未关闭。没有恢复旧工作区。Trellis源码规范和脚本实际存在；`get_context.py --mode packages`成功，单仓backend/frontend。prepare initialized:false不代表缺初始化；规范已是源码约束而非空模板，不重新bootstrap。已读before-dev/check及交付Skill要求references，补充可执行契约入workspace-contracts.md。

## 修改边界及事实

- `workspace.js`：Key启用/停用/过期/过渡/撤销/未知展示；新增应用、明确授权、真实已用/预留/预算列。后端仍决定权限和费用。原formatter确实只读revoked，不把静态修复写成原生通过。
- 无租户时needTenant原先append空态而留下加载占位；本轮源码审查确认会导致移除成员后refresh一直显示加载，改为replaceChildren并补实际函数回归。不是声称已在本地浏览器复现。
- 请求详情增加当前租户/Request ID日志入口，不用全局日志替代scope。
- 新Go runner只用production InitConsole/Open、真实监听及临时目录；每个视口重新init。初始只有init自带的空legacy租户；不隐藏这项事实，也不改DB删它。测试专用时钟/dispatch计数由另一个随机端口和随机capability保护，仅推进现有Config.Clock；没有产品测试接口或SQL改账。认证会话仍用真实wall clock，不声称推进了人员会话时间。
- 保留原所有HTTP/SSE/取消/未知签发/keyboard/迁移/健康红灯/Node/Go门禁，新continuous suite追加到原browser runner，固定入口仍为make verify。非监听检查不能证明新增浏览器运行成功。

## 连续旅程契约与当前状态

统一Actor由实际邀请/接受/登录产生；UI步骤禁止API补中间对象，API只最终读回，受控时钟/dispatch count不产生业务对象。统一Cleanup关闭浏览器contexts、Go/上游/控制server、Store，删除临时目录；不得接触用户配置。每行在1280/1440/390实际Go/SQLite执行；本节点所有新增UI-E2E结果均**Not Run，等待宿主**。

| 旅程 | Start State / Actor | 连续UI Steps | Final Fact / 原有门禁关系 |
|---|---|---|---|
| J01 | 独立init，无用户资产；平台→真实租户管理员/开发者/财务 | 键盘空必填提交→修正→建租户；错误服务地址保存→草稿保留/秘密清除→修正/核验；两模型/价格/grant/可信上界/预算；三角色邀请/接受；租户角色签Key/分配开发者→调用/同ID；最后管理员拒绝移除 | 当前核验适用；真实价14微美元/成本7；Key/tenant归属；角色只读及诊断分配。默认legacy空租户保留如实说明 |
| J02 | 同租户管理员会话/同应用/两获批模型 | 模型目录→Key编辑收窄→403零dispatch→修正→显式新调用/同ID | 旧请求价快照/费用未变；首Key签发接在J01同会话，非平台代签 |
| J03 | 有历史/已用的应用，无过渡 | 1h→新/旧调用→禁止24h叠加→提前结束→旧401；24h同链；额外Key撤销；主Key停用/恢复；真实期限Key测试时钟推进60s→过期；同Key历史 | 两代同池/不清已用；精确1h/24h安全元数据；状态表及秘密关闭不可重取。保留原立即轮换和Go纳秒截止门禁 |
| J04 | 硬费用预算/真实已用>0 | 配额看已用→预算调至已用→费用429→提高→新调用/同ID | tenant_budget_exceeded且零dispatch；提高前后已用相同；新账本14微美元。原次数/RPM/TPM/C门禁不改 |
| J05 | 两真实租户负责人，各有效Key；Alpha两个有效Key | Beta创建/接受/预算/授权/签Key；Alpha暂停→两Key403（同时普通/SSE）→Beta成功→恢复两Key成功；开发者降财务/删除；旧会话刷新 | 后端实时role/诊断分配删除、删除后scope404/空；应用Key仍可用；历史/已用保留；不把人员账户仍可登录误写成退出租户失败 |
| J06 | 真实旧价调用；同ID未知；时钟形成结束周期 | 发布新价→新21微美元/旧快照保留→未知调用→当前结账拒绝；推进UTC月/显式重新登录→pending拒绝→UI累计核对→结束周期结账→追加+5更正/原因/当前调整周期→详情/CSV/财务只读/另一租户404 | 原金额不变，pending归零，调整只计当前周期一次；旧周期CSV只导原周期责任，不把跨周期更正硬塞入旧周期总账；保留原核对幂等/并发Go门禁 |
| J07 | UI产生>25记录，平台具两合法scope | 同Key历史→页2→保留Key筛选/导出；同ID详情跳当前日志；hold实际旧scope响应→换租户→release/完成→范围不覆盖；注入GET503→错误不是空→重试；刷新/退出/重新登录 | 原范围过滤/导出隔离、旧响应不能覆盖、秘密/对话/本地存储不恢复；注入只改变读失败/时序，不伪造成功数据 |

旅程顺序J01/J02/J03/J04/J05/J07/J06：先同月分页，再推进月形成真实结束账单，避免SQL改日期；不是用户中途换身份或拼接API子集。每个viewport只有独立一套产品实例和真实会话，各角色context仅在该实例内。旧native兼容/UI资产继续单独保留。

## 验证记录和下游要求

本地原生沙箱稳定监听/Chromium限制已证，原命令/exit2/受限范围保存在acceptance/evidence/qa-make-verify.log及原attempt，不重复执行make verify，也不申请扩大权限。当前代码对该受限范围仍为待验；不能复用旧本地失败或seq21绿作为新候选结论。

同目录frontend-test.log保存新增及旧Node契约完整结果；checks.log保存format、module、vet、build、JS语法、7项后台定向in-process/exact-deadline回归和browser编译。Node/VM、in-process和编译均只属本层证据，不代表真实HTTP/浏览器。Go linker非fatal warning保留，临时路径脱敏。没有监听检查通过声明。

实际结果：frontend-test.log最终157 tests/157 pass/0 fail；初次155项和筛选保持之前156项日志分别保留frontend-test-initial.log、frontend-test-before-filter.log。checks-final-ui.log中format/module/vet/build/JS语法/diff-check、157 Node、下面七项定向Go race命令及browser compile-only均exit0；checks.log/checks-final.log保留较早补充检查。详情跳日志保留已有Key/模型/时间筛选，新增实际函数回归；测试时钟固定在当前UTC月首日中午，避免真实月末的60秒过期步骤意外切月。最后对该test-only clock和规范更新的format/vet/build/语法/Node/browser编译复查记录final-static.log；不重复同一已知监听失败。完成trellis-check：应用代码与测试/状态生命周期一致、权限仍Go服务端、scope/秘密清理保留、无门禁删除/无新依赖/无调试秘密输出。完整测试待宿主，不把本地质量审查勾成完整通过。

```sh
go test -race -count=1 -run '^(TestConsoleAccountingInProcess|TestConsoleGovernanceInProcess|TestConsoleResourceDimensionsInProcess|TestConsolePeopleScopeInProcess|TestConsoleGenerationExactDeadline|TestConsoleIdentityAndKeys|TestConsoleKeyCollectionContract)$' ./internal/relay
go test -run '^$' ./tests/browser
```

执行时使用项目同样的CGO_ENABLED=1、GOTOOLCHAIN=local及工作区缓存；上面第二条明确没有运行浏览器，不是UI Pass。

指纹生成首次直接`go run ./tests/evidence`未带Makefile的工作区GOCACHE，默认用户缓存访问被沙箱拒绝，source-evidence.json为空，未作为证据。那次组合shell继续运行，末尾exit0不代表Go命令成功，Go独立退出码未留存；保留该失败事实而不编造数值。按原Makefile配置工作区缓存重做，实际独立退出码见source-generation.log，只有完整非空JSON才用于交接。不是产品或监听失败，也未提权。

tests必须对source-evidence.json相同指纹执行完整 `make verify`；取得完整非截断log及exitcode后才交独立QA。新增成功需有3视口×7条journey passed日志，`test-results/continuous-{width}-{pid}.json`和before-rotation/key-states/before-correction/after-correction/final截图；失败保存同名JSON（stage/action/status/checkpoints及安全DOM布尔/count，不含秘密或raw错误）。截断时用平台read_command_output保留完整log，不能以缺失/截断作通过。不得删原断言、给未到达旅程补Pass、用static browser.check或API组合放行。

独立QA继续同冻结SHA及新代码指纹逐AC/H复核，并更新自己的QA22缺陷裁决。真实供应商/外部身份/支付、H缺证、正式Git/Actions/部署仍Not Run；授权由既有节点负责，无用户澄清或需求/设计返工阻塞。Codex/gpt-6.1-sol是指定配置；实际模型进程须由平台核对，本节点不自证。
