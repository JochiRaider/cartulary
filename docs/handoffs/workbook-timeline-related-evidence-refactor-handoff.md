# Timeline Create Related Evidence handoff

## Baseline, authority, and scope

Execution baseline: `main`, `f5354af46f4f9a14ccec9fb74a7a143e6d12a930`,
clean worktree, revalidated before editing. The digest read order, domain and
design boundaries, and adopted Core owners were inspected during planning.
Core 01 §§3.3.4–3.3.7, 7.4.1A, 7.4.1, 7.4.4 (REQ-01-328), 18 and 18B own
discovery, create, PATCH, replay, metadata and references. Core 02 §§12–13 and
19 own links, Evidence custody/blob separation and Parties. Core 03 §§2.3A,
3–4, 6, 8, 13, 15–16 and Core 04 §2 own interaction, continuity,
collaboration, capture state, and authorization. Design §§10–15 own bounded
presentation and accessibility. Research, digest, implementation, tests and
prior handoffs are evidence, not behavioral authority. No executable artifact
may depend on Markdown.

Only Timeline `create_related.evidence` targeting `cartulary.view.evidence.v1`
and subsequent attachment to the original Timeline record are authorized.
Retain one unsent draft; replacement requires explicit discard. Navigation or
inspector closure during creation pauses linking for renewed review. The
adopted capture-state transition is `reviewed -> enriched`; the request's
reviewed-to-rough wording was clarified in the approved plan.

Uploads, blob inputs, screenshot creation, existing Evidence attachment/access,
standalone creation, lifecycle editing, detach, Notes, coordination artifacts,
and Task/Decision workflows remain regression boundaries. No dependencies,
specification or public-contract changes, backend implementation, digest edits,
commits, pushes, deployment, or analyst-data changes are authorized. Any
required backend/public-contract/specification prerequisite blocks dependent
work until separately authorized.

## Execution tracker

| Workstream | Dependency | State | Exit |
| --- | --- | --- | --- |
| TRE-01 baseline, owners, characterization, prerequisites | none | DONE | Baseline unchanged; owner audit and characterization support frontend implementation. |
| TRE-02 retained metadata authoring | TRE-01 | DONE | Retention, minima, fields and Party discovery tests pass. |
| TRE-03 attempts, receipts, partial-success recovery | TRE-02 | DONE | Exact stage requests, replay and independent receipts pass. |
| TRE-04 reconciliation, security, accessibility, integration | TRE-03 | DONE | Required browser, security, continuity and reviewed visual evidence passes. |
| TRE-05 final validation and completed handoff | TRE-04 | DONE | Finalization, affected checks and completed-handoff scope review pass. |

Only the active row may be IN_PROGRESS. Append paths, commands, results, risks,
disposition and next action before starting its successor. Never advance past
a blocked dependency. Tests are added alongside implementation.

## Gap ledger

| ID / classification | Authority | Remediation / affected areas | Rationale and long-term benefit | Compatibility / unresolved risk | Binary validation |
| --- | --- | --- | --- | --- | --- |
| G1 confirmed | Core 03 §2.3A; authorized retention | Evidence-specific retained owner and detachable inspector presentation replace selection-bound reducer lifetime. | Preserve metadata while invalidating stale readiness. | Other targets retain ownership; lifecycle races require evidence. | Exact values and original identity survive row/version/sheet/closure; replacement requires discard. |
| G2 confirmed | Core 01 REQ-01-328, §§18–18B; Core 02 §13 | Metadata request preparation preserves omission, initial-state admission and field errors. | Untouched clearable controls currently become null and suppress defaults; generic feedback loses field detail. | Standalone creation unchanged; normalization is a test obligation. | Blank/default/Party-only fails; non-title minima pass; untouched fields omitted; invalid input retained. |
| G3 confirmed | Core 01 §18B; Core 02 §19 | Replace empty Timeline Party options with authorized candidate paging. | Existing reference inputs become usable without copied Party workflows. | No implicit Party creation; inaccessible targets remain explicit. | Loading/paging/empty/error/cancel/off-page selection pass; text and IDs stay independent. |
| G4 confirmed | Core 01 §3.3.5; Core 03 REQ-03-301–302 | Immutable creation/link attempts, complete envelopes and stage recovery. | Current adapter creates new keys, truncates receipts and returns link failure to creation. | Public topology unchanged; late results and uncertainty need evidence. | Link recovery never creates Evidence; exact replay preserves body, route and key. |
| G5 confirmed | Core 03 collaboration and continuity | Accept receipts independently of selection; retain reconciliation debt. | Current acceptance effects depend on current selected subject. | No manufactured counts/text; derived counts may change without Evidence version change. | Accepted results survive navigation and failed refresh; recovery reads only and preserves newer state. |
| H1 characterization obligations | Core 03 §§3–4, 6, 15; Core 04 §2; design | Cover activation, prior saves, conflicts, source removal, authority changes, projection order and focus. | Resolve integration risks through evidence. | Hypotheses until reproduced; required failure blocks dependent row. | All required scenarios pass or have explicit prerequisite disposition. |

## TRE-01 investigation

Inspected the requested Timeline hook, reducer/model, adapter and request
builders; inspector authoring and feature dispatch; shell/runtime lifecycle,
contextual creation and explicit PATCH owners; transport observation, candidate
paging and source reads; view/OpenAPI projections and Evidence/Timeline source
admission, idempotency, links and projections; related workflow tests and the
Timeline Workflow scenario in `sentinel.spec.ts`.

Evidence creation and Timeline attachment are separate public commits.
`create_related.evidence` has no seeds and requires no confirmation. Metadata
creation admits a qualifying non-title value or explicit requested state;
omitted defaults and Party IDs alone do not qualify. No blob is required for
requested, pending_receipt, received or quarantined initial Evidence.
Timeline attachments validate active same-incident Evidence without requiring
an available blob. Source PATCH derives attached_evidence links and refreshes
Evidence projections. Successful replay precedes fresh concurrency evaluation.
No blocking prerequisite was found at execution HEAD.

Planning characterization on this exact source passed:

- `make test-slice OWNER=web.workbook ROWS=web.workbook.regression.timeline_create_related_workflow_stale_effects_4b8ce2aa71,web.workbook.regression.inspector_related_record_model_20a3b73f6d`:
  PASS, `.cartulary/test-results/20260912T232030Z-p82881`.
- `make test-slice OWNER=module.evidence ROWS=module.evidence.unit.create_signal_and_initial_lifecycle_matrix_4f9a8b671c,module.evidence.unit.evidence_mutation_admission_and_replay_hashing_are_o_48d36a71bc,module.evidence.unit.evidence_direct_party_references_admit_exact_sta_d51415ee2a`:
  PASS, `.cartulary/test-results/20260912T232030Z-p82882`.
- `make test-slice OWNER=module.timeline ROWS=module.timeline.support_unit.unit_capture_state_and_payload_helpers_772e7f2e60`:
  PASS, `.cartulary/test-results/20260912T232044Z-p84090`.
- `make service-backed-test-slice OWNER=module.evidence ROWS=module.evidence.store.attached_evidence_create_and_patch_aa33ea0168`:
  PASS, `.cartulary/test-results/20260912T232044Z-p84099`.

The existing Timeline Workflow browser row is owned by `module.workbook`:
`module.workbook.browser_stateful.verify_timeline_inspector_workflow_create_relate_7fc4833af4`.
These passes characterize existing behavior; they do not close new recovery
requirements.

Execution task guides for `web.workbook`, `module.evidence` and
`module.timeline`: PASS. `git status --short` was empty before this handoff.
TRE-01 disposition: DONE; no prerequisite requiring separate authorization.
G1–G5 are confirmed; H1 remains an integrated-verification obligation. Next:
retained metadata authoring and target-owned Party discovery.

## TRE-02 retained authoring

Added the Evidence-specific retained owner, metadata preparation, form and
staged Party controls under `apps/web/src/workbook/features/evidence/`.
Extracted neutral authoring discovery, paged identity lookup and immutable
value helpers; the existing Task/Decision reader remains a compatibility
export. Authored source ownership and the focused authoring row are registered.
Untouched optional values remain omitted. Lifecycle starts unset; explicit
Requested qualifies. Metadata admission, NFC/string bounds, exact Party IDs,
calendar-valid zoned timestamps and reserved storage references are checked
without rewriting the editable draft. Suspension hides protected state;
account replacement retires it. Inspector/runtime dispatch integration follows
in TRE-03.

- `make test-slice OWNER=web.workbook ROWS=web.workbook.regression.timeline_related_evidence_authoring`:
  PASS (5 tests), `.cartulary/test-results/20260912T234234Z-p11239`.
- `make frontend-typecheck`: PASS,
  `.cartulary/test-results/20260912T234234Z-p11306`.
- Contextual authoring/discovery regression rows passed in
  `.cartulary/test-results/20260912T234141Z-p9722`; that combined run initially
  failed the new tests because unsupported assertion helpers were used.
  Corrected helpers and typed mocks; the focused rerun above passed.
- Initial routing checks rejected unsorted row IDs/titles; authored routing
  ordering was corrected before execution.

Disposition: DONE for retained authoring. Remaining risk: operation races,
source saves, reconciliation, integrated security and visuals are not yet
verified. No prerequisite or compatibility change was introduced. Next:
immutable stage attempts, complete receipts and partial-success recovery.

## TRE-03 attempts and recovery

The retained Evidence owner now captures separate immutable creation and link
attempts, complete correlated envelopes and independent reconciliation states.
Uncertain replay uses the original body, route and transaction identity. An
accepted creation removes the editable draft and becomes a recovery checkpoint;
link rejection cannot reopen creation. Navigation pauses automatic linking.
Definitive field rejection retains input and specific errors. Confirmed
`client_txn_conflict` requires explicit new-ID review. Existing collection conflict
resolution now supplies its complete returned envelope to the checkpoint.

Runtime composition, authority suspension/retirement, Timeline prior-save
coordination, source reads and the existing workbook recovery area are bound.
Explicit attempts contribute status without consuming autosave queue capacity.
Shared ordinary mutation-envelope validation was extracted for contextual
Task/Decision creation and the two Evidence stages. The attachment request
reuses the existing Timeline collection builder.

- `make test-slice OWNER=web.workbook ROWS=web.workbook.regression.timeline_related_evidence_authoring,web.workbook.regression.timeline_related_evidence_recovery`:
  PASS (17 tests), `.cartulary/test-results/20260913T000323Z-p25394`.
- The above rows plus contextual Task/Decision recovery: PASS,
  `.cartulary/test-results/20260913T000128Z-p19435` (before two additional
  Evidence transport/conflict tests, included in the later run).
- `make frontend-typecheck`: PASS,
  `.cartulary/test-results/20260913T000323Z-p25460`.
- `make frontend-import-boundary-check`: PASS,
  `.cartulary/test-results/20260912T235509Z-p15509`.
- `make format`: PASS, `.cartulary/test-results/20260913T000156Z-p20444`.
  No unrelated source changes were introduced by formatting.

Intermediate typechecks identified and resolved a literal-union test fixture
and the distinction between conflict keep-saved and mutation envelopes. These
were frontend typing/projection issues, not backend prerequisites.

Disposition: DONE for stage ownership and focused recovery. Remaining risk:
real-service integration, HTTP/socket ordering, newer projection observations,
visibility, focus and visual effects require TRE-04 evidence. Next: integrated
browser scenarios and affected regression boundaries, then final verification.

## TRE-04 integrated evidence

Removed the legacy Evidence create/link branch; Task/Decision and other targets
retain their existing owners. The retained owner is composed by the mutation
runtime and participates in Timeline save coordination without autosave capacity.
Collection-resolution receipts and refreshed conflict tokens remain associated
with the original link checkpoint. A second creation from the same Timeline
source is blocked while accepted Evidence still needs linking.

Socket observations and HTTP receipts are correlated independently. Refresh
work is invalidated by newer observations and rereads equal-version Evidence
counts. Successful scoped reads distinguish unavailable projections from failed
reads. History uses its existing bounded reader with a narrow related-projection
entry that preserves incident history authority when one record is unavailable.
Authentication and incident access failures still suspend protected state.

Real-service fixtures use public routes and isolated incidents. Response loss is
injected after forwarding the write to the real server. Public queries and full
history reads verify raw Timeline text, enriched capture state, existing links,
Evidence lifecycle/blob absence, derived counts, exact replay envelopes and no
repeated relationship or history effect.

| Selection | Result / run root suffix under `.cartulary/test-results/` |
| --- | --- |
| `module.evidence.browser.timeline_related_creation_replay` | PASS, `20260913T000926Z-p31930` (combined run's conflict assertion failed separately). |
| `module.evidence.browser.timeline_related_partial_creation` | PASS, `20260913T002931Z-p5387`. |
| `module.evidence.browser.timeline_related_role_loss` | PASS, `20260913T002440Z-p34501` (combined run's conflict history assertion failed separately). |
| `module.timeline.browser.related_evidence_link_replay`, `related_evidence_original_source`, `related_evidence_projection_recovery` | PASS, `20260913T001545Z-p96983`. |
| `module.timeline.browser.related_evidence_unavailable` | PASS, `20260913T002440Z-p34510`: deleted/superseded source and deleted target. |
| Existing `module.workbook` Timeline Workflow browser scenario | PASS, `20260913T003027Z-p40143` (new a11y fixture failed separately). |
| Existing `module.workbook` Task/Evidence inspector frontend row | PASS, `20260913T001545Z-p96985`. |
| Evidence authoring/recovery and Timeline runtime binding units | PASS, `20260913T002931Z-p5392`. |
| Evidence authoring/recovery and stale-workflow hook units | PASS, `20260913T003332Z-p8843`. |
| Evidence authoring/recovery and history-owner units | PASS, `20260913T004022Z-p14175`. |
| Frontend typecheck | PASS, `20260913T004025Z-p14864`. |

Browser assertions were corrected to count history entries separately from
change sets: each attachment contributes a Timeline patch and relationship
entry to source history; Evidence history gains the relationship entry. These
were test-expectation defects, not a changed backend contract. The initial
conflict run also exposed presentation invalidation on harmless source refresh;
actual row/sheet/closure changes now detach presentation, while version changes
invalidate review independently.

New a11y/visual fixtures initially omitted required Party kind and the a11y
selector named a nonexistent timestamp field. Both fixtures were corrected.
Keyboard validation then exposed unnecessary source review after a local field
error; correcting metadata now retains readiness when context is unchanged.
Initial visual runs have not promoted goldens. A selected browser invocation
started before generation finished was rejected for its missing generated group;
generation and the later selected runs resolve that routing failure.

A11y now passes (`20260913T004400Z-p78008`), including 1280/768/390 widths,
200% zoom, text spacing, independent Party text, picker Escape/cancel, field
error descriptions and retained recovery. Review buttons use `aria-disabled`
while reads are pending so keyboard focus remains on the initiating control.
The source-switching hook regression passes at `20260913T004535Z-p44115`;
changing create-related targets detaches presentation while preserving Evidence
metadata. Import boundaries pass at `20260913T005035Z-p48882`.

All four related-Evidence `module.timeline` browser rows pass together at
`20260913T005033Z-p48036`. The three new `module.evidence` browser rows and
both screenshot attachment/atomic Timeline creation boundary rows pass in
`20260913T004536Z-p44375`. That combined run failed only the existing Evidence
access stateful row after membership revocation; its standalone repeat also
failed at `20260913T004711Z-p79017`. The isolated clean baseline control at
`/tmp/cartulary-tre-baseline` confirms
the identical failure on HEAD `f5354af46f4f9a14ccec9fb74a7a143e6d12a930`:
`.cartulary/tre-baseline-control/20260913T005032Z-p47519` (copied intact from
the control's retained run). `git status --short` in the control was empty;
the task-created scratch worktree was then removed. Pre-existing worktrees
were not changed. The test expects the Evidence grid
at `evidence-integration.spec.ts:524` after membership revocation; its
`expectActiveEvidenceSurface` assertion fails at line 778 in both sources.
Disposition: pre-existing excluded-boundary failure, not caused by this seam
and not reported as a pass. No backend/spec/public-contract prerequisite is
introduced; the unrelated test/product-policy review is deferred. Baseline
bootstrap required `make frontend-install` before the control could run.

The full ordinary visual run `20260913T004400Z-p78208` reaches all 236 capture
intents. Reconciliation accounts for 231 existing active goldens, five missing
new goldens, zero orphan/ambiguous mappings and zero unresolved registered
fixtures. Earlier ordinary attempts `20260913T003027Z-p40341` and
`20260913T003721Z-p48817` exposed fixture setup and the visual reconciliation
schema's required 12-hex scenario identity. Authored scenario identities and
Make derivatives are corrected; no schema was relaxed.

All five actual screenshot attachments were individually inspected in that
run's `timeline-evidence-review/`. Review removed duplicate partial-success
copy. The updated five ordinary attachments at `20260913T004952Z-p16514` were
individually reviewed again: metadata/default lifecycle, independent Party
controls, source identity, receipts, recovery controls and bounded scrolling
are legible with no unintended overflow. Its narrow selection lists unselected
existing captures as orphans; those are not deletion candidates. The complete
accounting remains the full ordinary run above.

Accepted visual trigger: new retained authoring and partial-success states.
Semantic owner row: `module.workbook.visual.timeline_related_evidence`;
scenario `scenario_46f536da95f0`, project `chromium`, are authored in
`tools/test_families/module.workbook.json`.
The five stable capture IDs are `timeline-related-evidence-authoring`,
`timeline-related-evidence-authoring-narrow`,
`timeline-related-evidence-party-narrow`,
`timeline-related-evidence-partial-narrow`, and
`timeline-related-evidence-partial`. Each golden is its capture ID plus
`-linux.png` under `apps/web/e2e/workbook.visual.spec.ts-snapshots/`.
These are active nonregistry captures, permitted by the visual guide; existing
registered fixtures remain unchanged. Desktop is 1280×720, narrow is 768×640,
zoom 100%, inherited Timeline compact density and the pinned dark theme.
Existing fixture-local preferences, normalization, anchor and viewport capture
infrastructure are reused. No tolerance, renderer, font, mask or screenshot
scope change is used to obtain a pass.

`make browser-e2e-visual-update`: PASS, 12/12 units,
`20260913T005219Z-p14982`. All 236 goldens reconcile active with zero
missing/orphan/ambiguous mappings or unresolved registered fixtures. Exactly
five new PNGs and the Make-generated golden manifest changed; all 231 existing
PNGs are byte-identical. Every promoted image was individually inspected again.
No unintended clipping, overflow or control-state changes were accepted.
The first fresh full ordinary validation passes, 12/12 units,
`20260913T005755Z-p56142`: all 236 captures reconcile with no missing, orphan,
ambiguous or unresolved registered fixture. The concurrent second attempt
`20260913T005755Z-p56147` failed at service readiness before browser execution
(`failure_class=infra`, `service_readiness_timeout`). At this checkpoint its
isolated rerun was pending. Final source review also connected newer ordinary
HTTP/history observations to the same monotonic refresh invalidation used by
sockets. TRE-04 stayed active until the exit evidence below; TRE-05 had not
started.

## TRE-04 exit and gap dispositions

Second fresh full ordinary visual validation: PASS, 12/12 units,
`20260913T010524Z-p33208`. Together with `20260913T005755Z-p56142`, both fresh
ordinary passes validate all 236 captures after promotion. Final reconciliation
has zero missing/orphan/ambiguous mappings and no unresolved registered fixture.
The updated Evidence authoring/recovery plus history-owner unit selection passes
at `20260913T010523Z-p32901`; the stale-workflow hook row passes at
`20260913T005756Z-p56353`. Typecheck passed at `20260913T005319Z-p49349` before
final removal of obsolete legacy linking types and the last refresh-observation
check; final source typecheck is part of TRE-05.

| Gap | Final implementation disposition and binary evidence |
| --- | --- |
| G1 | DONE: exact raw metadata and original account/incident/source/feature survive presentation changes; replacing retained work requires discard. Authoring and hook units plus real navigation pass. |
| G2 | DONE: every owner-qualified minimum, omitted versus explicit lifecycle, four metadata initial states, timestamp/string/reference errors and payload omission pass; real collector-only creation succeeds. |
| G3 | DONE: authorized Party paging, loading, empty, retry/error, off-page selection and cancellation pass; keyboard controls preserve independent text/reference values. |
| G4 | DONE: immutable separate requests, complete receipts, malformed-response uncertainty, secure keys, duplicate activation, exact replay and partial-success link-only recovery pass with real writes/history. |
| G5 | DONE: independent acceptance and reconciliation, original-source reads, HTTP/socket ordering, newer versions, equal-version derived counts and read-only recovery pass. |
| H1 | DONE for this seam: real role loss, removed/superseded source, unavailable target, source edits/navigation and refresh failure pass; runtime suspension/retirement and focus tests pass. The excluded Evidence-access failure is reproduced unchanged at baseline and explicitly deferred above. |

TRE-04 disposition: DONE. No specification, public-contract or backend
prerequisite remains. Known limits are supported runtime retention, existing
compact-inspector presentation and the independently reproduced excluded-boundary
failure. No prohibited scope or analyst-data change occurred. Next: finalization,
affected broad frontend verification, final handoff and finished-document checks.

## TRE-05 final verification completed

`env -u RESULTS_DIR make agent-finalize` passed
at `20260913T011017Z-p68168` before broader end-of-run verification. Its
`unit-artifacts/finalize-summary.json` records passing JSON shape, catalog/tier
routing and generated-structure refresh/drift. Canonical retained-evidence and
scheduler drift maintenance were skipped with `results-dir-not-provided`:
no exact-source successful full warm check has been established. Focused,
visual and baseline-control runs do not qualify.

| Final command | Result / run root suffix under `.cartulary/test-results/` |
| --- | --- |
| `make frontend-typecheck` | PASS on final source, `20260913T011932Z-p72567`; earlier pass `20260913T011149Z-p72295`. |
| `make lint-biome` | PASS on final source, `20260913T011932Z-p72572`; corrected pass `20260913T011255Z-p93065`. Initial run `20260913T011149Z-p72349` rejected one unnecessary non-null assertion; the captured target ID now uses the existing narrowed receipt. |
| `make frontend-import-boundary-check` | PASS, `20260913T011149Z-p72343`. |
| `make generated-artifact-policy-check` | PASS, `20260913T011149Z-p72089`. |

Final Timeline Workflow and Evidence accessibility rerun:
`make service-backed-test-slice OWNER=module.workbook ROWS=module.workbook.browser_stateful.verify_timeline_inspector_workflow_create_relate_7fc4833af4,module.workbook.accessibility.timeline_related_evidence`:
PASS, 13/13 units, `20260913T011256Z-p93675`.

The initial full `make frontend-unit` run, `20260913T011149Z-p72294`, passed
596/597 units. The sole failure was the explicit owner-local wire-intent path
list in `apps/web/src/testing/sourceOwnershipPolicy.test.ts`; four new Evidence
owner/model/test paths were missing. Those exact paths are now admitted without
weakening the transaction-identity or presentation-boundary assertions.
`make test-slice OWNER=web.architecture ROWS=web.architecture.boundary_support.source_ownership_policy_suite_80cf87ef19`:
PASS, 2/2 units, `20260913T011906Z-p71624`.
The complete `make frontend-unit` rerun passes all 597/597 execution units at
`20260913T011932Z-p72541`; no frontend unit row remains failing.

TRE-05 disposition: DONE. All requested implementation, ownership/routing,
service scenarios and visual exit evidence are complete. No blocking
specification, public-contract or backend prerequisite was found or implemented.
The unrelated existing Evidence-access browser failure is preserved with its
clean-baseline reproduction and deferred disposition above. Post-completion
`make lint-markdown` passes at `20260913T012335Z-p36658`;
`git diff --check` and final scope review pass against the completed handoff.
Markdown lint and whitespace/scope checks are repeated after recording these
results. No additional product workstream follows.

Final scope review: 66 intended paths (47 frontend source/test files, 10 browser
fixture/scenario/PNG files, eight authored or Make-derived tool files, and this
handoff). Branch remains `main`, HEAD remains
`f5354af46f4f9a14ccec9fb74a7a143e6d12a930`. Initial worktree was clean. No
backend, public contract, adopted specification, digest, dependency/lockfile,
analyst data or unrelated worktree change is included. No commit, push or
deployment occurred.

Skipped broader verification: full repository `make check`, `make ci`,
`make release-check`, full backend/security suites and unrelated browser
scenarios. This frontend seam is covered by the complete frontend unit suite,
focused owner/backend characterization, real-service mutation scenarios,
accessibility and the full visual suite. A whole-repository green result is
not asserted. Retained-run maintenance remains skipped because `RESULTS_DIR`
was unset; focused successful runs do not qualify as full warm evidence.

## Final implementation map

- Retained draft, stage transitions and reconciliation:
  `apps/web/src/workbook/features/evidence/WorkbookTimelineRelatedEvidenceOwner.ts`,
  `timelineRelatedEvidenceModel.ts`, and `timelineRelatedEvidenceOperation.ts`.
- Detachable authoring and recovery: `TimelineRelatedEvidenceForm.tsx`,
  `RelatedEvidencePartyControl.tsx`, `TimelineRelatedEvidenceRecovery.tsx`,
  `TimelineRelatedEvidenceContext.ts`, and
  `useTimelineRelatedEvidenceAttachment.ts` in that Evidence feature directory.
- Shared discovery/read/transport mechanics:
  `apps/web/src/workbook/ports/WorkbookAuthoringReadPort.ts`;
  `createWorkbookAuthoringReader.ts`, `readWorkbookAuthoringRecord.ts`,
  `sendWorkbookRecordMutation.ts`, and
  `createTimelineRelatedEvidenceTransport.ts` under workbook adapters;
  `apps/web/src/workbook/utils/freezeWorkbookValue.ts`.
- Runtime integration: `WorkbookShell.tsx`, `WorkbookMutationRuntime.ts`,
  the shell infrastructure and Timeline composition/hooks, collaboration,
  conflict-resolution receipts and the history owner's scoped projection read.
  Legacy Evidence creation/linking is removed from the related-record adapter.
  Task/Decision uses compatibility exports for extracted neutral mechanics.
- Verification: Evidence authoring/recovery tests, existing inspector/hook/runtime
  regressions and `apps/web/src/testing/sourceOwnershipPolicy.test.ts`;
  `apps/web/e2e/timeline-related-evidence.spec.ts`, its isolated support fixture,
  existing sentinel, accessibility and visual scenario files.
- Authored ownership/routing: `tools/frontend_source_ownership.json` and
  `tools/test_families/{web.workbook,module.workbook,module.evidence,module.timeline}.json`.
  Make-derived changes: browser batch manifest, topology render index and
  frontend visual golden manifest. Five reviewed PNGs are added; existing
  goldens are unchanged.

## Compatibility and rollback

Keep public routes, schemas, field identities, server defaults, authorization,
custody and history stable. Retention is memory-only within the supported
account/incident runtime; reload, tab closure and browser persistence are not
guaranteed. Rollback reverts frontend and associated ownership/routing/visual
changes only. Preserve committed Evidence, attachments, receipts, custody
history and revisions. Do not delete or detach to compensate for partial
success, or discard live recovery state as a rollback step. Reverting frontend
code does not reverse committed operations.
