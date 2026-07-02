<template>
  <AppLayout>
    <div class="w-full space-y-6">
      <section class="rounded-lg border border-gray-200 bg-white p-6 shadow-sm dark:border-dark-700 dark:bg-dark-800">
        <div class="flex flex-col gap-4 lg:flex-row lg:items-start lg:justify-between">
          <div class="min-w-0">
            <p class="text-sm font-medium text-primary-600 dark:text-primary-400">{{ t('helpCenter.eyebrow') }}</p>
            <h1 class="mt-2 text-2xl font-semibold text-gray-900 dark:text-white">{{ config?.title || t('helpCenter.title') }}</h1>
            <p class="mt-2 max-w-3xl text-sm leading-6 text-gray-600 dark:text-gray-400">
              {{ config?.description || t('helpCenter.description') }}
            </p>
          </div>
          <div class="flex flex-wrap gap-2">
            <RouterLink to="/keys" class="btn btn-primary">
              <Icon name="key" size="sm" />
              {{ t('helpCenter.goToKeys') }}
            </RouterLink>
            <button type="button" class="btn btn-secondary" @click="loadHelpCenter">
              <Icon name="refresh" size="sm" />
              {{ t('common.refresh') }}
            </button>
          </div>
        </div>
      </section>

      <div v-if="loading" class="flex min-h-[320px] items-center justify-center rounded-lg border border-gray-200 bg-white dark:border-dark-700 dark:bg-dark-800">
        <div class="h-8 w-8 animate-spin rounded-full border-b-2 border-primary-600"></div>
      </div>

      <section v-else-if="!config?.enabled" class="rounded-lg border border-gray-200 bg-white p-8 text-center dark:border-dark-700 dark:bg-dark-800">
        <Icon name="book" size="xl" class="mx-auto text-gray-400" />
        <h2 class="mt-4 text-lg font-semibold text-gray-900 dark:text-white">{{ t('helpCenter.disabledTitle') }}</h2>
        <p class="mt-2 text-sm text-gray-500 dark:text-gray-400">{{ t('helpCenter.disabledDescription') }}</p>
      </section>

      <section v-else-if="!hasContent" class="rounded-lg border border-gray-200 bg-white p-8 text-center dark:border-dark-700 dark:bg-dark-800">
        <Icon name="document" size="xl" class="mx-auto text-gray-400" />
        <h2 class="mt-4 text-lg font-semibold text-gray-900 dark:text-white">{{ t('helpCenter.emptyTitle') }}</h2>
        <p class="mt-2 text-sm text-gray-500 dark:text-gray-400">{{ t('helpCenter.emptyDescription') }}</p>
      </section>

      <div v-else class="help-center-layout-container">
        <div :class="['help-center-layout', tutorials.length ? 'help-center-layout--with-nav' : '']">
          <aside v-if="tutorials.length" class="help-center-sidebar rounded-lg border border-gray-200 bg-white p-3 shadow-sm dark:border-dark-700 dark:bg-dark-800">
            <div class="help-center-client-nav">
              <div class="px-2 pb-3 text-xs font-semibold uppercase tracking-wide text-gray-500 dark:text-gray-400">
                {{ t('helpCenter.clients') }}
              </div>
              <div class="help-center-client-list">
                <button
                  v-for="tutorial in tutorials"
                  :key="tutorial.id"
                  type="button"
                  class="w-full rounded-md px-3 py-2.5 text-left transition"
                  :class="selectedTutorial?.id === tutorial.id
                    ? 'bg-primary-50 text-primary-700 dark:bg-primary-900/25 dark:text-primary-300'
                    : 'text-gray-700 hover:bg-gray-100 dark:text-gray-300 dark:hover:bg-dark-700'"
                  @click="selectedTutorialId = tutorial.id"
                >
                  <span class="flex items-center justify-between gap-3">
                    <span class="truncate text-sm font-medium">{{ tutorial.title }}</span>
                    <span v-if="tutorial.badge" class="shrink-0 rounded bg-gray-100 px-1.5 py-0.5 text-[11px] text-gray-500 dark:bg-dark-700 dark:text-gray-400">
                      {{ tutorial.badge }}
                    </span>
                  </span>
                  <span class="mt-1 line-clamp-2 block text-xs leading-5 text-gray-500 dark:text-gray-400">
                    {{ tutorial.summary }}
                  </span>
                </button>
              </div>
            </div>
          </aside>

          <div class="min-w-0 space-y-6">
            <article v-if="selectedTutorial && tutorials.length" class="min-w-0 rounded-lg border border-gray-200 bg-white p-6 shadow-sm dark:border-dark-700 dark:bg-dark-800">
            <header class="border-b border-gray-100 pb-5 dark:border-dark-700">
              <div class="flex flex-wrap items-center gap-2">
                <h2 class="text-2xl font-semibold text-gray-900 dark:text-white">{{ selectedTutorial.title }}</h2>
                <span v-if="selectedTutorial.badge" class="rounded bg-primary-50 px-2 py-1 text-xs font-medium text-primary-700 dark:bg-primary-900/30 dark:text-primary-300">
                  {{ selectedTutorial.badge }}
                </span>
              </div>
              <p v-if="selectedTutorial.summary" class="mt-2 text-sm leading-6 text-gray-600 dark:text-gray-400">
                {{ selectedTutorial.summary }}
              </p>
              <p v-if="config.base_url" class="mt-3 flex max-w-full flex-wrap items-center gap-2 rounded-md bg-gray-50 px-3 py-2 text-xs text-gray-600 dark:bg-dark-900 dark:text-gray-400">
                <Icon name="globe" size="xs" />
                {{ t('helpCenter.defaultBaseUrl') }}:
                <code class="break-all font-mono text-gray-900 dark:text-gray-200">{{ config.base_url }}</code>
              </p>
            </header>

            <div v-if="selectedTutorial.content_md" class="help-center-markdown mt-6" v-html="renderMarkdown(selectedTutorial.content_md)"></div>

            <section v-if="selectedTutorial.steps?.length" class="mt-8">
              <h3 class="text-base font-semibold text-gray-900 dark:text-white">{{ t('helpCenter.steps') }}</h3>
              <ol class="mt-4 space-y-3">
                <li
                  v-for="(step, index) in selectedTutorial.steps"
                  :key="`${selectedTutorial.id}-step-${index}`"
                  class="flex gap-3 rounded-lg border border-gray-100 bg-gray-50 p-4 dark:border-dark-700 dark:bg-dark-900"
                >
                  <span class="flex h-7 w-7 shrink-0 items-center justify-center rounded-full bg-primary-600 text-sm font-semibold text-white">
                    {{ index + 1 }}
                  </span>
                  <div class="min-w-0 flex-1">
                    <p class="text-sm font-semibold text-gray-900 dark:text-white">{{ step.title }}</p>
                    <p class="mt-1 whitespace-pre-line text-sm leading-6 text-gray-600 dark:text-gray-400">{{ step.description }}</p>

                    <div v-if="step.images?.length" class="mt-4 grid grid-cols-1 gap-3 sm:grid-cols-2">
                      <button
                        v-for="(image, imageIndex) in step.images"
                        :key="`${selectedTutorial.id}-step-${index}-image-${imageIndex}`"
                        type="button"
                        class="group overflow-hidden rounded-lg border border-gray-200 bg-white text-left shadow-sm transition hover:border-primary-200 hover:shadow-md focus:outline-none focus:ring-2 focus:ring-primary-500 focus:ring-offset-2 dark:border-dark-700 dark:bg-dark-800 dark:hover:border-primary-700 dark:focus:ring-offset-dark-900"
                        :disabled="!stepImageSrc(image)"
                        :aria-label="`${t('helpCenter.openImagePreview')}: ${image.label}`"
                        data-testid="help-center-step-image"
                        @click="openImagePreview(image)"
                      >
                        <div class="relative">
                          <img
                            v-if="stepImageSrc(image)"
                            :src="stepImageSrc(image)"
                            :alt="image.label"
                            class="aspect-video w-full bg-gray-100 object-contain dark:bg-dark-900"
                            loading="lazy"
                          />
                          <div v-else class="flex aspect-video w-full items-center justify-center bg-gray-100 text-sm text-gray-500 dark:bg-dark-900 dark:text-gray-400">
                            {{ t('common.loading') }}
                          </div>
                          <span
                            v-if="stepImageSrc(image)"
                            class="absolute right-2 top-2 rounded bg-gray-950/75 px-2 py-1 text-xs font-medium text-white opacity-0 transition group-hover:opacity-100 group-focus:opacity-100"
                          >
                            {{ t('helpCenter.previewImage') }}
                          </span>
                        </div>
                        <div class="flex items-center justify-between gap-3 px-3 py-2">
                          <span class="truncate text-sm font-medium text-gray-800 dark:text-gray-200">{{ image.label }}</span>
                          <Icon name="eye" size="xs" class="shrink-0 text-gray-400" />
                        </div>
                      </button>
                    </div>

                    <div v-if="step.attachments?.length" class="mt-4 rounded-lg border border-primary-100 bg-primary-50/70 p-3 dark:border-primary-900/50 dark:bg-primary-900/15">
                      <h4 class="flex items-center gap-2 text-sm font-semibold text-primary-700 dark:text-primary-300">
                        <Icon name="download" size="xs" />
                        {{ t('helpCenter.attachments') }}
                      </h4>
                      <div class="mt-3 grid grid-cols-1 gap-2 sm:grid-cols-2">
                        <button
                          v-for="(attachment, attachmentIndex) in step.attachments"
                          :key="`${selectedTutorial.id}-step-${index}-attachment-${attachmentIndex}`"
                          type="button"
                          class="flex items-center justify-between gap-3 rounded-md border border-primary-100 bg-white px-3 py-2.5 text-left text-sm shadow-sm transition hover:border-primary-300 hover:bg-primary-50 disabled:cursor-wait disabled:opacity-60 dark:border-primary-900/50 dark:bg-dark-800 dark:hover:border-primary-700 dark:hover:bg-primary-900/20"
                          :disabled="downloadingAttachment === attachment.url"
                          data-testid="help-center-step-attachment"
                          @click="downloadAttachment(attachment)"
                        >
                          <span class="min-w-0">
                            <span class="block truncate font-medium text-gray-900 dark:text-gray-100">{{ attachment.label }}</span>
                            <span v-if="attachment.file_name" class="block truncate text-xs text-gray-500 dark:text-gray-400">{{ attachment.file_name }}</span>
                          </span>
                          <span class="inline-flex shrink-0 items-center gap-1 text-primary-600 dark:text-primary-300">
                            <span v-if="downloadingAttachment === attachment.url" class="text-xs">{{ t('helpCenter.downloading') }}</span>
                            <Icon name="download" size="xs" />
                          </span>
                        </button>
                      </div>
                    </div>

                    <div v-if="step.code_blocks?.length" class="mt-4 space-y-3">
                      <div
                        v-for="(block, blockIndex) in step.code_blocks"
                        :key="`${selectedTutorial.id}-step-${index}-code-${blockIndex}`"
                        class="overflow-hidden rounded-lg border border-gray-200 bg-white dark:border-dark-700 dark:bg-dark-800"
                      >
                        <div class="flex items-center justify-between gap-3 border-b border-gray-200 bg-white px-4 py-2 dark:border-dark-700 dark:bg-dark-800">
                          <div class="min-w-0">
                            <p class="truncate text-sm font-medium text-gray-900 dark:text-white">{{ block.title }}</p>
                            <p v-if="block.language" class="text-xs text-gray-500 dark:text-gray-400">{{ block.language }}</p>
                          </div>
                          <button type="button" class="btn btn-secondary text-xs" @click="copyCode(block.content)">
                            <Icon name="copy" size="xs" />
                            {{ t('common.copy') }}
                          </button>
                        </div>
                        <pre class="max-h-[360px] overflow-auto bg-gray-950 p-4 text-sm leading-6 text-gray-100"><code>{{ block.content }}</code></pre>
                      </div>
                    </div>
                  </div>
                </li>
              </ol>
            </section>

            <section v-if="selectedTutorial.code_blocks?.length" class="mt-8">
              <h3 class="text-base font-semibold text-gray-900 dark:text-white">{{ t('helpCenter.codeBlocks') }}</h3>
              <div class="mt-4 space-y-4">
                <div
                  v-for="(block, index) in selectedTutorial.code_blocks"
                  :key="`${selectedTutorial.id}-code-${index}`"
                  class="overflow-hidden rounded-lg border border-gray-200 dark:border-dark-700"
                >
                  <div class="flex items-center justify-between gap-3 border-b border-gray-200 bg-gray-50 px-4 py-2 dark:border-dark-700 dark:bg-dark-900">
                    <div class="min-w-0">
                      <p class="truncate text-sm font-medium text-gray-900 dark:text-white">{{ block.title }}</p>
                      <p v-if="block.language" class="text-xs text-gray-500 dark:text-gray-400">{{ block.language }}</p>
                    </div>
                    <button type="button" class="btn btn-secondary text-xs" @click="copyCode(block.content)">
                      <Icon name="copy" size="xs" />
                      {{ t('common.copy') }}
                    </button>
                  </div>
                  <pre class="max-h-[460px] overflow-auto bg-gray-950 p-4 text-sm leading-6 text-gray-100"><code>{{ block.content }}</code></pre>
                </div>
              </div>
            </section>

            <section v-if="selectedTutorial.links?.length || selectedTutorial.attachments?.length" class="mt-8 grid grid-cols-1 gap-4 md:grid-cols-2">
              <div v-if="selectedTutorial.links?.length" class="rounded-lg border border-gray-200 p-4 dark:border-dark-700">
                <h3 class="text-sm font-semibold text-gray-900 dark:text-white">{{ t('helpCenter.links') }}</h3>
                <div class="mt-3 space-y-2">
                  <a
                    v-for="(link, index) in selectedTutorial.links"
                    :key="`${selectedTutorial.id}-link-${index}`"
                    :href="link.url"
                    class="flex items-center justify-between gap-3 rounded-md px-2 py-2 text-sm text-primary-700 hover:bg-primary-50 dark:text-primary-300 dark:hover:bg-primary-900/20"
                    :target="isExternalUrl(link.url) ? '_blank' : undefined"
                    :rel="isExternalUrl(link.url) ? 'noopener noreferrer' : undefined"
                  >
                    <span class="truncate">{{ link.label }}</span>
                    <Icon :name="isExternalUrl(link.url) ? 'externalLink' : 'arrowRight'" size="xs" />
                  </a>
                </div>
              </div>

              <div v-if="selectedTutorial.attachments?.length" class="rounded-lg border border-gray-200 p-4 dark:border-dark-700">
                <h3 class="text-sm font-semibold text-gray-900 dark:text-white">{{ t('helpCenter.attachments') }}</h3>
                <div class="mt-3 space-y-2">
                  <button
                    v-for="(attachment, index) in selectedTutorial.attachments"
                    :key="`${selectedTutorial.id}-attachment-${index}`"
                    type="button"
                    class="flex w-full items-center justify-between gap-3 rounded-md px-2 py-2 text-left text-sm text-primary-700 hover:bg-primary-50 disabled:cursor-wait disabled:opacity-60 dark:text-primary-300 dark:hover:bg-primary-900/20"
                    :disabled="downloadingAttachment === attachment.url"
                    @click="downloadAttachment(attachment)"
                  >
                    <span class="truncate">{{ attachment.label }}</span>
                    <span class="inline-flex items-center gap-1">
                      <span v-if="downloadingAttachment === attachment.url" class="text-xs">{{ t('helpCenter.downloading') }}</span>
                      <Icon name="download" size="xs" />
                    </span>
                  </button>
                </div>
              </div>
            </section>
          </article>

          <section v-if="faqs.length" class="rounded-lg border border-gray-200 bg-white p-6 shadow-sm dark:border-dark-700 dark:bg-dark-800">
            <div class="flex flex-col gap-4 md:flex-row md:items-start md:justify-between">
              <div>
                <div class="flex items-center gap-2">
                  <Icon name="questionCircle" size="sm" class="text-primary-600 dark:text-primary-400" />
                  <h2 class="text-xl font-semibold text-gray-900 dark:text-white">{{ t('helpCenter.faq') }}</h2>
                </div>
                <p class="mt-2 text-sm text-gray-500 dark:text-gray-400">{{ t('helpCenter.faqDescription') }}</p>
              </div>
              <label class="relative block w-full md:max-w-xs">
                <Icon name="search" size="xs" class="pointer-events-none absolute left-3 top-1/2 -translate-y-1/2 text-gray-400" />
                <input
                  v-model="faqQuery"
                  type="search"
                  class="input w-full pl-9"
                  :placeholder="t('helpCenter.faqSearchPlaceholder')"
                />
              </label>
            </div>

            <div v-if="filteredFAQs.length" class="mt-5 divide-y divide-gray-100 dark:divide-dark-700">
              <details
                v-for="(faq, index) in filteredFAQs"
                :key="faq.id"
                class="group py-4"
                :open="index === 0"
              >
                <summary class="flex cursor-pointer list-none items-center justify-between gap-4 text-left">
                  <span class="text-base font-semibold text-gray-900 dark:text-white">{{ faq.question }}</span>
                  <Icon name="chevronDown" size="sm" class="shrink-0 text-gray-400 transition group-open:rotate-180" />
                </summary>
                <div class="help-center-markdown mt-3" v-html="renderMarkdown(faq.answer_md)"></div>
                <div v-if="faq.tags?.length" class="mt-3 flex flex-wrap gap-2">
                  <span
                    v-for="tag in faq.tags"
                    :key="`${faq.id}-${tag}`"
                    class="rounded bg-gray-100 px-2 py-1 text-xs text-gray-600 dark:bg-dark-700 dark:text-gray-300"
                  >
                    {{ tag }}
                  </span>
                </div>
              </details>
            </div>
            <p v-else class="mt-5 rounded-lg border border-dashed border-gray-200 p-5 text-sm text-gray-500 dark:border-dark-700 dark:text-gray-400">
              {{ t('helpCenter.noFaqResults') }}
            </p>
            </section>
          </div>
        </div>
      </div>
    </div>

    <div
      v-if="previewImage"
      class="fixed inset-0 z-50 overflow-auto bg-gray-950/90 px-4 py-16 backdrop-blur-sm sm:px-8"
      role="dialog"
      aria-modal="true"
      :aria-label="previewImage.label || t('helpCenter.previewImage')"
      @click.self="closeImagePreview"
    >
      <div class="pointer-events-none fixed inset-x-4 top-4 z-[60] flex items-start justify-between gap-3">
        <div class="pointer-events-auto min-w-0 max-w-[calc(100vw-10rem)] rounded-md bg-gray-950/80 px-3 py-2 text-xs text-gray-200 shadow-lg backdrop-blur">
          <p class="truncate font-semibold text-gray-50">{{ previewImage.label }}</p>
          <p v-if="previewImage.file_name" class="truncate text-gray-400">{{ previewImage.file_name }}</p>
        </div>
        <div class="pointer-events-auto flex shrink-0 items-center gap-2">
          <button
            type="button"
            class="inline-flex items-center gap-1.5 rounded-md bg-gray-950/80 px-3 py-2 text-xs font-medium text-gray-100 shadow-lg transition hover:bg-gray-900 focus:outline-none focus:ring-2 focus:ring-primary-400"
            @click="previewFitMode = previewFitMode === 'fit' ? 'actual' : 'fit'"
          >
            <Icon :name="previewFitMode === 'fit' ? 'eye' : 'arrowsUpDown'" size="xs" />
            {{ previewFitMode === 'fit' ? t('helpCenter.actualSize') : t('helpCenter.fitToWindow') }}
          </button>
          <button
            type="button"
            class="inline-flex items-center gap-1.5 rounded-md bg-gray-950/80 px-3 py-2 text-xs font-medium text-gray-100 shadow-lg transition hover:bg-gray-900 focus:outline-none focus:ring-2 focus:ring-primary-400 disabled:cursor-not-allowed disabled:opacity-60"
            :disabled="downloadingAttachment === previewImage.url"
            @click="downloadAttachment(previewImage)"
          >
            <Icon name="download" size="xs" />
            {{ downloadingAttachment === previewImage.url ? t('helpCenter.downloading') : t('helpCenter.downloadImage') }}
          </button>
          <button
            type="button"
            class="inline-flex h-9 w-9 items-center justify-center rounded-md bg-gray-950/80 text-gray-100 shadow-lg transition hover:bg-gray-900 focus:outline-none focus:ring-2 focus:ring-primary-400"
            :aria-label="t('common.close')"
            data-testid="help-center-preview-close"
            @click="closeImagePreview"
          >
            <Icon name="x" size="sm" />
          </button>
        </div>
      </div>
      <div
        class="help-center-preview-frame"
        :class="previewFitMode === 'fit' ? 'help-center-preview-frame--zoom' : 'help-center-preview-frame--actual'"
        :style="previewNaturalWidth ? { '--preview-natural-width': `${previewNaturalWidth}px` } : undefined"
        data-testid="help-center-preview-frame"
        @click.self="closeImagePreview"
      >
        <img
          :src="previewImage.src"
          :alt="previewImage.label"
          class="help-center-preview-image"
          :class="previewFitMode === 'fit' ? 'help-center-preview-image--zoom' : 'help-center-preview-image--actual'"
          data-testid="help-center-preview-image"
          @load="handlePreviewImageLoad"
        />
      </div>
    </div>
  </AppLayout>
</template>

<script setup lang="ts">
import { computed, onMounted, onUnmounted, ref, watch } from 'vue'
import { RouterLink } from 'vue-router'
import { useI18n } from 'vue-i18n'
import { marked } from 'marked'
import DOMPurify from 'dompurify'
import { apiClient, helpCenterAPI } from '@/api'
import AppLayout from '@/components/layout/AppLayout.vue'
import Icon from '@/components/icons/Icon.vue'
import { useAppStore } from '@/stores/app'
import { useClipboard } from '@/composables/useClipboard'
import { getHelpCenterAttachmentAPIPath, isUploadedHelpCenterAttachment } from '@/utils/helpCenterAttachments'
import type { HelpCenterAttachment, HelpCenterConfig } from '@/types'

const { t } = useI18n()
const appStore = useAppStore()
const { copyToClipboard } = useClipboard()

const loading = ref(false)
const config = ref<HelpCenterConfig | null>(null)
const selectedTutorialId = ref('')
const faqQuery = ref('')
const downloadingAttachment = ref('')
const stepImageURLs = ref<Record<string, string>>({})
const previewImage = ref<(HelpCenterAttachment & { src: string }) | null>(null)
const previewFitMode = ref<'fit' | 'actual'>('fit')
const previewNaturalWidth = ref(0)
let imageURLGeneration = 0
let imageLoadAbortController: AbortController | null = null

marked.setOptions({
  breaks: true,
  gfm: true,
})

const tutorials = computed(() =>
  (config.value?.tutorials || [])
    .filter((tutorial) => tutorial.enabled)
    .sort((a, b) => a.sort_order - b.sort_order || a.title.localeCompare(b.title)),
)

const selectedTutorial = computed(() => {
  return tutorials.value.find((tutorial) => tutorial.id === selectedTutorialId.value) || tutorials.value[0] || null
})

const faqs = computed(() =>
  (config.value?.faqs || [])
    .filter((faq) => faq.enabled)
    .sort((a, b) => a.sort_order - b.sort_order || a.question.localeCompare(b.question)),
)

const filteredFAQs = computed(() => {
  const query = faqQuery.value.trim().toLowerCase()
  if (!query) return faqs.value
  return faqs.value.filter((faq) => {
    const text = [faq.question, faq.answer_md, ...(faq.tags || [])].join(' ').toLowerCase()
    return text.includes(query)
  })
})

const hasContent = computed(() => tutorials.value.length > 0 || faqs.value.length > 0)

watch(tutorials, (items) => {
  if (!items.length) {
    selectedTutorialId.value = ''
    return
  }
  if (!items.some((item) => item.id === selectedTutorialId.value)) {
    selectedTutorialId.value = items[0].id
  }
})

watch(selectedTutorial, (tutorial) => {
  loadStepImages(tutorial?.steps || [])
}, { immediate: true })

function renderMarkdown(content: string): string {
  const html = marked.parse(content || '') as string
  return DOMPurify.sanitize(html, {
    ADD_ATTR: ['target', 'rel'],
  })
}

function isExternalUrl(url: string): boolean {
  return /^https?:\/\//i.test(url)
}

function stepImageKey(image: HelpCenterAttachment): string {
  return image.url || `${image.file_name}-${image.label}`
}

function stepImageSrc(image: HelpCenterAttachment): string {
  if (!image.url) return ''
  if (stepImageURLs.value[stepImageKey(image)]) return stepImageURLs.value[stepImageKey(image)]
  if (!isUploadedHelpCenterAttachment(image.url)) return image.url
  return ''
}

function clearStepImageURLs(): void {
  Object.values(stepImageURLs.value).forEach((url) => {
    if (url.startsWith('blob:')) URL.revokeObjectURL(url)
  })
  stepImageURLs.value = {}
}

function setStepImageURL(key: string, url: string): void {
  const previous = stepImageURLs.value[key]
  if (previous && previous.startsWith('blob:') && previous !== url) {
    URL.revokeObjectURL(previous)
  }
  stepImageURLs.value = {
    ...stepImageURLs.value,
    [key]: url,
  }
}

async function copyCode(content: string): Promise<void> {
  await copyToClipboard(content, t('common.copiedToClipboard'))
}

function openImagePreview(image: HelpCenterAttachment): void {
  const src = stepImageSrc(image)
  if (!src) return
  previewFitMode.value = 'fit'
  previewNaturalWidth.value = 0
  previewImage.value = { ...image, src }
}

function closeImagePreview(): void {
  previewImage.value = null
  previewNaturalWidth.value = 0
}

function handlePreviewImageLoad(event: Event): void {
  const image = event.target as HTMLImageElement
  previewNaturalWidth.value = image.naturalWidth || 0
}

async function downloadAttachment(attachment: HelpCenterAttachment): Promise<void> {
  if (!attachment.url) return
  const apiPath = getHelpCenterAttachmentAPIPath(attachment.url)
  if (!apiPath) {
    window.open(attachment.url, isExternalUrl(attachment.url) ? '_blank' : '_self')
    return
  }

  downloadingAttachment.value = attachment.url
  try {
    const response = await apiClient.get<Blob>(apiPath, { responseType: 'blob' })
    const blobURL = URL.createObjectURL(response.data)
    const link = document.createElement('a')
    link.href = blobURL
    link.download = attachment.file_name || attachment.label || 'attachment'
    document.body.appendChild(link)
    link.click()
    link.remove()
    URL.revokeObjectURL(blobURL)
  } catch (error) {
    appStore.showError(t('helpCenter.downloadFailed'))
  } finally {
    downloadingAttachment.value = ''
  }
}

async function loadStepImages(steps: Array<{ images?: HelpCenterAttachment[] }>): Promise<void> {
  const generation = ++imageURLGeneration
  imageLoadAbortController?.abort()
  const abortController = new AbortController()
  imageLoadAbortController = abortController
  clearStepImageURLs()
  const uploadedImages = steps
    .flatMap((step) => step.images || [])
    .filter((image) => image.url && isUploadedHelpCenterAttachment(image.url))

  if (!uploadedImages.length) return

  const uniqueImages = Array.from(
    new Map(uploadedImages.map((image) => [stepImageKey(image), image])).values(),
  )

  await Promise.all(uniqueImages.map(async (image) => {
    const apiPath = getHelpCenterAttachmentAPIPath(image.url)
    if (!apiPath) return
    const key = stepImageKey(image)
    try {
      const response = await apiClient.get<Blob>(apiPath, {
        responseType: 'blob',
        signal: abortController.signal,
      })
      const objectURL = URL.createObjectURL(response.data)
      if (generation !== imageURLGeneration || abortController.signal.aborted) {
        URL.revokeObjectURL(objectURL)
        return
      }
      setStepImageURL(key, objectURL)
    } catch (error) {
      if (generation !== imageURLGeneration || abortController.signal.aborted) return
      setStepImageURL(key, '')
    }
  }))

  if (imageLoadAbortController === abortController) {
    imageLoadAbortController = null
  }
}

async function loadHelpCenter(): Promise<void> {
  loading.value = true
  try {
    const response = await helpCenterAPI.get()
    config.value = response.config
  } catch (error) {
    appStore.showError(t('helpCenter.failedToLoad'))
  } finally {
    loading.value = false
  }
}

onMounted(() => {
  loadHelpCenter()
})

onUnmounted(() => {
  imageURLGeneration += 1
  imageLoadAbortController?.abort()
  imageLoadAbortController = null
  clearStepImageURLs()
})
</script>

<style scoped>
.help-center-layout-container {
  container-type: inline-size;
}

.help-center-layout {
  display: grid;
  gap: 1.5rem;
  grid-template-columns: minmax(0, 1fr);
}

.help-center-sidebar {
  align-self: start;
}

.help-center-client-nav {
  width: 100%;
}

.help-center-client-list {
  display: grid;
  gap: 0.25rem;
  grid-template-columns: repeat(auto-fit, minmax(min(100%, 12rem), 1fr));
}

@container (min-width: 900px) {
  .help-center-layout--with-nav {
    align-items: stretch;
    grid-template-columns: 280px minmax(0, 1fr);
  }

  .help-center-sidebar {
    align-self: stretch;
    min-height: 100%;
  }

  .help-center-client-nav {
    position: sticky;
    top: 6rem;
  }

  .help-center-client-list {
    grid-template-columns: minmax(0, 1fr);
  }
}

.help-center-markdown {
  color: rgb(55 65 81);
  font-size: 0.9375rem;
  line-height: 1.75;
}

:global(.dark) .help-center-markdown {
  color: rgb(209 213 219);
}

.help-center-markdown :deep(h1),
.help-center-markdown :deep(h2),
.help-center-markdown :deep(h3) {
  color: rgb(17 24 39);
  font-weight: 650;
  line-height: 1.3;
  margin-top: 1.5rem;
  margin-bottom: 0.75rem;
}

:global(.dark) .help-center-markdown :deep(h1),
:global(.dark) .help-center-markdown :deep(h2),
:global(.dark) .help-center-markdown :deep(h3) {
  color: white;
}

.help-center-markdown :deep(h1) {
  font-size: 1.5rem;
}

.help-center-markdown :deep(h2) {
  font-size: 1.25rem;
}

.help-center-markdown :deep(h3) {
  font-size: 1.05rem;
}

.help-center-markdown :deep(p),
.help-center-markdown :deep(ul),
.help-center-markdown :deep(ol),
.help-center-markdown :deep(pre) {
  margin-top: 0.75rem;
}

.help-center-markdown :deep(ul),
.help-center-markdown :deep(ol) {
  padding-left: 1.25rem;
}

.help-center-markdown :deep(ul) {
  list-style: disc;
}

.help-center-markdown :deep(ol) {
  list-style: decimal;
}

.help-center-markdown :deep(code) {
  border-radius: 0.375rem;
  background: rgb(243 244 246);
  padding: 0.125rem 0.375rem;
  font-size: 0.875em;
}

:global(.dark) .help-center-markdown :deep(code) {
  background: rgb(31 41 55);
}

.help-center-markdown :deep(pre) {
  overflow: auto;
  border-radius: 0.5rem;
  background: rgb(3 7 18);
  padding: 1rem;
  color: rgb(243 244 246);
}

.help-center-markdown :deep(pre code) {
  background: transparent;
  padding: 0;
  color: inherit;
}

.help-center-markdown :deep(a) {
  color: rgb(29 78 216);
  text-decoration: underline;
  text-underline-offset: 3px;
}

:global(.dark) .help-center-markdown :deep(a) {
  color: rgb(147 197 253);
}

.help-center-preview-frame {
  display: flex;
  align-items: center;
  justify-content: center;
  min-height: calc(100vh - 8rem);
  width: 100%;
}

.help-center-preview-frame--zoom {
  min-width: min(140vw, var(--preview-natural-width, 1800px), 1800px);
}

.help-center-preview-frame--actual {
  min-width: max-content;
}

.help-center-preview-image {
  display: block;
  height: auto;
  object-fit: contain;
  box-shadow: 0 28px 80px rgb(0 0 0 / 45%);
}

.help-center-preview-image--zoom {
  max-width: none;
  width: min(140vw, var(--preview-natural-width, 1800px), 1800px);
}

.help-center-preview-image--actual {
  max-width: none;
  width: auto;
}
</style>
