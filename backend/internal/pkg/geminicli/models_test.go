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

func TestGoogleOneModels_ExcludeUnsupportedNewAndImageModels(t *testing.T) {
	t.Parallel()

	mapping := GoogleOneModelMapping()
	for _, id := range []string{"gemini-2.0-flash", "gemini-2.5-flash", "gemini-2.5-pro"} {
		if mapping[id] != id {
			t.Fatalf("expected Google One model %q to map to itself", id)
		}
	}
	for _, id := range []string{"gemini-2.5-flash-image", "gemini-3.1-flash-image", "gemini-3.5-flash"} {
		if _, ok := mapping[id]; ok {
			t.Fatalf("did not expect unsupported Google One model %q", id)
		}
	}

	mapping["gemini-2.5-flash"] = "mutated"
	if GoogleOneModelMapping()["gemini-2.5-flash"] != "gemini-2.5-flash" {
		t.Fatal("GoogleOneModelMapping must return a defensive copy")
	}
}
