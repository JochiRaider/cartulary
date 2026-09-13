# Workbook Task lifecycle refactor

## Authorization, baseline, and owners

Implement only Task lifecycle editing and necessary ordinary-patch, inspector,
runtime, conflict, and reconciliation integration. The user separately authorized
the Task-only correction for guard-valid in-state writes rejected as
`no_effective_change`. Compound conflicts retain their complete draft, use
revisionless Keep saved, and require a fresh explicit atomic submission.
No confirmation is added to `task.status.transition`.

Baseline and implementation recheck: clean `main`, HEAD
`740d7e6e1946d2df0b6bdfd42987c8eb8e3c63f1`; index empty. Only root `AGENTS.md`
applies. Preserve user changes; no commit, reset, push, deploy, or data cleanup.
The digest localized read order and relevant Core/design/domain owners were read
during planning and mappings revalidated against current source. Research and
historical handoffs provide context, not additional authorization.

Governing owners: Core 01 REQ-01-058–069, 086–088, 228–234, 336–338,
615–617; Core 02 §10.4.1/§10.4.1.1, especially REQ-02-100 and 102–109;
Core 03 REQ-03-035, 037–081, 282/283, 291/292, 301, §4.4 and §16.4;
Core 04 session lifecycle, current authorization, CSRF, and concealment;
design §§5.3, 7.3, 10, 12, and 14. Digest local-feedback/accessibility advice
is ADOPT; density advice is ADAPT; destructive confirmation is inapplicable.

Permitted authored areas: Task coordination features; necessary workbook patch,
runtime, inspector, query, continuity and composition integration; Task-only
in-state backend correction; focused service/browser tests and fixtures; selectors;
verification ownership/routing; reviewed affected visuals; this handoff.
Make alone produces generated derivatives. No dependency, storage, migration,
public-wire, Task creation, assignment administration, Decision approval, or
Timeline lifecycle redesign. Preserve digest and completed handoffs.

Actual doctor facts: Go 1.27.1, Node 24.15.0, pnpm 10.33.0, ShellCheck 0.11.0;
React 19.2.5, TypeScript 6.0.2, Vite 8.0.8, Vitest 4.1.4, Playwright 1.59.1.
Doctor reports only incomplete inspection of inotify usage. Current primary
verification owners are `web.workbook` and `module.tasksdecisions`; collaborators
include `module.workbook`, `module.revisions`, `package.ui`, and affected browser
and design rows. Latest artifact-publication/Decision handoffs record coordinated
full checks PASS 834/834 units and 1222/1222 rows, without waived failures.

## Gap ledger

| Gap | Owner | Areas and remediation | Rationale and long-term benefit | Compatibility | Risk | Binary exit |
| --- | --- | --- | --- | --- | --- | --- |
| G1 Independent picker and default blocked | Core 01 inspector; Core 03 291/292 | Selected Task, canonical Workflow admission | One subject and one action remove accidental cross-row writes | Existing feature and route | Retarget leaks | One editor, no target picker, accepted-state initialization |
| G2 Missing machine guidance and atomic owner path | Core 02 100, 102–109 | Task intent model, owner/reason/completion inputs | Valid compound intents avoid invalid intermediate states | Declared writable fields | Duplicated rules or invented defaults | Complete transition/guard matrix passes |
| G3 Separate command and conflict bypass | Core 01 058–069; Core 03 037–081 | Shared patch primitives and conflict registration | One transport interpretation, retained compound work | Existing PATCH and resolver | Partial conflict application | Compound draft survives Keep saved and fresh submission |
| G4 Duplicate admission and lost identity | Core 03 282/301 | Runtime explicit patch owner | Synchronous admission and immutable exact replay survive panels | Outside autosave FIFO; memory only | Unknown commit treated as rollback | Double submit once; loss/malformed success replay exactly |
| G5 Requested-status and refresh-owned completion | Core 01 088; Core 03 035/283 | Receipt-first row/version acceptance and separate refresh | Acknowledgement remains trustworthy through failed reads | Existing row/change-set response | Stale response regression | Authoritative values retained; refresh retry sends no write |
| G6 Filter exit loses context | Core 03 283/292 | Current-query reconciliation and continuity | Explain saved work without resurrecting rows | Filters and unrelated drafts unchanged | Late focus steals selection | Filter exit notice and deterministic visible fallback |
| G7 Task in-state write rejected | Core 02 102/109 | Narrow Task lifecycle patch admission correction | Adopted in-state legality reaches ordinary history/replay | User-authorized backend exception; no schema change | Decision or unrelated no-op drift | Valid in-state commit once; exact replay has no new effects |

## Tracker

| Row | Status | Exit |
| --- | --- | --- |
| TL-01 | DONE | Baseline, ledger, routed failing characterization |
| TL-02 | DONE | Selected editor, intent model, in-state correction |
| TL-03 | DONE | Shared patch, sequencing, conflicts, retained recovery |
| TL-04 | DONE | Reconciliation, accessibility, service/browser evidence |
| TL-05 | DONE | Final checks, scope audit, complete handoff |

## TL-01 log

Planning evidence roots below are under `.cartulary/test-results/`:

- `make doctor`: PASS `20260911T133610Z-p96796`.
- `make test-slice OWNER=module.tasksdecisions ROWS=module.tasksdecisions.frontend_unit.coordination_workflow_lifecycle_7a8d13c2f0,module.tasksdecisions.source_mutations.task_and_decision_mutation_admission_and_replay_6d4ea5900a`:
  PASS 3/3 `20260911T133831Z-p98112`.
- `make service-backed-test-slice OWNER=module.tasksdecisions ROWS=module.tasksdecisions.source_workbook.task_requests_and_decisions_persist_as_workbook_887d6b579e`:
  PASS 3/3 `20260911T133831Z-p98121`.
- `make test-slice OWNER=web.workbook ROWS=web.workbook.regression.use_generic_surface_query_owner_fab9b06dde,web.workbook.regression.workbookz_mutation_runtime_semantic_command_ports_7b2f1d02a4,web.workbook.regression.workbookz_mutation_runtime_responsibilities_a106000001`:
  PASS 4/4 `20260911T134348Z-p17700`.

Implementation baseline branch/HEAD/status audits and both requested task guides
passed again. Inspected the requested coordination/inspector/mutation/runtime
paths, Task contract, policy/admission/source, conflict resolver, query/continuity,
current tests and authored verification inputs. No owner contradiction found.
Next: retain failing characterization, then begin TL-02.

## Compatibility and rollback

Existing endpoints, wire shapes, stored schemas, dependencies, and analyst data
remain unchanged. Roll back implementation changes without deleting Tasks,
history, revisions, receipts, or analyst data. No follow-on refactor is authorized.

### TL-01 exit

`make generate` PASS `20260911T134832Z-p19691` after authored routing updates.
The routed frontend characterization FAIL 1/2 units at
`20260911T134855Z-p22770`: saved open Task incorrectly initializes blocked.
The routed service matrix FAIL 2/3 units at `20260911T134855Z-p22781`:
all five in-state pairs reject; the remaining 20 transition/replay cases pass.
These are expected product failures, retained before correction, not waived
checks. Paths: coordination controller test, Task store test, authored
Tasks/Decisions family, generated topology, and this handoff. Existing behavior
is unchanged. TL-01 exits are met; next TL-02 corrects the selected editor and
Task in-state acceptance.

## TL-02 exit

Selected-subject binding uses canonical `task.status.transition` admission,
Workflow placement, declared role and disabled conditions. The editor initializes
from accepted fields; workbook-memory drafts are keyed by Task identity and
changed guard siblings require local review. One Task intent model supplies the
transition matrix, guard guidance, compound writable fields, timestamp syntax,
and grid/ordinary-inspector validation. No confirmation or client timestamps.
Task-only in-state acceptance retains the existing version/history/replay path.

Paths: Task coordination model/controller/bindings/tests; generic inspector
composition/presentation and mutation controller; generic surface; runtime draft
lifetime; Task patch admission; authored family/generated topology.
`make format` PASS `20260911T135735Z-p43323`; `make generate` PASS
`20260911T135804Z-p47709`. Routed frontend slice PASS 2/2
`20260911T135855Z-p69378`; routed Task service slice PASS 3/3
`20260911T135823Z-p50813`, including all 25 pairs and exact replay.
Initial typecheck and frontend rerun failed only because new assertions used
an unavailable Chai matcher; replaced with the equivalent DOM `:disabled`
assertion. No assertion was weakened. Typecheck subsequently PASS 2/2
`20260911T135916Z-p70172`.
Compatibility unchanged; no new owner contradiction. Exit criteria met.
Retained transport, refresh acknowledgement, and late presentation handling are
TL-03/TL-04 work. Next: shared ordinary-patch primitives and runtime admission.

## TL-03 exit

Shared ordinary-patch capture/transport validates complete correlated rows and
nonempty receipts. Removed `updateTaskLifecycle` and its requested-status result.
Runtime explicit operations reserve synchronously outside the FIFO, wait for
earlier writes/required reads, compare affected guard fields, retain immutable
requests, and distinguish rejection, uncertainty, receipt, and refresh debt.
Replay keeps the exact request and secure ID; refresh-only retries send no patch.
Session loss conceals retained work; account/incident/runtime retirement clears
protected memory. Drafts remain keyed by Task and ordinary inspector field.
Compound conflicts retain full intent and block partial application; Keep saved
clears the field conflict, then fresh explicit submission reviews current fields.
Expired compound tokens require a successful current read before local conflict
retirement; this path performs no additional source write.

Paths include shared patch adapter, generic and queued adapters, command ports,
explicit runtime owner, conflict store/resolver, surface refresh coordination,
shell/query bindings, inspector drafts, recovery presentation, and focused tests.
`make generate` PASS `20260911T141209Z-p84449` after authored owner routing.
Routed explicit/runtime/command/query slice PASS 5/5
`20260911T141234Z-p91734`; typecheck PASS 2/2 at
`20260911T141234Z-p91809` and `20260911T141517Z-p98595`.
The surface regression initially failed because a supposed successful Task
fixture returned a different record ID. Repaired that fixture to match its
request, preserving its syncing/refresh assertions. Removed duplicate rejected
error text from the recovery notice. Surface/explicit/runtime slice then PASS
4/4 `20260911T141720Z-p4703`. `make format` PASS
`20260911T141647Z-p99765`; `make lint-biome` PASS
`20260911T141720Z-p4818`. Earlier formatting/lint failures were new dependency
and formatting diagnostics, repaired. An attempted unsupported lint flag was
rejected by public-input validation; normal public Make verification was used.

Compatibility: same PATCH and field-conflict routes, no public shapes or stored
data changed; autosave FIFO remains exclusive to autosave-origin edits.
TL-03 exits met. Remaining risks are browser timing, accessible filtered-row
fallback, and stale query/live-update reconciliation. Next TL-04 validates those
against real services and reviews the affected visual projection.

## TL-04 exit

Task receipts now feed a committed row/version high-water mark shared with
query/live updates, autosave, explicit patches, field conflicts, and history.
Receipts update only existing visible records. Acknowledgement precedes refresh;
failed refresh retains the receipt, and its recovery sends reads only. Filter
exit clears the selected subject and uses grid-root focus without changing
filters. Actual subsequent focus uses adapter-owned semantic row/cell identity
so a retained grid caret can select the same Task when it returns. This preserves
adapter notification deduplication and does not remount or reset the grid.

Retargeted presentation fences late feedback. Task drafts remain in runtime
memory; selected fields, guard siblings, and compound intent are retained.
Chronological completion errors remain server-owned and appear at the completion
control. Client validation uses declared normalization, including nonclearable
owner fields. Coordination timeout rejects before dispatch; transport timeout
retains an uncertain immutable request. Refresh debt on remount shares the same
runtime refresh boundary as prior writes.

Paths: generic surface/query/inspector integration, coordination controls and
recovery region, explicit owner/transport tests, Task service matrix fixture,
`apps/web/e2e/sentinel.spec.ts`, authored Task/workbook routing, selectors, and
Make-generated topology/browser manifest. No production backend change beyond
the authorized in-state acceptance. The ownerless terminal fixture explicitly
seeds a valid terminal state because Task creation defaults its owner; creation
behavior is unchanged. No new public field was introduced for `created_at`.

Exact routed commands and passing evidence (all roots under
`.cartulary/test-results/`):

- `make generate`: PASS `20260911T142356Z-p14868`,
  `20260911T142456Z-p22207`, `20260911T143410Z-p13854` after authored inputs.
- `make service-backed-test-slice OWNER=module.tasksdecisions ROWS=module.tasksdecisions.source_workbook.task_requests_and_decisions_persist_as_workbook_887d6b579e`:
  PASS 3/3 `20260911T142824Z-p64037`, including true ownerless atomic reopening,
  all 25 status pairs, rejected-write atomicity, clearing/defaulting, and exact
  replay with one receipt/change set and no extra version advancement.
- `make test-slice OWNER=module.tasksdecisions ROWS=module.tasksdecisions.frontend_unit.coordination_workflow_lifecycle_7a8d13c2f0`:
  PASS 2/2 `20260911T144354Z-p91562` (also `20260911T143146Z-p39138`).
- `make test-slice OWNER=web.workbook ROWS=web.workbook.regression.use_generic_surface_query_owner_fab9b06dde,web.workbook.regression.workbookshell_surfaces_suite_668e482b1e`:
  PASS 3/3 `20260911T142823Z-p63801`.
- `make test-slice OWNER=web.workbook ROWS=web.workbook.regression.explicit_task_patch_recovery,web.workbook.regression.workbookz_mutation_runtime_responsibilities_a106000001`:
  PASS 3/3 `20260911T144509Z-p29442`.
- `make test-slice OWNER=web.workbook ROWS=web.workbook.regression.explicit_task_patch_recovery,web.workbook.regression.workbookshell_surfaces_suite_668e482b1e`:
  PASS 3/3 `20260911T144921Z-p72756`. Includes actual autosave receipt/refresh
  sequencing, exact immutable replay, malformed success, session replacement,
  concealed late receipts, compound conflicts, and pre-dispatch timeout.
- `make service-backed-test-slice OWNER=module.tasksdecisions ROWS=module.tasksdecisions.browser.task_lifecycle_patch_recovery,module.tasksdecisions.browser.the_browser_workbook_opens_task_requests_and_dec_8048a75ceb`:
  PASS 11/11 units `20260911T144921Z-p72755`; all six browser scenarios pass.
  The five new scenarios exercise blocked/completed/reopened Tasks, no
  confirmation, atomic owner/reason submission, duplicate Task labels, filter
  exit/return, precommit loss, committed-but-lost response, malformed success,
  refresh-only retry, detached inspector, and compound conflict recovery.
  The existing native Task/Decision browser workflow also passes.
- `make service-backed-test-slice OWNER=module.workbook ROWS=module.workbook.accessibility.verify_coordination_surfaces_and_full_keyboard_c_83c0845d3e,module.workbook.accessibility.verify_inspector_tabs_relationship_links_evidenc_9ee9fd9ea2`:
  PASS 11/11 `20260911T143213Z-p78379`.
- `make service-backed-test-slice OWNER=module.workbook ROWS=module.workbook.visual.capture_task_requests_or_decisions_parties_link_558c8596cc`:
  PASS 11/11 `20260911T144023Z-p22873` against existing goldens.
- `make frontend-typecheck`: PASS 2/2 `20260911T144921Z-p72940`.
- `make frontend-import-boundary-check`: PASS 2/2
  `20260911T144921Z-p72935`.
- `make lint-biome`: PASS 2/2 `20260911T144652Z-p35622`.
- `make format`: PASS 2/2 `20260911T144854Z-p68223`.

The successful browser report retains exact request/receipt/history JSON,
including unchanged history after revisionless Keep saved and byte-identical
replay requests. Review root:
`20260911T144921Z-p72755/browser-e2e-webserver-backed/browser-groups/functional-support-default-sentinel/`.
Its `playwright-report.json` retains the editor screenshot and accessibility
snapshot; `task-lifecycle-editor-review.png` is the same attached image decoded
for human inspection. Reviewed compact controls, local guidance, distinct saved
status, keyboard order, visible focus, and preserved grid framing. The browser
asserts native terminal-option disabling, tab order, focus outline, and Enter
submission. These are implementation-support evidence, not new conformance
claims. The ordinary visual reconciliation passed with no golden changes,
renderer/viewport/mask/tolerance changes, or visual-update command.

Failures were investigated and repaired, never waived:

- Generation `20260911T142326Z-p11741` rejected unsorted scenario IDs; fixed
  the authored catalog before regeneration.
- Initial browser `20260911T142550Z-p25731` and surface/query
  `20260911T142550Z-p25732` exposed an omitted focus-anchor prop and a test
  comparing normalized cloned rows with raw fixture identity. Restored the prop;
  assert unrelated-row identity against the accepted pre-mutation row.
- Typecheck `20260911T142457Z-p22792` caught that prop and an invalid browser
  creation fixture with a null owner. Removed that out-of-contract input;
  ownerless terminal behavior is covered through the service fixture above.
- Browser `20260911T142835Z-p74299` reached the new editor but the disabled
  option matcher did not honor native option state. The assertion now checks
  its actual `disabled` property; prohibited transitions remain asserted.
- Browser `20260911T143136Z-p19325`, `20260911T144023Z-p22858`,
  `20260911T144354Z-p91557`, and `20260911T144652Z-p35474` characterized
  the returned-Task focus/selection defect. Repaired semantic focus integration;
  the same focused-record assertion now passes.
- Visual `20260911T143145Z-p35870` exposed a missing first grid-layout slot when
  the recovery component returned null. Preserved the empty slot; existing
  goldens then passed. No golden was refreshed to hide that layout defect.
- New sequencing test `20260911T144022Z-p22839` and
  `20260911T144354Z-p91533` used the wrong queue phase spelling. Corrected it
  to the existing typed `in_flight` state; receipt/version/refresh assertions
  remain intact. Typecheck `20260911T144354Z-p91666` caught unsupported
  `findLast`; used the repository-supported array operations.

Compatibility and exit: all TL-04 criteria met; existing interfaces/storage
remain unchanged. Remaining intentional limits are workbook-memory retention
and server-only chronological validation. No unresolved owner contradiction.
Next: TL-05 finalization, coordinated full verification, and final scope audit.

## TL-05 verification log

`make agent-finalize` PASS 1/1 `20260911T145143Z-p6624`, with `RESULTS_DIR`
unset. Retained-run selection/maintenance was skipped; no failed or stale full
run was supplied as successful warm evidence.

Pre-full checks passed: `make generate-drift` 4/4
`20260911T145222Z-p10253`; `make generated-artifact-policy-check` 3/3
`20260911T145222Z-p10257`; `make toolchain-drift` 2/2
`20260911T145222Z-p10308`; `make lint-biome` 2/2
`20260911T145222Z-p10612`; `make lint-markdown`
`20260911T145222Z-p10649`. `git diff --check` and branch/HEAD/index audits
passed. Digest, completed handoffs, public contracts, stored schema, and
lockfiles were unchanged.

First coordinated `make check` FAIL 832/835 units at
`20260911T145319Z-p16279`. Exactly three assertions failed, all related to this
change. No failure is waived:

- `web.architecture.boundary_support.source_ownership_policy_suite_80cf87ef19`:
  six new TypeScript files were missing from authored
  `tools/frontend_source_ownership.json`. Registered the Task fixture under
  `web.testing` and the five runtime/feature/adapter files under `web.workbook`.
- `web.architecture.boundary_support.workbook_layout_policy_suite_5f8fee9a2f`:
  moved recovery-region geometry into the existing shared
  `WorkbookSurfaceLayout.tsx` owner without changing its values or rendering.
- `package.ui.frontend_unit.workbook_interaction_contracts_dcad48388a`:
  the expected selector list still contained retired Task picker/submit IDs.
  Updated the projection test to assert the canonical feature-action selector
  and to reject both retired tokens explicitly. Other action expectations and
  all runtime assertions remain intact.

`make task-guide ROLE=module-author OWNER=web.architecture` confirms the routed
source/layout slice and import-boundary check. These corrections add the shared
layout owner and authored source-ownership manifest to the reviewed path set;
no new product scope or owner contradiction. Regenerate, verify those slices,
finalize again without `RESULTS_DIR`, and repeat the coordinated full check.

Corrections verified: `make format` PASS `20260911T150026Z-p78683`;
`make generate` PASS `20260911T150056Z-p83031`;
`make test-slice OWNER=web.architecture ROWS=web.architecture.boundary_support.source_ownership_policy_suite_80cf87ef19,web.architecture.boundary_support.workbook_layout_policy_suite_5f8fee9a2f`
PASS 3/3 `20260911T150117Z-p86318`;
`make test-slice OWNER=package.ui ROWS=package.ui.frontend_unit.workbook_interaction_contracts_dcad48388a`
PASS 2/2 `20260911T150117Z-p86325`;
`make frontend-typecheck` PASS 2/2 `20260911T150117Z-p86488`;
`make frontend-import-boundary-check` PASS 2/2 `20260911T150117Z-p86515`.
`make agent-finalize` passed again, 1/1 `20260911T150144Z-p88287`, with
`RESULTS_DIR` still unset and retained-run maintenance skipped.
`make generate-drift` PASS 4/4 `20260911T150207Z-p91700`;
`make lint-biome` PASS 2/2 `20260911T150207Z-p91855`.

Second coordinated `make check` FAIL 834/835 units at
`20260911T150237Z-p95733`. The only failure was
`web.networkflow.regression.exploration_focus`, after 15,334 ms with a Vitest
registration-stack error, consistent with the existing 15,000 ms test timeout.
That row passed in the preceding full run. Network Flow source and timeout
configuration are unchanged. An unchanged isolated rerun,
`make test-slice OWNER=web.networkflow ROWS=web.networkflow.regression.exploration_focus`,
PASS 2/2 `20260911T151048Z-p43205` (7,769 ms including target setup).
No assertions, routing, workload, or timeout limits were relaxed. Repeat the
coordinated full check against the same production source; the timeout is
retained evidence, not a waived result.

## TL-05 exit and final scope

Final coordinated `make check` PASS **835/835 units and 1223/1223 rows**, with
zero failed, skipped, or cancelled units, at
`.cartulary/test-results/20260911T151131Z-p44048` (268,829 ms).
Read `run-summary.json`, `target-summaries/check.json`, `rows/`, and
`unit-results/` within that root for exact accounting. This is a normal full
warm run: 666 cache hits and 169 required bypass executions under the current
harness policy. Its production-source digest is
`sha256:36651812a50033dafe86d2a3a03b65893334ecb4c44a947fbc3b9856b7c56e86`,
identical to the preceding full run. The Network Flow row passes without any
source or verification-policy changes. No waived failures remain.

The full check covers the coordinated backend/frontend unit and service,
contract, boundary, generation, lint, security, and harness work. Real browser,
accessibility, and visual evidence are the separate routed runs recorded in
TL-04; their scope is not inferred from `make check`. Decision supersession,
merge, history, inspector, save-status, autosave, and artifact-publication
regression inputs are preserved. New browser evidence proves Task exact replay
and refresh-only recovery; it does not authorize another workflow refactor.

`make lint-markdown` PASS `20260911T151633Z-p33314` before this final tracker
edit. `git diff --check`, `git branch --show-current`, `git rev-parse HEAD`,
`git diff --name-only --cached`, and `git status --short` passed the scope audit:
`main`, unchanged HEAD `740d7e6e1946d2df0b6bdfd42987c8eb8e3c63f1`, empty index,
and exactly 41 intended changed/new paths below. The required Markdown,
whitespace, and scope/status checks are repeated after this final edit.

Owner decisions: Core 01 ordinary patches and canonical inspector admission,
Core 02 Task guards/in-state legality, Core 03 conflict/recovery/continuity,
and Core 04 current authorization remain the adopted owners. Shared layout
and authored source ownership were repaired against their existing owners.
Only Task in-state acceptance changes production backend behavior. No public
endpoint, wire shape, stored schema, migration, dependency, lockfile, digest,
completed handoff, or committed golden changed. No commit, reset, push, deploy,
or analyst-data cleanup was performed.

Intentional limits: retained drafts and requests exist only for the permitted
workbook-memory lifetime; reload/retirement ends that lifetime. Completion
syntax is checked locally, while chronology against creation remains validated
by the server because the Task projection does not expose creation time.
The isolated Network Flow timeout is documented above and passed unchanged
in both isolated and coordinated verification; no unrelated workflow was
modified. No unresolved Task seam limitation or owner contradiction remains.

Rollback restores these implementation, test, routing, and generated-projection
changes. It must not delete Tasks, analyst data, history, revisions, receipts,
or change sets. All TL exits are met. No successor workstream is authorized.

### Exact changed paths

- `apps/web/e2e/sentinel.spec.ts`
- `apps/web/src/testing/taskWorkbookTestSupport.ts`
- `apps/web/src/workbook/WorkbookShell.surfaces.test.tsx`
- `apps/web/src/workbook/WorkbookShell.tsx`
- `apps/web/src/workbook/adapters/createWorkbookPendingMutationAdapter.ts`
- `apps/web/src/workbook/adapters/workbookRecordPatchTransport.ts`
- `apps/web/src/workbook/components/GenericWorkbookSurface.tsx`
- `apps/web/src/workbook/components/WorkbookSameFieldConflictResolver.tsx`
- `apps/web/src/workbook/features/coordination/CoordinationWorkflowBindings.tsx`
- `apps/web/src/workbook/features/coordination/TaskPatchRecovery.tsx`
- `apps/web/src/workbook/features/coordination/taskLifecycleModel.ts`
- `apps/web/src/workbook/features/coordination/useCoordinationWorkflowController.test.tsx`
- `apps/web/src/workbook/features/coordination/useCoordinationWorkflowController.ts`
- `apps/web/src/workbook/features/generic/GenericWorkbookInspectorPresentation.tsx`
- `apps/web/src/workbook/features/generic/createGenericMutationCommandPort.ts`
- `apps/web/src/workbook/features/generic/useGenericWorkbookInspectorComposition.tsx`
- `apps/web/src/workbook/hooks/useGenericSurfaceMutationController.ts`
- `apps/web/src/workbook/hooks/useWorkbookShellInfrastructure.ts`
- `apps/web/src/workbook/hooks/useWorkbookSurfaceQueries.ts`
- `apps/web/src/workbook/layout/WorkbookSurfaceLayout.tsx`
- `apps/web/src/workbook/mutations/createWorkbookMutationCommandPorts.test.ts`
- `apps/web/src/workbook/mutations/createWorkbookMutationCommandPorts.ts`
- `apps/web/src/workbook/mutations/workbookMutationCommandPorts.ts`
- `apps/web/src/workbook/query/useGenericSurfaceQuery.test.tsx`
- `apps/web/src/workbook/query/useGenericSurfaceQuery.ts`
- `apps/web/src/workbook/runtime/WorkbookConflictStore.ts`
- `apps/web/src/workbook/runtime/WorkbookExplicitPatchOwner.test.ts`
- `apps/web/src/workbook/runtime/WorkbookExplicitPatchOwner.ts`
- `apps/web/src/workbook/runtime/WorkbookMutationRuntime.ts`
- `apps/web/src/workbook/runtime/WorkbookSurfaceRegistry.ts`
- `apps/web/src/workbook/runtime/workbookConflictModel.ts`
- `docs/handoffs/workbook-task-lifecycle-refactor-handoff.md`
- `internal/modules/tasksdecisions/mutation_patch.go`
- `internal/modules/tasksdecisions/task_mutation_store_test.go`
- `packages/ui-contracts/src/entityEvidenceSelectors.ts`
- `packages/ui-contracts/src/workbook-interaction-selectors.test.ts`
- `tools/browser_e2e_batch_manifest.json`
- `tools/execution_topology_render_index.json`
- `tools/frontend_source_ownership.json`
- `tools/test_families/module.tasksdecisions.json`
- `tools/test_families/web.workbook.json`
