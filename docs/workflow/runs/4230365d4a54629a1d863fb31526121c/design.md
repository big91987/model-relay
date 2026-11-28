# 设计说明：交付工作台退出按钮配色与已证测试夹具修复

关联：Issue #118；需求基线 `requirements.md`（本目录）；目标基线 `70a4f55b5cfd737e0882936ab8f65bbd66fcaed2`（当前 HEAD 已核对一致）。
性质：局部改动，无接口、权限、数据或架构边界变化，沿用已验收范围，不新开方案评审。

## 1. 设计结论

采用 #117 已完整回归通过、#116 已独立 QA 通过的同一补丁版本，原样落到当前主干。不引入新抽象、新状态或新配置项。

## 2. 改动位置与关键行为

### 2.1 退出按钮配色（产品变更）

- 落点：`internal/relay/web/workspace.css` 文件末尾，媒体查询之外（无条件规则）：
  - `#logout{background:#fff1f2;color:#9f1239}`（常态浅红底、深红字）
  - `#logout:hover{background:#ffe4e6}`（悬停底色稍深；移开由浏览器原生 hover 结束自动恢复常态）
- 机制说明：
  - 元素由 `internal/relay/web/workspace.html` 的 `id="logout"` 提供；ID 选择器特异性高于现有按钮类规则，无需 `!important`，不影响其他按钮。
  - 规则位于所有 `@media` 块之外，桌面与 390px 窄屏同样生效（验收条件 3）。
  - 现有全局 `:focus-visible{outline:3px solid #8caeee;outline-offset:3px}` 不改动，焦点描边保持原样；新规则只设 background/color，不覆盖 outline。
  - 退出行为不变：`workspace.js` 中 `byId('logout').onclick` 仍 POST `/admin/api/logout` 并返回登录页；位置与文字「退出」不变。
- 失败处理：纯展示层改动，无运行时失败路径；若 CSS 未加载，按钮功能不受影响（仅退回原配色），不构成可用性风险。

### 2.2 Go 夹具就绪竞态修复（测试基础设施）

- 落点：`tests/verify/main_test.go` 的 `TestFixedRunnerReceipts`。
- 关键行为：
  1. 夹具脚本就绪回执改为「写临时文件 + `mv` 改名」原子发布，消除「文件可见但内容未写完」的竞态窗口。
  2. 断言侧新增 `awaitReadyFile` 轮询（5ms 间隔、5s 上限），等待回执内容完整（`pidReceipt` 校验 PID 合法；候选路径校验非空）。
  3. `timeout_gate` 场景：夹具门限 200ms → 5s（常量 `timeoutGateFixtureBudget`，覆盖宿主机进程拉起延迟观测值 ~300ms），父预算 8s → 15s 保持原有裕度；夹具脚本发布顺序调整为先发布 PID 回执再发布候选回执，保证有界超时时取消证据已就位；`cancel_gate` 则先发布候选回执，因为取消以 CANCEL_READY 为触发。
  4. 回执中记录的门限期望值随之改为 5000ms。
- 明确不变：生产门限 `gateTimeout` = 25 分钟（`main.go`）不动；全部取消/清理断言（进程组不存在、候选目录已清除等）保留。

## 3. 验收条件（直接引用需求第 5 节，设计未改动）

1. `#logout` 计算样式：常态 `background rgb(255,241,242)`、`color rgb(159,18,57)`；悬停 `background rgb(255,228,230)`；移开鼠标恢复常态。
2. Tab 聚焦后 `#logout` 满足 `:focus-visible` 且描边为 solid，底色仍为常态色。
3. 配色断言在桌面宽度与 390px 窄屏下均通过。
4. `#nav-toggle` 计算样式保持原值（白底、深蓝灰字），其他按钮不变。
5. 点击「退出」仍调用 `/admin/api/logout` 并返回登录页。
6. 静态断言：配色规则位于媒体查询之外；`workspace.js` 中退出绑定与 fetch 调用保持原逻辑。
7. Go 测试：`TestFixedRunnerReceipts` 全部场景（含 `cancel_gate`、`timeout_gate`）通过；`timeout_gate` 回执记录的门限为 5000ms，断言中的生产门限期望仍为 25 分钟；夹具子进程清理断言保留并通过。
8. 项目验证入口（`make verify`）整体通过，完整回归由固定节点执行。

## 4. 设计引入的验证风险（新增部分，供研发定向确认）

原补丁已一次完整回归通过、独立 QA 通过，以下为与基线核对相关的残留风险：

- **基线漂移**：补丁基于 70a4f55。研发落地前先 `git diff` 核对 `workspace.css`、`workspace.js`、`tests/verify/main_test.go` 与补丁上下文一致；本设计节点已确认 HEAD 即 70a4f55 且 `workspace.css` 尚无 `#logout` 规则，冲突风险低。
- **悬停断言的环境敏感性**：浏览器 hover 断言依赖合成鼠标事件，已有 `page.mouse.move(0,0)` 复位步骤；若浏览器测试环境变动需优先怀疑此段而非产品代码。
- **夹具轮询上限**：`awaitReadyFile` 5s 上限在极端负载宿主机上仍可能不足；属测试基础设施可调参数，不影响产品行为。

## 5. 架构核对

- `architecture.json` 与本变更相符，无需更新：配色改动位于 `console` 模块（`internal/relay/web/**`）内部，无跨模块依赖；`tests/**` 不在 `source_roots` 内。
- 未新增包、接口、权限、状态或异步机制，符合最小必要设计。
- 旧架构盘点发现按新流程留在架构审查报告，不进本次范围（需求第 4 节）。

## 6. 交接结论

设计就绪。研发任务：按基线核对并复用既定补丁，执行定向检查（补丁上下文一致性、`make verify`），完整回归交固定节点；随后走架构审查、独立 QA 与正式交付。
