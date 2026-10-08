# 修复模型保存后同范围页面刷新完成契约

Refs #30。基于 PR29 已部署 main `9bf6ecaf53758ca2258498968b830c535e02e1c0`，仅承接管理页面连续操作中的保存完成契约，不重建需求、架构或 design-v0.2.0，不恢复旧 Run。

## 变更与实证

- 修前正式 tests5 在真实 Go 服务、角色 context 和 UI 模型启停路径中，于 PATCH200 后分别暂停同 scope 的 tenants/catalog 响应：四组保存都提前返回；原样释放后严格最终 DOM 正确。证实本次完成契约缺口，不是持久数据错误的证明，也不能据此倒推旧 tests18 根因。该轮完整门禁失败，失败证据保留。
- `workspace.js` 复用已有 operation/epoch、busy、read/render 生命周期：写确认后显示读回状态，保留 loading/保存锁；只在同操作、当前 scope/page 成功提交 DOM 后标记完成。读失败显示“写入已确认但读取失败”，隐藏旧数据并仅允许 GET 恢复；旧 epoch 不能覆盖新范围或解锁新编辑。
- `journeys.save` 捕获本次 operation/tenant/page 并等待对应终态，再一次性严格断言；不再把 PATCH200、dialog 隐藏或旧 loading 隐藏等同保存完成。秘密/邀请路径、未知写入的幂等与恢复规则保持。
- 在原 `make verify` 内加入真实 tenants/catalog hold/release、GET503 恢复和旧 scope/第二次相反操作迟到回归；保留真正错状态、缺列、重复行仍红及原错误身份。没有 sleep、global networkidle、poll-until-green、assert 重试、提高超时或降低 race。
- 无 Go 后端接口、schema、鉴权、角色、依赖锁、固定入口、Platform、Connector、权限或时限变更；未操作已有 DB、账号、凭据、master key 或部署检查点。

## 已验证候选与质量门禁

QA8 结论：**Go with known issues，仅精确已验候选**。T001 为 Done，M01 为 Accepted（研发候选范围）；版本交付仍 Verifying，不代表新版本已合并或部署。

- report9 时 HEAD 仍为上述旧 main 加未提交候选，新修复 commit 尚未生成。任务分支：`workflow/2574d6be275f9eb630f8738837507e1e`，目标 `main`；候选 150 文件源码 SHA256：`0384f115adbbd545e8ce01021f238010ba66c89eb06e6acea08ce872ce0fa474`，development6/tests7/QA8/report9 核对一致。此摘要范围排除交付回执，不能把旧 HEAD 当已修复提交。
- 正式 tests7 一次完整 `make verify` exit 0，宿主 UTC `2026-10-08T22:19:51.041183Z`–`2026-10-08T22:30:55.912316Z`。format/module/vet、213/213 Node、全包 `go test -race`、真实 HTTP、build、迁移、零软预算/有效 UTC 及历史红分类器均通过；1280/1440/390 的 J01–J07、summary 和 30 页 visual 通过。
- 三视口共 12 保存组、24 个真实 tenants/catalog 暂停阶段均 save 未完成，释放后严格唯一行/列/状态正确；6 组 GET503 恢复和3组旧 scope/第二次相反保存边界通过。hold 转发真实响应，不伪造 catalog，不以 API/SQL 代 UI 写入。
- 冷起点 `node_modules_before=false`、`npm_cache_before=false`、`NODE_PATH=unset`，无手装依赖；Go1.25.6、Node v26.0.0、npm11.12.1、Playwright/core1.59.1，三个 probe 实测 Chromium147.0.7727.15。源码前后及候选保持，临时进程组与安装清理有 state 证据。
- QA8 自行完成 75/75 Node 契约、19 文件材料摘要、完整回执与源码复核，查看三宽度模型页截图；不是独立重跑动态 UI。report9 再核对材料及完整 log/state：项目 log 337036 字节，SHA256 `9197fff5ef1e77fd658f1a5a4c2a9e89b33598808f33e8708919967cee457c3f`。Connector 输出 337132 字节是另一口径，不补齐或混写。链接器 LC_DYSYMTAB 警告保留，不声称零警告。
- `TestConsoleClockRecoveryWorker` 顶层 skip 单列 Not Run；同轮父 InProcess/ActualHTTP 的真实子进程重启/备份恢复要求 `CLOCK_RECOVERY_PASS`，满足恢复门槛，不冒称顶层 worker 已运行。

## 已知问题与未验范围

| 范围 | 结论与边界 |
|---|---|
| R30-SYNC-01 | Closed，仅本次已证完成契约缺口；原层红转绿 |
| R28-04 | Open/P2；旧 tests18 严格断言的具体失败及根因仍未知，不能用本轮绿灯关闭或声称根治 |
| 持续最终产品错误 | 本轮无新 P0/P1 证据，不等于证明不存在；若刷新后仍持续错状态/缺列/重复行，按核心用户影响升级并保留原错误 |
| 独立 native 动态 QA | Blocked；未扩大监听权限，未用静态浏览器或 VM 冒充真实 Go UI |
| 真实供应商 | Not Run；所有宿主 upstream 均为显式隔离受控 HTTP，不是供应商联调或费用真实性验收 |
| Owner 实际已登录预览 | Not Run；宿主受控自动化不是 Owner 独立操作 |
| 新精确提交、合并、正式 prepare/部署 | report9 时 Not Run；旧 PR29 的 Deployment/health/schema3 仅为基线，不是本候选部署证据 |

冻结材料仍是 `design-v0.2.0` / `frozen-for-test-implementation` / `frontend-design-prototype`，archive SHA256 `5f51d35fdc15ed7214749f9413879b4238097f1d6b6bbe2cc0d508d8e103cef7`。实际引用同一回执下 `docs/prd.md`、`docs/interactions.md`、`docs/acceptance.md` 与原型资源，具体路径及各文件摘要见 QA8 材料检查。未将其改称重新批准，不执行材料包脚本。

## 仓库证据

以下为实际仓库与任务分支的完整 URL，ref 中的 `/` 已编码。**report9 尚未提交或推送这些新增文件，当前不声称远端可访问**；publish Connector 必须将它们随候选推送并核对可访问后再原样发布正文。不要改链到旧 main 或旧提交。

- [公共总览](https://github.com/big91987/model-relay/blob/workflow%2F2574d6be275f9eb630f8738837507e1e/docs/workflow/runs/2574d6be275f9eb630f8738837507e1e/README.md)、[T001](https://github.com/big91987/model-relay/blob/workflow%2F2574d6be275f9eb630f8738837507e1e/docs/workflow/runs/2574d6be275f9eb630f8738837507e1e/tasks/T001-model-save-readback.md)、[M01](https://github.com/big91987/model-relay/blob/workflow%2F2574d6be275f9eb630f8738837507e1e/docs/workflow/runs/2574d6be275f9eb630f8738837507e1e/milestones/M01-model-save-readback.md)
- [QA8 报告](https://github.com/big91987/model-relay/blob/workflow%2F2574d6be275f9eb630f8738837507e1e/docs/workflow/runs/2574d6be275f9eb630f8738837507e1e/acceptance/report-qa-08.md)、[逐项验收矩阵](https://github.com/big91987/model-relay/blob/workflow%2F2574d6be275f9eb630f8738837507e1e/docs/workflow/runs/2574d6be275f9eb630f8738837507e1e/acceptance/acceptance-matrix-qa-08.md)、[缺陷与限制](https://github.com/big91987/model-relay/blob/workflow%2F2574d6be275f9eb630f8738837507e1e/docs/workflow/runs/2574d6be275f9eb630f8738837507e1e/acceptance/defects-qa-08.md)
- [完整日志](https://github.com/big91987/model-relay/blob/workflow%2F2574d6be275f9eb630f8738837507e1e/docs/workflow/runs/2574d6be275f9eb630f8738837507e1e/acceptance/evidence/qa-08/host-tests-07/verify.log)、[state](https://github.com/big91987/model-relay/blob/workflow%2F2574d6be275f9eb630f8738837507e1e/docs/workflow/runs/2574d6be275f9eb630f8738837507e1e/acceptance/evidence/qa-08/host-tests-07/state.json)、[源码终态](https://github.com/big91987/model-relay/blob/workflow%2F2574d6be275f9eb630f8738837507e1e/docs/workflow/runs/2574d6be275f9eb630f8738837507e1e/acceptance/evidence/qa-08/source-final.json)
- [证据索引（包含三视口实际JSON及截图）](https://github.com/big91987/model-relay/blob/workflow%2F2574d6be275f9eb630f8738837507e1e/docs/workflow/runs/2574d6be275f9eb630f8738837507e1e/acceptance/evidence/index.md)、[原字节归档清单](https://github.com/big91987/model-relay/blob/workflow%2F2574d6be275f9eb630f8738837507e1e/docs/workflow/runs/2574d6be275f9eb630f8738837507e1e/acceptance/evidence/qa-08/archive-index.json)、[材料摘要与引用路径](https://github.com/big91987/model-relay/blob/workflow%2F2574d6be275f9eb630f8738837507e1e/docs/workflow/runs/2574d6be275f9eb630f8738837507e1e/acceptance/evidence/qa-08/material-check.json)
- [修前完整失败日志](https://github.com/big91987/model-relay/blob/workflow%2F2574d6be275f9eb630f8738837507e1e/docs/workflow/runs/2574d6be275f9eb630f8738837507e1e/verification/host-tests-05/verify.log)、[修前原层测量](https://github.com/big91987/model-relay/blob/workflow%2F2574d6be275f9eb630f8738837507e1e/docs/workflow/runs/2574d6be275f9eb630f8738837507e1e/verification/host-tests-05/model-save-readback-1280.json)、[原始研发归因](https://github.com/big91987/model-relay/blob/workflow%2F2574d6be275f9eb630f8738837507e1e/docs/workflow/runs/2574d6be275f9eb630f8738837507e1e/verification/development-06.md)
- [旧 R28-04 评估（已存在的基线提交）](https://github.com/big91987/model-relay/blob/9bf6ecaf53758ca2258498968b830c535e02e1c0/docs/workflow/runs/3e7496815b9922704072482cf44e73e2/acceptance/summary-assessment-qa-21.md)

## 后续授权链与发布门槛

交 publish Connector 将精确已验快照转为新 commit 并推送，随后按既有用户授权继续新 PR、精确提交核证/合并与 Owner 正式 preview 流程，不以草稿为交付终点。本报告节点不执行 Git 写入、合并、部署或 GitHub Workflow。

协调者仍需核证新 main 的正式 prepare、Owner 部署、Deployment、health200/version 等于精确新提交/storage_schema3，以及实际已登录用户入口；保持已有账号数据和 master key，不手改 DB/回执。任何同版门禁红或源码失配须 No-Go 返回 development，验收结论缺失/矛盾须回 QA；保留首个失败，不无限全套重跑。最终发布/部署状态须以新的真实回执更新公共事实，不能沿用本候选证据声称全面验收。
