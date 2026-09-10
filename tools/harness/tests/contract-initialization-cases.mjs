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
    "taskSurface", "topology", "rowMigrations", "catalog", "compiler",
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
  const supportURL = pathToFileURL(path.join(context.root, "tools/harness/tests/contract-suite-support.mjs")).href;
  const contextURL = pathToFileURL(path.join(context.root, "tools/harness/tests/contract-test-context.mjs")).href;
  const policyURL = pathToFileURL(path.join(context.root, "tools/harness/test-catalog/postgres-fixture-policy.mjs")).href;
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
      await import(process.argv[2]);
    `, context.root, supportURL], { encoding: "utf8", timeout: 30_000, env: childEnvironment });
    assert.equal(probe.status, 0, probe.stderr);

    const policyPath = path.join(directory, "invalid-policy.json");
    writeFileSync(policyPath, JSON.stringify({ schema_id: "invalid" }));
    const child = spawnSync(process.execPath, ["--test-reporter=tap", "--input-type=module", "--eval", `
      import { runContractSuite } from ${JSON.stringify(supportURL)};
      import { createContractTestContext } from ${JSON.stringify(contextURL)};
      import { validatePostgresFixturePolicy } from ${JSON.stringify(policyURL)};
      const root = process.argv[1];
      const policyPath = process.argv[2];
      const options = { contextFactory: () => createContractTestContext(root, {
        catalog: () => validatePostgresFixturePolicy(root, [], { policyPath }),
      }) };
      runContractSuite("evidence", options);
      runContractSuite("command_surface", options);
    `, context.root, policyPath], { encoding: "utf8", timeout: 60_000, env: childEnvironment });
    const output = child.stdout + child.stderr;
    assert.equal(child.status, 1, output);
    assert.match(output, /postgres_catalog_closure/u);
    assert.match(output, /not ok[^\n]*postgres_catalog_closure/u);
    assert.match(output, /(?:^|\n)ok[^\n]*public_command_identity_contract/u);
    assert.match(output, /(?:^|\n)ok[^\n]*go_selector_build_context/u);
    assert.match(output, /(?:^|\n)ok[^\n]*public_output_contract/u);
  } finally {
    rmSync(directory, { recursive: true, force: true });
  }
}
