# Harness remediation salvage

The user rejected the combined candidate after its first revised CI comparison
was 2.84% slower, and requires at least a 5% improvement for a future performance
candidate. This salvage is a smaller correctness and structural change. It makes
no CI speedup claim. Full CI and the comparison repetitions remain cancelled.

## Preserved experiment

The complete rejected working-tree change is saved outside executable inputs at
`/home/jochi/.codex/harness-remediation-rejected-20261024`:

- `tracked.patch` contains the binary diff against the original Git revision.
- `untracked.tar.gz` contains all 24 new files, including the full historical handoff.
- `snapshot.json` records the base revision, paths and SHA-256 archive checksums.

Apply that snapshot only to a separate clean checkout of its recorded base. Do not
apply it over the salvage. Retained test roots remain in the candidate and reference
worktrees; the snapshot is source preservation, not an archive of ignored test data.
The [original ledger](harness-remediation-20261023.md) retains all PostgreSQL and
browser dispositions, coverage successors and successful, failed and cancelled runs.

## Retained changes

- W1: reporting-neutral, per-row resolved Go execution contracts; explicit runtime
  binaries and dependencies; published exact selections; matching CPU/GOMAXPROCS.
- W2: seven approved rollback conversions, the administrative-audit split with an
  independent committed-state reader, narrow existing database ports, and isolation
  regression coverage. The original row-by-row audit remains historical evidence.
- W4: service-free visual registry validation, evidence successors, incident decoder
  coverage and representative browser/assembled HTTP coverage.
- Independently reproduced saved-graph review, container observation cleanup,
  never-started service retirement and owned frontend-publication fixes.
- Bounded Vitest lifecycle instrumentation, public plan/retained execution capacity
  parity, physical residency accounting assertions and exact selection invariants.

## Deferred changes

- W5 is restored atomically to the original cumulative Playwright producer,
  runner, reader and finalizer. The new case/disposition schemas are removed from
  active authored inputs. The complete incremental implementation and its security,
  teardown, interruption and linear-volume tests remain in the preserved snapshot.
- W3's four preference rows return to isolated execution. Their source proofs and
  successful neighbor tests are preserved; they do not qualify a new sharing change.
- W6's shard-count search, two-wave bound, resource-pool optimizer, calibration
  registry and tuning constants are removed. Browser and Vitest packing return to
  their original bounded first-fit policy. Go reuses that policy within its retained
  normalized contracts, without speculative parallel splits. Oversized rows stay
  explicit; hard timeouts, exact selection and native admission remain authoritative.

Schema producers and consumers move together; generated outputs are refreshed only
through Make. A future checkpoint cutover must again carry its matching loaders,
schemas and finalizers. A future performance attempt must establish benefit from
the affected bottleneck before another expensive whole-CI comparison is considered.

## Validation

Salvage implementation and focused validation are complete. The validated executable
source is `sha256:ad1830ecc92983066ce9313c99da2daa36c5850cc92c597d2d14b5249a447be6`.
Earlier evidence belongs to its recorded source; this change does not inherit the
rejected candidate's full-run acceptance or claim a performance improvement.

Every run below has the prefix `20261024-harness-salvage-` and is retained under
`.cartulary/test-results/` in the attached remediation worktree. Each command's
outer exit was independently observed as zero. Durations describe the particular
run, not a controlled performance comparison; some focused checks overlapped.

| Run suffix | Command or selection | Result |
| --- | --- | --- |
| `format-01` | `make format` | PASS; 8121 ms |
| `generate-01`, `generate-02` | `make generate` before and after final test changes | PASS; 10980 and 10540 ms |
| `registry-01` | `make test-slice OWNER=web.design ROWS=web.design.static.visual_fixture_registry` | PASS; 1/1 units, 758 ms; graph contains one operation, fixture `none`, no service dependencies |
| `regression-01` | `make harness-contract SUITE=regression` | PASS; 46/46 units, 251775 ms |
| `browser-contract-01` | `make test-slice OWNER=harness.browser` selecting `playwright_adapter`, `browser_lane_policy` and `native_fixtures` by their full catalog row IDs | PASS; 4/4 units, 61497 ms |
| `finalize-01` | `make agent-finalize` | PASS; 26040 ms; generated files unchanged |
| `generated-01` | `make source-check CHECK=generated` | PASS; 16246 ms |
| `schema-01` | `make source-check CHECK=schema` | PASS; 6747 ms |
| `publication-browser-01` | `make test-slice OWNER=harness.browser ROWS=harness.browser.boundary_support.frontend_artifact_lifetime` | PASS; 12/12 units, 99820 ms |

The registry, regression, browser-contract and publication-browser selections used
`CARTULARY_HARNESS_CACHE_MODE=off HARNESS_DIAGNOSTICS=basic`. Their manifests retain
the full exact row selections, capacity and graph identities. The native fixture
coverage includes cancellation before service admission; the final browser probe
exercises the retained owned-publication helper against the restored cumulative
reporter. Graph tests cover reporting-neutral grouping, incompatible dependencies,
exact selection, CPU claims, deterministic timeout packing and public-plan parity.

No full `make ci`, repeated comparison, full warm `make check`, or full visual suite
was run for the salvage. Verification was kept to the affected harness boundaries
and one browser integration probe. Retained-run maintenance was skipped because
`RESULTS_DIR` was unset and there is no new full warm check for this source. The
old full-run results remain historical evidence. No screenshot baselines changed.
