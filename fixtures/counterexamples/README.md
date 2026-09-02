# Independent counterexample fixture

`reference.json` and `candidate.json` intentionally describe the same CLOSED
case but differ in the ordered effect trace and terminal digest. Their
synthetic terminal digests are deliberately not trusted. The comparator must
return `REFUTED`; a candidate status of CLOSED cannot hide a semantic
divergence.
