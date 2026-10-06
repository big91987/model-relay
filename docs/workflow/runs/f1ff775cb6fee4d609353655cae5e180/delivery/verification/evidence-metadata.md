# 成功证据时间／配置版本设计补齐

## 委托与边界

2026-10-06维护者静态核对[交互§3](../../console-interaction.md)发现：概览成功行已显示ID/alias/scope，但缺少该次调用时间与上游配置revision。`console.go` 已从同一持久成功记录返回 `call_evidence.time / upstream_revision`，无需后端变更或迁移。精确记录页有时间不代表概览成功行满足设计。恢复[PRD](../../product-definition.md)、[架构](../../console-architecture.md)、[QA整改](../../acceptance/remediation-01.md)及[seq23重读返工](host-return-08.md)，沿用本Run，不改设计/QA报告，不扩大首次调用切片。

这是静态设计差距，不是原生浏览器失败回执。在途宿主测试按原对象正常完成，不重启、不重放，不把旧对象结果移植到本次指纹；本次完整验证另交固定入口。

## 实施与真实来源

- `app.js` 概览summary使用原textContent，在同ID/别名/范围行追加“调用时间 {known(overviewEvidence.time)} · 配置版本 {known(overviewEvidence.upstream_revision)}”。没有客户端当前时间、全局upstream.revision回退、格式化伪值或新对象。
- 保持原加载前清身份/summary、失败状态、配置变化失效、degraded拒绝证据、异步owner/load隔离与精确动作契约；缺失字段为“未知”，无证据时不展示旧字段。精确记录本身未新增字段或改变协议；原真实revision已供测试交叉核对。
- 只改产品该行、现有VM/原生旅程、Trellis规范与公共交付文档。Go/SQL/schema、Makefile/锁文件、需求/设计及QA资产未改，既有秘密清除和精确ID断言不删除。

## 原固定入口回归

`ui.test.mjs` 新回归故意让全局revision为900、成功证据revision为17，时间为固定历史2001-02-03T04:05:06Z，要求整行取证据值而非时钟/全局版本。继续验证重新读取pending先清、缺失字段为未知、配置变化返回null不沿用、degraded/503清旧值与动作，以及迟到旧GET不能覆盖新证据。原空/保存/成功恢复测试更新真实字段fixture及完整成功行期望，保留原身份/范围/计数检查。

`journey.mjs` 原两屏真实Go旅程增加 `evidenceSummary / evidenceOverview`，导航/刷新/重登录/重载前注册实际overview响应，要求成功行逐字段等于其call_evidence并核对布局无横向溢出。在既有分页旧成功ID场景里，精确动作仍查询同ID，再将显示时间/revision与真实record.time/revision逐项比较；停用上游造成无当前证据、明确注入503时不留旧字段/可点击动作。未知字段VM夹具不是供应商证据，也不冒称原生未返回字段。

所有回归仍在原 `make verify` 的frontend-test/browser入口；没有新替代入口或放宽门禁。真实供应商未测，受控上游仍明确标注。

## 实际证据

| 检查 | 结果 | 边界与证据 |
|---|---|---|
| 修复前新增VM | 77项，75通过/2失败，退出1 | [修复前日志](evidence-metadata-before.log)，Contract红测，不是浏览器红测 |
| 初次产品修复后VM | 76通过/1失败，退出1 | [日志](evidence-metadata-vm-initial.log)，原成功行精确期望尚未包含新字段；按设计补fixture和整行期望，旧ID/别名/范围断言保留 |
| 最终实际app.js VM | 77/77，退出0 | [日志](evidence-metadata-vm.log)，非原生/网络证据 |
| JS语法/Go格式/模块/vet/build/diff | 退出0 | [日志](checks-evidence-metadata.log) |
| 新56文件指纹 | 退出0，HEAD仍d62cd51 | [对象](source-evidence-metadata.json)：`7f6faaac87ad3501cf76108b850db068439f0c6a47c8e10003e5caa4694775c9` |
| 新原生两屏与完整make verify | Not Run | 稳定监听/Chromium沙箱限制未变，不重复必然受限命令；交固定宿主入口 |
| 独立QA/供应商/PR/合并/部署 | 无新增通过回执 | 保留QA首轮No-Go、QA-03驱动差异及全部历次失败；路线2–4未实现 |

Trellis before-dev packages发现成功，真实0.6.15初始化/规范存在，bootstrap不适用；按frontend索引和shared guides核对。trellis-check核对直接字段来源、textContent安全渲染、生命周期清除、回归/作用域和真实分项；update-spec记录不可用时不替代及两屏同记录核对契约。公共交付review保持T001/T002/T003 In Review、M01/M02/M03 Ready for Review，不宣称Done/Accepted。

## 下一责任阶段

等待在途原对象测试正常收尾，不干预或重启。新指纹原样完整 `make verify`，保留实际退出码和脱敏失败首尾；失败正常返工，通过后独立QA逐条18AC/交互§3及首轮整改原层复验，不覆盖先前报告。不能把静态修复、77VM或任何旧对象绿色称为新两屏原生通过。供应商与发布按实际回执，不提前宣称。
