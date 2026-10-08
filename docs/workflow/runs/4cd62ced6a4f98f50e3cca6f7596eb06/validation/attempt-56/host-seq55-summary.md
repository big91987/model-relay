# seq55 固定宿主入口失败事实

- 命令：`make verify`；Connector tests执行55；timeout900s；exit2。
- 保留日志103324字节、truncated=false。执行56读取offset98000至103324，eof=true；未把输入省略段或缺失日志当作通过。
- 输入候选editor-route-cleanup-r1；HEAD98511771871cf0951ecef716bbb55deb13e18da5；134件指纹6c674fcb21eba6c31818a9de7d564c520c3ef2c14a72fc0065308088a835c9a3。只绑定上一工作树。
- 原Go/race包、正式CLI迁移schema1/2→3三视口、原版本与routing浏览器路径已到达通过回执；这不是当前新候选完整通过。
- 新编辑器中断1280/1440/390均已输出清理成功后的passed：受控GET hold、新编辑器Save enabled、service_posts1、取消编辑器posts0。三个console-faults与console-real-go-ui分段也通过。
- continuous1280已输出J01/J02/J03/J04/J05/J07 passed；J06未完成。管理摘要、visual-current-period、63图最终门禁未到达，不计Pass。
- 最后失败为`page.waitForResponse`30000ms超时，原`tests/browser/journeys.mjs:257`创建等待，调用点476（核对后日志页详情）。Node未处理拒绝中止，原make browser/verify返回2；不是原生沙箱监听限制，也不是seq53路由重复continue错误。

## 最小归因及未证范围

实际产品records分支（`internal/relay/web/workspace.js`）每行提供“详情 / 核对”；“查看同Request ID日志”只在callForm创建。J06在调用日志页检查核对后列表，再误用默认call动作。执行56实际helper回归17/20、三项失败复现入口不匹配；修后20/20。HTTP等待/点击改同时观察，避免响应等待拒绝脱离原catch。

这是源码和驱动契约证据；不能证明Chromium修后已经到达、账本/CSV后续断言或全部宿主门禁通过。下一同Run固定tests对新source-evidence指纹完整执行，不删除原J06或用API查询替代UI详情入口。
