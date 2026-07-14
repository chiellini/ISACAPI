<template>
  <BaseDialog
    :show="show"
    :title="isEditing ? t('provider.form.editTitle') : t('provider.form.createTitle')"
    width="wide"
    @close="emit('close')"
  >
    <form id="provider-account-form" class="space-y-5" @submit.prevent="submit">
      <div class="grid gap-4 md:grid-cols-2">
        <div>
          <label class="input-label">{{ t('provider.form.name') }}</label>
          <input v-model="form.name" required class="input" />
        </div>
        <div>
          <label class="input-label">{{ t('provider.form.concurrency') }}</label>
          <input v-model.number="form.concurrency" type="number" min="1" required class="input" />
        </div>
        <div>
          <label class="input-label">{{ t('provider.form.platform') }}</label>
          <select v-model="form.platform" class="input" :disabled="isEditing">
            <option v-for="platform in platforms" :key="platform" :value="platform">
              {{ platformLabel(platform) }}
            </option>
          </select>
        </div>
        <div>
          <label class="input-label">{{ t('provider.form.type') }}</label>
          <select v-model="form.type" class="input" :disabled="isEditing">
            <option v-for="type in compatibleTypes" :key="type" :value="type">{{ type }}</option>
          </select>
        </div>
        <div v-if="isEditing">
          <label class="input-label">{{ t('provider.form.status') }}</label>
          <select v-model="form.status" class="input">
            <option value="active">{{ t('common.active') }}</option>
            <option value="inactive">{{ t('common.inactive') }}</option>
          </select>
        </div>
        <div class="md:col-span-2">
          <label class="input-label">{{ t('provider.form.notes') }}</label>
          <textarea v-model="form.notes" rows="2" class="input"></textarea>
        </div>
      </div>

      <div class="rounded-xl border border-gray-200 p-4 dark:border-dark-700">
        <div class="mb-3 flex items-center justify-between gap-3">
          <div>
            <div class="font-medium text-gray-900 dark:text-white">{{ t('provider.form.credentials') }}</div>
            <p class="mt-1 text-xs text-gray-500 dark:text-dark-400">{{ t('provider.form.credentialsHint') }}</p>
          </div>
          <label v-if="isEditing" class="flex items-center gap-2 text-sm text-gray-700 dark:text-gray-300">
            <input v-model="replaceCredentials" type="checkbox" class="rounded border-gray-300 text-primary-600" />
            {{ t('provider.form.replaceCredentials') }}
          </label>
        </div>

        <div v-if="!isEditing || replaceCredentials" class="space-y-3">
          <div v-if="form.type === 'apikey' || form.type === 'upstream'">
            <label class="input-label">{{ t('provider.form.baseUrl') }}</label>
            <input v-model="baseUrl" type="url" class="input" placeholder="https://api.example.com" />
          </div>
          <textarea
            v-model="credentialInput"
            rows="8"
            required
            spellcheck="false"
            autocomplete="off"
            class="input font-mono text-xs"
            :placeholder="t('provider.form.credentialsPlaceholder')"
          ></textarea>
          <label class="btn btn-secondary inline-flex cursor-pointer">
            {{ t('provider.form.uploadFile') }}
            <input type="file" accept=".json,application/json,text/plain" class="hidden" @change="loadCredentialFile" />
          </label>
        </div>
        <p class="mt-3 text-xs text-amber-700 dark:text-amber-300">{{ t('provider.form.secretsNotice') }}</p>
      </div>

      <div>
        <label class="input-label">{{ t('provider.form.groups') }}</label>
        <div v-if="compatibleGroups.length" class="grid gap-2 sm:grid-cols-2 lg:grid-cols-3">
          <label
            v-for="group in compatibleGroups"
            :key="group.id"
            class="flex cursor-pointer items-start gap-2 rounded-lg border border-gray-200 p-3 text-sm dark:border-dark-700"
          >
            <input v-model="form.group_ids" type="checkbox" :value="group.id" class="mt-0.5 rounded border-gray-300 text-primary-600" />
            <span>
              <span class="block font-medium text-gray-900 dark:text-white">{{ group.name }}</span>
              <span v-if="group.description" class="text-xs text-gray-500 dark:text-dark-400">{{ group.description }}</span>
            </span>
          </label>
        </div>
        <p v-else class="text-sm text-gray-500 dark:text-dark-400">{{ t('provider.form.noCompatibleGroups') }}</p>
      </div>
    </form>

    <template #footer>
      <div class="flex justify-end gap-3">
        <button type="button" class="btn btn-secondary" @click="emit('close')">{{ t('common.cancel') }}</button>
        <button type="submit" form="provider-account-form" class="btn btn-primary" :disabled="saving">
          {{ saving ? t('provider.form.saving') : t('common.save') }}
        </button>
      </div>
    </template>
  </BaseDialog>
</template>

<script setup lang="ts">
import { computed, reactive, ref, watch } from 'vue'
import { useI18n } from 'vue-i18n'
import BaseDialog from '@/components/common/BaseDialog.vue'
import { providerAPI, type ProviderGroupSummary } from '@/api/provider'
import { useAppStore } from '@/stores/app'
import { normalizeProviderCredentials, ProviderCredentialError } from '@/utils/providerCredentials'
import type { Account, AccountPlatform, AccountType } from '@/types'

const props = defineProps<{
  show: boolean
  account?: Account | null
  groups: ProviderGroupSummary[]
}>()
const emit = defineEmits<{
  close: []
  saved: [account: Account]
}>()
const { t } = useI18n()
const appStore = useAppStore()

const platforms: AccountPlatform[] = ['anthropic', 'openai', 'gemini', 'antigravity', 'grok']
const typesByPlatform: Record<AccountPlatform, AccountType[]> = {
  anthropic: ['oauth', 'setup-token', 'apikey', 'bedrock', 'service_account'],
  openai: ['oauth', 'apikey', 'upstream'],
  gemini: ['oauth', 'apikey', 'upstream', 'service_account'],
  antigravity: ['oauth', 'upstream'],
  grok: ['oauth', 'apikey'],
}

const form = reactive({
  name: '',
  notes: '',
  platform: 'anthropic' as AccountPlatform,
  type: 'oauth' as AccountType,
  concurrency: 1,
  status: 'active' as 'active' | 'inactive',
  group_ids: [] as number[],
})
const credentialInput = ref('')
const baseUrl = ref('')
const replaceCredentials = ref(false)
const saving = ref(false)
const isEditing = computed(() => Boolean(props.account))
const compatibleTypes = computed(() => typesByPlatform[form.platform])
const compatibleGroups = computed(() =>
  props.groups.filter((group) => group.status === 'active' && group.platform === form.platform)
)

const platformLabel = (platform: AccountPlatform) => {
  if (platform === 'openai') return 'OpenAI'
  if (platform === 'anthropic') return 'Anthropic'
  if (platform === 'antigravity') return 'Antigravity'
  if (platform === 'grok') return 'Grok'
  return 'Gemini'
}

const reset = () => {
  const account = props.account
  form.name = account?.name ?? ''
  form.notes = account?.notes ?? ''
  form.platform = account?.platform ?? 'anthropic'
  form.type = account?.type ?? 'oauth'
  form.concurrency = account?.concurrency ?? 1
  form.status = account?.status === 'inactive' ? 'inactive' : 'active'
  form.group_ids = [...(account?.group_ids ?? account?.groups?.map((group) => group.id) ?? [])]
  credentialInput.value = ''
  baseUrl.value = ''
  replaceCredentials.value = !account
}

watch(() => props.show, (show) => { if (show) reset() }, { immediate: true })
watch(() => props.account, () => { if (props.show) reset() })
watch(() => form.platform, () => {
  if (!compatibleTypes.value.includes(form.type)) form.type = compatibleTypes.value[0]
  const allowed = new Set(compatibleGroups.value.map((group) => group.id))
  form.group_ids = form.group_ids.filter((id) => allowed.has(id))
})

const loadCredentialFile = async (event: Event) => {
  const input = event.target as HTMLInputElement
  const file = input.files?.[0]
  if (file) credentialInput.value = await file.text()
  input.value = ''
}

const errorMessage = (error: unknown, fallback: string) => {
  if (error && typeof error === 'object' && 'message' in error && typeof error.message === 'string') return error.message
  return fallback
}

const submit = async () => {
  if (!form.name.trim()) {
    appStore.showError(t('provider.form.nameRequired'))
    return
  }
  saving.value = true
  try {
    let credentials: Record<string, unknown> | undefined
    if (!isEditing.value || replaceCredentials.value) {
      credentials = normalizeProviderCredentials(credentialInput.value, form.type)
      if (baseUrl.value.trim() && (form.type === 'apikey' || form.type === 'upstream')) {
        credentials.base_url = baseUrl.value.trim()
      }
    }

    const saved = props.account
      ? await providerAPI.updateAccount(props.account.id, {
          name: form.name,
          notes: form.notes || null,
          concurrency: form.concurrency,
          status: form.status,
          group_ids: form.group_ids,
          ...(credentials ? { credentials } : {}),
        })
      : await providerAPI.createAccount({
          name: form.name,
          notes: form.notes || null,
          platform: form.platform,
          type: form.type,
          credentials: credentials ?? {},
          concurrency: form.concurrency,
          group_ids: form.group_ids,
        })

    appStore.showSuccess(t(props.account ? 'provider.accounts.updated' : 'provider.accounts.created'))
    emit('saved', saved)
    emit('close')
  } catch (error) {
    appStore.showError(
      error instanceof ProviderCredentialError
        ? t('provider.form.credentialsInvalid')
        : errorMessage(error, t('provider.accounts.saveFailed'))
    )
  } finally {
    saving.value = false
  }
}
</script>

