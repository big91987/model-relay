# seq32 宿主门禁首失败恢复

- 来源：正式 tests Connector seq32，make verify exit2；归档93546 bytes、truncated=false。执行33用read_command_output读取offset87000→93546/eof=true/truncated=false；工具归档保留完整日志，本文件不是完整日志替代。
- 被测候选：HEAD98511771871cf0951ecef716bbb55deb13e18da5，124文件、Go1.25.6，SHA256 de6d4d88d809bedacf06d00693c51cb4fb7159b2d2c986c3bf80a025a270b978。
- 尾段显示Go relay包PASS、browser Go包PASS、build完成，兼容及生产迁移三视口、旧实际Go UI及路由旅程、新console1280/1440/390通过。连续1280 J01/J02/J03/J04/J05/J07 passed，J06失败；连续1440/390未到达，不能补记通过。
- 首失败：stage=J06-reconciliation，action=authoritative-read，HTTP200，journeys.mjs:412:137，AssertionError。编辑器已关闭、workspace可见、secret_present=false；非监听限制或保存失败。
- 源码该位置断言reconciled.charge_entries[0].fee==='21'。生产查询按ordinal,revision返回追加账本；settleAdmission先记录未知用量零确认revision0，核对追加差额而不改原记录。首条费用不是累计费用。该缺口与冻结PRD“未知待确认、追加核对只计差额”契约不符，应修driver而非生产账本。
- 原usage_facts全部守恒：fee126→126/reserved0→0/calls9→9/pending0→0/concurrency0→0；membership role=finance/diagnostic0/tenant0/options0。此前J05修补在此宿主候选实际通过，但不代表执行33或QA22已通过。
- 尚待证：执行33新累计及历史断言的真实浏览器结果、J06结账/更正/导出余程、1440/390连续完整旅程以及独立QA；不据此宣布唯一native根因已关闭。
