# 本机持续预览

本项目的 `.github/workflows/deploy-local.yml` 由 Agent Platform 维护源 `examples/github/model-relay-preview/install_workflow.py` 安装，版本摘要在 `deploy/platform-preview-source.json`。合并触发的准备只在代码经审查并合并到 `main` 后运行；该工作流不合并 PR，也不把草稿 PR 当成部署。用户已授权本目标维护者核对独立 QA 后自主完成测试仓 PR 合并与测试部署，无需等待用户另行执行或批准；阶段 Agent 仍交付草稿 PR，平台源仓 PR 不在合并授权内。

部署准备由受信任的同机 Runner 执行：精确检查当前 main SHA，在独立工作树运行本项目完整 `make verify`，然后把构建二进制标记为该 SHA。合并触发的运行只准备；获授权的目标维护者通过 GitHub Actions 显式运行 `Deploy Model Relay locally`、选择 `main` 并勾选 `deploy` 后，受信任控制器才会切换本机服务，检查 `/healthz` 回报的运行版本；失败尝试恢复上一二进制和正式备份。`local-preview` Environment 限制 main；现有配置使用显式部署入口而非合并即发布，授权不改变工作流机制、账户权限或部署成功证据要求。当前首版使用 Go 服务和 SQLite，不能复用旧静态网页的部署器。

安装与升级控制器、仓库变量、部署环境、回退边界，以 Agent Platform 维护源 `examples/github/model-relay-preview/README.md` 为准。目标目录必须私有并位于项目工作区外，服务仅监听 `127.0.0.1`，固定地址由仓库变量 `MODEL_RELAY_PREVIEW_URL` 指定。首次部署生成的管理员密码只保存在目标主机私有文件，不进仓库或 Actions 日志。上游供应商 URL 与凭据仍要由管理员通过产品网页配置；未获得真实响应前不能称供应商联调通过。

验收顺序：平台运行页查看草稿 PR 和独立 QA → 获授权目标维护者核对并合并测试仓 PR → GitHub Actions 自动准备 → 获授权目标维护者显式触发正式测试部署 → 比较 PR 合并提交、部署记录 SHA 和 `/healthz` 的 `version` → 打开 GitHub Deployment 的 `environment_url` 或平台 Run 页的效果地址，核验部署后浏览器体验。若 main 后续有提交，部署 SHA 应包含原 PR 合并提交，不能要求两者始终相等。部署失败时保留失败记录，通过维护源说明中的正式恢复路径处理；成功状态出现前不能把地址称为已发布。

首次真实发布已由 [PR #2](https://github.com/big91987/model-relay/pull/2) 合并及 [Actions 部署记录](https://github.com/big91987/model-relay/actions/runs/37306893618) 验证。发布后可用页面及版本应始终以当前 Deployment 状态和 `/healthz` 为准。
