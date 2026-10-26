# Command recipes

Run public Make operations from the selected checkout. Discover changed inputs
with `make help ALL=1` and `make target-plan`; these examples are not a second
target registry. Values in angle brackets are placeholders. Pass literal values
as quoted arguments, without shell command substitution.

## Selection

```sh
make task-guide OWNER=<discovered-owner>
make task-guide DETAIL=coverage OWNER=<discovered-owner> JSON=1
make target-plan TARGET=<target> DETAIL=rows JSON=1
make target-plan TARGET=<target> INPUTS=<selection-json-file> DETAIL=rows JSON=1
make target-plan JSON=1 TARGET=<target>
```

For `target-plan`, put the planned target's selection in an `INPUTS` JSON object,
for example `{"SUITE":"visual"}` or `{"OWNER":"harness.browser","ROWS":"<row-id>"}`.
Do not pass the planned target's `SUITE`, `OWNER` or `ROWS` directly to
`target-plan`; its own input contract rejects them. Owner/row IDs
come from the current catalog. Omitted rows select the owner scope excluding `explicit_only` rows; exact
rows authorize that subset, including explicit-only work. A service-backed slice closes only its
service-backed dependency scope.

Graph-backed plans include every execution child projection under the same
inputs and capacity. For example, `browser-e2e` with `SUITE=visual` includes the
`browser-e2e-visual` child used by partition audits. Planning launches no fixture
or service work. A successful plan is selection evidence only; verify retained
run validity and then audit the complete owner union separately.

Visual and UI-Review are explicit-only even for releases:

```sh
make browser-e2e SUITE=visual
make harness-contract SUITE=ui-review
make ci BROWSER_LANES=6
```

Browser lanes default to four and accept `1..6`; six is an upper bound subject to
host admission. Each group uses one worker. `PLAYWRIGHT_WORKERS` and browser
`SUITE=all` are retired without aliases. Snapshot update selects only screenshot
rows and requires its existing authorization.

## Final validation portfolio

Compare current target plans by exact selected rows/cases, runtime and fixture
profiles, and required non-row obligations. CI already contains its check policy;
do not automatically run check, CI, every browser suite and complete owners.
Visual, UI-review, release-only product measurements and independent order or
process witnesses can add obligations that CI omits. Record the missing coverage
or distinct purpose before each additional invocation. A reporting-suite label
alone does not justify repeating compatible execution.

After focused checks, run repository-required `make agent-finalize`, review its
changes, then freeze source and execute the necessary portfolio. Pass
`RESULTS_DIR` only for a compatible successful full warm-check root; otherwise
report retained maintenance skipped. Do not launch a redundant check to obtain
such a root. Evidence audits use explicit compatible roots and do not execute
missing tests. After a repair, select the affected obligations and re-evaluate
evidence compatibility; do not invent partial-source equivalence or repeat the
whole historical portfolio by habit.

## Execution

```sh
CARTULARY_TEST_RUN_ID=<fresh-literal-id> CARTULARY_OUTPUT_MODE=machine \
  make test-slice OWNER=<owner> ROWS=<exact-row-ids>
```

Use `test-slice SCOPE=services` for the corresponding authorized scope. Preserve
the native execution-session handle and its outer process exit. The default
result parent is `.cartulary/test-results`; `CARTULARY_TEST_RESULTS_DIR` overrides
the parent, not the joined run directory. Confirm the normalized emitted identity.

Machine output is one final JSON value, not progress. Use it with structured
observation; otherwise prefer supported bounded human output such as `ci` and
native-session observation. Do not pipe away Make's exit status or detach work
with shell jobs. A fresh-execution reproduction uses a fresh run ID and, when
required, supported `CARTULARY_HARNESS_CACHE_MODE=off`; preserve the old root.

## Observation

First confirm the capability exists in the current public task surface.

```sh
make test-run-status RESULTS_DIR=<exact-run-dir> JSON=1
make test-run-status RESULTS_DIR=<exact-run-dir> \
  AFTER_REVISION=<last-revision> WAIT_SECONDS=30 JSON=1
```

Alternatively supply `RESULTS_DIR=<parent> RUN_ID=<exact-id>`. `TARGET` selects a
projection inside that invocation. Do not combine `JSON=1` with global machine
mode. Cursor zero waits for an initial publication. A future cursor is invalid.

`operation_exit_code` belongs to the reader. Failed test summaries still produce
successful observations. Wait expiry is not a test timeout. `not_published`
establishes neither failure nor process death; `unavailable` reports a producer's
explicit limitation. A stale snapshot remains only the last observation.
`paused_measurement` ends waiting immediately; use the native session instead
of another polling command during the known pause.

If the command is absent, do not call private scripts or create a replacement.
Use native execution observation and existing terminal operations.

## Terminal investigation

```sh
make explain-run RESULTS_DIR=<exact-run-dir>
make explain-run RESULTS_DIR=<exact-run-dir> TARGET=<target> DETAIL=children
make explain-run RESULTS_DIR=<exact-run-dir> TARGET=<target> DETAIL=accounting
```

Only invoke these with an exact, identified published run. Their existing result
parent discovery is not authority to attach to a newest run. Prefer bounded
inspection of returned artifact references over broad `DETAIL=logs` output.
Run evidence-audit gates only when the requested acceptance claim requires them.

## Opt-in instrumentation

Graph-backed Make targets accept command-line `HARNESS_DIAGNOSTICS=basic`;
`off` is the default. Use the same narrow owner/row selection and fresh run-ID
rules above. The setting does not enable application telemetry or network export.

```sh
make test-slice OWNER=<owner> ROWS=<exact-row-ids> HARNESS_DIAGNOSTICS=basic
make explain-run RESULTS_DIR=<exact-run-dir> DETAIL=performance
make explain-run RESULTS_DIR=<exact-run-dir> DETAIL=performance JSON=1
make harness-observability-check RESULTS_DIR=<exact-run-dir>
```

Performance detail, including resource observations, requires an exact run (or
parent plus RUN_ID). Use `JSON=1` with `DETAIL=performance`; there is no separate
`DETAIL=resources` selection in the current public command. These are current-epoch
readers; older schemas are historical evidence, not silently translated inputs.
Live status includes the last resource sample when available, its graph-process
clock, age and staleness. Keep using the native session during measurement pause.

For explicitly requested invocation resource accounting, graph targets accept
`HARNESS_ACCOUNTING=required` together with `HARNESS_DIAGNOSTICS=basic` and
`CARTULARY_HARNESS_CACHE_MODE=off`. The prepared native harness requires a fresh
delegated cgroup v2 scope and provable local placement of owned service containers.
Prepare and execute with consistent native build settings. Workloads containing
nested ordinary Go-runner commands use `CGO_ENABLED=0` for preparation and the
outer command, matching the runner's existing cgo-disabled profile. A different
or stale controller executable fails required accounting before graph work.
Nested graphs inherit the proven outer accounting scope's basic diagnostics and
disabled execution-result reuse, including after an independent command clears
inherited environment variables. Unaccounted commands keep their normal defaults.
Unavailable prerequisites fail the request; do not enable privileges or change
host settings merely to obtain a measurement. The performance view reports
cumulative charged CPU and peak cgroup memory, with controller and shared-host
exclusions stated. These differ from sampled CPU lower bounds and sampled RSS.
`CHECK=all` requires a passing run. Use `CHECK=diagnostics` for deliberately
failed or cancelled runs; that check does not establish passing test assurance.

Compare two exact runs without selecting newer directories implicitly:

```sh
make explain-run RESULTS_DIR=<current-run> DETAIL=performance JSON=1 \
  COMPARE_RESULTS_DIR=<reference-run> COMPARISON=equivalent
```

Use `COMPARISON=instrumentation` only for matched same-source off/basic treatments.
Incompatible inputs retain separate observations and reason codes, with a null
delta. The comparison is descriptive and never a qualifying performance gate.
