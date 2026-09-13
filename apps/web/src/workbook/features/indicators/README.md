# workbook/features/indicators/

[Parent](../README.md) · [Source overview](../../../README.md)

Indicator creation and lifecycle, Observation capture/resolution, and operation reconciliation.

Creation, lifecycle, and Observation operations each retain drafts, captured
attempts, receipts, and reconciliation independently of presentation attachment.
Transport lives in [workbook adapters](../../adapters/README.md); shared
[History](../../history/README.md) and query owners supply current materialization.

## Canonical Indicator creation

| File | Responsibility |
| --- | --- |
| [IndicatorCanonicalAuthoring.tsx](IndicatorCanonicalAuthoring.tsx) | Canonical Indicator value-kind and value authoring controls. |
| [indicatorCreateAuthoring.test.tsx](indicatorCreateAuthoring.test.tsx) | Tests explicit canonical value kinds and draft preservation when Indicator type changes. |
| [IndicatorCreateContext.ts](IndicatorCreateContext.ts) | React context exposing Indicator creation ownership. |
| [IndicatorCreateDraftStore.ts](IndicatorCreateDraftStore.ts) | Retains Indicator creation drafts independently of presentation lifetime. |
| [IndicatorCreateFromObservation.test.tsx](IndicatorCreateFromObservation.test.tsx) | Tests accepted canonical creation, explicit Observation resolution, and independent replay. |
| [IndicatorCreateFromObservation.tsx](IndicatorCreateFromObservation.tsx) | Workflow for creating a canonical Indicator from an Observation and explicitly resolving it. |
| [indicatorCreateModel.ts](indicatorCreateModel.ts) | Indicator creation values, drafts, validation, availability, and Observation-derived seeds. |
| [indicatorCreateOperation.ts](indicatorCreateOperation.ts) | Indicator creation attempts, receipts, outcomes, snapshots, and authority bindings. |
| [IndicatorCreateOperationStatus.tsx](IndicatorCreateOperationStatus.tsx) | Indicator creation operation status and recovery feedback. |
| [indicatorCreateReconciliation.test.ts](indicatorCreateReconciliation.test.ts) | Tests paged target refresh, newer History evidence, and retained receipts on unavailable reads. |
| [reconcileIndicatorCreateReceipt.ts](reconcileIndicatorCreateReceipt.ts) | Reconciles accepted Indicator creation with current query and History materialization. |
| [WorkbookIndicatorCreateOwner.test.ts](WorkbookIndicatorCreateOwner.test.ts) | Tests creation admission, immutable replay, acknowledgement ordering, and refresh-only recovery. |
| [WorkbookIndicatorCreateOwner.ts](WorkbookIndicatorCreateOwner.ts) | Canonical Indicator creation admission, captured attempt, acknowledgement, and reconciliation ownership. |
| [WorkbookIndicatorCreateRecovery.tsx](WorkbookIndicatorCreateRecovery.tsx) | Retained Indicator creation replay and reconciliation recovery controls. |

## Observations and inspector workflows

| File | Responsibility |
| --- | --- |
| [indicatorInspectorHandlers.ts](indicatorInspectorHandlers.ts) | Resolves declared Indicator inspector actions to feature-owned handlers. |
| [IndicatorInspectorWorkflow.test.tsx](IndicatorInspectorWorkflow.test.tsx) | Tests accessible Observation loading, retry, empty states, and retained paged results. |
| [IndicatorInspectorWorkflow.tsx](IndicatorInspectorWorkflow.tsx) | Inspector composition for Observation browsing, capture, and resolution workflows. |
| [indicatorObservations.characterization.test.tsx](indicatorObservations.characterization.test.tsx) | Tests Observation capture exposes committed source text for accessible selection. |
| [observationAuthoring.test.tsx](observationAuthoring.test.tsx) | Tests exact source previews, optional-field omission, paged targets, and isolated transition drafts. |
| [ObservationCaptureEditor.tsx](ObservationCaptureEditor.tsx) | Observation source-text selection and capture authoring controls. |
| [ObservationCollection.test.ts](ObservationCollection.test.ts) | Tests identity-based page deduplication, failed-cursor retry, progress checks, and stale-read fencing. |
| [ObservationCollection.ts](ObservationCollection.ts) | Paged Observation collection state with retained results and request-generation fencing. |
| [ObservationContext.ts](ObservationContext.ts) | React context exposing Observation operation ownership. |
| [ObservationDetails.tsx](ObservationDetails.tsx) | Observation provenance, source span, and target detail presentation. |
| [ObservationDraftStore.ts](ObservationDraftStore.ts) | Retains Observation drafts by source field and operation subject. |
| [observationModel.test.ts](observationModel.test.ts) | Tests Unicode source spans, capability-based field discovery, and identity-isolated drafts. |
| [observationModel.ts](observationModel.ts) | Observation source, selection, target, and draft models with source-span validation. |
| [observationOperation.ts](observationOperation.ts) | Observation authority, subjects, intents, immutable attempts, pages, and failure classification. |
| [ObservationOperationStatus.tsx](ObservationOperationStatus.tsx) | Observation operation status and recovery feedback. |
| [ObservationPagingFeedback.tsx](ObservationPagingFeedback.tsx) | Observation collection loading, continuation, empty-state, and retry feedback. |
| [observationPresentation.test.tsx](observationPresentation.test.tsx) | Tests stale source-read fencing, protected-state concealment, retained uncertainty, and subject focus. |
| [observationReconciliation.test.tsx](observationReconciliation.test.tsx) | Tests source/Indicator reconciliation across pages and rejection of regressing HTTP projections. |
| [observationStyles.ts](observationStyles.ts) | Shared Observation form, source-text, and feedback styles. |
| [ObservationTargetPicker.tsx](ObservationTargetPicker.tsx) | Paged Indicator target selection for Observation transitions. |
| [reconcileObservationReceipt.ts](reconcileObservationReceipt.ts) | Reconciles Observation receipts with source, previous target, and new target materialization. |
| [useObservationTargetNames.ts](useObservationTargetNames.ts) | Loads display names for Observation targets without changing their stable identities. |
| [WorkbookObservationOwner.test.ts](WorkbookObservationOwner.test.ts) | Tests Observation admission, shell-recovery retention, exact replay, and account retirement. |
| [WorkbookObservationOwner.ts](WorkbookObservationOwner.ts) | Observation capture and transition ownership with source coordination, replay, and reconciliation. |
| [WorkbookObservationRecovery.tsx](WorkbookObservationRecovery.tsx) | Retained Observation operation replay and reconciliation recovery controls. |

## Indicator lifecycle

| File | Responsibility |
| --- | --- |
| [indicatorLifecycle.characterization.test.tsx](indicatorLifecycle.characterization.test.tsx) | Tests Indicator lifecycle content is reachable in the canonical History panel. |
| [IndicatorLifecycleContext.ts](IndicatorLifecycleContext.ts) | React context exposing Indicator lifecycle operation ownership. |
| [IndicatorLifecycleDraftStore.ts](IndicatorLifecycleDraftStore.ts) | Retains raw Indicator lifecycle drafts by subject identity. |
| [indicatorLifecycleModel.test.ts](indicatorLifecycleModel.test.ts) | Tests UTC interval validation, support identity/order, nullable fields, and isolated subject drafts. |
| [indicatorLifecycleModel.ts](indicatorLifecycleModel.ts) | Indicator lifecycle drafts, interval validation, support references, and reviewed values. |
| [indicatorLifecycleOperation.ts](indicatorLifecycleOperation.ts) | Indicator lifecycle authority, pages, attempts, outcomes, and operation scope. |
| [IndicatorLifecycleOperationStatus.tsx](IndicatorLifecycleOperationStatus.tsx) | Indicator lifecycle operation status and recovery feedback. |
| [indicatorLifecyclePaging.ts](indicatorLifecyclePaging.ts) | Paged Indicator lifecycle and support read state with navigation and recovery. |
| [indicatorLifecycleReconciliation.test.tsx](indicatorLifecycleReconciliation.test.tsx) | Tests affected-identity refresh and monotonic reconciliation against HTTP, socket, and History evidence. |
| [indicatorLifecycleRuntime.test.ts](indicatorLifecycleRuntime.test.ts) | Tests same-record write coordination, renewed version review, and recovery across shell detachment. |
| [indicatorLifecycleStyles.ts](indicatorLifecycleStyles.ts) | Shared Indicator lifecycle form and feedback styles. |
| [IndicatorLifecycleSupportPicker.tsx](IndicatorLifecycleSupportPicker.tsx) | Paged lifecycle support selection with explicit continuation and read feedback. |
| [IndicatorLifecycleWorkflow.test.tsx](IndicatorLifecycleWorkflow.test.tsx) | Tests interval details, field-local UTC errors, paged support selection, and retained drafts. |
| [IndicatorLifecycleWorkflow.tsx](IndicatorLifecycleWorkflow.tsx) | Indicator lifecycle interval details, authoring, support selection, and reviewed actions. |
| [reconcileIndicatorLifecycleReceipt.ts](reconcileIndicatorLifecycleReceipt.ts) | Reconciles accepted lifecycle changes across affected Indicator identities and History. |
| [WorkbookIndicatorLifecycleOwner.test.ts](WorkbookIndicatorLifecycleOwner.test.ts) | Tests lifecycle reservation, reviewed versions, exact replay, and authority-bound dispatch. |
| [WorkbookIndicatorLifecycleOwner.ts](WorkbookIndicatorLifecycleOwner.ts) | Indicator lifecycle operation admission, retained attempts, acknowledgements, and reconciliation. |
| [WorkbookIndicatorLifecycleRecovery.tsx](WorkbookIndicatorLifecycleRecovery.tsx) | Retained Indicator lifecycle replay and reconciliation recovery controls. |
