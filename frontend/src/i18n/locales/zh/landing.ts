export default {
  batchImageGuide: {
    title: '图片批量生成',
    description: '一次提交多条提示词，任务完成后可统一下载图片结果'
  },
  ccSwitchGuide: {
    heroBadge: '本机与远程配置',
    title: 'CC-Switch 使用教程',
    description:
      '导入 API Key 前先选对场景：如果 Claude Code、Codex 或 Gemini CLI 跑在你当前电脑上，就用本机 CC-Switch；如果它们跑在 VS Code Remote SSH 连接的服务器里，就要在远程终端里执行一键配置命令。',
    actions: {
      openKeys: '打开 API 密钥',
      releasePage: '官方 Release'
    },
    remoteWarning: {
      title: '远程 SSH 和本机桌面不是一回事',
      body:
        '在你的笔记本安装 CC-Switch，不能写入 VS Code Remote SSH 服务器里的 ~/.claude 或 ~/.codex。远程 Codex / Claude Code 应该打开已连接服务器的 VS Code 终端，把 API 密钥弹窗里的安装命令粘贴进去执行。'
    },
    scenarios: {
      local: {
        badge: '本机桌面',
        title: '在当前电脑使用 CC-Switch',
        body:
          '适合 Claude Code、Codex、Gemini CLI 就运行在本机 Windows 或 macOS 的情况。浏览器、CC-Switch 和客户端配置文件都在同一台机器上。',
        points: [
          '推荐用于日常本机开发，尤其是 Windows / macOS 桌面环境。',
          '在 API 密钥页点击导入，让 CC-Switch 接收并写入配置。',
          '导入后重启本机 CLI 或编辑器终端，让客户端读取新配置。'
        ]
      },
      remote: {
        badge: '远程 SSH',
        title: '配置 VS Code Remote SSH 里的终端',
        body:
          '适合 VS Code 通过 SSH 连接到服务器，并且 Codex 或 Claude Code 实际在那台服务器上运行的情况。',
        points: [
          '不要指望本机 CC-Switch 桌面应用去修改远程服务器文件。',
          '从密钥弹窗复制一键安装命令，粘贴到 VS Code 的远程终端执行。',
          '命令会在远程主机写入客户端配置、Shell 启动变量和 VS Code 终端环境。'
        ]
      }
    },
    download: {
      title: 'CC-Switch 官方下载',
      body:
        '请只使用官方来源。Windows ARM64、便携版和校验文件可以在 Release 页面选择对应资产。',
      officialSite: '官方网站',
      windows: 'Windows MSI',
      macos: 'macOS DMG',
      release: 'Release 页面'
    },
    local: {
      eyebrow: '路径 A',
      title: '本机 CC-Switch 配置',
      body:
        '当 AI 编程客户端运行在当前电脑上时，这是最省心的路径。',
      steps: [
        {
          title: '安装 CC-Switch',
          body:
            '从官方网站或 Release 页面下载 CC-Switch。Windows 安装 MSI，macOS 安装 DMG，然后至少启动一次 CC-Switch，让系统注册 ccswitch:// 协议。',
          bullets: [
            '从浏览器导入时保持 CC-Switch 正在运行。',
            '浏览器询问是否打开 CC-Switch 时，选择允许。',
            '需要 ARM64 或便携版时，到 Release 页面选择对应文件。'
          ]
        },
        {
          title: '从 API 密钥页导入',
          body:
            '打开 API 密钥，选择要使用的密钥，点击 CC-Switch 导入操作。根据你的工具选择客户端类型。',
          bullets: [
            'Claude Code 会使用 Anthropic 兼容端点。',
            'Codex 会使用 OpenAI 兼容端点和模型映射。',
            'Gemini CLI 会使用 Gemini 兼容端点。'
          ]
        },
        {
          title: '激活并在本机验证',
          body:
            '在 CC-Switch 里确认刚导入的 Provider 已设为当前配置。然后重启 Claude Code、Codex、Gemini CLI 或编辑器终端。',
          bullets: [
            '先发一个小请求验证密钥和端点是否正常。',
            '如果导入没有拉起 CC-Switch，重新安装 CC-Switch，或者改用安装命令。',
            '本机 CC-Switch 只会更新本机客户端配置。'
          ]
        }
      ]
    },
    remote: {
      eyebrow: '路径 B',
      title: '远程 SSH / VS Code / Codex / Claude Code 配置',
      body:
        '当命令实际运行在服务器上时，用这条路径。浏览器可以继续开在你的电脑上，但生成的命令必须粘贴到远程终端执行。',
      steps: [
        {
          title: '在 VS Code 连接服务器',
          body:
            '打开 VS Code，通过 Remote SSH 连接服务器，并确认终端提示符确实在远程主机上。',
          bullets: [
            '看 VS Code 状态栏里的 SSH 主机名。',
            '运行 pwd、hostname 或 whoami，确认不是本机终端。',
            '如果服务器还没有 Codex / Claude Code，先在远程服务器安装。'
          ]
        },
        {
          title: '从本站复制安装命令',
          body:
            '在网页后台打开 API 密钥，点击对应密钥，选择安装命令，再切到 Codex、Claude Code、Gemini CLI 或 OpenCode 标签。',
          bullets: [
            '选择远程服务器的系统标签，不是你笔记本的系统。',
            '大多数 SSH 服务器使用 Linux/macOS 命令。',
            '命令里包含当前密钥，请像保管密钥一样保管它。'
          ]
        },
        {
          title: '粘贴到远程终端执行',
          body:
            '把复制的命令粘贴到 VS Code Remote SSH 终端执行。它会为远程用户写入客户端配置文件和持久化 Shell 环境变量。',
          bullets: [
            'Claude Code 会写入 ~/.claude/settings.json 和 ANTHROPIC_* 环境变量。',
            'Codex 会写入 ~/.codex/config.toml、~/.codex/auth.json 和 OPENAI_* 环境变量。',
            '如果远程有 Node.js，还会把 VS Code terminal env 合并到远程 User settings.json。'
          ]
        },
        {
          title: '重启并测试',
          body:
            '关闭已有远程终端，重新打开一个 VS Code 远程终端，然后从新终端启动 Codex 或 Claude Code。',
          bullets: [
            '更新 profile 后，环境变量通常只会出现在新的 Shell 会话里。',
            '如果 VS Code 终端变量没有生效，重新打开 Remote SSH 窗口。',
            '正式长任务前，先用一个小 prompt 测试。'
          ]
        }
      ],
      filesTitle: '安装命令会修改的文件和变量',
      verifyTitle: '快速验证命令',
      files: [
        'Claude Code：~/.claude/settings.json，以及 ANTHROPIC_BASE_URL / ANTHROPIC_API_KEY。',
        'Codex：~/.codex/config.toml、~/.codex/auth.json，以及 OPENAI_BASE_URL、OPENAI_API_BASE、OPENAI_API_KEY。',
        'Shell profile：~/.bashrc、~/.bash_profile、~/.zshrc、~/.profile，或 PowerShell profile。',
        'VS Code 终端环境：terminal.integrated.env.linux / osx / windows。'
      ]
    },
    troubleshooting: {
      title: '常见检查',
      items: [
        {
          title: '浏览器没有拉起 CC-Switch',
          body:
            '先手动启动一次 CC-Switch，再重试导入。如果协议仍然没有注册，从官方 Release 页面重新安装。'
        },
        {
          title: '远程 Codex 仍然使用旧端点',
          body:
            '确认安装命令是在 Remote SSH 终端执行的，不是在本机终端执行的。命令完成后重新打开远程终端。'
        },
        {
          title: 'Claude Code 找不到密钥',
          body:
            '检查 ~/.claude/settings.json，并在新的终端会话里确认存在 ANTHROPIC_AUTH_TOKEN 或 ANTHROPIC_API_KEY。'
        },
        {
          title: 'VS Code 终端环境变量缺失',
          body:
            '在远程主机安装 Node.js 后重新执行安装命令；或者手动把变量加入 VS Code User settings 的 terminal.integrated.env.linux、osx 或 windows。'
        }
      ]
    }
  },
  pricingPage: {
    eyebrow: 'ISACAI 模型定价',
    title: '模型人民币价格，一目了然',
    description: '所有模型统一按分组倍率 ×1 计算，仅展示人民币每百万 token 的价格；当前充值到账倍率为 1 CNY = {rate} USD。',
    navHome: '首页',
    navPricing: '模型定价',
    searchPlaceholder: '搜索模型名称、厂商或系列…',
    filters: '筛选',
    filterDescription: '按模型系列快速筛选。',
    reset: '重置',
    providers: '模型系列',
    allModels: '全部模型',
    gptModels: 'GPT / OpenAI',
    claudeModels: 'Claude / Anthropic',
    visibleModels: '{count} 个模型',
    priceModeLabel: '价格口径',
    balanceMode: '人民币价格',
    cashMode: '人民币价格',
    unitLabel: '计价单位',
    perMillionShort: 'RMB / 百万',
    perThousandShort: 'RMB / 百万',
    sortLabel: '排序',
    sortName: '名称',
    sortInput: '输入价',
    sortOutput: '输出价',
    input: '输入',
    output: '输出',
    cacheRead: '缓存读取',
    balanceBadge: '人民币价格',
    cashBadge: '人民币价格',
    tokenBilling: '按量计费',
    groupRate: '分组 ×1',
    openaiCompatible: 'OpenAI 兼容',
    copyModel: '复制模型名称',
    copied: '已复制',
    noResultsTitle: '没有找到匹配的模型',
    noResultsDescription: '试试其他关键词，或清除当前模型系列筛选。',
    clearFilters: '清除筛选',
    rechargeLabel: '充值到账',
    rechargeValue: '1 元 = {usd} USD',
    rechargeHint: '充值到账倍率由后台配置，并同步用于实际到账和人民币价格换算。',
    valueLabel: '展示单位',
    valueValue: 'RMB / 百万 token',
    valueHint: '输入、输出与缓存读取价格均按每百万 token 展示。',
    billingLabel: '余额计费',
    billingValue: '分组倍率 ×1',
    billingHint: '输入、输出与缓存读取均统一按分组倍率 ×1 换算。',
    formulaTitle: '价格如何换算',
    formulaBalance: '本页统一按分组倍率 ×1 计算，不按不同分组拆分展示。',
    formulaCash: '人民币价格 = 模型美元基准价 ÷ {rate}，结果按每百万 token 展示。',
    formulaValue: '充值到账倍率更新后，页面中的人民币价格会同步重新换算。',
    disclaimer: '页面展示统一倍率 ×1 下的 token 基础价格；Priority、长上下文、缓存写入、图片及特殊计费规则以实际账单为准。',
    ctaTitle: '选好模型后，直接用同一个 API Key 开始',
    ctaDescription: '支持 CC-Switch、终端命令、远程服务器和本地客户端部署。',
    ctaPrimary: '注册并创建 Key',
    ctaAuthenticated: '进入控制台',
    ctaSecondary: '返回首页',
    modelDescriptions: {
      sol: 'GPT-5.6 旗舰档，适合高难度推理、编码与复杂代理任务。',
      terra: 'GPT-5.6 均衡档，在能力、速度与成本之间取得平衡。',
      luna: 'GPT-5.6 轻量档，适合高频对话与成本敏感型任务。',
      gpt55: '通用旗舰模型，面向复杂推理、代码和长流程工作。',
      opus: 'Claude 高能力档，适合复杂分析、代码与长任务。',
      sonnet: 'Claude 均衡档，适合日常编码、代理和内容处理。',
      haiku: 'Claude 高速轻量档，适合高吞吐与低延迟场景。'
    }
  },

  // Home Page
  home: {
    viewOnGithub: '在 GitHub 上查看',
    viewDocs: '查看文档',
    docs: '文档',
    switchToLight: '切换到浅色模式',
    switchToDark: '切换到深色模式',
    dashboard: '控制台',
    login: '登录',
    register: '注册',
    getStarted: '立即开始',
    goToDashboard: '进入控制台',
    nav: {
      home: '首页',
      dashboard: '控制台',
      tagline: '统一 AI API 网关',
      pricing: '模型定价',
      teams: '课题组与企业',
      integrations: '工具接入',
      openMenu: '打开菜单',
      closeMenu: '关闭菜单'
    },
    heroEyebrow: 'OpenAI 兼容接口 · 多模型统一转发',
    heroTitle: '一个 API Key',
    heroAccent: '连接主流 AI 模型',
    // 新增：面向用户的价值主张
    heroSubtitle: '一个密钥，畅用多个 AI 模型',
    heroDescription: '无需管理多个订阅账号，一站式接入 Claude、GPT、Gemini、DeepSeek、通义千问、豆包、Kimi 等主流 AI 服务',
    supportedAccess: '支持以下部署方式',
    apiDemo: {
      request: '请求',
      response: '响应',
      routed: '路由成功'
    },
    heroStats: {
      recharge: '人民币充值 : 美元余额',
      groupRate: '分组倍率',
      models: '核心模型',
      deployments: '部署方式'
    },
    ccSwitch: {
      badge: '推荐接入方式',
      title: '用 CC-Switch，一键配置你的 AI 开发环境',
      description: '选择一个 API 密钥，即可一键配置 Codex、Claude Code、Gemini CLI 与 OpenCode；也可通过主流 IDE 终端和远程 SSH 接入。',
      primaryAction: '立即一键配置',
      downloadAction: '下载 CC-Switch',
      guideAction: '查看本机与远程配置教程',
      panelTitle: '一个入口，连接常用 AI 开发工具',
      oneClick: '一键完成',
      ready: '已就绪',
      localSetup: '本机自动导入配置',
      remoteSetup: '远程 SSH 一键命令',
      noManual: '无需手动编辑密钥、端点或环境变量'
    },
    authPitch: '登录或注册后，即可通过 CC-Switch、终端命令、远程服务器和本地客户端部署同一套 API 配置。',
    deployment: {
      ccSwitch: {
        title: 'CC-Switch',
        description: '桌面端一键导入'
      },
      terminal: {
        title: '终端部署',
        description: 'CLI 一键命令'
      },
      server: {
        title: '服务器部署',
        description: '远程 SSH 配置'
      },
      client: {
        title: '客户端部署',
        description: 'Codex / Claude Code'
      }
    },
    integrations: {
      eyebrow: '开发工具接入',
      title: 'Codex、Claude Code 与主流 IDE 工作流，统一接入',
      description: '同一个 API 密钥覆盖命令行、编辑器终端和远程开发环境。',
      codexDescription: '自动写入 OpenAI 兼容端点与模型配置，开箱即用。',
      claudeCodeDescription: '自动配置 Anthropic 兼容端点，本机与远程环境都支持。',
      geminiDescription: '一键生成 Gemini CLI 所需配置，无需反复切换密钥。',
      ideDescription: '适配 VS Code、Cursor、Windsurf、JetBrains 等编辑器的内置终端与兼容 API。'
    },
    teams: {
      eyebrow: '高校课题组 · 企业团队',
      title: '一个主账号，清楚支持整个课题组或企业团队',
      description: '负责人注册后即可创建自己的小组，统一管理共享余额和成员月额度。成员继续使用独立账号与 API Key，并可在自己的控制台查看小组额度、已用额度和个人余额。',
      primaryAction: '注册并创建小组',
      manageAction: '创建或管理课题组',
      loginAction: '登录查看小组',
      features: {
        sharedBalance: {
          title: '共享余额，统一充值',
          description: '主账号余额仅资助符合条件的按量计费请求，并优先为成员整笔付款；订阅计费不进入共享余额。资助额度或主账号余额不足时，请求转由成员个人余额承担。'
        },
        independentAccounts: {
          title: '成员独立登录与密钥',
          description: '每位成员保留自己的登录账号和 API Key，负责人无需收集或控制成员密钥。'
        },
        visibleQuota: {
          title: '月额度对成员可见',
          description: '成员可随时查看本月小组额度、已用与剩余额度，同时保留个人充值和用量入口。'
        },
        billingTrace: {
          title: '付款与用量清楚追踪',
          description: '用量归属实际调用成员，扣款记录实际付款方，便于负责人查看由小组承担的成员用量。'
        }
      },
      business: {
        badge: '对公业务支持',
        title: '面向高校与企业的采购和交付支持',
        description: '适合需要统一采购、明确预算边界和持续接入 AI 能力的课题组、实验室及企业团队。',
        invoiceTitle: '正规对公发票',
        invoiceDescription: '在交易主体、业务内容及开票条件符合要求时，可确认正规对公发票的开具信息；不构成无条件开票承诺。',
        integrationTitle: '合规软件信息集成服务',
        integrationDescription: '可根据实际项目沟通软件、开发工具和业务系统的接口接入与信息集成，并按适用要求及双方约定实施。',
        note: '能否开票、发票项目、税务信息、服务范围和交付方式以实际业务、适用规则及双方约定为准；“合规”不表示适用于所有行业或场景的通用认证或法律结论。'
      }
    },
    workflow: {
      eyebrow: '快速开始',
      title: '三步开始调用',
      description: '从注册到在本机或服务器发出请求，只需完成一次统一配置。',
      steps: {
        account: {
          title: '注册并创建 API Key',
          description: '登录控制台，创建一个可供所有兼容客户端复用的 API Key。'
        },
        configure: {
          title: '选择配置方式',
          description: '使用 CC-Switch 一键导入，或复制终端命令配置本机和远程服务器。'
        },
        connect: {
          title: '开始调用模型',
          description: '通过 OpenAI 或 Anthropic 兼容接口，在 Codex、Claude Code 和 IDE 中直接使用。'
        }
      }
    },
    tags: {
      subscriptionToApi: '统一 API 接入',
      stickySession: '会话保持',
      realtimeBilling: '按量计费'
    },
    stats: {
      providers: '模型供应商',
      languages: '站点语言',
      compatible: '兼容协议'
    },
    apiCard: {
      protocol: 'OpenAI Compatible',
      title: '即刻接入统一 API',
      subtitle: '复制端点，使用现有 OpenAI SDK 即可调用。',
      endpoint: '接口端点',
      copy: '复制',
      copied: '已复制'
    },
    // 用户痛点区块
    painPoints: {
      title: '你是否也遇到这些问题？',
      items: {
        expensive: {
          title: '订阅费用高',
          desc: '每个 AI 服务都要单独订阅，每月支出越来越多'
        },
        complex: {
          title: '多账号难管理',
          desc: '不同平台的账号、密钥分散各处，管理起来很麻烦'
        },
        unstable: {
          title: '服务不稳定',
          desc: '单一账号容易触发限制，影响正常使用'
        },
        noControl: {
          title: '用量无法控制',
          desc: '不知道钱花在哪了，也无法限制团队成员的使用'
        }
      }
    },
    // 解决方案区块
    solutions: {
      title: '我们帮你解决',
      subtitle: '简单三步，开始省心使用 AI'
    },
    features: {
      unifiedGateway: '一键接入',
      unifiedGatewayDesc: '获取一个 API 密钥，即可调用所有已接入的 AI 模型，无需分别申请。',
      multiAccount: '稳定可靠',
      multiAccountDesc: '智能调度多个上游账号，自动切换和负载均衡，告别频繁报错。',
      balanceQuota: '用多少付多少',
      balanceQuotaDesc: '按实际使用量计费，支持设置配额上限，团队用量一目了然。'
    },
    // 优势对比
    comparison: {
      title: '为什么选择我们？',
      headers: {
        feature: '对比项',
        official: '官方订阅',
        us: '本平台'
      },
      items: {
        pricing: {
          feature: '付费方式',
          official: '固定月费，用不完也付',
          us: '按量付费，用多少付多少'
        },
        models: {
          feature: '模型选择',
          official: '单一服务商',
          us: '多模型随意切换'
        },
        management: {
          feature: '账号管理',
          official: '每个服务单独管理',
          us: '统一密钥，一站管理'
        },
        stability: {
          feature: '服务稳定性',
          official: '单账号易触发限制',
          us: '多账号池，自动切换'
        },
        control: {
          feature: '用量控制',
          official: '无法限制',
          us: '可设配额、查明细'
        }
      }
    },
    providers: {
      title: '精选核心模型',
      description: '只保留最常用的六大模型系列，一个密钥统一调用，选择更清晰。',
      supported: '已支持',
      soon: '即将推出',
      claude: 'Claude',
      gemini: 'Gemini',
      antigravity: 'Antigravity',
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
        qwen: '通义千问',
        doubao: '豆包',
        zhipu: 'GLM',
        moonshot: 'Kimi / Moonshot',
        baidu: '文心 ERNIE',
        tencent: '腾讯混元',
        iflytek: '讯飞星火',
        minimax: 'MiniMax',
        ai360: '360 智脑'
      }
    },
    domestic: {
      title: '国内厂商与国产模型已加入展示',
      description: '重点展示 DeepSeek、通义千问、豆包、Kimi、智谱、文心、混元、星火等国内模型生态。'
    },
    highlights: {
      gateway: {
        title: '统一网关',
        desc: '兼容 OpenAI 格式，将不同上游模型收敛到一套调用入口。'
      },
      providers: {
        title: '多供应商',
        desc: '参考 new-api 的多渠道理念，统一管理国际模型与国内模型供应商。'
      },
      billing: {
        title: '额度可控',
        desc: '面向 API Key、分组和模型倍率做清晰的用量与费用控制。'
      },
      languages: {
        title: '多语言',
        desc: '提供简体中文、繁体中文、日文、阿拉伯文，并保留英文兜底。'
      }
    },
    notice: {
      title: '系统公告',
      open: '打开系统公告',
      tabNotice: '通知',
      tabSystem: '系统公告',
      important: '（重要）',
      qqGroupLabel: '交流群：',
      qqGroupSuffix: '，有问题可以进群联系站长。',
      trust: '请认准 ISACAI 官方站点与当前 API 入口，避免使用不明来源链接。',
      status: '当前服务入口：',
      closeToday: '今日关闭',
      close: '关闭公告'
    },
    pricing: {
      eyebrow: '价格说明',
      title: '充值与 token 价格这样计算',
      description: '分组倍率为 1 时，余额按官网美元基准价正常扣费；人民币实付再按充值到账倍率换算，输入、输出价格均按每百万 token 展示。',
      fullPricingAction: '查看完整模型定价',
      rechargeLabel: '充值汇率',
      rechargeValue: '1 元 = {usd} USD',
      rechargeHint: '充值到账倍率由管理员配置，保存后同步用于实际到账、首页、充值页和模型价格页。',
      tokenLabel: '余额计费倍率',
      tokenValue: '分组倍率 × 1',
      tokenHint: '余额按模型官网美元基准价扣除；人民币实付 = 美元余额价 ÷ 充值到账倍率。',
      comparisonTitle: '主流模型输入 / 输出价格对比',
      comparisonDescription: '人民币价格统一按分组倍率 ×1 和当前充值到账倍率折算，均按每百万 token 展示。',
      effectiveRateBadge: '按 1 元 = {usd} USD',
      modelHeader: '模型',
      benchmarkInputHeader: '基准输入 USD / 百万',
      benchmarkOutputHeader: '基准输出 USD / 百万',
      effectiveInputHeader: '实付输入 RMB / 百万',
      effectiveOutputHeader: '实付输出 RMB / 百万',
      formula: '换算公式：分组倍率 1 时，官网基准 USD ÷ {usd} = 人民币实付。',
      disclaimer: '表格仅展示 token 基础价格；缓存、图片、分组倍率、峰值倍率及特殊渠道规则以实际账单为准。'
    },
    // CTA 区块
    cta: {
      title: '准备好开始了吗？',
      description: '注册即可获得免费试用额度，体验一站式 AI 服务',
      button: '免费注册'
    },
    footer: {
      allRightsReserved: '保留所有权利。'
    }
  },

  // Key Usage Query Page
  keyUsage: {
    title: 'API Key 用量查询',
    subtitle: '输入您的 API Key 以查看实时消费金额与使用状态',
    placeholder: 'sk-ant-mirror-xxxxxxxxxxxx',
    query: '查询',
    querying: '查询中...',
    privacyNote: '您的 Key 仅在浏览器本地处理，不会被存储',
    dateRange: '统计范围:',
    dateRangeToday: '今日',
    dateRange7d: '7 天',
    dateRange30d: '30 天',
    dateRange90d: '90 天',
    dateRangeCustom: '自定义',
    apply: '应用',
    used: '已使用',
    detailInfo: '详细信息',
    tokenStats: 'Token 统计',
    dailyDetail: '按日明细',
    modelStats: '模型用量统计',
    // Table headers
    date: '日期',
    model: '模型',
    requests: '请求数',
    inputTokens: '输入 Tokens',
    outputTokens: '输出 Tokens',
    cacheCreationTokens: '缓存创建',
    cacheReadTokens: '缓存读取',
    cacheWriteTokens: '缓存写入',
    totalTokens: '总 Tokens',
    cost: '费用',
    // Status
    quotaMode: 'Key 限额模式',
    walletBalance: '钱包余额',
    // Ring card titles
    totalQuota: '总额度',
    limit5h: '5 小时限额',
    limitDaily: '日限额',
    limit7d: '7 天限额',
    limitWeekly: '周限额',
    limitMonthly: '月限额',
    // Detail rows
    remainingQuota: '剩余额度',
    expiresAt: '过期时间',
    todayExpires: '(今日到期)',
    daysLeft: '({days} 天)',
    usedQuota: '已用额度',
    resetNow: '即将重置',
    subscriptionType: '订阅类型',
    subscriptionExpires: '订阅到期',
    // Usage stat cells
    todayRequests: '今日请求',
    todayInputTokens: '今日输入',
    todayOutputTokens: '今日输出',
    todayTokens: '今日 Tokens',
    todayCacheCreation: '今日缓存创建',
    todayCacheRead: '今日缓存读取',
    todayCost: '今日费用',
    rpmTpm: 'RPM / TPM',
    totalRequests: '累计请求',
    totalInputTokens: '累计输入',
    totalOutputTokens: '累计输出',
    totalTokensLabel: '累计 Tokens',
    totalCacheCreation: '累计缓存创建',
    totalCacheRead: '累计缓存读取',
    totalCost: '累计费用',
    avgDuration: '平均耗时',
    // Messages
    enterApiKey: '请输入 API Key',
    querySuccess: '查询成功',
    queryFailed: '查询失败',
    queryFailedRetry: '查询失败，请稍后重试',
    noDailyUsage: '暂无按日用量数据',
  },

  // Setup Wizard
  setup: {
    title: 'Sub2API 安装向导',
    description: '配置您的 Sub2API 实例',
    database: {
      title: '数据库配置',
      description: '连接到您的 PostgreSQL 数据库',
      host: '主机',
      port: '端口',
      username: '用户名',
      password: '密码',
      databaseName: '数据库名称',
      sslMode: 'SSL 模式',
      passwordPlaceholder: '密码',
      ssl: {
        disable: '禁用',
        require: '要求',
        verifyCa: '验证 CA',
        verifyFull: '完全验证'
      }
    },
    redis: {
      title: 'Redis 配置',
      description: '连接到您的 Redis 服务器',
      host: '主机',
      port: '端口',
      username: '用户名（可选）',
      password: '密码（可选）',
      database: '数据库',
      usernamePlaceholder: '默认用户留空',
      passwordPlaceholder: '密码',
      enableTls: '启用 TLS',
      enableTlsHint: '连接 Redis 时使用 TLS（公共 CA 证书）'
    },
    admin: {
      title: '管理员账户',
      description: '创建您的管理员账户',
      email: '邮箱',
      password: '密码',
      confirmPassword: '确认密码',
      passwordPlaceholder: '至少 8 个字符',
      confirmPasswordPlaceholder: '确认密码',
      passwordMismatch: '密码不匹配'
    },
    ready: {
      title: '准备安装',
      description: '检查您的配置并完成安装',
      database: '数据库',
      redis: 'Redis',
      adminEmail: '管理员邮箱'
    },
    status: {
      testing: '测试中...',
      success: '连接成功',
      testConnection: '测试连接',
      installing: '安装中...',
      completeInstallation: '完成安装',
      completed: '安装完成！',
      redirecting: '正在跳转到登录页面...',
      restarting: '服务正在重启，请稍候...',
      timeout: '服务重启时间超出预期，请手动刷新页面。'
    }
  },

  // Common
}
