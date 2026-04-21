# Cartulary Development Guide

## Status

- **Document class:** non-normative implementation-support guide.
- **Authority:** subordinate to Cartulary Normative Core 00 through Core 04 for implementation-conformance behavior, subordinate to Cartulary Normative Companion 05 for claim-bearing timed or fixture-sensitive publication behavior, and subordinate to any later adopted Cartulary NLSpec.
- **Intended audience:** implementers and maintainers of the Cartulary Go backend, TypeScript frontend, contracts, build pipeline, and deployment packaging.
- **Scope:** repo-local implementation baseline, contract derivation rules, module ownership, workspace structure, development workflow, release verification, and implementation-facing summaries of behavior already owned by the normative core.
- **Out of scope:** ownership of product behavior already defined in Core 00 through Core 04, ownership of claim-bearing timed or fixture-sensitive publication behavior defined in Core 05, contributor operating procedure delegated to the repo procedure owner (expected to be `AGENTS.md` when present), and ADR-style rationale beyond what is necessary to keep this guide operationally usable.

This guide MAY use **MUST**, **MUST NOT**, **SHOULD**, **SHOULD NOT**, and **MAY**. In this guide, those words govern the repository implementation baseline, artifact derivation rules, and guide interpretation only. They MUST NOT be read as widening, narrowing, or replacing product behavior already owned by the normative core.

When this guide and a primary owner section differ, the owner section governs. Behavior-affecting changes MUST land in the owner section first, then in the derived repo-local contract artifacts, then in generated code, then in this guide if the implementation baseline or developer workflow changed.

Implementation conformance and claim-bearing publication are separate. Nothing in this guide widens Base-profile or extension-profile implementation conformance through Core 05.

## Guide interpretation

### Section-class legend

Every major section in this guide declares one section class unless a subsection overrides it.

| Section class                    | Meaning                                                                            |
| -------------------------------- | ---------------------------------------------------------------------------------- |
| `Implementation baseline`        | Repo-local language, package, directory, command, and ownership choices.           |
| `Derived behavioral summary`     | A non-authoritative summary of product behavior owned elsewhere.                   |
| `Generated-artifact rule`        | Code-generation, drift, or hand-edit prohibition rules.                            |
| `Cross-reference`                | Navigation to the controlling owner sections.                                      |
| `Contributor-procedure boundary` | A pointer to `AGENTS.md` or another procedure owner, not procedure content itself. |

### Profile applicability matrix

Unless a narrower banner is stated for a subsection, this guide is written for the Base profile implementation baseline.

| Profile                                                 | Default in this guide                      | Implementation-guide meaning                                                                                                    |
| ------------------------------------------------------- | ------------------------------------------ | ------------------------------------------------------------------------------------------------------------------------------- |
| Base (`base`)                                           | Assumed unless a subsection says otherwise | Default architecture, repo layout, concurrency, evidence core, clipboard paste, local authentication, and baseline deployment.  |
| Import (`import`)                                       | Off unless explicitly claimed              | File-based structured import, import assistant objects, import-session persistence, and workbook parser compatibility behavior. |
| Snapshot and Reporting (`snapshot_reporting`)           | Off unless explicitly claimed              | Immutable snapshots, release records, report rendering, export redaction, and generated presentation artifacts.                 |
| Incident Portability (`incident_portability`)           | Off unless explicitly claimed              | Whole-incident export and import, bundle verification, and portability staging.                                                 |
| Reference Pack (`reference_pack`)                       | Off unless explicitly claimed              | Pack import, activation, refresh, attestation, and overlay behavior.                                                            |
| Enterprise Authentication (`enterprise_authentication`) | Off unless explicitly claimed              | OIDC and SAML provider integration.                                                                                             |

### Terminology used by this guide

| Term                                    | Required meaning in this guide                                                                                                                               |
| --------------------------------------- | ------------------------------------------------------------------------------------------------------------------------------------------------------------ |
| `implementation guide`                  | This document. It is not a sole behavioral authority.                                                                                                        |
| `repo-local derived contract artifacts` | Machine-readable contract files under `/contracts/*` that are derived from owner documents and drive code generation.                                        |
| `artifact`                              | The record family defined by Core 02. It is not synonymous with the Notes workbook surface.                                                                  |
| `Notes`                                 | One workbook surface backed by `artifact_type='note'`.                                                                                                       |
| `system view`                           | A workbook-native surface beyond the five built-in tabs.                                                                                                     |
| `party`                                 | An incident-scoped coordination identity, not a deployment-local login user.                                                                                 |
| `local authentication`                  | The Base-profile session and MFA implementation.                                                                                                             |
| `enterprise authentication`             | The Enterprise Authentication Extension Profile only.                                                                                                        |
| `import unit`                           | The canonical file-import contract object. Explanatory references to worksheet ranges or tables are examples of locator kinds, not alternate contract names. |

### Owner cross-reference map

| Topic family                                                                                                          | Primary owner |
| --------------------------------------------------------------------------------------------------------------------- | ------------- |
| Document status, precedence, profile model                                                                            | Core 00       |
| Architecture, route families, view schemas, projections, jobs, portability, extension route families                  | Core 01       |
| Record model, mention/entity semantics, parties, task requests, decisions, indicators, assessments, history substrate | Core 02       |
| Workbook surfaces, saved views, startup selection, collaboration, same-field conflicts, workflow behavior             | Core 03       |
| Authentication, authorization, runtime roots, trust boundaries, conformance fixtures                                  | Core 04       |
| Claim-bearing publication, benchmark profiles, benchmark manifests, measurement predicates, and reproducibility       | Core 05       |

---

## 1. Architecture summary

**Section class:** `Implementation baseline`  
**Profile applicability:** Base unless a subsection says otherwise.  
**Owner references:** Core 01 §1 through §3, Core 03 §1 through §2, Core 04 §5 through §7 and §12.

### 1.1 Deployable shape

The implementation baseline is a **modular monolith** deployed as **one application unit** plus two authoritative backing services:

- one application deployable that serves the browser UI, HTTP+JSON API, WebSocket collaboration stream, and background-job workers,
- one PostgreSQL service for authoritative structured state,
- one S3-compatible object storage service for authoritative binary evidence state.

The application deployable MUST remain one deployable unit in disconnected, on-prem, and cloud variants. Internal module boundaries MUST remain logical boundaries inside that unit rather than separate runtime services.

### 1.2 Human-facing runtime surface

The browser client is the primary interaction surface. It MUST preserve the workbook mental model at the view layer through:

- a virtualized workbook grid,
- inline editing,
- keyboard navigation,
- clipboard paste,
- low-friction row creation,
- a collapsible inspector for enrichment, relationships, history, and destructive actions,
- real-time presence and row updates,
- save-state and conflict-state presentation.

The implementation baseline uses the versioned public boundaries already owned by the normative core: `/api/v1/*` for HTTP+JSON and `/ws/v1/*` for the bounded collaboration stream.

### 1.3 Complexity budget

The design center is mutation semantics, projection maintenance, collaboration UX, import isolation, evidence handling, and same-field conflict behavior. The implementation baseline MUST spend complexity budget there rather than on early microservice decomposition.

A future split beyond the modular monolith baseline MUST be treated as an architecture change, not as an incidental refactor. Such a change MUST be authorized first in the owner architecture documents and SHOULD be accompanied by an ADR or equivalent governance artifact.

### Verification

- The repository build and packaging flow produce exactly one application artifact plus separate Postgres and object-storage service dependencies.
- No preview or download flow requires browser-constructed raw object-store access. Evidence upload MAY use a server-issued short-lived upload target under the object-store abstraction, but supported browser evidence access remains application-mediated.
- No repo-local document claims that this guide is the sole authority for product behavior.

---

## 2. Technology stack baseline and manifest verification boundary

**Section class:** `Implementation baseline`  
**Profile applicability:** Base  
**Owner references:** Core 01 §1 through §3, Core 04 §1, §4.4, §4.5, and §12.

### Pending repo-control revalidation

This section remains an intended stack baseline, not a manifest-verified current-repository inventory. Exact package presence, replacement, and version truth remain owned by `go.mod`, `go.sum`, `package.json`, workspace manifests, and lockfiles. Until those repo-control files are revalidated against the live repository, treat the package tables below as intended baseline only.

The planned stack families for the current implementation baseline are Go for the backend, React for the browser UI, and Vite for frontend development and bundling. The package names listed in this section are planned baseline choices, not an independently verified current-repository inventory. Exact package presence, replacement, and version truth are owned by `go.mod`, `go.sum`, `package.json`, workspace manifests, and lockfiles. In this guide, exact package-inventory assertions are current only to the extent those repo-control files are revalidated as of 2026-04-15.

| Statement family                              | Status in this guide    | Verification source                                                                                 | Default interpretation when repo-control files are not under review |
| --------------------------------------------- | ----------------------- | --------------------------------------------------------------------------------------------------- | ------------------------------------------------------------------- |
| Planned stack family                          | Implementation baseline | This guide                                                                                          | Authoritative guide baseline                                        |
| Exact direct dependency and version inventory | Repo-control fact       | `go.mod`, `go.sum`, `package.json`, workspace manifests, and lockfiles revalidated as of 2026-04-15 | Intended baseline only; not a verified current-repository claim     |

A guide revision that changes package assertions SHOULD update the package-inventory freshness anchor and revalidate the table against the repo-control files.

### 2.1 Backend runtime dependencies

All direct backend runtime dependencies MUST be permissive-licensed. Everything not listed below is Go standard library.

| Package                        | Purpose                                                                      | License      |
| ------------------------------ | ---------------------------------------------------------------------------- | ------------ |
| `github.com/jackc/pgx/v5`      | PostgreSQL driver, typed query execution, pooling, LISTEN/NOTIFY integration | MIT          |
| `github.com/coder/websocket`   | WebSocket server for the collaboration stream                                | ISC          |
| `github.com/minio/minio-go/v7` | S3-compatible object storage client                                          | Apache-2.0   |
| `github.com/BurntSushi/toml`   | Deployment configuration loading with fail-closed unknown-key validation     | MIT          |
| `golang.org/x/crypto`          | Argon2id password hashing and related primitives                             | BSD-3-Clause |
| `github.com/pquerna/otp`       | TOTP MFA                                                                     | Apache-2.0   |

Standard-library responsibilities include:

- HTTP routing via `net/http`,
- JSON encoding via `encoding/json`,
- cryptographic primitives via `crypto/*`,
- structured logging via `log/slog`,
- context propagation via `context`,
- concurrency primitives via `sync`,
- embedded frontend assets via `embed`,
- server-side shell templates via `html/template`,
- test harness utilities via `testing` and `net/http/httptest`.

The baseline MUST NOT introduce a Go web framework or ORM unless the repository intentionally changes the stack baseline. No `Gin`, `Echo`, `Chi`, `Fiber`, `GORM`, or `ent` dependency is part of the current implementation baseline.

### 2.2 Backend dev-only tools

These tools do not ship in the compiled application binary.

| Tool                | Purpose                                               |
| ------------------- | ----------------------------------------------------- |
| `sqlc`              | Generate typed Go code from authored SQL queries      |
| `goose`             | Apply numbered SQL migrations                         |
| `testcontainers-go` | Run integration tests against real Postgres and MinIO |

### 2.3 Frontend runtime dependencies

| Package           | Purpose                                                                            | License |
| ----------------- | ---------------------------------------------------------------------------------- | ------- |
| `react`           | UI runtime and declarative rendering                                               | MIT     |
| `react-dom`       | React DOM renderer                                                                 | MIT     |
| `react-data-grid` | Virtualized workbook grid, keyboard editing, paste, grouping, and custom renderers | MIT     |

### 2.3.1 Grid engine selection gate

The current baseline grid package remains `react-data-grid` until changed by an explicit stack-baseline revision. 

A pull request MUST NOT introduce a second runtime grid engine, replace the baseline grid engine, adopt a license-keyed runtime grid engine, or expose a selected-grid vendor instance outside the grid adapter unless the grid-engine selection gate is complete.

A grid-engine evaluation is complete only when every field below is present. Omission of any required field means `Evaluation incomplete`.

| Evaluation field          | Required content                                                                                                                            | Default when absent    |
| ------------------------- | ------------------------------------------------------------------------------------------------------------------------------------------- | ---------------------- |
| Dependency identity       | Package name, package scope, major version, package manager, and intended import path.                                                      | Evaluation incomplete. |
| License posture           | License name, commercial or evaluation terms, license-key requirement, and whether the dependency satisfies the permissive-runtime policy.  | Evaluation incomplete. |
| Runtime model             | Exactly one of `declarative_react_grid`, `imperative_engine_with_react_adapter`, or `other`.                                                | Evaluation incomplete. |
| Row identity model        | Whether source, physical, visual, and renderable indexes are distinct; how the engine exposes each.                                         | Evaluation incomplete. |
| Data ownership model      | Whether the engine mutates caller data, clones caller data, owns internal source data, or supports more than one mode.                      | Evaluation incomplete. |
| Extension surface         | Supported plugins, hooks, registries, renderers, editors, validators, menus, formulas, clipboard, sorting, filtering, grouping, and themes. | Evaluation incomplete. |
| Disabled-capability plan  | Closed list of vendor capabilities that are disabled, unavailable, or inaccessible through ordinary runtime.                                | Evaluation incomplete. |
| Adapter preservation plan | How `record_id`, `row_version`, `field_key`, conflict state, pending queue state, and server-canonical query state remain authoritative.    | Evaluation incomplete. |
| Test plan                 | Unit, integration, E2E, visual-regression, and performance evidence required before adoption.                                               | Evaluation incomplete. |

An ADR or equivalent rationale artifact MAY explain why a grid-engine change is desirable. It does not satisfy this gate by itself. The Development Guide MUST still record the stack-baseline consequence before the repository treats the dependency change as adopted.

#### Acceptance criteria

- **DG-R08-AC-001:** This subsection exists directly under `### 2.3 Frontend runtime dependencies`.
- **DG-R08-AC-002:** The subsection states that `react-data-grid` remains the baseline unless changed by explicit stack-baseline revision.
- **DG-R08-AC-003:** The subsection requires license-key and license-policy review before adopting a license-keyed grid runtime.
- **DG-R08-AC-004:** The subsection contains an evaluation-field table covering dependency identity, license posture, runtime model, row identity model, data ownership model, extension surface, disabled-capability plan, adapter preservation plan, and test plan.

The rest of the baseline frontend runtime uses browser-native APIs:

- `fetch` for HTTP,
- `WebSocket` for the collaboration stream,
- Clipboard APIs for paste,
- `KeyboardEvent` for shortcuts,
- `IntersectionObserver` and `ResizeObserver` for viewport management.

For evidence access, preview and download MUST continue to redeem through the Go application via same-origin opaque handles. The browser MUST NOT construct, persist, or treat arbitrary raw object-store URLs as evidence-access state. Evidence upload MAY use a server-issued short-lived upload target under the object-store abstraction.

### 2.4 Frontend dev and build tools

| Tool                               | Purpose                                            |
| ---------------------------------- | -------------------------------------------------- |
| `typescript`                       | Static type checking                               |
| `vite` plus `@vitejs/plugin-react` | Dev server and production bundler                  |
| `biome`                            | Formatting and linting                             |
| `vitest`                           | Unit, integration, protocol, and component testing |
| `@testing-library/react`           | Component test utilities                           |
| `@playwright/test`                 | Multi-context end-to-end testing                   |

### 2.5 Explicit exclusions

The current baseline excludes the following classes of dependency:

- Go web frameworks instead of the standard-library router.
- Go ORMs instead of authored SQL plus `sqlc`.
- SSR-oriented frontend frameworks such as `Next.js`.
- Commercial grid dependencies such as `AG Grid Enterprise`.
- Generic client-state libraries on day one if they would own Cartulary-specific conflict, presence, or pending-patch semantics.
- Monorepo frameworks such as `Nx`, `Turborepo`, `Bazel`, or `Pants`.

A stack-baseline change in any of those classes SHOULD be treated as a controlled repo-baseline change rather than as an incidental pull request.

### 2.6 Licensing, SBOM, and release verification

Every direct runtime dependency MUST be permissive-licensed. The release process MUST additionally produce:

- a direct and transitive dependency license report,
- a software bill of materials for the shipped application artifact set.

Developer verification and release verification are different tiers. `make check` is the developer verification gate. Release verification MUST run the developer gate plus license and SBOM checks. This guide does not assign a canonical user-facing release-gate command name unless and until the repo control surface is verified in repository files; the minimum requirement is a release gate, whether CI job or scripted task, that runs the full release-verification tier.

### Verification

- Exact package-inventory checks for this section are valid only when compared against `go.mod`, `go.sum`, `package.json`, workspace manifests, and lockfiles current as of 2026-04-15; absent that verification, the package tables above are interpreted as planned baseline choices.
- No prohibited license class appears in the shipped dependency graph.
- A release build produces a license report and an SBOM in addition to passing developer verification.

---

## 3. Monorepo structure

**Section class:** `Implementation baseline`  
**Profile applicability:** Base  
**Owner references:** Core 01 §2 through §3, Core 04 §6 and §12.

### Pending repo-control revalidation

This section remains an intended monorepo baseline, not a verified live-repository inventory. Until the live tree and repo-control files are revalidated, exact paths, filenames, generated-path locations, and root control surfaces below are intended baseline only.

The intended repository baseline is a polyglot modular-monolith monorepo with three implementation layers:

1. a Go application,
2. a TypeScript browser application,
3. a repo-local contract layer that keeps the Go and TypeScript sides aligned.

Business logic is not shared across Go and TypeScript. Shared structure exists through contracts, generated types, and stable route and view-schema identifiers.

The path tree below is an intended baseline shape, not an independently verified current-repository inventory. Treat exact filenames, path presence, and repo-control surfaces as current-state facts only after verification against the live repository.

```text
/
  AGENTS.md                          # Contributor and coding-agent procedure; not owned by this guide
  README.md
  Makefile                           # Repo-wide task surface
  docker-compose.dev.yml             # Local Postgres + MinIO only
  .editorconfig
  .env.example

  go.mod
  go.sum

  package.json
  pnpm-workspace.yaml
  pnpm-lock.yaml
  biome.json
  tsconfig.base.json

  /cmd
    /server                          # App entry point: UI, API, WS, jobs
    /migrate                         # Migration runner

  /internal
    /app                             # Composition root
    /platform
      /httpapi                       # Transport, middleware, envelopes, pagination
      /ws                            # WebSocket upgrade and stream lifecycle
      /jobs                          # Background-job runner
      /postgres                      # pgx pool and transaction helpers
      /objectstore                   # S3-compatible storage adapter
      /authn                         # Password, MFA, session implementation
      /config                        # Config loading, validation, runtime roots
    /modules
      /auth                          # Local users, sessions, MFA, deployment-local account concerns
      /incidents                     # Incidents, memberships, saved views, workbook preferences
      /timeline                      # Timeline capture and lifecycle behavior
      /entities                      # Hosts, identities, indicators, assessments, mentions, aliases
      /evidence                      # Evidence records, blob attachment, custody, access handles
      /imports                       # File-based import and shared tabular ingest
      /links                         # Links, tags, notes, parties, tasks, decisions, coordination artifacts
      /revisions                     # Change sets, history, rollback
      /projections                   # Projection maintenance and rebuilds
      /reference_data                # Reference packs and registries
      /reporting                     # Snapshots, releases, export models, renders
      /collaboration                 # Presence, replayable events, resume handling
    /gen
      /contracts                     # Generated Go code derived from /contracts
      /sql                           # Generated Go code derived from /db/queries

  /db
    /migrations                      # Numbered SQL migrations
    /queries                         # Authored SQL consumed by sqlc

  /contracts                         # Repo-local derived contract artifacts
    /openapi
      cartulary.openapi.yaml
    /ws
      *.schema.json
    /view-schemas
      *.json
    /errors
      *.json

  /apps
    /web                             # React application served by the app deployable

  /packages
    /ui                              # Reusable presentational components
    /protocol-ts                     # Generated TypeScript protocol types
    /view-contracts                  # TypeScript-consumable adapters around view-schema contracts
    /test-utils                      # Shared frontend and protocol test helpers

  /scripts                           # CI and helper scripts
  /tools                             # Pinned auxiliary Go tools
  /docs                              # Design notes, ADRs, threat model, and related support docs
```

### 3.1 Structural rules

- The backend MUST remain one root Go module.
- The TypeScript workspace MUST remain one `pnpm` workspace rooted at the repository top level.
- `internal/platform` owns transport, configuration, storage adapters, and cross-cutting runtime plumbing.
- `internal/modules` owns domain and application logic.
- Generated code MAY be checked in, but generated paths MUST be clearly marked and drift-checked.
- `AGENTS.md` owns contributor procedure. This guide may reference commands and paths, but it does not own execution procedure.

### 3.2 Authoritative-input versus generated-path policy

| Path family                              | Path status                         | Authoritative input                 | Edit policy                      |
| ---------------------------------------- | ----------------------------------- | ----------------------------------- | -------------------------------- |
| `/contracts/**`                          | Authored derived contract artifacts | Owner core docs and adopted NLSpecs | Hand-editable after owner change |
| `/db/queries/**`                         | Authored source                     | Repo-local query definitions        | Hand-editable                    |
| `/db/migrations/**`                      | Authored source                     | Schema evolution                    | Hand-editable                    |
| `/internal/gen/**`                       | Generated                           | `/contracts/**` or `/db/queries/**` | MUST NOT be hand-edited          |
| `/packages/protocol-ts/src/generated/**` | Generated                           | `/contracts/**`                     | MUST NOT be hand-edited          |
| `pnpm-lock.yaml`                         | Tool-managed                        | `pnpm install`                      | MUST NOT be hand-edited          |
| `go.sum`                                 | Tool-managed                        | `go mod tidy` or equivalent         | MUST NOT be hand-edited          |

### Verification

- The repository tree contains the path families declared above or explicit successor paths documented by a guide update.
- Generated paths contain `DO NOT EDIT` markers when checked in.
- No domain logic is introduced directly under transport packages.

---

## 4. Contract layer and code generation

**Section class:** `Generated-artifact rule`  
**Profile applicability:** Base plus any claimed extension profiles.  
**Owner references:** Core 00 §5.1, Core 01 §3.3, §7.4, §17, §18A, §18B, §19, §20, Core 02 §6 through §8 and §18 through §19, Core 03 §2 through §4.

### 4.1 Derivation model

The contract layer is a **repo-local derived artifact layer**, not the primary behavioral authority.

The required derivation chain is:

`owner section in Core or adopted NLSpec` → `/contracts/*` repo-local machine-readable artifact → generated Go and TypeScript code → runtime consumers

The guide MUST NOT refer to `/contracts/*` as the product “source of truth.” The only acceptable interpretation in this guide is that `/contracts/*` are the machine-readable contract artifacts from which the repository derives transport types, view adapters, validation helpers, and drift checks.

### 4.2 Contract family owner map

| Contract family                             | Primary owner                                                                                                      | Repo-local artifact                         | Generated or runtime consumers                                           | Edit rule                                          |
| ------------------------------------------- | ------------------------------------------------------------------------------------------------------------------ | ------------------------------------------- | ------------------------------------------------------------------------ | -------------------------------------------------- |
| HTTP route surface                          | Core 01 route-owner sections                                                                                       | `/contracts/openapi/cartulary.openapi.yaml` | Go transport types, handler tests, TypeScript request and response types | Change owner contract first, then OpenAPI          |
| WebSocket messages                          | Core 01 collaboration transport plus Core 03 collaboration behavior                                                | `/contracts/ws/*.schema.json`               | Go WebSocket transport, TypeScript message types, protocol tests         | Change owner contract first, then WS schemas       |
| View schemas and write-back contracts       | Core 01 view-schema registry and addenda, with field semantics from Core 02 and workflow consequences from Core 03 | `/contracts/view-schemas/*.json`            | Grid wrapper, filter builders, write routing, saved-view normalization   | Change owner contract first, then view schemas     |
| Error registries and reason-code registries | Core 01 error-envelope owner sections                                                                              | `/contracts/errors/*.json`                  | Go error helpers, TypeScript enums, test fixtures                        | Change owner contract first, then error registries |

### 4.3 Generated artifact rules

The following paths are generated and MUST NOT be hand-edited:

- `/internal/gen/**`
- `/packages/protocol-ts/src/generated/**`

`/packages/view-contracts` MUST remain downstream of `/contracts/view-schemas/*`. It MAY contain handwritten adapters, helpers, and package metadata, but it MUST NOT become an independent behavioral owner for field mutability, grouping, filter semantics, or write-back semantics.

### 4.4 Drift policy

Contract drift is a build failure.

At minimum, drift failure means one or more of the following:

- `make generate` changes tracked generated files,
- a view-schema change is not propagated into TypeScript- or Go-consumable outputs,
- a route or error-registry change is not propagated into generated protocol types,
- a code change requires a contract update but leaves `/contracts/*` stale.

The implementation sequence for a behavior-affecting change MUST be:

1. update the owner section in the normative core or adopted NLSpec,
2. update `/contracts/*`,
3. regenerate generated artifacts,
4. update Go and TypeScript implementation code,
5. update repo-local tests and this guide if the implementation baseline changed.

### Verification

- `make generate` followed by a clean `git diff` is part of developer verification.
- No pull request hand-edits generated paths.
- Every change to `/contracts/*` is traceable to an owner-section change or a repo-local representational refinement that does not alter behavior.

---

## 5. Backend architecture

**Section class:** `Implementation baseline`  
**Profile applicability:** Base unless a subsection says otherwise.  
**Owner references:** Core 01 §2 through §3, Core 02 domain-owner sections, Core 04 §1 through §3.

### 5.1 Composition root

`/cmd/server` starts the application unit. It wires platform services into module constructors, registers HTTP routes, installs the WebSocket handler, starts background jobs, and serves embedded frontend assets.

`/cmd/migrate` is the migration entry point. Schema DDL changes MUST move through `/db/migrations/*` and MUST NOT be embedded ad hoc inside application startup.

### 5.2 Platform layer

`internal/platform/httpapi` owns:

- `net/http` route registration,
- auth and CSRF middleware,
- request identifiers,
- structured logging,
- panic recovery,
- the common success and error envelopes,
- cursor-pagination helpers.

`internal/platform/ws` owns:

- WebSocket upgrade and origin validation,
- hello and resume handshake wiring,
- stream sequencing,
- replay-window enforcement,
- heartbeat handling,
- subscriber lifecycle and cleanup.

`internal/platform/postgres` owns:

- `pgx/v5` pool creation,
- transaction helpers,
- query-execution helpers shared across modules.

`internal/platform/objectstore` owns:

- S3-compatible storage configuration,
- upload-target creation,
- object fetch and range-read helpers,
- preview and download redeem helpers.

`internal/platform/authn` owns:

- password hashing,
- TOTP verification,
- session creation and revocation,
- session-expiry enforcement,
- the session inspection resource shape used by the UI.

`internal/platform/config` owns:

- config-file and environment loading,
- runtime-root resolution,
- unknown-key rejection,
- typed configuration binding.

`internal/platform/jobs` owns:

- common job-resource shape,
- background execution,
- status updates,
- cancellation,
- post-terminal retention.

### 5.3 Domain-module ownership baseline

| Module           | Authoritative implementation scope                                                                                                                           | Explicit non-scope                                                       |
| ---------------- | ------------------------------------------------------------------------------------------------------------------------------------------------------------ | ------------------------------------------------------------------------ |
| `auth`           | Local users, sessions, MFA, deployment-local user-account concerns                                                                                           | Incident data, workbook rows, evidence lifecycle                         |
| `incidents`      | Incident CRUD, memberships, saved views, workbook preferences, startup-surface persistence                                                                   | Password/MFA logic, blob lifecycle                                       |
| `timeline`       | `timeline_event` creation, update, capture-state transitions, timeline-specific projection effects                                                           | Canonical host or identity ownership outside mention workflows           |
| `entities`       | Hosts, identities, canonical indicators, indicator observations, indicator lifecycle intervals, compromise assessments, aliases, entity mentions             | Parties, task requests, decisions, artifact-backed coordination surfaces |
| `evidence`       | Evidence records, blob-slot lifecycle, attach-blob finalization, custody events, preview/download handle issuance and redemption                             | Saved views, reference-pack activation                                   |
| `imports`        | CSV and XLSX adapters, `import_session`, `import_unit`, preview, mapping, provenance, warning codes, import jobs                                             | Hot-path clipboard UI beyond the shared tabular-ingest contract          |
| `links`          | Typed record links, tags, Notes artifacts, `party`, `task_request`, `decision`, `comm_log`, `handoff`, `status_review`, `lesson`, coordination linking flows | Object storage internals, pack verification                              |
| `revisions`      | Change sets, mutation entries, record revisions, rollback operations                                                                                         | Live presence and WS transport                                           |
| `projections`    | `*_grid_projection` maintenance, rebuild commands, projection invalidation strategy                                                                          | Source-of-truth business decisions                                       |
| `reference_data` | Reference packs, manifests, activation state, attestation metadata, type registries                                                                          | Incident record lifecycle                                                |
| `reporting`      | Snapshots, releases, canonical export model, self-contained render artifacts                                                                                 | Live workbook write path                                                 |
| `collaboration`  | Presence state, `record_changed`, `job_progress`, replay and resume handling                                                                                 | HTTP mutation acceptance and persistence                                 |

This guide intentionally places `party`, `task_request`, `decision`, and artifact-backed coordination surfaces under the existing `links` module baseline because the current proposed module set does not define a separate `coordination` module. A future split into a dedicated coordination module MAY happen later, but it is not part of the current baseline.

### 5.4 Persistence and query rules

- `/db/queries/*.sql` are authored SQL source.
- `sqlc` generates typed Go into `/internal/gen/sql`.
- `/db/migrations/*.sql` are numbered migrations applied by `goose`.
- Projection tables are disposable caches and MUST be rebuildable from authoritative source state.
- Every mutable user-visible record MUST be addressed by stable `record_id` plus `row_version`. No backend API may rely on visible row position or display labels for mutation targeting.

### Verification

- Module directories and responsibilities match the ownership table.
- No XLSX or workbook-parser dependency is imported outside `imports`.
- No module bypasses `objectstore` to construct browser-visible storage URLs.
- Applying migrations to an empty database and a current baseline database succeeds in CI.

---

## 6. Frontend architecture

**Section class:** `Implementation baseline`  
**Profile applicability:** Base unless a subsection says otherwise.  
**Owner references:** Core 03 §1 through §4, Core 01 §3.3 through §3.3.10.1 and §7.4, Core 04 §1.1.1.

### 6.1 Application structure (`/apps/web`)

| Frontend module       | Responsibility                                                                                                                                                                                    | Required default or constraint                                                                                                                                                                         |
| --------------------- | ------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- | ------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------ |
| `workbook shell`      | Built-in tabs, system-view registration, saved-view selector, startup-surface resolution, view-level filters and grouping controls                                                                | MUST keep all workbook-native surfaces inside the workbook shell rather than separate application modules                                                                                              |
| `grid adapter`        | Cartulary-owned adapter around the selected grid engine, including `record_id` anchoring, column registration, renderer/editor registration, vendor-event translation, and imperative API fencing | The adapter MUST be the only code path that translates vendor-local coordinates, indexes, hooks, plugin events, selection objects, or imperative grid calls into Cartulary events or mutation requests |
| `cell renderers`      | Typed cell display and edit components                                                                                                                                                            | Save state, conflict state, and presence markers MUST render at cell level where applicable                                                                                                            |
| `sync engine`         | Pending-patch queue, autosave, HTTP mutation submission, retry behavior                                                                                                                           | MUST send `record_id`, `base_row_version`, and changed fields only                                                                                                                                     |
| `WebSocket client`    | `hello`, `resume`, replay tracking, `ping` and `pong`, reset handling                                                                                                                             | MUST treat the server stream as read-only and re-query via HTTP on reset                                                                                                                               |
| `collaboration state` | `record_changed`, `presence_snapshot`, `presence_delta`, `job_progress` handling                                                                                                                  | MUST preserve selection anchoring by `record_id` through patch, invalidate, and remove events                                                                                                          |
| `conflict resolver`   | Same-surface conflict review and explicit resolution                                                                                                                                              | MUST open from the affected cell and keep the grid visible                                                                                                                                             |
| `inspector drawer`    | Enrichment, relationships, history, rollback, evidence inspection, party and mention resolution                                                                                                   | MUST remain non-blocking relative to the grid                                                                                                                                                          |
| `presence UI`         | Header avatars, row-gutter indicators, same-cell editing hints                                                                                                                                    | MUST be keyed by stable identifiers rather than labels or row numbers                                                                                                                                  |

### 6.1.1 Grid adapter contract

The frontend MUST isolate the selected grid engine behind a Cartulary-owned grid adapter. The adapter is the only frontend boundary that may translate vendor coordinates into Cartulary anchors.

Definitions used by this section:

| Term                   | Required meaning                                                                                                                              |
| ---------------------- | --------------------------------------------------------------------------------------------------------------------------------------------- |
| `selected grid engine` | The grid dependency currently selected by the stack baseline.                                                                                 |
| `vendor coordinate`    | Any selected-grid-engine-local row, column, visual, physical, source, renderable, or selection coordinate.                                    |
| `Cartulary anchor`     | Stable identity tuple containing the relevant `view_schema_id`, `record_id`, `row_version` where needed, and `field_key` where cell-specific. |
| `render_binding_table` | Adapter-local mapping from current viewport and overscan rows to Cartulary row anchors and presentation-row classifications.                  |
| `column_binding_table` | Adapter-local mapping from rendered columns to stable `field_key`, writeability, renderer, editor, and layout metadata.                       |

The adapter MUST classify grid state as follows.

| State family                            | Authority classification           | Required handling                                                                                                                          |
| --------------------------------------- | ---------------------------------- | ------------------------------------------------------------------------------------------------------------------------------------------ |
| `record_id`                             | Authoritative                      | MUST be the row identity for focus, selection, edits, pending patches, presence, conflict markers, inspector routing, and history routing. |
| `row_version`                           | Authoritative                      | MUST be the base version for record-scoped optimistic writes.                                                                              |
| `field_key`                             | Authoritative                      | MUST be the cell write identity and query-field identity.                                                                                  |
| `view_schema_id`                        | Authoritative                      | MUST define field registry, mutability, read contracts, write contracts, grouping keys, sort keys, and surface identity.                   |
| `view_row_v1`                           | Authoritative row input            | MUST populate the adapter's render-binding table.                                                                                          |
| `view_row_patch_v1`                     | Authoritative sparse update input  | MUST update only the addressed `record_id` and changed `field_key` members.                                                                |
| `meta.query`                            | Authoritative query-state echo     | MUST be the displayed applied query state after server canonicalization.                                                                   |
| Sync-engine pending queue               | Authoritative local mutation state | MUST own unsent or retrying patches. The grid adapter MAY display queue state but MUST NOT reorder, drop, or reinterpret queue units.      |
| Conflict resolver state                 | Authoritative local conflict state | MUST own same-field conflict display and resolution actions.                                                                               |
| Vendor visual row index                 | Non-authoritative                  | MAY locate a visible cell only until translated to `record_id` and `field_key`.                                                            |
| Vendor physical/source/renderable index | Non-authoritative                  | MUST NOT be emitted outside the adapter as record identity.                                                                                |
| Vendor selection object                 | Non-authoritative                  | MUST NOT be persisted or sent to the server.                                                                                               |
| Vendor plugin state                     | Non-authoritative unless mapped    | MUST be inaccessible outside the adapter unless explicitly mapped to a Cartulary contract and test.                                        |
| Vendor imperative instance state        | Non-authoritative unless wrapped   | MUST be fenced behind adapter methods and tests.                                                                                           |

The adapter MUST expose one Cartulary event vocabulary to the rest of the frontend.

| Adapter event               | Required payload                                                                                                                       | Omission and error behavior                                                                                                                                                                                                 |
| --------------------------- | -------------------------------------------------------------------------------------------------------------------------------------- | --------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| `cell_focus_changed`        | `view_schema_id`, `record_id`, `field_key`, optional `row_version`                                                                     | If the vendor cell cannot be resolved to a data row and schema field, the adapter MUST emit no focus event and MUST clear or retain focus according to the prior valid Cartulary anchor, not according to the vendor index. |
| `selection_changed`         | `view_schema_id`, ordered selected anchors as `record_id` plus `field_key` ranges where applicable                                     | Group headers, spacer rows, and non-data rows MUST NOT become selected record anchors.                                                                                                                                      |
| `cell_edit_committed`       | `view_schema_id`, `record_id`, `base_row_version`, `field_key`, `value`, `client_txn_id`                                               | If any required anchor is missing, the adapter MUST reject the local commit before sync-engine enqueue.                                                                                                                     |
| `paste_committed`           | `view_schema_id`, ordered cell range mapped to existing `record_id` anchors or create-row anchors, field-keyed values, `client_txn_id` | The adapter MUST preserve visual paste order after translation to stable anchors. It MUST NOT rely on vendor source indexes as mutation identity.                                                                           |
| `query_state_requested`     | `view_schema_id`, requested sort, filters, grouping, and initiating UI source                                                          | The adapter MUST send only stable contract keys. It MUST NOT send visible labels, vendor column indexes, projection-table names, or storage-table names.                                                                    |
| `adapter_refresh_requested` | `view_schema_id`, refresh class from the grid-state refresh table, reason code                                                         | Unknown refresh classes are invalid.                                                                                                                                                                                        |

The adapter MUST normalize vendor cell events through this deterministic algorithm before emitting Cartulary events or enqueueing mutations.

```text
on vendor cell event:
  read vendor row coordinate and vendor column coordinate
  resolve row coordinate through current render_binding_table
  resolve column coordinate through current column_binding_table

  if row binding is absent:
    do not emit a Cartulary row event
    do not enqueue a mutation
    preserve the last valid Cartulary anchor only if the UI action did not claim a new data-row focus

  if column binding is absent or not bound to one field_key:
    do not emit a Cartulary cell event
    do not enqueue a mutation

  if the resolved row is a group header, spacer, loading row, summary row, or presentation-only row:
    do not emit a mutation-capable event

  emit the corresponding Cartulary adapter event using view_schema_id, record_id, row_version when required, and field_key
```

The adapter MUST maintain a render-binding table for the current viewport and overscan region. Each data-row binding MUST contain `view_schema_id`, `record_id`, `row_version`, and the row's current position in the adapter's displayed order. Each data-column binding MUST contain `field_key`, writeability state, read renderer, edit renderer, and default-hidden or visible layout state.

The adapter MUST treat group headers, loading rows, summary rows, spacer rows, and vendor-generated rows as non-data rows unless the active view contract explicitly declares them writable data rows. After sort, filter, grouping, cursor continuation, live update, hidden-field layout, virtual scrolling, or vendor plugin action, a commit MUST still target the original `record_id` and `field_key`. The terms `visual`, `physical`, `source`, and `renderable` are vendor-coordinate concepts only and MUST NOT become Cartulary identity.

#### Acceptance criteria

- **DG-R08-AC-006:** The frontend architecture table uses `grid adapter` as the vendor-grid boundary.
- **DG-R08-AC-007:** The architecture table states that vendor events and imperative APIs are mediated by the adapter.
- **DG-R08-AC-008:** No architecture row allows direct mutation routing from a renderer, editor, plugin, or component outside the adapter and sync engine.
- **DG-R08-AC-009:** This subsection defines authoritative and non-authoritative grid state in a table.
- **DG-R08-AC-010:** The event vocabulary table includes `cell_focus_changed`, `selection_changed`, `cell_edit_committed`, `paste_committed`, `query_state_requested`, and `adapter_refresh_requested`.
- **DG-R08-AC-011:** Every mutation-capable adapter event requires `record_id`, `field_key`, and the version value needed by the server-owned mutation contract.
- **DG-R08-AC-012:** The subsection explicitly forbids vendor row indexes as mutation identity.
- **DG-R08-AC-013:** The subsection specifies fail-closed behavior when a vendor coordinate cannot be translated to a Cartulary anchor.
- **DG-R08-AC-014:** The subsection uses `visual`, `physical`, `source`, and `renderable` only as vendor-coordinate concepts, not as Cartulary identity.
- **DG-R08-AC-015:** The subsection states that data-row commits after sorting, filtering, grouping, live update, and virtual scrolling target `record_id` and `field_key`.
- **DG-R08-AC-016:** The subsection states that group headers and presentation-only rows are never writable rows.

### 6.1.2 Grid capability allowlist

The grid adapter MUST define a closed capability allowlist for the selected grid engine. Plugin registration is closed by default. A vendor plugin or capability is enabled only when it appears in the allowlist with class, owner, and test evidence.

| Class            | Meaning                                                                                    | Required handling                                                                                                     |
| ---------------- | ------------------------------------------------------------------------------------------ | --------------------------------------------------------------------------------------------------------------------- |
| `adapter_owned`  | Required for Cartulary UI behavior and mediated by the grid adapter.                       | Enabled only through adapter code and covered by unit tests.                                                          |
| `server_owned`   | The grid may expose UI affordances, but canonical semantics are owned by server contracts. | Requests MUST round-trip through server-owned routes or canonical contract artifacts before display is authoritative. |
| `disabled`       | Conflicts with current Cartulary behavior or lacks current contract coverage.              | MUST NOT be registered, enabled, reachable, or callable in ordinary runtime.                                          |
| `future_profile` | Potentially useful but outside the current baseline.                                       | MUST require a later owner-contract or stack-baseline revision before enablement.                                     |

The current-profile default mapping is:

| Capability family                | Default class         | Required current-profile handling                                                                    |
| -------------------------------- | --------------------- | ---------------------------------------------------------------------------------------------------- |
| Virtualized rendering            | `adapter_owned`       | Adapter MAY use the grid engine for viewport rendering, but row identity remains `record_id`.        |
| Keyboard navigation              | `adapter_owned`       | Adapter MUST map key effects to Cartulary focus, selection, edit, and paste contracts.               |
| Inline scalar editors            | `adapter_owned`       | Editors MUST emit `cell_edit_committed` only through the adapter.                                    |
| Custom renderers                 | `adapter_owned`       | Renderers MAY display cell state but MUST NOT own authoritative mutation state.                      |
| Clipboard paste                  | `adapter_owned`       | Paste MUST translate into the shared Cartulary paste and tabular-ingest contract.                    |
| Sort UI                          | `server_owned`        | UI MAY request sort; server `meta.query.sort[]` is authoritative.                                    |
| Filter UI                        | `server_owned`        | UI MAY request filters; server normalized filter state is authoritative.                             |
| Group UI                         | `server_owned`        | UI MAY request one declared grouping key; group headers remain presentation-only.                    |
| Validation display               | `server_owned`        | Client MAY show local hints, but server validation and mutation result are authoritative.            |
| Presence markers                 | `adapter_owned`       | Markers MUST be keyed by `record_id` and `field_key`, not vendor coordinates.                        |
| Same-field conflict markers      | `adapter_owned`       | Markers MUST be keyed by conflict state from the sync/conflict subsystem.                            |
| Vendor source-data mutation APIs | `disabled`            | No direct use outside the adapter; no authoritative data mutation through vendor data APIs.          |
| Vendor formulas                  | `disabled`            | MUST NOT be enabled unless a later owner contract defines formula semantics.                         |
| Vendor undo/redo                 | `disabled`            | MUST NOT bypass Cartulary revision, conflict, or history contracts.                                  |
| Vendor hidden-row semantics      | `disabled`            | MUST NOT determine row visibility; server query state and view contracts own visibility.             |
| Vendor context-menu commands     | `disabled`            | MUST be unavailable unless every command is mapped to a Cartulary action and test.                   |
| Manual row movement              | `future_profile`      | MUST NOT mutate ordering or source state in the current baseline.                                    |
| Manual column movement           | `future_profile`      | MUST NOT become persisted layout unless a layout contract explicitly owns it.                        |
| Plugin registration              | `disabled` by default | Each enabled plugin MUST appear in the allowlist with class, owner, and tests.                       |
| Theme and CSS customization      | `adapter_owned`       | MAY affect presentation only and MUST NOT change row identity, field identity, or mutation behavior. |

#### Acceptance criteria

- **DG-R08-AC-017:** The allowlist defines exactly the four class values `adapter_owned`, `server_owned`, `disabled`, and `future_profile`, or stricter successors.
- **DG-R08-AC-018:** The default capability mapping includes every capability family listed in this subsection.
- **DG-R08-AC-019:** The subsection forbids vendor formulas, vendor undo/redo, vendor source-data mutation APIs, vendor hidden-row semantics, and vendor context-menu commands unless later mapped to Cartulary contracts and tests.
- **DG-R08-AC-020:** The subsection states that plugin registration is closed by default.

### 6.1.3 Renderer and editor lifecycle rules

Renderers MAY read `view_row_v1`, `view_row_patch_v1`, presence state, save state, and conflict state. Renderers MUST NOT mutate authoritative state.

Editors MUST emit commits only through the adapter event vocabulary. Editors MUST reset local state when prepared for a different `record_id`, `field_key`, or `row_version`. Editor reuse across rows MUST reset by stable Cartulary anchors, not vendor coordinates.

Renderers and editors MUST release portals, subscriptions, timers, observers, and stale closures during unmount, remount, surface switch, and engine destroy. A renderer or editor MUST NOT retain a vendor row index as a future mutation target. A renderer or editor MUST NOT call vendor data-mutation APIs directly. A renderer or editor MUST NOT call HTTP mutation routes directly; the sync engine remains the mutation submission owner.

#### Acceptance criteria

- **DG-R08-AC-021:** This subsection exists.
- **DG-R08-AC-022:** The subsection requires cleanup of portals, subscriptions, timers, observers, and stale closures.
- **DG-R08-AC-023:** The subsection forbids direct HTTP mutation calls from renderers and editors.
- **DG-R08-AC-024:** The subsection states that editor reuse across rows resets by stable Cartulary anchors, not vendor coordinates.

### 6.1.4 Grid-state refresh semantics

The grid adapter MUST use one of the refresh classes below whenever server results, live collaboration events, query replacement, or adapter lifecycle changes require visible grid refresh. Unknown refresh classes are invalid. The adapter MUST NOT invent best-effort refresh behavior without adding the class to this table and adding tests.

| Refresh class                    | Trigger examples                                                                                                               | Selection and focus                                                                                                        | Scroll                                                                                                | Pending local edits                                             | Inspector                                                           | Presence and conflict markers                   |
| -------------------------------- | ------------------------------------------------------------------------------------------------------------------------------ | -------------------------------------------------------------------------------------------------------------------------- | ----------------------------------------------------------------------------------------------------- | --------------------------------------------------------------- | ------------------------------------------------------------------- | ----------------------------------------------- |
| `cell_patch`                     | Sparse `view_row_patch_v1` changes visible fields.                                                                             | Preserve by `record_id` and `field_key`.                                                                                   | Preserve.                                                                                             | Preserve unless the same field enters conflict state.           | Preserve if the selected `record_id` remains active.                | Recompute markers by stable anchors.            |
| `row_refresh`                    | Full row returned after create, patch, attach, or explicit row refresh.                                                        | Preserve by `record_id`; update `row_version`.                                                                             | Preserve.                                                                                             | Preserve or clear only according to sync/conflict result.       | Preserve if selected `record_id` remains active.                    | Recompute markers by stable anchors.            |
| `invalidate_requery`             | Sparse patch cannot safely express the change; server emits invalidate.                                                        | Preserve by `record_id` and `field_key` when the record returns in the requery result; otherwise clear the invalid anchor. | Preserve when anchor remains visible; otherwise scroll behavior follows active query-result defaults. | Preserve sync-engine queue state; do not retarget.              | Preserve only when selected `record_id` remains visible and active. | Recompute after requery.                        |
| `query_replace_preserve_anchors` | Sort, filter, grouping, saved-view, or cursor-bound result replacement.                                                        | Preserve anchors when the same `record_id` and `field_key` remain in the result; otherwise clear invalid anchors.          | Keep anchor in view when possible.                                                                    | Preserve and do not retarget.                                   | Preserve only for surviving selected `record_id`.                   | Recompute by stable anchors.                    |
| `surface_remount`                | `view_schema_id` change, grid-engine reinitialization, capability registry change, or unrecoverable adapter invariant failure. | Clear surface-local selection and focus unless the new surface explicitly resolves the same stable anchor.                 | Reset to new surface default.                                                                         | Preserve sync-engine queue state; do not drop unsent mutations. | Close unless it can rebind to a valid same-incident `record_id`.    | Rehydrate from collaboration state after mount. |

Pending local edits MUST never retarget to a different `record_id`.

#### Acceptance criteria

- **DG-R08-AC-025:** The refresh table covers `cell_patch`, `row_refresh`, `invalidate_requery`, `query_replace_preserve_anchors`, and `surface_remount`.
- **DG-R08-AC-026:** Each refresh class defines selection/focus, scroll, pending-edit, inspector, presence, and conflict-marker behavior.
- **DG-R08-AC-027:** The subsection states that pending local edits must never retarget to a different `record_id`.

### 6.2 Workbook-surface rules

The frontend baseline MUST preserve the five built-in tabs:

- Timeline,
- Hosts,
- Identities,
- Evidence,
- Notes.

The workbook shell MUST also surface the pack-independent system views and workbook-native coordination surfaces already defined by the owner docs:

- Indicators,
- Assessments,
- Task Requests,
- Decisions,
- Parties,
- Communications Log,
- Handoff,
- Status Review,
- Lesson.

For `comm_log`, `handoff`, `status_review`, and `lesson`, the canonical workbook-surface identity MUST remain the standardized `view_schema_id` surface. An implementation-owned `scope='system'` saved view bound to the same `view_schema_id` MAY exist as an additional preset surface, but it MUST remain a distinct saved-view object and MUST NOT replace the required base surface. All such surfaces MUST remain within the workbook shell.

Saved-view scope MUST NOT be used as access control in the frontend. Scope affects discoverability and mutability of the saved-view configuration object only.

### 6.3 Startup-surface selection

The workbook shell MUST apply the ordered startup selection already owned by Core 03:

1. explicit launch `sheet_ref`,
2. caller `home_sheet_ref`,
3. incident `default_sheet_ref`,
4. `cartulary.view.timeline.v1`.

If a referenced saved view or view schema is missing, invisible, or invalid because a required optional pack is unavailable, the frontend MUST clear the invalid pointer and continue through the fallback chain rather than failing workbook open.

### 6.4 Workspace packages

| Package                    | Baseline responsibility                                                                                                                                                                                                                     |
| -------------------------- | ------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| `/packages/grid-adapter`   | Cartulary-owned frontend grid adapter. Owns selected-grid integration, vendor-coordinate translation, column and field binding, renderer/editor registration, grid capability allowlist, imperative API fencing, and adapter test fixtures. |
| `/packages/ui`             | Reusable presentational components with no workbook-state ownership and no direct vendor-grid mutation authority.                                                                                                                           |
| `/packages/protocol-ts`    | Generated protocol types and helpers derived from `/contracts/*`                                                                                                                                                                            |
| `/packages/view-contracts` | TypeScript-consumable adapters over `/contracts/view-schemas/*`                                                                                                                                                                             |
| `/packages/test-utils`     | Shared protocol, UI, grid-adapter, and visual-fixture test helpers, including mock WebSocket servers and row fixtures                                                                                                                       |

### 6.5 Frontend configuration

- `pnpm-workspace.yaml` MUST include only `/apps/web` and `/packages/*` unless the repository intentionally changes the workspace shape.
- `tsconfig.base.json` SHOULD contain strictness defaults only.
- Vite dev mode SHOULD use a fixed port with `strictPort: true`.
- Dev mode MUST proxy `/api/v1/*` and `/ws/v1/*` to the Go server rather than inventing a parallel API base contract.
- The frontend MUST use relative API and WebSocket paths unless the repository intentionally changes that baseline.
- Field writeability, `conflict_resolution_class`, `entity_binding_mode`, default sort, filterability, and grouping MUST derive from view contracts rather than from hard-coded visible labels.

### Verification

- The five built-in tabs and the required workbook-native system surfaces appear in the workbook shell.
- Required coordination surfaces retain the standardized `view_schema_id` identity even when additive `scope='system'` saved views are present.
- Startup-surface fallback is covered by automated tests.
- Mutations sent by the frontend are keyed by `record_id`, `base_row_version`, and `field_key`.
- No React component hard-codes a behavior that already belongs to the view-schema contracts by visible label alone.
- The grid adapter is the only package that imports or calls selected-grid mutation, plugin, or imperative instance APIs.
- Disabled grid capabilities are absent from ordinary runtime registration and fail adapter-level reachability tests.
- Pending local edits do not retarget after sort, filter, grouping, live update, virtual scrolling, or surface remount.

---

## 7. Development workflow

**Section class:** `Implementation baseline`  
**Profile applicability:** Base  
**Owner references:** Core 04 §9, plus the contract-derivation and generated-artifact rules in this guide.

### Pending repo-control revalidation

This section remains an intended command and workflow baseline, not a verified task-surface inventory. Until the live `Makefile`, scripts, and CI control surface are revalidated, the command table and verification-tier statements below are intended baseline only.

### 7.1 Canonical commands

If the repository exposes a root `Makefile`, it SHOULD remain the stable human-facing task surface. Treat the command table below as the intended implementation baseline until revalidated against the live repo control surface.

| Command              | Minimum responsibility                                                                                       |
| -------------------- | ------------------------------------------------------------------------------------------------------------ |
| `make bootstrap`     | Install tools, install workspace dependencies, prepare local services                                        |
| `make db-up`         | Start local Postgres and MinIO                                                                               |
| `make db-reset`      | Recreate the database and run migrations                                                                     |
| `make dev`           | Start the Go server and Vite dev server                                                                      |
| `make generate`      | Regenerate Go and TypeScript artifacts derived from `/db/queries/*` and `/contracts/*`                       |
| `make backend-store` | Run service-backed store-domain `U-*` backend evidence that keeps unit-layer IDs while using real Postgres   |
| `make test-fast`     | Run the narrower pure-unit, service-backed backend, and frontend unit/process loop                           |
| `make test`          | Run the authoritative full test corpus, including manifest-verified browser E2E plus explicit support suites |
| `make lint`          | Run backend vet/lint and frontend lint/type checks                                                           |
| `make check`         | Developer verification gate                                                                                  |
| `make ci`            | Provider-neutral CI enforcement entrypoint                                                                   |
| `make build`         | Produce the app build with embedded frontend assets                                                          |

If the repository uses both a root task-surface `Makefile` and `AGENTS.md`, both MUST be updated together when this task surface changes.

### 7.2 Verification tiers

| Verification tier      | Required checks                                                                                                                                                         | Blocking condition                                 |
| ---------------------- | ----------------------------------------------------------------------------------------------------------------------------------------------------------------------- | -------------------------------------------------- |
| Developer verification | formatting, lint or vet, Go build, Go tests, TypeScript build or type check, Vitest, browser E2E coverage, `make generate` drift check, migration application check     | Any failure blocks ordinary development completion |
| Release verification   | all developer verification checks, dependency license report, SBOM generation, release-artifact smoke build, any profile-specific release checks for claimed extensions | Any failure blocks release publication             |

`make check` is the required developer gate. It MUST include contract-generation drift detection, migration verification, and a failure condition when any authoritative phase-manifest row is absent from actual execution. When browser suites depend on the real Playwright web-server bootstrap, the gate must run those suites under one owned shared stack rather than parallelizing multiple independent startup attempts. `make ci` composes the same execution-truth guarantee into the provider-neutral CI surface.

The repository MUST distinguish **codegen drift** from **migration drift**:

- **codegen drift** means generated outputs change after `make generate`,
- **migration drift** means schema-affecting behavior or query changes are not represented in the numbered migration path or migrations do not apply cleanly in CI.

Browser verification MUST also distinguish shared-stack functional browser work from isolated browser classes. Stateful browser suites and fixture-sensitive measurement suites remain isolated after the heavy parallel block; ordinary web-server-backed browser suites belong to one serialized shared-stack orchestration path.

Backend verification MUST distinguish phase evidence from execution dependency. The repository now uses two axes:

- evidence layer: `U-*`, `I-*`, and `E-*` remain the authoritative phase IDs and must not be reclassified only because a harness uses real services,
- execution dependency: pure unit, service-backed store-domain, service-backed runtime, process, and isolated browser work.

`backend-unit` is reserved for pure backend tests with no `pgtest`, `s3test`, or real runtime startup. `backend-store` owns the service-backed store-domain `U-*` slice, while `backend-integration`, `backend-process`, and browser suites continue to own runtime, process, and browser-backed evidence respectively.

### 7.3 Local development loop

The default local loop is:

1. `make db-up`
2. `make dev`
3. use the Vite-served browser app against the Go server and local Postgres plus MinIO

Production packaging MUST embed the built frontend assets into the application deployable. The production deployable MUST NOT depend on the Vite dev server.

### 7.4 Test strategy

**Backend unit tests** cover module-local behavior with no external service dependency.

**Backend store tests** keep `U-*` evidence where the owner contract is still store-domain behavior, but they run against real Postgres rather than stubs. These tests execute through `make backend-store`, not `make backend-unit`.

**Backend integration tests** use real Postgres and MinIO through `testcontainers-go` or an equivalent real-service harness. Integration tests MUST exercise:

- HTTP routes,
- WebSocket behavior,
- object-store lifecycle,
- projection maintenance,
- migration application.

**Frontend unit and component tests** cover:

- cell renderers,
- the sync engine,
- conflict-state transitions,
- view-contract-driven behavior,
- WebSocket message handling.

**Frontend end-to-end tests** MUST cover at minimum:

- same-row different-field auto-merge,
- same-field conflict presentation,
- offline edit plus replay,
- multi-cell paste with partial conflicts,
- sort, filter, or group changes while an edit is pending,
- startup-surface fallback,
- evidence preview blocked-state handling.

### Verification

- `make check` passes on the repository baseline.
- `make generate` followed by a clean diff is enforced in CI.
- CI applies migrations successfully in an empty and upgrade path environment.
- Release verification produces license and SBOM artifacts.

---

## 8. Collaboration and concurrency model

**Section class:** `Derived behavioral summary`  
**Profile applicability:** Base  
**Owner references:** Core 03 §3 through §4, Core 01 §3.3.4 through §3.3.10.1, Core 04 §1.1.1.

This section is an implementation-facing summary. It does not own the behavioral contract.

### 8.1 Concurrency model summary

The implementation baseline assumes **field-level optimistic concurrency on top of row versioning**.

| Owned behavior                                                  | Implementation consequence                                                               |
| --------------------------------------------------------------- | ---------------------------------------------------------------------------------------- |
| Each visible row is identified by `record_id` and `row_version` | Grid state, selection anchoring, and pending patches MUST be keyed by stable identifiers |
| Different-field concurrent edits auto-rebase                    | Patch handling MUST be field-addressable rather than full-row replacement                |
| Same-field concurrent edits require explicit analyst resolution | Conflict UI MUST preserve saved value and unsaved local draft separately                 |
| Pending local work survives transient disconnect                | Client MUST retain a pending-patch queue distinct from server-committed state            |

### 8.2 Autosave and pending-patch handling

Autosave is the baseline interaction model. There is no normal workbook-wide Save button. Enter, Tab, blur, and paste completion are commit points unless a more specific surface contract says otherwise.

The pending-patch queue is client-local unsaved state. It MUST survive transient network interruption and ordinary reconnect. It MUST NOT be serialized as committed incident history.

### 8.3 Same-field conflicts

Only the affected cell enters conflict state. The saved server value remains the visible committed value. The analyst’s unsaved draft remains client-local. The resolver MUST open from the affected cell and keep the grid visible.

`text_compare_merge` and `collection_review` are contract-declared conflict classes. The frontend MUST derive those classes from the active view contract rather than from field labels or ad hoc component logic.

### 8.4 Collaboration stream

The collaboration stream is incident-scoped and read-only from the server’s perspective. All authoritative mutations continue to flow through HTTP routes.

Implementation consequences:

- replayable messages and ephemeral messages SHOULD remain separated in client handling,
- `record_changed` handling MUST preserve `record_id` anchoring,
- `invalidate` MUST trigger a view-shaped refresh rather than blind row synthesis,
- resume logic MUST fall back to HTTP re-query on reset.

### Verification

- Multi-context end-to-end tests demonstrate different-field auto-merge and same-field explicit conflict behavior.
- A lost connection followed by resume preserves unsaved local drafts and replays committed server events correctly.
- Selection never silently rebinds to a different record because of sorting, filtering, grouping, or replay.

---

## 9. Evidence lifecycle

**Section class:** `Derived behavioral summary`  
**Profile applicability:** Base  
**Owner references:** Core 01 evidence and handle route-owner sections, Core 02 evidence and blob lifecycle sections, Core 03 §8, Core 04 §2 and §4.5.

This section is an implementation-facing summary. It does not own the evidence or handle contract.

### 9.1 Two-step attachment

The baseline evidence path is:

1. create a blob slot,
2. upload bytes,
3. finalize by attaching the blob to an evidence record.

Step 1 creates a pending blob slot and returns the accepted create contract plus a server-issued upload target. The browser MAY upload bytes using that target, but attached evidence becomes visible only after successful finalize-and-attach against the evidence record.

A pending or failed blob MUST NOT create the appearance of attached evidence. Evidence counts, preview affordances, and download affordances MUST remain driven by committed evidence state, not by upload intent.

### 9.2 Handle issuance and redemption

The browser MUST redeem preview and download access through same-origin opaque handles. It MUST NOT construct raw object-store URLs.

Implementation summary:

| Browser objective                      | Required public contract                                                                                       | Required server-issued artifact       | Default supported browser behavior                                            | Unsupported interpretation                                                                 |
| -------------------------------------- | -------------------------------------------------------------------------------------------------------------- | ------------------------------------- | ----------------------------------------------------------------------------- | ------------------------------------------------------------------------------------------ |
| Upload bytes for one pending blob slot | `POST /api/v1/object-blobs` plus the returned upload target                                                    | `object_blob_id` plus `upload_target` | The browser MAY use the short-lived target for the accepted pending slot only | Persisting the target as evidence-access state or treating it as a preview or download URL |
| Preview evidence bytes                 | `POST /api/v1/evidence-records/{record_id}/preview-handle` then `GET /api/v1/evidence-handles/{handle_token}`  | preview handle                        | The browser redeems a same-origin opaque handle through the application       | Constructing a raw object-store URL                                                        |
| Download evidence bytes                | `POST /api/v1/evidence-records/{record_id}/download-handle` then `GET /api/v1/evidence-handles/{handle_token}` | download handle                       | The browser redeems a same-origin opaque handle through the application       | Reusing `upload_target` as a download path or constructing a raw object-store URL          |

- preview-handle issuance is intentionally non-idempotent,
- download-handle issuance is intentionally non-idempotent,
- preview handles are short-lived and multi-redeem within their validity window,
- download handles are short-lived and single-use,
- handle redemption MUST re-derive current authorization and current evidence or blob state.

### 9.3 Lifecycle separation

The blob-upload lifecycle and evidence-custody lifecycle remain separate machines. The implementation MUST keep them separate in code, storage, and tests.

An evidence record may exist with no blob at all. Requested or pending-receipt evidence therefore remains a first-class evidence row even before bytes are attached.

### 9.4 Preview safety boundary

Imported or uploaded content is hostile by default. Preview issuance and redemption MUST fail closed when the current evidence or blob state is missing, pending, failed, quarantined, or otherwise inconsistent. Active-content formats MUST not execute inside the main application origin. Blocked or unsupported preview states MUST fail explicitly; they MUST NOT silently downgrade to download.

### Verification

- Attaching a blob requires the two-step flow, and attached evidence state becomes visible only after successful finalization.
- Preview redemption works through the application origin only.
- Download handles fail after first successful consumption.
- Blocked or unsupported preview states fail explicitly and do not silently downgrade to download.
- State transitions to pending, failed, missing, or quarantined block preview and download.

---

## 10. Projection tables and workbook-surface mapping

**Section class:** `Implementation baseline`  
**Profile applicability:** Base unless a row says otherwise.  
**Owner references:** Core 01 §7.4 and §19, Core 02 record-owner sections, Core 03 §2 and §16.

### 10.1 Projection rules

Projection tables are the hot-path workbook realization layer. They are disposable caches, not authority.

The baseline rules are:

- one projection row per primary source record,
- projection updates occur in the same transaction as the source write or through a deterministic invalidation plus rebuild path,
- every mutable projection row surfaced to the client carries `record_id` and `row_version`,
- the client never infers identity from row position, visible grouping, or displayed values.

The implementation baseline MUST prefer projection tables over Postgres materialized views for hot workbook surfaces.

### 10.2 Surface ownership matrix

| Surface or object group | Authoritative model                                             | Backend owner | Projection or route realization                                                           | Frontend owner                                                   | Default surface class and notes                                                                                                                                                                                                                        |
| ----------------------- | --------------------------------------------------------------- | ------------- | ----------------------------------------------------------------------------------------- | ---------------------------------------------------------------- | ------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------ |
| Timeline                | `timeline_event`                                                | `timeline`    | `timeline_grid_projection`; `cartulary.view.timeline.v1`                                  | workbook shell, grid adapter, Timeline columns                   | Built-in tab. Zero-field create remains allowed only because the owner contract explicitly allows it.                                                                                                                                                  |
| Hosts                   | `host`                                                          | `entities`    | `host_grid_projection`; `cartulary.view.hosts.v1`                                         | workbook shell, grid adapter, Hosts columns                      | Built-in tab. `entity_origin` surface.                                                                                                                                                                                                                 |
| Identities              | `identity`                                                      | `entities`    | `identity_grid_projection`; `cartulary.view.identities.v1`                                | workbook shell, grid adapter, Identities columns                 | Built-in tab. `entity_origin` surface.                                                                                                                                                                                                                 |
| Evidence                | `evidence` plus joined blob metadata                            | `evidence`    | `evidence_grid_projection`; `cartulary.view.evidence.v1`                                  | workbook shell, grid adapter, Evidence columns, inspector        | Built-in tab. Visible text and hidden party links remain separate concepts.                                                                                                                                                                            |
| Notes                   | `artifact` filtered to `artifact_type='note'`                   | `links`       | `artifact_grid_projection`; `cartulary.view.notes.v1`                                     | workbook shell, grid adapter, Notes columns                      | Built-in tab. Notes are one artifact-backed surface, not the artifact family as a whole.                                                                                                                                                               |
| Indicators              | canonical `indicator` plus observations and lifecycle intervals | `entities`    | `indicator_grid_projection`; `cartulary.view.indicators.v1`                               | workbook shell, system-view registration, indicator columns      | Contract-backed system view. Existing rows derive mutability from the view contract and are not generic free-edit rows.                                                                                                                                |
| Assessments             | `assessment`                                                    | `entities`    | `assessment_grid_projection`; `cartulary.view.assessments.v1`                             | workbook shell, system-view registration, assessment columns     | Contract-backed system view.                                                                                                                                                                                                                           |
| Task Requests           | `task_request`                                                  | `links`       | `task_request_grid_projection`; `cartulary.view.task_requests.v1`                         | workbook shell, system-view registration, queue-oriented columns | Contract-backed system view.                                                                                                                                                                                                                           |
| Decisions               | `decision`                                                      | `links`       | `decision_grid_projection`; `cartulary.view.decisions.v1`                                 | workbook shell, system-view registration, decision columns       | Contract-backed system view.                                                                                                                                                                                                                           |
| Parties                 | `party`                                                         | `links`       | `party_grid_projection`; `cartulary.view.parties.v1`                                      | workbook shell, system-view registration, party columns          | Contract-backed system view. Party identity is incident-scoped coordination identity, not deployment-local account identity.                                                                                                                           |
| Communications Log      | `artifact` filtered to `artifact_type='comm_log'`               | `links`       | `artifact_grid_projection` filtered to `comm_log`; `cartulary.view.comm_log.v1`           | workbook shell, coordination-surface components                  | Workbook-native coordination surface. Canonical base surface is the standardized `view_schema_id`; an additional `scope='system'` saved view over the same schema MAY exist as an additive preset only and MUST NOT replace the required base surface. |
| Handoff                 | `artifact` filtered to `artifact_type='handoff'`                | `links`       | `artifact_grid_projection` filtered to `handoff`; `cartulary.view.handoff.v1`             | workbook shell, coordination-surface components                  | Workbook-native coordination surface. Canonical base surface is the standardized `view_schema_id`; an additional `scope='system'` saved view over the same schema MAY exist as an additive preset only and MUST NOT replace the required base surface. |
| Status Review           | `artifact` filtered to `artifact_type='status_review'`          | `links`       | `artifact_grid_projection` filtered to `status_review`; `cartulary.view.status_review.v1` | workbook shell, coordination-surface components                  | Workbook-native coordination surface. Canonical base surface is the standardized `view_schema_id`; an additional `scope='system'` saved view over the same schema MAY exist as an additive preset only and MUST NOT replace the required base surface. |
| Lesson                  | `artifact` filtered to `artifact_type='lesson'`                 | `links`       | `artifact_grid_projection` filtered to `lesson`; `cartulary.view.lesson.v1`               | workbook shell, coordination-surface components                  | Workbook-native coordination surface. Canonical base surface is the standardized `view_schema_id`; an additional `scope='system'` saved view over the same schema MAY exist as an additive preset only and MUST NOT replace the required base surface. |
| Saved views             | `saved_view`                                                    | `incidents`   | `/api/v1/incidents/{incident_id}/saved-views/*`                                           | workbook shell, saved-view selector, layout persistence          | Workbook configuration object, not a projection row. Scope controls discoverability and mutability of the configuration object only.                                                                                                                   |
| Workbook preferences    | `user_workbook_preferences`, `incident_workbook_preferences`    | `incidents`   | `/api/v1/incidents/{incident_id}/workbook-preferences/*`                                  | workbook bootstrap and startup-surface selection                 | Route-backed configuration object, not a workbook projection.                                                                                                                                                                                          |
| Deployment-local users  | `user`                                                          | `auth`        | `/api/v1/users/*`                                                                         | admin UI or admin CLI if present                                 | Not a workbook surface. Deployment-local administration only.                                                                                                                                                                                          |
| Incident memberships    | `incident_membership`                                           | `incidents`   | `/api/v1/incidents/{incident_id}/memberships/*`                                           | admin UI or admin CLI if present                                 | Not a workbook surface. Incident authorization administration only.                                                                                                                                                                                    |

If the implementation exposes standardized optional artifact-backed surfaces in the current profile, the baseline `links` module also owns `cartulary.view.findings.v1`, `cartulary.view.investigative_queries.v1`, and `cartulary.view.forensic_keywords.v1` over filtered artifact-backed or otherwise owner-defined storage. Those surfaces remain optional and MUST use the standardized `view_schema_id` values when exposed.

### Verification

- The repository exposes or plans exactly the required fourteen pack-independent base view-schema identifiers before optional overlays.
- The surface-ownership matrix above is reflected in backend package boundaries and frontend workbook routing.
- Required coordination surfaces retain their standardized `view_schema_id` identity even when additive `scope='system'` saved views are present.
- `comm_log`, `handoff`, `status_review`, and `lesson` remain workbook surfaces rather than separate application modules.

---

## 11. Import subsystem and tabular ingest

**Section class:** `Implementation baseline`  
**Profile applicability:** Base for clipboard paste; `import` for file-based structured import.  
**Owner references:** Core 01 §2.1 and extension import sections, Core 02 import provenance sections, Core 03 import workflow sections, Core 04 workbook-import trust-boundary sections.

### 11.1 Clipboard paste

Clipboard paste is part of the Base-profile workbook hot path.

The implementation baseline for paste is:

- paste occurs into the visible workbook surface,
- known columns map by stable `field_key`,
- paste uses the same stable tabular-ingest contract and mapping engine as file-based import where structure is shared,
- host and identity capture still follows the active field contract, including `entity_binding_mode`,
- paste remains one visible user action even if it materializes multiple record mutations.

### 11.2 File-based structured import

File-based structured import is not part of the default Base surface. When the Import Extension Profile is claimed, it MUST be realized through the dedicated `imports` module and the Phase 2 Workbook Import Assistant object model.

The guide uses the following canonical import objects:

| Object                | Default interpretation                                                  | Minimum implementation responsibility                                                                          |
| --------------------- | ----------------------------------------------------------------------- | -------------------------------------------------------------------------------------------------------------- |
| `import_session`      | One uploaded source file plus one operator-driven workflow              | Persist source identity, creator attribution, parser identity, session status, selected units, and diagnostics |
| `import_unit`         | One candidate ingestable unit discovered from the source file           | Persist locator identity, dimensions, unit status, warnings, and mapping identity when approved                |
| `mapping_fingerprint` | Deterministic identity for one approved source-to-field mapping         | Bind approved mapping persistence and later provenance                                                         |
| `warning_code[]`      | Closed warning vocabulary for downgraded or unsupported source features | Surface parser downgrades without leaking workbook semantics into other modules                                |

The implementation baseline MUST use canonical import-unit vocabulary rather than inventing alternate contract nouns such as “selected-sheet import” or “selected-region import.” A worksheet used range or named range is a locator kind, not the public contract object.

### 11.3 Boundary rules

- File adapters for CSV and XLSX belong only to `imports`.
- Workbook inspection, candidate-unit discovery, preview, header mapping, provenance capture, warning emission, and parser compatibility shims belong only to `imports`.
- Other backend modules MUST consume only the stable tabular-ingest contract and mapping engine.
- Formulas, macros, workbook automation, external links, and other active workbook behaviors MUST be treated as hostile content and MUST NOT execute during import.
- The guide MUST not use Excel-specific semantics as cross-module design shortcuts.

### Verification

- No Go package outside `imports` imports or depends directly on XLSX or OpenXML parsing libraries.
- File-based structured import can be disabled entirely without breaking clipboard paste or ordinary workbook editing.
- Import-created records, mentions, aliases, and observations preserve deterministic provenance.

---

## 12. Deployment, security, and release boundary

**Section class:** `Implementation baseline` unless a subsection says otherwise.  
**Profile applicability:** Base unless a subsection says otherwise.  
**Owner references:** Core 04 §1 through §7 and §12, Core 01 deployment and extension-route sections.

### 12.1 Deployment topology

The smallest useful deployment remains:

- one application container,
- one Postgres service,
- one S3-compatible object-storage service, typically MinIO in disconnected deployments.

On-prem and cloud deployments MAY swap the backing services for managed equivalents as long as the same behavioral contracts hold.

### 12.2 Configuration and runtime roots

`internal/platform/config` is the only implementation owner for loading and validating deployment configuration.

The guide baseline requires code support for:

- a file-backed configuration artifact,
- an explicit file-path override via `CARTULARY_CONFIG_FILE`,
- nested environment-variable overlay via `CARTULARY__*`,
- fail-closed handling for unknown keys,
- explicit runtime-root coverage matching the stable key map below.

| Runtime-root purpose   | Stable configuration key       | Default implementation meaning                                          |
| ---------------------- | ------------------------------ | ----------------------------------------------------------------------- |
| database storage       | `roots.database_storage`       | Authoritative structured-state storage                                  |
| object storage         | `roots.object_storage`         | Authoritative binary-evidence storage                                   |
| backup storage         | `roots.backup_storage`         | Operator-facing backup artifacts and retention material                 |
| reference-pack storage | `roots.reference_pack_storage` | Imported reference packs and extracted pack payloads                    |
| temporary work files   | `roots.temporary_work`         | Staging and transient processing artifacts that may carry incident data |
| export outputs         | `roots.export_outputs`         | Generated reports, presentations, and other rendered outputs            |

This guide intentionally does not restate the full operator-facing configuration artifact or discovery precedence as a second owner. Core 04 §12 remains the owner for those details.

### 12.3 Security boundaries

**Section class:** `Derived behavioral summary`

The implementation baseline MUST respect the following owner-defined security boundaries:

- browser authentication uses server-managed opaque sessions,
- state-changing cookie-authenticated routes use fail-closed CSRF protection,
- evidence upload MAY use a server-issued short-lived upload target under the object-store abstraction,
- evidence preview and download always flow through same-origin opaque handles and MUST NOT rely on browser-constructed raw object-store URLs,
- uploaded evidence, imported workbooks, reference packs, and portability bundles are hostile content until validated,
- workbook formulas, macros, automation, and active content do not execute during import or preview,
- preview allowlisting is derived from current validated media metadata and current object state, not from user hints alone,
- secrets for external integrations and enterprise-auth providers remain server-side deployment-local state.

#### 12.3.1 First-deployment admin bootstrap and deployment-admin boundary

The first deployment admin is created only from the deployment-local bootstrap manifest at `bootstrap.first_admin_manifest_path`. Startup MUST perform bootstrap preflight after deployment-configuration validation and before any HTTP listener, WebSocket listener, or background-job runner starts.

Implementation summary:

- if zero active `deployment_admin` users exist and no bootstrap-completion marker exists, a configured manifest path and a valid manifest are required or startup fails closed,
- if at least one active `deployment_admin` already exists, bootstrap preflight is skipped, and a stale, missing, or invalid manifest at that path does not block ordinary startup,
- if zero active `deployment_admin` users exist and a bootstrap-completion marker already exists, startup fails closed because the current profile defines no lost-admin re-bootstrap path,
- successful bootstrap creates exactly one local user with `is_deployment_admin=true`, `is_active=true`, and `mfa_required=true`, with no incident membership or other incident-scoped authorization.

The bootstrap-created admin then enters the ordinary local TOTP bootstrap flow on first valid password login. Because the created user starts with `mfa_required=true` and no active TOTP factor, the first login follows the existing `mfa_setup_required` -> `bootstrap_token` -> `POST /api/v1/auth/mfa/totp/begin` -> `POST /api/v1/auth/mfa/totp/complete` flow.

`deployment_admin` is deployment-scoped only. It authorizes deployment-local user-account inspection and administration, but it MUST NOT by itself grant incident read, write, preview, download, export, incident-scoped job, or incident WebSocket access. Incident access still requires ordinary incident membership and the incident-role model.

### 12.4 Profile-specific deployment concerns

| Profile                     | Additional implementation concern                                                                             |
| --------------------------- | ------------------------------------------------------------------------------------------------------------- |
| `import`                    | Temporary-work roots, hostile workbook parsing, size ceilings, background import jobs                         |
| `snapshot_reporting`        | Self-contained render assets, release-state handling, artifact approval gates, redaction-driven rendering     |
| `incident_portability`      | Staging under temporary-work roots, checksum verification before visibility, authoritative-source-only export |
| `reference_pack`            | Offline bundle import, activation state, attestation metadata, pack storage roots                             |
| `enterprise_authentication` | Server-side provider configuration, correlation-state storage, provider-to-session convergence                |

Extension-specific code paths MUST remain off by default unless the deployment explicitly claims the corresponding profile.

Runtime claim-state discovery for those deployment-gated families is `GET /api/v1/extensions`. That route returns the current extension registry plus reserved `route_families[]` roots. Requests under a reserved but unclaimed family fail with `404` and `error.code = extension_profile_not_claimed`; they are not treated as ordinary unknown routes. Frontend routing, deployment-gated feature flags, and integration tests SHOULD key off that discovery contract rather than ad hoc route probes.

### 12.5 Backup, restore, and failure modes

The implementation baseline remains:

- Postgres base backup plus WAL archiving or equivalent for structured state,
- object-storage snapshot or versioning for blobs and optional generated artifacts,
- projection tables rebuilt after restore as needed,
- backup execution, restore execution, and restore verification remain deployment-local operator-facing concerns; the current profile defines no public `/api/v1/backups*`, `/api/v1/restores*`, or `/api/v1/restore-verifications*` route family and no workbook-surface equivalent.

Operational backup baseline:

- each successful operational backup produces one retained `backup_set` bound to exactly one declared `consistency_point_at`,
- a successful `backup_set` includes restoreable Postgres state, restoreable object-store state for that same point, and a durable attestation record that includes at minimum `backup_set_id`, `consistency_point_at`, `postgres_restore_anchor`, `object_store_restore_anchor`, `created_at`, `retained_until`, `verification_state`, and `last_verified_restore_at`,
- `verification_state` uses exactly `unverified`, `verified`, or `failed`, and `last_verified_restore_at` MAY be `null` only while `verification_state='unverified'`,
- at least one successful retained `backup_set` MUST have `consistency_point_at` no older than 24 hours,
- each successful `backup_set` plus its restoreable artifacts MUST be retained for at least 30 days.

Restore and restore-verification baseline:

- restore selects exactly one retained `backup_set` and restores Postgres and object-store contents from that same set and same declared `consistency_point_at`,
- restore order remains Postgres, object storage, then projection rebuild,
- missing required Postgres or object-store artifacts, or missing required checksum or integrity proof for the chosen backup mechanism, fails restore before the environment is exposed as ready,
- full restore verification runs in an isolated environment at least every 7 days and after any change to the backup mechanism, `roots.database_storage` binding, `roots.object_storage` binding, or `roots.backup_storage` binding,
- a successful verification sets `verification_state='verified'` and updates `last_verified_restore_at`; a failed verification sets `verification_state='failed'` and updates `last_verified_restore_at`.

For flyaway or disconnected deployments, backup artifacts and restore-verification extracts that carry incident data MUST remain on encrypted storage. `roots.backup_storage` is the authoritative deployment-local root for those artifacts. `roots.export_outputs` and `roots.temporary_work` are not substitute backup roots.

Failure-mode baseline:

- application deployable unavailable: sessions drop, no committed data loss,
- Postgres unavailable: system unavailable,
- object storage unavailable: non-evidence editing remains possible; evidence upload and redeem paths fail,
- projection corruption: rebuild projections from source state.

### 12.6 Release boundary

Release-scoped rendering and external publication remain extension behavior, not ordinary workbook editing behavior.

Implementation consequences:

- release-state handling belongs under the reporting module, not the general row-edit path,
- generated outputs MUST be self-contained and MUST NOT depend on CDN-hosted runtime assets,
- release verification MUST include the developer gate plus SBOM and license checks,
- export redaction and recipient-specific withholding MUST happen at snapshot, render, and release time rather than by hiding live workbook content.

### Verification

- Unknown config keys fail closed.
- Effective configuration covers all six required runtime-root purpose keys or documented default-resolution paths.
- Preview and download never rely on browser-constructed raw object-store URLs as the supported evidence-access mechanism. Upload MAY use a server-issued short-lived upload target, but that target is not a substitute for same-origin preview or download handles.
- The implementation exposes no public `/api/v1/backups*`, `/api/v1/restores*`, or `/api/v1/restore-verifications*` route family in the current profile.
- Self-contained rendered outputs can be opened without live remote asset fetches.
- Extension-profile code paths are gated by explicit deployment claims.

---

## 13. Contributor procedure boundary

**Section class:** `Contributor-procedure boundary`  
**Profile applicability:** Base  
**Owner references:** repository procedure owner, expected to be `AGENTS.md` when present.

### Pending repo-control revalidation

This section treats `AGENTS.md` as the intended contributor-procedure owner only. Until a repository-root `AGENTS.md` is revalidated in the live repo control surface, nothing in this section should be read as proof that the file currently exists.

Contributor and coding-agent operating rules SHOULD live in a repository-root `AGENTS.md`. Until that file is revalidated in the repo control surface, treat it as the intended procedure owner and an adoption prerequisite rather than as an independently verified current-repository fact.

When `AGENTS.md` is the active procedure owner, it MUST own at minimum:

- repo map and path conventions for contributors,
- the canonical command surface,
- generated-file edit prohibitions,
- local execution procedure for routine repository work.

This guide MAY mention paths and commands because they are part of the implementation baseline. It does not own the contributor procedure itself.

This guide intentionally contains no assistant-specific or tool-specific work-allocation section. Assigning work to one assistant, IDE, or model family is not part of the durable development-guide surface.

### Verification

- The repository root contains `AGENTS.md`, or the guide adoption work tracks it as an explicit prerequisite.
- The development guide contains no embedded agent-operating contract beyond this boundary statement.
- The development guide contains no tool-specific work-split policy.

---

## 14. Performance reference fixtures and guide acceptance criteria

**Section class:** `Cross-reference`  
**Profile applicability:** Base plus any claimed extension profiles where noted.  
**Owner references:** Core 04 §9, Core 05 §1 through §5, and the guide sections above.

### 14.1 Performance reference fixtures

These fixtures are implementation-facing test anchors mirrored from owner criteria. They support local and CI engineering validation. Any claim-bearing timed or fixture-sensitive public statement MUST additionally satisfy Core 05 benchmark-profile, benchmark-manifest, measurement-predicate, and publication-manifest requirements.

| Measurement context                     | Governing owner                                                | Interpretation in this guide                                                                                                              |
| --------------------------------------- | -------------------------------------------------------------- | ----------------------------------------------------------------------------------------------------------------------------------------- |
| Local or CI engineering run             | Core 04 behavioral thresholds plus this guide's harness mirror | Informative implementation-facing validation                                                                                              |
| Public timed or fixture-sensitive claim | Core 05                                                        | Claim-bearing only when the benchmark profile, benchmark manifest, measurement predicate, and underlying implementation claim all conform |

Unless an owner criterion declares another benchmark profile explicitly, the current claim-bearing benchmark profile is `cartulary.perf.desktop_ref.v1`. Unless a publication criterion says otherwise, `reference incident` means Fixture A.

**Fixture A: large-grid incident**

- 20,000 timeline rows
- 1,000 host rows
- 1,000 identity rows
- 25 concurrently connected analyst sessions on one incident, with presence enabled and representative live row-update traffic
- representative tags, mentions, and links, but not evidence-heavy per row

**Fixture B: evidence-heavy incident**

- 5,000 timeline rows
- 10,000 evidence records
- tens of GB of binary evidence in object storage
- at least one timeline row linked to 100 evidence records
- evidence blobs MAY be stubbed for throughput tests, but evidence metadata, counts, attachment state, and preview handles MUST be real

The implementation baseline SHOULD mirror these fixtures in local and CI performance harnesses where practical. That mirroring is informative engineering validation only unless the run also satisfies Core 05 claim-publication requirements.

When repo task orchestration separates functional browser checks from timing-sensitive browser evidence, any browser test bound to Core 05 measurement predicates or other p95 fixture-sensitive envelope criteria SHOULD run in the isolated measurement suite after heavy backend verification has quiesced. These rows are not parallel-safe gate work because the fixture definition assumes a stable harness rather than incidental contention from unrelated verification tasks.

### 14.2 Guide acceptance criteria

This guide rewrite is complete only when all of the following are true:

1. The opening status block makes the guide explicitly non-normative relative to product behavior, explicitly subordinate to Core 00 through Core 04 for implementation-conformance behavior, and explicitly subordinate to Core 05 where claim-bearing publication behavior is discussed.
2. The guide no longer claims to be the single reference for what to build, where to put it, or how product behavior works.
3. `/contracts/*` is described only as a repo-local derived contract layer with an explicit derivation chain and drift policy.
4. Generated paths are explicitly marked non-editable and tied to code-generation drift checks.
5. Extension-only topics are visibly profile-gated, and Base-only behavior remains the default interpretation when no narrower banner is stated.
6. The guide contains an explicit section-class legend so a new implementer can distinguish implementation baseline from derived behavioral summary without inference.
7. The guide points contributor procedure to `AGENTS.md` and does not embed a second contributor operating contract.
8. The guide contains no tool-specific work-allocation section.
9. The guide includes a surface-ownership matrix that covers at least Timeline, Hosts, Identities, Evidence, Notes, Indicators, Assessments, Task Requests, Decisions, Parties, Communications Log, Handoff, Status Review, Lesson, Saved Views, Workbook Preferences, deployment-local users, and incident memberships.
10. The import section uses canonical import-assistant vocabulary and isolates workbook-parser concerns inside the `imports` module.
11. The deployment and release sections distinguish developer verification from release verification and explicitly require license and SBOM artifacts for release verification.
12. Any section that summarizes route, lifecycle, concurrency, or security behavior names its owner references instead of presenting itself as an independent behavioral owner.
13. The technology-stack section distinguishes planned stack-family choices from manifest-verified package inventory and names repo manifests and lockfiles current as of 2026-04-15 as the verification source for exact package truth.
14. The evidence and object-storage sections distinguish short-lived upload targets from same-origin preview and download handles and prohibit browser-constructed raw object-store URLs as supported evidence-access state.
15. Coordination-surface sections state that the standardized `view_schema_id` surface is mandatory and that any `scope='system'` saved view over the same schema is additive only.
16. Repo-fact sections that depend on live repository control files, including package inventory, the monorepo tree, task surface, and contributor-procedure owner, are explicitly marked as pending repo-control revalidation or independently revalidated.
17. Deployment runtime-root coverage includes `backup storage`, the deployment section includes first-deployment-admin bootstrap and `deployment_admin` boundary notes, and the backup or restore text states that backup, restore, and restore verification remain deployment-local operator-facing concerns rather than public workbook route families.
18. Section 14 explicitly distinguishes implementation-facing harness mirrors from Core 05 claim-bearing publication requirements, names Core 05 as the owner for benchmark-profile and reproducibility rules, and mirrors the current fixture qualifiers plus `cartulary.perf.desktop_ref.v1`.
19. The frontend section contains the grid-engine selection gate and continues to name `react-data-grid` as the current baseline.
20. The frontend architecture section contains the grid adapter contract, capability allowlist, renderer/editor lifecycle rules, and grid-state refresh semantics.
21. `/packages/grid-adapter` appears in the workspace package baseline and owns selected-grid integration, coordinate translation, capability allowlisting, and imperative API fencing.
22. No frontend section allows row-position identity, direct vendor-grid mutation outside the adapter, or mutation routing from renderers and editors outside the adapter and sync engine.

### Verification

- A reviewer can inspect this guide alone and determine document authority, section class, profile gating, contract derivation direction, and module ownership without cross-file guesswork.
- A maintainer can update stack choices, commands, or package ownership in this guide without accidentally changing product behavior already owned by the normative core.
