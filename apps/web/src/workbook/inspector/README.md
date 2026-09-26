# workbook/inspector/

[Parent](../README.md) · [Source overview](../../README.md)

Canonical inspector subjects, declared capability admission, related-record workflows, and History composition.

Canonical subjects and declared feature capabilities determine inspector content.
Source-specific composition belongs in [features](../features/README.md) and
[Timeline](../timeline/README.md). Shared History operation lifetime belongs in
[history](../history/README.md); this directory owns its inspector binding.

## Subdirectories

| Directory | Responsibility |
| --- | --- |
| [presentation/](presentation/README.md) | Shared inspector shells, panels, actions, feedback, and History event rendering. |

## Declared inspector composition

| File | Responsibility |
| --- | --- |
| [canonicalInspectorAdmission.ts](canonicalInspectorAdmission.ts) | Admits exact canonical inspector features and derives stable feature identity. |
| [inspectorCapabilityResolver.test.ts](inspectorCapabilityResolver.test.ts) | Tests exact canonical schema/feature tuples and contextual or History action classification. |
| [inspectorCapabilityResolver.ts](inspectorCapabilityResolver.ts) | Resolves canonical inspector capabilities and supported record History actions. |
| [useWorkbookInspectorCoordinator.test.tsx](useWorkbookInspectorCoordinator.test.tsx) | Direct tests for retargeting, lifecycle invalidation, action completion, idempotent close, and focus restoration. |
| [useWorkbookInspectorCoordinator.ts](useWorkbookInspectorCoordinator.ts) | Schema-bound inspector lifecycle coordinator for explicit open/close, stable row subjects, ordered feature invalidation, action completion, and focus restoration ports. |
| [WorkbookInspectorContextualActions.tsx](WorkbookInspectorContextualActions.tsx) | Renders admitted contextual inspector actions from canonical feature declarations. |
| [WorkbookInspectorDetails.tsx](WorkbookInspectorDetails.tsx) | Explicit editor attachments, focus and commands supplied by source owners. |
| [WorkbookInspectorSavedDetails.tsx](WorkbookInspectorSavedDetails.tsx) | Independent saved-value reading in declared order with optional presentation slots. |
| [WorkbookInspectorDeclaredPanelList.tsx](WorkbookInspectorDeclaredPanelList.tsx) | Renders declared inspector panels in their owner-defined order. |

## Related-record authoring

| File | Responsibility |
| --- | --- |
| [InspectorCreateRelatedWorkflow.tsx](InspectorCreateRelatedWorkflow.tsx) | Inspector related-record authoring presentation over declared feature commands. |
| [inspectorRelatedRecordModel.test.ts](inspectorRelatedRecordModel.test.ts) | Tests related-record field and input seeds. |
| [inspectorRelatedRecordModel.ts](inspectorRelatedRecordModel.ts) | Shared draft seed builder and retained-owner presentation types. |
| [useInspectorCreateRelatedWorkflow.test.tsx](useInspectorCreateRelatedWorkflow.test.tsx) | Checks every authored creation feature has exactly one retained owner and unbound actions cannot dispatch. |
| [useInspectorCreateRelatedWorkflow.ts](useInspectorCreateRelatedWorkflow.ts) | Presentation attachments to retained contextual Task/Decision, coordination and Note owners. |

## Record History

| File | Responsibility |
| --- | --- |
| [useWorkbookRecordHistoryController.ts](useWorkbookRecordHistoryController.ts) | Binds shared History reads/actions to the canonical Inspector subject or an explicit Recovery read locator and the existing History owner. |
| [useWorkbookRecordHistoryFocus.ts](useWorkbookRecordHistoryFocus.ts) | History action focus coordination using stable action and rollback identities. |
| [workbookHistoryPresentationModel.test.ts](workbookHistoryPresentationModel.test.ts) | Tests consistent History technical-field ordering and rollback/pending labels. |
| [workbookHistoryPresentationModel.ts](workbookHistoryPresentationModel.ts) | Builds History event, rollback-label, and pending-operation presentation. |
| [workbookHistoryRecovery.characterization.test.tsx](workbookHistoryRecovery.characterization.test.tsx) | Tests acknowledgement before refresh completion and selection-lifetime fencing of admitted effects. |
| [WorkbookInspectorRecordHistory.test.tsx](WorkbookInspectorRecordHistory.test.tsx) | Tests advertised History loading, stable rollback selectors, and exact tombstone restore versions. |
| [WorkbookInspectorRecordHistory.tsx](WorkbookInspectorRecordHistory.tsx) | Record History inspector composition and browsing controls. |
| [workbookRecordHistoryModel.test.ts](workbookRecordHistoryModel.test.ts) | Tests stable subject references and rejection of events from obsolete phases or identities. |
| [workbookRecordHistoryModel.ts](workbookRecordHistoryModel.ts) | Subject attachment and unsubmitted review over the sole accepted History browsing representation; display phases are derived. |
| [useWorkbookRecordHistoryState.ts](useWorkbookRecordHistoryState.ts) | Allocates one presentation state instance shared with the controller and, for Timeline, subject composition. |
| [workbookRecordHistoryOperation.ts](workbookRecordHistoryOperation.ts) | Builds inspector feedback for completed record History operations. |
| [workbookRecordHistoryOwnerEffects.ts](workbookRecordHistoryOwnerEffects.ts) | Coordinates History owner effects with inspector subject and presentation lifetime. |
| [WorkbookRecordHistoryPresentation.tsx](WorkbookRecordHistoryPresentation.tsx) | Loaded record History presentation with action and pending-state bindings. |

## Subjects and feedback

| File | Responsibility |
| --- | --- |
| [workbookInspectorErrorModel.test.ts](workbookInspectorErrorModel.test.ts) | Tests safe primary messages, decoded version-conflict identity, and sanitized details. |
| [workbookInspectorErrorModel.ts](workbookInspectorErrorModel.ts) | Shared inspector error and operation-feedback presentation models. |
| [workbookInspectorSubject.ts](workbookInspectorSubject.ts) | Constructs and compares canonical live/deleted inspector subjects and row bindings. |

## Ordinary edit drafts

| File | Responsibility |
| --- | --- |
| [WorkbookInspectorDraftStore.ts](WorkbookInspectorDraftStore.ts) | Account/incident-scoped raw drafts, explicit attachment, dependency review and revision-fenced retirement. |
| [useWorkbookInspectorEditDraft.ts](useWorkbookInspectorEditDraft.ts) | Canonical field/action binding and presentation detachment, independent from query object identity. |
| [WorkbookInspectorEditControl.tsx](WorkbookInspectorEditControl.tsx) | Accessible ordinary value control with retained reference identity and explicit clear intent. |
| [WorkbookInspectorDraftFeedback.tsx](WorkbookInspectorDraftFeedback.tsx) | Local Resume/Discard and changed saved-value review. |
| [prepareWorkbookInspectorChange.ts](prepareWorkbookInspectorChange.ts) | Patch capability and scalar/action serialization; creation remains separately admitted. |
| [workbookInspectorOrdinaryAttention.ts](workbookInspectorOrdinaryAttention.ts) | Immutable draft/PATCH attention projection; preserves distinct newer authoring and captured operations. |
| [WorkbookInspectorDraftStore.test.ts](WorkbookInspectorDraftStore.test.ts) | Draft, capability, review, security and captured-revision regressions. |
| [useWorkbookInspectorEditDraft.test.tsx](useWorkbookInspectorEditDraft.test.tsx) | Frozen field/action binding through refresh, detachment, no-row and explicit return. |

[WorkbookExplicitPatchRecovery.tsx](WorkbookExplicitPatchRecovery.tsx) keeps
ordinary explicit attempt recovery reachable on its surface after inspector closure.

## Query windows

`useRetainedInspectorRow.ts` retains one source-accepted observation independently
of query-window membership. Query and operation adapters attach record, version,
and the existing runtime account/session/incident scope and read-revocation epoch when accepting
evidence. Conversions preserve that observation; rendering cannot create it.
Admission filters all candidates by identity, version and authority before choosing
the newest. An absent query member does not establish deletion. Authority changes
require fresh evidence, while interaction permissions and window eviction do not
revoke readable observations. Late write receipts retain their dispatch scope and
independent recovery lifetime; they cannot grant current read authority.

## Explicit presentation and feedback

Every live declared panel requires a feature-owned presentation model. Data state
and access are independent: read-only content remains readable, while concealed
panels expose neither content nor feedback. Shared rendering does not inspect
React children to guess emptiness or dispatch requests. A missing contribution
fails coverage instead of becoming an empty panel.

`useWorkbookInspectorFieldFeedback.ts` captures canonical field/action identity,
authoring revision and attachment with each attempt. Only matching authoring
receives field feedback; unassociated failures stay beside the originating action.
Changing raw text, null intent or reference selection retires obsolete feedback.
The draft and mutation owners retain recovery independently of this presentation.

Ordinary retained failure notices remain readable without another live-region
announcement. The attached field or originating action announces the rejection;
uncertain-operation and saved-refresh recovery retain their own status semantics.

## Saved values and region delivery

`WorkbookInspectorDetails.tsx` presents accepted fields in contract order and one
explicit editor attachment. Update and Ctrl/Cmd+Enter submit. Blur retains the
attached editor; field switching, closing and Escape detach without submitting. Draft and PATCH owners retain
work and receipts independently. `useWorkbookInspectorNotice.ts` asks the retained
operation owner's notice ledger to admit one announcement for an attempt and
transition. A remounted view does not own or reset that ledger.

Panel contributions contain nonempty ordered region descriptors. Synchronous
accepted values use snapshot regions; subscribed owners deliver a typed region
model through a render callback. This delivery slot preserves component identity
and existing lazy reads without an aggregate ready wrapper or another cache.
Each readable region distinguishes data from commands and authoring. Concealed
variants contain no renderable payload. Unrequested History remains neutral.

## Closed shell and saved-value presentation

`WorkbookInspectorShell` accepts exactly one saved, empty or creation variant.
Saved subjects require the admitted section sequence; empty selection has no
sections; creation requires its source context and local admitted sections.
The same sequence drives navigation and mounted content. Source compositions
resolve the variant before rendering, and creation context must match the schema.

`WorkbookInspectorSavedDetails` consumes accepted rows independently of editing.
`WorkbookInspectorDetails` supplies field attachments through an explicit
presentation contract of controls, commands and feedback. It does not import a
draft hook or draft store type. Source owners retain authoring state and supply
commands; append-only sources use the reader directly. Preserve null, unloaded,
empty text, false and zero as distinct saved values.

Query port identity follows accepted read scope, including its revocation epoch;
an unchanged authorization recheck does not reset a browsing chain or interrupt
an acknowledged operation's refresh. Timeline partial action receipts carry the
transport's dispatch-time observation. They can advance a saved-row projection
only from a matching admitted base version and scope; receipt retention itself
remains independent of whether protected content may currently be shown.

Related creation requires an explicit retained owner in the capability resolver.
Contextual Task/Decision, coordination, Note and Timeline Evidence forms submit
through their own attachment leases; Assessment follow-on authoring uses its
Assessment owner. The inspector owns no fallback draft reducer, generic form or
direct create transport. Unknown additive actions are omitted. New features must
provide owner admission, captured attempts, receipts, authority and recovery before
joining the typed binding and authored-feature coverage check.

## History read continuity

[useWorkbookHistoryReadContinuity.ts](useWorkbookHistoryReadContinuity.ts) owns
presentation scroll/focus continuity for History reads. Its semantic control
anchor is scoped to the mounted panel, record/view, authorization lifetime and
request/cursor generation. Matching renders preserve its measured position;
user interaction, replacement, detachment or obsolete completion cancels it.
The initiating recovery control remains mounted and focusable while its admitted
read is pending, independently of the cleared failure flag. It is busy and
unavailable for repeat activation; a repeated failure restores the current
recovery action. Success moves focus to a surviving read control only while the
initiating interaction still owns it. If a newer interaction cancels that move
while the old control remains focused, the control shows a neutral completed
label until focus leaves. Focus restoration uses `preventScroll`. Accepted pages,
cursor chains and read admission remain with the retained History owner. All
four browser History surfaces share the continuation/retry and late-response
checks.

## Review and attachment lifetime

The shared coordinator publishes the same subject-change classification to its
owner-reset callback and reducer. `record_updated` advances `reviewGeneration`
for a new version of the same live record while preserving
`attachmentGeneration`. Forms, confirmations, version-bound navigation, and
registered inspector elements use review/version identity. Completion continuity
uses attachment identity plus its operation and current interaction/authority
scope. Closing, retargeting, lifecycle replacement, and explicit reopening cannot
revive an older attachment. Ordinary retained authoring remains source-owned.

Consumers that combine an inspector subject with its lifecycle metadata receive
the coordinator snapshot as a unit. Timeline observation management must not
combine a freshly selected subject with an earlier snapshot's generations or
cause during a React layout transition.
