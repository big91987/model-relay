# 研发检查与宿主待验契约

时间：2026-10-09（Asia/Shanghai）；执行：development第4次。结论：实现候选 Ready for Review，完整验证 Blocked by native sandbox；不是完整Passed、独立QA或正式部署通过。

## 版本和真实根因
HEAD/base均为 `cd413b5aa24a372a19c19af999a6ef61b5c41572`。首次环境无node_modules、NODE_PATH unset，Go1.25.6 darwin/arm64、Node26.0.0、npm11.12.1。原Makefile和锁摘要见clean-before.log。原 `make frontend-test` 实测192例191pass1fail，route-cleanup.test.mjs模块加载 `require.resolve('playwright-core/package.json')` 报MODULE_NOT_FOUND，make退出2。它发生于测试执行前，和用户给出的正式第二次prepare根因一致；不是健康方法的预期红灯分类器PASS，不证明被覆盖的首轮Actions日志已恢复。

添加顺序回归后、修复Makefile前，Go定向回归真实失败（exit1）：frontend-test和verify缺安装先行，独立browser原有安装先行已通过。此时首次 `make verify`（含新回归文件但产品入口未改）在review-history处退出2：`historical listener unavailable; not business red`。完整输出和当时代码戳617a4a3660812d88da23371019d838e0948536535a6171c806a21fb4c0275f29保留；未执行的后续门禁不记通过。

## 修复与分层结果
最终待验候选由 tests/evidence实测：142文件，工作树摘要 `8e82aaa123df25a4758a1939fb2d018713831507aee337e18dfa6100c49f619e`（source-after.json）。公共交付回执不在该摘要范围；未提交文件也纳入源码范围。定向检查后仅补充README区分历史验收与本次待验状态，Makefile、测试、规范和锁未再改变；该说明也纳入最终摘要，宿主须验最终候选。Makefile唯一共享安装recipe保持原npm参数；frontend-test/browser各加前置，verify各原步骤未改。Playwright及core仍1.59.1，package.json/package-lock无修改。

| 检查 | 实际结果 | 证据 |
|---|---|---|
| 干净旧frontend入口 | FAIL，192/191/1，MODULE_NOT_FOUND，exit2 | frontend-before.log、clean-before.log |
| 新顺序回归在旧入口 | FAIL，frontend-test/verify红、browser绿，exit1 | dependency-regression-before.log |
| 原生完整make verify | Blocked，历史真实HTTP监听不可用，exit2 | verify-before.log |
| 修复后Go race定向回归 | PASS，三入口顺序及两独立消费者安装失败停止，exit0 | dependency-regression-after.log |
| 实际修复后make frontend-test | Blocked，npm连接EPERM 127.0.0.1:7897，exit2；没有运行node测试 | frontend-after.log |
| format-check/go mod verify/vet/build/diff-check | 均exit0 | static-after.log |
| 宿主完整make verify、独立QA | Not Run，本节点交tests获取 | 尚无回执，不造绿灯 |
| 真实供应商、正式Actions/Deployment/healthz/admin | Not Run，本节点未访问/变更正式安装 | 发布仍待合法Connector |

npm的真实失败来自当前网络连接限制，而不是lockfile缺失或MODULE_NOT_FOUND；不会通过替换代理、全局安装、root/elevation或改驱动来绕过。保留npm完整stdout/stderr并移除私有debug日志地址、机器路径；未读取其私人debug日志。本次环境后态实测node_modules和两个包均仍不存在，见environment-after.log；没有安装成功或frontend绿灯，不能作为干净绿灯证明。定向fixture只验证make失败停止，不证明实际npm/Chromium行为。

Trellis-check：已检查变更范围、适用规范与指引；无产品API/持久化/鉴权/DOM/预算/时间变化，不需要数据流跨层重构。新测试在既有go test ./...门禁内，没有跳过/弱化route-cleanup、历史红分类、浏览器或迁移。必要规则写入 `.trellis/spec/frontend/dependency-preparation.md` 并在索引引用；现有规范有源码依据，bootstrap不适用。完整HTTP/浏览器测试无法在当前原生环境完成。监听限制已实证且环境未变，不反复执行必然在同处失败的完整入口；修复后代码与初次失败的对应关系通过前后源码戳及变更范围保留。

## tests宿主执行要求
1. 使用正式固定项目测试Connector，无额外权限，完整 `make verify` 不改门禁；在独立新clone/worktree开始，确认node_modules不存在、NODE_PATH unset。保留工具链版本、匹配Chromium/system libs前置证据、git HEAD和tests/evidence源码戳。不要在入口前手工npm ci；包源网络访问可用是正常宿主前提，不是产品新增权限。
2. 完整日志指向本Run新非空文件，不覆盖上述红灯。记录实际退出码和源戳、npm ci先于node --test、全部192+状态回归、安装后Playwright/core1.59.1、原历史HTTP分类、race/真实HTTP/CLI/schema1/2→3、budget-clock-r1零软预算/UTC、实际Go UI。不能以局部绿或dry-run代替全量。
3. 若失败先保留最小归因边界、完整日志、源码一致性，再正常返回研发；不是重复无区别全量尝试。QA基于当前候选独立判断，浏览器Blocked和真实供应商未验分别记录。
4. 按用户授权，后续协调者审查独立QA后由Connector新PR、精确已验提交合并和正式Owner deploy=true target=preview。Deployment/environment_url、healthz版本与schema3、真实admin同版闭环前，不宣称已上线。旧数据密码保持，禁止手工改库、重置或替代服务。通用同SHA日志覆盖归Platform维护源。
