# T001 登录页安全核对密码

- 状态：In Review
- 里程碑：M01；模块：internal/relay/web 与 tests/browser
- 优先级：High；负责人：研发节点；评审人：独立 QA
- 执行类型：行为回归驱动的实现／Validation
- 硬依赖：G1、G2、源码规范 bootstrap；软依赖：无
- 冲突范围：登录 DOM、秘密生命周期、两类前端测试；单一执行者
- 开始：2026-10-06；基线：73681a7df4c5cbe5914056ac13348154d58726e0

## 权威输入与目标
../product-definition.md 的 AC01—AC13 与 ../password-visibility-architecture.md 第 1／4 节。原来只有 password 输入，本任务在实际登录 DOM 中增加原生切换，不拦截或重写认证。普通选择已授权；无需重复审批。

## 范围与交付物
- index.html：明确非提交按钮、标签、关联和 live 状态。
- app.js：type 为唯一事实源，单一同步函数；signedIn 和 clearSecrets 重置。
- style.css：仅登录按钮间距。
- ui.test.mjs：保值、空值、零请求和生命周期回归。
- journey.mjs：真实键盘、语义、零 submit／invalid／请求、两态登录、退出和会话恢复。
- .trellis/spec：源码规范与本能力契约；本 Run 公共任务与真实验证证据。

非目标：后端、数据库、健康／版本协议、其他秘密输入显隐、偏好存储、权限、依赖变更、部署、供应商调用和 Git 发布。没有必要的行为重构。

## 验收与验证契约
环境：现有 Go 模块、锁定 Playwright、隔离临时目录／随机端口；准备沿用 Makefile browser 的 npm ci，不改门禁。
最小反馈 node --test tests/browser/ui.test.mjs；完整入口 make verify。成功要求固定入口退出 0 和独立 QA 逐项核对。失败保留首次和后续日志；沙箱受限交宿主 tests，不要求权限扩大。
AC01—02 对应初态、刷新、空值／特殊字符反复切换；03—05 对应 native Tab／Enter／Space、焦点、名称／描述／AX、零请求和表单事件；06—08 对应真实两态登录、拒绝、清空重置、退出、会话刷新及管理 401；09 对应 URL／存储不变、日志审计与代码检查；10—11 沿用完整管理台、健康／版本回归；12—13 对应当前回执、代码标识、历史 SHA-256 与 diff。
证据不打印秘密，布尔比较保值，无显示密码截图／trace。无需新增人工运维步骤。

## 实施结果与交接
实现及新增回归已形成，位置见上方 5 个源码／测试文件；未提交／发布。无接口、数据变化或设计偏差。实际完成 Trellis 源码 bootstrap、before-dev、update-spec 与 check，动作和真实结果见 [研发验证](../verification/M01-development.md)。27 项 VM、format、vet、模块、build 与健康／版本 handler 回归通过；make verify 退出 2，原生沙箱禁止 HTTP 监听，真实浏览器未执行。因此当前 In Review，不是 Done。

当前可依赖：原生非提交按钮与 type 单一状态、生命周期重置、回归代码、局部测试及历史保护证据。未完成：宿主固定门禁与独立 QA。下一步平台 tests 对最终工作树运行 make verify，然后 QA 逐项核对；保存本 Run 的当前失败和后续证据，不扩大权限。最近成功验证见 non-listening-checks.log；完整门禁失败见 make-verify-attempt-1.log。历史文档和日志已复原且 diff 为零，只是保护对象，不作为本轮测试通过。
