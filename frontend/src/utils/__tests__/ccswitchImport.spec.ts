import { describe, expect, it } from 'vitest'
import {
  CC_SWITCH_DOWNLOAD_LINKS,
  GROK_CC_SWITCH_MODEL,
  OPENAI_CC_SWITCH_CODEX_MODEL,
  buildCcSwitchImportDeeplink,
  getCcSwitchProtocolFallbackDelayMs,
  isAppleLikePlatform
} from '@/utils/ccswitchImport'
import type { GroupPlatform } from '@/types'

function paramsFromDeeplink(deeplink: string): URLSearchParams {
  const query = deeplink.split('?')[1] || ''
  return new URLSearchParams(query)
}

describe('ccswitchImport utils', () => {
  it('defaults OpenAI CC Switch imports to the current Codex model', () => {
    expect(OPENAI_CC_SWITCH_CODEX_MODEL).toBe('gpt-5.6-sol')
  })

  it('keeps CC-Switch download links on official channels', () => {
    expect(CC_SWITCH_DOWNLOAD_LINKS.officialSite).toBe('https://ccswitch.io/')
    expect(CC_SWITCH_DOWNLOAD_LINKS.releases).toBe('https://github.com/farion1231/cc-switch/releases/latest')
    expect(CC_SWITCH_DOWNLOAD_LINKS.windows).toMatch(
      /^https:\/\/github\.com\/farion1231\/cc-switch\/releases\/download\/v[\d.]+\/CC-Switch-v[\d.]+-Windows\.msi$/
    )
    expect(CC_SWITCH_DOWNLOAD_LINKS.macos).toMatch(
      /^https:\/\/github\.com\/farion1231\/cc-switch\/releases\/download\/v[\d.]+\/CC-Switch-v[\d.]+-macOS\.dmg$/
    )
  })

  it('uses a longer protocol fallback delay for Apple browsers', () => {
    expect(isAppleLikePlatform({
      platform: 'MacIntel',
      userAgent: 'Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7)',
      maxTouchPoints: 0
    })).toBe(true)
    expect(getCcSwitchProtocolFallbackDelayMs({
      platform: 'MacIntel',
      userAgent: 'Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7)',
      maxTouchPoints: 0
    })).toBe(5000)
    expect(getCcSwitchProtocolFallbackDelayMs({
      platform: 'Win32',
      userAgent: 'Mozilla/5.0 (Windows NT 10.0; Win64; x64)',
      maxTouchPoints: 0
    })).toBe(1800)
  })

  it('defaults Grok Build imports to the current Grok model', () => {
    expect(GROK_CC_SWITCH_MODEL).toBe('grok-4.5')
  })

  const baseInput = {
    baseUrl: 'https://api.example.com',
    providerName: 'ISACAPI',
    apiKey: 'sk-test',
    usageScript: 'return true'
  }

  it('adds the Codex model parameter for OpenAI imports', () => {
    const params = paramsFromDeeplink(
      buildCcSwitchImportDeeplink({
        ...baseInput,
        platform: 'openai',
        clientType: 'claude'
      })
    )

    expect(params.get('resource')).toBe('provider')
    expect(params.get('app')).toBe('codex')
    expect(params.get('endpoint')).toBe(baseInput.baseUrl)
    expect(params.get('model')).toBe(OPENAI_CC_SWITCH_CODEX_MODEL)
    expect(params.get('enabled')).toBe('true')
    expect(atob(params.get('usageScript') || '')).toBe(baseInput.usageScript)
  })

  it.each([
    'https://api.example.com',
    'https://api.example.com/',
    'https://api.example.com/v1',
    'https://api.example.com/v1/'
  ])('imports Grok Build with one /v1 suffix for base URL %s', (baseUrl) => {
    const params = paramsFromDeeplink(
      buildCcSwitchImportDeeplink({
        ...baseInput,
        baseUrl,
        platform: 'grok',
        clientType: 'claude'
      })
    )

    expect(params.get('app')).toBe('grokbuild')
    expect(params.get('endpoint')).toBe('https://api.example.com/v1')
    expect(params.get('model')).toBe(GROK_CC_SWITCH_MODEL)
  })

  it.each([
    { platform: 'anthropic' as GroupPlatform, clientType: 'claude' as const, app: 'claude' },
    { platform: 'gemini' as GroupPlatform, clientType: 'gemini' as const, app: 'gemini' }
  ])('does not add a model parameter for $platform imports', ({ platform, clientType, app }) => {
    const params = paramsFromDeeplink(
      buildCcSwitchImportDeeplink({
        ...baseInput,
        platform,
        clientType
      })
    )

    expect(params.get('app')).toBe(app)
    expect(params.get('endpoint')).toBe(baseInput.baseUrl)
    expect(params.get('enabled')).toBe('true')
    expect(params.has('model')).toBe(false)
  })

  it.each([
    { clientType: 'claude' as const, app: 'claude' },
    { clientType: 'gemini' as const, app: 'gemini' }
  ])('keeps Antigravity $app imports on the selected client endpoint without a model parameter', ({ clientType, app }) => {
    const params = paramsFromDeeplink(
      buildCcSwitchImportDeeplink({
        ...baseInput,
        platform: 'antigravity',
        clientType
      })
    )

    expect(params.get('app')).toBe(app)
    expect(params.get('endpoint')).toBe(`${baseInput.baseUrl}/antigravity`)
    expect(params.get('enabled')).toBe('true')
    expect(params.has('model')).toBe(false)
  })
})
