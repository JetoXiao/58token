<template>
  <BaseDialog
    :show="show"
    :title="t('keys.useKeyModal.title')"
    width="wide"
    @close="emit('close')"
  >
    <div class="space-y-4">
      <!-- No Group Assigned Warning -->
      <div v-if="!platform" class="flex items-start gap-3 rounded-lg border border-accent-200 bg-accent-50 p-4 dark:border-accent-800/40 dark:bg-accent-900/20">
        <svg class="mt-0.5 h-5 w-5 flex-shrink-0 text-accent-500" fill="none" stroke="currentColor" viewBox="0 0 24 24" stroke-width="1.5">
          <path stroke-linecap="round" stroke-linejoin="round" d="M12 9v3.75m-9.303 3.376c-.866 1.5.217 3.374 1.948 3.374h14.71c1.73 0 2.813-1.874 1.948-3.374L13.949 3.378c-.866-1.5-3.032-1.5-3.898 0L2.697 16.126zM12 15.75h.007v.008H12v-.008z" />
        </svg>
        <div>
          <p class="text-sm font-medium text-accent-800 dark:text-accent-200">
            {{ t('keys.useKeyModal.noGroupTitle') }}
          </p>
          <p class="mt-1 text-sm text-accent-700 dark:text-accent-300">
            {{ t('keys.useKeyModal.noGroupDescription') }}
          </p>
        </div>
      </div>

      <!-- Platform-specific content -->
      <template v-else>
        <!-- Description -->
        <p class="text-sm text-gray-600 dark:text-gray-400">
          {{ platformDescription }}
        </p>

        <!-- Client Tabs -->
        <div v-if="clientTabs.length" class="border-b border-gray-200 dark:border-dark-700">
          <nav class="-mb-px flex space-x-6" aria-label="Client">
            <button
              v-for="tab in clientTabs"
              :key="tab.id"
              @click="activeClientTab = tab.id"
              :class="[
                'whitespace-nowrap py-2.5 px-1 border-b-2 font-medium text-sm transition-colors',
                activeClientTab === tab.id
                  ? 'border-primary-500 text-primary-600 dark:text-primary-400'
                  : 'border-transparent text-gray-500 hover:text-gray-700 hover:border-gray-300 dark:text-gray-400 dark:hover:text-gray-300'
              ]"
            >
              <span class="flex items-center gap-2">
                <component :is="tab.icon" class="w-4 h-4" />
                {{ tab.label }}
              </span>
            </button>
          </nav>
        </div>

        <!-- OS/Shell Tabs -->
        <div v-if="showShellTabs" class="border-b border-gray-200 dark:border-dark-700">
          <nav class="-mb-px flex space-x-4" aria-label="Tabs">
            <button
              v-for="tab in currentTabs"
              :key="tab.id"
              @click="activeTab = tab.id"
              :class="[
                'whitespace-nowrap py-2.5 px-1 border-b-2 font-medium text-sm transition-colors',
                activeTab === tab.id
                  ? 'border-primary-500 text-primary-600 dark:text-primary-400'
                  : 'border-transparent text-gray-500 hover:text-gray-700 hover:border-gray-300 dark:text-gray-400 dark:hover:text-gray-300'
              ]"
            >
              <span class="flex items-center gap-2">
                <component :is="tab.icon" class="w-4 h-4" />
                {{ tab.label }}
              </span>
            </button>
          </nav>
        </div>

        <!-- Codex Import Script -->
        <div
          v-if="activeClientTab === 'codex-import'"
          class="space-y-4 rounded-lg border border-accent-200 bg-accent-50/70 p-4 dark:border-accent-800/40 dark:bg-accent-900/10"
        >
          <div class="space-y-2 text-sm leading-6 text-gray-700 dark:text-gray-300 [&>p:nth-last-child(-n+3)]:rounded-md [&>p:nth-last-child(-n+3)]:border [&>p:nth-last-child(-n+3)]:border-accent-200 [&>p:nth-last-child(-n+3)]:bg-accent-100/80 [&>p:nth-last-child(-n+3)]:px-3 [&>p:nth-last-child(-n+3)]:py-2 [&>p:nth-last-child(-n+3)]:font-semibold [&>p:nth-last-child(-n+3)]:text-accent-900 dark:[&>p:nth-last-child(-n+3)]:border-accent-700/70 dark:[&>p:nth-last-child(-n+3)]:bg-accent-900/30 dark:[&>p:nth-last-child(-n+3)]:text-accent-100">
            <p class="font-medium text-gray-900 dark:text-gray-100">Codex 一键导入脚本</p>
            <p>下载后双击运行 BAT 文件，根据窗口提示输入选项即可完成导入。</p>
            <p>选项 1：导入配置文件至用户目录%USERPROFILE%\.codex，已有文件备份为.bak。</p>
            <p>选项 2：只适用于配置文件导入后仍报错，且用户目录包含中文路径时。</p>
            <p>💡 温馨提示：本脚本安全无毒</p>
            <p>💡 如果弹出 “Windows 已保护你的电脑” 提示框，请不要惊慌，这是所有浏览器下载脚本的常规安全提醒。</p>
            <p>👉 解决方法：点击隐藏的 「更多信息」 链接，然后点击 「仍要运行」 按钮即可安全导入。</p>
          </div>
          <button
            type="button"
            class="btn w-fit bg-primary-600 text-white shadow-sm hover:bg-primary-700 focus:ring-2 focus:ring-primary-500 focus:ring-offset-2 dark:focus:ring-offset-dark-900"
            @click="downloadCodexImportScript"
          >
            下载导入脚本
          </button>
        </div>

        <!-- Code Blocks (Stacked for multi-file platforms) -->
        <div v-else class="space-y-4">
          <div
            v-for="(file, index) in currentFiles"
            :key="index"
            class="relative"
          >
            <!-- File Hint (if exists) -->
            <p v-if="file.hint" class="mb-1.5 flex items-center gap-1 text-xs text-accent-600 dark:text-accent-300">
              <Icon name="exclamationCircle" size="sm" class="flex-shrink-0" />
              {{ file.hint }}
            </p>
            <div class="bg-gray-900 dark:bg-dark-900 rounded-xl overflow-hidden">
              <!-- Code Header -->
              <div class="flex items-center justify-between px-4 py-2 bg-gray-800 dark:bg-dark-800 border-b border-gray-700 dark:border-dark-700">
                <span class="text-xs text-gray-400 font-mono">{{ file.path }}</span>
                <button
                  @click="copyContent(file.content, index)"
                  class="flex items-center gap-1.5 px-2.5 py-1 text-xs font-medium rounded-lg transition-colors"
                  :class="copiedIndex === index
                    ? 'bg-primary-500/20 text-primary-400'
                    : 'bg-gray-700 hover:bg-gray-600 text-gray-300 hover:text-white'"
                >
                  <svg v-if="copiedIndex === index" class="w-3.5 h-3.5" fill="none" stroke="currentColor" viewBox="0 0 24 24" stroke-width="2">
                    <path stroke-linecap="round" stroke-linejoin="round" d="M5 13l4 4L19 7" />
                  </svg>
                  <svg v-else class="w-3.5 h-3.5" fill="none" stroke="currentColor" viewBox="0 0 24 24" stroke-width="1.5">
                    <path stroke-linecap="round" stroke-linejoin="round" d="M15.666 3.888A2.25 2.25 0 0013.5 2.25h-3c-1.03 0-1.9.693-2.166 1.638m7.332 0c.055.194.084.4.084.612v0a.75.75 0 01-.75.75H9a.75.75 0 01-.75-.75v0c0-.212.03-.418.084-.612m7.332 0c.646.049 1.288.11 1.927.184 1.1.128 1.907 1.077 1.907 2.185V19.5a2.25 2.25 0 01-2.25 2.25H6.75A2.25 2.25 0 014.5 19.5V6.257c0-1.108.806-2.057 1.907-2.185a48.208 48.208 0 011.927-.184" />
                  </svg>
                  {{ copiedIndex === index ? t('keys.useKeyModal.copied') : t('keys.useKeyModal.copy') }}
                </button>
              </div>
              <!-- Code Content -->
              <pre class="p-4 text-sm font-mono text-gray-100 overflow-x-auto"><code v-if="file.highlighted" v-html="file.highlighted"></code><code v-else v-text="file.content"></code></pre>
            </div>
          </div>
        </div>

        <!-- Usage Note -->
        <div v-if="showPlatformNote" class="flex items-start gap-3 rounded-lg border border-primary-100 bg-primary-50 p-3 dark:border-primary-900/30 dark:bg-primary-900/20">
          <Icon name="infoCircle" size="md" class="mt-0.5 flex-shrink-0 text-primary-500" />
          <p class="text-sm text-primary-700 dark:text-primary-300">
            {{ platformNote }}
          </p>
        </div>
      </template>
    </div>

    <template #footer>
      <div class="flex justify-end">
        <button
          @click="emit('close')"
          class="btn btn-secondary"
        >
          {{ t('common.close') }}
        </button>
      </div>
    </template>
  </BaseDialog>
</template>

<script setup lang="ts">
import { ref, computed, h, watch, type Component } from 'vue'
import { useI18n } from 'vue-i18n'
import BaseDialog from '@/components/common/BaseDialog.vue'
import Icon from '@/components/icons/Icon.vue'
import { useClipboard } from '@/composables/useClipboard'
import type { GroupPlatform } from '@/types'

interface Props {
  show: boolean
  apiKey: string
  baseUrl: string
  platform: GroupPlatform | null
  allowMessagesDispatch?: boolean
  allowImageGeneration?: boolean
}

interface Emits {
  (e: 'close'): void
}

interface TabConfig {
  id: string
  label: string
  icon: Component
}

interface FileConfig {
  path: string
  content: string
  hint?: string  // Optional hint message for this file
  highlighted?: string
}

const props = defineProps<Props>()
const emit = defineEmits<Emits>()

const { t } = useI18n()
const { copyToClipboard: clipboardCopy } = useClipboard()

const copiedIndex = ref<number | null>(null)
const activeTab = ref<string>('unix')
const activeClientTab = ref<string>('claude')

// Reset tabs when platform changes
const defaultClientTab = computed(() => {
  switch (props.platform) {
    case 'openai':
      return 'codex'
    case 'gemini':
      return 'gemini'
    case 'antigravity':
      return 'claude'
    default:
      return 'claude'
  }
})

watch(() => props.platform, () => {
  activeTab.value = 'unix'
  activeClientTab.value = defaultClientTab.value
}, { immediate: true })

// Reset shell tab when client changes
watch(activeClientTab, () => {
  activeTab.value = 'unix'
})

// Icon components
const AppleIcon = {
  render() {
    return h('svg', {
      fill: 'currentColor',
      viewBox: '0 0 24 24',
      class: 'w-4 h-4'
    }, [
      h('path', { d: 'M18.71 19.5c-.83 1.24-1.71 2.45-3.05 2.47-1.34.03-1.77-.79-3.29-.79-1.53 0-2 .77-3.27.82-1.31.05-2.3-1.32-3.14-2.53C4.25 17 2.94 12.45 4.7 9.39c.87-1.52 2.43-2.48 4.12-2.51 1.28-.02 2.5.87 3.29.87.78 0 2.26-1.07 3.81-.91.65.03 2.47.26 3.64 1.98-.09.06-2.17 1.28-2.15 3.81.03 3.02 2.65 4.03 2.68 4.04-.03.07-.42 1.44-1.38 2.83M13 3.5c.73-.83 1.94-1.46 2.94-1.5.13 1.17-.34 2.35-1.04 3.19-.69.85-1.83 1.51-2.95 1.42-.15-1.15.41-2.35 1.05-3.11z' })
    ])
  }
}

const WindowsIcon = {
  render() {
    return h('svg', {
      fill: 'currentColor',
      viewBox: '0 0 24 24',
      class: 'w-4 h-4'
    }, [
      h('path', { d: 'M3 12V6.75l6-1.32v6.48L3 12zm17-9v8.75l-10 .15V5.21L20 3zM3 13l6 .09v6.81l-6-1.15V13zm7 .25l10 .15V21l-10-1.91v-5.84z' })
    ])
  }
}

// Terminal icon for Claude Code
const TerminalIcon = {
  render() {
    return h('svg', {
      fill: 'none',
      stroke: 'currentColor',
      viewBox: '0 0 24 24',
      'stroke-width': '1.5',
      class: 'w-4 h-4'
    }, [
      h('path', {
        'stroke-linecap': 'round',
        'stroke-linejoin': 'round',
        d: 'm6.75 7.5 3 2.25-3 2.25m4.5 0h3m-9 8.25h13.5A2.25 2.25 0 0 0 21 17.25V6.75A2.25 2.25 0 0 0 18.75 4.5H5.25A2.25 2.25 0 0 0 3 6.75v10.5A2.25 2.25 0 0 0 5.25 20.25Z'
      })
    ])
  }
}

// Sparkle icon for Gemini
const SparkleIcon = {
  render() {
    return h('svg', {
      fill: 'none',
      stroke: 'currentColor',
      viewBox: '0 0 24 24',
      'stroke-width': '1.5',
      class: 'w-4 h-4'
    }, [
      h('path', {
        'stroke-linecap': 'round',
        'stroke-linejoin': 'round',
        d: 'M9.813 15.904 9 18.75l-.813-2.846a4.5 4.5 0 0 0-3.09-3.09L2.25 12l2.846-.813a4.5 4.5 0 0 0 3.09-3.09L9 5.25l.813 2.846a4.5 4.5 0 0 0 3.09 3.09L15.75 12l-2.846.813a4.5 4.5 0 0 0-3.09 3.09ZM18.259 8.715 18 9.75l-.259-1.035a3.375 3.375 0 0 0-2.455-2.456L14.25 6l1.036-.259a3.375 3.375 0 0 0 2.455-2.456L18 2.25l.259 1.035a3.375 3.375 0 0 0 2.456 2.456L21.75 6l-1.035.259a3.375 3.375 0 0 0-2.456 2.456ZM16.894 20.567 16.5 21.75l-.394-1.183a2.25 2.25 0 0 0-1.423-1.423L13.5 18.75l1.183-.394a2.25 2.25 0 0 0 1.423-1.423l.394-1.183.394 1.183a2.25 2.25 0 0 0 1.423 1.423l1.183.394-1.183.394a2.25 2.25 0 0 0-1.423 1.423Z'
      })
    ])
  }
}

const clientTabs = computed((): TabConfig[] => {
  if (!props.platform) return []
  switch (props.platform) {
    case 'openai': {
      const tabs: TabConfig[] = [
        { id: 'codex', label: t('keys.useKeyModal.cliTabs.codexCli'), icon: TerminalIcon },
        { id: 'codex-ws', label: t('keys.useKeyModal.cliTabs.codexCliWs'), icon: TerminalIcon },
        { id: 'codex-import', label: 'Codex 一键导入', icon: TerminalIcon },
      ]
      if (props.allowMessagesDispatch) {
        tabs.push({ id: 'claude', label: t('keys.useKeyModal.cliTabs.claudeCode'), icon: TerminalIcon })
      }
      tabs.push({ id: 'opencode', label: t('keys.useKeyModal.cliTabs.opencode'), icon: TerminalIcon })
      return tabs
    }
    case 'gemini':
      return [
        { id: 'gemini', label: t('keys.useKeyModal.cliTabs.geminiCli'), icon: SparkleIcon },
        { id: 'opencode', label: t('keys.useKeyModal.cliTabs.opencode'), icon: TerminalIcon }
      ]
    case 'antigravity':
      return [
        { id: 'claude', label: t('keys.useKeyModal.cliTabs.claudeCode'), icon: TerminalIcon },
        { id: 'gemini', label: t('keys.useKeyModal.cliTabs.geminiCli'), icon: SparkleIcon },
        { id: 'opencode', label: t('keys.useKeyModal.cliTabs.opencode'), icon: TerminalIcon }
      ]
    default:
      return [
        { id: 'claude', label: t('keys.useKeyModal.cliTabs.claudeCode'), icon: TerminalIcon },
        { id: 'opencode', label: t('keys.useKeyModal.cliTabs.opencode'), icon: TerminalIcon }
      ]
  }
})

// Shell tabs (3 types for environment variable based configs)
const shellTabs: TabConfig[] = [
  { id: 'unix', label: 'macOS / Linux', icon: AppleIcon },
  { id: 'cmd', label: 'Windows CMD', icon: WindowsIcon },
  { id: 'powershell', label: 'PowerShell', icon: WindowsIcon }
]

// OpenAI tabs (2 OS types)
const openaiTabs: TabConfig[] = [
  { id: 'unix', label: 'macOS / Linux', icon: AppleIcon },
  { id: 'windows', label: 'Windows', icon: WindowsIcon }
]

const showShellTabs = computed(() => activeClientTab.value !== 'opencode' && activeClientTab.value !== 'codex-import')

const currentTabs = computed(() => {
  if (!showShellTabs.value) return []
  if (activeClientTab.value === 'codex' || activeClientTab.value === 'codex-ws') {
    return openaiTabs
  }
  return shellTabs
})

const platformDescription = computed(() => {
  switch (props.platform) {
    case 'openai':
      if (activeClientTab.value === 'codex-import') {
        return '下载 Windows 一键导入脚本，适用于 Codex CLI 和 Codex App。'
      }
      if (activeClientTab.value === 'claude') {
        return t('keys.useKeyModal.description')
      }
      return t('keys.useKeyModal.openai.description')
    case 'gemini':
      return t('keys.useKeyModal.gemini.description')
    case 'antigravity':
      return t('keys.useKeyModal.antigravity.description')
    default:
      return t('keys.useKeyModal.description')
  }
})

const platformNote = computed(() => {
  switch (props.platform) {
    case 'openai':
      if (activeClientTab.value === 'claude') {
        return t('keys.useKeyModal.note')
      }
      return activeTab.value === 'windows'
        ? t('keys.useKeyModal.openai.noteWindows')
        : t('keys.useKeyModal.openai.note')
    case 'gemini':
      return t('keys.useKeyModal.gemini.note')
    case 'antigravity':
      return activeClientTab.value === 'claude'
        ? t('keys.useKeyModal.antigravity.claudeNote')
        : t('keys.useKeyModal.antigravity.geminiNote')
    default:
      return t('keys.useKeyModal.note')
  }
})

const showPlatformNote = computed(() => activeClientTab.value !== 'opencode' && activeClientTab.value !== 'codex-import')

const escapeHtml = (value: string) => value
  .replace(/&/g, '&amp;')
  .replace(/</g, '&lt;')
  .replace(/>/g, '&gt;')
  .replace(/"/g, '&quot;')
  .replace(/'/g, '&#39;')

const wrapToken = (className: string, value: string) =>
  `<span class="${className}">${escapeHtml(value)}</span>`

const keyword = (value: string) => wrapToken('text-primary-300', value)
const variable = (value: string) => wrapToken('text-cyan-200', value)
const operator = (value: string) => wrapToken('text-slate-400', value)
const string = (value: string) => wrapToken('text-accent-200', value)
const comment = (value: string) => wrapToken('text-slate-500', value)

const CLAUDE_CODE_DEFAULT_MODEL = 'claude-opus-4-8'
const CLAUDE_CODE_TIMEOUT_MS = '3000000'
const CODEX_DEFAULT_MODEL = 'gpt-5.5'
const CODEX_REVIEW_MODEL = 'gpt-5.4'
const CODEX_REASONING_EFFORT = 'medium'
const OPENAI_IMAGE_MODEL = 'gpt-image-2'
const OPENAI_IMAGE_DEFAULT_SIZE = '1K'
const OPENCODE_TIMEOUT_MS = 3000000

const isOpenAIImageGenerationGroup = computed(() =>
  props.platform === 'openai' && props.allowImageGeneration === true
)

const claudeCodeEnabledPlugins = {
  'commit-commands@claude-plugins-official': true,
  'context7@claude-plugins-official': true,
  'frontend-design@claude-plugins-official': true,
  'playwright@claude-plugins-official': true,
  'pyright-lsp@claude-plugins-official': true,
  'superpowers@claude-plugins-official': true
}

// Syntax highlighting helpers
// Generate file configs based on platform and active tab
const currentFiles = computed((): FileConfig[] => {
  const baseUrl = props.baseUrl || window.location.origin
  const apiKey = props.apiKey
  const baseRoot = baseUrl.replace(/\/v1\/?$/, '').replace(/\/+$/, '')
  const ensureV1 = (value: string) => {
    const trimmed = value.replace(/\/+$/, '')
    return trimmed.endsWith('/v1') ? trimmed : `${trimmed}/v1`
  }
  const apiBase = ensureV1(baseRoot)
  const antigravityBase = ensureV1(`${baseRoot}/antigravity`)
  const antigravityGeminiBase = (() => {
    const trimmed = `${baseRoot}/antigravity`.replace(/\/+$/, '')
    return trimmed.endsWith('/v1beta') ? trimmed : `${trimmed}/v1beta`
  })()
  const geminiBase = (() => {
    const trimmed = baseRoot.replace(/\/+$/, '')
    return trimmed.endsWith('/v1beta') ? trimmed : `${trimmed}/v1beta`
  })()

  if (activeClientTab.value === 'opencode') {
    switch (props.platform) {
      case 'anthropic':
        return [generateOpenCodeConfig('anthropic', apiBase, apiKey)]
      case 'openai':
        return [generateOpenCodeConfig('openai', apiBase, apiKey, undefined, isOpenAIImageGenerationGroup.value)]
      case 'gemini':
        return [generateOpenCodeConfig('gemini', geminiBase, apiKey)]
      case 'antigravity':
        return [
          generateOpenCodeConfig('antigravity-claude', antigravityBase, apiKey, 'opencode.json (Claude)'),
          generateOpenCodeConfig('antigravity-gemini', antigravityGeminiBase, apiKey, 'opencode.json (Gemini)')
        ]
      default:
        return [generateOpenCodeConfig('openai', apiBase, apiKey)]
    }
  }

  switch (props.platform) {
    case 'openai':
      if (activeClientTab.value === 'claude') {
        return generateAnthropicFiles(baseUrl, apiKey)
      }
      if (activeClientTab.value === 'codex-ws') {
        if (isOpenAIImageGenerationGroup.value) {
          return generateOpenAIImageFiles(baseUrl, apiKey, { supportsWebsockets: true })
        }
        return generateOpenAIWsFiles(baseUrl, apiKey)
      }
      if (isOpenAIImageGenerationGroup.value) {
        return generateOpenAIImageFiles(baseUrl, apiKey)
      }
      return generateOpenAIFiles(baseUrl, apiKey)
    case 'gemini':
      return [generateGeminiCliContent(baseUrl, apiKey)]
    case 'antigravity':
      if (activeClientTab.value === 'gemini') {
        return [generateGeminiCliContent(`${baseUrl}/antigravity`, apiKey)]
      }
      return generateAnthropicFiles(`${baseUrl}/antigravity`, apiKey)
    default:
      return generateAnthropicFiles(baseUrl, apiKey)
  }
})

function generateAnthropicFiles(baseUrl: string, apiKey: string): FileConfig[] {
  let path: string
  let content: string
  const env = buildClaudeCodeEnv(baseUrl, apiKey)

  switch (activeTab.value) {
    case 'unix':
      path = 'Terminal'
      content = renderUnixEnv(env)
      break
    case 'cmd':
      path = 'Command Prompt'
      content = renderCmdEnv(env)
      break
    case 'powershell':
      path = 'PowerShell'
      content = renderPowerShellEnv(env)
      break
    default:
      path = 'Terminal'
      content = ''
  }

  const vscodeSettingsPath = activeTab.value === 'unix'
    ? '~/.claude/settings.json'
    : '%userprofile%\\.claude\\settings.json'

  const vscodeContent = JSON.stringify(
    {
      enabledPlugins: claudeCodeEnabledPlugins,
      env,
      includeCoAuthoredBy: false
    },
    null,
    2
  )

  return [
    { path, content },
    { path: vscodeSettingsPath, content: vscodeContent, hint: 'VSCode Claude Code' }
  ]
}

function buildClaudeCodeEnv(baseUrl: string, apiKey: string): Record<string, string> {
  return {
    ANTHROPIC_AUTH_TOKEN: apiKey,
    ANTHROPIC_BASE_URL: baseUrl,
    ANTHROPIC_DEFAULT_HAIKU_MODEL: CLAUDE_CODE_DEFAULT_MODEL,
    ANTHROPIC_DEFAULT_OPUS_MODEL: CLAUDE_CODE_DEFAULT_MODEL,
    ANTHROPIC_DEFAULT_SONNET_MODEL: CLAUDE_CODE_DEFAULT_MODEL,
    ANTHROPIC_MODEL: CLAUDE_CODE_DEFAULT_MODEL,
    API_TIMEOUT_MS: CLAUDE_CODE_TIMEOUT_MS,
    CLAUDE_CODE_ATTRIBUTION_HEADER: '0',
    CLAUDE_CODE_DISABLE_NONESSENTIAL_TRAFFIC: '1'
  }
}

function renderUnixEnv(env: Record<string, string>): string {
  return Object.entries(env)
    .map(([key, value]) => `export ${key}="${escapeShellDoubleQuoted(value)}"`)
    .join('\n')
}

function renderCmdEnv(env: Record<string, string>): string {
  return Object.entries(env)
    .map(([key, value]) => `set ${key}=${escapeCmdValue(value)}`)
    .join('\n')
}

function renderPowerShellEnv(env: Record<string, string>): string {
  return Object.entries(env)
    .map(([key, value]) => `$env:${key}="${escapePowerShellDoubleQuoted(value)}"`)
    .join('\n')
}

function escapeShellDoubleQuoted(value: string): string {
  return value.replace(/(["\\$`])/g, '\\$1')
}

function escapeCmdValue(value: string): string {
  return value.replace(/([&|<>^])/g, '^$1')
}

function escapePowerShellDoubleQuoted(value: string): string {
  return value.replace(/`/g, '``').replace(/"/g, '`"')
}

function generateGeminiCliContent(baseUrl: string, apiKey: string): FileConfig {
  const model = 'gemini-2.0-flash'
  const modelComment = t('keys.useKeyModal.gemini.modelComment')
  let path: string
  let content: string
  let highlighted: string

  switch (activeTab.value) {
    case 'unix':
      path = 'Terminal'
      content = `export GOOGLE_GEMINI_BASE_URL="${baseUrl}"
export GEMINI_API_KEY="${apiKey}"
export GEMINI_MODEL="${model}"  # ${modelComment}`
      highlighted = `${keyword('export')} ${variable('GOOGLE_GEMINI_BASE_URL')}${operator('=')}${string(`"${baseUrl}"`)}
${keyword('export')} ${variable('GEMINI_API_KEY')}${operator('=')}${string(`"${apiKey}"`)}
${keyword('export')} ${variable('GEMINI_MODEL')}${operator('=')}${string(`"${model}"`)}  ${comment(`# ${modelComment}`)}`
      break
    case 'cmd':
      path = 'Command Prompt'
      content = `set GOOGLE_GEMINI_BASE_URL=${baseUrl}
set GEMINI_API_KEY=${apiKey}
set GEMINI_MODEL=${model}`
      highlighted = `${keyword('set')} ${variable('GOOGLE_GEMINI_BASE_URL')}${operator('=')}${string(baseUrl)}
${keyword('set')} ${variable('GEMINI_API_KEY')}${operator('=')}${string(apiKey)}
${keyword('set')} ${variable('GEMINI_MODEL')}${operator('=')}${string(model)}
${comment(`REM ${modelComment}`)}`
      break
    case 'powershell':
      path = 'PowerShell'
      content = `$env:GOOGLE_GEMINI_BASE_URL="${baseUrl}"
$env:GEMINI_API_KEY="${apiKey}"
$env:GEMINI_MODEL="${model}"  # ${modelComment}`
      highlighted = `${keyword('$env:')}${variable('GOOGLE_GEMINI_BASE_URL')}${operator('=')}${string(`"${baseUrl}"`)}
${keyword('$env:')}${variable('GEMINI_API_KEY')}${operator('=')}${string(`"${apiKey}"`)}
${keyword('$env:')}${variable('GEMINI_MODEL')}${operator('=')}${string(`"${model}"`)}  ${comment(`# ${modelComment}`)}`
      break
    default:
      path = 'Terminal'
      content = ''
      highlighted = ''
  }

  return { path, content, highlighted }
}

function generateOpenAIFiles(baseUrl: string, apiKey: string): FileConfig[] {
  const isWindows = activeTab.value === 'windows'
  const configDir = isWindows ? '%userprofile%\\.codex' : '~/.codex'

  const { configContent, authContent } = buildCodexConfigFiles(baseUrl, apiKey, false)

  return [
    {
      path: `${configDir}/config.toml`,
      content: configContent,
      hint: t('keys.useKeyModal.openai.configTomlHint')
    },
    {
      path: `${configDir}/auth.json`,
      content: authContent
    }
  ]
}

function generateOpenAIWsFiles(baseUrl: string, apiKey: string): FileConfig[] {
  const isWindows = activeTab.value === 'windows'
  const configDir = isWindows ? '%userprofile%\\.codex' : '~/.codex'
  const { configContent, authContent } = buildCodexConfigFiles(baseUrl, apiKey, true)

  return [
    {
      path: `${configDir}/config.toml`,
      content: configContent,
      hint: t('keys.useKeyModal.openai.configTomlHint')
    },
    {
      path: `${configDir}/auth.json`,
      content: authContent
    }
  ]
}

function buildCodexConfigFiles(baseUrl: string, apiKey: string, supportsWebsockets: boolean) {
  const websocketLine = supportsWebsockets ? 'supports_websockets = true\n' : ''
  const configContent = `model_provider = "OpenAI"
model = "${CODEX_DEFAULT_MODEL}"
review_model = "${CODEX_REVIEW_MODEL}"
model_reasoning_effort = "${CODEX_REASONING_EFFORT}"
disable_response_storage = true
network_access = "enabled"
windows_wsl_setup_acknowledged = true
model_context_window = 1000000
model_auto_compact_token_limit = 900000

[model_providers.OpenAI]
name = "OpenAI"
base_url = "${baseUrl}"
wire_api = "responses"
${websocketLine}requires_openai_auth = true`

  const authContent = `{
  "OPENAI_API_KEY": "${apiKey}"
}`

  return { configContent, authContent }
}

function generateOpenAIImageFiles(
  baseUrl: string,
  apiKey: string,
  options: { supportsWebsockets?: boolean } = {}
): FileConfig[] {
  const isWindows = activeTab.value === 'windows'
  const configDir = isWindows ? '%userprofile%\\.codex' : '~/.codex'
  const websocketLine = options.supportsWebsockets
    ? '\nsupports_websockets = true'
    : ''

  const configContent = `model_provider = "OpenAI"
model = "${OPENAI_IMAGE_MODEL}"
disable_response_storage = true
network_access = "enabled"
windows_wsl_setup_acknowledged = true

[model_providers.OpenAI]
name = "OpenAI"
base_url = "${baseUrl}"
wire_api = "responses"${websocketLine}
requires_openai_auth = true`

  const authContent = `{
  "OPENAI_API_KEY": "${apiKey}"
}`

  const requestContent = JSON.stringify(
    {
      model: OPENAI_IMAGE_MODEL,
      prompt: 'Generate a cute orange cat astronaut sticker on a clean pastel background.',
      n: 1,
      size: OPENAI_IMAGE_DEFAULT_SIZE,
      response_format: 'b64_json'
    },
    null,
    2
  )

  return [
    {
      path: `${configDir}/config.toml`,
      content: configContent,
      hint: t('keys.useKeyModal.openai.configTomlHint')
    },
    {
      path: `${configDir}/auth.json`,
      content: authContent
    },
    {
      path: 'image-generation-request.json',
      content: requestContent
    }
  ]
}

function generateOpenCodeConfig(
  platform: string,
  baseUrl: string,
  apiKey: string,
  pathLabel?: string,
  useOpenAIImageConfig = false
): FileConfig {
  const provider: Record<string, any> = {
    [platform]: {
      options: {
        baseURL: baseUrl,
        apiKey,
        timeout: OPENCODE_TIMEOUT_MS,
        headerTimeout: OPENCODE_TIMEOUT_MS,
        chunkTimeout: OPENCODE_TIMEOUT_MS
      }
    }
  }
  const openaiImageModels = {
    [OPENAI_IMAGE_MODEL]: {
      name: 'GPT Image 2',
      modalities: {
        input: ['text', 'image'],
        output: ['image']
      },
      options: {
        store: false
      }
    }
  }
  const openaiModels = useOpenAIImageConfig ? openaiImageModels : {
    'gpt-5.2': {
      name: 'GPT-5.2',
      limit: {
        context: 400000,
        output: 128000
      },
      options: {
        store: false
      },
      variants: {
        low: {},
        medium: {},
        high: {},
        xhigh: {}
      }
    },
    'gpt-5.6-sol': {
      name: 'GPT-5.6 Sol',
      limit: {
        context: 1050000,
        output: 128000
      },
      options: {
        store: false
      },
      variants: {
        low: {},
        medium: {},
        high: {},
        xhigh: {}
      }
    },
    'gpt-5.6-terra': {
      name: 'GPT-5.6 Terra',
      limit: {
        context: 1050000,
        output: 128000
      },
      options: {
        store: false
      },
      variants: {
        low: {},
        medium: {},
        high: {},
        xhigh: {}
      }
    },
    'gpt-5.6-luna': {
      name: 'GPT-5.6 Luna',
      limit: {
        context: 1050000,
        output: 128000
      },
      options: {
        store: false
      },
      variants: {
        low: {},
        medium: {},
        high: {},
        xhigh: {}
      }
    },
    'gpt-5.5': {
      name: 'GPT-5.5',
      limit: {
        context: 1050000,
        output: 128000
      },
      options: {
        store: false
      },
      variants: {
        low: {},
        medium: {},
        high: {},
        xhigh: {}
      }
    },
    'gpt-5.4': {
      name: 'GPT-5.4',
      limit: {
        context: 1050000,
        output: 128000
      },
      options: {
        store: false
      },
      variants: {
        low: {},
        medium: {},
        high: {},
        xhigh: {}
      }
    },
    'gpt-5.4-mini': {
      name: 'GPT-5.4 Mini',
      limit: {
        context: 400000,
        output: 128000
      },
      options: {
        store: false
      },
      variants: {
        low: {},
        medium: {},
        high: {},
        xhigh: {}
      }
    },
    'gpt-5.3-codex-spark': {
      name: 'GPT-5.3 Codex Spark',
      limit: {
        context: 128000,
        output: 32000
      },
      options: {
        store: false
      },
      variants: {
        low: {},
        medium: {},
        high: {},
        xhigh: {}
      }
    },
    'gpt-5.3-codex': {
      name: 'GPT-5.3 Codex',
      limit: {
        context: 400000,
        output: 128000
      },
      options: {
        store: false
      },
      variants: {
        low: {},
        medium: {},
        high: {},
        xhigh: {}
      }
    },
    'codex-mini-latest': {
      name: 'Codex Mini',
      limit: {
        context: 200000,
        output: 100000
      },
      options: {
        store: false
      },
      variants: {
        low: {},
        medium: {},
        high: {}
      }
    }
  }
  const geminiModels = {
    'gemini-2.0-flash': {
      name: 'Gemini 2.0 Flash',
      limit: {
        context: 1048576,
        output: 65536
      },
      modalities: {
        input: ['text', 'image', 'pdf'],
        output: ['text']
      }
    },
    'gemini-2.5-flash': {
      name: 'Gemini 2.5 Flash',
      limit: {
        context: 1048576,
        output: 65536
      },
      modalities: {
        input: ['text', 'image', 'pdf'],
        output: ['text']
      }
    },
    'gemini-2.5-pro': {
      name: 'Gemini 2.5 Pro',
      limit: {
        context: 2097152,
        output: 65536
      },
      modalities: {
        input: ['text', 'image', 'pdf'],
        output: ['text']
      },
      options: {
        thinking: {
          budgetTokens: 24576,
          type: 'enabled'
        }
      }
    },
    'gemini-3.5-flash': {
      name: 'Gemini 3.5 Flash',
      limit: {
        context: 1048576,
        output: 65536
      },
      modalities: {
        input: ['text', 'image', 'pdf'],
        output: ['text']
      }
    },
    'gemini-3-flash-preview': {
      name: 'Gemini 3 Flash Preview',
      limit: {
        context: 1048576,
        output: 65536
      },
      modalities: {
        input: ['text', 'image', 'pdf'],
        output: ['text']
      }
    },
    'gemini-3-pro-preview': {
      name: 'Gemini 3 Pro Preview',
      limit: {
        context: 1048576,
        output: 65536
      },
      modalities: {
        input: ['text', 'image', 'pdf'],
        output: ['text']
      },
      options: {
        thinking: {
          budgetTokens: 24576,
          type: 'enabled'
        }
      }
    },
    'gemini-3.1-pro-preview': {
      name: 'Gemini 3.1 Pro Preview',
      limit: {
        context: 1048576,
        output: 65536
      },
      modalities: {
        input: ['text', 'image', 'pdf'],
        output: ['text']
      },
      options: {
        thinking: {
          budgetTokens: 24576,
          type: 'enabled'
        }
      }
    }
  }

  const antigravityGeminiModels = {
    'gemini-2.5-flash': {
      name: 'Gemini 2.5 Flash',
      limit: {
        context: 1048576,
        output: 65536
      },
      modalities: {
        input: ['text', 'image', 'pdf'],
        output: ['text']
      },
      options: {
        thinking: {
          budgetTokens: 24576,
          type: 'disable'
        }
      }
    },
    'gemini-2.5-flash-lite': {
      name: 'Gemini 2.5 Flash Lite',
      limit: {
        context: 1048576,
        output: 65536
      },
      modalities: {
        input: ['text', 'image', 'pdf'],
        output: ['text']
      },
      options: {
        thinking: {
          budgetTokens: 24576,
          type: 'enabled'
        }
      }
    },
    'gemini-2.5-flash-thinking': {
      name: 'Gemini 2.5 Flash (Thinking)',
      limit: {
        context: 1048576,
        output: 65536
      },
      modalities: {
        input: ['text', 'image', 'pdf'],
        output: ['text']
      },
      options: {
        thinking: {
          budgetTokens: 24576,
          type: 'enabled'
        }
      }
    },
    'gemini-3-flash': {
      name: 'Gemini 3 Flash',
      limit: {
        context: 1048576,
        output: 65536
      },
      modalities: {
        input: ['text', 'image', 'pdf'],
        output: ['text']
      },
      options: {
        thinking: {
          budgetTokens: 24576,
          type: 'enabled'
        }
      }
    },
    'gemini-3.1-pro-low': {
      name: 'Gemini 3.1 Pro Low',
      limit: {
        context: 1048576,
        output: 65536
      },
      modalities: {
        input: ['text', 'image', 'pdf'],
        output: ['text']
      },
      options: {
        thinking: {
          budgetTokens: 24576,
          type: 'enabled'
        }
      }
    },
    'gemini-3.1-pro-high': {
      name: 'Gemini 3.1 Pro High',
      limit: {
        context: 1048576,
        output: 65536
      },
      modalities: {
        input: ['text', 'image', 'pdf'],
        output: ['text']
      },
      options: {
        thinking: {
          budgetTokens: 24576,
          type: 'enabled'
        }
      }
    },
    'gemini-2.5-flash-image': {
      name: 'Gemini 2.5 Flash Image',
      limit: {
        context: 1048576,
        output: 65536
      },
      modalities: {
        input: ['text', 'image'],
        output: ['image']
      },
      options: {
        thinking: {
          budgetTokens: 24576,
          type: 'enabled'
        }
      }
    },
    'gemini-3.1-flash-image': {
      name: 'Gemini 3.1 Flash Image',
      limit: {
        context: 1048576,
        output: 65536
      },
      modalities: {
        input: ['text', 'image'],
        output: ['image']
      },
      options: {
        thinking: {
          budgetTokens: 24576,
          type: 'enabled'
        }
      }
    }
  }
  const claudeModels = {
    'claude-opus-4-8': {
      name: 'Claude Opus 4.8',
      limit: {
        context: 200000,
        output: 128000
      },
      modalities: {
        input: ['text', 'image', 'pdf'],
        output: ['text']
      },
      options: {
        thinking: {
          budgetTokens: 24576,
          type: 'enabled'
        }
      }
    },
    'claude-opus-4-6-thinking': {
      name: 'Claude 4.6 Opus (Thinking)',
      limit: {
        context: 200000,
        output: 128000
      },
      modalities: {
        input: ['text', 'image', 'pdf'],
        output: ['text']
      },
      options: {
        thinking: {
          budgetTokens: 24576,
          type: 'enabled'
        }
      }
    },
    'claude-sonnet-4-6': {
      name: 'Claude 4.6 Sonnet',
      limit: {
        context: 200000,
        output: 64000
      },
      modalities: {
        input: ['text', 'image', 'pdf'],
        output: ['text']
      },
      options: {
        thinking: {
          budgetTokens: 24576,
          type: 'enabled'
        }
      }
    }
  }

  if (platform === 'gemini') {
    provider[platform].npm = '@ai-sdk/google'
    provider[platform].models = geminiModels
  } else if (platform === 'anthropic') {
    provider[platform].npm = '@ai-sdk/anthropic'
    provider[platform].models = claudeModels
  } else if (platform === 'antigravity-claude') {
    provider[platform].npm = '@ai-sdk/anthropic'
    provider[platform].name = 'Antigravity (Claude)'
    provider[platform].models = claudeModels
  } else if (platform === 'antigravity-gemini') {
    provider[platform].npm = '@ai-sdk/google'
    provider[platform].name = 'Antigravity (Gemini)'
    provider[platform].models = antigravityGeminiModels
  } else if (platform === 'openai') {
    provider[platform].models = openaiModels
  }

  const agent =
    platform === 'openai'
      ? {
          build: {
            options: {
              store: false
            }
          },
          plan: {
            options: {
              store: false
            }
          }
        }
      : undefined
  const defaultModel = getOpenCodeDefaultModel(platform, useOpenAIImageConfig)

  const content = JSON.stringify(
    {
      model: defaultModel.model,
      small_model: defaultModel.smallModel,
      provider,
      ...(agent ? { agent } : {}),
      $schema: 'https://opencode.ai/config.json'
    },
    null,
    2
  )

  return {
    path: pathLabel ?? 'opencode.json',
    content,
    hint: t('keys.useKeyModal.opencode.hint')
  }
}

function getOpenCodeDefaultModel(platform: string, useOpenAIImageConfig = false): { model: string; smallModel: string } {
  switch (platform) {
    case 'anthropic':
      return {
        model: 'anthropic/claude-opus-4-8',
        smallModel: 'anthropic/claude-opus-4-8'
      }
    case 'openai':
      if (useOpenAIImageConfig) {
        return {
          model: `openai/${OPENAI_IMAGE_MODEL}`,
          smallModel: `openai/${OPENAI_IMAGE_MODEL}`
        }
      }
      return {
        model: 'openai/gpt-5.5',
        smallModel: 'openai/gpt-5.4-mini'
      }
    case 'gemini':
      return {
        model: 'gemini/gemini-2.5-flash',
        smallModel: 'gemini/gemini-2.5-flash'
      }
    case 'antigravity-claude':
      return {
        model: 'antigravity-claude/claude-opus-4-8',
        smallModel: 'antigravity-claude/claude-opus-4-8'
      }
    case 'antigravity-gemini':
      return {
        model: 'antigravity-gemini/gemini-2.5-flash',
        smallModel: 'antigravity-gemini/gemini-2.5-flash'
      }
    default:
      return {
        model: 'openai/gpt-5.5',
        smallModel: 'openai/gpt-5.4-mini'
      }
  }
}

const copyContent = async (content: string, index: number) => {
  const success = await clipboardCopy(content, t('keys.copied'))
  if (success) {
    copiedIndex.value = index
    setTimeout(() => {
      copiedIndex.value = null
    }, 2000)
  }
}

function encodeUtf8Base64(value: string): string {
  const bytes = new TextEncoder().encode(value)
  let binary = ''
  bytes.forEach((byte) => {
    binary += String.fromCharCode(byte)
  })
  return btoa(binary)
}

function getCodexImportScriptContent(): string {
  const baseUrl = props.baseUrl || window.location.origin
  const apiKey = props.apiKey
  const { configContent, authContent } = buildCodexConfigFiles(baseUrl, apiKey, false)
  const configBase64 = encodeUtf8Base64(configContent)
  const authBase64 = encodeUtf8Base64(authContent)

  return `@echo off
chcp 65001 >nul
set "CODEX_IMPORT_BAT=%~f0"
powershell.exe -NoLogo -NoProfile -Command "try { $ErrorActionPreference='Stop'; $p=$env:CODEX_IMPORT_BAT; $c=[System.IO.File]::ReadAllText($p,[System.Text.Encoding]::UTF8); $m='# POWERSHELL-CODE-START'; $i=$c.LastIndexOf($m); if($i -lt 0){ throw 'PowerShell script block was not found.' }; $s=$c.Substring($i+$m.Length); $block=[scriptblock]::Create($s); $block.Invoke() } catch { Write-Host ''; Write-Host ('启动失败：' + $_.Exception.Message); exit 1 }"
set "CODEX_IMPORT_EXIT=%ERRORLEVEL%"
if not "%CODEX_IMPORT_EXIT%"=="0" (
  echo.
  echo PowerShell 启动或解析失败，请将本窗口错误信息截图反馈给管理员。
  pause
)
exit /b %CODEX_IMPORT_EXIT%

# POWERSHELL-CODE-START
$ErrorActionPreference = 'Stop'
[Console]::OutputEncoding = [System.Text.Encoding]::UTF8
$OutputEncoding = [System.Text.Encoding]::UTF8

$ConfigBase64 = '${configBase64}'
$AuthBase64 = '${authBase64}'

function Decode-Base64Text {
  param([string]$Value)
  return [System.Text.Encoding]::UTF8.GetString([Convert]::FromBase64String($Value))
}

function Write-Utf8NoBomFile {
  param(
    [string]$Path,
    [string]$Content
  )
  $dir = Split-Path -Parent $Path
  if (-not (Test-Path -LiteralPath $dir)) {
    New-Item -ItemType Directory -Path $dir -Force | Out-Null
  }
  $bytes = [System.Text.Encoding]::UTF8.GetBytes($Content)
  [IO.File]::WriteAllBytes($Path, $bytes)
}

function Backup-FileIfExists {
  param([string]$Path)
  if (Test-Path -LiteralPath $Path) {
    Copy-Item -LiteralPath $Path -Destination "$Path.bak" -Force
    Write-Host "已备份：$Path.bak"
  }
}

function Write-CodexFiles {
  param(
    [string]$ConfigDir,
    [bool]$BackupExisting
  )
  if (-not (Test-Path -LiteralPath $ConfigDir)) {
    New-Item -ItemType Directory -Path $ConfigDir -Force | Out-Null
    Write-Host "已创建目录：$ConfigDir"
  }

  $configPath = Join-Path $ConfigDir 'config.toml'
  $authPath = Join-Path $ConfigDir 'auth.json'
  if ($BackupExisting) {
    Backup-FileIfExists $configPath
    Backup-FileIfExists $authPath
  }

  Write-Utf8NoBomFile $configPath (Decode-Base64Text $ConfigBase64)
  Write-Utf8NoBomFile $authPath (Decode-Base64Text $AuthBase64)
  Write-Host "已写入：$configPath"
  Write-Host "已写入：$authPath"
}

function Test-ReparsePath {
  param([string]$Path)
  $item = Get-Item -LiteralPath $Path -Force -ErrorAction SilentlyContinue
  return $null -ne $item -and (($item.Attributes -band [IO.FileAttributes]::ReparsePoint) -ne 0)
}

function Move-ExistingDefaultPath {
  param([string]$Path)
  if (-not (Test-Path -LiteralPath $Path)) {
    return $false
  }
  if (Test-ReparsePath $Path) {
    Write-Host "检测到默认路径已是目录联接：$Path"
    return $true
  }

  $parent = Split-Path -Parent $Path
  $stamp = Get-Date -Format 'yyyyMMdd-HHmmss'
  $backupName = ".codex.local-$stamp"
  Rename-Item -LiteralPath $Path -NewName $backupName
  Write-Host "已将原默认目录重命名为：$(Join-Path $parent $backupName)"
  return $false
}

function Ensure-Junction {
  param(
    [string]$LinkPath,
    [string]$TargetPath
  )
  $alreadyLinked = Move-ExistingDefaultPath $LinkPath
  if ($alreadyLinked) {
    Write-Host "目录联接已存在，未重复创建。"
    return
  }

  $output = cmd.exe /d /c "mklink /J ""$LinkPath"" ""$TargetPath""" 2>&1
  if ($LASTEXITCODE -ne 0) {
    $message = ($output | Out-String).Trim()
    throw "创建目录联接失败：$message"
  }
  $output | ForEach-Object { Write-Host $_ }
}

function Import-DefaultCodexConfig {
  $defaultDir = Join-Path $env:USERPROFILE '.codex'
  Write-CodexFiles $defaultDir $true
  Write-Host ''
  Write-Host '默认路径导入完成。请重新打开 Codex 后测试。'
}

function Import-EnglishCodexConfig {
  Write-Host '选项 2 用于默认路径导入后仍报错，尤其是用户目录包含中文路径时。'
  Write-Host '将写入 C:\\CodexConfig\\.codex，并把用户目录下的 .codex 创建为目录联接。'
  Write-Host '如果用户目录下已有普通 .codex 文件夹，会重命名为 .codex.local-时间戳，不会删除其中内容。'
  $confirm = Read-Host '确认执行请输入 y'
  if ($confirm -ne 'y') {
    Write-Host '已取消。'
    return
  }

  $targetDir = 'C:\\CodexConfig\\.codex'
  $linkDir = Join-Path $env:USERPROFILE '.codex'
  Write-CodexFiles $targetDir $false
  Ensure-Junction $linkDir $targetDir
  Write-Host ''
  Write-Host '英文路径导入完成。请重新打开 Codex 后测试。'
}

try {
  Write-Host 'Codex 一键导入脚本'
  Write-Host '===================='
  Write-Host '1. 默认路径导入：写入 %USERPROFILE%\\.codex，已有文件备份为 .bak 后覆盖'
  Write-Host '2. 英文路径修复：写入 C:\\CodexConfig\\.codex，并创建目录联接'
  Write-Host ''
  $choice = Read-Host '请输入选项 1 或 2'

  switch ($choice) {
    '1' { Import-DefaultCodexConfig }
    '2' { Import-EnglishCodexConfig }
    default { Write-Host '无效选项，未执行任何操作。' }
  }
} catch {
  Write-Host ''
  Write-Host "执行失败：$($_.Exception.Message)"
} finally {
  Write-Host ''
  Read-Host '按回车退出'
}
`
}

function downloadCodexImportScript() {
  const content = getCodexImportScriptContent().replace(/\r?\n/g, '\r\n')
  const blob = new Blob([content], { type: 'application/x-bat;charset=utf-8' })
  const url = URL.createObjectURL(blob)
  const link = document.createElement('a')
  link.href = url
  link.download = 'codex-import.bat'
  document.body.appendChild(link)
  link.click()
  link.remove()
  URL.revokeObjectURL(url)
}
</script>
