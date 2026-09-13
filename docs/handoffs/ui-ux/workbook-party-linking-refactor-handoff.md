# Workbook Party linking refactor handoff

## Baseline and authority

- Execution branch: `main`.
- Execution HEAD: `f3f195d2ec02a9ab1f481fb4a61f6af7f84ca2ca`.
- Worktree at execution start: clean; matches the recommendation baseline.
- Scope: Evidence collector/source and Task Request requester text/reference
  pairs, their contextual creation, discovery, explicit patches and recovery.
- Read order: root AGENTS, localized digest README and ordered overlay, then
  adopted owners. Digest and previous handoffs are evidence, not authority.
- Core 01: query/paging §§3.3.4, 3.3.7; ordinary create/patch/replay
  REQ-01-057–070; field contracts REQ-01-328, 336, 499–502, 668–669;
  inspector bindings REQ-01-615–617.
- Core 02: REQ-02-224–232 and 271 govern Party identity, independent text,
  exact-match reuse without enrichment, and active claims.
- Core 03: REQ-03-035, 266–271, 274, 283, 292 govern concurrency,
  explicit actions and inspector continuity. Core 04 REQ-04-022–024 govern
  incident authorization. Design §§7–10 and 14 govern presentation and access.
- No specification, public-contract or backend prerequisite was confirmed by
  the audit. A subsequently discovered prerequisite blocks the dependent row
  until separately authorized; frontend work must not conceal missing authority.

## Dependent tracker

| Workstream | Dependency | Status | Exit criterion |
| --- | --- | --- | --- |
| PL-01 audit and characterization | none | DONE | Owner map, baseline, characterization and prerequisite disposition recorded. |
| PL-02 presentation and discovery | PL-01 | DONE | Three pairs, reviewed authoring and paged candidates pass focused tests. |
| PL-03 retained attempts | PL-02 | DONE | Separate immutable create/patch requests and receipts pass recovery tests. |
| PL-04 integration | PL-03 | DONE | Reconciliation, authorization, keyboard and real-service browser evidence pass; affected visuals reviewed. |
| PL-05 final validation | PL-04 | DONE | Terminal results dispositioned; seam checks and completed handoff scope review pass. |

Only one row may be IN_PROGRESS. Append paths, commands, results, risks,
disposition and next action before advancing. Never advance past BLOCKED.

## Gap ledger

All remediations are frontend-compatible with existing public operations and
owner hashing. No storage or wire migration is intended. Shared-runtime changes
retain Task lifecycle, Evidence, Timeline mentions, generic editing and
Communications Log collections as regression boundaries.

| ID / classification | Evidence and affected areas | Authority | Remediation, rationale and long-term benefit | Compatibility impact | Risk and binary validation | Disposition |
| --- | --- | --- | --- | --- | --- | --- |
| P01 confirmed | Generic create command hard-coded person and extracted email; PartyLinkControls and creation adapter replace it. | REQ-01-499–501, 668; REQ-02-229–230 | Contract-backed review with deliberate kind and opt-in email; preserves analyst intent and one admission path. | Explicit review replaces silent defaults; same create contract. | No create without kind; no unaccepted proposed email; source unchanged; reused Party unmodified. Public receipt cannot distinguish insertion from reuse. | RESOLVED: controls and real-service match scenario pass; acceptance says Party saved. |
| P02 confirmed | Generic relationships consumed referenceOptions.parties; createPartyLinkReader/usePartyCandidates now own discovery. | Core 01 query/paging; REQ-02-226, 231 | Bounded Party candidate reader and existing picker expose partial inventory honestly without widening global references. | Existing authorized query, sorting and opaque cursors; no global reference change. | Page beyond 100 reachable; failed/stale/empty pages differ; cursor loops rejected; socket invalidates selection. Large inventories may require another bounded read. | RESOLVED: focused paging and 101-Party browser scenario pass. |
| P03 confirmed | Old hook retained ID only and reset on navigation; runtime owner and PartyLinkRecovery retain results. | REQ-01-070; REQ-03-271, 292 | Complete creation and source results outlive presentation; recovery remains attached to its source. | Existing incident/account memory lifetime; no browser persistence. | Row, pair, sheet and inspector changes retain receipts; account replacement retires protected recovery. | RESOLVED: owner and navigation/lost-response browser scenarios pass. |
| P04 confirmed | Creation and Evidence generic patches discarded uncertain identity; creation transport and explicit patch owner now retain it. | REQ-01-058–070, 669 | Immutable requests plus one explicit patch dispatcher prevent replay becoming another transaction. | Each operation keeps its existing backend hashing/replay semantics. | Lost responses replay original bytes/key; receipt and complete history equality prove no additional effect. | RESOLVED: create and link loss scenarios pass. |
| P05 confirmed | Task presentation returned null after navigation despite retained acceptance; Party owner now consumes operation results. | REQ-03-035, 271, 283 | Authoritative receipts replace presentation booleans, removing false failure coupling. | Task lifecycle validation and drafts remain Task-owned. | Accepted source receipt remains accepted after retarget; Task recovery regression passes. | RESOLVED: independent receipt, late acceptance and Task browser recovery pass. |
| P06 confirmed | Generic refresh failure was local feedback; explicit owner and surface registry now retain Party refresh debt. | REQ-01-070; Core 03 concurrency | Read-only refresh recovery protects accepted mutations across detached surfaces. | Source reads scan authorized pages; ordinary edits retain existing behavior. | Refresh retry sends zero writes and preserves history; incomplete reads remain recoverable. | RESOLVED: refresh-failure browser scenario and newer-source-version test pass. |
| P07 hardening | genericWorkbookModel lacked cohesive bindings; partyLinkModel now has exact descriptors. | REQ-01-502, 615–617; REQ-02-227–228; REQ-03-274 | Explicit source/type/field/binding ownership preserves three pairs without label inference. | Existing clear payloads preserved; no attendee/audience pair invented. | All 12 pair/state combinations and exact atomic clear requests pass. | RESOLVED: controls and three-pair owner tests pass. |
| P08 review hypotheses, characterized | Late callbacks, duplicate activation, roles and ordering covered in Party owner and collaboration tests. | REQ-03-035, 271, 283, 292; REQ-04-022–024 | Synchronous reservation, invalidatable review and monotonic source observations isolate attempts from current selection. | Shared runtime suspension/retirement and existing source dispatcher remain owners. | Zero duplicate effects, wrong-row writes, source-version regressions or protected recovery exposure in covered races. No claim that every hypothesis was a baseline defect. | RESOLVED: focused race/security tests and real-service navigation pass; newly exposed same-pair/focus/stale-choice wiring defects corrected. |

## Validation record

Planning baseline, not completion evidence:

- All four current task guides passed for `web.workbook`, `module.parties`,
  `module.evidence`, and `module.tasksdecisions`.
- `make test-slice OWNER=web.workbook` with explicit Task recovery, generic
  pair-model and surfaces-suite rows: PASS, 4/4 graph units,
  `.cartulary/test-results/20260912T065517Z-p13018`.
- Party admission/hash row: PASS, 1/1,
  `.cartulary/test-results/20260912T065517Z-p13019`.
- Evidence direct Party-reference admission row: PASS, 1/1,
  `.cartulary/test-results/20260912T065517Z-p13024`.
- `make service-backed-test-slice OWNER=module.parties` with idempotency,
  active same-incident references and exact-match reuse rows: PASS, 5/5,
  `.cartulary/test-results/20260912T065613Z-p14365`.
- Execution-start branch/HEAD/status recheck: unchanged and clean.

## Workstream checkpoints

PL-01 started. Existing implementation confirms P01–P06; P07 preserves existing
correct clear semantics. P08 remains a review hypothesis pending executable
characterization. No prerequisite authorization is currently needed.

## Compatibility and rollback

No dependencies, public contracts, backend behavior, analyst data, digest,
standalone Parties CRUD, account identity, Party merge, Communications Log
collections, Assessments or Indicators are changed by this seam.

Recovery is memory-only within the supported incident/account runtime. Reverting
frontend code never reverses committed Party creation/reuse, source text or
references, server receipts, change sets or history. Do not issue compensating
deletes/clears or discard live retained work during owner replacement.

## Final disposition

PL-01 through PL-05 are complete. All eight gaps are resolved with focused and
real-service evidence. The final full frontend suite has two explicitly recorded
unrelated failures: the baseline Timeline assertion and the intermittent Network
Flow focus row that passes alone. No Party implementation or prerequisite is
blocked. RESULTS_DIR remains unset; retained full-warm-run maintenance was
skipped. No commit, push or deployment was performed.

### PL-01 exit / PL-02 start

Added three-pair clear-shape characterization to WorkbookShell.surfaces.test.tsx
and its authored workbook selector. Existing Task recovery and surface selection
characterization passed through `make test-slice OWNER=web.workbook` with
`ROWS=web.workbook.regression.workbookshell_surfaces_suite_668e482b1e,web.workbook.regression.explicit_task_patch_recovery`:
3/3 graph units, `.cartulary/test-results/20260912T070134Z-p33059`.
No prerequisite or owner contradiction was found. Remaining race hypotheses
will be tested with the retained owner in PL-03. Next: isolated contract-backed
pair presentation and candidate discovery, then runtime assembly in PL-03.

### PL-02 exit / PL-03 start

Added explicit pair descriptors, contract-backed PartyLinkControls, the paged
candidate hook, and createPartyLinkReader using existing query operations.
Focused row `web.workbook.regression.party_link_controls`: PASS 2/2 graph units,
`.cartulary/test-results/20260912T070729Z-p37768`. Frontend typecheck: PASS 2/2,
`.cartulary/test-results/20260912T070714Z-p37277`.
Earlier focused runs exposed an invalid selector helper and incorrect test label
assumptions; corrected to the existing field selector and actual discovered
labels. No product contract changed. Assembly intentionally follows in PL-03 so
there is no temporary second mutation path. Next: retained owner, explicit patch
policy extension and replacement of the old component-local workflow.

### PL-03 exit / PL-04 start

Replaced the component-local mutation hook with WorkbookPartyLinkOperationOwner,
creation transport and runtime subscriptions. Party source actions use the
existing explicit PATCH dispatcher with captured source versions; Task lifecycle
validation stays on its existing path. Creation receipts are retained before
linking, and refresh recovery performs reads only. Moved the obsolete hook's
partial-completion and clear characterization into retained-owner tests covering
all three pairs, navigation, duplicate activation, independent receipts, exact
replay, late source acceptance, authorization suspension and newer source versions.

Focused controls, retained Party owner, explicit Task owner and shell surfaces:
PASS 5/5 graph units at `.cartulary/test-results/20260912T072704Z-p45994`.
Typecheck: PASS 2/2 at `.cartulary/test-results/20260912T072801Z-p47457`.
An earlier shell run failed two expectations: the now-empty reference correctly
disables Clear link, and a closed disclosure duplicated feedback in the DOM.
The assertion now checks action-specific availability and disclosure content is
mounted only while open. Intermediate typecheck failures were unused imports
and fixture typing, corrected without contract changes. Next: detached socket
observations, current authority, keyboard/focus, real-service fault scenarios and
visual review. Remaining risks are browser integration and final regression checks.

### PL-04 integration checkpoint

All seven real-service scenarios in `parties.recovery.spec.ts` passed:
`make service-backed-test-slice OWNER=module.parties ROWS=module.parties.browser.source_link_recovery`,
11/11 graph units, `.cartulary/test-results/20260912T074659Z-p45813`.
Requests are forwarded to the service before responses are suppressed; the
conflict scenario performs a competing source PATCH. Receipt equality, exact
request bytes, source wording, unchanged reuse, optional email omission,
original-row recovery and candidate paging are asserted. Existing Party sentinel
passed in `.cartulary/test-results/20260912T073912Z-p90900`; that run as a whole
failed because the new scenarios used the compact navigation menu on desktop.
The corrected helper selects the existing System views menu when available.

Browser investigation also exposed a confirmed P08 defect in new wiring:
selecting the already-selected pair cleared its review without changing effect
dependencies. Corrected and added a hook regression. A keyboard test exposed
focus restoration before the Create trigger was re-enabled; restoration now
runs after rendering. Initial fixtures incorrectly assumed source creation
preserved leading/trailing whitespace and omitted Task kind; fixtures now compare
against authoritative source receipts and supply required Task inputs.

Latest focused controls, Party owner, socket ordering, Task recovery and shell
surfaces passed 6/6 at `.cartulary/test-results/20260912T074943Z-p82428`;
typecheck passed at `.cartulary/test-results/20260912T074928Z-p81938`.
Further candidate invalidation and unavailable-target coverage was added after
those runs and requires final validation. Source patches now recheck current
role, incident state, source and target through existing authorized reads.
Detached Evidence/Task versions are observed before socket transaction suppression.

Reviewed browser attachments from the seven-scenario passing run at desktop and
768px. Corrected native pair/filter styling and fieldset chrome to use input
tokens. The existing compact inspector remains a full-height overlay with the
grid retained/inert behind it, per design §8; desktop retains the adjacent grid.
Recovery disclosure fits the compact viewport and restores its summary on Escape.
Golden reconciliation and final screenshots remain pending.

The first `make agent-finalize` failed before mutation because new workbook
selectors made generated topology inputs stale (`json-shape-check`,
`.cartulary/test-results/20260912T075200Z-p88365`). Running Make generation before
retrying. RESULTS_DIR is unset; retained-run and performance maintenance are
intentionally skipped. No specification, public-contract or backend prerequisite
has appeared.

### PL-04 exit / PL-05 start

Integrated Party socket invalidation, current-authority reads, source/target
preflight, detached source version observations, late-receipt retention and
responsive shell recovery. Reviewed current desktop/768px authoring and recovery
attachments from `.cartulary/test-results/20260912T080223Z-p47325`; themed controls,
grid context and bounded recovery are intact. That service-backed run passed
13/13 graph units, covering all seven new scenarios plus the existing Party
sentinel. Additional final assertions now compare complete history before and
after exact replay and refresh; they will run in PL-05.

`make browser-e2e-a11y` passed 14/14 at
`.cartulary/test-results/20260912T075418Z-p1256`.
`make browser-e2e-visual` passed 12/12 at
`.cartulary/test-results/20260912T075418Z-p1154`. Reviewed its reconciliation:
221 captures, 221 active committed goldens, all 28 registered fixtures resolved,
zero missing/ambiguous/orphan mappings and no errors. No golden changed or needed
promotion; renderer, viewport, masks and registry remain unchanged.

Broad checks exposed a prohibited type-import cycle through the runtime. Moved
the shared authority type to the leaf `mutations/workbookMutationAuthority.ts`;
kept the explicit owner's public alias. Removed a production selector string in
favor of a React ref and updated the existing wire-authoring ownership assertion
after removing the obsolete generic creation path. Architecture checks remain
strict. Typecheck, import boundaries and Biome now pass.

The full frontend suite passes 585/586 graph units at
`.cartulary/test-results/20260912T080606Z-p25931`. Its sole failure is the unchanged
Timeline inspector assertion at `WorkbookShell.inspector.test.tsx:607`, expecting
`timeline.mark_reviewed` to be absent. The exact baseline HEAD reproduces this
failure in a detached worktree at
`/tmp/cartulary-party-baseline-f3f195d2/.cartulary/test-results/20260912T080107Z-p35927`,
using `make test-slice OWNER=module.workbook`
`ROWS=module.workbook.frontend_unit.verify_inspector_selection_tab_state_details_rel_d2dc82a4bb`.
An earlier concurrent run also timed out in Network Flow exploration; its narrow
rerun passed 2/2 at `.cartulary/test-results/20260912T075813Z-p27929`, and the latest
full suite passes that row. Neither unrelated behavior was changed.

Generated topology was repaired with `make generate`. `make agent-finalize`
passed at `.cartulary/test-results/20260912T080149Z-p43761` before broad terminal
checks, with RESULTS_DIR unset. Retained-run/performance maintenance was skipped.
No owner contradiction or backend prerequisite remains. Next: final service and
browser assertions, second ordinary visual run, static/drift checks and final
scope review. The known baseline Timeline assertion is the remaining verification
limitation, not a prerequisite to this Party seam.

### PL-05 validation checkpoint

Final review found that a socket-invalidated Party selection could become
actionable again after the replacement page loaded. The selected ID now belongs
to its candidate revision; a new revision requires selection from refreshed
results. Earlier authorized rows still remain visible on failed reads. Added
the assertion to the existing candidate-invalidation test. Focused controls,
recovery and socket rows passed 4/4 at
`.cartulary/test-results/20260912T081153Z-p59002`.

Ran `make format`, `make generate`, then `make agent-finalize` before the final
broad verification. Generation passed at
`.cartulary/test-results/20260912T081159Z-p59707`; finalizer passed at
`.cartulary/test-results/20260912T081209Z-p62762`. RESULTS_DIR stayed unset:
retained successful full-warm-run and performance maintenance were skipped,
without treating narrow or baseline runs as eligible warm evidence.

All final run paths below are relative to `.cartulary/test-results/`.

| Command / selected scope | Result | Run |
| --- | --- | --- |
| `make frontend-typecheck` | PASS 2/2 | `20260912T081314Z-p66751` |
| `make frontend-import-boundary-check` | PASS 2/2 | `20260912T081335Z-p99908` |
| `make lint-biome` | PASS 2/2 | `20260912T081343Z-p2509` |
| `make frontend-unit` | FAIL 584/586: baseline Timeline assertion and intermittent Network Flow focus row; all Party rows pass | `20260912T081347Z-p2987` |
| `make test-slice OWNER=web.networkflow ROWS=web.networkflow.regression.exploration_focus` | PASS 2/2 isolated rerun | `20260912T081815Z-p97030` |
| `make service-backed-test-slice OWNER=module.parties` replay storage, active references, exact-match reuse rows | PASS 5/5 | `20260912T080534Z-p87105` |
| `make test-slice OWNER=module.evidence` direct Party references, stored idempotency payloads, mutation admission/hash rows | PASS 2/2 | `20260912T080534Z-p87132` |
| `make service-backed-test-slice OWNER=module.tasksdecisions` workbook idempotency adapter and direct-reference rows | PASS 3/3 | `20260912T080534Z-p87114` |
| `make service-backed-test-slice OWNER=module.parties ROWS=module.parties.browser.source_link_recovery,module.parties.browser.the_browser_workbook_surface_keeps_ordinary_evid_687c58e86e` | PASS 13/13; seven new scenarios plus existing sentinel | `20260912T081314Z-p66657` |
| `make test-catalog-check` | PASS | `20260912T081314Z-p67012` |
| `make json-shape-check` | PASS 3/3 | `20260912T081316Z-p67315` |
| `make generated-artifact-policy-check` | PASS 3/3 | `20260912T081322Z-p79151` |
| `make generate-drift` | PASS 4/4 | `20260912T081326Z-p88711` |

The browser row forwards real mutations, captures authoritative receipts and
suppresses only selected responses. It compares exact replay request bytes,
transaction IDs, receipts, record versions, change sets, full history, source
wording, unaffected rows and unchanged reused Party rows. Creation and recovery
controls exercise keyboard submission and Escape focus restoration. Desktop
Evidence and compact Task authoring/recovery screenshots are retained as
`reviewed-party-authoring` and `retained-party-recovery` report attachments.
Reviewed expanded form captures from `20260912T081007Z-p84657`: required kind,
unchecked email proposal, optional details and save/cancel controls are visible
within the existing inspector scrollport. No clipping or overflow defect was
found. The final browser run repeats these captures after selection invalidation.

Two ordinary visual runs passed before the final candidate-state hardening:
`20260912T075418Z-p1154` and `20260912T081007Z-p84865`, each 12/12. Both reconcile
221 active captures/goldens and all 28 registered fixtures without errors.
A final ordinary run after the candidate-state change passed 12/12 at
`20260912T081429Z-p17585`; its reconciliation again has no errors. No golden promotion,
renderer change or visual registry modification was necessary. Accessibility
remains PASS 14/14 at `20260912T075418Z-p1256`; the final state change alters
selection validity and does not alter markup, layout or accessible names.

## Changed paths and ownership

43 files comprise this seam, including this handoff. The worktree began clean.

| Paths | Change and ownership |
| --- | --- |
| `apps/web/src/workbook/features/parties/partyLinkModel.ts`, `partyLinkStyles.ts`, `PartyLinkControls.tsx`, `PartyLinkPanel.tsx`, `PartyLinkRecovery.tsx`, `usePartyCandidates.ts` in that directory | Exact descriptors, contract-backed review, authorized paged picker and accessible retained result presentation. |
| `apps/web/src/workbook/features/parties/WorkbookPartyLinkOperationOwner.ts`, `useGenericPartyLinkWorkflow.ts` | Runtime operation ownership and presentation-only subscription hook. |
| `apps/web/src/workbook/adapters/createPartyCreationTransport.ts`, `createPartyLinkReader.ts` | Existing generated HTTP operations, immutable create request and complete receipt, bounded source discovery. |
| `apps/web/src/workbook/runtime/WorkbookExplicitPatchOwner.ts`, `WorkbookMutationRuntime.ts`, `WorkbookSurfaceRegistry.ts` | Reused source dispatcher, bounded Party preparation policy, retained operation accounting, late acceptance and refresh debt. |
| `apps/web/src/workbook/mutations/workbookMutationAuthority.ts`, `workbookMutationCommandPorts.ts`; `apps/web/src/workbook/models/genericWorkbookModel.ts` | Leaf authority type, removal of obsolete generic Party creation, explicit descriptor routing. |
| `apps/web/src/workbook/WorkbookShell.tsx`; `hooks/useWorkbookShellInfrastructure.ts`, `useWorkbookSurfaceQueries.ts`, `useGenericSurfaceMutationController.ts`; `components/GenericWorkbookSurface.tsx` | Narrow incident/account assembly, authority, originating sheet context and shell recovery. |
| `apps/web/src/workbook/features/generic/createGenericMutationCommandPort.ts`, `useGenericWorkbookInspectorComposition.tsx`, `GenericWorkbookInspectorPresentation.tsx` | Removed local Party dispatch and boolean-result coupling; inspector delegates to Party owner. |
| `apps/web/src/workbook/query/useGenericSurfaceQuery.ts`; `collaboration/WorkbookCollaborationCoordinator.ts` | Evidence and Task source high-water reconciliation; detached/own-transaction socket observations and Party candidate invalidation. |
| `apps/web/src/workbook/components/WorkbookSameFieldConflictResolver.tsx`; `features/coordination/TaskPatchRecovery.tsx` | Complete-action conflict wording and Party recovery separated from existing Task presentation. |
| `apps/web/src/workbook/features/parties/partyLinkControls.test.tsx`, `WorkbookPartyLinkOperationOwner.test.ts`; `collaboration/WorkbookCollaborationCoordinator.test.ts`; `query/useGenericSurfaceQuery.test.tsx`; `WorkbookShell.surfaces.test.tsx`; `apps/web/src/testing/sourceOwnershipPolicy.test.ts` | Focused tests and regression adaptation; moved obsolete local-hook characterization to authoritative retained-owner tests; strict ownership policy preserved. |
| `apps/web/e2e/parties.recovery.spec.ts`, `apps/web/e2e/sentinel.spec.ts` | New real-service fault/reuse/paging evidence and explicit review in the existing Party sentinel. |
| `tools/frontend_source_ownership.json`, `tools/test_families/web.workbook.json`, `tools/test_families/module.parties.json` | Authored frontend ownership and verification routing with Parties/Evidence/Task collaborators. |
| `tools/browser_e2e_batch_manifest.json`, `tools/execution_topology_render_index.json` | Required derivatives generated through Make; no hand editing. |
| `.markdownlint-cli2.jsonc` | Include this handoff in the existing explicit Markdown lint inventory. |
| `docs/handoffs/workbook-party-linking-refactor-handoff.md` | Human authority audit, dependent tracker, gap disposition and verification record; no runtime dependency on Markdown. |

No generated protocol/UI source, contract, schema, SQL, backend implementation,
dependency file, digest or golden is modified. Scope review preserves ordinary
Evidence/Task and generic editing, Timeline mention recovery, Communications Log,
and standalone Parties behavior as regression boundaries. The baseline Timeline
assertion and intermittent Network Flow focus row are the known full-suite
limitations; there is no unresolved Party implementation or prerequisite blocker.
Release-wide backend, security,
performance and full warm checks were not selected because this is a frontend
seam with narrow source-owner and browser evidence.

The latest concurrent frontend run repeated the unchanged Network Flow focus
failure after approximately its 15-second limit; the isolated rerun passed. This
row also passed in the earlier 585/586 full run. Load sensitivity is an inference
from those results, not proof of an infrastructure defect. Its behavior and test
remain untouched. The reproducible Timeline failure is separately confirmed at
the exact baseline HEAD, as recorded under PL-04. Neither failure is hidden by
catalog changes, weakened assertions or golden updates.

The existing Markdown target passed at `20260912T081340Z-p1408`, but final review
found its explicit globs did not include new handoffs automatically. Added this
handoff to `.markdownlint-cli2.jsonc`; the post-completion rerun now lints this
document as well. This documentation-only routing change does not
affect product checks, generated artifacts or conformance.

### PL-05 exit

All scoped functional, source-owner, browser, accessibility, visual, type,
architecture, catalog and generated-drift checks have passed. Full frontend
failures are dispositioned above without changing unrelated behavior. Final
visual review includes separate saved-Party and saved-source recovery at desktop
and compact widths; both retain the originating pair and keyboard Escape focus.
There are no golden changes to promote. The final implementation and ownership
inventory contains 43 files; branch and HEAD remain the recorded baseline.

Final scope review confirms no digest, adopted owner, public contract, backend,
database, dependency, committed golden or out-of-scope workflow changes.
The retained baseline checkout under `/tmp/cartulary-party-baseline-f3f195d2`
exists only for the recorded Timeline failure evidence. It contains no commit
or analyst-data modification. Rollback is frontend-only and must preserve
committed Parties, source values, both receipts, change sets and history.

Post-completion `make lint-markdown` passed at
`.cartulary/test-results/20260912T082100Z-p98630`, including this handoff.
`git diff --check` and the final 43-file scope review passed. These checks are
repeated after recording this result so the completed handoff itself is covered.
Next action: user review of this seam. No additional implementation workstream
is proposed.
