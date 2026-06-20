package service

import (
	"context"
	"sync"
	"time"
)

// CaptureRecordVersion 是 CaptureRecord 的结构版本，便于演进而不破坏消费者。
const CaptureRecordVersion = 1

// CaptureSink 是对话采集的唯一投递接口。
//
// 第一版起就只保留一个方法：以后扩展只改 CaptureRecord 结构体，不改接口方法签名，
// 从而避免给现有 GatewayService interface 加方法导致所有 mock/stub 需要同步补全。
//
// 实现必须 fail-open 且非阻塞：队列满或后端故障时丢弃记录并计数，绝不阻塞转发主路径。
type CaptureSink interface {
	Submit(ctx context.Context, record CaptureRecord)
}

// CaptureRecord 是一次请求/响应采集的版本化快照。
type CaptureRecord struct {
	Version  int
	Identity ConversationIdentity
	Request  NormalizedRequest
	Response NormalizedResponse
	Error    *CaptureError
	Timing   CaptureTiming
}

// NormalizedRequest 是从客户端请求中规范化出的可归档部分（不含任何鉴权头/凭据）。
type NormalizedRequest struct {
	UserID    int64
	APIKeyID  *int64
	GroupID   int64
	Protocol  string
	Model     string
	RequestID string
	// 归一后的入站事件（用户文本、可选 system）。明文，落库前由 sink 加密 + 脱敏。
	Events []NormalizedEvent
}

// NormalizedResponse 是从上游响应（含 SSE 聚合后）规范化出的可归档部分。
type NormalizedResponse struct {
	Model        string
	ResponseID   string // 上游 response/message id（原值，sink 落库前哈希）
	FinishReason string
	Partial      bool  // 流中断导致不完整
	InputTokens  int64 // 用于会话聚合（可选）
	OutputTokens int64 // 用于会话聚合（可选）
	// 归一后的出站事件（最终 assistant 文本，第一版不含工具/思维链）。
	Events []NormalizedEvent
}

// NormalizedEvent 是协议无关的归一事件。Content 为明文，sink 负责加密/脱敏后落库。
type NormalizedEvent struct {
	Role       string // system / user / assistant / tool
	Kind       string // message / tool_call / tool_result / ...
	Content    string
	ToolCallID string // 上游稳定 id（原值，sink 落库前哈希）
}

// CaptureError 描述采集过程中的非致命错误（仅用于诊断，不影响转发）。
type CaptureError struct {
	Stage   string
	Message string
}

// CaptureTiming 记录采集相关时间点。
type CaptureTiming struct {
	StartedAt    time.Time
	FinishedAt   time.Time
	FirstTokenAt *time.Time
}

// NoopCaptureSink 丢弃所有记录，用于关闭采集或测试时注入。
type NoopCaptureSink struct{}

// Submit 实现 CaptureSink，什么都不做。
func (NoopCaptureSink) Submit(context.Context, CaptureRecord) {}

// MemoryCaptureSink 在内存中收集记录，用于测试。并发安全。
type MemoryCaptureSink struct {
	mu      sync.Mutex
	records []CaptureRecord
}

// Submit 实现 CaptureSink，记录到内存。
func (m *MemoryCaptureSink) Submit(_ context.Context, record CaptureRecord) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.records = append(m.records, record)
}

// Records 返回已收集记录的副本。
func (m *MemoryCaptureSink) Records() []CaptureRecord {
	m.mu.Lock()
	defer m.mu.Unlock()
	out := make([]CaptureRecord, len(m.records))
	copy(out, m.records)
	return out
}

var (
	_ CaptureSink = NoopCaptureSink{}
	_ CaptureSink = (*MemoryCaptureSink)(nil)
)
