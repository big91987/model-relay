# development执行8：migration-driver-r2

## 输入与边界

材料design-v0.2.0，archive SHA256=5f51d35fdc15ed7214749f9413879b4238097f1d6b6bbe2cc0d508d8e103cef7。
manifest：`.workflow-input/4cd62ced6a4f98f50e3cca6f7596eb06/5f51d35fdc15ed7214749f9413879b4238097f1d6b6bbe2cc0d508d8e103cef7/files/manifest.json`。
实际引用同root下`docs/acceptance.md`的AC18/H05/H06，承接本Run product-definition、architecture、implementation-contract对PRD/交互的冻结解释；不改材料或已批准设计。Codex/gpt-6.1-sol仍仅配置声明，原生模型由平台核对，不自证。

本次仅修宿主失败的迁移测试目标绑定：新增schema2兼容测试入口，保留原断言，追加schema1/2→3正式CLI/受控HTTP/native验证。产品源码未修改，不改业务数据过测、不增加默认假供应商、不发布Git或改平台。无需求/设计方向性缺陷，requirements/design返工不适用。

## 真实补充检查

Go go1.25.6，原生workspace-write沙箱，`GOCACHE=.cache/go-build`，`CGO_ENABLED=1`，`GOTOOLCHAIN=local`。下列串联命令总exit0；完整输出见`non-listening-checks.log`。仅脱敏临时目录前缀为`<temp>`，保留链接警告和结果。

```sh
make format-check
go mod verify
make vet
make frontend-test
go test -race -count=1 -v -run '^TestMigrationUpgradeTargetContracts$' ./tests/browser
go test -race -count=1 -v -run '^TestConsoleMigrationConservation$' ./internal/relay
make build
node --check tests/browser/migration-console.mjs
```

格式/模块/vet/build/JS语法通过，原127 Node测试全部通过。实际两类Go升级命令及数据库复开目标断言通过，非命令mock。schema1/2→3原守恒定向race通过；此层真实SQLite/进程内HTTP不能证明网络或native行为。

新浏览器未在原生沙箱运行。首次attempt01稳定bind限制仍保留，不反复执行同一必失败完整入口。宿主seq7不是环境失败，见`host-seq07-summary.md`。旧attempt08指纹已过期，当前见`source-evidence.json`；HEAD不能独立识别未提交候选。

最终来源：117文件，HEAD 98511771871cf0951ecef716bbb55deb13e18da5，工作树SHA256 f059c17ab3a1b0edf181d612d2907b0a755ba50605bca737901061abe99739ff。代码/测试在上述补充检查后未变，随后更新的公共矩阵和Trellis规范纳入最终指纹；delivery/validation回执排除以免自引用。`shasum -a 256`复核本Run archive.zip与输入摘要完全一致；`git diff --check`通过，attempt09公开证据未含绝对本机路径。最后仅补充本回执，不改变指纹范围。

## Trellis-check 与交付审查

`.trellis/.version=0.6.15`及源码规范已存在，不重做bootstrap空模板。before-dev承接backend/frontend索引、运行/路由/多租户/浏览器规范、cross-layer/reuse指南；`get_context.py --mode packages`确认单仓backend/frontend。update-spec在multitenant-contracts.md补充目标所有权、接口、错误矩阵及回归责任。

检查CLI→Store→health→native契约，两目标分别验，不将test-only入口称生产部署。对旧身份/Key/请求/映射只读比较，全部源数据由旧UI和正式restore/upgrade产生，不修改记录过测；旧二进制rollback保留。新脚本从实际workspace标签定位，用真实session及Go API判断Key启停/历史风险和三视口。诊断只schema/阶段/状态/受控计数，不输出秘密、对话或数据库原文。不添加依赖/权限、静态根浏览器证明或超范围清理。

仍T001 In Review/M01 Ready for Review，未Done/Accepted/Delivered。唯一任务/矩阵引用本轮证据，未重建PRD/架构。tests须对新指纹运行完整`make verify`；只有真实exit0后独立QA逐AC/H裁决。真实供应商、外部身份/支付、发布均Not Run。无需用户澄清。
