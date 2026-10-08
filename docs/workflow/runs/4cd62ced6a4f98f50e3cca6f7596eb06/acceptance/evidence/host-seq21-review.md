# 宿主固定门禁独立证据核对 — seq21

来源是本Run tests Connector执行21，不是QA自己成功运行网络测试。命令 `make verify`，回执exit0，timeout900秒，日志88327字节、truncated=false。

通过 `read_command_output(seq=21)` 按0→32768→65536→88327读取至eof=true，并使用39000、50000、56000重叠窗口核对当前Node/Go测试。归档未截断；工具展示会省略长消息中部，不把省略的文字补造成另一个本地“完整日志”。完整原始日志继续以平台seq21归档为权威；本文件是审查索引，不是逐字复制。下一节点仍可用同工具读取。

## 版本

- HEAD：98511771871cf0951ecef716bbb55deb13e18da5。
- 源码/测试/契约119文件指纹：71be54206d9850fa442f33c18ddd3c7aaba5590019543a78498a128ae0ab8350。
- Go：go1.25.6。QA在相同指纹读回，见source-before.json及source-after.json。
- 浏览器服务不是部署：临时Go服务、schema3、`console-controlled-verification`；受控真实HTTP/SSE上游。不是供应商、生产Actions或固定效果URL事实。
- 回执没有独立测试结束时第二个源码指纹。QA接收后两次一致，支持接收版本一致性；不补造宿主结束快照。

## 观察到的结果

1. health-history隔离重建旧版：内部exit1和21个方法失败是预期红灯，随后门禁写 `PASS expected historical HTTP red`。不是当前候选失败，也不是原开发时的成功记录。
2. format-check、modules、vet、Node152项（fail0/skip0）、Go race、build继续执行完成。Go当前Console治理、资源维度、身份范围、普通/SSE费用、迁移、真实进程崩溃恢复、价格精度测试均见实际PASS记录。
3. 正式CLI schema1→3和2→3，分别1280/1440/390浏览器迁移旅程通过；schema1→2属于显式兼容测试入口，不冒充生产CLI。
4. 旧native管理UI及多上游路由1280/390通过。不自动提升为新workspace相同功能的UI-E2E通过。
5. console-real-go-ui 1280/1440/390分别result=passed。断言源码为tests/browser/console.mjs：首次接入/核验/发布/授权/定价/Key签发和转交/普通及SSE/取消/即时轮换/邀请登录/未知签发读回/次数拒绝/十页入口/弹层13Tab及Escape返回/秘密清理。
6. 上轮keyboard-1280原路径当前已走完；不能据此宣称完整键盘业务操作或全部J01–J07已通过。现有脚本不包含预算提高后重发、租户暂停恢复、账单核对结账更正等必需步骤，详见整改交接。

## 边界

链接器LC_DYSYMTAB警告未阻断本次宿主门禁；不改写为产品错误或忽略真实失败。供应商、外部身份和支付、正式发布均Not Run。QA自己的make verify在health-history监听处Blocked/exit2，保留独立日志，不拼接成QA本地exit0。
