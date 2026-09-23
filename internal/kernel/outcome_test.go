package kernel

import (
	"os"
	"path/filepath"
	"testing"
)

func TestReadOutcomeRejectsTrailingJSONValue(t *testing.T) {
	value := int64(1)
	raw, err := RenderOutcome(Outcome{
		Schema:               "gooo.evaluation/v1",
		CaseID:               "case-1",
		SemanticID:           "semantic-1",
		EdgeID:               "edge-1",
		SemanticSchemaDigest: "sha256:schema",
		CorpusDigest:         "sha256:corpus",
		Status:               StatusClosed,
		TypedValue:           &Value{Type: "int", Int: &value},
	})
	if err != nil {
		t.Fatal(err)
	}
	raw = append(raw, []byte("{}\n")...)
	path := filepath.Join(t.TempDir(), "outcome.json")
	if err := os.WriteFile(path, raw, 0o600); err != nil {
		t.Fatal(err)
	}
	if _, _, err := ReadOutcome(path); err == nil {
		t.Fatal("ReadOutcome accepted a trailing JSON value")
	}
}
