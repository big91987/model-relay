# 执行45：非空弹窗截图的几何失败诊断

## 恢复与输入
- HEAD仍为 `3f9cee1615e87e002a2c2300f2d29778072ec3f5`，执行44前后154件候选均为 `6be11e323f52fc3bf47f20397ec84786f725c9bccf401cacebb797cc82e1d11f`，与执行43交接一致。恢复公共总览、T001/M01、执行43证据；已有实现与历次失败保留，不重放不确定调用或发布。
- 材料design-v0.2.0，archive SHA256 `5f51d35fdc15ed7214749f9413879b4238097f1d6b6bbe2cc0d508d8e103cef7`；读取原 `.workflow-input/b7f93c24bc1ea4d0fe182f12aa7622ae/5f51d35fdc15ed7214749f9413879b4238097f1d6b6bbe2cc0d508d8e103cef7/files/manifest.json`，沿用同root的 `docs/prd.md`、`docs/interactions.md`、`docs/acceptance.md`。不改冻结输入或材料状态。
- 已读取挂载研发交付、before-dev/check/update-spec及公共交付/审查references，读取实际frontend索引、生命周期、验证、workspace合同、依赖准备和共享复用指南。已有Trellis源码规范不是待填模板，不重新初始化或复制规范。

## 执行44原证据
- 275件原字节产物保留于 `host-tests-44/archive-index.json`。完整runner日志613961字节，SHA256 `60d642573908e25980493aba6c3ea9ddbd8407b72aaea774114cb7743d8f16b7`；Connector输出614142字节是不同封装。state证明冷起点、源不变、owned清理与候选删除。UTC为2026-10-09T08:31:02.611438Z至08:42:39.263198Z；gate_exit2、runner exit1、Connector exit2，非Pass。
- 原262 Node和前置Go race/实际HTTP/迁移/预算UTC回归通过。relay race包284.183秒，比执行42的320.136秒少35.953秒；这只是两次实测wall time差，不证明所有收益来自调度或保证下一次时间。
- 三视口J01–J07、summary/current-period通过；1280/1440完整10页visual与12步self_hosted通过。390在visual-keys-editor失败，后续visual/self_hosted未完成。错误发生于总预算到期前，非执行42的SIGTERM，亦非监听限制；不能拼接局部绿放行。

## 原图最小归因
- 已实际查看 `host-tests-44/browser/visual-390/keys-editor.png`：390x1227，有编辑弹窗与完整表单，非空图；SHA256 `0e37bd86cb1fad8f63ec660deb038d47de0c4a9fe79bbb66617f09de08be688e`。原侧车failure_boundary为integrity，before锚点像素均通过，随后document geometry断言失败。
- scrollY182→327，title.top273.75→418.75、content.top254.75→399.75，三者同移145px。root/body/main位置和尺寸不变；title/content视口相对位置与尺寸不变。原生命周期在截图call/return之间记录resize/scroll。
- `development-45/original-png-audit.json`用同一原PNG重算after锚点：title/content暗像素961/211、contrast255；before为69/7040。before的“有墨”不能独自证明采样在目标文本处。原PNG、原侧车与原失败不改写。
- 已证：本次是几何门禁失败，非全空/超时。待证：滚动的原生触发机制、实际computed弹窗定位及其坐标契约；历史空图唯一根因、R28-04根治均未证明。不据坐标相关性猜改CSS或免除弹窗几何断言。

## 最小变更与原门槛
- 本轮仅改 `tests/browser/capture-integrity.mjs`、其Node回归、`console-visual.mjs`与frontend既有规范。产品Go/JS/CSS、API/存储/费用/Token/权限无新增变化；既有13项并发父测试保持。
- 原前后边界增加固定五字段document/viewport有限数值delta、scroll delta、稳定布尔和allowlisted computed dialog_position。只有原截图失败时才对同一PNG补after锚点诊断；visual失败回执携带有界摘要。成功路径不重复PNG解码。
- 原 `assertCaptureIntegrity` 全部正文保持，视口稳定但文档移动的弹窗仍失败。after像素不是替代门禁，没有坐标归一化、截图方法替换、第二张图、新浏览器读取、sleep/重试、放宽等待/预算、依赖或权限变化。负例证明缺失/非有限数据不能稳定、真实位移及滚动相关弹窗仍失败、原错误身份和单次fullPage/两次原几何读取保留。
- 这是诊断能力，不是产品根治或完整交付；不把Node夹具当Chromium行为证明。下一次按实际computed定位、原图两组采样和同次生命周期选择有证据的修复，不循环猜测求绿。

## Trellis检查与本地结果
- 110相关Node通过；JS语法、git diff --check、原format/vet/build/go mod verify、verification wiring race通过。四项归档检查含UTC/exit/logSHA，见 `development-45/local-checks.json`。先前34项试运行通过，正式110项覆盖全部；局部绿不替完整入口。
- 已核源码边界、原层断言、复用、数据流与同层一致性；无新框架、依赖、生产日志或数据操作。规范仅捕获诊断签名、错误矩阵与负例，不改变产品设计。
- 当前154件候选 `88794346a0c678b418671a625365a555a28bd0626c2533492c692428cfc9a0a8`，Go1.25.6、HEAD同基线，见 `development-45/source-final.json`。历史原生监听限制环境未变，不重复同一必败原生完整命令，明确完整 `make verify` 留tests宿主，绝不写成本地Pass。

## 交接
T001 In Review、M01 Needs Revision、QA14 No-Go保持。交tests执行原固定冷 `make verify`：十二分钟总门禁/九百秒外层、parallel4、Playwright1.59.1/Chromium147锁和权限保持。不手装依赖、不改Connector或通用控制器。全部原Node/Go race/实际HTTP/预算UTC迁移/取消权限恢复、三视口J01–J07/summary/visual、首次self_hosted普通与SSE同ID日志费用/不可达及核验失败纠正恢复都必须完整执行。

再次失败返回development，读取新的 `capture_integrity` 摘要及原侧车dialog_position，按真实边界修复，不以这轮诊断宣称已解决。完整绿后才独立QA核实际安全图片、QA58-01与新self_hosted覆盖。真实供应商/支付/身份、独立动态原生浏览器与Owner preview已登录实操分别仍Not Run。受控HTTP不是供应商事实。本节点没有分支、提交、推送、部署或账号/DB副作用。
