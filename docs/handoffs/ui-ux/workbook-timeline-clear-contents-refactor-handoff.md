# Timeline clear contents and recovery

## Execution control

Implementation baseline revalidated on 2026-09-18: branch `main`, HEAD
`d6dfeabbe63d5d598308786a4629fbae9bb50539`, clean working tree, no pre-existing
edits. Only the repository-root AGENTS.md applies. The authorized seam comprises
owner amendments, public projections, Timeline batch execution, Grid Adapter
intent, Workbook retained recovery, and their verification. No analyst data,
dependencies, persistence, digest, commits, publishing, or other surfaces change.

| Workstream | Status | Dependency / exit |
| --- | --- | --- |
| TCC-01 Characterization and adopted clear contract | DONE | Owner amendments and complete matrices adopted; Markdown and whitespace checks pass. |
| TCC-02 Source command and public contract | DONE | Source and service slices pass; explicit null and public projections generated. |
| TCC-03 Grid interaction and retained recovery | DONE | Shared planner, opt-in adapter capture and retained recovery checks pass. |
| TCC-04 Behavioral and security evidence | DONE | Acceptance matrix passes, G7–G10 resolved, first ordinary visual validation passes. |
| TCC-05 Terminal validation and handoff | DONE | Final source, reviewed goldens, measurements, public contracts, routing and completed handoff validated. |

Only the current row becomes IN_PROGRESS. Append actual evidence and save DONE
before starting its successor. An applicable BLOCKED dependency prevents work
and completion. Historical handoffs are regression navigation, not fresh passes.

## Authority and characterization

Behavior: Core 01 §§3.3.5/7.4.1 and REQ-01-614/672; Core 02 §15; Core 03
§§3–4/6/13–15 and REQ-03-100/299; Core 04 §§1–2; design §§8/10/14.
Domain owns vocabulary/navigation only. The localized digest read order and the
range-entry, range-selection, paste/bulk, clipboard, grid-autosave, Find,
query-continuation, History, correction-access and recovery-navigation handoffs
were inspected during planning. The NLSpec research essay remains advisory.

Source placement is governed by frontend source/import manifests and established
Workbook/Timeline/Revisions boundaries. Verification is independently routed by
the verification registry, owner catalog and authored test families. No executable
consumer may read, stat, hash or otherwise depend on Markdown.

Current implementation trace:

- Grid Adapter `semanticKeyboardPolicy` treats Delete/Backspace as empty editor
  entry; `resolveSemanticCellRange` supplies ordered virtualized membership.
- Timeline's ten operational text fields already declare writable,
  grid-editable and clearable capabilities. `PatchChange` carries nullable text;
  batch-cell `Value` duplicates it as an unused string.
- Workbook authenticates/adopts a source provider at the existing bulk route.
  Timeline admits, locks and applies source batches, then appends one change set
  with a revision per changed record and projection/Collaboration effects.
- The retained Workbook batch owner captures predecessor identities, seals queue
  coalescing, orders overlap, captures exact attempts, retains receipts/conflicts,
  and separates acknowledgement from surface read debt.
- Find's external-action boundary demonstrates focus borrowing without blur
  submission. Selection, Inspector subject and record checkboxes are independent.
- Current date conversion runs after unrelated patches and can populate a null
  date from its counterpart. The adopted amendment makes conversion intent-aware.

## Adopted matrices

| Action/context | Admission and consequence |
| --- | --- |
| Unmodified Delete in Timeline navigation | Capture current committed cell or completed rectangle; admit one clear command. |
| View-bar Clear contents | Same capture/planner; borrow focus without submitting authoring. |
| Editor, native selection, nested control, menu, composition | Preserve local Delete ownership. |
| Backspace / other consumers | Preserve existing active-cell edit and selection behavior. |
| Ineligible member or overlapping unsubmitted draft | Reject whole selection locally; preserve all work and expose safe explanation/recovery. |
| Prior admitted save | Wait for captured predecessor; advance only undispatched base versions from its acknowledgement. |
| Later overlapping edit | Cannot overtake admitted clear; independent work remains available. |
| New selection/navigation after admission | Never retarget clear or restore obsolete focus. |

| Field category | Clear contract |
| --- | --- |
| Ten Timeline v2 operational text fields | Explicit authoritative null, gated by writable/grid-editable/clearable capabilities and current authority. |
| Activity Date pair | Submitted null wins; unselected source text unchanged. Derive existing pair/sort state. Unrelated edits cannot regenerate; later explicit date edits may generate an unsubmitted eligible counterpart. |
| Derived, system, collection, hidden, structural, draft, unauthorized or unsupported member | Ineligible; no silent skipping. |
| Source text with linked objects | Preserve mentions, observations, links, tags and Evidence. |

Eligible keys, individually covered by the authoritative all-fields service case:

| Operational field | Explicit null effect |
| --- | --- |
| `timeline.date_entered_text` | Null source text; no unrelated changes. |
| `timeline.analyst_text` | Null source text; attribution remains system-owned. |
| `timeline.mitre_stage_text` | Null source text; observations remain unchanged. |
| `timeline.device_object_text` | Null source text; entity links/mentions remain unchanged. |
| `timeline.ip_address_text` | Null source text; entity/indicator relationships remain unchanged. |
| `timeline.activity_utc_text` | Null UTC source; preserve unselected local source; recompute pair/projection. |
| `timeline.activity_local_text` | Null local source; preserve unselected UTC source; recompute pair/projection. |
| `timeline.raw_activity_text` | Null multiline source; no relationship removal. |
| `timeline.activity_synopsis_text` | Null synopsis source; no relationship removal. |
| `timeline.data_source_text` | Null source text; Evidence and links remain unchanged. |

The clear request has `kind=clear_cells_v1`, `view_schema_id`, `client_txn_id`,
`field_keys[]` (1–10 unique eligible keys) and `targets[]` (1–500 unique record
IDs and positive base versions). Every named field on every named record is an
explicit null assignment. Arrays preserve order for identity/results/conflicts;
object serialization and UUIDs canonicalize. Values, create targets, selectors,
coordinates and unrelated command members are forbidden.

| Outcome/lifetime | Required accounting |
| --- | --- |
| Invalid admission | Zero writes, revisions, conflicts or protected response material. |
| Accepted material changes | One attributable change set; one revision/version advance per changed record; ordinary capture-state effects. |
| Already null | Authoritative evaluation and replayable receipt; no artificial material change. |
| Same-field conflict | Preserve submitted null and exact saved/base values; ordered per-cell conflict; nonconflicting changes may commit. |
| All conflicts / all no-op | Empty rows; no change-set ID; ordered conflicts or empty conflicts respectively. |
| Uncertain / malformed success | Retain exact attempt; explicit retry uses identical bytes and identity. |
| Acknowledged, refresh failed | Retain receipt and read debt; recovery performs reads only. |
| Session suspension | Conceal protected presentation; retain required same-account work. |
| Incident retirement / account replacement | Retire protected retained state under existing owners; late effects cannot republish it. |

## Gap ledger

| Gap / classification | Remediation and areas | Rationale / long-term benefit | Compatibility / unresolved risk | Binary validation |
| --- | --- | --- | --- | --- |
| G1 Authorized enhancement | Adopt explicit clear and Delete semantics in Core/design, then Adapter/Timeline integration. | Predictable spreadsheet command with one semantic owner. | Intentional Timeline Delete change; Backspace/editors preserved. Risk: scope interception. | Only owned unmodified Delete dispatches one clear. |
| G2 Structural weakness | Retire unused string batch value; use nullable PatchChange through source/conflict/revision paths. | One value authority prevents null loss. | Paste/fill/tag meanings and receipt history unchanged; no data migration. Risk: null becomes empty text. | Null persists and resolves as null; empty/whitespace stay distinct. |
| G3 Source consequence reconciliation | Intent-aware date conversion across source writes and resolutions. | Explicit clearing survives unrelated work without extra persisted state. | Deliberate conversion behavior clarification; no schema migration. Risk: date pair silently refills or alters counterpart. | Clear remains null, unrelated writes preserve pair, later date edit may generate only unsubmitted counterpart. |
| G4 Retained admission | Check local drafts, reuse batch predecessors/reservations/exact attempts. | No implicit authoring loss or duplicate lifecycle owner. | Existing memory-local recovery retained; no reload persistence. Risk: draft submission/overtaking. | Whole-selection rejection or correctly ordered immutable command. |
| G5 Presentation/recovery extension | Shared planner, compact action, explicit null comparisons, command-specific receipt validation and labels. | Both entry points have identical semantics and honest outcomes. | Additive command; old servers reject without fallback. Risk: partial result appears complete. | Partial/conflict/no-op/uncertain/read-debt outcomes distinguishable and keyboard accessible. |
| G6 Verification boundary | Source, service, production browser, accessibility, visual/measurement and ownership evidence. | Correctness is established at the responsible layer. | Historical evidence not reused as a fresh pass. Risk: shared-grid regression. | Every adopted acceptance row PASS; N/A justified; no applicable BLOCKED. |
| G7 Confirmed integration defects | Batch projection previously flushed React per row; later queued/open drafts retained their pre-clear baseline. Use one receipt projection and a source-owned accepted-predecessor callback through the existing driver registry. | Large visible membership remains usable and later authoring follows the admitted clear without artificial stale rejection. | No persisted change or request rewrite. Only undispatched contexts advance; conflicted fields and immutable attempts remain untouched. Risk: advancing an unrelated draft baseline. | A 100-row production clear followed by queued and still-open later edits commits in order, with original targets, one clear request per activation, no lost draft or focus restoration. |
| G8 Confirmed layout regression | Use the compact Clear caption at base width and the eraser icon below it; opt Timeline into icon-only Inspector/Add row actions at narrow widths. Extend query-rail geometry assertions to include Find and Clear. | Prevent obscured active query chips while preserving Find's label, chip capacities, saved-view allocation and view-bar height. | Neutral opt-in presentation property; other surfaces keep their defaults. No query or input semantics change. Residual risk is covered by full visual and focused text-spacing checks. | 1024px and query-pressure accessibility/visual cases show nonoverlapping named reachable controls. |
| G9 Date intent plumbing | Preserve all submitted date-field identities through per-cell conflict filtering in the Timeline batch executor. | A conflicted submitted counterpart must not be regenerated indirectly by an accepted date cell. | Follows G3's adopted unsubmitted-counterpart rule; no request/hash/storage change. Risk: a mixed paste could overwrite a concurrent saved null while reporting its conflict. | After concurrent local-date clear, stale two-date paste changes UTC only, preserves saved local null, reports one conflict and conversion_unavailable. |
| G10 Delivery identity risk | Thread the accessible action's native event through Grid Adapter capture to the existing batch-owner delivery guard, matching Delete. | Repeated delivery of one click remains one semantic activation without suppressing later deliberate actions. | No new deduplication owner or timing heuristic. Risk: action callbacks create a fresh delivery identity on each invocation. | Production browser dispatches the same click twice after a separate keyboard activation and observes exactly one additional request. |
| G11 Confirmed baseline verification defect | Emit each browser group-target mapping through the finalizer's already-supported repeated flag in tools/harness/scheduler/work-graph/browser.mjs. | Necessary catalog aggregate verification otherwise cannot compile the existing full browser inventory. | Both HEAD and working manifests have the same 57 groups and 4,282-character mapping, exceeding the unchanged 4,096-character argument bound. No public command, schema, selection or scheduling change. Risk: incomplete finalizer mapping. | Existing aggregate graph determinism test passes and actual browser/measurement finalizers retain exact group accounting. |

## Evidence log

- Planning and execution baseline: branch/HEAD/status/AGENTS inspection passed;
  clean main at the stated commit. No behavioral tests were run during planning.
- Planning discovery: `make help-all`, task guides for `module.timeline`,
  `module.workbook`, `web.workbook`, `package.grid_adapter`, `platform.openapi`,
  and `git diff --check` passed.

## Compatibility, rollback and next action

The new command adds no route, storage schema or dependency. Released OpenAPI
baselines and committed receipts remain immutable. Older servers reject clear;
clients must not downgrade it to another command. Roll back UI entry points first,
retain compatible replay support until admitted attempts settle. An older server
rejects the new kind before receipt replay, so an unresolved clear attempt cannot
be handed to older code. Keep the intent-aware date rule during rollback: restoring
unconditional conversion would regenerate an accepted null through unrelated
edits. Coordinate specification/projection/generated/implementation/test rollback
without rewriting historical receipts or applying inverse writes. Preserve accepted
writes and historical receipts. No commit, push or deployment is authorized.

### TCC-01 exit

Changed this handoff, Core 01 mutation/Timeline owner sections, Core 03 keyboard,
bulk, write-back and materiality sections, and design action/keyboard presentation.
The matrices above resolve action, field, date, lifecycle, overlap and outcome
decisions. Existing Core 02 history and Core 04 authorization remain governing.
No implementation-dependent semantic decision remains. `make lint-markdown`
PASS at `.cartulary/test-results/20260918T130955Z-p3767150`;
`git diff --check` PASS. Nine current task guides (the five primary owners plus
protocol TS, architecture, design and test catalog) PASS. Historical evidence was
not promoted into current validation. No applicable blocked dependency.

Next action: TCC-02 source command, public projection and focused service evidence.

### TCC-02 exit evidence

- Added `ClearCellsCommand` and nullable source row planning through the existing
  Workbook provider and Timeline batch executor. Retired unused string-only
  batch-cell `Value`; ordinary paste/fill/tag fingerprints remain unchanged.
- Admission bounds, duplicate identities, capabilities, and source lifecycle are
  checked before mutation/conflict payloads. Date patch conversion now respects
  submitted null and only explicit date edits can generate an unsubmitted peer.
- Authored Workbook OpenAPI and six candidate release dispositions were updated;
  immutable releases remain unchanged. `make generate` PASS at
  `.cartulary/test-results/20260918T132204Z-p3779646`. Earlier generation failures
  were missing compatibility dispositions and unsorted new catalog selectors;
  both were corrected without weakening checks.
- `make format` PASS at `.cartulary/test-results/20260918T132421Z-p3807350`.
- `make test-slice OWNER=module.timeline
  ROWS=module.timeline.unit.timeline_admission_decoding_and_hashing_46ac663c5c,module.timeline.support_unit.clipboard_paste_parsing_mapping_provenance_and_b_43e8f0fd4f`
  PASS, 2/2 units, `.cartulary/test-results/20260918T132221Z-p3782940`.
- `make service-backed-test-slice OWNER=module.workbook
  ROWS=module.workbook.integration.timeline_clear_cells,module.workbook.integration.shared_ingest_workbook_clipboard_paste_persists_5ae6dc0770`
  PASS on final source, 3/3 units,
  `.cartulary/test-results/20260918T132435Z-p3811668`. Covers all ten fields,
  whitespace/empty/multiline to SQL null, ordered multiple records, one revision
  per changed record, attribution, no-op, exact replay, changed identity,
  mixed/conflicts-only outcomes, nullable conflict resolution, date persistence,
  invalid targets and viewer authorization; existing batch cases also pass.
- Paths: `internal/modules/timeline/{admission,mutationpolicy}` and source batch,
  date, facade, command and test files; Workbook integration tests;
  `internal/app/workbookassembly`; authored OpenAPI/candidate release;
  generated OpenAPI/Go/TypeScript projections; Timeline/Workbook test families.
- Remaining risk: presentation admission and retained clear receipts are not yet
  connected; no UI behavior is claimed. Final generation/drift will include the
  added date test selector. Next action: TCC-03 shared planner and grid intent.

### TCC-03 exit evidence

- Grid Adapter exposes neutral `GridClearIntent` and synchronous capture over the
  semantic presentation, with navigation-only Delete, native-event deduplication,
  held-key suppression, pointer-gesture exclusion and retained range geometry.
  Shared consumers opt in; Backspace/editor behavior remains unchanged.
- Timeline's single planner checks full rectangular membership, exact versions,
  field capabilities, lifecycle, authority and overlapping unsubmitted drafts.
  Captured admitted draft revisions remain predecessors; newer grid/Inspector
  drafts reject the entire action. The compact accessible view-bar action after
  Find borrows focus through the existing external-action boundary.
- Clear uses existing Workbook batch admission, reservations and immutable
  attempts. Transport validates null-valued clear receipts; recovery names the
  command and keeps partial/conflict/read-debt distinctions. Comparison labels
  distinguish cleared null from empty text without changing correction values.
- `make test-slice OWNER=web.workbook
  ROWS=web.workbook.regression.timeline_clear_planning,web.workbook.regression.batch_operation_retention`
  PASS, 3/3 units, `.cartulary/test-results/20260918T133556Z-p3838837`.
  Retained-owner cases exercise both fill and clear: ordering, exact retry,
  detachment/read debt, conflict retention and authority replacement/suspension.
- `make test-slice OWNER=web.workbook
  ROWS=web.workbook.regression.workbookz_mutation_runtime_paste_adapter_a108000001`
  PASS, `.cartulary/test-results/20260918T133720Z-p3840599`.
- `make test-slice OWNER=package.grid_adapter
  ROWS=package.grid_adapter.regression.index_suite_d804ef9789,package.grid_adapter.regression.range_keyboard_entry`
  PASS, 3/3 units, `.cartulary/test-results/20260918T133816Z-p3846439`;
  both production DOM and support bindings cover clear capture/native keys.
- `make frontend-typecheck` PASS,
  `.cartulary/test-results/20260918T133742Z-p3845837`;
  `make frontend-import-boundary-check` PASS,
  `.cartulary/test-results/20260918T133818Z-p3846728`.
- Development failures: corrected one missing method delimiter and two strict
  TypeScript shape errors. The catalog rejected dynamic describe-each titles;
  equivalent explicit test titles now run both commands inside each case.
- Changed areas: `packages/grid-adapter`, Workbook transport/command/recovery,
  Timeline bulk/editing/composition/presentation, frontend source ownership,
  authored unit families and source navigation. No new persistence or dependency.
- Remaining risk: production service/browser composition and supported layout
  need fresh evidence. Next action: TCC-04 browser, lifecycle and security cases.

### TCC-04 investigation evidence

- New production rows cover reversed rectangles, held Delete, no-op action,
  native editor Delete, unsent authoring, mixed derived membership, closed
  access, offscreen membership, expanded/collapsed groups, exact lost-receipt
  retry, detached recovery, reads-only acknowledgement recovery, and partial
  null conflict resolution. Disposable fixtures only.
- Initial five-row browser slice at
  `.cartulary/test-results/20260918T134231Z-p3852018` passed four rows; the
  eligibility fixture selected non-grid capture-state metadata. Replaced its
  locator with an actual visible derived Evidence count column; corrected row
  PASS at `.cartulary/test-results/20260918T134612Z-p3913843` (11/11 units).
- Clear action accessibility PASS at
  `.cartulary/test-results/20260918T134927Z-p3953316` (11/11 units): keyboard,
  visible focus, 1440/1280/768/390 layouts, 200% zoom, text spacing.
- Expanded service coverage PASS at
  `.cartulary/test-results/20260918T134424Z-p3893572` (3/3 units): exact revision
  null, actor attribution, capture-state demotion, unchanged relationships and
  foreign target rejection without writes or version disclosure.
- Malformed clear receipt, exact retry, planner and retained owner regression
  slice PASS at `.cartulary/test-results/20260918T135040Z-p4016478` (4/4 units).
- Offscreen browser investigation at
  `.cartulary/test-results/20260918T134926Z-p3953085` exposed G7; grouped membership
  passed. Source maps located repeated per-row synchronous projection commits.
  The corrected 100-row queued-edit case, reversed rectangle and grouped case
  PASS at `.cartulary/test-results/20260918T135931Z-p4025499` (11/11 units).
  Extending the same case to still-open later authoring exposed the matching
  draft-baseline gap at `.cartulary/test-results/20260918T140155Z-p4089170`;
  remediation is implemented and awaiting rerun.
- Source coordinator regression PASS at
  `.cartulary/test-results/20260918T140136Z-p4084065` (2/2 units). Typecheck PASS
  at `.cartulary/test-results/20260918T140137Z-p4085427` (2/2 units) after
  correcting a Timeline row type at the new batch projection boundary.
- Superseded admission extension initially failed because its assertion reused
  an incident-concealment helper that requires incident_not_found. Actual
  source result was the required illegal_transition with no writes. Corrected
  that fixture assertion; rerun pending.
- A direct browser target invocation with OWNER was rejected as unsupported
  configuration before execution; narrow browser routing uses the public
  service-backed-test-slice target.

Next action: finish current acceptance reruns, then save TCC-04 DONE before
starting TCC-05. No terminal completion claim is made by this investigation log.

- G7 final browser extension PASS at
  `.cartulary/test-results/20260918T140652Z-p19631` (11/11 units), including
  both queued and still-open later drafts and prior dispatched autosave ordering.
- Superseded rejection corrected service row PASS at
  `.cartulary/test-results/20260918T140418Z-p4128757` (3/3 units).
- Concurrent service-image warm stamp maintenance raced at
  `.cartulary/test-results/20260918T140418Z-p4128744` before browser execution.
  Isolated rerun above passed; no harness or product workaround introduced.
- `make agent-finalize` PASS at
  `.cartulary/test-results/20260918T140753Z-p57048`. RESULTS_DIR unset;
  retained-run maintenance skipped. This precedes broader acceptance verification;
  TCC-05 will repeat finalization against terminal source.
- Complete `test-slice` owners (which route all owned layers, including browser
  and service rows): `web.workbook` PASS 276/276 at
  `.cartulary/test-results/20260918T140928Z-p130358`; `package.grid_adapter`
  PASS 55/55 at `.cartulary/test-results/20260918T140901Z-p90895`;
  `platform.openapi` PASS 4/4 at `.cartulary/test-results/20260918T140901Z-p90785`.
  Import boundaries PASS 2/2 at `.cartulary/test-results/20260918T140930Z-p132848`.
- Explicit 34-row Timeline production regression selection PASS 13/13 execution
  units at `.cartulary/test-results/20260918T140927Z-p128366`. Covers all new
  clear rows, range entry/selection, clipboard fidelity, fill/batch conflicts,
  replay, typing, pointer entry, creation, rejection, refresh and virtualization.
- Full Timeline owner routing at `.cartulary/test-results/20260918T140901Z-p90766`
  passed 68/72 units; remaining failures were visual comparison and blank-row
  measurement threshold. Full Workbook owner routing at
  `.cartulary/test-results/20260918T140901Z-p90813` passed 103/105 units; only
  visual comparison and its summary failed. All routed service, functional,
  stateful and accessibility rows passed. These mixed runs are not full passes.
- Narrow visual/measurement run at
  `.cartulary/test-results/20260918T140438Z-p4176052` passed 18/22 units;
  pointer focus p95 257.9 ms exceeded 100 ms while concurrent owner checks ran.
  Three other measurement rows passed. A quiet four-row rerun is pending; no
  threshold, capture, retry or tolerance was weakened.
- Ordinary full visual run `.cartulary/test-results/20260918T140848Z-p65270`
  completed all capture assertions: 252 active captures/goldens, zero orphans,
  missing goldens or ambiguous mappings, 29 registered fixtures resolved.
  Its 61 failures were screenshot comparisons, with no functional assertion
  failure. Reviewing all changed-region sheets identified the expected clear
  action and a real 1024px query-chip overlap. The action now uses its accessible
  icon-only presentation below the base band. Added 1024px accessibility coverage
  and included Find/Clear in the existing query-rail overlap assertions. Golden
  refresh remains pending the corrected narrow-layout validation.

- Quiet four-row Timeline measurement rerun PASS, 20/20 units, at
  `.cartulary/test-results/20260918T142125Z-p336199`. Pointer focus, ArrowDown,
  committed typing acknowledgement and blank-row creation all satisfy their
  unchanged adopted thresholds. Earlier threshold failures are retained as
  failed concurrent-run evidence, superseded by this quiet run.

- Closed-incident service extension and clear accessibility including 1024px
  PASS, 12/12 units, at `.cartulary/test-results/20260918T143235Z-p452788`:
  `make service-backed-test-slice OWNER=module.workbook
  ROWS=module.workbook.integration.timeline_clear_cells,module.workbook.accessibility.timeline_clear`.
- Strengthened saved-view geometry exposed text-spacing overlap in the narrow
  band at `.cartulary/test-results/20260918T143150Z-p419298`. The final G8
  correction preserves Find's base/narrow caption and uses existing icon
  actions for Timeline Inspector/Add row in the narrow band only. The existing
  saved-view accessibility row then PASS at
  `.cartulary/test-results/20260918T143442Z-p486022`; its accompanying visual
  row reached all captures but failed expected old-golden comparisons (11/13
  aggregate units). Base and narrow pressure images were visually reviewed:
  named controls and semantic chip tokens are visible without overlap.
- `make test-slice OWNER=package.protocol_ts` PASS 8/8 at
  `.cartulary/test-results/20260918T143526Z-p522085`;
  `make test-slice OWNER=web.architecture` PASS 12/12 at
  `.cartulary/test-results/20260918T143526Z-p522122`;
  `make frontend-typecheck` PASS 2/2 at
  `.cartulary/test-results/20260918T143526Z-p522241`.

### Acceptance routing and disposition

The fresh owner runs above include existing adopted regression rows; historical
handoffs establish which boundaries to exercise, not their pass status. Source,
service, and browser assertions serve different purposes. A successful browser
receipt alone is not proof of persisted SQL null or revision attribution.

| Acceptance group | Verification layer and fresh evidence | Status |
| --- | --- | --- |
| All ten direct fields; explicit null versus empty, whitespace and multiline; no-op | Timeline clear unit/admission rows; Workbook timeline_clear_cells service row, including SQL null and revision JSON assertions | PASS |
| Stable target/field order, duplicates, bounds, wrong source/view/incident, inaccessible/deleted/superseded, closed/read-only | Source admission; clear and clipboard service rows; clear_eligibility production row | PASS |
| Date pair, selected null preservation, unselected text, unrelated edits and later explicit date edit | Date source unit cases and TestTimelineClearDateIntent_Integration, including the reproduced/corrected mixed date-conflict case | PASS |
| Capture-state demotion; unchanged relationship collections; authoritative history and attribution | TestTimelineClearPreservesRelationshipsAndHistory_Integration and authoritative multi-field/multi-record service case | PASS |
| Single/reversed rectangle; keyboard/pointer selection; virtualized/offscreen and expanded membership; hidden/reordered/evicted members | Adapter index/rangeKeyboard rows; clear planner; clear_rectangle, clear_membership, clear_groups and adopted range browser regressions | PASS |
| Whole-selection rejection, unsent grid/Inspector drafts, prior autosave, later queued/open drafts | Planner registry tests; clear_drafts, clear_predecessor, clear_membership production rows | PASS |
| Unrelated-field rebase; same-field, mixed and conflicts-only outcomes; null resolution | Shared batch service regressions, clear conflict service case, clear_conflicts production row | PASS |
| Duplicate delivery, held Delete, exact uncertain retry, malformed success, late acceptance and detached recovery | Adapter tests; transport test; parameterized retained owner; clear_rectangle and clear_recovery production rows | PASS |
| Acknowledged refresh failure is reads-only; no optimistic saved null | Retained owner and transport assertions; clear_recovery production request cardinality/receipt/history assertions | PASS |
| Closure, role loss, suspension, incident retirement and account replacement conceal protected retained work | Fill/clear parameterized batch owner tests; full Workbook runtime, Find authority, query-continuation authority/session and grid-autosave browser regressions | PASS |
| Selection/focus, Inspector context and bulk checkboxes; no stale restoration or retargeting | clear_rectangle, clear_membership and clear_recovery production rows; existing range/query invalidation and recovery-navigation checks | PASS |
| Typing, click entry, Enter/Tab, F2, native Delete/Backspace, clipboard/fill, Find, creation, History/correction access and shared consumers | Complete Grid Adapter and web.workbook owners; passing functional/stateful rows in full Timeline/Workbook runs and explicit 34-row Timeline regression slice | PASS |
| Keyboard action, local impediments, non-color feedback, supported width/zoom/text spacing | timeline_clear accessibility; strengthened saved-view-query accessibility row; existing full Workbook accessibility rows | PASS |
| Paint/focus/save/creation measurements | Quiet four-row measurement run at 20260918T142125Z-p336199, unchanged thresholds | PASS |
| Reviewed visual change and full ordinary validation | Reviewed 65 images; full ordinary validation at 20260918T144241Z-p595747 passes; second ordinary run required during terminal validation | PASS |
| New persistence, dependency/migration, collection clear, deletion, other-surface adoption, exhaustive selection | Excluded by user scope; no implementation or contract adoption | N/A |

- Final clear-action accessibility rerun after the narrow action presentation
  change PASS 11/11 units at `.cartulary/test-results/20260918T143741Z-p561863`:
  `make service-backed-test-slice OWNER=module.workbook
  ROWS=module.workbook.accessibility.timeline_clear`.
- Interim scope/whitespace review PASS. No lockfiles, migrations, digest files or
  immutable released baseline changed. The authored 2.0.0 candidate change-set
  records the additive clear command and description-only compatibility change.

### Reviewed golden refresh

Trigger: adopted Timeline Clear contents action and narrow action presentation,
after G8 geometry correction. `make browser-e2e-visual-update` PASS 12/12 at
`.cartulary/test-results/20260918T143621Z-p528307`. The full prior ordinary
reconciliation was reviewed before mutation; update promotion passed all functional
assertions and accounted for all 252 captures/goldens and 29 registered fixtures.
65 images changed; all were inspected as full-frame thumbnails plus detailed
changed regions, with full-size zoom/recovery/pressure review. Substantial changes
are confined to the Timeline view bar, including its banner-shifted and zoomed
positions. No unexplained grid, inspector, typography, focus or overflow change
was accepted. All 65 mappings below come from the retained reconciliation and
authored catalog, not filename-inferred ownership.

Viewport, zoom, density, masks, scroll normalization, screenshot scope, renderer,
fonts, tolerances and registry identities are unchanged. Only failing comparisons
were promoted by the Make-owned candidate workflow; no PNG was hand-edited.
No golden was deleted. Ordinary verification follows separately. These are
implementation/design regression artifacts, not Core 05 conformance claims.

All filenames in the following table are under
`apps/web/e2e/workbook.visual.spec.ts-snapshots/`. A dash means the active capture
has no registered fixture; its exact catalog/scenario/project identity reconciles.

| Changed golden | Authored semantic owner row | Registered fixture IDs |
| --- | --- | --- |
| `account-menu-controls-narrow-linux.png` | `web.design.visual.account_menu_root_and_nested_viewport_states_cef60727cc` | — |
| `account-menu-workbook-root-linux.png` | `web.design.visual.account_menu_root_and_nested_viewport_states_cef60727cc` | — |
| `collaboration-conflict-resolver-linux.png` | `module.collaboration.visual.capture_presence_markers_same_field_conflict_res_f0a62c52a1` | `visual.fixture.same_field_conflict` |
| `collaboration-grid-blocked-conflict-linux.png` | `module.collaboration.visual.the_visual_harness_asserts_syncing_same_field_co_df11cd99bc` | — |
| `collaboration-grid-conflict-resolver-compact-linux.png` | `module.collaboration.visual.the_visual_harness_asserts_same_field_conflict_m_c472bd3f9c` | — |
| `collaboration-grid-conflict-resolver-linux.png` | `module.collaboration.visual.the_visual_harness_asserts_same_field_conflict_m_c472bd3f9c` | — |
| `collaboration-grid-conflict-resolver-narrow-linux.png` | `module.collaboration.visual.the_visual_harness_asserts_same_field_conflict_m_c472bd3f9c` | — |
| `collaboration-presence-markers-linux.png` | `module.collaboration.visual.capture_presence_markers_same_field_conflict_res_f0a62c52a1` | `visual.fixture.presence_overflow` |
| `coordination-comm-log-authoring-linux.png` | `module.workbook.visual.coordination_create_authoring_recovery` | `visual.fixture.contextual_coordination_creation` |
| `coordination-handoff-authoring-linux.png` | `module.workbook.visual.coordination_create_authoring_recovery` | `visual.fixture.contextual_coordination_creation` |
| `coordination-lesson-authoring-linux.png` | `module.workbook.visual.coordination_create_authoring_recovery` | `visual.fixture.contextual_coordination_creation` |
| `coordination-recovery-linux.png` | `module.workbook.visual.coordination_create_authoring_recovery` | `visual.fixture.contextual_coordination_creation` |
| `coordination-recovery-narrow-linux.png` | `module.workbook.visual.coordination_create_authoring_recovery` | `visual.fixture.contextual_coordination_creation` |
| `coordination-source-narrow-linux.png` | `module.workbook.visual.coordination_create_authoring_recovery` | `visual.fixture.contextual_coordination_creation` |
| `coordination-status-review-authoring-linux.png` | `module.workbook.visual.coordination_create_authoring_recovery` | `visual.fixture.contextual_coordination_creation` |
| `entity-mention-chip-states-linux.png` | `module.entities.visual.capture_unresolved_token_resolved_chip_auto_reso_d3b74bd9d7` | `visual.fixture.mention_chip_state_matrix` |
| `incident-directory-compact-desktop-workbook-shell-linux.png` | `module.workbook.visual.capture_default_timeline_workbook_shell_with_vie_c06bbcbee0` | `visual.fixture.compact_desktop_workbook_shell` |
| `incident-directory-default-timeline-workbook-shell-linux.png` | `module.workbook.visual.capture_default_timeline_workbook_shell_with_vie_c06bbcbee0` | `visual.fixture.default_timeline_workbook_shell` |
| `incident-directory-narrow-desktop-workbook-shell-linux.png` | `module.workbook.visual.capture_default_timeline_workbook_shell_with_vie_c06bbcbee0` | `visual.fixture.narrow_desktop_workbook_shell` |
| `indicator-observation-authoring-linux.png` | `module.workbook.visual.indicator_observations_authoring` | `visual.fixture.indicator_observations_authoring` |
| `indicator-observation-authoring-narrow-linux.png` | `module.workbook.visual.indicator_observations_authoring` | `visual.fixture.indicator_observations_authoring` |
| `lifecycle-closed-linux.png` | `web.design.visual.lifecycle` | — |
| `lifecycle-confirmed-refresh-failure-linux.png` | `web.design.visual.lifecycle` | — |
| `lifecycle-pending-linux.png` | `web.design.visual.lifecycle` | — |
| `lifecycle-uncertain-linux.png` | `web.design.visual.lifecycle` | — |
| `linked-note-authoring-linux.png` | `module.workbook.visual.note_create_authoring_recovery` | — |
| `linked-note-authoring-narrow-linux.png` | `module.workbook.visual.note_create_authoring_recovery` | — |
| `linked-note-recovery-linux.png` | `module.workbook.visual.note_create_authoring_recovery` | — |
| `linked-note-recovery-narrow-linux.png` | `module.workbook.visual.note_create_authoring_recovery` | — |
| `linked-note-source-narrow-linux.png` | `module.workbook.visual.note_create_authoring_recovery` | — |
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

- First fresh ordinary `make browser-e2e-visual` PASS 12/12 at
  `.cartulary/test-results/20260918T144241Z-p595747`, against the promoted manifest.
- `make lint-biome` initially failed three noNonNullAssertion warnings in new
  test code at `.cartulary/test-results/20260918T144342Z-p630050`. Replaced the
  assertions with explicit narrowing/fixture guards. Corrected lint PASS 2/2
  at `.cartulary/test-results/20260918T144507Z-p635401`; planner slice PASS 2/2
  at `.cartulary/test-results/20260918T144507Z-p635267`; strengthened saved-view
  accessibility row PASS 11/11 at
  `.cartulary/test-results/20260918T144507Z-p635282`.
- Added task-guide discovery for module.collaboration and module.entities after
  reconciliation identified their affected Timeline-shell golden inputs. Their
  mapped visual rows pass in the full ordinary run; no source-domain behavior
  from those owners is changed.

- G9 confirmed by an added authoritative service assertion at
  `.cartulary/test-results/20260918T144714Z-p667870`: after clearing local date,
  a stale two-date paste reported its local conflict but regenerated local to
  `2026-09-18T08:00:00-05:00` instead of preserving null. The batch executor now
  applies only accepted cells while passing the complete originally submitted
  field set to date conversion. No request, hash, schema or merge policy changed.
  Source/admission slice PASS 2/2 at
  `.cartulary/test-results/20260918T144839Z-p690279`; service rerun pending.

- G9 corrected authoritative clear plus existing clipboard/fill/tag service
  selection PASS 3/3 at `.cartulary/test-results/20260918T144838Z-p690056`.
  The mixed date case now accepts UTC only, retains saved local null and reports
  its actionable conflict. Full final-source quiet measurements are running
  before closing TCC-04. No applicable blocker is being waived.

### Changed paths and retired implementation

The following authored and generated paths accompany the 65 golden files listed
above. Generated public contracts, topology and golden manifest were updated
only through their Make owners. Source navigation READMEs describe the new clear
planner and accepted-batch predecessor/projection responsibilities.

```text
apps/web/e2e/timeline-grid-entry.spec.ts
apps/web/e2e/timeline-range-selection.spec.ts
apps/web/e2e/workbook.a11y.spec.ts
apps/web/src/workbook/adapters/createWorkbookBatchTransport.ts
apps/web/src/workbook/adapters/createWorkbookClipboardPasteAdapter.test.ts
apps/web/src/workbook/components/WorkbookBatchRecovery.tsx
apps/web/src/workbook/components/WorkbookSameFieldConflictResolver.tsx
apps/web/src/workbook/components/WorkbookViewBar.tsx
apps/web/src/workbook/mutations/createWorkbookMutationCommandPorts.ts
apps/web/src/workbook/mutations/workbookMutationCommandPorts.ts
apps/web/src/workbook/runtime/README.md
apps/web/src/workbook/runtime/WorkbookBatchOperationOwner.test.ts
apps/web/src/workbook/runtime/WorkbookMutationDriverRegistry.ts
apps/web/src/workbook/runtime/WorkbookMutationRuntime.ts
apps/web/src/workbook/runtime/workbookBatchRecoveryItems.ts
apps/web/src/workbook/timeline/bulk/README.md
apps/web/src/workbook/timeline/bulk/useTimelineClearController.test.tsx
apps/web/src/workbook/timeline/bulk/useTimelineClearController.ts
apps/web/src/workbook/timeline/components/TimelineWorkbookGrid.tsx
apps/web/src/workbook/timeline/composition/useTimelineInteractionComposition.ts
apps/web/src/workbook/timeline/composition/useTimelineMutationComposition.ts
apps/web/src/workbook/timeline/composition/useTimelineWorkbookComposition.ts
apps/web/src/workbook/timeline/editing/useTimelineEditorDraftRegistry.ts
apps/web/src/workbook/timeline/hooks/useTimelineMutationRuntimeBindings.ts
apps/web/src/workbook/timeline/mutations/README.md
apps/web/src/workbook/timeline/mutations/WorkbookTimelineMutationOwner.ts
apps/web/src/workbook/timeline/mutations/createTimelineMutationDriver.ts
apps/web/src/workbook/timeline/mutations/useTimelineRowMutationCoordinator.ts
apps/web/src/workbook/timeline/presentation/TimelineWorkbookViewBarRegion.tsx
apps/web/src/workbook/timeline/presentation/useTimelineWorkbookPresentation.tsx
apps/web/src/workbook/timeline/useTimelineMutationRuntimeBindings.test.tsx
contracts/openapi-releases/2.0.0.change-set.json
contracts/openapi-source/owners/module.workbook/openapi.json
contracts/openapi/cartulary.openapi.yaml
docs/design.md
docs/handoffs/ui-ux/workbook-timeline-clear-contents-refactor-handoff.md
docs/spec/01_architecture_storage_and_view_contracts.md
docs/spec/03_workbook_interaction_collaboration_and_workflows.md
internal/app/workbookassembly/action_adapters.go
internal/app/workbookassembly/timeline_adapters.go
internal/app/workbookassembly/timeline_capabilities_test.go
internal/gen/contractopenapi/artifacts_gen.go
internal/gen/openapioperations/catalog_gen.go
internal/modules/timeline/admission/batch.go
internal/modules/timeline/admission/clear_cells_test.go
internal/modules/timeline/batch_mutation_store.go
internal/modules/timeline/clear_cells.go
internal/modules/timeline/clear_cells_test.go
internal/modules/timeline/clipboard_paste.go
internal/modules/timeline/commands.go
internal/modules/timeline/facade.go
internal/modules/timeline/mutationpolicy/policy.go
internal/modules/timeline/performance_fixture_store.go
internal/modules/timeline/store_patch.go
internal/modules/timeline/support_test.go
internal/modules/timeline/time_conversion_store.go
internal/modules/workbook/clear_cells_integration_test.go
internal/modules/workbook/clipboard_paste_integration_test.go
packages/grid-adapter/README.md
packages/grid-adapter/src/SemanticDataGrid.tsx
packages/grid-adapter/src/core.ts
packages/grid-adapter/src/index.test.tsx
packages/grid-adapter/src/index.tsx
packages/grid-adapter/src/rangeKeyboard.test.ts
packages/grid-adapter/src/semanticCapabilities.ts
packages/grid-adapter/src/semanticClear.ts
packages/grid-adapter/src/test-support.tsx
packages/protocol-ts/src/generated/core-http-types.ts
packages/protocol-ts/src/generated/core-http-validators.ts
tools/browser_e2e_batch_manifest.json
tools/execution_topology_render_index.json
tools/frontend_source_ownership.json
tools/frontend_visual_golden_manifest.json
tools/test_families/module.timeline.json
tools/test_families/module.workbook.json
tools/test_families/package.grid_adapter.json
tools/test_families/web.workbook.json
```

Retired path: the unused string-only `ownerBatchCellV1.Value` member and its
constructors. Nullable `PatchChange` is the sole batch-cell value authority.
No route, persistence table, dependency or consumer command was removed.
The UI action uses one Timeline planner; both entry points use the retained
Workbook batch executor. Grid Adapter adds an opt-in neutral intent only.

Limitations remain the adopted slice: current visible committed membership,
500 records by ten direct fields, existing in-memory retained lifetime, and
no reload persistence, Undo/Redo, collection clear, row deletion, hidden-member
selection or automatic page/group loading. Full release/CI and unrelated owner
suites are outside this seam; disposable fixtures supplied all writes.

- Quiet measurement rerun after the final date/rendering changes PASS 20/20 at
  `.cartulary/test-results/20260918T145001Z-p709188` with all four adopted
  thresholds unchanged. The subsequent G10 change only threads action-event
  identity; terminal validation will repeat measurements on the final source.
- G10 closes a delivery-identity gap identified during final review: the toolbar
  callback previously created a fresh opaque identity. It now passes the native
  click to the existing batch owner. The rectangle production scenario retains
  keyboard accessibility coverage and adds duplicate-click request-cardinality
  assertions. No timeout suppression or second execution path was introduced.

### TCC-04 exit

G10 rectangle/held-key/keyboard-action/duplicate-click production scenario PASS
11/11 units at `.cartulary/test-results/20260918T145609Z-p754718`.
Final frontend typecheck PASS 2/2 at
`.cartulary/test-results/20260918T145609Z-p754793`; Biome PASS 2/2 at
`.cartulary/test-results/20260918T145609Z-p754813`.

Every applicable acceptance group above passes at its responsible layer. G1–G10
are remediated with binary evidence; exclusions have explicit scope rationales.
All failed investigations have retained roots and a passing corrective disposition.
No applicable BLOCKED dependency remains. Behavioral source and public contracts
are complete. Next action: save TCC-05 IN_PROGRESS, finalize harness maintenance,
then terminal drift/boundary/lint checks, second ordinary visual pass and quiet
measurements on the final source. No commit, deployment or analyst write follows.

### TCC-05 terminal evidence

- `env -u RESULTS_DIR make agent-finalize` PASS 1/1 at
  `.cartulary/test-results/20260918T145747Z-p787177`; zero generated changes.
  RESULTS_DIR was unset. Retained-run canonical evidence, scheduler timing/order
  maintenance and retained performance-evidence maintenance were skipped with
  reason `results-dir-not-provided`; no successful full warm check was claimed.
- `make generate-drift` PASS 4/4 at
  `.cartulary/test-results/20260918T145836Z-p791158`.
- `make generated-artifact-policy-check` PASS 3/3 at
  `.cartulary/test-results/20260918T145836Z-p791138`;
  `make json-shape-check` PASS 3/3 at
  `.cartulary/test-results/20260918T145836Z-p791146`;
  `make test-catalog-check` PASS in the same public multi-target invocation.
- `make openapi-compatibility-check` PASS 4/4 at
  `.cartulary/test-results/20260918T145836Z-p791176`;
  `make frontend-import-boundary-check` PASS 2/2 at
  `.cartulary/test-results/20260918T145836Z-p791436`;
  `make backend-module-boundary-check` PASS 3/3 at
  `.cartulary/test-results/20260918T145836Z-p791447`.
- `make lint` PASS 11/11 at
  `.cartulary/test-results/20260918T145836Z-p791596`, covering authored Go,
  frontend, scripts and shell checks selected by the public target.
- `make test-slice OWNER=harness.test_catalog` failed aggregate graph
  determinism at `.cartulary/test-results/20260918T145836Z-p791278` (59/60
  internal cases passed). Both baseline and current browser manifests serialize
  the same webserver-backed group mapping to 4,282 characters, beyond the existing
  4,096-character argument bound. G11 replaces that one CSV argument with the
  finalizer's already-supported repeated mapping flag; limits, exact row
  selection and finalizer semantics remain unchanged. This narrow verification
  repair changes one additional authored path:
  `tools/harness/scheduler/work-graph/browser.mjs`.
- A task-guide lookup for `harness.scheduler` failed because that owner does not
  exist. Routing is owned by harness.test_catalog/command_surface and browser;
  the actual owner guide and adopted Testing Harness mechanism/selection boundary
  were read. No specification amendment is needed for an already-supported
  private argument representation. Corrective validation is pending.

- Second fresh ordinary `make browser-e2e-visual` PASS 12/12 at
  `.cartulary/test-results/20260918T145836Z-p791534`. Both required ordinary
  passes use the promoted manifest. G10 does not change rendering; all captures
  and functional assertions pass on its source.
- G11's first corrective catalog rerun at
  `.cartulary/test-results/20260918T150247Z-p849552` passed aggregate graph
  determinism, then failed the expected stale generated-input fingerprint check.
  `make generate` PASS at `.cartulary/test-results/20260918T150401Z-p853210`
  refreshed that fingerprint through the owner. Finalization and affected
  harness checks are being repeated on this final source.

- Repeated `env -u RESULTS_DIR make agent-finalize` PASS 1/1 at
  `.cartulary/test-results/20260918T150501Z-p856385`. Retained-run maintenance
  remains skipped because RESULTS_DIR is unset; current structural generation
  remains checked. This precedes the final affected harness verification.
- G11 final `make test-slice OWNER=harness.test_catalog` PASS 1/1 (all internal
  contract cases) at `.cartulary/test-results/20260918T150628Z-p860570`;
  `make test-slice OWNER=harness.command_surface` PASS 1/1 at
  `.cartulary/test-results/20260918T150628Z-p860520`.
- Final `make generate-drift` PASS 4/4 at
  `.cartulary/test-results/20260918T150628Z-p860448`;
  `make lint-scripts` PASS 2/2 at
  `.cartulary/test-results/20260918T150628Z-p860753`;
  `make generated-artifact-policy-check` PASS 3/3 at
  `.cartulary/test-results/20260918T150628Z-p860481`;
  `make json-shape-check` PASS 3/3 at
  `.cartulary/test-results/20260918T150628Z-p860490`.

The baseline routing defect is resolved without widening harness limits or
changing the public task surface. G11 may remain independently if the clear UI
is rolled back; any reversal of that repair must regenerate its topology source
fingerprint through Make. It has no analyst-data or historical-receipt effect.

Final scope review: 143 changed paths comprise the 65 reviewed goldens and 78
other authored/generated paths, including the separately recorded G11 repair.
Branch remains main at d6dfeabbe63d5d598308786a4629fbae9bb50539. No pre-existing
edits were overwritten. No file was deleted. No migrations, dependencies,
lockfiles, digest, configs, binary roots or immutable released baseline changed.
No commits, pushes, deployment, local analyst-database work or external messages
were performed. Current results use disposable harness fixtures only.

- Final-source quiet four-row `make service-backed-test-slice` measurement
  selection PASS 20/20 at `.cartulary/test-results/20260918T150834Z-p871410`.
  The four row IDs are the committed typing-acknowledgement, blank-row creation,
  ArrowDown-selection and Enter-focus rows recorded above. Thresholds and
  profiles remain unchanged. This also exercises G11's repeated mapping flags
  in real measurement target finalization with exact group accounting.

### Terminal disposition

All product acceptance groups and executable terminal gates are PASS. G1–G11
are remediated. No applicable BLOCKED dependency or unresolved semantic decision
remains. Public additions and generated consumers agree; old paste/fill/tag
meanings and hashes remain supported, and old servers reject clear without a
fallback. Recovery retains accepted writes, attributable history, unresolved
conflicts and exact uncertain attempts within the existing owner lifetime.

Retained-run maintenance was intentionally skipped because RESULTS_DIR is unset.
Full release/CI, unrelated owners, deployment and migration exercises are N/A to
this seam; affected backend/API/frontend/security/service/browser/accessibility/
visual/measurement and boundary checks are evidenced above. No previously failed
run was relabeled as a pass. The first measurement and layout failures, the G7/G9
product defects, lint warnings and G11 baseline defect retain their actual roots
and corrective dispositions.

Final documentation and whitespace review PASS. `make lint-markdown` passed at
`.cartulary/test-results/20260918T151421Z-p912330`; `git diff --check` and final
scope review passed. TCC-05 is DONE. The completed handoff records all adopted
decisions, actual failures and corrective evidence, changed paths, compatibility,
skipped maintenance, limitations and coherent rollback. No applicable acceptance
row is BLOCKED. Earlier pending investigation notes are superseded by the named
passing dispositions and completed workstream exits.

Next action: stop at this seam. The implementation is uncommitted for review;
no commit, push, deployment or analyst-data action follows. Rollback must preserve
accepted analyst writes and historical receipts as specified above.
