# seq23 上游重读完成边界返工

## 对象与真实失败

2026-10-06恢复[PRD](../../product-definition.md)、[架构](../../console-architecture.md)、[交互](../../console-interaction.md)、[QA整改](../../acceptance/remediation-01.md)、公共任务与[前次返工](host-return-07.md)，不重建方案、不关闭QA缺陷。

平台seq23原样 `make verify` 退出2，HEAD `d62cd51e8de2127548e1ac7469d349443d1bfe3d`，56文件指纹 `7e9d396e8fa48ff4386fe89c9e81f44068d6b7dcb368bb57ddc31a17f10dc826`。当前Go/race/HTTP通过（relay 112.345s）、build通过、1280px运行版本通过。以下为会话回执摘录，非完整源日志：

```text
FAIL actual Go service browser journey: viewport=1280
checkpoint="visible #error contains 上游地址不符合"
location="journey.mjs:15:56", error="TimeoutError"
passwordProbe=null, keyboardProbe=null, clipboardProbe=null
最近请求：PUT upstream 400；PUT upstream 400；GET upstream 200；
PUT upstream 200；GET upstream 200；PUT upstream 200；GET upstream 503；
GET upstream 200；GET upstream 200；PUT upstream 200；GET upstream 200
make[1]: *** [browser] Error 1
make: *** [verify] Error 2
```

history旧版预期红测不是本次产品失败；390px异常旅程没有本次通过回执，不复用seq18旧绿色。

## 定位与最小变更

`refresh-upstream.onclick → guarded → load('upstream')` 在GET完成后恢复公开表单状态，然后显示重读notice，finally恢复按钮。旧旅程只等待已存在的“凭据已配置”，不能证明本次GET完成；最后重读之后立即fill无效地址，迟到GET可以覆盖草稿，使PUT合法返回200。回执GET200→PUT200→GET200与该路径一致；这是源码定位的测试同步缺口，不宣称新宿主已证明修复。

只修改项目旅程、VM回归、Trellis验证规范与公共交付事实，不改产品、地址策略/协议、数据库、需求/设计、CSS、Makefile/锁文件或平台。

- `keyboardReloadUpstream(page)` 替换三个重复等待：先注册真实GET，明确测试route屏障延迟GET，原生Tab/Enter激活并要求pending按钮disabled；释放实际Go响应后要求200、重读专属notice、按钮恢复、持久地址和此前保存失败已清凭据，再允许后续输入。旧文案或重复notice不能独立满足完成屏障。
- 后续无效地址保存增加真实PUT400断言；原中文错误、notice为空、持久恢复断言全保留，无补偿点击/睡眠/API或SQL业务预填。
- 新VM执行实际app.js，证明旧credential状态在GET pending时仍存在而按钮未恢复；GET-only重读完成后再输入无效地址，核对PUT载荷与400中文反馈。
- 首次新VM错误期望“重读清任意新凭据”而失败；重读源码只更新公开状态。保留[初次失败](upstream-reread-vm-initial.log)，夹具改为此前保存失败已清凭据的实际起点，不改产品清密行为或既有断言。

## 验证与退出条件

| 检查 | 实际结果 | 证据与边界 |
|---|---|---|
| 实际app.js VM | 75/75，退出0 | [日志](upstream-reread-vm.log)，非原生/网络验收 |
| JS语法、Go格式/模块/vet/build、diff | 退出0 | [日志](checks-upstream-reread.log) |
| 新对象指纹 | 56文件、HEAD仍d62cd51、退出0 | [指纹](source-upstream-reread.json)：`069d3fa7b5e165366fcb33eedab7e234c2595c58acc88201ee8ef74cd1a920fe` |
| 新原生两屏/完整make verify | Not Run | 稳定沙箱监听/Chromium限制未改变，保留原受限日志，不重复必然受限命令；新对象交宿主 |
| 独立QA/供应商/PR/合并/部署 | 无新增通过回执 | QA No-Go及QA-03驱动差异保留，供应商Not Run；路线2–4未实现 |

Trellis before-dev：packages发现成功，0.6.15初始化与真实规范存在，bootstrap不适用；读取frontend三规范及共享guide。trellis-check核对作用域、原生动作、响应到UI数据流、保密、回归与分项结果；update-spec记录重读完成契约。公共交付审查保留编号/依赖/关键路径，T002/T003 In Review、M02/M03 Ready for Review，无Done/Accepted。历次失败和QA资产不改写，无用户决策阻塞。

## 交接

宿主原样 `make verify` 核对新指纹、实际退出码与脱敏失败首尾。GET屏障及PUT400须在真实Go两屏执行，不以VM替代原生通过。失败正式返工，通过后独立QA复验18AC及整改原层场景，缺陷由QA判断关闭。关键路径为新宿主验证→必要返工→独立QA→后续发布Connector，不宣称供应商或发布通过。
