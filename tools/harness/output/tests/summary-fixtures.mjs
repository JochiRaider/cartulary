import fs from 'node:fs';
import path from 'node:path';
import { spawnSync } from 'node:child_process';
import { repoRoot } from '../../contract/index.mjs';
import os from 'node:os';
export const instant = (ms = 0) => new Date(Date.UTC(2026, 0, 1) + ms).toISOString();
export function stepSummary(overrides = {}) {
    return { schema_id: 'cartulary.test_step_summary.v1', target: 'sample', label: 'sample', runner: 'shell',
        status: 'pass', start_time: instant(), end_time: instant(1000), accounting_mode: 'actual', counting_mode: 'counted',
        logical_duration_ms: 1000, executed_duration_ms: 1000, wall_duration_ms: 1000,
        counts: { tests: 7, support: 7, failed: 0, non_test: 0, non_test_failed: 0 }, owners: ['sample'], inventory: [], failures: [], ...overrides };
}
export function targetSummary(target, { tests = 7, steps = 2, executed = 1000, logical = executed, status = 'pass', ...overrides } = {}) {
    return { schema_id: 'cartulary.test_target_summary.v4', target, kind: 'leaf', status,
        start_time: instant(), end_time: instant(1000), executed_duration_ms: executed, logical_duration_ms: logical,
        wall_duration_ms: 1000, critical_path_wall_duration_ms: 1000, reused_duration_ms: 0, derived_duration_ms: 0, teardown_duration_ms: 0,
        accounting_modes: { actual: steps, reused: 0, derived: 0 }, counts: { steps, tests, support: tests, failed: 0, non_test: 0, non_test_failed: 0, packages: 1 },
        failures: [], artifacts: { dir: `fixture/${target}`, timing_json: '' }, ...overrides };
}
export function outputContext(t) {
    const root = fs.mkdtempSync(path.join(os.tmpdir(), 'cartulary-output-contract-'));
    t.after(() => fs.rmSync(root, { recursive: true, force: true }));
    const run = path.join(root, 'run');
    const env = { ...process.env, NODE: process.execPath, CARTULARY_TEST_RESULTS_DIR: root, CARTULARY_TEST_RUN_ID: 'run', CARTULARY_OUTPUT_MODE: 'quiet', CARTULARY_SUPPRESS_CHILD_SUCCESS: '0' };
    for (const name of ['CARTULARY_HARNESS_IDENTITY_PREPARED', 'CARTULARY_TEST_TARGET', 'VERBOSE', 'CI_VERBOSE', 'CARTULARY_HARNESS_GRAPH_CHILD', 'CARTULARY_HARNESS_GRAPH_ARTIFACT_CHILD', 'CARTULARY_HARNESS_SKIP_PREREQUISITES', 'CARTULARY_MAKE_INPUT_SOURCES'])
        delete env[name];
    const write = (relative, value) => {
        const file = path.join(run, relative);
        fs.mkdirSync(path.dirname(file), { recursive: true, mode: 0o700 });
        fs.writeFileSync(file, typeof value === 'string' ? value : JSON.stringify(value), { mode: 0o600 });
        return file;
    };
    const read = (relative) => JSON.parse(fs.readFileSync(path.join(run, relative), 'utf8'));
    const command = (executable, args, changes = {}) => {
        const childEnv = { ...env, ...changes };
        for (const [key, value] of Object.entries(childEnv))
            if (value === undefined)
                delete childEnv[key];
        const result = spawnSync(executable, args, { cwd: repoRoot, env: childEnv, encoding: 'utf8', maxBuffer: 8 * 1024 * 1024 });
        if (result.error)
            throw result.error;
        return { ...result, output: result.stdout + result.stderr };
    };
    return { root, run, env, write, read, command,
        step: (label, args, changes) => command('bash', ['tools/harness/execution/run-step.sh', label, '--', ...args], changes),
        output: (args, changes) => command('bash', ['tools/harness/output/test-output.sh', ...args], changes),
        explain: (args = []) => command(process.execPath, ['tools/harness/diagnostics/explain-run-cli.mjs', '--results-dir', run, ...args]),
    };
}
