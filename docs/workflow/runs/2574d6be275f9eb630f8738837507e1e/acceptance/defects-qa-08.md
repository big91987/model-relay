# Issue30 QA8 缺陷与残余风险

| ID | 等级 | 用户能看到的问题 | 复现步骤 | 影响 | 定位信息 | 当前状态 | 回归资产 |
|---|---|---|---|---|---|---|---|
| R30-SYNC-01 | P2，验证可靠性；未证持续最终错误 | 保存对话框消失后，自动化可能把本次读回未完成的页面当已完成 | tests5同role/scope实际PATCH200→hold tenants/catalog→旧save提前返回；释放后四组DOM正确 | 自动化过早严格断言可中断验收；不能据此声称写丢失或产品持续错状态 | verification/host-tests-05/model-save-readback-1280.json/log/state；workspace edit/sync/render与journeys.save | Closed（仅已证完成契约缺口）；tests7相同原层受控hold红转绿，QA8复核同版 | tests/browser/model-save-readback.mjs；model-save-readback.test.mjs；save-lifecycle.test.mjs；host-tests-07三视口实际JSON |
| R28-04 | P2，既有Owner/QA分级 | 旧tests18在1280摘要启停严格行/列/状态断言失败，验收被间歇阻断 | 原summary-model-context-toggle/form-submit；原failure_boundary不足，具体失败行/列/值与等待原因未知 | 已证交付延迟；未证明持久产品P1 | 旧Run acceptance/summary-assessment-qa-21.md 与原tests18证据；本轮tests7仅为新的成功实例 | Open；不因新控制复现修复/三视口绿而关闭，不倒推旧根因 | 原summary严格负例/原错误身份保持；同轮management-summary与原J01–J07 |

本轮没有新发现P0/P1；这不是“证明不存在缺陷”。若正式prepare/部署或后续原路径出现刷新完成后持续错状态、缺列/重复行或必须隐藏操作才能继续，保留该次原错误/安全JSON/UTC/版本，按核心用户结果建议P1并No-Go，返回development最小归因，不完整重试求绿。

## 非产品缺陷的未执行项

- 独立native动态QA：Blocked；保持原监听及依赖连接EPERM证据。本轮独立执行75项Node及源码/回执复核不是独立Chromium。
- 真实供应商、支付、外部身份联调：Not Run；受控HTTP不是供应商事实。
- Owner真实登录预览、新候选正式prepare/部署/Deployment/health/schema3核对：Not Run；交接后由已有授权链执行，不以旧部署补齐。
- 顶层 TestConsoleClockRecoveryWorker：Not Run（实际skip）；同轮父InProcess/ActualHTTP以子进程运行并断言CLOCK_RECOVERY_PASS，替代的是恢复门槛，不是改写该skip。
- 无需新增研发整改包：本轮当前强制候选门禁通过，无新实现/设计/需求缺陷待修。保留原R28-04风险随report交付。
