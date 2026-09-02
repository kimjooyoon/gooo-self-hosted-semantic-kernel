#!/usr/bin/env bash
set -euo pipefail

repo_root="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
cd "$repo_root"

schema_path=".gooo/semantic.gooo"
corpus_path=".gooo/corpus.gooo"
evidence_dir="evidence"
mkdir -p "$evidence_dir/reference" "$evidence_dir/candidate" "$evidence_dir/comparisons"

go run ./cmd/gooo validate --schema "$schema_path" --corpus "$corpus_path" --output "$evidence_dir/validation.json"
go run ./cmd/gooo summary --schema "$schema_path" --corpus "$corpus_path" --output "$evidence_dir/summary.json"
go run ./cmd/gooo generate --schema "$schema_path" --corpus "$corpus_path" --output generated > "$evidence_dir/generation.json"
go build -o "$evidence_dir/gooo-generated.bin" ./generated

jq -e '
  .status == "CLOSED" and
  .entity_counts.cells == 12 and
  .entity_counts.meta_activities == 12 and
  .entity_counts.proof_choices == 12 and
  .entity_counts.indicators == 12 and
  .entity_counts.cases == 12 and
  .proof_choices_by_status.CLOSED == 4 and
  .proof_choices_by_status.UNKNOWN == 4 and
  .proof_choices_by_status.REFUTED == 4 and
  .indicators_by_status.CLOSED == 4 and
  .indicators_by_status.UNKNOWN == 4 and
  .indicators_by_status.REFUTED == 4 and
  .cases_by_status.CLOSED == 4 and
  .cases_by_status.UNKNOWN == 4 and
  .cases_by_status.REFUTED == 4 and
  .decision_precedence == ["REFUTED", "UNKNOWN", "CLOSED"] and
  .unknown_fields == ["stage", "step", "reason", "unknown_class", "next_operation", "blocked_by"] and
  .fixed_point_keyword == "FIXED_POINT"
' "$evidence_dir/validation.json"

jq -e '
  (.vectors | length) == 12 and
  ((.vectors | map(select(.expected_status == "CLOSED")) | length) == 4) and
  ((.vectors | map(select(.expected_status == "UNKNOWN")) | length) == 4) and
  ((.vectors | map(select(.expected_status == "REFUTED")) | length) == 4)
' "$evidence_dir/summary.json"

jq -r '.vectors[].case_id' "$evidence_dir/summary.json" > "$evidence_dir/case-ids.txt"
while IFS= read -r case_id; do
	go run ./cmd/gooo bootstrap --schema "$schema_path" --corpus "$corpus_path" --case-id "$case_id" --output "$evidence_dir/reference/$case_id.json"
	"$evidence_dir/gooo-generated.bin" --case-id "$case_id" --output "$evidence_dir/candidate/$case_id.json"
	expected="$(jq -r --arg case_id "$case_id" '.vectors[] | select(.case_id == $case_id) | .expected_status' "$evidence_dir/summary.json")"
	go run ./cmd/gooo compare --schema "$schema_path" --corpus "$corpus_path" --case-id "$case_id" --expected "$expected" --reference "$evidence_dir/reference/$case_id.json" --candidate "$evidence_dir/candidate/$case_id.json" --output "$evidence_dir/comparisons/$case_id.json"
	jq -e --arg case_id "$case_id" '
		.case_id == $case_id and .matched == true and .byte_equal == true and
		.semantic_identity_equal == true and .typed_value_equal == true and
		.ordered_effect_trace_equal == true and .terminal_digest_equal == true and
		.reference_terminal_digest_valid == true and .candidate_terminal_digest_valid == true and
		.verdict == .expected_status
	' "$evidence_dir/comparisons/$case_id.json"
done < "$evidence_dir/case-ids.txt"

go run ./cmd/gooo compare --schema "$schema_path" --corpus "$corpus_path" --case-id case-02-closed-effect --expected REFUTED --reference fixtures/counterexamples/reference.json --candidate fixtures/counterexamples/candidate.json --output "$evidence_dir/counterexample.json"
jq -e '.verdict == "REFUTED" and .matched == false and .ordered_effect_trace_equal == false and .terminal_digest_equal == false' "$evidence_dir/counterexample.json"

jq -s '[.[] | {case_id, expected_status, verdict, matched, byte_equal, reference_bytes_digest, candidate_bytes_digest, semantic_identity_equal, typed_value_equal, ordered_effect_trace_equal, terminal_digest_equal, reference_terminal_digest_valid, candidate_terminal_digest_valid, reference_status, candidate_status}]' "$evidence_dir/comparisons"/*.json > "$evidence_dir/vectors.json"
"$repo_root/scripts/inventory.sh" "$evidence_dir/inventory.json"

pr_number="0"
merge_commit=""
if [[ -n "${GITHUB_EVENT_PATH:-}" && -f "$GITHUB_EVENT_PATH" ]]; then
	pr_number="$(jq -r '.number // .pull_request.number // 0' "$GITHUB_EVENT_PATH")"
	merge_commit="$(jq -r '.pull_request.merge_commit_sha // ""' "$GITHUB_EVENT_PATH")"
fi
jq -n \
	--arg repository "${GITHUB_REPOSITORY:-local/unknown}" \
	--arg commit "${GITHUB_SHA:-unknown}" \
	--arg ref "${GITHUB_REF:-unknown}" \
	--arg workflow "${GITHUB_WORKFLOW:-PR authoritative conformance}" \
	--arg job "${GITHUB_JOB:-conformance}" \
	--arg run_id "${GITHUB_RUN_ID:-unknown}" \
	--arg run_attempt "${GITHUB_RUN_ATTEMPT:-unknown}" \
	--arg pr_number "$pr_number" \
	--arg merge_commit "$merge_commit" \
	--slurpfile validation "$evidence_dir/validation.json" \
	--slurpfile summary "$evidence_dir/summary.json" \
	--slurpfile vectors "$evidence_dir/vectors.json" \
	--slurpfile counterexample "$evidence_dir/counterexample.json" \
	--slurpfile inventory "$evidence_dir/inventory.json" \
	'{schema:"gooo.conformance/v1", authority:"pull_request_actions", status:"CLOSED", repository:$repository, commit:$commit, ref:$ref, pull_request:{number:$pr_number, merge_commit:$merge_commit}, run:{id:$run_id, attempt:$run_attempt, workflow:$workflow, job:$job}, artifact:{name:"gooo-conformance", path:"evidence/"}, validation:$validation[0], corpus_summary:$summary[0], vectors:$vectors[0], counterexample:$counterexample[0], inventory:$inventory[0], claims:{bounded_self_hosting:{status:"CLOSED", scope:"schema to generated evaluator to frozen oracle over fixed 12-case corpus"}, improvement:{status:"UNKNOWN", reason:"no before/after artifacts with the same digest identities"}, external_utility:{status:"UNKNOWN", reason:"outside the bounded corpus is not measured"}}}' > "$evidence_dir/conformance.json"

jq -e '.status == "CLOSED" and .claims.improvement.status == "UNKNOWN" and .claims.external_utility.status == "UNKNOWN" and (.vectors | length == 12) and .counterexample.verdict == "REFUTED"' "$evidence_dir/conformance.json"
