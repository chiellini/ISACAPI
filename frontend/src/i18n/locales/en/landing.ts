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
  // Home Page
  home: {
    viewOnGithub: 'View on GitHub',
    viewDocs: 'View Documentation',
    docs: 'Docs',
    switchToLight: 'Switch to Light Mode',
    switchToDark: 'Switch to Dark Mode',
    dashboard: 'Dashboard',
    login: 'Login',
    getStarted: 'Get Started',
    goToDashboard: 'Go to Dashboard',
    nav: {
      tagline: 'Unified AI API Gateway'
    },
    heroEyebrow: 'OpenAI-compatible · Multi-model routing',
    // User-focused value proposition
    heroSubtitle: 'One Key, All AI Models',
    heroDescription: 'Access Claude, GPT, Gemini, DeepSeek, Qwen, Doubao, Kimi and more through one unified API key.',
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
    integrations: {
      eyebrow: 'Developer tool integrations',
      title: 'One connection for Codex, Claude Code, and popular IDE workflows',
      description: 'Use the same API key across CLIs, editor terminals, and remote development environments.',
      codexDescription: 'Automatically write the OpenAI-compatible endpoint and model configuration.',
      claudeCodeDescription: 'Configure the Anthropic-compatible endpoint for local and remote environments.',
      geminiDescription: 'Generate the Gemini CLI configuration in one click without juggling keys.',
      ideDescription: 'Works through compatible APIs and the integrated terminals in VS Code, Cursor, Windsurf, and JetBrains.'
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
      password: 'Password (optional)',
      database: 'Database',
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
