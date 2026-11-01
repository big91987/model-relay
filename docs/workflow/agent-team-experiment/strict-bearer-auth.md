# 严格 Bearer 应用密钥认证

日期：2026-11-01

## 缺陷证据

Schema 3 的 `controlledAppID` 原先使用：

```go
credential := strings.TrimPrefix(request.Header.Get("Authorization"), "Bearer ")
```

`TrimPrefix` 在前缀不存在时原样返回输入。因此，一个长度和 `mr_` 前缀均合法的裸密钥仍会进入 `generationIdentity`，使 `GET /v1/models` 接受 `Authorization: mr_...`。生成入口已有 `Bearer mr_` 前置判断，但模型列表直接通过 `appID` 调用该函数，两个公开入口的认证边界不一致。

## 修改

- `controlledAppID` 改用 `strings.CutPrefix`，要求请求头确实以大小写和空格均精确的 `Bearer ` 开头，再执行既有密钥前缀、长度、摘要、启用/撤销和过期校验。
- 新增 handler 级表驱动回归，验证 `/v1/models` 拒绝缺失头、裸密钥、错误 scheme、小写 scheme、前导空白、额外分隔空白、尾随空白、附加内容和空凭据。
- 合法 `Bearer <credential>` 仍成功，并断言仅返回该密钥获授权的 `governed` 模型，证明授权范围未扩大。
- 另验证生成入口拒绝裸密钥且受控上游访问次数为零。

未改变密钥生命周期、摘要查找、轮换/撤销/过期语义或权限数据模型。

## 验证记录

### 针对性测试

命令：

```sh
env GOCACHE="$PWD/.cache/go-build" CGO_ENABLED=1 GOTOOLCHAIN=auto \
  go test ./internal/relay \
  -run 'TestModelsRequireExactBearerAuthorization|TestGenerationRejectsCredentialWithoutBearerScheme' \
  -count=1 -v
```

结果：通过。两个顶层测试及 `/v1/models` 的十个子测试全部 `PASS`，包结果为 `ok model-relay/internal/relay 1.067s`。

首次未带项目本地缓存参数直接运行 `go test` 时，因用户级 Go build cache 为只读且工具链包发现失败，未进入编译；改用上述隔离的仓库缓存后成功。

### 固定入口

命令：

```sh
make verify
```

结果：未通过且未进入项目验证程序。Makefile 固定 `GOTOOLCHAIN=local`，环境本地 Go 为 1.21.6，而 `go.mod` 要求 Go 1.25.0：

```text
go: go.mod requires go >= 1.25.0 (running go 1.21.6; GOTOOLCHAIN=local)
make: *** [Makefile:9: verify] Error 1
```

额外使用 Go 1.25 自动工具链运行完整 `go test ./internal/relay -count=1`，测试可编译并运行，但因当前沙箱禁止监听 `127.0.0.1:0` 而失败；所有报告的失败均为 `socket: operation not permitted` 的真实 HTTP 测试。该结果不是本次认证断言失败，不能视为完整测试通过。

## 未测边界与限制

- 受沙箱监听限制，未完成独立 TCP listener 的端到端回归；新增测试调用真实 Go `ServeHTTP` handler、临时数据库和受控上游 transport，不访问供应商。
- `make verify` 的格式、vet、race、构建、前端和浏览器阶段均因本地工具链版本门槛而未执行。
- 未做真实供应商联调；本任务不需要且未读取任何供应商凭据。
- 当前会话未提供任务要求的 `agent_team_factory.get_task`、`send_message` 或 `handoff` 工具，因而无法调用控制器任务查询/交接接口；本文档保留了实现及验证证据供控制器接收。

## 第三轮固定环境复验

2026-11-01 控制器恢复完整 Git 基线、测试网络及固定工具链后，从原 worktree 继续复验，没有重做实现或改写上述首次失败记录。

工具路径与版本：

```text
/home/moffett/go/pkg/mod/golang.org/toolchain@v0.0.1-go1.25.6.linux-amd64/bin/go
go version go1.25.6 linux/amd64
/home/moffett/work/jiuzhou.zhao/node22/bin/node
v22.14.0
```

随后原样执行：

```sh
make verify
```

实际退出码为 2（外层 `make verify` 报 Error 1）。验证收据为 `test-results/verification/20261101T033439Z-2502115860`。Go、Node、npm 前置检查和 source evidence 均成功，隔离门禁随后在既有历史审查阶段失败：

```text
go run ./tests/review-history
review historical HTTP rejected: baseline unavailable: exit status 128
exit status 1
make: *** [Makefile:13: verify-gates] Error 1
GATE_END exit=2 cancelled=false
FAILED: exit status 2
isolated verification failed: exit status 2
make: *** [Makefile:9: verify] Error 1
```

本轮没有再次出现 Go 版本或监听端口错误，但门禁在 `review-history` 提前终止，未运行到格式、vet、race、Go 测试、构建、前端或浏览器阶段。因此只能继续确认先前针对性认证回归通过，不能声称 `make verify` 或全量项目验证通过。该失败属于验证基线/历史审查前置条件，输出中没有本次 Bearer 修改导致的测试失败证据。

## 第四轮完整历史复验

2026-11-01 维护源报告已执行 `git fetch --unshallow` 并确认历史提交 tree 后，继续原 worktree 复验，保留前三轮记录。

专项命令：

```sh
env GOCACHE="$PWD/.cache/go-build" \
  go test ./internal/relay \
  -run 'TestModelsRequireExactBearerAuthorization|TestGenerationRejectsCredentialWithoutBearerScheme' \
  -count=1 -v
```

结果：通过。两个顶层测试和 `/v1/models` 全部子测试均 `PASS`，包结果为 `ok model-relay/internal/relay 0.777s`。

随后原样执行 `make verify`。验证收据为 `test-results/verification/20261101T034529Z-2130396928`；实际结果仍为 `GATE_END exit=2`、外层 `make verify` Error 1。Go 1.25.6、Node 22.14.0、npm 10.9.2、source evidence 成功，仍在：

```text
go run ./tests/review-history
review historical HTTP rejected: baseline unavailable: exit status 128
```

只读诊断确认 `66f1328ada75ea90e5638fbf2bd58663ac86347e` 的 commit 对象可见、仓库报告 `false`（非 shallow），但门禁所需 archive 仍缺少一个 promisor blob，且 Git 对象目录在原生沙箱中只读：

```text
$ git archive 66f1328ada75ea90e5638fbf2bd58663ac86347e | wc -c
fatal: Unable to create temporary file '.../.git/objects/pack/tmp_pack_XXXXXX': Read-only file system
fatal: index-pack failed
fatal: could not fetch 7a41ad5b61c3b52d2b834d9f09f467211eed301c from promisor remote
0
```

因此“提交/tree 可见”尚不足以供 `git archive` 使用；历史对象仍不完整，而沙箱无法在只读 `.git/objects` 中按需补取。门禁仍未进入格式、vet、race、Go 全量测试、构建、前端或浏览器阶段。本轮专项认证回归通过，但固定入口未通过，不能宣称完整验证成功。

## 第五轮历史内容复验

2026-11-01 维护源确认已在相同原生沙箱条件下成功执行第四轮所需的 `git archive 66f1328...` 后，按指示不重复已绿专项，直接原样运行 `make verify`。

验证收据：`test-results/verification/20261101T035006Z-2838744462`。

第四轮阻塞点 `review-history` 本轮成功，输出了受控历史 HTTP 重建收据及三项预期历史断言，证明提交 `66f1328...` 的 archive 已可用。新的首错发生在下一必跑阶段：

```text
make health-history
go run ./tests/health-history
historical HTTP reproduction rejected: baseline archive unavailable: exit status 128
exit status 1
make[1]: *** [Makefile:25: health-history] Error 1
make: *** [Makefile:14: verify-gates] Error 2
GATE_END exit=2 cancelled=false
```

`health-history` 使用另一个历史基线 `9b6994ed80de2bdba79ff267419150dccaac1360`。只读诊断确认其 commit 对象可见，但 archive 仍缺 promisor blob `d0fc2c137556878ef97ed2e8ea0a3251578374ad`，按需补取因只读 Git 对象目录失败：

```text
fatal: Unable to create temporary file '.../.git/objects/pack/tmp_pack_XXXXXX': Read-only file system
fatal: index-pack failed
fatal: could not fetch d0fc2c137556878ef97ed2e8ea0a3251578374ad from promisor remote
```

因此第 5 轮固定入口仍未通过，也未进入后续格式、vet、race、全量 Go 测试、构建、前端和浏览器阶段。这是另一项历史基线内容准备问题；没有本次 Bearer 回归失败证据，仍不能宣称完整验证成功。

## 第六轮完整验证通过

2026-11-01 维护源完成项目测试引用的三个历史基线内容准备后，按指示不重复专项，原样执行：

```sh
make verify
```

结果：通过，退出码 0。验证收据为 `test-results/verification/20261101T035406Z-1091817144`。关键结果：

- `review-history` 与 `health-history` 均完成受控历史 HTTP 重建检查。
- format check、`go mod verify`、`go vet ./...` 均通过。
- Node 前端测试 286 项全部通过。
- `go test -race -count=1 -v ./... -parallel=4` 通过，其中本次 Bearer 两个顶层回归及所有格式子测试再次通过，实际 HTTP 测试也通过。
- Go build 通过。
- 延迟实验完成 2480 个样本、0 失败。
- 浏览器迁移、旧 UI、路由、控制台和完整多视口旅程全部通过；均使用受控上游，不是真实供应商验证。
- 最终输出为 `GATE_END exit=0`、依赖探针 exit 0，以及 `CLEAN_END gate_exit=0 source_unchanged=true NODE_PATH=unset`。

此前各轮失败记录保留，分别反映当时工具链、网络沙箱和 partial-clone 历史对象准备状态；第六轮是在这些设施条件修复后获得的完整绿色项目固定入口结果。
