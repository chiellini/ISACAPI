import { describe, expect, it } from 'vitest'
import {
  CC_SWITCH_DOWNLOAD_LINKS,
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
    expect(OPENAI_CC_SWITCH_CODEX_MODEL).toBe('gpt-5.5')
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
    { platform: 'anthropic' as GroupPlatform, clientType: 'claude' as const, app: 'claude' },
    { platform: 'anthropic' as GroupPlatform, clientType: 'claude-desktop' as const, app: 'claude-desktop' },
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

  it('routes Claude Desktop imports to a distinct app so CC-Switch does not fall back to Claude Code', () => {
    const claudeCode = paramsFromDeeplink(
      buildCcSwitchImportDeeplink({ ...baseInput, platform: 'anthropic', clientType: 'claude' })
    )
    const claudeDesktop = paramsFromDeeplink(
      buildCcSwitchImportDeeplink({ ...baseInput, platform: 'anthropic', clientType: 'claude-desktop' })
    )

    expect(claudeCode.get('app')).toBe('claude')
    expect(claudeDesktop.get('app')).toBe('claude-desktop')
    expect(claudeDesktop.get('app')).not.toBe(claudeCode.get('app'))
  })

  it.each([
    { clientType: 'claude' as const, app: 'claude' },
    { clientType: 'claude-desktop' as const, app: 'claude-desktop' },
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
