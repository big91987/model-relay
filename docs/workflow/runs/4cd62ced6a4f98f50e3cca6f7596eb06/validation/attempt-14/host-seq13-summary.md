# 宿主 seq13 失败边界

- 本Run tests执行13，固定命令`make verify`，exit2；86640 bytes、truncated=false。通过read_command_output(seq=13,offset=80000)恢复至86640/eof=true；不是凭截断预览推断通过。
- HEAD98511771871cf0951ecef716bbb55deb13e18da5；119文件工作树SHA256312cae2097f665d7e50ab6ac11290dd54fe444c6c6a37ed5dda233108d8bc101，与attempt12/source-evidence一致。该回执不认证本轮改动后的工作树。
- Go race所有包/build在该指纹通过；schema1→2兼容及正式schema1/2→3、三个迁移视口、原旧UI和路由journey部分通过。历史health重建exit1是验证入口要求的旧版预期红，不是当前产品health失败。
- 新控制台`stage=model-1280,result=failed,error_name=Error,write_status=201,check_status=200,check_revision_present=true,check_success=true,check_applicable=true,check_revision_matches=true`。服务核验修复已越过；模型与后续native旅程未通过。make browser exit1导致make verify exit2，不属于原生沙箱bind限制。

## 最小归因及候选

实际workspace.js的modelEditor生成remote标签“远端模型名”和routes标签“额外路由候选（每行 服务ID,远端模型名,优先级,权重,启用true/false）”。实际驱动旧getByLabel('远端模型名')无exact。VM执行实际字段构造及实际驱动片段，匹配数量2而非1，新增回归exit1；修精确选择后exit0。该源码缺陷已证，但seq13粗Error没有原始机制/具体子步信息，因此不是宿主唯一失败原因证明。

新增model-dialog/fields/create-http/grants-dialog/grants-http分段，保留原每次save状态与真实模型/授权旅程；失败仅输出数值/布尔和固定错误分类。下一宿主回执若仍失败，先据具体stage/error_kind/status/count收集最小诊断，区分驱动、产品、时序和环境，不能轮换假设性修复。不得输出错误正文、DOM、凭据、供应商URL、ID或对话。
