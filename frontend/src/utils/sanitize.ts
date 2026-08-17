import DOMPurify from 'dompurify'

export function sanitizeSvg(svg: string): string {
  if (!svg) return ''
  DOMPurify.addHook('uponSanitizeAttribute', (_node, data) => {
    if (/url\s*\(|(?:https?:)?\/\//i.test(data.attrValue)) {
      data.keepAttr = false
    }
  })
  try {
    return DOMPurify.sanitize(svg, {
      USE_PROFILES: { svg: true, svgFilters: true },
      // SVG snippets are rendered inline. Disallow elements and attributes that
      // can turn an icon into a navigation or an external-resource request.
      FORBID_TAGS: ['a', 'embed', 'foreignobject', 'iframe', 'image', 'object', 'script', 'style', 'use'],
      FORBID_ATTR: ['href', 'style', 'xlink:href'],
    })
  } finally {
    DOMPurify.removeHook('uponSanitizeAttribute')
  }
}
