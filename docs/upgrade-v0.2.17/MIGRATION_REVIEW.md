# 迁移逐文件审查清单

自动生成：`python tools/audit_sub2api_upgrade.py`。哈希按 UTF-8 内容 trim 后计算，与当前 runner 规则对照。生产数据库记录尚未获取，源码哈希不代表生产已应用状态。

“注释差异”是词法提示；不是 SQL 语义等价证明。下表未核实生产 schema、锁时长、执行结果。

| 文件 | 分类 | 建议处理 |
| --- | --- | --- |
| `001_init.sql` | different | 保留本项目原字节；目标仅注释差异提示 |
| `002_account_type_migration.sql` | different | 保留本项目原字节；目标仅注释差异提示 |
| `003_subscription.sql` | different | 保留本项目原字节；目标仅注释差异提示 |
| `137_add_request_params_to_request_logs.sql` | local-only-preserve | 保留已发布定制，审查上游结构交互 |
| `142_payment_order_pay_amount_precision.sql` | local-only-preserve | 保留已发布定制，审查上游结构交互 |
| `142_user_platform_quotas.sql` | target-only-review | 审查依赖/锁/覆盖价格映射/破坏性 DDL，生产副本验证 |
| `143_group_models_list_config.sql` | target-only-review | 审查依赖/锁/覆盖价格映射/破坏性 DDL，生产副本验证 |
| `143_model_marketplace.sql` | local-only-preserve | 保留已发布定制，审查上游结构交互 |
| `144_add_claude_opus_48_marketplace.sql` | local-only-preserve | 保留已发布定制，审查上游结构交互 |
| `144_add_opus48_to_model_mapping.sql` | target-only-review | 审查依赖/锁/覆盖价格映射/破坏性 DDL，生产副本验证 |
| `145_deleted_api_key_audit.sql` | target-only-review | 审查依赖/锁/覆盖价格映射/破坏性 DDL，生产副本验证 |
| `145_ops_metrics_ttft_sample_count.sql` | target-only-review | 审查依赖/锁/覆盖价格映射/破坏性 DDL，生产副本验证 |
| `145_update_marketplace_price_display.sql` | local-only-preserve | 保留已发布定制，审查上游结构交互 |
| `146_affiliate_partner_program.sql` | local-only-preserve | 保留已发布定制，审查上游结构交互 |
| `147_normalize_usdt_cny_exchange_rate_default.sql` | local-only-preserve | 保留已发布定制，审查上游结构交互 |
| `147_ops_error_log_api_key_prefix.sql` | target-only-review | 审查依赖/锁/覆盖价格映射/破坏性 DDL，生产副本验证 |
| `148_add_ops_error_logs_user_time_index_notx.sql` | target-only-review | 审查依赖/锁/覆盖价格映射/破坏性 DDL，生产副本验证 |
| `148_disable_retired_openai_marketplace_models.sql` | local-only-preserve | 保留已发布定制，审查上游结构交互 |
| `149_add_marketing_nav_items_setting.sql` | local-only-preserve | 保留已发布定制，审查上游结构交互 |
| `149_proxy_expiry_fallback.sql` | target-only-review | 审查依赖/锁/覆盖价格映射/破坏性 DDL，生产副本验证 |
| `150_account_group_scheduler_indexes_notx.sql` | target-only-review | 审查依赖/锁/覆盖价格映射/破坏性 DDL，生产副本验证 |
| `150_add_affiliate_partner_tiers_setting.sql` | local-only-preserve | 保留已发布定制，审查上游结构交互 |
| `151_account_autopause_expiry_index_notx.sql` | target-only-review | 审查依赖/锁/覆盖价格映射/破坏性 DDL，生产副本验证 |
| `151_add_claude_fable_5_marketplace.sql` | local-only-preserve | 保留已发布定制，审查上游结构交互 |
| `151_channel_monitor_jitter.sql` | target-only-review | 审查依赖/锁/覆盖价格映射/破坏性 DDL，生产副本验证 |
| `152_scheduler_outbox_dedup_key.sql` | target-only-review | 审查依赖/锁/覆盖价格映射/破坏性 DDL，生产副本验证 |
| `152_subscription_plan_limited_offer.sql` | local-only-preserve | 保留已发布定制，审查上游结构交互 |
| `153_add_sub_admin_menu_permissions.sql` | local-only-preserve | 保留已发布定制，审查上游结构交互 |
| `153_scheduler_outbox_pending_dedup_key_index_notx.sql` | target-only-review | 审查依赖/锁/覆盖价格映射/破坏性 DDL，生产副本验证 |
| `154_account_spark_shadow.sql` | target-only-review | 审查依赖/锁/覆盖价格映射/破坏性 DDL，生产副本验证 |
| `154_add_ops_system_logs_api_key_id.sql` | target-only-review | 审查依赖/锁/覆盖价格映射/破坏性 DDL，生产副本验证 |
| `154_free_quota_trial_cards.sql` | local-only-preserve | 保留已发布定制，审查上游结构交互 |
| `154a_account_spark_shadow_indexes_notx.sql` | target-only-review | 审查依赖/锁/覆盖价格映射/破坏性 DDL，生产副本验证 |
| `155_add_ops_system_logs_api_key_id_index_notx.sql` | target-only-review | 审查依赖/锁/覆盖价格映射/破坏性 DDL，生产副本验证 |
| `155_add_support_contact_config.sql` | local-only-preserve | 保留已发布定制，审查上游结构交互 |
| `156_content_moderation_matched_keyword.sql` | target-only-review | 审查依赖/锁/覆盖价格映射/破坏性 DDL，生产副本验证 |
| `156_redact_request_params_previews.sql` | local-only-preserve | 保留已发布定制，审查上游结构交互 |
| `157_add_balance_subscription_purchase_user_flag.sql` | local-only-preserve | 保留已发布定制，审查上游结构交互 |
| `157_user_platform_quotas_add_grok.sql` | target-only-review | 审查依赖/锁/覆盖价格映射/破坏性 DDL，生产副本验证 |
| `158_add_group_peak_rate_multiplier.sql` | target-only-review | 审查依赖/锁/覆盖价格映射/破坏性 DDL，生产副本验证 |
| `158_affiliate_settlements.sql` | local-only-preserve | 保留已发布定制，审查上游结构交互 |
| `158_enable_grok_media_generation_groups.sql` | target-only-review | 审查依赖/锁/覆盖价格映射/破坏性 DDL，生产副本验证 |
| `159_add_help_center.sql` | local-only-preserve | 保留已发布定制，审查上游结构交互 |
| `159_batch_image_foundation.sql` | target-only-review | 审查依赖/锁/覆盖价格映射/破坏性 DDL，生产副本验证 |
| `160_add_gpt56_and_claude_sonnet5_marketplace.sql` | local-only-preserve | 保留已发布定制，审查上游结构交互 |
| `160_add_user_frozen_balance.sql` | target-only-review | 审查依赖/锁/覆盖价格映射/破坏性 DDL，生产副本验证 |
| `160_batch_image_provider_refs.sql` | target-only-review | 审查依赖/锁/覆盖价格映射/破坏性 DDL，生产副本验证 |
| `161_add_visitor_analytics.sql` | local-only-preserve | 保留已发布定制，审查上游结构交互 |
| `161_batch_image_pricing_snapshot.sql` | target-only-review | 审查依赖/锁/覆盖价格映射/破坏性 DDL，生产副本验证 |
| `162_add_group_batch_image_generation_gate.sql` | target-only-review | 审查依赖/锁/覆盖价格映射/破坏性 DDL，生产副本验证 |
| `162_seed_visitor_channels.sql` | local-only-preserve | 保留已发布定制，审查上游结构交互 |
| `163_add_download_resources.sql` | local-only-preserve | 保留已发布定制，审查上游结构交互 |
| `163_batch_image_default_discount_and_hold_ratio.sql` | target-only-review | 审查依赖/锁/覆盖价格映射/破坏性 DDL，生产副本验证 |
| `164_batch_image_download_and_user_delete.sql` | target-only-review | 审查依赖/锁/覆盖价格映射/破坏性 DDL，生产副本验证 |
| `164_track_authenticated_visitors_and_downloads.sql` | local-only-preserve | 保留已发布定制，审查上游结构交互 |
| `165_hide_pre_upstream_batch_image_failures.sql` | target-only-review | 审查依赖/锁/覆盖价格映射/破坏性 DDL，生产副本验证 |
| `165_normalize_claude_marketplace_model_names.sql` | local-only-preserve | 保留已发布定制，审查上游结构交互 |
| `166_backfill_help_center_domain.sql` | local-only-preserve | 保留已发布定制，审查上游结构交互 |
| `166_batch_image_task_name.sql` | target-only-review | 审查依赖/锁/覆盖价格映射/破坏性 DDL，生产副本验证 |
| `167_clear_auto_batch_image_task_names.sql` | target-only-review | 审查依赖/锁/覆盖价格映射/破坏性 DDL，生产副本验证 |
| `167_model_marketplace_maintenance.sql` | local-only-preserve | 保留已发布定制，审查上游结构交互 |
| `168_model_marketplace_virtual_configs.sql` | local-only-preserve | 保留已发布定制，审查上游结构交互 |
| `168_restore_empty_batch_image_task_names.sql` | target-only-review | 审查依赖/锁/覆盖价格映射/破坏性 DDL，生产副本验证 |
| `169_batch_image_parent_batch.sql` | target-only-review | 审查依赖/锁/覆盖价格映射/破坏性 DDL，生产副本验证 |
| `169_model_marketplace_capability_tags.sql` | local-only-preserve | 保留已发布定制，审查上游结构交互 |
| `170_add_grok_video_pricing_controls.sql` | target-only-review | 审查依赖/锁/覆盖价格映射/破坏性 DDL，生产副本验证 |
| `170_sync_openai_marketplace_official_prices.sql` | local-only-preserve | 保留已发布定制，审查上游结构交互 |
| `171_allow_video_usage_without_image_size.sql` | target-only-review | 审查依赖/锁/覆盖价格映射/破坏性 DDL，生产副本验证 |
| `172_composite_model_routes.sql` | target-only-review | 审查依赖/锁/覆盖价格映射/破坏性 DDL，生产副本验证 |
| `172_video_per_second_billing_metadata.sql` | target-only-review | 审查依赖/锁/覆盖价格映射/破坏性 DDL，生产副本验证 |
| `173_allow_cyber_blocked_usage_request_type.sql` | target-only-review | 审查依赖/锁/覆盖价格映射/破坏性 DDL，生产副本验证 |
| `174_add_usage_log_long_context_billing.sql` | target-only-review | 审查依赖/锁/覆盖价格映射/破坏性 DDL，生产副本验证 |
| `174_add_usage_logs_api_key_latest_ip_index_notx.sql` | target-only-review | 审查依赖/锁/覆盖价格映射/破坏性 DDL，生产副本验证 |
| `174_group_web_search_price_per_call.sql` | target-only-review | 审查依赖/锁/覆盖价格映射/破坏性 DDL，生产副本验证 |
| `175_add_ops_system_logs_host.sql` | target-only-review | 审查依赖/锁/覆盖价格映射/破坏性 DDL，生产副本验证 |
| `175_default_openai_long_context_billing.sql` | target-only-review | 审查依赖/锁/覆盖价格映射/破坏性 DDL，生产副本验证 |
| `175a_add_ops_system_logs_host_index_notx.sql` | target-only-review | 审查依赖/锁/覆盖价格映射/破坏性 DDL，生产副本验证 |
| `176_channel_monitor_grok_provider.sql` | target-only-review | 审查依赖/锁/覆盖价格映射/破坏性 DDL，生产副本验证 |
| `177_add_subscription_plan_currency.sql` | target-only-review | 审查依赖/锁/覆盖价格映射/破坏性 DDL，生产副本验证 |
| `178_channel_image_input_price.sql` | target-only-review | 审查依赖/锁/覆盖价格映射/破坏性 DDL，生产副本验证 |
| `179_usage_log_image_input_tokens.sql` | target-only-review | 审查依赖/锁/覆盖价格映射/破坏性 DDL，生产副本验证 |
| `180_audit_logs.sql` | target-only-review | 审查依赖/锁/覆盖价格映射/破坏性 DDL，生产副本验证 |
| `181_group_duplicate_operation_id.sql` | target-only-review | 审查依赖/锁/覆盖价格映射/破坏性 DDL，生产副本验证 |
| `181_prompt_audit.sql` | target-only-review | 审查依赖/锁/覆盖价格映射/破坏性 DDL，生产副本验证 |
| `182_prompt_audit_full_prompt.sql` | target-only-review | 审查依赖/锁/覆盖价格映射/破坏性 DDL，生产副本验证 |
| `183_ops_ingress_reject_aggregates.sql` | target-only-review | 审查依赖/锁/覆盖价格映射/破坏性 DDL，生产副本验证 |
| `184_auth_cache_invalidation_outbox.sql` | target-only-review | 审查依赖/锁/覆盖价格映射/破坏性 DDL，生产副本验证 |
| `185_group_reasoning_effort_policy.sql` | target-only-review | 审查依赖/锁/覆盖价格映射/破坏性 DDL，生产副本验证 |
| `186_alipay_mobile_precreate_deep_link.sql` | target-only-review | 审查依赖/锁/覆盖价格映射/破坏性 DDL，生产副本验证 |
| `186_group_auth_cache_image_generation.sql` | target-only-review | 审查依赖/锁/覆盖价格映射/破坏性 DDL，生产副本验证 |
| `187_add_usage_log_session_id.sql` | target-only-review | 审查依赖/锁/覆盖价格映射/破坏性 DDL，生产副本验证 |
| `188_allow_live_usage_request_type.sql` | target-only-review | 审查依赖/锁/覆盖价格映射/破坏性 DDL，生产副本验证 |
| `189_add_group_allow_live.sql` | target-only-review | 审查依赖/锁/覆盖价格映射/破坏性 DDL，生产副本验证 |
| `190_add_users_email_alias_dedup_index_notx.sql` | target-only-review | 审查依赖/锁/覆盖价格映射/破坏性 DDL，生产副本验证 |
| `191_passkey_credentials.sql` | target-only-review | 审查依赖/锁/覆盖价格映射/破坏性 DDL，生产副本验证 |
| `192_group_profit_control.sql` | target-only-review | 审查依赖/锁/覆盖价格映射/破坏性 DDL，生产副本验证 |
| `193_group_profit_control_auth_cache_invalidation.sql` | target-only-review | 审查依赖/锁/覆盖价格映射/破坏性 DDL，生产副本验证 |
| `194_add_usage_log_upstream_response_model.sql` | target-only-review | 审查依赖/锁/覆盖价格映射/破坏性 DDL，生产副本验证 |
| `194_channel_monitor_v2.sql` | target-only-review | 审查依赖/锁/覆盖价格映射/破坏性 DDL，生产副本验证 |
| `195_add_usage_log_upstream_model_mismatch_index_notx.sql` | target-only-review | 审查依赖/锁/覆盖价格映射/破坏性 DDL，生产副本验证 |
| `195_channel_monitor_mode.sql` | target-only-review | 审查依赖/锁/覆盖价格映射/破坏性 DDL，生产副本验证 |
| `196_channel_monitor_v2_ignored_error_categories.sql` | target-only-review | 审查依赖/锁/覆盖价格映射/破坏性 DDL，生产副本验证 |
| `197_channel_monitor_v2_seed_popular_models.sql` | target-only-review | 审查依赖/锁/覆盖价格映射/破坏性 DDL，生产副本验证 |
| `198_channel_monitor_v2_health_thresholds.sql` | target-only-review | 审查依赖/锁/覆盖价格映射/破坏性 DDL，生产副本验证 |
| `199_channel_monitor_v2_fixed_rollups.sql` | target-only-review | 审查依赖/锁/覆盖价格映射/破坏性 DDL，生产副本验证 |
| `200_channel_monitor_v2_rollup_permissions.sql` | target-only-review | 审查依赖/锁/覆盖价格映射/破坏性 DDL，生产副本验证 |
| `201_channel_monitor_v2_refresh_5m.sql` | target-only-review | 审查依赖/锁/覆盖价格映射/破坏性 DDL，生产副本验证 |
| `202_channel_monitor_v2_full_table_permissions.sql` | target-only-review | 审查依赖/锁/覆盖价格映射/破坏性 DDL，生产副本验证 |
| `203_channel_monitor_v2_default_ignore_and_cache.sql` | target-only-review | 审查依赖/锁/覆盖价格映射/破坏性 DDL，生产副本验证 |
| `204_channel_monitor_hide_throughput.sql` | target-only-review | 审查依赖/锁/覆盖价格映射/破坏性 DDL，生产副本验证 |
| `205_channel_monitor_v2_reset_factory_cache_thresholds.sql` | target-only-review | 审查依赖/锁/覆盖价格映射/破坏性 DDL，生产副本验证 |
| `206_channel_monitor_v2_privacy_defaults.sql` | target-only-review | 审查依赖/锁/覆盖价格映射/破坏性 DDL，生产副本验证 |
| `217_group_video_model_prices.sql` | target-only-review | 审查依赖/锁/覆盖价格映射/破坏性 DDL，生产副本验证 |
| `218_group_audio_voice_pricing.sql` | target-only-review | 审查依赖/锁/覆盖价格映射/破坏性 DDL，生产副本验证 |
| `219_group_search_price_per_1k.sql` | target-only-review | 审查依赖/锁/覆盖价格映射/破坏性 DDL，生产副本验证 |
| `220_clear_non_grok_video_generation_config.sql` | target-only-review | 审查依赖/锁/覆盖价格映射/破坏性 DDL，生产副本验证 |
| `221_group_model_pricing.sql` | target-only-review | 审查依赖/锁/覆盖价格映射/破坏性 DDL，生产副本验证 |
| `222_group_usage_daily_rollups.sql` | target-only-review | 审查依赖/锁/覆盖价格映射/破坏性 DDL，生产副本验证 |
| `223_group_usage_rollup_timezone.sql` | target-only-review | 审查依赖/锁/覆盖价格映射/破坏性 DDL，生产副本验证 |
| `224_user_platform_quotas_add_cn_providers.sql` | target-only-review | 审查依赖/锁/覆盖价格映射/破坏性 DDL，生产副本验证 |
| `225_backfill_codex_fingerprint_seed.sql` | target-only-review | 审查依赖/锁/覆盖价格映射/破坏性 DDL，生产副本验证 |
| `225_channel_model_time_pricing.sql` | target-only-review | 审查依赖/锁/覆盖价格映射/破坏性 DDL，生产副本验证 |
| `226_add_usage_log_effective_model_indexes_notx.sql` | target-only-review | 审查依赖/锁/覆盖价格映射/破坏性 DDL，生产副本验证 |
| `226_channel_monitor_quota_mode.sql` | target-only-review | 审查依赖/锁/覆盖价格映射/破坏性 DDL，生产副本验证 |
| `227_composite_routes_add_cn_providers.sql` | target-only-review | 审查依赖/锁/覆盖价格映射/破坏性 DDL，生产副本验证 |
| `228_channel_pricing_multipliers.sql` | target-only-review | 审查依赖/锁/覆盖价格映射/破坏性 DDL，生产副本验证 |
| `229_plugins.sql` | target-only-review | 审查依赖/锁/覆盖价格映射/破坏性 DDL，生产副本验证 |
| `230_plugin_artifacts.sql` | target-only-review | 审查依赖/锁/覆盖价格映射/破坏性 DDL，生产副本验证 |
| `231_add_usage_log_native_compaction_v2.sql` | target-only-review | 审查依赖/锁/覆盖价格映射/破坏性 DDL，生产副本验证 |
| `231_add_usage_log_requested_reasoning_effort.sql` | target-only-review | 审查依赖/锁/覆盖价格映射/破坏性 DDL，生产副本验证 |
| `231_user_restrict_public_groups.sql` | target-only-review | 审查依赖/锁/覆盖价格映射/破坏性 DDL，生产副本验证 |
| `232_add_usage_log_upstream_request_id.sql` | target-only-review | 审查依赖/锁/覆盖价格映射/破坏性 DDL，生产副本验证 |
| `232_channel_cache_write_1h_pricing.sql` | target-only-review | 审查依赖/锁/覆盖价格映射/破坏性 DDL，生产副本验证 |
| `232_group_force_openai_fast.sql` | target-only-review | 审查依赖/锁/覆盖价格映射/破坏性 DDL，生产副本验证 |
| `232_group_reasoning_effort_over_limit.sql` | target-only-review | 审查依赖/锁/覆盖价格映射/破坏性 DDL，生产副本验证 |
| `233_add_usage_log_upstream_request_id_index_notx.sql` | target-only-review | 审查依赖/锁/覆盖价格映射/破坏性 DDL，生产副本验证 |
| `233_group_free_openai_fast.sql` | target-only-review | 审查依赖/锁/覆盖价格映射/破坏性 DDL，生产副本验证 |
| `234_channel_max_reasoning_effort_multiplier.sql` | target-only-review | 审查依赖/锁/覆盖价格映射/破坏性 DDL，生产副本验证 |
| `234_group_codex_models_manifest_config.sql` | target-only-review | 审查依赖/锁/覆盖价格映射/破坏性 DDL，生产副本验证 |
| `235_group_model_allowlist.sql` | target-only-review | 审查依赖/锁/覆盖价格映射/破坏性 DDL，生产副本验证 |
| `236_group_model_allowlist_repair.sql` | target-only-review | 审查依赖/锁/覆盖价格映射/破坏性 DDL，生产副本验证 |
| `237_add_minimax_platform.sql` | target-only-review | 审查依赖/锁/覆盖价格映射/破坏性 DDL，生产副本验证 |
| `238_opencode_go_platform.sql` | target-only-review | 审查依赖/锁/覆盖价格映射/破坏性 DDL，生产副本验证 |
| `238_purge_unlimited_user_platform_quotas.sql` | target-only-review | 审查依赖/锁/覆盖价格映射/破坏性 DDL，生产副本验证 |
| `238b_content_moderation_engine_meta.sql` | target-only-review | 审查依赖/锁/覆盖价格映射/破坏性 DDL，生产副本验证 |
| `239_channel_reasoning_effort_multipliers.sql` | target-only-review | 审查依赖/锁/覆盖价格映射/破坏性 DDL，生产副本验证 |
| `240_affiliate_ledger_operation_id.sql` | target-only-review | 审查依赖/锁/覆盖价格映射/破坏性 DDL，生产副本验证 |
| `241_add_payment_order_bonus_amount.sql` | target-only-review | 审查依赖/锁/覆盖价格映射/破坏性 DDL，生产副本验证 |
| `241_add_typesafe_platform.sql` | target-only-review | 审查依赖/锁/覆盖价格映射/破坏性 DDL，生产副本验证 |
| `242_drop_platform_check_constraints.sql` | target-only-review | 审查依赖/锁/覆盖价格映射/破坏性 DDL，生产副本验证 |
| `242_group_i18n.sql` | target-only-review | 审查依赖/锁/覆盖价格映射/破坏性 DDL，生产副本验证 |
| `243_gemini_cached_contents.sql` | target-only-review | 审查依赖/锁/覆盖价格映射/破坏性 DDL，生产副本验证 |
| `244_gemini_cached_contents_channel_usage.sql` | target-only-review | 审查依赖/锁/覆盖价格映射/破坏性 DDL，生产副本验证 |
| `245_gemini_cached_contents_mutation_lease.sql` | target-only-review | 审查依赖/锁/覆盖价格映射/破坏性 DDL，生产副本验证 |
