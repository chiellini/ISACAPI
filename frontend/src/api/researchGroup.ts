import { apiClient } from './client'
import type {
  PaginatedResponse,
  ResearchGroupContext,
  ResearchGroupMember,
  ResearchGroupMemberStatus,
  ResearchGroupStatus,
} from '@/types'

export interface CreateResearchGroupRequest {
  name: string
}

export interface UpdateResearchGroupRequest {
  name?: string
  status?: ResearchGroupStatus
}

export interface InviteResearchGroupMemberRequest {
  email: string
  monthly_limit_usd: number
}

export interface UpdateResearchGroupMemberRequest {
  monthly_limit_usd?: number
  status?: Exclude<ResearchGroupMemberStatus, 'pending'>
}

export interface ResearchGroupUsageItem {
  id: number
  user_id: number
  member_id: number
  request_id: string
  model: string
  total_cost: number
  created_at: string
  email: string
  username: string
}

export interface ResearchGroupUsageParams {
  page?: number
  page_size?: number
  member_id?: number
  start_time?: string
  end_time?: string
}

const BASE_PATH = '/user/research-group'

async function getContext(): Promise<ResearchGroupContext | null> {
  const { data } = await apiClient.get<ResearchGroupContext | null>(BASE_PATH)
  return data
}

async function create(payload: CreateResearchGroupRequest): Promise<ResearchGroupContext> {
  const { data } = await apiClient.post<ResearchGroupContext>(BASE_PATH, payload)
  return data
}

async function update(payload: UpdateResearchGroupRequest): Promise<ResearchGroupContext> {
  const { data } = await apiClient.patch<ResearchGroupContext>(BASE_PATH, payload)
  return data
}

async function dissolve(): Promise<void> {
  await apiClient.delete(BASE_PATH)
}

async function inviteMember(
  payload: InviteResearchGroupMemberRequest
): Promise<ResearchGroupMember> {
  const { data } = await apiClient.post<ResearchGroupMember>(`${BASE_PATH}/members`, payload)
  return data
}

async function updateMember(
  memberId: number,
  payload: UpdateResearchGroupMemberRequest
): Promise<ResearchGroupMember> {
  const { data } = await apiClient.patch<ResearchGroupMember>(
    `${BASE_PATH}/members/${memberId}`,
    payload
  )
  return data
}

async function resetMemberUsage(memberId: number): Promise<ResearchGroupMember> {
  const { data } = await apiClient.post<ResearchGroupMember>(
    `${BASE_PATH}/members/${memberId}/reset`
  )
  return data
}

async function removeMember(memberId: number): Promise<void> {
  await apiClient.delete(`${BASE_PATH}/members/${memberId}`)
}

async function listInvitations(): Promise<ResearchGroupMember[]> {
  const { data } = await apiClient.get<ResearchGroupMember[]>(`${BASE_PATH}/invitations`)
  return data
}

async function acceptInvitation(invitationId: number): Promise<ResearchGroupContext> {
  const { data } = await apiClient.post<ResearchGroupContext>(
    `${BASE_PATH}/invitations/${invitationId}/accept`
  )
  return data
}

async function rejectInvitation(invitationId: number): Promise<void> {
  await apiClient.post(`${BASE_PATH}/invitations/${invitationId}/reject`)
}

async function leave(): Promise<void> {
  await apiClient.post(`${BASE_PATH}/leave`)
}

async function getUsage(
  params: ResearchGroupUsageParams = {}
): Promise<PaginatedResponse<ResearchGroupUsageItem>> {
  const { data } = await apiClient.get<PaginatedResponse<ResearchGroupUsageItem>>(
    `${BASE_PATH}/usage`,
    { params }
  )
  return data
}

export const researchGroupAPI = {
  getContext,
  create,
  update,
  dissolve,
  inviteMember,
  updateMember,
  resetMemberUsage,
  removeMember,
  listInvitations,
  acceptInvitation,
  rejectInvitation,
  leave,
  getUsage,
}

export default researchGroupAPI
