package service

import (
	"fmt"

	"github.com/Wei-Shaw/sub2api/internal/config"
)

// ProvideConversationCaptureSink 按配置构造 CaptureSink。
//
// 未启用 → NoopCaptureSink（零成本）。启用 → ConversationArchiver。
// EncryptContent 开启但密钥非法时在启动阶段直接报错（fail-closed on misconfig），
// 避免运行期才发现无法加密。归档器为进程生命周期单例：进程退出时队列中少量未落库记录
// 允许丢失（fail-open 取舍），故不经 wire cleanup 处理。
func ProvideConversationCaptureSink(cfg *config.Config, repo ConversationRepository) (CaptureSink, error) {
	ca := cfg.ConversationArchive
	if !ca.Enabled {
		return NoopCaptureSink{}, nil
	}

	var enc *ContentEncryptor
	if ca.EncryptContent {
		e, err := NewContentEncryptor(ca.EncryptionKey)
		if err != nil {
			return nil, fmt.Errorf("conversation archive: encryption enabled but key invalid: %w", err)
		}
		enc = e
	}

	var redactor *SecretRedactor
	if ca.RedactSecrets {
		redactor = NewSecretRedactor()
	}

	archiver := NewConversationArchiver(repo, enc, redactor, ConversationArchiverOptions{
		Mode:             ca.Mode,
		QueueSize:        ca.QueueSize,
		MaxEventBytes:    ca.MaxEventBytes,
		MaxSessionEvents: ca.MaxSessionEvents,
		RedactSecrets:    ca.RedactSecrets,
		EncryptContent:   ca.EncryptContent,
		CaptureSystem:    ca.CaptureSystem,
	})
	return archiver, nil
}

// ConversationArchiveEnabled 返回是否启用对话存档（供 handler 快速短路）。
func ConversationArchiveEnabled(cfg *config.Config) bool {
	return cfg != nil && cfg.ConversationArchive.Enabled
}
