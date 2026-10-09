# 开发检查与宿主交接

## 对象、结论与版本
- 本 Run T001：In Review；M01：Ready for Review（待 tests 与独立 QA，不是 Accepted）。
- 基线 HEAD：`3f9cee1615e87e002a2c2300f2d29778072ec3f5`；没有自行提交/切分支/推送。
- 当前候选源码戳：见 `source-exec12-final.json`；执行11已验失败候选为 `4d2649461c3548ebbbb9a1efe328467b6d0f83efb513aeddad6bbec5fac4be47`，历次回执均原样保留。本节点不修改冻结设计或发布对象。
- 结论：执行11在同一最右列门槛失败，事件证明wheel投递时键盘滚动尚未结束；执行12仅修复驱动的阶段同步，不宣称Chromium内部根因或产品修复已实测通过。完整门禁、self_hosted与独立QA仍未通过。原生监听限制不能解释此次宿主失败。

## 执行12：执行11真实事件归因与最小驱动修复
- 恢复同Run总览、T001/M01、材料manifest及实际PRD/交互/AC相关条目；version design-v0.2.0与archive SHA256仍为prepare值，状态仍frozen-for-test-implementation。已有实质Trellis规范与scripts，实际包发现single-repo/backend/frontend；使用before-dev与check，非本次重新初始化，不需要bootstrap模板。
- 宿主 `make verify` 原收据 `test-results/verification/20261009T010049Z-2250598582`：UTC `2026-10-09T01:00:49.281705Z` → `01:12:05.913017Z`，gate_exit2/runner exit1/Connector exit2，source/candidate unchanged=true，冷起点=true，NODE_PATH unset，无手装依赖，owned_group_absent=true。完整67文件按原字节复制至 `host-tests-11/` 并逐一核SHA，索引 `host-tests-11/archive-index.json`。日志372342字节，SHA256 `8c7de82f0f833840dd20b604ae3cf1dfc85211462e249f6df236f8d0002b9ef4`。保留所有失败、测试内负向runner回执、截图及清理，不从平台截取摘要推断通过。
- 该宿主前段Node225/225、Go race/HTTP及预算UTC/迁移与migration/legacy/routing/console通过；full-console1280已走J01–J07/summary，再在visual-overview/rightmost-column超时。完整1440/390 visual、self_hosted三视口及独立QA未运行，不能把前段绿外推为整条绿。
- 原始 `browser/visual-1280/review.json` 与 `browser/continuous-1280-28625.json` 的文件名以索引为准：正确region/owner、非technical或nested hit；ArrowRight keydown1ms/keyup2ms，keyboard scroll21/29ms；wheel47ms、deltaX100000、trusted=true、default_prevented=false、scrollLeft7；同刻scrollend7后继续scroll55–146ms至40，最终scrollend163ms。events完整21条、omitted0。确证：原驱动只等scrollLeft>0即切wheel，输入阶段重叠；微任务观察未取消不等于所有后续监听均未取消。未确证：Chromium内部wheel/键盘动画为何相互影响；产品不可达、供应商行为或其他未跑视口。
- 改动边界：只改 `tests/browser/readability.mjs` 及其回归、Trellis合同和公共交付记录。`startKeyboardScrollCompletion(region)` 在ArrowRight前注册被动own-region scroll/scrollend，要求实际移动后结束（2000ms有界）；等待该promise后才执行原唯一wheel，再保留原2000ms最右列等待及精确bounds断言。缺结束事件直接失败 `native-keyboard-scrollend`；取消/超时/成功均清理，不fallback。原指针、事件白名单、截图秘密门槛/列/布局/API/DB/角色/预算不改，不sleep/轮换猜测/断言重试/延时放行。 键盘结束缺失时保留同一被动白名单 `keyboard_events`，不丢弃失败时序。
- 本地回归：`node-exec12-final-targeted.log` 14/14与 `node-exec12-final-related.log` 七文件93/93、0skip、exit0；包含own-target过滤、必须先移动、2000ms失败、取消清理、结束前不发wheel、结束缺失不得发wheel、原末列异常identity不变。夹具仅证明驱动契约，不证明浏览器机制。`static-exec12-final.log` format/vet/build/syntax/diff均exit0。
- trellis-check及交付review：变更限原T001验证范围；现有test helper复用，无新依赖/产品框架/跨层写入。合同写入 `.trellis/spec/frontend/workspace-contracts.md`；冻结输入、术语盘点及原全部门禁不变。公共状态仍T001 In Review/M01 Ready for Review，Ready with Non-blocking Gaps，不Accepted。
- 原生沙箱环境未变化，沿用首次监听受限收据，不再运行必然同处失败的完整入口；本地缺playwright-core的原全量Node失败仍保留，不手装依赖。新的当前候选交tests在固定 `make verify` 的隔离冷起点准备锁定依赖/Chromium，完整回执后独立QA。若仍失败，先比较新keyboard阶段scrollend与wheel事件，区分同步缺失/产品/驱动/环境，不移除原末列断言。
- self_hosted仍Implemented / Host Not Run；QA58-01仍Awaiting QA；动态原生独立浏览器、真实供应商/支付/身份、实际Owner preview登录分别Not Run。R28-04旧精确历史原因未知。无requirements/design偏差，不生成重复PRD或改变已完成旧Run。

## 执行10：执行9实测边界归档与时序诊断
- 原固定宿主命令 `make verify`，UTC `2026-10-09T00:42:38.247388Z` → `00:53:32.689746Z`；gate_exit2、runner exit1、Connector exit2。source/candidate前后同执行8候选且unchanged=true、冷起点=true、无手装依赖、NODE_PATH unset、owned_group_absent=true。原67文件（65浏览器产物含51图）字节复制并校验SHA，`host-tests-09/archive-index.json`；完整日志/state与锁定版本保持，不冒充本轮验证或独立QA图片审查。
- 前段Node221/221、Go race/HTTP/预算UTC迁移及migration/legacy/routing/console通过；1280的J01–J07/management-summary/current-period事实通过。完整full-console在visual-overview/rightmost-column失败；1440/390完整visual、新self_hosted三视口、独立QA仍未执行。
- 最小事实：before/after的point_in_viewport、point_inside_region、hit_inside_region、hit_scroll_owner_region均true，hit_technical_value与hit_nested_scroll_owner均false；scrollLeft1→40，client399/scroll650，末列表头right1476.265625→1437.265625，表格right1227。没有根溢出/加载失败或权限错误。输入定位和本表横滚存在已经证明，但wheel是否被投递/被取消/与尚未结束的ArrowRight滚动交叠仍待证，不能把scrollLeft40解释为浏览器根因。
- 编码边界：只改原 `tests/browser/readability.mjs` 失败处被动观察、对应单元及源码规范；不改产品/CSS/API/DB/保存ownership/readback，不换输入、不扩超时、不新增sleep/重试。`startScrollEvents`在原ArrowRight前监听，原wheel前标记phase，失败后收有限事件；保留原异常。32条白名单相对时间、scrollLeft、trusted/default_prevented和wheel数字，无文本/URL/ID/key/raw事件或敏感值；超量计数、不可用明确标注，结束移除监听与dispose。微任务中的prevented只证明观察时刻状态，不代表所有后续监听器。
- trellis-before-dev：现有初始化和源码规范已恢复，`get_context.py --mode packages`返回单仓frontend/backend；实际读取frontend索引/DOM/lifecycle/verification及共享思考指南，沿用冻结材料manifest及本Run任务，不建立第二套事实源。trellis-update-spec将新观察契约写入现有workspace-contracts；不需bootstrap或后台规格变更。
- trellis-check本地实测：12/12可读性/新诊断子集exit0（`node-exec10-targeted.log`），七文件相关91/91、0skip、exit0（`node-exec10-related.log`）；format/vet/build、两文件syntax、diff均exit0（`static-exec10.log`）。直接全量Node命令实际224项中223通过、1失败、exit1：`route-cleanup.test.mjs`在模块加载时缺 `playwright-core/package.json`，并非本轮诊断断言失败；完整记录 `node-exec10.log` 与 `node-exec10-redaction.json`。未手装依赖，原固定入口仍以锁定 `npm ci` 在隔离冷起点准备；此本地失败不算完整通过，宿主必须完整执行而非使用本地223结果放行。原日志留忽略目录，公共派生仅脱敏workspace绝对前缀，原/脱敏SHA分别保留。
- 源码审查：新增观察不赋值scrollLeft、不调dispatchEvent、不preventDefault；只在原region被动监听，32条白名单及omitted计数，原输入次数/异常身份/精确末列断言的负例保持。产品没有新增wheel/scrollend处理或scroll-behavior变更。没有后台/接口/凭据/权限/依赖锁/固定门禁变更，故本轮跳过跨层接口改造；冻结材料archive及实际引用PRD/interaction/AC SHA与manifest复核一致。VM只验证诊断管线，不证明浏览器机制。已知原生监听限制未发生相关环境改变，按既有首败回执不重复必败全入口；完整race/HTTP/browser留给tests宿主。
- 源码戳生成：首次直接 `go run ./tests/evidence` 因默认用户Go缓存不在可写根而报operation not permitted，未显式采集该子命令退出码，不将其写成通过。改用原Makefile既有 `GOCACHE` workspace缓存、`GOTOOLCHAIN=local`、`CGO_ENABLED=1` 后exit0，记录 `source-exec10-check.log` / `source-exec10.json`；无权限扩大、依赖安装或门禁变更。
- 下一步：tests对新候选运行原完整 `make verify`。若同边界仍失败，先读 `readability.scroll_probe.events` 的trusted wheel/delta与keyboard/wheel/scroll/scrollend时间顺序，再做最小机制修复；trace缺失或截断明确保留，不将事件夹具作为Chromium证明。完整通过后独立QA；真实供应商/支付/身份/实际preview登录与动态原生独立浏览器各保留NotRun。

## 执行8：执行7新边界归档与最小诊断
- 正式宿主入口仍是 `make verify`，UTC `2026-10-09T00:24:46.574986Z` → `00:35:44.923733Z`；gate_exit2、runner exit1、Connector exit2；source/candidate before/after均为执行6候选且 unchanged=true，cold start=true，NODE_PATH unset、无手装依赖，owned_group_absent=true。完整日志与65份浏览器产物（含51张图片）归档 `host-tests-07/`，原文件与SHA均保留，索引见 `host-tests-07/archive-index.json`。这不是本轮新运行或独立图片审查。
- 日志344655字节，SHA256 `cc06358ed47459030954e6c4bc27d5992303048e8430ededa649a5743b00c83e`；state SHA256 `b49a2cea0fae70f4dc1f5887fb0c65c0030930c1ab4e09288d29a3ac5df94ca5`。Connector外层344837字节含包装输出，不伪造为文件长度或截断。
- 已证：Node221/221、各Go race包、迁移/legacy-ui/routing/console阶段通过；full-console1280形成J01/J02/J03/J04/J05/J07/J06、visual-current-period、management-summary通过回执。原首次exact row-count边界未重现，不据此宣称R28-04历史原因已解决。
- 新失败：`host-tests-07/browser/visual-1280/review.json` 在overview readability/rightmost-column超时；HTTP200、无loading或错误节点、root fits以及全部布局布尔值true，2个表格中1个横向溢出。源码控制流显示该表格已通过ArrowRight的scrollLeft>0谓词，计数0是计数在整张表格完成之后才递增，不证明键盘输入失败。回执没有wheel坐标/命中/前后几何，不能证明是产品不可滚动、驱动命中错误、嵌套滚动消耗或事件时序问题。
- 编码前边界：只改 `tests/browser/readability.mjs` 的原失败处诊断、对应单元回归及已有规范；不改产品/CSS/API/DB/保存ownership/readback，不换输入方法，不移动指针到假设性新位置。新增 `scrollBoundaryFacts(region,pointer)` 在原wheel前及失败后测region/viewport/最后列尺寸、scrollLeft和实际point命中/最近横滚owner，只保留有限数字和固定布尔值；错误保持原对象，诊断失败写available=false。没有页面文本/地址/属性/ID/HTML/原始error/凭据。
- 原ArrowRight、一次wheel(100000,0)、原2000ms完成谓词、最右列完整边界断言与安全截图门禁均不变；没有sleep、断言重试、延长时限或程序化滚动替代。Node夹具仅证明记录白名单/异常身份/输入次数，不能证明真实负坐标、命中或Chromium机制。
- 本地实测：10/10诊断与可读性子集（`node-exec8-targeted.log`）；七文件相关89/89、0skip、exit0（`node-exec8.log`）；format/vet/build、两文件语法与diff检查exit0（`static-exec8.log`）。执行8没有新的Go产品变化。完整Node依赖/race/HTTP/browser仍由宿主固定入口执行，不以子集代完整绿。
- trellis-before-dev已恢复现有single-repo实质规范和冻结material manifest/version/sha及实际AC/交互引用；没有缺失初始化或待填模板，无需bootstrap。trellis-check审查诊断透传到visual和continuous失败回执、原异常不被替换、无业务数据写入/权限变化；更新workspace规范，不改设计基线。公共交付审查结论为Blocked by Validation，指完整验收未闭合，非需要用户输入或外部权限扩大的阻塞。
- 下一项：tests对 `source-exec8.json` 执行完整原入口；若相同失败，先读取 `readability.scroll_probe` / `readability_facts.scroll_probe` 的before/after，区分point越界/命中非本表/嵌套横滚与本表真实变化，再选择具体修复。诊断仍不可用时报告具体available=false，不猜根因。完整门禁通过后才交独立QA；本轮尚未新增原生通过事实。
- self_hosted三视口仍未开始；full-console1440/390、完整visual及右端截图未运行。QA58-01、动态原生独立浏览器、真实供应商/支付/身份和Owner preview已登录操作继续分别待验。51张归档图不是新QA58-01完整图片集。

## 执行6：宿主失败归因与最小修复
- 完整执行5宿主日志与清理状态归档：`host-tests-05/verify.log`、`host-tests-05/state.json`；实际失败摘要 `host-tests-05/continuous-1280.json`。未删改原回执；原受控浏览器产物仍位于 `test-results/verification/20261009T000843Z-3220031292/browser/`。日志340072字节，SHA256 `0dc2aa03b1261f82b48b38e630928195f8d2dc98e4d4ff650b50d8b9bc8c381e`；state SHA256 `b7ca24a840c357c2ce1da1af8abca0b9f86146c8452beb292c0c32501a68a6ba`。Connector外层输出340254字节，包含包装输出，与文件字节不同，不推断截断。
- 已证事实：UTC `2026-10-09T00:08:43.023306Z` → `00:18:58.967904Z`；宿主gate_exit2 / runner exit1 / Connector exit2，旧候选 `d78781505e2ca5bfd3fe5d357ea3379cb9a414d7cd127ea0248c2491bc633180` before/after相同，冷目录与真实清理已记录。Node220项、Go race各包及迁移/legacy/routing/console阶段通过；随后 full-console1280 的 `model-save-readback-disable` 在第一次启停前失败，summary row-count=0。读取HTTP200、workspace可见、无loading/弹层/秘密；未形成任何J01–J07 checkpoint，新self_hosted、三视口完整summary/visual未运行。不能把早期局部绿作为本次完整通过。
- 最小源码对照：生产 `technicalValue` 在完整标识文本外创建命名region，却把所有region的 `aria-label` 设置为通用滚动提示；原 `model-save-readback.mjs`、`console-summary.mjs`、`console-visual.mjs` 均保留原按完整公共值匹配的 exact cell 选择器。区域名称替换身份是明确的无障碍契约缺口，也是此次零精确行的候选原因；宿主没有记录完整无障碍树，因此原生归因仍须修复后实测，不声称已根治。
- 修复只在产品展示helper中将 `aria-label` 等于完整公开textContent，把滚动提示放 `aria-description`；空值/未知/键盘滚动/局部布局/秘密边界不变。没有改行选择器、保存ownership、summary状态/负例、Go后台或固定入口，没有sleep或断言重试。新增单元回归覆盖相似前缀模型、地址、安全标识、字面HTML和null，断言公开值与区域名称相同。
- 执行6实际检查：50项定位子集通过（`node-exec6-targeted.log`）；原七文件相关子集87/87、0 skip、exit0（`node-exec6.log`，UTC与退出码在日志）；`make format-check vet build` exit0（`static-exec6.log`）；workspace.js/readability.test.mjs语法及 `git diff --check` exit0。均非native浏览器证据。
- 源码戳首次单独调用 `go run ./tests/evidence` 未设置项目GOCACHE，因默认用户缓存写权限拒绝失败；未生成有效回执。随后以项目 `.cache/go-build`、GOTOOLCHAIN=local、CGO_ENABLED=1成功生成 `source-exec6.json`；不记录本机绝对缓存路径，也不把该失败算产品失败。
- trellis-before-dev：恢复已有实质规范与single-repo包发现，读取workspace/DOM/lifecycle/verification/共享指南；不重建冻结设计或模板。trellis-check：本轮三文件代码/规范变更限于原展示边界与单元回归，原跨层写入与业务门禁不变；更新workspace规范的命名合同。公共任务审查维持单一T001，M01尚未Accepted。
- 下一步：tests对 `source-exec6.json` 候选执行原固定 `make verify`；原生沙箱稳定监听限制未变，不重复同一必败命令。先核此次 exact row / 原保存屏障是否通过，再依全部既有门禁和新增self_hosted证据交独立QA。若仍是同一row-count失败，须采集匹配模型单元格DOM存在数与实际accessible-name边界的最小脱敏诊断，区分未渲染、身份命名和定位器，不轮换假设性修复。R28-04历史原因未知；QA58-01、动态独立浏览器、真实供应商/支付/身份及已登录Owner preview仍分别待验。

## 实际命令与证据
| 实际检查 | 结果 / 范围 | 证据 |
|---|---|---|
| `make verify` | shell exit2；内部 runner exit1 / gate exit2。UTC 2026-10-08T23:56:58.265233Z → 23:57:25.256346Z；原 review-history 明确 historical listener unavailable, not business red | `native-verify-state.json`、完整 `native-verify.log` |
| `make format-check vet build` | exit0；Go格式、所有包 go vet、产品 build，不是完整race/HTTP/browser | `static.log` |
| `node --test tests/browser/readability.test.mjs tests/browser/workspace.test.mjs tests/browser/journeys.test.mjs tests/browser/console-summary.test.mjs tests/browser/console-visual.test.mjs tests/browser/save-lifecycle.test.mjs tests/browser/model-save-readback.test.mjs` | exit0，86 tests/86 pass/0 skip；这是相关子集，未冒称全部锁定依赖 Node 门禁 | `node-current.log` |
| 同子集首次执行 | exit1，6个详情VM因新增实际 enumLabel 依赖未注入而失败；已注入实际生产 helper，保留原断言并复验 | `node-first-failure.log` |
| `go test -race ./tests/browser -run '^TestContinuousUpstream(Response\|Stream)Contract$' -count=1 -v`（项目本地GOCACHE、GOTOOLCHAIN=local、CGO_ENABLED=1） | exit0；known/unknown与实际flush的SSE夹具契约。未证明 TCP、浏览器或供应商机制 | `go-fixture.log` |
| `node --check` workspace.js / journeys.mjs / readability.mjs；`git diff --check` | exit0；语法与diff检查 | 本记录；不代替动态测试 |
| 当前材料archive及 PRD/交互/AC SHA256 | 全部与 prepare/manifest一致 | 下表 |

材料实际 root 为 `.workflow-input/b7f93c24bc1ea4d0fe182f12aa7622ae/5f51d35fdc15ed7214749f9413879b4238097f1d6b6bbe2cc0d508d8e103cef7/files`，version design-v0.2.0、状态 frozen-for-test-implementation。

| 文件（相对于 root） | 实测 SHA256 |
|---|---|
| docs/prd.md | e768b4f26b4471e714d92b0db6e3187c89f37a25a697c01a12c6079af07de547 |
| docs/interactions.md | 95f3490ef003fa9b19dd6b9e864dd3b0429619b8cd9d0ecec853919d6a12c801 |
| docs/acceptance.md | 2792740b1579d614780856a13b6257d192736e1efc0309f1dc41405a665bef7a |

## 环境限制与代码归因
中间85项成功结果与当时源码戳 `80562e54d8c264ffede6e6c25ada7a41da8941da0cfd01b6f95e6142780f06c6` 保留在 `node-final.log` / `source.json`，未覆盖旧记录。当前86项与最新源码分别见 `node-current.log` / `source-final.json`。
首次完整命令候选源码戳 `6ba7643b87596d54fa93558f4bb8581f27e1bd1c1f231fbe08bc23dd397940ad`；before/after相同，source_unchanged与candidate_unchanged均true。锁定工具前置已实际测 Go1.25.6、Node26.0.0、npm11.12.1，冷目录node_modules/npm cache均false；未手装依赖。失败发生在 npm准备/完整race/HTTP/browser 之前。
完整日志156959字节，SHA256 `cfc8c1149f39c3ef4df254ea1738ace98b7dd44231ff4e76f0c6444f011168fd`；原生 private candidate已移除，owned_group_absent=true。`native-verify-state.json`保留真实清理结果。
首次Node依赖问题已在完整命令前修复。完整命令后补强SSE夹具回归、角色行唯一定位、活动dialog滚动归属、有界失败诊断/完整秘密排除，并让空原因代码显示缺省破折号而非空滚动标识；均属于候选代码/测试调整，不能归入环境限制，也不声称已获宿主通过。当前源码戳重新生成见source-final.json。稳定沙箱监听限制及执行环境未变，因此不再重复同一必失败 `make verify`；完整验证明确交tests宿主，不删除历史门禁或更换更弱入口。

## Trellis 检查与跨层审查
- before-dev 实际执行包发现：single-repo，backend/frontend；已有Trellis初始化文件和实质规范，不需要bootstrap模板。读取frontend DOM/lifecycle/workspace/verification/dependency、backend runtime及共享thinking指南。
- trellis-check：改动列表与任务边界一致，新增展示/原生回归约定已写回 `.trellis/spec/frontend/workspace-contracts.md`；DOM全部textContent/native标签，未知枚举 own-property防原型键误认，无协议/API/DB/金额/Token变化。
- UI用户动作→既有受CSRF/角色约束API→既有持久化；新测试只GET核对，不SQL/API写入预填。受控TCP关闭/401来源只是隔离测试端点，不成为默认产品上游。
- R30-SYNC-01保存operation/readback/epoch严格合同不变，原summary negatives、J01–J07、撤销/预算/取消/错误恢复、UTC/迁移/race门禁不删、不重试断言。
- 新 native 滚动需要宿主证明：活动dialog而非inert背景；ArrowRight等待真实scrollLeft变化，再只断言一次；wheel触达最右列且额外安全截图。只读标识不ellipsis，原有秘密清理/权限不扩大。
- 公共交付审查：单一任务事实源、单个小型端到端里程碑，依赖与未验边界明确；不是重新搭建组织级计划。结论 Ready with Non-blocking Gaps，验证结果由宿主和QA决定。

## tests 与独立 QA 必须完成
1. 对当前源码戳执行原固定 `make verify`，保留完整UTC/源码一致性/退出/清理/冷起点/锁定Playwright1.59.1与实际Chromium回执。不要手装依赖或扩大Connector权限。
2. 全原 Node/Go race/HTTP/预算UTC迁移/三视口J01–J07/management-summary/visual；保留原63图和新增 `visual-{width}/*-right-columns-*.png` 的安全检查与摘要。
3. 新 self_hosted 在每个1280/1440/390独立临时新安装：先零来源/模型/租户，网页创建已有自托管→真实连接关闭失败→错误凭据401→网页纠正核验通过→模型关联/价格/租户授权→新角色会话/Key→普通/SSE→同ID详情/日志/价格快照/费用与用量核对。输出 `self-hosted-{width}-{pid}.json` 与其中截图路径；失败有bounded check/readability/steps诊断，不输出业务秘密。
4. 独立 QA 依据冻结输入、历史 QA58-01、当前真实截图与回执逐页判断，不把新增子集或入口health当全面Pass。失败先查看具体边界/HTTP状态/步骤及native截图，区分产品、驱动、时序、环境，再选择下一项修复。
5. self_hosted现为 Implemented / Host Not Run；QA58-01为Implemented / Awaiting QA；动态原生独立浏览器、真实供应商/支付/身份、Owner preview已登录实操分别 Not Run。受控HTTP绝不等于真实供应商联调。R28-04旧精确历史断言原因仍未知。
