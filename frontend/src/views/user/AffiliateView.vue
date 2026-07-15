<template>
  <AppLayout>
    <div class="space-y-6">
      <div v-if="loading" class="flex justify-center py-12">
        <div class="h-8 w-8 animate-spin rounded-full border-2 border-primary-500 border-t-transparent"></div>
      </div>

      <div v-else-if="detail && detail.status !== 'active'" class="card mx-auto max-w-2xl p-8 text-center">
        <div class="mx-auto flex h-14 w-14 items-center justify-center rounded-full bg-gray-100 text-gray-500 dark:bg-dark-800 dark:text-dark-300">
          <Icon name="users" size="lg" />
        </div>
        <h2 class="mt-4 text-xl font-semibold text-gray-900 dark:text-white">
          {{ t(`affiliate.status.${detail.status}.title`) }}
        </h2>
        <p class="mt-2 text-sm text-gray-500 dark:text-dark-400">
          {{ t(`affiliate.status.${detail.status}.description`) }}
        </p>
      </div>

      <template v-else-if="detail">
        <div class="grid gap-4 sm:grid-cols-2 xl:grid-cols-4">
          <StatCard :label="t('affiliate.stats.availableCommission')" :value="formatCurrency(availableCommission)" tone="green" />
          <StatCard :label="t('affiliate.stats.frozenCommission')" :value="formatCurrency(frozenCommission)" tone="amber" />
          <StatCard :label="t('affiliate.stats.withdrawalReserved')" :value="formatCurrency(withdrawalReserved)" />
          <StatCard :label="t('affiliate.stats.debt')" :value="formatCurrency(debt)" :tone="debt > 0 ? 'red' : 'default'" />
        </div>

        <div class="card p-6">
          <div class="flex flex-col gap-2 sm:flex-row sm:items-start sm:justify-between">
            <div>
              <h3 class="text-base font-semibold text-gray-900 dark:text-white">{{ t('affiliate.title') }}</h3>
              <p class="mt-1 text-sm text-gray-500 dark:text-dark-400">{{ t('affiliate.description') }}</p>
            </div>
            <div class="rounded-lg bg-primary-50 px-3 py-2 text-sm text-primary-700 dark:bg-primary-900/20 dark:text-primary-300">
              {{ t('affiliate.stats.rebateRate') }}：{{ formattedRebateRate }}%
              <span class="mx-2 text-primary-300 dark:text-primary-700">·</span>
              {{ t('affiliate.stats.invitedUsers') }}：{{ detail.aff_count.toLocaleString() }}
            </div>
          </div>

          <div class="mt-5 grid gap-4 md:grid-cols-2">
            <div class="space-y-2">
              <p class="text-sm font-medium text-gray-700 dark:text-gray-300">{{ t('affiliate.yourCode') }}</p>
              <div class="flex items-center gap-2 rounded-xl border border-gray-200 bg-gray-50 px-3 py-2 dark:border-dark-700 dark:bg-dark-900">
                <code class="flex-1 truncate text-sm font-semibold text-gray-900 dark:text-white">{{ detail.aff_code }}</code>
                <button class="btn btn-secondary btn-sm" @click="copyCode">
                  <Icon name="copy" size="sm" />
                  <span>{{ t('affiliate.copyCode') }}</span>
                </button>
              </div>
            </div>
            <div class="space-y-2">
              <p class="text-sm font-medium text-gray-700 dark:text-gray-300">{{ t('affiliate.inviteLink') }}</p>
              <div class="flex items-center gap-2 rounded-xl border border-gray-200 bg-gray-50 px-3 py-2 dark:border-dark-700 dark:bg-dark-900">
                <code class="flex-1 truncate text-sm text-gray-700 dark:text-gray-300">{{ inviteLink }}</code>
                <button class="btn btn-secondary btn-sm" @click="copyInviteLink">
                  <Icon name="copy" size="sm" />
                  <span>{{ t('affiliate.copyLink') }}</span>
                </button>
              </div>
            </div>
          </div>
        </div>

        <div class="grid gap-6 xl:grid-cols-2">
          <section class="card p-6">
            <div class="flex items-start justify-between gap-4">
              <div>
                <h3 class="text-base font-semibold text-gray-900 dark:text-white">{{ t('affiliate.paymentAccounts.title') }}</h3>
                <p class="mt-1 text-sm text-gray-500 dark:text-dark-400">{{ t('affiliate.paymentAccounts.description') }}</p>
              </div>
              <button class="btn btn-secondary btn-sm" @click="openAccountDialog()">
                <Icon name="plus" size="sm" />
                {{ t('affiliate.paymentAccounts.add') }}
              </button>
            </div>

            <div v-if="accountsLoading" class="flex justify-center py-8">
              <Icon name="refresh" class="animate-spin text-primary-500" />
            </div>
            <div v-else-if="paymentAccounts.length === 0" class="mt-4 rounded-xl border border-dashed border-gray-300 p-6 text-center text-sm text-gray-500 dark:border-dark-700 dark:text-dark-400">
              {{ t('affiliate.paymentAccounts.empty') }}
            </div>
            <div v-else class="mt-4 space-y-3">
              <div v-for="account in paymentAccounts" :key="account.id" class="flex items-center gap-3 rounded-xl border border-gray-200 p-4 dark:border-dark-700">
                <div class="flex h-10 w-10 shrink-0 items-center justify-center rounded-lg bg-gray-100 text-gray-600 dark:bg-dark-800 dark:text-dark-300">
                  <Icon name="creditCard" size="sm" />
                </div>
                <div class="min-w-0 flex-1">
                  <div class="flex items-center gap-2">
                    <span class="text-sm font-medium text-gray-900 dark:text-white">{{ paymentAccountTypeLabel(account.type) }}</span>
                    <span v-if="account.is_default" class="badge badge-primary">{{ t('affiliate.paymentAccounts.default') }}</span>
                  </div>
                  <p class="mt-0.5 truncate text-sm text-gray-500 dark:text-dark-400">{{ account.masked_summary }}</p>
                </div>
                <button class="btn btn-ghost btn-sm" :title="t('common.edit')" @click="openAccountDialog(account)">
                  <Icon name="edit" size="sm" />
                </button>
                <button class="btn btn-ghost btn-sm text-red-600" :title="t('common.delete')" @click="removePaymentAccount(account)">
                  <Icon name="trash" size="sm" />
                </button>
              </div>
            </div>
          </section>

          <section class="card p-6">
            <h3 class="text-base font-semibold text-gray-900 dark:text-white">{{ t('affiliate.withdrawal.title') }}</h3>
            <p class="mt-1 text-sm text-gray-500 dark:text-dark-400">
              {{ t('affiliate.withdrawal.description', { amount: formatCurrency(minimumWithdrawal) }) }}
            </p>

            <div v-if="debt > 0" class="mt-4 rounded-xl border border-red-200 bg-red-50 p-3 text-sm text-red-700 dark:border-red-900/40 dark:bg-red-900/20 dark:text-red-300">
              {{ t('affiliate.withdrawal.debtBlocked', { amount: formatCurrency(debt) }) }}
            </div>
            <form class="mt-4 space-y-4" @submit.prevent="submitWithdrawal">
              <div>
                <label class="mb-1 block text-sm font-medium text-gray-700 dark:text-gray-300">{{ t('affiliate.withdrawal.account') }}</label>
                <select v-model.number="withdrawalForm.payment_account_id" class="input" required>
                  <option :value="0" disabled>{{ t('affiliate.withdrawal.selectAccount') }}</option>
                  <option v-for="account in paymentAccounts" :key="account.id" :value="account.id">
                    {{ paymentAccountTypeLabel(account.type) }} · {{ account.masked_summary }}
                  </option>
                </select>
              </div>
              <div>
                <label class="mb-1 block text-sm font-medium text-gray-700 dark:text-gray-300">{{ t('affiliate.withdrawal.amount') }}</label>
                <input v-model.number="withdrawalForm.amount" class="input" type="number" step="0.01" :min="minimumWithdrawal" :max="availableCommission" required />
                <p class="mt-1 text-xs text-gray-500 dark:text-dark-400">
                  {{ t('affiliate.withdrawal.available', { amount: formatCurrency(availableCommission) }) }}
                </p>
              </div>
              <button class="btn btn-primary w-full" type="submit" :disabled="!canWithdraw || withdrawalSubmitting">
                <Icon v-if="withdrawalSubmitting" name="refresh" size="sm" class="animate-spin" />
                {{ withdrawalSubmitting ? t('affiliate.withdrawal.submitting') : t('affiliate.withdrawal.submit') }}
              </button>
            </form>
          </section>
        </div>

        <section class="card p-6">
          <div class="flex items-center justify-between gap-3">
            <h3 class="text-base font-semibold text-gray-900 dark:text-white">{{ t('affiliate.withdrawal.records') }}</h3>
            <button class="btn btn-secondary btn-sm" :disabled="withdrawalsLoading" @click="loadWithdrawals">
              <Icon name="refresh" size="sm" :class="withdrawalsLoading ? 'animate-spin' : ''" />
            </button>
          </div>
          <div v-if="withdrawals.length === 0 && !withdrawalsLoading" class="mt-4 rounded-xl border border-dashed border-gray-300 p-6 text-center text-sm text-gray-500 dark:border-dark-700 dark:text-dark-400">
            {{ t('affiliate.withdrawal.empty') }}
          </div>
          <div v-else class="mt-4 overflow-x-auto">
            <table class="w-full min-w-[760px] text-left text-sm">
              <thead>
                <tr class="border-b border-gray-200 text-gray-500 dark:border-dark-700 dark:text-dark-400">
                  <th class="px-3 py-2 font-medium">{{ t('affiliate.withdrawal.id') }}</th>
                  <th class="px-3 py-2 font-medium">{{ t('affiliate.withdrawal.account') }}</th>
                  <th class="px-3 py-2 text-right font-medium">{{ t('affiliate.withdrawal.amount') }}</th>
                  <th class="px-3 py-2 font-medium">{{ t('affiliate.withdrawal.statusLabel') }}</th>
                  <th class="px-3 py-2 font-medium">{{ t('affiliate.withdrawal.submittedAt') }}</th>
                  <th class="px-3 py-2 text-right font-medium">{{ t('common.actions') }}</th>
                </tr>
              </thead>
              <tbody>
                <tr v-for="item in withdrawals" :key="item.id" class="border-b border-gray-100 last:border-0 dark:border-dark-800">
                  <td class="px-3 py-3 font-mono text-gray-700 dark:text-gray-300">#{{ item.id }}</td>
                  <td class="px-3 py-3 text-gray-700 dark:text-gray-300">{{ paymentAccountTypeLabel(item.payment_account_type) }} · {{ item.payment_account_summary }}</td>
                  <td class="px-3 py-3 text-right font-medium text-gray-900 dark:text-white">{{ formatCurrency(item.amount) }}</td>
                  <td class="px-3 py-3"><span :class="withdrawalStatusClass(item.status)">{{ withdrawalStatusLabel(item.status) }}</span></td>
                  <td class="px-3 py-3 text-gray-500 dark:text-dark-400">{{ formatDateTime(item.submitted_at) }}</td>
                  <td class="px-3 py-3 text-right">
                    <button v-if="item.status === 'submitted'" class="btn btn-ghost btn-sm" :disabled="cancelingWithdrawalId === item.id" @click="cancelWithdrawal(item)">
                      {{ t('affiliate.withdrawal.cancel') }}
                    </button>
                    <span v-else>-</span>
                  </td>
                </tr>
              </tbody>
            </table>
          </div>
          <Pagination
            v-if="withdrawalPagination.total > withdrawalPagination.page_size"
            class="mt-4"
            :page="withdrawalPagination.page"
            :page-size="withdrawalPagination.page_size"
            :total="withdrawalPagination.total"
            @update:page="changeWithdrawalPage"
            @update:pageSize="changeWithdrawalPageSize"
          />
        </section>

        <section class="card p-6">
          <h3 class="text-base font-semibold text-gray-900 dark:text-white">{{ t('affiliate.invitees.title') }}</h3>
          <div v-if="detail.invitees.length === 0" class="mt-4 rounded-xl border border-dashed border-gray-300 p-6 text-center text-sm text-gray-500 dark:border-dark-700 dark:text-dark-400">
            {{ t('affiliate.invitees.empty') }}
          </div>
          <div v-else class="mt-4 overflow-x-auto">
            <table class="w-full min-w-[560px] text-left text-sm">
              <thead><tr class="border-b border-gray-200 text-gray-500 dark:border-dark-700 dark:text-dark-400">
                <th class="px-3 py-2 font-medium">{{ t('affiliate.invitees.columns.email') }}</th>
                <th class="px-3 py-2 font-medium">{{ t('affiliate.invitees.columns.username') }}</th>
                <th class="px-3 py-2 text-right font-medium">{{ t('affiliate.invitees.columns.rebate') }}</th>
                <th class="px-3 py-2 font-medium">{{ t('affiliate.invitees.columns.joinedAt') }}</th>
              </tr></thead>
              <tbody><tr v-for="item in detail.invitees" :key="item.user_id" class="border-b border-gray-100 last:border-0 dark:border-dark-800">
                <td class="px-3 py-3 text-gray-900 dark:text-white">{{ item.email || '-' }}</td>
                <td class="px-3 py-3 text-gray-700 dark:text-gray-300">{{ item.username || '-' }}</td>
                <td class="px-3 py-3 text-right font-medium text-emerald-600 dark:text-emerald-400">{{ formatCurrency(item.total_rebate) }}</td>
                <td class="px-3 py-3 text-gray-700 dark:text-gray-300">{{ formatDateTime(item.created_at) || '-' }}</td>
              </tr></tbody>
            </table>
          </div>
        </section>
      </template>
    </div>

    <BaseDialog :show="accountDialogOpen" :title="editingAccountId ? t('affiliate.paymentAccounts.edit') : t('affiliate.paymentAccounts.add')" @close="closeAccountDialog">
      <form class="space-y-4" @submit.prevent="savePaymentAccount">
        <div>
          <label class="mb-1 block text-sm font-medium text-gray-700 dark:text-gray-300">{{ t('affiliate.paymentAccounts.type') }}</label>
          <select v-model="accountForm.type" class="input" :disabled="editingAccountId !== null">
            <option value="alipay">{{ t('affiliate.paymentAccounts.types.alipay') }}</option>
            <option value="bank_card">{{ t('affiliate.paymentAccounts.types.bank_card') }}</option>
            <option value="usdt">{{ t('affiliate.paymentAccounts.types.usdt') }}</option>
          </select>
        </div>
        <p v-if="editingAccountId" class="rounded-lg bg-amber-50 p-3 text-sm text-amber-700 dark:bg-amber-900/20 dark:text-amber-300">{{ t('affiliate.paymentAccounts.reenterHint') }}</p>
        <div v-if="accountForm.type !== 'usdt'">
          <label class="mb-1 block text-sm font-medium text-gray-700 dark:text-gray-300">{{ t('affiliate.paymentAccounts.accountName') }}</label>
          <input v-model.trim="accountForm.account_name" class="input" required autocomplete="off" />
        </div>
        <div v-if="accountForm.type === 'bank_card'">
          <label class="mb-1 block text-sm font-medium text-gray-700 dark:text-gray-300">{{ t('affiliate.paymentAccounts.bankName') }}</label>
          <input v-model.trim="accountForm.bank_name" class="input" required autocomplete="off" />
        </div>
        <div v-if="accountForm.type !== 'usdt'">
          <label class="mb-1 block text-sm font-medium text-gray-700 dark:text-gray-300">{{ accountForm.type === 'alipay' ? t('affiliate.paymentAccounts.alipayAccount') : t('affiliate.paymentAccounts.cardNumber') }}</label>
          <input v-model.trim="accountForm.account_number" class="input" required autocomplete="off" />
        </div>
        <template v-else>
          <div>
            <label class="mb-1 block text-sm font-medium text-gray-700 dark:text-gray-300">{{ t('affiliate.paymentAccounts.network') }}</label>
            <select v-model="accountForm.usdt_network" class="input" required>
              <option value="TRC20">TRC20</option>
              <option value="ERC20">ERC20</option>
              <option value="BEP20">BEP20</option>
            </select>
          </div>
          <div>
            <label class="mb-1 block text-sm font-medium text-gray-700 dark:text-gray-300">{{ t('affiliate.paymentAccounts.walletAddress') }}</label>
            <input v-model.trim="accountForm.wallet_address" class="input font-mono" required autocomplete="off" />
          </div>
        </template>
        <label class="flex items-center gap-2 text-sm text-gray-700 dark:text-gray-300">
          <input v-model="accountForm.is_default" type="checkbox" class="rounded border-gray-300 text-primary-600" />
          {{ t('affiliate.paymentAccounts.setDefault') }}
        </label>
      </form>
      <template #footer>
        <button class="btn btn-secondary" :disabled="accountSaving" @click="closeAccountDialog">{{ t('common.cancel') }}</button>
        <button class="btn btn-primary" :disabled="accountSaving" @click="savePaymentAccount">
          <Icon v-if="accountSaving" name="refresh" size="sm" class="animate-spin" />
          {{ t('common.save') }}
        </button>
      </template>
    </BaseDialog>
  </AppLayout>
</template>

<script setup lang="ts">
import { computed, defineComponent, h, onMounted, reactive, ref, type PropType } from 'vue'
import { useI18n } from 'vue-i18n'
import AppLayout from '@/components/layout/AppLayout.vue'
import BaseDialog from '@/components/common/BaseDialog.vue'
import Pagination from '@/components/common/Pagination.vue'
import Icon from '@/components/icons/Icon.vue'
import userAPI from '@/api/user'
import type {
  AffiliatePaymentAccount,
  AffiliatePaymentAccountRequest,
  AffiliatePaymentAccountType,
  AffiliateWithdrawal,
  AffiliateWithdrawalStatus,
  UserAffiliateDetail,
} from '@/types'
import { useAppStore } from '@/stores/app'
import { useClipboard } from '@/composables/useClipboard'
import { formatCurrency, formatDateTime } from '@/utils/format'
import { extractApiErrorMessage } from '@/utils/apiError'

const { t } = useI18n()
const appStore = useAppStore()
const { copyToClipboard } = useClipboard()

const loading = ref(true)
const accountsLoading = ref(false)
const withdrawalsLoading = ref(false)
const accountSaving = ref(false)
const withdrawalSubmitting = ref(false)
const cancelingWithdrawalId = ref<number | null>(null)
const detail = ref<UserAffiliateDetail | null>(null)
const paymentAccounts = ref<AffiliatePaymentAccount[]>([])
const withdrawals = ref<AffiliateWithdrawal[]>([])
const withdrawalPagination = reactive({ page: 1, page_size: 20, total: 0 })

const accountDialogOpen = ref(false)
const editingAccountId = ref<number | null>(null)
const accountForm = reactive<AffiliatePaymentAccountRequest>({
  type: 'alipay', account_name: '', account_number: '', bank_name: '', usdt_network: 'TRC20', wallet_address: '', is_default: false,
})
const withdrawalForm = reactive({ payment_account_id: 0, amount: 10 })
let withdrawalRequestKey = ''
let withdrawalRequestFingerprint = ''

const availableCommission = computed(() => Number(detail.value?.available_commission ?? detail.value?.aff_quota ?? 0))
const frozenCommission = computed(() => Number(detail.value?.frozen_commission ?? detail.value?.aff_frozen_quota ?? 0))
const withdrawalReserved = computed(() => Number(detail.value?.withdrawal_reserved ?? 0))
const debt = computed(() => Number(detail.value?.debt ?? 0))
const minimumWithdrawal = computed(() => Number(detail.value?.minimum_withdrawal ?? 10))
const inviteLink = computed(() => {
  const code = detail.value?.aff_code || ''
  if (typeof window === 'undefined') return `/register?aff=${encodeURIComponent(code)}`
  return `${window.location.origin}/register?aff=${encodeURIComponent(code)}`
})
const formattedRebateRate = computed(() => {
  const rounded = Math.round(Number(detail.value?.effective_rebate_rate_percent ?? 0) * 100) / 100
  return Number.isInteger(rounded) ? String(rounded) : rounded.toString()
})
const canWithdraw = computed(() =>
  debt.value <= 0 &&
  withdrawalForm.payment_account_id > 0 &&
  withdrawalForm.amount >= minimumWithdrawal.value &&
  withdrawalForm.amount <= availableCommission.value,
)

const StatCard = defineComponent({
  props: {
    label: { type: String, required: true },
    value: { type: String, required: true },
    tone: { type: String as PropType<'default' | 'green' | 'amber' | 'red'>, default: 'default' },
  },
  setup(props) {
    const colors = { default: 'text-gray-900 dark:text-white', green: 'text-emerald-600 dark:text-emerald-400', amber: 'text-amber-600 dark:text-amber-400', red: 'text-red-600 dark:text-red-400' }
    return () => h('div', { class: 'card p-5' }, [
      h('p', { class: 'text-sm text-gray-500 dark:text-dark-400' }, props.label),
      h('p', { class: `mt-2 text-2xl font-semibold ${colors[props.tone]}` }, props.value),
    ])
  },
})

function newIdempotencyKey(prefix: string): string {
  return `${prefix}-${globalThis.crypto?.randomUUID?.() || `${Date.now()}-${Math.random().toString(36).slice(2)}`}`
}

async function loadAffiliateDetail(silent = false): Promise<void> {
  if (!silent) loading.value = true
  try {
    detail.value = await userAPI.getAffiliateDetail()
  } catch (error) {
    appStore.showError(extractApiErrorMessage(error, t('affiliate.loadFailed')))
  } finally {
    if (!silent) loading.value = false
  }
}

async function loadPaymentAccounts(): Promise<void> {
  accountsLoading.value = true
  try {
    paymentAccounts.value = await userAPI.listAffiliatePaymentAccounts()
    const preferred = paymentAccounts.value.find((item) => item.is_default) || paymentAccounts.value[0]
    if (preferred && !paymentAccounts.value.some((item) => item.id === withdrawalForm.payment_account_id)) {
      withdrawalForm.payment_account_id = preferred.id
    }
  } catch (error) {
    appStore.showError(extractApiErrorMessage(error, t('affiliate.paymentAccounts.loadFailed')))
  } finally {
    accountsLoading.value = false
  }
}

async function loadWithdrawals(): Promise<void> {
  withdrawalsLoading.value = true
  try {
    const response = await userAPI.listAffiliateWithdrawals(withdrawalPagination.page, withdrawalPagination.page_size)
    withdrawals.value = response.items
    withdrawalPagination.total = response.total
  } catch (error) {
    appStore.showError(extractApiErrorMessage(error, t('affiliate.withdrawal.loadFailed')))
  } finally {
    withdrawalsLoading.value = false
  }
}

function resetAccountForm(type: AffiliatePaymentAccountType = 'alipay'): void {
  Object.assign(accountForm, { type, account_name: '', account_number: '', bank_name: '', usdt_network: 'TRC20', wallet_address: '', is_default: false })
}

function openAccountDialog(account?: AffiliatePaymentAccount): void {
  editingAccountId.value = account?.id ?? null
  resetAccountForm(account?.type ?? 'alipay')
  if (account) accountForm.is_default = account.is_default
  accountDialogOpen.value = true
}

function closeAccountDialog(): void {
  if (accountSaving.value) return
  accountDialogOpen.value = false
  editingAccountId.value = null
}

async function savePaymentAccount(): Promise<void> {
  if ((accountForm.type !== 'usdt' && !accountForm.account_name.trim()) || accountSaving.value) return
  accountSaving.value = true
  try {
    const payload: AffiliatePaymentAccountRequest = {
      type: accountForm.type,
      account_name: accountForm.account_name.trim(),
      is_default: accountForm.is_default,
      ...(accountForm.type === 'usdt'
        ? { usdt_network: accountForm.usdt_network?.trim(), wallet_address: accountForm.wallet_address?.trim() }
        : { account_number: accountForm.account_number?.trim(), ...(accountForm.type === 'bank_card' ? { bank_name: accountForm.bank_name?.trim() } : {}) }),
    }
    if (editingAccountId.value) await userAPI.updateAffiliatePaymentAccount(editingAccountId.value, payload)
    else await userAPI.createAffiliatePaymentAccount(payload)
    appStore.showSuccess(t('affiliate.paymentAccounts.saved'))
    accountDialogOpen.value = false
    editingAccountId.value = null
    await loadPaymentAccounts()
  } catch (error) {
    appStore.showError(extractApiErrorMessage(error, t('affiliate.paymentAccounts.saveFailed')))
  } finally {
    accountSaving.value = false
  }
}

async function removePaymentAccount(account: AffiliatePaymentAccount): Promise<void> {
  if (!window.confirm(t('affiliate.paymentAccounts.deleteConfirm'))) return
  try {
    await userAPI.deleteAffiliatePaymentAccount(account.id)
    appStore.showSuccess(t('affiliate.paymentAccounts.deleted'))
    await loadPaymentAccounts()
  } catch (error) {
    appStore.showError(extractApiErrorMessage(error, t('affiliate.paymentAccounts.deleteFailed')))
  }
}

async function submitWithdrawal(): Promise<void> {
  if (!canWithdraw.value || withdrawalSubmitting.value) return
  withdrawalSubmitting.value = true
  try {
    const payload = { payment_account_id: withdrawalForm.payment_account_id, amount: Number(withdrawalForm.amount) }
    const fingerprint = `${payload.payment_account_id}:${payload.amount.toFixed(8)}`
    if (!withdrawalRequestKey || withdrawalRequestFingerprint !== fingerprint) {
      withdrawalRequestKey = newIdempotencyKey('affiliate-withdrawal')
      withdrawalRequestFingerprint = fingerprint
    }
    await userAPI.createAffiliateWithdrawal(payload, withdrawalRequestKey)
    withdrawalRequestKey = ''
    withdrawalRequestFingerprint = ''
    appStore.showSuccess(t('affiliate.withdrawal.submitted'))
    withdrawalPagination.page = 1
    await Promise.all([loadAffiliateDetail(true), loadWithdrawals()])
    withdrawalForm.amount = minimumWithdrawal.value
  } catch (error) {
    appStore.showError(extractApiErrorMessage(error, t('affiliate.withdrawal.submitFailed')))
  } finally {
    withdrawalSubmitting.value = false
  }
}

async function cancelWithdrawal(item: AffiliateWithdrawal): Promise<void> {
  if (item.status !== 'submitted' || cancelingWithdrawalId.value !== null) return
  cancelingWithdrawalId.value = item.id
  try {
    await userAPI.cancelAffiliateWithdrawal(item.id)
    appStore.showSuccess(t('affiliate.withdrawal.canceled'))
    await Promise.all([loadAffiliateDetail(true), loadWithdrawals()])
  } catch (error) {
    appStore.showError(extractApiErrorMessage(error, t('affiliate.withdrawal.cancelFailed')))
  } finally {
    cancelingWithdrawalId.value = null
  }
}

function changeWithdrawalPage(page: number): void { withdrawalPagination.page = page; void loadWithdrawals() }
function changeWithdrawalPageSize(pageSize: number): void { withdrawalPagination.page = 1; withdrawalPagination.page_size = pageSize; void loadWithdrawals() }
function paymentAccountTypeLabel(type: AffiliatePaymentAccountType): string { return t(`affiliate.paymentAccounts.types.${type}`) }
function withdrawalStatusLabel(status: AffiliateWithdrawalStatus): string { return t(`affiliate.withdrawal.statuses.${status}`) }
function withdrawalStatusClass(status: AffiliateWithdrawalStatus): string {
  const classes: Record<AffiliateWithdrawalStatus, string> = { submitted: 'badge badge-warning', approved: 'badge badge-primary', paid: 'badge badge-success', rejected: 'badge badge-danger', canceled: 'badge badge-gray' }
  return classes[status]
}
async function copyCode(): Promise<void> { if (detail.value?.aff_code) await copyToClipboard(detail.value.aff_code, t('affiliate.codeCopied')) }
async function copyInviteLink(): Promise<void> { if (inviteLink.value) await copyToClipboard(inviteLink.value, t('affiliate.linkCopied')) }

onMounted(async () => {
  await loadAffiliateDetail()
  if (detail.value?.status === 'active') {
    withdrawalForm.amount = minimumWithdrawal.value
    await Promise.all([loadPaymentAccounts(), loadWithdrawals()])
  }
})
</script>
