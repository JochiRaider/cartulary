# Ordinary inspector editing and recovery handoff

## Baseline, authority and scope

Implementation baseline: clean `main`, HEAD
`d4cdca1f15aebe8771e469e158e679fa501f5837`, revalidated before editing.
No unrelated authored changes were present. No commit, push, deployment,
analyst-data changes, browser persistence or digest edits are authorized.

The digest localized read order and maintained frontend guides were read during
planning and current implementation entry points rechecked. The digest and
historical handoffs are advisory evidence, not behavior authority or fresh passes.

Behavior owners: Core 01 §3.3.5, §7.4, §18/18A/18B and §19; Core 02 §6,
§8, §10, §12 and §15; Core 03 REQ-03-291/292, §3, REQ-03-282,
REQ-03-299/100, §13, §16, §18 and §19; Core 04 §1-2. Domain supplies
vocabulary and owner navigation; design supplies bounded interaction direction.

Source placement is independently owned by `tools/frontend_source_ownership.json`
and `tools/frontend_import_boundaries.json`. Verification routing is independently
owned by `contracts/verification`, the test catalog and authored test families.
Executable product/test/generation inputs do not depend on this handoff or Markdown.

Bounded changes: ordinary inspector editing and draft presentation; retained
explicit record PATCH machinery and actual Task/Party consumers; the internal
view-contract projection; source guides, semantic selectors and affected tests.
Timeline hot-path editing, creation, retained batches, specialized Task lifecycle,
Party linking, Decision supersession, Entity merge, Evidence attachment, History,
Notes/Assessment/contextual creation remain regression boundaries.

## Sequential workstreams

| Workstream | Status | Exit / next action |
| --- | --- | --- |
| IER-01 | DONE | Controls/transitions frozen; six executable gap reproductions and owner clarification recorded. |
| IER-02 | DONE | Canonical subjects, patch capability, scoped raw drafts and matrix binding pass. |
| IER-03 | DONE | Ordinary, Task and Party consumers migrated; exact replay and read recovery pass. |
| IER-04 | DONE | Integrated field/action, source, security, spreadsheet, browser and accessibility evidence passes. |
| IER-05 | DONE | Completed candidate validated before promotion; scoped checks pass, baseline failures isolated, rollback recorded. |

Only the current row may be IN_PROGRESS. Each exit is saved as DONE before
starting its dependent. An unresolved adopted-owner contradiction blocks progress.

## Retention decision and transition matrix

The user explicitly selected memory-local drafts per original record and
field/action, with Resume/Discard after detachment. This is an authorized Core 03
clarification: invalidation removes form/review authority, not raw authoring.

| Transition | Form / review / subject attachment | Raw authoring | Captured attempt / receipt |
| --- | --- | --- | --- |
| Same-row object replacement, unchanged values | Derive current subject; preserve eligible presentation. | Preserve exact dirty text and intent. | Unchanged. |
| New version, unrelated field | Invalidate prior authority; rederive eligible form from unchanged dependencies. | Preserve baseline field values and raw text. | Never change dispatched bytes. |
| Edited/dependent field changed | Stale review; submission disabled until explicit review. | Preserve complete raw draft separately from saved values. | Existing conflict/recovery remains independent. |
| Field/action switch | Detach old editor; new editor uses its own identity. | Park old draft; explicit Resume/Discard on return. | Unchanged. |
| Close, record retarget, no selection | Invalidate old form/errors synchronously; no old content under new subject. | Park per original record; no blocking navigation. | Retain independent lifetime. |
| Sheet/saved-view switch | Close before new surface becomes interactive. | Park; same original schema/record required for resume. | Retain read debt even when source is unmounted. |
| Delete, merge, filtered-out row, unavailable field | No actionable old form; never target a survivor automatically. | Park without displaying under another subject; resume only original eligible identity. | Preserve outcomes; authority controls recovery. |
| Incident closes / role becomes viewer | Invalidate write authority; no fresh/replay mutation. | Retain readable work only under current read authority. | Acknowledged reads may reconcile. |
| Uncertain authorization / same-account session loss | Conceal protected form, references and errors; require current authorized read. | Retain concealed within existing account runtime. | Retain concealed and accept correlated late receipts without presentation. |
| Incident access revoked | Existing incident-exit and scoped lifetime apply; no automatic reentry. | No cross-incident archive or disclosure. | Existing owner lifetime; never compensate committed writes. |
| Account replacement / runtime retirement / hard reload | Old attachment is invalid. | Clear on retirement; no reload/storage guarantee. | No obsolete callbacks into the replacement lifetime. |
| Older acknowledgement, newer typing/navigation | Effects require matching presentation and authoring revision. | Clear only captured revision; keep newer work. | Retain full receipt before read/presentation effects. |

Grid and inspector drafts remain independent. Raw inspector work alone does not
block grid typing or autosave. A saved grid change makes overlapping inspector
review stale; inspector submission never includes unrelated grid or parked work.
An admitted explicit operation coordinates existing writes and protects its
target until its outcome is resolved, without making the inspector mandatory.

## Control and lifecycle ownership

| Family | Current source / guarantee | Disposition and target boundary |
| --- | --- | --- |
| Entity scalar editor | Entity composition has independent `editRecordId`; `selectedEntity` owns header. Local value resets on row-object replacement; one-shot entity command mints fresh ID. | Remove row picker/state. Retain field picker/editor for patch-writable fields, bound only to canonical selected entity. |
| Entity alias add/remove | Selected entity is target; local alias text; one-shot patch and unguarded completion focus/selection. | Retain suggestion-only actions and stable item refs; move to retained lifecycle with revision-fenced completion. |
| Generic scalar/collection editor | Grid selection is target but row picker duplicates selection UI. Non-Task local values reset on row objects; generic command maps uncertainty to retryable rejection. | Remove row picker; retain existing fields/actions, with durable drafts and retained operation. |
| Ordinary Task inspector | Per-record Task draft store and retained explicit operations; baseline dependency checks exist. Completion clears by field without authoring revision. | Migrate ordinary draft ownership; retain Task guard contribution and exact operation semantics. |
| Task compound lifecycle | Task draft store and compound conflict restrictions inside explicit owner. | Keep domain form/guards with Coordination; neutral execution consumes its contribution. |
| Party explicit patch consumers | Source-review preparation and recovery attached to explicit owner. | Keep reviewed Party semantics with Parties; neutral execution consumes its contribution. |
| Specialized actions | Dedicated creation, linking, supersession, merge, Evidence and History owners. | Retain owner workflows; change only internal shared-lifecycle integration required by this seam. |

All retained ordinary edits use `PATCH /api/v1/records/{record_id}` and the active
schema's existing field/action contract. Scalar controls keep raw value plus saved
baseline/version; collections keep ordered actions and stable item/reference IDs.
Current account/incident/role/lifecycle determines admission. Query/Collaboration
owners admit saved rows monotonically; inspector attachment owns local focus;
retained operation ownership owns outcomes and read recovery.

Remove exposed create-only Host `aad_device_id`/`fqdn`, Identity
`aad_object_id`/`sid`, and Indicator fields from ordinary edit pickers. Preserve
creation capabilities and append-only Assessments. No public API or DB migration.

## Gap register

| ID | Remediation / affected areas | Rationale / long-term benefit | Compatibility and unresolved risk | Binary validation |
| --- | --- | --- | --- | --- |
| G01 | Canonical selected subject; Entity/generic inspector controls. | Remove duplicate target authority and wrong-record writes. | Retire row selectors; accessible field reach retained. Risk: obsolete callbacks. | Inspect A cannot patch B without canonical retarget; no-row sends no write. |
| G02 | Carry authored writable into internal patch capability; projection and preparation. | Separate create shape from existing-record permission. | Additive internal type only. Risk: create-only exposure. | Every retained field is patch writable; create-only fields never dispatch. |
| G03 | Memory-local draft owner independent of query objects; inspector and Task integration. | Preserve authoring with explicit identity and attachment. | Authorized retention clarification. Risk: accidental reattachment or data loss. | Refresh, field switch, close and retarget preserve original work without misbinding. |
| G04 | Dependency review and explicit clear/reference intent; neutral preparation and source contributions. | Avoid hidden rebasing, destructive normalization and lost option identity. | No new writable semantics. Risk: owner-coupled fields. | Unrelated refresh preserves work; edited/dependent changes require review; exact null/reference payloads. |
| G05 | Neutral explicit admission/capture/recovery; runtime and Task/Party owners. | One retained lifecycle instead of one-shot or copied engines. | Actual consumers migrate together. Risk: ordering deadlock or overlapping writes. | Single reservation; prior autosaves settle; exact uncertain bytes/ID; no automatic rekey. |
| G06 | Durable correlated receipt plus existing read debt. | Acknowledged success survives failed refresh/detachment. | Preserve Collaboration and version high-water marks. Risk: stale read acceptance. | Recovery after acknowledgement sends reads only and creates no extra revision. |
| G07 | Authoring revision and presentation fences. | Older completion cannot erase later work or restore obsolete selection. | No change to current applicable continuity. Risk: focus theft. | Newer typing/navigation survives every delayed success/error path. |
| G08 | Scoped security lifetimes, recovery presentation and matrix evidence. | Retention never becomes renewed authorization. | Existing session/incident boundaries preserved. Risk: protected disclosure. | Suspension conceals; same account recovers; replacement clears; obsolete callbacks have no effect. |

## Execution evidence

Planning baseline only: focused explicit Task and batch tests passed 3/3 units at
`.cartulary/test-results/20260914T134956Z-p55225`. This is not new-seam evidence.
Task guides were read for web.workbook, module.workbook, module.entities,
module.artifacts, module.parties, module.tasksdecisions, module.timeline,
package.view_contracts and web.architecture.

IER-01 exit: all ordinary controls inherit an explicit family lifecycle and each
field/action is enumerated below. Core 03 now records the user-selected ordinary
retention contract; specialized review and security lifetimes remain intact.
No adopted-owner contradiction was found.

Characterization commands:

- `make test-slice OWNER=web.workbook ROWS=web.workbook.regression.inspector_entity_editing,web.workbook.regression.inspector_ordinary_recovery`:
  expected FAIL at `.cartulary/test-results/20260914T140238Z-p60738`.
  Generic response loss minted a new request and acknowledged generic edits had
  no retained receipt. Entity tests demonstrated redundant targeting, absent
  draft return and obsolete selection restoration. One refresh assertion used an
  unavailable matcher; it was corrected before the final reproduction run.
- `make test-slice OWNER=web.workbook ROWS=web.workbook.regression.inspector_entity_editing`:
  expected FAIL at `.cartulary/test-results/20260914T140320Z-p61582`.
  Tests exercise actual wrong-record dispatch, dirty refresh, field return and
  detached completion. Detailed assertion evidence is in each run's
  `unit-logs/row-web.workbook.regression.inspector_entity_editing/runner.json`.

Tests are ordinary failing regressions, not skipped or expected-failure tests;
dependent work must make them pass. The earlier temporary expected-failure
characterization pass is not product readiness evidence. Risks G01-G08 remain
open until their implementation and integrated verification exits. Next: IER-02.

### IER-02 exit

Implemented `WorkbookInspectorDraftStore`, its canonical binding hook, local
feedback, field control and patch preparation under `workbook/inspector`.
Entity and generic composition now use canonical selected rows, with no ordinary
row picker or fallback field. Authored writable is projected as `patchWritable`
by the maintained generator; create-only capabilities remain separate.

Drafts retain exact raw values, explicit null intent, reference identities,
reviewed saved values and authoring revisions. Close/retarget/field/surface
changes detach; only explicit Resume reattaches. Saved-field/dependency changes
require review. The runtime scopes concealment and retirement to current
account/incident authority. An alias removal does not consume an add draft.
Collection removal preserves owner-defined prefixed `item_ref` identities;
these are not record UUIDs.

- `make generate`: PASS, `.cartulary/test-results/20260914T140619Z-p63041`.
- `make test-slice OWNER=web.workbook ROWS=web.workbook.regression.inspector_entity_editing,web.workbook.regression.inspector_draft_lifetime,web.workbook.regression.inspector_draft_binding`:
  PASS 4/4, `.cartulary/test-results/20260914T141823Z-p71887`.
  Binding evidence iterates every retained scalar and add/remove action.
- `make frontend-typecheck`: PASS 2/2,
  `.cartulary/test-results/20260914T141846Z-p72988`.
- Earlier typecheck/test failures were fixture typing and a wrong Task schema
  constant; both were corrected before the passes above. Initial `make format`
  and `make lint-biome` exposed an unnecessary effect dependency, empty import
  and non-null fixture assertions; corrected locally. Final lint remains part
  of IER-05.

No cross-record resume, storage, API or database migration was introduced.
G01-G04 draft/binding criteria pass focused checks; integrated evidence remains
IER-04. G05-G08 operation recovery and continuity require IER-03. Existing
one-shot transports are intentionally still present until that migration.
Next action: save this exit, then begin IER-03.

### IER-03 exit

`WorkbookExplicitPatchOwner` now owns neutral synchronous reservation, immutable
scope/intent/request capture, correlated receipts, replay and reconciliation.
Its policy contribution interface carries dependency review, preparation,
validation, authority rechecks and source effects. Task guards/drafts stay in
Coordination; Party review stays in Parties. Another contract-backed field uses
its patch metadata and a source contribution, without another operation owner.

Entity scalar/alias and generic ordinary edits now use this lifecycle. A captured
draft revision is acknowledged before refresh effects; newer authoring survives.
The shared surface registry retains refresh debt; generic and Entity query
owners reject regressing versions and admit receipts monotonically. Surface-local
recovery stays reachable with the inspector closed. No completion selects a row or reopens an inspector. Focus effects require the
captured presentation and control to remain applicable. Specialized Task notices remain with Tasks.

Retired: `createGenericMutationCommandPort.ts`, generic/Entity one-shot patch
interfaces and factory branches, `selectWorkbookEditTarget`, independent ordinary
Entity pending state, Task-shaped ordinary inspector draft storage and the
explicit owner's Task-only refresh callback. Actual Task/Party callers and tests
were migrated. No production compatibility shim remains.

- `make test-slice OWNER=web.workbook ROWS=web.workbook.regression.party_link_recovery,web.workbook.regression.workbookz_mutation_runtime_semantic_command_ports_7b2f1d02a4,web.workbook.regression.explicit_task_patch_recovery,web.workbook.regression.inspector_entity_editing,web.workbook.regression.inspector_ordinary_recovery`:
  PASS 6/6, `.cartulary/test-results/20260914T143714Z-p96033`.
- `make test-slice OWNER=web.workbook ROWS=web.workbook.regression.inspector_ordinary_recovery,web.workbook.regression.batch_operation_retention,web.workbook.regression.entity_merge_admission,web.workbook.regression.workbook_save_status_preserves_global_blockers_a_8d590e0883`:
  PASS 5/5, `.cartulary/test-results/20260914T143827Z-p1918`.
- `make frontend-import-boundary-check`: PASS 2/2,
  `.cartulary/test-results/20260914T143829Z-p2183`.
- `make frontend-typecheck`: PASS 2/2,
  `.cartulary/test-results/20260914T143941Z-p3731`.
- `make format`: PASS 2/2,
  `.cartulary/test-results/20260914T143940Z-p3531`.

Corrections during migration: preserved Party target-denial authority recheck;
removed a test-only duplicate surface refresh registration; updated old partial
Entity success fixtures to complete envelopes. Preparation failure retains raw
intent without claiming a dispatched request. A definitive transaction-ID
rejection permits a separately reviewed action; a refusal after uncertainty
remains uncertain. Neither path automatically rekeys or rewrites an attempt.

G05-G07 focused lifecycle criteria pass. Integrated source/security, browser,
accessibility and visual evidence remains IER-04. Shared query admission,
reservation ordering and removed selector expectations require those regressions.
Next action: save this exit, then begin IER-04.

### IER-04 exit

All G01-G08 criteria pass the integrated checks below. The field/action binding
suite iterates the complete frozen matrix through refresh, detachment, no-row,
unavailable field, explicit return, viewer/closed state, suspension, same-account
recovery and account replacement. Reference controls retain an off-page selected
identity and label; explicit clearing preserves null intent. The ordinary
operation suite verifies high-water refresh, edited-field re-review, exclusion
of later grid authoring, collection conflict class and immutable captured bytes.

Production-renderer scenarios in `apps/web/e2e/workbook-inspector-edit.spec.ts`
verify inspected A/B targeting, same-row dirty refresh, field return, stale review,
response loss after commit, malformed success, exact replay, later authoring,
closed-inspector recovery and newer grid focus through delayed completion.
History counts remain two (creation plus one patch), including replay/read recovery.
The existing service tests verify collection no-effect rejection, stable item
identity, source revisions and mutation replay. Tests use managed incident fixtures;
no existing analyst records were edited.

Integrated corrections and decisions:

- Entity grid active-cell selection now sets the canonical inspected record;
  draft/no-row focus clears it. Generic no-row selection also clears its subject.
- Alias acknowledgement restores its input only if the original inspector and
  focused control are still applicable. A detached completion cannot steal grid
  focus. Alias typing continues to invalidate the Entity merge review.
- The public-schema test fixture now omits internal `patchWritable`; no HTTP
  discovery field was added. Stale one-shot transaction-prefix expectations were
  updated to the neutral operation identity.
- Party creation preparation excludes exactly its own synchronous reservation,
  while other operations remain blocked. Party contributes complete source
  reconciliation, including mounted refresh or retained unmounted surface debt;
  the neutral default is the required originating-surface refresh.
- The old Coordination browser reference helper expected a permanently mounted
  select. It now opens the existing reference control and selects the fixture's
  declared reference surface. No reference product capability changed.
- New recovery buttons use the existing inspector button primitive. Reviewed
  captures at 1280x720 and 390x720 show named controls, visible focus, bounded
  content and existing independent drawer scrolling. No golden promotion needed.

Commands below use public root Make targets. Row selections are recorded in each
run's `run-manifest.json`, `rows/` results and target summary; browser evidence is
under its `browser-*/browser-groups/` directories.

| Command / selected evidence | Result | Run root under `.cartulary/test-results/` |
| --- | --- | --- |
| `make service-backed-test-slice OWNER=module.workbook` — Evidence create/patch, generic collections, Notes/Task/Decision mutation rows | PASS 3/3 | `20260914T144133Z-p9409` |
| `make service-backed-test-slice OWNER=module.workbook` — inspector subject retention, exact replay, detached refresh, a11y recovery | PASS 13/13 | `20260914T145228Z-p5040` |
| `make service-backed-test-slice OWNER=module.timeline` — spreadsheet pointer, keyboard, creation, refresh, batch replay | PASS 11/11 | `20260914T144652Z-p67207` |
| `make test-slice OWNER=web.workbook` — ordinary creation lifecycle, shell surfaces, Entity editing, binding and recovery | PASS 6/6 | `20260914T145449Z-p40321` |
| `make service-backed-test-slice OWNER=module.workbook` — ordinary creation matrix, continuity, detached refresh, closure, revocation, uncertain Timeline, Host/Identity paste recovery | PASS 15/15 | `20260914T145514Z-p41540` |
| `make test-slice OWNER=web.workbook` — inspector review authorization, History operation/Timeline/recovery surfaces, Party socket order, merge query concealment | PASS 7/7 | `20260914T145718Z-p42920` |
| `make service-backed-test-slice OWNER=module.entities` — Host/Identity merge exact replay and History rollback | PASS 4 browser scenarios; combined run's separate provenance row failed, repaired below | `20260914T145716Z-p42667` |
| `make service-backed-test-slice OWNER=module.entities` — Host/Identity/Notes grid provenance and inspector edits | PASS 11/11 | `20260914T150029Z-p41830` |
| `make service-backed-test-slice OWNER=module.workbook` — inspector a11y, tabs/Evidence/history accessibility, default-closed/no-row state, inspector visual fixture | PASS 17/17 | `20260914T150030Z-p42227` |
| `make service-backed-test-slice OWNER=module.tasksdecisions` — Task lifecycle and Decision supersession recovery | Seven scenarios PASS; combined run's stale option helper failed, repaired below | `20260914T145530Z-p69955` |
| `make service-backed-test-slice OWNER=module.tasksdecisions` — native Task/Decision workflow | PASS 11/11 | `20260914T150225Z-p10780` |
| `make service-backed-test-slice OWNER=module.parties ROWS=module.parties.browser.source_link_recovery` | PASS 11/11, seven scenarios | `20260914T150354Z-p10876` |
| `make service-backed-test-slice OWNER=module.artifacts` — source mutations and collection semantics | PASS 3/3 | `20260914T150226Z-p11056` |
| `make test-slice OWNER=web.workbook` — matrix binding, ordinary recovery and Party retained patches | PASS 4/4 | `20260914T150251Z-p76297` |
| `make test-slice OWNER=package.view_contracts` | PASS 5/5 | `20260914T150308Z-p79837` |
| `make test-slice OWNER=package.ui` | PASS 10/10 | `20260914T150314Z-p86729` |

Failed evidence retained: `20260914T144634Z-p38203` exposed missing Entity
active-cell binding; `20260914T144606Z-p33268`, `20260914T145227Z-p4790` and
`20260914T145335Z-p39223` exposed the schema fixture/internal-field mismatch and
alias focus assertion; `20260914T145531Z-p70712` exposed Party self-reservation;
`20260914T145855Z-p78143` exposed detached Party refresh ownership;
`20260914T145856Z-p78384` exposed the stale Coordination reference helper.
Their corrected paths have passing replacements above. A browser build in
`20260914T150224Z-p10555` correctly rejected inputs edited during its snapshot;
it ran no Party scenarios and was superseded by the stable-input pass.

No unresolved seam behavior or security risk remains from these failures.
Full frontend/source-boundary checks, drift, Markdown, applicable measurements
and the paste/bulk broad-failure audit remain terminal IER-05 work. No broad
repository readiness is inferred. Next action: save this DONE exit, then start
IER-05 and validate the complete handoff bytes.

### IER-05 terminal evidence

The final implementation retains one neutral explicit operation lifecycle and
one ordinary raw-draft owner. Final corrections also fence field-local errors on
retargeting, conceal operation snapshots when the role is unavailable, and return
keyboard focus after Resume/Discard/review only while the initiating attachment
and focused control remain applicable. Existing Entity grid/notice geometry now
uses the shared layout owner. These changes add no grid confirmation step.

Source placement and guides were updated under `workbook/inspector`, `runtime`,
`models`, `hooks`, and the affected Entity, generic and Coordination features.
`tools/frontend_source_ownership.json` registers every new authored frontend file;
`tools/test_families/web.workbook.json` and `module.workbook.json` register the
new regression and production-browser rows. The authored projection generator
and generated view facade change together. Browser/topology derivatives were
regenerated from the catalog. The Markdown lint input now includes this exact
handoff path; this is documentation maintenance, not product verification input.

The removed ordinary row-selector semantic helper remains only in negative
characterization/contract fixtures; no production row picker remains. Existing
runtime Entity/Decision write reservations retain actual creation/conflict
consumers. The unused Decision boundary interface and ordinary one-shot patch
ports are retired. There is no compatibility shim or second patch lifecycle.

Terminal commands below follow `make agent-finalize` (RESULTS_DIR unset).
Unit counts include Make-owned prerequisites. Exact selected row IDs and runner
commands are retained in each root's `run-manifest.json`, `rows/*.json`, and
`unit-logs/`; `run-summary.json` and `target-summaries/` contain terminal results.
All relative run roots in this section start at `.cartulary/test-results/`.

| Command / selection | Result | Run root |
| --- | --- | --- |
| `make agent-finalize` | PASS 1/1; retained-run maintenance skipped because RESULTS_DIR is unset | `20260914T152910Z-p1502` |
| `make test-slice OWNER=web.workbook ROWS=web.workbook.regression.inspector_entity_editing,web.workbook.regression.inspector_draft_binding,web.workbook.regression.inspector_ordinary_recovery,web.workbook.regression.inspector_draft_lifetime` | PASS 5/5 | `20260914T152430Z-p54484` |
| `make test-slice OWNER=web.workbook` — draft binding, Entity editing, ordinary recovery, Task retained patches and Party recovery | PASS 7/7 | `20260914T152123Z-p79089` |
| `make service-backed-test-slice OWNER=module.workbook ROWS=module.workbook.browser.inspector_subject_retention,module.workbook.browser.inspector_exact_replay,module.workbook.browser.inspector_detached_refresh,module.workbook.accessibility.inspector_edit_recovery` | PASS 13/13 | `20260914T152307Z-p20397` |
| `make service-backed-test-slice OWNER=module.workbook ROWS=module.workbook.accessibility.inspector_edit_recovery` | PASS 11/11; final root-zoom/text-spacing profile and keyboard Resume | `20260914T152429Z-p54220` |
| `make service-backed-test-slice OWNER=module.workbook` — four inspector scenarios plus native Coordination workflow | PASS 15/15 | `20260914T152126Z-p79690` |
| `make service-backed-test-slice OWNER=module.workbook ROWS=module.workbook.browser.canonical_create_independent_recovery` | PASS 11/11; previously intermittent Indicator recovery regression revalidated | `20260914T152902Z-p87096` |
| `make service-backed-test-slice OWNER=module.timeline` — committed typing acknowledgement, blank-row creation, ArrowDown and Enter measurement rows | PASS 20/20 | `20260914T150306Z-p79619` |
| `make frontend-unit` | FAIL 611/614; three baseline static-policy rows isolated below; all functional rows passed | `20260914T150855Z-p65161` |
| `make test-slice OWNER=web.architecture` | FAIL 10/12 after repairing touched-file violations; two unchanged baseline rows isolated below | `20260914T151311Z-p37835` |
| `make frontend-import-boundary-check` | PASS 2/2 | `20260914T152239Z-p19634` |
| `make frontend-typecheck` | PASS 2/2 | `20260914T152939Z-p21485` |
| `make lint-biome` | PASS 2/2 | `20260914T152939Z-p21492` |
| `make generate` | PASS 1/1 | `20260914T151603Z-p21133` |
| `make generate-drift` | PASS 4/4 | `20260914T152939Z-p21253` |
| `make generated-artifact-policy-check` | PASS 3/3 | `20260914T151507Z-p14922` |
| `make json-shape-check` | PASS 3/3 | `20260914T150759Z-p57135` |
| `make lint-markdown` — complete terminal candidate | PASS; updated candidate revalidated before promotion | `20260914T153222Z-p27593`, `adhoc/lint-markdown/tool-run-summary.json` |

The browser accessibility report at
`20260914T152429Z-p54220/browser-e2e-a11y/browser-groups/a11y-workbook-inspector-edit/playwright-report.json`
contains reviewed PNG attachments for 1280x720, 390x720 and 1280x720 at 200%
root zoom, plus the accessibility tree. Existing dark theme, inspector scrolling,
visible focus, named controls and readable local feedback remain usable with
text spacing and reduced motion. Controls below the viewport remain in the
inspector's scroll container. The maintained inspector visual fixture passed in
`20260914T150030Z-p42227`; no golden was changed or promoted. These are scoped
implementation-support artifacts, not a repository-wide accessibility claim.

### Broad-failure isolation

**Remediation follow-up (2026-09-14):** The table below is historical provenance
for three policy units and four browser failures. Baseline reproduction does not
classify all seven as stale tests. The [seven-gap remediation handoff](frontend-policy-remediation-handoff.md)
records corrected downstream assertions, the confirmed Indicator creation versus
patch eligibility defect, completed/cancellable focus, shared Coordination/Note
recovery layout, and refused Timeline draft loss exposed by the real overflow
scenario. Read its current dispositions and run evidence alongside this retained
comparison; none of the historical failures below is retroactively a pass.

A detached comparison checkout at `/tmp/cartulary-ier-baseline-d4cdca1` remains
clean at the exact baseline HEAD. It was bootstrapped through
`make frontend-install`; initial missing-AJV preflight failures and a transient
registry retry are setup evidence only. Baseline run roots below are under that
checkout's `.cartulary/test-results/`. No authored baseline files were changed.

| Failure / current evidence | Exact baseline comparison | Disposition |
| --- | --- | --- |
| Harness architecture literal schema bindings; broad frontend run `20260914T150855Z-p65161` | `make test-slice OWNER=harness.browser ROWS=harness.browser.boundary_support.architecturepolicy_suite_4e0bacb131`: FAIL 1/2, `20260914T151525Z-p20011` | Same five assertions in unchanged test-support files. Outside inspector lifecycle ownership. |
| Workbook layout policy; `20260914T151311Z-p37835` | `make test-slice OWNER=web.architecture` selecting layout/selector policy: FAIL 1/3, `20260914T151524Z-p19764` | Same existing Coordination creation `100vh` geometry. Touched Entity geometry repaired through shared owner; no new layout-policy violation remains. |
| Selector contract policy; `20260914T151311Z-p37835` | Same baseline policy run `20260914T151524Z-p19764` | Four unchanged selector assertions in `ordinaryCreateControls.test.tsx` and measurement `timingSupport.ts`. Touched reference and Notes selector violations repaired with semantic owner selectors. |
| Legacy live-grid shortcut assertion; `20260914T151313Z-p38461` | Baseline service slice FAIL 10/17, `20260914T151604Z-p21739` | Tab expects `date_entered`, receives `analyst_text` on both inputs. Existing Timeline spreadsheet and measurement selections pass; keyboard dispatch was not redesigned. |
| System-view switcher Indicator creation focus; `20260914T151313Z-p38461` | Same baseline service run `20260914T151604Z-p21739` | Same inactive `indicator_type` editor assertion. This seam removes unsupported existing-record Indicator edits; it does not change Indicator creation focus. |
| Queue-overflow save-status fixture; `20260914T151313Z-p38461` | Same baseline service run `20260914T151604Z-p21739` | Same missing/focus editor assertion before overflow verification. Other selected queue cases pass. Autosave/retained-batch ownership remains unchanged. |
| Saved-view browser helper layout payload; `20260914T151333Z-p69748` | Same baseline service run `20260914T151604Z-p21739` | Same expectation for unchanged `layout_json` in a PATCH that omits it. Saved-view serialization is outside this seam. |

The baseline service comparison selected exactly:

- `module.workbook.browser.required_keyboard_shortcuts_operate_on_the_live_3ff4a5c6a9`
- `module.workbook.browser.verify_system_views_switcher_keyboard_entry_rovi_90a2f62956`
- `module.workbook.browser_stateful.workbook_save_status_preserves_authoritative_tra_0fa8b0b16c`
- `module.workbook.browser_support.verify_browser_command_helpers_for_sort_filter_g_cfa68b33e4`

The overlapping Coordination option-availability and local-error selector failure
was in changed test support and was repaired; its fresh passing replacement is
`20260914T152126Z-p79690`. Earlier generation/topology and lint/type failures
(`20260914T150641Z-p50133`, `20260914T151506Z-p14714`,
`20260914T152058Z-p74722`, `20260914T152125Z-p79366`) were corrected before
passing replacements. Failed runs remain evidence, not hidden passes.

These reproduced baseline failures limit broad readiness claims; none establishes
an unresolved inspector-seam dependency. No full `make check`, CI, release,
security publication, visual golden promotion or repository-wide conformance
claim was attempted. Service-backed checks use isolated harness fixtures. Backend
algorithms, migrations and analyst data were not changed. Full broad suites were
not repeated after the focused repairs because unchanged failures are isolated
and the affected replacement rows have fresh passing evidence.

## Final acceptance and rollback

The digest acceptance assessment is scoped to the changed inspector/editing
boundary. PASS means the applicable part is supported by the evidence above;
it does not convert unrelated baseline failures into repository-wide passes.

| Acceptance | Result | Scope and evidence |
| --- | --- | --- |
| A001 Authority | PASS | Core clause map, authorized raw-lifetime clarification, independent source/verification owners. |
| A002 Scope | PASS | G01-G08 rubric, neutral draft/operation boundaries, metadata plus owner contribution extension path and retired callers. |
| A003 Repository state | PASS | Clean main baseline, maintained guides/manifests, final diff and clean comparison checkout. |
| A004 Tokens | PASS | Existing inspector action/control primitives and shared layout owner; no new token or theme registry. |
| A005 Theme | PASS | Current dark_graphite production captures and unchanged theme projection. |
| A006 Density | PASS | Grid density owner unchanged; existing inspector visual fixture and Timeline measurements pass. |
| A007 Creation | PASS | Ordinary creation/paste matrix, specialized native workflows and schema-fixture regression passes; patch capability remains separate. |
| A008 Responsive algorithm | N/A | No viewport, CSS-length accessor, breakpoint or inspector clamp algorithm changed. Local narrow/zoom controls are covered by A019-A020. |
| A009 Overflow | PASS | Existing independent grid/drawer ownership retained; bounded recovery strip and reviewed narrow/zoom captures. |
| A010 Inspector | PASS | Canonical subject, no-row, eligibility, matrix binding, review invalidation and detached completion tests. |
| A011 Continuity | PASS | Dirty refresh, field switch, explicit return, stale dependencies and newer authoring/focus browser evidence. |
| A012 Transactions | PASS | Secure IDs, synchronous duplicate reservation, immutable capture, exact replay and receipt-correlation tests. |
| A013 Acknowledgement/recovery | PASS | Read-only acknowledged recovery with server history counts; exact uncertain bytes; retained batch/Task/Party regressions. FIFO recovery remains its owner's path. |
| A014 Editing | PASS | Independent grid/inspector drafts, scalar validation/null/reference/action coverage, Timeline spreadsheet scenarios and no persistence. |
| A015 Conflict | PASS | Existing scalar/text/collection review classes and Task compound guards; current baseline versus retained draft remains explicit. |
| A016 Query/interaction states | PASS | Monotonic saved-row admission and required high-water refresh; no-row/unavailable, closed/viewer and unknown-authority matrix. |
| A017 Security lifetime | PASS | Draft/operation concealment, same-account resumption, incident closure/revocation boundaries, account retirement and obsolete callbacks. |
| A018 Evidence overlay matrix | N/A | Evidence attachment/preview/lifecycle overlay policy unchanged. Existing ordinary Evidence patch and accessibility regressions pass. |
| A019 Accessibility | PASS | Production keyboard Resume, visible focus, named local controls, non-color status, reduced motion and root-zoom profile. |
| A020 Components | PASS | Existing primitives; text spacing, narrow/zoom, off-page references, null intent and deterministic detached/stale states. |
| A021 Virtualization | PASS | No row virtualization algorithm changed; production continuity and owner-selected typing/creation/selection measurements pass. |
| A022 Visual fixtures | PASS | Maintained inspector visual case passes; current renderer captures inspected before accepting evidence; no golden promotion. |
| A023 Selectors | PASS | New/touched selectors use schema/record/field or semantic accessible names; package.ui 10/10. Unchanged broad-policy failures isolated above. |
| A024 Test authority | PASS | Product/tests/generators have no Markdown dependency; lint includes this document solely as documentation maintenance. |
| A025 Generated artifacts | PASS | Authored generator/catalog inputs precede generated facade/topology outputs; drift and artifact-policy checks pass. |
| A026 Compatibility | PASS | No API/schema/DB migration; actual reservation consumers retained, ordinary one-shot paths removed, coherent rollback below. |
| A027 Handoff | PASS | Complete sequential exits, gap dispositions, paths, commands/artifacts, failure isolation, limits, skipped checks and validated terminal bytes. |

Digest rule dispositions: ADOPT R001-R011 and R013-R015 for keyboard/focus,
local feedback, retained recovery, semantic controls, continuity and owner-aligned
verification. ADAPT R012, R017-R018, R022, R024-R025 and
R035 to existing Cartulary spreadsheet, security, component and verification
owners. Preserve R026-R034 rejection boundaries: no upstream palette, invented
navigation, generic redesign or unsupported product contract. R016, R019-R021
and R023 add no separate changed policy in this seam. The digest itself is intact.

No unresolved G01-G08 acceptance criterion remains. Residual limitations are the
explicit memory/runtime lifetime and the isolated baseline failures above.
Retained raw work is never renewed authorization. Reload persistence and
cross-account recovery remain excluded. RESULTS_DIR was unset throughout;
retained successful-run maintenance was skipped for that reason.

Rollback must restore the compatible Core clarification, authored projection
adapter and generated facade, neutral lifecycle owners and source contributions,
consumer integrations, tests/routing derivatives and source guides as one set.
No database or external API migration is needed. Preserve unrelated work and
committed analyst data. Never reverse accepted edits to compensate for a UI
rollback. Do not restore a mixed caller/owner interface set.

Terminal promotion uses a complete candidate with the DONE tracker bytes. The
public `make lint-markdown` target must pass on that candidate and the canonical
handoff before the identical candidate is promoted and its temporary copy
removed. The terminal target output supplies the exact lint run root and its
`adhoc/lint-markdown/tool-run-summary.json`; this final validation receipt
accompanies delivery. Final
`git diff --check` and scope inspection follow promotion. Thus the completed
handoff bytes, rather than an earlier draft, are validated before DONE is saved.
No digest, lockfile or database file changed; no commit, push or deployment was
performed. Next action: review this completed seam; implementation stops here.

## Frozen field/action matrix

The following exact rows are transcribed from the baseline authored projections.
Each row inherits the family lifecycle above. `value` means direct replacement;
collection verbs identify existing add/remove actions, not new capabilities.

| Schema suffix | Field | Write intent | Hidden by default | Disposition |
| --- | --- | --- | --- | --- |
| `comm_log.v1` | `comm_log.timestamp_utc` | value | no | Retain |
| `comm_log.v1` | `comm_log.comm_type` | value | no | Retain |
| `comm_log.v1` | `comm_log.audience` | value | no | Retain |
| `comm_log.v1` | `comm_log.channel_or_meeting` | value | no | Retain |
| `comm_log.v1` | `comm_log.summary` | value | no | Retain |
| `comm_log.v1` | `comm_log.next_report_at` | value | no | Retain |
| `comm_log.v1` | `comm_log.privilege_tag` | value | yes | Retain |
| `comm_log.v1` | `comm_log.decision_ids` | add/remove record ref | yes | Retain |
| `comm_log.v1` | `comm_log.action_task_ids` | add/remove record ref | yes | Retain |
| `comm_log.v1` | `comm_log.audience_party_ids` | add/remove party ref | yes | Retain |
| `comm_log.v1` | `comm_log.attendee_party_ids` | add/remove party ref | yes | Retain |
| `decisions.v1` | `decision.summary` | value | no | Retain |
| `decisions.v1` | `decision.status` | value | no | Retain |
| `decisions.v1` | `decision.owner_user_id` | value | no | Retain |
| `decisions.v1` | `decision.decision_type` | value | no | Retain |
| `decisions.v1` | `decision.decided_at` | value | no | Retain |
| `decisions.v1` | `decision.rationale` | value | no | Retain |
| `decisions.v1` | `decision.support_refs` | add/remove record ref | no | Retain |
| `decisions.v1` | `decision.affected_record_ids` | add/remove record ref | yes | Retain |
| `evidence.v1` | `evidence.title` | value | no | Retain |
| `evidence.v1` | `evidence.lifecycle_state` | value | no | Retain |
| `evidence.v1` | `evidence.requested_at` | value | no | Retain |
| `evidence.v1` | `evidence.received_at` | value | no | Retain |
| `evidence.v1` | `evidence.storage_ref` | value | no | Retain |
| `evidence.v1` | `evidence.collector_party_text` | value | no | Retain |
| `evidence.v1` | `evidence.collector_party_id` | value | yes | Retain |
| `evidence.v1` | `evidence.source_party_text` | value | no | Retain |
| `evidence.v1` | `evidence.source_party_id` | value | yes | Retain |
| `findings.v1` | `finding.statement` | value | no | Retain |
| `findings.v1` | `finding.kind` | value | no | Retain |
| `findings.v1` | `finding.state` | value | no | Retain |
| `findings.v1` | `finding.owner_user_id` | value | no | Retain |
| `findings.v1` | `finding.confidence_score` | value | no | Retain |
| `findings.v1` | `finding.supporting_refs` | add/remove record ref | yes | Retain |
| `findings.v1` | `finding.contradictory_refs` | add/remove record ref | yes | Retain |
| `forensic_keywords.v1` | `forensic_keyword.pattern` | value | no | Retain |
| `forensic_keywords.v1` | `forensic_keyword.reason` | value | no | Retain |
| `forensic_keywords.v1` | `forensic_keyword.match_mode` | value | no | Retain |
| `forensic_keywords.v1` | `forensic_keyword.case_sensitive` | value | no | Retain |
| `handoff.v1` | `handoff.timestamp_utc` | value | no | Retain |
| `handoff.v1` | `handoff.outgoing_owner_user_id` | value | no | Retain |
| `handoff.v1` | `handoff.incoming_owner_user_id` | value | no | Retain |
| `handoff.v1` | `handoff.current_state_summary` | value | no | Retain |
| `handoff.v1` | `handoff.open_task_ids` | add/remove record ref | yes | Retain |
| `handoff.v1` | `handoff.open_decision_ids` | add/remove record ref | yes | Retain |
| `handoff.v1` | `handoff.open_risk_refs` | add/remove risk ref | yes | Retain |
| `handoff.v1` | `handoff.next_checks` | value | no | Retain |
| `handoff.v1` | `handoff.acknowledged_at` | value | no | Retain |
| `hosts.v1` | `host.display_name` | value | no | Retain |
| `hosts.v1` | `host.hostname` | value | no | Retain |
| `hosts.v1` | `host.aliases` | add/remove alias | no | Retain |
| `hosts.v1` | `host.location` | value | no | Retain |
| `hosts.v1` | `host.os_platform` | value | no | Retain |
| `hosts.v1` | `host.business_owner` | value | no | Retain |
| `hosts.v1` | `host.criticality` | value | no | Retain |
| `hosts.v1` | `host.containment_status` | value | no | Retain |
| `identities.v1` | `identity.display_name` | value | no | Retain |
| `identities.v1` | `identity.upn` | value | no | Retain |
| `identities.v1` | `identity.email` | value | no | Retain |
| `identities.v1` | `identity.sam_account_name` | value | no | Retain |
| `identities.v1` | `identity.aliases` | add/remove alias | no | Retain |
| `identities.v1` | `identity.privilege_level` | value | no | Retain |
| `identities.v1` | `identity.mfa_state` | value | no | Retain |
| `identities.v1` | `identity.reset_status` | value | no | Retain |
| `investigative_queries.v1` | `investigative_query.platform` | value | no | Retain |
| `investigative_queries.v1` | `investigative_query.purpose` | value | no | Retain |
| `investigative_queries.v1` | `investigative_query.query_text` | value | no | Retain |
| `lesson.v1` | `lesson.timestamp_utc` | value | no | Retain |
| `lesson.v1` | `lesson.summary` | value | no | Retain |
| `lesson.v1` | `lesson.owner_user_id` | value | no | Retain |
| `lesson.v1` | `lesson.closure_state` | value | no | Retain |
| `lesson.v1` | `lesson.follow_up_task_ids` | add/remove record ref | no | Retain |
| `lesson.v1` | `lesson.evidence_refs` | add/remove record ref | no | Retain |
| `notes.v1` | `note.title` | value | no | Retain |
| `notes.v1` | `note.body` | value | no | Retain |
| `notes.v1` | `note.tags` | add/remove tag | no | Retain |
| `parties.v1` | `party.display_name` | value | no | Retain |
| `parties.v1` | `party.party_kind` | value | no | Retain |
| `parties.v1` | `party.organization_name` | value | no | Retain |
| `parties.v1` | `party.role_title` | value | no | Retain |
| `parties.v1` | `party.primary_email` | value | no | Retain |
| `parties.v1` | `party.timezone_name` | value | no | Retain |
| `parties.v1` | `party.external_ref` | value | no | Retain |
| `parties.v1` | `party.notes` | value | yes | Retain |
| `status_review.v1` | `status_review.timestamp_utc` | value | no | Retain |
| `status_review.v1` | `status_review.review_owner_user_id` | value | no | Retain |
| `status_review.v1` | `status_review.current_state_summary` | value | no | Retain |
| `status_review.v1` | `status_review.blocked_task_ids` | add/remove record ref | yes | Retain |
| `status_review.v1` | `status_review.pending_evidence_ids` | add/remove record ref | yes | Retain |
| `status_review.v1` | `status_review.open_decision_ids` | add/remove record ref | yes | Retain |
| `status_review.v1` | `status_review.active_risks_summary` | value | no | Retain |
| `status_review.v1` | `status_review.next_report_at` | value | no | Retain |
| `task_requests.v1` | `task.title` | value | no | Retain |
| `task_requests.v1` | `task.status` | value | no | Retain |
| `task_requests.v1` | `task.owner_user_id` | value | no | Retain |
| `task_requests.v1` | `task.priority` | value | no | Retain |
| `task_requests.v1` | `task.task_kind` | value | no | Retain |
| `task_requests.v1` | `task.workstream` | value | no | Retain |
| `task_requests.v1` | `task.due_at` | value | no | Retain |
| `task_requests.v1` | `task.requester_party_text` | value | no | Retain |
| `task_requests.v1` | `task.requester_party_id` | value | yes | Retain |
| `task_requests.v1` | `task.blocked_reason` | value | no | Retain |
| `task_requests.v1` | `task.completed_at` | value | no | Retain |
| `task_requests.v1` | `task.external_ticket_ref` | value | no | Retain |
| `task_requests.v1` | `task.closure_summary` | value | yes | Retain |
| `task_requests.v1` | `task.linked_record_ids` | add/remove record ref | yes | Retain |
| `task_requests.v1` | `task.decision_record_id` | value | yes | Retain |
