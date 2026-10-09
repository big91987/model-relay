# tests5退回归因与development6复验

日期2026-10-09；T001 In Review、M01 Ready for Review/Verifying；非产品Go。本记录补充而不覆盖results.md执行4历史。

## 恢复与版本

同一材料version `product-test33-handoff-r1`、ZIP SHA256 `a13fb6b9089cefbb9d2b6f9a333830e0ee6866bd7ff45d5ba881da97e9a66a49`；执行6再次核验manifest全部61项大小/SHA256一致，原材料及冻结设计未变。

tests5 HEAD仍`3f9cee1615e87e002a2c2300f2d29778072ec3f5`，工作树摘要`ebfc78a9bc25c69e3bacffe21f4aab956fd70d4d9af07dc12a8c97f7b61c7a9d`；host-tests5-state.json记录source_before/source_after相同、candidate_unchanged=true。新候选见candidate-source.json；前次源及代码摘要分别保留candidate-source-exec4.json、candidate-test-code-exec4.sha256。

host-tests5-full.log从宿主持久化回执逐字节归档，483297字节、SHA256 `fa5961f25f65accec113d6bd88c05084e5c8ad65a6b0745a51541fbf34db2aa8`，与原state一致。Connector483510字节还含外层失败行，不混作同一个摘要。原真实选择诊断归档host-tests5-selection.jsonl。原沙箱监听首失败继续保留，不改为通过。

## 已证事实、未证假设及责任

| 类别 | 已证事实 | 处理/责任 |
|---|---|---|
| DS-SAFETY-01“后端无默认”候选 | 旧候选SafetyActualHTTP正式bounds→普通/SSE受控上游有效64/显式32通过 | 内部实际HTTP不复现广义缺陷；非真实收费/隐藏推理证明 |
| RPM/TPM测试资产缺口 | 全部metering InProcess/ActualHTTP（安全、终态、窗口、共享/代际竞争、上界失信、未dispatch、timeout、cancel）及原Go包race通过，含数值/副作用原断言 | 旧候选内部证据非供应商事实；新候选宿主复验、独立计划/QA评审未完成 |
| 旧完整门禁 | relay包423.472秒；migration13.491秒、legacy-ui89.624秒成功；routing1280通过；390最后call73于12:33:39.673Z结束后等待冷却文字，十二分钟gate deadline触发SIGTERM | Connector exit2，state gate_exit=-1/exit1、cleanup成功；不是外层900秒超时或沙箱bind限制；完整门禁Failed/Incomplete |
| 冷却显示 | 旧报告无390边界API/DOM数值，原断言未自然结束 | 产品/驱动/冷却过期/时序根因尚未确认；补最小白名单诊断后交tests，不猜修产品 |
| 真实供应商普通/SSE、成功恢复、Token/费用对账 | 本研发新外部生成0；名称/收费/隐藏推理上界/历史预算尚未证明、公开材料无凭据 | 仍Blocked/Not Run；专用合法测试Pipeline负责，24次/64tokens/90秒授权不变、未知不清零 |

## 本轮改变边界

已证调度问题：本任务新增metering测试独占安装却串行。旧宿主Terminal两组各约19.7秒、Window/Competition四组各约9.9秒、Safety两组各约7.5秒。将十五个顶层TestMetering case在fixture创建前标记Parallel，复用原parallel=4上限。各组独占TempDir/SQLite/key/Server/clock/transport/channels/监听cleanup，无全局环境或DefaultTransport修改；共享fixture内步骤/竞争屏障不变。本地八组InProcess race共22.243秒，不是完整冷宿主门禁加速证明，不承诺门禁已适配。

冷却根因待证：cooldownObservation读取原UI导航实际GET响应，投影模型/候选冷却、过期、失败计数、next-page布尔及有限剩余毫秒，并记录读边界DOM精确目标按钮数/冷却文字布尔。可能先于UI完成渲染，不以瞬间差异判产品缺陷。无额外API请求/重试，无点击填充替换键盘，无断言/超时/31秒自然恢复等待变更；Node回归只证明白名单投影，不证明实际浏览器。

生产Go/UI/schema/依赖锁未变；Makefile、verify十二分钟deadline及原历史/进程/浏览器门禁未变。不修无关产品问题、不涉及平台通用开发、不发送真实供应商请求。

## 本地验证与Trellis检查

| 实际命令 | 结果 | 文件 |
|---|---|---|
| `GOCACHE=.cache/go-build CGO_ENABLED=1 GOTOOLCHAIN=local go test -race -count=1 -v ./internal/relay -parallel=4 -run '^TestMetering.*InProcess$'`（GOCACHE执行时解析为当前工作区） | exit0，八组通过，22.243秒 | metering-parallel-inprocess-exec6.log |
| `make format-check && go mod verify && make vet build && go test -run '^$' ./... && node --check tests/browser/routing.mjs && node --test tests/browser/cooldown-observation.test.mjs tests/browser/routing-progress.test.mjs`，同Go环境 | exit0；格式/模块/vet/build/全包编译，投影/计时3测试通过 | static-exec6.log |
| `node --test tests/browser/*.test.mjs` | exit1，213通过/1文件失败：缺playwright-core，非完整前端通过 | frontend-exec6.log |
| `npm_config_fetch_retries=0 npm_config_fetch_timeout=10000 make frontend-test` | exit2，锁定npm ci在沙箱connect EPERM退出，前端阶段未运行。仅缩短网络失败等待、不改锁/入口 | frontend-locked-exec6.log |
| 新候选完整make verify/实际HTTP/浏览器 | 本地Not Run：稳定监听限制未改变，不重复必然失败命令，交tests宿主 | baseline-make-verify、listener-diagnostic历史保留 |

frontend首失败的原摘要/脱敏规则保留evidence-index-exec6.json；去除本机路径后不称逐字节原日志。网络失败不当依赖已安装，不降低验收；锁定依赖安装仍由原宿主门禁执行。

trellis-before-dev已执行packages discovery，读取实际后端/前端规范及shared guides。Trellis脚本/非空规范存在，无需bootstrap/re-init；update-spec追加源码及回执支持的fixture调度/冷却观测契约。trellis-check审核：无生产跨层字段/状态/费用变化；fixture独占及清理、原断言/竞争屏障保留；相对模块可解析且有测试；不含秘密/body/本机路径；无新依赖/警告抑制。完整门禁和冷却事实仍需宿主，不能标全部检查通过。

## 下游及完成边界

tests对candidate-source.json精确工作树执行完整make verify，保留冷安装/review-history/health-history/race/进程/所有Go服务浏览器旅程，保存新state/full log。仍失败时先核对routing_cooldown的API冷却/剩余毫秒/精确目标行/DOM读边界及相位耗时，区分产品、驱动、时序、整体预算；保留原断言，不轮换无证据修复。

然后独立评审test-plan原AC/PLAN-01–12映射与数值，交独立QA。旧版HTTP绿不当新版完整绿；原#33 No-Go、真实供应商Blocked、未执行QA不改变。确认产品缺陷数仍0；本轮是已证测试调度改进及待归因诊断，不是已修好的生产故障。Git/PR/publish/deploy仅Connector操作。
