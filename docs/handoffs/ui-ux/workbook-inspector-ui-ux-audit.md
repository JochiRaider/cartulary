# Workbook inspector remediation and production readiness tracker

## 1. Scope and source posture

**Controlling artifact:** this file. **Current iteration: implementation DONE.**
The user authorized WS-11–WS-18 on 2026-09-21. Execute the eight slices serially,
saving each entry and passing exit before starting its successor. Preserve the
staged planning update and previous-iteration evidence. WS-11–WS-18 are DONE;
the mandatory final validation and handoff exit is recorded below.

**Previous iteration: DONE**, including WS-00–WS-10 and their validation/handoff
checkpoints. The following historical statement applies only to that iteration:
the user approved implementation of WS-00 through WS-10, the reading-first design defaults, and a
coordinated server/browser History cutover without a legacy response adapter.
Execute slices serially; record each passing exit and update this tracker before
beginning the next slice. Planning evidence does not establish product readiness.

Rebuild the inspector around reading a selected record, examining its connections,
and taking a deliberate action. Use a bounded record header, persistent section
navigation, aligned saved-value rows, and local editing/action disclosure. Repair
History's missing semantic detail through its source owners as well as its UI.
Retire competing styling and presentation paths as their callers migrate.

The user requested durable remediation, not preservation of incidental behavior.
Retain a capability because it fulfills an adopted obligation or has a concrete
future benefit. Existing appearance, wording, helper structure and passing tests
are not compatibility requirements. The audit's original presentation-only
assumption is superseded for History completeness (G10).

### Authority and limits

- Adopted subsystem owners and Core 00–04 govern their named behavior. Core 05
  applies only to claim-bearing publication; this plan makes no such claim.
- [Design](../design.md) owns presentation direction; [Domain](../domain.md)
  owns vocabulary and owner navigation. Neither becomes a competing route,
  permission, lifecycle or storage authority.
- [NLSpec research](../research/nlspec-spec.md) supplies evaluation methods:
  independent-implementer equivalence, explicit defaults, complete mappings,
  omission/error behavior and binary completion criteria. Its agent-directed
  appendices are source material, not instructions to execute this plan or a
  newly adopted Cartulary specification. Its generic definition of a design
  document does not override Core 00's adoption of Cartulary's design owner.
- This audit, the [planning framework](cartulary_modular_refactor_planning_framework.md)
  and the UI/UX digest are planning/support material. Requirements proposed here
  enter their actual owner before implementation. Tests and generated projections
  never read, stat or hash Markdown to obtain product behavior.
- A genuine conflict between adopted owners is `BLOCKED: owner contradiction`;
  stop only the affected dependency chain. No contradiction was established for
  the proposed presentation changes. Missing History detail is an implementation
  and projection deficiency, not permission to weaken REQ-03-139.

Exact Core navigation: [status and precedence](../spec/00_document_set_status_and_precedence.md),
[view and History contracts](../spec/01_architecture_storage_and_view_contracts.md),
[source model and history](../spec/02_domain_model_schema_and_history.md),
[workbook interaction](../spec/03_workbook_interaction_collaboration_and_workflows.md),
and [security](../spec/04_security_deployment_and_conformance.md).

Previous-iteration baseline: `main` at `67f3d98ac01b42a1da4409c0a6630d08d24e1761`.
At this planning task's start, this audit was already staged as an added file;
there were no other changes. The original audit used seven retained screenshots,
source inspection and scenario definitions, without a live browser or user study.
This pass rechecked source and owners; it does not turn retained evidence into a
fresh execution result. Only this tracker may be changed during planning.

Target label: `workbook-inspector`. The target is the set of presentation and
History seams inventoried below, not every file under Workbook. No new theme,
inspector density preference, workflow engine, generic retry engine, persistent
draft storage, grid vendor integration, or richer Evidence enumeration is planned.

### Planning execution checkpoints

Update this tracker after completing each checkpoint and before starting the
next. These planning statuses are independent of the implementation tracker.

| Checkpoint | Status | Evidence / exit |
| --- | --- | --- |
| P-01 — Evidence and owner review | DONE | Current presentation, Core 00–04 seams, design/domain/research guidance, History materialization/OpenAPI, source boundaries and verification routing inspected; History completeness escalated into G10. Recorded before P-02 began. |
| P-02 — Remediation and workstream definition | DONE | G00–G11 cover all nine audit findings plus specification, History data and validation gaps; eleven implementation workstreams have owners, migrations, dependencies, risks and exits. Recorded before P-03 began. |
| P-03 — Validation and handoff completion | DONE | Planning review, four owner task guides, whitespace checks and Markdown lint passed at unchanged baseline; lint run `20260921T033650Z-p1857`. The user then approved implementation. |

### Next iteration — cleanup and production readiness

Current baseline: `8af6bba8737e52ad545b4425879c92cf45cfd1bb` (Workbook inspector
iter 2), verified with a clean worktree before this document update. The target
is the inspector subsystem, connected authoring/History seams, and the backend
History read path; this is not an application-wide release certification.
The user selected bounded server-side History paging for this iteration.

G12–G18 and WS-11–WS-18 govern the new plan. G00–G11, WS-00–WS-10, P-01–P-03,
RB-001–RB-009 and their completed checkpoints remain historical. The previous
iteration's implementation descriptions, exclusions, run receipts and A001–A027
PASS claims are not fresh evidence for this baseline. Current inventory and
acceptance below supersede historical inventories for the new work only.

The tracker-only restriction belonged to the completed planning task. The current
authorized implementation includes the named source, tests, owner clarifications,
authored contracts/routing, generated projections and supporting documentation.
No product PASS or readiness follows from the planning update. Each workstream
must supply its own actual validation evidence.
The archived validation handoff is outside this task and is not a dependency
or blocker, per the user's clarification; do not restore or rewrite it.

Carry forward canonical retained snapshots and conformant imports, semantic
History identities, all current source projectors, unfinished authoring, captured
request bytes and acknowledged receipts. Remove competing state representations,
unused compatibility inputs and unnecessary exports as their consumers migrate.
Prefer cohesive owners with small interfaces over a proliferation of wrappers.
Future source families must extend existing catalogs and presentation ports,
not central source switches or another form/retry engine.

Preserve the adopted reading-first UX and its existing tokens, declared section
order and field identities. No new theme, durable drafts, richer Evidence
catalog, unrelated platform refactor or speculative domain feature is included.
Core 02 REQ-02-265 still excludes schema-less decoding, historical-shape inference
and backfill. No automatic database reset or history rewrite is authorized.
Additive indexes may be planned in WS-15 only when query evidence justifies them.

| Document checkpoint | Status | Exit / evidence |
| --- | --- | --- |
| P-04 — Current baseline, inventory and decisions | DONE | Clean baseline and all 74 target files accounted for; source/owner/routing inspection and user-selected subsystem/paging scope recorded. |
| P-05 — Tracker update and documentation checks | DONE | G12–G18, WS-11–18 TODO, acceptance, retirement, risks and handoff saved; public Markdown run `20260921T133544Z-p55519` and explicit tracker lint passed; whitespace, 74-file inventory, 62 historical WS rows and single-file diff verified. |

## 2. Current-state repository inventory

### Next-iteration inventory at 8af6bba8

The exhaustive target is `apps/web/src/workbook/inspector/` (52 files) and
`apps/web/src/workbook/history/` (22 files). The inventory below accounts for
all 74: 50 TypeScript files without a test suffix, 21 test files and three
README files. `history/workbookHistoryTestFixtures.ts` is one of the 50 by
filename classification but has only test consumers; it is not product behavior.

Paths in the first table are relative to `apps/web/src/workbook/`. Every file
was read for its declarations/imports or test surface. The exact behavior-changing
seams were inspected directly; the remaining rows are accounting and retention
dispositions, not claims of exhaustive behavioral audit. Consumer counts are
static relative-import/re-export references under `apps/web/src`; zero is not
proof of dead code. WS-11 and WS-17 must also reconcile exported symbols, aliases,
registration, package boundaries and tests before removal.

| Group | Responsibility / target owner | Outbound dependencies and contract owner | Existing validation neighbors |
| --- | --- | --- | --- |
| P | Shared inspector presentation and semantic focus | UI/view contracts, layout navigation, owner-supplied sections/actions; design §12.7 and Core 01 REQ-01-615–617 | Presentation/panel unit tests, inspector browser and accessibility rows |
| D | Ordinary drafts, editing, retained observations and field feedback | Query/operation projections and explicit draft owner; Core 03 §2.3A, REQ-03-299/100 | Draft store/binding, retained-row and field-feedback tests; Timeline/Generic/Entity recovery |
| H | History read, operation and presentation binding | History adapter, immutable source units, scope/selector contracts; Core 01 REQ-01-052A/054 and Core 02 REQ-02-216/218/265 | Browsing, owner, recovery, History presentation and Revisions tests |
| C | Canonical capability, subject and source workflow composition | View contracts and feature command owners; Core 01 §7.4 and Core 03 workflow/security owners | Capability, coordinator, related workflow and specialized-family tests |
| T | Test evidence and fixtures | Production public interfaces, existing test helpers and authored test routing | Owning unit suites; no production import from test support |
| N | Source guidance | Human-owned boundary explanations; no runtime dependency on Markdown | Markdown lint and inventory reconciliation |

Source ownership is `web.workbook`; verification routes are separately resolved
through §8. Risk is high for state/admission/recovery and medium for presentation
or interface cleanup. A test row marked migrate preserves intended behavior,
not obsolete reducer phases or prop arrangements.

| Target file | Responsibility / public surface | Static inbound references | Group | Disposition | Risk |
| --- | --- | --- | --- | --- | --- |
| `history/HistoryActionLookup.test.ts` | Tests bounded History action lookup, explicit continuation, and failed-page retry | owning test runner | T | keep / migrate assertions | medium |
| `history/HistoryActionLookup.ts` | Incremental paged lookup for a retained History action without treating paused scans as absence | 4 source / 1 test files | H | keep; migrate History interface in WS-12 | high |
| `history/HistoryLookupFeedback.tsx` | History action lookup progress, paused-search, failure, and continuation feedback | 4 source / 0 test files | H | keep; migrate History interface in WS-12 | high |
| `history/HistoryPageLookup.ts` | Shared bounded read search, page validation, cancellation and authority fencing | 3 source / 1 test files | H | keep; migrate History interface in WS-12 | high |
| `history/README.md` | Human source-boundary and consumer guide | human readers | N | keep; update source guidance | low |
| `history/WorkbookBatchHistoryReview.test.tsx` | Explicit off-window review, pruning lifetime, current action gates and existing reversal ownership | owning test runner | T | keep / migrate assertions | medium |
| `history/WorkbookHistoryContext.ts` | React access to History runtime, surface refresh, action permissions, and pending records | 13 source / 5 test files | H | keep; migrate History interface in WS-12 | high |
| `history/WorkbookHistoryLocalStatus.tsx` | Record-local History operation status | 1 source / 0 test files | H | keep; migrate History interface in WS-12 | high |
| `history/WorkbookHistoryRecovery.test.tsx` | Tests retained acknowledgements, read-only refresh retry, and session-bound recovery visibility | owning test runner | T | keep / migrate assertions | medium |
| `history/WorkbookHistoryRecovery.tsx` | Retained History operation replay, lookup, and reconciliation recovery presentation | 1 source / 1 test files | H | keep; migrate History interface in WS-12 | high |
| `history/WorkbookHistoryReview.tsx` | Explicit Recovery review using current row History and existing confirmed actions | 1 source / 0 test files | H | keep; migrate History interface in WS-12 | high |
| `history/WorkbookRecordHistoryOwner.test.ts` | Tests History authority lifetime, retained operations after detachment, and late effects | owning test runner | T | keep / migrate assertions | medium |
| `history/WorkbookRecordHistoryOwner.ts` | Record History operation admission, captured attempts, acknowledgement, and recovery ownership | 8 source / 10 test files | H | keep; migrate History interface in WS-12 | high |
| `history/historyOperationPresentation.ts` | Projects History operation state into user-facing status | 2 source / 0 test files | H | keep; migrate History interface in WS-12 | high |
| `history/workbookHistoryBrowsing.characterization.test.tsx` | Tests server-owned continuation and retained accepted History after refresh failure | owning test runner | T | keep / migrate assertions | medium |
| `history/workbookHistoryBrowsing.test.ts` | Tests short/empty page continuation, stable server ordering, deduplication, and eligibility refresh | owning test runner | T | keep / migrate assertions | medium |
| `history/workbookHistoryBrowsing.ts` | Pure History browsing state transitions for initial, continuation, and refresh reads | 2 source / 1 test files | H | keep; migrate History interface in WS-12 | high |
| `history/workbookHistoryItem.ts` | Normalizes History items and builds stable rollback targets and pending action identities | 5 source / 0 test files | H | keep; migrate History interface in WS-12 | high |
| `history/workbookHistoryOperation.ts` | History authority, intents, captured attempts, receipts, semantic ports, and bindings | 9 source / 7 test files | H | keep; migrate History interface in WS-12 | high |
| `history/workbookHistoryPage.ts` | History page requests, provenance, scope equality, and paging validation | 9 source / 3 test files | H | keep; migrate History interface in WS-12 | high |
| `history/workbookHistoryReview.ts` | Minimal authorized read locator and three-page review policy, distinct from pending actions | 3 source / 0 test files | H | keep; migrate History interface in WS-12 | high |
| `history/workbookHistoryTestFixtures.ts` | historyDiffFixture; thirteen test-only consumers | 0 source / 13 test files | T | move to existing app test support; G13 | medium |
| `inspector/InspectorCreateRelatedWorkflow.tsx` | Inspector related-record authoring presentation over declared feature commands | 4 source / 0 test files | C | keep; reconcile public surface | high |
| `inspector/README.md` | Human source-boundary and consumer guide | human readers | N | keep; update source guidance | low |
| `inspector/WorkbookExplicitPatchRecovery.tsx` | Explicit patch recovery after inspector detachment | 5 source / 0 test files | C | keep; reconcile public surface | high |
| `inspector/WorkbookInspectorContextualActions.tsx` | Renders admitted contextual inspector actions from canonical feature declarations | 1 source / 2 test files | P | keep / split presentation in WS-13 | medium |
| `inspector/WorkbookInspectorDeclaredPanelList.tsx` | Renders declared inspector panels in their owner-defined order | 4 source / 1 test files | P | keep / split presentation in WS-13 | medium |
| `inspector/WorkbookInspectorDetails.tsx` | Saved fields plus editable and read-only Details facades | 4 source / 1 test files | P | split reading from editing; G14 | medium |
| `inspector/WorkbookInspectorDraftFeedback.tsx` | Local Resume/Discard and changed saved-value review | 3 source / 0 test files | D | keep source lifetime; reconcile surface | high |
| `inspector/WorkbookInspectorDraftStore.test.ts` | Draft, capability, review, security and captured-revision regressions | owning test runner | T | keep / migrate assertions | medium |
| `inspector/WorkbookInspectorDraftStore.ts` | Account/incident-scoped raw drafts, explicit attachment, dependency review and revision-fenced retirement | 4 source / 5 test files | D | keep source lifetime; reconcile surface | high |
| `inspector/WorkbookInspectorEditControl.tsx` | Accessible ordinary value control with retained reference identity and explicit clear intent | 3 source / 2 test files | D | keep source lifetime; reconcile surface | high |
| `inspector/WorkbookInspectorRecordHistory.test.tsx` | Tests advertised History loading, stable rollback selectors, and exact tombstone restore versions | owning test runner | T | keep / migrate assertions | medium |
| `inspector/WorkbookInspectorRecordHistory.tsx` | Record History inspector composition and browsing controls | 8 source / 2 test files | H | keep; migrate History interface in WS-12 | high |
| `inspector/WorkbookRecordHistoryPresentation.tsx` | Loaded record History presentation with action and pending-state bindings | 1 source / 0 test files | H | keep; migrate History interface in WS-12 | high |
| `inspector/canonicalInspectorAdmission.ts` | Admits exact canonical inspector features and derives stable feature identity | 3 source / 1 test files | C | keep; reconcile public surface | high |
| `inspector/inspectorCapabilityResolver.test.ts` | Tests exact canonical schema/feature tuples and contextual or History action classification | owning test runner | T | keep / migrate assertions | medium |
| `inspector/inspectorCapabilityResolver.ts` | Resolves canonical inspector capabilities and supported record History actions | 17 source / 5 test files | C | keep; reconcile public surface | high |
| `inspector/inspectorRelatedRecordModel.test.ts` | Tests related-record seed construction and characterized workflow draft transitions | owning test runner | T | keep / migrate assertions | medium |
| `inspector/inspectorRelatedRecordModel.ts` | Related-record form seeds and pure inspector workflow state transitions | 11 source / 1 test files | C | keep; reconcile public surface | high |
| `inspector/prepareWorkbookInspectorChange.ts` | Patch capability and scalar/action serialization; creation remains separately admitted | 2 source / 2 test files | D | keep source lifetime; reconcile surface | high |
| `inspector/presentation/README.md` | Human source-boundary and consumer guide | human readers | N | keep; update source guidance | low |
| `inspector/presentation/WorkbookHistoryPresentation.tsx` | Shared History list and event rendering from presentation models | 1 source / 1 test files | H | keep; migrate History interface in WS-12 | high |
| `inspector/presentation/WorkbookInspectorActions.tsx` | Inspector action groups, contextual actions, and accessible action buttons | 64 source / 1 test files | P | keep / split presentation in WS-13 | medium |
| `inspector/presentation/WorkbookInspectorFeedback.tsx` | Inspector metadata, technical details, safe errors, feedback, and confirmation presentation | 21 source / 2 test files | P | keep / split presentation in WS-13 | medium |
| `inspector/presentation/WorkbookInspectorPanelContent.test.tsx` | Test evidence: Inspector panel states | owning test runner | T | keep / migrate assertions | medium |
| `inspector/presentation/WorkbookInspectorPanelContent.tsx` | Public surface: WorkbookInspectorPanelData, WorkbookInspectorPanelContentModel, inspectorReadData | 15 source / 5 test files | P | keep / split presentation in WS-13 | medium |
| `inspector/presentation/WorkbookInspectorPresentation.test.tsx` | Tests valid subject boundaries, ordered panels, deleted-record History, and explicit creation content | owning test runner | T | keep / migrate assertions | medium |
| `inspector/presentation/WorkbookInspectorShell.tsx` | Shared inspector shell and declared panel-section layout | 5 source / 1 test files | P | close saved/empty/creation variants; G14 | medium |
| `inspector/presentation/workbookInspectorPresentationModel.ts` | Inspector action bindings, disabled reasons, technical fields, and History event presentation types | 14 source / 1 test files | P | keep / split presentation in WS-13 | medium |
| `inspector/useInspectorCreateRelatedWorkflow.test.tsx` | Tests neutral related-creation feedback and detached late-result suppression | owning test runner | T | keep / migrate assertions | medium |
| `inspector/useInspectorCreateRelatedWorkflow.ts` | React controller for declared related-record authoring and presentation lifetime | 3 source / 2 test files | C | keep; reconcile public surface | high |
| `inspector/useRetainedInspectorRow.test.ts` | Window eviction, version and authority-retirement cases | owning test runner | T | keep / migrate assertions | medium |
| `inspector/useRetainedInspectorRow.ts` | Retained source observation admission and version comparison | 3 source / 1 test files | D | replace identity retirement with provenance; G15 | high |
| `inspector/useWorkbookInspectorCoordinator.test.tsx` | Direct tests for retargeting, lifecycle invalidation, action completion, idempotent close, and focus restoration | owning test runner | T | keep / migrate assertions | medium |
| `inspector/useWorkbookInspectorCoordinator.ts` | Schema-bound inspector lifecycle coordinator for explicit open/close, stable row subjects, ordered feature invalidation, action completion, and focus restoration ports | 4 source / 1 test files | C | keep; reconcile public surface | high |
| `inspector/useWorkbookInspectorEditDraft.test.tsx` | Frozen field/action binding through refresh, detachment, no-row and explicit return | owning test runner | T | keep / migrate assertions | medium |
| `inspector/useWorkbookInspectorEditDraft.ts` | Canonical field/action binding and presentation detachment, independent from query object identity | 8 source / 3 test files | D | keep source lifetime; reconcile surface | high |
| `inspector/useWorkbookInspectorFieldFeedback.ts` | Captured field/action feedback binding | 3 source / 1 test files | D | keep source lifetime; reconcile surface | high |
| `inspector/useWorkbookInspectorNotice.ts` | Owner-ledger announcement admission | 1 source / 0 test files | C | keep; reconcile public surface | high |
| `inspector/useWorkbookRecordHistoryController.ts` | Binds shared History reads/actions to the canonical Inspector subject or an explicit Recovery read locator and the existing History owner | 5 source / 2 test files | H | split binding from read authority; G12 | high |
| `inspector/useWorkbookRecordHistoryFocus.ts` | History action focus coordination using stable action and rollback identities | 2 source / 0 test files | H | keep; migrate History interface in WS-12 | high |
| `inspector/workbookHistoryPresentationModel.test.ts` | Tests consistent History technical-field ordering and rollback/pending labels | owning test runner | T | keep / migrate assertions | medium |
| `inspector/workbookHistoryPresentationModel.ts` | Builds History event, rollback-label, and pending-operation presentation | 1 source / 2 test files | H | keep; migrate History interface in WS-12 | high |
| `inspector/workbookHistoryRecovery.characterization.test.tsx` | Tests acknowledgement before refresh completion and selection-lifetime fencing of admitted effects | owning test runner | T | keep / migrate assertions | medium |
| `inspector/workbookInspectorErrorModel.test.ts` | Tests safe primary messages, decoded version-conflict identity, and sanitized details | owning test runner | T | keep / migrate assertions | medium |
| `inspector/workbookInspectorErrorModel.ts` | Shared inspector error and operation-feedback presentation models | 39 source / 3 test files | C | keep; reconcile public surface | high |
| `inspector/workbookInspectorFieldFeedback.test.ts` | Captured edit/revision feedback isolation | owning test runner | T | keep / migrate assertions | medium |
| `inspector/workbookInspectorFieldFeedback.ts` | Pure captured feedback identity and matching | 1 source / 1 test files | D | keep source lifetime; reconcile surface | high |
| `inspector/workbookInspectorSubject.ts` | Constructs and compares canonical live/deleted inspector subjects and row bindings | 39 source / 5 test files | C | keep; reconcile public surface | high |
| `inspector/workbookRecordHistoryModel.test.ts` | Tests stable subject references and rejection of events from obsolete phases or identities | owning test runner | T | keep / migrate assertions | medium |
| `inspector/workbookRecordHistoryModel.ts` | Inspector History state machine keyed by subject, read request, and operation identity | 17 source / 13 test files | H | split; remove duplicate read state; G12 | high |
| `inspector/workbookRecordHistoryOperation.ts` | Builds inspector feedback for completed record History operations | 1 source / 0 test files | H | keep; migrate History interface in WS-12 | high |
| `inspector/workbookRecordHistoryOwnerEffects.ts` | Coordinates History owner effects with inspector subject and presentation lifetime | 5 source / 0 test files | H | keep; migrate History interface in WS-12 | high |

### Next-iteration connected seams and exclusions

Paths below are repository-relative; brace lists identify the exact named files.
They extend the target only for the responsibility stated, not their entire
containing subsystem. Backend source semantics and operation recovery remain
owned by existing providers and feature owners.

| Connected paths | Inbound / outbound boundary | Tests and projections | Disposition / risk |
| --- | --- | --- | --- |
| `apps/web/src/workbook/timeline/hooks/{useTimelineHistoryState.ts,useTimelineHistoryActions.ts,useTimelineInspectorSelection.ts}` | Timeline binds inspector state, canonical rows and mutation coordination | Timeline History coordination and inspector retention rows | Migrate in WS-12/14; retain FIFO operation coordination; high |
| `apps/web/src/workbook/features/generic/{GenericWorkbookInspector.tsx,GenericWorkbookInspectorPresentation.tsx,useGenericWorkbookInspectorComposition.tsx}` | Generic query/operation owners supply subjects, sections and editor commands | Generic field/creation/recovery rows and view contracts | Migrate WS-13/14; high |
| `apps/web/src/workbook/features/entities/{EntityWorkbookInspector.tsx,useEntityWorkbookInspectorComposition.tsx,WorkbookEntityMergeRecovery.tsx}`; `apps/web/src/workbook/components/EntityWorkbookSurface.tsx` | Entity source, Details, converted committed rows and merge recovery consume shared contracts | Entity edit/merge/access tests | Migrate WS-12–14; preserve merge receipts; high |
| `apps/web/src/workbook/features/assessments/{AssessmentWorkbookInspector.tsx,useAssessmentWorkbookInspectorComposition.tsx}`; `apps/web/src/workbook/timeline/components/{TimelineWorkbookInspector.tsx,TimelineInspectorDetails.tsx}` | Append-only Assessment and Timeline consume shell/Details | Assessment append and Timeline explicit-edit matrices | Migrate WS-13; preserve distinct source lifetimes; high |
| `apps/web/src/workbook/adapters/{createWorkbookRecordHistoryAdapter.ts,createWorkbookRecordHistoryAdapter.test.ts}`; `apps/web/src/workbook/timeline/adapters/createTimelineActionAdapters.test.ts` | HTTP validation supplies typed History pages to History owners | Generated protocol and semantic response fixtures | Keep boundary validation; migrate page types/test fixtures; high |
| `apps/web/src/workbook/layout/{WorkbookSurfaceLayout.tsx,workbookInspectorNavigation.ts}`; `apps/web/src/workbook/hooks/useWorkbookQueryController.ts`; `apps/web/src/workbook/view-state/useWorkbookQueryState.ts` | Geometry/navigation and query observations stay separate from inspector retained rows | Geometry, subject retention, query and keyboard rows | Inspect source provenance at WS-14; preserve established owners; high |
| `apps/web/src/testing/workbookInspectorTestSupport.ts`; `apps/web/src/workbook/policies/workbookSurfaceOwnershipPolicy.test.ts`; `packages/test-utils/` | Existing app test support, shared semantic helpers and source policy | Unit/browser consumers and import-boundary rules | Move app-specific History fixture to app test support; do not move production dependencies into package test-utils; medium |
| `internal/modules/revisions/{application_ports.go,command_service.go,history_service.go,history_repository.go,history_materializer.go,history_model.go,history_attribution.go,history_rollback_actions.go}` | HTTP calls Revisions; Revisions coordinates retained facts, projectors, attribution and legal actions | History component/integration/import/rollback tests | Typed page query/result and bounded materialization in WS-15; high |
| `internal/modules/revisions/httpapi/{routes.go,history_pagination.go}`; `internal/platform/pagination/` | HTTP owns cursor encoding, request validation and envelope serialization | Record-bound paging and malformed-cursor tests | Replace full-list selection, keep generic codec ownership; high |
| `internal/modules/revisions/{history_projection.go,target_semantics_catalog.go,provider_contributions.go}`; `internal/app/revisionassembly/` | Source-owned pure projectors enter application-composed catalog | Ten record snapshots, fourteen mutation targets, nine source owners | Keep extension boundary; no central source switch; high |
| `db/migrations/{00007_records.sql,00008_revisions.sql}`; `db/queries/` | Existing history associations, selector storage and indexes support Revisions reads | Database migration policy and query-plan evidence | Inspect existing indexes; any additive index is a later authored migration, not an edit to applied history; high |
| `contracts/openapi-source/owners/module.revisions/openapi.json`; `contracts/design/presentation.v1.json`; `contracts/view-schemas/`; `packages/protocol-ts/src/generated/`; `packages/ui-contracts/src/generated/` | Owner projections and generated facades supply public fields, roles and schemas | Generation/drift/JSON-shape/OpenAPI compatibility | Preserve semantic transport and all 17 configurations; generated roots read-only; high |
| `tools/{frontend_source_ownership.json,frontend_import_boundaries.json,generated_artifact_policy.json,frontend_visual_fixture_registry.json,test_catalog_owner.json}`; `contracts/verification/`; `tools/test_families/{web.workbook,module.workbook,module.revisions,web.architecture}.json` | Source boundaries and independent verification routes | Public Make task guides, test slices and visual reconciliation | Migrate authored routing only when implementation requires it; medium |
| `apps/web/e2e/`; specialized Note/coordination/Task/Decision/Indicator/Evidence forms | Existing browser choreography and source-local authoring neighbors | Recovery, focus, accessibility and per-family scenario rows | Inspect only changed consumers or reproduced failures; no blanket form rewrite; high |

**Excluded:** unrelated application/platform release work; grid vendor internals;
Evidence blob storage and richer enumeration; unrelated Network Flow production
changes; all other source-provider behavior; global migration lifecycle changes;
archived handoff recovery; and speculative future schema families. Existing
related-record workflows and source recovery are retained even when their names
or tests mention legacy behavior. A live caller is evidence to migrate or retain,
not a justification for an indefinite adapter.

### Previous-iteration section 2 record (historical)

Paths below are repository-relative. Each row accounts for the named seam;
paths grouped in a row share its responsibility and risk. Source ownership is
`web.workbook` unless specified. Verification ownership is separately listed in
§8. Inspected code establishes structure, not a passing runtime baseline.

| Path | Responsibility / exported surface | Inbound callers | Outbound dependencies | Tests / contracts | Target owner and risk |
| --- | --- | --- | --- | --- | --- |
| `apps/web/src/workbook/inspector/presentation/WorkbookInspectorShell.tsx` | `WorkbookInspectorShell`, `WorkbookInspectorPanelSection`; record header, body, section headings | Timeline, Generic, Entity, Assessment compositions | UI/view contracts, layout, subject and presentation types | `WorkbookInspectorPresentation.test.tsx`; design §§7.3,12.7 | Shared inspector presentation; high focus/concealment risk |
| `apps/web/src/workbook/inspector/WorkbookInspectorDeclaredPanelList.tsx` | Ordered admitted panels and contextual actions | Feature inspector compositions | Capability resolver, panel models, section renderer | Presentation tests; Core 01 REQ-01-615–617 | Inspector composition; high coverage risk |
| `apps/web/src/workbook/inspector/presentation/WorkbookInspectorPanelContent.tsx` | Explicit data/access regions, owner delivery slots | All declared panel producers | Typed design states, owner-supplied content/notices | `WorkbookInspectorPanelContent.test.tsx`; `contracts/design/presentation.v1.json` | Keep shared state rendering; high lifetime/security risk |
| `apps/web/src/workbook/layout/WorkbookSurfaceLayout.tsx` | Adjacent/overlay composition, resize and return focus | Workbook surfaces | Existing layout metrics and overlay host | Browser geometry and inspector fixtures; layout tokens | Layout owner; integrate navigation without moving its geometry policy |
| `apps/web/src/workbook/inspector/WorkbookInspectorDetails.tsx`, `WorkbookInspectorEditControl.tsx`, `WorkbookInspectorDraftFeedback.tsx` | Accepted field rows, attached controls, draft/review feedback | Timeline Details, Generic/Entity editing | View contracts, draft port, reference and mutation controls | `TimelineInspectorDetails.test.tsx`, draft tests, inspector edit browser rows | Shared Details presentation; high authoring risk |
| `apps/web/src/workbook/inspector/presentation/WorkbookInspectorActions.tsx` and `apps/web/src/workbook/components/workbookFormStyles.ts` | Buttons, action groups, shared form roles | Inspector and specialized forms | Authored UI tokens, typed disabled reasons | Presentation tests and `package.ui` tests | Shared presentation; medium cross-consumer visual risk |
| `apps/web/src/workbook/timeline/components/TimelineMentionsPanel.tsx`, `TimelineMentionActionControls.tsx` | Source-local mention groups, selected controls, target filtering | Timeline inspector | Existing mention operation owner, candidate picker, semantic focus registry | Timeline relationship/mention regressions | Timeline presentation; high identity/recovery risk |
| `apps/web/src/workbook/timeline/components/TimelineEvidencePanel.tsx` | Count, file chooser, drop/paste and retained recovery | Timeline inspector | Evidence attachment context and file owner | `TimelineEvidencePanel.test.tsx`; Evidence owner contracts | Timeline/Evidence presentation; high security and duplicate-write risk |
| `apps/web/src/workbook/timeline/components/TimelineWorkbookStyles.ts` | Shared Timeline style exports used by inspector and other callers | Timeline components | Design variables plus local literals | Density/visual evidence | Migrate inspector callers; do not delete exports still used by grid or other surfaces |
| `apps/web/src/workbook/inspector/workbookHistoryPresentationModel.ts`, `WorkbookRecordHistoryPresentation.tsx`, `presentation/WorkbookHistoryPresentation.tsx` | Event mapping, event/reversal/delete/restore presentation | Inspector History composition | History action permissions, opaque selectors, safe technical fields | `workbookHistoryPresentationModel.test.ts`, History recovery tests | History-to-presentation adapter and stateless rendering; high review-scope risk |
| `apps/web/src/workbook/history/workbookHistoryItem.ts` | History normalization, selector construction, committed-content equality | History browsing and action owners | Protocol history model | History browsing/recovery tests | Keep History semantics here; high rollout/cache-epoch risk |
| `apps/web/src/workbook/features/generic/GenericWorkbookInspectorPresentation.tsx`, `features/entities/EntityWorkbookInspector.tsx`, `features/assessments/AssessmentWorkbookInspector.tsx` | Surface-specific region and editor composition | Respective feature surfaces | Shared inspector, feature-owned commands | Generic/Entity/Assessment tests and browser flows | Feature composition; medium presentation, high source-context risk |
| `apps/web/src/workbook/inspector/InspectorCreateRelatedWorkflow.tsx` | Dispatch into Note, coordination, contextual and related-Evidence forms | Inspector feature compositions | Existing creation contexts, capability metadata | Related-creation tests and specialized browser scenarios | Keep dispatch in composition; never move it into stateless presentation |
| `apps/web/src/workbook/features/indicators/IndicatorLifecycleWorkflow.tsx`, `indicatorLifecycleStyles.ts` | Lifecycle controls, retained draft binding, local styling | Indicator feature contribution | Lifecycle owner, paging, feedback and support picker | Lifecycle workflow/reconciliation tests | Share visuals only; retain feature semantics |
| `internal/modules/revisions/history_materializer.go`, `history_service.go`, `history_model.go` | Materialize and serialize row-centric History | Revisions History HTTP adapter | Retained mutations/revisions, attribution, rollback availability | `history_test.go`, `history_components_test.go`; Revisions OpenAPI | Revisions coordination; high disclosure/compatibility risk |
| `internal/modules/revisions/provider_contributions.go`, `target_semantics_catalog.go`; `internal/app/revisionassembly/` | Source-owned provider catalog and application composition | Revisions assembly | Source-owner providers, generated target requirements | Catalog and source-boundary tests | Revisions owns catalog validation; source modules own semantic history units |
| `contracts/openapi-source/owners/module.revisions/openapi.json`; `packages/protocol-ts/src/generated/core-http-types.ts` | Authored History transport projection and generated browser types | OpenAPI generation and History adapters | Owner requirements | `RecordHistoryDiffSummary.units` is currently open objects | Edit authored source only; generated types inspected read-only |
| `contracts/design/tokens.v1.json`, `presentation.v1.json`; `contracts/view-schemas/cartulary.view.timeline.v2.json` | Tokens, presentation states, immutable schema fields/panels | UI/view contract facades and workbook consumers | Adopted design/Core owners | UI contract/generation checks | `package.ui` / view projections; avoid new field or panel registries |
| `apps/web/e2e/workbook.visual.spec.ts`, `workbook-inspector-edit.spec.ts`; `tools/frontend_visual_fixture_registry.json` | Production-renderer visual and behavioral evidence | Browser harness | Semantic fixtures, selectors, declared capture parameters | Public browser targets | Test owners; missing first-viewport Details capture is G11 |
| `tools/frontend_source_ownership.json`, `frontend_import_boundaries.json`, `generated_artifact_policy.json`, `test_families/{web.workbook,module.workbook,module.revisions,package.ui}.json`, `task_surface_owner.json` | Source/import ownership, generation boundaries and independent routing | Repo harness and generators | Authored machine inputs | Public investigation/generation targets | Harness owners; do not infer requirements from routing |

The inspector, presentation and History README guides were read as navigation.
The package manifests confirm the existing React, UI/view/protocol package and
Lucide boundaries; no dependency or framework change is justified.

**Explicit exclusions:** remaining inspector controllers/stores, feature operation
owners, source-provider implementations, adapters, grid-adapter internals,
Evidence storage/access internals, unrelated modules and tests are regression
neighbors, not exhaustively audited targets. Their guides/imports/catalog rows
were used for navigation where stated. No file move is proposed without reading
its implementation and consumers in its workstream. WS-06 must enumerate every
participating source provider before changing its interface. Existing seven-image
review and its limitations remain historical; no images were recaptured here.

## 3. Module boundary diagnosis

### Next-iteration owner boundary decisions

| Responsibility | Current location / evidence | Correct owner | Disposition | Extension consequence |
| --- | --- | --- | --- | --- |
| Accepted History pages, provenance and read failures | Optional browsing plus repeated data in inspector phase union | `workbook/history` browsing model | split G12 | New History consumers share one read contract, without importing an inspector reducer. |
| Captured attempts, acknowledgement and replay | `WorkbookRecordHistoryOwner` and source operation owners | Existing retained operation owners | keep | A new view cannot shorten an operation lifetime. |
| Subject attachment, review disclosure and focus | Inspector controller/presentation, Timeline external state injection | Inspector binding and existing layout/focus owners | split G12/G14 | Eliminate parallel reducers while preserving Timeline transaction coordination. |
| Saved fields versus edit attachment | Details imports the draft hook's return type and read-only uses no-op commands | Pure presentation contract supplied by draft owner | split G14 | Append-only/read-only sources need no fake editing capability. |
| Accepted row provenance | `useRetainedInspectorRow` retires object references | Existing query/mutation observation owner, consumed by inspector | move admission evidence; G15 | Cloning or conversion cannot establish fresh authority. |
| Logical History page selection | Repository loads all facts; HTTP slices serialized maps | Revisions typed page application contract; HTTP retains cursor codec | split G16 | Growth adds source projectors, not unbounded materialization or a second history table. |
| Test-only fixture and incidental exports | Fixture beneath production root; external-use scan | Existing app test support; private implementation modules | move / privatize G13 | Tests do not enlarge the production facade. |
| Source policy, test routing and evidence | Authored ownership/catalog inputs and historical tracker | Existing verification owners; this tracker records outcomes | keep G17/G18 | No runtime dependency on planning phases or Markdown. |

These are internal structural changes except the explicit provenance correction
and disposable-cursor transition. Their intended behavior is defined in §4/§7;
all production execution remains a later task. No owner contradiction has been
established. Tests that assert old file placement or reducer phases must migrate
to the intended boundary without dropping useful behavioral coverage.

### Previous-iteration section 3 record (historical)

The inspector is a presentation/composition seam, not a new domain module. A
shared interface is justified when it hides the decision about presenting and
reaching an owner-supplied subject, section, field, action or state. It must not
absorb the decision about permission, request admission, domain validation or
operation lifetime.

| Responsibility | Current location | Owner candidate | Disposition | Evidence and design consequence |
| --- | --- | --- | --- | --- |
| Record frame, typography, accessible section navigation | Shared shell and layout | Inspector presentation, with geometry retained in layout | keep / extend | Shell already isolates header and one scroll body; derive one admitted section descriptor sequence for both navigation and body. |
| Panel access and contribution admission | Declared list and explicit region model | Inspector composition plus feature producers | keep | Concealed unions have no payload; reuse that decision, never scan React children or register a second permission map. |
| Ordinary field presentation and command placement | Details plus injected editor children | Shared Details presentation | split | Give editor content and owner-supplied actions explicit slots; group Update/Close without introspecting children or owning drafts. |
| Styling decisions repeated across forms | Timeline, Indicator, shared styles | Existing shared form/presentation boundary | move | Migrate actual callers and remove dead local alternatives; no generic form engine. |
| Source-specific authoring, retries and receipts | Feature/runtime owners | Existing source owners | keep | Their different validation, review and security lifetimes are material behavior, not duplicate visual machinery. |
| History semantic interpretation | Coarse Revisions materializer and frontend string presentation | Source-owned pure history contributions; Revisions coordinates; frontend renders typed facts | split | Core 02 REQ-02-218 prohibits central source-type logic; the current provider catalog is the extension seam. |
| Rich Evidence enumeration | Count plus separate owner access contributions | Evidence source/read owner | defer | Count does not supply item metadata or grant access. No new enumeration feature is needed to close this presentation plan. |
| Evidence accounting | Test catalog/fixture inputs | Current verification/harness owners | keep | Add missing scenario coverage through authored routing; no documentation-driven runtime tests. |

A new surface should supply subject, admitted ordered regions, fields and command
presentation through existing composition. It should need neither a new shell
copy nor a surface-name switch in shared rendering. A new history source supplies
its own semantic projector through a checked catalog; it does not add storage
knowledge to the frontend or a central Revisions switch.

| Coupling / boundary finding | Classification | Owner / required action |
| --- | --- | --- |
| Repeated local styling prevents complete role propagation | must_fix | G06; consolidate presentation decisions and migrate real callers. |
| Shared presentation already forbids feature/controller/mutation imports | intentional/no_action | Retain the existing import guard; navigation adds local presentation state only. |
| History units are untyped and the renderer loses semantic detail | must_fix | G10/G04; source-owned typed projection, Revisions coordination and a pure display adapter. |
| A universal form or retry engine would merge different owner lifetimes | intentional/no_action | Retain separate feature operation owners; share visual roles, not domain policy. |
| Rich Evidence enumeration lacks an established data contribution in the audited count panel | defer | G09/O-01; do not create a new read dependency to beautify a count. |
| Framework snapshots omit some current generated boundaries | must_fix | Planning correction: current generated policy also protects view-contract generated files and aggregate OpenAPI; use it rather than prose inventories. |

## 4. Public contract and behavior disposition map

### Next-iteration contract and behavior disposition

| Contract | Adopted owner | Current evidence | Required validation / disposition |
| --- | --- | --- | --- |
| Complete semantic History and representation generation | Core 01 REQ-01-052A; Core 02 REQ-02-218 | Closed generated types, adapter validator and source projector catalog | Preserve all unit families, exact absence/null/present values, imported actor attribution and generation-before-equality checks. No missing-detail success. |
| Stable item and rollback identity | Core 01 REQ-01-054; Core 02 REQ-02-216 | Materializer identity constructors and retained entry-ref repository | Preserve selectors and logical coalescing across page boundaries, imports, deletion and deployment. Retain live lazy selector allocation unless separately corrected with proof. |
| Paging and authority | Core 01 §3.3.7 and §3.3.4.2; Core 04 authorization owner | HTTP currently validates cursor after full materialization | Preserve authentication/incident admission and authorized error precedence; bounded page query plus versioned opaque ordering cursor. Old cursors restart disposable reads through existing invalid-pagination handling. |
| Canonical source facts | Core 02 REQ-02-265 | Snapshot validation and source-owned projectors | No schema-less decoder, backfill, snapshot rewrite or automatic reset. Malformed required selected facts fail the page safely. |
| Read provenance and account/incident lifetimes | Core 03 REQ-03-299/100 | Retained-row hook and Timeline/Generic/Entity observations | Distinguish authority from presentation/window scope; reject stale cloned observations; preserve permitted same-account operation recovery. |
| Authoring, draft and receipt ownership | Core 03 §2.3A and feature operation owners | Draft store, captured History owner and specialized creation/recovery | Preserve raw values, clear intent, explicit submission, duplicate fencing and late receipt handling; no new persistence or retry engine. |
| Inspector shell and Details | Design §12.7; Core 01 REQ-01-615–617 | Shared shell/Details and admitted section sequence | Preserve reading-first appearance, geometry, focus and 17 configurations; close contradictory modes and remove no-op edit facades. |
| Public protocol and generated projections | Core 01/02; immutable OpenAPI baseline and pending 2.0.0 boundary | Authored Revisions OpenAPI and generated protocol/UI facades | No new route, mutation envelope or dual protocol. Review any actual public schema change through existing compatibility process; internal paging alone does not invent a semantic generation change. |
| WebSocket, query windows and grid integration | Existing Workbook/query/collaboration owners; grid-adapter boundary | Surface consumers and query hooks | Preserve event semantics, saved-view/schema identities and authorized off-window subjects. No direct grid-vendor imports. |
| Verification and readiness | Testing Harness owner and authored routing; Core 05 only for publication | Historical failures/reruns and current task guides | Fresh scoped evidence, deterministic failure cases and no claim-bearing performance result inferred from local query experiments. |

G15 is a proposed correction to make existing admission obligations explicit;
G16 deliberately retires disposable cursor compatibility. Both are planned here,
not implemented or newly adopted. Any required owner clarification enters its
actual owner in WS-11 before dependent production work.

### Previous-iteration section 4 record (historical)

This is a justified contract boundary, not a freeze of incidental UI behavior.
Navigation, layout, disclosure, action copy and hierarchy deliberately change.
History projection is corrected. Other public contracts stay stable because the
identified gaps do not justify changing them.

| Contract | Governing owner | Current evidence / tests | Required characterization | Risk / disposition |
| --- | --- | --- | --- | --- |
| Schema identity, fields, panels, feature routes and order | Core 01 §7.4, REQ-01-615–617; Core 03 §2.3A | View contracts, capability resolver and presentation tests | Every emitted schema renders each declared readable panel/action once; concealed and absent differ; saved-view changes cannot override schema config | Keep canonical identities/order. Do not add a public view schema version for visual layout alone. |
| Ordinary authoring and explicit submission | Core 03 §2.3A; design §§12.5,12.7 | Details/editor/draft components; inspector edit browser scenarios | Update and Ctrl/Cmd+Enter only; IME/native picker precedence; detach/resume/discard and captured revision | Preserve useful raw authoring and exact intent; replace misleading completion copy. |
| Mutation HTTP, idempotency and recovery | Core 01 §3.3.5; Core 03 §§3–4 | Feature operation owners and History recovery tests | Duplicate activation, identical uncertain request bytes, late receipt, accepted-write/failed-refresh | No new route or retry policy; UI disclosure never owns attempts. |
| History read and reversal | Core 01 §3.3.4.2; Core 02 §15; Core 03 §§10.1–10.4, REQ-03-139/140 | Materializer, OpenAPI source, presentation mapper, selector/equality code | Required semantic unit families, actor/time, retained entries, opaque selectors, ordering, paging, all legal reversal scopes | Correct missing structured detail; explicitly classify wire changes and cache migration in WS-06. |
| Evidence preview, attachment and access | Core 03 §8; Core 04 §2.0A; design §11 | Timeline Evidence panel, separate Evidence owner contributions | Lifecycle/availability/access combinations, file/drop/paste parity, pending linking, replay and read recovery | Keep independent rights; a visible count never authorizes metadata or bytes. |
| Authorization, stale reads and session boundaries | Core 03 REQ-03-299/100; Core 04 §§1–2; design §§12.7–12.8 | Explicit region models and current owner lifetimes | Immediate concealment of contents/counts/destinations, rejected late reads, incident loss versus account replacement | No new hidden-content cache or generic clear-all policy. |
| Storage, revisions, immutable audit and source projection | Core 02 §15 and source owners | Revisions provider catalog and retained snapshots | Derive detail from retained source facts; do not rewrite committed history, mutate projections to invent history, or expose raw snapshots | No storage migration assumed; prove coverage of historical snapshots before claiming none is needed. |
| WebSocket, saved views and preference persistence | Core collaboration/view owners | Existing inspector integration only | Refresh/retarget and selection continuity | No protocol, saved-view JSON or persistent inspector preference change. |
| Grid boundary, semantic selectors and generated contracts | Repo boundaries; design §§8,14–15 | Source/import rules and UI/view facades | Semantic focus/selector continuity; direct grid imports remain in adapter; authored generation | Change tests tied to retired copy/layout; retain identities with actual consumers. |
| Verification accounting | Adopted harness owner; machine routing under `tools/` | Current task guides and catalogs | Active rows run once, new rows are owned, visual captures reconcile | Catalog coverage is not proof of specification completeness. |

## 5. Gap-by-gap remediation decisions

Classification uses `must_fix`, `should_fix`, `defer`, or `intentional/no_action`.
Priority is delivery priority, not an incident severity. Each gap names its change
areas, rationale, long-term value, migration impact, unresolved risk and binary
validation. Proposed UX benefits remain hypotheses until observed in task review.

### G00 — Proposed observable behavior is not yet a complete owner contract

**Finding:** specification completeness gap in this proposal; `must_fix`, P1.
Current owners already specify field order, explicit editing and state delivery;
they do not adopt this proposed navigation/disclosure design. The audit's broad
slice descriptions are not sufficient defaults or acceptance contracts.

**Remediation / areas:** specification, typed projections where needed, tests and
documentation; WS-00. Update design §§7.3,8.4–8.5,12.4–12.7,14–15 and matching §18
criteria with the decision table in §7. Update Core 01's History transport detail
and Core 03's presentation references only where G10 needs owner clarification.
Use domain terminology without copying requirements into `domain.md`; change that
file only if a vocabulary/navigation reference actually changes. Leave the NLSpec
research artifact unchanged.

**Rationale / benefit:** one observable rule per owner prevents independent
implementations from inventing incompatible focus, disclosure or missing-data
behavior. Keep component factoring intentionally unspecified.

**Compatibility / migration:** specification adoption is a deliberate UX change;
reconcile cross-references and machine projections together. No competing
inspector specification or versioned UI mode. Existing obligations are retained
for their value, not copied wholesale into the new clauses.

**Risk if unresolved:** code/tests become accidental requirements, and later
surfaces repeat incompatible navigation and completion semantics.

**Validation:** each G01–G11 exit traces to an owner or explicit design hypothesis;
all defaults, state transitions, omission cases and error paths in §7 have one
owner; no contradictory duplicate text or prose-dependent executable check.

### G01 — Long inspectors have no persistent section navigation

**Finding:** structural weakness; `should_fix`, P1; usability benefit unmeasured.
The shell renders a continuous stack and the declared list has no shared navigator.

**Remediation / areas:** design specification, implementation, tests, documentation;
WS-02 after WS-00/01. Add persistent Sections disclosure and current-section text.
Use the same admitted semantic descriptors as the body, preserving declared order.
Target the section heading or existing entry control without forcing lazy reads.
Keep sections mounted and one internal scrollport.

**Rationale / benefit:** navigation cost no longer grows with preceding content;
one identity/admission decision serves every schema and prevents concealed
sections from leaking through a separate navigation list.

**Compatibility / migration:** new local navigation replaces scroll-only access;
existing contextual openers and focus registry converge on it. No URL, saved-view,
persistence or grid-vendor change. Remove redundant navigation adapters, not
owner-required shortcuts.

**Risk if unresolved:** growing content becomes harder to find; independent
feature navigation increases coupling. Excess task time is a hypothesis.

**Validation:** every readable declared destination is reachable with at most two
control activations (open, choose; keyboard traversal is not counted as activation).
Passive scrolling never moves focus or requests data. Subject/access changes
remove stale destinations immediately; menu Escape precedes editor/inspector
Escape; supported widths retain visible Close, focus and status strip.

### G02 — Saved Details values have equal visual weight and long travel

**Finding:** structural weakness; `should_fix`, P1. Details stacks every declared
field; Timeline declares 24 fields.

**Remediation / areas:** design, implementation, tests; WS-03. Use aligned
label/value/action rows for short values, full-width narrative rows, responsive
stacking and explicit full-value disclosure. Classify through existing semantic
field metadata or a minimal authored presentation mapping, never display labels.
Give collection summaries a semantic route to existing management regions.

**Rationale / benefit:** field meaning and reading density become consistent
across surfaces without creating a second field registry or turning collections
into scalar mutations. New fields use deterministic defaults.

**Compatibility / migration:** retain declared fields/order, read-only content,
safe rendering and zero/false/empty/null/unloaded distinctions because they carry
meaning. Retire per-surface duplicate field-row layouts. Update layout assertions
and goldens; no field or stored-data migration.

**Risk if unresolved:** adding fields increases scanning effort and repeated
per-surface layout work; aggressive ad hoc hiding could remove required facts.

**Validation:** all readable declared fields occur once and in order; long labels,
URLs and prose remain fully accessible; unrecognized field presentation defaults
to a readable stacked row. Collections navigate only to admitted destinations;
read-only, zero, false, empty and missing states have separate assertions.

### G03 — Workflow and History actions compete with reading

**Finding:** structural weakness; `should_fix`, P1. Workflow wraps equal-weight
buttons; History repeats reversal buttons and leads with row deletion.

**Remediation / areas:** design, implementation, tests, documentation; WS-07/08.
Use an ordered Workflow action list with authored labels and explicit outcomes
(open authoring, review, or navigate). Put event reversal choices inside the
relevant event disclosure; place row delete/restore in a labelled Record actions
group after the event-reading region. Keep one primary affirmative action per
active authoring/review region, with destructive language on destructive actions.

**Rationale / benefit:** reading and action selection have a stable hierarchy;
new declared actions reuse presentation while capability owners remain explicit.

**Compatibility / migration:** migrate current action bindings and tests together;
preserve declared order, identity, legal disabled discovery and each typed shared
reason. No arbitrary More menu, duplicate action registry or global Save control.

**Risk if unresolved:** reviewers must scan repetitive controls and can mistake
row-level and event-level scope; wrong-action frequency is a hypothesis to test.

**Validation:** every action dispatches its canonical route exactly once, only
when currently admitted. Each reversal shows its real scope, including changes
to other records. Destructive confirmation starts at a safe control and invalidates
on owner-defined changes. Disabled reasons are deduplicated by identity/parameters.

### G04 — History reads like raw implementation output

**Finding:** confirmed rendering limitation plus structural weakness; `must_fix`
for required data display, `should_fix` for visual hierarchy, P1. The renderer uses
full ISO text; the mapper forwards summary/operation and no structured unit detail.
Its caller supplies no readable actor label.

**Remediation / areas:** design, implementation, tests; WS-07, dependent on G10.
Render concise authorized descriptions from typed semantic facts, absolute time
with explicit timezone, and visible attribution. Use an authorized actor label
when supplied; otherwise show the attributed actor ID as such, not an invented
person name. Disclosure exposes complete required semantic units, exact original
timestamp and technical identifiers. Never parse prose to recover semantics or
infer action eligibility.

**Rationale / benefit:** reviewers can tell what happened without decoding storage
names; new unit families gain an explicit mapping instead of string heuristics.

**Compatibility / migration:** change display copy/time formatting and associated
snapshots. Keep exact timestamp and opaque identities available. An unavailable
friendly label does not remove mandatory actor attribution. Historical missing
semantics remain a G10 completion failure, not a permanent cosmetic fallback.

**Risk if unresolved:** incomplete audit understanding and brittle text-derived
behavior as new source families arrive.

**Validation:** field, link, mention, tag, Evidence association and capture-state
changes render truthfully; absent optional names/details remain explicit. Timezone,
invalid-input handling and exact-time access are deterministic. Event expansion
preserves paging, order, opaque targets, requested-change highlighting and focus.

### G05 — Editing presents ambiguous completion choices

**Finding:** structural weakness; `should_fix`, P1; actual user confusion is a
hypothesis. Update and Done editing are produced in different contributors.

**Remediation / areas:** design, implementation, tests, documentation; WS-03.
Compose one Unsaved change region with accepted value still visible. Group Update
and Close editor; explain that closing retains unfinished work. Keep Clear at the
value control and Discard draft distinct. Expose retained-work status and eligible
Resume/Discard on the original field even when no editor is attached. Derive cues
from the canonical draft store, not another retention cache.

**Rationale / benefit:** command placement makes saving versus detachment explicit
while preserving source-specific validation and exact authoring identity.

**Compatibility / migration:** replace misleading Done editing copy and migrate
its tests without a label alias. Do not change explicit submission keys, null/clear
intent, account/incident scope or grid editing. No durable persistence is added.

**Risk if unresolved:** analysts may think detached work was saved or abandon
retained work; duplicate UI draft state would create correctness bugs.

**Validation:** Update and Ctrl/Cmd+Enter submit only the captured revision;
Close, Escape, blur, Tab and section navigation never submit/discard. Late success
clears only captured work; newer drafts survive. Retarget, stale field dependency,
read-only and reauthorization cases expose correct Resume/review/Discard behavior.
Global Saved cannot imply an unfinished local draft was submitted.

### G06 — Inspector controls do not consistently consume design roles

**Finding:** confirmed token-use defects; `must_fix`, P1. Title/section styles omit
full typography roles; buttons inherit font and use general spacing/radius;
Timeline/Indicator inputs have competing padding definitions.

**Remediation / areas:** implementation and tests; specification/projections only
for missing variants that prove necessary; documentation; WS-01. Apply complete title,
section, body, metadata, button and technical roles and existing component states.
Use shared field/group/action/feedback presentation. Adopt a compact variant only
if existing tokens cannot satisfy an observable requirement after layout changes.

**Rationale / benefit:** one styling decision keeps density, accessible focus and
future token changes coherent. Remove accidental native fieldset outlines while
retaining legends and group semantics.

**Compatibility / migration:** visual change across actual consumers; inventory
imports before editing shared styles. Move inspector callers first and delete
unused variants in the same slice. Keep grid-specific geometry separate. No new
theme, font, icon dependency or package is justified.

**Risk if unresolved:** token updates stop propagating reliably, and each future
workflow introduces another styling dialect or inaccessible compact control.

**Validation:** computed roles and component states match authored projections;
all supported densities, disabled/pending/destructive states, long labels, zoom
and focus remain usable. Migrated callers have no obsolete style path. Generated
facades are rebuilt from authored inputs if those inputs change.

### G07 — Selected relationships consume most of the viewport

**Finding:** structural weakness; `should_fix`, P2. One mention expands status,
target, resolution, filter, picker and operation controls into a tall form.

**Remediation / areas:** design, implementation, tests; WS-04. Use compact item
summaries with raw mention, resolution state and authorized target. Expand selected
correction controls inside the original source group; keep filter/read status,
retry and paging within the target chooser presentation. Preserve unfinished or
uncertain-operation cues when its controls are collapsed.

**Rationale / benefit:** source context remains legible with less repeated form
chrome; the existing mention owner remains the single source of operation truth.

**Compatibility / migration:** presentation-only cutover. Keep loaded-target scope,
selected identities across pages, raw versus canonical links, and the exact
Dismissed in this session qualification. No automatic enumeration or entity deletion.

**Risk if unresolved:** larger relationship collections become difficult to inspect;
local UI shortcuts can accidentally imply complete search or erase pending work.

**Validation:** selected controls remain beside the correct field/item after paging
and retargeting; failed/partial reads do not imply no targets or unresolve a known
mention. Resolve, correct, dismiss, restore and create-then-link preserve their
existing permissions, reviews, exact replay, receipts and local outcomes.

### G08 — Specialized authoring lacks a consistent visual and mode hierarchy

**Finding:** structural weakness; `should_fix`, P2. Assessment always supplies a
creation title even with a saved subject; specialized forms use different group
and control treatments.

**Remediation / areas:** design, implementation, tests, documentation; WS-09.
Use the saved subject for the record header and put the specialized operation title
in its attached region. No-subject creation gets an explicit creation header and
admitted context. Reuse field/group/action/feedback presentation while leaving
submission, review and lifecycle with each feature.

**Rationale / benefit:** users can distinguish record identity, operation and saved
outcome. New forms reuse a consistent presentation without inheriting another
feature's validation, authoring lifetime or submission semantics.

**Compatibility / migration:** migrate Assessment, Indicator, Note, coordination,
Task/Decision and related-Evidence callers by family. Keep append-only Assessment
and Indicator lifecycle rules because their domain meaning remains valuable.
Retire duplicate spacing, fieldset and action-row treatments as callers migrate;
no universal form engine or panel-global Save is introduced.

**Risk if unresolved:** creation can be mistaken for editing, and every new workflow
adds another visual and interaction dialect that must be maintained separately.

**Validation:** saved-record and no-subject headers differ correctly. Every declared
form retains explicit source identity, required minima, review and recovery.
Source replacement/clearing, duplicate submit, detached late outcomes and
permission loss are tested for each feature family, not inferred from shared CSS.

### G09 — Evidence presentation emphasizes mechanics over accepted information

**Finding:** structural weakness and bounded design opportunity; `should_fix`, P2.
The Timeline contribution presents an attachment count and file chooser, with
recovery above them; other Evidence-owned contributions supply access behavior.

**Remediation / areas:** design, implementation, tests, documentation; WS-05.
Separate authorized accepted information, independently admitted preview/download
controls, and attachment progress/recovery. Use a compact Attach file entry with a
named drop/paste region and keyboard alternative. Render item rows only from
already-authorized owner data; otherwise describe the observed count/scope honestly.

**Rationale / benefit:** inspecting evidence and attaching a file become distinct
activities. Independent access and upload/link outcomes stay visible, so the
presentation can accommodate future owner-provided metadata without owning reads
or inferring capabilities from counts.

**Compatibility / migration:** chooser, drop and paste retain current file-type,
size, lifecycle, authorization and upload contracts. No richer enumeration API or
storage migration is required. New Evidence item browsing is DEFERRED outside this
plan unless existing authorized data already supports its presentation.

**Risk if unresolved:** upload, Evidence creation and source linking can be mistaken
for one completed transaction; a richer-looking UI could falsely imply available
metadata, complete totals or permission to preview/download.

**Validation:** counts, preview/download rights, unavailable/quarantined/deleted
states, uncertain upload, saved-but-unlinked and saved-but-unrefreshed outcomes are
distinct. Chooser/drop/paste use the same admitted owner command. Recovery remains
source-local, preserves exact uncertain attempts, and reads after acknowledged
writes. Preview stays in the workbook and no protected metadata or bytes leak.

### G10 — History's transport and source projection cannot express required detail

**Finding:** confirmed end-to-end gap; `must_fix`, P1. Core 03 REQ-03-139 requires
unit-level change detail. The authored OpenAPI allows arbitrary unit objects;
`history_materializer.go` emits target metadata and before/after-presence flags,
not field or relationship deltas. The frontend then discards units entirely.

**Remediation / areas:** specification clarification, authored protocol projections,
backend and frontend adapters, source-owner contributions, tests and documentation;
WS-00/06/07. Define a closed, versioned semantic unit contract with explicit variants
for the required families and safe value/availability rules. Source owners derive
units from retained authoritative before/after facts through pure contributions;
Revisions validates/catalogs/composes them. No central source-table switch, browser
snapshot decoder, untyped map spreading, or projection-derived history fallback.

Define stable unit identity/order, public field or relation identities, operation,
value omission/null/redaction/unavailable meanings, actor provenance and unknown
version handling in the owner. Keep presentation text distinct from selector
identity. Include imported attribution in the producer/schema/consumer parity
review: `history_model.go` can emit `source_actor_id`, while the inspected closed
OpenAPI item does not declare it. Resolve against its adopted import owner rather
than silently accepting arbitrary additional properties.

**Rationale / benefit:** semantic History becomes reconstructable, testable and
extensible without coupling source storage formats to every client. Narrow typed
facts reduce accidental disclosure compared with exposing whole snapshots.

**Compatibility / migration:** use the approved coordinated server/browser
cutover under the pending OpenAPI 2.0.0 boundary, keeping `/api/v1` and the
immutable released baseline. No legacy response adapter is supported. Preserve
logical item/entry identities and canonical retained audit data. Core 02
REQ-02-265 prohibits schema-less readers, member-presence inference and backfill;
unsupported old databases remain subject to its operator-controlled reset policy.
This implementation does not reset a database. Missing detail in supported
canonical history blocks closure; never invent values or substitute current rows.

History content equality currently treats `diff_summary` as immutable. Changing
its representation during a retained browsing session needs a read-cache generation
cutover/reload policy that leaves captured mutation attempts and receipts intact.
Test upgrade and rollback; do not use a cache reset to bypass receipt recovery.
No storage migration is planned until retained-data inventory proves one necessary.

**Risk if unresolved:** cosmetic History can still violate required review behavior;
new sources multiply ambiguous payloads and unsafe ad hoc decoding.

**Validation:** required unit families are produced from representative current and
retained snapshots; source catalogs reject missing/duplicate/invalid contributions;
public schema, generated types and renderer agree; hidden data and raw storage
identifiers/bytes are absent. Paging/order/selectors and rollback semantics remain
stable. Compatibility report and retained-session upgrade/rollback evidence are
recorded, with every supported consumer accounted for.

The WS-00 contract review must close at least this semantic mapping; wire member
names and source-private decoding stay with their owners:

| Change family | Minimum truthful public detail | Semantic owner |
| --- | --- | --- |
| Field change | Public field identity, operation, authorized before/after value states; absence distinct from null, false, zero and empty text | Source record owner using retained snapshots |
| Link change | Public relationship identity/type, add/remove/replace and authorized endpoint references | Links or the source owner of the declared relationship |
| Mention change | Stable public mention reference, raw mention when visible, prior/result resolution state and authorized target facts | Entities mention owner |
| Tag change | Add/remove and authorized tag identity/value | Existing tag/source owner |
| Evidence association | Attach/detach and authorized Evidence/source references; no object bytes, upload tokens or preview handles | Evidence association owner |
| Capture-state transition | Prior/result capture state and admitted operation | Timeline owner |
| Record/merge and other source-owned events | Public record operation, merge participants when authorized, and source-defined semantic deltas for Assessment/Indicator and other supported histories | Respective source owner; no reinterpretation as arbitrary scalar patches |

Each unit needs deterministic identity/order within the stable logical item and
an explicit representation version. An omitted value is not automatically null;
unavailable or withheld detail uses a typed state only when the owner permits
disclosing its existence. Whole-item concealment exposes no such hint. Keep
presentation detail separate from the opaque rollback target already supplied by
History. This is a display projection of immutable source facts, not a second
event store or event-sourcing redesign.

### G11 — Validation evidence does not yet prove redesign completeness

**Finding:** confirmed coverage gap and untested usability hypotheses; `must_fix`,
P1 for delivery completion. The base visual scenario asserts Details content but
scrolls to Relationships before its first inspector screenshot. The original audit
sampled specialized workflows, not all seventeen configurations or accessibility states.

**Remediation / areas:** tests, fixture/catalog projections, documentation and
manual task review; WS-10 plus relevant tests in each earlier workstream. Add
first-viewport saved Details, attached edit and retained-draft cases; exercise all
schema configurations and the state/geometry matrix in §8. Compare representative
reading, editing, relationship, History and creation tasks.

**Rationale / benefit:** evidence tests the intended user outcomes and dangerous
transitions rather than preserving implementation-shaped snapshots. Future sources
inherit a finite coverage contract.

**Compatibility / migration:** update authored fixture declarations/catalog rows
and generated evidence only through public targets. Retire superseded captures only
after reconciliation proves zero consumers; keep useful existing scenarios.

**Risk if unresolved:** a visually improved happy path can ship with hidden focus,
authorization, disclosure or retained-work regressions; sampled screenshots may be
misreported as complete acceptance.

**Validation:** every applicable acceptance row has fresh evidence, every exclusion
has a scope/owner rationale, and no required row is blocked. No golden update hides
a functional failure. Task review records outcomes and remaining hypotheses without
claiming measured improvement that was not observed.

### G12 — History has competing state representations

**Classification / owners:** must_fix; History browsing and operation owners,
inspector binding; Core 01 REQ-01-052A and Core 03 recovery lifetimes; WS-12.

**Evidence:** `workbookRecordHistoryModel.ts` intersects optional browsing with
legacy phases containing their own accepted or retained data. Reconciliation in
`useWorkbookRecordHistoryController.ts` dispatches legacy load events and then
conditionally casts page data into browsing state. Timeline injects an external
reducer while the shared hook creates a local one. `HistoryBinding.reconcile`
and operation `currentHistory` erase paging through `RecordHistoryData`.

**Remediation / areas:** Implementation, types, tests and source documentation.
Make the browsing model the sole authority for accepted pages, provenance,
continuation and read failures. Require `HistoryPage` through reconciliation.
Keep captured attempts and receipts in their existing owner; inspector state
contains only attachment, presentation and unsubmitted review. Migrate Timeline
and Recovery/merge consumers before removing `legacyHistoryReducer`, duplicate
load events, page-less acceptance, unused initialization inputs and indirect
History type re-exports through inspector modules. Required production ports
must not silently fall back to an uncoordinated operation; preserve explicit
bounded test injection without creating a second production path.

**Rationale / long-term benefit:** One authoritative representation per lifetime
removes synchronization rules and lets new readers reuse History independently
of inspector presentation. It reduces coupling without a generic state engine.

**Compatibility / retirement:** Internal consumer migration; preserve explicit
reads, event identity, generation checks, Timeline FIFO coordination, captured
request bytes, acknowledgement and source refresh. Remove migrated compatibility
paths in the same slice; no forwarding reducer or page-less adapter remains.

**Risk if unresolved:** Divergent accepted content, stale review eligibility,
incompatible generations and receipt loss during attachment changes.

**Validation / exit:** All production consumers use the same browsing contract;
no duplicate accepted-page storage or legacy load-event bridge remains. Paging,
refresh failures, retargeting, generation restart, late outcomes, exact replay
and accepted-write read recovery pass without retiring unrelated local work.

### G13 — Public surfaces exceed demonstrated consumer needs

**Classification / owners:** should_fix, mandatory for this cleanup; source
owners and existing test support; WS-12/13 retire touched paths and WS-17 closes
the complete ledger.

**Evidence:** The current symbol scan found helpers/types exported without an
external production consumer. `historyDiffFixture` has thirteen test-file
consumers and no production consumer. `WorkbookInspectorPanelSection` still
renders inside Shell; its outside uses are tests/policy evidence. The controller
accepts unused `initialHistory`, `beginMutation` and `commands` inputs; verify
callers as well as reads before removing each input.

**Remediation / areas:** Implementation, tests, ownership projections and docs.
Maintain the symbol disposition ledger in §7. Remove unused inputs and obsolete
branches after migrating callers; privatize internal helpers/types. Move the
app-specific History fixture into existing `apps/web/src/testing/` support,
updating its thirteen test imports. Retain live components with named consumers.
Reconcile barrels, aliases, dynamic registration and policy tests before removal.

**Rationale / long-term benefit:** A small intentional interface prevents new
features and tests from depending on incidental internals. Future additions
extend meaningful owners instead of maintaining deprecated aliases.

**Compatibility / retirement:** Migrate repository callers atomically. No
forwarding exports, dead-code suppression or compatibility facade is introduced.
Do not delete a component merely because its name says legacy or its file is short.

**Risk if unresolved:** Accidental public APIs and test-only dependencies increase
coupling, maintenance cost and future migration obligations.

**Validation / exit:** The complete file/symbol ledger has named retained
consumers or removal evidence. Type/import-boundary and affected behavior tests
pass; test-only fixture imports cannot enter production. No unexplained shared
export or superseded implementation remains.

### G14 — Presentation interfaces permit contradictory configurations

**Classification / owners:** must_fix; shared inspector presentation and existing
source draft owners; design §12.7 and Core 01 REQ-01-615–617; WS-13.

**Evidence:** Read-only Details supplies empty editor slots and no-op callbacks
to the editable renderer. Shell independently permits a nullable subject,
creation mode, arbitrary children and optional sections. Details takes types
from the draft hook's return value rather than a pure presentation interface.

**Remediation / areas:** Implementation, typed presentation interfaces, tests and
docs. Separate saved-value reading from edit attachment. Use a pure editing
presentation contract supplied by its operation owner. Introduce closed saved
record, empty selection and creation variants. Saved records require admitted
sections; creation explicitly supplies source context and local content.

**Rationale / long-term benefit:** Interfaces represent valid combinations and
make read-only/append-only consumers simple. A new source supplies its own
commands and lifetime through existing slots, without a universal form engine.

**Compatibility / retirement:** Preserve adopted visuals, navigation, field
order, explicit commands and source-owned authoring. Migrate all consumers,
then remove no-op editing and conflicting shell props in the same slice.

**Risk if unresolved:** Future consumers bypass section admission or silently
receive ineffective controls; source operation state leaks into presentation.

**Validation / exit:** All seventeen schema configurations are accounted for;
read-only Details needs no editing callbacks. Saved/empty/creation modes retain
correct headers, mounted sections, focus destinations, disclosure and drafts.
False, zero, empty text, null and unloaded values remain distinct.

### G15 — Retained-row admission relies on object identity

**Classification / owners:** must_fix structural weakness; existing query and
mutation observation owners, consumed by inspector; Core 03 REQ-03-299/100;
WS-14. No actual unauthorized disclosure has been established by this inspection.

**Evidence:** `useRetainedInspectorRow` rejects a retired object reference but
does not receive evidence of the authority under which a row was accepted.
Timeline, Generic and Entity compose query and committed operation observations;
Entity also converts accepted API rows into presentation objects.

**Remediation / areas:** Implementation, source interfaces, security/lifecycle
tests and docs. Carry explicit accepted-read provenance from existing query and
mutation owners. Match both record and current authority scope before retaining
an observation. Keep query-window and presentation resets separate from that
scope. Remove object-reference retirement; do not mint fresh provenance merely
because a component cloned or rendered a row.

**Rationale / long-term benefit:** Admission remains understandable and correct
when data is converted, cloned, cached or delivered late. Future sources reuse
a stable observation contract without a second authorization registry.

**Compatibility / retirement:** Preserve authorized off-window reading and
source-owned recovery. No new permission policy, durable cache or draft lifetime.
Fresh evidence may arrive in the same render as a scope change; stale evidence
may not acquire the new scope by relabeling.

**Risk if unresolved:** Reconstructed stale observations can defeat an identity
check, and presentation behavior depends on incidental allocation choices.

**Validation / exit:** Reject cloned stale rows, wrong records, lower versions
and late old-scope observations. Cover account replacement, incident-access loss,
same-account recovery, fresh evidence concurrent with scope change and ordinary
window eviction. Retain drafts/receipts only as their existing owners permit.

### G16 — History pagination materializes the entire history

**Classification / owners:** must_fix for selected growth/readiness scope;
Revisions application/repository, HTTP cursor adapter and source projectors;
Core 01 §3.3.4.2/§3.3.7 and Core 02 REQ-02-216/218/265; WS-15.

**Evidence:** `GetHistory` calls `ListRecordHistory`, which loads mutation and
revision rows without a page bound, projects every item, resolves attribution
and rollback actions, then sorts serialized maps. HTTP `pageRecordHistory`
scans that full list for the cursor and slices it. Cursor validation follows
the expensive read. Existing GIN history-association and ordering indexes are
present; this finding does not assume a new index is automatically necessary.

**Remediation / areas:** Revisions implementation, transport adaptation, internal
typed contracts, integration tests and compatibility/operating documentation.
Validate pagination after authentication and incident admission but before
History materialization. Introduce typed application page requests/results.
Select at most the effective limit plus one logical event descriptors using
canonical ordering. Resolve mutation/revision coalescing before applying the
page boundary. Project complete facts, attribution, selectors and legal action
metadata only for selected events and necessary dependencies. A whole-change-set
rollback review may legitimately require that selected change set's full facts;
report that dependency separately from unrelated retained history.

Use a versioned opaque ordering cursor and retire the full-list anchor scan.
Keep cursor serialization in HTTP/platform pagination, not the Revisions domain
port. Use current indexes first; justify any additive index with query plans.
Do not add a second stored History representation, central source switch or
projection-derived fallback.

**Rationale / long-term benefit:** Application memory and projection work follow
the selected page and its required dependencies, rather than every retained
snapshot. New source families continue contributing pure projectors.

**Compatibility / retirement:** Preserve semantic payloads, canonical ordering,
coalesced item identities, rollback selectors, imports and mutation envelopes.
Old disposable cursors use the existing invalid-pagination restart path; no
legacy cursor adapter, forced reload, snapshot backfill, reset or history rewrite.
Cursor encoding changes alone do not change semantic representation generation.
Any new index is additive and may remain installed during application rollback.

**Risk if unresolved:** Growing histories cause excessive memory use, repeated
projection/eligibility work and long transactions even for small pages.

**Validation / exit:** Compare concatenated pages against canonical fixture
expectations across equal timestamps, mixed mutation/revision events, page-boundary
coalescing, concurrent additions, imports, deletion and all source families.
Malformed required selected facts fail safely; item selectors and legal reversal
scopes remain identical. Record query plans, loaded/projected-item counts and
memory observations at increasing history sizes. Ordinary page reads must not
load all retained snapshots. Lookahead establishes continuation without projecting
an extra event's public detail. Verify invalid cursors do not invoke materialization.

### G17 — Intermittent failures lack durable explanations

**Classification / owners:** must_fix verification/reliability gap; affected
Workbook source owners and browser harness owners; WS-16. Historical failures
are evidence of unresolved explanation, not proof of a current product defect.

**Evidence:** §10 records stateful run `20260921T073129Z-p96952` at 39/42 and
accessibility run `20260921T073453Z-p32834` at 18/20 before unchanged focused
reruns. Timeline recovery and Coordination focus require causal review.
The historical full and focused run directories remain present at planning time;
they have not been rerun or relabeled by this document update.

**Remediation / areas:** Tests, fixtures, affected implementation and verification
documentation. Reproduce relevant failures with controlled response ordering and
explicit state transitions. Identify source versus fixture/harness responsibility,
then fix that owner. Replace timing-dependent expectations with observable
completion boundaries; preserve the intended focus and recovery outcomes.
Record unrelated Network Flow evidence separately unless a shared harness defect
connects it to this scope.

**Rationale / long-term benefit:** Failures become reproducible, actionable
regressions rather than a reason to rerun broad suites until green.

**Compatibility / retirement:** Preserve intended behavior and replace unstable
choreography. No weakened assertions, arbitrary delays or increased retries to
obtain a pass; unrelated product refactors remain excluded.

**Risk if unresolved:** Timing failures may hide lost focus, blocked editing,
duplicate submission or incomplete accepted-write recovery.

**Validation / exit:** Each relevant failure has a documented cause, a deterministic
regression and a successful ordinary affected suite. An unchanged isolated pass
alone cannot close G17. If original artifacts are insufficient, collect fresh
reproduction evidence rather than inventing a historical explanation.

### G18 — Readiness needs a fresh, scoped handoff

**Classification / owners:** must_fix evidence boundary; controlling tracker,
authored acceptance/routing and affected source owners; WS-11 and mandatory WS-18.

**Evidence:** Existing DONE/PASS records describe WS-00–WS-10. This iteration has
new structural, provenance, scaling and reliability exits with no execution
results. The archived validation handoff is not required by this task.

**Remediation / areas:** Documentation, verification evidence and acceptance
mappings. Maintain the new acceptance, retirement and rollout ledgers independently
of historical records. Use fresh evidence for changed behavior and clearly
identify any qualifying retained evidence and its applicability.

**Rationale / long-term benefit:** Maintainers can distinguish planned cleanup,
completed changes, compatibility obligations and remaining hypotheses without
rediscovering the subsystem or treating planning prose as product authority.

**Compatibility / retirement:** Preserve canonical audit data and operation
recovery through coordinated server/browser rollout and rollback. Retain useful
historical records; do not rewrite previous failures as successful full runs.

**Risk if unresolved:** Historical passes may be mistaken for current production
readiness or unfinished work may be hidden behind the completed old tracker.

**Validation / exit:** WS-11–WS-18 are DONE with serial checkpoints, every new
mandatory acceptance has evidence, superseded code is retired and rollout/rollback
is verified. Fresh digest acceptance is recorded separately from historical
A001–A027. Participant usability measurements remain unavailable unless collected;
no timed completion, error-rate or scroll-distance benefit is invented.

## 6. Workstreams, dependencies and phase gates

### Next-iteration serial phases

Execute WS-11 through WS-18 serially. The old dependency table below applies only
to WS-00–WS-10. At each new entry save `IN_PROGRESS`; after its checks pass save
changed paths, decisions, commands/results/artifacts, compatibility, retirement,
resolved failures and exit evidence, then mark `DONE` before starting the next.
A failed exit remains `IN_PROGRESS` or `BLOCKED`. WS-18 is mandatory and final.

| Phase / workstream | Class / predecessor | Delivery and gap | Successor | Exit and phase risk |
| --- | --- | --- | --- | --- |
| A / WS-11 | root / none | Owner mappings, refreshed file/symbol inventory, compatibility and acceptance; G18 | WS-12 | Every finding has an owner and binary exit; clarify real owner conflicts before dependent changes. |
| B / WS-12 | chain / WS-11 | One History read authority and migrated bindings; G12 | WS-13 | All consumers migrated, read/recovery checks pass, old bridge removed; avoid losing operation ownership. |
| B / WS-13 | chain / WS-12 | Closed shell modes and separate saved reading/edit attachment; G14 | WS-14 | All configurations and consumers pass without contradictory props; preserve admission and focus. |
| B / WS-14 | chain / WS-13 | Explicit accepted-observation provenance; G15 | WS-15 | Authority/late-observation matrix passes; no stale row acquires current authority through cloning. |
| C / WS-15 | chain / WS-14 | Bounded typed History page queries; G16 | WS-16 | Semantic, selector, scaling, cursor restart and rollback evidence passes; incomplete logical events are blockers. |
| D / WS-16 | chain / WS-15 | Causal fixes for intermittent validation failures; G17 | WS-17 | Deterministic regressions and ordinary affected suites pass; assertion weakening is prohibited. |
| D / WS-17 | chain / WS-16 | Final symbol/consumer and retirement reconciliation; G13 | WS-18 | No superseded adapters, test-only production imports or unexplained shared exports. |
| E / WS-18 | chain / WS-11–17 | Fresh validation, retirement, rollout/rollback and handoff; G18 | none | Every mandatory new acceptance and final exit passes; historical evidence cannot substitute. |

Rollback structural changes as coherent caller/owner slices. Do not leave a
second reducer, shell or page protocol as a temporary release architecture.
Additive indexes may remain during application rollback; audit data is not reset.

### Previous-iteration section 6 record (historical)

Each WS is one separately reviewable implementation slice. Workstreams are
serial in the default delivery order below, even where minimum technical
dependencies would allow overlap. Before starting the next workstream, update §9
with changed paths, adopted decisions, commands/run roots, risks, retirement,
compatibility and its observed exit. `DONE` means the slice was executed and its
exit passed; writing its plan never qualifies. Allowed work statuses are `TODO`,
`IN_PROGRESS`, `BLOCKED`, `DONE`, `DEFERRED`, `DROPPED`.

| WS / phase | Class / technical prerequisites | Goal / gaps | Next workstream | Exit and checkpoint |
| --- | --- | --- | --- | --- |
| WS-00 / phase A | root / none | Owner cleanup and complete decisions; G00, G10 | WS-01 | Adopted text resolves every §7 decision, owner references and acceptance mapping; record any blocked owner issue before dependent work. |
| WS-01 / phase B | chain / WS-00 | Typed presentation projections and shared visual foundation; G06 | WS-02 | Authored roles/projections and migrated common controls agree; focused tests and token review pass; obsolete variants removed. |
| WS-02 / phase B | chain / WS-01 | Shared frame and section navigation; G01 | WS-03 | One admitted descriptor sequence drives navigation/body, focus and concealment checks pass in adjacent/overlay layouts. |
| WS-03 / phase C | chain / WS-02 | Saved Details and ordinary edit hierarchy; G02, G05 | WS-04 | Read/edit/close/resume/discard and revision-bound outcomes work across Timeline, Generic and Entity consumers. |
| WS-04 / phase C | chain / WS-02 | Relationship summary/correction disclosure; G07 | WS-05 | All mention transitions and bounded target recovery remain correct; hidden controls leave actionable unresolved-status cues. |
| WS-05 / phase C | chain / WS-02 | Evidence information and attachment hierarchy; G09 | WS-06 | Accepted data, independently authorized access and attachment/link/recovery states remain distinguishable. |
| WS-06 / phase D | chain / WS-00 | Semantic History contract, source contributions and compatibility cutover; G10 | WS-07 | Required units complete across source families, retained data and schema/generation/security/compatibility checks; no unresolved migration. |
| WS-07 / phase D | chain / WS-02, WS-06 | Readable History and deliberate corrective actions; G03, G04 | WS-08 | Mandatory attribution/time/units displayed, exact scopes and opaque targets preserved, History recovery and focus pass. |
| WS-08 / phase D | chain / WS-02 | Workflow action catalog presentation; G03 | WS-09 | Every declared action is discoverable once, ordered and correctly described/admitted; no hidden compatibility alias. |
| WS-09 / phase D | chain / WS-03, WS-05, WS-07, WS-08 | Specialized form and mode consistency; G08 | WS-10 | All feature families migrated to shared visuals with source identity, independent submission and recovery intact. |
| WS-10 / phase E | chain / WS-00 through WS-09 | Validation and handoff completion; G11 and all gaps | none | §12 product criteria met, acceptance/evidence recorded, retirement complete, rollout/rollback and next maintainer documented. |

| Phase | Main risk | Exit criteria | Stop/rollback rule |
| --- | --- | --- | --- |
| A — Specification cleanup | Design proposal becomes accidental authority or weakens required history | Complete owner decision tables and traceable binary criteria; no owner contradiction in dependent scope | Keep proposals in tracker until adopted; repair contradiction before dependent implementation. |
| B — Shared foundation | New navigation duplicates admission or changes mounts/focus | WS-01/02 exits and shared-consumer coverage pass | Revert coherent presentation/projection slice; do not keep two shell implementations behind an indefinite flag. |
| C — Reading and local authoring | Disclosure loses identity, masks draft or conflates Evidence rights | WS-03/04/05 exits and local failure/permission/recovery matrix pass | Revert affected visual composition; retained operation owners and receipts remain authoritative. |
| D — History and workflow | Payload disclosure, incompatible clients, incorrect rollback scope or form mode | WS-06–09 exits, explicit compatibility classification and source-family coverage pass | Roll back frontend/server/projection cutover coherently; preserve immutable audit and in-flight recovery. |
| E — Validation and handoff | Baseline failure or golden update is mistaken for completion | Every applicable acceptance row is evidenced; all mandatory WS are DONE | Do not mark effort complete on partial runs; record failing target/run root, cause and remaining work. |

## 7. Proposed slice implementation contracts

### Next-iteration slice implementation contracts

The following defines later production work; no production slice is executed
by this document update. WS-11 adopts any needed observable clarifications in
actual owner sections. It does not copy implementation decomposition into specs.

| Workstream | Intended change and affected seam | Characterization and completion | Validation / rollback |
| --- | --- | --- | --- |
| WS-11 | Recheck §2/§4, freeze intended semantic ordering/coalescing, authority provenance and cursor restart; owner amendments only where needed | All G12–G18 and I3 acceptance rows mapped; all public consumers/registrations accounted for | Refresh three task guides; owner/doc review and Markdown; no data migration |
| WS-12 | HistoryPage throughout binding/reconciliation; one browsing authority; Timeline retains coordination but drops external duplicate read reducer | Single accepted-page path; preserved explicit read, requested-change lookup, review and captured operation semantics | §8 History unit/service slices and frontend checks; revert coherent owner plus caller migration |
| WS-13 | Saved record/empty selection/creation shell union; saved-value renderer with an explicit edit attachment contract | No no-op edit facade or subject/mode conflict; four surface families and specialized creations retain behavior | §8 presentation/subject/edit tests and accessibility; revert shared component and consumers together |
| WS-14 | Propagate accepted observation scope from query/operation sources through conversion to the retention hook | No reference blacklist; current-record/current-authority checks and valid version precedence; no change to source recovery lifetimes | §8 provenance, permission, subject and late-response cases; revert source port and consumers together |
| WS-15 | Typed HistoryQuery page input and HistoryResult page output; page descriptor selection precedes fact projection and HTTP serialization | Canonical order: committed time descending, change-set identity descending, mutation before revision-only, mutation sequence ascending. Preserve first-mutation revision coalescing using complete dependency facts before page slicing. Never expose internal ordering identifiers as public action selectors | §8 Revisions integration, canonical/imported units, query-plan evidence and cursor transition; paired server/browser rollback; optional additive indexes may remain |
| WS-16 | Diagnose Timeline recovery/Coordination focus and any shared timing defect; fix owner or fixture | Deterministic response-order cases and ordinary suites demonstrate resolution; focused rerun alone is insufficient | Current routed stateful/a11y rows; revert fix and its characterization together if necessary |
| WS-17 | Reconcile full exported surface, retired implementations, test support imports and source guides | Every retained shared export names a consumer; no deprecated forwarding path; all retired callers migrated | Type/import/lint and affected units; retain removal ledger so rollback restores a coherent consumer set |
| WS-18 | Fresh acceptance and representative tasks, required full affected suites, golden reconciliation if changed, rollout and handoff | All §12 new-iteration exits pass with actual commands/run artifacts and limitations | agent-finalize before broad checks; qualifying retained RESULTS_DIR only; no deployment or reset implied |

**Page contract details:** The HTTP adapter owns opaque cursor encoding and
existing actor/record/limit binding. Revisions receives validated typed limit and
ordering position without depending on the platform pagination package. Select
logical descriptors plus one lookahead; fetch only selected facts and necessary
revision/change-set dependencies. Preserve imported attribution without account
directory access. Validate selected mandatory facts through source projectors.
A continuation cursor whose internal version is unsupported receives the existing
invalid-pagination response and restarts only disposable reads. Captured commands,
idempotency bytes and receipts never depend on a read cursor. Preserve current
schema generation when semantic content is unchanged. No public page-size or
ordering change is planned.

### Next-iteration symbol and retirement ledger

These are baseline dispositions, not completed removals. All entries are TODO
until their named slice proves caller migration and passes its checks. WS-11
expands this ledger to every shared export; WS-17 reconciles the final surface.

| Surface | Current evidence / named consumer | Decision | Slice / completion evidence |
| --- | --- | --- | --- |
| `legacyHistoryReducer`, repeated accepted data and legacy load events | Live controller reconciliation and Timeline external state | remove after migration; do not delete live behavior first | WS-12; single browsing representation and recovery tests |
| `HistoryBinding.reconcile`, operation `currentHistory` | History operation owner and controller; paging erased to RecordHistoryData | replace with mandatory HistoryPage | WS-12; no page-less acceptance or `paging in history` bridge |
| Controller `initialHistory`, `beginMutation`, `commands` inputs | Initial history has no caller; command inputs are declared but not read | remove unused inputs and stale caller arguments | WS-12; repository symbol/caller scan, type checks |
| History types re-exported through inspector model | History operation contracts import inspector-owned state/subject types | import canonical History types directly; keep only true attachment types in inspector | WS-12; dependency direction and named consumer ledger |
| `WorkbookInspectorEditorSlots`, `inspectorSavedValueKind` | Details-local consumers only | privatize unless new pure presentation contract requires a named external consumer | WS-13/17; no gratuitous public helper |
| `WorkbookInspectorPanelSection` | Shell renders it; external tests/policy reference it | keep implementation private; test through admitted public sections | WS-13/17; update ownership policy evidence without removing behavior |
| `InspectorEditDraft`, `InspectorDraftCapture`, `InspectorRelatedRecordFormModel`, `WorkbookInspectorNoticeDestination` | Baseline search found no external production use | privatize after full reference reconciliation | WS-17; preserve owner functionality |
| History request/operation ID types, `HistoryBrowseRequest`, `HistoryAcceptedChain` | State/browsing internals; factories and model remain live | minimize exported type surface; preserve semantic identity constructors | WS-12/17; explicit consumers for retained exports |
| `HistoryTransportOutcome` | Port owner uses it; one external test imports it | prefer testing through the public port; keep private unless a real production boundary needs it | WS-17; no test-driven production export |
| `historyDiffFixture` | Thirteen tests including adapter, shell, owner and Timeline fixtures | move to `apps/web/src/testing/workbookHistoryTestSupport.ts`; migrate all imports; remove old module | WS-17; test-support ownership and no production import |
| `pageRecordHistory` full-list anchor scan and map result boundary | HTTP route after unbounded service read | replace with typed bounded page result and cursor mapping | WS-15; no full snapshot materialization or redundant adapter |
| Source projectors/catalogs, lazy entry-ref allocation | Revisions source catalog and retained rollback selectors | keep; live supported-history behavior | WS-15/17; complete units and unchanged selectors |
| Related-record reducer/controller, History completion feedback and owner-effect port | Related workflow renderer/hook and feature/Timeline bindings | keep or consolidate only with named migrated consumers; short files and legacy test wording do not prove dead code | WS-17; preserve authoring and command ownership |

### WS-11 exported-module baseline

All 50 non-test-suffix TypeScript modules are listed below (the History fixture
is test support despite its old location). Symbols and direct import/re-export
consumers were reconciled against current source. Counts are navigation, not
deletion authority; the named retirement ledger above and WS-17 require final
symbol-level alias/registration review. Re-exports in the History model and page
modules migrate with their canonical owners in WS-12.

| Module | Declared exported symbols | Production consumers / disposition |
| --- | --- | --- |
| `history/HistoryActionLookup.ts` | `HistoryActionLookup` | `workbook/history/WorkbookRecordHistoryOwner.ts`, `workbook/history/workbookHistoryOperation.ts`, `workbook/inspector/useWorkbookRecordHistoryController.ts`, `workbook/inspector/workbookRecordHistoryModel.ts` |
| `history/HistoryLookupFeedback.tsx` | `HistoryLookupFeedback` | `workbook/history/WorkbookHistoryLocalStatus.tsx`, `workbook/history/WorkbookHistoryRecovery.tsx`, `workbook/history/WorkbookHistoryReview.tsx`, `workbook/inspector/WorkbookInspectorRecordHistory.tsx` |
| `history/HistoryPageLookup.ts` | `HistoryLookupState`, `HistoryPageLookup` | `workbook/history/HistoryActionLookup.ts`, `workbook/history/HistoryLookupFeedback.tsx`, `workbook/inspector/useWorkbookRecordHistoryController.ts` |
| `history/WorkbookHistoryContext.ts` | `WorkbookHistoryContext`, `useWorkbookHistoryRuntime`, `useWorkbookHistorySurfaceRefresh`, `useHistoryActionPermission`, `useHistoryRecordPending` | `workbook/WorkbookShell.tsx`, `workbook/features/assessments/useAssessmentWorkbookInspectorComposition.tsx`, `workbook/features/entities/useEntityWorkbookInspectorComposition.tsx`, `workbook/features/generic/useGenericWorkbookInspectorComposition.tsx`, `workbook/history/WorkbookHistoryLocalStatus.tsx`, `workbook/history/WorkbookHistoryRecovery.tsx`, `workbook/history/WorkbookHistoryReview.tsx`, `workbook/inspector/WorkbookInspectorRecordHistory.tsx`, `workbook/inspector/WorkbookRecordHistoryPresentation.tsx`, `workbook/inspector/useWorkbookRecordHistoryController.ts`, `workbook/timeline/hooks/useTimelineCommittedRows.ts`, `workbook/timeline/hooks/useTimelineHistoryActions.ts` |
| `history/WorkbookHistoryLocalStatus.tsx` | `WorkbookHistoryLocalStatus` | `workbook/inspector/WorkbookInspectorRecordHistory.tsx` |
| `history/WorkbookHistoryRecovery.tsx` | `WorkbookHistoryRecovery` | `workbook/WorkbookShell.tsx` |
| `history/WorkbookHistoryReview.tsx` | `WorkbookHistoryReview` | `workbook/components/WorkbookBatchRecovery.tsx` |
| `history/WorkbookRecordHistoryOwner.ts` | `WorkbookRecordHistoryOwner` | `workbook/features/coordination/reconcileDecisionReceipt.ts`, `workbook/features/indicators/reconcileIndicatorCreateReceipt.ts`, `workbook/features/indicators/reconcileIndicatorLifecycleReceipt.ts`, `workbook/features/indicators/reconcileObservationReceipt.ts`, `workbook/history/WorkbookHistoryRecovery.tsx`, `workbook/inspector/useWorkbookRecordHistoryController.ts`, `workbook/runtime/WorkbookMutationRuntime.ts`, `workbook/timeline/actions/reconcileTimelineCaptureReceipt.ts` |
| `history/historyOperationPresentation.ts` | `historyOperationStatus` | `workbook/history/WorkbookHistoryLocalStatus.tsx`, `workbook/history/WorkbookHistoryRecovery.tsx` |
| `history/workbookHistoryBrowsing.ts` | `HistoryReadKind`, `HistoryBrowseRequest`, `HistoryAcceptedChain`, `HistoryBrowsingState`, `initialHistoryBrowsing`, `beginHistoryRead`, `rejectHistoryRead`, `acceptHistoryPage` | `workbook/inspector/useWorkbookRecordHistoryController.ts`, `workbook/inspector/workbookRecordHistoryModel.ts` |
| `history/workbookHistoryItem.ts` | `RecordHistoryRollbackAction`, `RecordHistoryRollbackTarget`, `WorkbookRecordHistoryPendingAction`, `normalizeRecordHistoryData`, `buildRecordRollbackTargetFromHistoryAction`, `historyTargetEqual`, `historyItemContentEqual` | `workbook/history/HistoryActionLookup.ts`, `workbook/history/HistoryPageLookup.ts`, `workbook/history/workbookHistoryBrowsing.ts`, `workbook/history/workbookHistoryOperation.ts`, `workbook/inspector/workbookRecordHistoryModel.ts` |
| `history/workbookHistoryOperation.ts` | `HistoryAuthority`, `HistoryIntent`, `HistoryAttempt`, `HistoryReceipt`, `HistoryTransportOutcome`, `WorkbookRecordHistoryPort`, `HistoryBinding`, `HistoryOperation`, `historyActionPermitted`, `historyOperationLabel` | `workbook/adapters/createWorkbookRecordHistoryAdapter.ts`, `workbook/history/WorkbookHistoryRecovery.tsx`, `workbook/history/WorkbookRecordHistoryOwner.ts`, `workbook/history/historyOperationPresentation.ts`, `workbook/inspector/WorkbookRecordHistoryPresentation.tsx`, `workbook/inspector/useWorkbookRecordHistoryController.ts`, `workbook/inspector/workbookRecordHistoryOperation.ts`, `workbook/inspector/workbookRecordHistoryOwnerEffects.ts`, `workbook/mutations/workbookMutationCommandPorts.ts` |
| `history/workbookHistoryPage.ts` | `HistoryPage`, `HistoryPageRequest`, `HistoryReadScope`, `HistoryPageProvenance`, `sameHistoryReadScope`, `validHistoryPaging` | `workbook/adapters/createWorkbookRecordHistoryAdapter.ts`, `workbook/history/HistoryPageLookup.ts`, `workbook/history/WorkbookRecordHistoryOwner.ts`, `workbook/history/workbookHistoryBrowsing.ts`, `workbook/history/workbookHistoryItem.ts`, `workbook/history/workbookHistoryOperation.ts`, `workbook/history/workbookHistoryReview.ts`, `workbook/inspector/useWorkbookRecordHistoryController.ts`, `workbook/inspector/workbookRecordHistoryModel.ts` |
| `history/workbookHistoryReview.ts` | `HistoryReviewLocator`, `historyReviewAuthorized`, `historyReviewPageLimit` | `workbook/components/WorkbookBatchRecovery.tsx`, `workbook/history/WorkbookHistoryReview.tsx`, `workbook/inspector/useWorkbookRecordHistoryController.ts` |
| `history/workbookHistoryTestFixtures.ts` | `historyDiffFixture` | No production importer; privatize or move after symbol review |
| `inspector/InspectorCreateRelatedWorkflow.tsx` | `InspectorCreateRelatedWorkflow` | `workbook/features/assessments/AssessmentWorkbookInspector.tsx`, `workbook/features/entities/EntityWorkbookInspector.tsx`, `workbook/features/generic/GenericWorkbookInspector.tsx`, `workbook/timeline/components/TimelineWorkbookInspectorSections.tsx` |
| `inspector/WorkbookExplicitPatchRecovery.tsx` | `WorkbookExplicitPatchRecovery` | `workbook/components/EntityWorkbookSurface.tsx`, `workbook/components/GenericWorkbookSurface.tsx`, `workbook/features/entities/useEntityWorkbookInspectorComposition.tsx`, `workbook/features/generic/GenericWorkbookInspectorPresentation.tsx`, `workbook/timeline/components/TimelineInspectorDetails.tsx` |
| `inspector/WorkbookInspectorContextualActions.tsx` | `WorkbookInspectorContextualActions` | `workbook/inspector/WorkbookInspectorDeclaredPanelList.tsx` |
| `inspector/WorkbookInspectorDeclaredPanelList.tsx` | `WorkbookInspectorDeclaredPanelList` | `workbook/features/assessments/AssessmentWorkbookInspector.tsx`, `workbook/features/entities/EntityWorkbookInspector.tsx`, `workbook/features/generic/GenericWorkbookInspector.tsx`, `workbook/timeline/components/TimelineWorkbookInspector.tsx` |
| `inspector/WorkbookInspectorDetails.tsx` | `WorkbookInspectorEditorSlots`, `WorkbookInspectorReadOnlyDetails`, `WorkbookInspectorDetails`, `inspectorSavedValueKind` | `workbook/features/assessments/useAssessmentWorkbookInspectorComposition.tsx`, `workbook/features/entities/useEntityWorkbookInspectorComposition.tsx`, `workbook/features/generic/GenericWorkbookInspectorPresentation.tsx`, `workbook/timeline/components/TimelineInspectorDetails.tsx` |
| `inspector/WorkbookInspectorDraftFeedback.tsx` | `WorkbookInspectorDraftFeedback` | `workbook/features/entities/useEntityWorkbookInspectorComposition.tsx`, `workbook/features/generic/GenericWorkbookInspectorPresentation.tsx`, `workbook/timeline/components/TimelineInspectorDetails.tsx` |
| `inspector/WorkbookInspectorDraftStore.ts` | `InspectorEditIdentity`, `InspectorEditDraft`, `InspectorDraftCapture`, `inspectorEditKey`, `WorkbookInspectorDraftStore` | `workbook/inspector/useWorkbookInspectorEditDraft.ts`, `workbook/inspector/workbookInspectorFieldFeedback.ts`, `workbook/runtime/WorkbookMutationRuntime.ts`, `workbook/timeline/components/TimelineInspectorDetails.tsx` |
| `inspector/WorkbookInspectorEditControl.tsx` | `WorkbookInspectorEditControl` | `workbook/features/entities/useEntityWorkbookInspectorComposition.tsx`, `workbook/features/generic/GenericWorkbookInspectorPresentation.tsx`, `workbook/timeline/components/TimelineInspectorDetails.tsx` |
| `inspector/WorkbookInspectorRecordHistory.tsx` | `WorkbookInspectorRecordHistory`, `HistoryBrowsingControls`, `WorkbookRecordHistoryPanel` | `workbook/features/assessments/AssessmentWorkbookInspector.tsx`, `workbook/features/entities/EntityWorkbookInspector.tsx`, `workbook/features/entities/WorkbookEntityMergeRecovery.tsx`, `workbook/features/generic/GenericWorkbookInspector.tsx`, `workbook/history/WorkbookHistoryRecovery.tsx`, `workbook/history/WorkbookHistoryReview.tsx`, `workbook/timeline/components/TimelineHistoryPanel.tsx`, `workbook/timeline/components/TimelineWorkbookInspectorSections.tsx` |
| `inspector/WorkbookRecordHistoryPresentation.tsx` | `WorkbookRecordHistoryLoadedPresentation` | `workbook/inspector/WorkbookInspectorRecordHistory.tsx` |
| `inspector/canonicalInspectorAdmission.ts` | `admitCanonicalInspectorFeature`, `inspectorFeatureIdentity` | `workbook/features/coordination/CoordinationWorkflowBindings.tsx`, `workbook/inspector/inspectorCapabilityResolver.ts`, `workbook/timeline/actions/timelineMentionAuthority.ts` |
| `inspector/inspectorCapabilityResolver.ts` | `InspectorRecordHistoryAction`, `InspectorContextualCapability`, `inspectorContextualCapabilities`, `inspectorRecordHistoryActions` | `workbook/features/assessments/AssessmentWorkbookInspector.tsx`, `workbook/features/assessments/useAssessmentWorkbookInspectorComposition.tsx`, `workbook/features/entities/EntityWorkbookInspector.tsx`, `workbook/features/entities/WorkbookEntityMergeRecovery.tsx`, `workbook/features/entities/useEntityWorkbookInspectorComposition.tsx`, `workbook/features/generic/GenericWorkbookInspector.tsx`, `workbook/features/generic/useGenericWorkbookInspectorComposition.tsx`, `workbook/history/WorkbookHistoryRecovery.tsx`, `workbook/history/WorkbookHistoryReview.tsx`, `workbook/inspector/WorkbookInspectorContextualActions.tsx`, `workbook/inspector/WorkbookInspectorDeclaredPanelList.tsx`, `workbook/inspector/WorkbookInspectorRecordHistory.tsx`, `workbook/inspector/WorkbookRecordHistoryPresentation.tsx`, `workbook/inspector/presentation/workbookInspectorPresentationModel.ts`, `workbook/timeline/components/TimelineHistoryPanel.tsx`, `workbook/timeline/components/TimelineWorkbookInspector.tsx`, `workbook/timeline/hooks/useTimelineInspectorFeatureController.ts` |
| `inspector/inspectorRelatedRecordModel.ts` | `InspectorRelatedRecordFormModel`, `InspectorRelatedRecordWorkflowState`, `InspectorRelatedRecordWorkflowAction`, `inspectorRelatedRecordWorkflowReducer`, `buildInspectorRelatedRecordDraft` | `workbook/features/assessments/AssessmentWorkbookInspector.tsx`, `workbook/features/coordination/contextualCreateModel.ts`, `workbook/features/coordination/useContextualCreateAttachment.ts`, `workbook/features/coordination/useCoordinationCreateAttachment.ts`, `workbook/features/entities/EntityWorkbookInspector.tsx`, `workbook/features/evidence/useTimelineRelatedEvidenceAttachment.ts`, `workbook/features/generic/GenericWorkbookInspector.tsx`, `workbook/features/notes/useNoteCreateAttachment.ts`, `workbook/inspector/InspectorCreateRelatedWorkflow.tsx`, `workbook/inspector/useInspectorCreateRelatedWorkflow.ts`, `workbook/timeline/components/TimelineWorkbookInspectorSections.tsx` |
| `inspector/prepareWorkbookInspectorChange.ts` | `prepareWorkbookInspectorChange` | `workbook/features/entities/useEntityWorkbookInspectorComposition.tsx`, `workbook/features/generic/useGenericWorkbookInspectorComposition.tsx` |
| `inspector/presentation/WorkbookHistoryPresentation.tsx` | `WorkbookHistoryList`, `WorkbookHistoryEvent` | `workbook/inspector/WorkbookRecordHistoryPresentation.tsx` |
| `inspector/presentation/WorkbookInspectorActions.tsx` | `WorkbookInspectorActionGroup`, `WorkbookInspectorContextualAction`, `WorkbookInspectorActionButton`, `WorkbookInspectorDisabledReasonMessage` | `workbook/components/WorkbookActiveSurfaceFrame.tsx`, `workbook/components/WorkbookAuthoringReferenceControl.tsx`, `workbook/components/WorkbookAuthoringReferencePicker.tsx`, `workbook/components/WorkbookBatchRecordChoices.tsx`, `workbook/components/WorkbookBatchRecovery.tsx`, `workbook/components/WorkbookCandidateBrowsing.tsx`, `workbook/components/WorkbookCandidateSelection.tsx`, `workbook/components/WorkbookQueueOverflowNotice.tsx`, `workbook/components/WorkbookRecoveryPanel.tsx`, `workbook/components/WorkbookReferenceControl.tsx`, `workbook/features/assessments/useAssessmentWorkbookInspectorComposition.tsx`, `workbook/features/coordination/ContextualCreateForm.tsx`, `workbook/features/coordination/ContextualCreateRecovery.tsx`, `workbook/features/coordination/ContextualReferenceControl.tsx`, `workbook/features/coordination/CoordinationCreateForm.tsx`, `workbook/features/coordination/CoordinationCreateRecovery.tsx`, `workbook/features/coordination/CoordinationWorkflowBindings.tsx`, `workbook/features/coordination/DecisionSupersessionEditor.tsx`, `workbook/features/coordination/WorkbookDecisionSupersessionRecovery.tsx`, `workbook/features/entities/WorkbookEntityMergeRecovery.tsx`, `workbook/features/entities/useEntityWorkbookInspectorComposition.tsx`, `workbook/features/evidence/RelatedEvidencePartyControl.tsx`, `workbook/features/evidence/TimelineRelatedEvidenceForm.tsx`, `workbook/features/evidence/TimelineRelatedEvidenceRecovery.tsx`, `workbook/features/generic/GenericInspectorReferenceSummary.tsx`, `workbook/features/generic/GenericWorkbookInspectorPresentation.tsx`, `workbook/features/indicators/IndicatorCanonicalAuthoring.tsx`, `workbook/features/indicators/IndicatorCreateFromObservation.tsx`, `workbook/features/indicators/IndicatorCreateOperationStatus.tsx`, `workbook/features/indicators/IndicatorLifecycleOperationStatus.tsx`, `workbook/features/indicators/IndicatorLifecycleSupportPicker.tsx`, `workbook/features/indicators/IndicatorLifecycleWorkflow.tsx`, `workbook/features/indicators/ObservationCaptureEditor.tsx`, `workbook/features/indicators/ObservationDetails.tsx`, `workbook/features/indicators/ObservationOperationStatus.tsx`, `workbook/features/indicators/ObservationPagingFeedback.tsx`, `workbook/features/notes/NoteAssociationPanel.tsx`, `workbook/features/notes/NoteAssociationRecovery.tsx`, `workbook/features/notes/NoteCreateForm.tsx`, `workbook/features/notes/NoteCreateRecovery.tsx`, `workbook/features/notes/NoteSheetAuthoring.tsx`, `workbook/features/notes/NoteSourceControl.tsx`, `workbook/features/ordinary/OrdinaryCreateNotice.tsx`, `workbook/features/parties/PartyLinkControls.tsx`, `workbook/features/parties/PartyLinkPanel.tsx`, `workbook/features/parties/PartyLinkRecovery.tsx`, `workbook/history/HistoryLookupFeedback.tsx`, `workbook/history/WorkbookHistoryRecovery.tsx`, `workbook/history/WorkbookHistoryReview.tsx`, `workbook/inspector/InspectorCreateRelatedWorkflow.tsx`, `workbook/inspector/WorkbookExplicitPatchRecovery.tsx`, `workbook/inspector/WorkbookInspectorContextualActions.tsx`, `workbook/inspector/WorkbookInspectorDetails.tsx`, `workbook/inspector/WorkbookInspectorDraftFeedback.tsx`, `workbook/inspector/WorkbookInspectorEditControl.tsx`, `workbook/inspector/WorkbookInspectorRecordHistory.tsx`, `workbook/inspector/WorkbookRecordHistoryPresentation.tsx`, `workbook/inspector/presentation/WorkbookInspectorFeedback.tsx`, `workbook/inspector/presentation/WorkbookInspectorShell.tsx`, `workbook/timeline/actions/TimelineCaptureRecovery.tsx`, `workbook/timeline/actions/TimelineMentionRecovery.tsx`, `workbook/timeline/actions/TimelineSupersessionEditor.tsx`, `workbook/timeline/components/TimelineInspectorDetails.tsx`, `workbook/timeline/components/TimelineMentionActionControls.tsx` |
| `inspector/presentation/WorkbookInspectorFeedback.tsx` | `WorkbookInspectorCompactMetadata`, `WorkbookInspectorTechnicalDetails`, `WorkbookInspectorPublicError`, `WorkbookInspectorFeedbackView`, `WorkbookInspectorNoticeView`, `WorkbookInspectorConfirmation` | `workbook/components/WorkbookBatchRecordChoices.tsx`, `workbook/features/assessments/AssessmentWorkbookInspector.tsx`, `workbook/features/entities/EntityWorkbookInspector.tsx`, `workbook/features/entities/useEntityWorkbookInspectorComposition.tsx`, `workbook/features/generic/GenericWorkbookInspector.tsx`, `workbook/features/generic/GenericWorkbookInspectorPresentation.tsx`, `workbook/features/indicators/ObservationPagingFeedback.tsx`, `workbook/features/notes/NoteAssociationRecovery.tsx`, `workbook/features/parties/PartyLinkPanel.tsx`, `workbook/features/parties/PartyLinkRecovery.tsx`, `workbook/history/WorkbookHistoryRecovery.tsx`, `workbook/inspector/InspectorCreateRelatedWorkflow.tsx`, `workbook/inspector/WorkbookExplicitPatchRecovery.tsx`, `workbook/inspector/WorkbookInspectorRecordHistory.tsx`, `workbook/inspector/WorkbookRecordHistoryPresentation.tsx`, `workbook/inspector/presentation/WorkbookHistoryPresentation.tsx`, `workbook/inspector/presentation/WorkbookInspectorPanelContent.tsx`, `workbook/inspector/presentation/WorkbookInspectorShell.tsx`, `workbook/timeline/components/TimelineInspectorDetails.tsx`, `workbook/timeline/components/TimelineMentionsPanel.tsx`, `workbook/timeline/components/TimelineWorkbookInspector.tsx` |
| `inspector/presentation/WorkbookInspectorPanelContent.tsx` | `WorkbookInspectorPanelData`, `WorkbookInspectorPanelContentModel`, `inspectorReadData`, `WorkbookInspectorRegionModel`, `PresentInspectorRegion`, `WorkbookInspectorRegion`, `WorkbookInspectorPanelModel`, `inspectorPanel`, `savedInspectorRegion`, `ownedInspectorRegion`, `WorkbookInspectorPanelContent`, `WorkbookInspectorRegionContent` | `workbook/features/assessments/AssessmentWorkbookInspector.tsx`, `workbook/features/entities/EntityWorkbookInspector.tsx`, `workbook/features/entities/useEntityWorkbookInspectorComposition.tsx`, `workbook/features/evidence/useEvidenceWorkbookBindings.tsx`, `workbook/features/generic/GenericWorkbookInspector.tsx`, `workbook/features/generic/GenericWorkbookInspectorPresentation.tsx`, `workbook/features/generic/useGenericWorkbookInspectorComposition.tsx`, `workbook/features/indicators/IndicatorInspectorWorkflow.tsx`, `workbook/features/indicators/IndicatorLifecycleWorkflow.tsx`, `workbook/features/notes/NoteAssociationPanel.tsx`, `workbook/inspector/WorkbookInspectorDeclaredPanelList.tsx`, `workbook/inspector/WorkbookInspectorRecordHistory.tsx`, `workbook/timeline/components/TimelineHistoryPanel.tsx`, `workbook/timeline/components/TimelineWorkbookInspector.tsx`, `workbook/timeline/components/TimelineWorkbookInspectorSections.tsx` |
| `inspector/presentation/WorkbookInspectorShell.tsx` | `WorkbookInspectorSection`, `WorkbookInspectorShell`, `WorkbookInspectorPanelSection` | `workbook/features/assessments/AssessmentWorkbookInspector.tsx`, `workbook/features/entities/EntityWorkbookInspector.tsx`, `workbook/features/generic/GenericWorkbookInspector.tsx`, `workbook/inspector/WorkbookInspectorDeclaredPanelList.tsx`, `workbook/timeline/components/TimelineWorkbookInspector.tsx` |
| `inspector/presentation/workbookInspectorPresentationModel.ts` | `workbookInspectorNoRowMessage`, `WorkbookInspectorTechnicalField`, `WorkbookInspectorActionBinding`, `WorkbookHistoryEventPresentation`, `bindWorkbookInspectorAction`, `WorkbookInspectorDisabledReason`, `ownerInspectorDisabledReason`, `workbookInspectorDisabledReasonKey`, `workbookInspectorDisabledReasonText`, `workbookInspectorDisabledReason` | `workbook/features/coordination/CoordinationWorkflowBindings.tsx`, `workbook/features/generic/GenericWorkbookInspector.tsx`, `workbook/features/generic/useGenericWorkbookInspectorComposition.tsx`, `workbook/inspector/WorkbookInspectorContextualActions.tsx`, `workbook/inspector/WorkbookInspectorDeclaredPanelList.tsx`, `workbook/inspector/presentation/WorkbookHistoryPresentation.tsx`, `workbook/inspector/presentation/WorkbookInspectorActions.tsx`, `workbook/inspector/presentation/WorkbookInspectorFeedback.tsx`, `workbook/inspector/presentation/WorkbookInspectorShell.tsx`, `workbook/inspector/workbookHistoryPresentationModel.ts`, `workbook/inspector/workbookInspectorErrorModel.ts`, `workbook/timeline/actions/timelineMentionAuthority.ts`, `workbook/timeline/actions/useTimelineCaptureActions.ts`, `workbook/timeline/components/TimelineWorkbookInspector.tsx` |
| `inspector/useInspectorCreateRelatedWorkflow.ts` | `useInspectorCreateRelatedWorkflow` | `workbook/features/assessments/useAssessmentWorkbookInspectorComposition.tsx`, `workbook/features/entities/useEntityWorkbookInspectorComposition.tsx`, `workbook/features/generic/useGenericWorkbookInspectorComposition.tsx` |
| `inspector/useRetainedInspectorRow.ts` | `useRetainedInspectorRow` | `workbook/components/EntityWorkbookSurface.tsx`, `workbook/features/generic/useGenericWorkbookInspectorComposition.tsx`, `workbook/timeline/hooks/useTimelineInspectorSelection.ts` |
| `inspector/useWorkbookInspectorCoordinator.ts` | `useWorkbookInspectorCoordinator` | `workbook/features/assessments/useAssessmentWorkbookInspectorComposition.tsx`, `workbook/features/entities/useEntityWorkbookInspectorComposition.tsx`, `workbook/features/generic/useGenericWorkbookInspectorComposition.tsx`, `workbook/timeline/composition/useTimelineInspectorStateComposition.ts` |
| `inspector/useWorkbookInspectorEditDraft.ts` | `useWorkbookInspectorEditDraft`, `WorkbookInspectorEditDraft` | `workbook/features/entities/useEntityWorkbookInspectorComposition.tsx`, `workbook/features/generic/GenericWorkbookInspectorPresentation.tsx`, `workbook/features/generic/useGenericWorkbookInspectorComposition.tsx`, `workbook/inspector/WorkbookInspectorDetails.tsx`, `workbook/inspector/WorkbookInspectorDraftFeedback.tsx`, `workbook/inspector/WorkbookInspectorEditControl.tsx`, `workbook/inspector/useWorkbookInspectorFieldFeedback.ts`, `workbook/timeline/components/TimelineInspectorDetails.tsx` |
| `inspector/useWorkbookInspectorFieldFeedback.ts` | `useWorkbookInspectorFieldFeedback` | `workbook/features/entities/useEntityWorkbookInspectorComposition.tsx`, `workbook/features/generic/useGenericWorkbookInspectorComposition.tsx`, `workbook/timeline/components/TimelineInspectorDetails.tsx` |
| `inspector/useWorkbookInspectorNotice.ts` | `useWorkbookInspectorNotice` | `workbook/inspector/presentation/WorkbookInspectorFeedback.tsx` |
| `inspector/useWorkbookRecordHistoryController.ts` | `useWorkbookRecordHistoryController` | `workbook/features/entities/WorkbookEntityMergeRecovery.tsx`, `workbook/history/WorkbookHistoryRecovery.tsx`, `workbook/history/WorkbookHistoryReview.tsx`, `workbook/inspector/WorkbookInspectorRecordHistory.tsx`, `workbook/timeline/hooks/useTimelineHistoryActions.ts` |
| `inspector/useWorkbookRecordHistoryFocus.ts` | `useWorkbookRecordHistoryFocus`, `historyActionIdentity`, `historyRollbackActionIdentity` | `workbook/inspector/WorkbookInspectorRecordHistory.tsx`, `workbook/inspector/WorkbookRecordHistoryPresentation.tsx` |
| `inspector/workbookHistoryPresentationModel.ts` | `workbookHistoryEventPresentation`, `workbookHistoryRollbackLabel`, `workbookHistoryPendingTechnicalFields` | `workbook/inspector/WorkbookRecordHistoryPresentation.tsx` |
| `inspector/workbookInspectorErrorModel.ts` | `WorkbookInspectorNoticeDestination`, `WorkbookInspectorNotice`, `WorkbookInspectorNoticeLedger`, `workbookInspectorNoticeIdentity`, `WorkbookInspectorErrorPresentation`, `WorkbookInspectorFeedback`, `targetWorkbookInspectorFeedback`, `workbookInspectorErrorPresentation`, `workbookInspectorMessageFeedback`, `workbookInspectorOperationFailureFeedback`, `workbookInspectorLocalErrorFeedback`, `workbookInspectorLocalErrorPresentation` | `workbook/components/EntityWorkbookSurface.tsx`, `workbook/features/assessments/AssessmentWorkbookInspector.tsx`, `workbook/features/assessments/WorkbookAssessmentAuthoringOwner.ts`, `workbook/features/assessments/useAssessmentWorkbookInspectorComposition.tsx`, `workbook/features/entities/EntityWorkbookInspector.tsx`, `workbook/features/entities/useEntityClipboardPasteController.ts`, `workbook/features/entities/useEntityMergeController.ts`, `workbook/features/entities/useEntityWorkbookInspectorComposition.tsx`, `workbook/features/generic/GenericWorkbookInspector.tsx`, `workbook/features/generic/GenericWorkbookInspectorPresentation.tsx`, `workbook/features/generic/useGenericWorkbookInspectorComposition.tsx`, `workbook/features/indicators/ObservationPagingFeedback.tsx`, `workbook/features/notes/NoteAssociationRecovery.tsx`, `workbook/features/notes/WorkbookNoteAssociationOwner.ts`, `workbook/features/parties/PartyLinkPanel.tsx`, `workbook/features/parties/PartyLinkRecovery.tsx`, `workbook/history/WorkbookRecordHistoryOwner.ts`, `workbook/hooks/useEntityTimelinePreview.ts`, `workbook/hooks/useGenericSurfaceMutationController.ts`, `workbook/inspector/WorkbookExplicitPatchRecovery.tsx`, `workbook/inspector/inspectorRelatedRecordModel.ts`, `workbook/inspector/presentation/WorkbookInspectorFeedback.tsx`, `workbook/inspector/presentation/WorkbookInspectorPanelContent.tsx`, `workbook/inspector/useInspectorCreateRelatedWorkflow.ts`, `workbook/inspector/useWorkbookInspectorFieldFeedback.ts`, `workbook/inspector/useWorkbookInspectorNotice.ts`, `workbook/inspector/useWorkbookRecordHistoryController.ts`, `workbook/inspector/workbookRecordHistoryModel.ts`, `workbook/inspector/workbookRecordHistoryOperation.ts`, `workbook/runtime/WorkbookExplicitPatchOwner.ts`, `workbook/timeline/components/TimelineMentionsPanel.tsx`, `workbook/timeline/components/TimelineWorkbookInspector.tsx`, `workbook/timeline/composition/useTimelineInspectorWorkflowComposition.ts`, `workbook/timeline/hooks/useTimelineCreateRelatedWorkflow.ts`, `workbook/timeline/hooks/useTimelineEvidenceAttach.ts`, `workbook/timeline/hooks/useTimelineInspectorFeatureController.ts`, `workbook/timeline/hooks/useTimelineInspectorSelection.ts`, `workbook/timeline/hooks/useTimelineKeyboardController.ts`, `workbook/timeline/hooks/useTimelineMentionActions.ts` |
| `inspector/workbookInspectorFieldFeedback.ts` | `InspectorFieldFeedbackIdentity`, `inspectorFieldFeedback` | `workbook/inspector/useWorkbookInspectorFieldFeedback.ts` |
| `inspector/workbookInspectorSubject.ts` | `WorkbookInspectorSubject`, `WorkbookInspectorLiveSubject`, `WorkbookInspectorLiveRowBinding`, `buildWorkbookInspectorSubject`, `updateWorkbookInspectorSubject`, `workbookInspectorSubjectsEqual` | `workbook/features/assessments/AssessmentWorkbookInspector.tsx`, `workbook/features/assessments/useAssessmentWorkbookInspectorComposition.tsx`, `workbook/features/coordination/WorkbookContextualTaskDecisionCreateOwner.ts`, `workbook/features/coordination/WorkbookCoordinationCreateOwner.ts`, `workbook/features/coordination/contextualCreateModel.ts`, `workbook/features/coordination/coordinationCreateModel.ts`, `workbook/features/coordination/useContextualCreateAttachment.ts`, `workbook/features/coordination/useCoordinationCreateAttachment.ts`, `workbook/features/entities/EntityWorkbookInspector.tsx`, `workbook/features/entities/useEntityWorkbookInspectorComposition.tsx`, `workbook/features/evidence/WorkbookTimelineRelatedEvidenceOwner.ts`, `workbook/features/evidence/timelineRelatedEvidenceModel.ts`, `workbook/features/evidence/useTimelineRelatedEvidenceAttachment.ts`, `workbook/features/generic/GenericWorkbookInspector.tsx`, `workbook/features/generic/useGenericWorkbookInspectorComposition.tsx`, `workbook/features/indicators/IndicatorLifecycleWorkflow.tsx`, `workbook/features/notes/WorkbookNoteCreateOwner.ts`, `workbook/features/notes/noteCreateModel.ts`, `workbook/features/notes/useNoteCreateAttachment.ts`, `workbook/history/WorkbookHistoryRecovery.tsx`, `workbook/history/workbookHistoryOperation.ts`, `workbook/inspector/WorkbookInspectorDeclaredPanelList.tsx`, `workbook/inspector/WorkbookInspectorRecordHistory.tsx`, `workbook/inspector/WorkbookRecordHistoryPresentation.tsx`, `workbook/inspector/inspectorRelatedRecordModel.ts`, `workbook/inspector/presentation/WorkbookInspectorShell.tsx`, `workbook/inspector/useInspectorCreateRelatedWorkflow.ts`, `workbook/inspector/useWorkbookInspectorCoordinator.ts`, `workbook/inspector/useWorkbookRecordHistoryController.ts`, `workbook/inspector/useWorkbookRecordHistoryFocus.ts`, `workbook/inspector/workbookRecordHistoryModel.ts`, `workbook/models/workbookInspectorModel.ts`, `workbook/timeline/components/TimelineWorkbookInspector.tsx`, `workbook/timeline/components/TimelineWorkbookInspectorSections.tsx`, `workbook/timeline/focus/timelineInspectorElementRegistry.ts`, `workbook/timeline/hooks/useTimelineCreateRelatedWorkflow.ts`, `workbook/timeline/hooks/useTimelineHistoryActions.ts`, `workbook/timeline/hooks/useTimelineInspectorFeatureController.ts`, `workbook/timeline/models/timelineHistoryModel.ts` |
| `inspector/workbookRecordHistoryModel.ts` | `WorkbookRecordHistoryRequestId`, `WorkbookRecordHistoryOperationId`, `WorkbookRecordHistoryState`, `WorkbookRecordHistoryEvent`, `initialWorkbookRecordHistoryState`, `workbookRecordHistoryRequestId`, `workbookRecordHistoryOperationId`, `workbookRecordHistoryLoadedData`, `workbookRecordHistoryPendingAction`, `workbookRecordHistoryFeedback`, `workbookRecordHistoryLoadError`, `workbookRecordHistoryReducer` | `workbook/adapters/createWorkbookRecordHistoryAdapter.ts`, `workbook/history/WorkbookHistoryReview.tsx`, `workbook/history/WorkbookRecordHistoryOwner.ts`, `workbook/history/workbookHistoryOperation.ts`, `workbook/inspector/WorkbookInspectorRecordHistory.tsx`, `workbook/inspector/WorkbookRecordHistoryPresentation.tsx`, `workbook/inspector/useWorkbookRecordHistoryController.ts`, `workbook/inspector/useWorkbookRecordHistoryFocus.ts`, `workbook/inspector/workbookHistoryPresentationModel.ts`, `workbook/timeline/components/TimelineHistoryPanel.tsx`, `workbook/timeline/components/TimelineWorkbookInspectorSections.tsx`, `workbook/timeline/hooks/useTimelineHistoryActions.ts`, `workbook/timeline/hooks/useTimelineHistoryState.ts`, `workbook/timeline/hooks/useTimelineInspectorSelection.ts`, `workbook/timeline/hooks/useTimelineKeyboardController.ts`, `workbook/timeline/models/timelineHistoryModel.ts`, `workbook/timeline/presentation/useTimelineInspectorPresentation.tsx` |
| `inspector/workbookRecordHistoryOperation.ts` | `workbookRecordHistoryCompletionFeedback` | `workbook/inspector/useWorkbookRecordHistoryController.ts` |
| `inspector/workbookRecordHistoryOwnerEffects.ts` | `WorkbookRecordHistoryOwnerEffects` | `workbook/features/assessments/AssessmentWorkbookInspector.tsx`, `workbook/features/entities/EntityWorkbookInspector.tsx`, `workbook/features/generic/GenericWorkbookInspector.tsx`, `workbook/inspector/WorkbookInspectorRecordHistory.tsx`, `workbook/inspector/useWorkbookRecordHistoryController.ts` |

Owner decisions for this iteration:

- G12, G13 and G14 need no new normative decomposition: existing History,
  component and source-boundary owners suffice. Source guides describe the new
  interfaces after migration; product tests consume typed inputs, never Markdown.
- G15 clarifies Core 03 REQ-03-299: transformations do not create authority;
  genuine fresh same-render observations and ordinary window eviction retain
  their distinct treatment under existing REQ-03-100/299 lifetimes.
- G16 clarifies Core 01 REQ-01-056: page boundaries preserve complete logical
  items and avoid unrelated snapshot materialization. Core 02 REQ-02-218/265
  continues to own source facts/catalogs; no source schema or new decoder changes.
- Core 01 REQ-01-051/053/054 protects deterministic item order and stable public
  selectors. The existing coalescing/order algorithm is frozen as the chosen
  internal paging contract because it preserves audit/reversal identity.
- Cursor position versioning is History-local inside the current protected
  envelope. The shared pagination version and semantic representation generation
  remain unchanged; unsupported History positions restart disposable reads.
- G17 needs causal implementation/test evidence, not new product requirements.
  G18 is scoped evidence/handoff work. Design §12.7 and Domain need no amendment.

### Previous-iteration section 7 record (historical)

### Decisions to adopt in WS-00

The user approved these defaults. WS-00 adopts them in design §12.7 and Core 01
REQ-01-052A, Core 02 REQ-02-218/265 and Core 03 REQ-03-139. The explicit
implementation decisions above supersede historical conditional compatibility
recommendations. The History representation is `cartulary.history_diff.v1`,
canonical snapshots only, no partial-detail success and no legacy adapter.

| Concern / owner | Complete recommended rule |
| --- | --- |
| Navigation destinations / design §12.7 with Core 03 §2.3A | Derive from the active config and the same subject/access admission used by rendering. Include readable unrequested/failed sections; omit undeclared, concealed and unavailable-by-subject sections. Missing required implementation is an error, not omission. No subject: only explicitly admitted creation sections. Zero destinations: no Sections control, retain truthful no-row/creation status. |
| Active section / design §§7.3,12.7 | An explicit opener destination wins if admitted. Reopening the same subject preserves active panel under Core 03; a new subject/schema without explicit destination starts at first admitted panel. If active section disappears, use first remaining admitted panel. Passive scroll selects the last heading at/above the usable body top, or first visible heading when none precedes it; at body end select last admitted section. No automatic focus or fetch. |
| Section selection / design §§8.4–8.5,12.6 | Use a labelled disclosure containing ordinary semantic navigation controls, not tab roles. Enter/Space activates; Tab visits destinations. Selection closes disclosure, scrolls only inspector body, and focuses the heading or existing owner entry control. Unrequested History targets Open history without issuing a read. Escape closes the topmost popup before editor/inspector handling. Escape dismissal restores trigger focus; pointer outside dismissal preserves destination focus. |
| Transient lifetime / design §12.7 | Navigation does not unmount contributions; same-subject section selection retains an attached ordinary editor without submitting it. Retarget clears old-subject presentation and menu state. Scope disclosure selection to schema/record/field or item identity; presentation invalidation must not retire raw authoring or captured operations. On concealment remove content, counts, labels and navigation/focus destinations together. Restore focus only when the concealed region owned it, to a safe surviving control. |
| Geometry / design §§7.3,14 | Retain 420px default, 360px minimum, `min(560px, 45vw)` maximum and existing overlay bands/clamps. One body scrollport and bounded two-line title with explicit full-title access. Sections/Close remain visible. No persistent inspector preference. |
| Field classification / design §12.7 | Short scalar values use aligned rows; declared multiline/narrative values use full-width rows; unknown presentation stacks safely. Never omit/reorder fields for layout. Prefer wrapping; for lengthy narrative use a six-rendered-line preview with explicit Show full value / Show less controls. Project the line limit through the authored design contract if adopted. No truncation of editable control, retained-error or required recovery content. |
| Editor actions / design §§12.4–12.7 | Unsaved change region keeps accepted value visible. Owner supplies Update, Close editor, value-level Clear and distinct Discard bindings; Close explains retention. Exactly one ordinary editor attaches. Detached cues come from draft owner. Specialized workflows keep their own submit/review/detach policy. |
| Relationship disclosure / design §12.7 | Summaries always show raw mention and semantic resolution state; authorized target only when known. One selected correction region per current source collection binding; read/paging controls are local to picker. Collapsing hides controls, not pending/uncertain/retained-status or owner-required recovery entry. |
| History / Core 01 §3.3.4.2, Core 03 §10, design §§12.7,14 | One row per stable logical item; server order/paging. Display actor label if authorized, otherwise attributed ID, plus operation/description and absolute UTC time with numeric offset and unambiguous date. Exact committed timestamp and required typed units are accessible in event detail. No relative-only time, inferred diff or text-derived action target. |
| History wire/retained data / Core 01–02 | Core 01 REQ-01-052A owns closed units and explicit absent/null/present values. Validate canonical snapshots under REQ-02-265; reject unsupported payloads locally. No partial-detail success, schema-less translation or legacy adapter. |
| Workflow and creation / design §§12.4,12.7 | Show declared actions in order with outcome description. Saved subject remains the record title; local attached operation title identifies append/create/transition. Creation without a subject has explicit mode and source context. Primary treatment belongs to the active local affirmative decision; no panel-global Save. |
| Evidence / design §§11,12.7 | Separate accepted metadata, available access commands and attachment work. Counts never become item lists. File chooser, drop and file paste share owner admission/validation, prevent accidental duplicate activation, and have keyboard parity. Lifecycle, availability and access remain independent. |
| Unknown/additive facts / respective owner | Reuse existing capability omission rules; don't interpret unknown feature labels. History unknown unit/version handling is explicitly versioned and safe: no guessed value or action, preserve readable known authorized facts only when the adopted decoder allows it. Unsupported mandatory payload versions fail the read locally. |

### Slice delivery details

File groups refer to §2; this table is the concrete implementation handoff, not
permission to make those changes in the planning task.

| Slice | Likely authored files / boundary | Characterization and changes | Verification / completion | Rollback |
| --- | --- | --- | --- | --- |
| WS-00 | Design owner and targeted Core 01/03 History clauses; domain navigation only if needed | Resolve §7 decisions; map current obligations versus new design behavior; specify unit semantics and retained-data policy without mirroring helpers | Human owner review; Markdown lint; trace all gaps to exits; no machine product check reads prose | Revert proposed owner changes before implementation, or coordinate with dependent slices after adoption |
| WS-01 | Shared actions/form styles, shell typography, affected Timeline/Indicator style callers; authored design contracts if needed | Use complete roles; consolidate actual variants; retain semantic legends; characterize computed state/density rather than literal style object shape | Focused `package.ui`/`web.workbook`, frontend type/import checks; generation/drift if inputs changed; all common callers migrated | Revert authored and generated projections plus callers together |
| WS-02 | Shell, declared list, feature shell bindings, existing semantic focus integration; layout only as needed | Introduce one admitted section descriptor seam; keep body mounts; test menu/scroll/retarget/access/overlay transitions | Focused presentation tests plus browser focus/geometry; all G01 exits | Revert navigation and its bindings, keep operation owners unchanged |
| WS-03 | Details, editor, feedback, Timeline/Generic/Entity adapters | Explicit editor content/action slots; aligned fields/full value; owner-derived retained cues; remove duplicate Update placement | Draft/revision and Details rows; inspector edit browser/a11y; G02/G05 exits | Revert presentation/adapters as one unit, preserving draft identity/data |
| WS-04 | Mention panels/actions and candidate presentation integration | Compact summaries; selected correction disclosure; focus restore/reveal; qualified bounded-target status; no second operation store | Mention transition/reconciliation suites and browser relationship scenarios; G07 exit | Revert presentation with same owner attempts and semantic item IDs |
| WS-05 | Timeline Evidence and affected existing Evidence presentations | Distinct accepted/read/access/attachment regions; chooser/drop/paste admission; no new metadata enumeration | Evidence panel and attachment/recovery tests; public browser/a11y/visual cases; G09 exit | Revert visual grouping; never clear pending upload/link state as cleanup |
| WS-06 | Revisions OpenAPI owner, typed unit projection, History model/materializer/service, source provider contributions, revisionassembly and browser History adapter | Enumerate source families; add pure source-owned semantic projectors; close schema; preserve selectors/order; review imported attribution parity; test retained-data and session-generation cutover | Revisions/source-owner slices; protocol generation/drift; schema and OpenAPI compatibility; service-backed history/security cases; G10 exit | Coordinated server/client/schema revert; avoid data rewrite; no legacy adapter; deploy and roll back the bundled server/browser together |
| WS-07 | History event mapper/renderer/actions and focus binding | Visible actor/time, complete typed unit disclosure, event-local rollback, Record actions group; remove redundant operation labels | History mapping/browsing/recovery and browser reversal/a11y; G03 History/G04 exits | Revert UI/protocol consumer coherently with WS-06 if required |
| WS-08 | Contextual actions renderer and authored action presentation metadata where needed | Ordered action rows with result descriptions, existing reason association and canonical command binding | Capability/order/reason and browser action dispatch checks; G03 Workflow exit | Revert action layout and labels with callers/tests, no aliases |
| WS-09 | Assessment/Indicator, Note, coordination, Task/Decision, related-Evidence and fallback forms | Migrate family by family; record consumed shared styles; distinguish selected record from operation/creation; delete dead local styles | Each changed feature's authoring/recovery tests and representative browser flow; §8 schema audit complete | Revert each completed family coherently; no mixed duplicate form authority |
| WS-10 | Tests, authored routing/fixture registry, reconciled goldens, affected source guides and this tracker | Add missing evidence, reconcile owner acceptance, inspect every changed golden, document migration/retirement and residual limitations | All §8 gates and §12 product exits; final tracker/handoff update is the last required slice | Withhold cutover on blockers; record exact successful deployment-compatible rollback bundle |

A style export shared with non-inspector callers is retained only until those
callers are intentionally migrated or it remains a justified grid-specific role.
Do not delete it merely to satisfy a textual cleanup count. Other scope additions
require a separately identified gap/owner/exit here before work starts.

## 8. Validation plan

### Next-iteration validation routes and acceptance

The three public task guides (`web.workbook`, `module.workbook`,
`module.revisions`) were refreshed successfully for this document update.
They are routing evidence only. Their planning results are historical; current
implementation results are recorded in §10. Refresh routing at each implementation entry; add new
characterization rows to authored catalogs rather than deriving tests from this
Markdown. §8's older command table remains historical.

| Layer | Public command / selection | Required when | Current result |
| --- | --- | --- | --- |
| Routing | `make task-guide ROLE=module-author OWNER=web.workbook`; repeat for `module.workbook` and `module.revisions` | WS-11 and every affected slice entry | PASS; current document task |
| History unit | `make test-slice OWNER=web.workbook ROWS=web.workbook.regression.history_browsing_state,web.workbook.regression.history_browsing_characterization,web.workbook.regression.history_recovery_characterization,web.workbook.regression.history_recovery_surfaces,web.workbook.regression.history_timeline_coordination` | WS-12; augment current rows for new cases | Not run; future implementation |
| Inspector browser | `make service-backed-test-slice OWNER=module.workbook ROWS=module.workbook.browser.inspector_subject_retention` plus affected edit/creation/accessibility rows | WS-13/14 | Not run |
| Revisions | `make test-slice OWNER=module.revisions`; `make service-backed-test-slice OWNER=module.revisions ROWS=<current-paging-import-rollback-rows>` | WS-15 | Not run; current paging row includes `module.revisions.integration.history_pagination_remains_bound_to_record_id_re_a92fcdb263` |
| Static | `make frontend-typecheck`; `make frontend-import-boundary-check`; `make lint-biome` | Frontend changes, including fixture moves | Not run |
| Authored projections | `make generate`; `make generate-drift`; `make generated-artifact-policy-check`; `make json-shape-check`; `make openapi-compatibility-check` | Corresponding contracts/catalogs change; never edit generated outputs directly | Not run |
| Index/schema | `make migration-drift` plus refreshed owning migration/database slices | Only if WS-15 justifies a new additive index | Not run; no migration selected or applied |
| Reliability | Narrow `service-backed-test-slice` rows resolved by task guides; Coordination includes `module.workbook.accessibility.coordination_create_authoring_recovery` | WS-16; replay controlled response ordering and ordinary affected suites | Not run; prior reruns remain historical |
| Final implementation | `make agent-finalize`, then applicable `make test-fast`, `make browser-e2e-stateful`, `make browser-e2e-webserver-backed`, `make browser-e2e-a11y`, `make browser-e2e-visual` | WS-18, broaden only for changed ownership and unresolved risk | PASS; WS-18 current receipts below; earlier tracker-only skip remains historical |
| Golden maintenance | Ordinary reconciliation, `make browser-e2e-visual-update`, review every changed image, two successful ordinary visual runs | Only when applicable images intentionally change | Not run; no golden edits |
| Documentation | `make lint-markdown`; `git diff --check`; `git diff --cached --check` | Current tracker update and later source docs | PASS; configured lint plus explicit tracker lint, with exact scope/results in §10 |

Use `make explain-test-owner`, `make explain-target` and `make explain-run` when
routing or failures need investigation. There is no automatic full application
release/check requirement. Retained-run maintenance is skipped unless a qualifying
successful full warm check `RESULTS_DIR` is supplied. No such directory is selected
in this task; the presence of old focused runs does not qualify them.

### Next-iteration acceptance ledger

I3 identifiers separate new acceptance from historical A001–A027. Status is TODO
until implementation supplies evidence; planning/source inspection is not PASS.
WS-18 must separately reassess every applicable digest A001–A027 row at the new
candidate, with owner/scope rationale for N/A and no mandatory blocked outcome.

| ID | Gap / slice | Binary acceptance | Status / evidence required |
| --- | --- | --- | --- |
| I3-01 | G18 / WS-11 | All 74 target files and connected boundaries have dispositions; required owner clarifications adopted before dependent changes | PASS; WS-11 owner amendments, 50-module exported baseline and unchanged 74-file inventory; lint run 20260921T135352Z-p69697 |
| I3-02 | G12 / WS-12 | One accepted History browsing representation, mandatory page reconciliation, no legacy bridge or duplicate Timeline read reducer | PASS; source retirement, fourteen behavior rows and static checks; WS-12 exit |
| I3-03 | G12 / WS-12 | Explicit reads, requested-change lookup, paging, stale/unavailable states, generation restart and late responses preserve operation owners | PASS; lookup, browsing, operation, recovery and shell integration rows; integrated browser coverage reassessed in WS-18 |
| I3-04 | G14 / WS-13 | Saved/empty/creation modes are closed; readonly reading requires no editor callbacks; one admitted section sequence | PASS; closed props, component matrix and fresh browser evidence; WS-13 exit |
| I3-05 | G14 / WS-13 | All 17 schemas and specialized source families preserve field order, value distinctions, header, focus, disclosure and draft behavior | PASS; seventeen-schema saved-value/order tests, source composition, drafts and browser checks; WS-13 exit |
| I3-06 | G15 / WS-14 | Retained observations match record and authority provenance; cloning, conversion, stale versions and late responses cannot establish authority | PASS; source/retention regressions and ordinary browser evidence; WS-14 exit |
| I3-07 | G15 / WS-14 | Account/incident loss, same-account recovery, off-window retention and concurrent fresh evidence obey existing lifetimes | PASS; delayed reads/writes, revoked scope, role/closure/resumption and off-window matrix; WS-14 exit |
| I3-08 | G16 / WS-15 | Page windows preserve logical coalescing, ordering, imports, units, attributed actors, selectors and reversal scopes | PASS; 24 source configurations, equal-time/split/coalesced events, import and reversal rows; WS-15 evidence |
| I3-09 | G16 / WS-15 | Invalid cursors avoid materialization; old versions restart disposable reads; deployment/rollback preserve drafts, bytes and receipts | PASS; pre-materialization rejection, legacy-position unit and production browser restart/recovery; WS-18 coordinated final reassessment |
| I3-10 | G16 / WS-15 | Small-page reads do not materialize all snapshots as unrelated retained history grows; selected change-set dependency cost reported separately | PASS; 11/101/1001 events with bounded fetched/projected counts and separate selected-change-set costs; WS-15 artifact |
| I3-11 | G17 / WS-16 | Relevant intermittent failures have causes, deterministic regressions and successful ordinary affected suites | PASS; controlled Timeline pre-fix failure/post-fix pass and Coordination late-observation focus regression; ordinary affected suites pass |
| I3-12 | G13 / WS-17 | Unused paths removed, all fixture imports moved and retained shared exports have named production consumers/purposes | PASS; complete symbol ledger, fixture migration, ownership/import policy and fifteen affected unit rows |
| I3-13 | All / WS-18 | Relevant roles, incident closure, concealment, source replacement, uncertain/acknowledged writes and sibling failures pass | PASS; final complete server-backed and stateful aggregates, final fast and named source/recovery suites below |
| I3-14 | G14/G18 / WS-18 | Narrow/compact/overlay/below-minimum widths, densities, resize, zoom, long content, text spacing, reduced motion and keyboard/screen-reader paths remain usable | PASS; measurement 25/25, a11y 20/20, visual 12/12, nine fresh narrow/zoom/recovery captures and seven matched inspector/mention goldens reviewed |
| I3-15 | G18 / WS-18 | Applicable digest acceptance, retirement, coordinated rollout/rollback and final handoff are evidence-backed | PASS; all A001–A027 reviewed, final receipts, paired rollback proof, 173-path and 163-symbol ledgers, limitations and next maintainer action recorded |

Compare the same five representative tasks: locate Evidence; edit/close/resume;
correct a mention; inspect History and reversal scope; create a related task.
Record actual completion, incorrect actions, scroll travel and interpretation of
saving/closing only when observed. Scripted checks are not participant results;
unavailable participant measurements remain a usability hypothesis, not an
invented readiness failure or claimed measured improvement.

### Previous-iteration section 8 record (historical)

### Discovered public command routes (planning record)

This subsection records the pre-implementation selection, not current execution
status. WS-00–09 checkpoints and the WS-10 final evidence below supersede its
"Not run" entries.

Task guides for `web.workbook`, `module.workbook`, `package.ui` and
`module.revisions` ran successfully in this planning pass. `explain-test-owner`
confirmed 288 `web.workbook` rows and nine `package.ui` rows at this baseline.
These are routing observations, not tests or permanent inventory requirements.

| Layer | Public command | Scope / when required | Planning execution |
| --- | --- | --- | --- |
| Routing | `make task-guide ROLE=module-author OWNER=<active-owner>`; `make explain-test-owner OWNER=<active-owner>` | Refresh at slice entry; use authored catalog to select changed behavior | Four task guides and two owner explanations passed |
| Narrow frontend | `make test-slice OWNER=web.workbook ROWS=<selected-active-rows>` | Each presentation slice; add new behavior rows to authored catalog before relying on them | Not run; implementation pending |
| UI projections/selectors | `make test-slice OWNER=package.ui` | WS-01 and any later UI projection/selector change | Not run |
| Revisions | `make test-slice OWNER=module.revisions`; `make service-backed-test-slice OWNER=module.revisions ROWS=<selected-active-rows>` | WS-06/07, plus each changed source owner's current guide | Not run |
| Browser | `make service-backed-test-slice OWNER=module.workbook ROWS=<selected-active-rows>` | Changed stateful/webserver-backed scenarios per current routing | Not run |
| Static | `make frontend-typecheck`; `make frontend-import-boundary-check`; `make lint-biome` | Authored frontend changes and shared boundary migration | Not run |
| Generation | `make generate`; `make generate-drift`; `make generated-artifact-policy-check`; `make json-shape-check` | Only when corresponding authored contracts/catalogs change; include all generated roots declared by policy | Not run |
| HTTP compatibility | `make openapi-compatibility-check` | WS-06 against immutable release history, before choosing cutover | Not run |
| Finalization | `make agent-finalize` | Before broader implementation verification; pass RESULTS_DIR only for an actual qualifying successful full warm run | Skipped for documentation-only planning; it includes mutating generated-structure refresh |
| Accessibility / visuals | `make browser-e2e-a11y`; `make browser-e2e-visual` | Shared UI changes and WS-10 | Not run |
| Golden refresh | `make browser-e2e-visual-update`, then ordinary visual validation | Only after ordinary-run reconciliation and functional correctness; inspect every changed image | Not run |
| Broader end-to-end | `make browser-e2e-webserver-backed`; `make browser-e2e-stateful`; `make test-fast` | WS-10 only where changed owner breadth or unresolved risk requires beyond focused rows | Not run; no automatic full `make check` requirement for planning |
| Documentation | `make lint-markdown`; `git diff --check` | This tracker and later documentation changes | Pending P-03 |

Current relevant row examples, revalidate before execution:

- `web.workbook.regression.inspector_persistent_context`,
  `inspector_explicit_panel_states`, `inspector_grouped_permission_reasons`,
  `inspector_draft_binding`, `inspector_draft_lifetime`,
  `inspector_field_feedback_identity`, `inspector_timeline_explicit_details`,
  `history_inspector_recovery`, `history_recovery_characterization` all use the
  `web.workbook.regression.` prefix (apply it to each abbreviated name).
- `module.workbook.browser.inspector_subject_retention`,
  `module.workbook.browser.inspector_detached_refresh`,
  `module.workbook.browser.inspector_exact_replay` and
  `module.workbook.accessibility.inspector_edit_recovery`.
- History browser routing also belongs to `module.revisions`, including
  `module.revisions.browser.verify_history_and_rollback_preview_action_use_p_4201ff80a8`.
  Do not assume every inspector browser row belongs to `module.workbook`.

These existing rows do not yet assert every proposed behavior. Add targeted tests
for navigation, field disclosure, actor/unit presentation, source projectors and
specialized modes; avoid tests that merely mirror helper decomposition.

### Coverage and acceptance matrix

| Axis | Required cases / outcome | Evidence responsibility |
| --- | --- | --- |
| Schema coverage | Timeline; Hosts; Identities; Evidence; Notes; Indicators; Assessments; Task Requests; Decisions; Parties; Communications Log; Handoff; Status Review; Lesson; Findings; Investigative Queries; Forensic Keywords | Composition/capability table covers all 17 configurations. Optional surfaces are tested when implemented, otherwise documented N/A under Core 01; no silent omission. |
| Surface families | Direct behavioral review of Timeline, Entity, Generic, Assessment and Indicator compositions; each specialized authoring family has its own lifetime/recovery check | WS-03–09; tests can share fixtures, not substitute one owner's semantics for another's |
| Read state | Unrequested, initial loading, explicit empty/populated, refreshing with retained data, stale failure, unavailable; independently failing sibling regions | No invented absence, automatic read-all or disabled unrelated actions |
| Authority | Viewer/editor/reviewer/admin where relevant; open/closed incident; concealed data; incident-access loss, same-account recovery and account replacement | Contents, counts, navigation, feedback and late-result admission agree with current authority |
| Authoring/operation | Pristine, dirty, invalid, detached, review required, uncertain, accepted but unrefreshed, saved but unlinked; duplicate invocation and late completion | Captured identity/revision/request bytes stay with existing owners; recovery does not resend accepted writes |
| Geometry | 1280×720, 1024×720, 768×640, below-minimum containment; 360/420/maximum panel widths where admitted; vertical resize; all supported densities; 200% zoom; long text and text-spacing override; reduced motion | No document overflow, obscured focus or lost Close/status-strip access; compare actual owner fixture parameters |
| Keyboard/accessibility | Open/choose/dismiss sections, editor submit/detach, picker nesting, event disclosure, review/confirm/cancel, return to original field/grid; screen-reader names/announcements | Preserve semantic focus identity; one announcement per admitted transition/outcome; non-color state and supported contrast |
| History semantics | Each required unit family, current and retained snapshots, imported attribution, unknown/malformed version, concealed target/value, all advertised rollback scopes, pagination and eligibility change | Source-owner and Revisions integration tests plus production-renderer browser tests |
| Visual evidence | First-viewport saved Details, editing, retained draft, selected relationship, Evidence/recovery, History units/reversal, Workflow/specialized creation, narrow/read-only/error | Follow visual maintenance guide; original base-inspector screenshot did not prove first-viewport Details |
| Boundary/security | No presentation imports of request/mutation owners, no direct grid vendor imports, no raw history snapshots/object bytes/handles in generic diff, no prose-consuming executable tools | Source/import check, payload fixtures and human owner-to-projection review |

Assess every digest row A001–A027 in the implementation handoff, with actual
`PASS` evidence, `N/A` rationale, or `BLOCKED` reason. These are acceptance
outcomes, not work-item statuses. A004–A011, A014, A016–A020, A022–A026 are central;
A012–A013 and A015 apply to exposed mutation/recovery paths. A021 requires evidence
of grid continuity for shared layout changes; performance measurement is needed
only if virtualization/geometry behavior changes or a regression is suspected.
A001–A003 and A027 cover authority, repository state and handoff. Nothing is
silently dropped because this began as a visual audit.

For comparative task review, use representative records and the same five tasks:
locate Evidence; edit a scalar, close and resume; resolve/correct a mention;
inspect a History change and reversal scope; create a related task. Record task
completion, incorrect actions, scroll travel and Saved/Update/Close interpretation
for baseline and candidate. Required task paths must be completable without an
unintended mutation or inaccessible control. Record qualitative results honestly;
no invented time-saving threshold or conformance claim. Lack of participants must
be recorded as a remaining usability hypothesis, not an automated usability pass.

## 9. Top-level implementation tracker

### Current iteration implementation tracker

WS-11 through WS-18 are DONE under the user's implementation instruction. P-04/P-05 remain
historical document checkpoints. The older DONE table is unchanged for WS-00–WS-10.

| ID | Work item / gap | Status | Depends on | Evidence now | Exit condition |
| --- | --- | --- | --- | --- | --- |
| WS-11 | Owner, inventory and acceptance closure; G18 | DONE | none | Core 01 REQ-01-056 and Core 03 REQ-03-299 clarified; 74-file/50-module baseline, owner dispositions and public lint passed | I3-01 PASS; exit saved before WS-12 |
| WS-12 | Single History read authority; G12 | DONE | WS-11 | One browsing authority, mandatory HistoryPage, explicit coordination and migrated consumers; focused/static/generation checks pass | I3-02/03 PASS; exit saved before WS-13 |
| WS-13 | Closed presentation contracts; G14 | DONE | WS-12 | Closed shell variants, independent saved renderer and explicit editor presentation; all 17 schemas and affected browser rows pass | I3-04/05 PASS; exit saved before WS-14 |
| WS-14 | Accepted-row provenance; G15 | DONE | WS-13 | Source-accepted record/version/read scope; normalized evidence and source-only minting; authority, lifetime and browser rows pass | I3-06/07 PASS; exit saved before WS-15 |
| WS-15 | Bounded History paging; G16 | DONE | WS-14 | Typed bounded selection; 24 source configurations, imports, selectors, rollback, browser restart and growth measurements pass | I3-08/09/10 PASS; exit saved before WS-16 |
| WS-16 | Deterministic reliability; G17 | DONE | WS-15 | Timeline identity recovery settles once; terminal rejection settles immediately; failed Coordination refresh remains deliberate; controlled and ordinary suites pass | I3-11 PASS; exit saved before WS-17 |
| WS-17 | Dead-code and export retirement; G13 | DONE | WS-16 | Fixtures moved, forwarding exports and incidental types retired; 163 retained symbols have named consumers | I3-12 PASS; exit saved before WS-18 |
| WS-18 | Validation and handoff completion; G18 | DONE | WS-11–17 | Current fast/browser/stateful/measurement/a11y/visual, paired builds, static/contract checks and handoff pass; earlier failures retain causal dispositions | I3-13/14/15 PASS; mandatory final exit saved; no successor |

### Previous-iteration section 9 record (historical)

Implementation is authorized. Update the row and
append its evidence to §10 after completing each workstream, before the next starts.
If the exit fails, keep it `IN_PROGRESS` or `BLOCKED` with a concrete reason.

| ID | Work item | Status | Technical dependencies | Evidence to attach | Binary exit |
| --- | --- | --- | --- | --- | --- |
| WS-00 | Specification cleanup and decision closure | DONE | none | Design §12.7 and D-AC-082–086; Core 01 REQ-01-052A, Core 02 REQ-02-218, Core 03 REQ-03-139; lint and whitespace pass | Approved defaults and canonical-only History adopted without contradiction |
| WS-01 | Shared roles and authored projections | DONE | WS-00 | Shared roles and generated narrative limit; focused/UI/type/import/drift checks passed | G06 common-control migration complete; integrated visual evidence follows in WS-10 |
| WS-02 | Record frame and section navigation | DONE | WS-01 | Unit, browser, type/import/lint evidence in §10 | G01 navigation/focus exit passed; integrated matrix in WS-10 |
| WS-03 | Details and ordinary edit hierarchy | DONE | WS-02 | Saved reading, retained draft, browser and accessibility evidence in §10 | G02/G05 behavioral exits passed; integrated visual matrix remains WS-10 |
| WS-04 | Relationship summaries and correction | DONE | WS-02 | Source-local unit/browser evidence in §10 | G07 behavioral exits passed |
| WS-05 | Evidence reading and attachment | DONE | WS-02 | Access/lifecycle and attachment/link/recovery evidence | G09 complete |
| WS-06 | History semantic data contract and cutover | DONE | WS-00 | Source-family inventory, unit fixtures, compatibility and retained-data report | G10 complete |
| WS-07 | History reading and corrective actions | DONE | WS-02, WS-06 | Unit/actor/time and reversal/focus evidence | G04 and G03 History portion complete |
| WS-08 | Workflow action hierarchy | DONE | WS-02 | Declared action/order/reason/dispatch evidence | G03 Workflow portion complete |
| WS-09 | Specialized mode and form migration | DONE | WS-03, WS-05, WS-07, WS-08 | Per-family migration and source/receipt evidence | G08 complete; dead alternatives removed |
| WS-10 | Validation and handoff completion | DONE | WS-00–09 | Final 708/708 fast, all required browser scenarios, reviewed 255-capture reconciliation and two ordinary 12/12 visual runs; full evidence in validation handoff | G11 and all §12 exits complete; A001–A027 PASS; paired bundles and limitations recorded |
| O-01 | New rich Evidence enumeration/read capability | DEFERRED | Separate adopted owner scope | None; not required by this plan | Do not silently add to WS-05 or count as missing presentation work |

## 10. Session handoff log

### Current implementation checkpoints — 2026-09-21

| Workstream | Checkpoint | Evidence / next action |
| --- | --- | --- |
| WS-11 | IN_PROGRESS — entry | User authorized the complete serial implementation plan. Baseline `8af6bba8737e52ad545b4425879c92cf45cfd1bb`; only this tracker was staged before execution. Existing staged content is preserved. Clarify observable ownership, reconcile inventory and lint selection, then record the passing exit before WS-12. |
| WS-11 | DONE — exit | Changed this tracker, Core 01 REQ-01-056, Core 03 REQ-03-299 and `.markdownlint-cli2.jsonc`. All 74 files retained in the inventory; 50 non-test-suffix modules and their declarations/direct production consumers added to §7. No owner conflict, data/schema migration or new capability. G12–14/G17 need no normative decomposition; Design/Domain unchanged. Three public owner task guides passed, `git diff --check` and `git diff --cached --check` passed; `make lint-markdown` passed at `.cartulary/test-results/20260921T135352Z-p69697/adhoc/lint-markdown/tool-run-summary.json` including this tracker. Product checks/generation/finalizer skipped for this owner/document-only slice. I3-01 PASS. Next: enter WS-12. |
| WS-12 | IN_PROGRESS — entry | WS-11 exit saved. Migrate HistoryPage reconciliation and all consumers, remove competing read phases and Timeline reducer, preserve captured requests/receipts and FIFO coordination. |
| WS-12 | DONE — exit | G12 and I3-02/03 pass. One browsing state owns accepted pages, continuation, provenance and read failure; inspector state owns subject/review/submission attachment. Reconciliation requires HistoryPage. Shared state hook removes the unused local/Timeline duplicate reducer. Canonical page/item imports and neutral ports/WorkbookRecordSubject migrate every consumer. Production requires runtime coordination; standalone injection requires owner and coordinator. Legacy reducer/load events/page-less acceptance and unused beginMutation/commands interfaces are removed. Receipt reconciliation remains operation-owned; focus waits for that completion boundary. |

WS-12 changed `apps/web/src/workbook/{history,inspector,timeline,features,
components,models,ports,surfaces,adapters}/` and connected shell tests; authored
`tools/frontend_source_ownership.json` and `tools/test_families/web.workbook.json`
register the two new modules and presentation-state regression row. The generated
execution render index was refreshed through `make generate`, never edited by hand.
Source guides explain the shared page/subject boundary. Repository callers migrate
atomically; there is no persisted-data or wire change. Captured operation identity,
exact replay, acknowledgement and Timeline FIFO behavior remain covered.

WS-12 validation used the public Make targets with the pinned local Node PATH:

- `make test-slice OWNER=web.workbook ROWS=<fourteen affected History/inspector rows,
  plus ownership policy>`: fourteen behavior rows passed in
  `.cartulary/test-results/20260921T141601Z-p97437`. Exact selected row IDs and
  per-test results are retained in that run. The additional policy row failed an
  obsolete `config.panels.map` source-shape assertion (the existing admission
  owner uses flatMap). Updated it to require config-derived sections and delivery
  of the same sections to the consumer; existing forbidden-owner checks remain.
  Its ordinary row passes at `20260921T141701Z-p5450`.
- `make frontend-typecheck`: PASS `20260921T141601Z-p97495`.
- `make frontend-import-boundary-check`: PASS `20260921T141601Z-p97528`.
- `make lint-biome`: PASS `20260921T141601Z-p97546`.
- `make generate`: PASS `20260921T140737Z-p82408` after sorting the new selector
  titles in the authored catalog. `make generate-drift`: PASS
  `20260921T141609Z-p483`; `make generated-artifact-policy-check`: PASS
  `20260921T141609Z-p485`; `make json-shape-check`: PASS `20260921T141609Z-p487`.
- `make format-frontend`, `git diff --check`, `git diff --cached --check`: PASS.

All run IDs above are under `.cartulary/test-results/`; graph summaries and
`unit-logs/` retain commands/results. Migration failures in `20260921T140800Z-*`
were stale fixture props, an overbroad models-directory import restriction and
premature rollback focus. Moving the pure subject to ports preserves that
restriction; focus now waits for authoritative reconciliation. Subsequent
`20260921T141311Z-p90735`/`141312Z-p90886` exposed a duplicate test helper, removed;
`141449Z-p94557`/`141450Z-p94718` exposed two moved imports and the need to keep
monotonic record metadata reconciliation in the browsing owner, both repaired.
These failed aggregates remain failures. Fresh focused rows cover explicit reads,
continuation, failed refresh retaining provenance, generation restart, late reads,
bounded lookup, retargeting, exact replay, acknowledgement before refresh and
read-only recovery. Integrated browser/visual/a11y checks remain WS-18 obligations;
no unrelated broad run or finalizer was substituted for this slice's evidence.
Next action: save WS-13 IN_PROGRESS before changing presentation interfaces.

WS-13 IN_PROGRESS — entry: WS-12 exit saved. Close saved/empty/creation shell
variants, split saved rendering from explicit edit presentation, migrate all
source families together and validate the seventeen-schema matrix.

WS-13 DONE — exit: G14/I3-04/05 pass. Changed
`inspector/presentation/WorkbookInspectorShell.tsx`,
`inspector/WorkbookInspectorDetails.tsx`, new
`inspector/WorkbookInspectorSavedDetails.tsx`, their presentation tests,
Generic/Entity/Assessment inspector compositions, Timeline inspector composition,
Entity editing fixture, inspector README and authored frontend source ownership.
Shell variants require a subject plus admitted sections, an empty heading without
sections, or source-matching creation context plus local sections. Removed generic
children, nullable saved subjects and competing headings/mode defaults. Navigation
and mounting still consume the same admitted sequence. The independent saved
renderer needs no editor callbacks; the editable attachment uses an explicit pure
presentation type, not a hook return type. Append-only Assessment and all-schema
reading tests use the saved renderer directly. No compatibility wrapper remains.

WS-13 validation (run roots under `.cartulary/test-results/`):

- Fresh `make task-guide ROLE=module-author OWNER=web.workbook` and
  `OWNER=module.workbook`: PASS.
- `make test-slice OWNER=web.workbook ROWS=web.workbook.regression.inspector_persistent_context,web.workbook.regression.inspector_saved_reading,web.workbook.regression.inspector_entity_editing,web.workbook.regression.inspector_draft_binding,web.workbook.regression.inspector_draft_lifetime,web.workbook.regression.inspector_explicit_panel_states,web.workbook.regression.inspector_timeline_explicit_details,web.workbook.boundary_support.workbooksurfaceownershippolicy_suite_85a208f2dd`:
  PASS 9/9 units, `20260921T142402Z-p12921`. Covers all 17 schemas, null/zero/false/
  empty/unloaded values, concealed sections, header/navigation, disclosure,
  focus, saved values, explicit editing and draft lifetimes.
- `make test-slice OWNER=module.workbook ROWS=module.workbook.frontend_unit.verify_active_view_schema_id_selects_inspector_c_9c4dd5ce7c,module.workbook.frontend_unit.verify_active_view_schema_id_selects_inspector_c_fed994e037,module.workbook.frontend_unit.verify_inspector_panels_and_feature_groups_rende_b947f0007c`:
  PASS 4/4 units, `20260921T142534Z-p48832`; lifecycle, empty selection,
  configuration changes and related creation source composition.
- `make service-backed-test-slice OWNER=module.workbook ROWS=module.workbook.browser.inspector_subject_retention,module.workbook.browser.inspector_persistent_header`:
  PASS 13/13 units, `20260921T142402Z-p12954`. Browser reports are under
  `browser-e2e-webserver-backed/browser-groups/functional-support-default-workbook-inspector-edit/`
  and `functional-support-default-note-associations/`. Edit/close/resume and header
  access under narrow layouts, zoom and spacing pass.
- `make frontend-typecheck`: PASS `20260921T142533Z-p48677`;
  `make frontend-import-boundary-check`: PASS `20260921T142402Z-p13016`;
  `make lint-biome`: PASS `20260921T142533Z-p48687`. `make format-frontend`,
  `git diff --check` and `git diff --cached --check`: PASS.

Initial `20260921T142149Z-*` migration failures identified nullable test subjects,
a creation attachment string treated as an object, obsolete test children and the
Entity editing fixture's missing History provider. All were migrated; production
retains required coordination. A banned empty-object type in a type assertion
failed lint at `20260921T142402Z-p13034`; replaced with the precise partial type.
Failed runs remain failures. Compatibility is repository-only and coordinated:
no wire/storage changes or data migration. Invalid shell combinations are removed;
saved values and authoring owners retain their lifetimes. No goldens changed.
Broader a11y/visual integration and final documentation lint remain WS-18 duties.
Next: save WS-14 IN_PROGRESS before provenance implementation.

WS-14 IN_PROGRESS — entry: WS-13 exit saved. Replace reference retirement with
source-accepted record/version/authority evidence, migrate connected query and
mutation observations, preserve conversions and test authority transitions.

WS-14 intermediate evidence: source adapters now attach record/version and the
existing runtime authority epoch; conversions preserve evidence and the retention
hook filters candidates before version precedence. Source-only minting is enforced
by an authored import boundary. Focused source/retention tests passed at
`20260921T144852Z-p67388` (6/6) and `20260921T144853Z-p67608` (2/2).
Static checks passed at `20260921T144924Z-p99342` (types), `p99346` (imports),
`p99352` (Biome); generation/drift/policy/shape passed. Browser aggregate
`20260921T144854Z-p68000` failed 12/17: queries dispatched before incident authority
acceptance were correctly rejected but left the query owner waiting. Infrastructure
now replaces its source query port when accepted authority/session changes, allowing
the existing query lifecycle to reread. This is an affected-owner blocker until the
ordinary browser rows pass. A runtime lifetime test initially failed after its new
incident-closure setup left creation authority suspended; the fixture now explicitly
restores authority before authoring. The corrected lifetime/shell rows passed at
`20260921T145252Z-p19417` (5/5). No failure was treated as a successful exit. Follow-up browser aggregate
`20260921T145252Z-p19393` also failed 12/17. Diagnostic trace
`20260921T145609Z-p60703` established a second cause: mutation coordination resumed
while a successful query was pending and advanced the shared mutation epoch from
0 to 1. Read acceptance must not use that coordination epoch. The existing runtime
now owns a separate read-revocation counter, advanced on protected-read suspension;
resumption, incident closure and readable role changes preserve it. There is no new
authorization registry. The retention hook no longer holds its own suspended-scope
blacklist: source provenance owns revocation. Temporary diagnostic logging was
removed. Runtime regression covers resumption, role changes, closure and later
session loss; source/retention/lifetime rows passed at `20260921T145904Z-p98473`
(7/7), connected Mention/Decision rows at `20260921T145934Z-p30672` (3/3). The ordinary browser
aggregate `20260921T145904Z-p98505` passed Entity Find, Timeline Find and persistent
header, but failed the inspector edit row (15/17 units overall). Its trace identified
React maximum update depth after the accepted write: Entity conversion allocated a
new presentation on every render, retriggering draft observation. Conversion now
memoizes each source candidate independently; provenance filtering still precedes
version precedence. The ordinary inspector edit row subsequently passed at `20260921T150138Z-p37352` (11/11 units).



These execution checkpoints supersede the planning-only next-session statements
below. Historical run results remain historical; new results are recorded here.

### Next-iteration document update — 2026-09-21

Only this tracker is changed. The user requested implementation of the document
update plan, explicitly reserving production changes for subsequent work.
The archived `docs/archive/workbook-inspector-validation-handoff.md` is not
needed and will not be recovered or treated as a blocker. Historical completed
checkpoints below remain historical, including their old handoff references.

#### Current scope and authority handoff

| State | Files inspected / touched | Commands or review | Result / next action |
| --- | --- | --- | --- |
| P-04/P-05 DONE | AGENTS, framework, two active skills, Domain/design/Core 01–04, NLSpec research, digest navigation and this tracker | `git status --short`; `git rev-parse HEAD`; owner/source reads and current 74-file inventory | Clean baseline 8af6bba8; only tracker authorized to change; production WS-11–18 TODO |
| Scope decisions recorded | Current user plan and clarification | Inspector subsystem rather than application release; include bounded backend paging; archived handoff excluded | No planning blocker; WS-11 is the next implementation action after later authorization |


WS-14 DONE — exit: I3-06/07 pass. Changed source contracts/helpers are
`query/{WorkbookQueryRow,WorkbookViewQueryPort,workbookRowObservation,acceptWorkbookRowObservation,workbookQueryRowPatch}.ts`;
source adapters cover view/targeted reads, queued/explicit/batch writes, Assessment,
Indicator, Note, Party, Evidence, contextual/ordinary creation and Timeline mention
reading/entity creation. Infrastructure captures dispatch authority and refreshes
query readers at accepted-authority changes. Runtime and explicit-patch owners
supply the existing scope and read-revocation epoch. Generic, Entity, Assessment
and Timeline admissions filter candidates before version selection. Contract-row
and Timeline normalizers preserve matching observations; entity conversion retains
raw source rows. The retention hook's object/scope retirement blacklists are gone.
Source import policy, ownership/routing inputs, fixtures, authority tests and
inspector/ports source guidance were updated together.

Validation (all paths under `.cartulary/test-results/`): `make generate` passed at
`20260921T144815Z-p63941`; `make generate-drift`,
`make generated-artifact-policy-check`, `make json-shape-check` passed at
`20260921T144937Z-p805`, `p810`, `p816`. Focused `make test-slice` runs and exact rows
are recorded above; additional Party/Mention checks passed at
`20260921T150255Z-p71886` (3/3), Timeline action/query adapter rows at
`20260921T150305Z-p72901` (3/3). Final `make frontend-typecheck`,
`make frontend-import-boundary-check`, `make lint-biome` passed at
`20260921T150255Z-p71941`, `p71949`, `p71961`. `git diff --check` passed.
Browser command was `make service-backed-test-slice OWNER=module.workbook
ROWS=module.workbook.browser.inspector_subject_retention,module.workbook.browser.inspector_persistent_header,module.workbook.browser.find_authority,module.workbook.browser.entity_find_authority_host`.
The last three rows passed in aggregate `20260921T145904Z-p98505`, whose inspector
failure remains a recorded aggregate failure. The corrected inspector row passed
alone in `20260921T150138Z-p37352`; its ordinary Playwright report is
`browser-e2e-webserver-backed/browser-groups/functional-support-default-workbook-inspector-edit/playwright-report.json`.
Other reports use the corresponding `functional-support-default-{entity-find,timeline-find,note-associations}` group.

Compatibility: internal evidence only; no wire/schema/data migration. Same-account
recovery retains captured requests and receipts, but old read evidence cannot
cross a revocation. Role and closure changes preserve authorized reading. No
forced reload, historical rewrite or golden changes. Delivery risks found here
(startup read replacement, mutation/read epoch conflation, unstable Entity
conversion) have causal fixes and successful ordinary affected checks. Broader
cross-owner and layout coverage remains the explicit WS-18 obligation, not a
claimed pass here. Next: save WS-15 IN_PROGRESS before bounded paging changes.


WS-15 IN_PROGRESS — entry: WS-14 exit saved. Validate History cursors before
materialization, select bounded logical descriptors, and preserve canonical
coalescing, attribution, selectors and legal reversal dependencies.

WS-15 intermediate evidence: typed `HistoryQuery` now carries validated limit and
ordering position; `HistoryResult` returns selected items and continuation. HTTP
owns the unchanged envelope and a History-specific protected position version,
rejecting malformed/legacy positions before invoking History. Repository selection
fetches `limit + 1` metadata descriptors before selected snapshots, excludes the
lookahead from projection and uses complete change-set association metadata for
coalescing. The full-list anchor scan, application serialized maps and full-list
page assembler are removed. Current eligibility still reads selected change-set
reversal dependencies. Unit rows passed at `20260921T151102Z-p82713` (3/3); initial
repository/envelope/pagination service rows passed at `20260921T151102Z-p82733`
(5/5). New logical-page and existing cursor-binding integration rows passed in
`20260921T151736Z-p44075` (see their row evidence), while the aggregate failed its
new scaling fixture. Fixture failures at `20260921T151611Z-p17529`,
`20260921T151736Z-p44075` and `20260921T151915Z-p63144` were SQL parameter casts and
use of generic catalog fixture projectors instead of the fixture's collection and
Host display-name schema. The measurement regression is being corrected and rerun;
I3-10 remains open. Generation passed at `20260921T151703Z-p36023`; drift,
generated-policy and JSON shape checks passed at `20260921T151954Z-p82333`,
`p82335`, `p82337`. No index or migration has been added.


WS-15 validation checkpoint: final bounded/coalescing/selector/rollback integration
selection passed at `20260921T152425Z-p30476` (5/5 units), and source-projector,
materializer/attribution and cursor unit rows passed at `p30490` (3/3), then after
shared source-fixture extraction at `20260921T153239Z-p60802` (3/3). The measurement
artifact is `20260921T152425Z-p30476/module.revisions/history-paging-measurements.json`
beneath `.cartulary/test-results/`. At 11/101/1001 retained events, limit 1 fetched
2 descriptors, 1 mutation, 1 required revision and made 2 projector calls, with
zero eligibility dependency rows for a deleted record. Allocated bytes were
44048/25688/25688; descriptor plans used a 25 KB top-N heapsort and took
0.106/0.125/0.527 ms. These are local observations, not a production latency claim.
A separate selected-change-set reversal check fetched 2 dependency rows and
returned the expected precondition failure for its deliberately non-reversible
fixture. Existing indexes suffice at these measured sizes; no migration selected.
Later measurement-fixture corrections were UUID/text parameter typing at
`20260921T152050Z-p87549` and explicit expected reversal precondition at
`20260921T152216Z-p6772`; neither failed a production paging assertion.

The four production History overflow/keyboard/invalid-cursor/later-rollback browser
rows and acknowledged-delete failed-refresh recovery passed (5 scenarios, 13/13
units) at `20260921T153259Z-p79248`, via the five corresponding
`module.revisions.browser.history_browsing_*` / `history_recovery_fdfce91b2066`
rows. The all-source integration matrix initially failed two fixture shapes at
`20260921T153238Z-p60568`: Links' strict retained-value validator requires complete
canonical fields and target identities. Those authored fixtures are corrected;
all-source and imported attribution pagination checks remain pending.
`make lint-go-format` and `make lint-go-vet` passed (`20260921T153312Z-p421`,
`20260921T153316Z-p10219`). Staticcheck's touched fixture copy warning was fixed;
`20260921T153440Z-p37502` still fails pre-existing ST1005 at
`internal/app/workbookassembly/note_associations.go:62` (unchanged file).


WS-15 DONE — exit saved before WS-16. The complete source-family pagination matrix
passes (24 configurations) at `20260921T153729Z-p83990` (3/3 units). Its immediately
preceding attempt `20260921T153438Z-p37081` failed service readiness before tests
(`infra/service_readiness_timeout`); no product failure is hidden by the rerun.
The final semantic-source unit row passes at `20260921T153730Z-p84213` (1/1).
Imported History pagination, attributed actors, stable selectors and current
rollback actions pass at `20260921T153618Z-p65125` (3/3), command
`make service-backed-test-slice OWNER=module.incidentbundles
ROWS=module.incidentbundles.integration.incident_bundle_import_reuses_the_shared_upload_698c072755`.
The family matrix uses `module.revisions.integration.history_source_family_paging`;
canonical ordering/bounded/rollback selections and browser rows are recorded above.
Generation passed at `20260921T153222Z-p53338`; final drift/policy/shape pass at
`20260921T153512Z-p56049`, `p56051`, `p56053`. `git diff --check` passes.

Changed boundaries: Revisions `application_ports.go`, `command_service.go`,
`history_{model,repository,service,materializer}.go`, HTTP `routes.go` and
`history_pagination.go`; focused component/paging/source/integration tests,
`internal/app/revisionassembly/history_projection_test.go`, shared pure authored
`internal/testutil/historytest/fixtures.go`, Incident Bundles integration helpers,
Revisions README, authored Revisions test catalog and generated topology index.
Removed the full-list materializer/page assembler, anchor scan and serialized-map
application result. No compatibility shim, schema migration, history rewrite,
semantic generation bump or new index. Unsupported History positions restart
inside the unchanged protected cursor envelope. Drafts and captured operation
ownership are independent of disposable reads; production browser recovery passes.
Delivery and rollback require coherent application bundles as described in the
Revisions README; WS-18 rechecks the final candidate. Remaining risk: descriptor
selection still scans/sorts metadata as history grows; measured snapshots and
projectors are bounded, and selected reversal dependency costs remain explicit.
The unrelated pre-existing staticcheck finding remains recorded, not claimed pass.


WS-16 IN_PROGRESS — entry: WS-15 DONE checkpoint saved. Reproduce the retained
Timeline transaction-recovery editor timeout and Coordination recovery-focus
failure with controlled ordering and observable completion; fix the responsible
owner and require ordinary affected suites. No assertion weakening or retry-only
closure.


WS-16 causal checkpoint: retained Timeline trace `20260921T073129Z-p96952`
shows the next row's click completed but the prior row retained an editor labeled
"Conflict on Activity Synopsis" after Saved. The driver published a conflict
settlement on a recoverable halt and later published acceptance to the same
listener. Promise consumers could only retain the first settlement. The new
controlled owner regression fails on that exact premature conflict at
`20260921T155111Z-p53868` (1/2 units). The fix leaves the logical command pending
through halt; Retry/Discard or definitive conflict settles it once. No request is
resent by presentation. The same regression now passes, covering both recovery
outcomes, at `20260921T155203Z-p55264` with the Coordination regression (3/3).
The ordinary Timeline browser test now gates the retry transport and asserts the
acknowledged editor has detached before the next edit, exposing completion directly.

The retained Coordination trace `20260921T073453Z-p32834` shows Retry refresh
focused, followed by another aborted read and loss of focus. A late own socket
observation restarted a failed read, disabling that focused control. The owner
now retains failed refresh as explicit recovery; in-flight reads drain additional
observations without publishing an intermediate enabled control. A controlled
regression sends socket evidence during and after failure, asserts enabled focus,
one mutation, and successful read-only recovery. Ordinary affected accessibility
and Timeline rows passed before the Timeline settlement amendment at
`20260921T154443Z-p10040` (13/13, 3 scenarios); a fresh post-amendment run is pending.
An initial regression-fixture assertion used an unavailable matcher at
`20260921T154226Z-p8051`; replacing it with native disabled/activeElement assertions
passed at `20260921T154343Z-p9137` (2/2). This was test plumbing, not a product pass.


WS-16 DONE — passing exit saved before WS-17. Changed Timeline mutation driver,
its existing recovery fixture/tests and README, the public Timeline browser scenario,
Coordination creation owner/recovery tests, authored web regression routing and its
generated topology projection. The post-fix ordinary Timeline public-route and
Coordination accessibility selection passed at `20260921T155215Z-p55996` (13/13
units; 3 scenarios). Command: `make service-backed-test-slice OWNER=module.workbook
ROWS=module.workbook.accessibility.coordination_create_authoring_recovery,module.workbook.browser_stateful.verify_timeline_public_mutations_remount_safe_tr_84f645d6c5`.
Queue/retry/discard shell regressions pass at `20260921T155255Z-p87228` (2/2),
`make test-slice OWNER=module.workbook
ROWS=module.workbook.frontend_unit.verify_sync_engine_pending_queue_orders_creates_999bdc7b60`.
The two controlled regression rows passed at `20260921T155203Z-p55264`:
`web.workbook.regression.timeline_editor_recovery_settlement` and
`web.workbook.regression.coordination_create_recovery`.
Frontend type/import/Biome passed at `20260921T155216Z-p56255`, `p56259`, `p56265`;
generation at `20260921T155101Z-p51008`, drift/policy/shape at
`20260921T155316Z-p92243`, `p92245`, `p92247`; whitespace passes.

Compatibility: no wire/data/transaction identity change. The queue remains the
attention owner; pending editor settlement now follows the operation's actual
lifetime. Refresh failure remains visible until explicit read recovery, while late
observations accumulate all necessary refresh debt. No sleeps, increased retries,
weakened assertions or Network Flow edits. Residual risk is addressed by WS-18's
broader sibling/authority/browser matrix; these two retained failures have causal
owner fixes and successful ordinary evidence rather than unchanged reruns alone.


WS-17 IN_PROGRESS — entry: WS-16 DONE checkpoint saved. Move History fixtures into
existing application test support, reconcile exports against imports/re-exports,
registrations and policy consumers, and retire unnecessary public boundaries.


WS-17 DONE — exit saved before WS-18. The [final symbol ledger](workbook-inspector-symbol-ledger.tsv) reconciles 163 retained exports across 56 current modules and 28 moved/retired pre-iteration declarations. It names each consumer and purpose; the 74-file inventory and per-module export table above remain the historical baseline. Added modules and the moved fixture are represented in the final ledger. Named imports (including aliases), type imports, inline type imports, re-exports, namespace/dynamic import searches, source registrations and policy consumers were reviewed; no indirect consumer required a retired export.

Both `historyDiffFixture` and `historyPresentationFixture` now live in `apps/web/src/testing/workbookHistoryTestSupport.ts`; all fourteen consuming test files migrated and the old production module is deleted. Source ownership is `web.testing`; the authored production import policy expressly forbids this support module. Removed forwarding exports from HistoryActionLookup, workbookHistoryOperation and workbookHistoryPage; callers import canonical declarations. Privatized browsing internals, transport outcome, draft capture/value, related form, notice destination, Details slots/props and shell props/section. Tests exercise the public shell and infer its prop/port types. Shared source projectors, selector allocation, live workflow components, editing attachment and recovery ports retain named purposes. No deprecated alias or compatibility shim remains.

Changed paths: History/inspector declarations and their consumers/tests, the moved application fixture, History source guide, authored frontend source/import policies, generator-produced topology index, and the final ledger. Compatibility is internal-only: repository callers migrate atomically, with no wire/storage/data migration. Rollback must restore declarations and consumers together; removal does not alter captured requests or receipts.

Validation (all PASS, run roots under `.cartulary/test-results/`):

- `make format`; `make generate` (`20260921T155951Z-p4438`).
- `make test-slice OWNER=web.workbook ROWS=web.workbook.boundary_support.workbooksurfaceownershippolicy_suite_85a208f2dd,web.workbook.regression.history_action_lookup,web.workbook.regression.history_browsing_characterization,web.workbook.regression.history_browsing_state,web.workbook.regression.history_captured_transport,web.workbook.regression.history_operation_owner,web.workbook.regression.history_presentation_state,web.workbook.regression.history_recovery_surfaces,web.workbook.regression.history_timeline_coordination,web.workbook.regression.inspector_saved_reading,web.workbook.regression.batch_history_review,web.workbook.regression.workbook_history_presentation_model_af6496ddc4,web.workbook.regression.workbookshell_builds_rollback_targets_only_from_3a8ffdc380,web.workbook.regression.workbookshell__inspector_keeps_the_timeline_insp_b0ac344c3e`: 15/15 units (`20260921T160033Z-p7623`).
- `make frontend-typecheck frontend-import-boundary-check lint-biome`: `20260921T160033Z-p7734/p7745/p7761`.
- `make generate-drift generated-artifact-policy-check json-shape-check`: `20260921T160113Z-p10839/p10841/p10843`.

I3-12 PASS. Remaining risk is integrated regression coverage, owned by mandatory WS-18; no known retirement defect remains.

WS-18 IN_PROGRESS — entry: WS-11–WS-17 are DONE with saved serial exits. Run finalizer before broader verification, reassess I3-01–I3-15 and A001–A027, inspect visual evidence, establish coordinated rollout/rollback and complete the current handoff. No qualifying successful full warm check is supplied as RESULTS_DIR; retained-run maintenance will be reported skipped.

### WS-18 integrated validation findings (current)

The final slice added the previously missing concurrent-addition case to `TestHistoryLogicalPageBoundaries_Integration`: after page one, a newer committed event does not shift, omit or duplicate continuation items; a fresh chain discovers it. `make service-backed-test-slice OWNER=module.revisions ROWS=module.revisions.integration.history_logical_page_boundaries` passed 3/3 (`20260921T160406Z-p71483`). No production paging change was needed.

The first broad checks exposed additional affected-owner gaps; they are fixed here rather than relabeled unrelated or closed by unchanged reruns:

| Failure / original artifact | Supported cause and remediation | Regression / exit evidence |
| --- | --- | --- |
| `make test-fast`, `20260921T160254Z-p20077`, 707/710; Timeline action sequencing | A partial Mark reviewed receipt advanced raw row version but retained old-version provenance. Inspector admission correctly rejected it, preventing subsequent supersession after a stale query. The transport now attaches dispatch-time accepted provenance; committed/visible row projections preserve it only with the same admitted base identity, version and scope. Partial old-scope receipts cannot relabel unrelated saved fields. | Existing production-shell stale-query sequence passes (`20260921T160913Z-p60414`); transport regression controls a late response across session replacement and asserts original provenance (`20260921T161111Z-p95429`); coordinator suite passes (`20260921T161118Z-p2586`). |
| Same fast run; Assessment filter and production Recovery fixture rows; isolated reproduction `20260921T160954Z-p64129` | The filter fixture held the second request by ordinal, accidentally holding legitimate startup authority reads. It now holds only the explicitly triggered unfiltered query. Recovery restored individual operation authorities but left the shared read authority suspended; the fixture now restores it as the production shell does. Assertions are unchanged. | Both ordinary fixture rows pass, 3/3 (`20260921T161059Z-p75125`). |
| `make browser-e2e-stateful`, `20260921T160254Z-p20078`, 40/42; live query continuation group and aggregate | Fixture recorded `route.fetch` completion for cancelled startup reads as though the browser had accepted them, then compared continuation against that discarded token. It now records completed delivery and waits for the enabled continuation control. Query owner also declines dispatch when required accepted authority is absent; controlled port regression asserts zero network calls. | Affected ordinary scenario and action sequencing pass 11/11 (`20260921T161323Z-p20144`); final candidate recheck recorded below. |
| Initial server-backed run `20260921T161054Z-p66394`, FAIL 131/137 (five browser groups and aggregate); Assessment recovery and preferences | Infrastructure recreated the query port on every authorization-recheck revision even when read authority was identical. This aborted acknowledged refresh reads and restarted unrelated queries. Port identity now follows account/session/incident scope and revocation epoch, with actual scope changes still fenced. Same-authority rechecks do not reset browsing. | Full Assessment service-backed owner slice passes 23/23 (`20260921T161504Z-p60376`); preferences/live query and broad confirmation recorded below. |
| Interim typecheck `20260921T161524Z-p88867` | Browser fixture used `Array.findLast` beyond the current TypeScript library target. Replaced with supported filtering and `at(-1)`; no library/target change. | Final static confirmation recorded below. |

The final partial-receipt boundary regression first failed at `20260921T162229Z-p68940`: the committed-version ledger fabricated a newer raw row from old fields when accepting only a version. Removed that shortcut. `acceptVersion` now records only the highest known version; `latest` requires actual accepted row facts, while the prior observation remains available as prior evidence. Partial receipts promote fields only from an admitted matching base and scope; missing/old-scope provenance retains version/receipt information without promotion. The controlled regression checks both missing and replaced-session provenance, and the unchanged valid grouping/draft and action-sequencing cases remain covered. `make test-slice OWNER=module.timeline ROWS=module.timeline.frontend.timeline_row_mutation_coordinator_1a7e2c9b44,module.timeline.frontend.timeline_query_port_validation_6d44f87e20` passed 3/3 (`20260921T162354Z-p87699`); the ordinary action-sequencing row passed 2/2 (`20260921T162401Z-p88423`). Its new selector is authored in `tools/test_families/module.timeline.json`, then generated normally. Interim static failures were a nullable fixture spread and the now-unused hook dependency; both were corrected rather than suppressed. This is an additional source-admission correction within WS-18, not a new workstream or reopened historical completion claim.

The broad server-backed bundle also exposed Find and range navigation waiting after a terminal validation rejection (`timeline-find` and `timeline-range-selection` groups in `20260921T161054Z-p66394`). The WS-16 pending-command rule had incorrectly treated every queue halt as recoverable transaction identity. Controlled negative regression `20260921T163856Z-p67371` reproduced the missing settlement. The driver now defers settlement only for `client_txn_conflict`; terminal validation settles immediately and later Discard does not settle twice. Both transaction Retry/Discard and terminal validation pass in the ordinary row (`20260921T163949Z-p74011`, 2/2). Current Find/queue confirmation passes 13/13 (`20260921T164004Z-p75158`); range confirmation passes 11/11 (`20260921T164115Z-p17454`). The source guide and authored test catalog were updated; no retry allowance or assertion weakening was introduced.

Additional changed paths belong to existing source owners: Timeline capture adapter/protocol/port and committed/visible row projection, shared query adapter and infrastructure, controlled transport/query and shell tests, browser query fixture, and the concurrent History test. Source guides and final symbol/changed-path ledgers include the final consumer set. No new compatibility path, dependency, arbitrary sleep, retry allowance or weakened assertion was introduced.

### WS-18 current validation and rollout record

Candidate: `main` based on `8af6bba8737e52ad545b4425879c92cf45cfd1bb`, with the authored working-tree changes in [the changed-path inventory](workbook-inspector-changed-paths.tsv) and [the symbol/retirement ledger](workbook-inspector-symbol-ledger.tsv). No commit, deployment or application-wide certification is implied. The pre-existing staged tracker remains staged and untouched in the index; execution updates are additional working-tree changes.

Fresh current-candidate verification (run roots are under `.cartulary/test-results/`; the graph manifest, unit results, row results and referenced browser reports are the authoritative execution artifacts):

| Command / evidence | Result | Run root / artifact |
| --- | --- | --- |
| `make agent-finalize` before broader checks | PASS; retained-run maintenance SKIPPED because `RESULTS_DIR` was unset | `20260921T164050Z-p9638`; earlier finalizers `20260921T160222Z-p16005`, `20260921T161541Z-p91084` also passed |
| `make test-fast` after final validation-settlement correction | PASS 710/710 | `20260921T164111Z-p17044`; earlier final receipt/version run `20260921T162603Z-p10838` and post-authority-fix run `20260921T161604Z-p25166` also passed 710/710 |
| `make frontend-typecheck frontend-import-boundary-check lint-biome` | PASS | `20260921T164004Z-p75268/p75283/p75305` |
| `make browser-e2e-stateful` after the final recovery correction | PASS 42/42 | `20260921T170108Z-p75997`; previous 42/42 `20260921T162837Z-p50789` also passed; initial failing aggregate remains in the failure ledger |
| `make service-backed-test-slice OWNER=module.revisions` | PASS 24/24 | `20260921T162941Z-p2658`; includes HTTP paging, bounded projection, imports/source families, selectors, actions and History browser recovery |
| `make service-backed-test-slice OWNER=module.assessments` | PASS 23/23 | `20260921T161504Z-p60376`; fresh fix confirmation for acknowledged refresh recovery |
| `make service-backed-test-slice OWNER=module.workbook ROWS=module.workbook.browser.preferences_1,module.workbook.browser_stateful.query_continuation_live_recovery` | PASS 13/13 | `20260921T161541Z-p91121` |
| `make service-backed-test-slice OWNER=module.workbook ROWS=module.workbook.browser.note_associations` | PASS 11/11 | `20260921T163147Z-p5613`; fixes the same-authority port replacement that interrupted Notes refresh in the earlier browser bundle |
| `make service-backed-test-slice OWNER=module.timeline ROWS=module.timeline.browser.a_reviewer_session_browser_flow_visibly_performs_c3aaf62d89,module.timeline.browser.capture_action_exact_recovery,module.timeline.browser_support.reviewed_edit_demotion_preserves_query_filter_safety_0e2d9a8f10` | PASS 13/13 | `20260921T162956Z-p11956`; final partial-receipt and version-only semantics |
| `make browser-e2e-measurement` after the final recovery correction | PASS 25/25 | `20260921T170108Z-p76000`; production geometry, density, virtualized selection and measurement; earlier 25/25 `20260921T162104Z-p21747` also passed |
| `make build-server build-web` | PASS 4/4 and 2/2 | `20260921T164534Z-p14763/p14771`; earlier paired builds `20260921T162001Z-p95646/p95659` also passed |
| `make service-backed-test-slice OWNER=module.workbook ROWS=module.workbook.browser.find_edit_departure,module.workbook.browser_stateful.verify_timeline_public_mutations_remount_safe_tr_84f645d6c5` | PASS 13/13 | `20260921T164004Z-p75158`; validation correction and transaction retry both complete |
| `make service-backed-test-slice OWNER=module.timeline ROWS=module.timeline.browser.range_rejection_query` | PASS 11/11 | `20260921T164115Z-p17454`; terminal validation permits deliberate range/query navigation |
| `make generate-drift generated-artifact-policy-check json-shape-check` | PASS 4/4, 3/3, 3/3 | `20260921T164537Z-p15161/p15168/p15170` |
| Fresh History recovery image review | PASS; seven current captures inspected, no clipping or ambiguous action grouping observed | `20260921T162941Z-p2658/manual-review/index.json`; extracted original report image attachments, with unchanged pixels |
| `make browser-e2e-a11y` | PASS 20/20; 66 browser scenarios, no skipped or flaky outcomes | `20260921T164648Z-p62759`; nine original narrow/zoom/creation/recovery captures reviewed at `manual-review/index.json` |
| `make lint-markdown` | PASS documentation maintenance | Final `20260921T171344Z-p87171`; earlier `20260921T164841Z-p11463` also passed; controlling tracker included |
| `make browser-e2e-visual` | PASS 12/12; 47 scenarios, no skipped/flaky cases; no golden changes | `20260921T165056Z-p16545`; `browser-e2e-visual/frontend-visual-reconciliation.json`: 255 captures / 255 active goldens, zero orphan/missing/ambiguous mappings and 29 resolved registered fixtures |
| `make browser-e2e-webserver-backed` on the final implementation | PASS 137/137 | `20260921T164606Z-p31517`; initial aggregate remains FAIL 131/137 at `20260921T161054Z-p66394`, with all five browser causes corrected and confirmed by this fresh full run |

The fresh bounded-read artifact is `20260921T162941Z-p2658/module.revisions/history-paging-measurements.json`: at 11/101/1001 retained events, a one-item page fetches two ordering descriptors, one mutation fact, one revision fact, and invokes two source projectors. Measured allocation is 44,000 / 25,752 / 25,800 bytes; recorded query execution is 0.174 / 0.154 / 0.900 ms. These are fixture observations, not latency guarantees. PostgreSQL still examines ordering/association metadata as history grows; unrelated snapshots and semantic projections remain excluded. The separately exercised selected change set loads two reversal-dependency rows and correctly reports failed preconditions. Existing indexes suffice for the recorded plans; no index was added.

Rollout and rollback proof:

- Deliver the paired browser and server artifacts together. The History wire envelope, semantic generation `cartulary.history.1`, item/entry references, current write routes, source snapshots and request identities are unchanged. This iteration changes private browsing/projection interfaces and the disposable History ordering position.
- Forward: the new HTTP decoder rejects the former one-field `after_history_item_ref` anchor through the existing invalid-pagination path; the browser restarts its read chain. Reverse: inspection of the baseline decoder confirms it accepts exactly that one-field position and rejects the new six-field position. Both use the unchanged protected `pagination.cursor.v1` envelope; no legacy decoder is retained. The production restart/recovery rows and captured-transport/receipt tests verify that read restart does not clear drafts, captured bytes or acknowledgements and never resends an acknowledged write.
- Restore a coherent baseline browser/server pair for application rollback. Existing browser operations may finish against unchanged write contracts; no forced reload, draft serialization, history rewrite, backfill, database reset or storage migration is required. No additive index exists to remove or retain. No rollback or deployment was performed against a live installation.
- Keep source-provider catalog, projection selectors and recovery lifetimes intact when extending the subsystem. Add a source through its existing provider and declared presentation/authoring contracts. Do not reintroduce full-list materialization, presentation-created authority, version-created saved fields, compatibility forwards or duplicated History state.

Visual review: ordinary reconciliation passed without changing goldens. Reviewed the matched `workbook-inspector-details`, `workbook-inspector-history`, `workbook-inspector-rollback-preview`, `workbook-inspector-compact-actions`, `workbook-inspector-narrow-technical-details`, `workbook-inspector-retained-draft` and `entity-mention-chip-states` images. These preserve reading order, current section/record context, semantic History disclosure, reversal scope, explicit retained drafts and non-color mention state. Nine original accessibility captures at `20260921T164648Z-p62759/manual-review/index.json` additionally cover desktop, 390px, 200% zoom, reference selection, contextual task creation, Coordination and Evidence recovery. No unexpected clipping or inaccessible control was observed; scrolled content remains owned by the inspector. Golden-update and two-post-update-run requirements are inapplicable because no golden changed.

### WS-18 final-candidate acceptance assessment

This assessment uses the current source, authored catalogs and fresh execution receipts above. Digest rows supply review prompts; adopted owners remain authoritative. PASS records scripted behavior and source review, not participant usability measurements or an application-wide conformance claim. All mandatory current gates pass with the fresh receipts below; prior failed attempts remain explicitly failed in the causal ledger.

| Digest row | Current disposition | Evidence and scope |
| --- | --- | --- |
| A001 Authority | PASS | WS-11 exact Core 01/02/03/04 and Design §12.7 owner map; Core amendments precede code; implementation/query/codec details remain downstream. |
| A002 Scope | PASS | G12–G18 remediation rubric and WS-12–17 migration/retirement exits; one browsing authority, closed presentation, explicit source evidence and bounded semantic paging. |
| A003 Repository state | PASS | `main` / `8af6bba8737e52ad545b4425879c92cf45cfd1bb` plus 173 changed paths; current manifests, local guides, production imports and original staged tracker inspected. No direct grid boundary bypass or localization registry change. |
| A004 Tokens | PASS | Diff review: no new CSS/design literal or parallel theme/density registry; saved reading and creation reuse existing shell and density owners. |
| A005 Theme | PASS | Existing `dark_graphite` fixture and owner remain unchanged; no additional palette or theme exposure. |
| A006 Density | PASS | Measurement 25/25 and visual 12/12 with current density/geometry fixtures; saved/draft/read-only content uses the existing shared selection. |
| A007 Creation | PASS | All seventeen schemas and closed creation/saved/empty shell variants; contextual source retention, closed/read-only and uncertain creation in stateful 42/42, fast 710/710 and source browser matrix. No unsupported inline creation added. |
| A008 Responsive | PASS | Measurement 25/25, accessibility 20/20 and visual 12/12 cover token/accessor geometry, below-minimum/narrow/overlay layouts, resize and viewport fallback; reviewed 390px and 200% zoom captures. |
| A009 Overflow | PASS | Accessibility 20/20 and visual 12/12; nine reviewed source captures preserve shell-owned scrolling, persistent context, reachable editing/recovery and status/navigation regions. |
| A010 Inspector | PASS | All seventeen schemas, one admitted section sequence, semantic dispatcher, subject/role/source changes, retained drafts and late outcomes; fast 710/710 and stateful 42/42. |
| A011 Continuity | PASS | Controlled provenance, page/generation/late response, detachment and exact-replay tests; stateful 42/42, Revisions 24/24, Find/recovery 13/13 and range 11/11. |
| A012 Transactions | PASS | Exact captured request/identity and one-settlement regressions; final fast, Timeline exact recovery and ordinary browser queue recovery. No request reconstruction or weak random ID introduced. |
| A013 Acknowledgement/recovery | PASS | WS-16 controlled retry/focus regressions plus WS-18 terminal validation settlement; Assessment 23/23, Notes 11/11, History 24/24, Find/queue 13/13. Accepted writes recover through reads without resend. |
| A014 Editing | PASS | Saved-value/edit attachment separation, raw draft lifetime, validation correction, keyboard Find/range navigation, authority changes and current inspector edit browser scenarios. |
| A015 Conflict | PASS | Existing primary cell conflict and saved/draft separation retained; controlled rejection and ordinary recovery checks. No toast-only recovery introduced. |
| A016 Query/interaction | PASS | Source observation admission and stable query-port authority key; fast query/producer matrices, stateful continuation and recovery, final source/browser checks. Version knowledge cannot fabricate a saved row. |
| A017 Authorization scope | PASS | Dispatch-time provenance and record/version/scope filtering; old-session clones, session recovery, account replacement, incident revocation, role changes and off-window retention in controlled/ordinary source tests. Source rechecks preserve port identity only when actual authority is unchanged. |
| A018 Evidence | PASS | Lifecycle/preview states retain existing semantic/non-color presentation; Evidence query stateful and inspector/History source cases. |
| A019 Accessibility | PASS | Accessibility 20/20 (`20260921T164648Z-p62759`), including 66 scenarios with zero flaky/skipped cases; controlled Coordination recovery focus and ordinary confirmation also pass. No additional upstream conformance claim. |
| A020 Components | PASS | All-schema presentation tests, measurement 25/25, accessibility 20/20 and visual 12/12 cover densities, long content, spacing, zoom, reduced motion and component state precedence. |
| A021 Virtualization | PASS | Production measurement 25/25, stateful 42/42 and Find/range browser checks; accepted-row versions never derive facts from version-only knowledge. |
| A022 Visual fixtures | PASS | Visual 12/12; 255 captures reconcile to 255 active goldens with zero missing/orphan/ambiguous entries; all 29 registered fixtures resolved. Seven matching inspector/mention goldens and nine fresh accessibility images reviewed. No golden change or maintenance update was required. |
| A023 Selectors | PASS | Stable item/entry references and semantic selector allocation retained; final fast includes package UI owner rows, plus bounded History/source browser cases. |
| A024 Test authority | PASS | Changed tests/generators/runtime consume authored machine inputs and fixtures outside Markdown; Core amendments and handoff remain human review. Markdown lint is documentation maintenance only. |
| A025 Generated artifacts | PASS | Authored catalogs/policy updated before `make generate`; final generation drift 4/4, artifact policy 3/3 and JSON shape 3/3. No generated-root manual edit or dependency lock change. |
| A026 Compatibility | PASS | Consumer migrations and 28 moved/retired declarations reconciled; paired builds and forward/reverse disposable cursor restart proof. No legacy decoder, schema/storage migration, history rewrite or retained compatibility facade. |
| A027 Handoff | PASS | All serial checkpoints and mandatory gates complete; final receipts, failure causes, compatibility/rollback, 173 changed paths, 163 retained symbols, 28 retirements, limitations and next action recorded. |

All rows apply to the affected subsystem; no mandatory criterion is marked N/A. Optional features absent from the adopted owner scope are not newly required. Upstream publication certification, participant studies and restoring archived handoffs are outside this effort.

### Representative task evidence

The five tasks were exercised through production browser routes and semantic controls. These results establish scripted outcomes; no participants, task timing, error rate or scroll-distance comparison was collected.

| Task | Scripted outcome | Fresh evidence |
| --- | --- | --- |
| Locate Evidence | Open selected record, navigate its admitted Evidence section and inspect associated content while preserving record context. | `module.revisions.browser_stateful.verify_inspector_details_relationships_evidence_e32cd188c5`, stateful `20260921T170108Z-p75997`; final visual/a11y review supplements it. |
| Edit, close and resume | Dirty fields remain owner-retained on close/selection change; deliberate return resumes correct authoring; acknowledgement/failed refresh does not send another patch or steal newer focus. | `workbook-inspector-edit.spec.ts` production scenarios; final fast draft tests, Revisions recovery suite and final server-backed receipt. |
| Correct a mention | Manual resolution, dismissal, auto-resolution disclosure and Undo operate on current source/mention identity and refreshed rows. | `module.entities.browser_stateful.verify_manual_mention_resolution_dismissal_auto_dfa355e592`, stateful `20260921T170108Z-p75997`; captured operation and reconciliation tests in final fast. |
| Inspect History and reversal scope | Browse complete logical events, follow continuation, inspect current eligibility and invoke admitted server-supplied actions; recover uncertain and acknowledged outcomes distinctly. | Full `module.revisions` service slice `20260921T162941Z-p2658` 24/24, canonical paging/source matrix and seven reviewed recovery captures. |
| Create a related task | Open Timeline Workflow, create the related task through owner-declared context and retain the workbook shell; preserve source/authoring/recovery lifetimes. | `module.workbook.browser_stateful.verify_timeline_inspector_workflow_create_relate_7fc4833af4`, stateful `20260921T170108Z-p75997`; Coordination recovery browser and controlled tests. |

### Final limitations and maintainer handoff

- No deployment, rollback against a live installation, history rewrite, snapshot backfill, database reset, schema migration, index or forced reload was performed. Browser/server rollback is a reviewed paired-artifact procedure supported by cursor and recovery tests.
- Retained-run maintenance was skipped because `RESULTS_DIR` was unset; no successful full warm check was supplied. Agent finalization still passed before broader verification.
- The unrelated existing Go staticcheck finding `ST1005` at `internal/app/workbookassembly/note_associations.go:62` remains (`20260921T153440Z-p37502`). The failing target is recorded in the WS-15 exit; no full Go staticcheck success or application-wide release certification is claimed. Touched-owner tests, Go vet/format, frontend checks and generated-contract checks have their own receipts.
- SQL ordering/association metadata work can still grow with retained events. The bounded materializer excludes unrelated snapshots/projectors; selected change-set reversal dependencies remain an explicit additional cost. Recorded fixture allocations/timings are observations, not production guarantees.
- Participant usability benefit remains a hypothesis. Automated keyboard/accessibility and visual evidence do not measure human task time, comprehension or error rates.
- The next maintainer action is to review the final diff and paired delivery artifacts through the normal release process. Extend source-owned providers/presentation contracts and authored test catalogs; preserve explicit observation admission, captured operations and semantic History identities. No later remediation workstream or compatibility cleanup is deferred.

### WS-18 DONE — mandatory final exit

Completed 2026-09-21 17:13 UTC. The WS-18 entry followed WS-17's saved passing exit. All eight workstreams now have saved entry and completion checkpoints. I3-01–I3-15 and all applicable A001–A027 are PASS; no mandatory acceptance or superseded compatibility path remains open.

| Gap | Final disposition and owner |
| --- | --- |
| G12 | DONE — shared History browsing authority and operation/page reconciliation; web.workbook and Timeline consumers |
| G13 | DONE — intentional shared exports, fixture migration and 28 moved/retired declarations; web.workbook/web.testing and authored ownership/import policy |
| G14 | DONE — closed shell modes and independent saved-value reading across all seventeen schemas; presentation and source authoring owners |
| G15 | DONE — source-accepted record/version/authority observations, protected lifetime fencing and version-only evidence separation; query/mutation source owners |
| G16 | DONE — typed, bounded logical History pages, selected dependency accounting and protected disposable position; module.revisions/HTTP |
| G17 | DONE — causally reproduced Timeline/Coordination failures plus integrated provenance/query/validation corrections; responsible product or fixture owners and ordinary suites |
| G18 | DONE — owner clarification, inventory, fresh validation, rollout/rollback proof and this final handoff |

Final complete server-backed run `20260921T164606Z-p31517` passes 137/137; final stateful `20260921T170108Z-p75997` passes 42/42; final measurement `20260921T170108Z-p76000` passes 25/25. Current fast passes 710/710, accessibility 20/20 and visual 12/12 with no golden changes. Paired builds, frontend static/import/lint checks, generation drift, artifact policy, shape and documentation maintenance have separate receipts above. Final whitespace/index checks preserve the original staged tracker. Earlier failures are not relabeled passes; their supported fixes and final confirmations are retained.

Scope is complete at `main` / `8af6bba8737e52ad545b4425879c92cf45cfd1bb` plus the 173-path working-tree inventory. No commit or deployment was requested. Retained-run maintenance is explicitly skipped with `RESULTS_DIR` unset; the unrelated existing ST1005 and unmeasured participant outcomes remain limitations, not hidden product passes. The next maintainer action is normal review and coordinated delivery; no additional remediation slice is deferred.

### Planning handoff retained for historical evidence

The following planning-only results predate WS-11. The serial execution exits above and final WS-18 handoff supersede their state and next-action descriptions.

#### Planning checkpoint (historical) — backend handoff

| State | Files inspected | Result / risk | Next action |
| --- | --- | --- | --- |
| Planning only | Revisions application ports/service, repository, materializer, HTTP route/pagination, attribution/actions, source catalog and existing indexes | Full retained snapshot materialization precedes cursor selection; G16 proposes typed bounded pages without changing source truth | WS-15 must preserve coalescing/identities, measure query behavior and justify any additive index |

#### Planning checkpoint (historical) — frontend handoff

| State | Files inspected | Result / risk | Next action |
| --- | --- | --- | --- |
| Planning only | All inspector/History declarations and tests; exact model/controller, Shell/Details, retention, Timeline/Generic/Entity/Assessment and adapter seams | G12–G15 distinguish live compatibility paths, unused surface, contradictory props and a provenance weakness; no exploit or new runtime failure claimed | WS-12–14 migrate owners and consumers together; WS-17 closes symbol ledger |

#### Planning checkpoint (historical) — contract and generation handoff

| State | Files inspected | Result / constraints | Next action |
| --- | --- | --- | --- |
| Unchanged | Core 01 REQ-01-052A/054/§3.3.7, Core 02 REQ-02-216/218/265, Core 03 lifetimes, OpenAPI/design/view projections and generated policy | Current semantic wire contract, canonical source facts and protected generated roots retained; no dual protocol or automatic reset | WS-11 adopts only necessary owner clarifications; generation/compatibility checks only with later authored changes |

#### Planning checkpoint (historical) — tests and documentation handoff

| Check / artifact | Result | Interpretation / follow-up |
| --- | --- | --- |
| `make task-guide ROLE=module-author OWNER=web.workbook` | PASS | Current routing only; no tests executed |
| `make task-guide ROLE=module-author OWNER=module.workbook` | PASS | Current browser/a11y/visual routes identified |
| `make task-guide ROLE=module-author OWNER=module.revisions` | PASS | Current Revisions unit/service/browser routes identified |
| Historical run directories `20260921T073129Z-p96952`, `20260921T073453Z-p32834`, `20260921T074157Z-p46106`, `20260921T074150Z-p32164` under `.cartulary/test-results/` | Present at inspection | No rerun or fresh PASS; WS-16 uses available failure artifacts and fresh reproduction |
| `make lint-markdown` | PASS | `.cartulary/test-results/20260921T133544Z-p55519/adhoc/lint-markdown/tool-run-summary.json`; configured documentation scope, zero failures |
| Explicit tracker lint via `bash ./tools/harness/static-analysis/markdownlint.sh --no-globs docs/handoffs/workbook-inspector-ui-ux-audit.md` | PASS, exit 0 | Same Make-owned lint wrapper, explicit file selection; default configured globs omit this handoff, so the public target alone would not establish its lint result |
| `git diff --check`; `git diff --cached --check`; `git status --short`; file/checkpoint reconciliation | PASS | Only this tracker modified; 74/74 target files, all 12 numbered sections and all 62 historical WS table/checkpoint rows preserved; no staged changes |
| Product suites, builds, generation, migrations, golden changes and deployment | SKIPPED | Document-only scope; no production behavior changed |
| `make agent-finalize` and retained-run maintenance | SKIPPED | No broader implementation verification; finalizer may refresh generated structure; `RESULTS_DIR` unset |

All Make commands in this environment use the repository-local Node runtime via
`PATH="$PWD/tmp/node-runtime/bin:$PATH"`. Read-only `rg`, source/declaration and
relative-import inventory scripts establish planning evidence, not product tests.
The archived handoff remains outside scope per the user. Documentation review
corrected the navigation module extension from `.tsx` to `.ts`; every concrete
connected path in the new inventory now resolves. No failed product check is
claimed or waived; product verification was intentionally not executed.

#### Planning checkpoint (historical) — security handoff

| Finding | Evidence class | Required owner action |
| --- | --- | --- |
| G15 retired-object admission | Structural weakness; no demonstrated disclosure | Explicit accepted provenance; do not relabel old data with current authority |
| G12/G16 read and paging changes | Compatibility/recovery risk | Preserve captured requests, receipts, actor attribution and canonical facts; no read cursor controls write identity |

#### Planning checkpoint (historical) — next session

| Current state | Next action | Completion boundary |
| --- | --- | --- |
| Document plan and P-05 DONE | No further document work pending; production implementation has not begun | Exactly this tracker changed; no implementation workstream started |
| WS-11–WS-18 TODO | Later authorized implementation starts WS-11 IN_PROGRESS and saves its checkpoint | Serial slices; mandatory final WS-18; no archived handoff dependency |

### Previous-iteration section 10 record (historical)

### Implementation execution checkpoints

| Slice | Status / evidence | Changed boundary / next action |
| --- | --- | --- |
| WS-10 exit | DONE; 2026-09-21 07:54 UTC. Final fast 708/708 `20260921T073129Z-p97071`; final public visual update 12/12 `20260921T074355Z-p30667`, all changed images reviewed, then two fresh ordinary visual runs 12/12 each `20260921T074928Z-p69256/-p69271`. Final Markdown `20260921T075527Z-p42495`, whitespace, static/contract/generation checks and paired builds pass. Exact full/focused browser results and resolved failures are retained in the validation handoff. | G11 and overall effort complete; A001–A027 PASS with explicit optional-feature N/A rationale. Fixed canonical null/empty preservation, saved-text copy and access/presentation scope separation during integrated validation. All 358 changed paths inventoried, 151 images reviewed, 255 captures/mappings reconciled with no orphan/missing/ambiguous entries. Removed superseded presentations; retained exports have named consumers. Verified baseline/candidate 73-file bundles, coordinated rollout/rollback, skipped retained-run maintenance (RESULTS_DIR unset), and unavailable participant evidence recorded. No deployment, reset, backfill, data rewrite, legacy adapter or forced reload. This is the mandatory final slice; no later workstream remains. |
| WS-10 entry | IN_PROGRESS; WS-09 DONE checkpoint saved before entry | Complete integrated owner checks, browser/accessibility/visual reconciliation and review, two ordinary visual confirmations, acceptance A001–A027, retirement/cutover/rollback and final handoff. Run agent-finalize before broader final verification; retained-run maintenance is skipped unless a qualifying full warm check RESULTS_DIR is supplied. |
| WS-09 exit | DONE; 23 authoring/recovery selections passed 24/24 units (`20260921T055045Z-p28162`), final Assessment/subject/Task/Generic checks passed 5/5 (`20260921T055218Z-p69021`). Eight specialized accessibility rows passed 11/11 (`20260921T055157Z-p36472`). Assessment append-recovery/follow-on browser rows passed 13/13 (`20260921T055302Z-p75081`); ordinary creation matrix/continuity passed 11/11 (`20260921T055429Z-p9250`). Final format/type/lint/import/drift passed `20260921T055428Z-p9113/-p9093/-p9103/-p9097/-p9035`; Markdown `20260921T055509Z-p44468`; whitespace passed. Exact selectors are recorded in the retained run manifests and family review below. | G08 complete. Changed shared form roles, Shell/Details, Assessment/Generic compositions, Note/coordination/Task/Decision/Evidence/Indicator forms, related-row renderer, source guides/ownership and affected tests. Saved headers win over operation titles; creation modes and target labels are explicit; Assessment saved reading and append controls share the adopted presentation. Removed two Indicator style modules and migrated thirteen consumers; removed local form style alternatives. All seventeen configurations accounted for below. Existing source/draft/receipt owners and supported generic row-command adapter retained; no storage/protocol migration. Initial lint (`20260921T055116Z-p31387`) found the unused CSSProperties import after local-style retirement; removed and reran successfully. No check waived. Save before mandatory WS-10 entry. |
| WS-09 entry | IN_PROGRESS; WS-08 DONE checkpoint saved before this entry | Migrate specialized form presentation and creation modes while retaining each source owner, submission, detachment and operation recovery. Inventory all seventeen configurations and remove superseded local treatments. |
| WS-08 exit | DONE; `make test-slice OWNER=web.workbook ROWS=web.workbook.regression.inspector_ordered_actions,web.workbook.regression.inspector_grouped_permission_reasons,web.workbook.regression.contextual_task_decision_discovery,web.workbook.regression.contextual_task_decision_authoring,web.workbook.regression.timeline_inspector_feature_controller_a38c2d6f71` passed 6/6 (`20260921T054436Z-p79987`). Stateful default/closed/denied/surface-switch browser row `module.workbook.browser_stateful.verify_default_closed_inspector_state_no_row_sta_897d604d9e` passed 11/11 (`20260921T054457Z-p81569`). Format/type/lint/import passed `20260921T054426Z-p75155/-p75135/-p75145/-p75139`; generation `20260921T053918Z-p63699`, drift/shape `20260921T054510Z-p3670/-p3682`, Markdown `20260921T054555Z-p20037`, whitespace passed. | G03 Workflow presentation complete for bound actions. Changed shared ContextualActions, Actions/model, presentation contract/schema/generator/facade, presentation tests/catalog, README and inventory below. Canonical order and single dispatch replace wrapping/duplicate inputs; authored outcomes share accessible descriptions with deduplicated reasons. Existing command owners and unsupported-feature omission remain explicit. No stored-data/protocol migration; no wording alias. Initial type/reason checks failed on obsolete tests expecting no description on enabled actions (`20260921T053918Z-p63759`, `20260921T054321Z-p72851`, `20260921T054348Z-p73844`); fixed to validate outcome-plus-reason semantics and reran. One invalid combined target invocation applied OWNER to typecheck; corrected target routing. No acceptance waived. Save before WS-09 entry. |
| WS-08 entry | IN_PROGRESS; WS-07 DONE checkpoint saved before this entry | Replace equal-weight Workflow wrapping with ordered action rows using feature identity, authored labels and explicit outcome metadata. Preserve capability admission, dispatch and reason deduplication. |
| WS-07 exit | DONE; final pending-read/retained-recovery/batch-review units passed 5/5 (`20260921T053140Z-p16493`); public stale/error/reversal and stateful authorization browser rerun passed 13/13 (`20260921T053139Z-p16273`). Four-family paging and legal scopes plus accessibility evidence below passed. Generation `20260921T053117Z-p8731`, format `-p8811`, typecheck `-p8791`, Markdown `20260921T053223Z-p48494`, lint `-p48490`, import boundary `-p48484`, drift and whitespace passed. | G04 and History portion of G03 complete. Each event carries all admitted semantic detail and exact timestamps, attribution is visible, reversal review is event-local and explicitly scoped, record actions trail reading. Closing an event cancels only its unsubmitted review, including late reads; submitted requests/receipts retain their owner. Changed selectors, visual setups and README follow the same public disclosure flow. No protocol/data migration beyond WS-06; integrated visual review remains WS-10. Save before WS-08 entry. |
| WS-07 validation | Semantic unit/actor/value tests and existing reading/recovery selections passed except old hidden-control focus assumptions (`20260921T052425Z-p11841`, 7/8 units). After opening event disclosure in those fixtures, recovery passed 2/2 (`20260921T052601Z-p22104`); final presentation/model/recovery selection 4/4 (`20260921T052806Z-p34533`). All four family overflow/paging/retry/refresh/reversal browser scenarios and four legal-action scenarios passed in `20260921T052358Z-p81064`. Both inspector accessibility scenarios passed 11/11 (`20260921T052922Z-p73034`). | Changed History presentation model/types, event renderer, loaded action/confirmation composition, semantic focus handling, source guidance and public browser fixtures. Visible source attribution, UTC numeric-offset display, exact timestamp and complete units live under Event details; scoped reversal explanations and local confirmations; trailing Record actions. No display label grants an action. Final pending-read cancellation and public failure/authorization reruns follow before exit. |
| WS-07 resolved checks | Generation first rejected a new test title routed to the wrong file (`20260921T052219Z-p66220`); corrected and gave semantic detail its own authored row. Typecheck identified the generated nonempty-unit tuple requirement; fixture now constructs an explicit first unit. Two fixture focus failures clicked concealed actions, and browser inspector-actions fixtures omitted Event details or assumed a different owner error sentence; updated the user flow and asserted the current review requirement while keeping no-write/public-error checks. | `20260921T052358Z-p81064` and `20260921T052804Z-p34284` retain failing browser traces. Inspected disclosure cancellation and fixed pending-read review cancellation/clearing without aborting submitted owner operations; a focused late-response regression now covers this transition. Removed technical-only actor, raw-operation secondary paragraph, global reversal confirmation and leading destructive block. Shared event rendering remains used by record History and batch/merge review. |
| WS-07 entry | IN_PROGRESS; WS-06 DONE checkpoint saved before this entry | Render complete semantic History with visible attribution/time, event-local reversal and trailing Record actions. Preserve action metadata, confirmation safety, selection, paging and operation recovery. |
| WS-06 exit | DONE; final generation `20260921T051327Z-p15586`, typecheck `20260921T051432Z-p50356`, lint `20260921T051522Z-p58283`, imports `20260921T051108Z-p91432`, drift `-p91358`, generated policy `-p91360`, JSON shape `-p91362`, OpenAPI compatibility `-p91368`, Markdown `20260921T051400Z-p19103` and whitespace passed. Behavioral results and resolved failures are below. | G10 source/transport exit complete. Retired arbitrary units and coarse materializers, with no legacy response adapter. Source README and inventory below document canonical-only imports, coordinated 2.0.0 cutover/rollback, and disposable generation restarts preserving operation recovery. No stored-data change. Save this checkpoint before WS-07 entry. |
| WS-06 validation | Source/catalog/materialization/portability unit selection passed 6/6 (`20260921T051155Z-p46164`); frontend History browsing, lookup, transport, retained operations and batch review passed 8/8 (`20260921T050349Z-p3191`). History envelope and retained restart passed in `20260921T050012Z-p58487`; Indicator reversal rerun passed 3/3 (`20260921T050311Z-p82466`); link/tag integration 3/3 (`20260921T050458Z-p35453`); imported attribution/retained reversal 3/3 (`20260921T051259Z-p66812`). Browser entry reversal passed in `20260921T051106Z-p90931`; full keyboard browsing/recovery/reversal rerun passed 11/11 (`20260921T051359Z-p18836`). | Pure source projectors in nine modules, required catalog fields/compiler, Revisions historycontract/materializer/service/HTTP, authored OpenAPI, generated protocol, browser validation/generation boundaries, canonical fixtures and source README changed. All supported source families have complete deterministic units. Missing/malformed facts fail locally; values preserve absent/null/false/zero/empty. No raw snapshot or storage mutation ID is exposed. Imported actor fallback remains declared; no directory lookup. |
| WS-06 resolved failures | Early catalog fixtures lacked required schema/projector fields; corrected fixtures. Generation initially stopped on unreviewed compatibility fingerprints; reviewed eleven findings in pending 2.0.0 and regenerated. Old tests searched retired target IDs or expected assertive read alerts; migrated to semantic references and the existing polite notice owner. The paging scenario exposed absent end-of-history feedback; added it. Imported fixture had partial canonical Host fields and then an incorrect origin default; corrected both to source facts. | Initial failing roots include `20260921T050425Z-p4666`, `20260921T051106Z-p90931`, `20260921T050542Z-p63822` and `20260921T051107Z-p91156`; successful focused reruns above close them. A parallel service-image warm-stamp collision (`20260921T051300Z-p67161`, harness failure before browser launch) was resolved by sequential rerun, without changing harness policy. A mistyped unit row was rejected before execution and corrected to the authored catalog ID. Type/lint fixture errors and a missing callback dependency were corrected; final format passed `20260921T051503Z-p53691`. |
| WS-06 entry | IN_PROGRESS; WS-05 exit saved before entry | Implement source-owned semantic projectors, closed public History contract and coordinated representation cutover. Canonical retained snapshots and conformant imports only; no historical-shape inference, automatic reset or legacy response adapter. |
| WS-05 exit | DONE; module.evidence guide refreshed; web.workbook access slice passed 4/4 (`20260921T043242Z-p25928`), final sibling-failure check 2/2 (`20260921T043452Z-p76052`); Timeline Evidence 2/2 (`20260921T043243Z-p26722`); lifecycle/upload retained 6/6 (`20260921T043313Z-p31737`). Browser/accessibility/stateful run `20260921T043314Z-p32150` passed five selected rows but failed original-source review because its drop targeted the retired container. Updated that fixture to the named attachment region; focused rerun passed 11/11 (`20260921T043746Z-p77584`). Typecheck `20260921T043312Z-p31564`, imports `20260921T043402Z-p64662`, lint `20260921T043403Z-p65087`, format `20260921T043441Z-p71409`, Markdown `20260921T043806Z-p7501` and whitespace passed. | Changed EvidenceAttachmentEntry, EvidenceAccessActions, useEvidenceWorkbookBindings, EvidenceFileRecovery, TimelineEvidencePanel, source ownership, focused unit/browser fixtures and source guidance. Three independent regions distinguish saved information, authorized access and attachment recovery. Picker/drop/paste share admission and source-owned commands; acknowledged writes remain retained. Removed the old outer-container drop/paste path and combined access/attachment component. Compact row access remains a named consumer. Count scope stays explicit; no new enumeration, schema or stored-data migration. Failure artifacts and successful browser output remain under the named run roots. Next WS-06. |
| WS-05 entry | IN_PROGRESS; WS-04 exit saved before entry | Separate accepted Evidence information, independently authorized access and source-local attachment recovery. Keep all upload entry paths on the existing owner admission. |
| WS-04 exit | DONE; refreshed module.entities/module.timeline guides; web.workbook mention candidate/receipt/recovery/reconciliation/auto-resolution slices passed 6/6 units (`20260921T042428Z-p73919`); module.timeline collection inspection passed 2/2 (`20260921T042457Z-p5697`); module.entities dismissal/restoration and creation-recovery browser slice passed 13/13 (`20260921T042428Z-p73950`). Typecheck `20260921T042428Z-p74128`, format `20260921T042404Z-p69087`, lint `20260921T042520Z-p8789`, import `20260921T042522Z-p8984`, whitespace passed. | Changed TimelineMentionsPanel, TimelineMentionActionControls, candidate/collection tests, mentions.recovery browser scenario and presentation README. Compact raw-text/state/target summaries; complete provenance under Mention details; selected correction disclosure remains in original group with mounted inputs. Creation and mention receipts/replay/refresh stay visible outside closed controls. Empty partial candidate pages explicitly retain the continuation meaning. Existing source identity, permissions, canonical-link distinction and Dismissed in this session retained. No protocol/data migration. Expanded shared relationship chips remain used by EntityRelationships' matching Timeline preview; they are not an abandoned alternative. Browser results retained under the successful run's browser groups. Next WS-05. |
| WS-04 entry | IN_PROGRESS; WS-03 exit saved before entry | Compact mention summaries and source-local correction disclosure; preserve bounded candidate discovery and visible operation recovery while controls are collapsed. |
| WS-03 exit | DONE; owner guide refreshed; unit slice `20260921T041810Z-p13491` passed 7/7 units: saved reading, Entity editing, Timeline explicit Details, ordinary recovery, draft binding and lifetime. Browser/accessibility slice `20260921T041811Z-p13712` passed 13/13 units, including subject retention/narrative expansion, detached refresh, exact replay and inspector edit accessibility. Final typecheck `20260921T042045Z-p58296`, focused navigation `20260921T042043Z-p57294`, format `20260921T042042Z-p57026`, lint `20260921T041842Z-p48383`, import `20260921T041843Z-p48574`, generation `20260921T041727Z-p5291`, drift `20260921T041812Z-p14062`, Markdown `20260921T041943Z-p54588` and whitespace passed. | Reworked WorkbookInspectorDetails, ordinary draft read projections/hook, TimelineInspectorDetails and section bindings, Generic presentation, Entity composition, shared section focus descriptor, unit/browser fixtures, catalog and generated routing; updated presentation README. All 17 configurations preserve declared field order and value distinctions. Source string/read contracts classify rows; no second field registry. Explicit editor slots group Update/Close and retain owner feedback. Original-field cues select the correct collection action before Resume; captured receipts remain independent. Removed Done editing and Entity's unused control-stack style. Source owners supply collection destinations; semantic descriptors now carry their focus resolver explicitly. No stored-data or draft-lifetime migration. Next WS-04. |
| WS-03 resolved checks | Initial type/lint exposed an unused Entity style and an effect dependency; corrected both. Initial generation/type/format stopped because the new authored test row was not ASCII-sorted (`20260921T041650Z-p1508` generation); sorted the owner catalog and regenerated. | No generated output was edited by hand and no acceptance was waived. Browser artifacts are the successful run's `browser-e2e-webserver-backed/browser-groups/functional-support-default-workbook-inspector-edit/` and `browser-e2e-a11y/browser-groups/a11y-workbook-inspector-edit/` reports. |
| WS-03 entry | IN_PROGRESS; WS-02 exit saved before entry | Implement saved scalar/narrative rows and explicit editor slots; expose original-field retained work through the existing draft owner. Preserve submission, clear/null, revision captures and operation recovery. |
| WS-02 exit | DONE; `make test-slice OWNER=web.workbook ROWS=web.workbook.regression.inspector_persistent_context,web.workbook.regression.inspector_explicit_panel_states` passed (3/3, `20260921T040716Z-p77268`); `make service-backed-test-slice OWNER=module.workbook ROWS=module.workbook.browser.inspector_subject_retention` passed (11/11, `20260921T040648Z-p42149`); typecheck `20260921T040649Z-p42561`, import boundary `20260921T040716Z-p77362`, lint `20260921T040432Z-p3564`, format `20260921T040706Z-p68723`, whitespace passed | Changed declared panel admission, shared shell, new layout navigation hook/context and SurfaceLayout provider; migrated Timeline, Entity, Generic and Assessment callers; marked History Open focus destination; updated unit/browser tests, source ownership and presentation README. One sequence supplies mounted body/navigation. One bounded section selection survives close; no source/request state added. Removed standalone list rendering fallback. Existing semantic openers and operation lifetimes retained. Browser verifies adjacent/overlay widths 1440/760, zero navigation History reads, nested Escape and retained editing. Next WS-03. |
| WS-02 resolved failures | Initial format/lint rejected nonsemantic disclosure markup; unit fixture used unavailable matchers; typecheck rejected a testing-library option. Corrected and reran. Browser runs `20260921T040259Z-p64137` and `20260921T040523Z-p4281` failed grid focus during retarget; trace identified panel remount causing editor autofocus. | Kept section keys stable across subject changes and restricted concealment focus recovery to lost focus within the same subject. Added mount-continuity assertion. Successful ordinary browser rerun above verifies recovery; no test assertion waived. Artifacts are under each run root's `browser-e2e-webserver-backed/browser-groups/functional-support-default-workbook-inspector-edit/`. Invalid diagnostic flag invocation was rejected by harness configuration; ordinary lint provided diagnostics. |
| WS-02 entry | IN_PROGRESS; WS-01 exit recorded first | Derive one admitted descriptor sequence in composition and consume it in shell navigation/body. |
| WS-01 entry | IN_PROGRESS; WS-00 exit recorded before entry | Shared role/caller inventory, authored narrative line-limit projection and focused owner verification. |
| WS-01 exit | DONE; generation `20260921T034856Z-p13199`, format `20260921T034932Z-p16970`, typecheck `20260921T034951Z-p22064`, UI slice `20260921T035119Z-p25506` (10/10 units), Workbook slice `-p25543` (4/4), import `-p25779`, drift `-p25482` passed | Shared complete typography/button/input roles; shell focus/disabled presentation; Timeline/Indicator style declarations now consume common roles. Existing style export names retain current callers without separate values. Authored narrative limit generated; no dependency/theme change. Final browser/density/golden review remains WS-10. |
| WS-01 verification setup | `make generate`, `make format`, `make frontend-typecheck` and `make doctor` passed. Initial slices/import/drift runs failed before child launch with `infra/service_start_error` (runs `20260921T034951Z-p21785`, `-p21832`, `-p22136`, `-p21882`). `make explain-run` confirmed classification. | The shell had no `node` on PATH; Make's pinned runtime exists at `tmp/node-runtime/bin`. Rerun public Make targets with that runtime prepended to PATH; no harness policy change or global installation. |
| WS-00 entry | IN_PROGRESS; baseline `67f3d98ac01b42a1da4409c0a6630d08d24e1761`; audit already staged, no other changes | Adopt approved design/History decisions. Canonical retained snapshots only: Core 02 REQ-02-265 forbids schema-less history decoding or backfill. No database reset is authorized. |
| WS-00 exit | DONE; `make lint-markdown` passed, final owner-text run `.cartulary/test-results/20260921T034447Z-p9117`, summary `adhoc/lint-markdown/tool-run-summary.json`; `git diff --check` passed | Changed design, Core 01/02/03 and this tracker. Adopted closed History values/units, attribution, coordinated cutover and canonical-only admission; added explicit navigation, editing, Evidence, History and Workflow criteria. No product check claimed. Next WS-01. |

### WS-06 source and compatibility inventory

| Source owner | Required canonical contribution | Detail boundary |
| --- | --- | --- |
| Timeline | Timeline record | Public fields, generated-value flags, capture/review/supersession facts |
| Entities | Host and Identity; mention, alias, preserved identifier | Public fields, merge target, mention provenance/state/resolution, semantic identifiers |
| Evidence | Evidence record | Public metadata and lifecycle/upload state; excludes object bytes, object storage identifiers and access handles |
| Assessments | Assessment record | Subject, assessor, assessed time, confidence, state and rationale |
| Artifacts | Artifact record, all eight configured variants | Existing source catalog fields and readonly assigned identifiers/lifecycle facts |
| Tasks/Decisions | Task and Decision records | Existing source catalog public fields |
| Parties | Party record | Public party fields |
| Indicators | Indicator record, observation and interval | Identity values, observation resolution, interval rationale/support/confidence/times |
| Links | Record link and tag | Public endpoints/type/provenance and tag identity; attached Evidence has its own semantic kind |

The application catalog admits ten record snapshot types and fourteen mutation
target kinds from these nine owners. Twenty-four authored fixtures exercise
seventeen record configurations and seven non-row families; row target aliases
reuse their admitted record projector. Missing or duplicate contributions fail
catalog construction. Deterministic units expose semantic references and typed
absent/null/present values, never raw target IDs or snapshots. The browser uses
the generated closed response validator and validates duplicate identities.

The public owner source is
`contracts/openapi-source/owners/module.revisions/openapi.json`. Generated
protocol types/validators and assembled OpenAPI are downstream. The pending
2.0.0 change set records eleven reviewed findings; the released 1.0.0 baseline
is unchanged. `data.representation_generation` also exists on empty pages.
Continuation and action-proof reads compare this marker before immutable item
content and require a fresh first page across generations. Captured mutation
bytes, receipts, drafts, selectors, item references and `/api/v1` remain intact.

Deploy the bundled server/browser together and retain the coherent prior bundle
for rollback. No response adapter, SQL migration, automatic reset or historical
shape repair exists. Supported data is canonical retained history and conformant
imports. Unsupported legacy databases remain at the existing explicit operator
reset boundary. Imported attribution uses the portability owner's source actor,
without an account-directory dependency. The final visual display is WS-07.

### WS-10 integrated validation progress

The [validation handoff](workbook-inspector-validation-handoff.md) records exact
commands, run roots, failed-run resolution, the complete changed-file inventory,
151 individually mapped image reviews, digest acceptance and paired rollout/rollback.
This tracker remains the controlling status and decision artifact.

WS-10 is **DONE**. All substantive implementation corrections, required behavioral
checks, reviewed visual refresh, two ordinary confirmations and final documentation
checks passed. All applicable A001–A027 rows are PASS in the validation handoff.

- `make agent-finalize` passed `20260921T073043Z-p91957` before broader final
  verification. Retained-run maintenance was **SKIPPED** because `RESULTS_DIR`
  was unset; no qualifying full warm check was supplied.
- Final-source `make test-fast` passed **708/708** at
  `20260921T073129Z-p97071`. Integrated checks exposed and corrected exact saved-text
  copy, canonical null/empty/absent preservation and the separation of presentation
  reset from access scope. Regression checks and browser exercises pass.
- All nine failed scenarios from the initial full webserver-backed run
  `20260921T061637Z-p18330` passed their documented focused reruns. Final stateful
  run `20260921T073129Z-p96952` passed 39/42 units; its two transient failures passed
  unchanged in isolation at `20260921T073603Z-p78561` and
  `20260921T074157Z-p46106`. Final accessibility run `20260921T073453Z-p32834`
  passed 18/20 units; its Coordination focus scenario passed unchanged at
  `20260921T074150Z-p32164`. Earlier full stateful 42/42 and accessibility 20/20
  runs remain recorded separately. Failed full runs are never relabeled as passes.
- Revisions canonical-fixture failures were corrected without relaxing production
  admission; the three failed rows passed 4/4 at `20260921T062140Z-p68836` and
  all remaining full-run rows passed. No decoder, reset or backfill was introduced.
- Final generation, drift, generated policy, JSON shape, compatibility, type,
  import-boundary and lint checks pass; exact runs are in the handoff ledger.
  Server/web candidate builds pass at `20260921T073129Z-p97267/-p97300`.
- Baseline and final candidate archives each contain 73 verified server/web files;
  their manifests and SHA-256 digests are in the handoff. No external deployment,
  SQL migration, history rewrite or forced browser reload was performed.
- Visual reconciliation accounts for 255 captures/goldens/active mappings, zero
  orphans/missing/ambiguous mappings and 29 resolved registered fixtures. All 151
  changed images were reviewed; later changed captures were reviewed again in full.
  The Evidence anchor and Network Flow read-readiness fixture corrections remove
  timing-dependent placement/state. Final public refresh `20260921T074355Z-p30667`
  passed 12/12, followed by two fresh ordinary runs `20260921T074928Z-p69256` and
  `20260921T074928Z-p69271`, each 12/12. Final Markdown
  `20260921T075527Z-p42495` and staged/unstaged whitespace checks passed.
- All seventeen configurations and specialized authoring families have named
  source/verification coverage. Optional unbound features retain the explicit
  Core 01 `omit_feature` N/A rationale. Five-task comparison records scripted
  completion and source differences; participant completion/error/scroll/interpretation
  metrics were unavailable and remain usability hypotheses.

Run roots are relative to `.cartulary/test-results/`; manifests, row receipts,
logs and browser reports remain available. No mandatory acceptance is waived.

### WS-09 family and retirement review

| Schema configurations | Saved presentation / authoring boundary | Validation retained |
| --- | --- | --- |
| Timeline | Reading-first Details; source-local Note, Task/Decision, coordination and related-Evidence forms | WS-03/05 plus specialized authoring, exact replay and accepted-write recovery |
| Hosts, Identities | Shared saved Details and Entity relationships; Note and Task/Decision forms | WS-03/04 plus contextual creation source-retention matrix |
| Evidence | Shared Generic saved Details; distinct access/attachment; Note and Task/Decision forms | WS-05 plus contextual and related-Evidence owner tests |
| Notes | Shared Generic saved Details and explicitly named new-record draft; Note source control and local Note form | Note source replacement/clearing, detachment, revocation and recovery |
| Indicators | Shared Generic saved Details; separate canonical proposal, observation capture and lifecycle authoring | Each Indicator family's authoring/reconciliation tests and keyboard/accessibility flow |
| Assessments | Append-only saved values use shared read-only Details; selected record remains the header; append title is local | Stable selection, subject-only follow-on, close/resume/discard and append receipt recovery |
| Task Requests, Decisions | Shared Generic saved Details; dedicated Task lifecycle and Decision supersession; contextual creation | Guarded source/lifecycle review, independent submission and retained receipt tests |
| Parties | Shared Generic saved Details and named new-record draft | Ordinary schema matrix, required minima and continuity |
| Communications Log, Handoff, Status Review, Lesson | Shared Generic saved Details; source-owned coordination forms and contextual entries | Coordination source replacement/clearing, raw values, permission loss and recovery |
| Findings, Investigative Queries, Forensic Keywords | Shared Generic saved Details and named new-record draft; contextual entries where declared | Ordinary seventeen-schema registry and source-owned contextual tests |

Every configuration is accounted for. Optional unimplemented discovery surfaces
retain the explicit WS-08 owner-based N/A inventory. Assessment immutable history
is not ordinary field editing; its read-only facade supplies no mutation commands.

Removed `features/indicators/observationStyles.ts` and
`features/indicators/indicatorLifecycleStyles.ts` after migrating all thirteen
consumers to shared form roles. Removed local input/group/action styles in Note,
coordination, Decision, related-Evidence, Assessment and the generic related-row
presentation. Source-specific textarea resize, select appearance and source text
preservation remain meaningful variants of those roles. Shared group, heading,
field, action and preserved-text exports have named consumers in these families.

The related-row renderer and seed builder remain for their existing checked
row-creation command adapter; contextual source owners still intercept their
specialized families. This effort does not add a form engine or merge their
validation/recovery lifetimes. `Close draft` in an attached Assessment now invokes
the existing detach command next to its explicit append action; discard remains
separate. Generic creation labels now name the target instead of "Commit draft row".

### WS-08 action inventory and scope

The shared dispatcher admits 48 contextual command instances across all seventeen
configurations: 37 row-creation entries, four linked-Note entries, four Indicator
entries, two Timeline capture actions and one Decision supersession action.
They render once in canonical order, with authored labels and outcome descriptions.
Indicator mutation entry points open authoring; read entry points open review.
The other source-owned History, Evidence, mention, merge, reference and lifecycle
controls continue to use their existing commands and admission. `panel_read`
features describe content contributions, not additional buttons.

Core 01 REQ-01-615 explicitly specifies `unsupported_feature_behavior=omit_feature`.
The following optional discovery features have no current bound implementation;
they are **N/A for presentation migration**, not claimed implemented or removed
from the protocol. This scope decision does not waive supported field editing,
History or source-owned controls. No placeholder action or invented query is added.

| Configurations | Optional Workflow discovery not bound by this client | Existing supported destination |
| --- | --- | --- |
| Hosts, Identities | Timeline/Evidence/Assessment surface pivots | Entity Relationships preview and ordinary sheet navigation |
| Evidence, Notes | Linked/source-record surface pivots; Evidence Timeline pivot | Authorized association content and ordinary sheet navigation |
| Parties | Four usage pivots and reference-clear summary | Saved Party Details; reference controls belong to referencing records |
| Assessments | Subject/prior-assessment summary shortcuts | Saved subject Details and explicit History read |
| Communications Log, Handoff, Status Review | Named party/next-report/acknowledgement shortcuts and cross-record review summaries | Declared readable fields and admitted ordinary editors; related creation |
| Decisions, Findings, Lesson | Named owner/state/affected-record shortcuts | Admitted Details editors; Decision supersession retains its dedicated owner |
| Task Requests | Separate decision link/clear shortcuts | Admitted decision-reference editor; dedicated Task lifecycle control |
| Investigative Queries, Forensic Keywords | Named source/result/finding/Timeline reference shortcuts | Admitted collection editors in Details |
| Timeline, Indicators | No additional unsupported Workflow shortcuts | All declared Workflow creation entries use their bound owners |

This is an explicit migration inventory under the adopted omission boundary,
not a claim that the repository implements every feature in its discovery vocabulary.
Future implementations must supply a checked source binding and exact seeded query
or command, then join the same presentation hierarchy. Labels are never dispatch keys.

### Scope and authority

| Session | State / files | Commands or review | Result / next action |
| --- | --- | --- | --- |
| Original audit | Source/screenshot audit at the baseline above | Original evidence below | Historical findings retained; not fresh verification |
| Remediation planning P-01 | Only this tracker changed; four user documents, AGENTS, framework and skill guidance read | Branch/HEAD/status and owner/source inspection | DONE; updated checkpoint before beginning P-02 |
| Remediation planning P-02 | G00–G11 and WS-00–10 defined in this tracker | Owner-to-gap and compatibility review | DONE; each original finding separately accounted for, History unit mapping and migration gates recorded before P-03 |
| Remediation planning P-03 | Documentation validation and final handoff | Completeness, owner review, whitespace and public Markdown lint | DONE; fresh lint `20260921T033650Z-p1857`, four task guides and whitespace passed before implementation authorization |

### Backend boundary

| Session | Inspected / proposed boundary | Result / risk / next action |
| --- | --- | --- |
| P-01 | Revisions History materializer/service/model, provider contributions/catalog, revisionassembly navigation; Core 01 History and Core 02 §15 | Confirmed coarse units; WS-06 must inventory source providers, retained formats and imported actor parity before changing interfaces |

### Frontend boundary

| Session | Inspected / proposed boundary | Result / risk / next action |
| --- | --- | --- |
| P-01 | §2 shell/Details/History/Timeline/feature seams and source guides | Keep shared presentation stateless; navigation uses existing admission and focus; retain source-owned authoring/operations |
| P-02 | Shared-form migration and explicit editor action slots | No new form/retry engine; inventory remaining feature callers at each slice entry |

### Contract and code generation

| Session | Inspected / proposed boundary | Result / next action |
| --- | --- | --- |
| P-01 | Design/view contracts, Revisions OpenAPI source, generated History type, generated-artifact policy | Edit authored owners; generated policy additionally declares `packages/view-contracts/src/generated` and aggregate OpenAPI as generated; do not hand-edit them |
| P-02 | Public History unit correction | API classification and retained browsing generation transition required; no assertion of zero migration until WS-06 evidence |

### Tests and harness

| Session | Commands / evidence | Result / artifacts / next action |
| --- | --- | --- |
| P-01 | Four `make task-guide ROLE=module-author OWNER=...` calls: web.workbook, module.workbook, package.ui, module.revisions | All passed; routing only |
| P-01 | `make explain-test-owner OWNER=web.workbook` and `OWNER=package.ui` | Both passed; current catalog families/rows inspected |
| P-01/P-02 | Manual source searches, reads and `make help` | Located exact source paths and public command surface; confirmed 24 Timeline fields; no product verification target executed |
| P-03 | Documentation lint, whitespace and local-link review | PASS: Markdown `20260921T033650Z-p1857`; staged/unstaged whitespace and local references reviewed. Planning evidence only. |

### Security and authorization

| Session | Evidence / retained obligations | Result / next action |
| --- | --- | --- |
| P-01/P-02 | Core 04 authorization/Evidence clauses, explicit concealed region unions, History unit data boundary | No confirmed new authorization defect claimed. Require no hidden destinations/counts, no raw snapshot disclosure, source-authorized actor/value presentation and current request-time rights. |

### Open risks and next session

| Session | State | Next action |
| --- | --- | --- |
| Planning handoff | Production code, adopted docs, tests, contracts and generated artifacts unchanged | Later implementation task starts WS-00; inspect current HEAD/dirty state and this tracker, close owner decisions, then execute slices with tracker updates |

### Historical audit execution and advisory dispositions

The following record predates this remediation pass. Its commands, timings and
screenshot reviews remain historical evidence only.

`make task-guide ROLE=module-author OWNER=web.workbook` completed successfully
and selected `make test-slice OWNER=web.workbook` as the focused entry point.
`make task-guide ROLE=module-author OWNER=module.workbook` also completed and
identified the focused service-backed and browser verification routes.
Repository/source inspection and seven local image inspections completed;
all 28 document source/image links resolved locally.

One cache-suppressed offline UX query, `progressive disclosure hierarchy dense
navigation`, returned five results with no fallback. Applied dispositions:

- **ADOPT** semantic heading hierarchy and unobscured focus through R002/R013.
- **ADAPT** sticky navigation to the inspector's internal scrollport through
  R014/R018; generic page-padding examples do not govern workbook geometry.
- **ADAPT** typography hierarchy through R011/R019 using Cartulary's existing
  roles, rather than an upstream scale.
- **REJECT for this scope** breadcrumb navigation and hero-title balancing:
  neither addresses the inspector's record/section structure. No generated
  palette, framework, font, icon dependency, or design-system output was used.

Only this audit document was authored. Product tests, browser captures,
accessibility scans, generated-artifact changes, and full suites were skipped
because this task is planning only. `agent-finalize` was not run: no broader
product verification was performed, and its generated-structure refresh would
exceed this documentation-only scope. Retained-run maintenance was skipped
because `RESULTS_DIR` was unset. Documentation lint results are recorded below
for this audit. No product-readiness or accessibility-pass claim is made.

`make lint-markdown` passed (84.4 seconds). Run root:
`.cartulary/test-results/20260921T031046Z-p81814`; summary:
`adhoc/lint-markdown/tool-run-summary.json`. The new document was also inspected
with a no-index whitespace check; no whitespace diagnostics were emitted.
The documentation run reported no failing target. Final working-tree review
showed only this new Markdown document.

## 11. Open decisions, blockers and risk register

### Next-iteration risks and resolved defaults

There is no blocker to this document update. The risks below are implementation
obligations, not reasons to retain obsolete code. Existing RB-001–RB-009 retain
their previous-iteration status; new risks use RB-010 onward.

| ID | Risk / decision | Why it matters | Owner / evidence needed | Status / gate |
| --- | --- | --- | --- | --- |
| RB-010 | Competing History read authority and erased paging | Divergence can break review/recovery | G12; all History/Timeline/Recovery consumers | CLOSED; WS-12 exit |
| RB-011 | Object identity substitutes for read provenance | Conversion/cloning can defeat retirement | G15; query and operation source acceptance evidence | CLOSED; WS-14 accepted provenance and authority matrix; no disclosure claimed |
| RB-012 | Bounded paging changes coalescing or selectors | Audit/reversal meaning must stay complete | G16; canonical descriptors, source facts and expected page fixtures | CLOSED; WS-15 complete semantic pages and measured bounded projection |
| RB-013 | Cursor version transition and additive indexes | Disposable reads must restart without losing local work | HTTP/Revisions owners; compatibility and rollback tests | CLOSED; typed protected position, ordinary restart/recovery evidence, unchanged existing indexes |
| RB-014 | Intermittent failures closed by reruns alone | Unexplained focus/recovery behavior remains | G17; controlled response-order regressions and ordinary suites | CLOSED; WS-16 causal fixes and deterministic/ordinary evidence; unrelated Network Flow excluded |
| RB-015 | Dead-code removal drops live or registered consumers | Reference counts alone are insufficient | G13; full exports/imports/registration ledger and named consumers | CLOSED; WS-17 full symbol ledger and migrated consumers |
| RB-016 | Historical results mistaken for current readiness | New exits have no implementation evidence | G18; new I3 acceptance and fresh digest reassessment | CLOSED; WS-18 current aggregate/acceptance, compatibility and final handoff evidence |
| RB-017 | Scope and archive dependency | Avoid unrelated release work and unnecessary document recovery | User selected inspector subsystem plus backend paging and excluded archived handoff | DONE; resolved planning defaults; no archive restoration required |

An actual adopted-owner contradiction becomes `BLOCKED: owner contradiction`
only for its dependent chain. None was found for the planned boundaries.
Participant usability benefit remains an unmeasured hypothesis; no new study is
required to finish this document update and no participant outcome is fabricated.

### Previous-iteration section 11 record (historical)

There is no blocker to completing this plan. Implementation dependencies below
must be resolved in their named workstream; they do not justify inventing payload
facts or silently dropping required behavior.

| ID | Decision / risk | Why it matters | Needed authority / evidence | Status and resolution gate |
| --- | --- | --- | --- | --- |
| RB-001 | Adopt proposed navigation/disclosure/default copy and layout rules | These are intentional new UX behaviors, not current defects | Design owner and §7 complete decision table | DONE; user-approved defaults adopted in design §12.7, D-AC-082–086 |
| RB-002 | Close History unit semantics and safe visibility | Current payload lacks required detail; arbitrary units are not a durable public contract | Core 01/02/03, Revisions/source owner review and payload security fixtures | DONE; adopted Core 01/02/03 clauses and WS-06 closed contract, source projectors, safe disclosure and canonical validation implemented and tested |
| RB-003 | Supported HTTP consumers and released-contract classification | Cannot assume a tightened schema or new unit shape is compatible | Immutable OpenAPI release comparison; actual supported consumer inventory | DONE; eleven reviewed History fingerprints in pending 2.0.0 change set; immutable 1.0.0 baseline preserved; bundled client/server cutover |
| RB-004 | Retained History reconstruction and imported attribution parity | Older snapshots or undeclared actor fields may defeat a superficial UI fix | Source-provider/retained-data inventory and imported-attribution owner clauses | DONE; ten canonical record snapshot types / fourteen targets require source projectors; imported attribution and retained rollback verified; incomplete canonical fixtures corrected without a runtime decoder |
| RB-005 | History read equality across deployment | Changed diff payload can look like altered immutable committed content | Browsing generation and compatibility/rollback tests | DONE; opaque representation generation checked before item equality, fresh reads preserve operation owners, and WS-07 renders every admitted semantic unit with scoped actions |
| RB-006 | Full feature-family coverage beyond the original sample | Generic appearance does not prove specialized semantics | §8 17-schema and authoring-family matrix | DONE; WS-09 seventeen-schema/family inventory, source-local recovery tests and WS-10 integrated evidence |
| RB-007 | Baseline services/browser renderer and current test health | Historical passes/failures are not current evidence | Fresh focused runs; explain-target/run artifacts if failure | DONE; fresh owner/browser checks and resolved fixture/source failures are recorded in the validation handoff; no historical result substituted for current verification |
| RB-008 | Claimed reduction in effort/confusion | Design judgment is not a measured result | Comparative task review with disclosed participants/method | DONE with limitation; same five tasks compared through source/scripted behavior and reviewed captures; no participants, timed completion, error-rate or scroll-distance measurements; residual hypothesis recorded |
| RB-009 | Rich Evidence browsing without existing authorized item data | Count cannot manufacture an item catalog or access | Separate Evidence read capability proposal | DEFERRED outside required workstreams |

## 12. Binary completion criteria

### Completed planning-task criteria (historical; superseded by execution scope)

- P-04 and P-05 are DONE with actual fresh documentation evidence.
- Exactly this tracker changes; no source, tests, owner specs, projections,
  generated artifacts, dependencies, migrations or archived handoff changes.
- The twelve numbered sections and historical G00–G11/WS-00–WS-10 checkpoints
  remain intact. Their DONE/PASS statements are labeled historical.
- All 74 target files are accounted for, the test-only fixture is identified,
  connected seams/exclusions are explicit and each new gap has an owner,
  remediation, areas, rationale/benefit, compatibility, risk and binary exit.
- G12–G18, WS-11–WS-18, I3 acceptance, retirement and risk ledgers agree.
  At that planning checkpoint all new implementation workstreams remained TODO;
  the subsequent user implementation request supersedes that restriction.
- Public task guides, fresh Markdown lint, whitespace and single-file diff
  checks have recorded results; skipped product/finalization checks have reasons.

### Next-iteration implementation completion — mandatory WS-18 exit

- WS-11 through WS-18 are DONE, with entry and passing-exit checkpoints saved
  before the next workstream begins. No mandatory gap is silently deferred.
- Owner decisions and typed contracts agree; a new source extends existing
  catalog/presentation boundaries rather than a second authority or state engine.
- I3-01–I3-15 pass with fresh evidence. Every applicable A001–A027 digest row is
  reassessed for the new candidate; optional N/A has owner/scope rationale and
  no mandatory outcome remains blocked.
- History has one accepted browsing representation and mandatory page
  reconciliation. Captured operations and receipts remain independently owned.
- Closed presentation modes, pure saved reading and explicit edit attachment
  cover all seventeen schemas without no-op editing or duplicate section policy.
- Retained rows carry source-accepted authority provenance; old/cloned observations
  cannot acquire current authority through presentation changes.
- Bounded History pages preserve complete semantic units, coalescing, imports,
  ordering, attribution and selectors; measured work excludes unrelated retained
  snapshots. Required selected change-set dependency cost is disclosed separately.
- Every relevant intermittent failure has a causal fix and deterministic
  regression; ordinary affected suites pass without weakened assertions or
  retries replacing evidence.
- Superseded reducers, page-less paths, cursor scans, props and unnecessary
  exports are removed. Test-only helpers reside in test support. Every retained
  shared export has a named consumer and purpose; no deprecated forwarding path.
- Applicable browser, accessibility and visual validation covers the new
  candidate. Golden updates follow reconciliation, public update, review and
  two successful ordinary visual runs. No unsupported usability benefit is claimed.
- Coordinated server/browser rollout and rollback preserve canonical history,
  draft lifetimes, request bytes and receipts. Old read cursors restart safely;
  no forced reload, legacy decoder, data reset, history rewrite or snapshot backfill.
- Handoff records candidate revision, owners, changed paths, substantive edits,
  exact commands/results/artifacts, resolved failures, skipped checks and reasons,
  retirement, rollout/rollback, residual hypotheses and next maintainer action.
  Retained-run maintenance is recorded as used only with qualifying RESULTS_DIR;
  otherwise explicitly skipped. Archived handoff recovery is not an exit condition.

### Previous-iteration section 12 record (historical)

### Planning completion

- P-01, P-02 and P-03 are DONE with updates recorded in sequence.
- All nine original audit findings have G01–G09 coverage; specification closure,
  History data completeness and validation gaps are covered by G00/G10/G11.
- Each gap has remediation, change areas, rationale, long-term benefit,
  compatibility/migration impact, unresolved risk and testable exit.
- Every named target seam is inventoried; uninspected neighbors are explicitly
  excluded or assigned a workstream discovery gate. Owner, implementation and
  verification identities are distinguished.
- Each implementation slice is its own workstream with dependencies, risks,
  validation, retirement and rollback; WS-10 is the mandatory final slice.
- Fresh documentation validation and final diff scope are recorded. Only this
  tracker changed; its pre-existing staged content remains staged without index
  manipulation. No product PASS or implementation completion is claimed.

### Overall remediation completion after implementation

- WS-00 through WS-10 are DONE, with each exit and tracker update recorded before
  the next workstream began. No mandatory gap is silently DEFERRED or DROPPED.
- Adopted owner text is coherent, projections are generated from authored machine
  inputs, and code/tests/docs agree without any product dependency on Markdown.
- G01–G10 behavioral exits pass across the schema/state/authority matrix. History
  displays required source-owned detail; a cosmetic fallback cannot close G10.
- All migrated callers use the intended shared presentation. Redundant style,
  label and layout alternatives are removed; each retained adapter has a named
  consumer and removal gate. No indefinite dual inspector implementation remains.
- Applicable A001–A027 outcomes are PASS with evidence; N/A is justified by scope
  and owner, and no applicable BLOCKED row remains. Every required test failure
  has been fixed or resolved with evidence rather than waived as historical.
- Golden updates follow ordinary-run reconciliation, authorized update, image
  review and ordinary validation. First-viewport Details/edit/retained-work views
  are covered. Product readiness, visual support and usability observations are
  reported separately.
- History compatibility/migration and rollback are proven for supported clients
  and retained data. Logical audit identities and captured operation recovery
  survive the supported cutover. No unauthorized snapshot/actor/Evidence data is
  newly exposed.
- Final handoff states adopted decisions, changed paths, substantive edits,
  exact commands/results/run roots, failures and resolutions, skipped checks with
  reasons, compatibility/retirement, rollout/rollback, residual hypotheses and
  next maintainer action. RESULTS_DIR maintenance is explicitly recorded as used
  or skipped. Only then is the overall effort complete.
