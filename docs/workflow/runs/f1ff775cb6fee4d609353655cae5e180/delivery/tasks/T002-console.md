# T002 首次配置到可读调用

- 当前状态：**Done（本切片，2026-10-06）**。第二轮QA关闭失败清密/部分保存反馈/同ID证据动作，四项收尾及真实Go两屏旅程通过；见[当前报告](../../acceptance/report-02.md)。以下为历史实施与待验记录，发布未执行。

- 2026-10-06既有契约收尾：按维护者静态反馈修复GET-only撤销/共享保存恢复忙态、实际调用参数锁定、同记录created/未知、按alias/key ID即时行焦点恢复；不是新原生回执。101VM/helper、3Go in-process race及分项通过；原QA脚本恢复执行仍3/4，QA-03保持独立语义待评审，未改QA资产。新两屏/完整make verify交宿主，任务In Review、里程碑Ready for Review，QA No-Go与历史保留。见[当前对象与证据](../verification/closure-contracts.md)；下方旧对象不能放行本对象。

- 2026-10-06 seq27：宿主make verify退出2，390px启用目标index1仍停index0；旧OS弹层选择序列未奏效。本轮仅T003改闭合ArrowUp/启用序号ArrowDown/Tab，保留精确值/焦点/全部业务断言，85VM/helper与分项通过，新对象两屏待宿主原入口。任务In Review、里程碑Ready for Review，QA No-Go/驱动差异保留，见[最新记录](../verification/host-return-10.md)。以下为历史，不用于新对象放行。

- seq25：1280px真实Go完整旅程与QA新增场景通过，390px在旅程选择模型值断言失败；未发现产品焦点/CSS失败证据，不改产品。T003修正原生提交/诊断，完整新对象与独立QA仍待验；状态In Review，见[记录](../verification/host-return-09.md)，下方历史与QA No-Go保留。

- 2026-10-06设计补齐：概览同成功证据行补time/upstream_revision，用“调用时间／配置版本”与known未知表示；无新对象/迁移/全局版本或时钟回退。77VM与分项通过，两屏新对象待宿主和独立QA。仍In Review，见[记录](../verification/evidence-metadata.md)，不覆盖QA首轮结论。

- 最新seq23：源码定位旅程等待旧状态导致迟到重读覆盖草稿；本次仅T003修正同步，未改产品地址策略/清密。75VM及分项通过，完整新两屏与QA待宿主/独立QA。仍In Review，见[最新记录](../verification/host-return-08.md)，后文为历史。

- 当前seq21复制成功反馈超时；产品已有手动路径，此次仅T003明确测试能力与拒绝恢复/原生内容核对，不猜测改产品权限。74VM通过，新两屏/QA整改仍待宿主与独立QA。仍In Review，见[最新证据](../verification/host-return-07.md)，后文为历史。

- 当前状态：In Review。2026-10-06负责人研发接回seq19，按[整改包](../../acceptance/remediation-01.md)完成当前owner失败清密、保存/回读分离、概览明确成功ID动作与泛导航分离；74VM通过，原独立QA三条红转绿但QA-03旧泛导航driver仍失败并完整保留。新真实Go两屏原层异常需宿主复验和独立QA判断，三缺陷未自行关闭、不Done。见[本轮实施/边界](../verification/qa-rework-01.md)；下方为历史。

- 最新seq16：1280px完整真实Go旅程通过，390px在复用已成功配置时等待旧文案超时。T003修正响应及当前成功证据核对，60VM通过；本任务仍In Review，两屏与18AC仍需宿主和独立QA。见[最新证据](../verification/host-return-06.md)，下方为历史。

- 最新seq14：登录回归已越过原失败点，1280px运行版本旅程通过，但通用键盘焦点断言失败。本次未改产品；T003采用真实Tab到达及焦点分项诊断，原生两屏仍待宿主新对象验证。见[最新记录](../verification/host-return-05.md)。状态仍In Review，不将部分旅程当作全部AC通过。

- 最新seq12返工：旧含“请”的session错误提前满足错误密码登录等待，探针沿用显隐循环阶段。T003修正真实401/专属错误完成边界，未改产品显隐/鉴权；59VM通过，原生两屏仍待宿主。见[最新记录](../verification/host-return-04.md)。仍In Review；下段seq9为历史状态。

- 当前宿主seq9密码type回归失败，需复验；没有证据支持修改产品显隐/鉴权。T003补齐原生事件同步和安全诊断，详见[返工记录](../verification/host-return-03.md)。本任务仍In Review，原生两屏退出条件未满足。

- 状态：In Review；里程碑 M02；负责人研发；评审项目测试/独立QA；执行 SDD。2026-10-06领取，基线 d62cd51e8de2127548e1ac7469d349443d1bfe3d。
- 硬依赖 T001查询契约；冲突范围 internal/relay/web。
- 权威输入：[交互§1–10](../../console-interaction.md)，PRD AC-01–12/15。
- 目标：hash导航、真实概览、依赖空状态、可取消编辑、一次密钥、占位示例、普通/SSE可读输出与诊断/记录。
- 非目标：框架、多路路由、秘密持久化、自动重试。
- 验收：390px/键盘；重登录恢复安全页；所有await隔离；取消防重复；秘密离页清除。
- 验证：VM迟到回归、真实Go浏览器全UI闭环和 make verify；证据 ../verification/development.md。
- 实施结果：原生三文件实现六页hash/刷新/未知目标、真实计数和步骤、依赖引导、模型内联编辑/保留取消草稿、一次密钥显隐/清除、占位curl、启用模型选择、普通文本/工具结果、SSE/取消/防重复、精确记录及中文恢复。
- 异步所有权：会话/导航/调用身份、load/find序列、issuedEpoch剪贴板隔离；管理401在当前响应头时同步清除；到期检查、pagehide、退出清秘密。保存中的编辑不允许被另一编辑器/取消替换；签发中的取消禁用，不暗示已取消服务端提交。
- 已验证：41项VM测试含原密码/版本回归、迟到JSON/SSE/签发/401/finally、复制成功/失败边界、0/null、取消草稿、编辑保存隔离、未知hash；见[实际验证](../verification/development.md)。
- 无方向/接口级设计偏差。输出/诊断按设计有界；客户端示例保留模型占位符，不拼接供应商名。未完成：真实两屏布局、键盘/AX及Go全UI旅程的宿主验证。下一步T003/项目测试；不标Done。
- 维护者补充修正：首轮工具提示和60秒/网络失败会话分支弱于已接受契约，现已修正而非修改设计。普通工具/SSE增量按索引汇总数量和有限名称，不执行工具；可见长流每30秒检查，网络异常失效全部异步所有权并停止/清密，不冒充401，仅显式重新检查或登录恢复。产品概览移除所有用户固定“供应商未测”状态，实际供应商Not Run只在本轮报告保留。最新48项VM和新增宿主旅程见[补充记录](../verification/supplement-rework.md)；状态仍In Review。
- 第二次补充修正：本地已知到期由独立session-epoch/deadline计时器负责；pending查询不延误清理，跨页不撤销截止时间，正常检查不延寿、不复活旧会话，不明恢复悬挂也受截止约束。正常周期成功不修改生成状态/输入/输出/提示；只恢复不明状态才提示手动重新输入。当前58VM及修复前失败见[悬挂检查验证](../verification/expiry-watchdog.md)；仍In Review，浏览器/独立QA待宿主。
