package kernel

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"sort"
)

const (
	SchemaAuthority = "github.com/kimjooyoon/gooo-self-hosted-semantic-kernel/.gooo/semantic.gooo"
	CorpusAuthority = "github.com/kimjooyoon/gooo-self-hosted-semantic-kernel/.gooo/corpus.gooo"
)

func LoadSchema(path string) (SemanticSchema, []byte, error) {
	raw, err := os.ReadFile(path)
	if err != nil {
		return SemanticSchema{}, nil, err
	}
	var schema SemanticSchema
	if err := decodeStrict(raw, &schema, "semantic schema"); err != nil {
		return SemanticSchema{}, nil, err
	}
	if err := schema.Validate(); err != nil {
		return SemanticSchema{}, nil, err
	}
	return schema, raw, nil
}

func LoadCorpus(path string) (Corpus, []byte, error) {
	raw, err := os.ReadFile(path)
	if err != nil {
		return Corpus{}, nil, err
	}
	var corpus Corpus
	if err := decodeStrict(raw, &corpus, "corpus"); err != nil {
		return Corpus{}, nil, err
	}
	return corpus, raw, nil
}

func decodeStrict(raw []byte, destination any, label string) error {
	decoder := json.NewDecoder(bytes.NewReader(raw))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(destination); err != nil {
		return fmt.Errorf("parse %s: %w", label, err)
	}
	var extra any
	if err := decoder.Decode(&extra); !errors.Is(err, io.EOF) {
		if err == nil {
			return fmt.Errorf("parse %s: trailing JSON value", label)
		}
		return fmt.Errorf("parse %s: trailing data: %w", label, err)
	}
	return nil
}

func (s SemanticSchema) Validate() error {
	if s.Schema != "gooo.semantic-schema/v1" || s.Authority != SchemaAuthority || s.Language != "Gooo" || s.Version != "0.1.0" {
		return errors.New("semantic schema identity is not the frozen v0.1.0 authority")
	}
	if s.Scope == "" || len(s.Types) != 6 {
		return errors.New("semantic schema must declare the six minimum semantic types")
	}
	if !equalStrings(s.Types, []string{"Cell", "MetaActivity", "ProofChoice", "Indicator", "Decision", "Unknown"}) {
		return errors.New("semantic schema types are not the minimum declared order")
	}
	if err := validateIdentity(s.Identity); err != nil {
		return err
	}
	if err := validateDecision(s.Decision); err != nil {
		return err
	}
	if err := validateUnknownSchema(s.Unknown); err != nil {
		return err
	}
	if err := validateFixedPoint(s.FixedPoint); err != nil {
		return err
	}
	wantedOperations := []string{"RETURN_INT", "RETURN_BOOL", "RETURN_STRING", "EXTERNAL_INT", "EXTERNAL_BOOL", "EFFECT_INT", "EFFECT_BOOL", "FIXED_POINT", "REFUTE"}
	if len(s.Operations) != len(wantedOperations) {
		return errors.New("semantic schema operation set is not frozen")
	}
	for _, name := range wantedOperations {
		op, ok := s.Operations[name]
		if !ok || op.Input == "" || op.Output == "" || op.Description == "" {
			return fmt.Errorf("operation %q is missing or incomplete", name)
		}
	}
	return nil
}

func validateIdentity(identity IdentitySchema) error {
	if identity.SemanticID.Field != "semantic_id" || identity.EdgeID.Field != "edge_id" {
		return errors.New("stable identity fields must be semantic_id and edge_id")
	}
	if identity.SemanticID.Authority == "" || identity.EdgeID.Authority == "" || identity.SemanticID.Meaning == "" || identity.EdgeID.Meaning == "" {
		return errors.New("stable identity authority and meaning are required")
	}
	if identity.SemanticID.Uniqueness != "global" || identity.EdgeID.Uniqueness != "global" {
		return errors.New("stable semantic_id and edge_id uniqueness must be global")
	}
	return nil
}

func validateDecision(decision DecisionSchema) error {
	if !equalStatuses(decision.Statuses, []Status{StatusClosed, StatusUnknown, StatusRefuted}) {
		return errors.New("decision statuses are not CLOSED, UNKNOWN, REFUTED")
	}
	if !equalStatuses(decision.Precedence, []Status{StatusRefuted, StatusUnknown, StatusClosed}) {
		return errors.New("decision precedence must be REFUTED > UNKNOWN > CLOSED")
	}
	if !decision.FailClosedOnUnknown || decision.TopDecision == "" {
		return errors.New("top decision must be explicitly fail-closed")
	}
	if !equalStrings(decision.ComparisonFields, []string{"output_bytes", "output_digest", "terminal_digest"}) {
		return errors.New("comparison fields are not the frozen byte and digest fields")
	}
	return nil
}

func validateUnknownSchema(unknown UnknownSchema) error {
	wantedFields := []string{"stage", "step", "reason", "unknown_class", "next_operation", "blocked_by"}
	if !equalStrings(unknown.Fields, wantedFields) {
		return errors.New("UNKNOWN must declare exactly the six required fields")
	}
	wantedTemplates := []string{"external_value", "top_decision", "fixed_point_not_reached", "missing_effect_grant"}
	if len(unknown.Templates) != len(wantedTemplates) {
		return errors.New("UNKNOWN template set is not frozen")
	}
	for _, key := range wantedTemplates {
		template, ok := unknown.Templates[key]
		if !ok {
			return fmt.Errorf("UNKNOWN template %q is missing", key)
		}
		if err := template.Validate(); err != nil {
			return fmt.Errorf("UNKNOWN template %q: %w", key, err)
		}
	}
	return nil
}

func validateFixedPoint(fixedPoint FixedPointSchema) error {
	if fixedPoint.Keyword != "FIXED_POINT" || fixedPoint.Rule == "" || !fixedPoint.MaxStepsRequired {
		return errors.New("FIXED_POINT rule must be explicit and bounded")
	}
	if fixedPoint.UnstableStatus != StatusUnknown || fixedPoint.CycleStatus != StatusRefuted || fixedPoint.MissingRuleStatus != StatusRefuted {
		return errors.New("FIXED_POINT status outcomes are not fail-closed")
	}
	return nil
}

func ValidateCorpus(schema SemanticSchema, corpus Corpus) error {
	if corpus.Schema != "gooo.corpus/v1" || corpus.Authority != CorpusAuthority || corpus.Version != "0.1.0" {
		return errors.New("corpus identity is not the frozen v0.1.0 authority")
	}
	wantedCounts := map[string]int{"cells": 12, "meta_activities": 12, "proof_choices": 12, "indicators": 12, "cases": 12}
	for name, wanted := range wantedCounts {
		if corpus.EntityCounts[name] != wanted {
			return fmt.Errorf("corpus declared %s count %d, want %d", name, corpus.EntityCounts[name], wanted)
		}
	}
	if len(corpus.EntityCounts) != len(wantedCounts) || len(corpus.Cells) != 12 || len(corpus.MetaActivities) != 12 || len(corpus.ProofChoices) != 12 || len(corpus.Indicators) != 12 || len(corpus.Cases) != 12 {
		return errors.New("corpus must contain exactly twelve cells, meta activities, proof choices, indicators, and cases")
	}
	semanticIDs := map[string]string{}
	edgeIDs := map[string]string{}
	caseIDs := map[string]bool{}
	for index, cell := range corpus.Cells {
		if err := addIdentity(semanticIDs, edgeIDs, "cell", index, cell.SemanticID, cell.EdgeID); err != nil {
			return err
		}
		if cell.CaseID == "" || cell.Kind == "" || cell.Meaning == "" {
			return fmt.Errorf("cell %d is incomplete", index)
		}
	}
	for index, activity := range corpus.MetaActivities {
		if err := addIdentity(semanticIDs, edgeIDs, "meta_activity", index, activity.SemanticID, activity.EdgeID); err != nil {
			return err
		}
		if activity.CaseID == "" || activity.Action == "" || activity.Meaning == "" {
			return fmt.Errorf("meta activity %d is incomplete", index)
		}
	}
	for index, choice := range corpus.ProofChoices {
		if err := addIdentity(semanticIDs, edgeIDs, "proof_choice", index, choice.SemanticID, choice.EdgeID); err != nil {
			return err
		}
		if choice.CaseID == "" || !statusSet[choice.Status] || choice.Choice == "" {
			return fmt.Errorf("proof choice %d is incomplete", index)
		}
	}
	for index, indicator := range corpus.Indicators {
		if err := addIdentity(semanticIDs, edgeIDs, "indicator", index, indicator.SemanticID, indicator.EdgeID); err != nil {
			return err
		}
		if indicator.CaseID == "" || !statusSet[indicator.Status] || indicator.Name == "" {
			return fmt.Errorf("indicator %d is incomplete", index)
		}
	}
	cellByCase := map[string]Cell{}
	activityByCase := map[string]MetaActivity{}
	for _, cell := range corpus.Cells {
		cellByCase[cell.CaseID] = cell
	}
	for _, activity := range corpus.MetaActivities {
		activityByCase[activity.CaseID] = activity
	}
	proofByCase := map[string]ProofChoice{}
	indicatorByCase := map[string]Indicator{}
	for _, choice := range corpus.ProofChoices {
		proofByCase[choice.CaseID] = choice
	}
	for _, indicator := range corpus.Indicators {
		indicatorByCase[indicator.CaseID] = indicator
	}
	for index, c := range corpus.Cases {
		if err := addIdentity(semanticIDs, edgeIDs, "case", index, c.SemanticID, c.EdgeID); err != nil {
			return err
		}
		if c.CaseID == "" || caseIDs[c.CaseID] || !statusSet[c.ExpectedStatus] || len(c.Program) == 0 {
			return fmt.Errorf("case %d is incomplete or duplicated", index)
		}
		caseIDs[c.CaseID] = true
		cell, cellOK := cellByCase[c.CaseID]
		activity, activityOK := activityByCase[c.CaseID]
		proof, proofOK := proofByCase[c.CaseID]
		indicator, indicatorOK := indicatorByCase[c.CaseID]
		if !cellOK || !activityOK || !proofOK || !indicatorOK || c.CellID != cell.SemanticID || c.MetaActivityID != activity.SemanticID || c.ProofChoiceID != proof.SemanticID || c.IndicatorID != indicator.SemanticID {
			return fmt.Errorf("case %s has incomplete stable entity links", c.CaseID)
		}
		if proof.Status != c.ExpectedStatus || indicator.Status != c.ExpectedStatus {
			return fmt.Errorf("case %s status is not reflected by proof and indicator", c.CaseID)
		}
		if err := validateProgram(schema, c); err != nil {
			return fmt.Errorf("case %s: %w", c.CaseID, err)
		}
	}
	for caseID := range cellByCase {
		if !caseIDs[caseID] {
			return fmt.Errorf("cell %s has no case", caseID)
		}
	}
	for _, statuses := range []struct {
		name string
		list []Status
	}{
		{name: "proof choices", list: proofStatuses(corpus.ProofChoices)},
		{name: "indicators", list: indicatorStatuses(corpus.Indicators)},
		{name: "cases", list: caseStatuses(corpus.Cases)},
	} {
		if err := validateDistribution(statuses.name, statuses.list); err != nil {
			return err
		}
	}
	return nil
}

func addIdentity(semanticIDs, edgeIDs map[string]string, kind string, index int, semanticID, edgeID string) error {
	if semanticID == "" || edgeID == "" {
		return fmt.Errorf("%s %d has empty stable identity", kind, index)
	}
	if previous, exists := semanticIDs[semanticID]; exists {
		return fmt.Errorf("semantic_id %q is duplicated by %s and %s", semanticID, previous, kind)
	}
	if previous, exists := edgeIDs[edgeID]; exists {
		return fmt.Errorf("edge_id %q is duplicated by %s and %s", edgeID, previous, kind)
	}
	semanticIDs[semanticID] = kind
	edgeIDs[edgeID] = kind
	return nil
}

func validateProgram(schema SemanticSchema, c CaseSpec) error {
	for index, step := range c.Program {
		if _, ok := schema.Operations[step.Op]; !ok {
			return fmt.Errorf("program step %d uses undeclared operation %q", index, step.Op)
		}
		switch step.Op {
		case "RETURN_INT":
			if step.Int == nil {
				return fmt.Errorf("program step %d RETURN_INT needs int", index)
			}
		case "RETURN_BOOL":
			if step.Bool == nil {
				return fmt.Errorf("program step %d RETURN_BOOL needs bool", index)
			}
		case "RETURN_STRING":
			if step.String == "" {
				return fmt.Errorf("program step %d RETURN_STRING needs string", index)
			}
		case "EXTERNAL_INT", "EXTERNAL_BOOL":
			if step.Name == "" {
				return fmt.Errorf("program step %d external operation needs name", index)
			}
			if step.UnknownKey != "" {
				if _, ok := schema.Unknown.Templates[step.UnknownKey]; !ok {
					return fmt.Errorf("program step %d uses undeclared UNKNOWN key %q", index, step.UnknownKey)
				}
			}
		case "EFFECT_INT":
			if step.Effect == "" || step.Int == nil {
				return fmt.Errorf("program step %d EFFECT_INT needs effect and int", index)
			}
		case "EFFECT_BOOL":
			if step.Effect == "" || step.Bool == nil {
				return fmt.Errorf("program step %d EFFECT_BOOL needs effect and bool", index)
			}
		case "FIXED_POINT":
			if step.Rule == "" || step.PriorState == "" || step.NextState == "" || len(step.ObservedStates) == 0 || step.MaxSteps <= 0 {
				return fmt.Errorf("program step %d FIXED_POINT needs explicit bounded states", index)
			}
		case "REFUTE":
			if step.Reason == "" {
				return fmt.Errorf("program step %d REFUTE needs reason", index)
			}
		}
	}
	return nil
}

func validateDistribution(name string, statuses []Status) error {
	counts := map[Status]int{}
	for _, status := range statuses {
		counts[status]++
	}
	if counts[StatusClosed] != 4 || counts[StatusUnknown] != 4 || counts[StatusRefuted] != 4 {
		return fmt.Errorf("%s must contain exactly four CLOSED, four UNKNOWN, and four REFUTED vectors", name)
	}
	return nil
}

func proofStatuses(values []ProofChoice) []Status {
	statuses := make([]Status, 0, len(values))
	for _, value := range values {
		statuses = append(statuses, value.Status)
	}
	return statuses
}

func indicatorStatuses(values []Indicator) []Status {
	statuses := make([]Status, 0, len(values))
	for _, value := range values {
		statuses = append(statuses, value.Status)
	}
	return statuses
}

func caseStatuses(values []CaseSpec) []Status {
	statuses := make([]Status, 0, len(values))
	for _, value := range values {
		statuses = append(statuses, value.ExpectedStatus)
	}
	return statuses
}

func equalStrings(left, right []string) bool {
	if len(left) != len(right) {
		return false
	}
	for index := range left {
		if left[index] != right[index] {
			return false
		}
	}
	return true
}

func equalStatuses(left, right []Status) bool {
	if len(left) != len(right) {
		return false
	}
	for index := range left {
		if left[index] != right[index] {
			return false
		}
	}
	return true
}

func StatusRank(precedence []Status, status Status) int {
	for index, candidate := range precedence {
		if candidate == status {
			return len(precedence) - index
		}
	}
	return 0
}

func DominantStatus(precedence []Status, statuses []Status) Status {
	chosen := StatusClosed
	chosenRank := StatusRank(precedence, chosen)
	for _, status := range statuses {
		if rank := StatusRank(precedence, status); rank > chosenRank {
			chosen = status
			chosenRank = rank
		}
	}
	return chosen
}

func SortedCaseIDs(corpus Corpus) []string {
	ids := make([]string, 0, len(corpus.Cases))
	for _, c := range corpus.Cases {
		ids = append(ids, c.CaseID)
	}
	sort.Strings(ids)
	return ids
}

func FindCase(corpus Corpus, caseID string) (CaseSpec, error) {
	for _, c := range corpus.Cases {
		if c.CaseID == caseID {
			return c, nil
		}
	}
	return CaseSpec{}, fmt.Errorf("case %q is not in the fixed corpus", caseID)
}

func DigestBytes(raw []byte) string {
	digest := sha256.Sum256(raw)
	return "sha256:" + hex.EncodeToString(digest[:])
}
