# 停止升级后的 development 接续核验

## 状态、材料和版本
- 2026-10-09原生接续，沿用本Run、工作目录、固定入口、冻结方案、预算和权限。平台stop是用户报告的中断事实，不推导中断操作成功；本轮未切分支/提交/推送/发布/改变权限或重放不确定操作。T001 In Review、M01 Needs Revision、QA14 No-Go保持；只完成接续审计和最小诊断修正，不是空图修复或完整交付。
- HEAD `3f9cee1615e87e002a2c2300f2d29778072ec3f5`。停止前最新可核宿主候选154件 `cd3175b1ec0a5254aa6460274b3921b56d8cf4639d0485d95e2603280d4d9e45`；本轮最终154件 `7f0e8c2074c493a6176a94dad788c9c65a291d2464a70e484d1fde01e0b1a658`，最终局部检查前后相同，见 `development-resume/source-final.json`、`source-final-after.json`。首次编辑前没有新全量源码戳；恢复时改动清单相容不冒充逐字节核验。
- 已先读manifest `.workflow-input/b7f93c24bc1ea4d0fe182f12aa7622ae/5f51d35fdc15ed7214749f9413879b4238097f1d6b6bbe2cc0d508d8e103cef7/files/manifest.json`；design-v0.2.0、archive SHA256 `5f51d35fdc15ed7214749f9413879b4238097f1d6b6bbe2cc0d508d8e103cef7`，实际引用同root `docs/prd.md`、`docs/interactions.md`、`docs/acceptance.md`。三文件SHA复核见 `development-resume/recovery-audit.json`。仍frozen-for-test-implementation，不改为批准，不执行材料脚本。

## 停止前验证与副作用
- 执行38原入口make verify：`2026-10-09T06:33:15.943627Z`→`06:44:41.916723Z`，gate_exit2、runner exit1；source/candidate unchanged=true、冷起点已测、无手装依赖、owned_group_absent=true、candidate_removed。这只证明该次owned group清理；原生ps被不变的沙箱拒绝，不声称机器全部进程不存在，不杀猜测进程。
- 91件原字节产物保留在 `host-tests-38/`，逐件SHA/bytes见 `host-tests-38/archive-index.json`。完整runner日志560672字节、SHA `2d66a94afd054763a0e42ba616ad62d5ce31d122077f3e15e92e50be05a2e610`与state一致；截断的Connector包装输出不当完整日志，不猜包装层退出。
- 259Node及原前置Go/race/HTTP/UTC预算、migration/legacy-ui/routing/console通过；full-console1280 J01–J07/summary后visual-keys失败。完整1440/390 visual与self_hosted未执行到，其他套件视口绿不外推。原PNG已实际查看并重算：1280×1149、全单色RGB243/246/250、SHA仍 `78a7b304df77dac78964e27f6571f81f2ac48376e545786f9778a183028bff53`。
- trace complete=true但loss=true，仅call marker、8帧unattributed、ignored_events1345；每帧object_hex=true，其余身份形状false。屏外quad仍不能关联主页面截图区间；不是监听限制或已证产品overscroll。
- 初次只读恢复命令因zsh分隔符解释失败退出1，无编辑副作用；随后恢复成功。官方源文件curl因既有代理不可连接退出7，后续只读搜索工具检查官方源码，不执行下载内容。保留这些失败，不扩大权限。

## 最小变更与来源
- 缺口位于 `tests/browser/capture-integrity.mjs:observeCaptureFrameTrace`：原代码仅关联numeric args，宿主实际仅有hex对象ID。只修该身份关联、对应Node回归、已有Trellis合同；另补局部runner首败的原断言错误码。继承的产品展示/self_hosted改动保留，本轮不轮换未证CSS/滚动假设，不改产品/API/SQLite/费用/Token/授权。
- 宿主固定Chromium147.0.7727.15官方源码链：inspector_set_layer_tree_id::Data → ChromeClientImpl::GetLayerTreeId → WebFrameWidgetImpl::GetLayerTreeId → LayerTreeHost::GetId；CreateLayerTreeHostImpl传同一int给Impl，GenerateCompositorFrame以id_输出cc::LayerTreeHostImpl对象快照。不是凭夹具把任意指针当树ID。
- 新路径仅接受确切事件名、legacy O phase、无id2的plain hex正int32 ID，与主frame映射同一正safe-integer PID；最多8个进程键只留内存，映射晚到可关联，结束清空。别名/其他phase/scoped/overflow/zero/无PID仍unattributed，不同进程/树是other；原numeric路径保持。PID/对象ID/树ID/frameID/raw trace均不导出。
- 原类别/512KiB/markers、输入/同步/2000ms精确末列、唯一fullPage PNG、像素/几何/秘密门禁、锁定依赖及预算均不改。matched仅归属；loss/缺marker/非截图区间仍覆盖不完整，不能据此放绿或认定根因。Tracing仍可能扰动时序。
- 官方固定tag参考（非New API）：[trace事件](https://github.com/chromium/chromium/blob/147.0.7727.15/third_party/blink/renderer/core/inspector/inspector_trace_events.cc)、[ChromeClient](https://github.com/chromium/chromium/blob/147.0.7727.15/third_party/blink/renderer/core/page/chrome_client_impl.cc)、[Widget](https://github.com/chromium/chromium/blob/147.0.7727.15/third_party/blink/renderer/core/frame/web_frame_widget_impl.cc)、[Host](https://github.com/chromium/chromium/blob/147.0.7727.15/cc/trees/layer_tree_host.cc)、[Impl](https://github.com/chromium/chromium/blob/147.0.7727.15/cc/trees/layer_tree_host_impl.cc)。

## 本轮检查与首败
- 初次76相关Node通过；完整Go runner在TestFixedRunnerReceipts/cancel_probe报actual auxiliary child remains（exit1）。原日志 `development-resume/go-runner.log` 与命令/UTC/退出 `local-checks.json`保留。旧断言没有记录syscall错误，无法区分存活/权限/短暂时序，不归入监听限制。
- 仅在原断言区分PID解析失败并输出原Kill(child,0)错误，仍严格要求ESRCH；不输出PID、不增加sleep/重试、不修runner行为、不弱化断言。诊断版本单次cancel_probe及完整runner通过，首败原因仍未知，不称清理修复闭合。完整宿主保留此门禁。
- 最终76相关Node、Go runner/wiring race、format/vet/build、非监听普通/SSE fixture、JS syntax、git diff --check退出0；命令/UTC/日志见 `development-resume/local-recheck.json`。本地夹具不是Chromium/HTTP/供应商事实，也不是完整make verify通过。
- 原生监听环境未变，沿用 `development-check.md`、`native-verify-state.json`、`native-verify.log`原固定入口受限首败，不重复已确认必败命令。执行38空图和本轮取消清理首败分别保留，不冒充该环境限制。
- trellis-before-dev恢复single-repo前后端索引及相关DOM/lifecycle/验证/依赖/workspace/reuse/cross-layer合同，真实规范已有，无需bootstrap/重初始化。trellis-update-spec写入有来源的身份合同；trellis-check复核scope/兼容/隐私/回归/原门禁。交付Skill沿用唯一README/T001/M01，不改需求边界，requirements/design不适用返回修订，不补造阶段文档。

## tests交接
1. 本轮最终候选执行原固定冷make verify，保留版本前后/UTC/完整安全日志/PNG与sidecar/锁定依赖/退出/清理，不手装依赖或改权限预算；不重放停止前不确定命令。
2. 空图重复先核phase/同PID主frame匹配、loss、双marker、匹配帧是否在截图区间；仅identity_shape不等于关联，matched但丢失/缺return仍未知。若cancel_probe重复，先读signal_probe及原group回执，不轮换猜测修复。已证产品问题由development实现，不让tests改产品或门禁。
3. 全部原前置/race/HTTP/UTC预算、三视口J01–J07/summary/visual及新安装self_hosted创建→核验失败恢复→关联授权Key→普通/SSE→同请求日志费用完整绿后独立QA核材料/版本/真实图/QA14与QA58-01。T001不Done、M01不Accepted，历史空图/R28-04原因不闭合。
4. 动态原生独立QA、Owner preview已登录操作、真实供应商/支付/身份分别Not Run，受控HTTP不能证明供应商集成。没有需用户决定的新产品边界/权限，可继续交tests，无需再审批进入阶段。
