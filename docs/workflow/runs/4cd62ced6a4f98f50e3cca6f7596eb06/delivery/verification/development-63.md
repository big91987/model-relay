# Development63 当前整改交接

状态：development实现与必要非监听回归完成，T001 In Review；宿主完整tests和独立QA待重新执行。当前PR27 head为66f1328ada75ea90e5638fbf2bd58663ac86347e，未合并；本次未提交改动不能沿用QA58源戳为新版门禁。

只整改合并前P1-1零软预算和P1-2持久有效时钟，冻结design-v0.2.0 / architecture-r1范围不变。详细边界、实际输入摘要和验证记录见 ../../validation/attempt-63/checks.md。候选称budget-clock-r1，当前代码版本=上述HEAD+待验工作树。

关键路径：development63→tests原make verify（包括历史候选真实HTTP红灯重建和当前修复HTTP绿灯）→独立QA→report→publish→pr更新同一个PR27。禁止重复PR、替换main、手改数据或重用seq57/QA58给新代码背书。

QA58-01 P3及独立self_hosted首次网页分支仍保留，未在当前冻结修复混入；供应商、外部身份/支付与平台H未验项仍Not Run。用户已授权测试仓合并/正式部署，协调者在新门禁与QA核验后通过verified-main Deploy Model Relay locally执行，preview5545只是配置目标，不是已发布事实。最终交付仍须真实Deployment environment_url/admin/healthz与merge/Actions/version绑定，不用8793/旧main/临时服务冒充页面。

原生稳定监听限制仍适用，不重跑必败完整make verify、不提权；现有能力完成非监听检查后交给宿主固定tests。真实HTTP红→绿未实测前不标P1 Closed。

当前140件源戳ca8ec50aa168cf16ac2e24edfe214aa44d06a05846c34deadbc3c7cb6c02904a，实际[源回执](../../validation/attempt-63/source-evidence.json)、[材料19件重核](../../validation/attempt-63/material-integrity.json)。最终Console/新增期限排队与跨进程恢复race exit0（85.696s）、193Node/format/module/vet/build/历史分类器race exit0；实际日志race-current.log/static-current.log。源码新增Store持久有效UTC与正常schema3兼容建表、原锁取得后采样、零soft预算拒绝；原make verify仅追加准确PR27历史HTTP红灯重建，保留原所有层断言。Make verify未在本轮原生执行，tests须按同源戳完整运行900s配置并保留真实完整日志；不得把历史重建成功当原开发时网络记录。
