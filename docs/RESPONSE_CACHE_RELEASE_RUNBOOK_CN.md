# ResponseCache 发布上线 Runbook

本文档用于把本地 `dev` 分支上的 ResponseCache 改造提交到 GitHub，并在生产环境发布、配置第二 Redis、开启 shadow 统计，最后按 API Key / 分组灰度开启真实 exact cache。

## 0. 上线目标

第一阶段只做 shadow 统计：

```text
不真实返回缓存
不改变用户答案
不隐藏上游故障
只统计候选请求、重复命中、推荐开启条件
```

第一阶段生产推荐：

```env
RESPONSE_CACHE_ENABLED=false
RESPONSE_CACHE_SHADOW_ENABLED=true
RESPONSE_CACHE_RECOMMENDATION_ENABLED=true
```

不要第一阶段直接开启：

```env
RESPONSE_CACHE_ENABLED=true
```

## 1. 本地提交到 GitHub dev 分支

进入本地仓库：

```powershell
cd E:\software\python\pythonfile\useaifor\useAiDoAnything
git branch --show-current
```

确认当前分支是：

```text
dev
```

同步远端 dev：

```powershell
git fetch origin
git pull --ff-only origin dev
```

检查工作区：

```powershell
git status --short --untracked-files=all
```

确认不要提交这些本地文件：

```text
deploy/.env
deploy/redis_response_cache_data/
scripts/local-test-response-cache.ps1
```

这些文件已被 `.gitignore` / `deploy/.gitignore` 忽略。再次确认：

```powershell
git check-ignore -v deploy/.env
git check-ignore -v deploy/redis_response_cache_data/dump.rdb
git check-ignore -v scripts/local-test-response-cache.ps1
```

提交前检查：

```powershell
git diff --check
```

运行测试：

```powershell
$env:Path = [Environment]::GetEnvironmentVariable('Path','Machine') + ';' + [Environment]::GetEnvironmentVariable('Path','User') + ';' + $env:Path

docker run --rm -v "${PWD}:/src" -w /src/backend golang:1.26.3-alpine go test ./internal/config -run TestLoadResponseCacheListsFromEnv -count=1

docker run --rm -v "${PWD}:/src" -w /src/backend golang:1.26.3-alpine go test ./internal/service -run TestResponseCache -count=1

pnpm.cmd --dir frontend typecheck
```

推荐使用显式 `git add`，避免误提交其他本地文件：

```powershell
git add .gitignore `
  backend/cmd/server/wire_gen.go `
  backend/internal/config/config.go `
  backend/internal/config/config_test.go `
  backend/internal/handler/admin/ops_handler.go `
  backend/internal/handler/gateway_handler.go `
  backend/internal/handler/gateway_handler_chat_completions.go `
  backend/internal/handler/openai_chat_completions.go `
  backend/internal/handler/openai_gateway_handler.go `
  backend/internal/handler/wire.go `
  backend/internal/handler/response_cache_helper.go `
  backend/internal/repository/redis.go `
  backend/internal/server/routes/admin.go `
  backend/internal/service/openai_gateway_service.go `
  backend/internal/service/response_cache.go `
  backend/internal/service/response_cache_test.go `
  deploy/.env.example `
  deploy/.gitignore `
  deploy/config.example.yaml `
  deploy/docker-compose.dev.yml `
  deploy/docker-compose.yml `
  docs/RESPONSE_CACHE_SECOND_REDIS_DEPLOY_CN.md `
  docs/RESPONSE_CACHE_RELEASE_RUNBOOK_CN.md `
  frontend/pnpm-lock.yaml
```

确认暂存区：

```powershell
git status --short
git diff --cached --stat
```

提交并推送：

```powershell
git commit -m "feat: add response cache shadow metrics"
git push origin dev
```

## 2. 生产服务器拉取 dev 代码

以下命令在生产服务器执行。

进入生产项目目录：

```bash
cd /path/to/useAiDoAnything
```

确认远端和分支：

```bash
git remote -v
git branch --show-current
```

切到 dev：

```bash
git fetch origin
git checkout dev
git pull --ff-only origin dev
```

如果生产有本地未提交改动，先不要强拉，先查看：

```bash
git status --short
git diff
```

## 3. 生产环境配置第二 Redis

如果生产使用本项目的 `deploy/docker-compose.yml`，提交后的 compose 已包含：

```text
redis-response-cache
redis_response_cache_data
```

进入部署目录：

```bash
cd /path/to/useAiDoAnything/deploy
```

生成第二 Redis 密码：

```bash
openssl rand -base64 32
```

编辑生产 `.env`：

```bash
cp -n .env.example .env
nano .env
```

如果你已有生产 `.env`，不要覆盖，直接编辑原 `.env`，加入 ResponseCache 配置。

第一阶段推荐配置：

```env
RESPONSE_CACHE_ENABLED=false
RESPONSE_CACHE_SHADOW_ENABLED=true

RESPONSE_CACHE_REDIS_PASSWORD=替换为上面生成的强密码
RESPONSE_CACHE_REDIS_DB=0
RESPONSE_CACHE_REDIS_POOL_SIZE=128
RESPONSE_CACHE_REDIS_MIN_IDLE_CONNS=8
RESPONSE_CACHE_REDIS_MAXMEMORY=1gb

RESPONSE_CACHE_TTL_SECONDS=300
RESPONSE_CACHE_SHADOW_TTL_SECONDS=3600
RESPONSE_CACHE_REDIS_TIMEOUT_MS=10
RESPONSE_CACHE_MAX_BODY_BYTES=65536
RESPONSE_CACHE_MAX_VALUE_BYTES=1048576
RESPONSE_CACHE_MIN_PROMPT_CHARS=16
RESPONSE_CACHE_MAX_PROMPT_CHARS=12000

RESPONSE_CACHE_SINGLEFLIGHT_ENABLED=true
RESPONSE_CACHE_SINGLEFLIGHT_WAIT_TIMEOUT_MS=150
RESPONSE_CACHE_PREFIX_CACHE_ENABLED=false

RESPONSE_CACHE_ALLOWED_API_KEY_ID_LIST=
RESPONSE_CACHE_ALLOWED_GROUP_ID_LIST=
RESPONSE_CACHE_ALLOWED_MODEL_LIST=
RESPONSE_CACHE_BYPASS_API_KEY_ID_LIST=
RESPONSE_CACHE_BYPASS_GROUP_ID_LIST=
RESPONSE_CACHE_BYPASS_MODEL_LIST=
RESPONSE_CACHE_MONITOR_API_KEY_ID_LIST=
RESPONSE_CACHE_MONITOR_GROUP_ID_LIST=

RESPONSE_CACHE_RECOMMENDATION_ENABLED=true
RESPONSE_CACHE_RECOMMENDATION_WINDOW_HOURS=72
RESPONSE_CACHE_RECOMMENDATION_HIT_RATE_THRESHOLD=0.20
RESPONSE_CACHE_RECOMMENDATION_MIN_CANDIDATES=150
RESPONSE_CACHE_RECOMMENDATION_MIN_OBSERVED_HOURS=24
RESPONSE_CACHE_RECOMMENDATION_MAX_SPIKE_RATIO=5.0
```

## 4. 代码如何连接到第二 Redis

Docker Compose 部署时，后端服务通过 Docker 内部服务名连接第二 Redis：

```env
RESPONSE_CACHE_REDIS_HOST=redis-response-cache
RESPONSE_CACHE_REDIS_PORT=6379
RESPONSE_CACHE_REDIS_PASSWORD=${RESPONSE_CACHE_REDIS_PASSWORD}
RESPONSE_CACHE_REDIS_DB=0
```

这些变量已经在 `deploy/docker-compose.yml` 的后端 `environment:` 中配置，通常你只需要在生产 `.env` 里填：

```env
RESPONSE_CACHE_REDIS_PASSWORD=你的强密码
```

如果你不是 Docker Compose 部署，而是外部独立 Redis，需要在应用环境变量或配置文件中显式配置：

```env
RESPONSE_CACHE_REDIS_HOST=10.0.0.12
RESPONSE_CACHE_REDIS_PORT=6379
RESPONSE_CACHE_REDIS_PASSWORD=你的强密码
RESPONSE_CACHE_REDIS_DB=0
```

如果第二 Redis 使用 TLS：

```env
RESPONSE_CACHE_REDIS_ENABLE_TLS=true
```

## 5. 发布生产

在 `deploy` 目录执行：

```bash
docker compose pull
docker compose up -d --build
```

查看容器：

```bash
docker compose ps
```

应该看到：

```text
sub2api
postgres
redis
redis-response-cache
```

查看应用环境变量是否进入容器：

```bash
docker compose exec sub2api /bin/sh -lc 'env | sort | grep RESPONSE_CACHE'
```

查看第二 Redis 健康状态：

```bash
docker compose exec redis-response-cache redis-cli ping
```

如果设置了密码，使用：

```bash
docker compose exec redis-response-cache redis-cli -a "$RESPONSE_CACHE_REDIS_PASSWORD" ping
```

查看日志：

```bash
docker compose logs -f --tail=200 sub2api
docker compose logs -f --tail=100 redis-response-cache
```

## 6. 上线后验证 shadow 统计

第一阶段开启：

```env
RESPONSE_CACHE_ENABLED=false
RESPONSE_CACHE_SHADOW_ENABLED=true
```

请求不会返回缓存旧答案，只会在响应头看到：

```text
x-uado-cache: shadow
```

短 prompt、`hi`、`ping`、图片、工具调用、过大请求会绕过，可能看到：

```text
x-uado-cache: bypass; reason=prompt_too_short
x-uado-cache: bypass; reason=unsupported_request
x-uado-cache: bypass; reason=body_too_large
```

后台推荐接口：

```bash
curl -s \
  -H "Authorization: Bearer 管理员JWT" \
  "https://你的域名/api/v1/admin/ops/response-cache/recommendation"
```

重点看：

```json
{
  "total_candidates": 150,
  "shadow_hits": 35,
  "hit_rate": 0.23,
  "recommended": true,
  "decision": "recommend_enable_exact_cache"
}
```

判断规则：

```text
total_candidates >= 150
hit_rate >= 0.20
observed_hours >= 24
recommended = true
```

满足后仍然不会自动开启真实缓存，需要管理员手动配置。

## 7. 按 API Key / 分组灰度开启 exact cache

只有当 shadow 统计证明值得开，再进入这一阶段。

按 API Key 开启：

```env
RESPONSE_CACHE_ENABLED=true
RESPONSE_CACHE_ALLOWED_API_KEY_ID_LIST=101,102
```

按分组开启：

```env
RESPONSE_CACHE_ENABLED=true
RESPONSE_CACHE_ALLOWED_GROUP_ID_LIST=8,9
```

按模型开启：

```env
RESPONSE_CACHE_ENABLED=true
RESPONSE_CACHE_ALLOWED_MODEL_LIST=gpt-4o-mini,gpt-4.1-mini
```

可以组合：

```env
RESPONSE_CACHE_ENABLED=true
RESPONSE_CACHE_ALLOWED_GROUP_ID_LIST=8
RESPONSE_CACHE_ALLOWED_MODEL_LIST=gpt-4o-mini
```

注意：

```text
只要任意 ALLOWED_* 有值，exact cache 只对这些范围生效。
如果 ALLOWED_* 全为空，同时 RESPONSE_CACHE_ENABLED=true，就是全站开启 exact cache。
生产不建议第一步全站开启。
```

强制绕过缓存：

```env
RESPONSE_CACHE_BYPASS_API_KEY_ID_LIST=201,202
RESPONSE_CACHE_BYPASS_GROUP_ID_LIST=18
RESPONSE_CACHE_BYPASS_MODEL_LIST=o3,o4-mini
```

监控/探活 Key 单独统计，不参与推荐：

```env
RESPONSE_CACHE_MONITOR_API_KEY_ID_LIST=301,302
RESPONSE_CACHE_MONITOR_GROUP_ID_LIST=28
```

修改 `.env` 后重启：

```bash
docker compose up -d --build
```

验证 exact cache：

```text
第一次相同请求：x-uado-cache: miss
第二次相同请求：x-uado-cache: hit
```

## 8. 配置字段解释

### 总开关

```env
RESPONSE_CACHE_ENABLED=false
```

是否真实返回 exact cache。`true` 才会可能返回缓存答案。

```env
RESPONSE_CACHE_SHADOW_ENABLED=true
```

是否只统计重复率。不会返回缓存答案。

### 第二 Redis

```env
RESPONSE_CACHE_REDIS_PASSWORD=
```

第二 Redis 密码。生产必须设置强密码。

```env
RESPONSE_CACHE_REDIS_DB=0
```

Redis DB 编号。独立 Redis 一般用 0。

```env
RESPONSE_CACHE_REDIS_POOL_SIZE=128
RESPONSE_CACHE_REDIS_MIN_IDLE_CONNS=8
```

连接池大小和最小空闲连接数。

```env
RESPONSE_CACHE_REDIS_MAXMEMORY=1gb
```

第二 Redis 最大内存。Compose 中使用 `allkeys-lru`，内存满后会自动淘汰旧缓存。

```env
RESPONSE_CACHE_REDIS_TIMEOUT_MS=10
```

单次 Redis 操作超时。保持很小，避免 Redis 慢时拖慢用户请求。

### 缓存内容限制

```env
RESPONSE_CACHE_TTL_SECONDS=300
```

真实 exact cache 的 TTL。默认 5 分钟。

```env
RESPONSE_CACHE_SHADOW_TTL_SECONDS=3600
```

shadow 去重窗口。默认 1 小时。

```env
RESPONSE_CACHE_MAX_BODY_BYTES=65536
```

超过该请求体大小直接 bypass，避免大请求深度规范化拖慢请求。

```env
RESPONSE_CACHE_MAX_VALUE_BYTES=1048576
```

超过该响应大小不写入缓存。

```env
RESPONSE_CACHE_MIN_PROMPT_CHARS=16
```

短 prompt 不进入缓存统计和 exact cache，用于避免 `hi`、`ping`、探活被缓存遮蔽。

```env
RESPONSE_CACHE_MAX_PROMPT_CHARS=12000
```

超长 prompt 不进入缓存，避免规范化和存储成本过高。

### 并发去重

```env
RESPONSE_CACHE_SINGLEFLIGHT_ENABLED=true
RESPONSE_CACHE_SINGLEFLIGHT_WAIT_TIMEOUT_MS=150
```

exact cache 开启后，相同 key 的并发请求会短暂等待，减少缓存击穿。

### Prefix Cache

```env
RESPONSE_CACHE_PREFIX_CACHE_ENABLED=false
```

是否启用 prompt cache key 注入。第一阶段建议保持 false。

### 灰度范围

```env
RESPONSE_CACHE_ALLOWED_API_KEY_ID_LIST=
RESPONSE_CACHE_ALLOWED_GROUP_ID_LIST=
RESPONSE_CACHE_ALLOWED_MODEL_LIST=
```

exact cache 允许范围。逗号分隔。

```env
RESPONSE_CACHE_BYPASS_API_KEY_ID_LIST=
RESPONSE_CACHE_BYPASS_GROUP_ID_LIST=
RESPONSE_CACHE_BYPASS_MODEL_LIST=
```

强制绕过缓存范围。逗号分隔。

```env
RESPONSE_CACHE_MONITOR_API_KEY_ID_LIST=
RESPONSE_CACHE_MONITOR_GROUP_ID_LIST=
```

监控/探活 Key 或分组。单独统计，不参与开启推荐。

### 推荐阈值

```env
RESPONSE_CACHE_RECOMMENDATION_ENABLED=true
```

是否开启推荐接口。

```env
RESPONSE_CACHE_RECOMMENDATION_WINDOW_HOURS=72
```

统计最近多少小时。

```env
RESPONSE_CACHE_RECOMMENDATION_MIN_CANDIDATES=150
```

候选请求数达到多少才给推荐。

```env
RESPONSE_CACHE_RECOMMENDATION_MIN_OBSERVED_HOURS=24
```

至少观察多少小时。

```env
RESPONSE_CACHE_RECOMMENDATION_HIT_RATE_THRESHOLD=0.20
```

命中率阈值。`0.20` 表示 20%。

```env
RESPONSE_CACHE_RECOMMENDATION_MAX_SPIKE_RATIO=5.0
```

异常尖峰保护。某小时请求量异常突增时避免误判。

## 9. 回滚

最快回滚方式：

```env
RESPONSE_CACHE_ENABLED=false
RESPONSE_CACHE_SHADOW_ENABLED=false
RESPONSE_CACHE_RECOMMENDATION_ENABLED=false
```

重启：

```bash
docker compose up -d
```

如果要停掉第二 Redis：

```bash
docker compose stop redis-response-cache
```

如果要回滚代码：

```bash
git log --oneline -5
git revert <commit>
docker compose up -d --build
```

不要使用 `git reset --hard` 回滚生产，除非已经确认不会丢失本地配置或热修复。

## 10. 上线检查清单

```text
[ ] dev 分支已推送 GitHub
[ ] 生产已 git pull origin dev
[ ] 生产 .env 已配置 RESPONSE_CACHE_REDIS_PASSWORD
[ ] RESPONSE_CACHE_ENABLED=false
[ ] RESPONSE_CACHE_SHADOW_ENABLED=true
[ ] redis-response-cache 容器 healthy
[ ] sub2api 容器 healthy
[ ] x-uado-cache 能看到 shadow 或 bypass
[ ] 推荐接口能返回 total_candidates / hit_rate
[ ] 观察 24-72 小时后再决定是否开启 exact cache
```
