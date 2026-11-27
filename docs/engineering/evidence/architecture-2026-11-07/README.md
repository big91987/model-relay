# 架构工具整改证据

报告由本仓已安装的 archcheck 0.4.1 标准入口生成，文件内容未改写。所有模型判断都是待复核的候选意见，不代表产品运行通过。

| 目录 | 对应版本与用途 |
| --- | --- |
| before | main 42b79a0 的基线源码，5 项候选发现 |
| intermediate | 首次拆分后的中间源码；随后浏览器回归仍发现默认 fetch 读取时机改变，不能冒称最终验收 |
| after | 最终前端修复后的同版复扫，4 项候选发现；精确版本见 report.json 中 snapshot/hashes |
| cold-verify-state.json | 原始隔离冷安装状态回执；npm 下载超时，失败退出码保留 |

每个模型扫描目录含标准 report.md、report.json 和 model-result.json。SHA256.json 保存报告文件摘要。扫描时 git_head 尚为分支基线，不能单独用它识别工作区代码，应核对 snapshot 和逐文件 hashes。

[整改、回归范围与剩余问题](../../architecture-remediation-2026-11-07.md) 对模型意见作人工复核；它没有回写或改造模型原文。
