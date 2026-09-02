<template>
  <Teleport to="body">
    <Transition name="blessing-dialog">
      <div
        v-if="modelValue"
        class="fixed inset-0 z-[100000010] flex items-center justify-center overflow-y-auto bg-gray-950/55 p-4 backdrop-blur-sm"
        role="dialog"
        aria-modal="true"
        :aria-labelledby="titleId"
        :aria-describedby="descriptionId"
      >
        <section
          class="relative w-full max-w-lg overflow-hidden rounded-lg border border-gray-200 bg-white shadow-2xl dark:border-dark-700 dark:bg-dark-900"
        >
          <button
            type="button"
            class="absolute right-4 top-4 z-10 flex h-9 w-9 items-center justify-center rounded-md text-gray-400 transition hover:bg-gray-100 hover:text-gray-700 focus:outline-none focus:ring-2 focus:ring-primary-400 dark:hover:bg-dark-700 dark:hover:text-white"
            :aria-label="t('payment.blessing.closeLabel')"
            @click="close"
          >
            <Icon name="x" size="sm" />
          </button>

          <div class="relative px-6 pb-7 pt-9 text-center sm:px-10 sm:pb-9 sm:pt-10">
            <div class="relative mx-auto flex h-16 w-16 items-center justify-center">
              <div class="relative flex h-16 w-16 items-center justify-center overflow-hidden rounded-lg bg-white shadow-md ring-1 ring-gray-200 dark:bg-dark-800 dark:ring-dark-600">
                <img :src="siteLogo" alt="" class="h-full w-full object-contain" />
              </div>
              <span class="absolute -right-1 -top-1 flex h-6 w-6 items-center justify-center rounded-full border-2 border-white bg-emerald-500 text-white shadow-sm dark:border-dark-900">
                <Icon name="check" size="xs" :stroke-width="2.8" />
              </span>
            </div>

            <p class="mt-5 text-xs font-semibold text-primary-600 dark:text-primary-300">
              {{ t('payment.blessing.eyebrow') }}
            </p>
            <h2 :id="titleId" class="mt-2 text-2xl font-semibold text-gray-950 dark:text-white">
              {{ t('payment.blessing.title') }}
            </h2>
            <p :id="descriptionId" class="mt-2 text-sm leading-6 text-gray-500 dark:text-gray-400">
              {{ t('payment.blessing.description') }}
            </p>

            <div class="relative mt-6 rounded-lg border border-primary-100 bg-primary-50/60 px-6 py-6 dark:border-primary-900/50 dark:bg-primary-950/20 sm:px-8">
              <span class="absolute left-4 top-2 font-serif text-3xl leading-none text-primary-200 dark:text-primary-700">&ldquo;</span>
              <p class="relative text-base font-medium leading-8 text-gray-800 dark:text-gray-100 sm:text-lg">
                {{ blessing }}
              </p>
              <span class="absolute bottom-0 right-4 font-serif text-3xl leading-none text-primary-200 dark:text-primary-700">&rdquo;</span>
            </div>

            <button
              type="button"
              class="mt-6 inline-flex min-h-12 w-full items-center justify-center gap-2 rounded-lg bg-primary-600 px-6 text-sm font-semibold text-white shadow-sm transition hover:bg-primary-700 focus:outline-none focus:ring-2 focus:ring-primary-400 focus:ring-offset-2 dark:focus:ring-offset-dark-900"
              @click="close"
            >
              <Icon name="gift" size="sm" />
              {{ t('payment.blessing.accept') }}
            </button>
          </div>
        </section>
      </div>
    </Transition>
  </Teleport>
</template>

<script setup lang="ts">
import { computed, onBeforeUnmount, onMounted, ref, watch } from 'vue'
import { useI18n } from 'vue-i18n'
import Icon from '@/components/icons/Icon.vue'
import { useAppStore } from '@/stores'
import { BRAND_LOGO_URL } from '@/constants/brand'
import { randomPaymentBlessing } from './paymentBlessings'

const props = withDefaults(defineProps<{
  modelValue: boolean
  orderKey?: string | number
}>(), {
  orderKey: '',
})

const emit = defineEmits<{
  'update:modelValue': [value: boolean]
}>()

const i18n = useI18n()
const { t } = i18n
const appStore = useAppStore()
const blessing = ref('')
const instanceId = Math.random().toString(36).slice(2, 9)
const titleId = `payment-blessing-title-${instanceId}`
const descriptionId = `payment-blessing-description-${instanceId}`
let previousBodyOverflow = ''
let isScrollLocked = false
const siteLogo = computed(() => appStore.siteLogo || BRAND_LOGO_URL)

const currentLocale = computed(() => {
  const locale = i18n.locale as unknown
  if (typeof locale === 'string') return locale
  if (locale && typeof locale === 'object' && 'value' in locale) {
    return String((locale as { value?: string }).value || '')
  }
  return 'en'
})

const seenStorageKey = computed(() => {
  const orderKey = String(props.orderKey || '').trim()
  return orderKey ? `58token:payment-blessing:${orderKey}` : ''
})

function wasAlreadyShown(): boolean {
  if (!seenStorageKey.value || typeof window === 'undefined') return false
  try {
    return window.sessionStorage.getItem(seenStorageKey.value) === '1'
  } catch {
    return false
  }
}

function markAsShown(): void {
  if (!seenStorageKey.value || typeof window === 'undefined') return
  try {
    window.sessionStorage.setItem(seenStorageKey.value, '1')
  } catch {
    // Storage can be unavailable in strict privacy modes; showing the message still works.
  }
}

function chooseBlessing(): void {
  blessing.value = randomPaymentBlessing(currentLocale.value)
}

function close(): void {
  emit('update:modelValue', false)
}

function setPageScrollLocked(locked: boolean): void {
  if (typeof document === 'undefined') return
  if (locked && !isScrollLocked) {
    previousBodyOverflow = document.body.style.overflow
    document.body.style.overflow = 'hidden'
    isScrollLocked = true
    return
  }
  if (!locked && isScrollLocked) {
    document.body.style.overflow = previousBodyOverflow
    isScrollLocked = false
  }
}

function handleEscape(event: KeyboardEvent): void {
  if (props.modelValue && event.key === 'Escape') close()
}

watch(
  () => props.modelValue,
  (isOpen) => {
    setPageScrollLocked(isOpen)
    if (!isOpen) return
    if (wasAlreadyShown()) {
      close()
      return
    }
    markAsShown()
    chooseBlessing()
  },
  { immediate: true },
)

watch(currentLocale, () => {
  if (props.modelValue) chooseBlessing()
})

onMounted(() => window.addEventListener('keydown', handleEscape))
onBeforeUnmount(() => {
  window.removeEventListener('keydown', handleEscape)
  setPageScrollLocked(false)
})
</script>

<style scoped>
.blessing-dialog-enter-active,
.blessing-dialog-leave-active {
  transition: opacity 220ms ease;
}

.blessing-dialog-enter-active section,
.blessing-dialog-leave-active section {
  transition: transform 260ms cubic-bezier(.2, .8, .2, 1), opacity 220ms ease;
}

.blessing-dialog-enter-from,
.blessing-dialog-leave-to {
  opacity: 0;
}

.blessing-dialog-enter-from section,
.blessing-dialog-leave-to section {
  opacity: 0;
  transform: translate3d(0, 18px, 0) scale(.97);
}

@media (prefers-reduced-motion: reduce) {
  .blessing-dialog-enter-active,
  .blessing-dialog-leave-active,
  .blessing-dialog-enter-active section,
  .blessing-dialog-leave-active section {
    transition-duration: 1ms;
  }
}
</style>
