/**
 * API Client for ISACAPI Backend
 * Central export point for all API modules
 */

// Re-export the HTTP client
export { apiClient } from './client'

// Auth API
export { authAPI, isTotp2FARequired, type LoginResponse } from './auth'

// User APIs
export { keysAPI } from './keys'
export { usageAPI } from './usage'
export { userAPI } from './user'
export { redeemAPI, type RedeemHistoryItem } from './redeem'
export { paymentAPI } from './payment'
export { userGroupsAPI } from './groups'
export { userChannelsAPI } from './channels'
export * as batchImageAPI from './batchImage'
export { totpAPI } from './totp'
export { default as announcementsAPI } from './announcements'
export { channelMonitorUserAPI } from './channelMonitor'
export { providerAPI } from './provider'
export type {
  ProviderAccountCreateRequest,
  ProviderAccountUpdateRequest,
  ProviderGroupSummary,
  ProviderUsageResponse,
  ProviderUsageTotals,
} from './provider'
export {
  researchGroupAPI,
  type ResearchGroupUsageItem,
  type ResearchGroupUsageParams,
} from './researchGroup'

// Admin APIs
export { adminAPI } from './admin'

// Default export
export { default } from './client'
