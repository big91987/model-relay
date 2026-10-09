# 历史事实、候选与交付责任

当前联合候选6e359667，QA执行23。确认生产缺陷0；没有为关闭Issue制造P0–P3条目。下表分类不是缺陷等级，已有候选不自动变P1。当前范围无需development/design/requirements返工，下一节点为report。

| ID | 等级/分类 | 用户事实或复现/前置 | 影响与定位 | 当前状态 | 回归/Owner |
|---|---|---|---|---|---|
| DS-SAFETY-01 | Needs Evidence，未确认产品缺陷 | 原调用表单缺cap/thinking；当前受控正式bounds→上游默认64/显式32实际观察 | “后台无默认”未复现；真实收费/隐藏推理/思考策略仍未知 | Open（供应商路径） | SafetyActualHTTP＋正式bounds UI；专用测试Pipeline |
| PLAN01 | 材料前置缺口 | 当前ZIP/61项一致，原文实际引用完整 | 不回写旧#33缺材料时已验 | Closed（当前材料） | material-integrity、QA |
| PLAN03–11-INTERNAL | 测试资产缺口，非生产算法bug | 非零/窗口/计量/并发/池数值受控同版完整绿 | 可持续回归，不是“修好生产RPM/TPM” | Closed（受控层） | 当前完整make verify＋独立race；研发维护原断言 |
| JOINT-CANDIDATE | 联合交付证据缺口 | main6929376正式并入输入，HEAD仍8917f942；旧tests9/QA10不足 | 当前158文件已冷全绿，merge仍待publish | Closed（验证）；Ready for Retest（若源码再变） | tests22/QA23；report/publish准确新提交映射 |
| HOST-DEADLINE | 历史门禁不完整/容量风险 | tests5/7原720秒累计期限失败，原样保留；当前完整约653.728秒 | 慢环境可能再次超时，不放宽断言/deadline | Closed（当前门禁）；残余容量风险 | 原失败＋tests22同层绿；失败先最小归因 |
| EXEC17/21 | Git沙箱/npm代理环境 | exec17 Git拒绝，prepare18合法整合；exec21 npm EPERM无完整前端 | 不是业务实现故障；由宿主入口补足而不提权 | Closed（当前交付阻断） | integration-blocked-exec17、frontend-exec21.log及当前回执 |
| PLAN03–10-UI | 原数值网页未验 | 原两屏非零限额/自然窗口/双DA会话与池竞争未执行 | API/测试钟/一般UI不是替代 | Ready for Retest（前置Blocked） | 专用测试Pipeline保留原网页写入步骤 |
| DEEPSEEK-MODELS | 外部历史状态 | 原真实列表无精确flash、有pro，本轮无新账户核验 | 不能推永久故障、不静默rename | Open（Not Run） | 专用Pipeline合法入口精确核验 |
| PLAN12/REAL-LEDGER | 预算/真实上界/费用前置 | H/X历史not_attested，凭据不公开；生成成功及费用无事实 | 不重置24/64/90、不新增生成 | Open（Blocked） | 专用Pipeline先预算/安全独立评审，再有界复测 |

## 对原报告事实的处理

原16次本地认证/授权/零次数拒绝与同ID零attempt为历史子集，完整保留；本轮受控正式HTTP及一般UI同类路径通过，未重跑真实Provider。原网页纠正后503不证明成功恢复，本轮受控J02/J04/J05成功也不关闭真实恢复。原/models及官方HTML仅当时列表/页面取得事实，本轮未去账户端点核验，不以字符串或公开信息证明供应商语义。原普通/SSE成功、真实Tokens/平台与供应商费用对账仍Not Run。原控件检查是事实，“没有端到端安全默认”不是已确认根因；内部已补动态观察，但真实收费安全路径仍待证。原RPM/TPM计划/采证缺口当前内部计划经QA独立评审合格并已资产化，外部执行条件未补齐。

## report/publish/pr应执行的事项

1. 更新本Run旧pr.md，标题【产品测试】补齐受控RPM/TPM与安全调用计量回归，Refs #34；说明这是PR35 main6929376＋PR36测试资产联合补正，同一个Issue34/Run/PR36，不新建产品需求或供应商测试任务，不关闭#33。
2. 引用tests22及QA23当前证据和158文件完整SHA。HEAD8917f942是未提交整合前head，不能把它写成新联合提交；提交/推送由原Connector执行，Actions/部署没有回执前保持Pending。
3. 保留原首失败、tests5/7、exec17/21、旧tests9/QA10与原not_attested事实。历史绿不替新绿，当前绿也不抹旧失败。不为了关闭Issue宣称发现或修复了生产RPM/TPM故障。
4. 专用真实测试责任方继续通过合法配置入口核精确模型、收费输入/输出含隐藏推理上界、max_tokens/思考能力、历史H/X和每条潜在attempt预算。预算图必须H+X+潜在attempt≤24，单次64/90，未知/取消不自动返还；无前置不生成、不读取私有配置、不让协调者代验。
5. 原两屏/独立DA网页、自然分钟、非零RPM/TPM/并发/共享隔离、成功恢复与真账对照须独立评审后实跑；仅API或受控结果不能关闭该层。确认实现缺陷才返development，设计/需求矛盾才返相应节点，附同层首红/修后绿。
