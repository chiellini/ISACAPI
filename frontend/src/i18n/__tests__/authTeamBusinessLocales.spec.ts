import { describe, expect, it } from 'vitest'
import en from '@/i18n/locales/en'
import zh from '@/i18n/locales/zh'
import zhHant from '@/i18n/locales/zh-Hant'
import ja from '@/i18n/locales/ja'
import ar from '@/i18n/locales/ar'

describe('auth team business locales', () => {
  it.each([
    ['en', en],
    ['zh', zh],
    ['zh-Hant', zhHant],
    ['ja', ja],
    ['ar', ar],
  ])('provides team business copy in %s', (_locale, messages) => {
    expect(messages.auth.teamBusiness.audience).toBeTruthy()
    expect(messages.auth.teamBusiness.title).toBeTruthy()
    expect(messages.auth.teamBusiness.description).toBeTruthy()
    expect(messages.auth.teamBusiness.sharedBalance).toBeTruthy()
    expect(messages.auth.teamBusiness.quotaControl).toBeTruthy()
    expect(messages.auth.teamBusiness.corporateInvoice).toBeTruthy()
    expect(messages.auth.teamBusiness.compliantIntegration).toBeTruthy()
    expect(messages.auth.teamBusiness.disclaimer).toBeTruthy()
  })

  it('keeps the required Chinese corporate-service wording explicit', () => {
    expect(zh.auth.teamBusiness.corporateInvoice).toContain('正规对公发票')
    expect(zh.auth.teamBusiness.compliantIntegration).toContain('合规软件信息集成服务')
    expect(zh.auth.teamBusiness.description).toContain('独立账号')
    expect(zh.auth.teamBusiness.description).toContain('查看自己的课题组额度')
  })

  it.each([
    ['zh-Hant', zhHant, zh],
    ['ja', ja, en],
    ['ar', ar, en],
  ])('uses localized copy instead of the fallback for %s', (_locale, messages, fallback) => {
    expect(messages.auth.teamBusiness.title).not.toBe(fallback.auth.teamBusiness.title)
    expect(messages.auth.teamBusiness.description).not.toBe(
      fallback.auth.teamBusiness.description
    )
    expect(messages.auth.teamBusiness.corporateInvoice).not.toBe(
      fallback.auth.teamBusiness.corporateInvoice
    )
  })
})
