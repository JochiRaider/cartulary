import assert from "node:assert/strict";
import path from "node:path";
import test from "node:test";
import { spawnSync } from "../../workspace/child-process.mjs";
import { createFixture, repository } from "./fixtures.mjs";

function cliFixture(t) {
  const fixture = createFixture(t);
  fixture.write("src/example.ts", 'import { value as local } from "fixture-package";');
  fixture.write("boundaries.json", {
    schema_id: "cartulary.frontend_import_boundaries.v2",
    scan_roots: ["src"],
    rules: [{
      id: "synthetic-warning", level: "warning", message: "Fixture policy",
      applies_to: { include: ["**"] },
      restricted_imports: [{ kind: "package", name: "fixture-package" }],
    }],
  });
  return fixture;
}

function invoke(fixture, args = [], overrides = {}) {
  const env = { ...process.env, ...overrides };
  if (!Object.hasOwn(overrides, "CARTULARY_TEST_TARGET")) delete env.CARTULARY_TEST_TARGET;
  const child = spawnSync(process.execPath, [
    path.join(repository, "tools/harness/static-analysis/frontend-import-boundary-check-cli.mjs"),
    "--root", fixture.root, "--config", "boundaries.json", ...args,
  ], { env, encoding: "utf8", timeout: 10000 });
  assert.equal(child.error, undefined);
  return child;
}

test("frontend CLI: diagnostics and warnings", (t) => {
  const fixture = cliFixture(t);
  const warning = invoke(fixture);
  assert.equal(warning.status, 0, warning.stderr);
  assert.match(warning.stderr, /src\/example.ts:1:.*warning: synthetic-warning: Fixture policy/);
  assert.match(warning.stderr, /0 error|1 warning\(s\), 1 file\(s\) checked/);
  const blocking = invoke(fixture, ["--warnings-as-errors"]);
  assert.equal(blocking.status, 1);
  assert.match(blocking.stderr, /error: synthetic-warning/);
  assert.match(blocking.stderr, /failed: 0 error\(s\), 1 warning\(s\), 1 file\(s\) checked/);
  fixture.write("src/example.ts", "export const allowed = true;");
  const clean = invoke(fixture, [], { CARTULARY_REPO_ROOT: "/absent", FRONTEND_IMPORT_BOUNDARIES_CONFIG: "/absent" });
  assert.equal(clean.status, 0, clean.stderr);
  assert.match(clean.stdout, /frontend import boundaries verified: 1 file\(s\) checked/);
  assert.equal(clean.stderr, "");
});

test("frontend CLI: invalid arguments and configuration", (t) => {
  const fixture = cliFixture(t);
  for (const args of [["--unknown"], ["--config"], ["--root"]]) {
    const child = invoke(fixture, args);
    assert.equal(child.status, 1);
    assert.match(child.stderr, /usage:/);
  }
  for (const config of ["{invalid", {}, { schema_id: "cartulary.frontend_import_boundaries.v2", scan_roots: ["src"], rules: [{}] }]) {
    fixture.write("boundaries.json", config);
    const child = invoke(fixture);
    assert.equal(child.status, 1);
    assert.match(child.stderr, /frontend import boundary check failed:/);
  }
});

test("frontend CLI: target publication", (t) => {
  const fixture = cliFixture(t);
  fixture.write("src/example.ts", "export const allowed = true;");
  fixture.write("publish.mjs", 'import { writeFileSync } from "node:fs"; writeFileSync(new URL("./published.json", import.meta.url), JSON.stringify(process.argv.slice(2)));');
  const env = { CARTULARY_TEST_TARGET: "frontend-import-boundary-check", TEST_OUTPUT_SCRIPT: path.join(fixture.root, "publish.mjs") };
  const child = invoke(fixture, [], env);
  assert.equal(child.status, 0, child.stderr);
  assert.deepEqual(fixture.readJSON("published.json"), ["target-summary", "frontend-import-boundary-check", "pass", "--quiet-success", "--suppress-machine-output"]);
  fixture.write("publish.mjs", "process.exit(11);");
  assert.equal(invoke(fixture, [], env).status, 11);
});
