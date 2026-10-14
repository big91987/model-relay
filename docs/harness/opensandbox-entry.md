# Model Relay 的独立 OpenSandbox Harness 入口

本仓库复制并维护 `.harness/agent-platform` 内的固定SDK、模板及工具，Actions直接执行仓库内脚本。`SOURCE.json`记录初始源版本和文件摘要，项目调整由本仓库Git记录，升级需审阅差异。

原产品入口沿用AGENT_PLATFORM_WORKFLOW_CONFIG、Owner检查和Mac ARM64/he-full Runner，排除新标签。新入口选择标题【Harness验证】具体目标和harness-opensandbox标签，使用独立AGENT_PLATFORM_SANDBOX_CONFIG；手动入口也核当前Issue及标签。Owner评论续接原会话，机器人输出不作为用户输入。

Runner提交repository、固定repository_commit与原目标，要求expected_environment=opensandbox，不克隆Agent工作区。平台在所属Run中准备共享/workspace，各节点执行保持独立持久会话；同节点澄清接续保留会话，显式QA返工产生新step并保留旧历史。未知受理先按原request_id核对，不创建替代Run。

## 当前安装及使用

独立平台使用维护源3261b340a7d3b20b29f524739d346dd1b4025cda，标准升级保持原账号与运行历史。正式安装：[研发与独立QA流程](http://10.1.20.112:8795/workflows/83f08f81f7040fcd0c12e6ff4e2acd2d)。流程为准备→研发→固定make verify→独立QA→人工确认→报告→草稿PR，固定测试失败或QA返工返回研发。准备/发布节点选择私有Git凭据引用，Agent/tests不继承；报告、QA原文在发布前保留工作区，发布后通过GitHub产物查看。

Runner私有配置字段为base_url、workflow_id、repository、issue_label、base及token_file；这里使用big91987/model-relay、harness-opensandbox和main。仓库变量AGENT_PLATFORM_SANDBOX_CONFIG指向该私有文件。真实路径与值不写入仓库或Issue。标签与私有配置就绪并核对Runner能访问后，Owner创建唯一新Issue；原Issue重投找回原Run，完成Run不复活。

配置安装、任务受理、固定测试、独立QA、报告发布和合并必须分别核验。当前复制入口本身不修改产品，不代表新的仓库Pipeline已验证，更不代表产品商用验收。旧业务账号、数据、部署入口及产品Run不在本次范围。

## 维护

从明确源版本导出到新目录，比较SOURCE.json与已有项目调整，经审阅合入；不自动覆盖定制，不把平台源码路径塞回Actions。入口修改后更新.github/agent-platform*-entry-source.json实际摘要。GitHub通知使用正式App身份，审批与执行失败附处理入口；未知发送结果先查外部回执，不盲重发。配置缺失先修标准入口，不临时修改Run、凭据或权限。

入站支持私有配置的git_proxy，Issue/commit查询、Owner补充及通知使用同一GitHub子进程代理；平台API与全局环境保持。当前模板源版本以SOURCE.json为准，平台Go运行版本仍3261b34。入口失败日志提供安全阶段及HTTP/传输类别，不公开底层响应、命令或私有路径；先核原Run，再正式恢复原Issue。

固定commit准备同时保留完整可达祖先历史，历史回归不因浅克隆缺对象而失效。模板升级应用于新Run；旧冻结图不换脚本，需要取消/换环境时保留原证据并明确新旧任务的替代关系。源码未变的重复测试失败先定位，缺准备修复入口时blocked，不反复求绿。
