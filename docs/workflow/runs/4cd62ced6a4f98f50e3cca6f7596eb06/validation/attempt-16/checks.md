# development执行16：Key最小诊断候选

## 状态和责任

版本：implementation-r1 + commercial-comparison-r1 + migration-driver-r3 + service-check-r1 + model-driver-r1 + key-diagnostic-r1。本轮是诊断候选，不是Key产品修复、完整验收或交付。T001保持In Review、M01 Ready for Review；关键路径是当前指纹完整宿主`make verify` → 失败则development修复/复验，exit0才进入独立QA → Connector发布。没有需要用户澄清或requirements/design返工的已证问题。

## 材料承接与规范

同Run根目录、同prepare manifest及材料root，不恢复旧Run/旧工作区。archive SHA256 `5f51d35fdc15ed7214749f9413879b4238097f1d6b6bbe2cc0d508d8e103cef7`，版本design-v0.2.0，状态frozen-for-test-implementation；本轮重新核对archive和19文件，均一致，实际引用路径/文件摘要见material-integrity.json。重点承接PRD/AC06–08的一次秘密/代际/交集和AC16的真实普通/SSE、取消、同ID详情，interactions的J02及统一弹层清理；decisions不改变。未执行包内脚本、未改冻结输入或擅自增加New API/LiteLLM范围。

恢复architecture-r1、implementation-contract、verification-design、公共T001/总览、完整AC/H与商用功能/J01–J07矩阵以及attempt14检查/失败/来源。已恢复seq15完整末段，旧候选exit2详见host-seq15-summary.md。原生实际模型只能由平台核对；配置声明Codex/gpt-6.1-sol不当作本节点自证。

Trellis实际脚本/spec已有源码依据，`get_context.py --mode packages`返回single-repo/backend/frontend。prepare initialized:false是历史回执，不是本轮重新初始化成功声明；未重建模板。before-dev读取frontend index的DOM/lifecycle/verification/workspace规范、共享reuse/cross-layer guides；delivery Skill及公共契约/审查清单按既有单任务结构裁剪。update-spec新增Key粗粒度超时诊断契约。

## 变更边界与已证事实

行为缺口是固定浏览器驱动一个key stage覆盖多个异步/网络等待，原失败没有可归因的动作边界；所有者为tests/browser/console.mjs。仅改该驱动、workspace.test.mjs补充投影测试、相关Trellis规范及公共记录。Go、产品UI、HTTP/SQLite契约、依赖及Makefile不动。原动作/定位/断言和三视口仍在原入口，未加sleep或重试。

新增安全证据：Key导航/编辑/模型选择/签发/一次值/转交/调用/同ID详情11个阶段；catalog/members/Key POST/call/detail只保留数字状态；DOM只返回固定page枚举、布尔清理/故障状态和无效控件/选项计数。原异常正文、账号/租户/Key/Request ID、URL、响应体、notice/editor文本、提示词/回答均不输出。一次秘密只返回是否存在布尔值，不输出值或长度；page错误/请求失败只计数。页面取证失败保留snapshot_available:false，不制造DOM成功。

待证假设仍包括编辑依赖/模型选项、签发结果、秘密转交/表单校验、调用或详情读取；不能用源码的dialog.close/navigate顺序推定原生close事件时序，也不能用VM证明真实浏览器/网络。先拿宿主新stage/status/DOM facts，再选择修复。

## 已运行与限制

- 定向`node --test tests/browser/workspace.test.mjs`14项exit0；三新增回归执行实际snapshot/observer源码，分别验证秘密承载数据只投影白名单、缺控件安全输出及路由状态不读取body/URL/error。属于诊断契约测试，无native或后台通过声明。
- 首批串联检查exit0：diff/format/module/vet/frontend（140项）/build/两JS语法；non-listening-checks.log保留。追加observer回归后最终复验frontend（141项）及同组检查结果保存non-listening-final.log，以最终日志为准。
- 稳定沙箱bind限制首失败见attempt-01/make-verify.log，Chromium启动permission_denied见attempt14/local-browser-probe.json。环境未相关变化，本轮未重复必失败完整入口或Chromium启动。宿主seq15失败是另一事实，不能归到这些限制。真实Go HTTP/race/native三视口的完整当前版本结果仍由已配置tests Connector取得。

## Trellis-check与公共交付审查

已审查实际流程：keyEditor读取catalog/members → edit提交tenant Keys → 后台一次秘密 → secretDialog显式转交 → callForm → requestDetail；本轮只给测试事实观察分层，无跨层写契约变化。observer无响应体访问，snapshot无日志原文/秘密输出、无页面修改，捕获失败也不丢原stage。无新增产品依赖/忽略异常放行/无关重构/权限扩大/Git操作；局部检查不代替完整门禁。

维持原全部AC01–AC18/H01–H08、八类商用能力和J01–J07的实现/局部/宿主/独立QA/发布五层矩阵。下面当前源码指纹只用于tests绑定，不用于判Pass。真实供应商/外部身份/支付Not Run，独立QA及发布未通过。交付审查仍Blocked by Validation（退出门槛），但tests有合法可执行路由，不需用户等待。

## 下一责任

tests对source-evidence.json同候选保留前后指纹，完整`make verify`，不另改源码或trusted gates。失败保留原日志及新增安全JSON，按具体stage/status/控制状态返回development，不能把诊断候选称为已修复。只有exit0后才交独立QA，并携同manifest/matrix/report及历次失败；QA仍独立对照材料、版本及全部原断言，不把迁移/旧UI局部通过挪用到新控制台后续旅程或供应商事实。

最终来源：119文件，Go go1.25.6，HEAD `98511771871cf0951ecef716bbb55deb13e18da5`；工作树SHA256 `66b873dcdae343c12265140898fddf3fed9d5c298db1b89b507d6d806d029995`。在matrix/spec最终更新后生成source-evidence.json，再生成一次通过cmp完全一致；此后只补摘要范围外的回执。最终diff-check/两JS语法及attempt16路径/令牌模式扫描exit0（未发现本机绝对路径或Bearer/完整Key值）。完整当前宿主门禁尚未运行，本轮没有Key根因已修复的声明。
