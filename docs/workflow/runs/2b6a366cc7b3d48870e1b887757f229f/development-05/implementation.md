# 研发执行5：QA04 正式返工

2026-10-14，Issue #48，Run `2b6a366cc7b3d48870e1b887757f229f`，revision3；版本 `development05-lifecycle-rework`，HEAD `30b4e13c19e36d51b72ef1186468d6adf39ac822`。

先读取项目 AGENTS.md、冻结身份、development02实施记录、首轮原文、QA04拒绝和检查脚本、归档映射及项目源码摘要算法；未发现 docs/harness-output 下级AGENTS.md。修改前既有全仓离线 check_publish_files 和研发02/QA04 SHA256SUMS 均通过，当前文档与首轮原文一致。未重建工作区，未读取凭据值，仅检查环境变量存在性。

## 实质修订 QA04-D01

仅在新文档末尾追加“停止、恢复与环境释放”章节，包含正式approval安全点stop、重复stop去重实证、同Run同节点resume、前后材料摘要保持、未知结果查询及阻塞、保存后正式生命周期释放与已释放边界。说明前文执行2状态是历史观察，本次不是平台控制已完成的声明。保留前六章原字节、首轮快照及所有QA04原件；不改旧完整文档、产品业务、测试、模板、锁文件、预算、准备步骤或发布规则。不代写最终QA或report。

可调用工具无 apply_patch，终端标准位置亦不存在该命令；沿用既有环境以heredoc追加Markdown及写归档校验脚本文本，没有安装工具。证据只用Markdown、JSON、日志和.py.txt，不产生可编译源码副本。原路径/摘要见 archive-mapping.json；修订diff另存 document.patch.txt。

## QA04-E01 外层证据限制仍须后续处理

保留QA04原件及tests03内层log/state，完整外层仍保留于平台原seq3日志档案。用户正式回执元数据为1174629字节、truncated=false、exit_code=0、timeout_seconds=1800。本次实际调用read_command_output(seq=3, offset=1146880)，返回next_offset=1174629、eof=true、truncated=false；这只证明尾页可取，不代表补齐offset32768至1146880的缺页，也不宣称外层已完整审查或原字节已物化到仓库。现有工具仅分页返回文本，未提供直接落盘归档参数；不以重建/拼造正文代替原件。

此限制不被本次文档修订消除；已保存外层正式元数据和已读区间供固定tests/QA接续。后续QA须按原seq3从offset32768的next_offset继续读至EOF，或由平台正式原字节导出并校验；审查完整前不得关闭QA04-E01或宣称全链路通过。新固定tests也须留存完整外层原回执与内层材料，避免只保存开头和结尾。

## 复核及交接

归档checks记录源码163项逐文件核对及组合摘要、146模板摘要、HEAD/main祖先、旧文档和QA原件保持、文档原文前缀保持/新增章节、无网络发布扫描、冷起点路径存在性与凭据布尔值。旧研发SHA256SUMS仍绑定首轮文档版本，不覆写以冒充修订版；新版本使用本目录独立摘要。

研发未运行make verify或临时替代全套、未手装依赖；交固定tests原make verify，24分钟内层和1800秒外层保持不变。此次有QA04-D01实质修订，不是同版循环求绿。固定tests通过后仍须独立QA复核；正式控制验证、decision、report、固定publisher草稿PR、远端同版核验及保存后释放均未执行。不提交、不push、不建PR、不部署。
