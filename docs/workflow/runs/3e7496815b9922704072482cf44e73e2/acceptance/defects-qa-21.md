# 执行21 缺陷与剩余风险

本表更新当前候选范围；QA6、执行13红灯、执行18失败、所有历史state不覆盖、不改写。

| ID | 等级 | 用户能看到的问题 | 复现步骤 | 影响 | 定位信息 | 当前状态 | 回归资产 |
|---|---|---|---|---|---|---|---|
| P1-1 | P1 | 全新安装验证因缺少浏览器模块停止，不能发布 | 原main无node_modules/NODE_PATH unset运行frontend-test，192/191/1且MODULE_NOT_FOUND；正式Owner失败是用户提供事实 | 新安装及发布 | 原依赖只在最后browser安装；现在消费者显式依赖verification-deps | Closed（候选修复）；正式prepare仍待执行 | verification/frontend-before.log、clean-before.log、dependency-regression-before.log；当前host-seq20完整冷make verify及独立依赖顺序/安装失败停止 |
| E28-01 | 强制AC证据缺口 | 暖工作区绿灯不能证明首次安装 | QA6检查seq5起点，缺独立目录/无模块证明 | 发布审批无法放行 | 固定Go runner创建私有candidate并原样记录前后 | Closed | host-seq20/state.json的clean_start_measured/NODE_PATH/manual_preinstall/source_before/after |
| E28-02 | 强制AC证据缺口 | 无仓库内完整宿主日志，后续审查不能复核 | QA6仅有摘要与平台日志 | 追溯 | 本轮原样归档完整runner，不冒充Connector全文 | Closed | host-seq20/verify.log与state hash/bytes、archive-index.json |
| R28-01 | P2（原Owner提出Important） | 验证复制可经父链接读入根外文件 | 临时源tracked internal父目录替为根外symlink，copyFiles | 验证供应链边界 | os.Root＋逐级Lstat＋目标边界；未删除链接绕过 | Closed | development-13-red.log原红；本轮TestCopyRejectsParentEscape/TestCopyRootAndTargetBoundaries/失败清理独立race绿 |
| R28-02 | P2（原Owner提出Important） | 辅助失败只报exit，原stderr原因丢失 | evidence/probe失败fixture分别输出两路诊断 | 故障归因 | captureCommand两路采集，stdout另作JSON解析 | Closed | 原development13红；本轮TestFixedRunnerReceipts/TestCaptureChunkBoundariesAndFailure/TestDependencyProbeChromiumFailureOutput绿 |
| R28-03 | P2（原Owner提出Important） | 取消辅助阶段不能保证子进程退出和state落盘 | 私有helper在inventory/probe阶段接取消或SIGTERM | 现场与资源恢复 | 同一有界监督、Setpgid/WaitDelay/自有组清理 | Closed（当前工具）；不归因旧seq12中断 | 本轮TestInventoryCancellationReceipt/TestOwnedGroupCancellation/TestRunnerSIGTERMPersistsCleanup/各阶段cancel fixture绿 |
| R28-04 | P2（测试可靠性风险；非已证产品竞态） | 历史一次模型启停摘要断言失败，原因未定位 | tests18 1280px summary-model-context-toggle；无法保证稳定复现 | 正式prepare可能再红；当前未观测失败 | 仅新增白名单边界诊断，产品/同步未修 | Open；研发后续跟踪，非Owner已接受风险 | 原tests-18-continuous-1280.json/log保留；当前20三宽度原summary绿；再红使用summary_cell/failure_boundary/content_loading返development |

R28-04不静默升级或假定业务错误：历史已证的是断言失败，不知道行/列/值或调度边界。当前完整原层绿满足本轮用户规定的候选放行条件，不能称根因根治。其余环境Blocked与供应商/正式部署Not Run不是已修复产品缺陷。

维护者后续评估：见 `summary-assessment-qa-21.md`。原三类严格断言与false→true路径保留，产品dialog.close→异步syncTenants→renderPage顺序未改，driver.ready缺当前刷新代次屏障。仅静态同步契约缺口，不能认定seq18实际根因；旧failure_boundary仍未知、不得补造。4项定向Contract测试通过不等于动态UI重现。R28-04仍Open/P2验证可靠性风险，当前候选可进入强制正式prepare；一旦实证已完成刷新后的错状态/用户核心路径错误则按P1整改，不以P2规避强制AC。
