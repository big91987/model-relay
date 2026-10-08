# tests seq49：保留的首失败

- 原固定命令：`make verify`，timeout900s，exit2；平台保留日志98848bytes、truncated=false。
- 候选：management-summaries-r1；HEAD98511771871cf0951ecef716bbb55deb13e18da5；133件工作树指纹2b1297e865a8b595af98cd7594900d4a6c0506729c0e46de5ba3a05640a2a962。
- 研发执行50实际通过 `read_command_output(seq=49, offset=90000)` 恢复最后8848bytes，next_offset98848/eof=true/truncated=false。未重新执行宿主命令，未以日志缺失推断任何通过。
- 恢复段显示当前Go relay包race测试444.647s PASS、browser Go包11.153s PASS、health-history测试5.976s PASS、build完成；原兼容schema1→2及正式schema1/2→3迁移三视口、旧UI/路由两视口、console-faults1280及console-real-go-ui1280均输出PASS。
- 首失败：下一视口的`console.mjs`第97行service-create等待`/admin/api/upstreams` POST，30s waitForResponse超时，Node未处理拒绝导致退出，console浏览器exit1，make外层exit2。该行原有一个额外等待器，另一个在save中。没有落出原白名单snapshot；因此不能据宿主日志独自证明Save disabled、后端未收到请求或唯一根因。
- 健康历史输出针对隔离旧baseline的预期红灯，不是当前服务健康回归失败。宿主失败不是监听限制。
- 后续三视口完整新摘要/连续旅程/视觉门禁在这个候选未全部运行，不继承旧seq46的绿灯；原QA47 No-Go及QA47-01/02继续有效。

## 最小区分证据与修复

源码`edit`在提交时禁用Save，响应确认后关弹层，随后等待syncTenants/renderPage；导航提升epoch使旧finally不能解锁，而新edit仅重置hidden，不重置disabled。`editor-red.log`对实际函数的受控VM回归证实新弹层仍disabled；修复后同一测试`editor-green.log`exit0，另验证旧finally不能解锁后一个正在提交的弹层。这是已证产品生命周期缺陷，与宿主超时相符，但不是原宿主事件的浏览器追溯证明。

新原生有界旅程只在隔离测试服务hold这一GET，不替换后端写入：当前响应201后导航、打开下一服务编辑器、要求Save enabled、取消并检查实际POST总数仍1。1280/1440/390的真实产物由下次宿主产生；失败时保留stage/safe snapshot。不会用夹具定义证明真实网络取消机制。
