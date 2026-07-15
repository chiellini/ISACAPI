<template>
  <AppLayout>
    <TablePageLayout>
      <template #filters>
        <div class="flex flex-wrap items-center gap-3">
          <div class="relative w-full md:w-80">
            <Icon name="search" size="md" class="absolute left-3 top-1/2 -translate-y-1/2 text-gray-400" />
            <input v-model="filters.search" class="input pl-10" :placeholder="t('admin.affiliates.withdrawals.searchPlaceholder')" @input="debounceLoad" />
          </div>
          <select v-model="filters.status" class="input w-full sm:w-44" @change="reloadFromFirstPage">
            <option value="">{{ t('admin.affiliates.withdrawals.allStatuses') }}</option>
            <option v-for="status in withdrawalStatuses" :key="status" :value="status">{{ statusLabel(status) }}</option>
          </select>
          <button class="btn btn-secondary px-3" :disabled="loading" @click="loadWithdrawals"><Icon name="refresh" :class="loading ? 'animate-spin' : ''" /></button>
          <p v-if="!authStore.isSuperAdmin" class="text-xs text-gray-500 dark:text-dark-400">{{ t('admin.affiliates.readOnlyHint') }}</p>
        </div>
      </template>

      <template #table>
        <DataTable :columns="columns" :data="withdrawals" :loading="loading" row-key="id">
          <template #cell-id="{ row }"><span class="font-mono text-sm">#{{ row.id }}</span></template>
          <template #cell-user="{ row }">
            <div><p class="font-medium text-gray-900 dark:text-white">{{ row.user_email || '-' }}</p><p class="text-xs text-gray-500">#{{ row.user_id }} · {{ row.username || '-' }}</p></div>
          </template>
          <template #cell-amount="{ row }"><span class="font-semibold text-gray-900 dark:text-white">{{ formatCurrency(row.amount) }}</span></template>
          <template #cell-payment_account="{ row }"><div><p>{{ payoutTypeLabel(row.payment_account_type) }}</p><p class="text-xs text-gray-500">{{ row.payment_account_summary }}</p></div></template>
          <template #cell-status="{ row }"><span :class="statusClass(row.status)">{{ statusLabel(row.status) }}</span></template>
          <template #cell-submitted_at="{ row }">{{ formatDateTime(row.submitted_at) }}</template>
          <template #cell-actions="{ row }"><button class="btn btn-secondary btn-sm" @click="openDetail(row)">{{ t('common.view') }}</button></template>
        </DataTable>
      </template>
      <template #pagination>
        <Pagination v-if="pagination.total > 0" :page="pagination.page" :page-size="pagination.page_size" :total="pagination.total" @update:page="changePage" @update:pageSize="changePageSize" />
      </template>
    </TablePageLayout>

    <BaseDialog :show="detailOpen" :title="t('admin.affiliates.withdrawals.detailTitle')" width="wide" @close="closeDetail">
      <div v-if="detailLoading" class="flex justify-center py-10"><Icon name="refresh" class="animate-spin text-primary-500" /></div>
      <div v-else-if="selected" class="space-y-5">
        <div class="grid gap-3 sm:grid-cols-2 lg:grid-cols-4">
          <DetailItem :label="t('admin.affiliates.withdrawals.id')" :value="`#${selected.id}`" />
          <DetailItem :label="t('admin.affiliates.withdrawals.user')" :value="selected.user_email || `#${selected.user_id}`" />
          <DetailItem :label="t('admin.affiliates.withdrawals.amount')" :value="formatCurrency(selected.amount)" />
          <div class="rounded-lg bg-gray-50 p-3 dark:bg-dark-800"><p class="text-xs text-gray-500">{{ t('admin.affiliates.withdrawals.status') }}</p><p class="mt-1"><span :class="statusClass(selected.status)">{{ statusLabel(selected.status) }}</span></p></div>
        </div>

        <div class="rounded-xl border border-gray-200 p-4 dark:border-dark-700">
          <h4 class="text-sm font-semibold text-gray-900 dark:text-white">{{ t('admin.affiliates.withdrawals.paymentAccount') }}</h4>
          <p class="mt-2 text-sm text-gray-700 dark:text-gray-300">{{ payoutTypeLabel(selected.payment_account_type) }} · {{ selected.payment_account_summary }}</p>
        </div>
        <div v-if="authStore.isSuperAdmin && visiblePaymentDetails.length" class="rounded-xl border border-amber-200 bg-amber-50/60 p-4 dark:border-amber-900/40 dark:bg-amber-900/10">
          <h4 class="text-sm font-semibold text-gray-900 dark:text-white">{{ t('admin.affiliates.withdrawals.fullPaymentDetails') }}</h4>
          <p class="mt-1 text-xs text-amber-700 dark:text-amber-300">{{ t('admin.affiliates.withdrawals.sensitiveHint') }}</p>
          <dl class="mt-4 grid gap-3 text-sm sm:grid-cols-2">
            <div v-for="item in visiblePaymentDetails" :key="item.key" :class="item.key === 'wallet_address' ? 'sm:col-span-2' : ''">
              <dt class="text-gray-500">{{ paymentDetailLabel(item.key) }}</dt>
              <dd class="mt-0.5 break-all font-mono font-medium text-gray-900 dark:text-white">{{ item.value }}</dd>
            </div>
          </dl>
        </div>

        <div v-if="selected.status === 'submitted' && authStore.isSuperAdmin" class="rounded-xl border border-gray-200 p-4 dark:border-dark-700">
          <label class="mb-1 block text-sm font-medium text-gray-700 dark:text-gray-300">{{ t('admin.affiliates.withdrawals.rejectReason') }}</label>
          <textarea v-model.trim="rejectReason" class="input min-h-20" :placeholder="t('admin.affiliates.withdrawals.rejectReasonPlaceholder')"></textarea>
          <p class="mt-2 text-xs text-gray-500 dark:text-dark-400">{{ t('admin.affiliates.withdrawals.reviewHint') }}</p>
        </div>

        <form v-if="selected.status === 'approved' && authStore.isSuperAdmin" class="grid gap-4 rounded-xl border border-primary-200 bg-primary-50/50 p-4 dark:border-primary-900/40 dark:bg-primary-900/10 sm:grid-cols-2" @submit.prevent="markPaid">
          <div><label class="mb-1 block text-sm font-medium">{{ t('admin.affiliates.withdrawals.actualCurrency') }}</label><input v-model.trim="paidForm.actual_currency" class="input uppercase" required placeholder="CNY / USD / USDT" /></div>
          <div><label class="mb-1 block text-sm font-medium">{{ t('admin.affiliates.withdrawals.actualAmount') }}</label><input v-model.number="paidForm.actual_amount" class="input" type="number" min="0.00000001" step="0.00000001" required /></div>
          <div><label class="mb-1 block text-sm font-medium">{{ t('admin.affiliates.withdrawals.exchangeRate') }}</label><input v-model.number="paidForm.exchange_rate" class="input" type="number" min="0.00000001" step="0.00000001" required /></div>
          <div><label class="mb-1 block text-sm font-medium">{{ t('admin.affiliates.withdrawals.externalReference') }}</label><input v-model.trim="paidForm.external_reference" class="input" required /></div>
        </form>

        <div v-if="selected.status === 'paid'" class="rounded-xl border border-emerald-200 bg-emerald-50 p-4 dark:border-emerald-900/40 dark:bg-emerald-900/20">
          <p class="font-semibold text-emerald-700 dark:text-emerald-300">{{ t('admin.affiliates.withdrawals.statuses.paid') }}</p>
          <dl class="mt-3 grid gap-3 text-sm sm:grid-cols-2">
            <div><dt class="text-gray-500">{{ t('admin.affiliates.withdrawals.actualPayment') }}</dt><dd class="font-medium text-gray-900 dark:text-white">{{ selected.actual_currency }} {{ selected.actual_amount }}</dd></div>
            <div><dt class="text-gray-500">{{ t('admin.affiliates.withdrawals.exchangeRate') }}</dt><dd class="font-medium text-gray-900 dark:text-white">{{ selected.exchange_rate }}</dd></div>
            <div class="sm:col-span-2"><dt class="text-gray-500">{{ t('admin.affiliates.withdrawals.externalReference') }}</dt><dd class="break-all font-mono text-gray-900 dark:text-white">{{ selected.external_reference || '-' }}</dd></div>
          </dl>
        </div>

        <div v-if="selected.reject_reason" class="rounded-xl border border-red-200 bg-red-50 p-4 text-sm text-red-700 dark:border-red-900/40 dark:bg-red-900/20 dark:text-red-300">
          {{ t('admin.affiliates.withdrawals.rejectedBecause', { reason: selected.reject_reason }) }}
        </div>
      </div>
      <template #footer>
        <button class="btn btn-secondary" :disabled="actionLoading" @click="closeDetail">{{ t('common.close') }}</button>
        <template v-if="selected?.status === 'submitted' && authStore.isSuperAdmin">
          <button class="btn btn-danger" :disabled="actionLoading || !rejectReason" @click="reject">{{ t('admin.affiliates.withdrawals.reject') }}</button>
          <button class="btn btn-primary" :disabled="actionLoading" @click="approve">{{ t('admin.affiliates.withdrawals.approve') }}</button>
        </template>
        <button v-if="selected?.status === 'approved' && authStore.isSuperAdmin" data-test="mark-paid" class="btn btn-primary" :disabled="actionLoading || !paidFormValid" @click="markPaid">
          {{ t('admin.affiliates.withdrawals.markPaid') }}
        </button>
      </template>
    </BaseDialog>
  </AppLayout>
</template>

<script setup lang="ts">
import { computed, defineComponent, h, onMounted, reactive, ref } from 'vue'
import { useI18n } from 'vue-i18n'
import AppLayout from '@/components/layout/AppLayout.vue'
import TablePageLayout from '@/components/layout/TablePageLayout.vue'
import DataTable from '@/components/common/DataTable.vue'
import Pagination from '@/components/common/Pagination.vue'
import BaseDialog from '@/components/common/BaseDialog.vue'
import Icon from '@/components/icons/Icon.vue'
import type { Column } from '@/components/common/types'
import { affiliatesAPI, type AdminAffiliateWithdrawal } from '@/api/admin/affiliates'
import type { AffiliatePaymentAccountType, AffiliateWithdrawalStatus } from '@/types'
import { useAppStore } from '@/stores/app'
import { useAuthStore } from '@/stores/auth'
import { extractApiErrorMessage } from '@/utils/apiError'
import { formatCurrency, formatDateTime } from '@/utils/format'

const { t } = useI18n()
const appStore = useAppStore()
const authStore = useAuthStore()
const withdrawalStatuses: AffiliateWithdrawalStatus[] = ['submitted', 'approved', 'paid', 'rejected', 'canceled']
const loading = ref(false)
const detailLoading = ref(false)
const actionLoading = ref(false)
const detailOpen = ref(false)
const withdrawals = ref<AdminAffiliateWithdrawal[]>([])
const selected = ref<AdminAffiliateWithdrawal | null>(null)
const rejectReason = ref('')
const paidForm = reactive({ actual_currency: 'CNY', actual_amount: 0, exchange_rate: 1, external_reference: '' })
const filters = reactive<{ search: string; status: AffiliateWithdrawalStatus | '' }>({ search: '', status: '' })
const pagination = reactive({ page: 1, page_size: 20, total: 0 })
let debounceTimer: ReturnType<typeof setTimeout> | null = null

const columns = computed<Column[]>(() => [
  { key: 'id', label: t('admin.affiliates.withdrawals.id') },
  { key: 'user', label: t('admin.affiliates.withdrawals.user') },
  { key: 'amount', label: t('admin.affiliates.withdrawals.amount') },
  { key: 'payment_account', label: t('admin.affiliates.withdrawals.paymentAccount') },
  { key: 'status', label: t('admin.affiliates.withdrawals.status') },
  { key: 'submitted_at', label: t('admin.affiliates.withdrawals.submittedAt') },
  { key: 'actions', label: t('common.actions') },
])
const paidFormValid = computed(() => paidForm.actual_currency.trim() !== '' && paidForm.actual_amount > 0 && paidForm.exchange_rate > 0 && paidForm.external_reference.trim() !== '')
const visiblePaymentDetails = computed(() => Object.entries(selected.value?.payment_details || {}).filter((entry): entry is [string, string] => typeof entry[1] === 'string' && entry[1].trim() !== '').map(([key, value]) => ({ key, value })))
const DetailItem = defineComponent({ props: { label: { type: String, required: true }, value: { type: String, required: true } }, setup: (props) => () => h('div', { class: 'rounded-lg bg-gray-50 p-3 dark:bg-dark-800' }, [h('p', { class: 'text-xs text-gray-500' }, props.label), h('p', { class: 'mt-1 truncate font-medium text-gray-900 dark:text-white' }, props.value)]) })

function newIdempotencyKey(action: string, id: number): string { return `affiliate-${action}-${id}-${globalThis.crypto?.randomUUID?.() || `${Date.now()}-${Math.random().toString(36).slice(2)}`}` }
async function loadWithdrawals(): Promise<void> {
  loading.value = true
  try {
    const response = await affiliatesAPI.listWithdrawals({ page: pagination.page, page_size: pagination.page_size, search: filters.search.trim(), status: filters.status })
    withdrawals.value = response.items
    pagination.total = response.total
  } catch (error) { appStore.showError(extractApiErrorMessage(error, t('admin.affiliates.withdrawals.loadFailed'))) }
  finally { loading.value = false }
}
function debounceLoad(): void { if (debounceTimer) clearTimeout(debounceTimer); debounceTimer = setTimeout(() => { pagination.page = 1; void loadWithdrawals() }, 300) }
function reloadFromFirstPage(): void { pagination.page = 1; void loadWithdrawals() }
function changePage(page: number): void { pagination.page = page; void loadWithdrawals() }
function changePageSize(pageSize: number): void { pagination.page = 1; pagination.page_size = pageSize; void loadWithdrawals() }
async function openDetail(row: AdminAffiliateWithdrawal): Promise<void> {
  selected.value = row; detailOpen.value = true; rejectReason.value = ''
  Object.assign(paidForm, { actual_currency: 'CNY', actual_amount: row.amount, exchange_rate: 1, external_reference: '' })
  if (!authStore.isSuperAdmin) {
    detailLoading.value = false
    return
  }
  detailLoading.value = true
  try { selected.value = await affiliatesAPI.getWithdrawal(row.id) }
  catch (error) { appStore.showError(extractApiErrorMessage(error, t('admin.affiliates.withdrawals.detailFailed'))) }
  finally { detailLoading.value = false }
}
function closeDetail(): void { if (!actionLoading.value) { detailOpen.value = false; selected.value = null } }
async function approve(): Promise<void> {
  if (!authStore.isSuperAdmin || selected.value?.status !== 'submitted' || actionLoading.value) return
  actionLoading.value = true
  try { selected.value = await affiliatesAPI.approveWithdrawal(selected.value.id, newIdempotencyKey('approve', selected.value.id)); syncSelected(); appStore.showSuccess(t('admin.affiliates.withdrawals.approved')) }
  catch (error) { appStore.showError(extractApiErrorMessage(error, t('admin.affiliates.withdrawals.actionFailed'))) }
  finally { actionLoading.value = false }
}
async function reject(): Promise<void> {
  if (!authStore.isSuperAdmin || selected.value?.status !== 'submitted' || !rejectReason.value || actionLoading.value) return
  actionLoading.value = true
  try { selected.value = await affiliatesAPI.rejectWithdrawal(selected.value.id, { reason: rejectReason.value }, newIdempotencyKey('reject', selected.value.id)); syncSelected(); appStore.showSuccess(t('admin.affiliates.withdrawals.rejected')) }
  catch (error) { appStore.showError(extractApiErrorMessage(error, t('admin.affiliates.withdrawals.actionFailed'))) }
  finally { actionLoading.value = false }
}
async function markPaid(): Promise<void> {
  if (!authStore.isSuperAdmin || selected.value?.status !== 'approved' || !paidFormValid.value || actionLoading.value) return
  actionLoading.value = true
  try {
    selected.value = await affiliatesAPI.markWithdrawalPaid(selected.value.id, { actual_currency: paidForm.actual_currency.toUpperCase(), actual_amount: Number(paidForm.actual_amount), exchange_rate: Number(paidForm.exchange_rate), external_reference: paidForm.external_reference.trim() }, newIdempotencyKey('mark-paid', selected.value.id))
    syncSelected(); appStore.showSuccess(t('admin.affiliates.withdrawals.markedPaid'))
  } catch (error) { appStore.showError(extractApiErrorMessage(error, t('admin.affiliates.withdrawals.actionFailed'))) }
  finally { actionLoading.value = false }
}
function syncSelected(): void { if (!selected.value) return; const index = withdrawals.value.findIndex((item) => item.id === selected.value!.id); if (index >= 0) withdrawals.value[index] = selected.value }
function statusLabel(status: AffiliateWithdrawalStatus): string { return t(`admin.affiliates.withdrawals.statuses.${status}`) }
function statusClass(status: AffiliateWithdrawalStatus): string { return { submitted: 'badge badge-yellow', approved: 'badge badge-blue', paid: 'badge badge-green', rejected: 'badge badge-red', canceled: 'badge badge-gray' }[status] }
function payoutTypeLabel(type: AffiliatePaymentAccountType): string { return t(`affiliate.paymentAccounts.types.${type}`) }
function paymentDetailLabel(key: string): string {
  const keys: Record<string, string> = { account_name: 'accountName', account_number: selected.value?.payment_account_type === 'alipay' ? 'alipayAccount' : 'cardNumber', bank_name: 'bankName', usdt_network: 'network', wallet_address: 'walletAddress' }
  return t(`affiliate.paymentAccounts.${keys[key] || key}`)
}
onMounted(() => { void loadWithdrawals() })
</script>
