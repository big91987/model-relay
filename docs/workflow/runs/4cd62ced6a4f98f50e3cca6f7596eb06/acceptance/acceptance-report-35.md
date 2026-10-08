# design-v0.2.0 / reconciliation-ledger-r1 独立验收报告

> 日期：2026-10-08（+08）；qa执行35，Run4cd62ced6a4f98f50e3cca6f7596eb06，Issue25，workflow revision1。
> 仓库/分支：Model Relay / workflow/4cd62ced6a4f98f50e3cca6f7596eb06。
> HEAD：98511771871cf0951ecef716bbb55deb13e18da5；未提交/未跟踪实现包含在124文件指纹50fb764aac5443687e376c18906b36b765f02e9b05788fa4d5ac73617cbc98a4中。
> 环境：宿主临时真实Go服务、SQLite schema3、受控HTTP/SSE及Chromium；不是发布环境。Go1.25.6 / Node26.0.0 / npm11.12.1 / Playwright锁1.59.1。
> 模式：全量验收证据审查与独立非监听复验；上轮存在覆盖缺口，不能只短轮看J06。
> 结论：**No-Go，返回development。固定门禁及连续旅程主链通过，但安全读回失败后的恢复存在P1缺陷，强制故障覆盖仍不完整。**

## 1. 结论摘要

宿主seq34完整make verify exit0且94335字节归档未截断；QA独立算得同材料/同工作树指纹，实际检查三个continuous JSON和15张脱敏截图的来源与摘要。原J01–J07主链在1280/1440/390均完成，J06追加账本累计断言本轮已绿，QA22-02列表状态与摘要问题可以关闭。进一步对照AC08，发现写入丢响应后若第一次安全读回失败，页面把重试按钮删掉，又没有展示可保留的操作标识；实际生产处理函数的QA契约复现exit1。该缺陷会使签发/轮换结果无法在当前页面安全恢复，P1阻断；不是供应商缺凭据或沙箱监听问题。真实网页该故障分支、快速双提交以及多Key独立限额的缺用例必须由研发补入原门禁再复验，不能把21个绿色旅程标签提升为全部AC/H Go。

## 2. 基线、来源和执行责任

材料入口为`.workflow-input/4cd62ced6a4f98f50e3cca6f7596eb06/5f51d35fdc15ed7214749f9413879b4238097f1d6b6bbe2cc0d508d8e103cef7/files/manifest.json`；版本design-v0.2.0，frozen-for-test-implementation；archive SHA256为5f51d35fdc15ed7214749f9413879b4238097f1d6b6bbe2cc0d508d8e103cef7。QA35重新核对archive及19文件bytes/SHA全部一致，具体相对路径/摘要在evidence/attempt-35/material-integrity.json。

实际引用该root下docs/prd.md（e768b4f2…）、interactions.md（95f3490e…）、acceptance.md（2792740b…）、decisions.md（faed344c…）、baseline-review.md及prototype文件；完整摘要见上述JSON。最终整改闭合优先于baseline-review旧“待确认”段；A、内部身份、USD测试价和内部多租户授权已确定，不重复审批。本Run architecture.md、implementation-contract.md、verification-design.md及migration-and-operations.md为实施契约；冻结原材料没有修改。当前用户明确升级范围优先于项目旧v0.1.0首任务说明。

本节点完整读取挂载Skill、acceptance-sop及artifact-contracts。没有访问旧Run工作区、New API源码/秘密或GitHub评论，没有切换指定模型。Codex/gpt-6.1-sol是用户指定值；没有可核对的原生模型执行回执，保持Not Run，不拿配置或自报当实际模型证明。

## 3. 本轮覆盖

- **宿主执行，不冒充QA执行**：seq34完整固定入口、165 Node、Go race/实际HTTP/进程/迁移、build及真实后端UI。读回offset/eof与边界详见evidence/attempt-35/host-seq34-review.md。
- **QA实际执行**：项目缓存来源戳，19文件材料校验、18件宿主产物复制校验；format-check、module verify、vet、165 Node、七项针对性Go race与build组合exit0。七项为GovernanceInProcess、ResourceDimensionsInProcess、PeopleScopeInProcess、PauseUsageInProcess、AccountingInProcess、GenerationExactDeadline、MicroDollarPrecision。实际日志qa-independent-checks.log保存起止时间与exit；受控transport/handler测试不声称真实HTTP。
- **QA新增失败证据**：readback-recovery.test.mjs抽取当前workspace.js实际edit函数，在明确DOM替身（textContent删除子节点语义）中复现“已提交丢响应→一次安全读回失败→按钮消失”；exit1，mutation1/read1/retry=false/operation-visible=false。类型Contract，不是假装Chromium复现或固定门禁失败。
- **连续UI**：同一人员角色的子旅程沿原浏览器context和scope完成；平台交付、邀请接受、租户/开发者/财务责任切换使用各自真实登录context，不是前端切角色。邀请令牌/一次秘密从实际响应读出用于显式手工交付/调用测试；管理步骤仍由页面提交，无中间API/SQL补业务。API仅只读事实对账；测试clock推进和受控读失败/迟到注入明确披露。详见user-journeys-35.md。
- **UI/数据分层**：J01空安装及校验恢复、J02交集拒绝恢复、J03代际/撤销/停用/到期、J04硬费用恢复、J05暂停与成员退出、J06改价/核对/结账/追加更正/导出、J07分页/迟到/读取重试已通过所列步骤。模型/账单/身份API与账本事实另列，不拼接缺失UI步骤。

## 4. 缺陷和整改状态

| ID | 等级 | 用户可见问题/复现 | 影响与定位 | 当前状态/回归资产 |
|---|---|---|---|---|
| QA35-01 | P1 | Key写入提交后丢响应→安全读回发生短暂失败→提示稍后重试，但重试按钮被移除，操作标识未显示 | 负责人无法从原页面确认签发/轮换事实；可能丢失新秘密或被迫冒险再操作。workspace.js:94，output.textContent覆盖了button子节点；operation仅在closure | Open；QA生产函数Contract真实红灯readback-red.log；真实浏览器故障路径Not Run，必须回原层验证后关闭 |
| QA22-01 | P1（覆盖主题） | 原七条连续主链现已补齐，但AC08丢响应+读回失败/轮换、AC16快速双提交及AC09双scope独立限额竞争未完整证明 | console.mjs:137只有签发丢响应后首次成功读回；journeys.mjs:267/rotation正常响应；call helper等待完成；资源测试Key只独立命中Calls=0 | Fixing / 部分闭合；不得把Not Run写成产品Fail。保留原主题编号，本轮不重报“七旅程均没跑” |
| QA22-02 | P2 | 原列表混淆启用/停用/过期、缺授权/期限/预算摘要 | J03真实页面分别断言过渡/停用/恢复/过期/撤销；截图显示授权、期限、已用/预留/上限；实际formatter契约绿 | Closed；seq34三视口J03、key-states截图及Node状态边界回归。截图最终已恢复主Key，因此停用态以实际UI动作断言而非该截图证明 |

P0未发现；不把未执行用例等同产品已经失败。QA35-01的Contract足以确定当前处理函数缺陷，但关单仍必须有真实UI-E2E恢复，不降低证据层级。没有需要需求/设计返工或用户澄清的歧义。

## 5. 当前证据索引

- acceptance-matrix-35.md：AC01–AC18/H01–H08分层状态、八类商用五层对照。
- user-journeys-35.md：J01–J07连续步骤、Actor/Start/最终事实/清理/结果及未测分支。
- remediation-handoff-35.md：具体修复、必须新增故障用例和同版宿主复验要求。
- evidence/attempt-35/source-before.json、source-after.json：QA前后来源；不覆盖根矩阵/设计文件，避免改变验过的124文件快照。
- evidence/attempt-35/material-integrity.json、artifact-index.json、collect.mjs：材料完整摘要、原产物关联和18件字节一致证据。
- evidence/attempt-35/host-seq34/：三JSON、每视口before-rotation/key-states/before-correction/after-correction/final截图；是宿主运行产物，不是QA再次操作。
- evidence/attempt-35/readback-recovery.test.mjs、readback-red.log：新增P1复现；qa-independent-checks.log另证明既有测试集绿。两者不可混为“全部QA绿”。

历史seq21/seq24/26/28/30/32与历次失败仍保留；previous-22保存修改索引前原报告/矩阵/旅程/整改包，旧结论不能跨指纹当当前Pass。截图不记录一次性秘密；真实上游凭据未使用。

## 6. 未执行与阶段边界

QA不重复已知必败的监听make verify，执行22的原生Blocked日志仍保留；同轮宿主seq34替代固定门禁执行，不把旧Blocked改成Pass。Static browser.check不适用于真实Go后端，本节点未使用。

AC08失败读回UI与轮换丢响应、AC09全部Key维度/多Key和代际争用、AC16真实网页双提交仍Not Run，属可完成内部验收而非外部凭据限制。其余范围以矩阵具体通过子场景为准，不宣称未经覆盖的并发/故障组合全通过。

真实供应商/真实已有推理端点、外部身份/支付均Not Run，不阻断受控内部产品修复。H02重复入站、H03实际模型、H04材料负向恢复、H07同PR合并/Actions/Deployment/health固定URL、H08唯一平台维护源升级原入口复验仍缺完整独立证据，归平台/协调者；不要求研发手改平台数据库或自发部署。本轮没有“已商用对齐”“已发布”声明，冻结外支付税务/GPU/多协议/生产SLO仍为后续差距。

## 7. 环境恢复

QA没有启动持久产品服务、调用用户上游、修改DB或改业务实现。独立Go测试使用临时store并由测试清理；DOM替身没有网络。宿主runner使用独立临时目录、随机端口、真实登录context、服务/store/browser关闭及目录删除；代码/回执支持该清理契约，未做全机进程审计，不补造全局清理证明。被忽略test-results产物已脱敏复制到本Run，源文件及失败均未删除。

## 8. 质量趋势

相对QA22，七条主链21个三视口检查点由未执行提升到真实通过，Key状态P2关闭，追加账本误读的原native路径已绿；165 Node与既有Go/迁移/HTTP/SSE/keyboard门槛保留，质量明显提高。新增发现一个P1读回失败恢复缺陷，说明只测首次读回成功仍会漏安全异常分支；覆盖主题QA22-01部分闭合但不能整体关单。没有证据支持旧暂停费用不守恒或keyboard唯一根因的夸大结论；seq32失败是驱动累计读取问题，不是要求改原账本。剩余风险集中于明确失败恢复、独立资源组合以及Harness/发布事实。

## 9. Go / No-Go与路由

**No-Go：QA35-01 Open P1；QA22-01指定强制用例未闭合。** seq34绿灯是有效工程事实，但不覆盖缺失AC，也不消除真实处理函数红灯。沿正式handoff返回development修复并补回归；不请求原型/运营方式再次批准，不扩大权限，不提交/推送/创建PR/部署。研发使用同Run/同材料，保持通过主链及历史断言，形成新指纹完整make verify真实回执后再独立QA。当前不进入report作为可发布产品；平台和发布责任按矩阵另行闭合。
