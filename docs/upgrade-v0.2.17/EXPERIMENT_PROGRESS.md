# 三方合并实验进度（2026-10-20）

## 引用与隔离

- 本地基线 tag 已改为 `58token-pre-upgrade`，仍指向 `f2f785fd6af5233f570841003fb3fb23d830d78b`；旧日期 tag 已删除，没有推送或改动远端 tag。
- 主集成分支：`codex/upgrade-sub2api-v0.2.17`。
- 实验 checkout：`.upgrade-experiment/`，实验分支 `codex/experiment-sub2api-v0.2.17`。
- 合成快照提交：`ff9925f15f6c1de7ff75cc71a3178b3962842dbc`，父提交为可信官方基线 B；其 tree `d6c38e2ef467a5ac2133032ef672c4713704f710` 与冻结生产提交 tree 完全一致。
- 目标为固定上游 `060d0160417782cd962a0c7f460f6cd4565d4916`，执行 `merge --no-commit --no-ff` 后处于未完成合并状态。未生成合并提交，未将结果合入 dev/集成分支。

## 已处理与验证

初始 239 个冲突，已标记解决 36 个，目前索引剩余 203 个。具体列表/未合并 index blob 见 `conflicts-current.json`、`conflicts-index.txt`；这些是记录生成时的索引状态，不自动随工作区修改更新。

- Ent schema 保留本项目限时优惠、用户权限、余额购买、帮助提示、支付精度字段，并加入上游币种、冻结余额等字段；全部生成代码重新生成。
- 检查合并 index 中本项目 205 个 SQL 文件 blob 与生产快照全部一致；121 个上游新增 SQL 仍需数据库语义和生产副本审查。
- 合并协议转换和 usage 字段，保留缓存写入/audio 细节、上游 reasoning 信息及新模型名；模型别名测试暴露并修正了 canonical Sonnet 4.5 thinking 的迁移映射。
- 保留响应缓存配置，合并新增功能配置；入账超时保留 30 秒，溢出策略采用上游 sync，避免 sample 丢账。配置测试通过。
- CORS 同时保留客户端追踪 ID 和上游 UI/Server-Timing 头；Redis 同时支持第二缓存 Redis、ACL username 和 Server-Timing hook。
- 部署保留品牌域名、响应缓存服务变量，采用上游目录权限与空管理员邮箱默认值。

已通过：`go test ./ent/...`、`./internal/pkg/apicompat`、`./internal/pkg/antigravity`、`./internal/pkg/openai`、`./internal/pkg/usagestats`、`./internal/domain`、`./migrations`、`./internal/config`。非数据库实际迁移验收。

尚未执行完整应用编译/全量测试、前端依赖安装/构建、Compose 启动、真实模型/支付或生产数据库操作。实验目录仍含大量冲突标记，当前不能运行应用。

## 已编辑但尚未标记解决

前端品牌资源、favicon/main/App 和 store 出口、HTTP upstream 测试 import、admin_auth 的权限上下文已做候选整合，仍等待相关模块类型/回归验证。不能以消除冲突标记替代语义验证。

## 需要确认的业务口径

`backend/internal/service/payment_stats.go` 以及相应测试/前端看板存在真实行为冲突：

- 现有定制：人民币和 USDT 使用配置汇率折算成人民币，提供统一总额、均额、用户排名及原币信息。
- 上游：金额为 CurrencyAmounts 字典，按币种分别汇总、均额和排名，避免混币统计，不提供同样的人民币折算总额。

建议保留人民币折算汇总，并增加上游按币种明细；使用明确的报表汇率标签，不改变实际扣费/支付币种。需同步设计兼容 API 字段和前端测试，不能直接替换旧字段。

确认此口径后继续整合支付/返利/订阅；其余模块按“保留定制并整合上游能力”执行，无需为每个文本冲突单独确认。最终合入 dev、推送、数据库/生产发布仍分别审核。
