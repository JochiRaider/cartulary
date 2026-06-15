# Cartulary Browser Design Readiness Workflow

**Status**: Implementation-support guide
**Authority**: Core 00 through Core 04 own product behavior. `docs/design.md`
and `docs/guides/cartulary-ui-ux-design-guide.md` own design direction.
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
make build-migrate
CARTULARY_CONFIG_FILE="$PWD/configs/dev/config.toml" \
  CARTULARY_POSTGRES_POSTGRES_PRIMARY_DSN="postgres://cartulary:cartulary@localhost:5432/cartulary?sslmode=disable" \
  ./migrate up
make dev
```

`make bootstrap` is needed only when the pinned local toolchain is not already
installed. `make db-up` starts local Postgres and object storage but does not
upgrade an existing database. Run the migration binary before `make dev` so the
database schema matches the current server code.

For a clean design-review database, reset and migrate the local database
instead:

```bash
CARTULARY_DESTRUCTIVE_CONFIRM=db-reset make db-reset
```

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
- Accessibility feel: keyboard traversal, visible focus, Esc behavior, control
  names, and non-color state cues.

Record findings in this shape:

| Field | Required content |
| --- | --- |
| Viewport | Exact viewport and browser zoom. |
| Surface or state | The visible workbook surface, system view, or state under review. |
| Evidence | Screenshot, trace, retained artifact path, or concise observation. |
| Owner reference | Relevant `docs/design.md`, UI/UX guide, visual golden guide, or frontend guide section. |
| Decision | `accept`, `change before MVP`, or `defer`. |
| Classification | `design issue`, `product bug`, `fixture issue`, `golden stale`, or `accepted intentional state`. |

## Playwright Fixture Review

Prepare the built browser evidence shape:

```bash
make build-web build-server build-migrate playwright-install
```

For exploratory headed review, use Playwright directly as a developer
convenience. This is not canonical retained evidence:

```bash
tmp/node-runtime/bin/pnpm --dir apps/web exec playwright test apps/web/e2e/workbook.visual.spec.ts -g "FE-V-P2-01" --headed --debug --workers=1
```

Use Playwright UI mode when browsing scenarios interactively:

```bash
tmp/node-runtime/bin/pnpm --dir apps/web exec playwright test --ui apps/web/e2e/workbook.visual.spec.ts
```

Useful scenario greps:

| Grep | Review target |
| --- | --- |
| `FE-V-P2-01` | Default Timeline shell. |
| `FE-V-P6-01` | Evidence state and affordance matrix. |
| `FE-V-P7-01` | Presence, conflict, and save-state fixtures. |
| `FE-V-P9-01` | Inspector, relationships, history, rollback, and errors. |
| `FE-V-P10-01` | Coordination surfaces and grid affordances. |
| `FE-V-P11-03` | Exposed `dark_graphite` token and theme states. |

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
- `frontend-row-accounting.json`;
- `frontend-accessibility-summary.json`.

If a complete frontend evidence packet is needed, run
`make frontend-evidence-audit` only after fresh retained roots exist for
`make check`, `make browser-e2e-support`, `make browser-e2e-visual`, and
`make browser-e2e-a11y`:

```bash
make frontend-evidence-audit \
  PHASE_NAMESPACE=frontend \
  PHASE=FE-P11 \
  CHECK_RESULTS_DIR=<check-root> \
  BROWSER_SUPPORT_RESULTS_DIR=<support-root> \
  BROWSER_VISUAL_RESULTS_DIR=<visual-root> \
  BROWSER_A11Y_RESULTS_DIR=<a11y-root>
```

If those roots do not exist, record that frontend evidence audit was skipped
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
- any skipped `make frontend-evidence-audit` run is explicitly recorded with
  the missing retained roots.
