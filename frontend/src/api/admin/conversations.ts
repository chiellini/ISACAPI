/**
 * Admin Conversation Archive API (admin-only view + delete).
 */

import { apiClient } from '../client'
import type { PaginatedResponse } from '@/types'

export interface ConversationSessionView {
  id: string
  user_id: number
  api_key_id?: number
  group_id: number
  context_domain: string
  protocol: string
  title: string
  started_at: string
  last_active_at: string
  request_count: number
  total_input_tokens: number
  total_output_tokens: number
  status: string
}

export interface ConversationBranchView {
  id: string
  parent_branch_id?: string
  branch_reason: string
  event_count: number
  status: string
  created_at: string
}

export interface ConversationEventView {
  id: number
  branch_id: string
  sequence: number
  role: string
  kind: string
  content: string
  model: string
  partial: boolean
  encrypted: boolean
  decrypted: boolean
  created_at: string
}

export interface ConversationSessionDetail {
  session: ConversationSessionView
  branches: ConversationBranchView[]
  events: ConversationEventView[]
}

export interface ConversationListParams {
  page?: number
  page_size?: number
  user_id?: number
  group_id?: number
  status?: string
  context_domain?: string
  protocol?: string
  keyword?: string
  from?: string
  to?: string
}

async function list(
  params: ConversationListParams,
  options?: { signal?: AbortSignal }
): Promise<PaginatedResponse<ConversationSessionView>> {
  const { data } = await apiClient.get<PaginatedResponse<ConversationSessionView>>('/admin/conversations', {
    params,
    signal: options?.signal
  })
  return data
}

async function get(id: string): Promise<ConversationSessionDetail> {
  const { data } = await apiClient.get<ConversationSessionDetail>(`/admin/conversations/${id}`)
  return data
}

async function remove(id: string): Promise<void> {
  await apiClient.delete(`/admin/conversations/${id}`)
}

async function exportSession(id: string): Promise<Blob> {
  const response = await apiClient.get(`/admin/conversations/${id}/export`, { responseType: 'blob' })
  return response.data
}

async function exportAll(params: ConversationListParams): Promise<Blob> {
  const response = await apiClient.get('/admin/conversation-exports', { params, responseType: 'blob' })
  return response.data
}

export const conversationsAPI = {
  list,
  get,
  remove,
  exportSession,
  exportAll
}

export default conversationsAPI
