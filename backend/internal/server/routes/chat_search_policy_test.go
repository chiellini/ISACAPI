package routes

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestDecodeBuiltInChatSearchRequestUsesPositiveSchema(t *testing.T) {
	req, err := decodeBuiltInChatSearchRequest(strings.NewReader(`{
		"model":"claude",
		"query":"current news",
		"max_results":5
	}`))
	require.NoError(t, err)
	require.Equal(t, "claude", req.Model)
	require.Equal(t, "current news", req.Query)
	require.Equal(t, 5, req.MaxResults)

	_, err = decodeBuiltInChatSearchRequest(strings.NewReader(`{
		"model":"claude",
		"query":"current news",
		"provider_native_search":true
	}`))
	require.ErrorContains(t, err, "unknown field")

	_, err = decodeBuiltInChatSearchRequest(strings.NewReader(
		`{"model":"claude","query":"one"} {"model":"claude","query":"two"}`,
	))
	require.Error(t, err)
}
