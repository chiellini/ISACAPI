<template>
  <AppLayout>
    <div class="space-y-5">
      <div class="flex flex-wrap items-start justify-between gap-3">
        <div>
          <h1 class="text-2xl font-semibold text-gray-900 dark:text-white">{{ t('provider.accounts.title') }}</h1>
          <p class="mt-1 text-sm text-gray-500 dark:text-dark-400">{{ t('provider.accounts.description') }}</p>
        </div>
        <button class="btn btn-primary" @click="openCreate">
          <Icon name="plus" size="sm" class="mr-1.5" />
          {{ t('provider.accounts.addAccount') }}
        </button>
      </div>

      <div class="rounded-xl border border-blue-200 bg-blue-50 px-4 py-3 text-sm text-blue-800 dark:border-blue-800/60 dark:bg-blue-950/30 dark:text-blue-200">
        {{ t('provider.accounts.schedulingNotice') }}
      </div>

      <div class="card p-4">
        <div class="flex flex-wrap gap-3">
          <div class="relative min-w-[15rem] flex-1">
            <Icon name="search" size="sm" class="absolute left-3 top-1/2 -translate-y-1/2 text-gray-400" />
            <input
              v-model="filters.search"
              class="input pl-9"
              :placeholder="t('provider.accounts.searchPlaceholder')"
              @keyup.enter="applyFilters"
            />
          </div>
          <select v-model="filters.platform" class="input w-auto min-w-40" @change="applyFilters">
            <option value="">{{ t('provider.accounts.allPlatforms') }}</option>
            <option v-for="platform in platforms" :key="platform" :value="platform">{{ platform }}</option>
          </select>
          <select v-model="filters.status" class="input w-auto min-w-40" @change="applyFilters">
            <option value="">{{ t('provider.accounts.allStatuses') }}</option>
            <option value="active">{{ t('common.active') }}</option>
            <option value="inactive">{{ t('common.inactive') }}</option>
            <option value="error">{{ t('common.error') }}</option>
          </select>
          <button class="btn btn-secondary" :disabled="loading" @click="applyFilters">
            <Icon name="refresh" size="sm" :class="loading ? 'animate-spin' : ''" />
          </button>
        </div>
      </div>

      <div class="card overflow-hidden">
        <DataTable :columns="columns" :data="accounts" :loading="loading" row-key="id">
          <template #cell-name="{ row }">
            <div>
              <div class="font-medium text-gray-900 dark:text-white">{{ row.name }}</div>
              <div v-if="row.notes" class="max-w-56 truncate text-xs text-gray-500" :title="row.notes">{{ row.notes }}</div>
            </div>
          </template>
          <template #cell-platform="{ row }">
            <PlatformTypeBadge :platform="row.platform" :type="row.type" />
          </template>
          <template #cell-status="{ row }">
            <span :class="statusClass(row.status)">{{ statusLabel(row.status) }}</span>
          </template>
          <template #cell-groups="{ row }">
            <AccountGroupsCell :groups="groupsForAccount(row)" :max-display="4" />
          </template>
          <template #cell-concurrency="{ row }">
            <span class="font-mono">{{ row.concurrency }}</span>
          </template>
          <template #cell-last_used_at="{ value }">
            <span class="text-sm text-gray-500">{{ value ? formatDateTime(value) : '-' }}</span>
          </template>
          <template #cell-actions="{ row }">
            <div class="flex items-center gap-2">
              <button class="btn btn-ghost btn-sm" @click="openEdit(row)">{{ t('common.edit') }}</button>
              <button class="btn btn-ghost btn-sm text-red-600" @click="openDelete(row)">{{ t('common.delete') }}</button>
            </div>
          </template>
          <template #empty>
            <div class="py-10 text-center">
              <p class="font-medium text-gray-900 dark:text-white">{{ t('provider.accounts.emptyTitle') }}</p>
              <p class="mt-1 text-sm text-gray-500">{{ t('provider.accounts.emptyDescription') }}</p>
            </div>
          </template>
        </DataTable>
        <Pagination
          v-if="pagination.total > 0"
          :total="pagination.total"
          :page="pagination.page"
          :page-size="pagination.page_size"
          @update:page="changePage"
          @update:page-size="changePageSize"
        />
      </div>
    </div>

    <ProviderAccountFormModal
      :show="showForm"
      :account="editingAccount"
      :groups="groups"
      @close="showForm = false"
      @saved="handleSaved"
    />
    <ConfirmDialog
      :show="Boolean(deletingAccount)"
      :title="t('provider.accounts.deleteTitle')"
      :message="t('provider.accounts.deleteConfirm', { name: deletingAccount?.name })"
      :danger="true"
      @confirm="confirmDelete"
      @cancel="deletingAccount = null"
    />
  </AppLayout>
</template>

<script setup lang="ts">
import { computed, onMounted, reactive, ref } from 'vue'
import { useI18n } from 'vue-i18n'
import AppLayout from '@/components/layout/AppLayout.vue'
import DataTable from '@/components/common/DataTable.vue'
import Pagination from '@/components/common/Pagination.vue'
import ConfirmDialog from '@/components/common/ConfirmDialog.vue'
import Icon from '@/components/icons/Icon.vue'
import PlatformTypeBadge from '@/components/common/PlatformTypeBadge.vue'
import AccountGroupsCell from '@/components/account/AccountGroupsCell.vue'
import ProviderAccountFormModal from '@/components/provider/ProviderAccountFormModal.vue'
import { providerAPI, type ProviderGroupSummary } from '@/api/provider'
import { useAppStore } from '@/stores/app'
import { formatDateTime } from '@/utils/format'
import type { Account, AccountPlatform, Group } from '@/types'
import type { Column } from '@/components/common/types'

const { t } = useI18n()
const appStore = useAppStore()
const platforms: AccountPlatform[] = ['anthropic', 'openai', 'gemini', 'antigravity', 'grok']
const accounts = ref<Account[]>([])
const groups = ref<ProviderGroupSummary[]>([])
const loading = ref(false)
const showForm = ref(false)
const editingAccount = ref<Account | null>(null)
const deletingAccount = ref<Account | null>(null)
const filters = reactive({ search: '', platform: '' as AccountPlatform | '', status: '' as Account['status'] | '' })
const pagination = reactive({ page: 1, page_size: 20, total: 0 })

const columns = computed<Column[]>(() => [
  { key: 'name', label: t('provider.accounts.columns.name') },
  { key: 'platform', label: t('provider.accounts.columns.platform') },
  { key: 'status', label: t('provider.accounts.columns.status') },
  { key: 'groups', label: t('provider.accounts.columns.groups') },
  { key: 'concurrency', label: t('provider.accounts.columns.concurrency') },
  { key: 'last_used_at', label: t('provider.accounts.columns.lastUsed') },
  { key: 'actions', label: t('provider.accounts.columns.actions') },
])

const load = async () => {
  loading.value = true
  try {
    const result = await providerAPI.listAccounts({
      page: pagination.page,
      page_size: pagination.page_size,
      search: filters.search.trim() || undefined,
      platform: filters.platform || undefined,
      status: filters.status || undefined,
    })
    accounts.value = result.items
    pagination.total = result.total
  } catch {
    appStore.showError(t('provider.accounts.loadFailed'))
  } finally {
    loading.value = false
  }
}

const loadGroups = async () => {
  try {
    groups.value = await providerAPI.listGroups()
  } catch {
    appStore.showError(t('provider.accounts.loadFailed'))
  }
}

const applyFilters = () => { pagination.page = 1; void load() }
const changePage = (page: number) => { pagination.page = page; void load() }
const changePageSize = (size: number) => { pagination.page_size = size; pagination.page = 1; void load() }
const openCreate = () => { editingAccount.value = null; showForm.value = true }
const openEdit = (account: Account) => { editingAccount.value = account; showForm.value = true }
const openDelete = (account: Account) => { deletingAccount.value = account }
const handleSaved = () => { void Promise.all([load(), loadGroups()]) }

const confirmDelete = async () => {
  if (!deletingAccount.value) return
  try {
    await providerAPI.deleteAccount(deletingAccount.value.id)
    appStore.showSuccess(t('provider.accounts.deleted'))
    deletingAccount.value = null
    await load()
  } catch {
    appStore.showError(t('provider.accounts.deleteFailed'))
  }
}

const groupsForAccount = (account: Account): Group[] => {
  if (account.groups?.length) return account.groups
  const ids = new Set(account.group_ids ?? [])
  return groups.value
    .filter((group) => ids.has(group.id))
    .map((group) => ({ ...group, subscription_type: 'standard', rate_multiplier: 1 } as Group))
}

const statusLabel = (status: Account['status']) => t(`common.${status}`)
const statusClass = (status: Account['status']) => [
  'badge',
  status === 'active' ? 'badge-green' : status === 'error' ? 'badge-red' : 'badge-gray',
]

onMounted(() => { void Promise.all([load(), loadGroups()]) })
</script>

