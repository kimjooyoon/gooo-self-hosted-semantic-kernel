# OPERATIONAL_REFUTED retention

This directory is the durable home for a failed execution or wrong
predecessor release when one is intentionally preserved in the repository
evidence bundle. Record the GitHub run or release identity, immutable commit
or tag, observed failure, and the next operation. Never rewrite or delete the
original artifact.

The first public release is retained here as
`v0.1.0-lineage.json`: its release object reported `immutable:false`, and its
asset was built from PR Actions run `33597238722` at head SHA
`3a04de58a2e8620ae454adf5853edbd542dd94b7`, not from the merged main commit.
