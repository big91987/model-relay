# 登录页支持显示和隐藏管理员密码

Closes #12

管理员可在登录前核对密码；默认隐藏，切换保留原输入值，不提交表单或发送请求。此 PR 是登录页的局部增量，不改变后端认证、数据库、健康接口或版本协议，不增加运行依赖。

## 实现与回归

- 原生 `type="button"` 控件在“显示密码／隐藏密码”之间切换，支持 Tab 聚焦及 Enter、Space 激活；动态动作名称和 live 状态说明分别表达下一步动作与当前显隐状态。
- 以原密码输入节点的 `type` 为唯一可见性事实源，不复制或持久化密码，不新增凭据日志。空值、特殊字符和多次切换均保持输入值。
- 成功登录、退出和既有秘密清理入口统一恢复隐藏并清空密码；沿用现有会话恢复与 401 拒绝处理。
- 业务修改仅涉及 `internal/relay/web/index.html`、`app.js` 和登录区 `style.css`；测试扩展位于 `tests/browser/ui.test.mjs` 与 `journey.mjs`。
- VM 回归覆盖状态、保值、零请求及秘密生命周期；真实 Go 服务浏览器旅程扩展原生键盘、可访问语义、零 submit/invalid/network、错误重试、隐藏态/显示态登录、退出和会话恢复。既有管理台、健康方法和版本显示回归继续由固定门禁执行。
- 研发依据现有源码完成 8 份项目级 Trellis 规范 bootstrap，并实际执行 before-dev、update-spec 和 check；具体读取、执行和裁剪记录见研发验证文件。平台负责调度和 Git 发布，没有启动第二套调度器。

## 本轮验证与独立 QA

验收日期：2026-10-06。独立 QA 结论为 **Go with known issues**，AC01—AC13 有同轮证据，无未关闭 P0/P1；未发现 P0/P1/P2 产品缺陷。

| 证据来源 | 真实结果与边界 |
| --- | --- |
| 平台固定宿主 Connector（seq7） | `make verify` 退出 0。未修改串行 fail-fast 门禁，包含格式、模块校验、vet、前端测试、Go race/真实 HTTP、build 及启动实际 Go 服务的浏览器旅程。 |
| QA 本地固定门禁 | `make verify` 退出 2：health-history 旧版隔离重建遭沙箱监听限制。该次记为 Blocked，后续完整 HTTP/浏览器阶段为 Not Run，不冒充本地通过。 |
| QA 独立补充检查 | 格式、vet、模块校验、build、27 项前端 VM、race 健康/版本 handler 和 journey 语法检查通过；不替代真实 HTTP 或浏览器证据。 |
| 报告节点复核 | 仅运行源码身份、历史哈希与差异检查，未重新执行完整门禁。源码摘要仍匹配宿主/QA；10 份历史文档和 62 份历史日志全部 OK，历史路径无差异，`git diff --check` 通过。 |

验证基线 HEAD：`73681a7df4c5cbe5914056ac13348154d58726e0`；最终未提交工作树摘要：`0f6ee2d6894067323697f632231ee0d62b098be0261a4da7e975edb6e1efb37b`（36 文件的项目 evidence 范围，不是整个仓库或最终发布提交的摘要）。报告只新增本文，未改变已验源码。发布提交由后续 Connector 生成；业务、测试或锁发生变化必须重新固定验证和独立 QA。

宿主回执输出在旧版健康重建处截断：总体退出 0、源码身份和未修改的门禁/旅程契约支撑通过结论，但没有完整当前 HTTP/浏览器尾部或截图，不能声称 QA 亲自跑通浏览器或提供逐项原始输出。旧版健康拒绝方法重建中的退出 1 是门禁预期的历史失败，不是当前产品失败。

## 已知问题与交付边界

- **P3 QA12-HISTORY-WRITE（Open）**：既有 `tests/health-history/main.go` 会重写旧 `docs/workflow/health-methods-history.log`。本轮宿主与 QA 生成内容已分别留存到本 Run，再恢复旧路径并核验全部 72 份历史文件；当前现场已恢复，后续重跑仍需先保护、保存新输出、恢复并核验旧基线。未削弱门禁，未把该问题虚构关闭或标为 Accepted Risk。
- 研发首次沙箱失败、后续补充检查和 QA 失败均保留；旧 Issue #8/#9 文档、QA 和日志只作为保护对象/背景，不是本轮通过证据。
- 浏览器回归使用隔离临时数据、随机端口和明确标记的受控测试 upstream，不代表真实供应商联调。未部署、未调用真实供应商、未扩权；未验证屏幕阅读器音频、设备扩展或密码管理器兼容性。
- 本文是交付说明，不代表已经发布、合并或完成用户最终业务验收；Issue 回写、提交推送与草稿 PR 由平台 Connector/Hook 负责。

## 本轮产物与证据入口

- [需求与 AC 基线](docs/workflow/runs/836b817833886b3188f263800b0f092f/product-definition.md)、[G1 评审](docs/workflow/runs/836b817833886b3188f263800b0f092f/g1-review.md)。
- [轻量架构契约](docs/workflow/runs/836b817833886b3188f263800b0f092f/password-visibility-architecture.md)、[G2 评审](docs/workflow/runs/836b817833886b3188f263800b0f092f/g2-review.md)。
- [研发交付索引](docs/workflow/runs/836b817833886b3188f263800b0f092f/README.md)、[M01 里程碑](docs/workflow/runs/836b817833886b3188f263800b0f092f/milestones/M01-password-visibility.md)、[T001 任务](docs/workflow/runs/836b817833886b3188f263800b0f092f/tasks/T001-password-visibility.md)、[研发验证与 Trellis 执行记录](docs/workflow/runs/836b817833886b3188f263800b0f092f/verification/M01-development.md)。
- [独立 QA 报告](docs/workflow/runs/836b817833886b3188f263800b0f092f/acceptance/acceptance-report.md)、[分层验收矩阵](docs/workflow/runs/836b817833886b3188f263800b0f092f/acceptance/acceptance-matrix.md)、[核心旅程契约](docs/workflow/runs/836b817833886b3188f263800b0f092f/acceptance/journeys.md)、[缺陷与环境限制](docs/workflow/runs/836b817833886b3188f263800b0f092f/acceptance/defects.md)。
- [同轮宿主回执核验](docs/workflow/runs/836b817833886b3188f263800b0f092f/acceptance/evidence/host-connector-receipt.md)、[QA 本地失败](docs/workflow/runs/836b817833886b3188f263800b0f092f/acceptance/evidence/qa-make-verify.log)、[QA 补充检查](docs/workflow/runs/836b817833886b3188f263800b0f092f/acceptance/evidence/qa-non-listening-checks.log)、[最终身份与历史保护](docs/workflow/runs/836b817833886b3188f263800b0f092f/acceptance/evidence/final-integrity.log)。
