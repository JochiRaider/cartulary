import assert from 'node:assert/strict';
import fs from 'node:fs';
import path from 'node:path';
import test from 'node:test';
import { outputContext } from './summary-fixtures.mjs';
test('prepared identity rejects partial input before publishing artifacts', (t) => {
    const c = outputContext(t);
    const result = c.step('partial prepared identity', ['true'], { CARTULARY_TEST_RUN_ID: undefined, CARTULARY_TEST_TARGET: 'backend-unit', CARTULARY_HARNESS_IDENTITY_PREPARED: '1' });
    assert.equal(result.status, 2, result.output);
    assert.match(result.output, /prepared harness artifact identity is incomplete/);
    assert.deepEqual(fs.readdirSync(c.root), []);
});
test('a failing nested child retains its contained identity and parent artifacts', (t) => {
    const c = outputContext(t);
    const plan = c.write('test-slice/test-slice-plan.json', 'parent plan\n');
    const log = c.write('test-slice/scheduler-logs/unit.log', 'parent log\n');
    const childRoot = path.join(c.run, 'test-slice/child-results');
    const result = c.command('bash', ['-c', 'source tools/harness/execution/step-runtime.sh; ensure_harness_artifact_identity; exit 7'], { CARTULARY_TEST_RESULTS_DIR: childRoot, CARTULARY_TEST_RUN_ID: 'child-failure', CARTULARY_TEST_TARGET: 'backend-unit' });
    assert.equal(result.status, 7, result.output);
    assert.equal(fs.readFileSync(plan, 'utf8'), 'parent plan\n');
    assert.equal(fs.readFileSync(log, 'utf8'), 'parent log\n');
    assert.ok(fs.existsSync(path.join(childRoot, 'child-failure')));
    assert.ok(!fs.existsSync(path.join(childRoot, 'adhoc')));
});
test('native failure, argv boundaries, output labels and millisecond timestamps survive the wrapper', (t) => {
    const c = outputContext(t);
    const result = c.step('short failure', ['bash', '-c', 'printf "%s\\n" "$1" >&2; exit 7', 'bash', 'short-failure with spaces']);
    assert.equal(result.status, 7, result.output);
    for (const part of ['failure: short failure', 'coverage=non_test', 'runner=shell', 'message=short-failure with spaces', 'raw='])
        assert.ok(result.output.includes(part), result.output);
    assert.doesNotMatch(result.output, /== short failure ==/);
    const s = c.read('adhoc/short-failure/step-summary.json');
    assert.equal(s.counts.failed, 1);
    assert.equal(s.counts.non_test, 1);
    assert.equal(s.counts.non_test_failed, 1);
    assert.equal(s.counts.unmapped_failed, 0);
    assert.equal(s.failure_class, 'harness');
    assert.equal(s.failure_classes.harness, 1);
    assert.equal(s.failures[0].failure_class, 'harness');
    for (const field of ['start_time', 'end_time'])
        assert.match(s[field], /^\d{4}-\d{2}-\d{2}T\d{2}:\d{2}:\d{2}\.\d{3}Z$/);
});
test('quiet output suppresses successful child logs and overrides VERBOSE', (t) => {
    const c = outputContext(t);
    const result = c.step('success log replay', ['bash', '-c', 'echo keep-this-warning >&2'], { CARTULARY_OUTPUT_ALLOW_SUCCESS_LOG: '1', VERBOSE: '1' });
    assert.equal(result.status, 0, result.output);
    assert.match(result.output, /\[RESULT\] target=adhoc status=pass/);
    assert.doesNotMatch(result.output, /keep-this-warning|== success log replay ==/);
});
test('VERBOSE selects streaming when no explicit output mode exists', (t) => {
    const c = outputContext(t);
    const result = c.step('verbose default', ['bash', '-c', 'echo verbose-stream'], { CARTULARY_OUTPUT_MODE: undefined, VERBOSE: '1' });
    assert.equal(result.status, 0, result.output);
    assert.match(result.output, /== verbose default ==/);
    assert.match(result.output, /verbose-stream/);
});
test('monotonic step duration survives target collation without legacy ambiguity', (t) => {
    const c = outputContext(t);
    const result = c.step('monotonic', ['bash', '-c', 'sleep 0.66']);
    assert.equal(result.status, 0, result.output);
    const step = c.read('adhoc/monotonic/step-summary.json');
    assert.ok(step.wall_duration_ms >= 500);
    assert.ok(step.executed_duration_ms >= 500);
    assert.ok(step.logical_duration_ms >= 500);
    assert.equal(step.duration_ms, undefined);
    assert.equal(c.output(['target-summary', 'adhoc', 'pass']).status, 0);
    const target = c.read('adhoc/target-summary.json');
    const timing = c.read('adhoc/target-timing.json');
    assert.ok(target.totals.wall_duration_ms >= 500);
    assert.ok(target.totals.logical_duration_ms >= 500);
    assert.equal(target.duration_ms, undefined);
    assert.equal(timing.schema_id, 'cartulary.test_target_timing.v1');
});
test('secure capture preserves shell state, drains both streams, rejects symlinks and reports redaction failure', (t) => {
    const c = outputContext(t);
    const result = c.command('bash', ['-c', "\nset -euo pipefail\nsource tools/harness/execution/step-runtime.sh\nsource tools/harness/readiness/process-lifecycle.sh\nroot=\"$1\"\ntrap 'if [[ -n \"${captured_service:-}\" ]]; then stop_process_group \"$captured_service\"; fi' EXIT\numask 022\nstep_secure_mkdir \"$root/first/second/leaf\"\n[[ \"$(umask)\" == 0022 ]]\nfor dir in first first/second first/second/leaf; do [[ \"$(stat -c '%a' \"$root/$dir\")\" == 700 ]]; done\numask 000\nobserved_state=0\ncapture_probe() {\n [[ \"$(stat -c '%a' \"$1\")\" == 600 && \"$(stat -c '%a' \"$2\")\" == 600 ]]\n observed_state=1\n printf 'complete stdout\\n'\n (sleep 0.1; printf 'delayed stderr\\n' >&2) &\n}\nstep_capture_command quiet \"$root/stdout.log\" \"$root/stderr.log\" capture_probe \"$root/stdout.log\" \"$root/stderr.log\"\n[[ \"$observed_state\" == 1 ]]\n[[ \"$(cat \"$root/stdout.log\")\" == 'complete stdout' ]]\n[[ \"$(cat \"$root/stderr.log\")\" == 'delayed stderr' ]]\nprintf 'preserved\\n' >\"$root/protected\"\nln -s \"$root/protected\" \"$root/link\"\nif step_secure_files \"$root/link\" >/dev/null 2>&1; then exit 21; fi\n[[ \"$(cat \"$root/protected\")\" == preserved ]]\ncaptured_service=\"\"\nstart_captured_service() { start_process_group captured_service \"$root/service.log\" sleep 30; }\nstep_capture_command quiet \"$root/service-out.log\" \"$root/service-err.log\" start_captured_service\nprocess_group_running \"$captured_service\"\nstop_process_group \"$captured_service\"\ncaptured_service=\"\"\nstep_redact_stream() { cat >/dev/null; return 7; }\ncapture_status=0\nstep_capture_command quiet \"$root/failure-out.log\" \"$root/failure-err.log\" true || capture_status=$?\n[[ \"$capture_status\" == 11 ]]\n", 'bash', c.root]);
    assert.equal(result.status, 0, result.output);
    assert.equal(fs.statSync(path.join(c.root, 'stdout.log')).mode & 0o777, 0o600);
});
test('parallel step owners and empty shared logs publish separate complete artifacts', (t) => {
    const c = outputContext(t);
    const result = c.command('bash', ['-c', "\nset -euo pipefail\npids=()\nfor target in browser-e2e-webserver-backed browser-e2e-visual; do\n CARTULARY_TEST_TARGET=\"$target\" bash tools/harness/execution/run-step.sh frontend-toolchain -- true >/dev/null &\n pids+=(\"$!\")\ndone\nfor pid in \"${pids[@]}\"; do wait \"$pid\"; done\nmkdir -p \"$CARTULARY_TEST_RESULTS_DIR/run\"\n: >\"$CARTULARY_TEST_RESULTS_DIR/run/stdout.log\"\n: >\"$CARTULARY_TEST_RESULTS_DIR/run/stderr.log\"\npids=()\nfor target in step-a step-b; do\n mkdir -p \"$CARTULARY_TEST_RESULTS_DIR/run/$target\"\n CARTULARY_TEST_TARGET=\"$target\" CARTULARY_STEP_LABEL=\"$target\" CARTULARY_STEP_DIR=\"$CARTULARY_TEST_RESULTS_DIR/run/$target\" CARTULARY_STEP_COMMAND=: CARTULARY_STEP_COMMAND_ARGV='[\":\"]' CARTULARY_STEP_START_TIME=2026-01-01T00:00:00.000Z CARTULARY_STEP_END_TIME=2026-01-01T00:00:00.001Z CARTULARY_STEP_LOGICAL_DURATION_MS=1 CARTULARY_STEP_EXECUTED_DURATION_MS=1 CARTULARY_STEP_WALL_DURATION_MS=1 CARTULARY_STEP_EXIT_STATUS=0 CARTULARY_STEP_STDOUT_LOG=\"$CARTULARY_TEST_RESULTS_DIR/run/stdout.log\" CARTULARY_STEP_STDERR_LOG=\"$CARTULARY_TEST_RESULTS_DIR/run/stderr.log\" bash tools/harness/output/test-output.sh shell-step >/dev/null &\n pids+=(\"$!\")\ndone\nfor pid in \"${pids[@]}\"; do wait \"$pid\"; done\n"]);
    assert.equal(result.status, 0, result.output);
    for (const target of ['browser-e2e-webserver-backed', 'browser-e2e-visual'])
        assert.equal(c.read(`${target}/frontend-toolchain/step-summary.json`).status, 'pass');
    assert.ok(!fs.existsSync(path.join(c.run, 'frontend-toolchain')));
    for (const target of ['step-a', 'step-b'])
        assert.equal(c.read(`${target}/step-summary.json`).status, 'pass');
});
test('obsolete duration input does not supply logical or executed time', (t) => {
    const c = outputContext(t);
    const stdout = c.write('stdout.log', '');
    const stderr = c.write('stderr.log', '');
    fs.mkdirSync(path.join(c.run, 'step'));
    const result = c.output(['shell-step'], { CARTULARY_TEST_TARGET: 'sample', CARTULARY_STEP_LABEL: 'legacy duration', CARTULARY_STEP_DIR: path.join(c.run, 'step'), CARTULARY_STEP_COMMAND: ':', CARTULARY_STEP_START_TIME: '2026-01-01T00:00:00.000Z', CARTULARY_STEP_END_TIME: '2026-01-01T00:00:00.001Z', CARTULARY_STEP_DURATION_MS: '999', CARTULARY_STEP_WALL_DURATION_MS: '1', CARTULARY_STEP_EXIT_STATUS: '0', CARTULARY_STEP_STDOUT_LOG: stdout, CARTULARY_STEP_STDERR_LOG: stderr });
    assert.equal(result.status, 0, result.output);
    const s = c.read('step/step-summary.json');
    assert.equal(s.logical_duration_ms, 0);
    assert.equal(s.executed_duration_ms, 0);
});
