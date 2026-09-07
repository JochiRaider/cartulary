# Incident import workflow refactor handoff

**Current iteration:** [Production hardening, II-06–II-10](#production-hardening-iteration).
The approved iteration is complete, including validation and handoff. The original baseline,
II-01–II-05 completion record and evidence below remain historical records of
the implemented workflow.

## Baseline and authority

Execution baseline: `main`, HEAD `dd83a122008e7f410fc1dabff4094ba58fa5e6e6`,
clean working tree. The inspection recommendation is not a reset target.
Only the repository-root AGENTS.md applies. The React/Vite/pnpm stack and
generated-artifact policy were revalidated. No dependencies or backend changes
are authorized. Account/Security/Users, session/bootstrap, directory, creation,
menu and workbook behavior remain regression baselines.

Allowed scope: import presentation, bounded controller/binding/client adapter,
necessary App lifetime/capability/startup integration, existing-operation client
selection and generated projections, focused tests/routing/selectors, reviewed
visual evidence and this handoff. No new API, durable history, browser storage,
archive engine, membership mode, Reference Pack/export redesign or digest edit.

| Behavior | Adopted owner | Verification owners |
| --- | --- | --- |
| Deployment context and import availability | Core 01 REQ-01-608; design §4.5 | web.application, web.design |
| Admission and canonical job resource | Core 01 REQ-01-248/249, §3.3.9.1 | module.jobapi, platform.jobs, web.application |
| Multipart, exact bytes and idempotency | Core 01 §17.1.1, REQ-01-485/486 | module.incidentbundles, package.protocol_ts, web.application |
| Publication and initial membership | Core 01 REQ-01-447–450/609 | module.incidentbundles |
| Explicit ordinary workbook launch | Core 03 REQ-03-030/290 | web.application, web.design |
| Current authorization and retirement | Core 04 REQ-04-023/028/126 | module.jobapi, web.application |
| Forms, progress, safe feedback and focus | design §§4.5, 10.8, 12.4–12.8, 14 | web.design, web.application |

The localized digest was read in its documented order. It is advisory;
domain.md owns vocabulary and nlspec-spec.md informs specification quality.
Neither is a new behavioral owner. No owner contradiction was found.

## Collection and lifecycle decisions

The list contains only imports learned in this application/session lifetime;
it does not discover jobs or claim historical completeness. The user selected
serial observation of all locally known nonterminal jobs while the panel is
active. Known jobs are keyed by job_id; selection controls details/actions.

| Event | Required handling |
| --- | --- |
| Panel/route change or hidden tab | Pause/invalidate reads and navigation; retain selected input, unresolved admission and jobs in memory. |
| Admission finishes while hidden | Accept only the current attempt/lifetime/capability; no navigation or focus move. |
| Reactivation | Refresh retained jobs before enabling stale actions; respect explicit Pause. |
| Logout, session replacement, capability/profile loss, disposal | Clear protected state and file references, fence continuations and abort where possible. |
| Reload/tab close | Memory recovery is lost; native unload guard for pending/uncertain admission. Server work may continue. |

Deadlines: admission 120 seconds; reads, cancellation and handoff 30 seconds.
Reads are serial, with one second between automatic reads. Client aborts and
deadlines are never server cancellation evidence. Uncertain admission replays
the same Blob and transaction ID; known jobs recover only through reads.

## Tracker

| Workstream | Status | Exit |
| --- | --- | --- |
| II-01 — Baseline, authority and characterization | DONE | Seven deterministic failures reproduced; lifecycle and routing recorded |
| II-02 — Typed admission and job lifecycle | DONE | Sixteen focused adapter/lifecycle cases pass |
| II-03 — Presentation and application integration | DONE | Inline workflow, guarded App lifetime and explicit startup handoff pass |
| II-04 — Browser, accessibility and visual evidence | DONE | Real import, keyboard recovery and two scoped visual passes |
| II-05 — Final validation and completed handoff | DONE | Final checks pass; acceptance, compatibility and limitations recorded |

## Work log

### II-01

Inspected IncidentImportPanel, App, landing types/styles, AppSessionController,
browserApi/httpTransport, shared error/focus infrastructure, App.landing tests,
account/directory handoffs, authored Incident Bundles/job OpenAPI, protocol
selection/bindings, browser fixtures, visual guide and verification catalog.
Public Make guidance resolved for web.application, web.design,
module.incidentbundles, module.jobapi and platform.jobs.

Source findings: modal-local loose job state; unchecked partial successes;
new transaction IDs per activation; overlapping/unfenced reads; stalled polling
after failure; no cancellation or explicit read recovery; result navigation
based on refs without terminal-success validation. These are source findings
until the characterization results below reproduce them.

Baseline: three selected App.landing rows passed fresh using
`make test-slice OWNER=web.application ROWS=... CARTULARY_HARNESS_CACHE_MODE=off`
at `.cartulary/test-results/20260907T165434Z-p3453423`. Those rows cover import
entry, deployment routing and non-admin exclusion. The import fixture is a
partial HTTP-200 terminal response and must be corrected to complete HTTP-202
queued/running admission plus a terminal read.

Next action: deterministic characterization before production edits.

Characterization added IncidentImportPanel.test.tsx and complete typed job
fixtures in testing/incidentImportTestSupport.ts, routed through
`web.application.regression.incident_import_workflow`. The first run stopped
on an authored catalog title-sort error, corrected before execution.
`make test-slice OWNER=web.application ROWS=web.application.regression.incident_import_workflow CARTULARY_HARNESS_CACHE_MODE=off`
then reproduced seven failures at
`.cartulary/test-results/20260907T170304Z-p3456482`: duplicate requests,
unlocked file replacement, missing recovery for malformed/lost admission,
late older-job overwrite, missing failed-observation retry and premature
navigation from a running job. The lost response also rejects outside owned
recovery. No production source had changed. These failures are related to this
seam and are expected characterization evidence, not passing baselines.

II-01 exit complete. Next action: II-02 typed admission and lifecycle.

### II-02

Added api/incidentImportClient.ts and incidentImportModel.ts with captured
immutable uploads, separate admission/job/observation/cancel/navigation states,
serial fair reads, bounded waits, explicit exact replay and lifetime guards.
Generated response validation is supplemented by job identity, deployment scope,
timestamp/progress/lifecycle consistency and a separate safe result target.
Unknown future reference kinds remain supported without weakening envelopes.

Projection correction: the operation generator previously ignored multipart
request bodies and emitted undefined request aliases. It now selects the
existing multipart schema when no JSON request exists. Generation adds the
import operation and multipart request types; the already-selected createImportSession and importReferencePack
upload aliases receive the same mechanical correction. No wire schema or
operation behavior changes. Generated HTTP types/validators/bindings and the
generated topology render index were produced by Make, never hand-edited.

`make generate` PASS at `20260907T171011Z-p3461897`.
`make frontend-typecheck` PASS at `20260907T171339Z-p3466149`.
Focused client/lifecycle rows PASS (16 tests) at
`20260907T171453Z-p3466911` using `make test-slice OWNER=web.application`
with ROWS set to the incident_import_client and incident_import_lifecycle rows,
cache disabled. The preceding run failed one test because its Response fixture
was reused after consumption; fresh Response objects fixed the fixture.

II-02 exit complete. Next action: II-03 presentation and application integration.

### II-03

Replaced the modal with an inline native upload form, session-local list and
selected job details. Added useIncidentImport.ts to own React activation,
visibility, unload protection, input clearing and recovery-focus integration.
App pauses at route/panel departure, retires on session/capability changes and
refreshes membership before explicit semantic incident navigation. No sheet_ref
is supplied. Unknown/incomplete results cannot navigate; failed handoff remains
separate from committed import success. Added stable import selectors and removed
four now-unused shared import-dialog style exports.

Repaired App.landing admission to complete 202 queued plus terminal reads.
All seven characterization cases now pass. The combined four focused rows
(24 tests) passed at `20260907T172115Z-p3473388`; frontend typecheck and import
boundaries passed at `20260907T172115Z-p3473466` and
`20260907T172115Z-p3473474`. package.ui slice passed at
`20260907T172422Z-p3476010`. A new App lifetime test passed at
`20260907T172530Z-p3477974`, proving panel recovery retention and clearing before
demotion publication with an ignored-abort late job response. Its first run
failed because the credential fixture used another actor; matching the credential
resource to the session fixed the fixture, with no production change.

`make format` passed; it also rewrote two pre-existing unrelated Go formatting
sites. Those formatter-only edits were inspected and restored to baseline.

II-03 exit complete. Next action: II-04 browser/accessibility/visual evidence.

### II-04

Added a test-only v3 TAR builder in e2e/support/incidents/import.ts. It reads the
authored JSON source inventory, constructs complete incident/actor/time-profile
source and checksum/manifest members, and leaves every import decision to the
server. No Markdown, private store access or backend test route is involved.
The new incident-import.spec.ts routes real active/closed uploads through actual
admission, jobs, session membership and ordinary workbook startup. A deterministic
gate drops the response after actual admission; exact replay returns the same
job and bytes. A controlled failed GET recovers via explicit observation retry.

The first real run, `20260907T173548Z-p3487121`, exposed an overstrict client
assumption: an exact replay returns the current job, potentially terminal, with
HTTP 202. REQ-01-248 says the **initial** job is queued/running. The replay does
not start another operation. The adapter/controller now distinguish initial
admission from explicit replay: initial responses remain queued/running-only;
a complete replay receipt can establish the known ID and stale snapshot, then
a fresh GET is required before result actions. This is a frontend correction
against the existing API, not an owner contradiction or wire change. Focused
coverage now includes terminal replay. The real browser and cancellation recovery
rows both passed at `20260907T173926Z-p3537276` (11/11 units).

The browser cancellation scenario covers rejection, response loss, exact cancel
ID replay, cancel_requested, canceled, monotonic progress and rejected terminal
regression. Accessible cancellation controls remain mounted while reads refresh;
aria-disabled gates interaction without removing keyboard focus.

Accessibility passed in `20260907T174157Z-p3590923`: keyboard file chooser,
required-field association, indeterminate admission, recovery-focus transfer,
observation retry, cancellation, visible focus, control names, contrast, 640×480,
200% zoom and text spacing. The same combined run's visual row stopped because
the existing generic visual text masker assigned a native file input's value.
It now skips file inputs, whose names are deterministic fixture data. No other
mask changes were made.

The corrected narrow visual run `20260907T174352Z-p3674403` completed all 14
functional capture states; comparison failed because these are new goldens.
Reconciliation also exposed an authored scenario ID with the wrong shape.
The authored new browser IDs now use the required hexadecimal shape and were
regenerated through Make. The initial generator attempt with a dynamic test title
also failed exact catalog resolution; the test now uses a static routed title.

Focused owner evidence passed with cache disabled:

| Make route | Run root suffix | Result |
| --- | --- | --- |
| service-backed-test-slice module.incidentbundles, shared upload/idempotency/imported-open row | 20260907T174306Z-p3640506 | PASS, 3/3 units |
| service-backed-test-slice module.jobapi, current deployment job authorization row | 20260907T174307Z-p3640756 | PASS, 3/3 units |
| service-backed-test-slice platform.jobs, lifecycle/progress/expiry rows | 20260907T174353Z-p3674641 | PASS, 3/3 units |
| test-slice package.protocol_ts, generated HTTP bindings row | 20260907T174439Z-p3736704 | PASS, 2/2 units |

Typecheck caught a Node Buffer/DOM BodyInit mismatch and two Node.remove fixture
calls at `20260907T174158Z-p3591207`; fixtures now use Uint8Array and parent-node
removal. Cross-field timestamp validation caught the expiry fixture's started_at
being later than its historical finished_at at `20260907T174438Z-p3736385`;
the fixture now supplies a coherent start time. Format caught nullable Promise
truthiness in the gated browser fixture; explicit null checks fixed it.

`make agent-finalize` passed at `20260907T174556Z-p3741280` before full ordinary
visual reconciliation. RESULTS_DIR was unset: retained-run selection, retained
run checks and performance-evidence maintenance were skipped. Generated
maintenance reported unchanged. Next action: complete ordinary reconciliation,
review/promote only the new import captures, then require two fresh visual passes.

Full ordinary visual reconciliation at `20260907T174614Z-p3744451` accounted for
136 captures: all 122 existing goldens active, exactly 14 new import goldens
missing, zero orphans, zero ambiguous mappings and all 26 registered fixtures
resolved. Thirty existing browser tests passed; only the new missing comparisons
failed. The approved inline workflow is the refresh trigger. The new captures
are implementation-support nonregistry fixtures under the authored
web.design.visual.incident_import_workflow row, not new D-VFIX claims.

Additional bounded review reproduced two lifetime gaps before fixing them:
`20260907T175143Z-p3795204` showed same-session confirmation could not restore
an import controller retired after 403. The binding now re-admits authority only
on a confirmed session resource observation, independently of panel activation;
upload is visibly unavailable while that confirmation is pending. The old
protected data remains cleared. `20260907T175613Z-p3896601` showed a cancellation
response arriving while hidden could mark a retained snapshot fresh. Hidden
cancellation results are now stale and require a GET on return before actions.

Focused import rows passed (25 cases) at `20260907T175403Z-p3800316` and typecheck
passed at `20260907T175405Z-p3800577`. Fresh combined real-browser/cancellation/
accessibility rows passed at `20260907T175441Z-p3803478` (13/13 units). The new
hidden cancellation assertion is being rerun with the final lifecycle fix.

Visual update `20260907T175440Z-p3801834` correctly refused promotion: native
indeterminate progress could not produce consecutive stable screenshots. Native
progress is now styled with the existing tokens, a static striped track for an
unknown total and native count-derived fill for known totals. It retains native
accessible semantics without animation or invented percentages. Review also
identified shared-worker display-name contamination from the preceding menu
fixture; the import presentation fixture now normalizes only that display label
in an otherwise real session response. Real import tests retain the actual session
response. No account or shell behavior was changed to absorb fixture contamination.
The import grid explicitly contains its own tracks and wraps long text.

The next update `20260907T180313Z-p3900950` passed every import capture and
produced all 14 candidates, but refused promotion because the unchanged
module.savedviews visual row timed out waiting for workbook readiness. Its source
and goldens were not edited. A fresh narrow rerun is checking that baseline.
All 14 candidate images were individually inspected: empty, required, admission
pending, short/zoom/text-spacing recovery, queued indeterminate, running
determinate, observation failure, cancel requested, canceled, failed, succeeded
and handoff unavailable. Text wraps, controls fit, progress is truthful and no
private diagnostics appear. Existing tracked goldens and manifest remain unchanged.

A stronger keyboard assertion reproduced focus loss after a second uncertain
admission at `20260907T180514Z-p3994362`. The binding now restores focus to the
new Retry admission control after another loss, to the file input after rejection,
and to job details after accepted recovery or a cancellation action disappears.
It moves focus only if the originating control disappeared and focus is on body;
leaving the panel clears that intent. Open preserves focus while busy, and the
single live-region source can announce a repeated failure event without polling
announcements. The expanded accessibility row is being rerun before promotion.

Expanded keyboard recovery/cancellation focus passed at
`20260907T180818Z-p4040971`; the unchanged saved-view row passed fresh at
`20260907T180858Z-p4084501`. Focused import tests passed again at
`20260907T181007Z-p4131713`, and typecheck at `20260907T181007Z-p4131809`.
Boundary verification caught the E2E helper importing a private app test fixture;
the E2E helper now constructs its complete generated-type job fixture locally,
without adding a shared job framework. Lint caught direct document.cookie test
assignment; the test now spies on the cookie getter. Import boundaries and Biome
passed at `20260907T181338Z-p4181872` and `20260907T181338Z-p4181878`.

Canonical visual update passed at `20260907T181020Z-p4133594`: 12/12 units,
136/136 active captures, zero missing/orphan/ambiguous mappings, all 26 registered
fixtures resolved. The 14 new import images were reviewed individually; the
handoff-failure image correctly retains visible focus on Open imported incident.
The first ordinary validation passed at `20260907T181354Z-p4182649` (31 browser
tests, 12/12 units). Review also found nine unrelated existing golden rewrites
from the update. All nine were inspected and compared with their baseline pixels,
then restored byte-for-byte; none is part of the final change. Make regenerates
the golden manifest, and two fresh ordinary passes are required against that
final scoped manifest. No comparison tolerance, existing fixture, or screenshot
mask was weakened.

The scoped manifest was regenerated at `20260907T181746Z-p37376` and differs
only by 14 new entries. The first fresh ordinary pass against it passed at
`20260907T181815Z-p40421` (12/12 units). Existing tracked PNG bytes match HEAD.
The second scoped ordinary pass passed at `20260907T181942Z-p88518` (12/12 units). Both scoped runs reconcile all 136 captures with no missing, orphan or ambiguous mappings.

Accepted visual identities all belong to
`web.design.visual.incident_import_workflow`, scenario
`scenario_49044213afe8`, Chromium,
and the existing viewport snapshot helper. Their explicit filenames are:

| New golden basename (all end in -linux.png) | Viewport and review |
| --- | --- |
| incident-import-empty | 1280×720; inline form, session-local empty state, focus |
| incident-import-required | 1280×720; associated required-field feedback |
| incident-import-admission-pending | 1280×720; locked upload, static indeterminate progress |
| incident-import-recovery-short | 640×480; wrapped feedback and reachable Retry |
| incident-import-recovery-zoom | 1280×720 at 200%; native focus scroll retains recovery |
| incident-import-recovery-text-spacing | 768×640 with declared spacing; no horizontal clipping |
| incident-import-queued-indeterminate | 1280×900; unknown total, paused observation, job selection |
| incident-import-running-determinate | 1280×900; 2/8 native progress, no invented phase |
| incident-import-observation-unavailable | 1280×900; retained status, Retry, disabled cancellation |
| incident-import-cancel-requested | 1280×900; request distinguished from terminal outcome |
| incident-import-canceled | 1280×900; server cancellation, no opening |
| incident-import-failed | 1280×900; terminal failure, no opening |
| incident-import-succeeded | 1280×900; one explicit incident-entry action |
| incident-import-handoff-unavailable | 1280×900; committed success retained, Open focus/retry |

Capture scope remains the viewport. No crop or new dynamic mask was added.
The existing helper's dynamic text normalization and pinned renderer remain;
only native file values are exempted from illegal assignment. The recovery
fixtures declare their zoom/text spacing and scroll naturally through keyboard
focus. All images and changed image comparisons were reviewed, including the
nine existing candidates that were excluded from the final scope.

II-04 exit complete. Next action: II-05 final validation and completed handoff.

### II-05

`make agent-finalize` passed at `20260907T182354Z-p136950`, before broader final
verification. RESULTS_DIR remained unset because no qualifying exact-source full
warm check run was supplied. Retained-run selection/closure and performance
maintenance were skipped; this is not a retained full-check claim. Final
application, frontend boundary/type/lint, protocol/selector and projection checks
passed through public Make routes. Run roots below are under
`.cartulary/test-results/`; counts are harness units, including prerequisites.
All test slices used `CARTULARY_HARNESS_CACHE_MODE=off`.

| Command | Result | Run ID |
| --- | --- | --- |
| `make test-slice OWNER=web.application` | PASS 72/72 | 20260907T182539Z-p140812 |
| `make frontend-typecheck` | PASS 2/2 | 20260907T182539Z-p140990 |
| `make frontend-import-boundary-check` | PASS 2/2 | 20260907T182539Z-p141011 |
| `make lint-biome` | PASS 2/2 | 20260907T182539Z-p141044 |
| `make test-slice OWNER=package.protocol_ts ROWS=package.protocol_ts.frontend_unit.generated_http_operation_bindings,package.protocol_ts.boundary_support.browser_bundle_excludes_protected_audit_and_revi_13733d4a6b` | PASS 3/3 | 20260907T182539Z-p140855 |
| `make test-slice OWNER=package.ui` | PASS 10/10 | 20260907T182539Z-p140867 |
| `make generate-drift` | PASS 4/4 | 20260907T182539Z-p140709 |
| `make generated-artifact-policy-check` | PASS 3/3 | 20260907T182539Z-p140719 |
| `make json-shape-check` | PASS 3/3 | 20260907T182539Z-p140729 |

The application owner slice includes the completed account/security/users,
session/bootstrap, directory, incident creation, menu and startup baselines.
The service-backed import/job/platform, real-browser and accessibility evidence
is recorded in II-04 above. Both final visual passes used the final scoped
manifest. No product source changed after those passing browser runs.

Final scope review found 43 paths: 19 tracked modifications and 24 new files,
including exactly 14 new import PNGs. Existing tracked PNGs match HEAD. Branch
and HEAD remain the recorded baseline. No Go/backend, dependency/lockfile,
archive/storage/API owner, digest or unrelated documentation changed. New
runtime/test/generator code has no Markdown dependency. Final documentation
lint and whitespace checks close this workstream below.

`make lint-markdown` passed at `20260907T183244Z-p161434`.
`git diff --check` passed. The final audit used `git status --short`,
`git diff --name-only`, `git ls-files --others --exclude-standard`, branch/HEAD
reads, and inspection of the resulting authored/generated diff. It confirmed
the 43-path boundary above, with no unrelated changes retained. Documentation
lint, whitespace and scope checks are repeated after this closure edit.

II-05 exit complete. All five workstreams are DONE; no owner contradiction or
unresolved implementation failure remains. Earlier failed runs are preserved
above with their corrections or successful baseline reruns. There is no next
implementation action in this authorized seam. No commit, push or deployment
was performed.

## Final architecture and owner interpretation

The [Core 01 owner](../spec/01_architecture_storage_and_view_contracts.md)
provides deployment administration (REQ-01-608), initial admission (REQ-01-248),
canonical job fields, lifecycle, progress, retention, additive resource kinds and
cancel semantics (§3.3.9.1 and REQ-01-249), upload envelope (§17.1.1), and Incident
Portability (§17.5, especially REQ-01-485/486). REQ-01-447–450/609 own publication
and initial membership. Those server responsibilities remain unchanged.
[Core 03](../spec/03_workbook_interaction_collaboration_and_workflows.md)
REQ-03-030/290 supply ordinary explicit startup without sheet_ref.
[Core 04](../spec/04_security_deployment_and_conformance.md) REQ-04-023/028/126
supply current job authorization, session retirement and removal of protected
state. [Design](../design.md) §4.5 supplies the local import workflow/list/result
interaction, with applicable form, feedback, focus, progress and responsive rules.

The adapter owns generated operations, immutable multipart capture and resource
validation. The model owns admission, keyed known jobs, selection, observation,
cancellation, navigation and lifetime transitions. The React binding owns
activation/visibility, session-confirmation integration, unload protection and
focus restoration. Presentation receives state/actions and contains no fetch or
polling effect. App provides the existing capability/lifetime and semantic
workbook-startup boundary. There is no shared account controller, workbook queue
or general job framework in this workflow.

First-time admission requires complete HTTP-202 queued/running data. An uncertain
attempt replays exactly its original File/Blob and client_txn_id. The server's
current job receipt on replay can already be terminal; it still requires a fresh
GET before navigation. Once known, a job is never recovered by another import.
GET/cancel snapshots must bind the expected job identity and deployment scope,
retain submitter/submission identity, respect timestamp and progress monotonicity,
and never regress a terminal lifecycle. Unsupported result navigation is separate
from a valid succeeded status. Unknown additive reference kinds are ignored;
only a correct import result with one supported incident reference can open.
API reference routes are never browser destinations.

Cancellation uses a separate per-job transaction ID. Uncertainty preserves that
attempt for explicit replay only after a fresh eligible read. Rejection, lost
responses and completion races lead to authoritative observation. Hidden results
remain stale until reactivation reads them. Only server state establishes the
outcome. Client deadlines and aborts never establish cancellation or rollback.

## Final changed-path boundary

- apps/web/src/app: IncidentImportPanel.tsx, incidentImportModel.ts,
  useIncidentImport.ts, api/incidentImportClient.ts, focused tests, App.tsx,
  App.landing.test.tsx and removal of unused import-dialog styles in
  landingAdminStyles.ts.
- apps/web/src/testing/incidentImportTestSupport.ts: complete unit job fixtures.
- apps/web/e2e: incident-import.spec.ts, support/incidents/import.ts, appended
  accessibility/visual rows, and exactly the 14 new incident-import PNGs above.
- packages/ui-contracts/src: authored import selectors, exports and selector tests.
- contracts/protocol-ts/http-operations.v2.json and
  tools/protocol-ts/generate-protocol-types.mjs: existing import operation
  selection and multipart request projection correction.
- Generated protocol core-http-types.ts, core-http-validators.ts and
  http-operation-bindings.ts; generated browser batch/render indexes and the
  visual golden manifest. All generated through Make.
- tools/test_families/web.application.json and web.design.json: authored routing.
- This handoff. No other documentation or digest file changed.

## Advisory acceptance dispositions

PASS below means bounded implementation-support evidence, not a Core 05 claim.
Not applicable means the named subsystem behavior was preserved outside this seam.

| ID | Disposition | Evidence or boundary |
| --- | --- | --- |
| A001 | PASS | Adopted owner map above; digest never promoted to authority |
| A002 | PASS | One upload/admission/job/result seam; structure separated from behavioral fixes |
| A003 | PASS | Clean main/HEAD baseline, owners, stack, generated roots and current routing inspected |
| A004 | PASS | Existing tokens and native controls; no new theme/token registry |
| A005 | PASS | Current graphite theme only; no generated cybersecurity palette |
| A006 | Not applicable | Shared density and workbook geometry unchanged |
| A007 | Not applicable | Workbook row creation unchanged; no import modes or initial-role selection added |
| A008 | PASS in scope | Short/narrow/zoom/text-spacing import controls and panel containment tested |
| A009 | PASS in scope | Inline content scrolls through existing shell; ordinary workbook/navigation baselines pass |
| A010 | Not applicable | Inspector routes and feature ownership unchanged |
| A011 | PASS in scope | Keyed job selection, retained snapshots, gated late results and keyboard recovery focus |
| A012 | PASS | Existing Web Crypto transaction IDs; exact uncertain admission/cancel replay |
| A013 | Not applicable | Workbook blocked-edit queue unchanged; import uses its own exact replay rules |
| A014 | Not applicable | Workbook cell editing/paste unchanged |
| A015 | PASS in scope | Import field/operation feedback local; no toast-only unresolved state |
| A016 | PASS | Empty, admission, all six statuses, stale/failed/unavailable observation, cancellation and navigation distinct |
| A017 | PASS | Last validated snapshots retained on read failure; protected data removed on retirement |
| A018 | Not applicable | Evidence lifecycle and preview overlays unchanged |
| A019 | PASS | Real keyboard chooser/repeated recovery, focus, names, live source, contrast and constrained layout |
| A020 | PASS in scope | Focused state tests and 14 reviewed visual states; no new shared component matrix |
| A021 | Not applicable | Workbook virtualization unchanged; session-local job collection is not a worksheet |
| A022 | PASS | 14 new nonregistry captures; 26 registered fixtures unchanged; two scoped ordinary passes |
| A023 | PASS | Authored semantic import selectors and stable job_id selectors; package.ui slice |
| A024 | PASS | Test fixtures read authored JSON only; no runtime/test/generator Markdown dependency |
| A025 | PASS | Authored selection/routing inputs; Make generation, drift and artifact policy checks pass |
| A026 | PASS | Existing import/getJob/cancel APIs; no archive, auth, storage, schema or route invention |
| A027 | PASS | Baseline, behavior, paths, commands, failures, reviewed evidence, final checks and limits recorded |

Advisory classification: ADOPT the owner-compatible guidance for inline forms,
local feedback, focus continuity and truthful progress; ADAPT the job-list guidance
to the explicitly session-local known-ID collection; REJECT alternative palettes,
new dialog infrastructure, durable history/discovery, new API resources and import
modes outside adopted owners. No upstream skill or digest is an executable input.

## Compatibility, limitations and rollback

This is frontend workflow remediation using existing import/job APIs. There is
no archive, database or stored-data migration. Other administrative, account,
directory and workbook surfaces retain their owners and interfaces.

The collection is in-memory and deliberately incomplete. It survives ordinary
panel/route changes and hidden tabs; it is cleared on reload/tab closure,
sign-out, session replacement, capability/profile loss and disposal. Reload or
closure loses unresolved-file recovery even though server work can continue.
The native unload guard covers pending/uncertain admission. There is no browser
storage, cross-account history, general deployment job discovery or manual job-ID
entry. Failed handoff retries only opening, never import. Current job retention
and availability remain server-owned.

Browser evidence uses the pinned Chromium stack. Visual/accessibility evidence is
implementation support, not conformance or benchmark publication. Full repository
CI/release gates and unrelated owner suites are not claimed by these bounded runs.
Retained full-warm-run maintenance was intentionally skipped without RESULTS_DIR.

Rollback is cohesive: revert import panel/model/binding/adapter and App integration,
operation selection and its generated projections, scoped selectors/routing/tests,
14 import goldens and their manifest entries. Restore the prior workflow together;
do not retain an adapter/controller half against the old unchecked presentation.
The multipart projection correction can be retained independently only after
reviewing its two other already-selected upload aliases. No stored data needs
rollback. No commit, push or deployment is part of this work.

## Production hardening iteration

### Execution authority and baseline

The user approved implementation of the expanded II-06–II-10 remediation plan,
including specification cleanup and shared backend HTTP error safety. That
authorization supersedes the historical frontend-only boundary and the previous
document-update-only restrictions. II-01–II-05 above remain historical evidence;
their completion statements do not describe this iteration.

Execution baseline: main at b1624e9f30956453b877db62f9393a41fd64c566.
The handoff alone had staged user changes, retained intact as the starting point.
No reset, dependency change, new endpoint, stored-data migration, commit, push,
or deployment is included. Domain vocabulary remains owned by domain.md.

Fresh planning evidence: the four incident_import_client, incident_import_lifecycle,
incident_import_workflow and incident_import_application_lifetime rows passed
through make test-slice OWNER=web.application with cache off, 5/5 harness units,
at .cartulary/test-results/20260907T185737Z-p173041. This is a passing baseline,
not evidence for the new requirements.

### Remediation and acceptance map

| Gap | Remediation and areas | Rationale and lasting benefit | Compatibility and unresolved risk | Completion criteria |
| --- | --- | --- | --- | --- |
| G1 — Owner ambiguity | Core 01 defines replay receipts and safe internal_error; Core 03 owns action revalidation; Core 04 clarifies concealed job lookup and mandatory expiry; Design owns presentation. Update projections only from authored machine inputs. | Define behavior once so future implementations agree. | Preserve wire shapes and job concealment; restrict diagnostic text. Unresolved ambiguity risks inconsistent expiry, rejected replay and information disclosure. | Human owner review plus independent replay, authorization and exact-expiry tests. No executable Markdown dependency. |
| G2 — Raw client outcomes | Import-specific discriminated admission, observation and cancellation outcomes; preserve received status through malformed bodies, including multipart. Implementation/tests. | Normalize once; every operation settles without leaking transport details into transitions. | Preserve exact Blob, filename, transaction IDs and CSRF. Otherwise malformed failures can strand UI or duplicate operations. | Null, scalar, array, invalid JSON, HTML, empty body, throws, lost response and status matrix settle correctly. |
| G3 — Internal exception disclosure | Shared HTTP writer fixes internal_error to 500, fixed message, false retry hint, empty details, no conflict and server request ID. Remove raw cause copying in common and Job API helpers. Specification/backend/tests. | One enforced boundary protects every caller. | No data migration; exception-text consumers must use code/correlation. Otherwise private SQL, paths or infrastructure details can escape. | Hostile markers absent from complete shared and job responses; typed non-internal errors unchanged. |
| G4 — Coupled state and effects | Pure transitions/selectors, bounded executor, independent snapshot availability and read activity. Implementation/tests. | Cohesive state makes future growth testable without asynchronous branch sprawl. | Replace obsolete helpers and duplicate validation together. Otherwise refresh removes focused controls or cancels valid handoff. | State invariants, scheduling cleanup and stable DOM/focus through refresh pass. |
| G5 — Stale explicit actions | New serial GET after each Cancel/Open intent; priority next slot; bounded queue time, duplicate suppression and lifetime guards. Spec/implementation/tests. | Separate intent, observation and committed effects. | One GET per action; preserve exact cancel replay. Otherwise stale authorization and cancellation races can mutate or navigate incorrectly. | Queue, deadlines, terminal races, departure, duplicate intents and delayed responses pass. Failed preflight has no side effect. |
| G6 — Browser-clock expiry | Remove local expiry transitions and availability clock checks; server GET decides; monotonic local deadlines. Implementation/tests/docs. | Fewer timers and correct behavior under clock skew. | Retained result may remain visible until authoritative read; expiry never means output deletion. Otherwise valid results become locally inaccessible. | Positive/negative skew has no authority; server 404 does; durable outputs survive expiry. |
| G7 — Session object recovery | Explicit typed, lifetime-bound access confirmation, Retry access, current discovery when uncertain. Rename operation-observation interface without alias. App/session/binding/tests/docs. | Session publication and object allocation no longer grant authority implicitly. | Retirement clears files/jobs permanently; restoration starts empty. Otherwise recovery can stall or resurrect protected state. | Confirmation failure/retry, same identity, demotion, replacement, profile loss, ambiguous 404 and ignored-abort responses pass. |
| G8 — Conflicting results | Central immutable identity, progress, lifecycle and committed-target checks; allow skipped observations and ignore unknown additive refs. Spec/implementation/tests. | Strong target identity with forward-compatible summaries. | Unsupported/ambiguous results render but cannot open; conflicting snapshots do not overwrite validated state. Otherwise wrong incidents can open. | Target conflict, duplicate refs, additive refs, unknown result, lifecycle regressions and unchanged refresh pass. |
| G9 — Missing evidence | Deterministic regression, real browser, keyboard/a11y, reviewed scoped visuals and completed tracker. Tests/routing/docs. | Preserve a reliable regression boundary for later phases. | Replace assertions encoding removed defects; no unrelated golden refresh. Otherwise refactor completion can conceal unchanged bugs. | Every gap has owner mapping, routed results and final disposition; two ordinary visual passes after updates. |

### State, timing and lifecycle decisions

The import adapter owns generated wire operations and validation. Pure state owns
admission, retained snapshots, availability, read activity, cancellation attempts,
action intent, access and announcements. The executor owns requests, serial reads,
timers and cleanup. React owns visibility, native input clearing and focus. App
owns current session/capability confirmation and semantic workbook startup.
There is no generic job framework or extension registry.

Admission 4xx except 408 rejects even with malformed bodies; 401 retires the
session and 403 clears import authority before confirmation. Transport loss,
408, 5xx, unexpected success status or malformed success remains uncertain.
Read/cancel 404 means unavailable, not proof of global capability loss.
Unknown failures use fixed safe feedback; no raw message drives decisions.

Admission waits 120 seconds. Reads, action preflight (including queue time),
cancel dispatch and workbook handoff each have separate 30-second bounds.
A new action GET starts after its intent and takes the next available serial
slot. Duplicate intents are suppressed. Selection/panel/visibility/lifetime
changes invalidate undispatched actions. Dispatched cancel ownership survives
ordinary panel changes; uncertainty reuses its exact ID only after eligible
preflight. Pause stops automatic polling, not explicit recovery or actions.
Automatic reads remain fair across locally known nonterminal jobs, with one
second between automatic reads.

Routine refresh preserves validated snapshots and mounted controls. Server reads,
not browser time, establish availability. Open requires supported success,
current membership and ordinary startup without sheet_ref. Failed handoff retries
opening only. Job unavailable feedback never claims committed output deletion.

After 401/403 or confirmed capability loss, protected state is cleared and late
continuations fenced. Failed confirmation exposes Retry access. Confirmed access
after retirement restores an empty workflow. Ambiguous job 404 retains the
unavailable snapshot while confirmation disables new protected actions; confirmed
loss clears state before publication. Current discovery confirms uncertain profile
authority. Reload loses memory recovery, and server work may continue.

### Sequential tracker

Only one executable child row may be IN_PROGRESS. Parent II-06–II-10 phases are
rollups, not independent execution. Close evidence and update this tracker and
work log after each child, before beginning the next. Unexpected owner
contradictions block dependent work until reconciled.

| Parent phase | Child completion rollup |
| --- | --- |
| II-06 | Complete: II-06a and II-06b |
| II-07 | Complete: II-07a, II-07b and II-07c |
| II-08 | Complete: II-08a and II-08b |
| II-09 | Complete: II-09a and II-09b |
| II-10 | Complete: validation, acceptance audit and handoff closure |

| Workstream | Status | Dependency and exit |
| --- | --- | --- |
| II-06a — Specification and controlling plan | DONE | Owner wording and acceptance reviewed; Markdown and whitespace pass. |
| II-06b — Characterization | DONE | New hardening, access and shared/job error rows reproduce product failures before production edits. |
| II-07a — Shared server error safety | DONE | Shared and job-local sentinel tests pass; typed public errors and correlation preserved. |
| II-07b — Typed import boundary | DONE | Typed client/lifecycle and expanded malformed-response matrix pass; callers typecheck. |
| II-07c — Pure state and snapshot ownership | DONE | Pure state/client/lifecycle and focused-control tests pass; typecheck passes. |
| II-08a — Authoritative actions and expiry | DONE | Fresh serial preflight, queue deadlines, cancellation ownership and clock independence pass. |
| II-08b — Access and application integration | DONE | Typed recovery, retained ambiguous 404, retirement, session collaborators and focus pass. |
| II-09a — Browser and accessibility evidence | DONE | Real launch, action/access faults, profile loss and keyboard/layout evidence pass. |
| II-09b — Visual evidence | DONE | Three new/one changed scoped goldens reviewed; two full ordinary visual passes. |
| II-10 — Validation and handoff completion | DONE | Required checks pass; all gaps accepted; scope, limitations, compatibility and rollback recorded. |

Primary phase risks: II-06 can encode defects as requirements; II-07 can widen
shared HTTP compatibility or retain duplicate paths; II-08 can admit late actions
or restore authority accidentally; II-09 can hide defects in fixtures or unrelated
goldens; II-10 can overstate incomplete evidence. The exit gates above address
those risks and remain mandatory.

### Verification and handoff contract

Resolve Make routing for web.application, module.auth, module.jobapi,
module.incidentbundles, platform.jobs, app.server, platform.openapi,
package.protocol_ts, package.ui and web.design. Add authored routing for new
cases; generate topology/projections only through Make. Narrow fresh slices
precede broader checks. Preserve real active/closed import, exact upload replay,
six statuses and explicit launch; use deterministic gates for races.

Run make agent-finalize before broader final verification. Without qualifying
retained full-run evidence leave RESULTS_DIR unset and report maintenance skips.
Require applicable type/boundary/lint, generated drift/policy/JSON checks,
service-backed import/job evidence, browser/a11y/visual checks and make check.
Follow the visual golden maintenance guide: ordinary reconciliation, justified
update, individual image review and two fresh ordinary passes. Finish with
make lint-markdown and staged/unstaged whitespace and scope audits.

II-10 records changed files and behavior, commands/run roots, all failures and
corrections, skips with reasons, every G1–G9 disposition, memory/reload/retention
limits and no stored-data migration. Roll back frontend/state/binding/projections
cohesively. Shared error safety is independently revertible; workflow rollback
must not automatically restore disclosure. No production certification is implied.

### Hardening work log

#### II-06a — Complete

Revalidated main/HEAD and the staged handoff-only starting state. Updated Core 01
common internal errors and replay receipt semantics, Core 03 explicit actions and
recovery, Core 04 concealed jobs and exact expiry, and Design focus/presentation.
Existing typed error registry already declares internal_error=500; wire shape
does not change. Owner review found no remaining contradiction for this seam.
make lint-markdown PASS at 20260907T190845Z-p176935; git diff --check PASS.
No product code, test, generator or dependency changed in this workstream.
II-06a exit recorded before starting characterization.

#### II-06b — Complete

Added independent import hardening cases for malformed admission status, retained
actions during refresh, new action reads, cancellation completion, clock skew and
conflicting/additive targets. Strengthened the existing access case to require
failed explicit confirmation and Retry access instead of new object identity.
Added shared HTTP and job-local private-error sentinel tests. Authored routing
belongs to web.application, app.server and module.jobapi; make task-guide resolved
their narrow commands. make generate PASS at 20260907T191146Z-p179246.

Fresh cache-off make test-slice failures are expected characterization evidence:
web.application incident_import_hardening at 20260907T191213Z-p182201;
app.server internal_error_boundary at 20260907T191214Z-p182447;
module.jobapi internal_error_safety at 20260907T191308Z-p183387;
web.application incident_import_workflow at 20260907T191348Z-p184194.
All report product/test_assertion_failure rather than missing selectors. The
earlier passing four-row baseline remains separate. G1 is owner clarification;
G2–G8 now have targeted failing rows and G9 maps the missing evidence. No production
source changed before these reproductions. Service-backed replay/auth/expiry
fixtures remain existing owner evidence to rerun in final validation.
II-06b exit recorded before shared server safety implementation.

#### II-07a — Complete

The common HTTP writer now enforces the fixed internal_error envelope, clearing
caller diagnostics/details/conflicts and false retry hints while preserving
request correlation. Shared and job-local constructors no longer copy raw causes.
Typed non-internal responses retain their prior envelope. No logging was added.
Fresh app.server internal_error_boundary PASS at 20260907T191559Z-p187752
(including correlation and non-null empty details); module.jobapi
internal_error_safety PASS at 20260907T191535Z-p185628. Both use cache-off slices.
make format PASS at 20260907T191618Z-p188262; two unrelated baseline Go formatter
sites were inspected and restored before client conversion. G3's reproductions now pass; broader
shared-route compatibility remains a required II-10 gate.
II-07a exit recorded before typed client conversion.

#### II-07b — Complete

Admission, GET and cancellation now return distinct typed outcomes. The adapter
owns safe unknown-body inspection and received-status classification. Multipart
adds the same onResponse hook already available for JSON operations, before body
parsing; existing request framing/bytes/CSRF remain unchanged. The executor and
fixtures consume typed outcomes without raw payload branches. Existing client and
lifecycle plus admission-error rows PASS, 4/4 units, at 20260907T192145Z-p195391;
frontend-typecheck PASS at 20260907T192145Z-p195456. Expanded read/cancel malformed,
throw/loss/invalid-success matrix PASS at 20260907T192220Z-p196633.
An initial fixture conversion omitted parentheses around an async object return;
typecheck 20260907T192025Z-p193932 and lifecycle routing 20260907T192026Z-p194309
failed before correction. The reruns above pass. No production compatibility shim
was retained. G2 is complete; action/access characterization stays intentionally
red until its assigned slices. II-07b exit recorded before pure state extraction.

#### II-07c — Complete

Added incidentImportState.ts with pure events/transitions/selectors. The controller
now executes effects and dispatches events rather than patching workflow state.
Read activity is a separate flag; refresh preserves validated state. The result
control remains mounted through read failure/recovery. Central job advancement
now rejects changed committed target/code and established lifecycle timestamps,
while accepting unknown additive refs and skipped intermediate observations.
State/client/lifecycle rows PASS, 4/4, at 20260907T192805Z-p199548; focused-control
recovery PASS at 20260907T192915Z-p201053; typecheck PASS at
20260907T192828Z-p200421. Initial typecheck 20260907T192658Z-p198171 caught a
remaining presentation comparison against removed reading state; it was migrated
to the new activity flag. The mixed hardening row still fails pending action
preflight, as expected (20260907T192700Z-p198561). Selectors are clock-independent;
the remaining executor expiry timer is removed in II-08a. G4/G8 evidence passes.
II-07c exit recorded before action integration.

#### II-08a — Complete

Explicit Cancel/Open now create bounded intents with a fresh prioritized serial
GET. Queue time counts toward preflight's monotonic deadline. Duplicate intents,
departure and late responses cannot dispatch. Dispatched cancellation retains its
exact retry attempt across panel changes. Separate dispatch/handoff bounds remain.
Removed browser expiry decisions and timer; authoritative 404 controls availability.
Refresh no longer invalidates an unchanged workbook handoff. Paused automatic
observation still permits explicit actions.

Hardening/state/lifecycle/focus rows PASS, 5/5, at
20260907T194040Z-p208572; typecheck PASS at 20260907T194142Z-p210255.
The first mixed run 20260907T193221Z-p202614 exposed old local-expiry and cached
action fixture assumptions; updated lifecycle passed at 20260907T193429Z-p204270.
Typechecks at 20260907T193220Z-p202231, 20260907T193428Z-p203881,
20260907T193932Z-p207859 and 20260907T194042Z-p208839 caught union narrowing,
one leftover removed expiry helper call and a fixture's cancelled/canceled typo.
All corrected; direct compiler diagnostics were used only to expose errors omitted
by the graph summary, and the public Make rerun supplies passing evidence.
G5/G6 exit recorded before access integration.

#### II-08b — Complete

Replaced observedSession binding identity with a typed confirmation port, bound
authority and recovery generation. Retirement clears protected data; fresh typed
confirmation restores only an empty workflow. Failed/cancelled confirmation
settles into Retry access. Job 404 keeps its unavailable snapshot while protected
actions are suspended; confirmed loss clears it. Retry focus returns to the new
retry control or file input. Cancellation focus ownership spans preflight.

Renamed all account-specific session-observation callers to observeOperationSession,
without an alias. Current profile confirmation uses observeOperationExtensions
over the existing session observation executor and publishes no artificial loading
state. App confirms session/admin identity and current profile discovery before
authorization. Internal interfaces changed; wire/API and persistence did not.

Import/session/account collaborator rows PASS, 7/7, at
20260907T194552Z-p213555; typecheck PASS at 20260907T194553Z-p213860.
Expanded hardening/state/workflow/session rows PASS, 5/5, at
20260907T194844Z-p286496. module.auth owner slice PASS, 37/37, at
20260907T194618Z-p215267. Boundary check PASS at 20260907T194659Z-p267414;
Biome PASS at 20260907T194843Z-p286048; format PASS at
20260907T194827Z-p281719 with unrelated baseline Go formatting restored.
Typecheck 20260907T194424Z-p211765 caught a widened fixture literal, corrected.
Biome 20260907T194659Z-p267425 and 20260907T194804Z-p280905 caught new test
non-null assertions and subsequent formatting; explicit fixture checks and Make
format resolved them. G7 exit recorded before browser evidence work.

#### II-09a — Complete

Real active/closed archive tests now prove a new job GET precedes membership
recovery and ordinary workbook startup. Cancellation replay assertions wait for
dispatch after preflight. A separately routed browser scenario gates action
reads, suppresses duplicates, proves failed preflight cannot cancel, recovers
ambiguous 404 through repeated Retry access, and confirms current profile loss
removes the protected panel. Accessibility adds focused checking, failed preflight,
definitive rejection and disappearing cancellation controls to existing reduced
motion, keyboard, zoom, text spacing and constrained-layout coverage.

Fresh combined service-backed web.design rows PASS, 13/13 units, at
20260907T195207Z-p302616. The retained browser reports cover three functional
scenarios and the expanded accessibility scenario. Typecheck PASS at
20260907T195045Z-p291860; generate PASS at 20260907T195149Z-p295277.
Generator attempts 20260907T195017Z-p288043 and 20260907T195044Z-p291209
caught a missing scenario identity and unsorted identity array. The first browser
attempt likewise rejected the catalog before product execution. The new scenario
now has its own row/identity, preserving the historical scenario mapping.
Format PASS at 20260907T195151Z-p295933; unrelated Go formatting restored.
II-09a exit recorded before visual reconciliation.

#### II-09b — Complete

Ordinary reconciliation at 20260907T195335Z-p352168 accounts for 139 intents:
136 existing active goldens, three new missing captures, no orphans or ambiguous
mappings and all 26 registered fixtures resolved. All unrelated visual tests pass.
The sole existing changed comparison is incident-import-observation-unavailable:
Cancel remains enabled because activation performs fresh preflight. Its actual and
diff were inspected; the only changed pixels are that control's enabled text.

New planned captures are incident-import-action-checking,
incident-import-action-unconfirmed and incident-import-access-retry, all under
web.design.visual.incident_import_workflow / scenario_49044213afe8, Chromium,
1280×900 viewport. This is an approved workflow/presentation change; no masks,
zoom, scroll normalization, screenshot scope, renderer or tolerance changed.
The new ordinary captures completed their functional assertions before reporting
missing goldens. The canonical visual update passed at 20260907T195633Z-p405881, 12/12 units,
with all 139 captures active and no reconciliation defects.

Review also identified hidden-document focus restoration after a dispatched
cancellation. The binding now suppresses it; focused/hidden-focus row PASS at
20260907T195555Z-p404805. Typecheck and Biome PASS at
20260907T195500Z-p403770 and 20260907T195500Z-p403780. Make format PASS at
20260907T195657Z-p449561 and generate PASS at 20260907T195745Z-p458987.

All 14 changed/promoted images were individually inspected. The three new
1280×900 states retain visible focus, readable safe feedback and reachable
controls; the existing observation-failure image reflects preflight eligibility.
Ten incidental rewrites were inspected and restored byte-for-byte: four unchanged
import states (empty, canceled, queued, running) and six unrelated account/auth,
collaboration, creation and Timeline states. The final manifest regenerated at
20260907T200056Z-p463272 changes only three new entries and the justified existing
observation-failure hash. Two ordinary scoped passes satisfy the exit gate: 20260907T200117Z-p466414
and 20260907T200353Z-p515125, both 12/12 units. Each reconciles all 139 active
captures with zero missing/orphan/ambiguous mappings and all 26 registry fixtures.
Unrelated PNG bytes remain identical to HEAD. II-09b exit recorded before final
validation and handoff work.

#### II-10 — Complete

Final evidence audit characterized a remaining availability coupling at
20260907T200653Z-p562997: a failed read after an authoritative 404 made the
cached result actionable again. KnownImport now stores server-established
availability independently from the last observation outcome and read activity.
Only validated authoritative outcomes restore availability. Last snapshots,
failed-read feedback and unavailable feedback coexist without resurrecting actions.
The new pure regression and five import rows pass at 20260907T200850Z-p568640
(6/6). Positive and negative browser-clock offsets are covered.

Shared error tests now inspect individual secret, SQL and path markers across
the complete response. Actual Job GET and cancel handlers inject a failing
database read and assert the canonical envelope, separately from the private job
helper test. module.jobapi internal_error_safety PASS at
20260907T200851Z-p568949. The real imported-incident integration now expires its
job and proves 404 while ordinary workbook startup, imported projection, object
bytes and commit proof survive. It passes at 20260907T200853Z-p569461 (3/3).
No production backend change was needed for expiry.

Typecheck 20260907T200907Z-p584385 caught a Playwright-only exact option copied
into a Testing Library focus fixture; removed it. Typecheck/Biome now PASS at
20260907T201019Z-p591881 / 20260907T201019Z-p591891.
Generate PASS at 20260907T200908Z-p586183. agent-finalize PASS at
20260907T201020Z-p592276 before broader final verification. RESULTS_DIR was
unset; retained-run selection, canonical retained-evidence validation, scheduler
event/timing validation and performance-evidence maintenance were skipped.
Generated structure was unchanged; schema and tier coverage passed.

Final cache-off narrow owner evidence:

| Public Make route | Run root suffix | Result |
| --- | --- | --- |
| service-backed-test-slice platform.jobs: expiry/compaction, lifecycle/progress, transition contracts, failure security | 20260907T201131Z-p595832 | PASS 3/3 |
| service-backed-test-slice module.jobapi: all service rows | 20260907T201132Z-p596073 | PASS 3/3 |
| test-slice app.server: internal_error_boundary | 20260907T201133Z-p596939 | PASS 1/1 |
| test-slice platform.openapi | 20260907T201134Z-p599084 | PASS 4/4 |
| test-slice package.protocol_ts | 20260907T201140Z-p613775 | PASS 7/7 |
| test-slice package.ui | 20260907T201150Z-p640275 | PASS 10/10 |
| frontend-import-boundary-check | 20260907T201136Z-p602534 | PASS 2/2 |
| generate-drift | 20260907T201135Z-p602231 | PASS 4/4 |
| generated-artifact-policy-check | 20260907T201135Z-p602241 | PASS 3/3 |
| json-shape-check | 20260907T201135Z-p602251 | PASS 3/3 |

Run roots use .cartulary/test-results/<suffix>. The broader make check route is
justified by shared HTTP serialization and session caller migration.

The browser/accessibility/affected-visual slice passes on the availability fix at
20260907T201214Z-p642332 (15/15). The first make check at
20260907T201351Z-p692784 failed three of 747 work units; 744 passed. These are
verification regressions/pre-existing policy debt, not expected characterization
failures. The failing units and resolutions are:

- Selector-contract policy found 13 import heading-name readiness assertions in
  the three browser suites. They now locate the existing owned detail selector,
  then assert the heading text and applicable focus. No policy exception or new
  selector was introduced.
- Source-ownership policy found ten uncatalogued paths: seven existing import
  files and three new hardening/state files. The authored web.app/web.testing
  source ownership lists now cover them. Architecture rerun
  20260907T201827Z-p863427 passed two of three units while selector edits were
  still incomplete; corrected rerun 20260907T201924Z-p876326 passed all three.
- lint-go found pre-existing indentation in internal/app/server/module_settings.go
  and tools/contractgen/extensions_generation.go. The isolated lint-go-format
  failure at 20260907T202007Z-p881867 confirmed this. Make format at
  20260907T202020Z-p885954 corrected both. These whitespace-only changes are now
  retained to satisfy the required gate, superseding earlier work-log restorations.

agent-finalize passed again at 20260907T202052Z-p890274 before the second broad
check. RESULTS_DIR remained unset with the same retained-maintenance skips.
The generator's formatting changed its source-integrity digest; Make regenerated
only that digest and its enclosing artifact hash in
internal/gen/contractextensions/artifacts_gen.go. No extension schema or behavior
changed. No generated file was hand-edited.

Final typecheck and Biome pass at 20260907T202053Z-p890526 and
20260907T202053Z-p890536. make lint-go passes: format, vet and staticcheck
target summaries are 20260907T202053Z-p890633, 20260907T202055Z-p894616 and
20260907T202106Z-p905656 respectively. Fresh drift/policy/JSON reruns pass at
20260907T202215Z-p958367 (4/4), 20260907T202215Z-p958375 (3/3) and
20260907T202215Z-p958383 (3/3). The final browser/accessibility/visual rerun
after the selector corrections passes at 20260907T202215Z-p958442 (15/15).

The second make check passes at 20260907T202149Z-p909964: 747/747 work units,
1123 selected owner rows, zero failed/skipped/cancelled units, 326161 ms. This
broad run used normal cache mode (82 hits, 496 misses, 169 bypasses); the affected
narrow slices and final browser run above used cache off. No broad-run cache
hit is represented as a freshly executed characterization or browser result.
Final Markdown and whitespace checks follow the completed evidence record.

### Final gap acceptance disposition

PASS denotes implementation-support evidence. It does not publish a Core 05
conformance claim. The work-log tables supply the full run-root suffixes; all
new tests are routed through authored owner catalogs, never through Markdown.

| Gap | Final disposition and owner-to-evidence mapping |
| --- | --- |
| G1 | PASS. Human review reconciled Core 01 REQ-01-229/248/249, Core 03 REQ-03-290, Core 04 concealment and AC-261, and Design §4.5. Existing typed internal_error projection already matches 500; no wire projection change is required. module.jobapi service rows cover concealment/replay; platform.jobs.integration.expiry_and_compaction proves immediately-before and exact-cutoff behavior, including cancellation replay. The real module.incidentbundles integration proves expired-job output survival. |
| G2 | PASS. web.application.regression.incident_import_error_boundary/client and hardening exercise all three typed operations, malformed bodies/statuses, thrown/lost responses and invalid successes. Shared transport status survives parsing failure. Exact multipart recovery is also proved by web.design.browser.incident_import_real_workflow. |
| G3 | PASS. app.server.unit.internal_error_boundary and module.jobapi.unit.internal_error_safety prove fixed HTTP/envelope status, code/message, retry hint, empty details, omitted conflict and preserved shared request correlation. Individual private SQL/path/secret markers are absent from complete responses; actual GET/cancel handlers and typed public errors are covered. Full make check passes. |
| G4 | PASS. web.application.regression.incident_import_state/lifecycle/refresh_focus/hardening prove pure immutable transitions, independent snapshot availability/read activity, retained controls and unchanged handoff. The final availability regression ensures a failed read cannot reverse authoritative 404. |
| G5 | PASS. web.application.regression.incident_import_hardening/lifecycle prove fresh serial action reads, priority/deadline accounting, duplicate/departure fencing, loss of eligibility, exact dispatched cancellation replay and no side effect after failed preflight. Real browser action/access and import-to-workbook rows confirm integration. |
| G6 | PASS. Hardening covers large positive and negative browser-clock offsets; availability has no local clock dependency. platform.jobs expiry tests own server cutoff behavior. module.incidentbundles real integration proves workbook startup, imported projection, object bytes and commit proof remain after job expiry. |
| G7 | PASS. Import workflow/application-lifetime/hardening and session collaborator rows cover explicit confirmation failure/retry, stable identity, retirement and obsolete generations. module.auth passes 37/37 units. Real action/access browser coverage confirms ambiguous 404, repeated recovery and current profile loss; protected state is cleared before confirmed loss publication. |
| G8 | PASS. Central validation and state/client/hardening rows cover immutable job/submission/scope identity, timestamp/progress/lifecycle consistency, skipped states, duplicate/unsupported references, committed-target conflict and additive unknown references. Semantic incident IDs feed ordinary startup; API resource routes never navigate the browser. |
| G9 | PASS. Baseline, expected characterization failures, implementation failures/corrections, routed regressions and final results are distinct in this log. Real active/closed imports, upload/cancel recovery, six statuses, keyboard/layout/a11y and explicit launch remain covered. Three new and one changed golden were reviewed; two ordinary visual passes and the final 15/15 browser/a11y/visual run pass with unrelated golden bytes unchanged. |

### Final changed-file inventory and scope audit

This iteration changes 46 paths relative to its starting HEAD, including eight
new files. The original staged handoff edit remains staged and unchanged; all
iteration edits remain in the working tree. Historical II-01–II-05 path lists
above describe their earlier seam and are not this iteration's change inventory.

- Owners and controlling documentation: docs/spec/01_architecture_storage_and_view_contracts.md,
  docs/spec/03_workbook_interaction_collaboration_and_workflows.md,
  docs/spec/04_security_deployment_and_conformance.md, docs/design.md, and this
  handoff. docs/domain.md was inspected and remains unchanged.
- Import application/state/presentation: apps/web/src/app/incidentImportModel.ts,
  incidentImportState.ts (new), IncidentImportPanel.tsx, useIncidentImport.ts,
  api/incidentImportClient.ts, App.tsx, appSessionController.ts and
  accountSettingsModel.ts. The account caller change only adopts the neutral
  operation-observation name.
- Frontend transport and unit evidence: apps/web/src/services/browserApi.ts and
  httpTransport.ts; apps/web/src/app/incidentImportHardening.test.ts and
  incidentImportState.test.ts (new), incidentImportModel.test.ts,
  IncidentImportPanel.test.tsx, appSessionController.test.tsx and
  api/incidentImportClient.test.ts.
- Backend behavior and evidence: internal/platform/httpapi/httpapi.go,
  api_error.go and internal_error_test.go (new);
  internal/modules/jobapi/routes.go and error_test.go (new);
  internal/modules/incidentbundles/routes_admission_integration_test.go.
- Browser evidence: apps/web/e2e/incident-import.spec.ts,
  support/incidents/import.ts, workbook.a11y.spec.ts and workbook.visual.spec.ts.
  Under workbook.visual.spec.ts-snapshots, new
  incident-import-action-checking-linux.png,
  incident-import-action-unconfirmed-linux.png and
  incident-import-access-retry-linux.png; changed
  incident-import-observation-unavailable-linux.png. No other PNG changed.
- Authored verification/ownership: tools/test_families/app.server.json,
  module.jobapi.json, web.application.json and web.design.json;
  tools/frontend_source_ownership.json. Existing ownership omissions were repaired
  without changing routing authority or weakening architectural policy.
- Make-generated evidence/projections: tools/browser_e2e_batch_manifest.json,
  tools/execution_topology_render_index.json,
  tools/frontend_visual_golden_manifest.json and
  internal/gen/contractextensions/artifacts_gen.go. The latter changes only the
  formatted generator's integrity hash and containing artifact hash.
- Required formatter repair: internal/app/server/module_settings.go and
  tools/contractgen/extensions_generation.go, both whitespace-only. They add no
  feature behavior or stored-data change.

Inspection also covered domain navigation, existing protocol/error projections,
session/membership/workbook startup, auth/job admission/cancel/expiry paths,
verification routing and the visual maintenance guide. Package protocol/UI and
OpenAPI owner slices pass without changes to their wire shapes or selectors.
No lockfile, dependency, authored migration, endpoint or production configuration
changed. No runtime, generator, test, conformance or release-evidence input was
added under docs or other Markdown.

### Final verification limits and completion gate

All required product, service-backed, browser/accessibility, visual, type,
boundary, lint, generated-policy/drift and JSON checks have passed. Initial
characterization failures and subsequent implementation/policy failures have
successful closing evidence above; no product failure remains deferred.

Retained-run selection and retained canonical/scheduler/performance maintenance
were skipped by agent-finalize because RESULTS_DIR was unset at both invocations;
no qualifying successful full warm run was available then. The final make check
result does not retroactively claim those maintenance operations. Full CI,
release-check, alternate-browser certification, deployment and production
conformance publication are outside the authorized task and were not performed.
No required check is waived by those scope limits.

make lint-markdown passes at 20260907T202934Z-p1130629; its summary is
adhoc/lint-markdown/tool-run-summary.json. Both git diff --check and
git diff --cached --check pass. The scope audit confirms 46 changed paths, eight
new files, only the four approved PNG changes, and the original handoff-only
staged edit (151 insertions). Historical sections from the first section through
II-05's rollback are byte-identical to the original staged handoff; only the
iteration pointer and superseded hardening proposal were replaced.

All child workstreams and G1–G9 are complete. II-10 closes the overall effort;
no required work remains. A final Markdown/whitespace rerun checks this tracker
closure itself before the completion report. No commit, push or deployment was
performed.

### Final architecture, compatibility and rollback

The import-specific client validates transport outcomes once. Pure state owns
snapshots, independent availability/read activity, attempts, access and action
state. The executor owns serial scheduling, intent priority, monotonic deadlines,
abort/late-response fencing, and exact cancellation ownership. The React binding
owns visibility, native file clearing, subscriptions and conditional focus recovery.
App coordinates explicit session/profile confirmation, membership recovery and
ordinary workbook startup. No generic job framework or compatibility alias remains.

Core 01 owns fixed internal errors and admission/replay meaning; Core 04 owns
concealment and mandatory server expiry; Core 03 owns action/access requirements;
Design owns presentation and focus direction. domain.md was inspected and remains
unchanged as vocabulary/navigation authority. Executable checks do not consume
these documents. Existing generated wire shapes remain sufficient.

Compatibility effects are internal TypeScript interface changes, one fresh GET per
explicit Cancel/Open action, and restricted internal-error diagnostics. No new
endpoint, stored-data migration, dependency, job discovery, persistence, mode or
registry was added. Callers must use stable codes and request correlation instead
of private diagnostic strings.

Recovery remains local to this tab and session. Reload loses selected files,
attempts and locally known job history. After timeout or abort the server may
continue processing. Job retention controls resource observation, independently
of committed incident existence. Protected data cleared on retirement is not
resurrected after access recovery.

Rollback the frontend conversion coherently: adapter/transport, pure state,
executor, binding/panel, App/session callers, tests, authored routing, generated
routing and affected goldens/manifest. Restore owner interpretation consistently
with the behavior selected for rollback. The shared error security boundary and
its tests are independently revertible; a workflow rollback must not automatically
restore exception disclosure. No commit, push, deployment or production
certification is part of this effort.
