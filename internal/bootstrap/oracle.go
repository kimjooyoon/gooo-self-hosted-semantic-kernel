package bootstrap

import (
	"fmt"
	"strings"

	"github.com/kimjooyoon/gooo-self-hosted-semantic-kernel/internal/kernel"
)

// Evaluate is the frozen handwritten oracle. It intentionally shares only the
// parsed data model and output codec with the generator; it never imports or
// executes generated code.
func Evaluate(schema kernel.SemanticSchema, schemaRaw, corpusRaw []byte, c kernel.CaseSpec) (kernel.Outcome, error) {
	outcome := kernel.Outcome{
		Schema:               "gooo.evaluation/v1",
		CaseID:               c.CaseID,
		SemanticID:           c.SemanticID,
		EdgeID:               c.EdgeID,
		SemanticSchemaDigest: kernel.DigestBytes(schemaRaw),
		CorpusDigest:         kernel.DigestBytes(corpusRaw),
		OrderedEffectTrace:   []kernel.EffectEvent{},
	}
	statuses := make([]kernel.Status, 0, len(c.Program))
	var result *kernel.Value
	var unknown *kernel.Unknown
	refutedReason := ""
	for _, step := range c.Program {
		status, value, event, issueUnknown, issueReason, err := evaluateStep(schema, c, step, len(outcome.OrderedEffectTrace))
		if err != nil {
			return kernel.Outcome{}, err
		}
		statuses = append(statuses, status)
		if value != nil {
			valueCopy := *value
			result = &valueCopy
		}
		if event != nil {
			outcome.OrderedEffectTrace = append(outcome.OrderedEffectTrace, *event)
		}
		if issueUnknown != nil && unknown == nil {
			unknown = issueUnknown
		}
		if status == kernel.StatusRefuted && refutedReason == "" {
			refutedReason = issueReason
		}
	}
	outcome.Status = kernel.DominantStatus(schema.Decision.Precedence, statuses)
	switch outcome.Status {
	case kernel.StatusClosed:
		if result == nil {
			outcome.Status = kernel.StatusRefuted
			outcome.Reason = "CLOSED decision had no typed value"
		} else {
			outcome.TypedValue = result
		}
	case kernel.StatusUnknown:
		outcome.Unknown = unknown
		if unknown != nil {
			outcome.Reason = unknown.Reason
		} else {
			outcome.Reason = "UNKNOWN decision lost its six-field record"
		}
	case kernel.StatusRefuted:
		outcome.Reason = refutedReason
		if outcome.Reason == "" {
			outcome.Reason = "REFUTED observation dominates lower-status observations"
		}
	}
	if outcome.Status != kernel.StatusUnknown {
		outcome.Unknown = nil
	}
	if outcome.Status != kernel.StatusClosed {
		outcome.TypedValue = nil
	}
	if err := kernel.FinalizeOutcome(&outcome); err != nil {
		return kernel.Outcome{}, err
	}
	return outcome, nil
}

func evaluateStep(schema kernel.SemanticSchema, c kernel.CaseSpec, step kernel.Step, ordinal int) (kernel.Status, *kernel.Value, *kernel.EffectEvent, *kernel.Unknown, string, error) {
	switch step.Op {
	case "RETURN_INT":
		value := kernel.IntValue(*step.Int)
		return kernel.StatusClosed, &value, nil, nil, "", nil
	case "RETURN_BOOL":
		value := kernel.BoolValue(*step.Bool)
		return kernel.StatusClosed, &value, nil, nil, "", nil
	case "RETURN_STRING":
		value := kernel.StringValue(step.String)
		return kernel.StatusClosed, &value, nil, nil, "", nil
	case "EXTERNAL_INT":
		value, found := c.Externals[step.Name]
		if !found {
			key := step.UnknownKey
			if key == "" {
				key = "external_value"
			}
			u, err := unknownFor(schema, key, step.Name)
			return kernel.StatusUnknown, nil, nil, u, u.Reason, err
		}
		if value.Type != "int" || value.Int == nil {
			return kernel.StatusRefuted, nil, nil, nil, fmt.Sprintf("external %s is not an int", step.Name), nil
		}
		return kernel.StatusClosed, &value, nil, nil, "", nil
	case "EXTERNAL_BOOL":
		value, found := c.Externals[step.Name]
		if !found {
			key := step.UnknownKey
			if key == "" {
				key = "external_value"
			}
			u, err := unknownFor(schema, key, step.Name)
			return kernel.StatusUnknown, nil, nil, u, u.Reason, err
		}
		if value.Type != "bool" || value.Bool == nil {
			return kernel.StatusRefuted, nil, nil, nil, fmt.Sprintf("external %s is not a bool", step.Name), nil
		}
		return kernel.StatusClosed, &value, nil, nil, "", nil
	case "EFFECT_INT":
		if step.Effect != "emit" {
			return kernel.StatusRefuted, nil, nil, nil, fmt.Sprintf("EFFECT_INT cannot use effect %s", step.Effect), nil
		}
		if !hasGrant(c.Grants, step.Effect) {
			u, err := unknownFor(schema, "missing_effect_grant", step.Effect)
			return kernel.StatusUnknown, nil, nil, u, u.Reason, err
		}
		return kernel.StatusClosed, nil, &kernel.EffectEvent{Ordinal: ordinal, Effect: "emit", Value: kernel.IntValue(*step.Int)}, nil, "", nil
	case "EFFECT_BOOL":
		if step.Effect != "audit" {
			return kernel.StatusRefuted, nil, nil, nil, fmt.Sprintf("EFFECT_BOOL cannot use effect %s", step.Effect), nil
		}
		if !hasGrant(c.Grants, step.Effect) {
			u, err := unknownFor(schema, "missing_effect_grant", step.Effect)
			return kernel.StatusUnknown, nil, nil, u, u.Reason, err
		}
		return kernel.StatusClosed, nil, &kernel.EffectEvent{Ordinal: ordinal, Effect: "audit", Value: kernel.BoolValue(*step.Bool)}, nil, "", nil
	case "FIXED_POINT":
		if step.Rule != schema.FixedPoint.Keyword {
			return schema.FixedPoint.MissingRuleStatus, nil, nil, nil, "FIXED_POINT keyword does not match the authoritative rule", nil
		}
		if step.CycleDetected {
			return schema.FixedPoint.CycleStatus, nil, nil, nil, "FIXED_POINT cycle detected", nil
		}
		if step.PriorState == step.NextState && len(step.ObservedStates) <= step.MaxSteps {
			value := kernel.StringValue(step.NextState)
			return kernel.StatusClosed, &value, nil, nil, "", nil
		}
		if len(step.ObservedStates) <= step.MaxSteps {
			u, err := unknownFor(schema, "fixed_point_not_reached", step.Rule)
			return schema.FixedPoint.UnstableStatus, nil, nil, u, u.Reason, err
		}
		return schema.FixedPoint.CycleStatus, nil, nil, nil, "FIXED_POINT exceeded max_steps without stabilization", nil
	case "REFUTE":
		return kernel.StatusRefuted, nil, nil, nil, step.Reason, nil
	default:
		return kernel.StatusRefuted, nil, nil, nil, fmt.Sprintf("unsupported operation %s", step.Op), nil
	}
}

func unknownFor(schema kernel.SemanticSchema, key, subject string) (*kernel.Unknown, error) {
	template, ok := schema.Unknown.Templates[key]
	if !ok {
		return nil, fmt.Errorf("UNKNOWN template %q is not declared by semantic schema", key)
	}
	template.Reason = strings.ReplaceAll(template.Reason, "%s", subject)
	template.Step = strings.ReplaceAll(template.Step, "%s", subject)
	template.NextOperation = strings.ReplaceAll(template.NextOperation, "%s", subject)
	template.BlockedBy = strings.ReplaceAll(template.BlockedBy, "%s", subject)
	return &template, nil
}

func hasGrant(grants []string, wanted string) bool {
	for _, grant := range grants {
		if grant == wanted {
			return true
		}
	}
	return false
}
