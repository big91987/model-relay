# 宿主04失败与候选标签修订

2026-10-06；研发返工；D4/D5；AC-02/03/12。当前状态 In Review；完整门禁仍未通过。

## 1. 真实宿主边界

- seq9：04摘要 `5b30123a1570e5cef7c1cf18ddf6243753d6e5518d2962523984f2cdef4743e0`，`make verify`在300秒预算中断，exit -1，部分执行不能当整体Passed。平台提供已确认进程静止、工作树未变；维护源f23bc3e经回归/复审/标准升级，原manifest仅把tests预算改900秒。这是平台反馈来源，不是本节点独立复跑平台或修改安装的证据。
- seq10：同04摘要、Go1.25.6、HEAD `85f6d17c07c5fddeb2d6a39a5a067769a11bb4cc`，同入口900秒预算，真实exit2。平台保存完整脱敏日志74586 bytes，truncated=false；可按seq10/offset读取，不能用显示截断的output否定其完整存档。本节点读取46000段及68000至eof核对CLI与末端边界。
- CLI包全部PASS（含真实安装/重启）、relay包PASS（216.526s）、health-history分类PASS、build通过。历史旧版health测试预期红灯不是这次门禁根因；链接器警告保留，不把警告作为失败归因。
- 主迁移旅程PASS：精确旧main实际UI形成非空数据，schema1→2、aliases=2、keys=2、historical_requests=2、controlled_posts=4，旧binary回退保留升级前身份。实际Go控制台原旅程1280/390均PASS。只属于04受控产品证据；不是当前修订版整体通过，不替代独立QA、正式Actions或供应商。
- 实际失败：`Real routing browser journey failed at create candidate set 1280; no credential or response-body diagnostic`；browser子make exit1，顶层make exit2。多候选调度双屏旅程尚未通过。

## 2. 最小归因及修改边界

已证源码事实：`renderCandidates`把动态select及其option文本嵌入文本为“上游”的label；driver用`getByLabel('上游',{exact:true})`。本次锁定安装的Playwright注入代码`getElementLabels`对label调用`elementText`，后者递归拼接后代元素文字，精确label文本因包含所有option名称而不再等于“上游”。这是可从实际维护源码证明的标签/driver契约缺陷，而不是供应商、Go路由或数据库故障。

尚未证实：旧catch将整个候选阶段错误归并，未保留具体locator/键盘/提交原因，因此不能宣称它是宿主唯一故障或实测已消除。本机最小真实Chromium setContent诊断启动即Mach权限拒绝，exit1，见[受限探针](candidate-browser-probe-01.log)；未到DOM断言，不能以模拟夹具证明原生机制。原先本机监听限制未改变，不重复必然受限的完整make verify。

修订范围：仅产品候选select用create/edit+行号唯一id及独立`label.htmlFor`，不改变字段、原生键盘行为、候选值或配置API；VM新增create/edit两行关联回归。原routing driver只增加细分阶段和有界白名单诊断（locator数量、indices/counts/booleans、最后12项model method/status），保留Home/ArrowDown/Tab及最终选项完整匹配、原请求计数/调度/回退/停止/费用断言。没有selectOption/DOM值赋值兜底、API/SQL预置、超时放宽、产品延迟或门禁削弱。若再次失败依据精确新边界继续归因，而非轮换按键假设。

## 3. 复验及交接

- [VM日志](candidate-label-vm-01.log)：112/112，exit0；只证明DOM结构与原有VM契约。
- [静态与构建](candidate-label-validation-01.log)：实际局部检查结果见日志，不替代原生宿主。
- 当前身份见[05指纹](final-source-05.json)；04快照、旧失败及seq9/10回执全部保留。完整`make verify`交900秒宿主tests，须先核对05身份，保存新完整日志和真实exit。有故障须保留`routing_failure`白名单摘要及原断言，再回研发。通过后才独立QA。
- 正式Actions非空升级、获授权隔离失败恢复、真实供应商仍Not Run。平台源PR不自动合并，本节点没有修改平台controller/install/Workflow/DB、Git发布或触发部署。
