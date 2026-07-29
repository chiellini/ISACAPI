import { describe, expect, it } from 'vitest'

import ar from '../locales/ar'
import en from '../locales/en'
import ja from '../locales/ja'
import zh from '../locales/zh'
import zhHant from '../locales/zh-Hant'

type Messages = Record<string, unknown>

const requiredKeys = [
  'common.contactAdmin',
  'common.contactAdminHint',
  'common.recharge',
  'common.rechargeContactHint',
  'common.apply',
  'common.clear',
  'common.creating',
  'common.required',
  'common.sending',
  'common.tryAgain'
] as const

function resolveMessage(messages: Messages, key: string): unknown {
  return key.split('.').reduce<unknown>((value, segment) => {
    if (!value || typeof value !== 'object') return undefined
    return (value as Messages)[segment]
  }, messages)
}

describe('shared dashboard locale coverage', () => {
  const locales: Record<string, Messages> = { zh, 'zh-Hant': zhHant, en, ja, ar }

  it.each(Object.entries(locales))('defines contact and recharge labels for %s', (_locale, messages) => {
    const missing = requiredKeys.filter((key) => {
      const message = resolveMessage(messages, key)
      return typeof message !== 'string' || message.trim().length === 0
    })
    expect(missing).toEqual([])
  })
})
