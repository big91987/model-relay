# seq53 原固定门禁失败边界

2026-10-08；Connector tests执行53；`make verify` exit2，timeout900s；日志98085字节，truncated=false。本节点经read_command_output(seq=53,offset=93000)读取至eof=98085，未以入口省略文本代替尾部事实。

被测候选editor-readback-route-r1，HEAD98511771871cf0951ecef716bbb55deb13e18da5；133件指纹6b6564c93b0c15bfefcd33955c3affd852fb8716963dea08e9600004a67b4bab。完整日志由平台保留，本文是有界摘要，不是复制的完整日志或当前候选通过证据。

尾部实际Go race包internal/relay PASS(437.985s)、tests/browser PASS(10.908s)、health-history PASS(6.041s)，随后build和browser。历史health重建的预期红属于health-history门禁，不是当前产品health失败。实际浏览器迁移schema1/2→3三视口与此前legacy/路由旅程有PASS；不据此将整套门禁标通过。

关键边界：1280控制台受控readback已到达，打印console-editor-interruption passed、controlled_get_hold=true、new_editor_save_enabled=true、service_posts=1、canceled_editor_posts=0。紧随其后Node v26.0.0报`route.continue: Route is already handled!`，位置console.mjs:90的barrier.then继续；console退出1，browser/verify最终exit2。

这次不是监听环境限制，也不是seq51分页检查点未到。该分段passed写在清理之前，因此不能证明其清理成功或后续1280完整控制台成功，更不能证明1440/390、summary、J01–J07/独立QA已完成。本节点不读取其他Run或旧工作区、不重演旧成功。

最小归因：锁定Playwright1.59.1源码Page.removeRequestInterceptor→notifyRoutesInFlightAboutRemovedHandler→Route.removeHandler会自动continue当前在途handler。旧finally并行release continuation/page.unroute，形成双继续竞争。实际锁定Route方法的隔离delegate诊断保留于route-diagnostic.log及playwright-source-diagnostic.json；证明源码机制，不冒充原生浏览器/网络取消事实。宿主再次执行原完整入口仍是必要验证。
