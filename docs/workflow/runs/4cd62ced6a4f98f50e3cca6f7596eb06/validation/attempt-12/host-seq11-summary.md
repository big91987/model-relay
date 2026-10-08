# 宿主seq11失败边界

- 命令：`make verify`；tests执行11，exit2，900秒上限。保留原回执，不属于原生监听环境限制。
- 回执来源：本Run `read_command_output(seq=11,offset=80000)`，next_offset86204/eof=true/truncated=false。恢复相关末段，不声称未读取的前段全量日志已归档到本文件；原完整86204字节仍在平台保留命令日志。
- 已测试指纹：HEAD98511771871cf0951ecef716bbb55deb13e18da5，119文件，Go1.25.6，工作树bd81e49e060f7e0494e1217947a0e9367fab394b19a64bfe0845b427addbecc6。属于attempt10候选，不能替代当前候选Pass。
- 当前Go relay race包PASS（330.653s）、迁移CLI目标契约包PASS、build通过。health-history是期望旧基线红的重建门禁，不将其旧HTTP405断言失败误判为当前产品失败。
- 浏览器：schema1→2兼容路径通过；正式schema1/2→3的identity/keys/pending-history均在1280/1440/390通过；旧控制台实际Go UI、恢复/分页/health和多上游路由1280/390通过。受控HTTP/SSE不证明真实供应商或正式部署。
- 原失败行：`{"suite":"console-real-go-ui","stage":"service-1280","result":"failed","error_name":"AssertionError"}`；随后console browser exit1、make browser/verify exit2。新控制台模型/价格/Key/邀请/预算及后续视口未完成，不能推导通过。

## 最小诊断与归因

该stage含服务保存201断言和核验200断言，旧日志没有status或具体断言，单靠该行不能认定唯一原因。源码边界已证：workspace renderPage服务行调用POST核验发送{}，configuration.go checkUpstream要求expected_execution_revision且缺失返回400（在网络dispatch前）。修复前实际render VM两例均红，记录red-workspace.log；schema3实际handler回归证明空值400、版本不符409且无dispatch。

补充回归初次200状态但核验不成功，最小检查确认现有Go协议必须object=list和data数组；本回归及console.go的受控列表缺object标记。red-go-check.log保留这一受控fixture失败，按原协议修正夹具，而非放宽产品；完整native的新成功/applicable/版本断言仍留宿主证明。未用夹具机制代替浏览器事实，未猜测性更改时序。

## 下一项真实验证

当前候选完整make verify仍待宿主。保留所有迁移/native/HTTP/进程门禁；失败按service-create/check-http/check-contract/check-dialog和白名单status/版本布尔诊断返回研发。seq11局部成功不可跨工作树复用为完整通过。
