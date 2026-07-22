<template>
  <BaseDialog
    :show="show"
    :title="isEditing ? t('provider.form.editTitle') : t('provider.form.createTitle')"
    width="wide"
    @close="emit('close')"
  >
    <form v-if="step === 1" id="provider-account-form" class="space-y-5" @submit.prevent="submit">
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
          <p v-if="oauthFlowSupported" class="rounded-lg bg-blue-50 px-3 py-2 text-xs text-blue-700 dark:bg-blue-900/30 dark:text-blue-300">
            {{ t('provider.form.oauthHint') }}
          </p>
          <div v-if="form.type === 'apikey' || form.type === 'upstream'">
            <label class="input-label">{{ t('provider.form.baseUrl') }}</label>
            <input v-model="baseUrl" type="url" class="input" placeholder="https://api.example.com" />
          </div>
          <textarea
            v-model="credentialInput"
            rows="8"
            :required="!oauthFlowSupported"
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

    <!-- Step 2: OAuth authorization, mirroring the admin create-account flow -->
    <div v-else class="space-y-4">
      <div v-if="form.platform === 'gemini'">
        <label class="input-label">{{ t('provider.form.geminiOauthType') }}</label>
        <select v-model="geminiOauthType" class="input">
          <option value="code_assist">Code Assist</option>
          <option value="google_one">Google One</option>
          <option value="ai_studio">AI Studio</option>
        </select>
      </div>
      <OAuthAuthorizationFlow
        ref="oauthFlowRef"
        :add-method="oauthAddMethod"
        :auth-url="oauthAuthUrl"
        :session-id="oauthSessionId"
        :loading="oauthLoading"
        :error="oauthError"
        :show-help="form.platform === 'anthropic'"
        :show-proxy-warning="false"
        :allow-multiple="false"
        :show-cookie-option="false"
        :show-refresh-token-option="false"
        :show-mobile-refresh-token-option="false"
        :show-session-token-option="false"
        :show-access-token-option="false"
        :show-codex-session-import-option="false"
        :show-agent-identity-option="false"
        :show-codex-pat-option="false"
        :show-sso-option="false"
        :show-manual-option="true"
        :initial-input-method="'manual'"
        :platform="form.platform"
        :show-project-id="geminiOauthType !== 'ai_studio'"
        @generate-url="handleGenerateUrl"
      />
    </div>

    <template #footer>
      <div v-if="step === 1" class="flex justify-end gap-3">
        <button type="button" class="btn btn-secondary" @click="emit('close')">{{ t('common.cancel') }}</button>
        <button type="submit" form="provider-account-form" class="btn btn-primary" :disabled="saving">
          {{ saving ? t('provider.form.saving') : goesToOAuthStep ? t('common.next') : t('common.save') }}
        </button>
      </div>
      <div v-else class="flex justify-between gap-3">
        <button type="button" class="btn btn-secondary" @click="backToStep1">{{ t('common.back') }}</button>
        <button type="button" class="btn btn-primary" :disabled="!canExchange" @click="handleExchange">
          {{ oauthLoading ? t('provider.form.saving') : t('provider.form.oauthConfirm') }}
        </button>
      </div>
    </template>
  </BaseDialog>
</template>

<script setup lang="ts">
import { computed, reactive, ref, watch } from 'vue'
import { useI18n } from 'vue-i18n'
import BaseDialog from '@/components/common/BaseDialog.vue'
import OAuthAuthorizationFlow from '@/components/account/OAuthAuthorizationFlow.vue'
import { providerAPI, type ProviderGroupSummary } from '@/api/provider'
import { providerOAuthAPI } from '@/api/providerOauth'
import { useAppStore } from '@/stores/app'
import { normalizeProviderCredentials, ProviderCredentialError } from '@/utils/providerCredentials'
import {
  buildAntigravityOAuthCredentials,
  buildGeminiOAuthCredentials,
  buildGrokOAuthCredentials,
  buildOpenAIOAuthCredentials,
} from '@/utils/oauthCredentialBuilders'
import type { Account, AccountPlatform, AccountType } from '@/types'

interface OAuthFlowExposed {
  authCode: string
  oauthState: string
  projectId: string
  reset: () => void
}

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
const step = ref(1)
const isEditing = computed(() => Boolean(props.account))
const compatibleTypes = computed(() => typesByPlatform[form.platform])
const compatibleGroups = computed(() =>
  props.groups.filter((group) => group.status === 'active' && group.platform === form.platform)
)

// OAuth link flow state (create mode only)
const oauthFlowRef = ref<OAuthFlowExposed | null>(null)
const oauthAuthUrl = ref('')
const oauthSessionId = ref('')
const oauthServerState = ref('')
const oauthLoading = ref(false)
const oauthError = ref('')
const geminiOauthType = ref<'code_assist' | 'google_one' | 'ai_studio'>('code_assist')

const oauthFlowSupported = computed(
  () => !isEditing.value && (form.type === 'oauth' || (form.platform === 'anthropic' && form.type === 'setup-token'))
)
const goesToOAuthStep = computed(() => oauthFlowSupported.value && !credentialInput.value.trim())
const oauthAddMethod = computed(() =>
  form.platform === 'anthropic' && form.type === 'setup-token' ? 'setup-token' : 'oauth'
)
const canExchange = computed(() =>
  Boolean(!oauthLoading.value && oauthSessionId.value && (oauthFlowRef.value?.authCode || '').trim())
)

const platformLabel = (platform: AccountPlatform) => {
  if (platform === 'openai') return 'OpenAI'
  if (platform === 'anthropic') return 'Anthropic'
  if (platform === 'antigravity') return 'Antigravity'
  if (platform === 'grok') return 'Grok'
  return 'Gemini'
}

const resetOAuthState = () => {
  oauthAuthUrl.value = ''
  oauthSessionId.value = ''
  oauthServerState.value = ''
  oauthLoading.value = false
  oauthError.value = ''
  oauthFlowRef.value?.reset()
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
  step.value = 1
  geminiOauthType.value = 'code_assist'
  resetOAuthState()
}

watch(() => props.show, (show) => { if (show) reset() }, { immediate: true })
watch(() => props.account, () => { if (props.show) reset() })
watch(() => form.platform, () => {
  if (!compatibleTypes.value.includes(form.type)) form.type = compatibleTypes.value[0]
  const allowed = new Set(compatibleGroups.value.map((group) => group.id))
  form.group_ids = form.group_ids.filter((id) => allowed.has(id))
})
watch([() => form.platform, () => form.type], resetOAuthState)
// 换 OAuth 类型后旧链接的 session 参数不再匹配，必须重新生成
watch(geminiOauthType, resetOAuthState)

const loadCredentialFile = async (event: Event) => {
  const input = event.target as HTMLInputElement
  const file = input.files?.[0]
  if (file) credentialInput.value = await file.text()
  input.value = ''
}

const errorMessage = (error: unknown, fallback: string) => {
  const detail = (error as { response?: { data?: { detail?: string; message?: string } } })?.response?.data
  if (detail?.detail) return detail.detail
  if (detail?.message) return detail.message
  if (error && typeof error === 'object' && 'message' in error && typeof error.message === 'string') return error.message
  return fallback
}

const parseStateFromUrl = (url: string): string => {
  try {
    return new URL(url).searchParams.get('state') || ''
  } catch {
    return ''
  }
}

const backToStep1 = () => {
  step.value = 1
  resetOAuthState()
}

const handleGenerateUrl = async () => {
  oauthLoading.value = true
  oauthError.value = ''
  oauthAuthUrl.value = ''
  oauthSessionId.value = ''
  oauthServerState.value = ''
  try {
    let result
    if (form.platform === 'anthropic') {
      result = await providerOAuthAPI.generateAnthropicAuthUrl(form.type === 'setup-token')
    } else if (form.platform === 'openai') {
      result = await providerOAuthAPI.generateOpenAIAuthUrl()
    } else if (form.platform === 'gemini') {
      const projectId = (oauthFlowRef.value?.projectId || '').trim()
      if (geminiOauthType.value !== 'ai_studio' && !projectId) {
        oauthError.value = t('admin.accounts.oauth.gemini.missingProjectId')
        appStore.showError(oauthError.value)
        return
      }
      result = await providerOAuthAPI.generateGeminiAuthUrl({
        project_id: projectId || undefined,
        oauth_type: geminiOauthType.value,
      })
    } else if (form.platform === 'antigravity') {
      result = await providerOAuthAPI.generateAntigravityAuthUrl()
    } else {
      result = await providerOAuthAPI.generateGrokAuthUrl()
    }
    oauthAuthUrl.value = result.auth_url
    oauthSessionId.value = result.session_id
    oauthServerState.value = result.state || parseStateFromUrl(result.auth_url)
  } catch (error) {
    oauthError.value = errorMessage(error, t('provider.form.oauthGenerateFailed'))
    appStore.showError(oauthError.value)
  } finally {
    oauthLoading.value = false
  }
}

const handleExchange = async () => {
  const code = (oauthFlowRef.value?.authCode || '').trim()
  if (!code || !oauthSessionId.value) return
  oauthLoading.value = true
  oauthError.value = ''
  try {
    const sessionId = oauthSessionId.value
    const state = (oauthFlowRef.value?.oauthState || oauthServerState.value || '').trim()
    let credentials: Record<string, unknown>
    if (form.platform === 'anthropic') {
      const tokenInfo = await providerOAuthAPI.exchangeAnthropicCode(
        { session_id: sessionId, code },
        form.type === 'setup-token'
      )
      credentials = { ...tokenInfo }
    } else if (!state) {
      oauthError.value = t('provider.form.oauthMissingState')
      appStore.showError(oauthError.value)
      return
    } else if (form.platform === 'openai') {
      const tokenInfo = await providerOAuthAPI.exchangeOpenAICode({ session_id: sessionId, code, state })
      credentials = buildOpenAIOAuthCredentials(tokenInfo)
    } else if (form.platform === 'gemini') {
      const tokenInfo = await providerOAuthAPI.exchangeGeminiCode({
        session_id: sessionId,
        state,
        code,
        oauth_type: geminiOauthType.value,
      })
      credentials = buildGeminiOAuthCredentials(tokenInfo)
    } else if (form.platform === 'antigravity') {
      const tokenInfo = await providerOAuthAPI.exchangeAntigravityCode({ session_id: sessionId, state, code })
      credentials = buildAntigravityOAuthCredentials(tokenInfo)
    } else {
      const tokenInfo = await providerOAuthAPI.exchangeGrokCode({ session_id: sessionId, code, state })
      credentials = buildGrokOAuthCredentials(tokenInfo)
    }
    await createAccount(credentials)
  } catch (error) {
    oauthError.value = errorMessage(error, t('provider.form.oauthExchangeFailed'))
    appStore.showError(oauthError.value)
  } finally {
    oauthLoading.value = false
  }
}

// 自行处理保存错误,避免上层把创建失败误报成授权失败
const createAccount = async (credentials: Record<string, unknown>) => {
  saving.value = true
  try {
    const saved = await providerAPI.createAccount({
      name: form.name,
      notes: form.notes || null,
      platform: form.platform,
      type: form.type,
      credentials,
      concurrency: form.concurrency,
      group_ids: form.group_ids,
    })
    appStore.showSuccess(t('provider.accounts.created'))
    emit('saved', saved)
    emit('close')
  } catch (error) {
    appStore.showError(errorMessage(error, t('provider.accounts.saveFailed')))
  } finally {
    saving.value = false
  }
}

const submit = async () => {
  if (!form.name.trim()) {
    appStore.showError(t('provider.form.nameRequired'))
    return
  }
  if (goesToOAuthStep.value) {
    step.value = 2
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
