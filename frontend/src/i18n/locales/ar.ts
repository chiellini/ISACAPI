import en from './en'

export default {
  ...en,
  nav: {
    ...en.nav,
    ccSwitchDownload: 'تنزيل CC-Switch'
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
      releasePage: 'صفحة التنزيل الرسمية'
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
