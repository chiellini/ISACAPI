import { describe, expect, it } from 'vitest'
import en from '@/i18n/locales/en'
import zh from '@/i18n/locales/zh'
import zhHant from '@/i18n/locales/zh-Hant'
import ja from '@/i18n/locales/ja'
import ar from '@/i18n/locales/ar'

describe('research group locales', () => {
  it.each([
    ['en', en],
    ['zh', zh],
    ['zh-Hant', zhHant],
    ['ja', ja],
    ['ar', ar],
  ])('provides navigation, owner, member, invitation, and billing copy in %s', (_locale, messages) => {
    expect(messages.nav.researchGroup).toBeTruthy()
    expect(messages.researchGroup.title).toBeTruthy()
    expect(messages.researchGroup.members.title).toBeTruthy()
    expect(messages.researchGroup.invitations.accept).toBeTruthy()
    expect(messages.researchGroup.dashboard.personalBalance).toBeTruthy()
    expect(messages.researchGroup.errors.bothBalancesInsufficient).toBeTruthy()
  })
})
