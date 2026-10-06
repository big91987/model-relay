# seq21 剪贴板能力前置条件返工

## 对象与真实失败

2026-10-06恢复[PRD](../../product-definition.md)、[交互](../../console-interaction.md)、[架构](../../console-architecture.md)、[QA整改包](../../acceptance/remediation-01.md)、公共任务和[上轮实施](qa-rework-01.md)。不重建需求/设计，不关闭QA缺陷。

平台seq21原样 `make verify` 退出2，HEAD `d62cd51e8de2127548e1ac7469d349443d1bfe3d`，56文件摘要 `6cdd533c1dd449f9aa28dcf1422228c399e94ed2a589175f465ff16839cd7a52`。当前Go/race/HTTP通过（relay 112.092s）、build通过、1280px运行版本旅程通过；随后第一组密钥复制反馈超时。health-history内旧版预期红测不是当前产品失败。以下为平台失败尾部摘录，非完整源日志：

```text
PASS runtime version actual health, refresh, failure/reentry, nonblocking and text/layout 1280px
FAIL actual Go service browser journey: viewport=1280
checkpoint="visible #notice contains 已复制，请保管"
location="journey.mjs:14:56", error="TimeoutError"
passwordProbe=null, keyboardProbe=null
最近请求：POST /admin/api/keys 201；GET /admin/api/keys 200
exit status 1
make[1]: *** [browser] Error 1
make: *** [verify] Error 2
```

回执没有剪贴板权限或实际error文案，不能确证OS拒绝或产品根因。源码确定的是旅程前置缺口：`newContext`无剪贴板权限设置却硬要求成功。产品 `copy-key.onclick` 已在writeText拒绝时给出中文手动路径且按owner/issuedEpoch隔离；不应修改产品以谎报成功。

## 最小变更与执行契约

仅改项目 `journey.mjs`、`ui.test.mjs`、Trellis验证规范和公共任务。已核对锁定Playwright安装的Chromium适配器支持clipboard-read/write映射。不改产品JS/HTML/CSS、Go协议/数据库、Makefile/锁文件、平台或系统权限；无新依赖，不放宽成功/键盘/保密断言。

`copyIssuedKey(page, context)` 在两屏每次签发后：
- 明确注入writeText的NotAllowedError，原生Tab/Enter激活，要求完整中文手动错误、保留遮罩一次展示和受控key、按钮可重试；finally恢复原API。这是API故障注入，非OS拒绝回执。
- 只对隔离Go服务origin授予context剪贴板读写能力，重新键盘激活，等待当前操作settled并保留原成功反馈门禁。
- 原生readText在页面内比较剪贴板与当前签发字段，只返回boolean，不返回key内容。finally清空受控剪贴板值并撤销context权限，不读取/恢复用户既有剪贴板历史。
- 失败诊断仅stage及secureContext/documentFocused/manualFallback/successNotice布尔值，无key/正文/截图。VM原隔离回归补拒绝→保值/可重试→成功，并保留关闭/跳页/晚到断言。

现有占位示例的真实成功或手动路径仍独立验证，不拿它替代密钥复制内容核对；没有修改原QA资产或增加替代验证入口。

## 当前证据与退出条件

| 检查 | 实际结果 | 证据/边界 |
|---|---|---|
| 真实app.js VM | 74/74，退出0 | [日志](clipboard-recovery-vm.log)，不是原生/网络验收 |
| JS语法、Go格式/vet/build、模块、diff | 退出0 | [日志](checks-clipboard-recovery.log) |
| 工作树版本 | 56文件、HEAD仍d62cd51、退出0 | [指纹](source-clipboard-recovery.json)：`7e9d396e8fa48ff4386fe89c9e81f44068d6b7dcb368bb57ddc31a17f10dc826` |
| 新原生剪贴板/完整两屏异常/分页 | Not Run | 稳定沙箱监听/Chromium限制未变，不重复必然受限完整命令，交宿主原样make verify |
| 独立QA/供应商/PR/合并/部署 | 无新增通过回执 | 原QA No-Go和QA-03旧泛导航driver失败保留；供应商Not Run；路线2–4未实现 |

Trellis before-dev：packages发现成功，存在0.6.15初始化及源码规范，bootstrap不适用；读取frontend三规范及复用/跨层guide。trellis-check核对作用域、原生/所有权、依赖、回归和实际分项；update-spec沉淀权限前置与故障注入/原生内容区别。公共交付审查复用T002/T003、M02/M03，依赖/范围不变，仍In Review/Ready for Review，不Done/Accepted。

## 宿主与独立QA交接

原样 `make verify` 核对新指纹、真实退出码和clipboardProbe脱敏尾部；失败正常返工，不用seq18旧绿色或74VM放行。通过后独立QA复核18AC及整改全部原层场景，包括旧QA-03驱动与明确证据动作差异；研发不擅删断言/关缺陷。自动复制成功必须证明实际内容相等，权限拒绝必须提供保值手动恢复，不能只接受任选一个notice来改绿。
