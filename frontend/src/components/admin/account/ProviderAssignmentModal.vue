<template>
  <BaseDialog :show="show" :title="t('provider.admin.assignTitle')" width="normal" @close="emit('close')">
    <div class="space-y-4">
      <p class="text-sm text-gray-600 dark:text-gray-300">{{ t('provider.admin.assignDescription') }}</p>
      <div v-if="account" class="rounded-lg bg-gray-50 px-3 py-2 text-sm dark:bg-dark-800">
        <span class="font-medium text-gray-900 dark:text-white">{{ account.name }}</span>
        <span class="ml-2 font-mono text-xs text-gray-500">#{{ account.id }}</span>
      </div>
      <div>
        <label class="input-label">{{ t('provider.admin.selectProvider') }}</label>
        <select v-model="selectedProvider" class="input" :disabled="loading">
          <option value="">{{ t('provider.admin.unassigned') }}</option>
          <option v-for="provider in providers" :key="provider.id" :value="String(provider.id)">
            {{ providerLabel(provider) }} (#{{ provider.id }})
          </option>
        </select>
      </div>
      <div v-if="loading" class="text-sm text-gray-500">{{ t('common.loading') }}</div>
    </div>

    <template #footer>
      <div class="flex justify-end gap-3">
        <button type="button" class="btn btn-secondary" @click="emit('close')">{{ t('common.cancel') }}</button>
        <button type="button" class="btn btn-primary" :disabled="loading || saving || !account" @click="save">
          {{ saving ? t('provider.form.saving') : t('common.save') }}
        </button>
      </div>
    </template>
  </BaseDialog>
</template>

<script setup lang="ts">
import { ref, watch } from 'vue'
import { useI18n } from 'vue-i18n'
import BaseDialog from '@/components/common/BaseDialog.vue'
import { adminAPI } from '@/api/admin'
import { useAppStore } from '@/stores/app'
import type { Account, AdminUser } from '@/types'

const props = defineProps<{ show: boolean; account: Account | null }>()
const emit = defineEmits<{ close: []; saved: [account: Account] }>()
const { t } = useI18n()
const appStore = useAppStore()
const providers = ref<AdminUser[]>([])
const selectedProvider = ref('')
const loading = ref(false)
const saving = ref(false)

const providerLabel = (provider: AdminUser) => provider.username || provider.email

// 后端校验的是 IsProvider():provider 和 admin_provider 都是合法供号方,且必须为 active
const PROVIDER_ROLES = ['provider', 'admin_provider'] as const

const loadProviders = async () => {
  loading.value = true
  try {
    const collected: AdminUser[] = []
    for (const role of PROVIDER_ROLES) {
      let page = 1
      let pages = 1
      do {
        const result = await adminAPI.users.list(page, 100, { role, status: 'active' })
        collected.push(...result.items)
        pages = result.pages || Math.ceil(result.total / result.page_size) || 1
        page += 1
      } while (page <= pages)
    }
    providers.value = collected
  } catch {
    appStore.showError(t('provider.admin.loadFailed'))
  } finally {
    loading.value = false
  }
}

watch(() => props.show, (show) => {
  if (!show) return
  selectedProvider.value = props.account?.provider_id ? String(props.account.provider_id) : ''
  void loadProviders()
})

const save = async () => {
  if (!props.account) return
  saving.value = true
  try {
    const updated = await adminAPI.accounts.assignProvider(
      props.account.id,
      selectedProvider.value ? Number(selectedProvider.value) : null
    )
    appStore.showSuccess(t('provider.admin.saved'))
    emit('saved', updated)
    emit('close')
  } catch {
    appStore.showError(t('provider.admin.saveFailed'))
  } finally {
    saving.value = false
  }
}
</script>

