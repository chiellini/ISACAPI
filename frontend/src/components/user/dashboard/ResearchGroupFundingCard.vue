<template>
  <section data-testid="research-group-funding-card" class="card overflow-hidden border border-indigo-100 dark:border-indigo-900/40">
    <div v-if="isPending" class="flex flex-col gap-4 bg-indigo-50/70 p-5 dark:bg-indigo-950/20 sm:flex-row sm:items-center sm:justify-between">
      <div>
        <p class="text-sm font-semibold text-indigo-900 dark:text-indigo-100">{{ t('researchGroup.dashboard.invitationTitle') }}</p>
        <p class="mt-1 text-sm text-indigo-700 dark:text-indigo-300">
          {{ t('researchGroup.dashboard.invitationDescription', { group: context.group.name, owner: context.group.owner_username || context.group.owner_email }) }}
        </p>
      </div>
      <div class="flex shrink-0 gap-2">
        <button type="button" class="btn btn-secondary" :disabled="busy" @click="respond(false)">
          {{ t('researchGroup.invitations.reject') }}
        </button>
        <button type="button" class="btn btn-primary" :disabled="busy" @click="respond(true)">
          {{ busy ? t('common.processing') : t('researchGroup.invitations.accept') }}
        </button>
      </div>
    </div>

    <div v-else class="grid gap-5 p-5 md:grid-cols-2">
      <div>
        <div class="flex items-start justify-between gap-3">
          <div>
            <p class="text-xs font-semibold uppercase tracking-wide text-indigo-500">{{ t('researchGroup.dashboard.groupFunding') }}</p>
            <h2 class="mt-1 text-lg font-semibold text-gray-900 dark:text-white">{{ context.group.name }}</h2>
          </div>
          <span class="rounded-full px-2.5 py-1 text-xs font-medium" :class="isPaused ? 'bg-amber-100 text-amber-700 dark:bg-amber-900/30 dark:text-amber-300' : 'bg-emerald-100 text-emerald-700 dark:bg-emerald-900/30 dark:text-emerald-300'">
            {{ t(`researchGroup.status.${member.status}`) }}
          </span>
        </div>

        <div class="mt-4 flex items-end justify-between gap-3">
          <div>
            <p class="text-sm text-gray-500 dark:text-gray-400">{{ t('researchGroup.dashboard.monthlyRemaining') }}</p>
            <p class="text-2xl font-bold text-indigo-600 dark:text-indigo-400">{{ formatUsd(member.monthly_remaining_usd) }}</p>
          </div>
          <p class="text-right text-xs text-gray-500 dark:text-gray-400">
            {{ t('researchGroup.dashboard.usedOfLimit', { used: formatUsd(member.monthly_usage_usd), limit: formatUsd(member.monthly_limit_usd) }) }}
          </p>
        </div>
        <div class="mt-2 h-2 overflow-hidden rounded-full bg-gray-100 dark:bg-dark-700">
          <div class="h-full rounded-full bg-indigo-500 transition-all" :style="{ width: `${usagePercent}%` }" />
        </div>
        <p class="mt-2 text-xs text-gray-500 dark:text-gray-400">{{ t('researchGroup.dashboard.resetsAt', { date: formatDate(member.resets_at) }) }}</p>
      </div>

      <div class="rounded-xl bg-gray-50 p-4 dark:bg-dark-800/70">
        <p class="text-xs font-semibold uppercase tracking-wide text-gray-500 dark:text-gray-400">{{ t('researchGroup.dashboard.personalBalance') }}</p>
        <p class="mt-1 text-2xl font-bold text-emerald-600 dark:text-emerald-400">{{ formatUsd(personalBalance) }}</p>
        <p class="mt-3 text-sm text-gray-600 dark:text-gray-300">{{ isPaused ? t('researchGroup.dashboard.pausedFallback') : t('researchGroup.dashboard.fundingPriority') }}</p>
        <RouterLink to="/research-group" class="mt-3 inline-flex text-sm font-medium text-primary-600 hover:text-primary-700 dark:text-primary-400">
          {{ t('researchGroup.dashboard.viewDetails') }}
        </RouterLink>
      </div>
    </div>
  </section>
</template>

<script setup lang="ts">
import { computed, ref } from 'vue'
import { useI18n } from 'vue-i18n'
import { RouterLink } from 'vue-router'
import researchGroupAPI from '@/api/researchGroup'
import { useAppStore } from '@/stores/app'
import { useAuthStore } from '@/stores/auth'
import { extractApiErrorMessage } from '@/utils/apiError'
import type { ResearchGroupContext } from '@/types'

const props = defineProps<{ context: ResearchGroupContext; personalBalance: number }>()
const emit = defineEmits<{ (event: 'updated'): void }>()
const { t } = useI18n()
const appStore = useAppStore()
const authStore = useAuthStore()
const busy = ref(false)

const member = computed(() => props.context.member!)
const isPending = computed(() => member.value?.status === 'pending')
const isPaused = computed(() => member.value?.status === 'paused' || props.context.group.status === 'paused')
const usagePercent = computed(() => {
  const limit = member.value?.monthly_limit_usd ?? 0
  if (limit <= 0) return 0
  const committed = (member.value?.monthly_usage_usd ?? 0) + (member.value?.monthly_reserved_usd ?? 0)
  return Math.min(100, Math.max(0, (committed / limit) * 100))
})

function formatUsd(value: number): string {
  return new Intl.NumberFormat(undefined, { style: 'currency', currency: 'USD', minimumFractionDigits: 2, maximumFractionDigits: 4 }).format(value || 0)
}

function formatDate(value: string): string {
  if (!value) return '-'
  const date = new Date(value)
  return Number.isNaN(date.getTime()) ? value : date.toLocaleString()
}

async function respond(accept: boolean): Promise<void> {
  const invitationId = member.value?.id
  if (!invitationId || busy.value) return
  busy.value = true
  try {
    if (accept) {
      await researchGroupAPI.acceptInvitation(invitationId)
      appStore.showSuccess(t('researchGroup.invitations.accepted'))
    } else {
      await researchGroupAPI.rejectInvitation(invitationId)
      appStore.showSuccess(t('researchGroup.invitations.rejected'))
    }
    await authStore.refreshUser()
    emit('updated')
  } catch (error) {
    appStore.showError(extractApiErrorMessage(error, t('researchGroup.errors.operationFailed')))
  } finally {
    busy.value = false
  }
}
</script>
