# seq30 宿主失败归因边界（执行31恢复）

Connector `make verify` exit2，92708 bytes，truncated=false；HEAD `98511771871cf0951ecef716bbb55deb13e18da5`，124文件指纹 `6263734f52bb1631801f5db0f27d44f7a12de0e64f1fad4ffdf003c244b6dd92`，Go go1.25.6。

`read_command_output(seq=30,offset=86000)` 到92708/eof/truncated=false完整恢复尾段；offset64000的补充读取到相同eof，但本报告不把工具展示省略部分当逐行完整审查。全量归档由Connector保留。尾段受控HTTP/browser包PASS、迁移/兼容及新console1280/1440/390 passed；连续1280 J01–J04 passed，J05失败。J05后续/J06/J07及1440/390连续未完成，不跨版本复用Pass。

原实际失败JSON从既有 `test-results/continuous-1280-89946.json` 原样保存在 [host-seq30-continuous-1280.json](host-seq30-continuous-1280.json)，不是重建日志。原receipt为 `J05-resume / authoritative-read / last_status:200`，expected call200/dispatch_delta1；usage_facts期间相同、费用126→126、预留0→0、calls9→9、pending0→0、concurrency0→0。

## 纠正此前过度归因

执行29及更早总览把粗标签解释为“恢复读取后的费用比较失败”；seq30新增值并不支持这个结论。driver在恢复比较后还有两次成功调用及成员读取/降权/诊断回收/退出/跨租户断言，直到J05完成都未更新stage。现有证据只证明 **J05某后续断言失败**，不能证明费用守恒失败或唯一位置。最后一次守恒检查中全部字段相等且action又变回authoritative-read，支持“可能在后续成员步骤”的推断；缺安全的源码位置或阶段记录时不升级为已证根因。

源码另外确认独立driver缺口：product refresh.onclick先await syncTenants，之后navigate才替换loading；原测试refresh click后ready只等待loading隐藏，旧页面本来就满足，可在新scope同步前返回。新增延迟helper回归实际执行原函数，首先red（finished=true），补等待预期零选项后green。此回归证明helper不能保证刷新完成；**不证明Chromium seq30实际失败于此**。保留原零选项/status/费用/dispatch断言；新增成员阶段、read expected/actual和安全source line/column供宿主确认，生产计费/权限/DB/UI不改。

下一步完整原make verify；若同一失败仍存在，按新增stage/failure_boundary/read_facts/membership_facts选择下一项修改，不再根据旧resume标签修改计费或轮换假设。当前宿主失败不是环境限制；稳定原生监听限制是另一份既有证据，不混淆。
