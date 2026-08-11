# Cartulary Browser Design Readiness Workflow

**Status**: Implementation-support guide
**Authority**: Core 00 through Core 04 own product behavior. `docs/design.md`
is the sole normative design-direction owner.
`docs/guides/cartulary-ui-ux-design-guide.md` is non-normative design support.
`docs/testing-harness-nlspec.md` owns harness mechanics. This guide does not
define Base Profile conformance, extension-profile conformance, or Core 05
claim-publication evidence.

## Purpose

Use this workflow before MVP stand-up when the question is, "What does the
current UI actually look and feel like in a browser, and what design changes
are needed?"

The workflow is intentionally hybrid:

1. Inspect the running application manually and make design decisions from the
   live layout.
2. Use Playwright fixture states to revisit deterministic browser states.
3. Retain visual and accessibility artifacts after review.

Current visual goldens are coverage hints and regression inputs. They are not
design authority until reviewed and accepted by a human reviewer. Do not refresh
goldens during discovery. Refresh them only after a design decision accepts the
current layout or an intentional layout change.

## Review Boundary

This workflow produces design-direction evidence only. It does not prove Base
Profile conformance, release readiness, disconnected-profile behavior, Core 05
publication criteria, or extension-profile conformance.

Use this guide for:

- first-viewport layout review;
- density, spacing, typography, and status-state review;
- keyboard and focus feel;
- visual fixture triage before accepting or rejecting goldens;
- deciding which design changes block MVP stand-up.

Do not use this guide for:

- product conformance claims;
- route, schema, authorization, storage, or lifecycle ownership changes;
- snapshot update approval by itself;
- release verification;
- MVP deployment stand-up.

## Manual Browser Review

Run the local development stack from the repository root:

```bash
make bootstrap
make db-up
make db-migrate
make dev
```

`make bootstrap` is needed only when the pinned local toolchain is not already
installed. `make db-up` starts local Postgres and object storage but does not
upgrade an existing database. Run `make db-migrate` before `make dev` so the
database schema matches the current server code.

`make db-migrate` is the Make-owned wrapper around the migration application's
`migrate up` surface. It passes `CONFIG_FILE` through as
`CARTULARY_CONFIG_FILE`. For the default development config, which uses
`roots.database_storage.service_ref = "postgres_primary"`, the selected DSN
environment variables are
`CARTULARY_POSTGRES_POSTGRES_PRIMARY_MIGRATION_DSN` for `make db-migrate` and
`CARTULARY_POSTGRES_POSTGRES_PRIMARY_RUNTIME_DSN` for `make dev`. When the
selected variable is already set, the matching command preserves it. When it
is unset, the command derives only its purpose-specific local Compose DSN.

To use a non-default config or database for browser review, pass the same config
and selected managed-service DSN to migration and dev startup:

```bash
CARTULARY_POSTGRES_POSTGRES_PRIMARY_MIGRATION_DSN="postgres://migration-user:pass@db.example:5432/cartulary?sslmode=require" \
  CONFIG_FILE="$PWD/configs/dev/browser-review.toml" \
  make db-migrate
CARTULARY_POSTGRES_POSTGRES_PRIMARY_RUNTIME_DSN="postgres://runtime-user:pass@db.example:5432/cartulary?sslmode=require" \
  CONFIG_FILE="$PWD/configs/dev/browser-review.toml" \
  make dev
```

A historical-line or contaminated local database must be destroyed and
recreated before the current server can start. The v2 cutover has no
export/import transition or compatibility bridge. For a clean review database
on the default local Compose Postgres service, reset and migrate the local
database:

```bash
make db-reset CARTULARY_DESTRUCTIVE_CONFIRM=db-reset
```

The destructive confirmation is a Make command-line input for the target, not a
leading environment assignment.

`make dev` starts the Go server and Vite dev server for manual inspection. The
browser URL is:

```text
http://127.0.0.1:5173
```

Use the local bootstrap admin credentials from
`configs/dev/bootstrap-admin.json`:

```text
email: dev-admin@example.test
password: DevBootstrap1!
```
oathtool --totp -b 'LZNWD7TWKSM2IFYZS42C7FFIHUNITKIA'

Complete any prompted TOTP enrollment before starting layout review. Inspect at
100 percent browser zoom. Use browser device emulation or window sizing for
these viewports:

| Viewport | Purpose |
| --- | --- |
| `1440x900` | Default visual-shell fixture and first-viewport workbook review. |
| `1280x720` | Base design viewport and required shell-control reachability. |
| `1024x720` | Narrow desktop overflow and inspector behavior. |
| `768x640` | Compact desktop overflow and degraded-density behavior. |

Review these surfaces before judging changes:

- Default Timeline shell: top bar, built-in tabs, `System views`, view bar,
  grid, inspector, and status strip.
- Workbook interactions: row creation, inline edit, paste, save-state changes,
  conflict state, filters, grouping, and saved views.
- System views: Indicators, Compromise Assessments, Task Requests, Decisions,
  Parties, Communications Log, Handoff, Status Review, and Lesson.
- State-heavy areas: evidence affordances, mention chips, conflict resolver,
  rollback and history, destructive confirmations, and successful empty query.
- Inspector feature registry states: default-closed inspector,
  `no_row_selected`, Details, Relationships, Evidence, History, Workflow
  panels, create-related workflow actions, merge confirmation, rollback
  preview, supersede confirmation, blocked evidence preview, handoff
  acknowledgement, status-review blocked-work review, and stale-state
  invalidation after row change.
- Accessibility feel: keyboard traversal, visible focus, Esc behavior, control
  names, and non-color state cues.

A missing inspector-heavy visual state is a fixture issue when the Core-owned
behavior is already implemented and not represented; it is a product bug when
the behavior violates Core 01/Core 03/Core 04; it is a design issue when the
behavior satisfies Core but violates `design.md`. A mismatch with the UI/UX
guide is a guide-maintenance issue unless it also conflicts with an owner.

Record findings in this shape:

| Field | Required content |
| --- | --- |
| Viewport | Exact viewport and browser zoom. |
| Surface or state | The visible workbook surface, system view, or state under review. |
| Evidence | Screenshot, trace, retained artifact path, or concise observation. |
| Owner reference | Relevant Core section or `docs/design.md` section; supporting-guide references may be recorded separately but are not owners. |
| Decision | `accept`, `change before MVP`, or `defer`. |
| Classification | `design issue`, `product bug`, `fixture issue`, `golden stale`, or `accepted intentional state`. |

## Playwright Fixture Review

Prepare the built browser evidence shape:

```bash
make build-web build-server build-migrate playwright-install
```

Browser execution requires the same Make-owned isolated service lifecycle and
validated v4 attachment as retained evidence. Run the applicable target and
inspect its retained Playwright report, screenshots, traces, and session logs:

```bash
make browser-e2e-visual
```

Raw `playwright test`, Playwright UI mode, and IDE launchers are unsupported
because they do not own or validate an isolated browser session. They are not
equivalent developer-convenience evidence paths. A future interactive surface
must be Make-owned and establish the same attachment contract before enabling
Playwright UI or debug execution.

Useful scenario greps:

| Grep | Review target |
| --- | --- |
| `visual.workbook-shell.row-01` | Default Timeline shell. |
| `visual.evidence-workflow.row-01` | Evidence state and affordance matrix. |
| `visual.collaboration.row-01` | Presence, conflict, and save-state fixtures. |
| `visual.inspector-history.row-01` | Inspector, relationships, history, rollback, and errors. |
| `visual.coordination-review.row-01` | Coordination surfaces and grid affordances. |
| `visual.design-readiness.row-03` | Exposed `dark_graphite` token and theme states. |

When an exploratory run exposes a mismatch, classify it before changing
anything:

- Design issue: the live UI conflicts with accepted design direction.
- Product bug: behavior breaks Core-owned runtime semantics.
- Fixture issue: seed data, masking, viewport, scroll normalization, or
  screenshot scope is wrong.
- Golden stale: the UI is intentionally correct but the committed golden is
  old.
- Accepted intentional state: no change needed.

## Retained Evidence

After manual review identifies the states worth retaining, run the canonical
browser evidence targets:

```bash
make browser-e2e-visual
make browser-e2e-a11y
```

Inspect retained artifacts under the reported run roots:

```text
.cartulary/test-results/<run-id>/.../playwright-output/
```

Prioritize these artifacts:

- actual and diff PNGs from visual failures;
- Playwright traces when present;
- `*-render-diagnostics` attachments;
- `*-grid-diagnostics` attachments;
- font manifest digest attachments;
- `test-evidence-accounting.json`;
- `frontend-accessibility-summary.json`.

If a complete owner evidence packet is needed, run
`make test-evidence-audit` only after fresh retained roots exist for
`make check`, `make browser-e2e-support`, `make browser-e2e-visual`, and
`make browser-e2e-a11y`:

```bash
make test-evidence-audit \
  OWNER=web.workbook \
  CHECK_RESULTS_DIR=<check-root> \
  BROWSER_SUPPORT_RESULTS_DIR=<support-root> \
  BROWSER_VISUAL_RESULTS_DIR=<visual-root> \
  BROWSER_A11Y_RESULTS_DIR=<a11y-root>
```

If those roots do not exist, record that owner evidence audit was skipped
because the workflow was still in design-discovery mode.

## Acceptance

The pre-MVP browser design review is complete when:

- every reviewed issue has a viewport, surface or state, evidence reference,
  owner reference, decision, and classification;
- no decision relies only on an existing golden image;
- the default Timeline shell review confirms that an admin or control card
  stack does not dominate the first viewport above the active grid;
- `make browser-e2e-visual` either passes or has reviewed diffs with clear
  design, product, fixture, or stale-golden classification;
- `make browser-e2e-a11y` either passes or has actionable keyboard, focus,
  accessible-name, contrast, or state-communication findings;
- any skipped `make test-evidence-audit` run is explicitly recorded with
  the missing retained roots.
