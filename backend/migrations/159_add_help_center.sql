ALTER TABLE users
  ADD COLUMN IF NOT EXISTS help_center_key_prompt_dismissed BOOLEAN NOT NULL DEFAULT FALSE;

COMMENT ON COLUMN users.help_center_key_prompt_dismissed IS 'Whether the user permanently dismissed the Help Center key-created prompt.';

INSERT INTO settings (key, value, updated_at)
VALUES (
  'user_menu_items',
  '["dashboard","api_keys","help_center","image_generation","usage","channel_status","subscriptions","purchase","orders","redeem","affiliate","support_contact","profile"]',
  NOW()
)
ON CONFLICT (key) DO UPDATE
SET
  value = CASE
    WHEN settings.value IS NULL OR btrim(settings.value) = '' THEN EXCLUDED.value
    WHEN btrim(settings.value) NOT LIKE '[%' THEN EXCLUDED.value
    WHEN settings.value::jsonb ? 'help_center' THEN settings.value
    ELSE (
      SELECT jsonb_agg(item ORDER BY ord)
      FROM (
        SELECT elem AS item, ord
        FROM jsonb_array_elements(settings.value::jsonb) WITH ORDINALITY AS existing(elem, ord)
        UNION ALL
        SELECT '"help_center"'::jsonb AS item, 2.5::numeric AS ord
      ) merged
    )::text
  END,
  updated_at = NOW();

INSERT INTO settings (key, value, updated_at)
VALUES (
  'help_center_draft_config',
  $help_center_config${"enabled":true,"base_url":"https://useaifor.me","title":"帮助中心","description":"查看 Codex、Claude Code、OpenClaw、Hermes 等客户端接入教程，并前往 API 密钥页完成真实配置。","key_created_prompt":{"enabled":true,"title":"API 密钥已创建","description":"下一步可以进入帮助中心查看 Codex、Claude Code 等客户端的配置教程，或回到 API 密钥页使用现有配置入口。","primary_action_label":"查看帮助中心","primary_action_url":"/help-center","secondary_action_label":"留在 API 密钥页","secondary_action_url":"/keys","dismiss_label":"不再提示"},"tutorials":[{"id":"codex","enabled":true,"sort_order":1,"title":"Codex","badge":"Desktop/CLI","summary":"适用于 Codex Desktop 与 Codex CLI 的快速接入说明。","content_md":"Codex 可以通过平台生成的 API Key 进行访问。可以通过ccSwitch快捷配置。官网地址：https://ccswitch.io/zh/","steps":[{"title":"准备 API Key","description":"进入 API 密钥页创建或选择一个可用 key。","code_blocks":[],"images":[{"label":"image","url":"/api/v1/help-center/attachments/image-98fa390334ad.png","file_name":"image-98fa390334ad.png"},{"label":"image","url":"/api/v1/help-center/attachments/image-4e096373241a.png","file_name":"image-4e096373241a.png"}],"attachments":[]},{"title":"下载ccSwitch","description":"进入到官网后，点击免费下载，会跳转到github页面，然后选择版本进行下载，也可以直接下载平台提供的ccSwitch免安装包，这里只准备了macos和Windows的。","code_blocks":[],"images":[{"label":"image","url":"/api/v1/help-center/attachments/image-dcec1183237a.png","file_name":"image-dcec1183237a.png"},{"label":"image","url":"/api/v1/help-center/attachments/image-6ead8840f9fb.png","file_name":"image-6ead8840f9fb.png"},{"label":"image","url":"/api/v1/help-center/attachments/image-e82a09b83bb7.png","file_name":"image-e82a09b83bb7.png"}],"attachments":[{"label":"CC-Switch-v3.16.4-Windows-macos.zip","url":"/api/v1/help-center/attachments/CC-Switch-v3.16.4-Windows-macos-7c77aba6a27e.zip","file_name":"CC-Switch-v3.16.4-Windows-macos-7c77aba6a27e.zip"}]},{"title":"接入apikey","description":"点击导入到ccs","code_blocks":[],"images":[{"label":"image","url":"/api/v1/help-center/attachments/image-811ee2caf887.png","file_name":"image-811ee2caf887.png"},{"label":"image","url":"/api/v1/help-center/attachments/image-93e41dabd7a7.png","file_name":"image-93e41dabd7a7.png"},{"label":"image","url":"/api/v1/help-center/attachments/image-a7b158c6e591.png","file_name":"image-a7b158c6e591.png"},{"label":"image","url":"/api/v1/help-center/attachments/image-3c6af26af39d.png","file_name":"image-3c6af26af39d.png"}],"attachments":[]},{"title":"验证codex","description":"打开codex，开启会话，发送你好，如果有回复，说明就连通了","code_blocks":[],"images":[],"attachments":[]}],"code_blocks":[],"links":[],"attachments":[]},{"id":"claudecode","enabled":true,"sort_order":40,"title":"Claude code","badge":"Cli","summary":"适用于 Claude code的快速接入说明。","content_md":"Claude code 可以通过平台生成的 API Key 进行访问。可以通过ccSwitch快捷配置。官网地址：https://ccswitch.io/zh/","steps":[{"title":"准备 API Key","description":"进入 API 密钥页创建或选择一个可用 key。","code_blocks":[],"images":[{"label":"image","url":"/api/v1/help-center/attachments/image-98fa390334ad.png","file_name":"image-98fa390334ad.png"},{"label":"image","url":"/api/v1/help-center/attachments/image-74a6342665fa.png","file_name":"image-74a6342665fa.png"}],"attachments":[]},{"title":"下载ccSwitch","description":"进入到官网后，点击免费下载，会跳转到github页面，然后选择版本进行下载，也可以直接下载平台提供的ccSwitch免安装包，这里只准备了macos和Windows的。","code_blocks":[],"images":[{"label":"image","url":"/api/v1/help-center/attachments/image-dcec1183237a.png","file_name":"image-dcec1183237a.png"},{"label":"image","url":"/api/v1/help-center/attachments/image-6ead8840f9fb.png","file_name":"image-6ead8840f9fb.png"},{"label":"image","url":"/api/v1/help-center/attachments/image-e82a09b83bb7.png","file_name":"image-e82a09b83bb7.png"}],"attachments":[{"label":"CC-Switch-v3.16.4-Windows-macos.zip","url":"/api/v1/help-center/attachments/CC-Switch-v3.16.4-Windows-macos-7c77aba6a27e.zip","file_name":"CC-Switch-v3.16.4-Windows-macos-7c77aba6a27e.zip"}]},{"title":"接入apikey","description":"点击导入到ccs","code_blocks":[],"images":[{"label":"image","url":"/api/v1/help-center/attachments/image-8472a30ddf2b.png","file_name":"image-8472a30ddf2b.png"},{"label":"image","url":"/api/v1/help-center/attachments/image-9989879625e5.png","file_name":"image-9989879625e5.png"},{"label":"image","url":"/api/v1/help-center/attachments/image-0a49958ac1cb.png","file_name":"image-0a49958ac1cb.png"},{"label":"image","url":"/api/v1/help-center/attachments/image-e3fc9ae4788b.png","file_name":"image-e3fc9ae4788b.png"}],"attachments":[]},{"title":"验证Claude code","description":"打开Claude code cli，开启会话，发送你好，如果有回复，说明就连通了","code_blocks":[],"images":[],"attachments":[]}],"code_blocks":[],"links":[],"attachments":[]}],"faqs":[{"id":"where-to-create-key","enabled":true,"sort_order":10,"question":"在哪里创建或选择 API Key？","answer_md":"请先进入 **API 密钥** 页面创建或选择一个 key。帮助中心只负责提供客户端配置教程","tags":["API 密钥","新手"]}]}$help_center_config$,
  NOW()
)
ON CONFLICT (key) DO UPDATE
SET value = EXCLUDED.value,
    updated_at = NOW()
WHERE settings.value IS NULL
   OR btrim(settings.value) = ''
   OR jsonb_typeof(settings.value::jsonb) <> 'object'
   OR COALESCE(jsonb_array_length(settings.value::jsonb -> 'tutorials'), 0) = 0
   OR (settings.value::jsonb -> 'tutorials' -> 0 ->> 'id') IN ('codex-desktop', 'codex-cli');

INSERT INTO settings (key, value, updated_at)
SELECT 'help_center_published_config', value, NOW()
FROM settings
WHERE key = 'help_center_draft_config'
ON CONFLICT (key) DO UPDATE
SET value = EXCLUDED.value,
    updated_at = NOW()
WHERE settings.value IS NULL
   OR btrim(settings.value) = ''
   OR jsonb_typeof(settings.value::jsonb) <> 'object'
   OR COALESCE(jsonb_array_length(settings.value::jsonb -> 'tutorials'), 0) = 0
   OR (settings.value::jsonb -> 'tutorials' -> 0 ->> 'id') IN ('codex-desktop', 'codex-cli');
