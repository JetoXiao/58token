# ResponseCache 第二 Redis 生产部署指南

本文档用于说明在生产环境中为 AI ResponseCache / shadow 统计单独部署第二个 Redis 的推荐方案。

> 当前 dev 代码已接入独立 ResponseCache Redis 配置，并按 fail-open 方式工作：缓存 Redis 异常时请求继续走原上游链路，不应影响用户正常调用。

## 1. 为什么要部署第二个 Redis

现有 Redis 通常承担登录、鉴权、限流、Token 缓存、分布式锁、运维指标等业务功能。AI 完整响应缓存的 value 可能较大，读写频率也可能随请求量快速上升。如果直接复用业务 Redis，可能带来以下风险：

- 缓存响应占用大量内存，挤压业务 Redis 的 key。
- Redis 延迟升高，影响登录、鉴权、限流等核心链路。
- 缓存淘汰策略误删业务 key。
- 缓存请求占满连接池，影响现有功能。
- 故障排查时难以区分业务 Redis 压力和缓存 Redis 压力。

推荐拆分为：

```text
redis                 -> 现有业务 Redis：登录、限流、Token、锁、监控
redis-response-cache  -> 新增缓存 Redis：shadow 统计、exact response cache
```

## 2. 部署前确认

在服务器上进入项目部署目录，例如：

```bash
cd /path/to/useAiDoAnything/deploy
```

确认当前使用的 compose 文件。常见情况：

```bash
docker compose -f docker-compose.yml ps
```

如果你使用的是 `docker-compose.local.yml` 或 `docker-compose.standalone.yml`，后续步骤中的文件名需要替换为你实际使用的 compose 文件。

建议先备份当前部署文件：

```bash
cp docker-compose.yml docker-compose.yml.bak.$(date +%Y%m%d%H%M%S)
cp .env .env.bak.$(date +%Y%m%d%H%M%S)
```

## 3. 新增第二个 Redis 服务

编辑 `deploy/docker-compose.yml`，在 `services:` 下新增服务：

```yaml
  redis-response-cache:
    image: redis:8-alpine
    container_name: sub2api-redis-response-cache
    restart: unless-stopped
    ulimits:
      nofile:
        soft: 100000
        hard: 100000
    volumes:
      - redis_response_cache_data:/data
    command: >
      sh -c '
        redis-server
        --save ""
        --appendonly no
        --maxmemory ${RESPONSE_CACHE_REDIS_MAXMEMORY:-1gb}
        --maxmemory-policy ${RESPONSE_CACHE_REDIS_MAXMEMORY_POLICY:-allkeys-lru}
        ${RESPONSE_CACHE_REDIS_PASSWORD:+--requirepass "$RESPONSE_CACHE_REDIS_PASSWORD"}'
    environment:
      - TZ=${TZ:-Asia/Shanghai}
      - REDISCLI_AUTH=${RESPONSE_CACHE_REDIS_PASSWORD:-}
    networks:
      - sub2api-network
    healthcheck:
      test: ["CMD", "redis-cli", "ping"]
      interval: 10s
      timeout: 5s
      retries: 5
      start_period: 5s
```

然后在文件底部 `volumes:` 下新增：

```yaml
  redis_response_cache_data:
    driver: local
```

最终结构类似：

```yaml
services:
  app:
    ...

  redis:
    ...

  redis-response-cache:
    ...

volumes:
  sub2api_data:
    driver: local
  postgres_data:
    driver: local
  redis_data:
    driver: local
  redis_response_cache_data:
    driver: local
```

## 4. 配置环境变量

编辑 `deploy/.env`，新增以下配置：

```env
# =============================================================================
# ResponseCache Redis
# AI 响应缓存 / shadow 统计专用 Redis
# =============================================================================

RESPONSE_CACHE_REDIS_HOST=redis-response-cache
RESPONSE_CACHE_REDIS_PORT=6379
RESPONSE_CACHE_REDIS_PASSWORD=请换成强密码
RESPONSE_CACHE_REDIS_DB=0
RESPONSE_CACHE_REDIS_POOL_SIZE=128
RESPONSE_CACHE_REDIS_MIN_IDLE_CONNS=8
RESPONSE_CACHE_REDIS_MAXMEMORY=1gb
RESPONSE_CACHE_REDIS_MAXMEMORY_POLICY=allkeys-lru
RESPONSE_CACHE_REDIS_ENABLE_TLS=false
RESPONSE_CACHE_ENABLED=false
RESPONSE_CACHE_SHADOW_ENABLED=true
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
```

建议生产环境密码使用随机强密码，例如：

```bash
openssl rand -base64 32
```

注意：

- 不要复用业务 Redis 密码。
- 不要把缓存 Redis 端口暴露到公网。
- `RESPONSE_CACHE_REDIS_MAXMEMORY` 应按机器内存和请求规模调整。
- 如果只是 shadow 统计，`256mb` 通常也够用。
- 如果开始缓存完整响应，建议从 `1gb` 起步，并观察内存增长。

## 5. 后端服务需要增加的环境变量

`app` 服务需要读取上面的环境变量。可以在 `docker-compose.yml` 的后端服务 `environment:` 中加入：

```yaml
      - RESPONSE_CACHE_REDIS_HOST=${RESPONSE_CACHE_REDIS_HOST:-redis-response-cache}
      - RESPONSE_CACHE_REDIS_PORT=${RESPONSE_CACHE_REDIS_PORT:-6379}
      - RESPONSE_CACHE_REDIS_PASSWORD=${RESPONSE_CACHE_REDIS_PASSWORD:-}
      - RESPONSE_CACHE_REDIS_DB=${RESPONSE_CACHE_REDIS_DB:-0}
      - RESPONSE_CACHE_REDIS_POOL_SIZE=${RESPONSE_CACHE_REDIS_POOL_SIZE:-128}
      - RESPONSE_CACHE_REDIS_MIN_IDLE_CONNS=${RESPONSE_CACHE_REDIS_MIN_IDLE_CONNS:-8}
      - RESPONSE_CACHE_REDIS_ENABLE_TLS=${RESPONSE_CACHE_REDIS_ENABLE_TLS:-false}
      - RESPONSE_CACHE_ENABLED=${RESPONSE_CACHE_ENABLED:-false}
      - RESPONSE_CACHE_SHADOW_ENABLED=${RESPONSE_CACHE_SHADOW_ENABLED:-false}
      - RESPONSE_CACHE_TTL_SECONDS=${RESPONSE_CACHE_TTL_SECONDS:-300}
      - RESPONSE_CACHE_SHADOW_TTL_SECONDS=${RESPONSE_CACHE_SHADOW_TTL_SECONDS:-3600}
      - RESPONSE_CACHE_REDIS_TIMEOUT_MS=${RESPONSE_CACHE_REDIS_TIMEOUT_MS:-10}
      - RESPONSE_CACHE_MAX_BODY_BYTES=${RESPONSE_CACHE_MAX_BODY_BYTES:-65536}
      - RESPONSE_CACHE_MAX_VALUE_BYTES=${RESPONSE_CACHE_MAX_VALUE_BYTES:-1048576}
      - RESPONSE_CACHE_MIN_PROMPT_CHARS=${RESPONSE_CACHE_MIN_PROMPT_CHARS:-16}
      - RESPONSE_CACHE_MAX_PROMPT_CHARS=${RESPONSE_CACHE_MAX_PROMPT_CHARS:-12000}
      - RESPONSE_CACHE_SINGLEFLIGHT_ENABLED=${RESPONSE_CACHE_SINGLEFLIGHT_ENABLED:-true}
      - RESPONSE_CACHE_SINGLEFLIGHT_WAIT_TIMEOUT_MS=${RESPONSE_CACHE_SINGLEFLIGHT_WAIT_TIMEOUT_MS:-150}
      - RESPONSE_CACHE_PREFIX_CACHE_ENABLED=${RESPONSE_CACHE_PREFIX_CACHE_ENABLED:-false}
      - RESPONSE_CACHE_ALLOWED_API_KEY_ID_LIST=${RESPONSE_CACHE_ALLOWED_API_KEY_ID_LIST:-}
      - RESPONSE_CACHE_ALLOWED_GROUP_ID_LIST=${RESPONSE_CACHE_ALLOWED_GROUP_ID_LIST:-}
      - RESPONSE_CACHE_ALLOWED_MODEL_LIST=${RESPONSE_CACHE_ALLOWED_MODEL_LIST:-}
      - RESPONSE_CACHE_BYPASS_API_KEY_ID_LIST=${RESPONSE_CACHE_BYPASS_API_KEY_ID_LIST:-}
      - RESPONSE_CACHE_BYPASS_GROUP_ID_LIST=${RESPONSE_CACHE_BYPASS_GROUP_ID_LIST:-}
      - RESPONSE_CACHE_BYPASS_MODEL_LIST=${RESPONSE_CACHE_BYPASS_MODEL_LIST:-}
      - RESPONSE_CACHE_MONITOR_API_KEY_ID_LIST=${RESPONSE_CACHE_MONITOR_API_KEY_ID_LIST:-}
      - RESPONSE_CACHE_MONITOR_GROUP_ID_LIST=${RESPONSE_CACHE_MONITOR_GROUP_ID_LIST:-}
      - RESPONSE_CACHE_RECOMMENDATION_ENABLED=${RESPONSE_CACHE_RECOMMENDATION_ENABLED:-true}
      - RESPONSE_CACHE_RECOMMENDATION_WINDOW_HOURS=${RESPONSE_CACHE_RECOMMENDATION_WINDOW_HOURS:-24}
      - RESPONSE_CACHE_RECOMMENDATION_HIT_RATE_THRESHOLD=${RESPONSE_CACHE_RECOMMENDATION_HIT_RATE_THRESHOLD:-0.20}
      - RESPONSE_CACHE_RECOMMENDATION_MIN_CANDIDATES=${RESPONSE_CACHE_RECOMMENDATION_MIN_CANDIDATES:-150}
      - RESPONSE_CACHE_RECOMMENDATION_MIN_OBSERVED_HOURS=${RESPONSE_CACHE_RECOMMENDATION_MIN_OBSERVED_HOURS:-24}
      - RESPONSE_CACHE_RECOMMENDATION_MAX_SPIKE_RATIO=${RESPONSE_CACHE_RECOMMENDATION_MAX_SPIKE_RATIO:-5.0}
```

同时建议后端 ResponseCache 配置具备以下开关：

```env
RESPONSE_CACHE_ENABLED=false
RESPONSE_CACHE_SHADOW_ENABLED=true
RESPONSE_CACHE_TTL_SECONDS=300
RESPONSE_CACHE_MAX_BODY_BYTES=65536
RESPONSE_CACHE_MAX_VALUE_BYTES=1048576
RESPONSE_CACHE_REDIS_TIMEOUT_MS=10
RESPONSE_CACHE_MIN_PROMPT_CHARS=16
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
RESPONSE_CACHE_RECOMMENDATION_WINDOW_HOURS=24
RESPONSE_CACHE_RECOMMENDATION_HIT_RATE_THRESHOLD=0.20
RESPONSE_CACHE_RECOMMENDATION_MIN_CANDIDATES=150
RESPONSE_CACHE_RECOMMENDATION_MIN_OBSERVED_HOURS=24
RESPONSE_CACHE_RECOMMENDATION_MAX_SPIKE_RATIO=5.0
```

建议上线顺序：

```text
先 shadow，不返回缓存命中结果
观察命中率、延迟、Redis 内存、推荐接口结果
再对指定 API Key / 分组开启真实 exact cache
```

### 5.1 配置 exact cache 灰度范围

开启真实缓存时，建议不要全站直接打开。先打开总开关：

```env
RESPONSE_CACHE_ENABLED=true
```

然后指定允许命中真实缓存的 API Key ID、分组 ID 或模型：

```env
RESPONSE_CACHE_ALLOWED_API_KEY_ID_LIST=101,102
RESPONSE_CACHE_ALLOWED_GROUP_ID_LIST=8,9
RESPONSE_CACHE_ALLOWED_MODEL_LIST=gpt-4o-mini,gpt-4.1-mini
```

只要配置了任意 `ALLOWED_*`，exact cache 就只对这些范围生效。都为空时，`RESPONSE_CACHE_ENABLED=true` 会变成全局开启，所以生产不建议第一步这样做。

如果某些 API Key / 分组 / 模型必须永远绕过缓存，可以配置：

```env
RESPONSE_CACHE_BYPASS_API_KEY_ID_LIST=201,202
RESPONSE_CACHE_BYPASS_GROUP_ID_LIST=18
RESPONSE_CACHE_BYPASS_MODEL_LIST=o3,o4-mini
```

### 5.2 配置监控 Key 排除

如果你有平台自己的探活/预警 API Key，建议把这些 Key 的 ID 填入：

```env
RESPONSE_CACHE_MONITOR_API_KEY_ID_LIST=101,102
```

如果探活统一挂在某些分组，也可以填分组 ID：

```env
RESPONSE_CACHE_MONITOR_GROUP_ID_LIST=8,9
```

这些请求仍会被 shadow 观测，但会进入 `monitor_*` 计数，不参与“是否建议开启 exact cache”的命中率计算，避免探活脚本把命中率刷高。

### 5.2 查看 exact cache 开启建议

管理员接口：

```bash
curl -H "Authorization: Bearer <admin-token>" \
  "https://你的域名/api/v1/admin/ops/response-cache/recommendation"
```

默认规则：

```text
shadow 命中率连续 24 小时 >= 20%
且缓存候选请求数 >= 150
且小时级流量没有明显突刺
且不把监控 Key 计入业务候选
=> 返回 recommend_enable_exact_cache
=> 管理员确认后再开启 exact cache
```

快速验证接口逻辑时，可以临时缩短窗口：

```bash
curl -H "Authorization: Bearer <admin-token>" \
  "https://你的域名/api/v1/admin/ops/response-cache/recommendation?window_hours=1&min_observed_hours=1&min_candidates=10&hit_rate_threshold=0.2"
```

返回字段重点看：

```text
recommended: 是否建议开启
decision: recommend_enable_exact_cache / not_recommended / already_enabled
reasons: 未达标原因
total_candidates: 非监控候选请求数
shadow_hits: shadow 命中数
hit_rate: shadow 命中率
monitor_candidates: 监控/探活请求数
hours: 每小时明细
```

## 6. 启动第二 Redis

只启动新增 Redis：

```bash
docker compose -f docker-compose.yml up -d redis-response-cache
```

查看容器状态：

```bash
docker compose -f docker-compose.yml ps redis-response-cache
```

查看日志：

```bash
docker compose -f docker-compose.yml logs -f redis-response-cache
```

如果状态为 `healthy`，说明 Redis 已正常启动。

## 7. 验证连接

在 compose 网络内执行 ping：

```bash
docker compose -f docker-compose.yml exec redis-response-cache redis-cli ping
```

如果配置了密码但没有自动带入 `REDISCLI_AUTH`，可以手动执行：

```bash
docker compose -f docker-compose.yml exec redis-response-cache redis-cli -a "$RESPONSE_CACHE_REDIS_PASSWORD" ping
```

预期输出：

```text
PONG
```

测试写入和 TTL：

```bash
docker compose -f docker-compose.yml exec redis-response-cache redis-cli setex uado:rc:test 60 ok
docker compose -f docker-compose.yml exec redis-response-cache redis-cli get uado:rc:test
docker compose -f docker-compose.yml exec redis-response-cache redis-cli ttl uado:rc:test
```

预期：

```text
OK
ok
TTL 为 1 到 60 之间的数字
```

## 8. 推荐 Redis 策略

缓存 Redis 和业务 Redis 的目标不同。缓存 Redis 的数据允许丢失，建议策略是：

```text
appendonly no
save ""
maxmemory 512mb 或更高
maxmemory-policy allkeys-lru
短 TTL
不暴露公网端口
独立密码
独立连接池
Redis 异常时后端 fail-open，直接 bypass
```

不建议缓存 Redis 使用业务 Redis 的高持久化策略，因为完整响应缓存不需要强持久化，AOF 反而可能增加 IO 压力。

## 9. 监控指标

上线后建议定期查看：

```bash
docker compose -f docker-compose.yml exec redis-response-cache redis-cli info memory
docker compose -f docker-compose.yml exec redis-response-cache redis-cli info stats
docker compose -f docker-compose.yml exec redis-response-cache redis-cli info clients
```

重点关注：

```text
used_memory_human
maxmemory_human
evicted_keys
expired_keys
connected_clients
instantaneous_ops_per_sec
latency
```

如果出现以下情况，需要降低缓存范围或扩大 Redis 资源：

```text
used_memory 持续接近 maxmemory
evicted_keys 快速增长
connected_clients 接近上限
Redis 延迟升高
后端出现缓存 Redis timeout
```

## 10. 回滚方式

如果第二 Redis 导致异常，可以按以下顺序回滚。

先关闭后端缓存开关：

```env
RESPONSE_CACHE_ENABLED=false
RESPONSE_CACHE_SHADOW_ENABLED=false
```

重启后端服务：

```bash
docker compose -f docker-compose.yml up -d app
```

如果只是停掉缓存 Redis：

```bash
docker compose -f docker-compose.yml stop redis-response-cache
```

如果要完全删除缓存 Redis 容器和数据卷：

```bash
docker compose -f docker-compose.yml down
docker volume rm deploy_redis_response_cache_data
docker compose -f docker-compose.yml up -d
```

注意：`docker volume rm` 会删除缓存 Redis 数据。正常情况下这没有问题，因为 ResponseCache 数据应该是可丢失的。

## 11. 常见问题

### 11.1 可以只用同一个 Redis 的不同 DB 吗

可以，但只建议用于 shadow 阶段或小流量测试。

不同 DB 仍然共享同一份内存、连接、CPU、淘汰策略。完整响应缓存变大后，仍可能影响业务 Redis。

### 11.2 第二 Redis 要不要持久化

一般不需要。

ResponseCache 是性能优化，不是业务数据源。缓存丢失后最多回源上游，不应影响请求正确性。

### 11.3 Redis 挂了会不会影响用户请求

后端实现时必须保证不会影响。

推荐行为：

```text
Redis 正常   -> 执行 shadow / cache 查询
Redis 超时   -> 直接 bypass，继续请求上游
Redis 错误   -> 记录日志，继续请求上游
Redis 不可用 -> 自动降级为无缓存
```

### 11.4 缓存 Redis 应该设置多大

初始建议：

```text
shadow only：256mb
小规模 exact cache：512mb
中等规模 exact cache：1gb-4gb
```

真实值应根据：

```text
每分钟请求量
平均响应大小
TTL
缓存命中率
可接受的淘汰频率
```

估算公式：

```text
所需内存 ~= 每分钟可缓存请求数 * 平均响应大小 * TTL分钟数 * 1.3
```

例如：

```text
每分钟 1000 个可缓存响应
平均响应 20KB
TTL 5 分钟
内存 ~= 1000 * 20KB * 5 * 1.3 = 130MB
```

如果响应平均 100KB，同样请求量和 TTL 下约为 650MB。

## 12. 后续代码接入建议

后端实现 ResponseCache 时建议遵循：

- 独立 Redis client，不复用业务 Redis client。
- 独立连接池，避免缓存请求占满业务 Redis 连接。
- Redis 操作设置极短 timeout，例如 5-10ms。
- 默认 shadow，不默认全站真实缓存。
- 短句、探活词、生图、多模态、文件、工具调用、报错响应强制 bypass。
- 只缓存成功响应。
- 响应头标记：`x-uado-cache: hit|miss|bypass|shadow`。
- 记录 bypass reason，方便评估是否值得放开缓存。
- Redis 异常 fail-open，不能影响用户请求。

