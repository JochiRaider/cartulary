#!/usr/bin/env bash
set -euo pipefail
ROOT_DIR="$(unset CDPATH && cd -- "$(dirname "$0")/../../.." && pwd)"
cd "$ROOT_DIR"
"${NODE_BIN:?Make-owned Node required}" --input-type=module - <<'NODE'
import assert from 'node:assert/strict';
import { mkdtempSync, writeFileSync, readFileSync, existsSync, rmSync } from 'node:fs';
import { tmpdir } from 'node:os';
import path from 'node:path';
import { spawnSync } from 'node:child_process';
import { WorkGraphCompiler } from './tools/harness/scheduler/work-graph/index.mjs';
import { buildOperationInvocation } from './tools/harness/scheduler/work-graph/operations.mjs';
const compiler = new WorkGraphCompiler(process.cwd());
const plan = compiler.compileAggregatePlan('release-check');
const producer = 'operation:harness.release_inventory_artifacts';
assert.equal(plan.graph.units.filter((unit) => unit.unit_id === producer).length, 1, 'one inventory producer serves both consumers');
for (const id of ['harness.license_report', 'harness.sbom']) {
  const consumer = plan.graph.units.find((unit) => unit.unit_id === `operation:${id}`);
  assert.ok(consumer.needs.includes(producer));
}
const env = { ...process.env, MAKEFLAGS: '', MFLAGS: '' };
for (const name of ['MAKEOVERRIDES', 'CARTULARY_HARNESS_IDENTITY_PREPARED', 'CARTULARY_TEST_TARGET', 'CARTULARY_MAKE_INPUT_SOURCES']) delete env[name];
const help = spawnSync('make', ['--no-print-directory', 'help', 'ALL=1', 'JSON=1'], { env, encoding: 'utf8' });
assert.equal(help.status, 0, help.stderr);
const text = help.stdout;
assert.match(text, /release-check/u);
for (const name of ['license-report', 'sbom', 'release-readiness-evidence', 'release-inventory-artifacts']) {
  assert.equal(compiler.taskSurface.targets.some((entry) => entry.name === name), false, `${name}: private operation only`);
  const rejected = spawnSync('make', ['--no-print-directory', name], { env, encoding: 'utf8' });
  assert.equal(rejected.status, 2);
  assert.match(rejected.stderr, /No rule to make target/u);
}
const directory = mkdtempSync(path.join(tmpdir(), 'release-artifact-consumers-'));
try {
  for (const [operation, label, valid, legacy] of [
    ['harness.license_report', 'license report', { schema_id: 'cartulary.license_report.v2', semantic_input_digest: `sha256:${'a'.repeat(64)}`, entries: [] }, { schema_id: 'cartulary.license_report.v1', entries: [] }],
    ['harness.sbom', 'SBOM', { bomFormat: 'CycloneDX', specVersion: '1.7', serialNumber: 'urn:uuid:aaaaaaaa-aaaa-5aaa-8aaa-aaaaaaaaaaaa', version: 1, metadata: { component: { type: 'application', name: 'fixture', properties: [{ name: 'cartulary:semantic_input_digest', value: `sha256:${'a'.repeat(64)}` }] } } }, null],
  ]) {
    for (const [scenario, bytes] of [['missing', null], ['empty', ''], ['valid', JSON.stringify(valid)], ...(legacy ? [['legacy', JSON.stringify(legacy)]] : [])]) {
      const file = path.join(directory, `${operation}-${scenario}.json`);
      if (bytes !== null) writeFileSync(file, bytes);
      // Execute only the already-scheduled consumer. A warm inventory cache must
      // neither repair malformed evidence nor create a missing artifact.
      const invocation = buildOperationInvocation(compiler.owner, operation, { ...env, NODE_BIN: process.execPath, RUN_STEP_SCRIPT: path.resolve('tools/harness/execution/run-step.sh'), CARTULARY_HARNESS_GRAPH_CHILD: '1', LICENSE_REPORT_ARTIFACT: file, SBOM_ARTIFACT: file });
      const result = spawnSync(invocation.command, invocation.args, { env: invocation.env, encoding: 'utf8' });
      assert.ifError(result.error);
      assert.equal(result.status === 0, scenario === 'valid', `${label} ${scenario}: ${result.stderr}`);
      if (scenario !== 'valid') assert.match(result.stderr, /artifact missing|artifact is empty|unsupported license report schema/u);
      assert.equal(existsSync(file), bytes !== null, 'consumer does not generate evidence');
      if (bytes !== null) assert.equal(readFileSync(file, 'utf8'), bytes, 'consumer is read-only');
    }
  }
} finally { rmSync(directory, { recursive: true, force: true }); }
process.stdout.write('release surface, shared inventory and read-only consumer boundaries passed\n');
NODE
