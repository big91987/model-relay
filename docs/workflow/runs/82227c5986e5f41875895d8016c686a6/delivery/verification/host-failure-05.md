# seq19完成任务回看超时：驱动身份契约修复

2026-10-06；D4/D5；In Review，未Done/Accepted；研发负责修复，独立QA负责验收。

## 恢复和变更边界
承接[seq17整改包](../../acceptance/remediation-01.md)、[原独立QA报告](../../acceptance/report-01.md)、[研发整改记录](qa-remediation-01.md)、PRD AC-09/R6及交互的已完成回看/同请求身份约束。不重开需求/设计，不改原负控、QA历史矩阵或门禁。本轮最小差距在 `tests/browser/routing.mjs` 的完成回看等待谓词，而非Go路由/持久化或产品文案。只改该driver、已有VM回归、前端验证规范与D4/D5公共状态/证据；不改产品源码、按键策略、超时预算、Makefile、Workflow/manifest、平台源/安装/运行数据或历史Run。

## 已证事实与待证假设
- 宿主seq19固定 `make verify`，900秒预算，真实exit2；平台完整脱敏日志83109 bytes、truncated=false。源码[10身份](final-source-10.json)：85文件，`ff456d824299075fe8d98104a2cd51671ca41f08421f18cb1dd053d80efe93c8`。平台存档可用 `read_command_output(seq=19, offset=78500)` 读尾至eof，不能把摘要当完整重跑。
- 实际CLI/Go测试包PASS；新增 `TestConnectionRealHTTPHeaderAndBodyBudget` 的headers/body通过。旧main UI非空升级/旧binary恢复、原控制台1280/390主旅程PASS。新增routing在 `completed task fold and keyboard review 1280` 抛TimeoutError，因此10整体失败；新多上游390旅程和完整收尾不能声明通过。历史health红灯是预期旧版负控，不是本次失败归因。
- 原始[83807选择轨迹](host-seq19-selection.jsonl)直接归档自ignored `test-results/routing-selection-83807.jsonl`，42行、272543 bytes，SHA256 `56c3f6ad698a3fc1041c4a2220c57cf8d130d8c33d461b3a92f8ff3b08b5508f`；最后stage为actual weighted selection1280，目标index4实际index4、原options身份不变/mutation0。它是此前选择轨迹，不是折叠/详情证据，不再轮换选择按键。
- 原完成组依次检查待办0/完成5/default closed、原生展开、详情、返回收起，但共用一个stage。源码的详情等待是 `#record-detail` 的 `hasText: '请求'`；实际 `recordValues` 以ID和时间开头，schema2成功详情呈现尝试/用量/配置，并无必须包含“请求”这一正文词的契约。这是真实driver契约缺陷。不能用向产品添加无需求文案掩盖它。
- 未证：原宿主究竟卡在展开还是该详情谓词；旧catch没有子阶段/详情响应诊断。本轮修正已证的不正确谓词，并细分阶段/最小白名单诊断，不能预报唯一宿主根因已消除。VM不能证明原生details切换、焦点或网络时序。

## 修复、回归和下游步骤
由UI进入概览的真实GET取得服务器 `call_evidence.request_id`；保持原0/5行、默认折叠、原生Tab/Enter/focus-visible、展开和返回收起。先注册精确记录GET再激活原按钮，要求200及响应ID相同、可见正文第一行和查找输入均相同。比原通用字眼断言更严格，不拿最后本地生成ID替代概览证据，不API预置业务步骤。新增 `reviewProbe` 仅计数/open/status/布尔，禁止记录正文、秘密或实际ID；子阶段分开展开/精确动作/渲染/收起。

既有completed-row VM回归扩展：恢复完成任务→保留原按钮节点→页面一为空仍精确GET→schema2未知用量详情保持原ID。自动随原 `make frontend-test` 入固定入口。实际局部检查与最终11版本见[总验证记录](report.md)及[11身份](final-source-11.json)。稳定本机listen/Mach限制未变，不重复必然受限整条命令，不扩大权限。下一步核对11和[模板联合身份](platform-template-provenance-01.json)，900秒宿主同 `make verify` 完整执行、保存新实际exit/源码/完整脱敏日志与PID关联JSONL；再次失败先依子阶段/reviewProbe归因，保留原断言，不猜测性改产品/加睡眠/放宽预期。之后独立QA复验原13AC。DEP02～04/06/07及真实供应商仍Not Run，平台维护源PR不自动合并，发布由维护者/Connector负责。
