import zh from './zh'
import { zhHantResearchGroup } from './researchGroupExtra'

export default {
  ...zh,
  nav: {
    ...zh.nav,
    researchGroup: '課題組'
  },
  researchGroup: zhHantResearchGroup,
  home: {
    ...zh.home,
    viewOnGithub: '在 GitHub 上查看',
    viewDocs: '查看文件',
    docs: '文件',
    switchToLight: '切換到淺色模式',
    switchToDark: '切換到深色模式',
    dashboard: '控制台',
    login: '登入',
    register: '註冊',
    getStarted: '立即開始',
    goToDashboard: '進入控制台',
    nav: {
      home: '首頁',
      dashboard: '儀表板',
      tagline: '統一 AI API 閘道',
      pricing: '模型價格',
      integrations: '工具整合',
      openMenu: '開啟選單',
      closeMenu: '關閉選單'
    },
    heroEyebrow: 'OpenAI 相容介面 · 多模型統一轉發',
    heroTitle: '一個 API 金鑰',
    heroAccent: '串接主流 AI 模型',
    heroSubtitle: '一個金鑰，暢用多個 AI 模型',
    heroDescription:
      '無需管理多個訂閱帳號，一站式接入 Claude、GPT、Gemini、DeepSeek、通義千問、豆包、Kimi 等主流 AI 服務',
    supportedAccess: '支援以下部署方式',
    apiDemo: {
      request: '請求',
      response: '回應',
      routed: '已成功轉送'
    },
    heroStats: {
      recharge: '人民幣充值 : 美元餘額',
      groupRate: '分組倍率',
      models: '主要模型',
      deployments: '部署選項'
    },
    ccSwitch: {
      badge: '推薦接入方式',
      title: '用 CC-Switch，一鍵設定你的 AI 開發環境',
      description: '選擇一個 API 金鑰，即可一鍵設定 Codex、Claude Code、Gemini CLI 與 OpenCode；也可透過主流 IDE 終端和遠端 SSH 接入。',
      primaryAction: '立即一鍵設定',
      downloadAction: '下載 CC-Switch',
      guideAction: '查看本機與遠端設定教學',
      panelTitle: '一個入口，連接常用 AI 開發工具',
      oneClick: '一鍵完成',
      ready: '已就緒',
      localSetup: '本機自動匯入設定',
      remoteSetup: '遠端 SSH 一鍵命令',
      noManual: '無需手動編輯金鑰、端點或環境變數'
    },
    integrations: {
      eyebrow: '開發工具接入',
      title: 'Codex、Claude Code 與主流 IDE 工作流程，統一接入',
      description: '同一個 API 金鑰覆蓋命令列、編輯器終端與遠端開發環境。',
      codexDescription: '自動寫入 OpenAI 相容端點與模型設定，開箱即用。',
      claudeCodeDescription: '自動設定 Anthropic 相容端點，本機與遠端環境都支援。',
      geminiDescription: '一鍵產生 Gemini CLI 所需設定，無需反覆切換金鑰。',
      ideDescription: '適配 VS Code、Cursor、Windsurf、JetBrains 等編輯器的內建終端與相容 API。'
    },
    workflow: {
      eyebrow: '快速開始',
      title: '三步開始呼叫',
      description: '從註冊到在本機或伺服器送出請求，只需完成一次統一設定。',
      steps: {
        account: {
          title: '註冊並建立 API 金鑰',
          description: '登入控制台，建立一個可供所有相容用戶端重複使用的 API 金鑰。'
        },
        configure: {
          title: '選擇設定方式',
          description: '使用 CC-Switch 一鍵匯入，或複製終端命令設定本機與遠端伺服器。'
        },
        connect: {
          title: '開始呼叫模型',
          description: '透過 OpenAI 或 Anthropic 相容介面，直接在 Codex、Claude Code 與 IDE 中使用。'
        }
      }
    },
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
      title: '精選核心模型',
      description: '只保留最常用的六大模型系列，一個金鑰統一呼叫，選擇更清晰。',
      supported: '已支援',
      soon: '即將推出',
      more: '更多',
      items: {
        openai: 'GPT',
        claude: 'Claude',
        gemini: 'Gemini',
        xai: 'Grok',
        meta: 'Meta',
        mistral: 'Mistral',
        cohere: 'Cohere',
        midjourney: 'Midjourney',
        perplexity: 'Perplexity',
        openrouter: 'OpenRouter',
        deepseek: 'DeepSeek',
        qwen: '通義千問',
        doubao: '豆包',
        zhipu: 'GLM',
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
      open: '開啟系統公告',
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
