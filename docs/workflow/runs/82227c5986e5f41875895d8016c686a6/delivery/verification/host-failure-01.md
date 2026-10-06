# 首次宿主失败与CLI最小归因

2026-10-06；研发修复。来源：平台传入的 tests 节点 seq=7 Connector 回执，不自行读取GitHub或平台数据库。此文件保存已收到内容的摘要，不冒充原始完整日志。

## 宿主已证事实
- 命令 `make verify`，exit 2；Go go1.25.6；HEAD `85f6d17c07c5fddeb2d6a39a5a067769a11bb4cc`，工作树摘要与 [03](final-source-03.json) 完全相同：`39c1a53b0851952ced57519c97a5e5a146c3c0a53f49158df24a676a035287b9`，80文件。
- 回执开头是health-history隔离重建旧 `9b6994ed80de2bdba79ff267419150dccaac1360` 的失败HTTP断言；旧版GET/HEAD通过、POST等错误方法返回200而不是405。这是原历史红灯重建，不是当前产品方法门禁失败的证明，不能删除或放宽分类。
- 回执尾部当前 `model-relay/internal/relay` 包PASS（207.878s），含真实选择/有限切换/冷却/禁用/已提交SSE、升级及health版本测试；`tests/health-history` 单测PASS（4.473s）。最后整体FAIL，`make[1]: *** [test] Error 1`、`make: *** [verify] Error 2`。
- 中段由平台明确标为 `output omitted`。收到的内容没有展示导致当前Go总测试失败的具体包/测试断言，不能断言唯一失败原因；也不能把此宿主失败归为本机沙箱监听限制。固定入口停在test，build/browser未形成通过证据。

## 最小诊断及修复边界
源码核对发现 `cmd/model-relay/main_test.go` 的真实CLI安装旅程未跟随route-2迁移：仍GET单数upstream、POST旧model字段并期待200、重启后把模型列表解为数组。权威 [架构](../../routing-architecture.md) 和 [Trellis契约](../../../../../../.trellis/spec/backend/routing-contracts.md) 规定410弃用旧管理入口、新candidates写入201、分页envelope。此不一致是已证的测试驱动缺陷，不是需要恢复单上游产品行为的需求。

本机包级诊断 `go test -race -count=1 -v ./cmd/model-relay` exit 1：deployment-contract通过，CLI安装旅程在端口占用检查就因监听权限退出，尚未到旧管理调用；[原失败保留](cli-repair-before-01.log)。对应诊断时03源码。这个结果不能证明宿主的后半段具体失败位置；确认稳定监听限制后不再重跑该完整监听旅程或make verify，不提权。

修复只更新真实CLI测试使用的管理契约。共享 `configureCLIInstallation(t, admin)` 不做旧输入翻译：读取真实空集合、创建元数据上游、以返回的id创建真实候选模型；主测试保留进程启动/端口拒绝/SIGTERM/reset-password/会话撤销/密钥启停和公开访问状态断言，增加health实际schema与重启后上游/候选身份核对。原固定Makefile、历史分类及产品接口不变。

新增免监听 `TestCLIManagementConfigurationContract` 通过真实Go Store/Server/认证与管理handler断言单数410、旧模型400以及同一共享配置步骤成功；无Fixture旧接口自动翻译。仅配置公开IP字面量元数据，避免DNS，未发核验或生成网络请求，不是供应商验证。首轮 [契约回归](cli-management-01.log) exit 0；最终 [静态/race/CLI契约/历史分类/编译/构建检查](cli-repair-validation-01.log) 各exit 0，临时编译路径已脱敏。

## 待宿主辨别
当前版本为 [04](final-source-04.json)。宿主应保存完整固定入口输出（含失败包/测试名与退出码），若回执容量不足，完整日志保存在本Run delivery/verification 新文件中，并回传白名单摘要和实际文件路径，不覆盖本失败。

完整CLI进程、reset/restart、真实HTTP、原生两屏与精确旧UI非空迁移需要同版本make verify。若仍失败，先依据完整断言确定是尚未展示的其他故障、测试驱动还是产品/异步问题，再做最小诊断；不能继续猜测性修复。正式Actions非空升级/隔离失败恢复、独立QA与真实供应商仍未通过，本次免监听绿灯不关闭这些边界。
