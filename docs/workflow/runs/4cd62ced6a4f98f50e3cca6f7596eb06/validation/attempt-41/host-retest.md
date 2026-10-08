# tests交接：先获得真实失败边界，不轮换猜测修复

候选visual-boundary-r1；HEAD `98511771871cf0951ecef716bbb55deb13e18da5`；129文件指纹 `fb766c0aca5f33fbb0960e13286a13461d20999485e93a19a841eb51114f92a4`。材料版本/包/19文件沿prepare，摘要见material-integrity.json。原make verify及全部前置门禁未修改；本次诊断有意**保持原失败断言**，不能假定下一次会绿。

1. 宿主正式执行完整 `make verify`，保留命令/exit/完整未截断日志/当前源戳。不要在原生重试监听、不扩大权限、不重放已完成外部副作用、不改数据库。
2. 失败时按新`console-visual`安全摘要和 `test-results/visual-1280/review.json`恢复 `pending_capture` / `failure`。保存到本Run新的验证产物位置，带文件摘要和实际路径；保留本轮host-seq40副本。不能把Node Contract中模拟的HTTP200/空态当作宿主回执。
3. 若`capture_assertion`仍是ready：检查实际overview GET status/scope-match、title/role/scope、loading/error节点、empty_states（包括不可见节点）、summary/fields/actions/rows、尺寸。只有真实facts证明是已落定合法“近7日暂无调用”，才返回development针对适用空态做最小修复；仍保留全表/摘要/角色/字段/动作/根尺寸要求。不得单凭源码假设允许所有.empty，不造新请求/改时钟隐藏问题。
4. 若是其他断言/导航/详情/editor失败，按明确check处理；HTTP错误、权限/scope、字段和布局分别归因，不能套用空态假设。没有facts要记录null，不能补造成功。由development负责最小driver或产品修复，再重新进入原完整门禁。
5. 仅当同指纹完整门禁真正通过，携本次实际十页review和全部PNG、J01–J07/故障/资源/迁移等原证据交独立QA。旧test-results中有同名/旧PID历史图，不按数量冒充当前63件；以新review引用/hash及同次receipt为准。独立QA逐页对冻结A原型核验AC01，QA38-01保持Open直到它自行关闭。

当前未决：seq40真实capture失败字段/根因；当前候选完整宿主门禁；1280/1440/390十页实际采集与独立QA；正式发布及外部事实。176 Node/语法/vet/build绿色仅为局部。没有需求/架构改动或用户澄清。

入口：公共任务 `../../delivery/tasks/T001-console.md`；十页要求 `../attempt-39/page-mapping.md`；独立QA返研 `../../acceptance/remediation-handoff-38.md`；矩阵 `../../acceptance/acceptance-matrix-38.md`；当前局部结论 `checks.md`；当前产物索引 `artifact-index.json`。
