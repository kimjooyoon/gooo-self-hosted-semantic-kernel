package kernel

import (
	"os"
	"path/filepath"
	"testing"
)

func TestReadOutcomeRejectsNonCanonicalRefutedPayload(t *testing.T) {
	cases := []struct {
		name   string
		mutate func(*Outcome)
	}{
		{
			name: "typed value",
			mutate: func(outcome *Outcome) {
				value := IntValue(7)
				outcome.TypedValue = &value
			},
		},
		{
			name: "unknown evidence",
			mutate: func(outcome *Outcome) {
				outcome.Unknown = &Unknown{Stage: "stage", Step: "step", Reason: "reason", UnknownClass: "class", NextOperation: "next", BlockedBy: "blocked"}
			},
		},
	}

	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			outcome := Outcome{
				Schema:               "gooo.evaluation/v1",
				CaseID:               "case",
				SemanticID:           "semantic",
				EdgeID:               "edge",
				SemanticSchemaDigest: "schema-digest",
				CorpusDigest:         "corpus-digest",
				Status:               StatusRefuted,
				Reason:               "refuted",
				OrderedEffectTrace:   []EffectEvent{},
			}
			testCase.mutate(&outcome)
			raw, err := RenderOutcome(outcome)
			if err != nil {
				t.Fatal(err)
			}
			path := filepath.Join(t.TempDir(), "outcome.json")
			if err := os.WriteFile(path, raw, 0o644); err != nil {
				t.Fatal(err)
			}
			if _, _, err := ReadOutcome(path); err == nil {
				t.Fatal("ReadOutcome accepted a non-canonical REFUTED payload")
			}
		})
	}
}
