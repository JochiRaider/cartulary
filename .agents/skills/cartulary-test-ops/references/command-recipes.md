# Command recipes

Run public Make operations from the selected checkout. Discover changed inputs
with `make help-all` and `make explain-target`; these examples are not a second
target registry. Values in angle brackets are placeholders. Pass literal values
as quoted arguments, without shell command substitution.

## Selection

```sh
make task-guide ROLE=module-author OWNER=<discovered-owner>
make explain-test-owner OWNER=<discovered-owner> JSON=1
make explain-target TARGET=<target> DETAIL=rows JSON=1
make target-plan-json TARGET=<target>
```

Supply the target's discovered selection inputs when required. Owner/row IDs
come from the current catalog. Omitted rows select the command's complete owner
scope; explicit rows select a subset. A service-backed slice closes only its
service-backed dependency scope.

## Execution

```sh
CARTULARY_TEST_RUN_ID=<fresh-literal-id> CARTULARY_OUTPUT_MODE=machine \
  make test-slice OWNER=<owner> ROWS=<exact-row-ids>
```

Use `service-backed-test-slice` for the corresponding authorized scope. Preserve
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
