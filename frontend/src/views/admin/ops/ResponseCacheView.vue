<template>
  <AppLayout>
    <div class="space-y-6 pb-12">
      <div class="flex flex-col gap-4 lg:flex-row lg:items-start lg:justify-between">
        <div>
          <div class="flex items-center gap-3">
            <h1 class="text-2xl font-semibold text-gray-900 dark:text-white">响应缓存</h1>
            <span
              class="inline-flex items-center rounded-md px-2.5 py-1 text-xs font-semibold"
              :class="decisionBadgeClass"
            >
              {{ decisionLabel }}
            </span>
          </div>
          <p class="mt-2 max-w-3xl text-sm text-gray-500 dark:text-dark-400">
            查看 exact cache 的 shadow 命中率、候选请求数和开启建议。当前页面只读，不会自动打开真实缓存。
          </p>
        </div>

        <div class="flex flex-wrap items-center gap-2">
          <button
            type="button"
            class="inline-flex h-10 items-center gap-2 rounded-lg border border-gray-200 bg-white px-4 text-sm font-medium text-gray-700 transition-colors hover:bg-gray-50 disabled:cursor-not-allowed disabled:opacity-60 dark:border-dark-700 dark:bg-dark-800 dark:text-dark-200 dark:hover:bg-dark-700"
            :disabled="loading"
            @click="fetchRecommendation"
          >
            <Icon name="refresh" size="sm" :class="loading ? 'animate-spin' : ''" />
            刷新
          </button>
          <button
            type="button"
            class="inline-flex h-10 items-center gap-2 rounded-lg bg-primary-600 px-4 text-sm font-medium text-white transition-colors hover:bg-primary-700 disabled:cursor-not-allowed disabled:opacity-60"
            :disabled="!configSnippet"
            @click="copyConfig"
          >
            <Icon name="copy" size="sm" />
            复制 .env
          </button>
        </div>
      </div>

      <div
        v-if="errorMessage"
        class="rounded-lg border border-red-200 bg-red-50 p-4 text-sm text-red-700 dark:border-red-900/50 dark:bg-red-900/20 dark:text-red-300"
      >
        {{ errorMessage }}
      </div>

      <div class="grid grid-cols-1 gap-4 md:grid-cols-2 xl:grid-cols-4">
        <div class="rounded-lg border border-gray-200 bg-white p-5 dark:border-dark-700 dark:bg-dark-800">
          <div class="flex items-center justify-between gap-3">
            <p class="text-sm font-medium text-gray-500 dark:text-dark-400">候选请求数</p>
            <Icon name="database" size="sm" class="text-gray-400" />
          </div>
          <p class="mt-3 text-3xl font-semibold text-gray-900 dark:text-white">{{ formatInt(totalCandidates) }}</p>
          <div class="mt-4 h-2 overflow-hidden rounded bg-gray-100 dark:bg-dark-700">
            <div class="h-full rounded bg-primary-500" :style="candidateProgressStyle"></div>
          </div>
          <p class="mt-2 text-xs text-gray-500 dark:text-dark-400">
            目标 {{ formatInt(minCandidates) }}，达到后才满足数量条件。
          </p>
        </div>

        <div class="rounded-lg border border-gray-200 bg-white p-5 dark:border-dark-700 dark:bg-dark-800">
          <div class="flex items-center justify-between gap-3">
            <p class="text-sm font-medium text-gray-500 dark:text-dark-400">Shadow 命中率</p>
            <Icon name="chart" size="sm" class="text-gray-400" />
          </div>
          <p class="mt-3 text-3xl font-semibold text-gray-900 dark:text-white">{{ formatPercent(hitRate) }}</p>
          <div class="mt-4 h-2 overflow-hidden rounded bg-gray-100 dark:bg-dark-700">
            <div class="h-full rounded bg-emerald-500" :style="hitRateProgressStyle"></div>
          </div>
          <p class="mt-2 text-xs text-gray-500 dark:text-dark-400">
            阈值 {{ formatPercent(threshold) }}，命中 {{ formatInt(shadowHits) }} 次。
          </p>
        </div>

        <div class="rounded-lg border border-gray-200 bg-white p-5 dark:border-dark-700 dark:bg-dark-800">
          <div class="flex items-center justify-between gap-3">
            <p class="text-sm font-medium text-gray-500 dark:text-dark-400">观察小时</p>
            <Icon name="clock" size="sm" class="text-gray-400" />
          </div>
          <p class="mt-3 text-3xl font-semibold text-gray-900 dark:text-white">
            {{ formatInt(observedHours) }} / {{ formatInt(minObservedHours) }}
          </p>
          <div class="mt-4 h-2 overflow-hidden rounded bg-gray-100 dark:bg-dark-700">
            <div class="h-full rounded bg-sky-500" :style="observedProgressStyle"></div>
          </div>
          <p class="mt-2 text-xs text-gray-500 dark:text-dark-400">
            统计窗口 {{ formatInt(windowHours) }} 小时。
          </p>
        </div>

        <div class="rounded-lg border border-gray-200 bg-white p-5 dark:border-dark-700 dark:bg-dark-800">
          <div class="flex items-center justify-between gap-3">
            <p class="text-sm font-medium text-gray-500 dark:text-dark-400">监控 Key</p>
            <Icon name="shield" size="sm" class="text-gray-400" />
          </div>
          <p class="mt-3 text-3xl font-semibold text-gray-900 dark:text-white">{{ formatInt(monitorCandidates) }}</p>
          <p class="mt-4 text-sm text-gray-600 dark:text-dark-300">
            命中 {{ formatInt(monitorHits) }} 次，监控 Key 建议保持 bypass。
          </p>
          <p class="mt-2 text-xs text-gray-500 dark:text-dark-400">
            用于避免探活脚本被缓存结果误导。
          </p>
        </div>
      </div>

      <div class="grid grid-cols-1 gap-4 md:grid-cols-3">
        <div class="rounded-lg border border-gray-200 bg-white p-5 dark:border-dark-700 dark:bg-dark-800">
          <div class="flex items-center justify-between gap-3">
            <p class="text-sm font-medium text-gray-500 dark:text-dark-400">唯一缓存 Key</p>
            <Icon name="key" size="sm" class="text-gray-400" />
          </div>
          <p class="mt-3 text-3xl font-semibold text-gray-900 dark:text-white">{{ formatInt(uniqueKeys) }}</p>
          <p class="mt-2 text-xs text-gray-500 dark:text-dark-400">
            目标 {{ formatInt(minUniqueKeys) }}，用于判断命中是否足够分散。
          </p>
        </div>

        <div class="rounded-lg border border-gray-200 bg-white p-5 dark:border-dark-700 dark:bg-dark-800">
          <div class="flex items-center justify-between gap-3">
            <p class="text-sm font-medium text-gray-500 dark:text-dark-400">Top1 命中贡献</p>
            <Icon name="chartBar" size="sm" class="text-gray-400" />
          </div>
          <p class="mt-3 text-3xl font-semibold" :class="top1TooHigh ? 'text-amber-600 dark:text-amber-300' : 'text-gray-900 dark:text-white'">
            {{ formatPercent(top1HitShare) }}
          </p>
          <p class="mt-2 text-xs text-gray-500 dark:text-dark-400">
            建议不超过 {{ formatPercent(top1MaxHitShare) }}。
          </p>
        </div>

        <div class="rounded-lg border border-gray-200 bg-white p-5 dark:border-dark-700 dark:bg-dark-800">
          <div class="flex items-center justify-between gap-3">
            <p class="text-sm font-medium text-gray-500 dark:text-dark-400">Top5 命中贡献</p>
            <Icon name="chartBar" size="sm" class="text-gray-400" />
          </div>
          <p class="mt-3 text-3xl font-semibold" :class="top5TooHigh ? 'text-amber-600 dark:text-amber-300' : 'text-gray-900 dark:text-white'">
            {{ formatPercent(top5HitShare) }}
          </p>
          <p class="mt-2 text-xs text-gray-500 dark:text-dark-400">
            建议不超过 {{ formatPercent(top5MaxHitShare) }}。
          </p>
        </div>
      </div>

      <div class="grid grid-cols-1 gap-6 xl:grid-cols-3">
        <section class="rounded-lg border border-gray-200 bg-white p-5 dark:border-dark-700 dark:bg-dark-800 xl:col-span-1">
          <div class="flex items-center justify-between gap-3">
            <div>
              <h2 class="text-base font-semibold text-gray-900 dark:text-white">推荐条件</h2>
              <p class="mt-1 text-sm text-gray-500 dark:text-dark-400">临时调整阈值后刷新，可验证推荐逻辑。</p>
            </div>
          </div>

          <div class="mt-5 grid grid-cols-1 gap-4 sm:grid-cols-2 xl:grid-cols-1">
            <label class="block">
              <span class="text-sm font-medium text-gray-700 dark:text-dark-200">统计窗口小时</span>
              <input v-model.number="filters.windowHours" type="number" min="1" max="168" class="mt-1 h-10 w-full rounded-lg border border-gray-200 bg-white px-3 text-sm text-gray-900 outline-none focus:border-primary-500 focus:ring-2 focus:ring-primary-500/20 dark:border-dark-700 dark:bg-dark-900 dark:text-white" />
            </label>
            <label class="block">
              <span class="text-sm font-medium text-gray-700 dark:text-dark-200">最小候选数</span>
              <input v-model.number="filters.minCandidates" type="number" min="1" class="mt-1 h-10 w-full rounded-lg border border-gray-200 bg-white px-3 text-sm text-gray-900 outline-none focus:border-primary-500 focus:ring-2 focus:ring-primary-500/20 dark:border-dark-700 dark:bg-dark-900 dark:text-white" />
            </label>
            <label class="block">
              <span class="text-sm font-medium text-gray-700 dark:text-dark-200">命中率阈值 %</span>
              <input v-model.number="filters.hitRateThreshold" type="number" min="0" max="100" step="0.1" class="mt-1 h-10 w-full rounded-lg border border-gray-200 bg-white px-3 text-sm text-gray-900 outline-none focus:border-primary-500 focus:ring-2 focus:ring-primary-500/20 dark:border-dark-700 dark:bg-dark-900 dark:text-white" />
            </label>
            <label class="block">
              <span class="text-sm font-medium text-gray-700 dark:text-dark-200">最小观察小时</span>
              <input v-model.number="filters.minObservedHours" type="number" min="1" max="168" class="mt-1 h-10 w-full rounded-lg border border-gray-200 bg-white px-3 text-sm text-gray-900 outline-none focus:border-primary-500 focus:ring-2 focus:ring-primary-500/20 dark:border-dark-700 dark:bg-dark-900 dark:text-white" />
            </label>
            <label class="block">
              <span class="text-sm font-medium text-gray-700 dark:text-dark-200">流量尖刺倍数</span>
              <input v-model.number="filters.maxSpikeRatio" type="number" min="0" step="0.1" class="mt-1 h-10 w-full rounded-lg border border-gray-200 bg-white px-3 text-sm text-gray-900 outline-none focus:border-primary-500 focus:ring-2 focus:ring-primary-500/20 dark:border-dark-700 dark:bg-dark-900 dark:text-white" />
            </label>
            <label class="block">
              <span class="text-sm font-medium text-gray-700 dark:text-dark-200">最小唯一 Key 数</span>
              <input v-model.number="filters.minUniqueKeys" type="number" min="0" class="mt-1 h-10 w-full rounded-lg border border-gray-200 bg-white px-3 text-sm text-gray-900 outline-none focus:border-primary-500 focus:ring-2 focus:ring-primary-500/20 dark:border-dark-700 dark:bg-dark-900 dark:text-white" />
            </label>
            <label class="block">
              <span class="text-sm font-medium text-gray-700 dark:text-dark-200">Top1 最大贡献 %</span>
              <input v-model.number="filters.top1MaxHitShare" type="number" min="0" max="100" step="0.1" class="mt-1 h-10 w-full rounded-lg border border-gray-200 bg-white px-3 text-sm text-gray-900 outline-none focus:border-primary-500 focus:ring-2 focus:ring-primary-500/20 dark:border-dark-700 dark:bg-dark-900 dark:text-white" />
            </label>
            <label class="block">
              <span class="text-sm font-medium text-gray-700 dark:text-dark-200">Top5 最大贡献 %</span>
              <input v-model.number="filters.top5MaxHitShare" type="number" min="0" max="100" step="0.1" class="mt-1 h-10 w-full rounded-lg border border-gray-200 bg-white px-3 text-sm text-gray-900 outline-none focus:border-primary-500 focus:ring-2 focus:ring-primary-500/20 dark:border-dark-700 dark:bg-dark-900 dark:text-white" />
            </label>
          </div>

          <button
            type="button"
            class="mt-5 inline-flex h-10 w-full items-center justify-center gap-2 rounded-lg border border-gray-200 bg-white px-4 text-sm font-medium text-gray-700 transition-colors hover:bg-gray-50 disabled:cursor-not-allowed disabled:opacity-60 dark:border-dark-700 dark:bg-dark-800 dark:text-dark-200 dark:hover:bg-dark-700"
            :disabled="loading"
            @click="fetchRecommendation"
          >
            <Icon name="refresh" size="sm" />
            按当前阈值重新计算建议
          </button>

          <div class="mt-5 rounded-lg bg-gray-50 p-4 dark:bg-dark-900">
            <p class="text-sm font-medium text-gray-700 dark:text-dark-200">判断结果</p>
            <p class="mt-2 text-sm text-gray-600 dark:text-dark-300">{{ decisionHelp }}</p>
            <ul v-if="reasonLabels.length" class="mt-3 space-y-2">
              <li v-for="reason in reasonLabels" :key="reason" class="flex gap-2 text-sm text-gray-600 dark:text-dark-300">
                <Icon name="infoCircle" size="sm" class="mt-0.5 flex-shrink-0 text-amber-500" />
                <span>{{ reason }}</span>
              </li>
            </ul>
          </div>
        </section>

        <section class="rounded-lg border border-gray-200 bg-white p-5 dark:border-dark-700 dark:bg-dark-800 xl:col-span-2">
          <div class="flex flex-col gap-3 sm:flex-row sm:items-center sm:justify-between">
            <div>
              <h2 class="text-base font-semibold text-gray-900 dark:text-white">上线建议配置</h2>
              <p class="mt-1 text-sm text-gray-500 dark:text-dark-400">按左侧当前阈值生成建议值，复制时仍会生成生产 `.env` 可用配置。</p>
            </div>
            <button
              type="button"
              class="inline-flex h-9 items-center justify-center gap-2 rounded-lg border border-gray-200 bg-white px-3 text-sm font-medium text-gray-700 transition-colors hover:bg-gray-50 disabled:cursor-not-allowed disabled:opacity-60 dark:border-dark-700 dark:bg-dark-800 dark:text-dark-200 dark:hover:bg-dark-700"
              :disabled="!configSnippet"
              @click="copyConfig"
            >
              <Icon name="copy" size="sm" />
              复制 .env
            </button>
          </div>
          <div class="mt-4 grid grid-cols-1 gap-4 md:grid-cols-2">
            <label v-for="field in configFields" :key="field.key" class="block">
              <span class="text-sm font-medium text-gray-700 dark:text-dark-200">{{ field.label }}</span>
              <input
                :value="field.value"
                readonly
                class="mt-1 h-10 w-full rounded-lg border border-gray-200 bg-white px-3 text-sm text-gray-900 outline-none dark:border-dark-700 dark:bg-dark-800 dark:text-white"
              />
              <p v-if="field.help" class="mt-1 text-xs text-gray-500 dark:text-dark-400">{{ field.help }}</p>
            </label>
          </div>
          <p class="mt-4 text-xs text-gray-500 dark:text-dark-400">生成时间：{{ generatedAtLabel }}</p>
        </section>
      </div>

      <section class="rounded-lg border border-gray-200 bg-white dark:border-dark-700 dark:bg-dark-800">
        <div class="flex flex-col gap-2 border-b border-gray-200 p-5 dark:border-dark-700 sm:flex-row sm:items-start sm:justify-between">
          <div>
            <h2 class="text-base font-semibold text-gray-900 dark:text-white">脱敏命中明细 Top</h2>
            <p class="mt-1 text-sm text-gray-500 dark:text-dark-400">
              按缓存 key 聚合展示，不包含用户原始问题、完整请求体或模型回答。
            </p>
          </div>
          <div class="flex flex-wrap items-center gap-2">
            <select v-model="keyStatsFilters.sort" class="h-9 rounded-lg border border-gray-200 bg-white px-3 text-sm text-gray-700 outline-none dark:border-dark-700 dark:bg-dark-800 dark:text-dark-200" @change="fetchKeyStats">
              <option value="hit_count">按命中数</option>
              <option value="total_count">按候选数</option>
              <option value="hit_rate">按命中率</option>
              <option value="last_seen_at">按最近出现</option>
            </select>
            <select v-model="keyStatsFilters.monitor" class="h-9 rounded-lg border border-gray-200 bg-white px-3 text-sm text-gray-700 outline-none dark:border-dark-700 dark:bg-dark-800 dark:text-dark-200" @change="fetchKeyStats">
              <option value="no">排除监控 Key</option>
              <option value="all">全部</option>
              <option value="yes">仅监控 Key</option>
            </select>
          </div>
        </div>
        <div v-if="concentrationDetected" class="mx-5 mt-5 rounded-lg border border-amber-200 bg-amber-50 p-4 text-sm text-amber-800 dark:border-amber-900/50 dark:bg-amber-900/20 dark:text-amber-200">
          命中贡献过于集中，可能是少数重复请求或探活脚本撑高了整体命中率，建议继续 shadow 观察，不建议直接全局开启缓存。
        </div>
        <div class="overflow-x-auto">
          <table class="min-w-full divide-y divide-gray-200 dark:divide-dark-700">
            <thead class="bg-gray-50 dark:bg-dark-900">
              <tr>
                <th class="whitespace-nowrap px-5 py-3 text-left text-xs font-semibold uppercase tracking-wide text-gray-500 dark:text-dark-400">缓存 Key</th>
                <th class="whitespace-nowrap px-5 py-3 text-left text-xs font-semibold uppercase tracking-wide text-gray-500 dark:text-dark-400">模型</th>
                <th class="whitespace-nowrap px-5 py-3 text-right text-xs font-semibold uppercase tracking-wide text-gray-500 dark:text-dark-400">API Key</th>
                <th class="whitespace-nowrap px-5 py-3 text-right text-xs font-semibold uppercase tracking-wide text-gray-500 dark:text-dark-400">分组</th>
                <th class="whitespace-nowrap px-5 py-3 text-right text-xs font-semibold uppercase tracking-wide text-gray-500 dark:text-dark-400">候选</th>
                <th class="whitespace-nowrap px-5 py-3 text-right text-xs font-semibold uppercase tracking-wide text-gray-500 dark:text-dark-400">命中</th>
                <th class="whitespace-nowrap px-5 py-3 text-right text-xs font-semibold uppercase tracking-wide text-gray-500 dark:text-dark-400">命中率</th>
                <th class="whitespace-nowrap px-5 py-3 text-right text-xs font-semibold uppercase tracking-wide text-gray-500 dark:text-dark-400">贡献</th>
                <th class="whitespace-nowrap px-5 py-3 text-left text-xs font-semibold uppercase tracking-wide text-gray-500 dark:text-dark-400">最近出现</th>
                <th class="whitespace-nowrap px-5 py-3 text-left text-xs font-semibold uppercase tracking-wide text-gray-500 dark:text-dark-400">类型</th>
              </tr>
            </thead>
            <tbody class="divide-y divide-gray-100 dark:divide-dark-700">
              <tr v-if="keyStatsLoading && !keyStats">
                <td colspan="10" class="px-5 py-10 text-center text-sm text-gray-500 dark:text-dark-400">加载中...</td>
              </tr>
              <tr v-else-if="!keyStatsItems.length">
                <td colspan="10" class="px-5 py-10 text-center text-sm text-gray-500 dark:text-dark-400">暂无脱敏 key 明细。产生候选请求后会显示。</td>
              </tr>
              <tr v-for="item in keyStatsItems" v-else :key="item.cache_key_hash" class="hover:bg-gray-50 dark:hover:bg-dark-700/50">
                <td class="whitespace-nowrap px-5 py-3 font-mono text-sm text-gray-700 dark:text-dark-200">{{ item.cache_key_hash }}</td>
                <td class="whitespace-nowrap px-5 py-3 text-sm text-gray-700 dark:text-dark-200">{{ item.model || '-' }}</td>
                <td class="whitespace-nowrap px-5 py-3 text-right text-sm text-gray-700 dark:text-dark-200">{{ item.api_key_id || '-' }}</td>
                <td class="whitespace-nowrap px-5 py-3 text-right text-sm text-gray-700 dark:text-dark-200">{{ item.group_id ?? '-' }}</td>
                <td class="whitespace-nowrap px-5 py-3 text-right text-sm text-gray-700 dark:text-dark-200">{{ formatInt(item.total_count) }}</td>
                <td class="whitespace-nowrap px-5 py-3 text-right text-sm text-gray-700 dark:text-dark-200">{{ formatInt(item.hit_count) }}</td>
                <td class="whitespace-nowrap px-5 py-3 text-right text-sm text-gray-700 dark:text-dark-200">{{ formatPercent(item.hit_rate) }}</td>
                <td class="whitespace-nowrap px-5 py-3 text-right text-sm text-gray-700 dark:text-dark-200">{{ formatPercent(item.hit_share) }}</td>
                <td class="whitespace-nowrap px-5 py-3 text-sm text-gray-700 dark:text-dark-200">{{ formatDate(item.last_seen_at) }}</td>
                <td class="whitespace-nowrap px-5 py-3">
                  <span class="inline-flex rounded-md px-2 py-1 text-xs font-semibold" :class="item.monitor ? 'bg-amber-100 text-amber-700 dark:bg-amber-900/30 dark:text-amber-300' : 'bg-gray-100 text-gray-700 dark:bg-dark-700 dark:text-dark-200'">
                    {{ item.monitor ? '监控' : '业务' }}
                  </span>
                </td>
              </tr>
            </tbody>
          </table>
        </div>
      </section>

      <section class="rounded-lg border border-gray-200 bg-white dark:border-dark-700 dark:bg-dark-800">
        <div class="flex flex-col gap-2 border-b border-gray-200 p-5 dark:border-dark-700 sm:flex-row sm:items-center sm:justify-between">
          <div>
            <h2 class="text-base font-semibold text-gray-900 dark:text-white">小时明细</h2>
            <p class="mt-1 text-sm text-gray-500 dark:text-dark-400">按小时查看 shadow 候选、命中和监控 Key 数据。</p>
          </div>
          <span class="text-sm text-gray-500 dark:text-dark-400">
            {{ visibleHours.length }} 条
          </span>
        </div>

        <div class="overflow-x-auto">
          <table class="min-w-full divide-y divide-gray-200 dark:divide-dark-700">
            <thead class="bg-gray-50 dark:bg-dark-900">
              <tr>
                <th class="whitespace-nowrap px-5 py-3 text-left text-xs font-semibold uppercase tracking-wide text-gray-500 dark:text-dark-400">时间</th>
                <th class="whitespace-nowrap px-5 py-3 text-right text-xs font-semibold uppercase tracking-wide text-gray-500 dark:text-dark-400">候选</th>
                <th class="whitespace-nowrap px-5 py-3 text-right text-xs font-semibold uppercase tracking-wide text-gray-500 dark:text-dark-400">命中</th>
                <th class="whitespace-nowrap px-5 py-3 text-left text-xs font-semibold uppercase tracking-wide text-gray-500 dark:text-dark-400">命中率</th>
                <th class="whitespace-nowrap px-5 py-3 text-right text-xs font-semibold uppercase tracking-wide text-gray-500 dark:text-dark-400">监控候选</th>
                <th class="whitespace-nowrap px-5 py-3 text-right text-xs font-semibold uppercase tracking-wide text-gray-500 dark:text-dark-400">监控命中</th>
              </tr>
            </thead>
            <tbody class="divide-y divide-gray-100 dark:divide-dark-700">
              <tr v-if="loading && !recommendation">
                <td colspan="6" class="px-5 py-10 text-center text-sm text-gray-500 dark:text-dark-400">加载中...</td>
              </tr>
              <tr v-else-if="!visibleHours.length">
                <td colspan="6" class="px-5 py-10 text-center text-sm text-gray-500 dark:text-dark-400">暂无 shadow 统计数据。</td>
              </tr>
              <tr v-for="row in visibleHours" v-else :key="row.hour" class="hover:bg-gray-50 dark:hover:bg-dark-700/50">
                <td class="whitespace-nowrap px-5 py-3 text-sm text-gray-700 dark:text-dark-200">{{ formatDate(row.hour) }}</td>
                <td class="whitespace-nowrap px-5 py-3 text-right text-sm text-gray-700 dark:text-dark-200">{{ formatInt(row.total) }}</td>
                <td class="whitespace-nowrap px-5 py-3 text-right text-sm text-gray-700 dark:text-dark-200">{{ formatInt(row.hit) }}</td>
                <td class="min-w-[180px] px-5 py-3">
                  <div class="flex items-center gap-3">
                    <div class="h-2 flex-1 overflow-hidden rounded bg-gray-100 dark:bg-dark-700">
                      <div class="h-full rounded bg-emerald-500" :style="rateBarStyle(row.hit_rate)"></div>
                    </div>
                    <span class="w-14 text-right text-sm text-gray-700 dark:text-dark-200">{{ formatPercent(row.hit_rate) }}</span>
                  </div>
                </td>
                <td class="whitespace-nowrap px-5 py-3 text-right text-sm text-gray-700 dark:text-dark-200">{{ formatInt(row.monitor_total) }}</td>
                <td class="whitespace-nowrap px-5 py-3 text-right text-sm text-gray-700 dark:text-dark-200">{{ formatInt(row.monitor_hit) }}</td>
              </tr>
            </tbody>
          </table>
        </div>
      </section>
    </div>
  </AppLayout>
</template>

<script setup lang="ts">
import { computed, onMounted, reactive, ref } from 'vue'
import axios from 'axios'
import AppLayout from '@/components/layout/AppLayout.vue'
import Icon from '@/components/icons/Icon.vue'
import {
  opsAPI,
  type OpsResponseCacheKeyStatsQuery,
  type OpsResponseCacheKeyStatsResponse,
  type OpsResponseCacheRecommendation
} from '@/api/admin/ops'
import { useClipboard } from '@/composables/useClipboard'
import { formatDateTime } from './utils/opsFormatters'

const loading = ref(false)
const keyStatsLoading = ref(false)
const errorMessage = ref('')
const recommendation = ref<OpsResponseCacheRecommendation | null>(null)
const keyStats = ref<OpsResponseCacheKeyStatsResponse | null>(null)
const abortController = ref<AbortController | null>(null)
const keyStatsAbortController = ref<AbortController | null>(null)
const { copyToClipboard } = useClipboard()

const filters = reactive({
  windowHours: 72,
  minCandidates: 150,
  hitRateThreshold: 20,
  minObservedHours: 24,
  maxSpikeRatio: 5,
  minUniqueKeys: 20,
  top1MaxHitShare: 50,
  top5MaxHitShare: 80
})

type KeyStatsSort = NonNullable<OpsResponseCacheKeyStatsQuery['sort']>
type KeyStatsMonitor = NonNullable<OpsResponseCacheKeyStatsQuery['monitor']>

const keyStatsFilters = reactive<{
  sort: KeyStatsSort
  monitor: KeyStatsMonitor
  limit: number
}>({
  sort: 'hit_count',
  monitor: 'no',
  limit: 50
})

const activeThresholds = computed(() => ({
  windowHours: normalizePositiveInt(filters.windowHours),
  minCandidates: normalizePositiveInt(filters.minCandidates),
  hitRateThreshold: normalizePercentInput(filters.hitRateThreshold),
  minObservedHours: normalizePositiveInt(filters.minObservedHours),
  maxSpikeRatio: normalizeNumber(filters.maxSpikeRatio),
  minUniqueKeys: normalizeNonNegativeInt(filters.minUniqueKeys),
  top1MaxHitShare: normalizePercentInput(filters.top1MaxHitShare),
  top5MaxHitShare: normalizePercentInput(filters.top5MaxHitShare)
}))

const totalCandidates = computed(() => recommendation.value?.total_candidates ?? 0)
const shadowHits = computed(() => recommendation.value?.shadow_hits ?? 0)
const monitorCandidates = computed(() => recommendation.value?.monitor_candidates ?? 0)
const monitorHits = computed(() => recommendation.value?.monitor_hits ?? 0)
const hitRate = computed(() => recommendation.value?.hit_rate ?? 0)
const threshold = computed(() => activeThresholds.value.hitRateThreshold)
const observedHours = computed(() => recommendation.value?.observed_hours ?? 0)
const minObservedHours = computed(() => activeThresholds.value.minObservedHours)
const minCandidates = computed(() => activeThresholds.value.minCandidates)
const windowHours = computed(() => activeThresholds.value.windowHours)
const uniqueKeys = computed(() => recommendation.value?.unique_keys ?? keyStats.value?.summary?.unique_keys ?? 0)
const minUniqueKeys = computed(() => activeThresholds.value.minUniqueKeys)
const top1HitShare = computed(() => recommendation.value?.top1_hit_share ?? keyStats.value?.summary?.top1_hit_share ?? 0)
const top5HitShare = computed(() => recommendation.value?.top5_hit_share ?? keyStats.value?.summary?.top5_hit_share ?? 0)
const top1MaxHitShare = computed(() => activeThresholds.value.top1MaxHitShare)
const top5MaxHitShare = computed(() => activeThresholds.value.top5MaxHitShare)
const concentrationDetected = computed(() => recommendation.value?.concentration_detected ?? keyStats.value?.summary?.concentration_detected ?? false)
const top1TooHigh = computed(() => top1MaxHitShare.value > 0 && top1HitShare.value > top1MaxHitShare.value)
const top5TooHigh = computed(() => top5MaxHitShare.value > 0 && top5HitShare.value > top5MaxHitShare.value)
const keyStatsItems = computed(() => keyStats.value?.items ?? [])

const visibleHours = computed(() => {
  const rows = recommendation.value?.hours ?? []
  return rows.filter((row) => row.total > 0 || row.hit > 0 || row.monitor_total > 0 || row.monitor_hit > 0)
})

const decisionLabel = computed(() => decisionText(recommendation.value?.decision || 'loading'))

const decisionHelp = computed(() => {
  const rec = recommendation.value
  if (!rec) return '正在读取后端统计接口。'
  if (rec.decision === 'already_enabled') return '真实 exact cache 已开启，继续关注命中率、监控 Key 和异常趋势。'
  if (rec.recommended) return '当前 shadow 数据满足阈值，可以考虑对选定 API Key 或分组开启真实 exact cache。'
  return '当前数据还不建议打开真实缓存，建议继续 shadow 观察或先降低本地阈值验证链路。'
})

const decisionBadgeClass = computed(() => {
  const rec = recommendation.value
  if (!rec) return 'bg-gray-100 text-gray-700 dark:bg-dark-700 dark:text-dark-200'
  if (rec.decision === 'already_enabled') return 'bg-blue-100 text-blue-700 dark:bg-blue-900/30 dark:text-blue-300'
  if (rec.recommended) return 'bg-emerald-100 text-emerald-700 dark:bg-emerald-900/30 dark:text-emerald-300'
  if (rec.reasons?.includes('redis_unavailable') || rec.reasons?.includes('redis_error') || rec.reasons?.includes('response_cache_not_configured')) {
    return 'bg-red-100 text-red-700 dark:bg-red-900/30 dark:text-red-300'
  }
  return 'bg-amber-100 text-amber-700 dark:bg-amber-900/30 dark:text-amber-300'
})

const reasonLabels = computed(() => {
  return (recommendation.value?.reasons ?? []).map(reasonText)
})

const generatedAtLabel = computed(() => {
  const value = recommendation.value?.generated_at
  return value ? formatDate(value) : '-'
})

const candidateProgressStyle = computed(() => ({ width: progressWidth(totalCandidates.value, minCandidates.value) }))
const hitRateProgressStyle = computed(() => ({ width: progressWidth(hitRate.value, threshold.value) }))
const observedProgressStyle = computed(() => ({ width: progressWidth(observedHours.value, minObservedHours.value) }))

const configFields = computed(() => {
  const rec = recommendation.value
  const thresholds = activeThresholds.value
  const realCacheEnabled =
    rec?.decision === 'already_enabled' ? 'true' : rec?.recommended ? 'true（建议灰度开启）' : 'false（继续 shadow）'
  return [
    {
      key: 'enabled',
      label: '真实缓存开关',
      value: realCacheEnabled,
      help: '是否用缓存结果直接响应用户。未验证前建议保持关闭。'
    },
    {
      key: 'shadowEnabled',
      label: 'Shadow 统计',
      value: 'true',
      help: '只统计如果开启缓存会不会命中，不影响用户真实响应。'
    },
    {
      key: 'recommendationEnabled',
      label: '开启推荐判断',
      value: 'true',
      help: '让后端计算是否建议开启 exact cache。'
    },
    {
      key: 'windowHours',
      label: '统计窗口小时',
      value: String(thresholds.windowHours),
      help: '默认观察最近 72 小时。'
    },
    {
      key: 'minCandidates',
      label: '最小候选数',
      value: String(thresholds.minCandidates),
      help: '样本量低于这个值时不建议开启真实缓存。'
    },
    {
      key: 'threshold',
      label: '命中率阈值',
      value: formatPercent(thresholds.hitRateThreshold),
      help: 'shadow 命中率超过该值才满足命中条件。'
    },
    {
      key: 'minObservedHours',
      label: '最小观察小时',
      value: String(thresholds.minObservedHours),
      help: '避免只看短时间偶然重复请求。'
    },
    {
      key: 'maxSpikeRatio',
      label: '流量尖刺倍数',
      value: String(thresholds.maxSpikeRatio),
      help: '用于识别异常集中流量，降低误判。'
    },
    {
      key: 'minUniqueKeys',
      label: '最小唯一 Key 数',
      value: String(thresholds.minUniqueKeys),
      help: '避免少数重复问题撑高整体命中率。'
    },
    {
      key: 'top1MaxHitShare',
      label: 'Top1 最大命中贡献',
      value: formatPercent(thresholds.top1MaxHitShare),
      help: '单个缓存 key 贡献过高时不建议直接开启。'
    },
    {
      key: 'top5MaxHitShare',
      label: 'Top5 最大命中贡献',
      value: formatPercent(thresholds.top5MaxHitShare),
      help: '前 5 个缓存 key 贡献过高时视为命中过于集中。'
    },
    {
      key: 'monitorApiKeys',
      label: '监控 Key 列表',
      value: '空',
      help: '生产上建议填探活/预警脚本使用的 API Key ID。'
    },
    {
      key: 'bypassApiKeys',
      label: '强制跳过缓存 Key',
      value: '空',
      help: '探活脚本或敏感业务 Key 可以放这里。'
    }
  ]
})

const configSnippet = computed(() => {
  const rec = recommendation.value
  const thresholds = activeThresholds.value
  if (!rec) {
    return [
      'RESPONSE_CACHE_ENABLED=false',
      'RESPONSE_CACHE_SHADOW_ENABLED=true',
      'RESPONSE_CACHE_RECOMMENDATION_ENABLED=true'
    ].join('\n')
  }

  const lines = [
    '# ResponseCache 第一阶段推荐配置',
    'RESPONSE_CACHE_SHADOW_ENABLED=true',
    'RESPONSE_CACHE_RECOMMENDATION_ENABLED=true',
    `RESPONSE_CACHE_RECOMMENDATION_WINDOW_HOURS=${thresholds.windowHours}`,
    `RESPONSE_CACHE_RECOMMENDATION_MIN_CANDIDATES=${thresholds.minCandidates}`,
    `RESPONSE_CACHE_RECOMMENDATION_HIT_RATE_THRESHOLD=${formatConfigRatio(thresholds.hitRateThreshold)}`,
    `RESPONSE_CACHE_RECOMMENDATION_MIN_OBSERVED_HOURS=${thresholds.minObservedHours}`,
    `RESPONSE_CACHE_RECOMMENDATION_MAX_SPIKE_RATIO=${thresholds.maxSpikeRatio}`,
    `RESPONSE_CACHE_RECOMMENDATION_MIN_UNIQUE_KEYS=${thresholds.minUniqueKeys}`,
    `RESPONSE_CACHE_RECOMMENDATION_TOP1_MAX_HIT_SHARE=${formatConfigRatio(thresholds.top1MaxHitShare)}`,
    `RESPONSE_CACHE_RECOMMENDATION_TOP5_MAX_HIT_SHARE=${formatConfigRatio(thresholds.top5MaxHitShare)}`,
    '',
    '# 监控/探活 Key 建议加入 bypass 或 monitor 列表',
    'RESPONSE_CACHE_MONITOR_API_KEY_ID_LIST=',
    'RESPONSE_CACHE_BYPASS_API_KEY_ID_LIST='
  ]

  if (rec.recommended) {
    lines.unshift('# 当前 shadow 数据达到开启建议，但仍建议先按 API Key 或分组灰度')
    lines.push('', '# 灰度开启时再打开真实缓存')
    lines.push('RESPONSE_CACHE_ENABLED=true')
    lines.push('RESPONSE_CACHE_ALLOWED_API_KEY_ID_LIST=')
    lines.push('RESPONSE_CACHE_ALLOWED_GROUP_ID_LIST=')
  } else if (rec.decision === 'already_enabled') {
    lines.unshift('# 当前真实 exact cache 已开启')
    lines.push('RESPONSE_CACHE_ENABLED=true')
  } else {
    lines.unshift('# 当前暂不建议开启真实缓存，继续 shadow 观察')
    lines.push('RESPONSE_CACHE_ENABLED=false')
  }

  return lines.join('\n')
})

async function fetchRecommendation() {
  abortController.value?.abort()
  const controller = new AbortController()
  abortController.value = controller
  loading.value = true
  errorMessage.value = ''
  try {
    const thresholds = activeThresholds.value
    recommendation.value = await opsAPI.getResponseCacheRecommendation(
      {
        window_hours: thresholds.windowHours,
        min_candidates: thresholds.minCandidates,
        hit_rate_threshold: thresholds.hitRateThreshold,
        min_observed_hours: thresholds.minObservedHours,
        max_spike_ratio: thresholds.maxSpikeRatio,
        min_unique_keys: thresholds.minUniqueKeys,
        top1_max_hit_share: thresholds.top1MaxHitShare,
        top5_max_hit_share: thresholds.top5MaxHitShare
      },
      { signal: controller.signal }
    )
    await fetchKeyStats()
  } catch (error: any) {
    if (axios.isCancel(error) || error?.code === 'ERR_CANCELED') return
    errorMessage.value = error?.message || '读取响应缓存统计失败'
  } finally {
    if (abortController.value === controller) {
      abortController.value = null
      loading.value = false
    }
  }
}

async function fetchKeyStats() {
  keyStatsAbortController.value?.abort()
  const controller = new AbortController()
  keyStatsAbortController.value = controller
  keyStatsLoading.value = true
  try {
    const thresholds = activeThresholds.value
    keyStats.value = await opsAPI.getResponseCacheKeyStats(
      {
        window_hours: thresholds.windowHours,
        limit: normalizePositiveInt(keyStatsFilters.limit),
        sort: keyStatsFilters.sort,
        monitor: keyStatsFilters.monitor
      },
      { signal: controller.signal }
    )
  } catch (error: any) {
    if (axios.isCancel(error) || error?.code === 'ERR_CANCELED') return
    errorMessage.value = error?.message || '读取缓存 key 明细失败'
  } finally {
    if (keyStatsAbortController.value === controller) {
      keyStatsAbortController.value = null
      keyStatsLoading.value = false
    }
  }
}

async function copyConfig() {
  await copyToClipboard(configSnippet.value, '配置片段已复制')
}

function decisionText(value: string): string {
  const labels: Record<string, string> = {
    loading: '读取中',
    not_available: '不可用',
    not_recommended: '继续观察',
    recommend_enable_exact_cache: '建议开启',
    already_enabled: '已开启'
  }
  return labels[value] || value
}

function reasonText(value: string): string {
  const labels: Record<string, string> = {
    response_cache_not_configured: '后端没有注入 ResponseCache 服务，请确认新代码和配置已经发布。',
    recommendation_disabled: '推荐统计未启用，请确认 RESPONSE_CACHE_RECOMMENDATION_ENABLED=true。',
    redis_unavailable: '第二 Redis 不可用或未配置，shadow 统计无法读取。',
    redis_error: '读取第二 Redis 统计失败，请检查容器、密码和网络。',
    insufficient_observed_hours: '观察小时数不足，还没覆盖足够长的真实流量周期。',
    insufficient_candidates: '候选请求数不足，样本量还不够。',
    low_shadow_hit_rate: 'shadow 命中率低于当前阈值。',
    hourly_hit_rate_below_threshold: '部分小时命中率低于阈值，命中不够稳定。',
    traffic_spike_detected: '检测到小时流量尖刺，建议排查是否是脚本压测或异常流量。',
    key_stats_unavailable: '缓存 key 明细暂不可用，无法判断命中是否集中。',
    insufficient_unique_keys: '唯一缓存 Key 数不足，命中样本还不够分散。',
    hit_concentration_detected: '命中贡献过于集中，可能被少数重复请求撑高。',
    exact_cache_already_enabled: '真实 exact cache 已经开启。'
  }
  return labels[value] || value
}

function formatInt(value: number | null | undefined): string {
  const n = Number(value || 0)
  return Number.isFinite(n) ? Math.round(n).toLocaleString() : '0'
}

function formatPercent(value: number | null | undefined): string {
  const n = Number(value || 0)
  if (!Number.isFinite(n)) return '0%'
  return `${(n * 100).toFixed(n > 0 && n < 0.01 ? 2 : 1)}%`
}

function formatDate(value: string): string {
  return formatDateTime(value) || '-'
}

function progressWidth(value: number, target: number): string {
  if (!Number.isFinite(value) || !Number.isFinite(target) || target <= 0) return '0%'
  return `${Math.max(0, Math.min(100, (value / target) * 100)).toFixed(1)}%`
}

function rateBarStyle(value: number) {
  return { width: `${Math.max(0, Math.min(100, Number(value || 0) * 100)).toFixed(1)}%` }
}

function normalizePositiveInt(value: number): number {
  const n = Number(value)
  return Number.isFinite(n) && n > 0 ? Math.floor(n) : 0
}

function normalizeNonNegativeInt(value: number): number {
  const n = Number(value)
  return Number.isFinite(n) && n >= 0 ? Math.floor(n) : 0
}

function normalizeNumber(value: number): number {
  const n = Number(value)
  return Number.isFinite(n) && n >= 0 ? n : 0
}

function normalizePercentInput(value: number): number {
  const n = normalizeNumber(value)
  return n > 1 ? n / 100 : n
}

function formatConfigRatio(value: number): string {
  const n = Number(value)
  if (!Number.isFinite(n)) return '0.20'
  return n.toFixed(2)
}

onMounted(fetchRecommendation)
</script>
