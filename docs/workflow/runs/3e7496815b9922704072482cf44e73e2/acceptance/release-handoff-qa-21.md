# 执行21 → report：精确发布及正式部署待办

## 当前事实
独立QA候选结论Go with known issues；不是已经上线。基线HEAD/base cd413b5aa24a372a19c19af999a6ef61b5c41572；147件已验SHA256 f8befd158a543afeb1ba77778e4caf72e876f6e4b0b988dd4e6ab73de5e5eaab。源码原件 `verification/source-development-19.json` 与 `acceptance/evidence/qa-21/source-qa-21.json` 一致。工作区仍是研发未提交改动；不得把base HEAD当已验新commit。

执行20固定make verify正式宿主完整冷exit0、Connector310815bytes/未截断/读取至EOF；实际runner310718bytes/hash7c4d3952943721deae9188a6f52f860572037ac8ea39ba55d278d6bf2c7abb7b。完整原样log/state、三视口continuous/summary/visual/故障与截图可访问，索引列每个具体文件及hash。QA实际局部回归绿；QA原生make verify监听Blocked保留。供应商/正式部署Not Run。

## report和授权协调者必须继续
1. 在本Run根 `pr.md` 写真实任务/证据/限制，不复制旧PR27；handoff只列本Run实际非空文件，目录用证据索引而非空目录占位。
2. 由合法Connector形成新PR，记录确切head、base、变更、QA候选摘要。仅说明/本Run验收证据新增不改变147件摘要；任何产品/测试/锁/规范被改须更新候选并新tests→QA。
3. 审查后按已有用户授权精确合并已验提交，记录merge/main SHA与source摘要映射。不要止于草稿PR，不恢复旧Run/Issue25，不扩大权限。
4. 新main通过 `Deploy Model Relay locally` 正式prepare，再Owner正式workflow_dispatch `deploy=true target=preview`。prepare须独立工作树原入口真实通过；失败保留新运行日志/plan/归因，产品问题返development，不跳route/browser/历史红分类器。日志同SHA覆盖问题归Platform维护源，不改本仓控制器。
5. 真正Deployment成功后，记录environment_url、实际healthz.version对应main SHA及schema3；使用真实管理员页面完成登录/当前功能烟测，确认新版本页面与后端一致及存量数据/原密码保持。不得init/reset安装/手改DB/临时替代服务。没有这些回执，不把地址称为已上线。
6. 完成后更新T001/M01最终状态并给出实际证据路径；阶段放行不是整个用户任务完成。若凭据/外部权限真实缺失，由责任节点明确等待，不让QA伪造结论。

## 剩余风险与不可回退
维护者追加的独立评估见 `acceptance/summary-assessment-qa-21.md`，具体安全历史/源码记录 `acceptance/evidence/qa-21/summary-history-review.json`、`summary-source-review.json`、`summary-focused-contract.log`。结论是历史Open/根因未知，原断言保真，P2验证可靠性风险不阻止本候选进入精确prepare；没有复现修复/关单事实。补充报告第4节给出若再失败返development的受控真实GET屏障最小复现、同scope post-write渲染完成同步契约与邻近回归要求；不得凭静态候选解释先改产品或放松assert。本次没有再次运行完整门禁或独立动态浏览器。
执行18summary原断言一次失败原因未定位，本轮只是安全诊断变更且三个宽度原路径绿；保留P2可靠性风险，不说已经根治。正式prepare再红需携带具体summary_cell/failure_boundary/content_loading与原层log/state返development；不能轮换猜测修复、增限或改预期。budget-clock-r1零软预算/有效UTC/迁移不变量、全部UI/HTTP门禁及依赖锁保持。

## 权威产物
当前报告 `acceptance/report-qa-21.md`；矩阵 `acceptance/matrix-qa-21.md`；旅程 `acceptance/journeys-qa-21.md`；缺陷 `acceptance/defects-qa-21.md`；证据说明 `acceptance/evidence/qa-21/index.md`；逐件归档 `acceptance/evidence/qa-21/archive-index.json`。旧QA6/seq5/8/10/12/14/16/18仅历史，原失败/中断保留。
