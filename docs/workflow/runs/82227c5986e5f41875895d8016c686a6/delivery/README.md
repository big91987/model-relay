# 多上游研发交付计划

2026-10-06；交付判断：Delivered（当前产品切片，未正式发布）；负责人：研发／交付节点；评审：独立 QA。

## 目标和边界
依据 [PRD](../product-definition.md)、[架构](../routing-architecture.md)、[交互](../routing-interaction.md)、[ADR](../adr-001-storage-upgrade.md) 和 [验证计划](../routing-validation.md)。交付公共路线2的多上游运营、有界真实路由和可解释控制台；首次使用、普通/SSE/工具协议及安全边界持续回归。不提前实现路线3/4，不复制平台控制器或手改安装／运行数据。受管理Workflow/manifest仅由原标准安装器按授权同步，见[来源记录](verification/platform-template-provenance-01.json)。

## 当前状态和关键路径
Delivered（仅当前产品切片）：版本13宿主seq28原固定 `make verify` 完整exit0，独立QA [report-03](../acceptance/report-03.md) Go；两项合并前Important/P1 Closed，原四项P1本版未复发。D1–D5 Done，M01–M03 Accepted；逐AC范围与联合身份见[矩阵03](../acceptance/matrix-03.md)和[版本锁03](../acceptance/evidence/version-lock-03.json)。原report-01 No-Go、report-02历史Go及返工红控保留，不拿历史绿灯放行新代码。

Connector此前已提交推送HEAD `8ae4e5173aa0dc45c28ad61dbc64df20afe229a6` 并创建Draft PR #19；版本13修复与新证据尚未提交推送，未合并或部署。关键路径：报告收尾→原发布Connector提交推送并更新原PR #19→维护者真实可信部署联调及正式Actions验收；无待领取研发任务，无需重做需求设计。DEP02–04、DEP06–07、DEP05未覆盖边缘及真实供应商仍Not Run。

## 里程碑
| 里程碑 | 可观察结果／退出条件 | 状态 | 依赖 | 退出证据 |
|---|---|---|---|---|
| M01 | 旧安装经显式升级保留身份／历史，能运营多个独立上游及公共模型 | Accepted | G2 | D1/D2 Done；seq21精确旧main非空UI备份→升级→回退；核验和配置API/SQLite契约；matrix-02 AC01–03/11 |
| M02 | 公共模型真实有界调度且同请求尝试、未知用量和停止可解释 | Accepted | M01 | D3/D4 Done；版本13 seq28完整门禁及report-03 AC04–10/12；SSE两项P1原层级关闭 |
| M03 | 两屏连续管理任务、旧UI升级和完整固定门禁可复验 | Accepted | M02 | D5 Done；版本13 seq28真实双屏/旧UI非空升级及report-03全量覆盖复验 |

## 任务摘要
任务文件是状态事实源；本表是投影。负责人为研发节点，指定评审为独立QA，收尾节点根据真实验收同步。

| 任务 | 结果 | 里程碑 | 状态 | 硬依赖 |
|---|---|---|---|---|
| [D1](tasks/D1-upgrade.md) | 保留旧资产的显式schema2升级 | M01 | Done | G2 |
| [D2](tasks/D2-configuration.md) | 多上游与候选权威配置及核验 | M01 | Done | D1契约 |
| [D3](tasks/D3-routing.md) | 有界真实调度、冷却及停止 | M02 | Done | D2契约 |
| [D4](tasks/D4-console.md) | 逐尝试诊断与连续用户操作 | M02 | Done | D2/D3契约 |
| [D5](tasks/D5-verification.md) | 固定门禁及真实旧UI升级证据 | M03 | Done | D1–D4实现 |

## 风险、裁剪和交付审查
- 三个结果里程碑继续原位维护；本轮G4产品切片检查通过，正式发布门槛尚未通过。公共说明见[PR正文](../pr.md)，产品Accepted不等于正式发布通过。
- 取消／提交止损不承诺撤销供应商费用；核验只请求模型列表、不生成，仍保留供应商可能计费提示。未知用量不估算为0。
- 回退只恢复升级前完整备份，不合并升级后新增配置、密钥、请求或停用状态。CLI非空迁移通过不等于可信controller或Actions迁移通过。
- 标准安装及用户提供平台回归按来源分层；平台源PR不自动合并，维护者不得人工预upgrade或改平台DB替代可信activate验证。
- 上游追踪、任务退出、共享入口、指定评审、原位状态及交接审查依据见[收尾审查](verification/report-closeout-02.md)。无需用户重新批准普通阶段选择；无新增需求／设计偏差。

## 历史验证与变化
首轮[QA No-Go](../acceptance/report-01.md)、[整改包](../acceptance/remediation-01.md)及[研发整改](verification/qa-remediation-01.md)保留；不覆盖为当前Go。历次宿主失败分别见[03 CLI](verification/host-failure-01.md)、[04 标签](verification/host-failure-02.md)、[05 选择](verification/host-failure-03.md)、[06 原生轨迹](verification/host-failure-04.md)、[10 回看身份](verification/host-failure-05.md)。seq9预算中断及本机沙箱失败继续保留于[验证记录](verification/report.md)。版本11成功回执见[seq21索引](../acceptance/evidence/host-seq21.md)，独立QA专项与辅助证据在[验收索引](../acceptance/README.md)。不把旧版本局部PASS用于当前放行。
