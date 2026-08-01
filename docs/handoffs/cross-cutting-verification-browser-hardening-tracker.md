# Cross-Cutting Verification and Browser Hardening Tracker

## 1. Scope and authority

This tracker records the hardening work that followed the completed
Tasks/Decisions refactor. The historical failure record remains unchanged in
[Section 13 of the completed refactor tracker](./tasksdecisions-module-refactor-tracker.md#13-production-readiness-cleanup-iteration).
This tracker does not reopen that refactor and does not own product behavior.

The adopted Testing Harness NLSpec owns the harness requirements changed here.
Its checked-in JSON inputs and generated outputs are machine projections.
Core 00 through Core 04 and `docs/domain.md` are unchanged. No runtime command,
test, generator, conformance check, or release check reads this tracker.

The work has no public API, OpenAPI, database, persisted-data, route, or view-ID
migration. It changes only private harness contracts, test evidence, and
browser-support mechanics.

## 2. Adopted contract decisions

| Contract | Final decision | Compatibility posture |
| --- | --- | --- |
| Fixture profiles | Every profile declares `postgres_policy`; focused and broad Go planning share one assignment facade. | All profiles migrated atomically; no inference fallback. |
| Catalog closure | Repository-wide owner, family, row, and selector totals are derived diagnostics. | Removed redundant authored count assertions; relational closure remains fail-closed. |
| Go duration baselines | `cartulary.go_test_duration_baselines.v5` owns explicit test, package, and command defaults. | v4 is unsupported after migration. Missing non-raw subjects are reported as defaulted. |
| Finalization | Existing action IDs and report schema remain; closure precedes generation and retained refresh precedes retained generation. | Action order and cache contract versions changed; consumers retain stable IDs. |
| Backend boundaries | `cartulary.backend_module_boundaries.v2` requires `match_kind=exact|subtree`. | v1 and implicit-prefix rules are unsupported. Existing rules were migrated explicitly. |
| Worker-admin state | `cartulary.playwright_worker_admin_manifest.v1` is strict, closed, ephemeral, and atomically published. | No legacy ephemeral reader or persistent migration. |
| Service-backed schedule sources | `cartulary.service_backed_schedule_sources.v2` carries authored readiness attribution. | v1 and target-name inference are unsupported. |
| Timing evidence | Authored execution topology owns readiness role, class, threshold, and rationale; diagnostics require pressure-summary v4. | Thresholds and balance policy are unchanged; missing or obsolete summaries fail as artifacts. |

## 3. Workstream status

| Workstream | Status | Implemented closure |
| --- | --- | --- |
| Specification ownership | Complete | Testing Harness requirements now cover fixture parity, derived catalog totals, duration defaults, finalizer order, explicit boundary matching, atomic secret state, stable record polling, and topology-owned timing attribution. |
| Fixture parity | Complete | Fixture policy is read from authored profiles, incorporated in catalog semantics, and projected through the backend target-planning owner facade into owner-slice child environments. |
| Catalog maintenance | Complete | Hard-coded global totals were removed while reference, owner, identity, selector, partition, and digest invariants remain. Synthetic catalog growth/closure tests remain contract evidence. |
| Duration bootstrap | Complete | v5 defaults are used by shard planning and reported by coverage; raw aggregate coverage remains strict; eligible retained refresh replaces observed defaults and retained drift rejects missing observed subjects. |
| Finalizer sequencing | Complete | Both retained and non-retained orders, explicit skips, fail-fast boundaries, catalog validation, and cache version changes are contract-tested. |
| Boundary semantics | Complete | All forbidden-import rules declare match kind; checker diagnostics name rule, match kind, importer, and candidate import; source-table access remains file-exact. |
| Browser secret state | Complete | Worker-admin JSON and shared TOTP state use one exclusive 0600 temporary-file, flush, close, and atomic-rename helper with cleanup on failure. |
| Browser row identity | Complete | Collaboration polling reacquires the `record_id` row and row-local editor/display each time and reports mounted identities on failure. |
| Idempotency precedence | Complete | Adapter-owner regression evidence covers changed-hash conflict before malformed typed replay decoding, exact-hash decode failure, and absence of source/revision/projection effects. |
| Timing attribution | Complete | Hard-coded target-name maps were removed; check and service-backed generated manifests carry the same authored metadata; pressure-summary v4 is mandatory diagnostic input. |

## 4. Validation ledger

The final handoff records only Make-owned verification. Run roots are retained
under `.cartulary/test-results/` and are diagnostic evidence, not requirements.

| Validation | Result | Evidence |
| --- | --- | --- |
| Formatting | Pass | `make format`; `20260801T231952Z-p2417911` |
| JSON shape | Pass | `make json-shape-check`; `20260801T230210Z-p2171440` |
| Catalog closure | Pass | `make test-catalog-check`; `20260801T230237Z-p2174836` |
| Boundary v2 | Pass | `make backend-module-boundary-check`; `20260801T230210Z-p2171632` |
| Harness contracts | Pass | `make harness-contract`; `20260801T232011Z-p2423403` |
| Generation drift | Pass | `make generate-drift`; `20260801T230236Z-p2174267` |
| Generated artifact policy | Pass | `make generated-artifact-policy-check`; `20260801T230236Z-p2174284` |
| Atomic private-state row | Pass | `make test-slice OWNER=harness.browser ROWS=...6e8cd66ca6`; `20260801T224507Z-p2032943` |
| Record-stable polling row | Pass | `make test-slice OWNER=harness.browser ROWS=...604f761c28`; `20260801T225910Z-p2159976` |
| Harness browser owner | Pass, 16/16 rows | `make test-slice OWNER=harness.browser`; `20260801T230210Z-p2171511` |
| Tasks/Decisions idempotency row | Pass | `make service-backed-test-slice OWNER=module.tasksdecisions ROWS=...90a3399bf6`; `20260801T224552Z-p2038313` |
| Tasks/Decisions conflict row | Pass with `postgres_template_clone` | `make service-backed-test-slice OWNER=module.tasksdecisions ROWS=...10e9976746`; `20260801T224607Z-p2039707` |
| Tasks/Decisions owner slices | Pass | non-service `20260801T224739Z-p2066930`; service-backed `20260801T224629Z-p2041221` |
| Collaboration exact browser row, retry-free | Pass three independent runs | `20260801T225328Z-p2097191`, `20260801T225405Z-p2116039`, `20260801T225442Z-p2134524` |
| Known timing rejection | Expected reject | Prior root `20260801T215105Z-p1883892` still rejects `build-server` at 16,209 ms over 15,000 ms and `check-service-backed` at 174,479 ms over 155,000 ms; diagnostic run `20260801T230054Z-p2162284` |
| Webserver-backed browser suite | Pass | `make browser-e2e-webserver-backed`; `20260801T230322Z-p2180699` |
| Repository test aggregate | Pass, 1,021 tests | `make test`; `20260801T230835Z-p2211558` |
| Mandatory non-retained finalizer | Pass | `make agent-finalize`; `20260801T232055Z-p2425008`; schema/catalog, coverage, and generation executed in order and retained-only actions were explicitly skipped |
| Full check | Pass, 194/194 work units and 872 tests | `make check`; `20260801T232123Z-p2431771` |
| Retained maintenance | Inapplicable; not run | The successful final check exceeded unchanged eligibility policy: `build-server` 16,601/15,000 ms and `check-service-backed` 178,518/155,000 ms. Diagnostic root `20260801T232445Z-p2539508`. |

All planned narrow and broad functional gates passed. Retained duration
maintenance was intentionally not attempted because the successful full-check
root was timing-ineligible; policy and baselines were not relaxed.

## 5. Explicit non-actions

- Tasks/Decisions routes, OpenAPI, storage, persisted behavior, view identities,
  and the hash-before-decode implementation were not changed.
- Row-ID generation, ASCII ordering, duplicate `ROWS` rejection, fixture
  mismatch enforcement, exact source-table permissions, and contaminated-run
  refusal were not weakened.
- The removed Tasks/Decisions Store facade and other legacy adapters were not
  restored.
- Timing thresholds were not relaxed in response to the prior concurrent
  outlier.

## 6. Handoff exit criteria

Closed. Narrow owner evidence, webserver-backed browser evidence, repository
tests, generated drift, non-retained finalization, and full `make check` all
passed. No threshold-compliant retained root was produced, so retained
finalization remained correctly inapplicable with exact timing facts recorded
above.
