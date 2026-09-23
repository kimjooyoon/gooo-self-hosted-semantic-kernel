package kernel

import (
	"path/filepath"
	"testing"
)

func TestReadOutcomeRejectsRefutedTypedValue(t *testing.T) {
	path := filepath.Join(t.TempDir(), "refuted.json")
	value := IntValue(1)
	outcome := Outcome{
		Schema: "gooo.evaluation/v1", CaseID: "case-07-refuted", SemanticID: "semantic-07", EdgeID: "edge-07",
		SemanticSchemaDigest: "sha256:schema", CorpusDigest: "sha256:corpus", Status: StatusRefuted,
		TypedValue: &value, Reason: "refuted fixture",
	}
	if err := WriteOutcome(path, outcome); err != nil {
		t.Fatal(err)
	}
	if _, _, err := ReadOutcome(path); err == nil {
		t.Fatal("REFUTED outcome with typed value was accepted")
	}
}
