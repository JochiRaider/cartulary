# Timeline collection-input cleanup

## Execution control

Execution baseline revalidated on 2026-09-18: clean `main`, HEAD
`f31a0fda2faf2269fb4721739ce3204a7a768d1a`. Only root `AGENTS.md` applies.
No pre-existing edits. Preserve user work. Scope is Host reference, Identity
reference and Tag token authoring, immediate Grid/Inspector consumers, necessary
shared primitives, owner clarification and verification. No digest edits,
commits, pushes, deployment, analyst data, dependencies, persistence, new APIs,
relationship capabilities or matching policies are authorized.

| Workstream | Status | Exit |
| --- | --- | --- |
| TCI-01 Characterization and owners | DONE | Production baseline, bounded gap ledger and unambiguous owner contracts. |
| TCI-02 Input lifetime | DONE | Passive cells omit inputs; surface-specific retained authoring survives valid updates. |
| TCI-03 Interaction retirement | DONE | One transition per action, duplicate machinery retired, capture steps preserved. |
| TCI-04 Integrated evidence | DONE | Applicable behavior, production continuity, measurement and visual gates pass. |
| TCI-05 Terminal handoff | DONE | Final verification and acceptance assessment complete without applicable blockers. |

Only the current row may be IN_PROGRESS. Append actual paths, evidence,
commands/results, risks, dispositions and next action; save DONE before starting
its successor. An unresolved applicable blocker stops dependent work.

## Authority and accepted decisions

Read order followed during planning: localized digest README, START_HERE,
LOCAL_AGENT_PROMPT, REPO_MAP, OWNER_MAP, rules, acceptance, QUERY_RECIPES and
UPSTREAM_MAP. Digest and historical handoffs are advisory. The NLSpec research
essay is not another product request or product authority.

Behavior owners: Core 01 §7.4.1 and REQ-01-313–322 (collection values/actions),
Core 02 §§6–8/15 (mentions, provenance/history), Core 03 §§3–4/12–14/16/18 and
REQ-03-217–219/283/298–300 (authoring, native keyboard, continuity and authority),
Core 04 §§1–2 (authorization). Domain owns vocabulary/navigation; design owns
declared presentation. Source/import manifests and verification routing are
independent of behavior authority. Executable inputs never consume Markdown.

The user approved independent grid and Inspector collection drafts, and Escape
cancellation of the local unsubmitted token before Inspector close. Composition
and popup dismissal keep priority. Ordinary collection-cell clicks remain
selection, not scalar editing. Add remains the explicit, keyboard-accessible
raw-entry control. No picker-first requirement is introduced.

Completed relationship/collection, mention-action, grid-autosave, range-selection,
range-entry, clearing, Find and frozen-column handoffs are regression baselines.
Their historical passes are not fresh execution evidence.

## Planning evidence

- `make help`, `make help-all`, task guides for `web.workbook`, `module.timeline`
  and `package.grid_adapter`, and measurement target explanation passed.
- Three Timeline frontend rows passed at
  `.cartulary/test-results/20260918T221635Z-p3861164`; two Workbook frontend rows
  passed at `.cartulary/test-results/20260918T221635Z-p3861175` (fresh execution).
- Planning left tracked files unchanged; `git diff --check` passed.

## TCI-01 investigation

Source-confirmed structures, not yet production defect claims:

- `TimelineCollectionCell` mounts an invisible one-pixel inactive input and keys
  the input by row version. Retained nonempty drafts automatically expose it.
- Foundation activation state changes the collection renderer and column factory.
- Renderer, draft capture/materialization and settlement alias collection drafts
  to grid identity across both surfaces; mounted refs themselves are distinct.
- Keyboard/blur deduplication remembers strings and can suppress settlement
  callbacks. Find and readiness consumers also depend on draft identity.
- Selected-row draft subscriptions enter root composition through observation
  readiness. Actual render work and user impact require profiling.

Next action: author and run isolated production characterization before source
behavior changes; record exact observations and adopt the approved narrow owners.

### Production evidence and remediation ledger

Production characterization uses `apps/web/e2e/timeline-collection-input.spec.ts`
with the existing production build, isolated API-created fixtures and authored
`module.timeline.browser.collection_input_characterization` routing. React
commit/column-prop observations and CDP work counters are diagnostic profiling,
not production timing measurements. The application has no profiling changes.

The successful 24-case selection characterization is retained at
`.cartulary/test-results/20260918T222744Z-p3971363`, attachment
`collection-input-observations` in the collection-input group's Playwright report.
Each case covers one of three fields, 1/100/200/300 loaded records and closed/open
Inspector. One Add activation exposed raw entry. Each intentional fixture-side
same-row Analyst patch replaced the input; all 24 original nodes disconnected,
lost focus and collapsed the backward selection from 4–8 to 22–22. Exact raw
text survived. Activation/typing/live reception issued zero browser record
mutations and zero browser HTTP requests within the sampled action interval.

At one loaded row, the closed Inspector presentation had six collection inputs
(three recordless, three committed); three were initially invisible. Larger
windows mounted 105 inputs closed/108 open, with 101 invisible after activation.
This is viewport/overscan work, not one mounted control per loaded record.
Typical larger-window activation rendered 105/108 collection components; one
fill rendered all 105/108 again. Column-prop identity replacements also occurred
(typically six on activation, five on input). These counters identify unnecessary
work, not a measured speedup or an AC-043 failure.

| Gap / classification | Remediation and affected areas | Rationale / long-term benefit | Compatibility / unresolved risk | Binary validation |
| --- | --- | --- | --- | --- |
| G1 Confirmed focus/selection defect | Remove row-version input identity; explicit accepted-revision reconciliation in cell/registry. | Preserve native authoring through unrelated committed changes. | No wire/data migration; IME and newer text must not be cleared by an older receipt. | Same DOM input, text, caret/selection and composition survive permitted row updates. |
| G2 Confirmed structural/render cost | Mount committed-grid input only for explicit authoring; scope activation subscriptions in existing registry. | Fewer controls and no activation propagation through column assembly. | Preserve Add, trailing draft and visible Inspector entry; detachment must retain text. | Zero passive hidden inputs; explicit activation remains one action; no column invalidation attributable to activation. |
| G3 Approved authoring correction | Carry surface through renderer, materialization, capture, settlement, Find and readiness consumers. | Prevent independently edited presentations from aliasing retained state. | Browser-memory identity only; all callers migrate together; coalesced contributions need exact revision ownership. | Accept/cancel grid A leaves Inspector B and newer grid C unchanged. |
| G4 Approved cancellation clarification | Implement local collection Escape before Inspector close; use existing semantic focus anchors. | One predictable native/editor/Inspector hierarchy. | No mutation cancellation API; admitted attempts remain owned by queue. | Local Escape dispatches nothing, cancels only unsubmitted current context and restores its visible anchor. |
| G5 Source-confirmed duplicate coordination risk | Replace string-only keyboard/blur tracking with revision-specific shared settlement in existing command owner. | Repeated departure callers cannot duplicate dispatch or lose callbacks. | Preserve distinct repeated mention tokens and ordered coalescing; uncertainty retains exact request. | Enter/Tab plus blur produces one attempt and settles every waiting caller; later identical text is a new logical action. |
| G6 Confirmed broad readiness subscription | Subscribe readiness where consumed; commands read current drafts in both surfaces. | Typing does not rerender unrelated collection cells through root readiness. | Source-action eligibility unchanged; readiness must still update when drafts change. | Current eligibility updates locally; typing has no unrelated collection render propagation. |
| G7 Confirmed retained-presentation gap | Include collection drafts in the registry projection consumed by existing Unsaved cells UI. | Preserve readable work when its original grid target is unavailable or read-only, as REQ-03-298 requires. | No store or persistence change; empty cancellation revisions are not displayed; Inspector remains independent. | Filtered drafts are readable, role loss retains exact raw text, and incident access loss conceals it. |

Retain: collection action construction, mutation queue/replay, committed-version
high-water marks, source-bound mention semantics, auto-resolution/Undo, chip and
overflow inspection/return registrations, active input and recordless focus refs,
semantic Grid navigation, Find borrowing and accepted-range policies.
Simplify: collection cell/renderer, registry surface identity and subscriptions,
source readiness, keyboard departure and acceptance reconciliation.
Remove after migrating consumers: invisible inputs, row-version keys, foundation
activation snapshot/callback threading, string-only deduplication, cross-surface
DOM clearing and unconsumed compatibility paths.

The approved normative change is in Core 03 REQ-03-219, with bounded design §9.3
alignment. No adopted owner contradiction remains. Collection capabilities and
the scalar editor eligibility map are unchanged. An implementation-only source
capture review also found that collection contributions must be filtered to the
actual submitted fields before revision settlement; otherwise unrelated drafts
could be mistaken for submitted work. This is covered by G3/G5, not a new API.

### Commands and disposition

- `make generate` passed at `20260918T222251Z-p3864857`; generated browser
  topology changed only for the new authored scenario.
- Initial characterization failed at `20260918T222322Z-p3868029`: fixture used
  a scalar-only selector. Existing collection inspection passed in that run.
  Corrected the characterization to use the semantic relationship selector.
- Second run failed at `20260918T222554Z-p3931767`: fixture PATCH omitted
  required `view_schema_id`. Replaced it with the typed existing patch helper.
  Both failures belong to new test choreography, not product or infrastructure.
- `make frontend-typecheck` passed at `20260918T222752Z-p3981820`.
- `make service-backed-test-slice OWNER=module.timeline` with the four existing
  measurement rows passed at `20260918T222322Z-p3868038`. Baseline p95 ms:
  selection 33.2, focus-edit 39.8, typing 32.2, blank creation 107.1. Existing
  thresholds remain 100/100/100/150 ms, with the unchanged AC-043 fixture,
  traffic and sampling policy. No timing failure was established.
- Characterization extension now exercises collapsed nonterminal caret and
  Chromium composition in addition to selection. Next action: inspect that
  result, complete TCI-01, then start lifetime changes.

TCI-01 exit: the expanded 24-case run passed at
`20260918T222958Z-p4003918`. All six collapsed-caret cases lost focus and moved
7–7 to 22–22; all six Chromium composition cases disconnected the composing
input and moved 5–5 to 19–19 while retaining interim text. The other twelve
cases reproduced selection loss. This confirms native-input interruption, not
raw-text loss or duplicate dispatch. Markdown lint passed at
`20260918T223051Z-p4035082`; `git diff --check` passed. The owners, bounded ledger
and control dispositions above close characterization. No applicable blocker
remains. Next action: TCI-02 stable lifetime and independent draft capture.

## TCI-02 implementation and exit

Changed collection cell/renderer, foundation/interaction/presentation composition,
`useTimelineEditorDraftRegistry`, `WorkbookLocalDraftStore`, mutation draft capture,
and the immediate Observation/source-write/capture-action readiness consumers.
Passive committed grid cells now omit input DOM; Add mounts and focuses one input.
The existing runtime store retains independent surface drafts. Row versions no
longer key input instances. Component subscriptions reconcile accepted revisions;
composition begins a fresh authoring revision, protecting it from older receipts.
Inspector and recordless controls remain visible. Detached drafts display a Draft
marker and require explicit grid reactivation. Existing active-input registration
and exact chip/overflow return registrations remain.

Find already reads and submits its active input's surface identity; no aliasing
fallback was needed there. Source-readiness checks now inspect both surfaces.
Observation readiness subscription moved into the consuming Inspector workflow.
Captured revisions are restricted to fields actually carried by patch payloads.
No wire schema, stored data, or dependency migration is required.

Fresh evidence:

- Timeline collection/registry/keyboard rows passed: `20260918T224042Z-p4040473`.
- Foundation/keyboard rows passed: `20260918T224052Z-p4041409`.
- TypeScript passed: `20260918T224052Z-p4041481`. Earlier local compile failures
  at `20260918T223503Z-p4037814` and `20260918T223941Z-p4039705` exposed stale
  test props and a remaining foundation snapshot reference; both were repaired.
- Autosave, capture authoring and Find rows passed: `20260918T224328Z-p4107516`.
- Observation/source-readiness and evicted-source coordination passed:
  `20260918T224503Z-p4110609`, `20260918T224504Z-p4110848`.
- Production characterization and collection inspection passed together:
  `20260918T224250Z-p4077158` (13/13 execution units). All 24 scenarios preserve
  the original connected input, focus, exact caret/selection and composition text;
  zero invisible inputs and zero browser record mutations. Larger windows mount
  four active/recordless inputs closed or seven with Inspector, versus 105/108.
- Prior integrated run `20260918T224052Z-p4041390` passed characterization but
  failed the historical assertion that Inspector copied a grid token. Updated
  that assertion to the explicitly adopted independent-draft contract; all exact
  inspection focus and no-mutation assertions remain.

Diagnostic counter review found that React work flags remain on reused fibers;
the initial whole-tree traversal overcounts renders. Those initial render counts
are withdrawn as render-work evidence. The instrumentation now excludes fibers
reused from the preceding tree. Corrected comparable profiling remains TCI-04
work; AC-043 timings and input/focus evidence are unaffected.

TCI-02 exit PASS. Remaining risks belong to TCI-03/04: shared keyboard/blur
settlement, cancellation/departure, asynchronous navigation and terminal behavior
coverage. Next action: consolidate those paths in the existing command owner.

## TCI-03 implementation and exit

Collection settlement now belongs to `useTimelineMutationCommands.ts`, with a
revision-specific shared outcome in `timelinePendingSaves.ts`. Keyboard and blur
join that outcome. Distinct subsequent mention tokens receive distinct logical
admission even when their raw strings are identical. Removed the obsolete
collection payload-signature duplicate branch, keyboard/blur source discriminator,
unreachable collection keyboard intent mapper and foundation activation plumbing.
The existing queue still owns ordering, conflicts, uncertainty and replay.

`useTimelineKeyboardController.ts` captures navigation before awaiting acceptance,
checks authoring revision and attachment eligibility, and rejects superseded focus
intent. `useTimelineGridAnchorController.ts` shares recordless navigation planning;
`SemanticDataGrid.tsx` exposes preparation using its existing navigation decision.
`domInteraction.ts` captures native Inspector Tab destinations. The cell owns local
Escape cancellation, with composition priority. No extra capture step was added.

Production Escape exposed the grid vendor's automatic focus redirect from a cell
to its first chip. `semanticFocusRequest.ts` now shields only the synchronous,
explicit cell-focus event from that redirect; normal chip focus and Tab remain.
The focused regression in `semanticFocusRequest.test.ts` passes. Existing input,
recordless, chip and overflow registrations remain necessary and are retained.
The legacy no-revision settlement path remains for runtime recovery contexts;
normal collection admissions capture exact revisions and never use it.

Targeted callback stabilization in presentation, Grid environment, mutation and
Inspector composition removes wrapper-object dependency churn. Corrected profiling
shows ordinary typing was already local, while an open Observation workflow caused
108 unrelated collection renders and six column-prop replacements per fill.
Moving its readiness subscription to `IndicatorInspectorWorkflow.tsx` reduces that
case to one collection render and zero column replacements. G6 is confirmed only
for that active consumer; the broader ordinary-typing hypothesis is rejected.
Selection work is being separated from explicit Add activation in final profiling.

Fresh commands/results:

- `make test-slice OWNER=module.timeline` with collection, registry, keyboard and
  mutation-model rows: PASS, `20260918T232105Z-p470759` (5/5 units).
- `make test-slice OWNER=web.workbook` with keyboard, foundation and Find rows:
  PASS, `20260918T232106Z-p470984` (4/4 units).
- Grid Adapter full narrow owner slice: PASS, `20260918T225911Z-p4167353`;
  subsequent explicit-focus row: PASS, `20260918T231306Z-p272195`.
- Production characterization, independent authoring and promotion rows:
  PASS, `20260918T232344Z-p504945` (11/11 units). All three fields preserve
  newer typing, surface isolation, exact Escape focus and recordless promotion.
  Enter plus blur dispatches once. Later identical mentions remain distinct;
  the existing duplicate-tag `no_effective_change` rejection preserves raw text
  and focus, with one authoritative tag and no duplicate effect.
- `make format`: PASS, `20260918T232357Z-p530589`.
- `make frontend-typecheck`: PASS, `20260918T232402Z-p538290`.
- Intermediate authoring runs exposed test selectors/diagnostics and the real
  Escape redirect described above. `20260918T231307Z-p272453` failed only the
  incorrect expectation that duplicate tags return success; corrected the test
  to existing owner behavior, then PASS at `20260918T231507Z-p398835`.
- Promotion's initial timeout was fixture horizontal positioning; corrected the
  semantic scroll preparation, then PASS at `20260918T231335Z-p301789`.
- Biome at `20260918T231530Z-p428127` required formatting/import organization;
  subsequent Make formatting repaired it. An attempted unknown Biome flag and
  guessed short row IDs were rejected by harness admission, without execution
  or acceptance changes. All subsequent selectors come from the authored catalog.

TCI-03 exit PASS. No owner contradiction or dependent blocker remains. Remaining
risks are integrated attachment/authority coverage and shared Grid regression
coverage. Next action: TCI-04 production lifecycle, regression, timing and visual
validation, including corrected comparable profiling.

## TCI-04 integrated validation

Added production `collection_input_lifetime` routing in
`tools/test_families/module.timeline.json`, generated browser manifests through
`make generate`, and extended `timeline-collection-input.spec.ts`. Isolated real
fixtures cover long Unicode drafts in all three fields, virtualization, hidden
columns, filtering, failed refresh, same-account recovery, account replacement,
read-only role changes, incident revocation and acknowledgement after suspension.
The accepted write is checked against authoritative saved values; navigation,
inspection and retained drafts dispatch zero writes. No analyst incident is used.

Integrated regression findings and repairs:

- The shared explicit-focus shield exposed a support-adapter reliance on bubbling
  focus. Its semantic selection observer now uses capture, matching production's
  explicit selection phase. The unchanged consumer-anchor row passed at
  `20260918T233359Z-p1044221`; full frontend rerun remains required.
- Historical auto-resolution evidence requires Inspector Enter to return to the
  original row's Summary cell. The keyboard owner now captures that destination.
  A second source scroll-restoration request competed with the accepted keyboard
  destination. Collection callers awaiting settlement now exclusively own their
  departure; ordinary blur still uses the existing continuity owner. Auto-resolution
  regression PASS at `20260918T233855Z-p1194030` after earlier failures at
  `20260918T233024Z-p771826` (focus) and `20260918T233400Z-p1045110` (visibility).
- Inspector acceptance can legitimately virtualize a retained grid input when it
  returns to Summary. The new test now reactivates explicitly before checking the
  exact retained draft. Duplicate-tag rejection blocks later queued mutations by
  existing policy; the test checks Inspector acceptance before deliberately
  creating that rejection. Neither expectation changes product policy.
- Profiling identified remaining identity churn in `TimelineWorkbookGrid.tsx`
  (surface, active row and cell-state adapter) and the clipboard callback's whole
  composition dependency. Stabilized those exact values; no blanket component
  memoization or new store was added. Final comparable profiling is pending.

Fresh integrated evidence so far:

- Collection lifetime initial PASS: `20260918T232849Z-p670603`; expanded hidden
  column/refresh/role/revocation PASS: `20260918T233029Z-p779308` (11/11 units).
- All four new collection rows PASS: `20260918T233210Z-p922167` (11/11 units),
  before the additional Inspector acceptance regression assertion.
- Narrow Timeline browser regression slice PASS: `20260918T232820Z-p612561`
  (15/15 units): collection inspection/display profiles, clipboard, clear,
  range entry/cancellation/native keys/authority/scrolling and scalar creation,
  keyboard, refresh, rejection and uncertain creation.
- Narrow Workbook browser regression slice PASS: `20260918T232823Z-p613789`
  (15/15 units): Find borrowing/departure/geometry, frozen columns and scalar
  successive writes, collaboration, refresh, closure, roles, sessions and replay.
- Grid accessibility PASS: `20260918T233106Z-p839808`; mention accessibility
  PASS: `20260918T233246Z-p962096` (11/11 units each).
- Ordinary visual suite PASS: `20260918T233105Z-p838751` (12/12 units), including
  successful `frontend-visual-reconciliation.json`. Reviewed Timeline default,
  collection/Inspector chips and Grid Adapter snapshots against the maintenance
  guide. No golden bytes changed; no refresh trigger or update was needed.
- Frontend full run `20260918T233107Z-p841023` failed two focus-observer rows
  (651/653 units); the capture correction above repairs that shared cause.
- A subsequent all-owner Grid slice passed its unit rows but its browser build
  rejected in-flight source changes (`20260918T233414Z-p1053599`). This is valid
  harness protection; no source-snapshot check was bypassed.
- An overly broad `make test-slice OWNER=module.timeline` at
  `20260918T233245Z-p960430` unintentionally selected unrelated service suites.
  It was stopped through the harness cancellation handler. Its final collation
  rejected secret-capable syntax in a temporary unrelated Evidence trace. That
  raw trace was not read or used as evidence. The run also recorded a creation
  threshold failure and a range-scroll timeout under concurrent verification;
  neither is waived or counted as PASS. Narrow fresh reruns are required.
- `make agent-finalize` PASS: `20260918T233025Z-p772705`, repeated after source
  corrections at `20260918T233940Z-p1263343`. RESULTS_DIR was unset; retained-run
  maintenance skipped because there is no qualifying full-warm-check evidence.
- Typecheck, import boundary and Biome PASS:
  `20260918T234001Z-p1274640`, `20260918T234023Z-p1275650`,
  `20260918T234032Z-p1276139`.
- Generated drift, JSON shape and generated-artifact policy PASS:
  `20260918T233935Z-p1256615`, `20260918T233952Z-p1268368`,
  `20260918T233958Z-p1273731`.

Next action: settle final collection/scrolling and frontend reruns, then run the
four unchanged AC-043 gates without competing verification and close TCI-04 only
on PASS. No timing speedup is inferred from component counts.

### Integrated corrections and resumed execution

The full frontend rerun passed all 653 units at `20260918T234154Z-p1286131`.
Final clipboard fidelity, range scrolling and characterization passed 15/15
units at `20260918T234755Z-p1423002`. The earlier loaded range-scroll timeouts
are not accepted evidence; this narrow successful rerun retains their assertions.

Direct collection-led recordless creation exposed a missing destination after
promotion. The existing draft navigation planner now captures the fresh same-field
input for Enter, and keyboard settlement recognizes accepted promotion without
requiring the retired recordless DOM node. Other detached targets remain invalid.
Host, Identity and Tag creation each issue one POST, save one token and continue
in the fresh immediate input. Authoring and promotion passed at
`20260918T234856Z-p1459087` after the direct-entry failure at
`20260918T234629Z-p1386908`. Wheel navigation also invalidates obsolete delayed
focus intent. No new creation policy or request shape was introduced.

The interrupted stream stopped AC-043 run `20260918T235059Z-p1492104` before
it produced a summary. It is incomplete, not a PASS or a product failure. The
temporary diagnostic baseline worktree and its corrected profiling evidence were
also lost. Resumed execution revalidated `main` and the original HEAD, retained
all slice edits, and recreated the diagnostic baseline at
`/home/jochi/code/cartulary-tci-baseline` on the exact original commit. Only the
characterization test and its authored/generated route differ in that worktree.

Final owner review found G7: the existing parked-draft projection only enumerated
scalar fields, leaving retained collection text unreadable after write-role loss.
`useTimelineEditorDraftRegistry.ts` now includes nonempty grid collection drafts
in the existing `WorkbookParkedGridDrafts` consumer. Empty cancellation revisions
remain invisible and independent Inspector drafts are untouched. The lifetime
browser scenario now asserts exact readable text after filtering and role loss,
then complete concealment after incident membership removal. No second draft
store or new presentation component was needed. Empty cancelled drafts no longer
request draft sizing in passive cells.

Fresh resumed evidence:

- Four collection production rows PASS, `20260919T030650Z-p9856` (11/11 units):
  characterization, independent authoring, direct/pending creation promotion and
  extended lifetime/readability/authority coverage.
- Focused Timeline unit rows PASS, `20260919T030649Z-p9635` (5/5 units);
  Workbook keyboard/foundation/Find PASS, `20260919T030904Z-p49455` (4/4);
  Grid semantic-focus/index PASS, `20260919T030911Z-p50247` (3/3).
- `make generate` PASS, `20260919T030855Z-p46501`, after adding the new registry
  assertion to its authored verification row. The managed outputs were generated.
- Baseline-only generator initially rejected an unsorted appended route at
  `20260919T030757Z-p42432`; sorted its authored catalog and regenerated successfully
  at `20260919T030916Z-p50658`. Main product source was unaffected.

Next action: retain recreated comparable profiling and rerun the interrupted
four-gate AC-043 measurement without competing verification.

### Comparable production observations

Corrected baseline profiling PASS: baseline worktree run
`/home/jochi/code/cartulary-tci-baseline/.cartulary/test-results/20260919T030925Z-p53503`
(11/11 units). Final implementation profiling PASS:
`.cartulary/test-results/20260919T030650Z-p9856`. Both retain the
`collection-input-observations` JSON attachment in
`browser-e2e-webserver-backed/browser-groups/functional-support-default-timeline-collection-input/playwright-report.json`.
They use the same fixtures, viewport, actions, corrected React instrumentation
and 1/100/200/300 loaded windows. Ordinary semantic selection settles before
the measured Add action. The baseline asserts the observed interruption; the
implementation asserts preservation. Neither changes a timing acceptance gate.

| Observation | Baseline | Final implementation |
| --- | --- | --- |
| Add actions to raw input | 1 | 1 |
| Mounted inputs, one row, Inspector closed/open | 6 / 9 | 4 / 7 |
| Mounted inputs, 100–300 rows, Inspector closed/open | 105 / 108 | 4 / 7 |
| Invisible inputs after activation, 100–300 rows | 101 | 0 |
| Collection renders per Add, 100–300 rows, Inspector closed/open | 105–106 / 108–109 | 3 / 6 |
| Column-prop replacements per Add | 6 | 0 |
| Ordinary typing collection renders / column replacements | 0 / 0 | 1 / 0 |
| Typing with Observation workflow open, collection renders / column replacements | 108 / 6 | 1 / 0 |
| Same-row unrelated update: original input/focus survives | 0/24 | 24/24 |
| Exact nonterminal caret, selection or composition state survives | 0/24 | 24/24 |
| Exact interim raw text survives | 24/24 | 24/24 |
| Sampled browser HTTP / record mutations for activation, typing and live reception | 0 / 0 | 0 / 0 |

Mounted counts are viewport/overscan observations, not a claim that every loaded
record mounts. The final four inputs are three immediate recordless controls plus
one active committed input; Inspector adds its three visible controls. One local
typing render now updates the subscribed retained state. The ordinary-typing
regression hypothesis is rejected: broad typing work existed only with the
Observation readiness consumer open. Fixture-side Analyst PATCH requests are
separate from the browser request counts. No speedup is inferred from render
counts; the demonstrated interaction improvement is uninterrupted native editing
without another activation or caret-repair action after a row-version update.

The new registry readability test's explicit routed row also passed after catalog
generation: `20260919T031052Z-p88178` (2/2 units).

### AC-043 timing and exit

All four unchanged measurement rows PASS, 20/20 execution units, at
`.cartulary/test-results/20260919T031112Z-p88760`. The authoritative
`browser-e2e-measurement/frontend-measurement-aggregate.json` is qualified, with
zero scheduler overlaps, one canonical snapshot builder and four isolated clones.
Baseline and final use `ac043_large_grid_snapshot_v1`, snapshot key
`35df4bc4fcfd9d2604015a7b04405d7c60ebae4edf6d02cbd09138a0c7e16b2d`,
unchanged traffic, one excluded warm-up and 100 accepted samples per predicate.
No competing verification ran during the final measurements.

| Existing paint-qualified gate | Baseline p95 ms | Final p95 ms | Limit ms | Result |
| --- | --- | --- | --- | --- |
| Summary selection down | 33.2 | 32.4 | 100 | PASS |
| Summary focus/edit | 39.8 | 28.6 | 100 | PASS |
| Typing acknowledgement | 32.2 | 32.7 | 100 | PASS |
| Blank-row creation | 107.1 | 89.4 | 150 | PASS |

These are comparable gate observations, not a causal speedup claim. All existing
limits pass; the 0.5 ms typing difference establishes no threshold regression.
Collection diagnostic profiling is separate from these uninstrumented timings.

TCI-04 exit PASS. G1–G7 binary validations pass through the production scenarios,
focused units, retained shared mutation regressions and unchanged timing gates.
No applicable behavior, owner, geometry or performance blocker remains. Next
action: TCI-05 refresh routing discovery, finalize harness maintenance, complete
terminal checks and publish the reviewable handoff without committing.

## TCI-05 terminal validation and exit

Verification starts from `web.workbook` and `module.timeline`; Grid Adapter is
required for prepared navigation and explicit cell focus, Entities for mention
inspection and auto-resolution, and Design for affected visual/accessibility
projections. Architecture coverage is the frontend import-boundary check because
the changed boundary is between existing frontend owners. No backend, API, SQL,
dependency, migration or analyst-data change is in scope.

Next action: refresh public target/owner discovery and run `make agent-finalize`
with RESULTS_DIR unset before broader terminal verification. No qualifying current
full-warm-check run exists; retained-run maintenance is intentionally skipped.

Discovery (`make help`, `make help-all`, the five owner task guides) passed.
`make agent-finalize` then passed at `20260919T031634Z-p31843` (1/1 unit), with
RESULTS_DIR unset and the maintenance skip described above.

Terminal verification results:

| Command / selected scope | Result | Run root below `.cartulary/test-results/` |
| --- | --- | --- |
| `make frontend-typecheck` | PASS, 2/2 | `20260919T031725Z-p36020` |
| `make frontend-unit` | PASS, 653/653 | `20260919T031725Z-p36051` |
| `make frontend-import-boundary-check` | PASS, 2/2 | `20260919T031829Z-p39109` |
| `make lint-biome` | PASS, 2/2 | `20260919T031846Z-p49493` |
| `make frontend-fallow-static` | PASS, 2/2, advisory static analysis | `20260919T031858Z-p54864` |
| `make generate-drift` | PASS, 4/4 | `20260919T031943Z-p71133` |
| `make json-shape-check` | PASS, 3/3 | `20260919T032016Z-p84672` |
| `make generated-artifact-policy-check` | PASS, 3/3 | `20260919T032027Z-p89168` |
| `make service-backed-test-slice OWNER=module.timeline`, three visual rows | PASS, 11/11 | `20260919T031725Z-p35893` |
| `make service-backed-test-slice OWNER=package.grid_adapter`, visual and accessibility rows | PASS, 13/13 | `20260919T031725Z-p35952` |
| `make service-backed-test-slice OWNER=module.entities`, two visual rows | PASS, 11/11 | `20260919T032122Z-p7782` |
| `make service-backed-test-slice OWNER=module.entities`, mention accessibility | PASS, 11/11 | `20260919T032230Z-p42693` |
| `make service-backed-test-slice OWNER=module.entities`, auto-resolution/disclosure/Undo | PASS, 11/11 | `20260919T032320Z-p73937` |
| `make lint-markdown` | PASS, final substantive handoff | `20260919T032529Z-p7632` |

The terminal Entities command combined visual, accessibility and functional
profiles in one run (`20260919T031725Z-p35930`). Artifact validation stopped the
run on secret-capable syntax in a temporary auto-resolution browser trace while
another group was active; the functional group and final visual aggregation were
cancelled. The restricted raw resource was not opened. Passed child rows are not
substituted for the cancelled final target. This is a harness evidence-collection
issue, not an accepted product failure or a waiver. Next action: rerun each
Entities profile in its own ordinary Make-owned run root, preserving artifact
validation, and finish the full frontend unit run and Markdown check.

Disposition: the separate ordinary Entities runs above all passed, including
their final artifact aggregation. No trace validator, functional assertion or
acceptance limit changed. The cancelled mixed-profile run is superseded by this
complete routed evidence; no applicable blocker remains from it.

Final visual review used the maintenance guide and the successful ordinary
Timeline, Entities and Grid Adapter runs. Reviewed the matched
`timeline-grid-timeline-default-linux.png`,
`record-relationships-mention-chips-linux.png` and
`timeline-grid-adapter-fixtures-linux.png` snapshots. Compact row scanning,
Add/chip/overflow controls, immediate Inspector entry and semantic focus geometry
remain consistent with the adopted presentation. Each final visual reconciliation
passes under the pinned renderer. No golden bytes were changed or refreshed.

Fallow's blocking package-surface projection reports zero findings for Grid
Adapter and UI Contracts. Its whole-workspace report contains 418 advisory
findings and a nonblocking health subcommand exit of 1; the existing profile
retains these without failing the target. No baseline comparison or clean
whole-workspace static-analysis claim is made. No rule or acceptance threshold
was weakened. The complete current frontend unit run passed all 653 routed units.

### Final scope, compatibility and rollback

Final product scope remains the three existing Timeline collection fields and
their immediate consumers. The implementation lives in collection presentation
(`TimelineCollectionCell.tsx`, `useTimelineCollectionRenderer.tsx`), the existing
draft registry/store, source mutation admission/capture/settlement and keyboard
navigation owners. Composition and `TimelineWorkbookGrid.tsx` no longer carry
global collection activation churn. Inspector Observation/source-readiness and
Find use the originating surface. Narrow shared Grid Adapter changes prepare
semantic navigation and preserve explicit cell focus. Tests and authored
verification rows cover those owners; generated changes are limited to
`tools/browser_e2e_batch_manifest.json` and
`tools/execution_topology_render_index.json`.

The adopted clarification is confined to Core 03 REQ-03-219 and design §9.3.
Core 01 field capabilities, Core 02 mention/tag/history semantics and Core 04
authority handling remain the controlling owners for their scopes. The digest,
research essay, backend behavior, wire contracts, storage, dependencies and
analyst data are unchanged. The historical mention test now expects an empty
independent Inspector draft instead of a copied grid draft; its inspection and
no-mutation assertions remain.

Compatibility impact is deliberately local: ordinary cell selection remains
selection, Add still activates raw entry in one action, chips and overflow retain
inspection, and visible Inspector/recordless inputs remain immediately usable.
Independent surface drafts and local Escape are the approved behavior changes.
Rejected and detached local text remains distinct from committed values; the
existing Unsaved cells UI exposes unavailable grid work. There is no persistent
draft migration, request-version change or database migration.

Retained mechanisms are the runtime draft store, mutation queue/recovery,
committed-version high-water marks, active/recordless input registration, semantic
navigation and chip/overflow return targets. Removed mechanisms are hidden
inactive collection inputs, row-version input keys, foundation activation state
and callback propagation, cross-surface collection DOM clearing, the string-only
keyboard/blur discriminator and superseded collection keyboard-intent/duplicate
branches. Draft revisions and one source-owned settlement now coordinate those
actions. No generic event bus, second editable draft store or editor framework
was introduced.

Review rollback as one source/test/owner/routing slice. Revert this slice's authored
changes under `apps/web/src/workbook/`, its immediate test support and E2E tests,
the shared `packages/grid-adapter/src/` changes, the two owner/design amendments,
and the three authored `tools/test_families/` changes together. Regenerate managed
outputs through `make generate`; do not hand-edit them. Re-run the same narrow
behavior, authority, timing and visual gates. No golden image bytes, database rows,
APIs or dependencies require rollback. Ordinary runtime-local unsent work has no
reload/crash guarantee; preserve any desired text before replacing a running
client. No rollback, commit, push or deployment has been performed.

Evidence limits: Chromium production fixtures and CDP composition establish the
observed browser behavior, not every operating-system IME implementation. Timing
evidence establishes the existing four AC-043 gates on their canonical workload;
diagnostic render counts establish removed work, not a causal speedup. Fixtures
are isolated and do not represent analyst incident data. The baseline worktree
and run artifacts remain available for review. Visuals are implementation-support
evidence, not a Core 05 conformance publication.

Skipped checks with owner/scope rationale: retained-run maintenance is N/A without
a qualifying full-warm-check RESULTS_DIR; visual update and its mandatory two
post-refresh passes are N/A because no golden was refreshed; backend, migration,
API compatibility, release/deployment and broad unrelated service suites are N/A
because no corresponding authored behavior or input changed. Generated routing
drift and shared frontend boundary coverage are applicable and passed.

### Acceptance assessment

| Applicable obligation | Status | Evidence / disposition |
| --- | --- | --- |
| Ordinary collection selection, explicit keyboard-accessible Add and direct raw entry in all three fields | PASS | Collection presentation/keyboard units; production authoring and direct creation; no scalar capability expansion. |
| No passive hidden inputs; localized activation and column work | PASS | 24-case comparable production profiling; four/seven mounted controls and zero activation column replacements. |
| Same-session text, nonterminal caret, selection and composition through row updates | PASS | 24/24 production continuity cases; component acceptance-during-composition and exact-revision units. |
| Independent surface capture, cancellation and newer typing during acknowledgement | PASS | Authoring browser row for Host, Identity and Tag; registry materialization/capture/settlement units. |
| Enter/Tab/blur deduplication, repeated mentions, tag deduplication and saved values | PASS | Source-owner shared-outcome units and authoritative browser assertions; one admitted request per token. |
| Inspection/overflow/navigation do not mutate; exact original return focus | PASS | Collection inspection/display-profile, mention-action, auto-resolution and semantic-focus regressions. |
| First-input creation and recordless promotion | PASS | All three fields preserve pending later text/selection and directly create one saved token with fresh-input focus. |
| Virtualization, filtering, hidden fields, failed refresh and readable rejected work | PASS | Extended collection lifetime production row; explicit reactivation restores original text; G7 readable-draft assertions. |
| Remote/accepted versions, target invalidation, rejection, conflict, uncertainty and replay ordering | PASS | Collection remote/receipt/rejection scenarios plus existing source queue, coordinator, stale-target and scalar/range recovery regressions, all in the passing frontend suite and routed browser slices. |
| Role loss, incident revocation, same-account recovery, account replacement and late acknowledgement | PASS | Collection lifetime production row; retained read-only text, complete concealment and no unauthorized focus restoration. |
| Density, narrow layout, supported zoom/text spacing, frozen columns and accessibility | PASS | Existing collection display-profile/frozen regressions; final ordinary visual and accessibility rows with reviewed matched screenshots. |
| Scalar, range, clipboard, Clear and Find compatibility | PASS | Routed Timeline/Workbook browser slices, final clipboard/range-scrolling rerun, full frontend suite and focused Find/keyboard rows. |
| AC-043 selection, focus/edit, typing acknowledgement and blank creation | PASS | Qualified four-predicate aggregate at `20260919T031112Z-p88760`; unchanged samples, traffic and thresholds. |
| Types, imports, authored/generated routing and static package exports | PASS | Final type/import/Biome, generation drift, shape, policy and blocking Fallow package-surface checks. |
| Visual refresh and post-refresh two-pass requirement | N/A | Design maintenance guide makes these conditional on refreshed golden bytes; none changed. |
| Persistent draft migration, backend/API/database verification and deployment | N/A | No corresponding source, contract, schema, dependency or deployment change; existing frontend owners retain those ports. |
| Retained successful full-warm-run maintenance | N/A | AGENTS.md permits leaving RESULTS_DIR unset; no qualifying full-warm-check evidence exists. |

Coverage is by the changed owner boundary, not every Cartesian combination:
new production scenarios directly exercise all three collection fields; unchanged
generic conflict/uncertainty, stale-target and recovery decisions also retain
their existing owner regressions. No new matching or bulk-operation semantics
were introduced or claimed as verified by this slice.

TCI-05 exit PASS on 2026-09-19. Final `make lint-markdown` and `git diff --check`
passed. All five workstreams are DONE. Final branch remains `main`, HEAD remains
`f31a0fda2faf2269fb4721739ce3204a7a768d1a`; the uncommitted slice contains 56
modified tracked files and two new files (the characterization/behavior E2E file
and this handoff). No digest or golden file changed. No pending product decision,
approval request or applicable blocker remains.

Next action: review the uncommitted slice and retained evidence. Implementation,
verification and handoff are complete; no commit, push or deployment was made.
