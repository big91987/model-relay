# Issue #14 验证历史保护独立验收报告

> 日期/判定时点：2026-10-06 05:01 +08:00；QA 执行自 04:59 起。
> 仓库/分支/Commit：model-relay / workflow/cf691e39d382439f92a3ac12f5facfff / 066f035bc551b2b4d031166cebdab0849084afd4，加未提交修复。
> 当前制品：本地构建，未部署；源码摘要 b33aa9f6f848a8cda1bf5af08ba9c2e540d0aa09f91090930ccafa6d4e636702（36 文件）。macOS arm64、Go 1.25.6、CGO、Node 26.0.0、npm 11.12.1，锁定依赖；独立临时测试数据，无真实供应商。
> 验收模式：短轮复验 + 完整公共固定门禁回执核验。
> 结论：**Go**（本轮同版宿主成功；QA 沙箱完整入口 Blocked 单独保留）。

## 1. 结论摘要
QA12-HISTORY-WRITE 已在原发现的验证命令输出层关闭：QA 用原 main overlay 独立复现三个新回归红灯，修复后的进程级 race 回归绿灯。QA 实际原样 make verify 因监听受限退出 2，完整保留 REJECTED 和失败传播，不宣称本地完整通过。同轮宿主固定 make verify 总退出 0，首段版本摘要与 QA 前后完全一致； unchanged fail-fast 门禁链支持正常历史分类、当前健康/版本、race、前端、构建及真实 Go 浏览器全部执行成功。106 份研发执行前历史清单在宿主后、QA 前后都一致，没有人为恢复旧日志。工程质量提高，无新增产品缺陷；草稿 PR 和交付状态闭环仍属后续报告/Connector，不是已发布。

## 2. 本轮覆盖
权威基线：原始用户任务、本 Run tasks/T001-history-output.md AC、既有运行契约，历史 QA 仅提供缺陷主题和连续旅程契约。独立分析 diff：仅 history main/main_test、项目说明与运行契约改变；Makefile、产品、认证、数据模型、依赖与前端实现均未变，故无需新 PRD/设计或组织级验收文档。

输出正常/拒绝/关闭 stdout、历史哨兵、精确分类 14 负例、实际沙箱异常、固定门禁和历史哈希均覆盖。产品角色、首次使用、生命周期、失败恢复、重启持久事实、秘密排除与版本 UI 沿用既有真实 Go browser 连续旅程；宿主同轮实际重新运行，详见 journeys.md UI01。每 viewport 同一管理员/同一 context，从入口到最终结果；夹具仅用于初始安装、受控依赖与故障/对账，不用隐藏 API 补产品操作。UI-E2E Pass 来源为同轮 Connector 总回执及受信 fail-fast 旅程链，不来自单测或页面静态存在。QA 本地 UI-E2E Not Run 不改写。

## 3. 缺陷及进展
见 defects.md。沿用 P3 QA12-HISTORY-WRITE 主题，本轮 Closed；不是 Accepted Risk 或人工恢复。原实现两个场景改写历史，关闭 stdout 场景吞错；修复全部通过。无新增、复发产品缺陷和未关闭 P0/P1。没有供应商凭据，真实供应商能力不在此任务范围，也不作联调结论。

## 4. 证据
- evidence/snapshot.log、evidence/final-integrity.log：版本、环境、Makefile 无 diff、完整哈希与源码摘要前后一致。
- evidence/host-connector-receipt.md：平台 seq=5 command exit 0 的实际字段摘录、同版核验和截断边界，不是完整宿主日志。
- evidence/qa-original-regression.log：HEAD 原 main overlay + 当前测试，race exit 1；失败发生于临时 fixture。此 overlay 不修改工作树源码。
- evidence/qa-output-regression.log：QA 实际 Go race exit 0，3 个 main 子进程场景及分类负例通过；这是受控命令输出回归，不是网络替代。
- evidence/qa-make-verify.log：QA 实际固定入口 exit 2；新输出可定位，监听拒绝后 REJECTED + stderr，保留完整本地 stdout/stderr，仅系统 make 的绝对路径脱敏为 `<system-make>`，未改变事件和退出结果。
- evidence/qa-non-listening-checks.log：格式、module、vet、27 前端 VM、build exit 0，不拼接成完整 make verify 或 UI Pass。
- ../verification 下研发失败、红转绿、Trellis 和哈希证据原样保留；当前判定不以研发自评代替 QA。

## 5. 未执行和环境限制
QA 固定入口在 health-history 停止；该次后续格式/module/vet/frontend/full race/current HTTP/build/browser 均 Not Run。后续另行补验只对应实际命令范围；当前 HTTP/full race/真实 browser 的发布门禁依据是同轮宿主，与旧 Run 成功无关，不请求扩权。宿主输出在旧版事件处截断，无后段原始日志/截图，不能报细粒度宿主测试数量或视觉质量分数。可信总退出 0 + 同版不吞错链足以核验固定门禁，本地 Blocked 保持原结果。首次 overlay 工具命令被策略拒绝、未执行；后续安全临时文件清理方式成功，拒绝事实见 defects.md。

## 6. 阶段边界
本任务只修复验证工程输出边界，不修改业务实现、不做新供应商/部署/权限能力。静态核验 main 仅把历史 WriteFile/unchecked Print 替换为检查 stdout 错误；assess 与真实旧 baseline 保持，Makefile 全门禁未删减。README 说明默认 stdout/stderr、调用方选择新路径及保存退出码；运行契约原位维护。缺失路径/写权限时由调用方重定向失败非零，不要求新增无必要输出 flag；closed_stdout 回归验证程序内部输出错误。普通技术选择符合用户委托。

## 7. 环境恢复
106 份旧文档/日志全部哈希一致，禁止恢复旧文件掩盖副作用，本轮未做恢复。测试只使用 t.TempDir 和程序临时目录/端口；Go 程序 defer 关闭服务/数据库/移除目录。overlay 的临时源码/映射已删除；忽略的 bin/cache 为本地验证产物，无凭据或用户数据库修改。无共享部署、付费上游或长期测试服务启动；宿主清理依据同版 harness 的 defer，本报告不声称单独观察过宿主进程列表。

## 8. 质量趋势
相对上一轮：关闭 1 个既有 P3，新增 0、复发 0；沉淀 3 个进程级输出/失败/保护场景并进入既有 full Go race 门禁，保留原精确分类负例。独立旧版红转绿证明测试有检错能力，成功/失败不再要求人工恢复历史。残余证据风险为宿主输出截断和本地监听限制，不是未关闭实现缺陷；旧失败日志与所有限制完整保留。产品/UI 没有改动，不据此宣称新业务能力提高。

## 9. Go / No-Go 与报告交接
**Go**：同轮同版固定门禁成功，当前强制修复 AC 有正常、异常和历史完整性证据，无未关闭 P0/P1 或其他产品缺陷。进入 report，不返回开发/设计/需求；report 的公共文档同步责任尚未完成，不把它误列为已 Done 的发布事实。

report 必须按本报告与 seq=5 真实宿主结果同步本 Run T001/M01/README，记录报告时点：宿主测试和独立 QA 已完成，不再写等待执行；保留研发/QA 本地 Blocked 作为历史执行限制，而非继续宣称宿主待验。按挂载交付 Skill 的状态模型确定任务/里程碑/总览当前状态，然后写本 Run pr.md。不得只加 QA 链接保留矛盾，不修改旧 Issue #8/#9/#12 原文，不把报告完成等同提交、PR、合并或部署。尚未发生的发布动作由平台 Connector 处理。
