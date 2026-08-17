import { renderSafeMarkdown } from '@/utils/safeMarkdown'

/** 把 Markdown 渲染为经过 DOMPurify 清洗的安全 HTML。 */
export function renderMarkdown(text: string): string {
  return renderSafeMarkdown(text, { allowSameOriginImages: true })
}
