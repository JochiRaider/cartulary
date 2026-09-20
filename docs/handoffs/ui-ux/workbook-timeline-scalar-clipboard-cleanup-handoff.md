# Timeline scalar clipboard cleanup handoff

## Control, authority and scope

Execution baseline: `main`, commit `95c750b8e5a74b4d73e6c14d71afe711bf322479`,
clean checkout. The user authorized the existing Timeline scalar clipboard and
paste-settlement correction, its bounded Adapter consumers, tests, routing and
this handoff. No backend/API, persistence, dependencies, codec redesign, digest
edits, commits, pushes, deployment or analyst-data changes are authorized.

Planning read AGENTS.md; the digest README, START_HERE, LOCAL_AGENT_PROMPT,
REPO_MAP, OWNER_MAP, rules, acceptance and QUERY_RECIPES; domain/design and the
NLSpec research essay; the completed clipboard-fidelity, committed-grid editing,
range-entry, collection-input, autosave-feedback, auto-resolution-feedback and
bulk-tag handoffs. Advisory and historical material supplies context, not fresh
verification or additional tasks. The root AGENTS.md is the only applicable one.

Behavior owners: Core 03 §11.1 REQ-03-147 (native editor versus grid clipboard),
§4.1 REQ-03-087/088 (paste-completion autosave), REQ-03-298/300 and §13
(revision retention, acceptance and keyboard ownership); Core 01 §3.3.5
(mutation identity and replay); Core 04 §§1–2 and REQ-04-053 (authorization,
clipboard bounds and spreadsheet-export protection); design §§8,10,12,14.
Domain owns vocabulary/navigation. No owner contradiction has been found.

Source ownership is `web.workbook`, with neutral vendor/session mechanics in
`packages/grid-adapter`. Independent verification begins with `web.workbook`
and `module.timeline`, adding `package.grid_adapter` for that boundary. Current
task guides confirm narrow `test-slice`/`service-backed-test-slice` routing.
Authored manifests own routing; generators own derived outputs. Product tests
and runtime never consume Markdown. Current stack: React 19.2.5, RDG
7.0.0-beta.59, Playwright 1.59.1.

## Workstreams and binary exits

| Workstream | Status | Required exit |
| --- | --- | --- |
| SC-01 Production characterization and owners | PASS | Native reproduction and surface comparison recorded below; owner matrix and rubric complete. |
| SC-02 Clipboard and settlement correction | PASS | Focused ownership tests and native clipboard/autosave/departure proof pass. |
| SC-03 Continuity, recovery and hot-path proof | PASS | Production matrix, retained regressions and unchanged AC-043 gates pass. |
| SC-04 Terminal verification and handoff | PASS | Finalizer, terminal gates and evidence-backed acceptance complete. |

## Interaction and lifetime matrix

| Context or transition | Required behavior and owner |
| --- | --- |
| Active scalar control, including readable read-only | Native selected-substring copy; collapsed copy follows the browser; no grid selection or Inspector side effect. |
| Editable scalar cut/paste/history | Browser owns insertion, normalization, selection, caret and Undo/Redo. Actual completed DOM value feeds Timeline retention and paste autosave. |
| Grid navigation copy/paste | Existing representations, formula neutralization, geometry, parsing and stable target planning remain. |
| Evidence file/image | Existing capture-phase owner precedes scalar/grid text. |
| Native input | Mounted buffer, retained source revision and Adapter session have different lifetimes; synchronize once without replacing the native control. |
| Paste then Enter/Tab/blur | One source-owned operation per captured revision; departure awaits its authoritative acceptance. |
| Another paste or newer typing | Fresh revision; older completion cannot clear text or complete obsolete navigation. Equal text is not delivery identity. |
| Creation input/paste | First-input capture admits once; promotion preserves source identity and current focus/caret. |
| Grid and Inspector | Independent authoring and captured revision ownership. |
| Rejection/uncertainty | Exact draft and correction path retained; captured uncertain attempts replay unchanged. |
| Accepted write followed by failed read | Acceptance persists; recovery performs reads only. |
| Escape | Cancel unsubmitted local editing and its navigation; do not roll back an admitted or accepted record write. |
| Read loss/account replacement | Existing concealment, suspension and retirement owners remain authoritative. |

## Selection rubric

| Gap and classification | Remediation and change areas | Boundary, rationale, future use and capability | Retirement, compatibility, risk and validation |
| --- | --- | --- | --- |
| G1 Source-confirmed whole-buffer copy | Restore browser default in the scalar control; production copy evidence. | Native control owns text selection for all three existing presentations; future scalar consumers reuse that control. Copy remains useful without mutation. | Remove clipboard writing/focus-anchor side effects. No wire migration. Wrong copied text remains the concrete risk. Selected `beta` must copy exactly. |
| G2 Source-confirmed reconstructed paste; caret/history consequences require browser evidence | Observe completed native input; retain size admission and isolation. Implementation, callers and tests. | Browser owns text editing; Timeline owns completed save notification. This preserves native normalization and enables local history without a new subsystem. | Remove reconstruction and paste/capture callback chains. No codec or data migration. Actual DOM value/caret/history and autosave must agree. |
| G3 Separate paste/departure submissions; duplicate-write hypothesis | Share scalar settlement by semantic authoring revisions in the existing source mutation owner. | One admission outcome serves several event consumers; later presentations use the same source owner. Existing queue/replay/recovery remain valuable. | Retire content-based scalar duplicate/replanning paths. Keep FIFO/coalescing. Risk: duplicate writes or premature departure. Held acceptance/rejection must settle one gesture once. |
| G4 Equal-value native input does not advance all revision guards | Advance source/Adapter authoring on native input, with neutral optional Adapter metadata. | Editing identity is distinct from string equality; generic consumers retain default behavior. | Keep mounted/session/runtime layers for distinct lifetimes; remove duplicate registry writes. Internal TypeScript callers migrate together. Older receipt must not clear newer equal-text paste. |
| G5 Native fidelity evidence gap and structural work | Add production probes and regression scenarios, preserve four existing AC-043 predicates. | Real browser behavior establishes fidelity; corrected diagnostic counters establish work only. | Retire synthetic prevented-paste assertions as fidelity evidence. Risks include false confidence and unsupported speed claims. No latency claim from render counts. |

Offline queries executed during planning returned four UX and four React results,
without fallback. R001/R002/R013 ADOPT for keyboard, visible focus and semantic
controls; R012 ADAPT for state/effect advice according to actual ownership and
lifetime; R033/R034 REJECT for invented authority and incidental selectors.
No upstream design system was generated.

## Evidence ledger

- Planning baseline: `make service-backed-test-slice OWNER=module.timeline
  ROWS=module.timeline.browser.clipboard_fidelity` PASS, 11/11 graph units,
  `.cartulary/test-results/20260920T185119Z-p56044`. Existing editor coverage
  selects all and replaces text; it does not establish partial-copy or history.
- Execution recheck: branch/commit unchanged; checkout clean before this slice.
- Characterization setup: `make generate` PASS `20260920T192615Z-p98638`;
  `make format` PASS `20260920T192651Z-p2415`; `make frontend-typecheck` PASS
  `20260920T192953Z-p71803`.
- Initial browser run `20260920T192705Z-p6809` captured actual clipboard
  `alpha beta gamma` for partial, backward and collapsed copy; middle replacement
  `alpha B🙂 gamma` had caret 15 rather than 9; continued typing appended at 15;
  a second native Undo did not undo paste. Cut copied `alpha` and removed that
  selected text correctly. Single-line reconstructed paste stripped newlines in
  the DOM but started from unsanitized submission text. The probe's final Saved
  expectation failed at Conflict; this is retained production evidence, not an
  assertion to weaken in final regressions. Later characterization records this
  outcome to continue observing the other surfaces.
- Probe corrections: `20260920T192839Z-p39277` exposed a missing required
  view-schema argument to the Inspector selector. The next run used a field not
  offered by Inspector and was interrupted (`20260920T192953Z-p71744`). Inspector
  offers RAW Activity and Activity Synopsis only; the matrix now uses those actual
  controls. These are fixture errors. Event diagnostics use a next-task read:
  capture-listener microtasks run before later listeners and cannot establish
  final default prevention. Clipboard contents and observed DOM/caret above are
  unaffected by that diagnostic correction.

## Delivery record

A001–A027 are assessed below with PASS evidence or a specific N/A rationale.
The handoff records changed files, removed/retained paths, compatibility,
commands, artifacts, failures/disposition, limitations, skipped checks, rollback
and next action. `make agent-finalize` ran before broader terminal verification;
retained-run maintenance was skipped because `RESULTS_DIR` was unset. No visual
golden was added or refreshed.

## SC-01 exit

Production characterization PASS 11/11 graph units at
`.cartulary/test-results/20260920T193243Z-p5396`. Its JSON attachment records 60
observations across committed grid single-line/multiline and both actual Inspector
scalar controls, requests, real clipboard contents, element identity and native
event/work diagnostics. All four controls copied the whole buffer for partial,
backward and collapsed selections. Middle paste placed the caret at 15, and
native Undo did not undo insertion. Cut was native. No clipboard-only mutation
occurred. Controls stayed connected and identical; remount is not the cause.

Grid continued editing after the first accepted paste reached Conflict without
another HTTP patch; Inspector continued successfully. This establishes an
additional baseline/settlement defect in the managed-grid path, not a backend
rejection. G3/G4 remediation includes starting new authoring from current accepted
source facts after the prior revision retires, while retaining existing baselines
for unresolved authoring. Subsequent production held/rejected, creation, read-only,
independent-surface, IME and recovery cases are required SC-02/03 proof gates.

Native cut and typing produced real input events; reconstructed paste prevented
default and produced no insertion input event. The correction will observe native
`input` after default insertion, with `insertFromPaste` as the completion cause.
Production verification must prove this mechanism before SC-02 can pass. The
complete selection rubric and authority matrix above apply. No authority conflict
or additional product decision remains. SC-01 PASS recorded before SC-02 source
edits; next action is the bounded control/source-settlement correction.


## SC-02 exit

PASS before beginning dependent continuity work. The native-input correction and
revision-owned settlement are implemented. Focused Timeline slice PASS 5/5
(`20260920T194308Z-p49267`); workbook autosave/sentinel slice PASS 13/13
(`20260920T194614Z-p25643`); Adapter owner slice PASS 56/56
(`20260920T194549Z-p88050`). The subsequently expanded equal-value Adapter case
will run again in SC-03. Native characterization plus held/rejected rapid-departure
production tests PASS 11/11 graph units, 2 browser cases
(`20260920T194612Z-p23875`). All eight grid/Inspector Enter, Tab, blur and Escape
paths produce exactly one accepted patch. Rejection preserves exact whitespace,
Unicode and multiline text and repeated departure does not resubmit it.

Partial/backward copy now yields `beta`; collapsed copy leaves the clipboard
sentinel unchanged. Middle replacement has caret 9 before and after acceptance;
continued typing inserts at 9; native Undo restores typing then the original
selected `beta`; Redo restores both. Copy and local history add no mutation.
Single-line native insertion produces `雪\t"quoted" line`; textareas produce
`雪\t"quoted"\nline\n`. Submitted payloads match those actual DOM values.
The same mounted node stays focused through all observations. All four controls
finish Saved, with three patches each and no table-paste transport.

Removed whole-value copy, selection reconstruction, `onPasteCommit`,
`onCaptureInput`, scalar paste renderer typing/controller threading, content-based
scalar duplicate admission and the departure replanning exception. Kept the local
control buffer, runtime draft store and Adapter session for their distinct
lifetimes. Managed retention now goes through the Adapter once, with an optional
revision flag and a current accepted-row baseline. Scalar settlement shares one
captured operation across callers; subsequent native edits allocate revisions.
Existing replay/FIFO/coalescing remains the request owner. Read-only copy and
Evidence capture precedence remain native/existing respectively.

Intermediate typecheck failures were migration issues (a missing test fixture map,
two synthetic Event clipboard typings, and unused imports/request diagnostics).
They were corrected; final typecheck remains a terminal gate. No baseline runtime
failure was waived. Source complexity/render counts are not latency evidence.


### SC-03 discovered settlement gap

`20260920T195109Z-p99700` confirms that two identical native paste gestures are
admitted separately, but the later unchanged patch receives HTTP 400
`invalid_mutation_payload/no_effective_change` after the predecessor succeeds.
G6 (confirmed): retain distinct operation identity, then settle an unsent scalar
head only after the source driver verifies its accepted baseline and requested
values against the authoritative row. This is bounded to source settlement and
a neutral guarded queue retirement method; captured/uncertain attempts are
excluded. Benefit and extension: other verified no-change source outcomes can
use the guard without changing replay semantics. Risk: premature retirement;
validate FIFO/authorization/never-dispatched guards plus real repeated-paste
continuity. No backend/API change or content-based gesture suppression.

Inspector keyboard save also navigated before receipt; its existing navigation
now waits for acceptance and checks authoring revision, interaction sequence,
connected input and current focus. This is G3's owner-required departure remedy,
not a new shortcut. The same run passed lost-response/failed-read recovery; two
other failures were fixture mistakes (caret offset 7 instead of 8 and ARIA on
the grid shell instead of its grid). The expanded Adapter case had one stale
expected call count after adding a second explicit departure; corrected to two.


### SC-03 continuity and regression evidence

- `20260920T195712Z-p53760`: rapid-departure and captured-replay cases passed.
  Remaining failures were fixture assumptions: opening Inspector intentionally
  closes the previous clean grid editor; explicit Enter after last-row creation
  intentionally advances to the next capture row; readonly grid rendering uses
  Adapter semantic cells, not writable producer spans. These fixtures now use
  their actual supported targets. Paste without departure stays on the promoted
  record; Enter's explicit destination is not an unsolicited focus move.
- `20260920T200333Z-p38588`: creation promotion PASS. Find borrowing must use the
  existing Find button while a native editor owns text keys; no new Ctrl+F
  interception is introduced. The readonly fixture was still locating a writable
  producer span and was corrected to the row/field semantic selector.
- `20260920T200706Z-p87255`: newer identical paste, independent grid/Inspector,
  Find borrowing and detached acceptance PASS; bulk-tag rejection PASS after
  marking its form as an external editor action. Composition/scroll continuity
  passed before the remaining readonly locator timeout. No product assertion
  was removed to make the timeout pass.
- `make test-slice OWNER=web.workbook`: 282/283 units at
  `20260920T195850Z-p91084`. The sole failure was the websocket continuity fixture
  still sending synthetic `change`; migrated to `input`. Its focused rerun PASS
  2/2 at `20260920T200426Z-p76876`.
- Expanded Adapter revision test PASS 2/2 `20260920T195606Z-p50698`; keyboard
  ownership PASS 2/2 `20260920T195607Z-p51224`; guarded no-change queue and existing
  FIFO/replay suite PASS 2/2 `20260920T200135Z-p68717`. Queue verification adds
  `module.workbook`, whose current task guide was consulted.
- Evidence file paste and both creation orderings PASS 11/11 at
  `20260920T200134Z-p65482`, selected via `module.evidence.browser.file_draft_orderings`
  after consulting its task guide. Evidence capture still precedes scalar input.
- 32 existing clipboard, spreadsheet, range-entry/departure, collection-input,
  autosave, auto-resolution and bulk-tag rows ran at `20260920T200130Z-p63936`:
  19/21 graph units; all except bulk-tag rejection passed. That concrete focus
  regression is repaired and its unchanged test passes in the run above.
- Full `module.timeline` owner selection also includes service-backed, support,
  measurement and existing visual stages. No goldens were edited or refreshed.
  The four AC-043 measurement rows qualified at `20260920T195849Z-p90849`:
  `browser-e2e-measurement/frontend-measurement-aggregate.json` records one fixture
  build, four clones, zero scheduler overlaps and the following unchanged gates.

| Existing AC-043 predicate | Samples | p95 ms | Budget ms | Outcome |
| --- | --- | --- | --- | --- |
| Timeline focus/edit | 100 | 28.0 | 100 | PASS |
| Committed synopsis typing acknowledgement | 100 | 32.4 | 100 | PASS |
| Synopsis ArrowDown selection | 100 | 33.1 | 100 | PASS |
| Blank-row creation | 100 | 109.0 | 150 | PASS |

Predicates, fixtures, warmups, load envelope, budgets and timing boundaries were
not changed. No before/after latency comparison was collected and no speedup is
claimed.

### Render work, separate from latency

The production diagnostic excludes reused fibers. Comparable baseline
`20260920T193243Z-p5396` and final corrected `20260920T201312Z-p67662` workloads give:

| Workload | Baseline commits / renders | Corrected commits / renders |
| --- | --- | --- |
| Three further copy gestures, every control | 0 / 0 | 0 / 0 |
| Grid single-line paste through acceptance | 13 / 603 | 12 / 603 |
| Grid multiline paste through acceptance | 13 / 659 | 11 / 610 |
| Inspector RAW paste through acceptance | 12 / 988 | 13 / 1007 |
| Inspector synopsis paste through acceptance | 13 / 986 | 10 / 906 |
| One grid typing input after acceptance | 1 / 8 | 1 / 8 |
| One Inspector typing input after acceptance | 3 / 100 | 1 / 3 |

Both paste paths observed nine row-prop replacements. Grid column-prop
replacements stayed 12; Inspector changed 13 to 12. Typing and copy observed no
row/column replacements. These are instrumented work observations, not timing
measurements or evidence of perceived latency. Baseline/corrected post-paste text
history differs by design; comparisons stop at equivalent first-paste acceptance
and one typing input rather than treating the entire different history as equal.


## SC-03 exit

PASS before terminal work: required production continuity/recovery cases and all
four unchanged AC-043 gates have successful retained evidence. Final composition,
multiline scroll and readable readonly native copy/cut/paste coverage PASS 11/11
at `20260920T201032Z-p26542`. The readonly fixture now targets the actual semantic
row/field cell before opening its readable Inspector. Real clipboard copied
`line`; cut/paste did not change the readonly control or submit a mutation.

The new scalar cases together cover actual browser clipboard/input/history,
authoritative paste completion, eight rapid-departure combinations, exact rejected
text, newer identical revisions, independent surfaces, Find borrowing, detachment,
creation promotion, composition across older acceptance, uncertainty and read-only
refresh recovery. Source settlement callbacks survive presentation detachment and
are retired with the account runtime; their mounted consumers fence obsolete
navigation independently. The large owner selection remains additional regression
evidence and any failed rows are recorded with their corrective reruns.


## Terminal verification routing and compatibility

`make agent-finalize` PASS 1/1 at `20260920T201228Z-p63529` before terminal
verification. `RESULTS_DIR` was unset: retained-run maintenance skipped. None of
this slice's partial/failed owner selections is presented as a successful full
warm check. Final authored formatting PASS `20260920T201311Z-p67493`.

The expanded `make test-slice OWNER=module.timeline` finished 81/84 graph units
at `20260920T195849Z-p90849`. Its failed nodes were the known bulk-tag focus case,
readonly fixture selector case and their aggregate. Other routed unit, Go,
support, production browser, AC-043 and existing visual rows passed. Bulk-tag's
unchanged failing test passed after the external-action marker correction;
readonly passed after targeting the correct semantic cell. Final focused browser
selection repeats all six new scalar cases plus clipboard fidelity and bulk-tag
rejection against the final authored source.

No backend/API, schema, persistence, dependency, codec, formula-protection,
authorization or analyst-data migration. The optional Adapter authoring flag
defaults to existing behavior for every other consumer. The clipboard byte-limit
helper is exposed through the existing Adapter facade; implementation and test
facades agree. The source queue's guarded unchanged-head settlement cannot retire
in-flight, previously dispatched, blocked, unauthorized or non-head work. It
recognizes a no-change outcome only after source validation; it does not treat
identical clipboard text as operation identity. Existing captured retries keep
exact bytes and transaction IDs. No workbook Undo subsystem or History rollback
was added.

Source settlement now survives presentation detach. Account-runtime retirement
clears scalar settlement references and driver callbacks; mounted keyboard and
Adapter consumers independently refuse obsolete navigation. The independent
bulk-tag form borrows focus using the established external-editor-action marker,
so a rejected scalar receipt cannot hijack that control's authoring.

### Changed files

- `apps/web/e2e/timeline-scalar-clipboard.spec.ts`
- `apps/web/src/app/App.timeline-invalidation.support.test.tsx`
- `apps/web/src/testing/timelineWorkbookTestSupport.ts`
- `apps/web/src/workbook/WorkbookShell.autosave.test.tsx`
- `apps/web/src/workbook/WorkbookShell.sentinel.test.tsx`
- `apps/web/src/workbook/timeline/components/TimelineBulkTagControl.tsx`
- `apps/web/src/workbook/timeline/components/TimelineScalarEditor.test.tsx`
- `apps/web/src/workbook/timeline/components/TimelineScalarEditor.tsx`
- `apps/web/src/workbook/timeline/components/TimelineWorkbookRendererTypes.ts`
- `apps/web/src/workbook/timeline/components/TimelineWorkbookRenderers.tsx`
- `apps/web/src/workbook/timeline/components/useTimelineColumnAssembly.tsx`
- `apps/web/src/workbook/timeline/components/useTimelineScalarRenderers.tsx`
- `apps/web/src/workbook/timeline/composition/useTimelineInteractionComposition.ts`
- `apps/web/src/workbook/timeline/editing/useTimelineEditorDraftRegistry.test.tsx`
- `apps/web/src/workbook/timeline/hooks/useTimelineClipboardPasteController.ts`
- `apps/web/src/workbook/timeline/hooks/useTimelineKeyboardController.ts`
- `apps/web/src/workbook/timeline/hooks/useTimelineMutationCommands.ts`
- `apps/web/src/workbook/timeline/models/timelineMutationModels.test.ts`
- `apps/web/src/workbook/timeline/models/timelineMutationQueueAdmission.ts`
- `apps/web/src/workbook/timeline/models/timelinePendingSaves.ts`
- `apps/web/src/workbook/timeline/mutations/WorkbookTimelineMutationOwner.ts`
- `apps/web/src/workbook/timeline/mutations/createTimelineMutationDriver.ts`
- `apps/web/src/workbook/timeline/presentation/useTimelineWorkbookPresentation.tsx`
- `apps/web/src/workbook/timeline/useTimelineCommittedRecordIdle.test.tsx`
- `apps/web/src/workbook/timeline/useTimelineKeyboardController.test.tsx`
- `apps/web/src/workbook/utils/workbookPendingQueue.test.ts`
- `apps/web/src/workbook/utils/workbookPendingQueue.ts`
- `docs/handoffs/ui-ux/workbook-timeline-scalar-clipboard-cleanup-handoff.md`
- `packages/grid-adapter/src/core.ts`
- `packages/grid-adapter/src/index.test.tsx`
- `packages/grid-adapter/src/index.tsx`
- `packages/grid-adapter/src/rdgCompiler.tsx`
- `packages/grid-adapter/src/test-support.tsx`
- `tools/browser_e2e_batch_manifest.json`
- `tools/execution_topology_render_index.json`
- `tools/test_families/module.timeline.json`
- `tools/test_families/module.workbook.json`


### Limits, skipped work, rollback and next action

- Native clipboard, insertion, caret, history and composition evidence uses the
  production Chromium renderer. Composition is driven through Chromium's native
  CDP IME API; physical OS input methods and other browser engines were not run.
  No cross-platform fidelity or OS latency claim is made.
- Creation changes semantic identity from a recordless draft to a committed row.
  Evidence proves one record, current text/selection/focus transfer and continued
  typing on that record. It does not assert that a native Undo stack spans that
  intentional control promotion. Ordinary accepted edits retain their native
  control and Undo/Redo.
- Byte-limit rejection has focused handler coverage; no eight-mebibyte system
  clipboard stress measurement was performed. All decisive normal editing
  evidence uses actual clipboard shortcuts, not synthetic insertion.
- No intentional layout/theme/density change, new golden or golden refresh.
  Existing visual rows happened to run through owner selection and passed.
  Full repository release/CI/backend security suites are outside this frontend
  correction; owner-routed coverage and terminal gates are the review evidence.
- Rollback only this slice's authored files and handoff, then regenerate the two
  derived routing outputs through Make. No database or data rollback is needed.
  Preserve any subsequent unrelated edits; no reset/clean command is required.
- No commit, push or deployment was performed. Next action is human review of the diff and this handoff.


### Terminal results

- Final production browser selection PASS 15/15 graph units, eight browser cases,
  `20260920T201312Z-p67662`: all six new scalar scenarios, existing clipboard
  fidelity/malformed representations and unchanged bulk-tag rejection. Attachments
  in `browser-e2e-webserver-backed/browser-groups/*/playwright-report.json` contain
  actual clipboard, DOM, selection, identity, focus, request and mutation evidence.
- Owner-selected accessibility PASS 11/11, `20260920T201313Z-p68869`, row
  `module.workbook.accessibility.verify_grid_navigation_edit_entry_exit_paste_fee_3de2b0afac`.
- `make frontend-typecheck` PASS 2/2 `20260920T201337Z-p31252`;
  `make frontend-import-boundary-check` PASS 2/2 `20260920T201337Z-p31256`;
  `make lint-biome` PASS 2/2 `20260920T201337Z-p31262`;
  `make generate-drift` PASS 4/4 `20260920T201337Z-p31178`;
  `make generated-artifact-policy-check` PASS 3/3 `20260920T201544Z-p43789`.
- `make lint-markdown` PASS `20260920T201337Z-p31266`; completed-handoff
  documentation rerun PASS `20260920T202452Z-p24259`. `git diff --check` passed.
- Final review narrowed the scalar settlement key to the active scalar revision and other contributing scalar
  revisions (creation still includes all contributing capture authoring). An
  unrelated retained collection input must not split paste/departure settlement.
  The source-registry regression exercises that case; the final affected reruns
  below passed.

The visual-golden maintenance guide was reviewed. No existing capture contract,
layout, font, density, crop or renderer profile changed, and existing owner visual
rows passed. The local oversized-text alert supplies admission feedback without
a new visual fixture or a golden refresh.

## Acceptance assessment

| Row | Result | Evidence or scoped rationale |
| --- | --- | --- |
| A001 | PASS | Exact owner map above; adopted native editing, autosave, revision and mutation identity obligations govern changes. Routing and digest remain advisory/projection evidence. |
| A002 | PASS | G1–G6 complete bounded remedy/rationale/extension/risk/retirement assessments; one scalar input/settlement boundary, with neutral Adapter and queue guards. |
| A003 | PASS | Clean main/commit baseline, local guides, stack, actual consumers and authored routes inspected; final dirty files are only this slice. New feedback uses the existing application language, with no localization registry change. |
| A004 | PASS | No design tokens, component style literals or token/theme/density registry added. Existing byte-limit helper reused. |
| A005 | N/A | Theme selection and supported palette are unchanged; this task adds no theme behavior. |
| A006 | N/A | Density ownership and geometry are unchanged; existing Adapter/visual regressions and AC-043 still pass. No density redesign claim. |
| A007 | PASS | Native paste/first-input promotion creates exactly one record and transfers current authoring; continued typing targets it. Readonly and Evidence creation-ordering regressions pass. |
| A008 | N/A | No responsive accessor, breakpoint, viewport fallback or Inspector clamp changed. Existing bulk-tag layout/zoom regressions ran successfully. |
| A009 | PASS | Native multiline scrolling survives older receipt; existing shell/bulk layout regressions retain reachable controls. |
| A010 | PASS | Existing Inspector dispatch preserved; same-field Inspector/grid revisions stay independent, with readonly controls and detached source settlement verified. |
| A011 | PASS | Connected input, selection/caret, Find borrowing, current destinations, promotion and detached newer text covered by production attachments. |
| A012 | PASS | Fresh native revisions identify gestures; one paste/departure operation, exact captured replay bytes/transaction IDs and unchanged secure ID source. |
| A013 | PASS | Rejected draft retained; existing recovery path preserved; queue/replay tests and real lost response prove captured identity; failed refresh does not resend accepted write. |
| A014 | PASS | Native clipboard/history and existing grid entry/navigation/collection tests; exact validation drafts, Escape, IME and permission lifecycle regressions. |
| A015 | PASS | Existing field conflict/recovery model and UI untouched; owner regression rows pass. Generic rejection retains original recovery and precise local authoring. |
| A016 | PASS | Accepted-write/failed-query production case keeps Saved independent from stale data; closed readonly copies without mutation. Adapter data/interaction tests pass. |
| A017 | PASS | Existing runtime/lifetime and range authority regressions pass; account retirement clears source callbacks/settlement map; no authorization or concealment bypass added. |
| A018 | N/A | Evidence lifecycle/preview/overlay design is unchanged. Applicable paste precedence and capture ordering are separately PASS in Evidence's routed production case. |
| A019 | PASS | Owner-selected accessibility case, native keyboard/history, acceptance-gated focus, readable readonly and accessible local admission alert; no extra conformance claim. |
| A020 | PASS | Existing component/Adapter and bulk-tag layout/spacing/zoom regressions; native long multiline editing/scroll and rejection controls remain usable. |
| A021 | PASS | Stable native identity across ordinary updates, existing range/cycle/collection scenarios and four unchanged large-grid AC-043 qualifications. |
| A022 | PASS | Existing owner-selected visual rows passed without fixture/golden changes. No intentional rendering contract or new golden to reconcile. |
| A023 | PASS | New tests use stable incident/record/field/surface identities and existing UI selector facade; readonly uses the actual semantic row/field selector. No selector API changed. |
| A024 | PASS | Source/test diff audit and import-boundary check: no runtime, test, generator or release dependency on docs/Markdown. |
| A025 | PASS | Authored module.timeline/module.workbook catalog changes generated via Make; finalizer, drift and generated-policy checks pass. Two derived routing outputs only. |
| A026 | PASS | Compatibility and no-migration result above; existing source mutation/replay/authorization owners retained; optional Adapter parameter preserves other callers. |
| A027 | PASS | This completed tracker records sequential exits, evidence, failures/disposition, limits, skipped checks, exact files, rollback and review next action. All four workstreams have passed their binary exits. |


Additional intermediate failures were resolved without relaxing requirements:
`20260920T195603Z-p49119` typecheck found a nullable binding guard and a test error
with status nested in the wrong object; both were corrected. The earlier
`20260920T194106Z-p40260`, `20260920T194248Z-p46362` and
`20260920T194641Z-p60047` typecheck failures were the fixture-map/event/import
migrations already described. `20260920T195046Z-p67381` was the expanded Adapter
fixture's stale expected commit count; its corrected unchanged-source rerun
passed. All final terminal type/lint/import/generation checks pass.

The final key also includes the focused revision when its text is already saved:
that gesture can settle without a write and retire only its own draft, while
unrelated collection authoring remains retained. Focused source tests cover both
cases. Native regression rerun `20260920T201813Z-p50170` PASS 11/11, six cases;
focused ownership rerun `20260920T201811Z-p49539` PASS 5/5; typecheck/Biome PASS
`20260920T201812Z-p49849` / `20260920T201812Z-p49859`. Final verification after adding
the focused unchanged-value retirement assertion passed:

- Focused Timeline ownership slice PASS 5/5, `20260920T201953Z-p88265`.
- `make frontend-typecheck` PASS 2/2, `20260920T201955Z-p88542`.
- `make lint-biome` PASS 2/2, `20260920T201955Z-p88552`.
- Six production native clipboard cases PASS, 11/11 graph units,
  `20260920T201956Z-p88863`.

SC-04 exit: PASS. Final source and production reruns pass; all applicable
acceptance rows have evidence or a scoped N/A rationale. Final branch/commit
remain the baseline, with 37 changed/new paths belonging to this slice and no
unrelated edits. No implementation work remains. Documentation lint also passed, with its
retained artifact at
`.cartulary/test-results/20260920T202452Z-p24259/adhoc/lint-markdown/tool-run-summary.json`.
