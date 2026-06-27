import en from './en'

export default {
  ...en,
  home: {
    ...en.home,
    viewOnGithub: 'GitHub で見る',
    viewDocs: 'ドキュメントを見る',
    docs: 'ドキュメント',
    switchToLight: 'ライトモードに切り替え',
    switchToDark: 'ダークモードに切り替え',
    dashboard: 'ダッシュボード',
    login: 'ログイン',
    getStarted: '今すぐ始める',
    goToDashboard: 'ダッシュボードへ',
    nav: {
      tagline: '統合 AI API ゲートウェイ'
    },
    heroEyebrow: 'OpenAI 互換 · 複数モデルの統合ルーティング',
    heroSubtitle: '1 つのキーで、複数の AI モデルへ',
    heroDescription:
      'Claude、GPT、Gemini、DeepSeek、Qwen、Doubao、Kimi などの主要 AI サービスを、1 つの API キーから利用できます。',
    stats: {
      providers: 'モデル提供元',
      languages: '対応言語',
      compatible: '互換プロトコル'
    },
    apiCard: {
      protocol: 'OpenAI Compatible',
      title: '統合 API にすぐ接続',
      subtitle: 'エンドポイントをコピーして、既存の OpenAI SDK から呼び出せます。',
      endpoint: 'エンドポイント',
      copy: 'コピー',
      copied: 'コピー済み'
    },
    tags: {
      subscriptionToApi: '統合 API 接続',
      stickySession: 'セッション維持',
      realtimeBilling: '従量課金'
    },
    providers: {
      ...en.home.providers,
      title: '多くの大規模モデル提供元に対応',
      description: '海外の主要モデル、中国国内ベンダー、国産モデルを 1 つの API で利用できます。',
      supported: '対応済み',
      soon: '近日対応',
      more: 'さらに追加',
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
        qwen: 'Qwen',
        doubao: 'Doubao',
        zhipu: 'Zhipu GLM',
        moonshot: 'Kimi / Moonshot',
        baidu: 'ERNIE',
        tencent: 'Tencent Hunyuan',
        iflytek: 'iFlytek Spark',
        minimax: 'MiniMax',
        ai360: '360 AI'
      }
    },
    domestic: {
      title: '中国国内ベンダーと国産モデルも掲載',
      description: 'DeepSeek、Qwen、Doubao、Kimi、Zhipu、ERNIE、Hunyuan、Spark などを重点的に表示します。'
    },
    highlights: {
      gateway: {
        title: '統合ゲートウェイ',
        desc: 'OpenAI 互換形式のリクエストを、複数の上流モデル提供元へルーティングします。'
      },
      providers: {
        title: '複数プロバイダー',
        desc: 'new-api のマルチチャネル思想を参考に、海外モデルと中国国内モデルをまとめて管理します。'
      },
      billing: {
        title: '利用枠を管理',
        desc: 'API キー、グループ、モデル倍率ごとに利用量とコストを把握できます。'
      },
      languages: {
        title: '多言語対応',
        desc: '簡体字中国語、繁体字中国語、日本語、アラビア語に対応し、英語をフォールバックとして保持します。'
      }
    },
    notice: {
      title: 'システム公告',
      tabNotice: '通知',
      tabSystem: 'システム',
      important: '（重要）',
      qqGroupLabel: 'QQ グループ：',
      qqGroupSuffix: '。問題があればグループで管理者に連絡できます。',
      trust: 'ISACAI 公式サイトと現在の API エントリを利用し、不明なリンクは避けてください。',
      status: '現在のサービス入口：',
      closeToday: '今日は閉じる',
      close: '公告を閉じる'
    },
    footer: {
      allRightsReserved: 'All rights reserved.'
    }
  }
}
