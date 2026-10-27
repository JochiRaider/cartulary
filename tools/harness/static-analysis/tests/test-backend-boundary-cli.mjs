import assert from "node:assert/strict";
import { createHash } from "node:crypto";
import path from "node:path";
import test from "node:test";
import { spawnSync } from "../../workspace/child-process.mjs";
import { createFixture, repository } from "./fixtures.mjs";
import { emptyPolicy } from "./backend-boundary-fixtures.mjs";

export function registerBoundariesTests() {
  test("backend CLI: success, sorted violations and invalid input preserve process contract", (t) => {
    const fixture = createFixture(t);
    const policy = emptyPolicy({ scan_roots: ["src"], forbidden_go_imports: [{
      id: "private-owner", match_kind: "subtree", imports: ["example/private"],
      scan_paths: ["src/**"], allowed_paths: [], production_only: true,
    }] });
    const support = { schema_id: "cartulary.test_support_inventory.v1", go_support_roots: [{
      path: "support", runtime_scan: "excluded", support_scan: "included", service_starting: false,
    }] };
    fixture.write("boundaries.json", policy);
    fixture.write("support.json", support);
    fixture.write("src/allowed.go", "package allowed");
    const invoke = (args = []) => spawnSync(process.execPath, [
      path.join(repository, "tools/harness/static-analysis/backend-module-boundary-check-cli.mjs"),
      "--root", fixture.root, "--manifest", "boundaries.json", "--support-inventory", "support.json", ...args,
    ], {
      encoding: "utf8", timeout: 10000,
      env: { ...process.env, CARTULARY_REPO_ROOT: "/absent", BACKEND_MODULE_BOUNDARIES_MANIFEST: "/absent", TEST_SUPPORT_INVENTORY: "/absent" },
    });
    const success = invoke();
    assert.equal(success.error, undefined);
    assert.equal(success.status, 0, success.stderr);
    assert.equal(success.stderr, "");
    const summary = JSON.parse(success.stdout);
    const digest = (value) => `sha256:${createHash("sha256").update(JSON.stringify(value)).digest("hex")}`;
    assert.match(summary.checked_at, /^\d{4}-\d{2}-\d{2}T.+Z$/u);
    assert.deepEqual(summary, {
      schema_id: "cartulary.backend_module_boundary_summary.v1", checked_at: summary.checked_at,
      repo_root: fixture.root, manifest_digest: digest(policy), support_inventory_digest: digest(support),
      effective_scan_excludes: ["support/**"], dynamic_relation_sources: [], violations: [], result: "pass",
    });
    for (const name of ["z", "a"]) fixture.write(`src/${name}.go`, 'package forbidden\nimport "example/private/child"');
    const failure = invoke();
    assert.equal(failure.status, 1);
    assert.equal(failure.stderr, "");
    const failedSummary = JSON.parse(failure.stdout);
    assert.equal(failedSummary.result, "fail");
    assert.deepEqual(failedSummary.violations, ["a", "z"].map((name) => ({
      code: "forbidden_go_import", path: `src/${name}.go`, symbol_or_import: "example/private/child",
      rule_id: "private-owner", match_kind: "subtree", result: "fail",
    })));
    fixture.write("boundaries.json", "{invalid");
    const invalid = invoke();
    assert.equal(invalid.status, 1);
    assert.equal(invalid.stdout, "");
    assert.match(invalid.stderr, /^backend module boundary check failed:/u);
    const usage = invoke(["--unknown"]);
    assert.equal(usage.status, 1);
    assert.equal(usage.stdout, "");
    assert.match(usage.stderr, /^backend module boundary check failed: usage:/u);
  });
}
