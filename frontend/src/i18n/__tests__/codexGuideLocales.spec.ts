import { describe, expect, it } from 'vitest'

import ar from '../locales/ar'
import en from '../locales/en'
import ja from '../locales/ja'
import zh from '../locales/zh'
import zhHant from '../locales/zh-Hant'

describe('Codex guide locale copy', () => {
  const locales = { zh, 'zh-Hant': zhHant, en, ja, ar }

  it.each(Object.entries(locales))('defines the complete visible guide for %s', (_locale, messages) => {
    expect(messages.nav.codexGuide).toBeTruthy()
    expect(messages.codexGuide.title).toBeTruthy()
    expect(messages.codexGuide.quickStart.items).toHaveLength(4)
    expect(messages.codexGuide.install.idePoints.length).toBeGreaterThanOrEqual(3)
    expect(messages.codexGuide.configure.steps).toHaveLength(3)
    expect(messages.codexGuide.use.commands).toHaveLength(4)
    expect(messages.codexGuide.bestPractices.items.length).toBeGreaterThanOrEqual(4)
    expect(messages.codexGuide.troubleshooting.items.length).toBeGreaterThanOrEqual(5)
  })

  it('keeps the new navigation label localized', () => {
    for (const messages of [zh, zhHant, ja, ar]) {
      expect(messages.nav.codexGuide).not.toBe(en.nav.codexGuide)
    }
  })
})
