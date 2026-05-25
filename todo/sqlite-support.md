# SQLite support

## Context

A downstream project uses SQLite (WAL mode, per-tenant databases) for all data storage. The schema has 15 tables with foreign keys, indexes, and constraints. Currently the schema is managed via hand-written DDL in Python and a custom forward-only migration system (schema_version table + sequential DDL functions).

pgdesign's TOML → schema → migration flow would be ideal for managing this schema declaratively, but pgdesign currently only targets PostgreSQL.

## Question

Can pgdesign's architecture support SQLite as a second target, or is it fundamentally Postgres-specific?

## Analysis of compatibility

**What would work as-is:**
- TOML schema definition (tables, columns, types, constraints, comments)
- Dependency ordering (topological sort for CREATE TABLE)
- Canonical formatting (`pgdesign fmt`)
- Validation rules (naming conventions, NOT NULL by default, FK requires on_delete)
- D2 diagram generation (schema visualization)

**What would need adaptation:**
- DDL generation: SQLite DDL syntax differs (no `CREATE SCHEMA`, no `ALTER TABLE DROP COLUMN`, limited `ALTER TABLE ADD COLUMN`, no `CREATE INDEX CONCURRENTLY`)
- Type system: SQLite has 5 types (NULL, INTEGER, REAL, TEXT, BLOB) vs Postgres's rich type system. The semantic type system would need a SQLite type mapping layer.
- Migration safety: SQLite doesn't support transactional DDL for all operations. Column renames/drops require table rebuild (CREATE new → copy → DROP old → RENAME).
- Introspection: SQLite uses `sqlite_master` / `PRAGMA table_info` instead of `pg_catalog`. Different IR construction path.
- Extensions: not applicable to SQLite (no opclass, no custom types beyond column affinities).
- Normal form auditing: works at the logical level, should be engine-independent.
- FD discovery (TANE): engine-independent algorithm, only needs row access.

**What would NOT work:**
- Schema-per-tenant isolation (SQLite has no schemas — isolation is via separate database files)
- Advisory locks (SQLite uses file locks)
- Non-transactional DDL detection (not relevant for SQLite)
- pg_version conditional syntax

## Options

**Option A: SQLite as a first-class target alongside Postgres.**
Add `engine = "sqlite"` to `pgdesign.toml`. Shared TOML schema, type mapping layer, separate DDL generators. Migrations generated for the target engine. Biggest effort but most useful.

**Option B: TOML schema → SQLite DDL export only (no migrations).**
pgdesign generates CREATE TABLE statements for SQLite from the same TOML files. No migration support, no introspection. The downstream project keeps its own migration system. Moderate effort, useful for schema validation and visualization.

**Option C: Out of scope.**
pgdesign stays Postgres-only. The downstream project continues managing SQLite schemas manually, or builds its own lightweight TOML → SQLite tool.

## Effort estimate

- Option A: Large (3-4 weeks). New DDL generator, type mapper, introspection backend, migration adapter, test suite.
- Option B: Medium (1 week). DDL generator + type mapper only.
- Option C: Zero.

## Downstream project details

- 15 tables, ~100 columns, 12 foreign keys
- WAL mode, per-tenant databases (100s-1000s of SQLite files)
- Forward-only migrations (no rollback)
- Python application, schema defined in Python strings today
- Would benefit most from: TOML schema definition, validation, diagram generation, migration generation
