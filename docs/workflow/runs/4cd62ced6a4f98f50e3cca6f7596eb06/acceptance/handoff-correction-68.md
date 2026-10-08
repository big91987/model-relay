# QA68 文件级交接纠正

> 日期：2026-10-08，Asia/Shanghai；同Run执行68，模式：只读证据与交接契约复核。
> 产品验收继续以QA65报告及矩阵为准；report66继续有效，不重建产品验收或改写历史回执。
> 判定保持：budget-clock-r1 **Go with known issues**；发布失败待正式重试，未部署最新页面。

## 失败及本次责任

正式publish seq67 exit1，归档978 bytes/truncated=false。通过read_command_output(seq67, offset0)读至eof，错误明确为QA65交接包含 `acceptance/evidence/attempt-65/host` 目录，而发布契约只接受同Run文档根内非空实际文件。不是源码测试失败；不放宽路径权限、不改发布器、不改历史回执。

本次仅纠正交接产物路径、新增核对记录与文件清单。此前未提交业务修复原样保留，不将本轮文档修正声称为业务修复；未提交/推送/部署，未运行GitHub动作。此轮没有启动监听服务、再次执行make verify或扩权，未重跑QA65已有副作用。

## 版本及证据复核

- 仓库big91987/model-relay，分支workflow/4cd62ced6a4f98f50e3cca6f7596eb06；当前HEAD `66f1328ada75ea90e5638fbf2bd58663ac86347e` 加现有未提交修复，暂存为空。旧PR27 head不包含该修复。
- 逐项核QA65 `integrity.json` 的140件source_files，按路径/NUL/字节数/NUL/原字节重算聚合SHA256，预期为 `ca8ec50aa168cf16ac2e24edfe214aa44d06a05846c34deadbc3c7cb6c02904a`。实际结果由 [recheck-68.json](evidence/recheck-68.json)记录；原QA65证据不覆写。该范围不等于全仓全部文件，报告与留存证据不在产品源戳内。
- prepare同一manifest/材料根：`.workflow-input/4cd62ced6a4f98f50e3cca6f7596eb06/5f51d35fdc15ed7214749f9413879b4238097f1d6b6bbe2cc0d508d8e103cef7/files/manifest.json`。version design-v0.2.0，status frozen-for-test-implementation；19件按manifest摘要/字节数逐项重核，archive SHA256 `5f51d35fdc15ed7214749f9413879b4238097f1d6b6bbe2cc0d508d8e103cef7`。实际引用路径逐项保存在recheck-68.json；不执行包脚本、不改Frozen材料。
- QA65的 [ui-artifact-index.json](evidence/attempt-65/ui-artifact-index.json) 所列51件文件全部逐字节重核。1280/1440/390宿主连续旅程仍归seq64来源；本轮只验证留存摘要，不声称重新浏览器执行或新增截图。
- 正式seq64回执exit0、110897 bytes/truncated=false；本轮按offset0/32768/65536/98304再次读取，最后next_offset110897/eof=true。首段源戳、旧版HTTP红灯、当前零软预算/回拨ActualHTTP绿灯和末段三视口旅程结果与QA65审查一致。工具显示存在长度裁剪，不能把显示摘要当完整原日志副本；完整来源判断沿用 [QA65宿主审查](evidence/attempt-65/host-seq64-review.md) 与平台保留归档，不补造原日志。宿主执行不改称本轮原生QA执行。

## 不变的验收结论及未决项

本轮实际核对结果：140/140源码摘要及聚合源戳一致，19/19材料摘要/字节数一致，51/51留存UI摘要一致；HEAD未变、暂存0件。生成清单73件均为同Run根内非空实际文件，另有清单自身；`git diff --check` exit0。没有新的产品测试执行结果。

P1-1零软预算绕过、P1-2时钟回拨复活仍为Closed；当前没有新源码或新测试结果支持重新判定，沿用QA65原层红转绿证据。QA58-01 P3 Open非阻断。产品候选仍Go with known issues，完整Harness/发布退出仍未完成。

原生独立QA UI **Blocked**；独立self_hosted首次UI、真实供应商、外部身份/支付及未验平台H项 **Not Run**；H07最新正式部署/页面仍Not Run。本次没有使这些项目转Pass。质量相对QA65产品层持平，仅修复交接路径可靠性，不增加覆盖率。

## 可直接使用的文件级交接

[artifact-files-68.json](artifact-files-68.json) 是明确非空文件清单，含QA65全部留存证据（包括host目录内51件实际文件）、验收报告/矩阵、report66及公共说明。目录不是artifact；该清单自身作为独立非空文件交接，因避免自引用摘要不列自身。已有UI索引继续作为51件的来源/摘要入口，不新增重复图片。

report本轮须只整理此次纠正与当前PR正文引用，随后沿report→publish→pr正式重试更新**同一PR27/Issue25**，不得新建PR。后续handoff.artifacts只传同Run根内实际非空文件；需要传图片等证据时，逐项使用清单中的path，不传目录，不要求放宽递归发布权限。不得覆写或删除seq67失败事实。

仅报告/证据文件变化且140件源戳一致可复用tests64/QA65；publish须记录实际新head并再核源戳。产品/测试/锁/哈希内契约任何变化均需正式make verify和受影响QA。report66远端可访问性仍待推送后核对。

协调者继续按已授权测试仓merge与verified-main正式Actions部署闭合实际merge SHA、Actions/Deployment environment_url、/admin登录页、/healthz运行version/schema3；5545只是目标，不能用旧main、8793或临时服务冒充最新效果。安全首次初始化/迁移恢复步骤保持待正式安装实测，不在文档记录秘密。无需用户新审批。

## 环境与复验入口

仅Node读取现有文件及git只读信息，生成本Run证据；未操作数据库、上游或共享运行服务，无新测试服务需清理。运行 `node docs/workflow/runs/4cd62ced6a4f98f50e3cca6f7596eb06/acceptance/evidence/recheck-68.mjs` 可复算摘要/材料/51件UI及非空文件清单；这不是产品make verify，不替代原门禁。
