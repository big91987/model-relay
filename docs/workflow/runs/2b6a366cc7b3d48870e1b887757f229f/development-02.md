# 首轮研发实施记录（执行2）

2026-10-14，Issue #48，Run `2b6a366cc7b3d48870e1b887757f229f`，Workflow revision3；HEAD `30b4e13c19e36d51b72ef1186468d6adf39ac822`。本记录仅为研发首轮，不是 QA 或最终交付报告。

## 授权故障注入

按用户明确要求，新文档 `docs/harness/opensandbox-publish-recovery.md` 故意遗漏“停止、恢复与环境释放”章节及其操作说明。不得首轮提前补齐；须在本轮固定测试通过后，由独立 QA 实际拒绝并通过原生 revise 返工，再由研发修订、重新固定测试及独立复核。`first-round/opensandbox-publish-recovery.md` 保留首轮原字节，后续不可覆盖。已有 `docs/harness/opensandbox-run-sharing.md` 不改。

## 读取及核验

读取根 AGENTS.md；遍历未发现下级 AGENTS.md。读取 prepare 冻结输入任务全文及身份白名单、正式 prepare 回执和准备身份；不读任何凭据值。冻结 issue_number 为48。初始工作区干净，HEAD及准备身份匹配；声明 main 对象存在且是 HEAD 祖先。SOURCE.json 的146项模板文件全部摘要匹配维护源 bc43314。既有全仓无网络发布扫描通过，没有放宽规则或执行发布。

按 `tests/evidence/main.go` 相同路径排序/长度/NUL/原字节算法、不启动 Go 编译，重算163项源码摘要，匹配任务的 `991c79ca919b3d0ccf148c066c3b24d993090fcb78ab864c0e4637cfff8040ba`。初始 node_modules、.cache/npm、.cache/go-build、bin、test-results 均不存在。这里只证明工作区冷起点，不读取全局缓存、不宣称全局缓存为空。GITHUB_TOKEN/GH_TOKEN/GIT_ASKPASS 仅检查存在性，均 false。

## 实际改动与工具边界

只新增新交付文档及本 Run 的冻结身份、实施记录、文档原字节快照、源码身份清单、离线检查结果和具体交付校验清单。未修改产品业务、任何测试、Makefile、依赖锁、模板、旧 QA/报告或旧完整文档。不创建分支、不提交、push、PR、部署或改平台。当前可调用工具没有 apply_patch，环境亦未找到该命令，因此新增文本由终端 heredoc 写入，身份及校验产物用 Python 标准库生成；未安装工具或依赖。

证据与构建输入分离：仅文档 Markdown 原字节快照；163项源码仅归档原路径/字节数/SHA-256 清单，不创建可编译副本。清单算法与范围明确，不将其冒充执行 `go run ./tests/evidence` 的回执。原路径映射见归档 manifest；清单自身不包含自身摘要，避免自指。完整新文件清单供下游核对，不要求 QA 或 report 将未生成的文件视为已存在。

本轮不运行 make verify、临时替代全套、Go构建、npm 或浏览器，保持固定 tests 冷起点。研发校验限于原字节快照比较、文件摘要、文档故障边界、git diff --check、离线发布扫描与版本核对；实际结果见 checks.json。固定24分钟内层及1800秒外层保持不变。

## 待执行与交接

交原生 tests 执行冻结的原 make verify，保存完整外层原回执与内层 verify.log/state、退出码、冷起点、源码/候选一致性和清理。失败先完整读日志，修复或有新证据再交接，不同版循环求绿。通过后独立 QA 必须在本轮实际拒绝授权缺章，审查当前测试设施及证据；旧 tests12/QA13 仅历史。

首轮故意缺章是本轮唯一预期 QA 返工，不声称 tests 必然通过。QA 放行后的正式操作者控制核验、decision、report、固定 publisher 真实草稿 PR、远端同版核验均仍待对应角色执行。研发不代 QA、不代确认或发布；材料保存和环境生命周期完成也不可臆造。

## 归档后清理

导入既有离线扫描函数产生两个未跟踪 Python 字节码文件，仅删除本轮新生成的 materials/repository 字节码并移除已空的 __pycache__，未删除工作区、源码或原失败证据。最终扫描使用 python3 -B 禁止再次生成字节码；不归档缓存。源码及旧文件保持原字节，新交付12个具体文件经 SHA256SUMS 核验。该清理不构成平台环境释放，也不替代固定 tests 清理结果。
