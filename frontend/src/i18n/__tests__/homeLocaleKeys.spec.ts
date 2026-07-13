import { describe, expect, it } from 'vitest'

import homeViewSource from '../../views/HomeView.vue?raw'
import ar from '../locales/ar'
import en from '../locales/en'
import ja from '../locales/ja'
import zh from '../locales/zh'
import zhHant from '../locales/zh-Hant'

type Messages = Record<string, unknown>

const homeKeys = [...new Set(homeViewSource.match(/home\.[A-Za-z0-9_.]+/g) ?? [])].sort()
const providerBlock = homeViewSource.match(/const providerLogos = \[([\s\S]*?)\]\s+as const/)?.[1] ?? ''
const providerModels = [...providerBlock.matchAll(/model: '([^']+)'/g)].map((match) => match[1])
const explicitlyLocalizedHomeKeys = [
  'home.register',
  'home.nav.home',
  'home.nav.dashboard',
  'home.nav.tagline',
  'home.nav.pricing',
  'home.nav.integrations',
  'home.nav.openMenu',
  'home.nav.closeMenu',
  'home.heroTitle',
  'home.heroAccent',
  'home.supportedAccess',
  'home.apiDemo.request',
  'home.apiDemo.response',
  'home.apiDemo.routed',
  'home.heroStats.recharge',
  'home.heroStats.groupRate',
  'home.heroStats.models',
  'home.heroStats.deployments',
  'home.workflow.eyebrow',
  'home.workflow.title',
  'home.workflow.description',
  'home.workflow.steps.account.title',
  'home.workflow.steps.account.description',
  'home.workflow.steps.configure.title',
  'home.workflow.steps.configure.description',
  'home.workflow.steps.connect.title',
  'home.workflow.steps.connect.description',
  'home.notice.open'
] as const

function resolveMessage(messages: Messages, key: string): unknown {
  return key.split('.').reduce<unknown>((value, segment) => {
    if (!value || typeof value !== 'object') {
      return undefined
    }
    return (value as Messages)[segment]
  }, messages)
}

describe('HomeView locale coverage', () => {
  const locales: Record<string, Messages> = { zh, 'zh-Hant': zhHant, en, ja, ar }

  it('finds the translation keys referenced by HomeView', () => {
    expect(homeKeys.length).toBeGreaterThan(0)
  })

  it.each(Object.entries(locales))('defines every visible HomeView message for %s', (_locale, messages) => {
    const missing = homeKeys.filter((key) => {
      const message = resolveMessage(messages, key)
      return typeof message !== 'string' || message.trim().length === 0
    })
    expect(missing).toEqual([])
  })

  it('keeps the new homepage copy localized instead of inheriting the base locale', () => {
    for (const key of [
      'home.ccSwitch.title',
      'home.ccSwitch.description',
      'home.integrations.title',
      ...explicitlyLocalizedHomeKeys
    ]) {
      expect(resolveMessage(zhHant, key)).not.toBe(resolveMessage(zh, key))
      expect(resolveMessage(ja, key)).not.toBe(resolveMessage(en, key))
      expect(resolveMessage(ar, key)).not.toBe(resolveMessage(en, key))
    }
  })
})

describe('HomeView product focus', () => {
  it('shows only the six requested model families', () => {
    expect(providerModels).toEqual([
      'gpt-4.1',
      'claude-3-5-sonnet',
      'gemini-2.5-pro',
      'grok-4',
      'deepseek-chat',
      'glm-4-plus'
    ])
  })

  it('makes CC-Switch and developer-tool setup the primary journey', () => {
    expect(homeViewSource).toContain('CC_SWITCH_DOWNLOAD_LINKS.officialSite')
    expect(homeViewSource).toContain('to="/keys"')
    expect(homeViewSource).toContain('to="/cc-switch"')
    expect(homeViewSource).toContain("name: 'Codex'")
    expect(homeViewSource).toContain("name: 'Claude Code'")
    expect(homeViewSource).toContain("name: 'Gemini CLI'")
    expect(homeViewSource).toContain("name: 'VS Code / Cursor / IDE'")
  })

  it('restores the versioned homepage notice without a permanent dismissal', () => {
    expect(homeViewSource).toContain("const NOTICE_VERSION = '2026-07-pricing-v2'")
    expect(homeViewSource).toContain('sessionStorage.getItem(NOTICE_SESSION_KEY)')
    expect(homeViewSource).toContain('showNotice.value = !closedForSession && !closedToday')
    expect(homeViewSource).not.toContain('NOTICE_PERMANENT_KEY')
  })
})
