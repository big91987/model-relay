# 缺陷与强制验收缺口

| ID | 等级 | 用户能看到的问题 | 复现步骤 | 影响 | 定位信息 | 当前状态 | 回归资产 |
|---|---|---|---|---|---|---|---|
| Issue28-CI | P1，沿原发布阻塞主题 | 全新安装prepare缺包，无法激活新版本 | 旧入口无node_modules时make frontend-test；正式prepare原失败来自用户提供事实 | 全新安装发布受阻 | 原Makefile frontend-test先于最终browser安装 | Ready for Retest：实现已修，普通宿主全量绿；干净独立宿主原路径未证实 | tests/verification/dependencies_test.go；原route-cleanup；固定make verify |
| E28-01 | 强制AC证据缺口，发布阻塞；不是新业务P1结论 | 不能证明新安装不会再次靠工作区缓存掩盖问题 | 执行5日志起点仅源码戳；对照T001第3项，缺独立目录/无模块/NODE_PATH/未预装记录 | 无法独立批准用户明确要求的干净供应链修复 | seq5 exit0有效，但执行条件不足；QA开始时模块已在 | Open，development组织宿主补证 | 保留原全部门禁、真实干净前置及完整同版日志 |
| E28-02 | 交付证据缺口，随E28-01一起补档 | 仓库交接只有宿主摘要，不能直接复核完整执行日志 | 本Runverification/acceptance仅有摘要和原生/研发日志，seq5完整日志只在平台保留 | 审查/发布后追溯欠完整 | 平台retained111871 bytes、truncated=false；不可手抄摘要冒充完整 | Open | 正式Connector保存脱敏完整stdout/stderr/exit/源戳、具体相对文件及索引 |

无新产品实现Fail；No-Go来源是明确强制证据未闭合，不能借此放宽门禁。独立QA原生监听限制为环境Blocked，不作为产品缺陷、不要求elevation。正式部署尚在后续合法Connector责任域，不在本轮伪造缺陷已关闭或已上线。
