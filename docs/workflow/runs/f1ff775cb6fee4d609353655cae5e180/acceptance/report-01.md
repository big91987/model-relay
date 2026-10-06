# 首次配置、调用与诊断切片独立验收报告

> 日期：2026-10-06（Asia/Shanghai）；独立执行开始10:14。
> 仓库/分支：big91987/model-relay，workflow/f1ff775cb6fee4d609353655cae5e180。
> HEAD：d62cd51e8de2127548e1ac7469d349443d1bfe3d；56文件工作树摘要：dc9990248951dbfa54fda356975ffbef0dcfc13fef4add95e4a001eda74d69f4。
> 制品/环境：当前未提交前后端，隔离真实Go/SQLite服务，healthz=dev；不是已部署main验收。Go1.25.6/Darwin arm64、Node26.0.0、npm11.12.1、Playwright1.59.1；Chromium具体build无回执。
> 验收模式：首次独立全量，仅公共路线第1项。
> 结论：**No-Go，返回 development**。

## 1. 结论摘要

seq18宿主原样 `make verify` 退出0，源指纹与QA当前对象完全一致，真实Go首配1280px及存量390px主旅程通过。QA实际重跑60项前端回归和3项后端race契约均通过；但新增四个设计边界测试失败：上游保存拒绝未清凭据、生成拒绝未清key/message、写成功后读取失败未说明已写、刷新后概览成功证据动作未定位其同ID。前两项为同一失败清密主题；清密和同请求动作建议P1，未关闭，不因门禁绿灯放行。失败是当前app.js的确定性Contract证据，原生异常旅程尚未执行；必须补原层UI回归，不能称网页缺陷已红转绿。设计已明确预期，无需求方向或架构冲突，推荐实现/覆盖返工而非改基线。

## 2. 本轮覆盖

基线顺序为最新授权→[PRD](../product-definition.md)→[架构](../console-architecture.md)/[交互](../console-interaction.md)→测试/回执；实现不反向决定预期。验收Skill、SOP及产物契约完整读取，按单管理员/单实例小应用裁剪，不造组织级平台文档。

[矩阵](matrix-01.md)逐项覆盖18AC及其分层子场景，[旅程](journeys-01.md)记录连续步骤。1280px正式Init只准备管理员，业务空配置通过UI保存上游、创建别名、签发、普通/SSE、同ID记录；390px同服务从真实持久成功态继续，不称第二次空安装。每个视口同角色/会话顺序操作，显式测试退出、重启后失效和重新登录，非他人代填。API仅对账，控制端仅夹具流、重启和日志审计，没有API/SQL预填业务中间步骤。

已测：六页直达/刷新/前后，密码显隐/AX/零提交，局部原生键盘和两屏布局，版本原值/5秒隔离，数量和空/保存/成功区分，创建/映射编辑/取消/停用key，一次密钥与占位示例，普通/工具/SSE增量/DONE/取消/防重复，同ID安全元数据和0/null，中文错误、存储/日志/截图保密。Go补强超1000计数、权限/安全拒绝/存储fail-closed/配置竞态；前端补强忽略abort晚到、pending session已知期限、工具有界不执行。

这些只代表矩阵列明的子项。完整键盘创建/启停/定位/分页及依赖重新启用等无原生连续证据，QA-G01须补用例。成员/租户、集群、商业运营、删除对象及后续路由能力不适用，不补造阶段。

## 3. 缺陷及进展

| ID | 等级 | 用户能看到的问题 | 复现步骤 | 影响 | 定位信息 | 当前状态 | 回归资产 |
|---|---|---|---|---|---|---|---|
| QA-D01 | P1，新提出的强制保密契约缺陷 | 保存/调用失败后敏感输入留在页面 | 非法地址+非空凭据保存→400；或输入应用key/message调用→应用401 | 交互§4/§7要求失败清密保公开选择，实际失败不清；未发现持久存储/日志外泄，不夸大为泄露事件 | app.js guarded/上游提交281行/生成catch；AC07及AC12设计分支 | Open；Contract Fail，UI专项Not Run | QA-01/04；真实Go两屏异常回归待补 |
| QA-D02 | P2，不单独阻塞 | 保存已成功但只见通用读取故障 | PUT200→紧接GET503或网络拒绝 | 不能判断已写；可能重复保存，有手动重读绕行，无自动重写证据 | app.js上游提交只在load后message；交互§4 | Open；Contract Fail，UI专项Not Run | QA-02 |
| QA-D03 | P1，新提出的强制同请求契约缺陷 | 概览“核对同请求记录”没有携带显示的成功ID | 当前启动有success→刷新/重登录→概览显示ID→步骤5 | 进入泛列表需手抄ID；同请求动作未兑现。已有不同本地ID时可能串ID为源码风险，尚未执行验证 | overview只显示ID；navigate records只用lastRequestID；index.html步骤5；AC05/13、交互§3 | Open；Contract Fail，UI专项Not Run | QA-03第二轮；不同本地ID/分页原生回归待补 |
| QA-G01 | 必测证据缺口，非已证实产品缺陷 | 完整键盘/依赖恢复/分页失败链尚不能判断 | J03-FULL、J04、J05和矩阵Not Run | 不能用HTTP/VM绿灯拼成UI通过 | journey.mjs多配置/定位动作仍click/fill，未触发more-records/各依赖重新启用 | Open；UI-E2E Not Run | [整改包](remediation-01.md) |

P1依据明确安全和同请求承诺，不是主观“最好优化”。没有Owner已接受P2/P3被静默改级；若后续裁决不同需保留新决策，不改本轮历史。D02保持P2，本身不阻塞。3列窄屏导航而非设计2列、button而非真实链接是偏差观察，当前深链/已测键盘可用，不另造P1；不声称全可访问性认证。

## 4. 证据

- [宿主seq18](evidence/host-seq18.md)：Connector实际执行固定入口退出0。源指纹、当前Go race/HTTP、构建、真实Go两屏尾部明确，中段省略不补造。
- [QA基线](evidence/baseline.log)：同HEAD/分支/56文件摘要；不是只验旧HEAD。
- [QA前端](evidence/frontend-independent.log)：实际 `make frontend-test` 退出0，60/60，无skip。[QA后端](evidence/console-contract-independent.log)：实际 `go test -race -count=1 -v -run '^TestConsole(OverviewAndExactRecords|EvidenceConfigurationRace|DeniedMetadataAndStorage)$' ./internal/relay` 退出0、3/3。Go linker warning不是race失败，临时路径脱敏。
- [新增入口](evidence/independent-boundaries.mjs)/[当前第三轮](evidence/independent-boundaries-attempt-03.log)：node --test退出1，4/4 Fail，复用现有DOM/fetch夹具运行真实app.js，不是浏览器/后端结果。[首轮](evidence/independent-boundaries-attempt-01.log)QA-03误用items，[第二轮](evidence/independent-boundaries-attempt-02.log)纠正为真实data并增加无error/面板可见断言后仍失败；第三轮进一步固定22字符合法ID和修复后精确GET应答，仍在同ID传递断言失败。首轮该项不作为产品证据，所有失败保留。
- [1280px](evidence/host-seq18-console-1280.png)/[390px](evidence/host-seq18-console-390.png)/[摘要](evidence/screenshots.sha256)：宿主概览清密截图，QA目视归档，无完整凭据或对话；截图不代表新增异常/所有写前写后步骤。
- 历次研发失败完整保留。seq7尾部不可恢复、seq11broken pipe结果未知，不补造成功。seq9/12/14/16原生等待/焦点/跨屏问题在seq18同主链通过，不用更底层绿色代关网页问题。

## 5. 未执行和环境限制

QA未重复 `make verify` 的host-only监听部分；沙箱监听/Chromium已有明确限制，按AGENTS采用匹配版本Connector回执，没有扩权、改入口或绕过门禁，也没有静态browser.check替代真服务。

D01–D03新增异常原生旅程、QA-G01完整键盘/恢复/分页为Not Run，须纳入固定入口宿主复验。Contract失败不说成浏览器已复现/已关闭。真实供应商无凭据，Not Run；当前新部署制品、PR/合并/部署后体验无回执，Not Run。旧main部署不验证当前工作树。无新正式操作手册或伪造步骤截图。

## 6. 阶段边界

仅公共路线第1项。第2–4项已授权持续目标、前切片后维护者以独立Issue接续，本轮未实现或创建后续编号。维护者已授权QA后自主测试仓合并/正式测试部署，不改变本节点职责，不包含平台源PR自动合并。当前No-Go不推进发布，普通修复无需新逐阶段批准。

## 7. 环境恢复

QA VM无真实网络/持久存储；Go专项是临时installation和in-process recorder/transport，清理钩子关闭/删除，不触达用户数据。宿主main.go有defer关闭服务/store和临时目录，退出0与正常清理一致；QA没有OS级进程/目录独立清理审计，不冒称已经做过。保留既有忽略的cache/npm/build/test-results，只归档清密图片；无SQLite/凭据/node_modules提交，无数据库造绿、平台修改或Git发布。堆栈本机路径脱敏。

## 8. 质量趋势

相对研发原生多次失败，seq18门禁和两屏主链明显改善，登录异步完成/Tab可达/成功概览复用已有同层通过。QA新增3缺陷主题（4条红灯）及1组覆盖缺口为前序漏测，不能被60VM绿灯掩盖。后端事实/会话watchdog/日志保护证据强；失败态清密/同ID动作/完整键盘退出条件弱。没有确定的原主题产品缺陷复发，不把QA首轮夹具误设算产品缺陷。

公共状态：T001 In Review、M01 Ready for Review（后端专项/当前宿主HTTP已有证据）；T002/T003和M02/M03 Needs Rework，不标Done/Accepted/Delivered。[验收索引](README.md)指向当前报告，不覆盖历史或其他Run。

## 9. Go / No-Go

**No-Go**：未关闭QA-D01/QA-D03 P1及必测UI证据缺口；QA-D02为P2不单独阻塞。不是否定seq18退出0，更不是因不能重复沙箱监听拒绝宿主证据。走development：修实现/补覆盖，不能改PRD/设计放宽。修后新指纹原样 `make verify`、原层异常浏览器与同根/相邻回归，再独立QA；旧绿色不得放行新实现。
