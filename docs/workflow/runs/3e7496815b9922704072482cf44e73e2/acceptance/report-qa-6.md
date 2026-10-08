# Issue28 干净安装修复独立验收（执行6）

> 日期：2026-10-09 Asia/Shanghai（独立检查开始 2026-10-08T16:48:23Z）
> 仓库：big91987/model-relay；分支：workflow/3e7496815b9922704072482cf44e73e2
> HEAD/base：cd413b5aa24a372a19c19af999a6ef61b5c41572；候选为未提交工作树
> 142件源码 SHA256：8e82aaa123df25a4758a1939fb2d018713831507aee337e18dfa6100c49f619e
> 模式：最小供应链修复短轮复验，公共固定门禁仍为完整 make verify
> 结论：Blocked；当前不放行发布，返回 development 补充强制验收证据，不要求无依据修改业务实现。

## 1. 结论摘要
Makefile根因与修复正确对应：旧frontend-test在浏览器依赖安装之前加载playwright-core；新共享phony前置先运行锁定npm ci，独立browser也复用该前置。独立顺序/失败停止race回归和原route-cleanup两例均通过，锁未变，未删减原门禁或修改测试预期。tests Connector执行5真实make verify回执exit0，源码戳与独立QA完全相同，193个frontend测试通过，实际Go UI、零软预算/有效UTC和迁移关键段为绿；这不是QA自行执行完整门禁的结果。

但强制AC要求的“独立候选目录在开始时无node_modules、NODE_PATH unset，未预先手工安装”的宿主证明没有随执行5提供。该日志从源码摘要开始，npm ci成功只能证明本轮安装执行，不能反向证明其开始状态；研发旧版干净红灯与另一次宿主绿灯不能拼成完整干净红转绿。QA开始时node_modules已存在，不能补写为干净起点。正式新PR/精确合并/Owner preview部署未执行，不宣布上线。

## 2. 本轮覆盖与权威基线
- 当前用户Issue28修复边界优先；任务及退出条件见 `../tasks/T001-clean-install.md`、`../milestones/M01-clean-install.md`。不重做需求、架构或UI；不恢复旧Run、不重复PR27。
- 材料manifest：`.workflow-input/3e7496815b9922704072482cf44e73e2/5f51d35fdc15ed7214749f9413879b4238097f1d6b6bbe2cc0d508d8e103cef7/files/manifest.json`。archive SHA256为路径中的完整摘要，version=design-v0.2.0，status=frozen-for-test-implementation；QA校验全部19个材料文件，未提升批准状态。
- 实际引用同一材料root下 `docs/acceptance.md`（2792740b1579d614780856a13b6257d192736e1efc0309f1dc41405a665bef7a）及 `docs/interactions.md`（95f3490ef003fa9b19dd6b9e864dd3b0429619b8cd9d0ecec853919d6a12c801）。覆盖AC01–18已有不变量，重点AC09–11/18与H05–07。沿用 `docs/local-deployment.md` 的部署顺序，不以历史部署作为本候选证明。
- 差异仅Makefile、README、新verification回归、Trellis依赖规范/索引；运行时Go、UI、DB、供应商与部署控制器未改。全部verify recipe保持，新增回归自动进入go test ./...。
- 工具链：Go1.25.6 darwin/arm64、Node26.0.0、npm11.12.1；本地安装及lock的Playwright/core均1.59.1。Chromium解析至1217；解析路径不是本QA成功启动浏览器的证据。
- 受控测试Go服务与上游为独立临时夹具，不是供应商。宿主J01涉及平台管理员创建与邀请、租户管理员Key/调用、开发者和财务权限的明确角色交接，不能表述为单一管理员完成全部动作。其余旅程按角色自己的浏览器context继续，详见 `journeys-qa-6.md`。

## 3. 缺陷与进展
详见 `defects-qa-6.md`。原干净安装P1主题处于Ready for Retest：根因已修、定向绿与宿主普通全量绿已取得，但强制干净宿主原路径仍未有证据，不能Closed。新增的是证据缺口，不是已观测到新的运行时失败。未发现需要新增产品修补的P0/P1或应放宽既有安全语义的理由。

## 4. 证据
权威映射为 `matrix-qa-6.md`，具体文件见 `evidence/index.md`。
执行5宿主命令回执：exit0、111871 bytes、retained truncated=false。通过read_command_output分段取回至offset111871/eof=true，并审查源码、安装前置、frontend汇总、零预算/UTC、迁移和browser关键段；终端工具的展示仍会省略长片段。本Run目前没有执行5完整日志仓库副本，`evidence/connector-5-review.md`仅为回执及关键段审查摘要，不冒充完整日志。须补档并核对其完整性；不能以截图或摘要替代完整回执。
本QA实际 `make verify` 退出2，停止于review-history的监听受限；定向Go race exit0；route-cleanup 2/2 exit0。已校验当前材料与源码。宿主visual review三个视口的白名单投影及实际1280工作台截图保留为辅助证据，QA查看该截图：布局、租户/角色、真实用量与schema3标识正常。截图不是本QA新执行UI旅程。

## 5. 未执行与环境限制
- 本QA完整门禁：Blocked by listener sandbox，不是产品Fail。原生UI未到浏览器门禁，保持Blocked，不要求扩大沙箱权限；同版宿主受控UI结果单独记Pass。
- 干净独立宿主完整入口：Not Run（缺执行条件证明，不能推断）。顺序dry-run、失败安装fixture及已安装本地route测试均不替代此AC。
- 外部真实供应商：Not Run；无凭据/真实响应，所有上游绿灯仅受控HTTP/SSE。
- 当前部署的真实admin、Deployment/environment_url及healthz.version/schema3：Not Run。用户所述5545旧版状态未由QA重新访问，不能作为本候选当前环境版本或已经更新的依据。
- 顶层TestConsoleClockRecoveryWorker skip仍是Not Run；同轮TestConsoleClockRecoveryActualHTTP/InProcess为Pass，代码要求私有子进程实际运行并输出CLOCK_RECOVERY_PASS，含重启及备份恢复。这是对应不变量的同轮替代，不把顶层skip改成Pass。

## 6. 阶段边界
本节点只验产品修复与证据，不提交/推送/合并/启动Actions，不修改业务实现、数据、密码或控制器。通用部署日志同SHA覆盖归Platform维护源，本任务不复制修改控制器。正式部署在干净宿主证据补齐、QA放行后由授权协调者/Connector执行，不以草稿PR结束，不用临时替代服务。

## 7. 环境恢复
QA定向回归由testing.TempDir管理隔离目录；固定入口监听失败未激活服务或改变安装。未访问或写入用户数据/上游。Connector生成的忽略node_modules、缓存、test-results保留，不提交；本次仅新增验收文档及脱敏证据。Go默认缓存探测最初受限，后使用项目既有GOCACHE成功核对，不修改产品配置。

## 8. 质量趋势
相对用户报告的干净安装192/191/1 MODULE_NOT_FOUND，本轮宿主frontend达到193/193（原无法加载的测试模块现在执行两例，数量增加不是删除红例），完整公共门禁exit0；顺序及失败停止回归新增并独立复验。预算/时间/迁移/UI既有关键结果未发现退步。尚不能确认在正式prepare原发现环境红转绿，干净安装闭环与正式上线风险仍未关闭；不无限重跑同一沙箱失败。

## 9. Go / No-Go
实现方向可接受、同版普通宿主全量门禁可信，但发布硬条件的干净起点及独立工作树证据缺失，结论Blocked，当前发布No-Go。返回development只补齐可复验的真实宿主契约与完整回执，不为了绿灯改测试、跳门禁或扩大权限。补齐后独立QA再次核对同版；通过再走report及Connector精确发布。
