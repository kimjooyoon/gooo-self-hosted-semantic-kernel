package kernel

import (
	"path/filepath"
	"testing"
)

func TestValidateProgramRejectsConflictingTypedStepFields(t *testing.T) {
	schema, _, err := LoadSchema(filepath.Join("..", "..", ".gooo", "semantic.gooo"))
	if err != nil {
		t.Fatal(err)
	}
	value := int64(7)
	flag := true
	caseSpec := CaseSpec{Program: []Step{{Op: "RETURN_INT", Int: &value, Bool: &flag}}}
	if err := validateProgram(schema, caseSpec); err == nil {
		t.Fatal("validateProgram accepted conflicting typed fields")
	}
}
