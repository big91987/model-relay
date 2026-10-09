# 执行27：页级输入观测补缺，空图仍未根治

## 基线与恢复
- HEAD仍为 `3f9cee1615e87e002a2c2300f2d29778072ec3f5`；恢复同Run总览/T001/M01、执行25和独立QA14整改，不重启Run或重建产品方案。
- 材料：design-v0.2.0；archive SHA256 `5f51d35fdc15ed7214749f9413879b4238097f1d6b6bbe2cc0d508d8e103cef7`。已读取 `.workflow-input/b7f93c24bc1ea4d0fe182f12aa7622ae/5f51d35fdc15ed7214749f9413879b4238097f1d6b6bbe2cc0d508d8e103cef7/files/manifest.json`，引用同root `docs/prd.md`、`docs/interactions.md`、`docs/acceptance.md` 的既有网页/秘密/同ID费用契约；不执行材料脚本，frozen-for-test-implementation不改为批准。
- 使用研发交付Skill及两份references维护现有公共事实；trellis-before-dev实际执行get_context识别单repo，读取frontend index及DOM/lifecycle/dependency/verification/workspace和共享reuse/cross-layer。现有初始化脚本与规范均有源码合同，不以prepare initialized=false宣称重建或缺失；bootstrap不适用。trellis-update-spec将本轮最小观察边界写入既有workspace-contracts。

## 执行26真实失败
- 原固定冷起点make verify：UTC `2026-10-09T04:00:50.769388Z` 至 `2026-10-09T04:11:53.712778Z`；Connector exit2、gate_exit2、runner exit1。原日志实测249 Node通过，原前置门禁和1280 J01–J07/summary通过，full-console在visual-keys失败；其余完整三视口visual与本次self_hosted未执行完成。
- 源前后154件戳均为 `bce88d0683991cef837feea8b321cc67a75a18da9b09a71760105b816f4c7824`。cold/cache/locked依赖和清理状态保留在原state；owned_group_absent=true、candidate_removed。不是本地监听限制，不能写成产品已根治或完整通过。
- 91件原字节产物逐件SHA核对归档，索引 `host-tests-26/archive-index.json`；原runner verify.log为477430bytes，SHA `424a6fcba37ed649d4c0394daa820d6c21fb9ad4ce2cc68d41b1d43db3493431`。Connector外包装477612bytes，不混为同一文件。已调用read_command_output(seq26)确认retained日志，未把首段工具输出当全部；归档的原runner完整文件是事实源。
- 实看原keys PNG：1280×1149全单色RGB243/246/250，四锚点零墨；SHA `78a7b304df77dac78964e27f6571f81f2ac48376e545786f9778a183028bff53` 与执行20/22/24相同。前后root scrollY148、末列scrollLeft32/clientWidth988/scrollWidth1020已达；几何和可见DOM不证明像素可用。实际4个LayerTree变化与10个paint事件也不证明栅格正确。
- 执行25修复已产出真实同钟trace：keyboard scrollend170ms；wheel-dispatch-start178ms；capture-start193ms/end282ms；未观测到表格wheel。此事实排除“失败后trace被覆盖”的旧证据缺陷，但表格冒泡观察不能区分投递到表外与没有观察到DOM wheel。没有wheel记录不证明浏览器没有处理输入；上述相关性不证明空图机制。白名单摘要：`development-27/host-failure-summary.json`。

## 最小修改边界
- 只改 `tests/browser/readability.mjs`、其测试及既有Trellis合同：将唯一wheel被动观察从表格冒泡移到ownerDocument捕获；键盘/本表scroll/scrollend继续在原表。新增有限坐标client_x/client_y与target_inside_region布尔，最多32事件/3标记，finish移除两个作用域监听并dispose。
- 已有输入序列、唯一100000 wheel、keyboard-scrollend、2000ms、末列精确bounds、原单次fullPage PNG及像素/秘密/几何断言全部保留。没有产品/API/DB/Go/锁定依赖/权限/控制器改变；没有重拍、sleep、raf、重试或推测性补丁。
- 新Node回归验证页级内/外目标、单次监听、白名单、上限和清理；既有失败后trace保留与original capture异常测试继续执行。这仅证明诊断合同，不能模拟证明Chromium输入/合成或宣称空图修复。

## 本地检查与trellis-check
- `development-27/local-checks.json`包含实际命令、UTC、退出、日志；沿用原Makefile工作区Go缓存/CGO/local toolchain，不手装依赖。
- 66相关Node通过；format-check/vet/build通过；普通/SSE非监听Go race合同通过；Node syntax及git diff --check通过。新增目标测试首次16/16通过；正式保存组合日志66/66通过，无本轮测试失败被隐藏。
- 源前后154件戳 `0211a610cb72b4435455a6d746d7648e9668d414a967d021c1b645ff97825e96`一致，见source.json/source-after.json。trellis-check核范围、复用、观测异常、秘密投影、cleanup与规范同步；没有新跨业务层数据流、类型系统或UI框架，不新增不适用检查。
- 最终证据索引首次因把无stdout的成功syntax-diff.log纳入非空检查而exit1；保留 `development-27/integrity-first-failure.json`。不伪造命令输出或删除原空log，其真实命令/UTC/exit在local-checks.json；交接不列空文件，最终重新核对见final-integrity.json。
- 稳定本地监听限制未改变，未重复运行必败完整入口；首次native失败证据继续保留。本地子集不是完整make verify绿；正式完整测试仍交已配置宿主Connector。

## 交tests与未决事项
- 保持原固定 `make verify` 冷起点、原npm ci与锁定Playwright1.59.1/真实Chromium，全部Node/Go race/HTTP/UTC预算迁移、三视口J01–J07/summary/visual及首次self_hosted。保存完整安全log/UTC/退出/清理/源码前后与原PNG/sidecar；新Node数量实测，不从本地推算宿主Pass。
- 若再失败，先核页级wheel的target_inside_region与client坐标/同钟阶段，对照原PNG与合成事实。观测到表外wheel可定位输入归属；观测到表内wheel仅证明该DOM边界；仍无记录继续保持未知，不以另一假设性产品修改或放弱断言求绿。此轮不是历史空图或R28-04根因修复。
- T001 In Review、M01 Needs Revision；QA14/QA58未Closed。完整绿仍须独立QA核当前真实安全截图、术语分项和self_hosted覆盖。动态原生独立QA、Owner preview已登录操作、真实供应商/支付/身份各自Not Run；受控HTTP不冒称真实供应商。
- requirements/design无新范围偏差故沿用冻结输入并跳过，不补造新文档。业务账号/配置/DB、发布/部署不操作，Git发布仍归Connector。
