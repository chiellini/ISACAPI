import type { GroupPlatform } from '@/types'

export const OPENAI_CC_SWITCH_CODEX_MODEL = 'gpt-5.5'

const CC_SWITCH_LATEST_VERSION = 'v3.16.5'
const CC_SWITCH_RELEASE_BASE = `https://github.com/farion1231/cc-switch/releases/download/${CC_SWITCH_LATEST_VERSION}`

export const CC_SWITCH_DOWNLOAD_LINKS = {
  officialSite: 'https://ccswitch.io/',
  releases: 'https://github.com/farion1231/cc-switch/releases/latest',
  windows: `${CC_SWITCH_RELEASE_BASE}/CC-Switch-${CC_SWITCH_LATEST_VERSION}-Windows.msi`,
  macos: `${CC_SWITCH_RELEASE_BASE}/CC-Switch-${CC_SWITCH_LATEST_VERSION}-macOS.dmg`
} as const

export type CcSwitchClientType = 'claude' | 'claude-desktop' | 'gemini'
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

export function resolveCcSwitchImportConfig(
  platform: GroupPlatform | undefined | null,
  clientType: CcSwitchClientType,
  baseUrl: string
): CcSwitchImportConfig {
  switch (platform || 'anthropic') {
    case 'antigravity':
      return {
        app: clientType === 'gemini' ? 'gemini' : clientType,
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
    default:
      return {
        app: clientType === 'claude-desktop' ? 'claude-desktop' : 'claude',
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
