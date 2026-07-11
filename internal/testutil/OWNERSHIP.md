# internal/testutil Ownership

`internal/testutil` is shared test infrastructure, not a product module. New
product-domain helpers should not be added here by default.

Composed Postgres, object-store, and application test-server startup is owned
by `internal/app/testsupport`. Owner-local helpers select dependencies,
privileged route mode, and semantic fixtures while reusing that composition
boundary.

Retained shared helper families:

| Package family | Owner posture |
| --- | --- |
| `authcookietest` | Shared auth-cookie assertion helper over platform auth cookie names. |
| `configtest`, `diagnosticstest`, `processtest` | Platform/runtime diagnostics and process harness support. |
| `auditassert`, `contractassert`, `securityassert` | Owner-neutral mutation/audit, protocol-contract, and recursive secret-safety assertions. |
| `fixtures`, `golden` | Shared fixture and golden facades for bootstrap, config, HTTP envelope, WebSocket, and adopted OTel evidence; new product-semantic data is not admitted here. |
| `httptestx` | Shared app test-server, request, and HTTP-envelope support; NT-02 and NT-03 remove implicit privilege and product-semantic assertions. |
| `pgschema`, `pgtest`, `s3test`, `suiteservices`, `testcontainersx` | Platform service, storage, and suite lifecycle infrastructure. |
| `wstest` | Generic WebSocket client/assertion infrastructure. |

Current owner-local support families:

| Owner | Current posture |
| --- | --- |
| Auth | Enterprise-auth, bootstrap, auth-flow, route-inventory, and store helpers under owner-semantic support packages. |
| Incidents | Incident scenario, route-inventory, mutation, fault, and store helpers under owner-semantic support packages. |
| Timeline and Collaboration | Timeline scenario/store assertions and generic incident WebSocket lifecycle support under owner-semantic support packages. |
| Workbook and Records | Workbook scenario/route helpers and record fixture/assertion/store support under owner-semantic support packages. |

Owner-local support packages are private implementation details. Phase labels
remain valid in test IDs and accounting, but they do not establish package or
symbol compatibility. New product-semantic helpers must use an owner and
responsibility name and must be registered in the canonical test-support
inventory introduced by NT-05.

The controlling migration artifact is
`docs/handoffs/test-util-module-refactor-tracker.md`.
