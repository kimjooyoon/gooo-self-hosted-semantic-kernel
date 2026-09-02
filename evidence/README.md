# Evidence layout

GitHub Actions writes this directory during the authoritative pull-request
run. The generated files and JSON outcomes are not committed. The retained
artifact contains:

- `validation.json` and `summary.json`;
- one bootstrap and one generated outcome for every case;
- one differential comparison for every case;
- `counterexample.json` with verdict `REFUTED`;
- `inventory.json` excluding the root README, with actual counts, lines,
  generated-file count, wall milliseconds, and peak RSS KiB;
- `conformance.json` with PR, merge-ready, run, job, artifact, and digest
  identities.

If a workflow or release is invalid, preserve its identifiers and classify it
as `OPERATIONAL_REFUTED`. Do not delete it to make the evidence look green.
