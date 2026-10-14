# QA执行7追加记录：完整外层原档能力阻塞

2026-10-14；Run2b6a366cc7b3d48870e1b887757f229f/revision3；候选development05-lifecycle-rework/HEAD30b4e13c19e36d51b72ef1186468d6adf39ac822未变。原qa-07/qa.md、20项文件清单/摘要和QA04拒绝不覆盖；此前revise handoff未受理（用户输入待送达），没有发生正式研发交接。

## 新增旁证与独立观察

维护者给出的正式连续分页/EOF、外层字节数与摘要只作为旁证保存在maintainer-corroboration.json；seq3摘要754f5242f9a567dd52b2f40b10667b34c18819099b6094d28940e48f661d310e、seq6摘要ec2540257de3d6a6c0717f165a045ceee8712b30380bb1cba4c56691e4bf6caf。维护者表示原件保留在平台，未复制进工作区。本QA不能将维护者阅读写成自己的阅读，不能凭这些摘要关闭QA04-E01。

本QA此前亲自调用seq6的offset0和1146880，分别得到next_offset32768/eof=false及1177433/eof=true。收到旁证后另实际调用seq3 offset32768，next_offset65536/eof=false/truncated=false，看到源码/历史材料路径清单。**工具响应呈现包含省略标记，没有可取的原档文件引用、下载句柄或直接写入工作区参数。** 这些调用证实原生分页可用，但不等于每页原字节均完整展示/归档；不把返回next_offset跨度当作完整内容独立审阅。不宣称独立读完seq3/seq6，亦不宣称平台原档truncated或丢失。原报告的完整审阅缺口仍成立；新增页也不足以完整关闭它。

已完整读取/摘要核验的两个内层原件及限定独立测试结论不变，文档QA04-D01修复确认不变。这是外层证据接口能力阻塞，不是新产品缺陷，不需研发改产品或重跑固定测试。

## 精确能力缺口及解除条件

当前暴露的read_command_output只有seq/offset输入，返回呈现文本/分页元数据，没有直接按原字节输出为受授权本地文件或完整原档文件引用的功能。当前终端可以保存自己读取的文件，但没有这两份平台外层原档的受授权文件句柄。已有仓库SDK公开只读路径用无凭据HTTP请求得到401；本QA没有获取或读取凭据值，也不要求扩大Agent权限。平台已保留原档的旁证不自动赋予终端读取原档的认证路径。

可靠解除方式：平台通过已有正式文件API/原字节导出能力，将seq3和seq6完整原始脱敏外层档案提供为本QA可访问的精确文件（保留原Run/seq与来源映射），或在当前原生工具提供不经呈现省略的完整字节文件引用/授权原档读取。由QA独立核对1174629/1177433字节、给定摘要、退出/timeout及内层对应后完整复核。不能从显示的省略文本补造外层副本、用内层替换外层、交凭据值、重跑求绿或绕过固定发布规则。导出日志是证据保存，不是手动Git发布。

依最新用户指示使用wait_for_input(kind=blocked)保留在当前QA节点，等待原有权限范围的可靠原档读取/导出能力；不重启Run、不新增节点、不把blocked当通过，也不提前进入approval。

## files（追加具体文件）

- `docs/workflow/runs/2b6a366cc7b3d48870e1b887757f229f/qa-07/maintainer-corroboration.json`
- `docs/workflow/runs/2b6a366cc7b3d48870e1b887757f229f/qa-07/capability-blocker.md`
- `docs/workflow/runs/2b6a366cc7b3d48870e1b887757f229f/qa-07/followup-files.json`
