<template>
  <article
    class="overflow-hidden rounded-2xl border border-slate-200 bg-white shadow-sm dark:border-dark-700 dark:bg-dark-900"
  >
    <div
      :class="[
        'flex flex-col gap-3 border-b border-slate-200 bg-slate-50/80 dark:border-dark-700 dark:bg-dark-950/60 sm:flex-row sm:items-start sm:justify-between',
        compact ? 'px-4 py-4' : 'px-5 py-5 md:px-6',
      ]"
    >
      <div>
        <h3 :class="['font-bold text-slate-950 dark:text-white', compact ? 'text-base' : 'text-lg']">
          {{ t('home.pricing.comparisonTitle') }}
        </h3>
        <p class="mt-1 max-w-3xl text-xs leading-5 text-slate-500 dark:text-dark-400 sm:text-sm">
          {{ t('home.pricing.comparisonDescription') }}
        </p>
      </div>
      <span
        class="inline-flex shrink-0 self-start rounded-full border border-emerald-200 bg-emerald-50 px-3 py-1.5 text-xs font-semibold text-emerald-700 dark:border-emerald-900/60 dark:bg-emerald-950/30 dark:text-emerald-300"
      >
        {{ t('home.pricing.effectiveRateBadge', { usd: formatCompactNumber(normalizedUsdPerCny) }) }}
      </span>
    </div>

    <div class="overflow-x-auto">
      <table class="w-full min-w-[820px] border-collapse text-left text-xs sm:text-sm">
        <thead class="bg-slate-50 text-[11px] uppercase tracking-wide text-slate-500 dark:bg-dark-950/40 dark:text-dark-400">
          <tr>
            <th class="px-4 py-3 font-semibold md:px-5">{{ t('home.pricing.modelHeader') }}</th>
            <th class="px-4 py-3 text-right font-semibold">{{ t('home.pricing.benchmarkInputHeader') }}</th>
            <th class="px-4 py-3 text-right font-semibold">{{ t('home.pricing.benchmarkOutputHeader') }}</th>
            <th class="bg-emerald-50/70 px-4 py-3 text-right font-semibold text-emerald-700 dark:bg-emerald-950/20 dark:text-emerald-300">
              {{ t('home.pricing.effectiveInputHeader') }}
            </th>
            <th class="bg-emerald-50/70 px-4 py-3 text-right font-semibold text-emerald-700 dark:bg-emerald-950/20 dark:text-emerald-300 md:px-5">
              {{ t('home.pricing.effectiveOutputHeader') }}
            </th>
          </tr>
        </thead>
        <tbody class="divide-y divide-slate-100 dark:divide-dark-800">
          <tr
            v-for="model in PUBLIC_MODEL_PRICES"
            :key="model.id"
            class="transition-colors hover:bg-slate-50/80 dark:hover:bg-dark-800/50"
          >
            <td class="whitespace-nowrap px-4 py-3.5 font-semibold text-slate-900 dark:text-white md:px-5">
              <span class="inline-flex items-center gap-2">
                <span
                  :class="[
                    'h-2 w-2 rounded-full',
                    model.family === 'gpt' ? 'bg-emerald-500' : 'bg-orange-500',
                  ]"
                ></span>
                {{ model.name }}
              </span>
            </td>
            <td class="whitespace-nowrap px-4 py-3.5 text-right font-mono text-slate-600 dark:text-dark-300">
              {{ formatUsd(model.benchmarkInputUsdPerMillion) }}
            </td>
            <td class="whitespace-nowrap px-4 py-3.5 text-right font-mono text-slate-600 dark:text-dark-300">
              {{ formatUsd(model.benchmarkOutputUsdPerMillion) }}
            </td>
            <td class="whitespace-nowrap bg-emerald-50/40 px-4 py-3.5 text-right font-mono font-semibold text-emerald-700 dark:bg-emerald-950/10 dark:text-emerald-300">
              {{ formatCny(model.benchmarkInputUsdPerMillion) }}
            </td>
            <td class="whitespace-nowrap bg-emerald-50/40 px-4 py-3.5 text-right font-mono font-semibold text-emerald-700 dark:bg-emerald-950/10 dark:text-emerald-300 md:px-5">
              {{ formatCny(model.benchmarkOutputUsdPerMillion) }}
            </td>
          </tr>
        </tbody>
      </table>
    </div>

    <div
      :class="[
        'border-t border-slate-200 bg-slate-50/70 text-xs leading-5 text-slate-500 dark:border-dark-700 dark:bg-dark-950/40 dark:text-dark-400',
        compact ? 'px-4 py-3' : 'px-5 py-4 md:px-6',
      ]"
    >
      <p class="font-medium text-slate-700 dark:text-dark-200">
        {{
          t('home.pricing.formula', {
            divisor: INTERNAL_TOKEN_PRICE_DIVISOR,
            usd: formatCompactNumber(normalizedUsdPerCny),
            total: formatCompactNumber(INTERNAL_TOKEN_PRICE_DIVISOR * normalizedUsdPerCny),
          })
        }}
      </p>
      <p class="mt-1">{{ t('home.pricing.disclaimer') }}</p>
    </div>
  </article>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import { useI18n } from 'vue-i18n'
import {
  INTERNAL_TOKEN_PRICE_DIVISOR,
  PUBLIC_MODEL_PRICES,
  PUBLIC_RECHARGE_USD_PER_CNY,
  benchmarkUsdToEffectiveCny,
  formatCompactNumber,
} from '@/utils/pricing'

const props = withDefaults(
  defineProps<{
    usdPerCny?: number
    compact?: boolean
  }>(),
  {
    usdPerCny: PUBLIC_RECHARGE_USD_PER_CNY,
    compact: false,
  },
)

const { t } = useI18n()

const normalizedUsdPerCny = computed(() =>
  Number.isFinite(props.usdPerCny) && props.usdPerCny > 0
    ? props.usdPerCny
    : PUBLIC_RECHARGE_USD_PER_CNY,
)

function formatUsd(value: number): string {
  return `$${formatCompactNumber(value, 2)}`
}

function formatCny(benchmarkUsd: number): string {
  return `¥${formatCompactNumber(
    benchmarkUsdToEffectiveCny(benchmarkUsd, normalizedUsdPerCny.value),
    6,
  )}`
}
</script>
