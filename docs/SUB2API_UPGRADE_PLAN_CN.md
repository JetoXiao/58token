# 58Token 升级 Sub2API v0.2.17 实施方案

编制日期：2026-10-19。本文交付升级方案，不执行代码升级、发布或生产数据操作。

## 1. 已确认范围与版本依据

| 项目 | 确认值 |
| --- | --- |
| 当前生产代码 | `dev`，检查时 HEAD 为 `f2f785fd6af5233f570841003fb3fb23d830d78b` |
| 当前远端 | `https://github.com/JetoXiao/58token.git` |
| 当前版本文件 | `backend/cmd/server/VERSION` 为 `0.1.130`；这是文件值，不能直接证明与官方该 tag 内容相同 |
| 升级目标 | 官方 `v0.2.17`，固定 commit `060d0160417782cd962a0c7f460f6cd4565d4916` |
| 部署/数据库 | Docker Compose / PostgreSQL |
| 发布时间策略 | 先独立测试一段时间，再安排数小时维护窗口发布 |
| 国产模型 | 优先沿用新版官方支持，不重复开发适配器 |
| 视频协议 | 尚未选定；按目标版本实际支持的平台与客户端做契约验收 |

上游发布于北京时间 2026-10-16 12:41:43。已通过 GitHub API 核实 tag、commit、发布说明、部分源码及目录；尚未下载完整上游历史、执行真实合并、构建或迁移。因此本文不声称已获得全部冲突列表、已证明兼容或已测得迁移时间。

参考：

- [v0.2.17 发布说明](https://github.com/Wei-Shaw/sub2api/releases/tag/v0.2.17)
- [目标版本源码](https://github.com/Wei-Shaw/sub2api/tree/060d0160417782cd962a0c7f460f6cd4565d4916)
- 本仓库：`backend/internal/repository/migrations_runner.go`、`backend/migrations/README.md`、`.github/workflows/backend-ci.yml`。

### 1.1 已发现的关键事实

1. 当前只有本地 `dev` 和 `origin/dev`，没有本地 tag；完整本地历史起点为独立的 `Initial commit`。须先核实是否与官方有共同祖先，不能直接假设可普通 merge。
2. 当前 SQL 迁移最高数字前缀为 `170`，目标目录最高为 `245`。编号不是版本水位：相同前缀可能有多个文件，上游新增迁移也可能落在较小前缀。必须按完整文件名、内容、依赖与数据库记录逐项比较。
3. 当前迁移以完整文件名为主键，计算去掉首尾空白后的 SHA256，并按文件名字典序执行。存在历史 checksum 精确兼容规则，合并时须保留其依据，不得用关闭校验解决冲突。
4. 当前曾有 Sora 迁移，但 `090_drop_sora.sql` 已删除相关表和字段；不能把历史 SQL 当作仍可用的视频实现。
5. 当前 Go 为 `1.26.3`，目标 `go.mod` 为 `1.27.2`。本地前端 CI 使用 Node 20、pnpm 9；目标前端工具链须在完整拉取后核对。
6. `deploy/docker-compose.local.yml` 当前仍是 `weishaw/sub2api:latest`。生产升级必须使用本项目合并后自建镜像的 digest，直接换成官方镜像会丢失二次开发代码。
7. v0.2.17 包含支付退款、易支付查单、Stripe secret 传递、模型计费及路由修复；GPT-5.5 在 ChatGPT OAuth/SetupToken 账号退役，API Key 账号不受同样限制，须检查线上模型映射。

## 2. 推荐路线、分支和交付物

推荐：固定上游快照 → 找到可信比较基线 → 按功能合并 → 迁移专项处理 → 独立测试 7–14 天 → 维护窗口切换。7–14 天是建议值，可按实际支付与业务周期延长。

分支约定：

| 名称 | 用途与规则 |
| --- | --- |
| `dev` | 保持生产稳定；升级验收前不合入集成半成品 |
| `codex/upgrade-sub2api-v0.2.17` | 升级集成，合并检查和验收通过后 PR 至 `dev` |
| `codex/upgrade-v0.2.17-<domain>` | 短期模块分支，例如 migrations、gateway、payment、frontend |
| `codex/hotfix-<issue>` | 从生产基线修复；同步至 `dev` 和升级分支 |
| `58token-pre-upgrade` | 升级前代码 tag，固定名称并保留原生产提交 |
| `v0.2.17-58token.1-rc.1` / `v0.2.17-58token.1` | 候选 / 正式 tag；与官方 tag 分开 |

不要 rebase/force-push 已共享的 `dev`。已从 `dev` 创建集成分支时，旧定制已在分支中，不再整批 cherry-pick 相同提交。只有从干净上游重建的备选路线才需要移植定制。

每个阶段交付：

- P0：版本锁定表、生产配置/镜像清单、定制功能清单、备份恢复演练记录。
- P1：上游比较基线、代码差异分类表、迁移冲突表、功能取舍记录。
- P2：模块 PR、构建镜像 digest、自动化验证结果。
- P3：生产副本迁移/恢复计时、业务回归与金额核对报告。
- P4：试运行报告、上线/回滚执行单、最终放行记录。

建议由开发负责人负责合并与接口，DBA/运维负责备份、迁移及恢复，测试负责人维护用例证据，业务负责人确认账务与功能。若一人兼任，也须分别记录验收结论。

## 3. Git 准备和合并步骤

以下 Git 命令在开发环境运行，尖括号参数须先替换。文档提交后以实际冻结 commit 为准。

```bash
git status --short
git fetch origin
git switch dev
git pull --ff-only origin dev
git rev-parse HEAD
git tag -a 58token-pre-upgrade -m "58Token production baseline"
# tag 存在时先核对，不覆盖；审核后单独推送这个 tag
git push origin 58token-pre-upgrade
git remote -v
# upstream 尚不存在时添加；存在时先检查 URL
git remote add upstream https://github.com/Wei-Shaw/sub2api.git
git fetch upstream --no-tags '+refs/heads/main:refs/remotes/upstream/main'
git fetch upstream refs/tags/v0.2.17:refs/tags/upstream-v0.2.17
git rev-parse upstream-v0.2.17^{commit}
# 必须等于 060d0160417782cd962a0c7f460f6cd4565d4916
git switch -c codex/upgrade-sub2api-v0.2.17
git merge-base dev upstream-v0.2.17
```

### 3.1 有共同祖先

记录 merge-base 为 B，分别检查 B→dev 定制和 B→目标上游变更。确认 B 可信后：

```bash
git diff --stat <B> dev
git diff --stat <B> upstream-v0.2.17
git merge --no-commit --no-ff upstream-v0.2.17
git diff --name-only --diff-filter=U
```

不要对所有冲突统一 `ours` 或 `theirs`。按下节处理，迁移审查完成后再提交 merge。中止实验可在确认未混入额外修改后使用 `git merge --abort`。

### 3.2 无共同祖先：优先恢复比较基线

禁止直接 `--allow-unrelated-histories` 合并两棵完整目录，也不要给生产历史伪造祖先。

1. 找到二次开发最初导入 commit、原始源码来源或当时官方 tag；`VERSION=0.1.130` 仅作为候选线索。
2. 对照候选官方快照，检查网关、Ent schema、迁移、支付、前端路由与依赖锁文件，形成“基线匹配证据表”。初始快照可能已含前置定制，须一并盘点。
3. 如果能确认基线 B，在隔离升级目录创建合成三方合并实验：基于 B 导入已审核的生产源码快照，再合并目标上游。此实验只用于获得文件结果和冲突报告，不修改 `dev` 历史。
4. 将审查后的文件结果按模块移入从 `dev` 创建的升级分支，每模块提交并注明上游 commit、定制来源与冲突决策。
5. 若基线无法确认，采用干净 v0.2.17 快照逐功能移植定制，再将结果作为升级分支的受审查文件变更提交。代价较高，必须覆盖初始导入就存在的功能，不能只回放最近提交。

两条路线都使用双端点差异检查：`git diff dev upstream-v0.2.17`；无共同祖先时不用 `dev...upstream-v0.2.17` 代替它。正式集成保留 `dev` 历史，并在提交/清单中记录上游来源，后续升级沿用此次已确认的比较基线。

### 3.3 合并顺序与功能保留

| 顺序 | 功能/当前代码线索 | 合并原则和验收证据 |
| --- | --- | --- |
| 1 | 依赖、配置、migrations、Ent schema | 先解决 schema 和依赖；Wire/Ent 由工具生成并审核结果，避免手工拼生成文件 |
| 2 | `openai_account_scheduler*`、`failover_loop*`、`runtime_model_probe_service*` | 保留模型级隔离、熔断、探测恢复、粘性和图片重试语义；新官方调度若替代旧逻辑须有等价测试 |
| 3 | `gateway_forward_as_responses*`、`openai_gateway*`、`response_cache*` | 检查流式 usage、协议转换、缓存命中计费、重试与取消，防止多层重复重试/扣款 |
| 4 | `payment*`、subscription、affiliate、鉴权 | 合并官方退款修复与定制余额购买/返利，保持原有 OAuth 绑定、组授权及子管理员边界 |
| 5 | `model_marketplace*`、`upstream_pricing*`、账号余额 | 保留市场配置/虚拟配置/能力标签/上游报价；明确官方价、自定义价、倍率和兜底优先级 |
| 6 | visitor analytics、downloads、help center、运维面板 | 保留接口、持久化、资源权限、隐私脱敏及已有数据 |
| 7 | 前端 API/types/router/i18n、branding | 保留 58Token 品牌、菜单和业务入口；国产模型页面复用官方实现并整合权限 |
| 8 | Docker、CI、release、自动更新来源 | 镜像和下载/更新仓库指向本项目；核对被隐藏页面是否仍有接口权限约束 |

每个功能登记：现有入口、文件/提交、关联表、现网配置、目标行为、处理方式（保留/官方替代/停用）、用例、负责人。业务功能默认保留，停用须明确业务决策。

工具链同步检查 `go.mod`、所有 Dockerfile、CI 内写死的 `go1.26.3`、golangci-lint 兼容性，以及目标 Node/pnpm 与锁文件。不得仅改 go.mod 而保留旧构建镜像。保留许可证和上游署名。

## 4. 数据库升级专项

### 4.1 先制作冲突表，再启动候选应用

现有应用会在启动时自动迁移，测试环境必须指向隔离数据库。正式迁移执行入口和 Atlas 行为以合并后 runner 为准，检查 `schema_migrations` 与存在时的 `atlas_schema_revisions`。

逐文件输出：完整文件名、当前内容哈希、目标内容哈希、生产记录哈希、DDL/DML 摘要、依赖、锁风险、处置和验证 SQL。

| 差异 | 处置 |
| --- | --- |
| 同名同内容 | 保持原样并验证生产 checksum |
| 同名不同内容 | 最高风险；保留已发布旧文件，拆出目标变更为新补偿迁移，审查依赖；不得随意重命名旧文件或覆盖数据库 checksum |
| 同编号不同文件名 | runner 可能都执行；检查重复建表/字段和字典序依赖，不能只把编号改大就认为解决 |
| 仅本项目存在且已应用 | 保留；包括品牌、帮助中心、市场、统计等定制迁移 |
| 仅上游存在 | 审查是否引用被定制改变的结构、是否覆盖模型映射/价格、是否破坏旧数据，再决定原样采用或拆解 |

保留既有精确 checksum 兼容规则与测试，不新增全局跳过、不手工伪造迁移成功记录。补偿文件安排在依赖之后，使用统一编号预留规则；每次上游同步重新审查。`_notx.sql` 按 runner 限制只用于并发索引，检查失败后无效索引和重入行为。

本仓库迁移 README 中举例的 `make migrate-up/down` 未出现在当前 backend Makefile，不作为实施命令；没有通用自动降级流程，也不能把 Down SQL 放进同一迁移文件。

### 4.2 双路径演练

必须同时通过：

1. 空库 → 合并版本：证明新部署没有因保留历史 SQL 产生依赖错误。
2. 脱敏生产副本 → 合并版本 → 再次启动：证明增量迁移、checksum、幂等及业务数据可用。

备份可恢复性另做独立演练。记录备份、恢复、迁移、索引及启动耗时，观察 `pg_stat_activity`、锁等待、磁盘/WAL 增长、数据库连接和无效索引。大表回填分批，避免把历史数据全表更新塞进长事务。

验收不仅统计行数：按主键比较用户余额、订阅到期时间、分组权限、订单状态/金额、兑换记录、佣金、模型映射和定价。为预期变更单列白名单及前后值。真实表名/金额字段从合并后的 Ent schema 和数据库目录生成，不预设存在 `user_balances`、`subscriptions` 等表。

只读记录示例：

```sql
SELECT filename, checksum, applied_at FROM schema_migrations ORDER BY filename;
SELECT table_name FROM information_schema.tables WHERE table_schema = 'public';
SELECT table_name, column_name, data_type FROM information_schema.columns
WHERE table_schema = 'public' ORDER BY table_name, ordinal_position;
```

## 5. 国产模型与视频接口决策

目标源码存在国产供应商余额、配额和限流实现，不另建通用适配框架。测试前选择实际要开的供应商、模型 ID、账号类型、套餐与区域，并验证官方支持覆盖这些组合；模型名称出现在列表里不等于完整支持。

已核实目标路由存在：

- `POST /v1/videos`、`/v1/videos/generations`。
- `GET /v1/videos/:request_id`、`/v1/videos/:request_id/content`；另有 generations/edits/extensions 形式。
- Seedance：`POST /api/v3/contents/generations/tasks`，以及同路径任务 ID 的 GET/DELETE；存在 `/v3`、`/v1`、根路径等别名。

路由受平台限制，不能承诺所有国产供应商都能使用 `/v1/videos`。建议保留官方协议：需要 OpenAI 风格客户端时先验证 `/v1/videos` 的平台和参数；Seedance 客户端优先使用官方任务接口。只有客户端契约确有缺口才增加兼容层，避免本次升级顺带重写视频架构。

视频专项用例：文生/图生参数、任务创建/查询、允许时取消、成功/失败/超时、重启后查询、跨用户越权、结果下载及过期、并发额度、重复请求/查询的扣费、计费单位（秒/分辨率/请求等）和无 usage 情况。无取消/回调/退款支持的组合明确标为不支持，不假定存在通用 webhook 或退款。

创建请求超时可能已在上游生成任务：不能不加判断地换号重试。验证任务与原供应商/账号绑定及幂等/查询能力，无法确认时返回可追踪状态并人工对账，防止重复生成和重复费用。

## 6. 测试环境、回归与证据

### 6.1 环境隔离

优先使用单独测试服务器；同机仅改 Compose project name 不够，当前有固定 `container_name`、端口和相对 bind mount，均须调整。隔离 PostgreSQL、Redis/第二缓存 Redis、data 目录、S3 bucket/prefix、域名、邮件、OAuth 和支付回调。

生产副本先封锁外部出站与回调，再导入并清理敏感数据，禁用定时任务、真实支付/退款/邮件、自动探测、自动更新及其他会写上游的作业。配置专用低额度模型账号，不能直接用克隆的生产 OAuth token 发起刷新，避免影响生产。备份及配置文件限权，不提交 Git。

### 6.2 构建检查

以下为当前仓库可识别命令，在合并版本和目标工具链环境执行；先解决原版基线已有失败并记录，不能把全部失败归因于升级：

```bash
cd backend
make test-unit
make test-integration
golangci-lint run ./...
cd ../frontend
pnpm install --frozen-lockfile
pnpm run lint:check
pnpm run typecheck
pnpm run test:run
pnpm run build
```

集成测试依赖按测试代码准备 Docker/PostgreSQL/Redis。已有 CI 前端只跑一组选定关键 Vitest，不等于全量业务端到端回归；发布前补实际浏览器流程和目标镜像启动验证。依赖漏洞检查按合并后的 CI 与例外清单执行。生成代码执行一次后再次生成应无意外差异。

### 6.3 回归矩阵

| 等级 | 范围 | 核心验证与证据 |
| --- | --- | --- |
| P0 | 登录/OAuth/TOTP/权限 | 老用户登录、旧 Key、绑定身份、私有组和子管理员越权；记录状态码和用户结果 |
| P0 | 余额/支付/订阅/返利 | 创建、成功/重复/乱序回调、超时查单、部分/待定/失败退款、余额购买；订单→账务→余额逐笔核对 |
| P0 | 既有网关 | chat/responses/messages/Gemini 原生及实际客户端；流式事件、工具调用、图像、断流 usage、取消、模型映射 |
| P0 | 调度/计费/缓存 | 429/5xx/超时/模型错误、模型级熔断与恢复、粘性、重试上限、缓存命中；一请求一账且不跨用户复用 |
| P0 | 国产模型 | 每个启用厂商至少一个真实低额度模型，普通/流式/推理/工具能力按支持范围验证；配额和价格正确 |
| P0 | 视频 | 第 5 节完整生命周期，跨租户与金额测试；未开功能不算支持通过 |
| P1 | 市场/报价/运营 | 定制价优先级、虚拟配置、标签、上游余额、资源下载、访问统计、帮助中心和运维面板 |
| P1 | 部署/恢复 | 固定镜像、冷启动、重启、迁移重复启动、备份恢复、监控与日志脱敏 |

测试采用两层：mock 固定 usage/错误注入，精确核对协议与计费；真实供应商低额度验证能力和上游账单。真实模型回答内容和生成 token 数可能变化，不用逐字相同作为通过条件。金额按既有 decimal 精度及已确认取整规则精确核对，不能用宽泛容差掩盖重复扣费。

每条用例保存 commit、digest、配置版本、时间、脱敏请求/响应、request/task ID、使用账号、预期/实际账务与截图/日志。建议增加语义契约测试覆盖已有定制，避免只测试函数实现细节。

### 6.4 试运行与门禁

建议连续 7–14 天内部试用，覆盖实际支付查单/退款和订阅周期边界；至少一次重启、峰值并发压测、供应商故障注入及恢复演练。测试期生产仍运行旧版本，不能把新旧应用接同一个生产库试用。

建议放行阈值（基线和容量确认后冻结）：P0 全通过；P1 无未解释回归；账务重复扣费/丢账/越权为 0；同负载普通接口 P95 劣化不超过 20%；非预期 5xx 相比基线不增加超过 1 个百分点。视频延迟和成功率按供应商独立设基线，不与文本接口混算。任何账务/权限异常立即阻断发布。

## 7. Compose 发布 Runbook

采用一次维护切换，默认不做新旧程序共享数据库的百分比灰度。测试验收后，在维护模式内先由白名单验证，再全量开放。

### 7.1 上线前一天

- PR 审核通过，构建一次自有候选镜像；验证后按同一 digest 晋级，保存旧镜像且可离线恢复。
- Compose app 镜像改为本项目 registry digest，例如 `registry.example.com/58token/sub2api@sha256:<实际摘要>`。禁止直接部署 `weishaw/sub2api:0.2.17` 或 `latest` 代替合并版本。
- 对实际 Compose 文件、env、数据挂载、DB/Redis 版本、JWT/TOTP/其他加密材料、对象存储和反代配置登记；保持现有密钥，避免会话失效和凭据不可解密。
- 本次不同时升级 PostgreSQL/Redis 大版本；检查新增环境变量、配置优先级、反代超时和媒体大小限制。
- `docker compose config --quiet` 检查配置；展开配置可能带密钥，不上传展开全文。确认 registry 拉取、容量、备份空间和恢复权限。
- 根据演练确定维护时长，至少预留完整恢复加冒烟时间。建议窗口上限 4 小时，最晚回滚决策时间 = 窗口结束减恢复耗时减验证余量，而非固定等到最后一小时。

### 7.2 维护窗口顺序

1. 反代进入维护模式，暂停新生成、充值和写操作；排空流式请求和媒体任务，记录未完成任务/订单。支付回调必须有持久接收/重试或事后主动查单方案，不能简单返回成功丢弃事件。
2. 停止所有应用实例、worker、cron 和其他业务写入者，保留数据库运行。记录停写时刻 T0 和最后订单/用量标记。
3. 做 T0 一致备份，导出迁移记录和逐用户账务基线；备份配置、挂载数据及对象存储版本。再次核实备份可读且恢复方法已演练。
4. 切到自建候选 digest，只启动一个应用实例执行迁移，观察迁移锁、错误与健康。不要用 `up -d` 无差别更新所有基础设施。
5. 检查迁移/账务、旧用户登录、旧 Key、国产模型、视频小额请求、支付沙箱/专用订单及核心页面。
6. 通过后开放白名单，再开放业务；记录 T1。观察 30–60 分钟并延续次日账务核查。

命令模板适用于 Linux Compose 主机，不在 PowerShell 里直接粘贴。先填实际部署目录、服务名和数据库参数；缺省示例服务名为 `sub2api`、`postgres`。数据库 dump 写在容器中再复制，避免 shell 管道损坏二进制：

```bash
cd /srv/58token
docker compose -f docker-compose.local.yml config --quiet
docker compose -f docker-compose.local.yml stop sub2api
# 其他 worker/定时写入者也须先停止
mkdir -p ./backup/20261019
chmod 700 ./backup/20261019
docker compose -f docker-compose.local.yml exec -T postgres sh -c \
  'pg_dump -U "$POSTGRES_USER" -d "$POSTGRES_DB" -Fc -f /tmp/58token-pre-upgrade.dump'
docker compose -f docker-compose.local.yml cp \
  postgres:/tmp/58token-pre-upgrade.dump ./backup/20261019/database.dump
docker compose -f docker-compose.local.yml exec -T postgres \
  pg_restore --list /tmp/58token-pre-upgrade.dump
# 单独更新 app 镜像配置为已验证 digest 后：
docker compose -f docker-compose.local.yml pull sub2api
docker compose -f docker-compose.local.yml up -d --no-deps sub2api
docker compose -f docker-compose.local.yml logs --tail=200 sub2api
```

检查每条命令退出码，备份失败即停止发布。不得对运行中的 PostgreSQL 数据目录直接打包当作一致备份；宿主机数据库或外置数据库需要对应替换备份命令。环境/数据库身份参数按实际值核对，禁止 `down -v` 或删除旧数据目录。

## 8. 回滚 Runbook 与数据边界

触发：迁移失败/锁表超预算、无法登录/调用、金额或权限错误、关键指标越阈值，或到达最晚回滚决策时间。

### 8.1 T1 开放业务前

保持维护 → 停所有新应用与 worker → 保存候选日志和失败库现场 → 恢复 T0 数据库到新库/新数据卷 → 配回旧数据库连接及旧 digest/配置 → 恢复必要的业务状态存储 → 冒烟及逐笔核账 → 开放流量。

推荐先恢复至新数据库/数据卷，保留现场；不要直接在故障库执行未经演练的 DROP/CLEAN。恢复脚本中的角色、扩展、owner、权限、连接参数均须演练。应用启动可能自动迁移，必须先完成旧配置与恢复库切换再启动旧版。

### 8.2 T1 开放业务后

不能盲目恢复 T0：会丢失 T0 后真实充值、订单、用量扣费及生成任务。

1. 再次维护并冻结写入，保存最新库、订单/用量/供应商任务清单和外部支付账单。
2. 若演练证明旧程序兼容新 schema，可仅回旧镜像，但必须满足旧后台任务不会破坏新数据的条件。
3. 若不兼容，优先修复前进；必须恢复 T0 时，先确定增量账务提取与幂等补录、支付查单和媒体任务补偿路径，由业务/运维共同确认后执行。
4. 对 T0 后供应商已发生的真实模型费用和用户消费逐项对账，不能因数据库恢复而重新发起生成。

Redis 分别识别会话/限流/缓存/队列：可重建缓存通常不恢复；旧快照中的待处理队列不能直接重放造成重复扣费。第二响应缓存按版本清理或命名空间隔离，确保不串用户/模型。对象存储保留版本和新增媒体，避免回滚删除有效结果。

恢复目标：维护停写到开放前的数据 RPO 为 0（仍须补外部回调）；开放后 RPO 由增量日志与对账能力决定，不承诺无损自动回滚。RTO 以真实恢复演练计时验收。

## 9. 风险登记

| 风险 | 等级 | 控制与停止条件 |
| --- | --- | --- |
| 独立历史/比较基线不明，漏掉初始定制 | 高 | 基线证据与功能清单；无法解释差异不得机械合并 |
| 同名迁移差异/上游覆盖定制结构 | 极高 | 全量迁移冲突表、生产副本与空库双路径；checksum 未解决阻断 |
| 新调度与旧重试/缓存相互放大 | 高 | 故障注入、重试总次数上限、单请求账务核对 |
| 支付退款/返利/余额语义改变 | 极高 | 沙箱及重复/乱序事件测试、逐笔账务；任何差异阻断 |
| 国产平台/视频接口支持范围误判 | 高 | 实际账号/套餐/客户端验证，未验证的组合不开通 |
| 旧视频任务账号绑定丢失或重复创建 | 高 | 排空、保留供应商任务 ID、幂等对账、超时不盲重试 |
| 官方镜像/自动更新覆盖定制 | 极高 | 自有 digest、更新来源审查、禁自动覆盖 |
| 新工具链或 CI 不兼容 | 中高 | Go 1.27.2 与目标前端工具链、CI/Docker 全同步 |
| 共享环境污染生产 | 极高 | 独立 DB/Redis/挂载/域名/凭据；测试前隔离出站 |
| 回滚丢失开放后账务 | 极高 | 分 T1 前后流程，保存增量、幂等补账，未演练不承诺无损 |

## 10. 排期、放行清单和待补参数

粗估：勘察/基线 1–2 人日，合并与迁移适配 4–8 人日，自动化/业务回归及恢复演练 3–5 人日，内部试运行 7–14 个自然日，正式维护 2–4 小时。没有共同基线或账务迁移冲突严重时，单独重估；不是已经测得的承诺。

上线前逐项签字：

- [ ] 生产实际 commit 与冻结 tag 一致；目标上游 SHA 固定。
- [ ] 全部二次开发功能已登记，并有保留/替代决策和用例。
- [ ] 所有迁移差异已审核，生产副本、空库、重复启动通过。
- [ ] 账务、权限、国产模型、视频 P0 通过；真实使用额度已核对。
- [ ] 自建镜像 digest、旧 digest、Compose 配置和构建记录齐全。
- [ ] 测试环境隔离，工具链与 CI 通过，试运行周期完成。
- [ ] 数据备份已实际恢复，RTO/RPO 和最晚回滚决策时刻已记录。
- [ ] 维护回调接收/补查、在途任务、开放后增量对账方案已演练。
- [ ] 开发/测试/运维/业务负责人确认发布执行单。

实施开始时继续采集以下参数即可，不影响当前方案编制：生产 Compose 文件与服务名/数据路径、PostgreSQL 版本/容量/最大表/日增量、备份方式和恢复耗时、是否有外部 worker/第二 Redis/S3、实际国产厂商与账号套餐、首批视频模型和使用客户端、支付通道与回调重试规则、测试期长度/预算、业务验收负责人。不要在文档或聊天中记录原始密钥。
