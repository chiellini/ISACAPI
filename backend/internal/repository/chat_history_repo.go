package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/Wei-Shaw/sub2api/internal/service"
)

// chatHistoryRepository 用原生 SQL 实现聊天历史，刻意不依赖 ent 生成代码。
type chatHistoryRepository struct {
	db *sql.DB
}

// NewChatHistoryRepository 创建聊天历史仓储。
func NewChatHistoryRepository(db *sql.DB) service.ChatHistoryRepository {
	return &chatHistoryRepository{db: db}
}

func (r *chatHistoryRepository) ListSessions(ctx context.Context, userID int64) ([]service.ChatHistorySession, error) {
	rows, err := r.db.QueryContext(ctx,
		`SELECT id, title, model, updated_at FROM chat_sessions WHERE user_id = $1 ORDER BY updated_at DESC`,
		userID,
	)
	if err != nil {
		return nil, fmt.Errorf("list chat sessions: %w", err)
	}
	defer rows.Close()

	out := make([]service.ChatHistorySession, 0)
	for rows.Next() {
		var s service.ChatHistorySession
		if err := rows.Scan(&s.ID, &s.Title, &s.Model, &s.UpdatedAt); err != nil {
			return nil, err
		}
		out = append(out, s)
	}
	return out, rows.Err()
}

func (r *chatHistoryRepository) GetSession(ctx context.Context, userID, id int64) (*service.ChatHistorySession, error) {
	var s service.ChatHistorySession
	err := r.db.QueryRowContext(ctx,
		`SELECT id, title, model, updated_at FROM chat_sessions WHERE id = $1 AND user_id = $2`,
		id, userID,
	).Scan(&s.ID, &s.Title, &s.Model, &s.UpdatedAt)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, service.ErrChatSessionNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("get chat session: %w", err)
	}

	rows, err := r.db.QueryContext(ctx,
		`SELECT role, content FROM chat_messages WHERE session_id = $1 ORDER BY id`,
		id,
	)
	if err != nil {
		return nil, fmt.Errorf("get chat messages: %w", err)
	}
	defer rows.Close()

	s.Messages = make([]service.ChatHistoryMessage, 0)
	for rows.Next() {
		var m service.ChatHistoryMessage
		if err := rows.Scan(&m.Role, &m.Content); err != nil {
			return nil, err
		}
		s.Messages = append(s.Messages, m)
	}
	return &s, rows.Err()
}

func (r *chatHistoryRepository) CreateSession(ctx context.Context, userID int64, title, model string) (int64, error) {
	var id int64
	err := r.db.QueryRowContext(ctx,
		`INSERT INTO chat_sessions (user_id, title, model) VALUES ($1, $2, $3) RETURNING id`,
		userID, title, model,
	).Scan(&id)
	if err != nil {
		return 0, fmt.Errorf("create chat session: %w", err)
	}
	return id, nil
}

// UpdateSession 在一个事务内校验归属、更新会话元数据、整体替换消息列表。
func (r *chatHistoryRepository) UpdateSession(ctx context.Context, userID, id int64, title, model string, msgs []service.ChatHistoryMessage) error {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer func() { _ = tx.Rollback() }()

	res, err := tx.ExecContext(ctx,
		`UPDATE chat_sessions SET title = $1, model = $2, updated_at = NOW() WHERE id = $3 AND user_id = $4`,
		title, model, id, userID,
	)
	if err != nil {
		return fmt.Errorf("update chat session: %w", err)
	}
	affected, _ := res.RowsAffected()
	if affected == 0 {
		return service.ErrChatSessionNotFound
	}

	// msgs == nil 表示仅更新元数据（如重命名），保留原有消息不动。
	if msgs == nil {
		return tx.Commit()
	}

	if _, err := tx.ExecContext(ctx, `DELETE FROM chat_messages WHERE session_id = $1`, id); err != nil {
		return fmt.Errorf("clear chat messages: %w", err)
	}
	for _, m := range msgs {
		if _, err := tx.ExecContext(ctx,
			`INSERT INTO chat_messages (session_id, role, content) VALUES ($1, $2, $3)`,
			id, m.Role, m.Content,
		); err != nil {
			return fmt.Errorf("insert chat message: %w", err)
		}
	}
	return tx.Commit()
}

func (r *chatHistoryRepository) DeleteSession(ctx context.Context, userID, id int64) error {
	// chat_messages 通过 ON DELETE CASCADE 自动清理。
	res, err := r.db.ExecContext(ctx,
		`DELETE FROM chat_sessions WHERE id = $1 AND user_id = $2`,
		id, userID,
	)
	if err != nil {
		return fmt.Errorf("delete chat session: %w", err)
	}
	if affected, _ := res.RowsAffected(); affected == 0 {
		return service.ErrChatSessionNotFound
	}
	return nil
}
