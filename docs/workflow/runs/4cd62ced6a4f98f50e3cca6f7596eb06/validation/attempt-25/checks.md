# development执行25：fresh-install-driver-r1

## 来源与范围

承接design-v0.2.0 / architecture-r1、QA22 No-Go和seq24失败。材料沿用prepare实际manifest：`.workflow-input/4cd62ced6a4f98f50e3cca6f7596eb06/5f51d35fdc15ed7214749f9413879b4238097f1d6b6bbe2cc0d508d8e103cef7/files/manifest.json`。本轮实际读取manifest、docs/prd.md、docs/interactions.md、docs/acceptance.md；实际路径/摘要见material-source.log，archive SHA仍为5f51d35fdc15ed7214749f9413879b4238097f1d6b6bbe2cc0d508d8e103cef7。未执行材料脚本，未读取旧工作区或New API源码。

最小差距为fresh installation误当legacy migration。仅改journeys.mjs、journeys.test.mjs、console_v3_test.go及相关规范/交付文档；不改产品初始化、身份、存储、UI或财务，不造legacy数据、不API预种业务对象、不SQL改账、不改冻结材料。原门禁、迁移守恒、J01–J07、视口及后续动作/断言不变。

实际Trellis脚本及源码支持的规范存在，get_context --mode packages返回single-repo/backend/frontend；按before-dev/check读取frontend相关规范、backend当前契约及共享guide，update-spec追加新安装/迁移区分与诊断契约。prepare initialized=false仍是原事实，未改写为成功；现有规范不是待填模板，本轮无须重建。按delivery-review-checklist裁剪为T001持续整改，不新建组织级文档或重复需求。T001 In Review/M01 Ready for Review；退出Blocked by Validation，不Done/Accepted。无用户澄清或需求/设计返工。

## 真实验证

- journey-test.log：node --test tests/browser/journeys.test.mjs exit0、6项，实际freshInstallation源码段拒绝非空tenant、null集合及错误范围选项。expected-red.log复植旧count=1契约，expected-red实验exit0（成功验证预期拒绝），不是重跑旧native候选。
- fresh-install-test.log：GOCACHE=<workspace>/.cache/go-build GOTOOLCHAIN=local CGO_ENABLED=1 go test -race -count=1 -run '^TestConsoleFreshInstallationEmptyCollections$' -v ./internal/relay exit0。真实隔离安装/持久化/登录/handlers，不监听、不是真实网络/Chromium。ld warning保留，临时绝对路径替换为<temp>，未绕过失败。
- checks.log：2026-10-08 06:50:50–06:50:52 +08:00，make format-check、go mod verify、make vet、make frontend-test（158项）、make build、go test -run '^$' ./tests/browser、node --check tests/browser/journeys.mjs、git diff --check均exit0。browser compile无用例执行且cached，不声称native Pass。
- 本沙箱稳定监听限制未改变，不重复已知必失败make verify；历史受限证据仍在validation/report.md。seq24宿主失败属于契约错配，不归入环境限制。完整当前make verify留给tests Connector。
- 平台Codex/gpt-6.1-sol配置声明未独立核对实际运行模型，不切模型。供应商、外部身份/支付、正式部署及Harness缺证仍Not Run/未闭合。QA22-01/02均待复验，不Closed；AC/H及五层商用状态未提升。

## 宿主与QA复验契约

对本轮source-evidence.json指纹完整make verify，保留前后指纹、非截断全日志/exitcode；不得用定向Go/Node替代。原Go/Node/health-history/迁移/真实HTTP/SSE/取消/未知签发/keyboard断言完整保留。continuous-console 1280/1440/390各必须J01–J07七项实际passed；test-results/continuous-{width}-{pid}.json和before-rotation/key-states/before-correction/after-correction/final脱敏截图复制到本Run真实验证目录。

失败保存新stage/action/status/initial_counts及安全DOM投影，先最小归因再返development；未到达不写Pass。全部通过后独立QA按同材料/指纹复验QA22缺陷及AC01–AC18/H01–H08、八类商用功能/J01–J07五层状态。门禁绿不自动代表完整Go、已对齐New API/LiteLLM或已发布。关键路径：宿主完整门禁→独立QA→Connector发布。
