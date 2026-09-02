package service

import (
	"context"
	"encoding/json"
	"fmt"
	"net/url"
	"regexp"
	"sort"
	"strings"
)

var helpCenterIDPattern = regexp.MustCompile(`^[a-z0-9][a-z0-9_-]{1,63}$`)

type HelpCenterConfig struct {
	Enabled          bool                       `json:"enabled"`
	BaseURL          string                     `json:"base_url"`
	Title            string                     `json:"title"`
	Description      string                     `json:"description"`
	KeyCreatedPrompt HelpCenterKeyCreatedPrompt `json:"key_created_prompt"`
	Tutorials        []HelpCenterTutorial       `json:"tutorials"`
	FAQs             []HelpCenterFAQ            `json:"faqs"`
}

type HelpCenterKeyCreatedPrompt struct {
	Enabled              bool   `json:"enabled"`
	Title                string `json:"title"`
	Description          string `json:"description"`
	PrimaryActionLabel   string `json:"primary_action_label"`
	PrimaryActionURL     string `json:"primary_action_url"`
	SecondaryActionLabel string `json:"secondary_action_label"`
	SecondaryActionURL   string `json:"secondary_action_url"`
	DismissLabel         string `json:"dismiss_label"`
}

type HelpCenterTutorial struct {
	ID          string                 `json:"id"`
	Enabled     bool                   `json:"enabled"`
	SortOrder   int                    `json:"sort_order"`
	Title       string                 `json:"title"`
	Badge       string                 `json:"badge"`
	Summary     string                 `json:"summary"`
	ContentMD   string                 `json:"content_md"`
	Steps       []HelpCenterStep       `json:"steps"`
	CodeBlocks  []HelpCenterCodeBlock  `json:"code_blocks"`
	Links       []HelpCenterLink       `json:"links"`
	Attachments []HelpCenterAttachment `json:"attachments"`
}

type HelpCenterStep struct {
	Title       string                 `json:"title"`
	Description string                 `json:"description"`
	CodeBlocks  []HelpCenterCodeBlock  `json:"code_blocks"`
	Images      []HelpCenterAttachment `json:"images"`
	Attachments []HelpCenterAttachment `json:"attachments"`
}

type HelpCenterCodeBlock struct {
	Title    string `json:"title"`
	Language string `json:"language"`
	Content  string `json:"content"`
}

type HelpCenterLink struct {
	Label string `json:"label"`
	URL   string `json:"url"`
}

type HelpCenterAttachment struct {
	Label    string `json:"label"`
	URL      string `json:"url"`
	FileName string `json:"file_name"`
}

type HelpCenterFAQ struct {
	ID        string   `json:"id"`
	Enabled   bool     `json:"enabled"`
	SortOrder int      `json:"sort_order"`
	Question  string   `json:"question"`
	AnswerMD  string   `json:"answer_md"`
	Tags      []string `json:"tags"`
}

type HelpCenterService struct {
	settingRepo SettingRepository
}

func NewHelpCenterService(settingRepo SettingRepository) *HelpCenterService {
	return &HelpCenterService{settingRepo: settingRepo}
}

const defaultHelpCenterConfigLiteral = `{"enabled":true,"base_url":"https://58token.vip","title":"帮助中心","description":"查看 Codex、Claude Code、OpenClaw、Hermes 等客户端接入教程，并前往 API 密钥页完成真实配置。","key_created_prompt":{"enabled":true,"title":"API 密钥已创建","description":"下一步可以进入帮助中心查看 Codex、Claude Code 等客户端的配置教程，或回到 API 密钥页使用现有配置入口。","primary_action_label":"查看帮助中心","primary_action_url":"/help-center","secondary_action_label":"留在 API 密钥页","secondary_action_url":"/keys","dismiss_label":"不再提示"},"tutorials":[{"id":"codex","enabled":true,"sort_order":1,"title":"Codex","badge":"Desktop/CLI","summary":"适用于 Codex Desktop 与 Codex CLI 的快速接入说明。","content_md":"Codex 可以通过平台生成的 API Key 进行访问。可以通过ccSwitch快捷配置。官网地址：https://ccswitch.io/zh/","steps":[{"title":"准备 API Key","description":"进入 API 密钥页创建或选择一个可用 key。","code_blocks":[],"images":[{"label":"image","url":"/api/v1/help-center/attachments/image-98fa390334ad.png","file_name":"image-98fa390334ad.png"},{"label":"image","url":"/api/v1/help-center/attachments/image-4e096373241a.png","file_name":"image-4e096373241a.png"}],"attachments":[]},{"title":"下载ccSwitch","description":"进入到官网后，点击免费下载，会跳转到github页面，然后选择版本进行下载，也可以直接下载平台提供的ccSwitch免安装包，这里只准备了macos和Windows的。","code_blocks":[],"images":[{"label":"image","url":"/api/v1/help-center/attachments/image-dcec1183237a.png","file_name":"image-dcec1183237a.png"},{"label":"image","url":"/api/v1/help-center/attachments/image-6ead8840f9fb.png","file_name":"image-6ead8840f9fb.png"},{"label":"image","url":"/api/v1/help-center/attachments/image-e82a09b83bb7.png","file_name":"image-e82a09b83bb7.png"}],"attachments":[{"label":"CC-Switch-v3.16.4-Windows-macos.zip","url":"/api/v1/help-center/attachments/CC-Switch-v3.16.4-Windows-macos-7c77aba6a27e.zip","file_name":"CC-Switch-v3.16.4-Windows-macos-7c77aba6a27e.zip"}]},{"title":"接入apikey","description":"点击导入到ccs","code_blocks":[],"images":[{"label":"image","url":"/api/v1/help-center/attachments/image-811ee2caf887.png","file_name":"image-811ee2caf887.png"},{"label":"image","url":"/api/v1/help-center/attachments/image-93e41dabd7a7.png","file_name":"image-93e41dabd7a7.png"},{"label":"image","url":"/api/v1/help-center/attachments/image-a7b158c6e591.png","file_name":"image-a7b158c6e591.png"},{"label":"image","url":"/api/v1/help-center/attachments/image-3c6af26af39d.png","file_name":"image-3c6af26af39d.png"}],"attachments":[]},{"title":"验证codex","description":"打开codex，开启会话，发送你好，如果有回复，说明就连通了","code_blocks":[],"images":[],"attachments":[]}],"code_blocks":[],"links":[],"attachments":[]},{"id":"claudecode","enabled":true,"sort_order":40,"title":"Claude code","badge":"Cli","summary":"适用于 Claude code的快速接入说明。","content_md":"Claude code 可以通过平台生成的 API Key 进行访问。可以通过ccSwitch快捷配置。官网地址：https://ccswitch.io/zh/","steps":[{"title":"准备 API Key","description":"进入 API 密钥页创建或选择一个可用 key。","code_blocks":[],"images":[{"label":"image","url":"/api/v1/help-center/attachments/image-98fa390334ad.png","file_name":"image-98fa390334ad.png"},{"label":"image","url":"/api/v1/help-center/attachments/image-74a6342665fa.png","file_name":"image-74a6342665fa.png"}],"attachments":[]},{"title":"下载ccSwitch","description":"进入到官网后，点击免费下载，会跳转到github页面，然后选择版本进行下载，也可以直接下载平台提供的ccSwitch免安装包，这里只准备了macos和Windows的。","code_blocks":[],"images":[{"label":"image","url":"/api/v1/help-center/attachments/image-dcec1183237a.png","file_name":"image-dcec1183237a.png"},{"label":"image","url":"/api/v1/help-center/attachments/image-6ead8840f9fb.png","file_name":"image-6ead8840f9fb.png"},{"label":"image","url":"/api/v1/help-center/attachments/image-e82a09b83bb7.png","file_name":"image-e82a09b83bb7.png"}],"attachments":[{"label":"CC-Switch-v3.16.4-Windows-macos.zip","url":"/api/v1/help-center/attachments/CC-Switch-v3.16.4-Windows-macos-7c77aba6a27e.zip","file_name":"CC-Switch-v3.16.4-Windows-macos-7c77aba6a27e.zip"}]},{"title":"接入apikey","description":"点击导入到ccs","code_blocks":[],"images":[{"label":"image","url":"/api/v1/help-center/attachments/image-8472a30ddf2b.png","file_name":"image-8472a30ddf2b.png"},{"label":"image","url":"/api/v1/help-center/attachments/image-9989879625e5.png","file_name":"image-9989879625e5.png"},{"label":"image","url":"/api/v1/help-center/attachments/image-0a49958ac1cb.png","file_name":"image-0a49958ac1cb.png"},{"label":"image","url":"/api/v1/help-center/attachments/image-e3fc9ae4788b.png","file_name":"image-e3fc9ae4788b.png"}],"attachments":[]},{"title":"验证Claude code","description":"打开Claude code cli，开启会话，发送你好，如果有回复，说明就连通了","code_blocks":[],"images":[],"attachments":[]}],"code_blocks":[],"links":[],"attachments":[]}],"faqs":[{"id":"where-to-create-key","enabled":true,"sort_order":10,"question":"在哪里创建或选择 API Key？","answer_md":"请先进入 **API 密钥** 页面创建或选择一个 key。帮助中心只负责提供客户端配置教程","tags":["API 密钥","新手"]}]}`

func DefaultHelpCenterConfig() HelpCenterConfig {
	var cfg HelpCenterConfig
	if err := json.Unmarshal([]byte(defaultHelpCenterConfigLiteral), &cfg); err == nil {
		return cfg
	}
	return HelpCenterConfig{
		Enabled:     true,
		BaseURL:     "https://58token.vip",
		Title:       "帮助中心",
		Description: "查看 Codex、Claude Code、OpenClaw、Hermes 等客户端接入教程，并前往 API 密钥页完成真实配置。",
		KeyCreatedPrompt: HelpCenterKeyCreatedPrompt{
			Enabled:              true,
			Title:                "API 密钥已创建",
			Description:          "下一步可以进入帮助中心查看 Codex、Claude Code 等客户端的配置教程，或回到 API 密钥页使用现有配置入口。",
			PrimaryActionLabel:   "查看帮助中心",
			PrimaryActionURL:     "/help-center",
			SecondaryActionLabel: "留在 API 密钥页",
			SecondaryActionURL:   "/keys",
			DismissLabel:         "不再提示",
		},
		Tutorials: []HelpCenterTutorial{},
		FAQs: []HelpCenterFAQ{
			defaultHelpCenterFAQ("where-to-create-key", 10, "在哪里创建或选择 API Key？", "请先进入 **API 密钥** 页面创建或选择一个 key。帮助中心只负责提供客户端配置教程，不重复密钥管理能力。", []string{"API 密钥", "新手"}),
		},
	}
}

func DefaultHelpCenterConfigJSON() string {
	cfg := normalizedDefaultHelpCenterConfig()
	data, err := json.Marshal(cfg)
	if err != nil {
		return "{}"
	}
	return string(data)
}

func normalizedDefaultHelpCenterConfig() HelpCenterConfig {
	cfg, err := NormalizeHelpCenterConfig(DefaultHelpCenterConfig())
	if err != nil {
		return DefaultHelpCenterConfig()
	}
	return cfg
}

func defaultHelpCenterKeyLinks() []HelpCenterLink {
	return []HelpCenterLink{
		{Label: "前往 API 密钥页", URL: "/keys"},
	}
}

func defaultCodexDesktopTutorial() HelpCenterTutorial {
	return HelpCenterTutorial{
		ID:        "codex-desktop",
		Enabled:   true,
		SortOrder: 10,
		Title:     "Codex Desktop",
		Badge:     "Desktop",
		Summary:   "适用于 Codex 桌面端或支持 OpenAI Responses API 的桌面客户端。包含 OpenAI 兼容模式和 Codex backend API 两种地址写法。",
		ContentMD: "Codex Desktop 通常只需要 Base URL、API Key 和模型名。平台同时提供 OpenAI 兼容入口与 Codex 专用入口：客户端选择 OpenAI-compatible 时使用 Base URL + /v1；如果客户端字段写的是 Codex backend API 或 Codex API Base URL，使用 Base URL + /backend-api/codex。",
		Steps: []HelpCenterStep{
			{
				Title:       "准备 API Key",
				Description: "进入 API 密钥页创建或选择一个可用 key。帮助中心不保存真实密钥，只提供配置路径；复制密钥后请只粘贴到本机 Codex Desktop。",
			},
			{
				Title:       "打开 Codex Desktop 的 Provider 设置",
				Description: "在 Codex Desktop 中进入 Settings / Providers / Custom Provider。字段名称可能叫 OpenAI-compatible、Custom OpenAI、API Provider 或 Backend API。",
			},
			{
				Title:       "选择合适的 Base URL",
				Description: "如果客户端选择 OpenAI-compatible，Base URL 填：帮助中心顶部 Base URL + /v1。如果客户端明确要求 Codex backend API 地址，填：帮助中心顶部 Base URL + /backend-api/codex。",
			},
			{
				Title:       "填写鉴权方式",
				Description: "API Key 填 API 密钥页复制的 key。请求头优先使用 Authorization: Bearer <API Key>；如果客户端只支持 x-api-key，也可以填同一个 key。",
			},
			{
				Title:       "选择模型并保存",
				Description: "模型名填写平台可用模型名称。使用 OpenAI/Codex 模型时，请确保 key 绑定的分组或上游支持 OpenAI 兼容入口；使用 Claude 模型时建议改用 Claude/Anthropic 兼容教程。",
			},
			{
				Title:       "发送测试消息",
				Description: "新建会话发送一个简单问题。401 通常是 key 错误或未启用，403 通常是余额/权限/分组不可用，404 多数是 Base URL 多写或少写了 /v1、/backend-api/codex。",
			},
		},
		CodeBlocks: []HelpCenterCodeBlock{
			{
				Title:    "OpenAI-compatible 模式",
				Language: "json",
				Content: `{
  "provider": "openai-compatible",
  "baseURL": "<帮助中心显示的 Base URL>/v1",
  "apiKey": "<你的 API Key>",
  "model": "<平台模型名>"
}`,
			},
			{
				Title:    "Codex backend API 模式",
				Language: "json",
				Content: `{
  "provider": "codex",
  "codexApiBaseURL": "<帮助中心显示的 Base URL>/backend-api/codex",
  "apiKey": "<你的 API Key>",
  "model": "<平台模型名>"
}`,
			},
		},
		Links: defaultHelpCenterKeyLinks(),
	}
}

func defaultCodexCLITutorial() HelpCenterTutorial {
	return HelpCenterTutorial{
		ID:        "codex-cli",
		Enabled:   true,
		SortOrder: 20,
		Title:     "Codex CLI",
		Badge:     "CLI",
		Summary:   "适用于 Codex CLI 自定义 provider。推荐在 ~/.codex/config.toml 配置 model_provider，并把 API Key 放到环境变量。",
		ContentMD: "Codex CLI 推荐使用配置文件管理自定义 provider，API Key 通过环境变量读取。代理或中转场景建议使用 Responses API；如果你只跑 chat/completions 型客户端，也可以把 wire_api 调整为 chat。",
		Steps: []HelpCenterStep{
			{
				Title:       "创建或选择 API Key",
				Description: "在 API 密钥页复制 key。不要把真实 key 写进公开仓库或截图，建议放在系统环境变量里。",
			},
			{
				Title:       "设置环境变量",
				Description: "把 USEAIFOR_API_KEY 设置为你的平台 key。PowerShell、macOS/Linux shell 都可以，示例见下方配置块。",
			},
			{
				Title:       "编辑 Codex CLI 配置",
				Description: "打开 ~/.codex/config.toml，新增 useaifor provider，并把 model_provider 指向 useaifor。Windows 对应路径通常在用户目录下的 .codex/config.toml。",
			},
			{
				Title:       "填写 Base URL 与协议",
				Description: "base_url 使用帮助中心顶部 Base URL + /v1。wire_api 建议填 responses，因为平台已提供 /v1/responses；如你的客户端版本只支持 Chat Completions，可改为 chat。",
			},
			{
				Title:       "选择模型",
				Description: "model 填平台可用模型名，例如你在平台上能调用的 OpenAI/Codex 模型名称。模型是否可用取决于 key 绑定的分组和上游渠道。",
			},
			{
				Title:       "重启终端并测试",
				Description: "保存配置后重新打开终端，再启动 codex。若仍走默认官方地址，通常是 model_provider 没切到 useaifor 或环境变量没有在当前终端生效。",
			},
		},
		CodeBlocks: []HelpCenterCodeBlock{
			{
				Title:    "macOS / Linux 环境变量",
				Language: "bash",
				Content:  `export USEAIFOR_API_KEY="<你的 API Key>"`,
			},
			{
				Title:    "Windows PowerShell 环境变量",
				Language: "powershell",
				Content:  `$env:USEAIFOR_API_KEY="<你的 API Key>"`,
			},
			{
				Title:    "~/.codex/config.toml 示例",
				Language: "toml",
				Content: `model_provider = "useaifor"
model = "<平台模型名>"

[model_providers.useaifor]
name = "UseAiFor"
base_url = "<帮助中心显示的 Base URL>/v1"
env_key = "USEAIFOR_API_KEY"
wire_api = "responses"`,
			},
		},
		Links: defaultHelpCenterKeyLinks(),
	}
}

func defaultClaudeCodeTutorial() HelpCenterTutorial {
	return HelpCenterTutorial{
		ID:        "claude-code",
		Enabled:   true,
		SortOrder: 30,
		Title:     "Claude Code",
		Badge:     "CLI",
		Summary:   "适用于 Claude Code 或 Anthropic Messages API 兼容客户端。使用 ANTHROPIC_BASE_URL 与 ANTHROPIC_AUTH_TOKEN 接入。",
		ContentMD: "Claude Code 走 Anthropic 兼容接口时，平台会接收 /v1/messages 请求。请使用支持 Claude/Anthropic 的 key 或分组；如果你要调用 OpenAI 模型，请改看 Codex CLI 或 OpenAI 兼容教程。",
		Steps: []HelpCenterStep{
			{
				Title:       "准备 Claude/Anthropic 可用的 API Key",
				Description: "进入 API 密钥页复制 key，并确认该 key 对应的分组、订阅或上游渠道允许调用 Claude/Anthropic 兼容模型。",
			},
			{
				Title:       "设置 Base URL",
				Description: "ANTHROPIC_BASE_URL 填帮助中心顶部 Base URL + /v1。平台会在该前缀下处理 /messages、/models 等 Anthropic 兼容请求。",
			},
			{
				Title:       "设置鉴权 token",
				Description: "Claude Code 网关场景优先使用 ANTHROPIC_AUTH_TOKEN，值就是 API 密钥页复制的 key。部分 Anthropic 兼容客户端字段名可能叫 ANTHROPIC_API_KEY，值保持一致即可。",
			},
			{
				Title:       "选择或指定模型",
				Description: "如果客户端让你填写 model，使用平台可用的 Claude 模型名。不要把 OpenAI 模型名填到 Claude Code 的 Anthropic 模式里。",
			},
			{
				Title:       "重启 Claude Code",
				Description: "环境变量改完后关闭并重新打开终端或 IDE，再运行 Claude Code。很多连接失败只是因为旧进程没读取到新的变量。",
			},
			{
				Title:       "排查常见错误",
				Description: "401 检查 token 是否复制完整；403 检查余额、订阅、分组或模型权限；404 检查 Base URL 是否为 Base URL + /v1，且客户端没有额外重复拼接 /v1。",
			},
		},
		CodeBlocks: []HelpCenterCodeBlock{
			{
				Title:    "macOS / Linux",
				Language: "bash",
				Content: `export ANTHROPIC_BASE_URL="<帮助中心显示的 Base URL>/v1"
export ANTHROPIC_AUTH_TOKEN="<你的 API Key>"

# 如果你的客户端版本只识别 ANTHROPIC_API_KEY，可改用：
export ANTHROPIC_API_KEY="<你的 API Key>"`,
			},
			{
				Title:    "Windows PowerShell",
				Language: "powershell",
				Content: `$env:ANTHROPIC_BASE_URL="<帮助中心显示的 Base URL>/v1"
$env:ANTHROPIC_AUTH_TOKEN="<你的 API Key>"

# 如果你的客户端版本只识别 ANTHROPIC_API_KEY，可改用：
$env:ANTHROPIC_API_KEY="<你的 API Key>"`,
			},
			{
				Title:    "字段映射",
				Language: "text",
				Content: `Base URL / API Base URL: <帮助中心显示的 Base URL>/v1
API Key / Auth Token: <你的 API Key>
Header: Authorization: Bearer <你的 API Key>
Endpoint: /v1/messages`,
			},
		},
		Links: defaultHelpCenterKeyLinks(),
	}
}

func defaultOpenClawTutorial() HelpCenterTutorial {
	return HelpCenterTutorial{
		ID:        "openclaw",
		Enabled:   true,
		SortOrder: 40,
		Title:     "OpenClaw",
		Badge:     "Client",
		Summary:   "适用于 OpenClaw 的 OpenAI-compatible 配置。主要填写 Provider 类型、Base URL、API Key 和模型名。",
		ContentMD: "OpenClaw 如果提供 Provider 类型，请选择 OpenAI Compatible / Custom OpenAI。平台支持 Authorization Bearer 鉴权，也兼容 x-api-key。Claude 模型请在 OpenClaw 中选择 Anthropic/Claude 兼容模式，仍使用同一个 Base URL + /v1。",
		Steps: []HelpCenterStep{
			{
				Title:       "进入 Provider 或 Model 配置页",
				Description: "在 OpenClaw 中打开 Settings / Provider / Model Provider，新增一个自定义服务，名称可以写 UseAiFor。",
			},
			{
				Title:       "选择 OpenAI Compatible",
				Description: "调用 OpenAI、Codex 或大多数 OpenAI 兼容模型时，Provider Type 选择 OpenAI Compatible、Custom OpenAI 或 OpenAI API Compatible。",
			},
			{
				Title:       "填写 Base URL",
				Description: "Base URL 填帮助中心顶部 Base URL + /v1。不要再额外拼 /chat/completions；客户端会自动拼接接口路径。",
			},
			{
				Title:       "填写 API Key",
				Description: "API Key 填 API 密钥页复制的 key。若 OpenClaw 有 Header 设置，使用 Authorization: Bearer <API Key>；若只有 x-api-key 字段，也填同一个 key。",
			},
			{
				Title:       "填写模型名",
				Description: "模型名使用平台可用模型名称。配置保存后先用一个轻量模型测试，确认网络、余额和权限都正常。",
			},
			{
				Title:       "Claude 模型的特殊情况",
				Description: "如果 OpenClaw 里要接 Claude/Anthropic 模型，Provider Type 改选 Anthropic Compatible / Claude Compatible，Base URL 仍填帮助中心顶部 Base URL + /v1。",
			},
		},
		CodeBlocks: []HelpCenterCodeBlock{
			{
				Title:    "OpenAI-compatible 字段",
				Language: "text",
				Content: `Provider Name: UseAiFor
Provider Type: OpenAI Compatible
Base URL: <帮助中心显示的 Base URL>/v1
API Key: <你的 API Key>
Model: <平台模型名>`,
			},
			{
				Title:    "可选自定义 Header",
				Language: "http",
				Content:  `Authorization: Bearer <你的 API Key>`,
			},
			{
				Title:    "Claude-compatible 字段",
				Language: "text",
				Content: `Provider Name: UseAiFor Claude
Provider Type: Anthropic Compatible / Claude Compatible
Base URL: <帮助中心显示的 Base URL>/v1
API Key: <你的 API Key>
Model: <平台 Claude 模型名>`,
			},
		},
		Links: defaultHelpCenterKeyLinks(),
	}
}

func defaultHermesTutorial() HelpCenterTutorial {
	return HelpCenterTutorial{
		ID:        "hermes",
		Enabled:   true,
		SortOrder: 50,
		Title:     "Hermes",
		Badge:     "Client",
		Summary:   "适用于 Hermes 的多 Provider 配置。可同时添加 OpenAI-compatible 与 Claude/Anthropic-compatible 两个入口。",
		ContentMD: "Hermes 如果支持多个 provider，建议分别创建 UseAiFor OpenAI 和 UseAiFor Claude。两者都使用帮助中心顶部 Base URL + /v1，区别在于 Provider 类型和模型名。",
		Steps: []HelpCenterStep{
			{
				Title:       "决定使用哪种兼容模式",
				Description: "OpenAI、Codex、Responses 或 Chat Completions 模型选择 OpenAI Compatible；Claude 模型选择 Anthropic Compatible / Claude Compatible。",
			},
			{
				Title:       "创建 OpenAI provider",
				Description: "在 Hermes 的 Providers 页面新增 UseAiFor OpenAI，Base URL 填帮助中心顶部 Base URL + /v1，API Key 填 API 密钥页复制的 key。",
			},
			{
				Title:       "按需创建 Claude provider",
				Description: "如果你也要在 Hermes 里使用 Claude 模型，再新增 UseAiFor Claude，Provider Type 选择 Anthropic Compatible，Base URL 同样填 Base URL + /v1。",
			},
			{
				Title:       "绑定模型",
				Description: "在模型列表中填写平台模型名，并把模型绑定到对应 provider。OpenAI 模型不要绑定到 Claude provider，Claude 模型不要绑定到 OpenAI provider。",
			},
			{
				Title:       "保存并测试",
				Description: "保存后分别用两个 provider 各发一条测试消息。OpenAI 兼容一般会请求 /v1/chat/completions 或 /v1/responses；Claude 兼容一般会请求 /v1/messages。",
			},
			{
				Title:       "排查 URL 拼接",
				Description: "如果报 404，优先检查 Hermes 是否要求填写根地址还是 /v1 地址。大多数兼容模式填写 Base URL + /v1；不要把完整接口路径当 Base URL。",
			},
		},
		CodeBlocks: []HelpCenterCodeBlock{
			{
				Title:    "Hermes OpenAI provider 示例",
				Language: "yaml",
				Content: `providers:
  useaifor-openai:
    type: openai-compatible
    base_url: "<帮助中心显示的 Base URL>/v1"
    api_key: "<你的 API Key>"
    models:
      - "<平台 OpenAI/Codex 模型名>"`,
			},
			{
				Title:    "Hermes Claude provider 示例",
				Language: "yaml",
				Content: `providers:
  useaifor-claude:
    type: anthropic-compatible
    base_url: "<帮助中心显示的 Base URL>/v1"
    api_key: "<你的 API Key>"
    models:
      - "<平台 Claude 模型名>"`,
			},
			{
				Title:    "通用字段对照",
				Language: "text",
				Content: `OpenAI Compatible Base URL: <帮助中心显示的 Base URL>/v1
Claude Compatible Base URL: <帮助中心显示的 Base URL>/v1
API Key: <你的 API Key>
OpenAI endpoints: /v1/responses, /v1/chat/completions
Claude endpoint: /v1/messages`,
			},
		},
		Links: defaultHelpCenterKeyLinks(),
	}
}

func defaultHelpCenterFAQ(id string, order int, question string, answer string, tags []string) HelpCenterFAQ {
	return HelpCenterFAQ{
		ID:        id,
		Enabled:   true,
		SortOrder: order,
		Question:  question,
		AnswerMD:  answer,
		Tags:      tags,
	}
}

func NormalizeHelpCenterConfig(cfg HelpCenterConfig) (HelpCenterConfig, error) {
	cfg.Title = strings.TrimSpace(cfg.Title)
	if cfg.Title == "" {
		cfg.Title = "帮助中心"
	}
	cfg.Description = strings.TrimSpace(cfg.Description)
	cfg.BaseURL = strings.TrimRight(strings.TrimSpace(cfg.BaseURL), "/")
	if cfg.BaseURL != "" && !isAllowedHelpCenterURL(cfg.BaseURL) {
		return HelpCenterConfig{}, fmt.Errorf("invalid url for base_url")
	}

	cfg.KeyCreatedPrompt.Title = strings.TrimSpace(cfg.KeyCreatedPrompt.Title)
	cfg.KeyCreatedPrompt.Description = strings.TrimSpace(cfg.KeyCreatedPrompt.Description)
	cfg.KeyCreatedPrompt.PrimaryActionLabel = strings.TrimSpace(cfg.KeyCreatedPrompt.PrimaryActionLabel)
	cfg.KeyCreatedPrompt.PrimaryActionURL = strings.TrimSpace(cfg.KeyCreatedPrompt.PrimaryActionURL)
	cfg.KeyCreatedPrompt.SecondaryActionLabel = strings.TrimSpace(cfg.KeyCreatedPrompt.SecondaryActionLabel)
	cfg.KeyCreatedPrompt.SecondaryActionURL = strings.TrimSpace(cfg.KeyCreatedPrompt.SecondaryActionURL)
	cfg.KeyCreatedPrompt.DismissLabel = strings.TrimSpace(cfg.KeyCreatedPrompt.DismissLabel)
	if cfg.KeyCreatedPrompt.Title == "" {
		cfg.KeyCreatedPrompt.Title = "API 密钥已创建"
	}
	if cfg.KeyCreatedPrompt.PrimaryActionLabel == "" {
		cfg.KeyCreatedPrompt.PrimaryActionLabel = "查看帮助中心"
	}
	if cfg.KeyCreatedPrompt.PrimaryActionURL == "" {
		cfg.KeyCreatedPrompt.PrimaryActionURL = "/help-center"
	}
	if cfg.KeyCreatedPrompt.SecondaryActionLabel == "" {
		cfg.KeyCreatedPrompt.SecondaryActionLabel = "留在 API 密钥页"
	}
	if cfg.KeyCreatedPrompt.SecondaryActionURL == "" {
		cfg.KeyCreatedPrompt.SecondaryActionURL = "/keys"
	}
	if cfg.KeyCreatedPrompt.DismissLabel == "" {
		cfg.KeyCreatedPrompt.DismissLabel = "不再提示"
	}
	if !isAllowedHelpCenterURL(cfg.KeyCreatedPrompt.PrimaryActionURL) {
		return HelpCenterConfig{}, fmt.Errorf("invalid url for key_created_prompt.primary_action_url")
	}
	if !isAllowedHelpCenterURL(cfg.KeyCreatedPrompt.SecondaryActionURL) {
		return HelpCenterConfig{}, fmt.Errorf("invalid url for key_created_prompt.secondary_action_url")
	}

	if cfg.Tutorials == nil {
		cfg.Tutorials = []HelpCenterTutorial{}
	}
	seen := make(map[string]struct{}, len(cfg.Tutorials))
	for i := range cfg.Tutorials {
		tutorial := &cfg.Tutorials[i]
		tutorial.ID = strings.ToLower(strings.TrimSpace(tutorial.ID))
		if !helpCenterIDPattern.MatchString(tutorial.ID) {
			return HelpCenterConfig{}, fmt.Errorf("invalid tutorial id %q", tutorial.ID)
		}
		if _, ok := seen[tutorial.ID]; ok {
			return HelpCenterConfig{}, fmt.Errorf("duplicate tutorial id %q", tutorial.ID)
		}
		seen[tutorial.ID] = struct{}{}

		tutorial.Title = strings.TrimSpace(tutorial.Title)
		if tutorial.Enabled && tutorial.Title == "" {
			return HelpCenterConfig{}, fmt.Errorf("tutorial %q title is required", tutorial.ID)
		}
		tutorial.Badge = strings.TrimSpace(tutorial.Badge)
		tutorial.Summary = strings.TrimSpace(tutorial.Summary)
		tutorial.ContentMD = strings.TrimSpace(tutorial.ContentMD)
		if tutorial.Steps == nil {
			tutorial.Steps = []HelpCenterStep{}
		}
		for j := range tutorial.Steps {
			step := &tutorial.Steps[j]
			step.Title = strings.TrimSpace(step.Title)
			step.Description = strings.TrimSpace(step.Description)
			if step.CodeBlocks == nil {
				step.CodeBlocks = []HelpCenterCodeBlock{}
			}
			for k := range step.CodeBlocks {
				step.CodeBlocks[k].Title = strings.TrimSpace(step.CodeBlocks[k].Title)
				step.CodeBlocks[k].Language = strings.TrimSpace(step.CodeBlocks[k].Language)
			}
			if step.Images == nil {
				step.Images = []HelpCenterAttachment{}
			}
			for k := range step.Images {
				step.Images[k].Label = strings.TrimSpace(step.Images[k].Label)
				step.Images[k].URL = strings.TrimSpace(step.Images[k].URL)
				step.Images[k].FileName = strings.TrimSpace(step.Images[k].FileName)
				if step.Images[k].Label == "" || step.Images[k].URL == "" {
					return HelpCenterConfig{}, fmt.Errorf("tutorial %q step image label and url are required", tutorial.ID)
				}
				if !isAllowedHelpCenterURL(step.Images[k].URL) {
					return HelpCenterConfig{}, fmt.Errorf("invalid url for tutorial %q step image", tutorial.ID)
				}
			}
			if step.Attachments == nil {
				step.Attachments = []HelpCenterAttachment{}
			}
			for k := range step.Attachments {
				step.Attachments[k].Label = strings.TrimSpace(step.Attachments[k].Label)
				step.Attachments[k].URL = strings.TrimSpace(step.Attachments[k].URL)
				step.Attachments[k].FileName = strings.TrimSpace(step.Attachments[k].FileName)
				if step.Attachments[k].Label == "" || step.Attachments[k].URL == "" {
					return HelpCenterConfig{}, fmt.Errorf("tutorial %q step attachment label and url are required", tutorial.ID)
				}
				if !isAllowedHelpCenterURL(step.Attachments[k].URL) {
					return HelpCenterConfig{}, fmt.Errorf("invalid url for tutorial %q step attachment", tutorial.ID)
				}
			}
		}
		if tutorial.CodeBlocks == nil {
			tutorial.CodeBlocks = []HelpCenterCodeBlock{}
		}
		for j := range tutorial.CodeBlocks {
			tutorial.CodeBlocks[j].Title = strings.TrimSpace(tutorial.CodeBlocks[j].Title)
			tutorial.CodeBlocks[j].Language = strings.TrimSpace(tutorial.CodeBlocks[j].Language)
		}
		if tutorial.Links == nil {
			tutorial.Links = []HelpCenterLink{}
		}
		for j := range tutorial.Links {
			tutorial.Links[j].Label = strings.TrimSpace(tutorial.Links[j].Label)
			tutorial.Links[j].URL = strings.TrimSpace(tutorial.Links[j].URL)
			if tutorial.Links[j].Label == "" || tutorial.Links[j].URL == "" {
				return HelpCenterConfig{}, fmt.Errorf("tutorial %q link label and url are required", tutorial.ID)
			}
			if !isAllowedHelpCenterURL(tutorial.Links[j].URL) {
				return HelpCenterConfig{}, fmt.Errorf("invalid url for tutorial %q link", tutorial.ID)
			}
		}
		if tutorial.Attachments == nil {
			tutorial.Attachments = []HelpCenterAttachment{}
		}
		for j := range tutorial.Attachments {
			tutorial.Attachments[j].Label = strings.TrimSpace(tutorial.Attachments[j].Label)
			tutorial.Attachments[j].URL = strings.TrimSpace(tutorial.Attachments[j].URL)
			tutorial.Attachments[j].FileName = strings.TrimSpace(tutorial.Attachments[j].FileName)
			if tutorial.Attachments[j].Label == "" || tutorial.Attachments[j].URL == "" {
				return HelpCenterConfig{}, fmt.Errorf("tutorial %q attachment label and url are required", tutorial.ID)
			}
			if !isAllowedHelpCenterURL(tutorial.Attachments[j].URL) {
				return HelpCenterConfig{}, fmt.Errorf("invalid url for tutorial %q attachment", tutorial.ID)
			}
		}
	}
	sort.SliceStable(cfg.Tutorials, func(i, j int) bool {
		if cfg.Tutorials[i].SortOrder == cfg.Tutorials[j].SortOrder {
			return cfg.Tutorials[i].ID < cfg.Tutorials[j].ID
		}
		return cfg.Tutorials[i].SortOrder < cfg.Tutorials[j].SortOrder
	})

	if cfg.FAQs == nil {
		cfg.FAQs = []HelpCenterFAQ{}
	}
	seenFAQs := make(map[string]struct{}, len(cfg.FAQs))
	for i := range cfg.FAQs {
		faq := &cfg.FAQs[i]
		faq.ID = strings.ToLower(strings.TrimSpace(faq.ID))
		if !helpCenterIDPattern.MatchString(faq.ID) {
			return HelpCenterConfig{}, fmt.Errorf("invalid faq id %q", faq.ID)
		}
		if _, ok := seenFAQs[faq.ID]; ok {
			return HelpCenterConfig{}, fmt.Errorf("duplicate faq id %q", faq.ID)
		}
		seenFAQs[faq.ID] = struct{}{}
		faq.Question = strings.TrimSpace(faq.Question)
		faq.AnswerMD = strings.TrimSpace(faq.AnswerMD)
		if faq.Enabled && (faq.Question == "" || faq.AnswerMD == "") {
			return HelpCenterConfig{}, fmt.Errorf("faq %q question and answer are required", faq.ID)
		}
		tags := faq.Tags[:0]
		seenTags := map[string]struct{}{}
		for _, tag := range faq.Tags {
			tag = strings.TrimSpace(tag)
			if tag == "" {
				continue
			}
			if _, ok := seenTags[tag]; ok {
				continue
			}
			seenTags[tag] = struct{}{}
			tags = append(tags, tag)
		}
		faq.Tags = tags
	}
	sort.SliceStable(cfg.FAQs, func(i, j int) bool {
		if cfg.FAQs[i].SortOrder == cfg.FAQs[j].SortOrder {
			return cfg.FAQs[i].ID < cfg.FAQs[j].ID
		}
		return cfg.FAQs[i].SortOrder < cfg.FAQs[j].SortOrder
	})

	return cfg, nil
}

func (s *HelpCenterService) GetDraft(ctx context.Context) (HelpCenterConfig, error) {
	return s.getConfig(ctx, SettingKeyHelpCenterDraftConfig)
}

func (s *HelpCenterService) SaveDraft(ctx context.Context, cfg HelpCenterConfig) error {
	normalized, err := NormalizeHelpCenterConfig(cfg)
	if err != nil {
		return err
	}
	return s.saveConfig(ctx, SettingKeyHelpCenterDraftConfig, normalized)
}

func (s *HelpCenterService) GetPublished(ctx context.Context) (HelpCenterConfig, error) {
	return s.getConfig(ctx, SettingKeyHelpCenterPublishedConfig)
}

func (s *HelpCenterService) PublishDraft(ctx context.Context) (HelpCenterConfig, error) {
	draft, err := s.GetDraft(ctx)
	if err != nil {
		return HelpCenterConfig{}, err
	}
	normalized, err := NormalizeHelpCenterConfig(draft)
	if err != nil {
		return HelpCenterConfig{}, err
	}
	if err := s.saveConfig(ctx, SettingKeyHelpCenterPublishedConfig, normalized); err != nil {
		return HelpCenterConfig{}, err
	}
	return normalized, nil
}

func (s *HelpCenterService) getConfig(ctx context.Context, key string) (HelpCenterConfig, error) {
	if s == nil || s.settingRepo == nil {
		return normalizedDefaultHelpCenterConfig(), nil
	}
	value, err := s.settingRepo.GetValue(ctx, key)
	if err != nil {
		if err == ErrSettingNotFound {
			return normalizedDefaultHelpCenterConfig(), nil
		}
		return HelpCenterConfig{}, err
	}
	var cfg HelpCenterConfig
	if err := json.Unmarshal([]byte(value), &cfg); err != nil {
		return normalizedDefaultHelpCenterConfig(), nil
	}
	normalized, err := NormalizeHelpCenterConfig(cfg)
	if err != nil {
		return normalizedDefaultHelpCenterConfig(), nil
	}
	return normalized, nil
}

func (s *HelpCenterService) saveConfig(ctx context.Context, key string, cfg HelpCenterConfig) error {
	data, err := json.Marshal(cfg)
	if err != nil {
		return err
	}
	return s.settingRepo.Set(ctx, key, string(data))
}

func isAllowedHelpCenterURL(raw string) bool {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return true
	}
	if strings.HasPrefix(raw, "/") && !strings.HasPrefix(raw, "//") {
		return true
	}
	parsed, err := url.Parse(raw)
	if err != nil || parsed.Scheme == "" || parsed.Host == "" {
		return false
	}
	return parsed.Scheme == "http" || parsed.Scheme == "https"
}
