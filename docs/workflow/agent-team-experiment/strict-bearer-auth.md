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
