# seq17独立QA整改实现与待验

日期：2026-10-06。Owner：研发；状态：In Review / Pending Host and Independent QA。
恢复依据：[QA整改包](../../acceptance/remediation-01.md)、[首轮验收矩阵](../../acceptance/matrix-01.md)、[原真实失败](../../acceptance/evidence/connection-contract-01.log)、PRD AC-01/02/03/06/07/09、既有架构和交互。未重启需求/设计、另建路线、改验收预期或清除历史失败。seq16只证明08完整门禁通过，seq17明确No-Go，不给当前源码背书。

## 根因与最小修改

| 整改 | 已证缺口与责任层 | 实现及回归 | 当前证据边界 |
|---|---|---|---|
| QA-01 | checkUpstream仅要求data非nil，错误协议可成功 | configuration.go检查JSON Content-Type、object:list、非null数组及逐项非空字符串id，保留1MiB上限/完整JSON；真实受控成功列表补object:list；原QA负控不变 | 原独立负控局部真实转绿；新UI受控核验待宿主 |
| QA-02 | 出站context为30秒而非10秒 | 统一10秒父请求派生context，涵盖headers/body；新真实HTTP分别挂起头与body，要求实际超时/出站取消及API/存储失败一致 | 出站deadline负控真实通过；10秒实际头/body监听测试仅编译通过，未运行 |
| QA-03 | 默认unknown掩盖坏JSON/429/503/network确定失败 | 默认failed，协议/HTTP/网络安全code；deadline failed/upstream_timeout；父取消canceled，历史无终态仍未知 | API/持久化负控真实通过；真实两屏非法列表/401/429/10秒与编辑恢复待宿主 |
| QA-04 | 完成步骤仅标class，完整常驻当前任务 | renderSetup移动原五行/原按钮到details回看，保持原序号/ID/同请求证据；未知读取不保留完成事实，重启仅前三配置步骤完成 | VM原节点、顺序、失败、重启回归通过；双屏折叠/展开/精确详情/重启待宿主 |

新增取消核验按钮是当前操作控制器的abort入口，不另写配置/健康；忙锁去重、父取消传递、释放后显式重新核验由Go回归实际验证。UI对已abort但迟到成功显式隔离；离页/到期同原生命周期，119项VM含这些负控。取消只清本次核验，不发生成或清冷却，保留应用取消/秘密清理契约。

## 强制边界补齐资产（不冒称已验收）

- `connection_lifecycle_test.go`：免监听handler单上游去重、出站context取消、API/存储canceled、启停/执行代次/健康不变、gate释放及显式重入；真实HTTP头和body分别10秒超时、供应商context停止及落库核对。
- `routing_contract_regression_test.go`：优先级-1/101/非整数、权重0/101/非整数、空远端/未知引用/null拒绝无部分写；合法0/100及1/100；全冷却503零出站、较旧成功晚到不抹冷却、启停保留冷却、凭据/候选新代次隔离、重启健康未知。
- `routing_http_test.go`：双候选首工具增量已提交后EOF、idle超时、客户端cancel分别核对终态/一尝试/备用零访问/无DONE；不是单上游工具回归的推定延伸。
- `routing.mjs`：每屏从真实页面新建A/B；非法/401/429核验、忙时取消并核对实际供应商GET/取消计数及持久状态、10秒超时、离页abort；重名/变址缺凭据拒绝→保留页面非敏感草稿→纠正并新write/readback→再核验。后续创建/编辑/取消/恢复候选、签key、发生成及停止全部原生Tab/typing/Space/Enter/type-ahead，去除fill/selectOption/click/check后备；保留精确标签、SWRR计数、优先级、冷却、SSE与不重放断言。
- 同driver每屏成功后核对五步进入默认折叠区、键盘展开/进入精确记录/收起；隔离真实服务重启后重新登录、前三已完成后二待调用。原journey只增加完成区原生展开再做原精确ID断言，未删或降低断言。
- Native选择轨迹及控制外部行为只作可归因证据，无DOM设值/业务API预置/SQL改业务数据/凭据正文打印。历史迁移主旅程仍由原入口执行。当前新增HTTP/UI资产的存在与编译不等于Pass。

## 真实局部验证

| 记录 | 命令/结果 | 支持范围 |
|---|---|---|
| [contract-01](qa-repair-contract-01.log) | Go race原QAConnection负控选择，exit0 | 出站预算、非法协议/已知失败合同；不证明监听/浏览器 |
| [boundaries-01](qa-repair-boundaries-01.log)、[boundaries-02](qa-repair-boundaries-02.log) | Go race选择原QA负控/晚核验及新增Cancel/Integer/HealthEpoch，exit0 | handler/transport/持久化及健康状态边界；非真实网络机制 |
| [VM-01](qa-repair-vm-01.log)、[VM-02](qa-repair-vm-02.log) | node --test，分别116和119项exit0 | 当前UI状态层；不作原生键盘/截图/浏览器证据 |
| [validation-01](qa-repair-validation-01.log) | format-check、module verify、vet、119项VM、全包零测试编译、build、driver语法、diff逐命令exit0 | 静态/局部层，零测试编译不是全Go suite |

运行输出已脱敏临时本机目录，保留ld警告，不把警告视为断言错误。当前[10源码指纹](final-source-10.json)与最后同版本检查见后续回执；公共delivery receipts不进入摘要，避免自引用。Workflow/manifest未改，继续与[平台模板联合身份](platform-template-provenance-01.json)核对；10产品摘要本身不覆盖它们。

## Trellis-check与宿主交接

最终10真实检查完成：85文件，sha256 `ff456d824299075fe8d98104a2cd51671ca41f08421f18cb1dd053d80efe93c8`；[同版本最后检查](qa-repair-final-validation-02.log)绑定该摘要，10条命令全部exit0，119/119 VM及所选Go race实际通过。format/module/vet/全包零测试编译/build/driver语法/diff通过；不扩展为完整Go/HTTP/浏览器Pass。[最终审计](qa-repair-audit-01.log)9文档/106相对链接问题0，新增DOM ID唯一，driver无设置值/click后备；QA原测试、Workflow及manifest与独立QA版锁定摘要均未变，原Makefile/health-history与HEAD无差异。包含Workflow的整体HEAD差异检查exit1来自前序已授权安装模板升级，已分层记录，不误报本节点篡改或门禁通过。脱敏检查未发现本机绝对路径。宿主须使用10，不使用保留09或08。

收尾静态审查发现原页面hash在重启后保留records，不应只等overview；仅修driver为重登录后显式原生导航概览，并要求computed focus outline实际可见。09（[保留指纹](final-source-09.json)及[同版本局部检查](qa-repair-final-validation-01.log)）保留，不覆盖它；最终10与[最后检查](qa-repair-final-validation-02.log)绑定。此为静态driver修正，不声称曾实测浏览器失败或已证明原生行为；实际新的完整宿主入口仍待运行。

已恢复真实Trellis backend/frontend index、runtime/console/routing、DOM/ownership/verification及shared guides；当前脚本packages输出Single-repo、backend/frontend，规范有源码合同而非空模板，不需bootstrap。新协议/错误矩阵与UI生命周期已原位update-spec；跨层从POST输入、出站context/Body、CheckResult存储、当前API到UI安全状态核对，取消不写路由配置，折叠不克隆证据节点。未改认证/CSRF/SSRF/Go依赖/schema/可信门禁/平台源或历史Run。

宿主下一步仍为900秒 **make verify**：先核对10＋模板联合身份；保留完整新日志、退出码、实际源摘要及路由JSONL路径，不覆盖seq16或QA失败。新真实HTTP与双屏旅程走原入口，不用局部选择替代。若失败按实际stage/test/断言和安全事件归因，分产品、driver、异步或环境，不换按键/放宽合同。完成后交独立QA逐AC判断，本节点不修改其首轮矩阵结论。

既有本机listen/Mach稳定限制未改变，故不重跑同一必然受限完整命令，原[make失败](make-verify-01.log)保留并关联当前10；必要静态/免监听检查继续实际运行。新HTTP、浏览器、完整当前固定入口 **Pending Host**。DEP02～04可信controller新旧Go链、DEP06～07正式Actions升级/隔离恢复及真实供应商 **Not Run**；不扩大权限、不自动合平台源PR、不操作安装/Workflow/平台数据，不声称发布成功。
