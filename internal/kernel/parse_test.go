package kernel

import "testing"

func TestFrozenCorpusShape(t *testing.T) {
	schema, _, err := LoadSchema("../../.gooo/semantic.gooo")
	if err != nil {
		t.Fatal(err)
	}
	corpus, _, err := LoadCorpus("../../.gooo/corpus.gooo")
	if err != nil {
		t.Fatal(err)
	}
	if err := ValidateCorpus(schema, corpus); err != nil {
		t.Fatal(err)
	}
	if got := len(SortedCaseIDs(corpus)); got != 12 {
		t.Fatalf("case count = %d, want 12", got)
	}
}

func TestDecisionPrecedence(t *testing.T) {
	precedence := []Status{StatusRefuted, StatusUnknown, StatusClosed}
	if got := DominantStatus(precedence, []Status{StatusClosed, StatusUnknown}); got != StatusUnknown {
		t.Fatalf("dominant status = %s, want UNKNOWN", got)
	}
	if got := DominantStatus(precedence, []Status{StatusClosed, StatusUnknown, StatusRefuted}); got != StatusRefuted {
		t.Fatalf("dominant status = %s, want REFUTED", got)
	}
}
