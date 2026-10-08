# scope-refresh-attribution-r1 实施候选

## 来源与范围

2026-10-08，Run `4cd62ced6a4f98f50e3cca6f7596eb06` development执行31，延续同T001/M01、architecture-r1、冻结design-v0.2.0。材料manifest为 `.workflow-input/4cd62ced6a4f98f50e3cca6f7596eb06/5f51d35fdc15ed7214749f9413879b4238097f1d6b6bbe2cc0d508d8e103cef7/files/manifest.json`；archive SHA256 `5f51d35fdc15ed7214749f9413879b4238097f1d6b6bbe2cc0d508d8e103cef7`。实际引用同root的docs/prd.md（成员移除/降权即时收回诊断）、docs/interactions.md、docs/acceptance.md、docs/decisions.md，摘要核对见material-source.log。材料脚本不执行，冻结内容不改，不重复批准原型或恢复旧Run。

恢复现有delivery、architecture/implementation-contract、acceptance-matrix与独立QA22 report/remediation；[宿主事实及归因修正](host-seq30-summary.md)是本轮切入点。无需求/设计歧义，不返回requirements/design。

## 最小行为差距与变更

行为所有者是实际driver的刷新等待和失败归因。`tests/browser/journeys.mjs` removed-member refresh最多5秒等待期望零租户选项，再执行原严格count断言；细分十个成员阶段，GET绑定activePage与numeric expected/actual status、白名单role/count、自己的源码line/column（不打印raw stack/异常值/host路径）。普通刷新不加scope期望，无任意sleep、业务写重试、API预填或取消原断言。`journeys.test.mjs`实际helper延迟回归先红后绿及脱敏/阶段回归仍不证明真实Chromium机制。

`internal/relay/console_contract_test.go`已有PeopleScope in-process/actual HTTP两个入口共用生产handler/SQLite场景，补developer→finance、现有session角色readback、诊断清空、角色恢复及正式Key编辑重新分配、成员退出后的空tenant集合；重新分配确保原DELETE CASCADE断言仍有待删除记录。生产Go/UI/协议/计费/DB/权限不改，不调记录以过测，不加产品依赖。seq30原native唯一根因仍待新安全位置证据确认。

## 实测与未测

- refresh-red.log保留修补前Node退出1，旧helper提前返回；refresh-green.log退出0（12 tests），仅helper层回归。日志本机路径已脱敏，原断言和退出事实保留；最终修改后全部Node回归又由checks.log复验。
- checks-first.log首次format/module/vet/frontend通过，定向Go因命令包装未export GOCACHE而走只读用户cache导致setup失败exit1。修检查调用器显式export项目既有.cache/go-build，不改权限/测试代码；保留失败，不归入产品失败或稳定监听失败。
- checks.log最终全部exit0：format-check、go mod verify、vet、frontend-test（164 tests）、三项顶层Go race（PeopleScopeInProcess、PauseUsageInProcess、AccountingInProcess）、ContinuousUpstreamResponseContract及known/unknown、build、browser测试编译、journeys语法、diff检查。链接器LC_DYSYMTAB warning原样保留，临时路径脱敏。
- 当前版本/完整工作树摘要见source-evidence.json；局部race/编译不升级为当前actual HTTP/native通过。PeopleScope ActualHTTP扩展与J01–J07三视口仍待宿主。稳定原生监听环境未变，按前序限制证据不重复必然失败的完整make verify；seq30产品/driver失败不因此变成环境失败。

## Trellis-check 与交付审查

实际get_context --mode packages为Single repo backend/frontend，读取frontend索引、DOM/lifecycle/verification/workspace，backend runtime/console/routing/multitenant及shared thinking guides；既有规范有源码契约，不是模板，无需bootstrap。保留prepare initialized:false，不声称首次初始化成功。update-spec在workspace-contracts记录刷新等待、归因、状态/秘密边界及必需回归；check审查范围只涉及测试、规范和本Run状态，新增回归仍由原make verify的go test/frontend-test覆盖，Makefile不改。无产品API/存储/授权变更，不做无关重构。

T001 In Review、M01 Ready for Review，整体Blocked by Validation；QA22-01/02未关闭、原独立No-Go有效，AC01–AC18/H01–H08/八类商用/J01–J07五层状态不减少或勾完成。供应商/外部身份/外部支付Not Run；无发布/合并/Issue评论/用户DB操作。Codex/gpt-6.1-sol是指定平台配置，工具无实际模型证明，不伪称核对成功。

## 下游契约

最终版本：HEAD `98511771871cf0951ecef716bbb55deb13e18da5`，124文件、Go go1.25.6，工作树SHA256 `de6d4d88d809bedacf06d00693c51cb4fb7159b2d2c986c3bf80a025a270b978`。公共状态更新后生成完整指纹；产品/测试源码与checks.log实际检查时一致。最终diff检查exit0，attempt31证据扫描无凭据模式/绝对本机路径；archive及四份引用文档摘要与manifest一致。无提交/切分支/推送/部署。

tests对source-evidence.json同指纹运行原完整make verify，保留before/after指纹、完整非截断日志/exit、原层失败及continuous三视口JSON和before-rotation/key-states/before-correction/after-correction/final脱敏截图至本Run验证目录。PeopleScope ActualHTTP扩展必须运行；1280/1440/390各需J01–J07七项actual passed，原其他Go/Node/health-history/迁移/HTTP/SSE/取消/未知签发/keyboard门槛不减。旧新console三视口绿或VM green不能作连续旅程绿。

若J05仍失败，用failure_boundary（当前journeys版本line/column）、十个成员stage和read/membership/usage/call事实确认首次失败；位置缺失记录unavailable，不输出raw异常/秘密。失败返development；完整exit0才交独立QA按同材料/工作树逐AC/H及商业旅程五层状态复验，由QA决定是否撤销No-Go。本轮修已证driver等待缺口，**不宣称seq30唯一native根因修复**。
