# QA采证失败和恢复记录

以下是qa执行10真实命令过程，不是产品失败或完整门禁结果；未删除原失败，完整宿主门禁来源始终是tests9。

1. QA原生 `go run ./tests/evidence` 首次因默认用户Go缓存不在沙箱可写根而权限拒绝。改用工作区GOCACHE和local工具链后source-evidence exit0，154文件摘要与宿主一致。不请求提权，不重复已知监听失败的完整门禁。
2. collect第一次exit1：归档扫描在独立Go日志拒绝个人临时linker对象绝对路径。此前已复制宿主原state/log及browser文件，没有宣称采证完成。原始独立日志在忽略缓存保留；仅将Darwin linker warning里的临时对象路径替换为`<temporary-object>`，不删除warning、不改变测试结果，raw/archived字节摘要在local-log-transforms.json。
3. collect第二次exit1：检查文字中“auth-header payloads”的早期写法被扫描器当做头载荷。只改说明文字，不放宽扫描规则或修改runner/state/test断言。
4. collect第三次exit0：61材料+ZIP、宿主日志SHA/字节数、前后source、依赖/退出/清理全部核对，161条文件摘要记录加archive-index自身，共162个生成文件。日志转换不让未经同意的本机路径进入公共产物。
5. QA独立Go race exit0、8组/29.082秒，出现Darwin LC_DYSYMTAB linker warning但测试完成，无race失败。初次Node smoke把普通selection-trace模块也列入node --test，共4个报告项；正式focused只跑两项测试文件，3个断言全部通过，不把模块加载当第四个功能验收。
6. 报告及当前索引写入后再次collect整合exit0，额外生成handoff-artifacts，交接索引分别绑定10项报告/索引和161项原始归档证据。独立逐项核验这些文件摘要与研发历史索引，均一致；最终source仍154文件2dfe14a3，git diff --check exit0。此次再归档不是重跑make verify，也不增加供应商请求。

宿主原verify.log已经无个人绝对路径/秘密，逐字节保留。PNG由运行脚本断言秘密字段/存储清空，QA抽查两张最终屏，不宣称通用secret扫描能够验证所有图像。
