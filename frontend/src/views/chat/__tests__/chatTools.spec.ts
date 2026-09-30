import { describe, expect, it } from 'vitest'
import { buildToolList, formatCurrentTime, parseToolArguments } from '../chatTools'

function toolNames(tools: unknown[]): string[] {
  return tools.map((tool) => (tool as { function: { name: string } }).function.name)
}

describe('buildToolList', () => {
  it('returns no tools for image models', () => {
    const tools = buildToolList({ isChatModel: false, webToolsOn: true, imageModelId: 'gpt-image-2' })
    expect(tools).toEqual([])
  })

  it('always offers current_time for chat models', () => {
    const tools = buildToolList({ isChatModel: true, webToolsOn: false, imageModelId: '' })
    expect(toolNames(tools)).toEqual(['current_time'])
  })

  it('offers web tools only when web tools are on', () => {
    const tools = buildToolList({ isChatModel: true, webToolsOn: true, imageModelId: '' })
    expect(toolNames(tools)).toEqual(['web_search', 'web_fetch', 'current_time'])
  })

  it('offers generate_image only when an image model exists', () => {
    const tools = buildToolList({ isChatModel: true, webToolsOn: false, imageModelId: 'gpt-image-2' })
    expect(toolNames(tools)).toEqual(['generate_image', 'current_time'])
  })

  it('offers everything when all capabilities are available', () => {
    const tools = buildToolList({ isChatModel: true, webToolsOn: true, imageModelId: 'gpt-image-2' })
    expect(toolNames(tools)).toEqual(['web_search', 'web_fetch', 'generate_image', 'current_time'])
  })
})

describe('parseToolArguments', () => {
  it('parses string fields', () => {
    expect(parseToolArguments('{"query":"天气"}')).toEqual({ query: '天气' })
  })

  it('drops non-string values and unknown shapes', () => {
    expect(parseToolArguments('{"url":"https://a","n":1}')).toEqual({ url: 'https://a' })
    expect(parseToolArguments('[]')).toEqual({})
  })

  it('returns empty object for invalid JSON or empty input', () => {
    expect(parseToolArguments('not json')).toEqual({})
    expect(parseToolArguments('')).toEqual({})
  })
})

describe('formatCurrentTime', () => {
  it('formats a readable timestamp', () => {
    expect(typeof formatCurrentTime('')).toBe('string')
    expect(formatCurrentTime('').length).toBeGreaterThan(0)
  })

  it('falls back to the local timezone for an invalid IANA name', () => {
    expect(typeof formatCurrentTime('Not/AZone')).toBe('string')
  })
})
