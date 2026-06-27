/**
 * Public (no-auth) service status API.
 *
 * Returns availability aggregated by provider/model. Internal channel and group
 * names are never included. Gated server-side by the `public_status_enabled`
 * opt-in flag — when disabled the endpoint returns `enabled: false`.
 */

import { apiClient } from './client'

export type PublicStatusValue = 'operational' | 'degraded' | 'down'

export interface PublicStatusModel {
  model: string
  status: PublicStatusValue
  availability_7d: number | null
  /** Groups (分组) this model belongs to. May be empty. */
  groups: string[] | null
}

export interface PublicStatusProvider {
  provider: string
  status: PublicStatusValue
  availability_7d: number | null
  models: PublicStatusModel[]
}

export interface PublicStatusResponse {
  enabled: boolean
  overall_status: PublicStatusValue
  providers: PublicStatusProvider[]
}

/**
 * Fetch the public service status. No authentication required.
 */
export async function getPublicStatus(options?: {
  signal?: AbortSignal
}): Promise<PublicStatusResponse> {
  const { data } = await apiClient.get<PublicStatusResponse>('/status/public', {
    signal: options?.signal,
  })
  return data
}

export const publicStatusAPI = { get: getPublicStatus }
