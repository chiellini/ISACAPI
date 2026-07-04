import { describe, expect, it } from 'vitest'

import ar from '../locales/ar'
import en from '../locales/en'
import ja from '../locales/ja'
import zh from '../locales/zh'

describe('CC-Switch and one-click setup locale copy', () => {
  const locales = { zh, en, ja, ar }

  it.each(Object.entries(locales))('defines new setup copy for %s', (_locale, messages) => {
    expect(messages.nav.ccSwitchDownload).toBeTruthy()
    expect(messages.keys.useKeyModal.oneClick.hint).toContain('VS Code')
    expect(messages.keys.useKeyModal.oneClick.runHint).toContain('VS Code')
    expect(messages.keys.ccsFallback.downloadTitle).toBeTruthy()
    expect(messages.keys.ccsFallback.downloadDescription).toBeTruthy()
    expect(messages.keys.ccsFallback.windowsDownload).toBe('Windows MSI')
    expect(messages.keys.ccsFallback.macosDownload).toBe('macOS DMG')
    expect(messages.keys.ccsFallback.releasePage).toBeTruthy()
  })

  it('does not let Japanese or Arabic fall back to English for the new visible strings', () => {
    for (const messages of [ja, ar]) {
      expect(messages.nav.ccSwitchDownload).not.toBe(en.nav.ccSwitchDownload)
      expect(messages.keys.useKeyModal.oneClick.hint).not.toBe(en.keys.useKeyModal.oneClick.hint)
      expect(messages.keys.useKeyModal.oneClick.runHint).not.toBe(en.keys.useKeyModal.oneClick.runHint)
      expect(messages.keys.ccsFallback.downloadTitle).not.toBe(en.keys.ccsFallback.downloadTitle)
      expect(messages.keys.ccsFallback.downloadDescription).not.toBe(en.keys.ccsFallback.downloadDescription)
      expect(messages.keys.ccsFallback.releasePage).not.toBe(en.keys.ccsFallback.releasePage)
    }
  })
})
