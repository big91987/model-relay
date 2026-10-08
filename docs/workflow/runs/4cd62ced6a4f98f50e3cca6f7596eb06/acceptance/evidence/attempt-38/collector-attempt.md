# QA38 归档器尝试

首次执行collect.mjs真实exit1：读取test-results/console-first-call-1440.png得到ENOENT。这是QA采集清单错误，不是产品或宿主make verify失败；旧native首调用截图不覆盖全部新控制台视口。已删除额外的旧native首调用引用，改为只收当前console三视口、fault及明确PID的continuous产物。未生成/未存在的截图不补造，不删除已有运行或失败证据。首次已复制部分字节，第二次仅对同一源字节复制并核对摘要。

首次误收的旧console-first-call-1280.png副本已移至忽略缓存.cache/qa38/collector-partial-console-first-call-1280.png，保留采集现场但不进入本轮交付证据；源test-results文件未修改。当前host-seq37目录仅包含artifact-index列出的63件同版证据，避免无来源的旧图被后续打包误收。
