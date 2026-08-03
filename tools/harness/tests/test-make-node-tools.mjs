#!/usr/bin/env node

import assert from "node:assert/strict";
import { readFileSync } from "node:fs";
import path from "node:path";

import {
  buildMakeNodeToolChildEnv,
  buildMakeNodeToolInvocation,
  makeNodeToolNames,
} from "../command-surface/make-node-tools.mjs";

const root = path.resolve(import.meta.dirname, "../../..");
const owner = JSON.parse(readFileSync(path.join(root, "tools/task_surface_owner.json"), "utf8"));
const expected = Object.entries(owner.make_recipes)
  .filter(([, recipe]) => new Set(["node_tool", "owner_command"]).has(recipe.type))
  .map(([name]) => name)
  .sort();
assert.deepEqual(makeNodeToolNames(), expected);
assert.deepEqual(
  buildMakeNodeToolInvocation("harness-performance-check", { EVIDENCE_ROOTS_FILE: "/tmp/window.json" }).args,
  ["check", "--evidence-roots-file", "/tmp/window.json"],
);
assert.deepEqual(
  buildMakeNodeToolInvocation("go-test-duration-baselines", { RESULTS_DIR: "/tmp/run" }).args,
  ["--results-dir", "/tmp/run", "observe", "--view", "backend_go"],
);
assert.deepEqual(
  buildMakeNodeToolInvocation("test-evidence-audit", { OWNER: "module.auth", EVIDENCE_ROOTS_FILE: "/tmp/roots.json" }).args,
  ["--owner", "module.auth", "--evidence-roots-file", "/tmp/roots.json"],
);
const child = buildMakeNodeToolChildEnv("harness-performance-check", {
  EVIDENCE_ROOTS_FILE: "/tmp/window.json",
  PATH: "/bin",
});
assert.equal(child.PATH, "/bin");
assert.equal(Object.hasOwn(child, "EVIDENCE_ROOTS_FILE"), false);
const source = readFileSync(path.join(root, "tools/harness/command-surface/make-node-tools.mjs"), "utf8");
for (const retired of [
  "GO_TEST_DURATION_BASELINE",
  "BROWSER_E2E_DURATION_BASELINE",
  "SERVICE_BACKED_MAKE_TARGET_DURATION_BASELINE",
  "HARNESS_SMOKE_DURATION_BASELINE",
]) {
  assert.equal(source.includes(retired), false, `${retired} must not remain a live input`);
}
