# testing/

[Source overview / parent](../README.md)

Reusable frontend fixtures, Vitest setup, and source/import/selector policy tests.

Runtime application code must not import this directory. Colocated tests remain
with the behavior they cover; these fixtures support reuse across owners.
Source accounting and policy checks consume machine-readable inputs, not READMEs.

## App and administrative fixtures

| File | Responsibility |
| --- | --- |
| [administrativeAuditTestSupport.ts](administrativeAuditTestSupport.ts) | Deployment audit actor, event, page, and response-envelope fixtures. |
| [appShellTestSupport.ts](appShellTestSupport.ts) | Shared app-shell test helpers and fixtures. |
| [incidentImportTestSupport.ts](incidentImportTestSupport.ts) | Incident import job identities, resources, envelopes, and captured-upload helpers. |
| [incidentMembershipAuditTestSupport.tsx](incidentMembershipAuditTestSupport.tsx) | Incident membership audit composition fixture for presentation and lifecycle tests. |
| [incidentMembershipManagementTestSupport.ts](incidentMembershipManagementTestSupport.ts) | Membership authorities, resources, pages, envelopes, and deferred-response helpers. |
| [incidentMetadataTestSupport.ts](incidentMetadataTestSupport.ts) | Incident metadata authorities, resources, envelopes, and deferred-response helpers. |
| [referencePackTestSupport.ts](referencePackTestSupport.ts) | Reference Pack actors, versions, catalog pages, and job fixtures. |
| [incidentMetadataSurfaceTestSupport.tsx](incidentMetadataSurfaceTestSupport.tsx) | Incident metadata presentation composition; tests supply fixture authority. |
| [incidentLifecycleSurfaceTestSupport.tsx](incidentLifecycleSurfaceTestSupport.tsx) | Incident lifecycle presentation composition; tests supply fixture authority. |
| [incidentMembershipManagementSurfaceTestSupport.tsx](incidentMembershipManagementSurfaceTestSupport.tsx) | Incident membership management presentation composition; tests supply fixture authority. |

## Feature workflow fixtures

| File | Responsibility |
| --- | --- |
| [decisionSupersessionTestSupport.ts](decisionSupersessionTestSupport.ts) | Decision supersession authority, row, review, request, and receipt fixtures. |
| [entityMergeTestSupport.ts](entityMergeTestSupport.ts) | Entity merge survivor/loser identities, authority, rows, reviews, and receipts. |
| [indicatorCreateTestSupport.ts](indicatorCreateTestSupport.ts) | Canonical Indicator creation contracts, values, responses, and composed fixtures. |
| [indicatorLifecycleTestSupport.ts](indicatorLifecycleTestSupport.ts) | Indicator lifecycle authorities, intervals, rows, and receipt fixtures. |
| [observationTestSupport.ts](observationTestSupport.ts) | Observation source/target identities, authority, receipts, and reader fixtures. |

## Workbook and Timeline fixtures

| File | Responsibility |
| --- | --- |
| [extensionAvailabilityTestSupport.ts](extensionAvailabilityTestSupport.ts) | Deterministic ready extension-availability controller fixture for workbook and feature tests. |
| [fetchMockTestSupport.ts](fetchMockTestSupport.ts) | Fetch mock helpers for unit and integration-style frontend tests. |
| [taskWorkbookTestSupport.ts](taskWorkbookTestSupport.ts) | Task lifecycle authority, record, intent, and receipt fixtures. |
| [timelineCaptureActionTestSupport.ts](timelineCaptureActionTestSupport.ts) | Reviewed Timeline capture-action fixture construction. |
| [timelineMentionTestSupport.ts](timelineMentionTestSupport.ts) | Timeline mention review, receipt, workbook row, inspector, and entity-creation fixtures. |
| [timelineWorkbookRenderTestSupport.tsx](timelineWorkbookRenderTestSupport.tsx) | Shared Timeline render helpers used by component characterization tests. |
| [TimelineWorkbookRuntimeFixture.tsx](TimelineWorkbookRuntimeFixture.tsx) | Production-composed Timeline runtime fixture with configurable shell, layout, query, entity and interaction inputs, borrowing the production recovery-focus owner. |
| [timelineWorkbookTestSupport.test.tsx](timelineWorkbookTestSupport.test.tsx) | Tests for Timeline workbook test-support helpers. |
| [timelineWorkbookTestSupport.ts](timelineWorkbookTestSupport.ts) | Shared Timeline workbook fixture helpers, route mocks, and row builders for tests. |
| [workbookAuthorizationTestSupport.ts](workbookAuthorizationTestSupport.ts) | Deterministic workbook authorization-recovery port fixture. |
| [workbookHistoryTestSupport.ts](workbookHistoryTestSupport.ts) | Complete semantic History events and explicitly accepted presentation pages shared by tests; never a production fallback. |
| [workbookImportTestSupport.ts](workbookImportTestSupport.ts) | Workbook import scopes, mappings, units, approvals, previews, and session fixtures. |
| [workbookInspectorTestSupport.test.tsx](workbookInspectorTestSupport.test.tsx) | Tests for delayed entity-inspector readiness and safe subject diagnostics. |
| [workbookInspectorTestSupport.ts](workbookInspectorTestSupport.ts) | Entity-inspector readiness waits and diagnostics keyed by stable surface, record, and row-version identity. |
| [workbookPreferenceTestSupport.ts](workbookPreferenceTestSupport.ts) | Workbook preference authorities, sheet targets, and home/default resource fixtures. |
| [workbookSchemaTestSupport.ts](workbookSchemaTestSupport.ts) | Public discovery fixtures projected from typed contracts. |
| [workbookSavedViewTestSupport.ts](workbookSavedViewTestSupport.ts) | Saved-view authorities, resources, controllers, and React application bindings for tests. |

Isolated query consumers use the required browsing provider through
`workbookQueryTestSupport` render helpers. `TimelineWorkbookRuntimeFixture`
composes that provider directly. Selection fixtures admit real query results
before exercising membership-dependent commands; they do not synthesize a
missing-provider fallback.

## Setup and architecture policies

Timeline scalar fixtures exercise current control values on Enter/Tab, matching
the production Grid Adapter. Its test-support renderer must preserve that contract
even when DOM input and retained React state temporarily differ.

| File | Responsibility |
| --- | --- |
| [selectorContractPolicy.test.ts](selectorContractPolicy.test.ts) | Selector ownership policy test that guards raw `data-testid` literals and shared selector facade usage. |
| [sourceOwnershipPolicy.test.ts](sourceOwnershipPolicy.test.ts) | Exact non-Markdown frontend source-ownership parity and manifest-shape policy. |
| [testSetup.dom.ts](testSetup.dom.ts) | DOM-specific Vitest setup. |
| [testSetup.ts](testSetup.ts) | Common Vitest setup for frontend tests. |
| [transportBoundaryPolicy.test.ts](transportBoundaryPolicy.test.ts) | Static same-origin transport policy; raw fetch is limited to the shared transport and server-issued Evidence upload target. |
