export default {
  researchGroup: {
    title: '课题组',
    description: '主账号统一管理余额，成员保留独立登录账号和 API 密钥。',
    owner: { description: '使用您的余额资助成员，设置月额度，并仅查看由您付款的调用。' },
    status: { active: '启用', paused: '已暂停', pending: '待确认' },
    actions: { pauseGroup: '暂停资助', resumeGroup: '恢复资助', dissolve: '解散课题组', pause: '暂停', resume: '恢复', remove: '移除', leave: '退出课题组' },
    fields: { groupName: '课题组名称', memberEmail: '已注册用户的准确邮箱', monthlyLimit: '月额度（美元）' },
    stats: { sharedBalance: '共享余额', monthFunding: '本月已资助', activeMembers: '启用成员', fundedRequests: '资助请求数' },
    invite: { title: '邀请成员', description: '只能邀请已注册且启用的普通用户；成员需在自己的账号中确认。', send: '发送邀请', zeroLimitHint: '月额度设为 $0 表示不使用课题组资助，调用将使用成员个人余额。' },
    members: { title: '成员管理', description: '调整额度不会清零本月已经产生的用量。', empty: '暂无成员。', member: '成员', usage: '本月用量', resetAt: '重置时间', reserved: '待完成任务已预留 {amount}' },
    usage: { title: '课题组资助用量', description: '这里只显示实际从主账号余额付款的成员请求。', allMembers: '全部成员', startDate: '开始日期', endDate: '结束日期', empty: '没有符合筛选条件的资助请求。', model: '模型', requestId: '请求 ID', cost: '实付费用', time: '时间' },
    member: { owner: '资助人：{owner}', limit: '月额度', used: '已用', reserved: '已预留', remaining: '剩余', billingTitle: '调用如何扣费', activeBilling: '符合条件的按量调用优先使用课题组资助；主账号余额或月额度不可用时，整笔调用改用您的个人余额。', pausedBilling: '课题组资助已暂停，当前按量调用使用您的个人余额。', zeroLimitBilling: '您的课题组额度为 $0，按量调用使用您的个人余额。', resetsAt: '月额度将在 {date} 重置。' },
    invitations: { title: '课题组邀请', description: '接受后，符合条件的按量调用会优先由课题组付款；您的账号和 API 密钥仍然独立。', limit: '每月资助额度：{amount}', accept: '接受', reject: '拒绝', accepted: '已接受课题组邀请。', rejected: '已拒绝课题组邀请。' },
    create: { title: '创建课题组', description: '您的余额将作为共享资助来源；成员仍是独立用户，并自行管理 API 密钥。', namePlaceholder: '例如：计算生物学课题组', action: '创建课题组' },
    dashboard: { invitationTitle: '收到课题组邀请', invitationDescription: '{owner} 邀请您加入 {group}。', groupFunding: '课题组月额度', monthlyRemaining: '本月剩余', usedOfLimit: '已用 {used} / 共 {limit}', resetsAt: '{date} 重置', personalBalance: '个人余额', fundingPriority: '优先使用课题组额度；不可用时，整笔调用改用您的个人余额。', pausedFallback: '课题组资助已暂停，当前调用使用您的个人余额。', viewDetails: '查看课题组详情' },
    messages: { created: '课题组已创建。', updated: '课题组信息已更新。', paused: '课题组资助已暂停。', resumed: '课题组资助已恢复。', invited: '邀请已发送。', limitUpdated: '月额度已更新。', memberPaused: '成员资助已暂停。', memberResumed: '成员资助已恢复。', dissolved: '课题组已解散。', left: '您已退出课题组。', usageReset: '本月用量已重置。', removed: '成员已移除。' },
    confirm: { dissolveTitle: '解散课题组？', dissolveMessage: '所有成员将立即恢复个人余额扣费，此操作不可撤销。', leaveTitle: '退出课题组？', leaveMessage: '后续调用将使用您的个人余额，账号和 API 密钥不会被删除。', resetTitle: '重置本月用量？', resetMessage: '确认重置 {member} 本月的课题组资助用量？此操作立即生效。', removeTitle: '移除成员？', removeMessage: '确认将 {member} 移出课题组？其后续调用将使用个人余额。' },
    errors: { operationFailed: '课题组操作失败。', memberIneligible: '无法添加该账号，请核对准确邮箱及账号状态。', alreadyAssigned: '该账号已属于其他课题组或已有待确认邀请。', forbidden: '您没有权限管理该课题组。', bothBalancesInsufficient: '课题组余额和个人余额均不足。' }
  }
}
