# development执行10：migration-driver-r3

2026-10-08；基线HEAD 98511771871cf0951ecef716bbb55deb13e18da5。当前候选 implementation-r1 + commercial-comparison-r1 + migration-driver-r3。

## 输入及变更边界

材料design-v0.2.0，archive SHA256 5f51d35fdc15ed7214749f9413879b4238097f1d6b6bbe2cc0d508d8e103cef7，本轮shasum再核对一致。manifest为 `.workflow-input/4cd62ced6a4f98f50e3cca6f7596eb06/5f51d35fdc15ed7214749f9413879b4238097f1d6b6bbe2cc0d508d8e103cef7/files/manifest.json`。实际引用同root的docs/acceptance.md AC18/H05/H06，承接本Run PRD/交互解释、architecture-r1及implementation-contract；不改冻结输入、不执行材料脚本。

before-dev确认现有Trellis0.6.15及源码规范、get_context单仓backend/frontend，无需bootstrap。恢复任务和失败证据，读取frontend workspace/DOM/lifecycle/verification、backend运行/console/routing/multitenant及cross-layer/reuse规范；update-spec维护workspace-contracts。

最小缺口为迁移驱动误读schema3列表包络。所有者是migration-console.mjs断言，不是产品API。仅改items读取和安全分段诊断，新增migration-contract模块/Node回归、console_v3_test实际handler回归，维护公共任务/矩阵/验证。不改产品源码、公开API、CLI、迁移数据、依赖、Makefile、权限、部署或旧断言，不修未经验证的时序假设。稳定原生bind限制未改变，不反复运行完整原生入口；首次失败保留attempt01。宿主seq9不是环境失败，见host-seq09-summary.md。

## 真实检查

以下串联命令exit0，完整日志见non-listening-checks.log，仅脱敏临时目录前缀。使用项目环境GOCACHE=.cache/go-build、CGO_ENABLED=1、GOTOOLCHAIN=local。

```sh
git diff --check
make format-check
go mod verify
make vet
make frontend-test
go test -race -count=1 -v -run '^(TestConsoleKeyCollectionContract|TestConsoleMigrationConservation)$' ./internal/relay
make build
node --check tests/browser/migration-console.mjs
```

135个Node测试全部通过（原127+新8），包括拒绝data包络、缺/多Key、total不符、缺身份及改启停。Go定向race实际SQLite/Go登录/签发/列表包络及schema1/2→3守恒通过；Recorder非监听，不证明网络/browser。链接器警告保留，非测试失败。format/module/vet/build/JS通过。最终Go/HEAD/工作树指纹见source-evidence.json，矩阵/规范变更纳入指纹，delivery/validation按项目入口排除以免自引用。

最终来源：119文件，Go go1.25.6，HEAD 98511771871cf0951ecef716bbb55deb13e18da5，工作树SHA256 bd81e49e060f7e0494e1217947a0e9367fab394b19a64bfe0845b427addbecc6。验证后仅更新公共文档；此指纹纳入最新矩阵/Trellis规范，产品/测试在验证后未变。公开本轮日志13940 bytes，无绝对本机路径或秘密；测试名称中的password仅为描述，不是凭据。

## Trellis-check / 公共交付审查

数据流检查：schema3 replyCollection → items → 两身份/启停/count/total断言；产品UI仍使用原collection，不加替代适配。新回归自动进入固定make verify，DOM行可见、待确认历史、溢出及两迁移目标原断言全部保留。每视口重置白名单status/schema/count；schema验证整数，不打印任意字段/原异常/响应正文/凭据。未加跳过、豁免或门禁降级，无额外依赖/抽象或无关改动。可执行Trellis-check完成，完整网络/native检查待宿主，不标整体通过。

T001 In Review / M01 Ready for Review，完整门槛仍Blocked by Validation；可交tests，非Done/Accepted/Delivered。AC01–AC18/H01–H08及商用八类功能/J01–J07五层状态不缩减。下一步宿主对当前指纹完整make verify、保留前后指纹，仅exit0后独立QA裁决；失败按最小诊断返回development。不需requirements/design返工或用户澄清。真实供应商/外部身份/支付、发布Not Run；Codex/gpt-6.1-sol仅配置声明，实际原生模型待平台核对，不自证。
