import zh from './zh'

export default {
  ...zh,
  home: {
    ...zh.home,
    viewOnGithub: '在 GitHub 上查看',
    viewDocs: '查看文件',
    docs: '文件',
    switchToLight: '切換到淺色模式',
    switchToDark: '切換到深色模式',
    dashboard: '控制台',
    login: '登入',
    getStarted: '立即開始',
    goToDashboard: '進入控制台',
    nav: {
      tagline: '統一 AI API 閘道'
    },
    heroEyebrow: 'OpenAI 相容介面 · 多模型統一轉發',
    heroSubtitle: '一個金鑰，暢用多個 AI 模型',
    heroDescription:
      '無需管理多個訂閱帳號，一站式接入 Claude、GPT、Gemini、DeepSeek、通義千問、豆包、Kimi 等主流 AI 服務',
    stats: {
      providers: '模型供應商',
      languages: '站點語言',
      compatible: '相容協定'
    },
    apiCard: {
      protocol: 'OpenAI Compatible',
      title: '即刻接入統一 API',
      subtitle: '複製端點，使用現有 OpenAI SDK 即可呼叫。',
      endpoint: '介面端點',
      copy: '複製',
      copied: '已複製'
    },
    tags: {
      subscriptionToApi: '統一 API 接入',
      stickySession: '會話保持',
      realtimeBilling: '按量計費'
    },
    providers: {
      ...zh.home.providers,
      title: '支援眾多的大模型供應商',
      description: '覆蓋國際主流模型、國內廠商與國產模型，一個 API，多種選擇',
      supported: '已支援',
      soon: '即將推出',
      more: '更多',
      items: {
        openai: 'OpenAI',
        claude: 'Claude',
        gemini: 'Gemini',
        xai: 'xAI',
        meta: 'Meta',
        mistral: 'Mistral',
        cohere: 'Cohere',
        midjourney: 'Midjourney',
        perplexity: 'Perplexity',
        openrouter: 'OpenRouter',
        deepseek: 'DeepSeek',
        qwen: '通義千問',
        doubao: '豆包',
        zhipu: '智譜 GLM',
        moonshot: 'Kimi / Moonshot',
        baidu: '文心 ERNIE',
        tencent: '騰訊混元',
        iflytek: '訊飛星火',
        minimax: 'MiniMax',
        ai360: '360 智腦'
      }
    },
    domestic: {
      title: '國內廠商與國產模型已加入展示',
      description: '重點展示 DeepSeek、通義千問、豆包、Kimi、智譜、文心、混元、星火等國內模型生態。'
    },
    highlights: {
      gateway: {
        title: '統一閘道',
        desc: '相容 OpenAI 格式，將不同上游模型收斂到一套呼叫入口。'
      },
      providers: {
        title: '多供應商',
        desc: '參考 new-api 的多通道理念，統一管理國際模型與國內模型供應商。'
      },
      billing: {
        title: '額度可控',
        desc: '面向 API Key、分組和模型倍率做清晰的用量與費用控制。'
      },
      languages: {
        title: '多語言',
        desc: '提供簡體中文、繁體中文、日文、阿拉伯文，並保留英文兜底。'
      }
    },
    notice: {
      title: '系統公告',
      tabNotice: '通知',
      tabSystem: '系統公告',
      important: '（重要）',
      qqGroupLabel: '交流群：',
      qqGroupSuffix: '，有問題可以進群聯繫站長。',
      trust: '請認準 ISACAI 官方站點與目前 API 入口，避免使用不明來源連結。',
      status: '目前服務入口：',
      closeToday: '今日關閉',
      close: '關閉公告'
    },
    footer: {
      allRightsReserved: '保留所有權利。'
    }
  }
}
