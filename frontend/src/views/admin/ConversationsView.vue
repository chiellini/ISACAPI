<script setup lang="ts">
import { ref, reactive, computed, onMounted } from 'vue'
import { useI18n } from 'vue-i18n'
import { useAppStore } from '@/stores/app'
import { adminAPI } from '@/api/admin'
import type {
  ConversationSessionView,
  ConversationSessionDetail,
  ConversationEventView
} from '@/api/admin/conversations'
import type { Column } from '@/components/common/types'
import AppLayout from '@/components/layout/AppLayout.vue'
import TablePageLayout from '@/components/layout/TablePageLayout.vue'
import DataTable from '@/components/common/DataTable.vue'
import Pagination from '@/components/common/Pagination.vue'
import BaseDialog from '@/components/common/BaseDialog.vue'
import ConfirmDialog from '@/components/common/ConfirmDialog.vue'
import EmptyState from '@/components/common/EmptyState.vue'
import Select from '@/components/common/Select.vue'

const { t } = useI18n()
const appStore = useAppStore()

const loading = ref(false)
const sessions = ref<ConversationSessionView[]>([])
const pagination = reactive({ page: 1, page_size: 20, total: 0 })

const filters = reactive({
  user_id: '',
  status: ''
})

const statusOptions = computed(() => [
  { value: '', label: t('admin.conversations.filters.allStatuses') },
  { value: 'active', label: t('admin.conversations.status.active') },
  { value: 'archived', label: t('admin.conversations.status.archived') },
  { value: 'expired', label: t('admin.conversations.status.expired') }
])

const columns = computed<Column[]>(() => [
  { key: 'last_active_at', label: t('admin.conversations.columns.lastActive') },
  { key: 'user_id', label: t('admin.conversations.columns.user') },
  { key: 'context_domain', label: t('admin.conversations.columns.contextDomain') },
  { key: 'request_count', label: t('admin.conversations.columns.requests') },
  { key: 'tokens', label: t('admin.conversations.columns.tokens') },
  { key: 'status', label: t('admin.conversations.columns.status') },
  { key: 'actions', label: t('admin.conversations.columns.actions') }
])

function formatTime(value: string): string {
  if (!value) return '-'
  return new Date(value).toLocaleString()
}

async function load() {
  loading.value = true
  try {
    const params: Record<string, unknown> = {
      page: pagination.page,
      page_size: pagination.page_size
    }
    if (filters.user_id.trim()) params.user_id = Number(filters.user_id.trim())
    if (filters.status) params.status = filters.status
    const data = await adminAPI.conversations.list(params)
    sessions.value = data.items || []
    pagination.total = data.total || 0
  } catch (err) {
    appStore.showError(t('admin.conversations.loadFailed'))
  } finally {
    loading.value = false
  }
}

function applyFilters() {
  pagination.page = 1
  load()
}

function triggerDownload(blob: Blob, filename: string) {
  const url = window.URL.createObjectURL(blob)
  const link = document.createElement('a')
  link.href = url
  link.download = filename
  document.body.appendChild(link)
  link.click()
  document.body.removeChild(link)
  window.URL.revokeObjectURL(url)
}

function exportFilters(): Record<string, unknown> {
  const params: Record<string, unknown> = {}
  if (filters.user_id.trim()) params.user_id = Number(filters.user_id.trim())
  if (filters.status) params.status = filters.status
  return params
}

const exportingAll = ref(false)
async function exportAll() {
  exportingAll.value = true
  try {
    const blob = await adminAPI.conversations.exportAll(exportFilters())
    const today = new Date().toISOString().split('T')[0]
    triggerDownload(blob, `conversations-${today}.zip`)
  } catch (err) {
    appStore.showError(t('admin.conversations.exportFailed'))
  } finally {
    exportingAll.value = false
  }
}

const exportingOne = ref(false)
async function exportOne(id: string) {
  exportingOne.value = true
  try {
    const blob = await adminAPI.conversations.exportSession(id)
    triggerDownload(blob, `conversation-${id}.txt`)
  } catch (err) {
    appStore.showError(t('admin.conversations.exportFailed'))
  } finally {
    exportingOne.value = false
  }
}

function handlePageChange(page: number) {
  pagination.page = page
  load()
}

function handlePageSizeChange(size: number) {
  pagination.page_size = size
  pagination.page = 1
  load()
}

// ---- detail ----
const detailOpen = ref(false)
const detailLoading = ref(false)
const detail = ref<ConversationSessionDetail | null>(null)

async function openDetail(row: ConversationSessionView) {
  detailOpen.value = true
  detailLoading.value = true
  detail.value = null
  try {
    detail.value = await adminAPI.conversations.get(row.id)
  } catch (err) {
    appStore.showError(t('admin.conversations.loadDetailFailed'))
    detailOpen.value = false
  } finally {
    detailLoading.value = false
  }
}

function roleLabel(ev: ConversationEventView): string {
  return t(`admin.conversations.roles.${ev.role}`, ev.role)
}

function roleClass(role: string): string {
  switch (role) {
    case 'user':
      return 'role-user'
    case 'assistant':
      return 'role-assistant'
    case 'tool':
      return 'role-tool'
    case 'system':
      return 'role-system'
    default:
      return ''
  }
}

// ---- delete ----
const confirmOpen = ref(false)
const deleting = ref(false)
const pendingDelete = ref<ConversationSessionView | null>(null)

function askDelete(row: ConversationSessionView) {
  pendingDelete.value = row
  confirmOpen.value = true
}

async function confirmDelete() {
  if (!pendingDelete.value) return
  deleting.value = true
  try {
    await adminAPI.conversations.remove(pendingDelete.value.id)
    appStore.showSuccess(t('admin.conversations.deleteSuccess'))
    confirmOpen.value = false
    if (detail.value?.session.id === pendingDelete.value.id) {
      detailOpen.value = false
    }
    load()
  } catch (err) {
    appStore.showError(t('admin.conversations.deleteFailed'))
  } finally {
    deleting.value = false
    pendingDelete.value = null
  }
}

onMounted(load)
</script>

<template>
  <AppLayout>
    <TablePageLayout
      :title="t('admin.conversations.title')"
      :description="t('admin.conversations.description')"
    >
      <template #filters>
        <div class="filter-row">
          <input
            v-model="filters.user_id"
            type="text"
            inputmode="numeric"
            class="input"
            :placeholder="t('admin.conversations.filters.userId')"
            @keyup.enter="applyFilters"
          />
          <div class="w-40">
            <Select v-model="filters.status" :options="statusOptions" />
          </div>
          <button class="btn btn-primary" @click="applyFilters">
            {{ t('common.search') }}
          </button>
          <button class="btn" :disabled="exportingAll" @click="exportAll">
            {{ exportingAll ? t('admin.conversations.exporting') : t('admin.conversations.exportAll') }}
          </button>
        </div>
      </template>

      <template #table>
        <div class="flex min-h-0 flex-1 flex-col overflow-hidden">
          <DataTable :columns="columns" :data="sessions" :loading="loading">
            <template #cell-last_active_at="{ value }">
              <span class="text-sm">{{ formatTime(value) }}</span>
            </template>
            <template #cell-tokens="{ row }">
              <span class="text-sm">{{ row.total_input_tokens }} / {{ row.total_output_tokens }}</span>
            </template>
            <template #cell-status="{ value }">
              <span class="badge">{{ t(`admin.conversations.status.${value}`, value) }}</span>
            </template>
            <template #cell-actions="{ row }">
              <div class="actions">
                <button class="btn btn-sm" @click="openDetail(row)">
                  {{ t('admin.conversations.view') }}
                </button>
                <button class="btn btn-sm btn-danger" @click="askDelete(row)">
                  {{ t('common.delete') }}
                </button>
              </div>
            </template>
          </DataTable>
        </div>
      </template>

      <template #pagination>
        <Pagination
          :page="pagination.page"
          :total="pagination.total"
          :page-size="pagination.page_size"
          @update:page="handlePageChange"
          @update:pageSize="handlePageSizeChange"
        />
      </template>
    </TablePageLayout>

    <!-- 详情 -->
    <BaseDialog :show="detailOpen" :title="t('admin.conversations.detailTitle')" width="wide" @close="detailOpen = false">
      <div v-if="detailLoading" class="py-8 text-center text-sm text-gray-500">
        {{ t('common.loading') }}
      </div>
      <div v-else-if="detail" class="detail">
        <div class="detail-meta">
          <span>{{ t('admin.conversations.columns.user') }}: {{ detail.session.user_id }}</span>
          <span>{{ detail.session.context_domain }}</span>
          <span>{{ detail.session.protocol }}</span>
          <span>{{ formatTime(detail.session.last_active_at) }}</span>
        </div>
        <div class="detail-actions">
          <button class="btn btn-sm btn-primary" :disabled="exportingOne" @click="exportOne(detail.session.id)">
            {{ t('admin.conversations.downloadTxt') }}
          </button>
          <button class="btn btn-sm btn-danger" @click="askDelete(detail.session)">
            {{ t('admin.conversations.deleteAfterDownload') }}
          </button>
        </div>
        <EmptyState v-if="detail.events.length === 0" :description="t('admin.conversations.noEvents')" />
        <div v-else class="timeline">
          <div v-for="ev in detail.events" :key="ev.id" class="event" :class="roleClass(ev.role)">
            <div class="event-head">
              <span class="event-role">{{ roleLabel(ev) }}</span>
              <span class="event-kind">{{ ev.kind }}</span>
              <span v-if="ev.partial" class="event-partial">{{ t('admin.conversations.partial') }}</span>
              <span v-if="ev.encrypted && !ev.decrypted" class="event-locked">
                {{ t('admin.conversations.previewOnly') }}
              </span>
            </div>
            <pre class="event-content">{{ ev.content }}</pre>
          </div>
        </div>
      </div>
    </BaseDialog>

    <ConfirmDialog
      :show="confirmOpen"
      :title="t('admin.conversations.deleteTitle')"
      :message="t('admin.conversations.deleteConfirm')"
      :loading="deleting"
      @confirm="confirmDelete"
      @cancel="confirmOpen = false"
    />
  </AppLayout>
</template>

<style scoped>
.filter-row {
  display: flex;
  gap: 0.5rem;
  align-items: center;
  flex-wrap: wrap;
}
.actions {
  display: flex;
  gap: 0.5rem;
}
.detail-meta {
  display: flex;
  gap: 1rem;
  flex-wrap: wrap;
  font-size: 0.8rem;
  color: var(--text-secondary, #6b7280);
  margin-bottom: 0.5rem;
}
.detail-actions {
  display: flex;
  gap: 0.5rem;
  margin-bottom: 0.75rem;
}
.timeline {
  display: flex;
  flex-direction: column;
  gap: 0.75rem;
  max-height: 60vh;
  overflow-y: auto;
}
.event {
  border-left: 3px solid #d1d5db;
  padding: 0.25rem 0.75rem;
}
.event.role-user {
  border-left-color: #3b82f6;
}
.event.role-assistant {
  border-left-color: #10b981;
}
.event.role-tool {
  border-left-color: #f59e0b;
}
.event.role-system {
  border-left-color: #9ca3af;
}
.event-head {
  display: flex;
  gap: 0.5rem;
  align-items: center;
  font-size: 0.75rem;
  color: var(--text-secondary, #6b7280);
  margin-bottom: 0.25rem;
}
.event-role {
  font-weight: 600;
}
.event-partial,
.event-locked {
  color: #f59e0b;
}
.event-content {
  white-space: pre-wrap;
  word-break: break-word;
  font-size: 0.85rem;
  margin: 0;
}
</style>
