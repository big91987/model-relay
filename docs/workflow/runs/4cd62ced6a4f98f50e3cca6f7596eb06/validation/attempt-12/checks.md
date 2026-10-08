# development执行12：service-check-r1

基线HEAD98511771871cf0951ecef716bbb55deb13e18da5；当前候选implementation-r1 + commercial-comparison-r1 + migration-driver-r3 + service-check-r1。T001 In Review/M01 Ready for Review；不是Done、完整验收或发布。

## 承接与变更边界

沿用requirements-r1/architecture-r1/implementation-contract、同Run任务/矩阵及历次失败。材料design-v0.2.0/Frozen for Test Implementation；本轮archive重新shasum得到5f51d35fdc15ed7214749f9413879b4238097f1d6b6bbe2cc0d508d8e103cef7，与prepare一致。manifest：`.workflow-input/4cd62ced6a4f98f50e3cca6f7596eb06/5f51d35fdc15ed7214749f9413879b4238097f1d6b6bbe2cc0d508d8e103cef7/files/manifest.json`，实际引用同root的docs/prd.md、interactions.md、acceptance.md AC01/02、decisions.md及本Run当前契约；完整AC/H和商用目标不缩减，不更改材料/授权/评审状态、不执行包内脚本。

独立本地摘要核对19个manifest文件的SHA256及字节数全部一致，完整版本/摘要/仓库相对路径在material-integrity.json。该核对不执行包内代码、不证明产品验收。

现有Trellis scripts/config及有源码依据的spec已存在，get_context --mode packages实际返回single-repo/backend/frontend；prepare最初initialized=false是历史回执，不能据此捏造本轮受信初始化。未初始化或重建Trellis，无待填规范需bootstrap。before-dev恢复任务、读取frontend workspace/DOM/lifecycle/verification、backend runtime/console/routing/multitenant及shared guides；update-spec维护workspace核验契约。

最小缺口所有者为workspace.js服务行POST：遗漏expected_execution_revision。仅修该字段；Node测试执行实际render action、Go测试使用真实schema3登录/创建/核验handler和SQLite验证拒绝及连接-only语义。console.go受控列表缺object=list据实际协议纠正。console.mjs保留原断言、追加成功/applicable/版本一致及白名单诊断，不删失败或替换原native旅程。不改产品Go协议/存储/CLI、依赖、Makefile、权限、Git发布/部署，不修未经证明的时序。

## 真实检查与失败保留

- `node --test tests/browser/workspace.test.mjs`修复前exit1（8原用例通过、2新用例因{}载荷失败），修复后exit0（10项）；修复前日志red-workspace.log。
- `go test -race -count=1 -v -run '^TestConsoleServiceCheckRevisionContract$' ./internal/relay`首次exit1：实际handler200但受控fixture缺object=list，未假称通过；red-go-check.log保留。只修fixture后，下面定向命令exit0。
- 首次未指定项目GOCACHE的source-evidence尝试因默认用户缓存目录写入受限exit1；改用既有项目.cache/go-build后成功。不是产品/网络通过事实，不扩大沙箱权限。
- 以下串联命令最终exit0，完整脱敏输出在non-listening-checks.log。环境GOCACHE=.cache/go-build、CGO_ENABLED=1、GOTOOLCHAIN=local；保留链接器warning，非失败。

```sh
git diff --check
make format-check
go mod verify
make vet
make frontend-test
go test -race -count=1 -v -run '^(TestConsoleServiceCheckRevisionContract|TestConsoleKeyCollectionContract|TestConsoleMigrationConservation)$' ./internal/relay
make build
node --check internal/relay/web/workspace.js
node --check tests/browser/console.mjs
```

137 Node均通过；实际handler/SQLite核验、Key包络、schema1/2→3守恒定向race均通过。隔离transport和Recorder不是真实HTTP或native证明。原生稳定bind限制未改变，不重复必失败的完整入口；attempt01原命令/退出码/指纹/受限范围继续保留。宿主seq11失败不归入该限制。完整make verify及三视口新控制台仍须宿主新指纹复验。

## Trellis-check / 公共交付审查

核验数据流：upstream列表execution_revision → 实际服务行 → POST expected_execution_revision → 原Go版本guard → GET-only协议核验 → 当前结果详情。未引入管理revision替代或隐式重试，空/不符输入仍400/409、零dispatch，成功不代表供应商。Node与Go精确回归自动进入既有固定make verify；原各native/迁移/assert保留。诊断仅status数值和命名布尔，逐视口重置，不输出原异常/正文/秘密/绝对本机路径。无新增依赖或无关重构。检查通过的是可执行非监听部分，完整Trellis/native门禁待宿主，不标整体Pass。

当前工作树指纹由source-evidence.json绑定，公共matrix/Trellis规范入摘要，delivery/validation回执按既有入口排除以免自引用。独立QA/发布未通过，供应商/外部身份/支付Not Run，实际原生模型待平台核对。无需用户澄清/requirements/design返工。下一责任为tests：保存前后指纹，完整make verify exit0后方可携同manifest、矩阵/报告/历次证据交独立QA。

最终来源：119文件、Go go1.25.6、HEAD98511771871cf0951ecef716bbb55deb13e18da5、工作树SHA256312cae2097f665d7e50ab6ac11290dd54fe444c6c6a37ed5dda233108d8bc101。最终摘要在矩阵/规范更新后生成；此后仅追加本回执，不改产品/测试/摘要范围内文件。Makefile相对HEAD已有的1/1差异属于承接候选，本次修复没有修改它。归档日志已排除本机绝对路径、秘密及对话；源码常量中的controlled凭据仅用于隔离测试，不是供应商凭据。
