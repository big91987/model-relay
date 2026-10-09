# 执行37：丢失绘制帧诊断的过滤修正

## 结论与版本
- 仅修已证诊断配置缺口并补最小身份形状观测，不是空图修复或完整交付。T001 In Review、M01 Needs Revision、QA14 No-Go保持；QA58-01、完整三视口/self_hosted与独立QA未闭合。
- HEAD仍main基线 `3f9cee1615e87e002a2c2300f2d29778072ec3f5`；宿主执行36的154件候选 `d885da49ab2fe52cbc062c56dbd67293078c3fa8ed45235f776159b3572d519c` 前后一致。当前154件候选 `cd3175b1ec0a5254aa6460274b3921b56d8cf4639d0485d95e2603280d4d9e45`，本地检查前后一致，见 `development-37/source.json`、`source-after.json`。未切分支/提交/推送/发布。
- 沿用design-v0.2.0，archive SHA256 `5f51d35fdc15ed7214749f9413879b4238097f1d6b6bbe2cc0d508d8e103cef7`。manifest实际路径 `.workflow-input/b7f93c24bc1ea4d0fe182f12aa7622ae/5f51d35fdc15ed7214749f9413879b4238097f1d6b6bbe2cc0d508d8e103cef7/files/manifest.json`；实际引用同root的 `docs/prd.md`、`docs/interactions.md`、`docs/acceptance.md`，逐件SHA与archive核对见 `development-37/material-integrity.json`。保留frozen-for-test-implementation，不改为批准。

## 宿主执行36事实
- 固定make verify：`2026-10-09T06:02:16.03872Z` → `06:13:17.058101Z`，gate2/runner1/Connector2。冷起点、源码不变、owned group absent=true见原state。258Node、原Go race/HTTP/UTC预算前置门禁及迁移/legacy/routing/console通过，1280 J01–J07/summary通过；full-console在visual-keys像素门禁失败，完整1440/390 visual与self_hosted未运行到，不混用其他套件的视口绿。
- 91件原字节文件归档在 `host-tests-36/`，逐件bytes/SHA见 `host-tests-36/archive-index.json`。完整runner日志546664byte，SHA `4c1ca6cd64ae2944172d77cf8d271be6e0bc7201204aaf16a9ea7cb7458a2193` 与state相符；Connector包装日志546846byte是另一层，本档案不声称包含完整包装日志。
- 实际查看原PNG；sidecar记1280×1149、uniform_rgb=true、RGB243/246/250，SHA `78a7b304df77dac78964e27f6571f81f2ac48376e545786f9778a183028bff53` 仍同历史空图。原图和sidecar在 `host-tests-36/browser/visual-1280/keys-right-columns-1.png` 及其integrity文件。DOM存在/生命周期稳定不能替代可用截图。
- trace available/complete=true、data_loss=true；有mainframe SetLayerTreeId，4帧全部unattributed，只有call marker，缺return，ignored_events=6854。最小白名单摘要见 `development-37/host-failure-summary.json`。部分quad显著负X屏外，但丢失/未关联的浏览器级观测不能证明主页面root overscroll。原wheel195ms早于capture-start202ms，不能继续以未投递解释失败。

## 已证缺口与最小修改
- 缺口位于 `observeCaptureFrameTrace(session)`：仅包含disabled-by-default类别不排除默认类别。与宿主147.0.7727.15对应的[官方过滤源码](https://github.com/chromium/chromium/blob/147.0.7727.15/base/trace_event/trace_config_category_filter.cc)（blob `470dbbb42f78c3975920b95cc7143394339ccf5e`）把这些包含项归入disabled列表，普通include为空时未排除默认类别仍启用；显式包含优先于通配排除。原配置仅排除截图，不是真正的窄类别白名单。这解释无关事件可能大量进入，不证明全部丢失或空图由它造成。
- 只改capture helper、所属Node回归及已有Trellis合同：增加排除 `*`，保留原3个明确包含类别/截图排除、512KiB recordUntilFull与无sampling。不是扩buffer/延时/重试；raw trace不落盘。未读取或复用New API代码。
- 原最多8帧新增5个固定 `identity_shape` 布尔值：numeric argument、numeric snapshot字段、hex object/local/global字段存在性。不输出ID/任意键/值，不猜hex或snapshot字段与mainframe等价；仍只有原显式numeric `args.layerTreeId`可匹配，其余unattributed。下次按实际形状找对应运行时源码合同，不用fixture证明真实格式。
- 原单次fullPage截图、原像素/几何/秘密断言、输入/同步/2000ms末列条件及全部门禁不变。无产品/API/SQLite/费率/Token/权限/锁/Chromium/平台变更；Tracing可扰动时序，新绿也不证明历史空图根治。

## 可运行验证与Skill检查
- trellis-before-dev恢复single repo frontend/backend context，读取实际索引、DOM/lifecycle、验证/依赖、workspace与reuse/cross-layer指南；现有Trellis scripts/spec具源码依据，无需bootstrap或重初始化。研发交付Skill沿用唯一README/T001/M01和公共产物/审查契约，不重建冻结方案。
- trellis-update-spec维护过滤/隐私/未知合同；trellis-check复核最小scope、失败身份/清理及原断言。新增形状回归验证非有限/非数字/非hex、同值hex不猜关联、隐私和原numeric匹配；原配置精确断言要求通配排除。测试进入原Node wildcard；fixture只证明配置/序列化，不能证明Chromium实测。
- 75相关Node、完整Go runner/wiring race、format/vet/build、非监听普通/SSE fixture、JS syntax及git diff --check全部退出0，命令/UTC/退出/日志见 `development-37/local-checks.json`。linker原警告保留；一处owned临时路径脱敏并记录原SHA。
- 原生监听环境未改，沿用 `development-check.md`、`native-verify-state.json`、`native-verify.log` 首次固定入口受限事实，不重复必然同处退出的完整命令。局部通过不是make verify通过；完整HTTP/SQLite/原生浏览器归配置宿主tests。
- requirements/design无新增边界，跳过返回修订；当前是可交tests的验证依赖，不是需用户输入的外部阻塞。不放绿、不写Done。

## tests下一步
1. 当前candidate执行原固定冷 `make verify`，不手装依赖/改Connector权限或超时，不增sleep/断言重试/重拍。保留Playwright1.59.1/Chromium锁、源码前后戳、完整安全日志/UTC/退出/清理、原PNG与sidecar。
2. 同空图先查loss、双marker、identity_shape、mainframe匹配；缺失/unattributed仍未知。完整匹配后才比较同trace截图区间内quad/opacity与PNG；不混用LayerTree与trace矩阵布局。不因新绿闭合历史空图或R28-04。
3. 全部原Node/Go race/HTTP/预算UTC/迁移、1280/1440/390 J01–J07/summary/visual及新安装self_hosted网页创建→核验失败/恢复→关联→授权→Key→普通/SSE→同请求日志费用须完整通过，UI写入只能用户动作。绿后独立QA核材料、源码、回执、安全截图及QA58-01。
4. 动态原生独立QA、真实供应商/支付/身份、Owner preview已登录操作分别Not Run。受控HTTP不冒称供应商联调；无需用户澄清，T001/M01/QA14未结案。
