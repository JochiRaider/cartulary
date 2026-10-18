#!/usr/bin/env node

import { loadExecutionTopology } from "../generated-artifacts/execution-topology.mjs";
import assert from "node:assert/strict";
import { readFileSync } from "node:fs";
import path from "node:path";

import {
  buildMakeNodeToolChildEnv,
  buildMakeNodeToolInvocation,
  makeNodeToolNames,
} from "../command-surface/make-node-tools.mjs";

const root = path.resolve(import.meta.dirname, "../../..");
const owner = loadExecutionTopology({ root }).taskSurface;
const operations = JSON.parse(readFileSync(path.join(root, "tools/harness_work_graph_owner.json"))).operations;
const expected = [...new Set(Object.values(operations).filter((op) => ["node_tool", "owner_command"].includes(op.execution.type)).map((op) => op.execution.tool_id))].sort();
assert.deepEqual(makeNodeToolNames(), expected);
assert.deepEqual(buildMakeNodeToolInvocation("harness-observability-check", { RESULTS_DIR: "/tmp/run", CHECK: "timing" }).args,
  ["--results-dir", "/tmp/run", "--check", "timing"]);
for (const name of ["harness-performance-check", "harness-public-target-duration-baselines", "go-test-duration-baselines"]) {
  assert.throws(() => buildMakeNodeToolInvocation(name, {}), /unknown make node tool/u);
}
assert.deepEqual(
  buildMakeNodeToolInvocation("test-evidence-audit", { OWNER: "module.auth", EVIDENCE_ROOTS_FILE: "/tmp/roots.json" }).args,
  ["--owner", "module.auth", "--evidence-roots-file", "/tmp/roots.json"],
);
const child = buildMakeNodeToolChildEnv("test-evidence-audit", {
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

assert.deepEqual(buildMakeNodeToolInvocation("target-plan", {
  HOME: process.env.HOME,
  TARGET: "check", DETAIL: "logs", JSON: "", TASK_SURFACE_MANIFEST: "/unavailable/internal-override.json",
  CARTULARY_MAKE_INPUT_SOURCES: "TARGET=cli DETAIL=env JSON=unset",
}).args, ["--target", "check", "--view", "summary"], "dispatch honors canonical selector sources and ignores inherited internal overrides");
assert.deepEqual(buildMakeNodeToolInvocation("task-guide", {
  HOME: process.env.HOME,
  OWNER: "module.networkflow", JSON: "", CARTULARY_MAKE_INPUT_SOURCES: "OWNER=cli JSON=unset",
}).args, ["--mode", "task-guide", "--owner", "module.networkflow", "--detail", "summary"], "false boolean defaults are not forwarded as explicit flags");
