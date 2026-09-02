package generator

import (
	"bytes"
	_ "embed"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/kimjooyoon/gooo-self-hosted-semantic-kernel/internal/kernel"
)

// The evaluator and entrypoint templates are kept outside Go source so the
// generated program remains inspectable without sharing the oracle package.
//
//go:embed evaluator.tmpl
var generatedEvaluator string

//go:embed main.tmpl
var generatedMain string

func Generate(schema kernel.SemanticSchema, schemaRaw []byte, corpus kernel.Corpus, corpusRaw []byte, outputDir string) error {
	if err := schema.Validate(); err != nil {
		return err
	}
	if err := kernel.ValidateCorpus(schema, corpus); err != nil {
		return err
	}
	compactCorpus := bytes.Buffer{}
	if err := json.Compact(&compactCorpus, corpusRaw); err != nil {
		return fmt.Errorf("compact corpus for generated descriptor: %w", err)
	}
	if err := os.MkdirAll(outputDir, 0o755); err != nil {
		return err
	}
	evaluator := generatedEvaluator
	evaluator = strings.ReplaceAll(evaluator, "__SCHEMA_DIGEST__", goString(kernel.DigestBytes(schemaRaw)))
	evaluator = strings.ReplaceAll(evaluator, "__CORPUS_DIGEST__", goString(kernel.DigestBytes(corpusRaw)))
	evaluator = strings.ReplaceAll(evaluator, "__CORPUS_JSON__", goString(compactCorpus.String()))
	evaluator = strings.ReplaceAll(evaluator, "__TYPES__", stringSliceLiteral(schema.Types))
	evaluator = strings.ReplaceAll(evaluator, "__OPERATIONS__", stringSliceLiteral(operationNames(schema.Operations)))
	evaluator = strings.ReplaceAll(evaluator, "__COMPARISON_FIELDS__", stringSliceLiteral(schema.Decision.ComparisonFields))
	evaluator = strings.ReplaceAll(evaluator, "__IDENTITY_FIELDS__", stringSliceLiteral([]string{schema.Identity.SemanticID.Field, schema.Identity.EdgeID.Field}))
	evaluator = strings.ReplaceAll(evaluator, "__UNKNOWN_FIELDS__", stringSliceLiteral(schema.Unknown.Fields))
	evaluator = strings.ReplaceAll(evaluator, "__PRECEDENCE__", statusLiteral(schema.Decision.Precedence))
	evaluator = strings.ReplaceAll(evaluator, "__UNKNOWN_TEMPLATES__", unknownTemplatesLiteral(schema.Unknown.Templates))
	evaluator = strings.ReplaceAll(evaluator, "__FIXED_POINT_KEYWORD__", goString(schema.FixedPoint.Keyword))
	evaluator = strings.ReplaceAll(evaluator, "__FIXED_POINT_UNSTABLE__", goString(string(schema.FixedPoint.UnstableStatus)))
	evaluator = strings.ReplaceAll(evaluator, "__FIXED_POINT_CYCLE__", goString(string(schema.FixedPoint.CycleStatus)))
	if err := os.WriteFile(filepath.Join(outputDir, "evaluator.go"), []byte(evaluator), 0o644); err != nil {
		return err
	}
	if err := os.WriteFile(filepath.Join(outputDir, "main.go"), []byte(generatedMain), 0o644); err != nil {
		return err
	}
	return nil
}

func stringSliceLiteral(values []string) string {
	parts := make([]string, 0, len(values))
	for _, value := range values {
		parts = append(parts, goString(value))
	}
	return "[]string{" + strings.Join(parts, ", ") + "}"
}

func operationNames(operations map[string]kernel.OperationSchema) []string {
	names := make([]string, 0, len(operations))
	for name := range operations {
		names = append(names, name)
	}
	sort.Strings(names)
	return names
}

func goString(value string) string {
	raw, _ := json.Marshal(value)
	return string(raw)
}

func statusLiteral(statuses []kernel.Status) string {
	parts := make([]string, 0, len(statuses))
	for _, status := range statuses {
		parts = append(parts, "Status("+goString(string(status))+")")
	}
	return "[]Status{" + strings.Join(parts, ", ") + "}"
}

func unknownTemplatesLiteral(templates map[string]kernel.Unknown) string {
	keys := make([]string, 0, len(templates))
	for key := range templates {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	parts := make([]string, 0, len(keys))
	for _, key := range keys {
		template := templates[key]
		parts = append(parts, goString(key)+": {Stage: "+goString(template.Stage)+", Step: "+goString(template.Step)+", Reason: "+goString(template.Reason)+", UnknownClass: "+goString(template.UnknownClass)+", NextOperation: "+goString(template.NextOperation)+", BlockedBy: "+goString(template.BlockedBy)+"}")
	}
	return "map[string]Unknown{" + strings.Join(parts, ", ") + "}"
}
