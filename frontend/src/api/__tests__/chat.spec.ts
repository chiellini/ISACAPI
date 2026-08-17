import { beforeEach, describe, expect, it, vi } from 'vitest'

const apiClient = vi.hoisted(() => ({
  get: vi.fn(),
  post: vi.fn(),
  put: vi.fn(),
  delete: vi.fn(),
}))
const refreshAuthTokens = vi.hoisted(() => vi.fn())

vi.mock('@/api/client', () => ({
  apiClient,
}))
vi.mock('@/api/tokenRefresh', () => ({
  refreshAuthTokens,
}))

import { completeChat, generateImage, listModels } from '@/api/chat'

describe('chat api', () => {
  const fetchMock = vi.fn()

  beforeEach(() => {
    localStorage.clear()
    fetchMock.mockReset()
    refreshAuthTokens.mockReset()
    apiClient.get.mockReset()
    vi.stubGlobal('fetch', fetchMock)
  })

  it('requests image generation as a stream and parses JSON fallback responses', async () => {
    localStorage.setItem('auth_token', 'jwt-token')
    fetchMock.mockResolvedValue(new Response(
      JSON.stringify({ data: [{ b64_json: 'aGVsbG8=' }] }),
      { status: 200, headers: { 'Content-Type': 'application/json' } },
    ))

    const images = await generateImage({ model: 'gpt-image-2', prompt: 'draw a cat' })

    expect(images).toEqual(['data:image/png;base64,aGVsbG8='])
    expect(fetchMock).toHaveBeenCalledWith(
      '/api/v1/chat/v1/images/generations',
      expect.objectContaining({ method: 'POST' }),
    )
    expect((fetchMock.mock.calls[0][1].headers as Headers).get('Authorization'))
      .toBe('Bearer jwt-token')
    const requestBody = JSON.parse(String(fetchMock.mock.calls[0][1].body))
    expect(requestBody).toEqual({
      model: 'gpt-image-2',
      prompt: 'draw a cat',
      stream: true,
      response_format: 'b64_json',
    })
  })

  it('parses image generation SSE and ignores partial images', async () => {
    fetchMock.mockResolvedValue(new Response(
      [
        'event: image_generation.partial_image',
        'data: {"type":"image_generation.partial_image","b64_json":"cGFydGlhbA=="}',
        '',
        'event: image_generation.completed',
        'data: {"type":"image_generation.completed","b64_json":"ZmluYWw="}',
        '',
        'data: [DONE]',
        '',
      ].join('\n'),
      { status: 200, headers: { 'Content-Type': 'text/event-stream' } },
    ))

    const images = await generateImage({ model: 'gpt-image-2', prompt: 'draw a cat' })

    expect(images).toEqual(['data:image/png;base64,ZmluYWw='])
  })

  it('parses Responses image results nested under response output', async () => {
    fetchMock.mockResolvedValue(new Response(
      [
        'data: {"type":"response.completed","response":{"output":[{"type":"image_generation_call","result":"d2VicA==","output_format":"webp"}]}}',
        '',
        'data: [DONE]',
        '',
      ].join('\n'),
      { status: 200, headers: { 'Content-Type': 'text/event-stream' } },
    ))

    const images = await generateImage({ model: 'gpt-image-2', prompt: 'draw a cat' })

    expect(images).toEqual(['data:image/webp;base64,d2VicA=='])
  })

  it('parses Responses output_item.done image results', async () => {
    fetchMock.mockResolvedValue(new Response(
      [
        'data: {"type":"response.output_item.done","item":{"type":"image_generation_call","result":"anBn","output_format":"jpg"}}',
        '',
        'data: [DONE]',
        '',
      ].join('\n'),
      { status: 200, headers: { 'Content-Type': 'text/event-stream' } },
    ))

    const images = await generateImage({ model: 'gpt-image-2', prompt: 'draw a cat' })

    expect(images).toEqual(['data:image/jpeg;base64,anBn'])
  })

  it('does not load upstream-controlled image URLs or active image formats', async () => {
    fetchMock.mockResolvedValue(new Response(
      JSON.stringify({ data: [
        { url: 'https://tracker.example/pixel.png' },
        { b64_json: 'data:image/svg+xml;base64,PHN2Zz48L3N2Zz4=' },
      ] }),
      { status: 200, headers: { 'Content-Type': 'application/json' } },
    ))

    await expect(generateImage({ model: 'gpt-image-2', prompt: 'draw a cat' })).resolves.toEqual([])
  })

  it('normalizes HTML gateway timeout errors', async () => {
    fetchMock.mockResolvedValue(new Response(
      '<html><body><h1>504 Gateway Time-out</h1></body></html>',
      { status: 504, headers: { 'Content-Type': 'text/html' } },
    ))

    await expect(generateImage({ model: 'gpt-image-2', prompt: 'draw a cat' }))
      .rejects.toThrow('image request failed: 504')
  })

  it('completeChat accumulates streamed deltas into the full reply', async () => {
    localStorage.setItem('auth_token', 'jwt-token')
    fetchMock.mockResolvedValue(new Response(
      [
        'data: {"choices":[{"delta":{"content":"Hello"}}]}',
        '',
        'data: {"choices":[{"delta":{"content":", "}}]}',
        '',
        'data: {"choices":[{"delta":{"content":"world"}}]}',
        '',
        'data: [DONE]',
        '',
      ].join('\n'),
      { status: 200, headers: { 'Content-Type': 'text/event-stream' } },
    ))

    const text = await completeChat({
      model: 'gpt-5.5',
      messages: [{ role: 'user', content: 'hi' }],
    })

    expect(text).toBe('Hello, world')
    expect(fetchMock).toHaveBeenCalledWith(
      '/api/v1/chat/v1/chat/completions',
      expect.objectContaining({ method: 'POST' }),
    )
    const requestBody = JSON.parse(String(fetchMock.mock.calls[0][1].body))
    expect(requestBody.stream).toBe(true)
    expect(requestBody.messages).toEqual([{ role: 'user', content: 'hi' }])
  })

  it('refreshes an expired token and retries a native streaming request once', async () => {
    localStorage.setItem('auth_token', 'expired-token')
    localStorage.setItem('refresh_token', 'refresh-token')
    refreshAuthTokens.mockResolvedValue({
      access_token: 'fresh-token',
      refresh_token: 'rotated-refresh-token',
      expires_in: 3600,
      token_type: 'Bearer',
    })
    fetchMock
      .mockResolvedValueOnce(new Response('expired', { status: 401 }))
      .mockResolvedValueOnce(new Response(
        ['data: {"choices":[{"delta":{"content":"retried"}}]}', '', 'data: [DONE]', ''].join('\n'),
        { status: 200, headers: { 'Content-Type': 'text/event-stream' } },
      ))

    await expect(completeChat({
      model: 'claude-sonnet',
      messages: [{ role: 'user', content: 'hi' }],
    })).resolves.toBe('retried')

    expect(refreshAuthTokens).toHaveBeenCalledWith({ failedAccessToken: 'expired-token' })
    expect(fetchMock).toHaveBeenCalledTimes(2)
    expect((fetchMock.mock.calls[1][1].headers as Headers).get('Authorization'))
      .toBe('Bearer fresh-token')
  })

  it('loads the authenticated model list and ignores malformed entries', async () => {
    apiClient.get.mockResolvedValue({ data: {
      data: [
        { id: 'claude-sonnet-4-5', display_name: 'Claude', owned_by: 'anthropic', default: true, capabilities: { vision: true, web_search: false } },
        { id: 'gemini-2.5-pro', display_name: 'Gemini', owned_by: 'gemini', capabilities: { image: false } },
        { id: 42 },
        null,
      ],
    } })

    await expect(listModels()).resolves.toEqual([
      {
        id: 'claude-sonnet-4-5', display_name: 'Claude', owned_by: 'anthropic', default: true,
        capabilities: { vision: true, image: undefined, web_search: false, context_limit: undefined },
      },
      {
        id: 'gemini-2.5-pro', display_name: 'Gemini', owned_by: 'gemini', default: false,
        capabilities: { vision: undefined, image: false, web_search: undefined, context_limit: undefined },
      },
    ])
    expect(apiClient.get).toHaveBeenCalledWith('/chat/v1/models')
  })

  it('returns an empty model list for an unexpected gateway payload', async () => {
    apiClient.get.mockResolvedValue({ data: { data: null } })

    await expect(listModels()).resolves.toEqual([])
  })
})
