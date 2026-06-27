package service

import (
	"testing"
)

func TestAggregateStatus(t *testing.T) {
	cases := []struct {
		name        string
		operational int
		known       int
		want        string
	}{
		{"all operational", 3, 3, PublicStatusValueOperational},
		{"all down", 0, 3, PublicStatusValueDown},
		{"partial degraded", 1, 3, PublicStatusValueDegraded},
		{"no data treated as down", 0, 0, PublicStatusValueDown},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if got := aggregateStatus(tc.operational, tc.known); got != tc.want {
				t.Fatalf("aggregateStatus(%d,%d)=%q want %q", tc.operational, tc.known, got, tc.want)
			}
		})
	}
}

func TestRollupStatus(t *testing.T) {
	cases := []struct {
		name            string
		known, op, down int
		want            string
	}{
		{"all operational", 2, 2, 0, PublicStatusValueOperational},
		{"all down", 2, 0, 2, PublicStatusValueDown},
		{"mixed degraded", 3, 1, 1, PublicStatusValueDegraded},
		{"one down one ok degraded", 2, 1, 1, PublicStatusValueDegraded},
		{"empty down", 0, 0, 0, PublicStatusValueDown},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if got := rollupStatus(tc.known, tc.op, tc.down); got != tc.want {
				t.Fatalf("rollupStatus(%d,%d,%d)=%q want %q", tc.known, tc.op, tc.down, got, tc.want)
			}
		})
	}
}

func TestBuildProviderStatus_AveragesAvailabilityAndSkipsUnknownModels(t *testing.T) {
	models := map[string]*modelStatusAccumulator{
		// operational, two channels report availability -> averaged
		"claude-opus": {operational: 2, known: 2, availSum: 0.99 + 0.97, availCount: 2},
		// degraded: one of two channels operational
		"claude-sonnet": {operational: 1, known: 2, availSum: 0.95, availCount: 1},
		// no known status at all -> must be skipped entirely
		"ghost-model": {operational: 0, known: 0},
	}

	ps := buildProviderStatus(MonitorProviderAnthropic, models)

	if len(ps.Models) != 2 {
		t.Fatalf("expected 2 models (ghost skipped), got %d", len(ps.Models))
	}
	// Sorted alphabetically: claude-opus before claude-sonnet.
	if ps.Models[0].Model != "claude-opus" || ps.Models[1].Model != "claude-sonnet" {
		t.Fatalf("models not sorted: %+v", ps.Models)
	}
	opus := ps.Models[0]
	if opus.Status != PublicStatusValueOperational {
		t.Fatalf("opus status = %q want operational", opus.Status)
	}
	if !opus.HasAvailability || opus.Availability7d <= 0.97 || opus.Availability7d >= 0.99 {
		t.Fatalf("opus availability avg wrong: %v (has=%v)", opus.Availability7d, opus.HasAvailability)
	}
	if ps.Models[1].Status != PublicStatusValueDegraded {
		t.Fatalf("sonnet status = %q want degraded", ps.Models[1].Status)
	}
	// Provider: one operational + one degraded -> degraded.
	if ps.Status != PublicStatusValueDegraded {
		t.Fatalf("provider status = %q want degraded", ps.Status)
	}
	if !ps.HasAvailability {
		t.Fatalf("provider should have an aggregate availability")
	}
}

func TestBuildProviderStatus_CollectsDedupedSortedGroups(t *testing.T) {
	models := map[string]*modelStatusAccumulator{
		// Same model seen across channels in two distinct groups (one duplicated).
		"claude-opus": {
			operational: 2, known: 2,
			groups: map[string]struct{}{"vip": {}, "default": {}},
		},
		// Model with no group info -> Groups must be nil (omitted as label).
		"claude-sonnet": {operational: 1, known: 1},
	}

	ps := buildProviderStatus(MonitorProviderAnthropic, models)

	opus := ps.Models[0] // sorted: claude-opus first
	if opus.Model != "claude-opus" {
		t.Fatalf("unexpected first model %q", opus.Model)
	}
	if len(opus.Groups) != 2 || opus.Groups[0] != "default" || opus.Groups[1] != "vip" {
		t.Fatalf("groups not deduped/sorted: %+v", opus.Groups)
	}
	if ps.Models[1].Groups != nil {
		t.Fatalf("model without groups should have nil Groups, got %+v", ps.Models[1].Groups)
	}
}

func TestSortProviders_KnownOrderThenAlpha(t *testing.T) {
	providers := []PublicProviderStatus{
		{Provider: "zeta"},
		{Provider: MonitorProviderGemini},
		{Provider: MonitorProviderAnthropic},
		{Provider: "alpha"},
		{Provider: MonitorProviderOpenAI},
	}
	sortProviders(providers)
	want := []string{
		MonitorProviderAnthropic,
		MonitorProviderOpenAI,
		MonitorProviderGemini,
		"alpha",
		"zeta",
	}
	for i, w := range want {
		if providers[i].Provider != w {
			t.Fatalf("position %d = %q want %q (full: %+v)", i, providers[i].Provider, w, providers)
		}
	}
}

func TestAggregateOverallStatus(t *testing.T) {
	allOp := []PublicProviderStatus{
		{Status: PublicStatusValueOperational},
		{Status: PublicStatusValueOperational},
	}
	if got := aggregateOverallStatus(allOp); got != PublicStatusValueOperational {
		t.Fatalf("all operational overall = %q want operational", got)
	}
	mixed := []PublicProviderStatus{
		{Status: PublicStatusValueOperational},
		{Status: PublicStatusValueDown},
	}
	if got := aggregateOverallStatus(mixed); got != PublicStatusValueDegraded {
		t.Fatalf("mixed overall = %q want degraded", got)
	}
}
