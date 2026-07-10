import en from './en'

type LocaleMessages = Record<string, any>

const base = en as LocaleMessages

export default {
  ...en,
  nav: {
    ...en.nav,
    ccSwitchDownload: 'CC-Switch をダウンロード',
    ccSwitchGuide: 'CC-Switch ガイド'
  },
  keys: {
    ...en.keys,
    useKeyModal: {
      ...en.keys.useKeyModal,
      oneClick: {
        ...(base.keys.useKeyModal.oneClick ?? {}),
        hint:
          '以下の 1 つのコマンドをコピーしてターミナルに貼り付け、Enter を押してください。クライアント設定、シェル起動ファイル、VS Code のターミナル環境設定を書き込みます。',
        runHint:
          '反映するには、クライアント（Claude Code / Codex / Gemini CLI / OpenCode）と VS Code のターミナルを再起動してください。'
      }
    },
    ccsFallback: {
      ...(base.keys.ccsFallback ?? {}),
      downloadTitle: 'CC-Switch をダウンロード',
      downloadDescription:
        '公式チャネルのみを使用してください。Windows ARM64、ポータブル版、チェックサムファイルは公式ダウンロードページで該当アセットを選択してください。',
      windowsDownload: 'Windows MSI',
      macosDownload: 'macOS DMG',
      releasePage: '公式ダウンロードページ',
      macosHint:
        'macOS: DMG をインストールしたら CC-Switch を一度手動で開き、Gatekeeper に止められた場合はシステム設定で許可してください。その後この画面で「もう一度開く」を押します。',
      retryOpen: 'もう一度開く',
      retryOpenDesc: 'ccswitch:// インポートを再試行',
      guide: '完全ガイド',
      guideDesc: 'ローカル CC-Switch と Remote SSH の設定'
    }
  },
  admin: {
    ...en.admin,
    accounts: {
      ...en.admin.accounts,
      oauth: {
        ...en.admin.accounts.oauth,
        gemini: {
          ...en.admin.accounts.oauth.gemini,
          projectIdLabel: 'Project ID（必須）',
          projectIdHint:
            'Code Assist と Google One OAuth では必須です。Google Cloud プロジェクトを作成または選択し、Gemini for Google Cloud API を有効化してから Project ID を入力してください。',
          missingProjectId:
            'Code Assist / Google One OAuth には Project ID が必要です。Google Cloud プロジェクトを作成または選択し、Gemini for Google Cloud API を有効化し、この Google アカウントにアクセス権があることを確認してから認可 URL を生成してください。',
          needsProjectId: '組み込み OAuth（Code Assist / Google One）',
          needsProjectIdDesc: 'Gemini for Google Cloud API が有効な Project ID が必要です'
        }
      },
      gemini: {
        ...en.admin.accounts.gemini,
        oauthType: {
          ...en.admin.accounts.gemini.oauthType,
          googleOneDesc: '個人の Google アカウント。Google Cloud Project ID が必要です',
          codeAssistDesc: 'エンタープライズ向け。Google Cloud Project ID が必要です',
          badges: {
            ...en.admin.accounts.gemini.oauthType.badges,
            noGcp: 'Project ID が必要'
          }
        }
      }
    }
  },
  ccSwitchGuide: {
    heroBadge: 'ローカルとリモートの設定',
    title: 'CC-Switch 設定ガイド',
    description:
      'API キーをインポートする前に、実行場所を確認してください。CLI が現在の PC で動くならローカル CC-Switch を使います。Codex や Claude Code が VS Code Remote SSH のサーバーで動くなら、リモート端末で一括設定コマンドを実行します。',
    actions: {
      openKeys: 'API キーを開く',
      releasePage: '公式リリース'
    },
    remoteWarning: {
      title: 'Remote SSH はローカルデスクトップとは別環境です',
      body:
        'ノート PC に CC-Switch を入れても、VS Code Remote SSH サーバー内の ~/.claude や ~/.codex は更新されません。リモートの Codex / Claude Code では、サーバーに接続された VS Code 端末で API キーモーダルのインストールコマンドを実行してください。'
    },
    scenarios: {
      local: {
        badge: 'ローカルデスクトップ',
        title: 'この PC で CC-Switch を使う',
        body:
          'Claude Code、Codex、Gemini CLI が Windows または macOS のローカル環境で動いている場合の手順です。ブラウザ、CC-Switch、設定ファイルが同じマシン上にあります。',
        points: [
          'Windows / macOS のローカル開発に適しています。',
          'API キーページからインポートし、CC-Switch に設定を受け渡します。',
          'インポート後、ローカル CLI またはエディタ端末を再起動します。'
        ]
      },
      remote: {
        badge: 'Remote SSH',
        title: 'VS Code Remote SSH の端末を設定する',
        body:
          'VS Code が SSH でサーバーに接続しており、Codex や Claude Code がそのサーバー上で実行される場合の手順です。',
        points: [
          'ローカルの CC-Switch デスクトップアプリでリモートファイルを変更しようとしないでください。',
          'キーのモーダルから一括インストールコマンドをコピーし、VS Code のリモート端末に貼り付けます。',
          'コマンドはリモートホスト上のクライアント設定、シェル起動変数、VS Code 端末環境を更新します。'
        ]
      }
    },
    download: {
      title: 'CC-Switch 公式ダウンロード',
      body:
        '必ず公式ソースを使用してください。Windows ARM64、ポータブル版、チェックサムはリリースページから選択できます。',
      officialSite: '公式サイト',
      windows: 'Windows MSI',
      macos: 'macOS DMG',
      release: 'リリースページ'
    },
    local: {
      eyebrow: 'パス A',
      title: 'ローカル CC-Switch 設定',
      body:
        'AI コーディングクライアントが現在の PC で動いている場合は、この方法が最も簡単です。',
      steps: [
        {
          title: 'CC-Switch をインストールする',
          body:
            '公式サイトまたはリリースページから CC-Switch をダウンロードします。Windows は MSI、macOS は DMG をインストールし、一度起動して ccswitch:// プロトコルを登録します。',
          bullets: [
            'ブラウザからインポートする間は CC-Switch を起動しておきます。',
            'ブラウザが CC-Switch を開く許可を求めたら許可します。',
            'ARM64 やポータブル版が必要な場合はリリースページを使います。'
          ]
        },
        {
          title: 'API キーページからインポートする',
          body:
            'API キーを開き、使いたいキーの CC-Switch インポートを実行します。利用するツールに合わせてクライアント種別を選びます。',
          bullets: [
            'Claude Code は Anthropic 互換エンドポイントを使います。',
            'Codex は OpenAI 互換エンドポイントとモデルマッピングを使います。',
            'Gemini CLI は Gemini 互換エンドポイントを使います。'
          ]
        },
        {
          title: 'ローカルで有効化して確認する',
          body:
            'CC-Switch でインポートした Provider が有効なプロファイルになっていることを確認し、Claude Code、Codex、Gemini CLI、またはエディタ端末を再起動します。',
          bullets: [
            'まず短いリクエストでキーとエンドポイントを確認します。',
            'インポートで CC-Switch が開かない場合は再インストールするか、インストールコマンドを使います。',
            'ローカル CC-Switch が更新するのはローカルの設定だけです。'
          ]
        }
      ]
    },
    remote: {
      eyebrow: 'パス B',
      title: 'Remote SSH / VS Code / Codex / Claude Code 設定',
      body:
        'コマンドが実際にサーバー上で動く場合はこちらを使います。ブラウザはローカル PC のままで構いませんが、生成されたコマンドはリモート端末に貼り付けます。',
      steps: [
        {
          title: 'VS Code でサーバーに接続する',
          body:
            'VS Code の Remote SSH でサーバーへ接続し、端末がリモートホスト上で動いていることを確認します。',
          bullets: [
            'VS Code ステータスバーの SSH ホスト名を確認します。',
            'pwd、hostname、whoami などでローカル端末ではないことを確認します。',
            'Codex / Claude Code が未導入なら、リモートサーバー側にインストールします。'
          ]
        },
        {
          title: 'このサイトからインストールコマンドをコピーする',
          body:
            'Web ダッシュボードで API キーを開き、対象キーの Install Command を選び、Codex、Claude Code、Gemini CLI、OpenCode のタブを選択します。',
          bullets: [
            'ノート PC ではなく、リモートサーバーの OS に合うタブを選びます。',
            '多くの SSH サーバーでは Linux/macOS コマンドを使います。',
            'コマンドには API キーが含まれるため、秘密情報として扱います。'
          ]
        },
        {
          title: 'リモート端末に貼り付けて実行する',
          body:
            'コピーしたコマンドを VS Code Remote SSH の端末で実行します。リモートユーザーのクライアント設定と永続的な環境変数が書き込まれます。',
          bullets: [
            'Claude Code は ~/.claude/settings.json と ANTHROPIC_* 変数を書き込みます。',
            'Codex は ~/.codex/config.toml、~/.codex/auth.json、OPENAI_* 変数を書き込みます。',
            'Node.js がある場合、VS Code terminal env もリモート User settings.json にマージされます。'
          ]
        },
        {
          title: '再起動してテストする',
          body:
            '既存のリモート端末を閉じ、新しい VS Code リモート端末を開いてから Codex または Claude Code を起動します。',
          bullets: [
            'profile 更新後の環境変数は新しいシェルで有効になります。',
            'VS Code 端末変数が反映されない場合は Remote SSH ウィンドウを開き直します。',
            '長い作業の前に短いプロンプトで確認します。'
          ]
        }
      ],
      filesTitle: 'コマンドが更新するファイルと変数',
      verifyTitle: '確認用コマンド',
      files: [
        'Claude Code: ~/.claude/settings.json と ANTHROPIC_BASE_URL / ANTHROPIC_API_KEY。',
        'Codex: ~/.codex/config.toml、~/.codex/auth.json、OPENAI_BASE_URL、OPENAI_API_BASE、OPENAI_API_KEY。',
        'Shell profile: ~/.bashrc、~/.bash_profile、~/.zshrc、~/.profile、または PowerShell profile。',
        'VS Code terminal env: terminal.integrated.env.linux / osx / windows。'
      ]
    },
    troubleshooting: {
      title: 'よくある確認点',
      items: [
        {
          title: 'ブラウザから CC-Switch が開かない',
          body:
            'CC-Switch を一度手動で起動してから再試行します。プロトコルが登録されない場合は公式リリースから再インストールしてください。'
        },
        {
          title: 'リモート Codex が古いエンドポイントを使う',
          body:
            'インストールコマンドをローカル端末ではなく Remote SSH 端末で実行したか確認し、完了後にリモート端末を開き直します。'
        },
        {
          title: 'Claude Code がキーを見つけられない',
          body:
            '~/.claude/settings.json を確認し、新しい端末セッションで ANTHROPIC_AUTH_TOKEN または ANTHROPIC_API_KEY が存在するか確認します。'
        },
        {
          title: 'VS Code 端末の環境変数がない',
          body:
            'リモートホストに Node.js を入れてコマンドを再実行するか、terminal.integrated.env.linux、osx、windows に手動で変数を追加します。'
        }
      ]
    }
  },
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
