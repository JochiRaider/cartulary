import assert from 'node:assert/strict';
import fs from 'node:fs';
import path from 'node:path';
import test from 'node:test';
import { repoRoot, validateSchemaSync } from '../../contract/index.mjs';
import { buildEvidenceUnit, buildReaderEvidence, persistEvidence } from '../../observability/tests/retained-fixture.mjs';
import { outputContext, targetSummary, stepSummary, instant } from './summary-fixtures.mjs';
import { validFindings, malformedFindings, startupFailure } from './diagnostic-fixtures.mjs';
function assertExit(result, status = 0) { assert.equal(result.status, status, result.output); return result.output; }
const resolveArtifact = (ref) => path.isAbsolute(ref) ? ref : path.join(repoRoot, ref);
test('current finalizer child failure replaces the generic shell failure', (t) => {
    const c = outputContext(t);
    const artifact = { target: 'agent-finalize', status: 'fail', failures: [{ action_id: 'generated_structure_refresh', substep_id: 'generate', target: 'generate', failure_class: 'artifact', failure_reason: 'artifact_error', headline: 'generated output could not be published', summary_json: null }] };
    const file = c.write('finalizer-input.json', artifact);
    const result = c.step('agent-finalize', ['bash', '-c', 'cp "$1" "$(dirname "$CARTULARY_STEP_ARTIFACT_DIR")/finalize-summary.json"; echo "synthetic finalizer wrapper failure" >&2; exit 2', 'bash', file], { CARTULARY_TEST_TARGET: 'agent-finalize' });
    assertExit(result, 2);
    assert.match(result.output, /reason=artifact_error/);
    const s = c.read('agent-finalize/tool-run-summary.json');
    assert.equal(s.failure_class, 'artifact');
    assert.equal(s.failures[0].child_target, 'generate');
    assert.equal(s.failures.length, 1);
});
for (const { name, target, label, message, reason, publicExit } of [
    { name: 'watcher exhaustion', target: 'browser-e2e-webserver-backed', label: 'browser-e2e startup frontend ready', message: 'frontend exited before readiness\nError: ENOSPC: System limit for number of file watchers reached, watch /fixture/vite.config.ts', reason: 'service_start_error', publicExit: 3 },
    { name: 'listener conflict', target: 'browser-e2e-stateful', label: 'browser-e2e startup services', message: 'listen tcp 127.0.0.1:8333: bind: address already in use', reason: 'resource_conflict', publicExit: 4 },
])
    test(`${name} retains native exit and emits the classified public exit`, (t) => {
        const c = outputContext(t);
        const result = c.step(label, ['bash', '-c', 'printf "%s\\n" "$1" >&2; exit 1', 'bash', message], { CARTULARY_TEST_TARGET: target });
        assertExit(result, 1);
        assert.match(result.output, new RegExp(`reason=${reason}`));
        assertExit(c.output(['target-summary', target, 'fail']));
        const summary = c.read(`${target}/target-summary.json`);
        assert.equal(summary.failure_class, 'infra');
        assert.equal(summary.failure_reason, reason);
        assert.equal(c.read(`${target}/tool-run-summary.json`).exit_code, publicExit);
    });
test('filesystem exhaustion names the affected path and available capacity', (t) => {
    const c = outputContext(t);
    const dir = path.join(c.root, 'go-tmp');
    fs.mkdirSync(dir);
    const result = c.step('build server', ['bash', '-c', 'printf "mkdir %s/go-build-fixture: no space left on device (ENOSPC)\\n" "$1" >&2; exit 1', 'bash', dir], { CARTULARY_TEST_TARGET: 'build-server', GO_TMP_DIR: dir });
    assertExit(result, 1);
    const s = c.read('build-server/build-server/step-summary.json');
    assert.equal(s.failure_class, 'infra');
    assert.equal(s.failure_reason, 'resource_conflict');
    assert.ok(s.failures[0].message.includes(`filesystem_capacity path=${dir}`));
    assert.match(s.failures[0].message, /available_bytes=/);
    assert.equal(c.read('build-server/tool-run-summary.json').exit_code, 4);
});
test('browser startup evidence overrides an uninformative shell message', (t) => {
    const c = outputContext(t);
    c.write('_shared/test-services/suite-test/browser-sessions/session-test/startup-diagnostics.json', structuredClone(startupFailure));
    assertExit(c.step('browser-e2e startup frontend ready', ['bash', '-c', 'echo "frontend exited before readiness" >&2; exit 1'], { CARTULARY_TEST_TARGET: 'browser-e2e-webserver-backed' }), 1);
    const s = c.read('browser-e2e-webserver-backed/target-summary.json');
    assert.equal(s.failure_class, 'infra');
    assert.equal(s.failure_reason, 'resource_conflict');
    assert.equal(c.read('browser-e2e-webserver-backed/tool-run-summary.json').exit_code, 4);
});
for (const { name, target, message, expected, reason, status = 1 } of [
    { name: 'progress', target: 'adhoc', message: '[TARGET] start browser-e2e-webserver-backed service_backed=1 expected_steps=0 expected_tests=0\nreal-shell-failure', expected: 'real-shell-failure', reason: 'unknown_failure', status: 9 },
    { name: 'shellcheck stdout', target: 'adhoc', message: "tools/harness/readiness/bootstrap-node-runtime.sh\nIn tools/harness/generated-artifacts/tests/test-json-shapes.sh line 200:\n^-- SC2016 (info): Expressions don't expand in single quotes, use double quotes for that.", expected: "ShellCheck SC2016 at tools/harness/generated-artifacts/tests/test-json-shapes.sh:200: Expressions don't expand in single quotes, use double quotes for that.", reason: 'unknown_failure', status: 11 },
    { name: 'shellcheck', target: 'lint-shell', message: 'In scripts/example.sh line 5:\nunused_var=1\n^-- SC2034 (warning): unused_var appears unused. Verify use (or export if used externally).', expected: 'ShellCheck SC2034 at scripts/example.sh:5: unused_var appears unused. Verify use (or export if used externally).', reason: 'tool_diagnostic_failure' },
    { name: 'biome', target: 'lint-biome', message: 'apps/web/src/example.ts:12:8 lint/style/noNonNullAssertion ━━━━━━━━━━\n  ! Forbidden non-null assertion.', expected: 'Biome lint/style/noNonNullAssertion at apps/web/src/example.ts:12:8', reason: 'tool_diagnostic_failure' },
    { name: 'gosec', target: 'go-gosec-targeted', message: 'G304: Potential file inclusion via variable', expected: 'G304: Potential file inclusion via variable', reason: 'security_finding' },
])
    test(`${name} diagnostics are promoted from real child output`, (t) => {
        const c = outputContext(t);
        const result = c.step('diagnostic', ['bash', '-c', 'printf "%s\\n" "$1"; exit "$2"', 'bash', message, String(status)], { CARTULARY_TEST_TARGET: target });
        assertExit(result, status);
        assert.ok(result.output.includes(expected), result.output);
        const s = c.read(`${target}/diagnostic/step-summary.json`);
        assert.equal(s.dossiers[0].message, expected);
        assert.equal(s.failure_reason, reason);
        if (reason === 'security_finding')
            assert.equal(s.failure_class, 'security');
    });
for (const valid of [true, false])
    test(`Govulncheck ${valid ? 'findings' : 'malformed evidence'} propagates through step, target and tool publication`, (t) => {
        const c = outputContext(t);
        const input = c.write('findings-input.json', structuredClone(valid ? validFindings : malformedFindings));
        assertExit(c.step('go-vulncheck', ['bash', '-c', 'cp "$1" "$CARTULARY_STEP_ARTIFACT_DIR/govulncheck-findings.json"; echo "=== Symbol Results ==="; exit 1', 'bash', input], { CARTULARY_TEST_TARGET: 'go-vulncheck' }), 1);
        for (const file of ['go-vulncheck/step-summary.json', 'target-summary.json', 'tool-run-summary.json']) {
            const s = c.read(`go-vulncheck/${file}`);
            assert.equal(s.failure_class, valid ? 'security' : 'artifact');
            assert.equal(s.failure_reason, valid ? 'security_finding' : 'artifact_error');
            if (file !== 'go-vulncheck/step-summary.json')
                assert.equal(s.extensions?.['cartulary.security']?.govulncheck?.blocking_count, valid ? 1 : undefined);
        }
        if (valid) {
            const tool = c.read('go-vulncheck/tool-run-summary.json');
            assert.ok(tool.summary_artifacts.some(a => a.role === 'govulncheck_findings' && a.path_kind === 'file' && a.format === 'json' && a.path.endsWith('govulncheck-findings.json') && !a.path.startsWith('/') && !a.path.split('/').includes('..')));
        }
    });
test('missing run and summary-group evidence exits with artifact failure', (t) => {
    const c = outputContext(t);
    assertExit(c.output(['run-summary', 'missing-run', 'fail', '0', '1', '-', 'missing-target']), 11);
    let s = c.read('run-summary.json');
    assert.equal(s.failure_class, 'artifact');
    assert.equal(s.counts.failed, 1);
    assert.deepEqual(s.summary_targets.missing, ['missing-target']);
    c.write('present/target-summary.json', targetSummary('present'));
    assertExit(c.output(['run-summary', 'group-run', 'pass', '1', '1', '-', '--summary-groups', 'browser=missing-browser', 'present']), 11);
    s = c.read('run-summary.json');
    assert.equal(s.failure_reason, 'artifact_error');
    assert.deepEqual(s.summary_groups[0].missing_summary_targets, ['missing-browser']);
});
test('failed infrastructure timing rejects otherwise passing tests; scheduled retry remains nonterminal', (t) => {
    const c = outputContext(t);
    c.write('infra/pass/step-summary.json', stepSummary({ target: 'infra' }));
    const event = { schema_id: 'cartulary.test_services.journal_event.v2', producer_id: 'fixture-infra', seq: 1, type: 'timing-span', timestamp: instant(2000), pid: 100, details: { target: 'infra', bucket: 'service_wait', label: 'test-services start object-store', start_time: instant(1000), end_time: instant(2000), duration_ms: 1000, status: 'fail' } };
    c.write('_shared/test-services/suite/journals/fixture-infra.ndjson', JSON.stringify(event) + '\n');
    assertExit(c.output(['target-summary', 'infra', 'pass']));
    let s = c.read('infra/target-summary.json');
    assert.equal(s.failure_class, 'infra');
    assert.equal(s.failure_classes.infra, 1);
    assert.equal(s.failures[0].kind, 'timing');
    assert.match(s.failure_headline, /tests passed/);
    event.details = { ...event.details, target: 'retry', startup_attempt: true, attempt: 1, max_attempts: 2, retryable: true, retry_scheduled: true, retry_blocked_by_context: false };
    c.write('_shared/test-services/suite/journals/fixture-infra.ndjson', JSON.stringify(event) + '\n');
    assertExit(c.output(['target-summary', 'retry', 'pass']));
    s = c.read('retry/target-summary.json');
    assert.equal(s.status, 'pass');
    assert.equal(s.failures.length, 0);
    assert.equal(s.own.timing_failures.length, 0);
    assert.equal(s.totals.counts.failed, 0);
});
test('child publication preserves own and inclusive counts, detail selection and separate shared work', (t) => {
    const c = outputContext(t);
    c.write('child-a/target-summary.json', targetSummary('child-a', { tests: 7, steps: 2, executed: 1000, logical: 1200 }));
    c.write('child-b/target-summary.json', targetSummary('child-b', { tests: 11, steps: 3, executed: 2000 }));
    assertExit(c.step('parent', ['true'], { CARTULARY_TEST_TARGET: 'parent' }));
    const quiet = assertExit(c.output(['target-summary', 'parent', 'pass', '--children', 'child-a,child-b']));
    assert.doesNotMatch(quiet, /\[CHILD\]/);
    assert.match(quiet, /\[ARTIFACTS\]/);
    const parent = c.read('parent/target-summary.json');
    assert.equal(parent.kind, 'aggregate');
    assert.equal(parent.schema_id, 'cartulary.test_target_summary.v4');
    assert.equal(parent.duration_ms, undefined);
    assert.match(parent.own.artifacts.timing_json, /target-timing.json$/);
    assert.equal(parent.own.accounting_modes.actual, 1);
    assert.deepEqual(parent.children.missing, []);
    const timing = c.read('parent/target-timing.json');
    assert.deepEqual(timing.buckets.map(b => b.name), ['test_command', 'report_collation']);
    assert.deepEqual(parent.own.slowest_lifecycle_bucket, timing.slowest_lifecycle_bucket);
    assert.equal(parent.own.counts.tests, 0);
    assert.equal(parent.children.counts.tests, 18);
    assert.equal(parent.totals.counts.tests, 18);
    assert.deepEqual(parent.children.present.map(x => x.target), ['child-a', 'child-b']);
    assertExit(c.output(['shared-execution', 'backend-integration-shards', 'shard-01', 'pass', instant(), instant(7000), '7000', '0', path.join(c.run, '_shared/shard-01/shared-execution.json')]));
    assertExit(c.output(['run-summary', 'check', 'pass', '1', '1', '-', '--summary-groups', 'backend=child-a,child-b;browser=child-b', 'parent']));
    const run = c.read('run-summary.json');
    assert.equal(run.counts.tests, 18);
    assert.equal(run.executed_duration_ms, parent.totals.executed_duration_ms);
    assert.equal(run.logical_duration_ms, parent.totals.logical_duration_ms);
    assert.equal(run.accounting_modes.actual, 6);
    assert.equal(run.summary_groups[0].executed_duration_ms, 3000);
    assert.equal(run.summary_groups[0].wall_duration_ms, 1000);
    assert.equal(run.summary_groups[1].summary_targets[0], 'child-b');
    assert.equal(run.shared_execution_groups[0].executed_duration_ms, 7000);
    assert.equal(run.slowest_target.target, 'parent');
    assert.match(assertExit(c.explain(['--target', 'parent', '--detail', 'children'])), /child-a/);
    assert.match(assertExit(c.explain(['--detail', 'logs']), 1), /DETAIL=logs requires TARGET/);
    assert.match(assertExit(c.explain(['--detail', 'progress']), 1), /DETAIL=progress requires TARGET/);
});
test('skipped descendants preserve the failed root without manufacturing missing artifacts', (t) => {
    const c = outputContext(t);
    assertExit(c.output(['target-summary', 'failed-check', 'fail']));
    assertExit(c.output(['run-summary', 'skipped-run', 'fail', '0', '1', 'failed-check', '--summary-groups', 'harness=failed-check,skipped-check', '--skipped-after-failure', 'skipped-check', 'failed-check', 'skipped-check']), 1);
    const s = c.read('run-summary.json');
    assert.equal(s.counts.failed, 1);
    assert.equal(s.counts.non_test_failed, 1);
    assert.equal(s.work_units.aborted_after, 'failed-check');
    assert.deepEqual(s.summary_targets.missing, []);
    assert.deepEqual(s.summary_targets.skipped_after_failure, ['skipped-check']);
    assert.deepEqual(s.summary_groups[0].skipped_after_failure, ['skipped-check']);
    assert.deepEqual(s.summary_groups[0].missing_summary_targets, []);
});
for (const source of ['explicit', 'child', 'scheduler'])
    test(`${source} skipped-child attribution does not become a missing-child failure`, (t) => {
        const c = outputContext(t);
        c.write('failed/target-summary.json', targetSummary('failed', { status: 'fail', counts: { tests: 1, authoritative: 1, authoritative_failed: 1, failed: 1, steps: 1 } }));
        const skipped = { target: 'later', work_unit: 'later-work', reason: 'schedule_stopped_after_failure', failed_dependency: 'failed' };
        let args = ['--skipped-after-failure', 'later', '--failed-dependency', 'failed'];
        if (source === 'child') {
            c.write('source/target-summary.json', { schema_id: 'cartulary.test_target_summary.v4', target: 'source', status: 'fail', children: { skipped: [skipped] } });
            args = ['--skipped-from-child', 'source'];
        }
        if (source === 'scheduler') {
            c.write('source/scheduler-summary.json', { schema_id: 'cartulary.scheduler_summary.v1', target: 'source', status: 'fail', failed_work_unit: 'failed', skipped_work_units: [{ label: 'later-work', id: 'later-work', aggregate_target: 'later', reason: skipped.reason, failed_dependency: 'failed' }] });
            args = ['--skipped-from-scheduler', 'source'];
        }
        const output = assertExit(c.output(['target-summary', 'parent', 'fail', '--children', 'failed,later', ...args], { CARTULARY_OUTPUT_MODE: 'verbose' }));
        assert.match(output, /\[CHILD-SKIPPED\] parent later/);
        assert.doesNotMatch(output, /\[CHILD-MISSING\] parent later/);
        const s = c.read('parent/target-summary.json');
        assert.deepEqual(s.children.failed_targets, ['failed']);
        assert.deepEqual(s.children.missing, []);
        assert.equal(s.children.skipped[0].target, 'later');
        assert.equal(s.children.skipped[0].failed_dependency, 'failed');
        assert.equal(s.own.counts.non_test_failed, 0);
    });
test('missing and status-only children reject, while scheduler-only failures name the actual failed unit', (t) => {
    const c = outputContext(t);
    assertExit(c.output(['target-summary', 'missing-parent', 'pass', '--children', 'missing']));
    const missing = c.read('missing-parent/target-summary.json');
    assert.equal(missing.status, 'fail');
    assert.equal(missing.failure_class, 'artifact');
    assert.deepEqual(missing.children.missing, ['missing']);
    c.write('status-only/target-summary.json', targetSummary('status-only', { status: 'fail', tests: 0, steps: 0 }));
    assertExit(c.output(['target-summary', 'status-parent', 'pass', '--children', 'status-only']));
    const failed = c.read('status-parent/target-summary.json');
    assert.equal(failed.status, 'fail');
    assert.equal(failed.failure_reason, 'child_target_failure');
    c.write('scheduler-only/scheduler-summary.json', { schema_id: 'cartulary.scheduler_summary.v1', target: 'scheduler-only', status: 'fail', failed_work_unit: 'lint-shell', failed_work_unit_detail: { label: 'lint-shell', id: 'lint-shell', aggregate_target: 'lint-shell', kind: 'work_unit', status: 1, duration_ms: 12, log_file: 'fixture/lint-shell.log' }, skipped_work_units: [{ label: 'source-check-schema', id: 'source-check-schema', aggregate_target: 'source-check-schema', reason: 'schedule_stopped_after_failure', failed_dependency: 'lint-shell' }] });
    const output = assertExit(c.output(['target-summary', 'scheduler-only', 'fail']));
    assert.match(output, /work_unit=lint-shell/);
    assert.match(output, /child_target=lint-shell/);
    assert.doesNotMatch(output, /child_target=source-check-schema/);
});
test('same-run helper evidence is digest-bound and later child summaries exclude parent artifacts', (t) => {
    const c = outputContext(t);
    assertExit(c.step('helper-target', ['bash', '-c', 'echo "helper stdout"; echo "helper stderr" >&2'], { CARTULARY_TEST_TARGET: 'helper-target' }));
    c.write('target-summaries/helper-target.json', { schema_id: 'cartulary.harness_target_summary.v1', target: 'helper-target', status: 'pass' });
    assertExit(c.output(['run-summary', 'check', 'pass', '1', '1', '-', '--helper-units', 'helper-target', '--completed-helper-units', 'helper-target']));
    const s = c.read('run-summary.json');
    const ref = s.helper_units.same_run_artifact_refs[0];
    assert.equal(s.helper_units.artifacts[0].same_run_output_digest_sha256, ref.output_digest_sha256);
    const artifact = JSON.parse(fs.readFileSync(resolveArtifact(ref.artifact), 'utf8'));
    validateSchemaSync('cartulary.same_run_helper_artifact_ref.v2', artifact);
    assert.equal(artifact.reuse_scope, 'same_run_only');
    assert.equal(artifact.accounting_mode, 'helper_reused');
    assert.equal(artifact.scheduler_reused, false);
    assertExit(c.step('late target', ['true'], { CARTULARY_TEST_TARGET: 'late-target' }));
    const late = c.read('late-target/tool-run-summary.json');
    assert.ok(!late.summary_artifacts.some(a => ['run_summary', 'run_tool_run_summary'].includes(a.role)));
    const explanation = assertExit(c.explain());
    for (const part of ['[HELPER] helper-target status=pass steps=1', '[HELPER-REF] helper-target accounting=helper_reused scheduler_reused=false', '[HELPER-STEP] helper-target label=helper-target status=pass'])
        assert.ok(explanation.includes(part), explanation);
});
test('tool-only finalizer diagnostics expose actions, children, logs and absent reports', (t) => {
    const c = outputContext(t);
    c.write('agent-finalize/tool-run-summary.json', { schema_id: 'cartulary.tool_run_summary.v5', target: 'agent-finalize', status: 'pass', exit_code: 0, duration_ms: 1200, output_mode: 'summary', run_root: c.run, summary_artifacts: [], log_artifacts: [], counts: { tests: 0, failed: 0 }, failure_class: null, failure_reason: null, failures: [] });
    assert.match(assertExit(c.explain()), /\[FINALIZE\] missing/);
    const log = c.write('agent-finalize/agent-finalize/stdout.log', 'finalize child stdout\n');
    c.write('agent-finalize/finalize-summary.json', { target: 'agent-finalize', status: 'pass', results_dir_status: 'valid', generated: { status: 'unchanged', updated_file_count: 0 }, run_checks: { status: 'pass' }, actions: [{ action_id: 'generated_structure_refresh', status: 'pass', execution_state: 'executed', duration_ms: 1200, cache: { state: 'miss', reason_code: 'cache_record_missing' }, substeps: [{ id: 'generate', target: 'generate', status: 'pass', summary_json: 'none', stdout_log: log, stderr_log: null }] }], failures: [] });
    const output = assertExit(c.explain());
    assert.match(output, /\[RUN\] tool-summary-only target=agent-finalize/);
    assert.match(output, /\[FINALIZE\] agent-finalize status=pass results_dir_status=valid/);
    assert.match(output, /\[FINALIZE-ACTION\] generated_structure_refresh status=pass execution_state=executed cache_state=miss/);
    assert.match(assertExit(c.explain(['--target', 'agent-finalize', '--detail', 'children'])), /\[FINALIZE-SUBSTEP\] action=generated_structure_refresh id=generate/);
    assert.match(assertExit(c.explain(['--target', 'agent-finalize', '--detail', 'logs'])), /finalize child stdout/);
});
test('canonical reader diagnostics use the shared fixture and reject unknown generations', (t) => {
    const c = outputContext(t);
    const fixture = buildReaderEvidence({ runID: 'run', target: 'sample', commandID: 'cartulary.harness.command.sample.v1', units: [buildEvidenceUnit('row:sample')], outcomes: { 'row:sample': 'failed' } });
    persistEvidence(c.run, fixture);
    c.write('unit-results/row-sample.json', { schema_id: 'cartulary.harness_unit_result.v1', unit_id: 'row:sample', semantic_digest: `sha256:${'b'.repeat(64)}`, status: 'failed', exit_code: 1, signal: null, failure_class: 'product', failure_reason: 'test_assertion_failure', evidence_outputs: [], missing_outputs: [] });
    const summary = assertExit(c.explain(['--target', 'sample']));
    assert.match(summary, /units=1/);
    assert.match(summary, /\[TARGET\] sample status=fail failure_class=product units=1/);
    assert.doesNotMatch(summary, /undefined/);
    const accounting = assertExit(c.explain(['--target', 'sample', '--detail', 'accounting']));
    assert.match(accounting, /\[UNIT\] row:sample status=failed/);
    assert.match(accounting, /\[ACCOUNTING\] units=1/);
    const malformed = c.read('run-summary.json');
    malformed.schema_id = 'cartulary.harness_run_summary.v999';
    c.write('run-summary.json', malformed);
    assert.match(assertExit(c.explain(), 11), /unsupported run summary schema/);
});

test('owned and service cleanup observations publish the disjoint teardown window and failure', (t) => {
 const c=outputContext(t);const target='browser-e2e-webserver-backed';
 for(const [start,end,label] of [[0,1100,'owned'],[500,1500,'overlap'],[1800,2100,'remove-root']]) assertExit(c.output(['timing-span'],{CARTULARY_TEST_TARGET:target,CARTULARY_TIMING_BUCKET:'teardown',CARTULARY_TIMING_LABEL:label,CARTULARY_TIMING_START_TIME:instant(start),CARTULARY_TIMING_END_TIME:instant(end),CARTULARY_TIMING_DURATION_MS:String(end-start)}));
 c.write('_shared/test-services/suite/journals/fixture-cleanup.ndjson',JSON.stringify({schema_id:'cartulary.test_services.journal_event.v2',producer_id:'fixture-cleanup',seq:1,type:'timing-span',timestamp:instant(1800),pid:100,details:{target,bucket:'teardown',label:'failed service cleanup',start_time:instant(1100),end_time:instant(1800),duration_ms:700,status:'fail'}})+'\n');
 assertExit(c.output(['target-summary',target,'pass']));const s=c.read(`${target}/target-summary.json`);const timing=c.read(`${target}/target-timing.json`);
 assert.equal(s.status,'fail');assert.equal(s.failure_class,'artifact');assert.equal(s.failure_classes.artifact,1);assert.equal(s.own.counts.non_test_failed,1);assert.equal(s.totals.teardown_status,'fail');assert.equal(s.totals.teardown_failures[0].label,'failed service cleanup');assert.equal(s.totals.timing_failures[0].bucket,'teardown');
 for(const key of ['wall_duration_ms','critical_path_wall_duration_ms','teardown_duration_ms'])assert.equal(s.totals[key],2100);
 assert.equal(s.start_time,instant());assert.equal(s.end_time,instant(2100));assert.equal(timing.start_time,instant());assert.equal(timing.end_time,instant(2100));assert.equal(timing.status,'fail');assert.equal(timing.buckets[0].duration_ms,2100);assert.equal(timing.buckets[0].spans.length,4);
});
