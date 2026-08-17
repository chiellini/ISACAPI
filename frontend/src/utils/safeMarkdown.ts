import DOMPurify from 'dompurify'
import { marked } from 'marked'
import { sanitizeImageUrl, sanitizeUrl } from '@/utils/url'

export interface SafeMarkdownOptions {
  allowSameOriginImages?: boolean
}

const SAFE_MARKDOWN_TAGS = [
  'a',
  'b',
  'blockquote',
  'br',
  'code',
  'del',
  'div',
  'em',
  'h1',
  'h2',
  'h3',
  'h4',
  'h5',
  'h6',
  'hr',
  'i',
  'li',
  'ol',
  'p',
  'pre',
  's',
  'span',
  'strong',
  'table',
  'tbody',
  'td',
  'th',
  'thead',
  'tr',
  'ul',
]

export function sanitizeMarkdownImageUrl(value: string): string {
  const source = value.trim()
  if (!source) return ''

  if (source.toLowerCase().startsWith('data:')) {
    return sanitizeImageUrl(source, { allowDataUrl: true })
  }

  const currentOrigin = typeof window === 'undefined' ? '' : window.location.origin
  if (source.toLowerCase().startsWith('blob:')) {
    if (!currentOrigin) return ''
    try {
      const parsed = new URL(source)
      return parsed.origin === currentOrigin ? parsed.toString() : ''
    } catch {
      return ''
    }
  }

  const safeUrl = sanitizeUrl(source, { allowRelative: true })
  if (!safeUrl) return ''
  if (safeUrl.startsWith('/')) return safeUrl
  if (!currentOrigin) return ''

  try {
    return new URL(safeUrl).origin === currentOrigin ? safeUrl : ''
  } catch {
    return ''
  }
}

/** Render public/admin-authored Markdown without active UI or tracking media. */
export function renderSafeMarkdown(content: string, options: SafeMarkdownOptions = {}): string {
  if (!content) return ''

  const html = marked.parse(content, { breaks: true, gfm: true }) as string
  const rejectedImages = new WeakSet<Node>()
  DOMPurify.addHook('uponSanitizeAttribute', (node, data) => {
    if (data.attrName === 'href') {
      const safeUrl = sanitizeUrl(data.attrValue, { allowRelative: true })
      if (!safeUrl) {
        data.keepAttr = false
        return
      }
      data.attrValue = safeUrl
      return
    }

    if (data.attrName === 'src' && node.nodeName.toLowerCase() === 'img') {
      const safeUrl = options.allowSameOriginImages ? sanitizeMarkdownImageUrl(data.attrValue) : ''
      if (!safeUrl) {
        data.keepAttr = false
        rejectedImages.add(node)
        return
      }
      data.attrValue = safeUrl
    }
  })
  DOMPurify.addHook('afterSanitizeAttributes', (node) => {
    if (node.nodeName.toLowerCase() !== 'img') return
    const image = node as Element
    if (rejectedImages.has(node) || !image.getAttribute('src')) {
      node.parentNode?.removeChild(node)
    }
  })
  try {
    return DOMPurify.sanitize(html, {
      ALLOWED_TAGS: options.allowSameOriginImages ? [...SAFE_MARKDOWN_TAGS, 'img'] : SAFE_MARKDOWN_TAGS,
      ALLOWED_ATTR: options.allowSameOriginImages
        ? ['alt', 'colspan', 'href', 'rowspan', 'src', 'title']
        : ['colspan', 'href', 'rowspan', 'title'],
      ALLOWED_URI_REGEXP: /^(?:(?:https?|blob|data):|\/)/i,
      ALLOW_ARIA_ATTR: false,
      ALLOW_DATA_ATTR: false,
    })
  } finally {
    DOMPurify.removeHook('uponSanitizeAttribute')
    DOMPurify.removeHook('afterSanitizeAttributes')
  }
}
