# 公开产品行为观察与裁剪

2026-10-06 通过公开官方文档查阅用户提供的七个页面。仅研究信息架构、操作反馈和行为；没有访问本机 New API、产品实现仓库、数据库模型、图片资产或凭据。下列观察用于减少状态语义和诊断闭环不确定性，不是复刻清单或供应商联调证据。

| 官方来源 | 本次查到的公开行为 | 对本轮的采用边界 |
|---|---|---|
| [New API 渠道管理](https://docs.newapi.pro/zh/docs/guide/feature-guide/admin/channel) | 渠道列表显示状态，保存后可另执行测试并得到结果；包含编辑和启停操作 | 分离保存与调用核验、状态与上下文操作；不复制页面资产、批量功能、多 Key、商业范围或多渠道实现 |
| [New API 日志与统计](https://docs.newapi.pro/zh/docs/guide/feature-guide/admin/log) | 独立调用日志入口，按模型／渠道／令牌等条件查询 | 采用调用后进入诊断的独立入口；本轮精确找回同请求，保留分页，不引入完整统计系统 |
| [LiteLLM Admin UI](https://docs.litellm.ai/docs/proxy/ui) | 管理界面连接建模、签发密钥和使用观察，具有明确登录入口 | 采用连贯管理任务引导；沿用现有正式初始化和权限隔离，不照搬其登录和成员模型 |
| [LiteLLM Model Management](https://docs.litellm.ai/docs/proxy/model_management) | 模型列表显示公共名与提供者模型；详情可编辑及另行测试连接 | 采用别名／真实模型和状态区分、保存与测试分开；不引入价格、配置来源管理、多供应商系统 |
| [LiteLLM Virtual Keys](https://docs.litellm.ai/docs/proxy/virtual_keys) | 文档描述虚拟密钥用于调用的访问管理能力，涵盖模型访问与预算等治理 | 本轮只采用应用访问入口这一概念；一次展示和清除来自用户与本项目保密契约，不宣称参考产品具备完全相同语义；高级治理后置 |
| [LiteLLM UI Logs](https://docs.litellm.ai/docs/proxy/ui_logs) | 文档区分成功／错误日志，默认不存请求／响应内容，另有内容记录选项 | 采用安全元数据诊断思路；本轮坚持不持久化完整对话，不添加内容记录开关 |
| [LiteLLM Routing](https://docs.litellm.ai/docs/routing-load-balancing) | 文档目录单列负载均衡、路由及故障切换 | 仅核对公共路线第 2 项边界，本轮不实现路由、重试或自动切换 |

原创设计原则：围绕“理解状态→准备访问→调用→查同请求”组织，使用真实数据、明确依赖和可恢复反馈；不复刻商标、整套导航、商业功能、截图或文案。具体原型由 design 节点评审，本次没有创建或验证视觉原型。
