# workbook/features/coordination/

[Parent](../README.md) · [Source overview](../../../README.md)

Task lifecycle, Decision supersession, and contextual or coordination record-creation workflows.

Contextual Task/Decision creation and coordination action variants have distinct
draft and operation owners. Decision supersession coordinates both reviewed
participants; Task lifecycle uses explicit record-patch ownership in the
[mutation runtime](../../runtime/README.md).

## Contextual Task and Decision creation

| File | Responsibility |
| --- | --- |
| [contextualCreateAuthoring.test.tsx](contextualCreateAuthoring.test.tsx) | Tests declared contextual entry points and draft identity across detachment and source changes. |
| [ContextualCreateContext.ts](ContextualCreateContext.ts) | React context exposing contextual Task/Decision creation ownership. |
| [contextualCreateDiscovery.test.ts](contextualCreateDiscovery.test.ts) | Tests contextual route/capability validation and paged target-reference discovery. |
| [ContextualCreateForm.tsx](ContextualCreateForm.tsx) | Contextual Task/Decision authoring form with source context and reviewed submission. |
| [contextualCreateModel.ts](contextualCreateModel.ts) | Contextual Task/Decision feature identities, source-bound drafts, and reference requirements. |
| [contextualCreateOperation.ts](contextualCreateOperation.ts) | Contextual creation reviews, immutable attempts, receipts, outcomes, and owner entries. |
| [contextualCreateRecovery.test.tsx](contextualCreateRecovery.test.tsx) | Tests synchronous reservation, prior-save coordination, timeout recovery, and late receipt acceptance. |
| [ContextualCreateRecovery.tsx](ContextualCreateRecovery.tsx) | Recovery presentation for retained contextual Task/Decision creation attempts. |
| [ContextualReferenceControl.tsx](ContextualReferenceControl.tsx) | Paged reference selection for contextual Task/Decision creation. |
| [useContextualCreateAttachment.ts](useContextualCreateAttachment.ts) | Attaches contextual Task/Decision presentation to retained source-bound creation state. |
| [WorkbookContextualTaskDecisionCreateOwner.ts](WorkbookContextualTaskDecisionCreateOwner.ts) | Contextual Task/Decision draft, reviewed request, replay, and reconciliation ownership. |

## Coordination creation and bindings

| File | Responsibility |
| --- | --- |
| [coordinationCreateAuthoring.test.tsx](coordinationCreateAuthoring.test.tsx) | Tests coordination action variants, target minima, and retained authoring across navigation. |
| [CoordinationCreateContext.ts](CoordinationCreateContext.ts) | React context exposing coordination creation ownership. |
| [CoordinationCreateForm.tsx](CoordinationCreateForm.tsx) | Coordination creation form for source-specific action variants. |
| [coordinationCreateModel.ts](coordinationCreateModel.ts) | Coordination source/action variants, draft fields, validation, and request preparation. |
| [coordinationCreateOperation.ts](coordinationCreateOperation.ts) | Coordination creation reviews, captured attempts, receipts, and semantic outcomes. |
| [coordinationCreateRecovery.test.tsx](coordinationCreateRecovery.test.tsx) | Tests exact coordination request capture, same-frame admission, and replay after response loss. |
| [CoordinationCreateRecovery.tsx](CoordinationCreateRecovery.tsx) | Recovery presentation for retained coordination creation attempts. |
| [CoordinationWorkflowBindings.tsx](CoordinationWorkflowBindings.tsx) | Coordination-owned task lifecycle and decision supersession presentation over semantic commands. |
| [useCoordinationCreateAttachment.ts](useCoordinationCreateAttachment.ts) | Attaches coordination creation presentation to its source-bound owner state. |
| [useCoordinationWorkflowController.test.tsx](useCoordinationWorkflowController.test.tsx) | Tests saved Task initialization, independent Decision state, and late-outcome invalidation. |
| [useCoordinationWorkflowController.ts](useCoordinationWorkflowController.ts) | React coordination of Task lifecycle actions and Decision supersession capabilities. |
| [WorkbookCoordinationCreateOwner.ts](WorkbookCoordinationCreateOwner.ts) | Coordination creation draft, atomic request, replay, and reconciliation ownership. |

## Decision supersession

| File | Responsibility |
| --- | --- |
| [decisionSupersession.characterization.test.tsx](decisionSupersession.characterization.test.tsx) | Tests History-panel placement, complete supersession acknowledgements, and declared role admission. |
| [DecisionSupersessionContext.ts](DecisionSupersessionContext.ts) | React context exposing Decision supersession ownership. |
| [DecisionSupersessionEditor.test.tsx](DecisionSupersessionEditor.test.tsx) | Tests replacement review, candidate paging failures, review invalidation, and retained drafts. |
| [DecisionSupersessionEditor.tsx](DecisionSupersessionEditor.tsx) | Decision replacement selection, reason authoring, and explicit supersession review. |
| [decisionSupersessionModel.ts](decisionSupersessionModel.ts) | Decision authority, reviewed participant versions, eligibility, and supersession review models. |
| [decisionSupersessionOperation.ts](decisionSupersessionOperation.ts) | Decision write coordination, captured supersession attempts, receipts, and candidate read ports. |
| [decisionSupersessionReconciliation.test.tsx](decisionSupersessionReconciliation.test.tsx) | Tests both-participant refresh and monotonic query/live reconciliation of accepted versions. |
| [decisionSupersessionRuntime.test.ts](decisionSupersessionRuntime.test.ts) | Tests supersession coordination with queued writes and accepted History versions. |
| [reconcileDecisionReceipt.ts](reconcileDecisionReceipt.ts) | Reconciles supersession receipts against both Decision records and their histories. |
| [useDecisionCandidates.ts](useDecisionCandidates.ts) | Paged eligible Decision candidate discovery with explicit read-state feedback. |
| [WorkbookDecisionSupersessionOwner.test.ts](WorkbookDecisionSupersessionOwner.test.ts) | Tests participant reservation, write coordination, exact replay, and current-authority admission. |
| [WorkbookDecisionSupersessionOwner.ts](WorkbookDecisionSupersessionOwner.ts) | Decision supersession admission, participant coordination, retained attempts, and reconciliation. |
| [WorkbookDecisionSupersessionRecovery.tsx](WorkbookDecisionSupersessionRecovery.tsx) | Decision supersession replay, receipt completion, and reconciliation recovery controls. |

## Task lifecycle

| File | Responsibility |
| --- | --- |
| [taskLifecycleModel.ts](taskLifecycleModel.ts) | Task lifecycle status, guarded fields, patch changes, and transition preparation. |
| [TaskPatchRecovery.tsx](TaskPatchRecovery.tsx) | Recovery presentation for retained Task lifecycle patches. |

## Ordinary creation seam

| File | Responsibility |
| --- | --- |
| [coordinationOrdinaryCreate.ts](coordinationOrdinaryCreate.ts) | Ordinary coordination preparation, defaults and reference restrictions without contextual source seeding. |
| [initialCoordinationCreateRules.ts](initialCoordinationCreateRules.ts) | Shared initial Task/Decision lifecycle guards for ordinary and contextual preparation. |

Task explicit patches contribute their guard dependencies and validation through
[taskExplicitPatchContribution.ts](taskExplicitPatchContribution.ts). The Task
draft store remains here; the neutral runtime does not own Task or Party forms.
