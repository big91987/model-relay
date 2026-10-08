# design-v0.2.0 / safe-readback-resource-matrix-r1 独立验收报告

> 日期：2026-10-08（+08）；qa执行38，Run4cd62ced6a4f98f50e3cca6f7596eb06，Issue25，workflow revision1。
> 仓库/分支/Commit：Model Relay / workflow/4cd62ced6a4f98f50e3cca6f7596eb06 / HEAD98511771871cf0951ecef716bbb55deb13e18da5。
> 当前未提交候选：127文件工作树3ee2122eadac33a67342523a082f791d8a81154cc86b7321d53de856866a15ef，未把main HEAD当候选实现已提交或发布。
> 环境：宿主隔离真实Go服务、SQLite schema3、Chromium及受控HTTP/SSE；Go1.25.6 / Node26.0.0 / npm11.12.1 / Playwright锁1.59.1。不是生产部署。
> 模式：全量验收证据审查与独立非监听复验；上轮全量覆盖仍有遗漏，不仅看新增安全读回。
> 结论：**No-Go，返回development补强制AC01实际页面设计对照证据。既有产品P1读回缺陷及指定故障/资源覆盖均已复验关闭，固定make verify同版通过；不能将这些事实写成整套AC/H已Go或已发布。**

## 1. 结论摘要

宿主seq37完整make verify真实exit0，101748字节归档truncated=false；QA独立源戳与该回执一致，重新核对archive及19文件、归档63件当前产物。QA35-01已回到真实Go+Chromium的签发/已用Key轮换故障路径，在三视口验证提交丢响应后GET503/404/200恢复、mutation1、安全标识/retry保留及秘密不重放，可以关闭。原QA22-01指定残余AC08/09/16覆盖本轮补齐：两scope五维边界、multi-Key/旧新代8争2、另一租户独立，普通/SSE真实用户双发送各dispatch1；原J01–J07共21主链检查点仍通过。QA22-02列表状态问题保持Closed。尚未完成的AC01完整批准原型逐页实际对照在QA35矩阵已是Not Run，本轮63件图仍主要覆盖Key/账单，十页导航只检查h1及根宽度，不能替代设计对照；这是一项强制证据门槛缺口，而非新增未经执行的产品运行失败。

## 2. 权威基线与版本一致性

材料仍为design-v0.2.0 / frozen-for-test-implementation；archive SHA2565f51d35fdc15ed7214749f9413879b4238097f1d6b6bbe2cc0d508d8e103cef7。实际manifest为`.workflow-input/4cd62ced6a4f98f50e3cca6f7596eb06/5f51d35fdc15ed7214749f9413879b4238097f1d6b6bbe2cc0d508d8e103cef7/files/manifest.json`。evidence/attempt-38/material-integrity.json保存实际引用路径/逐件完整摘要/bytes；19件与archive均一致，没有执行包内脚本或改材料。

承接材料docs/prd.md（Key/限额/账本规则）、docs/interactions.md（十页A与J01–J07/状态/抽屉）、docs/acceptance.md（AC01–18/H01–08）、docs/decisions.md和最后整改后的baseline-review；A/内部身份/USD测试计价及用户原型授权保持有效。本Run architecture.md、implementation-contract.md、verification-design.md、migration-and-operations.md、development36交接及QA35报告/矩阵/整改契约是实施与复验来源。本次明确v0.2任务优先于AGENTS旧v0.1首任务说明，不重开需求或设计批准。

已完整读取挂载验收Skill、SOP及产物契约。没有访问其他Run工作区、New API实现/秘密、GitHub评论或改平台数据库。用户指定Codex/gpt-6.1-sol；本节点没有可核对的原生模型执行回执，H03-runtime保留Not Run，不以配置或自报充当证据。

127文件指纹覆盖产品/测试/依赖锁/项目契约/Trellis规范；QA前后source-before/after.json一致。验收产物/回执不在该源戳范围，避免自引用；这不等于整棵Git所有文件的哈希。受控测试服务的runtime标签与schema3不是正式版本部署证明，浏览器JSON无独立签名版戳，来源关联限制在artifact-index.json明确保存。

## 3. 本轮覆盖和真实执行责任

- **宿主tests执行37**：完整固定入口含format、modules、vet、169 Node、Go race/实际HTTP/进程/备份迁移、build及真实Go+Chromium。日志从平台read_command_output读至eof；health-history旧版红为明确预期隔离重建，当前健康测试为绿。详细边界见evidence/attempt-38/host-seq37-review.md。
- **QA自己实际执行**：19件材料摘要/63件源与副本校验；Node169项全部通过、0skip；七项顶层Go race为GovernanceInProcess、GenerationExactDeadline、PeopleScopeInProcess、PauseUsageInProcess、ResourceMatrixInProcess、SharedResourcesInProcess、MicroDollarPrecision，包48.879s exit0。InProcess不声称网络事实。带set -e的format/modules/vet/build组合真实exit0。实际日志分别qa-node.log、qa-go.log、qa-static-final.log。
- **QA独立审查**：当前生产editor的unknown锁/reference/status/retry、GET忙去重和epoch/dialog-open迟到隔离；fault脚本真实POST route.fetch后abort、GET503/404再真实200、浏览器native双提交、Key数量/代际/usage/价快照守恒；资源fixture/HTTP客户端/barrier/另一租户与两scope最终事实。代码仅增强实际回执，不代替缺少UI路径。
- **连续UI主链**：J01空安装/非法修正/首次授权调用/真实邀请；J02授权收窄拒绝恢复；J03轮换/启停/撤销/期限/历史；J04硬预算拒绝修正；J05暂停/另一租户/成员退出；J06旧新价/pending/核对/结账/更正/导出；J07搜索分页/同ID/迟到/重试/退出。在1280/1440/390同角色子旅程用同context/scope；角色责任切换使用各自真实登录context，不是前端角色切换。API仅只读最终事实或显式准备，不代办中途页面步骤；测试clock与故障注入明确标记。user-journeys-38.md给出Actor/Start/UI结果/事实/清理。
- **新增原路径故障**：三视口签发/轮换提交成功丢响应，安全读503→404→200后仍只一次写；轮换usage/历史请求/价格快照不变，叠加409不改变对象。普通/SSE各一次准入/attempt/真实上游dispatch和Tokens3+4。修复后重建旧editor的浏览器红明确标记，不能冒充旧开发阶段的原始Chromium红灯；原QA35生产handler Contract红仍保留。
- **界面设计对照**：已人工查看当前1280Key/签发恢复、390轮换503、1440账单更正与批准Key原型；确有A框架、统一dialog/drawer、期限/用量/状态/真实USD口径。其余页面只有代码/部分旅程或h1烟测，没有本轮完整实际设计对照，不能以这些底层绿色拼成AC01完整Pass。

## 4. 缺陷、门槛及进展

| ID | 等级/类别 | 用户影响与复现 | 当前状态 | 关闭/补证依据 |
|---|---|---|---|---|
| QA35-01 | P1产品缺陷 | 提交已成功丢响应后第一次安全读回失败，原页面丢retry及operation标识，不能安全恢复 | Closed | 当前asset三视口签发及已用Key轮换GET503/404/200；mutation1/不回放secret/原历史价快照usage不变，Node原handler契约同时绿 |
| QA22-01 | P1覆盖主题 | 原主链及AC08故障/AC09两scope独立资源竞争/AC16双发送缺用例 | Closed（原主题列明范围） | J01–J07本轮21检查点，F三视口故障和双发送，ResourceMatrix/SharedResources ActualHTTP；不是把所有可能组合都宣称测完 |
| QA22-02 | P2产品缺陷 | 原Key列表状态和授权/期限/预算摘要不准确 | Closed（保持） | J03真实启停/过期/撤销/过渡动作及当前key-states截图/Node状态边界继续通过 |
| QA38-01 | 强制验证门槛缺口，不冒称产品P1 | 十页批准设计的实际页面信息与操作尚不能逐页核对；旧AC01-UI-full-review仍Not Run | Open / owner development测试资产 | visual-review-38.md逐页说明：console.mjs:140仅h1/宽度，:142最终为Key图；当前图集中Key/账单，不支持十页设计对照 |

本轮未发现新的已执行产品P0/P1，未把P2视觉装饰建议升级阻断。No-Go来自强制AC01证据缺失，不是“最好再多测点”。若研发能提供已有同版十页真实证据，应精确交接而非重复实现；否则在原真实Go浏览器入口补必要页面和关键编辑/详情断言及脱敏截图。本轮没有需求/架构歧义，不返回requirements/design，不需要用户再确认授权。

## 5. 证据与可复用资产

- acceptance-matrix-38.md：AC01–18/H01–08分层结果、八类功能五层对照；没有把局部/API/宿主/独立QA/发布混同。
- user-journeys-38.md、visual-review-38.md：连续旅程及实际设计对照范围/门槛。
- remediation-handoff-38.md：下一轮唯一产品侧补证责任、既有门槛及新指纹复验要求。
- evidence/attempt-38/source-before.json、source-after.json、material-integrity.json、artifact-index.json、collect.mjs、collect-final.log：版本/材料/当前63件产物来源。
- evidence/attempt-38/host-seq37/：三个fault JSON/39故障PNG、三个continuous JSON/15旅程PNG、三个console PNG；平台seq37完整日志另由read_command_output取得，不伪造本地完整日志。
- evidence/attempt-38/qa-node.log、qa-go.log、qa-static-final.log、host-seq37-review.md、collector-attempt.md：QA实际运行与宿主审查分别记录。

首次采集器错误引用不存在的console-first-call-1440.png得到exit1；这是QA清单假设错误，不是产品失败。已删旧native额外引用并按现存新控制台三视口重收exit0；首次已复制的旧1280首调用图不在63件索引，不作为当前设计对照证据。qa-static.log首次组合未开启errexit，最终用带set -e的新日志，旧文件不删除。此前所有开发/QA/宿主失败和QA35原红均保留。

qa-go.log链接器警告中的一处本机临时目录前缀已替换为`<temp>`，警告及全部测试结果未改；log-redaction.json记录原始/脱敏bytes与SHA256及唯一替换规则。原始日志只保存在忽略缓存，不列交付artifact，不提交本机绝对路径。

## 6. 未执行、环境限制与阶段边界

AC01完整批准原型实际设计对照仍Not Run；不是供应商凭据问题，当前内部可完成且是本轮返研目标。self_hosted API来源/核验有当前受控证明；首次self_hosted来源选择的UI独立分支与真实供应商联调不擅自写Pass。

QA原生make verify的已知监听Blocked保持QA22原证据，本轮按项目约定使用同版本宿主seq37完整入口替代，没有重跑必败命令、沙箱提权或static browser.check假后端。此替代足以证明固定工程门禁，不足以弥补缺少的AC01实际页面对照。

真实供应商/已有推理端点、外部身份/支付Not Run；受控HTTP/SSE仅证明项目语义，USD是测试分账。H02重复入站、H03-runtime实际模型、H04材料负向恢复、H07同PR合并/Actions/Deployment/health固定URL、H08平台维护源安装升级复验仍缺可核对事实，归平台/协调者，不要求研发越权补这些回执。H05全部设计符合不提前通过；H06独立验收返工和原层关闭已有事实通过。不能声称已商用全面对齐New API/LiteLLM或已发布。冻结外GPU/其他协议/税务支付/生产SLO/多实例协调记录为未来差距，不增本轮范围。

## 7. 环境恢复

QA未启动持久服务、修改业务源码/DB/用户上游或平台状态。自身InProcess测试使用测试拥有临时store并由cleanup处理；Node为隔离DOM/脚本契约，不访问真实供应商。宿主runner的独立临时目录、随机端口、context/browser/server/store关闭与目录清理契约有代码和运行完成回执；未独立进行全机进程审计，不编造全局清理证明。仅在本Run增加验收报告/矩阵/采集验证资产和脱敏证据，缓存/运行二进制/本机路径不提交。

## 8. 质量趋势

相对QA35，读回临时失败的P1已由当前asset原用户路径三视口关闭，原覆盖P1主题的故障/双发送/资源组合从Not Run提升到实际通过，旧P2无复发。自动化由165 Node扩至169、新增两scope五维及共享资源ActualHTTP并保留旧原账本/价快照断言；七条主链无退步，质量提高。剩余产品侧风险收敛到十页批准设计的完整实际对照，没有因为门禁绿而隐藏原Not Run行。新发现是证据门槛遗漏，不是捏造新的代码失败；未来供应商/平台/发布事实边界仍清晰。

## 9. Go / No-Go与路由

**No-Go：AC01强制设计对照门槛尚未实际完成；不是固定make verify失败。** QA35-01、QA22-01原主题范围、QA22-02已关闭，不要求研发重复修这些功能或篡改账本。沿正式handoff回development补同版实际页面证据或给出已存在的具体等价产物；如测试/源代码变化形成新指纹，必须宿主原完整make verify与独立QA重新核对，不拿3ee2122…旧绿放新版本。平台H/发布在各自责任节点后续闭合，此次不进入report自称完整交付、不切分支提交推送、不创建PR/部署、不重启Run。
