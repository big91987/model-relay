# design-v0.2.0 / keyboard-ready-r1 独立验收报告

> 节点：qa，执行22，Run 4cd62ced6a4f98f50e3cca6f7596eb06，冻结工作流版本1，Issue #25。
> 时间：宿主seq21及QA命令日志记录2026-10-08（+08:00）；时间是命令环境事实，不推定部署时间。容量中断后在原节点/会话正式恢复，没有重跑已完成的副作用。
> 仓库/分支：Model Relay / workflow/4cd62ced6a4f98f50e3cca6f7596eb06；HEAD 98511771871cf0951ecef716bbb55deb13e18da5，候选包含未提交和未跟踪产品文件。
> 快照：119文件，working_tree_sha256 71be54206d9850fa442f33c18ddd3c7aaba5590019543a78498a128ae0ab8350，前后独立来源一致。
> 环境：本地Go临时服务+SQLite schema3+受控HTTP/SSE；不是正式产品发布。Go1.25.6 darwin/arm64、Node26.0.0、npm11.12.1、Playwright锁定1.59.1。
> 模式：全量验收（首次完整独立QA，历史覆盖不完整），不是仅复验keyboard。
> 结论：**No-Go，返回development补齐强制UI验收覆盖**。

## 1. 结论摘要

同材料、同工作树指纹的宿主make verify seq21真实exit0，88327字节且归档未截断；新控制台1280/1440/390现有脚本全部走完。原keyboard-1280失败路径本轮已通过，不能据此认定宿主唯一根因或整套产品通过。独立审查发现当前脚本没有完成冻结J01–J07及architecture-r1验证设计要求的租户角色业务、授权收窄恢复、费用预算恢复、暂停恢复、账单核对结账更正和新workspace跨scope分页恢复，登记QA22-01 P1覆盖缺陷。另有Key列表无法区分有效/停用/过期的状态表达问题QA22-02 P2，不单独作为阻断理由。未执行的产品网页路径、Harness和正式发布事实不能因固定门禁绿而标Go。

## 2. 权威材料与实际引用

材料root：`.workflow-input/4cd62ced6a4f98f50e3cca6f7596eb06/5f51d35fdc15ed7214749f9413879b4238097f1d6b6bbe2cc0d508d8e103cef7/files`；manifest入口为该root的manifest.json。版本design-v0.2.0、frozen-for-test-implementation；archive SHA256 5f51d35fdc15ed7214749f9413879b4238097f1d6b6bbe2cc0d508d8e103cef7。QA独立校验19文件与archive全部一致，详见evidence/material-integrity.json。

| root下实际引用 | SHA256/用途 |
|---|---|
| docs/prd.md | e768b4f26b4471e714d92b0db6e3187c89f37a25a697c01a12c6079af07de547；完整规则、四角色、首次正式入口、独立预算、微美元、全部dispatch计费、账单责任 |
| docs/interactions.md | 95f3490ef003fa9b19dd6b9e864dd3b0429619b8cd9d0ecec853919d6a12c801；A工作台、十页动作、J01–J07和失败状态 |
| docs/acceptance.md | 2792740b1579d614780856a13b6257d192736e1efc0309f1dc41405a665bef7a；全部AC01–AC18/H01–H08 |
| docs/decisions.md | faed344c8d63ca9bf909b4c1255a301005c20d04de8b3120fa3b67eb40063220；自主授权/正式身份/账单决定，无需重问原型批准 |
| docs/baseline-review.md | 1074fb6993a13de6d60a2b87381506f1a8bb1f367c326930b02d753829572271；最后整改闭环优先，R02–R04不重新开启 |
| prototype/index.html / app.js / styles.css | 摘要逐项在material-integrity.json；入口及Key列表/轮换信息架构参考，包内脚本没有执行 |
| docs/screenshots/keys.jpg | b992c2fec02cc55b4e5087bb2edb3f6e331686894428dfbe443b6d8ab0c6fe0e；实际查看批准Key页，不作为产品运行证据 |

承接本Run product-definition.md、architecture.md、verification-design.md、根acceptance-matrix.md及attempt20真实来源/局部检查/seq19失败记录。冻结正文不改写；本轮矩阵另放acceptance目录以保留119文件测试快照。旧Run、旧工作区及本机New API源码/数据/秘密均未读取。用户补充商用参考沿用已交接八类能力，不扩大到支付/GPU/生产SLO等冻结范围外能力。

## 3. 本轮覆盖与真实性

- **工程门禁**：宿主seq21执行完整make verify，含历史预期红灯门禁、format/modules/vet、152 Node测试、Go race及真实HTTP、build、真实Go服务浏览器和迁移。QA不是该宿主命令执行者；日志索引见evidence/host-seq21-review.md。用户恢复消息补充维护者已核对三页完整归档，仍不冒充QA本地网络执行。
- **QA独立执行**：format-check/modules/vet、152 Node、六项针对性的Go race非监听测试和build，组合exit0。Go选择GovernanceInProcess、ResourceDimensionsInProcess、PeopleScopeInProcess、AccountingInProcess、GenerationExactDeadline、MicroDollarPrecision；它们使用实际产品handler和明确受控transport，不是真实网络或供应商证明。
- **本地固定命令**：QA独立make verify实际exit2，在health-history被sandbox bind denied阻断。日志保留；正式恢复后按用户要求不再次重跑已知必败监听命令、不扩大权限、不删门禁。初次源码go run默认用户cache也被拒，改用项目既有.cache/go-build后读回成功，不改产品。
- **UI正常子集**：平台管理员在三视口真实登录、建租户、接入核验服务、发布并显式授权模型、定价、签发及显式转交秘密、普通/SSE调用和同ID详情、取消、立即轮换、未知签发读回、次数拒绝、十页导航、弹层焦点和退出清理。成员在单独真实身份context接受邀请/登录/看scope/退出，不能与平台签发拼成租户管理员首次链。
- **API/数据/恢复**：宿主治理、人员scope、资源维度、价格精度、计费追加、schema1/2→3正式CLI及三视口迁移、进程kill/restart测试记录通过。各局部证明与未闭合不变量分列矩阵；不因测试名存在便认定全AC覆盖。
- **视觉对照**：实际查看批准keys.jpg与当前工作区console-1280.png/console-390.png，确认明亮稳定导航/紧凑表格框架和Key状态/字段缺口。复制脱敏截图作为辅助观察；截图缺独立版本绑定，不用其代替UI-E2E或宣称逐页对照完成。
- **完整旅程**：user-journeys.md分别记录Actor/Start State/UI Steps/Expected UI/Final Fact/Cleanup。J01–J07完整UI-E2E均Not Run；上述子集在acceptance-matrix.md单独Pass，不跨角色、跨scope或跨UI/API拼接。

## 4. 缺陷与进展

| ID | 等级 | 用户能看到的问题/证据缺口 | 复现/定位 | 影响 | 状态 | 回归资产/责任 |
|---|---|---|---|---|---|---|
| QA22-01 | P1 | 未证明租户管理员和费用责任人能在正式UI完成承诺工作；当前绿色测试集漏核心路径 | 比对冻结J01–J07、verification-design§4与tests/browser/console.mjs:95起；quota仅次数拒绝，无暂停或账单动作，member仅登录/退出 | 强制AC及三视口退出标准覆盖缺口，阻断放行；不声称未测产品一定失败 | Open | development补完整实际UI测试、请求及最终事实；remediation-handoff.md列具体用例 |
| QA22-02 | P2 | 未撤销Key一律显示“有效 / 按启用与期限判断”，不能辨认停用/过期；列表缺授权/预算摘要 | workspace.js:156列只取revoked；evidence/key-status-review.mjs抽取实际formatter，三种状态显示相同 | 日常安全运营需要开详情或自行推断，信息层级偏离基线；不独立阻塞 | Open | development改准确状态并补真实浏览器停用/过期/过渡/撤销行回归；本轮证据为Static Review，非过期Key UI复现 |
| 历史keyboard-1280 | 保留原主题 | 上轮editor仍开/断言失败 | attempt20/host-seq19-summary.md；本轮宿主B原13Tab/Escape/返回点 | 原发现层路径当前已绿 | 当前路径回归通过 | 保留keyboard-ready-r1及原断言，不对唯一根因补造结论 |

P0未发现。QA22-01是新发现的覆盖遗漏，不将测试缺失改写为业务执行Fail；对应未跑旅程仍Not Run。QA22-02分级建议为P2，未有已批准等级被静默升级。QA不修改业务实现，整改归development；没有需要需求/设计重决策的阻塞。

## 5. 证据索引与来源

- acceptance-matrix.md：全AC/H、已执行局部证据和八类商用能力五层状态。
- user-journeys.md：J01–J07连续旅程及缺步、夹具和清理边界。
- remediation-handoff.md：实际返研动作、验收语句、不可回退门禁及下轮复验。
- evidence/source-before.json、source-after.json：119文件版本一致；源码算法包括产品/测试/锁/契约，排除QA交付回执以避免自引用。
- evidence/material-integrity.json：实际原资料相对路径、19文件SHA及archive核对。
- evidence/host-seq21-review.md：平台保留日志seq21/offset/eof索引和边界；不是本地伪造完整日志。
- evidence/qa-make-verify.log：QA实际失败完整日志；监听限制单列，未当产品失败。
- evidence/qa-independent-checks.log：QA独立六项非监听Go race/152 Node/format/modules/vet/build；临时链接器路径以占位符脱敏，其他结果不改。
- evidence/key-status-review.mjs、key-status-review.json：静态实际formatter验证，distinguishes_enabled_disabled_expired=false；脚本exit0表示采集成功，不表示产品状态正确。
- evidence/console-1280-observed.png、console-390-observed.png：工作区现有截图只作辅助视觉证据；evidence/receipt.json记录来源和文件摘要，不能代替操作回执。

## 6. 未执行、环境限制与阶段边界

供应商/真实自托管端点、外部身份和支付没有凭据，均Not Run；不阻断当前可完成内部身份和受控产品验收。新workspace缺用例是可补齐的研发责任，不归入供应商或sandbox限制。H02唯一Run/重复入站、H03实际原生模型、H04材料负向恢复、H07同PR合并/Actions/Deployment/health固定URL及H08平台维护源正式升级事实仍无完整QA可用证据，保持Not Run/责任归平台和协调者，不要求产品研发修平台数据库。用户指定Codex/gpt-6.1-sol持续有效，QA没有选择或切换模型，也不拿配置/自然语言自报当实际模型回执。

当前支持工程候选测试集绿色，不支持“商用能力已全面对齐”“完整产品Go”“已经发布”。正式测试仓合并/部署授权仍有效，但须先关闭强制缺口并复验，随后report/Connector处理；不创建分支/提交/推送/PR，不调用生产部署或旧Run。

## 7. 环境恢复

QA没有访问用户上游/数据，没有修改数据库让测试通过。独立InProcess使用testing.TempDir及store清理，未启动持久监听服务；失败的health-history由Go工具自身退出。宿主runner代码拥有临时目录、server/store/browser关闭，未单独取得运行后的全机进程清单，故不声称宿主资源清理经独立全机审计。QA仅写本Run文档、证据和忽略的本地构建缓存/bin；绝对主机路径和秘密不进入新交接产物。

## 8. 质量趋势

相对seq19，当前固定入口第一次完整走过Key转交/调用/账本/配额/keyboard并完成三视口，工程可靠性提高。152项Node与既有原生断言保留，迟打开editor回归资产有效。新增发现1项P1核心UI覆盖遗漏、1项P2列表表达偏差；没有证据表明旧keyboard路径复发。未用“全部入口可打开”掩盖账单及租户治理未测；剩余风险为角色连续旅程、原型信息表达、Harness事实和正式发布链，而非供应商凭据本身。

## 9. Go / No-Go及下一轮

**No-Go：QA22-01未关闭，强制连续UI退出标准未覆盖。** QA22-02是P2默认不阻塞，列整改和风险。按用户正式恢复指令向development handoff；不因为容量中断或宿主门禁绿改变结论。

下一轮补入明确UI旅程/测试并运行新指纹完整make verify；携同manifest、矩阵、真实宿主日志前后来源、失败与整改证据再独立QA。原119文件71be…绿色只对应当前候选；改测试或产品后不能沿用。没有需要再次澄清的产品决策，容量问题若未再发生不虚构为外部阻塞。
