# 执行11：修复 npm 前置配置及诊断缺口

## 已恢复基线和失败
- 基线 HEAD：`cd413b5aa24a372a19c19af999a6ef61b5c41572`。执行9候选144件、`af09459112ac061976688f434891d3f859e476c0efdfd285b645f504e9759ca3`仅覆盖旧代码。
- 正式tests10：冻结argv `make verify`，exit2；239bytes/truncated=false，已通过平台read_command_output seq10读至EOF。完整平台原文保留在 `/api/workflow-runs/3e7496815b9922704072482cf44e73e2/steps/10/output?format=text`，没有将手抄摘要当全文。
- 正式入口实际打印 `test-results/verification/20261008T173346Z-400132109`。从该目录直接复制的 [verify.log](tests-10-verify.log) 与 [state.json](tests-10-state.json) 为原始失败产物。日志40bytes、SHA256 `5e6ca2ab71099bfc2923c05fe4c383f93c6965809cc8da7470c58291041a87e8`与state一致。失败在npm版本前置，gate_exit=-1；尚未执行npm ci、历史HTTP或浏览器，不能称为完整冷验证。
- seq5旧exit0、QA6 No-Go、seq8 Owner stop和执行7/9证据全部保留，没有改写为新通过。

## 最小诊断与改动边界
已证源码将NPM_CONFIG_USERCONFIG和NPM_CONFIG_GLOBALCONFIG同时指向/dev/null。真实本机npm `--version`在该环境下退出1并输出double-loading config，见 [最小复现](development-11-npm-before.log)。这不是网络、监听、业务或异步时序故障。旧runner使用Output丢弃stderr，正式seq10没有该原始npm诊断；本机诊断支持相同配置缺陷的归因，不冒充seq10宿主stderr。

最小修复在 `tests/verify/main.go`：创建副本内两个不同的空私有配置文件；继续清除继承的NODE_/npm配置并保持原有冷cache约束。版本探针改为CommandContext/CombinedOutput，将stdout和stderr一起进入已有脱敏日志；退出失败仍终止原门禁。没有手装依赖、global module、NODE_PATH、放宽权限或改Connector。

回归在 `tests/verify/main_test.go`：实际npm解析user/global路径并执行--version，确认未装依赖；新增失败前置夹具验证stderr保留、非零传播、gates未启动、完整日志hash/bytes以及只清理自身副本。原gates通过/失败断言保持。实际取消测试仍只检查本次拥有的进程组，不声称旧seq8现场或所有端口消失。

README与Trellis依赖契约同步两个独立配置路径及诊断约定。Makefile、锁定依赖、Go业务API/UI/DB、budget-clock-r1/UTC和历史分类器没有本轮改动；不复制Platform部署控制器。本Run只维护现有T001/M01，不增加需求、设计或平行任务。

## 研发检查与限制
- before-dev：实际Trellis脚本识别single-repo及backend/frontend层；规范已具源码契约，不是空模板，不重新初始化，也不声称prepare initialized=true。本轮读取实际依赖/验证规范、公共交付契约及共享思考指南；按update-spec同步实际缺陷经验。
- [定向最终race日志](development-11-tests.log)：真实npm配置测试、自有子进程取消、采集负向夹具、固定入口顺序及安装失败停止回归，exit0。夹具中故意exit23/24是PASS的负向测试，不是完整产品门禁结果。
- [最终静态检查](development-11-static.log)：format、go mod verify、go vet、build、git diff --check各exit0，并输出当前源码戳。最终身份见 [source-development-11.json](source-development-11.json)，须与日志候选一致。
- 最终候选144件，SHA256 `4d484086e12d71308e42b963ec0266f5f756924f8073defd1c5622a6260c9025`，HEAD仍为上述main基线。最终定向race和静态日志同版；本轮初次及中间检查保留在索引，不能代替最终候选证据。
- trellis-check：范围、source/env→独立配置→gates→日志/清理路径、失败传播、原断言、locked modules和复用tests/evidence已审查；没有抑制警告或业务行为变化。业务跨层读写本轮不适用。原生网络/监听限制未改变，保留首次失败，不再重复必然受限的完整make verify；真实npm前置可执行，已实测，而非归为环境限制。
- 本轮没有启动产品或浏览器服务，没有杀他人进程、访问旧数据或重置密码。没有完整冷安装/HTTP/browser新绿灯，QA仍No-Go，真实供应商及正式部署没有新回执。

## 正式下一步及归档契约
1. tests只执行冻结的 `make verify`。不得先安装依赖、删除原node_modules、执行执行7脚本或更改Connector。固定入口自动复制当前144件新摘要候选，实测无node_modules/npm cache、NODE_PATH未设并打印唯一证据目录；本轮源码摘要与执行9不同，旧结果不覆盖新候选。
2. 从实际新目录归档完整脱敏verify.log/state.json到本Run具体文件并更新索引，核对日志bytes/SHA256、源码前后匹配、clean_start_measured=true、前置版本与实际安装后的本地锁/Chromium匹配、finished_utc、gate_exit=0/exit_code=0及owned_group_absent=true。完整新日志必须包含预算UTC/迁移、历史精确分类、193Node及新增Go工具回归、三视口J01–J07/十页等全部既有门禁。非零按实际诊断返研发，不推断缺失日志成功；强制中断仍是不完整。
3. 独立QA审查受控真实Go浏览器产物隐私后，按白名单选实际文件及具体索引归档；原生QA浏览器Blocked、真实供应商Not Run保持分层，不由本轮单测替代。
4. 新完整冷绿及独立QA后，由授权Connector/协调者新PR、精确已验提交合并，再正式Owner deploy=true target=preview；Deployment/environment_url、healthz version/schema3与真实admin页面闭合前不宣布上线。保持旧安装、密码与数据。

交付状态：T001 In Review，M01 Blocked by Validation；研发修复Ready for Dispatch到tests，不是Accepted/Delivered。材料继续使用同一design-v0.2.0回执，sha256和实际引用路径在T001，没有新批准或新增供应商事实。
