import type { GroupPlatform } from '@/types'
import { OPENAI_CODEX_DEFAULT_MODEL } from '@/constants/codex'

export const OPENAI_CC_SWITCH_CODEX_MODEL = OPENAI_CODEX_DEFAULT_MODEL
export const GROK_CC_SWITCH_MODEL = 'grok-4.5'

const CC_SWITCH_LATEST_VERSION = 'v3.16.5'
const CC_SWITCH_RELEASE_BASE = `https://github.com/farion1231/cc-switch/releases/download/${CC_SWITCH_LATEST_VERSION}`

export const CC_SWITCH_DOWNLOAD_LINKS = {
  officialSite: 'https://ccswitch.io/',
  releases: 'https://github.com/farion1231/cc-switch/releases/latest',
  windows: `${CC_SWITCH_RELEASE_BASE}/CC-Switch-${CC_SWITCH_LATEST_VERSION}-Windows.msi`,
  macos: `${CC_SWITCH_RELEASE_BASE}/CC-Switch-${CC_SWITCH_LATEST_VERSION}-macOS.dmg`
} as const

/** CC Switch provider targets accepted by its v1 deeplink importer. */
export type CcSwitchClientType =
  | 'claude'
  | 'codex'
  | 'gemini'
  | 'grokbuild'
  | 'openclaw'
  | 'hermes'
  | 'opencode'
export type CcSwitchNavigatorSnapshot = Pick<Navigator, 'platform' | 'userAgent' | 'maxTouchPoints'>

export interface CcSwitchImportConfig {
  app: string
  endpoint: string
  model?: string
}

export interface CcSwitchImportDeeplinkInput {
  baseUrl: string
  platform?: GroupPlatform | null
  clientType: CcSwitchClientType
  providerName: string
  apiKey: string
  usageScript: string
}

function resolveCcSwitchAppType(clientType: CcSwitchClientType): string {
  return clientType
}

function encodeBase64Utf8(value: string): string {
  const bytes = new TextEncoder().encode(value)
  let binary = ''
  for (const byte of bytes) {
    binary += String.fromCharCode(byte)
  }
  return btoa(binary)
}

function normalizeBaseUrl(baseUrl: string): string {
  return baseUrl.replace(/\/+$/, '')
}

function withV1Endpoint(baseUrl: string): string {
  const normalizedBaseUrl = normalizeBaseUrl(baseUrl)
  return normalizedBaseUrl.endsWith('/v1') ? normalizedBaseUrl : `${normalizedBaseUrl}/v1`
}

function isOpenAICompatibleTarget(clientType: CcSwitchClientType): boolean {
  return clientType === 'codex'
    || clientType === 'grokbuild'
    || clientType === 'openclaw'
    || clientType === 'hermes'
    || clientType === 'opencode'
}

function resolveTargetEndpoint(baseUrl: string, clientType: CcSwitchClientType): string {
  const normalizedBaseUrl = normalizeBaseUrl(baseUrl)
  return isOpenAICompatibleTarget(clientType) ? withV1Endpoint(normalizedBaseUrl) : normalizedBaseUrl
}

function resolveTargetModel(
  platform: GroupPlatform | undefined | null,
  clientType: CcSwitchClientType
): string | undefined {
  if (platform === 'openai' && clientType === 'codex') return OPENAI_CC_SWITCH_CODEX_MODEL
  if (platform === 'grok' && clientType === 'grokbuild') return GROK_CC_SWITCH_MODEL
  return undefined
}

export function resolveCcSwitchImportConfig(
  platform: GroupPlatform | undefined | null,
  clientType: CcSwitchClientType,
  baseUrl: string
): CcSwitchImportConfig {
  switch (platform || 'anthropic') {
    case 'antigravity':
      return {
        app: resolveCcSwitchAppType(clientType),
        endpoint: `${normalizeBaseUrl(baseUrl)}/antigravity`
      }
    case 'openai':
      return {
        app: resolveCcSwitchAppType(clientType),
        endpoint: resolveTargetEndpoint(baseUrl, clientType),
        ...(resolveTargetModel(platform, clientType)
          ? { model: resolveTargetModel(platform, clientType) }
          : {})
      }
    case 'gemini':
      return {
        app: resolveCcSwitchAppType(clientType),
        endpoint: resolveTargetEndpoint(baseUrl, clientType)
      }
    case 'grok':
      return {
        app: resolveCcSwitchAppType(clientType),
        endpoint: resolveTargetEndpoint(baseUrl, clientType),
        ...(resolveTargetModel(platform, clientType)
          ? { model: resolveTargetModel(platform, clientType) }
          : {})
      }
    default:
      return {
        app: resolveCcSwitchAppType(clientType),
        endpoint: resolveTargetEndpoint(baseUrl, clientType)
      }
  }
}

export function buildCcSwitchImportDeeplink(input: CcSwitchImportDeeplinkInput): string {
  const config = resolveCcSwitchImportConfig(input.platform, input.clientType, input.baseUrl)
  const entries: [string, string][] = [
    ['resource', 'provider'],
    ['app', config.app],
    ['name', input.providerName],
    ['homepage', input.baseUrl],
    ['endpoint', config.endpoint],
    ['apiKey', input.apiKey],
    ['enabled', 'true'],
    ['configFormat', 'json'],
    ['usageEnabled', 'true'],
    ['usageScript', encodeBase64Utf8(input.usageScript)],
    ['usageAutoInterval', '30']
  ]

  if (config.model) {
    entries.splice(2, 0, ['model', config.model])
  }

  return `ccswitch://v1/import?${new URLSearchParams(entries).toString()}`
}

function getCurrentNavigator(): CcSwitchNavigatorSnapshot | undefined {
  return typeof navigator === 'undefined' ? undefined : navigator
}

export function isAppleLikePlatform(nav: CcSwitchNavigatorSnapshot | undefined = getCurrentNavigator()): boolean {
  if (!nav) return false
  const platform = nav.platform || ''
  const userAgent = nav.userAgent || ''
  return /Mac|iPhone|iPad|iPod/i.test(platform) || (/Macintosh/i.test(userAgent) && nav.maxTouchPoints > 1)
}

export function getCcSwitchProtocolFallbackDelayMs(nav: CcSwitchNavigatorSnapshot | undefined = getCurrentNavigator()): number {
  return isAppleLikePlatform(nav) ? 5000 : 1800
}

export function sanitizeCcSwitchDeeplink(value: string): string {
  const deeplink = value.trim()
  if (!deeplink) return ''
  try {
    const parsed = new URL(deeplink)
    return parsed.protocol === 'ccswitch:'
      && parsed.hostname === 'v1'
      && parsed.pathname === '/import'
      && !parsed.username
      && !parsed.password
      && !parsed.port
      && !parsed.hash
      ? deeplink
      : ''
  } catch {
    return ''
  }
}

export function openCcSwitchDeeplink(deeplink: string): void {
  if (typeof window === 'undefined') return

  const safeDeeplink = sanitizeCcSwitchDeeplink(deeplink)
  if (!safeDeeplink) return

  if (typeof document === 'undefined' || !document.body) {
    window.location.assign(safeDeeplink)
    return
  }

  const anchor = document.createElement('a')
  anchor.href = safeDeeplink
  anchor.style.display = 'none'
  anchor.setAttribute('aria-hidden', 'true')
  document.body.appendChild(anchor)
  anchor.click()
  window.setTimeout(() => {
    anchor.remove()
  }, 0)
}
