# RFC: bounded self-hosted semantic kernel v0.1.0

## Decision

`.gooo/semantic.gooo` is the sole semantic authority for this bounded slice.
Go parses it and `.gooo/corpus.gooo`, generates a self-contained Go descriptor
and evaluator, and compares the generated evaluator with the handwritten
bootstrap oracle over a fixed corpus.

The path is intentionally bounded to nine declared operations, twelve cases,
and the entity cardinalities recorded by the schema. It is not general
self-hosting, does not claim completeness, and does not measure external
utility.

## Status and unknown semantics

Decision precedence is `REFUTED > UNKNOWN > CLOSED`. The join is applied after
all program observations, so a later fallback cannot close an UNKNOWN. A top
decision that remains UNKNOWN has no typed value and retains the six-field
record declared in the schema.

`FIXED_POINT` is explicit: a state is CLOSED only when canonical next state
equals prior state before `max_steps`; an unresolved bounded state is UNKNOWN;
a cycle or missing rule is REFUTED.

## Differential evidence

Each case output contains stable semantic and edge identities, the typed value,
ordered effect trace, and a terminal digest. CI compares canonical output
bytes, output byte digests, semantic identities, typed values, traces, and
terminal digests. The independent counterexample fixture must remain REFUTED.

## Operational policy

PR Actions are the authority. Merge follows green PR checks. Failed workflow
runs and incorrect predecessor releases are never deleted; preserve them as
`OPERATIONAL_REFUTED` evidence. Improvement remains UNKNOWN without matching
before/after artifacts carrying the same digest identities.
