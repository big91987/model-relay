# seq51 宿主失败边界

- 原入口make verify，timeout900s；Connector exit2，98283 bytes，truncated=false。
- editor-lock-r1；HEAD `98511771871cf0951ecef716bbb55deb13e18da5`；133件指纹 `5b0822a01c8d0a92236fdf1a70d5e4e1fe6114782ddb643b86e002c0dee3d836`。
- 本节点read_command_output读取offset87000到98283/eof；完整原日志由平台seq51保留。不声明逐字复核全部98283字节，不把界面省略或缺失日志作通过证据。
- 尾段实证：internal/relay race包PASS（440.840s）、tests/browser Go包PASS（10.876s）、health-history包PASS、build推进browser。旧版health-history预期红灯不是最终失败原因。
- 旧UI兼容、schema1/2→3迁移三视口、health/recovery1280/390、多上游1280/390已输出PASS，只属于上述指纹，不移植为新候选通过。
- 新console停editor-held-readback-1280，Error/other、POST201、catalog200、members200、page_error_count0、request_failed_count1、editor已关闭、Save enabled。完整门禁Fail，不是监听限制；后续console1440/390、continuous/summary/visual等未到达。

实际collection('tenants')带分页查询，seq50 glob **/admin/api/tenants不匹配，检查点5秒内无法触发。锁定Playwright实际urlMatches边界见route-diagnostic.log：旧glob对分页/cursor false，修正regex true，成员/相似路径仍false。证明路由匹配缺陷，不证明原生取消机制；request_failed_count1具体原因尚未归因。只修驱动及回归，保留seq50产品save-lock/epoch修复和全部原断言。不改后台/业务数据/时钟、不延长等待，完整实际Go/Chromium仍交tests。
