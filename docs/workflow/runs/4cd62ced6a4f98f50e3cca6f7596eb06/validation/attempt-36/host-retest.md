# 执行36宿主复验要求

候选 safe-readback-resource-matrix-r1，HEAD98511771871cf0951ecef716bbb55deb13e18da5，来源见本目录 `source-evidence.json` / `source-after.json`。沿用design-v0.2.0/archive5f51d35fdc15ed7214749f9413879b4238097f1d6b6bbe2cc0d508d8e103cef7及原Run；不得读其他工作区、手改DB或平台检查点。

1. 运行未改的原完整 `make verify`，保存真实exit和全量未截断日志；入口source摘要须与本目录127文件当前指纹一致。原健康历史预期红灯、format/modules/vet/Node/所有Go race/build/migration/CLI/process/HTTP/SSE/取消及原native/路由/keyboard门禁不删。不单独用新的fault suite代替完整入口。
2. 新 `TestConsoleResourceMatrixActualHTTP` 要看到tenant/Key各五维0/null/正有限边界共10分例通过；`TestConsoleSharedResourcesActualHTTP` 要看到multi_key/old_new_generation的8争2、6拒绝、另租户1次，以及额度历史/双scope原子性读回。失败归属产品/driver/时序/环境按首次实证，不笼统归入沙箱。
3. 原console1280/1440/390在既有普通/SSE/cancel/立即轮换/真实成员之后调用 `consoleFaults`。先做标记清楚的historical旧editor隔离重建red，再解除route/reload当前asset。原操作标识/按钮消失须为expected failure；当前签发201/轮换200丢响应→GET503/404→同operation200必须确实恢复，秘密不重放、double-save/Enter仅一次已提交mutation，stacked rotation409，应用历史/usage/价快照不变。该旧行为重建不是原开发时真实UI执行事实。
4. 普通/SSE各native双发送：一POST、一准入、一同ID、一次实际测试上游dispatch、Tokens3+4与同ID详情；按钮恢复，原cancel费用可能待确认、不重放断言保留。每视口须生成 `test-results/fault-{width}.json` 和fault截图。事实JSON只允许枚举阶段、状态/次数/摘要，无秘密或原始异常。
5. 当前三视口原J01–J07连续旅程继续全部通过；J06保留fee21/cost7、Tokens3+4、旧revision0/快照相同、追加delta revision=before+1；Governance账本19/60及delta5/16不变。
6. 收集当前实际 `test-results/fault-*`（JSON及历史red/写前/unknown/503/404/recovered/双发送截图）和 `continuous-*` 等原产物，同字节脱敏归档到本Run新validation/acceptance证据目录，交接真实路径及摘要。没有跑出的文件不得在清单标通过。失败也保留phase和已有安全产物，勿仅转发被截断的整段日志。
7. 任一失败返回development，携首个归因边界和完整回执；全量绿后交独立QA按QA35报告及 `remediation-handoff-35.md` 验证是否关QA35-01/QA22-01。当前本地绿不跨层放行，真实供应商/外部身份支付/平台模型与发布仍分层Not Run。
