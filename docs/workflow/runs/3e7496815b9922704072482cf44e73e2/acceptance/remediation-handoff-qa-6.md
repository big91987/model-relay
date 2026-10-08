# development 补证交接

## 本轮结论与版本
Blocked，不放行发布。HEAD cd413b5aa24a372a19c19af999a6ef61b5c41572，142件源码摘要8e82aaa123df25a4758a1939fb2d018713831507aee337e18dfa6100c49f619e。实现最小正确，tests宿主seq5 exit0且同版；不要无依据改产品。独立QA报告/矩阵/缺口见同目录report-qa-6.md、matrix-qa-6.md、defects-qa-6.md，具体证据索引evidence/index.md。

## 必须补齐的行动
1. 由正式宿主测试Connector在独立候选工作树/副本运行原make verify。源码必须同戳；记录开始HEAD/142件摘要、独立目录标识（相对/脱敏）、node_modules不存在、NODE_PATH unset、没有全局模块或手工预先npm ci。匹配Chromium/系统库及包源为已有宿主前置，不扩大Agent权限。
2. 同一执行保存完整脱敏stdout/stderr、时间、最终退出码及开始/结束状态，安装后Playwright/core1.59.1与锁一致。必须从make verify自行安装依赖到所有历史/Go/race/build/browser门禁通过，不靠定向绿或make -n。若宿主此前确实已完成合格干净运行，可提供当时真实前置回执；不得事后写证明冒充当时测量，否则需补做一次有效干净执行。
3. 正式测试回执存本Runverification下实际非空文件，完整seq5原日志也补档（保持首次失败和历史证据）。索引指向具体文件，不传目录充当日志；未经审查的原始凭据/会话/机器路径不提交。缺片段/截断不算完整。
4. 再由独立QA复核同版与干净条件即可；没有业务改动不用重建需求/设计，不重复沙箱监听失败。若真实新失败出现，保留事实并最小修复、回归；严禁更改期待或删原门禁。

## 不得回退
零软预算拒绝、有效UTC高水位/期限/重启/备份恢复、schema1/2→3守恒、design-v0.2.0三视口真实Go UI及原route-cleanup。原生QA UIBlocked、真实供应商Not Run继续保留，与宿主受控Pass分开。

## 后续发布
补证QA通过后走report，由授权协调者/Connector新PR、精确已验提交合并，再正式Deploy Model Relay locally prepare及Owner deploy=true target=preview。成功Deployment/environment_url、healthz.version/schema3与真实admin同版才闭环；旧数据/密码不重置、不改库、不替代服务。控制器同SHA日志覆盖由Platform维护源负责。当前没有新发布或上线回执，不宣称完成用户全部任务。
