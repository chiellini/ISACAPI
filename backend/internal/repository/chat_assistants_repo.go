package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/Wei-Shaw/sub2api/internal/service"
)

// chatAssistantsRepository 用原生 SQL 实现用户自定义助手（与聊天历史同一模式，不依赖 ent）。
type chatAssistantsRepository struct {
	db *sql.DB
}

// NewChatAssistantsRepository 创建用户助手仓储。
func NewChatAssistantsRepository(db *sql.DB) service.ChatAssistantsRepository {
	return &chatAssistantsRepository{db: db}
}

func (r *chatAssistantsRepository) List(ctx context.Context, userID int64) ([]service.ChatAssistant, error) {
	rows, err := r.db.QueryContext(ctx,
		`SELECT id, name, description, system_prompt, model, opening_message, updated_at
		   FROM chat_assistants WHERE user_id = $1 ORDER BY updated_at DESC`,
		userID,
	)
	if err != nil {
		return nil, fmt.Errorf("list chat assistants: %w", err)
	}
	defer func() { _ = rows.Close() }()

	out := make([]service.ChatAssistant, 0)
	for rows.Next() {
		var a service.ChatAssistant
		if err := rows.Scan(&a.ID, &a.Name, &a.Description, &a.SystemPrompt, &a.Model, &a.OpeningMessage, &a.UpdatedAt); err != nil {
			return nil, err
		}
		out = append(out, a)
	}
	return out, rows.Err()
}

func (r *chatAssistantsRepository) Get(ctx context.Context, userID, id int64) (*service.ChatAssistant, error) {
	var a service.ChatAssistant
	err := r.db.QueryRowContext(ctx,
		`SELECT id, name, description, system_prompt, model, opening_message, updated_at
		   FROM chat_assistants WHERE id = $1 AND user_id = $2`,
		id, userID,
	).Scan(&a.ID, &a.Name, &a.Description, &a.SystemPrompt, &a.Model, &a.OpeningMessage, &a.UpdatedAt)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, service.ErrChatAssistantNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("get chat assistant: %w", err)
	}
	return &a, nil
}

func (r *chatAssistantsRepository) Create(ctx context.Context, userID int64, in service.ChatAssistantInput) (int64, error) {
	var id int64
	err := r.db.QueryRowContext(ctx,
		`INSERT INTO chat_assistants (user_id, name, description, system_prompt, model, opening_message)
		 VALUES ($1, $2, $3, $4, $5, $6) RETURNING id`,
		userID, in.Name, in.Description, in.SystemPrompt, in.Model, in.OpeningMessage,
	).Scan(&id)
	if err != nil {
		return 0, fmt.Errorf("create chat assistant: %w", err)
	}
	return id, nil
}

func (r *chatAssistantsRepository) Update(ctx context.Context, userID, id int64, in service.ChatAssistantInput) error {
	res, err := r.db.ExecContext(ctx,
		`UPDATE chat_assistants
		    SET name = $1, description = $2, system_prompt = $3, model = $4, opening_message = $5, updated_at = NOW()
		  WHERE id = $6 AND user_id = $7`,
		in.Name, in.Description, in.SystemPrompt, in.Model, in.OpeningMessage, id, userID,
	)
	if err != nil {
		return fmt.Errorf("update chat assistant: %w", err)
	}
	if affected, _ := res.RowsAffected(); affected == 0 {
		return service.ErrChatAssistantNotFound
	}
	return nil
}

func (r *chatAssistantsRepository) Delete(ctx context.Context, userID, id int64) error {
	// 关联会话的 assistant_id 由 ON DELETE SET NULL 置空，会话与消息保留。
	res, err := r.db.ExecContext(ctx,
		`DELETE FROM chat_assistants WHERE id = $1 AND user_id = $2`,
		id, userID,
	)
	if err != nil {
		return fmt.Errorf("delete chat assistant: %w", err)
	}
	if affected, _ := res.RowsAffected(); affected == 0 {
		return service.ErrChatAssistantNotFound
	}
	return nil
}

func (r *chatAssistantsRepository) CountByUser(ctx context.Context, userID int64) (int, error) {
	var count int
	if err := r.db.QueryRowContext(ctx,
		`SELECT COUNT(*) FROM chat_assistants WHERE user_id = $1`, userID,
	).Scan(&count); err != nil {
		return 0, fmt.Errorf("count chat assistants: %w", err)
	}
	return count, nil
}
