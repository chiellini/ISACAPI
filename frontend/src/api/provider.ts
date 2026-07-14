import { apiClient } from './client'
import type { Account, AccountPlatform, AccountType, Group, PaginatedResponse } from '@/types'

export type ProviderGroupSummary = Pick<Group, 'id' | 'name' | 'description' | 'platform' | 'status'>

export interface ProviderAccountListParams {
  page?: number
  page_size?: number
  platform?: AccountPlatform | ''
  type?: AccountType | ''
  status?: Account['status'] | ''
  search?: string
}

export interface ProviderAccountCreateRequest {
  name: string
  notes?: string | null
  platform: AccountPlatform
  type: AccountType
  credentials: Record<string, unknown>
  concurrency?: number
  group_ids?: number[]
}

export interface ProviderAccountUpdateRequest {
  name?: string
  notes?: string | null
  credentials?: Record<string, unknown>
  concurrency?: number
  status?: 'active' | 'inactive'
  group_ids?: number[]
}

export interface ProviderUsageTotals {
  total_requests: number
  total_input_tokens: number
  total_output_tokens: number
  total_cache_creation_tokens: number
  total_cache_read_tokens: number
  total_tokens: number
}

export interface ProviderAccountUsage extends ProviderUsageTotals {
  account_id: number
  account_name: string
  platform: AccountPlatform | string
}

export interface ProviderUsageResponse {
  provider_id: number
  start_time: string
  end_time: string
  totals: ProviderUsageTotals
  accounts: ProviderAccountUsage[]
}

export interface ProviderUsageParams {
  start_time?: string
  end_time?: string
}

const toCount = (value: unknown): number => {
  const parsed = typeof value === 'number' ? value : Number(value)
  return Number.isFinite(parsed) && parsed >= 0 ? Math.trunc(parsed) : 0
}

const pickCount = (source: Record<string, unknown>, current: string, legacy: string): number =>
  toCount(source[current] ?? source[legacy])

const normalizeTotals = (value: unknown): ProviderUsageTotals => {
  const source = value && typeof value === 'object' ? value as Record<string, unknown> : {}
  const total_input_tokens = pickCount(source, 'total_input_tokens', 'input_tokens')
  const total_output_tokens = pickCount(source, 'total_output_tokens', 'output_tokens')
  const total_cache_creation_tokens = pickCount(source, 'total_cache_creation_tokens', 'cache_creation_tokens')
  const total_cache_read_tokens = pickCount(source, 'total_cache_read_tokens', 'cache_read_tokens')
  const fallbackTotal = total_input_tokens + total_output_tokens + total_cache_creation_tokens + total_cache_read_tokens

  return {
    total_requests: pickCount(source, 'total_requests', 'requests'),
    total_input_tokens,
    total_output_tokens,
    total_cache_creation_tokens,
    total_cache_read_tokens,
    total_tokens: toCount(source.total_tokens ?? source.tokens ?? fallbackTotal),
  }
}

export const normalizeProviderUsage = (value: unknown): ProviderUsageResponse => {
  const source = value && typeof value === 'object' ? value as Record<string, unknown> : {}
  const rows = Array.isArray(source.accounts)
    ? source.accounts
    : Array.isArray(source.per_account)
      ? source.per_account
      : []

  return {
    provider_id: toCount(source.provider_id),
    start_time: typeof source.start_time === 'string' ? source.start_time : '',
    end_time: typeof source.end_time === 'string' ? source.end_time : '',
    totals: normalizeTotals(source.totals ?? source),
    accounts: rows.map((row) => {
      const record = row && typeof row === 'object' ? row as Record<string, unknown> : {}
      return {
        account_id: toCount(record.account_id),
        account_name: typeof record.account_name === 'string' ? record.account_name : '',
        platform: typeof record.platform === 'string' ? record.platform : '',
        ...normalizeTotals(record),
      }
    }),
  }
}

const sanitizeCreate = (request: ProviderAccountCreateRequest): ProviderAccountCreateRequest => ({
  name: request.name.trim(),
  notes: request.notes?.trim() || null,
  platform: request.platform,
  type: request.type,
  credentials: request.credentials,
  concurrency: Math.max(1, Math.trunc(request.concurrency ?? 1)),
  group_ids: [...new Set(request.group_ids ?? [])],
})

const sanitizeUpdate = (request: ProviderAccountUpdateRequest): ProviderAccountUpdateRequest => {
  const payload: ProviderAccountUpdateRequest = {}
  if (request.name !== undefined) payload.name = request.name.trim()
  if (request.notes !== undefined) payload.notes = request.notes?.trim() || null
  if (request.credentials !== undefined) payload.credentials = request.credentials
  if (request.concurrency !== undefined) payload.concurrency = Math.max(1, Math.trunc(request.concurrency))
  if (request.status !== undefined) payload.status = request.status
  if (request.group_ids !== undefined) payload.group_ids = [...new Set(request.group_ids)]
  return payload
}

export async function listAccounts(params: ProviderAccountListParams = {}): Promise<PaginatedResponse<Account>> {
  const page = params.page ?? 1
  const pageSize = params.page_size ?? 20
  const { data } = await apiClient.get<Partial<PaginatedResponse<Account>>>('/provider/accounts', {
    params: { ...params, page, page_size: pageSize }
  })
  const total = toCount(data.total)
  return {
    items: Array.isArray(data.items) ? data.items : [],
    total,
    page: toCount(data.page) || page,
    page_size: toCount(data.page_size) || pageSize,
    pages: toCount(data.pages) || Math.ceil(total / pageSize),
  }
}

export async function getAccount(id: number): Promise<Account> {
  const { data } = await apiClient.get<Account>(`/provider/accounts/${id}`)
  return data
}

export async function createAccount(request: ProviderAccountCreateRequest): Promise<Account> {
  const { data } = await apiClient.post<Account>('/provider/accounts', sanitizeCreate(request))
  return data
}

export async function updateAccount(id: number, request: ProviderAccountUpdateRequest): Promise<Account> {
  const { data } = await apiClient.put<Account>(`/provider/accounts/${id}`, sanitizeUpdate(request))
  return data
}

export async function deleteAccount(id: number): Promise<void> {
  await apiClient.delete(`/provider/accounts/${id}`)
}

export async function listGroups(): Promise<ProviderGroupSummary[]> {
  const { data } = await apiClient.get<ProviderGroupSummary[] | { items: ProviderGroupSummary[] }>('/provider/groups')
  return Array.isArray(data) ? data : data.items ?? []
}

export async function getUsage(params: ProviderUsageParams = {}): Promise<ProviderUsageResponse> {
  const { data } = await apiClient.get<unknown>('/provider/usage', { params })
  return normalizeProviderUsage(data)
}

export async function getAdminUsage(providerId: number, params: ProviderUsageParams = {}): Promise<ProviderUsageResponse> {
  const { data } = await apiClient.get<unknown>(`/admin/providers/${providerId}/usage`, { params })
  return normalizeProviderUsage(data)
}

export const providerAPI = {
  listAccounts,
  getAccount,
  createAccount,
  updateAccount,
  deleteAccount,
  listGroups,
  getUsage,
  getAdminUsage,
}

export default providerAPI
