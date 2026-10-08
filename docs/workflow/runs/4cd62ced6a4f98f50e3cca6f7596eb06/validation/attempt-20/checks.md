# development 执行20 — keyboard-ready-r1

## 承接及变更边界

同Run、main基线、requirements-r1、architecture-r1和完整冻结design-v0.2.0；沿用T001/M01和八类商用能力/J01–J07五层矩阵，不缩为v0.1，也不引入基线外能力。材料archive SHA256 `5f51d35fdc15ed7214749f9413879b4238097f1d6b6bbe2cc0d508d8e103cef7`；prepare指定manifest及19文件校验见material-integrity.json，引用其files根docs/prd.md、docs/interactions.md、docs/acceptance.md的对话框/键盘规则；原型A授权不变。未读取旧Run/旧工作区/供应商代码，未执行材料脚本。

读取研发交付Skill、公共契约与审查清单，按已有单任务公共事实续接；trellis-before-dev实际读取frontend index、DOM/lifecycle/verification/workspace规范、共享reuse/cross-layer指南，get_context.py --mode packages识别single-repo的backend/frontend。现有Trellis脚本与非空源码规范存在；prepare initialized:false为历史回执，不声称重新初始化，不需要模板bootstrap。trellis-update-spec记录键盘归因契约。产品Go、workspace.js/html/css、数据库、模块锁、Makefile不改；只修改console driver、其辅助契约测试、规范与当前Run证据。

最小差距：seq19粗keyboard断言无法定位首次Tab、循环边界、Escape或返回焦点。进一步源码对照确认keyEditor发起Promise.all.then异步打开editor，但旧keyboard.press立即发送Tab、不等待dialog可见；修复位于驱动准备状态，而不修改产品原生dialog。新增延迟打开回归在无等待版本实际失败（regression-before.log exit1 EXPECTED RED），加入locator.waitFor visible后通过。真实行为仍在原生dialog；**仅修复确定的驱动ready缺口，不认定这是seq19唯一根因，不声称产品焦点已修复**。无权限扩大、需求或架构返工、用户澄清项。

## 实际回归与检查

- actual driver新增7项辅助回归：原创建动作/13 Tab/Escape顺序、异步打开准备、四种原始拒绝路径与分阶段、固定布尔投影；最终定向workspace.test.mjs共25项exit0（会话记录），持续检查见non-listening-final.log。初步诊断候选151项/exit0日志non-listening-checks.log保留，最终候选152项结果以final日志为准。这些VM调度不证明Chromium焦点或真实身份/网络。
- 仍执行原先全部断言：第一次Tab包含、12次循环后的包含、Escape立即不可见、准确触发按钮恢复。trace最多16条，只含固定phase、tab_count与5个布尔；不输出控制ID/值/文本/原始异常/URL/账号/秘密。不以额外每次Tab断言改写验收；无强制focus、sleep、重试、API预置或门禁降级。
- trellis-check执行非监听format/module/vet/frontend/build/JS语法/diff检查，最终152项Node全部通过、组内exit0，实际命令及退出见non-listening-final.log（完整文件，不依赖会话显示截断）；旧初步候选检查不替代最终检查。完整make verify稳定原生bind限制未改变，不反复执行必失败的入口；与第一次attempt01对应，不把宿主seq19失败归于此限制。
- 候选来源见source-evidence.json；证据工具范围包含本次测试、规范和验收矩阵，排除delivery/validation防自引用。首次直接go run ./tests/evidence未带Makefile的GOCACHE设置，尝试个人默认缓存被沙箱拒绝exit1；未获得摘要，不能作为来源成功证据。随后按Makefile设置使用工作区.cache/go-build重跑并二次比对（source-consistency.log），不扩大权限或修改代码门禁。HEAD本身不代表未提交候选版本。原生实际模型仍须平台核对，配置声明不是自证。

## 交接与退出

T001 In Review、M01 Ready for Review，退出门槛Blocked by Validation。已证旧转交卡点在seq19越过，但当前全量make verify仍失败；AC01–AC18/H01–H08、八类/J01–J07完整宿主/独立QA/发布均不升级为Pass。真实供应商/外部身份/支付Not Run。

tests对source-evidence同指纹运行完整make verify，保留前后指纹与未截断日志。失败收集keyboard-*及keyboard_focus_trace（最多16条）并回development，据事实区别product/driver/timing/environment；当前版本是驱动ready修正+诊断候选，不要求工具把失败包装为成功。只有当前指纹exit0才独立QA，对同manifest与全AC/H及商用旅程逐项验收，之后Connector发布。不得使用静态browser.check代替实际Go服务。
