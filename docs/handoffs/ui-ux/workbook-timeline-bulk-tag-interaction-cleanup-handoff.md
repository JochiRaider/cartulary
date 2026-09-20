# Timeline bulk-selection and tag-control cleanup

## Control and scope

Execution baseline: clean `main`, HEAD
`f6da8ce3ae5600b613c4751472cf2323e0a8501e`, revalidated before edits.
Only root AGENTS.md applies. Preserve unrelated work. Scope is Timeline
record-checkbox selection, tag authoring/admission presentation and immediate
query, layout and recovery consumers. No new commands, routes, backend tag
semantics, dependencies, persistence, cross-page selection, analyst-data changes,
digest edits, commits, pushes or deployment.

The user chose explicit assignment admission behind preceding saves and visible
retained tag authoring when selection becomes empty. Readiness never submits an
unsubmitted draft. The existing batch owner retains admitted operations.

| Workstream | Status | Binary exit |
| --- | --- | --- |
| BTI-01 Characterization and owners | COMPLETE | Production baseline, full gap rubric and unambiguous decisions recorded. |
| BTI-02 Selection and controls | COMPLETE | Focused ownership, admission, lifetime and layout tests pass; retired callers removed. |
| BTI-03 Production proof | COMPLETE | Production interactions, recovery, unchanged AC-043 and reviewed visuals pass. |
| BTI-04 Terminal handoff | COMPLETE | Finalizer, terminal checks and evidence-backed acceptance complete. |

Record each completed exit before starting its dependent. Historical handoffs
and planning evidence are context, not fresh completion proof.

## Authority and ownership

Core 03 REQ-03-297 owns committed identity and accepted query-window selection;
§13.3 REQ-03-221/222 owns multi-row tag assignment and batch semantics;
REQ-03-298/299/100 and §§3–4/13–14 own editing, continuity and scoped authority.
Core 01 §3.3.5 owns mutation identity, versions and replay; Core 04 §§1–2 owns
authorization. Design §§7–8/10/12/14 owns layout, local feedback and accessibility.
Domain supplies vocabulary; the NLSpec research essay is advisory.

The digest README, START_HERE, LOCAL_AGENT_PROMPT, REPO_MAP, OWNER_MAP, rules,
acceptance and QUERY_RECIPES were consulted during planning. Completed clipboard/
bulk recovery, batch History, working-set controls, range entry, collection input,
autosave feedback and auto-resolution handoffs were reviewed as regression
context. Source ownership is `web.workbook`; independent verification begins with
`web.workbook` and `module.timeline`. Additional owners follow affected boundaries.
No executable consumer reads Markdown.

| Concern | Decision and retained owner |
| --- | --- |
| Selection | One Timeline controller; authorized committed accepted query members, independently of pending saves. Adapter consumes this neutral membership predicate. |
| Submission | Revalidate all intended targets. Capture once on an explicit native form event; batch owner coordinates admitted prerequisites and permitted versions. |
| Known failed/unsubmitted edit | Keep selection and tag draft; explain locally and prevent unsafe dispatch. |
| Query/authority | Prune true departures and authority loss; append does not expand selection; pins/drafts/groups excluded. |
| Admitted batch | Captured targets survive later presentation changes; existing partial success, conflicts, exact replay and read-only reconciliation retained. |
| Authoring | One mounted leaf owns raw tag text; keep it visible at zero selection; explicit dismissal only; no persistence extension. |
| Completion | Observe owner facts without clearing newer authoring, changing focus/ranges or retargeting Inspector. |
| Layout | Compact contextual above-grid control in the existing feedback region; preserve working-set rail, auto-resolution disclosures and fixed row geometry. |

## Selection rubric and advisory decisions

| Gap | Classification / affected action | Remediation and areas | Boundary, rationale and future use | Capability / retirement / compatibility | Risk and binary validation |
| --- | --- | --- | --- | --- | --- |
| G1 | Source-confirmed pending selection pruning; edit then tag | Core clarification, Timeline membership/planner and tests | Identity membership distinct from execution readiness; later adopted consumers use the same neutral Adapter port | Retain multi-row tagging; retire pending-based membership; no wire/data migration | Intended targets silently disappear; selected identities and checkboxes survive held/rejected saves |
| G2 | Source-confirmed root draft observation; tag typing | Leaf authoring/controller boundary and tests | Local authoring updates only its consumer; future presentation can consume semantic selection commands without mirrored state | Retain raw text; retire root tag/message/submitting projections | Unrelated composition work; comparable diagnostics show no tag-only grid work, separately from latency |
| G3 | Source-confirmed shared query-rail placement; geometry hypothesis | Contextual feedback-region form, styles and renderer evidence | Separate working-set navigation from selected-record action; reuse existing layout ownership | Retain short keyboard/pointer workflow; remove supplemental controls prop/styles; no permanent row or overlay | Crowding and lost access; full-workbook supported geometry and long feedback pass |
| G4 | Source-confirmed ignored admission result and microtask pending | Semantic admission result and batch observation, tests | Existing retained owner owns operation truth and delivery identity; future callers reuse it | Retain execution/recovery; remove completion-like microtask state; internal callers migrate together | Hidden refusal or misleading completion; owner failure is local and pending/accepted states match actual operation |
| G5 | Source-confirmed conditional control disappearance | Always-mounted leaf, explicit lifetime and result revision fences | Presentation updates do not own authoring/operation lifetime | Retain mounted-memory scope; no reload/cross-tab guarantee; explicit dismissal only | Lost focus or newer text; native input/caret, ranges and Inspector survive relevant transitions |

Material offline advice: R001/R002/R006/R008/R010/R013/R014 ADOPT for semantic
controls, local recovery and continuity; R012/R018/R035 ADAPT through current
React owners, dense desktop geometry and bounded overflow; R026–R034 REJECT for
invented design authority, frameworks and incidental test selectors. Planning
queries returned 5 focus/recovery, 5 overflow and 8 React-state results without
fallback. No upstream design system was generated.

## Evidence ledger

Planning baseline controller/planner slice PASS 3/3 at
`.cartulary/test-results/20260920T151556Z-p67993`; existing production
spreadsheet-bulk PASS 11/11 at
`.cartulary/test-results/20260920T153006Z-p72062`. Execution-entry task guides
for web.workbook and module.timeline pass. Existing unit expectations that reject
pending checkbox eligibility are implementation evidence to replace.

## Compatibility, rollback and completion

Public request/receipt shapes and data remain unchanged. Roll back this slice's
source, bounded owner clarifications, verification inputs, generated routing and
intentional goldens together. Never roll back accepted records or History.
Acceptance, changed files, failures and their disposition, limitations, skipped
checks and next action are recorded below with retained execution evidence.

## BTI-01 exit

Production baseline PASS 11/11 at
`.cartulary/test-results/20260920T155417Z-p54703`; the report retains
`bulk-tag-observations` and full-workbook captures at 1440/1024/768 pixels.
Held autosave removes the selected record checkbox and membership; acknowledgement
restores the checkbox but leaves it deselected; rejection keeps it absent. At
768 pixels, the tag input/count overlap Group/Filters/Columns. The inspected
full-workbook capture confirms that G3 is a defect, not only a width hypothesis.
Grid top/height remain 122.39/749.61 pixels at these baseline widths.

Rejected hypothesis: a passive accepted row update alone did not replace the tag
input or lose raw text, focus or backward caret selection (3–8). Ten typed
characters generated 12 observed commits, no column/row prop replacements and no
HTTP requests. This is diagnostic work, not latency evidence. A repeated baseline
adds performed-grid-render counts; all after comparisons use the same probe.
Query/authorization/ordinary-entry boundary behavior is retained and receives the
explicit production regression matrix in BTI-03; no additional defect is inferred.

Probe corrections: typecheck failed at `20260920T154953Z-p12522` for an unused
import and a selector-helper argument, both fixed. The first browser attempt at
`20260920T155021Z-p17294` timed out because the observational probe waited on the
correctly observed absent checkbox; checking existence first fixed the probe.
These are fixture failures, not product failures. Make generation/format passed
at `20260920T154934Z-p9227` and `20260920T155412Z-p50353`.

Core 03 and design clarify only the approved correction. No owner contradiction
remains. The full rubric and state matrix above define the source/verification
boundaries and binary validation. BTI-01 is complete; next is BTI-02.

## BTI-02 exit

Focused controller/planner PASS 3/3 at `20260920T161313Z-p23897`;
frontend typecheck PASS 2/2 at `20260920T161338Z-p29059`; import boundaries
PASS 2/2 at `20260920T161400Z-p59660`. Production held-save capture/version
advancement and draft/caret retention passed in `20260920T160427Z-p40324`;
failed-save blocking and compact supported-width/zoom geometry PASS 11/11 at
`20260920T160712Z-p73731`. Existing batch retention/History PASS 3/3 at
`20260920T160713Z-p73994`. These complete the focused implementation exit;
the broader regression, recovery and measurement matrix follows in BTI-03.

Selection and checkbox policy now share Timeline's committed accepted-member
predicate; Adapter pruning mechanics remain unchanged. Whole-target planning and
readiness read current source owners synchronously. Native submit identity goes
to the existing batch owner. The command result exposes admission ID or refusal.
The leaf owns one raw draft and observes readiness and operations directly.
No root draft/message/submitting projection or microtask completion guard remains.
The supplemental query-bar prop/styles and their sole caller are retired together.

Intermediate fixture/build corrections: the first migrated unit test expected a
protected admission error while the owner was suspended; it now exercises an
observable closed-incident refusal. Import checking rejected a bulk-directory
test importing a component; the integration test moved to Timeline's root.
The zoom assertion now compares actionable input geometry with the query bar,
not the form's border box. Typecheck caught an incorrect collection selector
helper signature; corrected to the existing positional contract. A mistyped
planner row ID was rejected before execution; the catalog ID above passed.
No product assertion or measurement budget was weakened.

## BTI-03 evidence and review

The comparable baseline probe at `20260920T155543Z-p91351` observed 11 React
commits and 10 performed grid renders for ten tag characters; the after probe
at `20260920T161343Z-p29467` observed 10 commits and **zero** performed grid
renders. Both had zero row/column prop replacements and zero HTTP requests.
This is reduced unrelated render work, not a measured speedup. Native input,
raw text and backward caret selection survived passive updates in both runs.
After the correction, both selected identities and their checkboxes survive
held, acknowledged and rejected edits. Renderer observations and full-workbook
captures are retained as report attachments in those run roots.

The compact row moves grid top from 122.39 to 161.69 pixels at 1440/1024/768
width with 900-pixel height: a conditional 39.30-pixel allocation. Grid row
geometry is unchanged. No allocation is retained after selection/draft/feedback
and contained focus all end. The common feedback region reuses the existing
25cqh bound. Coexistence testing caught nested disclosure clipping; flex-column
allocation now lets disclosures shrink inside their existing scroll owner while
tag controls retain their height. This is the only additional layout correction.

Production evidence:

- `20260920T161343Z-p29467`: comparable characterization and duplicate native
  delivery versus deliberate repetition with ordinary collection/scalar entry.
- `20260920T161658Z-p71284`: all three densities, widths, Inspector minimum/
  maximum, text spacing and 200% zoom; batch tag History and whole-change reversal;
  range/Inspector/checkbox distinction, range entry/Find, collection authoring,
  autosave feedback, auto-resolution arrival and scoped authority behavior;
  conflicts-only tag receipts and independent local correction.
- `20260920T161955Z-p53013`: captured prerequisites succeed with permitted version
  advancement; a subsequent prerequisite failure retains both selected records and
  prevents transport of the admitted tag action.
- `20260920T162219Z-p86340`: membership scenario PASS 11/11, including stale failed
  refresh, 100/200/300 loaded-row append, eviction, return without reselection,
  accepted tag-filter replacement and closed-incident authority loss.
- `20260920T162003Z-p65465`: Find full retained window/eviction and creation-pin
  exclusion, plus default Timeline shell visual comparisons.
- `20260920T161659Z-p71512`: complete Grid Adapter focused slice PASS 56/56;
  no Adapter source change. Its semantic core-record filtering and neutral
  layout-effect pruning both consume Timeline's source-owned predicate.

Intermediate browser failures were retained, not counted as whole-run passes.
Membership initially attempted unsupported Summary filtering (`161343`); it now
uses the declared Tags filter. The next run (`161955`) exposed the fixture's
incorrect singular status expectation; matching the existing status text fixed
it. The replay test (`161658`, `162140`) incorrectly assumed one History item per
accepted action; it now compares captured post-acceptance History with post-replay
History, alongside identical request bytes and change-set identity. Coexistence
geometry failed at `162140` and required the flex allocation correction above.

### Visual maintenance

Ordinary visual evidence at `20260920T162003Z-p65465` showed exactly the intended
checkbox additions in pending/recovery states. Reviewed full-workbook actual/diff
captures: `timeline-mutation-pending-replay-status.png`,
`timeline-mutation-transaction-recovery-panel.png`, and its `-narrow` and
`-compact` variants. The only marked pixels were the now-retained row checkbox;
query rail, recovery, grid rows, status and drawer geometry were unchanged.
The retained reconciliation v3 had zero missing or ambiguous mappings and valid
renderer/manifest identity. Its target failure came from these expected
comparisons; unexecuted captures in this narrow run are not deletion candidates.
Owner row: `module.workbook.visual.capture_save_state_pending_replay_transaction_re_70f3e80a67`.
No viewport, masking, scroll normalization, crop, font or renderer pin changed.

The first Make-owned update (`20260920T162157Z-p55259`) was cancelled by the
harness with `cleanup_error` / interrupted group before promotion. Goldens and
manifest stayed unchanged. It supplies no successful update evidence; the
successful retry and ordinary validations are recorded below.

The corrected coexistence/replay/layout slice PASS 13/13 at
`20260920T162445Z-p32266`. It includes all three densities with tag authoring beside
four auto-resolution disclosures, native recovery, Inspector clamps, supported
200% effective viewport, and the existing below-minimum fallback. Full-workbook
captures were reviewed at compact Inspector-open, default maximum Inspector,
comfortable zoom, and comfortable supported zoom. No hidden query rail is claimed
available below the existing supported minimum; that pre-existing fallback is
unchanged. Exact replay compares both submitted bytes/change-set IDs and the
accepted History before/after replay, rather than assuming one History item per
logical operation. Failed refresh performs reads only.

Make-owned golden update PASS 12/12 at `20260920T162705Z-p70548`.
All 252 captures reconcile as active, with zero missing, orphan, ambiguous or
unresolved registered fixtures. Only the four listed checkbox goldens and their
manifest hashes changed; every promoted full-workbook image was inspected.
These captures are active nonregistry cases from `scenario_79fa228c8b2e`, not new
registered fixtures. Stable capture IDs are `visual.capture.a5356d72e06448c07560`
(pending), `visual.capture.124cf27d3d0177dea828` (recovery),
`visual.capture.c49072efd841f8cb0cea` (narrow) and
`visual.capture.af13dd93abd19940247c` (compact). The scenario's registered edit-cell
fixture remains unchanged. No tolerance, normalization or selector crop changed.

## Changed paths and retirement

- Timeline bulk selection/planning, readiness reader and command port/adapter:
  membership/readiness separation, full-set admission and native delivery identity.
- Timeline foundation/interaction/composition/presentation: pass semantic bindings;
  no root tag authoring observation or unused bulk command projection.
- New `TimelineBulkTagControl.tsx`: sole raw draft and revision-scoped local feedback
  owner; live batch observation and accessible native form.
- `TimelineWorkbookView.tsx`: contextual form and existing disclosures in the
  bounded feedback slot; shared `WorkbookSurfaceLayout.tsx` owns its unchanged
  geometry through `workbookSurfaceFeedbackStyle`; retained grid/scroll owners.
- `TimelineWorkbookViewBarRegion.tsx` and shared `WorkbookViewBar.tsx`: remove the
  replaced tag fieldset, supplemental prop/allocation styles and sole caller.
- Planner/adapter tests and new root `timelineBulkTagInteraction.test.tsx` replace
  incidental pending-pruning expectations; obsolete bulk-directory test removed.
- Production `timeline-bulk-tag.spec.ts`, auto-resolution coexistence assertions,
  and Find creation-pin eligibility assertions; existing semantic selectors used.
- Authored source inventory and Timeline catalog updated; browser batching and
  topology index generated only by `make generate`.
- Bounded Core 03/design clarification, bulk/component source guides, this handoff
  and its Markdown-lint inclusion; four reviewed golden PNGs and generated hashes.

Retained paths include the single Workbook batch owner, transport/version/replay
contracts, Recovery navigation, exact History, prerequisite save queue, draft
registry, neutral Adapter selection mechanics, query ownership, fixed row sizes,
virtualization, range/Inspector identity and the existing responsive fallback.
There is no backend, endpoint, dependency, persisted preference, reload or
cross-tab state change and no analyst-data migration. Test fixtures are isolated
harness data. No compatibility alias or duplicate owner remains.

Baseline stack revalidation used authored manifests: React 19.2.5, TypeScript
6.0.2, Vite 8.0.8, Playwright 1.59.1, Node 24.15.0, pnpm 10.33.0 and Go toolchain
1.27.1. Production `react-data-grid` imports remain confined to Grid Adapter.
Source/import inventory, verification registry/catalog, generator policy and
local Timeline/Workbook guides were inspected independently of the digest.
No localization mirror was updated. The working tree contains only this slice;
the original branch and commit remain unchanged.

Final focused refinements PASS: `20260920T163229Z-p40964` (11/11) explicitly
covers a single selected row receiving an update and nonzero grid scroll through
captured prerequisite acknowledgement, alongside the comparable diagnostics.
`20260920T163320Z-p77615` (11/11) adds production checkbox exclusion for a visible
out-of-query creation pin and verifies that it is unchecked when the accepted
query subsequently includes it. The final focused controller/planner slice is
PASS 3/3 at `20260920T162612Z-p69335` after removal of unused command projections.

Shared accessibility and responsive-frame production slice PASS 13/13 at
`20260920T163528Z-p10689` (`web.design`, the global keyboard/name/focus/non-color
matrix and viewport/panel overflow scenario). Focused controller/planner/adapter
slice PASS 4/4 at `20260920T163751Z-p84049`. The latter also exercises actual
retained draft revisions: an admitted scalar revision permits waiting, newer
unsubmitted text blocks, unselected drafts do not block, and selected Inspector
collection authoring blocks without executing or flushing it.

Ordinary post-promotion `make browser-e2e-visual` PASS 12/12 at
`20260920T163156Z-p6972` and `20260920T174530Z-p68663`. These are two fresh
successful full validations against the promoted manifest.

The second ordinary visual attempt (`20260920T163634Z-p45379`) failed one
unrelated Decision supersession capture: 161 pixels in the blank-row Owner UUID.
The inspected full-workbook diff contains only that dynamic identifier, while
all four promoted Timeline images passed. This slice changes neither Decision
ownership nor its visual fixture. The first ordinary full run and update had
already passed that unchanged capture. No Decision golden, mask, tolerance or
fixture was changed; the fresh ordinary run above established the second
successful full validation. The incidental capture instability is recorded rather
than hidden by snapshot promotion.

### AC-043 timing evidence

Each existing measurement ran alone through
`make service-backed-test-slice OWNER=module.timeline ROWS=<existing-row-id>`.
All four targets passed 14/14 execution units, their aggregate status is
`qualified`, and each reports zero scheduler overlaps and 100 samples.
Fixtures, predicates and budgets were unchanged. These results establish budget
compliance; no before/after latency improvement is claimed.

| Existing predicate | p95 ms | Limit ms | Run root under `.cartulary/test-results/` |
| --- | --- | --- | --- |
| `perf.typing_ack.v1` | 31.2 | 100 | `20260920T173141Z-p95715` |
| `perf.timeline_summary_selection_down.v1` | 32.4 | 100 | `20260920T173425Z-p31723` |
| `perf.timeline_summary_focus_edit.v1` | 27.3 | 100 | `20260920T173714Z-p64990` |
| `perf.timeline_blank_row_create.v1` | 76.3 | 150 | `20260920T174011Z-p98196` |

Each root retains `browser-e2e-measurement/frontend-measurement-aggregate.json`
and its referenced group summary/observation artifacts. The shared profile is
`ac043_large_grid_snapshot_v1`, snapshot key
`55233e62069d0026ce0f0925c1efaa9863cc7e43efd6d45f08722f047be53cb4`.
The workload retains 25 analysts, 24 background analysts and 4.8 updates/second.

Final delivery/entry/partial-conflict production slice PASS 11/11 at
`20260920T174634Z-p8275`. The same native event admits once, a deliberate repeat
has a new transaction ID, and scalar/collection entry preserves selection.
A held third assignment receives a concurrent tag edit on one target: the server
returns the other target as accepted plus one `collection_review` conflict.
Both checkboxes and raw tag text remain, and the local recovery reason disables
another assignment. No intended target was omitted before transport.
The first extended run (`20260920T174445Z-p36808`) failed its final message
assertion: it expected generic completion instead of the selected conflict's
blocking reason. Receipt assertions already passed. The corrected assertion
checks the recovery reason and disabled submission; product behavior was unchanged.

### BTI-03 exit

Native clipboard fidelity regression PASS 11/11 at `20260920T174808Z-p41724`
through the existing `module.timeline.browser.clipboard_fidelity` row.
Production interaction, recovery, full-workbook geometry, accessibility and
regression evidence is complete. All four unchanged AC-043 predicates qualify
within budget; both required ordinary full visual runs pass. Every intentional
golden change was reviewed, and no unrelated golden was promoted.
BTI-03 is complete as of 2026-09-20 17:50 UTC, before invoking the finalizer.
BTI-04 now runs `make agent-finalize` with `RESULTS_DIR` unset, then terminal checks.

## Acceptance assessment

This assesses the digest's complete A001–A027 checklist only within the changed
owners. Unchanged subsystem requirements do not create additional feature work.
Artifacts below are implementation evidence, not Core 05 publication claims.

| Row | Status | Applicable evidence / scope rationale |
| --- | --- | --- |
| A001 Authority | PASS | Core 03 REQ-03-297 and §13.3 REQ-03-221/222, Core 01 §3.3.5 and Core 04 authorization map to the selection/admission matrix; advisory research/digest did not define behavior. |
| A002 Scope | PASS | G1–G5 apply every rubric category; one selection owner and one draft owner; obsolete controls/projections removed with callers. |
| A003 Repository state | PASS | Clean branch/commit baseline, authored stack/source/import/routing inputs, local guides and sole vendor boundary revalidated; no digest edits. |
| A004 Tokens | PASS | Existing semantic colors, borders, radii and spacing; reused feedback bound and saved-view input allocation token; no theme/density registry. |
| A005 Theme | PASS | Existing dark_graphite only; reviewed production captures and token/theme visual fixture. |
| A006 Density | PASS | Three-density production geometry, four unchanged AC-043 predicates and two ordinary full visual validations pass. |
| A007 Creation | PASS | Production creation pin exclusion/return and unchanged blank-row creation timing gate pass. No creation policy or capability added. |
| A008 Responsive | PASS | Three-density 1440/1024/768 layouts, effective supported 200% zoom, Inspector min/max and text spacing; shared responsive-frame production scenario passes. Existing below-minimum fallback retained. |
| A009 Overflow | PASS | Bounded combined feedback, independently scrolling disclosures, unchanged grid/Inspector/document scroll ownership; protected toolbar and account navigation remain reachable in supported layouts. |
| A010 Inspector | PASS | Range presentation test distinguishes Inspector subject and record checkboxes; tag control has no Inspector mutation command. Existing dispatch/confirmation owners remain unchanged. |
| A011 Continuity | PASS | Same native input/raw text/caret through passive and single-row updates, zero selection and late outcomes; production scroll and selected identity assertions, query/Find and recovery tests. |
| A012 Transactions | PASS | Native delivery duplicate admits once; intentional repeated actions use distinct transaction IDs; exact replay bytes and change-set identity verified through production. Random identity owner unchanged. |
| A013 Recovery | PASS | Captured pending save succeeds/advances or fails/blocks without smaller dispatch; accepted tag/failed read retries reads only; original Recovery and exact History retained. |
| A014 Editing | PASS | Scalar and collection entry with bulk selection, selected unsubmitted draft revision blocking, range entry/Find regression, raw tag retention/dismissal and scoped authority tests. No persistence extension. |
| A015 Conflict | PASS | Failed selected edits retain checkboxes and local reason with existing cell/Recovery conflict feedback; conflicts-only collection review and attributed correction remain valid. |
| A016 Query/interaction states | PASS | Failed refresh retains authorized selection; accepted replacement/eviction prunes; append does not expand; closed authority disables authoring/assignment while preserving mounted raw text. |
| A017 Refresh/authorization | PASS | Query-membership production matrix; shared batch-owner suspension/late-receipt tests and auto-resolution session/account/access-loss production regressions. No new cross-account retained state. |
| A018 Evidence | N/A | No Evidence lifecycle, overlay, preview or access capability changed. Shared toolbar removal has no remaining Evidence consumer. |
| A019 Accessibility | PASS | Semantic native form, labelled input/buttons, disabled reasons, Enter/pointer parity, visible focus, zoom/spacing captures and existing global keyboard/name/focus/non-color matrix pass. No new motion or contrast palette. |
| A020 Components | PASS | Long raw Unicode text, local rejection, pending/accepted/uncertain states, zero-selection retained input, supported density/Inspector geometry and disclosure coexistence pass. |
| A021 Virtualization | PASS | Production 100/200/300 append/eviction, pin exclusion, range/Find, fixed row geometry and all four unchanged AC-043 predicates pass. |
| A022 Visual fixtures | PASS | Production full-workbook captures reviewed; four intentional checkbox goldens promoted by Make, all 252 reconcile; two fresh ordinary full validations pass. |
| A023 Selectors | PASS | Existing view schema, record/field IDs and accessible roles/names reused; no positional or incidental style selector defines target membership. |
| A024 Test authority | PASS | Executable source/tests never read Markdown; prose/source guides remain manual context and Markdown lint only. |
| A025 Generated artifacts | PASS | Authored source/catalog inputs and golden hashes generated through Make; finalizer/catalog, generated policy, JSON shape and generation drift checks pass. |
| A026 Compatibility | PASS | Internal port callers migrated together; unchanged wire/data/mutation/recovery contracts, no migration, explicit retirement and code-only rollback. |
| A027 Handoff | PASS | All four exits complete; finalizer and terminal checks pass; paths, command/artifact evidence, failures/dispositions, limits, skipped checks, rollback and next action recorded. |

### Explicit state lifetimes

Selection belongs to the mounted Timeline controller and the currently authorized
accepted loaded-query window. It survives transient saves, retains explicit
checkbox choices, prunes true scope departures, and does not grow on append.
The one raw tag draft belongs to the mounted leaf; selection/readiness/query/row
updates do not replace it. Submitted normalization never rewrites the raw input.
The form remains present for selected records, retained text, relevant feedback,
or focus inside it. At zero selection it explains the disabled action. Clear tag
draft retires raw authoring and local feedback only; leaving the empty form then
allows it to disappear without reserving height.

Local validation and submitted-operation feedback match both the authoring
revision and the selection identity captured by that submission. New authoring
or selection suppresses older local results. Accepted/rejected/uncertain batch
facts remain in their existing owner and Recovery even after local dismissal or
presentation detachment. No async result restores checkbox choices, resets the
input, requests focus, changes a range or retargets Inspector. No new reload,
cross-tab or cross-account lifetime is implied.

### Exact changed-file inventory

The deleted bulk-directory controller test is replaced by the root integration
test listed below; all other paths are additions or modifications in this slice.

```text
.markdownlint-cli2.jsonc
apps/web/e2e/timeline-auto-resolution-feedback.spec.ts
apps/web/e2e/timeline-bulk-tag.spec.ts
apps/web/e2e/timeline-find.spec.ts
apps/web/e2e/workbook.visual.spec.ts-snapshots/timeline-mutation-pending-replay-status-linux.png
apps/web/e2e/workbook.visual.spec.ts-snapshots/timeline-mutation-transaction-recovery-panel-compact-linux.png
apps/web/e2e/workbook.visual.spec.ts-snapshots/timeline-mutation-transaction-recovery-panel-linux.png
apps/web/e2e/workbook.visual.spec.ts-snapshots/timeline-mutation-transaction-recovery-panel-narrow-linux.png
apps/web/src/workbook/components/WorkbookViewBar.tsx
apps/web/src/workbook/layout/README.md
apps/web/src/workbook/layout/WorkbookSurfaceLayout.tsx
apps/web/src/workbook/timeline/adapters/createTimelineActionAdapters.test.ts
apps/web/src/workbook/timeline/adapters/createTimelineBulkTagCommandAdapter.ts
apps/web/src/workbook/timeline/bulk/README.md
apps/web/src/workbook/timeline/bulk/createTimelineBulkTagReadiness.ts
apps/web/src/workbook/timeline/bulk/useTimelineBulkTagController.test.tsx
apps/web/src/workbook/timeline/bulk/useTimelineBulkTagController.ts
apps/web/src/workbook/timeline/components/README.md
apps/web/src/workbook/timeline/components/TimelineBulkTagControl.tsx
apps/web/src/workbook/timeline/composition/useTimelineInteractionComposition.ts
apps/web/src/workbook/timeline/composition/useTimelineSurfaceFoundation.ts
apps/web/src/workbook/timeline/composition/useTimelineWorkbookComposition.ts
apps/web/src/workbook/timeline/models/timelineBulkTagPlan.test.ts
apps/web/src/workbook/timeline/models/timelineBulkTagPlan.ts
apps/web/src/workbook/timeline/ports/TimelineBulkTagCommandPort.ts
apps/web/src/workbook/timeline/presentation/README.md
apps/web/src/workbook/timeline/presentation/TimelineWorkbookView.tsx
apps/web/src/workbook/timeline/presentation/TimelineWorkbookViewBarRegion.tsx
apps/web/src/workbook/timeline/presentation/useTimelineWorkbookPresentation.tsx
apps/web/src/workbook/timeline/timelineBulkTagInteraction.test.tsx
docs/design.md
docs/handoffs/ui-ux/workbook-timeline-bulk-tag-interaction-cleanup-handoff.md
docs/spec/03_workbook_interaction_collaboration_and_workflows.md
tools/browser_e2e_batch_manifest.json
tools/execution_topology_render_index.json
tools/frontend_source_ownership.json
tools/frontend_visual_golden_manifest.json
tools/test_families/module.timeline.json
```

## Terminal verification and completion

`make agent-finalize` PASS 1/1 at `20260920T175025Z-p74467` and again after
the architecture corrections at `20260920T175637Z-p20258`, before the respective
broader terminal checks. Its `unit-artifacts/finalize-summary.json` reports
no generated changes and passing schema, catalog/tier and structure checks.
Retained-run maintenance was skipped because `RESULTS_DIR` was unset: no
successful full warm-check root qualifies for that input. Canonical retained-run
evidence and scheduler timing/order maintenance were therefore not selected.
The separate isolated AC-043 results above remain valid focused evidence.

All run roots below are under `.cartulary/test-results/`. Counts are harness
execution units, not individual assertions. Public Make targets use their normal
cache policy; an aggregate pass does not imply every unchanged test re-executed.

| Command | Result | Run root |
| --- | --- | --- |
| `make frontend-typecheck` | PASS 2/2 | `20260920T175658Z-p28749` |
| `make frontend-unit` | PASS 656/656 | `20260920T175723Z-p36354` |
| `make frontend-import-boundary-check` | PASS 2/2 | `20260920T175658Z-p28777` |
| `make lint-biome` | PASS 2/2 | `20260920T175708Z-p31641` |
| `make lint-markdown` | PASS; completed handoff | `20260920T180451Z-p38937` |
| `make generated-artifact-policy-check` | PASS 3/3 | `20260920T175051Z-p78384` |
| `make json-shape-check` | PASS 3/3 | `20260920T175055Z-p79512` |
| `make generate-drift` | PASS 4/4 | `20260920T175712Z-p32090` |

The first broad `make frontend-unit` failed 654/656 execution units at
`20260920T175112Z-p85728`. Both failures are preserved under the run's
`unit-logs/*/vitest-failure-details.json`. The new feedback sizing violated the
architecture policy requiring surface geometry in the shared layout owner.
The selector check exposed a pre-existing raw dynamic Inspector selector in the
auto-resolution scenario extended by this slice (confirmed against HEAD).
The identical feedback style object moved into the existing shared layout module;
the existing probe now uses
`dataTestIdSelector` with existing semantic builders. Neither policy test changed.
Focused `web.architecture` repair verification PASS 3/3 at
`20260920T175556Z-p59512`. This ownership-only style move changes no CSS value,
DOM structure, state, admission or transport behavior.

Final-source production characterization, tag geometry and the complete
auto-resolution family PASS 13/13 at `20260920T175616Z-p60719`. The repeated
diagnostic remains 10 commits, zero grid renders, zero row/column replacements
and zero HTTP requests for ten typed characters. Raw text, backward caret 3–8,
native input identity and selected committed identities remain stable. Fresh
full-workbook 768-pixel, supported 200% comfortable and combined disclosure/tag
captures were reviewed; geometry is unchanged from the earlier proof.
The full frontend unit target now passes all 656/656 execution units, including
both unchanged architecture policies.

The additional final-source full visual run at `20260920T175616Z-p60873`
passed all 46 Workbook scenarios but failed the separate Network Analysis
scenario (aggregate 10/12 execution units). Full actual/diff images show only
the enabled/disabled appearance of the existing Save current graph button in
`network-flow-analysis-compact-saved-graphs` and `network-flow-analysis-delete-dialog`.
Network Analysis does not consume the moved feedback style or shared Workbook
surface component; its source, fixture and goldens are unchanged. This is an
incidental out-of-scope capture-state mismatch, not a Timeline geometry failure.
The two earlier successful ordinary full visual runs remain recorded above;
the unchanged Network Analysis row passes its narrow isolated recheck, 11/11 at
`20260920T180313Z-p6505`, using
`make service-backed-test-slice OWNER=module.networkflow ROWS=module.networkflow.visual.capture_deterministic_claimed_network_analysis_a_47b1c2cce6`.
No additional golden promotion or unrelated product change was made.

### Limitations and skipped scope

- Raw authoring and selection retain their existing mounted-memory lifetime; no
  reload or cross-tab persistence is promised. Existing access/account lifetime
  owners still fence protected state.
- The pre-existing below-minimum effective viewport fallback is unchanged.
  Reachability claims apply to supported effective viewport sizes, including
  the supported 200% zoom case, not to an arbitrarily undersized viewport.
- Render counts demonstrate less unrelated work. The unchanged timing gates
  demonstrate budget compliance, not a comparative speedup or Core 05 claim.
- The unrelated intermittent Decision blank-row UUID capture mismatch remains
  recorded as incidental visual-fixture instability. It passed the final full
  ordinary run without changing its source, golden, mask or tolerance.
  The additional Network Analysis button-state mismatch passed its isolated
  recheck unchanged. Neither failed aggregate run is represented as a full pass.
- No full backend `make test`, full warm `make check`, `make ci`, release gates,
  deployment, security audit or unrelated browser-owner sweep was run. No
  backend, dependency, endpoint or persisted-data change calls for those here;
  production verification uses isolated harness services and fixtures.
- No digest or localization/provenance files were modified. No upstream design
  generation, analyst-data modification, commit, push or deployment occurred.

Rollback restores this slice's source, bounded owner wording, authored
verification/source inventories, generated routing and four intentional visual
goldens/manifest together. No accepted record or History data rollback is needed.
The next action is review of this completed working-tree diff and handoff;
there is no pending implementation or migration step.

### BTI-04 exit

Complete. Both finalizer runs passed with retained-run maintenance explicitly
skipped because `RESULTS_DIR` was unset. Final-source typecheck, all 656 frontend
unit execution units, import boundaries, Biome, Markdown, generated policy,
JSON shape and generation drift pass. Production proof and failure dispositions
are retained above; all applicable acceptance rows pass, with A018 justified N/A.
`git diff --check` passes. Final review confirms 38 changed paths, including the
replaced test deletion, with no digest, dependency, backend or analyst-data edits.
Branch remains `main` at `f6da8ce3ae5600b613c4751472cf2323e0a8501e`;
the working tree contains this uncommitted slice only.
