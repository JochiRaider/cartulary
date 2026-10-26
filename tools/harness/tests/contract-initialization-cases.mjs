import assert from "node:assert/strict";
import { spawnSync } from "node:child_process";
import { mkdtempSync, rmSync, writeFileSync } from "node:fs";
import { tmpdir } from "node:os";
import path from "node:path";
import { pathToFileURL } from "node:url";
import { createContractTestContext } from "./contract-test-context.mjs";

export function assertLazyCaseContext(context) {
  const loads = [];
  const overrides = Object.fromEntries([
    "taskSurface", "topology", "rowMigrations", "catalog", "catalogMetadata", "compiler",
    "cacheRegistry", "fixtureBuilderPolicy",
  ].map((name) => [name, () => { loads.push(name); return { name }; }]));
  const first = createContractTestContext(context.root, overrides);
  const second = createContractTestContext(context.root, overrides);
  assert.deepEqual(loads, []);
  assert.equal(first.taskSurface, first.taskSurface);
  assert.deepEqual(loads, ["taskSurface"]);
  assert.notEqual(first.taskSurface, second.taskSurface);
  assert.deepEqual(loads, ["taskSurface", "taskSurface"]);
}

export function assertImportAndFailureIsolation(context) {
  const supportModules = [
    "tests/contract-test-context.mjs", "tests/initialization-tests.mjs",
    "command-surface/tests/test-contracts.mjs", "generated-artifacts/tests/test-boundary-projections.mjs",
    "performance-fixture/tests/test-evidence-generation.mjs", "runtime/tests/test-producer-umask.mjs",
    "test-catalog/tests/test-contracts.mjs", "scheduler/work-graph/tests/test-contracts.mjs",
    "scheduler/work-graph/tests/test-scheduler-contracts.mjs", "readiness/tests/test-frontend-artifact.mjs",
    "static-analysis/tests/test-exact-file-sets.mjs", "tests/test-command-failure.mjs",
    "tests/test-frontend-producer-graph.mjs", "tests/test-frontend-producer-lifecycle.mjs",
  ].map((file) => pathToFileURL(path.join(context.root, "tools/harness", file)).href);
  const contextURL = pathToFileURL(path.join(context.root, "tools/harness/tests/contract-test-context.mjs")).href;
  const policyURL = pathToFileURL(path.join(context.root, "tools/harness/test-support/owner-fixtures.mjs")).href;
  const childEnvironment = { ...process.env };
  delete childEnvironment.NODE_TEST_CONTEXT;
  const directory = mkdtempSync(path.join(tmpdir(), "cartulary-contract-isolation."));
  try {
    const probe = spawnSync(process.execPath, ["--input-type=module", "--eval", `
      import fs from "node:fs";
      import { syncBuiltinESMExports } from "node:module";
      const root = process.argv[1];
      for (const name of ["readFileSync", "existsSync", "statSync", "lstatSync", "readdirSync", "mkdirSync", "mkdtempSync", "writeFileSync"]) {
        const original = fs[name];
        fs[name] = (...args) => {
          const file = String(args[0]);
          const code = /\\.[cm]?js$/.test(file) || file.includes("node_modules");
          if (!code && (file.startsWith(root) || ["mkdirSync", "mkdtempSync", "writeFileSync"].includes(name)))
            throw new Error("unexpected import I/O: " + name + " " + file);
          return original(...args);
        };
      }
      syncBuiltinESMExports();
      for (const module of process.argv.slice(2)) await import(module);
    `, context.root, ...supportModules], { encoding: "utf8", timeout: 30_000, env: childEnvironment });
    assert.equal(probe.status, 0, probe.stderr);
    assert.equal(probe.stdout, "", "support imports must not register or run tests");

    const policyPath = path.join(directory, "invalid-policy.json");
    writeFileSync(policyPath, JSON.stringify({ schema_id: "invalid" }));
    const child = spawnSync(process.execPath, ["--test-reporter=tap", "--input-type=module", "--eval", `
      import test from "node:test";
      import assert from "node:assert/strict";
      import { createContractTestContext } from ${JSON.stringify(contextURL)};
      import { validatePostgresFixturePolicy } from ${JSON.stringify(policyURL)};
      const root = process.argv[1];
      const policyPath = process.argv[2];
      test("invalid_policy", () => {
        const context = createContractTestContext(root, {
          catalogMetadata: () => validatePostgresFixturePolicy(root, [], { policyPath }),
        });
        void context.catalogMetadata;
      });
      test("independent_command_identity", () => {
        const context = createContractTestContext(root);
        assert.equal(context.taskSurface.targets.find((target) => target.name === "test-slice").command_id,
          "cartulary.harness.command.test_slice.v5");
      });
      test("independent_context_lifetime", () => {
        const first = createContractTestContext(root, { topology: () => ({ fresh: true }) });
        const second = createContractTestContext(root, { topology: () => ({ fresh: true }) });
        assert.deepEqual(first.topology, { fresh: true });
        assert.notEqual(first.topology, second.topology);
      });
    `, context.root, policyPath], { encoding: "utf8", timeout: 60_000, env: childEnvironment });
    const output = child.stdout + child.stderr;
    assert.equal(child.status, 1, output);
    assert.match(output, /(?:^|\n)not ok[^\n]*invalid_policy/u);
    for (const name of ["independent_command_identity", "independent_context_lifetime"]) {
      assert.equal(output.split("# Subtest: " + name).length - 1, 1, output);
      assert.match(output, new RegExp("(?:^|\\n)ok[^\\n]*" + name, "u"));
    }
    assert.match(output, /# tests 3(?:\n|$)/u);
    assert.match(output, /# pass 2(?:\n|$)/u);
    assert.match(output, /# fail 1(?:\n|$)/u);
  } finally {
    rmSync(directory, { recursive: true, force: true });
  }
}
