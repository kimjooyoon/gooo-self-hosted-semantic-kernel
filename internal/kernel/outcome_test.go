package kernel

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

func TestReadOutcomeRejectsForgedTerminalDigest(t *testing.T) {
	value := IntValue(7)
	outcome := Outcome{
		Schema:               "gooo.evaluation/v1",
		CaseID:               "case-1",
		SemanticID:           "semantic-1",
		EdgeID:               "edge-1",
		SemanticSchemaDigest: "sha256:schema",
		CorpusDigest:         "sha256:corpus",
		Status:               StatusClosed,
		TypedValue:           &value,
		OrderedEffectTrace:   []EffectEvent{},
		TerminalDigest:       "sha256:forged",
	}
	raw, err := json.MarshalIndent(outcome, "", "  ")
	if err != nil {
		t.Fatal(err)
	}
	raw = append(raw, '\n')
	path := filepath.Join(t.TempDir(), "outcome.json")
	if err := os.WriteFile(path, raw, 0o600); err != nil {
		t.Fatal(err)
	}
	if _, _, err := ReadOutcome(path); err == nil {
		t.Fatal("forged terminal digest was accepted")
	}
}
