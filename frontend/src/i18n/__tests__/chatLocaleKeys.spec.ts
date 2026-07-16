import { describe, expect, it } from 'vitest'

import chatViewSource from '../../views/chat/ChatView.vue?raw'
import ar from '../locales/ar'
import en from '../locales/en'
import ja from '../locales/ja'
import zh from '../locales/zh'
import zhHant from '../locales/zh-Hant'

type Messages = Record<string, unknown>

const chatKeys = [...new Set(chatViewSource.match(/chat\.[A-Za-z0-9_.]+/g) ?? [])].sort()

function resolveMessage(messages: Messages, key: string): unknown {
  return key.split('.').reduce<unknown>((value, segment) => {
    if (!value || typeof value !== 'object') return undefined
    return (value as Messages)[segment]
  }, messages)
}

describe('ChatView locale coverage', () => {
  const locales: Record<string, Messages> = { zh, 'zh-Hant': zhHant, en, ja, ar }

  it.each(Object.entries(locales))('defines every ChatView message for %s', (_locale, messages) => {
    const missing = chatKeys.filter((key) => {
      const message = resolveMessage(messages, key)
      return typeof message !== 'string' || message.trim().length === 0
    })
    expect(missing).toEqual([])
  })

  it.each(Object.entries(locales))('defines the chat navigation label for %s', (_locale, messages) => {
    expect(resolveMessage(messages, 'nav.chat')).toEqual(expect.any(String))
  })
})
