# 第一阶段基线与上游锁定报告

生成日期：2026-10-20（Asia/Shanghai）

## 已完成

- 当前生产基线 tag：`58token-pre-upgrade`，指向 `f2f785fd6af5233f570841003fb3fb23d830d78b`。
- 上游 tag：`upstream-v0.2.17`，目标 commit `060d0160417782cd962a0c7f460f6cd4565d4916`。
- 当前升级分支：`codex/upgrade-sub2api-v0.2.17`。
- 已确认当前基线和官方目标没有共同祖先，后续不能直接执行普通 `git merge`。
- 升级方案已纳入当前分支：`a6843a5 docs: add Sub2API v0.2.17 upgrade plan`。

## 快照规模

| 项目 | 当前基线 | v0.2.17 |
| --- | ---: | ---: |
| Git 文件数 | 2388 | 4338 |
| SQL 迁移文件数 | 205 | 296 |
| 同名迁移 | 175 | 175 |
| 仅当前基线迁移 | 30 | — |
| 仅目标版本迁移 | — | 121 |

同名迁移中已经发现 `001_init.sql`、`002_account_type_migration.sql`、`003_subscription.sql` 等内容差异；迁移文件不能按文件名直接覆盖。当前定制迁移主要集中在 151–170，目标版本则在相同编号段使用了不同的上游迁移，需逐文件建立冲突表。

## 工具链和部署差异

- 基线 Go：`1.26.3`。
- 目标 Go：`1.27.2`。
- 前端目标构建新增 i18n 检查，依赖锁文件有较大差异。
- `backend/go.mod`、`backend/go.sum`、`frontend/package.json`、`frontend/pnpm-lock.yaml`、`deploy/docker-compose.local.yml` 和 CI workflow 均需要模块级审查。
- Compose 生产镜像必须改为本项目合并后构建的固定 digest，不能直接部署官方 `latest` 或官方 tag 镜像替代定制代码。

## 下一步门禁

1. 完成可信官方共同基线或初始导入差异审查。
2. 建立完整迁移冲突表，禁止覆盖已应用迁移和关闭 checksum 校验。
3. 按依赖、后端网关/调度、计费支付、前端、部署分模块移植。
4. 完成隔离 Compose 构建后，再进入 PostgreSQL 副本迁移演练。

本报告只记录 Git 和源码快照，不包含生产数据库备份；生产备份必须在实际 Docker Compose/PostgreSQL 主机执行。
