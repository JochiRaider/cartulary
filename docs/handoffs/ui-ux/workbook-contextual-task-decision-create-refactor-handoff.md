# Contextual Task Request and Decision creation handoff

## Baseline, authority, and scope

Execution baseline: `main`, `82c0b4c7de2f6eaa6b6810ebffb78b922a468bb3`, clean
worktree. Rechecked before implementation. The digest read order and adopted
owners were inspected during planning; the execution baseline is unchanged.
Core 01 §§3.3.4–3.3.7, 6, 7.4.1A, 7.4.8–7.4.9, and 18B own discovery,
creation, replay, fields, and references. Core 02 §10.4 owns Task/Decision
semantics. Core 03 §§2.3A, 4, 16.4 (REQ-03-255–256) and Core 04 §2 own
interaction, continuity, collaboration, and authorization. `docs/design.md`
owns presentation within its boundary. Digest, research, tests, and prior
handoffs are evidence, not behavioral authority. No executable artifact reads
Markdown as a behavioral input.

Only contextual `create_related.task_request` and `create_related.decision`
are in scope. One shared editable draft is retained in the existing account /
incident runtime and can resume from the workbook recovery area. Dispatched
attempts and receipts have independent identity. No browser persistence.
Standalone creation, existing-row editing, Task lifecycle, Decision
supersession, Assessment/Party/Notes/coordination creation, and Timeline
Evidence create-and-link are regression boundaries. The subsequently authorized
backend exception is the empty-array serializer repair and list/get regression
coverage described below. No dependencies, public-contract, or specification
changes are authorized. No commit, push,
deployment, digest edit, or analyst-data modification is authorized.

## Execution tracker

| Workstream | Dependency | State | Exit |
| --- | --- | --- | --- |
| CTD-01 baseline, owners, characterization, prerequisites | none | DONE | Current matrix and gaps checked; no blocking prerequisite. |
| CTD-02 retained drafts, authoring, discovery | CTD-01 | DONE | Retention, form, paging, and selection tests pass. |
| CTD-03 transport, attempts, receipts, recovery | CTD-02 | DONE | Unified dispatch and exact replay evidence passes. |
| CTD-04 reconciliation, security, accessibility, browsers | CTD-03 | DONE | Correlation/security, live browsers, a11y, reviewed promotion and two fresh visual validations pass. |
| CTD-05 final validation and completed handoff | CTD-04 | DONE | Finalization, all 595 frontend units, typecheck/lint, owner slices and scope review pass; handoff complete. |

Only the active row may be IN_PROGRESS. Record paths, commands, results, risks,
disposition, and next action before advancing. Never advance past a blocked
dependency.

## Source-feature-target matrix

Task means `create_related.task_request` → `cartulary.view.task_requests.v1`,
seed selected record ID into `task.linked_record_ids`; its reference inputs
are incident-member owner, requester Party, Decision, and linked records.
Decision means `create_related.decision` → `cartulary.view.decisions.v1`,
seed selected record ID into `decision.support_refs`; its reference inputs
are incident-member owner, supporting records, and affected records.

Every entry uses `view_row_create` / `view_row_create_route`, minimum role
`editor`, `requires_confirmation=false`, disabled conditions
`no_row_selected`, `incident_closed`, `authorization_lost`, success
`preserve_selected_row`, and failure
`show_same_shell_error_invalidate_pending_action`. Invalidation withdraws
stale submission readiness; it does not delete retained draft values.

| Source | Features | Composition | Source verification owner |
| --- | --- | --- | --- |
| Timeline | Task, Decision | Timeline inspector | module.workbook |
| Hosts | Task, Decision | Entity inspector | module.entities |
| Identities | Task, Decision | Entity inspector | module.entities |
| Evidence | Task, Decision | Generic inspector | module.evidence |
| Notes | Task, Decision | Generic inspector | module.artifacts |
| Indicators | Task, Decision | Generic inspector | module.indicators |
| Assessments | Task, Decision | Assessment inspector | module.assessments |
| Decisions | Task | Generic inspector | module.tasksdecisions |
| Communications Log | Task | Generic inspector | module.artifacts |
| Handoff | Task | Generic inspector | module.artifacts |
| Status Review | Task | Generic inspector | module.artifacts |
| Lesson | Task | Generic inspector | module.artifacts |
| Findings | Task, Decision | Generic inspector | module.artifacts |
| Investigative Queries | Task | Generic inspector | module.artifacts |
| Forensic Keywords | Task | Generic inspector | module.artifacts |

There are 23 entries on 15 sources. The three optional sources are already
implemented and emitted by `viewschema.ListPublicResources`. Task Requests and
Parties declare neither feature. Frontend routing is `web.workbook`; target
semantics use `module.tasksdecisions`. Source integration uses the table's
owners through narrow `test-slice` / `service-backed-test-slice` selections.
The existing Timeline Workflow browser row belongs to `module.workbook`.

## Gap ledger

| ID / classification | Authority | Remediation / affected areas | Rationale and long-term benefit | Compatibility impact / unresolved risk | Binary validation |
| --- | --- | --- | --- | --- | --- |
| G1 confirmed | Core 03 inspector/continuity; authorized retention | Replace Task/Decision component-local reducer lifetime with retained coordination owner and separate presentation attachment. | Keep analyst work while withdrawing stale readiness. | Excluded flows remain unchanged; lifecycle races require evidence. | Selection, sheet, version, and closure retain exact values and origin. |
| G2 confirmed | Core 01 queries/reference contracts; REQ-03-256 | Target-owned paged references replace empty/source-based arrays in Entity, Assessment, Timeline, and generic compositions. | Editable seeds and consistent discovery on every source. | Page absence is not deletion; local read failure is not access loss. | Both targets pass owner, Party, Decision, collection, paging, cancel, and stale-selection scenarios. |
| G3 confirmed | Core 01 §§7.4.8–7.4.9; Core 02 §10.4 | Contextual create validation, defaults, writable/input handling, and field-local feedback. | Prevent invalid initial lifecycle/scalar/reference writes without hiding server errors. | Standalone and lifecycle editors retain semantics. | Minima, timestamps, enums, lifecycle guards, exact IDs, and computed-field exclusion pass. |
| G4 confirmed | Core 01 ordinary create/replay and envelopes | One source-neutral transport with immutable attempts and full receipts replaces Task/Decision Timeline adapter use. | Recovery cannot accidentally create duplicates. | Unknown outcomes cannot be treated as rejection. | Lost/malformed response recovery sends identical body, route, actor scope, and key. |
| G5 confirmed | Core 03 collaboration/continuity | Runtime reconciliation and refresh debt for source, target, history, references. | Accepted results survive presentation changes. | Do not infer query membership or regress newer versions. | Failed post-acceptance refresh recovers with reads only in both HTTP/socket orderings. |
| G6 confirmed prerequisite, discovered in CTD-04 | Core 01 §18B public discovery requires empty `create_inputs=[]`; OpenAPI ViewSchemaResource requires an array | Applied the separately authorized fix in `internal/platform/viewschema/registry.go`: preserve empty arrays in `cloneCreateInputs`; added list/get serialization assertions in the existing owner tests. | Make authoritative discovery consumable by the existing typed HTTP validator. | Separate authorization received; resolved by red/green owner tests and live browser creation. No schema, route, storage, or migration change. | List/get serialize every empty create-input collection as `[]`; focused owner tests and real-service contextual creation pass. |
| H1 hypotheses | Core 03, Core 04, design | Characterize source removal/merge, role/session/account changes, late results, paging races, focus and containment. | Close race and security risks before completion. | No inherited failure classification. | Every scenario has passing evidence or blocks its dependent row. |

Observed nonblocking projection discrepancy: some optional Task create OpenAPI
properties are non-nullable while field-level contracts admit clearability.
Unset optional values can be omitted on this create seam. No public projection
repair is authorized or required for that behavior.

## CTD-01 investigation

Inspected the requested shared and Timeline workflow hooks/model/form/adapter;
Entity, generic, Assessment, and Timeline compositions; generic create builder
and decoder; reference options, membership adapter, bounded candidate readers,
picker controls; runtime registry/lifecycle, Assessment retained owner,
collaboration and surface refresh owners; target view schemas and workbook
OpenAPI; Task/Decision admission, transactional create, reference validation,
idempotency and history; related-workflow tests and `sentinel.spec.ts`.

Confirmed backend create includes target associations, projection, history,
collaboration intent, and replay result in one transaction. The HTTP boundary
rechecks current incident role before replay. No source concurrency token is
part of the target create contract. Membership and view queries expose paging;
current frontend adapters drop it. The generic all-record option inventory
omits supported source families; the new seam must resolve field-eligible,
publicly discovered surfaces instead.

Commands: both requested `make task-guide ROLE=module-author` calls passed
again at execution HEAD. Planning characterization passed on the same source:

- `make test-slice OWNER=web.workbook ROWS=web.workbook.regression.inspector_create_related_workflow_stale_results_31f0cbe2d1,web.workbook.regression.inspector_related_record_model_20a3b73f6d,web.workbook.regression.timeline_create_related_workflow_stale_effects_4b8ce2aa71`
  PASS; `.cartulary/test-results/20260912T205516Z-p76336`.
- `make test-slice OWNER=module.tasksdecisions ROWS=module.tasksdecisions.source_mutations.task_and_decision_mutation_admission_and_replay_6d4ea5900a,module.tasksdecisions.unit.task_decision_direct_references_admit_exact_stab_7bf2beff80`
  PASS; `.cartulary/test-results/20260912T205516Z-p76338`.
- `make service-backed-test-slice OWNER=module.tasksdecisions ROWS=module.tasksdecisions.source_workbook.task_requests_and_decisions_persist_as_workbook_887d6b579e,module.tasksdecisions.idempotency_adapter.application_workbook_idempotency_adapter_boundar_90a3399bf6`
  PASS; `.cartulary/test-results/20260912T205822Z-p80048`.

## Compatibility and rollback

CTD-01 disposition: DONE. Current code, contracts, owner clauses and unchanged
baseline evidence support the frontend remedy. No blocking prerequisite was
found. G1–G5 are confirmed; H1 remains a verification obligation. Next: retained
authoring and target-owned discovery, with colocated tests.

Keep public routes, schemas, saved views, field identities, authorization, and
server defaults stable. Do not introduce source patches, supersession, or
create-then-link for Task/Decision fields. Rollback reverts frontend and related
routing/visual changes only. Committed Tasks, Decisions, relationships, server
receipts, and history remain committed. Do not discard live retained recovery
state as a rollback step. Client retention has the existing runtime lifetime;
reload, tab closure, and browser persistence are outside its guarantee.

## CTD-02 disposition

Added the coordination-owned contextual draft model, retained owner, attachment
interface, target form and staged reference controls under
`apps/web/src/workbook/features/coordination/`. Added
`adapters/createContextualCreateReader.ts` with public schema discovery,
capability verification, bounded 100-row queries and membership paging.
Extracted the existing Assessment pager mechanics to
`hooks/useWorkbookCandidates.ts`; Assessment eligibility remains unchanged.
Registered the retained owner in runtime/session lifecycle, authored frontend
ownership and the focused semantic test row.

Validation:

- Authoring slice `web.workbook.regression.contextual_task_decision_authoring`:
  PASS, `.cartulary/test-results/20260912T211855Z-p5606` (five tests, including
  the full 23-entry mapping).
- `make frontend-typecheck`: PASS,
  `.cartulary/test-results/20260912T211856Z-p5888`.
- Assessment discovery regression slice: PASS,
  `.cartulary/test-results/20260912T211943Z-p6884`.
- Initial checks found missing frontend operation-family entries, exact-title
  catalog routing and unsupported test assertions; corrected before these passes.

Disposition: DONE. No external prerequisite. The new form and attachment are
ready for transport cutover in CTD-03; existing inspectors still use their prior
submission path until that cutover. HTTP discovery faults, cross-surface browser
behavior and responsive presentation remain validation risks for CTD-03/04.
Next: immutable attempts, complete receipt validation and retained recovery.

## CTD-03 disposition

Connected both inspector hooks to the retained contextual attachment and shared
form. Timeline's legacy target inventory now excludes Tasks and Decisions;
Notes, Evidence and coordination artifact paths remain operational. The
source-neutral `createContextualCreateTransport.ts` captures the immutable
request/body/path/API base, reviewed values, actor/incident scope and secure
transaction ID. Complete validated HTTP receipts precede presentation effects.
The runtime coordinates earlier source saves and accounts explicit creations
separately from autosave queue capacity. `ContextualCreateRecovery.tsx` exposes
retained drafts, uncertain attempts and read-only refresh recovery in the shell.

Validation:

- Focused authoring/discovery/recovery plus legacy shared/Timeline slice:
  PASS 6/6, `.cartulary/test-results/20260912T213425Z-p15518`.
- Recovery includes synchronous duplicate activation, prior-save invalidation,
  exact replay, timeout with late original acceptance, confirmed field rejection,
  role changes, suspension/retirement, both version-observation orderings,
  replacement drafts and accepted-plus-failed-refresh using reads only.
- All implemented reference surfaces and membership paging:
  PASS, `.cartulary/test-results/20260912T213349Z-p14417`.
- Import boundaries: PASS,
  `.cartulary/test-results/20260912T213013Z-p12199`.
- Typecheck: PASS, `.cartulary/test-results/20260912T213528Z-p21363`.
- Initial formatting/lint exposed new-code assertion/dependency/typing issues;
  corrected. `make format`: PASS,
  `.cartulary/test-results/20260912T213736Z-p22647`.
- An attempted diagnostic-only `BIOME_CHECK_FLAGS` override was rejected by
  public Make input admission; ordinary Make lint supplied the diagnostics.

Disposition: DONE. No backend/public-contract/specification prerequisite.
CTD-04 must validate live socket observation, source-specific browser assembly,
authorization-sensitive picker labels, focus and visual containment. No browser
or visual success is claimed by these unit results.

Final CTD-03 exit: ordinary Biome PASS
`.cartulary/test-results/20260912T213748Z-p27182`. A final typecheck found that
`PagingMeta` is not exported by the protocol facade; changed the local annotation
to the existing operation response's paging type. Typecheck PASS
`.cartulary/test-results/20260912T213848Z-p28177` before beginning CTD-04.

## CTD-04 historical blocked disposition

Implemented runtime collaboration observations, version high-water handling,
authorized-label invalidation, staged-picker focus handling, responsive recovery
controls, and three routed real-service browser scenarios in
`apps/web/e2e/contextual-create.spec.ts`. The new browser rows belong to
`module.entities`, `module.assessments`, and `module.evidence`; existing Timeline
coverage remains with `module.workbook`. Visual/a11y work is incomplete. The
support helper `apps/web/e2e/support/workbook/contextualCreate.ts` is prepared
for those checks but is not yet used by a visual or accessibility test.

Current evidence:

- `make test-slice OWNER=web.workbook ROWS=web.workbook.regression.contextual_task_decision_authoring,web.workbook.regression.contextual_task_decision_discovery,web.workbook.regression.contextual_task_decision_recovery`:
  PASS 4/4, `.cartulary/test-results/20260912T214720Z-p34958`.
- `CARTULARY_GENERATE_DRIFT_REFRESH=1 make generate-drift`: PASS 4/4,
  `.cartulary/test-results/20260912T214616Z-p30852`. Make generated the affected
  browser batch manifest and execution topology render index from authored rows.
- `make service-backed-test-slice OWNER=module.entities ROWS=module.entities.browser.contextual_task_replay`:
  FAIL, `.cartulary/test-results/20260912T214719Z-p34725`. The first run exposed
  an incorrect new test helper: grid focus was mistaken for Entity inspector
  selection. Corrected the helper to use the existing Entity inspect control.
- The same browser command after that correction: FAIL,
  `.cartulary/test-results/20260912T215000Z-p67259` (9/11 execution units pass;
  browser assertion and aggregate fail). The form renders, but review cannot
  verify the source/target contracts because live schema responses are invalid.
- `make task-guide ROLE=module-author OWNER=platform.viewschema`: PASS.
- `make test-slice OWNER=platform.viewschema ROWS=platform.viewschema.unit.source_and_public_projection,platform.viewschema.unit.view_schema_discovery_exposes_exact_sorting_and_94fda0d98a`:
  PASS, `.cartulary/test-results/20260912T215438Z-p99813`. Existing assertions
  compare slice contents/length and do not check the serialized empty array.

The second browser trace contains HTTP 200 responses from both
`/api/v1/view-schemas/cartulary.view.hosts.v1` and
`/api/v1/view-schemas/cartulary.view.task_requests.v1` with
`data.create_inputs: null`. The trace is under that run's
`browser-e2e-webserver-backed/browser-groups/functional-support-default-contextual-create/playwright-output/`
directory. This is current real-service evidence, independent of prior failure
classifications.

Core 01's public discovery clause at
`docs/spec/01_architecture_storage_and_view_contracts.md:4755` explicitly requires
`[]` when no create-only input is declared. The matching authored OpenAPI
projection at
`contracts/openapi-source/owners/platform.viewschema/openapi.json:591` requires
an array. `internal/platform/viewschema/registry.go:318` clones using a nil
destination slice, producing JSON `null` for empty input. Both public list and
lookup paths use this helper. The frontend reader uses the existing typed HTTP
validator; no tolerance for the malformed response has been added.

Concrete backend proposal at the blocked checkpoint, **not yet applied then**:

```diff
 func cloneCreateInputs(inputs []CreateInputDescriptor) []CreateInputDescriptor {
-    return append([]CreateInputDescriptor(nil), inputs...)
+    return append([]CreateInputDescriptor{}, inputs...)
 }
```

Extend `TestOpenAPIViewSchemaPublicProjectionContract_Unit` in
`internal/platform/viewschema/viewschema_test.go` to assert the declared array
type and serialized array value for every list resource and corresponding
lookup resource, retaining exact descriptor contents and order. This repairs
an implementation projection of the existing contract; it proposes no public
API or specification change and no changes to Task/Decision persistence.

Disposition: BLOCKED on G6. The earlier no-prerequisite findings record what was
known at their exits; this browser evidence supersedes that assumption. Do not
advance to CTD-05. Next action: obtain separate authorization for this bounded
backend helper repair and regression coverage, then rerun the discovery owner
slice and resume CTD-04 browser, accessibility, visual, and regression work.

Outstanding risks include full socket-envelope correlation coverage, all
source-specific browser paths, protected-label transitions in the live UI,
focus/containment, and excluded-workflow browser regressions. Frontend unit
passes do not establish these outcomes. No affected visual has been promoted.
`make agent-finalize`, broader end-of-run verification, and CTD-05 completion
remain pending the blocked dependency. `RESULTS_DIR` remains unset; retained-run
maintenance has not run and no exact-source full warm evidence is claimed.

Independent checks of the frontend work already written, before requesting
authorization:

- `make format`: PASS,
  `.cartulary/test-results/20260912T215802Z-p1383`.
- `make frontend-typecheck`: PASS,
  `.cartulary/test-results/20260912T215916Z-p6330`.
- `make lint-biome`: PASS,
  `.cartulary/test-results/20260912T215916Z-p6364`.
- `make frontend-import-boundary-check`: PASS,
  `.cartulary/test-results/20260912T215916Z-p6341`.
- `make lint-markdown`: PASS,
  `.cartulary/test-results/20260912T215916Z-p6382`.
- `git diff --check`: PASS. Scope review confirms no backend, contracts,
  generated protocol/UI-contract code, database, or digest edits. All changes
  remain uncommitted on the recorded branch and HEAD.

These checks do not complete CTD-04 or CTD-05. Markdown validation of this blocked
handoff and a final whitespace/scope recheck are recorded in the turn report.

## CTD-04 resumed authorization

The user separately authorized the proposed empty-array helper repair and
list/get serialization regression tests, and instructed resumption through
CTD-05. Branch and HEAD are unchanged; the worktree contains this seam's retained
implementation. The authorization covers this bounded backend implementation
repair; public contracts, specifications, storage, and other backend behavior
remain outside the authorized change. G6 remains open until the owner tests and
live contextual creation pass. CTD-04 is the only IN_PROGRESS row.

The serializer regression failed against the original helper at
`.cartulary/test-results/20260912T220711Z-p11583`, then the same two-row
`platform.viewschema` selection passed after the authorized one-line repair at
`.cartulary/test-results/20260912T220731Z-p12118`. Assertions cover public list
and lookup JSON arrays and descriptor contents/order. G6 is now resolved: live
discovery and contextual creation pass. The optional Task clearability
discrepancy remains nonblocking and unchanged.

Further CTD-04 characterization and remediation:

- Public seed bindings are compared structurally, preserving binding order but
  ignoring object property order. The first browser rerun exposed the incorrect
  serialization comparison; a new discovery regression covers all 23 entries
  and rejection of changed create capability. PASS at
  `.cartulary/test-results/20260912T221218Z-p8801`.
- Corrected the new Assessment fixture to supply its required initial state.
  Corrected both stages of the new picker-cancel browser helper to locate the
  surface combobox by its accessible role/name. Failed runs
  `20260912T220742Z-p12566`, `20260912T220757Z-p38200`,
  `20260912T220757Z-p38216`, `20260912T221247Z-p9640`,
  `20260912T221659Z-p47783`, and `20260912T222006Z-p47847` document those
  superseded failures; the latter two repeated the helper's second-stage locator.
- Full socket envelopes are retained and correlated with HTTP change sets.
  Removal high-water marks are scoped to their declared views; empty pages
  never establish deletion. Late related-source socket observations extend the
  runtime's read refresh plan and mounted/unmounted projection debt. Replay
  rechecks current declared capability using the immutable original draft.
- Presentation attachments move keyboard focus into the resumed form and restore
  the originating control only while that form still owns focus. The real
  Entity test delays replay acceptance until focus moves to the newly selected
  surface, then proves completion does not steal focus.
- Reviewed ordinary visuals exposed browser-default reference-surface styling
  and a misleading Conflict save label after accepted creation with refresh
  debt. Applied existing input tokens, kept the accepted operation Saved,
  identified the default owner as Current actor, and made stable reference IDs
  and full receipt details expandable and keyboard-accessible.

Current browser evidence (all commands use `make service-backed-test-slice`):

| Owner / rows | Result | Run root below `.cartulary/test-results/` |
| --- | --- | --- |
| `module.entities` / `module.entities.browser.contextual_task_replay` | PASS 11/11; exact replay, edited seeds, owner/Party/Decision selection, picker cancel, late focus, one creation/history effect, no source PATCH | `20260912T222818Z-p60642` |
| `module.assessments` / `module.assessments.browser.contextual_decision_refresh` | PASS 11/11; retained closed inspector, accepted creation remains Saved, refresh recovery performs reads only, narrow recovery and Escape | `20260912T222818Z-p60632` |
| `module.evidence` / `module.evidence.browser.contextual_creation_discard` | PASS 11/11; retained Task draft requires explicit discard before Decision replacement | `20260912T221247Z-p9673` |
| `module.workbook` / `module.workbook.browser_stateful.verify_timeline_inspector_workflow_create_relate_7fc4833af4` | PASS 11/11; both contextual targets and excluded Timeline Evidence/coordination workflows | `20260912T221247Z-p9682` |
| `module.workbook` / `module.workbook.accessibility.contextual_task_decision_creation` | PASS 11/11; both targets, field errors, timestamp errors, picker cancel/Escape, focus, 1280/768/390 widths, 200% zoom, text spacing and recovery | `20260912T222818Z-p60686` |

The three focused frontend rows passed 4/4 at
`.cartulary/test-results/20260912T222733Z-p55354`. Typecheck passed at
`.cartulary/test-results/20260912T222752Z-p59535`; the subsequent Biome run
`20260912T222808Z-p60022` found only the newly extended test's formatting, now
handled through Make formatting. Earlier typecheck
`20260912T221659Z-p47902` exposed union narrowing across a callback, corrected
before the current pass. Make generation passed at
`.cartulary/test-results/20260912T222732Z-p55102` after authored browser scenario
IDs were corrected to the harness's required `scenario_` plus 12-hex format.

First full ordinary visual run
`.cartulary/test-results/20260912T222006Z-p48094` passed all existing visual
scenarios and reached all ten new contextual captures. Missing new goldens
caused comparison failures; the initially invalid authored scenario ID also
prevented reconciliation publication. All ten screenshot attachments were
extracted without image modification into that run's `contextual-review/`
directory and individually reviewed before the presentation corrections above.
No golden promotion is claimed yet. The corrected ordinary run, reconciliation,
review, promotion, and two fresh ordinary validations remain CTD-04 exit checks.


## CTD-04 final evidence

The corrected ordinary visual run
`.cartulary/test-results/20260912T222818Z-p60924` reached every capture and
failed only for the ten new missing goldens. Its
`browser-e2e-visual/frontend-visual-reconciliation.json` reports 231 capture
intents, 221 active committed goldens, ten missing, no orphan or ambiguous
mapping, and no unresolved registered fixture. All ten ordinary screenshot
attachments were individually inspected in `contextual-review/` before
promotion. Inputs, selected references, default owner, field containment,
scrolling and retained/accepted recovery were reviewed at 1280×720 and 768×640.

`make browser-e2e-visual-update`: PASS 12/12,
`.cartulary/test-results/20260912T223431Z-p98085`. Its
`browser-e2e-visual-update/frontend-visual-reconciliation.json` reports all
231 goldens active, zero missing/orphan/ambiguous mappings and zero unresolved
registered fixtures. The ten promoted files were individually inspected again;
the 221 existing goldens are unchanged. Make regenerated the golden manifest.
No renderer, viewport, tolerance, theme, density, mask or pinned-browser change
was used to obtain a pass. These are implementation/design evidence, not a
Core 05 claim publication.

All ten captures belong to `module.workbook`, row
`module.workbook.visual.contextual_task_decision_creation`, scenario
`scenario_4a08391df78c`, project `chromium`. Paths below are relative to
`apps/web/e2e/workbook.visual.spec.ts-snapshots/`.

| Capture ID | Reviewed golden |
| --- | --- |
| `visual.capture.44e851b61fcacf540838` | `contextual-decision-authoring-linux.png` |
| `visual.capture.b0c05164823b82329162` | `contextual-decision-authoring-narrow-linux.png` |
| `visual.capture.597c2cd3c6d4b9ed069b` | `contextual-decision-recovery-linux.png` |
| `visual.capture.ea7796e054f33716ba12` | `contextual-decision-recovery-narrow-linux.png` |
| `visual.capture.684e63e6fb29722ef1e9` | `contextual-decision-references-narrow-linux.png` |
| `visual.capture.010db75b102f203f2286` | `contextual-task-request-authoring-linux.png` |
| `visual.capture.67c158eea44215a72519` | `contextual-task-request-authoring-narrow-linux.png` |
| `visual.capture.0a3f3690592b622a6288` | `contextual-task-request-recovery-linux.png` |
| `visual.capture.3c66cdc666f7ed8903f6` | `contextual-task-request-recovery-narrow-linux.png` |
| `visual.capture.08039628dabdd01901dd` | `contextual-task-request-references-narrow-linux.png` |

Excluded-workflow browser regressions also pass through public
`make service-backed-test-slice OWNER=<owner> ROWS=<rows>` selections:

| Owner / rows | Result | Run root below `.cartulary/test-results/` |
| --- | --- | --- |
| `module.tasksdecisions` / `module.tasksdecisions.browser.task_lifecycle_patch_recovery,module.tasksdecisions.browser.decision_supersession_exact_recovery` | PASS 11/11 | `20260912T223505Z-p29239` |
| `module.assessments` / `module.assessments.browser.append_recovery,module.assessments.browser.the_browser_workbook_drives_the_assessments_surf_4f277414a9` | PASS 13/13 | `20260912T223505Z-p29251` |
| `module.parties` / `module.parties.browser.the_browser_workbook_surface_keeps_ordinary_evid_687c58e86e` | PASS 11/11 | `20260912T223505Z-p29270` |
| `module.links` / `module.links.browser.the_notes_tab_supports_browser_visible_creation_ae6bdcc2da` | PASS 11/11 | `20260912T223711Z-p32245` |

Latest focused discovery/recovery tests pass 3/3 at
`.cartulary/test-results/20260912T223827Z-p63534`; typecheck passes 2/2 at
`.cartulary/test-results/20260912T223833Z-p64204`. These cover stricter route,
feature, field/input capability comparison and error-envelope correlation.
The preceding typecheck caught unsupported Testing Library `exact` options;
those assertions now use the supported accessible-name matcher. Both-target
stale/failed-page and role-change tests pass at
`.cartulary/test-results/20260912T223259Z-p97034`. The new visual and a11y
scenarios use `e2e/support/workbook/contextualCreate.ts`; its earlier unused
status was limited to the blocked checkpoint.


Additional exit evidence: the refreshed Evidence contextual-discard browser
selection passes 11/11 at `20260912T224506Z-p33851`; the refreshed existing
Timeline Workflow selection passes 11/11 at `20260912T224507Z-p34792`. The
both-target source removal/merge-notification characterization added to the
retention test passes 2/2 at `20260912T224607Z-p99919`: exact source identity,
seed IDs and entered values survive, with review readiness invalidated.
Runtime receipt forwarding updates monotonic Task/Decision caches and history;
`WorkbookSurfaceRegistry.refreshIfMounted` retains unmounted refresh debt.
These paths do not insert target rows into current query membership.

Exact initial seam paths inspected include:

- `apps/web/src/workbook/inspector/useInspectorCreateRelatedWorkflow.ts`,
  `inspectorRelatedRecordModel.ts`, and `InspectorCreateRelatedWorkflow.tsx`.
- `apps/web/src/workbook/timeline/hooks/useTimelineCreateRelatedWorkflow.ts`
  and `timeline/adapters/createTimelineRelatedRecordCommandAdapter.ts`.
- `apps/web/src/workbook/runtime/WorkbookMutationRuntime.ts`,
  `WorkbookSurfaceRegistry.ts`, and `WorkbookExplicitPatchOwner.ts`.
- `apps/web/src/workbook/models/genericWorkbookModel.ts` and
  `workbookRequestDecoders.ts`; the Assessment candidate reader and picker
  consumers before the bounded pager extraction.
- `apps/web/e2e/sentinel.spec.ts`, the source inspector compositions and their
  existing related-workflow unit tests; Task/Decision schema/OpenAPI projections
  and `internal/modules/tasksdecisions/` admission/create/replay owners.

Final frontend ownership routes the coordination model, owner, form, recovery,
reference control, transport, reader and tests through `web.workbook`. The
source-specific browser routes remain with Entities, Assessments, Evidence and
Workbook; visual/a11y rows remain with Workbook. No digest or adopted owner was
used as a runtime/test input or edited as part of this implementation.


CTD-04 disposition: DONE. Both fresh ordinary `make browser-e2e-visual`
validations passed 12/12 after promotion:
`20260912T224331Z-p66024` and `20260912T224332Z-p66221`. Each reconciliation
reports 231 active goldens, zero missing/orphan/ambiguous mappings and zero
unresolved registered fixtures. All 43 visual scenarios pass in each run.
No excluded-workflow or current prerequisite failure remains unresolved.

H1 characterization is covered by source/version/removal retention tests,
both-target stale reference and role transitions, session/account retirement,
late HTTP acceptance, both socket orderings, off-page IDs, and real keyboard /
responsive browser evidence. Merge removal keeps original IDs; it never
silently selects the survivor. Invalid references remain subject to authoritative
create admission. The source-specific browser suite uses isolated harness data.
Residual limits are the existing runtime lifetime and server-authoritative
visibility/admission; no browser persistence or public-contract change was
introduced. Next: CTD-05 finalization and broader current-source validation.

## CTD-05 final validation

`RESULTS_DIR` is unset: no exact-source successful full warm `check` run exists
for this implementation. Retained-run maintenance is intentionally excluded;
`make agent-finalize` is run before broader end-of-run verification.


Initial `make agent-finalize` failed at shape validation in
`20260912T224802Z-p2016`. Direct `make json-shape-check` reproduced the reason
at `20260912T224830Z-p2532`: authored `web.workbook` test titles had changed
since the previous generated topology fingerprint. This is change-related
metadata drift, not a product or backend prerequisite. Remediation uses
`CARTULARY_GENERATE_DRIFT_REFRESH=1 make generate-drift` before retrying the
ordinary finalizer; generated files are not edited by hand.


Generated refresh passed 4/4 at `20260912T224849Z-p3115`. The subsequent
ordinary `make agent-finalize` passed 1/1 at `20260912T224905Z-p6885` before
broader verification. Its `unit-artifacts/finalize-summary.json` records
schema-shape validation, catalog/tier coverage and generated-structure refresh
as passing. Canonical retained evidence and scheduler drift maintenance are
explicitly skipped with `results-dir-not-provided`; no full warm performance,
conformance or release evidence is claimed.


The first full frontend-unit run at `20260912T224937Z-p11201` exposed three
change-related integration issues: the exact source-ownership allowlist did not
name the new coordination request/envelope owners and their characterization
tests; three new test consumers used raw reference-picker IDs; the older
Timeline inspector fixture did not attach the retained contextual owner.
Updated the explicit ownership assertion, used accessible picker role/name
selectors, and attached the retained owner and source-neutral transport to that
focused Timeline test. Its assertion now checks the authoritative retained
receipt instead of obsolete inspector-local success text. This changes test
composition/ownership, not product semantics or the reviewed visuals.

Already passing broader checks: frontend typecheck `20260912T224937Z-p11165`,
import boundaries `20260912T224937Z-p11290`, Biome `20260912T224937Z-p11344`,
full ViewSchema owner slice `20260912T224937Z-p10985`, generated artifact policy
`20260912T224937Z-p10968`, and Go formatting `20260912T224937Z-p12240`.
The full Task/Decision owner slice passes 21/21 at `20260912T224937Z-p11080`;
the persistence/replay service selection passes 4/4 at
`20260912T224937Z-p11046`. Narrow Artifacts surface projection/registry rows
pass 2/2 at `20260912T224959Z-p38118`; the Indicator inspector metadata/paging
row passes 2/2 at `20260912T225000Z-p40032`.


The first full frontend run finished with 592/595 units passing; only the three
issues above failed. Focused corrections pass: architecture ownership/selector
rows 3/3 at `20260912T225321Z-p39189`, the existing Timeline inspector row 2/2
at `20260912T225321Z-p39196`, and contextual authoring 2/2 at
`20260912T225321Z-p39257`. `make task-guide ROLE=module-author OWNER=web.architecture`
confirmed the narrow routing. The full frontend suite is repeated after another
ordinary finalizer pass; no successful full-suite result is inferred from the
focused passes.


The updated accessible selectors also pass in real-service browser execution:
Entity contextual replay 11/11 at `20260912T225321Z-p39270`, and both-target
contextual accessibility 11/11 at `20260912T225321Z-p39353`. The existing
rendered UI and goldens are unchanged by these test-only corrections.

## Final ledger disposition and compatibility

| Gap | Disposition / binary evidence |
| --- | --- |
| G1 | RESOLVED: all 23 mappings, exact retained values/source IDs, detachment/navigation/closure, relevant version/removal invalidation and explicit replacement pass in authoring, recovery and source-browser rows. |
| G2 | RESOLVED: all 17 implemented public reference surfaces use bounded pages of 100; membership paging, owner/Party/Decision controls, editable seeds, off-page IDs, stale/failure/retry and staged Cancel/Escape pass. |
| G3 | RESOLVED: both targets' normalized minima, defaults, writable boundaries, enums/timestamps, exact IDs and initial lifecycle guards pass; unset optional fields are omitted. |
| G4 | RESOLVED: one retained source-neutral port, synchronous duplicate reservation, earlier-save ordering, complete receipt/error correlation, immutable original replay and deliberate new attempts pass; live aborted-response replay creates one target/history effect. |
| G5 | RESOLVED: accepted receipt precedes effects, survives replacement/suspension and filtered/page absence, preserves newer versions in both HTTP/socket orderings, and failed refresh recovers using reads only. Related-view observations extend retained projection refresh debt. |
| G6 | RESOLVED: authorized one-line serializer repair preserves `[]` in public list/get; red/green serialization tests, full ViewSchema owner slice and live contextual creates pass. |
| H1 | CHARACTERIZED: source merge/removal, current capability/role/session/account transitions, protected-label invalidation, paging races, late acceptance, focus/scroll continuity and responsive/keyboard recovery have current passing tests and reviewed browser artifacts. |

No specification or public API change, dependency, migration, standalone
creation behavior or excluded workflow business rule was introduced. Existing
Task/Decision atomic relationship/history and replay behavior remain the
backend authority. The only backend change is the separately authorized
empty-array clone repair. The optional Task nullability/clearability discrepancy
remains documented and untouched; this seam omits unset optional create values.

Retention lasts for the supported account/incident runtime only. Full warm
`make check`, release/conformance publication and retained performance/scheduler
maintenance were not run; owner-selected unit/service/browser coverage and
ordinary visual/a11y evidence are reported here without extending those claims.
No remaining blocker or unclassified observed failure is known after the
specific dispositions above. Final-suite completion is recorded below.

Rollback removes this frontend seam and its associated authored/generated
routing and visual changes. The independently correct serializer repair may
remain. Coordinate frontend replacement with recovery of live retained
attempts; do not discard or force-reload that state as a rollback mechanism.
Committed Tasks, Decisions, relationships, receipts and history survive rollback.
Reverting frontend code never undoes those committed operations.


The second finalizer passed at `20260912T225415Z-p4386`. Subsequent typecheck
`20260912T225502Z-p8536` caught an inferred `string` role in the updated
Timeline unit fixture. Added its explicit `WorkbookMutationAuthority` annotation;
this is a type-only correction. The full frontend rerun has no observed failures
at this checkpoint; final results follow below.

Scope review: `main` remains at
`82c0b4c7de2f6eaa6b6810ebffb78b922a468bb3`. The worktree contains only this seam:
frontend/runtime/tests, authored routing/ownership, Make-generated derivatives,
ten added goldens, the two authorized ViewSchema backend files, and this
handoff. No existing golden changed. No contracts, packages, dependencies,
lockfiles, database, digest, commit, push, deployment or analyst-data change.
The source-owner assertion in `apps/web/src/testing/sourceOwnershipPolicy.test.ts`
and Timeline inspector test are intentional final-validation updates.


Final typecheck passes 2/2 at `20260912T225659Z-p43556`; final Biome passes
2/2 at `20260912T225700Z-p43948`. Preliminary Markdown lint passed at
`20260912T225503Z-p8613` and whitespace checks pass. The completed handoff will
receive the required final Markdown/whitespace/scope recheck after CTD-05 is
marked DONE.

| Final verification command | Result / run root below `.cartulary/test-results/` |
| --- | --- |
| `make agent-finalize` | PASS `20260912T225415Z-p4386`; retained-run actions explicitly skipped |
| `make frontend-unit` | PASS 595/595 `20260912T225502Z-p8525` |
| `make frontend-typecheck` | PASS 2/2 `20260912T225659Z-p43556` |
| `make frontend-import-boundary-check` | PASS 2/2 `20260912T224937Z-p11290` |
| `make lint-biome` | PASS 2/2 `20260912T225700Z-p43948` |
| `make lint-go-format` | PASS `20260912T224937Z-p12240` |
| `make generated-artifact-policy-check` | PASS 3/3 `20260912T224937Z-p10968` |
| `make test-slice OWNER=platform.viewschema` | PASS 1/1 `20260912T224937Z-p10985` |
| `make test-slice OWNER=module.tasksdecisions` | PASS 21/21 `20260912T224937Z-p11080` |
| `make service-backed-test-slice OWNER=module.tasksdecisions ROWS=module.tasksdecisions.source_workbook.task_requests_and_decisions_persist_as_workbook_887d6b579e,module.tasksdecisions.idempotency_adapter.application_workbook_idempotency_adapter_boundar_90a3399bf6` | PASS 4/4 `20260912T224937Z-p11046` |

Frontend/source integration, architecture, accessibility, excluded workflows and
both ordinary post-promotion visual runs are recorded in the earlier exact row
and artifact tables. No broad full warm check, complete service/browser inventory,
Go static-analysis suite or release gate is claimed: validation broadened only
to the owners, frontend boundary and visual surface affected by this seam.


CTD-05 disposition: DONE. The corrected full `make frontend-unit` run passes
595/595 units with zero failed, skipped or cancelled units at
`20260912T225502Z-p8525`. The only subsequent unit-source adjustment was the
explicit fixture authority type, validated by the passing final typecheck;
product and visual code did not change. All ledger gaps are resolved or
characterized as recorded. No blocker remains. Authoritative backend behavior,
excluded workflows, ownership/routing and generated derivatives are preserved.

The completed handoff is followed by one final `make lint-markdown`,
`git diff --check`, and branch/HEAD/scope review; their terminal results are
reported with this handoff in the completion response. No additional product
work is planned. Stop after this seam. All changes remain uncommitted.
