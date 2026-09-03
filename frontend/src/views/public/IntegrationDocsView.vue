<template>
  <div class="relative flex min-h-screen flex-col overflow-x-clip bg-[#f7f8fb] text-gray-950 dark:bg-[#05060a] dark:text-white">
    <div class="pointer-events-none absolute inset-0 overflow-hidden">
      <div class="absolute left-1/2 top-[-18rem] h-[42rem] w-[42rem] -translate-x-1/2 rounded-full bg-cyan-300/30 blur-3xl dark:bg-cyan-400/20"></div>
      <div class="absolute right-[-14rem] top-24 h-[34rem] w-[34rem] rounded-full bg-violet-300/25 blur-3xl dark:bg-violet-500/18"></div>
      <div class="absolute bottom-[-18rem] left-[-10rem] h-[38rem] w-[38rem] rounded-full bg-emerald-300/20 blur-3xl dark:bg-emerald-400/10"></div>
      <div class="absolute inset-0 bg-[linear-gradient(rgba(15,23,42,0.045)_1px,transparent_1px),linear-gradient(90deg,rgba(15,23,42,0.045)_1px,transparent_1px)] bg-[size:72px_72px] [mask-image:radial-gradient(circle_at_top,black,transparent_78%)] dark:bg-[linear-gradient(rgba(255,255,255,0.045)_1px,transparent_1px),linear-gradient(90deg,rgba(255,255,255,0.045)_1px,transparent_1px)]"></div>
    </div>

    <MarketingNavbar
      :site-name="siteName"
      :subtitle="siteSubtitle"
      :logo="siteLogo"
      :doc-url="docUrl"
      docs-to="/docs"
      :docs-label="t('nav.integrationDocs')"
      :cta-to="isAuthenticated ? dashboardPath : '/login'"
      :cta-label="isAuthenticated ? t('home.dashboard') : t('home.login')"
      model-marketplace-to="/available-channels"
      :model-marketplace-label="t('nav.availableChannels')"
      partner-to="/partners"
      :partner-label="t('gateway.common.partner')"
      resources-to="/downloads"
      :resources-label="t('gateway.common.resources')"
      :visible-items="marketingNavItems"
    >
      <template #tools>
        <LocaleSwitcher />
        <button
          type="button"
          class="rounded-xl border border-gray-200/70 bg-white/70 p-2 text-gray-600 transition hover:bg-white hover:text-gray-950 dark:border-white/10 dark:bg-white/[0.04] dark:text-slate-300 dark:hover:bg-white/10 dark:hover:text-white"
          :title="isDark ? t('home.switchToLight') : t('home.switchToDark')"
          @click="toggleTheme"
        >
          <Icon v-if="isDark" name="sun" size="md" />
          <Icon v-else name="moon" size="md" />
        </button>
      </template>
    </MarketingNavbar>

    <main class="relative z-10 flex-1 px-4 pb-16 pt-6 sm:px-6">
      <section class="mx-auto max-w-7xl overflow-hidden rounded-[2rem] border border-gray-200/70 bg-white/72 p-5 shadow-[0_26px_90px_rgba(15,23,42,0.10)] backdrop-blur-2xl dark:border-white/10 dark:bg-white/[0.045] dark:shadow-[0_26px_90px_rgba(0,0,0,0.30)] md:p-8">
        <div class="grid gap-8 lg:grid-cols-[minmax(0,1fr)_360px] lg:items-end">
          <div>
            <div class="inline-flex items-center gap-2 rounded-full border border-primary-200 bg-primary-50 px-3 py-1 text-xs font-semibold text-primary-700 dark:border-primary-800/60 dark:bg-primary-900/20 dark:text-primary-300">
              <Icon name="book" size="sm" />
              {{ t('integrationDocs.hero.eyebrow') }}
            </div>
            <h1 class="mt-5 max-w-4xl text-4xl font-semibold tracking-normal text-gray-950 dark:text-white md:text-6xl">
              {{ t('integrationDocs.title') }}
            </h1>
            <p class="mt-5 max-w-3xl text-base leading-8 text-gray-600 dark:text-slate-300">
              {{ t('integrationDocs.hero.subtitle') }}
            </p>
            <div class="mt-6 flex flex-wrap gap-3">
              <RouterLink
                to="/register"
                class="inline-flex items-center justify-center rounded-2xl border border-gray-900/10 bg-gray-950 px-5 py-3 text-sm font-semibold text-white shadow-[0_18px_44px_rgba(15,23,42,0.18)] transition hover:-translate-y-0.5 hover:bg-gray-800 dark:border-white/15 dark:bg-white dark:text-slate-950 dark:hover:bg-slate-100"
              >
                {{ t('integrationDocs.hero.primaryCta') }}
              </RouterLink>
              <RouterLink
                to="/available-channels"
                class="inline-flex items-center justify-center rounded-2xl border border-gray-200/80 bg-white/75 px-5 py-3 text-sm font-semibold text-gray-700 shadow-sm backdrop-blur-xl transition hover:border-primary-300 hover:text-primary-700 dark:border-white/10 dark:bg-white/[0.05] dark:text-slate-200 dark:hover:border-primary-700 dark:hover:text-primary-300"
              >
                {{ t('integrationDocs.hero.secondaryCta') }}
              </RouterLink>
            </div>
          </div>

          <div class="rounded-[1.5rem] border border-emerald-200/80 bg-white/62 p-5 shadow-sm backdrop-blur dark:border-emerald-900/50 dark:bg-white/[0.04]">
            <p class="text-xs font-semibold uppercase tracking-[0.18em] text-emerald-700 dark:text-emerald-300">
              {{ t('integrationDocs.baseUrl.label') }}
            </p>
            <div class="mt-3 space-y-3">
              <div>
                <p class="text-xs font-semibold text-gray-500 dark:text-slate-400">{{ t('integrationDocs.baseUrl.gatewayLabel') }}</p>
                <div class="mt-1 flex items-center gap-2 rounded-2xl border border-gray-200 bg-white/80 px-4 py-3 dark:border-white/10 dark:bg-slate-950/60">
                  <code class="min-w-0 flex-1 break-all font-mono text-sm text-gray-950 dark:text-slate-100">{{ gatewayOrigin }}</code>
                  <button
                    type="button"
                    class="inline-flex shrink-0 items-center gap-1.5 rounded-xl border border-gray-200 bg-white px-3 py-2 text-xs font-semibold text-gray-700 shadow-sm transition hover:border-primary-300 hover:text-primary-700 dark:border-white/10 dark:bg-white/10 dark:text-slate-100 dark:hover:bg-white/15"
                    @click="copyConfig(gatewayOrigin, 'gateway-origin')"
                  >
                    <Icon :name="copiedKey === 'gateway-origin' ? 'checkCircle' : 'copy'" size="xs" />
                    {{ copiedKey === 'gateway-origin' ? t('common.copied') : t('common.copy') }}
                  </button>
                </div>
                <p class="mt-2 text-xs leading-5 text-gray-500 dark:text-slate-400">{{ t('integrationDocs.baseUrl.gatewayDescription') }}</p>
              </div>
            </div>
          </div>
        </div>
      </section>

      <div class="mx-auto mt-6 grid max-w-7xl gap-6 md:grid-cols-[240px_minmax(0,1fr)] xl:grid-cols-[280px_minmax(0,1fr)]">
        <div ref="tocColumnRef" class="relative z-20 md:min-h-full">
          <aside
            ref="tocPanelRef"
            class="z-20 max-h-[calc(100vh-9rem)] self-start overflow-y-auto rounded-[1.5rem] border border-gray-200/70 bg-white/85 p-5 shadow-sm backdrop-blur-2xl dark:border-white/10 dark:bg-slate-950/75"
            :style="tocPanelStyle"
          >
            <div class="flex items-center justify-between gap-3">
              <h2 class="text-base font-semibold text-gray-950 dark:text-white">{{ t('integrationDocs.toc.title') }}</h2>
              <span class="font-mono text-xs font-semibold text-primary-700 dark:text-primary-300">{{ scrollProgressPercent }}%</span>
            </div>
            <div class="mt-3 h-1.5 overflow-hidden rounded-full bg-gray-100 dark:bg-white/10">
              <div class="h-full rounded-full bg-primary-500 transition-[width] duration-150 dark:bg-primary-300" :style="{ width: `${scrollProgressPercent}%` }"></div>
            </div>
            <nav class="mt-4 space-y-2">
              <div v-for="item in tocItems" :key="item.id">
                <div
                  v-if="item.id === 'clients'"
                  :class="tocGroupClass(item.id)"
                >
                  <div class="flex items-center">
                    <a
                      :href="`#${item.id}`"
                      :class="tocLinkClass(item.id, true)"
                    >
                      {{ item.label }}
                    </a>
                    <button
                      type="button"
                      class="mr-1 rounded-lg p-2 text-gray-500 transition hover:bg-white hover:text-gray-950 dark:text-slate-400 dark:hover:bg-white/10 dark:hover:text-white"
                      :aria-expanded="clientsTocExpanded"
                      :aria-label="t('integrationDocs.toc.toggleClients')"
                      @click="clientsTocExpanded = !clientsTocExpanded"
                    >
                      <Icon :name="clientsTocExpanded ? 'chevronUp' : 'chevronDown'" size="xs" />
                    </button>
                  </div>
                  <div v-if="clientsTocExpanded" class="space-y-1 pb-2 pl-3 pr-2">
                    <a
                      v-for="guide in clientGuides"
                      :key="`toc-${guide.id}`"
                      :href="`#${guide.id}`"
                      :class="tocSubLinkClass(guide.id)"
                    >
                      {{ guide.title }}
                    </a>
                  </div>
                </div>
                <a
                  v-else
                  :href="`#${item.id}`"
                  :class="tocLinkClass(item.id)"
                >
                  <span>{{ item.label }}</span>
                  <Icon name="chevronRight" size="xs" />
                </a>
              </div>
            </nav>
          </aside>
        </div>

        <section class="min-w-0 space-y-6">
          <DocsSection id="quick-start" :title="t('integrationDocs.quickStart.title')" :description="t('integrationDocs.quickStart.description')">
            <div class="grid gap-3 md:grid-cols-2">
              <article v-for="(step, index) in quickStartSteps" :key="step.title" class="rounded-2xl border border-gray-200/70 bg-white/72 p-5 shadow-sm dark:border-white/10 dark:bg-white/[0.04]">
                <div class="flex items-center gap-3">
                  <span class="flex h-8 w-8 shrink-0 items-center justify-center rounded-xl bg-primary-100 text-sm font-semibold text-primary-700 dark:bg-primary-900/30 dark:text-primary-300">{{ index + 1 }}</span>
                  <h3 class="text-base font-semibold text-gray-950 dark:text-white">{{ step.title }}</h3>
                </div>
                <p class="mt-3 text-sm leading-6 text-gray-600 dark:text-slate-400">{{ step.description }}</p>
              </article>
            </div>
          </DocsSection>

          <DocsSection id="concepts" :title="t('integrationDocs.concepts.title')" :description="t('integrationDocs.concepts.description')">
            <div class="grid gap-3 lg:grid-cols-3">
              <article v-for="item in conceptCards" :key="item.title" class="rounded-2xl border border-gray-200/70 bg-white/72 p-5 shadow-sm dark:border-white/10 dark:bg-white/[0.04]">
                <h3 class="text-base font-semibold text-gray-950 dark:text-white">{{ item.title }}</h3>
                <p class="mt-3 text-sm leading-6 text-gray-600 dark:text-slate-400">{{ item.description }}</p>
              </article>
            </div>
          </DocsSection>

          <DocsSection id="endpoints" :title="t('integrationDocs.endpoints.title')" :description="t('integrationDocs.endpoints.description')">
            <div class="overflow-hidden rounded-2xl border border-gray-200/70 dark:border-white/10">
              <div class="overflow-x-auto">
                <table class="min-w-[760px] w-full text-left text-sm">
                  <thead class="bg-gray-50/90 text-xs uppercase text-gray-500 dark:bg-white/[0.04] dark:text-slate-400">
                    <tr>
                      <th class="px-4 py-3">{{ t('integrationDocs.endpoints.columns.endpoint') }}</th>
                      <th class="px-4 py-3">{{ t('integrationDocs.endpoints.columns.method') }}</th>
                      <th class="px-4 py-3">{{ t('integrationDocs.endpoints.columns.auth') }}</th>
                      <th class="px-4 py-3">{{ t('integrationDocs.endpoints.columns.usage') }}</th>
                    </tr>
                  </thead>
                  <tbody>
                    <tr v-for="endpoint in endpointRows" :key="endpoint.endpoint" class="border-t border-gray-100 dark:border-white/10">
                      <td class="px-4 py-4 font-mono text-xs text-gray-950 dark:text-white">{{ endpoint.endpoint }}</td>
                      <td class="px-4 py-4 text-gray-600 dark:text-slate-300">{{ endpoint.method }}</td>
                      <td class="px-4 py-4 text-gray-600 dark:text-slate-300">{{ endpoint.auth }}</td>
                      <td class="px-4 py-4 text-gray-600 dark:text-slate-300">{{ endpoint.usage }}</td>
                    </tr>
                  </tbody>
                </table>
              </div>
            </div>
          </DocsSection>

          <DocsSection id="clients" :title="clientSectionTitle" :description="clientSectionDescription">
            <div class="mb-5 overflow-hidden rounded-2xl border border-primary-200/70 dark:border-primary-900/50">
              <div class="overflow-x-auto">
                <table class="min-w-[760px] w-full text-left text-sm">
                  <thead class="bg-primary-50/80 text-xs uppercase text-primary-700 dark:bg-primary-900/20 dark:text-primary-300">
                    <tr>
                      <th class="px-4 py-3">{{ t('integrationDocs.clients.matrix.columns.client') }}</th>
                      <th class="px-4 py-3">{{ t('integrationDocs.clients.matrix.columns.field') }}</th>
                      <th class="px-4 py-3">{{ t('integrationDocs.clients.matrix.columns.baseUrl') }}</th>
                      <th class="px-4 py-3">{{ t('integrationDocs.clients.matrix.columns.reason') }}</th>
                    </tr>
                  </thead>
                  <tbody>
                    <tr v-for="row in clientMatrixRows" :key="`${row.client}-${row.field}`" class="border-t border-primary-100/70 dark:border-primary-900/40">
                      <td class="px-4 py-4 font-semibold text-gray-950 dark:text-white">{{ row.client }}</td>
                      <td class="px-4 py-4 text-gray-600 dark:text-slate-300">{{ row.field }}</td>
                      <td class="px-4 py-4">
                        <code class="rounded-lg bg-white/80 px-2 py-1 font-mono text-xs text-gray-950 dark:bg-slate-950/60 dark:text-slate-100">{{ row.baseUrl }}</code>
                      </td>
                      <td class="px-4 py-4 text-gray-600 dark:text-slate-300">{{ row.reason }}</td>
                    </tr>
                  </tbody>
                </table>
              </div>
            </div>

            <div class="grid gap-4">
              <article v-for="guide in clientGuides" :id="guide.id" :key="guide.title" class="scroll-mt-6 rounded-2xl border border-gray-200/70 bg-white/72 p-5 shadow-sm dark:border-white/10 dark:bg-white/[0.04]">
                <div class="flex flex-col gap-3 md:flex-row md:items-start md:justify-between">
                  <div>
                    <div class="flex flex-wrap items-center gap-2">
                      <h3 class="text-lg font-semibold text-gray-950 dark:text-white">{{ guide.title }}</h3>
                      <span class="rounded-full border border-primary-200 bg-primary-50 px-2.5 py-1 text-xs font-semibold text-primary-700 dark:border-primary-800/60 dark:bg-primary-900/20 dark:text-primary-300">{{ guide.badge }}</span>
                    </div>
                    <p class="mt-2 text-sm leading-6 text-gray-600 dark:text-slate-400">{{ guide.description }}</p>
                  </div>
                  <div class="rounded-2xl border border-emerald-200 bg-emerald-50/70 px-4 py-3 text-sm dark:border-emerald-900/50 dark:bg-emerald-900/20">
                    <div class="flex items-start justify-between gap-3">
                      <div class="min-w-0">
                        <p class="font-semibold text-emerald-800 dark:text-emerald-200">{{ guide.baseUrlLabel }}</p>
                        <code class="mt-1 block break-all font-mono text-emerald-950 dark:text-emerald-100">{{ guide.baseUrl }}</code>
                      </div>
                      <button
                        type="button"
                        class="inline-flex shrink-0 items-center gap-1.5 rounded-xl border border-emerald-300/80 bg-white/75 px-3 py-2 text-xs font-semibold text-emerald-800 shadow-sm transition hover:border-emerald-400 hover:bg-white dark:border-emerald-800/70 dark:bg-emerald-950/45 dark:text-emerald-100 dark:hover:bg-emerald-900/55"
                        @click="copyConfig(guide.baseUrl, `base-url-${guide.title}`)"
                      >
                        <Icon :name="copiedKey === `base-url-${guide.title}` ? 'checkCircle' : 'copy'" size="xs" />
                        {{ copiedKey === `base-url-${guide.title}` ? t('common.copied') : t('common.copy') }}
                      </button>
                    </div>
                  </div>
                </div>

                <p class="mt-4 rounded-2xl border border-gray-200/70 bg-gray-50/80 p-4 text-sm leading-6 text-gray-600 dark:border-white/10 dark:bg-slate-950/35 dark:text-slate-300">
                  {{ guide.baseUrlNote }}
                </p>

                <div class="mt-4 grid gap-3 md:grid-cols-3">
                  <div v-for="os in guide.osRows" :key="os.system" class="rounded-2xl border border-gray-200/70 bg-white/72 p-4 dark:border-white/10 dark:bg-white/[0.03]">
                    <p class="text-sm font-semibold text-gray-950 dark:text-white">{{ os.system }}</p>
                    <p class="mt-2 break-all font-mono text-xs text-gray-500 dark:text-slate-400">{{ os.path }}</p>
                    <code class="mt-2 block break-all rounded-xl bg-gray-50 px-3 py-2 font-mono text-xs text-gray-700 dark:bg-slate-950/50 dark:text-slate-200">{{ os.command }}</code>
                  </div>
                </div>

                <div class="mt-4 grid gap-3">
                  <div v-for="block in guide.configBlocks" :key="block.title" class="overflow-hidden rounded-2xl border border-gray-200/70 bg-slate-950 dark:border-white/10">
                    <div class="flex items-center justify-between gap-3 border-b border-white/10 px-4 py-3">
                      <div>
                        <p class="text-sm font-semibold text-white">{{ block.title }}</p>
                        <p class="mt-1 text-xs text-slate-400">{{ block.description }}</p>
                      </div>
                      <button
                        type="button"
                        class="inline-flex shrink-0 items-center gap-1.5 rounded-xl border border-white/10 bg-white/10 px-3 py-2 text-xs font-semibold text-slate-100 transition hover:bg-white/15"
                        @click="copyConfig(block.code, block.title)"
                      >
                        <Icon :name="copiedKey === block.title ? 'checkCircle' : 'copy'" size="xs" />
                        {{ copiedKey === block.title ? t('common.copied') : t('common.copy') }}
                      </button>
                    </div>
                    <pre class="overflow-x-auto p-4 text-xs leading-6 text-slate-100"><code>{{ block.code }}</code></pre>
                  </div>
                </div>
              </article>
            </div>
          </DocsSection>

          <DocsSection id="examples" :title="t('integrationDocs.examples.title')" :description="t('integrationDocs.examples.description')">
            <div class="grid gap-4">
              <article v-for="example in codeExamples" :key="example.title" class="overflow-hidden rounded-2xl border border-gray-200/70 bg-slate-950 shadow-sm dark:border-white/10">
                <div class="flex items-center justify-between border-b border-white/10 px-4 py-3">
                  <div>
                    <h3 class="text-sm font-semibold text-white">{{ example.title }}</h3>
                    <p class="mt-1 text-xs text-slate-400">{{ example.description }}</p>
                  </div>
                  <div class="flex shrink-0 items-center gap-2">
                    <span class="rounded-full bg-white/10 px-2.5 py-1 text-xs font-medium text-slate-300">{{ example.badge }}</span>
                    <button
                      type="button"
                      class="inline-flex items-center gap-1.5 rounded-xl border border-white/10 bg-white/10 px-3 py-2 text-xs font-semibold text-slate-100 transition hover:bg-white/15"
                      @click="copyConfig(example.code, `example-${example.title}`)"
                    >
                      <Icon :name="copiedKey === `example-${example.title}` ? 'checkCircle' : 'copy'" size="xs" />
                      {{ copiedKey === `example-${example.title}` ? t('common.copied') : t('common.copy') }}
                    </button>
                  </div>
                </div>
                <pre class="overflow-x-auto p-4 text-xs leading-6 text-slate-100"><code>{{ example.code }}</code></pre>
              </article>
            </div>
          </DocsSection>

          <DocsSection id="workflow" :title="t('integrationDocs.workflow.title')" :description="t('integrationDocs.workflow.description')">
            <div class="grid gap-4 lg:grid-cols-2">
              <article v-for="section in workflowSections" :key="section.title" class="rounded-2xl border border-gray-200/70 bg-white/72 p-5 shadow-sm dark:border-white/10 dark:bg-white/[0.04]">
                <h3 class="text-base font-semibold text-gray-950 dark:text-white">{{ section.title }}</h3>
                <p class="mt-2 text-sm leading-6 text-gray-600 dark:text-slate-400">{{ section.description }}</p>
                <ul class="mt-4 space-y-2">
                  <li v-for="bullet in section.bullets" :key="bullet" class="flex gap-2 text-sm leading-6 text-gray-600 dark:text-slate-300">
                    <Icon name="checkCircle" size="sm" class="mt-0.5 shrink-0 text-primary-500" />
                    <span>{{ bullet }}</span>
                  </li>
                </ul>
              </article>
            </div>
          </DocsSection>

          <DocsSection id="troubleshooting" :title="t('integrationDocs.troubleshooting.title')" :description="t('integrationDocs.troubleshooting.description')">
            <div class="grid gap-3">
              <article v-for="(item, index) in troubleshootingItems" :key="item.title" class="rounded-2xl border border-gray-200/70 bg-white/72 p-5 shadow-sm dark:border-white/10 dark:bg-white/[0.04]">
                <h3 class="text-base font-semibold text-gray-950 dark:text-white">{{ item.title }}</h3>
                <p class="mt-2 text-sm leading-6 text-gray-600 dark:text-slate-400">{{ item.description }}</p>
                <a
                  v-if="index === 1"
                  href="#clients"
                  class="mt-4 inline-flex items-center gap-2 rounded-xl border border-primary-200 bg-primary-50 px-3 py-2 text-sm font-semibold text-primary-700 transition hover:border-primary-300 hover:bg-primary-100 dark:border-primary-800/60 dark:bg-primary-900/20 dark:text-primary-300 dark:hover:bg-primary-900/35"
                >
                  {{ t('integrationDocs.troubleshooting.viewClientConfig') }}
                  <Icon name="arrowRight" size="xs" />
                </a>
              </article>
            </div>
          </DocsSection>
        </section>
      </div>
    </main>
  </div>
</template>

<script setup lang="ts">
import { computed, defineComponent, h, onMounted, onUnmounted, ref } from 'vue'
import { useI18n } from 'vue-i18n'
import { useAuthStore, useAppStore } from '@/stores'
import LocaleSwitcher from '@/components/common/LocaleSwitcher.vue'
import Icon from '@/components/icons/Icon.vue'
import MarketingNavbar from '@/components/marketing/MarketingNavbar.vue'
import { BRAND_LOGO_URL } from '@/constants/brand'

interface LocalizedCard {
  title: string
  description: string
  bullets: string[]
}

interface EndpointRow {
  endpoint: string
  method: string
  auth: string
  usage: string
}

interface CodeExample {
  title: string
  description: string
  badge: string
  code: string
}

interface ClientMatrixRow {
  client: string
  field: string
  baseUrl: string
  reason: string
}

interface ClientGuide {
  id: string
  title: string
  badge: string
  description: string
  baseUrlLabel: string
  baseUrl: string
  baseUrlNote: string
  osRows: Array<{
    system: string
    path: string
    command: string
  }>
  configBlocks: Array<{
    title: string
    description: string
    code: string
  }>
}

const { t, tm, rt, te, locale } = useI18n()
const authStore = useAuthStore()
const appStore = useAppStore()

const siteName = computed(() => appStore.cachedPublicSettings?.site_name || appStore.siteName || '58Token')
const siteLogo = computed(() => appStore.siteLogo || BRAND_LOGO_URL)
const siteSubtitle = computed(() => appStore.cachedPublicSettings?.site_subtitle || t('home.heroSubtitle'))
const docUrl = computed(() => appStore.cachedPublicSettings?.doc_url || appStore.docUrl || '')
const marketingNavItems = computed(() => appStore.cachedPublicSettings?.marketing_nav_items)
const isAuthenticated = computed(() => authStore.isAuthenticated)
const isAdmin = computed(() => authStore.isAdmin)
const dashboardPath = computed(() => (isAdmin.value ? '/admin/dashboard' : '/dashboard'))
const isDark = ref(document.documentElement.classList.contains('dark'))
const copiedKey = ref('')
const clientsTocExpanded = ref(false)
const activeTocId = ref('quick-start')
const scrollProgressPercent = ref(0)
const tocColumnRef = ref<HTMLElement | null>(null)
const tocPanelRef = ref<HTMLElement | null>(null)
const tocPanelStyle = ref<Record<string, string>>({})
const normalizeOrigin = (value: string) => value.replace(/\/+$/, '')
const origin = computed(() => (typeof window === 'undefined' ? 'https://58token.vip' : window.location.origin))
const gatewayOrigin = computed(() => {
  if (origin.value.includes('localhost') || origin.value.includes('127.0.0.1')) {
    return 'https://58token.vip'
  }
  return normalizeOrigin(origin.value)
})
const openAiSdkBaseUrl = computed(() => `${gatewayOrigin.value}/v1`)
const clientSectionTitle = computed(() => tr('integrationDocs.clients.title', '客户端配置', 'Client Setup'))
const clientSectionDescription = computed(() => tr(
  'integrationDocs.clients.description',
  '按客户端复制正确的接入地址。不同客户端对接口路径的拼接方式不同，所以 Base URL 不一定相同。',
  'Copy the right address for each client. Different clients append endpoint paths differently, so the Base URL is not always the same.'
))

const tocItems = computed(() => [
  { id: 'quick-start', label: t('integrationDocs.quickStart.title') },
  { id: 'concepts', label: t('integrationDocs.concepts.title') },
  { id: 'endpoints', label: t('integrationDocs.endpoints.title') },
  { id: 'clients', label: clientSectionTitle.value },
  { id: 'examples', label: t('integrationDocs.examples.title') },
  { id: 'workflow', label: t('integrationDocs.workflow.title') },
  { id: 'troubleshooting', label: t('integrationDocs.troubleshooting.title') }
])

const quickStartSteps = computed(() => localizedCards('integrationDocs.quickStart.steps'))
const conceptCards = computed(() => localizedCards('integrationDocs.concepts.cards'))
const workflowSections = computed(() => localizedCards('integrationDocs.workflow.sections'))
const troubleshootingItems = computed(() => localizedCards('integrationDocs.troubleshooting.items'))
const endpointRows = computed(() => localizedRows<EndpointRow>('integrationDocs.endpoints.rows'))
const clientMatrixRows = computed(() => localizedRows<ClientMatrixRow>('integrationDocs.clients.matrix.rows').map((row) => ({
  ...row,
  baseUrl: row.baseUrl === 'gatewayOrigin' ? gatewayOrigin.value : openAiSdkBaseUrl.value
})))

const clientGuides = computed<ClientGuide[]>(() => [
  {
    id: 'client-codex-desktop',
    title: t('integrationDocs.clients.codexDesktop.title'),
    badge: t('integrationDocs.clients.codexDesktop.badge'),
    description: t('integrationDocs.clients.codexDesktop.description'),
    baseUrlLabel: t('integrationDocs.clients.baseUrlLabel'),
    baseUrl: gatewayOrigin.value,
    baseUrlNote: t('integrationDocs.clients.codexDesktop.baseUrlNote'),
    osRows: localizedRows('integrationDocs.clients.codexDesktop.osRows'),
    configBlocks: [
      {
        title: t('integrationDocs.clients.codexDesktop.blocks.configTitle'),
        description: t('integrationDocs.clients.codexDesktop.blocks.configDescription'),
        code: `baseurl = "${gatewayOrigin.value}"`
      }
    ]
  },
  {
    id: 'client-codex-cli',
    title: t('integrationDocs.clients.codexCli.title'),
    badge: t('integrationDocs.clients.codexCli.badge'),
    description: t('integrationDocs.clients.codexCli.description'),
    baseUrlLabel: t('integrationDocs.clients.baseUrlLabel'),
    baseUrl: gatewayOrigin.value,
    baseUrlNote: t('integrationDocs.clients.codexCli.baseUrlNote'),
    osRows: localizedRows('integrationDocs.clients.codexCli.osRows'),
    configBlocks: [
      {
        title: t('integrationDocs.clients.codexCli.blocks.configTitle'),
        description: t('integrationDocs.clients.codexCli.blocks.configDescription'),
        code: `model_provider = "OpenAI"
model = "gpt-5.5"
review_model = "gpt-5.4"
model_reasoning_effort = "medium"
disable_response_storage = true
network_access = "enabled"
windows_wsl_setup_acknowledged = true
model_context_window = 1000000
model_auto_compact_token_limit = 900000

[model_providers.OpenAI]
name = "OpenAI"
base_url = "${gatewayOrigin.value}"
wire_api = "responses"
requires_openai_auth = true`
      },
      {
        title: t('integrationDocs.clients.codexCli.blocks.envTitle'),
        description: t('integrationDocs.clients.codexCli.blocks.envDescription'),
        code: `{
  "OPENAI_API_KEY": "sk-your-api-key"
}`
      }
    ]
  },
  {
    id: 'client-claude-code',
    title: t('integrationDocs.clients.claudeCode.title'),
    badge: t('integrationDocs.clients.claudeCode.badge'),
    description: t('integrationDocs.clients.claudeCode.description'),
    baseUrlLabel: t('integrationDocs.clients.baseUrlLabel'),
    baseUrl: gatewayOrigin.value,
    baseUrlNote: t('integrationDocs.clients.claudeCode.baseUrlNote'),
    osRows: localizedRows('integrationDocs.clients.claudeCode.osRows'),
    configBlocks: [
      {
        title: t('integrationDocs.clients.claudeCode.blocks.settingsTitle'),
        description: t('integrationDocs.clients.claudeCode.blocks.settingsDescription'),
        code: `{
  "enabledPlugins": {
    "commit-commands@claude-plugins-official": true,
    "context7@claude-plugins-official": true,
    "frontend-design@claude-plugins-official": true,
    "playwright@claude-plugins-official": true,
    "pyright-lsp@claude-plugins-official": true,
    "superpowers@claude-plugins-official": true
  },
  "env": {
    "ANTHROPIC_BASE_URL": "${gatewayOrigin.value}",
    "ANTHROPIC_AUTH_TOKEN": "sk-your-api-key",
    "ANTHROPIC_DEFAULT_HAIKU_MODEL": "claude-opus-4-8",
    "ANTHROPIC_DEFAULT_OPUS_MODEL": "claude-opus-4-8",
    "ANTHROPIC_DEFAULT_SONNET_MODEL": "claude-opus-4-8",
    "ANTHROPIC_MODEL": "claude-opus-4-8",
    "API_TIMEOUT_MS": "3000000",
    "CLAUDE_CODE_ATTRIBUTION_HEADER": "0",
    "CLAUDE_CODE_DISABLE_NONESSENTIAL_TRAFFIC": "1"
  },
  "includeCoAuthoredBy": false
}`
      },
      {
        title: t('integrationDocs.clients.claudeCode.blocks.envTitle'),
        description: t('integrationDocs.clients.claudeCode.blocks.envDescription'),
        code: `# Windows PowerShell
[Environment]::SetEnvironmentVariable("ANTHROPIC_BASE_URL", "${gatewayOrigin.value}", "User")
[Environment]::SetEnvironmentVariable("ANTHROPIC_AUTH_TOKEN", "sk-your-api-key", "User")
[Environment]::SetEnvironmentVariable("ANTHROPIC_DEFAULT_HAIKU_MODEL", "claude-opus-4-8", "User")
[Environment]::SetEnvironmentVariable("ANTHROPIC_DEFAULT_OPUS_MODEL", "claude-opus-4-8", "User")
[Environment]::SetEnvironmentVariable("ANTHROPIC_DEFAULT_SONNET_MODEL", "claude-opus-4-8", "User")
[Environment]::SetEnvironmentVariable("ANTHROPIC_MODEL", "claude-opus-4-8", "User")
[Environment]::SetEnvironmentVariable("API_TIMEOUT_MS", "3000000", "User")
[Environment]::SetEnvironmentVariable("CLAUDE_CODE_ATTRIBUTION_HEADER", "0", "User")
[Environment]::SetEnvironmentVariable("CLAUDE_CODE_DISABLE_NONESSENTIAL_TRAFFIC", "1", "User")

# macOS / Linux
export ANTHROPIC_BASE_URL="${gatewayOrigin.value}"
export ANTHROPIC_AUTH_TOKEN="sk-your-api-key"
export ANTHROPIC_DEFAULT_HAIKU_MODEL="claude-opus-4-8"
export ANTHROPIC_DEFAULT_OPUS_MODEL="claude-opus-4-8"
export ANTHROPIC_DEFAULT_SONNET_MODEL="claude-opus-4-8"
export ANTHROPIC_MODEL="claude-opus-4-8"
export API_TIMEOUT_MS="3000000"
export CLAUDE_CODE_ATTRIBUTION_HEADER="0"
export CLAUDE_CODE_DISABLE_NONESSENTIAL_TRAFFIC="1"`
      }
    ]
  },
  {
    id: 'client-gemini-cli',
    title: t('integrationDocs.clients.geminiCli.title'),
    badge: t('integrationDocs.clients.geminiCli.badge'),
    description: t('integrationDocs.clients.geminiCli.description'),
    baseUrlLabel: t('integrationDocs.clients.baseUrlLabel'),
    baseUrl: gatewayOrigin.value,
    baseUrlNote: t('integrationDocs.clients.geminiCli.baseUrlNote'),
    osRows: localizedRows('integrationDocs.clients.geminiCli.osRows'),
    configBlocks: [
      {
        title: t('integrationDocs.clients.geminiCli.blocks.nativeTitle'),
        description: t('integrationDocs.clients.geminiCli.blocks.nativeDescription'),
        code: `# Windows PowerShell
[Environment]::SetEnvironmentVariable("GOOGLE_GEMINI_BASE_URL", "${gatewayOrigin.value}", "User")
[Environment]::SetEnvironmentVariable("GEMINI_API_KEY", "sk-your-api-key", "User")

# macOS / Linux
export GOOGLE_GEMINI_BASE_URL="${gatewayOrigin.value}"
export GEMINI_API_KEY="sk-your-api-key"`
      },
      {
        title: t('integrationDocs.clients.geminiCli.blocks.openaiTitle'),
        description: t('integrationDocs.clients.geminiCli.blocks.openaiDescription'),
        code: `# OpenAI-compatible Gemini wrapper or provider
baseURL="${openAiSdkBaseUrl.value}"
apiKey="sk-your-api-key"`
      }
    ]
  },
  {
    id: 'client-nodejs',
    title: t('integrationDocs.clients.nodejs.title'),
    badge: t('integrationDocs.clients.nodejs.badge'),
    description: t('integrationDocs.clients.nodejs.description'),
    baseUrlLabel: t('integrationDocs.clients.baseUrlLabel'),
    baseUrl: openAiSdkBaseUrl.value,
    baseUrlNote: t('integrationDocs.clients.nodejs.baseUrlNote'),
    osRows: localizedRows('integrationDocs.clients.nodejs.osRows'),
    configBlocks: [
      {
        title: t('integrationDocs.clients.nodejs.blocks.installTitle'),
        description: t('integrationDocs.clients.nodejs.blocks.installDescription'),
        code: `node -v
npm -v
npm install openai dotenv`
      },
      {
        title: t('integrationDocs.clients.nodejs.blocks.envTitle'),
        description: t('integrationDocs.clients.nodejs.blocks.envDescription'),
        code: `USE_AIFORME_API_KEY=sk-your-api-key
USE_AIFORME_BASE_URL=${openAiSdkBaseUrl.value}`
      },
      {
        title: t('integrationDocs.clients.nodejs.blocks.scriptTitle'),
        description: t('integrationDocs.clients.nodejs.blocks.scriptDescription'),
        code: `import "dotenv/config";
import OpenAI from "openai";

const client = new OpenAI({
  apiKey: process.env.USE_AIFORME_API_KEY,
  baseURL: process.env.USE_AIFORME_BASE_URL
});

const response = await client.chat.completions.create({
  model: "gpt-5.5",
  messages: [{ role: "user", content: "Reply with OK." }]
});

console.log(response.choices[0]?.message?.content);`
      }
    ]
  },
  {
    id: 'client-trae-solo',
    title: t('integrationDocs.clients.traeSolo.title'),
    badge: t('integrationDocs.clients.traeSolo.badge'),
    description: t('integrationDocs.clients.traeSolo.description'),
    baseUrlLabel: t('integrationDocs.clients.baseUrlLabel'),
    baseUrl: openAiSdkBaseUrl.value,
    baseUrlNote: t('integrationDocs.clients.traeSolo.baseUrlNote'),
    osRows: localizedRows('integrationDocs.clients.traeSolo.osRows'),
    configBlocks: [
      {
        title: t('integrationDocs.clients.traeSolo.blocks.openaiTitle'),
        description: t('integrationDocs.clients.traeSolo.blocks.openaiDescription'),
        code: `{
  "provider": "openai-compatible",
  "baseURL": "${openAiSdkBaseUrl.value}",
  "apiKey": "sk-your-api-key",
  "model": "gpt-5.5"
}`
      },
      {
        title: t('integrationDocs.clients.traeSolo.blocks.anthropicTitle'),
        description: t('integrationDocs.clients.traeSolo.blocks.anthropicDescription'),
        code: `{
  "provider": "anthropic-compatible",
  "baseURL": "${gatewayOrigin.value}",
  "apiKey": "sk-your-api-key",
  "model": "claude-opus-4-8"
}`
      }
    ]
  },
  {
    id: 'client-openclaw',
    title: t('integrationDocs.clients.openClaw.title'),
    badge: t('integrationDocs.clients.openClaw.badge'),
    description: t('integrationDocs.clients.openClaw.description'),
    baseUrlLabel: t('integrationDocs.clients.baseUrlLabel'),
    baseUrl: openAiSdkBaseUrl.value,
    baseUrlNote: t('integrationDocs.clients.openClaw.baseUrlNote'),
    osRows: localizedRows('integrationDocs.clients.openClaw.osRows'),
    configBlocks: [
      {
        title: t('integrationDocs.clients.openClaw.blocks.envTitle'),
        description: t('integrationDocs.clients.openClaw.blocks.envDescription'),
        code: `OPENAI_API_KEY=sk-your-api-key
OPENAI_BASE_URL=${openAiSdkBaseUrl.value}
OPENAI_MODEL=gpt-5.5`
      },
      {
        title: t('integrationDocs.clients.openClaw.blocks.jsonTitle'),
        description: t('integrationDocs.clients.openClaw.blocks.jsonDescription'),
        code: `{
  "baseUrl": "${openAiSdkBaseUrl.value}",
  "apiKey": "sk-your-api-key",
  "model": "gpt-5.5"
}`
      }
    ]
  },
  {
    id: 'client-hermes',
    title: t('integrationDocs.clients.hermes.title'),
    badge: t('integrationDocs.clients.hermes.badge'),
    description: t('integrationDocs.clients.hermes.description'),
    baseUrlLabel: t('integrationDocs.clients.baseUrlLabel'),
    baseUrl: openAiSdkBaseUrl.value,
    baseUrlNote: t('integrationDocs.clients.hermes.baseUrlNote'),
    osRows: localizedRows('integrationDocs.clients.hermes.osRows'),
    configBlocks: [
      {
        title: t('integrationDocs.clients.hermes.blocks.openaiTitle'),
        description: t('integrationDocs.clients.hermes.blocks.openaiDescription'),
        code: `provider: openai-compatible
base_url: ${openAiSdkBaseUrl.value}
api_key: sk-your-api-key
model: gpt-5.5`
      },
      {
        title: t('integrationDocs.clients.hermes.blocks.claudeTitle'),
        description: t('integrationDocs.clients.hermes.blocks.claudeDescription'),
        code: `provider: anthropic-compatible
base_url: ${gatewayOrigin.value}
api_key: sk-your-api-key
model: claude-opus-4-8`
      }
    ]
  },
  {
    id: 'client-api-script',
    title: t('integrationDocs.clients.apiScript.title'),
    badge: t('integrationDocs.clients.apiScript.badge'),
    description: t('integrationDocs.clients.apiScript.description'),
    baseUrlLabel: t('integrationDocs.clients.baseUrlLabel'),
    baseUrl: openAiSdkBaseUrl.value,
    baseUrlNote: t('integrationDocs.clients.apiScript.baseUrlNote'),
    osRows: localizedRows('integrationDocs.clients.apiScript.osRows'),
    configBlocks: [
      {
        title: t('integrationDocs.clients.apiScript.blocks.openaiTitle'),
        description: t('integrationDocs.clients.apiScript.blocks.openaiDescription'),
        code: `curl ${openAiSdkBaseUrl.value}/chat/completions \\
  -H "Authorization: Bearer sk-your-api-key" \\
  -H "Content-Type: application/json" \\
  -d '{"model":"gpt-5.5","messages":[{"role":"user","content":"ping"}]}'`
      },
      {
        title: t('integrationDocs.clients.apiScript.blocks.claudeTitle'),
        description: t('integrationDocs.clients.apiScript.blocks.claudeDescription'),
        code: `curl ${openAiSdkBaseUrl.value}/messages \\
  -H "x-api-key: sk-your-api-key" \\
  -H "anthropic-version: 2023-06-01" \\
  -H "Content-Type: application/json" \\
  -d '{"model":"claude-opus-4-8","max_tokens":128,"messages":[{"role":"user","content":"ping"}]}'`
      }
    ]
  }
])

const codeExamples = computed<CodeExample[]>(() => [
  {
    title: t('integrationDocs.examples.curl.title'),
    description: t('integrationDocs.examples.curl.description'),
    badge: 'curl',
    code: `curl ${openAiSdkBaseUrl.value}/chat/completions \\
  -H "Authorization: Bearer sk-your-api-key" \\
  -H "Content-Type: application/json" \\
  -d '{
    "model": "gpt-5.5",
    "messages": [
      { "role": "user", "content": "Say hello in one sentence." }
    ]
  }'`
  },
  {
    title: t('integrationDocs.examples.node.title'),
    description: t('integrationDocs.examples.node.description'),
    badge: 'Node.js',
    code: `import OpenAI from "openai";

const client = new OpenAI({
  apiKey: process.env.USE_AIFORME_API_KEY,
  baseURL: "${openAiSdkBaseUrl.value}"
});

const completion = await client.chat.completions.create({
  model: "gpt-5.5",
  messages: [{ role: "user", content: "Write a short onboarding tip." }]
});

console.log(completion.choices[0]?.message?.content);`
  },
  {
    title: t('integrationDocs.examples.python.title'),
    description: t('integrationDocs.examples.python.description'),
    badge: 'Python',
    code: `from openai import OpenAI

client = OpenAI(
    api_key="sk-your-api-key",
    base_url="${openAiSdkBaseUrl.value}",
)

response = client.chat.completions.create(
    model="gpt-5.4-mini",
    messages=[{"role": "user", "content": "Return a JSON health check."}],
)

print(response.choices[0].message.content)`
  },
  {
    title: t('integrationDocs.examples.anthropic.title'),
    description: t('integrationDocs.examples.anthropic.description'),
    badge: 'Claude',
    code: `curl ${openAiSdkBaseUrl.value}/messages \\
  -H "x-api-key: sk-your-api-key" \\
  -H "anthropic-version: 2023-06-01" \\
  -H "Content-Type: application/json" \\
  -d '{
    "model": "claude-opus-4-8",
    "max_tokens": 512,
    "messages": [
      { "role": "user", "content": "Summarize the integration steps." }
    ]
  }'`
  }
])

function localizedCards(path: string): LocalizedCard[] {
  const raw = tm(path) as unknown
  if (!Array.isArray(raw)) return []
  return raw.map((item: any) => ({
    title: localize(item.title),
    description: localize(item.description),
    bullets: Array.isArray(item.bullets) ? item.bullets.map((bullet: unknown) => localize(bullet)) : []
  }))
}

function localizedRows<T extends object>(path: string): T[] {
  const raw = tm(path) as unknown
  if (!Array.isArray(raw)) return []
  return raw.map((row: Record<string, unknown>) => {
    return Object.fromEntries(Object.entries(row).map(([key, value]) => [key, localize(value)])) as T
  })
}

function localize(value: unknown): string {
  return typeof value === 'string' ? value : rt(value as any)
}

function tr(key: string, fallbackZh: string, fallbackEn: string): string {
  if (te(key)) return t(key)
  return String(locale.value).startsWith('zh') ? fallbackZh : fallbackEn
}

function isTocItemActive(id: string): boolean {
  return activeTocId.value === id || (id === 'clients' && activeTocId.value.startsWith('client-'))
}

function tocGroupClass(id: string): string {
  const active = isTocItemActive(id)
  return [
    'rounded-xl border transition hover:border-gray-200 hover:bg-gray-50 dark:hover:border-white/10 dark:hover:bg-white/10',
    active
      ? 'border-primary-200 bg-primary-50/80 dark:border-primary-800/50 dark:bg-primary-900/20'
      : 'border-transparent'
  ].join(' ')
}

function tocLinkClass(id: string, nested = false): string {
  const active = isTocItemActive(id)
  return [
    nested ? 'min-w-0 flex-1' : 'flex items-center justify-between',
    'rounded-xl border px-3 py-2 text-sm transition hover:border-gray-200 hover:bg-gray-50 hover:text-gray-950 dark:hover:border-white/10 dark:hover:bg-white/10 dark:hover:text-white',
    active
      ? 'border-primary-200 bg-primary-50 font-semibold text-primary-800 dark:border-primary-800/50 dark:bg-primary-900/25 dark:text-primary-200'
      : 'border-transparent text-gray-600 dark:text-slate-300'
  ].join(' ')
}

function tocSubLinkClass(id: string): string {
  const active = activeTocId.value === id
  return [
    'block rounded-lg px-3 py-1.5 text-xs transition hover:bg-white hover:text-primary-700 dark:hover:bg-white/10 dark:hover:text-primary-300',
    active
      ? 'bg-white font-semibold text-primary-700 shadow-sm dark:bg-white/10 dark:text-primary-200'
      : 'text-gray-500 dark:text-slate-400'
  ].join(' ')
}

function toggleTheme() {
  isDark.value = !isDark.value
  document.documentElement.classList.toggle('dark', isDark.value)
  localStorage.setItem('theme', isDark.value ? 'dark' : 'light')
}

async function copyConfig(code: string, key: string) {
  await navigator.clipboard.writeText(code)
  copiedKey.value = key
  window.setTimeout(() => {
    if (copiedKey.value === key) copiedKey.value = ''
  }, 1600)
}

function initTheme() {
  const savedTheme = localStorage.getItem('theme')
  const prefersDark = window.matchMedia('(prefers-color-scheme: dark)').matches
  isDark.value = savedTheme === 'dark' || (!savedTheme && prefersDark)
  document.documentElement.classList.toggle('dark', isDark.value)
}

const scrollTrackedIds = computed(() => [
  ...tocItems.value.flatMap((item) =>
    item.id === 'clients'
      ? [item.id, ...clientGuides.value.map((guide) => guide.id)]
      : [item.id]
  )
])

let scrollRaf = 0

function updateDocsScrollState() {
  if (typeof window === 'undefined') return

  const doc = document.documentElement
  const maxScroll = Math.max(1, doc.scrollHeight - window.innerHeight)
  scrollProgressPercent.value = Math.min(100, Math.max(0, Math.round((window.scrollY / maxScroll) * 100)))

  const offset = 156
  let current = scrollTrackedIds.value[0] || 'quick-start'
  for (const id of scrollTrackedIds.value) {
    const section = document.getElementById(id)
    if (!section) continue
    if (section.getBoundingClientRect().top <= offset) {
      current = id
    } else {
      break
    }
  }
  activeTocId.value = current
  updateTocPanelPosition()
}

function updateTocPanelPosition() {
  const column = tocColumnRef.value
  const panel = tocPanelRef.value
  if (!column || !panel || window.innerWidth < 768) {
    tocPanelStyle.value = {}
    return
  }

  const top = window.innerWidth >= 1280 ? 128 : 112
  const columnRect = column.getBoundingClientRect()
  const columnTop = columnRect.top + window.scrollY
  const columnBottom = columnTop + column.offsetHeight
  const panelHeight = panel.offsetHeight
  const columnWidth = columnRect.width
  const columnLeft = columnRect.left
  const fixedTopDocumentY = window.scrollY + top

  if (fixedTopDocumentY <= columnTop) {
    tocPanelStyle.value = {}
    return
  }

  if (fixedTopDocumentY + panelHeight >= columnBottom) {
    tocPanelStyle.value = {
      position: 'absolute',
      left: '0px',
      right: '0px',
      top: `${Math.max(0, column.offsetHeight - panelHeight)}px`,
      width: `${columnWidth}px`
    }
    return
  }

  tocPanelStyle.value = {
    position: 'fixed',
    left: `${columnLeft}px`,
    top: `${top}px`,
    width: `${columnWidth}px`
  }
}

function handleDocsScroll() {
  if (scrollRaf) return
  scrollRaf = window.requestAnimationFrame(() => {
    scrollRaf = 0
    updateDocsScrollState()
  })
}

onMounted(() => {
  initTheme()
  authStore.checkAuth()
  if (!appStore.publicSettingsLoaded) {
    appStore.fetchPublicSettings()
  }
  updateDocsScrollState()
  window.addEventListener('scroll', handleDocsScroll, { passive: true })
  window.addEventListener('resize', handleDocsScroll)
})

onUnmounted(() => {
  window.removeEventListener('scroll', handleDocsScroll)
  window.removeEventListener('resize', handleDocsScroll)
  if (scrollRaf) {
    window.cancelAnimationFrame(scrollRaf)
  }
})

const DocsSection = defineComponent({
  props: {
    id: { type: String, required: true },
    title: { type: String, required: true },
    description: { type: String, required: true }
  },
  setup(props, { slots }) {
    return () =>
      h('section', {
        id: props.id,
        class: 'scroll-mt-6 rounded-[1.5rem] border border-gray-200/70 bg-white/75 p-5 shadow-sm backdrop-blur-2xl dark:border-white/10 dark:bg-white/[0.045] md:p-6'
      }, [
        h('div', { class: 'mb-5' }, [
          h('h2', { class: 'text-2xl font-semibold tracking-normal text-gray-950 dark:text-white' }, props.title),
          h('p', { class: 'mt-2 text-sm leading-6 text-gray-600 dark:text-slate-400' }, props.description)
        ]),
        slots.default?.()
      ])
  }
})
</script>
