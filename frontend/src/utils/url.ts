/**
 * 验证并规范化 URL
 * 默认只接受绝对 URL（以 http:// 或 https:// 开头），可按需允许相对路径
 * @param value 用户输入的 URL
 * @returns 规范化后的 URL，如果无效则返回空字符串
 */
type SanitizeOptions = {
  allowRelative?: boolean
  allowDataUrl?: boolean
}

export function hasUnsafeUrlPathCharacters(value: string): boolean {
  if (value.includes('\\')) return true
  for (const character of value) {
    const codePoint = character.codePointAt(0) ?? 0
    if (codePoint <= 0x1f || codePoint === 0x7f) return true
  }
  return false
}

const SAFE_RASTER_DATA_IMAGE = /^data:image\/(?:avif|bmp|gif|jpeg|png|webp);base64,[a-z\d+/=\s]+$/i

export function sanitizeImageUrl(value: string, options: { allowRelative?: boolean; allowDataUrl?: boolean } = {}): string {
  const source = value.trim()
  if (!source) return ''
  if (source.toLowerCase().startsWith('data:')) {
    return options.allowDataUrl && SAFE_RASTER_DATA_IMAGE.test(source) ? source : ''
  }
  return sanitizeUrl(source, { allowRelative: options.allowRelative })
}

export function sanitizeUrl(value: string, options: SanitizeOptions = {}): string {
  const trimmed = value.trim()
  if (!trimmed) {
    return ''
  }

  if (options.allowRelative && trimmed.startsWith('/')) {
    // Browsers treat backslashes as slashes for special URLs. Values such as
    // `/\\attacker.example` can therefore become cross-origin even though they
    // do not start with `//` at the string level.
    if (/^\/[\\/]/.test(trimmed) || hasUnsafeUrlPathCharacters(trimmed)) {
      return ''
    }
    return trimmed
  }

  // 允许 data:image/ 开头的 data URL（仅限图片类型）
  if (options.allowDataUrl && trimmed.startsWith('data:image/')) {
    return trimmed
  }

  // 只接受绝对 URL，不使用 base URL 来避免相对路径被解析为当前域名
  // 检查是否以 http:// 或 https:// 开头
  if (!trimmed.match(/^https?:\/\//i)) {
    return ''
  }

  try {
    const parsed = new URL(trimmed)
    const protocol = parsed.protocol.toLowerCase()
    if (protocol !== 'http:' && protocol !== 'https:') {
      return ''
    }
    return parsed.toString()
  } catch {
    return ''
  }
}
