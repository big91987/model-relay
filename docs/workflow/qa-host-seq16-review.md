# seq 16 宿主验证回执独立复核

日期：2026-10-05。原始完整回执为平台 previous_results seq 16，kind=command，exit_code=0。本文是独立 QA 摘录/审核，不是另一份命令执行，也不以研发 delivery 的状态作为通过依据。

## 代码对应

```json
{"files":32,"go":"go1.25.6","head":"fdcb16cb3426d403327da727ae2a8f95c001f1af","working_tree_sha256":"fbc07625a4d9819bba5aa809de5dfea636450cf42489d7d7c72ab66ccba06efe"}
```

独立 QA 使用项目 tests/evidence 再次计算相同值，见 qa-sandbox-recheck.log。摘要涵盖产品、网页、测试、构建锁、README 与验收/设计基线，排除交付/QA 回执避免自引用。HEAD 本身仍是初始提交；工作树实现尚未提交。

## 原回执结果

- format-check、go mod verify（all modules verified）、vet 成功。
- frontend-test 4 项成功，fail/cancelled/skipped/todo 均为 0。
- go test -race -count=1 -v ./...：CLI 包成功（4.355s），relay 包成功（98.380s）。所有原用例及子用例继续通过；新增 TestHTTPConfigurationChangesAndReenable 通过（2.57s）。CrashProcessHelper 是辅助入口，不重复算作独立验收。
- go build -trimpath -o bin/model-relay ./cmd/model-relay 成功。
- npm ci 后实际执行 go run ./tests/browser；不是静态原型。原输出：

```text
PASS actual Go service browser journey 1280px (controlled upstream, not supplier validation)
PASS actual Go service browser journey 390px (controlled upstream, not supplier validation)
```

独立审核了 lifecycle_test.go 的新模型/地址/token、拒绝不变、禁用无消耗和重新启用断言，以及 journey.mjs/main.go 的网页映射编辑、多映射正常关停重开、失效会话/重新登录、管理/DOM 双核验、真实敏感失败计数、记录 DOM 和捕获服务日志排除断言。测试控制动作仅用于运维夹具/日志审计，产品对象由网页建立；没有数据库直接写入补状态。

## 证据边界

完整工程门禁成功，QA-GAP-01/02 关闭详见 qa.md 第 9 节。历史 seq 13 及失败回执保留，不作为当前摘要的新执行证明。没有截图/trace 可附，不编造。链接器 warning 和故意注入 storage_write_failed 不导致断言失败。执行者是宿主 Connector；Agent QA 自查仅非监听资产。真实供应商联调 Not Run，不以受控服务证明供应商兼容或可用性。
