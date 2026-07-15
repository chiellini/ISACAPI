export default {
  agentsDescription: '开通、暂停代理账号并查看佣金负债',
  withdrawalsDescription: '审核提现申请，并在线下转账后登记打款信息',
  readOnlyHint: '当前为只读模式，仅超级管理员可执行状态和资金操作。',
  statuses: { inactive: '未开通', active: '已开通', suspended: '已暂停' },
  agents: {
    searchPlaceholder: '搜索邮箱、用户名或用户 ID', allStatuses: '全部状态', user: '用户', status: '代理状态', affCode: '邀请码', rebateRate: '佣金比例',
    invitedCount: '直属用户', available: '可提现', frozen: '冻结', reserved: '提现预留', debt: '欠款', activate: '开通代理', suspend: '暂停代理',
    activated: '代理已开通', suspended: '代理已暂停', loadFailed: '加载代理账号失败', updateFailed: '更新代理状态失败'
  },
  withdrawals: {
    searchPlaceholder: '搜索提现单、邮箱、用户名或用户 ID', allStatuses: '全部状态', detailTitle: '提现详情', id: '提现单', user: '代理用户', amount: '申请金额',
    paymentAccount: '收款账户', fullPaymentDetails: '完整收款资料', sensitiveHint: '仅超级管理员可见，请仅在线下转账时使用并妥善保护。', status: '状态', submittedAt: '申请时间', rejectReason: '拒绝原因', rejectReasonPlaceholder: '拒绝时必须填写原因',
    reviewHint: '批准后进入待转账状态；拒绝后预留佣金将释放。', actualCurrency: '实际币种', actualAmount: '实付金额', exchangeRate: '汇率',
    externalReference: '外部流水号', actualPayment: '实际转账', rejectedBecause: '拒绝原因：{reason}', approve: '审核通过', reject: '拒绝申请',
    markPaid: '标记为已经转账', approved: '提现申请已批准', rejected: '提现申请已拒绝', markedPaid: '已标记为已经转账',
    loadFailed: '加载提现列表失败', detailFailed: '加载提现详情失败', actionFailed: '更新提现状态失败',
    statuses: { submitted: '待审核', approved: '待转账', paid: '已经转账', rejected: '已拒绝', canceled: '已取消' }
  }
}
