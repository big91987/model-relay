# 07 上游配置与密钥脱敏静态审查

## 结论与边界

本报告只对候选源码做静态审查；没有启动产品、运行测试、读取数据库、备份、key file 或真实供应商令牌，也不构成真实供应商联调或产品通过结论。网络目标限制（DNS、拨号、重定向等）归第 18 项，本报告仅在解释 token 的发送边界时提及，不评价其完备性。

结论：未发现供应商 token 经管理响应、错误、check 结果或请求诊断直接回显的 P0 证据。代码建立了 AES-GCM + AAD 的静态保护、管理响应隐藏密文、供应商错误归一化和 execution revision 晚到隔离。确认的主要缺口是安装主加密密钥没有轮换入口（P1 运维恢复能力）；另有解封失败在 connection check 中被误分类、明文生命周期未主动缩短等 P2 问题。删除上游不在当前直接合同中，但 v0.1.0 只承诺修改/禁用，因此不作为本范围缺陷。

## 令牌数据流（确认事实）

1. **输入与校验。** POST/PATCH 将 `token` 解入 Go `string`；限制 8192 字节并拒绝 CR/LF/NUL。创建、URL 改变或当前无凭据时不允许空 token；普通元数据修改可用空值表示“保留原凭据”［`internal/relay/configuration.go:172-180,211-214`］。
2. **AAD 与加密。** AAD 是 `upstream:v2:<id>:<normalized URL>`［`internal/relay/upgrade.go:13-15`］。`seal` 使用调用者提供的 AES key、随机 GCM nonce，并将 nonce 前置于密文；URL 在加密前经过策略规范化［`internal/relay/security.go:28-41`; `internal/relay/configuration.go:188-192,247-254`］。因此交换记录、修改绑定 URL 或密文/AAD 被篡改都会导致认证解密失败。
3. **原子存储。** 创建先在同一事务中取得自增 ID，再以该 ID 和规范化 URL 生成 AAD，随后 UPDATE 写入密文；事务提交前的空 token 不会成为已提交中间状态［`internal/relay/configuration.go:215-255,277-291`］。更新未提供新 token 时沿用已有密文字节；改变 URL 必须同时提供新 token，避免旧 AAD 下的密文被错误复用［`internal/relay/configuration.go:211-214,244-255`］。
4. **读取与管理响应。** 存储层把 BLOB 扫入 `Upstream.Encrypted`，只由其长度派生 `credential_configured`；`Encrypted` 标注 `json:"-"`，GET/list/create/PATCH 返回 `Upstream` 时不会序列化密文或明文［`internal/relay/configuration.go:16-27,41-62,128-137,154-164,282-291`］。概览同样只查询是否存在凭据并返回布尔值［`internal/relay/console.go:97-104,143`］。
5. **使用。** connection check 解封后只把 token 放入出站 `Authorization: Bearer …`，响应体仅在 2xx 时限读 1 MiB 并校验模型列表；持久化/响应只有归一化 error code、HTTP status、时间和执行版本［`internal/relay/configuration.go:631-668,679-695`］。生成路径同样解封并设置 Authorization；解封失败对调用方归一化为 `storage_unavailable`，不会返回底层错误［`internal/relay/routed_proxy.go:69-76,101-106,116-126`］。
6. **升级与启动校验。** schema 1 升级先用旧 AAD `upstream:<URL>` 解封，再用 v2 AAD 重封；迁移事务提交前调用 `validateMaterial` 逐条认证解密［`internal/relay/upgrade.go:156-179,193-218`］。每次打开安装时先用固定 installation AAD 校验 sentinel，再按 schema 版本验证全部非空上游材料，失败即关闭 Store［`internal/relay/store.go:193-229`; `internal/relay/upgrade.go:17-40`］。

## revision 与轮换核对

- 每次成功创建的 `revision=1`、`execution_revision=1`。每次 PATCH 均使资源 `revision+1`；URL 或非空 token 变化使 `execution_revision+1`，仅名称、启停或 `source_kind` 变化不推进执行版本［`internal/relay/configuration.go:230-246,255-267`］。
- token PATCH 是当前供应商凭据轮换方式：新 token 用新 nonce 重封并原子替换；即使提交与原值相同，代码也无法比较明文而仍会推进 execution revision。这是保守失效行为，不暴露旧值。
- check 必须携带预期 execution revision；开始时记录该版本，结束时重新读取当前版本并设置 `applicable`，因此轮换期间完成的旧 check 会保留为历史结果但不会继续标为适用［`internal/relay/configuration.go:588-616,690-695`］。现有直接合同也明确覆盖“check 中轮换后晚到成功不可适用”［`internal/relay/configuration_test.go:11-39`］。
- 路由健康键包含执行版本；凭据轮换后旧健康状态不会污染新执行 epoch，直接合同在 token 修改后验证这一点［`internal/relay/routing_contract_regression_test.go:45-89`］。
- **主加密密钥轮换：未实现。** `openInstallation` 只接受能解开 sentinel 和现有材料的单一 32 字节 key；仓库未见把 sentinel 与全部上游密文原子重封到新 key 的命令或双 key 过渡合同［`internal/relay/store.go:193-229`］。这与供应商 token 的 PATCH 轮换是两个不同层次。

## 已有防护与直接合同

- 管理 API 路由位于管理员边界；Schema 3 的上游写入只记录操作者、动作与 upstream ID，不记录 token［`internal/relay/configuration.go:259-275`］。
- 所有管理失败均返回稳定 code；所审路径没有把 `seal`/`unseal`/SQL/HTTP 的原始 error 文本放进响应［`internal/relay/configuration.go:180-291,583-695`］。
- check 不持久化供应商响应体或模型 ID，只持久化状态、归一化 code 和数值 status［`internal/relay/configuration.go:644-668,685-695`］。
- 请求诊断脱敏合同将应用密钥、两个供应商 token、上游 URL 和已流出内容逐项断言为不存在［`internal/relay/routing_http_test.go:246-255`］。这是源码中的测试意图，本任务未执行该测试。
- 升级合同验证旧 AAD 不再能解开迁移后材料，且新 AAD 能恢复原受控 token［`internal/relay/upgrade_test.go:75-86`］。connection check 合同覆盖 10 秒预算、协议错误、401/429/503 与网络失败的归一化结果［`internal/relay/qa_connection_contract_test.go:13-83`］；本任务同样未执行。
- 路线/发布合同明确要求“凭据加密保存、查询不回显”和“不记录密钥或敏感上游响应”［`docs/releases/v0.1.0.md:23-27`］，代码结构与此一致。

## 发现与优先级

### P0

无。静态审查未找到 token 明文/密文进入管理 JSON、审计、请求诊断、持久化错误或供应商错误透传的路径。该结论不等于动态证明不存在旁路。

### P1 — 安装主加密密钥无可用轮换合同（确认缺口）

**证据：** Store 只加载一个 key，并以同一 key 校验 installation sentinel 和所有 upstream material；现有升级只在 schema 1→2 时改变 AAD，仍使用原 key［`internal/relay/store.go:197-229`; `internal/relay/upgrade.go:163-169`］。

**影响：** key file 疑似泄露时，操作者无法通过受支持的原子命令更换数据加密密钥；仅替换文件会使安装无法打开。恢复只能依赖另行重建/重新录入所有供应商凭据，增加停机、误操作和旧密文长期可解风险。

**建议：** 后续提供离线、持锁、先备份的 key rotation：验证旧 key 与全部材料，生成/接收新 key，事务内以新 key 重封 sentinel 和每条上游材料，完整复验后提交，并以原子方式安装权限为 0600 的新 key file；失败必须保留旧安装可启动。不要在响应或日志输出 key/token。此项是建议，未实现、未执行。

### P2 — check 将材料解封失败误报为连接失败（确认缺陷）

**证据：** check 先把结果初始化为 `upstream_connection_failed`；`unseal` 失败时跳过 HTTP，但没有改成 storage/material 错误或触发 degraded，最终持久化并返回该连接错误［`internal/relay/configuration.go:633-671,679-695`］。生成路径对相同失败则记录并返回 `storage_unavailable`，并调用 `storageFailure`［`internal/relay/routed_proxy.go:75,101-106`］。

**影响：** 启动后发生内存/存储材料损坏时，管理员会被引导排查网络或供应商；健康诊断与生成路径语义不一致。启动时 `validateMaterial` 已降低正常启动即遇到此状态的概率，但不能覆盖启动后的损坏。

**建议：** 将解封失败映射为不含细节的内部材料/存储错误并置 degraded；不得返回 AEAD 错误文本。增加不读取真实 token 的隔离合同，验证无 HTTP 访问、无明文响应且状态分类一致。

### P2 — 明文 token 生命周期未主动缩短（加固假设）

**证据：** 输入与 `unseal` 返回 Go `string`；生成/check 将其拼成 Authorization 字符串，之后依赖 GC，没有显式清理能力［`internal/relay/security.go:43-56`; `internal/relay/configuration.go:635-640`; `internal/relay/routed_proxy.go:75,125`］。这不是 API/持久化泄露证据，且 Go string 通常无法可靠清零。

**影响：** 进程内存转储、调试器或同进程内存破坏场景下，明文可能比一次系统调用存活更久。当前威胁模型文档未证明要求抗内存取证，因此列为加固假设而非已证实产品违约。

**建议：** 避免复制或格式化 token、禁止 header/请求对象调试日志，并在安全设计中明确内存转储边界；若未来引入可清零缓冲区，需评估 `net/http` Header 自身仍会复制字符串，不能声称完全清零。

## 未执行项与剩余风险

- 未运行 Go 测试、`make verify`、服务、浏览器、HTTP 请求或供应商联调；上文“直接合同”仅表示测试源码存在及其断言内容。
- 未读取或查询任何 SQLite 数据、备份、运行日志、key file、环境变量或真实 token；无法确认部署文件权限、日志采集器、反向代理、崩溃转储或备份介质的实际状态。
- 未动态验证 JSON 序列化、SQL 事务回滚、nonce 随机源、竞态、重启、升级失败恢复或损坏材料路径。
- 未评价第 18 项负责的网络目标限制，也未把受控测试上游当作真实供应商证据。
- 本审查未发现直接暴露不代表不存在旁路；尤其应由后续独立 QA 用合成 secret 对管理 GET/list/create/PATCH、check 各失败分支、请求诊断、审计和日志做负向扫描，同时保持不接触真实凭据。
