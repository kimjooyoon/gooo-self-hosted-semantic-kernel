# Release v0.1.0 evidence contract

The release is created draft-first from the merged `main` commit only after a
green pull-request Actions run. The release tag is created once and verified
before publication. The release assets are the conformance evidence bundle,
including all twelve vectors, byte/digest comparisons, generated source, and
inventory.

No failed run or invalid predecessor release is deleted. Such an artifact is
retained under the `OPERATIONAL_REFUTED` label with its run or tag identity.

The release does not claim general self-hosting, completeness, aggregate
quality, or external utility. Those claims are outside the bounded evidence;
utility and improvement without same-identity before/after material remain
`UNKNOWN`.
