# app/api/

[Parent](../README.md) · [Source overview](../../README.md)

App-shell operation adapters and private protocol types for authentication, accounts, and administration.

These authored adapters call generated protocol operations through the shared
[browser services](../../services/README.md). App controllers own drafts, read
lifetimes, review, and recovery; clients own request/response boundaries.

## Files

| File | Responsibility |
| --- | --- |
| [administrativeAuditClient.ts](administrativeAuditClient.ts) | Generated-operation adapter for deployment audit pages and safe read failures. |
| [authAccountClient.ts](authAccountClient.ts) | Authentication, credential, account, and supporting-resource adapters over generated operations. |
| [deploymentUserClient.ts](deploymentUserClient.ts) | Deployment-user and enterprise-binding administration adapters over generated operations. |
| [incidentClient.ts](incidentClient.ts) | Incident directory and creation adapters over generated operations. |
| [incidentImportClient.ts](incidentImportClient.ts) | Incident bundle upload, job observation, and cancellation with captured attempts and typed outcomes. |
| [incidentLifecycleClient.ts](incidentLifecycleClient.ts) | Incident lifecycle reads and mutations through validated generated operations. |
| [incidentMembershipAuditClient.ts](incidentMembershipAuditClient.ts) | Incident-scoped membership audit page adapter with typed read outcomes. |
| [incidentMembershipManagementClient.ts](incidentMembershipManagementClient.ts) | Membership page reads and mutations with validated resources and semantic failures. |
| [incidentMetadataClient.ts](incidentMetadataClient.ts) | Validated incident metadata reads and sparse patch operations. |
| [incidentMetadataContracts.ts](incidentMetadataContracts.ts) | Private generated incident-resource type alias for metadata editing. |
| [incidentResourceClient.ts](incidentResourceClient.ts) | Validated incident summary read adapter. |
| [incidentResourceContracts.ts](incidentResourceContracts.ts) | Private generated incident-resource type alias for shared summary reads. |
| [publicHttpTypes.ts](publicHttpTypes.ts) | Public app-shell HTTP request and response type exports from the generated protocol facade. |

## Tests

| File | Responsibility |
| --- | --- |
| [administrativeAuditClient.test.ts](administrativeAuditClient.test.ts) | Tests exact deployment audit requests and safe, complete response-page validation. |
| [incidentDirectoryClient.test.ts](incidentDirectoryClient.test.ts) | Tests exact directory queries, opaque cursors, response validation, and cancellation. |
| [incidentImportClient.test.ts](incidentImportClient.test.ts) | Tests exact multipart replay, import admission validation, terminal jobs, and cancellation identity. |
| [incidentLifecycleClient.test.ts](incidentLifecycleClient.test.ts) | Tests captured lifecycle requests, correlated responses, and safe typed failure distinctions. |
| [incidentMembershipAuditClient.test.ts](incidentMembershipAuditClient.test.ts) | Tests exact incident audit requests, additive read vocabulary, redaction, and page correlation. |
| [incidentMembershipManagementClient.test.ts](incidentMembershipManagementClient.test.ts) | Tests membership page validation and exact versioned create, patch, and delete requests. |
| [incidentMetadataClient.test.ts](incidentMetadataClient.test.ts) | Tests sparse metadata patches, null preservation, no-op acknowledgements, and response correlation. |
| [shellHttpClients.routeBoundary.test.ts](shellHttpClients.routeBoundary.test.ts) | Route-boundary tests for app-shell client helpers. |
