/**
 * 聊天 agent 工具注册表（纯逻辑，便于单测）。
 *
 * buildToolList 按上下文（模型能力 / 用户开关）组装 OpenAI tools 数组；
 * parseToolArguments 容错解析模型输出的工具参数。
 *
 * 服务端 ChatPolicyMiddleware 会按白名单把 tools 重写为规范定义（见后端
 * internal/server/middleware/chat_policy.go 的 chatToolSpecs），因此这里的
 * schema 只表达客户端意图，真正到达模型的是服务端审核过的定义。
 */

export type ChatToolName = 'web_search' | 'web_fetch' | 'generate_image' | 'current_time'

export interface ChatToolContext {
  /** 当前选中模型是否为文本模型（工具仅对文本模型开放）。 */
  isChatModel: boolean
  /** 联网工具是否开启（平台已配置搜索 + 档案能力允许 + 用户打开开关）。 */
  webToolsOn: boolean
  /** 可用的生图模型 ID（空 = 平台未提供生图模型）。 */
  imageModelId: string
}

interface ChatToolEntry {
  name: ChatToolName
  schema: unknown
  available: (ctx: ChatToolContext) => boolean
}

const CHAT_TOOL_ENTRIES: ChatToolEntry[] = [
  {
    name: 'web_search',
    available: (ctx) => ctx.isChatModel && ctx.webToolsOn,
    schema: {
      type: 'function',
      function: {
        name: 'web_search',
        description: '联网搜索获取实时/最新信息。回答需要最新或可引用的外部信息时调用。',
        parameters: {
          type: 'object',
          properties: {
            query: { type: 'string', description: '检索关键词或问题，使用与用户相同的语言' },
          },
          required: ['query'],
        },
      },
    },
  },
  {
    name: 'web_fetch',
    available: (ctx) => ctx.isChatModel && ctx.webToolsOn,
    schema: {
      type: 'function',
      function: {
        name: 'web_fetch',
        description: '抓取一个公开网页的正文。搜索后需要深入阅读某个结果，或用户给出具体链接时调用。',
        parameters: {
          type: 'object',
          properties: {
            url: { type: 'string', description: '要阅读的完整 http(s) 链接' },
          },
          required: ['url'],
        },
      },
    },
  },
  {
    name: 'generate_image',
    available: (ctx) => ctx.isChatModel && !!ctx.imageModelId,
    schema: {
      type: 'function',
      function: {
        name: 'generate_image',
        description: '根据文字描述生成图片并附在回复里。用户要求画图、配图、示意图时调用。',
        parameters: {
          type: 'object',
          properties: {
            prompt: { type: 'string', description: '完整、自洽的生图提示词' },
          },
          required: ['prompt'],
        },
      },
    },
  },
  {
    name: 'current_time',
    available: (ctx) => ctx.isChatModel,
    schema: {
      type: 'function',
      function: {
        name: 'current_time',
        description: '获取当前日期与时间（可选 IANA 时区，如 Asia/Shanghai）。涉及“现在/今天”等时间问题时调用。',
        parameters: {
          type: 'object',
          properties: {
            timezone: { type: 'string', description: '可选的 IANA 时区名，默认本地时区' },
          },
        },
      },
    },
  },
]

/** 组装当前可用的 tools 数组（OpenAI 函数工具形态）。 */
export function buildToolList(ctx: ChatToolContext): unknown[] {
  return CHAT_TOOL_ENTRIES.filter((entry) => entry.available(ctx)).map((entry) => entry.schema)
}

/** 容错解析模型输出的工具参数：非 JSON / 非对象 / 非字符串值一律丢弃。 */
export function parseToolArguments(raw: string): Record<string, string> {
  try {
    const parsed: unknown = JSON.parse(raw || '{}')
    if (!parsed || typeof parsed !== 'object' || Array.isArray(parsed)) return {}
    const source = parsed as Record<string, unknown>
    const out: Record<string, string> = {}
    for (const key of Object.keys(source)) {
      if (typeof source[key] === 'string') out[key] = source[key] as string
    }
    return out
  } catch {
    return {}
  }
}

/** 按可选 IANA 时区格式化当前时间；时区非法时回退本地时区。 */
export function formatCurrentTime(timezone: string): string {
  const now = new Date()
  const options: Intl.DateTimeFormatOptions = { dateStyle: 'full', timeStyle: 'long' }
  try {
    return new Intl.DateTimeFormat('default', timezone ? { ...options, timeZone: timezone } : options).format(now)
  } catch {
    return new Intl.DateTimeFormat('default', options).format(now)
  }
}
