# 日志升级后交宿主重新验证

2026-10-06，同Run同研发会话。此记录更新当前动作，不覆盖[前次阻塞记录](host-return-01.md)、seq7或历次日志。T001/T002/T003 In Review，M03 Ready for Review，未Accepted。按研发交付管理与评审反馈Skill核对维护反馈、工作树和公共状态后，正式handoff next到tests。

## 维护反馈及证据归属

维护者确认seq7宿主回执只保留开头，实际失败尾部已丢弃、无法恢复；不再需要用户提供丢失日志。可见旧版HTTP断言属于预期历史红测，不据此修复当前产品。

维护者报告：已在唯一平台维护源修复首尾日志和截断前脱敏；完整平台门禁通过；d5e1403已升级开发实例且通过原manifest升级。以上仅记录维护者反馈，本节点未读取或修改平台源、运行库或安装配置，也未独立执行平台门禁；这不是本项目验证通过的证据。

## 当前工作树核对

- 实际运行`git status --short`、`git diff --stat`及`go run ./tests/evidence`，证据命令退出0；未运行监听测试。
- HEAD仍为d62cd51e8de2127548e1ac7469d349443d1bfe3d；56文件指纹仍为9254855737fbe400acf1d8c3a038767ca7807e0e1ea43160b7b5ed87d5456d0a，与seq6交接、seq7前缀及[第8次日志](verify-attempt-08.log)一致。机器可读结果见[source-host-retry.json](source-host-retry.json)。
- git状态中的产品、测试、项目契约及Trellis规范变更均对应前序已交接实现与需求/设计。现有未跟踪console.go/console_test.go和两份console规范均已列于seq6 artifacts；本Run目录的新增文件为既有需求/设计/交付证据及本轮公共状态更新。无未解释产品变更；不是干净工作树声明。
- 本轮仅修改公共计划、T003、M03，新增本记录和指纹回执；不改产品、测试、Makefile、锁文件或规范，不猜测修复、不放宽门禁。原生沙箱监听限制已有多次真实回执，无须重复相同受限命令。无需新增Trellis规范或bootstrap。
- 文档更新后`git diff --check`退出0；当前文档新增引用路径存在。前序Trellis检查及分项回执继续保留，不作为本轮重新运行的结果。

## 正式交接与退出边界

tests Connector按固定入口原样执行`make verify`，保留新的真实宿主退出码、首尾诊断及脱敏证据，不重放/改写seq7，不跳过历史门禁。证据归档于本Run根；当前只等待新宿主结果，不等待旧日志或阶段批准。

若失败，由正常失败路由回研发，使用新尾部定位并针对性返工；需求/设计缺陷按相应合法路由处理。若通过，继续独立QA逐条18AC与设计边界验收；不能把tests结果提前当成独立QA或发布结论。

继续采用既有[PRD](../../product-definition.md)、[架构](../../console-architecture.md)、[交互](../../console-interaction.md)及[研发AC映射](development.md)，重点保留真实Go全UI首次配置、390px/桌面键盘、请求ID核对、拒绝零上游触达、HTTP/SSE取消与晚到隔离、独立到期计时器、工具摘要和秘密排除。受控上游不等于真实供应商联调。

本轮没有新项目make verify结果、供应商、浏览器、独立QA、PR、合并或部署成功证据；公共路线2–4仍属于后续独立Issue。
