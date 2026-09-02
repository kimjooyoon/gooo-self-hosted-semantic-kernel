package kernel

import "errors"

type Status string

const (
	StatusClosed  Status = "CLOSED"
	StatusUnknown Status = "UNKNOWN"
	StatusRefuted Status = "REFUTED"
)

var statusSet = map[Status]bool{
	StatusClosed:  true,
	StatusUnknown: true,
	StatusRefuted: true,
}

type Value struct {
	Type   string  `json:"type"`
	Int    *int64  `json:"int,omitempty"`
	Bool   *bool   `json:"bool,omitempty"`
	String *string `json:"string,omitempty"`
}

type EffectEvent struct {
	Ordinal int    `json:"ordinal"`
	Effect  string `json:"effect"`
	Value   Value  `json:"value"`
}

type Unknown struct {
	Stage         string `json:"stage"`
	Step          string `json:"step"`
	Reason        string `json:"reason"`
	UnknownClass  string `json:"unknown_class"`
	NextOperation string `json:"next_operation"`
	BlockedBy     string `json:"blocked_by"`
}

func (u Unknown) Validate() error {
	if u.Stage == "" || u.Step == "" || u.Reason == "" || u.UnknownClass == "" || u.NextOperation == "" || u.BlockedBy == "" {
		return errors.New("UNKNOWN requires stage, step, reason, unknown_class, next_operation, and blocked_by")
	}
	return nil
}

type Outcome struct {
	Schema               string        `json:"schema"`
	CaseID               string        `json:"case_id"`
	SemanticID           string        `json:"semantic_id"`
	EdgeID               string        `json:"edge_id"`
	SemanticSchemaDigest string        `json:"semantic_schema_digest"`
	CorpusDigest         string        `json:"corpus_digest"`
	Status               Status        `json:"status"`
	TypedValue           *Value        `json:"typed_value,omitempty"`
	OrderedEffectTrace   []EffectEvent `json:"ordered_effect_trace"`
	Reason               string        `json:"reason,omitempty"`
	Unknown              *Unknown      `json:"unknown,omitempty"`
	TerminalDigest       string        `json:"terminal_digest"`
}

type IdentityField struct {
	Field      string `json:"field"`
	Authority  string `json:"authority"`
	Uniqueness string `json:"uniqueness"`
	Meaning    string `json:"meaning"`
}

type IdentitySchema struct {
	SemanticID IdentityField `json:"semantic_id"`
	EdgeID     IdentityField `json:"edge_id"`
}

type DecisionSchema struct {
	Statuses            []Status `json:"statuses"`
	Precedence          []Status `json:"precedence"`
	FailClosedOnUnknown bool     `json:"fail_closed_on_unknown"`
	ComparisonFields    []string `json:"comparison_fields"`
	TopDecision         string   `json:"top_decision"`
}

type UnknownSchema struct {
	Fields    []string           `json:"fields"`
	Templates map[string]Unknown `json:"templates"`
}

type FixedPointSchema struct {
	Keyword           string `json:"keyword"`
	Rule              string `json:"rule"`
	MaxStepsRequired  bool   `json:"max_steps_required"`
	UnstableStatus    Status `json:"unstable_status"`
	CycleStatus       Status `json:"cycle_status"`
	MissingRuleStatus Status `json:"missing_rule_status"`
}

type OperationSchema struct {
	Input       string `json:"input"`
	Output      string `json:"output"`
	Description string `json:"description"`
}

type SemanticSchema struct {
	Schema     string                     `json:"schema"`
	Authority  string                     `json:"authority"`
	Language   string                     `json:"language"`
	Version    string                     `json:"version"`
	Scope      string                     `json:"scope"`
	Types      []string                   `json:"types"`
	Identity   IdentitySchema             `json:"identity"`
	Decision   DecisionSchema             `json:"decision"`
	Unknown    UnknownSchema              `json:"unknown"`
	FixedPoint FixedPointSchema           `json:"fixed_point"`
	Operations map[string]OperationSchema `json:"operations"`
}

type Cell struct {
	SemanticID string `json:"semantic_id"`
	EdgeID     string `json:"edge_id"`
	CaseID     string `json:"case_id"`
	Kind       string `json:"kind"`
	Meaning    string `json:"meaning"`
}

type MetaActivity struct {
	SemanticID string `json:"semantic_id"`
	EdgeID     string `json:"edge_id"`
	CaseID     string `json:"case_id"`
	Action     string `json:"action"`
	Meaning    string `json:"meaning"`
}

type ProofChoice struct {
	SemanticID string `json:"semantic_id"`
	EdgeID     string `json:"edge_id"`
	CaseID     string `json:"case_id"`
	Status     Status `json:"status"`
	Choice     string `json:"choice"`
}

type Indicator struct {
	SemanticID string `json:"semantic_id"`
	EdgeID     string `json:"edge_id"`
	CaseID     string `json:"case_id"`
	Status     Status `json:"status"`
	Name       string `json:"name"`
}

type Step struct {
	Op             string   `json:"op"`
	Int            *int64   `json:"int,omitempty"`
	Bool           *bool    `json:"bool,omitempty"`
	String         string   `json:"string,omitempty"`
	Name           string   `json:"name,omitempty"`
	UnknownKey     string   `json:"unknown_key,omitempty"`
	Effect         string   `json:"effect,omitempty"`
	Rule           string   `json:"rule,omitempty"`
	PriorState     string   `json:"prior_state,omitempty"`
	NextState      string   `json:"next_state,omitempty"`
	ObservedStates []string `json:"observed_states,omitempty"`
	MaxSteps       int      `json:"max_steps,omitempty"`
	CycleDetected  bool     `json:"cycle_detected,omitempty"`
	Reason         string   `json:"reason,omitempty"`
}

type CaseSpec struct {
	CaseID         string           `json:"case_id"`
	SemanticID     string           `json:"semantic_id"`
	EdgeID         string           `json:"edge_id"`
	CellID         string           `json:"cell_id"`
	MetaActivityID string           `json:"meta_activity_id"`
	ProofChoiceID  string           `json:"proof_choice_id"`
	IndicatorID    string           `json:"indicator_id"`
	ExpectedStatus Status           `json:"expected_status"`
	Externals      map[string]Value `json:"externals,omitempty"`
	Grants         []string         `json:"grants,omitempty"`
	Program        []Step           `json:"program"`
}

type Corpus struct {
	Schema         string         `json:"schema"`
	Authority      string         `json:"authority"`
	Version        string         `json:"version"`
	EntityCounts   map[string]int `json:"entity_counts"`
	Cells          []Cell         `json:"cells"`
	MetaActivities []MetaActivity `json:"meta_activities"`
	ProofChoices   []ProofChoice  `json:"proof_choices"`
	Indicators     []Indicator    `json:"indicators"`
	Cases          []CaseSpec     `json:"cases"`
}

func IntValue(value int64) Value { return Value{Type: "int", Int: &value} }

func BoolValue(value bool) Value { return Value{Type: "bool", Bool: &value} }

func StringValue(value string) Value { return Value{Type: "string", String: &value} }

func (v Value) Validate() error {
	switch v.Type {
	case "int":
		if v.Int == nil || v.Bool != nil || v.String != nil {
			return errors.New("int value must contain only int")
		}
	case "bool":
		if v.Bool == nil || v.Int != nil || v.String != nil {
			return errors.New("bool value must contain only bool")
		}
	case "string":
		if v.String == nil || v.Int != nil || v.Bool != nil {
			return errors.New("string value must contain only string")
		}
	default:
		return errors.New("value type must be int, bool, or string")
	}
	return nil
}
