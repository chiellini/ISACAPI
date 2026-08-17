import { apiClient } from '../client'

export type ChatProvider = 'openai' | 'anthropic' | 'gemini'

export interface ChatCapabilities {
  vision: boolean
  image: boolean
  web_search: boolean
  context_limit?: number
}

export interface ChatSkill {
  id: string
  name: string
  description?: string
  instructions: string
  enabled: boolean
}

export interface ChatProfile {
  id: string
  name: string
  provider: ChatProvider
  public_model: string
  upstream_model: string
  group_id: number
  system_prompt: string
  skill_ids: string[]
  enabled: boolean
  default: boolean
  capabilities: ChatCapabilities
}

export interface ChatPolicy {
  enabled: boolean
  profiles: ChatProfile[]
  skills: ChatSkill[]
}

export async function get(): Promise<ChatPolicy> {
  const { data } = await apiClient.get<ChatPolicy>('/admin/settings/chat-policy')
  return data
}

export async function update(policy: ChatPolicy): Promise<ChatPolicy> {
  const { data } = await apiClient.put<ChatPolicy>('/admin/settings/chat-policy', policy)
  return data
}

export default { get, update }
