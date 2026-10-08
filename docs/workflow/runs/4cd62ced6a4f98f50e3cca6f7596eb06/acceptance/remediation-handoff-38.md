# QA38 补证交接（development）

## 当前已验证版本与不应重复修的内容

design-v0.2.0/archive5f51d35fdc15ed7214749f9413879b4238097f1d6b6bbe2cc0d508d8e103cef7；manifest沿用prepare回执。HEAD98511771871cf0951ecef716bbb55deb13e18da5，127文件工作树3ee2122eadac33a67342523a082f791d8a81154cc86b7321d53de856866a15ef。seq37原完整make verify exit0/101748 bytes/truncated=false，QA source-before/after一致、19件材料一致、63件宿主产物已归档。QA38自己169 Node、七项Go race及format/modules/vet/build实际通过；没有重复沙箱必败监听。

QA35-01当前editor的原层故障三视口已Closed；QA22-01原J01–J07及AC08/09/16缺用例范围Closed；QA22-02保持Closed。不要因为此轮No-Go重做身份/预算/计费或再改原账本。当前没有新已执行产品P0/P1或需求/设计歧义。

## QA38-01：完成已承诺的AC01实际设计对照

依据冻结材料docs/acceptance.md AC01、docs/interactions.md「页面及行为」及本Run verification-design.md§2。QA35矩阵AC01-UI-full-review明确Not Run；development36的目标修复及seq37绿色没有覆盖该行。当前console.mjs:140十页循环仅h1和根宽度，:142截图已回Key；63件截图集中Key/账单，不能供独立QA核对全部批准页面。

最小必要动作：

1. 在既有真实Go+Chromium入口中，为十页分别留下**实际page/role/scope与非空或适用空态**的证据，优先复用现有J/C执行对象，不以HTTP预制完成态代替首次旅程。1280/1440/390沿既有视口，保存当前页面和关键编辑/详情的脱敏截图及字段/操作断言；不要求所有失败态组合都截图。
2. 按材料页面表整理“批准主要信息/动作→实际入口/截图/断言”的逐页映射，明确合理装饰简化、真实无样本/未知口径和权限下不可见，不把字段存在于源码/接口JSON视作用户已能在正式页面读懂。具体尚需核对页面见acceptance/visual-review-38.md。
3. 模型/服务/预算/分析/成员/日志/调用页应有当前实际页面证据；若发现重要信息/动作确实缺失，按冻结设计最小修复并补原层测试。不借此扩大高级协议/供应商/SLO或商业支付范围；不私自改冻结材料来适配实现。
4. 每张图等真实请求/界面落定、截图前敏感输入清除；页面详情截图不能在drawer出现前截图。不收Key/邀请/上游token/完整对话/密码/数据库/绝对本机路径。既有截图不能改造或标成新运行。
5. 若已有同版真实产物与断言可满足，交接精确仓库相对路径及生成来源即可，不必为了流程重跑已完成副作用或重写业务；目前纳入索引63件/h1烟测未达到该门槛。

验收语句：**独立QA可以在同一127文件候选（或新的已验证指纹）上逐页对照十页批准主要信息/动作和三视口实际可用页面；UI结果、API/数据事实分层，差异有可核对依据，缺页/缺真实操作仍Not Run，不以导航标题勾AC01。**

这是强制门槛未验证，不是假定产品运行失败，也不是P2色值/装饰建议阻断。返回研发是补可复验资产与必要最小实现，不要求其承担QA最终判断。

## 必须保留的绿色/红色证据与复验规则

- J01–J07三视口21检查点、原13Tab/Escape/触发点、实际HTTP/SSE/取消/重启/迁移/backup及健康历史门禁不减少。
- F签发/已用Key轮换POST成功丢响应→GET503/404/200、native双提交mutation1、未知锁/reference/retry/epoch与秘密清理；旧editor重建红必须继续标“修复后隔离重建”，保留QA35真实Contract红，不能追溯伪称旧Chromium事实。
- ResourceMatrix tenant/Key各Calls/RPM/TPM/C/budget独立0/finite/null、拒绝双scope原子性；multi-Key/双代8争2、另一tenant可用与原历史usage不清零。
- 普通/SSE双发送各proxy POST/准入/attempt/实际dispatch1、Tokens3+4；原J06 fee21/cost7/entries2/旧revision0及快照、原Governance19/60及delta5/16保持，不改DB造绿。
- 不改Makefile门禁或为浏览器方便放宽后端权限/计费；新增验证源码/测试后重新计算工作树并走原完整宿主make verify，保留exit/未截断全日志/产物/指纹再交独立QA。仅新增同版文档不伪称代码已升级。
- 沙箱限制未变无需再次监听或广泛提权。静态browser.check不能验证实际Go后台。保留历次失败、当前63件来源和自己的采集错误说明。

## 责任边界及交接输入

当前不需要用户澄清/requirements/design返工。H02重复入站、H03实际原生模型、H04材料负向恢复、H07同PR/Actions/Deployment/health固定URL、H08唯一平台源安装升级由平台/协调者提供证据；不是本次产品研发补证任务，不手改平台/触发Workflow/部署。真实供应商/外部身份支付Not Run，不阻断内部页面补证。

实际输入：acceptance-report-38.md、acceptance-matrix-38.md、user-journeys-38.md、visual-review-38.md与evidence/attempt-38/artifact-index.json、source-before/after.json、material-integrity.json、host-seq37-review.md。旧QA35整改及development36记录继续作为历史，不覆盖。完成补证且独立QA强制门槛闭合后才进入report；目前不报整套Go/已发布。
