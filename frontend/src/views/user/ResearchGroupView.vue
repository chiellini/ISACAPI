<template>
  <AppLayout>
    <div class="space-y-6" data-testid="research-group-view">
      <div v-if="loading" class="flex justify-center py-16">
        <LoadingSpinner />
      </div>

      <template v-else-if="isOwner && context">
        <section class="card p-6">
          <div class="flex flex-col gap-4 lg:flex-row lg:items-start lg:justify-between">
            <div>
              <div class="flex flex-wrap items-center gap-2">
                <h1 class="text-2xl font-semibold text-gray-900 dark:text-white">{{ context.group.name }}</h1>
                <StatusPill :status="context.group.status" />
              </div>
              <p class="mt-1 text-sm text-gray-500 dark:text-gray-400">{{ t('researchGroup.owner.description') }}</p>
            </div>
            <div class="flex flex-wrap gap-2">
              <button class="btn btn-secondary" :disabled="working" @click="toggleGroupStatus">
                {{ context.group.status === 'active' ? t('researchGroup.actions.pauseGroup') : t('researchGroup.actions.resumeGroup') }}
              </button>
              <button class="btn btn-danger" :disabled="working" @click="requestDissolve">
                {{ t('researchGroup.actions.dissolve') }}
              </button>
            </div>
          </div>

          <form class="mt-5 flex max-w-xl gap-2" @submit.prevent="saveGroupName">
            <input v-model.trim="groupName" class="input flex-1" maxlength="100" :placeholder="t('researchGroup.fields.groupName')" />
            <button class="btn btn-primary" :disabled="working || !groupName || groupName === context.group.name">
              {{ t('common.save') }}
            </button>
          </form>
        </section>

        <section class="grid gap-4 sm:grid-cols-2 xl:grid-cols-4">
          <StatCard :label="t('researchGroup.stats.sharedBalance')" :value="formatUsd(context.group.owner_balance ?? 0)" tone="emerald" />
          <StatCard :label="t('researchGroup.stats.monthFunding')" :value="formatUsd(context.usage_summary?.total_cost_usd || 0)" tone="indigo" />
          <StatCard :label="t('researchGroup.stats.activeMembers')" :value="String(context.usage_summary?.active_member_count || 0)" tone="blue" />
          <StatCard :label="t('researchGroup.stats.fundedRequests')" :value="formatInteger(context.usage_summary?.request_count || 0)" tone="purple" />
        </section>

        <section class="card p-6">
          <h2 class="text-lg font-semibold text-gray-900 dark:text-white">{{ t('researchGroup.invite.title') }}</h2>
          <p class="mt-1 text-sm text-gray-500 dark:text-gray-400">{{ t('researchGroup.invite.description') }}</p>
          <form class="mt-4 grid gap-3 md:grid-cols-[minmax(0,1fr)_180px_auto]" @submit.prevent="inviteMember">
            <input v-model.trim="inviteEmail" type="email" class="input" required :placeholder="t('researchGroup.fields.memberEmail')" />
            <div class="relative">
              <span class="pointer-events-none absolute inset-y-0 left-3 flex items-center text-sm text-gray-400">$</span>
              <input v-model.number="inviteLimit" type="number" min="0" step="0.01" class="input w-full pl-7" required :aria-label="t('researchGroup.fields.monthlyLimit')" />
            </div>
            <button class="btn btn-primary" :disabled="working || !inviteEmail || inviteLimit < 0">{{ t('researchGroup.invite.send') }}</button>
          </form>
          <p class="mt-2 text-xs text-gray-500 dark:text-gray-400">{{ t('researchGroup.invite.zeroLimitHint') }}</p>
        </section>

        <section class="card overflow-hidden">
          <div class="border-b border-gray-200 p-6 dark:border-dark-700">
            <h2 class="text-lg font-semibold text-gray-900 dark:text-white">{{ t('researchGroup.members.title') }}</h2>
            <p class="mt-1 text-sm text-gray-500 dark:text-gray-400">{{ t('researchGroup.members.description') }}</p>
          </div>
          <div v-if="members.length === 0" class="p-10 text-center text-sm text-gray-500 dark:text-gray-400">{{ t('researchGroup.members.empty') }}</div>
          <div v-else class="overflow-x-auto">
            <table class="w-full min-w-[980px] text-left text-sm">
              <thead class="bg-gray-50 text-xs uppercase tracking-wide text-gray-500 dark:bg-dark-800 dark:text-gray-400">
                <tr>
                  <th class="px-5 py-3 font-medium">{{ t('researchGroup.members.member') }}</th>
                  <th class="px-5 py-3 font-medium">{{ t('common.status') }}</th>
                  <th class="px-5 py-3 font-medium">{{ t('researchGroup.fields.monthlyLimit') }}</th>
                  <th class="px-5 py-3 font-medium">{{ t('researchGroup.members.usage') }}</th>
                  <th class="px-5 py-3 font-medium">{{ t('researchGroup.members.resetAt') }}</th>
                  <th class="px-5 py-3 text-right font-medium">{{ t('common.actions') }}</th>
                </tr>
              </thead>
              <tbody class="divide-y divide-gray-100 dark:divide-dark-700">
                <tr v-for="member in members" :key="member.id" class="text-gray-700 dark:text-gray-300">
                  <td class="px-5 py-4">
                    <p class="font-medium text-gray-900 dark:text-white">{{ member.username || '-' }}</p>
                    <p class="text-xs text-gray-500">{{ member.email }}</p>
                  </td>
                  <td class="px-5 py-4"><StatusPill :status="member.status" /></td>
                  <td class="px-5 py-4">
                    <div class="flex items-center gap-2">
                      <input v-model.number="limitDrafts[member.id]" type="number" min="0" step="0.01" class="input w-28" :aria-label="t('researchGroup.fields.monthlyLimit')" />
                      <button class="btn btn-ghost btn-sm" :disabled="working || !limitChanged(member)" @click="saveMemberLimit(member)">{{ t('common.save') }}</button>
                    </div>
                  </td>
                  <td class="px-5 py-4">
                    <p class="font-medium">{{ formatUsd(member.monthly_usage_usd) }} / {{ formatUsd(member.monthly_limit_usd) }}</p>
                    <p v-if="member.monthly_reserved_usd > 0" class="text-xs text-amber-600 dark:text-amber-400">{{ t('researchGroup.members.reserved', { amount: formatUsd(member.monthly_reserved_usd) }) }}</p>
                  </td>
                  <td class="px-5 py-4 text-xs">{{ formatDate(member.resets_at) }}</td>
                  <td class="px-5 py-4">
                    <div class="flex justify-end gap-1">
                      <button v-if="member.status !== 'pending'" class="btn btn-ghost btn-sm" :disabled="working" @click="setMemberStatus(member, member.status === 'active' ? 'paused' : 'active')">
                        {{ member.status === 'active' ? t('researchGroup.actions.pause') : t('researchGroup.actions.resume') }}
                      </button>
                      <button v-if="member.status !== 'pending'" class="btn btn-ghost btn-sm" :disabled="working" @click="requestMemberReset(member)">{{ t('common.reset') }}</button>
                      <button class="btn btn-ghost btn-sm text-red-600 hover:text-red-700" :disabled="working" @click="requestMemberRemoval(member)">{{ t('researchGroup.actions.remove') }}</button>
                    </div>
                  </td>
                </tr>
              </tbody>
            </table>
          </div>
        </section>

        <section class="card overflow-hidden">
          <div class="border-b border-gray-200 p-6 dark:border-dark-700">
            <h2 class="text-lg font-semibold text-gray-900 dark:text-white">{{ t('researchGroup.usage.title') }}</h2>
            <p class="mt-1 text-sm text-gray-500 dark:text-gray-400">{{ t('researchGroup.usage.description') }}</p>
            <div class="mt-4 grid gap-3 sm:grid-cols-2 lg:grid-cols-4">
              <select v-model="usageMemberId" class="input" @change="reloadUsage">
                <option value="">{{ t('researchGroup.usage.allMembers') }}</option>
                <option v-for="member in members" :key="member.id" :value="String(member.id)">{{ member.username || member.email }}</option>
              </select>
              <input v-model="usageStart" type="date" class="input" :aria-label="t('researchGroup.usage.startDate')" />
              <input v-model="usageEnd" type="date" class="input" :aria-label="t('researchGroup.usage.endDate')" />
              <button class="btn btn-secondary" :disabled="usageLoading" @click="reloadUsage">{{ t('common.search') }}</button>
            </div>
          </div>
          <div v-if="usageLoading" class="flex justify-center p-10"><LoadingSpinner /></div>
          <div v-else-if="usage.items.length === 0" class="p-10 text-center text-sm text-gray-500 dark:text-gray-400">{{ t('researchGroup.usage.empty') }}</div>
          <div v-else class="overflow-x-auto">
            <table class="w-full min-w-[760px] text-left text-sm">
              <thead class="bg-gray-50 text-xs uppercase tracking-wide text-gray-500 dark:bg-dark-800 dark:text-gray-400">
                <tr>
                  <th class="px-5 py-3 font-medium">{{ t('researchGroup.members.member') }}</th>
                  <th class="px-5 py-3 font-medium">{{ t('researchGroup.usage.model') }}</th>
                  <th class="px-5 py-3 font-medium">{{ t('researchGroup.usage.requestId') }}</th>
                  <th class="px-5 py-3 text-right font-medium">{{ t('researchGroup.usage.cost') }}</th>
                  <th class="px-5 py-3 font-medium">{{ t('researchGroup.usage.time') }}</th>
                </tr>
              </thead>
              <tbody class="divide-y divide-gray-100 dark:divide-dark-700">
                <tr v-for="item in usage.items" :key="item.id">
                  <td class="px-5 py-4"><p class="font-medium text-gray-900 dark:text-white">{{ item.username || '-' }}</p><p class="text-xs text-gray-500">{{ item.email }}</p></td>
                  <td class="px-5 py-4 text-gray-700 dark:text-gray-300">{{ item.model || '-' }}</td>
                  <td class="max-w-[240px] truncate px-5 py-4 font-mono text-xs text-gray-500" :title="item.request_id">{{ item.request_id || '-' }}</td>
                  <td class="px-5 py-4 text-right font-medium text-indigo-600 dark:text-indigo-400">{{ formatUsd(item.total_cost) }}</td>
                  <td class="px-5 py-4 text-xs text-gray-500">{{ formatDate(item.created_at) }}</td>
                </tr>
              </tbody>
            </table>
          </div>
          <Pagination v-if="usage.total > usage.page_size" :total="usage.total" :page="usage.page" :page-size="usage.page_size" :show-page-size-selector="false" @update:page="changeUsagePage" />
        </section>
      </template>

      <template v-else-if="activeMemberContext && context?.member">
        <section class="card p-6">
          <div class="flex flex-col gap-4 sm:flex-row sm:items-start sm:justify-between">
            <div>
              <div class="flex flex-wrap items-center gap-2"><h1 class="text-2xl font-semibold text-gray-900 dark:text-white">{{ context.group.name }}</h1><StatusPill :status="context.member.status" /></div>
              <p class="mt-2 text-sm text-gray-500 dark:text-gray-400">{{ t('researchGroup.member.owner', { owner: context.group.owner_username || context.group.owner_email }) }}</p>
            </div>
            <button class="btn btn-danger" :disabled="working" @click="requestLeave">{{ t('researchGroup.actions.leave') }}</button>
          </div>
        </section>
        <section class="grid gap-4 sm:grid-cols-2 xl:grid-cols-4">
          <StatCard :label="t('researchGroup.member.limit')" :value="formatUsd(context.member.monthly_limit_usd)" tone="indigo" />
          <StatCard :label="t('researchGroup.member.used')" :value="formatUsd(context.member.monthly_usage_usd)" tone="purple" />
          <StatCard :label="t('researchGroup.member.reserved')" :value="formatUsd(context.member.monthly_reserved_usd)" tone="amber" />
          <StatCard :label="t('researchGroup.member.remaining')" :value="formatUsd(context.member.monthly_remaining_usd)" tone="emerald" />
        </section>
        <section class="card p-6">
          <h2 class="text-lg font-semibold text-gray-900 dark:text-white">{{ t('researchGroup.member.billingTitle') }}</h2>
          <p class="mt-2 text-sm text-gray-600 dark:text-gray-300">{{ memberBillingDescription }}</p>
          <p class="mt-3 text-sm text-gray-500 dark:text-gray-400">{{ t('researchGroup.member.resetsAt', { date: formatDate(context.member.resets_at) }) }}</p>
        </section>
      </template>

      <template v-else-if="invitations.length > 0">
        <section class="card p-6">
          <h1 class="text-2xl font-semibold text-gray-900 dark:text-white">{{ t('researchGroup.invitations.title') }}</h1>
          <p class="mt-1 text-sm text-gray-500 dark:text-gray-400">{{ t('researchGroup.invitations.description') }}</p>
          <div class="mt-5 space-y-3">
            <div v-for="invitation in invitations" :key="invitation.id" class="flex flex-col gap-4 rounded-xl border border-gray-200 p-4 dark:border-dark-700 sm:flex-row sm:items-center sm:justify-between">
              <div>
                <p class="font-medium text-gray-900 dark:text-white">{{ invitation.research_group_name || '-' }}</p>
                <p class="mt-1 text-sm text-gray-600 dark:text-gray-300">
                  {{ t('researchGroup.dashboard.invitationDescription', {
                    group: invitation.research_group_name || '-',
                    owner: invitation.owner_username || invitation.owner_email || '-',
                  }) }}
                </p>
                <p class="mt-1 text-sm text-gray-500">{{ t('researchGroup.invitations.limit', { amount: formatUsd(invitation.monthly_limit_usd) }) }}</p>
              </div>
              <div class="flex gap-2"><button class="btn btn-secondary" :disabled="working" @click="rejectInvitation(invitation)">{{ t('researchGroup.invitations.reject') }}</button><button class="btn btn-primary" :disabled="working" @click="acceptInvitation(invitation)">{{ t('researchGroup.invitations.accept') }}</button></div>
            </div>
          </div>
        </section>
      </template>

      <section v-else class="mx-auto max-w-2xl card p-6 sm:p-8">
        <h1 class="text-2xl font-semibold text-gray-900 dark:text-white">{{ t('researchGroup.create.title') }}</h1>
        <p class="mt-2 text-sm text-gray-500 dark:text-gray-400">{{ t('researchGroup.create.description') }}</p>
        <form class="mt-6 space-y-4" @submit.prevent="createGroup">
          <div><label class="mb-1.5 block text-sm font-medium text-gray-700 dark:text-gray-300">{{ t('researchGroup.fields.groupName') }}</label><input v-model.trim="groupName" class="input w-full" maxlength="100" required :placeholder="t('researchGroup.create.namePlaceholder')" /></div>
          <button class="btn btn-primary w-full sm:w-auto" :disabled="working || !groupName">{{ t('researchGroup.create.action') }}</button>
        </form>
      </section>
    </div>

    <ConfirmDialog v-if="confirmation" :show="true" :title="confirmation.title" :message="confirmation.message" :danger="confirmation.danger" @confirm="runConfirmation" @cancel="confirmation = null" />
  </AppLayout>
</template>

<script setup lang="ts">
import { computed, defineComponent, h, onMounted, reactive, ref } from 'vue'
import { useI18n } from 'vue-i18n'
import AppLayout from '@/components/layout/AppLayout.vue'
import ConfirmDialog from '@/components/common/ConfirmDialog.vue'
import LoadingSpinner from '@/components/common/LoadingSpinner.vue'
import Pagination from '@/components/common/Pagination.vue'
import researchGroupAPI, { type ResearchGroupUsageItem } from '@/api/researchGroup'
import { useAppStore } from '@/stores/app'
import { useAuthStore } from '@/stores/auth'
import { extractApiErrorMessage } from '@/utils/apiError'
import type { PaginatedResponse, ResearchGroupContext, ResearchGroupMember, ResearchGroupMemberStatus } from '@/types'

const { t } = useI18n()
const appStore = useAppStore()
const authStore = useAuthStore()
const loading = ref(true)
const working = ref(false)
const usageLoading = ref(false)
const context = ref<ResearchGroupContext | null>(null)
const invitations = ref<ResearchGroupMember[]>([])
const groupName = ref('')
const inviteEmail = ref('')
const inviteLimit = ref(10)
const limitDrafts = reactive<Record<number, number>>({})
const usageMemberId = ref('')
const usageStart = ref('')
const usageEnd = ref('')
const usage = ref<PaginatedResponse<ResearchGroupUsageItem>>({ items: [], total: 0, page: 1, page_size: 20, pages: 0 })
const confirmation = ref<{ title: string; message: string; danger: boolean; action: () => Promise<void> } | null>(null)

const isOwner = computed(() => context.value?.role === 'owner')
const activeMemberContext = computed(() => context.value?.role === 'member' && context.value.member?.status !== 'pending')
const members = computed(() => context.value?.members || [])
const memberBillingDescription = computed(() => {
  if (!context.value?.member) return ''
  if (context.value.group.status === 'paused' || context.value.member.status === 'paused') return t('researchGroup.member.pausedBilling')
  if (context.value.member.monthly_limit_usd <= 0) return t('researchGroup.member.zeroLimitBilling')
  return t('researchGroup.member.activeBilling')
})

const StatusPill = defineComponent({
  props: { status: { type: String, required: true } },
  setup(props) {
    return () => h('span', { class: ['inline-flex rounded-full px-2.5 py-1 text-xs font-medium', props.status === 'active' ? 'bg-emerald-100 text-emerald-700 dark:bg-emerald-900/30 dark:text-emerald-300' : props.status === 'pending' ? 'bg-blue-100 text-blue-700 dark:bg-blue-900/30 dark:text-blue-300' : 'bg-amber-100 text-amber-700 dark:bg-amber-900/30 dark:text-amber-300'] }, t(`researchGroup.status.${props.status}`))
  }
})

const StatCard = defineComponent({
  props: { label: { type: String, required: true }, value: { type: String, required: true }, tone: { type: String, default: 'indigo' } },
  setup(props) {
    const tones: Record<string, string> = { emerald: 'text-emerald-600 dark:text-emerald-400', indigo: 'text-indigo-600 dark:text-indigo-400', blue: 'text-blue-600 dark:text-blue-400', purple: 'text-purple-600 dark:text-purple-400', amber: 'text-amber-600 dark:text-amber-400' }
    return () => h('div', { class: 'card p-5' }, [h('p', { class: 'text-sm text-gray-500 dark:text-gray-400' }, props.label), h('p', { class: ['mt-2 text-2xl font-semibold', tones[props.tone] || tones.indigo] }, props.value)])
  }
})

function formatUsd(value: number): string { return new Intl.NumberFormat(undefined, { style: 'currency', currency: 'USD', minimumFractionDigits: 2, maximumFractionDigits: 4 }).format(value || 0) }
function formatInteger(value: number): string { return new Intl.NumberFormat().format(value || 0) }
function formatDate(value?: string | null): string { if (!value) return '-'; const date = new Date(value); return Number.isNaN(date.getTime()) ? value : date.toLocaleString() }
function limitChanged(member: ResearchGroupMember): boolean { return Number(limitDrafts[member.id]) !== member.monthly_limit_usd && Number(limitDrafts[member.id]) >= 0 }

function operationError(error: unknown): string {
  const code = String((error as { code?: unknown })?.code || '')
  const keys: Record<string, string> = {
    RESEARCH_GROUP_MEMBER_INELIGIBLE: 'researchGroup.errors.memberIneligible',
    RESEARCH_GROUP_MEMBER_NOT_ELIGIBLE: 'researchGroup.errors.memberIneligible',
    RESEARCH_GROUP_MEMBER_ALREADY_ASSIGNED: 'researchGroup.errors.alreadyAssigned',
    RESEARCH_GROUP_MEMBER_ALREADY_HAS_GROUP: 'researchGroup.errors.alreadyAssigned',
    RESEARCH_GROUP_FORBIDDEN: 'researchGroup.errors.forbidden',
    RESEARCH_GROUP_BOTH_BALANCES_INSUFFICIENT: 'researchGroup.errors.bothBalancesInsufficient',
    RESEARCH_GROUP_AND_PERSONAL_BALANCE_INSUFFICIENT: 'researchGroup.errors.bothBalancesInsufficient',
  }
  return keys[code] ? t(keys[code]) : extractApiErrorMessage(error, t('researchGroup.errors.operationFailed'))
}

async function loadPage(): Promise<void> {
  loading.value = true
  try {
    context.value = await researchGroupAPI.getContext()
    groupName.value = context.value?.group.name || ''
    syncLimitDrafts()
    if (context.value?.role === 'owner') {
      invitations.value = []
      await loadUsage()
    } else {
      invitations.value = await researchGroupAPI.listInvitations()
    }
  } catch (error) { appStore.showError(operationError(error)) } finally { loading.value = false }
}

function syncLimitDrafts(): void { for (const member of members.value) limitDrafts[member.id] = member.monthly_limit_usd }
async function refreshAfterMutation(): Promise<void> { await Promise.all([loadPage(), authStore.refreshUser().catch(() => undefined)]) }
async function mutate(action: () => Promise<unknown>, successKey: string): Promise<void> { if (working.value) return; working.value = true; try { await action(); appStore.showSuccess(t(successKey)); await refreshAfterMutation() } catch (error) { appStore.showError(operationError(error)) } finally { working.value = false } }

async function createGroup(): Promise<void> { await mutate(() => researchGroupAPI.create({ name: groupName.value }), 'researchGroup.messages.created') }
async function saveGroupName(): Promise<void> { await mutate(() => researchGroupAPI.update({ name: groupName.value }), 'researchGroup.messages.updated') }
async function toggleGroupStatus(): Promise<void> { if (!context.value) return; const status = context.value.group.status === 'active' ? 'paused' : 'active'; await mutate(() => researchGroupAPI.update({ status }), status === 'active' ? 'researchGroup.messages.resumed' : 'researchGroup.messages.paused') }
async function inviteMember(): Promise<void> { await mutate(() => researchGroupAPI.inviteMember({ email: inviteEmail.value, monthly_limit_usd: Number(inviteLimit.value) }), 'researchGroup.messages.invited'); inviteEmail.value = ''; inviteLimit.value = 10 }
async function saveMemberLimit(member: ResearchGroupMember): Promise<void> { await mutate(() => researchGroupAPI.updateMember(member.id, { monthly_limit_usd: Number(limitDrafts[member.id]) }), 'researchGroup.messages.limitUpdated') }
async function setMemberStatus(member: ResearchGroupMember, status: Exclude<ResearchGroupMemberStatus, 'pending'>): Promise<void> { await mutate(() => researchGroupAPI.updateMember(member.id, { status }), status === 'active' ? 'researchGroup.messages.memberResumed' : 'researchGroup.messages.memberPaused') }
async function acceptInvitation(invitation: ResearchGroupMember): Promise<void> { await mutate(() => researchGroupAPI.acceptInvitation(invitation.id), 'researchGroup.invitations.accepted') }
async function rejectInvitation(invitation: ResearchGroupMember): Promise<void> { await mutate(() => researchGroupAPI.rejectInvitation(invitation.id), 'researchGroup.invitations.rejected') }

function requestDissolve(): void { confirmation.value = { title: t('researchGroup.confirm.dissolveTitle'), message: t('researchGroup.confirm.dissolveMessage'), danger: true, action: async () => { await mutate(() => researchGroupAPI.dissolve(), 'researchGroup.messages.dissolved') } } }
function requestLeave(): void { confirmation.value = { title: t('researchGroup.confirm.leaveTitle'), message: t('researchGroup.confirm.leaveMessage'), danger: true, action: async () => { await mutate(() => researchGroupAPI.leave(), 'researchGroup.messages.left') } } }
function requestMemberReset(member: ResearchGroupMember): void { confirmation.value = { title: t('researchGroup.confirm.resetTitle'), message: t('researchGroup.confirm.resetMessage', { member: member.username || member.email }), danger: false, action: async () => { await mutate(() => researchGroupAPI.resetMemberUsage(member.id), 'researchGroup.messages.usageReset') } } }
function requestMemberRemoval(member: ResearchGroupMember): void { confirmation.value = { title: t('researchGroup.confirm.removeTitle'), message: t('researchGroup.confirm.removeMessage', { member: member.username || member.email }), danger: true, action: async () => { await mutate(() => researchGroupAPI.removeMember(member.id), 'researchGroup.messages.removed') } } }
async function runConfirmation(): Promise<void> { const action = confirmation.value?.action; confirmation.value = null; if (action) await action() }

function dateBoundary(value: string, endOfDay: boolean): string | undefined {
  if (!value) return undefined
  const localDate = new Date(`${value}T${endOfDay ? '23:59:59.999' : '00:00:00.000'}`)
  return Number.isNaN(localDate.getTime()) ? undefined : localDate.toISOString()
}
function usageParams(page = usage.value.page) { return { page, page_size: usage.value.page_size, member_id: usageMemberId.value ? Number(usageMemberId.value) : undefined, start_time: dateBoundary(usageStart.value, false), end_time: dateBoundary(usageEnd.value, true) } }
async function loadUsage(page = 1): Promise<void> { if (context.value?.role !== 'owner') return; usageLoading.value = true; try { usage.value = await researchGroupAPI.getUsage(usageParams(page)) } catch (error) { appStore.showError(operationError(error)) } finally { usageLoading.value = false } }
function reloadUsage(): void { void loadUsage(1) }
function changeUsagePage(page: number): void { void loadUsage(page) }

onMounted(() => { void loadPage() })
</script>
