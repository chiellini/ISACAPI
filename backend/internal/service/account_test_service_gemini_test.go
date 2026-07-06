//go:build unit

package service

import (
	"context"
	"encoding/json"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"
)

func TestCreateGeminiTestPayload_ImageModel(t *testing.T) {
	t.Parallel()

	payload := createGeminiTestPayload("gemini-2.5-flash-image", "draw a tiny robot")

	var parsed struct {
		Contents []struct {
			Parts []struct {
				Text string `json:"text"`
			} `json:"parts"`
		} `json:"contents"`
		GenerationConfig struct {
			ResponseModalities []string `json:"responseModalities"`
			ImageConfig        struct {
				AspectRatio string `json:"aspectRatio"`
			} `json:"imageConfig"`
		} `json:"generationConfig"`
	}

	require.NoError(t, json.Unmarshal(payload, &parsed))
	require.Len(t, parsed.Contents, 1)
	require.Len(t, parsed.Contents[0].Parts, 1)
	require.Equal(t, "draw a tiny robot", parsed.Contents[0].Parts[0].Text)
	require.Equal(t, []string{"TEXT", "IMAGE"}, parsed.GenerationConfig.ResponseModalities)
	require.Equal(t, "1:1", parsed.GenerationConfig.ImageConfig.AspectRatio)
}

func TestNormalizeGeminiAPIImageModelID(t *testing.T) {
	t.Parallel()

	tests := map[string]string{
		"gemini-3.1-flash-image":         "gemini-3-pro-image-preview",
		"gemini-3.1-flash-image-preview": "gemini-3-pro-image-preview",
		"gemini-3-pro-image":             "gemini-3-pro-image-preview",
		"gemini-3-pro-image-preview":     "gemini-3-pro-image-preview",
		"gemini-2.5-flash-image-preview": "gemini-2.5-flash-image",
		"models/gemini-2.5-flash-image":  "gemini-2.5-flash-image",
	}

	for input, want := range tests {
		if got := normalizeGeminiAPIImageModelID(input); got != want {
			t.Fatalf("normalizeGeminiAPIImageModelID(%q) = %q, want %q", input, got, want)
		}
	}
}

func TestResolveGeminiAccountTestModel_AppliesOAuthMapping(t *testing.T) {
	t.Parallel()

	account := &Account{
		Platform: PlatformGemini,
		Type:     AccountTypeOAuth,
		Credentials: map[string]any{
			"oauth_type": "google_one",
			"model_mapping": map[string]any{
				"gemini-3.1-pro-preview": "gemini-2.5-pro",
				"gemini-3.5-flash":       "gemini-2.5-flash",
			},
		},
	}

	require.Equal(t, "gemini-2.5-pro", resolveGeminiAccountTestModel(account, "gemini-3.1-pro-preview"))
	require.Equal(t, "gemini-2.5-flash", resolveGeminiAccountTestModel(account, "gemini-3.5-flash"))
}

func TestResolveGeminiAccountTestModel_NormalizesMappedOAuthImageModel(t *testing.T) {
	t.Parallel()

	account := &Account{
		Platform: PlatformGemini,
		Type:     AccountTypeOAuth,
		Credentials: map[string]any{
			"oauth_type": "google_one",
			"model_mapping": map[string]any{
				"gemini-3-pro-image": "gemini-3.1-flash-image",
			},
		},
	}

	require.Equal(t, "gemini-3-pro-image-preview", resolveGeminiAccountTestModel(account, "gemini-3-pro-image"))
}

func TestBuildGeminiOAuthRequest_GoogleOneWithoutProjectIDRejectsAIStudioDirect(t *testing.T) {
	t.Parallel()

	svc := &AccountTestService{}
	account := &Account{
		Platform: PlatformGemini,
		Type:     AccountTypeOAuth,
		Credentials: map[string]any{
			"oauth_type":   "google_one",
			"access_token": "ya29.test-token",
		},
	}

	req, err := svc.buildGeminiOAuthRequest(context.Background(), account, "gemini-2.5-flash", createGeminiTestPayload("gemini-2.5-flash", "hi"))
	require.Nil(t, req)
	require.Error(t, err)
	require.Contains(t, err.Error(), "Google One OAuth token uses Gemini CLI / Code Assist scopes")
	require.Contains(t, err.Error(), "project_id")
	require.Contains(t, err.Error(), "AI Studio OAuth/API-key")
}

func TestProcessGeminiStream_EmitsImageEvent(t *testing.T) {
	t.Parallel()
	gin.SetMode(gin.TestMode)

	ctx, recorder := newTestContext()
	svc := &AccountTestService{}

	stream := strings.NewReader("data: {\"candidates\":[{\"content\":{\"parts\":[{\"text\":\"ok\"},{\"inlineData\":{\"mimeType\":\"image/png\",\"data\":\"QUJD\"}}]}}]}\n\ndata: [DONE]\n\n")

	err := svc.processGeminiStream(ctx, stream)
	require.NoError(t, err)

	body := recorder.Body.String()
	require.Contains(t, body, "\"type\":\"content\"")
	require.Contains(t, body, "\"text\":\"ok\"")
	require.Contains(t, body, "\"type\":\"image\"")
	require.Contains(t, body, "\"image_url\":\"data:image/png;base64,QUJD\"")
	require.Contains(t, body, "\"mime_type\":\"image/png\"")
}
