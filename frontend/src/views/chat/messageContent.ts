import type { ContentPart } from '@/api/chat'

export interface MessageAttachmentContent {
  kind: 'image' | 'text'
  name: string
  dataUrl?: string
  text?: string
}

// Do not forward historical image parts after the user switches to a model
// whose server policy disables vision. Keep an explicit text marker so the
// conversation remains understandable without implying that the model saw it.
export function buildApiMessageContent(
  content: string,
  attachments: readonly MessageAttachmentContent[] | undefined,
  allowImages: boolean,
): string | ContentPart[] {
  if (!attachments?.length) return content

  let text = content
  for (const attachment of attachments) {
    if (attachment.kind === 'text' && attachment.text) {
      text += `\n\n[${attachment.name}]\n${attachment.text}`
    } else if (attachment.kind === 'image' && !allowImages) {
      text += `\n\n[Image attachment not sent to this model: ${attachment.name}]`
    }
  }
  if (!allowImages) return text

  const parts: ContentPart[] = [{ type: 'text', text }]
  for (const attachment of attachments) {
    if (attachment.kind === 'image' && attachment.dataUrl) {
      parts.push({ type: 'image_url', image_url: { url: attachment.dataUrl } })
    }
  }
  return parts
}
