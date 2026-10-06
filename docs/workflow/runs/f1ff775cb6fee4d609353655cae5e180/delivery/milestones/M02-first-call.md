# M02 管理员完成首次配置与可追踪调用

- 当前状态：**Accepted（本切片实现与验收，2026-10-06）**。首次空配置全UI、390px存量主链、同请求诊断、清密恢复与四项收尾退出条件满足；见[当前报告](../../acceptance/report-02.md)。以下为历史实施与待验记录，不是Delivered/Published。

- 2026-10-06既有契约收尾：按维护者静态反馈修复GET-only撤销/共享保存恢复忙态、实际调用参数锁定、同记录created/未知、按alias/key ID即时行焦点恢复；不是新原生回执。101VM/helper、3Go in-process race及分项通过；原QA脚本恢复执行仍3/4，QA-03保持独立语义待评审，未改QA资产。新两屏/完整make verify交宿主，任务In Review、里程碑Ready for Review，QA No-Go与历史保留。见[当前对象与证据](../verification/closure-contracts.md)；下方旧对象不能放行本对象。

- 2026-10-06 seq27：宿主make verify退出2，390px启用目标index1仍停index0；旧OS弹层选择序列未奏效。本轮仅T003改闭合ArrowUp/启用序号ArrowDown/Tab，保留精确值/焦点/全部业务断言，85VM/helper与分项通过，新对象两屏待宿主原入口。任务In Review、里程碑Ready for Review，QA No-Go/驱动差异保留，见[最新记录](../verification/host-return-10.md)。以下为历史，不用于新对象放行。

- seq25后仍Ready for Review：1280px旧对象通过，390px模型选择失败未完成；T003补原生提交与诊断，77产品VM+4helper契约通过不能替代两屏。新对象宿主与独立QA仍须完成，见[记录](../verification/host-return-09.md)，以下历史保留。

- 维护者静态设计核对补齐成功证据time/revision的清楚表述与原层验证，77VM通过；不等于两屏或独立QA通过。保持Ready for Review，见[记录](../verification/evidence-metadata.md)，QA原报告保留。

- 最新seq23等待旧状态的问题按真实GET完成边界修正，仅旅程/VM变更，75VM与分项通过。新完整两屏待宿主、QA首轮No-Go仍待独立复验，保持Ready for Review、不Accepted，见[记录](../verification/host-return-08.md)。

- 最新seq21复制反馈超时；T003明确测试能力前置并补拒绝恢复/原生内容门禁，产品契约未改。74VM通过，新两屏闭环待宿主和QA。仍Ready for Review，不Accepted，见[最新证据](../verification/host-return-07.md)，后文为历史。

- 当前状态：Ready for Review。seq19整改实现及74VM通过，失败清密/写读分离/明确同ID动作已补，两屏完整控件键盘/恢复/分页用例进入原门禁；原生新对象待宿主及独立QA，原QA-03泛导航驱动仍红不隐瞒，不关闭P1、不Accepted。见[整改实施证据](../verification/qa-rework-01.md)；下方为历史。

- 最新seq16：1280px真实Go首次全UI调用旅程通过（受控上游），390px在复用成功配置时因旧摘要关键词等待失败。T003已修正真实overview状态核对；仍Ready for Review，原生两屏与独立QA退出条件未满足。见[最新记录](../verification/host-return-06.md)，下方为历史。

- 状态 Ready for Review；负责人研发，验收人独立QA；任务[T002](../tasks/T002-console.md)。
- 进入：G2交互与M01实现契约已存在；硬依赖M01。
- 退出：真实Go空配置全UI→普通/SSE→同ID记录；两屏/键盘/深链/过期；取消、防重复、晚到/秘密保护和中文恢复均通过。
- 范围：PRD AC-01–12/15；原生前端，无自动供应商调用或重放。
- 实际证据：[development.md](../verification/development.md)。41项VM通过；浏览器已扩展但宿主运行未完成，尚不Accepted。
- 风险：浏览器原生布局/焦点/AX不能由VM证明；真实供应商未测。
- 同Run补充证据：[supplement-rework.md](../verification/supplement-rework.md)。工具与会话边界已补齐，48项VM通过；宿主真实Go首旅程、两屏新增边界及18AC独立验收仍是退出条件，状态不升为Accepted。
- 第二次补充当前证据：[expiry-watchdog.md](../verification/expiry-watchdog.md)。独立到期与一次性timer回归58项VM通过；仍不以确定性隔离诊断替代宿主浏览器或独立QA验收，状态不变。
