# design-v0.2.0 / visual-editor-owner-r1 独立验收报告

> 日期：2026-10-08；Run 4cd62ced6a4f98f50e3cca6f7596eb06，QA执行47。
> 仓库：big91987/model-relay；分支 workflow/4cd62ced6a4f98f50e3cca6f7596eb06。
> HEAD：98511771871cf0951ecef716bbb55deb13e18da5；候选未提交工作树，129文件 SHA256：7e0b20ddbdea4585959365d8ad3b6121abaa91cd3f450fa2c6c21f7afc05a9ab。
> 模式：恢复既有全量验收范围、审查当前完整宿主回归并独立逐页对照；不是只复验定位器。
> 环境：宿主实际Go临时服务、SQLite schema3、受控HTTP/SSE上游、真实内部用户会话、Playwright三视口；非生产部署。Go1.25.6、Playwright锁版本1.59.1；QA采集Node v26.0.0。
> 结论：**No-Go，返回 development**。固定门禁通过，但强制AC01设计符合性失败。

## 1. 结论摘要

当前版本宿主 `make verify`（seq46）exit0，104299 bytes、truncated=false。129文件指纹与QA独立计算一致；三视口J01–J07、当前周期真实活动、十页采集均完成。QA实际查看30张页面图、1280四张详情图及390六张编辑图，并检查63张PNG的字节摘要。十页证据采集缺口已补齐，不再以“没有截图”返工。

实际页面却减少了批准设计要求的主要管理摘要：模型列表没有来源/当前价格/延迟等信息，租户页面没有负责人、授权模型、Key与预算摘要，服务列表缺少模型关联数量和核验时间。现有采集断言按当前实现选字段，故绿色不能证明没有删减冻结设计。记录QA47-01 P1，AC01仍Fail。另有QA47-02 P2：已核对请求在分析页显示确认输入3/输出4，日志却把未标明口径的“Token总量”显示为未知；原始供应商事实可保持不可变，但应与核对后的事实区别表达。这不是已证账本丢失或少收费。

无新的需求或架构歧义，无需用户重复批准。真实供应商、外部身份/支付、完整Harness负向/实际模型来源、PR/合并/部署仍Not Run；不能称已商用全面对齐或最终交付。

## 2. 本轮覆盖

- 权威规则：同prepare archive SHA256 5f51d35fdc15ed7214749f9413879b4238097f1d6b6bbe2cc0d508d8e103cef7；manifest19文件逐项长度/hash一致。实际路径及逐件摘要在 `evidence/attempt-47/material-integrity.json`。读取冻结PRD、interactions、AC、decisions及baseline-review最终承接，参考原型源码/样式与overview/tenant图片，不执行包内脚本。架构沿用architecture-r1、implementation-contract、verification-design；不重开R02–R04。
- 宿主：固定入口含历史健康预期红重建、format/modules/vet、182项Node契约、Go race真实HTTP、双scope五维资源竞争、幂等计费/重启、构建、正式CLI旧schema1/2→3及三视口浏览器旅程。历史健康FAIL及旧editor预期失败属于隔离历史重建，不是当前生产失败；当前健康HTTP、故障读回和双发送回归通过。
- 连续网页：J01首次配置及邀请/登录/Key/调用；J02授权收窄拒绝与修正；J03轮换共享池及退出；J04硬预算拒绝恢复；J05成员失权与多Key租户暂停；J06价格/待确认/结账/追加更正；J07真实筛选分页导出及迟到隔离。各角色使用独立真实context，各单Actor子链连续执行，角色交接不伪装为一个会话。参见 `user-journeys-47.md`。
- 逐页UI：10页×1280/1440/390全部实际截图，对照结果见 `visual-review-47.md`。白底后台/稳定桌面导航/窄屏菜单/统一编辑/右抽屉基本成立，非空管理数据及当前周期活动有真实证据。没有因配色/品牌或缺少装饰判P1。
- 分层结果：`acceptance-matrix-47.md`逐项AC01–AC18/H01–H08，UI-E2E、API-E2E、Contract、Data Reconciliation和External State分别记录；截图不代替连续旅程，底层守恒测试不代替设计符合性。

## 3. 缺陷及进展

| ID | 等级 | 用户能看到的问题 / 影响 | 复现与定位 | 状态 / 回归 |
|---|---|---|---|---|
| QA38-01 | 原强制证据门槛 | 原先缺十页同版实际对照证据 | 当前三个review+63PNG摘要及QA30页人工查看 | Closed（仅采集缺口）；AC01不因此通过 |
| QA47-01 | P1 | 多模型/多租户日常管理不能在批准列表/摘要中识别来源、当前价、负责人和资源边界；需逐对象进入原始字段或切页，部分摘要根本未提供 | 已配置两模型/两租户/四Key，进入模型、服务、租户页；三视口均同样缺项。workspace.js:184–185、services分支；console-visual.mjs:7–10只断言当前删减后的字段 | Open，owner development；原真实Go采集及AC字段/实际值断言必须补齐 |
| QA47-02 | P2 | 已核对请求分析Tokens3+4、日志Token总量“未知”且待确认0，未解释原始/核对口径，排障人员容易误判 | 同Request ID筛选usage/records→详情；三视口。workspace.js:190直接用total_tokens；console_reports.go读取原requests/attempts，usage读取charge_entries | Open，owner development；不因本项单独阻断，不要求覆写原始用量或账本 |
| QA35-01 / QA22-01 / QA22-02 | 既有分级保留 | 未知读回、连续覆盖、列表状态历史主题 | 当前seq46 fault及J01–07继续通过；历史报告/原红保留 | Closed，不重开、不降级，不重复研发副作用 |

QA47-01不是给轻微P2样式问题升级；冻结PRD§7明确不减少结果，interactions“页面及行为”明确这些列表/摘要，AC01要求实际批准设计对照。可用等价紧凑摘要，不要求像素或列顺序完全一致，也不要求搬入演示角色/GPU/支付/其他协议。

## 4. 证据与来源

- 平台seq46归档：通过read_command_output取offset0→32768→65536→98304→104299/eof；每段archive truncated=false。原完整日志由平台保管，未伪造本地完整stdout。摘要及真实命令责任见 `evidence/attempt-47/host-seq46-review.md`。
- 当前宿主产物75件保存在 `evidence/attempt-47/host-seq46/`：63PNG、3review、3continuous（34454/34476/34501）、3visual-activity、3fault；`artifact-index.json`记录源路径、仓库相对保存路径、bytes/SHA256。图未重绘、裁剪或拼装；来源保留受控上游标签。
- QA实际执行：源戳Go入口（只读、不监听）、Node证据完整性采集、材料archive/19文件核对、真实PNG查看、静态基线/源码/测试资产对照。`collect-final.log` exit0仅证完整性，不是新跑make或浏览器。
- QA初次未指定本地GOCACHE，源戳运行被个人cache权限拒绝，改用项目GOCACHE后取得同指纹；未提权。采集器首轮把fault误当有总体result而失败，修正为phase及五case逐项检查后成功；不是产品回归失败，说明保存在 `collector-attempt.md`。

## 5. 未执行和环境限制

QA原生固定入口历史监听Blocked保留，按用户明确指令不再重跑已知必败 `make verify`；同版本完整宿主seq46为固定门禁等价证据。QA没有亲自运行这些真实浏览器链，不将宿主执行记成QA执行；独立验收是审核当前回执、版本、字节及页面设计。

self_hosted来源的API/核验语义有受控证据，独立首次self_hosted UI分支仍Not Run。真实供应商/真实自托管凭据联调Not Run，不阻断当前可完成的内部整改。外部支付/身份未接入且非本期默认承诺。

H02重复接单、H03实际原生模型、H04材料负向恢复、H07同PR/Actions/Deployment/health固定URL、H08平台维护源修复安装升级均由平台/协调者提供；本节点不能自证gpt-6.1-sol，只能记录指定配置，未切换模型或委派。最终AC/H完整退出仍未满足。

## 6. 阶段边界

本轮只要求兑现design-v0.2.0现有信息/操作与口径，不新增New API/LiteLLM全套高级功能。公开参考沿用冻结references.md和已核对官方来源，不读取New API实现/资产/本机数据。后续GPU、多协议、公开支付税务、生产SLO、多实例协调等仍在既有差距列表，不进入返工门槛。USD金额为平台测试价，不代表实际供应商报价。

## 7. 环境恢复

QA没有启动持久服务、改业务实现/测试驱动/DB/冻结材料或用户上游；仅新增本Run验收文档与证据。宿主runner声明隔离目录、随机端口、browser/context/server/store cleanup，当前continuous回执标cleanup；未独立进行全机进程盘点，不声称生产已恢复。无提交、推送、PR、部署或GitHub评论副作用；缓存/秘密/绝对本机路径不列交付。

## 8. 质量趋势

相对QA38与seq44失败，限额editor owner已由当前三视口真实打开关闭，十页采集从Not Run变成可审查的真实63图；全部连续旅程保留，Node资产182/182。当前周期空数据采集也由新真实调用闭合，没有造数据或改账本。新发现是冻结主要摘要删减P1及已核对Tokens解释P2，不是定位器故障复发。工程质量改善，但门禁断言尚未覆盖全部冻结信息；需把本次发现沉淀为断言，避免“测试跟着缩水实现走”。

## 9. Go / No-Go与路由

**No-Go：QA47-01 Open/P1，AC01设计符合性Fail；make verify当前Pass。** 正式回development按 `remediation-handoff-47.md`最小整改并同版本完整入口复验。P2单独不阻断；H/供应商/发布未测不推给研发越权补造。本节点不进入report自称交付；后续无产品P1且强制门禁闭合时再交report整理PR和剩余外部责任。
