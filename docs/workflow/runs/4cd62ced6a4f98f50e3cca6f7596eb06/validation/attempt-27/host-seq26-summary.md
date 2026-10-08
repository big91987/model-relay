# tests执行26失败边界

- 固定入口`make verify`：Connector exit_code=2；完整归档89976 bytes、truncated=false。HEAD `98511771871cf0951ecef716bbb55deb13e18da5`，122文件指纹 `ed726cdd9609135b02904e2022ef20faa65e1b0b1796e723f9d1af1293a8f071`。
- 执行27恢复归档 offset84500→89976/eof=true，结合入口原回执确认末段；没有声称逐字复核全部89976字节。完整原输出保留于平台seq26，不把入口省略的中段当通过依据。
- health-history旧基线的预期红不是当前产品失败。尾部当前Go测试、build、schema1/2→3迁移三视口、原UI普通/SSE/取消与路由、console-real-go-ui三视口报告通过；这些是该旧指纹的局部事实，不等于整个入口exit0。
- continuous-console 1280：result=failed，stage=J01-responsibility-chain，action=proxy-call，last_status=502；initial_counts upstreams/models/tenants各0；checkpoints=[]。J01未完成，J02–J07及1440/390连续旅程未到达。editor_open=false、editor_invalid=0、editor_failure=false、workspace_visible=true、secret_present=true（仅布尔，不保留秘密）。
- 已越过执行24的新安装空态问题。本次不是监听限制。生产validResponse要求普通完成message.role=assistant；实际journeys.go已知/未知受控响应均遗漏role。本轮实际fixture回归先红（两个分支相同失败），生产validator保留缺role/错role拒绝。源码和局部回归证明确定契约错误；seq26没有终态明细，不能认定这是所有后续宿主失败的唯一原因。
- 本轮只修测试上游，附加有界GET诊断和回归；不改产品权限、协议校验、SQL事实或冻结材料。原seq26 exit2、QA22 No-Go均保留。宿主必须完整运行新指纹make verify，不将未到达旅程补写Pass。
