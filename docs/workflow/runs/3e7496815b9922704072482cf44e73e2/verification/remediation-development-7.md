# development7 宿主补证契约

## 结论与边界
2026-10-09：QA6返工成立，发布仍 No-Go。保留产品候选142件摘要 `8e82aaa123df25a4758a1939fb2d018713831507aee337e18dfa6100c49f619e`，HEAD `cd413b5aa24a372a19c19af999a6ef61b5c41572`。不改业务、锁版本、测试期待或平台控制器。原始缺包红灯及原生监听失败继续保留；不重跑稳定受限的完整入口。

本轮新增仅为本Run宿主证据采集辅助脚本和公共任务更新。before-dev确认实际Trellis源码规范、packages单仓上下文，无需初始化/bootstrap；没有新的产品规范变更，不制造新架构事实。

## 正式宿主执行
由 tests Connector 在原候选checkout运行：

```sh
bash docs/workflow/runs/3e7496815b9922704072482cf44e73e2/verification/host-clean-verify.sh tests-8
```

脚本不替代固定入口：仅准备独立临时源码副本，实际测试命令仍是原 `make verify`。副本来自Git非忽略文件（含候选未提交修改和新增文件），不复制node_modules、缓存、数据、凭据或材料包；通过原仓Git只读历史支持历史重建，未创建分支/提交，也不写Git历史。开始时断言node_modules和本地npm cache不存在，unset NODE_PATH/NODE_OPTIONS，不手工安装；验证前后源码摘要相同，安装后两包1.59.1必须解析到候选node_modules。日志由文件重定向收集，非tail/工具投影，原make退出码与采集检查退出码分开保留，失败同样写入回执。

若配置的tests命令不支持此准备包装，tests须采用等价宿主隔离准备和采集：不能仅在已有node_modules工作区再运行make verify声称干净，也不能把本脚本的静态检查当宿主实测。无法执行则报告配置缺口，不伪造通过。匹配Chromium/系统库和包源沿用现有宿主前置；不另安装全局模块、扩大权限或降低门禁。

## 待宿主产物（不是当前已有证据）
脚本将写本目录 `host-clean-tests-8.log`、`host-clean-tests-8.state.log`、`host-clean-tests-8.sha256`。这些文件尚未执行生成，不列入本节点artifacts。浏览器产物会保留在原checkout忽略的 `.cache/host-clean-tests-8-test-results`，不是可交接证据；tests须审查路径/凭据脱敏、补充实际Chromium前置版本，选取必要浏览器白名单证据复制到本Run实际文件，并更新索引。不要仅传缓存目录或临时目录（临时源码结束会清理）。新的完整日志必须覆盖历史红分类器、frontend193、Go/race/budget-clock-r1/有效UTC/迁移/build及真实Go UI原门禁。

seq5平台保留111871bytes、truncated=false、exit0，但没有当时干净起点证明。完整历史日志本Run副本仍待tests从受信留存导出；不要把本轮读取投影/审查摘要标为完整日志，也不要改写seq5前置。所有失败/旧证据保留，新的独立运行才可补闭干净AC。

## 交付审查与下游
T001继续In Review，M01为Blocked；本次 Ready for Dispatch（宿主补证），不是Accepted/Delivered。研发可执行静态与定向检查见 development-7-check.log；完整宿主执行未在Agent沙箱进行。QA6矩阵、报告和缺陷是后续QA审查基线，无需重做需求/设计。

trellis-check：定向race顺序/失败停止、格式、diff、辅助脚本语法均exit0；原生隔离副本预检实证无node_modules/本地npm cache且同142件源码摘要，仅为Agent准备检查，不是宿主完整验收。UI/API/DB/预算/时间契约未变，跨层业务流检查不适用；复用原source-evidence和make入口，无新依赖、全局模块或弱化原断言。seq5按offset 0/32768/65536/98304续读至111871 EOF，工具声明truncated=false，但展示投影有缩略；未据此手抄完整文件，日志导出缺口仍明确保留。

tests真实补证后交独立QA；QA通过才进入获授权的新PR/精确已验提交合并/正式Owner preview部署。Deployment/environment_url、healthz version/schema3和真实admin页面未有新回执；旧数据密码不重置、不手改库、不替代服务。原生QA浏览器Blocked与真实供应商Not Run保持，受控测试不得升级为供应商或部署事实。
