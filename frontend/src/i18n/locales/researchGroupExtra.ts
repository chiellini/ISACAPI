export const zhHantResearchGroup = {
  title: '課題組', description: '主帳號統一管理餘額，成員保留獨立登入帳號和 API 金鑰。',
  owner: { description: '使用您的餘額資助成員、設定月額度，並僅查看由您付款的呼叫。' },
  status: { active: '啟用', paused: '已暫停', pending: '待確認' },
  actions: { pauseGroup: '暫停資助', resumeGroup: '恢復資助', dissolve: '解散課題組', pause: '暫停', resume: '恢復', remove: '移除', leave: '退出課題組' },
  fields: { groupName: '課題組名稱', memberEmail: '已註冊使用者的準確信箱', monthlyLimit: '月額度（美元）' },
  stats: { sharedBalance: '共享餘額', monthFunding: '本月已資助', activeMembers: '啟用成員', fundedRequests: '資助請求數' },
  invite: { title: '邀請成員', description: '只能邀請已註冊且啟用的普通使用者；成員需在自己的帳號中確認。', send: '傳送邀請', zeroLimitHint: '月額度設為 $0 表示不使用課題組資助，呼叫將使用成員個人餘額。' },
  members: { title: '成員管理', description: '調整額度不會清除本月已產生的用量。', empty: '暫無成員。', member: '成員', usage: '本月用量', resetAt: '重設時間', reserved: '待完成工作已預留 {amount}' },
  usage: { title: '課題組資助用量', description: '此處只顯示實際從主帳號餘額付款的成員請求。', allMembers: '全部成員', startDate: '開始日期', endDate: '結束日期', empty: '沒有符合篩選條件的資助請求。', model: '模型', requestId: '請求 ID', cost: '實付費用', time: '時間' },
  member: { owner: '資助人：{owner}', limit: '月額度', used: '已用', reserved: '已預留', remaining: '剩餘', billingTitle: '呼叫如何扣費', activeBilling: '符合條件的按量呼叫優先使用課題組資助；主帳號餘額或月額度不可用時，整筆呼叫改用您的個人餘額。', pausedBilling: '課題組資助已暫停，目前按量呼叫使用您的個人餘額。', zeroLimitBilling: '您的課題組額度為 $0，按量呼叫使用您的個人餘額。', resetsAt: '月額度將於 {date} 重設。' },
  invitations: { title: '課題組邀請', description: '接受後，符合條件的按量呼叫會優先由課題組付款；您的帳號和 API 金鑰仍然獨立。', limit: '每月資助額度：{amount}', accept: '接受', reject: '拒絕', accepted: '已接受課題組邀請。', rejected: '已拒絕課題組邀請。' },
  create: { title: '建立課題組', description: '您的餘額將作為共享資助來源；成員仍是獨立使用者，並自行管理 API 金鑰。', namePlaceholder: '例如：計算生物學課題組', action: '建立課題組' },
  dashboard: { invitationTitle: '收到課題組邀請', invitationDescription: '{owner} 邀請您加入 {group}。', groupFunding: '課題組月額度', monthlyRemaining: '本月剩餘', usedOfLimit: '已用 {used} / 共 {limit}', resetsAt: '{date} 重設', personalBalance: '個人餘額', fundingPriority: '優先使用課題組額度；不可用時，整筆呼叫改用您的個人餘額。', pausedFallback: '課題組資助已暫停，目前呼叫使用您的個人餘額。', viewDetails: '查看課題組詳情' },
  messages: { created: '課題組已建立。', updated: '課題組資訊已更新。', paused: '課題組資助已暫停。', resumed: '課題組資助已恢復。', invited: '邀請已傳送。', limitUpdated: '月額度已更新。', memberPaused: '成員資助已暫停。', memberResumed: '成員資助已恢復。', dissolved: '課題組已解散。', left: '您已退出課題組。', usageReset: '本月用量已重設。', removed: '成員已移除。' },
  confirm: { dissolveTitle: '解散課題組？', dissolveMessage: '所有成員將立即恢復個人餘額扣費，此操作無法復原。', leaveTitle: '退出課題組？', leaveMessage: '後續呼叫將使用您的個人餘額，帳號和 API 金鑰不會被刪除。', resetTitle: '重設本月用量？', resetMessage: '確認重設 {member} 本月的課題組資助用量？此操作立即生效。', removeTitle: '移除成員？', removeMessage: '確認將 {member} 移出課題組？其後續呼叫將使用個人餘額。' },
  errors: { operationFailed: '課題組操作失敗。', memberIneligible: '無法加入該帳號，請核對準確信箱及帳號狀態。', alreadyAssigned: '該帳號已屬於其他課題組或已有待確認邀請。', forbidden: '您沒有權限管理此課題組。', bothBalancesInsufficient: '課題組餘額和個人餘額均不足。' }
}

export const jaResearchGroup = {
  title: '研究グループ', description: '代表アカウントが残高を管理し、メンバーは個別のログインと API キーを維持します。',
  owner: { description: '残高からメンバーを支援し、月額上限を設定して、実際に支払った利用のみ確認できます。' },
  status: { active: '有効', paused: '一時停止', pending: '承認待ち' },
  actions: { pauseGroup: '支援を停止', resumeGroup: '支援を再開', dissolve: 'グループを解散', pause: '停止', resume: '再開', remove: '削除', leave: 'グループを退出' },
  fields: { groupName: 'グループ名', memberEmail: '登録済みの正確なメールアドレス', monthlyLimit: '月額上限（USD）' },
  stats: { sharedBalance: '共有残高', monthFunding: '今月の支援額', activeMembers: '有効なメンバー', fundedRequests: '支援リクエスト数' },
  invite: { title: 'メンバーを招待', description: '有効な登録ユーザーのみ招待できます。本人のアカウントで承認が必要です。', send: '招待を送信', zeroLimitHint: '$0 を設定するとグループ支援は無効になり、個人残高が使用されます。' },
  members: { title: 'メンバー管理', description: '上限を変更しても今月の利用額はリセットされません。', empty: 'メンバーはいません。', member: 'メンバー', usage: '今月の利用額', resetAt: 'リセット日時', reserved: '処理中ジョブが {amount} を予約中' },
  usage: { title: 'グループ支援の利用', description: '代表アカウントの残高で実際に支払ったリクエストだけを表示します。', allMembers: 'すべてのメンバー', startDate: '開始日', endDate: '終了日', empty: '条件に一致する支援リクエストはありません。', model: 'モデル', requestId: 'リクエスト ID', cost: '支払額', time: '日時' },
  member: { owner: '支援者：{owner}', limit: '月額上限', used: '使用済み', reserved: '予約済み', remaining: '残り', billingTitle: '課金の優先順位', activeBilling: '対象の従量課金はグループ支援を優先します。共有残高または月額枠が使えない場合、そのリクエスト全額を個人残高から支払います。', pausedBilling: 'グループ支援は停止中です。現在は個人残高が使用されます。', zeroLimitBilling: 'グループ枠が $0 のため、個人残高が使用されます。', resetsAt: '月額枠は {date} にリセットされます。' },
  invitations: { title: '研究グループへの招待', description: '承認後、対象の従量課金はグループが優先して支払います。アカウントと API キーは引き続き独立しています。', limit: '月額支援上限：{amount}', accept: '承認', reject: '拒否', accepted: '招待を承認しました。', rejected: '招待を拒否しました。' },
  create: { title: '研究グループを作成', description: 'あなたの残高を共有支援元にします。メンバーは個別ユーザーのまま API キーを管理します。', namePlaceholder: '例：計算生物学研究室', action: 'グループを作成' },
  dashboard: { invitationTitle: '研究グループへの招待', invitationDescription: '{owner} が {group} に招待しました。', groupFunding: 'グループ月額枠', monthlyRemaining: '今月の残り', usedOfLimit: '{limit} のうち {used} 使用', resetsAt: '{date} にリセット', personalBalance: '個人残高', fundingPriority: 'グループ枠を優先し、利用できない場合はリクエスト全額を個人残高で支払います。', pausedFallback: 'グループ支援は停止中のため、現在は個人残高が使用されます。', viewDetails: 'グループ詳細を見る' },
  messages: { created: '研究グループを作成しました。', updated: 'グループ情報を更新しました。', paused: 'グループ支援を停止しました。', resumed: 'グループ支援を再開しました。', invited: '招待を送信しました。', limitUpdated: '月額上限を更新しました。', memberPaused: 'メンバー支援を停止しました。', memberResumed: 'メンバー支援を再開しました。', dissolved: '研究グループを解散しました。', left: '研究グループを退出しました。', usageReset: '今月の利用額をリセットしました。', removed: 'メンバーを削除しました。' },
  confirm: { dissolveTitle: '研究グループを解散しますか？', dissolveMessage: '全メンバーは直ちに個人残高での課金に戻ります。この操作は取り消せません。', leaveTitle: '研究グループを退出しますか？', leaveMessage: '今後は個人残高が使用されます。アカウントと API キーは削除されません。', resetTitle: '今月の利用額をリセットしますか？', resetMessage: '{member} の今月の支援利用額をリセットします。', removeTitle: 'メンバーを削除しますか？', removeMessage: '{member} をグループから削除します。今後は個人残高が使用されます。' },
  errors: { operationFailed: '研究グループの操作に失敗しました。', memberIneligible: 'このアカウントは追加できません。メールアドレスと状態を確認してください。', alreadyAssigned: 'このアカウントは既にグループ所属または招待待ちです。', forbidden: 'この研究グループを管理する権限がありません。', bothBalancesInsufficient: 'グループ残高と個人残高の両方が不足しています。' }
}

export const arResearchGroup = {
  title: 'مجموعة البحث', description: 'يدير الحساب الرئيسي الرصيد مع احتفاظ كل عضو بحساب مستقل ومفاتيح API خاصة به.',
  owner: { description: 'موّل الأعضاء من رصيدك وحدد سقفاً شهرياً وراجع فقط الاستخدام الذي دفعته.' },
  status: { active: 'نشط', paused: 'متوقف', pending: 'بانتظار الموافقة' },
  actions: { pauseGroup: 'إيقاف التمويل', resumeGroup: 'استئناف التمويل', dissolve: 'حل المجموعة', pause: 'إيقاف', resume: 'استئناف', remove: 'إزالة', leave: 'مغادرة المجموعة' },
  fields: { groupName: 'اسم المجموعة', memberEmail: 'البريد الدقيق للمستخدم المسجل', monthlyLimit: 'الحد الشهري (USD)' },
  stats: { sharedBalance: 'الرصيد المشترك', monthFunding: 'تمويل هذا الشهر', activeMembers: 'الأعضاء النشطون', fundedRequests: 'الطلبات الممولة' },
  invite: { title: 'دعوة عضو', description: 'يمكن دعوة حساب مسجل ونشط فقط، ويجب على العضو الموافقة من حسابه.', send: 'إرسال الدعوة', zeroLimitHint: 'الحد $0 يعطل تمويل المجموعة ويستخدم الرصيد الشخصي للعضو.' },
  members: { title: 'إدارة الأعضاء', description: 'تغيير الحد لا يصفر الاستخدام المسجل هذا الشهر.', empty: 'لا يوجد أعضاء.', member: 'العضو', usage: 'استخدام الشهر', resetAt: 'موعد التصفير', reserved: 'محجوز {amount} للمهام المعلقة' },
  usage: { title: 'الاستخدام الممول من المجموعة', description: 'تظهر هنا فقط الطلبات المدفوعة فعلياً من رصيد الحساب الرئيسي.', allMembers: 'كل الأعضاء', startDate: 'تاريخ البدء', endDate: 'تاريخ الانتهاء', empty: 'لا توجد طلبات ممولة تطابق المرشحات.', model: 'النموذج', requestId: 'معرف الطلب', cost: 'التكلفة المدفوعة', time: 'الوقت' },
  member: { owner: 'التمويل من {owner}', limit: 'الحد الشهري', used: 'المستخدم', reserved: 'المحجوز', remaining: 'المتبقي', billingTitle: 'كيفية احتساب المكالمات', activeBilling: 'تستخدم مكالمات الرصيد المؤهلة تمويل المجموعة أولاً. عند تعذر الرصيد المشترك أو الحصة الشهرية تُخصم المكالمة كاملة من رصيدك الشخصي.', pausedBilling: 'تمويل المجموعة متوقف، وتستخدم المكالمات حالياً رصيدك الشخصي.', zeroLimitBilling: 'حصة المجموعة $0، لذلك تستخدم المكالمات رصيدك الشخصي.', resetsAt: 'تتجدد الحصة الشهرية في {date}.' },
  invitations: { title: 'دعوة إلى مجموعة بحث', description: 'بعد القبول تدفع المجموعة أولاً للمكالمات المؤهلة، بينما يبقى حسابك ومفاتيح API مستقلين.', limit: 'حد التمويل الشهري: {amount}', accept: 'قبول', reject: 'رفض', accepted: 'تم قبول الدعوة.', rejected: 'تم رفض الدعوة.' },
  create: { title: 'إنشاء مجموعة بحث', description: 'يصبح رصيدك مصدر التمويل المشترك، ويبقى الأعضاء مستخدمين مستقلين يديرون مفاتيحهم.', namePlaceholder: 'مثال: مختبر الأحياء الحاسوبية', action: 'إنشاء المجموعة' },
  dashboard: { invitationTitle: 'دعوة إلى مجموعة بحث', invitationDescription: 'دعاك {owner} إلى {group}.', groupFunding: 'تمويل المجموعة الشهري', monthlyRemaining: 'المتبقي هذا الشهر', usedOfLimit: 'استخدم {used} من {limit}', resetsAt: 'يتجدد في {date}', personalBalance: 'الرصيد الشخصي', fundingPriority: 'يستخدم تمويل المجموعة أولاً، وعند تعذره تُخصم المكالمة كاملة من رصيدك الشخصي.', pausedFallback: 'تمويل المجموعة متوقف، لذلك تستخدم المكالمات رصيدك الشخصي حالياً.', viewDetails: 'عرض تفاصيل المجموعة' },
  messages: { created: 'تم إنشاء مجموعة البحث.', updated: 'تم تحديث بيانات المجموعة.', paused: 'تم إيقاف التمويل.', resumed: 'تم استئناف التمويل.', invited: 'تم إرسال الدعوة.', limitUpdated: 'تم تحديث الحد الشهري.', memberPaused: 'تم إيقاف تمويل العضو.', memberResumed: 'تم استئناف تمويل العضو.', dissolved: 'تم حل مجموعة البحث.', left: 'غادرت مجموعة البحث.', usageReset: 'تم تصفير استخدام هذا الشهر.', removed: 'تمت إزالة العضو.' },
  confirm: { dissolveTitle: 'حل مجموعة البحث؟', dissolveMessage: 'سيعود جميع الأعضاء فوراً إلى الدفع من رصيدهم الشخصي ولا يمكن التراجع.', leaveTitle: 'مغادرة مجموعة البحث؟', leaveMessage: 'ستستخدم المكالمات القادمة رصيدك الشخصي ولن يحذف حسابك أو مفاتيحك.', resetTitle: 'تصفير استخدام الشهر؟', resetMessage: 'تصفير الاستخدام الممول لهذا الشهر للعضو {member}؟', removeTitle: 'إزالة العضو؟', removeMessage: 'إزالة {member} من المجموعة؟ ستستخدم طلباته القادمة رصيده الشخصي.' },
  errors: { operationFailed: 'فشلت عملية مجموعة البحث.', memberIneligible: 'لا يمكن إضافة هذا الحساب. تحقق من البريد وحالة الحساب.', alreadyAssigned: 'الحساب عضو بالفعل في مجموعة أو لديه دعوة معلقة.', forbidden: 'ليس لديك إذن لإدارة مجموعة البحث.', bothBalancesInsufficient: 'رصيد المجموعة والرصيد الشخصي غير كافيين.' }
}
