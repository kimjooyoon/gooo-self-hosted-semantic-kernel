package main

import (
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"os"

	"github.com/kimjooyoon/gooo-self-hosted-semantic-kernel/internal/bootstrap"
	"github.com/kimjooyoon/gooo-self-hosted-semantic-kernel/internal/generator"
	"github.com/kimjooyoon/gooo-self-hosted-semantic-kernel/internal/kernel"
)

func main() {
	if len(os.Args) < 2 {
		fatal(errors.New("command is required: validate, generate, bootstrap, compare, or summary"))
	}
	switch os.Args[1] {
	case "validate":
		validateCommand(os.Args[2:])
	case "generate":
		generateCommand(os.Args[2:])
	case "bootstrap":
		bootstrapCommand(os.Args[2:])
	case "compare":
		compareCommand(os.Args[2:])
	case "summary":
		summaryCommand(os.Args[2:])
	default:
		fatal(fmt.Errorf("unknown command %q", os.Args[1]))
	}
}

func loadInputs(schemaPath, corpusPath string) (kernel.SemanticSchema, []byte, kernel.Corpus, []byte) {
	schema, schemaRaw, err := kernel.LoadSchema(schemaPath)
	if err != nil {
		fatal(err)
	}
	corpus, corpusRaw, err := kernel.LoadCorpus(corpusPath)
	if err != nil {
		fatal(err)
	}
	if err := kernel.ValidateCorpus(schema, corpus); err != nil {
		fatal(err)
	}
	return schema, schemaRaw, corpus, corpusRaw
}

func validateCommand(args []string) {
	flags := flag.NewFlagSet("validate", flag.ExitOnError)
	schemaPath := flags.String("schema", ".gooo/semantic.gooo", "semantic schema path")
	corpusPath := flags.String("corpus", ".gooo/corpus.gooo", "fixed corpus path")
	output := flags.String("output", "", "JSON report path; stdout when empty")
	flags.Parse(args)
	schema, schemaRaw, corpus, corpusRaw := loadInputs(*schemaPath, *corpusPath)
	report := validationReport{
		Schema:                "gooo.validation/v1",
		Status:                kernel.StatusClosed,
		SemanticSchemaDigest: kernel.DigestBytes(schemaRaw),
		CorpusDigest:          kernel.DigestBytes(corpusRaw),
		DecisionPrecedence:    schema.Decision.Precedence,
		UnknownFields:         schema.Unknown.Fields,
		FixedPointKeyword:     schema.FixedPoint.Keyword,
		EntityCounts:          corpus.EntityCounts,
		ProofChoicesByStatus:  countProofChoices(corpus.ProofChoices),
		IndicatorsByStatus:     countIndicators(corpus.Indicators),
		CasesByStatus:          countCases(corpus.Cases),
	}
	writeJSON(*output, report)
}

func generateCommand(args []string) {
	flags := flag.NewFlagSet("generate", flag.ExitOnError)
	schemaPath := flags.String("schema", ".gooo/semantic.gooo", "semantic schema path")
	corpusPath := flags.String("corpus", ".gooo/corpus.gooo", "fixed corpus path")
	output := flags.String("output", "generated", "generated output directory")
	flags.Parse(args)
	schema, schemaRaw, corpus, corpusRaw := loadInputs(*schemaPath, *corpusPath)
	if err := generator.Generate(schema, schemaRaw, corpus, corpusRaw, *output); err != nil {
		fatal(err)
	}
	writeJSON("", map[string]any{
		"schema":                 "gooo.generation/v1",
		"status":                 kernel.StatusClosed,
		"semantic_schema_digest": kernel.DigestBytes(schemaRaw),
		"corpus_digest":           kernel.DigestBytes(corpusRaw),
		"generated_files":         []string{"evaluator.go", "main.go"},
	})
}

func bootstrapCommand(args []string) {
	flags := flag.NewFlagSet("bootstrap", flag.ExitOnError)
	schemaPath := flags.String("schema", ".gooo/semantic.gooo", "semantic schema path")
	corpusPath := flags.String("corpus", ".gooo/corpus.gooo", "fixed corpus path")
	caseID := flags.String("case-id", "", "fixed corpus case identifier")
	output := flags.String("output", "", "outcome path; stdout when empty")
	flags.Parse(args)
	if *caseID == "" {
		fatal(errors.New("--case-id is required"))
	}
	schema, schemaRaw, corpus, corpusRaw := loadInputs(*schemaPath, *corpusPath)
	c, err := kernel.FindCase(corpus, *caseID)
	if err != nil {
		fatal(err)
	}
	outcome, err := bootstrap.Evaluate(schema, schemaRaw, corpusRaw, c)
	if err != nil {
		fatal(err)
	}
	if err := kernel.WriteOutcome(*output, outcome); err != nil {
		fatal(err)
	}
}

func compareCommand(args []string) {
	flags := flag.NewFlagSet("compare", flag.ExitOnError)
	schemaPath := flags.String("schema", ".gooo/semantic.gooo", "semantic schema path")
	corpusPath := flags.String("corpus", ".gooo/corpus.gooo", "fixed corpus path")
	caseID := flags.String("case-id", "", "fixed corpus case identifier")
	expected := flags.String("expected", "", "expected verdict")
	referencePath := flags.String("reference", "", "bootstrap outcome path")
	candidatePath := flags.String("candidate", "", "generated outcome path")
	output := flags.String("output", "", "comparison report path; stdout when empty")
	flags.Parse(args)
	if *caseID == "" || *expected == "" || *referencePath == "" || *candidatePath == "" {
		fatal(errors.New("--case-id, --expected, --reference, and --candidate are required"))
	}
	schema, _, corpus, _ := loadInputs(*schemaPath, *corpusPath)
	if _, err := kernel.FindCase(corpus, *caseID); err != nil {
		fatal(err)
	}
	reference, referenceRaw, err := kernel.ReadOutcome(*referencePath)
	if err != nil {
		fatal(err)
	}
	candidate, candidateRaw, err := kernel.ReadOutcome(*candidatePath)
	if err != nil {
		fatal(err)
	}
	report := kernel.Compare(schema.Decision.Precedence, *caseID, kernel.Status(*expected), reference, referenceRaw, candidate, candidateRaw)
	writeJSON(*output, report)
	if err := kernel.ValidateComparison(report); err != nil {
		fatal(err)
	}
}

func summaryCommand(args []string) {
	flags := flag.NewFlagSet("summary", flag.ExitOnError)
	schemaPath := flags.String("schema", ".gooo/semantic.gooo", "semantic schema path")
	corpusPath := flags.String("corpus", ".gooo/corpus.gooo", "fixed corpus path")
	output := flags.String("output", "", "JSON report path; stdout when empty")
	flags.Parse(args)
	_, schemaRaw, corpus, corpusRaw := loadInputs(*schemaPath, *corpusPath)
	vectors := make([]vectorSummary, 0, len(corpus.Cases))
	for _, c := range corpus.Cases {
		vectors = append(vectors, vectorSummary{CaseID: c.CaseID, SemanticID: c.SemanticID, EdgeID: c.EdgeID, ExpectedStatus: c.ExpectedStatus})
	}
	writeJSON(*output, map[string]any{
		"schema":                 "gooo.corpus-summary/v1",
		"status":                 kernel.StatusClosed,
		"semantic_schema_digest": kernel.DigestBytes(schemaRaw),
		"corpus_digest":           kernel.DigestBytes(corpusRaw),
		"vectors":                vectors,
	})
}

type validationReport struct {
	Schema                 string                     `json:"schema"`
	Status                 kernel.Status              `json:"status"`
	SemanticSchemaDigest   string                     `json:"semantic_schema_digest"`
	CorpusDigest           string                     `json:"corpus_digest"`
	DecisionPrecedence     []kernel.Status             `json:"decision_precedence"`
	UnknownFields          []string                   `json:"unknown_fields"`
	FixedPointKeyword      string                     `json:"fixed_point_keyword"`
	EntityCounts           map[string]int             `json:"entity_counts"`
	ProofChoicesByStatus   map[kernel.Status]int      `json:"proof_choices_by_status"`
	IndicatorsByStatus     map[kernel.Status]int      `json:"indicators_by_status"`
	CasesByStatus          map[kernel.Status]int      `json:"cases_by_status"`
}

type vectorSummary struct {
	CaseID         string        `json:"case_id"`
	SemanticID     string        `json:"semantic_id"`
	EdgeID         string        `json:"edge_id"`
	ExpectedStatus kernel.Status  `json:"expected_status"`
}

func countProofChoices(values []kernel.ProofChoice) map[kernel.Status]int {
	counts := map[kernel.Status]int{}
	for _, value := range values {
		counts[value.Status]++
	}
	return counts
}

func countIndicators(values []kernel.Indicator) map[kernel.Status]int {
	counts := map[kernel.Status]int{}
	for _, value := range values {
		counts[value.Status]++
	}
	return counts
}

func countCases(values []kernel.CaseSpec) map[kernel.Status]int {
	counts := map[kernel.Status]int{}
	for _, value := range values {
		counts[value.ExpectedStatus]++
	}
	return counts
}

func writeJSON(path string, value any) {
	raw, err := json.MarshalIndent(value, "", "  ")
	if err != nil {
		fatal(err)
	}
	raw = append(raw, '\n')
	if path == "" {
		if _, err := os.Stdout.Write(raw); err != nil {
			fatal(err)
		}
		return
	}
	if err := os.WriteFile(path, raw, 0o644); err != nil {
		fatal(err)
	}
}

func fatal(err error) {
	fmt.Fprintln(os.Stderr, err)
	os.Exit(1)
}
