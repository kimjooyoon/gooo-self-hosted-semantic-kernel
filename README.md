# Gooo bounded self-hosted semantic kernel

This public repository implements one deliberately bounded self-hosting path:

```text
.gooo semantic schema
  -> generated Go descriptor/evaluator
  -> frozen bootstrap oracle
  -> fixed 12-case corpus
  -> byte and terminal-digest comparison
```

It is not a general self-hosting system and makes no completeness claim. The
only semantic authority for stable semantic IDs, edge IDs, decision precedence,
the six-field UNKNOWN record, and `FIXED_POINT` is
[`.gooo/semantic.gooo`](.gooo/semantic.gooo). The corpus is data interpreted by
that schema in [`.gooo/corpus.gooo`](.gooo/corpus.gooo).

## Bounded contract

The corpus fixes exactly 12 Cells, 12 MetaActivities, 12 ProofChoices, 12
Indicators, and 12 cases. Proof choices and indicators each have four CLOSED,
four UNKNOWN, and four REFUTED vectors. Cases have the same 4/4/4 split. The
decision order is `REFUTED > UNKNOWN > CLOSED`; UNKNOWN at the top decision is
fail-closed and never receives a default value.

The corpus includes a stable `FIXED_POINT` case, an unresolved fixed-point
case, a fixed-point cycle, and the `unknown-top-decision` fail-closed case.
Every UNKNOWN retains exactly these fields: `stage`, `step`, `reason`,
`unknown_class`, `next_operation`, and `blocked_by`.

The bootstrap oracle in `internal/bootstrap` is handwritten and does not import
the generated evaluator. The counterexample fixture in
`fixtures/counterexamples` proves that a changed effect trace or terminal
digest is reported as REFUTED even when the candidate status says CLOSED.

## Authority and verification

GitHub Actions on pull requests is the verification authority. The workflow
uses Go 1.27, runs tests, vet, formatting checks, `bash -n`, `jq` assertions,
and the conformance path. Local test/build/vet/fmt/actionlint/bash/jq/conformance
executions are intentionally not release evidence.

The conformance artifact records every vector, output byte digest, terminal
digest, inventory counts, wall time, and peak RSS. The root README is excluded
from inventory. Missing before/after artifacts with the same digest identities
leave improvement claims at `UNKNOWN`; external utility is also `UNKNOWN`.
Failed runs and invalid predecessor releases are retained under
`OPERATIONAL_REFUTED` instead of being deleted.

## Commands

```text
go run ./cmd/gooo validate --schema .gooo/semantic.gooo --corpus .gooo/corpus.gooo
go run ./cmd/gooo generate --schema .gooo/semantic.gooo --corpus .gooo/corpus.gooo --output generated
go run ./cmd/gooo bootstrap --schema .gooo/semantic.gooo --corpus .gooo/corpus.gooo --case-id case-01-closed-value
go run ./cmd/gooo compare --schema .gooo/semantic.gooo --case-id case-01-closed-value --expected CLOSED --reference reference.json --candidate candidate.json
```

The commands above describe the CI path. Do not treat local execution as
authority for this repository.

## Go 1.27 notes

The implementation targets the Go 1.27 module language version and CI toolchain.
The relevant primary sources are the [Go 1.27 release notes](https://go.dev/doc/go1.27),
the [Go 1.27 language specification](https://go.dev/ref/spec), and the
[official release history](https://go.dev/doc/devel/release).
