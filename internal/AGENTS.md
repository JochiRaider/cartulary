# Backend Instructions

Supplement the [root procedure](../AGENTS.md) for backend work. Paths below are repository-relative.

- `internal/app/server`, `internal/app/migrate`, and `internal/app/operator` are the exact facades for matching `cmd/*` roots. The `internal/app` root is not a Go package.
- `internal/app/revisionassembly` aggregates source-owner revision contributions and injects platform dependencies. Source owners construct providers; Revisions validates the complete catalog and owns generic coordination.
- `internal/app/serverprocess` is process-level test evidence, not production assembly. Reusable application test composition belongs in `internal/testutil/appsupport`; reusable backend harnesses and fixtures belong in `internal/testutil/*`.
- Non-nil `server.Options.Postgres` and `server.Options.ObjectStore` are borrowed. The server closes only resources it creates, in reverse acquisition order; `Runtime.Close` is idempotent.
- Platform ownership includes configuration, storage adapters, auth primitives, and job shells. Modules own domain, application, and owner-specific supporting-subsystem logic.
- `internal/platform/postgres` owns the narrow `DB` port, settings/binding resolution, connectivity, and PostgreSQL telemetry—not a generic transaction lifecycle facade. Network Flow owns its private transaction helper over `postgres.DB`.
- `internal/modules/database_migrations` owns migration lifecycle mechanics over already-opened handles, not connectivity, secrets, generic query/transaction ports, source-owner schema meaning, or authored SQL. `db/migrations` and `db/queries` contain authored SQL inputs.
- Harness development under `internal/testutil` also uses the [tooling instructions](../tools/AGENTS.md).
