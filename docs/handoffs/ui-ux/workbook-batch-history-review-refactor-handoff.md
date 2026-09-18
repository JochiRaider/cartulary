# Timeline batch-result review and exact History navigation

## Execution control

Implementation baseline revalidated: branch `main`, HEAD
`cd03bf6a2d53d884c667de344c6af16e3a7567a6`, clean working tree, no existing
edits, zero commits behind and one ahead of `origin/main`. Only root AGENTS.md
applies. Preserve that commit. Scope is Timeline clear, clipboard table paste,
fill-down and multi-row tag review through existing Recovery and History owners,
their narrow Core 03/design amendments, and necessary verification/support.
No digest changes, new routes, mutation semantics, browser persistence,
dependencies, analyst-data operations, commits, pushes or deployments.

| Workstream | Status | Exit / next action |
| --- | --- | --- |
| BHR-01 Receipt and review contract | DONE | Owner matrices and gap dispositions adopted; documentation checks pass. |
| BHR-02 Cohesive outcome presentation | DONE | Outcome matrix, bounded chooser and retained-batch regression pass; types pass. |
| BHR-03 Exact History navigation | DONE | Exact current History context, bounded lifetime and existing reversal pass. |
| BHR-04 Integrated evidence | DONE | Applicable behavioral, security and accessibility criteria pass. |
| BHR-05 Final validation | DONE | Final source and completed handoff covered by required checks. |

Only the current row is IN_PROGRESS. Save actual exit evidence and DONE before
starting its successor. Applicable BLOCKED dependencies prevent dependent work
and completion. Historical results are neither fresh passes nor fresh failures.

## Authority and investigation

Behavior: Core 01 §3.3.4.2, §3.3.5, §3.3.5.0 and REQ-01-672; Core 02 §15;
Core 03 §§2.3A,3–4,10,11.1,13–14 and REQ-03-100/285/299; Core 04 §§1–2 and
REQ-04-127; design §§7.3–7.4,10,14. Domain vocabulary unchanged. The NLSpec
research essay, localized digest and historical handoffs remain advisory.
Source placement follows frontend ownership/import manifests; verification
routing follows the owner catalog and authored test families. No executable
consumer may depend on Markdown.

Planning followed the localized README read order and inspected the completed
clear, paste/bulk, History browsing, History mutation recovery, Recovery navigation
and query-continuation handoffs. Implementation rechecked root instructions,
branch, HEAD and dirty state before editing. Main source trace:

- Timeline clear/fill controllers, clipboard planner/controller and bulk-tag
  planner/controller capture semantic targets; adapters admit retained batches.
- `runtime/WorkbookBatchOperationOwner` captures immutable attempts, retains
  outcomes/conflicts/read obligations, and prunes older completed receipts per
  surface. `adapters/createWorkbookBatchTransport` validates complete receipts.
- `runtime/workbookBatchRecoveryItems` and `components/WorkbookBatchRecovery`
  currently duplicate labels and classification; requested existing-record
  counts omit paste creation slots and are not explicitly labeled requested.
- `history/HistoryActionLookup`, `workbookHistoryBrowsing`, the shared Inspector
  History controller/presentation and `WorkbookRecordHistoryOwner` own reads,
  legal-action review and captured mutations. Recovery already reuses that
  presentation, independently from the canonical Inspector subject.
- Timeline source batches omit unchanged targets and append source-owned
  mutation entries, revisions, projections and immutable replay receipts.
  Receipt rows do not inventory every related record affected by a change set.
- Shell/runtime History reconciliation retains acknowledgement before reads and
  source refresh. Recovery navigation currently detaches a removed item, so
  receipt pruning requires a minimal active review contribution.

## Adopted matrices

| Operation | Requested scope | Confirmed receipt claims |
| --- | --- | --- |
| Clear contents | Explicit target records and eligible fields | Each returned materially changed Timeline record once; no changed-cell count. |
| Clipboard table paste | Existing-record targets plus creation slots | Returned Timeline records and ordered conflicts; returned IDs support review, not complete change-set scope. |
| Fill-down | Recipient records and one field/value, excluding source | Returned Timeline records and ordered conflicts; no inferred changed-cell count. |
| Tag assignment | Explicit target records and tag label | Returned Timeline records and collection conflicts; no inferred tag-mutation count. |
| Hosts/Identities regression | Existing entity-origin request semantics | Duplicate returned identities remain legal where source reuse allows them; no new review entry. |

The following outcome matrix applies to every in-scope operation. Original
conflicts belong to the immutable receipt; unresolved conflicts come from the
current conflict store. Resolving a conflict creates separate attributed history.

| Outcome facts | Feedback and actions |
| --- | --- |
| Waiting | Awaiting earlier work; existing discard admission. |
| Preparing/submitting | Applying; no completion or review implied. |
| Uncertain | Outcome uncertain; current-authority exact retry only. |
| Definitive rejection | Rejected; safe failure/correction context and existing discard. A denied uncertain replay does not prove rejection. |
| Material receipt, no unresolved conflicts | Saved changes; returned-record choices permit read review. |
| Material receipt plus unresolved conflicts | Saved changes and current conflicts independently visible; review and existing conflict access. |
| Empty receipt with original conflicts | No batch changes saved; original and remaining conflicts distinguished even after all are resolved. No change-set navigation. |
| Empty receipt without conflicts | Acknowledged no-op. No change-set navigation. |
| Any acknowledgement with incomplete reads | Preserve outcome; independently show refreshing/refresh required and read-only retry. |

| Subject/lifetime | Required behavior |
| --- | --- |
| Passive completion | No panel opening, selection replacement, focus movement or typing interruption. |
| Explicit review | One selected returned Timeline record and fixed requested change-set identity; no write until explicit existing History action. |
| Inspector | Preserve canonical subject/drafts while Recovery detaches its presentation. |
| Filtered/evicted record | Read record History directly; no query pivot, page expansion or receipt-as-current-state assumption. |
| Deleted/unavailable record | Current History supplies deletion/version/actions; unavailable History clears protected read content without inventing deletion or incident revocation. |
| Lookup | Three reads per activation, existing 30-second observation bound; Continue/Retry/Restart/Cancel; paused is not absence. |
| Browsing | At most three response pages and bounded continuation bookkeeping for active batch review; discarded earlier content requires restart. |
| Newer completion/pruning | Keep only active locator, small navigation/label context and bounded History state. Receipt and chooser inventory are not copied or pinned. |
| Close/detail switch/panel detachment | Cancel disposable reads and release review state; retain admitted mutations in their existing owner. Restore semantic focus only if still eligible. |
| Closure/demotion with read access | Invalidate confirmation; keep authorized reading and current operation-specific action gates. |
| Suspension/revocation/account replacement | Conceal disposable review; retire it on incident/account replacement. Existing owners retain required same-account unsaved work and attempts. |

Record choices use local search and pages of 20 compact historical receipt labels.
Only an explicit Review this change activation reads History. The locator contains
read authority scope, source/view, record and change-set identities, never a
rollback selector or current version. All matching loaded logical items are
identified; a match does not imply complete loaded change-set membership.
Whole-change-set reversal includes all reversible entries and can affect records
beyond the selected row and returned receipt list. Existing confirmations,
reviewer/admin rollback permission, concurrency, exact replay and source
reconciliation remain authoritative. Reversal does not rewrite original batch
receipts or clear their unresolved conflicts.

## Gap ledger

| Gap / classification | Remediation / affected areas | Rationale / long-term benefit | Compatibility / unresolved risk | Binary validation |
| --- | --- | --- | --- | --- |
| G1 Confirmed wording weakness | One typed outcome/requested-scope projection for Recovery list/detail. | Facts have one presentation owner; creation targets cannot be omitted. | Internal presentation change; no wire migration. Risk: unsupported counts. | Every operation/outcome reports only requested or contract-proven returned scope. |
| G2 Confirmed ambiguity | Separate material/no-op/original/current conflicts and read debt in projection/tests. | Acknowledgement cannot be erased by read failure or later correction. | Existing save-state and conflict ownership retained. Risk: conflicts-only described as no-op. | All outcome matrix rows pass; read retry sends zero writes. |
| G3 Authorized enhancement | Add bounded Timeline record choices and explicit review entry in Recovery. | Exact inspection stays local without a global activity surface. | Other consumers unchanged; no persistence. Risk: eager reads or large DOM. | At most 20 choices rendered; zero History reads before record activation. |
| G4 Confirmed lifetime mismatch | Active minimal locator contributes existing navigation identity after receipt pruning. | Review survives completion without pinning payloads or a second archive. | Existing pruning preserved. Risk: silently replacing subject or retaining receipts. | Newer completion prunes receipt while current subject/change set remain fixed; close releases locator. |
| G5 Shared lookup extension | Extract neutral paging mechanics; add read change-set target distinct from action validation. | One cancellation/validation owner prevents scanner drift. | Existing action lookup contract preserved. Risk: arbitrary-entry selection, incomplete search called absence. | Later-page/multiple-entry/budget/cursor/cancel/stale tests pass with exact identities. |
| G6 Authorized subject reconciliation | Narrow Core 03/design amendments and shared History locator mode. | Off-window review no longer depends on grid/Inspector current state. | Canonical Inspector consumers and backend contracts retained. Risk: draft submission or current version inferred from receipt. | Direct History reads, current action metadata and preserved drafts/selection/focus. |
| G7 Confirmed unbounded review risk | Optional three-page retention in shared browsing and bounded lookup bookkeeping. | Long review does not accumulate full history or hidden receipt copies. | Existing Inspector browsing behavior retained; review policy only. Risk: lost continuation or stale action proof. | Working set bounded across explicit continuation; discarded pages require restart; action admission remains current. |
| G8 Behavioral/security evidence | Focused owner, service, production browser, accessibility and visual/measurement rows. | Establish correct behavior at each responsible layer. | No historical run promoted to feature evidence. Risk: authority/focus/multi-record regression. | Every applicable acceptance row PASS; N/A justified; no applicable BLOCKED. |

## Evidence log

- Planning baseline only: four requested task guides passed; selected batch
  retention, History lookup/browsing and Recovery navigation tests passed 5/5
  execution units at `.cartulary/test-results/20260918T154606Z-p924679`.
- Implementation entry: branch/HEAD/status/instruction inspection passed and
  exactly matched the clean planning baseline. No pre-existing edits.

## Compatibility, rollback and next action

No public route, schema, mutation, stored-data or dependency migration is planned.
Remove redundant presentation paths only after their callers migrate. Rollback
must coordinate owner wording, source, tests and generated routing. Never remove
accepted writes, reversal revisions, replay receipts or history. No global
Undo/Redo, PATCH-based reversal or new reverse-batch command is introduced.

### BHR-01 exit

Changed this handoff, Core 03 §10.2 and design §7.4. The operation/outcome/subject
matrices resolve count, scope, permission and retention assumptions. The current
lookup budget is retained; the new three-page review working set is explicitly
adopted without changing existing Inspector consumers. Domain vocabulary unchanged.
The requested four task guides passed again at implementation entry.
`make lint-markdown` PASS at
`.cartulary/test-results/20260918T170456Z-p941483`; `git diff --check` PASS.
The handoff's matrix/heading structure was also inspected directly (the broad
Markdown target's configured handoff globs are selective). No applicable blocked
dependency. Next: BHR-02 typed outcome projection and compact record choices.

### BHR-02 exit

Added `runtime/workbookBatchOutcome.ts` and focused tests; migrated Recovery
item/detail wording to that one projection. Removed duplicated kind labels,
ambiguous no-op/conflicts-only wording and unqualified requested-record counts.
Added `components/WorkbookBatchRecordChoices.tsx` and tests for local search,
twenty-choice pages and explicit activation. BHR-03 connects its read callback.
Updated authored source ownership, test-family rows and local source guides.

`make format` PASS at `.cartulary/test-results/20260918T171107Z-p945361`.
The first slice/typecheck identified new fixture issues (unsupported matcher,
required nonempty tuple and session identity); corrected, no product regression.
Failure roots: `20260918T171129Z-p949829` and `20260918T171129Z-p949925` under
`.cartulary/test-results`. Rerun `make test-slice OWNER=web.workbook
ROWS=web.workbook.regression.batch_outcome_facts,web.workbook.regression.batch_record_choices,web.workbook.regression.batch_operation_retention`
PASS 4/4 at `.cartulary/test-results/20260918T171229Z-p951402`.
`make frontend-typecheck` PASS 2/2 at
`.cartulary/test-results/20260918T171229Z-p951480`.

All four operation/outcome classifications, original/current conflicts, requested
creation targets, Entity reuse regression, completed grouping and concealment
are covered. Existing retained owner tests cover acknowledged read-only retry.
No public contract or retention change. Remaining integration/focus/security
risks belong to BHR-03/04. Next: exact History lookup and review lifetime.

### BHR-03 exit

Shared `HistoryPageLookup` now owns scanning mechanics; action lookup supplies its
separate selector/admission rule. Recovery uses a minimal authorized locator,
current History-derived subject and shared controller/presentation. All matching
loaded entries are marked; lookup retains the three-read/30-second budget and
review browsing evicts beyond three pages. Current metadata and the existing
History owner still authorize and execute separately confirmed reversal.
Whole-change-set confirmation identifies effects beyond the displayed record.
Receipt pruning preserves only the active navigation descriptor/locator. Closing
or detachment releases it. Recovery activation preserves unsubmitted grid drafts.

Fresh evidence: `make format` PASS at
`.cartulary/test-results/20260918T172656Z-p967653`; focused batch review/action
lookup/browsing slice PASS 4/4 at
`.cartulary/test-results/20260918T172717Z-p972028`; `make frontend-typecheck` PASS
2/2 at `.cartulary/test-results/20260918T172717Z-p972109`; existing History
controller/transport/owner/recovery/coordination and navigation slice PASS 10/10
at `.cartulary/test-results/20260918T172745Z-p973310`.
The focused review demonstrates explicit later-page/off-window reads, multiple
matching entries, editor action gates, bounded retention, pruning/new completion
without focus movement, and one confirmed mutation through the existing owner.

Resolved validation failures: a redundant effect dependency failed formatting
at `20260918T171832Z-p954036` and `20260918T172547Z-p961359`; intermediate lint
failed at `20260918T172436Z-p960541`. Test fixture paging/mock types and an
ambiguous Completed selector failed types/slice at `20260918T172614Z-p966289`
and `20260918T172614Z-p966225`; fixed and rerun above. Diagnostic Make overrides
were rejected as unsupported inputs and were removed. No routing policy changed
for diagnostics. No blocker remains for BHR-04; next: integrated production
browser/service, authority and failure evidence.

### BHR-04 evidence and dispositions

The new routed `batch_history_review` frontend row uses the production Recovery
attachment and shared History controller. It covers explicit off-window reads,
multiple matching entries, three-page retention, failed/expired-page continuation,
cancellation/late completion, current deletion/absence, pruning/newer completion,
editor gates, closure/demotion, History suspension, scoped revocation and account
replacement. Existing owner/transport rows retain exact replay, late acknowledgement,
confirmation/version fencing and refresh-only recovery. No second mutation owner
or receipt archive was introduced.

The four new `module.timeline.browser.batch_history_review_*` rows exercise
production batch controls and live service History/rollback. Clear searches past
100 newer entries while its row is filtered out and preserves the query. Current
metadata offers row restore but no change-set reversal for that older clear;
the UI correctly omits the unavailable action. Paste/fill/tag use exact change-set
selectors and current versions, append attributed reversal entries, preserve
original batch entries, and restore all returned targets. Tag History has multiple
logical entries under one change set and each is marked. Paste review preserves an
unrelated unsubmitted grid draft. Keyboard activation, close focus, narrow geometry
and text spacing are asserted. Desktop captures are retained in the browser report;
visual review led to removing redundant detail headings/scope and combining technical
identifiers in the shared History disclosure.

Existing coordination deliberately blocks History action preparation while the
runtime has unresolved conflicts. Read-only review still works, and the existing
Review conflicts control remains available. This gate is preserved rather than
bypassed by the receipt locator. A focused partial-success test asserts that neither
receipt nor unresolved conflicts are changed and no reversal is submitted. This
is a compatibility disposition, not additional rollback authority.

Fresh service/owner evidence (all roots under `.cartulary/test-results`):

| Public Make selection | Result / run root |
| --- | --- |
| `service-backed-test-slice OWNER=module.workbook`, shared-ingest batch and Timeline clear integration rows | PASS 3/3, `20260918T173243Z-p982626` |
| `service-backed-test-slice OWNER=module.revisions`, atomic consequences and History pagination integration rows | PASS 3/3, `20260918T173243Z-p982637` |
| `test-slice OWNER=module.timeline`, clipboard mapping, shared planning and rollback provider rows | PASS 3/3, `20260918T173348Z-p1049459` |
| `test-slice OWNER=module.revisions`, History keyset and changed-field completeness rows | PASS 2/2, `20260918T173348Z-p1049482` |
| `test-slice OWNER=module.workbook`, keep-saved conflict row | PASS 3/3, `20260918T173348Z-p1049488` |
| `service-backed-test-slice OWNER=module.revisions`, four History browsing and eight mutation-recovery browser rows | PASS 13/13, `20260918T174236Z-p1192175` |
| `service-backed-test-slice OWNER=module.timeline`, four new review rows | PASS 11/11, `20260918T174525Z-p1302316`; final compact presentation PASS 11/11, `20260918T174926Z-p1414121` |
| `service-backed-test-slice OWNER=module.workbook`, retained batch accessibility row | PASS 11/11, `20260918T174526Z-p1303553`; final compact presentation PASS 11/11, `20260918T174927Z-p1414501` |

Mixed-run evidence is accounted by scenario, never called a passing full run:
`20260918T173730Z-p1075784` passed six clear/range and fourteen spreadsheet,
clipboard/fill, creation and correction scenarios; one clear assertion expected
obsolete conflict copy. Its corrected scenario passed in
`20260918T174235Z-p1191937`, which also passed all thirteen range lifecycle,
authority, native editing, gestures, selection and accessibility scenarios.
`20260918T174234Z-p1191697` passed five Find and both Hosts/Identities clipboard
scenarios. Its accessibility failure exposed a real loss of distinguishable
batch input previews. Bounded previews were restored in the typed projection;
the final accessibility rerun above passes without weakening that assertion.
`20260918T173729Z-p1075549` passed Timeline clear accessibility, query-continuation
History, and both shared History/Recovery visual scenarios (existing goldens
unchanged); only an obsolete accepted-result wording assertion failed and was
updated. No golden was edited or refreshed.

Feature fixture corrections: `20260918T173326Z-p1019505` used an unsupported filter,
the wrong rollback response property and a single-entry tag expectation.
`20260918T173815Z-p1140018` incorrectly attempted an unavailable older change-set
action and treated a collection value as an array. `20260918T174235Z-p1191937`
expected the unsupported button disabled rather than correctly absent. Corrected
expectations follow current server contracts and the final four-row run passes.
The new partial-conflict fixture failed in `20260918T174738Z-p1402832` (missing
fixture conflicts) and `20260918T174925Z-p1413894` (incorrectly expecting the
existing conflict gate to admit reversal); both fixture assumptions are corrected.
Types passed at `20260918T174200Z-p1189966`; focused outcomes/review/chooser/retention
passed 5/5 at `20260918T174200Z-p1189891`. Final source checks follow in BHR-05.

Generation ran through `make generate` (PASS at `20260918T173241Z-p981970` and
`20260918T174132Z-p1186769`). The initial new-browser selection required generated
groups and correctly refused before generation; an unused test import caused
`frontend-typecheck` failure at `20260918T173208Z-p980533` and was removed.

Performance: `20260918T174602Z-p1365668` passed typing acknowledgement (p95 35.3 ms),
ArrowDown selection (35 ms), and pointer focus (39.3 ms), each below 100 ms.
Blank-row creation measured 190.3 ms against 150 ms while other browser work ran;
that applicable row remains pending an isolated rerun. No threshold or fixture
has been weakened. BHR-04 stays IN_PROGRESS until its remaining evidence passes.

### BHR-04 exit

The isolated blank-row creation rerun PASS 14/14 at
`20260918T175128Z-p1488341`, p95 110 ms against 150 ms, closes the applicable
measurement failure. The final focused outcome/chooser/review suite, including
the preserved conflict gate, PASS 4/4 at `20260918T175129Z-p1488558`.
The four production review scenarios passed again after adding explicit local
Show requested entries navigation (no additional reads or late focus movement),
11/11 at `20260918T175514Z-p1526222`. Captures from
`20260918T174926Z-p1414121` were inspected at desktop and narrow/text-spacing
sizes: ordinary scrolling, visible focus, readable wrapping, no horizontal
clipping. Newer entries remain in server order; the explicit local control moves
to the marked entry without another History request.

| Applicable acceptance | Disposition / evidence |
| --- | --- |
| Four operations and every outcome; requested/returned/original/current facts | PASS: typed matrix, adapter/retained owner and production batch regressions. |
| Exact later-page match, multiple logical entries, off-window subject, current metadata | PASS: new review rows, shared lookup/browsing and live History service rows. |
| Three-read budget, retry/restart/cancel/stale fencing, bounded payloads, pruning | PASS: lookup/browsing/review unit rows and active-subject/pruning tests. |
| Read-only entry, legal permissions, concurrency, exact replay, attributed reversal | PASS: integrated review, current owner/transport and twelve History browser regressions; service atomic consequences. |
| Authority concealment/retirement and confirmation invalidation | PASS: integrated suspension/revocation/replacement/closure/demotion tests plus unchanged retained owner security rows. |
| Passive completion, drafts, range/Find/correction/creation, other consumers | PASS: new draft/focus browser assertions and fresh range, spreadsheet, Find, query-continuation, Hosts/Identities, and History regressions. |
| Keyboard, live/textual feedback, twenty-choice bound, narrow/text spacing | PASS: bounded chooser unit rows, production review controls, batch/clear accessibility and inspected captures. |
| Existing History/Recovery visual obligations | PASS: two selected existing visual scenarios; no golden change. |
| Applicable Timeline performance | PASS: four predicates, isolated creation rerun above; no limit change. |
| Global Undo, additional batch surfaces, new routes/storage/mutation semantics | N/A: explicitly excluded; no activation or contract migration introduced. |

No applicable blocker remains. G1–G8 remediations and compatibility dispositions
are implemented and tested. Next action: BHR-05 final maintenance, broad affected
frontend checks, final routing/drift/documentation validation and scope review.

### BHR-05 final maintenance and source review

Final scope remains the four Timeline batch families, shared read/presentation
primitives and their existing consumers. No backend route, request/receipt wire
schema, database migration, dependency or lockfile changed. The digest is untouched.
The authored source inventory and local component/runtime/History/Inspector guides
cover the new files. New frontend rows and four live-browser rows are routed by
`tools/test_families/web.workbook.json` and `module.timeline.json`; Make regenerated
only browser grouping and topology input-index derivatives. The Markdown lint
configuration now includes this controlling handoff, without adding documentation
to product verification inputs.

Final source review removed a redundant scanner result-page payload for bounded
navigation. Navigation publishes accepted pages through the existing browsing
owner, while action lookup still retains its reviewed result page for admission.
This also prevents a rejected continuation from retaining an additional response
payload alongside the three accepted browsing pages. The focused
`history_action_lookup`, `history_browsing_state` and `batch_history_review` slice
PASS 4/4 at `20260918T180855Z-p1817474`; final types PASS 2/2 at
`20260918T180855Z-p1817576`. No mutation or action-admission behavior changed.

| Public Make command / selection | Actual result / run root |
| --- | --- |
| `make agent-finalize` | Initial failure `20260918T175725Z-p1562283`: generated routing stale after the last test title. `make json-shape-check` reproduced it at `20260918T175813Z-p1562920`. `make generate` repaired it at `20260918T175840Z-p1563522`; finalize PASS at `20260918T175944Z-p1566735`. |
| `make frontend-unit` | First run 650/651 at `20260918T180026Z-p1571144`: one new browser heading selector violated the existing selector policy. Corrected to the scoped semantic heading used by Recovery; policy slice PASS 2/2 at `20260918T180323Z-p1687992`. Full rerun PASS 651/651 at `20260918T180524Z-p1745547`; the subsequent retention audit is additionally covered by the focused current-source run above. |
| `make frontend-typecheck` | PASS 2/2 at `20260918T180026Z-p1571146`; final retention source rerun above. |
| `make frontend-import-boundary-check` | PASS 2/2 at `20260918T180026Z-p1571153`. |
| `make lint-biome` | PASS 2/2 at `20260918T180026Z-p1571221`; subsequent authored source formatting through Make. |
| `make generated-artifact-policy-check` | PASS 3/3 at `20260918T180026Z-p1570830`. |
| `make generate-drift` | PASS 4/4 at `20260918T180026Z-p1570769`; final retention test title regenerated through Make at `20260918T180855Z-p1817387`. |
| `make lint-markdown` | PASS, including the new handoff glob, at `20260918T180026Z-p1571393`; completed-content rerun remains below. |
| `make service-backed-test-slice OWNER=module.workbook`, batch/clear accessibility and both affected History/Recovery visual rows | PASS 13/13 at `20260918T180056Z-p1586127`; no golden change. |
| `make service-backed-test-slice OWNER=module.timeline`, four new review rows | PASS 11/11 at `20260918T180058Z-p1586521`; corrected selector rerun PASS 11/11 at `20260918T180509Z-p1718055`. |

Final-source follow-up: `make agent-finalize` PASS at
`20260918T181108Z-p1825214`, with RESULTS_DIR unset and retained-run maintenance
skipped. The four production review rows PASS 11/11 at
`20260918T181333Z-p1829519` after the bounded scanner payload fix;
`make lint-biome` PASS 2/2 at `20260918T181334Z-p1829797`.

Retired paths are the duplicate batch outcome/operation wording, the separate
action-only page-scanning implementation, and redundant Recovery detail headings
and requested-scope text. The retained batch owner, conflict store, canonical
Inspector subject and History mutation owner remain the sole owners of their
existing state. No second receipt archive, executor or persistence path was added.

The selector-policy failure required the `web.architecture` task guide and its
narrow owning row; its implementation and policy were not weakened or changed.
No additional product owner inputs changed. Source/HEAD review still finds `main`
at `cd03bf6a2d53d884c667de344c6af16e3a7567a6`, zero behind/one ahead of origin,
with only this seam's uncommitted edits. `git diff --check` passes.

Retained-run maintenance is skipped: RESULTS_DIR is unset, as recorded by the
finalizer. No historical warm run was promoted to final-source evidence. Broad
backend lint/migration/release/deployment checks are N/A for unchanged backend,
schema, dependency and deployment inputs; relevant backend behavior is covered by
the fresh service slices. `web.workbook` has no service-backed route in its task
guide; live service evidence is routed through Timeline, Workbook and Revisions.

Compatibility limits are intentional: review is ephemeral, supports one record
at a time and does not inventory all change-set effects; old receipts remain
subject to existing pruning. Continue/Retry/Restart/Cancel expose bounded lookup.
Current History metadata may offer fewer actions than older observations, and
unresolved conflicts keep the existing action-coordination gate. Rollback of this
feature must revert owner wording, implementation, source/test inventories and
Make-generated routing together, while preserving all accepted writes, attributed
reversal revisions, receipts and History. No data rollback is part of code rollback.

### BHR-05 terminal evidence

The final isolated four-row Timeline measurement selection passed 20/20 execution
units at `.cartulary/test-results/20260918T181923Z-p1862764`. Actual p95 values:
typing acknowledgement 36.6 ms, ArrowDown selection 34 ms and focus/edit 36.1 ms
(each below 100 ms); blank-row creation 111.7 ms (below 150 ms). This covers the
frozen final source with unchanged fixtures and thresholds.

The finalizer's current-source generated transaction and catalog/schema checks
passed with no generated changes. Final scope inspection confirms the original
HEAD and ahead-of-origin commit, unchanged digest/lockfiles/generated production
roots, and only the listed owner, implementation, test, routing and guide changes.
The final diff whitespace check passes. Completed-content `make lint-markdown`
PASS at `.cartulary/test-results/20260918T182441Z-p1904285`; the status/evidence
closeout is checked again before returning the work. All applicable acceptance
rows pass; no product or acceptance gap remains. BHR-05 is DONE. Next action:
review the uncommitted diff; implementation stops at this seam. No commit, push,
deployment or analyst-data operation was performed or is required.
