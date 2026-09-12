# Workbook Assessment authoring refactor handoff

## Baseline and authority

Execution starts on clean `main` at
`c6494175f95ae0f2862c016751fea0811e42133d`, rechecked after planning.
The user authorized the Assessment authoring, discovery, and retained append
recovery seam. One editable draft is retained per incident runtime; replacement
requires explicit discard. Attempts and authoritative receipts are separate.

Read `AGENTS.md` and the digest's prescribed read order, Core 00 precedence,
Core 01 ordinary create/replay/query/discovery and §7.4.7 (REQ-01-332–335),
Assessment inspector bindings, Core 02 §10.3, Core 03 §16.3
(REQ-03-251–254,304–305), collaboration/continuity/pending-work owners, Core 04
authorization, design §§10,12–15, and the visual golden maintenance guide.
The digest and existing handoffs are evidence, not behavior authority.
No specification, public-contract, or backend prerequisite was found.

Inspected Assessment features, surface, model, support hook, surface query,
mutation commands, runtime/collaboration, subject queries, candidate picker,
Assessment view/OpenAPI projections, backend admission/facade/idempotency,
model/controller/shell tests, and `workbook.assessments.spec.ts`.

## Execution tracker

| Workstream | Dependency | Status | Exit criteria |
| --- | --- | --- | --- |
| AA-01 baseline and owner audit | none | DONE | Gaps characterized; no public/backend prerequisite. |
| AA-02 discovery and authoring | AA-01 | DONE | Stable subjects, honest paged support, contract-backed payload tests pass. |
| AA-03 retained append operations | AA-02 | DONE | Immutable attempts, receipts, duplicate guard and exact replay tests pass. |
| AA-04 integrated behavior | AA-03 | DONE | Reconciliation, authority, continuity, browser and visual evidence pass. |
| AA-05 final validation | AA-04 | DONE | Final verification, scope review and completed handoff. |

Only the active row may be IN_PROGRESS. Record exit evidence and next action
before advancing; never advance past a blocked dependency.

## Gap ledger

| ID/status | Authority | Remediation and affected areas | Rationale and long-term benefit | Compatibility impact | Unresolved risk and binary validation |
| --- | --- | --- | --- | --- | --- |
| G1 confirmed in code | Core 01 §7.4.7; Core 03 continuity; requested subject stability | Remove first-row subject defaults/replacement effects in surface/controller; independent paged subject reader. | Selected identity survives changing inventories. | Standalone requires deliberate subject selection. | PASS iff filtering, paging, refresh, disappearance and type changes never silently select another record. |
| G2 confirmed in code | REQ-03-304; Core 01 paging; design §§10,12 | Replace array-only support hook and Timeline runtime normalization with narrow candidate reader and explicit async states. | Honest discovery independent of Timeline editor fields. | Timeline-only discovery; existing other support remains readable. | PASS iff failed reads preserve selected IDs, later pages are reachable and candidates require only identity/display text. |
| G3 confirmed in code | REQ-01-332–335; REQ-03-251–254 | Contract-backed admission and scalar/support validation in Assessment model. | Prevent undeclared writes while retaining editable invalid input. | Existing wire grammar and omitted defaults preserved. | PASS iff minima, band scores, timestamp omission/validation and 1–64 action bounds hold. |
| G4 confirmed in code | REQ-01-070; REQ-03-301,305; runtime lifecycle | Replace component submission state with Assessment runtime owner, frozen attempts and complete receipts. | Navigation and uncertain outcomes cannot lose results or regenerate writes accidentally. | Internal frontend interfaces only; no browser persistence. | PASS iff duplicate activation sends once, replay preserves exact body/key and detached acceptance remains recoverable. |
| G5 confirmed in code | REQ-03-035,096–098,283,305 | Accept before refresh, retain refresh debt and protect query/socket/receipt version observations. | Accepted appends stay accepted and newer state cannot regress. | Filters, selection and history remain stable. | PASS iff refresh failure never resubmits and both HTTP/socket orderings are monotonic. |
| H1 hypotheses to exercise | Core 03 inspector/continuity; Core 04 authority; design accessibility | Characterize stale targets, roles, detached recovery, picker cancellation, focus and containment. | Resolve race/security risks without claiming unobserved baseline defects. | Existing suspension/retirement and semantic controls. | PASS iff named scenarios disclose no protected state, target no wrong record, lose no retained work and steal no late focus. |

## Evidence log

### AA-01 start

Rechecked branch, HEAD and clean status. Re-ran
`make task-guide ROLE=module-author OWNER=web.workbook` and
`make task-guide ROLE=module-author OWNER=module.assessments`: PASS.
Prior planning evidence is exact-source evidence for this unchanged HEAD:

| Command/selection | Result | Run root under `.cartulary/test-results/` |
| --- | --- | --- |
| `make test-slice OWNER=web.workbook` selected Assessment model/controller/shell rows | PASS 5/5 | `20260912T142422Z-p74768` |
| `make test-slice OWNER=module.assessments` query/model/follow-on/admission rows | PASS 5/5 | `20260912T142424Z-p74999` |
| `make service-backed-test-slice OWNER=module.assessments` create/replay/publication and validation rows | PASS 3/3 | `20260912T142534Z-p76890` |
| `make test-slice OWNER=web.networkflow ROWS=web.networkflow.regression.exploration_focus` | PASS 2/2 isolated | `20260912T142459Z-p75948` |
| `make test-slice OWNER=module.workbook ROWS=module.workbook.frontend_unit.verify_inspector_selection_tab_state_details_rel_d2dc82a4bb` | FAIL 1/2, unchanged Timeline assertion | `20260912T142741Z-p94891` |

The Timeline assertion expects `timeline.mark_reviewed` absent at
`WorkbookShell.inspector.test.tsx:607`. The Party handoff additionally records
intermittent Network Flow focus failure under broader runs. Preserve both
limitations; isolated success does not establish an all-green baseline.

### AA-01 exit / AA-02 start

The inspected public contract already supplies paged view queries, atomic support
creation and complete original-result replay. All confirmed gaps have bounded
frontend remedies. Risks H1 remain integration test obligations, not prerequisite
blockers. Next: independent candidate discovery and contract-backed authoring.

### AA-02 exit / AA-03 start

Added AssessmentCandidateReadPort, createAssessmentCandidateReader,
useAssessmentCandidates and AssessmentDiscovery. Removed the array-only support
hook and its Timeline model dependency; surface/composition now query independent
Host/Identity candidates. Subject type changes clear only subject identity.
Support choices stage until Apply; Escape preserves the draft. Admission consumes
create discovery and scalar validation while handling nullable confidence under
its Assessment create contract rather than existing-row clearability.

Authored source ownership and focused test rows were updated, then `make generate`
passed at `20260912T144501Z-p780`. Initial typecheck and focused runs exposed
integration leftovers, invalid fixture identities and nullable-confidence
admission; those change-related failures were repaired, with no weakened product
assertions. Failure roots: `20260912T144438Z-p99846`,
`20260912T144551Z-p4201`, `20260912T144704Z-p5304`.

- `make test-slice OWNER=web.workbook ROWS=web.workbook.regression.assessment_discovery,web.workbook.regression.assessment_admission,web.workbook.regression.workbookshell_assessments_suite_f2e4a841a2`: PASS 4/4 at `20260912T144754Z-p6402`.
- `make test-slice OWNER=module.assessments ROWS=module.assessments.frontend.the_assessment_model_owns_confidence_mapping_and_f155fcd292,module.assessments.frontend.the_assessment_surface_preserves_stable_selectio_ab289a0995`: PASS 3/3 at `20260912T144754Z-p6409`.
- `make frontend-typecheck`: PASS 2/2 at `20260912T144844Z-p7872`.

G1–G3 have focused evidence; security invalidation and browser integration remain
AA-04 obligations. No prerequisite appeared. Next: replace component-lifetime
submission with retained immutable operations and receipts.

### AA-03 exit / AA-04 start

Added `WorkbookAssessmentAuthoringOwner`, immutable `assessmentOperation` types,
`createAssessmentAppendTransport` and `AssessmentAppendRecovery`. The creation
hook now owns presentation attachment only. The incident runtime owns drafts,
explicit-operation accounting, exact replay, authoritative envelopes and refresh
debt. Command assembly exposes capture/send instead of regenerating requests.
Shell infrastructure shares its current-authority reader with the Assessment
owner; runtime suspension/retirement includes retained Assessment state.

Confirmed rejection retains editable values/support. Uncertainty reserves the
original attempt. Acceptance is recorded before refresh, including detached or
late responses; accepted recovery cannot issue another create. Explicit discard
and resume replace implicit draft resets. Shell fixtures now provide complete
incident resources required by current-authority admission.

- `make generate`: PASS at `20260912T150024Z-p14324` and `20260912T150310Z-p21837` after authored selector updates.
- `make test-slice OWNER=web.workbook ROWS=web.workbook.regression.assessment_append_recovery,web.workbook.regression.workbookz_assessment_creation_controller_2f0f2487ec,web.workbook.regression.workbookz_mutation_runtime_semantic_command_ports_7b2f1d02a4`: PASS 4/4 at `20260912T150033Z-p17318`.
- Shell/owner/controller selection: PASS 4/4 at `20260912T150217Z-p19962`.
- `make test-slice OWNER=module.assessments ROWS=module.assessments.frontend.the_assessment_surface_preserves_stable_selectio_ab289a0995`: PASS 2/2 at `20260912T150241Z-p20874`.
- `make frontend-typecheck`: PASS 2/2 at `20260912T150241Z-p20952`.
- Expanded owner/controller deadline, secure-ID and editable-invalid-input cases: PASS 3/3 at `20260912T150327Z-p24884`.

Intermediate change-related failures were repaired: unsorted selectors
(`20260912T150014Z-p11220`), test type annotations
(`20260912T150033Z-p17383`, `20260912T150118Z-p19253`), and incomplete incident
fixture (`20260912T150110Z-p18662`). G4 has focused evidence. No prerequisite
appeared. Next: monotonic query/socket reconciliation, authority transitions,
continuity, real-service faults and visual review.

### AA-04 exit / AA-05 start

Integrated version/removal observations in `useAssessmentSurfaceQuery`, the
collaboration coordinator and the runtime's history subscription. Socket
observations precede transaction-echo suppression. Full query responses cannot
regress observed versions, and receipts never manufacture filtered membership.
The query's startup authority remains independent of mutation-owner readiness;
actual suspension aborts the query and hides protected observations. Relevant
subject/support changes invalidate review without changing retained selections.

Current authority is rechecked for dispatch, exact replay and read-only refresh.
Local subject/support failures trigger the existing authority reader rather than
assuming incident access loss. Presentation attachment fences selection changes,
close and late completion; global recovery retains receipts without navigation
or focus effects. Discovery now uses declared text equality as well as enum
filters. New controls use existing theme tokens and bounded scroll containers.

| Command/selection | Result | Run root under `.cartulary/test-results/` |
| --- | --- | --- |
| `make test-slice OWNER=web.workbook` discovery, reconciliation, socket order, retained recovery and shell rows | PASS 6/6 | `20260912T152311Z-p19557` |
| Assessment owner/controller rows, including detached inline feedback | PASS 3/3 | `20260912T152521Z-p91010` |
| `make service-backed-test-slice OWNER=module.assessments` recovery and filtered follow-on rows | PASS 13/13 | `20260912T151925Z-p72657` |
| `make service-backed-test-slice OWNER=module.assessments ROWS=module.assessments.browser.append_recovery` | PASS 11/11 | `20260912T152322Z-p21351` |
| `make service-backed-test-slice OWNER=web.design ROWS=web.design.accessibility.assessment_authoring` | PASS 11/11 | `20260912T152618Z-p96260` |
| `make service-backed-test-slice OWNER=module.workbook ROWS=module.workbook.visual.capture_inspector_details_relationships_evidence_a56cae74ea` | PASS 11/11 | `20260912T151912Z-p47328` |
| `make frontend-typecheck` | PASS 2/2 | `20260912T152311Z-p19684` |
| `make lint-biome` | PASS 2/2 | `20260912T152311Z-p19722` |
| `make frontend-import-boundary-check` | PASS 2/2 | `20260912T152733Z-p28969` |
| `make format` | PASS 2/2 | `20260912T152732Z-p28761` |
| `make generate` after authored selector updates | PASS | `20260912T152414Z-p83514` |

The real-service recovery spec captures a server-committed response and loses it
at the browser boundary. Exact replay returns the same receipt with one created
Assessment, one create change set, one support set and one supporting-link
history delta. The accepted-create/failed-refresh case proves the write request
list remains unchanged during recovery. Additional cases prove late acceptance
does not steal focus after navigation and rejected support validation creates
neither an Assessment nor partial supporting links. These use isolated harness
fixtures, not analyst data. Existing browsing, history and sentinel cases passed
in `20260912T151200Z-p35321`, whose two recovery failures were subsequently fixed.

Reviewed inline Playwright screenshots for subject discovery and authoring at
1280, 768 and 390 pixels; support discovery at 1280 and 768; and rejection,
uncertainty and accepted-refresh recovery. Keyboard, Escape, reduced motion,
200% zoom, validation associations and compact containment passed. The ordinary
shared-inspector visual reconciliation reports PASS, seven active captures,
zero missing or ambiguous goldens and no errors. Its 214 out-of-slice goldens are
not a cleanup claim. There was no existing Assessment golden to replace; no
golden or visual manifest changed and no promotion occurred. The two ordinary
post-promotion runs are therefore inapplicable. A full visual suite was not run.
Screenshots remain review evidence in the named browser reports.

Intermediate failures and disposition: the startup query guard
(`20260912T150805Z-p29659`) was repaired and characterized; a compact navigation
test and unfocusable recovery region (`20260912T151200Z-p35321`) were repaired;
the recovery test's incorrect assumption that creating a support link leaves
support history unchanged (`20260912T151621Z-p78065`) was replaced by the exact
one-delta assertion and unchanged history after replay. A discovery test exposed
the missing declared text-filter UI (`20260912T151622Z-p78304`). Missing fixture
imports, semantic labels, formatting and direct protocol-type import failures
were repaired; the latter now uses the existing workbook protocol facade.
Relevant failure roots include `20260912T151607Z-p77562`,
`20260912T151900Z-p43000`, `20260912T152106Z-p11300` and
`20260912T152618Z-p96371`. An unsupported Biome flag was rejected by Make input
validation; the ordinary supported target subsequently passed. No unrelated
assertion was weakened. No prerequisite appeared. Next: agent finalization,
broader scoped verification and the final handoff review.

### Acceptance disposition

| Ledger item | Result and implementation evidence | Remaining boundary |
| --- | --- | --- |
| G1 | PASS: independent paged subject reader, pinned selected identity, deliberate type reset; discovery and shell tests cover page/filter/order/refresh/disappearance. | Server remains authoritative for subject type, visibility, incident and state. |
| G2 | PASS: narrow candidate port, explicit failed/empty/stale/paging/retry states and staged support; failed reads retain IDs. | Interactive discovery is Timeline-only; inspection retains other readable references. |
| G3 | PASS: model admission tests cover required values, omitted defaults, editable timestamps, band mapping and support bounds; follow-on copies only subject. | No score editor, assessor override or ordinary semantic patch. |
| G4 | PASS: runtime owner captures immutable route/body/context/key, reserves synchronously, retains complete receipts, exact replay and detached results. | One editable incident-runtime draft; no browser persistence. |
| G5 | PASS: accepted receipt precedes refresh, query and socket high-water observations are monotonic; real-service refresh recovery sends no writes. | An accepted append outside filters remains successful without forced visibility. |
| H1 | PASS for characterized cases: stale review, session/role changes, suspension/retirement, close/resume, detached completion, picker cancellation and responsive keyboard access. | Inherited unrelated Timeline/Network Flow verification limitations remain separately tracked. |

### AA-05 exit

Authored ownership remains in `tools/frontend_source_ownership.json`; focused
verification routing is in `tools/test_families/web.workbook.json`,
`module.assessments.json` and `web.design.json`. The required Make-generated
derivatives are `tools/browser_e2e_batch_manifest.json` and
`tools/execution_topology_render_index.json`. Markdown lint includes this handoff
through `.markdownlint-cli2.jsonc`. No generator, contract or backend repair was
needed.

Ran `env -u RESULTS_DIR make agent-finalize` before broader final verification:
PASS 1/1 at `20260912T153240Z-p34779`. Its
`unit-artifacts/finalize-summary.json` records zero generated updates, passing
schema/catalog/tier/structure checks, and skipped retained-run selection,
canonical evidence, scheduler drift and performance maintenance because
RESULTS_DIR was not supplied. There is no qualifying successful full warm run.

| Final command | Result | Run root under `.cartulary/test-results/` |
| --- | --- | --- |
| `make frontend-unit` | FAIL 589/591; only the two inherited unrelated owner rows below | `20260912T153303Z-p39047` |
| `make frontend-typecheck` | PASS 2/2 | `20260912T153303Z-p39053` |
| `make frontend-import-boundary-check` | PASS 2/2; finalizer changed no source | `20260912T152733Z-p28969` |
| `make lint-biome` | PASS 2/2 | `20260912T153303Z-p39151` |
| `make generate-drift` | PASS 4/4 | `20260912T153303Z-p38703` |
| `make json-shape-check` | PASS 3/3 | `20260912T153303Z-p38772` |
| `make generated-artifact-policy-check` | PASS 3/3 | `20260912T153303Z-p38792` |
| `make test-catalog-check` | PASS, exit 0; also verified by finalizer | `20260912T153304Z-p39727` |
| `make lint-markdown` | PASS; `adhoc/lint-markdown/tool-run-summary.json` | `20260912T153345Z-p56918` |
| `make test-slice OWNER=web.networkflow ROWS=web.networkflow.regression.exploration_focus` | FAIL 1/2, inherited intermittent focus row | `20260912T153345Z-p56722` |
| `git diff --check`, branch/HEAD and scope review | PASS | local inspection |

The full frontend unit run's `run-summary.json` and `unit-results/` identify
exactly these failures:

- `module.workbook.frontend_unit.verify_inspector_selection_tab_state_details_rel_d2dc82a4bb`: the unchanged Timeline inspector assertion already failing at baseline.
- `web.networkflow.regression.exploration_focus`: the inherited intermittent focus failure, also reproduced by the narrow recheck.

All Assessment rows, including all newly routed discovery, admission, retained
recovery, reconciliation and socket-order cases, passed in that full frontend
run. No source or assertion in either unrelated failing test was modified.
Their disposition is follow-up outside this seam, not weakened validation or an
all-green claim. Existing source-preserving Party and other frontend owner rows
passed in this run.

Scope review includes tracked and untracked paths: frontend implementation and
tests, authored ownership/routing, the two Make-generated derivatives, Markdown
lint coverage and this handoff only. `main` and HEAD remain
`c6494175f95ae0f2862c016751fea0811e42133d`. No digest, adopted specification,
public contract, backend, lockfile or golden was edited. No dependencies or
commits were added. Full `make check`, full backend/browser/accessibility/visual
and release suites were not run: relevant narrow service/browser/design rows and
the broader frontend gate cover this bounded frontend seam. No benchmark or
release-readiness claim is made.

G1–G5 are repaired and H1's named scenarios have passing scoped evidence. No
dependency is blocked and no additional implementation is authorized. Next and
final action: rerun Markdown lint, `git diff --check` and scope review against
this completed handoff; report their terminal results with delivery, then stop.

## Compatibility and rollback

Preserve Assessment browsing, pivots, append-only fields/support, readable
non-Timeline support and history. No new creation surface, endpoint, dependency,
succession link, entity compromise property, operational posture, general
workflow engine or browser persistence. No digest, specification, public API or
backend edits are authorized by this seam. No commit, push, deployment or analyst
data modification.

Rollback reverts frontend behavior and corresponding tests/routing/visuals; it
performs no data cleanup, deletes no assessments/support/receipts/history and
does not undo committed appends. Local recovery has the supported incident
runtime lifetime, including its existing suspension and retirement boundaries.

## Final disposition

AA-01 through AA-05 are complete. Scoped Assessment behavior, real-service
recovery and reviewed accessibility/visual evidence pass. Broader frontend
verification retains only the two explicitly recorded unrelated failures.
RESULTS_DIR was unset and retained full-run maintenance was skipped. The seam
stops here with frontend changes and this handoff; rollback never deletes data
or undoes committed appends.
