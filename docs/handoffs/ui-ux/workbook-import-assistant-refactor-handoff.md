# Workbook Import Assistant lifecycle and recovery handoff

## Baseline, authority, and scope

- Baseline: clean `main`, `a5bf1748c6aa9825ebe9e85206e7a197e496bcbf`;
  revalidated before implementation on 2026-09-08. No unrelated changes exist.
- Read AGENTS.md and the digest's localized read order during planning. The
  digest is unchanged advisory review support. Its accessibility, local error,
  recovery, and containment concerns are ADOPT/ADAPT; it creates no product policy.
- Core 01 §§3.3.9.1, 17.1, 17.2 owns requests, receipts, cancellation, mapping
  transport, and durable resources. Core 02 §7.2 owns import identity/provenance.
  Core 03 §11.2 owns approval, persisted selection, overlap, deterministic apply,
  duplicate detection, and outcomes; §4.3.2 owns closed-incident interaction.
- Core 04 §2 and REQ-04-053 own current authorization and inert hostile input.
  Extensions §18 (especially EXT-REQ-201/212) owns fencing and cleanup. Network
  Flow §10 owns analytical preview/approval fingerprint parity. Design §§5, 7,
  8, 12, 14 owns compact accessible presentation. domain.md supplies vocabulary;
  nlspec-spec.md is research support, not product authority.
- Completed incident-import, Network Analysis, and Imports-module handoffs were
  inspected as implementation evidence only. Those projects are not reopened.
- Allowed changes: frontend import workflow, adapters, narrow application and
  incident-control lifetime bindings, necessary Network Flow caller adaptations,
  focused tests, authored routing and generated routing, narrow Imports public
  cancellation-state corrections, and this handoff.
- Excluded: parser/kernel/backend scheduler/source-owner redesign; incident
  bundles, clipboard, new public routes/targets/storage/dependencies, generic
  workflow engines, compatibility facades, commits, resets, pushes, deployment.

## Ordered tracker

| Workstream | Status | Exit |
| --- | --- | --- |
| WI-01 Baseline and characterization | DONE | Verified gaps, owner map, focused reproductions |
| WI-02 Transport and observation | DONE | Projection dependency regenerated and transport checks pass |
| WI-03 Workflow ownership | DONE | Controller reconciliation and deterministic lifecycle tests pass |
| WI-04 Presentation and browsers | DONE | Five Import browsers, Network Flow browser, and reviewed containment evidence pass |
| WI-05 Final validation | DONE | All required checks, binary acceptance, and final scope audit pass |

Never advance a blocked dependency. An adopted-owner contradiction is recorded
as `BLOCKED: owner contradiction` and requires direction.

## Findings and owner-to-change map

| Demonstrated gap / reproduction | Owner | Correction boundary |
| --- | --- | --- |
| Discovery submission succeeds, then polling/preview fails: no receipt survives | Core 01 job and Import contracts | Typed transport and import controller |
| Approve mapping, fail select: only a boolean survives; returned mapping and selection envelope are discarded | Core 01 REQ-01-474; Core 03 REQ-03-172/178 | Separate draft, approved unit, and session selection |
| Close/reopen drawer after upload/apply: feature unmount destroys all local job/file state | Core 03 REQ-03-168; Extensions cleanup matrix | Controller above presentation with application authority binding |
| Role loss/incident closure after discovery: later handlers do not recheck current authority | Core 04 §2; Core 03 §4.3.2 | Synchronous admission and dispatch fencing |
| Cancel twice, or cancel a nonterminal non-cancelable job: current control permits duplicate requests | Core 01 §3.3.9.1 | Independent cancel intent, current read, exact replay |
| Slow job exceeds 30 polls: coordinator reports timeout as failure without resume | Core 01 job status vocabulary | Bounded resumable observation |
| Partial apply: links come from submitted draft targets, not durable applied units | Core 03 REQ-03-098/179/184 | Outcome projection and validated semantic navigation |
| Duplicate headers suggest duplicate fields; eligibility uses direct_value only | Core 01 REQ-01-474 | Draft model consuming generated createWritable/binding facts |

Structural movement: transport, state transitions/execution, mapping drafts,
outcome projection, React binding, and presentation become separate cohesive
import-owned concerns. Source parsing, kernel algorithms, fingerprints, and
analytical owner semantics remain in their existing owners.

## Work log

### WI-01

Inspected the four primary frontend files, drawer/shell/App bindings, generated
view/target projections, authored Imports and common-job OpenAPI schemas,
Imports schemas, Network Flow import controller, current tests and routing.
The reproduced sequences above follow the current handlers' receipt discard,
unmount cleanup, and missing dispatch gates. Existing tests characterize
skip/reselect, durable regions, typed target consumption and Network Flow parity.

Current `make task-guide ROLE=module-author OWNER=module.imports` and
`make task-guide ROLE=module-author OWNER=module.jobapi` succeed and route to
owner slices. Additional discovered owners are web.application, web.workbook,
web.networkflow, and package.ui when shared selectors change.

Exact-source planning evidence:

- Focused Imports frontend `make test-slice OWNER=module.imports ROWS=...`:
  PASS 6/6, `.cartulary/test-results/20260908T201919Z-p2220331`.
- Shared coordinator `make test-slice OWNER=web.application
  ROWS=web.application.regression.importcoordinator_suite_fd42b7df92`:
  PASS 2/2, `.cartulary/test-results/20260908T202023Z-p2222131`.
- `make test-slice OWNER=module.jobapi`: PASS 4/4,
  `.cartulary/test-results/20260908T202023Z-p2222136`.
- `git diff --check`: PASS. No service-backed/browser result is inferred from
  these unit runs.

Decisions: one current workbook workflow; secure immutable write attempts;
drawer closure continues the current bounded observation window; volatile
known-resource recovery only; no import-history or reload-recovery promise.
Cancellation follows common-job authorization separately from source writes.
Generated createWritable already projects writable OR create_writable.
No adopted-owner contradiction was found.

WI-01 exit: PASS. Next action: typed transport and observation.

### WI-02

Added importRequests (immutable bytes/JSON attempts), importClient (typed
submission and separate resource reads), importJobContract (cross-resource job
validation), and importJobObservation (bounded serial windows). Expanded
importContractAdapter aliases without changing public schemas. Added exact-byte,
CSRF, structured rejection, malformed success, scope, mapping, pagination,
preview-bound, cancellation-state, hung-read, resume, late-response and terminal
observation tests. Existing coordinator consumers remain until their ordered
integration; the old composites will be removed, not retained as aliases.

- `make frontend-typecheck`: PASS 2/2,
  `.cartulary/test-results/20260908T203351Z-p2243984`.
- `make frontend-import-boundary-check`: PASS 2/2,
  `.cartulary/test-results/20260908T203403Z-p2244485`.
- `make test-slice OWNER=module.imports
  ROWS=module.imports.frontend.typed_transport,module.imports.frontend.bounded_observation`:
  PASS 3/3 execution units (10 tests),
  `.cartulary/test-results/20260908T203451Z-p2245976`.
- Earlier routing preflights rejected suite-only titles and unsorted titles;
  authored selectors now enumerate exact sorted tests. The initial replay test
  reused a consumed Response body; corrected the fixture to create each response.
  Related failed run: `.cartulary/test-results/20260908T203414Z-p2245121`.

WI-02 exit: PASS. Next action: workflow owner and durable reconciliation.

WI-02 dependency reopened during WI-03 type integration: authored OpenAPI
ImportSession omitted `created`/`mapped` and admitted retired `uploaded`/
`discovering`; source-column empty policies admitted retired values instead of
`write_null`; unit state incorrectly admitted `canceled`; action/preview status
projections were unconstrained strings. Core 03 REQ-03-164/172/188 establishes the exact vocabulary. The
backend agrees for mapping and session states; WI-04 subsequently found a
cancellation-only unit serialization gap, corrected below. This is projection drift, not an owner contradiction.
Repairing these exact enum projections through `make generate` is necessary for
correct durable reconciliation. No route or stored data is added. WI-03 is
paused until this dependency and transport checks pass.

Dependency resolved: authored release change-set entries describe the exact
Imports enum corrections; unrelated entries and immutable baseline are retained.
The compatibility checker classifies session/empty-policy corrections as
breaking projection differences against the published 1.0.0 baseline within the
existing 2.0.0 release. Runtime requests, routes, and stored data are unchanged.
`make generate` initially failed the stale fingerprint check at
`.cartulary/test-results/20260908T204359Z-p2249057`; after updating authored
declarations it passed at `.cartulary/test-results/20260908T204851Z-p2251590`.
`make frontend-typecheck` passed 2/2 at
`.cartulary/test-results/20260908T204909Z-p2254890`; the same WI-02 transport and
observation slice passed 3/3 at
`.cartulary/test-results/20260908T204909Z-p2254824`.
WI-02 exit reconfirmed: PASS. Next action: resume WI-03 controller integration.

### WI-03

Added WorkbookImportController, workbookImportState, workbookImportMapping, and
the generated-target/view-contract join. The controller owns admission, bounded
execution, current and late receipts, independent selection and cancellation,
authority fencing, preview queue, resource reconciliation and navigation review.
Draft suggestions/duplicate-field feedback and durable outcome projection are
separate modules. Mapping algorithms and fingerprints remain server-owned.

`make test-slice OWNER=module.imports
ROWS=module.imports.frontend.lifecycle,module.imports.frontend.typed_transport,module.imports.frontend.bounded_observation`
passed 4/4 execution units (22 tests) at
`.cartulary/test-results/20260908T205359Z-p2258927`. Coverage includes preserved
approval after selection failure, exact upload replay, current late acceptance,
replaced lifetime, drawer closure, empty discovery, serial previews, durable
regions, changed selection, overlap, stale reads, role/closure/claim gates,
authentication pause, cancellation races and retained applied-unit navigation.
Earlier cancellation fixture returned the discovery job ID for an apply cancel;
the correlation check correctly rejected it. Corrected fixture run passed at
`.cartulary/test-results/20260908T205320Z-p2257850`.
Type-only fixture corrections passed `make frontend-typecheck` (2/2) at
`.cartulary/test-results/20260908T205453Z-p2260409`; preceding failed type runs
are `.cartulary/test-results/20260908T205227Z-p2256644` and
`.cartulary/test-results/20260908T205358Z-p2258541`.

WI-03 exit: PASS. Next action: bind application lifetime, replace presentation,
adapt the shared coordinator consumer, then collect browser evidence.

### WI-04

Replaced ImportAssistantFeature with compact Source, Units and mapping, Apply,
and Outcomes presentation plus ImportUnitCard and ImportOperationNotice.
App/useWorkbookImport binds the controller to the application session; the
workbook hook supplies incident lifecycle, role, extension availability, and
access confirmation. Closing the drawer retains the controller and current
observation window; changing authorization generations fences callbacks.
The coordinator now contains only explicit submission, bounded observation, and
resource loading operations. Its former composite interface is removed.
Network Flow retains candidate construction, preview decoding, fingerprint
comparison, analytical publication and navigation in its own owner.

Observed additional correction: Core 01 §§3.3.9.1/17.2 and Core 03 REQ-03-172
agree that canceled jobs use `job_canceled` and public units have no `canceled`
state. The existing Imports cancellation test asserted a family code in the
job table and only private outcome rows. The actual HTTP projection exposed
that drift. Imports now publishes the common job code, records stopped unit
processing as `failed`, and reads retained older unit records with that same
public state. Private cancellation outcomes, reason codes, committed effects,
and session `partially_applied` remain intact. The existing service test now
also reads the terminal job and both unit resources through authenticated HTTP.
This is a narrow owner-backed correction, not a scheduler or outcome-journal
redesign. No adopted-owner contradiction was found.

Verification and artifacts:

- `make test-slice OWNER=module.imports ROWS=<all nine frontend rows>`:
  lifecycle, transport, observation, presentation and region tests pass;
  `.cartulary/test-results/20260908T213230Z-p2510584` was 9/10 because the
  static target-consumption test expected the old component-local import.
  Its boundary assertion now follows the adapter/mapping separation; the exact
  failed row passes 2/2 at
  `.cartulary/test-results/20260908T213351Z-p2628191`.
- Lifecycle includes exact apply/cancel replay and a 24-case stage matrix:
  upload/mapping/select/region/apply/cancel crossed with role loss, closure,
  claim withdrawal, and incident replacement. Deferred transport responses
  cannot restore retired state. Authentication pause and session replacement
  have separate tests.
- `make test-slice OWNER=web.application
  ROWS=web.application.regression.workbook_import_lifetime`: PASS 2/2,
  `.cartulary/test-results/20260908T211801Z-p2397797`.
- Shared coordinator application row: PASS 2/2,
  `.cartulary/test-results/20260908T210615Z-p2266370`.
- `make test-slice OWNER=web.networkflow`: PASS 39/39,
  `.cartulary/test-results/20260908T211802Z-p2399314`.
- `make test-slice OWNER=module.imports ROWS=<all five browser rows>`:
  PASS 12/12 execution units,
  `.cartulary/test-results/20260908T213240Z-p2520741`. These are the authored
  base-profile absence, claimed production CSV, density_containment,
  keyboard_recovery, and partial_outcomes rows.
- `make test-slice OWNER=module.networkflow
  ROWS=module.networkflow.browser.verify_claimed_network_analysis_discovery_import_f977242343`:
  PASS 11/11, `.cartulary/test-results/20260908T213242Z-p2521155`.
- `make service-backed-test-slice OWNER=module.imports ROWS=<ten affected
  integration rows>`: PASS 4/4,
  `.cartulary/test-results/20260908T211926Z-p2455221`. Rows cover current-state
  apply admission, CSV exhaustive mapping, explicit new-session re-import,
  cancellation, overlap/reselection, atomic rollback, exact upload replay,
  bounded XLSX, durable region replay, and Network Flow atomic unit commit.
- Public cancellation correction rerun:
  `make service-backed-test-slice OWNER=module.imports
  ROWS=module.imports.integration.partial_cancellation_outcomes_a14b66f830`:
  PASS 3/3, `.cartulary/test-results/20260908T213230Z-p2510601`.
- `make service-backed-test-slice OWNER=module.jobapi`: PASS 3/3,
  `.cartulary/test-results/20260908T211928Z-p2455440`.
- `make frontend-typecheck`: PASS 2/2,
  `.cartulary/test-results/20260908T213230Z-p2510716`.
- `make frontend-import-boundary-check`: PASS 2/2,
  `.cartulary/test-results/20260908T211335Z-p2339055`.
- `CARTULARY_GENERATE_DRIFT_REFRESH=1 make generate-drift`: PASS 4/4,
  `.cartulary/test-results/20260908T213145Z-p2502703`. Authored owner rows drive
  generated browser grouping; runtime/tests do not consume documentation.

Browser review: inspected all five new evidence PNGs below the successful
Imports run's
`browser-e2e-webserver-backed/browser-groups/functional-support-network-flow-claimed-import-assistant/playwright-output/`:
`import-compact-narrow.png`, `import-default-narrow.png`,
`import-comfortable-narrow.png`, `import-recovery-zoom.png`, and
`import-partial-outcomes.png`. These show native workbook tokens, visible focus,
390px containment, long wrapping headers, bounded preview content, 200% zoom,
and a distinct applied Timeline / failed Evidence result with only Timeline
navigation. The keyboard row also tests short layout, reduced motion, Escape,
focus restoration, same-job resume, and selection-only recovery after closure.
The XLSX fixture is deterministic authored test data with an invalid Evidence
date; its Timeline unit commits through the real service.

The visual maintenance guide was read. No existing golden captures the new
Import drawer states and no committed golden was changed or promoted. New PNGs
are retained browser review evidence, not claim-bearing visual goldens.
The ordinary comparison result is recorded in WI-05.

Earlier implementation failures were corrected rather than accepted: legacy
shell test props and malformed job fixtures (type/NF tests); expected text tied
to the old combined approval state; the XLSX field locator (now uses its exact
accessible role/name); missing generated browser groups; and an incorrect
application test family selector. The failed partial browser run is
`.cartulary/test-results/20260908T211429Z-p2347319`; the final five-browser run
above supersedes it. Formatting found hook dependency and accessibility lint
issues, corrected using the existing hook and scroll-region conventions.

WI-04 exit: PASS. Next action: agent-finalize and final owner/readiness checks.

### WI-05

Final review tightened immutable-source correlation on session/unit refresh and
rejects ready/applied unit resources without durable approval. It also retires
unit-specific attempts after a unit disappears, removes files/drafts when the
session is unavailable, and confirms authorization after a resource-specific
404 without assuming all incident access was revoked. Existing receipt identity,
local denial, and lifecycle tests cover these cases. A first contradictory-source
test accidentally used the fixture's unchanged digest; corrected it to a distinct
digest. Related failed run: `.cartulary/test-results/20260908T214025Z-p2782372`;
the lifecycle rerun passed at
`.cartulary/test-results/20260908T214201Z-p2784307`.

`make agent-finalize` passed before broader verification at
`.cartulary/test-results/20260908T213653Z-p2633873`, and again after final
correlation hardening at `.cartulary/test-results/20260908T214206Z-p2785044`.
The last finalizer pass, after resource-denial retirement review, is
`.cartulary/test-results/20260908T215051Z-p2917509`.
The final lifecycle denial/purge tests passed 2/2 execution units at
`.cartulary/test-results/20260908T214839Z-p2915408`.
`RESULTS_DIR` was unset: retained-run selection, duration/performance maintenance,
and retained run checks were explicitly SKIPPED. No qualifying exact-source
successful full warm `check` run was supplied or inferred from focused slices.
The finalizer's `unit-artifacts/finalize-summary.json` records those skips.

Final verification:

| Public command / exact catalog selection | Result | Run root below `.cartulary/test-results/` |
| --- | --- | --- |
| `make test-slice OWNER=module.imports ROWS=...` (all 17 catalog rows with `fixture_capability=none`) | PASS 14/14 execution units; 17/17 rows | `20260908T214221Z-p2788292` |
| `make service-backed-test-slice OWNER=module.imports` | PASS 14/14 execution units; all 20 service/browser rows | `20260908T214221Z-p2788266` |
| `make test-slice OWNER=web.application` | PASS 95/95 | `20260908T213730Z-p2637985` |
| `make test-slice OWNER=web.networkflow` | PASS 39/39 | `20260908T213730Z-p2638007` |
| `make test-slice OWNER=module.jobapi` | PASS 4/4 | `20260908T213730Z-p2638039` |
| `make test-slice OWNER=web.workbook ROWS=web.workbook.regression.workbookshell_assessments_suite_f2e4a841a2,web.workbook.regression.workbookshell_surfaces_suite_668e482b1e` | PASS 3/3 | `20260908T213751Z-p2679303` |
| `make frontend-typecheck` | PASS 2/2 | `20260908T215205Z-p2921136` |
| `make frontend-import-boundary-check` | PASS 2/2 | `20260908T215205Z-p2921174` |
| `make lint-biome` | PASS 2/2 | `20260908T215205Z-p2921177` |
| `make backend-module-boundary-check` | PASS 3/3 | `20260908T213750Z-p2677334` |
| `make generated-artifact-policy-check` | PASS 3/3 | `20260908T213749Z-p2675822` |
| `make json-shape-check` | PASS 3/3 | `20260908T213749Z-p2675989` |
| `make generate-drift` | PASS 4/4 | `20260908T213749Z-p2676017` |
| `make openapi-compatibility-check` | PASS 4/4 | `20260908T213749Z-p2676375` |
| `make browser-e2e-a11y` | PASS 12/12 execution units; 25/25 rows | `20260908T213730Z-p2638511` |
| `make browser-e2e-visual` (ordinary comparison) | PASS 12/12 execution units; 35/35 rows; no golden promotion | `20260908T213730Z-p2638288` |
| `make browser-e2e-measurement` | PASS 23/23 execution units; 7/7 rows | `20260908T214359Z-p2840852` |

The two final Imports slices account for every authored Imports row. Exact row
IDs, command selection, per-row terminal results, and artifact paths are retained
in each run's manifest, `rows/`, and `target-summaries/`. The browser review roots
in WI-04 remain the inspected presentation evidence; the final service slice
repeats all five browser workflows on the final correlated resource behavior.
No comparison image changed, so golden update/promotion and its two subsequent
ordinary runs were not applicable. The ordinary 35-row suite passed unchanged.

## Binary acceptance

| Criterion | Result | Evidence |
| --- | --- | --- |
| One import-specific owner; presentation closes without canceling or discarding acknowledged work | PASS | Controller lifecycle, App lifetime tests, keyboard browser closure/reopen |
| Secure immutable upload/write identities; exact retry retains bytes/body/ID; changed intent is new | PASS | Typed transport, lifecycle deferred replay, service upload/region/apply replay |
| Receipts survive failed observation/resource/preview reads | PASS | Typed transport, observation/lifecycle tests, same-job browser resume |
| Bounded cancellable observation; no timeout/abort claim of server failure; late/stale observations fenced | PASS | Injected-clock hung reads and deadlines; generation and terminal-state tests |
| Opaque paging, contradictory identities, empty discovery and preview failures remain distinct | PASS | Transport paging/preview tests, immutable-source refresh tests, CSV/XLSX services |
| Draft, acknowledged mapping/fingerprint, and persisted selection remain separate | PASS | Approval-then-selection-failure, skip/reselect unit and browser tests |
| Returned durable regions remain unapproved/unselected; overlap/readiness block apply | PASS | Region unit/browser/service tests, current-state selection/preflight tests |
| Mapping eligibility, target/binding and unknown-column facts come from current projections | PASS | Generated-target contract test, mapping tests, owner registry/service checks |
| Every write rechecks role, lifecycle, claim and lifetime; resource denial stays local until confirmed | PASS | 24-case command-stage matrix, App authority tests, resource denial/purge tests |
| Cancelability, duplicate cancel, cancel_requested, exact replay and completion races are distinct | PASS | Common-job owner slices, controller cancellation tests, authenticated public cancellation service test |
| Full/partial/canceled outcomes retain committed work and expose only durable result targets | PASS | Real partial XLSX, atomic/cancellation/recovery services, validated navigation tests |
| Network Flow preserves owner preview, fingerprint approval and analytical navigation | PASS | 39/39 frontend execution units; real claimed import browser; atomic owner service tests |
| Accessible names, keyboard/focus/announcements, non-color states, reduced motion, all densities, narrow/short layout, zoom and long content | PASS | Five reviewed Import PNGs, keyboard/density browser assertions, 25-row accessibility suite |
| No new route, target, dependency, browser storage, generic engine, compatibility facade or migration; generated outputs use public generators | PASS | Final source scope, generated-policy/drift/compatibility checks; explicit impact below |
| Final measurement, Markdown and whitespace checks and final handoff-byte audit | PASS | Measurement 7/7 rows; repeated Markdown/whitespace/scope checks |

## Compatibility, limitations, and rollback

Actual impact: internal frontend interfaces plus corrected authored public enum
projections and canceled-job/unit state serialization. No route, target,
dependency, storage mechanism, schema migration, or mapping fingerprint change.
Newly canceled unit processing stores the existing `failed` state; immutable
private outcome journals keep their cancellation reason. Existing retained units
remain readable without a data migration.
Recovery is volatile and scoped to known resources in this tab. Public routes
cannot discover historical imports or restore this state after a browser reload.
The UI shows public session/unit diagnostics when available; it does not expose
private per-unit outcome journals that the existing read routes do not return. Observation
timeout/abort cannot prove server failure or cancellation. Committed units are
retained after cancellation. Rollback must restore the frontend seam and its
tests/routing, cancellation serialization, and generated projections coherently;
no data rollback is planned.

Broader release/CI, full `check`, unrelated stateful browser suites, dependency
security scans, and destructive maintenance were not run: the changed owners,
common job seam, shared analytical consumer, real service workflows, and frontend
readiness have focused evidence above. This task makes no release or deployment
claim. No commit, push, reset, deployment, or unrelated digest/handoff edit was
performed.

Final scope review covers the import owner/models/transports, App/workbook
lifetime bindings and presentation, narrow Network Flow adaptation, three
Imports cancellation implementation files and their existing integration test,
authored OpenAPI/release and test routing inputs, generated projections, focused
frontend/browser tests, one deterministic XLSX fixture, and this handoff.
No parser, mapping kernel, backend scheduler, source-owner creation facade,
incident-bundle import, clipboard, or workbook-shell redesign is included.

Final pre-completion checks: `make lint-markdown` passed at
`.cartulary/test-results/20260908T215321Z-p2922634`; `git diff --check` and the
scope review passed. The review finds 36 tracked modifications and 17 new files,
only this handoff among Markdown changes, one 2,973-byte XLSX fixture, and no
migration, dependency, or committed golden changes. The same Markdown, whitespace, and 53-file scope checks are repeated against
the completed handoff bytes after marking WI-05 DONE.

WI-05 exit: PASS. All five workstreams and every binary acceptance criterion
are complete. Next action: none.
