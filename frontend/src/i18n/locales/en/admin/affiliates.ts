export default {
  agentsDescription: 'Activate or suspend affiliate accounts and review commission liabilities',
  withdrawalsDescription: 'Review withdrawal requests and record completed offline transfers',
  readOnlyHint: 'Read-only mode. Only a super administrator can change status or funds.',
  statuses: { inactive: 'Not Enabled', active: 'Active', suspended: 'Suspended' },
  agents: {
    searchPlaceholder: 'Search email, username, or user ID', allStatuses: 'All Statuses', user: 'User', status: 'Affiliate Status', affCode: 'Affiliate Code', rebateRate: 'Commission Rate',
    invitedCount: 'Direct Referrals', available: 'Available', frozen: 'Frozen', reserved: 'Reserved', debt: 'Debt', activate: 'Activate', suspend: 'Suspend',
    activated: 'Affiliate activated', suspended: 'Affiliate suspended', loadFailed: 'Failed to load affiliates', updateFailed: 'Failed to update affiliate status'
  },
  withdrawals: {
    searchPlaceholder: 'Search withdrawal, email, username, or user ID', allStatuses: 'All Statuses', detailTitle: 'Withdrawal Details', id: 'Withdrawal', user: 'Affiliate', amount: 'Requested Amount',
    paymentAccount: 'Payout Account', fullPaymentDetails: 'Full Payout Details', sensitiveHint: 'Visible only to super administrators. Use only for the offline transfer and protect this data.', status: 'Status', submittedAt: 'Submitted At', rejectReason: 'Rejection Reason', rejectReasonPlaceholder: 'A reason is required to reject',
    reviewHint: 'Approval moves the request to awaiting transfer; rejection releases reserved commission.', actualCurrency: 'Actual Currency', actualAmount: 'Actual Amount', exchangeRate: 'Exchange Rate',
    externalReference: 'External Reference', actualPayment: 'Actual Transfer', rejectedBecause: 'Rejection reason: {reason}', approve: 'Approve', reject: 'Reject', markPaid: 'Mark as Transferred',
    approved: 'Withdrawal approved', rejected: 'Withdrawal rejected', markedPaid: 'Marked as transferred', loadFailed: 'Failed to load withdrawals', detailFailed: 'Failed to load withdrawal details', actionFailed: 'Failed to update withdrawal',
    statuses: { submitted: 'Pending Review', approved: 'Awaiting Transfer', paid: 'Transferred', rejected: 'Rejected', canceled: 'Canceled' }
  }
}
