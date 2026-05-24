# migrations/

Directory where generated migration files live. Committed to git.

## File naming

Semver: `0.1.0.toml`, `0.2.0.toml`, `0.3.0.toml`. Sorted by semver (not lexicographic) for execution order.

## Lifecycle

1. User edits schema TOML files
2. `pgdesign migrate generate --db <url>` diffs desired vs live, writes a new migration file here
3. User reviews the file (or CI reviews it)
4. `pgdesign migrate apply --db <url>` applies pending files in semver order

## This directory is NOT the schema source of truth

The TOML schema files are the source of truth. Migration files are derived artifacts (the diff between schema versions). They exist for: audit trail, rollback capability, DML operations, and reproducibility.

## Version numbering

The version is always user-specified. When running `pgdesign migrate generate`, the tool prompts for the next version (or accepts `--version 0.2.0` flag). No auto-heuristic -- the user decides whether a change is patch or minor.

## Collision handling

If two developers generate migrations from the same baseline, they may pick the same version. This manifests as a git merge conflict on the filename. Resolution: dev B re-generates with the next available version. This is a workflow convention, not a tool-enforced constraint.

## State

Applied migrations tracked in `pgdesign_migrations` table in the target DB. This directory + that table = full migration history.
