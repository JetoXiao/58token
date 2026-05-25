<template>
  <BaseDialog
    :show="show"
    :title="t('keys.testModal.title')"
    width="normal"
    @close="handleClose"
  >
    <div class="space-y-4">
      <div
        v-if="apiKey"
        class="rounded-xl border border-gray-200 bg-gradient-to-r from-gray-50 to-gray-100 p-3 dark:border-dark-500 dark:from-dark-700 dark:to-dark-600"
      >
        <div class="flex items-center justify-between gap-3">
          <div class="flex min-w-0 items-center gap-3">
            <div class="flex h-10 w-10 flex-shrink-0 items-center justify-center rounded-lg bg-gradient-to-br from-primary-500 to-primary-600">
              <Icon name="play" size="md" class="text-white" :stroke-width="2" />
            </div>
            <div class="min-w-0">
              <div class="truncate font-semibold text-gray-900 dark:text-gray-100">
                {{ apiKey.name }}
              </div>
              <div class="mt-0.5 flex flex-wrap items-center gap-1.5 text-xs text-gray-500 dark:text-gray-400">
                <span class="rounded bg-gray-200 px-1.5 py-0.5 font-mono text-[10px] dark:bg-dark-500">
                  {{ maskApiKey(apiKey.key) }}
                </span>
                <span v-if="apiKey.group">{{ apiKey.group.name }}</span>
              </div>
            </div>
          </div>
          <span
            :class="[
              'rounded-full px-2.5 py-1 text-xs font-semibold',
              apiKey.status === 'active'
                ? 'bg-green-100 text-green-700 dark:bg-green-500/20 dark:text-green-400'
                : 'bg-gray-100 text-gray-600 dark:bg-gray-700 dark:text-gray-400'
            ]"
          >
            {{ t(`keys.status.${apiKey.status}`) }}
          </span>
        </div>
      </div>

      <div
        v-if="apiKey && !apiKey.group"
        class="rounded-xl border border-yellow-200 bg-yellow-50 p-3 text-sm text-yellow-800 dark:border-yellow-800 dark:bg-yellow-900/20 dark:text-yellow-200"
      >
        {{ t('keys.testModal.noGroup') }}
      </div>

      <div
        v-else-if="apiKey && apiKey.status !== 'active'"
        class="rounded-xl border border-yellow-200 bg-yellow-50 p-3 text-sm text-yellow-800 dark:border-yellow-800 dark:bg-yellow-900/20 dark:text-yellow-200"
      >
        {{ t('keys.testModal.inactive') }}
      </div>

      <p class="text-sm text-gray-600 dark:text-gray-400">
        {{ t('keys.testModal.description') }}
      </p>

      <div class="space-y-1.5">
        <label class="text-sm font-medium text-gray-700 dark:text-gray-300">
          {{ t('keys.testModal.selectModel') }}
        </label>
        <Select
          v-model="selectedModelId"
          :options="availableModels"
          :disabled="status === 'connecting' || !canTest"
          :placeholder="t('keys.testModal.selectModel')"
        />
      </div>

      <div class="group relative">
        <div
          ref="terminalRef"
          class="max-h-[260px] min-h-[150px] overflow-y-auto rounded-xl border border-gray-700 bg-gray-900 p-4 font-mono text-sm dark:border-gray-800 dark:bg-black"
        >
          <div v-if="status === 'idle'" class="flex items-center gap-2 text-gray-500">
            <Icon name="play" size="sm" :stroke-width="2" />
            <span>{{ t('keys.testModal.ready') }}</span>
          </div>
          <div v-else-if="status === 'connecting'" class="flex items-center gap-2 text-yellow-400">
            <Icon name="refresh" size="sm" class="animate-spin" :stroke-width="2" />
            <span>{{ t('keys.testModal.testing') }}</span>
          </div>

          <div v-for="(line, index) in outputLines" :key="index" :class="line.class">
            {{ line.text }}
          </div>

          <div v-if="streamingContent" class="text-green-400">
            {{ streamingContent }}<span class="animate-pulse">_</span>
          </div>

          <div
            v-if="status === 'success'"
            class="mt-3 flex items-center gap-2 border-t border-gray-700 pt-3 text-green-400"
          >
            <Icon name="check" size="sm" :stroke-width="2" />
            <span>{{ t('keys.testModal.success') }}</span>
          </div>
          <div
            v-else-if="status === 'error'"
            class="mt-3 flex items-center gap-2 border-t border-gray-700 pt-3 text-red-400"
          >
            <Icon name="x" size="sm" :stroke-width="2" />
            <span>{{ errorMessage }}</span>
          </div>
        </div>

        <button
          v-if="outputLines.length > 0 || streamingContent"
          @click="copyOutput"
          class="absolute right-2 top-2 rounded-lg bg-gray-800/80 p-1.5 text-gray-400 opacity-0 transition-all hover:bg-gray-700 hover:text-white group-hover:opacity-100"
          :title="t('keys.testModal.copyOutput')"
        >
          <Icon name="copy" size="sm" :stroke-width="2" />
        </button>
      </div>

      <div class="flex flex-wrap items-center justify-between gap-2 px-1 text-xs text-gray-500 dark:text-gray-400">
        <span class="flex items-center gap-1">
          <Icon name="grid" size="sm" :stroke-width="2" />
          {{ endpointLabel }}
        </span>
        <span class="flex items-center gap-1">
          <Icon name="chat" size="sm" :stroke-width="2" />
          {{ t('keys.testModal.testPrompt') }}
        </span>
      </div>

      <p class="rounded-lg bg-blue-50 px-3 py-2 text-xs text-blue-700 dark:bg-blue-900/20 dark:text-blue-300">
        {{ t('keys.testModal.smallUsageHint') }}
      </p>
    </div>

    <template #footer>
      <div class="flex justify-end gap-3">
        <button
          @click="handleClose"
          class="rounded-lg bg-gray-100 px-4 py-2 text-sm font-medium text-gray-700 transition-colors hover:bg-gray-200 dark:bg-dark-600 dark:text-gray-300 dark:hover:bg-dark-500"
        >
          {{ t('common.close') }}
        </button>
        <button
          @click="startTest"
          :disabled="status === 'connecting' || !selectedModelId || !canTest"
          :class="[
            'flex items-center gap-2 rounded-lg px-4 py-2 text-sm font-medium transition-all',
            status === 'connecting' || !selectedModelId || !canTest
              ? 'cursor-not-allowed bg-primary-400/70 text-white'
              : status === 'success'
                ? 'bg-green-500 text-white hover:bg-green-600'
                : status === 'error'
                  ? 'bg-orange-500 text-white hover:bg-orange-600'
                  : 'bg-primary-500 text-white hover:bg-primary-600'
          ]"
        >
          <Icon
            v-if="status === 'connecting'"
            name="refresh"
            size="sm"
            class="animate-spin"
            :stroke-width="2"
          />
          <Icon v-else-if="status === 'idle'" name="play" size="sm" :stroke-width="2" />
          <Icon v-else name="refresh" size="sm" :stroke-width="2" />
          <span>
            {{
              status === 'connecting'
                ? t('keys.testModal.testing')
                : status === 'idle'
                  ? t('keys.testModal.start')
                  : t('keys.testModal.retry')
            }}
          </span>
        </button>
      </div>
    </template>
  </BaseDialog>
</template>

<script setup lang="ts">
import { computed, nextTick, ref, watch } from 'vue'
import { useI18n } from 'vue-i18n'
import BaseDialog from '@/components/common/BaseDialog.vue'
import Select from '@/components/common/Select.vue'
import Icon from '@/components/icons/Icon.vue'
import { useClipboard } from '@/composables/useClipboard'
import { maskApiKey } from '@/utils/maskApiKey'
import type { ApiKey, GroupPlatform } from '@/types'

interface OutputLine {
  text: string
  class: string
}

type ModelOption = Record<string, unknown> & {
  value: string
  label: string
}

const props = defineProps<{
  show: boolean
  apiKey: ApiKey | null
}>()

const emit = defineEmits<{
  (e: 'close'): void
}>()

const { t } = useI18n()
const { copyToClipboard } = useClipboard()

const terminalRef = ref<HTMLElement | null>(null)
const selectedModelId = ref('')
const status = ref<'idle' | 'connecting' | 'success' | 'error'>('idle')
const outputLines = ref<OutputLine[]>([])
const streamingContent = ref('')
const errorMessage = ref('')
let abortController: AbortController | null = null

const modelPresets: Record<GroupPlatform, ModelOption[]> = {
  anthropic: [
    { value: 'claude-sonnet-4-6', label: 'Claude Sonnet 4.6' },
    { value: 'claude-opus-4-7', label: 'Claude Opus 4.7' },
    { value: 'claude-opus-4-6', label: 'Claude Opus 4.6' },
    { value: 'claude-haiku-4-5', label: 'Claude Haiku 4.5' }
  ],
  openai: [
    { value: 'gpt-5.5', label: 'GPT-5.5' },
    { value: 'gpt-5.4', label: 'GPT-5.4' },
    { value: 'gpt-5.4-mini', label: 'GPT-5.4 Mini' },
    { value: 'gpt-5.2', label: 'GPT-5.2' },
    { value: 'gpt-5.3-codex', label: 'GPT-5.3 Codex' }
  ],
  gemini: [
    { value: 'gemini-2.5-flash', label: 'Gemini 2.5 Flash' },
    { value: 'gemini-2.5-pro', label: 'Gemini 2.5 Pro' },
    { value: 'gemini-3-pro-preview', label: 'Gemini 3 Pro Preview' }
  ],
  antigravity: [
    { value: 'claude-sonnet-4-6', label: 'Claude Sonnet 4.6' },
    { value: 'claude-opus-4-7', label: 'Claude Opus 4.7' },
    { value: 'gemini-2.5-flash', label: 'Gemini 2.5 Flash' },
    { value: 'gemini-2.5-pro', label: 'Gemini 2.5 Pro' }
  ]
}

const platform = computed(() => props.apiKey?.group?.platform ?? null)
const availableModels = computed(() => (platform.value ? modelPresets[platform.value] : []))
const canTest = computed(() => !!props.apiKey?.group && props.apiKey.status === 'active')
const endpoint = computed(() => (platform.value === 'openai' || platform.value === 'gemini' ? '/v1/chat/completions' : '/v1/messages'))
const endpointLabel = computed(() => `${t('keys.testModal.endpoint')}: ${endpoint.value}`)

watch(
  () => props.show,
  (visible) => {
    if (visible) {
      resetState()
      selectDefaultModel()
    } else {
      abortStream()
    }
  }
)

watch(platform, () => {
  selectDefaultModel()
})

const selectDefaultModel = () => {
  selectedModelId.value = availableModels.value[0]?.value ?? ''
}

const resetState = () => {
  status.value = 'idle'
  outputLines.value = []
  streamingContent.value = ''
  errorMessage.value = ''
}

const handleClose = () => {
  abortStream()
  emit('close')
}

const abortStream = () => {
  if (abortController) {
    abortController.abort()
    abortController = null
  }
}

const scrollToBottom = async () => {
  await nextTick()
  if (terminalRef.value) {
    terminalRef.value.scrollTop = terminalRef.value.scrollHeight
  }
}

const addLine = (text: string, className = 'text-gray-300') => {
  outputLines.value.push({ text, class: className })
  scrollToBottom()
}

const buildRequestBody = () => {
  if (endpoint.value === '/v1/chat/completions') {
    return {
      model: selectedModelId.value,
      messages: [{ role: 'user', content: 'hi' }],
      max_tokens: 32,
      stream: true
    }
  }

  return {
    model: selectedModelId.value,
    max_tokens: 32,
    stream: true,
    messages: [
      {
        role: 'user',
        content: [{ type: 'text', text: 'hi' }]
      }
    ]
  }
}

const startTest = async () => {
  if (!props.apiKey || !selectedModelId.value || !canTest.value) return

  resetState()
  status.value = 'connecting'
  addLine(t('keys.testModal.starting', { name: props.apiKey.name }), 'text-blue-400')
  addLine(t('keys.testModal.groupLine', { group: props.apiKey.group?.name || '-' }), 'text-gray-400')
  addLine(t('keys.testModal.usingModel', { model: selectedModelId.value }), 'text-cyan-400')
  addLine('', 'text-gray-300')

  abortStream()
  abortController = new AbortController()

  try {
    const response = await fetch(endpoint.value, {
      method: 'POST',
      headers: {
        Authorization: `Bearer ${props.apiKey.key}`,
        'Content-Type': 'application/json',
        'anthropic-version': '2023-06-01'
      },
      body: JSON.stringify(buildRequestBody()),
      signal: abortController.signal
    })

    if (!response.ok) {
      throw new Error(await extractGatewayError(response))
    }

    addLine(t('keys.testModal.connected'), 'text-green-400')
    addLine(t('keys.testModal.sending'), 'text-gray-400')
    addLine(t('keys.testModal.response'), 'text-yellow-400')

    await readStream(response)
    finalizeSuccess()
  } catch (error: unknown) {
    if (error instanceof DOMException && error.name === 'AbortError') {
      status.value = 'idle'
      return
    }

    const message = error instanceof Error ? error.message : String(error)
    status.value = 'error'
    errorMessage.value = message || t('keys.testModal.requestFailed')
    addLine(`Error: ${errorMessage.value}`, 'text-red-400')
  }
}

const readStream = async (response: Response) => {
  const reader = response.body?.getReader()
  if (!reader) {
    const text = await response.text()
    if (text) appendContent(text)
    return
  }

  const decoder = new TextDecoder()
  let buffer = ''

  while (true) {
    const { done, value } = await reader.read()
    if (done) break

    buffer += decoder.decode(value, { stream: true })
    const lines = buffer.split('\n')
    buffer = lines.pop() || ''

    for (const rawLine of lines) {
      const line = rawLine.trim()
      if (!line || line.startsWith(':') || line.startsWith('event:')) continue
      const payload = line.startsWith('data:') ? line.slice(5).trim() : line
      if (!payload || payload === '[DONE]') continue
      handleStreamPayload(payload)
    }
  }

  if (buffer.trim()) {
    handleStreamPayload(buffer.trim())
  }
}

const handleStreamPayload = (payload: string) => {
  try {
    const event = JSON.parse(payload)
    const text =
      event.delta?.text ||
      event.text ||
      event.choices?.[0]?.delta?.content ||
      event.choices?.[0]?.message?.content ||
      event.output_text ||
      event.content?.[0]?.text

    if (event.type === 'error' || event.error) {
      const message = event.error?.message || event.error || t('keys.testModal.requestFailed')
      throw new Error(message)
    }

    if (text) appendContent(text)
  } catch (error) {
    if (error instanceof SyntaxError) {
      return
    }
    throw error
  }
}

const appendContent = (text: string) => {
  streamingContent.value += text
  scrollToBottom()
}

const finalizeSuccess = () => {
  if (streamingContent.value) {
    addLine(streamingContent.value, 'text-green-300')
    streamingContent.value = ''
  }
  status.value = 'success'
}

const extractGatewayError = async (response: Response) => {
  const text = await response.text()
  if (!text) return `HTTP ${response.status}`

  try {
    const parsed = JSON.parse(text)
    const message = parsed.error?.message || parsed.error || parsed.message || parsed.detail
    if (message) return `HTTP ${response.status}: ${message}`
  } catch {
    // Fall through to raw body.
  }

  return `HTTP ${response.status}: ${text.slice(0, 500)}`
}

const copyOutput = () => {
  const text = [...outputLines.value.map((line) => line.text), streamingContent.value]
    .filter(Boolean)
    .join('\n')
  copyToClipboard(text, t('keys.testModal.outputCopied'))
}
</script>
