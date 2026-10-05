# 本机持续预览

本项目的 `.github/workflows/deploy-local.yml` 由 Agent Platform 维护源 `examples/github/model-relay-preview/install_workflow.py` 安装，版本摘要在 `deploy/platform-preview-source.json`。它只在代码经人工审查并合并到 `main` 后触发；不合并 PR，也不把草稿 PR 当成部署。

部署准备由受信任的同机 Runner 执行：精确检查当前 main SHA，在独立工作树运行本项目完整 `make verify`，然后把构建二进制标记为该 SHA。合并触发的运行只准备；Owner 在 GitHub Actions 手动运行 `Deploy Model Relay locally`、选择 `main` 并勾选 `deploy` 后，受信任控制器才会切换本机服务，检查 `/healthz` 回报的运行版本；失败尝试恢复上一二进制和正式备份。`local-preview` Environment 限制 main；当前私有仓库套餐不支持 Required reviewer，所以手动部署动作是明确的人审入口。当前首版使用 Go 服务和 SQLite，不能复用旧静态网页的部署器。

安装与升级控制器、仓库变量、部署环境、回退边界，以 Agent Platform 维护源 `examples/github/model-relay-preview/README.md` 为准。目标目录必须私有并位于项目工作区外，服务仅监听 `127.0.0.1`，固定地址由仓库变量 `MODEL_RELAY_PREVIEW_URL` 指定。首次部署生成的管理员密码只保存在目标主机私有文件，不进仓库或 Actions 日志。上游供应商 URL 与凭据仍要由管理员通过产品网页配置；未获得真实响应前不能称供应商联调通过。

验收顺序：平台运行页查看 PR → Owner 检查并自行合并 → GitHub Actions 自动准备 → Owner 手动触发部署 → 比较合并的 main SHA、部署记录 SHA 和 `/healthz` 的 `version` → 打开固定管理页验证效果。部署尚未完成时，固定 URL 不得写成已发布地址。当前草稿 PR 的代码和部署 Workflow 尚未合并，首次真实 Actions 发布仍待用户合并后验收。
