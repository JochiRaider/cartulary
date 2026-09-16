# Workbook recovery navigation and panel coordination

## Baseline and execution control

Implementation began on clean `main` at
`bbc26cdbdca3f60274ae0f08f54fa61ba934ebe4`. `git status --short` was empty;
there were no pre-existing changes. The recommendation baseline is unchanged.
No commit, push, deployment, dependency addition, or analyst-data mutation is
authorized. This document is the sole execution tracker.

The localized digest read order was followed during planning. Its advice and
historical handoffs are navigation and historical evidence, not requirements or
fresh verification. Core 01/03/04 and the applicable adopted feature owners
govern behavior; design owns bounded presentation and domain owns vocabulary.
Markdown is not an executable product input.

| Workstream | Status | Dependency | Exit |
| --- | --- | --- | --- |
| WRN-01 Inventory and contract reconciliation | DONE | None | All consumers classified; status and presentation lifetimes owner-grounded. |
| WRN-02 Navigation projection and attachment ownership | DONE | WRN-01 DONE | Stable identities, counts, selection, authorization and disposal tests pass. |
| WRN-03 Shell integration and retirement | DONE | WRN-02 DONE | Scoped consumers use one coordinated panel and retain their capabilities. |
| WRN-04 Integrated workflow, security and accessibility | DONE | WRN-03 DONE | Applicable cross-feature production-control scenarios pass. |
| WRN-05 Final validation and handoff | DONE | WRN-04 DONE | Terminal evidence, acceptance, rollback and handoff are complete. |

Before each workstream only its row becomes `IN_PROGRESS`. Evidence is appended
and the row saved `DONE` before starting its dependent. Applicable blocked
evidence prevents completion and dependent work.

## Agreed presentation and scope

- One compact `Recovery (N)` entry counts unfinished scoped obligations, grouped
  into Needs attention, In progress and Retained drafts. Existing completed
  notices are in a collapsed Completed section outside the count.
- One active shell recovery panel uses shell-owned work-area bounds and internal
  scrolling. Explicit recovery activation detaches the inspector; opening the
  inspector detaches recovery. Owner-required work survives either transition.
- Network Analysis import, table changes and Indicator links join navigation.
  Saved graph operation/job/result recovery stays workspace-local; its dialogs
  participate only in presentation coordination.
- Routine grid authoring, ordinary inspector drafts, local file feedback,
  saved-view controls and query paging remain local. No draft/job/notification
  catalog or mutation engine is introduced.

## Planning evidence

The baseline discovery commands `make help`, `make help-all` and task guides for
`web.workbook`, `module.workbook`, `web.design`, `web.networkflow`,
`web.architecture`, `package.grid_adapter` and `module.extensions` passed.
The two existing status/runtime responsibility rows passed in
`.cartulary/test-results/20260915T210033Z-p69856/run-summary.json`.
This is baseline evidence only, not acceptance of the new seam.

## WRN-01 investigation

Implementation revalidation: `git branch --show-current`, `git rev-parse HEAD`,
`git status --short`, repository `AGENTS.md`, and applicable instruction search.
Only the root AGENTS.md applies. The completed source and owner inventory follows.

### Owner and consumer matrix

Paths below are relative to `apps/web/src/`. Every Core consumer is scoped to
the current incident/client runtime and originating account; source-owner
authority hides its snapshot on suspension. Readable closure/demotion retains
work with actions limited by the owner. REQ-03-100 preserves required same-account
work across session recovery; REQ-03-299 exits revoked incidents; account
replacement/runtime disposal retires protected material. Navigation adds no
authorization or persistence.

State notation: D retained draft; W admitted/waiting/preparing; S submitting;
U uncertain; R rejected/review required; C same-field conflict; A acknowledged;
F refresh required/refreshing; X complete. Only listed states are supported by
that producer. A phase is not a shared workflow machine. Unless specified,
closing detaches presentation only and no completion notice may dismiss a draft
or an unresolved receipt. Source metadata is kept out of ordinary summary copy.

| Producer and source implementation | Behavioral owner | Identity and origin | Supported states and actions | Presentation, completion and verification disposition |
| --- | --- | --- | --- | --- |
| `workbook/components/WorkbookSurfaceRefreshNotice.tsx`; runtime surface registry | Core 03 §§2.3A, 3–4; design §10 | Surface registry debt by `view_schema_id`; exact originating sheet when supplied | F/X; current-authority read only | Existing top-bar notice; remove on accepted refresh. Integrate only unrepresented debt. `web.workbook` batch refresh rows. |
| `workbook/components/WorkbookBatchRecovery.tsx`; `WorkbookBatchOperationOwner` | Core 01 §3.3.5; Core 03 §§11.1, 13.3 | Stable entry ID, immutable plan, originating sheet/range; child conflict keys | W/S/U/R/C/A/F/X; exact retry, read refresh, review conflicts, owner-permitted discard | Independent button/open state and absolute panel. One logical batch with child conflicts, not one count per row. Latest retained completion remains outside count. `web.workbook` batch retention/debt; `module.workbook` clipboard/bulk browser rows. |
| `workbook/history/WorkbookHistoryRecovery.tsx`; `WorkbookRecordHistoryOwner` | Core 01 §3.3.5.0; Core 02 §15; Core 03 §10 | Attempt ID and bound schema/record/action | W/S/U/R/A/F/X; exact replay, history/current review, read refresh, permitted notice dismissal | Local review/replacement UI remains. Independent focus/open/absolute panel retires. `web.workbook` History rows. |
| `workbook/features/entities/WorkbookEntityMergeRecovery.tsx`; `WorkbookEntityMergeOwner` | Core 01 §3.3.5.4; Core 03 §2.3A | Attempt ID; survivor/loser identities and reviewed versions | W/S/U/R/A/F/X; replay, current/history review, refresh, dismiss settled notice | Keep entity semantics and nested history. Replace local placement/focus. `web.workbook` merge rows. |
| `workbook/features/coordination/WorkbookDecisionSupersessionRecovery.tsx`; matching owner | Core 01 §3.3.5; Core 02 §10; Core 03 §16.4 | Attempt ID; original/replacement Decision | W/S/U/R/A/F/X; replay, reviewed successor, refresh, notice dismissal | Keep compound review; retire local panel ownership. `web.workbook` Decision rows. |
| `workbook/features/evidence/TimelineRelatedEvidenceRecovery.tsx`; matching owner | Core 01 §§3.3.5, 7.4.1A, 18/18B; Core 03 §§2.3A, 8 | Draft ID → checkpoint ID; original Timeline and created Evidence; stage request IDs | D/W/S/U/R/C/A/F/X; resume/discard draft, exact stage replay, explicit permitted new-ID review, review/link original source, read refresh | Count checkpoint once across create/link/refresh. Creation receipt does not prove linking. Fixed disclosure retires; forms/reviews stay owned. `web.workbook` related Evidence rows. |
| `workbook/features/notes/NoteCreateRecovery.tsx`; `WorkbookNoteCreateOwner` | Core 01 create/feature contracts; Core 03 §2.3A | Numeric owner draft ID, captured revision, attempt ID; source or Notes sheet | D/W/S/U/R/A/F/X; resume/discard, exact replay, read refresh | Owner attachment symbol; rejected authoring stays editable; completed entries already omitted by this producer. Shared overlay but independent disclosure/focus retires. `web.workbook` Note rows. |
| `workbook/features/coordination/CoordinationCreateRecovery.tsx`; matching owner | Core 01 create/feature contracts; Core 03 §§2.3A, 16.4 | Draft ID/revision, variant, source sheet/record, captured attempt | D/W/S/U/R/A/F/X; resume/discard, exact replay, read refresh | Same attachment policy as Notes, with source-owned variant and capture guards. Completed entries omitted. `web.workbook` Coordination rows. |
| `workbook/features/coordination/ContextualCreateRecovery.tsx`; `WorkbookContextualTaskDecisionCreateOwner` | Core 01 §7.4.1A; Core 02 §10; Core 03 §§2.3A, 16.4 | Draft ID/revision, target type, original source and sheet, attempt ID | D/W/S/U/R/A/F/X; resume/discard, exact replay, read refresh | Keep contextual form and raw reference identities. Fixed disclosure retires. `web.workbook` contextual Task/Decision rows. |
| `workbook/features/assessments/AssessmentAppendRecovery.tsx`; `WorkbookAssessmentAuthoringOwner` | Core 01 Assessment create contract; Core 03 §16.3 | Append attempt ID; subject/support refs and origin sheet | W/S/U/R/A/F/X; exact replay, read refresh | Ordinary append draft stays in its authoring owner; only already-retained append results contribute. Fixed disclosure retires. `web.workbook` Assessment append rows. |
| `workbook/features/indicators/WorkbookIndicatorCreateRecovery.tsx`; matching owner | Core 01 Indicator/observation contracts; Core 03 §9.1 | Attempt ID, original observation/source, canonical Indicator result | W/S/U/R/A/F/X; replay, duplicate/existing result review, refresh, settled dismissal | Keep specialized canonicalization/review. Independent fixed panel retires. `web.workbook` canonical owner rows. |
| `workbook/features/indicators/WorkbookObservationRecovery.tsx`; `WorkbookObservationOwner` | Core 01 observation route contracts; Core 03 §9.1 | Attempt ID; source span or observation identity | W/S/U/R/A/F/X; replay, read refresh, settled dismissal | Owner action labels retained; independent absolute panel retires. `web.workbook` observation owner rows. |
| `workbook/features/indicators/WorkbookIndicatorLifecycleRecovery.tsx`; matching owner | Core 01 interval routes; Core 03 §9.1 | Attempt ID; Indicator and reviewed interval draft | W/S/U/R/A/F/X; replay, read refresh, settled dismissal | Separate append semantics, no same-field resolver substitute. Absolute panel retires. `web.workbook` lifecycle runtime rows. |
| `workbook/timeline/actions/TimelineMentionRecovery.tsx`; matching operation owner | Core 01 §3.3.5.5; Core 03 §9 | Creation key and linked resolution key; original mention/source/span | W/S/U/R/A/F/X; exact creation/resolution replay, independent refresh, return to original mention for review | Correlate created-entity and link stages without counting twice. Fixed disclosure retires. `web.workbook` Timeline mention rows. |
| `workbook/timeline/actions/TimelineCaptureRecovery.tsx`; matching action owner | Core 01 §3.3.5; Core 03 §§6–7 | Attempt ID; original Timeline/action/replacement | W/S/U/R/A/F/X; replay, current review, refresh, settled dismissal | Keep action-specific review. Independent panel/focus retires. `web.workbook` capture action rows. |
| `workbook/features/parties/PartyLinkRecovery.tsx`; `WorkbookPartyLinkOperationOwner` | Core 01 §19; Core 03 §20 | Creation ID, link ID, source record and reference-pair identity | W/S/U/R/C/A/F/X; creation/source exact replay, refresh, original-source link review | Correlate creation with its link; independent unrelated pair changes remain distinct. Keep local `PartyPatchFeedback`. Fixed disclosure retires. `web.workbook` Party rows. |
| `networkFlow/NetworkFlowImportSurface.tsx`; `NetworkFlowImportController` | Network Flow §§5, 7–8, 10, 13–14; Core Import owner | Retained workflow/attempt identity, import session, target extension workspace | D/W/S/U/R/A/F/X; owner mapping/review/apply, exact recovery, permitted fresh upload, observation/read recovery | Gate `access=active` and claimed routes; controller controls presented state. Refactor shell detail placement, preserve mapping decisions. `web.networkflow` import rows. |
| `networkFlow/NetworkFlowTableLifecycle.tsx`; `NetworkFlowTableController` | Network Flow §§5, 7–8, 17, 19; Core 04 §2 | Table ID/action draft; captured transaction ID, reviewed table version | D/W/S/U/R/A/F/X; rename/delete review, exact replay, reviewed new action, current reads | Gate hidden/current capabilities; distinguish retained draft from independent operation. Controller dialog presentation migrates; action controls stay local. `web.networkflow` table rows. |
| `networkFlow/IndicatorLinkDialog.tsx`; `NetworkFlowIndicatorLinkController` | Network Flow §§5, 7, 15; Core 04 §2 | Candidate selector/target, retained draft, captured transaction ID | D/W/S/U/R/A/F/X; target review, exact replay, current read, close | Gate hidden/current role/claim; preserve binding receipt and selector identities. Dialog moves into coordinated detail. `web.networkflow` linking rows. |
| FIFO blocker, overflow, same-field resolver in `workbook/components/WorkbookActiveSurfaceFrame.tsx` | Core 03 REQ-03-041–051, 089, 099–100, 301–302; design §§10.1, 10.5 | FIFO unit ID, refused-work identity, conflict `record_id:field_key`; exact sheet | W/S/U/R/C; FIFO-specific new-ID retry/discard, conflict-class resolution, notice close | Preserve urgent notices, stable priority, origin/cell focus and raw local values; move expanded panel ownership out of mutation snapshots. `module.workbook` queue/conflict/status rows; `web.workbook` status/focus rows. |
| Session status action in `useWorkbookRecoveryFocus.ts` | Core 03 REQ-03-100/299; Core 04 §§1–2 | Current authorization scope, never a retained-work identity | Authentication pause; authoritative account recovery | Remains an authorization action, not another counted draft. Concealment precedes navigation publication. App/Collaboration authorization rows. |

Source owners are Workbook and Network Flow through existing source guides;
verification IDs differ from source placement (`web.networkflow` versus
`web.network_flow`). Shared navigation is presentation support, not a new domain
owner. `tools/frontend_source_ownership.json` and import boundaries own new source
placement; `tools/test_families/**` owns runnable test selectors.

Excluded consumers: ordinary grid/create drafts, inspector draft store, explicit
patch local feedback (except existing Party contributions), file attachment
feedback, saved-view lifecycle, Import Assistant incident controls, query paging,
and saved graph jobs/results. They retain their own controls. Their contributions
to primary save facts are corrected, without adding navigation catalog entries.

### Confirmed gap ledger

| Gap and classification | Remediation and affected artifacts | Rationale and long-term benefit | Compatibility and unresolved risk | Binary validation |
| --- | --- | --- | --- | --- |
| G1 confirmed source defect: broad `explicitRecoveryBlocked` creates Conflict with no cause-specific target | Core 03 REQ-03-089 clarification; runtime projector and owner save-only counts; Indicator/Note/status tests; this inventory | Distinguish saved state from required attention; avoid unrelated resolver routing | Visible label correction; no wire or execution changes. Every owner phase must be classified. | Uncertain write Syncing; rejected feature review/acknowledged refresh alone Saved; genuine conflict/FIFO/overflow Conflict. |
| G2 confirmed source defect: `pendingCount` includes acknowledged reads and status counts explicit writes as replay units | Separate `unsettledMutationCount` from operational admission counts; projector and owner lifecycle tests | Acknowledgement remains authoritative; read recovery cannot masquerade as another write | Admission counts remain unchanged. Late receipts must dominate stale transport state. | Save-only count is zero during failed/in-progress acknowledged refresh and one for a retained uncertain write. |
| G3 confirmed structural defect: independently owned disclosures and competing fixed/absolute placements | Design §§7/10; shared presentation boundary; consumer migrations and coexistence tests | One attachment owner prevents overlap and scales without adding chrome controls | Forms/actions retained. Browser bounds require fresh integrated evidence. | Two activations show only the selected detail, with unchanged owner drafts/attempts. |
| G4 source-level focus risk: several consumers remove focused content without a common fallback | Shared identity selection/focus ownership; consumer tests and browser accessibility | Completion/authority changes cannot select unrelated work or strand focus | Owner security withdrawal is immediate; background completion must not regain focus. | Removal while focused lands on a safe live target; background removal never moves external focus. |
| G5 source-level duplication risk: staged work and receipts have unrelated presentation counts | Owner-produced logical identity and grouping, batch conflict deduplication, completed section | Count obligations rather than implementation stages; stable selection survives settlement | No receipt archive or new execution engine. Separate owner obligations must remain independently reachable. | Draft/attempt/receipt/refresh projections count once; unrelated work counts separately. |
| G6 browser hypothesis: many families overflow narrow chrome or obscure work-area controls | Compact entry and shared geometry; production shell narrow/zoom/visual evidence | Bounded presentation keeps routine grid work dominant | No unrelated visual updates; geometry evidence remains required in WRN-04. | Supported viewport/zoom cases preserve status/account/navigation and internal scrolling. |

### Status and lifetime decisions

REQ-03-089 now explicitly distinguishes retained drafts, admitted unsettled
writes, definitive feature rejection and acknowledged refresh. Existing
`pendingCount`/`blockedCount` remain operation-admission facts; dedicated
`unsettledMutationCount` supplies save presentation. The projector no longer
promotes all feature attention to Conflict or adds explicit writes to the FIFO
in-flight count. No request capture, retry, transport or receipt transition was
changed.

REQ-03-302 and design §§7.4/10.5 now make urgent notices independent of explicitly
opened detail. Their mandatory safe message, actions, FIFO order and keyboard
entry remain; notices do not replace an active panel. Design owns compact scope,
counts, ordering, completion grouping and inspector coordination. These are the
smallest presentation-owner corrections authorized by the plan.

### WRN-01 exit evidence

- `make frontend-typecheck`: PASS, run
  `.cartulary/test-results/20260915T210836Z-p74009`. Initial routing rejection
  required ASCII-sorting the newly added authored selector title; corrected.
- `make test-slice OWNER=web.workbook ROWS=...recovery_navigation,...workbook_save_status_preserves_global_blockers_a_8d590e0883`:
  existing status row passed. New characterization initially used an unsupported
  test helper; corrected to native toggle events.
- Indicator canonical, lifecycle and Observation owner rows passed in
  `.cartulary/test-results/20260915T210959Z-p75505`; that aggregate also contained
  two new fixture assertion failures (wrong Lesson field and an early refresh
  observation), both corrected without production behavior changes.
- Recovery characterization and Note recovery rows: PASS 3/3 execution units,
  `.cartulary/test-results/20260915T211227Z-p77133`. Two real independent retained
  authoring disclosures are simultaneously present; suspension hides content
  and same-account authority recovery retains raw authoring. Note save facts
  distinguish raw draft, uncertainty and acknowledged refresh failure.
- Browser geometry and interaction during asynchronous completion remain
  explicitly classified risks/hypotheses for the production-shell WRN-04 matrix;
  jsdom observations are not claimed as browser layout evidence.

Exit: consumer scope, identity/counting policy, save meaning, urgent notices and
panel lifetimes are specified. No unresolved adopted-owner contradiction blocks
the presentation seam. No wire, backend, retry or storage contract changed.
Next action: implement the instance-scoped navigation/attachment boundary.

## WRN-02 implementation and evidence

`shared/workbookRecoveryNavigation.ts` holds only safe contribution summaries,
selected identity, active attachment callbacks and completed-notice suppression.
`shared/WorkbookRecoveryBoundary.tsx` binds source subscriptions/presentation to
that instance and provides secondary-panel coordination. No module owns a second
copy of authoring, requests or receipts. Updates compare only bounded navigation
metadata; registration tokens reject obsolete publishers and cleanup.

Concrete source projections are `noteRecoveryItems.ts`,
`runtime/workbookBatchRecoveryItems.ts` and
`networkFlow/networkFlowTableRecoveryItems.ts`. Note draft identity survives
capture; batch receipt/conflicts remain one item; extension table dialog identity
survives capture and detachment. Extension projection is exposed through
`NetworkFlowOperations.ts`, preserving its public source boundary.

- Focused navigation model, batch owner and Note owner: PASS 4/4 units,
  `.cartulary/test-results/20260915T211815Z-p79344`.
- `make frontend-typecheck`: PASS,
  `.cartulary/test-results/20260915T211824Z-p80259`.
- Network Flow table projection/owner: PASS 2/2 units,
  `.cartulary/test-results/20260915T211921Z-p81364`. The preceding run exposed a
  test assumption: table dialog detachment stops browser observation and marks
  the still-captured write uncertain. The test now asserts that existing owner
  policy, rather than imposing a universal progress state.

Model tests cover deterministic ordering, stable selection under unrelated
updates, exactly scoped detachment, stale activation, completed-notice dismissal,
duplicate identity rejection, replaced registrations, external-panel switching
and disposal. Source ownership and test selector inputs are authored updates;
no generated files were edited. Shell migration remains WRN-03 work.

WRN-02 exit: the final navigation-model rerun passed (2/2 execution units),
`.cartulary/test-results/20260915T212027Z-p82190`. Creation, batch and extension
projections use their existing owner identities; no execution state is held by
the navigation boundary. Next action: migrate shell presentation in WRN-03.

## WRN-03 implementation and exit evidence

All matrix producers now publish owner-derived metadata and attach their existing
forms through `shared/WorkbookRecoveryBoundary.tsx`. Workbook and extension
work areas provide the shared overlay host. `WorkbookRecoveryPanel.tsx` owns the
entry, grouped list, detail attachment, internal scrolling and focus transitions.
`WorkbookShellTopBar.tsx` no longer receives a growing inventory of triggers.
Core FIFO, overflow and same-field details share that host; urgent detached
notices remain on the active surface. Status and direct cell actions resolve the
same semantic parent, including batch, Evidence conflict keys and compound Party
operations. Acknowledged surface debt is hidden from the list when its operation
already represents the read obligation.

Inspector, incident drawer, Network Analysis inspector, contributor panel and
saved-graph dialog activation coordinate presentation. Network import mapping,
table lifecycle and indicator linking retain their controllers and captured
requests. Mapping and linking are non-modal panel content. Table detachment keeps
its existing observation/uncertainty policy. No navigation action submits work.

Retired: `components/RecoverySurface.tsx`, the workbook-local overlay module,
Network Flow mapping modal placement, feature top-bar recovery triggers and local
open/focus state, runtime conflict-panel state and redundant Batch announcements.
Source-owned nested review and reference controls, receipts and allowed dismissal
remain. Generic Completed dismissal only suppresses navigation notice metadata.
Canonical Indicator creation and independent Observation resolution retain two
items when both owners hold independently recoverable requests/receipts; linked
mention, Party and Evidence stages use their existing correlated identity.

G4 refinement: REQ-03-302's retry-to-resolver sentence now explicitly conditions
continuation on the initiating panel still owning interaction. The core adapter
retains only unit/key/activation presentation metadata through retry. A background
result cannot reopen recovery or replace newer work. Network Analysis row-refresh
focus restoration now requires grid focus, preserving external panel interaction.
These are presentation changes, with unchanged queue and transport transitions.

Fresh passing evidence:

- Navigation/UI model, production shell ordinary creation: 4/4 units,
  `.cartulary/test-results/20260915T214828Z-p7090`.
- History, merge, save status and runtime responsibilities: 5/5 units,
  `.cartulary/test-results/20260915T214531Z-p95252`.
- Network Analysis import mapping: 2/2,
  `.cartulary/test-results/20260915T214604Z-p96249`; exploration, graph-local
  recovery and captured indicator linking: 4/4,
  `.cartulary/test-results/20260915T214830Z-p7320`.
- Navigation, batch, Evidence, Party, contextual and Coordination slices: 9/9,
  `.cartulary/test-results/20260915T220152Z-p28568`.
- Grouped paste plus FIFO retry/discard production controls: 3/3,
  `.cartulary/test-results/20260915T220315Z-p37950`.
- Collaboration queue/security and retry-to-conflict rows passed in
  `.cartulary/test-results/20260915T220315Z-p37956`; its third row initially
  expected the retired panel border. The corrected shared-host assertion passes
  in `.cartulary/test-results/20260915T220407Z-p40686` (2/2).
- Existing keep-saved, merged-value and use-unsaved resolver actions: 4/4,
  `.cartulary/test-results/20260915T220354Z-p39448`.
- Latest frontend typecheck: 2/2,
  `.cartulary/test-results/20260915T220355Z-p39739`.
- Import boundary: 2/2, `.cartulary/test-results/20260915T214605Z-p96542`.
- Make-owned formatting passes; a newly introduced test non-null assertion was
  removed after lint identified it. Earlier typecheck failures were migration
  imports/props and exact optional helper types, corrected before these passes.

Historical automatic-opening and feature-summary focus assertions were migrated
to explicit status activation and the common heading. Request/content assertions
remain. Browser layout and cross-feature accessibility still require WRN-04;
these unit results do not claim browser evidence. Next action: exercise the
production shell with simultaneous work and current authority transitions.

## WRN-04 integrated investigation

Production-shell coexistence passed in
`.cartulary/test-results/20260915T221315Z-p94754`: actual Recovery controls host a
retained Note draft, uncertain batch, acknowledged Note refresh and rejected
Indicator lifecycle review together. Counts, destination identity, raw authoring
retention, nested source picker during another result, read-only refresh,
same-account conceal/recovery, account replacement and disposal are asserted.
The current navigation/production-shell/runtime-registration slice also passed
in `.cartulary/test-results/20260915T224009Z-p464` (5/5 execution units).

Fresh browser evidence:

- Note, Coordination and batch accessibility passed in
  `20260915T221029Z-p52968`. The expanded batch scenario adds a retained Note,
  switches between both families, and passes in `20260915T221917Z-p7692`.
  That aggregate also passed contextual creation, Indicator lifecycle,
  observations and Timeline Evidence. Its four remaining failures were isolated
  for correction; the aggregate itself is not acceptance.
- Network Analysis import, table lifecycle and protected-state retirement passed
  in `20260915T222100Z-p76552`; both Network Analysis accessibility rows passed
  after removing their retired modal-loop expectations in
  `20260915T222447Z-p83233`.
- Timeline clipboard fidelity, exact uncertain replay with later typing and
  read-only acknowledged refresh, ranges/fill, first-input creation, pointer
  navigation and spreadsheet keyboard cases passed in
  `20260915T222059Z-p75570`. Grouped-paste conflict recovery subsequently passed
  in `20260915T223049Z-p52331`; the separate tagging continuation subsequently passed in
  `20260915T224821Z-p74782`.

Browser review confirmed G4/G6 refinements: the common heading needed a visible
focus ring; activation needed to reset the internal scrollport; the compact
entry needed a non-wrapping label. Those corrections live in
`WorkbookRecoveryPanel.tsx`. Completion of an action that disables its invoker
also exposed competing grid focus restoration. Source review found obsolete
surface-registry conflict-focus callbacks and unconditional post-resolution
focus in Generic, Entity and Timeline surfaces. These are removed; reconciliation
still applies authoritative rows and clears only captured draft revisions.
`contracts/design/presentation.v1.json` now projects design's invoker/list
fallback, generated through `make generate` (`20260915T223757Z-p29228`, PASS).
A browser focus trace then identified `GridHandle.cancelEdit` requesting vendor
cell focus when FIFO discard cleared an editor behind the panel. Its production
binding now requests cell focus only if focus is already inside the grid. The
existing adapter editor-lifetime test asserts external recovery focus survives.
Keyboard cancellation stays unchanged; no request/retry contract changes.

G4 binary acceptance remains: explicit activation enters the panel; completion
keeps panel-owned focus; a newer external interaction wins; background outcomes
never reopen an inspector. Browser completion verification passed in `20260915T224730Z-p42712`.
G6 acceptance remains bounded host geometry and visible controls at narrow,
compact, zoom and text-spacing settings, followed by reviewed visual evidence.

Two full ordinary visual attempts (`20260915T222058Z-p75437` and
`20260915T223228Z-p17701`) exposed expected chrome/panel differences plus obsolete
fixture activation assumptions. They are failed investigation evidence. At those
checkpoints no visual goldens had been updated. The later reconciliation and
Make-owned update sequence below resolves these investigation failures.

The tagging browser reproduction confirmed a geometry defect rather than a grid
checkbox failure: inserting an urgent notice as the first child consumed the
active frame's only `1fr` row. `WorkbookActiveSurfaceFrame.tsx` now allocates an
`auto` notice row followed by `minmax(0, 1fr)` for the sheet. The reproduction
uses the original production checkbox without viewport-forcing workarounds.

Fresh correction evidence: FIFO completion focus and viewport accessibility
passed in `20260915T224730Z-p42712`; the real Grid Adapter editor-lifetime and
semantic-focus rows passed in `20260915T224729Z-p42486`. The original tagging
checkbox and independent conflicts-only batch path passed in
`20260915T224821Z-p74782`, and same-field accessibility/geometry passed in
`20260915T224822Z-p75964`.

The expanded production-shell test (closure/demotion, exact duplicate-guarded
retry and receipt-preserving notice dismissal) passed in
`20260915T225110Z-p86104`. The 23 scoped Workbook owner rows passed in
`20260915T224846Z-p36990`; eight Network Analysis rows passed in
`20260915T225033Z-p80457`; architecture source/import/geometry policy rows passed
in `20260915T225032Z-p79570`. Core FIFO/resolver/status rows passed in
`20260915T225132Z-p86945` and `20260915T225133Z-p87170`.

All four paint-qualified Timeline measurement rows passed in
`20260915T225131Z-p86728` (20/20 execution units). Extension startup, lazy Network
Analysis loading and Base client identity continuity passed in
`20260915T225630Z-p69852`.

The complete ordinary visual run `20260915T224937Z-p46051` resolved all 252
captures/goldens and all 29 registered fixtures, with zero orphan, missing,
ambiguous or unresolved mappings. Every functional assertion passed; 141 image
comparisons require intentional shell-entry/panel refresh. All 141 actual images
were reviewed through diagnostic contact sheets and full-size representative
recovery panels. The urgent action's default browser styling was corrected to
the existing design button before update. Viewport, renderer, fonts, masks and
comparison tolerances are unchanged. Fixture activation now explicitly opens
recovery; measurements use the common work-area host instead of retired wrappers.

One narrow visual retry, `20260915T224043Z-p30836`, failed before tests with
`infra/service_readiness_timeout`; its missing group summary was consequential.
The subsequent full ordinary run resolved every fixture and provides replacement
evidence. This infrastructure failure is separate from product acceptance.

The many-entry browser fixture first used single-cell paste, which correctly
remained ordinary field editing (Saved and no additional recovery item). The
fixture now uses a two-column paste to exercise an owner-admitted batch. This
preserves the documented exclusion of routine edits from navigation.

Verification routing enforces exact owner membership, including rows that list
`web.workbook` as a collaborator. An attempted combined collaborator slice was
rejected before execution; feature browser rows are now run under their actual
`module.entities`, `module.assessments`, `module.parties`, `module.timeline`,
`module.revisions`, `module.networkflow` and `module.workbook` owners.

### Additional integrated findings

- G7 reproduced: acceptance-required Entity reference refresh swallowed failed
  reads while Timeline was active, so mention creation lost its acknowledged
  refresh obligation. Remediation: `query/useEntitySurfaceQuery.ts` rejects failed,
  aborted, inactive and superseded acceptance-required reads; ordinary reference
  reads remain silent. Core 03 acknowledged-read ownership and the mention owner
  already require this distinction; no specification change or retry semantics.
  Focused Entity reader coverage and the real mention recovery browser row prove
  rejection retains recovery and a later accepted read completes it. The query
  guide documents the contract. Benefit: receipts never imply refreshed views.
  Compatibility: failed refresh remains discoverable; no wire/storage change.
  Risk: reference replacement must still fence obsolete results. Binary exit:
  failed read keeps Refresh created entity; successful read removes it without
  another create or resolution request.
- G3/G5 refinement: detaching graph contributors initially cleared its endpoint
  selection and invalidated the Indicator link draft opened from that endpoint.
  `NetworkFlowExplorationPanel.tsx` now detaches only drawer presentation. The
  existing navigation owner retains selection and query state; explicit selection
  reattaches it. Its focus test and extension browser linking/table-continuity
  rows cover this. Source guide updated. Benefit: shell presentation cannot alter
  feature applicability. No request/identity migration. Risk: explicit close must
  retain its existing selection-clearing policy. Binary exit: recovery hides the
  drawer, selection remains, new link is editable, and explicit selection returns
  to contributors without a new graph query.
- Batch list summaries now expose bounded original-input previews, record counts
  and accessible descriptions, distinguishing same-family work at narrow widths.
  The preview only inspects a bounded prefix; full original input stays with the
  batch owner. Mention/Timeline acknowledged refresh metadata also participates
  in existing surface-debt deduplication.
- Feature sweep: Assessment `230252Z-p68995` and History `230539Z-p63919`
  passed. Entity merge passed in `230136Z-p3584`; mention exposed G7. Party
  `230344Z-p831`, Timeline capture `230447Z-p32712`, Network Flow
  `230637Z-p95299` and Indicator lifecycle `230838Z-p27498` require reruns
  after explicit-panel expectations and the graph detachment fix. These failed
  runs are diagnostic evidence, not acceptance.
- Visual update `225709Z-p1428` did not promote: a notice moved into collapsed
  Completed between observation and activation. The production-control fixture
  helper now retries through the Completed disclosure when that transition occurs.
  No forced clicks, screenshot tolerances or masks were added.

- G6 reproduced during Party conflict continuation (`231335Z-p43617`): the
  common overlay at layer 12 intercepted the view-bar's layer-9 surface menu.
  Its work-area layer is now 8, matching the coordinated inspector allocation
  below navigation. The unchanged production menu click is the binary browser
  check; no forced navigation is used. Affected shared placement and its guide;
  benefit is reachable surface/saved-view controls during recovery. No execution
  or storage migration. Risk: nested owner pickers must remain within the panel.

### Visual refresh record

`make browser-e2e-visual-update` passed in
`.cartulary/test-results/20260915T231139Z-p67678` (12/12 execution units).
All 252 captures reconcile; 141 failing goldens were promoted and 111 passing
files retained. No missing, orphan, ambiguous or unresolved capture remains.
122 promoted files are pixel-identical to the previously reviewed actuals;
the remaining 19 were reviewed again, including the urgent recovery button,
common heading focus and accepted notice. No unrelated golden was regenerated.
Two fresh ordinary visual passes remain required in WRN-05.

The following exact files are under
`apps/web/e2e/workbook.visual.spec.ts-snapshots/`. Registry fixture `—` means
that the routed scenario is inline rather than registered as a named fixture.
Owner and capture identity are recorded by the reconciliation artifact above.

| Golden filename | Routed row | Registered fixture |
| --- | --- | --- |
| `account-menu-controls-compact-linux.png` | `web.design.visual.account_menu_root_and_nested_viewport_states_cef60727cc` | — |
| `account-menu-controls-narrow-linux.png` | `web.design.visual.account_menu_root_and_nested_viewport_states_cef60727cc` | — |
| `account-menu-controls-short-linux.png` | `web.design.visual.account_menu_root_and_nested_viewport_states_cef60727cc` | — |
| `account-menu-long-label-text-spacing-linux.png` | `web.design.visual.account_menu_root_and_nested_viewport_states_cef60727cc` | — |
| `account-menu-long-label-zoom-linux.png` | `web.design.visual.account_menu_root_and_nested_viewport_states_cef60727cc` | — |
| `account-menu-workbook-root-linux.png` | `web.design.visual.account_menu_root_and_nested_viewport_states_cef60727cc` | — |
| `collaboration-conflict-resolver-linux.png` | `module.collaboration.visual.capture_presence_markers_same_field_conflict_res_f0a62c52a1` | `visual.fixture.same_field_conflict` |
| `collaboration-grid-blocked-conflict-linux.png` | `module.collaboration.visual.the_visual_harness_asserts_syncing_same_field_co_df11cd99bc` | — |
| `collaboration-grid-conflict-resolver-compact-linux.png` | `module.collaboration.visual.the_visual_harness_asserts_same_field_conflict_m_c472bd3f9c` | — |
| `collaboration-grid-conflict-resolver-linux.png` | `module.collaboration.visual.the_visual_harness_asserts_same_field_conflict_m_c472bd3f9c` | — |
| `collaboration-grid-conflict-resolver-narrow-linux.png` | `module.collaboration.visual.the_visual_harness_asserts_same_field_conflict_m_c472bd3f9c` | — |
| `collaboration-presence-markers-linux.png` | `module.collaboration.visual.capture_presence_markers_same_field_conflict_res_f0a62c52a1` | `visual.fixture.presence_overflow` |
| `contextual-decision-authoring-linux.png` | `module.workbook.visual.contextual_task_decision_creation` | — |
| `contextual-decision-authoring-narrow-linux.png` | `module.workbook.visual.contextual_task_decision_creation` | — |
| `contextual-decision-recovery-linux.png` | `module.workbook.visual.contextual_task_decision_creation` | — |
| `contextual-decision-recovery-narrow-linux.png` | `module.workbook.visual.contextual_task_decision_creation` | — |
| `contextual-decision-references-narrow-linux.png` | `module.workbook.visual.contextual_task_decision_creation` | — |
| `contextual-task-request-authoring-linux.png` | `module.workbook.visual.contextual_task_decision_creation` | — |
| `contextual-task-request-authoring-narrow-linux.png` | `module.workbook.visual.contextual_task_decision_creation` | — |
| `contextual-task-request-recovery-linux.png` | `module.workbook.visual.contextual_task_decision_creation` | — |
| `contextual-task-request-recovery-narrow-linux.png` | `module.workbook.visual.contextual_task_decision_creation` | — |
| `contextual-task-request-references-narrow-linux.png` | `module.workbook.visual.contextual_task_decision_creation` | — |
| `coordination-comm-log-authoring-linux.png` | `module.workbook.visual.coordination_create_authoring_recovery` | `visual.fixture.contextual_coordination_creation` |
| `coordination-handoff-authoring-linux.png` | `module.workbook.visual.coordination_create_authoring_recovery` | `visual.fixture.contextual_coordination_creation` |
| `coordination-lesson-authoring-linux.png` | `module.workbook.visual.coordination_create_authoring_recovery` | `visual.fixture.contextual_coordination_creation` |
| `coordination-recovery-linux.png` | `module.workbook.visual.coordination_create_authoring_recovery` | `visual.fixture.contextual_coordination_creation` |
| `coordination-recovery-narrow-linux.png` | `module.workbook.visual.coordination_create_authoring_recovery` | `visual.fixture.contextual_coordination_creation` |
| `coordination-source-narrow-linux.png` | `module.workbook.visual.coordination_create_authoring_recovery` | `visual.fixture.contextual_coordination_creation` |
| `coordination-status-review-authoring-linux.png` | `module.workbook.visual.coordination_create_authoring_recovery` | `visual.fixture.contextual_coordination_creation` |
| `decision-supersession-accepted-linux.png` | `module.workbook.visual.decision_supersession_review_recovery` | — |
| `decision-supersession-review-linux.png` | `module.workbook.visual.decision_supersession_review_recovery` | — |
| `decision-supersession-review-narrow-linux.png` | `module.workbook.visual.decision_supersession_review_recovery` | — |
| `entity-mention-chip-states-linux.png` | `module.entities.visual.capture_unresolved_token_resolved_chip_auto_reso_d3b74bd9d7` | `visual.fixture.mention_chip_state_matrix` |
| `incident-directory-compact-desktop-workbook-shell-linux.png` | `module.workbook.visual.capture_default_timeline_workbook_shell_with_vie_c06bbcbee0` | `visual.fixture.compact_desktop_workbook_shell` |
| `incident-directory-default-timeline-workbook-shell-linux.png` | `module.workbook.visual.capture_default_timeline_workbook_shell_with_vie_c06bbcbee0` | `visual.fixture.default_timeline_workbook_shell` |
| `incident-directory-narrow-desktop-workbook-shell-linux.png` | `module.workbook.visual.capture_default_timeline_workbook_shell_with_vie_c06bbcbee0` | `visual.fixture.narrow_desktop_workbook_shell` |
| `indicator-lifecycle-authoring-linux.png` | `module.workbook.visual.indicator_lifecycle_authoring` | `visual.fixture.indicator_lifecycle_authoring` |
| `indicator-lifecycle-authoring-narrow-linux.png` | `module.workbook.visual.indicator_lifecycle_authoring` | `visual.fixture.indicator_lifecycle_authoring` |
| `indicator-observation-authoring-linux.png` | `module.workbook.visual.indicator_observations_authoring` | `visual.fixture.indicator_observations_authoring` |
| `indicator-observation-authoring-narrow-linux.png` | `module.workbook.visual.indicator_observations_authoring` | `visual.fixture.indicator_observations_authoring` |
| `lifecycle-closed-linux.png` | `web.design.visual.lifecycle` | — |
| `lifecycle-comfortable-linux.png` | `web.design.visual.lifecycle` | — |
| `lifecycle-compact-linux.png` | `web.design.visual.lifecycle` | — |
| `lifecycle-confirmed-refresh-failure-linux.png` | `web.design.visual.lifecycle` | — |
| `lifecycle-pending-linux.png` | `web.design.visual.lifecycle` | — |
| `lifecycle-reason-linux.png` | `web.design.visual.lifecycle` | — |
| `lifecycle-review-linux.png` | `web.design.visual.lifecycle` | — |
| `lifecycle-review-narrow-linux.png` | `web.design.visual.lifecycle` | — |
| `lifecycle-review-spacing-linux.png` | `web.design.visual.lifecycle` | — |
| `lifecycle-review-zoom-linux.png` | `web.design.visual.lifecycle` | — |
| `lifecycle-uncertain-linux.png` | `web.design.visual.lifecycle` | — |
| `linked-note-authoring-linux.png` | `module.workbook.visual.note_create_authoring_recovery` | — |
| `linked-note-authoring-narrow-linux.png` | `module.workbook.visual.note_create_authoring_recovery` | — |
| `linked-note-recovery-linux.png` | `module.workbook.visual.note_create_authoring_recovery` | — |
| `linked-note-recovery-narrow-linux.png` | `module.workbook.visual.note_create_authoring_recovery` | — |
| `linked-note-source-narrow-linux.png` | `module.workbook.visual.note_create_authoring_recovery` | — |
| `membership-audit-comfortable-linux.png` | `web.design.visual.membership_audit_browsing` | — |
| `membership-audit-compact-linux.png` | `web.design.visual.membership_audit_browsing` | — |
| `membership-audit-cursor-recovery-linux.png` | `web.design.visual.membership_audit_browsing` | — |
| `membership-audit-empty-linux.png` | `web.design.visual.membership_audit_browsing` | — |
| `membership-audit-inspected-linux.png` | `web.design.visual.membership_audit_browsing` | — |
| `membership-audit-inspected-narrow-linux.png` | `web.design.visual.membership_audit_browsing` | — |
| `membership-audit-inspected-spacing-linux.png` | `web.design.visual.membership_audit_browsing` | — |
| `membership-audit-inspected-zoom-linux.png` | `web.design.visual.membership_audit_browsing` | — |
| `membership-audit-loading-linux.png` | `web.design.visual.membership_audit_browsing` | — |
| `membership-audit-stale-linux.png` | `web.design.visual.membership_audit_browsing` | — |
| `membership-management-comfortable-linux.png` | `web.design.visual.membership_management_visual` | — |
| `membership-management-compact-linux.png` | `web.design.visual.membership_management_visual` | — |
| `membership-management-confirmed-refresh-failure-linux.png` | `web.design.visual.membership_management_visual` | — |
| `membership-management-loading-linux.png` | `web.design.visual.membership_management_visual` | — |
| `membership-management-pending-linux.png` | `web.design.visual.membership_management_visual` | — |
| `membership-management-removal-narrow-linux.png` | `web.design.visual.membership_management_visual` | — |
| `membership-management-removal-spacing-linux.png` | `web.design.visual.membership_management_visual` | — |
| `membership-management-removal-zoom-linux.png` | `web.design.visual.membership_management_visual` | — |
| `membership-management-role-linux.png` | `web.design.visual.membership_management_visual` | — |
| `membership-management-uncertain-linux.png` | `web.design.visual.membership_management_visual` | — |
| `metadata-closed-linux.png` | `web.design.visual.metadata_editing` | — |
| `metadata-comfortable-linux.png` | `web.design.visual.metadata_editing` | — |
| `metadata-compact-linux.png` | `web.design.visual.metadata_editing` | — |
| `metadata-confirmed-refresh-failure-linux.png` | `web.design.visual.metadata_editing` | — |
| `metadata-conflict-linux.png` | `web.design.visual.metadata_editing` | — |
| `metadata-dirty-linux.png` | `web.design.visual.metadata_editing` | — |
| `metadata-loading-linux.png` | `web.design.visual.metadata_editing` | — |
| `metadata-review-narrow-linux.png` | `web.design.visual.metadata_editing` | — |
| `metadata-review-spacing-linux.png` | `web.design.visual.metadata_editing` | — |
| `metadata-review-zoom-linux.png` | `web.design.visual.metadata_editing` | — |
| `metadata-saving-linux.png` | `web.design.visual.metadata_editing` | — |
| `metadata-uncertain-linux.png` | `web.design.visual.metadata_editing` | — |
| `network-flow-analysis-accepted-inspector-linux.png` | `module.networkflow.visual.capture_deterministic_claimed_network_analysis_a_47b1c2cce6` | `visual.fixture.claimed_network_analysis_workspace_states` |
| `network-flow-analysis-compact-saved-graphs-linux.png` | `module.networkflow.visual.capture_deterministic_claimed_network_analysis_a_47b1c2cce6` | `visual.fixture.claimed_network_analysis_compact_workspace` |
| `network-flow-analysis-delete-dialog-linux.png` | `module.networkflow.visual.capture_deterministic_claimed_network_analysis_a_47b1c2cce6` | `visual.fixture.claimed_network_analysis_workspace_states` |
| `network-flow-analysis-graph-contributors-linux.png` | `module.networkflow.visual.capture_deterministic_claimed_network_analysis_a_47b1c2cce6` | `visual.fixture.claimed_network_analysis_workspace_states` |
| `network-flow-analysis-mapping-dialog-linux.png` | `module.networkflow.visual.capture_deterministic_claimed_network_analysis_a_47b1c2cce6` | `visual.fixture.claimed_network_analysis_workspace_states` |
| `network-flow-analysis-narrow-query-controls-linux.png` | `module.networkflow.visual.capture_deterministic_claimed_network_analysis_a_47b1c2cce6` | `visual.fixture.claimed_network_analysis_narrow_workspace` |
| `network-flow-analysis-rejected-diagnostics-linux.png` | `module.networkflow.visual.capture_deterministic_claimed_network_analysis_a_47b1c2cce6` | `visual.fixture.claimed_network_analysis_workspace_states` |
| `network-flow-analysis-saved-graph-result-linux.png` | `module.networkflow.visual.capture_deterministic_claimed_network_analysis_a_47b1c2cce6` | `visual.fixture.claimed_network_analysis_workspace_states` |
| `ordinary-closed-retained-narrow-linux.png` | `module.workbook.visual.ordinary_create_authoring_recovery` | — |
| `ordinary-recovery-1280-linux.png` | `module.workbook.visual.ordinary_create_authoring_recovery` | — |
| `ordinary-recovery-390-linux.png` | `module.workbook.visual.ordinary_create_authoring_recovery` | — |
| `ordinary-reference-authoring-linux.png` | `module.workbook.visual.ordinary_create_authoring_recovery` | — |
| `record-relationships-mention-chips-linux.png` | `module.entities.visual.the_visual_harness_captures_unresolved_mention_a_4b882068c7` | — |
| `timeline-grid-timeline-default-linux.png` | `module.timeline.visual.the_visual_harness_captures_a_deterministic_time_a19d57e206` | — |
| `timeline-mutation-pending-replay-status-linux.png` | `module.workbook.visual.capture_save_state_pending_replay_transaction_re_70f3e80a67` | — |
| `timeline-mutation-transaction-recovery-panel-compact-linux.png` | `module.workbook.visual.capture_save_state_pending_replay_transaction_re_70f3e80a67` | — |
| `timeline-mutation-transaction-recovery-panel-linux.png` | `module.workbook.visual.capture_save_state_pending_replay_transaction_re_70f3e80a67` | — |
| `timeline-mutation-transaction-recovery-panel-narrow-linux.png` | `module.workbook.visual.capture_save_state_pending_replay_transaction_re_70f3e80a67` | — |
| `timeline-related-evidence-authoring-linux.png` | `module.workbook.visual.timeline_related_evidence` | — |
| `timeline-related-evidence-authoring-narrow-linux.png` | `module.workbook.visual.timeline_related_evidence` | — |
| `timeline-related-evidence-partial-linux.png` | `module.workbook.visual.timeline_related_evidence` | — |
| `timeline-related-evidence-partial-narrow-linux.png` | `module.workbook.visual.timeline_related_evidence` | — |
| `timeline-related-evidence-party-narrow-linux.png` | `module.workbook.visual.timeline_related_evidence` | — |
| `timeline-supersession-accepted-linux.png` | `module.workbook.visual.timeline_capture_actions` | — |
| `timeline-supersession-authoring-linux.png` | `module.workbook.visual.timeline_capture_actions` | — |
| `timeline-supersession-review-linux.png` | `module.workbook.visual.timeline_capture_actions` | — |
| `timeline-supersession-review-narrow-linux.png` | `module.workbook.visual.timeline_capture_actions` | — |
| `workbook-inspector-compact-actions-linux.png` | `module.workbook.visual.capture_inspector_details_relationships_evidence_a56cae74ea` | `visual.fixture.inspector_compact_actions` |
| `workbook-inspector-history-linux.png` | `module.workbook.visual.capture_inspector_details_relationships_evidence_a56cae74ea` | `visual.fixture.base_inspector` |
| `workbook-inspector-narrow-technical-details-linux.png` | `module.workbook.visual.capture_inspector_details_relationships_evidence_a56cae74ea` | `visual.fixture.inspector_narrow_technical_details` |
| `workbook-inspector-public-error-linux.png` | `module.workbook.visual.capture_inspector_details_relationships_evidence_a56cae74ea` | `visual.fixture.base_inspector` |
| `workbook-inspector-relationships-linux.png` | `module.workbook.visual.capture_inspector_details_relationships_evidence_a56cae74ea` | `visual.fixture.base_inspector` |
| `workbook-inspector-rollback-preview-linux.png` | `module.workbook.visual.capture_inspector_details_relationships_evidence_a56cae74ea` | `visual.fixture.base_inspector` |
| `workbook-preferences-comfortable-linux.png` | `module.workbook.visual.preferences` | — |
| `workbook-preferences-compact-linux.png` | `module.workbook.visual.preferences` | — |
| `workbook-preferences-confirmed-stale-linux.png` | `module.workbook.visual.preferences` | — |
| `workbook-preferences-uncertain-linux.png` | `module.workbook.visual.preferences` | — |
| `workbook-preferences-uncertain-narrow-linux.png` | `module.workbook.visual.preferences` | — |
| `workbook-preferences-unset-linux.png` | `module.workbook.visual.preferences` | — |
| `workbook-query-empty-text-spacing-linux.png` | `module.savedviews.visual.capture_saved_view_selector_active_chips_grouped_3da7859cdc` | `visual.fixture.empty_successful_query` |
| `workbook-query-empty-zoom-200-linux.png` | `module.savedviews.visual.capture_saved_view_selector_active_chips_grouped_3da7859cdc` | `visual.fixture.empty_successful_query` |
| `workbook-query-saved-view-query-controls-linux.png` | `module.savedviews.visual.capture_saved_view_selector_active_chips_grouped_3da7859cdc` | `visual.fixture.saved_view_query_controls_and_grouped_result` |
| `workbook-view-bar-filter-editing-overflow-linux.png` | `module.savedviews.visual.capture_saved_view_selector_active_chips_grouped_3da7859cdc` | — |
| `workbook-view-bar-long-columns-linux.png` | `module.savedviews.visual.capture_saved_view_selector_active_chips_grouped_3da7859cdc` | — |
| `workbook-view-bar-maximum-pressure-base-linux.png` | `module.savedviews.visual.capture_saved_view_selector_active_chips_grouped_3da7859cdc` | — |
| `workbook-view-bar-maximum-pressure-compact-linux.png` | `module.savedviews.visual.capture_saved_view_selector_active_chips_grouped_3da7859cdc` | — |
| `workbook-view-bar-maximum-pressure-narrow-linux.png` | `module.savedviews.visual.capture_saved_view_selector_active_chips_grouped_3da7859cdc` | — |
| `workbook-view-bar-ordered-maximum-sort-linux.png` | `module.savedviews.visual.capture_saved_view_selector_active_chips_grouped_3da7859cdc` | — |
| `workbook-view-bar-saved-view-actions-linux.png` | `module.savedviews.visual.capture_saved_view_selector_active_chips_grouped_3da7859cdc` | — |
| `workbook-view-bar-saved-view-clean-linux.png` | `module.savedviews.visual.capture_saved_view_selector_active_chips_grouped_3da7859cdc` | — |
| `workbook-view-bar-saved-view-modified-linux.png` | `module.savedviews.visual.capture_saved_view_selector_active_chips_grouped_3da7859cdc` | — |
| `workbook-view-bar-text-spacing-linux.png` | `module.savedviews.visual.capture_saved_view_selector_active_chips_grouped_3da7859cdc` | — |
| `workbook-view-bar-zoom-200-linux.png` | `module.savedviews.visual.capture_saved_view_selector_active_chips_grouped_3da7859cdc` | — |

## WRN-04 exit evidence

All previously failing applicable feature workflows now pass through production
controls. Run roots below are under `.cartulary/test-results/20260915T`:

| Check | Result | Run suffix |
| --- | --- | --- |
| Entity reference acceptance, replacement and cleanup | PASS, 2/2 units | `231304Z-p36201` |
| Mention create/resolve exact recovery and acknowledged reads | PASS, 11/11 | `231243Z-p6432` |
| Party creation/link conflict, replay, navigation and read recovery | PASS, 11/11 | `231626Z-p13237` |
| Timeline capture exact request and receipt preservation | PASS, 11/11 | `231715Z-p76167` |
| Network indicator link and table lifecycle continuity | PASS, 11/11 | `231626Z-p13009` |
| Indicator lifecycle late outcome and closed-incident recovery | PASS, 11/11 | `231723Z-p86309` |
| Production shell composition including saved-view detachment | PASS, 5/5 | `231953Z-p3903` |
| Navigation, current shell, Timeline and mention owner rows | PASS, 6/6 | `231400Z-p72464` |
| Status priority and save-fact responsibilities | PASS, 3/3 | `231638Z-p53757` |
| Graph focus/detachment retaining endpoint selection | PASS, 2/2 | `231242Z-p6378` |
| Many batches plus Note, narrow/zoom/text spacing and descriptions | PASS, 11/11 | `231836Z-p39853` |
| Network graph and indicator keyboard accessibility | PASS, 13/13 | `231849Z-p64403` |
| Frontend typechecking | PASS, 2/2 | `231837Z-p40106` |
| Full reviewed visual candidate update | PASS, 12/12 | `231139Z-p67678` |

The Timeline rerun fixed a test-observation race: history was sampled before the
first uncertain operation settled. It now waits for the owner's uncertain
feedback before recording history. Exact receipt/request equality remains
asserted. The Party menu failure was fixed in product geometry, not bypassed.
The final shell suite corrected one old assertion that kept acknowledged Party
writes Syncing while reads ran; admission remains blocked as required, but save
state is Saved. Its authored exact-title selector was updated before generation.

### Integrated acceptance dispositions

| Applicable scenario | Disposition and evidence |
| --- | --- |
| All scoped producers and source-specific actions | PASS: consumer matrix plus focused owner slices and feature browser sweeps above; no producer is silently excluded. |
| Four families, count, destination, rejected feature without conflict | PASS: production-shell row, including saved Note refresh, retained Note, uncertain batch and rejected interval. |
| FIFO/overflow/conflict/status priority | PASS: status/runtime rows, FIFO accessibility, same-field accessibility and original bulk checkbox/conflicts-only workflow. |
| A→B→close→A, nested picker during another outcome | PASS: production-shell and navigation tests retain raw Note authoring, captured batch and Note picker. |
| Stable ordering, stale activation, withdrawal, disposal | PASS: navigation model and panel focus tests; owner subscriptions are independently fenced. |
| Base/saved-view/extension navigation | PASS: saved-view production-shell addition, Party real menu, extension startup/Base identity, Network replay after Timeline navigation. |
| Exact retry, duplicate activation, late acknowledgement | PASS: production-shell duplicate guard, capture/mention/Party/Network/Indicator browser requests and receipts. |
| Acknowledged refresh reads only | PASS: Note shell, mention, Assessment, Party and existing affected owner browser checks. |
| Completed dismissal preserves receipts | PASS: production-shell compares retained batch entry object after notice dismissal; completed items are outside count. |
| Closure/role changes/session recovery/account replacement | PASS: shell authority scenario, feature late-closed replay, Network protected-state retirement and owner lifecycle slices. |
| Incident access loss/protected content concealment | PASS: owner authority and Entity access-loss tests; current Network protected-state browser; navigation clears safe projections on withdrawal. |
| Focus entry/Escape/return/completion/newer interaction | PASS: common panel tests, FIFO completion browser, graph focus, production owner keyboard checks and Grid Adapter cancellation test. |
| Non-color states, accessible names, long/many entries, narrow/zoom/text spacing | PASS: final batch and Network accessibility plus other scoped owner a11y rows and reviewed visual captures. |
| Timeline direct editing/caret/range/clipboard/fill/first-input | PASS: fresh browser workflow rows recorded above, four paint-qualified measurements and real adapter lifetime tests. Final repeat after the adapter change is routed in WRN-05. |
| Backend/API/new execution semantics/persistent storage/bulk recovery | N/A: explicitly excluded; no such implementation is added. |
| Ordinary grid/inspector drafts, query paging, local errors, saved-graph jobs as recovery entries | N/A: remain with their adopted owners; save facts and external-panel coordination are covered where applicable. |

WRN-04 exit: confirmed gaps have product/test/documentation remediation; every
applicable integrated workflow has passing evidence. No blocked dependency
remains. Risks remaining for terminal verification are generated-input drift,
format/import policy and two ordinary visual runs against promoted goldens.
Next action: WRN-05 finalization and terminal evidence.

## WRN-05 terminal verification

`make generate` passed after authored selector maintenance in
`20260915T232024Z-p9001`. `make agent-finalize` passed before broader terminal
checks in `20260915T232117Z-p12181`; its authoritative result is
`unit-artifacts/finalize-summary.json`. It changed zero files. `RESULTS_DIR` was
unset: retained-run canonical/scheduler maintenance and warm-check performance
review were skipped with `results-dir-not-provided`. Focused runs are not
misrepresented as full warm-check evidence.

Current task guides were consulted for each affected unit owner. The final sweep
selects cataloged Vitest rows for the changed test files, keeping exact owner
routing. Two old expectations required migration: a Timeline batch completion
must be opened through Completed, and Network table/link activation focuses the
common heading with an accessible route out rather than trapping focus in the
old dialog. Mutation target, request, validation and retention assertions remain.
Initial diagnostic roots `232251Z-p68995`, `232307Z-p89008` and
`232352Z-p96582` are not acceptance; corrected reruns are recorded below.

### Final implementation and compatibility

The shell owns one instance-scoped navigation projection and one work-area host.
Each source publishes only safe metadata and presentation callbacks; its own
subscription, drafts, requests, receipts, validation and authorization remain
feature-owned. Identities survive authoring/capture/settlement/refresh, and
completed notices are presentation-only. Global save state stays separate.

Retired paths include `workbook/components/RecoverySurface.tsx`, the old
`workbook/layout/WorkbookWorkAreaOverlay.tsx` placement and
`networkFlow/NetworkFlowMappingModal.tsx` (replaced by the shared host and
`NetworkFlowMappingPanel.tsx`). Independent top-bar disclosures, conflict panel
flags in mutation snapshots, duplicate focus-return callbacks and feature-owned
absolute/fixed recovery wrappers are removed. All scoped owners use the new
entry; saved graph dialogs participate only in external-panel coordination.

Behavioral compatibility changes are intentional: retained drafts and rejected
feature reviews no longer falsely show Conflict; uncertain admitted writes show
Syncing; acknowledged read recovery shows Saved. Opening recovery detaches an
inspector; reopening it requires explicit activation. Owner authoring and
captured execution survive. Urgent notices stay visible without opening panels.
There is no API, wire identity, storage, dependency, retry-policy or data migration.
The Grid Adapter cancellation correction preserves external panel focus while
keeping cancellation inside the grid unchanged.

### Changed non-golden files

The exact 141 golden filenames and their owner/fixture mappings appear in the
visual refresh record above. The remaining changed files are listed here;
deleted paths are marked. The visual manifest and generated design/topology
projections were produced through Make-owned tooling.

- `apps/web/e2e/assessments.recovery.spec.ts`
- `apps/web/e2e/contextual-coordination-create.spec.ts`
- `apps/web/e2e/contextual-create.spec.ts`
- `apps/web/e2e/history-browsing.spec.ts`
- `apps/web/e2e/history-recovery.spec.ts`
- `apps/web/e2e/indicator-canonical-create.spec.ts`
- `apps/web/e2e/indicator-lifecycle.spec.ts`
- `apps/web/e2e/indicator-observations.spec.ts`
- `apps/web/e2e/inspector-actions.spec.ts`
- `apps/web/e2e/keyboard.spec.ts`
- `apps/web/e2e/linked-note-create.spec.ts`
- `apps/web/e2e/mentions.recovery.spec.ts`
- `apps/web/e2e/merge-recovery.spec.ts`
- `apps/web/e2e/network-flow.spec.ts`
- `apps/web/e2e/parties.recovery.spec.ts`
- `apps/web/e2e/sentinel.spec.ts`
- `apps/web/e2e/support/collaboration/replay.ts`
- `apps/web/e2e/support/workbook/coordinationCreate.ts`
- `apps/web/e2e/support/workbook/noteCreate.ts`
- `apps/web/e2e/support/workbook/recovery.ts`
- `apps/web/e2e/support/workbook/timelineRelatedEvidence.ts`
- `apps/web/e2e/timeline-grid-entry.spec.ts`
- `apps/web/e2e/timeline-related-evidence.spec.ts`
- `apps/web/e2e/timeline-workbook.spec.ts`
- `apps/web/e2e/workbook.a11y.spec.ts`
- `apps/web/e2e/workbook.visual.spec.ts`
- `apps/web/src/networkFlow/IndicatorLinkDialog.tsx`
- `apps/web/src/networkFlow/NetworkAnalysisWorkspace.test.tsx`
- `apps/web/src/networkFlow/NetworkAnalysisWorkspace.tsx`
- `apps/web/src/networkFlow/NetworkFlowExplorationPanel.tsx`
- `apps/web/src/networkFlow/NetworkFlowImportController.ts`
- `apps/web/src/networkFlow/NetworkFlowImportSurface.tsx`
- `apps/web/src/networkFlow/NetworkFlowIndicatorLinkController.ts`
- `apps/web/src/networkFlow/NetworkFlowMappingModal.tsx` (deleted)
- `apps/web/src/networkFlow/NetworkFlowMappingPanel.tsx`
- `apps/web/src/networkFlow/NetworkFlowSemanticGrid.tsx`
- `apps/web/src/networkFlow/NetworkFlowTableController.test.ts`
- `apps/web/src/networkFlow/NetworkFlowTableLifecycle.tsx`
- `apps/web/src/networkFlow/README.md`
- `apps/web/src/networkFlow/explorationFocus.test.tsx`
- `apps/web/src/networkFlow/networkFlowImportState.ts`
- `apps/web/src/networkFlow/networkFlowIndicatorLinkOperation.test.ts`
- `apps/web/src/networkFlow/networkFlowIndicatorLinkOperation.ts`
- `apps/web/src/networkFlow/networkFlowIndicatorRecoveryItems.ts`
- `apps/web/src/networkFlow/networkFlowTableRecoveryItems.ts`
- `apps/web/src/networkFlow/useNetworkFlowModalFocus.ts`
- `apps/web/src/shared/README.md`
- `apps/web/src/shared/WorkbookRecoveryBoundary.tsx`
- `apps/web/src/shared/WorkbookWorkAreaOverlay.tsx`
- `apps/web/src/shared/workbookRecoveryNavigation.test.ts`
- `apps/web/src/shared/workbookRecoveryNavigation.ts`
- `apps/web/src/testing/TimelineWorkbookRuntimeFixture.tsx`
- `apps/web/src/testing/WorkbookRecoveryFixture.tsx`
- `apps/web/src/workbook/README.md`
- `apps/web/src/workbook/WorkbookShell.collaboration.test.tsx`
- `apps/web/src/workbook/WorkbookShell.sentinel.test.tsx`
- `apps/web/src/workbook/WorkbookShell.surfaces.test.tsx`
- `apps/web/src/workbook/WorkbookShell.tsx`
- `apps/web/src/workbook/components/EntityWorkbookSurface.tsx`
- `apps/web/src/workbook/components/GenericWorkbookSurface.tsx`
- `apps/web/src/workbook/components/README.md`
- `apps/web/src/workbook/components/RecoverySurface.tsx` (deleted)
- `apps/web/src/workbook/components/WorkbookActiveSurfaceFrame.test.tsx`
- `apps/web/src/workbook/components/WorkbookActiveSurfaceFrame.tsx`
- `apps/web/src/workbook/components/WorkbookBatchRecovery.tsx`
- `apps/web/src/workbook/components/WorkbookEditRecoveryPanel.tsx`
- `apps/web/src/workbook/components/WorkbookIncidentControlsPresentation.tsx`
- `apps/web/src/workbook/components/WorkbookQueueOverflowNotice.tsx`
- `apps/web/src/workbook/components/WorkbookRecoveryPanel.tsx`
- `apps/web/src/workbook/components/WorkbookSameFieldConflictResolver.tsx`
- `apps/web/src/workbook/components/WorkbookShellTopBar.tsx`
- `apps/web/src/workbook/components/WorkbookSurfaceRefreshNotice.tsx`
- `apps/web/src/workbook/features/NetworkFlowOperations.ts`
- `apps/web/src/workbook/features/assessments/AssessmentAppendRecovery.tsx`
- `apps/web/src/workbook/features/assessments/WorkbookAssessmentAuthoringOwner.ts`
- `apps/web/src/workbook/features/coordination/ContextualCreateRecovery.tsx`
- `apps/web/src/workbook/features/coordination/CoordinationCreateRecovery.tsx`
- `apps/web/src/workbook/features/coordination/DecisionSupersessionEditor.tsx`
- `apps/web/src/workbook/features/coordination/README.md`
- `apps/web/src/workbook/features/coordination/WorkbookContextualTaskDecisionCreateOwner.ts`
- `apps/web/src/workbook/features/coordination/WorkbookCoordinationCreateOwner.ts`
- `apps/web/src/workbook/features/coordination/WorkbookDecisionSupersessionOwner.ts`
- `apps/web/src/workbook/features/coordination/WorkbookDecisionSupersessionRecovery.tsx`
- `apps/web/src/workbook/features/coordination/coordinationRecoveryItems.ts`
- `apps/web/src/workbook/features/entities/WorkbookEntityMergeOwner.ts`
- `apps/web/src/workbook/features/entities/WorkbookEntityMergeRecovery.test.tsx`
- `apps/web/src/workbook/features/entities/WorkbookEntityMergeRecovery.tsx`
- `apps/web/src/workbook/features/entities/useEntityMergeController.ts`
- `apps/web/src/workbook/features/evidence/TimelineRelatedEvidenceRecovery.tsx`
- `apps/web/src/workbook/features/evidence/WorkbookEvidenceAttachmentOwner.ts`
- `apps/web/src/workbook/features/evidence/WorkbookTimelineFileOwner.ts`
- `apps/web/src/workbook/features/evidence/WorkbookTimelineRelatedEvidenceOwner.ts`
- `apps/web/src/workbook/features/generic/useGenericWorkbookInspectorComposition.tsx`
- `apps/web/src/workbook/features/indicators/WorkbookIndicatorCreateOwner.test.ts`
- `apps/web/src/workbook/features/indicators/WorkbookIndicatorCreateOwner.ts`
- `apps/web/src/workbook/features/indicators/WorkbookIndicatorCreateRecovery.tsx`
- `apps/web/src/workbook/features/indicators/WorkbookIndicatorLifecycleOwner.ts`
- `apps/web/src/workbook/features/indicators/WorkbookIndicatorLifecycleRecovery.tsx`
- `apps/web/src/workbook/features/indicators/WorkbookObservationOwner.test.ts`
- `apps/web/src/workbook/features/indicators/WorkbookObservationOwner.ts`
- `apps/web/src/workbook/features/indicators/WorkbookObservationRecovery.tsx`
- `apps/web/src/workbook/features/indicators/indicatorLifecycleRuntime.test.ts`
- `apps/web/src/workbook/features/notes/NoteCreateRecovery.tsx`
- `apps/web/src/workbook/features/notes/README.md`
- `apps/web/src/workbook/features/notes/WorkbookNoteCreateOwner.ts`
- `apps/web/src/workbook/features/notes/noteCreateRecovery.test.tsx`
- `apps/web/src/workbook/features/notes/noteRecoveryItems.ts`
- `apps/web/src/workbook/features/ordinary/WorkbookOrdinaryCreateOwner.ts`
- `apps/web/src/workbook/features/parties/PartyLinkRecovery.tsx`
- `apps/web/src/workbook/features/parties/WorkbookPartyLinkOperationOwner.ts`
- `apps/web/src/workbook/history/WorkbookHistoryLocalStatus.tsx`
- `apps/web/src/workbook/history/WorkbookHistoryRecovery.test.tsx`
- `apps/web/src/workbook/history/WorkbookHistoryRecovery.tsx`
- `apps/web/src/workbook/history/WorkbookRecordHistoryOwner.ts`
- `apps/web/src/workbook/hooks/README.md`
- `apps/web/src/workbook/hooks/useWorkbookRecoveryFocus.ts`
- `apps/web/src/workbook/inspector/WorkbookInspectorRecordHistory.test.tsx`
- `apps/web/src/workbook/inspector/useWorkbookRecordHistoryController.ts`
- `apps/web/src/workbook/layout/README.md`
- `apps/web/src/workbook/layout/WorkbookSurfaceLayout.tsx`
- `apps/web/src/workbook/layout/WorkbookWorkAreaOverlay.tsx` (deleted)
- `apps/web/src/workbook/ports/WorkbookTimelineActionRuntimePort.ts`
- `apps/web/src/workbook/query/README.md`
- `apps/web/src/workbook/query/useEntitySurfaceQuery.test.tsx`
- `apps/web/src/workbook/query/useEntitySurfaceQuery.ts`
- `apps/web/src/workbook/runtime/README.md`
- `apps/web/src/workbook/runtime/WorkbookBatchOperationOwner.test.ts`
- `apps/web/src/workbook/runtime/WorkbookBatchOperationOwner.ts`
- `apps/web/src/workbook/runtime/WorkbookConflictStore.ts`
- `apps/web/src/workbook/runtime/WorkbookExplicitPatchOwner.ts`
- `apps/web/src/workbook/runtime/WorkbookManagedPatchDriver.ts`
- `apps/web/src/workbook/runtime/WorkbookMutationRuntime.ts`
- `apps/web/src/workbook/runtime/WorkbookRuntimeResponsibilities.test.ts`
- `apps/web/src/workbook/runtime/WorkbookSurfaceRegistry.ts`
- `apps/web/src/workbook/runtime/workbookBatchRecoveryItems.ts`
- `apps/web/src/workbook/runtime/workbookMutationStatusProjector.ts`
- `apps/web/src/workbook/timeline/actions/TimelineCaptureRecovery.tsx`
- `apps/web/src/workbook/timeline/actions/TimelineMentionRecovery.tsx`
- `apps/web/src/workbook/timeline/actions/WorkbookTimelineCaptureActionOwner.ts`
- `apps/web/src/workbook/timeline/actions/WorkbookTimelineMentionOperationOwner.ts`
- `apps/web/src/workbook/timeline/hooks/useTimelineMutationRuntimeBindings.ts`
- `apps/web/src/workbook/timeline/mutations/useTimelineRowMutationCoordinator.ts`
- `apps/web/src/workbook/timeline/presentation/TimelineWorkbookInspectorRegion.tsx`
- `apps/web/src/workbook/timeline/useTimelineMutationRuntimeBindings.test.tsx`
- `apps/web/src/workbook/utils/workbookPendingQueue.ts`
- `apps/web/src/workbook/utils/workbookStatusSecondary.ts`
- `apps/web/src/workbook/workbookRecoveryNavigation.test.tsx`
- `apps/web/src/workbook/workbookSaveStatus.test.tsx`
- `contracts/design/presentation.v1.json`
- `docs/design.md`
- `docs/handoffs/ui-ux/workbook-recovery-navigation-refactor-handoff.md`
- `docs/spec/03_workbook_interaction_collaboration_and_workflows.md`
- `packages/grid-adapter/README.md`
- `packages/grid-adapter/src/SemanticDataGrid.tsx`
- `packages/grid-adapter/src/index.test.tsx`
- `packages/ui-contracts/src/generated/design-presentation.ts`
- `tools/execution_topology_render_index.json`
- `tools/frontend_source_ownership.json`
- `tools/frontend_visual_golden_manifest.json`
- `tools/test_families/web.networkflow.json`
- `tools/test_families/web.workbook.json`

### Rollback and limits

The captured baseline was clean, so this change has no pre-existing user edits
to merge around. Before rollback, capture a fresh status/diff and preserve any
subsequent user work. Restore only this handoff's listed tracked changes from
`bbc26cdbdca3f60274ae0f08f54fa61ba934ebe4`, remove only its listed new files, and
restore the Make-owned generated projections and visual manifest/goldens as a
matching set. Do not use broad clean/reset commands or remove retained evidence.

Full repository/backend, release/deployment, vulnerability and full warm-check
suites are outside this frontend presentation change and were not run. Narrow
service-backed tests use isolated harness fixtures; no analyst data, deployment,
commit or push was performed. Visual/accessibility results are implementation
and design-support evidence, not a new Core 05 product-conformance claim.

### Terminal results

All roots below are under `.cartulary/test-results/20260915T`; each graph run's
`run-summary.json` records its exact selected rows and execution units.

| Command / owner slice | Result | Run suffix |
| --- | --- | --- |
| `make frontend-typecheck` | PASS | `232206Z-p16111` |
| `make lint-biome` | PASS | `232227Z-p45890` |
| `make frontend-import-boundary-check` | PASS | `232231Z-p46678` |
| `make json-shape-check` | PASS | `232239Z-p47981` |
| `make generated-artifact-policy-check` | PASS | `232245Z-p54065` |
| `make generate-drift` | PASS | `232249Z-p64440` |
| `make lint-markdown` | PASS | `232306Z-p87945` (`adhoc/lint-markdown/tool-run-summary.json`) |
| `make frontend-fallow-static` | PASS | `232559Z-p14413` |
| `make test-slice OWNER=module.collaboration ROWS=…` | PASS, 3 selected rows | `232223Z-p45401` |
| `make test-slice OWNER=package.grid_adapter ROWS=…` | PASS, 3 rows | `232234Z-p47233` |
| `make test-slice OWNER=module.entities ROWS=…` | PASS, 1 row | `232242Z-p49261` |
| `make test-slice OWNER=module.workbook ROWS=…` | PASS, 5 rows | `232711Z-p15830` |
| `make test-slice OWNER=module.networkflow ROWS=…` | PASS, 1 row | `232722Z-p21148` |
| `make test-slice OWNER=web.workbook ROWS=…` | PASS, 31 rows | `232320Z-p90570` |
| `make test-slice OWNER=web.networkflow ROWS=…` | PASS, 21 rows | `232729Z-p21667` |
| `make test-slice OWNER=module.evidence ROWS=…` | PASS, 1 row | `232414Z-p380` |
| `make service-backed-test-slice OWNER=module.timeline ROWS=…` | PASS, clipboard + batch replay/headers + bulk + creation + keyboard + pointer | `232241Z-p48426` |

The 66 unit rows are the authored-catalog rows selecting changed test files,
including the full shell composition suite. Intermediate corrected-unit attempts
`232509Z-p5866`, `232522Z-p10921`, `232530Z-p11444` found remaining old
completion-text/focus assumptions; the passing roots above replace them. No
terminal unit failure remains. Fresh Timeline browser evidence confirms normal
spreadsheet interaction after the Grid Adapter focus correction.

The first ordinary visual verification (`232206Z-p16155`) resolved all 252
captures and had zero functional errors, but 37 screenshot comparisons differed
after lowering the overlay beneath view-bar navigation. Review found small text
rasterization changes at the controls, consistent with the changed paint order;
forms, counts and feedback remain intact. Every affected actual was reviewed,
including full-size diff inspection. A second Make-owned update is required;
no mask, tolerance, renderer, viewport or unrelated workflow change is used.
Two ordinary passing runs must follow that final update.

The final test-only cleanup passes `make frontend-typecheck` in
`233056Z-p59418` and `make lint-biome` in `233056Z-p59428`. The preceding
`233013Z-p58447` failure was an unsupported Testing Library `exact` option in
updated test queries; removing that redundant option preserves exact string
matching and changes no product behavior. Static retirement checks found no
remaining `RecoverySurface`, mutation-owned conflict panel flag, old mapping
modal import or old top-bar recovery prop in production TypeScript.

### Final layer refresh

`make browser-e2e-visual-update` passed in `232911Z-p24984` (12/12 units).
Reconciliation again accounts for 252 active captures/goldens and 29 registered
fixtures with zero missing, orphan, ambiguous or unresolved entries. The layer
correction refresh uses the following subset of the existing 141-file record;
its exact owner rows and fixture identities are in that table. The accepted
trigger is the verified navigation stacking fix. Screenshot scope, zoom, masks,
scroll normalization, renderer and comparison tolerances are unchanged.

- `lifecycle-closed-linux.png`
- `lifecycle-confirmed-refresh-failure-linux.png`
- `lifecycle-pending-linux.png`
- `lifecycle-reason-linux.png`
- `lifecycle-review-linux.png`
- `lifecycle-uncertain-linux.png`
- `metadata-conflict-linux.png`
- `metadata-dirty-linux.png`
- `metadata-loading-linux.png`
- `metadata-saving-linux.png`
- `metadata-uncertain-linux.png`
- `membership-audit-cursor-recovery-linux.png`
- `membership-audit-empty-linux.png`
- `membership-audit-inspected-linux.png`
- `membership-audit-loading-linux.png`
- `membership-audit-stale-linux.png`
- `membership-management-confirmed-refresh-failure-linux.png`
- `membership-management-loading-linux.png`
- `membership-management-pending-linux.png`
- `membership-management-role-linux.png`
- `membership-management-uncertain-linux.png`
- `timeline-related-evidence-partial-narrow-linux.png`
- `account-menu-controls-compact-linux.png`
- `account-menu-controls-narrow-linux.png`
- `account-menu-long-label-text-spacing-linux.png`
- `account-menu-workbook-root-linux.png`
- `contextual-decision-recovery-narrow-linux.png`
- `contextual-task-request-recovery-narrow-linux.png`
- `linked-note-recovery-narrow-linux.png`
- `ordinary-reference-authoring-linux.png`
- `workbook-preferences-confirmed-stale-linux.png`
- `workbook-preferences-uncertain-linux.png`
- `workbook-preferences-unset-linux.png`
- `timeline-mutation-transaction-recovery-panel-compact-linux.png`
- `timeline-mutation-transaction-recovery-panel-narrow-linux.png`
- `workbook-query-empty-text-spacing-linux.png`
- `workbook-view-bar-maximum-pressure-narrow-linux.png`

Final bounds audit moved the existing 38rem recovery width into the design-owned
`layout.recoveryMaxWidth` token, projected in `contracts/design/tokens.v1.json`
and consumed by the shared host. Its rendered value is unchanged; spacing and
work-area clamps remain token-based. This closes the plan's token-owned geometry
requirement without resizing the panel. Generation and final visual runs cover
the authored projection. No new behavior or dependency is introduced.

Additional final projection files: `contracts/design/tokens.v1.json` and
`packages/ui-contracts/src/generated/design-tokens.ts`. `make generate` passed
in `233606Z-p917`. The current routing owner is `package.ui` (an attempted
`package.ui_contracts` guide lookup was rejected before any test execution).
All four token/presentation projection rows pass in `233708Z-p8570` (5/5 units).
The final token value is identical to the reviewed 38rem panel width.

Final-source maintenance rerun: `make agent-finalize` PASS in
`233648Z-p4390`, still with `RESULTS_DIR` unset; frontend typechecking PASS in
`233648Z-p4432`; Biome PASS in `233648Z-p4442`. The earlier visual verification
started before this token-only source projection and is supplemental; two fresh
runs using the final generated token are required for completion.

The final layer promotion was pixel-identical to 32 of the 37 reviewed actuals.
The remaining five (lifecycle reason, membership role, Timeline Evidence partial
recovery, long account label/text spacing, and empty query/text spacing) were
reviewed again after promotion; meaningful content, focus, bounds and controls
remain intact. The refresh record's filenames and owner/fixture mappings cover
all affected files.

Supplemental ordinary visual verification passed in `233426Z-p66900` (12/12
units, all 252 captures). This run used the same rendered width before its final
token projection. Documentation lint passed again in `233817Z-p39238`;
`git diff --check` passes. Final import-boundary verification passed in
`233648Z-p4436`. The handoff inventory accounts for all 304 changed paths,
including exactly 141 goldens, and no excluded backend/API, lockfile, advisory
digest or analyst-data path changed.

Final-token ordinary visual run `233839Z-p45540` passed all 252 captures (12/12
units). Parallel run `233745Z-p10234` passed 251 captures but exposed one
nondeterministic internal scroll position in `collaboration-grid-blocked-conflict`.
Its old `scrollIntoViewIfNeeded()` ran before presentation normalization and had
no declared post-normalization anchor. The fixture now uses the existing visual
anchor helper to center the Keep saved control within the shared recovery
scrollport, with clamping and three-frame observation. This is a fixture
preparation correction, not a product focus override or pixel-offset workaround.
No tolerance/mask/golden change is intended; the narrow owner row and another
ordinary run must pass before completion.

The focused anchor check `234435Z-p80929` passed its functional and geometry
assertions but showed that centering the bottom action clips the common heading.
The final declared anchor is therefore the start of the Recovery navigation
region in the same scrollport, preserving the common navigation controls and
heading. This is the correct semantic framing for the coordinated panel; the
old arbitrarily scrolled golden needs one reviewed refresh. A new Make-owned
update and two ordinary runs are required. The single affected golden is
`collaboration-grid-blocked-conflict-linux.png`, owned by
`module.collaboration.visual.the_visual_harness_asserts_syncing_same_field_co_df11cd99bc`;
its fixture mapping is in the refresh table. No product geometry, scope, mask,
normalization value or comparison tolerance is changed to match old pixels.

After the capture-anchor correction, frontend typechecking passes in
`234646Z-p50356`, Biome in `234646Z-p50366`, and documentation lint in
`234646Z-p50370`. The product sources and generated token remain unchanged;
only the affected visual fixture's declared anchor changed.

The final anchor refresh passed in `234624Z-p17493` (12/12 units). Reconciliation
still resolves all 252 captures, all 252 goldens and all 29 registered fixtures,
with zero missing/orphan/ambiguous/unresolved entries. The promoted conflict
image was reviewed at full size: the common All recovery, Close recovery and
Return to workbook controls and heading are visible; feature controls remain
inside the panel's scrollport; urgent status and grid remain visible. This
single-golden framing change is covered by the owner/fixture mapping above.

### Resume after interrupted terminal runs

The stream/environment restart interrupted ordinary visual runs
`20260915T235104Z-p58474` and `20260915T235104Z-p58488`. Their retained
`run-summary.json` and target summary files are empty, and their processes no
longer exist. They supply no terminal acceptance evidence. Fresh ordinary runs
are required against the final promoted manifest; no source or golden change is
needed to resume them.

Resume revalidation confirms `main` at
`bbc26cdbdca3f60274ae0f08f54fa61ba934ebe4`, with the same 304 changed paths
(including 141 goldens), all accounted for in this handoff. No additional user
changes were discovered. Final-promotion `make json-shape-check` passed in
`20260915T235127Z-p15942` and `make generated-artifact-policy-check` passed in
`20260915T235127Z-p15940`, each with 3/3 execution units.

The resumed run `20260916T000254Z-p3895` has a pre-execution loader failure in
the claimed Network Analysis visual group: `SyntaxError: The requested module
'./browserSession' does not provide an export named 'csrfHeaders'`. The retained
group stdout and `browser-group-result.json` report zero row execution time and
`infrastructure_failed`; scheduler cleanup subsequently reports `artifact_error`.
The unchanged source exports that function, and the companion run loaded and
passed the same group. The underlying transient loader cause is not established;
this is diagnostic evidence only. A fresh complete ordinary run must replace it,
without changing source, goldens or comparison settings.

The interrupted-run replacement `20260916T000254Z-p3904` passed
`make browser-e2e-visual` with 12/12 execution units and all 252 captures. Its
`browser-e2e-visual/frontend-visual-reconciliation.json` resolves all 252 goldens
and 29 registered fixtures, with zero missing, orphan, ambiguous or unresolved
entries. The promoted manifest SHA-256 is
`c17516184f7887e257dcceeaed8ac39353f4ea3764831d4773ec2f61c1dc91d0`.
The failed companion `20260916T000254Z-p3895` finished with 10/12 passing units;
its 46 workbook visual scenarios passed, but the loader failure and consequent
reconciliation failure prevent accepting that run. A second complete ordinary
pass remains required.

### Final acceptance and completion

The second complete ordinary visual run, `20260916T000737Z-p72311`, passes
12/12 execution units and all 252 captures. Both final runs reconcile 252 active
goldens and 29 registered fixtures with zero errors. Their source digests match,
and both attest the final manifest SHA-256 recorded above, which also matches the
current file. This completes the required two fresh ordinary passes after the
final promotion. No further source, fixture or golden change was made.

| Final gate | Disposition | Evidence |
| --- | --- | --- |
| Recovery ownership, scope and status reconciliation | PASS | WRN-01 owner matrix, gap ledger and save-state tests. |
| Identity, counting, ordering, attachment and disposal | PASS | WRN-02 focused model/panel tests and production-shell coexistence. |
| Consumer migration, status routing and retained capabilities | PASS | WRN-03 implementation and WRN-04 integrated acceptance matrix. |
| Workflow, concealment, accessibility and Timeline editing | PASS | WRN-04 acceptance rows and final seven-row Timeline browser slice. |
| Final selected unit coverage | PASS | 66 affected test-file rows plus four `package.ui` token/presentation rows, all passing. |
| Final ordinary visual comparisons and reconciliation | PASS | `20260916T000254Z-p3904` and `20260916T000737Z-p72311`, 252 captures each. |
| Typechecking, Biome, imports and unused-source checks | PASS | Terminal results and final-source reruns above. |
| Authored projection generation, drift, JSON and generated policy | PASS | Make generation/finalization and final-promotion checks above. |
| Documentation, final bytes, scope and whitespace | PASS | `make lint-markdown` in `20260916T001228Z-p107018`; UTF-8/newline/whitespace and inventory inspection; `git diff --check`. |
| Full backend/release/warm-check execution | N/A | Outside the changed frontend presentation boundaries; no full-run claim is made. |
| Retained-run maintenance | N/A | `RESULTS_DIR` unset because no successful full warm-check run is supplied. |

All applicable product and integration gates pass. The transient test-loader
failure remains a harness diagnostic with an unestablished root cause; the two
complete passing runs replace its missing acceptance evidence. No known scoped
product defect or blocked dependency remains. Evidence is limited to the current
owner-routed automated checks and reviewed pinned-renderer visual artifacts;
it does not claim release or new Core 05 conformance.

Final scope remains the 304 documented paths, including 141 reviewed goldens.
The migration, retired paths, compatibility and selective rollback instructions
above are final. No commit, push, deployment, dependency, backend/API or analyst
data change was made. Final documentation lint passed; the handoff decodes as
UTF-8, ends in one newline, and contains no carriage returns or trailing
whitespace. `git diff --check` passes. Branch and HEAD remain the captured
baseline, and the final inventory contains no unrecorded change.

WRN-05 exit criteria and terminal handoff are complete. Next action: review the
working-tree diff and this handoff; no further implementation is outstanding.
