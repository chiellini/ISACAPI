package service

import (
	"context"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestChatSearchRejectsOversizedQueryBeforeProviderLookup(t *testing.T) {
	svc := &ChatHistoryService{}
	_, err := svc.SearchWeb(context.Background(), strings.Repeat("界", chatSearchMaxQueryRunes+1), 5)
	require.ErrorIs(t, err, ErrChatSearchQueryTooLong)
}
