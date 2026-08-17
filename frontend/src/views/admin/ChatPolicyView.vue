<template>
  <AppLayout>
    <div class="space-y-6">
      <div class="flex flex-col gap-4 lg:flex-row lg:items-start lg:justify-between">
        <div>
          <h1 class="text-2xl font-semibold text-gray-900 dark:text-white">Chat 模型策略</h1>
          <p class="mt-1 max-w-3xl text-sm text-gray-500 dark:text-gray-400">
            配置内置聊天提供的 GPT、Claude、Gemini 模型。Prompt 与 Skill 由服务端注入，普通用户无法覆盖。
          </p>
        </div>
        <button
          class="btn btn-primary"
          type="button"
          data-testid="save-chat-policy"
          :disabled="!loaded || loading || saving"
          @click="save"
        >
          {{ saving ? '保存中…' : '保存策略' }}
        </button>
      </div>

      <div
        v-if="loadError"
        class="flex flex-col gap-3 rounded-lg border border-red-200 bg-red-50 p-4 text-sm text-red-700 dark:border-red-800 dark:bg-red-950/30 dark:text-red-300 sm:flex-row sm:items-center sm:justify-between"
        role="alert"
      >
        <span>{{ loadError }}</span>
        <button
          class="btn btn-secondary shrink-0"
          type="button"
          data-testid="retry-chat-policy"
          :disabled="loading"
          @click="load"
        >
          重试
        </button>
      </div>

      <div v-if="loading" class="flex justify-center py-20">
        <div class="h-8 w-8 animate-spin rounded-full border-b-2 border-primary-600"></div>
      </div>

      <template v-else-if="loaded">
        <section class="card p-6">
          <label class="flex items-start gap-3">
            <input v-model="policy.enabled" type="checkbox" class="mt-1 h-4 w-4 rounded" />
            <span>
              <span class="block font-medium text-gray-900 dark:text-white">启用服务端 Chat 策略</span>
              <span class="mt-1 block text-sm text-gray-500 dark:text-gray-400">
                启用后必须配置一个默认模型；未配置的模型不会向用户展示，也不能通过请求绕过。
              </span>
            </span>
          </label>
        </section>

        <section class="space-y-4">
          <div class="flex items-center justify-between">
            <div>
              <h2 class="text-lg font-semibold text-gray-900 dark:text-white">模型 Profile</h2>
              <p class="text-sm text-gray-500 dark:text-gray-400">公开模型名映射到一个实际分组和上游模型。</p>
            </div>
            <button class="btn btn-secondary" type="button" @click="addProfile">新增 Profile</button>
          </div>

          <div v-if="policy.profiles.length === 0" class="card p-8 text-center text-sm text-gray-500">尚未配置模型。</div>
          <article v-for="(profile, index) in policy.profiles" :key="`${profile.id}-${index}`" class="card p-6">
            <div class="mb-5 flex items-center justify-between gap-4">
              <div class="flex flex-wrap items-center gap-4">
                <label class="flex items-center gap-2 text-sm"><input v-model="profile.enabled" type="checkbox" />启用</label>
                <label class="flex items-center gap-2 text-sm">
                  <input :checked="profile.default" type="radio" name="default-chat-profile" @change="setDefault(index)" />默认
                </label>
              </div>
              <button class="text-sm text-red-600 hover:text-red-700" type="button" @click="removeProfile(index)">删除</button>
            </div>

            <div class="grid gap-4 md:grid-cols-2 xl:grid-cols-3">
              <label class="block text-sm">显示名称<input v-model.trim="profile.name" class="input mt-1" placeholder="GPT" /></label>
              <label class="block text-sm">Profile ID<input v-model.trim="profile.id" class="input mt-1 font-mono" placeholder="gpt" /></label>
              <label class="block text-sm">
                提供商
                <select v-model="profile.provider" class="input mt-1" @change="ensureCompatibleGroup(profile)">
                  <option value="openai">OpenAI / GPT</option>
                  <option value="anthropic">Anthropic / Claude</option>
                  <option value="gemini">Google / Gemini</option>
                </select>
              </label>
              <label class="block text-sm">公开模型名<input v-model.trim="profile.public_model" class="input mt-1 font-mono" placeholder="gpt" /></label>
              <label class="block text-sm">上游模型名<input v-model.trim="profile.upstream_model" class="input mt-1 font-mono" placeholder="gpt-5" /></label>
              <label class="block text-sm">
                目标分组
                <select v-model.number="profile.group_id" class="input mt-1">
                  <option :value="0">请选择</option>
                  <option v-for="group in groupsForProvider(profile.provider)" :key="group.id" :value="group.id">
                    {{ group.name }} (#{{ group.id }})
                  </option>
                </select>
              </label>
            </div>

            <label class="mt-4 block text-sm">
              服务端 System Prompt
              <textarea v-model="profile.system_prompt" class="input mt-1 min-h-32 resize-y" placeholder="仅由超级管理员维护的可信系统指令"></textarea>
            </label>

            <div class="mt-4 grid gap-4 lg:grid-cols-2">
              <fieldset>
                <legend class="text-sm font-medium">绑定 Skill</legend>
                <div class="mt-2 flex min-h-10 flex-wrap gap-3 rounded-lg border border-gray-200 p-3 dark:border-dark-600">
                  <label v-for="skill in enabledSkills" :key="skill.id" class="flex items-center gap-2 text-sm">
                    <input v-model="profile.skill_ids" type="checkbox" :value="skill.id" />{{ skill.name }}
                  </label>
                  <span v-if="enabledSkills.length === 0" class="text-sm text-gray-400">没有可用 Skill</span>
                </div>
              </fieldset>
              <fieldset>
                <legend class="text-sm font-medium">能力</legend>
                <div class="mt-2 flex flex-wrap gap-4 rounded-lg border border-gray-200 p-3 dark:border-dark-600">
                  <label class="flex items-center gap-2 text-sm"><input v-model="profile.capabilities.vision" type="checkbox" />视觉</label>
                  <label class="flex items-center gap-2 text-sm"><input v-model="profile.capabilities.image" type="checkbox" :disabled="profile.provider !== 'openai'" />生图（仅 OpenAI）</label>
                  <label class="flex items-center gap-2 text-sm"><input v-model="profile.capabilities.web_search" type="checkbox" />联网搜索</label>
                  <label class="flex items-center gap-2 text-sm">
                    上下文字符上限
                    <input
                      v-model.number="profile.capabilities.context_limit"
                      class="input w-32"
                      type="number"
                      min="0"
                      max="10000000"
                      step="1"
                      placeholder="0"
                    />
                  </label>
                </div>
              </fieldset>
            </div>
          </article>
        </section>

        <section class="space-y-4">
          <div class="flex items-center justify-between">
            <div>
              <h2 class="text-lg font-semibold text-gray-900 dark:text-white">Prompt Skill</h2>
              <p class="text-sm text-gray-500 dark:text-gray-400">Skill 只能包含文本指令，不支持上传代码或任意工具定义。</p>
            </div>
            <button class="btn btn-secondary" type="button" @click="addSkill">新增 Skill</button>
          </div>

          <div v-if="policy.skills.length === 0" class="card p-8 text-center text-sm text-gray-500">尚未配置 Skill。</div>
          <article v-for="(skill, index) in policy.skills" :key="`${skill.id}-${index}`" class="card p-6">
            <div class="grid gap-4 md:grid-cols-2">
              <label class="block text-sm">名称<input v-model.trim="skill.name" class="input mt-1" placeholder="Research" /></label>
              <label class="block text-sm">Skill ID<input v-model.trim="skill.id" class="input mt-1 font-mono" placeholder="research" /></label>
            </div>
            <label class="mt-4 block text-sm">说明<input v-model.trim="skill.description" class="input mt-1" /></label>
            <label class="mt-4 block text-sm">Skill 指令<textarea v-model="skill.instructions" class="input mt-1 min-h-32 resize-y"></textarea></label>
            <div class="mt-4 flex items-center justify-between">
              <label class="flex items-center gap-2 text-sm"><input v-model="skill.enabled" type="checkbox" />启用</label>
              <button class="text-sm text-red-600 hover:text-red-700" type="button" @click="removeSkill(index)">删除</button>
            </div>
          </article>
        </section>
      </template>
    </div>
  </AppLayout>
</template>

<script setup lang="ts">
import { computed, onMounted, reactive, ref } from 'vue'
import AppLayout from '@/components/layout/AppLayout.vue'
import { adminAPI, type ChatPolicy, type ChatProfile, type ChatProvider, type ChatSkill } from '@/api/admin'
import type { AdminGroup } from '@/types'
import { useAppStore } from '@/stores/app'

const appStore = useAppStore()
const loading = ref(true)
const saving = ref(false)
const loaded = ref(false)
const loadError = ref('')
const groups = ref<AdminGroup[]>([])
const policy = reactive<ChatPolicy>({ enabled: false, profiles: [], skills: [] })

const enabledSkills = computed(() => policy.skills.filter((skill) => skill.enabled && skill.id.trim()))

function getErrorMessage(error: unknown, fallback: string): string {
  if (error instanceof Error && error.message) return error.message
  if (error && typeof error === 'object' && 'message' in error) {
    const message = (error as { message?: unknown }).message
    if (typeof message === 'string' && message.trim()) return message
  }
  return fallback
}

function groupsForProvider(provider: ChatProvider) {
  return groups.value.filter((group) => group.platform === provider && group.status === 'active')
}

function ensureCompatibleGroup(profile: ChatProfile) {
  if (!groupsForProvider(profile.provider).some((group) => group.id === profile.group_id)) profile.group_id = 0
  if (profile.provider !== 'openai') profile.capabilities.image = false
}

function createProfile(provider: ChatProvider = 'openai'): ChatProfile {
  return {
    id: '', name: '', provider, public_model: '', upstream_model: '', group_id: 0,
    system_prompt: '', skill_ids: [], enabled: true, default: policy.profiles.length === 0,
    capabilities: { vision: false, image: false, web_search: false }
  }
}

function createSkill(): ChatSkill {
  return { id: '', name: '', description: '', instructions: '', enabled: true }
}

function addProfile() { policy.profiles.push(createProfile()) }
function addSkill() { policy.skills.push(createSkill()) }
function setDefault(index: number) { policy.profiles.forEach((profile, i) => { profile.default = i === index }) }
function removeProfile(index: number) {
  const wasDefault = policy.profiles[index]?.default
  policy.profiles.splice(index, 1)
  if (wasDefault && policy.profiles.length > 0) setDefault(0)
}
function removeSkill(index: number) {
  const [removed] = policy.skills.splice(index, 1)
  if (removed) policy.profiles.forEach((profile) => { profile.skill_ids = profile.skill_ids.filter((id) => id !== removed.id) })
}

async function load() {
  loading.value = true
  loaded.value = false
  loadError.value = ''
  try {
    const [stored, activeGroups] = await Promise.all([adminAPI.chatPolicy.get(), adminAPI.groups.getAll()])
    Object.assign(policy, stored, { profiles: stored.profiles || [], skills: stored.skills || [] })
    groups.value = activeGroups.filter((group) => ['openai', 'anthropic', 'gemini'].includes(group.platform))
    loaded.value = true
  } catch (error) {
    loadError.value = getErrorMessage(error, '加载 Chat 策略失败')
    appStore.showError(loadError.value)
  } finally {
    loading.value = false
  }
}

async function save() {
  if (!loaded.value || loading.value || saving.value) return
  saving.value = true
  try {
    const updated = await adminAPI.chatPolicy.update(JSON.parse(JSON.stringify(policy)) as ChatPolicy)
    Object.assign(policy, updated, { profiles: updated.profiles || [], skills: updated.skills || [] })
    appStore.showSuccess('Chat 策略已保存')
  } catch (error) {
    appStore.showError(getErrorMessage(error, '保存 Chat 策略失败'))
  } finally {
    saving.value = false
  }
}

onMounted(load)
</script>
