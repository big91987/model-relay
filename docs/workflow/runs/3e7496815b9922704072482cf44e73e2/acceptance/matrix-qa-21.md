# 执行21 权威验收矩阵

源：用户Issue28/T001强制AC、同材料design-v0.2.0 docs/acceptance.md与docs/interactions.md；版本147件f8befd，HEAD仍旧base。证据路径以本目录为起点，H=`evidence/qa-21/host-seq20`、Q=`evidence/qa-21`。结果只记实际证据层，未把UI/API/外部事实合并成一个Pass。

| ID | 基线要求 | 角色/范围 | 起始状态 | 连续用户步骤 | 最终事实 | 证据类型 | 结果 | 证据位置 |
|---|---|---|---|---|---|---|---|---|
| COLD-01 | 无模块固定入口全门禁 | 正式宿主Connector20/私有候选 | node_modules/cache absent、NODE_PATH unset、无手装 | make verify自身创建隔离候选→锁安装→全部门禁→probe→清理 | exit0/前后147件相同/依赖本地/自有组不存在 | Contract | Pass | H/verify.log、H/state.json、Q/archive-index.json |
| ORDER-01 | 消费者前依赖/失败停止 | QA测试fixture | 临时Makefile/私有fake命令，仅测失败传播 | 两个消费者与全入口顺序；安装失败禁止消费 | 断言保留原全部门禁并非fixture冒充真实安装 | Contract | Pass | Q/independent-go.log；真实冷安装另见COLD-01 |
| QA-GATE | QA实际完整固定入口 | QA沙箱 | 私有冷候选 | make verify→历史实际HTTP前置 | listener unavailable非business red；清理闭合 | API-E2E | Blocked | Q/qa-native/verify.log、state.json |
| TOOL-01 | 根/父/最终链接不可越界 | QA临时目录 | 根外和根内文件/链接 | copy边界→拒绝→失败清理 | 不读取复制根外，原层回归绿 | Contract | Pass | Q/independent-go.log，原红verification/development-13-red.log |
| TOOL-02 | 辅助两路诊断与独立JSON解析 | QA私有子进程 | 失败stdout/stderr及分块路径 | evidence/probe失败→两路采集→state | 原原因保留且脱敏，机器结构独立 | Contract | Pass | Q/independent-go.log |
| TOOL-03 | 有界取消/监督/落盘 | QA私有helper | 各阶段运行子进程 | cancel/SIGTERM→退出/自有组清理→state | 监督不无界、失败现场可恢复 | Contract | Pass | Q/independent-go.log |
| UI-J01 | 首次空配置AC02–06/12/16 | 管理员/受邀人员各自context | 临时空安装 | 见journeys-qa-21 J01 | 创建与首次调用真实完成 | UI-E2E | Pass | H/browser/continuous-{1280-62042,1440-62062,390-62137}.json、H/verify.log |
| UI-J02 | 授权交集AC03/06 | 租户管理员 | J01Key | 收窄→拒绝→修正→新调用 | 无越权自动重放 | UI-E2E | Pass | 同三件continuous收据/J02 |
| UI-J03 | 轮换退出AC07–08 | 租户管理员 | 历史Key | 过渡→关闭→撤销/期限 | 秘密清理、共享用量/历史 | UI-E2E | Pass | 同三件continuous收据/J03 |
| UI-J04 | 额度恢复AC09–11 | 额度管理员各自context | 非零已用 | 耗尽→调整→重新发送 | 不清零、拒绝无上游 | UI-E2E | Pass | 同三件continuous收据/J04 |
| UI-J05 | 租户生命周期AC04–05 | 管理员/租户用户 | 两租户 | 暂停/恢复/成员权限 | 全Key停止与另一租户独立 | UI-E2E | Pass | 同三件continuous收据/J05 |
| UI-J06 | 价格/账单AC12–15 | 管理员/财务 | 历史测试周期 | 价格快照→结账→更正→导出 | 原金额/明细/范围保留 | UI-E2E | Pass | 同三件continuous收据/J06 |
| UI-J07 | 操作导航AC08/15/17 | 当前租户人员 | 已有请求 | 分页/scope/读错/刷新 | 迟到隔离、秘密不恢复 | UI-E2E | Pass | 同三件continuous收据/J07 |
| UI-QA | QA亲自复跑J01–J07/真实页面 | QA沙箱 | HTTP监听前置受限 | 固定入口未到browser | 没有浏览器执行成功事实 | UI-E2E | Blocked | Q/qa-native/verify.log；同轮等价用户结果替代为UI-J01–07宿主行 |
| HTTP-AC | 权限/预算/UTC/SSE/故障AC02–17 | Go完整实际HTTP测试/受控上游 | 测试私有安装与端口 | 并发/边界/拒绝/恢复/跨租户/终态 | 零软预算、有效UTC、期限/快照/计量守恒 | API-E2E | Pass | H/verify.log，TestConsoleReviewZeroSoftBudgetActualHTTP/ClockRollbackActualHTTP等 |
| DATA-AC | 用量/价格/账单与拒绝守恒AC09–14 | 隔离Go及continuous API读回 | 已知/未知attempts与额度 | UI步骤后读回核对 | 不把未知当零，不重复扣计、历史不变 | Data Reconciliation | Pass | H/verify.log、H/browser/visual-activity-{1280,1440,390}.json |
| HISTORY | 不放宽历史红分类器 | 宿主隔离重建旧版本 | 明确旧版快照 | 旧HTTP红断言→分类器→当前HTTP绿 | 预期业务红不接受bind/缺包/超时作替代 | Contract | Pass | H/verify.log；TestHistoricalAssessment |
| MIGRATE | AC18旧身份/额度/历史保留 | 宿主真实CLI/受控UI | 隔离schema1/schema2 seed | 升级1→3/2→3→三视口身份/Key/历史 | 正式产品CLI升级守恒；不是正式Owner部署 | API-E2E | Pass | H/verify.log migration阶段及TestMigrationUpgradeTargetContracts；没有独立migration JSON收据 |
| CLOCK-WORKER | 顶层私有worker测试 | 完整Go测试顶层 | 无私有state env | 跳过仅子进程入口 | 顶层没有执行 | Contract | Not Run | H/verify.log SKIP；同轮ActualHTTP/InProcess要求真实worker PASS作不变量替代 |
| UI-SUMMARY | 三宽度实时摘要/上下文动作 | 平台管理员当前tenant | J01–07真实对象 | 模型启停及Key/租户/服务上下文→原摘要断言 | API值一致、动作完成；未证明历史偶发根因消除 | UI-E2E | Pass | H/browser/management-summary-{1280,1440,390}.json |
| UI-FAULT | 未知写/重复/取消恢复AC08/16 | 当前角色独立contexts | 受控丢响应/读错/held GET | 页面签发/轮换→安全读回、双击/切编辑 | 单mutation/dispatch、无秘密重放 | UI-E2E | Pass | H/browser/fault-{1280,1440,390}.json、editor-interruption-*.json |
| UI-VISUAL | AC01十页信息层级/窄屏/弹层 | 当前tenant平台管理员 | 连续旅程非空对象 | 原页面/详情/编辑键盘及ready/overflow/secret检查 | 3宽度十页review通过 | UI-E2E | Pass | H/browser/visual-{1280,1440,390}/review.json |
| UI-SAMPLE | 原型与当前截图对照 | QA只读 | 同轮截图 | 查看1280overview/390models/1440key editor | 辅助信息层级核对；不代表QA完成旅程 | Static Review | Pass | H/browser/visual-1280/overview-page.png、visual-390/models-page.png、visual-1440/keys-editor.png |
| SUPPLIER | 真实供应商响应/费用 | 用户真实供应商 | 无实际凭据/响应 | 未执行 | 不能由受控服务推出供应商事实 | External State | Not Run | report-qa-21.md第5节 |
| RELEASE-H07 | 精确PR/merge/部署 | 授权协调者/Connector | 未发布的当前候选 | 新PR→精确merge→正式prepare/Owner deploy preview→health/admin | Deployment/environment_url/版本/schema3/旧数据密码一致尚未证实 | External State | Not Run | release-handoff-qa-21.md |

矩阵中的花括号是可读的同组实际文件缩写，不作为handoff artifacts占位。所有具体路径及sha由Q/archive-index.json列出。正式上线不能由COLD-01或UI-VISUAL行代替。
