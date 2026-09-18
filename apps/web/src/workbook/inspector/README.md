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
| [WorkbookInspectorDeclaredPanelList.tsx](WorkbookInspectorDeclaredPanelList.tsx) | Renders declared inspector panels in their owner-defined order. |

## Related-record authoring

| File | Responsibility |
| --- | --- |
| [InspectorCreateRelatedWorkflow.tsx](InspectorCreateRelatedWorkflow.tsx) | Inspector related-record authoring presentation over declared feature commands. |
| [inspectorRelatedRecordModel.test.ts](inspectorRelatedRecordModel.test.ts) | Tests related-record seed construction and characterized workflow draft transitions. |
| [inspectorRelatedRecordModel.ts](inspectorRelatedRecordModel.ts) | Related-record form seeds and pure inspector workflow state transitions. |
| [useInspectorCreateRelatedWorkflow.test.tsx](useInspectorCreateRelatedWorkflow.test.tsx) | Tests neutral related-creation feedback and detached late-result suppression. |
| [useInspectorCreateRelatedWorkflow.ts](useInspectorCreateRelatedWorkflow.ts) | React controller for declared related-record authoring and presentation lifetime. |

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
| [workbookRecordHistoryModel.ts](workbookRecordHistoryModel.ts) | Inspector History state machine keyed by subject, read request, and operation identity. |
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
| [WorkbookInspectorDraftStore.test.ts](WorkbookInspectorDraftStore.test.ts) | Draft, capability, review, security and captured-revision regressions. |
| [useWorkbookInspectorEditDraft.test.tsx](useWorkbookInspectorEditDraft.test.tsx) | Frozen field/action binding through refresh, detachment, no-row and explicit return. |

[WorkbookExplicitPatchRecovery.tsx](WorkbookExplicitPatchRecovery.tsx) keeps
ordinary explicit attempt recovery reachable on its surface after inspector closure.

## Query windows

`useRetainedInspectorRow.ts` retains one authorized source independently of query
window membership. An absent query member does not establish deletion. Authority
changes retire the old source; newer committed evidence can update it off-window.
