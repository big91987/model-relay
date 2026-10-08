# QA采集尝试保留

1. 首次只读 `go run ./tests/evidence` 未设置GOCACHE，访问用户Go cache被沙箱拒绝（exit1）；改用项目 `.cache/go-build` / GOTOOLCHAIN=local 后成功，未扩大权限、未监听。错误未作产品缺陷。
2. collect.mjs首轮exit1：错误假设fault JSON有总体result=passed，实际该回执有phase=completed及case.result；首轮collect.log为空，原错误保留在执行47原生命令归档。只修QA采集脚本为五case逐项校验，历史红须expected_failure_reproduced，其余passed，不删任何断言或伪造字段。
3. 修正后collect-final.log exit0：材料19/19、源戳129一致、63PNG与三个review/J/J活动/fault，75件按原bytes/hash保存。采集成功不等于AC01通过；实际观察Fail另在capture-contract-review.json及主报告。

所有持久引用均仓库相对路径；没有保存用户cache路径、凭据或原始秘密。未删除上轮失败证据，重复采集仅同一宿主产物字节拷贝。
