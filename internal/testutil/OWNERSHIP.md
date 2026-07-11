# internal/testutil Ownership

`internal/testutil` is shared test infrastructure, not a product module. New
product-domain helpers should not be added here by default.

Retained shared helper families:

| Package family | Owner posture |
| --- | --- |
| `authcookietest` | Shared auth-cookie assertion helper over platform auth cookie names. |
| `configtest`, `diagnosticstest`, `processtest` | Platform/runtime diagnostics and process harness support. |
| `crosscutting` | Owner-neutral mutation attribution and secret-safe payload assertions. |
| `fixtures`, `golden` | Shared fixture and golden facades; product-specific fixture data is migration-bound by the tracker. |
| `httptestx` | Shared app test-server and HTTP helper; it stays out of platform packages because it imports `internal/app`. |
| `pgschema`, `pgtest`, `s3test`, `suiteservices`, `testcontainersx` | Platform service, storage, and suite lifecycle infrastructure. |
| `wstest` | Generic WebSocket client/assertion infrastructure. |

Migration-only helper families:

| Package family | Successor posture |
| --- | --- |
| `enterpriseauthtest`, `phase0test`, `phase1test`, `phase1storetest` | Move to auth owner test support. |
| `phase2test`, `phase2storetest` | Move to incident-scoped owner test support. |
| `phase3test`, `phase3storetest`, `timelinetest`, `timelinestoretest` | Move to Timeline owner test support. |
| `phase4test`, `phase4storetest`, `assertx`, Phase 4 fixture/golden data | Move or split into owner-local workbook/record test support. |
| `phase6test`, `incidentwstest` | Move to workbook/collaboration owner test support. |

The controlling migration artifact is
`docs/handoffs/test-util-module-refactor-tracker.md`.
