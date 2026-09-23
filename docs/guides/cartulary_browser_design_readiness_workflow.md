# Cartulary Browser Design Readiness Workflow

**Status**: Implementation-support guide. Reviewed against the September 2026
application and harness owners.

**Authority**: Adopted subsystem NLSpecs and Core 00–04 own product behavior.
[Design direction](../design.md) owns presentation within its stated boundary;
[domain vocabulary](../domain.md) owns vocabulary and owner navigation.
The [Testing Harness NLSpec](../testing-harness-nlspec.md) owns execution and
artifact mechanics. This guide supports human design review; it does not establish
Base/extension conformance, release readiness, or Core 05 publication claims.

## Prepare and start a review

Run from the repository root on the supported Linux/WSL2 environment with Docker
available. Use `make doctor` to diagnose prerequisites and `make bootstrap` when
the pinned toolchain and dependencies need installation. Discover current commands
with `make help` and `make help-all`.

```bash
make browser-design-review
```

This helper prepares the production frontend and server harness, obtains an
isolated service suite, creates a private database and object namespace, validates
the current browser attachment, and seeds synthetic review data. It prints the
actual loopback URL, a scenario index, sample-file directory, private login
instructions, and retained diagnostics. Use that URL, including its allocated
port. The helper stays running until Ctrl-C.

The default is `REVIEW_PROFILE=network_flow_claimed`. It claims Import, Incident
Portability, Reference Pack, Snapshot/Reporting, and Network Flow Activity. The
comparison session leaves Network Flow unclaimed:

```bash
make browser-design-review REVIEW_PROFILE=default
```

`default` retains the other four claims; it is not a Base-only deployment.
Enterprise Authentication is unclaimed in both review sessions. A profile claim,
packaged client support, visible entry point, and user authorization are separate
facts. Do not infer availability from a configuration flag alone.

Each invocation starts fresh. It does not use or reset the Compose development
database. Ctrl-C closes owned processes and services and removes private runtime
material, including credentials. Retained diagnostics and synthetic sample files
remain under the reported `.cartulary/test-results/design-review-...` directory.
Startup and seed failure also trigger cleanup. Restart after changing source to
review a new build: a running session continues serving its own sealed artifact.

For automated setup verification, use the finite helper:

```bash
make browser-design-review-smoke
make browser-design-review-smoke REVIEW_PROFILE=default
```

These helpers verify the workbook, Evidence upload/download, workbook CSV import,
reference-pack import, incident-bundle handoff, zero/one/many incident navigation,
role-dependent controls, and claimed/unclaimed Network Analysis before cleanup.
They do not close product verification rows or replace a human design decision.

### Sign in and choose a scenario

Open the printed private `access.json` locally. It contains the bootstrap admin's
session-generated authenticator setup key and dedicated review-user credentials.
Add that key to a TOTP authenticator; use its current code at sign-in. There is no
committed review TOTP secret and no required `oathtool` installation. Do not copy
credentials or the private access file into findings, screenshots, or commits.

Use separate browser profiles or private contexts to compare roles:

| Account | Starting context |
| --- | --- |
| Bootstrap admin | Deployment administration and admin access to populated/empty review incidents. |
| Review editor | Two visible incidents, writable workbook, no deployment administration. |
| Review viewer | One visible incident, read-only workbook. |
| Review empty | Zero visible incidents; successful empty directory and ordinary creation entry. |

All cardinalities remain in the incident directory until explicit Open. The
populated incident includes Timeline rows, revision history, coordination records,
Evidence, and a shared saved view. The claimed profile also contains a Network
Analysis table imported through its production workflow. Follow the scenario index
for URLs and use the application's navigation for Account settings, Incident
controls, and inspector sections.

Sample files include a Timeline CSV, an XLSX with partial-import cases, the
owner-maintained minimal Network Flow CSV, a verified-format reference pack,
a current-format incident bundle, and a text Evidence file. The bundle contains
an independent synthetic incident; importing it changes the admin's visible
incident count. Repeated imports or mutations may intentionally change review
state. Restart the session to recover the initial scenario.

### Existing development environments

For iterative development with persistent data, follow the
[development guide](cartulary-dev-guide.md). `make db-up`, `make db-migrate`, and
`make dev` retain their separate meanings. Migration uses the selected migration
DSN; the server uses the selected runtime DSN. Use the same configuration and
managed-service binding for both. Follow their current task contracts instead of
copying a credential-bearing example into this guide.

An existing database must meet the current migration-lineage contract. Reset is
an explicit destructive operation, not a routine readiness prerequisite. The
isolated review helper avoids that decision entirely. A dev server is not a valid
attachment for canonical browser evidence.

## Review the current application

Start at `1440x900`, 100% browser zoom, `dark_graphite`, and a cleared account
density override. Timeline's resolved default is compact. Review `1280x720`,
`1024x720`, and `768x640`, then short-window and below-minimum-width cases that
exercise focused editors and recovery controls. Repeat relevant states at 200%
zoom, supported text spacing, and reduced motion. Keep browser zoom distinct from
a fixture's explicitly declared CSS zoom.

Select compact, default, and comfortable through Account settings; also test
clearing the override. There is one exposed theme, not a theme-selection workflow.
Use exact fixture profiles for retained captures rather than assuming every
scenario shares the initial viewport or density.

| Review area | Observe and exercise |
| --- | --- |
| Authentication and root | Local sign-in, authenticator setup, validation, focus, provider-discovery failure, zero/one/many incidents, editable search during refresh, previous results, continuation, and creation recovery. |
| Account and administration | Profile, Appearance, Security, modal containment, Users, dirty departure, Reference packs, Incident import, and Administrative audit. Keep deployment and incident authorization distinct. |
| Incident controls | Metadata, workbook preferences, membership/audit, lifecycle, import entry, retained operations, and permission loss. |
| First viewport | Grid prominence; compact view bar; navigation; selected/focused cell; closed inspector with reachable opener; status and presence without competing control-card stacks. |
| Grid and working set | Single-click and keyboard entry, range selection, Find/match navigation, copy/paste, clearing, fill-down, correction access, row creation, batch outcomes, and exact History navigation. |
| Layout and views | Column resizing, frozen columns, saved layouts, saved-view discovery/actions, sorting/filtering/grouping, query replacement/continuation, empty results, scroll continuity, and retained authorized rows during refresh. |
| Inspector | Persistent Sections navigation, one body scrollport, complete saved values, narrative expansion, attached editing, explicit Update/Close editor, retained work after detachment, and row retargeting. |
| Relationships and Evidence | Stable source identity, candidate discovery, collection overflow, mentions, local attachment feedback, accepted information versus access, upload/preview/download states, and recovery. |
| History and Workflow | Readable attribution and UTC times, continuation, rollback/restore, canonical action order, contextual creation, explicit operation scope, acknowledgement, and refresh failure after accepted writes. |
| System views | Indicators, Compromise Assessments, Task Requests, Decisions, Parties, Communications Log, Handoff, Status Review, and Lesson; include supported create paths and read-only roles. |
| Collaboration and lifecycle | Presence accuracy, save/conflict states, uncertain replay, duplicate activation, detached operation completion, incident access loss, account-session loss, and clearing protected content before root navigation. |
| Accessibility | Keyboard-only use, Escape priority, visible focus, focus return, accessible names, non-color cues, one announcement per event, native modal containment, reachable correction actions, zoom, spacing, and motion. |

For every asynchronous flow distinguish initial loading, successful empty,
filtered empty, retained stale results, failed initial load, accepted mutation with
failed refresh, uncertain mutation, and authorization loss. Refresh recovery must
not resend an accepted write. Existing fixtures are the repeatable route to states
that cannot safely or reliably be created during ordinary manual interaction.

## Cover all six current extension profiles

Core 00 §4.2 recognizes exactly six profiles. The following is review navigation,
not a replacement profile or test-row registry. Use current owner catalogs to
select evidence; source modules and verification owner IDs can differ.

| Profile and owner | Prerequisites and browser entry | Review and evidence route | Limitations |
| --- | --- | --- | --- |
| Import — Core 01 import owner | Claimed, supported client, writable incident; Incident controls → Import assistant. | CSV/XLSX upload, discovery, mapping/preview/approval, selected units, apply, partial outcomes, cancel and exact recovery. Navigate with `module.imports` and its current browser rows. | Workbook import is distinct from whole-incident portability and analytical import targets. |
| Incident Portability — Core 01 portability owner | Claimed; deployment-admin import from Deployment administration → Incident import. | Real bundle admission, queued/running/terminal jobs, cancellation, observation recovery, access confirmation and Open. Use `module.incidentbundles` and incident-import browser rows. | Backend export and descriptor coverage do not imply an export browser UI or permission to access every incident. |
| Reference Pack — Core 01 reference-data owner | Claimed; Deployment administration → Reference packs. | Catalog/search/filtering, staged import, activate/disable/reverify, exact-scope refresh, retained jobs and recovery. Use `module.reference_data`. | The subsystem NLSpec remains draft. Its proposed formats and workflows are not current setup requirements. |
| Network Flow Activity — adopted Network Flow and Graph Projection NLSpecs | Claimed with Import dependency and distinct generated key material; Network Analysis tab. | Import/mapping, tables, query pages, temporal graphs, contributors, saved graphs, explicit indicator linking, local selection/focus and stale recovery. Use `module.networkflow`, `web.networkflow`, and applicable graph owner evidence. | The `default` comparison omits this workspace. Graph labels and coordinates do not replace semantic identities. |
| Enterprise Authentication — Core 01/Core 04 | Presentation fixtures, or a separately configured claimed deployment with valid OIDC/SAML provider manifest. Anonymous sign-in and existing Account/Users controls. | Provider discovery/begin, local fallback, session/root convergence, accessible errors and recovery. Use `module.auth` and its enterprise browser selectors. | Existing presentation fixtures simulate provider responses. They do not exercise a real IdP or prove OIDC/SAML verification. |
| Snapshot/Reporting — adopted Reporting and Report Composition NLSpecs | Claimed, valid template/source inputs, and authorized incident access. | Use `module.reporting` and `module.reportcomposition` owner evidence for snapshots, compositions, immutable versions, render/release and outputs. Record applicable output inspection separately. | The current frontend has no report-composition workspace. A passing backend slice or an API resource is not browser design coverage; record this product-surface gap. |

### Optional real enterprise authentication

Use a separate development/deployment configuration and the current Enterprise
Authentication sections of Core 01/Core 04. Set the claim and provider-manifest
path together, provide provider secret references through the configured secret
mechanism, and configure the IdP's registered redirect/ACS origins to match that
deployment's public origin. Migrate with its migration binding, then start with
its runtime binding. Preserve local bootstrap recovery.

The isolated helper deliberately provides neither an IdP nor provider secrets.
Record provider type, non-secret provider identity, origin, and which round-trip
was exercised. Never describe simulated discovery/begin fixtures as successful
real-provider login. Keep external-provider credentials out of retained artifacts.

## Reproduce states and retain evidence

Discover selection and ownership before running broad suites:

```bash
make task-guide ROLE=module-author OWNER=web.design
make explain-target TARGET=browser-e2e-visual DETAIL=rows
make explain-test-owner OWNER=module.auth
```

The current graph runner launches `node` by name. If the pinned runtime is installed
but `node` is absent from the shell's PATH, prefix these verification commands with
`PATH="$PWD/tmp/node-runtime/bin:$PATH"`. The review helper itself uses the pinned
runtime directly. Record this environment adjustment with the run.

Choose exact current row IDs from that output and use
`make service-backed-test-slice OWNER=<owner> ROWS=<row-id,...>` for browser rows.
Use `make test-slice` for applicable non-service owner checks. Historical title
fragments and screenshot filenames do not define row ownership. Do not add a
parallel list of catalog IDs to this guide.

Canonical retained visual/accessibility runs remain:

```bash
make browser-e2e-visual
make browser-e2e-a11y
```

Their work graphs own builds, service sessions, profile separation and cleanup;
there is no need to precede each run with a copied build-target list. Playwright
workers validate the v7 stack, runtime identities, live process proofs, and the
sealed frontend receipt before assertions. Raw Playwright test/UI/IDE launchers
and an already-running dev server cannot substitute for that attachment.

Inspect the printed run root with `make explain-run RESULTS_DIR=<run-root>`.
Follow summary artifact references to actual/diff PNGs, traces, session startup
diagnostics, grid/render diagnostics, font digest, row accounting and accessibility
summary. Preparation failures precede screenshot interpretation. The review
helper's screenshots and scenario index are manual inspection support; they are
not catalog-row success records.

Use the [golden maintenance guide](cartulary_visual_golden_maintenance.md) for the
pinned renderer, capture intent v2, registry v6, reconciliation v3, golden manifest,
and accepted refresh procedure. Catalog visual rows are the capture inventory;
the design registry names only its declared semantic fixtures. An unregistered
capture is not automatically an orphan. Do not update goldens during discovery.
Reviewed goldens remain regression inputs, not design authority.

An owner evidence audit is optional for discovery. When needed, discover its
current inputs with `make explain-target TARGET=test-evidence-audit DETAIL=summary`
and supply the required successful retained roots for the chosen owner. Report
missing roots and skipped audits explicitly; do not create release claims from
an incomplete evidence packet.

## Record findings and accept the review

Keep a review record alongside the work's handoff. Record the source revision and
working-tree status, run root, runtime profile and observed claims, actor role,
exact viewport, browser/zoom, density override and resolved density, motion/spacing
settings, surface identity, reproduction steps, observation, owner reference,
evidence location, classification, disposition, and responsible follow-up.

Classify findings as `design issue`, `product bug`, `fixture issue`, `golden stale`,
`guide/setup issue`, `coverage gap`, or `accepted intentional state`. Distinguish a
confirmed defect from a hypothesis. Identify evidence as live application,
simulated presentation, canonical visual, accessibility, or service/API evidence.
Choose `accept`, `change before readiness acceptance`, or `defer` with rationale.
The earlier phrase “before MVP” is a review milestone, not a universal gate.

The review is complete when every applicable checklist/profile entry has evidence
or an explicit limitation; actionable findings have a disposition; the first
viewport preserves the grid-first contract; visual differences and accessibility
failures are classified; and skipped checks are explained. A missing browser
surface remains a coverage gap even when its backend passes. Actual readiness
acceptance is a human decision. Neither a successful helper nor a golden refresh
establishes that decision automatically.

The [overhaul record](../handoffs/browser-design-readiness-overhaul.md) tracks this
workflow's implementation, verification and known limits. Consult it for setup
validation, not as a permanent product-readiness certificate.
