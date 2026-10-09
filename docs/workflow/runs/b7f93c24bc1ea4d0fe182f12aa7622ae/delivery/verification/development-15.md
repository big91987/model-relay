# 执行15：QA14截图失配的有界原层诊断

## 输入、版本与裁剪
- Run b7f93c24bc1ea4d0fe182f12aa7622ae，development执行15；HEAD main基线 `3f9cee1615e87e002a2c2300f2d29778072ec3f5` 不变。本轮候选154件 SHA256 `4a5046dee6b66429a75652f7ad8c5344a60dff1448a10c7293f9423bad1cf10c`，见 `development-15/source.json`。前序152件 `65201f…` 宿主执行13绿只适用于前序版本，不冒称当前完整通过。
- 恢复同Run总览/T001/M01与 `../../acceptance/{acceptance-report-14,acceptance-matrix-14,visual-review-14,remediation-14}.md`；保留全部210件原宿主资产、181PNG及4全单色事实。没有覆盖QA14报告/原图、DB、检查点或既有正式账号。
- 恢复原材料 manifest：`.workflow-input/b7f93c24bc1ea4d0fe182f12aa7622ae/5f51d35fdc15ed7214749f9413879b4238097f1d6b6bbe2cc0d508d8e103cef7/files/manifest.json`。版本design-v0.2.0，仍Frozen for Test Implementation；archive SHA256 `5f51d35fdc15ed7214749f9413879b4238097f1d6b6bbe2cc0d508d8e103cef7`。实际引用同root的 `docs/prd.md`（e768b4f26b4471e714d92b0db6e3187c89f37a25a697c01a12c6079af07de547）、`docs/interactions.md`（95f3490ef003fa9b19dd6b9e864dd3b0429619b8cd9d0ecec853919d6a12c801）、`docs/acceptance.md`（2792740b1579d614780856a13b6257d192736e1efc0309f1dc41405a665bef7a），本轮重新hash与manifest相符。AC01/02/06/15/17要求真实三视口、用户写入/失败恢复、秘密不回显及同ID核对，不能由像素门禁或设计夹具代替。
- 冻结产品/架构没有冲突，requirements/design不重跑；不另建PRD/设计或第二套任务状态。使用现有实质Trellis规范，包发现single repo/backend/frontend；prepare initialized=false并不抹去当前已存在的scripts/spec，不声称重新初始化。bootstrap不适用。

## 已证事实与待证假设
1. QA14原图确有4全单色、另有主体缺失/偏移；本轮实际解码181PNG，尺寸及uniform结果全部匹配QA14独立Go解码器，见 `development-15/original-png-audit.json`。另实际打开QA14 `visual-1280/keys-page.png`（全背景）和 `visual-1280/usage-page.png`（侧栏可见、主体空缺），与QA观察一致；未声称本轮人工重看其余179张。这只证明PNG事实/解码器可读性，不证明live页面的位置或Chromium合成机制。
2. 既有DOM root_fits / content检查在截图之前；既有截图入口未对截图的实际主体像素或截图前后document坐标设门槛。这是已证验收工程缺口。原层图片缺陷不能根据仅绿色DOM关闭，也不能推断业务/API全量失败。
3. 真实页面移位、焦点造成祖先滚动、截图采集/合成失配等仍是待证假设。没有原图同时刻geometry，不选择其中任一作为根因；本轮不加入滚动复位、CSS替换、sleep、断言重试或图片加工。

## 实施边界与回归
- 共用 `tests/browser/capture-integrity.mjs` 接管visual十页/详情/编辑及continuous/self_hosted的既有安全截图，秘密/存储/对话原门槛在前。每图仅一次 `page.screenshot({fullPage:true})`，原PNG不改、不拼接、不替换；原 keyboard own scrollend → 唯一wheel →2000ms精确末列断言完整保留。
- 同一截图前后记录UTC、window/visualViewport/root/body/main/content/dialog/title、最多8表region数值位置/尺寸/滚动、focus类型布尔值。无DOM文本、账号/ID/URL、属性值、对话或raw异常。旁证路径固定为 `<原截图>.integrity.json`；现有Go runner失败时也复制整个test-results，故无需改Connector或额外导出权限。截图自身报错也尝试一次只读after诊断，保留原错误；不重试截图/断言。
- 最多18个实际投影锚点：页面或dialog标题、非同一标题的实质内容、各表当前可见表头/首行单元格。要求锚点落入图片、contrast≥32且dark pixels≥8，截图宽度与layout viewport一致，title/content的document几何前后差≤1px。不是仅判断整图非单色，侧栏-only、标题可见但表格主体缺墨及移出图像均失败。门禁是必要完整性检查，不声称证明每个动作/列可读，仍需独立QA看实际PNG。
- 7项新增Node测试涵盖正常/空白/侧栏、位移/缺锚点/缺数据像素、前后几何变化、白名单及真实DOM回调、PNG支持/截断、单次截图、原错误身份/失败sidecar及两原入口秘密门槛。纯夹具不证明native截图机制。
- API/DB/业务CSS与JS本轮零新改动；已有术语/首次自托管实现保留。价格/Token/角色/授权/安全数据口径未改，用户写入仍由页面动作。未添加依赖、没有发布、切分支、push或供应商调用。

## 本地检查与证据
最终当前版本命令、UTC及退出码以 `development-15/local-checks.json` / 对应日志为准。相关Node50/50、Go race非监听的普通/SSE夹具、format/vet/build和syntax/diff通过；不是完整HTTP/浏览器门禁。

保留失败：`development-15/node-all-missing-dependency.log` 记录本地一次 `node --test tests/browser/*.test.mjs` exit1，231 pass /1模块加载失败，缺 `playwright-core/package.json`（执行在新增第7项前）。与源码行为失败不同；未改测试、未跳过该文件冒充全量绿、未临时预装依赖，宿主原入口按lock冷安装。`development-15/node-fixture-first-failure.log` 记录新增geometry夹具首次exit1：夹具cell宽12小于实际门槛24，修正测试尺寸32后原门槛不变；最终50/50可定位，不删失败。

当前沙箱环境/监听权限未改变。原 `make verify` shell exit2 / runner exit1 / gate exit2，在2026-10-08T23:56:58.265233Z至23:57:25.256346Z受原review-history监听限制，原152件6ba7643…版本及清理/日志在 `native-verify-state.json`、`native-verify.log`。按稳定限制不重复必然退出的完整命令；将该首次受限事实与当前154件待验候选分别记录，不把旧限制当当前通过或当前产品失败。新完整make verify明确交正式tests。

## Trellis-check与交接条件
- 已按frontend index读取workspace/DOM/lifecycle/verification/dependency preparation及共享reuse/cross-layer指南；本轮只测试采集层，不改后端层。检查helper复用、原secret gates、无新增写操作、无assert重试/timeout扩大/调试敏感输出、无需TypeScript。更新workspace-contracts具体签名、失败矩阵、好/坏例与所需原层回归；规范非模板，不创建重复规范。
- T001 In Review仅表示诊断门禁可交宿主，不表示QA14修复Done；M01 Needs Revision。原执行13行为Pass保留为历史，新源码全量/截图实际恢复仍Not Run，QA58-01与QA14-01未Closed。R28-04原因未知；动态独立原生浏览器、Owner preview已登录实操、真实供应商/支付/身份分别Not Run。
- tests在当前源码戳上用原固定 `make verify` 冷起点完整执行，保留源前后、锁定Playwright1.59.1/实际Chromium、UTC、退出、完整日志与清理，以及所有旧Node/Go race/HTTP/UTC预算/迁移/J01–J07/summary/visual和self_hosted。新增7测试正常锁依赖总量预期234，但以真实回执为准。
- 若内容门禁失败，保留原PNG及 `.integrity.json` 返回development：若before/after锚点document几何移位，可进一步区分focus/祖先/viewport滚动；若几何稳定且像素缺失，说明捕获与DOM失配仍在，不据此直接断言合成原因。先据边界选择下一项最小修复，不能反复完整求绿。若全门禁绿，独立QA仍逐张复核三视口十页和新分支，与冻结设计/QA14缺失图对照后才作Go/No-Go。
