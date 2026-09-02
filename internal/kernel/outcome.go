package kernel

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"strings"
)

type terminalPayload struct {
	Status             Status        `json:"status"`
	TypedValue         *Value        `json:"typed_value,omitempty"`
	OrderedEffectTrace []EffectEvent `json:"ordered_effect_trace"`
	Reason             string        `json:"reason,omitempty"`
	Unknown            *Unknown      `json:"unknown,omitempty"`
}

func FinalizeOutcome(outcome *Outcome) error {
	if outcome.OrderedEffectTrace == nil {
		outcome.OrderedEffectTrace = []EffectEvent{}
	}
	payload := terminalPayload{
		Status:             outcome.Status,
		TypedValue:         outcome.TypedValue,
		OrderedEffectTrace: outcome.OrderedEffectTrace,
		Reason:             outcome.Reason,
		Unknown:            outcome.Unknown,
	}
	raw, err := json.Marshal(payload)
	if err != nil {
		return err
	}
	outcome.TerminalDigest = DigestBytes(raw)
	return nil
}

func RenderOutcome(outcome Outcome) ([]byte, error) {
	if err := FinalizeOutcome(&outcome); err != nil {
		return nil, err
	}
	raw, err := json.MarshalIndent(outcome, "", "  ")
	if err != nil {
		return nil, err
	}
	return append(raw, '\n'), nil
}

func WriteOutcome(path string, outcome Outcome) error {
	raw, err := RenderOutcome(outcome)
	if err != nil {
		return err
	}
	if path == "" {
		_, err = os.Stdout.Write(raw)
		return err
	}
	return os.WriteFile(path, raw, 0o644)
}

func ReadOutcome(path string) (Outcome, []byte, error) {
	raw, err := os.ReadFile(path)
	if err != nil {
		return Outcome{}, nil, err
	}
	decoder := json.NewDecoder(bytes.NewReader(raw))
	decoder.DisallowUnknownFields()
	var outcome Outcome
	if err := decoder.Decode(&outcome); err != nil {
		return Outcome{}, nil, fmt.Errorf("parse outcome %s: %w", path, err)
	}
	if outcome.Schema != "gooo.evaluation/v1" || outcome.CaseID == "" || !statusSet[outcome.Status] || outcome.SemanticID == "" || outcome.EdgeID == "" || outcome.SemanticSchemaDigest == "" || outcome.CorpusDigest == "" || outcome.TerminalDigest == "" {
		return Outcome{}, nil, fmt.Errorf("outcome %s has an incomplete identity or status", path)
	}
	if outcome.Status == StatusUnknown {
		if outcome.Unknown == nil || outcome.Unknown.Validate() != nil || outcome.TypedValue != nil {
			return Outcome{}, nil, fmt.Errorf("outcome %s UNKNOWN is not fail-closed", path)
		}
	}
	if outcome.Status == StatusClosed && outcome.TypedValue == nil {
		return Outcome{}, nil, fmt.Errorf("outcome %s CLOSED has no typed value", path)
	}
	if outcome.TypedValue != nil {
		if err := outcome.TypedValue.Validate(); err != nil {
			return Outcome{}, nil, fmt.Errorf("outcome %s: %w", path, err)
		}
	}
	return outcome, raw, nil
}

func ValidTerminalDigest(outcome Outcome) bool {
	copy := outcome
	original := copy.TerminalDigest
	if FinalizeOutcome(&copy) != nil {
		return false
	}
	return original == copy.TerminalDigest
}

type ComparisonReport struct {
	Schema                    string   `json:"schema"`
	CaseID                    string   `json:"case_id"`
	ExpectedStatus            Status   `json:"expected_status"`
	Verdict                   Status   `json:"verdict"`
	Matched                   bool     `json:"matched"`
	ByteEqual                 bool     `json:"byte_equal"`
	ReferenceBytesDigest      string   `json:"reference_bytes_digest"`
	CandidateBytesDigest      string   `json:"candidate_bytes_digest"`
	SemanticIdentityEqual     bool     `json:"semantic_identity_equal"`
	TypedValueEqual           bool     `json:"typed_value_equal"`
	OrderedEffectTraceEqual   bool     `json:"ordered_effect_trace_equal"`
	TerminalDigestEqual       bool     `json:"terminal_digest_equal"`
	ReferenceTerminalValid    bool     `json:"reference_terminal_digest_valid"`
	CandidateTerminalValid    bool     `json:"candidate_terminal_digest_valid"`
	Compared                  []string `json:"compared"`
	ReferenceStatus           Status   `json:"reference_status"`
	CandidateStatus           Status   `json:"candidate_status"`
	Reason                    string   `json:"reason"`
}

func Compare(precedence []Status, caseID string, expected Status, reference Outcome, referenceRaw []byte, candidate Outcome, candidateRaw []byte) ComparisonReport {
	byteEqual := bytes.Equal(referenceRaw, candidateRaw)
	semanticIdentityEqual := reference.CaseID == caseID && candidate.CaseID == caseID && reference.CaseID == candidate.CaseID && reference.SemanticID == candidate.SemanticID && reference.EdgeID == candidate.EdgeID && reference.SemanticSchemaDigest == candidate.SemanticSchemaDigest && reference.CorpusDigest == candidate.CorpusDigest
	typedValueEqual := valuesEqual(reference.TypedValue, candidate.TypedValue)
	traceEqual := valuesEqual(reference.OrderedEffectTrace, candidate.OrderedEffectTrace)
	terminalDigestEqual := reference.TerminalDigest == candidate.TerminalDigest
	valid := ValidTerminalDigest(reference) && ValidTerminalDigest(candidate)
	matched := byteEqual && semanticIdentityEqual && typedValueEqual && traceEqual && terminalDigestEqual && valid
	verdict := DominantStatus(precedence, []Status{reference.Status, candidate.Status})
	reason := "reference oracle and generated evaluator agree"
	if !matched {
		verdict = StatusRefuted
		switch {
		case !byteEqual:
			reason = "canonical output bytes diverged"
		case !semanticIdentityEqual:
			reason = "stable semantic or edge identity diverged"
		case !typedValueEqual:
			reason = "typed value diverged"
		case !traceEqual:
			reason = "ordered effect trace diverged"
		case !terminalDigestEqual:
			reason = "terminal digest diverged"
		default:
			reason = "one terminal digest is invalid"
		}
	}
	return ComparisonReport{
		Schema:                  "gooo.differential-comparison/v1",
		CaseID:                  caseID,
		ExpectedStatus:          expected,
		Verdict:                 verdict,
		Matched:                 matched,
		ByteEqual:               byteEqual,
		ReferenceBytesDigest:    DigestBytes(referenceRaw),
		CandidateBytesDigest:    DigestBytes(candidateRaw),
		SemanticIdentityEqual:   semanticIdentityEqual,
		TypedValueEqual:         typedValueEqual,
		OrderedEffectTraceEqual: traceEqual,
		TerminalDigestEqual:     terminalDigestEqual,
		ReferenceTerminalValid:  ValidTerminalDigest(reference),
		CandidateTerminalValid:  ValidTerminalDigest(candidate),
		Compared:                []string{"output_bytes", "output_digest", "terminal_digest"},
		ReferenceStatus:         reference.Status,
		CandidateStatus:         candidate.Status,
		Reason:                  reason,
	}
}

func ValidateComparison(report ComparisonReport) error {
	if report.Verdict != report.ExpectedStatus {
		return fmt.Errorf("case %s verdict %s, want %s", report.CaseID, report.Verdict, report.ExpectedStatus)
	}
	if report.Verdict == StatusUnknown && !report.Matched {
		return fmt.Errorf("case %s is UNKNOWN without a matched oracle result", report.CaseID)
	}
	return nil
}

func valuesEqual(left, right any) bool {
	leftRaw, leftErr := json.Marshal(left)
	rightRaw, rightErr := json.Marshal(right)
	return leftErr == nil && rightErr == nil && strings.TrimSpace(string(leftRaw)) == strings.TrimSpace(string(rightRaw))
}
