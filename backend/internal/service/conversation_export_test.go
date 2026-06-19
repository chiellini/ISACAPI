package service

import (
	"archive/zip"
	"bytes"
	"context"
	"strings"
	"testing"

	"github.com/google/uuid"
)

func TestConversationService_ExportSessionText(t *testing.T) {
	repo := newFakeConversationRepo()
	a := newTestArchiver(repo, ConversationModeUserAssistantText)
	a.Submit(context.Background(), mkRecord(1, "ak1", ContextDomainOpenAIAPI, "req-1", "what is go?", "go is a language", "resp-1", ""))
	a.Stop()

	var id uuid.UUID
	for sid := range repo.sessions {
		id = sid
	}
	svc := &ConversationService{repo: repo}

	filename, content, err := svc.ExportSessionText(context.Background(), id)
	if err != nil {
		t.Fatalf("export: %v", err)
	}
	if !strings.HasSuffix(filename, ".txt") {
		t.Fatalf("filename = %q", filename)
	}
	text := string(content)
	if !strings.Contains(text, "what is go?") || !strings.Contains(text, "go is a language") {
		t.Fatalf("transcript missing content:\n%s", text)
	}
	if !strings.Contains(text, "USER") || !strings.Contains(text, "ASSISTANT") {
		t.Fatalf("transcript missing roles:\n%s", text)
	}
}

func TestConversationService_ExportZip(t *testing.T) {
	repo := newFakeConversationRepo()
	a := newTestArchiver(repo, ConversationModeUserAssistantText)
	a.Submit(context.Background(), mkRecord(1, "ak1", ContextDomainOpenAIAPI, "req-1", "q1", "a1", "resp-1", ""))
	a.Submit(context.Background(), mkRecord(1, "ak2", ContextDomainOpenAIAPI, "req-2", "q2", "a2", "resp-2", ""))
	a.Stop()

	svc := &ConversationService{repo: repo}
	filename, content, err := svc.ExportZip(context.Background(), ConversationSessionListFilters{})
	if err != nil {
		t.Fatalf("export zip: %v", err)
	}
	if !strings.HasSuffix(filename, ".zip") {
		t.Fatalf("filename = %q", filename)
	}
	zr, err := zip.NewReader(bytes.NewReader(content), int64(len(content)))
	if err != nil {
		t.Fatalf("read zip: %v", err)
	}
	if len(zr.File) != 2 {
		t.Fatalf("expected 2 entries, got %d", len(zr.File))
	}
}
