# 2026-05-12 Harness Security Remediation Observed Failures RCA Handoff

## Purpose

This handoff preserves the failure evidence observed while implementing the harness security remediation. It is for a separate root-cause analysis thread. Do not treat this document as the RCA itself.

The key distinction: one `go test` invocation failed. Several additional `suite failure ...` lines appeared in the same output, but those lines come from negative-path test fixtures and did not independently fail the targeted reruns.

## Environment and Toolchain Context

- Repository: `/home/askahn/code/cartulary`
- Date observed: `2026-05-12`
- UTC evidence timestamp collected after the failure: `2026-05-12T04:43:43Z`
- Host profile: Linux/WSL-style repo path, user `askahn`
- Go: `go version go1.26.3 linux/amd64`
- Node: `v24.15.0`
- pnpm: `10.33.0`
- Working tree at collection time had harness security remediation changes in progress, including `tools/harness/backend/go-target-runner.mjs`, `tools/scheduler_manifest.json`, `tools/execution_topology_render_index.json`, and redaction/permission changes.
- Full command that produced the observed failing output:

```sh
go test ./internal/testutil/harnessredact ./internal/testutil/suiteservices ./tools/testservices
```

The command ran for about `90.941s` and exited nonzero.

## Failure 1: `TestMakeTestFastSharesSingleSuiteAcrossServiceBackedWorkUnits`

### Observed Failure Mode

The `tools/testservices` package failed in `TestMakeTestFastSharesSingleSuiteAcrossServiceBackedWorkUnits`. The inner workflow `make test-fast` exited through `test-fast-service-backed` with status `2`.

The harness failure headline was:

```text
tests passed; harness reason=unknown_failure failure: unknown shard backend-integration-entities-shard-02 for backend-integration-support
```

The public harness result line classified the failure as:

```text
[FAIL] target=test-fast exit_code=1 failure_class=harness reason=unknown_failure work_unit=test-fast-service-backed child_target=test-fast-service-backed duration_ms=67148 headline="tests passed; harness reason=unknown_failure failure: unknown shard backend-integration-entities-shard-02 for backend-integration-support"
```

### Trigger

Triggered by:

```sh
go test ./internal/testutil/harnessredact ./internal/testutil/suiteservices ./tools/testservices
```

The failing test itself calls:

```go
runMakeAndLoadScopeWithEvents(t, "test-fast-suite", nil, "test-fast")
```

Source anchor:

- `tools/testservices/integration_test.go`, `TestMakeTestFastSharesSingleSuiteAcrossServiceBackedWorkUnits`

### Exact Result Path and Artifact Directory

Surviving result root:

```text
/tmp/TestMakeTestFastSharesSingleSuiteAcrossServiceBackedWorkUnits853069809/001/test-fast-suite
```

Surviving service artifact directory:

```text
/tmp/TestMakeTestFastSharesSingleSuiteAcrossServiceBackedWorkUnits853069809/001/test-fast-suite/_shared/test-services/562c491a774fbb76dfcfce8a
```

Surviving files:

```text
/tmp/TestMakeTestFastSharesSingleSuiteAcrossServiceBackedWorkUnits853069809/001/test-fast-suite/_shared/test-services/562c491a774fbb76dfcfce8a/service-scope.json
/tmp/TestMakeTestFastSharesSingleSuiteAcrossServiceBackedWorkUnits853069809/001/test-fast-suite/_shared/test-services/562c491a774fbb76dfcfce8a/events/2026-05-12t04-36-17-63909014z-99047-000001-timing-span.json
/tmp/TestMakeTestFastSharesSingleSuiteAcrossServiceBackedWorkUnits853069809/001/test-fast-suite/_shared/test-services/562c491a774fbb76dfcfce8a/events/2026-05-12t04-36-17-640043379z-99047-000002-timing-span.json
```

The harness output referenced this log artifact, but it was not present when the handoff was collected:

```text
/tmp/TestMakeTestFastSharesSingleSuiteAcrossServiceBackedWorkUnits853069809/001/test-fast-suite/test-fast-service-backed/test-fast-service-backed/stdout.log
```

### Relevant Artifact Content

`service-scope.json` shows services did not start:

```json
{
  "suite_id": "562c491a774fbb76dfcfce8a",
  "run_id": "test-fast-suite",
  "preflight": {
    "docker_ok": false,
    "reaper_ready": false
  },
  "postgres": {
    "started": false
  },
  "minio": {
    "started": false
  },
  "wrapper": {
    "owned_count": 0,
    "pass_through_count": 0
  }
}
```

File modes at collection time were owner-only:

```text
drwx------ /tmp/TestMakeTestFastSharesSingleSuiteAcrossServiceBackedWorkUnits853069809/001/test-fast-suite
drwx------ /tmp/TestMakeTestFastSharesSingleSuiteAcrossServiceBackedWorkUnits853069809/001/test-fast-suite/_shared/test-services/562c491a774fbb76dfcfce8a
-rw------- /tmp/TestMakeTestFastSharesSingleSuiteAcrossServiceBackedWorkUnits853069809/001/test-fast-suite/_shared/test-services/562c491a774fbb76dfcfce8a/service-scope.json
```

### Determinism Assessment

Likely deterministic in the current worktree if the same generated scheduler manifest and Go shard plan remain mismatched. It may be environment-gated because this integration test requires Docker, but the headline points to a scheduler/shard-plan lookup mismatch rather than a raw Docker startup failure.

### Known Reproduction Steps

```sh
go test ./tools/testservices -run TestMakeTestFastSharesSingleSuiteAcrossServiceBackedWorkUnits -count=1 -v
```

For the broader original failure:

```sh
go test ./internal/testutil/harnessredact ./internal/testutil/suiteservices ./tools/testservices
```

### Prior Related Failures or Attempted Fixes

- During this remediation, `make agent-finalize` was run and regenerated `tools/scheduler_manifest.json` and `tools/execution_topology_render_index.json` on its first pass.
- Later `make agent-finalize` reruns were unchanged.
- The generated `tools/scheduler_manifest.json` now contains entries for `backend-integration-entities-shard-02`.
- The failure specifically says `unknown shard backend-integration-entities-shard-02 for backend-integration-support`, so the suspected mismatch is not simply absence from the scheduler manifest; it may be target-specific shard-plan availability for `backend-integration-support`.
- A targeted `go test ./tools/testservices -run 'TestRunRedactsServiceDiagnostics|TestRunRecordsMinIOStartupFailureWithStructuredSummary|TestServiceStartupAttemptTimingSummarizesRetries|TestRunRecordsPostgresTemplateFailureWithStructuredSummary'` passed and does not cover this integration path.

### Hypotheses Already Considered

- **Generated scheduler/shard plan drift:** plausible. `tools/scheduler_manifest.json` changed materially after `agent-finalize`, including browser shard redistribution and entries involving `backend-integration-entities-shard-02`.
- **Target/shard ownership mismatch:** plausible. `tools/harness/backend/go-shard-plan.mjs` and `tools/harness/backend/go-target-runner.mjs` both have `unknown shard` checks; the error names target `backend-integration-support`, not `backend-integration`.
- **Docker unavailable:** possibly present in environment and service summary (`docker_ok=false`), but it does not match the headline `unknown shard ... for backend-integration-support`.
- **Security remediation regression:** possible because `tools/harness/backend/go-target-runner.mjs` was edited for secure writes/redaction, but the same error text names shard resolution rather than permissions/redaction.

### Evidence Useful for RCA

- Surviving service scope path:
  `/tmp/TestMakeTestFastSharesSingleSuiteAcrossServiceBackedWorkUnits853069809/001/test-fast-suite/_shared/test-services/562c491a774fbb76dfcfce8a/service-scope.json`
- Console output from original failed command, including:

```text
make[1]: *** [tools/task_surface.generated.mk:600: test-fast-service-backed] Error 1
[FAIL] target=test-fast exit_code=1 failure_class=harness reason=unknown_failure work_unit=test-fast-service-backed child_target=test-fast-service-backed duration_ms=67148 headline="tests passed; harness reason=unknown_failure failure: unknown shard backend-integration-entities-shard-02 for backend-integration-support"
[ARTIFACTS] target=test-fast root=/tmp/TestMakeTestFastSharesSingleSuiteAcrossServiceBackedWorkUnits853069809/001/test-fast-suite summary_json=tool-run-summary.json log_artifact=test-fast-service-backed/test-fast-service-backed/stdout.log scheduler_json=- progress_log=-
[RERUN] command="make test-fast"
[INVESTIGATE] command="make explain-run RESULTS_DIR=/tmp/TestMakeTestFastSharesSingleSuiteAcrossServiceBackedWorkUnits853069809/001/test-fast-suite"
make: *** [tools/task_surface.generated.mk:595: test-fast] Error 2
```

### Missing Evidence / Instrumentation to Collect

- Re-run the focused test with retained verbose logs:

```sh
CARTULARY_OUTPUT_MODE=verbose go test ./tools/testservices -run TestMakeTestFastSharesSingleSuiteAcrossServiceBackedWorkUnits -count=1 -v
```

- Before rerunning, capture:

```sh
make target-plan-json
node -e 'import("./tools/harness/backend/go-shard-plan.mjs").then(() => console.log("go-shard-plan import ok"))'
rg -n "backend-integration-entities-shard-02|backend-integration-support" tools/scheduler_manifest.json tools/*duration* scripts/lib
```

- Collect the referenced but missing `test-fast-service-backed/test-fast-service-backed/stdout.log` if the rerun retains it.
- Capture `make explain-run RESULTS_DIR=<new-run-root> TARGET=test-fast DETAIL=logs` immediately after reproduction.
- Capture Docker state if Docker availability remains suspect:

```sh
docker version
docker info
```

## Failure 2: Docker Endpoint Refused in Preflight Fixture Output

### Observed Failure Mode

The original failed command printed:

```text
suite failure failure_class=infra startup-preflight: ping docker endpoint unix:///var/run/docker.sock: connection refused [artifacts: /tmp/TestRunFailsFastWhenSuitePreflightFails1869306065/001/wrapper-tests/_shared/test-services/suite-redaction]
```

### Trigger

Same original command:

```sh
go test ./internal/testutil/harnessredact ./internal/testutil/suiteservices ./tools/testservices
```

Specific fixture source appears to be `TestRunFailsFastWhenSuitePreflightFails` in `tools/testservices/main_test.go`.

### Exact Result Path and Artifacts

Referenced artifact directory:

```text
/tmp/TestRunFailsFastWhenSuitePreflightFails1869306065/001/wrapper-tests/_shared/test-services/suite-redaction
```

At handoff collection time this directory no longer existed. It was likely removed by Go test temp cleanup.

### Determinism Assessment

Likely deterministic when Docker is not reachable at `unix:///var/run/docker.sock`. However, in this test this may be an expected negative fixture rather than an unexpected failure.

### Known Reproduction Steps

```sh
go test ./tools/testservices -run TestRunFailsFastWhenSuitePreflightFails -count=1 -v
```

### Prior Related Failures or Attempted Fixes

No fix attempted in this thread. This output was not treated as the primary failing test because targeted service redaction tests later passed.

### Hypotheses Already Considered

- Docker daemon or Docker Desktop integration was unavailable.
- The test intentionally stubs or exercises a preflight failure path and prints the expected suite failure.

### Evidence Useful for RCA

- The output line itself includes the failing endpoint and class/reason context.
- `service-scope.json` from Failure 1 also shows `docker_ok=false`, supporting a host-level Docker availability issue during the run.

### Missing Evidence / Instrumentation to Collect

- Confirm whether this fixture is expected to pass by asserting failure classification.
- Capture `docker version`, `docker info`, and `DOCKER_HOST`.
- Preserve the per-test temp directory by rerunning with a wrapper that copies `t.TempDir()` artifacts before test cleanup, if needed.

## Failure 3: Postgres Startup Fixture Failure Output

### Observed Failure Mode

The original failed command printed:

```text
suite failure failure_class=infra postgres-start (postgres): postgres refused startup [artifacts: /tmp/TestRunCleansUpMinIOWhenPostgresStartupFails1050143735/001/wrapper-tests/_shared/test-services/suite-redaction]
```

### Trigger

Same original command. Specific fixture appears to be `TestRunCleansUpMinIOWhenPostgresStartupFails` in `tools/testservices/main_test.go`.

### Exact Result Path and Artifacts

Referenced artifact directory:

```text
/tmp/TestRunCleansUpMinIOWhenPostgresStartupFails1050143735/001/wrapper-tests/_shared/test-services/suite-redaction
```

Directory was not present during handoff collection.

### Determinism Assessment

Likely deterministic inside the fixture because the test name implies it injects a Postgres startup failure. Treat as expected negative-path output unless a focused rerun fails.

### Known Reproduction Steps

```sh
go test ./tools/testservices -run TestRunCleansUpMinIOWhenPostgresStartupFails -count=1 -v
```

### Prior Related Failures or Attempted Fixes

No RCA attempted. Security remediation changed redaction of service failure messages but not startup control flow.

### Hypotheses Already Considered

- Expected stubbed failure validating cleanup behavior.
- If unexpected, investigate service startup dependency injection in `tools/testservices/main_test.go`.

### Missing Evidence / Instrumentation to Collect

- Focused rerun output.
- `service-scope.json` from that fixture, copied before temp cleanup.
- Whether cleanup assertions passed after startup failure.

## Failure 4: MinIO Startup Failure Fixture Output

### Observed Failure Mode

The original failed command printed:

```text
suite failure failure_class=infra minio-start (minio): docker.sock connection refused password=[REDACTED] access_key=[REDACTED] (retry blocked by context: context ended before retry) [artifacts: /tmp/TestRunRecordsMinIOStartupFailureWithStructuredSummary3608815161/001/wrapper-tests/_shared/test-services/suite-redaction]
```

### Trigger

Same original command. Specific fixture: `TestRunRecordsMinIOStartupFailureWithStructuredSummary`.

### Exact Result Path and Artifacts

Referenced artifact directory:

```text
/tmp/TestRunRecordsMinIOStartupFailureWithStructuredSummary3608815161/001/wrapper-tests/_shared/test-services/suite-redaction
```

Directory was not present during handoff collection.

### Determinism Assessment

Deterministic fixture behavior. A targeted rerun passed:

```sh
go test ./tools/testservices -run 'TestRunRedactsServiceDiagnostics|TestRunRecordsMinIOStartupFailureWithStructuredSummary|TestServiceStartupAttemptTimingSummarizesRetries|TestRunRecordsPostgresTemplateFailureWithStructuredSummary'
```

Result:

```text
ok  	github.com/JochiRaider/cartulary/tools/testservices	0.016s
```

### Known Reproduction Steps

```sh
go test ./tools/testservices -run TestRunRecordsMinIOStartupFailureWithStructuredSummary -count=1 -v
```

### Prior Related Failures or Attempted Fixes

The harness security remediation intentionally changed this output by routing service diagnostics through manifest-backed redaction. The observed output confirms `password` and `access_key` were redacted.

### Hypotheses Already Considered

- This is expected negative-fixture output.
- Redaction path is functioning for key/value secret forms.

### Missing Evidence / Instrumentation to Collect

- None urgent unless a future focused rerun fails. If it does, preserve the service scope JSON and check `failure.message`.

## Failure 5: Postgres Template Failure Fixture Output

### Observed Failure Mode

The original failed command printed:

```text
suite failure failure_class=infra postgres-template (postgres): open postgres admin handle: postgres://cartulary:[REDACTED:dsn]@127.0.0.1:5432/postgres?sslmode=disable [artifacts: /tmp/TestRunRecordsPostgresTemplateFailureWithStructuredSummary669201596/001/wrapper-tests/_shared/test-services/suite-redaction]
```

### Trigger

Same original command. Specific fixture: `TestRunRecordsPostgresTemplateFailureWithStructuredSummary`.

### Exact Result Path and Artifacts

Referenced artifact directory:

```text
/tmp/TestRunRecordsPostgresTemplateFailureWithStructuredSummary669201596/001/wrapper-tests/_shared/test-services/suite-redaction
```

Directory was not present during handoff collection.

### Determinism Assessment

Deterministic fixture behavior. The targeted rerun listed in Failure 4 passed.

### Known Reproduction Steps

```sh
go test ./tools/testservices -run TestRunRecordsPostgresTemplateFailureWithStructuredSummary -count=1 -v
```

### Prior Related Failures or Attempted Fixes

Test expectations were adjusted during remediation to allow URL scheme/host to remain while password/userinfo secret segments are redacted.

### Hypotheses Already Considered

- Expected negative-fixture output.
- Redaction is preserving diagnostic DSN shape while removing the password segment.

### Missing Evidence / Instrumentation to Collect

- If investigating redaction semantics, capture full `service-scope.json` and verify no raw password appears in `failure.message`.

## Failure 6: Child Start Failure Fixture Output

### Observed Failure Mode

The original failed command printed:

```text
suite failure failure_class=harness child-start: exec child failed password=[REDACTED] [artifacts: /tmp/TestRunRecordsChildStartFailureWithStructuredSummary2378349095/001/wrapper-tests/_shared/test-services/suite-redaction]
```

### Trigger

Same original command. Specific fixture appears to be `TestRunRecordsChildStartFailureWithStructuredSummary`.

### Exact Result Path and Artifacts

Referenced artifact directory:

```text
/tmp/TestRunRecordsChildStartFailureWithStructuredSummary2378349095/001/wrapper-tests/_shared/test-services/suite-redaction
```

Directory was not present during handoff collection.

### Determinism Assessment

Likely deterministic fixture behavior.

### Known Reproduction Steps

```sh
go test ./tools/testservices -run TestRunRecordsChildStartFailureWithStructuredSummary -count=1 -v
```

### Prior Related Failures or Attempted Fixes

No direct fix attempted. The redacted output indicates the new redaction path applied to the child-start failure message.

### Hypotheses Already Considered

- Expected negative fixture validating child-start failure classification.

### Missing Evidence / Instrumentation to Collect

- Preserve the fixture `service-scope.json` if a future focused rerun fails.

## Failure 7: `prepare-web-e2e` Required-Environment Diagnostics

### Observed Failure Mode

The original failed command printed:

```text
prepare-web-e2e requires CARTULARY_TEST_SERVICES_ACTIVE=1
prepare-web-e2e requires CARTULARY_PGTEST_TEMPLATE_DB to clone the migrated suite template database
```

### Trigger

Same original command. These lines likely originate from `prepare-web-e2e` negative-path tests in `tools/testservices/main_test.go`.

### Exact Result Path and Artifacts

No artifact path was printed for these two lines in the observed output.

### Determinism Assessment

Likely deterministic negative-fixture behavior.

### Known Reproduction Steps

Search and run the focused `prepare-web-e2e` tests:

```sh
rg -n "prepare-web-e2e requires" tools/testservices
go test ./tools/testservices -run PrepareWebE2E -count=1 -v
```

### Prior Related Failures or Attempted Fixes

No fix attempted. Security remediation touched browser env/session file modes in `scripts/start-web-e2e.sh`, but these diagnostics are from the Go `tools/testservices` helper.

### Hypotheses Already Considered

- Expected validation diagnostics for missing required env.

### Missing Evidence / Instrumentation to Collect

- Focused test names and stdout/stderr separation.
- Whether these diagnostics should be suppressed or captured in retained artifacts rather than printed during otherwise passing tests.

## Failure 8: Browser E2E Cleanup Leak Fixture Output

### Observed Failure Mode

The original failed command printed:

```text
suite failure failure_class=artifact cleanup-web-e2e: browser e2e fixture database "ct_web" has 1 active postgres connection [artifacts: /tmp/TestCleanupOwnedServicesFailsFastOnBrowserFixtureLeak2872211222/001/wrapper-tests/_shared/test-services/suite-web-e2e-cleanup-fail]
```

### Trigger

Same original command. Specific fixture: `TestCleanupOwnedServicesFailsFastOnBrowserFixtureLeak`.

### Exact Result Path and Artifacts

Referenced artifact directory:

```text
/tmp/TestCleanupOwnedServicesFailsFastOnBrowserFixtureLeak2872211222/001/wrapper-tests/_shared/test-services/suite-web-e2e-cleanup-fail
```

Directory was not present during handoff collection.

### Determinism Assessment

Likely deterministic negative-fixture behavior. It intentionally appears to validate that cleanup fails fast on leaked browser fixture DB connections.

### Known Reproduction Steps

```sh
go test ./tools/testservices -run TestCleanupOwnedServicesFailsFastOnBrowserFixtureLeak -count=1 -v
```

### Prior Related Failures or Attempted Fixes

No direct fix attempted.

### Hypotheses Already Considered

- Expected fixture leak simulation.
- Not related to the shard mismatch primary failure.

### Missing Evidence / Instrumentation to Collect

- Preserve service scope and cleanup diagnostics on focused rerun.
- Confirm whether stdout emission is expected for passing negative-path tests.

## Failure 9: Browser E2E Env Write Failure Fixture Output

### Observed Failure Mode

The original failed command printed:

```text
write browser e2e env: open /tmp/TestPrepareWebE2ECleansFixtureWhenEnvWriteFails2922203932/002: is a directory
```

### Trigger

Same original command. Specific fixture: `TestPrepareWebE2ECleansFixtureWhenEnvWriteFails`.

### Exact Result Path and Artifacts

Referenced path:

```text
/tmp/TestPrepareWebE2ECleansFixtureWhenEnvWriteFails2922203932/002
```

Path was not present during handoff collection.

### Determinism Assessment

Likely deterministic negative-fixture behavior: the test appears to pass a directory where an env file should be written.

### Known Reproduction Steps

```sh
go test ./tools/testservices -run TestPrepareWebE2ECleansFixtureWhenEnvWriteFails -count=1 -v
```

### Prior Related Failures or Attempted Fixes

No direct fix attempted. The harness security remediation did change some browser E2E shell env/session file modes in `scripts/start-web-e2e.sh`, but this observed line comes from `tools/testservices`.

### Hypotheses Already Considered

- Expected write-failure fixture output.

### Missing Evidence / Instrumentation to Collect

- Focused rerun output.
- The fixture’s cleanup assertions and any generated metadata before temp cleanup.

## Verification That Passed After the Observed Failure

These checks passed after the failure and should help narrow RCA scope:

```sh
node --check scripts/lib/harness-contract.mjs scripts/lib/test-output/cli.mjs scripts/run-make-node-tool.mjs scripts/agent-finalize.mjs
bash -n scripts/lib/run-phase-common.sh scripts/start-web-e2e.sh
bash scripts/test-run-make-sequence-fast.sh
bash scripts/test-run-go-target.sh
go test ./internal/testutil/harnessredact ./internal/testutil/suiteservices
go test ./tools/testservices -run 'TestRunRedactsServiceDiagnostics|TestRunRecordsMinIOStartupFailureWithStructuredSummary|TestServiceStartupAttemptTimingSummarizesRetries|TestRunRecordsPostgresTemplateFailureWithStructuredSummary'
make json-shape-check
make lint-shell
git diff --check
make agent-finalize
```

`make agent-finalize` final run:

```text
run_root=.cartulary/test-results/20260512T044049Z-p14937
status=pass
generated=unchanged
```

## Recommended Next RCA Flow

1. Start with Failure 1. It is the only observed actual failing test.
2. Do not spend time debugging the negative-fixture `suite failure` lines unless a focused rerun of those tests fails.
3. Reproduce `TestMakeTestFastSharesSingleSuiteAcrossServiceBackedWorkUnits` with verbose retained artifacts.
4. Compare `backend-integration-support` shard definitions across:
   - `tools/scheduler_manifest.json`
   - `tools/harness/backend/go-shard-plan.mjs`
   - `tools/harness/backend/go-target-runner.mjs`
   - any generated target-plan output
5. Collect Docker availability evidence in the same shell before rerun, because surviving service scope also recorded `docker_ok=false`.
6. Preserve `/tmp/TestMakeTestFastSharesSingleSuiteAcrossServiceBackedWorkUnits...` before Go cleanup if possible, especially the referenced `stdout.log` and target summaries.

## Remediation Update Collected 2026-05-12

The follow-up remediation resolved the observed current `make check` gaps.

Root causes fixed:

- Structured redaction used broad key substring matching and replaced `service_sessions` with a redaction string. The redaction manifest and shared Go/Node redactors now use exact or anchored credential-key patterns and preserve schema-owned structural fields while still redacting nested secret values.
- `generate-drift` used hand-maintained scratch copy lists and omitted `scripts/harness-contract.mjs`. Scratch replay is now driven by `tools/generate_drift_scratch_inputs.json`, and the shell smoke covers missing declared scratch inputs.
- Scheduler/runner shard ownership is now guarded by a harness contract test that ensures each generated `go_shard` unit is executable by its declared target.
- Secure retained artifact writes are centralized through `scripts/lib/artifact-writer.mjs`, with existing `harness-contract` exports kept as compatibility entrypoints.
- The frontend failures from the retained run did not reproduce against the remediated tree; focused affected Vitest files and `make frontend-unit` passed without product-code changes.

Validation completed:

```text
go test ./internal/testutil/harnessredact ./internal/testutil/suiteservices
node --test ./scripts/test-harness-contracts.mjs
bash scripts/test-generate-drift.sh
make harness-contract-tests
make json-shape-check
make check-harness-smoke
make generate-drift
make frontend-unit
make service-backed-unit-check
make phase-schedule-drift
go test ./tools/testservices -run TestMakeTestFastSharesSingleSuiteAcrossServiceBackedWorkUnits -count=1 -v
make generated-artifact-policy-check
make agent-finalize
make lint-shell
make check
git diff --check
```

Final successful `make check` evidence:

```text
run_root=.cartulary/test-results/20260512T120842Z-p24517
status=pass
work_units=89/89
tests=507
failed=0
```

`make agent-finalize` result:

```text
status=pass
generated=unchanged
RESULTS_DIR maintenance skipped because RESULTS_DIR was unset
```
