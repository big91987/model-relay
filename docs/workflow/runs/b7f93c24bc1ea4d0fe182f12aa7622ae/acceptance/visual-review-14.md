# 执行14 实际安全图片与QA58-01复核

候选65201f488ae0969684625a92c23fb2c1e76bed30916b2c92a405b1e9e00f92a9；图片来自宿主执行13，不是原型或Owner预览。独立QA通过image viewer实际查看30张三视口主页面（十页×三）和19张补充图，共49张；其余132张仅解码/摘要保留，不宣称已人工查看。

路径以下均相对于 `evidence/attempt-14/host/browser`。原210件字节完整性见integrity.json；89张visual声明SHA全部匹配；181PNG都可解码，**4张全部像素同色**见png-audit.json。单色检测能证明空图，不能判断非单色图可用或证明渲染原因。

## 逐页对照（1280/1440/390均实际打开）

| 页面 | 现有基线与QA58-01项 | 本轮观察 / 处理 | 当前结论 |
|---|---|---|---|
| 工作台 | 真实当前周期/来源/长名及状态 | 三图可见真实1次/21微美元/来源中文；390局部表格和模型/预算入口 | 本项可接受；不因截图金额当供应商报价 |
| 模型托管 | 长公共名/来源/能力、状态/操作 | 三主图模型、来源中文、协议必要名称、完整状态和动作可识别；390长模型补图保留聚焦完整文本region | 分项改善成立，完整名可访问由真实键盘/DOM增强，不凭PNG可见前缀单独判完整 |
| 供应商与服务 | 原success英文、地址字母内断词 | 1280/1440核验通过中文、地址单行；390主图与右列图表头/操作可达；390失败/纠正图状态中文、原原因代码完整 | QA58-01服务中文项实测改善；1440self_hosted地址补图主体消失，长地址三视口视觉整体待复验 |
| 调用测试 | 协议名/同ID、取消和秘密边界 | 三图为未使用表单空态、秘密为空；不是普通/SSE成功截图；桌面fullPage侧栏顶点不一致 | 行为Pass来自实际UI链，不能用空表单拼出调用Pass；截图坐标归因待研判 |
| API Keys | 安全标识不内部断词、状态/轮换 | 390/1440完整安全hint、状态和动作；1280主图/右列图都是背景；1280首次分支Key图只有侧栏，1440分支主体向左偏移 | 1280视觉Fail，不整体关闭安全标识项；没有把hint当完整秘密 |
| 租户与成员 | 角色/邀请机器枚举 | 390/1280成员中文/邀请已接受；390self_hosted角色、身份、未发送邮件可见；1440主体移出大半、右列图全背景；1440首次角色图全背景 | 角色DOM行为通过，1440实际视觉Fail；未当外部身份认证 |
| 配额与预算 | 五维中文、费用/Token/无限0语义 | 390/1280可读，保留费用未知非零；1440左侧品牌/导航被裁去部分 | 未认定费用错误；1440截图布局证据不合格 |
| 用量分析 | 长模型与已确认/未知、筛选导出 | 390/1440指标可读；1280只有侧栏，标题/表格/筛选全部不可见；390首次长名聚焦文本可访问 | 1280视觉Fail，数据行为另记Pass |
| 定价与账单 | 长名/价格标识、USD历史不回算 | 390/1440价格、账单和支付未接通明确；1280主体整体向左，仅剩窄片 | 1280视觉Fail，不推断账单数据不正确 |
| 调用日志 | 已知状态中文/精确原因/完整安全标识 | 三主图成功/已拒绝中文，390详情公共名/ID完整region、未知Token忠实；桌面长页fullPage侧栏不位于顶端 | 文字项改善；fullPage采集/页面坐标失配需归因，不能依DOM全绿保证PNG布局 |

## 最小明确异常

| 文件 | 实际像素/人工观察 | 同回执声明 |
|---|---|---|
| visual-1280/keys-page.png | 1280×1149，全图单色背景 | keys rootFits/secretsClear/readability=true、非空4行 |
| visual-1280/keys-right-columns-1.png | 同尺寸，全图单色背景 | 对应表最右列已精确到达 |
| visual-1440/tenants-right-columns-1.png | 1440×1486，全图单色背景 | tenants最右列截图 |
| continuous-1440-self-hosted-roles-48273.png | 1440×1216，全图单色背景 | roles readability/root_fits=true |
| visual-1280/usage-page.png | 侧栏可见，主体全部背景 | 非空用量行、rootFits=true |
| visual-1280/billing-page.png | 主体左移，标题/账单内容不可读 | 页面/关键列检查全部绿色 |
| visual-1440/tenants-page.png | 主体仅左边残片，大部分为空背景 | rootFits=true/角色与操作检查通过 |
| continuous-1440-self-hosted-service-long-address-48273.png | 侧栏可见、主体空缺 | service-long-address、完整region键盘检查通过 |
| continuous-1280-self-hosted-key-safe-identifier-48143.png | 侧栏可见、主体空缺 | key-safe-identifier checkpoint通过 |
| continuous-1440-self-hosted-key-safe-identifier-48273.png | 页面主标题/左列移出可见区 | key-safe-identifier checkpoint通过 |

以上是图片事实，不是“产品必定有页面根横向溢出”或“Chromium动画已证明”的结论；DOM回执与PNG不同层需研发采集同一时刻事实解释。没有图片编辑、裁剪、涂改、替换或重截旧图。

## 人工查看的19张补充图

- visual-390/records-detail.png、services-right-columns-1.png；visual-1280/keys-right-columns-1.png；visual-1440/tenants-right-columns-1.png。
- continuous-390-self-hosted-{upstream_connection_failed,upstream_rejected,verified,service-long-address,roles,key-safe-identifier,normal-same-request,sse-same-request,usage-confirmed,model-long-name}-48450.png（全部十张）。
- continuous-1280-self-hosted-{service-long-address,key-safe-identifier}-48143.png。
- continuous-1440-self-hosted-{service-long-address,key-safe-identifier,roles}-48273.png。

PNG可见文本没有密码/完整应用Key/供应商凭据或完整对话；安全采集脚本清空输入/storage及秘密全文排除断言继续有效。图片没有秘密不等于图片已可用。QA58-01按分项记录改善，整体保持待复验；新增QA14-01返回development。
