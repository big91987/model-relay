# QA23 采证与失败保留

第一次audit.mjs在归档material-integrity时EISDIR退出1：mkdir路径包含未规范化的文件名/..，Node先建成文件名目录。尚未归档宿主日志、未改变业务源码/数据/Git。最小修正使用dirname，只rmdir本轮刚创建的空目录后复跑exit0。原工具回执保留在本执行日志；这是QA采证脚本故障，非生产故障或门禁失败。

随后实际独立Go8组race exit0（21.991秒）、Node53项exit0。完整make verify的执行者是tests22宿主，不是本QA。read_command_output只读了seq22首32768字节，未用缺失或截断部分推导成功；落盘完整verify.log按state长度/hash核验，Connector声明959549字节未截断，97字节前缀重构长度一致。未声称独立逐字节取回平台整段。

保存全部本轮browser PNG/JSON/JSONL，无修改/裁剪图像；文本以绝对本机路径、Bearer、sk-模式安全检查后归档。未归档私有Provider、密码、session或cache。两张390截图由QA实际抽看，剩余截图按原PNG采证/几何/keyboard断言及数值回执审查，不伪称逐张人工视觉验收。

旧tests5/7和exec17/21失败及旧tests9/QA10证据未删/未覆盖。当前报告/证据路径由index指向QA23，旧报告保留为历史。

finalize首次退出1：安全检查发现独立Go日志链接器warning含临时绝对路径（malformed LC_DYSYMTAB）。实际8组race仍exit0；未忽略安全检查。最小修正保留raw在忽略缓存，仅将该warning的一处临时路径替换为<TMP_LINKER_PATH>，记录原/公开字节数与SHA于local-log-transforms.json；测试/数值行不变，宿主完整日志原样。数值与PNG核验已在首次finalize完成，安全修正后再完整归档。
