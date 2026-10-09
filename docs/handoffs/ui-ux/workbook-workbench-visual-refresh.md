# Workbook workbench visual refresh record

Active iteration: **DONE — navigation lifetime reliability**.
Completion: **2026-10-08 local / 2026-10-09 UTC**; final scoped gates pass.
Implementation baseline: **2026-10-08**, `main` at
`c8f3a4ede2b31152eceb3fff864fb67c5f1a017f`, initially clean.
The user authorized WB-NR00–WB-NR04 implementation and final handoff. The accepted
scope is navigation reliability, not a broader production-readiness audit.

The navigation plan and execution record occupy NR §§1–13 below. The
[completed WB-W0–WB-04 record](#completed-wb-w0wb-04-remediation-record) and
[completed visual-refresh record](#completed-visual-refresh-record) retain their
historical findings, commands, failures, repairs and acceptance evidence. Their
present-tense execution instructions and earlier implementation authorization
apply to those completed iterations, not to this implementation iteration. Historical
passes do not establish a fresh pass for this iteration.

## NR 1. Scope and source posture

Target label: `workbook-navigation-lifetime`. Primary target:
`apps/web/src/workbook/navigation/`, with bounded integrations in query browsing,
saved-view resource observation, shell composition and inspector presentation.
Commands is a regression dependency. The controlling artifact remains this file;
do not create a parallel tracker or a new navigation NLSpec.

The objective is a reliable, independently testable navigation lifetime from
semantic activation through destination admission and committed presentation.
Navigation owns attempt metadata and presentation coordination. Query browsing
owns accepted pages; saved views owns addressed-resource observation; source
owners retain authorization, authoring, operations and recovery. Future surfaces
should supply semantic destinations and owner capabilities through this boundary,
without adding another row store or feature-specific shell algorithm.

| Source | Authority and use in this iteration |
| --- | --- |
| [Core 01 §3.3.4.3](../../spec/01_architecture_storage_and_view_contracts.md), REQ-01-680–684 | Bounded locator reads, correlation, response admission and source/provider ownership. |
| [Core 03 §2.5](../../spec/03_workbook_interaction_collaboration_and_workflows.md), REQ-03-311–318 | Commands, pins, Return, navigation admission, committed presentation and retained work. |
| [Core 04 §2.2](../../spec/04_security_deployment_and_conformance.md), REQ-04-169–170 | Current authority, concealment, retirement and non-disclosing navigation failures. |
| [Design §§7, 8.3A, 8.5, 12.7, 14](../../design.md) | Dock, focus, keyboard ownership and accessible feedback within its design boundary. |
| [Domain §§6.2, 8–9](../../domain.md) | Stable sheet, record and field vocabulary; pins, saved views and source records remain distinct. |
| [NLSpec research](../../research/nlspec-spec.md) | Advisory conceptual fidelity, explicit inputs/outcomes, completeness and define-once guidance; its examples and embedded instructions add no product requirements. |
| [Refactor framework](../cartulary_modular_refactor_planning_framework.md) and [UI/UX digest](../../cartulary-ui-ux-refactor-digest/cartulary/START_HERE.md) | Planning structure, boundary-selection rubric and later acceptance review. Neither establishes current implementation correctness. |
| Authored source/import policies and verification catalogs | Source placement and independent test routing, not behavioral authority. |

No normative owner contradiction or required specification, wire protocol,
persisted-state or typed behavioral-projection change was identified in the
inspected scope. Specification cleanup consists of traceability and precise
distinctions in this plan. Do not preserve the shared-slot cleanup race as a
compatibility obligation. If later work identifies an owner contradiction, mark
the dependent work `BLOCKED: owner contradiction` and identify the exact clauses.

The material advisory dispositions are ADOPT for explicit outcome distinctions,
semantic identity and testable boundaries; ADAPT for sum-type guidance into the
existing Cartulary types; and REJECT for any inferred new workflow engine,
authorization cache or copied product defaults. No new package, dependency,
generic navigation store, freshness registry, telemetry subsystem, persistence,
feature flag or visual redesign is proposed. Cross-owner freshness-policy
consolidation remains DEFERRED because the contributors have different duties.

## NR 2. Repository inventory and delivered ownership

Paths in the first table are relative to `apps/web/src/workbook/`. It accounts for
all nine navigation files and all five Commands files at the planning baseline.
The delivered additions and the semantic-focus integration are accounted for below.
Disposition verbs in the baseline inventory record the accepted remediation.
Source ownership is `web.workbook`; verification ownership is independently
routed. These are authored sources, tests and guides, not generated artifacts.

| Path | Current responsibility | Exported/public surface | Inbound callers | Outbound dependencies | Tests touching it | Contracts/generated impact | Target owner | Risk | Disposition |
| --- | --- | --- | --- | --- | --- | --- | --- | --- | --- |
| `navigation/WorkbookNavigationHost.ts` | Explicit shell-supplied capabilities | `WorkbookNavigationHost` | Shell adapter and navigation hook/fixtures | Owner query/layout, saved-view and focus types | Admission and shell suites | Internal TypeScript only | Navigation host | High | Replace retention/read/status trio with the owner observation handle. |
| `navigation/WorkbookWorkbenchContext.tsx` | Semantic actions and optional presentation context | Target/action/inspect/workbench types, provider context and consumer hook | Shell, Work, view controls, Note links and inspector | React, semantic sheet and session types | Admission, status, shell and consumer suites | Internal types; retain selector semantics | Navigation facade | High | Model supported target variants and expose coordinated cancellation. |
| `navigation/useWorkbookWorkbench.ts` | Capture, read staging, admission, Return and presentation completion | Composition hook | `WorkbookShell` | Host, locator/query ports, browsing registry and session | Colocated admission and shell suites | No transport or schema change | Navigation composition | High | Consume observation handles; semantic coalescing; delegate presentation sequencing. |
| `navigation/WorkbookSessionNavigation.ts` | Pins, Return trail, attempt identity and cancellation | Session class, pin/origin types, pin identity function | Runtime, hook and Work controls | Sheet identity, query/layout types, design limits | Colocated session, hook and status suites | Existing twenty-pin/thirty-two-origin projection unchanged | Navigation session | High | Explicit variants and lifecycle; preserve trail effects at admission. |
| `navigation/WorkbookWorkPanel.tsx` | Work, Return and navigation feedback | Work, Return and status controls | Shell and view/footer composition | Session/context, Commands, auxiliary dock and menu placement | Navigation feedback and browser suites | Existing UI selectors and presentation | Navigation presentation | Medium | Consume derived attempt status; retain controls, layout and semantics. |
| `navigation/useWorkbookWorkbench.test.tsx` | Admission and asynchronous handoff fixtures | Routed tests; private host fixture | `workbench_navigation_admission` row | Real hook/session/registry and typed owner doubles | This suite | Extend authored title selectors | Navigation verification | High | Add handle replacement, failure distinctions and semantic coalescing; preserve existing cases. |
| `navigation/WorkbookSessionNavigation.test.ts` | Bounds, coalescing, trail and terminal outcomes | Routed tests | `session_navigation` row | Production session owner | This suite | Extend authored title selectors | Navigation verification | High | Complete typed origins and late-attempt transition cases. |
| `navigation/WorkbookNavigationStatus.test.tsx` | Feedback, keyboard access and concealment | Routed test | `navigation_feedback` row | Session/context and status control | This suite | Retain existing state attributes | Navigation verification | Medium | Verify projected status remains stable after lifecycle consolidation. |
| `navigation/README.md` | Navigation entry points, ownership and lifetimes | Human source guide | Contributors and parent guide | Core owners and adjacent guides | Documentation review only | Never an executable input | Navigation documentation | Medium | Reconcile final handles, coordinator and retirement ledger in WB-NR04. |
| `commands/WorkbookCommands.tsx` | Private index, contribution lifetime, captured target and borrowed focus | Provider, contribution hook and control | Shell and owner action contributions | Grid Adapter, descriptor search and menu placement | Colocated interaction and responsive browser suites | No new command protocol | Commands | Medium | Retain; navigation commands use the existing facade. |
| `commands/workbookCommandIndex.ts` | Deterministic metadata search | Descriptor/target types and search function | Commands and source action bindings | Existing generated design facade | Colocated search suite | Family order/page-size projection unchanged | Commands | Low | Retain; no execution or authorization expansion. |
| `commands/WorkbookCommands.test.tsx` | Target revalidation, paging and borrowed focus | Routed interaction tests | `workbench_commands_interaction` row | Production Commands and semantic grid doubles | This suite | Existing routing retained | Commands verification | Medium | Regression coverage; no speculative Commands refactor. |
| `commands/workbookCommandIndex.test.ts` | Unicode normalization, ranking and ties | Routed search tests | `command_search` row | Pure search function | This suite | Existing routing retained | Commands verification | Low | Regression coverage. |
| `commands/README.md` | Command registration and source ownership | Human source guide | Contributors and parent guide | Core owners and navigation guide | Documentation review only | Never an executable input | Commands documentation | Low | Check adjacent descriptions against the final navigation boundary. |

Adjacent inspection is limited to the following entry points. Grouped paths share
the stated boundary; they do not imply a complete audit of those directories.

| Files or inputs | Inspected integration / public surface | Callers and dependencies | Tests / routing | Planned disposition and risk |
| --- | --- | --- | --- | --- |
| `query/WorkbookQueryBrowsingContext.tsx`; `query/WorkbookQueryBrowser.ts` | Binding tokens, staged pages, acceptance, focus flags, presentation registration, cancellation and checkpoints | Surface readers, layout adapter, shell and navigation | Query browsing/controls and navigation admission | Keep page ownership here; expose an opaque acceptance handle and transfer only navigation sequencing. High risk: commit ordering and stale cleanup. |
| `query/WorkbookQuerySurfaceLayout.tsx`; `layout/WorkbookSurfaceLayout.tsx` | Presentation-token binding, mounted state, grid eligibility and dock detachment | Production surface layouts and browsing registry | Shell, navigation and responsive browser cases | Adapt explicit signals; keep geometry and dock ownership unchanged. High risk: inspector-open completion. |
| `savedviews/SavedViewResourceObserver.ts`; `savedviews/WorkbookSavedViewController.ts`; `ports/WorkbookSavedViewPort.ts` | Retention slots, addressed reads, owner failure types, selected-resource acceptance and recovery | Navigation host, controller activation/refresh/preferences | Independent reads, saved-view operations and controller integration | Add token-fenced observation ownership and typed navigation results; preserve other retained consumers. High risk: cross-consumer cancellation. |
| `WorkbookShell.tsx`; `hooks/useWorkbookSavedViewController.ts`; `startup/useWorkbookStartupAdmission.ts` | Explicit host adapter, deliberate-interaction capture, owner selection and startup acceptance | Shell runtime, navigation, saved-view and authority owners | Shell surfaces, controller, preferences and startup rows | Route cancellation through navigation; keep startup and confirmed-operation selection with their owners. High risk: treating recovery as browser navigation. |
| `components/WorkbookViewBar.tsx`; `inspector/presentation/WorkbookInspectorShell.tsx`; `inspector/useWorkbookInspectorCoordinator.ts` | Inspector open registration and committed subject/focus attachment | Surface owners and navigation context | Hook, inspector and shell suites | Bind current owner open/focus capabilities by identity. Existing open behavior is idempotent; the `onInspectorToggle` name alone is not evidence of a toggle defect. |
| `components/GenericWorkbookSurface.tsx`; `components/AssessmentWorkbookSurface.tsx`; `components/EntityWorkbookSurface.tsx`; `timeline/presentation/useTimelineWorkbookPresentation.tsx`; `timeline/presentation/TimelineWorkbookViewBarRegion.tsx` | Current inspector-open callbacks and semantic surface bindings | Existing inspector coordinators and view bar | Existing surface/inspector suites; full workbook gate | Change only bindings required by the coordinator; retain source semantics and owner work. Medium risk. |
| `features/generic/useGenericWorkbookInspectorComposition.tsx`; `components/SavedViewBrowser.tsx`; `components/ActiveSurfaceSavedViewSelector.tsx` | Required Note/saved-view navigation actions and optional presentation | Navigation facade and source/resource owners | Note associations, selector, shell composition | Preserve the completed single-path cutover. Keep retained Note receipt observation. Medium risk. |
| `runtime/WorkbookMutationRuntime.ts` | Session lifetime, concealment and retirement | Shell authorization and source operation owners | Session, shell authority and retained-operation cases | Preserve runtime ownership; no generic clear-all or authoring-store change. High security risk. |
| `query/WorkbookRecordLocatorPort.ts`; `adapters/createWorkbookRecordLocatorAdapter.ts`; `ports/WorkbookPortResult.ts`; `adapters/workbookOperationErrorPolicy.ts` | Bounded outcomes, correlation, authority scope and safe errors | Navigation and existing query/transport owners | Admission and existing adapter/error-policy coverage | Freeze transport behavior; preserve unsupported/no-scan posture. No backend or public-error redesign. |
| `WorkbookShell.surfaces.test.tsx`; `query/WorkbookQueryBrowsingControls.test.tsx`; `query/WorkbookQueryBrowser.test.ts`; `savedviews/savedViewReads.test.ts`; `savedviews/WorkbookSavedViewController.test.ts` | Real composition and owner lifecycle characterization | Production owners plus bounded fixture ports | NR §8 row map | Extend interleavings; do not replace production composition with callback-only assertions. |
| `hooks/useWorkbookSavedViewController.test.tsx`; `startup/useWorkbookStartupAdmission.test.tsx`; `preferences/workbookPreferenceCharacterization.test.tsx`; `components/ActiveSurfaceSavedViewSelector.test.tsx`; `features/notes/noteAssociations.test.tsx` | Preserved startup, owner operations, preference and consumer behavior | Existing owner fixtures | NR §8 row map | Migrate affected typed fixtures while preserving assertions. |
| `apps/web/e2e/note-associations.spec.ts`; `apps/web/e2e/workbook.support.spec.ts`; `apps/web/e2e/workbook.a11y.spec.ts` | Existing Note, pending-navigation/Columns and responsive Commands scenarios | Production shell and semantic browser helpers | Three selected `module.workbook` rows in NR §8 | Extend relevant timing/focus assertions; keep golden and selector contracts unchanged. |
| `tools/frontend_source_ownership.json`; `tools/frontend_import_boundaries.json`; `contracts/verification/owners/web.workbook.json`; `contracts/verification/owners/module.workbook.json`; `tools/test_families/web.workbook.json`; `tools/test_families/module.workbook.json`; `tools/generated_artifact_policy.json` | Independent source, import, verification and generated boundaries | Public Make harness and authored catalogs | Ownership/import/catalog checks | Ownership and workbook row selectors updated with the implementation; generated output changes only through Make. See NR §13. |

Delivered additions, all assigned to source owner `web.workbook`:

| Path (relative to workbook) | Responsibility / consumers | Verification / risk |
| --- | --- | --- |
| `navigation/WorkbookNavigationIntent.ts` | Shared supported destinations, pins, Return origins and semantic equality; session/context/hook consume it | Session and intent rows; typecheck rejects invalid anchors. |
| `navigation/WorkbookNavigationPresentation.ts` | Navigation sequencing over opaque query acceptance, selection acknowledgement, grid and inspector capabilities | Dedicated `navigation_presentation` row and production shell/browser cases; commit-order risk. |
| `navigation/WorkbookNavigationPresentation.test.ts` | Direct deterministic sequencing and replacement tests | Dedicated authored row; no product test-only machinery. |
| `hooks/useWorkbookSemanticGridFocus.ts`; `hooks/useWorkbookSemanticGridFocus.test.tsx` | Existing source selection acknowledgements and grid registrations, reviewed without source changes | `useworkbookstartupcontroller_zsemantic_grid_focus_6f31bf7e92`; now feeds the coordinator through the registry. |

Backend, SQL, source authoring internals, extension-specific record navigation
and broad freshness consolidation remain excluded. Navigation/query/saved-view
and parent guides were reconciled; the Commands guide was reviewed and remains
accurate without edits.

## NR 3. Module boundary diagnosis

| Responsibility | Current location | Correct owner | Keep / move / split / defer | Evidence and decision |
| --- | --- | --- | --- | --- |
| Addressed saved-view observation lifetime | Hook sets/clears a named resource slot; observer owns reads | Saved-view resource owner | Move | The hook's unconditional `finally` can clear a successor's slot. Owner-issued handles hide token matching, cancellation and retention release. |
| Resource failure classification | Observer retains `SavedViewProblem`; navigation receives resource/null plus a status probe | Saved-view owner, with navigation presentation mapping | Split | Return structured outcomes through the host; do not recreate resource or authority policy in navigation. |
| Attempt identity, coalescing and terminal outcome | Session class plus hook readiness ref | Navigation session | Keep / simplify | Session remains the canonical attempt owner; derive pending status from phase and accept only current-attempt presentation events. |
| Selection/focus/inspector sequence | Query registry focus state and hook registration/completion refs | Navigation presentation coordinator | Move | This is one common navigation decision. Query storage and accepted-page lifetime remain independent. |
| Page staging, acceptance and bounded checkpoints | Query browser and registry | Query browsing | Keep | Supply an opaque acceptance handle; no second page, row or checkpoint store in the coordinator. |
| Semantic destinations and Return origins | Broad optional-field types and serialized input key | Navigation types/session | Simplify | Represent extension roots and supported workbook destinations distinctly; compare meaningful fields rather than object encoding. |
| Deliberate-interaction cancellation | Shell calls session and registry separately | Navigation facade, invoked by shell | Move | One cancellation entry cancels obsolete work without consuming the user's event or clearing source work. |
| Commands discovery and owner dispatch | Private Commands index and source bindings | Commands and source owners | Keep | No new registry or execution engine; required action homes and borrowed focus remain useful. |
| Freshness contributions | Existing query/source owners and shell admission predicate | Existing source owners | Defer | Different contributor responsibilities do not justify a universal registry in this iteration. |

### Delivered handle and presentation contracts

The host's saved-view observation capability returns a handle with a typed result
promise and an idempotent release operation. Its signal belongs to the current
navigation attempt. The saved-view owner fences cancellation and release by
handle identity, including same-resource replacement. A stale handle must not
cancel a newer read or remove its retained observation. On admission, establish
selected-resource retention before releasing navigation retention. Cancellation
releases the abandoned observation without removing selected, operation, home or
default retention. Do not merely guard the hook's `finally` with a resource ID:
that cannot distinguish two attempts for the same resource.

| Observation result | Navigation effect | Recovery / disclosure |
| --- | --- | --- |
| Accepted authorized resource | Stage its current query/layout and continue existing admission | Apply even when reselecting the same saved-view ID; accept the resource as comparison baseline on commit. |
| Rejected unavailable target | Preserve the eligible origin; conceal an unavailable pin label | Generic unavailable feedback; explicit authorized base fallback for Return where its owner permits. |
| Rejected operational or contract failure | Preserve origin and leave selection/trail uncommitted | Safe local read retry; no unavailable inference or unavailable-only fallback. |
| Rejected authority failure | Delegate to the existing source/authority recovery policy | Conceal/cancel when authority becomes uncertain; do not infer incident loss from generic resource unavailability. |
| Aborted, superseded or retired observation | No new attachment, selection, failure notice or trail effect | Current attempt controls feedback; older cleanup is harmless. |

Preserve existing `SavedViewProblem` distinctions rather than inventing a second
error taxonomy. Remove the nullable navigation read and separate status probe;
migrate affected internal callers in the same slice. Existing resource-operation
and preference retention are legitimate source responsibilities, not compatibility
aliases to remove.

The session owns the attempt's pending-admission, admitted, succeeded, failed and
cancelled phases. It projects the established outcome names to existing controls
and selectors. Admission settles the existing Return push/pop effect exactly
once; later presentation cancellation does not undo an admitted pivot. Pending
admission and incomplete presentation remain distinguishable.

The presentation coordinator sequences current query acceptance, committed
selection, eligible Grid Adapter focus, and explicit inspector attachment/focus.
It holds semantic identity and an opaque query-acceptance handle, never rows,
drafts, receipts or an authorization cache. Mounted presentation, grid eligibility
and inspector readiness are separate owner signals. Opening Record can make the
grid ineligible without invalidating successfully completed inspector focus.
Calling open, reveal or scroll does not itself complete presentation. Registration
replacement, unmount, authority withdrawal and deliberate interaction fence all
pending callbacks. Known terminal failures settle once; waiting for an eligible
mount does not introduce an invented timeout or polling loop.

Model extension-root origins without workbook query/record anchors. Preserve
record-bearing saved-view Return origins with their captured compatible query,
layout and comparison version. Pin records retain their canonical base schema.
Coalescing compares semantic destination, anchor, entry mode, inspect intent and
applicable captured Return configuration; property insertion order and labels
must not create a different intent. Ordered query/layout arrays retain their
owner meaning. Same-ID saved-view activation after a prior attempt completes
still performs a fresh observation.

## NR 4. Public contract and behavior freeze map

| Contract | Behavioral owner | Evidence / current boundary | Existing tests | Required characterization | Risk / intended change |
| --- | --- | --- | --- | --- | --- |
| Bounded locator and query envelopes | Core 01 REQ-01-680–684 | Locator port/adapter and query owner | Admission and adapter cases | Retain unsupported/no-scan, malformed, outside-query and unavailable distinctions | No HTTP, cursor, schema or backend change. |
| Saved-view observation and recovery | Core 03 saved-view owner requirements and REQ-03-314–315 | Resource observer/controller, host adapter | Independent reads, operations, controller, startup | Old release after successor starts; same-ID replacement; failure-to-recovery mapping | Internal handle cutover; do not change current-resource unavailable recovery or confirmed-operation opening. |
| Navigation/Return/pins | Core 03 REQ-03-313–315 | Session and navigation hook | Session/admission suites | Semantic coalescing; typed extension/workbook origins; stale terminal events | Keep twenty pins, thirty-two origins, admission trail effects and explicit fallback. |
| Presentation and semantic focus | Core 03 REQ-03-314/317; Design §§7, 8.5, 14 | Registry, Grid Adapter, layout and inspector bindings | Admission, shell and selected browser cases | Acceptance → committed selection → focus → optional inspector; replaced registrations | Move coordination, not grid/vendor semantics or visual geometry. |
| Authority and safe disclosure | Core 04 REQ-04-169–170 | Runtime lifetime, source/authority recovery and readable state | Shell authority and Commands cases | Abort/release across uncertainty, replacement, retirement and late replies | Retain same-account source work; no permission from cached labels. Closed incidents remain readable. |
| Retained authoring, writes and receipts | Core 03 REQ-03-311/318 and existing source owners | Source operation owners outside navigation | Note and saved-view owner suites; full workbook gate | Navigation causes no submit/discard/replay; acknowledged-write recovery remains reads only | No mutation, revision, WebSocket or operation-lifetime redesign. |
| Commands | Core 03 REQ-03-312/317 | Private metadata index and owner bindings | Search, interaction and responsive accessibility | Regression only | No new shortcut, search scope or command registry. |
| Generated contracts, selectors and test accounting | Adopted owners; authored source/verification policies | Existing design/view/protocol/UI facades and catalogs | Import, type and catalog gates | New coordinator ownership and executed test selectors | No generated hand edits; no runtime/test dependency on Markdown. |

## NR 5. Coupling and boundary findings

The table retains the baseline diagnosis and accepted remediation. WB-NR00
reproduced NR-F01/02/04; WB-NR01–03 repaired all five findings. Fresh evidence,
including integration failures and repairs, is recorded in NR §§10 and 13.

| Finding | Evidence and classification | Areas / remediation | Rationale and long-term benefit | Compatibility / migration | Risk if unresolved | Validation criteria |
| --- | --- | --- | --- | --- | --- | --- |
| NR-F01 | `must_fix`; source-confirmed cleanup race: hook releases `retainNavigation(null)` in `finally`; shell maps all attempts to the same observer slot; dropping an unretained ID cancels its read | Implementation, tests, docs: owner-issued token-fenced observation handles bound to attempt cancellation | Encapsulates lifetime ownership; older work cannot cancel a successor, including the same resource | Atomic private TypeScript cutover; keep other consumer slots; no persisted migration | A cancelled predecessor can abort the user's latest saved-view selection | Hold A, start B, settle/release A: B remains retained and completes; repeat with the same ID, abort and rejection. |
| NR-F02 | `must_fix`; interface information loss: `read` returns resource/null and navigation separately reads unavailable status | Implementation, types, tests, docs: accepted/rejected/aborted owner result; remove the nullable navigation read/status probe | Failure presentation and recovery remain deterministic without a second resource/authority policy | Migrate affected owner callers and fixtures together; no legacy alias | Operational failures, cancellation and unavailability acquire ambiguous recovery paths as navigation grows | Distinct safe outcomes; retry only through a new read intent; unavailable fallback only where allowed; no protected disclosure. |
| NR-F03 | `should_fix`; structural weakness: session outcome, hook completion ref, registry focus booleans and shell cancellation calls jointly settle presentation | Implementation, types, tests, docs: one navigation-owned coordinator; query-owned acceptance handle; session-owned terminal outcome | One independently testable sequencing decision supports future destinations without scattered feature branches | Preserve selectors/outcome names and owner boundaries; remove transferred state and forwarding in the cutover | Late focus, stale cleanup and indefinitely pending terminal outcomes become harder to diagnose | Current committed acceptance/selection/focus/inspector signals settle exactly once; stale registrations and callbacks have no effect. |
| NR-F04 | `should_fix`; structural/semantic weakness: key serializes whole input objects; target/origin optional fields admit unsupported combinations | Implementation, types, tests, docs: explicit semantic equality and supported destination/origin variants | Removes encoding-dependent coalescing and makes valid future extension paths legible to the compiler | Internal constructor/fixture migration; preserve saved-view record origins and extension root entry; no data conversion | Equivalent activations can supersede instead of coalescing; new consumers can invent unsupported anchors | Reordered equivalent object members coalesce; materially different intent does not; complete typed fixtures compile and existing pin/Return semantics pass. |
| NR-F05 | `must_fix`; verification gap: existing owner tests cover retention and generation fencing separately, while hook fixtures do not establish the shared-slot interleaving | Tests, authored routing, docs: adversarial owner tests plus real shell composition; dedicated coordinator tests | Proves integration and observable focus rather than only callback invocation | Extend owning rows/selectors; preserve old coverage; add a row only for the new owned suite | Isolated passes conceal production cancellation or late attachment defects | Public Make execution includes all added cases; assert selected resource, accepted record, focus, retention and no writes. |

Commands metadata search, owner authorization, saved-view confirmed operations,
startup selection, Note receipt subscriptions and source authoring are
`intentional/no_action` except necessary interface adaptations. Direct vendor,
storage or backend movement is not supported by the inspected scope. Broad
freshness consolidation is `defer`, not a required exit for this iteration.

## NR 6. Refactor workstreams

Execution order is **WB-NR00 → WB-NR01 → WB-NR02 → WB-NR03 → WB-NR04**.
Each slice is a separate workstream. After completing its required checks,
update this controlling tracker and its handoff log before beginning its dependent.
Record changed paths, findings, commands, run roots, failures, compatibility,
risks and rollback. A failed required exit remains BLOCKED, not DONE.

| Workflow ID | Name | Class | Previous | Subsequent | Goal / likely files | Validation | Handoff checkpoint |
| --- | --- | --- | --- | --- | --- | --- | --- |
| WB-NR00 | Contract alignment and characterization | root | None | WB-NR01 | Revalidate inventory, owners and baseline; characterize retention replacement and semantic coalescing in owning tests | NR §8 baseline and targeted reproductions | Separate expected pre-fix failures from passing baseline; all contracts and scenarios mapped. |
| WB-NR01 | Saved-view observation ownership | chain | WB-NR00 | WB-NR02 | Observer/controller, host, hook, shell adapter and affected fixtures | Independent reads, operations, admission, controller, preferences/startup, typecheck | Successor-safe handles and typed recovery complete; NR-F01/02 evidence recorded. |
| WB-NR02 | Semantic intent and attempt state | chain | WB-NR01 | WB-NR03 | Context/host/session types, hook capture/coalescing and constructors | Session/admission/status, shell and typecheck | Semantic equality, valid variants and admission trail effects complete; NR-F04 evidence recorded. |
| WB-NR03 | Presentation coordination cutover | chain | WB-NR02 | WB-NR04 | New navigation coordinator, registry/browser acceptance bridge, hook, shell, layout/inspector bindings | Coordinator, query browsing/controls, admission, shell, selected browser/accessibility and static gates | NR-F03 and integrated NR-F05 cases complete; obsolete completion/focus plumbing removed. |
| WB-NR04 | Validation and handoff completion | chain | WB-NR03 | None | Source guides, authored ownership/routing, affected generated projections and this tracker | NR §8 final gates and digest assessment | All applicable criteria PASS or justified N/A; removal, compatibility, limitations and restart handoff complete. |

## NR 7. Accepted implementation slices

All source/test slices below are authorized for this implementation. NR-F01
and NR-F02 are owner-aligned behavior corrections; NR-F03 and NR-F04
primarily restructure ownership/types while correcting incidental coalescing.
Do not broaden these changes into new product features.

| Slice | Depends on | Intended change | Files/packages likely involved | Contract risk | Tests to add/preserve | Validation command / selection | Rollback | Binary exit |
| --- | --- | --- | --- | --- | --- | --- | --- | --- |
| WB-NR00 | None | Revalidate current HEAD/dirty state, requirements and callers; write deterministic reproductions before fixes | Existing admission/session/independent-read tests; authored selectors; tracker | Treating historical evidence as current or a failing reproduction as a completed correction | A→B retention release, same-ID replacement and equivalent intent with reordered members | Public owner guides and selected `make test-slice OWNER=web.workbook ROWS=...` from NR §8 | Characterization/routing changes are independently reversible; keep failure evidence in the log | Complete owner-to-gap map and repeatable pre-fix characterization; unrelated required baseline failures block dependent work. Expected defect failures remain explicitly open for their repair slices. |
| WB-NR01 | WB-NR00 | Introduce handle identity, cancellation/release and typed outcomes; transfer selected retention before release; remove old navigation trio | Saved-view observer/controller/port, host, shell adapter, hook and fixtures | Cancelling another consumer's read or breaking selected-resource recovery | Old cleanup/rejection/abort cannot cancel B; startup, open-confirmed, preferences and deletion/unavailability recovery retained | Independent reads, owner operations, controller, startup, preferences and admission rows; `make frontend-typecheck` | Revert complete owner/host/caller/fixture cutover; no one-sided interface rollback | NR-F01/02 reproductions and all affected owner behavior pass; no nullable navigation read/status probe remains. |
| WB-NR02 | WB-NR01 | Model target/origin variants, explicit semantic equality and session phase; derive status without another source store | Navigation/session/context and all changed constructors/tests | Rejecting legitimate saved-view Return anchors or altering trail effects | Equivalent coalescing, distinct intent, pin limits, bounded trail, extension roots and stale terminal events | Session/admission/feedback/shell rows; `make frontend-typecheck` | Revert types, constructors, equality and session state together | No object-order dependency; valid variants compile; current-attempt outcomes and existing admission push/pop rules pass. |
| WB-NR03 | WB-NR02 | Introduce the coordinator and opaque acceptance handle; migrate explicit open/focus registration; route shell cancellation once; remove transferred state | Navigation, query registry/browser, shell, layout adapter, inspector/view-bar bindings and tests | Commit ordering, Strict Mode, source acceptance versus mount/focus, late cleanup | Delayed/replaced mounts, acceptance failure, committed selection, grid fallback, explicit inspect, interaction delivery and authority loss | Coordinator and query/navigation/shell rows; selected browser/accessibility invocation; type/import checks | Revert complete coordinator, bindings, acceptance bridge, routing and tests together | Accepted selection precedes focus; current explicit inspector attachment completes; terminal failure/cancel settles; old callbacks cannot attach or steal focus. |
| WB-NR04 | WB-NR03 | Reconcile guides, account for files/tests, generate affected downstream artifacts, execute final verification and hand off | Navigation/query/saved-view/Commands guides, parent guide, authored policies/catalogs and this tracker | Incomplete routing or documentation lint mistaken for product readiness | Preserve original regression coverage plus all NR cases | Full final gate set in NR §8, after `make agent-finalize` | Retain evidence; roll back full implementation slices if needed, never only a facade or adapter | All workstreams DONE; every applicable acceptance row PASS or justified N/A; no required check deferred. |

The cutover is private and atomic per slice: no HTTP/database migration, browser
storage conversion, compatibility alias, deprecation window or dual implementation.
Retain source-owned `activateResource`/`openConfirmed`, startup acceptance and
automatic recovery of a currently selected unavailable resource. Remove the old
navigation retention/read/status trio, hook completion ref, transferred registry
focus flags and obsolete cancellation forwarding with their migrated consumers.
Do not remove query binding tokens, Grid Adapter focus cancellation or source
operation state merely because their names resemble the retired state.

## NR 8. Validation plan

Use public Make targets from the repository root. Source ownership and test
routing remain independent; rerun the owner guides if the implementation baseline
changes. Historical run roots below the completed-record boundary remain historical.
No product test result is claimed by this plan.

| Validation layer | Command / scope | Required before implementation? | Evidence / use |
| --- | --- | --- | --- |
| Owner discovery | `make task-guide ROLE=module-author OWNER=web.workbook`; repeat for `module.workbook` and `web.architecture` | Yes, in WB-NR00 | Current routing guidance; these commands completed during planning. |
| Focused unit/composition | `make test-slice OWNER=web.workbook ROWS=<selected rows below>` | Baseline and characterization in WB-NR00 | Add title selectors with tests; record pre-fix failures separately from passing exits. |
| Selected browser/accessibility | Exact invocation below | Required for WB-NR03 and final handoff | Real Note selection/inspector focus, Columns interaction delivery and responsive Commands. |
| Authored/generated accounting | `make test-catalog-check`; `make generate` only for affected authored inputs; applicable generated-policy/drift targets | No generated edits for this document task | No manual edits to generated topology or contract artifacts. Record actual outputs and why they changed. |
| Finalizer | `make agent-finalize` | After focused implementation checks, before broad final checks | Leave `RESULTS_DIR` unset unless eligible successful full warm-check evidence exists; record retained-run maintenance skip when unset. |
| Full scoped frontend | `make test-slice OWNER=web.workbook`; `make test-slice OWNER=web.architecture` | Final implementation gates | Whole affected frontend ownership and architecture, not release-wide qualification. |
| Static/build | `make frontend-typecheck`; `make frontend-import-boundary-check`; `make lint-biome`; `make test-catalog-check`; `make build-web` | Final implementation gates | New coordinator owns a real common decision; no vendor leakage, circular dependency or missing routed tests. |
| Documentation | `make lint-markdown`; `git diff --check` | Required for WB-NR04 | Documentation/whitespace evidence only; never product certification. |
| Backend, release-wide, broad visual | Select through public guides only if a changed boundary requires them | No default gate | Excluded for this frontend structural scope; unexpected visual changes require investigation, not automatic golden updates. |

Focused frontend rows below use the prefix `web.workbook.regression.`. Preserve
the original seven, and select additions by the listed workstream. These are
verified active rows, with additions identified by the implementation workstreams.

| Row suffix | Owning evidence | Required use |
| --- | --- | --- |
| `workbench_navigation_admission` | `navigation/useWorkbookWorkbench.test.tsx` | All changed navigation behavior and new interleavings. |
| `session_navigation` | `navigation/WorkbookSessionNavigation.test.ts` | Intent equality, phases, pins and Return. |
| `navigation_feedback` | `navigation/WorkbookNavigationStatus.test.tsx` | Stable outcomes, accessible feedback and concealment. |
| `command_search` | `commands/workbookCommandIndex.test.ts` | Preserved deterministic matching. |
| `workbench_commands_interaction` | `commands/WorkbookCommands.test.tsx` | Captured target, borrowed focus and revalidation. |
| `note_associations` | `features/notes/noteAssociations.test.tsx` | Preserved source association/removal and retained-operation behavior. |
| `activesurfacesavedviewselector_resource_action_focus_a103000001` | `components/ActiveSurfaceSavedViewSelector.test.tsx` | Browser actions retain the required navigation path and owner focus. |
| `query_browsing` | `query/WorkbookQueryBrowser.test.ts` | Staged acceptance, bounded checkpoints and query-owner recovery. |
| `query_browsing_controls` | `query/WorkbookQueryBrowsingControls.test.tsx` | Binding replacement, Strict Mode, committed read lifetime and control focus. |
| `saved_view_independent_reads` | `savedviews/savedViewReads.test.ts` | Token-fenced retention, cancellation and distinct read outcomes. |
| `saved_view_operation_owner` | `savedviews/WorkbookSavedViewController.test.ts` | Confirmed-operation opening and retained resource operations. |
| `useworkbooksavedviewcontroller_suite_d3f1a57e5d` | `hooks/useWorkbookSavedViewController.test.tsx` | Controller binding and current-resource recovery. |
| `use_workbook_startup_admission_suite_919ed32f45` | `startup/useWorkbookStartupAdmission.test.tsx` | Startup acceptance remains source-owned. |
| `preference_characterization` | `preferences/workbookPreferenceCharacterization.test.tsx` | Home/default observation and selection behavior. |
| `workbookshell_surfaces_suite_668e482b1e` | `WorkbookShell.surfaces.test.tsx` | Production composition, same-ID reselection, authority and committed focus. |

Added `web.workbook.regression.navigation_presentation` for the direct coordinator
suite and `web.workbook.regression.workbench_navigation_intent` for semantic
coalescing. The existing semantic-focus row
`web.workbook.regression.useworkbookstartupcontroller_zsemantic_grid_focus_6f31bf7e92`
is included in the integration selection. Catalog checks account for every title.
No test-only machinery was added to product runtime.
Existing shell tests must use the real resource observer and presentation wiring
for the new integration cases; transport holding/delivery remains fixture-owned.

The required selected browser/accessibility invocation is:

```bash
make test-slice OWNER=module.workbook ROWS=module.workbook.browser.note_associations,module.workbook.browser_support.pending_navigation_preserves_interaction,module.workbook.accessibility.workbench_commands
```

| Scenario | Observable pass/fail criterion | Primary evidence |
| --- | --- | --- |
| A superseded by B | A's cancel, rejected result, release and late cleanup cannot cancel B or remove B's retention; B selects once | Resource-owner tests and real shell composition. |
| Same-resource replacement | Identical pending semantic activation coalesces; a distinct newer intent for the same ID has its own handle; old cleanup is harmless | Session, hook and owner tests. |
| Semantic identity | Equivalent member ordering coalesces; different sheet/record/field/inspect/entry mode or relevant Return configuration does not | Navigation/session tests with complete typed fixtures. |
| Admission-to-focus | Accepted owner page precedes committed selection; selection precedes focus; inspect waits for matching committed attachment and actual focus | Coordinator tests and production shell/Note browser case. |
| Delayed and replaced attachment | Delayed mount, changed grid handle, changed inspector binding, unmount and Strict Mode reject predecessor callbacks and cleanup | Coordinator, query controls and composition. |
| Deliberate interaction | Columns receives the user's pending-navigation click/key action; obsolete focus and inspector intent cannot resume afterward | Existing selected pending-navigation browser row with relevant assertions. |
| Failure distinctions | Outside-query, unavailable, operational, malformed and unsupported results preserve eligible origin; cancellation does not manufacture failure; unsupported never scans pages | Admission/adapter and saved-view outcome cases. |
| Saved-view reselection | Same selected ID freshly observes and applies authorized persisted configuration; pending failure preserves origin | Shell surfaces and controller integration. |
| Return | Captured compatible query/layout and latest resource comparison remain distinct; failed/cancelled admission retains entry; successful fallback consumes it; admitted trail effects occur once | Session/admission and shell cases. |
| Retained work | No navigation submits/discards drafts, dismisses receipts, replays writes or replaces source identity; acknowledged recovery dispatches reads only | Existing Note/saved-view owner cases and composition assertions. |
| Authority | Uncertainty conceals and cancels; same-account recovery reauthorizes; account replacement/incident loss retires state; closed remains readable; late observations cannot disclose | Shell authority and Commands regression. |
| Regression boundaries | Commands matching/pages/borrowed focus, duplicate-before-capacity pins, startup, confirmed operations and selected-resource recovery continue | Retained focused rows and full workbook gate. |

Keep tokens, geometry, selectors and visual goldens unchanged. An unexpected
pixel difference is a regression to investigate. Any intentional visual correction
requires the repository UI-review and golden-maintenance procedure as a separately
identified change. Required failures must be reported with target, run root or
summary, relation to the change and blocked exit. Do not hide an unrelated failure
or weaken an assertion to mark a workstream complete.

## NR 9. Top-level work tracker

| ID | Work item | Workstream | Status | Depends on | Evidence / artifact | Exit condition |
| --- | --- | --- | --- | --- | --- | --- |
| NR-DOC | Publish this accepted next-iteration plan | Document update only | DONE | Accepted navigation-reliability scope | NR §§1–12; current source/catalog inspection; documentation evidence in NR §10 | Only this tracker changed; Markdown and whitespace checks pass; completed historical body preserved. |
| WB-NR00 | Contract alignment and reproducible characterization | WB-NR00 | DONE | Authorized implementation | Executed NR §8 baseline/scenarios | Current inventory/owner mapping and diagnostic reproductions complete. |
| WB-NR01 | Saved-view observation ownership | WB-NR01 | DONE | WB-NR00 | NR-F01/02 | Handles, failure mapping and affected owner behavior pass. |
| WB-NR02 | Semantic intent and attempt state | WB-NR02 | DONE | WB-NR01 | NR-F04 | Semantic coalescing, valid origins and lifecycle checks pass. |
| WB-NR03 | Presentation coordination cutover | WB-NR03 | DONE | WB-NR02 | NR-F03/05 | Committed selection/focus, cancellation and production integration pass; old plumbing removed. |
| WB-NR04 | Final validation and handoff | WB-NR04 | DONE | WB-NR03 | NR §8 gates; digest assessment | All required validation complete and final tracker/handoff updated. |

Do not mark a future workstream DONE because this planning document is complete.
During execution, update each row and the appropriate log after its required exit
and before beginning its dependent. Record limitations separately from required
checks; deferring a required check does not satisfy the exit.

## NR 10. Session handoff log

### Implementation execution checkpoints

- **WB-NR03 DONE (2026-10-08):** Navigation now owns the presentation
  coordinator; query browsing owns the opaque acceptance handle. Grid registration
  notifications are independent of React data subscriptions. Selection tokens
  fence stale acknowledgements; a changed grid handle aborts pending focus without
  undoing committed selection/completed focus during inspector attachment.
  The first integration run (`nr03-integration-20261008a`) exposed a shell render
  loop, repaired by separating presentation notifications. The next shell run
  (`nr03-coordination-20261008b`) exposed a Note-navigation timeout, repaired by
  distinguishing grid-handle replacement from selection replacement. Direct
  coordinator and shell checks passed at `nr03-shell-20261008c`; all seven focused
  rows passed at `nr03-final-focused-20261008d`. Query acceptance/rejection/release
  characterization passed at `nr03-acceptance-20261008a`.
- Typecheck passed at `nr03-static-20261008c`. Combining additional public targets
  with that explicit run ID was rejected because the directory was nonempty;
  reran them separately: import boundaries PASS at `nr03-imports-20261008a`,
  catalog PASS at `nr03-catalog-20261008b`. All roots above are beneath
  `.cartulary/test-results/`.
- Browser run `nr03-browser-20261008a`: Note navigation and Commands accessibility
  passed; pending-navigation/Columns timed out waiting for a response to the now
  correctly aborted fetch. Trace confirmed the click reached Columns. Updated
  the browser case to assert request cancellation and drain late routing while
  retaining all existing menu/geometry/no-attachment checks. Focused browser
  rerun `nr03-browser-support-20261008b` PASS. Added a retained-refresh detachment
  guard and owner regression during review; owner reads/operations/admission
  PASS at `nr03-owner-lifetime-20261008b`. Final selected browser/accessibility
  PASS at `nr03-browser-final-20261008c` (15/15 units), typecheck PASS at
  `nr03-types-final-20261008d`, catalog PASS at `nr03-catalog-final-20261008c`.
  Formatting PASS at `nr03-format-20261008b`. No required WB-NR03 check remains
  open. Roll back coordinator, query bridge, bindings, callers and routing
  together. Recorded before WB-NR04 guide reconciliation and broad validation.
- **WB-NR04 DONE (2026-10-08 local / 2026-10-09 UTC):** Reconciled the four
  source guides and active tracker wording; preserved both historical completed
  records. Registered sources/tests, regenerated the topology render index and
  completed finalizer before broad gates. Full workbook (326/326 units), full
  architecture (13/13), required browser/accessibility (15/15), typecheck, imports,
  Biome, catalog, web build, generated policy and Markdown pass. Final acceptance
  review added an exhausted-focus assertion to an existing coordinator test;
  its affected row, typecheck and Biome reruns pass. Production code is unchanged
  since the final browser run. NR §13 records exact roots, failures/repairs,
  all 27 digest dispositions, migration/rollback and justified skips.
  `RESULTS_DIR` was unset because no eligible successful full warm-check run
  exists; retained-run maintenance was explicitly skipped. No required check or
  unresolved navigation defect is deferred.

- **WB-NR02 DONE (2026-10-08):** Added owned semantic intent/origin/pin types,
  semantic equality and phase-derived pending state. Equivalent admitted attempts
  coalesce until terminal presentation; subsequent activation observes afresh.
  Typed fixtures now include real query/layout data. Session/admission/status,
  semantic-intent and real shell coverage ran at
  `.cartulary/test-results/nr02-focused-20261008a`: four rows passed; the pin
  fixture needed explicit completion before simulating another activation.
  Corrected that bounded fixture and admission passed at
  `.cartulary/test-results/nr02-admission-20261008b`.
  Initial typecheck exposed incomplete fixtures and overly narrow root-reference
  unions; repaired them, with PASS at `.cartulary/test-results/nr02-types-20261008b`.
  All required WB-NR02 evidence is passing. No wire/storage conversion or string
  compatibility overload remains. Rollback types, equality, lifecycle and callers
  together. Recorded before WB-NR03 implementation.

- **WB-NR01 DONE (2026-10-08):** Owner-issued observation handles now fence
  release/cancellation by identity; navigation uses accepted/rejected/aborted
  results and retains the selected resource before release. Confirmed activation
  and discovery dismissal release only their own handle. Removed navigation's
  shared retention/read/status facade. Existing retained owner refresh remains
  source-owned. No protocol, persistence or compatibility alias was introduced.
  Six focused rows passed at `.cartulary/test-results/nr01-focused-20261008a`;
  typecheck passed at `.cartulary/test-results/nr01-types-20261008a`.
  The intentional NR-F04 failure is separately routed under
  `web.workbook.regression.workbench_navigation_intent` for WB-NR02.
  Changed observer/controller, host/shell/hook, owner/admission tests and routing.
  Rollback requires the complete owner/host/caller/fixture cutover. No required
  WB-NR01 check remains open. Recorded before WB-NR02 implementation.

- **WB-NR00 DONE (2026-10-08):** Revalidated clean baseline
  `c8f3a4ede2b31152eceb3fff864fb67c5f1a017f` and all three owner guides.
  Inspected the governing Core 01/03/04 clauses, Design/Domain and current
  navigation, query, saved-view, shell, inspector and semantic-grid-focus owners.
  The inventory additionally includes `hooks/useWorkbookSemanticGridFocus.ts`
  and its test suite, routed by
  `web.workbook.regression.useworkbookstartupcontroller_zsemantic_grid_focus_6f31bf7e92`.
  Cross-consumer cancellation includes discovery closing/confirmed activation.
- Baseline `make test-slice OWNER=web.workbook` with independent-read, session
  and admission rows passed at `.cartulary/test-results/nr00-baseline-20261008a`.
  Added three routed admission regressions. The first reproduction launch failed
  before execution (`artifact/artifact_error`); the explicit catalog diagnostic
  at `.cartulary/test-results/nr00-catalog-20261008a` identified unsorted authored
  title selectors. Sorted them and reran with a fresh run identity.
- `.cartulary/test-results/nr00-reproductions-20261008b` reproduced all three
  intended assertion failures: predecessor cleanup aborts its successor,
  equivalent member ordering dispatches twice, and operational saved-view
  Return offers unavailable fallback. Existing admission cases passed.
  NR-F01/02 repair belongs to WB-NR01; semantic coalescing to WB-NR02. No unrelated
  baseline defect remains. Reproductions and routing are independently reversible.
  This checkpoint was recorded before WB-NR01 production changes.


The following planning-session entries are historical. Their document-only
restrictions and skipped product checks do not apply to the authorized execution
checkpoints above or the final handoff below.

### NR scope and authority

| Time | Session | Current state | Files inspected / touched | Commands | Result | Blockers | Next action |
| --- | --- | --- | --- | --- | --- | --- | --- |
| 2026-10-08 | Planning and accepted document update | User selected navigation reliability; implementation remains future work | Root instructions, owner clauses, Design/Domain, NLSpec research, framework/digest and this tracker | Source reads; `git rev-parse HEAD`; `git branch --show-current`; `git status --short` | Baseline `b28709d065a37c4748fc1c42dcb783ab613925a0`, `main`, initially clean; only tracker write authorized | No owner contradiction identified | Publish and validate document; later implementation starts WB-NR00. |

### NR backend boundary

| Time | Session | Current state | Files inspected / touched | Commands | Result | Blockers | Next action |
| --- | --- | --- | --- | --- | --- | --- | --- |
| 2026-10-08 | Planning | Backend/protocol/storage frozen | Locator port/adapter and governing Core interface; no backend files changed | Targeted interface reads | No backend audit or product pass claimed; no new route/schema/store proposed | None for document scope | Expand verification only if later changes actually cross this boundary. |

### NR frontend boundary

| Time | Session | Current state | Files inspected / touched | Commands | Result | Blockers | Next action |
| --- | --- | --- | --- | --- | --- | --- | --- |
| 2026-10-08 | Planning/document update | Fourteen target files accounted for; adjacent integrations bounded in NR §2 | Navigation, Commands, registry/browser, saved-view observer/controller, shell, layout and inspector bindings | `rg --files`, caller searches, exact source reads | NR-F01 source-confirmed; NR-F02 information loss and NR-F03/04 structural weaknesses recorded; no production edit | Runtime reproductions not yet run | WB-NR00 creates deterministic characterization under later authorization. |

### NR contract and codegen

| Time | Session | Current state | Files inspected / touched | Commands | Result | Blockers | Next action |
| --- | --- | --- | --- | --- | --- | --- | --- |
| 2026-10-08 | Planning | Internal interfaces may change later; adopted contracts stay controlling | Source/import policies, generated policy and selected verification inputs | Targeted manifest/source reads | No owner, projection, generated output, dependency or catalog edited; tests/runtime remain independent of Markdown | None | Later slices register source/tests through authored owners and generate only affected outputs. |

### NR tests and harness

| Time | Session | Current state | Files inspected / touched | Commands | Result | Blockers | Next action |
| --- | --- | --- | --- | --- | --- | --- | --- |
| 2026-10-08 | Planning | Active rows and public owner guidance discovered | Navigation/query/saved-view/shell tests and authored row selectors | `make task-guide ROLE=module-author OWNER=web.workbook`; same for `module.workbook` and `web.architecture` | Guides passed; NR §8 records selected routing. No fresh product suite executed | New interleavings await WB-NR00 | Preserve baseline rows and extend owning selectors with each new case. |
| 2026-10-08 local / 2026-10-09 UTC | Document update complete | NR-DOC DONE; all five implementation workstreams TODO | This tracker only | `make lint-markdown`; `git diff --check`; `git status --short`; manual comparison with `git show HEAD:<path>` | Markdown PASS at `.cartulary/test-results/20261009T002509Z-p10571`, summary `adhoc/lint-markdown/tool-run-summary.json`; whitespace PASS; one-file diff; completed body byte-identical from `Execution checkpoints` onward; nine navigation and five Commands files accounted for | None for document update; no product pass claimed | Recheck lint after recording evidence; later implementation starts WB-NR00. |

During the historical document-only task, product tests, builds, generation,
visual comparison and finalization were skipped because no executable input
changed. Current implementation verification and justified skips are in NR §13.

### NR security and authorization

| Time | Session | Current state | Files inspected / touched | Commands | Result | Blockers | Next action |
| --- | --- | --- | --- | --- | --- | --- | --- |
| 2026-10-08 | Planning | Authority and operation lifetime remain owner-held | Core 04 §2.2, session/readable behavior, resource-owner failure recovery and shell cancellation | Targeted reads and call-site review | Plan retains concealment, reauthorization and scoped retirement; no ACL/cache/persistence added | No contradiction identified | Validate cancellation/release across authority transitions with late results. |

### NR open risks and next session

| Time | Session | Current state | Files inspected / touched | Commands | Result | Blockers | Next action |
| --- | --- | --- | --- | --- | --- | --- | --- |
| 2026-10-08 | Document update | Future workstreams remain TODO | This tracker and source baseline | `git status --short`; inventory/reference review | Main risks are same-ID cleanup, cross-consumer cancellation, commit ordering and incomplete routing | Later implementation task required; not a defect in the document | Revalidate HEAD/dirty state, rerun owner guidance and start WB-NR00; update tracker before each dependent. |

## NR 11. Open questions and blockers

No unresolved product preference or owner contradiction blocks this plan. The
user chose navigation reliability over a broader readiness audit. Reproductions and implementation validation are recorded in the execution log;
WB-NR04 closed the final gates in NR §13, without borrowing historical passes.

| ID | Question / gate | Why it matters | Required authority or evidence | Status |
| --- | --- | --- | --- | --- |
| NR-RB01 | Start implementation | User explicitly authorized the accepted plan | Current implementation request | RESOLVED |
| NR-RB02 | Establish reproducible NR-F01/04 evidence | Source inspection is not an executed regression demonstration | WB-NR00 routed characterization and classified results in NR §10 | RESOLVED |

If a genuine owner contradiction appears, add its exact source clauses and mark
only dependent work BLOCKED. If required verification fails, record the failure
and blocked exit; do not relabel it as out-of-scope debt to finish the iteration.

## NR 12. Binary completion criteria

The planning-only NR-DOC deliverable is historical. The authorized implementation
iteration is complete only when WB-NR04 establishes:

- All five workstreams are DONE with each completed tracker update preceding
  its dependent; all required reproductions and final gates pass.
- No stale handle can cancel its successor; resource outcomes remain distinct;
  semantic coalescing, current-attempt status and committed presentation are proven.
- Old navigation retention/read/status calls, duplicated completion state and
  transferred focus/cancellation forwarding have no remaining consumers or aliases.
- Query pages, checkpoints, source work, authorization and recovery remain with
  their owners; navigation adds no row, draft, receipt or authorization store.
- Final source guides, ownership and authored verification routing match code;
  generated changes have owner/generator evidence and no Markdown dependency.
- Every applicable [digest acceptance row](../../cartulary-ui-ux-refactor-digest/cartulary/acceptance.tsv)
  is PASS with fresh evidence or N/A with a specific scope/owner rationale.
  Applicable BLOCKED criteria prevent completion. Prior assessments are historical.
- The final handoff records removed paths, retained capabilities, commands/run
  roots, failures/repairs, compatibility, justified skips, residual out-of-scope
  debt, complete-slice rollback and restart instructions. Scoped navigation
  readiness is distinguished from repository-wide production qualification.

## NR 13. Final validation and handoff

### Delivered changes and owner accounting

The correction keeps observation identity in saved views, attempt identity in the
session, presentation sequencing in navigation, and accepted pages/checkpoints in
query browsing. No new normative requirement was needed: Core 01 REQ-01-680–684,
Core 03 REQ-03-311–318 and Core 04 REQ-04-169–170 govern the changes. Design and
Domain remain presentation/vocabulary references. Source ownership is
`web.workbook`; verification is independently routed by `web.workbook`,
`module.workbook` and `web.architecture`.

Changed source paths below are relative to `apps/web/src/workbook/`; NR §2 gives
caller/dependency detail. Unchanged inspected owners include runtime authority,
startup, preference, Commands, semantic grid focus, layout and inspector focus.

| Changed paths | Final responsibility / substantive edit |
| --- | --- |
| `savedviews/SavedViewResourceObserver.ts`; `savedviews/WorkbookSavedViewController.ts` | Handle-specific retention/cancellation and typed outcomes; transfer selected retention before release; background refresh stops when its last concrete consumer leaves. |
| `navigation/WorkbookNavigationIntent.ts`; `navigation/WorkbookSessionNavigation.ts`; `navigation/WorkbookWorkbenchContext.tsx` | Valid destination/origin/pin variants, semantic coalescing until terminal presentation, phase-derived feedback and exactly-once admission trail effects. |
| `navigation/WorkbookNavigationPresentation.ts`; `navigation/WorkbookNavigationHost.ts`; `navigation/useWorkbookWorkbench.ts`; `WorkbookShell.tsx` | Narrow observation capability, one presentation coordinator, typed failure mapping and one shell cancellation call per triggering event. |
| `query/WorkbookQueryBrowser.ts`; `query/WorkbookQueryBrowsingContext.tsx` | Query-owned opaque acceptance handle; token-fenced selection/presentation registrations; independent presentation notifications; removed navigation focus sequencing. |
| `components/WorkbookViewBar.tsx`; `components/GenericWorkbookSurface.tsx`; `components/AssessmentWorkbookSurface.tsx`; `components/EntityWorkbookSurface.tsx`; `timeline/presentation/useTimelineWorkbookPresentation.tsx`; `timeline/presentation/TimelineWorkbookViewBarRegion.tsx` | Coordinator-backed inspector registration and accurate private `onInspectorOpen` naming; existing idempotent owner open behavior retained. |
| `navigation/WorkbookNavigationPresentation.test.ts`; `navigation/WorkbookSessionNavigation.test.ts`; `navigation/useWorkbookWorkbench.test.tsx`; `navigation/WorkbookNavigationStatus.test.tsx`; `savedviews/savedViewReads.test.ts`; `query/WorkbookQueryBrowser.test.ts`; `WorkbookShell.surfaces.test.tsx` | Direct sequencing, owner replacement/failure, semantic intent, acceptance and real-shell race regressions; typed fixtures and retained existing coverage. |
| `README.md`; `navigation/README.md`; `query/README.md`; `savedviews/README.md` | Reconciled final owner boundaries and lifetimes; Commands guide reviewed unchanged. |
| `apps/web/e2e/workbook.support.spec.ts` (repository relative) | Assert the intentionally aborted request while proving delivery of Columns interaction and absence of late destination attachment. |
| `tools/frontend_source_ownership.json`; `tools/test_families/web.workbook.json` (repository relative) | Three new source/test files registered; two new semantic rows and updated exact test-title selectors. |
| `tools/execution_topology_render_index.json` (repository relative) | Generator-produced digest update after catalog changes; no hand edit. |
| This tracker | Owner traceability, ordered workstream checkpoints, failures/repairs, acceptance and restart handoff. |

The digest localization remains a historical September snapshot; current authored
source/import/catalog inputs supersede its inventory for this implementation.
No digest update, direct grid-vendor import, token literal, theme/density registry,
new dependency, backend change, wire/schema change or persisted-state conversion
was introduced.

### Retirement and justified retention

| Removed or retained | Decision and reason |
| --- | --- |
| Shared navigation/activation retention slots and resource-ID cancellation | Removed. A caller now releases the acquired observation handle; stale cleanup cannot target the latest occupant of an ID. |
| Navigation `retainNavigation` / nullable `read` / `unavailable` host trio | Removed atomically with callers and fixtures; no compatibility alias. Typed accepted/rejected/aborted results retain source failure meaning. |
| Whole-intent serialized keys and string navigation overload | Removed. Semantic equality and supported variants express intent. Fixed semantic pin-key tuple encoding remains appropriate for pin identity. |
| Hook completion ref and query-registry navigation focus flags/algorithm | Removed. One navigation coordinator sequences acceptance, committed selection, grid focus and explicit inspector focus. |
| Shell session-plus-query cancellation forwarding | Removed. One facade cancels navigation without consuming pointer/key input. |
| Selected, operation, home and default retention | Retained for actual source/resource consumers. Source-owner `read` remains an auto-released refresh helper, not a navigation compatibility facade. |
| Startup, confirmed-operation activation and selected-resource recovery | Retained as distinct owner duties; discovery cleanup releases only its owner activation. |
| Ordinary entry-focus owner and source selection acknowledgements | Retained because creation-first entry and committed source selection remain source responsibilities; the coordinator consumes their capabilities. |
| Query page/row/checkpoint storage | Retained solely in query browsing; acceptance handle release preserves an already accepted page. |
| Commands, pins, Return, drafts, receipts, operation/recovery ownership | Retained to satisfy adopted behavior; navigation never writes, discards authoring, retries mutations or owns authorization. |
| Visual selectors, tokens, geometry and goldens | Unchanged. Browser assertions exercise the existing production presentation without baseline replacement. |

### Fresh verification ledger

All run IDs below resolve beneath `.cartulary/test-results/`. Each invocation used
its own fresh `CARTULARY_TEST_RUN_ID`. Work-unit counts are harness units, not test
counts. WB-NR00–03 focused selections and failures are in NR §10.

| Command / selection | Result | Run ID / evidence |
| --- | --- | --- |
| `make agent-finalize` (first attempt); diagnostic `make json-shape-check` | Failed: new row catalog made generated topology stale; related to this change, no product defect | `nr04-finalize-20261008a`, `unit-artifacts/finalize-summary.json`; `nr04-shape-20261008a`, `unit-logs/target-json-shape-check/stderr.log`. |
| `make generate` | PASS; refreshed only the topology render index in tracked generated output | `nr04-generate-20261008a`, `generate/tool-run-summary.json`. |
| `make agent-finalize` (after generation) | PASS before broad gates; schema, catalog, tier coverage and generation/drift transaction pass | `nr04-finalize-20261008b`, `unit-artifacts/finalize-summary.json`. No further generated mutation. |
| `make test-slice OWNER=web.workbook` | PASS, 326/326 units, zero skipped | `nr04-workbook-20261008a`; final production implementation. |
| `make test-slice OWNER=web.architecture` | PASS, 13/13 units | `nr04-architecture-20261008a`. |
| `make frontend-typecheck` | PASS | `nr04-types-20261008a`; final test refinement also PASS at `nr04-types-20261008b`. |
| `make frontend-import-boundary-check` | PASS | `nr04-imports-20261008a`. |
| `make lint-biome` | PASS | `nr04-biome-20261008a`; final test refinement also PASS at `nr04-biome-20261008b`. |
| `make test-catalog-check` | PASS | `nr04-catalog-20261008a`. |
| `make build-web` | PASS | `nr04-build-20261008a`. |
| `make generated-artifact-policy-check` | PASS | `nr04-generated-policy-20261008a`. |
| Required three-row `module.workbook` browser/accessibility selection in NR §8 | PASS, 15/15 units; final production code | `nr03-browser-final-20261008c`; no later production change invalidates this evidence. |
| `make lint-markdown`; `git diff --check` | PASS; completion entry rechecked | `nr04-markdown-20261008a` and final `nr04-markdown-20261008b`, `adhoc/lint-markdown/tool-run-summary.json`; whitespace check has no artifact. |
| `make test-slice OWNER=web.workbook ROWS=web.workbook.regression.navigation_presentation` | PASS, five coordinator tests | `nr04-focus-exhaustion-20261008a`; final review added all-targets-unavailable to the existing terminal-outcome case. No production or catalog change. |

Retained-run maintenance was skipped by the finalizer because `RESULTS_DIR` was
unset: this effort has no eligible successful full warm-check run. This does not
skip generation or authored-catalog validation. Broad backend, full release,
performance/measurement and visual-golden suites were not selected: no backend,
protocol, geometry, theme, density, token or golden input changed. The scoped
browser/accessibility selection is current evidence; it is not a new broad visual
or Core 05 publication claim.

### Navigation acceptance matrix

Evidence abbreviations below refer to the rows in NR §8 executed by the full
workbook gate and the final selected browser run. Row terminal artifacts are
`rows/<row-id>.json`; exact executed assertion results are retained in
`unit-logs/row-<row-id>/runner.json`. Catalog validation checks authored title
accounting. The coordinator's final targeted run supplements the full owner pass
with an extra exhausted-focus assertion inside an existing case; production
sources and the other suites are unchanged.

| Required behavior | Result | Current evidence |
| --- | --- | --- |
| Successor-safe cleanup, different/identical resource IDs and retained consumers | PASS | Independent reads, admission and real shell held-response regression; old release/rejection cannot abort the successor. |
| Fresh same-ID saved-view reselection | PASS | Shell surfaces and session lifecycle: active equivalent intents coalesce; terminal reactivation freshly observes authorized configuration. |
| Distinct unavailable, operational, contract, unsupported and authority outcomes | PASS | Independent reads plus admission/locator cases; only eligible unavailability offers base fallback, abort has no failure notice, unsupported does not scan. |
| Captured Return configuration versus fresh comparison baseline | PASS | Admission Return fixtures restore captured query/layout against current authorized resource/version. |
| Exactly-once trail effects, fallback consumption and cancellation | PASS | Session/admission current-attempt assertions, failed/cancelled Return retention and admitted fallback consumption. |
| Accepted page → committed selection → grid focus → explicit inspector focus | PASS | Direct coordinator, real registry/semantic-grid integration, Note browser actual committed inspector control focus. Exhausted/rejected focus settles failure. |
| Delayed/replaced binding, unmount, Strict Mode and delivered input | PASS | Coordinator replacement/reconnect, query-controls Strict Mode and semantic focus; Columns browser click reaches its control while obsolete observation aborts. |
| Concealment, reauthorization, account/incident retirement and readable closure | PASS | Shell authority/retained operation and Commands rows in the full workbook gate; saved-view late-result retirement checks. |
| Retained drafts, operations, receipts and read-only acknowledged recovery | PASS | Production shell retained-work cases, Note associations and saved-view operation owner; navigation creates no writes or operation replacement. |
| Commands, bounded pins/trail, startup, ordinary entry and bounded query behavior | PASS | Commands/search, session limits, startup/preferences, query browsing/controls and full workbook gate. No extra query cache or page traversal. |

### Digest acceptance assessment

Every row of `acceptance.tsv` is assessed below for the changed navigation seam.
PASS means
the applicable regression boundary is satisfied, not fresh qualification of every
unmodified subsystem named by the general digest.

| Row | Assessment | Evidence / scope rationale |
| --- | --- | --- |
| A001 Authority | PASS | NR §§1/4 map exact adopted Core clauses; source placement and verification routing are separate. No advisory prose became behavior. |
| A002 Scope | PASS | NR §§3/5 and retirement ledger: one sequencing decision, owner handles, semantic variants, explicit extension path and no duplicated resource/page store. |
| A003 Repository state | PASS | Clean starting `main`/SHA recorded; current guides, authored manifests/import policies and generated policy inspected. September digest inventory is historical; changed files fully accounted above. |
| A004 Tokens | PASS | Diff audit: no CSS/design-token/theme/density literals or registry added. |
| A005 Theme | N/A | Theme producers, selection and fixtures are unchanged; no theme choice is exposed by the navigation correction. |
| A006 Density | N/A | No density, editor sizing, typography or row/header geometry change. No new measurement claim. |
| A007 Creation | PASS | Full workbook gate preserves startup/ordinary creation-first entry and source-owned creation capabilities; navigation adds no authoring dispatch. |
| A008 Responsive | PASS | Selected Columns cases at 1440×900, 1024×720 and 768×640 and Commands cases at 390×480/200% zoom preserve interaction and reachable controls. CSS accessors, thresholds, viewport fallback and resize algorithms unchanged. |
| A009 Overflow | PASS | Selected Columns/Commands browser checks retain reachable controls and shell geometry; no scroll container or shell overflow change. |
| A010 Inspector | PASS | Direct coordinator and Note browser current subject/focus pass; full workbook gate covers retained source/inspector detachment and authority. Feature dispatch unchanged. |
| A011 Continuity | PASS | Semantic-focus, direct coordinator and production-shell race cases pass; full workbook gate preserves raw authoring, receipts, selection and late-response continuity. |
| A012 Transactions | PASS | No transaction construction/replay changes; full workbook retained-write suites verify navigation does not submit, duplicate or replay owner requests. |
| A013 Acknowledgement/recovery | PASS | Saved-view operations and admission focused rows pass; full workbook Note/saved-view recovery proves acknowledged recovery remains read-only. |
| A014 Editing | PASS | One non-consuming cancellation call preserves deliberate input; browser Columns and full workbook editing/retained-draft cases cover affected lifetime transitions. |
| A015 Conflict | N/A | Cell conflict rendering/state is unchanged; no conflict-locus or retry-policy correction is part of this seam. |
| A016 Query/interaction states | PASS | Typed saved-view outcomes, query acceptance/rejection, preserved stale rows and admission failures pass; full workbook operational/interaction matrices remain routed. |
| A017 Refresh/authorization | PASS | Observation retirement/late results and shell uncertainty/account/incident cases; full workbook gate covers retained source authority without a generic clear-all. |
| A018 Evidence | N/A | No Evidence lifecycle, overlay or preview behavior is changed. Existing workbook Evidence suites remain in the full owner run. |
| A019 Accessibility | PASS | Required Commands accessibility browser selection, Note actual inspector focus and Columns interaction delivery pass; direct selection-before-focus and feedback tests retain keyboard semantics. No expanded conformance claim. |
| A020 Components | PASS | No component styling/variant/compound-state rule changes; affected menus, feedback and inspector attachment pass focused and browser cases. Zoom/text-spacing/density algorithms untouched. |
| A021 Virtualization | PASS | Semantic grid-focus and actual shell navigation use record/field identities, not vendor row indices; full workbook continuity remains required. No row-store, fake-row or virtualization algorithm change requiring performance claims. |
| A022 Visual fixtures | N/A | No visual fixture/golden/renderer styling change; no screenshot refresh or new visual comparison claim. Selected production browser geometry assertions pass. |
| A023 Selectors | PASS | Added assertions use semantic record/view identities and existing navigation/menu selectors. UI-contract package and generated selectors unchanged; no package.ui rerun needed for this diff. |
| A024 Test authority | PASS | New tests depend only on TypeScript fixtures/production capabilities and authored JSON routing. Diff/dependency review finds no Markdown reads, stats or hashes in executable inputs. |
| A025 Generated artifacts | PASS | Public `make generate` refreshed the catalog-derived render index; finalizer generation/drift transaction and generated policy check pass. No generated hand edits. |
| A026 Compatibility | PASS | Private atomic cutover, no aliases, feature flag, schema/wire/storage migration or new permission source. Retained capabilities and complete-slice rollback are explicit. |
| A027 Handoff | PASS | NR §10 records each ordered exit before its dependent; all five workstreams DONE, fresh gates pass, and NR §13 supplies complete accounting, acceptance, rollback and restart. Completed historical body was compared with HEAD and remains byte-identical. |

### Migration, rollback and restart

No persisted data, HTTP contract or storage migration is necessary. Compile all
private owner interfaces, adapters, callers and fixtures together. Roll back each
implementation slice coherently: observer/host/callers; intent/session/constructors;
coordinator/query bridge/bindings/tests. Never restore only the nullable facade or
retain two competing presentation algorithms. Revert authored source/test routing
with the corresponding source changes, then regenerate through Make and rerun
owner gates; preserve the execution evidence and failure history.

No new required debt was deferred. Cross-owner freshness consolidation, backend
qualification and visual redesign remain outside this iteration. Future consumers
must provide valid semantic targets and owner capabilities rather than adding
another shell sequence, cache or shared retention slot. To resume after a later
change, inspect current dirty state and owner guides, select affected rows from
NR §8, regenerate if authored machine inputs changed, then invalidate/rerun only
the evidence affected by that change. No historical pass establishes new readiness.

## Completed WB-W0–WB-04 remediation record

Completed iteration: **DONE — workbench legacy retirement and scoped production readiness**.
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
