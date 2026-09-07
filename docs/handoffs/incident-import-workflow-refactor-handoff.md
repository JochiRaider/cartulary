# Incident import workflow refactor handoff

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
