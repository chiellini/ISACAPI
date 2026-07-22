export default {
  batchImageGuide: {
    title: 'Batch Image Generation',
    description: 'Submit multiple prompts in one job and download the generated images when complete'
  },
  ccSwitchGuide: {
    heroBadge: 'Local and remote setup',
    title: 'CC-Switch setup guide',
    description:
      'Choose the right setup path before importing an API key. Local CC-Switch is best when the CLI runs on the same computer as your browser. If Codex or Claude Code runs inside a remote SSH server, configure that remote terminal instead.',
    actions: {
      openKeys: 'Open API Keys',
      releasePage: 'Official releases'
    },
    remoteWarning: {
      title: 'Remote SSH is different from local desktop use',
      body:
        'Installing CC-Switch on your laptop cannot write files inside a VS Code Remote SSH server. For remote Codex / Claude Code, open the integrated terminal that is connected to the server and run the install command from the API key modal there.'
    },
    scenarios: {
      local: {
        badge: 'Local desktop',
        title: 'Use CC-Switch on this computer',
        body:
          'Use this path when Claude Code, Codex, or Gemini CLI runs on your Windows or macOS desktop. The browser, CC-Switch, and the CLI config files are all on the same machine.',
        points: [
          'Recommended for daily local development on Windows or macOS.',
          'Click the import button from the API key page and let CC-Switch receive the configuration.',
          'After import, restart the local CLI so it reads the new profile.'
        ]
      },
      remote: {
        badge: 'Remote SSH',
        title: 'Configure VS Code Remote SSH terminals',
        body:
          'Use this path when VS Code is connected to a Linux/macOS/Windows server over SSH and Codex or Claude Code runs on that server.',
        points: [
          'Do not rely on the local CC-Switch desktop app for remote files.',
          'Copy the one-click install command from the key modal and paste it into the remote VS Code terminal.',
          'The command writes client config files, shell startup variables, and VS Code terminal environment settings on the remote host.'
        ]
      }
    },
    download: {
      title: 'Official CC-Switch downloads',
      body:
        'Use official sources only. Windows ARM64, portable builds, and checksum files are available from the release page.',
      officialSite: 'Official site',
      windows: 'Windows MSI',
      macos: 'macOS DMG',
      release: 'Release page'
    },
    local: {
      eyebrow: 'Path A',
      title: 'Local CC-Switch setup',
      body:
        'This is the simplest path when the AI coding client runs on your current computer.',
      steps: [
        {
          title: 'Install CC-Switch',
          body:
            'Download CC-Switch from the official site or release page. Install the Windows MSI or macOS DMG, then launch CC-Switch once so the ccswitch:// protocol handler is registered.',
          bullets: [
            'Keep CC-Switch running while importing from the browser.',
            'If the browser asks for permission to open CC-Switch, allow it.',
            'Use the release page when you need ARM64 or portable assets.'
          ]
        },
        {
          title: 'Import from the API key page',
          body:
            'Open API Keys, choose the key you want to use, and click the CC-Switch import action. Select the client type that matches your tool.',
          bullets: [
            'Claude Code profiles use the Anthropic-compatible endpoint.',
            'Codex profiles use the OpenAI-compatible endpoint and model mapping.',
            'Gemini CLI profiles use the Gemini-compatible endpoint.'
          ]
        },
        {
          title: 'Activate and verify locally',
          body:
            'In CC-Switch, confirm the imported provider is selected as the active profile. Restart Claude Code, Codex, Gemini CLI, or your editor terminal.',
          bullets: [
            'Run a small request first to verify the key and endpoint.',
            'If import does not open CC-Switch, reinstall CC-Switch or use the install command fallback.',
            'Local CC-Switch only updates local client configuration.'
          ]
        }
      ]
    },
    remote: {
      eyebrow: 'Path B',
      title: 'Remote SSH / VS Code / Codex / Claude Code setup',
      body:
        'Use this path when the command actually runs on a server. The browser can stay on your laptop, but the generated command must be pasted into the remote terminal.',
      steps: [
        {
          title: 'Connect to the server in VS Code',
          body:
            'Open VS Code, connect with Remote SSH, and make sure the terminal prompt is running on the remote host.',
          bullets: [
            'Check the VS Code status bar for the SSH host name.',
            'Run pwd, hostname, or whoami to confirm you are not in a local terminal.',
            'Install Codex / Claude Code on the remote server if they are not already available there.'
          ]
        },
        {
          title: 'Copy the install command from this site',
          body:
            'Open API Keys in the web dashboard, click the key, choose Install Command, and select the tab for Codex, Claude Code, Gemini CLI, or OpenCode.',
          bullets: [
            'Choose the OS tab that matches the remote server, not your laptop.',
            'For most SSH servers, use the Linux/macOS command.',
            'The command contains the selected key, so treat it like a secret.'
          ]
        },
        {
          title: 'Paste into the remote terminal',
          body:
            'Run the copied command inside the VS Code Remote SSH terminal. It writes the client config files and persistent shell environment for that remote user.',
          bullets: [
            'Claude Code writes ~/.claude/settings.json and ANTHROPIC_* environment variables.',
            'Codex writes ~/.codex/config.toml, ~/.codex/auth.json, and OPENAI_* environment variables.',
            'VS Code terminal env is merged into the remote User settings.json when Node.js is available.'
          ]
        },
        {
          title: 'Restart and test',
          body:
            'Close existing remote terminals, open a new VS Code terminal, then start Codex or Claude Code again from that terminal.',
          bullets: [
            'Environment variables only appear in new shells after profile files are updated.',
            'If VS Code terminal variables do not update, reopen the Remote SSH window.',
            'Run a small prompt before starting long sessions.'
          ]
        }
      ],
      filesTitle: 'Files and variables touched by the command',
      verifyTitle: 'Quick verification commands',
      files: [
        'Claude Code: ~/.claude/settings.json and ANTHROPIC_BASE_URL / ANTHROPIC_API_KEY.',
        'Codex: ~/.codex/config.toml, ~/.codex/auth.json, OPENAI_BASE_URL, OPENAI_API_BASE, OPENAI_API_KEY.',
        'Shell profiles: ~/.bashrc, ~/.bash_profile, ~/.zshrc, ~/.profile, or PowerShell profile files.',
        'VS Code terminal env: terminal.integrated.env.linux / osx / windows.'
      ]
    },
    troubleshooting: {
      title: 'Common checks',
      items: [
        {
          title: 'CC-Switch does not open from the browser',
          body:
            'Launch CC-Switch manually once, then retry the import. If the protocol handler is still missing, reinstall CC-Switch from the official release page.'
        },
        {
          title: 'Remote Codex still uses the old endpoint',
          body:
            'Make sure you ran the install command in the Remote SSH terminal, not in a local terminal. Reopen the remote terminal after the command finishes.'
        },
        {
          title: 'Claude Code cannot find the key',
          body:
            'Check ~/.claude/settings.json and confirm ANTHROPIC_AUTH_TOKEN or ANTHROPIC_API_KEY exists in a new terminal session.'
        },
        {
          title: 'VS Code terminal variables are missing',
          body:
            'Install Node.js on the remote host and rerun the install command, or manually add the listed variables to VS Code User settings under terminal.integrated.env.linux, osx, or windows.'
        }
      ]
    }
  },
  pricingPage: {
    eyebrow: 'ISACAI model pricing',
    title: 'Clear model prices in CNY',
    description: 'All models use one unified 1× group rate and show only CNY prices per million tokens. The current credit rate is 1 CNY = {rate} USD.',
    navHome: 'Home',
    navPricing: 'Model pricing',
    searchPlaceholder: 'Search models, providers, or families…',
    filters: 'Filters',
    filterDescription: 'Narrow the catalog by model family.',
    reset: 'Reset',
    providers: 'Model families',
    allModels: 'All models',
    gptModels: 'GPT / OpenAI',
    claudeModels: 'Claude / Anthropic',
    visibleModels: '{count} models',
    priceModeLabel: 'Price basis',
    balanceMode: 'CNY price',
    cashMode: 'CNY price',
    unitLabel: 'Billing unit',
    perMillionShort: 'CNY / 1M',
    perThousandShort: 'CNY / 1M',
    sortLabel: 'Sort',
    sortName: 'Name',
    sortInput: 'Input price',
    sortOutput: 'Output price',
    input: 'Input',
    output: 'Output',
    cacheRead: 'Cache read',
    balanceBadge: 'CNY price',
    cashBadge: 'CNY price',
    tokenBilling: 'Usage based',
    groupRate: 'Group ×1',
    openaiCompatible: 'OpenAI compatible',
    copyModel: 'Copy model name',
    copied: 'Copied',
    noResultsTitle: 'No matching models',
    noResultsDescription: 'Try another keyword or clear the current family filter.',
    clearFilters: 'Clear filters',
    rechargeLabel: 'Top-up credit',
    rechargeValue: '1 CNY = {usd} USD',
    rechargeHint: 'The top-up credit rate is configured in admin and drives both actual credits and CNY price conversion.',
    valueLabel: 'Display unit',
    valueValue: 'CNY / 1M tokens',
    valueHint: 'Input, output, and cache-read prices are all shown per million tokens.',
    billingLabel: 'Balance billing',
    billingValue: 'Group rate ×1',
    billingHint: 'Input, output, and cache-read prices all use the unified 1× group rate.',
    formulaTitle: 'How prices are converted',
    formulaBalance: 'This page uses one unified 1× group rate and does not split prices by group.',
    formulaCash: 'CNY price = benchmark USD model price ÷ {rate}, shown per million tokens.',
    formulaValue: 'When the top-up credit rate changes, all CNY prices on this page are recalculated automatically.',
    disclaimer: 'Token base prices at the unified 1× rate only. Priority, long-context, cache-write, image, and special billing rules follow the final bill.',
    ctaTitle: 'Choose a model and start with one API key',
    ctaDescription: 'Deploy through CC-Switch, terminal commands, remote servers, or local clients.',
    ctaPrimary: 'Register and create a key',
    ctaAuthenticated: 'Open dashboard',
    ctaSecondary: 'Back to home',
    modelDescriptions: {
      sol: 'The flagship GPT-5.6 tier for difficult reasoning, coding, and agentic work.',
      terra: 'A balanced GPT-5.6 tier across capability, speed, and cost.',
      luna: 'A lightweight GPT-5.6 tier for high-volume and cost-sensitive work.',
      gpt55: 'A general flagship model for reasoning, code, and long workflows.',
      opus: 'Claude’s high-capability tier for complex analysis, code, and long tasks.',
      sonnet: 'Claude’s balanced tier for everyday coding, agents, and content work.',
      haiku: 'Claude’s fast lightweight tier for high-throughput, low-latency use.'
    }
  },

  // Home Page
  home: {
    viewOnGithub: 'View on GitHub',
    viewDocs: 'View Documentation',
    docs: 'Docs',
    switchToLight: 'Switch to Light Mode',
    switchToDark: 'Switch to Dark Mode',
    dashboard: 'Dashboard',
    login: 'Login',
    register: 'Sign up',
    getStarted: 'Get Started',
    goToDashboard: 'Go to Dashboard',
    nav: {
      home: 'Home',
      dashboard: 'Dashboard',
      tagline: 'Unified AI API Gateway',
      pricing: 'Model pricing',
      teams: 'Teams & Business',
      integrations: 'Integrations',
      openMenu: 'Open menu',
      closeMenu: 'Close menu'
    },
    heroEyebrow: 'OpenAI-compatible · Multi-model routing',
    heroTitle: 'One API key for',
    heroAccent: 'every leading AI model',
    // User-focused value proposition
    heroSubtitle: 'One Key, All AI Models',
    heroDescription: 'Access Claude, GPT, Gemini, DeepSeek, Qwen, Doubao, Kimi and more through one unified API key.',
    supportedAccess: 'Supported deployment options',
    apiDemo: {
      request: 'Request',
      response: 'Response',
      routed: 'Routed successfully'
    },
    heroStats: {
      recharge: 'CNY paid : USD balance',
      groupRate: 'Group rate',
      models: 'Core models',
      deployments: 'Deployment options'
    },
    ccSwitch: {
      badge: 'Recommended setup',
      title: 'Configure your AI development environment in one click with CC-Switch',
      description: 'Choose one API key to configure Codex, Claude Code, Gemini CLI, and OpenCode in one click, then connect through popular IDE terminals or Remote SSH.',
      primaryAction: 'Configure in One Click',
      downloadAction: 'Download CC-Switch',
      guideAction: 'View the local and remote setup guide',
      panelTitle: 'One setup hub for the AI development tools you use',
      oneClick: 'One-click setup',
      ready: 'Ready',
      localSetup: 'Import local client settings automatically',
      remoteSetup: 'Ready-made command for Remote SSH',
      noManual: 'No manual editing of keys, endpoints, or environment variables'
    },
    authPitch: 'Sign in or register to deploy the same API configuration through CC-Switch, terminal commands, remote servers, and local clients.',
    deployment: {
      ccSwitch: {
        title: 'CC-Switch',
        description: 'One-click desktop import'
      },
      terminal: {
        title: 'Terminal',
        description: 'Ready-made CLI command'
      },
      server: {
        title: 'Server',
        description: 'Remote SSH setup'
      },
      client: {
        title: 'Client',
        description: 'Codex / Claude Code'
      }
    },
    integrations: {
      eyebrow: 'Developer tool integrations',
      title: 'One connection for Codex, Claude Code, and popular IDE workflows',
      description: 'Use the same API key across CLIs, editor terminals, and remote development environments.',
      codexDescription: 'Automatically write the OpenAI-compatible endpoint and model configuration.',
      claudeCodeDescription: 'Configure the Anthropic-compatible endpoint for local and remote environments.',
      geminiDescription: 'Generate the Gemini CLI configuration in one click without juggling keys.',
      ideDescription: 'Works through compatible APIs and the integrated terminals in VS Code, Cursor, Windsurf, and JetBrains.'
    },
    teams: {
      eyebrow: 'Research groups · Business teams',
      title: 'One owner account can support an entire research group or business team',
      description: 'A group lead can register and create a team, then manage shared balance and monthly member sponsorship. Members keep independent logins and API keys, and can see their group quota, usage, and personal balance in their own dashboard.',
      primaryAction: 'Register and create a team',
      manageAction: 'Create or manage a group',
      loginAction: 'Sign in to view your team',
      features: {
        sharedBalance: {
          title: 'Shared balance, centrally funded',
          description: 'The owner balance sponsors eligible pay-as-you-go requests only and pays each member request in full. Subscriptions do not use shared balance; insufficient sponsorship or owner balance sends the request to the member balance.'
        },
        independentAccounts: {
          title: 'Independent member accounts',
          description: 'Every member keeps their own login and API keys, so a group lead does not need to collect or control private keys.'
        },
        visibleQuota: {
          title: 'Monthly quotas members can see',
          description: 'Members can view their monthly group limit, usage, and remaining sponsorship while retaining personal top-up and usage access.'
        },
        billingTrace: {
          title: 'Clear usage and payer records',
          description: 'Usage stays attributed to the member who made the call, while billing records the actual payer for group-funded reporting.'
        }
      },
      business: {
        badge: 'Business procurement support',
        title: 'Procurement and delivery support for universities and companies',
        description: 'Designed for research groups, laboratories, and business teams that need centralized purchasing, clear budget boundaries, and ongoing AI integration.',
        invoiceTitle: 'Official business invoices',
        invoiceDescription: 'Invoice details can be confirmed when the transaction parties, service, and invoicing conditions qualify; this is not an unconditional promise to issue an invoice.',
        integrationTitle: 'Compliant software information integration services',
        integrationDescription: 'API access and information integration for existing software, developer tools, and business systems can be scoped to the actual project and applicable requirements.',
        note: 'Invoice availability, invoice items, tax details, service scope, and delivery terms depend on the transaction, applicable rules, and mutual agreement. “Compliant” does not imply universal certification or a legal conclusion for every industry or use case.'
      }
    },
    workflow: {
      eyebrow: 'Quick start',
      title: 'Start calling models in three steps',
      description: 'Go from registration to your first local or server request with one shared setup.',
      steps: {
        account: {
          title: 'Register and create an API key',
          description: 'Open the dashboard and create one API key to reuse across compatible clients.'
        },
        configure: {
          title: 'Choose a setup method',
          description: 'Import with CC-Switch or copy a terminal command for your local machine or remote server.'
        },
        connect: {
          title: 'Start calling models',
          description: 'Use the OpenAI- or Anthropic-compatible endpoint directly from Codex, Claude Code, or your IDE.'
        }
      }
    },
    tags: {
      subscriptionToApi: 'Unified API Access',
      stickySession: 'Session Persistence',
      realtimeBilling: 'Pay As You Go'
    },
    stats: {
      providers: 'Providers',
      languages: 'Languages',
      compatible: 'Protocol'
    },
    apiCard: {
      protocol: 'OpenAI Compatible',
      title: 'Connect to one unified API',
      subtitle: 'Copy the endpoint and keep using your existing OpenAI SDK.',
      endpoint: 'Endpoint',
      copy: 'Copy',
      copied: 'Copied'
    },
    // Pain points section
    painPoints: {
      title: 'Sound Familiar?',
      items: {
        expensive: {
          title: 'High Subscription Costs',
          desc: 'Paying for multiple AI subscriptions that add up every month'
        },
        complex: {
          title: 'Account Chaos',
          desc: 'Managing scattered accounts and API keys across different platforms'
        },
        unstable: {
          title: 'Service Interruptions',
          desc: 'Single accounts hitting rate limits and disrupting your workflow'
        },
        noControl: {
          title: 'No Usage Control',
          desc: "Can't track where your money goes or limit team member usage"
        }
      }
    },
    // Solutions section
    solutions: {
      title: 'We Solve These Problems',
      subtitle: 'Three simple steps to stress-free AI access'
    },
    features: {
      unifiedGateway: 'One-Click Access',
      unifiedGatewayDesc: 'Get a single API key to call all connected AI models. No separate applications needed.',
      multiAccount: 'Always Reliable',
      multiAccountDesc: 'Smart routing across multiple upstream accounts with automatic failover. Say goodbye to errors.',
      balanceQuota: 'Pay What You Use',
      balanceQuotaDesc: 'Usage-based billing with quota limits. Full visibility into team consumption.'
    },
    // Comparison section
    comparison: {
      title: 'Why Choose Us?',
      headers: {
        feature: 'Comparison',
        official: 'Official Subscriptions',
        us: 'Our Platform'
      },
      items: {
        pricing: {
          feature: 'Pricing',
          official: 'Fixed monthly fee, pay even if unused',
          us: 'Pay only for what you use'
        },
        models: {
          feature: 'Model Selection',
          official: 'Single provider only',
          us: 'Switch between models freely'
        },
        management: {
          feature: 'Account Management',
          official: 'Manage each service separately',
          us: 'Unified key, one dashboard'
        },
        stability: {
          feature: 'Stability',
          official: 'Single account rate limits',
          us: 'Multi-account pool, auto-failover'
        },
        control: {
          feature: 'Usage Control',
          official: 'Not available',
          us: 'Quotas & detailed analytics'
        }
      }
    },
    providers: {
      title: 'Curated Core Models',
      description: 'Six widely used model families, one API key, and a much clearer choice.',
      supported: 'Supported',
      soon: 'Soon',
      claude: 'Claude',
      gemini: 'Gemini',
      antigravity: 'Antigravity',
      more: 'More',
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
        qwen: 'Qwen',
        doubao: 'Doubao',
        zhipu: 'GLM',
        moonshot: 'Kimi / Moonshot',
        baidu: 'ERNIE',
        tencent: 'Tencent Hunyuan',
        iflytek: 'iFlytek Spark',
        minimax: 'MiniMax',
        ai360: '360 AI'
      }
    },
    domestic: {
      title: 'Chinese vendors and domestic models are included',
      description: 'Highlights DeepSeek, Qwen, Doubao, Kimi, Zhipu, ERNIE, Hunyuan, Spark and more.'
    },
    highlights: {
      gateway: {
        title: 'Unified Gateway',
        desc: 'OpenAI-compatible requests are routed to multiple upstream model providers.'
      },
      providers: {
        title: 'Multi-provider',
        desc: 'Inspired by new-api style multi-channel management for global and Chinese providers.'
      },
      billing: {
        title: 'Quota Control',
        desc: 'Track usage and cost by API key, group, and model pricing multiplier.'
      },
      languages: {
        title: 'Multilingual',
        desc: 'Simplified Chinese, Traditional Chinese, Japanese, Arabic, with English fallback.'
      }
    },
    notice: {
      title: 'System Notice',
      open: 'Open system notice',
      tabNotice: 'Notice',
      tabSystem: 'System',
      important: '(Important)',
      qqGroupLabel: 'QQ group: ',
      qqGroupSuffix: '. Join the group to contact the site admin.',
      trust: 'Use the official ISACAI site and current API entry to avoid unknown links.',
      status: 'Current service entry: ',
      closeToday: 'Close today',
      close: 'Close notice'
    },
    pricing: {
      eyebrow: 'Pricing',
      title: 'How top-ups and token prices are calculated',
      description: 'At a 1× group rate, balance is charged at the benchmark USD price. Effective CNY is then converted through the credited top-up rate. Input and output are shown per million tokens.',
      fullPricingAction: 'View complete model pricing',
      rechargeLabel: 'Top-up rate',
      rechargeValue: '1 CNY = {usd} USD',
      rechargeHint: 'The top-up credit rate is configured by an administrator. Saving it updates actual credits, the home page, checkout, and model pricing.',
      tokenLabel: 'Balance billing rate',
      tokenValue: 'Group rate × 1',
      tokenHint: 'Balance follows the benchmark USD model price. Effective CNY = USD balance price ÷ credited top-up rate.',
      comparisonTitle: 'Model input / output price comparison',
      comparisonDescription: 'CNY prices use the unified 1× group rate and the current top-up credit rate, all shown per million tokens.',
      effectiveRateBadge: 'At 1 CNY = {usd} USD',
      modelHeader: 'Model',
      benchmarkInputHeader: 'Benchmark input USD / 1M',
      benchmarkOutputHeader: 'Benchmark output USD / 1M',
      effectiveInputHeader: 'Effective input CNY / 1M',
      effectiveOutputHeader: 'Effective output CNY / 1M',
      formula: 'Formula at a 1× group rate: benchmark USD ÷ {usd} = effective CNY paid.',
      disclaimer: 'Token base prices only. Cache, image, group, peak-time, and channel-specific multipliers follow the final bill.'
    },
    // CTA section
    cta: {
      title: 'Ready to Get Started?',
      description: 'Sign up now and get free trial credits to experience seamless AI access',
      button: 'Sign Up Free'
    },
    footer: {
      allRightsReserved: 'All rights reserved.'
    }
  },

  // Key Usage Query Page
  keyUsage: {
    title: 'API Key Usage',
    subtitle: 'Enter your API Key to view real-time spending and usage status',
    placeholder: 'sk-ant-mirror-xxxxxxxxxxxx',
    query: 'Query',
    querying: 'Querying...',
    privacyNote: 'Your Key is processed locally in the browser and will not be stored',
    dateRange: 'Date Range:',
    dateRangeToday: 'Today',
    dateRange7d: '7 Days',
    dateRange30d: '30 Days',
    dateRange90d: '90 Days',
    dateRangeCustom: 'Custom',
    apply: 'Apply',
    used: 'Used',
    detailInfo: 'Detail Information',
    tokenStats: 'Token Statistics',
    dailyDetail: 'Daily Detail',
    modelStats: 'Model Usage Statistics',
    // Table headers
    date: 'Date',
    model: 'Model',
    requests: 'Requests',
    inputTokens: 'Input Tokens',
    outputTokens: 'Output Tokens',
    cacheCreationTokens: 'Cache Creation',
    cacheReadTokens: 'Cache Read',
    cacheWriteTokens: 'Cache Write',
    totalTokens: 'Total Tokens',
    cost: 'Cost',
    // Status
    quotaMode: 'Key Quota Mode',
    walletBalance: 'Wallet Balance',
    // Ring card titles
    totalQuota: 'Total Quota',
    limit5h: '5-Hour Limit',
    limitDaily: 'Daily Limit',
    limit7d: '7-Day Limit',
    limitWeekly: 'Weekly Limit',
    limitMonthly: 'Monthly Limit',
    // Detail rows
    remainingQuota: 'Remaining Quota',
    expiresAt: 'Expires At',
    todayExpires: '(expires today)',
    daysLeft: '({days} days)',
    usedQuota: 'Used Quota',
    resetNow: 'Resetting soon',
    subscriptionType: 'Subscription Type',
    subscriptionExpires: 'Subscription Expires',
    // Usage stat cells
    todayRequests: 'Today Requests',
    todayInputTokens: 'Today Input',
    todayOutputTokens: 'Today Output',
    todayTokens: 'Today Tokens',
    todayCacheCreation: 'Today Cache Creation',
    todayCacheRead: 'Today Cache Read',
    todayCost: 'Today Cost',
    rpmTpm: 'RPM / TPM',
    totalRequests: 'Total Requests',
    totalInputTokens: 'Total Input',
    totalOutputTokens: 'Total Output',
    totalTokensLabel: 'Total Tokens',
    totalCacheCreation: 'Total Cache Creation',
    totalCacheRead: 'Total Cache Read',
    totalCost: 'Total Cost',
    avgDuration: 'Avg Duration',
    // Messages
    enterApiKey: 'Please enter an API Key',
    querySuccess: 'Query successful',
    queryFailed: 'Query failed',
    queryFailedRetry: 'Query failed, please try again later',
    noDailyUsage: 'No daily usage data',
  },

  // Setup Wizard
  setup: {
    title: 'Sub2API Setup',
    description: 'Configure your Sub2API instance',
    database: {
      title: 'Database Configuration',
      description: 'Connect to your PostgreSQL database',
      host: 'Host',
      port: 'Port',
      username: 'Username',
      password: 'Password',
      databaseName: 'Database Name',
      sslMode: 'SSL Mode',
      passwordPlaceholder: 'Password',
      ssl: {
        disable: 'Disable',
        require: 'Require',
        verifyCa: 'Verify CA',
        verifyFull: 'Verify Full'
      }
    },
    redis: {
      title: 'Redis Configuration',
      description: 'Connect to your Redis server',
      host: 'Host',
      port: 'Port',
      username: 'Username (optional)',
      password: 'Password (optional)',
      database: 'Database',
      usernamePlaceholder: 'Leave empty for default user',
      passwordPlaceholder: 'Password',
      enableTls: 'Enable TLS',
      enableTlsHint: 'Use TLS when connecting to Redis (public CA certs)'
    },
    admin: {
      title: 'Admin Account',
      description: 'Create your administrator account',
      email: 'Email',
      password: 'Password',
      confirmPassword: 'Confirm Password',
      passwordPlaceholder: 'Min 8 characters',
      confirmPasswordPlaceholder: 'Confirm password',
      passwordMismatch: 'Passwords do not match'
    },
    ready: {
      title: 'Ready to Install',
      description: 'Review your configuration and complete setup',
      database: 'Database',
      redis: 'Redis',
      adminEmail: 'Admin Email'
    },
    status: {
      testing: 'Testing...',
      success: 'Connection Successful',
      testConnection: 'Test Connection',
      installing: 'Installing...',
      completeInstallation: 'Complete Installation',
      completed: 'Installation completed!',
      redirecting: 'Redirecting to login page...',
      restarting: 'Service is restarting, please wait...',
      timeout: 'Service restart is taking longer than expected. Please refresh the page manually.'
    }
  },

  // Common
}
