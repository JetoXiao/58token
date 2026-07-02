<template>
  <AppLayout>
    <div class="mx-auto max-w-6xl space-y-6">
      <div v-if="loading" class="flex items-center justify-center py-20">
        <div class="h-8 w-8 animate-spin rounded-full border-4 border-primary-500 border-t-transparent"></div>
      </div>
      <template v-else>
        <!-- Tab Switcher -->
        <div v-if="tabs.length > 1 && paymentPhase === 'select'" class="flex space-x-1 rounded-xl bg-gray-100 p-1 dark:bg-dark-800">
          <button v-for="tab in tabs" :key="tab.key"
            class="flex-1 rounded-lg px-4 py-2.5 text-sm font-medium transition-all"
            :class="activeTab === tab.key ? 'bg-white text-gray-900 shadow dark:bg-dark-700 dark:text-white' : 'text-gray-500 hover:text-gray-700 dark:text-gray-400 dark:hover:text-gray-300'"
            @click="activeTab = tab.key">{{ tab.label }}</button>
        </div>
        <!-- Payment in progress (shared by recharge and subscription) -->
        <template v-if="paymentPhase === 'paying'">
          <PaymentStatusPanel
            :order-id="paymentState.orderId"
            :qr-code="paymentState.qrCode"
            :expires-at="paymentState.expiresAt"
            :payment-type="paymentState.paymentType"
            :pay-url="paymentState.payUrl"
            :order-type="paymentState.orderType"
            :currency="paymentState.currency || selectedCurrency"
            @done="onPaymentDone"
            @success="onPaymentSuccess"
            @settled="onPaymentSettled"
          />
        </template>
        <!-- Tab content (select phase) -->
        <template v-else>
          <!-- Top-up Tab -->
          <template v-if="activeTab === 'recharge'">
            <div class="space-y-6">
              <div class="relative overflow-hidden rounded-2xl border border-gray-200 bg-white p-6 shadow-sm dark:border-dark-700 dark:bg-dark-900 sm:p-8">
                <div class="pointer-events-none absolute -right-24 -top-24 h-64 w-64 rounded-full bg-cyan-400/20 blur-3xl dark:bg-cyan-400/10"></div>
                <div class="pointer-events-none absolute -bottom-28 left-10 h-64 w-64 rounded-full bg-violet-500/15 blur-3xl dark:bg-violet-500/10"></div>
                <div class="relative grid gap-6">
                  <div>
                    <p class="inline-flex rounded-full border border-primary-200 bg-primary-50 px-3 py-1 text-xs font-semibold text-primary-700 dark:border-primary-900/50 dark:bg-primary-900/20 dark:text-primary-300">
                      {{ t('payment.rechargeHeroEyebrow') }}
                    </p>
                    <h2 class="mt-5 max-w-2xl text-3xl font-black tracking-tight text-gray-950 dark:text-white sm:text-4xl">
                      {{ t('payment.rechargeHeroTitle') }}
                    </h2>
                    <p class="mt-3 max-w-2xl text-base leading-7 text-gray-600 dark:text-gray-300">
                      {{ t('payment.rechargeHeroDescription') }}
                    </p>
                    <div class="mt-5 grid gap-3 sm:grid-cols-2">
                      <div class="rounded-2xl border border-gray-200/80 bg-white/70 p-5 shadow-sm backdrop-blur dark:border-dark-700 dark:bg-dark-800/70">
                        <p class="text-sm font-medium text-gray-500 dark:text-gray-400">{{ t('payment.currentBalance') }}</p>
                        <p class="mt-2 text-4xl font-black text-gray-950 dark:text-white">{{ formatUsdBalance(user?.balance || 0) }}</p>
                        <p class="mt-2 text-xs text-gray-500 dark:text-gray-400">{{ user?.username || t('payment.rechargeAccount') }}</p>
                      </div>
                      <div class="rounded-2xl border border-emerald-200 bg-emerald-50/70 p-4 text-sm dark:border-emerald-900/50 dark:bg-emerald-900/20">
                        <p class="font-semibold text-emerald-800 dark:text-emerald-200">{{ t('payment.usdtExchangeRateTitle') }}</p>
                        <p class="mt-1 text-2xl font-black text-emerald-700 dark:text-emerald-300">{{ usdtCnyExchangeRateLabel }}</p>
                      </div>
                    </div>
                  </div>
                </div>
              </div>

              <div class="grid gap-6 lg:grid-cols-[minmax(0,1fr)_360px]">
                <div class="space-y-6">
                  <div class="card p-5">
                    <div class="mb-4 flex flex-col gap-2 sm:flex-row sm:items-end sm:justify-between">
                      <div>
                        <p class="text-sm font-semibold text-gray-900 dark:text-white">{{ t('payment.paymentRailTitle') }}</p>
                      </div>
                      <span class="w-fit rounded-full border border-emerald-200 bg-emerald-50 px-3 py-1 text-xs font-semibold text-emerald-700 dark:border-emerald-900/60 dark:bg-emerald-900/20 dark:text-emerald-300">
                        {{ t('payment.rechargeValueTitle') }}
                      </span>
                    </div>
                    <div class="grid gap-3 md:grid-cols-2">
                      <button
                        type="button"
                        :disabled="!rmbRailAvailable"
                        :class="[
                          'rounded-xl border p-4 text-left transition-all',
                          selectedRechargeRail === 'rmb'
                            ? 'border-primary-400 bg-primary-50/80 ring-4 ring-primary-100 dark:border-primary-500 dark:bg-primary-900/20 dark:ring-primary-950'
                            : rmbRailAvailable
                              ? 'border-gray-200 bg-white/80 hover:border-primary-200 dark:border-dark-700 dark:bg-dark-800/80 dark:hover:border-primary-800'
                              : 'cursor-not-allowed border-gray-200 bg-gray-50/80 opacity-60 dark:border-dark-700 dark:bg-dark-800/80',
                        ]"
                        @click="rmbRailAvailable && (selectedRechargeRail = 'rmb')"
                      >
                        <div class="flex items-center justify-between gap-2">
                          <p class="text-sm font-bold text-gray-800 dark:text-gray-100">{{ t('payment.paymentRailRmb') }}</p>
                          <span
                            :class="[
                              'rounded-full px-2 py-0.5 text-[11px] font-semibold',
                              rmbRailAvailable
                                ? 'border border-emerald-200 bg-emerald-50 text-emerald-700 dark:border-emerald-900/60 dark:bg-emerald-900/20 dark:text-emerald-300'
                                : 'bg-gray-200 text-gray-600 dark:bg-dark-700 dark:text-gray-300',
                            ]"
                          >
                            {{ rmbRailAvailable ? t('payment.paymentRailOpen') : t('payment.paymentRailUnavailable') }}
                          </span>
                        </div>
                        <p class="mt-1 text-xs leading-5 text-gray-500 dark:text-gray-400">{{ t('payment.paymentRailRmbDesc') }}</p>
                      </button>
                      <button
                        type="button"
                        :disabled="!usdtRailAvailable"
                        :class="[
                          'rounded-xl border p-4 text-left transition-all',
                          selectedRechargeRail === 'usdt'
                            ? 'border-primary-400 bg-primary-50/80 ring-4 ring-primary-100 dark:border-primary-500 dark:bg-primary-900/20 dark:ring-primary-950'
                            : usdtRailAvailable
                              ? 'border-gray-200 bg-white/80 hover:border-primary-200 dark:border-dark-700 dark:bg-dark-800/80 dark:hover:border-primary-800'
                              : 'cursor-not-allowed border-gray-200 bg-gray-50/80 opacity-60 dark:border-dark-700 dark:bg-dark-800/80',
                        ]"
                        @click="usdtRailAvailable && (selectedRechargeRail = 'usdt')"
                      >
                        <div class="flex items-center justify-between gap-2">
                          <p class="text-sm font-bold text-gray-950 dark:text-white">{{ t('payment.paymentRailUsdt') }}</p>
                          <span
                            :class="[
                              'rounded-full px-2 py-0.5 text-[11px] font-semibold',
                              usdtRailAvailable
                                ? 'border border-emerald-200 bg-emerald-50 text-emerald-700 dark:border-emerald-900/60 dark:bg-emerald-900/20 dark:text-emerald-300'
                                : 'bg-gray-200 text-gray-600 dark:bg-dark-700 dark:text-gray-300',
                            ]"
                          >
                            {{ usdtRailAvailable ? t('payment.paymentRailOpen') : t('payment.paymentRailUnavailable') }}
                          </span>
                        </div>
                        <p class="mt-1 text-xs leading-5 text-gray-600 dark:text-gray-300">{{ t('payment.paymentRailUsdtDesc') }}</p>
                      </button>
                    </div>
                  </div>

                  <div>
                    <div class="mb-4 flex flex-col gap-2 sm:flex-row sm:items-end sm:justify-between">
                      <div>
                        <h3 class="text-xl font-bold text-gray-950 dark:text-white">{{ t('payment.packageTitle') }}</h3>
                        <p class="mt-1 text-sm text-gray-500 dark:text-gray-400">{{ t('payment.packageSubtitle') }}</p>
                      </div>
                    </div>
                    <div class="grid gap-4 sm:grid-cols-2 xl:grid-cols-4">
                      <button
                        v-for="pkg in rechargePackages"
                        :key="pkg.amount"
                        type="button"
                        :class="[
                          'group relative min-h-[220px] rounded-2xl border p-5 text-left shadow-sm transition-all hover:-translate-y-0.5 hover:shadow-lg',
                          validAmount === pkg.amount
                            ? 'border-primary-400 bg-primary-50/80 ring-4 ring-primary-100 dark:border-primary-500 dark:bg-primary-900/20 dark:ring-primary-950'
                            : 'border-gray-200 bg-white hover:border-primary-200 dark:border-dark-700 dark:bg-dark-900 dark:hover:border-primary-800',
                        ]"
                        @click="selectRechargePackage(pkg.amount)"
                      >
                        <span v-if="pkg.badgeKey" class="absolute right-4 top-4 rounded-full border border-amber-200 bg-amber-50 px-2.5 py-1 text-xs font-bold text-amber-700 dark:border-amber-900/50 dark:bg-amber-900/20 dark:text-amber-300">
                          {{ t(pkg.badgeKey) }}
                        </span>
                        <p class="text-sm font-semibold text-gray-500 dark:text-gray-400">{{ t(pkg.nameKey) }}</p>
                        <div class="mt-5">
                          <p class="text-sm text-gray-500 dark:text-gray-400">{{ t('payment.payRmb') }}</p>
                          <p class="mt-1 text-4xl font-black tracking-tight text-gray-950 dark:text-white">{{ formatRmbAmount(pkg.amount) }}</p>
                          <p v-if="selectedRechargeRail === 'usdt'" class="mt-2 text-xs font-semibold text-primary-600 dark:text-primary-300">
                            {{ t('payment.usdtEquivalent', { amount: formatUsdtAmount(usdtPaymentForAmount(pkg.amount)) }) }}
                          </p>
                          <p v-else class="mt-2 text-xs font-semibold text-primary-600 dark:text-primary-300">
                            {{ t('payment.rmbOnlinePayHint') }}
                          </p>
                        </div>
                        <div class="mt-5 rounded-xl border border-emerald-200 bg-emerald-50/80 p-3 dark:border-emerald-900/50 dark:bg-emerald-900/20">
                          <div class="flex items-start justify-between gap-3">
                            <p class="text-xs font-medium text-emerald-700 dark:text-emerald-300">{{ t('payment.receiveUsd') }}</p>
                            <span v-if="bonusForAmount(pkg.amount) > 0" class="rounded-full bg-emerald-100 px-2 py-0.5 text-[11px] font-bold text-emerald-700 dark:bg-emerald-900/40 dark:text-emerald-200">
                              {{ t('payment.rechargeBonusBadge', { amount: formatUsdBalance(bonusForAmount(pkg.amount)) }) }}
                            </span>
                          </div>
                          <p class="mt-1 text-2xl font-black text-emerald-700 dark:text-emerald-300">{{ formatUsdBalance(creditedForAmount(pkg.amount)) }}</p>
                        </div>
                      </button>
                    </div>
                  </div>

                  <div class="card p-6">
                    <div class="mb-4">
                      <h3 class="text-base font-bold text-gray-950 dark:text-white">{{ t('payment.customRechargeTitle') }}</h3>
                      <p class="mt-1 text-sm text-gray-500 dark:text-gray-400">{{ t('payment.customRechargeDescription') }}</p>
                    </div>
                    <AmountInput
                      v-model="amount"
                      :amounts="[]"
                      :min="globalMinAmount"
                      :max="globalMaxAmount"
                      prefix="¥"
                    />
                    <p v-if="amountError" class="mt-2 text-xs text-amber-600 dark:text-amber-300">{{ amountError }}</p>
                  </div>
                </div>

                <div class="space-y-4 lg:sticky lg:top-24 lg:self-start">
                  <div v-if="rechargeMethods.length >= 1" class="card p-6">
                    <PaymentMethodSelector
                      :methods="methodOptions"
                      :selected="selectedRechargeMethod"
                      @select="selectedRechargeMethod = $event"
                    />
                  </div>
                  <div v-else class="card p-6">
                    <h3 class="text-base font-bold text-gray-950 dark:text-white">{{ t('payment.paymentMethod') }}</h3>
                    <p class="mt-2 text-sm leading-6 text-gray-500 dark:text-gray-400">{{ t('payment.methodUnavailableHint') }}</p>
                  </div>
                  <div class="card p-6">
                    <h3 class="text-base font-bold text-gray-950 dark:text-white">{{ t('payment.checkoutSummary') }}</h3>
                    <div v-if="validAmount > 0" class="mt-5 space-y-3 text-sm">
                      <div class="flex justify-between gap-4">
                        <span class="text-gray-500 dark:text-gray-400">{{ t('payment.paymentAmount') }}</span>
                        <span class="font-semibold text-gray-900 dark:text-white">{{ formatRmbAmount(validAmount) }}</span>
                      </div>
                      <div v-if="selectedRechargeRail === 'usdt'" class="flex justify-between gap-4">
                        <span class="text-gray-500 dark:text-gray-400">{{ t('payment.usdtPayAmount') }}</span>
                        <span class="font-semibold text-gray-900 dark:text-white">{{ formatUsdtAmount(usdtPaymentAmount) }}</span>
                      </div>
                      <div v-if="feeRate > 0" class="flex justify-between gap-4">
                        <span class="text-gray-500 dark:text-gray-400">{{ t('payment.fee') }} ({{ feeRate }}%)</span>
                        <span class="font-semibold text-gray-900 dark:text-white">{{ formatRechargePaymentAmount(rechargeFeeAmount) }}</span>
                      </div>
                      <div class="flex justify-between gap-4 border-t border-gray-200 pt-3 dark:border-dark-600">
                        <span class="font-medium text-gray-700 dark:text-gray-300">{{ t('payment.actualPay') }}</span>
                        <span class="text-lg font-black text-gray-950 dark:text-white">{{ formatRechargePaymentAmount(rechargeTotalPaymentAmount) }}</span>
                      </div>
                      <div v-if="rechargeBonusAmount > 0" class="flex justify-between gap-4">
                        <span class="text-gray-500 dark:text-gray-400">{{ t('payment.rechargeBonus') }}</span>
                        <span class="font-semibold text-emerald-700 dark:text-emerald-300">{{ formatUsdBalance(rechargeBonusAmount) }}</span>
                      </div>
                      <div class="rounded-xl bg-emerald-50 p-4 dark:bg-emerald-900/20">
                        <div class="flex items-center justify-between gap-4">
                          <span class="text-sm font-semibold text-emerald-700 dark:text-emerald-300">{{ t('payment.creditedBalance') }}</span>
                          <span class="text-2xl font-black text-emerald-700 dark:text-emerald-300">{{ formatUsdBalance(creditedAmount) }}</span>
                        </div>
                        <p class="mt-2 text-xs leading-5 text-emerald-700/80 dark:text-emerald-300/80">
                          {{ t('payment.rechargeRatePreview', { usd: balanceRechargeMultiplier.toFixed(2) }) }}
                        </p>
                      </div>
                    </div>
                    <div v-else class="mt-5 rounded-xl border border-dashed border-gray-200 p-5 text-sm text-gray-500 dark:border-dark-700 dark:text-gray-400">
                      {{ t('payment.noPackageSelected') }}
                    </div>
                    <button :class="['btn mt-5 w-full py-3 text-base font-medium', paymentButtonClass]" :disabled="!canSubmit || submitting" @click="handleSubmitRecharge">
                      <span v-if="submitting" class="flex items-center justify-center gap-2">
                        <span class="h-4 w-4 animate-spin rounded-full border-2 border-white border-t-transparent"></span>
                        {{ t('common.processing') }}
                      </span>
                      <span v-else-if="rechargeMethods.length === 0">{{ t('payment.notAvailable') }}</span>
                      <span v-else>{{ t('payment.createOrder') }} {{ formatRechargePaymentAmount(rechargeTotalPaymentAmount) }}</span>
                    </button>
                    <div class="mt-5 space-y-2 border-t border-gray-200 pt-4 text-xs text-gray-500 dark:border-dark-700 dark:text-gray-400">
                      <p>{{ t('payment.rechargeRuleRate') }}</p>
                      <p>{{ t('payment.rechargeRuleNoExpiry') }}</p>
                      <p>{{ t('payment.rechargeRuleAllModels') }}</p>
                    </div>
                  </div>
                </div>
              </div>
            </div>
          </template>
          <!-- Subscribe Tab -->
          <template v-else-if="activeTab === 'subscription'">
            <div v-if="checkout.plans.length === 0" class="card py-16 text-center">
              <Icon name="gift" size="xl" class="mx-auto mb-3 text-gray-300 dark:text-dark-600" />
              <p class="text-gray-500 dark:text-gray-400">{{ t('payment.noPlans') }}</p>
            </div>
            <div v-else class="grid gap-6 lg:grid-cols-[minmax(0,1fr)_360px]">
              <div class="space-y-6">
                <div>
                  <div class="mb-4 flex flex-col gap-2 sm:flex-row sm:items-end sm:justify-between">
                    <div>
                      <h3 class="text-xl font-bold text-gray-950 dark:text-white">{{ t('payment.subscriptionPlanTitle') }}</h3>
                      <p class="mt-1 text-sm text-gray-500 dark:text-gray-400">{{ t('payment.subscriptionPlanSubtitle') }}</p>
                    </div>
                  </div>
                  <div :class="planGridClass">
                    <SubscriptionPlanCard
                      v-for="plan in checkout.plans"
                      :key="plan.id"
                      :plan="plan"
                      :active-subscriptions="activeSubscriptions"
                      :selected="selectedPlan?.id === plan.id"
                      @select="selectPlan"
                    />
                  </div>
                </div>

                <div v-if="activeSubscriptions.length > 0">
                  <p class="mb-2 text-xs font-medium text-gray-400 dark:text-gray-500">{{ t('payment.activeSubscription') }}</p>
                  <div class="space-y-2">
                    <div
                      v-for="sub in activeSubscriptions"
                      :key="sub.id"
                      class="flex items-center gap-3 rounded-xl border border-gray-100 bg-white px-3 py-2 shadow-sm dark:border-dark-700 dark:bg-dark-800"
                    >
                      <div :class="['h-6 w-1 shrink-0 rounded-full', platformAccentBarClass(sub.group?.platform || '')]" />
                      <div class="min-w-0 flex-1">
                        <div class="flex items-center gap-1.5">
                          <span class="truncate text-xs font-semibold text-gray-900 dark:text-white">{{ sub.group?.name || t('payment.groupFallback', { id: sub.group_id }) }}</span>
                          <span :class="['shrink-0 rounded-full px-1.5 py-0.5 text-[9px] font-medium', platformBadgeLightClass(sub.group?.platform || '')]">{{ platformLabel(sub.group?.platform || '') }}</span>
                        </div>
                        <div class="flex flex-wrap gap-x-3 text-[11px] text-gray-400 dark:text-gray-500">
                          <span>{{ t('payment.planCard.rate') }}: x{{ sub.group?.rate_multiplier ?? 1 }}</span>
                          <span v-if="sub.group?.daily_limit_usd == null && sub.group?.weekly_limit_usd == null && sub.group?.monthly_limit_usd == null">{{ t('payment.planCard.quota') }}: {{ t('payment.planCard.unlimited') }}</span>
                          <span v-if="sub.expires_at">{{ t('userSubscriptions.daysRemaining', { days: getDaysRemaining(sub.expires_at) }) }}</span>
                          <span v-else>{{ t('userSubscriptions.noExpiration') }}</span>
                        </div>
                      </div>
                      <span class="badge badge-success shrink-0 text-[10px]">{{ t('userSubscriptions.status.active') }}</span>
                    </div>
                  </div>
                </div>
              </div>

              <div class="space-y-4 lg:sticky lg:top-24 lg:self-start">
                <div v-if="subscriptionMethods.length >= 1" class="card p-6">
                  <PaymentMethodSelector
                    :methods="subMethodOptions"
                    :selected="selectedSubscriptionMethod"
                    @select="selectedSubscriptionMethod = $event"
                  />
                </div>
                <div v-else class="card p-6">
                  <h3 class="text-base font-bold text-gray-950 dark:text-white">{{ t('payment.paymentMethod') }}</h3>
                  <p class="mt-2 text-sm leading-6 text-gray-500 dark:text-gray-400">{{ t('payment.methodUnavailableHint') }}</p>
                </div>

                <div v-if="selectedSubscriptionMethod === 'balance' && balanceSubscriptionAllowed" class="card relative z-40 overflow-visible p-6">
                  <h3 class="text-base font-bold text-gray-950 dark:text-white">{{ t('payment.subscriptionTargetTitle') }}</h3>
                  <p class="mt-2 text-sm leading-6 text-gray-500 dark:text-gray-400">{{ t('payment.subscriptionTargetDesc') }}</p>
                  <label class="mt-4 block text-sm font-semibold text-gray-700 dark:text-gray-200" for="subscription-target-email">
                    {{ t('payment.subscriptionTargetEmail') }}
                  </label>
                  <div class="relative mt-2">
                    <Icon name="search" size="sm" class="pointer-events-none absolute left-4 top-1/2 -translate-y-1/2 text-gray-400" />
                    <input
                      id="subscription-target-email"
                      v-model="subscriptionTargetKeyword"
                      type="search"
                      autocomplete="off"
                      class="w-full rounded-xl border border-gray-200 bg-white py-3 pl-11 pr-11 text-sm text-gray-900 outline-none transition focus:border-primary-400 focus:ring-4 focus:ring-primary-100 dark:border-dark-700 dark:bg-dark-800 dark:text-white dark:focus:border-primary-500 dark:focus:ring-primary-950"
                      :placeholder="t('payment.subscriptionTargetPlaceholder')"
                      @focus="openSubscriptionTargetDropdown"
                      @input="handleSubscriptionTargetInput"
                      @keydown.down.prevent="moveSubscriptionTargetHighlight(1)"
                      @keydown.up.prevent="moveSubscriptionTargetHighlight(-1)"
                      @keydown.enter.prevent="confirmHighlightedSubscriptionTarget"
                      @keydown.esc="subscriptionTargetDropdownOpen = false"
                    />
                    <button
                      v-if="subscriptionTargetKeyword"
                      type="button"
                      class="absolute right-3 top-1/2 rounded-lg p-1 text-gray-400 transition hover:bg-gray-100 hover:text-gray-600 dark:hover:bg-dark-700 dark:hover:text-gray-200"
                      :title="t('common.clear')"
                      @click="clearSubscriptionTarget"
                    >
                      <Icon name="x" size="xs" :stroke-width="2" />
                    </button>

                    <div
                      v-if="subscriptionTargetDropdownVisible"
                      class="absolute left-0 right-0 top-full z-50 mt-2 max-h-64 overflow-y-auto rounded-xl border border-gray-200 bg-white p-1 shadow-xl dark:border-dark-700 dark:bg-dark-900"
                    >
                      <div v-if="subscriptionTargetSearching" class="flex items-center gap-2 px-3 py-3 text-sm text-gray-500 dark:text-gray-400">
                        <span class="h-4 w-4 animate-spin rounded-full border-2 border-primary-400 border-t-transparent"></span>
                        <span>{{ t('payment.subscriptionTargetSearching') }}</span>
                      </div>
                      <button
                        v-for="(option, index) in subscriptionTargetOptions"
                        :key="option.id"
                        type="button"
                        :class="[
                          'flex w-full items-center gap-3 rounded-lg px-3 py-2.5 text-left transition',
                          index === highlightedSubscriptionTargetIndex
                            ? 'bg-primary-50 text-primary-700 dark:bg-primary-900/20 dark:text-primary-200'
                            : 'text-gray-700 hover:bg-gray-50 dark:text-gray-200 dark:hover:bg-dark-800',
                        ]"
                        @mousedown.prevent="selectSubscriptionTarget(option)"
                      >
                        <span class="flex h-8 w-8 shrink-0 items-center justify-center rounded-full bg-primary-100 text-xs font-bold text-primary-700 dark:bg-primary-900/40 dark:text-primary-200">
                          {{ subscriptionTargetInitial(option) }}
                        </span>
                        <span class="min-w-0 flex-1">
                          <span class="block truncate text-sm font-semibold">{{ option.username || option.email }}</span>
                          <span class="block truncate text-xs text-gray-500 dark:text-gray-400">{{ option.email }}</span>
                        </span>
                        <Icon v-if="option.email === subscriptionTargetEmailTrimmed" name="check" size="sm" class="shrink-0 text-primary-600 dark:text-primary-300" />
                      </button>
                      <div
                        v-if="!subscriptionTargetSearching && subscriptionTargetKeywordReady && subscriptionTargetOptions.length === 0"
                        class="px-3 py-3 text-sm text-gray-500 dark:text-gray-400"
                      >
                        {{ t('payment.subscriptionTargetNoResults') }}
                      </div>
                    </div>
                  </div>
                  <p class="mt-2 text-xs leading-5 text-gray-500 dark:text-gray-400">
                    {{ subscriptionTargetStatusText }}
                  </p>
                </div>

                <div class="card p-6">
                  <h3 class="text-base font-bold text-gray-950 dark:text-white">{{ t('payment.subscriptionSummaryTitle') }}</h3>
                  <div v-if="selectedPlan" class="mt-5 space-y-4">
                    <div class="rounded-xl border border-gray-200 bg-gray-50/80 p-4 dark:border-dark-700 dark:bg-dark-800/70">
                      <div class="mb-3 flex flex-wrap items-center gap-2">
                        <span :class="['rounded-md border px-2 py-0.5 text-xs font-medium', planBadgeClass]">
                          {{ platformLabel(selectedPlan.group_platform || '') }}
                        </span>
                        <span v-if="selectedPlanHasSavings" :class="['rounded-full px-2 py-0.5 text-xs font-semibold', planDiscountClass]">
                          {{ selectedPlanDiscountText }}
                        </span>
                        <span v-if="selectedPlanLimitedOfferActive" class="rounded-full bg-rose-100 px-2 py-0.5 text-xs font-bold text-rose-700 dark:bg-rose-900/30 dark:text-rose-300">
                          {{ t('payment.planCard.limitedOffer') }}
                        </span>
                      </div>
                      <h4 class="text-lg font-black text-gray-950 dark:text-white">{{ selectedPlan.name }}</h4>
                      <p v-if="selectedPlan.description" class="mt-1 text-sm leading-6 text-gray-500 dark:text-gray-400">
                        {{ selectedPlan.description }}
                      </p>
                      <div class="mt-4 flex flex-wrap items-end justify-between gap-3">
                        <div>
                          <p class="text-xs text-gray-500 dark:text-gray-400">{{ t('payment.planCard.currentPrice') }}</p>
                          <p class="mt-1">
                            <span :class="['text-3xl font-black', planTextClass]">{{ formatSelectedPlanAmount(selectedPlan.price) }}</span>
                            <span class="text-sm text-gray-500 dark:text-gray-400"> / {{ planValiditySuffix }}</span>
                          </p>
                        </div>
                        <div v-if="selectedPlan.original_price || selectedPlanLimitedOfferActive" class="text-right">
                          <p class="text-xs text-gray-400 dark:text-gray-500">{{ selectedPlanLimitedOfferActive ? t('payment.planCard.restorePrice') : t('payment.planCard.originalPrice') }}</p>
                          <p class="text-sm font-semibold text-gray-400 line-through dark:text-gray-500">
                            {{ formatSelectedPlanAmount(selectedPlanLimitedOfferActive ? selectedPlanRegularPrice : (selectedPlan.original_price || 0)) }}
                          </p>
                          <p v-if="selectedPlanHasSavings" class="mt-1 text-xs font-bold text-emerald-700 dark:text-emerald-300">
                            {{ selectedPlanSavingsText }}
                          </p>
                        </div>
                      </div>
                    </div>

                    <div
                      v-if="selectedPlanLimitedOfferActive"
                      class="rounded-xl border border-rose-200 bg-rose-50 px-4 py-3 text-xs font-semibold text-rose-700 dark:border-rose-900/50 dark:bg-rose-900/20 dark:text-rose-200"
                    >
                      {{ t('payment.planCard.limitedOfferUntil', { time: selectedPlanLimitedOfferEndText, price: formatSelectedPlanAmount(selectedPlanRegularPrice) }) }}
                    </div>

                    <div class="grid grid-cols-2 gap-3 text-sm">
                      <div class="rounded-xl border border-gray-200 bg-white/80 p-3 dark:border-dark-700 dark:bg-dark-800/80">
                        <span class="text-xs text-gray-400 dark:text-gray-500">{{ t('payment.planCard.rate') }}</span>
                        <div :class="['mt-1 text-lg font-black', planTextClass]">{{ selectedPlanRateDisplay }}</div>
                      </div>
                      <div class="rounded-xl border border-gray-200 bg-white/80 p-3 dark:border-dark-700 dark:bg-dark-800/80">
                        <span class="text-xs text-gray-400 dark:text-gray-500">{{ t('payment.subscriptionValidityLabel') }}</span>
                        <div class="mt-1 text-lg font-black text-gray-950 dark:text-white">{{ selectedPlanValidityDuration }}</div>
                      </div>
                      <div v-if="selectedPlan.daily_limit_usd != null" class="rounded-xl border border-gray-200 bg-white/80 p-3 dark:border-dark-700 dark:bg-dark-800/80">
                        <span class="text-xs text-gray-400 dark:text-gray-500">{{ t('payment.planCard.dailyLimit') }}</span>
                        <div class="mt-1 text-lg font-black text-gray-950 dark:text-white">{{ formatUsdQuotaValue(selectedPlan.daily_limit_usd) }}</div>
                      </div>
                      <div v-if="selectedPlan.weekly_limit_usd != null" class="rounded-xl border border-gray-200 bg-white/80 p-3 dark:border-dark-700 dark:bg-dark-800/80">
                        <span class="text-xs text-gray-400 dark:text-gray-500">{{ t('payment.planCard.weeklyLimit') }}</span>
                        <div class="mt-1 text-lg font-black text-gray-950 dark:text-white">{{ formatUsdQuotaValue(selectedPlan.weekly_limit_usd) }}</div>
                      </div>
                      <div v-if="selectedPlan.monthly_limit_usd != null" class="rounded-xl border border-gray-200 bg-white/80 p-3 dark:border-dark-700 dark:bg-dark-800/80">
                        <span class="text-xs text-gray-400 dark:text-gray-500">{{ t('payment.planCard.monthlyLimit') }}</span>
                        <div class="mt-1 text-lg font-black text-gray-950 dark:text-white">{{ formatUsdQuotaValue(selectedPlan.monthly_limit_usd) }}</div>
                      </div>
                      <div
                        v-if="selectedPlan.daily_limit_usd == null && selectedPlan.weekly_limit_usd == null && selectedPlan.monthly_limit_usd == null"
                        class="rounded-xl border border-gray-200 bg-white/80 p-3 dark:border-dark-700 dark:bg-dark-800/80"
                      >
                        <span class="text-xs text-gray-400 dark:text-gray-500">{{ t('payment.planCard.quota') }}</span>
                        <div class="mt-1 text-lg font-black text-gray-950 dark:text-white">{{ t('payment.planCard.unlimited') }}</div>
                      </div>
                    </div>

                    <div
                      v-if="selectedPlanWorkdayFriendly"
                      class="rounded-xl border border-emerald-200 bg-emerald-50 px-4 py-3 dark:border-emerald-900/50 dark:bg-emerald-900/20"
                    >
                      <div class="flex items-center gap-2 text-sm font-bold text-emerald-800 dark:text-emerald-200">
                        <Icon name="calendar" size="sm" :stroke-width="2" />
                        <span>{{ t('payment.planCard.workdayFriendly') }}</span>
                      </div>
                      <p class="mt-1 text-xs leading-5 text-emerald-700/90 dark:text-emerald-200/80">
                        {{ t('payment.planCard.workdayFriendlyDesc', { daily: formatUsdQuotaValue(selectedPlan.daily_limit_usd), weekly: formatUsdQuotaValue(selectedPlan.weekly_limit_usd) }) }}
                      </p>
                    </div>

                    <div v-if="selectedPlanHasSavings" class="rounded-xl bg-emerald-50 px-4 py-3 text-xs font-bold text-emerald-700 dark:bg-emerald-900/20 dark:text-emerald-300">
                      {{ t('payment.planCard.dealHint', { original: formatSelectedPlanAmount(selectedPlanSavingsComparePrice), price: formatSelectedPlanAmount(selectedPlan.price) }) }}
                    </div>

                    <div class="space-y-3 border-t border-gray-200 pt-4 text-sm dark:border-dark-700">
                      <div class="flex justify-between gap-4">
                        <span class="text-gray-500 dark:text-gray-400">{{ t('payment.amountLabel') }}</span>
                        <span class="font-semibold text-gray-900 dark:text-white">{{ formatSelectedPaymentAmount(subscriptionBaseAmount) }}</span>
                      </div>
                      <div v-if="selectedSubscriptionMethod !== 'balance' && feeRate > 0 && selectedPlan.price > 0" class="flex justify-between gap-4">
                        <span class="text-gray-500 dark:text-gray-400">{{ t('payment.fee') }} ({{ feeRate }}%)</span>
                        <span class="font-semibold text-gray-900 dark:text-white">{{ formatSelectedPaymentAmount(subFeeAmount) }}</span>
                      </div>
                      <div v-if="selectedSubscriptionMethod === 'balance'" class="flex justify-between gap-4">
                        <span class="text-gray-500 dark:text-gray-400">{{ t('payment.subscriptionTargetSummary') }}</span>
                        <span class="max-w-[180px] truncate text-right font-semibold text-gray-900 dark:text-white">{{ subscriptionTargetDisplay }}</span>
                      </div>
                      <div class="flex justify-between gap-4 border-t border-gray-200 pt-3 dark:border-dark-600">
                        <span class="font-medium text-gray-700 dark:text-gray-300">{{ t('payment.actualPay') }}</span>
                        <span class="text-lg font-black text-gray-950 dark:text-white">{{ formatSelectedPaymentAmount(subscriptionPayableAmount) }}</span>
                      </div>
                    </div>
                  </div>
                  <div v-else class="mt-5 rounded-xl border border-dashed border-gray-200 p-5 text-sm text-gray-500 dark:border-dark-700 dark:text-gray-400">
                    {{ t('payment.noSubscriptionPlanSelected') }}
                  </div>

                  <button :class="['btn mt-5 w-full py-3 text-base font-medium', paymentButtonClass]" :disabled="!canSubmitSubscription || submitting" @click="confirmSubscribe">
                    <span v-if="submitting" class="flex items-center justify-center gap-2">
                      <span class="h-4 w-4 animate-spin rounded-full border-2 border-white border-t-transparent"></span>
                      {{ t('common.processing') }}
                    </span>
                    <span v-else-if="subscriptionMethods.length === 0">{{ t('payment.notAvailable') }}</span>
                    <span v-else-if="!selectedPlan">{{ t('payment.selectPlan') }}</span>
                    <span v-else>{{ t('payment.createOrder') }} {{ formatSelectedPaymentAmount(subscriptionPayableAmount) }}</span>
                  </button>
                  <button v-if="selectedPlan" class="btn btn-secondary mt-3 w-full" @click="selectedPlan = null">{{ t('common.cancel') }}</button>

                  <div class="mt-5 space-y-2 border-t border-gray-200 pt-4 text-xs text-gray-500 dark:border-dark-700 dark:text-gray-400">
                    <p>{{ t('payment.subscriptionRuleWorkday') }}</p>
                    <p>{{ t('payment.subscriptionRuleValidity') }}</p>
                    <p>{{ t('payment.subscriptionRuleAutoActivate') }}</p>
                    <p>{{ t('payment.subscriptionRuleOfferVolatile') }}</p>
                  </div>
                </div>
              </div>
            </div>
          </template>
        </template>
        <div v-if="(checkout.help_text || checkout.help_image_url) && paymentPhase === 'select'" class="card p-4">
          <div class="flex flex-col items-center gap-3">
            <img v-if="checkout.help_image_url" :src="checkout.help_image_url" alt=""
              class="h-40 max-w-full cursor-pointer rounded-lg object-contain transition-opacity hover:opacity-80"
              @click="previewImage = checkout.help_image_url" />
            <p v-if="checkout.help_text" class="text-center text-sm text-gray-500 dark:text-gray-400">{{ checkout.help_text }}</p>
          </div>
        </div>
      </template>
    </div>
    <!-- Renewal Plan Selection Modal -->
    <Teleport to="body">
      <Transition name="modal">
        <div v-if="showRenewalModal" class="fixed inset-0 z-50 flex items-center justify-center bg-black/60 backdrop-blur-sm p-4" @click.self="closeRenewalModal">
          <div class="relative w-full max-w-lg rounded-2xl border border-gray-200 bg-white p-6 shadow-2xl dark:border-dark-700 dark:bg-dark-900">
            <!-- Close button -->
            <button class="absolute right-4 top-4 rounded-lg p-1 text-gray-400 transition-colors hover:bg-gray-100 hover:text-gray-600 dark:hover:bg-dark-700 dark:hover:text-gray-200" @click="closeRenewalModal">
              <svg class="h-5 w-5" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="2"><path stroke-linecap="round" stroke-linejoin="round" d="M6 18L18 6M6 6l12 12" /></svg>
            </button>
            <h3 class="mb-4 text-lg font-semibold text-gray-900 dark:text-white">{{ t('payment.selectPlan') }}</h3>
            <div class="space-y-4">
              <SubscriptionPlanCard v-for="plan in renewalPlans" :key="plan.id" :plan="plan" :active-subscriptions="activeSubscriptions" @select="selectPlanFromModal" />
            </div>
          </div>
        </div>
      </Transition>
    </Teleport>
    <!-- Image Preview Overlay -->
    <Teleport to="body">
      <Transition name="modal">
        <div v-if="previewImage" class="fixed inset-0 z-[60] flex items-center justify-center bg-black/70 backdrop-blur-sm" @click="previewImage = ''">
          <img :src="previewImage" alt="" class="max-h-[85vh] max-w-[90vw] rounded-xl object-contain shadow-2xl" />
        </div>
      </Transition>
    </Teleport>
  </AppLayout>
</template>

<script setup lang="ts">
import { ref, computed, onMounted, watch } from 'vue'
import { useI18n } from 'vue-i18n'
import { useRoute, useRouter } from 'vue-router'
import { useAuthStore } from '@/stores/auth'
import { usePaymentStore } from '@/stores/payment'
import { useSubscriptionStore } from '@/stores/subscriptions'
import { useAppStore } from '@/stores'
import { paymentAPI } from '@/api/payment'
import { extractApiErrorMessage, extractI18nErrorMessage } from '@/utils/apiError'
import { isMobileDevice } from '@/utils/device'
import type { SubscriptionPlan, CheckoutInfoResponse, CreateOrderResult, OrderType, PaymentSubscriptionTargetUser } from '@/types/payment'
import AppLayout from '@/components/layout/AppLayout.vue'
import AmountInput from '@/components/payment/AmountInput.vue'
import PaymentMethodSelector from '@/components/payment/PaymentMethodSelector.vue'
import { METHOD_ORDER, getPaymentPopupFeatures } from '@/components/payment/providerConfig'
import {
  PAYMENT_RECOVERY_STORAGE_KEY,
  buildCreateOrderPayload,
  clearPaymentRecoverySnapshot,
  decidePaymentLaunch,
  getVisibleMethods,
  normalizeVisibleMethod,
  readPaymentRecoverySnapshot,
  type PaymentRecoverySnapshot,
  writePaymentRecoverySnapshot,
} from '@/components/payment/paymentFlow'
import {
  platformAccentBarClass,
  platformBadgeLightClass,
  platformBadgeClass,
  platformDiscountClass,
  platformTextClass,
  platformLabel,
} from '@/utils/platformColors'
import SubscriptionPlanCard from '@/components/payment/SubscriptionPlanCard.vue'
import PaymentStatusPanel from '@/components/payment/PaymentStatusPanel.vue'
import Icon from '@/components/icons/Icon.vue'
import { formatPaymentAmount, normalizePaymentCurrency } from '@/components/payment/currency'
import type { PaymentMethodOption } from '@/components/payment/PaymentMethodSelector.vue'
import { buildPaymentErrorToastMessage, describePaymentScenarioError } from './paymentUx'
import { hasWechatResumeQuery, parseWechatResumeRoute, stripWechatResumeQuery } from './paymentWechatResume'

const i18n = useI18n()
const { t } = i18n
const route = useRoute()
const router = useRouter()
const authStore = useAuthStore()
const paymentStore = usePaymentStore()
const subscriptionStore = useSubscriptionStore()
const appStore = useAppStore()

const user = computed(() => authStore.user)
const activeSubscriptions = computed(() => subscriptionStore.activeSubscriptions)

function getDaysRemaining(expiresAt: string): number {
  const diff = new Date(expiresAt).getTime() - Date.now()
  return Math.max(0, Math.ceil(diff / (1000 * 60 * 60 * 24)))
}

const loading = ref(true)
const submitting = ref(false)
const errorMessage = ref('')
const errorHintMessage = ref('')
const activeTab = ref<'recharge' | 'subscription'>('recharge')
const amount = ref<number | null>(null)
const selectedRechargeMethod = ref('')
const selectedSubscriptionMethod = ref('')
const selectedRechargeRail = ref<'rmb' | 'usdt'>('rmb')
const selectedPlan = ref<SubscriptionPlan | null>(null)
const subscriptionTargetEmail = ref('')
const subscriptionTargetKeyword = ref('')
const subscriptionTargetOptions = ref<PaymentSubscriptionTargetUser[]>([])
const subscriptionTargetSearching = ref(false)
const subscriptionTargetDropdownOpen = ref(false)
const highlightedSubscriptionTargetIndex = ref(0)
const previewImage = ref('')
let subscriptionTargetSearchTimer: ReturnType<typeof setTimeout> | null = null
let subscriptionTargetSearchSeq = 0

const paymentPhase = ref<'select' | 'paying'>('select')

interface CreateOrderOptions {
  openid?: string
  wechatResumeToken?: string
  paymentType?: string
  paymentAmount?: number
  targetUserEmail?: string
  isResume?: boolean
  mobileQrFallbackAttempted?: boolean
}

interface WeixinJSBridgeLike {
  invoke(
    action: string,
    payload: Record<string, unknown>,
    callback: (result: Record<string, unknown>) => void,
  ): void
}

function emptyPaymentState(): PaymentRecoverySnapshot {
  return {
    orderId: 0,
    amount: 0,
    qrCode: '',
    expiresAt: '',
    paymentType: '',
    payUrl: '',
    outTradeNo: '',
    clientSecret: '',
    intentId: '',
    currency: '',
    countryCode: '',
    paymentEnv: '',
    payAmount: 0,
    orderType: '',
    paymentMode: '',
    resumeToken: '',
    createdAt: 0,
  }
}

function getWeixinJSBridge(): WeixinJSBridgeLike | undefined {
  return (window as Window & { WeixinJSBridge?: WeixinJSBridgeLike }).WeixinJSBridge
}

function waitForWeixinJSBridge(timeoutMs = 4000): Promise<WeixinJSBridgeLike | null> {
  const existing = getWeixinJSBridge()
  if (existing) return Promise.resolve(existing)

  return new Promise((resolve) => {
    let settled = false
    const finish = (bridge: WeixinJSBridgeLike | null) => {
      if (settled) return
      settled = true
      document.removeEventListener('WeixinJSBridgeReady', handleReady)
      document.removeEventListener('onWeixinJSBridgeReady', handleReady)
      window.clearTimeout(timer)
      resolve(bridge)
    }
    const handleReady = () => finish(getWeixinJSBridge() ?? null)
    const timer = window.setTimeout(() => finish(getWeixinJSBridge() ?? null), timeoutMs)
    document.addEventListener('WeixinJSBridgeReady', handleReady, false)
    document.addEventListener('onWeixinJSBridgeReady', handleReady, false)
  })
}

async function invokeWechatJsapiPayment(payload: Record<string, unknown>): Promise<Record<string, unknown>> {
  const bridge = await waitForWeixinJSBridge()
  if (!bridge) {
    throw new Error('WECHAT_JSAPI_UNAVAILABLE')
  }
  return new Promise((resolve) => {
    bridge.invoke('getBrandWCPayRequest', payload, (result) => resolve(result || {}))
  })
}

const paymentState = ref<PaymentRecoverySnapshot>(emptyPaymentState())

function persistRecoverySnapshot(snapshot: PaymentRecoverySnapshot) {
  if (typeof window === 'undefined' || !snapshot.orderId) return
  writePaymentRecoverySnapshot(window.localStorage, snapshot, PAYMENT_RECOVERY_STORAGE_KEY)
}

function removeRecoverySnapshot() {
  if (typeof window === 'undefined') return
  clearPaymentRecoverySnapshot(window.localStorage, PAYMENT_RECOVERY_STORAGE_KEY)
}

function resetPayment() {
  paymentPhase.value = 'select'
  paymentState.value = emptyPaymentState()
  removeRecoverySnapshot()
}

async function redirectToPaymentResult(state: PaymentRecoverySnapshot): Promise<void> {
  const query: Record<string, string | undefined> = {}
  if (state.orderId > 0) {
    query.order_id = String(state.orderId)
  }
  if (state.outTradeNo) {
    query.out_trade_no = state.outTradeNo
  }
  if (state.resumeToken) {
    query.resume_token = state.resumeToken
  }
  await router.push({
    path: '/payment/result',
    query,
  })
}

function buildWechatOAuthAuthorizeUrl(
  authorizeUrl: string,
  context: { paymentType: string; orderType: OrderType; planId?: number; orderAmount: number },
): string {
  const normalizedUrl = authorizeUrl.trim()
  if (!normalizedUrl || typeof window === 'undefined') {
    return normalizedUrl
  }

  try {
    const targetUrl = new URL(normalizedUrl, window.location.origin)
    const redirectPath = targetUrl.searchParams.get('redirect') || '/purchase'
    const redirectUrl = new URL(redirectPath, window.location.origin)
    const paymentType = normalizeVisibleMethod(context.paymentType) || context.paymentType.trim() || 'wxpay'

    redirectUrl.searchParams.set('payment_type', paymentType)
    redirectUrl.searchParams.set('order_type', context.orderType)

    if (context.planId) {
      redirectUrl.searchParams.set('plan_id', String(context.planId))
    } else {
      redirectUrl.searchParams.delete('plan_id')
    }

    if (context.orderAmount > 0) {
      redirectUrl.searchParams.set('amount', String(context.orderAmount))
    } else {
      redirectUrl.searchParams.delete('amount')
    }

    targetUrl.searchParams.set('redirect', `${redirectUrl.pathname}${redirectUrl.search}`)
    return targetUrl.toString()
  } catch {
    return normalizedUrl
  }
}

function onPaymentDone() {
  const wasSubscription = paymentState.value.orderType === 'subscription'
  resetPayment()
  selectedPlan.value = null
  if (wasSubscription) {
    subscriptionStore.fetchActiveSubscriptions(true).catch(() => {})
  }
}

function onPaymentSuccess() {
  removeRecoverySnapshot()
  authStore.refreshUser()
  if (paymentState.value.orderType === 'subscription') {
    subscriptionStore.fetchActiveSubscriptions(true).catch(() => {})
  }
}

function onPaymentSettled() {
  removeRecoverySnapshot()
}

// All checkout data from single API call
const checkout = ref<CheckoutInfoResponse>({
  methods: {}, global_min: 0, global_max: 0,
  plans: [], balance_disabled: false, balance_recharge_multiplier: 1, recharge_fee_rate: 0, recharge_bonus_threshold: 100, recharge_bonus_amount: 10, usdt_cny_exchange_rate: 7, help_text: '', help_image_url: '', stripe_publishable_key: '', allow_balance_subscription_purchase: false,
})

const tabs = computed(() => {
  const result: { key: 'recharge' | 'subscription'; label: string }[] = []
  if (!checkout.value.balance_disabled) result.push({ key: 'recharge', label: t('payment.tabTopUp') })
  result.push({ key: 'subscription', label: t('payment.tabSubscribe') })
  return result
})

const visibleMethods = computed(() => getVisibleMethods(checkout.value.methods))
const rmbPaymentMethods = ['alipay', 'wxpay']
const usdtPaymentMethods = ['usdt']
function isVisibleMethodOpen(method: string): boolean {
  const limit = visibleMethods.value[method]
  return !!limit && limit.available !== false
}
const rmbRailAvailable = computed(() => rmbPaymentMethods.some((method) => isVisibleMethodOpen(method)))
const usdtRailAvailable = computed(() => usdtPaymentMethods.some((method) => isVisibleMethodOpen(method)))
const rechargeMethods = computed(() => {
  const allowed = selectedRechargeRail.value === 'rmb' ? rmbPaymentMethods : usdtPaymentMethods
  return allowed.filter((method) => isVisibleMethodOpen(method))
})
const enabledMethods = computed(() => Object.keys(visibleMethods.value))
const balanceSubscriptionAllowed = computed(() => checkout.value.allow_balance_subscription_purchase === true)
const subscriptionTargetEmailTrimmed = computed(() => subscriptionTargetEmail.value.trim())
const subscriptionTargetKeywordTrimmed = computed(() => subscriptionTargetKeyword.value.trim())
const subscriptionTargetKeywordReady = computed(() => subscriptionTargetKeywordTrimmed.value.length >= 2)
const subscriptionTargetNeedsSelection = computed(() =>
  selectedSubscriptionMethod.value === 'balance'
    && balanceSubscriptionAllowed.value
    && subscriptionTargetKeywordTrimmed.value !== ''
    && subscriptionTargetEmailTrimmed.value === ''
)
const subscriptionTargetDropdownVisible = computed(() =>
  subscriptionTargetDropdownOpen.value
    && subscriptionTargetKeywordTrimmed.value !== ''
    && (
      subscriptionTargetSearching.value
      || subscriptionTargetKeywordReady.value
      || subscriptionTargetOptions.value.length > 0
    )
)
const subscriptionTargetDisplay = computed(() => {
  if (subscriptionTargetEmailTrimmed.value) {
    return subscriptionTargetEmailTrimmed.value
  }
  return user.value?.email || user.value?.username || t('payment.subscriptionTargetCurrentUser')
})
const subscriptionTargetStatusText = computed(() => {
  if (subscriptionTargetNeedsSelection.value) {
    return t('payment.subscriptionTargetSelectRequired')
  }
  if (subscriptionTargetEmailTrimmed.value) {
    return t('payment.subscriptionTargetSelectedHint', { email: subscriptionTargetEmailTrimmed.value })
  }
  return t('payment.subscriptionTargetHint')
})
const subscriptionMethods = computed(() => {
  const methods = enabledMethods.value.filter((method) => method !== 'balance')
  if (!balanceSubscriptionAllowed.value) return methods
  return ['balance', ...methods]
})

function subscriptionTargetInitial(option: PaymentSubscriptionTargetUser): string {
  const source = option.username || option.email || '?'
  return source.trim().slice(0, 1).toUpperCase() || '?'
}

function openSubscriptionTargetDropdown() {
  subscriptionTargetDropdownOpen.value = true
  scheduleSubscriptionTargetSearch()
}

function handleSubscriptionTargetInput() {
  subscriptionTargetEmail.value = ''
  highlightedSubscriptionTargetIndex.value = 0
  subscriptionTargetDropdownOpen.value = true
  scheduleSubscriptionTargetSearch()
}

function clearSubscriptionTarget() {
  subscriptionTargetKeyword.value = ''
  subscriptionTargetEmail.value = ''
  subscriptionTargetOptions.value = []
  subscriptionTargetSearching.value = false
  subscriptionTargetDropdownOpen.value = false
  highlightedSubscriptionTargetIndex.value = 0
  if (subscriptionTargetSearchTimer) {
    clearTimeout(subscriptionTargetSearchTimer)
    subscriptionTargetSearchTimer = null
  }
}

function selectSubscriptionTarget(option: PaymentSubscriptionTargetUser) {
  subscriptionTargetKeyword.value = option.username
    ? `${option.username} <${option.email}>`
    : option.email
  subscriptionTargetEmail.value = option.email
  subscriptionTargetDropdownOpen.value = false
  highlightedSubscriptionTargetIndex.value = 0
}

function moveSubscriptionTargetHighlight(delta: number) {
  if (!subscriptionTargetDropdownVisible.value || subscriptionTargetOptions.value.length === 0) {
    openSubscriptionTargetDropdown()
    return
  }
  const count = subscriptionTargetOptions.value.length
  highlightedSubscriptionTargetIndex.value = (highlightedSubscriptionTargetIndex.value + delta + count) % count
}

function confirmHighlightedSubscriptionTarget() {
  if (!subscriptionTargetDropdownVisible.value || subscriptionTargetOptions.value.length === 0) return
  selectSubscriptionTarget(subscriptionTargetOptions.value[highlightedSubscriptionTargetIndex.value] || subscriptionTargetOptions.value[0])
}

function scheduleSubscriptionTargetSearch() {
  if (subscriptionTargetSearchTimer) {
    clearTimeout(subscriptionTargetSearchTimer)
    subscriptionTargetSearchTimer = null
  }
  const keyword = subscriptionTargetKeywordTrimmed.value
  if (keyword.length < 2) {
    subscriptionTargetOptions.value = []
    subscriptionTargetSearching.value = false
    return
  }
  subscriptionTargetSearchTimer = setTimeout(() => {
    searchSubscriptionTargets(keyword)
  }, 250)
}

async function searchSubscriptionTargets(keyword: string) {
  const requestSeq = ++subscriptionTargetSearchSeq
  subscriptionTargetSearching.value = true
  try {
    const response = await paymentAPI.searchSubscriptionTargets(keyword, 8)
    if (requestSeq !== subscriptionTargetSearchSeq) return
    subscriptionTargetOptions.value = response.data || []
    highlightedSubscriptionTargetIndex.value = 0
  } catch {
    if (requestSeq !== subscriptionTargetSearchSeq) return
    subscriptionTargetOptions.value = []
  } finally {
    if (requestSeq === subscriptionTargetSearchSeq) {
      subscriptionTargetSearching.value = false
    }
  }
}

function firstSortedMethod(methods: string[]): string {
  if (methods.length === 0) return ''
  const order: readonly string[] = METHOD_ORDER
  return [...methods].sort((a, b) => {
    const ai = order.indexOf(a)
    const bi = order.indexOf(b)
    return (ai === -1 ? 999 : ai) - (bi === -1 ? 999 : bi)
  })[0] || ''
}

const validAmount = computed(() => amount.value ?? 0)
const balanceRechargeMultiplier = computed(() => {
  const multiplier = checkout.value.balance_recharge_multiplier
  return multiplier > 0 ? multiplier : 1
})
const rechargeBonusThreshold = computed(() => normalizeNonNegativeNumber(checkout.value.recharge_bonus_threshold, 100))
const rechargeBonusUnitAmount = computed(() => normalizeNonNegativeNumber(checkout.value.recharge_bonus_amount, 10))
const rechargeBonusAmount = computed(() => bonusForAmount(validAmount.value))
const creditedAmount = computed(() => creditedForAmount(validAmount.value))
const usdtCnyExchangeRate = computed(() => {
  const rate = Number(checkout.value.usdt_cny_exchange_rate)
  return Number.isFinite(rate) && rate > 0 ? rate : 7
})
const usdtCnyExchangeRateLabel = computed(() => `1 USDT = ${formatRateNumber(usdtCnyExchangeRate.value)} CNY`)
const usdtPaymentAmount = computed(() => usdtPaymentForAmount(validAmount.value))
const usdtFeeAmount = computed(() =>
  feeRate.value > 0 && usdtPaymentAmount.value > 0
    ? Math.ceil(((usdtPaymentAmount.value * feeRate.value) / 100) * 10000) / 10000
    : 0
)
const usdtTotalPaymentAmount = computed(() =>
  feeRate.value > 0 && usdtPaymentAmount.value > 0
    ? Math.round((usdtPaymentAmount.value + usdtFeeAmount.value) * 10000) / 10000
    : usdtPaymentAmount.value
)
const rmbFeeAmount = computed(() =>
  feeRate.value > 0 && validAmount.value > 0
    ? Math.ceil(((validAmount.value * feeRate.value) / 100) * 100) / 100
    : 0
)
const rmbTotalPaymentAmount = computed(() =>
  feeRate.value > 0 && validAmount.value > 0
    ? Math.round((validAmount.value + rmbFeeAmount.value) * 100) / 100
    : validAmount.value
)
const rechargeFeeAmount = computed(() => selectedRechargeRail.value === 'usdt' ? usdtFeeAmount.value : rmbFeeAmount.value)
const rechargeTotalPaymentAmount = computed(() => selectedRechargeRail.value === 'usdt' ? usdtTotalPaymentAmount.value : rmbTotalPaymentAmount.value)

const rechargePackages = [
  { amount: 30, nameKey: 'payment.packageStarter' },
  { amount: 100, nameKey: 'payment.packageStandard' },
  { amount: 300, nameKey: 'payment.packagePro', badgeKey: 'payment.packageRecommended' },
  { amount: 500, nameKey: 'payment.packageTeam', badgeKey: 'payment.packageBestValue' },
]

function creditedForAmount(value: number): number {
  return Math.round(((value * balanceRechargeMultiplier.value) + bonusForAmount(value)) * 100) / 100
}

function normalizeNonNegativeNumber(value: number | undefined, fallback: number): number {
  const amountValue = Number(value)
  return Number.isFinite(amountValue) && amountValue >= 0 ? amountValue : fallback
}

function bonusForAmount(value: number): number {
  const amountValue = Number.isFinite(value) ? value : 0
  const threshold = rechargeBonusThreshold.value
  const bonusAmount = rechargeBonusUnitAmount.value
  if (amountValue <= 0 || threshold <= 0 || bonusAmount <= 0) return 0
  return Math.floor(amountValue / threshold) * bonusAmount
}

function usdtPaymentForAmount(value: number): number {
  if (!Number.isFinite(value) || value <= 0) return 0
  return Math.ceil((value / usdtCnyExchangeRate.value) * 10000) / 10000
}

function amountFitsAnyMethod(value: number): boolean {
  if (rechargeMethods.value.length === 0) return true
  return rechargeMethods.value.some((m) => amountFitsMethod(value, m))
}

function selectRechargePackage(value: number) {
  if (!amountFitsAnyMethod(value)) return
  amount.value = value
}

function formatRmbAmount(value: number): string {
  try {
    return new Intl.NumberFormat(localeCode.value || undefined, {
      style: 'currency',
      currency: 'CNY',
      currencyDisplay: 'narrowSymbol',
      minimumFractionDigits: 0,
      maximumFractionDigits: 2,
    }).format(value)
  } catch {
    return `¥${formatRateNumber(value)}`
  }
}

function formatRateNumber(value: number): string {
  const amountValue = Number.isFinite(value) ? value : 0
  return amountValue.toFixed(2).replace(/\.?0+$/, '')
}

function formatUsdtAmount(value: number): string {
  const amountValue = Number.isFinite(value) ? value : 0
  return `${amountValue.toFixed(4).replace(/\.?0+$/, '')} USDT`
}

function formatRechargePaymentAmount(value: number): string {
  return selectedRechargeRail.value === 'usdt'
    ? formatUsdtAmount(value)
    : formatRmbAmount(value)
}

function formatUsdBalance(value: number): string {
  try {
    return new Intl.NumberFormat(localeCode.value || undefined, {
      style: 'currency',
      currency: 'USD',
      currencyDisplay: 'narrowSymbol',
      minimumFractionDigits: 2,
      maximumFractionDigits: 2,
    }).format(value)
  } catch {
    return `$${value.toFixed(2)}`
  }
}

// Adaptive grid: keep subscription cards roomy beside the checkout column.
const planGridClass = computed(() => {
  const n = checkout.value.plans.length
  if (n <= 2) return 'grid grid-cols-1 gap-4 xl:grid-cols-2'
  return 'grid grid-cols-1 gap-4 xl:grid-cols-2'
})

// Check if an amount fits a method's [min, max]. 0 = no limit.
function amountFitsMethod(amt: number, methodType: string): boolean {
  if (amt <= 0) return true
  if (methodType === 'balance') return true
  const ml = visibleMethods.value[methodType]
  if (!ml) return false
  const paymentAmount = amountForMethodLimit(amt, ml)
  if (ml.single_min > 0 && paymentAmount < ml.single_min) return false
  if (ml.single_max > 0 && paymentAmount > ml.single_max) return false
  return true
}

function amountForMethodLimit(value: number, limit: { currency?: string }): number {
  return normalizePaymentCurrency(limit.currency) === 'USDT'
    ? usdtPaymentForAmount(value)
    : value
}

function rmbAmountFromMethodLimit(value: number, limit: { currency?: string }): number {
  return normalizePaymentCurrency(limit.currency) === 'USDT'
    ? Math.ceil((value * usdtCnyExchangeRate.value) * 100) / 100
    : value
}

// Visible methods decide the amount range shown to users.
const globalMinAmount = computed(() => {
  const limits = rechargeMethods.value.map(method => visibleMethods.value[method]).filter(Boolean)
  if (limits.length === 0) return 0
  if (limits.some(limit => limit.single_min <= 0)) return 0
  return Math.min(...limits.map(limit => rmbAmountFromMethodLimit(limit.single_min, limit)))
})
const globalMaxAmount = computed(() => {
  const limits = rechargeMethods.value.map(method => visibleMethods.value[method]).filter(Boolean)
  if (limits.length === 0) return 0
  if (limits.some(limit => limit.single_max <= 0)) return 0
  return Math.max(...limits.map(limit => rmbAmountFromMethodLimit(limit.single_max, limit)))
})

const currentSelectedMethod = computed(() =>
  activeTab.value === 'subscription' ? selectedSubscriptionMethod.value : selectedRechargeMethod.value
)
const selectedRechargeLimit = computed(() => visibleMethods.value[selectedRechargeMethod.value])
const selectedSubscriptionLimit = computed(() => visibleMethods.value[selectedSubscriptionMethod.value])
// Selected method's limits (for validation and error messages)
const selectedLimit = computed(() =>
  activeTab.value === 'subscription' ? selectedSubscriptionLimit.value : selectedRechargeLimit.value
)
const selectedSubscriptionCurrency = computed(() =>
  selectedSubscriptionMethod.value === 'balance' ? 'USD' : normalizePaymentCurrency(selectedSubscriptionLimit.value?.currency)
)
const selectedCurrency = computed(() =>
  activeTab.value === 'subscription' ? selectedSubscriptionCurrency.value : normalizePaymentCurrency(selectedLimit.value?.currency)
)
const localeCode = computed(() => {
  const raw = i18n.locale as unknown
  if (typeof raw === 'string') return raw
  if (raw && typeof raw === 'object' && 'value' in raw) {
    return String((raw as { value?: string }).value || '')
  }
  return undefined
})

function formatSelectedPaymentAmount(value: number): string {
  return formatPaymentAmount(value, selectedCurrency.value, localeCode.value)
}

function formatSelectedPlanAmount(value: number): string {
  return formatPaymentAmount(value, 'CNY', localeCode.value)
}

function formatUsdQuotaValue(value: number | null | undefined): string {
  const amountValue = Number(value ?? 0)
  if (!Number.isFinite(amountValue)) return '$0'
  return `$${amountValue.toFixed(2).replace(/\.?0+$/, '')}`
}

const methodOptions = computed<PaymentMethodOption[]>(() =>
  rechargeMethods.value.map((type) => {
    const ml = visibleMethods.value[type]
    return {
      type,
      fee_rate: ml?.fee_rate ?? 0,
      available: ml?.available !== false && amountFitsMethod(validAmount.value, type),
    }
  })
)

const feeRate = computed(() => checkout.value?.recharge_fee_rate ?? 0)

const amountError = computed(() => {
  if (validAmount.value <= 0) return ''
  // No method can handle this amount
  if (!rechargeMethods.value.some((m) => amountFitsMethod(validAmount.value, m))) {
    return t('payment.amountNoMethod')
  }
  // Selected method can't handle this amount (but others can)
  const ml = selectedRechargeLimit.value
  if (ml) {
    const paymentAmount = amountForMethodLimit(validAmount.value, ml)
    if (ml.single_min > 0 && paymentAmount < ml.single_min) return t('payment.amountTooLow', { min: formatRmbAmount(rmbAmountFromMethodLimit(ml.single_min, ml)) })
    if (ml.single_max > 0 && paymentAmount > ml.single_max) return t('payment.amountTooHigh', { max: formatRmbAmount(rmbAmountFromMethodLimit(ml.single_max, ml)) })
  }
  return ''
})

const canSubmit = computed(() =>
  validAmount.value > 0
    && rechargeMethods.value.includes(selectedRechargeMethod.value)
    && amountFitsMethod(validAmount.value, selectedRechargeMethod.value)
    && selectedRechargeLimit.value?.available !== false
)

// Subscription-specific: method options based on plan price
const subMethodOptions = computed<PaymentMethodOption[]>(() => {
  const planPrice = selectedPlan.value?.price ?? 0
  return subscriptionMethods.value.map((type) => {
    const ml = visibleMethods.value[type]
    if (type === 'balance') {
      return {
        type,
        fee_rate: 0,
        available: planPrice <= 0 || (user.value?.balance ?? 0) >= balanceSubscriptionCost(planPrice),
      }
    }
    return {
      type,
      fee_rate: ml?.fee_rate ?? 0,
      available: ml?.available !== false && amountFitsMethod(planPrice, type),
    }
  })
})

function balanceSubscriptionCost(value: number): number {
  if (!Number.isFinite(value) || value <= 0) return 0
  return Math.round((value * balanceRechargeMultiplier.value) * 100) / 100
}

const subscriptionBaseAmount = computed(() => {
  const price = selectedPlan.value?.price ?? 0
  if (selectedSubscriptionMethod.value === 'balance') {
    return balanceSubscriptionCost(price)
  }
  if (selectedSubscriptionCurrency.value === 'USDT') {
    return usdtPaymentForAmount(price)
  }
  return price
})

const subFeeAmount = computed(() => {
  if (selectedSubscriptionMethod.value === 'balance') return 0
  const amountValue = subscriptionBaseAmount.value
  if (feeRate.value <= 0 || amountValue <= 0) return 0
  const scale = selectedSubscriptionCurrency.value === 'USDT' ? 10000 : 100
  return Math.ceil(((amountValue * feeRate.value) / 100) * scale) / scale
})

const subTotalAmount = computed(() => {
  if (selectedSubscriptionMethod.value === 'balance') return subscriptionBaseAmount.value
  const amountValue = subscriptionBaseAmount.value
  if (feeRate.value <= 0 || amountValue <= 0) return amountValue
  const scale = selectedSubscriptionCurrency.value === 'USDT' ? 10000 : 100
  return Math.round((amountValue + subFeeAmount.value) * scale) / scale
})

const subscriptionPayableAmount = computed(() =>
  feeRate.value > 0 ? subTotalAmount.value : subscriptionBaseAmount.value
)

const canSubmitSubscription = computed(() =>
  selectedPlan.value !== null
    && subscriptionMethods.value.includes(selectedSubscriptionMethod.value)
    && amountFitsMethod(selectedPlan.value.price, selectedSubscriptionMethod.value)
    && (
      selectedSubscriptionMethod.value === 'balance'
        ? (user.value?.balance ?? 0) >= balanceSubscriptionCost(selectedPlan.value.price) && !subscriptionTargetNeedsSelection.value
        : selectedSubscriptionLimit.value?.available !== false
    )
)

// Auto-switch to first available method when current selection can't handle the amount
watch(() => [validAmount.value, selectedRechargeMethod.value, selectedRechargeRail.value, rechargeMethods.value.join(',')] as const, ([amt, method]) => {
  if (amt > 0 && rechargeMethods.value.includes(method) && amountFitsMethod(amt, method)) return
  const available = rechargeMethods.value.find((m) => amountFitsMethod(amt, m))
  if (available) selectedRechargeMethod.value = available
})

watch(() => [selectedPlan.value?.price ?? 0, selectedSubscriptionMethod.value, subscriptionMethods.value.join(',')] as const, ([price, method]) => {
  if (method && subscriptionMethods.value.includes(method) && amountFitsMethod(price, method)) return
  const available = subscriptionMethods.value.find((m) => amountFitsMethod(price, m))
  if (available) selectedSubscriptionMethod.value = available
})

watch([rmbRailAvailable, usdtRailAvailable], ([rmbAvailable, usdtAvailable]) => {
  if (selectedRechargeRail.value === 'rmb' && !rmbAvailable && usdtAvailable) {
    selectedRechargeRail.value = 'usdt'
  } else if (selectedRechargeRail.value === 'usdt' && !usdtAvailable && rmbAvailable) {
    selectedRechargeRail.value = 'rmb'
  }
}, { immediate: true })

// Payment button class: follows selected payment method color
const paymentButtonClass = computed(() => {
  const m = currentSelectedMethod.value
  if (!m) return 'btn-primary'
  if (m.includes('alipay')) return 'btn-alipay'
  if (m.includes('wxpay')) return 'btn-wxpay'
  if (m === 'balance') return 'btn-usdt'
  if (m === 'usdt') return 'btn-usdt'
  if (m === 'stripe') return 'btn-stripe'
  if (m === 'airwallex') return 'btn-airwallex'
  return 'btn-primary'
})

// Subscription confirm: platform accent colors (clean card, no gradient)
const planBadgeClass = computed(() => platformBadgeClass(selectedPlan.value?.group_platform || ''))
const planDiscountClass = computed(() => platformDiscountClass(selectedPlan.value?.group_platform || ''))
const planTextClass = computed(() => platformTextClass(selectedPlan.value?.group_platform || ''))
const selectedPlanRateDisplay = computed(() => {
  const rate = selectedPlan.value?.rate_multiplier ?? 1
  return `x${Number(rate.toPrecision(10))}`
})

const selectedPlanRegularPrice = computed(() =>
  selectedPlan.value?.regular_price || selectedPlan.value?.price || 0
)

const selectedPlanSavingsComparePrice = computed(() => {
  if (!selectedPlan.value) return 0
  const originalPrice = selectedPlan.value.original_price || 0
  if (originalPrice > selectedPlan.value.price) return originalPrice
  if (selectedPlanLimitedOfferActive.value && selectedPlanRegularPrice.value > selectedPlan.value.price) {
    return selectedPlanRegularPrice.value
  }
  return 0
})

const selectedPlanLimitedOfferActive = computed(() => {
  const plan = selectedPlan.value
  if (!plan) return false
  if (plan.limited_offer_active != null) return plan.limited_offer_active
  if (!plan.limited_offer_price || !plan.limited_offer_expires_at) return false
  return new Date(plan.limited_offer_expires_at).getTime() > Date.now()
})

const selectedPlanLimitedOfferEndText = computed(() => {
  const value = selectedPlan.value?.limited_offer_expires_at
  if (!value) return ''
  const date = new Date(value)
  if (Number.isNaN(date.getTime())) return value
  return date.toLocaleString()
})

// Renewal modal state
const showRenewalModal = ref(false)
const renewGroupId = ref<number | null>(null)
const renewalPlans = computed(() => {
  if (renewGroupId.value == null) return []
  return checkout.value.plans.filter(p => p.group_id === renewGroupId.value)
})

const planValiditySuffix = computed(() => {
  if (!selectedPlan.value) return ''
  const u = normalizePlanValidityUnit(selectedPlan.value.validity_unit)
  if (u === 'month') return t('payment.perMonth')
  if (u === 'year') return t('payment.perYear')
  return `${selectedPlan.value.validity_days}${t('payment.days')}`
})

const selectedPlanValidityDuration = computed(() => {
  if (!selectedPlan.value) return ''
  return formatPlanValidityDuration(selectedPlan.value.validity_days, selectedPlan.value.validity_unit)
})

function normalizePlanValidityUnit(unit: string | undefined): string {
  const value = (unit || 'day').toLowerCase()
  if (value === 'days') return 'day'
  if (value === 'weeks') return 'week'
  if (value === 'months') return 'month'
  if (value === 'years') return 'year'
  return value
}

function formatPlanValidityDuration(value: number, unit: string | undefined): string {
  const normalized = normalizePlanValidityUnit(unit)
  if (normalized === 'month') return t('payment.planCard.durationMonths', { count: value })
  if (normalized === 'week') return t('payment.planCard.durationWeeks', { count: value })
  if (normalized === 'year') return t('payment.planCard.durationYears', { count: value })
  return t('payment.planCard.durationDays', { count: value })
}

const selectedPlanHasSavings = computed(() => {
  if (!selectedPlan.value) return false
  return selectedPlanSavingsComparePrice.value > selectedPlan.value.price
})

const selectedPlanDiscountText = computed(() => {
  if (!selectedPlanHasSavings.value || !selectedPlan.value) return ''
  const pct = Math.round((1 - selectedPlan.value.price / selectedPlanSavingsComparePrice.value) * 100)
  return pct > 0 ? t('payment.planCard.discountPercent', { percent: pct }) : ''
})

const selectedPlanSavingsText = computed(() => {
  if (!selectedPlanHasSavings.value || !selectedPlan.value) return ''
  return t('payment.planCard.saveAmount', {
    amount: formatSelectedPlanAmount(selectedPlanSavingsComparePrice.value - selectedPlan.value.price),
  })
})

const selectedPlanWorkdayFriendly = computed(() => {
  const daily = selectedPlan.value?.daily_limit_usd
  const weekly = selectedPlan.value?.weekly_limit_usd
  if (daily == null || weekly == null || daily <= 0 || weekly <= 0) return false
  return true
})

function selectPlan(plan: SubscriptionPlan) {
  selectedPlan.value = plan
  errorMessage.value = ''
}

function selectPlanFromModal(plan: SubscriptionPlan) {
  showRenewalModal.value = false
  renewGroupId.value = null
  selectedPlan.value = plan
  errorMessage.value = ''
}

function closeRenewalModal() {
  showRenewalModal.value = false
  renewGroupId.value = null
}

async function handleSubmitRecharge() {
  if (!canSubmit.value || submitting.value) return
  const selectedPaymentCurrency = normalizePaymentCurrency(selectedRechargeLimit.value?.currency).toUpperCase()
  await createOrder(validAmount.value, 'balance', undefined, {
    paymentType: selectedRechargeMethod.value,
    paymentAmount: selectedPaymentCurrency === 'USDT' ? usdtPaymentAmount.value : undefined,
  })
}

async function confirmSubscribe() {
  if (!selectedPlan.value || submitting.value) return
  const selectedPaymentCurrency = selectedSubscriptionCurrency.value.toUpperCase()
  await createOrder(selectedPlan.value.price, 'subscription', selectedPlan.value.id, {
    paymentType: selectedSubscriptionMethod.value,
    paymentAmount: selectedPaymentCurrency === 'USDT' ? subscriptionBaseAmount.value : undefined,
    targetUserEmail: selectedSubscriptionMethod.value === 'balance' ? subscriptionTargetEmailTrimmed.value : undefined,
  })
}

async function createOrder(orderAmount: number, orderType: OrderType, planId?: number, options: CreateOrderOptions = {}) {
  submitting.value = true
  errorMessage.value = ''
  errorHintMessage.value = ''
  const fallbackMethod = orderType === 'subscription' ? selectedSubscriptionMethod.value : selectedRechargeMethod.value
  const requestType = normalizeVisibleMethod(options.paymentType || fallbackMethod) || options.paymentType || fallbackMethod
  try {
    const payload = buildCreateOrderPayload({
      amount: orderAmount,
      paymentAmount: options.paymentAmount,
      paymentType: requestType,
      orderType,
      planId,
      targetUserEmail: options.targetUserEmail,
      origin: typeof window !== 'undefined' ? window.location.origin : '',
      isMobile: isMobileDevice(),
      isWechatBrowser: typeof window !== 'undefined' && /MicroMessenger/i.test(window.navigator.userAgent),
      forceQRCode: !!(checkout.value.alipay_force_qrcode && normalizeVisibleMethod(requestType) === 'alipay'),
    })
    if (options.openid) {
      payload.openid = options.openid
    }
    if (options.wechatResumeToken) {
      payload.wechat_resume_token = options.wechatResumeToken
    }

    const result = await paymentStore.createOrder(payload) as CreateOrderResult & { resume_token?: string }
    const openWindow = (url: string) => {
      const win = window.open(url, 'paymentPopup', getPaymentPopupFeatures())
      if (!win || win.closed) {
        window.location.href = url
      }
    }
    const visibleMethod = normalizeVisibleMethod(requestType) || requestType
    // When user clicks the dedicated Stripe button, leave method blank so the
    // landing page renders Stripe's full Payment Element (card/link/alipay/wxpay).
    const stripeMethod = visibleMethod === 'stripe'
      ? ''
      : visibleMethod === 'wxpay' ? 'wechat_pay' : 'alipay'
    const stripeRouteUrl = result.client_secret && visibleMethod !== 'airwallex'
      ? router.resolve({
        path: '/payment/stripe',
        query: {
          order_id: String(result.order_id),
          client_secret: result.client_secret,
          method: stripeMethod || undefined,
          resume_token: result.resume_token || undefined,
        },
      }).href
      : ''
    const airwallexRouteUrl = result.client_secret && result.intent_id
      ? router.resolve({
        path: '/payment/airwallex',
        query: {
          order_id: String(result.order_id),
          out_trade_no: result.out_trade_no || undefined,
          resume_token: result.resume_token || undefined,
        },
      }).href
      : ''
    const decision = decidePaymentLaunch(result, {
      visibleMethod,
      orderType,
      isMobile: isMobileDevice(),
      isWechatBrowser: typeof window !== 'undefined' && /MicroMessenger/i.test(window.navigator.userAgent),
      forceQRCode: !!(checkout.value.alipay_force_qrcode && visibleMethod === 'alipay'),
      stripePopupUrl: stripeRouteUrl,
      stripeRouteUrl,
      airwallexRouteUrl,
    })

    if (visibleMethod === 'balance' && String(result.status || '').toUpperCase() === 'COMPLETED') {
      removeRecoverySnapshot()
      authStore.refreshUser()
      if (orderType === 'subscription') {
        selectedPlan.value = null
        subscriptionStore.fetchActiveSubscriptions(true).catch(() => {})
      }
      appStore.showSuccess(t('payment.balanceSubscriptionSuccess'))
      return
    }

    if (decision.kind === 'wechat_oauth' && decision.oauth?.authorize_url) {
      window.location.href = buildWechatOAuthAuthorizeUrl(decision.oauth.authorize_url, {
        paymentType: visibleMethod,
        orderType,
        planId,
        orderAmount,
      })
      return
    }

    if (decision.kind === 'unhandled') {
      applyScenarioError({ reason: 'UNHANDLED_PAYMENT_SCENARIO' }, visibleMethod)
      return
    }

    paymentState.value = decision.paymentState
    paymentPhase.value = 'paying'
    persistRecoverySnapshot(decision.recovery)

    if (decision.kind === 'stripe_popup') {
      openWindow(decision.paymentState.payUrl)
      return
    }
    if (decision.kind === 'stripe_route') {
      window.location.href = decision.paymentState.payUrl
      return
    }
    if (decision.kind === 'airwallex_route') {
      window.location.href = decision.paymentState.payUrl
      return
    }
    if (decision.kind === 'wechat_jsapi' && decision.jsapi) {
      try {
        const jsapiResult = await invokeWechatJsapiPayment(decision.jsapi as Record<string, unknown>)
        const errMsg = String(jsapiResult.err_msg || '').toLowerCase()
        if (errMsg.includes('cancel')) {
          appStore.showInfo(t('payment.qr.cancelled'))
          resetPayment()
        } else if (errMsg && !errMsg.includes('ok')) {
          resetPayment()
          const fallbackApplied = await attemptMobileQrFallback(
            { reason: 'WECHAT_JSAPI_FAILED', message: errMsg },
            {
              orderAmount,
              paymentAmount: options.paymentAmount,
              orderType,
              planId,
              paymentType: visibleMethod,
              attempted: options.mobileQrFallbackAttempted === true,
            },
          )
          if (!fallbackApplied) {
            applyScenarioError({ reason: 'WECHAT_JSAPI_FAILED', message: errMsg }, visibleMethod)
          }
        } else {
          const resultState = { ...decision.paymentState }
          resetPayment()
          await redirectToPaymentResult(resultState)
        }
      } catch (err: unknown) {
        resetPayment()
        const fallbackApplied = await attemptMobileQrFallback(err, {
          orderAmount,
          paymentAmount: options.paymentAmount,
          orderType,
          planId,
          paymentType: visibleMethod,
          attempted: options.mobileQrFallbackAttempted === true,
        })
        if (!fallbackApplied) {
          throw err
        }
      }
      return
    }
    if (decision.kind === 'redirect_waiting' && decision.paymentState.payUrl) {
      if (isMobileDevice()) {
        window.location.href = decision.paymentState.payUrl
        return
      }
      if (visibleMethod === 'usdt') {
        window.location.href = decision.paymentState.payUrl
        return
      }
      openWindow(decision.paymentState.payUrl)
    }
  } catch (err: unknown) {
    const apiErr = err as Record<string, unknown>
    if (apiErr.reason === 'TOO_MANY_PENDING') {
      const metadata = apiErr.metadata as Record<string, unknown> | undefined
      errorMessage.value = t('payment.errors.tooManyPending', { max: metadata?.max || '' })
      errorHintMessage.value = ''
    } else if (apiErr.reason === 'CANCEL_RATE_LIMITED') {
      errorMessage.value = t('payment.errors.cancelRateLimited')
      errorHintMessage.value = ''
    } else if (await attemptMobileQrFallback(err, {
      orderAmount,
      paymentAmount: options.paymentAmount,
      orderType,
      planId,
      paymentType: requestType,
      attempted: options.mobileQrFallbackAttempted === true,
    })) {
      return
    } else {
      const handled = applyScenarioError(
        err,
        normalizeVisibleMethod(options.paymentType || fallbackMethod) || fallbackMethod,
      )
      if (!handled) {
        errorMessage.value = extractI18nErrorMessage(err, t, 'payment.errors', extractApiErrorMessage(err, t('payment.result.failed')))
        errorHintMessage.value = ''
      }
      if (handled) {
        return
      }
    }
    appStore.showError(buildPaymentErrorToastMessage(errorMessage.value, errorHintMessage.value))
  } finally {
    submitting.value = false
  }
}

interface MobileQrFallbackContext {
  orderAmount: number
  paymentAmount?: number
  orderType: OrderType
  planId?: number
  paymentType: string
  attempted: boolean
}

function shouldFallbackToDesktopQr(err: unknown, paymentMethod: string, attempted: boolean): boolean {
  if (attempted || !isMobileDevice()) {
    return false
  }

  const normalizedMethod = normalizeVisibleMethod(paymentMethod) || paymentMethod
  const reason = typeof err === 'object' && err && 'reason' in err && typeof err.reason === 'string'
    ? err.reason
    : ''
  const message = err instanceof Error
    ? err.message
    : (typeof err === 'object' && err && 'message' in err && typeof err.message === 'string'
      ? err.message
      : '')
  const normalizedMessage = message.toLowerCase()

  if (normalizedMethod === 'wxpay') {
    return reason === 'WECHAT_H5_NOT_AUTHORIZED'
      || reason === 'WECHAT_PAYMENT_MP_NOT_CONFIGURED'
      || reason === 'WECHAT_JSAPI_FAILED'
      || reason === 'PAYMENT_GATEWAY_ERROR'
      || reason === 'UNHANDLED_PAYMENT_SCENARIO'
      || normalizedMessage.includes('weixinjsbridge is unavailable')
      || normalizedMessage.includes('wechat_jsapi_unavailable')
  }

  if (normalizedMethod === 'alipay') {
    return reason === 'PAYMENT_GATEWAY_ERROR' || reason === 'UNHANDLED_PAYMENT_SCENARIO'
  }

  return false
}

async function attemptMobileQrFallback(err: unknown, context: MobileQrFallbackContext): Promise<boolean> {
  if (!shouldFallbackToDesktopQr(err, context.paymentType, context.attempted)) {
    return false
  }

  try {
    const visibleMethod = normalizeVisibleMethod(context.paymentType) || context.paymentType
    const payload = buildCreateOrderPayload({
      amount: context.orderAmount,
      paymentAmount: context.paymentAmount,
      paymentType: visibleMethod,
      orderType: context.orderType,
      planId: context.planId,
      origin: typeof window !== 'undefined' ? window.location.origin : '',
      isMobile: false,
      isWechatBrowser: false,
    })
    const result = await paymentStore.createOrder(payload) as CreateOrderResult & { resume_token?: string }
    const stripeMethod = visibleMethod === 'wxpay' ? 'wechat_pay' : 'alipay'
    const stripeRouteUrl = result.client_secret
      ? router.resolve({
        path: '/payment/stripe',
        query: {
          order_id: String(result.order_id),
          client_secret: result.client_secret,
          method: stripeMethod,
          resume_token: result.resume_token || undefined,
        },
      }).href
      : ''
    const decision = decidePaymentLaunch(result, {
      visibleMethod,
      orderType: context.orderType,
      isMobile: false,
      isWechatBrowser: false,
      stripePopupUrl: stripeRouteUrl,
      stripeRouteUrl,
    })

    if (decision.kind !== 'qr_waiting' || !decision.paymentState.qrCode) {
      return false
    }

    errorMessage.value = ''
    errorHintMessage.value = ''
    paymentState.value = decision.paymentState
    paymentPhase.value = 'paying'
    persistRecoverySnapshot(decision.recovery)
    appStore.showWarning(t('payment.errors.mobilePaymentFallbackToQr'))
    return true
  } catch {
    return false
  }
}

function applyScenarioError(err: unknown, paymentMethod: string): boolean {
  const descriptor = describePaymentScenarioError(err, {
    paymentMethod,
    isMobile: isMobileDevice(),
    isWechatBrowser: typeof window !== 'undefined' && /MicroMessenger/i.test(window.navigator.userAgent),
  })
  if (!descriptor) {
    errorMessage.value = ''
    errorHintMessage.value = ''
    return false
  }
  errorMessage.value = t(descriptor.messageKey)
  errorHintMessage.value = descriptor.hintKey ? t(descriptor.hintKey) : ''
  appStore.showError(buildPaymentErrorToastMessage(errorMessage.value, errorHintMessage.value))
  return true
}

async function resumeWechatPaymentFromQuery() {
  const resume = parseWechatResumeRoute(route.query, checkout.value.plans, validAmount.value)
  if (!resume) {
    return
  }

  if (resume.orderType === 'subscription') {
    selectedSubscriptionMethod.value = resume.paymentType
  } else {
    selectedRechargeMethod.value = resume.paymentType
  }
  if (resume.orderType === 'balance' && resume.orderAmount > 0) {
    amount.value = resume.orderAmount
  }
  if (resume.orderType === 'subscription' && resume.planId) {
    selectedPlan.value = checkout.value.plans.find(plan => plan.id === resume.planId) ?? null
  }

  await router.replace({ path: route.path, query: stripWechatResumeQuery(route.query) })

  if (resume.wechatResumeToken) {
    await createOrder(0, resume.orderType, resume.planId, {
      wechatResumeToken: resume.wechatResumeToken,
      paymentType: resume.paymentType,
      isResume: true,
    })
    return
  }

  if (resume.orderAmount > 0 && resume.openid) {
    await createOrder(resume.orderAmount, resume.orderType, resume.planId, {
      openid: resume.openid,
      paymentType: resume.paymentType,
      isResume: true,
    })
  }
}

onMounted(async () => {
  try {
    const res = await paymentAPI.getCheckoutInfo()
    checkout.value = res.data
    selectedRechargeMethod.value = firstSortedMethod(rechargeMethods.value) || firstSortedMethod(enabledMethods.value)
    selectedSubscriptionMethod.value = firstSortedMethod(subscriptionMethods.value)
    if (typeof window !== 'undefined') {
      if (hasWechatResumeQuery(route.query)) {
        removeRecoverySnapshot()
      }
      const routeResumeToken = typeof route.query.resume_token === 'string'
        ? route.query.resume_token
        : typeof route.query.wechat_resume_token === 'string'
          ? route.query.wechat_resume_token
          : undefined
      const restored = readPaymentRecoverySnapshot(
        window.localStorage.getItem(PAYMENT_RECOVERY_STORAGE_KEY),
        { resumeToken: routeResumeToken },
      )
      if (restored) {
        paymentState.value = restored
        paymentPhase.value = 'paying'
        const restoredMethod = normalizeVisibleMethod(restored.paymentType)
        if (restoredMethod) {
          if (restored.orderType === 'subscription') {
            selectedSubscriptionMethod.value = restoredMethod
          } else {
            selectedRechargeMethod.value = restoredMethod
          }
        }
      } else {
        removeRecoverySnapshot()
      }
    }
    await resumeWechatPaymentFromQuery()
    if (checkout.value.balance_disabled) {
      activeTab.value = 'subscription'
    }
    // Handle renewal navigation: ?tab=subscription&group=123
    if (route.query.tab === 'subscription') {
      activeTab.value = 'subscription'
      if (typeof route.query.target_user === 'string') {
        subscriptionTargetEmail.value = route.query.target_user
      } else if (typeof route.query.target_email === 'string') {
        subscriptionTargetEmail.value = route.query.target_email
      }
      if (subscriptionTargetEmail.value) {
        subscriptionTargetKeyword.value = subscriptionTargetEmail.value
      }
      if (subscriptionTargetEmail.value && balanceSubscriptionAllowed.value) {
        selectedSubscriptionMethod.value = 'balance'
      }
      if (route.query.group) {
        const groupId = Number(route.query.group)
        const groupPlans = checkout.value.plans.filter(p => p.group_id === groupId)
        if (groupPlans.length === 1) {
          selectedPlan.value = groupPlans[0]
        } else if (groupPlans.length > 1) {
          renewGroupId.value = groupId
          showRenewalModal.value = true
        }
      }
    }
  } catch (err: unknown) { appStore.showError(extractI18nErrorMessage(err, t, 'payment.errors', t('common.error'))) }
  finally { loading.value = false }
  // Fetch active subscriptions (uses cache, non-blocking)
  subscriptionStore.fetchActiveSubscriptions().catch(() => {})
})
</script>
