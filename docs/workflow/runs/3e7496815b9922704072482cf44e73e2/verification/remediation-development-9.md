# 执行9：固定入口冷验证与完整采集

## 真实诊断与边界
用户回退指出已确认的接口事实：tests8的冻结argv只有make verify，不执行development7推荐的本Run shell脚本。seq8是在已有node_modules工作区重跑并中断（exit-1，102563bytes，平台truncated=false、用户已读EOF）；部分门禁通过不是完整通过。seq5完整111871bytes原exit0已由维护者读至EOF，但其旧开始状态不能事后补成冷证明。原日志正式读取路径 `/api/workflow-runs/3e7496815b9922704072482cf44e73e2/steps/5/output?format=text`；本节点不手抄为完整副本、不修改QA6 No-Go。

最小缺口在产品验证编排，不在产品UI/业务。新增项目Go工具tests/verify，由固定make verify实际调用。原配方全部原序移入内部verify-gates，原历史红灯、格式、module、vet、frontend、Go/race、build、browser保留。锁定npm准备前置修复继续保持，package-lock/版本不变。无API/DB/UI/预算UTC语义、供应商、控制器、Connector argv、权限或用户数据修改。

材料沿用prepare回执：design-v0.2.0；archive sha256 `5f51d35fdc15ed7214749f9413879b4238097f1d6b6bbe2cc0d508d8e103cef7`。实际manifest路径 `.workflow-input/3e7496815b9922704072482cf44e73e2/5f51d35fdc15ed7214749f9413879b4238097f1d6b6bbe2cc0d508d8e103cef7/files/manifest.json`；需求/交互/AC路径仍是该root下docs/prd.md、docs/interactions.md、docs/acceptance.md（原任务追踪不变）。manifest状态frozen-for-test-implementation，不冒充新批准，不执行包内脚本。现有Trellis源码规范与scripts可用且非空白模板；本次使用before-dev/check/update-spec，不声称本节点执行了prepare初始化。

## 可复用执行契约
- tests固定命令仍make verify；无外部推荐命令、先删依赖、手工npm安装、NODE_PATH或全局node模块。Go runner按当前Git tracked/unignored清单复制新候选（ignored依赖/cache/运行库不入副本），保留只读原Git历史。
- 入口实测副本node_modules/.cache/npm缺失，清除NODE_PATH/NODE_OPTIONS、npm用户/全局配置，设置新Go build cache；前后核对原树与副本源码戳/HEAD。源码变化、安装/任一门禁、日志写入失败均非零。首次Chromium/系统库仍是已有宿主前置，不要求平台临时预装node_modules来通过。
- 每次唯一ignored目录test-results/verification/<UTC>-<unique>，自动打印相对路径；verify.log为完整脱敏stdout/stderr并流式传给Connector，state.json原子更新：测量标志、时间、工具版本、开始/结束状态、源码、门禁/最终退出、日志字节/SHA256、锁匹配本地Playwright/core与Chromium版本。没有硬编码本Run/旧摘要。browser/保留原始产物，只供独立QA白名单隐私审查。
- SIGINT/SIGTERM清理自身进程组和临时副本。仅针对本次拥有的group作有界signal-0探测，记录owned_group_absent，残留则非零；不将其冒充所有历史/脱离进程或端口已消失的证明。SIGKILL可能留未完成state，必须判不完整。原工作区依赖、旧安装数据与密码不变；不杀其他进程。

## 研发验证与未决
最终定向race/静态/同版检查见development-9-check.log；此前development-9-tests.log、development-9-tests-final.log、development-9-tests-candidate.log和development-9-static.log保留。最新源码见source-development-9.json。这些不等于正式完整冷门禁。首次定向执行发现两个新增测试驱动错误：substring "old"误匹配临时路径folders、非CommandContext却设置Cancel；已最小修复测试并重新执行，未改产品断言。初次终端回执非绿（Go exit1），此处保留失败诊断，不伪造未保存的完整首轮日志。

原生环境稳定监听/网络限制已有本Run首次verify-before.log/frontend-after.log等证据，不重复必然失败的完整make verify。检查seq8残留现场时ps被原生沙箱拒绝（operation not permitted）；未获可归属PID/端口事实，不声称旧现场已清理、不做任何广泛kill。新增actual subprocess取消回归只操作自身新进程组，并非seq8现场证明。

Trellis-check按实际项目规范检查变更范围、复用tests/evidence、保留全部原门禁、源码/回执一致性、无凭据日志/依赖版本变化；新增夹具只证明隔离、采集、失败传播。UI/API/DB跨层业务流不适用，基础设施env→copy→gates→capture→cleanup路径已定向覆盖。公共计划/任务/里程碑已同步；Ready for Dispatch到正式tests，但任务仍In Review、M01仍Blocked by Validation，非Accepted/Delivered。

## 下游必须完成

最终实测：development-9-check.log整体exit0，定向race两包PASS、格式/module/vet/build/diff均exit0、前后源码一致。HEAD `cd413b5aa24a372a19c19af999a6ef61b5c41572`；144件源码sha256 `af09459112ac061976688f434891d3f859e476c0efdfd285b645f504e9759ca3`。日志中的failed_fixture / `FAILED: exit status 23`是故意失败的隔离采集负向用例，外层该用例PASS，不能误读成完整产品门禁红或绿。此前旧142件摘要不覆盖新工具；新候选完整宿主仍Not Run。

1. 正式宿主执行原make verify，使用入口打印的新唯一目录；核对state有clean_start_measured=true、开始依赖缺失/NODE_PATH unset、source_before/source_after与本次候选一致、原树与副本不变、最终finished_utc、gate_exit=0及exit_code=0。确认完整日志所有历史/预算UTC/迁移/三视口J01–J07十页/browser门禁均在新源码上通过，而非夹具或旧日志。
2. 将新verify.log/state.json真实脱敏文件归档本Runverification，验证字节数及SHA256，保留失败记录；独立QA审查选取实际browser白名单文件/具体索引。执行器已采集本地完整日志，不依赖Connector缩略输出充当完整文件；不得用目录引用冒充日志。若工具容量截断仍按offset读EOF，不能从缺失内容推断成功。
3. 新完整冷绿后独立QA重审QA6报告/矩阵和冷AC，保持原生QA浏览器Blocked/供应商Not Run与宿主受控Pass分层。旧seq5完整日志可引用正式API/read_command_output，未自动归档的历史副本仍不能捏造。
4. QA通过后由获授权Connector/协调者新PR、精确已验提交合并，正式Owner deploy=true target=preview；Deployment/environment_url、healthz.version/schema3和真实admin页面闭环前不宣称上线。无正式部署新回执，不重置旧数据或启动替代服务。同SHA日志覆盖属Platform维护源，本产品不改控制器。
