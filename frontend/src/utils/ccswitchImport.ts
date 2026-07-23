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

export type CcSwitchClientType = 'claude' | 'gemini'
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

function withV1Endpoint(baseUrl: string): string {
  const normalizedBaseUrl = baseUrl.replace(/\/+$/, '')
  return normalizedBaseUrl.endsWith('/v1') ? normalizedBaseUrl : `${normalizedBaseUrl}/v1`
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
        endpoint: `${baseUrl}/antigravity`
      }
    case 'openai':
      return {
        app: 'codex',
        endpoint: baseUrl,
        model: OPENAI_CC_SWITCH_CODEX_MODEL
      }
    case 'gemini':
      return {
        app: 'gemini',
        endpoint: baseUrl
      }
    case 'grok':
      return {
        app: 'grokbuild',
        endpoint: withV1Endpoint(baseUrl),
        model: GROK_CC_SWITCH_MODEL
      }
    default:
      return {
        app: resolveCcSwitchAppType(clientType),
        endpoint: baseUrl
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
    ['usageScript', btoa(input.usageScript)],
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

export function openCcSwitchDeeplink(deeplink: string): void {
  if (typeof window === 'undefined') return

  if (typeof document === 'undefined' || !document.body) {
    window.location.assign(deeplink)
    return
  }

  const anchor = document.createElement('a')
  anchor.href = deeplink
  anchor.style.display = 'none'
  anchor.setAttribute('aria-hidden', 'true')
  document.body.appendChild(anchor)
  anchor.click()
  window.setTimeout(() => {
    anchor.remove()
  }, 0)
}
