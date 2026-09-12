# Vitest diagnostics and remediation handoff

This guide is implementation-support material. The Testing Harness NLSpec owns
harness behavior; Core 01/Core 03 and the Network Flow Activity NLSpec own the
product interactions described here. No executable check consumes this guide.

## Reading failures

Use repository-root public Make targets and their retained run roots. Start with
`make explain-run RESULTS_DIR=<run-root>` and the target summary. For a selected
Vitest row, the unit's evidence roster links these files under
`unit-logs/row-<row-id>/`:

- `vitest-failure-details.json`: schema `cartulary.vitest_failure_details.v2`;
- `runner.json`: the redacted built-in reporter output;
- `stdout.log` and `stderr.log`: ordinary unit logs.

Full frontend wrapper diagnostics retain the same filenames in the target's raw
artifact directory. Successful cached rows execute no Vitest invocation and do
not synthesize observations or diagnostic artifacts. Disable the harness cache
when measuring a test or reproducing execution behavior.

The v2 sidecar identifies its run and invocation. Observations retain project,
repository-relative file, full title, leaf test name, terminal status, runner
duration, effective deadline, and original error name/message/stack. Failure
records preserve every original cause and its normalized classification.
Catalog selectors may use an exact full title or an unambiguous exact leaf name;
a collision is an accounting error rather than last-result-wins attribution.

| Observation | Classification | Normalized exit |
| --- | --- | ---: |
| Runner-confirmed Vitest test or hook timeout | `timing/timeout_failure` | 13 |
| Assertion or other product-execution error | `product/test_assertion_failure` | 10 |
| Suite collection/setup error | `infra/preflight_error` | 3 |
| Marker-only error without a preserved cause | `unknown/unknown_failure` | 1 |
| Missing, unsafe, or malformed required diagnostics | `artifact/artifact_error` | 11 |
| Duplicate or contradictory observations | `harness/scheduler_accounting_error` | 11 |
| Parent watchdog expiry | Existing `timing/timeout_failure` handling | 13 |
| Cancellation | Existing interrupted handling | Signal-dependent |

The outer GNU Make process may return its own recipe-failure status. Read the
normalized failure fields in the retained summaries when distinguishing these
outcomes. Playwright's per-test timeout policy is unchanged.

`STACK_TRACE_ERROR` is a stack-formatting marker, not a cause. The pinned Vitest
runner can preserve the real timeout message while its built-in JSON reporter
chooses the registration stack. The harness collector reads the supported
reporter lifecycle before that information is lost. Inspect the original
message and failure kind; do not infer a timeout from elapsed time or treat a
registration frame as a failed focus assertion.

The browser-unit limit remains 15 seconds. Do not increase it or enable retries
to hide a missing readiness condition. Profile rendering, readiness, queries,
and interactions separately. Focus assertions must observe the result of the
interaction, without moving focus to manufacture the expected postcondition.

## Current-only diagnostic cutover

The v2 collector and both parent consumers move together. V1 writers/readers and
optional-sidecar fallbacks are removed. Historical v1 files are investigation
material only, with no automatic translation or dual-reading path. Diagnostic
helpers and schemas already participate in the unit-aware test cache dependency
closure, so their changes invalidate prior cached success.

Collectors write invocation-private observations. Parents validate, correlate,
and redact those observations and the built-in JSON before retention. Only the
parent writes canonical row results. Test duration is measured independently
from process wall duration; a cache receipt is not a new duration measurement.
Rollback of the harness cutover must revert the specification, schema,
collector, consumers, and generated projections together.

## Remediation acceptance mapping

| Gap and owner | Remediation and evidence |
| --- | --- |
| Timeline inspector; Core 01 §7.4.1A and Core 03 REQ-03-104/291/292 | Replace the stale absence assertion with disabled-state, accessible permission-description, and zero-request assertions. Keep the existing selection/refresh/history checks. The dedicated presentation regression proves keyboard exclusion and reviewer/admin eligibility; existing action-owner tests cover role/state/closure admission. |
| Network Flow; NF-REQ-170e/199 and NF-AC-120 | Split off-page Reveal/Close from Saved/Explore, metadata-refresh, and Escape continuity. Both fixtures use 501 vertices and one valid edge; the dedicated rendering-limit case retains 501 vertices and 1,001 edges. Existing navigation and withdrawal tests cover edge/vertex limits and stale focus. |
| Harness; TH-HARNESS-REQ-302 and AC-014/021 | Preserve original runner errors, classify deadlines as timing failures, record exact durations, reject ambiguous evidence, retain all causes, and remove v1 fallback. Real pinned-runner fixtures exercise test/hook timeout, misleading stack, marker-only failure, suite load, repeated leaf names, and redaction through the row adapter and full wrapper. Wrapper fixtures exercise watchdogs and cancellation. |

No production focus logic, product route, view schema, domain vocabulary, stored
state, or graph rendering limit changed. The Timeline failure was test drift;
this investigation did not establish a production focus race.

## September 12, 2026 profiling

Baseline: clean `d6e401ad`, pinned repository toolchain, cache disabled. The
instrumented baseline retained at
`.cartulary/test-results/20260912T194320Z-p60223` took 3,503 ms of actual test time
and 8,141 ms of total harness time. The two after-profile cases at
`.cartulary/test-results/20260912T200526Z-p86666` took 741 ms and 289 ms, with
5,608 ms total harness time. Temporary timing instrumentation was removed.

| Stage | Original monolithic test (ms) | Split cases (ms) |
| --- | ---: | ---: |
| Initial rendering | 504 | 382 in off-page case; 137 including paging in continuity case |
| Initial selection/contributors | 213 | 106 |
| Page away | 174 | 53 |
| Reveal | 457 | 115 |
| Page and Close | 390 | 81 |
| Continuity selection/contributors | 118 | 32 |
| Saved/Explore | 263 | 42 |
| Metadata refresh | 1,116 | 51 |
| Escape | 251 | 26 |

These are local diagnostic samples, not a cross-machine performance claim.
Five independent uncached repetitions passed while the full frontend graph
was also running. Focused runs used the declared default of four Vitest workers;
the concurrent frontend run used the supported two-worker setting. The host was
Linux/WSL2 x86-64, with 19 available CPU tokens and approximately 31 GiB memory
recorded by the harness. These timings are runner observations, excluding
startup and harness overhead.

| Run root under `.cartulary/test-results/` | Reveal/Close (ms) | Continuity/Escape (ms) |
| --- | ---: | ---: |
| `20260912T200820Z-p94818` | 2,731 | 1,073 |
| `20260912T200837Z-p1465` | 1,941 | 798 |
| `20260912T200856Z-p7236` | 2,335 | 898 |
| `20260912T200916Z-p13917` | 1,986 | 765 |
| `20260912T200932Z-p19898` | 3,206 | 1,290 |

Both cases remain below ten seconds in every repetition. The concurrent
`make frontend-unit VITEST_MAX_WORKERS=2 CARTULARY_HARNESS_CACHE_MODE=off`
passed all 592 units in 216.5 seconds at `20260912T200820Z-p94930`.
Historical timeouts remain distinct failure evidence.

## Verification record

All commands run from the repository root through public Make targets. Focused
slices used `CARTULARY_HARNESS_CACHE_MODE=off`. Run roots below are relative to
`.cartulary/test-results/`. The product regression commands were:

```sh
make test-slice OWNER=module.workbook ROWS=module.workbook.frontend_unit.verify_inspector_selection_tab_state_details_rel_d2dc82a4bb CARTULARY_HARNESS_CACHE_MODE=off
make test-slice OWNER=web.workbook ROWS=web.workbook.regression.inspector_review_authorization CARTULARY_HARNESS_CACHE_MODE=off
make test-slice OWNER=web.networkflow ROWS=web.networkflow.regression.exploration_focus CARTULARY_HARNESS_CACHE_MODE=off
```

The harness owner slices use `make test-slice OWNER=<owner>` with owners
`harness.command_surface`, `harness.test_catalog`, and
`harness.evidence_accounting`, again with the cache disabled.

| Verification | Result and run root |
| --- | --- |
| Original Workbook inspector regression | Pass; `20260912T195617Z-p65617` |
| New Workbook presentation authorization regression | Pass; `20260912T195818Z-p66540` |
| Harness command-surface slice, including real diagnostic fixtures | Pass; `20260912T201836Z-p75490`; cancellation-after-publication case also passed at `20260912T202608Z-p64042` |
| Harness test-catalog slice | Pass; `20260912T201540Z-p70476` |
| Harness evidence-accounting slice | Pass; `20260912T201916Z-p79033` |
| `make format` | Pass; `20260912T201915Z-p78898` |
| `make generate` | Pass; final regeneration `20260912T202946Z-p73400` |
| `make agent-finalize` | Pass; final run `20260912T202956Z-p82849` |
| `make harness-contract CARTULARY_HARNESS_CACHE_MODE=off` | Pass; final run `20260912T203012Z-p97276` |
| `make generate-drift` | Pass; `20260912T202004Z-p90187` |
| `make generated-artifact-policy-check` | Pass; `20260912T202017Z-p95645` |
| `make json-shape-check` | Pass; `20260912T202021Z-p96298` |
| `make frontend-typecheck` | Pass; `20260912T202026Z-p97257` |
| `make lint-scripts` | Pass; final run `20260912T203915Z-p68768` |
| `make lint-biome` | Pass; `20260912T202055Z-p43533` |
| `make lint-shell` | Pass; `20260912T202059Z-p48108` |
| `make lint-markdown` | Pass; `20260912T203917Z-p69184` |
| `make check CARTULARY_HARNESS_CACHE_MODE=off` | Pass; all 876 units, `20260912T203058Z-p1222` |

`agent-finalize` ran without `RESULTS_DIR`; retained-run maintenance was skipped
because no qualifying successful full warm check was supplied. No cleanup,
destructive command, or retained-evidence rewriting was requested.

The following development failures remain retained and have been resolved:

- `20260912T195441Z-p64676`: exact-title accounting exposed existing catalog leaf
  selectors; the adapter now requires an unambiguous exact match.
- `20260912T200010Z-p67711` and `20260912T200107Z-p72561`: generated topology was
  stale during the contract migration; regeneration restored parity.
- `20260912T200245Z-p79867`: deliberate negative fixtures inherited the enclosing
  command's failure channel; isolated child contexts fixed the false failure.
- `20260912T200820Z-p94796` and `20260912T201121Z-p56518`: the full-wrapper fixture
  supplied an empty prepared target identity; the corrected fixture uses the
  existing ad hoc identity procedure.
- `20260912T201708Z-p73305`: the full-wrapper negative fixture exposed loss of
  accounting classification when invalid observations prevented retention.
  Summaries now preserve the private invocation failure channel's classification.

The diagnostic matrix also covers required-artifact rejection, malformed JSON,
contradictory observations, duplicate/ambiguous titles, suite-load failure,
watchdog termination, cancellation, redaction, private permissions, and cleanup.
No temporary profiling code, obsolete monolithic selector, or current v1
compatibility path remains. The only v1 literal in executable remediation code
is a fixture asserting its rejection.

The first uncached `make check`, `20260912T202028Z-p97659`, passed 873 of 876
units. Two generation checks rejected the stale topology index after a final
runner edit; the final regeneration and finalizer above resolve that drift.
The remaining failure was `TestResetBucketPreservesNamespaceAndProvesMutation`
in `harness.browser.unit.object_store_fixture_admission`: object-store capability
readiness expired after 32 attempts, with failed cleanup reported. Its retained
row result classifies this as `infra/service_readiness_timeout`, exit 3. This
appears unrelated to the Vitest changes. The uncached isolated command below
passed all three units at `20260912T202949Z-p74829`; no object-store code or
readiness deadlines were changed. The final uncached `make check` at
`20260912T203058Z-p1222` passed all 876 units in 429.9 seconds, including that
object-store fixture and both generation checks. It used the declared default
four Vitest workers. The Network Flow continuity row passed in 3,153 ms of
runner time. This passing run supersedes the failed check for acceptance; both
roots remain retained for investigation.

```sh
make test-slice OWNER=harness.browser ROWS=harness.browser.unit.object_store_fixture_admission CARTULARY_HARNESS_CACHE_MODE=off
```

## Changed files and review boundaries

- Specifications: `docs/testing-harness-nlspec.md` and
  `docs/network-flow-activity-nlspec.md`. `docs/domain.md` was inspected and remains
  unchanged; the existing Core 01/Core 03 action declarations remain the owners.
- Product evidence: `WorkbookShell.inspector.test.tsx`,
  `WorkbookInspectorPresentation.test.tsx`, and `NetworkAnalysisWorkspace.test.tsx`,
  with authored selectors in `tools/test_families/web.workbook.json` and
  `tools/test_families/web.networkflow.json`.
- Diagnostic boundary: `tools/harness/diagnostics/vitest-collector.mjs`,
  `vitest-observations.mjs`, and the `vitest-failure-details.mjs` facade, plus
  `tools/harness/execution/vitest-invocation-cli.mjs`.
- Consumers and retention: the Vitest and row runner adapters, both Vitest shell
  wrappers, the Vitest test-output adapter, and the work-graph unit evidence writer.
- Contracts and routing: v2 failure-details schema replacing v1; schema
  attachments, helper ownership, Fallow entrypoints, test-output schema identity,
  authored task-surface inputs, and their generated Make/manifest/topology outputs.
- Regression fixtures: `tools/harness/execution/tests/test-vitest-diagnostics.mjs`,
  `vitest-diagnostic-fixture.mjs`, and `test-run-vitest-step.sh`.

The inspector resolver/presentation/action-owner boundaries and Network Flow
navigation/focus implementation were inspected to preserve the existing product
contract. The local pinned Vitest reporter implementation informed the collector;
no dependency files or installed runner code were modified.
