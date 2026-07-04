import en from './en'

export default {
  ...en,
  nav: {
    ...en.nav,
    ccSwitchDownload: 'تنزيل CC-Switch',
    ccSwitchGuide: 'دليل CC-Switch'
  },
  keys: {
    ...en.keys,
    useKeyModal: {
      ...en.keys.useKeyModal,
      oneClick: {
        ...en.keys.useKeyModal.oneClick,
        hint:
          'انسخ الأمر الواحد أدناه والصقه في الطرفية ثم اضغط Enter. سيكتب إعدادات العميل وملفات بدء تشغيل الصدفة وإعدادات بيئة طرفية VS Code.',
        runHint:
          'أعد تشغيل العميل (Claude Code / Codex / Gemini CLI / OpenCode) وطرفية VS Code بعد ذلك لتطبيق الإعدادات.'
      }
    },
    ccsFallback: {
      ...en.keys.ccsFallback,
      downloadTitle: 'تنزيل CC-Switch',
      downloadDescription:
        'استخدم القنوات الرسمية فقط. افتح صفحة التنزيل الرسمية لاختيار Windows ARM64 أو النسخة المحمولة أو ملفات التحقق.',
      windowsDownload: 'Windows MSI',
      macosDownload: 'macOS DMG',
      releasePage: 'صفحة التنزيل الرسمية',
      guide: 'الدليل الكامل',
      guideDesc: 'إعداد CC-Switch المحلي و Remote SSH'
    }
  },
  ccSwitchGuide: {
    heroBadge: 'إعداد محلي وعن بعد',
    title: 'دليل إعداد CC-Switch',
    description:
      'اختر المسار الصحيح قبل استيراد مفتاح API. إذا كان Claude Code أو Codex أو Gemini CLI يعمل على جهازك الحالي فاستخدم CC-Switch محليا. إذا كان يعمل داخل خادم متصل عبر VS Code Remote SSH، فشغل أمر الإعداد داخل الطرفية البعيدة.',
    actions: {
      openKeys: 'فتح مفاتيح API',
      releasePage: 'الإصدارات الرسمية'
    },
    remoteWarning: {
      title: 'Remote SSH ليس نفس بيئة سطح المكتب المحلية',
      body:
        'تثبيت CC-Switch على جهازك المحمول لا يكتب ملفات ~/.claude أو ~/.codex داخل خادم VS Code Remote SSH. لإعداد Codex / Claude Code البعيد، افتح طرفية VS Code المتصلة بالخادم وشغل أمر التثبيت من نافذة مفتاح API هناك.'
    },
    scenarios: {
      local: {
        badge: 'سطح المكتب المحلي',
        title: 'استخدام CC-Switch على هذا الجهاز',
        body:
          'هذا المسار مناسب عندما يعمل Claude Code أو Codex أو Gemini CLI على Windows أو macOS المحلي. المتصفح و CC-Switch وملفات إعداد العميل كلها على نفس الجهاز.',
        points: [
          'مناسب للتطوير اليومي على Windows أو macOS.',
          'استورد الإعداد من صفحة مفاتيح API ودع CC-Switch يستقبله.',
          'بعد الاستيراد أعد تشغيل CLI المحلي أو طرفية المحرر لقراءة الإعداد الجديد.'
        ]
      },
      remote: {
        badge: 'Remote SSH',
        title: 'إعداد طرفيات VS Code Remote SSH',
        body:
          'هذا المسار مناسب عندما يكون VS Code متصلا بخادم عبر SSH ويعمل Codex أو Claude Code فعليا على ذلك الخادم.',
        points: [
          'لا تعتمد على تطبيق CC-Switch المحلي لتعديل ملفات الخادم البعيد.',
          'انسخ أمر التثبيت بنقرة واحدة من نافذة المفتاح والصقه في طرفية VS Code البعيدة.',
          'سيكتب الأمر إعدادات العميل ومتغيرات بدء تشغيل الصدفة وبيئة طرفية VS Code على الخادم البعيد.'
        ]
      }
    },
    download: {
      title: 'تنزيلات CC-Switch الرسمية',
      body:
        'استخدم المصادر الرسمية فقط. تتوفر ملفات Windows ARM64 والنسخ المحمولة وملفات التحقق من صفحة الإصدارات.',
      officialSite: 'الموقع الرسمي',
      windows: 'Windows MSI',
      macos: 'macOS DMG',
      release: 'صفحة الإصدارات'
    },
    local: {
      eyebrow: 'المسار A',
      title: 'إعداد CC-Switch المحلي',
      body:
        'هذا هو المسار الأبسط عندما يعمل عميل البرمجة بالذكاء الاصطناعي على جهازك الحالي.',
      steps: [
        {
          title: 'تثبيت CC-Switch',
          body:
            'نزّل CC-Switch من الموقع الرسمي أو صفحة الإصدارات. ثبّت Windows MSI أو macOS DMG، ثم شغّل CC-Switch مرة واحدة لتسجيل بروتوكول ccswitch:// في النظام.',
          bullets: [
            'اترك CC-Switch قيد التشغيل أثناء الاستيراد من المتصفح.',
            'إذا طلب المتصفح إذنا لفتح CC-Switch فاسمح بذلك.',
            'استخدم صفحة الإصدارات عند الحاجة إلى ARM64 أو نسخة محمولة.'
          ]
        },
        {
          title: 'الاستيراد من صفحة مفاتيح API',
          body:
            'افتح مفاتيح API، اختر المفتاح المطلوب، ثم اضغط إجراء استيراد CC-Switch. اختر نوع العميل الذي يطابق أداتك.',
          bullets: [
            'ملفات Claude Code تستخدم نقطة نهاية متوافقة مع Anthropic.',
            'ملفات Codex تستخدم نقطة نهاية متوافقة مع OpenAI وتعيين النموذج.',
            'ملفات Gemini CLI تستخدم نقطة نهاية متوافقة مع Gemini.'
          ]
        },
        {
          title: 'التفعيل والتحقق محليا',
          body:
            'في CC-Switch تأكد أن المزوّد المستورد هو الملف النشط. أعد تشغيل Claude Code أو Codex أو Gemini CLI أو طرفية المحرر.',
          bullets: [
            'أرسل طلبا صغيرا أولا للتحقق من المفتاح ونقطة النهاية.',
            'إذا لم يفتح الاستيراد CC-Switch، أعد تثبيته أو استخدم أمر التثبيت.',
            'CC-Switch المحلي يحدّث إعدادات الجهاز المحلي فقط.'
          ]
        }
      ]
    },
    remote: {
      eyebrow: 'المسار B',
      title: 'إعداد Remote SSH / VS Code / Codex / Claude Code',
      body:
        'استخدم هذا المسار عندما يعمل الأمر فعليا على خادم. يمكن أن يبقى المتصفح على جهازك، لكن الأمر المولّد يجب أن يلصق في الطرفية البعيدة.',
      steps: [
        {
          title: 'الاتصال بالخادم في VS Code',
          body:
            'افتح VS Code واتصل بالخادم باستخدام Remote SSH، ثم تأكد أن الطرفية تعمل على المضيف البعيد.',
          bullets: [
            'تحقق من اسم مضيف SSH في شريط حالة VS Code.',
            'شغّل pwd أو hostname أو whoami للتأكد أنك لست في طرفية محلية.',
            'ثبّت Codex / Claude Code على الخادم البعيد إذا لم تكن مثبتة.'
          ]
        },
        {
          title: 'نسخ أمر التثبيت من الموقع',
          body:
            'افتح مفاتيح API في لوحة الويب، اضغط المفتاح، اختر Install Command، ثم اختر تبويب Codex أو Claude Code أو Gemini CLI أو OpenCode.',
          bullets: [
            'اختر تبويب نظام التشغيل الخاص بالخادم البعيد وليس جهازك المحمول.',
            'معظم خوادم SSH تستخدم أمر Linux/macOS.',
            'الأمر يحتوي على المفتاح المحدد، لذلك عامله كمعلومة سرية.'
          ]
        },
        {
          title: 'اللصق والتنفيذ في الطرفية البعيدة',
          body:
            'شغّل الأمر المنسوخ داخل طرفية VS Code Remote SSH. سيكتب ملفات إعداد العميل ومتغيرات البيئة الدائمة لذلك المستخدم البعيد.',
          bullets: [
            'Claude Code يكتب ~/.claude/settings.json ومتغيرات ANTHROPIC_*.',
            'Codex يكتب ~/.codex/config.toml و ~/.codex/auth.json ومتغيرات OPENAI_*.',
            'عند توفر Node.js يتم دمج VS Code terminal env في remote User settings.json.'
          ]
        },
        {
          title: 'إعادة التشغيل والاختبار',
          body:
            'أغلق الطرفيات البعيدة الحالية، افتح طرفية VS Code جديدة، ثم شغّل Codex أو Claude Code منها.',
          bullets: [
            'متغيرات البيئة تظهر عادة في جلسات الصدفة الجديدة بعد تحديث ملفات profile.',
            'إذا لم تتحدث متغيرات طرفية VS Code، أعد فتح نافذة Remote SSH.',
            'اختبر بموجه صغير قبل بدء جلسات طويلة.'
          ]
        }
      ],
      filesTitle: 'الملفات والمتغيرات التي يلمسها الأمر',
      verifyTitle: 'أوامر تحقق سريعة',
      files: [
        'Claude Code: ‏~/.claude/settings.json و ANTHROPIC_BASE_URL / ANTHROPIC_API_KEY.',
        'Codex: ‏~/.codex/config.toml و ~/.codex/auth.json و OPENAI_BASE_URL و OPENAI_API_BASE و OPENAI_API_KEY.',
        'ملفات الصدفة: ‏~/.bashrc و ~/.bash_profile و ~/.zshrc و ~/.profile أو ملفات PowerShell profile.',
        'بيئة طرفية VS Code: ‏terminal.integrated.env.linux / osx / windows.'
      ]
    },
    troubleshooting: {
      title: 'فحوصات شائعة',
      items: [
        {
          title: 'المتصفح لا يفتح CC-Switch',
          body:
            'شغّل CC-Switch يدويا مرة واحدة ثم أعد الاستيراد. إذا بقي البروتوكول غير مسجل، أعد التثبيت من صفحة الإصدارات الرسمية.'
        },
        {
          title: 'Codex البعيد ما زال يستخدم نقطة النهاية القديمة',
          body:
            'تأكد أنك شغّلت أمر التثبيت في طرفية Remote SSH وليس في طرفية محلية. افتح طرفية بعيدة جديدة بعد انتهاء الأمر.'
        },
        {
          title: 'Claude Code لا يجد المفتاح',
          body:
            'افحص ~/.claude/settings.json وتأكد في جلسة طرفية جديدة من وجود ANTHROPIC_AUTH_TOKEN أو ANTHROPIC_API_KEY.'
        },
        {
          title: 'متغيرات طرفية VS Code غير موجودة',
          body:
            'ثبّت Node.js على المضيف البعيد وأعد تشغيل الأمر، أو أضف المتغيرات يدويا إلى terminal.integrated.env.linux أو osx أو windows في User settings.'
        }
      ]
    }
  },
  home: {
    ...en.home,
    viewOnGithub: 'عرض على GitHub',
    viewDocs: 'عرض الوثائق',
    docs: 'الوثائق',
    switchToLight: 'التبديل إلى الوضع الفاتح',
    switchToDark: 'التبديل إلى الوضع الداكن',
    dashboard: 'لوحة التحكم',
    login: 'تسجيل الدخول',
    getStarted: 'ابدأ الآن',
    goToDashboard: 'الانتقال إلى لوحة التحكم',
    nav: {
      tagline: 'بوابة API موحدة للذكاء الاصطناعي'
    },
    heroEyebrow: 'متوافق مع OpenAI · توجيه موحد لعدة نماذج',
    heroSubtitle: 'مفتاح واحد لكل نماذج الذكاء الاصطناعي',
    heroDescription:
      'استخدم Claude و GPT و Gemini و DeepSeek و Qwen و Doubao و Kimi وغيرها عبر مفتاح API موحد واحد.',
    stats: {
      providers: 'مزودو النماذج',
      languages: 'لغات الموقع',
      compatible: 'البروتوكول'
    },
    apiCard: {
      protocol: 'OpenAI Compatible',
      title: 'اتصل بواجهة API موحدة',
      subtitle: 'انسخ نقطة النهاية واستخدم SDK الحالي المتوافق مع OpenAI.',
      endpoint: 'نقطة النهاية',
      copy: 'نسخ',
      copied: 'تم النسخ'
    },
    tags: {
      subscriptionToApi: 'وصول API موحد',
      stickySession: 'ثبات الجلسة',
      realtimeBilling: 'الدفع حسب الاستخدام'
    },
    providers: {
      ...en.home.providers,
      title: 'دعم واسع لمزودي النماذج الكبيرة',
      description: 'نماذج عالمية، ومزودون صينيون، ونماذج محلية عبر API واحد.',
      supported: 'مدعوم',
      soon: 'قريبا',
      more: 'المزيد',
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
      title: 'تمت إضافة المزودين الصينيين والنماذج المحلية',
      description: 'إبراز DeepSeek و Qwen و Doubao و Kimi و Zhipu و ERNIE و Hunyuan و Spark وغيرها.'
    },
    highlights: {
      gateway: {
        title: 'بوابة موحدة',
        desc: 'تجميع طلبات OpenAI المتوافقة وتوجيهها إلى عدة مزودين للنماذج.'
      },
      providers: {
        title: 'عدة مزودين',
        desc: 'إدارة متعددة القنوات مستوحاة من new-api للمزودين العالميين والصينيين.'
      },
      billing: {
        title: 'تحكم في الحصة',
        desc: 'تتبع الاستخدام والتكلفة حسب مفتاح API والمجموعة ومعامل سعر النموذج.'
      },
      languages: {
        title: 'متعدد اللغات',
        desc: 'يدعم الصينية المبسطة والتقليدية واليابانية والعربية مع الإنجليزية كلغة احتياطية.'
      }
    },
    notice: {
      title: 'إعلان النظام',
      tabNotice: 'إشعار',
      tabSystem: 'النظام',
      important: '(مهم)',
      qqGroupLabel: 'مجموعة QQ: ',
      qqGroupSuffix: '. يمكنك الانضمام للتواصل مع مدير الموقع عند وجود مشكلة.',
      trust: 'استخدم موقع ISACAI الرسمي ومدخل API الحالي وتجنب الروابط غير المعروفة.',
      status: 'مدخل الخدمة الحالي: ',
      closeToday: 'إغلاق اليوم',
      close: 'إغلاق الإعلان'
    },
    footer: {
      allRightsReserved: 'All rights reserved.'
    }
  }
}
