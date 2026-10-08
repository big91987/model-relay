# QA65 → report 交接

## 真实结论和唯一版本

产品候选budget-clock-r1 **Go with known issues**；不是完整Harness/发布最终Go。repo big91987/model-relay，Issue25，原PR27，分支workflow/4cd62ced6a4f98f50e3cca6f7596eb06。当前HEAD66f1328ada75ea90e5638fbf2bd58663ac86347e + 未提交140件源戳ca8ec50aa168cf16ac2e24edfe214aa44d06a05846c34deadbc3c7cb6c02904a。seq64新make verify exit0/110897bytes/truncated=false；QA65恢复到eof，独立重算一致，不能使用QA58/seq57替当前候选放行。

本轮关闭P1-1零软预算绕过、P1-2时钟回拨复活：准确PR27历史HTTP红灯重建+新当前ActualHTTP绿灯，QA另做六组非监听race24.750s/exit0及193Node/exit0。三视口J01–J07/十页/摘要/故障/正式CLI迁移通过是宿主实际UI证据，QA原生实操Blocked不改Pass。完整报告/矩阵/来源见同目录acceptance-report-65.md、acceptance-matrix-65.md及evidence/attempt-65/host-seq64-review.md、integrity.json、ui-artifact-index.json。

## report必须处理

1. 更新本Run根 `pr.md`，绑定本候选工作树源戳、tests64与QA65，以及用户“通过产品自身Pipeline提供最新实际页面”要求。publish/pr更新同一个PR27，不新建；Connector随后返回的新head才是含本修复的提交。现有PR27旧66f1328不等于已含修复。
2. 所有产品/测试/锁/哈希内契约修改必须新make verify与受影响QA；仅报告/脱敏证据提交也要重核140件源戳未变，绑定新提交而非用HEAD变化误称产品变化或忽略内容变化。
3. 保留QA58-01 P3非阻断；独立self_hosted首次网页分支Not Run（用户明确先本候选发布后补测）、原生QA实操Blocked、供应商/外部身份支付Not Run。H02、H03实际模型凭证、H04负向、H08与最终H07等仍未验，矩阵逐项不隐藏。
4. 用户已授权测试仓合并和正式Actions部署；协调者核验后经verified-main Deploy Model Relay locally既有流程执行。preview5545仅为配置目标；真实Deployment environment_url/admin/healthz必须绑定merge SHA/prepare/Actions/运行版本，不能用旧main/8793/临时服务/预写URL冒充最新页面。Agent不代发GitHub动作。
5. 给安全初始化/首次登录说明：沿实际CLI与docs/local-deployment.md/迁移运营合同，使用私密文件/安全分发，不在PR/日志写管理员密码、应用Key、主密钥或本机绝对路径。未在正式最新部署重跑的操作步骤标待实测/草稿，不能出“已验证部署手册”。

## 当前实际产物

- acceptance/acceptance-report-65.md：最终独立报告/缺陷表/质量趋势。
- acceptance/acceptance-matrix-65.md：AC01–18/H01–08分层及七个连续旅程契约/商用五层状态。
- acceptance/evidence/attempt-65/check-integrity.mjs、integrity.json、qa-inprocess.log、qa-node.log、host-seq64-review.md、ui-artifact-index.json、rejected-artifact-selection.json、retain-ui-evidence.mjs：独立脚本/结果、日志与证据索引；51件宿主UI文件的实际路径逐项见ui-artifact-index.json，不将目录作为交接artifact。
- acceptance/README.md及index.md：当前报告/矩阵索引；QA58/47等原报告和全部开发失败证据保留。

不需要需求或设计返工/用户再次审批；余下发布职责经report→publish→pr正常继续。所有外部未验事实由对应Connector/协调者闭合，不由产品测试数量替代。

执行68纠正：seq67因目录artifact失败，原回执保留。后续使用acceptance/handoff-correction-68.md及acceptance/artifact-files-68.json的非空文件级清单；QA65结论、矩阵和源戳不变。
