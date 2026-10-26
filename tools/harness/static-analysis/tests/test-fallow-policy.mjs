import assert from "node:assert/strict";
import { rmSync } from "node:fs";
import path from "node:path";
import test from "node:test";
import {
  runFallowBatch, readPackageSurfaceBindings, collectBlockingPackageSurfaceFindings,
  collectBlockingWebSurfaceFindings, validateFallowReport,
} from "../fallow-policy.mjs";
import { buildResolvedFallowConfig } from "../fallow-reachability.mjs";
import { createFixture } from "./fixtures.mjs";
import { createReachabilityFixture } from "./fallow-fixtures.mjs";

test("Fallow batch starts all children and drains reversed settlement", async () => {
  const specs = ["dead-code", "dupes", "health"].map((name) => ({ name, reportRoot: "/fixture", args: [name], outputFile: `/fixture/${name}.json` }));
  const gates = specs.map(() => Promise.withResolvers());
  const started = [];
  let completed = false;
  const pending = runFallowBatch(specs, (_root, name) => {
    started.push(name);
    return gates[specs.findIndex((spec) => spec.name === name)].promise;
  }).then((result) => { completed = true; return result; });
  await Promise.resolve();
  assert.deepEqual(started, specs.map((spec) => spec.name));
  gates[2].resolve({ name: "health", status: "pass", exitCode: 0 });
  gates[1].reject(new Error("start failure"));
  await Promise.resolve();
  assert.equal(completed, false);
  gates[0].resolve({ name: "dead-code", status: "fail", exitCode: 1 });
  const results = await pending;
  assert.deepEqual(results.map((result) => result.name), started);
  assert.deepEqual(results.map((result) => result.exitCode), [1, 127, 0]);
  assert.deepEqual(results.filter((result) => result.status === "fail").map((result) => result.name), ["dead-code", "dupes"]);
});

test("Fallow batch drains siblings after a synchronous start failure", async () => {
  const started = [];
  const result = await runFallowBatch(["first", "second"].map((name) => ({ name, reportRoot: "/fixture", args: [], outputFile: `/fixture/${name}.json` })), (_root, name) => {
    started.push(name);
    if (name === "first") throw new Error("cannot start");
    return { name, status: "pass", exitCode: 0 };
  });
  assert.deepEqual(started, ["first", "second"]);
  assert.deepEqual(result.map((entry) => entry.exitCode), [127, 0]);
});

test("Fallow package bindings use only the supplied root", (t) => {
  const fixture = createFixture(t);
  fixture.write("packages/example/src/index.ts", 'export { source as publicValue, type PublicType } from "./surface"; export const local = 1; export type LocalType = string; export type { Generated } from "./generated/types";');
  fixture.write("packages/example/src/surface.ts", "export const source = 1; export type PublicType = string; export type PrivateType = number;");
  fixture.write("packages/example/src/generated/types.ts", "export type Generated = string;");
  const policy = { packages: [{ package_name: "@fixture/example", entrypoint: "packages/example/src/index.ts" }], rules: ["unused_exports", "unused_types"] };
  const bindings = readPackageSurfaceBindings(fixture.root, policy);
  const findings = collectBlockingPackageSurfaceFindings({
    unused_exports: [{ path: "packages/example/src/surface.ts", export_name: "source" }, { path: "packages/example/src/index.ts", export_name: "local" }],
    unused_types: [{ path: "packages/example/src/surface.ts", export_name: "PublicType" }, { path: "packages/example/src/surface.ts", export_name: "PrivateType" }, { path: "packages/example/src/generated/types.ts", export_name: "Generated" }],
  }, policy, bindings);
  assert.deepEqual(findings.map((entry) => [entry.symbol, entry.source_symbol]), [["publicValue", "source"], ["local", "local"], ["PublicType", "PublicType"]]);
  assert.ok(bindings.has("packages/example/src/index.ts\u0000LocalType"));
  fixture.write("packages/example/src/index.ts", 'export * from "./surface";');
  assert.throws(() => readPackageSurfaceBindings(fixture.root, policy), /wildcard or namespace/);
});

test("Fallow reachability resolves declared roots without hiding peers", (t) => {
  const fixture = createReachabilityFixture(t);
  const result = buildResolvedFallowConfig({ root: fixture.root, outputFile: path.join(fixture.root, "resolved.json") });
  assert.equal(result.stats.vite_module_entry_points, 2);
  assert.equal(result.stats.typescript_compile_entry_points, 1);
  assert.ok(result.stats.task_surface_entry_points >= 3);
  assert.equal(result.stats.harness_entry_points, 1);
  assert.equal(result.stats.harness_dynamic_export_files, 1);
  assert.equal(result.stats.harness_dynamic_exports, 1);
  for (const file of ["apps/web/src/measurement/main.ts", "packages/example/src/contract.compile.ts", "tools/harness/static-analysis/shell-child.mjs"]) assert.ok(result.config.entry.includes(file), file);
  assert.equal(result.config.entry.includes("packages/example/src/orphan.ts"), false);
});

test("Fallow web findings validate bounded policy reports", () => {
  const policy = { root: "apps/web", rules: ["unused_files", "unused_exports", "unused_types"] };
  const report = { unused_files: [{ path: "apps/web/src/orphan.ts" }, { path: "packages/example/src/orphan.ts" }], unused_exports: [{ path: "apps/web/src/surface.ts", export_name: "dead" }], unused_types: [{ path: "apps/web/src/surface.ts", export_name: "DeadType" }] };
  assert.deepEqual(collectBlockingWebSurfaceFindings(report, policy).map(({ rule, path, symbol }) => [rule, path, symbol]), [
    ["unused_files", "apps/web/src/orphan.ts", null], ["unused_exports", "apps/web/src/surface.ts", "dead"], ["unused_types", "apps/web/src/surface.ts", "DeadType"],
  ]);
  assert.deepEqual(collectBlockingWebSurfaceFindings({ unused_files: [], unused_exports: [], unused_types: [], unused_class_members: [{ path: "apps/web/src/live.ts", name: "read" }], duplicate_exports: [{}] }, policy), []);
  assert.throws(() => collectBlockingWebSurfaceFindings({}, policy), /missing unused_files/);
  for (const unsafe of ["apps/web/../escape.ts", "/apps/web/file.ts", "apps\\web\\file.ts"]) assert.throws(() => collectBlockingWebSurfaceFindings({ ...report, unused_files: [{ path: unsafe }] }, policy), /invalid relative path/);
  assert.throws(() => collectBlockingWebSurfaceFindings({ ...report, unused_types: [{ path: "apps/web/src/file.ts" }] }, policy), /no export name/);
});

test("Fallow validates only current report kinds and consumed fields", () => {
  validateFallowReport({ kind: "dead-code", unused_files: [], unused_exports: [], unused_types: [] }, "dead-code");
  validateFallowReport({ kind: "dupes", clone_groups: [], clone_families: [] }, "dupes");
  validateFallowReport({ kind: "health", findings: [], future_field: true }, "health");
  for (const value of [null, [], {}, { kind: "dead-code" }, { kind: "health", findings: [null] }, { kind: "health", findings: [], error: true }]) {
    assert.throws(() => validateFallowReport(value, "health"));
  }
});

for (const [name, file] of [
  ["public asset", "apps/web/public/assets/fonts/fonts.css"],
  ["task script", "tools/harness/static-analysis/example-cli.mjs"],
  ["setup", "apps/web/src/testing/testSetup.ts"],
  ["harness entrypoint", "tools/harness/test-support/example-direct.mjs"],
]) {
  test(`Fallow rejects missing ${name}`, (t) => {
    const fixture = createReachabilityFixture(t);
    rmSync(path.join(fixture.root, file));
    assert.throws(() => buildResolvedFallowConfig({ root: fixture.root, outputFile: path.join(fixture.root, "resolved.json") }), /missing|exist|file/i);
  });
}
