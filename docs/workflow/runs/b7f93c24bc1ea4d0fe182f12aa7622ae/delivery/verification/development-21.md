# 执行21：空白截图最小归因诊断

## 实施前边界
- 基线 HEAD `3f9cee1615e87e002a2c2300f2d29778072ec3f5`；执行20宿主前后154件候选戳均为 `566a5eb651efd268740b9d2f876b8e860015da062acc7e3fcd7bee63403f2e15`。
- 恢复本Run总览、T001、M01与执行19记录，继续同一任务；design-v0.2.0保持冻结。材料manifest为 `.workflow-input/b7f93c24bc1ea4d0fe182f12aa7622ae/5f51d35fdc15ed7214749f9413879b4238097f1d6b6bbe2cc0d508d8e103cef7/files/manifest.json`，archive SHA256 `5f51d35fdc15ed7214749f9413879b4238097f1d6b6bbe2cc0d508d8e103cef7`；沿用原任务引用的 `docs/acceptance.md` 与 `docs/interactions.md`，不执行材料脚本、不重新批准设计。
- Trellis目录及源码依据规范实际存在；`python3 .trellis/scripts/get_context.py --mode packages` 返回single-repo，已读取frontend index、workspace、DOM/lifecycle、verification与共享guide及研发交付公共契约/审查清单。prepare初始化false为历史回执，不将空模板或本轮初始化冒充已完成。
- 最小差距：执行20截图前后几何稳定且最后列可达，但keys原PNG全空白、四个锚点零墨；现有receipt只有几何，无法区分CSS/祖先隐藏、页面生命周期、命中遮盖与截图栅格边界。不是执行18的单个裁剪锚点缺陷，不据此归因于滚动、监听或浏览器内部机制。
- 仅改 `tests/browser/capture-integrity.mjs` 与对应Node测试：在原单次截图前后采集有界白名单渲染条件，并审计原PNG全图均匀性。失败仍在原像素/几何门禁抛出原异常；无第二次截图、sleep、重试、raf同步、超时扩大或DOM写入。规范同步到既有workspace-contracts，不改产品、供应商、DB、锁文件、平台或Connector。
- 当前需宿主在原固定 `make verify` 采集新诊断；诊断夹具只能证明序列/输出契约，不能证明真实浏览器渲染或关闭QA14-01。

## 原始宿主事实与归因边界
- 执行20原入口UTC `2026-10-09T02:52:11.896716Z` → `2026-10-09T03:03:04.405587Z`；Connector exit2、runner receipt exit1/gate exit2，源码前后相同，owned process group absent且隔离candidate已清理。91件原runner产物逐字节复制/SHA核对见 `host-tests-20/archive-index.json`，没有修改原state/PNG/sidecar/log。
- 原runner完整日志428447bytes，SHA256 `33d24fd385ac05cfeab2c22233b1b104e569b6dfb905a19e3da436a83cdd230c`；Connector显示428629bytes包含外层包装，不把runner日志宣称成逐字节相同的Connector全文。
- `keys-right-columns-1.png` SHA256 `78a7b304df77dac78964e27f6571f81f2ac48376e545786f9778a183028bff53`，实看原图及全像素审计均为RGB243/246/250。原sidecar四锚点contrast0/dark0，title/content前后位置尺寸相同；表格末列到达、根无溢出、秘密清理断言此前通过。原图不是有内容但采样点错位。详细白名单摘要见 `development-21/host-failure-summary.json`。
- 执行20已执行的原legacy/routing/console三视口与1280 J01–J07/summary日志继续保留；full-console止于1280 visual-keys，当前candidate完整visual三视口与新self_hosted尚未获本次通过。历史执行13行为绿与执行14No-Go不覆盖。
- 已证事实：原像素确为空白、DOM几何稳定、滚动已达；未证假设：CSS祖先/可见性、生命周期/字体/焦点、覆盖命中、截图栅格边界。执行21没有CDP或浏览器内部机制证据，新增白名单只排查这些外部可观察条件；若重现，先比较新sidecar，再选择针对性下一项修改。

## 实施与本地检查
- `captureRenderFacts(page)`不读取文本/属性/URL/凭据/任意CSS串，固定五个scope，每个最多16祖先；观察失败只记available:false。原单次PNG前后同步只增加观测，不等待帧或更改DOM。原图均匀性不参与放行：blank/sidebar仍在原完整像素门禁失败，原异常保留。
- 新4项Node回归进入原 `tests/browser/*.test.mjs` 通配；原相关四文件61项exit0。`make format-check vet build`、非监听Go普通/SSE fixture race、语法与diff均exit0。命令/UTC/退出码见 `development-21/local-checks.json`，原日志同目录保留。
- 尝试本地全Node通配exit1：243项通过，route-cleanup文件因 `MODULE_NOT_FOUND: playwright-core/package.json` 无法加载；不是产品回归、不是监听失败，也不是完整绿。未临时手装依赖、未删除用例或修改锁文件，宿主仍经原make verify冷起点执行npm ci及原Node门禁。原失败日志 `development-21/node-current.log` 保留。
- Go fixture链接器LC_DYSYMTAB warning保留，测试exit0；不涉及本轮测试源码，不修改无关内容。静态/VM/非监听夹具不冒充真实HTTP/SSE网页旅程。
- Trellis-check：前后源码154件戳均 `66cc3cd66fd062d0716ceff32bc9889cc2140145164e0cabc4b116ae3707a3c8`；本轮仅测试观测/回归与既有workspace规范更新。无业务写、API/DB/费用/Token/权限改动，跨层业务改造不适用。原secret guard、一次截图、mandatory anchors、像素阈值、稳定几何、原键盘/scrollend/wheel/2000ms及所有业务断言保持。source-before/after与最终SHA清单同目录。

## 交接结论
- T001 In Review；M01 Needs Revision；不是已修复或已交付。QA14-01/QA58-01未Closed，R28-04历史原因未知，尚无实际供应商/支付/身份联调证明。
- 稳定原生监听限制与环境未变；按首次失败 `delivery/verification/native-verify-state.json` / `native-verify.log` 不重复必败完整入口。当前缺锁定浏览器依赖也未以手装规避；继续交tests原固定make verify宿主冷起点获取本次实际诊断与完整回执。
- tests保留完整安全日志、UTC/退出/清理、源码前后、锁定Playwright1.59.1与实际Chromium、原PNG和sidecar。若再失败，返回准确image/sidecar/render.before/after及失败边界，不重试图片或只输出截断长转储；完整白名单在sidecar可访问。若完整绿，仍须独立QA实际图/术语/新self_hosted分项判定。
- 动态原生独立QA、实际Owner preview已登录实操、真实供应商/支付/身份继续分别Not Run，不由本轮诊断或受控HTTP替代。requirements/design不适用：冻结输入未变，跳过理由仍沿用总览；发布由Connector负责。
