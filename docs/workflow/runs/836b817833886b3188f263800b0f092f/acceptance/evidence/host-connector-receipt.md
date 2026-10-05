# 同轮固定宿主 Connector 回执核验

来源：本节点平台 previous_results，seq=7，node=tests；非 QA 自己执行，也非历史 Issue #8/#9 记录。

- kind：command；exit_code：0；平台摘要：命令已结束 · 退出码 0。
- 交接固定命令：make verify；Makefile 未修改，无 skip/忽略错误分支。顺序为 evidence → health-history → format-check → go mod verify → vet → frontend-test → go test -race → build → npm ci / 实际 Go browser。
- 输出首段 evidence：files=36；go=go1.25.6；head=73681a7df4c5cbe5914056ac13348154d58726e0；working_tree_sha256=0f6ee2d6894067323697f632231ee0d62b098be0261a4da7e975edb6e1efb37b。
- 可见输出为旧健康方法隔离重建，baseline=9b6994ed80de2bdba79ff267419150dccaac1360；expected old failure exit_code=1，拒绝方法实际 200/503 而非 405。这是受信门禁特意验证的旧版失败，不是当前产品失败。
- 回执文本在旧版 HTTP 输出处标记 [output truncated]；没有完整前端、当前 HTTP 或浏览器尾部日志、截图。不得虚构逐项宿主输出或测试数量。根据退出 0 与串行 fail-fast 的固定 Makefile/Go browser Harness 判断完整门禁成功；真实 UI 结论来自该同轮执行及对应脚本断言，不来自 VM 拼接。
- 宿主实际生成的重建日志单独保存在 host-health-history.log。旧路径被脚本改写；QA 保留当前输出后，从 HEAD 恢复旧文件并用交接的 SHA-256 基线确认，而不是把新日志当旧历史。后续本地尝试再次保护并恢复。

证据边界：可信 Connector 总退出状态可作为同轮固定门禁结果；截断导致细粒度诊断/视觉截图不足，明确保留限制，不宣称 QA 亲自完成真实浏览器或提供不存在的原始尾部记录。
