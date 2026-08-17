import { describe, expect, it } from 'vitest'
import { OPENAI_CODEX_DEFAULT_MODEL } from '@/constants/codex'
import chatViewSource from '@/views/chat/ChatView.vue?raw'
import {
  createChatModelOptions,
  fallbackChatModelOptions,
  isImageModelOption,
  resolveAvailableModel,
  resolvePromptAgentModel,
} from '../modelOptions'

describe('chat model options', () => {
  it('discovers model options before loading or creating a chat session', () => {
    expect(chatViewSource).toContain('createChatModelOptions(await listModels())')
    expect(chatViewSource).toContain('modelOptionsLoading.value = false')
    expect(chatViewSource.indexOf('loadModelOptions(),')).toBeLessThan(
      chatViewSource.indexOf('await loadSessions()'),
    )
  })

  it('normalizes and deduplicates models returned by the authenticated gateway', () => {
    expect(
      createChatModelOptions([' claude-sonnet-4-5 ', 'gemini-2.5-pro', 'claude-sonnet-4-5', '']),
    ).toEqual([
      { id: 'claude-sonnet-4-5', label: 'claude-sonnet-4-5', kind: 'chat', provider: undefined, default: false, capabilities: { vision: undefined, image: undefined, webSearch: undefined, contextLimit: undefined } },
      { id: 'gemini-2.5-pro', label: 'gemini-2.5-pro', kind: 'chat', provider: undefined, default: false, capabilities: { vision: undefined, image: undefined, webSearch: undefined, contextLimit: undefined } },
    ])
  })

  it('uses typed metadata rather than a model-name prefix for image capability', () => {
    const options = createChatModelOptions(['gpt-image-2', 'gpt-image-experimental'])

    expect(isImageModelOption(options, 'gpt-image-2')).toBe(true)
    expect(isImageModelOption(options, 'gpt-image-experimental')).toBe(false)
  })

  it('falls back to the first available model when a saved model is unavailable', () => {
    const options = createChatModelOptions([
      { id: 'claude-sonnet-4-5' },
      { id: 'gemini-2.5-pro', default: true },
    ])

    expect(resolveAvailableModel(options, 'retired-model')).toBe('gemini-2.5-pro')
    expect(resolveAvailableModel(options, 'gemini-2.5-pro')).toBe('gemini-2.5-pro')
  })

  it('uses configured capabilities and a dynamic text model for helper calls', () => {
    const options = createChatModelOptions([
      { id: 'draw', display_name: 'Image', owned_by: 'openai', default: true, capabilities: { image: true } },
      { id: 'claude', display_name: 'Claude', owned_by: 'anthropic', capabilities: { vision: true, web_search: true } },
    ])
    expect(options[0]).toMatchObject({ label: 'Image', kind: 'image', provider: 'openai', default: true })
    expect(options[1]?.capabilities).toMatchObject({ vision: true, webSearch: true })
    expect(resolvePromptAgentModel(options, 'draw')).toBe('claude')
  })

  it('provides a stable local fallback when model discovery is unavailable', () => {
    const fallback = fallbackChatModelOptions()

    expect(fallback[0]?.id).toBe(OPENAI_CODEX_DEFAULT_MODEL)
    expect(fallback.some((option) => option.kind === 'image')).toBe(true)
  })
})
