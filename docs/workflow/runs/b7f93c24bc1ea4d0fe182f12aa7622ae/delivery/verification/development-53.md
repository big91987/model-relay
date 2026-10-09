# 执行53：模型详情关闭边界的最小诊断

## 恢复与版本
- HEAD仍是 `3f9cee1615e87e002a2c2300f2d29778072ec3f5`；执行52冷宿主before/after均为154件候选 `454ba32f380c184b2bfde486bf3c38a1854c2290563c38ec26a72580c518fec8`，与执行51交接相同。现有产品修改、冻结方案、QA14 No-Go及失败历史保持，不重放未知写入或供应商调用。
- 沿用已有非空源码规范与Trellis scripts；get_context实测single-repo、backend/frontend。prepare initialized=false不改写成新初始化成功；本轮无需bootstrap。
- design-v0.2.0材料archive SHA256实核 `5f51d35fdc15ed7214749f9413879b4238097f1d6b6bbe2cc0d508d8e103cef7`，状态仍frozen-for-test-implementation。manifest在 `.workflow-input/b7f93c24bc1ea4d0fe182f12aa7622ae/5f51d35fdc15ed7214749f9413879b4238097f1d6b6bbe2cc0d508d8e103cef7/files/manifest.json`。引用同root的docs/prd.md、docs/interactions.md、docs/acceptance.md；实际路径/摘要/匹配结果见 `development-53/retained-integrity.json`。interactions第17行明确关闭返回原触发点，不删除该失败断言。

## 原宿主证据与最小归因
- 原state、完整verify.log及浏览器证据79件逐字节归档在 `host-tests-52/archive-index.json`。日志705216bytes，SHA256 `bce53e5cb908a7dd0f4c5b0de5aff3367f4a509953ab24cd5762ac18feae311f`，与state匹配；Connector总输出705398bytes含外层输出，不是该文件大小。UTC10:17:34.989092至10:27:43.957583；gate_exit2、runner exit1、Connector exit2；冷起点实测、source/candidate未变、owned组消失及候选清理均有真实回执。
- 270 Node、七包Go race及migration/legacy-ui/routing/console阶段通过。full-console在1280完成J01–J07、当前周期visual和management-summary后，models-detail后续关闭段失败。新self_hosted未进入，1440/390 full-console未完成，不能拼前轮绿。
- `host-tests-52/browser/visual-1280/models-detail.png.integrity.json`实际passed；原PNG已查看，有实质模型详情，不是空图。root宽1280，modal root clip/body fixed/dialog auto，原一次scrollBy位移0、focus_inside=true；title/content像素和五项document几何通过，checkpoint已保存该截图。
- 旧报告仍标check=screenshot-content-integrity，capture_integrity=null/readability unavailable。调用路径定位到截图后的关闭段，但无法区分native ownership、root/body解除锁与精确trigger焦点四项失败。关闭事件异步交付、锁清理、焦点错位均待证；不再轮换猜测性CSS/截图/产品时序修复。

## 改动与验证边界
- 本轮仅改 `tests/browser/console-visual.mjs`、对应单元测试、Trellis合同及本Run公共证据。workspace.js/css、后端、schema、账号、费用/Token/授权、安全口径与冻结设计未改。诊断就绪不等于QA58-01或self_hosted交付。
- 原close-button click前注册只读native close监听，保留原hidden wait及四项严格断言。成功把关闭事实附到截图；失败标close-detail/close-editor并记录固定assertion boundary、root/body枚举、有限坐标、trigger连接/精确焦点、close事件次数及事件时锁状态。监听在产品既有监听之后注册；零事件只是观察，不推断根因或放行。
- 无DOM文本/值/ID/URL/秘密或原始error输出；无额外截图、sleep、RAF、重试、settlement等待、trigger替换、deadline/锁/权限修改。监听移除并dispose；清理失败不覆盖原断言失败。夹具只证明诊断合同，不证明真实浏览器机制。
- 两次apply_patch分别因规范/里程碑标题锚点不存在被整体拒绝；核查均无部分写入后，用真实锚点重做，没有重放业务操作。

## trellis-check与公共交付审查
- `development-53/local-checks.json`记录准确命令/UTC/退出码/日志摘要：131相关Node通过，覆盖visual、capture、readability、journeys、workspace、readback、save-lifecycle；JS语法、diff、format、vet、build、mod verify及verification wiring race通过。新增三项Node回归，宿主全量预计273项，原270项保留。
- 复核原scope、根溢出、局部滚动、秘密清理、一次fullPage/像素/document几何与四条关闭断言未减弱；observer无产品写入，复用限两个既有关闭调用，合同已同步。无新增框架、公共API或跨层数据流；公共任务仍唯一T001，无重复状态源。
- 原生监听环境未变，沿用先前限制记录，不重复已知必败完整make verify；未手装本地缺失playwright-core。局部检查不是完整固定门禁替代，宿主须原冷npm ci/锁定Chromium。静态构建、隔离单元测试与证据归档是本轮副作用，无账号/DB/供应商/发布操作。
- 新154件候选 `00d8519e5b54d97eccce2cae8652735bec46319108044d0d075c96da9b33ce8c`，Go1.25.6。见 `development-53/source-final.json`、`development-53/final-integrity.json`。

## 交接与未决
T001 In Review/M01 Needs Revision/QA14 No-Go保持。交tests原冷 `make verify`，保留十二分钟总门禁/九百秒外层/parallel4/Playwright1.59.1 Chromium147与原权限。失败依据新的modal_close事实回development，再实施有事实依据的根因修复；不能把未收到事件视为可放行。

固定入口仍须完整Node/Go race/HTTP/预算UTC迁移/取消权限撤销恢复、三视口J01–J07/summary/visual及首次self_hosted网页写入、不可达/核验失败纠正、普通/SSE同ID日志费用。完整绿后独立QA核冻结设计、QA58-01及安全截图。旧空图/R28-04根因、动态原生独立浏览器、真实供应商/支付/身份、Owner preview已登录实操继续分别未闭合；受控HTTP不能证明真实供应商。本轮不声称产品根治或验收完成。
