# 宿主执行24失败：最小归因

Connector完整make verify exit2，89307 bytes/truncated=false。本节点恢复末段offset82500→89307/eof=true；未声称本节点逐页审阅完整日志。HEAD98511771871cf0951ecef716bbb55deb13e18da5、122文件指纹182077621a4dbc04d73b8517a40f70811c9c8e46f435d7d9596edcfb5018bab2，与attempt23来源一致。

末段Go relay race、迁移契约、构建、原迁移/旧UI/路由通过，console-real-go-ui 1280/1440/390 passed；这些是旧候选局部事实，不等于新增连续UI通过。历史health-history旧版预期红不是最终退出原因。宿主真实监听，失败不归入沙箱限制。

原断点投影：

```json
{"suite":"continuous-console","width":1280,"result":"failed","stage":"J01-empty-state","action":"authoritative-read","last_status":null,"checkpoints":[],"error_kind":"assertion","editor_open":false,"editor_invalid":0,"editor_failure":false,"workspace_visible":true,"secret_present":false}
```

J01–J07没有完成，1440/390连续旅程未到达。原日志由平台read_command_output(seq24)留存，未把截取末段伪装成完整日志副本。

## 已证契约与未决边界

journeys.go以InitConsole创建全新隔离schema3安装；store.go:initSchema在version3时显式移除legacy tenant/membership/limit。旧journeys.mjs却要求tenants.length=1、id=legacy，再读取legacy keys。确定的新安装/迁移契约错配，不应通过添加产品legacy数据解决。

新增非监听回归通过真实InitConsole/Open/login/ServeHTTP/SQLite验证三个集合均非null空数组、legacy keys 404。生产syncTenants仍为平台管理员保留一个空值“全部租户”选项，它不是已有租户。驱动按该契约检查空态，然后全部业务对象仍经UI创建；独立迁移套件legacy守恒断言不改。

原投影不能证明具体失败读取。候选加入分集合stage、read实际数字status和固定三个initial_counts，只记录白名单事实、不打印API正文/异常/秘密。已修确定源码契约，不声称宿主唯一原因或剩余连续旅程已通过。
