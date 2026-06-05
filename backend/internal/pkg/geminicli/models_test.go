package geminicli

import "testing"

func TestDefaultModels_MatchesCuratedTextModels(t *testing.T) {
	t.Parallel()

	got := make([]string, len(DefaultModels))
	for i, model := range DefaultModels {
		got[i] = model.ID
	}

	want := []string{
		"gemini-2.5-flash",
		"gemini-2.5-pro",
		"gemini-3-flash-preview",
		"gemini-3-pro-preview",
		"gemini-3.1-pro-preview",
		"gemini-3.5-flash",
	}

	if len(got) != len(want) {
		t.Fatalf("expected %d curated Gemini models, got %d: %v", len(want), len(got), got)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("model %d mismatch: got %q, want %q", i, got[i], want[i])
		}
	}
}
