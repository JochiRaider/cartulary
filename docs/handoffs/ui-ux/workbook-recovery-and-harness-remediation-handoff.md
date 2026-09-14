# Workbook recovery and harness remediation handoff

Status: implementation and focused recovery validation complete; broader verification is in progress. This document is a review record, not a requirement owner or an executable input.

## Owner decisions and compatibility

Core 03 owns contextual Task Request/Decision retention and same-account authentication recovery, with Core 01 supplying the declared action matrix. The adopted Testing Harness NLSpec owns service leases, secure artifact production, output completion, and failure attribution. Clarifications were made in those owners before implementation. `docs/domain.md` was inspected; no vocabulary or owner-navigation change was necessary.

The repairs preserve public routes, mutation payloads, transaction identities, database schemas, and the existing account/incident/browser-runtime retention boundary. There is no persistent queue, source concurrency token on creation requests, insecure artifact mode, or compatibility switch for the defective behavior. The generated topology was refreshed with `make generate`; generated artifacts were not edited by hand. Human review of the owner-to-projection correspondence remains a review responsibility, not a test claim.

## Final gap ledger

| Gap | Implemented remediation and ownership | Long-term benefit and unresolved-risk boundary | Migration and validation |
| --- | --- | --- | --- |
| A: contextual creation depends on a mounted source | Core 03 clarification; `WorkbookMutationRuntime.coordinateSourceWrites` settles source writes and returns an accepted-version floor. Retained owners verify records through the incident-bound authoring reader, independently of filters and mounted callbacks. Explicit PATCH has a required source reader and identifies its own reservation. Other explicit writes and uncertain source outcomes participate in coordination. | Presentation lifetime no longer defines mutation authority. Missing materialization and refresh debt cannot prevent a valid retained creation; incomplete, stale, mismatched, or unavailable reads cannot authorize one. | Migrated contextual Task/Decision, coordination artifacts, Notes, related Evidence, Party links, and explicit PATCH. Removed `coordinateExplicitPatch`. Real-runtime tests cover detached Assessments, missing materialization, debt, stale/changed/unavailable sources, cancellation, and earlier writes. The reported Assessment browser row passes. |
| A: accepted refresh recovery races socket echoes | Accepted receipts are stored before projection effects. A late socket echo cannot restart a failed refresh; new affected views extend a completed refresh only when necessary. Unmounted views retain debt and reconcile on registration. | Acceptance survives failed reads. Explicit refresh recovery cannot accidentally become another creation or move the active Hosts surface. | No wire or persisted-state migration. The browser scenario creates exactly one Decision, preserves source identity and target values, recovers using reads, and checks focus/viewport behavior. |
| B: session loss is treated as incident access loss | Shared `WorkbookFailureLifecycle` distinguishes session authority, incident authority uncertainty, record availability, and operation rejection. Query/startup/reference/preview/candidate/action/reconciliation consumers request authoritative recovery. Only the authority owner or validated incident revocation exits the incident. Protected presentation is concealed separately from retained work. | Prevents lost edits and destructive interpretations of record-level or operation-level failures. Role downgrade and true membership loss retain their separate behavior. | Removed `workbookOperationFailureIsAccessLoss` and local observation access-loss classification. Renamed internal uncertain-authority callbacks; the actual incident-exit callback remains at its authoritative boundary. Current consumers were migrated together. |
| B: replay depends on mounted Timeline | `WorkbookTimelineMutationOwner` retains the queue driver, receipt settlement, source reader, and draft cleanup. The hook attaches presentation only. Session epochs and attachment identities fence obsolete observations and focus effects; retired lifetimes clear retained state. Incomplete source reads retain the queue and request recovery instead of establishing record absence. | FIFO replay remains executable on Hosts or across authentication-shell transitions. Late acceptance remains authoritative without exposing suspended presentation. | No persistent or cross-account storage. Integration uses the real runtime/coordinator and a detached Timeline, requires HTTP acceptance, materializes versions at dispatch, and retains stable transaction IDs. Session tests cover delayed old 401 responses around login and account replacement. |
| C: session fixture masks an upstream failure | One validated session-response rewrite helper is used by membership audit/management, metadata, administration, reference packs, import, and account editing fixtures. Only successful contract-valid sessions are rewritten. Upstream failures are forwarded with bounded operation/status/public-code/request-ID diagnostics. | Server failures remain diagnosable; malformed success or non-JSON responses cannot manufacture membership state. | Test-support API migration only. Fixture tests cover valid responses, malformed/error bodies, private-text exclusion, and transport failure. All three reported membership-audit browser rows pass. |
| D: concurrent suites can delete live services | The service janitor no longer infers abandonment from age. Reclamation requires a matching successful terminal cleanup summary for the owning run and suite. Each removal records exact container/run/suite/service attribution first. | Prevents another test run from destroying infrastructure used by a long-running browser suite. Unproven orphans remain under their lease-specific cleanup authority. | No backend session or readiness contract changed. Lifecycle regression rejects reclamation of another run even after the former ten-minute cutoff, while retaining confirmed cleanup and concurrent-removal handling. Historical causal disposition is below. |
| D: failed preparation and replacement lose their cause | Browser database preparation checks every required publication before sourcing it and explicitly propagates failure. Startup/reset steps and process diagnostics live under the exact browser session and reset attempt. Bounded redacted backend/frontend output is retained before private cleanup. | Failed prerequisites cannot produce misleading readiness; later attempts cannot overwrite an earlier attempt's log/summary pair. | Existing session/generation identities are reused; historical evidence is immutable. Failure-injection lifecycle smoke covers failed database preparation and missing environment/metadata publication, alongside existing replacement/cleanup cases. |
| D: explain-run misreads current summaries | Current run, target, and unit schemas are validated and read from their canonical locations. Summaries report unit counts and identity; accounting/progress use unit results. Unsupported schema versions fail explicitly. | Inspection cannot silently turn 109 units into zero work or undefined labels. | Supported historical formats retain their explicit handling. Canonical/unsupported-schema smoke tests pass; inspection of the reported broad root reports 109 units (47 passed, 4 failed, 58 skipped). |
| E: logs are insecure or incomplete at publication | Shared secure writers reject symlink/type/ownership violations and establish directory/file modes before bytes. Shell capture owns and joins redaction/output workers before publication. Detached services close inherited capture descriptors. Browser log/environment/staging producers use secure creation; replacement files are exclusive. | Secure artifact production is an invariant. Output completion no longer races summary publication or hangs on inherited descriptors held by a service. | No command changes or validator weakening. Execution smoke checks permissive umask, creation-time modes, delayed stderr, retained shell state, detached services, symlink rejection, and redaction failure. Existing timeout/cancellation and lifecycle coverage remains enabled. |

## Service failure reconstruction

The historical broad run `20260914T172458Z-p41830` used suite `491c0b8374480a86cfb97b26`. PostgreSQL started at `2026-09-14T17:25:45.144870759Z`; object storage started at `17:26:00.783Z`.

A concurrent run, `20260914T173625Z-p26170`, executed suite preflight from `17:37:20.697815148Z` through `17:37:22.204440728Z`. Its event at `17:37:22.188722587Z` reports four scanned containers and two removals. The broad run's session request `req-181` began at `17:37:20.892Z` and returned 500; Timeline query `req-178` began at `17:37:20.853Z` and also returned 500. Both failures occurred inside that cleanup window. The broad suite's two services had crossed the janitor's former ten-minute age threshold; the newer concurrent services had not.

Source inspection proves that the old janitor authorized deleting other runs' managed services solely by age. The timing, counts, and later connection refusal strongly support that defect as the service-disruption cause. Historical logs did not retain the deleted container IDs or the backend database exception, and Docker no longer retained events for that window. Consequently, the specific database exception behind the original session 500 cannot be recovered. The demonstrated destructive ownership defect is repaired and regression-tested; fixture hardening alone is not the causal closure. New cleanup attribution and retained process diagnostics remove that historical evidence gap for subsequent incidents.

The broad root is evidence for a **dirty `cd0be11dd321e48acbdcdbc507b6371099db48f9` snapshot**, source digest `sha256:ec1f71561d918a177025890a17a75244aa193aeb7a6619df455c744411b53c94`. It is not execution evidence for the current base commit. The supplied baseline and focused roots remain unchanged.

## Verification ledger

All commands run from the repository root through Make. Each run's `run-manifest.json` owns its exact source commit, dirty/clean state, and source digest. The implementation runs below started from `8e88d1f485d51f07b2cc93a73c05d41c056e296a` with the recorded uncommitted changes; they must not be relabeled as clean-commit runs.

| Command / scope | Result | Run root under `.cartulary/test-results/` |
| --- | --- | --- |
| `make test-slice OWNER=web.workbook` | Pass, 258/258 units | `20260914T192924Z-p10213` |
| `make test-slice OWNER=web.collaboration ROWS=web.collaboration.regression.workbookcollaborationcoordinator_suite_514be174a8` | Pass before additional incomplete-read regression; latest result recorded below | `20260914T192007Z-p36518` |
| `make test-slice OWNER=web.application ROWS=web.application.regression.app_session_lifecycle_d770146af2` | Pass, 2/2 units | `20260914T191022Z-p54997` |
| Assessment reproduction below | Pass, 11/11 units | `20260914T192009Z-p36775` |
| Collaboration reproduction below | Pass, 11/11 units | `20260914T193011Z-p29077` |
| Membership-audit reproduction below | Pass, 11/11 units | `20260914T192145Z-p69973` |
| `make test-slice OWNER=harness.browser ROWS=harness.browser.boundary_support.fixtures_suite_b570a8a829,harness.browser.unit.testservices_lifecycle_contract` | Pass, 3/3 units | `20260914T193347Z-p85945` |
| `make frontend-typecheck` | Pass, 2/2 units | `20260914T193348Z-p86680` |
| `make frontend-import-boundary-check` | Pass, 2/2 units; later broad result recorded below | `20260914T192016Z-p48872` |
| `make lint-shell` | Pass, 4/4 units | `20260914T192640Z-p97878` |
| `make lint-scripts` | Pass, 2/2 units | `20260914T192815Z-p4658` |
| `make lint-markdown` | Pass before this handoff was added; final result recorded below | `20260914T192816Z-p5700` |
| `make generate` | Pass | `20260914T193300Z-p77769` |
| `make agent-finalize` | Pass, 1/1 units | `20260914T193345Z-p85475` |

The three reported reproductions are:

```bash
make service-backed-test-slice OWNER=module.assessments ROWS=module.assessments.browser.contextual_decision_refresh
make service-backed-test-slice OWNER=module.collaboration ROWS=module.collaboration.browser.within_one_browser_runtime_queued_unsent_writes_3efbd48054
make service-backed-test-slice OWNER=module.incidents ROWS=module.incidents.browser.membership_audit_read_recovery,module.incidents.browser.membership_audit_access_recovery,module.incidents.browser.membership_audit_response_inspection
```

`RESULTS_DIR` was unset for finalization. Retained-run maintenance was therefore skipped; no qualifying successful full warm-check root was supplied. Required product rows were not counted as passing when dependency-skipped.

Development failures are retained rather than rewritten: the original Assessment navigation failure advanced to a refresh-echo race before passing; the replay test advanced to a Secure-cookie API-client mismatch before passing. Browser run `20260914T192404Z-p71173` was rejected because formatting changed frontend inputs during its build, and is not product evidence. Finalization root `20260914T193108Z-p73716` failed for stale generated topology; Make-owned generation corrected that. Early capture retries exposed inherited descriptors and were terminated before browser assertions; the detached-service smoke regression now covers that fault.

## Final validation and handoff

Broader browser, repository, generated-drift, and current integration results will be recorded here after completion. Intentional visual changes were not made; visual goldens are unchanged. Historical missing backend/container diagnostics remain the explicit limit on reconstructing the exact original 500.
