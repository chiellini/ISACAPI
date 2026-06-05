package gemini

import "testing"

func TestDefaultModels_ContainsFallbackCatalogModels(t *testing.T) {
	t.Parallel()

	models := DefaultModels()
	byName := make(map[string]Model, len(models))
	for _, model := range models {
		byName[model.Name] = model
	}

	required := []string{
		"models/gemini-2.5-flash",
		"models/gemini-2.5-pro",
		"models/gemini-3-flash-preview",
		"models/gemini-3-pro-preview",
		"models/gemini-3.1-pro-preview",
		"models/gemini-3.5-flash",
	}

	for _, name := range required {
		model, ok := byName[name]
		if !ok {
			t.Fatalf("expected fallback model %q to exist", name)
		}
		if len(model.SupportedGenerationMethods) == 0 {
			t.Fatalf("expected fallback model %q to advertise generation methods", name)
		}
	}
}

func TestHasFallbackModel_RecognizesCuratedModel(t *testing.T) {
	t.Parallel()

	if !HasFallbackModel("gemini-3.1-pro-preview") {
		t.Fatalf("expected curated model to exist in fallback catalog")
	}
	if !HasFallbackModel("models/gemini-3.1-pro-preview") {
		t.Fatalf("expected prefixed curated model to exist in fallback catalog")
	}
	if HasFallbackModel("gemini-unknown") {
		t.Fatalf("did not expect unknown model to exist in fallback catalog")
	}
}
