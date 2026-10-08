# 宿主执行37独立核对

2026-10-08；qa执行38。固定入口make verify由tests Connector执行，不是QA原生执行。平台回执exit0、101748 bytes、truncated=false；read_command_output(seq=37)按offset 0→32768→65536→98304→101748/eof=true读回，另offset44000定位中段回归。完整原始脱敏日志保留在平台，本文不是重建完整日志，也不以会话首尾摘要代替归档回执。工具展示可能省略长正文，未将展示省略称为归档截断。

## 版本与结果

- 日志首行go run ./tests/evidence：HEAD98511771871cf0951ecef716bbb55deb13e18da5，Go1.25.6，127文件，工作树3ee2122eadac33a67342523a082f791d8a81154cc86b7321d53de856866a15ef；QA source-before.json独立重复一致。
- 历史health-history隔离重建旧版的exit1是明确预期红灯，随后历史评估及当前HealthMethodsHTTP/Handler通过，不解释为本候选失败或旧开发时成功记录。链接器LC_DYSYMTAB警告不是隐藏的测试skip。
- format-check、go mod verify、vet、frontend-test、go test -race -count=1 -v ./...、build、browser依次完成；Node169项0失败0skip。当前internal/relay包412.191s通过，tests/browser包9.688s通过；其他包按归档结果，不以耗时作为正确性证据。
- 新ResourceMatrixActualHTTP两scope×五维均通过；SharedResourcesActualHTTP记录multi_key及old_new_generation各8争2、accepted2/denied6/other_tenant1。InProcess同名资产另通过，不能把InProcess写成HTTP。
- 原健康/鉴权/隔离/治理/暂停/费用/崩溃/备份恢复/迁移门槛保留。正式CLI schema1/2→3与旧schema1→2兼容边界分别输出；真实Go浏览器、旧native恢复/分页/路由及键盘门槛通过。
- console-faults在1280/1440/390均记录：隔离重建旧editor预期失败；签发POST201、轮换POST200已提交后丢响应，GET503→404→200；mutations1/no secret replay/轮换叠加409；普通和SSE各proxy_posts1/admissions1/attempts1/actual_upstream_dispatches1/Tokens3+4。
- continuous-console三视口各J01/J02/J03/J04/J05/J07/J06通过，共21检查点，不把标签当全部异常组合证据。console-real-go-ui三视口完成十页导航、关键操作及秘密清理。

## 产物来源与限制

artifact-index.json记录63件源/副本相对路径、bytes/SHA256及源mtime；三continuous PID明确为5283/5310/5330，checkpoint与当前回执一致，fault JSON phase=completed。所有纳入索引的副本与源字节一致。截图的无秘密条件来自runner实际断言，QA另外查看了1280签发恢复、390轮换读503、1440追加更正、1280Key列表及批准原型Key截图；不声称逐像素OCR或全部截图人工检查。

浏览器JSON自身没有签名版戳；关联依赖Connector首行源戳、匹配checkout、当前runner及10:20–10:21产物时间窗，不夸大为制品加密证明。当前runtime标签console-controlled-verification / console-continuous-controlled-ui、schema3是隔离测试服务，不是正式发布版本。63件证据集中Key与账单页面，不能证明十页均逐页对照批准原型；具体缺口见visual-review-38.md。

QA另实际运行Node169、七项顶层Go race（含resource两个入口中的InProcess、治理/成员/暂停/精确期限/精度）及format/modules/vet/build，日志qa-node.log、qa-go.log、qa-static-final.log分别对应，不冒充QA再次执行make verify/真实HTTP/Chromium。qa-static.log首次组合没有errexit，不作为最终组合门禁；带set -e的qa-static-final.log才是最终静态检查。原生监听Blocked仍用原QA22证据，不重复必败命令、不提权、不用静态browser.check代替。

真实供应商/外部身份支付、指定原生模型执行证明、正式PR/Actions/Deployment/固定URL均不由本日志证明。H02/H04/H08平台负向和恢复未执行，不能用产品门禁勾选。
