<template>
  <AppLayout>
    <TablePageLayout>
      <template #filters>
        <div class="flex flex-wrap items-center gap-3">
          <div class="relative w-full md:w-80">
            <Icon name="search" size="md" class="absolute left-3 top-1/2 -translate-y-1/2 text-gray-400" />
            <input v-model="filters.search" class="input pl-10" :placeholder="t('admin.affiliates.agents.searchPlaceholder')" @input="debounceLoad" />
          </div>
          <select v-model="filters.status" class="input w-full sm:w-44" @change="reloadFromFirstPage">
            <option value="">{{ t('admin.affiliates.agents.allStatuses') }}</option>
            <option value="inactive">{{ statusLabel('inactive') }}</option>
            <option value="active">{{ statusLabel('active') }}</option>
            <option value="suspended">{{ statusLabel('suspended') }}</option>
          </select>
          <button class="btn btn-secondary px-3" :disabled="loading" @click="loadAgents">
            <Icon name="refresh" size="md" :class="loading ? 'animate-spin' : ''" />
          </button>
          <p v-if="!authStore.isSuperAdmin" class="text-xs text-gray-500 dark:text-dark-400">{{ t('admin.affiliates.readOnlyHint') }}</p>
        </div>
      </template>

      <template #table>
        <DataTable :columns="columns" :data="agents" :loading="loading" row-key="user_id">
          <template #cell-user="{ row }">
            <div>
              <p class="font-medium text-gray-900 dark:text-white">{{ row.email || '-' }}</p>
              <p class="text-xs text-gray-500 dark:text-dark-400">#{{ row.user_id }} · {{ row.username || '-' }}</p>
            </div>
          </template>
          <template #cell-status="{ row }"><span :class="statusClass(row.status)">{{ statusLabel(row.status) }}</span></template>
          <template #cell-aff_code="{ row }"><code class="text-xs text-gray-700 dark:text-gray-300">{{ row.aff_code || '-' }}</code></template>
          <template #cell-rebate_rate_percent="{ row }">{{ formatPercent(row.rebate_rate_percent) }}</template>
          <template #cell-invited_count="{ row }">{{ row.invited_count.toLocaleString() }}</template>
          <template #cell-available_commission="{ row }"><span class="font-medium text-emerald-600 dark:text-emerald-400">{{ formatCurrency(row.available_commission) }}</span></template>
          <template #cell-frozen_commission="{ row }">{{ formatCurrency(row.frozen_commission) }}</template>
          <template #cell-withdrawal_reserved="{ row }">{{ formatCurrency(row.withdrawal_reserved) }}</template>
          <template #cell-debt="{ row }"><span :class="row.debt > 0 ? 'font-medium text-red-600 dark:text-red-400' : ''">{{ formatCurrency(row.debt) }}</span></template>
          <template #cell-actions="{ row }">
            <div v-if="authStore.isSuperAdmin" class="flex justify-end gap-2">
              <button v-if="row.status !== 'active'" class="btn btn-primary btn-sm" :disabled="updatingUserId === row.user_id" @click="setAgentStatus(row, 'active')">
                {{ t('admin.affiliates.agents.activate') }}
              </button>
              <button v-else class="btn btn-secondary btn-sm" :disabled="updatingUserId === row.user_id" @click="setAgentStatus(row, 'suspended')">
                {{ t('admin.affiliates.agents.suspend') }}
              </button>
            </div>
            <span v-else class="text-xs text-gray-400">-</span>
          </template>
        </DataTable>
      </template>

      <template #pagination>
        <Pagination v-if="pagination.total > 0" :page="pagination.page" :page-size="pagination.page_size" :total="pagination.total" @update:page="changePage" @update:pageSize="changePageSize" />
      </template>
    </TablePageLayout>
  </AppLayout>
</template>

<script setup lang="ts">
import { computed, onMounted, reactive, ref } from 'vue'
import { useI18n } from 'vue-i18n'
import AppLayout from '@/components/layout/AppLayout.vue'
import TablePageLayout from '@/components/layout/TablePageLayout.vue'
import DataTable from '@/components/common/DataTable.vue'
import Pagination from '@/components/common/Pagination.vue'
import Icon from '@/components/icons/Icon.vue'
import type { Column } from '@/components/common/types'
import { affiliatesAPI, type AffiliateAgentEntry } from '@/api/admin/affiliates'
import type { AffiliateStatus } from '@/types'
import { useAppStore } from '@/stores/app'
import { useAuthStore } from '@/stores/auth'
import { extractApiErrorMessage } from '@/utils/apiError'
import { formatCurrency } from '@/utils/format'

const { t } = useI18n()
const appStore = useAppStore()
const authStore = useAuthStore()
const loading = ref(false)
const updatingUserId = ref<number | null>(null)
const agents = ref<AffiliateAgentEntry[]>([])
const filters = reactive<{ search: string; status: AffiliateStatus | '' }>({ search: '', status: '' })
const pagination = reactive({ page: 1, page_size: 20, total: 0 })
let debounceTimer: ReturnType<typeof setTimeout> | null = null

const columns = computed<Column[]>(() => [
  { key: 'user', label: t('admin.affiliates.agents.user') },
  { key: 'status', label: t('admin.affiliates.agents.status') },
  { key: 'aff_code', label: t('admin.affiliates.agents.affCode') },
  { key: 'rebate_rate_percent', label: t('admin.affiliates.agents.rebateRate') },
  { key: 'invited_count', label: t('admin.affiliates.agents.invitedCount') },
  { key: 'available_commission', label: t('admin.affiliates.agents.available') },
  { key: 'frozen_commission', label: t('admin.affiliates.agents.frozen') },
  { key: 'withdrawal_reserved', label: t('admin.affiliates.agents.reserved') },
  { key: 'debt', label: t('admin.affiliates.agents.debt') },
  { key: 'actions', label: t('common.actions'), class: 'text-right' },
])

async function loadAgents(): Promise<void> {
  loading.value = true
  try {
    const response = await affiliatesAPI.listAgents({ page: pagination.page, page_size: pagination.page_size, search: filters.search.trim(), status: filters.status })
    agents.value = response.items
    pagination.total = response.total
  } catch (error) {
    appStore.showError(extractApiErrorMessage(error, t('admin.affiliates.agents.loadFailed')))
  } finally {
    loading.value = false
  }
}

function debounceLoad(): void {
  if (debounceTimer) clearTimeout(debounceTimer)
  debounceTimer = setTimeout(() => { pagination.page = 1; void loadAgents() }, 300)
}
function reloadFromFirstPage(): void { pagination.page = 1; void loadAgents() }
function changePage(page: number): void { pagination.page = page; void loadAgents() }
function changePageSize(pageSize: number): void { pagination.page = 1; pagination.page_size = pageSize; void loadAgents() }

async function setAgentStatus(agent: AffiliateAgentEntry, status: AffiliateStatus): Promise<void> {
  if (!authStore.isSuperAdmin || updatingUserId.value !== null || agent.status === status) return
  updatingUserId.value = agent.user_id
  try {
    const idempotencyKey = `affiliate-agent-status-${agent.user_id}-${globalThis.crypto?.randomUUID?.() || `${Date.now()}-${Math.random().toString(36).slice(2)}`}`
    const updated = await affiliatesAPI.updateAgentStatus(agent.user_id, status, idempotencyKey)
    const index = agents.value.findIndex((item) => item.user_id === agent.user_id)
    if (index >= 0) agents.value[index] = updated
    appStore.showSuccess(t(status === 'active' ? 'admin.affiliates.agents.activated' : 'admin.affiliates.agents.suspended'))
  } catch (error) {
    appStore.showError(extractApiErrorMessage(error, t('admin.affiliates.agents.updateFailed')))
  } finally {
    updatingUserId.value = null
  }
}

function statusLabel(status: AffiliateStatus): string { return t(`admin.affiliates.statuses.${status}`) }
function statusClass(status: AffiliateStatus): string {
  return { inactive: 'badge badge-gray', active: 'badge badge-green', suspended: 'badge badge-yellow' }[status]
}
function formatPercent(value: number): string { return `${Number(value || 0).toFixed(2).replace(/\.00$/, '')}%` }

onMounted(() => { void loadAgents() })
</script>
