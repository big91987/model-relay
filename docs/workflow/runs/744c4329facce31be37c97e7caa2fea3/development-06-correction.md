# 维护者纠正后的资源/关键路径归因与 blocked

2026-10-14，Issue #44，Run `744c4329facce31be37c97e7caa2fea3`，研发执行 6 同一会话接续，Workflow revision 2。先核对现有根 AGENTS.md、冻结身份、源码 diff、执行 5 主 state/log 和执行 6 既有材料；没有重建工作区或恢复旧 Run。

## 撤回此前预算修复及交接状态

按维护者纠正，**撤回执行 6 的全部验证器源码修改**：`tests/verify/main.go`、`tests/verify/main_test.go` 已与当前冻结 HEAD `1d581652736fbac4e8862b70367181c8a9f2f8b2` 原字节相同。累计门禁仍为原 `context.WithTimeout(ctx, 12*time.Minute)`，原断言、门禁顺序、进程组清理与辅助期限均不变。没有新增 sleep/retry 或执行全套重试，也没有改 CPU/内存限额、冻结准备、产品业务、锁文件或数据库。

此前 24 分钟补丁、定向回执、源码摘要 `79467e60...7316c` 和执行 6 原记录仍以原字节保存于 `development-06/`，**仅为已经撤回的历史尝试，不是当前候选，不是修复验收**。该历史 SHA256SUMS 指向的旧验证器哈希不再等于当前源码；不能将旧清单验证失败误判为归档损坏，也不能用它证明当前同版。当前清单由本纠正目录另存。原三份候选测试源码仍原字节，执行 4 的 candidate 归档模块边界保留。

此前 handoff tests 返回“not accepted: user input pending delivery”，因此没有接受该交接、没有本次 24 分钟版固定 tests 执行；现已先处理维护者输入，不重试那个交接。本记录正式取代执行 6 原记录中的继续 tests 建议；不能把原建议写成平台已经完成交接。

## 两层完整回执与主 attempt 身份

维护者通过正式 output API 读取执行 5 外层全文至 EOF，提供 exit=2、1075469 字节、SHA-256 `46e19fdbf713e234a00fadde4f3c88d0561a96b4e7b3849e8d49bae88f347f8d`、truncated=false。此事实以维护者回执保存于 `development-06-correction/maintainer-output-receipt.json`，不是研发自行读取 API 全文的声明。

研发独立全量读取并重新核验主项目回执 `20261014T154026Z-2753749201` 的 state 与 log：1054470 字节、SHA-256 `8361a3d64091f4557be14e80d436754d016bccfa1395b4f91bf3e041e1265735` 与 state 一致，解析 526 个 JSON 记录。原字节仍在 `development-06/state.json`、`development-06/verify.log`，不重复复制或覆盖。

主 CLEAN_START 为该项目日志第 31 行，主 CLEAN_END 为**最后一个** CLEAN_END（第 4188 行）。原包级 race 测试区间第 1935–3836 行包含验证器单测自有嵌套 attempt；其中的 COMMAND/CLEAN/FAILED 与短 deadline 是回归 fixture，不是主门禁时序。`analyze-main.cjs` 明确该区间并保留主 state 的 auxiliary_commands，未用嵌套回执推算主阶段耗时。

## 全部阶段及关键路径

以下行号均属于完整项目 verify.log，可用归档解析器复验。原验证器没有记录每个非浏览器门禁的绝对起止时间，因此这些阶段的全量时长为 unknown；不将包耗时、TAP 耗时、npm 提示或嵌套 fixture 时间伪装成阶段 wall-clock。

| 主阶段 | 开始行 | 实际结论/可用耗时 |
| --- | ---: | --- |
| source evidence | 32 | 进入下一阶段，源码摘要记录完整；单阶段时长 unknown |
| review-history | 34 | 原分类结果通过，进入下一阶段；时长 unknown |
| health-history | 38 | 第 241 行明确 21 method 红例/12 query-recovery 通过；时长 unknown |
| format-check | 242 | 无格式拒绝，进入下一阶段；时长 unknown |
| go mod verify | 243 | all modules verified，进入下一阶段；时长 unknown |
| vet | 245 | 完成后进入 frontend；时长 unknown |
| frontend-test | 247 | 278 pass/0 fail/0 skip；TAP 1634.151786ms 仅是测试执行，首次 npm 提示 3s，不代表该阶段整体 |
| race test | 1935 | 当前七包 ok，relay 包 102.070s；多包并行，不把各包耗时相加作为主阶段时长 |
| build | 3837 | 进入 browser，无构建拒绝；时长 unknown |
| browser | 3839 | 原依赖准备提示 467ms，随后串行阶段如下；整体被总截止取消 |

| browser 阶段 | UTC 起止（2026-10-14） | 结果 |
| --- | --- | --- |
| migration | 15:45:49.114684828 → 15:46:04.523263344 | end ok=true，15.408s |
| legacy-ui | 15:46:04.736472378 → 15:49:02.902224848 | end ok=true，178.165s |
| routing | 15:49:02.902534005 → 15:52:36.969101595 | end ok=true，214.066s；调用有连续成功 end |
| console | 15:52:36.969904929 开始 | editor interruption 有 passed；没有阶段 end；在约 5 秒后被主总截止终止 |
| full-console | 未开始 | 未验，不称完成 |

主 state 全程 `15:40:26.835575262Z` → `15:52:42.63948322Z`，约 735.804s（含门禁外准备/归档清理）。从主 state 起点到首个 migration start 约 322.279s，含门禁外准备，不能全部计为 gate 内。三个已完成 browser phase 标记耗时合计 407.639s；加上阶段间约 0.216s 后，console 已在 main 全程约第 730.134s 才开始。主 prerequisite-npm 于 `15:40:41.613750092Z` 开始、执行 115ms 后完成，其后创建原 12 分钟 gate context；首个主清理 source-evidence 于 `15:52:41.802656755Z` 开始。这个顺序与源码 12 分钟累计截止一致，没有使用嵌套 fixture 的超时估计主 gate。

### 失败分类

- **总量超期：已确认。** 主 gate_exit=-1，末尾 signal:terminated/context deadline exceeded；维护者确认不是 Owner stop，且未到 Connector 外层 1800 秒。不能把它写成用户主动停止或进程 OOM。
- **历史预期红：不是当前产品红。** 26 个 JSON fail 事件全部来自 HealthMethodsHTTP 历史红例（含父测试/包聚合），正式分类标记通过。
- **当前产品断言失败：没有独立证明。** console invitation-1280 的 failed/error_kind=other 诊断在 make Terminated 之后出现，page_error_count=0、request_failed_count=4、snapshot_available=false。它不含足以独立归因某个断言的错误详情，不能据此假定 console 原本正确，也不能把取消后请求失败直接写成产品缺陷。
- **单步卡住：没有证据。** 已完成 migration/legacy-ui/routing，routing 每步有实际成功时序；console 才开始约 5 秒即终止，未发生可由日志确认的独立单步 timeout。没有实证支持删动作、改单步超时、加 retry 或其他设施补丁。

## 当前资源只读采样（非本轮差值）

通过当前进程 `/proc/self/cgroup` 的 v2 namespace `/` 定位自身可见 cgroup，仅只读 cpu.max/cpu.stat、memory.max/memory.events、cpuset 与进程 CPU affinity；未检查凭据值。最终归档快照的 cgroup 读取时间为 **2026-10-14 16:09:07.260Z–16:09:07.261Z**，运行时 CPU 可见性紧接着读取。另有 16:05:02Z 的探索性读取，不是固定执行 5 起点；本记录不计算两次后验读取之间的差，也不充当执行 5 起点。

| 项目 | 最终采样原值/含义 |
| --- | --- |
| cpu.max | `200000 100000`，quota/period=2 个 CPU 等效配额；不是 112 核无限 CPU |
| cpu.stat usage/user/system | `1386224592 / 1158551808 / 227672783` usec，cgroup 累计 |
| cpu.stat periods/throttled | `nr_periods=19511`、`nr_throttled=2254`、`throttled_usec=101809176`，累计节流约 101.809176s |
| cpu.stat 其他 | force_idle=0、nr_bursts=0、burst_usec=0 |
| memory.max | `8589934592`，8GiB，符合任务声明的内存前提 |
| memory.events | low/high/max/oom/oom_kill/oom_group_kill 全部 0 |
| cpuset/进程 affinity | `0-111` |
| CPU 可见性 | Node `availableParallelism()=2`；Node logical CPUs、nproc、getconf 均为 112，后者是可见/affinity 数量，不是可用 quota |

原始样本及只读采样器位于本纠正目录。上述 cpu.stat 是当前可见 cgroup 在采样时的**累计计数**；没有执行 5 前后采样、进程归属或每阶段 CPU 使用/节流数据，包含本 Run 其他活动的可能性不能排除。不得说“执行 5 节流了 101.809 秒”、计算执行 5 CPU 占用率，或把 2CPU/累计节流直接当作本次失败的唯一原因。memory.events 当前为 0，无本环境 OOM 证据；不移用原 Run 2Gi 的事后 OOM 累计值。

2CPU 与当前串行关键路径及历史节流构成环境容量需核查的证据，而不是已完成的性能因果实验；未实测 Go GOMAXPROCS、单阶段 CPU 瓶颈或足够 CPU 的最低数值。现有证据确证当前这次冷路径无法在原 12 分钟总门禁内完成，不能证明单纯增 CPU 必然解决。

## 正式阻塞、影响及外部下一步

当前没有明确、可最小实证的项目测试设施缺陷可修；源码恢复后没有新的绿色同版证据。继续在同环境重跑原全套只会循环求绿，扩大到 24 分钟则违背维护者保留门禁的纠正，均不执行。

正式 blocked：需要 **Harness Builder** 通过标准配置入口核对实际 2CPU 是否符合本验证任务的资源前提，解决资源/关键路径与原 12 分钟门禁的匹配问题。研发不能在容器内改 quota、装依赖、改冻结准备或扩大门禁。下一步应由 Builder 核对标准 CPU 配置及可用运行时并发，若配置不符通过标准入口修正，记录真实配置与冻结身份；保持本 Run 现有仓库和材料，不恢复旧 Run 或丢弃工作区。不能由研发擅自指定已证明足够的 CPU 数值。

恢复后由固定 tests 针对原 12 分钟/原断言候选执行一次原入口，并在实际执行起点和结束记录资源/CPU counter，保存完整 stdout/stderr、退出码、主与嵌套 attempt 区分、阶段关键路径及清理。仅在标准环境已有实质变更或其他新的可复验证据时接续，不在当前状态再次 handoff tests。

影响：本 Run 全量测试仍未通过，独立 QA 缺章拒绝/返工尚未进入；操作者确认、报告、真实同版草稿 PR 均未完成。首轮 11 项清单原字节仍通过，原文档及首轮快照的授权缺章均保留。真实供应商/支付/外部身份 NotRun；没有读取凭据值、提交、push、PR、发布、合并或变更容器限额。
