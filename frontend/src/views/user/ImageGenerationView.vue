<template>
  <AppLayout>
    <div class="mx-auto min-w-0 max-w-7xl space-y-4 overflow-x-hidden sm:space-y-6">
      <section class="relative min-w-0 overflow-hidden rounded-[1.25rem] border border-gray-200/70 bg-white/75 p-4 text-gray-950 shadow-[0_24px_90px_rgba(15,23,42,0.10)] backdrop-blur-2xl dark:border-white/10 dark:bg-slate-950/70 dark:text-white dark:shadow-[0_24px_90px_rgba(0,0,0,0.32)] sm:rounded-[1.75rem] sm:p-6">
        <div class="absolute inset-x-0 top-0 h-px bg-gradient-to-r from-transparent via-cyan-300/70 to-transparent dark:via-cyan-200/40"></div>
        <div class="relative min-w-0">
          <div class="min-w-0 max-w-3xl">
            <div class="mb-3 inline-flex items-center gap-2 rounded-full border border-cyan-200/70 bg-cyan-50/80 px-3 py-1 text-xs font-semibold text-cyan-700 dark:border-cyan-300/20 dark:bg-cyan-300/10 dark:text-cyan-200">
              <Icon name="sparkles" size="xs" />
              <span>{{ t('imageGeneration.eyebrow') }}</span>
            </div>
            <h2 class="text-2xl font-semibold sm:text-4xl">
              {{ t('imageGeneration.title') }}
            </h2>
            <p class="mt-3 max-w-2xl break-words text-sm leading-6 text-gray-600 dark:text-slate-400">
              {{ t('imageGeneration.description') }}
            </p>
          </div>
        </div>
      </section>

      <div class="grid min-w-0 gap-4 sm:gap-6 xl:grid-cols-[minmax(0,0.95fr)_minmax(0,1.05fr)]">
        <section class="min-w-0 space-y-4">
          <div class="min-w-0 overflow-hidden rounded-[1.25rem] border border-gray-200/70 bg-white/80 p-4 shadow-sm backdrop-blur-xl dark:border-white/10 dark:bg-white/[0.045] sm:rounded-[1.5rem] sm:p-6">
            <div class="mb-5 flex min-w-0 items-start justify-between gap-3 sm:gap-4">
              <div class="min-w-0">
                <h3 class="text-lg font-semibold text-gray-950 dark:text-white">{{ t('imageGeneration.compose.title') }}</h3>
                <p class="mt-1 break-words text-sm text-gray-500 dark:text-slate-400">{{ t('imageGeneration.compose.subtitle') }}</p>
              </div>
              <button type="button" class="btn btn-ghost btn-icon shrink-0" :title="t('common.refresh')" @click="loadKeys">
                <Icon name="refresh" size="sm" :class="{ 'animate-spin': loadingKeys }" />
              </button>
            </div>

            <div class="space-y-5">
              <div>
                <label class="input-label">{{ t('imageGeneration.form.apiKey') }}</label>
                <div v-if="imageKeys.length > 0" class="flex min-w-0 flex-col gap-2 sm:flex-row sm:items-center">
                  <select v-model.number="selectedKeyId" class="input min-w-0 flex-1 truncate">
                    <option v-for="key in imageKeys" :key="key.id" :value="key.id">
                      {{ formatApiKeyOption(key) }}
                    </option>
                  </select>
                  <router-link to="/keys" class="btn btn-secondary w-full shrink-0 justify-center sm:w-auto">
                    <Icon name="key" size="xs" />
                    {{ t('imageGeneration.emptyKeys.action') }}
                  </router-link>
                </div>
                <div v-if="imageKeys.length > 0 && selectedKey" class="mt-2 flex min-w-0 flex-wrap items-center gap-2 text-xs text-gray-500 dark:text-slate-400">
                  <span class="max-w-full break-words">{{ selectedKey.group?.name || t('imageGeneration.form.noGroup') }}</span>
                  <span class="text-gray-300 dark:text-slate-600">/</span>
                  <span class="max-w-full break-all">{{ maskApiKey(selectedKey.key) }}</span>
                  <span class="rounded-full bg-emerald-100 px-2 py-0.5 font-semibold text-emerald-700 dark:bg-emerald-400/10 dark:text-emerald-300">OpenAI</span>
                </div>
                <div v-if="imageKeys.length === 0" class="min-w-0 rounded-2xl border border-dashed border-gray-300 bg-gray-50/80 p-5 text-sm text-gray-600 dark:border-white/10 dark:bg-white/[0.035] dark:text-slate-300">
                  <p class="font-medium text-gray-900 dark:text-white">{{ t('imageGeneration.emptyKeys.title') }}</p>
                  <p class="mt-1 leading-6">{{ loadingKeys ? t('common.loading') : t('imageGeneration.emptyKeys.description') }}</p>
                  <router-link to="/keys" class="btn btn-secondary btn-sm mt-4">
                    <Icon name="key" size="xs" />
                    {{ t('imageGeneration.emptyKeys.action') }}
                  </router-link>
                </div>
              </div>

              <div>
                <label class="input-label">{{ t('imageGeneration.form.model') }}</label>
                <select v-model="model" class="input min-w-0" :disabled="modelOptions.length === 0">
                  <option v-if="modelOptions.length === 0" value="">
                    {{ t('imageGeneration.form.noSupportedModels') }}
                  </option>
                  <option v-for="option in modelOptions" :key="option" :value="option">
                    {{ option }}
                  </option>
                </select>
              </div>

              <TextArea
                v-model="prompt"
                :label="t('imageGeneration.form.prompt')"
                :placeholder="t('imageGeneration.form.promptPlaceholder')"
                :hint="t('imageGeneration.form.promptHint')"
                rows="8"
              />

              <div class="grid min-w-0 gap-4 sm:grid-cols-2">
                <div class="min-w-0">
                  <div class="mb-1.5 flex items-center gap-1.5">
                    <label class="input-label !mb-0">{{ t('imageGeneration.form.size') }}</label>
                    <HelpTooltip trigger="both" width-class="w-80 max-w-[calc(100vw-2rem)]">
                      <template #trigger>
                        <button
                          type="button"
                          class="inline-flex h-5 w-5 items-center justify-center rounded-full text-gray-400 transition hover:bg-gray-100 hover:text-primary-600 focus:outline-none focus:ring-2 focus:ring-primary-500/30 dark:text-slate-500 dark:hover:bg-white/10 dark:hover:text-primary-300"
                          :aria-label="t('imageGeneration.sizeRules.title')"
                          :title="t('imageGeneration.sizeRules.title')"
                        >
                          <Icon name="infoCircle" size="sm" />
                        </button>
                      </template>
                      <div class="space-y-2 pr-3">
                        <p class="text-sm font-semibold">{{ t('imageGeneration.sizeRules.title') }}</p>
                        <p>{{ t('imageGeneration.sizeRules.summary') }}</p>
                        <ul class="list-disc space-y-1 pl-4">
                          <li>{{ t('imageGeneration.sizeRules.maxSide') }}</li>
                          <li>{{ t('imageGeneration.sizeRules.multiple') }}</li>
                          <li>{{ t('imageGeneration.sizeRules.ratio') }}</li>
                          <li>{{ t('imageGeneration.sizeRules.pixels') }}</li>
                        </ul>
                        <p class="text-gray-300">{{ t('imageGeneration.sizeRules.presets') }}</p>
                      </div>
                    </HelpTooltip>
                  </div>
                  <select :value="sizeSelectValue" class="input min-w-0 truncate" @change="handleSizeSelect">
                    <option v-for="option in sizeOptions" :key="option.value" :value="option.value">
                      {{ option.label }}
                    </option>
                  </select>
                  <p class="input-hint">
                    {{ t('imageGeneration.customSize.current', { size: displaySizeWithTier(size) }) }}
                  </p>
                  <div v-if="showCustomSizeEditor" class="mt-3 min-w-0 rounded-2xl border border-cyan-100 bg-cyan-50/60 p-3 dark:border-cyan-300/20 dark:bg-cyan-300/10">
                    <div class="grid min-w-0 grid-cols-[minmax(0,1fr)_1rem_minmax(0,1fr)] items-end gap-2">
                      <div class="min-w-0">
                        <label class="input-label text-xs">{{ t('imageGeneration.customSize.width') }}</label>
                        <input
                          v-model.trim="customWidth"
                          type="number"
                          min="1"
                          step="16"
                          inputmode="numeric"
                          class="input min-w-0"
                          :class="{ 'input-error': customSizeError }"
                          placeholder="1024"
                        />
                      </div>
                      <span class="pb-3 text-sm font-semibold text-gray-400 dark:text-slate-500">x</span>
                      <div class="min-w-0">
                        <label class="input-label text-xs">{{ t('imageGeneration.customSize.height') }}</label>
                        <input
                          v-model.trim="customHeight"
                          type="number"
                          min="1"
                          step="16"
                          inputmode="numeric"
                          class="input min-w-0"
                          :class="{ 'input-error': customSizeError }"
                          placeholder="1024"
                        />
                      </div>
                    </div>
                    <p class="mt-2 text-xs leading-5 text-gray-500 dark:text-slate-400">{{ t('imageGeneration.customSize.hint') }}</p>
                    <p v-if="customSizeError" class="input-error-text">{{ customSizeError }}</p>
                    <div class="mt-3 flex min-w-0 flex-col gap-2 sm:flex-row">
                      <button type="button" class="btn btn-primary btn-sm flex-1" @click="saveCustomSize">
                        <Icon name="check" size="xs" />
                        {{ t('common.save') }}
                      </button>
                      <button type="button" class="btn btn-secondary btn-sm w-full sm:w-auto" @click="cancelCustomSize">
                        {{ t('common.cancel') }}
                      </button>
                    </div>
                  </div>
                </div>
                <div class="min-w-0">
                  <label class="input-label">{{ t('imageGeneration.form.quality') }}</label>
                  <select v-model="quality" class="input min-w-0">
                    <option v-for="option in qualityOptions" :key="option.value" :value="option.value">
                      {{ option.label }}
                    </option>
                  </select>
                </div>
              </div>

              <div>
                <label class="input-label">{{ t('imageGeneration.form.count') }}</label>
                <div class="grid grid-cols-2 gap-2 sm:grid-cols-4">
                  <button
                    v-for="count in countOptions"
                    :key="count"
                    type="button"
                    class="min-w-0 rounded-xl border px-3 py-2 text-sm font-semibold transition"
                    :class="n === count ? 'border-gray-950 bg-gray-950 text-white dark:border-white dark:bg-white dark:text-slate-950' : 'border-gray-200 bg-white text-gray-600 hover:border-cyan-300 hover:text-gray-950 dark:border-white/10 dark:bg-white/[0.035] dark:text-slate-300 dark:hover:border-cyan-300/30 dark:hover:text-white'"
                    @click="n = count"
                  >
                    {{ count }}
                  </button>
                </div>
              </div>

              <div class="min-w-0 rounded-2xl border border-cyan-100 bg-cyan-50/60 p-4 dark:border-cyan-300/20 dark:bg-cyan-300/10">
                <div class="flex min-w-0 flex-wrap items-start justify-between gap-3">
                  <div class="min-w-0">
                    <p class="text-sm font-semibold text-gray-950 dark:text-white">{{ t('imageGeneration.pricing.title') }}</p>
                    <p class="mt-1 break-words text-xs leading-5 text-gray-600 dark:text-slate-300">{{ pricePreview.description }}</p>
                  </div>
                  <span class="shrink-0 rounded-full bg-white px-2.5 py-1 text-xs font-semibold text-cyan-700 shadow-sm dark:bg-white/10 dark:text-cyan-200">
                    {{ pricePreview.tierLabel }}
                  </span>
                </div>
                <div v-if="pricePreview.rows.length > 0" class="mt-3 grid min-w-0 gap-2 sm:grid-cols-3">
                  <div v-for="row in pricePreview.rows" :key="row.label" class="min-w-0 rounded-xl bg-white/75 px-3 py-2 dark:bg-white/[0.06]">
                    <p class="text-[11px] font-medium uppercase text-gray-400 dark:text-slate-500">{{ row.label }}</p>
                    <p class="mt-1 break-words text-sm font-semibold text-gray-950 dark:text-white">{{ row.value }}</p>
                  </div>
                </div>
              </div>

              <div>
                <label class="input-label">{{ t('imageGeneration.form.templates') }}</label>
                <div class="grid min-w-0 gap-2 sm:grid-cols-2">
                  <button
                    v-for="template in templates"
                    :key="template.key"
                    type="button"
                    class="min-w-0 rounded-2xl border border-gray-200 bg-white/65 px-4 py-3 text-left transition hover:border-cyan-300 hover:bg-cyan-50/60 dark:border-white/10 dark:bg-white/[0.035] dark:hover:border-cyan-300/30 dark:hover:bg-white/[0.06]"
                    @click="applyTemplate(template.prompt)"
                  >
                    <span class="block text-sm font-semibold text-gray-900 dark:text-white">{{ template.title }}</span>
                    <span class="mt-1 block text-xs leading-5 text-gray-500 dark:text-slate-400">{{ template.description }}</span>
                  </button>
                </div>
              </div>

              <div class="flex min-w-0 flex-col gap-3 border-t border-gray-100 pt-5 dark:border-white/10 sm:flex-row">
                <button type="button" class="btn btn-primary btn-lg flex-1" :disabled="!canGenerate" @click="startGenerate">
                  <span v-if="generating" class="spinner h-4 w-4"></span>
                  <Icon v-else name="sparkles" size="sm" />
                  {{ generating ? t('imageGeneration.actions.generating') : t('imageGeneration.actions.generate') }}
                </button>
                <button type="button" class="btn btn-secondary btn-lg w-full sm:w-auto" :disabled="generating" @click="resetForm">
                  <Icon name="refresh" size="sm" />
                  {{ t('common.reset') }}
                </button>
              </div>
            </div>
          </div>
        </section>

        <section class="min-w-0 space-y-4">
          <div class="min-w-0 overflow-hidden rounded-[1.25rem] border border-gray-200/70 bg-white/80 p-4 shadow-sm backdrop-blur-xl dark:border-white/10 dark:bg-white/[0.045] sm:rounded-[1.5rem] sm:p-6">
            <div class="mb-5 flex min-w-0 flex-col gap-3 sm:flex-row sm:items-start sm:justify-between">
              <div class="min-w-0">
                <h3 class="text-lg font-semibold text-gray-950 dark:text-white">{{ t('imageGeneration.preview.title') }}</h3>
                <p class="mt-1 break-words text-sm text-gray-500 dark:text-slate-400">{{ statusText }}</p>
                <button
                  v-if="activeClientRequestId && !generating"
                  type="button"
                  class="mt-2 block max-w-full overflow-hidden truncate rounded-full bg-cyan-50 px-2.5 py-1 text-left font-mono text-[11px] font-semibold text-cyan-700 transition hover:bg-cyan-100 dark:bg-cyan-300/10 dark:text-cyan-200 dark:hover:bg-cyan-300/15"
                  :title="activeClientRequestId"
                  @click="copyClientRequestId"
                >
                  {{ t('imageGeneration.preview.requestId') }}: {{ activeClientRequestId }}
                </button>
              </div>
              <button
                type="button"
                class="btn btn-secondary btn-sm w-full sm:w-auto sm:shrink-0"
                :disabled="generatedImages.length === 0"
                @click="downloadAll"
              >
                <Icon name="download" size="xs" />
                {{ t('imageGeneration.actions.downloadAll') }}
              </button>
            </div>

            <div
              v-if="generating"
              class="relative flex min-h-[24rem] min-w-0 flex-col items-center justify-center overflow-hidden rounded-[1.25rem] border border-cyan-200/80 bg-cyan-50/70 p-4 text-center dark:border-cyan-300/20 dark:bg-cyan-300/10 sm:min-h-[32rem] sm:p-6"
            >
              <div class="absolute inset-x-8 top-8 h-24 rounded-full bg-cyan-200/35 blur-3xl dark:bg-cyan-300/10"></div>
              <div class="relative mb-5 flex h-20 w-20 items-center justify-center rounded-3xl bg-white text-cyan-700 shadow-sm ring-1 ring-cyan-100 dark:bg-slate-950/80 dark:text-cyan-200 dark:ring-cyan-300/20">
                <span class="spinner h-8 w-8"></span>
              </div>
              <h4 class="relative text-base font-semibold text-gray-950 dark:text-white">{{ t('imageGeneration.preview.runningTitle') }}</h4>
              <p class="relative mt-2 max-w-md text-sm leading-6 text-gray-600 dark:text-slate-300">
                {{ t('imageGeneration.preview.runningDescription') }}
              </p>
              <div class="relative mt-5 w-full min-w-0 max-w-md">
                <div class="progress h-2.5 bg-white/80 dark:bg-slate-900/80">
                  <div class="progress-bar h-full rounded-full" :style="{ width: `${generationProgress}%` }"></div>
                </div>
                <div class="mt-3 grid min-w-0 gap-2 text-left sm:grid-cols-2">
                  <div class="min-w-0 rounded-xl bg-white/80 px-3 py-2 text-xs shadow-sm dark:bg-slate-950/55">
                    <p class="font-medium text-gray-400 dark:text-slate-500">{{ t('imageGeneration.preview.elapsed') }}</p>
                    <p class="mt-1 font-semibold text-gray-950 dark:text-white">{{ t('imageGeneration.preview.elapsedSeconds', { seconds: generationElapsedSeconds }) }}</p>
                  </div>
                  <div class="min-w-0 rounded-xl bg-white/80 px-3 py-2 text-xs shadow-sm dark:bg-slate-950/55">
                    <p class="font-medium text-gray-400 dark:text-slate-500">{{ t('imageGeneration.preview.requestId') }}</p>
                    <button
                      type="button"
                      class="mt-1 max-w-full break-all text-left font-mono font-semibold text-cyan-700 hover:text-cyan-900 dark:text-cyan-200 dark:hover:text-cyan-100"
                      :title="activeClientRequestId || ''"
                      @click="copyClientRequestId"
                    >
                      {{ activeClientRequestId || '-' }}
                    </button>
                  </div>
                </div>
                <p class="mt-3 text-xs leading-5 text-gray-500 dark:text-slate-400">
                  {{ t('imageGeneration.preview.logHint') }}
                </p>
              </div>
            </div>

            <div
              v-else-if="generatedImages.length === 0"
              class="flex min-h-[24rem] min-w-0 flex-col items-center justify-center rounded-[1.25rem] border border-dashed border-gray-300 bg-gray-50/80 p-4 text-center dark:border-white/10 dark:bg-white/[0.025] sm:min-h-[32rem] sm:p-6"
            >
              <div class="mb-4 flex h-16 w-16 items-center justify-center rounded-2xl bg-cyan-100 text-cyan-700 dark:bg-cyan-300/10 dark:text-cyan-200">
                <Icon name="sparkles" size="xl" />
              </div>
              <h4 class="text-base font-semibold text-gray-950 dark:text-white">{{ t('imageGeneration.preview.emptyTitle') }}</h4>
              <p class="mt-2 max-w-md text-sm leading-6 text-gray-500 dark:text-slate-400">{{ t('imageGeneration.preview.emptyDescription') }}</p>
            </div>

            <div v-else class="grid min-w-0 gap-4 sm:grid-cols-2">
              <figure
                v-for="(image, index) in generatedImages"
                :key="`${image.url}-${index}`"
                class="group min-w-0 overflow-hidden rounded-[1.25rem] border border-gray-200 bg-white shadow-sm transition hover:border-cyan-300 dark:border-white/10 dark:bg-slate-950/60 dark:hover:border-cyan-300/30"
              >
                <button type="button" class="block w-full max-w-full overflow-hidden bg-gray-100 dark:bg-slate-900" @click="previewImage = image.url">
                  <img :src="image.url" :alt="t('imageGeneration.preview.imageAlt', { number: index + 1 })" class="block aspect-square h-auto w-full max-w-full object-contain" />
                </button>
                <figcaption class="flex min-w-0 items-center justify-between gap-3 p-3">
                  <span class="min-w-0 truncate text-xs font-medium text-gray-500 dark:text-slate-400">
                    {{ t('imageGeneration.preview.imageNumber', { number: index + 1 }) }}
                  </span>
                  <div class="flex shrink-0 items-center gap-1">
                    <button type="button" class="btn btn-ghost btn-icon !p-2" :title="t('common.copy')" @click="copyImage(image.url)">
                      <Icon name="copy" size="xs" />
                    </button>
                    <button type="button" class="btn btn-ghost btn-icon !p-2" :title="t('imageGeneration.actions.download')" @click="downloadImage(image.url, index)">
                      <Icon name="download" size="xs" />
                    </button>
                  </div>
                </figcaption>
              </figure>
            </div>

            <div v-if="errorMessage" class="mt-4 rounded-2xl border border-red-200 bg-red-50 px-4 py-3 text-sm text-red-700 dark:border-red-400/20 dark:bg-red-400/10 dark:text-red-200">
              {{ errorMessage }}
            </div>
          </div>

          <div class="min-w-0 overflow-hidden rounded-[1.25rem] border border-gray-200/70 bg-white/80 p-4 shadow-sm backdrop-blur-xl dark:border-white/10 dark:bg-white/[0.045] sm:rounded-[1.5rem] sm:p-6">
            <div class="mb-4 flex min-w-0 items-center justify-between gap-3">
              <div class="min-w-0">
                <h3 class="text-lg font-semibold text-gray-950 dark:text-white">{{ t('imageGeneration.history.title') }}</h3>
                <p class="mt-1 break-words text-sm text-gray-500 dark:text-slate-400">{{ t('imageGeneration.history.subtitle') }}</p>
              </div>
              <button type="button" class="btn btn-ghost btn-sm shrink-0" :disabled="history.length === 0" @click="clearHistory">
                {{ t('imageGeneration.history.clear') }}
              </button>
            </div>

            <div v-if="history.length === 0" class="rounded-2xl border border-dashed border-gray-300 bg-gray-50/70 p-5 text-sm text-gray-500 dark:border-white/10 dark:bg-white/[0.025] dark:text-slate-400">
              {{ t('imageGeneration.history.empty') }}
            </div>
            <div v-else class="space-y-3">
              <button
                v-for="item in history"
                :key="item.id"
                type="button"
                class="flex w-full items-center gap-3 rounded-2xl border border-gray-200 bg-white/65 p-3 text-left transition hover:border-cyan-300 hover:bg-cyan-50/50 dark:border-white/10 dark:bg-white/[0.035] dark:hover:border-cyan-300/30 dark:hover:bg-white/[0.06]"
                @click="restoreHistory(item)"
              >
                <img :src="item.thumbnail" :alt="item.prompt" class="h-14 w-14 shrink-0 rounded-xl object-cover" />
                <span class="min-w-0">
                  <span class="block truncate text-sm font-semibold text-gray-900 dark:text-white">{{ item.prompt }}</span>
                  <span class="mt-1 block truncate text-xs text-gray-500 dark:text-slate-400">
                    {{ item.model }} / {{ item.size }} / {{ item.createdAt }}
                  </span>
                </span>
              </button>
            </div>
          </div>
        </section>
      </div>
    </div>

    <div v-if="previewImage" class="fixed inset-0 z-50 flex items-center justify-center bg-black/80 p-4 backdrop-blur-sm" @click="previewImage = ''">
      <button type="button" class="absolute right-4 top-4 rounded-xl bg-white/10 p-2 text-white transition hover:bg-white/20" :title="t('common.close')" @click.stop="previewImage = ''">
        <Icon name="x" size="md" />
      </button>
      <img :src="previewImage" :alt="t('imageGeneration.preview.fullscreenAlt')" class="max-h-[92vh] max-w-[92vw] rounded-2xl object-contain shadow-2xl" @click.stop />
    </div>
  </AppLayout>
</template>

<script setup lang="ts">
import { computed, onBeforeUnmount, onMounted, ref, watch } from 'vue'
import { useI18n } from 'vue-i18n'
import AppLayout from '@/components/layout/AppLayout.vue'
import Icon from '@/components/icons/Icon.vue'
import HelpTooltip from '@/components/common/HelpTooltip.vue'
import TextArea from '@/components/common/TextArea.vue'
import { keysAPI, imageGenerationAPI } from '@/api'
import { useAppStore } from '@/stores/app'
import { maskApiKey } from '@/utils/maskApiKey'
import type { ApiKey } from '@/types'
import type { GeneratedImage, ImageGenerationQuality, ImageGenerationSize } from '@/api/imageGeneration'

interface HistoryItem {
  id: string
  prompt: string
  model: string
  size: ImageGenerationSize
  quality: ImageGenerationQuality
  thumbnail: string
  images: string[]
  createdAt: string
}

interface GenerateRequestedImagesParams {
  apiKey: string
  model: string
  prompt: string
  size: ImageGenerationSize
  quality: ImageGenerationQuality
  requestedCount: number
  clientRequestId: string
  signal: AbortSignal
}

const HISTORY_KEY = 'sub2api_image_generation_history'
const HISTORY_DB_NAME = 'sub2api_image_generation'
const HISTORY_DB_VERSION = 1
const HISTORY_STORE_NAME = 'history'
const HISTORY_RECORD_KEY = 'items'
const HISTORY_LIMIT = 8
const HISTORY_IMAGE_MAX_SIDE = 1024
const HISTORY_IMAGE_QUALITY = 0.82
const HISTORY_IMAGE_OPTIMIZE_THRESHOLD = 450_000
const CUSTOM_SIZES_KEY = 'sub2api_image_generation_custom_sizes'
const CUSTOM_SIZE_SELECT_VALUE = '__custom__'
const MIN_IMAGE_PIXELS = 655360
const MAX_IMAGE_PIXELS = 8294400
const MAX_IMAGE_SIDE = 3840
const IMAGE_SIZE_STEP = 16
const MAX_IMAGE_RATIO = 3

type ImageSizeTier = '1K' | '2K' | '4K'
const imageGenerationQualities: ImageGenerationQuality[] = ['low', 'medium', 'high', 'auto']

interface SizeOption {
  value: ImageGenerationSize
  label: string
  custom?: boolean
}

interface ParsedImageSize {
  width: number
  height: number
  value: ImageGenerationSize
  pixels: number
  tier: ImageSizeTier
}

const { t, locale } = useI18n()
const appStore = useAppStore()

const keys = ref<ApiKey[]>([])
const loadingKeys = ref(false)
const selectedKeyId = ref<number | null>(null)
const model = ref('gpt-image-2')
const prompt = ref('')
const size = ref<ImageGenerationSize>('1024x1024')
const quality = ref<ImageGenerationQuality>('high')
const n = ref(1)
const generating = ref(false)
const generatedImages = ref<GeneratedImage[]>([])
const errorMessage = ref('')
const previewImage = ref('')
const history = ref<HistoryItem[]>([])
const customSizes = ref<ImageGenerationSize[]>([])
const showCustomSizeEditor = ref(false)
const customWidth = ref('1024')
const customHeight = ref('1024')
const customSizeError = ref('')
const activeClientRequestId = ref('')
const generationStartedAt = ref<number | null>(null)
const generationNow = ref(Date.now())
const keyModelIds = ref<string[]>([])
const keyModelsLoaded = ref(false)
let abortController: AbortController | null = null
let keyModelsAbortController: AbortController | null = null
let generationTimer: number | null = null
let historyRevision = 0
let historySaveQueue: Promise<void> = Promise.resolve()

const countOptions = [1, 2, 3, 4]
const allImageModelOptions = ['gpt-image-2', 'gpt-image-1.5', 'gpt-image-1']
const imageModelScopeAliases: Record<string, string[]> = {
  'gpt-image-2': ['gpt-image-2', 'gpt_image_2', 'gptimage2', 'image2', 'image_2', 'image-2', 'openai_image_2', 'openai-image-2'],
  'gpt-image-1.5': ['gpt-image-1.5', 'gpt_image_1_5', 'gptimage15', 'image1.5', 'image_1_5', 'image-1.5', 'openai_image_1_5', 'openai-image-1.5'],
  'gpt-image-1': ['gpt-image-1', 'gpt_image_1', 'gptimage1', 'image1', 'image_1', 'image-1', 'openai_image_1', 'openai-image-1']
}
const allImageModelScopeAliases = ['gpt-image', 'gpt_image', 'gptimage', 'openai_image', 'openai-image', 'openaiimage', 'image_generation', 'image-generation', 'imagegeneration', 'image']
const presetSizeOptions = computed(() => [
  { value: '1024x1024', label: formatSizeOptionLabel('1024x1024', t('imageGeneration.sizes.square1k')) },
  { value: '1536x1024', label: formatSizeOptionLabel('1536x1024', t('imageGeneration.sizes.landscape1_5k')) },
  { value: '1024x1536', label: formatSizeOptionLabel('1024x1536', t('imageGeneration.sizes.portrait1_5k')) },
  { value: '2048x2048', label: formatSizeOptionLabel('2048x2048', t('imageGeneration.sizes.square2k')) },
  { value: '2048x1152', label: formatSizeOptionLabel('2048x1152', t('imageGeneration.sizes.landscape2k')) },
  { value: '2880x2880', label: formatSizeOptionLabel('2880x2880', t('imageGeneration.sizes.square2880')) },
  { value: '3840x2160', label: formatSizeOptionLabel('3840x2160', t('imageGeneration.sizes.landscape4k')) },
  { value: '2160x3840', label: formatSizeOptionLabel('2160x3840', t('imageGeneration.sizes.portrait4k')) }
] as SizeOption[])

const imageKeys = computed(() =>
  keys.value.filter((key) => {
    return key.status === 'active' &&
      key.group?.platform === 'openai' &&
      key.group?.allow_image_generation === true
  })
)

const selectedKey = computed(() => imageKeys.value.find((key) => key.id === selectedKeyId.value) || null)

const modelOptions = computed(() => {
  if (!selectedKey.value) return []

  const scopedOptions = imageModelsFromScopes(selectedKey.value.group?.supported_model_scopes)
  const namedOptions = scopedOptions.explicit
    ? []
    : imageModelsFromText(`${selectedKey.value.group?.name || ''} ${selectedKey.value.group?.description || ''}`)
  const configuredOptions = scopedOptions.explicit
    ? scopedOptions.models
    : namedOptions.length > 0
      ? namedOptions
      : [...allImageModelOptions]
  const availableOptions = imageModelsFromAvailableIds(keyModelIds.value)

  if (keyModelsLoaded.value && keyModelIds.value.length > 0 && availableOptions.length === 0) {
    return []
  }

  if (availableOptions.length > 0 && availableOptions.length < allImageModelOptions.length) {
    const intersection = configuredOptions.filter((option) => availableOptions.includes(option))
    return intersection.length > 0 ? intersection : availableOptions
  }

  return configuredOptions
})

const canGenerate = computed(() => {
  return !generating.value &&
    !!selectedKey.value &&
    model.value.trim() !== '' &&
    modelOptions.value.includes(model.value) &&
    prompt.value.trim().length >= 3 &&
    validateImageSize(size.value).valid
})

const sizeOptions = computed(() => [
  ...presetSizeOptions.value,
  ...customSizes.value.map((value) => ({
    value,
    label: formatSizeOptionLabel(value, t('imageGeneration.customSize.optionLabel', { size: displaySize(value) })),
    custom: true
  })),
  { value: CUSTOM_SIZE_SELECT_VALUE, label: t('imageGeneration.customSize.add') }
] as SizeOption[])

const sizeSelectValue = computed(() => {
  return sizeOptions.value.some((option) => option.value === size.value)
    ? size.value
    : CUSTOM_SIZE_SELECT_VALUE
})

const qualityOptions = computed(() => [
  { value: 'high', label: t('imageGeneration.qualities.high') },
  { value: 'medium', label: t('imageGeneration.qualities.medium') },
  { value: 'low', label: t('imageGeneration.qualities.low') },
  { value: 'auto', label: t('imageGeneration.qualities.auto') }
] as Array<{ value: ImageGenerationQuality; label: string }>)

const templates = computed(() => [
  {
    key: 'product',
    title: t('imageGeneration.templates.product.title'),
    description: t('imageGeneration.templates.product.description'),
    prompt: t('imageGeneration.templates.product.prompt')
  },
  {
    key: 'poster',
    title: t('imageGeneration.templates.poster.title'),
    description: t('imageGeneration.templates.poster.description'),
    prompt: t('imageGeneration.templates.poster.prompt')
  },
  {
    key: 'icon',
    title: t('imageGeneration.templates.icon.title'),
    description: t('imageGeneration.templates.icon.description'),
    prompt: t('imageGeneration.templates.icon.prompt')
  },
  {
    key: 'infographic',
    title: t('imageGeneration.templates.infographic.title'),
    description: t('imageGeneration.templates.infographic.description'),
    prompt: t('imageGeneration.templates.infographic.prompt')
  }
])

const statusText = computed(() => {
  if (generating.value) return t('imageGeneration.preview.generating', { seconds: generationElapsedSeconds.value })
  if (generatedImages.value.length > 0) {
    const count = generatedImages.value.length
    return count === 1
      ? t('imageGeneration.preview.readyOne')
      : t('imageGeneration.preview.readyMany', { count })
  }
  return t('imageGeneration.preview.idle')
})

const selectedSizeInfo = computed(() => parseImageSize(size.value))

const generationElapsedSeconds = computed(() => {
  if (!generationStartedAt.value) return 0
  return Math.max(0, Math.floor((generationNow.value - generationStartedAt.value) / 1000))
})

const generationProgress = computed(() => {
  if (!generating.value) return 0
  const seconds = generationElapsedSeconds.value
  return Math.min(92, Math.max(8, 8 + seconds * 3.5))
})

const pricePreview = computed(() => {
  const group = selectedKey.value?.group
  const sizeInfo = selectedSizeInfo.value
  const tier = sizeInfo?.tier || '2K'
  const tierLabel = t('imageGeneration.pricing.tier', { tier })
  if (!group) {
    return {
      tierLabel,
      description: t('imageGeneration.pricing.noKey'),
      rows: [] as Array<{ label: string; value: string }>
    }
  }

  const configuredPrice = imagePriceForTier(group, tier)
  const multiplier = effectiveImageMultiplier(group)
  const rows = [
    { label: t('imageGeneration.pricing.multiplier'), value: formatMultiplier(multiplier) },
    { label: t('imageGeneration.pricing.unitPrice'), value: configuredPrice == null ? t('imageGeneration.pricing.serverDefault') : formatUSD(configuredPrice * multiplier) },
    { label: t('imageGeneration.pricing.estimatedTotal'), value: configuredPrice == null ? t('imageGeneration.pricing.serverCalculated') : formatUSD(configuredPrice * multiplier * n.value) }
  ]

  return {
    tierLabel,
    description: configuredPrice == null
      ? t('imageGeneration.pricing.defaultDescription')
      : t('imageGeneration.pricing.configuredDescription', { count: n.value }),
    rows
  }
})

function formatApiKeyOption(key: ApiKey): string {
  const groupName = key.group?.name || t('imageGeneration.form.noGroup')
  return `${key.name} / ${groupName} / ${maskApiKey(key.key)}`
}

function normalizeModelScope(value: string): string {
  return value.trim().toLowerCase().replace(/[^a-z0-9]/g, '')
}

function imageModelsFromScopes(scopes: string[] | undefined): { explicit: boolean; models: string[] } {
  if (!Array.isArray(scopes) || scopes.length === 0) {
    return { explicit: false, models: [...allImageModelOptions] }
  }

  const normalizedScopes = new Set(scopes.map(normalizeModelScope).filter(Boolean))
  if (normalizedScopes.size === 0) {
    return { explicit: false, models: [...allImageModelOptions] }
  }
  if (allImageModelScopeAliases.some((scope) => normalizedScopes.has(normalizeModelScope(scope)))) {
    return { explicit: true, models: [...allImageModelOptions] }
  }

  return {
    explicit: true,
    models: allImageModelOptions.filter((option) => {
      const aliases = [option, ...(imageModelScopeAliases[option] || [])]
      return aliases.some((scope) => normalizedScopes.has(normalizeModelScope(scope)))
    })
  }
}

function imageModelsFromAvailableIds(ids: string[]): string[] {
  if (!Array.isArray(ids) || ids.length === 0) return []
  const normalizedIds = new Set(ids.map((id) => id.trim().toLowerCase()).filter(Boolean))
  return allImageModelOptions.filter((option) => normalizedIds.has(option))
}

function imageModelsFromText(value: string): string[] {
  const text = value.trim().toLowerCase()
  if (!text) return []
  return allImageModelOptions.filter((option) => {
    if (option === 'gpt-image-2') {
      return /gpt[-_\s]?image[-_\s]?2|openai[-_\s]?image[-_\s]?2|image[-_\s]?2/.test(text)
    }
    if (option === 'gpt-image-1.5') {
      return /gpt[-_\s]?image[-_\s]?1(?:\.|[-_\s])?5|openai[-_\s]?image[-_\s]?1(?:\.|[-_\s])?5|image[-_\s]?1(?:\.|[-_\s])?5/.test(text)
    }
    return /gpt[-_\s]?image[-_\s]?1(?![\d.])|openai[-_\s]?image[-_\s]?1(?![\d.])|image[-_\s]?1(?![\d.])/.test(text)
  })
}

function displaySize(value: string): string {
  const parsed = parseImageSize(value)
  if (!parsed) return value
  return `${parsed.width} x ${parsed.height}`
}

function displaySizeWithTier(value: string): string {
  const parsed = parseImageSize(value)
  if (!parsed) return value
  return `${parsed.tier} / ${parsed.width} x ${parsed.height}`
}

function formatSizeOptionLabel(value: string, label: string): string {
  const parsed = parseImageSize(value)
  return parsed ? `${parsed.tier} / ${label}` : label
}

function parseImageSize(value: string): ParsedImageSize | null {
  const match = /^(\d+)x(\d+)$/i.exec(value.trim())
  if (!match) return null
  const width = Number(match[1])
  const height = Number(match[2])
  if (!Number.isInteger(width) || !Number.isInteger(height) || width <= 0 || height <= 0) {
    return null
  }
  const pixels = width * height
  return {
    width,
    height,
    value: `${width}x${height}`,
    pixels,
    tier: classifyImageSizeTier(width, height)
  }
}

function classifyImageSizeTier(width: number, height: number): ImageSizeTier {
  const maxEdge = Math.max(width, height)
  if (maxEdge <= 1024) return '1K'
  if (maxEdge <= 2048) return '2K'
  return '4K'
}

function validateImageSize(value: string): { valid: boolean; normalized?: ImageGenerationSize; error?: string } {
  const parsed = parseImageSize(value)
  if (!parsed) {
    return { valid: false, error: t('imageGeneration.customSize.errors.format') }
  }
  const { width, height, pixels } = parsed
  if (width > MAX_IMAGE_SIDE || height > MAX_IMAGE_SIDE) {
    return { valid: false, error: t('imageGeneration.customSize.errors.maxSide', { max: MAX_IMAGE_SIDE }) }
  }
  if (width % IMAGE_SIZE_STEP !== 0 || height % IMAGE_SIZE_STEP !== 0) {
    return { valid: false, error: t('imageGeneration.customSize.errors.multiple', { step: IMAGE_SIZE_STEP }) }
  }
  const ratio = Math.max(width, height) / Math.min(width, height)
  if (ratio > MAX_IMAGE_RATIO) {
    return { valid: false, error: t('imageGeneration.customSize.errors.ratio', { ratio: `${MAX_IMAGE_RATIO}:1` }) }
  }
  if (pixels < MIN_IMAGE_PIXELS || pixels > MAX_IMAGE_PIXELS) {
    return {
      valid: false,
      error: t('imageGeneration.customSize.errors.pixels', {
        min: formatInteger(MIN_IMAGE_PIXELS),
        max: formatInteger(MAX_IMAGE_PIXELS)
      })
    }
  }
  return { valid: true, normalized: parsed.value }
}

function imagePriceForTier(group: ApiKey['group'], tier: ImageSizeTier): number | null {
  if (!group) return null
  switch (tier) {
    case '1K':
      return group.image_price_1k
    case '2K':
      return group.image_price_2k
    case '4K':
      return group.image_price_4k
    default:
      return null
  }
}

function effectiveImageMultiplier(group: ApiKey['group']): number {
  if (!group) return 1
  const value = group.image_rate_independent ? group.image_rate_multiplier : group.rate_multiplier
  return Number.isFinite(value) && value >= 0 ? value : 1
}

function formatUSD(value: number): string {
  if (!Number.isFinite(value)) return '-'
  return `$${value.toFixed(6).replace(/0+$/, '').replace(/\.$/, '')}`
}

function formatMultiplier(value: number): string {
  if (!Number.isFinite(value)) return '1x'
  return `${value.toFixed(4).replace(/0+$/, '').replace(/\.$/, '')}x`
}

function formatInteger(value: number): string {
  return new Intl.NumberFormat(locale.value).format(value)
}

function createClientRequestId(): string {
  const randomPart = Math.random().toString(36).slice(2, 10)
  return `img-${Date.now().toString(36)}-${randomPart}`
}

function normalizeGenerateCount(value: number): number {
  if (!Number.isFinite(value)) return 1
  return Math.min(Math.max(...countOptions), Math.max(1, Math.trunc(value)))
}

function startGenerationTimer() {
  stopGenerationTimer()
  generationStartedAt.value = Date.now()
  generationNow.value = generationStartedAt.value
  generationTimer = window.setInterval(() => {
    generationNow.value = Date.now()
  }, 1000)
}

function stopGenerationTimer() {
  if (generationTimer) {
    window.clearInterval(generationTimer)
    generationTimer = null
  }
}

function handleSizeSelect(event: Event) {
  const value = (event.target as HTMLSelectElement).value
  customSizeError.value = ''
  if (value === CUSTOM_SIZE_SELECT_VALUE) {
    const parsed = parseImageSize(size.value)
    customWidth.value = parsed ? String(parsed.width) : '1024'
    customHeight.value = parsed ? String(parsed.height) : '1024'
    showCustomSizeEditor.value = true
    return
  }
  size.value = value
  showCustomSizeEditor.value = false
}

function saveCustomSize() {
  const result = validateImageSize(`${customWidth.value}x${customHeight.value}`)
  if (!result.valid || !result.normalized) {
    customSizeError.value = result.error || t('imageGeneration.customSize.errors.format')
    appStore.showError(customSizeError.value)
    return
  }
  customSizeError.value = ''
  size.value = result.normalized
  if (!presetSizeOptions.value.some((option) => option.value === result.normalized) && !customSizes.value.includes(result.normalized)) {
    customSizes.value = [result.normalized, ...customSizes.value].slice(0, 12)
    saveCustomSizes()
  }
  showCustomSizeEditor.value = false
  appStore.showSuccess(t('imageGeneration.customSize.saved'))
}

function cancelCustomSize() {
  showCustomSizeEditor.value = false
  customSizeError.value = ''
}

async function loadKeys() {
  loadingKeys.value = true
  try {
    const response = await keysAPI.list(1, 100, { status: 'active' })
    keys.value = response.items
    if (!selectedKeyId.value || !imageKeys.value.some((key) => key.id === selectedKeyId.value)) {
      selectedKeyId.value = imageKeys.value[0]?.id ?? null
    }
  } catch (error) {
    const message = error instanceof Error ? error.message : t('imageGeneration.errors.loadKeysFailed')
    appStore.showError(message)
  } finally {
    loadingKeys.value = false
  }
}

async function loadSelectedKeyModels(key: ApiKey | null) {
  keyModelsAbortController?.abort()
  keyModelsAbortController = null
  keyModelIds.value = []
  keyModelsLoaded.value = false
  if (!key) return

  const controller = new AbortController()
  keyModelsAbortController = controller
  try {
    const models = await imageGenerationAPI.listModels(key.key, controller.signal)
    if (keyModelsAbortController !== controller) return
    keyModelIds.value = models
    keyModelsLoaded.value = true
  } catch (error) {
    if (error instanceof DOMException && error.name === 'AbortError') return
  } finally {
    if (keyModelsAbortController === controller) {
      keyModelsAbortController = null
    }
  }
}

function applyTemplate(value: string) {
  prompt.value = value
}

function resetForm() {
  prompt.value = ''
  model.value = modelOptions.value[0] || ''
  size.value = '1024x1024'
  quality.value = 'high'
  n.value = 1
  errorMessage.value = ''
  showCustomSizeEditor.value = false
  customSizeError.value = ''
}

async function startGenerate() {
  if (!canGenerate.value || !selectedKey.value) return
  const currentKey = selectedKey.value
  const requestedCount = normalizeGenerateCount(n.value)
  n.value = requestedCount
  const sizeValidation = validateImageSize(size.value)
  if (!sizeValidation.valid || !sizeValidation.normalized) {
    const message = sizeValidation.error || t('imageGeneration.customSize.errors.format')
    errorMessage.value = message
    appStore.showError(message)
    return
  }

  abortController?.abort()
  abortController = new AbortController()
  activeClientRequestId.value = createClientRequestId()
  generating.value = true
  startGenerationTimer()
  errorMessage.value = ''
  generatedImages.value = []

  try {
    const images = await generateRequestedImages({
      apiKey: currentKey.key,
      model: model.value,
      prompt: prompt.value,
      size: sizeValidation.normalized,
      quality: quality.value,
      requestedCount,
      clientRequestId: activeClientRequestId.value,
      signal: abortController.signal
    })
    if (images.length === 0) {
      throw new Error(t('imageGeneration.errors.noImages'))
    }
    generatedImages.value = images
    pushHistory(images.map((image) => image.url))
    if (images.length < requestedCount) {
      const message = t('imageGeneration.errors.partialImages', { count: images.length, requested: requestedCount })
      errorMessage.value = message
      appStore.showError(message)
    } else {
      appStore.showSuccess(t('imageGeneration.messages.generated'))
    }
  } catch (error) {
    if (error instanceof DOMException && error.name === 'AbortError') {
      return
    }
    const message = error instanceof Error ? error.message : t('imageGeneration.errors.generateFailed')
    errorMessage.value = message
    appStore.showError(message)
  } finally {
    stopGenerationTimer()
    generationNow.value = Date.now()
    generating.value = false
    abortController = null
  }
}

async function generateRequestedImages(params: GenerateRequestedImagesParams): Promise<GeneratedImage[]> {
  const images: GeneratedImage[] = []
  let remaining = params.requestedCount
  let requestIndex = 0

  while (remaining > 0) {
    requestIndex += 1
    const response = await imageGenerationAPI.generate({
      apiKey: params.apiKey,
      model: params.model,
      prompt: params.prompt,
      size: params.size,
      quality: params.quality,
      n: requestIndex === 1 ? remaining : 1,
      clientRequestId: requestIndex === 1 ? params.clientRequestId : `${params.clientRequestId}-${requestIndex}`,
      signal: params.signal
    })
    if (response.images.length === 0) break

    images.push(...response.images)
    generatedImages.value = images.slice(0, params.requestedCount)
    remaining = params.requestedCount - images.length
  }

  return images.slice(0, params.requestedCount)
}

function pushHistory(images: string[]) {
  if (images.length === 0) return
  historyRevision += 1
  history.value = [
    {
      id: `${Date.now()}-${Math.random().toString(36).slice(2, 8)}`,
      prompt: prompt.value.trim(),
      model: model.value.trim(),
      size: size.value,
      quality: quality.value,
      thumbnail: images[0],
      images,
      createdAt: new Intl.DateTimeFormat(locale.value, {
        month: 'short',
        day: '2-digit',
        hour: '2-digit',
        minute: '2-digit'
      }).format(new Date())
    },
    ...history.value
  ].slice(0, HISTORY_LIMIT)
  scheduleHistorySave()
}

function restoreHistory(item: HistoryItem) {
  prompt.value = item.prompt
  model.value = item.model
  size.value = item.size
  rememberCustomSize(item.size)
  quality.value = item.quality
  generatedImages.value = item.images.map((url) => ({ url, mimeType: 'image/png' }))
  errorMessage.value = ''
}

async function openHistoryDatabase(): Promise<IDBDatabase | null> {
  if (typeof window === 'undefined' || !window.indexedDB) return null
  return new Promise((resolve, reject) => {
    const request = window.indexedDB.open(HISTORY_DB_NAME, HISTORY_DB_VERSION)
    request.onupgradeneeded = () => {
      const db = request.result
      if (!db.objectStoreNames.contains(HISTORY_STORE_NAME)) {
        db.createObjectStore(HISTORY_STORE_NAME)
      }
    }
    request.onsuccess = () => resolve(request.result)
    request.onerror = () => reject(request.error)
    request.onblocked = () => resolve(null)
  })
}

function readHistoryFromDatabase(db: IDBDatabase): Promise<unknown> {
  return new Promise((resolve, reject) => {
    const transaction = db.transaction(HISTORY_STORE_NAME, 'readonly')
    const store = transaction.objectStore(HISTORY_STORE_NAME)
    const request = store.get(HISTORY_RECORD_KEY)
    request.onsuccess = () => resolve(request.result)
    request.onerror = () => reject(request.error)
  })
}

function writeHistoryToDatabase(db: IDBDatabase, items: HistoryItem[]): Promise<void> {
  return new Promise((resolve, reject) => {
    const transaction = db.transaction(HISTORY_STORE_NAME, 'readwrite')
    const store = transaction.objectStore(HISTORY_STORE_NAME)
    store.put(items, HISTORY_RECORD_KEY)
    transaction.oncomplete = () => resolve()
    transaction.onerror = () => reject(transaction.error)
    transaction.onabort = () => reject(transaction.error)
  })
}

function clearHistoryDatabase(db: IDBDatabase): Promise<void> {
  return new Promise((resolve, reject) => {
    const transaction = db.transaction(HISTORY_STORE_NAME, 'readwrite')
    const store = transaction.objectStore(HISTORY_STORE_NAME)
    store.delete(HISTORY_RECORD_KEY)
    transaction.oncomplete = () => resolve()
    transaction.onerror = () => reject(transaction.error)
    transaction.onabort = () => reject(transaction.error)
  })
}

function normalizeHistoryItems(value: unknown): HistoryItem[] {
  if (!Array.isArray(value)) return []
  return value
    .map((item): HistoryItem | null => {
      if (!item || typeof item !== 'object') return null
      const candidate = item as Partial<HistoryItem>
      const images = Array.isArray(candidate.images)
        ? candidate.images.filter((url): url is string => typeof url === 'string' && url.trim().length > 0)
        : []
      const thumbnail = typeof candidate.thumbnail === 'string' && candidate.thumbnail.trim()
        ? candidate.thumbnail
        : images[0]
      const quality = imageGenerationQualities.includes(candidate.quality as ImageGenerationQuality)
        ? candidate.quality as ImageGenerationQuality
        : 'high'
      if (
        typeof candidate.id !== 'string' ||
        typeof candidate.prompt !== 'string' ||
        typeof candidate.model !== 'string' ||
        typeof candidate.size !== 'string' ||
        typeof candidate.createdAt !== 'string' ||
        !thumbnail ||
        images.length === 0
      ) {
        return null
      }
      return {
        id: candidate.id,
        prompt: candidate.prompt,
        model: candidate.model,
        size: candidate.size,
        quality,
        thumbnail,
        images,
        createdAt: candidate.createdAt
      }
    })
    .filter((item): item is HistoryItem => item !== null)
    .slice(0, HISTORY_LIMIT)
}

function loadHistoryFromLocalStorage(): HistoryItem[] {
  try {
    const raw = localStorage.getItem(HISTORY_KEY)
    if (!raw) return []
    const parsed = JSON.parse(raw)
    return normalizeHistoryItems(parsed)
  } catch {
    return []
  }
}

function historyItemsNeedOptimization(items: HistoryItem[]): boolean {
  return items.some((item) => item.images.some(shouldOptimizeHistoryImage))
}

function mergeHistoryItems(...collections: HistoryItem[][]): HistoryItem[] {
  const seen = new Set<string>()
  const merged: HistoryItem[] = []
  collections.forEach((items) => {
    items.forEach((item) => {
      if (seen.has(item.id)) return
      seen.add(item.id)
      merged.push(item)
    })
  })
  return merged.slice(0, HISTORY_LIMIT)
}

async function loadHistory() {
  try {
    const db = await openHistoryDatabase()
    if (db) {
      try {
        const storedItems = normalizeHistoryItems(await readHistoryFromDatabase(db))
        const localItems = loadHistoryFromLocalStorage()
        const mergedItems = mergeHistoryItems(localItems, storedItems)
        if (mergedItems.length > 0) {
          history.value = mergedItems
          if (localItems.length > 0 || historyItemsNeedOptimization(mergedItems)) {
            scheduleHistorySave()
          }
          return
        }
        return
      } finally {
        db.close()
      }
    }

    history.value = loadHistoryFromLocalStorage()
  } catch {
    history.value = loadHistoryFromLocalStorage()
  }
}

function scheduleHistorySave(): void {
  historySaveQueue = historySaveQueue
    .catch(() => undefined)
    .then(() => saveHistory())
}

async function saveHistory() {
  const revision = historyRevision
  const items = await prepareHistoryItemsForStorage(history.value.slice(0, HISTORY_LIMIT))
  if (revision !== historyRevision) return
  history.value = mergePreparedHistoryItems(history.value, items)
  try {
    const db = await openHistoryDatabase()
    if (revision !== historyRevision) {
      db?.close()
      return
    }
    if (db) {
      try {
        try {
          await writeHistoryToDatabase(db, items)
        } catch (error) {
          try {
            await clearHistoryDatabase(db)
            await writeHistoryToDatabase(db, items)
          } catch {
            try {
              await clearHistoryDatabase(db)
            } catch {
              // Keep falling back to localStorage below.
            }
            throw error
          }
        }
        localStorage.removeItem(HISTORY_KEY)
        return
      } finally {
        db.close()
      }
    }
    if (revision !== historyRevision) return
    localStorage.setItem(HISTORY_KEY, JSON.stringify(items))
  } catch {
    try {
      if (revision !== historyRevision) return
      localStorage.setItem(HISTORY_KEY, JSON.stringify(items))
    } catch {
      const reducedItems = items.slice(0, Math.max(1, Math.floor(items.length / 2)))
      try {
        localStorage.setItem(HISTORY_KEY, JSON.stringify(reducedItems))
      } catch {
        // The current in-memory history remains usable if persistent storage is temporarily full.
      }
    }
  }
}

async function prepareHistoryItemsForStorage(items: HistoryItem[]): Promise<HistoryItem[]> {
  const prepared = await Promise.all(items.map(async (item) => {
    const images = await Promise.all(item.images.map(optimizeHistoryImage))
    return {
      ...item,
      thumbnail: images[0] || item.thumbnail,
      images
    }
  }))
  return prepared.slice(0, HISTORY_LIMIT)
}

function mergePreparedHistoryItems(currentItems: HistoryItem[], preparedItems: HistoryItem[]): HistoryItem[] {
  const preparedById = new Map(preparedItems.map((item) => [item.id, item]))
  return currentItems
    .slice(0, HISTORY_LIMIT)
    .map((item) => preparedById.get(item.id) || item)
}

function shouldOptimizeHistoryImage(url: string): boolean {
  return url.startsWith('data:image/') && url.length > HISTORY_IMAGE_OPTIMIZE_THRESHOLD
}

async function optimizeHistoryImage(url: string): Promise<string> {
  if (!shouldOptimizeHistoryImage(url)) return url
  try {
    const optimized = await resizeHistoryImage(url)
    return optimized.length < url.length ? optimized : url
  } catch {
    return url
  }
}

async function resizeHistoryImage(url: string): Promise<string> {
  const image = await loadHistoryImage(url)
  const sourceWidth = image.naturalWidth || image.width
  const sourceHeight = image.naturalHeight || image.height
  if (sourceWidth <= 0 || sourceHeight <= 0) return url

  const scale = Math.min(1, HISTORY_IMAGE_MAX_SIDE / Math.max(sourceWidth, sourceHeight))
  const width = Math.max(1, Math.round(sourceWidth * scale))
  const height = Math.max(1, Math.round(sourceHeight * scale))
  const canvas = document.createElement('canvas')
  canvas.width = width
  canvas.height = height
  const ctx = canvas.getContext('2d')
  if (!ctx) return url

  ctx.fillStyle = '#ffffff'
  ctx.fillRect(0, 0, width, height)
  ctx.drawImage(image, 0, 0, width, height)
  const blob = await canvasToHistoryBlob(canvas)
  return blobToDataUrl(blob)
}

function loadHistoryImage(url: string): Promise<HTMLImageElement> {
  return new Promise((resolve, reject) => {
    const image = new Image()
    image.onload = () => resolve(image)
    image.onerror = () => reject(new Error('Failed to load history image'))
    image.src = url
  })
}

function canvasToHistoryBlob(canvas: HTMLCanvasElement): Promise<Blob> {
  return new Promise((resolve, reject) => {
    canvas.toBlob((blob) => {
      if (blob) {
        resolve(blob)
      } else {
        reject(new Error('Failed to optimize history image'))
      }
    }, 'image/jpeg', HISTORY_IMAGE_QUALITY)
  })
}

function blobToDataUrl(blob: Blob): Promise<string> {
  return new Promise((resolve, reject) => {
    const reader = new FileReader()
    reader.onload = () => resolve(String(reader.result || ''))
    reader.onerror = () => reject(reader.error || new Error('Failed to read optimized image'))
    reader.readAsDataURL(blob)
  })
}

function loadCustomSizes() {
  try {
    const raw = localStorage.getItem(CUSTOM_SIZES_KEY)
    if (!raw) return
    const parsed = JSON.parse(raw)
    if (!Array.isArray(parsed)) return
    customSizes.value = parsed
      .map((value) => validateImageSize(String(value)).normalized)
      .filter((value: ImageGenerationSize | undefined): value is ImageGenerationSize => !!value)
      .filter((value, index, values) => values.indexOf(value) === index)
      .filter((value) => !presetSizeOptions.value.some((option) => option.value === value))
      .slice(0, 12)
  } catch {
    customSizes.value = []
  }
}

function saveCustomSizes() {
  try {
    localStorage.setItem(CUSTOM_SIZES_KEY, JSON.stringify(customSizes.value))
  } catch {
    // Custom sizes are convenience-only; generation still works without persistence.
  }
}

function rememberCustomSize(value: ImageGenerationSize) {
  const result = validateImageSize(value)
  if (!result.valid || !result.normalized) return
  if (presetSizeOptions.value.some((option) => option.value === result.normalized)) return
  if (customSizes.value.includes(result.normalized)) return
  customSizes.value = [result.normalized, ...customSizes.value].slice(0, 12)
  saveCustomSizes()
}

async function clearHistory() {
  history.value = []
  localStorage.removeItem(HISTORY_KEY)
  try {
    const db = await openHistoryDatabase()
    if (!db) return
    try {
      await clearHistoryDatabase(db)
    } finally {
      db.close()
    }
  } catch {
    // Clearing local state is enough if persistent storage is temporarily unavailable.
  }
}

async function copyImage(url: string) {
  try {
    await navigator.clipboard.writeText(url)
    appStore.showSuccess(t('common.copied'))
  } catch {
    appStore.showError(t('common.copyFailed'))
  }
}

async function copyClientRequestId() {
  if (!activeClientRequestId.value) return
  try {
    await navigator.clipboard.writeText(activeClientRequestId.value)
    appStore.showSuccess(t('imageGeneration.messages.requestIdCopied'))
  } catch {
    appStore.showError(t('common.copyFailed'))
  }
}

function downloadImage(url: string, index: number) {
  const link = document.createElement('a')
  link.href = url
  link.download = `ai-image-${new Date().toISOString().replace(/[:.]/g, '-')}-${index + 1}.png`
  document.body.appendChild(link)
  link.click()
  document.body.removeChild(link)
}

function downloadAll() {
  generatedImages.value.forEach((image, index) => {
    window.setTimeout(() => downloadImage(image.url, index), index * 120)
  })
}

watch(imageKeys, (items) => {
  if (!selectedKeyId.value && items.length > 0) {
    selectedKeyId.value = items[0].id
  }
})

watch(modelOptions, (options) => {
  if (options.length === 0) {
    model.value = ''
    return
  }
  if (!options.includes(model.value)) {
    model.value = options[0]
  }
}, { immediate: true })

watch(selectedKey, (key) => {
  loadSelectedKeyModels(key)
}, { immediate: true })

onMounted(() => {
  loadCustomSizes()
  void loadHistory()
  loadKeys()
})

onBeforeUnmount(() => {
  abortController?.abort()
  keyModelsAbortController?.abort()
  stopGenerationTimer()
})
</script>
