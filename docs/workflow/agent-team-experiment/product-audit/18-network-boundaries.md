# 外部端点与网络目标边界静态审查

## 结论

审查对象为当前 worktree 的 `internal/relay/security.go`、`internal/relay/configuration.go`、代理请求路径、共享 HTTP client/transport 及启动参数直接合同。结论：默认公网供应商路径具备较完整的 SSRF 防线；配置时与连接时均解析 DNS，拒绝任一非公网地址，连接时固定拨号到已校验 IP，且禁用连接复用和自动重定向。HTTPS 请求的 URL host 仍由 `net/http` 用作 HTTP Host 与 TLS ServerName，未因固定 IP 拨号而改成 IP。

未发现 P0。记录 2 项 P1（私网例外的明文 HTTP 合同、关键网络安全性质缺少充分回归覆盖）和 2 项 P2（任意公网端口、TLS 策略依赖标准库默认值）。这些项目均不等同于已证明可利用漏洞；威胁主体主要是已获管理面写权限的操作者，或能控制/污染端点 DNS 的对手。

## 范围、方法与证据状态

- **事实**：以下代码行引用来自本 worktree 的静态读取；未读取供应商密钥存储实现，也不评价密钥存储。
- **未执行**：依任务边界，未联网、未探测任何真实目标、未启动服务、未发起 HTTP/DNS/TLS 请求、未运行测试或 `make verify`，也未修改实现。
- **假设**：按 `go.mod` 的 `go 1.25.0` 合同，采用该版本 Go `net/http`/`crypto/tls` 的通常语义：请求 URL hostname 决定 Host/SNI 和证书主机名校验；未设置 `TLSClientConfig` 时使用系统根证书与标准库默认 TLS 策略。此项本任务未动态证明。
- **任务工具回执**：首次执行时任务工具未暴露，因而当次无法取得控制器消息或提交 handoff；这是此前退出成功但任务失败的历史事实。本次正式接续（attempt 2）已成功调用 `agent_team_factory.get_task`，确认三项依赖均为 `merged`，并收到“仅静态核对现有报告、不运行产品测试、通过真实 handoff 交付”的控制器消息。本报告仍是唯一产物。

## 已有防护（事实）

### URL 与协议

- `Policy.URL` 设置 5 秒解析/解析器上下文，拒绝空 hostname、userinfo、query、fragment、opaque URL、反斜杠和 CR/LF；拒绝 RawPath、NUL、`..` 与双斜杠路径，并清理基础路径（`internal/relay/security.go:138-161`）。这阻断凭据嵌入、URL 解析差异及后续 `+ "/models"` / `+ "/chat/completions"` 拼接的常见越界。
- 默认只接受 HTTPS；HTTP 仅可用于精确匹配的测试 authority 或运维显式配置的私网 authority（`internal/relay/security.go:145-147`）。管理 API 在创建和更新上游时调用该校验，失败不落库（`internal/relay/configuration.go:172-192`）。
- 显式端口必须为十进制 1..65535（`internal/relay/security.go:151-156`）。未显式限制到 443，见 P2-01。

### IP 分类、DNS 重绑定与多地址

- 字面 IP 直接解析；域名经注入 resolver 或 `net.DefaultResolver` 解析（`internal/relay/security.go:110-119`）。
- 地址先 `Unmap`，再拒绝非 global-unicast、RFC1918/ULA、环回、链路本地，以及 CGNAT、文档、基准、保留、IPv4/IPv6 转换/隧道等显式前缀（`internal/relay/security.go:79-109`）。因此 IPv4、IPv6 和 IPv4-mapped IPv6 均进入同一分类边界。
- `validateIPs` 遍历解析结果并在**任一**地址不合规时整体拒绝；没有只检查首地址的 mixed-answer 绕过（`internal/relay/security.go:120-136`）。私网 allowlist 仅放行 private，仍拒绝 loopback 与 link-local；测试 authority 才可放行 loopback（`internal/relay/security.go:125-133`）。
- 配置时做一次 DNS 校验（`internal/relay/security.go:158-160`）；真正连接时重新解析、重新校验全部地址，并只向本次已校验的首地址拨号，而不是再按 hostname 解析（`internal/relay/security.go:163-178`）。这消除了“校验后由系统拨号再次解析”的经典 DNS 重绑定窗口。
- transport 禁用 keep-alive，所以每次请求均重新进入上述 Dial policy；同时不读环境代理（`Proxy: nil`），避免由代理绕开目标 IP policy（`internal/relay/server.go:74-75`）。代价是可用性和握手开销，而非边界缺失。

### 重定向、Host、SNI 与直接请求合同

- client 的 `CheckRedirect` 始终返回 `http.ErrUseLastResponse`，3xx 不被跟随（`internal/relay/server.go:75`）；生成路径把 3xx 归类为 `upstream_redirect`（`internal/relay/routed_proxy.go:159-163`）。因此 bearer credential 不会随重定向发送到第二目标。
- 探测固定为 `GET <base>/models` 并只添加 bearer header（`internal/relay/configuration.go:631-640`）；代理固定为 `POST <base>/chat/completions`，设置 JSON、bearer、请求 ID 与 Accept（`internal/relay/routed_proxy.go:116-141`）。两者复用同一受约束 client。
- `DialContext` 只替换底层连接目的 IP；`http.Request.URL.Host` 未被改写。依 Go 标准库通常语义，HTTP Host、TLS SNI 与证书 hostname 校验仍一致指向已配置 hostname，而非选中的 IP。代码没有设置 `InsecureSkipVerify`、自定义 `ServerName` 或自定义根（`internal/relay/server.go:74-75`）。这是基于标准库合同的静态判断，未做 TLS 抓包或握手测试。
- 共享 transport 设置 5 秒拨号和 TLS handshake timeout、response-header timeout；探测另有 10 秒总上下文（`internal/relay/security.go:163-178`、`internal/relay/server.go:74`、`internal/relay/configuration.go:631-632`）。

## 发现与分级

### P1-01：私网 allowlist 同时授权明文 HTTP

**事实**：`--private-upstream-authorities` 接收精确 `host:port` 字符串并交给 Policy（`cmd/model-relay/main.go:34,85-93,118`）；只要 authority 精确匹配，`Policy.URL` 允许 `http://`，`validateIPs` 允许 private 地址（`internal/relay/security.go:68-75,125-146`）。allowlist 参数只验证可被 `net.SplitHostPort` 分割、字段非空和无首尾空白，没有按 scheme 分离授权。

**影响**：显式启用私网上游后，供应商 bearer credential、请求提示和响应可在本地网络以明文传输；同网段观察者或中间人可读取/篡改流量。此风险需要运营者先授予精确 authority，默认部署不受影响，但凭据与模型数据的影响高。

**建议**：把“可访问私网”与“允许明文 HTTP”拆成两个显式开关；私网 allowlist 默认仍要求 HTTPS。若确需 HTTP，要求单独逐 authority 的高摩擦配置并在 UI/启动日志标示无 TLS。补充验证：私网 allowlist + HTTPS 成功、相同 authority + HTTP 默认失败、单独明文授权才成功。

### P1-02：关键边界的回归测试覆盖不足

**事实**：现有单元测试覆盖 userinfo/query/fragment/path、部分 IPv4 私网、mapped loopback、CGNAT、一次 DNS 重绑定和 mixed public/private answer（`internal/relay/relay_test.go:301-323`），也覆盖不跟随 307（`internal/relay/relay_test.go:608-650`）。静态检索未见对以下直接合同的专门测试：TLS 证书 hostname/SNI 与 Host 一致；DNS 两次解析后由公网切换私网且确认未产生连接；多个全公网地址只拨已校验 IP；IPv6 ULA/loopback/link-local/multicast/保留段；私网 allowlist 仍拒绝 loopback/link-local/mixed 越界；显式非 443 端口；探测路径的重定向与 credential 不外泄。

**影响**：当前实现的核心性质集中在少量定制代码与 Go transport 隐式行为中，未来重构 resolver、keepalive、proxy、TLS 或 redirect 设置时，可能无回归信号地削弱 SSRF/credential 边界。

**建议**：用仅监听临时目录/临时端口的隔离测试 resolver、Dialer 与 TLS server 建立表驱动合同；测试不得访问真实网络。至少断言“所有答案先校验、拨号使用返回的已校验 IP、URL hostname 保持为 Host/SNI、证书错名失败、3xx 第二目标零请求/零 Authorization”。

### P2-01：HTTPS 公网端点可使用任意有效端口

**事实**：端口只校验 1..65535（`internal/relay/security.go:151-156`），所有 global-unicast 公网 IP 均可被管理面配置并拨号。不存在 443 或供应商端口 allowlist。

**影响**：已获上游配置权限的主体可让服务连接公网任意 TCP 端口并尝试 TLS/HTTP；这扩大端口扫描、协议探测和非预期出口面。由于目标仍须公网、TLS 验证仍启用、响应按 HTTP 解析，且该主体已是管理权限，评为 P2。

**建议**：记录这是产品合同还是收紧为默认 443；如需兼容供应商自定义端口，可提供显式端口 allowlist。无论选择何者，都补充 443、非 443、端口 0/65536 和 IPv6 bracket authority 测试。

### P2-02：TLS 最低版本和信任策略依赖 Go 默认值

**事实**：transport 没有显式 `TLSClientConfig`（`internal/relay/server.go:74`）。代码因此依赖所构建 Go 版本的默认最低 TLS 版本、cipher suites、系统 trust store 与 hostname 验证；当前代码也没有证书钉扎或每供应商自定义 CA 合同。

**影响**：当前标准库默认值通常安全且证书验证开启，但升级 Go、跨发行版部署或未来加入私有 CA 时，安全合同不够显式，审计和回归较弱。未发现禁用验证的事实，故为 P2 而非漏洞结论。

**建议**：显式设置最低 TLS 版本（至少 TLS 1.2，若产品兼容性允许则 TLS 1.3），保留标准 hostname 验证；将系统根/私有 CA 决策写入部署合同并添加错名、自签名、过低版本失败测试。

## 未发现的问题与剩余假设

- **未发现 P0**：没有看到默认允许 HTTP、自动重定向、环境代理、校验后按 hostname 二次拨号、只校验首个 DNS answer、接受 userinfo、禁用证书校验等路径。
- **私网/环回/链路本地**：默认均拒绝；显式私网 allowlist 只放宽 `IsPrivate`，不放宽 loopback/link-local（事实）。云厂商 metadata 若位于 link-local 会被拒绝；若某部署把 metadata 暴露在普通 RFC1918 地址，显式 allowlist 该精确 authority 后会被当作获授权私网目标，这是运维授权模型的剩余假设。
- **多 IP 可用性**：全部答案合规后仅拨第一个（`internal/relay/security.go:136,178`），没有逐地址 fallback。它可能造成双栈/首地址故障时的可用性问题，但不会扩大网络权限。
- **authority 精确匹配**：私网例外按原始 `parsed.Host` 字符串匹配，而非 DNS 名规范化（`internal/relay/security.go:68-75,145,158`）。大小写、尾点、默认端口省略与 IPv6 文本差异更可能导致 fail-closed/运维困惑，静态上未见由此获得未授权私网访问。
- **路径拼接**：基础 URL 禁止 query/fragment/RawPath/`..`/`//` 并规范化，因此两处字符串拼接未见 authority 切换；但这是当前 validator 与消费者必须共同保持的直接合同。

## 优先级汇总

| 优先级 | 项目 | 当前防护 | 主要影响 |
|---|---|---|---|
| P0 | 无 | 默认 HTTPS、双阶段 DNS/IP 校验、固定 IP 拨号、无重定向 | 未发现立即边界突破 |
| P1 | 私网 allowlist 同时允许 HTTP | 精确 authority；仍拒绝 loopback/link-local | 私网流量与 bearer 明文泄露/篡改 |
| P1 | 网络安全性质回归覆盖不足 | 已有基础 SSRF 与 redirect 测试 | 重构后静默回归风险 |
| P2 | 公网任意端口 | 仍要求公网 IP、TLS、HTTP | 扩大出口探测面 |
| P2 | TLS 策略依赖默认值 | 标准库默认验证，未见跳过验证 | 部署差异与审计不明确 |
