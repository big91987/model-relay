# 执行21 证据索引

## 维护者后续历史摘要独立审查
- `../../summary-assessment-qa-21.md`：是否阻止发布、P2依据/用户影响/未知边界及若再红的真实原层最小重现与同步契约。
- `summary-history-review.json`：原18完整log/hash/state及安全failed JSON、本次20三宽度summary、来源/结果层严格分离；旧根因未知，不回填state。
- `summary-source-review.json`：原严格断言/当前源码hash/具体片段/未修改产品刷新顺序；标注Static Review而非Chromium重现。
- `summary-focused-contract.log`：实际4项原摘要断言/诊断负例测试通过，仅Contract。
- `summary-review.mjs`：只读QA采集器，不是产品同步修复或固定入口替代。未重跑整套门禁，动态独立浏览器Blocked、供应商Not Run保持。

本目录全为当前Run产物。完整字节级归档148件列于 `archive-index.json`，不以文件夹代替具体产物。归档脚本 `collect.mjs` 与独立检查收尾 `finalize.mjs` 仅QA证据工具，不是供应链/固定入口替代。

## 正式宿主20
- `host-seq20/verify.log`：原样实际runner全文310718bytes，SHA2567c4d3952943721deae9188a6f52f860572037ac8ea39ba55d278d6bf2c7abb7b。
- `host-seq20/state.json`：原样冷前后、UTC、exit、source/candidate不变、工具版本、最终probe、cleanup及owned group状态。
- Connector回执310815bytes/exit0/truncated=false，平台游标顺序读取到EOF记录在archive-index。上述runner文件不是Connector全文，不手抄/补造前缀。
- `host-seq20/browser/continuous-1280-62042.json`、`continuous-1440-62062.json`、`continuous-390-62137.json`：原样连续七旅程。
- `host-seq20/browser/management-summary-1280.json`、`management-summary-1440.json`、`management-summary-390.json`：原摘要动作/最终事实。
- `host-seq20/browser/visual-1280/review.json`、`visual-1440/review.json`、`visual-390/review.json`：十页、范围、角色、ready、内容、截图及敏感信息检查。
- `host-seq20/browser/fault-1280.json`、`fault-1440.json`、`fault-390.json`：五种故障场景；没有顶层result，四个当前场景passed及一个历史expected_failure_reproduced逐项核对。
- QA查看样本：`host-seq20/browser/visual-1280/overview-page.png`、`visual-390/models-page.png`、`visual-1440/keys-editor.png`；其余截图作为自动化辅助资产不冒称全部人工审图。
- `collection-check.log`：实际归档完整性/材料19件一致性检查；首次collector错误期待fault顶层result已纠正，只改QA工具，原始收据未改。记录于independent-review.json，不是产品门禁失败。

## QA亲自执行
- `source-qa-21.json`：当前147件f8befd与宿主前后一致；源码文件具体hash在 `independent-review.json`。
- `independent-go.log`：runner/依赖实际独立race exit0；包含实际临时目录链接/失败输出/有界取消/真实npm配置和fixture回归。
- `independent-node.log`：197项Node通过、0skip；使用工作区已有锁定依赖，只是独立回归，不冒充冷安装证据。
- `independent-static.log`：format/module/vet/build/diff exit0。
- `qa-native/verify.log`：QA实际固定入口完整失败runner146593bytes、SHA2565057aaccbb758d93b356ee9d86afa0b5ae2af2192fe0f7f6283e6a0c6badc186；历史listener unavailable，browser未到达，结果Blocked。
- `qa-native/state.json`：该次原样退出/冷起点/source一致/清理闭合，不与宿主绿拼成QA亲自完整通过。
- `independent-review.json`：来源/退出/完整性/源码锁/本轮采集工具纠错及脱敏记录。独立日志无实际字节改动，原始/归档hash一致。

历史失败继续引用本Runverification目录中的frontend-before、development-13-red、tests-18-verify/state/continuous JSON及QA6报告；不覆盖旧文件，不用历史source放行新候选。供应商与正式部署新版本/admin没有本轮事实，保持Not Run。
