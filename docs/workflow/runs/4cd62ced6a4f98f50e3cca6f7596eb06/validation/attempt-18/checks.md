# development 执行18 — key-transfer-r1

## 承接与边界

同Run、同main基线及architecture-r1；恢复公共T001/总览、implementation-contract、verification-design、AC/H和商用八类/J01–J07矩阵及attempt16失败/来源。读取prepare指定manifest、PRD/AC相关规则、当前交互契约及实际源码，不恢复旧Run或外部工作区。design-v0.2.0 archive SHA256 `5f51d35fdc15ed7214749f9413879b4238097f1d6b6bbe2cc0d508d8e103cef7`，19文件实际校验结果见material-integrity.json，路径保留manifest相对入口。一次秘密不保存、显式转交、退出/切租户/迟到隔离的既有规则不变。

delivery Skill及公共契约/审查清单按现有单任务结构维护。trellis-before-dev读取实际frontend index、DOM/lifecycle/verification/workspace和共享reuse/cross-layer指南；get_context.py --mode packages返回single-repo/backend/frontend。现有初始化脚本和源码规范已存在；prepare initialized:false是历史回执，不宣称本轮重新初始化。变更仅workspace.js、两项现有测试文件、相关spec和本Run文档；Go/后台/schema/依赖/Makefile不改。

## 实际结果

- 宿主旧失败绑定及归因边界见host-seq17-summary.md。没有把宿主失败归入原生环境限制。
- 实际close/transfer回调四项回归：关闭处理早于/晚于转交settlement、重开dialog保护、跨epoch迟到隔离；定向workspace.test.mjs共18项exit0。旧源码重建相同两行行为、相同fixture，两项契约EXPECTED RED，日志regression-before.log；仅受控调度，不证明真实Chromium机制。
- 非监听检查：format-check、go mod verify、vet、frontend-test（145项）、build、两JS语法及diff-check，组内exit0见non-listening-checks.log。外层打印回执时zsh只读变量status赋值退出1，发生在组内检查完成后；没有隐瞒或将该外壳当成功命令。最终同组重新执行使用正确退出处理，结果见non-listening-final.log。
- 固定完整入口仍make verify；保留原Go/HTTP/race/进程/迁移与native三视口门禁。追加key-secret-transfer-value非空布尔等待，不代替原HTTP200、同ID详情和后续秘密清理断言；原生测试不能被VM替代。
- 最终补充同一UI签发值的只读布尔相等断言，不在测试中赋值秘密；所有当前检查再次追加至non-listening-final.log，以最后一组exit0为准。
- 稳定原生bind限制见attempt-01/make-verify.log，Chromium permission_denied见attempt-14/local-browser-probe.json。无相关环境变化，不重复必失败的完整入口；新候选完整检查由tests宿主执行。

## Trellis-check及审查

已对照spec审查敏感值流：签发响应→一次editor input→显式转交的内存变量→当前epoch call-key；关闭只清本dialog，invalidate仍全局清理；旧epoch不能复原秘密。HTTP、权限、金额及存储契约无变化，未引入依赖/日志值/客户端权限或测试夹具产品默认值。update-spec记录执行契约、错误矩阵与回归要求。全部AC01–AC18/H01–H08和商用五层矩阵保持，T001 In Review/M01 Ready for Review，交付审查仍Blocked by Validation（完整退出门槛）。真实供应商/外部身份/支付Not Run；原生模型由平台核对，不把配置Codex/gpt-6.1-sol当自证。

## 下一责任

最终来源：119文件，go1.25.6，HEAD `98511771871cf0951ecef716bbb55deb13e18da5`，工作树SHA256 `b61394208bd1f0c4fd58d466390fbfc794c703c5eafeec47b6fe903ce6bc5bdf`。按项目Makefile相同的仓内GOCACHE运行tests/evidence两次，cmp一致；此前直接go run因默认用户级缓存读权限失败，未产生有效摘要，随后按项目既有缓存约定成功，不扩大权限。最终diff-check及本轮证据的本机路径/令牌模式扫描通过。两次文档批量patch因末尾标题匹配失败均未应用，随后使用实际标题成功；没有漏写公共状态。

tests在同source-evidence.json工作树前后记录指纹并完整make verify，不改源码或trusted gates；失败保留原日志及安全stage/status/control投影后返回development。仅完整exit0后交独立QA，携同manifest、矩阵、report及本轮和历次失败。当前候选未宣称Key原生问题已验证解决、商用对齐、QA或发布通过。
