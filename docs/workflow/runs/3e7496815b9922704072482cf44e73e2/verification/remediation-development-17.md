# 执行17：累计deadline归因与独立场景有界调度

## 当前结论与版本
- T001 In Review；M01 Blocked；独立QA6 No-Go保持。新候选尚未获得正式完整冷门禁；本轮不是上线/部署验收，也不保证调度改动已经使全套运行低于deadline。
- HEAD/base仍为 `cd413b5aa24a372a19c19af999a6ef61b5c41572`。147件新候选摘要 `28571f199f257d5c0bb09fc679150ddb8c5beff822c1a8c813a3bb54c064eeb7`，由现有 `tests/evidence` 实际生成，见 [source-development-17.json](source-development-17.json)。旧147件 `24725bbb…` 为执行16失败被测版，不可替新版本。
- 仍只修本产品验证供应链及相关回归调度；产品Go后台/UI、schema3、零软预算/有效UTC/迁移契约均未改。锁文件、Go模块、Playwright1.59.1、冻结Connector argv/900秒、内层12分钟、管理员部署控制器均未改。没有创建分支/提交/推送/PR/部署或改DB。
- 冻结材料沿用design-v0.2.0；archive SHA256 `5f51d35fdc15ed7214749f9413879b4238097f1d6b6bbe2cc0d508d8e103cef7`；实际manifest `.workflow-input/3e7496815b9922704072482cf44e73e2/5f51d35fdc15ed7214749f9413879b4238097f1d6b6bbe2cc0d508d8e103cef7/files/manifest.json`。本轮无需改需求/架构，材料评审状态不变。

## 执行16原始事实（失败保留）
原固定入口打印 `test-results/verification/20261008T183317Z-269232865`；完整runner日志原样归档 [tests-16-verify.log](tests-16-verify.log)、[state](tests-16-state.json)。实际293790bytes/SHA256 `a17eceef5a36d90431fdf1353c7a85ff32ca6f78305c01409bea4d5fa9fbfdb2`，读取字节及hash与state一致。Connector回执294002bytes/truncated=false/exit2是外层输出；本轮read_command_output从offset288000读到294002/EOF，未把尾页或runner文件声称为已读Connector全文。

原state：起点2026-10-08T18:33:17.456Z；结束18:45:21.644141Z；开始无node_modules/npm cache、NODE_PATH unset、无手工预装；源和副本摘要均不变；gate_exit=-1、exit_code=1；owned_group_absent=true、candidate_removed。失败为内层deadline，非缺包/Owner stop。依赖/Chromium最终probe只有完整gate_exit0才运行，因此执行16没有最终成功探针，不据Node测试成功推断全套通过。

实际分界：
- `internal/relay` race包耗时450.421秒，而不是“Go编译一定慢”的假设；其他包也按原日志保留。
- migration在18:41:56.714194Z才开始，即总起点约519.258秒后；耗时13.139秒、ok=true。
- legacy-ui耗时88.239秒、ok=true；routing在18:43:38.204818Z开始、没有end。
- 实际149次routing_call start、149次end、0次pending；最后390px sequence73在18:45:08.586Z结束，elapsed152ms，各原生边界均结束。1280px路由整旅程已有PASS；390px最后cooldown model rendering被进程终止，未完成，不是已证渲染产品缺陷。
- [tests-16-routing-selection.jsonl](tests-16-routing-selection.jsonl) 原样153行/994029bytes/SHA256 `925644e26e103e66ebeada06ebdc89a6cbeb27c5adda56a2d59c70334bde3bd8`；检查均为白名单键盘/DOM边界，没有标签值/密码/生成内容。此文件不证明未完成的冷却或后续console旅程。

据此选择缩减独立测试的串行调度开销，而不是修改浏览器动作/产品冷却或延长deadline。并未证明最终累计时长已足够；需新宿主冷门禁。

## 最小改动与隔离依据
原Go门禁追加 `-parallel=4`，保留 `go test -race -count=1 -v ./...` 原命令/全包/无缓存测试；原门禁顺序不变。只增加 `T.Parallel()`：两类路由状态码子例、资源矩阵tenant/key维度子例、上游失败/超时子例、连接header/body子例、SSE DONE终端子例、独立人员范围和汇总场景。所有原断言/状态码/计数/回读/重启/等待及实际HTTP变体保留。

这些场景各自调用 `installation`、`newConsoleFixture` 或 `governedInstallation`，拥有独立TempDir/SQLite/Server/时钟/计数/受控transport；实际HTTP使用自己的ephemeral listener与test.Cleanup。没有修改http.DefaultClient/DefaultTransport、进程环境、口令算法成本或共享全局状态。共享同一安装的场景内部步骤仍顺序执行；Go1.25语言的range变量各次迭代独立。四个测试槽明确约束CPU/内存并发；不是无限并发。

`tests/verification/dependencies_test.go` 在原固定入口回归中要求并行上限，并继续逐项要求原历史红灯分类器、格式/module/vet/Node/race/build/browser及锁定安装早于消费。独立消费入口仍先npm ci。原runner的冷复制/两配置/两路日志/取消与文件根边界修复均保持。

## 真实局部验证
全部以下文件已保留，路径脱敏仅将本机workspace/temp/toolchain/cache路径替为占位，不删堆栈/断言/退出码。局部使用项目 `.cache/go-build`，不是正式独立冷起点证明。

| 实际执行 | 结果/范围 | 文件 |
|---|---|---|
| 默认环境的13状态码定向race | exit1，默认Go cache写入受限，未执行场景；不归为产品失败 | development-17-serial.log |
| 设置项目GOCACHE后的同13状态码原串行race、count1、parallel4 | exit0；父30.58秒，包32.244秒 | development-17-serial-project-cache.log |
| 同13状态码/原断言/count1/race、场景Parallel后 | exit0；包11.765秒（父0.00不包含子例，不使用该值当性能证据） | development-17-parallel.log |
| 五组新调度场景一起跑in-process race、parallel4 | exit0/31.548秒；路由allowlist、资源矩阵、人员范围、汇总、SSE终端，所有子例通过 | development-17-inprocess-race.log |
| 194Node + runner/依赖/browser Go包race | 194Node通过；runner12.850秒及依赖3.195秒通过；随后browser实际HTTP监听bind operation not permitted/panic，组合exit1；静态链未继续 | development-17-check.log |
| 单独无监听BrowserPhase回归 + format/module/vet/build/diff | exit0，BrowserPhase1.443秒；all modules verified | development-17-static.log |

本轮宽局部检查触及已有沙箱监听限制的实际命令及堆栈已保留；不扩大权限、不重复完整make verify或该必然受限检查。完整入口在这种原生环境的既有受限记录仍有效；实际HTTP变体（包括此次标记的并行子例）、浏览器原生动作、registry/Chromium及冷全套时长只能由正式tests宿主执行。局部in-process通过不是监听、真实供应商或OS浏览器证据。

## Trellis与交付检查
已恢复总览/T001/M01/QA6与执行15失败交接，读取实际Skill、公共交付references、Trellis前端/后端索引及相关运行/路由/验证/依赖规范、共享思考指南，并通过get_context确认single-repo。Trellis已有真实spec/版本/工具，prepare initialized=false仅保留为历史回执，不据此重新初始化或伪称新bootstrap。此轮before-dev边界明确为验证调度；update-spec维护原dependency-preparation契约；不复制PRD/设计。

check实际对照：源码差异只有测试调度、Go门禁并行上限、对应守卫与说明；产品函数/迁移/预算/UTC/锁无diff；无断言删改、禁用race、skip、fill/selectOption替换、增限或权限扩张。format/module/vet/build/diff实际绿；新组合race与194Node绿；完整make verify/真实HTTP/浏览器仍未通过，不能将check描述为整套通过。文档/格式最后核对通过，源码戳由当前内容实际重算。原执行5/6/8/10/12/14/16证据未覆盖。

清理范围：原state测得本次owned group absent/candidate removed；本轮对应原候选路径实际不存在。只读ps检查53219被沙箱拒绝，故未能独立复查该PID是否存在，也不从路径不存在推广为所有子进程无残留。没有kill任何进程、监听替代服务或触碰用户预览。首次宽局部失败由httptest立即panic退出；其他测试自有临时目录由标准test cleanup管理。未知残留保持未知，宿主应按真实进程归属证据判断，不能广泛kill。

## 正式tests交接（原入口，不推荐替代命令）
1. 冻结Connector仍只执行 `make verify`。新候选digest见source-development-17；开始时由原入口新建独立副本，无node_modules/npm cache、NODE_PATH unset/无手工预装；保留完整日志与版本绑定。不得复用旧seq5/16或原工作区缓存当冷证明。
2. 读取实际打印的唯一receipt目录，归档完整verify.log和原state.json至本Run索引具体文件；对比log_bytes/SHA256和before/after/source/candidate不变，确认全门禁/race/历史红分类器/预算UTC迁移/194Node/全部browser阶段及最终依赖/Chromium探针完成。cleanup/结束/退出字段缺失、race/断言/超时都是失败或未完成。
3. 保留每个browser阶段UTC/elapsed、各包耗时和routing_call最小诊断。再次失败须据新边界（累计、单步、断言或网络）返研发，不盲增限或重跑同一完整失败。HTTP子例可能因并发暴露新竞态，按产品/测试同步事实修复，不能松断言。
4. 隐私检查后仅归档所需具体browser文件；实际截图/边界输出按host receipt报告，raw trace/密码不进公共目录。完整日志若Connector投影省略，使用实际runner完整文件，明确长度差别；不手抄全文或修state。
5. 新完整冷绿后独立QA审查原材料与实际版本。QA原生browser Blocked/供应商未验保持。通过后授权Connector新PR/精确已验merge及协调者正式Owner deploy=true/target=preview；需成功Deployment/environment_url/healthz version/schema3/真实admin页面，研发未完成这些，不能宣布上线。按最后用户事实旧预览9851177/schema2，未作新健康查询，不重置数据/密码或复刻部署控制器。

阶段Ready for Dispatch（验证），不是Accepted/Delivered。全部具体文件入口见 [index.md](index.md)，现有QA6报告、矩阵与补证交接仍为下游审查基线。

最后一致性收据 [development-17-final.log](development-17-final.log) 实际重算147件摘要与本轮source文件一致，源码/格式/diff检查exit0、15项交接文件非空、执行16原日志长度/hash匹配；其计数不包括该检查日志自身。之后仅补此索引与事实说明，二者均在源码戳排除的交付记录范围，不改待验代码。
