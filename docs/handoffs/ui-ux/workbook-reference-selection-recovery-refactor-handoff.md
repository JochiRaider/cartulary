# Workbook reference selection and recovery

## Baseline and execution control

Authorized scope is ordinary existing-record direct references in grids and
inspectors, and declared inspector reference collection additions/removals.
Execution began on clean `main` at
`540fbb03310d3fb632688c861107ba87d91b5621`. Root AGENTS.md is the only applicable
repository instruction file. There were no pre-existing changes. Planning
inspection followed the localized digest read order; the digest and research
material remain advisory. No commit, push, deployment, analyst-data modification,
new dependency, browser persistence, or main-sheet query-window redesign is in
scope.

This handoff is the sole execution tracker. Save only the current workstream as
IN_PROGRESS before beginning it, append actual paths, decisions, commands,
results, risks and next action at its exit, then save DONE before advancing.
An applicable blocked dependency prevents advancement and completion.

| Workstream | Status | Exit |
| --- | --- | --- |
| RSR-01 Field and target ownership | DONE | Owner-backed field matrix and characterized gaps. |
| RSR-02 Discovery and selection ownership | DONE | Bounded paged reads, typed outcomes and retained identity. |
| RSR-03 Editing integration and retirement | DONE | Parent editing owners remain authoritative; consumers accounted for. |
| RSR-04 Integrated evidence | DONE | Field/lifecycle and production accessibility evidence passes. |
| RSR-05 Validation and completion | DONE | Terminal checks and complete handoff pass. |

## Authority and boundaries

Behavior belongs to Core 01 §§3.3.4–3.3.7, §§6–7, §18B and §19; Core 02's
record identity, Party, coordination, artifact, relationship and history owners;
Core 03 §§2.3A,3–4,13 and REQ-03-298/299/100; and Core 04 §§1–2.
Domain vocabulary and design §§7–8,12,14 apply within their own boundaries.
The completed query, grid/autosave, inspector, relationship, creation and
revocation handoffs are historical evidence, not fresh test passes.

Source ownership is `web.workbook` for the affected workbook files. Verification
routing is independently defined by contracts/verification, tools/test_catalog_owner.json
and tools/test_families. Confirmed entry owners are web.workbook,
module.workbook, package.grid_adapter, package.view_contracts, web.architecture,
module.tasksdecisions and module.artifacts. Additional changed consumers select
verification through their current task guides.

No product test, runtime, generator or release evidence may consume Markdown.
Authored projections and Make-owned generators precede generated output changes.

## Evidence before implementation

- Checkout revalidation: `git status --short --branch`, `git rev-parse HEAD`,
  `git diff --stat`: clean main, exact baseline above.
- Planning focused reference broker / inspector binding / grid intent / reference
  option slice: PASS, 6/6 execution units, run
  `.cartulary/test-results/20260915T171034Z-p12227`; summary
  `target-summaries/test-slice.json`. This proves retained behavior only.
- Planning `make help`, `make help-all`, owner task guides and explain-test-owner
  completed successfully. `git diff --check` passed before implementation.

## Browsing decision

The user selected Previous/Next pages. The picker retains one accepted 100-row
candidate page and at most ten earlier request checkpoints. Selections are
independent and pending collection actions retain the existing 64-action limit.
Previous re-fetches a checkpoint; First resets the chain. Reads use the existing
30-second deadline, explicit retries, instance-local sharing and cancellation.
Query/target replacement and dismissal release candidate browsing state.

## RSR-01 field matrix

All direct fields below are available in grid and ordinary inspector Details;
collection actions are inspector-only. Direct Party IDs additionally have
specialized Party linking controls, which retain their separate owner. Task
owner also appears in the Task lifecycle form with its guarded-field review.

All candidates require current incident read authority. Record references require
same-incident, non-soft-deleted targets; collections exclude self-links. Host and
Identity candidate surfaces expose their live entities, while retained historical
IDs remain subject to authoritative general-record admission. Member submission
requires active current membership, independently of a retained label. Membership
resources do not expose user activity, so discovery cannot claim that a displayed
member guarantees write admission. Field labels such as Open or Blocked do not
add undeclared lifecycle filters. Party text and IDs remain independent.

| Schema | Field | Kind / target | Clear / item identity | Owner / verification |
| --- | --- | --- | --- | --- |
| `cartulary.view.comm_log.v1` | `comm_log.action_task_ids` | collection / task | returned item_ref | Core 01 §19 / Core 02 §10.4; module.artifacts, web.workbook, module.workbook |
| `cartulary.view.comm_log.v1` | `comm_log.attendee_party_ids` | collection / party | returned item_ref | Core 01 §19 / Core 02 §10.4; module.artifacts, web.workbook, module.workbook |
| `cartulary.view.comm_log.v1` | `comm_log.audience_party_ids` | collection / party | returned item_ref | Core 01 §19 / Core 02 §10.4; module.artifacts, web.workbook, module.workbook |
| `cartulary.view.comm_log.v1` | `comm_log.decision_ids` | collection / decision | returned item_ref | Core 01 §19 / Core 02 §10.4; module.artifacts, web.workbook, module.workbook |
| `cartulary.view.decisions.v1` | `decision.affected_record_ids` | collection / record | returned item_ref | Core 01 §§7.4.8–9 / REQ-01-333–334; module.tasksdecisions, web.workbook, module.workbook |
| `cartulary.view.decisions.v1` | `decision.owner_user_id` | direct / incident_member | non-clearable | Core 01 §18B, REQ-01-628; module.tasksdecisions, web.workbook, module.workbook |
| `cartulary.view.decisions.v1` | `decision.support_refs` | collection / record | returned item_ref | Core 01 §§7.4.8–9 / REQ-01-333–334; module.tasksdecisions, web.workbook, module.workbook |
| `cartulary.view.evidence.v1` | `evidence.collector_party_id` | direct / party | explicit null | Core 01 §18B, REQ-01-628; module.evidence, web.workbook, module.workbook |
| `cartulary.view.evidence.v1` | `evidence.source_party_id` | direct / party | explicit null | Core 01 §18B, REQ-01-628; module.evidence, web.workbook, module.workbook |
| `cartulary.view.findings.v1` | `finding.contradictory_refs` | collection / record | returned item_ref | Core 01 §19 / Core 02 §10.4; module.artifacts, web.workbook, module.workbook |
| `cartulary.view.findings.v1` | `finding.owner_user_id` | direct / incident_member | non-clearable | Core 01 §18B, REQ-01-628; module.artifacts, web.workbook, module.workbook |
| `cartulary.view.findings.v1` | `finding.supporting_refs` | collection / record | returned item_ref | Core 01 §19 / Core 02 §10.4; module.artifacts, web.workbook, module.workbook |
| `cartulary.view.handoff.v1` | `handoff.incoming_owner_user_id` | direct / incident_member | non-clearable | Core 01 §18B, REQ-01-628; module.artifacts, web.workbook, module.workbook |
| `cartulary.view.handoff.v1` | `handoff.open_decision_ids` | collection / decision | returned item_ref | Core 01 §19 / Core 02 §10.4; module.artifacts, web.workbook, module.workbook |
| `cartulary.view.handoff.v1` | `handoff.open_task_ids` | collection / task | returned item_ref | Core 01 §19 / Core 02 §10.4; module.artifacts, web.workbook, module.workbook |
| `cartulary.view.handoff.v1` | `handoff.outgoing_owner_user_id` | direct / incident_member | non-clearable | Core 01 §18B, REQ-01-628; module.artifacts, web.workbook, module.workbook |
| `cartulary.view.lesson.v1` | `lesson.evidence_refs` | collection / evidence | returned item_ref | Core 01 §19 / Core 02 §10.4; module.artifacts, web.workbook, module.workbook |
| `cartulary.view.lesson.v1` | `lesson.follow_up_task_ids` | collection / task | returned item_ref | Core 01 §19 / Core 02 §10.4; module.artifacts, web.workbook, module.workbook |
| `cartulary.view.lesson.v1` | `lesson.owner_user_id` | direct / incident_member | non-clearable | Core 01 §18B, REQ-01-628; module.artifacts, web.workbook, module.workbook |
| `cartulary.view.status_review.v1` | `status_review.blocked_task_ids` | collection / task | returned item_ref | Core 01 §19 / Core 02 §10.4; module.artifacts, web.workbook, module.workbook |
| `cartulary.view.status_review.v1` | `status_review.open_decision_ids` | collection / decision | returned item_ref | Core 01 §19 / Core 02 §10.4; module.artifacts, web.workbook, module.workbook |
| `cartulary.view.status_review.v1` | `status_review.pending_evidence_ids` | collection / evidence | returned item_ref | Core 01 §19 / Core 02 §10.4; module.artifacts, web.workbook, module.workbook |
| `cartulary.view.status_review.v1` | `status_review.review_owner_user_id` | direct / incident_member | non-clearable | Core 01 §18B, REQ-01-628; module.artifacts, web.workbook, module.workbook |
| `cartulary.view.task_requests.v1` | `task.decision_record_id` | direct / decision | explicit null | Core 01 §18B, REQ-01-628; module.tasksdecisions, web.workbook, module.workbook |
| `cartulary.view.task_requests.v1` | `task.linked_record_ids` | collection / record | returned item_ref | Core 01 §§7.4.8–9 / REQ-01-333–334; module.tasksdecisions, web.workbook, module.workbook |
| `cartulary.view.task_requests.v1` | `task.owner_user_id` | direct / incident_member | non-clearable | Core 01 §18B, REQ-01-628; module.tasksdecisions, web.workbook, module.workbook |
| `cartulary.view.task_requests.v1` | `task.requester_party_id` | direct / party | explicit null | Core 01 §18B, REQ-01-628; module.tasksdecisions, web.workbook, module.workbook |

### Discovery, dependencies and submission

Record domains use the existing view-query port, including server paging and
declared filters. Membership uses listIncidentMemberships, ordered by joined_at
and user_id, with no invented search endpoint. All seventeen implemented record
surfaces are eligible for the unrestricted record domain; the six omitted by the
old inventory are Assessments, Indicators, Communications, Handoffs, Status
Reviews and Lessons. Artifacts without a declared surface have no invented
candidate endpoint; exact-ID authoring and authoritative submission remain valid.

There is no generic record singleton or record_id view filter in the current
public read contract. Selected presentation uses current authorized committed
rows, returned collection display_text and previously selected labels. Failure
to independently resolve a label is not evidence of deletion or eligibility.
No exhaustive lookup scan is introduced.

Grid writes remain with WorkbookGridDraftStore / WorkbookGridAutosave; inspector
writes remain with WorkbookInspectorDraftStore / explicit patch runtime. Task
owner edits retain Task guard-field dependencies; other ordinary fields review
the edited field and existing source-owner dependencies. Related Party text is
not implicitly changed. Collection actions preserve field-scoped item_ref and
owner ordering/limits. Links and source owners validate at dispatch.

Excluded from ordinary selection: mention-backed Timeline Host/Identity fields,
Timeline Evidence association workflows, tags, aliases, child risk references,
append-only Assessment authoring, Indicator workflows and contextual creation.
These are regression boundaries, not missing generic target mappings.

### Gap register

| Gap | Remediation / affected areas | Rationale and benefit | Compatibility / risk | Binary validation |
| --- | --- | --- | --- | --- |
| G1 first-page-only broker | Typed page service; broker/controller/UI/tests | Reach eligible targets with bounded reads | Internal interface migration; cursor correlation risk | Target 101 reachable without eager traversal |
| G2 aggregate failure and erased typing | Per-source typed outcomes and local retry; hook/broker/tests | Recover one source without losing unrelated authorized choices | Existing lifecycle owner remains; stale-publication risk | Failed source/continuation retains usable page and choices; retry writes zero times |
| G3 duplicated incomplete inventories | Versioned owner-backed target projection; contracts/facade/policies/tests | One explicit target decision instead of scattered lists | No HTTP change; accidental widening risk | All 27 bindings and eligible seventeen surfaces match owners; specialized fields excluded |
| G4 empty reference becomes null | Exact direct-reference preparation; inspector/tests | Preserve intentional clearing and raw draft distinctions | Invalid empty values now reject; retained raw input preserved | Empty/whitespace reject, permitted explicit null succeeds |
| G5 popup events captured by parent | Nested grid interaction boundary; adapter/control/browser tests | Picker reads/navigation cannot submit or cancel parent work | Small adapter interface; focus/blur risk | Escape closes one layer; Enter/Tab within picker do not write |
| G6 eager snapshot consumers | Migrate ordinary edits, Task member selector and label/receipt consumers; guides/tests | Remove duplicate state while preserving source-owned operations | No data migration; specialized consumer regression risk | No eager inventories remain; named consumers pass |

### RSR-01 exit — PASS

Implemented the 27-field projection in contracts/view-references/index.json,
registered it in contracts/index.json and generated the Go/TypeScript derivatives
through `make generate`. packages/view-contracts/src/references.ts exposes the
field contract without adding HTTP fields or routes. Inspector preparation now
preserves exact IDs and rejects empty input independently of explicit null.

Focused characterization proves discarded continuation and aggregate failure;
existing broker tests prove delayed-consumer cancellation and authority-instance
replacement. Existing inspector binding tests prove off-page selected-label/ID
retention. These are confirmed implementation observations, not claims that
selected-value retention was absent. Future integrated lifecycle risks remain
RSR-04 acceptance obligations.

Commands and actual results:

- `make generate`: PASS, run
  `.cartulary/test-results/20260915T172314Z-p21731/generate/tool-run-summary.json`.
  Earlier related failures at 172035Z-p16527 (projection output location) and
  172232Z-p18603 (unsorted selector titles) were corrected in authored inputs.
- `make test-slice OWNER=web.workbook ROWS=web.workbook.regression.referencequerybroker_suite_7bd36b209a,web.workbook.regression.inspector_draft_lifetime,web.workbook.regression.inspector_draft_binding`:
  PASS, `.cartulary/test-results/20260915T172315Z-p22332`.
- `make test-slice OWNER=package.view_contracts ROWS=package.view_contracts.frontend_unit.references`:
  PASS, `.cartulary/test-results/20260915T172316Z-p23014`.
  Slice summaries are target-summaries/test-slice.json within each run.

No unresolved owner contradiction blocks discovery. Target availability remains
separate from eligibility; member activity is checked by the mutation owner.
Next action: implement bounded typed discovery and staged selection under RSR-02.

## RSR-02 exit — PASS

Added ports/WorkbookReferenceReadPort.ts, services/workbookReferenceReader.ts,
services/WorkbookReferenceSelection.ts and
adapters/createWorkbookReferenceMemberReader.ts under apps/web/src/workbook.
The reader preserves accepted/rejected/aborted outcomes, canonical metadata and
opaque paging. It shares in-flight work only within its authority instance;
one cancellation releases one consumer and the final cancellation aborts transport.
Disposal and the 30-second deadline fence late publication. No settled inventory
or eager page traversal exists in the new boundary.

The controller retains one page, ten earlier checkpoints and ordered staged
identities independently. Failed continuation retains the accepted page; source
replacement releases browsing state while retaining choices. Authority failure
conceals picker presentation and requests the existing lifecycle owner. Member
paging uses its own route and rejects record filters. Source eligibility is
checked against the field projection; current surface availability is learned
through the authorized source read. A missing surface does not remove target
eligibility or silently retarget an exact-ID choice.

- `make test-slice OWNER=web.workbook ROWS=web.workbook.regression.reference_selection`:
  PASS, 2/2 execution units, eight focused scenarios,
  `.cartulary/test-results/20260915T173356Z-p28387`.
- `make frontend-typecheck`: PASS,
  `.cartulary/test-results/20260915T173441Z-p29657`. Earlier related failure
  at 173357Z-p28677 was a test mock's zero-argument inferred tuple; corrected.
- Source ownership and local ports/services/adapter guides include the new files.

No parent authoring or mutation interface enters the controller. Resource tests
traverse sixteen pages, then Previous/Next/First, with exactly nineteen reads,
100 retained candidates and at most ten earlier checkpoints. Off-page identities
remain stable. Typed failures and membership identity remain distinct.

Remaining integration risks are focus/blur, consumer retirement and lifecycle
wiring; they belong to RSR-03/04. Next action: integrate production editors and
remove the superseded eager state owners.

## RSR-03 exit — PASS

WorkbookReferenceControl now supplies exact-ID input and a native top-layer
picker to WorkbookGridEditorControl and WorkbookInspectorEditControl. Accept
uses the grid's existing commit override; inspector acceptance changes only its
retained draft. Removal remains the native collection-item control and serializes
returned item_ref values. prepareWorkbookInspectorChange projects ordered
reference actions explicitly and enforces the existing 64-action pending limit.
Empty direct input rejects; permitted null remains an explicit Clear action.

Grid Adapter production and DOM-unit capture handlers defer nested popup keys.
Escape dismisses one layer, returning focus to the input; candidate paging,
filtering and retry do not invoke parent writes. Authority failures conceal the
picker, suspend the existing explicit-patch owner (which owns draft readability),
and request the existing authorization lifecycle. Role/closure admission remains
independent of candidate reads.

Retired useOwnerReferenceOptions, workbookReferenceOptions buckets/tests,
referenceQueryBroker/tests, ReferenceRequirement inventories and unused policy
refresh consequences. Removed the inventory-only incident listMembers port.
The neutral typed member reader owns membership paging instead.

Consumer accounting:

- Note creation retains its own source/creation owner. Receipt changes restart
  only an open Note candidate source; unopened pickers read fresh on demand.
- Decision labels read existing committed-row/receipt evidence. No second receipt
  store exists. Raw IDs never imply current authorization or eligibility.
- Ordinary creation keeps its authoring reader and retained selected labels;
  initial snapshot fallback now uses exact IDs honestly.
- Task lifecycle uses the incident-member picker while retaining Task validation,
  dependencies and its existing Update interaction.
- Timeline Host/Identity observation was a real additional broker consumer,
  reproduced during migration. It now uses the existing query port for its named
  bounded observation, independently of authored Entity queries. Mention
  resolution retains its specialized candidate owner.
- Indicator lifecycle retains its own IndicatorSupportReference value and
  specialized picker. genericReferenceOptionsFromRows remains solely a neutral
  row-label projection for that supported consumer. Assessment, Party,
  contextual creation and other specialized candidate helpers keep their
  separate contracts; empty aggregate props were removed.

Actual validation:

- Workbook integration/inspector/adapters/ordinary creation/model slice: PASS,
  8/8 units, `.cartulary/test-results/20260915T175235Z-p49056`.
- Entity observation owner slice: PASS, 2/2,
  `.cartulary/test-results/20260915T175133Z-p42320`.
- Task lifecycle owner slice: PASS, 2/2,
  `.cartulary/test-results/20260915T175236Z-p49516`.
- Grid Adapter nested interaction slice: PASS, 2/2,
  `.cartulary/test-results/20260915T174916Z-p35803`.
- Final changed reference-control/model/binding slice: PASS, 4/4,
  `.cartulary/test-results/20260915T175446Z-p56546`.
- `make format`: PASS, 175435Z-p52177; `make frontend-typecheck`: PASS,
  175446Z-p56708; `make lint-biome`: PASS, 175446Z-p56713.
  These run roots all reside under .cartulary/test-results.

Related intermediate failures were corrected: missing membership fixture
request_id (174915Z-p35555), a policy-removal syntax remainder
(175132Z-p42102/175134Z-p42709), and new hook-dependency/accessibility/test typing
issues found by format/typecheck. A diagnostic-only BIOME_CHECK_FLAGS override
was rejected by the harness as an undeclared input; standard public targets
were used successfully. No unrelated source formatting was retained.

Next action: RSR-04 production-route evidence including >100 real targets,
failures, selected identity, authority recovery and keyboard/layout behavior.

### RSR-04 merged-target characterization correction

The production target-lifecycle test at
`.cartulary/test-results/20260915T182303Z-p44393` disproved its hypothesis that
absence from Host candidates requires reference rejection. A merged Host remains
a first-class historical record (Core 02 REQ-02-066). Core 01's general record
collection contract rejects foreign, wrong-type and soft-deleted targets; it
does not declare an additional merged-state restriction. Core 02 REQ-02-220
owns repointing the live graph at merge, not a frontend inference from later
candidate absence. Current Links admission accepts the exact historical ID.

The seam's binary requirement is no silent retargeting: the test now checks the
unchanged submitted ID and returned `record_ref:<loser-id>` item_ref. Deleted Parties
and removed members still require actual owner rejection. A proposed extra
Links rejection was withdrawn before validation: `merged_into` is a permitted
relationship kind, not a guaranteed persisted merge-lineage record. No backend,
merge, HTTP or relationship contract change is retained. This is a characterized
non-gap for selection ownership, not evidence that labels confer eligibility.
General-record target admission stays with its existing owner; no page scan or
frontend lifecycle filter is introduced. No blocking dependency remains from
this disproved blanket-rejection hypothesis.

### Field matrix contract details

These rules complete each row of the field matrix; they are shared owner rules,
not inferred from labels or local surface inventories.

| Target domain | Candidate surface / exact add identity | Collection action identity |
| --- | --- | --- |
| Party | `cartulary.view.parties.v1`; exact party_id | `add_party_ref {party_id}` / `remove_party_ref {item_ref}` |
| Decision | `cartulary.view.decisions.v1`; exact record_id | `add_record_ref {linked_record_id}` / `remove_record_ref {item_ref}` |
| Task | `cartulary.view.task_requests.v1`; exact record_id | Same record-reference actions |
| Evidence | `cartulary.view.evidence.v1`; exact record_id | Same record-reference actions |
| General record | Ten first-class types and seventeen declared surfaces in the projection; exact record_id | Same record-reference actions |
| Member | `listIncidentMemberships`; exact user_id | Direct scalar only, no collection-removal identity |

Direct changes use `{field_key, value: exact_id}`. Clearable fields accept
explicit `value: null`; untouched fields are omitted. Empty text is invalid,
not a clear request. Collection changes use the declared field_key and
`collection_actions_v1`, preserving action order and the existing 64-action
limit. Returned item_ref values are passed through, not reconstructed from a
record selection. All source schema/action bindings live in the typed
projection and are checked against the existing field contracts.

Direct grid submission remains WorkbookGridAutosave; ordinary inspector
submission remains explicit patch/WorkbookInspectorDraftStore. Task guard
siblings remain with taskGuardFields and taskExplicitPatchContribution; other
fields use their owner-declared dependencies. No reference load or presentation
update changes those baselines. Specialized Party linking retains its own
captured commands and recovery. Task lifecycle's member picker borrows the
new membership capability and retains its explicit Update and Task guards.

### RSR-04 evidence and corrections

Production evidence uses actual authenticated routes and isolated harness data.
Fault injection interrupts those real reads/writes; it does not substitute a
mocked editor for production controls. Model/adapter evidence covers malformed
metadata, correlation, shared consumers and resource bounds that cannot be
reliably inferred from screenshots.

| Acceptance area | Evidence / disposition |
| --- | --- |
| All 27 fields and exact field/action ownership | PASS: view-contracts reference matrix, inspector direct intent matrix, reference-control collection payload matrix; every declared field tested |
| More than 100 targets and cross-page staging | PASS: 105 real Parties in grid, 105 real Notes plus Indicator in inspector; staged off-page identities retained |
| Typed initial/continuation/source failures, empty results, retry | PASS: reader/controller, member adapter and real inspector read interruption; failed continuation retains the accepted page |
| Explicit server filter and query replacement | PASS: rapid replacement model and final production Created By filtering run |
| Cancellation, delayed responses, independent consumers | PASS: shared-reader cancellation/disposal/deadline tests; same-account browser gate releases obsolete response after concealment |
| Bounded retention and requests | PASS: 16 pages plus Previous/Next/First uses 19 reads, one page and at most ten checkpoints; selection limit remains pending actions only |
| Direct ID/null and collection removal | PASS: direct grid invalid-empty correction, exact accepted Party ID, explicit null, real returned item_ref removal; every collection action contract tested |
| Retained parent authoring, detachment and resumption | PASS: grid and inspector owner suites and real retained editing/exact replay/acknowledged-refresh rows |
| Current eligibility and dependency review | PASS: deleted Party and removed member reject while retaining choice; Task guard and all-surface editor owner regressions; merged historical target characterized separately above |
| Closure, role change and incident access loss | PASS: production reference admission lifecycle plus existing grid closure/role rows; readable copyable drafts and write admission remain independent |
| Same-account recovery and replacement account | PASS: production reference draft restored only to the same account; late picker read cannot reopen UI; replacement retires prior authoring |
| Acknowledged write followed by failed refresh | PASS: inspector reference retry/browse submits zero additional writes; existing parent refresh recovery is required before the next mutation |
| Keyboard, names, feedback, focus and clipping | PASS: real reference control at 1280px, 390px, 200% zoom, text spacing and reduced motion, with native top-layer popup, named controls, announced status/error and one-layer Escape |
| Timeline spreadsheet behavior | PASS: pointer, keyboard Enter/Shift+Enter/Tab/Shift+Tab and shell exit, caret, ranges, paste/fill, first input, refresh/rejection and virtualization browser rows; Grid Adapter 48/48 units |
| Specialized consumer regression | PASS: Task lifecycle, Entity observation, ordinary/contextual creation, Note and Decision recovery evidence; Indicator/Assessment/Party owner tests in the 265-unit workbook slice |

Material corrections found by integrated evidence:

- A pending grid reference input briefly became disabled after invalid input;
  removing that redundant pending gate preserves correction focus and the
  existing captured-write owner.
- Popup coordinates now account for inherited CSS zoom. Native top-layer
  placement stays within the visual viewport; tall content scrolls inside it.
- Nested picker text selection is excluded from the parent grid caret capture,
  in addition to its navigation/dismissal keys.
- Selected presentation reads current committed evidence, including Decision
  receipts, even while staging. It does not alter selected identity or parent
  authoring. A dedicated control test verifies this separation.
- Default current-actor creation labels now use already-authorized session
  presentation plus the exact user ID. Other unresolved IDs remain literal.
  No member inventory or receipt cache was restored.
- Inspector continuation/acknowledgement tests use the existing explicit parent
  refresh action before starting another write; reference retry itself stays
  read-only. Test selector ambiguity and exact public merge-path/item_ref
  expectations were corrected without changing product admission.

### RSR-04 exit — PASS

The final production filtering/grid/accessibility slice passes 15/15 execution
units at `.cartulary/test-results/20260915T183805Z-p21546`. Its exact command:

```bash
make service-backed-test-slice OWNER=module.workbook ROWS=module.workbook.browser.reference_selection_recovery,module.workbook.browser.reference_grid_commit,module.workbook.accessibility.reference_picker
```

The Notes filter uses the declared Created By field. Title is not filterable;
the UI correctly excludes it. Typing a filter causes no read until Apply;
empty and restored server results preserve all staged identities. This closes
the pending filtering row above. Candidate read count remains at most ten in
the production scenario including deliberate failures and explicit retries.

Other current evidence (run roots under `.cartulary/test-results/`):

| Command / selected scope | Result and run |
| --- | --- |
| `make test-slice OWNER=web.workbook` | PASS 265/265, 20260915T180534Z-p9935 |
| Changed reader/control/ordinary-control owner slice | PASS 4/4, 20260915T183255Z-p25559; includes rapid query replacement and committed presentation |
| Member adapter/control/reader slice | PASS 4/4, 20260915T182126Z-p2113 |
| `make test-slice OWNER=package.grid_adapter` | PASS 48/48, 20260915T181532Z-p13785 |
| Grid Adapter index and nested-interaction rows after caret fencing | PASS 3/3, 20260915T183806Z-p21785 |
| View-contracts owner slice, then exact facade row correction | Other rows PASS at 20260915T181532Z-p13828; corrected facade PASS 2/2 at 20260915T181718Z-p91048 |
| Architecture owner slice, then selector-boundary correction | Other rows PASS at 20260915T181532Z-p13857; corrected selector PASS 2/2 at 20260915T181719Z-p91670 |
| `make test-slice OWNER=package.protocol_ts` | PASS 8/8, 20260915T183131Z-p97317 |
| Workbook existing grid/inspector recovery, accessibility and inspector visual rows | PASS 19/19, 20260915T181635Z-p58783 |
| Timeline keyboard/pointer/create/bulk/virtualization/refresh/rejection browser rows | PASS 11/11, 20260915T181717Z-p90355 |
| Reference closure/role/access, same-account/account replacement, historical merge, and creation/Decision visual rows | Individual rows PASS at 20260915T183336Z-p28897; its one failed filter test is superseded by the final PASS above |
| Contextual creation, Note visual and ordinary/contextual accessibility rows | PASS rows at 20260915T182018Z-p35011; creation/Decision visual differences subsequently resolved without golden changes |
| Workbook real coordination/generic collection and direct-decoder slice | PASS 3/3, 20260915T182127Z-p3874 |
| Artifact collection mutation matrix | PASS 3/3, 20260915T182128Z-p5518 |
| Evidence direct-reference and lifecycle/failure policy slice | PASS 2/2, 20260915T182129Z-p6150 |
| Task/Decision lifecycle and direct-reference owner slice | PASS 6/6, 20260915T183339Z-p29367 |

Each run retains `run-summary.json`, `target-summaries/<target>.json`, row results
and unit logs. Browser group directories retain Playwright reports, traces,
accessibility trees and production screenshots. Visual review examined the
normal/narrow/200% reference popup; focus is visible, long IDs wrap in the
selection summary, native options retain full accessible names, and tall popup
content scrolls within the viewport. The popup remains in editor DOM ownership
while occupying the native top layer. No golden was changed.

Intermediate failures are resolved or isolated:

- New test selectors initially matched creation and inspector controls together;
  selectors now scope the existing inspector panel. The architecture policy
  caught an unowned test-only ID; tests use existing public selector builders.
- Initial generic facade expectations omitted the two intended exports; the
  exact facade contract was updated and passed.
- The new merge fixture first used the wrong route parameter, then assumed
  merged IDs reject, then used a record-type prefix for item_ref. The corrected
  owner-backed test checks `survivor_record_id`, unchanged historical identity
  and the returned record-reference item identity; product removal never
  constructs this string.
- The first filter fixture incorrectly asked for undeclared Note-title search
  and timed out at 183336Z-p28897. Declared Created By filtering now passes.
- A broader `make browser-e2e-visual-update` attempt at 182317Z-p71131 failed
  before promotion on the unrelated saved-view grouping fixture. Committed
  golden bytes and their manifest remained unchanged. The exact saved-view
  visual row passed independently at 183337Z-p29131 (11/11 units), so this is
  classified as transient fixture failure, not an established baseline defect.
  The earlier three label-only differences were resolved by existing session
  presentation; no golden refresh remains necessary.
- The combined run at 183129Z-p96631 stopped at build-web artifact preparation
  while inputs were being finalized; no browser scenario ran. Fresh production
  builds and the final affected rows pass. Incorrect guessed slice row IDs
  were rejected by the harness and replaced with catalog-owned identifiers.

All applicable field/lifecycle acceptance areas now have passing evidence. No
owner contradiction or blocking implementation dependency remains. Remaining
limits are declared source filters and the absence of a generic targeted read;
selected IDs remain honest and independent. Next action: RSR-05 finalization,
terminal verification and completed handoff.


## RSR-05 terminal validation

Finalization and verification use the current public Make surface and authored
owner catalog. `make help`, `make help-all` and module-author task guides selected
web.workbook, module.workbook and package.grid_adapter first. The changed
contract facade, imports and consumer routes additionally selected
package.view_contracts, package.protocol_ts, web.architecture, module.artifacts,
module.evidence and module.tasksdecisions. The generated-family shape check
selected harness.generated_artifacts through its current task guide. These are
verification routes; they do not replace the behavioral owners above.

`make agent-finalize` passed before the broader terminal owner/type/lint checks.
RESULTS_DIR was unset: retained-run maintenance, canonical retained-evidence
validation and scheduler/performance checks were explicitly skipped. The
finalizer's `unit-artifacts/finalize-summary.json` records those dispositions.
Its schema/catalog validation and Make-owned generation/drift transaction passed.

| Exact terminal command | Result | Run under `.cartulary/test-results/` |
| --- | --- | --- |
| `make generate` | PASS | 20260915T184218Z-p59730; generate/tool-run-summary.json |
| `make agent-finalize` | PASS 1/1 | 20260915T184242Z-p62841 |
| `make test-slice OWNER=web.workbook` | PASS 265/265 | 20260915T184322Z-p66795 |
| `make test-slice OWNER=package.view_contracts` | PASS 6/6 | 20260915T184322Z-p66810 |
| `make test-slice OWNER=web.architecture` | PASS 12/12 | 20260915T184322Z-p66846 |
| `make test-slice OWNER=harness.generated_artifacts` | PASS 5/5 | 20260915T184341Z-p75282 |
| `make lint-scripts` | PASS 2/2 | 20260915T184342Z-p75787 |
| `make frontend-typecheck` | PASS 2/2 | 20260915T184704Z-p7105 |
| `make test-slice OWNER=web.workbook ROWS=web.workbook.regression.reference_selection` | PASS 2/2 after test type annotation | 20260915T184705Z-p7311 |
| `make generate-drift` | PASS 4/4 | 20260915T184706Z-p7544 |
| `make lint-biome` | PASS 2/2 after test annotation | 20260915T184943Z-p13085 |
| `make generated-artifact-policy-check` | PASS 3/3 | 20260915T184944Z-p13210 |
| `make lint-markdown` | PASS, zero failures; repeated on completed handoff | 20260915T185011Z-p14144 and 20260915T185114Z-p15990; adhoc/lint-markdown/tool-run-summary.json |
| `git diff --check` | PASS | Terminal output; no whitespace errors |

Graph run roots contain `run-summary.json`, `run-manifest.json` with the exact
selected owner/rows, `target-summaries/<target>.json`, row results and unit logs.
The RSR-04 table retains narrow backend, Grid Adapter, browser, accessibility
and visual evidence. They were not repeated after the final test-only type
annotation. No production change followed the final RSR-04 passing browser run.

Finalization found one additional authored projection requirement:
check-json-shapes.mjs enumerates active contract families in output order.
Adding view-references there and regenerating the topology resolved the initial
184021Z-p57172 / 184058Z-p57760 failures. The intermediate 184146Z-p58578 /
184203Z-p59219 failures reported the then-stale generated renderer digest;
Make regeneration resolved it. No generator, runtime or verification logic
reads this handoff or other Markdown.

The first terminal typecheck at 184322Z-p67098 caught a test fixture whose
operator literal widened to string. Giving the fixture its WorkbookQueryState
type resolved it; the affected test and typecheck pass above. These failures
were related and corrected. No unexplained applicable failure remains.

### Final scope and owner decisions

The seam covers all 27 projected ordinary direct/collection reference fields,
plus the named membership and presentation consumers required to retire the
old aggregate inventory. Source schema files are
`contracts/view-schemas/<schema-id>.json` for each exact schema ID in the matrix.
Those existing field contracts continue to own write kind, clearability and
direct-reference contract IDs; the new projection owns their downstream target
mapping. No schema HTTP fields were added.

Substantive corrections are the complete owner-permitted target set, explicit
null versus invalid empty input, independently recoverable paged reads, exact
collection item identity, nested keyboard/caret handling and zoom placement.
The historical merged-target disposition is documented above: a retained exact
ID does not silently become the survivor. Current source owners decide admission.

The G1-G6 gap register is closed by passing evidence. Its implementation/test
paths appear in the inventory below. Core 01 query/continuation and mutation
clauses drive G1/G2/G4; field/relationship owners drive G3/G6; Core 03 keyboard,
draft and authority lifetimes drive G2/G5/G6; Core 04 governs concealment in all
read recovery. Source guides for ports, services, components, query and affected
specialized consumers document the new ownership. No adopted owner specification
needed amendment. Long-term benefits are bounded browsing, one target decision,
separate read/write recovery and fewer retained state owners.

### Compatibility, limits and skipped work

Existing HTTP routes and payload shapes remain. There is no dependency,
persistent browser storage, record migration or analyst-data change. Empty
reference text now rejects locally; clearable fields still require explicit
null. General record choices now reach all seventeen eligible implemented
surfaces. Party IDs, member user IDs and removal item_ref values remain distinct.

The retired files and supported retained consumers are listed in RSR-03. Native
reference controls use the browser top layer; production Chromium route tests
cover keyboard, narrow layout, zoom, text spacing and focus. Other browser
engines were not exercised in this run. A missing generic singleton read or
undeclared filter cannot be manufactured by this seam: unresolved selections
show a retained label or exact ID and await authoritative mutation admission.
A candidate page, cached label or current actor label grants no authority.

- Full release/CI, global browser and security audit suites were not run. The
  affected owner, source-boundary, authorization-recovery, browser, accessibility
  and visual slices cover this change; server authorization policy, deployment,
  dependencies and release claims are outside scope.
- Broad measurement and main query-window checks are N/A to changed behavior:
  that window is unchanged, and picker resource bounds have focused executable
  evidence. Existing Timeline navigation/virtualization regression rows passed.
- Retained-run maintenance is SKIPPED because RESULTS_DIR was unset and no
  qualifying full warm-check root was supplied.
- Visual golden promotion is N/A: no golden or manifest bytes changed. The
  failed maintenance attempt and successful affected rendering evidence are
  recorded in RSR-04. No unrelated generated change remains.
- No commit, push, deployment or advisory digest edit was performed.

### Rollback and next action

The captured starting checkout was clean main at
540fbb03310d3fb632688c861107ba87d91b5621 and remains on that branch/HEAD.
Rollback should restore only the modified/deleted paths below to that captured
HEAD and remove only the listed new files, after checking for any newer user
changes. Revert authored projection/routing inputs together, then use their
Make-owned generators. Do not use a broad hard reset or clean. Local run evidence
may be retained independently. There was no pre-existing work to overwrite;
any work added after this handoff must be preserved.

No blocking owner or implementation risk remains. The practical limits are
source-declared filters, historical/unresolved presentation and the browser
coverage stated above. Recommended next action is review of this working-tree
diff and the retained production evidence; merging or releasing remains outside
this task.

### Final changed-file inventory

The inventory includes all additions, edits and deletions; generated derivatives
are included only as Make outputs. Paths under the workbook include small
consumer-prop removals as well as the new boundary. No backend domain source,
lockfile, visual golden or advisory digest is changed.

Total: 97 paths. A = added, M = modified, D = deleted.

- M: `apps/web/e2e/workbook-grid-autosave.spec.ts`
- M: `apps/web/e2e/workbook-inspector-edit.spec.ts`
- M: `apps/web/src/workbook/WorkbookShell.tsx`
- M: `apps/web/src/workbook/adapters/README.md`
- M: `apps/web/src/workbook/adapters/createWorkbookIncidentAdapter.ts`
- A: `apps/web/src/workbook/adapters/createWorkbookReferenceMemberReader.ts`
- M: `apps/web/src/workbook/adapters/createWorkbookStartupAndIncidentAdapters.test.ts`
- M: `apps/web/src/workbook/components/EntityWorkbookSurface.tsx`
- M: `apps/web/src/workbook/components/GenericMutationControl.test.tsx`
- M: `apps/web/src/workbook/components/GenericMutationControl.tsx`
- M: `apps/web/src/workbook/components/GenericWorkbookSurface.tsx`
- M: `apps/web/src/workbook/components/README.md`
- M: `apps/web/src/workbook/components/WorkbookGridEditorControl.tsx`
- A: `apps/web/src/workbook/components/WorkbookReferenceControl.test.tsx`
- A: `apps/web/src/workbook/components/WorkbookReferenceControl.tsx`
- M: `apps/web/src/workbook/components/genericMutationControlModel.ts`
- M: `apps/web/src/workbook/features/assessments/AssessmentWorkbookInspector.tsx`
- M: `apps/web/src/workbook/features/assessments/useAssessmentWorkbookInspectorComposition.tsx`
- M: `apps/web/src/workbook/features/coordination/ContextualCreateForm.tsx`
- M: `apps/web/src/workbook/features/coordination/CoordinationWorkflowBindings.tsx`
- M: `apps/web/src/workbook/features/coordination/README.md`
- M: `apps/web/src/workbook/features/coordination/useCoordinationWorkflowController.test.tsx`
- M: `apps/web/src/workbook/features/entities/EntityWorkbookInspector.tsx`
- M: `apps/web/src/workbook/features/entities/useEntityWorkbookInspectorComposition.tsx`
- M: `apps/web/src/workbook/features/evidence/TimelineRelatedEvidenceForm.tsx`
- M: `apps/web/src/workbook/features/generic/GenericWorkbookInspector.tsx`
- M: `apps/web/src/workbook/features/generic/GenericWorkbookInspectorPresentation.tsx`
- M: `apps/web/src/workbook/features/generic/useGenericWorkbookInspectorComposition.tsx`
- M: `apps/web/src/workbook/features/indicators/IndicatorCanonicalAuthoring.tsx`
- M: `apps/web/src/workbook/features/indicators/IndicatorLifecycleSupportPicker.tsx`
- M: `apps/web/src/workbook/features/indicators/README.md`
- M: `apps/web/src/workbook/features/indicators/indicatorLifecycle.characterization.test.tsx`
- M: `apps/web/src/workbook/features/indicators/indicatorLifecycleModel.ts`
- M: `apps/web/src/workbook/features/ordinary/OrdinaryCreateControl.tsx`
- M: `apps/web/src/workbook/features/ordinary/README.md`
- M: `apps/web/src/workbook/features/ordinary/ordinaryCreateControls.test.tsx`
- M: `apps/web/src/workbook/features/parties/PartyLinkControls.tsx`
- M: `apps/web/src/workbook/hooks/README.md`
- D: `apps/web/src/workbook/hooks/useOwnerReferenceOptions.ts`
- M: `apps/web/src/workbook/hooks/useWorkbookShellInfrastructure.ts`
- M: `apps/web/src/workbook/hooks/useWorkbookSurfaceQueries.ts`
- M: `apps/web/src/workbook/inspector/InspectorCreateRelatedWorkflow.tsx`
- M: `apps/web/src/workbook/inspector/WorkbookInspectorDraftStore.test.ts`
- M: `apps/web/src/workbook/inspector/WorkbookInspectorEditControl.tsx`
- M: `apps/web/src/workbook/inspector/prepareWorkbookInspectorChange.ts`
- M: `apps/web/src/workbook/inspector/useWorkbookInspectorEditDraft.test.tsx`
- M: `apps/web/src/workbook/models/README.md`
- D: `apps/web/src/workbook/models/workbookReferenceOptions.test.ts`
- D: `apps/web/src/workbook/models/workbookReferenceOptions.ts`
- M: `apps/web/src/workbook/models/workbookSurfaceRegistration.ts`
- M: `apps/web/src/workbook/policies/artifactSurfacePolicies.ts`
- M: `apps/web/src/workbook/policies/coordinationSurfacePolicies.ts`
- M: `apps/web/src/workbook/policies/evidenceSurfacePolicies.ts`
- M: `apps/web/src/workbook/policies/workbookSurfacePolicy.ts`
- M: `apps/web/src/workbook/ports/README.md`
- M: `apps/web/src/workbook/ports/WorkbookIncidentPort.ts`
- A: `apps/web/src/workbook/ports/WorkbookReferenceReadPort.ts`
- M: `apps/web/src/workbook/query/README.md`
- M: `apps/web/src/workbook/query/useEntitySurfaceQuery.test.tsx`
- M: `apps/web/src/workbook/query/useEntitySurfaceQuery.ts`
- M: `apps/web/src/workbook/services/README.md`
- A: `apps/web/src/workbook/services/WorkbookReferenceSelection.ts`
- D: `apps/web/src/workbook/services/referenceQueryBroker.test.ts`
- D: `apps/web/src/workbook/services/referenceQueryBroker.ts`
- A: `apps/web/src/workbook/services/workbookReferenceReader.ts`
- A: `apps/web/src/workbook/services/workbookReferenceSelection.test.ts`
- M: `apps/web/src/workbook/surfaces/WorkbookSurfacesFacade.tsx`
- M: `apps/web/src/workbook/timeline/components/TimelineMentionActionControls.tsx`
- M: `apps/web/src/workbook/timeline/components/TimelineWorkbookInspectorSections.tsx`
- M: `apps/web/src/workbook/timeline/composition/useTimelineInspectorWorkflowComposition.ts`
- M: `apps/web/src/workbook/timeline/presentation/useTimelineWorkbookPresentation.tsx`
- M: `contracts/index.json`
- M: `contracts/protocol-ts/frontend-entrypoints.v2.json`
- A: `contracts/view-references/index.json`
- A: `docs/handoffs/ui-ux/workbook-reference-selection-recovery-refactor-handoff.md`
- A: `internal/gen/contractviewreferences/artifacts_gen.go`
- M: `packages/grid-adapter/README.md`
- M: `packages/grid-adapter/src/index.test.tsx`
- M: `packages/grid-adapter/src/rdgCompiler.tsx`
- M: `packages/grid-adapter/src/test-support.tsx`
- M: `packages/protocol-ts/src/entrypoints/view-schemas.ts`
- A: `packages/protocol-ts/src/generated/view-reference-registry.ts`
- A: `packages/view-contracts/README.md`
- M: `packages/view-contracts/src/facade.test.ts`
- M: `packages/view-contracts/src/index.ts`
- M: `packages/view-contracts/src/projection.ts`
- A: `packages/view-contracts/src/references.test.ts`
- A: `packages/view-contracts/src/references.ts`
- M: `tools/browser_e2e_batch_manifest.json`
- M: `tools/execution_topology_render_index.json`
- M: `tools/frontend_import_boundaries.json`
- M: `tools/frontend_source_ownership.json`
- M: `tools/harness/generated-artifacts/check-json-shapes.mjs`
- M: `tools/test_families/module.workbook.json`
- M: `tools/test_families/package.grid_adapter.json`
- M: `tools/test_families/package.view_contracts.json`
- M: `tools/test_families/web.workbook.json`


### RSR-05 exit — PASS

All applicable acceptance rows have passing evidence; there is no BLOCKED row.
The final working-tree inventory matches the 97 paths above. Branch and HEAD
still match the captured clean baseline; no commit or staging operation was
performed. A final search finds no remaining production import or source-owner
entry for the retired broker, option hook, buckets or ReferenceRequirement.

Documentation lint and diff whitespace checks passed. Final handoff byte review
confirms valid UTF-8, a terminating LF, no CR bytes and no trailing whitespace.
The completion-row change is followed by the same byte and diff checks.
This completes the authorized implementation and validation. Next action:
review the working-tree diff and retained evidence; no additional implementation
or approval is required for this scoped handoff.
