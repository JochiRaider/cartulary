# workbook/features/assessments/

[Parent](../README.md) · [Source overview](../../../README.md)

Assessment subject/support discovery, append authoring, inspector composition, and recovery.

Append ownership retains reviewed drafts and receipts independently of inspector
attachment. Candidate reads use a semantic port; generated operation transport
lives in the [workbook adapters](../../adapters/README.md).

## Files

| File | Responsibility |
| --- | --- |
| [AssessmentAppendRecovery.tsx](AssessmentAppendRecovery.tsx) | Recovery presentation for retained Assessment append operations. |
| [assessmentCandidatePort.ts](assessmentCandidatePort.ts) | Semantic capability for paged Assessment subject and support discovery. |
| [AssessmentDiscovery.tsx](AssessmentDiscovery.tsx) | Assessment subject and support pickers over paged authorized candidates. |
| [assessmentOperation.ts](assessmentOperation.ts) | Assessment drafts, reviews, captured append attempts, receipts, and transport outcomes. |
| [AssessmentWorkbookInspector.tsx](AssessmentWorkbookInspector.tsx) | Assessment inspector facade over feature-owned composition and presentation. |
| [useAssessmentCreationController.ts](useAssessmentCreationController.ts) | React controller for Assessment creation drafts, support selection, and append commands. |
| [useAssessmentWorkbookInspectorComposition.tsx](useAssessmentWorkbookInspectorComposition.tsx) | Assembles Assessment inspector sections, subject state, and feature workflows. |
| [WorkbookAssessmentAuthoringOwner.ts](WorkbookAssessmentAuthoringOwner.ts) | Assessment append draft, reviewed attempt, acknowledgement, and recovery ownership. |

## Tests

| File | Responsibility |
| --- | --- |
| [assessmentDiscovery.test.tsx](assessmentDiscovery.test.tsx) | Tests authorized candidate queries, opaque paging, and failed-page retry without candidate loss. |
| [useAssessmentCreationController.test.tsx](useAssessmentCreationController.test.tsx) | Tests accepted creation feedback and retained append drafts after rejection. |

AssessmentDiscovery uses one-page Workbook observation and shared explicit query
controls. Host/Identity subjects change only on deliberate selection. Support
selection stages at most 64 additions from Timeline and retains off-page identities;
existing non-Timeline support remains readable. Apply readiness is independent of
page exhaustion, and stale presentation is concealed through authority/freshness
signals. Assessment append attempts and refresh debt remain parent-owned.
