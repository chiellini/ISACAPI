export default {
  affiliate: {
    title: 'Affiliate Center',
    description: 'Share your invite link and manage commission withdrawals',
    yourCode: 'My Affiliate Code', inviteLink: 'Invite Link', copyCode: 'Copy Code', copyLink: 'Copy Link',
    codeCopied: 'Affiliate code copied', linkCopied: 'Invite link copied', loadFailed: 'Failed to load affiliate data',
    status: {
      inactive: { title: 'Affiliate access not enabled', description: 'Contact an administrator to review and activate your affiliate account.' },
      suspended: { title: 'Affiliate account suspended', description: 'New referrals and commissions are paused. Contact an administrator.' },
      active: { title: 'Affiliate active', description: 'Your affiliate account is active.' }
    },
    stats: {
      rebateRate: 'Commission Rate', rebateRateHint: 'Commission earned on invitees’ real payments', invitedUsers: 'Direct Referrals',
      availableCommission: 'Available Commission', frozenCommission: 'Frozen Commission', withdrawalReserved: 'Withdrawal Reserved', debt: 'Affiliate Debt',
      availableQuota: 'Available Commission', frozenQuota: 'Frozen Commission', frozenQuotaHint: 'Available after the hold period', totalQuota: 'Historical Commission'
    },
    paymentAccounts: {
      title: 'Payout Accounts', description: 'Account details are encrypted and only masked information is shown.', add: 'Add Account', edit: 'Edit Account', empty: 'No payout account yet',
      default: 'Default', setDefault: 'Set as default payout account', type: 'Payout Method', accountName: 'Account Holder', bankName: 'Bank Name', alipayAccount: 'Alipay Account',
      cardNumber: 'Card Number', network: 'USDT Network', walletAddress: 'Wallet Address', reenterHint: 'For security, enter the complete payout details again when editing.',
      deleteConfirm: 'Delete this payout account?', saved: 'Payout account saved', deleted: 'Payout account deleted', loadFailed: 'Failed to load payout accounts',
      saveFailed: 'Failed to save payout account', deleteFailed: 'Failed to delete payout account', types: { alipay: 'Alipay', bank_card: 'Bank Card', usdt: 'USDT' }
    },
    withdrawal: {
      title: 'Request Withdrawal', description: 'Minimum {amount}, no fee. An administrator will review and transfer offline.', account: 'Payout Account', selectAccount: 'Select a payout account',
      amount: 'Amount', available: 'Available: {amount}', debtBlocked: 'You have affiliate debt of {amount}. Withdrawals are unavailable until it is repaid.', submit: 'Submit Withdrawal',
      submitting: 'Submitting...', submitted: 'Withdrawal submitted', submitFailed: 'Failed to submit withdrawal', records: 'Withdrawal History', empty: 'No withdrawals yet',
      loadFailed: 'Failed to load withdrawals', id: 'Withdrawal', statusLabel: 'Status', submittedAt: 'Submitted At', cancel: 'Cancel', canceled: 'Withdrawal canceled',
      cancelFailed: 'Failed to cancel withdrawal', statuses: { submitted: 'Pending Review', approved: 'Awaiting Transfer', paid: 'Transferred', rejected: 'Rejected', canceled: 'Canceled' }
    },
    invitees: { title: 'Direct Referrals', empty: 'No referrals yet', columns: { email: 'Email', username: 'Username', rebate: 'Commission', joinedAt: 'Joined At' } },
    tips: { title: 'How It Works', line1: 'Share your affiliate code or invite link with new users.', line2: 'Earn {rate} commission on invitees’ real payments.', line3: 'Request an offline payout after the hold period.', line4: 'Refunds reverse commission and may create affiliate debt.' }
  }
}
