# development 执行14：model-driver-r1

2026-10-08；Run4cd62ced6a4f98f50e3cca6f7596eb06。候选是修复后的实现与测试，不是完整验收交付。基线HEAD98511771871cf0951ecef716bbb55deb13e18da5，最终工作树以source-evidence.json为准。冻结材料design-v0.2.0/Frozen for Test Implementation及archive SHA2565f51d35fdc15ed7214749f9413879b4238097f1d6b6bbe2cc0d508d8e103cef7不变；先读prepare指定manifest，19文件长度/摘要及archive重新实测一致，version、sha256、实际引用的PRD/interactions/acceptance/decisions路径见material-integrity.json。未执行材料脚本、未改冻结状态/授权、未接触旧workspace。

## 恢复与变更边界

恢复T001/交付总览、architecture-r1与implementation-contract、完整matrix及attempt12失败/检查/来源。seq13完整末段已恢复，见host-seq13-summary.md。源码缺口是驱动的模型label子串选择，不是后台协议或产品缺标签。仅修改console.mjs、workspace.test.mjs、相关Trellis规范及公共证据。原save状态、模型创建与租户授权和后续native断言全部保留；不改Go产品、UI文案、附加路由、CLI/数据、依赖/Makefile或权限。不做无证据时序修复，不用.first()/API预置替代用户旅程。

现有Trellis脚本/有源码依据的spec已存在，get_context --mode packages实测single-repo/backend/frontend；prepare initialized:false为历史回执，不宣称本轮重新受信初始化。before-dev读取实际frontend index/workspace/DOM/lifecycle/verification、共享guides及任务指定契约；本轮无Go变更，不重复整套架构或bootstrap模板。update-spec记录模型label契约。

## 实测与失败保留

- 原驱动+新实际字段/驱动片段回归：node --test tests/browser/workspace.test.mjs exit1，10通过/1失败，匹配2而非1；red-model-driver.log保留并替换本机路径为<workspace>，非删除失败。
- 初次驱动改精确后新VM缺assert注入，测试又报ReferenceError；这是新增fixture绑定错误，非产品/浏览器问题。补入assert后11项exit0，不把中间失败标成功。
- 本轮非监听串联命令全部exit0：git diff --check、make format-check、go mod verify、make vet、make frontend-test（138项）、make build、node --check console.mjs/workspace.test.mjs。non-listening-checks.log保存命令、结果及最终exit0；环境用项目GOCACHE=.cache/go-build、CGO_ENABLED=1、GOTOOLCHAIN=local。
- 本地额外尝试真实Chromium DOM label验证：实际字段source probe两处子串/一处精确，browser launch在页面/网络请求前失败exit1；最小第二次诊断permission_denied=true、missing_executable=false，之后不重试。白名单实测事实见local-browser-probe.json。VM匹配模拟不是Chromium机制证明。
- 稳定原生bind限制仍按attempt01原命令/退出码/来源保留，不重复必失败make verify。本轮宿主seq13 model失败不能归入该限制。完整Go HTTP/native及三视口仍交已配置tests Connector原入口实测。

## Trellis-check / 交付审查

实际modelEditor fields → edit生成label/for控制关系 → console精确locator → 原POST models及PUT tenant-grants；没有改变端到端payload/认证/授权/CSRF/价格/存储。新VM执行实际构造与驱动片段并断言只写alias/upstream/remote三字段，保留routes；通过的是补充契约，不是native事实。frontend-test原glob自然纳入新回归，未削弱固定make verify。细分阶段/允许的错误kind只输出安全枚举，numeric write/check status、remote_label_matches及既有booleans，不打印原异常/正文/秘密。无新增依赖、代码注释、无关重构或Git操作。完整Trellis/native门禁待宿主，不写整体Pass。

T001保持In Review，M01 Ready for Review，全部AC01–AC18/H01–H08及八类商用功能/J01–J07五层状态不升完整Pass；供应商/外部身份/支付Not Run，独立QA/发布未通过，原生实际模型由平台核对。无需requirements/design返工、无用户澄清。tests须保存最终前后指纹并完整make verify，仅exit0后交独立QA；若失败据新分段诊断继续本任务。

最终来源：119文件，Go go1.25.6，HEAD98511771871cf0951ecef716bbb55deb13e18da5；工作树SHA25634f3a82f0c7687194a60e0cbce6af66e6badff55b0b50a373a0690764e6e7fa3。摘要在matrix/spec最终更新后生成，之后只追加摘要范围外回执。最终git diff --check及两新增/改动JS语法检查再次exit0；本轮证据路径扫描无本机绝对路径或Bearer值，失败栈保留<workspace>替代位置。完整宿主门禁仍未通过。
