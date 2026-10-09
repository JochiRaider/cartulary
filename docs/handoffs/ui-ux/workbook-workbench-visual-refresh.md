# Workbook workbench visual refresh record

Active iteration: **DONE — workbench legacy retirement and scoped production readiness**.
Date: **2026-10-08**. The user authorized implementation of WB-W0 through WB-04,
including retirement of standalone navigation. The completed visual-refresh evidence begins at
[Completed visual refresh record](#completed-visual-refresh-record).

## Execution checkpoints

This section records the current execution. Earlier planning-session entries are
historical evidence; their document-only restrictions do not restrict this
authorized implementation. Update this tracker after completing each workstream
and before starting the next one. Required failed exits remain BLOCKED.

| Workstream | Status | Dependencies | Exit / evidence |
| --- | --- | --- | --- |
| WB-W0: specification alignment and baseline | DONE | None | Revalidated unchanged HEAD and callers; expanded baseline PASS at `.cartulary/test-results/20261008T232521Z-p76916` (four units). |
| WB-01: narrow navigation dependencies | DONE | WB-W0 | Explicit host/actions and typed fixture; focused tests, typecheck, source ownership and import checks passed (execution log). |
| WB-02: retire alternate consumers | DONE | WB-01 | Required Note/saved-view navigation; composition, authority and browser checks passed. Inspector completion correction included. |
| WB-03: remove dead surfaces | DONE | WB-02 | Removed wrapper/forwarding and stored controller; startup, saved-view operations, cancellation and composition pass. |
| WB-04: validation and handoff completion | DONE | WB-03 | All final gates PASS; every digest acceptance row PASS or justified N/A; completed handoff in §13. |

### Remediation decisions

| Finding | Areas | Remediation and rationale / long-term benefit | Compatibility | Unresolved risk | Validation |
| --- | --- | --- | --- | --- | --- |
| F-01 | Implementation, types, tests, docs | Navigation declares only consumed host capabilities; shell adapts them. Compiler-checked fixtures replace the aggregate-runtime cast, reducing coupling and authority. | Private TypeScript cutover, no alias. | Shell growth hides dependencies and missing fixture behavior. | No shell-runtime type dependency; typed fixture and admission/readiness checks. |
| F-02 | Implementation, tests, docs | Required semantic Note navigation replaces providerless page lookup and its row/error state. One owner admits destinations. | Retire unsupported standalone execution; migrate callers and fixtures together. | Divergent stale-result, disclosure, Return and selection behavior. | Real Note navigation, no scanning, retained authoring and fenced attachment. |
| F-03 | Implementation, tests, docs | Saved-view browser uses required navigation for base and saved selections. Keep resource operations in their owner. | Remove browser fallback props; retain current-resource recovery and confirmed-operation opening. | Tests exercise different admission from production. | Same-view reselection, origin retention, owner operations and preference fixtures. |
| F-04 | Implementation, tests, docs | Remove unused forwarded selection command, keeping the internal owner binding and startup acceptance. | Internal removal without deprecation alias. | Future callers revive the bypass. | Caller review, typecheck, startup and saved-view owner tests. |
| F-05 | Implementation, tests | Remove stored pending controller; retain local controller, active intent and request identity. Clarifies cancellation ownership. | No external/state migration. | Ambiguous cancellation ownership complicates maintenance. | Coalescing, cancellation, supersession and completion tests. |
| F-06 | Docs | Add navigation/Commands guides and parent links. Make ownership and verification discoverable. | No runtime effect. | Boundaries drift or retired paths return. | Guide/source review and Markdown lint. |
| F-07 | Retention, tests, docs | Keep live authoring reads, saved-view owner operations, startup and required invoker metadata. | Retain actual supported operations, not duplicate navigation. | Broad deletion breaks creation/recovery or owner obligations. | Live caller/owner evidence and affected tests. |
| F-08 | Architecture, docs | Reuse query, Grid Adapter and source-operation boundaries; no generic store/engine/freshness registry. | No new package, protocol or persistence. | Duplicate authority increases coupling. | Semantic injection, no extra row/draft/receipt store. |
| V-01 | Tests, routing, docs | Add real composition and authority-lifetime evidence alongside unit doubles. | Update authored selectors with cases; retain coverage. | Isolated tests conceal wiring and late-attachment defects. | Routed production admission, selection, retention, concealment and focus cases. |

### Execution baseline and scenario placement

HEAD remains `c7716b3dc25332b637b90ec2fd6b3915db2c94ed` on `main`.
The tracker was already modified by the previous planning task; those edits and
the completed visual-refresh record are preserved. Core 01 REQ-01-680–684,
Core 03 REQ-03-311–318, Core 04 REQ-04-169–170, Design and Domain already own
the required behavior. No normative contradiction or required Core, Design,
Domain, protocol or typed behavioral projection change was found. The changes
below repair implementation and documentation; historical code is not authority.

The seven frontend rows in §8 passed during planning at
`.cartulary/test-results/20261008T231123Z-p71452` (eight execution units including
installation). The execution baseline additionally selects
`web.workbook.regression.preference_characterization`,
`web.workbook.regression.useworkbooksavedviewcontroller_suite_d3f1a57e5d`, and
`web.workbook.regression.workbookshell_surfaces_suite_668e482b1e`.

Admission/outcome and session lifetime cases belong in the existing navigation
tests. Real Note and saved-view cutover cases belong in the shell surface suite;
selector/preferences fixtures supply typed action doubles. Existing Note replay,
saved-view owner, responsive Commands and interrupted-navigation browser cases
remain required. Authored selectors change alongside new cases. Broader freshness
policy consolidation remains DEFERRED because source contributor sets differ.

### Execution session log

| Checkpoint | Changes / inspected boundaries | Commands and results | Risks / next action |
| --- | --- | --- | --- |
| WB-W0 complete, 2026-10-08 | Revalidated navigation, shell/startup, saved-view composition and preference fixture; updated tracker only. No owner contradiction or normative correction needed. | `git status --short`, `git rev-parse HEAD`, `make task-guide ROLE=module-author OWNER=web.workbook`; expanded three-row `make test-slice` PASS at `20261008T232521Z-p76916`. Previous seven-row baseline remains at `20261008T231123Z-p71452`. | No failing baseline. Next: WB-01 explicit host; preserve async observation and presentation fencing. |
| WB-01 complete, 2026-10-08 | Added navigation host, narrow action type, explicit shell adapter, complete typed fixture, ownership entry and navigation guide/parent link. No shell-hook return type remains in navigation. | Focused five-row test-slice PASS `20261008T232720Z-p79477`; import check PASS `20261008T232720Z-p79645`; typecheck PASS `20261008T232842Z-p85252`; `make format-frontend` passed. | Typecheck first exposed widened fixture filter operator at `20261008T232720Z-p79620`; fixed with the owner query type. Roll back host, shell adapter and fixture together. Next: WB-02. |
| WB-01 gate repair, 2026-10-08 | Architecture run exposed a pre-existing `getByTestId("candidates")` in the authoring picker test (confirmed at HEAD). Replaced this single test lookup with its labeled Parties group; no product or selector policy change. | Architecture 12/13 passed at `20261008T232720Z-p79479`; failed selector row passed after repair at `20261008T232842Z-p85150`. Full authoring candidate row passed at `20261008T232842Z-p85157`. | Unrelated baseline test maintenance needed for the required gate; final architecture gate will rerun in WB-04. |
| WB-02 complete, 2026-10-08 | Required action capability now flows through shell, surface facade, generic inspector, view-bar binding, selector and browser. Deleted Note page-scan/row/error branch and browser activation/base fallbacks. Added shell Note/Return, saved-view reselection/base, closed-read and authority-lifetime cases; unsupported locator and action-dispatch cases; authored row selectors updated. | Six-row test-slice `20261008T233505Z-p89510` passed five rows; new Note completion assertion failed. Follow-up shell/admission PASS `20261008T234142Z-p42921`; typecheck PASS `20261008T234143Z-p43234`; `make test-catalog-check` PASS. Selected three browser/accessibility rows PASS, 15/15 units, `20261008T234226Z-p45534`, including added Note completion/focus assertions. | Found a real V-01/F-02 completion defect: successful inspector attachment makes grid focus unavailable, so an extra grid-readiness requirement left navigation admitted forever. Separated mounted accepted destination completion from grid-only eligibility in the browsing registry/hook. No geometry/token/golden change. Roll back this correction with the cutover, callers and tests. Next: WB-03. |
| WB-02 failed-check classification | New test typing corrected (`fetch` mock and role options); no runtime type failure. Note readiness reproductions at `20261008T233726Z-p92562`, `20261008T233819Z-p94085`, `20261008T233916Z-p38482`, `20261008T234020Z-p41218` traced the completion defect; temporary diagnostics removed. | Initial typecheck failed `20261008T233507Z-p89919`, then passed `20261008T233726Z-p92641`. Pre-correction browser scenarios passed `20261008T233820Z-p94439` but lacked explicit completion assertion; superseded by the enhanced browser pass above. | These failed runs are diagnostic evidence, not exits. All WB-02 required checks now pass. |
| WB-03 complete, 2026-10-08 | Removed the returned `selectSavedView` wrapper and shell forwarding, plus `pending.controller`. Reference review leaves only the internal controller selection binding. Kept `activateResource`/`openConfirmed`, startup `upsertSavedView`, shared authoring reads, local abort intent and Return invoker metadata. | Four-row cancellation/admission/controller/shell slice PASS `20261008T234405Z-p91086` (5/5 units); saved-view operations/startup slice PASS `20261008T234417Z-p92655` (3/3); typecheck PASS `20261008T234407Z-p91402`; `git diff --check` PASS. | No newly orphaned helper found; final WB-04 review also retains the Note association receipt subscription. No alias, persisted state, HTTP or database migration. Roll back the three removal files together with any callers. Next: WB-04 final guides, full scoped gates and acceptance handoff. |
| WB-04 validation coverage, 2026-10-08 | Final acceptance review added Return cases for captured query/layout restoration against a changed authorized saved view, failed/canceled restoration retention and explicit base fallback. Existing admission row selector extended. Added Commands guide and completed navigation/adjacent guides. | Admission slice PASS `20261008T234938Z-p4540`; typecheck PASS `20261008T234939Z-p4857`; source-guide Markdown PASS `20261008T235053Z-p13247`. | This closes a coverage gap without adding another runtime policy or changing adopted owners. Final scoped gates remain in progress. |
| WB-04 generated accounting, 2026-10-08 | Catalog title additions invalidated the generated topology input digest. `agent-finalize` first stopped at JSON shape validation; public generation refreshed only `tools/execution_topology_render_index.json`. | Failed finalization `20261008T234445Z-p93574`; isolated JSON-shape failure `20261008T234512Z-p94136`. `make generate` PASS `20261008T234633Z-p95231`; finalization PASS `20261008T234648Z-p98420`. After Return selector additions, generation PASS `20261008T235003Z-p5758` and finalization PASS `20261008T235028Z-p8841`, before broad verification. | Change-related generated staleness repaired through Make; no hand-edited generated output. `RESULTS_DIR` unset, so retained-run canonical/performance maintenance was intentionally skipped. |
| WB-04 broad-gate repair, 2026-10-08 | Full workbook gate found two pre-existing reference-picker test rows still asserting `option`/`listbox` for the established checkbox/fieldset presentation. Verified the stale assertions at HEAD and unchanged production picker; corrected four lookups in coordination/contextual authoring tests. | Full workbook run `20261008T235115Z-p15723`: 322/324 units passed, failures `contextual_task_decision_authoring` and `coordination_create_authoring`; diagnostic files under matching `unit-logs/row-web.workbook.regression.*`. Architecture PASS `20261008T235115Z-p15726`; final browser PASS `20261008T235220Z-p38297` (15/15 units). | Unrelated baseline assertion drift, not navigation behavior. WB-04 BLOCKED pending narrow repair verification and complete workbook rerun; no production picker or coverage removal. |
| WB-04 retained observation review | Caller review confirmed generic inspector `latestRow` still consumes Note association receipts. Retained the source-owner subscription, without the removed fallback's `associationSnapshot` variable/authority lookup; it must not rely on query refresh to rerender its retained subject. | Coordination repair slice PASS `20261008T235524Z-p21707`; typecheck PASS `20261008T235553Z-p22848`; Biome PASS `20261008T235554Z-p23149`; finalizer PASS `20261008T235552Z-p22567`. | This is F-07 retention, not an alternate navigation path. Source frozen after this review; final whole-slice results must follow it. |
| WB-04 broad-gate recovery | The repaired workbook gate passed 324/324 units at `20261008T235715Z-p30310`; architecture passed 13/13 at `20261008T235716Z-p30585`. Because the retained-observation review finished during this run, a final stable-source run follows. | Finalizer PASS `20261008T235854Z-p61542` with no updated files and retained-run checks skipped; source-frozen typecheck PASS `20261008T235952Z-p82860`, Biome PASS `20261008T235953Z-p83473`, production build PASS `20261008T235955Z-p84536`. | Required failed exits are repaired. WB-04 returns to IN_PROGRESS pending final evidence recording and documentation handoff. |
| WB-04 complete, 2026-10-08 local / 2026-10-09 UTC | Completed navigation/Commands and adjacent guides, source/test ownership, final acceptance assessment, compatibility/removal ledger and rollback/restart handoff. Findings F-01–F-06 and V-01 resolved; F-07/F-08 retained with explicit rationale. Original visual-refresh record remains byte-identical from `Status:` onward. | Final W/A/B runs PASS at `20261009T000100Z-p87681` (324/324), `20261009T000100Z-p87684` (13/13), `20261009T000100Z-p87685` (15/15); type/import/Biome/catalog/build/finalizer gates PASS as §13. Final handoff Markdown PASS `20261009T000523Z-p95561`; whitespace PASS. | No required check or workstream remains deferred. No migration or feature flag. Broader freshness work stays out of scope; retained-run maintenance skipped only because `RESULTS_DIR` was unset. Changes remain uncommitted in the shared checkout for review. |

## 1. Scope and source posture

The primary target is the workbook navigation and composition seam under
`apps/web/src/workbook/navigation`, with the adjacent Commands boundary under
`apps/web/src/workbook/commands`. Target label: `workbook-workbench`.
The controlling plan and execution artifact is
`docs/handoffs/ui-ux/workbook-workbench-visual-refresh.md`.
The audience is the implementing engineer and reviewer. The current task
authorizes the scoped source, tests, authored accounting and documentation changes.

Inspection baseline: `main`, commit
`c7716b3dc25332b637b90ec2fd6b3915db2c94ed`. The original planning inspection
started clean; execution inherited the tracker edits recorded above. All eleven files in the two target directories were
inventoried. Adjacent owners were inspected only at the integration points
listed below; this is not a whole-workbook audit.

The goal is one understandable, independently testable navigation admission
boundary, with obsolete alternate execution paths and dead forwarding removed.
Future surfaces should supply semantic targets and existing owner capabilities
without making navigation depend on an expanding shell-runtime object.
No new package, generic workflow engine, navigation data store, persistent pins,
route, schema, feature family, or visual redesign is proposed.

Behavior follows adopted subsystem owners and Core 00–04. This plan uses
[Core 01 §3.3.4.3](../../spec/01_architecture_storage_and_view_contracts.md),
[Core 03 §2.5](../../spec/03_workbook_interaction_collaboration_and_workflows.md),
[Core 04 §2.2](../../spec/04_security_deployment_and_conformance.md), and bounded
[Design §§7, 8.3A, 8.5, 12.7, 14](../../design.md).
[Domain §§1, 6.2, 8–9](../../domain.md) supplies vocabulary and owner navigation;
[NLSpec research](../../research/nlspec-spec.md) supplies planning guidance on
interfaces, explicit boundaries, and testable acceptance, not new runtime rules.
Instructions embedded in reference documents do not expand the scoped remediation.

The [refactor framework](../cartulary_modular_refactor_planning_framework.md)
and [UI/UX overlay](../../cartulary-ui-ux-refactor-digest/cartulary/START_HERE.md)
guide the tracker structure and selection rubric. Existing code, tests, and the
[implementation handoff](workbook-workbench-implementation.md) establish current
implementation evidence only. No owner contradiction was identified in the
inspected clauses. Any later contradiction blocks its dependent slice rather
than permitting a silent choice of behavior.

## 2. Baseline repository inventory

This table preserves the pre-remediation inventory and rationale. The execution
ledger and final path inventory describe the delivered state. Paths in this first
table are relative to `apps/web/src/workbook/`.
All inventoried source belongs to `web.workbook`; verification routing is separately owned
by the active catalogs. None of these eleven files is a generated artifact.

| Path | Current responsibility | Exported/public symbols or package surface | Inbound callers | Outbound dependencies | Tests touching it | Generated artifacts or contracts touched | Suspected target owner module | Risk level | Notes |
| --- | --- | --- | --- | --- | --- | --- | --- | --- | --- |
| `navigation/WorkbookWorkbenchContext.tsx` | Workbench context and semantic action interface | `WorkbookWorkbench`, navigation/inspect types, context and consumer hook | Shell provider; Work, view controls, inspector and reference presentations | React, `SheetRef`, session metadata types | Navigation feedback and hook tests; shell consumers | Semantic sheet/record/field identities; no generated edit | `web.workbook` navigation | High | Nullable context currently permits alternate execution in two adjacent compositions. |
| `navigation/useWorkbookWorkbench.ts` | Destination preparation, admission, Return, pins and inspector handoff | Composition hook | `WorkbookShell` | Full shell-runtime type, query/locator ports, browsing registry, saved-view owner, session | Colocated admission test | Existing query/layout and locator contracts | `web.workbook` navigation | High | Replace broad runtime dependency; retain staged reads and committed presentation handshake. |
| `navigation/useWorkbookWorkbench.test.tsx` | Admission and asynchronous attachment characterization | Test cases; private fixture | `web.workbook.regression.workbench_navigation_admission` | Production hook, registry, typed read ports, grid doubles | This file | Semantic test rows; no generated contract change | `web.workbook` verification | High | Fixture currently casts a partial runtime through `unknown`; replace with the narrow interface. |
| `navigation/WorkbookSessionNavigation.ts` | Memory-local pins, bounded Return trail and navigation attempt lifetime | Session class, pin/origin types, `workbookPinIdentity` | Mutation runtime, navigation hook and Work presentation | `SheetRef`, query/layout types, generated design facade | Colocated session test and hook/status tests | Consumes existing pin/trail limits | `web.workbook` navigation | High | Keep owner-required origin metadata; remove only the redundant stored pending controller reference. |
| `navigation/WorkbookSessionNavigation.test.ts` | Capacity, duplicate, coalescing, trail and completion cases | Test cases | `web.workbook.regression.session_navigation` | Session owner | This file | Existing bounds and attempt semantics | `web.workbook` verification | Medium | Preserve admitted-versus-presented distinctions and cancellation evidence. |
| `navigation/WorkbookWorkPanel.tsx` | Work destination, Return control and navigation feedback | `WorkbookWorkControls`, `WorkbookReturnControl`, `WorkbookNavigationStatus` | Shell and footer/top-bar composition | Context, session, Commands, auxiliary dock, menu placement, surface IDs | Navigation status test and shell/browser coverage | Existing selector and design facades | `web.workbook` navigation presentation | Medium | Retain current controls and placement; file size alone does not justify splitting it. |
| `navigation/WorkbookNavigationStatus.test.tsx` | Feedback control identity, keyboard access and concealment | Test case | `web.workbook.regression.navigation_feedback` | Context, session and status presentation | This file | Existing navigation selector/state attributes | `web.workbook` verification | Medium | Preserve stable control footprint and current-attempt feedback. |
| `commands/workbookCommandIndex.ts` | Pure deterministic metadata search and descriptor types | Search function and command types | Commands provider/control and owner contributions | Generated design presentation facade | Colocated search test | Existing family ordering; no generated edit | `web.workbook` Commands | Low | Keep metadata matching separate from execution/authorization. |
| `commands/workbookCommandIndex.test.ts` | NFC, Unicode whitespace, ranking and identity ties | Test cases | `web.workbook.regression.command_search` | Pure command search | This file | Current matching contract | `web.workbook` verification | Low | Retain; no speculative search replacement. |
| `commands/WorkbookCommands.tsx` | Private registration index, owner bindings, target capture and search presentation | Provider, contribution hook and control | Shell; Work, Find, grid/view controls, Timeline tags/presets, inspector and Recovery contributions | Semantic Grid Adapter facade, descriptor search, menu placement | Colocated interaction test and responsive browser row | Existing command page size and semantic grid interface | `web.workbook` Commands | High | Existing private index is justified; do not turn it into a registry of source workflows. |
| `commands/WorkbookCommands.test.tsx` | Target revalidation, paging, borrowed focus and authority concealment | Test cases | `web.workbook.regression.workbench_commands_interaction` | Production Commands components and semantic grid double | This file | Existing owner eligibility and command limits | `web.workbook` verification | Medium | Preserve behavior while adjacent navigation capabilities become explicit. |

Adjacent inspection is bounded as follows. Paths are repository-relative;
inspection of a call site does not claim a complete audit of its owner.

| Adjacent files or inputs | Inspected integration point | Planned disposition |
| --- | --- | --- |
| `apps/web/src/workbook/WorkbookShell.tsx` | Workbench construction, current freshness predicate, provider placement, Work/Return/Commands composition | Adapt navigation dependency wiring; retain source-owner freshness contributions and their existing scope. |
| `apps/web/src/workbook/hooks/useWorkbookShellRuntime.ts`; `useWorkbookSavedViewController.ts` in the same directory | Returned command surfaces, startup `upsertSavedView`, internal selection callback and forwarded `selectSavedView` | Remove only the unconsumed forwarded command; preserve startup and owner callbacks. |
| `apps/web/src/workbook/features/generic/useGenericWorkbookInspectorComposition.tsx` | `navigateNote`, `navigatedNote`, navigation error/identity state, retained subject candidates | Remove the providerless navigation branch and state used solely by it. Other inspector state remains owner-held. |
| `apps/web/src/workbook/components/SavedViewBrowser.tsx`; `ActiveSurfaceSavedViewSelector.tsx`; `WorkbookShellViewBarControls.tsx`; `WorkbookViewBar.tsx` | Providerless `onBase`/`activateResource` selection and forwarding | Migrate navigation to the required capability; remove obsolete callback plumbing from these bindings. |
| `apps/web/src/workbook/components/GenericWorkbookSurface.tsx`; `apps/web/src/workbook/surfaces/WorkbookSurfacesFacade.tsx` | Production caller path to generic inspector composition | Carry the explicit capability from the workbench composition; do not redesign surface rendering. |
| `apps/web/src/workbook/savedviews/WorkbookSavedViewController.ts` | Resource observation/acceptance, `activateResource`, `openConfirmed` | Retain controller activation for confirmed owner operations and other live callers. |
| `apps/web/src/workbook/adapters/readWorkbookAuthoringRecord.ts` and its inbound references | Identity verification by paged authoring reads | Retain shared helper and live creation/association/Evidence/composition uses; retire only its navigation fallback use. |
| `apps/web/src/workbook/runtime/WorkbookMutationRuntime.ts` | Session construction and invalidation; source-version contributions | Retain runtime lifetime wiring; do not move session ownership or generalize version admission in this iteration. |
| `apps/web/src/workbook/query/WorkbookRecordLocatorPort.ts`; `apps/web/src/workbook/adapters/createWorkbookRecordLocatorAdapter.ts` | Typed outcomes, read scope, cancellation and locator admission | Preserve public transport and owner boundaries; adapter inspected at admission integration, not comprehensively audited. |
| `apps/web/src/workbook/components/ActiveSurfaceSavedViewSelector.test.tsx`; `apps/web/src/workbook/preferences/workbookPreferenceCharacterization.test.tsx`; `apps/web/src/workbook/features/notes/noteAssociations.test.tsx` | Standalone fixture composition and existing navigation assertions | Supply explicit test capabilities; add production-composition coverage rather than preserving a second runtime for fixtures. |
| `apps/web/src/workbook/components/WorkbookRelationshipChip.tsx`; `apps/web/src/workbook/features/generic/GenericInspectorReferenceSummary.tsx`; inspector presentation consumers | Semantic target construction and context consumers | Preserve admitted reference navigation; optional read-only presentation is not a second execution path. |
| Source READMEs, `apps/web/package.json`, `tools/frontend_source_ownership.json`, `tools/frontend_import_boundaries.json`, `tools/generated_artifact_policy.json`, and selected `tools/test_families/{web.workbook,module.workbook}.json` entries | Private application package, exact file ownership, import policy, generated roots and verification selectors | Update authored maps only during implementation when affected; execution updates the host ownership entry and affected test selectors. |

Remaining Workbook feature implementations, backend locator/provider internals,
SQL, storage, vendor integration, and the visual image contents are outside this
inspection. Execution additionally inspected and exercised the selected browser scenarios;
the Note scenario now asserts completion and inspector focus. Historical visual evidence is preserved, not re-certified.

## 3. Module boundary diagnosis

The target is application navigation orchestration plus presentation, not a new
domain module. Its common decision is whether a semantic destination can replace
the current authorized context and when its selection/focus is actually ready.

| Responsibility found | Current location | Correct owner candidate | Keep / move / split / defer | Evidence | Notes |
| --- | --- | --- | --- | --- | --- |
| Destination preparation and admission | Workbench hook | Navigation-owned interface and orchestrator | keep | Query/locator staging and current-attempt checks | Replace the dependency surface, not the admission algorithm. |
| Shell-runtime adaptation | Navigation imports the composition hook's return type | Workbook composition | move | `Runtime = ReturnType<typeof useWorkbookShellRuntime>` | Composition supplies only consumed capabilities; later shell growth does not enlarge navigation authority. |
| Note and saved-view user navigation | Workbench plus providerless local branches | Workbench navigation | move | Two distinct code paths selected by context presence | Cut over consumers together; no indefinite aliases or hidden fallback. |
| Query pages, committed selection and focus readiness | Browsing registry, query owners and Grid Adapter | Existing owners | keep | Staged page and selection-before-focus tests | Do not copy rows or checkpoints into a navigation store. |
| Session metadata and retirement | Session class, wired through mutation runtime | Existing incident/account runtime lifetime | keep | Runtime invalidation and session tests | Keep read authority separate from authoring permission. |
| Command discovery and owner dispatch | Private index and bound contributions | Commands plus source action owners | keep | Current registration/search and invocation checks | New actions add a descriptor and declared owner binding, not central feature-specific execution. |
| Cross-owner freshness consolidation | Shell predicate and source-specific read policies | Separate later analysis | defer | Policies have different current contributor sets | A generic version registry is not needed to retire these paths safely. |

## 4. Public contract and behavior freeze map

| Contract | Current owner | Evidence | Existing tests | Required characterization tests | Refactor risk | Notes |
| --- | --- | --- | --- | --- | --- | --- |
| Locator HTTP, rows and normal cursor continuation | Core 01 REQ-01-680–684; Workbook/query and source providers | Locator port/adapter and governing clauses | Existing locator catalog coverage; workbench admission row | Unsupported location must fail safely without page scanning; preserve operational/unavailable distinction | High | No route, envelope, cursor or backend change. |
| Navigation, selection and Return | Core 03 REQ-03-314–315 | Workbench hook, session and registry integration tests | Admission/session rows | Note/saved-view cutover through production composition, same-view reselection, delayed attachment and cancellation | High | Preserve origin until admission; completion requires current presentation, not just a read result. |
| Pins and Commands | Core 03 REQ-03-312–313 | Session, command search and contribution code | Search, interaction and session rows | Preserve duplicate-before-capacity, ordering, captured targets and invocation-time eligibility | Medium | Twenty pins, thirty-two Return origins and twenty command results remain existing projected limits. |
| Query/layout, saved views and presets | Core 03 REQ-03-315–316 and existing saved-view owners | Query/layout conversion and saved-view observation calls | Selector and workbench admission rows | Preserve captured working configuration, fresh resource comparison, explicit base opening and source-owner confirmed-operation activation | High | Removing browser fallback does not remove owner activation, startup admission or automatic current-resource recovery. |
| Authorization and lifetime | Core 04 REQ-04-169–170; existing mutation/query owners | Readability, actor change and runtime invalidation paths | Session, admission and command concealment cases | Same-account recovery, account replacement, incident exit/loss and closed-but-readable state | High | No permission cache, new auth service or persisted session metadata. |
| Source writes, drafts, receipts, history and collaboration | Core 03 REQ-03-311,317–318 and respective source owners | Detachment boundaries and retained operation composition | Existing Note association replay and shell/Commands tests | Preserve drafts/receipts through navigation; refresh retry must not resubmit accepted writes | High | No mutation, revision, WebSocket, storage, or replay policy changes. |
| Shell, inspector, selectors and accessible feedback | Design §§7,8.3A,8.5,12.7,14; Core 03 §2.5 | Work/Return/status presentation and inspector handshake | Navigation feedback and responsive Commands rows | Stable controls, keyboard return, no focus theft or lost deliberate interaction | Medium | Preserve graphite tokens, dock geometry, semantic selectors and goldens. |
| Machine projections and verification accounting | Authored generated policy, source/import manifests and test families | Inspected policy/catalog inputs | Architecture and routed owner checks | Update changed test selectors and file ownership without reducing coverage | Medium | No executable dependency on Markdown; historical passes do not certify this iteration. |

## 5. Coupling and boundary findings

| Finding | Baseline evidence | Risk | Classification | Owner | Execution disposition |
| --- | --- | --- | --- | --- | --- |
| F-01: broad composition dependency | Workbench hook imports full shell-runtime return type; its fixture requires an `unknown` cast | Shell expansion can unnecessarily widen navigation dependencies and obscure test obligations | must_fix | Navigation interface; composition adapter | RESOLVED in WB-01: explicit host and complete typed fixture. |
| F-02: alternate Note navigation and retained row | `navigateNote` switches to authoring page lookup without a provider; `navigatedNote` becomes another subject candidate | Parallel admission, failure and stale-result behavior; unnecessary compatibility burden | must_fix | Workbench navigation | RESOLVED in WB-02: required semantic action, no retained navigation row/error/scan branch. |
| F-03: alternate saved-view activation | `SavedViewBrowser` falls back to `onBase` or controller activation | Test fixtures can exercise a path unlike the application and omit navigation semantics | must_fix | Workbench navigation; saved-view owner remains independent | RESOLVED in WB-02: required base/saved navigation; no owner-activation fallback. |
| F-04: unused forwarded selection command | `selectSavedView` is returned by the saved-view hook and forwarded by shell runtime without a consuming production call site | Exposes an obsolete entry point that future code could adopt | should_fix | Shell composition | RESOLVED in WB-03: returned/forwarded wrapper removed; internal owner callback retained. |
| F-05: redundant stored abort reference | Pending request stores `controller`; execution closes over the local controller and cancellation uses `intent` | Extra apparent state owner makes cancellation harder to explain | should_fix | Session navigation | RESOLVED in WB-03: stored member removed; local controller, intent and completion retained. |
| F-06: absent source navigation guides | Both populated target directories lack README files and parent links | Future changes lack a clear source entry point despite machine ownership | should_fix | Workbook source guides | RESOLVED in WB-04: navigation/Commands guides, parent links and adjacent lifetime descriptions; source-guide Markdown lint passed. |
| F-07: live helpers and required metadata can resemble legacy code | Authoring reader has live source-owner callers; controller `openConfirmed` uses activation; Core 03 REQ-03-315 requires semantic invoker metadata | Broad deletion could remove behavior outside this seam | intentional/no_action | Existing owners | RETAINED: live callers reviewed in WB-03; invoker metadata remains required by REQ-03-315. |
| F-08: growth does not justify another framework | Existing query, Grid Adapter, Commands and source-operation boundaries already exist | New generic stores/registries would duplicate authority and complicate expansion | intentional/no_action | Existing owners | RETAINED boundaries: no new workflow/store/registry or persistence; explicit host only. |
| V-01: incomplete integration evidence | Callback doubles did not prove Note/saved-view composition, full authority lifetimes or inspector completion | Wiring and focus defects could remain hidden | must_fix | Existing shell/navigation tests and authored verification catalog | WB-02 added composition/browser cases and corrected inspector completion; WB-04 added saved-view Return restoration/failure/cancellation/fallback evidence. Final gates below. |

No direct storage or grid-vendor implementation dependency was found in the
eleven-file target. This does not establish a repository-wide boundary audit.
The previous implementation handoff deliberately retained standalone Note lookup;
this iteration changes that compatibility decision for the private application,
not the adopted navigation contract. The framework's package vocabulary does
not require promoting these app-local files into a shared package.

## 6. Refactor workstreams

| Workflow ID | Name | Class: root/chain/parallel | Required previous workflows | Required subsequent workflows | Goal | Files likely involved | Validation | Handoff checkpoint |
| --- | --- | --- | --- | --- | --- | --- | --- | --- |
| WB-W0 | Revalidate baseline and freeze coverage | root | User implementation request | WB-W1 | Confirm callers, owner clauses, test rows and clean/dirty state at implementation HEAD | Inventory and active catalogs | Read-only discovery; current task guides; selected baseline rows | Record drift and characterization gaps before source edits. |
| WB-W1 | Narrow navigation dependencies | chain | WB-W0 | WB-W2 | Establish explicit consumed capabilities | Navigation hook/types, shell composition and fixture | V-U, V-T, V-A below | WB-01 exit and changed interface inventory. |
| WB-W2 | Cut over legacy consumers | chain | WB-W1 | WB-W3 | One navigation path for Note and saved-view actions | Generic inspector, selector/browser, view-bar binding, tests | V-U, V-B below | WB-02 migrated callers and removal evidence. |
| WB-W3 | Remove dead surface | chain | WB-W2 | WB-W4 | Remove forwarding and redundant stored state | Shell hooks, session class and affected types | V-U, V-T, V-A; reference searches | WB-03 zero remaining consumers/aliases. |
| WB-W4 | Close ownership and readiness evidence | chain | WB-W3 | None | Complete guides, accounting, checks and handoff | Source guides, affected authored ownership/catalog entries, this tracker | V-D and final scoped implementation gates | WB-04 evidence, residual risks and rollback recorded. |

Update each completed workstream and its exit before beginning the dependent
workstream. Keep documentation and catalog adjustments needed by a slice with
that slice; WB-W4 verifies their final completeness rather than allowing stale
ownership between checkpoints.

## 7. Implementation slices

The user authorized all slices, including retirement of the unsupported standalone
paths. Execution status and evidence appear in the checkpoint table above. WB-01
and WB-03 narrow structural surfaces; WB-02 also corrects the inspector-completion
defect exposed by composition coverage. WB-04 closes documentation, verification
and handoff accounting. Adopted owners required no amendment.

| Slice ID | Depends on | Intended change | Files/packages likely involved | Contract risks | Tests to add or preserve | Validation command | Rollback note | Completion criterion |
| --- | --- | --- | --- | --- | --- | --- | --- | --- |
| WB-01 | WB-W0 | Declare navigation-owned host interface; wire consumed shell capabilities; replace broad fixture cast | `navigation/useWorkbookWorkbench.ts`, new app-local navigation interface, shell composition, hook tests | Accidental stale snapshot or missing saved-view/readiness capability | Existing admission, cancellation, stale-version, focus and authority cases | V-U, V-T, V-A | Revert interface, adapter and fixture together | No dependency on shell-hook return type; typed fixture constructs exactly the required capability. |
| WB-02 | WB-01 | Require navigation for Note and saved-view action compositions; delete alternate execution/state; migrate fixtures | Generic inspector, browser/selector, view-bar forwarding and affected tests | Lost Return, selection, authoring or recovery continuity | Add production-composition Note/saved-view cases; preserve incoming Note direction, saved-view owner operations and delayed-navigation interaction | V-U, V-B | Revert caller cutover, removals and tests as one slice | Both actions have one admission path; provider absence cannot trigger scans or alternate activation. |
| WB-03 | WB-02 | Delete unused forwarded command, redundant pending controller member and newly orphaned imports/props/exports | Shell hooks, session metadata owner, affected bindings | Deleting live internal selection or cancellation | Preserve startup, confirmed saved-view operation and session completion tests | V-U, V-T, V-A; `rg` caller checks | Revert deletions and type/fixture changes together | No retired API alias or test-only compatibility path remains; live operations still have their owner entry. |
| WB-04 | WB-03 | Finish source guides, authored accounting, acceptance and restart notes | Navigation/Commands READMEs, parent guide, affected manifests/catalogs, this handoff | Mistaking historical or visual evidence for readiness | Preserve test coverage and semantic identifiers; assess applicable digest acceptance rows with fresh evidence | V-D and final scoped implementation checks | Revert metadata/docs with associated source slice | All exits and relevant checks recorded; remaining debt is explicit, with no unsupported production-readiness claim. |

The navigation host interface exposes only the current semantic surface/sheet
identity, active contract, selected saved-view observation and grid-entry state;
query/layout read and application commands; identity/surface/extension selection;
grid-entry cancellation; and saved-view retain/read/status/accept operations.
Keep query, locator, page-admission, extension availability and authority-failure
inputs explicit as they are today. Declare the interface from owner types, not
`ReturnType` or `Pick` over the shell hook. Composition supplies the adapter;
navigation owns no new source operations or resource cache.

For the two action compositions, use a required narrow navigation capability
carried from the workbench provider through their production parents. Unit
fixtures supply typed capability doubles; integration cases exercise the real
navigation composition. Do not make every passive inspector/reference renderer
require the full workbench context merely to render saved data. Remove only the
base-selection plumbing that serves the deleted browser fallback; preserve
owner-declared selected-resource deletion/unavailability recovery.

The future extension path is an existing registered surface supplying a semantic
destination plus its owner capabilities. Command contributions remain descriptor
plus owner binding. No speculative surface or action is implemented to demonstrate
that path. Pins, Return, source reading, presets, Commands and Recovery are
retained because they provide adopted navigation, context, reading and recovery
capabilities; none is retained solely because code already exists.

No persisted data migration, HTTP compatibility period, feature flag or dual
runtime is needed for this private application cutover. Retain stable selectors
and generated contract facades. Roll back complete slices with their callers,
tests and ownership changes; do not restore only one side of an interface.

## 8. Validation plan

The rows below define execution coverage. Actual commands, results and run roots
are recorded in the execution log and final gate table; historical passes are
not substituted for changed-source verification.
Current discovery used `make help`, task guides for `web.workbook`,
`module.workbook` and `web.architecture`, and target explanations for
`lint-markdown` and `agent-finalize`.

| Validation layer | Command | Scope | Required before implementation? | Notes |
| --- | --- | --- | --- | --- |
| V-U: focused frontend behavior | `make test-slice OWNER=web.workbook ROWS=<exact selected rows below>` | Admission, session, Commands, feedback, Note associations and selector behavior | Yes, establish affected baseline | Repeat affected rows after each slice; retain any unrelated failures separately. |
| V-T: type surface | `make frontend-typecheck` | Narrow interface, removed command/prop consumers and typed fixtures | No; after source changes | No unsafe broad runtime cast or compatibility alias should be needed. |
| V-A: ownership/imports | `make test-slice OWNER=web.architecture`; `make frontend-import-boundary-check` | Authored ownership, acyclic direct imports and vendor isolation | No; after affected changes | New files require ownership and source-guide updates in their slice. |
| V-B: browser/accessibility | `make test-slice OWNER=module.workbook ROWS=<exact selected rows below>` | Real Note navigation, interrupted saved-view navigation, responsive Commands | No; after cutover | Catalog selects browser services and topology. No direct Playwright invocation. |
| Backend integration | Existing `module.workbook` locator catalog coverage | Locator contracts | No | No backend edit planned; broaden only if transport/provider behavior changes or a concrete regression points there. |
| Generated drift/accounting | Repository-owned generator and drift targets selected from affected authored input | Changed source/test accounting only | No | Do not hand-edit generated outputs. Do not run regeneration solely because this Markdown changed. |
| V-D: documentation | `make lint-markdown`; `git diff --check` | Controlling tracker and changed source guides | No | Documentation validation is separate from product evidence. |
| End-of-run maintenance | `make agent-finalize` | Required maintenance before broader end-of-run checks | No | Use `RESULTS_DIR` only for eligible successful full warm-check evidence; otherwise report retained-run maintenance skipped. |
| Broader product checks | `make test-slice OWNER=web.workbook`; affected browser/visual targets from current guides | Broaden for interface reach or unresolved risk after focused checks | No | Full scoped workbook/architecture checks are required; release-wide and broad visual checks are conditional on affected boundaries. |

Use these exact active frontend row IDs. The final selector row covers both
resource operations that remain and browser presentation whose fixture changes.

```text
web.workbook.regression.workbench_navigation_admission
web.workbook.regression.session_navigation
web.workbook.regression.navigation_feedback
web.workbook.regression.command_search
web.workbook.regression.workbench_commands_interaction
web.workbook.regression.note_associations
web.workbook.regression.activesurfacesavedviewselector_resource_action_focus_a103000001
```

The original seven-row focused invocation is:

```bash
make test-slice OWNER=web.workbook ROWS=web.workbook.regression.workbench_navigation_admission,web.workbook.regression.session_navigation,web.workbook.regression.navigation_feedback,web.workbook.regression.command_search,web.workbook.regression.workbench_commands_interaction,web.workbook.regression.note_associations,web.workbook.regression.activesurfacesavedviewselector_resource_action_focus_a103000001
```

The selected browser/accessibility invocation is:

```bash
make test-slice OWNER=module.workbook ROWS=module.workbook.browser.note_associations,module.workbook.browser_support.pending_navigation_preserves_interaction,module.workbook.accessibility.workbench_commands
```

Characterization cases were added to their existing owning test files/rows and
authored selectors updated alongside them. A test's historical row membership is not
proof that it covers the new case. The required scenario matrix is:

| Scenario group | Required outcome | Existing posture and required addition |
| --- | --- | --- |
| Note and saved-view cutover | One owner-admitted destination; same selected saved view observes current authorized persisted state | Existing standalone selector/association coverage remains; add real workbench composition cases so doubles cannot conceal lost navigation. |
| Read outcomes | Preserve origin for pending, failed, outside-query, unavailable, unsupported, canceled and superseded location; explicit base opening is a new bounded request | Existing admission cases cover several outcomes; add explicit unsupported/no-scan and cutover assertions. |
| Selection and presentation | Accepted page precedes owner selection; selection commit precedes eligible focus; inspector completion and delayed mounting remain fenced | Preserve existing admission/registry cases and deliberate Columns-click browser scenario. |
| Session and authority | Duplicate check precedes twenty-pin capacity; thirty-two-entry Return semantics; uncertainty conceals, replacement/exit retires, closed incident remains readable | Preserve session/command tests; cover full lifetime transitions in production composition where current unit evidence is insufficient. |
| Commands and retained work | Deterministic twenty-result paging; captured targets revalidated; borrowed focus restored only when eligible; draft/receipt retention through detachment | Preserve command search/interaction, responsive browser and Note replay coverage; no newly admitted write or replay path. |

Goldens, token inputs and renderer settings remain unchanged by default. If a
later authorized correction changes pixels, follow the
[visual golden maintenance guide](../../guides/cartulary_visual_golden_maintenance.md)
and record its exact cause separately; do not refresh images to hide a regression.
Failures must name the target, summary artifact/run root, relationship to this
slice, and dependent exit they block. A passing documentation check cannot
satisfy a product acceptance row.

## 9. Top-level work tracker

| ID | Work item | Workstream | Status | Depends on | Evidence or artifact | Exit condition |
| --- | --- | --- | --- | --- | --- | --- |
| WB-P0 | Inspect and record current plan | Documentation task | DONE | None | Baseline, inventory, owner mapping and retirement decisions above | Decision-complete document; no production edit. |
| WB-P1 | Revalidate implementation HEAD and characterize gaps | WB-W0 | DONE | User implementation request | Execution baseline above | Drift and failing baseline classified before source movement. |
| WB-P2 | Explicit navigation interface | WB-W1 | DONE | WB-P1 | WB-01 execution checkpoint | Narrow interface and affected checks pass. |
| WB-P3 | Legacy consumer cutover | WB-W2 | DONE | WB-P2 | WB-02 caller/removal ledger and tests | Single navigation path and continuity evidence. |
| WB-P4 | Dead-surface removal | WB-W3 | DONE | WB-P3 | WB-03 reference checks and live-owner tests | No unconsumed forwarding or retired compatibility alias. |
| WB-P5 | Ownership and readiness handoff | WB-W4 | DONE | WB-P4 | WB-04 guides, maps, runs and acceptance assessment | All implementation exits and limitations recorded. |
| WB-P6 | Consolidate broader freshness policies | Outside iteration | DEFERRED | Separate owner analysis | Different current contributor sets; no unification proposed | Not required for this iteration. |

## 10. Session handoff log

These planning-session entries are historical. Current implementation and validation
are recorded in the execution session log above; document-only restrictions below
do not apply to the authorized implementation.

### Scope and authority

| Time | Agent/session | Current state | Files inspected or touched | Commands run | Result | Blockers | Next action |
| --- | --- | --- | --- | --- | --- | --- | --- |
| 2026-10-08 | Planning and document update | Scope fixed to workbench seam; implementation remains TODO | AGENTS, cited Core/Design/Domain/research/framework/overlay, existing handoffs; only this handoff edited | `git status --short`, `git branch --show-current`, `git rev-parse HEAD`; targeted reads | Baseline is clean `main` at the recorded commit; reference documents did not expand task scope | None for document update | Implement WB-W0 only in later authorized production task. |

### Backend boundary

| Time | Agent/session | Current state | Files inspected or touched | Commands run | Result | Blockers | Next action |
| --- | --- | --- | --- | --- | --- | --- | --- |
| 2026-10-08 | Planning | Backend implementation excluded | Core locator/security clauses and client locator port/adapter integration | Targeted reads and reference searches | No backend, SQL, transport contract or migration change planned | None identified in inspected interface | Preserve backend boundary; investigate only if a scoped regression implicates it. |

### Frontend boundary

| Time | Agent/session | Current state | Files inspected or touched | Commands run | Result | Blockers | Next action |
| --- | --- | --- | --- | --- | --- | --- | --- |
| 2026-10-08 | Planning | Eleven target files inventoried; adjacent source inspected as bounded above | Navigation/Commands and listed integrations; source unchanged | `rg --files`; symbol/import searches; exact source reads | Alternate navigation paths and dead forwarding distinguished from live owner helpers | No production pass claimed | Narrow host interface before caller cutover. |

### Contract and codegen

| Time | Agent/session | Current state | Files inspected or touched | Commands run | Result | Blockers | Next action |
| --- | --- | --- | --- | --- | --- | --- | --- |
| 2026-10-08 | Planning | Existing contracts frozen | Source/import/generated policies and selected catalogs read | Targeted reads | No generated or machine input changed; no Markdown-dependent product check added | None | Update authored ownership/routing with later affected source; generate only downstream outputs required by those changes. |

### Tests and harness

| Time | Agent/session | Current state | Files inspected or touched | Commands run | Result | Blockers | Next action |
| --- | --- | --- | --- | --- | --- | --- | --- |
| 2026-10-08 | Planning | Canonical routing discovered | Active frontend and browser rows; target guidance | `make help`; `make task-guide ROLE=module-author OWNER=web.workbook`; same guide for `module.workbook` and `web.architecture`; `make explain-target TARGET=lint-markdown DETAIL=summary`; same explanation for `agent-finalize` | Discovery commands passed; proposed checks in §8 | None for planning | Run current guides again if implementation HEAD changes. |
| 2026-10-08 | Document update | Documentation verification passed | This handoff only | `make lint-markdown`; `git diff --check`; `git status --short`; manual byte comparison with `git show HEAD:<path>` | Markdown PASS at `.cartulary/test-results/20261008T225814Z-p61694`, summary `adhoc/lint-markdown/tool-run-summary.json`; whitespace PASS; one-file additions-only diff; original status-through-end content is byte-identical | None for document update; no product test attempted | Re-run lint after this evidence entry; leave production slices TODO. |

In that historical planning session, product suites, generation, visual comparison
and release checks were skipped because it changed documentation only. `make agent-finalize` was not run:
there is no broader product verification or requested harness maintenance in this
task. Retained-run maintenance was skipped because `RESULTS_DIR` was unset.
Historical run roots below remain historical evidence, not the current baseline.

### Security and authorization

| Time | Agent/session | Current state | Files inspected or touched | Commands run | Result | Blockers | Next action |
| --- | --- | --- | --- | --- | --- | --- | --- |
| 2026-10-08 | Planning | Read authority and operation lifetime remain separate | Core 04 §2.2; hook/session/runtime invalidation and relevant tests | Targeted reads and actor/readability searches | No new ACL, account clearing policy or persistence proposed | No contradiction identified in inspected scope | Preserve concealment/retirement and test late responses in cutover. |

### Open risks and next session

| Time | Agent/session | Current state | Files inspected or touched | Commands run | Result | Blockers | Next action |
| --- | --- | --- | --- | --- | --- | --- | --- |
| 2026-10-08 | Planning | Next production step is WB-W0, followed by WB-01 | Tracker and scoped source inventory | `git status --short` and reference searches | No code executed from the proposed slices; standalone test migration remains planned work | Later production task required | Restart with `make task-guide ROLE=module-author OWNER=web.workbook`; revalidate callers and baseline before editing source. |

## 11. Open questions and blockers

No owner contradiction or unresolved product preference remains. The original
implementation-authorization blocker was resolved by the user's explicit request.
WB-04 is complete: final stable-source gates pass and the handoff is recorded.

| ID | Question or blocker | Resolution / required evidence | Current status |
| --- | --- | --- | --- |
| RB-001 | Historical document-only scope | Explicit implementation request authorizes WB-W0 through WB-04 and path retirement. | RESOLVED |
| RB-002 | Fresh cutover and full lifetime evidence | WB-01–03 focused exits and all WB-04 final gates pass; acceptance complete in §13. | RESOLVED |

Broader freshness-policy consolidation is deferred, not a completion blocker:
source contributor sets have distinct owners and were not unified here. Any
future supported consumer must establish its owner and concrete contract before
introducing another entry point.

## 12. Binary completion criteria

Implementation completion requires the following:

- WB-W0 through WB-04 and WB-P1 through WB-P5 are DONE, with tracker updates
  recorded between workstreams.
- Findings F-01–F-06 and V-01 have tested remedies; F-07/F-08 retention is explicit.
- Required scoped gates pass; failed intermediate runs are classified and linked
  to their repair and successful replacement evidence.
- Source guides, ownership/routing, removals, migration impact, rollback and
  restart guidance describe the final state.
- Historical visual-refresh evidence and heading anchors remain intact; tokens,
  geometry, selectors and goldens remain unchanged.

The implementation is complete only when WB-P1–WB-P5 are DONE, all slice
exits pass, retirement searches find no obsolete aliases/callers, and the handoff
contains fresh relevant verification evidence. Assess applicable
[digest acceptance rows](../../cartulary-ui-ux-refactor-digest/cartulary/acceptance.tsv)
in that handoff with PASS evidence or justified N/A; an applicable BLOCKED row
prevents completion. No new status schema or Markdown dependency belongs in
product checks. Local readiness does not imply whole-application accessibility,
performance publication, deployment or release qualification.

## 13. Implementation handoff

### Delivered boundary and paths

Short workbook subpaths below are relative to `apps/web/src/workbook/`; full
paths identify other repository roots. The original eleven-file inventory remains
above for traceability. Unlisted source owners were read at
integration points only and were not claimed as fully audited.

| Area | Delivered paths / change |
| --- | --- |
| Navigation boundary | New `apps/web/src/workbook/navigation/WorkbookNavigationHost.ts`; explicit host use in `useWorkbookWorkbench.ts`; action capability in `WorkbookWorkbenchContext.tsx`; simplified `WorkbookSessionNavigation.ts`. |
| Shell adaptation and owner surfaces | `apps/web/src/workbook/WorkbookShell.tsx`; `hooks/useWorkbookShellRuntime.ts`; `hooks/useWorkbookSavedViewController.ts`. Host capabilities are adapted explicitly; the returned selection wrapper is gone. |
| Note cutover | `apps/web/src/workbook/surfaces/WorkbookSurfacesFacade.tsx`; `components/GenericWorkbookSurface.tsx`; `features/generic/useGenericWorkbookInspectorComposition.tsx`; `features/notes/NoteAssociationPanel.tsx`. Required semantic navigation replaces the providerless reader and extra retained row/error. |
| Saved-view cutover | `apps/web/src/workbook/components/WorkbookShellViewBarControls.tsx`, `ActiveSurfaceSavedViewSelector.tsx`, `SavedViewBrowser.tsx`. Required navigation covers base/saved selection; optional context supplies presentation only. |
| Completion correction | `apps/web/src/workbook/query/WorkbookQueryBrowsingContext.tsx` plus the navigation hook. Mounted accepted destination and grid accessibility are separate; an inspector's completed focus handoff can finish while the inspector remains open. |
| Evidence | `apps/web/src/workbook/WorkbookShell.surfaces.test.tsx`, `navigation/useWorkbookWorkbench.test.tsx`, `components/ActiveSurfaceSavedViewSelector.test.tsx`, `preferences/workbookPreferenceCharacterization.test.tsx`; `apps/web/e2e/note-associations.spec.ts`. Existing cases retained; seven routed cases added. |
| Baseline test repair | `apps/web/src/workbook/components/WorkbookAuthoringReferencePicker.test.tsx`: one unsupported test-ID lookup replaced by the existing Parties accessible group. Four additional stale role lookups corrected in `features/coordination/{contextualCreateAuthoring,coordinationCreateAuthoring}.test.tsx`. All verified pre-existing at HEAD; no product or selector contract change. |
| Accounting | Authored `tools/frontend_source_ownership.json` and `tools/test_families/web.workbook.json`; generated `tools/execution_topology_render_index.json` refreshed by Make. No other generated artifact changed. |
| Guides | New `apps/web/src/workbook/{navigation,commands}/README.md`; parent workbook guide and adjacent `hooks`, `query`, `savedviews`, `features/generic`, `features/notes` guides; this tracker. |

### Compatibility and removal ledger

| Decision | Owner / consequence |
| --- | --- |
| Remove broad shell runtime import/cast | Navigation has a complete host fixture and explicit composition adapter; no shell return-type dependency. |
| Remove providerless Note lookup | No scan, `navigatedNote`, navigation-local errors or exclusive fencing remain. The source owner still reads records for authoring and recovery. |
| Remove alternate saved/base browser activation | Required `open` capability replaces `onSelectBaseSurface`/`onBase` browser forwarding and direct browser `activateResource`; fixture owner deletion/unavailability callbacks remain intentional. |
| Remove returned `selectSavedView` | Only the internal hook callback bound to the saved-view owner remains. No alias or deprecated wrapper. |
| Remove stored pending controller | The local controller, active abort intent, request identity and presentation outcomes retain their distinct roles. |
| Retain live operations | Shared authoring reader still serves creation, associations, Evidence and composition. Saved-view `openConfirmed` still uses owner `activateResource`; startup accepts authorized resources; current-resource deletion/unavailability remains owner recovery. |
| Retain useful adopted capabilities | Commands, pins, bounded Return with invoker metadata, semantic source reading, query owners and Grid Adapter remain independent. No registry/store/engine/freshness consolidation introduced. |

This is an atomic private-application TypeScript cutover. No HTTP/API or database
migration, persisted-data conversion, feature flag, compatibility alias or dual
implementation is required. No Core/Design/Domain, typed behavioral contract,
selector, token, geometry or visual golden input changed. Tests and runtime do
not read, stat or hash this tracker or other Markdown.

### Rollback, limitations and restart

Rollback a complete slice with its callers, fixtures and authored accounting:
WB-01 host/hook/shell adaptation together; WB-02 consumer cutover and the query/
inspector completion correction together; WB-03 wrapper and stored-state removals
together. Reconcile source guides and regenerate catalog-derived topology through
Make after reverting authored inputs. Avoid a blanket checkout reset: the tracker
included user-owned planning edits before execution. No production data rollback
is necessary. Rerun the affected slice and selected browser cases after rollback.

No required scoped check may be deferred at completion. Backend, database, release,
full repository check and broad visual/measurement suites are outside this
frontend structural boundary. They are not implied by a workbook pass. Browser
screenshots are implementation-support artifacts; no new pixel comparison or
whole-application accessibility qualification is claimed. Retained-run canonical
and performance maintenance was skipped because `RESULTS_DIR` was unset and no
eligible successful full warm-check evidence was selected.

Residual out-of-scope debt is the previously deferred cross-owner freshness
analysis. The actual selected contributors remain unchanged. There is no need
to recreate retired compatibility paths when extending the workbook: supply a
semantic destination through the navigation action and keep source operations
with their owners.

Restart by checking branch/dirty state against the recorded baseline and this
handoff, then use `make task-guide ROLE=module-author OWNER=web.workbook` (or the
owning boundary's guide) for the new change. Reuse the exact selected browser
invocation in §8 when navigation wiring/focus changes. There is no remaining
implementation slice once WB-04 and its final evidence table are DONE.

### Final scoped validation

All IDs below are under `.cartulary/test-results/`. For graph targets, the summary
is `target-summaries/<target>.json` inside the run root. Unit counts include setup;
they are not counts of test assertions. These are local scoped readiness results,
not repository-wide release qualification.

| Key | Command | Result / evidence |
| --- | --- | --- |
| W | `make test-slice OWNER=web.workbook` | PASS, 324/324 units, `20261009T000100Z-p87681/target-summaries/test-slice.json`. Includes the original seven rows, preferences, saved-view controller, shell composition, new Return cases and retained owner operations. |
| A | `make test-slice OWNER=web.architecture` | PASS, 13/13 units, `20261009T000100Z-p87684/target-summaries/test-slice.json`. |
| B | Three-row `module.workbook` invocation in §8 | PASS, 15/15 units, `20261009T000100Z-p87685/target-summaries/test-slice.json`. Note direction/removal, exact replay/read recovery, inspector completion/focus, Columns click with late attachment fencing, responsive Commands/axe. |
| T | `make frontend-typecheck` | PASS, `20261008T235952Z-p82860/target-summaries/frontend-typecheck.json`. |
| I | `make frontend-import-boundary-check` | PASS, `20261009T000100Z-p87883/target-summaries/frontend-import-boundary-check.json`. |
| L | `make lint-biome` | PASS, `20261008T235953Z-p83473/target-summaries/lint-biome.json`. |
| C | `make test-catalog-check` | PASS, `20261009T000100Z-p88347/test-catalog-check/tool-run-summary.json`; seven new exact case titles in existing authored rows. |
| Build | `make build-web` | PASS, `20261008T235955Z-p84536/target-summaries/build-web.json`. |
| G | `make generate`; `make agent-finalize` | PASS; generator `20261008T235003Z-p5758/generate/tool-run-summary.json`; finalizer `20261008T235854Z-p61542/unit-artifacts/finalize-summary.json`. Generated catalog input digest current; finalizer changed no files. |
| D | `make lint-markdown`; `git diff --check` | PASS, `20261009T000523Z-p95561/adhoc/lint-markdown/tool-run-summary.json`; `git diff --check` also PASS. Completion checkpoint updates retain the verified document structure. |
| Review | Baseline/retirement/source diff review | HEAD remains `c7716b3dc25332b637b90ec2fd6b3915db2c94ed`; historical visual record from `Status:` through end is byte-identical to HEAD. Host has one authored ownership entry; no retired navigation branch/alias remains. |

Production source was frozen before these final W/A/B/T/I/L/Build results.
Only handoff documentation changed afterward. No final product gate failed or
was deferred. Intermediate failures, their relation to the work and successful
replacement evidence remain in the execution log above.

### Digest acceptance assessment

This assesses the remediation's affected behavior only. `N/A` denotes an
unchanged owner boundary, not unverified work silently treated as complete.
Evidence keys refer to the final gate table; source/diff review is human evidence,
not a new executable dependency on the digest.

| ID | Status | Scoped evidence or owner rationale |
| --- | --- | --- |
| A001 | PASS | Owner-to-gap map in §§1–5: Core 01 locator reads, Core 03 navigation/Commands/Return and Core 04 disclosure/lifetimes; source owner and verification routing distinguished. No normative contradiction or new requirement. |
| A002 | PASS | One semantic admission decision, explicit host/actions, complete caller cutover and removal/retention ledger. No speculative package or framework. W/A/I. |
| A003 | PASS | WB-W0 revalidated baseline and inherited dirty tracker; actual authored source/import/generated policies and active catalogs inspected. New host ownership registered; A/I/G. Digest maps were navigation, not current authority. |
| A004 | PASS | Diff introduces no component-local design literal, token, theme or density registry; UI package/token files unchanged. |
| A005 | N/A | Theme behavior is unchanged; current dark-graphite ownership remains with existing generated presentation. No new palette or theme control. |
| A006 | N/A | Density selection and computed geometry were not edited. No density/measurement requalification claimed. |
| A007 | PASS | W covers ordinary/contextual/Note authoring and source retention; B preserves related-Note creation/association/recovery. Navigation introduces no write capability or creation policy. |
| A008 | N/A | Responsive branching, viewport fallback and inspector clamp geometry are unchanged. B still verifies required scoped Commands visibility at 1440×900, 390×480 and 1280×900 with 2× zoom; broader layout matrix is not requalified. |
| A009 | N/A | Scroll ownership, status/account navigation layout and overflow styles are unchanged. No shell geometry rewrite. |
| A010 | PASS | W/B prove admitted selection, inspector attachment/completion and cancellation fencing; existing dispatcher/authority/draft cases pass. Retained association receipt subscription kept independently of navigation. |
| A011 | PASS | W/B cover Note Return, saved-view reselection, compatible captured query/layout restoration, failed/canceled Return retention, semantic selection/focus and source-owned draft/receipt continuity. |
| A012 | PASS | No transaction-generation/transport changes. W owner cases and B exact-replay scenario retain captured identity/bytes; shell navigation cases assert no unintended writes. |
| A013 | PASS | W retained-operation/recovery cases and B acknowledgement with failed read recovery pass. Owner recovery remains read-only after acceptance and separate from presentation. |
| A014 | PASS | W editing/authoring/navigation cases and B late-interaction delivery pass. No new draft store or persistence; source-owned raw authoring remains retained. |
| A015 | N/A | Conflict presentation and saved/local value separation are untouched; no toast or cell-conflict behavior changed. |
| A016 | PASS | W authority/composition cases preserve separate read admission and interaction permission, including closed but readable incidents; unsupported/malformed/operational/outside-query/unavailable outcomes remain distinct. |
| A017 | PASS | W hook cancellation and shell lifetime cases prove uncertainty concealment, same-account restoration, account replacement and incident exit/access-loss retirement, including late read fencing. |
| A018 | N/A | Evidence lifecycle/overlay/preview state machinery and non-color encoding are unchanged. Note association operations retain their existing Evidence owner. |
| A019 | PASS | B responsive Commands keyboard/axe checks and Note inspector focus, W borrowed-focus and navigation status tests, plus semantic accessible selector repairs. No whole-application conformance claim. |
| A020 | PASS | Affected navigation-status and Commands states pass W/B, including narrow width and zoom. No new component variant or compound-state styling; unchanged long-content geometry is not requalified. |
| A021 | PASS | W staged page/selection/focus tests and B real Grid Adapter navigation/late-interaction cases preserve semantic row identity. Virtualizer implementation and performance/measurement boundaries are untouched. |
| A022 | N/A | No visual fixture, renderer setting or golden changed. Historical visual evidence is preserved, not presented as a fresh pixel comparison; no intentional pixel correction was made. |
| A023 | PASS | Existing UI-contract helpers and semantic identities/accessible names used; A selector-ownership check and T pass. No new selector schema or raw unowned test ID. |
| A024 | PASS | Diff/source review and A/C: no test, runtime, generator or product evidence dependency on Markdown was introduced. Documentation lint is a separate maintenance check. |
| A025 | PASS | Only authored source ownership/test selectors changed; G regenerated the downstream input digest. No generated file was hand-edited; no unrelated generated drift repaired. |
| A026 | PASS | Private atomic cutover and no-data-migration result explicit. Internal selection, startup, authoring readers, confirmed-operation opening and invoker metadata have live callers/adopted obligations; obsolete compatibility paths removed. |
| A027 | PASS | All scoped product/documentation gates pass. WB-W0–WB-04 checkpoints and sequencing, complete path/removal/retention ledger, failed-run repairs, compatibility, skipped checks, rollback and restart guidance are recorded. |

## Completed visual refresh record

The status and all remaining sections below are the preserved historical record.
They describe the completed visual refresh, not the remediation iteration above.

Status: all changed images reviewed; two final ordinary comparisons passed. See the
[implementation handoff](workbook-workbench-implementation.md).
This is implementation-support evidence, not behavioral or release authority.

## Accepted change and evidence

The adopted Design 0.6 and Core 03 workbench requirements intentionally change
the incident/view/footer bands, auxiliary destination, inspector reading surface,
and command/query menus. Existing administration and directory consumers of the
shared top-bar minimum-height token also change: the reviewed audit header/content
origin moves from 64 to 58 pixels without losing controls or data. The corrective
refresh keeps the closed lifecycle label within the 40-pixel incident band at
390 pixels. No renderer, font, viewport, zoom, masking, scroll normalization,
screenshot scope, or comparison tolerance was changed for this refresh.

The initial ordinary baseline is
`.cartulary/test-results/20261007T035002Z-p13696`. Its reconciliation reported
no missing or ambiguously mapped goldens; expected pixel comparisons failed.
The initial transactional update passed 12/12 units at
`.cartulary/test-results/20261007T044457Z-p76167`. The corrective transactional
update passed 12/12 units at
`.cartulary/test-results/20261007T050710Z-p2162`.

The final reconciliation accounts for 255 active captures/goldens and 29 registered
fixtures with zero missing goldens, ambiguous mappings, or orphans. Of these,
232 goldens changed relative to the starting commit. Each final changed image was
visually consumed through a private artifact session after byte-length and SHA-256
verification. Review included the single corrected lifecycle image after the
second update. No unexplained visual difference remains. Standalone image imports
provide pixels; they do not provide DOM, accessibility, or performance evidence.
All review sessions were stopped with successful cleanup receipts.

The final `tools/frontend_visual_golden_manifest.json` SHA-256 is
`233a2d274e7cb3abc16035d761f401c404c0292ad1f0415a688418f96133b854`.
Renderer identity remains
`visual.renderer.playwright_1_63_0_chromium_1243_linux_amd64`.
The two final ordinary comparisons both passed 12/12 units against this exact
manifest at `.cartulary/test-results/20261007T051745Z-p41032` and
`.cartulary/test-results/20261007T051745Z-p41039`. Their reconciliations account for
the same 255 active goldens with no missing, ambiguous, or orphan entries. The
implementation handoff records broader functional/accessibility evidence.

## Exact changed-golden ownership

The following mappings were read from the final reconciliation's exact
`golden_path`, `catalog_row_ids`, `owner_ids`, and `fixture_ids` fields, after
matching the changed paths and SHA-256 values. Ownership was not inferred from
filenames. Every filename below is relative to
`apps/web/e2e/workbook.visual.spec.ts-snapshots/`. “Active nonregistry capture”
means an admitted capture without a registry fixture, not an orphan.

### `module.collaboration.visual.capture_presence_markers_same_field_conflict_res_f0a62c52a1`

Owner: `module.collaboration`.

| Changed golden filename | Stable fixture ID |
| --- | --- |
| `collaboration-conflict-resolver-linux.png` | `visual.fixture.same_field_conflict` |
| `collaboration-conflict-strip-linux.png` | `visual.fixture.save_state_strip` |
| `collaboration-presence-markers-linux.png` | `visual.fixture.presence_overflow` |
| `collaboration-recovered-saved-strip-linux.png` | `visual.fixture.save_state_strip` |

### `module.collaboration.visual.the_visual_harness_asserts_deterministic_timelin_22b64f5dec`

Owner: `module.collaboration`.

| Changed golden filename | Stable fixture ID |
| --- | --- |
| `collaboration-grid-presence-markers-linux.png` | Active nonregistry capture |

### `module.collaboration.visual.the_visual_harness_asserts_same_field_conflict_m_c472bd3f9c`

Owner: `module.collaboration`.

| Changed golden filename | Stable fixture ID |
| --- | --- |
| `collaboration-grid-conflict-resolver-compact-linux.png` | Active nonregistry capture |
| `collaboration-grid-conflict-resolver-linux.png` | Active nonregistry capture |
| `collaboration-grid-conflict-resolver-narrow-linux.png` | Active nonregistry capture |

### `module.collaboration.visual.the_visual_harness_asserts_syncing_same_field_co_df11cd99bc`

Owner: `module.collaboration`.

| Changed golden filename | Stable fixture ID |
| --- | --- |
| `collaboration-grid-blocked-conflict-linux.png` | Active nonregistry capture |
| `collaboration-grid-recovered-saved-strip-linux.png` | Active nonregistry capture |
| `collaboration-grid-saved-strip-linux.png` | Active nonregistry capture |
| `collaboration-grid-syncing-strip-linux.png` | `visual.fixture.save_state_strip` |

### `module.entities.visual.capture_unresolved_token_resolved_chip_auto_reso_d3b74bd9d7`

Owner: `module.entities`.

| Changed golden filename | Stable fixture ID |
| --- | --- |
| `entity-mention-chip-states-linux.png` | `visual.fixture.mention_chip_state_matrix` |

### `module.entities.visual.the_visual_harness_captures_unresolved_mention_a_4b882068c7`

Owner: `module.entities`.

| Changed golden filename | Stable fixture ID |
| --- | --- |
| `record-relationships-mention-chips-linux.png` | Active nonregistry capture |

### `module.evidence.visual.capture_evidence_count_affordance_available_requ_cfada809e4`

Owner: `module.evidence`.

| Changed golden filename | Stable fixture ID |
| --- | --- |
| `evidence-affordance-states-linux.png` | `visual.fixture.evidence_affordance` |
| `evidence-timeline-evidence-count-linux.png` | Active nonregistry capture |

### `module.evidence.visual.the_visual_harness_captures_blocked_evidence_acc_779473e830`

Owner: `module.evidence`.

| Changed golden filename | Stable fixture ID |
| --- | --- |
| `evidence-grid-blocked-preview-linux.png` | Active nonregistry capture |
| `evidence-grid-timeline-evidence-badge-linux.png` | Active nonregistry capture |

### `module.evidence.visual.the_visual_harness_captures_evidence_surface_acc_8c22a3c9bc`

Owner: `module.evidence`.

| Changed golden filename | Stable fixture ID |
| --- | --- |
| `record-relationships-evidence-access-linux.png` | Active nonregistry capture |

### `module.evidence.visual.the_visual_harness_captures_requested_evidence_a_1eb50235af`

Owner: `module.evidence`.

| Changed golden filename | Stable fixture ID |
| --- | --- |
| `evidence-grid-available-evidence-linux.png` | Active nonregistry capture |
| `evidence-grid-requested-evidence-linux.png` | Active nonregistry capture |

### `module.networkflow.visual.capture_deterministic_claimed_network_analysis_a_47b1c2cce6`

Owner: `module.networkflow`.

| Changed golden filename | Stable fixture ID |
| --- | --- |
| `network-flow-analysis-accepted-inspector-linux.png` | `visual.fixture.claimed_network_analysis_workspace_states` |
| `network-flow-analysis-compact-saved-graphs-linux.png` | `visual.fixture.claimed_network_analysis_compact_workspace` |
| `network-flow-analysis-delete-dialog-linux.png` | `visual.fixture.claimed_network_analysis_workspace_states` |
| `network-flow-analysis-filtered-empty-grid-linux.png` | Active nonregistry capture |
| `network-flow-analysis-graph-contributors-linux.png` | `visual.fixture.claimed_network_analysis_workspace_states` |
| `network-flow-analysis-mapping-dialog-linux.png` | `visual.fixture.claimed_network_analysis_workspace_states` |
| `network-flow-analysis-narrow-query-controls-linux.png` | `visual.fixture.claimed_network_analysis_narrow_workspace` |
| `network-flow-analysis-rejected-diagnostics-linux.png` | `visual.fixture.claimed_network_analysis_workspace_states` |
| `network-flow-analysis-saved-graph-result-linux.png` | `visual.fixture.claimed_network_analysis_workspace_states` |

### `module.savedviews.visual.capture_saved_view_selector_active_chips_grouped_3da7859cdc`

Owner: `module.savedviews`.

| Changed golden filename | Stable fixture ID |
| --- | --- |
| `workbook-query-empty-closed-read-only-linux.png` | `visual.fixture.empty_successful_query` |
| `workbook-query-empty-compact-linux.png` | `visual.fixture.empty_successful_query` |
| `workbook-query-empty-density-comfortable-linux.png` | `visual.fixture.empty_successful_query` |
| `workbook-query-empty-density-compact-linux.png` | `visual.fixture.empty_successful_query` |
| `workbook-query-empty-narrow-linux.png` | `visual.fixture.empty_successful_query` |
| `workbook-query-empty-successful-query-linux.png` | `visual.fixture.empty_successful_query` |
| `workbook-query-empty-text-spacing-linux.png` | `visual.fixture.empty_successful_query` |
| `workbook-query-empty-zoom-200-linux.png` | `visual.fixture.empty_successful_query` |
| `workbook-query-filtered-empty-linux.png` | `visual.fixture.empty_successful_query` |
| `workbook-query-saved-view-query-controls-linux.png` | `visual.fixture.saved_view_query_controls_and_grouped_result` |
| `workbook-view-bar-filter-editing-overflow-linux.png` | Active nonregistry capture |
| `workbook-view-bar-long-columns-linux.png` | Active nonregistry capture |
| `workbook-view-bar-maximum-pressure-base-linux.png` | Active nonregistry capture |
| `workbook-view-bar-maximum-pressure-compact-linux.png` | Active nonregistry capture |
| `workbook-view-bar-maximum-pressure-narrow-linux.png` | Active nonregistry capture |
| `workbook-view-bar-ordered-maximum-sort-linux.png` | Active nonregistry capture |
| `workbook-view-bar-saved-view-actions-linux.png` | Active nonregistry capture |
| `workbook-view-bar-saved-view-clean-linux.png` | Active nonregistry capture |
| `workbook-view-bar-saved-view-modified-linux.png` | Active nonregistry capture |
| `workbook-view-bar-text-spacing-linux.png` | Active nonregistry capture |
| `workbook-view-bar-zoom-200-linux.png` | Active nonregistry capture |

### `module.timeline.visual.the_visual_harness_captures_a_deterministic_grou_ac01b2d810`

Owner: `module.timeline`.

| Changed golden filename | Stable fixture ID |
| --- | --- |
| `timeline-grid-grouped-grid-linux.png` | Active nonregistry capture |

### `module.timeline.visual.the_visual_harness_captures_a_deterministic_time_a19d57e206`

Owner: `module.timeline`.

| Changed golden filename | Stable fixture ID |
| --- | --- |
| `timeline-grid-timeline-default-linux.png` | Active nonregistry capture |

### `module.timeline.visual.the_visual_harness_drives_the_real_timeline_work_0977c1d4cf`

Owner: `module.timeline`.

| Changed golden filename | Stable fixture ID |
| --- | --- |
| `timeline-grid-active-edit-cell-linux.png` | `visual.fixture.edit_cell` |
| `timeline-grid-conflict-strip-linux.png` | Active nonregistry capture |
| `timeline-grid-saved-strip-linux.png` | Active nonregistry capture |
| `timeline-grid-syncing-strip-linux.png` | Active nonregistry capture |

### `module.workbook.visual.capture_default_timeline_workbook_shell_with_vie_c06bbcbee0`

Owner: `module.workbook`.

| Changed golden filename | Stable fixture ID |
| --- | --- |
| `incident-directory-compact-desktop-workbook-shell-linux.png` | `visual.fixture.compact_desktop_workbook_shell` |
| `incident-directory-default-timeline-workbook-shell-linux.png` | `visual.fixture.default_timeline_workbook_shell` |
| `incident-directory-narrow-desktop-workbook-shell-linux.png` | `visual.fixture.narrow_desktop_workbook_shell` |

### `module.workbook.visual.capture_inspector_details_relationships_evidence_a56cae74ea`

Owner: `module.workbook`.

| Changed golden filename | Stable fixture ID |
| --- | --- |
| `workbook-inspector-attached-edit-linux.png` | `visual.fixture.base_inspector` |
| `workbook-inspector-compact-actions-linux.png` | `visual.fixture.inspector_compact_actions` |
| `workbook-inspector-destructive-confirmation-linux.png` | `visual.fixture.destructive_actions` |
| `workbook-inspector-details-linux.png` | `visual.fixture.base_inspector` |
| `workbook-inspector-history-linux.png` | `visual.fixture.base_inspector` |
| `workbook-inspector-narrow-technical-details-linux.png` | `visual.fixture.inspector_narrow_technical_details` |
| `workbook-inspector-public-error-linux.png` | `visual.fixture.base_inspector` |
| `workbook-inspector-relationships-linux.png` | `visual.fixture.base_inspector` |
| `workbook-inspector-retained-draft-linux.png` | `visual.fixture.base_inspector` |
| `workbook-inspector-rollback-preview-linux.png` | `visual.fixture.base_inspector` |

### `module.workbook.visual.capture_save_state_pending_replay_transaction_re_70f3e80a67`

Owner: `module.workbook`.

| Changed golden filename | Stable fixture ID |
| --- | --- |
| `timeline-mutation-active-edit-cell-linux.png` | `visual.fixture.edit_cell` |
| `timeline-mutation-empty-timeline-query-linux.png` | Active nonregistry capture |
| `timeline-mutation-pending-replay-status-linux.png` | Active nonregistry capture |
| `timeline-mutation-transaction-recovery-panel-compact-linux.png` | Active nonregistry capture |
| `timeline-mutation-transaction-recovery-panel-linux.png` | Active nonregistry capture |
| `timeline-mutation-transaction-recovery-panel-narrow-linux.png` | Active nonregistry capture |

### `module.workbook.visual.capture_task_requests_or_decisions_parties_link_558c8596cc`

Owner: `module.workbook`.

| Changed golden filename | Stable fixture ID |
| --- | --- |
| `record-relationships-task-requests-linux.png` | `visual.fixture.task_requests_or_decisions` |

### `module.workbook.visual.contextual_task_decision_creation`

Owner: `module.workbook`.

| Changed golden filename | Stable fixture ID |
| --- | --- |
| `contextual-decision-authoring-linux.png` | Active nonregistry capture |
| `contextual-decision-authoring-narrow-linux.png` | Active nonregistry capture |
| `contextual-decision-recovery-linux.png` | Active nonregistry capture |
| `contextual-decision-recovery-narrow-linux.png` | Active nonregistry capture |
| `contextual-decision-references-narrow-linux.png` | Active nonregistry capture |
| `contextual-task-request-authoring-linux.png` | Active nonregistry capture |
| `contextual-task-request-authoring-narrow-linux.png` | Active nonregistry capture |
| `contextual-task-request-recovery-linux.png` | Active nonregistry capture |
| `contextual-task-request-recovery-narrow-linux.png` | Active nonregistry capture |
| `contextual-task-request-references-narrow-linux.png` | Active nonregistry capture |

### `module.workbook.visual.coordination_create_authoring_recovery`

Owner: `module.workbook`.

| Changed golden filename | Stable fixture ID |
| --- | --- |
| `coordination-comm-log-authoring-linux.png` | `visual.fixture.contextual_coordination_creation` |
| `coordination-handoff-authoring-linux.png` | `visual.fixture.contextual_coordination_creation` |
| `coordination-lesson-authoring-linux.png` | `visual.fixture.contextual_coordination_creation` |
| `coordination-recovery-linux.png` | `visual.fixture.contextual_coordination_creation` |
| `coordination-recovery-narrow-linux.png` | `visual.fixture.contextual_coordination_creation` |
| `coordination-source-narrow-linux.png` | `visual.fixture.contextual_coordination_creation` |
| `coordination-status-review-authoring-linux.png` | `visual.fixture.contextual_coordination_creation` |

### `module.workbook.visual.decision_supersession_review_recovery`

Owner: `module.workbook`.

| Changed golden filename | Stable fixture ID |
| --- | --- |
| `decision-supersession-accepted-linux.png` | Active nonregistry capture |
| `decision-supersession-review-linux.png` | Active nonregistry capture |
| `decision-supersession-review-narrow-linux.png` | Active nonregistry capture |

### `module.workbook.visual.indicator_lifecycle_authoring`

Owner: `module.workbook`.

| Changed golden filename | Stable fixture ID |
| --- | --- |
| `indicator-lifecycle-authoring-linux.png` | `visual.fixture.indicator_lifecycle_authoring` |
| `indicator-lifecycle-authoring-narrow-linux.png` | `visual.fixture.indicator_lifecycle_authoring` |

### `module.workbook.visual.indicator_observations_authoring`

Owner: `module.workbook`.

| Changed golden filename | Stable fixture ID |
| --- | --- |
| `indicator-observation-authoring-linux.png` | `visual.fixture.indicator_observations_authoring` |
| `indicator-observation-authoring-narrow-linux.png` | `visual.fixture.indicator_observations_authoring` |

### `module.workbook.visual.note_create_authoring_recovery`

Owner: `module.workbook`.

| Changed golden filename | Stable fixture ID |
| --- | --- |
| `linked-note-authoring-linux.png` | Active nonregistry capture |
| `linked-note-authoring-narrow-linux.png` | Active nonregistry capture |
| `linked-note-recovery-linux.png` | Active nonregistry capture |
| `linked-note-recovery-narrow-linux.png` | Active nonregistry capture |
| `linked-note-source-narrow-linux.png` | Active nonregistry capture |

### `module.workbook.visual.ordinary_create_authoring_recovery`

Owner: `module.workbook`.

| Changed golden filename | Stable fixture ID |
| --- | --- |
| `ordinary-closed-retained-narrow-linux.png` | Active nonregistry capture |
| `ordinary-recovery-1280-linux.png` | Active nonregistry capture |
| `ordinary-recovery-390-linux.png` | Active nonregistry capture |
| `ordinary-reference-authoring-linux.png` | Active nonregistry capture |

### `module.workbook.visual.preferences`

Owner: `module.workbook`.

| Changed golden filename | Stable fixture ID |
| --- | --- |
| `workbook-preferences-comfortable-linux.png` | Active nonregistry capture |
| `workbook-preferences-compact-linux.png` | Active nonregistry capture |
| `workbook-preferences-confirmed-stale-linux.png` | Active nonregistry capture |
| `workbook-preferences-uncertain-linux.png` | Active nonregistry capture |
| `workbook-preferences-uncertain-narrow-linux.png` | Active nonregistry capture |
| `workbook-preferences-unset-linux.png` | Active nonregistry capture |

### `module.workbook.visual.timeline_capture_actions`

Owner: `module.workbook`.

| Changed golden filename | Stable fixture ID |
| --- | --- |
| `timeline-supersession-accepted-linux.png` | Active nonregistry capture |
| `timeline-supersession-authoring-linux.png` | Active nonregistry capture |
| `timeline-supersession-review-linux.png` | Active nonregistry capture |
| `timeline-supersession-review-narrow-linux.png` | Active nonregistry capture |

### `module.workbook.visual.timeline_related_evidence`

Owner: `module.workbook`.

| Changed golden filename | Stable fixture ID |
| --- | --- |
| `timeline-related-evidence-authoring-linux.png` | Active nonregistry capture |
| `timeline-related-evidence-authoring-narrow-linux.png` | Active nonregistry capture |
| `timeline-related-evidence-partial-linux.png` | Active nonregistry capture |
| `timeline-related-evidence-partial-narrow-linux.png` | Active nonregistry capture |
| `timeline-related-evidence-party-narrow-linux.png` | Active nonregistry capture |

### `package.grid_adapter.visual.capture_test_only_grid_adapter_support_specimens_9c222633ba`

Owner: `package.grid_adapter`.

| Changed golden filename | Stable fixture ID |
| --- | --- |
| `timeline-grid-adapter-fixtures-linux.png` | `visual.fixture.drag_fill_handle`, `visual.fixture.edit_cell`, `visual.fixture.frozen_column`, `visual.fixture.resize_handle`, `visual.fixture.tree_group_row` |

### `web.design.visual.account_menu_root_and_nested_viewport_states_cef60727cc`

Owner: `web.design`.

| Changed golden filename | Stable fixture ID |
| --- | --- |
| `account-menu-controls-compact-linux.png` | Active nonregistry capture |
| `account-menu-controls-narrow-linux.png` | Active nonregistry capture |
| `account-menu-controls-short-linux.png` | Active nonregistry capture |
| `account-menu-deployment-root-linux.png` | Active nonregistry capture |
| `account-menu-directory-root-linux.png` | Active nonregistry capture |
| `account-menu-long-label-text-spacing-linux.png` | Active nonregistry capture |
| `account-menu-long-label-zoom-linux.png` | Active nonregistry capture |
| `account-menu-workbook-root-linux.png` | Active nonregistry capture |

### `web.design.visual.account_settings_editing_states_5af731dc29`

Owner: `web.design`.

| Changed golden filename | Stable fixture ID |
| --- | --- |
| `account-settings-appearance-conflict-linux.png` | Active nonregistry capture |
| `account-settings-appearance-dirty-linux.png` | Active nonregistry capture |
| `account-settings-appearance-pending-linux.png` | Active nonregistry capture |
| `account-settings-appearance-recovery-linux.png` | Active nonregistry capture |
| `account-settings-appearance-short-linux.png` | Active nonregistry capture |
| `account-settings-appearance-zoom-linux.png` | Active nonregistry capture |
| `account-settings-profile-conflict-linux.png` | Active nonregistry capture |
| `account-settings-profile-dirty-linux.png` | Active nonregistry capture |
| `account-settings-profile-long-name-linux.png` | Active nonregistry capture |
| `account-settings-profile-pending-linux.png` | Active nonregistry capture |
| `account-settings-profile-recovery-linux.png` | Active nonregistry capture |

### `web.design.visual.administrative_audit_browsing`

Owner: `web.design`.

| Changed golden filename | Stable fixture ID |
| --- | --- |
| `administrative-audit-empty-linux.png` | Active nonregistry capture |
| `administrative-audit-inspected-linux.png` | Active nonregistry capture |
| `administrative-audit-loading-linux.png` | Active nonregistry capture |
| `administrative-audit-page-two-linux.png` | Active nonregistry capture |
| `administrative-audit-stale-linux.png` | Active nonregistry capture |
| `administrative-audit-unavailable-linux.png` | Active nonregistry capture |

### `web.design.visual.capture_test_only_exposed_dark_graphite_token_an_7cc73db04c`

Owner: `web.design`.

| Changed golden filename | Stable fixture ID |
| --- | --- |
| `design-delayed-initial-loading-linux.png` | `visual.fixture.delayed_initial_loading` |
| `design-exposed-theme-states-linux.png` | `visual.fixture.component_state_matrix` |
| `design-grid-background-refresh-linux.png` | `visual.fixture.error_presentation_loci` |
| `design-grid-closed-read-only-rows-linux.png` | Active nonregistry capture |
| `design-grid-stale-refresh-linux.png` | `visual.fixture.error_presentation_loci` |
| `design-grid-unavailable-initial-load-linux.png` | `visual.fixture.error_presentation_loci` |
| `design-immediate-initial-loading-linux.png` | `visual.fixture.delayed_initial_loading` |

### `web.design.visual.incident_creation_form_errors_pending_recovery_a_0d7c2a3cde`

Owner: `web.design`.

| Changed golden filename | Stable fixture ID |
| --- | --- |
| `incident-create-confirmed-handoff-failure-linux.png` | Active nonregistry capture |
| `incident-create-expanded-details-linux.png` | Active nonregistry capture |
| `incident-create-initial-linux.png` | Active nonregistry capture |
| `incident-create-pending-linux.png` | Active nonregistry capture |
| `incident-create-recovery-short-linux.png` | Active nonregistry capture |
| `incident-create-recovery-zoom-linux.png` | Active nonregistry capture |
| `incident-create-required-errors-linux.png` | Active nonregistry capture |

### `web.design.visual.incident_import_workflow`

Owner: `web.design`.

| Changed golden filename | Stable fixture ID |
| --- | --- |
| `incident-import-access-retry-linux.png` | Active nonregistry capture |
| `incident-import-action-checking-linux.png` | Active nonregistry capture |
| `incident-import-action-unconfirmed-linux.png` | Active nonregistry capture |
| `incident-import-admission-pending-linux.png` | Active nonregistry capture |
| `incident-import-cancel-requested-linux.png` | Active nonregistry capture |
| `incident-import-canceled-linux.png` | Active nonregistry capture |
| `incident-import-empty-linux.png` | Active nonregistry capture |
| `incident-import-failed-linux.png` | Active nonregistry capture |
| `incident-import-handoff-unavailable-linux.png` | Active nonregistry capture |
| `incident-import-observation-unavailable-linux.png` | Active nonregistry capture |
| `incident-import-queued-indeterminate-linux.png` | Active nonregistry capture |
| `incident-import-required-linux.png` | Active nonregistry capture |
| `incident-import-running-determinate-linux.png` | Active nonregistry capture |
| `incident-import-succeeded-linux.png` | Active nonregistry capture |

### `web.design.visual.lifecycle`

Owner: `web.design`.

| Changed golden filename | Stable fixture ID |
| --- | --- |
| `lifecycle-closed-linux.png` | Active nonregistry capture |
| `lifecycle-comfortable-linux.png` | Active nonregistry capture |
| `lifecycle-compact-linux.png` | Active nonregistry capture |
| `lifecycle-confirmed-refresh-failure-linux.png` | Active nonregistry capture |
| `lifecycle-pending-linux.png` | Active nonregistry capture |
| `lifecycle-reason-linux.png` | Active nonregistry capture |
| `lifecycle-review-linux.png` | Active nonregistry capture |
| `lifecycle-review-narrow-linux.png` | Active nonregistry capture |
| `lifecycle-review-spacing-linux.png` | Active nonregistry capture |
| `lifecycle-review-zoom-linux.png` | Active nonregistry capture |
| `lifecycle-uncertain-linux.png` | Active nonregistry capture |

### `web.design.visual.membership_audit_browsing`

Owner: `web.design`.

| Changed golden filename | Stable fixture ID |
| --- | --- |
| `membership-audit-comfortable-linux.png` | Active nonregistry capture |
| `membership-audit-compact-linux.png` | Active nonregistry capture |
| `membership-audit-cursor-recovery-linux.png` | Active nonregistry capture |
| `membership-audit-empty-linux.png` | Active nonregistry capture |
| `membership-audit-inspected-linux.png` | Active nonregistry capture |
| `membership-audit-inspected-narrow-linux.png` | Active nonregistry capture |
| `membership-audit-inspected-spacing-linux.png` | Active nonregistry capture |
| `membership-audit-inspected-zoom-linux.png` | Active nonregistry capture |
| `membership-audit-loading-linux.png` | Active nonregistry capture |
| `membership-audit-stale-linux.png` | Active nonregistry capture |

### `web.design.visual.membership_management_visual`

Owner: `web.design`.

| Changed golden filename | Stable fixture ID |
| --- | --- |
| `membership-management-comfortable-linux.png` | Active nonregistry capture |
| `membership-management-compact-linux.png` | Active nonregistry capture |
| `membership-management-confirmed-refresh-failure-linux.png` | Active nonregistry capture |
| `membership-management-loading-linux.png` | Active nonregistry capture |
| `membership-management-pending-linux.png` | Active nonregistry capture |
| `membership-management-removal-narrow-linux.png` | Active nonregistry capture |
| `membership-management-removal-spacing-linux.png` | Active nonregistry capture |
| `membership-management-removal-zoom-linux.png` | Active nonregistry capture |
| `membership-management-role-linux.png` | Active nonregistry capture |
| `membership-management-uncertain-linux.png` | Active nonregistry capture |

### `web.design.visual.metadata_editing`

Owner: `web.design`.

| Changed golden filename | Stable fixture ID |
| --- | --- |
| `metadata-closed-linux.png` | Active nonregistry capture |
| `metadata-comfortable-linux.png` | Active nonregistry capture |
| `metadata-compact-linux.png` | Active nonregistry capture |
| `metadata-confirmed-refresh-failure-linux.png` | Active nonregistry capture |
| `metadata-conflict-linux.png` | Active nonregistry capture |
| `metadata-dirty-linux.png` | Active nonregistry capture |
| `metadata-loading-linux.png` | Active nonregistry capture |
| `metadata-review-narrow-linux.png` | Active nonregistry capture |
| `metadata-review-spacing-linux.png` | Active nonregistry capture |
| `metadata-review-zoom-linux.png` | Active nonregistry capture |
| `metadata-saving-linux.png` | Active nonregistry capture |
| `metadata-uncertain-linux.png` | Active nonregistry capture |

### `web.design.visual.reference_pack_administration`

Owner: `web.design`.

| Changed golden filename | Stable fixture ID |
| --- | --- |
| `reference-pack-cancel-requested-linux.png` | Active nonregistry capture |
| `reference-pack-canceled-linux.png` | Active nonregistry capture |
| `reference-pack-catalog-linux.png` | Active nonregistry capture |
| `reference-pack-failed-linux.png` | Active nonregistry capture |
| `reference-pack-observation-recovery-linux.png` | Active nonregistry capture |
| `reference-pack-queued-linux.png` | Active nonregistry capture |
| `reference-pack-running-linux.png` | Active nonregistry capture |
| `reference-pack-selection-linux.png` | Active nonregistry capture |
| `reference-pack-submitted-linux.png` | Active nonregistry capture |
| `reference-pack-succeeded-linux.png` | Active nonregistry capture |
