# PR36 联合候选独立验收报告

> 日期：2026-10-10，Asia/Shanghai；Run 637705186406b80a2ee4afaf0d6b2167，QA执行23。
> 仓库/分支：big91987/model-relay / workflow/637705186406b80a2ee4afaf0d6b2167。
> HEAD：8917f942cb38af2f56aa93914ab62e763bf9cc6e；origin/main与MERGE_HEAD：6929376089535038112da48aed715fd729615d76。
> 受测制品：尚未提交的联合工作树，158文件源码SHA256：6e359667d5f7521f6b8c421454c6fc7cbe5ccb8c74aa87bb1e5955c194c202cc。
> 模式：联合受控回归全量门禁＋独立受影响复验；外部供应商部分只读评估。
> 结论：**Go，仅限联合M01受控回归资产交付；#33产品Incomplete-No-Go保持。**

## 1. 结论摘要

tests22在联合候选上完整冷执行固定 `make verify`，Connector、runner、内部门禁均exit0。QA复算材料61项、源码逐文件库存、主线/分支保留关系和完整日志摘要，均与本轮回执一致；没有引用旧tests9/QA10替代联合绿。QA另实际执行8组进程内计量race和53项Node采证/可读性/摘要/冷却回归，均exit0。

本轮没有确认新的生产实现缺陷，也没有修改业务实现。内部安全默认、非零RPM/TPM、并发、计量和隔离已有当前受控HTTP数值证据；PR35带入的三宽首次自托管普通/SSE网页路径同时通过。它们均不是真实DeepSeek、供应商收费或原两屏非零限额网页验收。可进入report并由原publish/pr更新同一PR36，不能关闭#33。

## 2. 版本、范围与基线

工作区非干净，处于受信prepare18未提交整合；无unmerged路径。QA只读Git元数据，未写.git、切分支、提交或推送。QA接手时复算PR35主线3255条变更路径全部与main逐字一致；PR36原分支231条路径仅4份授权状态/计划文档不同，其余保留。随后仅新增QA报告/证据并更新本Run验收索引，不改变业务/测试源码。接手时完整库存见 `evidence/qa-23/integration-recomputed.json`；Go源码、Makefile和测试入口确实共同包含两项交付，不以无冲突替代测试。

权威顺序：本次人工回退限定联合受控交付→原材料PRD/AC/交互→已有implementation-contract §4–5→本Run test-plan A/B/C→当前测试与回执。材料版本product-test33-handoff-r1；ZIP SHA256 a13fb6b9089cefbb9d2b6f9a333830e0ee6866bd7ff45d5ba881da97e9a66a49。原设计ZIP声明5f51d35fdc15ed7214749f9413879b4238097f1d6b6bbe2cc0d508d8e103cef7保留，不声称重建原ZIP。完整材料引用路径和61项摘要见 `evidence/qa-23/material-integrity.json`。未执行包内脚本，未改变冻结设计或待评审材料状态。

验收覆盖租户/稳定Key共享与隔离、模型bounds、普通/SSE、拒绝/失败/取消/未知/重启、价格快照、权限和跨层对账。原安装、迁移、history门禁也实际执行。浏览器由tests22宿主启动真实Go服务，provider及self_hosted分支、1280/1440/390三宽、J01–J07连续旅程通过。各角色使用其独立真实登录会话；管理员向租户交接是显式授权，不声称所有步骤由一个人完成。详见分层矩阵、旅程表和计划独立评审。

## 3. 缺陷及进展

新增/复发/本轮关闭的确认生产缺陷均0，P0/P1/P2/P3均无新增记录。DS-SAFETY-01的“无控件意味着无后台安全默认”推断在内部动态链未复现；真实供应商收费上界/隐藏推理与思考策略仍Blocked，不把候选关闭为已修。

历史tests5/7失败、exec17 Git拒绝、exec21 npm代理EPERM及旧版本证据保留，不回写为通过。当前联合门禁缺证已由tests22关闭；这是交付采证缺口关闭，不是生产算法故障修复。原主线历史失败日志尾空白原样保留；不为整合diff --check清理历史证据。当前QA新增文本diff检查通过。

## 4. 证据与执行来源

宿主tests22时间2026-10-09T16:22:25.516679Z–16:33:19.244284Z，即上海2026-10-10 00:22:25–00:33:19；总约653.728秒。原720秒内部deadline未扩展。node_modules/npm cache起初不存在，NODE_PATH unset，无手工预装；模块、格式、vet、锁定npm ci、278项Node测试（0 fail/skip）、全包race、build、migration/legacy/routing/console/full-console全部运行。relay race包257.388秒，7组ActualHTTP计量顶层用例通过。嵌套runner自测的fixture故障/取消日志是预期测试，不把其中的CLEAN_END当顶层门禁结果；顶层以state.json和末尾对应source指纹定位。

环境Go1.25.6 darwin/arm64、Node26.0.0、npm11.12.1、Playwright/core1.59.1、Chromium147.0.7727.15；隔离临时SQLite和受控上游，非线上部署。五段浏览器耗时12.863/89.580/123.492/15.222/81.021秒。三宽自托管各12步骤，普通/SSE输入3输出4、租户费用21微美元、成本7微美元、同请求一致；这些与metering夹具租户费7/成本33是不同测试价，不混算。

归档 `evidence/qa-23/host-tests22/state.json`、完整 `verify.log` 和全部browser PNG/JSON/JSONL。完整日志959452字节，SHA256 ad3d5808e2f16420ed7ffd0ebda3955d37c1d1932091528e715c91ecccaf8746；加Connector97字节前缀为959549，与正式回执一致。只读取过read_command_output首段；**未声称逐字节独立比对平台全段输出**，判定依据是未截断Connector回执＋完整落盘日志hash/长度＋版本/具体用例证据，不是提示中的省略输出。

QA本人实际执行及证据：`audit.mjs`材料/库存/回执复算；`independent-go.log`8组无监听race 21.991秒exit0；`independent-node.log`53项exit0；`independent-checks.json`记录命令/来源/摘要。`finalize.mjs`独立核对七类关键快照数值（双scope/重复均匹配）及45份PNG integrity的原字节hash；仅将本地Go链接器warning的一处临时绝对路径脱敏，原/公开hash留在local-log-transforms，宿主日志不改。本人抽看390自托管同请求截图及limits页面，结合几何/键盘与PNG采证回归，无新增确认偏差；不是本人重跑浏览器。静态browser.check不支持此后端，未使用其伪造动态通过。

## 5. 未执行和限制

QA沙箱没有重复已知受限的完整监听门禁；本人HTTP/进程/浏览器重跑为Not Run，使用当前同版tests22的实际宿主运行证据，明确来源。供应商普通/SSE成功、成功恢复、自然UTC窗口、DA双会话非零RPM/TPM/池竞争网页、真实费用均Not Run/Blocked。没有读取私有Provider配置、调用供应商或消耗24次预算；历史not_attested、每次64 tokens/90秒和未知attempt不重置。账号中精确deepseek-v4-flash/pro的当前可用性未知，不用网络公开资料推定。

## 6. 阶段边界

本轮只交付联合受控测试资产，不能宣称生产RPM/TPM故障已修、产品可商用或真实供应商已验。公开H07精确提交/Actions/合并/部署由后续Connector承担，目前无新merge SHA、发布制品或部署回执。report必须更新旧pr.md，标题为【产品测试】补齐受控RPM/TPM与安全调用计量回归，Refs #34；不接管#32、不关闭#33。无需求/设计冲突，不重建这些阶段。

## 7. 环境恢复

宿主state：candidate_removed、owned_group_absent=true，owned_process_group_cleanup_attempted；本地补充测试由测试自身临时目录与清理管理，不碰用户数据。QA只保留可交付的无秘密证据，缓存/运行目录仍忽略。未产生供应商工作负载或私有配置变更。

## 8. 质量趋势

相比旧tests9/QA10，受控语义不变；提高的是联合候选证据完整度，PR35首次自托管/可读性与PR36数值计量首次在同一精确候选完整绿。受控资产继续可复验，生产缺陷关闭0、新增0、复发0。约66秒顶层时间余量不能保证慢环境永不超时，不得增deadline或删断言求绿。真实产品覆盖缺口没有减少为已验，仍按原责任方复测。

## 9. Go / No-Go

**联合M01受控回归资产Go；允许整理交付说明并推进原publish/pr。** 强制完整冷门禁和独立QA已完成，当前范围无未关闭P0/P1。该Go只绑定158文件指纹；publish之后必须准确关联实际新提交，源码若变化需要重新验证。**Issue33产品Incomplete-No-Go不变；供应商完整安全/费用与原数值网页验收仍Blocked。**
