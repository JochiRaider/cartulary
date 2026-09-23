# workbook/runtime/

The runtime also owns ordinary committed-cell drafts through `gridDrafts`.
Presentation detachment does not retire them. Current authority controls access;
account/incident retirement clears them. Managed queue and accepted-row tests in
`WorkbookGridAutosave.test.ts` cover revision settlement and first-dispatch bases.

Managed enqueue returns an admission ticket whose completion represents server
acknowledgement. The driver retains each field's contributed draft revision;
coalescing explicitly supersedes replaced contributions. The queue prepares a
base synchronously only before first dispatch. Transport capture retains the
route, body and transaction identity through uncertainty. Complete accepted
receipts enter the existing committed-record capability before dependent work
is released. Surface refresh runs as a separate retained read obligation.

[Parent](../README.md) · [Source overview](../../README.md)

Workbook mutation lifetime, pending replay, conflict recovery, write coordination, and status projection.

The runtime coordinates pending work, explicit operations, conflicts, and
surface refresh obligations. It retains the ordinary creation owner by incident
and account, outside Timeline dispatch admission. It stays independent of Timeline implementation;
surface-specific commands enter through registered semantic capabilities.

The [pending queue](pending/README.md) owns memory-local FIFO replay, admission, coalescing and conflict settlement.

## Composition and lifecycle

`createWorkbookMutationRuntime` is the construction entry for production and
fixtures. Shared queue, conflict, driver, scheduling and surface resources exist
before `WorkbookMutationFeatureAssembly` creates the fixed feature set. Callbacks
and subscriptions activate only after all peers exist. Source owners retain their
own validation, captured attempts and receipt/reconciliation policies.

Retained owners expose `unsettledMutationCount` for admitted writes awaiting
settlement. Recovery and local action availability use their subject-specific
snapshots and capabilities. An acknowledged receipt with pending or failed
refresh is read recovery, not another in-flight mutation. Aggregate pending,
blocked and uncertain counters are not supported feature-owner interfaces.
Presentation has no manual mutation-reporting capability. The runtime's private
conflict-submission accounting covers its real request only and ends before read
recovery. Timeline History reads retain source-write ordering without contributing
another mutation to the status strip.

`WorkbookFeatureLifecycle` requires a lifecycle contribution for every feature.
The explicit PATCH contribution adapts incident closure to its authority API;
Timeline supplies its specialized authority projection at its construction edge.
Accepted read authority belongs to the runtime and is independent of individual
feature initialization or local operation denial. Terminal retirement is one
idempotent transition, including source subscriptions, deferred waits, retry,
queued microtasks and mounted surface registrations.

## Files

| File | Responsibility |
| --- | --- |
| [useWorkbookMutationRuntime.ts](useWorkbookMutationRuntime.ts) | React external-store subscription hook for shell-owned workbook mutation state. |
| [WorkbookClientTransactionLedger.ts](WorkbookClientTransactionLedger.ts) | Bounded client-transaction identity retention and pending-queue settlement lookup. |
| [workbookConflictModel.ts](workbookConflictModel.ts) | Same-field conflict parsing and common envelope types. |
| [WorkbookConflictStore.ts](WorkbookConflictStore.ts) | Conflict registration, compatible draft preservation, refresh commands, and panel state. |
| [WorkbookExplicitPatchOwner.ts](WorkbookExplicitPatchOwner.ts) | Neutral retained record PATCH admission, immutable capture, exact replay and read-only reconciliation; source owners contribute review and effects. |
| [workbookLifecycleModel.ts](workbookLifecycleModel.ts) | Shared load, refresh, save, conflict, and recovery lifecycle reducer. |
| [WorkbookManagedPatchDriver.ts](WorkbookManagedPatchDriver.ts) | Managed-patch admission, transport dispatch, settlement, local projection, conflict registration, and refresh. |
| [WorkbookMutationDriverRegistry.ts](WorkbookMutationDriverRegistry.ts) | Closed managed-patch/Timeline-row owner envelopes, exact driver registration, duplicate rejection, and absence-safe dispatch. |
| [createWorkbookMutationRuntime.ts](createWorkbookMutationRuntime.ts) | Fixed construction entry with injectable clock, scheduler and transport. |
| [WorkbookMutationFeatureAssembly.ts](WorkbookMutationFeatureAssembly.ts) | Concrete feature construction through bounded shared coordination capabilities. |
| [WorkbookFeatureLifecycle.ts](WorkbookFeatureLifecycle.ts) | Complete typed feature lifecycle and subscription membership. |
| [WorkbookMutationRuntime.ts](WorkbookMutationRuntime.ts) | Shell-lifetime facade over queue coordination, conflicts, surface registration, managed patches, retry, transaction settlement, and status projection. |
| [WorkbookMutationRuntimeRegistry.ts](WorkbookMutationRuntimeRegistry.ts) | App-owned registry retaining one incident-scoped mutation runtime across shell detachment and retiring it on scope replacement. |
| [workbookMutationStatusProjector.ts](workbookMutationStatusProjector.ts) | Pure queue/conflict/explicit-operation status and save-state projection. |
| [workbookPendingMutationSettlement.ts](workbookPendingMutationSettlement.ts) | Maps semantic mutation failures to common queue settlement outcomes. |
| [workbookPendingReplayRuntime.ts](workbookPendingReplayRuntime.ts) | Pending-replay runtime state, admission contracts, and refresh barriers. |
| [WorkbookRetryScheduler.ts](WorkbookRetryScheduler.ts) | Injected, single-flight retry scheduling and cancellation. |
| [WorkbookRuntimeLifecycle.ts](WorkbookRuntimeLifecycle.ts) | Listener lifetime, coalesced drain notification, and terminal disposal. |
| [workbookRuntimePorts.ts](workbookRuntimePorts.ts) | Clock and scheduler ports with the browser composition defaults. |
| [WorkbookSurfaceRegistry.ts](WorkbookSurfaceRegistry.ts) | Surface command registration, replacement-safe cleanup, and retained refresh debt. |

## Tests

| File | Responsibility |
| --- | --- |
| [workbookConflictModel.test.ts](workbookConflictModel.test.ts) | Tests for conflict parsing, queue entries, resolution payloads, and collection actions. |
| [WorkbookEntityMergeAdmission.test.ts](WorkbookEntityMergeAdmission.test.ts) | Tests merge coordination with queued/direct writes and overlapping participant reservation. |
| [WorkbookExplicitPatchOwner.test.ts](WorkbookExplicitPatchOwner.test.ts) | Tests explicit Task patch reservation, guarded-field review, and exact uncertain replay. |
| [WorkbookMutationRuntime.test.ts](WorkbookMutationRuntime.test.ts) | Tests for shell-lifetime queue retention, autosave, refresh debt, conflicts, and mutation coordination. |
| [WorkbookGridAutosave.test.ts](WorkbookGridAutosave.test.ts) | Admission versus acknowledgement, coalesced contributors, successive revisions, guarded preparation, immutable replay, refusal, discard and read recovery. |
| [WorkbookRuntimeResponsibilities.test.ts](WorkbookRuntimeResponsibilities.test.ts) | Deterministic tests for responsibility boundaries, injected time/scheduling, registration cleanup, conflict drafts, transaction settlement, and disposal. |

## Retained workbook batches

| File | Responsibility |
| --- | --- |
| [workbookBatchOperation.ts](workbookBatchOperation.ts) | Gesture admission, prepared plan, immutable attempt, complete receipt and presentation-independent lifecycle contracts. |
| [WorkbookBatchOperationOwner.ts](WorkbookBatchOperationOwner.ts) | Incident/account retention, delivery deduplication, record/type reservations, prerequisite autosaves, explicit exact retry and reads-only acknowledged recovery. |
| [WorkbookBatchOperationOwner.test.ts](WorkbookBatchOperationOwner.test.ts) | Exact capture/replay, overlap, conflicts-only acceptance, read debt, role/closure/session changes and obsolete callbacks. |
| [WorkbookSurfaceRegistry.test.ts](WorkbookSurfaceRegistry.test.ts) | Registration, authorization and debt generation fencing plus failed-read/remount recovery. |

The owner serves Timeline paste, selected-cell clear, Hosts/Identities paste, fill
and collection tagging.
Source planners retain semantic fields, rows, values and create/reuse rules.
Later overlapping actions wait through uncertainty and unresolved conflicts.
Batch boundaries prevent pending autosave coalescing across admission; preceding
local writes may advance only undispatched versions. The prerequisite identity
set is fixed at admission, even while an explicit preceding save is unfinished;
later edits cannot become prerequisites or overtake an overlapping batch. The conflict store keeps
batch grouping separate from compound-operation restrictions. Unresolved payloads
are never evicted. Settled receipts may be released after their read/conflict
obligations; the latest completed receipt per surface remains available.

Accepted batch predecessors are delivered through the registered source driver
before presentation or later queue dispatch. Timeline advances only later local
clear successors and preserves dispatched attempts and conflicted fields. Batch
receipts apply as one mounted projection commit, including virtualized members.

Surface registrations follow mounted lifetime and dereference current callbacks.
Refresh completion must match registration, authority and debt generations.
Late effects do not restore old selection or overwrite newer editor drafts.

Ordinary inspector raw drafts live in `inspector/WorkbookInspectorDraftStore`.
Task compound drafts and validation stay in Coordination; Party review and
recovery reads stay in Parties. `WorkbookSurfaceRegistry` is the sole mounted
refresh/debt registry. Explicit patches do not carry a Task refresh callback.

An explicit source contribution may own complete reconciliation, including its
source reads and mounted refresh or deferred surface debt. Otherwise the neutral
owner requires the originating surface refresh. Party preparation excludes only
its own reservation ID; other retained operations still block admission.

The committed cache accepts validated contiguous collaboration patches for every
known Core view before surface attachment. A missing predecessor raises only the
version floor. The current source baseline must then be read before preparation.
`surfaceRefreshRequired` and `refreshSurface` expose the existing surface registry
under current read authority; the notice is presentation only.

Local stale-field validation may admit a successor behind an existing grid write:
a collaboration echo is committed evidence but does not settle that predecessor's
HTTP attempt. Preparation waits for predecessor settlement and then validates the
original field/dependency baseline. Pending feedback does not offer premature
review of the predecessor's own echoed value.

## Save facts and recovery discovery

`workbookMutationStatusProjector.ts` derives Saved/Syncing/Conflict independently
of navigation attention. Owner `unsettledMutationCount` excludes retained unsent
authoring and acknowledged refresh reads; existing admission counts retain their
execution meaning. Only FIFO units enter replay counts. Feature review alone is
not a same-field conflict. `workbookBatchRecoveryItems.ts` projects one batch item
including its receipt and conflicts; the surface registry exposes only remaining
read debt. Conflict storage owns drafts and resolution, never panel-open state.

`workbookBatchOutcome.ts` owns factual batch outcome and requested-scope wording
for Recovery list/detail. Its focused tests cover every Timeline operation and
receipt/conflict/read state. `workbookBatchRecoveryItems.ts` adapts those facts
to navigation without changing save-state accounting.
