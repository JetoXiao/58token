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
          :disabled="status === 'connecting' || marketplaceLoading || !canTest || availableModels.length === 0"
          :placeholder="marketplaceLoading ? t('common.loading') : t('keys.testModal.selectModel')"
        />
      </div>

      <div
        v-if="selectedImageModelNeedsEnabled"
        class="rounded-xl border border-yellow-200 bg-yellow-50 p-3 text-sm text-yellow-800 dark:border-yellow-800 dark:bg-yellow-900/20 dark:text-yellow-200"
      >
        {{ t('keys.testModal.imageGenerationDisabled') }}
      </div>

      <div
        v-else-if="imageModelUnsupported"
        class="rounded-xl border border-yellow-200 bg-yellow-50 p-3 text-sm text-yellow-800 dark:border-yellow-800 dark:bg-yellow-900/20 dark:text-yellow-200"
      >
        {{ t('keys.testModal.unsupportedImageModel') }}
      </div>

      <div v-if="supportsImageTest" class="space-y-4">
        <TextArea
          v-model="testPrompt"
          :label="t('keys.testModal.imagePromptLabel')"
          :placeholder="t('keys.testModal.imagePromptPlaceholder')"
          :hint="t('keys.testModal.imageTestHint')"
          :disabled="status === 'connecting'"
          rows="3"
        />

        <div class="space-y-1.5">
          <label class="text-sm font-medium text-gray-700 dark:text-gray-300">
            {{ t('keys.testModal.imageSizeLabel') }}
          </label>
          <Select
            v-model="selectedImageSizeTier"
            :options="imageSizeOptions"
            :disabled="status === 'connecting'"
          />
        </div>
      </div>

      <div class="group relative">
        <div
          ref="terminalRef"
          class="max-h-[260px] min-h-[150px] overflow-y-auto rounded-xl border border-gray-700 bg-gray-900 p-4 font-mono text-sm dark:border-gray-800 dark:bg-black"
        >
          <div v-if="status === 'idle'" class="flex items-center gap-2 text-gray-500">
            <Icon name="play" size="sm" :stroke-width="2" />
            <span>{{ supportsImageTest ? t('keys.testModal.imageReady') : t('keys.testModal.ready') }}</span>
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

      <div v-if="generatedImages.length > 0" class="space-y-2">
        <div class="text-xs font-medium text-gray-600 dark:text-gray-300">
          {{ t('keys.testModal.imagePreview') }}
        </div>
        <div class="flex flex-wrap justify-center gap-3">
          <div
            v-for="(image, index) in generatedImages"
            :key="`${image.url}-${index}`"
            class="group/img relative cursor-pointer overflow-hidden rounded-xl border border-gray-200 bg-white shadow-sm transition hover:border-primary-300 hover:shadow-md dark:border-dark-500 dark:bg-dark-700"
            @click="previewImageUrl = image.url"
          >
            <img :src="image.url" :alt="`key-test-image-${index + 1}`" class="max-h-[360px] w-full object-contain" />
            <div class="absolute inset-0 flex items-center justify-center bg-black/0 transition-colors group-hover/img:bg-black/20">
              <Icon name="eye" size="lg" class="text-white opacity-0 drop-shadow-lg transition-opacity group-hover/img:opacity-100" :stroke-width="2" />
            </div>
            <div class="border-t border-gray-100 px-3 py-1.5 text-xs text-gray-500 dark:border-dark-500 dark:text-gray-300">
              {{ image.mimeType || 'image/*' }}
            </div>
          </div>
        </div>
      </div>

      <Teleport to="body">
        <Transition name="fade">
          <div
            v-if="previewImageUrl"
            class="fixed inset-0 z-[100] flex items-center justify-center bg-black/80 p-4"
            @click.self="previewImageUrl = ''"
          >
            <button
              class="absolute right-4 top-4 rounded-full bg-black/50 p-2 text-white transition-colors hover:bg-black/70"
              @click="previewImageUrl = ''"
            >
              <Icon name="x" size="lg" :stroke-width="2" />
            </button>
            <img
              :src="previewImageUrl"
              alt="preview"
              class="max-h-[90vh] max-w-[90vw] rounded-lg object-contain shadow-2xl"
            />
          </div>
        </Transition>
      </Teleport>

      <div class="flex flex-wrap items-center justify-between gap-2 px-1 text-xs text-gray-500 dark:text-gray-400">
        <span class="flex items-center gap-1">
          <Icon name="grid" size="sm" :stroke-width="2" />
          {{ endpointLabel }}
        </span>
        <span class="flex items-center gap-1">
          <Icon name="chat" size="sm" :stroke-width="2" />
          {{
            supportsImageTest
              ? t('keys.testModal.imageTestModeWithSize', { size: selectedImageSizeTier })
              : t('keys.testModal.testPrompt')
          }}
        </span>
      </div>

      <p class="rounded-lg bg-blue-50 px-3 py-2 text-xs text-blue-700 dark:bg-blue-900/20 dark:text-blue-300">
        {{ supportsImageTest ? t('keys.testModal.imageSmallUsageHint') : t('keys.testModal.smallUsageHint') }}
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
          :disabled="status === 'connecting' || !canStartTest"
          :class="[
            'flex items-center gap-2 rounded-lg px-4 py-2 text-sm font-medium transition-all',
            status === 'connecting' || !canStartTest
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
import TextArea from '@/components/common/TextArea.vue'
import Icon from '@/components/icons/Icon.vue'
import { useClipboard } from '@/composables/useClipboard'
import { maskApiKey } from '@/utils/maskApiKey'
import type { ApiKey, GroupPlatform } from '@/types'

interface OutputLine {
  text: string
  class: string
}

interface PreviewImage {
  url: string
  mimeType?: string
}

type ModelOption = Record<string, unknown> & {
  value: string
  label: string
}

type ImageSizeTier = '1K' | '2K' | '4K'

interface MarketplaceItemResponse {
  model_name: string
  pricing_aliases?: string[]
  vendor_name?: string
  groups?: string[]
  endpoints?: string[]
  sort_order?: number
  enabled?: boolean
}

interface MarketplaceResponse {
  items: MarketplaceItemResponse[]
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
const marketplaceItems = ref<MarketplaceItemResponse[]>([])
const marketplaceLoading = ref(false)
const testPrompt = ref('')
const selectedImageSizeTier = ref<ImageSizeTier>('1K')
const generatedImages = ref<PreviewImage[]>([])
const previewImageUrl = ref('')
let abortController: AbortController | null = null

const platform = computed(() => props.apiKey?.group?.platform ?? null)
const groupName = computed(() => props.apiKey?.group?.name ?? '')
const canTest = computed(() => !!props.apiKey?.group && props.apiKey.status === 'active')
const normalizedSelectedModelId = computed(() => normalizeModelId(selectedModelId.value))
const groupAllowsImageGeneration = computed(() => props.apiKey?.group?.allow_image_generation === true)
const isOpenAIImageModel = computed(() => normalizedSelectedModelId.value.startsWith('gpt-image-'))
const isGeminiImageModel = computed(() => normalizedSelectedModelId.value.startsWith('gemini-') && normalizedSelectedModelId.value.includes('image'))
const isSelectedImageModel = computed(() => isOpenAIImageModel.value || isGeminiImageModel.value)
const supportsOpenAIImageTest = computed(() => isOpenAIImageModel.value && platform.value === 'openai' && groupAllowsImageGeneration.value)
const supportsGeminiImageTest = computed(() => isGeminiImageModel.value && (platform.value === 'gemini' || platform.value === 'antigravity') && groupAllowsImageGeneration.value)
const supportsImageTest = computed(() => supportsOpenAIImageTest.value || supportsGeminiImageTest.value)
const selectedImageModelNeedsEnabled = computed(() => !!props.apiKey?.group && isSelectedImageModel.value && !groupAllowsImageGeneration.value)
const imageModelUnsupported = computed(() => isSelectedImageModel.value && groupAllowsImageGeneration.value && !supportsImageTest.value)
const canStartTest = computed(() => !!selectedModelId.value && canTest.value && !selectedImageModelNeedsEnabled.value && !imageModelUnsupported.value)
const textEndpoint = computed(() => (platform.value === 'openai' || platform.value === 'gemini' ? '/v1/chat/completions' : '/v1/messages'))
const endpoint = computed(() => {
  if (supportsOpenAIImageTest.value) return '/v1/images/generations'
  if (supportsGeminiImageTest.value) return buildGeminiImageEndpoint()
  return textEndpoint.value
})
const endpointLabel = computed(() => `${t('keys.testModal.endpoint')}: ${endpoint.value}`)
const imageSizeOptions = computed(() => [
  { value: '1K', label: t('keys.testModal.imageSize1k') },
  { value: '2K', label: t('keys.testModal.imageSize2k') },
  { value: '4K', label: t('keys.testModal.imageSize4k') }
])
const imageFallbackModelOptions = computed<ModelOption[]>(() => {
  if (!groupAllowsImageGeneration.value) return []

  switch (platform.value) {
    case 'openai':
      return [{
        value: 'gpt-image-2',
        label: 'GPT Image 2',
        source: 'image-fallback'
      }]
    case 'gemini':
    case 'antigravity':
      return [{
        value: 'gemini-2.5-flash-image',
        label: 'Gemini 2.5 Flash Image',
        source: 'image-fallback'
      }]
    default:
      return []
  }
})

const marketplaceModelOptions = computed<ModelOption[]>(() => {
  const enabledItems = marketplaceItems.value.filter((item) => item.enabled !== false && item.model_name)
  if (enabledItems.length === 0) return []

  const groupMatchedItems = groupName.value
    ? enabledItems.filter((item) => includesNormalized(item.groups, groupName.value))
    : []
  const scopedItems = groupMatchedItems.length > 0
    ? groupMatchedItems
    : enabledItems.filter((item) => isMarketplaceItemForPlatform(item, platform.value))

  const seen = new Set<string>()
  return scopedItems
    .slice()
    .sort((a, b) => (a.sort_order ?? 0) - (b.sort_order ?? 0) || a.model_name.localeCompare(b.model_name))
    .flatMap((item) => {
      const modelName = item.model_name.trim()
      const requestModelName = marketplaceRequestModelName(item)
      if (!modelName || !requestModelName || seen.has(requestModelName)) return []
      seen.add(requestModelName)
      return [{
        value: requestModelName,
        label: humanizeModelName(modelName),
        marketplaceModelName: modelName,
        vendorName: item.vendor_name,
        pricingAliases: item.pricing_aliases ?? []
      }]
    })
})

const availableModels = computed(() => mergeModelOptions([
  ...imageFallbackModelOptions.value,
  ...marketplaceModelOptions.value
]))

watch(
  () => props.show,
  (visible) => {
    if (visible) {
      resetState()
      selectedModelId.value = ''
      testPrompt.value = ''
      selectedImageSizeTier.value = '1K'
      loadMarketplaceModels()
    } else {
      abortStream()
    }
  }
)

watch(platform, () => {
  selectDefaultModel()
})

watch(groupName, () => {
  selectDefaultModel()
})

watch(availableModels, () => {
  selectDefaultModel()
})

watch([selectedModelId, supportsImageTest], () => {
  if (supportsImageTest.value && !testPrompt.value.trim()) {
    testPrompt.value = t('keys.testModal.imagePromptDefault')
  }
  if (!supportsImageTest.value) {
    generatedImages.value = []
    previewImageUrl.value = ''
  }
})

const selectDefaultModel = () => {
  if (availableModels.value.some((option) => option.value === selectedModelId.value)) return
  selectedModelId.value = availableModels.value[0]?.value ?? ''
}

const loadMarketplaceModels = async () => {
  marketplaceLoading.value = true
  try {
    const response = await fetch('/api/v1/public/model-marketplace', {
      headers: { Accept: 'application/json' }
    })
    if (!response.ok) throw new Error(`HTTP ${response.status}`)
    const payload = unwrapMarketplacePayload(await response.json())
    marketplaceItems.value = payload.items
  } catch (error) {
    console.error('Failed to load model marketplace items:', error)
    marketplaceItems.value = []
  } finally {
    marketplaceLoading.value = false
    selectDefaultModel()
  }
}

const unwrapMarketplacePayload = (payload: unknown): MarketplaceResponse => {
  if (!payload || typeof payload !== 'object') return { items: [] }
  const record = payload as Record<string, unknown>
  const data = record.data
  if (data && typeof data === 'object') {
    const dataRecord = data as Record<string, unknown>
    return { items: Array.isArray(dataRecord.items) ? dataRecord.items as MarketplaceItemResponse[] : [] }
  }
  return { items: Array.isArray(record.items) ? record.items as MarketplaceItemResponse[] : [] }
}

function normalizeText(value: string) {
  return value.trim().toLowerCase()
}

function normalizeModelId(value: string) {
  return normalizeText(value).replace(/^models\//, '')
}

const includesNormalized = (values: string[] | undefined, target: string) => {
  const normalizedTarget = normalizeText(target)
  return !!normalizedTarget && (values ?? []).some((value) => normalizeText(value) === normalizedTarget)
}

const marketplaceRequestModelName = (item: MarketplaceItemResponse) => {
  const aliases = Array.isArray(item.pricing_aliases) ? item.pricing_aliases : []
  const alias = aliases.map((value) => value.trim()).find(Boolean)
  return alias || item.model_name.trim()
}

function mergeModelOptions(options: ModelOption[]) {
  const seen = new Set<string>()
  return options.filter((option) => {
    const value = option.value.trim()
    if (!value || seen.has(value)) return false
    seen.add(value)
    return true
  })
}

const isMarketplaceItemForPlatform = (item: MarketplaceItemResponse, currentPlatform: GroupPlatform | null) => {
  const vendorName = normalizeText(item.vendor_name ?? '')
  switch (currentPlatform) {
    case 'anthropic':
      return vendorName === 'anthropic'
    case 'openai':
      return vendorName === 'openai'
    case 'gemini':
      return vendorName === 'google' || vendorName === 'gemini'
    case 'antigravity':
      return vendorName === 'anthropic' || vendorName === 'gemini' || vendorName === 'google'
    default:
      return false
  }
}

const titleCaseWords = (value: string) => value.replace(/\b[a-z]/g, (char) => char.toUpperCase())

const humanizeModelName = (modelName: string) => {
  if (modelName.startsWith('claude-')) {
    return `Claude ${titleCaseWords(modelName.slice('claude-'.length).replace(/-/g, ' '))}`
  }
  if (modelName.startsWith('gpt-')) {
    return `GPT-${titleCaseWords(modelName.slice('gpt-'.length).replace(/-/g, ' '))}`
  }
  if (modelName.startsWith('gemini-')) {
    return `Gemini ${titleCaseWords(modelName.slice('gemini-'.length).replace(/-/g, ' '))}`
  }
  return titleCaseWords(modelName.replace(/-/g, ' '))
}

const resetState = () => {
  status.value = 'idle'
  outputLines.value = []
  streamingContent.value = ''
  errorMessage.value = ''
  generatedImages.value = []
  previewImageUrl.value = ''
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

const buildTextRequestBody = () => {
  if (textEndpoint.value === '/v1/chat/completions') {
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
  if (!props.apiKey || !canStartTest.value) return
  if (supportsImageTest.value) {
    await startImageTest()
    return
  }
  await startTextTest()
}

const beginTest = (apiKey: ApiKey) => {
  resetState()
  status.value = 'connecting'
  addLine(t('keys.testModal.starting', { name: apiKey.name }), 'text-blue-400')
  addLine(t('keys.testModal.groupLine', { group: apiKey.group?.name || '-' }), 'text-gray-400')
  addLine(t('keys.testModal.usingModel', { model: selectedModelId.value }), 'text-cyan-400')
  if (supportsImageTest.value) {
    addLine(t('keys.testModal.usingImageSize', { size: selectedImageSizeTier.value }), 'text-cyan-300')
  }
  addLine('', 'text-gray-300')

  abortStream()
  abortController = new AbortController()
}

const startTextTest = async () => {
  const apiKey = props.apiKey
  if (!apiKey || !selectedModelId.value || !canTest.value) return

  beginTest(apiKey)

  try {
    const response = await fetch(textEndpoint.value, {
      method: 'POST',
      headers: {
        Authorization: `Bearer ${apiKey.key}`,
        'Content-Type': 'application/json',
        'anthropic-version': '2023-06-01'
      },
      body: JSON.stringify(buildTextRequestBody()),
      signal: abortController?.signal
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

const startImageTest = async () => {
  const apiKey = props.apiKey
  if (!apiKey || !selectedModelId.value || !supportsImageTest.value) return

  beginTest(apiKey)

  try {
    const response = await fetch(endpoint.value, {
      method: 'POST',
      headers: {
        Authorization: `Bearer ${apiKey.key}`,
        'Content-Type': 'application/json'
      },
      body: JSON.stringify(buildImageRequestBody()),
      signal: abortController?.signal
    })

    if (!response.ok) {
      throw new Error(await extractGatewayError(response))
    }

    addLine(t('keys.testModal.connected'), 'text-green-400')
    addLine(t('keys.testModal.sendingImage'), 'text-gray-400')
    addLine(t('keys.testModal.response'), 'text-yellow-400')

    const payload = await response.json()
    const images = supportsOpenAIImageTest.value
      ? extractOpenAIImages(payload)
      : extractGeminiImages(payload)

    if (images.length === 0) {
      throw new Error(t('keys.testModal.noImagesReturned'))
    }

    generatedImages.value = images
    images.forEach((image, index) => {
      addLine(t('keys.testModal.imageReceived', { count: index + 1 }), 'text-green-300')
      addLine(image.mimeType || 'image/*', 'text-gray-400')
    })
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

const buildImageRequestBody = () => {
  if (supportsOpenAIImageTest.value) {
    return {
      model: selectedModelId.value,
      prompt: imagePromptText(),
      n: 1,
      size: selectedImageSizeTier.value,
      response_format: 'b64_json'
    }
  }

  return {
    contents: [
      {
        role: 'user',
        parts: [{ text: imagePromptText() }]
      }
    ],
    generationConfig: {
      responseModalities: ['TEXT', 'IMAGE'],
      imageConfig: {
        aspectRatio: '1:1',
        imageSize: selectedImageSizeTier.value
      }
    }
  }
}

const imagePromptText = () => testPrompt.value.trim() || t('keys.testModal.imagePromptDefault')

const buildGeminiImageEndpoint = () => {
  const model = selectedModelId.value.trim().replace(/^models\//i, '')
  const prefix = platform.value === 'antigravity' ? '/antigravity/v1beta' : '/v1beta'
  return `${prefix}/models/${encodeURIComponent(model)}:generateContent`
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
    const event = JSON.parse(payload) as Record<string, any>
    const text =
      (event.type === 'response.output_text.delta' && typeof event.delta === 'string' ? event.delta : '') ||
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

const extractOpenAIImages = (payload: unknown): PreviewImage[] => {
  const root = asRecord(payload)
  if (!root) return []

  const images: PreviewImage[] = []
  collectOpenAIImageItems(root.data, images)
  collectOpenAIImageItems(root.output, images)

  const directImage = imageFromRecord(root)
  if (directImage) images.push(directImage)

  return images
}

const collectOpenAIImageItems = (value: unknown, images: PreviewImage[]) => {
  for (const item of asArray(value)) {
    const record = asRecord(item)
    if (!record) continue

    const image = imageFromRecord(record)
    if (image) {
      images.push(image)
      continue
    }

    collectOpenAIImageItems(record.content, images)
  }
}

const imageFromRecord = (record: Record<string, unknown>): PreviewImage | null => {
  const url = asString(record.url) || asString(record.image_url)
  if (url) return { url, mimeType: asString(record.mime_type) || asString(record.mimeType) || undefined }

  const base64 =
    asString(record.b64_json) ||
    asString(record.base64) ||
    asString(record.image_base64) ||
    asString(record.result)
  if (!base64) return null

  const outputFormat = asString(record.output_format)
  const mimeType = asString(record.mime_type) || asString(record.mimeType) || (outputFormat ? `image/${outputFormat}` : 'image/png')
  return { url: dataUrlFromBase64(base64, mimeType), mimeType }
}

const extractGeminiImages = (payload: unknown): PreviewImage[] => {
  const root = asRecord(payload)
  if (!root) return []

  const data = asRecord(root.response) ?? root
  const images: PreviewImage[] = []
  const texts: string[] = []

  for (const candidateValue of asArray(data.candidates)) {
    const candidate = asRecord(candidateValue)
    const content = asRecord(candidate?.content)
    for (const partValue of asArray(content?.parts)) {
      const part = asRecord(partValue)
      if (!part) continue

      const text = asString(part.text)
      if (text) texts.push(text)

      const inlineData = asRecord(part.inlineData) ?? asRecord(part.inline_data)
      if (!inlineData) continue

      const base64 = asString(inlineData.data)
      if (!base64) continue

      const mimeType = asString(inlineData.mimeType) || asString(inlineData.mime_type) || 'image/png'
      images.push({ url: dataUrlFromBase64(base64, mimeType), mimeType })
    }
  }

  texts.forEach((text) => addLine(text, 'text-green-300'))
  return images
}

const asRecord = (value: unknown): Record<string, unknown> | null =>
  value && typeof value === 'object' && !Array.isArray(value) ? value as Record<string, unknown> : null

const asArray = (value: unknown): unknown[] => Array.isArray(value) ? value : []

const asString = (value: unknown): string => typeof value === 'string' ? value.trim() : ''

const dataUrlFromBase64 = (base64: string, mimeType = 'image/png') =>
  base64.startsWith('data:') ? base64 : `data:${mimeType};base64,${base64}`

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

<style>
.fade-enter-active,
.fade-leave-active {
  transition: opacity 0.2s ease;
}

.fade-enter-from,
.fade-leave-to {
  opacity: 0;
}
</style>
