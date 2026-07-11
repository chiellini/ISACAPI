export default {
  conversations: {
    title: 'Conversation Archive',
    description: 'Review, export, and clean up gateway conversation sessions',
    filters: {
      userId: 'Search by user ID',
      allStatuses: 'All statuses'
    },
    columns: {
      lastActive: 'Last Active',
      user: 'User',
      contextDomain: 'Context Domain',
      requests: 'Requests',
      tokens: 'Input / Output Tokens',
      status: 'Status',
      actions: 'Actions'
    },
    status: {
      active: 'Active',
      archived: 'Archived',
      expired: 'Expired'
    },
    roles: {
      user: 'User',
      assistant: 'Assistant',
      tool: 'Tool',
      system: 'System'
    },
    view: 'View',
    exportAll: 'Export All',
    exporting: 'Exporting...',
    downloadTxt: 'Download TXT',
    deleteAfterDownload: 'Delete Archive',
    detailTitle: 'Conversation Details',
    noEvents: 'No conversation events',
    partial: 'Partial',
    previewOnly: 'Preview only',
    deleteTitle: 'Delete Conversation Archive',
    deleteConfirm: 'Delete this conversation archive? This action cannot be undone.',
    deleteSuccess: 'Conversation archive deleted',
    deleteFailed: 'Failed to delete conversation archive',
    loadFailed: 'Failed to load conversation archives',
    loadDetailFailed: 'Failed to load conversation details',
    exportFailed: 'Failed to export conversation archive'
  }
}
