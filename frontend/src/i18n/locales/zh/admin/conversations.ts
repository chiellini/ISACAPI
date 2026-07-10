export default {
  conversations: {
    title: '会话归档',
    description: '查看、导出和清理网关保存的对话会话',
    filters: {
      userId: '按用户 ID 搜索',
      allStatuses: '全部状态'
    },
    columns: {
      lastActive: '最后活跃',
      user: '用户',
      contextDomain: '上下文域',
      requests: '请求数',
      tokens: '输入 / 输出 Token',
      status: '状态',
      actions: '操作'
    },
    status: {
      active: '活跃',
      archived: '已归档',
      expired: '已过期'
    },
    roles: {
      user: '用户',
      assistant: '助手',
      tool: '工具',
      system: '系统'
    },
    view: '查看',
    exportAll: '全部导出',
    exporting: '导出中...',
    downloadTxt: '下载 TXT',
    deleteAfterDownload: '删除归档',
    detailTitle: '会话详情',
    noEvents: '暂无会话事件',
    partial: '部分内容',
    previewOnly: '仅预览',
    deleteTitle: '删除会话归档',
    deleteConfirm: '确定要删除这个会话归档吗？此操作无法撤销。',
    deleteSuccess: '会话归档已删除',
    deleteFailed: '删除会话归档失败',
    loadFailed: '加载会话归档失败',
    loadDetailFailed: '加载会话详情失败',
    exportFailed: '导出会话归档失败'
  }
}
