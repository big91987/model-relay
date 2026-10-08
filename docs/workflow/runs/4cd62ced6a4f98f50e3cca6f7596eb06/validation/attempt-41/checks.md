# 执行41：visual-boundary-r1（诊断候选，未修复宿主断言）

2026-10-08；development；同Run、同checkout、HEAD `98511771871cf0951ecef716bbb55deb13e18da5`。当前129文件指纹 `fb766c0aca5f33fbb0960e13286a13461d20999485e93a19a841eb51114f92a4`，见 [source-evidence.json](source-evidence.json)。T001 In Review / M01 Ready for Review；**No-Go、QA38-01 Open、当前完整验证 Not Run**。诊断完成不代表产品缺陷已修复。

## 恢复与证据归属

- 正式 `read_command_output(seq=40)` 按offset `0 → 32768 → 65536 → 98304 → 101576`读至eof；归档101576 bytes，truncated=false，`make verify` exit2。原日志仍以平台seq40归档为准，本文件不是完整日志副本。
- seq40源戳为129文件 `bb30f8179c08e5e9dad248382132ec1018959c934efcd5951b1cab55e5b486da`；[host-seq40-source.json](host-seq40-source.json)复制自attempt39同摘要source-after，不声称该副本是seq40命令新生成的文件。
- [host-seq40-continuous-1280.json](host-seq40-continuous-1280.json)保存实际宿主进程11460的连续失败回执，与归档末尾一致；[host-seq40-visual-1280.json](host-seq40-visual-1280.json)保存实际visual review，pages=[] / phase=overview / failed。尚无本次十页PNG；1440/390连续及新视觉采集未到达。不能跨用seq37绿或混入旧test-results图。
- seq40旧候选173 Node、Go race/build、原迁移/旧UI/多上游/console/fault门禁通过；1280连续J01–J07检查点全部通过，随后visual-overview失败。健康历史命令开头的exit1是原门禁要求的隔离旧版红重建，后续当前健康测试通过，不是本次新增故障。
- `last_status=404`与J06跨租户账单预期读一致；不是overview响应。`failure_boundary.available=false`且visual文件没有facts，**现有回执不能证明是哪条capture断言失败，更不能证明overview HTTP404**。

权威输入保持design-v0.2.0 / frozen-for-test-implementation，archive SHA256 `5f51d35fdc15ed7214749f9413879b4238097f1d6b6bbe2cc0d508d8e103cef7`。实际入口为 `.workflow-input/4cd62ced6a4f98f50e3cca6f7596eb06/5f51d35fdc15ed7214749f9413879b4238097f1d6b6bbe2cc0d508d8e103cef7/files/manifest.json`；当前读取manifest、docs/acceptance.md，沿用本Runarchitecture-r1、T001及QA38返研包。19文件长度/摘要和archive重新核对一致，实际完整引用路径/逐文件摘要见 [material-integrity.json](material-integrity.json)。未执行材料脚本、不修改冻结资料、不访问旧Run。Codex/gpt-6.1-sol沿平台指定配置；实际原生模型由平台核对，不在产品日志自证。

## 已证事实与待证假设

源码事实：`journeys.mjs`的J06正常通过测试钟推进下一月以验证结账；`workspace.js`有合法的“当前范围近7日暂无调用”呈现；`assertCaptureFacts`一律拒绝任何`.empty`。这些事实使空态假设合理，但**不能证明seq40真实DOM满足该状态或该断言就是唯一根因**。标题、服务端角色、scope、字段/动作、数据行和布局仍可能失败。

本轮最小差异是失败不可归因，不是未经证明的业务错误。所有者在`tests/browser/console-visual.mjs`：它原来在断言之后才写facts，失败丢失了边界。本轮只修这个可观察性缺口：

1. 导航前创建pending_capture，断言前保存安全DOM facts：原title/scope/role及字段/动作布尔值、行/card数、loading/error-node数、root/content实际尺寸、选中范围/选项数量、固定已知空态分类及其他空态计数（总数/可见数分开）。不存DOM文本/账号/Key/消息。
2. 当前页面监听实际overview GET响应，仅精确白名单资源/status/scope-match，最多16条；不保存地址、query、headers或body。J06旧404不能进入这个列表。退出必移除监听。
3. 保留实际失败phase、具体check及capture断言的固定消息；导航失败facts=null，不填造DOM。完整安全facts写既有review.json，失败摘要进入固定入口日志。
4. 保留原ready/行数/字段/角色/范围/宽度/秘密全部断言，不允许任何新增空态，不改生产UI、Go接口、身份、预算、账本、迁移、测试钟或夹具业务状态。未改变Makefile，也没有增加额外依赖。

## 实际局部验证

| 检查 | 实际结果 | 证据与边界 |
|---|---|---|
| 首次定向visual Contract | 4/4 exit0 | 原工具回执；模拟receipt写入，不是浏览器事实 |
| 首次 `make frontend-test` | 175/175 exit0 | frontend-check.log |
| 最终 `make frontend-test` | 176/176 exit0 | frontend-check-final.log；原173保留，新增3项白名单/断言前留证/导航失败脱敏Contract |
| JS语法、format-check、modules、vet、build、diff-check | 初次及最终exit0 | static-check.log / static-check-final.log；构建不等于真实HTTP/浏览器通过 |
| archive及19件材料摘要 | exit0 / 全一致 | material-integrity.json |
| 当前源戳 | exit0 / 129文件 | source-evidence.json / source-after.json |
| 原完整 `make verify` / 新63PNG / 独立QA | 当前候选Not Run | 正式宿主tests执行；本轮没有宣称已修复visual-overview |

原生监听限制没有相关环境变化，沿用validation/report.md及attempt33/36保留的首次命令、退出码、版本和范围；没有重复已知必败监听、提权或static browser.check。seq40宿主失败是待诊断测试/产品边界，**不归类为环境限制**。Contract测试的人工facts只证明记录和脱敏机制，不证明Chromium、API或实际空态。

Trellis现场已有初始化文件/脚本及源码规范，`get_context.py --mode packages`成功（single-repo backend/frontend），无需bootstrap空模板。按trellis-before-dev读取frontend索引及DOM/lifecycle/verification/workspace契约、reuse guide后编码；trellis-update-spec记录安全失败边界。trellis-check检查原断言无降低、监听退出清理、字段不泄秘密、现有JSON职责复用、176回归及构建；单测试层变动，无跨层接口/存储变更。交付Skill沿唯一T001维护公共任务和总览，未重建方案/任务。审查结论Blocked by Validation；关键路径为有界真实诊断→有证据最小修复（必要时）→同指纹原完整门禁→独立十页QA→Connector发布。

没有用户决定事项。实际供应商、外部身份/支付、H02/H03/H04/H07/H08平台事实及发布仍Not Run；不得把受控上游或局部测试写成商用已对齐。后续见 [host-retest.md](host-retest.md)。
