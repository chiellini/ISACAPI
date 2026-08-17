import { OPENAI_CODEX_DEFAULT_MODEL } from '@/constants/codex'
import type { ChatModelDescriptor } from '@/api/chat'

export type ChatModelKind = 'chat' | 'image'

export interface ChatModelOption {
  id: string
  label: string
  kind: ChatModelKind
  provider?: string
  default: boolean
  capabilities: {
    vision?: boolean
    image?: boolean
    webSearch?: boolean
    contextLimit?: number
  }
}

const KNOWN_MODEL_OPTIONS: Readonly<Partial<Record<string, Pick<ChatModelOption, 'label' | 'kind'>>>> = {
  [OPENAI_CODEX_DEFAULT_MODEL]: {
    label: OPENAI_CODEX_DEFAULT_MODEL,
    kind: 'chat',
  },
  'gpt-image-2': {
    label: 'GPT Image 2',
    kind: 'image',
  },
}

export function fallbackChatModelOptions(): ChatModelOption[] {
  return [OPENAI_CODEX_DEFAULT_MODEL, 'gpt-image-2'].map((id) => toModelOption(id))
}

export function createChatModelOptions(models: readonly (string | ChatModelDescriptor)[]): ChatModelOption[] {
  const seen = new Set<string>()
  const options: ChatModelOption[] = []

  for (const value of models) {
    const descriptor = typeof value === 'string' ? { id: value } : value
    const id = descriptor.id.trim()
    if (!id || seen.has(id)) continue
    seen.add(id)
    options.push(toModelOption(id, descriptor))
  }

  return options
}

export function resolveAvailableModel(
  options: readonly ChatModelOption[],
  preferred?: string | null,
): string {
  const normalized = preferred?.trim() || ''
  if (normalized && options.some((option) => option.id === normalized)) return normalized
  return options.find((option) => option.default)?.id || options[0]?.id || ''
}

export function isImageModelOption(
  options: readonly ChatModelOption[],
  modelId: string,
): boolean {
  return options.find((option) => option.id === modelId)?.kind === 'image'
}

export function resolvePromptAgentModel(options: readonly ChatModelOption[], selected?: string): string {
  const selectedOption = options.find((option) => option.id === selected && option.kind === 'chat')
  return selectedOption?.id
    || options.find((option) => option.default && option.kind === 'chat')?.id
    || options.find((option) => option.kind === 'chat')?.id
    || ''
}

function toModelOption(id: string, descriptor?: ChatModelDescriptor): ChatModelOption {
  const known = KNOWN_MODEL_OPTIONS[id]
  // Prefer server policy metadata and only use the legacy table when an older
  // gateway returns IDs without capabilities.
  const image = descriptor?.capabilities?.image ?? (known?.kind === 'image')
  return {
    id,
    label: descriptor?.display_name || known?.label || id,
    kind: image ? 'image' : 'chat',
    provider: descriptor?.owned_by,
    default: descriptor?.default === true,
    capabilities: {
      vision: descriptor?.capabilities?.vision,
      image: descriptor?.capabilities?.image,
      webSearch: descriptor?.capabilities?.web_search,
      contextLimit: descriptor?.capabilities?.context_limit,
    },
  }
}
