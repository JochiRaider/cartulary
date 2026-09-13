# workbook/policies/

[Parent](../README.md) · [Source overview](../../README.md)

Contract-backed surface policy declarations, reference requirements, owner bindings, and application shortcuts.

Policies declare create defaults, advisory minima, collection semantics,
reference requirements, refresh consequences, errors, and owner UI bindings.
They remain pure and perform neither transport nor authorization.

## Files

| File | Responsibility |
| --- | --- |
| [artifactSurfacePolicies.ts](artifactSurfacePolicies.ts) | Notes, Findings, Investigative Queries, and Forensic Keywords policy owner. |
| [assessmentSurfacePolicies.ts](assessmentSurfacePolicies.ts) | Assessments policy owner. |
| [captureTimelineSurfacePolicies.ts](captureTimelineSurfacePolicies.ts) | Timeline surface policy owner. |
| [coordinationSurfacePolicies.ts](coordinationSurfacePolicies.ts) | Parties, Tasks, Decisions, Communications, Handoff, Status Review, and Lessons policy owner. |
| [entitiesObservationsSurfacePolicies.ts](entitiesObservationsSurfacePolicies.ts) | Hosts, Identities, and Indicators policy owner. |
| [evidenceSurfacePolicies.ts](evidenceSurfacePolicies.ts) | Evidence policy owner. |
| [workbookApplicationShortcuts.ts](workbookApplicationShortcuts.ts) | Pure capability-gated Workbook application shortcuts for quick link, Evidence preview, History, and inspector close. |
| [workbookSurfacePolicy.ts](workbookSurfacePolicy.ts) | Pure `WorkbookSurfacePolicy` types, immutable defaults, and stable reference declarations. |

## Tests

| File | Responsibility |
| --- | --- |
| [workbookApplicationShortcuts.test.ts](workbookApplicationShortcuts.test.ts) | Exhaustive application-shortcut admission and event-consumption tests. |
| [workbookSurfaceOwnershipPolicy.test.ts](workbookSurfaceOwnershipPolicy.test.ts) | Static policy-purity, owner-completeness, and common-surface boundary checks. |
